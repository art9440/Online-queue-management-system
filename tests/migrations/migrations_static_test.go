package migrations_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestMigrationFiles_WhenCollectedByGoose_ShouldHaveUniqueIncreasingVersions(t *testing.T) {
	migrationsDir := migrationsDir(t)

	migrations, err := goose.CollectMigrations(migrationsDir, 0, goose.MaxVersion)
	if err != nil {
		t.Fatalf("collect migrations: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected at least one migration")
	}

	versions := make([]int64, 0, len(migrations))
	seen := make(map[int64]string, len(migrations))
	for _, migration := range migrations {
		if previous, ok := seen[migration.Version]; ok {
			t.Fatalf("duplicate migration version %d in %q and %q", migration.Version, previous, migration.Source)
		}
		seen[migration.Version] = migration.Source
		versions = append(versions, migration.Version)
	}

	if !sort.SliceIsSorted(versions, func(i, j int) bool { return versions[i] < versions[j] }) {
		t.Fatalf("migration versions are not increasing: %v", versions)
	}
}

func TestMigrationFiles_WhenInspected_ShouldHaveGooseUpAndDownSections(t *testing.T) {
	for _, path := range migrationFiles(t) {
		content := readFile(t, path)

		assertContains(t, path, content, "-- +goose Up")
		assertContains(t, path, content, "-- +goose Down")

		up := sectionBetween(content, "-- +goose Up", "-- +goose Down")
		if strings.TrimSpace(removeGooseDirectives(up)) == "" {
			t.Fatalf("%s has an empty Up section", filepath.Base(path))
		}

		down := sectionAfter(content, "-- +goose Down")
		if strings.TrimSpace(removeGooseDirectives(down)) == "" {
			t.Fatalf("%s has an empty Down section", filepath.Base(path))
		}
	}
}

func TestSeedMigrations_WhenInspected_ShouldBeSafeToReapply(t *testing.T) {
	tests := []struct {
		file     string
		patterns []string
	}{
		{
			file:     "20260326103218_seed_roles.sql",
			patterns: []string{"ON CONFLICT (name) DO NOTHING"},
		},
		{
			file: "20260420193000_seed_mock_branches_employees_services.sql",
			patterns: []string{
				"WHERE NOT EXISTS",
				"FROM branches br",
				"FROM employees e",
				"FROM services s",
				"FROM employee_services es",
				"FROM employee_schedules es",
			},
		},
	}

	dir := migrationsDir(t)
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			content := readFile(t, filepath.Join(dir, tt.file))
			for _, pattern := range tt.patterns {
				assertContains(t, tt.file, content, pattern)
			}
		})
	}
}

func TestSchemaMigrations_WhenInspected_ShouldKeepCriticalConstraints(t *testing.T) {
	tests := []struct {
		file     string
		patterns []string
	}{
		{
			file: "20260326103157_init.sql",
			patterns: []string{
				"CONSTRAINT fk_users_role",
				"CONSTRAINT fk_users_business",
				"CONSTRAINT chk_employee_schedules_day_of_week",
				"CONSTRAINT chk_employee_schedules_time_range",
				"CONSTRAINT chk_services_duration",
				"CONSTRAINT chk_appointments_status",
				"CREATE INDEX idx_appointments_employee_start_time",
			},
		},
		{
			file: "20260326104048_add_link_to_businesses.sql",
			patterns: []string{
				"ADD COLUMN registration_slug",
				"CREATE UNIQUE INDEX businesses_registration_slug_uindex",
				"WHERE registration_slug IS NOT NULL",
			},
		},
		{
			file: "20260420161500_update_employee_schedules_to_timestamptz.sql",
			patterns: []string{
				"ADD COLUMN timezone text NOT NULL DEFAULT 'UTC'",
				"starts_at    timestamptz NOT NULL",
				"ends_at      timestamptz NOT NULL",
				"CONSTRAINT chk_employee_schedules_time_range",
				"CREATE INDEX idx_employee_schedules_employee_starts_at",
			},
		},
	}

	dir := migrationsDir(t)
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			content := readFile(t, filepath.Join(dir, tt.file))
			for _, pattern := range tt.patterns {
				assertContains(t, tt.file, content, pattern)
			}
		})
	}
}

func migrationFiles(t *testing.T) []string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(migrationsDir(t), "*.sql"))
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected migration sql files")
	}
	sort.Strings(files)
	return files
}

func migrationsDir(t *testing.T) string {
	t.Helper()

	dir := repoRoot(t)
	return filepath.Join(dir, "migrations")
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if fileExists(filepath.Join(dir, "go.mod")) && dirExists(filepath.Join(dir, "migrations")) {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repository root with go.mod and migrations directory was not found")
		}
		dir = parent
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func assertContains(t *testing.T, filename, content, pattern string) {
	t.Helper()

	if !strings.Contains(content, pattern) {
		t.Fatalf("%s does not contain %q", filepath.Base(filename), pattern)
	}
}

func sectionBetween(content, start, end string) string {
	startIdx := strings.Index(content, start)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(start)

	endIdx := strings.Index(content[startIdx:], end)
	if endIdx == -1 {
		return content[startIdx:]
	}
	return content[startIdx : startIdx+endIdx]
}

func sectionAfter(content, marker string) string {
	idx := strings.Index(content, marker)
	if idx == -1 {
		return ""
	}
	return content[idx+len(marker):]
}

func removeGooseDirectives(content string) string {
	lines := strings.Split(content, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "-- +goose") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
