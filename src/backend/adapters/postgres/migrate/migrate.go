package migrate

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var migrations embed.FS

// Up executes all embedded SQL migrations in filename order in one transaction.
func Up(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.ReadDir(migrations, "sql")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Serialize schema changes across processes/packages sharing the same database.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(70700402)`); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := "sql/" + entry.Name()
		query, err := migrations.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.Exec(ctx, string(query)); err != nil {
			return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}
