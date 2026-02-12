package pgx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	casbinMigrate "github.com/onlyin32bit/casbin-policy-migrate"
)

// Adapter implements migration.Adapter using pgx.
type Adapter struct {
	pool      *pgxpool.Pool
	tableName string
}

// NewAdapter creates a new Adapter instance.
func NewAdapter(pool *pgxpool.Pool) *Adapter {
	return &Adapter{
		pool:      pool,
		tableName: "casbin_policy_migrations",
	}
}

func (a *Adapter) Initialize(ctx context.Context) error {
	// Create migrations table if not exists
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(255) PRIMARY KEY,
			checksum VARCHAR(255),
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`, a.tableName)
	_, err := a.pool.Exec(ctx, query)
	return err
}

func (a *Adapter) GetAppliedMigrations(ctx context.Context) ([]casbinMigrate.Migration, error) {
	query := fmt.Sprintf("SELECT id, checksum, applied_at FROM %s ORDER BY id ASC", a.tableName)
	rows, err := a.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []casbinMigrate.Migration
	for rows.Next() {
		var m casbinMigrate.Migration
		if err := rows.Scan(&m.ID, &m.Checksum, &m.AppliedAt); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}
	return migrations, nil
}

func (a *Adapter) MarkMigrationApplied(ctx context.Context, m casbinMigrate.Migration) error {
	query := fmt.Sprintf("INSERT INTO %s (id, checksum, applied_at) VALUES ($1, $2, $3)", a.tableName)
	_, err := a.pool.Exec(ctx, query, m.ID, m.Checksum, time.Now())
	return err
}

func (a *Adapter) MarkMigrationRolledBack(ctx context.Context, m casbinMigrate.Migration) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", a.tableName)
	_, err := a.pool.Exec(ctx, query, m.ID)
	return err
}

// Policy Management

func (a *Adapter) AddPolicy(ctx context.Context, sec string, ptype string, rule []string) error {
	// This depends on the actual Casbin storage schema.
	// Common schema is: casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
	// We assume a standard table 'casbin_rule'.
	// NOTE: If the user has a custom table name for rules, this adapter needs configuration.
	// The standard casbin casbin-pg-adapter uses "casbin_rule".

	// Helper to handle varying rule lengths (v0-v5)
	args := make([]interface{}, 7) // ptype + v0..v5
	args[0] = ptype
	for i, r := range rule {
		if i < 6 {
			args[i+1] = r
		}
	}
	// Fill rest with empty strings if needed (some adapters use NULL, some empty string)
	// Standard casbin usually uses empty strings for unused columns in SQL adapters.
	for i := len(rule) + 1; i < 7; i++ {
		args[i] = ""
	}

	query := `INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := a.pool.Exec(ctx, query, args...)
	return err
}

func (a *Adapter) RemovePolicy(ctx context.Context, sec string, ptype string, rule []string) error {
	// Construct WHERE clause
	// We only match on provided fields
	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "ptype = $1")
	args = append(args, ptype)

	for i, r := range rule {
		if i < 6 {
			whereClauses = append(whereClauses, fmt.Sprintf("v%d = $%d", i, i+2))
			args = append(args, r)
		}
	}

	query := fmt.Sprintf("DELETE FROM casbin_rule WHERE %s", strings.Join(whereClauses, " AND "))
	_, err := a.pool.Exec(ctx, query, args...)
	return err
}
