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

// GetGuild returns a guild by ID
func (c *Client) GetGuild(ctx context.Context, guildID string, withCounts bool) (*models.Guild, error) {
	path := "/guilds/" + guildID
	if withCounts {
		path += "?with_counts=true"
	}

	data, err := c.GET(ctx, path)
	if err != nil {
		return nil, err
	}

	var guild models.Guild
	if err := json.Unmarshal(data, &guild); err != nil {
		return nil, fmt.Errorf("failed to parse guild: %w", err)
	}

	return &guild, nil
}

// GetGuildPreview returns a guild preview (for discoverable guilds)
func (c *Client) GetGuildPreview(ctx context.Context, guildID string) (*models.Guild, error) {
	data, err := c.GET(ctx, "/guilds/"+guildID+"/preview")
	if err != nil {
		return nil, err
	}

	var guild models.Guild
	if err := json.Unmarshal(data, &guild); err != nil {
		return nil, fmt.Errorf("failed to parse guild preview: %w", err)
	}

	return &guild, nil
}

// GetGuildChannels returns all channels in a guild
func (c *Client) GetGuildChannels(ctx context.Context, guildID string) ([]models.Channel, error) {
	data, err := c.GET(ctx, "/guilds/"+guildID+"/channels")
	if err != nil {
		return nil, err
	}

	var channels []models.Channel
	if err := json.Unmarshal(data, &channels); err != nil {
		return nil, fmt.Errorf("failed to parse channels: %w", err)
	}

	return channels, nil
}

// GetGuildMember returns a guild member
func (c *Client) GetGuildMember(ctx context.Context, guildID, userID string) (*models.Member, error) {
	data, err := c.GET(ctx, "/guilds/"+guildID+"/members/"+userID)
	if err != nil {
		return nil, err
	}

	var member models.Member
	if err := json.Unmarshal(data, &member); err != nil {
		return nil, fmt.Errorf("failed to parse member: %w", err)
	}

	return &member, nil
}

// ListGuildMembersParams for listing members
type ListGuildMembersParams struct {
	Limit int    // Max members to return (1-1000, default 1)
	After string // Get members after this user ID
}

// ListGuildMembers returns members in a guild
func (c *Client) ListGuildMembers(ctx context.Context, guildID string, params *ListGuildMembersParams) ([]models.Member, error) {
	path := "/guilds/" + guildID + "/members"

	if params != nil {
		q := url.Values{}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.After != "" {
			q.Set("after", params.After)
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}

	data, err := c.GET(ctx, path)
	if err != nil {
		return nil, err
	}

	var members []models.Member
	if err := json.Unmarshal(data, &members); err != nil {
		return nil, fmt.Errorf("failed to parse members: %w", err)
	}

	return members, nil
}

// SearchGuildMembers searches for members by username/nickname
func (c *Client) SearchGuildMembers(ctx context.Context, guildID, query string, limit int) ([]models.Member, error) {
	q := url.Values{}
	q.Set("query", query)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	data, err := c.GET(ctx, "/guilds/"+guildID+"/members/search?"+q.Encode())
	if err != nil {
		return nil, err
	}

	var members []models.Member
	if err := json.Unmarshal(data, &members); err != nil {
		return nil, fmt.Errorf("failed to parse members: %w", err)
	}

	return members, nil
}

// LeaveGuild leaves a guild
func (c *Client) LeaveGuild(ctx context.Context, guildID string) error {
	_, err := c.DELETE(ctx, "/users/@me/guilds/"+guildID)
	return err
}

// GetGuildRoles returns all roles in a guild
func (c *Client) GetGuildRoles(ctx context.Context, guildID string) ([]models.Role, error) {
	data, err := c.GET(ctx, "/guilds/"+guildID+"/roles")
	if err != nil {
		return nil, err
	}

	var roles []models.Role
	if err := json.Unmarshal(data, &roles); err != nil {
		return nil, fmt.Errorf("failed to parse roles: %w", err)
	}

	return roles, nil
}

// GetGuildEmojis returns all emojis in a guild
func (c *Client) GetGuildEmojis(ctx context.Context, guildID string) ([]models.Emoji, error) {
	data, err := c.GET(ctx, "/guilds/"+guildID+"/emojis")
	if err != nil {
		return nil, err
	}

	var emojis []models.Emoji
	if err := json.Unmarshal(data, &emojis); err != nil {
		return nil, fmt.Errorf("failed to parse emojis: %w", err)
	}

	return emojis, nil
}

// GetGuildStickers returns all stickers in a guild
func (c *Client) GetGuildStickers(ctx context.Context, guildID string) ([]models.Sticker, error) {
	data, err := c.GET(ctx, "/guilds/"+guildID+"/stickers")
	if err != nil {
		return nil, err
	}

	var stickers []models.Sticker
	if err := json.Unmarshal(data, &stickers); err != nil {
		return nil, fmt.Errorf("failed to parse stickers: %w", err)
	}

	return stickers, nil
}

// Invite represents a guild invite
type Invite struct {
	Code                     string         `json:"code"`
	Guild                    *models.Guild  `json:"guild,omitempty"`
	Channel                  *models.Channel `json:"channel,omitempty"`
	Inviter                  *models.User   `json:"inviter,omitempty"`
	TargetType               int            `json:"target_type,omitempty"`
	TargetUser               *models.User   `json:"target_user,omitempty"`
	ApproximatePresenceCount int            `json:"approximate_presence_count,omitempty"`
	ApproximateMemberCount   int            `json:"approximate_member_count,omitempty"`
	ExpiresAt                string         `json:"expires_at,omitempty"`
	Uses                     int            `json:"uses,omitempty"`
	MaxUses                  int            `json:"max_uses,omitempty"`
	MaxAge                   int            `json:"max_age,omitempty"`
	Temporary                bool           `json:"temporary,omitempty"`
	CreatedAt                string         `json:"created_at,omitempty"`
}

// GetInvite returns an invite by code
func (c *Client) GetInvite(ctx context.Context, code string, withCounts, withExpiration bool) (*Invite, error) {
	q := url.Values{}
	if withCounts {
		q.Set("with_counts", "true")
	}
	if withExpiration {
		q.Set("with_expiration", "true")
	}

	path := "/invites/" + code
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	data, err := c.GET(ctx, path)
	if err != nil {
		return nil, err
	}

	var invite Invite
	if err := json.Unmarshal(data, &invite); err != nil {
		return nil, fmt.Errorf("failed to parse invite: %w", err)
	}

	return &invite, nil
}

// AcceptInvite joins a guild via invite
func (c *Client) AcceptInvite(ctx context.Context, code string) (*Invite, error) {
	data, err := c.POST(ctx, "/invites/"+code, nil)
	if err != nil {
		return nil, err
	}

	var invite Invite
	if err := json.Unmarshal(data, &invite); err != nil {
		return nil, fmt.Errorf("failed to parse invite: %w", err)
	}

	return &invite, nil
}

// GetGuildInvites returns all invites in a guild
func (c *Client) GetGuildInvites(ctx context.Context, guildID string) ([]Invite, error) {
	data, err := c.GET(ctx, "/guilds/"+guildID+"/invites")
	if err != nil {
		return nil, err
	}

	var invites []Invite
	if err := json.Unmarshal(data, &invites); err != nil {
		return nil, fmt.Errorf("failed to parse invites: %w", err)
	}

	return invites, nil
}

// VoiceRegion represents a voice region
type VoiceRegion struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Optimal    bool   `json:"optimal"`
	Deprecated bool   `json:"deprecated"`
	Custom     bool   `json:"custom"`
}

// GetVoiceRegions returns available voice regions
func (c *Client) GetVoiceRegions(ctx context.Context) ([]VoiceRegion, error) {
	data, err := c.GET(ctx, "/voice/regions")
	if err != nil {
		return nil, err
	}

	var regions []VoiceRegion
	if err := json.Unmarshal(data, &regions); err != nil {
		return nil, fmt.Errorf("failed to parse voice regions: %w", err)
	}

	return regions, nil
}

// GetGuildVoiceRegions returns voice regions for a guild
func (c *Client) GetGuildVoiceRegions(ctx context.Context, guildID string) ([]VoiceRegion, error) {
	data, err := c.GET(ctx, "/guilds/"+guildID+"/regions")
	if err != nil {
		return nil, err
	}

	var regions []VoiceRegion
	if err := json.Unmarshal(data, &regions); err != nil {
		return nil, fmt.Errorf("failed to parse voice regions: %w", err)
	}

	return regions, nil
}
