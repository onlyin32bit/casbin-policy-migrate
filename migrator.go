package casbinMigrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migrator handles the migration process.
type Migrator struct {
	adapter     Adapter
	policiesDir string
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(adapter Adapter, policiesDir string) *Migrator {
	return &Migrator{
		adapter:     adapter,
		policiesDir: policiesDir,
	}
}

// Up applies all pending migrations.
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.adapter.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize adapter: %w", err)
	}

	applied, err := m.adapter.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, mgr := range applied {
		appliedMap[mgr.ID] = true
	}

	files, err := os.ReadDir(m.policiesDir)
	if err != nil {
		return fmt.Errorf("failed to read policies directory: %w", err)
	}

	var pendingFiles []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), ".csv") {
			continue
		}
		if !appliedMap[f.Name()] {
			pendingFiles = append(pendingFiles, f.Name())
		}
	}

	sort.Strings(pendingFiles)

	if len(pendingFiles) == 0 {
		return nil
	}

	for _, filename := range pendingFiles {
		fmt.Printf("Applying migration: %s\n", filename)
		if err := m.applyMigration(ctx, filename); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", filename, err)
		}
	}

	return nil
}

// Down rolls back the last n migrations. If n <= 0, it rolls back all.
func (m *Migrator) Down(ctx context.Context, n int) error {
	if err := m.adapter.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize adapter: %w", err)
	}

	applied, err := m.adapter.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return nil
	}

	// Sort applied migrations by ID descending to rollback in reverse order
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].ID > applied[j].ID
	})

	toRollback := applied
	if n > 0 && n < len(applied) {
		toRollback = applied[:n]
	}

	for _, mgr := range toRollback {
		fmt.Printf("Rolling back migration: %s\n", mgr.ID)
		if err := m.rollbackMigration(ctx, mgr); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", mgr.ID, err)
		}
	}

	return nil
}

func (m *Migrator) applyMigration(ctx context.Context, filename string) error {
	path := filepath.Join(m.policiesDir, filename)

	// Parse the file
	ops, err := ParseMigrationFile(path)
	if err != nil {
		return err
	}

	for _, op := range ops {
		if op.Type == OperationAdd {
			if err := m.adapter.AddPolicy(ctx, op.Sec, op.PType, op.Rule); err != nil {
				return fmt.Errorf("failed to add policy: %w", err)
			}
		} else {
			if err := m.adapter.RemovePolicy(ctx, op.Sec, op.PType, op.Rule); err != nil {
				return fmt.Errorf("failed to remove policy: %w", err)
			}
		}
	}

	return m.adapter.MarkMigrationApplied(ctx, Migration{
		ID: filename,
	})
}

func (m *Migrator) rollbackMigration(ctx context.Context, mgr Migration) error {
	path := filepath.Join(m.policiesDir, mgr.ID)

	ops, err := ParseMigrationFile(path)
	if err != nil {
		return err
	}

	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i]
		if op.Type == OperationAdd {
			if err := m.adapter.RemovePolicy(ctx, op.Sec, op.PType, op.Rule); err != nil {
				return fmt.Errorf("failed to remove policy (rollback): %w", err)
			}
		} else {
			if err := m.adapter.AddPolicy(ctx, op.Sec, op.PType, op.Rule); err != nil {
				return fmt.Errorf("failed to add policy (rollback): %w", err)
			}
		}
	}

	return m.adapter.MarkMigrationRolledBack(ctx, mgr)
}

// Status returns the status of migrations (applied and pending).
func (m *Migrator) Status(ctx context.Context) ([]Migration, []string, error) {
	if err := m.adapter.Initialize(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize adapter: %w", err)
	}

	applied, err := m.adapter.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, mgr := range applied {
		appliedMap[mgr.ID] = true
	}

	files, err := os.ReadDir(m.policiesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read policies directory: %w", err)
	}

	var pendingFiles []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), ".csv") {
			continue
		}
		if !appliedMap[f.Name()] {
			pendingFiles = append(pendingFiles, f.Name())
		}
	}

	sort.Strings(pendingFiles)

	return applied, pendingFiles, nil
}
