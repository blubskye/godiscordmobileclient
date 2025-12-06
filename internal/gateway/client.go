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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

const (
	GatewayURL = "wss://gateway.discord.gg/?v=10&encoding=json"
)

// EventHandler is called when a gateway event is received
type EventHandler func(eventType string, data json.RawMessage)

// Client manages the Discord gateway WebSocket connection
type Client struct {
	token    string
	conn     *websocket.Conn
	handlers []EventHandler

	// Session state
	sessionID        string
	resumeGatewayURL string
	sequence         int64
	sequenceMu       sync.Mutex

	// Heartbeat
	heartbeatInterval time.Duration
	lastHeartbeatAck  time.Time
	heartbeatMu       sync.Mutex

	// Connection state
	connected bool
	connMu    sync.RWMutex

	// Channels for coordination
	done   chan struct{}
	cancel context.CancelFunc
}

// NewClient creates a new gateway client
func NewClient(token string) *Client {
	return &Client{
		token:    token,
		handlers: make([]EventHandler, 0),
	}
}

// OnEvent registers an event handler
func (c *Client) OnEvent(handler EventHandler) {
	c.handlers = append(c.handlers, handler)
}

// SetToken updates the authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// Connect establishes a connection to the Discord gateway
func (c *Client) Connect(ctx context.Context) error {
	c.connMu.Lock()
	if c.connected {
		c.connMu.Unlock()
		return errors.New("already connected")
	}
	c.connMu.Unlock()

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.done = make(chan struct{})

	// Determine gateway URL
	gatewayURL := GatewayURL
	if c.resumeGatewayURL != "" {
		gatewayURL = c.resumeGatewayURL + "?v=10&encoding=json"
	}

	// Connect to gateway
	conn, _, err := websocket.Dial(ctx, gatewayURL, &websocket.DialOptions{
		CompressionMode: websocket.CompressionContextTakeover,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}
	c.conn = conn
	conn.SetReadLimit(16 * 1024 * 1024) // 16MB

	c.connMu.Lock()
	c.connected = true
	c.connMu.Unlock()

	// Start read loop
	go c.readLoop(ctx)

	return nil
}

// Disconnect closes the gateway connection
func (c *Client) Disconnect() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false
	if c.cancel != nil {
		c.cancel()
	}

	if c.conn != nil {
		return c.conn.Close(websocket.StatusNormalClosure, "disconnecting")
	}

	return nil
}

// IsConnected returns true if connected to the gateway
func (c *Client) IsConnected() bool {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.connected
}

// Send sends a payload to the gateway
func (c *Client) Send(ctx context.Context, op Opcode, data interface{}) error {
	payload := GatewayPayload{
		Op: op,
	}

	if data != nil {
		d, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		payload.Data = d
	}

	msg, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	c.connMu.RLock()
	conn := c.conn
	c.connMu.RUnlock()

	if conn == nil {
		return errors.New("not connected")
	}

	return conn.Write(ctx, websocket.MessageText, msg)
}

// readLoop reads messages from the gateway
func (c *Client) readLoop(ctx context.Context) {
	defer close(c.done)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, msg, err := c.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				log.Printf("gateway closed: %v", err)
			} else {
				log.Printf("gateway read error: %v", err)
			}
			c.handleDisconnect(ctx)
			return
		}

		var payload GatewayPayload
		if err := json.Unmarshal(msg, &payload); err != nil {
			log.Printf("failed to unmarshal gateway payload: %v", err)
			continue
		}

		c.handlePayload(ctx, &payload)
	}
}

// handlePayload processes incoming gateway payloads
func (c *Client) handlePayload(ctx context.Context, payload *GatewayPayload) {
	// Update sequence number
	if payload.Sequence != nil {
		c.sequenceMu.Lock()
		c.sequence = *payload.Sequence
		c.sequenceMu.Unlock()
	}

	switch payload.Op {
	case OpcodeHello:
		var hello HelloData
		if err := json.Unmarshal(payload.Data, &hello); err != nil {
			log.Printf("failed to unmarshal hello: %v", err)
			return
		}
		c.heartbeatInterval = time.Duration(hello.HeartbeatInterval) * time.Millisecond
		go c.heartbeatLoop(ctx)
		c.identify(ctx)

	case OpcodeHeartbeat:
		// Server requested immediate heartbeat
		c.sendHeartbeat(ctx)

	case OpcodeHeartbeatACK:
		c.heartbeatMu.Lock()
		c.lastHeartbeatAck = time.Now()
		c.heartbeatMu.Unlock()

	case OpcodeReconnect:
		log.Println("gateway requested reconnect")
		c.handleDisconnect(ctx)

	case OpcodeInvalidSession:
		var resumable bool
		json.Unmarshal(payload.Data, &resumable)
		if !resumable {
			c.sessionID = ""
			c.sequence = 0
		}
		// Wait before reconnecting
		time.Sleep(1 * time.Second + time.Duration(float64(time.Second)*2*0.5))
		c.handleDisconnect(ctx)

	case OpcodeDispatch:
		c.handleEvent(payload.Type, payload.Data)
	}
}

// handleEvent processes dispatch events
func (c *Client) handleEvent(eventType string, data json.RawMessage) {
	// Handle READY specially to store session info
	if eventType == EventReady {
		var ready ReadyData
		if err := json.Unmarshal(data, &ready); err == nil {
			c.sessionID = ready.SessionID
			c.resumeGatewayURL = ready.ResumeGatewayURL
		}
	}

	// Call registered handlers
	for _, handler := range c.handlers {
		handler(eventType, data)
	}
}

// identify sends the IDENTIFY payload
func (c *Client) identify(ctx context.Context) {
	// Check if we can resume
	if c.sessionID != "" && c.sequence > 0 {
		c.resume(ctx)
		return
	}

	identify := IdentifyData{
		Token: c.token,
		Properties: IdentifyProperties{
			OS:      "Android",
			Browser: "Discord Mobile",
			Device:  "discord-mobile",
		},
		Compress:       false,
		LargeThreshold: 250,
		Capabilities:   30717, // Standard client capabilities
	}

	if err := c.Send(ctx, OpcodeIdentify, identify); err != nil {
		log.Printf("failed to send identify: %v", err)
	}
}

// resume sends the RESUME payload
func (c *Client) resume(ctx context.Context) {
	c.sequenceMu.Lock()
	seq := c.sequence
	c.sequenceMu.Unlock()

	resume := ResumeData{
		Token:     c.token,
		SessionID: c.sessionID,
		Sequence:  seq,
	}

	if err := c.Send(ctx, OpcodeResume, resume); err != nil {
		log.Printf("failed to send resume: %v", err)
	}
}

// heartbeatLoop sends heartbeats at the specified interval
func (c *Client) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()

	// Send initial heartbeat
	c.sendHeartbeat(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.heartbeatMu.Lock()
			lastAck := c.lastHeartbeatAck
			c.heartbeatMu.Unlock()

			// Check if we received ACK for last heartbeat
			if !lastAck.IsZero() && time.Since(lastAck) > c.heartbeatInterval*2 {
				log.Println("heartbeat ACK timeout, reconnecting")
				c.handleDisconnect(ctx)
				return
			}

			c.sendHeartbeat(ctx)
		}
	}
}

// sendHeartbeat sends a heartbeat
func (c *Client) sendHeartbeat(ctx context.Context) {
	c.sequenceMu.Lock()
	seq := c.sequence
	c.sequenceMu.Unlock()

	var data interface{}
	if seq > 0 {
		data = seq
	}

	if err := c.Send(ctx, OpcodeHeartbeat, data); err != nil {
		log.Printf("failed to send heartbeat: %v", err)
	}
}

// handleDisconnect handles disconnection and reconnection
func (c *Client) handleDisconnect(ctx context.Context) {
	c.connMu.Lock()
	c.connected = false
	if c.conn != nil {
		c.conn.Close(websocket.StatusGoingAway, "reconnecting")
	}
	c.connMu.Unlock()

	// Attempt to reconnect with exponential backoff
	backoff := time.Second
	maxBackoff := 2 * time.Minute

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		log.Printf("attempting to reconnect...")
		if err := c.Connect(ctx); err != nil {
			log.Printf("reconnect failed: %v", err)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		log.Println("reconnected successfully")
		return
	}
}

// UpdatePresence updates the client's presence
func (c *Client) UpdatePresence(ctx context.Context, presence *PresenceUpdate) error {
	return c.Send(ctx, OpcodePresenceUpdate, presence)
}

// JoinVoiceChannel joins a voice channel
func (c *Client) JoinVoiceChannel(ctx context.Context, guildID, channelID string, mute, deaf bool) error {
	return c.Send(ctx, OpcodeVoiceStateUpdate, VoiceStateUpdateData{
		GuildID:   guildID,
		ChannelID: channelID,
		SelfMute:  mute,
		SelfDeaf:  deaf,
	})
}

// LeaveVoiceChannel leaves the current voice channel
func (c *Client) LeaveVoiceChannel(ctx context.Context, guildID string) error {
	return c.Send(ctx, OpcodeVoiceStateUpdate, VoiceStateUpdateData{
		GuildID:   guildID,
		ChannelID: "", // null to disconnect
		SelfMute:  false,
		SelfDeaf:  false,
	})
}

// RequestGuildMembers requests guild members (for large guilds)
func (c *Client) RequestGuildMembers(ctx context.Context, guildID string, query string, limit int) error {
	return c.Send(ctx, OpcodeRequestGuildMembers, RequestGuildMembersData{
		GuildID: guildID,
		Query:   query,
		Limit:   limit,
	})
}

// GetSessionID returns the current session ID
func (c *Client) GetSessionID() string {
	return c.sessionID
}

// GetSequence returns the current sequence number
func (c *Client) GetSequence() int64 {
	c.sequenceMu.Lock()
	defer c.sequenceMu.Unlock()
	return c.sequence
}
