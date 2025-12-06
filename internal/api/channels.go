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

// GetChannel returns a channel by ID
func (c *Client) GetChannel(ctx context.Context, channelID string) (*models.Channel, error) {
	data, err := c.GET(ctx, "/channels/"+channelID)
	if err != nil {
		return nil, err
	}

	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return nil, fmt.Errorf("failed to parse channel: %w", err)
	}

	return &channel, nil
}

// GetUserDMChannels returns the current user's DM channels
func (c *Client) GetUserDMChannels(ctx context.Context) ([]models.Channel, error) {
	data, err := c.GET(ctx, "/users/@me/channels")
	if err != nil {
		return nil, err
	}

	var channels []models.Channel
	if err := json.Unmarshal(data, &channels); err != nil {
		return nil, fmt.Errorf("failed to parse channels: %w", err)
	}

	return channels, nil
}

// CreateDM opens a DM channel with a user
func (c *Client) CreateDM(ctx context.Context, recipientID string) (*models.Channel, error) {
	body := map[string]string{"recipient_id": recipientID}
	data, err := c.POST(ctx, "/users/@me/channels", body)
	if err != nil {
		return nil, err
	}

	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return nil, fmt.Errorf("failed to parse channel: %w", err)
	}

	return &channel, nil
}

// CreateGroupDM creates a group DM
func (c *Client) CreateGroupDM(ctx context.Context, accessTokens []string, nicks map[string]string) (*models.Channel, error) {
	body := map[string]interface{}{
		"access_tokens": accessTokens,
		"nicks":         nicks,
	}
	data, err := c.POST(ctx, "/users/@me/channels", body)
	if err != nil {
		return nil, err
	}

	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return nil, fmt.Errorf("failed to parse channel: %w", err)
	}

	return &channel, nil
}

// ModifyChannel updates a channel's settings
func (c *Client) ModifyChannel(ctx context.Context, channelID string, params *ModifyChannelParams) (*models.Channel, error) {
	data, err := c.PATCH(ctx, "/channels/"+channelID, params)
	if err != nil {
		return nil, err
	}

	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return nil, fmt.Errorf("failed to parse channel: %w", err)
	}

	return &channel, nil
}

// ModifyChannelParams for updating channels
type ModifyChannelParams struct {
	Name             string `json:"name,omitempty"`
	Icon             string `json:"icon,omitempty"` // Group DM icon
	Topic            string `json:"topic,omitempty"`
	NSFW             *bool  `json:"nsfw,omitempty"`
	RateLimitPerUser *int   `json:"rate_limit_per_user,omitempty"`
	Bitrate          *int   `json:"bitrate,omitempty"`
	UserLimit        *int   `json:"user_limit,omitempty"`
	Position         *int   `json:"position,omitempty"`
	ParentID         string `json:"parent_id,omitempty"`
	RTCRegion        string `json:"rtc_region,omitempty"`
	VideoQualityMode *int   `json:"video_quality_mode,omitempty"`
}

// DeleteChannel deletes a channel (or closes a DM)
func (c *Client) DeleteChannel(ctx context.Context, channelID string) (*models.Channel, error) {
	data, err := c.DELETE(ctx, "/channels/"+channelID)
	if err != nil {
		return nil, err
	}

	var channel models.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return nil, fmt.Errorf("failed to parse channel: %w", err)
	}

	return &channel, nil
}

// TriggerTypingIndicator starts a typing indicator in a channel
func (c *Client) TriggerTypingIndicator(ctx context.Context, channelID string) error {
	_, err := c.POST(ctx, "/channels/"+channelID+"/typing", nil)
	return err
}

// GetPinnedMessages returns pinned messages in a channel
func (c *Client) GetPinnedMessages(ctx context.Context, channelID string) ([]models.Message, error) {
	data, err := c.GET(ctx, "/channels/"+channelID+"/pins")
	if err != nil {
		return nil, err
	}

	var messages []models.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse messages: %w", err)
	}

	return messages, nil
}

// PinMessage pins a message
func (c *Client) PinMessage(ctx context.Context, channelID, messageID string) error {
	_, err := c.PUT(ctx, "/channels/"+channelID+"/pins/"+messageID, nil)
	return err
}

// UnpinMessage unpins a message
func (c *Client) UnpinMessage(ctx context.Context, channelID, messageID string) error {
	_, err := c.DELETE(ctx, "/channels/"+channelID+"/pins/"+messageID)
	return err
}

// AddGroupDMRecipient adds a user to a group DM
func (c *Client) AddGroupDMRecipient(ctx context.Context, channelID, userID, accessToken, nick string) error {
	body := map[string]string{
		"access_token": accessToken,
		"nick":         nick,
	}
	_, err := c.PUT(ctx, "/channels/"+channelID+"/recipients/"+userID, body)
	return err
}

// RemoveGroupDMRecipient removes a user from a group DM
func (c *Client) RemoveGroupDMRecipient(ctx context.Context, channelID, userID string) error {
	_, err := c.DELETE(ctx, "/channels/"+channelID+"/recipients/"+userID)
	return err
}
