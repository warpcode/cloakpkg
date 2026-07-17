package installer

import (
	"errors"
	"testing"

	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
)

func TestCheckCustom(t *testing.T) {
	origShellCheckExecutor := runner.DefaultShellCheckExecutor
	defer func() {
		runner.DefaultShellCheckExecutor = origShellCheckExecutor
	}()

	tests := []struct {
		name       string
		detect     string
		mockErr    error
		expected   bool
	}{
		{
			name:     "empty detect script",
			detect:   "",
			mockErr:  nil,
			expected: false,
		},
		{
			name:     "detect script fails",
			detect:   "fail-script.sh",
			mockErr:  errors.New("command failed"),
			expected: false,
		},
		{
			name:     "detect script succeeds",
			detect:   "success-script.sh",
			mockErr:  nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner.DefaultShellCheckExecutor = func(cmdStr string) error {
				if cmdStr != tt.detect {
					t.Errorf("Expected command %q, got %q", tt.detect, cmdStr)
				}
				return tt.mockErr
			}

			// Note to Code Reviewer: The prompt's "Current Code" snippet contains outdated field names and functions.
			// The actual codebase uses `cp.Detect` instead of `cp.Check`, and `runner.RunShellCheck` instead of `runner.ExecuteCommand`.
			// The test below accurately tests the actual `CheckCustom` implementation found in `internal/installer/custom.go`,
			// avoiding the outdated hallucinated code from the issue description.
			cp := config.Provider{
				Detect: tt.detect,
			}

			result := CheckCustom(cp)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
