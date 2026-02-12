package casbinMigrate

import (
	"time"
)

// Migration represents a single migration record.
type Migration struct {
	// ID is the unique identifier for the migration, usually the filename (e.g., "001_init.csv").
	ID string
	// Checksum is the hash of the migration file content to detect changes (optional).
	Checksum string
	// AppliedAt is the timestamp when the migration was applied.
	AppliedAt time.Time
}
