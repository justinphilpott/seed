// Package main - main.go
//
// PURPOSE:
// This is the CLI entry point for the seed tool.
// It's responsible for:
// - Parsing command-line arguments
// - Displaying usage/help information
// - Orchestrating the wizard -> scaffolder flow
// - Handling errors and providing user-friendly messages
//
// DESIGN PATTERNS:
// - Thin orchestration layer (delegates to wizard.go and scaffold.go)
// - Fail-fast error handling with clear messages
// - Single responsibility: CLI argument handling and flow control
//
// USAGE:
// seed <directory>
// seed myproject     -> Creates ./myproject/
// seed ~/dev/myapp   -> Creates ~/dev/myapp/

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Minimal styles for output messages
var (
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")) // green
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))            // gray
)

// Version is set at build time via ldflags. Falls back to "dev" for local builds.
var Version = "dev"

type usageError struct {
	msg string
}

func (e usageError) Error() string {
	return e.msg
}

func main() {
	// Run main logic and exit with appropriate code
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, formatErrorOutput(displayVersion(), err))
		os.Exit(1)
	}
}

func displayVersion() string {
	return strings.TrimPrefix(Version, "v")
}

func renderStartBanner(version string) string {
	return fmt.Sprintf("🌱 Seed %s - Simple project scaffolding. Setup wizard:", version)
}

func renderErrorBanner(version, message string) string {
	return fmt.Sprintf("🌱 Seed %s - Error: %s", version, message)
}

func renderScaffoldingLine() string {
	return "scaffolding..."
}

func formatErrorOutput(version string, err error) string {
	var b strings.Builder
	b.WriteString(renderErrorBanner(version, err.Error()))

	var usageErr usageError
	if errors.As(err, &usageErr) {
		b.WriteString("\n\nUsage: seed <directory>")
	}

	return b.String()
}

// run contains the main program logic.
// Separated from main() to enable clean error handling and testing.
//
// Flow:
// 1. Parse CLI arguments -> get target directory
// 2. Run TUI wizard -> collect user input
// 3. Initialize scaffolder -> prepare template engine
// 4. Scaffold project -> render templates and write files
// 5. Print success message
//
// Returns:
// - error: If any step fails
func run() error {
	// Step 1: Parse command-line arguments
	targetDir, err := parseArgs()
	if err != nil {
		return err
	}

	// Step 2: Show startup context
	fmt.Println(renderStartBanner(displayVersion()))
	fmt.Println()

	targetDirExisted, err := targetDirectoryExists(targetDir)
	if err != nil {
		return err
	}

	// Step 3: Check target directory and confirm if non-empty
	allowNonEmpty, err := checkTargetDir(targetDir)
	if err != nil {
		return err
	}

	// Capture existing files before scaffolding so we can report created files by phase.
	beforeFiles, err := snapshotProjectFiles(targetDir)
	if err != nil {
		return fmt.Errorf("failed to inspect existing files: %w", err)
	}

	// Step 4: Run interactive wizard
	wizardData, err := RunWizard(filepath.Base(targetDir))
	if err != nil {
		// User cancelled (Ctrl+C) or validation error
		return fmt.Errorf("wizard cancelled: %w", err)
	}

	// Step 5: Initialize scaffolder with embedded templates
	scaffolder, err := NewScaffolder()
	if err != nil {
		// This should never happen if templates are valid
		return fmt.Errorf("failed to initialize scaffolder: %w", err)
	}

	fmt.Println(renderScaffoldingLine())
	fmt.Println()

	if !targetDirExisted {
		fmt.Printf("Created directory: %s\n", targetDir)
	}

	// Step 6: Convert wizard data to template data and scaffold
	templateData := wizardData.ToTemplateData()
	if err := scaffolder.Scaffold(targetDir, templateData, allowNonEmpty); err != nil {
		return fmt.Errorf("failed to scaffold project: %w", err)
	}

	afterScaffoldFiles, err := snapshotProjectFiles(targetDir)
	if err != nil {
		return fmt.Errorf("failed to inspect scaffolded files: %w", err)
	}
	scaffoldCreatedFiles := createdFileList(beforeFiles, afterScaffoldFiles)
	for _, file := range scaffoldCreatedFiles {
		fmt.Printf("%s created %s\n", successStyle.Render("✓"), file)
	}

	// Step 7: Install agent skills into the project
	_, err = installSkillsWithReport(targetDir)
	if err != nil {
		return fmt.Errorf("failed to install skills: %w", err)
	}

	afterSkillsFiles, err := snapshotProjectFiles(targetDir)
	if err != nil {
		return fmt.Errorf("failed to inspect created files: %w", err)
	}
	skillsCreatedFiles := createdFileList(afterScaffoldFiles, afterSkillsFiles)
	for _, file := range skillsCreatedFiles {
		fmt.Printf("%s created %s\n", successStyle.Render("✓"), file)
	}

	gitActions := []string{}
	// Step 8: Optionally initialize git repository
	if wizardData.InitGit {
		gitActions, err = initGitRepo(targetDir, wizardData.ProjectName)
		if err != nil {
			return fmt.Errorf("failed to initialize git: %w", err)
		}
		for _, action := range gitActions {
			fmt.Printf("%s %s\n", successStyle.Render("✓"), action)
		}
	}

	fmt.Println("Done.")

	return nil
}

func targetDirectoryExists(targetDir string) (bool, error) {
	info, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cannot access %s: %w", targetDir, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s exists but is not a directory", targetDir)
	}
	return true, nil
}

// snapshotProjectFiles returns all file paths under root as slash-normalized
// paths relative to root. Missing roots return an empty set.
func snapshotProjectFiles(root string) (map[string]struct{}, error) {
	files := make(map[string]struct{})

	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return files, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// createdFileList returns sorted file paths present in after but not before.
func createdFileList(before, after map[string]struct{}) []string {
	var created []string
	for file := range after {
		if _, exists := before[file]; !exists {
			created = append(created, file)
		}
	}
	sort.Strings(created)
	return created
}

// checkTargetDir validates the target directory before launching the wizard.
// Returns (allowNonEmpty, error). If the directory is non-empty, prompts the
// user for confirmation via TUI. Returns true if user confirmed overwrite.
func checkTargetDir(targetDir string) (bool, error) {
	info, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		// Target doesn't exist yet -> validate that the parent is reachable and writable
		parentDir := filepath.Dir(targetDir)
		parentInfo, err := os.Stat(parentDir)
		if os.IsNotExist(err) {
			return false, fmt.Errorf("parent directory %s does not exist", parentDir)
		}
		if err != nil {
			return false, fmt.Errorf("cannot access parent directory %s: %w", parentDir, err)
		}
		if !parentInfo.IsDir() {
			return false, fmt.Errorf("parent path %s is not a directory", parentDir)
		}
		// Probe writability by attempting to create and remove a temp file
		f, err := os.CreateTemp(parentDir, ".seed-check-*")
		if err != nil {
			return false, fmt.Errorf("cannot write to parent directory %s: %w", parentDir, err)
		}
		f.Close()
		os.Remove(f.Name())
		return false, nil // will be created later
	}
	if err != nil {
		return false, fmt.Errorf("cannot access %s: %w", targetDir, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s exists but is not a directory", targetDir)
	}
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return false, fmt.Errorf("cannot read %s: %w", targetDir, err)
	}
	if len(entries) == 0 {
		return false, nil // empty dir, all good
	}

	// Non-empty -> ask user to confirm
	var confirm bool
	err = huh.NewConfirm().
		Title(fmt.Sprintf("Directory %s contains %d items. Continue anyway?", targetDir, len(entries))).
		Description("Existing files will NOT be overwritten, but new files will be added").
		Value(&confirm).
		Run()
	if err != nil {
		return false, fmt.Errorf("cancelled: %w", err)
	}
	if !confirm {
		return false, fmt.Errorf("aborted -> directory is not empty")
	}
	return true, nil
}

// initGitRepo runs git init, git add, and an initial commit in the target directory.
func initGitRepo(targetDir, projectName string) ([]string, error) {
	commands := []struct {
		args  []string
		label string
	}{
		{args: []string{"git", "init"}, label: "git init"},
		{args: []string{"git", "add", "."}, label: "git add ."},
		{args: []string{"git", "commit", "-m", fmt.Sprintf("Initial scaffold for %s (via seed)", projectName)}, label: "git commit -m \"Initial scaffold for <project> (via seed)\""},
	}

	executed := make([]string, 0, len(commands))
	for _, c := range commands {
		cmd := exec.Command(c.args[0], c.args[1:]...)
		cmd.Dir = targetDir
		cmd.Stdout = nil // suppress output
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			return executed, fmt.Errorf("%s failed: %w", c.label, err)
		}
		executed = append(executed, c.label)
	}
	return executed, nil
}

// parseArgs parses command-line arguments and returns the target directory.
//
// Expected usage:
// - seed <directory>
//
// Returns:
// - string: Target directory path
// - error: If arguments are invalid
//
// Handles:
// - No arguments -> show usage
// - Too many arguments -> usageError
// - --help, -h, help -> show usage
// - --version, -v -> show version
// - --verbose -> accepted for backward compatibility; ignored
func parseArgs() (string, error) {
	args := os.Args[1:] // Skip program name

	// Handle no arguments
	if len(args) == 0 {
		showUsage()
		os.Exit(0)
	}

	// Handle help flags
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		showUsage()
		os.Exit(0)
	}

	// Handle version flags
	if args[0] == "--version" || args[0] == "-v" {
		fmt.Printf("seed version %s\n", Version)
		os.Exit(0)
	}

	if args[0] == "--verbose" {
		args = args[1:]
	}

	if len(args) == 0 {
		return "", usageError{msg: "missing directory argument"}
	}

	// Handle too many arguments
	if len(args) > 1 {
		return "", usageError{msg: "too many arguments"}
	}

	// Return the target directory
	return args[0], nil
}

// showUsage prints usage information to stdout.
// Called when user runs: seed, seed --help, seed -h, or seed help
func showUsage() {
	fmt.Printf(`🌱 seed v%s — rapid agentic project scaffolder

USAGE:
  seed <directory>

WHAT IT DOES:
  Runs an interactive wizard that asks about your project, then generates
  a set of structured markdown docs and agent skill files designed for AI
  agents to work with from day one.

  The wizard collects: project name, description, language/framework,
  and optional devcontainer setup.

GENERATED FILES:
  README.md                        Project overview
  AGENTS.md                        Agent context and constraints
  DECISIONS.md                     Key architectural decisions
  TODO.md                          Active work and next steps
  LEARNINGS.md                     Validated discoveries
  .gitignore                       Git ignore rules (language-aware)
  .editorconfig                    Editor formatting defaults
  LICENSE                          Open-source license (optional)
  .devcontainer/devcontainer.json  Dev container config (optional)
  .devcontainer/setup.sh           AI chat continuity (optional)
  skills/                          Reusable agent skill files

EXAMPLES:
  seed myproject                Create ./myproject/
  seed ~/dev/myapp              Create ~/dev/myapp/
  seed .                        Use current directory (if empty)

FLAGS:
  -h, --help      Show this help message
  -v, --version   Show version number

LEARN MORE:
  https://github.com/justinphilpott/seed
`, Version)
}
