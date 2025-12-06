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

  
package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blubskye/godiscordmobileclient/internal/models"
)

// GetCurrentUser returns the current authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*models.User, error) {
	data, err := c.GET(ctx, "/users/@me")
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}

	return &user, nil
}

// GetUser returns a user by ID
func (c *Client) GetUser(ctx context.Context, userID string) (*models.User, error) {
	data, err := c.GET(ctx, "/users/"+userID)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}

	return &user, nil
}

// ModifyCurrentUser updates the current user's settings
func (c *Client) ModifyCurrentUser(ctx context.Context, params *ModifyUserParams) (*models.User, error) {
	data, err := c.PATCH(ctx, "/users/@me", params)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}

	return &user, nil
}

// ModifyUserParams for updating user profile
type ModifyUserParams struct {
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar,omitempty"` // base64 data URI
	Banner   string `json:"banner,omitempty"` // base64 data URI
}

// GetCurrentUserGuilds returns guilds the current user is in
func (c *Client) GetCurrentUserGuilds(ctx context.Context) ([]models.Guild, error) {
	data, err := c.GET(ctx, "/users/@me/guilds")
	if err != nil {
		return nil, err
	}

	var guilds []models.Guild
	if err := json.Unmarshal(data, &guilds); err != nil {
		return nil, fmt.Errorf("failed to parse guilds: %w", err)
	}

	return guilds, nil
}

// GetRelationships returns the current user's relationships (friends, blocked)
func (c *Client) GetRelationships(ctx context.Context) ([]models.Relationship, error) {
	data, err := c.GET(ctx, "/users/@me/relationships")
	if err != nil {
		return nil, err
	}

	var relationships []models.Relationship
	if err := json.Unmarshal(data, &relationships); err != nil {
		return nil, fmt.Errorf("failed to parse relationships: %w", err)
	}

	return relationships, nil
}

// AddRelationship sends a friend request or blocks a user
func (c *Client) AddRelationship(ctx context.Context, userID string, relType models.RelationshipType) error {
	body := map[string]int{"type": int(relType)}
	_, err := c.PUT(ctx, "/users/@me/relationships/"+userID, body)
	return err
}

// RemoveRelationship removes a friend or unblocks a user
func (c *Client) RemoveRelationship(ctx context.Context, userID string) error {
	_, err := c.DELETE(ctx, "/users/@me/relationships/"+userID)
	return err
}

// SendFriendRequest sends a friend request by username
func (c *Client) SendFriendRequest(ctx context.Context, username string) error {
	body := map[string]string{"username": username}
	_, err := c.POST(ctx, "/users/@me/relationships", body)
	return err
}

// UserSettings represents user settings
type UserSettings struct {
	Locale                     string   `json:"locale,omitempty"`
	Theme                      string   `json:"theme,omitempty"`
	ShowCurrentGame            bool     `json:"show_current_game,omitempty"`
	InlineAttachmentMedia      bool     `json:"inline_attachment_media,omitempty"`
	InlineEmbedMedia           bool     `json:"inline_embed_media,omitempty"`
	RenderEmbeds               bool     `json:"render_embeds,omitempty"`
	RenderReactions            bool     `json:"render_reactions,omitempty"`
	AnimateEmoji               bool     `json:"animate_emoji,omitempty"`
	EnableTTSCommand           bool     `json:"enable_tts_command,omitempty"`
	MessageDisplayCompact      bool     `json:"message_display_compact,omitempty"`
	ConvertEmoticons           bool     `json:"convert_emoticons,omitempty"`
	ExplicitContentFilter      int      `json:"explicit_content_filter,omitempty"`
	DisableGamesTab            bool     `json:"disable_games_tab,omitempty"`
	DeveloperMode              bool     `json:"developer_mode,omitempty"`
	GIFAutoPlay                bool     `json:"gif_auto_play,omitempty"`
	AnimateStickers            int      `json:"animate_stickers,omitempty"`
	Status                     string   `json:"status,omitempty"`
	CustomStatus               *CustomStatus `json:"custom_status,omitempty"`
	RestrictedGuilds           []string `json:"restricted_guilds,omitempty"`
	FriendSourceFlags          *FriendSourceFlags `json:"friend_source_flags,omitempty"`
}

type CustomStatus struct {
	Text      string `json:"text,omitempty"`
	EmojiID   string `json:"emoji_id,omitempty"`
	EmojiName string `json:"emoji_name,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

type FriendSourceFlags struct {
	All           bool `json:"all,omitempty"`
	MutualFriends bool `json:"mutual_friends,omitempty"`
	MutualGuilds  bool `json:"mutual_guilds,omitempty"`
}

// GetUserSettings returns the current user's settings
func (c *Client) GetUserSettings(ctx context.Context) (*UserSettings, error) {
	data, err := c.GET(ctx, "/users/@me/settings")
	if err != nil {
		return nil, err
	}

	var settings UserSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	return &settings, nil
}

// UpdateUserSettings updates the current user's settings
func (c *Client) UpdateUserSettings(ctx context.Context, settings *UserSettings) (*UserSettings, error) {
	data, err := c.PATCH(ctx, "/users/@me/settings", settings)
	if err != nil {
		return nil, err
	}

	var updated UserSettings
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	return &updated, nil
}

// UpdateStatus updates the user's online status
func (c *Client) UpdateStatus(ctx context.Context, status models.Status) error {
	settings := &UserSettings{Status: string(status)}
	_, err := c.UpdateUserSettings(ctx, settings)
	return err
}

// SetCustomStatus sets a custom status
func (c *Client) SetCustomStatus(ctx context.Context, text string) error {
	settings := &UserSettings{
		CustomStatus: &CustomStatus{Text: text},
	}
	_, err := c.UpdateUserSettings(ctx, settings)
	return err
}

// ClearCustomStatus clears the custom status
func (c *Client) ClearCustomStatus(ctx context.Context) error {
	settings := &UserSettings{
		CustomStatus: nil,
	}
	_, err := c.UpdateUserSettings(ctx, settings)
	return err
}
