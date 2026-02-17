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

	for _, name := range []string{"README.md", "AGENTS.md", "DECISIONS.md", "TODO.md", "LEARNINGS.md", ".gitignore", ".editorconfig"} {
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

	if dc.Build.Dockerfile != "Dockerfile" {
		t.Errorf("wrong build.dockerfile: got %q", dc.Build.Dockerfile)
	}
	if dc.Name != "test-dc-minimal (Dev Container)" {
		t.Errorf("wrong name: got %q", dc.Name)
	}
	if len(dc.Mounts) != 2 {
		t.Errorf("expected 2 mounts (gh credentials + extensions volume), got %d", len(dc.Mounts))
	}
	if !strings.Contains(dc.PostCreateCommand, "ln -sfn") || !strings.Contains(dc.PostCreateCommand, ".vscode-extensions-cache") {
		t.Errorf("expected postCreateCommand with extensions symlink, got %q", dc.PostCreateCommand)
	}
	if dc.ContainerEnv["GH_TOKEN"] != "${localEnv:GH_TOKEN}" {
		t.Errorf("expected GH_TOKEN env forwarding, got %v", dc.ContainerEnv)
	}

	// Dockerfile should exist and reference the correct image
	dfPath := filepath.Join(target, ".devcontainer", "Dockerfile")
	dfRaw, err := os.ReadFile(dfPath)
	if err != nil {
		t.Fatalf("Dockerfile should exist: %v", err)
	}
	dfContent := string(dfRaw)
	if !strings.Contains(dfContent, "mcr.microsoft.com/devcontainers/python:3-3.12") {
		t.Error("Dockerfile should contain the correct base image")
	}
	if !strings.Contains(dfContent, ".vscode-server") {
		t.Error("Dockerfile should pre-create .vscode-server directory")
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

	if dc.Build.Dockerfile != "Dockerfile" {
		t.Errorf("wrong build.dockerfile: got %q", dc.Build.Dockerfile)
	}

	// Should have mounts for all known AI tools plus gh credentials plus extensions volume
	expectedMounts := len(knownAITools) + 2 // +1 for gh config, +1 for extensions volume
	if len(dc.Mounts) != expectedMounts {
		t.Fatalf("expected %d mounts (AI tools + gh credentials + extensions volume), got %d", expectedMounts, len(dc.Mounts))
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
	script := generateSetupScript("ln -sfn /home/vscode/.vscode-extensions-cache /home/vscode/.vscode-server/extensions")

	if !strings.HasPrefix(script, "#!/bin/bash\n") {
		t.Error("setup script should start with shebang")
	}

	if !strings.Contains(script, ".vscode-extensions-cache") {
		t.Error("script should symlink extensions cache")
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
	script := generateSetupScript("ln -sfn /home/vscode/.vscode-extensions-cache /home/vscode/.vscode-server/extensions")

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

			if dc.Build.Dockerfile != "Dockerfile" {
				t.Errorf("expected build.dockerfile \"Dockerfile\", got %q", dc.Build.Dockerfile)
			}

			// Verify the Dockerfile contains the correct base image
			dfRaw, err := os.ReadFile(filepath.Join(target, ".devcontainer", "Dockerfile"))
			if err != nil {
				t.Fatalf("Dockerfile should exist: %v", err)
			}
			expected := "mcr.microsoft.com/devcontainers/" + img
			if !strings.Contains(string(dfRaw), expected) {
				t.Errorf("Dockerfile should contain image %q", expected)
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

func TestGitignoreContent(t *testing.T) {
	t.Run("common patterns always present", func(t *testing.T) {
		target := mustScaffold(t, TemplateData{
			ProjectName: "test-gitignore",
			Description: "A test project",
		})
		raw, err := os.ReadFile(filepath.Join(target, ".gitignore"))
		if err != nil {
			t.Fatalf(".gitignore should exist: %v", err)
		}
		content := string(raw)
		for _, pattern := range []string{".DS_Store", ".env", ".idea/"} {
			if !strings.Contains(content, pattern) {
				t.Errorf(".gitignore should contain %q", pattern)
			}
		}
	})

	t.Run("Go-specific patterns with Go image", func(t *testing.T) {
		target := mustScaffold(t, TemplateData{
			ProjectName:         "test-gitignore-go",
			Description:         "A test project",
			IncludeDevContainer: true,
			DevContainerImage:   "go:2-1.25-trixie",
		})
		raw, err := os.ReadFile(filepath.Join(target, ".gitignore"))
		if err != nil {
			t.Fatalf(".gitignore should exist: %v", err)
		}
		content := string(raw)
		if !strings.Contains(content, "vendor/") {
			t.Error(".gitignore should contain Go-specific pattern 'vendor/'")
		}
	})

	t.Run("Node-specific patterns with Node image", func(t *testing.T) {
		target := mustScaffold(t, TemplateData{
			ProjectName:         "test-gitignore-node",
			Description:         "A test project",
			IncludeDevContainer: true,
			DevContainerImage:   "typescript-node:20-bookworm",
		})
		raw, err := os.ReadFile(filepath.Join(target, ".gitignore"))
		if err != nil {
			t.Fatalf(".gitignore should exist: %v", err)
		}
		content := string(raw)
		if !strings.Contains(content, "node_modules/") {
			t.Error(".gitignore should contain Node-specific pattern 'node_modules/'")
		}
	})
}

func TestEditorconfigContent(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName: "test-editorconfig",
		Description: "A test project",
	})
	raw, err := os.ReadFile(filepath.Join(target, ".editorconfig"))
	if err != nil {
		t.Fatalf(".editorconfig should exist: %v", err)
	}
	content := string(raw)
	for _, expected := range []string{"root = true", "charset = utf-8", "end_of_line = lf", "insert_final_newline = true"} {
		if !strings.Contains(content, expected) {
			t.Errorf(".editorconfig should contain %q", expected)
		}
	}
}

func TestLicenseMIT(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName: "test-mit",
		Description: "A test project",
		License:     "MIT",
		Year:        2025,
	})
	raw, err := os.ReadFile(filepath.Join(target, "LICENSE"))
	if err != nil {
		t.Fatalf("LICENSE should exist: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, "MIT License") {
		t.Error("LICENSE should contain 'MIT License'")
	}
	if !strings.Contains(content, "2025") {
		t.Error("LICENSE should contain the year")
	}
	if !strings.Contains(content, "test-mit") {
		t.Error("LICENSE should contain the project name")
	}
}

func TestLicenseApache(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName: "test-apache",
		Description: "A test project",
		License:     "Apache-2.0",
		Year:        2025,
	})
	raw, err := os.ReadFile(filepath.Join(target, "LICENSE"))
	if err != nil {
		t.Fatalf("LICENSE should exist: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, "Apache License") {
		t.Error("LICENSE should contain 'Apache License'")
	}
	if !strings.Contains(content, "2025") {
		t.Error("LICENSE should contain the year")
	}
}

func TestLicenseNone(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName: "test-no-license",
		Description: "A test project",
		License:     "none",
	})
	if _, err := os.Stat(filepath.Join(target, "LICENSE")); !os.IsNotExist(err) {
		t.Error("LICENSE should not exist when license is 'none'")
	}
}

func TestLicenseDefaultEmpty(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName: "test-no-license-default",
		Description: "A test project",
	})
	if _, err := os.Stat(filepath.Join(target, "LICENSE")); !os.IsNotExist(err) {
		t.Error("LICENSE should not exist when license is empty")
	}
}

func TestDevcontainerHasGitHubCLIFeature(t *testing.T) {
	target := mustScaffold(t, TemplateData{
		ProjectName:         "test-dc-ghcli",
		Description:         "A test project",
		IncludeDevContainer: true,
		DevContainerImage:   "go:2-1.25-trixie",
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

	if _, ok := dc.Features["ghcr.io/devcontainers/features/github-cli:1"]; !ok {
		t.Error("devcontainer should include github-cli feature")
	}
}

func TestDevContainerVSCodeExtensions(t *testing.T) {
	t.Run("extensions included when selected", func(t *testing.T) {
		target := mustScaffold(t, TemplateData{
			ProjectName:         "test-dc-ext",
			Description:         "A test project",
			IncludeDevContainer: true,
			DevContainerImage:   "go:2-1.25-trixie",
			VSCodeExtensions:    []string{"anthropics.claude-code", "openai.chatgpt"},
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

		if dc.Customizations == nil {
			t.Fatal("customizations should be present when extensions are selected")
		}
		exts := dc.Customizations.VSCode.Extensions
		if len(exts) != 2 {
			t.Fatalf("expected 2 extensions, got %d", len(exts))
		}
		if exts[0] != "anthropics.claude-code" {
			t.Errorf("expected first extension to be anthropics.claude-code, got %q", exts[0])
		}
		if exts[1] != "openai.chatgpt" {
			t.Errorf("expected second extension to be openai.chatgpt, got %q", exts[1])
		}
	})

	t.Run("no customizations when no extensions selected", func(t *testing.T) {
		target := mustScaffold(t, TemplateData{
			ProjectName:         "test-dc-no-ext",
			Description:         "A test project",
			IncludeDevContainer: true,
			DevContainerImage:   "go:2-1.25-trixie",
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

		if dc.Customizations != nil {
			t.Error("customizations should be absent when no extensions are selected")
		}
	})
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
