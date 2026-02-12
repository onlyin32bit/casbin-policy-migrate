package casbinMigrate_test

import (
	"strings"
	"testing"

	casbinMigrate "github.com/onlyin32bit/casbin-policy-migrate"
)

func TestParseReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []casbinMigrate.Operation
		wantErr  bool
	}{
		{
			name: "Basic Add",
			input: `p, alice, data1, read
g, alice, admin`,
			expected: []casbinMigrate.Operation{
				{Type: casbinMigrate.OperationAdd, Sec: "p", PType: "p", Rule: []string{"alice", "data1", "read"}},
				{Type: casbinMigrate.OperationAdd, Sec: "g", PType: "g", Rule: []string{"alice", "admin"}},
			},
		},
		{
			name: "Basic Remove",
			input: `-p, bob, data2, write
-g, bob, user`,
			expected: []casbinMigrate.Operation{
				{Type: casbinMigrate.OperationRemove, Sec: "p", PType: "p", Rule: []string{"bob", "data2", "write"}},
				{Type: casbinMigrate.OperationRemove, Sec: "g", PType: "g", Rule: []string{"bob", "user"}},
			},
		},
		{
			name: "Mixed and Comments/Empty Lines",
			input: `
p, role:admin, data:nested, read

-p, old:role, data:old, write
`,
			expected: []casbinMigrate.Operation{
				{Type: casbinMigrate.OperationAdd, Sec: "p", PType: "p", Rule: []string{"role:admin", "data:nested", "read"}},
				{Type: casbinMigrate.OperationRemove, Sec: "p", PType: "p", Rule: []string{"old:role", "data:old", "write"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := casbinMigrate.ParseReader(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.expected) {
				t.Errorf("ParseReader() got %d ops, want %d", len(got), len(tt.expected))
				return
			}

			for i := range got {
				if got[i].Type != tt.expected[i].Type {
					t.Errorf("op[%d].Type = %v, want %v", i, got[i].Type, tt.expected[i].Type)
				}
				if got[i].Sec != tt.expected[i].Sec {
					t.Errorf("op[%d].Sec = %v, want %v", i, got[i].Sec, tt.expected[i].Sec)
				}
				if got[i].PType != tt.expected[i].PType {
					t.Errorf("op[%d].PType = %v, want %v", i, got[i].PType, tt.expected[i].PType)
				}
				if len(got[i].Rule) != len(tt.expected[i].Rule) {
					t.Errorf("op[%d].Rule length = %d, want %d", i, len(got[i].Rule), len(tt.expected[i].Rule))
				} else {
					for j, v := range got[i].Rule {
						if v != tt.expected[i].Rule[j] {
							t.Errorf("op[%d].Rule[%d] = %s, want %s", i, j, v, tt.expected[i].Rule[j])
						}
					}
				}
			}
		})
	}
}
