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

package storage

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/blubskye/godiscordmobileclient/internal/gateway"
	"github.com/blubskye/godiscordmobileclient/internal/models"
)

// HybridCache combines in-memory cache with SQLite persistence
type HybridCache struct {
	mu     sync.RWMutex
	config *Config
	db     *Database

	// In-memory cache (hot data)
	user          *models.User
	guilds        map[string]*models.Guild
	channels      map[string]*models.Channel
	guildChannels map[string][]string
	dmChannels    []string
	messages      map[string][]*models.Message // channel_id -> messages
	relationships map[string]*models.Relationship
	presences     map[string]*models.Presence
	voiceStates   map[string]*models.VoiceState
	typing        map[string]map[string]int64 // channel_id -> user_id -> timestamp

	// LRU tracking for memory management
	channelAccess map[string]time.Time // channel_id -> last access time
	guildAccess   map[string]time.Time // guild_id -> last access time

	// Background tasks
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}

	// Event callbacks
	onGuildCreate        func(*models.Guild)
	onGuildDelete        func(string)
	onChannelCreate      func(*models.Channel)
	onChannelDelete      func(*models.Channel)
	onMessageCreate      func(*models.Message)
	onMessageUpdate      func(*models.Message)
	onMessageDelete      func(channelID, messageID string)
	onTypingStart        func(*models.TypingStart)
	onPresenceUpdate     func(*models.Presence)
	onRelationshipAdd    func(*models.Relationship)
	onRelationshipRemove func(string)
	onVoiceStateUpdate   func(*models.VoiceState)
	onVoiceServerUpdate  func(*gateway.VoiceServerUpdateData)
	onReady              func()
}

// NewHybridCache creates a new hybrid cache
func NewHybridCache(config *Config) (*HybridCache, error) {
	// Calculate auto limits if needed
	if err := config.CalculateAutoLimits(); err != nil {
		log.Printf("Warning: failed to calculate auto limits: %v", err)
	}

	db, err := NewDatabase(config)
	if err != nil {
		return nil, err
	}

	hc := &HybridCache{
		config:        config,
		db:            db,
		guilds:        make(map[string]*models.Guild),
		channels:      make(map[string]*models.Channel),
		guildChannels: make(map[string][]string),
		dmChannels:    make([]string, 0),
		messages:      make(map[string][]*models.Message),
		relationships: make(map[string]*models.Relationship),
		presences:     make(map[string]*models.Presence),
		voiceStates:   make(map[string]*models.VoiceState),
		typing:        make(map[string]map[string]int64),
		channelAccess: make(map[string]time.Time),
		guildAccess:   make(map[string]time.Time),
		stopCleanup:   make(chan struct{}),
	}

	// Load guilds from database
	if err := hc.loadFromDatabase(); err != nil {
		log.Printf("Warning: failed to load from database: %v", err)
	}

	// Start background cleanup
	hc.startCleanupRoutine()

	return hc, nil
}

// Close closes the cache and database
func (hc *HybridCache) Close() error {
	close(hc.stopCleanup)
	if hc.cleanupTicker != nil {
		hc.cleanupTicker.Stop()
	}
	return hc.db.Close()
}

// loadFromDatabase loads essential data from SQLite
func (hc *HybridCache) loadFromDatabase() error {
	ctx := context.Background()

	// Load guilds
	guilds, err := hc.db.GetAllGuilds(ctx)
	if err != nil {
		return err
	}

	hc.mu.Lock()
	for _, g := range guilds {
		hc.guilds[g.ID] = g
		hc.guildAccess[g.ID] = time.Now()
	}
	hc.mu.Unlock()

	// Load channels for each guild
	for _, g := range guilds {
		channels, err := hc.db.GetGuildChannels(ctx, g.ID)
		if err != nil {
			continue
		}
		hc.mu.Lock()
		channelIDs := make([]string, 0, len(channels))
		for _, ch := range channels {
			hc.channels[ch.ID] = ch
			channelIDs = append(channelIDs, ch.ID)
		}
		hc.guildChannels[g.ID] = channelIDs
		hc.mu.Unlock()
	}

	return nil
}

// startCleanupRoutine starts the background cleanup
func (hc *HybridCache) startCleanupRoutine() {
	interval := time.Duration(hc.config.CleanupIntervalMinutes) * time.Minute
	hc.cleanupTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-hc.cleanupTicker.C:
				hc.runCleanup()
			case <-hc.stopCleanup:
				return
			}
		}
	}()
}

// runCleanup performs memory and database cleanup
func (hc *HybridCache) runCleanup() {
	ctx := context.Background()

	// Check memory usage
	if hc.config.CleanupOnLowMemory {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		// If using more than 80% of available memory, be more aggressive
		// This is a simplified check - real implementation would check system memory
	}

	// Trim in-memory message cache
	hc.trimMemoryMessages()

	// Trim presences
	hc.trimPresences()

	// Run database cleanup
	if err := hc.db.Cleanup(ctx); err != nil {
		log.Printf("Database cleanup failed: %v", err)
	}
}

// trimMemoryMessages removes old messages from memory, keeping them in database
func (hc *HybridCache) trimMemoryMessages() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	maxPerChannel := hc.config.MemoryMessagesPerChannel

	for channelID, msgs := range hc.messages {
		if len(msgs) > maxPerChannel {
			// Keep only the newest messages in memory
			hc.messages[channelID] = msgs[len(msgs)-maxPerChannel:]
		}
	}
}

// trimPresences removes old presences from memory
func (hc *HybridCache) trimPresences() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if len(hc.presences) <= hc.config.MemoryPresencesMax {
		return
	}

	// Simple approach: remove offline presences first
	toRemove := len(hc.presences) - hc.config.MemoryPresencesMax
	removed := 0
	for userID, presence := range hc.presences {
		if presence.Status == models.StatusOffline {
			delete(hc.presences, userID)
			removed++
			if removed >= toRemove {
				break
			}
		}
	}
}

// Event handlers setters
func (hc *HybridCache) OnGuildCreate(fn func(*models.Guild))                 { hc.onGuildCreate = fn }
func (hc *HybridCache) OnGuildDelete(fn func(string))                        { hc.onGuildDelete = fn }
func (hc *HybridCache) OnChannelCreate(fn func(*models.Channel))             { hc.onChannelCreate = fn }
func (hc *HybridCache) OnChannelDelete(fn func(*models.Channel))             { hc.onChannelDelete = fn }
func (hc *HybridCache) OnMessageCreate(fn func(*models.Message))             { hc.onMessageCreate = fn }
func (hc *HybridCache) OnMessageUpdate(fn func(*models.Message))             { hc.onMessageUpdate = fn }
func (hc *HybridCache) OnMessageDelete(fn func(channelID, messageID string)) { hc.onMessageDelete = fn }
func (hc *HybridCache) OnTypingStart(fn func(*models.TypingStart))           { hc.onTypingStart = fn }
func (hc *HybridCache) OnPresenceUpdate(fn func(*models.Presence))           { hc.onPresenceUpdate = fn }
func (hc *HybridCache) OnRelationshipAdd(fn func(*models.Relationship))      { hc.onRelationshipAdd = fn }
func (hc *HybridCache) OnRelationshipRemove(fn func(string))                 { hc.onRelationshipRemove = fn }
func (hc *HybridCache) OnVoiceStateUpdate(fn func(*models.VoiceState))       { hc.onVoiceStateUpdate = fn }
func (hc *HybridCache) OnVoiceServerUpdate(fn func(*gateway.VoiceServerUpdateData)) {
	hc.onVoiceServerUpdate = fn
}
func (hc *HybridCache) OnReady(fn func()) { hc.onReady = fn }

// HandleEvent processes gateway events
func (hc *HybridCache) HandleEvent(eventType string, data json.RawMessage) {
	switch eventType {
	case gateway.EventReady:
		hc.handleReady(data)
	case gateway.EventGuildCreate:
		hc.handleGuildCreate(data)
	case gateway.EventGuildUpdate:
		hc.handleGuildUpdate(data)
	case gateway.EventGuildDelete:
		hc.handleGuildDelete(data)
	case gateway.EventChannelCreate:
		hc.handleChannelCreate(data)
	case gateway.EventChannelUpdate:
		hc.handleChannelUpdate(data)
	case gateway.EventChannelDelete:
		hc.handleChannelDelete(data)
	case gateway.EventMessageCreate:
		hc.handleMessageCreate(data)
	case gateway.EventMessageUpdate:
		hc.handleMessageUpdate(data)
	case gateway.EventMessageDelete:
		hc.handleMessageDelete(data)
	case gateway.EventTypingStart:
		hc.handleTypingStart(data)
	case gateway.EventPresenceUpdate:
		hc.handlePresenceUpdate(data)
	case gateway.EventVoiceStateUpdate:
		hc.handleVoiceStateUpdate(data)
	case gateway.EventVoiceServerUpdate:
		hc.handleVoiceServerUpdate(data)
	case gateway.EventRelationshipAdd:
		hc.handleRelationshipAdd(data)
	case gateway.EventRelationshipRemove:
		hc.handleRelationshipRemove(data)
	case gateway.EventUserUpdate:
		hc.handleUserUpdate(data)
	}
}

func (hc *HybridCache) handleReady(data json.RawMessage) {
	var ready gateway.ReadyData
	if err := json.Unmarshal(data, &ready); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	hc.user = ready.User

	// Store DM channels
	for _, ch := range ready.PrivateChannels {
		channel := ch
		hc.channels[ch.ID] = &channel
		hc.dmChannels = append(hc.dmChannels, ch.ID)
		// Save to database
		go hc.db.SaveChannel(ctx, &channel)
	}

	// Store relationships
	for _, rel := range ready.Relationships {
		r := rel
		hc.relationships[rel.ID] = &r
	}

	// Store presences
	for _, p := range ready.Presences {
		presence := p
		if p.User != nil {
			hc.presences[p.User.ID] = &presence
		}
	}
	hc.mu.Unlock()

	// Save user to database
	if ready.User != nil {
		go hc.db.SaveUser(ctx, ready.User)
	}

	if hc.onReady != nil {
		hc.onReady()
	}
}

func (hc *HybridCache) handleGuildCreate(data json.RawMessage) {
	var guild models.Guild
	if err := json.Unmarshal(data, &guild); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	hc.guilds[guild.ID] = &guild
	hc.guildAccess[guild.ID] = time.Now()

	// Store channels
	channelIDs := make([]string, 0, len(guild.Channels))
	for _, ch := range guild.Channels {
		channel := ch
		channel.GuildID = guild.ID
		hc.channels[ch.ID] = &channel
		channelIDs = append(channelIDs, ch.ID)
	}
	hc.guildChannels[guild.ID] = channelIDs

	// Store voice states
	for _, vs := range guild.VoiceStates {
		voiceState := vs
		hc.voiceStates[vs.UserID] = &voiceState
	}

	// Store presences
	for _, p := range guild.Presences {
		presence := p
		if p.User != nil {
			hc.presences[p.User.ID] = &presence
		}
	}
	hc.mu.Unlock()

	// Save to database asynchronously
	go func() {
		hc.db.SaveGuild(ctx, &guild)
		for _, ch := range guild.Channels {
			channel := ch
			channel.GuildID = guild.ID
			hc.db.SaveChannel(ctx, &channel)
		}
	}()

	if hc.onGuildCreate != nil {
		hc.onGuildCreate(&guild)
	}
}

func (hc *HybridCache) handleGuildUpdate(data json.RawMessage) {
	var guild models.Guild
	if err := json.Unmarshal(data, &guild); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	if existing, ok := hc.guilds[guild.ID]; ok {
		guild.Channels = existing.Channels
		guild.Members = existing.Members
	}
	hc.guilds[guild.ID] = &guild
	hc.guildAccess[guild.ID] = time.Now()
	hc.mu.Unlock()

	go hc.db.SaveGuild(ctx, &guild)
}

func (hc *HybridCache) handleGuildDelete(data json.RawMessage) {
	var guild models.UnavailableGuild
	if err := json.Unmarshal(data, &guild); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	if channelIDs, ok := hc.guildChannels[guild.ID]; ok {
		for _, id := range channelIDs {
			delete(hc.channels, id)
			delete(hc.messages, id)
		}
		delete(hc.guildChannels, guild.ID)
	}
	delete(hc.guilds, guild.ID)
	delete(hc.guildAccess, guild.ID)
	hc.mu.Unlock()

	go hc.db.DeleteGuild(ctx, guild.ID)

	if hc.onGuildDelete != nil {
		hc.onGuildDelete(guild.ID)
	}
}

func (hc *HybridCache) handleChannelCreate(data json.RawMessage) {
	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	hc.channels[channel.ID] = &channel
	if channel.GuildID != "" {
		hc.guildChannels[channel.GuildID] = append(hc.guildChannels[channel.GuildID], channel.ID)
	} else if channel.IsDM() {
		hc.dmChannels = append(hc.dmChannels, channel.ID)
	}
	hc.mu.Unlock()

	go hc.db.SaveChannel(ctx, &channel)

	if hc.onChannelCreate != nil {
		hc.onChannelCreate(&channel)
	}
}

func (hc *HybridCache) handleChannelUpdate(data json.RawMessage) {
	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	hc.channels[channel.ID] = &channel
	hc.mu.Unlock()

	go hc.db.SaveChannel(ctx, &channel)
}

func (hc *HybridCache) handleChannelDelete(data json.RawMessage) {
	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return
	}

	hc.mu.Lock()
	delete(hc.channels, channel.ID)
	delete(hc.messages, channel.ID)
	if channel.GuildID != "" {
		if channels, ok := hc.guildChannels[channel.GuildID]; ok {
			for i, id := range channels {
				if id == channel.ID {
					hc.guildChannels[channel.GuildID] = append(channels[:i], channels[i+1:]...)
					break
				}
			}
		}
	}
	hc.mu.Unlock()

	if hc.onChannelDelete != nil {
		hc.onChannelDelete(&channel)
	}
}

func (hc *HybridCache) handleMessageCreate(data json.RawMessage) {
	var message models.Message
	if err := json.Unmarshal(data, &message); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	hc.messages[message.ChannelID] = append(hc.messages[message.ChannelID], &message)
	hc.channelAccess[message.ChannelID] = time.Now()

	// Trim if over memory limit
	if len(hc.messages[message.ChannelID]) > hc.config.MemoryMessagesPerChannel {
		hc.messages[message.ChannelID] = hc.messages[message.ChannelID][1:]
	}
	hc.mu.Unlock()

	// Save to database
	go hc.db.SaveMessage(ctx, &message)

	// Save author if present
	if message.Author != nil {
		go hc.db.SaveUser(ctx, message.Author)
	}

	if hc.onMessageCreate != nil {
		hc.onMessageCreate(&message)
	}
}

func (hc *HybridCache) handleMessageUpdate(data json.RawMessage) {
	var message models.Message
	if err := json.Unmarshal(data, &message); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	if messages, ok := hc.messages[message.ChannelID]; ok {
		for i, msg := range messages {
			if msg.ID == message.ID {
				hc.messages[message.ChannelID][i] = &message
				break
			}
		}
	}
	hc.mu.Unlock()

	go hc.db.SaveMessage(ctx, &message)

	if hc.onMessageUpdate != nil {
		hc.onMessageUpdate(&message)
	}
}

func (hc *HybridCache) handleMessageDelete(data json.RawMessage) {
	var del gateway.MessageDeleteData
	if err := json.Unmarshal(data, &del); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	if messages, ok := hc.messages[del.ChannelID]; ok {
		for i, msg := range messages {
			if msg.ID == del.ID {
				hc.messages[del.ChannelID] = append(messages[:i], messages[i+1:]...)
				break
			}
		}
	}
	hc.mu.Unlock()

	go hc.db.DeleteMessage(ctx, del.ID)

	if hc.onMessageDelete != nil {
		hc.onMessageDelete(del.ChannelID, del.ID)
	}
}

func (hc *HybridCache) handleTypingStart(data json.RawMessage) {
	var typing models.TypingStart
	if err := json.Unmarshal(data, &typing); err != nil {
		return
	}

	hc.mu.Lock()
	if _, ok := hc.typing[typing.ChannelID]; !ok {
		hc.typing[typing.ChannelID] = make(map[string]int64)
	}
	hc.typing[typing.ChannelID][typing.UserID] = typing.Timestamp
	hc.mu.Unlock()

	if hc.onTypingStart != nil {
		hc.onTypingStart(&typing)
	}
}

func (hc *HybridCache) handlePresenceUpdate(data json.RawMessage) {
	var presence models.Presence
	if err := json.Unmarshal(data, &presence); err != nil {
		return
	}

	hc.mu.Lock()
	if presence.User != nil {
		hc.presences[presence.User.ID] = &presence
	}
	hc.mu.Unlock()

	if hc.onPresenceUpdate != nil {
		hc.onPresenceUpdate(&presence)
	}
}

func (hc *HybridCache) handleVoiceStateUpdate(data json.RawMessage) {
	var vs models.VoiceState
	if err := json.Unmarshal(data, &vs); err != nil {
		return
	}

	hc.mu.Lock()
	if vs.ChannelID == "" {
		delete(hc.voiceStates, vs.UserID)
	} else {
		hc.voiceStates[vs.UserID] = &vs
	}
	hc.mu.Unlock()

	if hc.onVoiceStateUpdate != nil {
		hc.onVoiceStateUpdate(&vs)
	}
}

func (hc *HybridCache) handleVoiceServerUpdate(data json.RawMessage) {
	var vs gateway.VoiceServerUpdateData
	if err := json.Unmarshal(data, &vs); err != nil {
		return
	}

	if hc.onVoiceServerUpdate != nil {
		hc.onVoiceServerUpdate(&vs)
	}
}

func (hc *HybridCache) handleRelationshipAdd(data json.RawMessage) {
	var rel models.Relationship
	if err := json.Unmarshal(data, &rel); err != nil {
		return
	}

	hc.mu.Lock()
	hc.relationships[rel.ID] = &rel
	hc.mu.Unlock()

	if hc.onRelationshipAdd != nil {
		hc.onRelationshipAdd(&rel)
	}
}

func (hc *HybridCache) handleRelationshipRemove(data json.RawMessage) {
	var rel gateway.RelationshipRemoveData
	if err := json.Unmarshal(data, &rel); err != nil {
		return
	}

	hc.mu.Lock()
	delete(hc.relationships, rel.ID)
	hc.mu.Unlock()

	if hc.onRelationshipRemove != nil {
		hc.onRelationshipRemove(rel.ID)
	}
}

func (hc *HybridCache) handleUserUpdate(data json.RawMessage) {
	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return
	}

	ctx := context.Background()

	hc.mu.Lock()
	hc.user = &user
	hc.mu.Unlock()

	go hc.db.SaveUser(ctx, &user)
}

// Getters

func (hc *HybridCache) GetUser() *models.User {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.user
}

func (hc *HybridCache) GetGuild(id string) *models.Guild {
	hc.mu.RLock()
	guild := hc.guilds[id]
	hc.mu.RUnlock()

	if guild != nil {
		hc.mu.Lock()
		hc.guildAccess[id] = time.Now()
		hc.mu.Unlock()
		return guild
	}

	// Try database
	ctx := context.Background()
	guild, _ = hc.db.GetGuild(ctx, id)
	if guild != nil {
		hc.mu.Lock()
		hc.guilds[id] = guild
		hc.guildAccess[id] = time.Now()
		hc.mu.Unlock()
	}
	return guild
}

func (hc *HybridCache) GetGuilds() []*models.Guild {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	guilds := make([]*models.Guild, 0, len(hc.guilds))
	for _, g := range hc.guilds {
		guilds = append(guilds, g)
	}
	return guilds
}

func (hc *HybridCache) GetChannel(id string) *models.Channel {
	hc.mu.RLock()
	channel := hc.channels[id]
	hc.mu.RUnlock()

	if channel != nil {
		hc.mu.Lock()
		hc.channelAccess[id] = time.Now()
		hc.mu.Unlock()
		return channel
	}

	// Try database
	ctx := context.Background()
	channel, _ = hc.db.GetChannel(ctx, id)
	if channel != nil {
		hc.mu.Lock()
		hc.channels[id] = channel
		hc.channelAccess[id] = time.Now()
		hc.mu.Unlock()
	}
	return channel
}

func (hc *HybridCache) GetGuildChannels(guildID string) []*models.Channel {
	hc.mu.RLock()
	ids, ok := hc.guildChannels[guildID]
	if !ok {
		hc.mu.RUnlock()
		// Try database
		ctx := context.Background()
		channels, _ := hc.db.GetGuildChannels(ctx, guildID)
		if channels != nil {
			hc.mu.Lock()
			channelIDs := make([]string, 0, len(channels))
			for _, ch := range channels {
				hc.channels[ch.ID] = ch
				channelIDs = append(channelIDs, ch.ID)
			}
			hc.guildChannels[guildID] = channelIDs
			hc.mu.Unlock()
		}
		return channels
	}

	channels := make([]*models.Channel, 0, len(ids))
	for _, id := range ids {
		if ch, ok := hc.channels[id]; ok {
			channels = append(channels, ch)
		}
	}
	hc.mu.RUnlock()
	return channels
}

func (hc *HybridCache) GetDMChannels() []*models.Channel {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	channels := make([]*models.Channel, 0, len(hc.dmChannels))
	for _, id := range hc.dmChannels {
		if ch, ok := hc.channels[id]; ok {
			channels = append(channels, ch)
		}
	}
	return channels
}

func (hc *HybridCache) GetMessages(channelID string) []*models.Message {
	hc.mu.RLock()
	msgs := hc.messages[channelID]
	hc.mu.RUnlock()

	if len(msgs) > 0 {
		hc.mu.Lock()
		hc.channelAccess[channelID] = time.Now()
		hc.mu.Unlock()
		return msgs
	}

	// Try database
	ctx := context.Background()
	msgs, _ = hc.db.GetMessages(ctx, channelID, hc.config.MemoryMessagesPerChannel, "", "")
	if len(msgs) > 0 {
		hc.mu.Lock()
		hc.messages[channelID] = msgs
		hc.channelAccess[channelID] = time.Now()
		hc.mu.Unlock()
	}
	return msgs
}

func (hc *HybridCache) GetRelationships() []*models.Relationship {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	rels := make([]*models.Relationship, 0, len(hc.relationships))
	for _, r := range hc.relationships {
		rels = append(rels, r)
	}
	return rels
}

func (hc *HybridCache) GetFriends() []*models.Relationship {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	friends := make([]*models.Relationship, 0)
	for _, r := range hc.relationships {
		if r.Type == models.RelationshipTypeFriend {
			friends = append(friends, r)
		}
	}
	return friends
}

func (hc *HybridCache) GetPresence(userID string) *models.Presence {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.presences[userID]
}

func (hc *HybridCache) GetVoiceState(userID string) *models.VoiceState {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.voiceStates[userID]
}

func (hc *HybridCache) AddMessages(channelID string, messages []*models.Message) {
	ctx := context.Background()

	hc.mu.Lock()
	// Prepend messages (they're usually older)
	hc.messages[channelID] = append(messages, hc.messages[channelID]...)

	// Trim if over limit
	if len(hc.messages[channelID]) > hc.config.MemoryMessagesPerChannel {
		hc.messages[channelID] = hc.messages[channelID][len(hc.messages[channelID])-hc.config.MemoryMessagesPerChannel:]
	}
	hc.channelAccess[channelID] = time.Now()
	hc.mu.Unlock()

	// Save to database
	go hc.db.SaveMessages(ctx, messages)
}

// SearchMessages searches messages in the database
func (hc *HybridCache) SearchMessages(query string, channelID string, limit int) []*models.Message {
	ctx := context.Background()
	msgs, _ := hc.db.SearchMessages(ctx, query, channelID, limit)
	return msgs
}

// GetStats returns storage statistics
func (hc *HybridCache) GetStats() *StorageStats {
	ctx := context.Background()
	stats, err := hc.db.GetStats(ctx)
	if err != nil {
		return &StorageStats{}
	}

	// Add memory stats
	hc.mu.RLock()
	for _, msgs := range hc.messages {
		stats.MemoryMessagesCount += len(msgs)
	}
	hc.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats.MemoryUsedBytes = int64(m.Alloc)

	return stats
}

// GetConfig returns the current config
func (hc *HybridCache) GetConfig() *Config {
	return hc.config
}

// UpdateConfig updates the cache configuration
func (hc *HybridCache) UpdateConfig(config *Config) error {
	hc.config = config
	return config.Save()
}
