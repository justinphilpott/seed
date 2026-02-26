package main

import (
	"errors"
	"os"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantDir      string
		wantErr      bool
		wantUsageErr bool
	}{
		{
			name:         "directory only",
			args:         []string{"seed", "myproject"},
			wantDir:      "myproject",
			wantErr:      false,
			wantUsageErr: false,
		},
		{
			name:         "verbose flag accepted and ignored",
			args:         []string{"seed", "--verbose", "myproject"},
			wantDir:      "myproject",
			wantErr:      false,
			wantUsageErr: false,
		},
		{
			name:         "missing directory after verbose",
			args:         []string{"seed", "--verbose"},
			wantDir:      "",
			wantErr:      true,
			wantUsageErr: true,
		},
		{
			name:         "too many args",
			args:         []string{"seed", "one", "two"},
			wantDir:      "",
			wantErr:      true,
			wantUsageErr: true,
		},
	}

	originalArgs := os.Args
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			gotDir, err := parseArgs()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				var usageErr usageError
				gotUsageErr := errors.As(err, &usageErr)
				if gotUsageErr != tt.wantUsageErr {
					t.Fatalf("usage error mismatch: got %v, want %v", gotUsageErr, tt.wantUsageErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotDir != tt.wantDir {
				t.Fatalf("directory mismatch: got %q, want %q", gotDir, tt.wantDir)
			}
		})
	}
}

func TestRenderBanners(t *testing.T) {
	if got, want := renderStartBanner("0.1.0"), "🌱 Seed 0.1.0 - Simple project scaffolding. Setup wizard:"; got != want {
		t.Fatalf("start banner mismatch: got %q, want %q", got, want)
	}

	if got, want := renderErrorBanner("0.1.0", "boom"), "🌱 Seed 0.1.0 - Error: boom"; got != want {
		t.Fatalf("error banner mismatch: got %q, want %q", got, want)
	}

	if got, want := renderScaffoldingLine(), "scaffolding..."; got != want {
		t.Fatalf("scaffolding line mismatch: got %q, want %q", got, want)
	}
}

func TestFormatErrorOutput(t *testing.T) {
	usage := formatErrorOutput("0.1.0", usageError{msg: "missing directory argument"})
	if want := "🌱 Seed 0.1.0 - Error: missing directory argument\n\nUsage: seed <directory>"; usage != want {
		t.Fatalf("usage error output mismatch:\n got: %q\nwant: %q", usage, want)
	}

	nonUsage := formatErrorOutput("0.1.0", errors.New("failed to scaffold project"))
	if want := "🌱 Seed 0.1.0 - Error: failed to scaffold project"; nonUsage != want {
		t.Fatalf("non-usage error output mismatch:\n got: %q\nwant: %q", nonUsage, want)
	}
}
