package migrator

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func TestRenderStatusTable(t *testing.T) {
	batchOne := 1
	statuses := []MigrationStatus{
		{Ran: true, Migration: "2026_09_03_000001_create_employees_table", Batch: &batchOne},
		{Ran: false, Migration: "2026_09_03_000002_create_users_table", Batch: nil},
	}

	rendered := RenderStatusTable(statuses)
	if !strings.Contains(rendered, "Yes") || !strings.Contains(rendered, "Pending") {
		t.Errorf("Expected status table to contain Yes and Pending, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "2026_09_03_000001_create_employees_table") {
		t.Errorf("Expected status table to contain migration name, got:\n%s", rendered)
	}

	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		t.Fatalf("Expected rendered table to contain lines")
	}
	firstLineLen := len(lines[0])
	for i, line := range lines {
		if len(line) != firstLineLen {
			t.Errorf("Line %d length %d != first line length %d: %q", i, len(line), firstLineLen, line)
		}
	}
}

func TestRenderStatusTableEmpty(t *testing.T) {
	rendered := RenderStatusTable(nil)
	if rendered != "No migrations found." {
		t.Errorf("Expected 'No migrations found.', got %q", rendered)
	}
}

func TestMakeMigration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "migrations-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	upPath, downPath, err := MakeMigration(tempDir, "create_roles_table")
	if err != nil {
		t.Fatalf("MakeMigration failed: %v", err)
	}

	if _, err := os.Stat(upPath); os.IsNotExist(err) {
		t.Errorf("Expected up migration file to exist: %s", upPath)
	}
	if _, err := os.Stat(downPath); os.IsNotExist(err) {
		t.Errorf("Expected down migration file to exist: %s", downPath)
	}

	upContent, _ := os.ReadFile(upPath)
	if !strings.Contains(string(upContent), "CREATE TABLE IF NOT EXISTS `roles`") {
		t.Errorf("Expected table creation template for roles, got:\n%s", string(upContent))
	}

	downContent, _ := os.ReadFile(downPath)
	if !strings.Contains(string(downContent), "DROP TABLE IF EXISTS `roles`") {
		t.Errorf("Expected table drop template for roles, got:\n%s", string(downContent))
	}
}

func TestMakeMigrationGeneric(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "migrations-test-generic-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	upPath, downPath, err := MakeMigration(tempDir, "add_status_to_users")
	if err != nil {
		t.Fatalf("MakeMigration failed: %v", err)
	}

	upContent, _ := os.ReadFile(upPath)
	if !strings.Contains(string(upContent), "-- Migration up SQL statement") {
		t.Errorf("Expected generic up template, got:\n%s", string(upContent))
	}

	downContent, _ := os.ReadFile(downPath)
	if !strings.Contains(string(downContent), "-- Migration down SQL statement") {
		t.Errorf("Expected generic down template, got:\n%s", string(downContent))
	}
}

func TestExtractTableName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"create_users_table", "users"},
		{"create_employee_records_table", "employee_records"},
		{"add_index_to_users", ""},
	}

	for _, tt := range tests {
		got := extractTableName(tt.input)
		if got != tt.expected {
			t.Errorf("extractTableName(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestGetAvailableMigrations(t *testing.T) {
	mockFS := fstest.MapFS{
		"2026_09_03_000002_create_users_table.up.sql":       &fstest.MapFile{Data: []byte("CREATE TABLE users;")},
		"2026_09_03_000002_create_users_table.down.sql":     &fstest.MapFile{Data: []byte("DROP TABLE users;")},
		"2026_09_03_000001_create_employees_table.up.sql":   &fstest.MapFile{Data: []byte("CREATE TABLE employees;")},
		"2026_09_03_000001_create_employees_table.down.sql": &fstest.MapFile{Data: []byte("DROP TABLE employees;")},
		"readme.txt":                                        &fstest.MapFile{Data: []byte("ignore me")},
	}

	m := New(nil, mockFS)
	migrations, err := m.GetAvailableMigrations()
	if err != nil {
		t.Fatalf("GetAvailableMigrations failed: %v", err)
	}

	expected := []string{
		"2026_09_03_000001_create_employees_table",
		"2026_09_03_000002_create_users_table",
	}

	if len(migrations) != len(expected) {
		t.Fatalf("Expected %d migrations, got %d", len(expected), len(migrations))
	}

	for i, name := range expected {
		if migrations[i] != name {
			t.Errorf("At index %d: expected %s, got %s", i, name, migrations[i])
		}
	}
}
