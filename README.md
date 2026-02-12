# Casbin Policy Migrate

[![Go Report Card](https://goreportcard.com/badge/github.com/onlyin32bit/casbin-policy-migrate)](https://goreportcard.com/report/github.com/onlyin32bit/casbin-policy-migrate)
[![Godoc](https://godoc.org/github.com/onlyin32bit/casbin-policy-migrate?status.svg)](https://godoc.org/github.com/onlyin32bit/casbin-policy-migrate)

A migration system for managing Casbin policies with version control, similar to database migrations.

## Features

- **CSV-based migrations** - Simple, readable migration files
- **Up/Down migrations** - Apply and rollback policy changes
- **Adapter system** - Support for multiple storage backends (PostgreSQL, MySQL, etc.)
- **CLI tool** - Manage migrations from the command line

## Installation

### CLI Tool

```bash
go install github.com/onlyin32bit/casbin-policy-migrate/cli/casbin-migrate@latest
```

### Package

```bash
go get github.com/onlyin32bit/casbin-policy-migrate
```

## Quick Start

### 1. Initialize Project Structure

```bash
casbin-migrate init
```

This creates:

```txt
casbin/
├── model.conf
└── policy_migrations/
```

### 2. Create a Migration

```bash
casbin-migrate new add_admin_policies
```

This creates a timestamped migration file: `./casbin/policy_migrations/20260212120000_add_admin_policies.csv`

Alternatively, create a sequential migration with the `--sequential` flag:

```bash
casbin-migrate new add_admin_policies --sequential
```

This creates: `./casbin/policy_migrations/0001_add_admin_policies.csv`

### 3. Edit the Migration File

```csv
p, admin, data:*, read
p, admin, data:*, write
g, user:1, admin
```

### 4. Run Migrations

```bash
casbin-migrate up --db-url "postgres://user:pass@localhost:5432/dbname"
```

Or use environment variable:

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/dbname"
casbin-migrate up
```

## Migration File Format

Migration files are CSV files with the following format:

### Adding Policies

```csv
p, subject, object, action
g, user, role
```

### Removing Policies

Prefix the policy type with `-`:

```csv
-p, subject, object, action
-g, user, role
```

### Example Migration

```csv
# Add new permissions
p, system:root, user:*, read
p, system:root, user:*, write

# Add role assignment
g, user:1, system:root

# Remove old permission
-p, system:root, user:*, delete
```

### Rules

- **Add operations** (`p`, `g`) - Must succeed. If the policy already exists, it will fail.
- **Remove operations** (`-p`, `-g`) - Ignore if not exist. Safe to remove non-existent policies.

## CLI Commands

### `init`

Initialize the migration directory structure:

```bash
casbin-migrate init
```

### `new <name>`

Create a new migration file:

```bash
casbin-migrate new add_user_permissions
```

**Flags:**

- `--sequential` - Use sequential numbering (0001, 0002, etc) instead of timestamp prefix
- `--migration-dir` - Specify custom migration directory (default: `./casbin/policy_migrations`)

**Examples:**

```bash
# Create timestamped migration (default)
casbin-migrate new add_user_permissions
# Creates: 20260212120000_add_user_permissions.csv

# Create sequential migration
casbin-migrate new add_user_permissions --sequential
# Creates: 0005_add_user_permissions.csv

# Use custom migration directory
casbin-migrate new add_user_permissions --migration-dir "./migrations"
```

### `up`

Apply all pending migrations:

```bash
casbin-migrate up --db-url "postgres://..."
```

### `down [n]`

Rollback the last `n` migrations (default: 1):

```bash
casbin-migrate down      # Rollback last migration
casbin-migrate down 3    # Rollback last 3 migrations
casbin-migrate down 0    # Rollback all migrations
```

### `status`

Show migration status:

```bash
casbin-migrate status --db-url "postgres://..."
```

## Programmatic Usage

### Basic Example

```go
package main

import (
    "context"
    "log"

    "github.com/jackc/pgx/v5/pgxpool"
    casbinMigrate "github.com/onlyin32bit/casbin-policy-migrate"
    "github.com/onlyin32bit/casbin-policy-migrate/adapter/pgx"
)

func main() {
    // Connect to database
    pool, err := pgxpool.New(context.Background(), "postgres://...")
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // Create adapter and migrator
    adapter := pgx.NewAdapter(pool)
    migrator := casbinMigrate.NewMigrator(adapter, "casbin/policy_migrations")

    // Run migrations
    ctx := context.Background()
    if err := migrator.Up(ctx); err != nil {
        log.Fatal(err)
    }

    log.Println("Migrations applied successfully!")
}
```

### Check Migration Status

```go
applied, pending, err := migrator.Status(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Applied: %d, Pending: %d\n", len(applied), len(pending))
```

### Rollback Migrations

```go
// Rollback last migration
if err := migrator.Down(ctx, 1); err != nil {
    log.Fatal(err)
}
```

## Custom Adapters

You can create custom adapters for different storage backends by implementing the `migration.Adapter` interface:

```go
type Adapter interface {
    // Migration History Management
    Initialize(ctx context.Context) error
    GetAppliedMigrations(ctx context.Context) ([]Migration, error)
    MarkMigrationApplied(ctx context.Context, migration Migration) error
    MarkMigrationRolledBack(ctx context.Context, migration Migration) error

    // Policy Management
    AddPolicy(ctx context.Context, sec string, ptype string, rule []string) error
    RemovePolicy(ctx context.Context, sec string, ptype string, rule []string) error
}
```

### Example: Custom Adapter

```go
type MyAdapter struct {
    db *sql.DB
}

func (a *MyAdapter) Initialize(ctx context.Context) error {
    // Create migrations table
    _, err := a.db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS casbin_policy_migrations (
            id VARCHAR(255) PRIMARY KEY,
            checksum VARCHAR(255),
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    return err
}

func (a *MyAdapter) AddPolicy(ctx context.Context, sec string, ptype string, rule []string) error {
    // Insert policy into casbin_rule table
    // Implementation depends on your schema
    return nil
}

// Implement other methods...
```

## Project Structure

```txt
.
├── adapter/
│   └── pgx/              # PostgreSQL adapter using pgx
├── cli/
│   ├── cmd/              # CLI commands
│   └── main.go           # CLI entry point
├── example/              # Example usage
│   ├── casbin/
│   │   ├── model.conf
│   │   └── policy_migrations/
│   │       ├── 0001_init.csv
│   │       ├── 0002_auth_sessions.csv
│   │       ├── 0003_root_user.csv
│   │       └── 0004_system_root_fix.csv
│   └── main.go
├── pkg/
│   └── migration/        # Core migration logic
│       ├── adapter.go    # Adapter interface
│       ├── migrator.go   # Migration engine
│       ├── parser.go     # CSV parser
│       └── types.go      # Type definitions
└── README.md
```

## Testing

Run tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
