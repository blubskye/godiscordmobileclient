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

import "time"

// Message represents a Discord message
type Message struct {
	ID                   string             `json:"id"`
	ChannelID            string             `json:"channel_id"`
	Author               *User              `json:"author,omitempty"`
	Content              string             `json:"content"`
	Timestamp            time.Time          `json:"timestamp"`
	EditedTimestamp      *time.Time         `json:"edited_timestamp,omitempty"`
	TTS                  bool               `json:"tts"`
	MentionEveryone      bool               `json:"mention_everyone"`
	Mentions             []User             `json:"mentions"`
	MentionRoles         []string           `json:"mention_roles"`
	MentionChannels      []ChannelMention   `json:"mention_channels,omitempty"`
	Attachments          []Attachment       `json:"attachments"`
	Embeds               []Embed            `json:"embeds"`
	Reactions            []Reaction         `json:"reactions,omitempty"`
	Nonce                interface{}        `json:"nonce,omitempty"` // string or int
	Pinned               bool               `json:"pinned"`
	WebhookID            string             `json:"webhook_id,omitempty"`
	Type                 MessageType        `json:"type"`
	Activity             *MessageActivity   `json:"activity,omitempty"`
	Application          *Application       `json:"application,omitempty"`
	ApplicationID        string             `json:"application_id,omitempty"`
	MessageReference     *MessageReference  `json:"message_reference,omitempty"`
	Flags                int                `json:"flags,omitempty"`
	ReferencedMessage    *Message           `json:"referenced_message,omitempty"`
	Interaction          *MessageInteraction `json:"interaction,omitempty"`
	Thread               *Channel           `json:"thread,omitempty"`
	Components           []Component        `json:"components,omitempty"`
	StickerItems         []StickerItem      `json:"sticker_items,omitempty"`
	Position             int                `json:"position,omitempty"`
	RoleSubscriptionData *RoleSubscription  `json:"role_subscription_data,omitempty"`

	// Only for MESSAGE_CREATE event
	GuildID string  `json:"guild_id,omitempty"`
	Member  *Member `json:"member,omitempty"`
}

// MessageType represents the type of message
type MessageType int

const (
	MessageTypeDefault                                 MessageType = 0
	MessageTypeRecipientAdd                            MessageType = 1
	MessageTypeRecipientRemove                         MessageType = 2
	MessageTypeCall                                    MessageType = 3
	MessageTypeChannelNameChange                       MessageType = 4
	MessageTypeChannelIconChange                       MessageType = 5
	MessageTypeChannelPinnedMessage                    MessageType = 6
	MessageTypeUserJoin                                MessageType = 7
	MessageTypeGuildBoost                              MessageType = 8
	MessageTypeGuildBoostTier1                         MessageType = 9
	MessageTypeGuildBoostTier2                         MessageType = 10
	MessageTypeGuildBoostTier3                         MessageType = 11
	MessageTypeChannelFollowAdd                        MessageType = 12
	MessageTypeGuildDiscoveryDisqualified              MessageType = 14
	MessageTypeGuildDiscoveryRequalified               MessageType = 15
	MessageTypeGuildDiscoveryGracePeriodInitialWarning MessageType = 16
	MessageTypeGuildDiscoveryGracePeriodFinalWarning   MessageType = 17
	MessageTypeThreadCreated                           MessageType = 18
	MessageTypeReply                                   MessageType = 19
	MessageTypeChatInputCommand                        MessageType = 20
	MessageTypeThreadStarterMessage                    MessageType = 21
	MessageTypeGuildInviteReminder                     MessageType = 22
	MessageTypeContextMenuCommand                      MessageType = 23
	MessageTypeAutoModerationAction                    MessageType = 24
	MessageTypeRoleSubscriptionPurchase                MessageType = 25
	MessageTypeInteractionPremiumUpsell                MessageType = 26
	MessageTypeStageStart                              MessageType = 27
	MessageTypeStageEnd                                MessageType = 28
	MessageTypeStageSpeaker                            MessageType = 29
	MessageTypeStageTopic                              MessageType = 31
	MessageTypeGuildApplicationPremiumSubscription     MessageType = 32
)

// IsSystem returns true if this is a system message (not user-generated)
func (m *Message) IsSystem() bool {
	return m.Type != MessageTypeDefault && m.Type != MessageTypeReply
}

// ChannelMention represents a mentioned channel
type ChannelMention struct {
	ID      string      `json:"id"`
	GuildID string      `json:"guild_id"`
	Type    ChannelType `json:"type"`
	Name    string      `json:"name"`
}

// Attachment represents a message attachment
type Attachment struct {
	ID           string  `json:"id"`
	Filename     string  `json:"filename"`
	Description  string  `json:"description,omitempty"`
	ContentType  string  `json:"content_type,omitempty"`
	Size         int     `json:"size"`
	URL          string  `json:"url"`
	ProxyURL     string  `json:"proxy_url"`
	Height       int     `json:"height,omitempty"`
	Width        int     `json:"width,omitempty"`
	Ephemeral    bool    `json:"ephemeral,omitempty"`
	DurationSecs float64 `json:"duration_secs,omitempty"`
	Waveform     string  `json:"waveform,omitempty"`
	Flags        int     `json:"flags,omitempty"`
}

// IsImage returns true if attachment is an image
func (a *Attachment) IsImage() bool {
	switch a.ContentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	}
	return false
}

// IsVideo returns true if attachment is a video
func (a *Attachment) IsVideo() bool {
	switch a.ContentType {
	case "video/mp4", "video/webm", "video/quicktime":
		return true
	}
	return false
}

// Embed represents a rich embed
type Embed struct {
	Title       string          `json:"title,omitempty"`
	Type        string          `json:"type,omitempty"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url,omitempty"`
	Timestamp   string          `json:"timestamp,omitempty"`
	Color       int             `json:"color,omitempty"`
	Footer      *EmbedFooter    `json:"footer,omitempty"`
	Image       *EmbedImage     `json:"image,omitempty"`
	Thumbnail   *EmbedThumbnail `json:"thumbnail,omitempty"`
	Video       *EmbedVideo     `json:"video,omitempty"`
	Provider    *EmbedProvider  `json:"provider,omitempty"`
	Author      *EmbedAuthor    `json:"author,omitempty"`
	Fields      []EmbedField    `json:"fields,omitempty"`
}

type EmbedFooter struct {
	Text         string `json:"text"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

type EmbedImage struct {
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

type EmbedThumbnail struct {
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

type EmbedVideo struct {
	URL      string `json:"url,omitempty"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

type EmbedProvider struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type EmbedAuthor struct {
	Name         string `json:"name"`
	URL          string `json:"url,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// Reaction represents a message reaction
type Reaction struct {
	Count        int           `json:"count"`
	CountDetails *CountDetails `json:"count_details,omitempty"`
	Me           bool          `json:"me"`
	MeBurst      bool          `json:"me_burst"`
	Emoji        *ReactionEmoji `json:"emoji"`
	BurstColors  []string      `json:"burst_colors,omitempty"`
}

type CountDetails struct {
	Burst  int `json:"burst"`
	Normal int `json:"normal"`
}

type ReactionEmoji struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Animated bool   `json:"animated,omitempty"`
}

// MessageActivity for Rich Presence
type MessageActivity struct {
	Type    int    `json:"type"`
	PartyID string `json:"party_id,omitempty"`
}

// Application for Rich Presence
type Application struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image,omitempty"`
}

// MessageReference for replies and forwards
type MessageReference struct {
	MessageID       string `json:"message_id,omitempty"`
	ChannelID       string `json:"channel_id,omitempty"`
	GuildID         string `json:"guild_id,omitempty"`
	FailIfNotExists bool   `json:"fail_if_not_exists,omitempty"`
}

// MessageInteraction for slash command responses
type MessageInteraction struct {
	ID     string  `json:"id"`
	Type   int     `json:"type"`
	Name   string  `json:"name"`
	User   *User   `json:"user"`
	Member *Member `json:"member,omitempty"`
}

// Component for buttons/selects
type Component struct {
	Type        int         `json:"type"`
	CustomID    string      `json:"custom_id,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Style       int         `json:"style,omitempty"`
	Label       string      `json:"label,omitempty"`
	Emoji       *ReactionEmoji `json:"emoji,omitempty"`
	URL         string      `json:"url,omitempty"`
	Options     []SelectOption `json:"options,omitempty"`
	Placeholder string      `json:"placeholder,omitempty"`
	MinValues   int         `json:"min_values,omitempty"`
	MaxValues   int         `json:"max_values,omitempty"`
	Components  []Component `json:"components,omitempty"`
	MinLength   int         `json:"min_length,omitempty"`
	MaxLength   int         `json:"max_length,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Value       string      `json:"value,omitempty"`
}

type SelectOption struct {
	Label       string        `json:"label"`
	Value       string        `json:"value"`
	Description string        `json:"description,omitempty"`
	Emoji       *ReactionEmoji `json:"emoji,omitempty"`
	Default     bool          `json:"default,omitempty"`
}

// StickerItem is a partial sticker in messages
type StickerItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	FormatType int    `json:"format_type"`
}

// RoleSubscription for role subscription messages
type RoleSubscription struct {
	RoleSubscriptionListingID string `json:"role_subscription_listing_id"`
	TierName                  string `json:"tier_name"`
	TotalMonthsSubscribed     int    `json:"total_months_subscribed"`
	IsRenewal                 bool   `json:"is_renewal"`
}

// CreateMessage is the payload for sending messages
type CreateMessage struct {
	Content          string            `json:"content,omitempty"`
	Nonce            string            `json:"nonce,omitempty"`
	TTS              bool              `json:"tts,omitempty"`
	Embeds           []Embed           `json:"embeds,omitempty"`
	AllowedMentions  *AllowedMentions  `json:"allowed_mentions,omitempty"`
	MessageReference *MessageReference `json:"message_reference,omitempty"`
	Components       []Component       `json:"components,omitempty"`
	StickerIDs       []string          `json:"sticker_ids,omitempty"`
	Flags            int               `json:"flags,omitempty"`
}

// AllowedMentions controls who gets pinged
type AllowedMentions struct {
	Parse       []string `json:"parse,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Users       []string `json:"users,omitempty"`
	RepliedUser bool     `json:"replied_user,omitempty"`
}
