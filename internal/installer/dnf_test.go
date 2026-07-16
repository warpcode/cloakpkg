package installer

import (
	"testing"
)

func TestExpandRepoVariablesDnf(t *testing.T) {
	// Trigger the once execution so it doesn't overwrite our custom values during tests
	expandRepoVariablesDnf("")

	// Save original values and restore them after the test
	origDistro := dnfDistro
	origVersionID := dnfVersionID
	defer func() {
		dnfDistro = origDistro
		dnfVersionID = origVersionID
	}()

	// Set predictable values for testing
	dnfDistro = "testfedora"
	dnfVersionID = "99"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no variables",
			input:    "https://example.com/repo",
			expected: "https://example.com/repo",
		},
		{
			name:     "lowercase distro",
			input:    "https://example.com/${distro}/repo",
			expected: "https://example.com/testfedora/repo",
		},
		{
			name:     "uppercase distro",
			input:    "https://example.com/${DISTRO}/repo",
			expected: "https://example.com/testfedora/repo",
		},
		{
			name:     "lowercase version_id",
			input:    "https://example.com/${version_id}/repo",
			expected: "https://example.com/99/repo",
		},
		{
			name:     "uppercase version_id",
			input:    "https://example.com/${VERSION_ID}/repo",
			expected: "https://example.com/99/repo",
		},
		{
			name:     "mixed variables",
			input:    "https://example.com/${distro}/${VERSION_ID}/repo",
			expected: "https://example.com/testfedora/99/repo",
		},
		{
			name:     "multiple occurrences",
			input:    "${distro}-${distro}-${version_id}-${version_id}",
			expected: "testfedora-testfedora-99-99",
		},
		{
			name:     "unhandled variables",
			input:    "https://example.com/${arch}/${version_id}/repo",
			expected: "https://example.com/${arch}/99/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandRepoVariablesDnf(tt.input)
			if result != tt.expected {
				t.Errorf("expandRepoVariablesDnf(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkExpandRepoVariablesDnf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expandRepoVariablesDnf("test string with ${DISTRO} and ${VERSION_ID}")
	}
}
