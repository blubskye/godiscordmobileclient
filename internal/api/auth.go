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
	"net/http"

	"github.com/blubskye/godiscordmobileclient/internal/models"
)

// LoginRequest represents the login payload
type LoginRequest struct {
	Login           string `json:"login"`
	Password        string `json:"password"`
	UnderageSelfDel bool   `json:"undelete,omitempty"`
	CaptchaKey      string `json:"captcha_key,omitempty"`
	LoginSource     string `json:"login_source,omitempty"`
	GiftCodeSKUId   string `json:"gift_code_sku_id,omitempty"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token          string   `json:"token,omitempty"`
	MFA            bool     `json:"mfa,omitempty"`
	SMS            bool     `json:"sms,omitempty"`
	Ticket         string   `json:"ticket,omitempty"`
	BackupCodes    bool     `json:"backup,omitempty"`
	TOTP           bool     `json:"totp,omitempty"`
	WebAuthn       string   `json:"webauthn,omitempty"`
	UserID         string   `json:"user_id,omitempty"`
	CaptchaKey     []string `json:"captcha_key,omitempty"`
	CaptchaSiteKey string   `json:"captcha_sitekey,omitempty"`
	CaptchaService string   `json:"captcha_service,omitempty"`
}

// MFARequest represents an MFA verification request
type MFARequest struct {
	Code   string `json:"code"`
	Ticket string `json:"ticket"`
}

// MFAResponse represents an MFA verification response
type MFAResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id,omitempty"`
}

// Login attempts to authenticate with email/password
// Note: This may be blocked by Discord and is against ToS for automation
func (c *Client) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	req := &LoginRequest{
		Login:    email,
		Password: password,
	}

	data, err := c.POST(ctx, "/auth/login", req)
	if err != nil {
		return nil, err
	}

	var resp LoginResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w", err)
	}

	// If we got a token directly, set it
	if resp.Token != "" {
		c.SetToken(resp.Token)
	}

	return &resp, nil
}

// VerifyMFA verifies an MFA code
func (c *Client) VerifyMFA(ctx context.Context, code, ticket string) (*MFAResponse, error) {
	req := &MFARequest{
		Code:   code,
		Ticket: ticket,
	}

	data, err := c.POST(ctx, "/auth/mfa/totp", req)
	if err != nil {
		return nil, err
	}

	var resp MFAResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse MFA response: %w", err)
	}

	if resp.Token != "" {
		c.SetToken(resp.Token)
	}

	return &resp, nil
}

// VerifySMSCode verifies an SMS MFA code
func (c *Client) VerifySMSCode(ctx context.Context, code, ticket string) (*MFAResponse, error) {
	req := &MFARequest{
		Code:   code,
		Ticket: ticket,
	}

	data, err := c.POST(ctx, "/auth/mfa/sms", req)
	if err != nil {
		return nil, err
	}

	var resp MFAResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse SMS MFA response: %w", err)
	}

	if resp.Token != "" {
		c.SetToken(resp.Token)
	}

	return &resp, nil
}

// ValidateToken checks if the current token is valid by fetching user info
func (c *Client) ValidateToken(ctx context.Context) (*models.User, error) {
	return c.GetCurrentUser(ctx)
}

// Logout invalidates the current session
func (c *Client) Logout(ctx context.Context) error {
	_, err := c.Request(ctx, http.MethodPost, "/auth/logout", nil)
	if err != nil {
		return err
	}
	c.SetToken("")
	return nil
}

// RefreshToken attempts to refresh the session
// Note: This endpoint may not work for user tokens
func (c *Client) RefreshToken(ctx context.Context) (string, error) {
	data, err := c.POST(ctx, "/auth/token/refresh", nil)
	if err != nil {
		return "", err
	}

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse refresh response: %w", err)
	}

	if resp.Token != "" {
		c.SetToken(resp.Token)
	}

	return resp.Token, nil
}
