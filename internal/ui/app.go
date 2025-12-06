// Copyright (C) 2025 blubskye
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// Source code: https://github.com/blubskye/godiscordmobileclient

package ui

import (
	"context"
	"log"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"

	"github.com/blubskye/godiscordmobileclient/internal/api"
	"github.com/blubskye/godiscordmobileclient/internal/gateway"
	"github.com/blubskye/godiscordmobileclient/internal/models"
	"github.com/blubskye/godiscordmobileclient/internal/state"
	"github.com/blubskye/godiscordmobileclient/internal/storage"
	"github.com/blubskye/godiscordmobileclient/internal/ui/screens"
	"github.com/blubskye/godiscordmobileclient/internal/ui/theme"
)

// Screen represents different app screens
type Screen int

const (
	ScreenLogin Screen = iota
	ScreenGuilds
	ScreenChannels
	ScreenChat
	ScreenDMs
	ScreenSettings
	ScreenAbout
)

// App holds the application state
type App struct {
	window *app.Window
	theme  *theme.Theme
	screen Screen
	ctx    context.Context
	cancel context.CancelFunc

	// Discord clients
	api     *api.Client
	gateway *gateway.Client
	cache   state.CacheInterface

	// Screens
	loginScreen    *screens.LoginScreen
	guildsScreen   *screens.GuildsScreen
	channelsScreen *screens.ChannelsScreen
	chatScreen     *screens.ChatScreen
	aboutScreen    *screens.AboutScreen

	// Current state
	currentGuildID   string
	currentChannelID string
}

// NewApp creates a new application instance
func NewApp(w *app.Window) *App {
	ctx, cancel := context.WithCancel(context.Background())

	// Load storage config and create hybrid cache
	config, err := storage.LoadConfig("")
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		config = storage.DefaultConfig()
	}

	cache, err := storage.NewHybridCache(config)
	if err != nil {
		log.Printf("Failed to create hybrid cache, falling back to in-memory: %v", err)
		// Fall back to in-memory cache if hybrid fails
		cache = nil
	}

	// Use CacheInterface - either HybridCache or fall back to in-memory Cache
	var cacheInterface state.CacheInterface
	if cache != nil {
		cacheInterface = cache
	} else {
		cacheInterface = state.NewCache()
	}

	apiClient := api.NewClient("")
	gatewayClient := gateway.NewClient("")

	a := &App{
		window:  w,
		theme:   theme.New(),
		screen:  ScreenLogin,
		ctx:     ctx,
		cancel:  cancel,
		api:     apiClient,
		gateway: gatewayClient,
		cache:   cacheInterface,
	}

	// Initialize screens
	a.loginScreen = screens.NewLoginScreen(a.theme, a.onLogin)
	a.guildsScreen = screens.NewGuildsScreen(a.theme, cacheInterface, a.onGuildSelect)
	a.channelsScreen = screens.NewChannelsScreen(a.theme, cacheInterface, a.onChannelSelect, a.onBackToGuilds)
	a.chatScreen = screens.NewChatScreen(a.theme, cacheInterface, apiClient, a.onBackToChannels)
	a.aboutScreen = screens.NewAboutScreen(a.theme, a.onBackFromAbout)

	// Register gateway event handler
	gatewayClient.OnEvent(cacheInterface.HandleEvent)

	// Set up cache callbacks to invalidate UI
	cacheInterface.OnReady(func() {
		w.Invalidate()
	})
	cacheInterface.OnMessageCreate(func(m *models.Message) {
		w.Invalidate()
	})
	cacheInterface.OnGuildCreate(func(g *models.Guild) {
		w.Invalidate()
	})

	return a
}

// Run starts the application event loop
func (a *App) Run() error {
	var ops op.Ops

	for {
		switch e := a.window.Event().(type) {
		case app.DestroyEvent:
			a.cancel()
			if a.gateway.IsConnected() {
				a.gateway.Disconnect()
			}
			// Close cache if it supports closing (HybridCache does)
			if closer, ok := a.cache.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					log.Printf("Error closing cache: %v", err)
				}
			}
			return e.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

// Layout renders the current screen
func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	// Fill background
	FillBackground(gtx, a.theme.Background, gtx.Constraints.Max)

	switch a.screen {
	case ScreenLogin:
		return a.loginScreen.Layout(gtx)
	case ScreenGuilds:
		return a.guildsScreen.Layout(gtx)
	case ScreenChannels:
		return a.channelsScreen.Layout(gtx, a.currentGuildID)
	case ScreenChat:
		return a.chatScreen.Layout(gtx, a.currentChannelID)
	case ScreenAbout:
		return a.aboutScreen.Layout(gtx)
	default:
		return layout.Dimensions{}
	}
}

// onLogin handles successful login
func (a *App) onLogin(token string) {
	a.api.SetToken(token)
	a.gateway.SetToken(token)

	// Connect to gateway
	go func() {
		if err := a.gateway.Connect(a.ctx); err != nil {
			log.Printf("Gateway connection failed: %v", err)
			return
		}
	}()

	a.screen = ScreenGuilds
	a.window.Invalidate()
}

// onGuildSelect handles guild selection
func (a *App) onGuildSelect(guildID string) {
	a.currentGuildID = guildID
	a.screen = ScreenChannels
	a.window.Invalidate()
}

// onChannelSelect handles channel selection
func (a *App) onChannelSelect(channelID string) {
	a.currentChannelID = channelID
	a.chatScreen.SetChannel(channelID)

	// Load messages
	go func() {
		messages, err := a.api.GetMessages(a.ctx, channelID, &api.GetMessagesParams{Limit: 50})
		if err != nil {
			log.Printf("Failed to load messages: %v", err)
			return
		}
		// Convert to pointers and add to cache
		msgPtrs := make([]*models.Message, len(messages))
		for i := range messages {
			msgPtrs[i] = &messages[i]
		}
		a.cache.AddMessages(channelID, msgPtrs)
		a.window.Invalidate()
	}()

	a.screen = ScreenChat
	a.window.Invalidate()
}

// onBackToGuilds returns to guild list
func (a *App) onBackToGuilds() {
	a.screen = ScreenGuilds
	a.window.Invalidate()
}

// onBackToChannels returns to channel list
func (a *App) onBackToChannels() {
	a.screen = ScreenChannels
	a.window.Invalidate()
}

// onBackFromAbout returns from about screen
func (a *App) onBackFromAbout() {
	a.screen = ScreenGuilds
	a.window.Invalidate()
}

// ShowAbout navigates to the about screen
func (a *App) ShowAbout() {
	a.screen = ScreenAbout
	a.window.Invalidate()
}
