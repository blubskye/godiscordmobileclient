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
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// CacheLimitMode defines how cache limits are managed
type CacheLimitMode int

const (
	// CacheLimitAuto automatically manages cache based on available storage
	CacheLimitAuto CacheLimitMode = iota
	// CacheLimitManual uses user-defined limits
	CacheLimitManual
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

// Save saves the config to a file
func (c *Config) Save() error {
	if err := os.MkdirAll(c.DataDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(c.DataDir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Load loads config from file, or returns defaults if not found
func LoadConfig(dataDir string) (*Config, error) {
	if dataDir == "" {
		dataDir = defaultDataDir()
	}

	path := filepath.Join(dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			cfg.DataDir = dataDir
			return cfg, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Ensure data dir is set
	if cfg.DataDir == "" {
		cfg.DataDir = dataDir
	}

	return &cfg, nil
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
