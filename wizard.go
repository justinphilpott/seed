// Package main - wizard.go
//
// PURPOSE:
// This file implements the interactive TUI wizard using Charm's Huh library.
// It's responsible for:
// - Collecting user input (ProjectName, Description)
// - Validating input within sensible bounds
// - Providing a beautiful, polished user experience
// - Returning structured data ready for template rendering
//
// DESIGN PATTERNS:
// - Form-based input collection (Huh's NewForm pattern)
// - Input validation with clear error messages
// - Separation of concerns: wizard doesn't know about templates or file I/O
//
// USAGE:
// data, err := RunWizard()
// if err != nil { handle error }
// // data is now ready to pass to scaffolder

package main

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
)

// WizardData holds the user's responses from the wizard.
// This is a temporary struct used during wizard execution.
// After collection, it's converted to TemplateData for rendering.
type WizardData struct {
	ProjectName         string
	Description         string
	License             string   // "none", "MIT", or "Apache-2.0"
	InitGit             bool     // Whether to run git init + initial commit
	IncludeDevContainer bool     // Whether to scaffold .devcontainer/
	DevContainerImage   string   // MCR image tag, e.g. "go:2-1.25-trixie"
	AIChatContinuity    bool     // Whether to enable AI chat continuity
	AgentExtensions     []string // Selected extension IDs (e.g. "anthropics.claude-code")
}

// RunWizard launches the interactive TUI wizard and collects user input.
// It displays a form with fields:
// 1. Project Name (text input with validation)
// 2. Description (multi-line text area with validation)
//
// Returns:
// - WizardData: Collected and validated user input
// - error: If user cancels (Ctrl+C) or validation fails unexpectedly
//
// Validation:
// - Project Name: 1-100 chars, non-empty when trimmed
// - Description: 1-500 chars, non-empty when trimmed
func RunWizard(defaultName string) (WizardData, error) {
	var data WizardData
	data.ProjectName = defaultName

	// Create the form with input groups
	// Huh's NewForm accepts one or more Groups
	// Each Group contains related fields that are displayed together
	form := huh.NewForm(
		// Group 1: Core project info
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Value(&data.ProjectName).
				Validate(validateProjectName),

			huh.NewText().
				Title("Description").
				CharLimit(500).
				Value(&data.Description).
				Validate(validateDescription),
		),

		// Group 2: Project setup options
		huh.NewGroup(
			huh.NewConfirm().
				Title("Initialize git repository?").
				Value(&data.InitGit),

			huh.NewConfirm().
				Title("Include a dev container?").
				Value(&data.IncludeDevContainer),
		),

		// Group 3: Dev container details (only shown if opted in)
		huh.NewGroup(
			// Image tags reference MCR defaults at time of release.
			// Check https://mcr.microsoft.com for current versions.
			huh.NewSelect[string]().
				Title("Tech stack").
				Options(
					huh.NewOption("Go", "go:2-1.25-trixie"),
					huh.NewOption("Node/TypeScript", "typescript-node:20-bookworm"),
					huh.NewOption("Python", "python:3-3.12"),
					huh.NewOption("Rust", "rust:1-bookworm"),
					huh.NewOption("Java", "java"),
					huh.NewOption(".NET", "dotnet"),
					huh.NewOption("C++", "cpp"),
					huh.NewOption("Universal (all languages)", "universal"),
				).
				Value(&data.DevContainerImage),

			huh.NewConfirm().
				Title("Enable AI chat continuity?").
				Value(&data.AIChatContinuity),

			huh.NewMultiSelect[string]().
				Title("Agent extensions").
				Options(
					huh.NewOption("Claude Code", "anthropics.claude-code"),
					huh.NewOption("Codex", "openai.chatgpt"),
				).
				Value(&data.AgentExtensions),
		).WithHideFunc(func() bool {
			return !data.IncludeDevContainer
		}),

		// Group 4: License selection (kept last intentionally)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("License").
				Options(
					huh.NewOption("None", "none"),
					huh.NewOption("MIT", "MIT"),
					huh.NewOption("Apache-2.0", "Apache-2.0"),
				).
				Value(&data.License),
		),
	)

	// Run the form and wait for user to complete or cancel
	// form.Run() blocks until user submits (Enter) or cancels (Ctrl+C/Esc)
	if err := form.Run(); err != nil {
		// User cancelled (Ctrl+C) or unexpected error
		return WizardData{}, err
	}

	// Trim whitespace from text inputs
	// This ensures "  myproject  " becomes "myproject"
	data.ProjectName = strings.TrimSpace(data.ProjectName)
	data.Description = strings.TrimSpace(data.Description)

	return data, nil
}

// validateProjectName validates the project name input.
// Called automatically by Huh during form input.
//
// Validation rules (sensible bounds):
// - Required (non-empty after trimming)
// - Maximum 100 characters
// - Minimum 1 character after trimming
//
// Returns:
// - nil if valid
// - error with user-friendly message if invalid
func validateProjectName(s string) error {
	trimmed := strings.TrimSpace(s)

	// Check minimum length (required field)
	if len(trimmed) == 0 {
		return errors.New("project name is required")
	}

	// Check maximum length (sensible bound)
	if len(trimmed) > 100 {
		return errors.New("project name is too long (max 100 characters)")
	}

	return nil
}

// validateDescription validates the description input.
// Called automatically by Huh during form input.
//
// Validation rules (sensible bounds):
// - Required (non-empty after trimming)
// - Maximum 500 characters
// - Minimum 1 character after trimming
//
// Returns:
// - nil if valid
// - error with user-friendly message if invalid
func validateDescription(s string) error {
	trimmed := strings.TrimSpace(s)

	// Check minimum length (required field)
	if len(trimmed) == 0 {
		return errors.New("description is required")
	}

	// Check maximum length (sensible bound)
	// 500 chars is enough for 2-3 sentences
	if len(trimmed) > 500 {
		return errors.New("description is too long (max 500 characters)")
	}

	return nil
}

// ToTemplateData converts WizardData to TemplateData.
// This is a simple mapping function that bridges the wizard layer
// and the scaffolding layer.
//
// Note: Year is NOT set here - it's auto-populated
// by the Scaffolder to ensure it's always current.
func (w WizardData) ToTemplateData() TemplateData {
	return TemplateData{
		ProjectName:         w.ProjectName,
		Description:         w.Description,
		License:             w.License,
		IncludeDevContainer: w.IncludeDevContainer,
		DevContainerImage:   w.DevContainerImage,
		AIChatContinuity:    w.AIChatContinuity,
		VSCodeExtensions:    w.AgentExtensions,
	}
}
