// Package main - main.go
//
// PURPOSE:
// This is the CLI entry point for the seed tool.
// It's responsible for:
// - Parsing command-line arguments
// - Displaying usage/help information
// - Orchestrating the wizard → scaffolder flow
// - Handling errors and providing user-friendly messages
//
// DESIGN PATTERNS:
// - Thin orchestration layer (delegates to wizard.go and scaffold.go)
// - Fail-fast error handling with clear messages
// - Single responsibility: CLI argument handling and flow control
//
// USAGE:
// seed <directory>
// seed myproject     → Creates ./myproject/
// seed ~/dev/myapp   → Creates ~/dev/myapp/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Minimal styles for output messages
var (
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")) // green
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))            // gray
	errorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")) // red
)

// Version is set at build time via ldflags. Falls back to "dev" for local builds.
var Version = "dev"

func main() {
	// Run main logic and exit with appropriate code
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", errorStyle.Render("Error:"), err)
		os.Exit(1)
	}
}

// run contains the main program logic.
// Separated from main() to enable clean error handling and testing.
//
// Flow:
// 1. Parse CLI arguments → get target directory
// 2. Run TUI wizard → collect user input
// 3. Initialize scaffolder → prepare template engine
// 4. Scaffold project → render templates and write files
// 5. Print success message
//
// Returns:
// - error: If any step fails
func run() error {
	// Handle subcommands before normal arg parsing
	if len(os.Args) >= 2 && os.Args[1] == "skills" {
		return runSkills()
	}

	// Step 1: Parse command-line arguments
	targetDir, err := parseArgs()
	if err != nil {
		return err
	}

	// Step 2: Check target directory and confirm if non-empty
	allowNonEmpty, err := checkTargetDir(targetDir)
	if err != nil {
		return err
	}

	// Step 3: Run interactive wizard
	fmt.Println("🌱 Seed - Project Scaffolder")
	fmt.Println()

	wizardData, err := RunWizard(filepath.Base(targetDir))
	if err != nil {
		// User cancelled (Ctrl+C) or validation error
		return fmt.Errorf("wizard cancelled: %w", err)
	}

	// Step 4: Initialize scaffolder with embedded templates
	scaffolder, err := NewScaffolder()
	if err != nil {
		// This should never happen if templates are valid
		return fmt.Errorf("failed to initialize scaffolder: %w", err)
	}

	// Step 5: Convert wizard data to template data and scaffold
	templateData := wizardData.ToTemplateData()
	if err := scaffolder.Scaffold(targetDir, templateData, allowNonEmpty); err != nil {
		return fmt.Errorf("failed to scaffold project: %w", err)
	}

	// Step 6: Optionally initialize git repository
	if wizardData.InitGit {
		if err := initGitRepo(targetDir, wizardData.ProjectName); err != nil {
			return fmt.Errorf("failed to initialize git: %w", err)
		}
	}

	// Step 7: Success! Print confirmation
	fmt.Println()
	fmt.Printf("%s Project %s created in: %s\n",
		successStyle.Render("✓"),
		successStyle.Render(wizardData.ProjectName),
		targetDir)
	fmt.Println()
	fmt.Println(dimStyle.Render("Next steps:"))
	fmt.Printf(dimStyle.Render("  cd %s")+"\n", targetDir)
	if templateData.IncludeDevContainer {
		fmt.Println(dimStyle.Render("  # Open in VS Code and 'Reopen in Container'"))
	}
	fmt.Println(dimStyle.Render("  # Start building!"))

	return nil
}

// checkTargetDir validates the target directory before launching the wizard.
// Returns (allowNonEmpty, error). If the directory is non-empty, prompts the
// user for confirmation via TUI. Returns true if user confirmed overwrite.
func checkTargetDir(targetDir string) (bool, error) {
	info, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		// Target doesn't exist yet — validate that the parent is reachable and writable
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

	// Non-empty — ask user to confirm
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
		return false, fmt.Errorf("aborted — directory is not empty")
	}
	return true, nil
}

// initGitRepo runs git init, git add, and an initial commit in the target directory.
func initGitRepo(targetDir, projectName string) error {
	commands := []struct {
		args []string
	}{
		{[]string{"git", "init"}},
		{[]string{"git", "add", "."}},
		{[]string{"git", "commit", "-m", fmt.Sprintf("Initial scaffold for %s (via seed)", projectName)}},
	}

	for _, c := range commands {
		cmd := exec.Command(c.args[0], c.args[1:]...)
		cmd.Dir = targetDir
		cmd.Stdout = nil // suppress output
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s failed: %w", c.args[0], err)
		}
	}
	return nil
}

// runSkills handles the `seed skills <directory>` subcommand.
// Installs embedded skill files into the target project's skills/ directory.
func runSkills() error {
	args := os.Args[2:] // Skip "seed" and "skills"

	if len(args) == 0 {
		return fmt.Errorf("missing target directory\n\nUsage: seed skills <directory>")
	}
	if args[0] == "--help" || args[0] == "-h" {
		fmt.Println("Install agent skill files into an existing project.")
		fmt.Println()
		fmt.Println("Usage: seed skills <directory>")
		fmt.Println()
		fmt.Println("Copies skill files (e.g., doc-health-check.md) into <directory>/skills/.")
		fmt.Println("Skills are agent instructions — markdown files that define reusable procedures.")
		os.Exit(0)
	}
	if len(args) > 1 {
		return fmt.Errorf("too many arguments\n\nUsage: seed skills <directory>")
	}

	targetDir := args[0]
	if err := InstallSkills(targetDir); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("%s Skills installed to: %s\n",
		successStyle.Render("✓"),
		filepath.Join(targetDir, "skills"))
	fmt.Println()
	fmt.Println(dimStyle.Render("Installed skills:"))
	fmt.Println(dimStyle.Render("  doc-health-check.md  — Audit project docs for informational coverage"))
	fmt.Println()
	fmt.Println(dimStyle.Render("Point your agent at skills/ to use them."))

	return nil
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
// - No arguments → show usage
// - Too many arguments → show usage
// - --help, -h, help → show usage
// - --version, -v → show version
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

	// Handle too many arguments
	if len(args) > 1 {
		return "", fmt.Errorf("too many arguments\n\nUsage: seed <directory>")
	}

	// Return the target directory
	return args[0], nil
}

// showUsage prints usage information to stdout.
// Called when user runs: seed, seed --help, seed -h, or seed help
func showUsage() {
	fmt.Printf(`seed v%s - Project Scaffolder

USAGE:
  seed <directory>              Scaffold a new project
  seed skills <directory>       Install agent skills into a project

DESCRIPTION:
  Creates a new project with minimal, agent-friendly documentation.
  Runs an interactive wizard to collect project details.

EXAMPLES:
  seed myproject                Create ./myproject/
  seed ~/dev/myapp              Create ~/dev/myapp/
  seed .                        Use current directory (if empty)
  seed skills ./myproject       Install skills into existing project

FLAGS:
  -h, --help      Show this help message
  -v, --version   Show version number

GENERATED FILES:
  README.md                        Project overview
  AGENTS.md                        Agent context and constraints
  DECISIONS.md                     Key architectural decisions
  TODO.md                          Active work and next steps
  LEARNINGS.md                     Validated discoveries
  .devcontainer/devcontainer.json  Dev container config (optional)
  .devcontainer/setup.sh           AI chat continuity (optional)

INSTALL:
  go install github.com/justinphilpott/seed@latest

LEARN MORE:
  https://github.com/justinphilpott/seed
`, Version)
}
