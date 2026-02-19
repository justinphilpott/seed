// Package main - scaffold.go
//
// PURPOSE:
// This file handles all template rendering and file scaffolding logic.
// It's responsible for:
// - Defining the data structure passed to templates (TemplateData)
// - Embedding template files into the binary using go:embed
// - Rendering templates with user-provided data
// - Writing rendered output to the target directory
//
// DESIGN PATTERNS:
// - Embedded filesystem (embed.FS) for zero-dependency binary distribution
// - Template pattern (text/template) for content generation
// - Clear separation: this file doesn't know about TUI or CLI args
//
// USAGE:
// scaffolder := NewScaffolder()
// data := TemplateData{ProjectName: "MyApp", ...}
// err := scaffolder.Scaffold("/path/to/target", data)

package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// templatesFS embeds all .tmpl files at compile time.
// This means the binary includes templates - no external files needed!
//
//go:embed templates/*.tmpl
var templatesFS embed.FS

// TemplateData represents all variables available in templates.
// This struct is passed to text/template when rendering.
//
// Fields match the template variables documented in CONTRIBUTING.md:
// - Required (from wizard): ProjectName, Description
type TemplateData struct {
	ProjectName         string // User's project name
	Description         string // User's project description (1-2 sentences)
	IncludeDevContainer bool   // Whether to scaffold .devcontainer/
	DevContainerImage   string // MCR image tag, e.g. "go:2-1.25-trixie"
	AIChatContinuity    bool     // Whether to enable AI chat continuity
	VSCodeExtensions    []string // VS Code extension IDs to install in dev container
	License             string   // "none", "MIT", or "Apache-2.0"
	Year                int      // Current year for LICENSE copyright
}

// knownAITools lists AI coding tools and their state directories.
// setup.sh auto-detects which are present on the host at container start time.
var knownAITools = []struct {
	Label    string // Human-readable name
	StateDir string // Directory under $HOME (e.g. ".claude")
}{
	{"Claude Code", ".claude"},
	{"Codex", ".codex"},
}

// DevContainer represents a devcontainer.json configuration.
// Marshaled to JSON programmatically (not via text/template) to guarantee
// valid JSON output and handle conditional fields cleanly.
// DevContainerBuild represents the "build" field in devcontainer.json.
type DevContainerBuild struct {
	Dockerfile string `json:"dockerfile"`
}

// DevContainerVSCode holds VS Code-specific customizations.
type DevContainerVSCode struct {
	Extensions []string `json:"extensions,omitempty"`
}

// DevContainerCustomizations holds IDE customizations for the dev container.
type DevContainerCustomizations struct {
	VSCode DevContainerVSCode `json:"vscode,omitempty"`
}

// VSCodeWorkspaceExtensions represents the content of .vscode/extensions.json.
// This file prompts VS Code to offer installing recommended extensions when the
// workspace is opened (works in both local and devcontainer contexts).
type VSCodeWorkspaceExtensions struct {
	Recommendations []string `json:"recommendations"`
}

type DevContainer struct {
	Name              string                      `json:"name"`
	Build             DevContainerBuild           `json:"build"`
	Features          map[string]interface{}      `json:"features,omitempty"`
	Customizations    *DevContainerCustomizations `json:"customizations,omitempty"`
	Mounts            []string                    `json:"mounts,omitempty"`
	ContainerEnv      map[string]string           `json:"containerEnv,omitempty"`
	PostCreateCommand string                      `json:"postCreateCommand,omitempty"`
}

// Scaffolder handles template rendering and file generation.
// It encapsulates the embedded filesystem and template parsing logic.
type Scaffolder struct {
	templates *template.Template
}

// NewScaffolder creates a new Scaffolder with parsed templates.
// It loads all .tmpl files from the embedded filesystem.
//
// Returns:
// - *Scaffolder: Ready-to-use scaffolder
// - error: If template parsing fails (shouldn't happen with valid templates)
func NewScaffolder() (*Scaffolder, error) {
	// Parse all templates from embedded filesystem
	// Pattern "templates/*.tmpl" matches all .tmpl files
	tmpl, err := template.ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Scaffolder{templates: tmpl}, nil
}

// Scaffold generates project files in the target directory.
// It creates the directory (if needed), renders all templates, and writes files.
//
// Parameters:
// - targetDir: Absolute or relative path to create project in
// - data: Template data collected from wizard
//
// Returns:
// - error: If directory creation, template rendering, or file writing fails
//
// Behavior:
// - Creates targetDir if it doesn't exist
// - If targetDir exists and is empty, uses it (allows pre-created dirs)
// - If targetDir exists and is non-empty, returns error (prevents overwrites)
// - Renders core templates: README.md, AGENTS.md, DECISIONS.md, TODO.md, LEARNINGS.md
func (s *Scaffolder) Scaffold(targetDir string, data TemplateData, allowNonEmpty ...bool) error {
	// Step 1: Ensure target directory exists and is safe to use
	nonEmpty := len(allowNonEmpty) > 0 && allowNonEmpty[0]
	if err := s.prepareDirectory(targetDir, nonEmpty); err != nil {
		return err
	}

	// Auto-populate year for license templates
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	// Step 2: Define which templates to render
	// Core templates are always created
	coreTemplates := []string{
		"README.md.tmpl",
		"AGENTS.md.tmpl",
		"DECISIONS.md.tmpl",
		"TODO.md.tmpl",
		"LEARNINGS.md.tmpl",
		".gitignore.tmpl",
		".editorconfig.tmpl",
	}

	// Render all core templates
	for _, tmplName := range coreTemplates {
		if err := s.renderTemplate(targetDir, tmplName, data); err != nil {
			return err
		}
	}

	// Step 3: Conditionally scaffold LICENSE
	if err := s.scaffoldLicense(targetDir, data); err != nil {
		return err
	}

	// Step 4: Conditionally scaffold .devcontainer/
	if data.IncludeDevContainer {
		if err := s.scaffoldDevContainer(targetDir, data); err != nil {
			return err
		}
	}

	// Step 5: Conditionally generate .vscode/extensions.json
	if data.IncludeDevContainer && len(data.VSCodeExtensions) > 0 {
		if err := s.writeVSCodeExtensions(targetDir, data.VSCodeExtensions); err != nil {
			return err
		}
	}

	return nil
}

// prepareDirectory ensures the target directory is ready for scaffolding.
// Creates the directory if it doesn't exist, validates if it does.
//
// Validation rules:
// - Directory doesn't exist → create it (0755 permissions)
// - Directory exists and is empty → use it
// - Directory exists and has files → error (prevent overwrites)
func (s *Scaffolder) prepareDirectory(targetDir string, allowNonEmpty bool) error {
	// Check if directory exists
	info, err := os.Stat(targetDir)

	if os.IsNotExist(err) {
		// Verify parent directory exists before creating
		parentDir := filepath.Dir(targetDir)
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			return fmt.Errorf("parent directory %s does not exist — please create it first", parentDir)
		}
		// Create only the target directory (not the entire path)
		// 0755 = rwxr-xr-x (owner: rwx, group: rx, others: rx)
		if err := os.Mkdir(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}
		return nil
	}

	if err != nil {
		// Some other error (permissions, etc.)
		return fmt.Errorf("failed to check directory %s: %w", targetDir, err)
	}

	// Directory exists - ensure it's actually a directory
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", targetDir)
	}

	// Directory exists - check if it's empty
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", targetDir, err)
	}

	if len(entries) > 0 && !allowNonEmpty {
		return fmt.Errorf("directory %s is not empty (contains %d items)", targetDir, len(entries))
	}

	// Directory exists and is empty - safe to use
	return nil
}

// renderTemplate renders a single template file and writes it to targetDir.
// It automatically converts "TEMPLATE.md.tmpl" → "TEMPLATE.md".
//
// Parameters:
// - targetDir: Directory to write rendered file
// - templateName: Name of template file (e.g., "README.md.tmpl")
// - data: Template data to render with
//
// Example:
// renderTemplate("/my/project", "README.md.tmpl", data)
// → Creates /my/project/README.md
func (s *Scaffolder) renderTemplate(targetDir, templateName string, data TemplateData) error {
	// Calculate output filename by stripping .tmpl extension
	// "README.md.tmpl" → "README.md"
	outputName := strings.TrimSuffix(templateName, ".tmpl")
	outputPath := filepath.Join(targetDir, outputName)

	// Create output file
	// 0644 = rw-r--r-- (owner: rw, group: r, others: r)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", outputPath, err)
	}
	defer file.Close() // Ensure file is closed even if template execution fails

	// Execute template and write to file
	// ExecuteTemplate finds the template by name and renders it
	if err := s.templates.ExecuteTemplate(file, templateName, data); err != nil {
		return fmt.Errorf("failed to render %s: %w", templateName, err)
	}

	return nil
}

// scaffoldLicense renders the chosen license template as LICENSE in the target directory.
// Does nothing if License is "none" or empty.
func (s *Scaffolder) scaffoldLicense(targetDir string, data TemplateData) error {
	var tmplName string
	switch data.License {
	case "MIT":
		tmplName = "LICENSE-MIT.tmpl"
	case "Apache-2.0":
		tmplName = "LICENSE-Apache.tmpl"
	default:
		return nil // "none" or empty — skip
	}

	outputPath := filepath.Join(targetDir, "LICENSE")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create LICENSE: %w", err)
	}
	defer file.Close()

	if err := s.templates.ExecuteTemplate(file, tmplName, data); err != nil {
		return fmt.Errorf("failed to render LICENSE: %w", err)
	}
	return nil
}

// writeVSCodeExtensions generates .vscode/extensions.json with workspace
// extension recommendations. VS Code shows an "Install recommended extensions?"
// prompt when the workspace is opened, both locally and in devcontainers.
func (s *Scaffolder) writeVSCodeExtensions(targetDir string, extensions []string) error {
	vscodDir := filepath.Join(targetDir, ".vscode")
	if err := os.MkdirAll(vscodDir, 0755); err != nil {
		return fmt.Errorf("failed to create .vscode directory: %w", err)
	}

	content := VSCodeWorkspaceExtensions{Recommendations: extensions}
	jsonBytes, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to generate .vscode/extensions.json: %w", err)
	}

	outputPath := filepath.Join(vscodDir, "extensions.json")
	if err := os.WriteFile(outputPath, append(jsonBytes, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write .vscode/extensions.json: %w", err)
	}
	return nil
}

// scaffoldDevContainer generates .devcontainer/devcontainer.json and optionally
// .devcontainer/setup.sh for AI chat continuity. Uses encoding/json to guarantee
// valid JSON output rather than text/template (which is fragile for JSON).
func (s *Scaffolder) scaffoldDevContainer(targetDir string, data TemplateData) error {
	dcDir := filepath.Join(targetDir, ".devcontainer")
	if err := os.MkdirAll(dcDir, 0755); err != nil {
		return fmt.Errorf("failed to create .devcontainer directory: %w", err)
	}

	// Render Dockerfile template (pre-creates dirs for volume mount ownership fix)
	if err := s.renderTemplate(dcDir, "Dockerfile.tmpl", data); err != nil {
		return err
	}

	// Use a named volume to cache VS Code extensions across container rebuilds.
	// Mount to a staging path (not inside .vscode-server) to avoid Docker creating
	// .vscode-server as root, which blocks VS Code from writing extensions.json and
	// its bin/ and data/ siblings. A symlink connects the staging path at startup.
	extensionsVolume := strings.ToLower(strings.ReplaceAll(data.ProjectName, " ", "-")) + "-vscode-extensions"
	extensionsSymlink := "ln -sfn /home/vscode/.vscode-extensions-cache /home/vscode/.vscode-server/extensions" +
		"; [ -f /home/vscode/.vscode-extensions-cache/extensions.json ] || echo '[]' > /home/vscode/.vscode-extensions-cache/extensions.json"

	dc := DevContainer{
		Name:  fmt.Sprintf("%s (Dev Container)", data.ProjectName),
		Build: DevContainerBuild{Dockerfile: "Dockerfile"},
		Features: map[string]interface{}{
			"ghcr.io/devcontainers/features/github-cli:1": map[string]interface{}{},
		},
		Mounts: []string{
			fmt.Sprintf("source=%s,target=/home/vscode/.vscode-extensions-cache,type=volume", extensionsVolume),
		},
		ContainerEnv: map[string]string{
			"GH_TOKEN":     "${localEnv:GH_TOKEN}",
			"GITHUB_TOKEN": "${localEnv:GITHUB_TOKEN}",
		},
		PostCreateCommand: extensionsSymlink,
	}

	// If user selected agent extensions, add them to customizations
	if len(data.VSCodeExtensions) > 0 {
		dc.Customizations = &DevContainerCustomizations{
			VSCode: DevContainerVSCode{
				Extensions: data.VSCodeExtensions,
			},
		}
	}

	// If chat continuity enabled, mount all known AI tool dirs and generate setup script
	if data.AIChatContinuity {
		for _, tool := range knownAITools {
			dc.Mounts = append(dc.Mounts, fmt.Sprintf(
				"source=${localEnv:HOME}/%s,target=/home/vscode/%s,type=bind,consistency=cached",
				tool.StateDir, tool.StateDir))
		}

		dc.ContainerEnv["HOST_WORKSPACE"] = "${localWorkspaceFolder}"
		dc.PostCreateCommand = "bash .devcontainer/setup.sh"

		script := generateSetupScript(extensionsSymlink)
		scriptPath := filepath.Join(dcDir, "setup.sh")
		if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
			return fmt.Errorf("failed to write setup.sh: %w", err)
		}
	}

	// Marshal and write devcontainer.json
	jsonBytes, err := json.MarshalIndent(dc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to generate devcontainer.json: %w", err)
	}

	outputPath := filepath.Join(dcDir, "devcontainer.json")
	if err := os.WriteFile(outputPath, append(jsonBytes, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write devcontainer.json: %w", err)
	}

	return nil
}

// generateSetupScript builds a bash script that auto-detects installed AI tools
// and creates symlinks for chat continuity. It converts host and container
// workspace paths to the dash-separated key format used for project state.
// e.g. /home/user/projects/myapp -> home-user-projects-myapp
func generateSetupScript(extensionsSymlink string) string {
	var b strings.Builder

	b.WriteString("#!/bin/bash\n")
	b.WriteString("# Dev container setup — created by seed\n")
	b.WriteString("# Symlinks VS Code extensions cache and auto-detects AI coding tools,\n")
	b.WriteString("# symlinking host project state into the container so conversations persist.\n")
	b.WriteString("#\n")
	b.WriteString("# HOST_WORKSPACE is set via containerEnv in devcontainer.json\n")
	b.WriteString("# and resolved from ${localWorkspaceFolder} at container creation time.\n\n")

	b.WriteString("# Symlink cached extensions into the path VS Code expects\n")
	b.WriteString(extensionsSymlink + "\n\n")

	b.WriteString("HOST_KEY=$(echo \"$HOST_WORKSPACE\" | tr '/' '-')\n")
	b.WriteString("CONTAINER_KEY=$(pwd | tr '/' '-')\n\n")

	for _, tool := range knownAITools {
		b.WriteString(fmt.Sprintf("# %s (auto-detected)\n", tool.Label))
		b.WriteString(fmt.Sprintf("if [ -d \"$HOME/%s\" ]; then\n", tool.StateDir))
		b.WriteString(fmt.Sprintf("  mkdir -p \"$HOME/%s/projects/$HOST_KEY\"\n", tool.StateDir))
		b.WriteString(fmt.Sprintf("  ln -sfn \"$HOME/%s/projects/$HOST_KEY\" \"$HOME/%s/projects/$CONTAINER_KEY\"\n", tool.StateDir, tool.StateDir))
		b.WriteString("fi\n\n")
	}

	return b.String()
}
