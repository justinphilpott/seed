// Package main - skills.go
//
// PURPOSE:
// This file handles installing skill files into target projects.
// Skills are markdown-based agent instructions (e.g., doc health check)
// that are embedded in the binary and copied to the target project.
//
// DESIGN PATTERNS:
// - Embedded filesystem (embed.FS) for zero-dependency distribution
// - Same pattern as scaffold.go: embed at compile time, copy to target
// - Separation of concerns: this file doesn't know about TUI or CLI args
//
// USAGE:
// err := InstallSkills("/path/to/project")

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// skillsFS embeds all skill files at compile time.
//
//go:embed skills/*.md
var skillsFS embed.FS

type skillsInstallReport struct {
	Skipped []string
}

// InstallSkills copies all embedded skill files into targetDir/skills/.
// Creates the skills/ directory if it doesn't exist.
func InstallSkills(targetDir string) error {
	_, err := installSkillsWithReport(targetDir)
	return err
}

// installSkillsWithReport performs skills installation and returns structured
// reporting data so the caller can handle all user-facing output centrally.
func installSkillsWithReport(targetDir string) (skillsInstallReport, error) {
	report := skillsInstallReport{}

	// Verify target directory exists
	info, err := os.Stat(targetDir)
	if err != nil {
		return report, fmt.Errorf("target directory %s does not exist", targetDir)
	}
	if !info.IsDir() {
		return report, fmt.Errorf("%s is not a directory", targetDir)
	}

	// Create skills/ subdirectory
	skillsDir := filepath.Join(targetDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return report, fmt.Errorf("failed to create skills directory: %w", err)
	}

	// Walk embedded skills and copy each one
	entries, err := fs.ReadDir(skillsFS, "skills")
	if err != nil {
		return report, fmt.Errorf("failed to read embedded skills: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		outputPath := filepath.Join(skillsDir, entry.Name())

		// Skip files that already exist to avoid clobbering user modifications
		if _, err := os.Stat(outputPath); err == nil {
			report.Skipped = append(report.Skipped, entry.Name())
			continue
		}

		content, err := skillsFS.ReadFile(filepath.Join("skills", entry.Name()))
		if err != nil {
			return report, fmt.Errorf("failed to read skill %s: %w", entry.Name(), err)
		}

		if err := os.WriteFile(outputPath, content, 0644); err != nil {
			return report, fmt.Errorf("failed to write %s: %w", outputPath, err)
		}
	}

	sort.Strings(report.Skipped)
	return report, nil
}
