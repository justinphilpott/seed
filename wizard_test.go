package main

import (
	"strings"
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string // substring of expected error, empty if valid
	}{
		{"valid simple name", "my-project", ""},
		{"valid single char", "x", ""},
		{"valid at max length", strings.Repeat("a", 100), ""},
		{"empty string", "", "required"},
		{"whitespace only", "   ", "required"},
		{"tabs only", "\t\t", "required"},
		{"exceeds max length", strings.Repeat("a", 101), "too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectName(tt.input)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"valid description", "A project to explore GraphQL", ""},
		{"valid single char", "x", ""},
		{"valid at max length", strings.Repeat("a", 500), ""},
		{"empty string", "", "required"},
		{"whitespace only", "   ", "required"},
		{"exceeds max length", strings.Repeat("a", 501), "too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDescription(tt.input)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}

func TestToTemplateData(t *testing.T) {
	wd := WizardData{
		ProjectName:            "test-project",
		Description:            "A test description",
		License:                "MIT",
		InitGit:                true,
		IncludeDevContainer:    true,
		DevContainerImage:      "go:2-1.25-trixie",
		AIChatContinuity:       true,
		AgentExtensions:        []string{"anthropics.claude-code", "openai.chatgpt"},
	}

	td := wd.ToTemplateData()

	if td.ProjectName != wd.ProjectName {
		t.Errorf("ProjectName: got %q, want %q", td.ProjectName, wd.ProjectName)
	}
	if td.Description != wd.Description {
		t.Errorf("Description: got %q, want %q", td.Description, wd.Description)
	}
	if td.License != wd.License {
		t.Errorf("License: got %q, want %q", td.License, wd.License)
	}
	if td.IncludeDevContainer != wd.IncludeDevContainer {
		t.Errorf("IncludeDevContainer: got %v, want %v", td.IncludeDevContainer, wd.IncludeDevContainer)
	}
	if td.DevContainerImage != wd.DevContainerImage {
		t.Errorf("DevContainerImage: got %q, want %q", td.DevContainerImage, wd.DevContainerImage)
	}
	if td.AIChatContinuity != wd.AIChatContinuity {
		t.Errorf("AIChatContinuity: got %v, want %v", td.AIChatContinuity, wd.AIChatContinuity)
	}
	if len(td.VSCodeExtensions) != len(wd.AgentExtensions) {
		t.Errorf("VSCodeExtensions length: got %d, want %d", len(td.VSCodeExtensions), len(wd.AgentExtensions))
	}
	for i, ext := range td.VSCodeExtensions {
		if ext != wd.AgentExtensions[i] {
			t.Errorf("VSCodeExtensions[%d]: got %q, want %q", i, ext, wd.AgentExtensions[i])
		}
	}
}
