//go:build integration

package repo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// testPool connects to a real Postgres instance for integration tests.
//
// It targets the local docker-compose database by default (see
// docker-compose.yaml / Makefile `up` + `migrate`), overridable via
// TEST_DB_* env vars for CI. The test is skipped — not failed — when the
// database isn't reachable, since these tests exercise real SQL/constraints
// and aren't meant to run as part of the default `go test ./...`.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	host := envOr("TEST_DB_HOST", "127.0.0.1")
	port := envOr("TEST_DB_PORT", "5438")
	user := envOr("TEST_DB_USER", "tracker")
	pass := envOr("TEST_DB_PASS", "tracker")
	name := envOr("TEST_DB_NAME", "tracker")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("integration: cannot create pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("integration: postgres not reachable at %s:%s (start it with `make up && make migrate`): %v", host, port, err)
	}

	t.Cleanup(pool.Close)
	return pool
}

// testUser inserts a throwaway user row and returns its id. The row (and
// anything FK-cascaded from it, e.g. user_custom_timers/user_timer_settings
// rows) is deleted when the test finishes.
func testUser(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	ctx := context.Background()
	tgUserID := time.Now().UnixNano()

	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO users (tg_user_id, language, timezone) VALUES ($1, 'en', 'UTC') RETURNING id;`,
		tgUserID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1;`, id)
	})

	return id
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
