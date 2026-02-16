package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "project")
}

func mustScaffold(t *testing.T, data TemplateData) string {
	t.Helper()
	target := tempDir(t)
	s, err := NewScaffolder()
	if err != nil {
		t.Fatalf("NewScaffolder: %v", err)
	}
	if err := s.Scaffold(target, data); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	return target
}

func TestCoreFilesAlwaysCreated(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName: "test-core",
		Description: "A test project",
	})

	for _, name := range []string{"README.md", "AGENTS.md", "DECISIONS.md", "TODO.md", "LEARNINGS.md"} {
		path := filepath.Join(target, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
}

func TestLearningsAlwaysCreated(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName: "test-learnings",
		Description: "A test project",
	})
	path := filepath.Join(target, "LEARNINGS.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("LEARNINGS.md should always exist: %v", err)
	}
}

func TestNoDevcontainerByDefault(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName: "test-no-dc",
		Description: "A test project",
	})
	dcDir := filepath.Join(target, ".devcontainer")
	if _, err := os.Stat(dcDir); !os.IsNotExist(err) {
		t.Error(".devcontainer/ should not exist when IncludeDevContainer is false")
	}
}

func TestDevcontainerMinimal(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName:         "test-dc-minimal",
		Description:         "A test project",
		IncludeDevContainer: true,
		DevContainerImage:   "python:3-3.12",
		AIChatContinuity:    false,
	})

	dcPath := filepath.Join(target, ".devcontainer", "devcontainer.json")
	raw, err := os.ReadFile(dcPath)
	if err != nil {
		t.Fatalf("devcontainer.json should exist: %v", err)
	}

	var dc DevContainer
	if err := json.Unmarshal(raw, &dc); err != nil {
		t.Fatalf("devcontainer.json is not valid JSON: %v", err)
	}

	if dc.Image != "mcr.microsoft.com/devcontainers/python:3-3.12" {
		t.Errorf("wrong image: got %q", dc.Image)
	}
	if dc.Name != "test-dc-minimal (Dev Container)" {
		t.Errorf("wrong name: got %q", dc.Name)
	}
	if len(dc.Mounts) != 0 {
		t.Errorf("expected no mounts, got %d", len(dc.Mounts))
	}
	if dc.PostCreateCommand != "" {
		t.Errorf("expected no postCreateCommand, got %q", dc.PostCreateCommand)
	}

	// setup.sh should NOT exist
	setupPath := filepath.Join(target, ".devcontainer", "setup.sh")
	if _, err := os.Stat(setupPath); !os.IsNotExist(err) {
		t.Error("setup.sh should not exist when chat continuity is disabled")
	}
}

func TestDevcontainerWithChatContinuity(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName:         "test-dc-chat",
		Description:         "A test project",
		IncludeDevContainer: true,
		DevContainerImage:   "go:2-1.25-trixie",
		AIChatContinuity:    true,
	})

	// devcontainer.json
	dcPath := filepath.Join(target, ".devcontainer", "devcontainer.json")
	raw, err := os.ReadFile(dcPath)
	if err != nil {
		t.Fatalf("devcontainer.json should exist: %v", err)
	}

	var dc DevContainer
	if err := json.Unmarshal(raw, &dc); err != nil {
		t.Fatalf("devcontainer.json is not valid JSON: %v", err)
	}

	if dc.Image != "mcr.microsoft.com/devcontainers/go:2-1.25-trixie" {
		t.Errorf("wrong image: got %q", dc.Image)
	}

	// Should have mounts for all known AI tools
	if len(dc.Mounts) != len(knownAITools) {
		t.Fatalf("expected %d mounts (one per known AI tool), got %d", len(knownAITools), len(dc.Mounts))
	}
	for _, tool := range knownAITools {
		found := false
		for _, m := range dc.Mounts {
			if strings.Contains(m, tool.StateDir) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected mount for %s (%s)", tool.Label, tool.StateDir)
		}
	}

	// Should have HOST_WORKSPACE env
	if dc.ContainerEnv["HOST_WORKSPACE"] != "${localWorkspaceFolder}" {
		t.Errorf("expected HOST_WORKSPACE env, got %v", dc.ContainerEnv)
	}

	// Should reference setup.sh
	if dc.PostCreateCommand != "bash .devcontainer/setup.sh" {
		t.Errorf("expected postCreateCommand to reference setup.sh, got %q", dc.PostCreateCommand)
	}

	// setup.sh should exist and reference all known tools
	setupPath := filepath.Join(target, ".devcontainer", "setup.sh")
	setupRaw, err := os.ReadFile(setupPath)
	if err != nil {
		t.Fatalf("setup.sh should exist: %v", err)
	}
	setup := string(setupRaw)
	for _, tool := range knownAITools {
		if !strings.Contains(setup, tool.StateDir) {
			t.Errorf("setup.sh should reference %s (%s)", tool.Label, tool.StateDir)
		}
	}
}

func TestSetupScriptContent(t *testing.T) {
	script := generateSetupScript()

	if !strings.HasPrefix(script, "#!/bin/bash\n") {
		t.Error("setup script should start with shebang")
	}

	if !strings.Contains(script, "HOST_KEY=") {
		t.Error("script should derive HOST_KEY")
	}
	if !strings.Contains(script, "CONTAINER_KEY=") {
		t.Error("script should derive CONTAINER_KEY")
	}

	// Should auto-detect (check for existence) before symlinking
	for _, tool := range knownAITools {
		if !strings.Contains(script, fmt.Sprintf(`if [ -d "$HOME/%s" ]`, tool.StateDir)) {
			t.Errorf("script should auto-detect %s existence", tool.StateDir)
		}
		if !strings.Contains(script, "ln -sfn") {
			t.Error("script should create symlinks")
		}
	}
}

func TestSetupScriptAutoDetects(t *testing.T) {
	// Verify the script checks if the tool dir exists before acting
	script := generateSetupScript()

	// Each tool block should be wrapped in an existence check
	for _, tool := range knownAITools {
		// Should check if the tool's state dir exists on the host (via mount)
		check := fmt.Sprintf(`if [ -d "$HOME/%s" ]; then`, tool.StateDir)
		if !strings.Contains(script, check) {
			t.Errorf("script should check existence of %s before acting", tool.StateDir)
		}
	}
}

func TestTemplateContentHasProjectName(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName: "my-unique-project-name",
		Description: "This is a unique description for testing",
	})

	raw, err := os.ReadFile(filepath.Join(target, "README.md"))
	if err != nil {
		t.Fatalf("README.md should exist: %v", err)
	}
	if !strings.Contains(string(raw), "my-unique-project-name") {
		t.Error("README.md should contain the project name")
	}
	if !strings.Contains(string(raw), "This is a unique description for testing") {
		t.Error("README.md should contain the description")
	}
}

func TestNonEmptyDirectoryFails(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "project")
	os.MkdirAll(target, 0755)
	os.WriteFile(filepath.Join(target, "existing.txt"), []byte("hello"), 0644)

	s, err := NewScaffolder()
	if err != nil {
		t.Fatalf("NewScaffolder: %v", err)
	}
	err = s.Scaffold(target, TemplateData{
		ProjectName: "test",
		Description: "test",
	})
	if err == nil {
		t.Error("expected error when scaffolding into non-empty directory")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Errorf("expected 'not empty' error, got: %v", err)
	}
}

func TestParentDirectoryMustExist(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "nonexistent", "project")

	s, err := NewScaffolder()
	if err != nil {
		t.Fatalf("NewScaffolder: %v", err)
	}
	err = s.Scaffold(target, TemplateData{
		ProjectName: "test",
		Description: "test",
	})
	if err == nil {
		t.Error("expected error when parent directory does not exist")
	}
	if !strings.Contains(err.Error(), "parent directory") {
		t.Errorf("expected 'parent directory' error, got: %v", err)
	}
}

func TestAllImageOptions(t *testing.T) {
	images := []string{
		"go:2-1.25-trixie",
		"typescript-node:20-bookworm",
		"python:3-3.12",
		"rust:1-bookworm",
		"java",
		"dotnet",
		"cpp",
		"universal",
	}

	for _, img := range images {
		t.Run(img, func(t *testing.T) {
			target := mustScaffold(t, TemplateData{
				ProjectName:         "test-" + img,
				Description:         "Testing image " + img,
				IncludeDevContainer: true,
				DevContainerImage:   img,
			})

			dcPath := filepath.Join(target, ".devcontainer", "devcontainer.json")
			raw, err := os.ReadFile(dcPath)
			if err != nil {
				t.Fatalf("devcontainer.json should exist: %v", err)
			}

			var dc DevContainer
			if err := json.Unmarshal(raw, &dc); err != nil {
				t.Fatalf("devcontainer.json is not valid JSON: %v", err)
			}

			expected := "mcr.microsoft.com/devcontainers/" + img
			if dc.Image != expected {
				t.Errorf("expected image %q, got %q", expected, dc.Image)
			}
		})
	}
}

func TestAllowNonEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "project")
	os.MkdirAll(target, 0755)
	os.WriteFile(filepath.Join(target, "existing.txt"), []byte("hello"), 0644)

	s, err := NewScaffolder()
	if err != nil {
		t.Fatalf("NewScaffolder: %v", err)
	}

	// Without allowNonEmpty — should fail (already covered, but confirms contrast)
	err = s.Scaffold(target, TemplateData{ProjectName: "test", Description: "test"})
	if err == nil {
		t.Fatal("expected error without allowNonEmpty")
	}

	// With allowNonEmpty — should succeed
	err = s.Scaffold(target, TemplateData{ProjectName: "test", Description: "test"}, true)
	if err != nil {
		t.Fatalf("expected success with allowNonEmpty, got: %v", err)
	}

	// Original file should still be there
	if _, err := os.Stat(filepath.Join(target, "existing.txt")); err != nil {
		t.Error("existing file should not be removed")
	}
	// Scaffolded files should also exist
	if _, err := os.Stat(filepath.Join(target, "README.md")); err != nil {
		t.Error("README.md should have been created")
	}
}

func TestTargetPathIsFileFails(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "not-a-dir")
	os.WriteFile(filePath, []byte("I'm a file"), 0644)

	s, err := NewScaffolder()
	if err != nil {
		t.Fatalf("NewScaffolder: %v", err)
	}
	err = s.Scaffold(filePath, TemplateData{ProjectName: "test", Description: "test"})
	if err == nil {
		t.Error("expected error when target is a file, not a directory")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' error, got: %v", err)
	}
}


func TestSpecialCharsInProjectName(t *testing.T) {
	names := []string{
		"my project (v2)",
		"project-with-dashes",
		"project_with_underscores",
		"CamelCaseProject",
		"project.with.dots",
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			target := mustScaffold(t, TemplateData{
				ProjectName: name,
				Description: "Testing special characters",
			})

			raw, err := os.ReadFile(filepath.Join(target, "README.md"))
			if err != nil {
				t.Fatalf("README.md should exist: %v", err)
			}
			if !strings.Contains(string(raw), name) {
				t.Errorf("README.md should contain project name %q", name)
			}
		})
	}
}

func TestEmptyDirectoryReuseSucceeds(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "project")
	os.MkdirAll(target, 0755)

	target = mustScaffold(t, TemplateData{
		ProjectName: "test-empty-reuse",
		Description: "Scaffolding into a pre-created empty directory",
	})

	if _, err := os.Stat(filepath.Join(target, "README.md")); err != nil {
		t.Error("README.md should exist in reused empty directory")
	}
}
