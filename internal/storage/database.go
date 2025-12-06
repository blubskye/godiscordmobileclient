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

package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/blubskye/godiscordmobileclient/internal/models"
)

// Database wraps the SQLite database
type Database struct {
	db     *sql.DB
	config *Config
}

// NewDatabase creates and initializes the database
func NewDatabase(config *Config) (*Database, error) {
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(config.DataDir, "discord.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite only supports one writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	d := &Database{
		db:     db,
		config: config,
	}

	if err := d.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return d, nil
}

// Close closes the database
func (d *Database) Close() error {
	return d.db.Close()
}

// migrate creates or updates the database schema
func (d *Database) migrate() error {
	schema := `
	-- Schema version tracking
	CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY
	);

	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		discriminator TEXT,
		global_name TEXT,
		avatar TEXT,
		bot INTEGER DEFAULT 0,
		data TEXT,  -- Full JSON for extra fields
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

	-- Guilds table
	CREATE TABLE IF NOT EXISTS guilds (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		icon TEXT,
		owner_id TEXT,
		member_count INTEGER,
		data TEXT,  -- Full JSON
		updated_at INTEGER NOT NULL
	);

	-- Channels table
	CREATE TABLE IF NOT EXISTS channels (
		id TEXT PRIMARY KEY,
		guild_id TEXT,
		type INTEGER NOT NULL,
		name TEXT,
		position INTEGER,
		parent_id TEXT,
		topic TEXT,
		last_message_id TEXT,
		data TEXT,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_channels_guild ON channels(guild_id);
	CREATE INDEX IF NOT EXISTS idx_channels_parent ON channels(parent_id);

	-- Messages table (most important for caching)
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		channel_id TEXT NOT NULL,
		author_id TEXT,
		content TEXT,
		timestamp INTEGER NOT NULL,
		edited_timestamp INTEGER,
		type INTEGER DEFAULT 0,
		pinned INTEGER DEFAULT 0,
		has_attachments INTEGER DEFAULT 0,
		has_embeds INTEGER DEFAULT 0,
		data TEXT,  -- Full JSON for complete message
		cached_at INTEGER NOT NULL,
		FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
		FOREIGN KEY (author_id) REFERENCES users(id)
	);
	CREATE INDEX IF NOT EXISTS idx_messages_channel ON messages(channel_id);
	CREATE INDEX IF NOT EXISTS idx_messages_channel_time ON messages(channel_id, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_messages_author ON messages(author_id);
	CREATE INDEX IF NOT EXISTS idx_messages_cached ON messages(cached_at);

	-- Attachments metadata (not the actual files)
	CREATE TABLE IF NOT EXISTS attachments (
		id TEXT PRIMARY KEY,
		message_id TEXT NOT NULL,
		filename TEXT NOT NULL,
		content_type TEXT,
		size INTEGER,
		url TEXT,
		proxy_url TEXT,
		width INTEGER,
		height INTEGER,
		cached_locally INTEGER DEFAULT 0,
		local_path TEXT,
		FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_attachments_message ON attachments(message_id);

	-- Relationships (friends, blocked)
	CREATE TABLE IF NOT EXISTS relationships (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		type INTEGER NOT NULL,
		nickname TEXT,
		data TEXT,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	CREATE INDEX IF NOT EXISTS idx_relationships_type ON relationships(type);

	-- Read states (which messages have been read)
	CREATE TABLE IF NOT EXISTS read_states (
		channel_id TEXT PRIMARY KEY,
		last_message_id TEXT,
		mention_count INTEGER DEFAULT 0,
		updated_at INTEGER NOT NULL
	);

	-- DM channels (private channels)
	CREATE TABLE IF NOT EXISTS dm_channels (
		channel_id TEXT PRIMARY KEY,
		recipient_ids TEXT,  -- JSON array of user IDs
		last_message_id TEXT,
		updated_at INTEGER NOT NULL
	);

	-- Guild settings per guild
	CREATE TABLE IF NOT EXISTS guild_settings (
		guild_id TEXT PRIMARY KEY,
		muted INTEGER DEFAULT 0,
		suppress_everyone INTEGER DEFAULT 0,
		suppress_roles INTEGER DEFAULT 0,
		notification_level INTEGER DEFAULT 0,
		data TEXT,
		updated_at INTEGER NOT NULL
	);

	-- Channel settings
	CREATE TABLE IF NOT EXISTS channel_settings (
		channel_id TEXT PRIMARY KEY,
		muted INTEGER DEFAULT 0,
		collapsed INTEGER DEFAULT 0,
		data TEXT,
		updated_at INTEGER NOT NULL
	);

	-- Search index for message content (FTS5)
	CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
		content,
		content=messages,
		content_rowid=rowid
	);

	-- Triggers to keep FTS in sync
	CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
		INSERT INTO messages_fts(rowid, content) VALUES (new.rowid, new.content);
	END;
	CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
		INSERT INTO messages_fts(messages_fts, rowid, content) VALUES('delete', old.rowid, old.content);
	END;
	CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
		INSERT INTO messages_fts(messages_fts, rowid, content) VALUES('delete', old.rowid, old.content);
		INSERT INTO messages_fts(rowid, content) VALUES (new.rowid, new.content);
	END;
	`

	_, err := d.db.Exec(schema)
	return err
}

// SaveUser saves or updates a user
func (d *Database) SaveUser(ctx context.Context, user *models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	_, err = d.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO users (id, username, discriminator, global_name, avatar, bot, data, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, user.ID, user.Username, user.Discriminator, user.GlobalName, user.Avatar, user.Bot, string(data), time.Now().Unix())
	return err
}

// GetUser retrieves a user by ID
func (d *Database) GetUser(ctx context.Context, id string) (*models.User, error) {
	var data string
	err := d.db.QueryRowContext(ctx, "SELECT data FROM users WHERE id = ?", id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// SaveGuild saves or updates a guild
func (d *Database) SaveGuild(ctx context.Context, guild *models.Guild) error {
	data, err := json.Marshal(guild)
	if err != nil {
		return err
	}

	_, err = d.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO guilds (id, name, icon, owner_id, member_count, data, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, guild.ID, guild.Name, guild.Icon, guild.OwnerID, guild.MemberCount, string(data), time.Now().Unix())
	return err
}

// GetGuild retrieves a guild by ID
func (d *Database) GetGuild(ctx context.Context, id string) (*models.Guild, error) {
	var data string
	err := d.db.QueryRowContext(ctx, "SELECT data FROM guilds WHERE id = ?", id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var guild models.Guild
	if err := json.Unmarshal([]byte(data), &guild); err != nil {
		return nil, err
	}
	return &guild, nil
}

// GetAllGuilds retrieves all guilds
func (d *Database) GetAllGuilds(ctx context.Context) ([]*models.Guild, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT data FROM guilds ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guilds []*models.Guild
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var guild models.Guild
		if err := json.Unmarshal([]byte(data), &guild); err != nil {
			continue
		}
		guilds = append(guilds, &guild)
	}
	return guilds, rows.Err()
}

// SaveChannel saves or updates a channel
func (d *Database) SaveChannel(ctx context.Context, channel *models.Channel) error {
	data, err := json.Marshal(channel)
	if err != nil {
		return err
	}

	_, err = d.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO channels (id, guild_id, type, name, position, parent_id, topic, last_message_id, data, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, channel.ID, channel.GuildID, channel.Type, channel.Name, channel.Position, channel.ParentID, channel.Topic, channel.LastMessageID, string(data), time.Now().Unix())
	return err
}

// GetChannel retrieves a channel by ID
func (d *Database) GetChannel(ctx context.Context, id string) (*models.Channel, error) {
	var data string
	err := d.db.QueryRowContext(ctx, "SELECT data FROM channels WHERE id = ?", id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var channel models.Channel
	if err := json.Unmarshal([]byte(data), &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// GetGuildChannels retrieves all channels for a guild
func (d *Database) GetGuildChannels(ctx context.Context, guildID string) ([]*models.Channel, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT data FROM channels WHERE guild_id = ? ORDER BY position", guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*models.Channel
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var channel models.Channel
		if err := json.Unmarshal([]byte(data), &channel); err != nil {
			continue
		}
		channels = append(channels, &channel)
	}
	return channels, rows.Err()
}

// SaveMessage saves a message
func (d *Database) SaveMessage(ctx context.Context, msg *models.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	hasAttachments := 0
	if len(msg.Attachments) > 0 {
		hasAttachments = 1
	}
	hasEmbeds := 0
	if len(msg.Embeds) > 0 {
		hasEmbeds = 1
	}

	var authorID string
	if msg.Author != nil {
		authorID = msg.Author.ID
	}

	var editedTs *int64
	if msg.EditedTimestamp != nil {
		ts := msg.EditedTimestamp.Unix()
		editedTs = &ts
	}

	_, err = d.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO messages (id, channel_id, author_id, content, timestamp, edited_timestamp, type, pinned, has_attachments, has_embeds, data, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, msg.ID, msg.ChannelID, authorID, msg.Content, msg.Timestamp.Unix(), editedTs, msg.Type, msg.Pinned, hasAttachments, hasEmbeds, string(data), time.Now().Unix())

	if err != nil {
		return err
	}

	// Save attachments
	for _, att := range msg.Attachments {
		_, err = d.db.ExecContext(ctx, `
			INSERT OR REPLACE INTO attachments (id, message_id, filename, content_type, size, url, proxy_url, width, height)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, att.ID, msg.ID, att.Filename, att.ContentType, att.Size, att.URL, att.ProxyURL, att.Width, att.Height)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveMessages saves multiple messages in a transaction
func (d *Database) SaveMessages(ctx context.Context, messages []*models.Message) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO messages (id, channel_id, author_id, content, timestamp, edited_timestamp, type, pinned, has_attachments, has_embeds, data, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Unix()
	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}

		hasAttachments := 0
		if len(msg.Attachments) > 0 {
			hasAttachments = 1
		}
		hasEmbeds := 0
		if len(msg.Embeds) > 0 {
			hasEmbeds = 1
		}

		var authorID string
		if msg.Author != nil {
			authorID = msg.Author.ID
		}

		var editedTs *int64
		if msg.EditedTimestamp != nil {
			ts := msg.EditedTimestamp.Unix()
			editedTs = &ts
		}

		_, err = stmt.ExecContext(ctx, msg.ID, msg.ChannelID, authorID, msg.Content, msg.Timestamp.Unix(), editedTs, msg.Type, msg.Pinned, hasAttachments, hasEmbeds, string(data), now)
		if err != nil {
			continue
		}
	}

	return tx.Commit()
}

// GetMessages retrieves messages for a channel
func (d *Database) GetMessages(ctx context.Context, channelID string, limit int, before, after string) ([]*models.Message, error) {
	query := "SELECT data FROM messages WHERE channel_id = ?"
	args := []interface{}{channelID}

	if before != "" {
		query += " AND id < ?"
		args = append(args, before)
	}
	if after != "" {
		query += " AND id > ?"
		args = append(args, after)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var msg models.Message
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}
	return messages, rows.Err()
}

// GetMessage retrieves a single message by ID
func (d *Database) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	var data string
	err := d.db.QueryRowContext(ctx, "SELECT data FROM messages WHERE id = ?", id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var msg models.Message
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// DeleteMessage deletes a message
func (d *Database) DeleteMessage(ctx context.Context, id string) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM messages WHERE id = ?", id)
	return err
}

// SearchMessages performs full-text search on messages
func (d *Database) SearchMessages(ctx context.Context, query string, channelID string, limit int) ([]*models.Message, error) {
	sqlQuery := `
		SELECT m.data FROM messages m
		JOIN messages_fts fts ON m.rowid = fts.rowid
		WHERE messages_fts MATCH ?
	`
	args := []interface{}{query}

	if channelID != "" {
		sqlQuery += " AND m.channel_id = ?"
		args = append(args, channelID)
	}

	sqlQuery += " ORDER BY m.timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := d.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var msg models.Message
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}
	return messages, rows.Err()
}

// GetStats returns database statistics
func (d *Database) GetStats(ctx context.Context) (*StorageStats, error) {
	stats := &StorageStats{}

	// Get database file size
	dbPath := filepath.Join(d.config.DataDir, "discord.db")
	if info, err := os.Stat(dbPath); err == nil {
		stats.DatabaseSizeBytes = info.Size()
		stats.DatabaseSizeMB = info.Size() / (1024 * 1024)
	}

	// Count messages
	d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM messages").Scan(&stats.MessageCount)

	// Count guilds
	d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM guilds").Scan(&stats.GuildCount)

	// Count channels
	d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM channels").Scan(&stats.ChannelCount)

	// Count attachments
	d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM attachments").Scan(&stats.AttachmentCount)

	// Calculate percent of limit used
	if d.config.DatabaseMaxSizeMB > 0 {
		stats.PercentLimitUsed = float64(stats.DatabaseSizeMB) / float64(d.config.DatabaseMaxSizeMB) * 100
	}

	return stats, nil
}

// Cleanup removes old data to stay within limits
func (d *Database) Cleanup(ctx context.Context) error {
	// Get current stats
	stats, err := d.GetStats(ctx)
	if err != nil {
		return err
	}

	// Check if we need to cleanup
	if stats.DatabaseSizeMB < d.config.DatabaseMaxSizeMB &&
		stats.MessageCount < d.config.DatabaseMessagesMax {
		return nil
	}

	// Delete oldest messages until we're under the limit
	targetMessages := d.config.DatabaseMessagesMax * 80 / 100 // Target 80% of limit
	if stats.MessageCount > targetMessages {
		deleteCount := stats.MessageCount - targetMessages
		_, err = d.db.ExecContext(ctx, `
			DELETE FROM messages WHERE id IN (
				SELECT id FROM messages ORDER BY cached_at ASC LIMIT ?
			)
		`, deleteCount)
		if err != nil {
			return err
		}
	}

	// Vacuum if configured
	if d.config.VacuumOnCleanup {
		_, err = d.db.ExecContext(ctx, "VACUUM")
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteGuild removes a guild and all related data
func (d *Database) DeleteGuild(ctx context.Context, guildID string) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete messages in guild channels
	_, err = tx.ExecContext(ctx, `
		DELETE FROM messages WHERE channel_id IN (
			SELECT id FROM channels WHERE guild_id = ?
		)
	`, guildID)
	if err != nil {
		return err
	}

	// Delete channels
	_, err = tx.ExecContext(ctx, "DELETE FROM channels WHERE guild_id = ?", guildID)
	if err != nil {
		return err
	}

	// Delete guild
	_, err = tx.ExecContext(ctx, "DELETE FROM guilds WHERE id = ?", guildID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
