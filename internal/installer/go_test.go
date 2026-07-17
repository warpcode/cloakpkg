package installer

import (
	"testing"
)

func TestGetGoBinaryName(t *testing.T) {
	tests := []struct {
		name     string
		pkgName  string
		expected string
	}{
		{
			name:     "simple name without version",
			pkgName:  "goimports",
			expected: "goimports",
		},
		{
			name:     "simple name with version",
			pkgName:  "goimports@latest",
			expected: "goimports",
		},
		{
			name:     "full path without version",
			pkgName:  "golang.org/x/tools/cmd/goimports",
			expected: "goimports",
		},
		{
			name:     "full path with version",
			pkgName:  "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2",
			expected: "golangci-lint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getGoBinaryName(tt.pkgName)
			if result != tt.expected {
				t.Errorf("getGoBinaryName(%q) = %q; want %q", tt.pkgName, result, tt.expected)
			}
		})
	}
}
