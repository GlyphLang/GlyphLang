package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newInitCmd creates a fresh init cobra command for testing.
func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "init [name]",
		Args: cobra.ExactArgs(1),
		RunE: runInit,
	}
	cmd.Flags().StringP("template", "t", "crud", "Project template (crud, rest-api, hello-world)")
	return cmd
}

func TestRunInit_CreatesProjectStructure(t *testing.T) {
	dir := t.TempDir()
	projName := "test-project"
	projDir := filepath.Join(dir, projName)

	cmd := newInitCmd()
	cmd.SetArgs([]string{projName})

	// Run from the temp directory so the project is created there
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	if err := cmd.Execute(); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(projDir)
	if err != nil {
		t.Fatalf("project directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("project path is not a directory")
	}

	// Verify main.glyph exists
	mainFile := filepath.Join(projDir, "main.glyph")
	if _, err := os.Stat(mainFile); err != nil {
		t.Fatalf("main.glyph not created: %v", err)
	}

	// Verify .gitignore exists
	gitignoreFile := filepath.Join(projDir, ".gitignore")
	gitignoreContent, err := os.ReadFile(gitignoreFile)
	if err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}

	// Verify .gitignore contents
	gitignoreStr := string(gitignoreContent)
	for _, expected := range []string{"*.glyphc", "generated/", ".env", ".glyph/"} {
		if !strings.Contains(gitignoreStr, expected) {
			t.Errorf(".gitignore missing entry: %s", expected)
		}
	}
}

func TestRunInit_CrudTemplate(t *testing.T) {
	dir := t.TempDir()
	projName := "crud-project"
	projDir := filepath.Join(dir, projName)

	cmd := newInitCmd()
	cmd.SetArgs([]string{projName, "--template", "crud"})

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	if err := cmd.Execute(); err != nil {
		t.Fatalf("runInit with crud template failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(projDir, "main.glyph"))
	if err != nil {
		t.Fatalf("failed to read main.glyph: %v", err)
	}

	source := string(content)

	// Verify project name appears in template header
	if !strings.Contains(source, projName) {
		t.Error("template does not contain project name")
	}

	// Verify expected CRUD route patterns
	expectedPatterns := []string{
		"@ GET /api/items",
		"@ GET /api/items/:id",
		"@ POST /api/items",
		"@ PUT /api/items/:id",
		"@ DELETE /api/items/:id",
		"@ GET /health",
	}
	for _, pattern := range expectedPatterns {
		if !strings.Contains(source, pattern) {
			t.Errorf("CRUD template missing route pattern: %s", pattern)
		}
	}

	// Verify type definitions
	if !strings.Contains(source, ": Item {") {
		t.Error("CRUD template missing Item type definition")
	}
	if !strings.Contains(source, ": CreateItemInput {") {
		t.Error("CRUD template missing CreateItemInput type definition")
	}
	if !strings.Contains(source, ": UpdateItemInput {") {
		t.Error("CRUD template missing UpdateItemInput type definition")
	}
}

func TestRunInit_HelloWorldTemplate(t *testing.T) {
	dir := t.TempDir()
	projName := "hello-project"
	projDir := filepath.Join(dir, projName)

	cmd := newInitCmd()
	cmd.SetArgs([]string{projName, "--template", "hello-world"})

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	if err := cmd.Execute(); err != nil {
		t.Fatalf("runInit with hello-world template failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(projDir, "main.glyph"))
	if err != nil {
		t.Fatalf("failed to read main.glyph: %v", err)
	}

	if !strings.Contains(string(content), "Hello World") {
		t.Error("hello-world template missing expected content")
	}
}

func TestRunInit_RestAPITemplate(t *testing.T) {
	dir := t.TempDir()
	projName := "api-project"
	projDir := filepath.Join(dir, projName)

	cmd := newInitCmd()
	cmd.SetArgs([]string{projName, "--template", "rest-api"})

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	if err := cmd.Execute(); err != nil {
		t.Fatalf("runInit with rest-api template failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(projDir, "main.glyph"))
	if err != nil {
		t.Fatalf("failed to read main.glyph: %v", err)
	}

	if !strings.Contains(string(content), "REST API") {
		t.Error("rest-api template missing expected content")
	}
}

func TestRunInit_UnknownTemplate(t *testing.T) {
	dir := t.TempDir()

	cmd := newInitCmd()
	cmd.SetArgs([]string{"bad-project", "--template", "nonexistent"})

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}

	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("error message should mention unknown template, got: %s", err.Error())
	}
}

func TestRunInit_DirectoryAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	projName := "existing-project"
	projDir := filepath.Join(dir, projName)

	// Pre-create the directory
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed to pre-create directory: %v", err)
	}

	cmd := newInitCmd()
	cmd.SetArgs([]string{projName})

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	// Should succeed even if directory already exists (MkdirAll)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("runInit should succeed when directory exists: %v", err)
	}

	// Verify files were created inside existing directory
	if _, err := os.Stat(filepath.Join(projDir, "main.glyph")); err != nil {
		t.Error("main.glyph not created in existing directory")
	}
	if _, err := os.Stat(filepath.Join(projDir, ".gitignore")); err != nil {
		t.Error(".gitignore not created in existing directory")
	}
}

func TestGetRestCrudTemplate(t *testing.T) {
	tmpl := getRestCrudTemplate("my-service")

	if !strings.Contains(tmpl, "my-service") {
		t.Error("template should contain the project name")
	}

	// Ensure fmt.Sprintf correctly resolved the format placeholder
	if strings.Contains(tmpl, "%"+"s") {
		t.Error("template should not contain unresolved format placeholder")
	}

	// The dependency injection line should use single %
	if !strings.Contains(tmpl, "% db: Database") {
		t.Error("template should contain '% db: Database' for dependency injection")
	}
}
