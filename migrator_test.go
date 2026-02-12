package casbinMigrate_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	casbinMigrate "github.com/onlyin32bit/casbin-policy-migrate"
)

// MockAdapter is a simple in-memory adapter for testing.
type MockAdapter struct {
	Applied  []casbinMigrate.Migration
	Policies map[string]bool // key: "ptype:rule_joined"
}

func NewMockAdapter() *MockAdapter {
	return &MockAdapter{
		Policies: make(map[string]bool),
	}
}

func (m *MockAdapter) Initialize(ctx context.Context) error { return nil }

func (m *MockAdapter) GetAppliedMigrations(ctx context.Context) ([]casbinMigrate.Migration, error) {
	return m.Applied, nil
}

func (m *MockAdapter) MarkMigrationApplied(ctx context.Context, mgr casbinMigrate.Migration) error {
	mgr.AppliedAt = time.Now()
	m.Applied = append(m.Applied, mgr)
	return nil
}

func (m *MockAdapter) MarkMigrationRolledBack(ctx context.Context, mgr casbinMigrate.Migration) error {
	for i, v := range m.Applied {
		if v.ID == mgr.ID {
			m.Applied = append(m.Applied[:i], m.Applied[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockAdapter) AddPolicy(ctx context.Context, sec string, ptype string, rule []string) error {
	key := ptype + ":" + joinRule(rule)
	m.Policies[key] = true
	return nil
}

func (m *MockAdapter) RemovePolicy(ctx context.Context, sec string, ptype string, rule []string) error {
	key := ptype + ":" + joinRule(rule)
	delete(m.Policies, key)
	return nil
}

func joinRule(rule []string) string {
	res := ""
	for _, r := range rule {
		res += r + ","
	}
	return res
}

func TestMigrator_Up_Down(t *testing.T) {
	// Setup temporary directory for policies
	tmpDir, err := os.MkdirTemp("", "casbin_migrations")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test migration file
	content := `p, alice, data1, read`
	if err := os.WriteFile(filepath.Join(tmpDir, "001_init.csv"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	adapter := NewMockAdapter()
	migrator := casbinMigrate.NewMigrator(adapter, tmpDir)
	ctx := context.Background()

	// Test Up
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() failed: %v", err)
	}

	if len(adapter.Applied) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(adapter.Applied))
	}
	if !adapter.Policies["p:alice,data1,read,"] {
		t.Error("Policy not added")
	}

	// Test Down
	if err := migrator.Down(ctx, 1); err != nil {
		t.Fatalf("Down() failed: %v", err)
	}

	if len(adapter.Applied) != 0 {
		t.Errorf("Expected 0 applied migrations after rollback, got %d", len(adapter.Applied))
	}
	if adapter.Policies["p:alice,data1,read,"] {
		t.Error("Policy not removed")
	}
}
