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
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "modernc.org/sqlite"
)

// CacheLimitMode defines how cache limits are managed
type CacheLimitMode int

const (
	// CacheLimitAuto automatically manages cache based on available storage
	CacheLimitAuto CacheLimitMode = iota
	// CacheLimitManual uses user-defined limits
	CacheLimitManual
)

// DebugLevel represents the logging verbosity level
type DebugLevel int

const (
	DebugLevelOff DebugLevel = iota
	DebugLevelError
	DebugLevelWarn
	DebugLevelInfo
	DebugLevelDebug
	DebugLevelTrace
)

// Config holds storage configuration
type Config struct {
	// Mode determines how limits are calculated
	Mode CacheLimitMode `json:"mode"`

	// Memory cache limits (in-memory, fast access)
	// These are always active for performance
	MemoryMessagesPerChannel int `json:"memory_messages_per_channel"` // Messages kept in RAM per channel
	MemoryGuildsMax          int `json:"memory_guilds_max"`           // Max guilds to keep fully loaded in RAM
	MemoryPresencesMax       int `json:"memory_presences_max"`        // Max presence records in RAM

	// Database limits (SQLite, persistent)
	DatabaseMaxSizeMB      int64 `json:"database_max_size_mb"`     // Max database file size in MB
	DatabaseMessagesMax    int   `json:"database_messages_max"`    // Total messages to store
	DatabaseAttachmentsMax int   `json:"database_attachments_max"` // Max cached attachments

	// Auto mode settings
	AutoReserveStorageMB  int64   `json:"auto_reserve_storage_mb"`  // Storage to keep free
	AutoMaxStoragePercent float64 `json:"auto_max_storage_percent"` // Max % of free storage to use

	// Cleanup settings
	CleanupIntervalMinutes int  `json:"cleanup_interval_minutes"` // How often to run cleanup
	CleanupOnLowMemory     bool `json:"cleanup_on_low_memory"`    // Cleanup when memory is low
	VacuumOnCleanup        bool `json:"vacuum_on_cleanup"`        // VACUUM database on cleanup

	// Debug settings
	DebugEnabled     bool       `json:"debug_enabled"`      // Enable debug mode
	DebugLevel       DebugLevel `json:"debug_level"`        // Logging verbosity (0=off, 5=trace)
	DebugStackTraces bool       `json:"debug_stack_traces"` // Include stack traces on errors
	DebugLogToFile   bool       `json:"debug_log_to_file"`  // Write logs to file
	DebugLogFile     string     `json:"debug_log_file"`     // Log file path (empty = auto)

	// Data directory
	DataDir string `json:"data_dir"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Mode: CacheLimitAuto,

		// Memory limits (conservative for mobile)
		MemoryMessagesPerChannel: 100,
		MemoryGuildsMax:          50,
		MemoryPresencesMax:       500,

		// Database limits
		DatabaseMaxSizeMB:      500,   // 500MB max
		DatabaseMessagesMax:    50000, // 50k messages
		DatabaseAttachmentsMax: 1000,

		// Auto mode
		AutoReserveStorageMB:  500,  // Keep 500MB free
		AutoMaxStoragePercent: 10.0, // Use max 10% of free storage

		// Cleanup
		CleanupIntervalMinutes: 30,
		CleanupOnLowMemory:     true,
		VacuumOnCleanup:        false, // Can be slow

		// Debug (off by default)
		DebugEnabled:     false,
		DebugLevel:       DebugLevelInfo,
		DebugStackTraces: false,
		DebugLogToFile:   false,
		DebugLogFile:     "",

		DataDir: defaultDataDir(),
	}
}

// defaultDataDir returns the default data directory for the platform
func defaultDataDir() string {
	switch runtime.GOOS {
	case "android":
		// Android app data directory
		return "/data/data/com.blubskye.discordclient/files"
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "DiscordMobile")
	case "windows":
		appData := os.Getenv("APPDATA")
		return filepath.Join(appData, "DiscordMobile")
	default:
		// Linux/Unix
		home, _ := os.UserHomeDir()
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData == "" {
			xdgData = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(xdgData, "discord-mobile")
	}
}

// LowMemoryConfig returns config optimized for low memory devices
func LowMemoryConfig() *Config {
	cfg := DefaultConfig()
	cfg.MemoryMessagesPerChannel = 50
	cfg.MemoryGuildsMax = 20
	cfg.MemoryPresencesMax = 200
	cfg.DatabaseMaxSizeMB = 200
	cfg.DatabaseMessagesMax = 20000
	cfg.CleanupOnLowMemory = true
	return cfg
}

// HighPerformanceConfig returns config for devices with more resources
func HighPerformanceConfig() *Config {
	cfg := DefaultConfig()
	cfg.MemoryMessagesPerChannel = 200
	cfg.MemoryGuildsMax = 100
	cfg.MemoryPresencesMax = 2000
	cfg.DatabaseMaxSizeMB = 2000
	cfg.DatabaseMessagesMax = 200000
	cfg.VacuumOnCleanup = true
	return cfg
}

// Save saves the config to the SQLite database
// This method opens a temporary connection to save settings
// For better performance, use Database.SaveConfig when you have an open database
func (c *Config) Save() error {
	if err := os.MkdirAll(c.DataDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(c.DataDir, "discord.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return err
	}
	defer db.Close()

	// Ensure settings table exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			value_type TEXT NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Save each setting in a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO app_settings (key, value, value_type, updated_at)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Unix()

	// Helper to save a setting
	saveSetting := func(key string, value interface{}, valueType string) error {
		_, err := stmt.ExecContext(ctx, key, value, valueType, now)
		return err
	}

	// Save all settings
	if err := saveSetting("cache_mode", int(c.Mode), "int"); err != nil {
		return err
	}
	if err := saveSetting("memory_messages_per_channel", c.MemoryMessagesPerChannel, "int"); err != nil {
		return err
	}
	if err := saveSetting("memory_guilds_max", c.MemoryGuildsMax, "int"); err != nil {
		return err
	}
	if err := saveSetting("memory_presences_max", c.MemoryPresencesMax, "int"); err != nil {
		return err
	}
	if err := saveSetting("database_max_size_mb", c.DatabaseMaxSizeMB, "int"); err != nil {
		return err
	}
	if err := saveSetting("database_messages_max", c.DatabaseMessagesMax, "int"); err != nil {
		return err
	}
	if err := saveSetting("database_attachments_max", c.DatabaseAttachmentsMax, "int"); err != nil {
		return err
	}
	if err := saveSetting("auto_reserve_storage_mb", c.AutoReserveStorageMB, "int"); err != nil {
		return err
	}
	if err := saveSetting("auto_max_storage_percent", c.AutoMaxStoragePercent, "float"); err != nil {
		return err
	}
	if err := saveSetting("cleanup_interval_minutes", c.CleanupIntervalMinutes, "int"); err != nil {
		return err
	}
	boolVal := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	if err := saveSetting("cleanup_on_low_memory", boolVal(c.CleanupOnLowMemory), "bool"); err != nil {
		return err
	}
	if err := saveSetting("vacuum_on_cleanup", boolVal(c.VacuumOnCleanup), "bool"); err != nil {
		return err
	}
	if err := saveSetting("debug_enabled", boolVal(c.DebugEnabled), "bool"); err != nil {
		return err
	}
	if err := saveSetting("debug_level", int(c.DebugLevel), "int"); err != nil {
		return err
	}
	if err := saveSetting("debug_stack_traces", boolVal(c.DebugStackTraces), "bool"); err != nil {
		return err
	}
	if err := saveSetting("debug_log_to_file", boolVal(c.DebugLogToFile), "bool"); err != nil {
		return err
	}
	if err := saveSetting("debug_log_file", c.DebugLogFile, "string"); err != nil {
		return err
	}

	return tx.Commit()
}

// LoadConfig loads config from SQLite database, or returns defaults if not found
func LoadConfig(dataDir string) (*Config, error) {
	if dataDir == "" {
		dataDir = defaultDataDir()
	}

	cfg := DefaultConfig()
	cfg.DataDir = dataDir

	// Ensure directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return cfg, nil // Return defaults on error
	}

	dbPath := filepath.Join(dataDir, "discord.db")

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return cfg, nil // Return defaults, will be created on first save
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return cfg, nil // Return defaults on error
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Helper to get settings
	getInt := func(key string, def int) int {
		var value int
		err := db.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = ?", key).Scan(&value)
		if err != nil {
			return def
		}
		return value
	}

	getInt64 := func(key string, def int64) int64 {
		var value int64
		err := db.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = ?", key).Scan(&value)
		if err != nil {
			return def
		}
		return value
	}

	getFloat := func(key string, def float64) float64 {
		var value float64
		err := db.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = ?", key).Scan(&value)
		if err != nil {
			return def
		}
		return value
	}

	getBool := func(key string, def bool) bool {
		var value int
		err := db.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = ?", key).Scan(&value)
		if err != nil {
			return def
		}
		return value == 1
	}

	getString := func(key string, def string) string {
		var value string
		err := db.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = ?", key).Scan(&value)
		if err != nil {
			return def
		}
		return value
	}

	// Load all settings
	cfg.Mode = CacheLimitMode(getInt("cache_mode", int(cfg.Mode)))
	cfg.MemoryMessagesPerChannel = getInt("memory_messages_per_channel", cfg.MemoryMessagesPerChannel)
	cfg.MemoryGuildsMax = getInt("memory_guilds_max", cfg.MemoryGuildsMax)
	cfg.MemoryPresencesMax = getInt("memory_presences_max", cfg.MemoryPresencesMax)
	cfg.DatabaseMaxSizeMB = getInt64("database_max_size_mb", cfg.DatabaseMaxSizeMB)
	cfg.DatabaseMessagesMax = getInt("database_messages_max", cfg.DatabaseMessagesMax)
	cfg.DatabaseAttachmentsMax = getInt("database_attachments_max", cfg.DatabaseAttachmentsMax)
	cfg.AutoReserveStorageMB = getInt64("auto_reserve_storage_mb", cfg.AutoReserveStorageMB)
	cfg.AutoMaxStoragePercent = getFloat("auto_max_storage_percent", cfg.AutoMaxStoragePercent)
	cfg.CleanupIntervalMinutes = getInt("cleanup_interval_minutes", cfg.CleanupIntervalMinutes)
	cfg.CleanupOnLowMemory = getBool("cleanup_on_low_memory", cfg.CleanupOnLowMemory)
	cfg.VacuumOnCleanup = getBool("vacuum_on_cleanup", cfg.VacuumOnCleanup)
	cfg.DebugEnabled = getBool("debug_enabled", cfg.DebugEnabled)
	cfg.DebugLevel = DebugLevel(getInt("debug_level", int(cfg.DebugLevel)))
	cfg.DebugStackTraces = getBool("debug_stack_traces", cfg.DebugStackTraces)
	cfg.DebugLogToFile = getBool("debug_log_to_file", cfg.DebugLogToFile)
	cfg.DebugLogFile = getString("debug_log_file", cfg.DebugLogFile)

	return cfg, nil
}

// CalculateAutoLimits adjusts limits based on available storage
func (c *Config) CalculateAutoLimits() error {
	if c.Mode != CacheLimitAuto {
		return nil
	}

	// Get available storage
	available, err := getAvailableStorage(c.DataDir)
	if err != nil {
		// Fall back to defaults
		return nil
	}

	// Calculate max database size
	maxFromPercent := int64(float64(available) * c.AutoMaxStoragePercent / 100)
	maxAfterReserve := available - (c.AutoReserveStorageMB * 1024 * 1024)

	// Use the smaller of the two
	maxSize := maxFromPercent
	if maxAfterReserve < maxSize {
		maxSize = maxAfterReserve
	}

	// Convert to MB and apply
	maxSizeMB := maxSize / (1024 * 1024)
	if maxSizeMB > 0 && maxSizeMB < c.DatabaseMaxSizeMB {
		c.DatabaseMaxSizeMB = maxSizeMB
	}

	// Scale message limits based on database size
	// Rough estimate: ~10KB per message with attachments metadata
	estimatedMessagesCapacity := (c.DatabaseMaxSizeMB * 1024) / 10
	if estimatedMessagesCapacity < int64(c.DatabaseMessagesMax) {
		c.DatabaseMessagesMax = int(estimatedMessagesCapacity)
	}

	return nil
}

// getAvailableStorage returns available storage in bytes
func getAvailableStorage(path string) (int64, error) {
	// Ensure directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return 0, err
	}

	// Platform-specific implementation would go here
	// For now, return a reasonable default
	return 10 * 1024 * 1024 * 1024, nil // 10GB default
}

// StorageStats holds storage usage statistics
type StorageStats struct {
	DatabaseSizeBytes   int64   `json:"database_size_bytes"`
	DatabaseSizeMB      int64   `json:"database_size_mb"`
	MessageCount        int     `json:"message_count"`
	GuildCount          int     `json:"guild_count"`
	ChannelCount        int     `json:"channel_count"`
	AttachmentCount     int     `json:"attachment_count"`
	AvailableStorageMB  int64   `json:"available_storage_mb"`
	PercentLimitUsed    float64 `json:"percent_limit_used"`
	MemoryUsedBytes     int64   `json:"memory_used_bytes"`
	MemoryMessagesCount int     `json:"memory_messages_count"`
}
