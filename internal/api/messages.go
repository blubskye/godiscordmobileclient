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
	"net/url"
	"strconv"

	"github.com/blubskye/godiscordmobileclient/internal/models"
)

// GetMessagesParams for fetching messages
type GetMessagesParams struct {
	Around string // Get messages around this message ID
	Before string // Get messages before this message ID
	After  string // Get messages after this message ID
	Limit  int    // Max number of messages (1-100, default 50)
}

// GetMessages returns messages from a channel
func (c *Client) GetMessages(ctx context.Context, channelID string, params *GetMessagesParams) ([]models.Message, error) {
	path := "/channels/" + channelID + "/messages"

	if params != nil {
		q := url.Values{}
		if params.Around != "" {
			q.Set("around", params.Around)
		}
		if params.Before != "" {
			q.Set("before", params.Before)
		}
		if params.After != "" {
			q.Set("after", params.After)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}

	data, err := c.GET(ctx, path)
	if err != nil {
		return nil, err
	}

	var messages []models.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse messages: %w", err)
	}

	return messages, nil
}

// GetMessage returns a specific message
func (c *Client) GetMessage(ctx context.Context, channelID, messageID string) (*models.Message, error) {
	data, err := c.GET(ctx, "/channels/"+channelID+"/messages/"+messageID)
	if err != nil {
		return nil, err
	}

	var message models.Message
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	return &message, nil
}

// CreateMessage sends a message to a channel
func (c *Client) CreateMessage(ctx context.Context, channelID string, msg *models.CreateMessage) (*models.Message, error) {
	data, err := c.POST(ctx, "/channels/"+channelID+"/messages", msg)
	if err != nil {
		return nil, err
	}

	var message models.Message
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	return &message, nil
}

// SendMessage is a convenience method for sending a simple text message
func (c *Client) SendMessage(ctx context.Context, channelID, content string) (*models.Message, error) {
	return c.CreateMessage(ctx, channelID, &models.CreateMessage{
		Content: content,
	})
}

// ReplyToMessage sends a reply to a message
func (c *Client) ReplyToMessage(ctx context.Context, channelID, messageID, content string, mention bool) (*models.Message, error) {
	return c.CreateMessage(ctx, channelID, &models.CreateMessage{
		Content: content,
		MessageReference: &models.MessageReference{
			MessageID: messageID,
			ChannelID: channelID,
		},
		AllowedMentions: &models.AllowedMentions{
			RepliedUser: mention,
		},
	})
}

// EditMessageParams for editing messages
type EditMessageParams struct {
	Content         string                  `json:"content,omitempty"`
	Embeds          []models.Embed          `json:"embeds,omitempty"`
	Flags           *int                    `json:"flags,omitempty"`
	AllowedMentions *models.AllowedMentions `json:"allowed_mentions,omitempty"`
	Components      []models.Component      `json:"components,omitempty"`
}

// EditMessage edits a message
func (c *Client) EditMessage(ctx context.Context, channelID, messageID string, params *EditMessageParams) (*models.Message, error) {
	data, err := c.PATCH(ctx, "/channels/"+channelID+"/messages/"+messageID, params)
	if err != nil {
		return nil, err
	}

	var message models.Message
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	return &message, nil
}

// DeleteMessage deletes a message
func (c *Client) DeleteMessage(ctx context.Context, channelID, messageID string) error {
	_, err := c.DELETE(ctx, "/channels/"+channelID+"/messages/"+messageID)
	return err
}

// BulkDeleteMessages deletes multiple messages (2-100, not older than 14 days)
func (c *Client) BulkDeleteMessages(ctx context.Context, channelID string, messageIDs []string) error {
	body := map[string][]string{"messages": messageIDs}
	_, err := c.POST(ctx, "/channels/"+channelID+"/messages/bulk-delete", body)
	return err
}

// AddReaction adds a reaction to a message
func (c *Client) AddReaction(ctx context.Context, channelID, messageID, emoji string) error {
	// emoji should be URL encoded: name:id for custom, or URL-encoded unicode
	_, err := c.PUT(ctx, "/channels/"+channelID+"/messages/"+messageID+"/reactions/"+url.PathEscape(emoji)+"/@me", nil)
	return err
}

// RemoveOwnReaction removes your own reaction
func (c *Client) RemoveOwnReaction(ctx context.Context, channelID, messageID, emoji string) error {
	_, err := c.DELETE(ctx, "/channels/"+channelID+"/messages/"+messageID+"/reactions/"+url.PathEscape(emoji)+"/@me")
	return err
}

// RemoveUserReaction removes another user's reaction (requires permissions)
func (c *Client) RemoveUserReaction(ctx context.Context, channelID, messageID, emoji, userID string) error {
	_, err := c.DELETE(ctx, "/channels/"+channelID+"/messages/"+messageID+"/reactions/"+url.PathEscape(emoji)+"/"+userID)
	return err
}

// GetReactionsParams for fetching reactions
type GetReactionsParams struct {
	After string // Get users after this user ID
	Limit int    // Max number of users (1-100, default 25)
}

// GetReactions returns users who reacted with a specific emoji
func (c *Client) GetReactions(ctx context.Context, channelID, messageID, emoji string, params *GetReactionsParams) ([]models.User, error) {
	path := "/channels/" + channelID + "/messages/" + messageID + "/reactions/" + url.PathEscape(emoji)

	if params != nil {
		q := url.Values{}
		if params.After != "" {
			q.Set("after", params.After)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}

	data, err := c.GET(ctx, path)
	if err != nil {
		return nil, err
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to parse users: %w", err)
	}

	return users, nil
}

// DeleteAllReactions removes all reactions from a message
func (c *Client) DeleteAllReactions(ctx context.Context, channelID, messageID string) error {
	_, err := c.DELETE(ctx, "/channels/"+channelID+"/messages/"+messageID+"/reactions")
	return err
}

// DeleteAllReactionsForEmoji removes all reactions for a specific emoji
func (c *Client) DeleteAllReactionsForEmoji(ctx context.Context, channelID, messageID, emoji string) error {
	_, err := c.DELETE(ctx, "/channels/"+channelID+"/messages/"+messageID+"/reactions/"+url.PathEscape(emoji))
	return err
}

// SearchMessagesParams for searching messages
type SearchMessagesParams struct {
	Content   string `json:"content,omitempty"`
	AuthorID  string `json:"author_id,omitempty"`
	Mentions  string `json:"mentions,omitempty"`
	Has       string `json:"has,omitempty"` // link, embed, file, video, image, sound, sticker
	MinID     string `json:"min_id,omitempty"`
	MaxID     string `json:"max_id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	Pinned    *bool  `json:"pinned,omitempty"`
	Offset    int    `json:"offset,omitempty"`
	Limit     int    `json:"limit,omitempty"`      // Default 25, max 25
	SortBy    string `json:"sort_by,omitempty"`    // relevance or timestamp
	SortOrder string `json:"sort_order,omitempty"` // asc or desc
}

// SearchResult from message search
type SearchResult struct {
	TotalResults int                `json:"total_results"`
	Messages     [][]models.Message `json:"messages"` // Each result is an array with context
}

// SearchGuildMessages searches messages in a guild
func (c *Client) SearchGuildMessages(ctx context.Context, guildID string, params *SearchMessagesParams) (*SearchResult, error) {
	q := url.Values{}
	if params.Content != "" {
		q.Set("content", params.Content)
	}
	if params.AuthorID != "" {
		q.Set("author_id", params.AuthorID)
	}
	if params.ChannelID != "" {
		q.Set("channel_id", params.ChannelID)
	}
	if params.Has != "" {
		q.Set("has", params.Has)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}

	path := "/guilds/" + guildID + "/messages/search"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	data, err := c.GET(ctx, path)
	if err != nil {
		return nil, err
	}

	var result SearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search result: %w", err)
	}

	return &result, nil
}

// SearchChannelMessages searches messages in a channel
func (c *Client) SearchChannelMessages(ctx context.Context, channelID string, params *SearchMessagesParams) (*SearchResult, error) {
	q := url.Values{}
	if params.Content != "" {
		q.Set("content", params.Content)
	}
	if params.AuthorID != "" {
		q.Set("author_id", params.AuthorID)
	}
	if params.Has != "" {
		q.Set("has", params.Has)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}

	path := "/channels/" + channelID + "/messages/search"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	data, err := c.GET(ctx, path)
	if err != nil {
		return nil, err
	}

	var result SearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search result: %w", err)
	}

	return &result, nil
}
