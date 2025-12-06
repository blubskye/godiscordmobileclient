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

package models

// Guild represents a Discord server
type Guild struct {
	ID                          string         `json:"id"`
	Name                        string         `json:"name"`
	Icon                        string         `json:"icon,omitempty"`
	IconHash                    string         `json:"icon_hash,omitempty"`
	Splash                      string         `json:"splash,omitempty"`
	DiscoverySplash             string         `json:"discovery_splash,omitempty"`
	Owner                       bool           `json:"owner,omitempty"`
	OwnerID                     string         `json:"owner_id"`
	Permissions                 string         `json:"permissions,omitempty"`
	Region                      string         `json:"region,omitempty"`
	AFKChannelID                string         `json:"afk_channel_id,omitempty"`
	AFKTimeout                  int            `json:"afk_timeout"`
	WidgetEnabled               bool           `json:"widget_enabled,omitempty"`
	WidgetChannelID             string         `json:"widget_channel_id,omitempty"`
	VerificationLevel           int            `json:"verification_level"`
	DefaultMessageNotifications int            `json:"default_message_notifications"`
	ExplicitContentFilter       int            `json:"explicit_content_filter"`
	Roles                       []Role         `json:"roles"`
	Emojis                      []Emoji        `json:"emojis"`
	Features                    []string       `json:"features"`
	MFALevel                    int            `json:"mfa_level"`
	ApplicationID               string         `json:"application_id,omitempty"`
	SystemChannelID             string         `json:"system_channel_id,omitempty"`
	SystemChannelFlags          int            `json:"system_channel_flags"`
	RulesChannelID              string         `json:"rules_channel_id,omitempty"`
	MaxPresences                int            `json:"max_presences,omitempty"`
	MaxMembers                  int            `json:"max_members,omitempty"`
	VanityURLCode               string         `json:"vanity_url_code,omitempty"`
	Description                 string         `json:"description,omitempty"`
	Banner                      string         `json:"banner,omitempty"`
	PremiumTier                 int            `json:"premium_tier"`
	PremiumSubscriptionCount    int            `json:"premium_subscription_count,omitempty"`
	PreferredLocale             string         `json:"preferred_locale"`
	PublicUpdatesChannelID      string         `json:"public_updates_channel_id,omitempty"`
	MaxVideoChannelUsers        int            `json:"max_video_channel_users,omitempty"`
	MaxStageVideoChannelUsers   int            `json:"max_stage_video_channel_users,omitempty"`
	ApproximateMemberCount      int            `json:"approximate_member_count,omitempty"`
	ApproximatePresenceCount    int            `json:"approximate_presence_count,omitempty"`
	WelcomeScreen               *WelcomeScreen `json:"welcome_screen,omitempty"`
	NSFWLevel                   int            `json:"nsfw_level"`
	Stickers                    []Sticker      `json:"stickers,omitempty"`
	PremiumProgressBarEnabled   bool           `json:"premium_progress_bar_enabled"`
	SafetyAlertsChannelID       string         `json:"safety_alerts_channel_id,omitempty"`

	// These fields are only sent in GUILD_CREATE
	JoinedAt    string       `json:"joined_at,omitempty"`
	Large       bool         `json:"large,omitempty"`
	Unavailable bool         `json:"unavailable,omitempty"`
	MemberCount int          `json:"member_count,omitempty"`
	VoiceStates []VoiceState `json:"voice_states,omitempty"`
	Members     []Member     `json:"members,omitempty"`
	Channels    []Channel    `json:"channels,omitempty"`
	Threads     []Channel    `json:"threads,omitempty"`
	Presences   []Presence   `json:"presences,omitempty"`
}

// IconURL returns the URL for the guild's icon
func (g *Guild) IconURL(size int) string {
	if g.Icon == "" {
		return ""
	}
	ext := "png"
	if len(g.Icon) > 2 && g.Icon[:2] == "a_" {
		ext = "gif"
	}
	return "https://cdn.discordapp.com/icons/" + g.ID + "/" + g.Icon + "." + ext + "?size=" + itoa(size)
}

// Role represents a Discord role
type Role struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Color        int    `json:"color"`
	Hoist        bool   `json:"hoist"`
	Icon         string `json:"icon,omitempty"`
	UnicodeEmoji string `json:"unicode_emoji,omitempty"`
	Position     int    `json:"position"`
	Permissions  string `json:"permissions"`
	Managed      bool   `json:"managed"`
	Mentionable  bool   `json:"mentionable"`
	Flags        int    `json:"flags"`
}

// Emoji represents a custom emoji
type Emoji struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Roles         []string `json:"roles,omitempty"`
	User          *User    `json:"user,omitempty"`
	RequireColons bool     `json:"require_colons,omitempty"`
	Managed       bool     `json:"managed,omitempty"`
	Animated      bool     `json:"animated,omitempty"`
	Available     bool     `json:"available,omitempty"`
}

// Sticker represents a sticker
type Sticker struct {
	ID          string `json:"id"`
	PackID      string `json:"pack_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Tags        string `json:"tags"`
	Type        int    `json:"type"`
	FormatType  int    `json:"format_type"`
	Available   bool   `json:"available,omitempty"`
	GuildID     string `json:"guild_id,omitempty"`
	User        *User  `json:"user,omitempty"`
	SortValue   int    `json:"sort_value,omitempty"`
}

// Member represents a guild member
type Member struct {
	User                       *User    `json:"user,omitempty"`
	Nick                       string   `json:"nick,omitempty"`
	Avatar                     string   `json:"avatar,omitempty"`
	Roles                      []string `json:"roles"`
	JoinedAt                   string   `json:"joined_at"`
	PremiumSince               string   `json:"premium_since,omitempty"`
	Deaf                       bool     `json:"deaf"`
	Mute                       bool     `json:"mute"`
	Flags                      int      `json:"flags"`
	Pending                    bool     `json:"pending,omitempty"`
	Permissions                string   `json:"permissions,omitempty"`
	CommunicationDisabledUntil string   `json:"communication_disabled_until,omitempty"`
}

// DisplayName returns the member's display name (nickname or username)
func (m *Member) DisplayName() string {
	if m.Nick != "" {
		return m.Nick
	}
	if m.User != nil {
		return m.User.DisplayName()
	}
	return ""
}

// VoiceState represents a user's voice connection status
type VoiceState struct {
	GuildID                 string  `json:"guild_id,omitempty"`
	ChannelID               string  `json:"channel_id,omitempty"`
	UserID                  string  `json:"user_id"`
	Member                  *Member `json:"member,omitempty"`
	SessionID               string  `json:"session_id"`
	Deaf                    bool    `json:"deaf"`
	Mute                    bool    `json:"mute"`
	SelfDeaf                bool    `json:"self_deaf"`
	SelfMute                bool    `json:"self_mute"`
	SelfStream              bool    `json:"self_stream,omitempty"`
	SelfVideo               bool    `json:"self_video"`
	Suppress                bool    `json:"suppress"`
	RequestToSpeakTimestamp string  `json:"request_to_speak_timestamp,omitempty"`
}

// WelcomeScreen represents the welcome screen shown to new members
type WelcomeScreen struct {
	Description     string                 `json:"description,omitempty"`
	WelcomeChannels []WelcomeScreenChannel `json:"welcome_channels"`
}

type WelcomeScreenChannel struct {
	ChannelID   string `json:"channel_id"`
	Description string `json:"description"`
	EmojiID     string `json:"emoji_id,omitempty"`
	EmojiName   string `json:"emoji_name,omitempty"`
}

// UnavailableGuild is sent when a guild becomes unavailable
type UnavailableGuild struct {
	ID          string `json:"id"`
	Unavailable bool   `json:"unavailable"`
}
