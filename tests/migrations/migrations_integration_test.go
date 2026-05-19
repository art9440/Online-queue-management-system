//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const latestMigrationVersion int64 = 20260518100000

func TestMigrations_WhenAppliedToPostgres_ShouldCreateSeededSchemaAndRollback(t *testing.T) {
	dsn := os.Getenv("MIGRATIONS_TEST_DSN")
	if dsn == "" {
		t.Skip("MIGRATIONS_TEST_DSN is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	adminDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}
	defer adminDB.Close()

	schema := fmt.Sprintf("migration_test_%d", time.Now().UnixNano())
	if _, err := adminDB.ExecContext(ctx, `CREATE SCHEMA `+quoteIdentifier(schema)); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, _ = adminDB.ExecContext(cleanupCtx, `DROP SCHEMA IF EXISTS `+quoteIdentifier(schema)+` CASCADE`)
	})

	db, err := sql.Open("pgx", dsnWithSearchPath(t, dsn, schema))
	if err != nil {
		t.Fatalf("open migration db: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}

	if err := goose.UpContext(ctx, db, migrationsDir(t)); err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
		t.Fatalf("apply migrations: %v", err)
	}

	version, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		t.Fatalf("get db version: %v", err)
	}
	if version != latestMigrationVersion {
		t.Fatalf("expected latest version %d, got %d", latestMigrationVersion, version)
	}

	assertTablesExist(t, ctx, db, schema, []string{
		"businesses",
		"roles",
		"branches",
		"users",
		"employees",
		"employee_schedules",
		"services",
		"employee_services",
		"clients",
		"appointments",
	})
	assertColumnType(t, ctx, db, schema, "branches", "timezone", "text")
	assertColumnType(t, ctx, db, schema, "employee_schedules", "starts_at", "timestamp with time zone")
	assertColumnType(t, ctx, db, schema, "employee_schedules", "ends_at", "timestamp with time zone")

	assertCount(t, ctx, db, "roles", 4)
	assertCount(t, ctx, db, "branches", 3)
	assertCount(t, ctx, db, "employees", 6)
	assertCount(t, ctx, db, "services", 9)
	assertCount(t, ctx, db, "employee_services", 15)
	assertCount(t, ctx, db, "employee_schedules", 24)
	assertDemoBusinessSlug(t, ctx, db)

	if err := goose.DownToContext(ctx, db, migrationsDir(t), 0); err != nil {
		t.Fatalf("rollback migrations: %v", err)
	}
	assertTablesDoNotExist(t, ctx, db, schema, []string{
		"businesses",
		"roles",
		"branches",
		"users",
		"employees",
		"employee_schedules",
		"services",
		"employee_services",
		"clients",
		"appointments",
	})
}

func assertDemoBusinessSlug(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	var slug string
	err := db.QueryRowContext(ctx, `
		SELECT registration_slug
		FROM businesses
		WHERE name = 'Demo Business'
		  AND type = 'service_company'
	`).Scan(&slug)
	if err != nil {
		t.Fatalf("query demo business slug: %v", err)
	}

	if slug != "demo-business" {
		t.Fatalf("expected demo business slug demo-business, got %q", slug)
	}
}

func dsnWithSearchPath(t *testing.T, dsn, schema string) string {
	t.Helper()

	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}

	query := parsed.Query()
	query.Set("options", "-c search_path="+schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func assertTablesExist(t *testing.T, ctx context.Context, db *sql.DB, schema string, tables []string) {
	t.Helper()

	for _, table := range tables {
		if !tableExistsInSchema(t, ctx, db, schema, table) {
			t.Fatalf("expected table %s.%s to exist", schema, table)
		}
	}
}

func assertTablesDoNotExist(t *testing.T, ctx context.Context, db *sql.DB, schema string, tables []string) {
	t.Helper()

	for _, table := range tables {
		if tableExistsInSchema(t, ctx, db, schema, table) {
			t.Fatalf("expected table %s.%s to be rolled back", schema, table)
		}
	}
}

func tableExistsInSchema(t *testing.T, ctx context.Context, db *sql.DB, schema, table string) bool {
	t.Helper()

	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = $1
			  AND table_name = $2
		)
	`, schema, table).Scan(&exists)
	if err != nil {
		t.Fatalf("check table %s.%s: %v", schema, table, err)
	}
	return exists
}

func assertColumnType(t *testing.T, ctx context.Context, db *sql.DB, schema, table, column, expectedType string) {
	t.Helper()

	var dataType string
	err := db.QueryRowContext(ctx, `
		SELECT data_type
		FROM information_schema.columns
		WHERE table_schema = $1
		  AND table_name = $2
		  AND column_name = $3
	`, schema, table, column).Scan(&dataType)
	if err != nil {
		t.Fatalf("read column type %s.%s.%s: %v", schema, table, column, err)
	}
	if dataType != expectedType {
		t.Fatalf("expected %s.%s.%s to be %q, got %q", schema, table, column, expectedType, dataType)
	}
}

func assertCount(t *testing.T, ctx context.Context, db *sql.DB, table string, expected int) {
	t.Helper()

	var actual int
	query := `SELECT count(*) FROM ` + quoteIdentifier(table)
	if err := db.QueryRowContext(ctx, query).Scan(&actual); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if actual != expected {
		t.Fatalf("expected %s count %d, got %d", table, expected, actual)
	}
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
