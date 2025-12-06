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

// User represents a Discord user
type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	GlobalName    string `json:"global_name,omitempty"`
	Avatar        string `json:"avatar,omitempty"`
	Bot           bool   `json:"bot,omitempty"`
	System        bool   `json:"system,omitempty"`
	MFAEnabled    bool   `json:"mfa_enabled,omitempty"`
	Banner        string `json:"banner,omitempty"`
	AccentColor   int    `json:"accent_color,omitempty"`
	Locale        string `json:"locale,omitempty"`
	Verified      bool   `json:"verified,omitempty"`
	Email         string `json:"email,omitempty"`
	Flags         int    `json:"flags,omitempty"`
	PremiumType   int    `json:"premium_type,omitempty"`
	PublicFlags   int    `json:"public_flags,omitempty"`
	Phone         string `json:"phone,omitempty"`
}

// DisplayName returns the user's display name (global name or username)
func (u *User) DisplayName() string {
	if u.GlobalName != "" {
		return u.GlobalName
	}
	return u.Username
}

// AvatarURL returns the URL for the user's avatar
func (u *User) AvatarURL(size int) string {
	if u.Avatar == "" {
		// Default avatar based on discriminator or user ID
		defaultIndex := 0
		if u.Discriminator != "0" && u.Discriminator != "" {
			// Legacy: use discriminator mod 5
			for _, c := range u.Discriminator {
				defaultIndex = (defaultIndex*10 + int(c-'0')) % 5
			}
		} else {
			// New: use (user_id >> 22) mod 6
			var id uint64
			for _, c := range u.ID {
				id = id*10 + uint64(c-'0')
			}
			defaultIndex = int((id >> 22) % 6)
		}
		return "https://cdn.discordapp.com/embed/avatars/" + string(rune('0'+defaultIndex)) + ".png"
	}
	ext := "png"
	if len(u.Avatar) > 2 && u.Avatar[:2] == "a_" {
		ext = "gif"
	}
	return "https://cdn.discordapp.com/avatars/" + u.ID + "/" + u.Avatar + "." + ext + "?size=" + itoa(size)
}

// Presence represents a user's online status
type Presence struct {
	User         *User         `json:"user"`
	GuildID      string        `json:"guild_id,omitempty"`
	Status       Status        `json:"status"`
	Activities   []Activity    `json:"activities"`
	ClientStatus *ClientStatus `json:"client_status,omitempty"`
}

// Status represents online/offline status
type Status string

const (
	StatusOnline    Status = "online"
	StatusDND       Status = "dnd"
	StatusIdle      Status = "idle"
	StatusInvisible Status = "invisible"
	StatusOffline   Status = "offline"
)

// ClientStatus shows status per device
type ClientStatus struct {
	Desktop Status `json:"desktop,omitempty"`
	Mobile  Status `json:"mobile,omitempty"`
	Web     Status `json:"web,omitempty"`
}

// Activity represents what a user is doing
type Activity struct {
	Name          string         `json:"name"`
	Type          ActivityType   `json:"type"`
	URL           string         `json:"url,omitempty"`
	CreatedAt     int64          `json:"created_at"`
	Timestamps    *Timestamps    `json:"timestamps,omitempty"`
	ApplicationID string         `json:"application_id,omitempty"`
	Details       string         `json:"details,omitempty"`
	State         string         `json:"state,omitempty"`
	Emoji         *ActivityEmoji `json:"emoji,omitempty"`
	Party         *Party         `json:"party,omitempty"`
	Assets        *Assets        `json:"assets,omitempty"`
	Secrets       *Secrets       `json:"secrets,omitempty"`
	Instance      bool           `json:"instance,omitempty"`
	Flags         int            `json:"flags,omitempty"`
	Buttons       []string       `json:"buttons,omitempty"`
}

type ActivityType int

const (
	ActivityTypePlaying ActivityType = iota
	ActivityTypeStreaming
	ActivityTypeListening
	ActivityTypeWatching
	ActivityTypeCustom
	ActivityTypeCompeting
)

type Timestamps struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
}

type ActivityEmoji struct {
	Name     string `json:"name"`
	ID       string `json:"id,omitempty"`
	Animated bool   `json:"animated,omitempty"`
}

type Party struct {
	ID   string `json:"id,omitempty"`
	Size []int  `json:"size,omitempty"` // [current_size, max_size]
}

type Assets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
	SmallImage string `json:"small_image,omitempty"`
	SmallText  string `json:"small_text,omitempty"`
}

type Secrets struct {
	Join     string `json:"join,omitempty"`
	Spectate string `json:"spectate,omitempty"`
	Match    string `json:"match,omitempty"`
}

// Relationship represents a friend/blocked user
type Relationship struct {
	ID       string           `json:"id"`
	Type     RelationshipType `json:"type"`
	Nickname string           `json:"nickname,omitempty"`
	User     *User            `json:"user"`
	Since    time.Time        `json:"since,omitempty"`
}

type RelationshipType int

const (
	RelationshipTypeNone RelationshipType = iota
	RelationshipTypeFriend
	RelationshipTypeBlocked
	RelationshipTypePendingIncoming
	RelationshipTypePendingOutgoing
	RelationshipTypeImplicit
)

// Helper to convert int to string without strconv import
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
