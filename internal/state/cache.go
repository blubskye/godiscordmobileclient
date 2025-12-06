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
	"sync"

	"github.com/blubskye/godiscordmobileclient/internal/gateway"
	"github.com/blubskye/godiscordmobileclient/internal/models"
)

// Cache holds the application state in memory
type Cache struct {
	mu sync.RWMutex

	// Current user
	User *models.User

	// Guilds indexed by ID
	Guilds map[string]*models.Guild

	// Channels indexed by ID
	Channels map[string]*models.Channel

	// Guild channels indexed by guild ID
	GuildChannels map[string][]string

	// DM channels
	DMChannels []string

	// Messages indexed by channel ID
	Messages map[string][]*models.Message

	// Relationships (friends/blocked)
	Relationships map[string]*models.Relationship

	// Presences indexed by user ID
	Presences map[string]*models.Presence

	// Voice states indexed by user ID
	VoiceStates map[string]*models.VoiceState

	// Typing indicators: channel ID -> user ID -> timestamp
	Typing map[string]map[string]int64

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

// NewCache creates a new state cache
func NewCache() *Cache {
	return &Cache{
		Guilds:        make(map[string]*models.Guild),
		Channels:      make(map[string]*models.Channel),
		GuildChannels: make(map[string][]string),
		DMChannels:    make([]string, 0),
		Messages:      make(map[string][]*models.Message),
		Relationships: make(map[string]*models.Relationship),
		Presences:     make(map[string]*models.Presence),
		VoiceStates:   make(map[string]*models.VoiceState),
		Typing:        make(map[string]map[string]int64),
	}
}

// Event handlers setters
func (c *Cache) OnGuildCreate(fn func(*models.Guild))                 { c.onGuildCreate = fn }
func (c *Cache) OnGuildDelete(fn func(string))                        { c.onGuildDelete = fn }
func (c *Cache) OnChannelCreate(fn func(*models.Channel))             { c.onChannelCreate = fn }
func (c *Cache) OnChannelDelete(fn func(*models.Channel))             { c.onChannelDelete = fn }
func (c *Cache) OnMessageCreate(fn func(*models.Message))             { c.onMessageCreate = fn }
func (c *Cache) OnMessageUpdate(fn func(*models.Message))             { c.onMessageUpdate = fn }
func (c *Cache) OnMessageDelete(fn func(channelID, messageID string)) { c.onMessageDelete = fn }
func (c *Cache) OnTypingStart(fn func(*models.TypingStart))           { c.onTypingStart = fn }
func (c *Cache) OnPresenceUpdate(fn func(*models.Presence))           { c.onPresenceUpdate = fn }
func (c *Cache) OnRelationshipAdd(fn func(*models.Relationship))      { c.onRelationshipAdd = fn }
func (c *Cache) OnRelationshipRemove(fn func(string))                 { c.onRelationshipRemove = fn }
func (c *Cache) OnVoiceStateUpdate(fn func(*models.VoiceState))       { c.onVoiceStateUpdate = fn }
func (c *Cache) OnVoiceServerUpdate(fn func(*gateway.VoiceServerUpdateData)) {
	c.onVoiceServerUpdate = fn
}
func (c *Cache) OnReady(fn func()) { c.onReady = fn }

// HandleEvent processes gateway events and updates state
func (c *Cache) HandleEvent(eventType string, data json.RawMessage) {
	switch eventType {
	case gateway.EventReady:
		c.handleReady(data)
	case gateway.EventGuildCreate:
		c.handleGuildCreate(data)
	case gateway.EventGuildUpdate:
		c.handleGuildUpdate(data)
	case gateway.EventGuildDelete:
		c.handleGuildDelete(data)
	case gateway.EventChannelCreate:
		c.handleChannelCreate(data)
	case gateway.EventChannelUpdate:
		c.handleChannelUpdate(data)
	case gateway.EventChannelDelete:
		c.handleChannelDelete(data)
	case gateway.EventMessageCreate:
		c.handleMessageCreate(data)
	case gateway.EventMessageUpdate:
		c.handleMessageUpdate(data)
	case gateway.EventMessageDelete:
		c.handleMessageDelete(data)
	case gateway.EventTypingStart:
		c.handleTypingStart(data)
	case gateway.EventPresenceUpdate:
		c.handlePresenceUpdate(data)
	case gateway.EventVoiceStateUpdate:
		c.handleVoiceStateUpdate(data)
	case gateway.EventVoiceServerUpdate:
		c.handleVoiceServerUpdate(data)
	case gateway.EventRelationshipAdd:
		c.handleRelationshipAdd(data)
	case gateway.EventRelationshipRemove:
		c.handleRelationshipRemove(data)
	case gateway.EventUserUpdate:
		c.handleUserUpdate(data)
	}
}

func (c *Cache) handleReady(data json.RawMessage) {
	var ready gateway.ReadyData
	if err := json.Unmarshal(data, &ready); err != nil {
		return
	}

	c.mu.Lock()
	c.User = ready.User

	// Store DM channels
	for _, ch := range ready.PrivateChannels {
		channel := ch
		c.Channels[ch.ID] = &channel
		c.DMChannels = append(c.DMChannels, ch.ID)
	}

	// Store relationships
	for _, rel := range ready.Relationships {
		r := rel
		c.Relationships[rel.ID] = &r
	}

	// Store presences
	for _, p := range ready.Presences {
		presence := p
		if p.User != nil {
			c.Presences[p.User.ID] = &presence
		}
	}
	c.mu.Unlock()

	if c.onReady != nil {
		c.onReady()
	}
}

func (c *Cache) handleGuildCreate(data json.RawMessage) {
	var guild models.Guild
	if err := json.Unmarshal(data, &guild); err != nil {
		return
	}

	c.mu.Lock()
	c.Guilds[guild.ID] = &guild

	// Store channels
	channelIDs := make([]string, 0, len(guild.Channels))
	for _, ch := range guild.Channels {
		channel := ch
		channel.GuildID = guild.ID
		c.Channels[ch.ID] = &channel
		channelIDs = append(channelIDs, ch.ID)
	}
	c.GuildChannels[guild.ID] = channelIDs

	// Store voice states
	for _, vs := range guild.VoiceStates {
		voiceState := vs
		c.VoiceStates[vs.UserID] = &voiceState
	}

	// Store presences
	for _, p := range guild.Presences {
		presence := p
		if p.User != nil {
			c.Presences[p.User.ID] = &presence
		}
	}
	c.mu.Unlock()

	if c.onGuildCreate != nil {
		c.onGuildCreate(&guild)
	}
}

func (c *Cache) handleGuildUpdate(data json.RawMessage) {
	var guild models.Guild
	if err := json.Unmarshal(data, &guild); err != nil {
		return
	}

	c.mu.Lock()
	if existing, ok := c.Guilds[guild.ID]; ok {
		// Preserve channels/members since update doesn't include them
		guild.Channels = existing.Channels
		guild.Members = existing.Members
	}
	c.Guilds[guild.ID] = &guild
	c.mu.Unlock()
}

func (c *Cache) handleGuildDelete(data json.RawMessage) {
	var guild models.UnavailableGuild
	if err := json.Unmarshal(data, &guild); err != nil {
		return
	}

	c.mu.Lock()
	// Remove guild channels
	if channelIDs, ok := c.GuildChannels[guild.ID]; ok {
		for _, id := range channelIDs {
			delete(c.Channels, id)
		}
		delete(c.GuildChannels, guild.ID)
	}
	delete(c.Guilds, guild.ID)
	c.mu.Unlock()

	if c.onGuildDelete != nil {
		c.onGuildDelete(guild.ID)
	}
}

func (c *Cache) handleChannelCreate(data json.RawMessage) {
	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return
	}

	c.mu.Lock()
	c.Channels[channel.ID] = &channel
	if channel.GuildID != "" {
		c.GuildChannels[channel.GuildID] = append(c.GuildChannels[channel.GuildID], channel.ID)
	} else if channel.IsDM() {
		c.DMChannels = append(c.DMChannels, channel.ID)
	}
	c.mu.Unlock()

	if c.onChannelCreate != nil {
		c.onChannelCreate(&channel)
	}
}

func (c *Cache) handleChannelUpdate(data json.RawMessage) {
	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return
	}

	c.mu.Lock()
	c.Channels[channel.ID] = &channel
	c.mu.Unlock()
}

func (c *Cache) handleChannelDelete(data json.RawMessage) {
	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return
	}

	c.mu.Lock()
	delete(c.Channels, channel.ID)
	// Remove from guild channels list
	if channel.GuildID != "" {
		if channels, ok := c.GuildChannels[channel.GuildID]; ok {
			for i, id := range channels {
				if id == channel.ID {
					c.GuildChannels[channel.GuildID] = append(channels[:i], channels[i+1:]...)
					break
				}
			}
		}
	}
	c.mu.Unlock()

	if c.onChannelDelete != nil {
		c.onChannelDelete(&channel)
	}
}

func (c *Cache) handleMessageCreate(data json.RawMessage) {
	var message models.Message
	if err := json.Unmarshal(data, &message); err != nil {
		return
	}

	c.mu.Lock()
	c.Messages[message.ChannelID] = append(c.Messages[message.ChannelID], &message)
	// Keep only last 100 messages per channel
	if len(c.Messages[message.ChannelID]) > 100 {
		c.Messages[message.ChannelID] = c.Messages[message.ChannelID][1:]
	}
	c.mu.Unlock()

	if c.onMessageCreate != nil {
		c.onMessageCreate(&message)
	}
}

func (c *Cache) handleMessageUpdate(data json.RawMessage) {
	var message models.Message
	if err := json.Unmarshal(data, &message); err != nil {
		return
	}

	c.mu.Lock()
	if messages, ok := c.Messages[message.ChannelID]; ok {
		for i, msg := range messages {
			if msg.ID == message.ID {
				c.Messages[message.ChannelID][i] = &message
				break
			}
		}
	}
	c.mu.Unlock()

	if c.onMessageUpdate != nil {
		c.onMessageUpdate(&message)
	}
}

func (c *Cache) handleMessageDelete(data json.RawMessage) {
	var del gateway.MessageDeleteData
	if err := json.Unmarshal(data, &del); err != nil {
		return
	}

	c.mu.Lock()
	if messages, ok := c.Messages[del.ChannelID]; ok {
		for i, msg := range messages {
			if msg.ID == del.ID {
				c.Messages[del.ChannelID] = append(messages[:i], messages[i+1:]...)
				break
			}
		}
	}
	c.mu.Unlock()

	if c.onMessageDelete != nil {
		c.onMessageDelete(del.ChannelID, del.ID)
	}
}

func (c *Cache) handleTypingStart(data json.RawMessage) {
	var typing models.TypingStart
	if err := json.Unmarshal(data, &typing); err != nil {
		return
	}

	c.mu.Lock()
	if _, ok := c.Typing[typing.ChannelID]; !ok {
		c.Typing[typing.ChannelID] = make(map[string]int64)
	}
	c.Typing[typing.ChannelID][typing.UserID] = typing.Timestamp
	c.mu.Unlock()

	if c.onTypingStart != nil {
		c.onTypingStart(&typing)
	}
}

func (c *Cache) handlePresenceUpdate(data json.RawMessage) {
	var presence models.Presence
	if err := json.Unmarshal(data, &presence); err != nil {
		return
	}

	c.mu.Lock()
	if presence.User != nil {
		c.Presences[presence.User.ID] = &presence
	}
	c.mu.Unlock()

	if c.onPresenceUpdate != nil {
		c.onPresenceUpdate(&presence)
	}
}

func (c *Cache) handleVoiceStateUpdate(data json.RawMessage) {
	var vs models.VoiceState
	if err := json.Unmarshal(data, &vs); err != nil {
		return
	}

	c.mu.Lock()
	if vs.ChannelID == "" {
		delete(c.VoiceStates, vs.UserID)
	} else {
		c.VoiceStates[vs.UserID] = &vs
	}
	c.mu.Unlock()

	if c.onVoiceStateUpdate != nil {
		c.onVoiceStateUpdate(&vs)
	}
}

func (c *Cache) handleVoiceServerUpdate(data json.RawMessage) {
	var vs gateway.VoiceServerUpdateData
	if err := json.Unmarshal(data, &vs); err != nil {
		return
	}

	if c.onVoiceServerUpdate != nil {
		c.onVoiceServerUpdate(&vs)
	}
}

func (c *Cache) handleRelationshipAdd(data json.RawMessage) {
	var rel models.Relationship
	if err := json.Unmarshal(data, &rel); err != nil {
		return
	}

	c.mu.Lock()
	c.Relationships[rel.ID] = &rel
	c.mu.Unlock()

	if c.onRelationshipAdd != nil {
		c.onRelationshipAdd(&rel)
	}
}

func (c *Cache) handleRelationshipRemove(data json.RawMessage) {
	var rel gateway.RelationshipRemoveData
	if err := json.Unmarshal(data, &rel); err != nil {
		return
	}

	c.mu.Lock()
	delete(c.Relationships, rel.ID)
	c.mu.Unlock()

	if c.onRelationshipRemove != nil {
		c.onRelationshipRemove(rel.ID)
	}
}

func (c *Cache) handleUserUpdate(data json.RawMessage) {
	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return
	}

	c.mu.Lock()
	c.User = &user
	c.mu.Unlock()
}

// Getters

func (c *Cache) GetUser() *models.User {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.User
}

func (c *Cache) GetGuild(id string) *models.Guild {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Guilds[id]
}

func (c *Cache) GetGuilds() []*models.Guild {
	c.mu.RLock()
	defer c.mu.RUnlock()
	guilds := make([]*models.Guild, 0, len(c.Guilds))
	for _, g := range c.Guilds {
		guilds = append(guilds, g)
	}
	return guilds
}

func (c *Cache) GetChannel(id string) *models.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Channels[id]
}

func (c *Cache) GetGuildChannels(guildID string) []*models.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ids, ok := c.GuildChannels[guildID]
	if !ok {
		return nil
	}
	channels := make([]*models.Channel, 0, len(ids))
	for _, id := range ids {
		if ch, ok := c.Channels[id]; ok {
			channels = append(channels, ch)
		}
	}
	return channels
}

func (c *Cache) GetDMChannels() []*models.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	channels := make([]*models.Channel, 0, len(c.DMChannels))
	for _, id := range c.DMChannels {
		if ch, ok := c.Channels[id]; ok {
			channels = append(channels, ch)
		}
	}
	return channels
}

func (c *Cache) GetMessages(channelID string) []*models.Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Messages[channelID]
}

func (c *Cache) GetRelationships() []*models.Relationship {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rels := make([]*models.Relationship, 0, len(c.Relationships))
	for _, r := range c.Relationships {
		rels = append(rels, r)
	}
	return rels
}

func (c *Cache) GetFriends() []*models.Relationship {
	c.mu.RLock()
	defer c.mu.RUnlock()
	friends := make([]*models.Relationship, 0)
	for _, r := range c.Relationships {
		if r.Type == models.RelationshipTypeFriend {
			friends = append(friends, r)
		}
	}
	return friends
}

func (c *Cache) GetPresence(userID string) *models.Presence {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Presences[userID]
}

func (c *Cache) GetVoiceState(userID string) *models.VoiceState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.VoiceStates[userID]
}

func (c *Cache) AddMessages(channelID string, messages []*models.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Prepend messages (they're usually older)
	c.Messages[channelID] = append(messages, c.Messages[channelID]...)
}
