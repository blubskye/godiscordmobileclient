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

// Channel represents a Discord channel
type Channel struct {
	ID                            string                 `json:"id"`
	Type                          ChannelType            `json:"type"`
	GuildID                       string                 `json:"guild_id,omitempty"`
	Position                      int                    `json:"position,omitempty"`
	PermissionOverwrites          []PermissionOverwrite  `json:"permission_overwrites,omitempty"`
	Name                          string                 `json:"name,omitempty"`
	Topic                         string                 `json:"topic,omitempty"`
	NSFW                          bool                   `json:"nsfw,omitempty"`
	LastMessageID                 string                 `json:"last_message_id,omitempty"`
	Bitrate                       int                    `json:"bitrate,omitempty"`
	UserLimit                     int                    `json:"user_limit,omitempty"`
	RateLimitPerUser              int                    `json:"rate_limit_per_user,omitempty"`
	Recipients                    []User                 `json:"recipients,omitempty"`
	Icon                          string                 `json:"icon,omitempty"`
	OwnerID                       string                 `json:"owner_id,omitempty"`
	ApplicationID                 string                 `json:"application_id,omitempty"`
	Managed                       bool                   `json:"managed,omitempty"`
	ParentID                      string                 `json:"parent_id,omitempty"`
	LastPinTimestamp              string                 `json:"last_pin_timestamp,omitempty"`
	RTCRegion                     string                 `json:"rtc_region,omitempty"`
	VideoQualityMode              int                    `json:"video_quality_mode,omitempty"`
	MessageCount                  int                    `json:"message_count,omitempty"`
	MemberCount                   int                    `json:"member_count,omitempty"`
	ThreadMetadata                *ThreadMetadata        `json:"thread_metadata,omitempty"`
	Member                        *ThreadMember          `json:"member,omitempty"`
	DefaultAutoArchiveDuration    int                    `json:"default_auto_archive_duration,omitempty"`
	Permissions                   string                 `json:"permissions,omitempty"`
	Flags                         int                    `json:"flags,omitempty"`
	TotalMessageSent              int                    `json:"total_message_sent,omitempty"`
	AvailableTags                 []ForumTag             `json:"available_tags,omitempty"`
	AppliedTags                   []string               `json:"applied_tags,omitempty"`
	DefaultReactionEmoji          *DefaultReaction       `json:"default_reaction_emoji,omitempty"`
	DefaultThreadRateLimitPerUser int                    `json:"default_thread_rate_limit_per_user,omitempty"`
	DefaultSortOrder              int                    `json:"default_sort_order,omitempty"`
	DefaultForumLayout            int                    `json:"default_forum_layout,omitempty"`
}

// ChannelType represents the type of channel
type ChannelType int

const (
	ChannelTypeGuildText          ChannelType = 0
	ChannelTypeDM                 ChannelType = 1
	ChannelTypeGuildVoice         ChannelType = 2
	ChannelTypeGroupDM            ChannelType = 3
	ChannelTypeGuildCategory      ChannelType = 4
	ChannelTypeGuildAnnouncement  ChannelType = 5
	ChannelTypeAnnouncementThread ChannelType = 10
	ChannelTypePublicThread       ChannelType = 11
	ChannelTypePrivateThread      ChannelType = 12
	ChannelTypeGuildStageVoice    ChannelType = 13
	ChannelTypeGuildDirectory     ChannelType = 14
	ChannelTypeGuildForum         ChannelType = 15
	ChannelTypeGuildMedia         ChannelType = 16
)

// IsText returns true if this is a text-based channel
func (c *Channel) IsText() bool {
	switch c.Type {
	case ChannelTypeGuildText, ChannelTypeDM, ChannelTypeGroupDM,
		ChannelTypeGuildAnnouncement, ChannelTypeAnnouncementThread,
		ChannelTypePublicThread, ChannelTypePrivateThread:
		return true
	}
	return false
}

// IsVoice returns true if this is a voice channel
func (c *Channel) IsVoice() bool {
	return c.Type == ChannelTypeGuildVoice || c.Type == ChannelTypeGuildStageVoice
}

// IsDM returns true if this is a DM or group DM
func (c *Channel) IsDM() bool {
	return c.Type == ChannelTypeDM || c.Type == ChannelTypeGroupDM
}

// IsCategory returns true if this is a category
func (c *Channel) IsCategory() bool {
	return c.Type == ChannelTypeGuildCategory
}

// IsThread returns true if this is a thread
func (c *Channel) IsThread() bool {
	return c.Type == ChannelTypeAnnouncementThread ||
		c.Type == ChannelTypePublicThread ||
		c.Type == ChannelTypePrivateThread
}

// DMName returns a display name for DM channels
func (c *Channel) DMName() string {
	if c.Name != "" {
		return c.Name
	}
	if len(c.Recipients) == 1 {
		return c.Recipients[0].DisplayName()
	}
	if len(c.Recipients) > 1 {
		name := ""
		for i, r := range c.Recipients {
			if i > 0 {
				name += ", "
			}
			name += r.DisplayName()
			if i >= 2 && len(c.Recipients) > 3 {
				name += " and " + itoa(len(c.Recipients)-3) + " others"
				break
			}
		}
		return name
	}
	return "Unknown"
}

// PermissionOverwrite represents channel-specific permissions
type PermissionOverwrite struct {
	ID    string `json:"id"`
	Type  int    `json:"type"` // 0 = role, 1 = member
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

// ThreadMetadata contains thread-specific fields
type ThreadMetadata struct {
	Archived            bool   `json:"archived"`
	AutoArchiveDuration int    `json:"auto_archive_duration"`
	ArchiveTimestamp    string `json:"archive_timestamp"`
	Locked              bool   `json:"locked"`
	Invitable           bool   `json:"invitable,omitempty"`
	CreateTimestamp     string `json:"create_timestamp,omitempty"`
}

// ThreadMember represents a member of a thread
type ThreadMember struct {
	ID            string  `json:"id,omitempty"`
	UserID        string  `json:"user_id,omitempty"`
	JoinTimestamp string  `json:"join_timestamp"`
	Flags         int     `json:"flags"`
	Member        *Member `json:"member,omitempty"`
}

// ForumTag represents a tag in a forum channel
type ForumTag struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Moderated bool   `json:"moderated"`
	EmojiID   string `json:"emoji_id,omitempty"`
	EmojiName string `json:"emoji_name,omitempty"`
}

// DefaultReaction for forum posts
type DefaultReaction struct {
	EmojiID   string `json:"emoji_id,omitempty"`
	EmojiName string `json:"emoji_name,omitempty"`
}

// TypingStart event data
type TypingStart struct {
	ChannelID string  `json:"channel_id"`
	GuildID   string  `json:"guild_id,omitempty"`
	UserID    string  `json:"user_id"`
	Timestamp int64   `json:"timestamp"`
	Member    *Member `json:"member,omitempty"`
}
