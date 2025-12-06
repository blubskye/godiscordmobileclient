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

  
package gateway

import (
	"encoding/json"

	"github.com/blubskye/godiscordmobileclient/internal/models"
)

// Opcode represents gateway opcodes
type Opcode int

const (
	OpcodeDispatch            Opcode = 0  // Receive: An event was dispatched
	OpcodeHeartbeat           Opcode = 1  // Send/Receive: Heartbeat
	OpcodeIdentify            Opcode = 2  // Send: Identify
	OpcodePresenceUpdate      Opcode = 3  // Send: Update presence
	OpcodeVoiceStateUpdate    Opcode = 4  // Send: Join/leave/move voice
	OpcodeResume              Opcode = 6  // Send: Resume connection
	OpcodeReconnect           Opcode = 7  // Receive: Server wants reconnect
	OpcodeRequestGuildMembers Opcode = 8  // Send: Request guild members
	OpcodeInvalidSession      Opcode = 9  // Receive: Session invalidated
	OpcodeHello               Opcode = 10 // Receive: Hello with heartbeat interval
	OpcodeHeartbeatACK        Opcode = 11 // Receive: Heartbeat acknowledged
)

// GatewayPayload is the base structure for all gateway messages
type GatewayPayload struct {
	Op       Opcode          `json:"op"`
	Data     json.RawMessage `json:"d,omitempty"`
	Sequence *int64          `json:"s,omitempty"`
	Type     string          `json:"t,omitempty"`
}

// HelloData is sent with Opcode 10
type HelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// IdentifyData is sent with Opcode 2
type IdentifyData struct {
	Token          string             `json:"token"`
	Properties     IdentifyProperties `json:"properties"`
	Compress       bool               `json:"compress,omitempty"`
	LargeThreshold int                `json:"large_threshold,omitempty"`
	Shard          []int              `json:"shard,omitempty"`
	Presence       *PresenceUpdate    `json:"presence,omitempty"`
	Capabilities   int                `json:"capabilities,omitempty"`
}

// IdentifyProperties for client identification
type IdentifyProperties struct {
	OS              string `json:"os"`
	Browser         string `json:"browser"`
	Device          string `json:"device"`
	SystemLocale    string `json:"system_locale,omitempty"`
	BrowserVersion  string `json:"browser_version,omitempty"`
	OSVersion       string `json:"os_version,omitempty"`
	Referrer        string `json:"referrer,omitempty"`
	ReferringDomain string `json:"referring_domain,omitempty"`
}

// PresenceUpdate for updating presence
type PresenceUpdate struct {
	Since      *int64            `json:"since"`
	Activities []models.Activity `json:"activities"`
	Status     models.Status     `json:"status"`
	AFK        bool              `json:"afk"`
}

// ResumeData is sent with Opcode 6
type ResumeData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Sequence  int64  `json:"seq"`
}

// VoiceStateUpdateData is sent with Opcode 4
type VoiceStateUpdateData struct {
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id,omitempty"` // null to disconnect
	SelfMute  bool   `json:"self_mute"`
	SelfDeaf  bool   `json:"self_deaf"`
}

// RequestGuildMembersData is sent with Opcode 8
type RequestGuildMembersData struct {
	GuildID   string   `json:"guild_id"`
	Query     string   `json:"query,omitempty"`
	Limit     int      `json:"limit"`
	Presences bool     `json:"presences,omitempty"`
	UserIDs   []string `json:"user_ids,omitempty"`
	Nonce     string   `json:"nonce,omitempty"`
}

// ReadyData is received with READY event
type ReadyData struct {
	Version         int                     `json:"v"`
	User            *models.User            `json:"user"`
	Guilds          []models.UnavailableGuild `json:"guilds"`
	SessionID       string                  `json:"session_id"`
	ResumeGatewayURL string                 `json:"resume_gateway_url"`
	Shard           []int                   `json:"shard,omitempty"`
	Application     *ApplicationData        `json:"application,omitempty"`

	// User-specific fields
	PrivateChannels   []models.Channel      `json:"private_channels,omitempty"`
	Relationships     []models.Relationship `json:"relationships,omitempty"`
	UserSettings      json.RawMessage       `json:"user_settings,omitempty"`
	ReadState         json.RawMessage       `json:"read_state,omitempty"`
	GuildJoinRequests json.RawMessage       `json:"guild_join_requests,omitempty"`
	Presences         []models.Presence     `json:"presences,omitempty"`
}

type ApplicationData struct {
	ID    string `json:"id"`
	Flags int    `json:"flags"`
}

// GuildCreateData is received with GUILD_CREATE event
type GuildCreateData = models.Guild

// GuildDeleteData is received with GUILD_DELETE event
type GuildDeleteData = models.UnavailableGuild

// MessageCreateData is received with MESSAGE_CREATE event
type MessageCreateData = models.Message

// MessageUpdateData is received with MESSAGE_UPDATE event
type MessageUpdateData = models.Message

// MessageDeleteData is received with MESSAGE_DELETE event
type MessageDeleteData struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id,omitempty"`
}

// MessageDeleteBulkData is received with MESSAGE_DELETE_BULK event
type MessageDeleteBulkData struct {
	IDs       []string `json:"ids"`
	ChannelID string   `json:"channel_id"`
	GuildID   string   `json:"guild_id,omitempty"`
}

// ChannelCreateData is received with CHANNEL_CREATE event
type ChannelCreateData = models.Channel

// ChannelUpdateData is received with CHANNEL_UPDATE event
type ChannelUpdateData = models.Channel

// ChannelDeleteData is received with CHANNEL_DELETE event
type ChannelDeleteData = models.Channel

// TypingStartData is received with TYPING_START event
type TypingStartData = models.TypingStart

// PresenceUpdateData is received with PRESENCE_UPDATE event
type PresenceUpdateData = models.Presence

// VoiceStateUpdateEventData is received with VOICE_STATE_UPDATE event
type VoiceStateUpdateEventData = models.VoiceState

// VoiceServerUpdateData is received with VOICE_SERVER_UPDATE event
type VoiceServerUpdateData struct {
	Token    string `json:"token"`
	GuildID  string `json:"guild_id"`
	Endpoint string `json:"endpoint"`
}

// MemberAddData is received with GUILD_MEMBER_ADD event
type MemberAddData struct {
	models.Member
	GuildID string `json:"guild_id"`
}

// MemberRemoveData is received with GUILD_MEMBER_REMOVE event
type MemberRemoveData struct {
	GuildID string       `json:"guild_id"`
	User    *models.User `json:"user"`
}

// MemberUpdateData is received with GUILD_MEMBER_UPDATE event
type MemberUpdateData struct {
	GuildID                    string       `json:"guild_id"`
	Roles                      []string     `json:"roles"`
	User                       *models.User `json:"user"`
	Nick                       string       `json:"nick,omitempty"`
	Avatar                     string       `json:"avatar,omitempty"`
	JoinedAt                   string       `json:"joined_at,omitempty"`
	PremiumSince               string       `json:"premium_since,omitempty"`
	Deaf                       bool         `json:"deaf,omitempty"`
	Mute                       bool         `json:"mute,omitempty"`
	Pending                    bool         `json:"pending,omitempty"`
	CommunicationDisabledUntil string       `json:"communication_disabled_until,omitempty"`
}

// GuildMembersChunkData is received with GUILD_MEMBERS_CHUNK event
type GuildMembersChunkData struct {
	GuildID    string            `json:"guild_id"`
	Members    []models.Member   `json:"members"`
	ChunkIndex int               `json:"chunk_index"`
	ChunkCount int               `json:"chunk_count"`
	NotFound   []string          `json:"not_found,omitempty"`
	Presences  []models.Presence `json:"presences,omitempty"`
	Nonce      string            `json:"nonce,omitempty"`
}

// RelationshipAddData is received with RELATIONSHIP_ADD event
type RelationshipAddData = models.Relationship

// RelationshipRemoveData is received with RELATIONSHIP_REMOVE event
type RelationshipRemoveData struct {
	ID   string `json:"id"`
	Type int    `json:"type"`
}

// ReactionAddData is received with MESSAGE_REACTION_ADD event
type ReactionAddData struct {
	UserID          string              `json:"user_id"`
	ChannelID       string              `json:"channel_id"`
	MessageID       string              `json:"message_id"`
	GuildID         string              `json:"guild_id,omitempty"`
	Member          *models.Member      `json:"member,omitempty"`
	Emoji           *models.ReactionEmoji `json:"emoji"`
	MessageAuthorID string              `json:"message_author_id,omitempty"`
	Burst           bool                `json:"burst"`
	BurstColors     []string            `json:"burst_colors,omitempty"`
}

// ReactionRemoveData is received with MESSAGE_REACTION_REMOVE event
type ReactionRemoveData struct {
	UserID    string              `json:"user_id"`
	ChannelID string              `json:"channel_id"`
	MessageID string              `json:"message_id"`
	GuildID   string              `json:"guild_id,omitempty"`
	Emoji     *models.ReactionEmoji `json:"emoji"`
	Burst     bool                `json:"burst"`
}

// Event names
const (
	EventReady                  = "READY"
	EventResumed                = "RESUMED"
	EventGuildCreate            = "GUILD_CREATE"
	EventGuildUpdate            = "GUILD_UPDATE"
	EventGuildDelete            = "GUILD_DELETE"
	EventChannelCreate          = "CHANNEL_CREATE"
	EventChannelUpdate          = "CHANNEL_UPDATE"
	EventChannelDelete          = "CHANNEL_DELETE"
	EventMessageCreate          = "MESSAGE_CREATE"
	EventMessageUpdate          = "MESSAGE_UPDATE"
	EventMessageDelete          = "MESSAGE_DELETE"
	EventMessageDeleteBulk      = "MESSAGE_DELETE_BULK"
	EventMessageReactionAdd     = "MESSAGE_REACTION_ADD"
	EventMessageReactionRemove  = "MESSAGE_REACTION_REMOVE"
	EventTypingStart            = "TYPING_START"
	EventPresenceUpdate         = "PRESENCE_UPDATE"
	EventVoiceStateUpdate       = "VOICE_STATE_UPDATE"
	EventVoiceServerUpdate      = "VOICE_SERVER_UPDATE"
	EventGuildMemberAdd         = "GUILD_MEMBER_ADD"
	EventGuildMemberRemove      = "GUILD_MEMBER_REMOVE"
	EventGuildMemberUpdate      = "GUILD_MEMBER_UPDATE"
	EventGuildMembersChunk      = "GUILD_MEMBERS_CHUNK"
	EventRelationshipAdd        = "RELATIONSHIP_ADD"
	EventRelationshipRemove     = "RELATIONSHIP_REMOVE"
	EventUserUpdate             = "USER_UPDATE"
)
