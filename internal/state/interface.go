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

package state

import (
	"encoding/json"

	"github.com/blubskye/godiscordmobileclient/internal/gateway"
	"github.com/blubskye/godiscordmobileclient/internal/models"
)

// CacheInterface defines the interface for cache implementations
// Both the in-memory Cache and hybrid HybridCache implement this
type CacheInterface interface {
	// Event handling
	HandleEvent(eventType string, data json.RawMessage)

	// Event callbacks
	OnGuildCreate(fn func(*models.Guild))
	OnGuildDelete(fn func(string))
	OnChannelCreate(fn func(*models.Channel))
	OnChannelDelete(fn func(*models.Channel))
	OnMessageCreate(fn func(*models.Message))
	OnMessageUpdate(fn func(*models.Message))
	OnMessageDelete(fn func(channelID, messageID string))
	OnTypingStart(fn func(*models.TypingStart))
	OnPresenceUpdate(fn func(*models.Presence))
	OnRelationshipAdd(fn func(*models.Relationship))
	OnRelationshipRemove(fn func(string))
	OnVoiceStateUpdate(fn func(*models.VoiceState))
	OnVoiceServerUpdate(fn func(*gateway.VoiceServerUpdateData))
	OnReady(fn func())

	// Getters
	GetUser() *models.User
	GetGuild(id string) *models.Guild
	GetGuilds() []*models.Guild
	GetChannel(id string) *models.Channel
	GetGuildChannels(guildID string) []*models.Channel
	GetDMChannels() []*models.Channel
	GetMessages(channelID string) []*models.Message
	GetRelationships() []*models.Relationship
	GetFriends() []*models.Relationship
	GetPresence(userID string) *models.Presence
	GetVoiceState(userID string) *models.VoiceState

	// Setters
	AddMessages(channelID string, messages []*models.Message)
}

// Ensure Cache implements CacheInterface
var _ CacheInterface = (*Cache)(nil)
