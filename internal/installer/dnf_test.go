package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"strings"
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

func TestDnfAddRepositories(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "dnf"
	}

	dnf := &Dnf{}

	repos := []config.Repository{
		{
			Source:  "[testrepo]\nname=Test Repo\nbaseurl=http://example.com/repo\nenabled=1\ngpgcheck=0",
			Keyring: "test.repo",
		},
		{
			Source: "copr:user/project",
			Type:   "copr",
		},
		{
			Source: "http://example.com/normal-repo",
		},
	}

	err := dnf.AddRepositories(false, false, repos)
	if err != nil {
		t.Fatalf("AddRepositories failed: %v", err)
	}

	if len(executedCmds) != 3 {
		t.Fatalf("Expected 3 commands executed, got %d:\n%v", len(executedCmds), executedCmds)
	}

	// First command: cp <tmp_path> /etc/yum.repos.d/test.repo (or sudo cp ...)
	cmd1 := executedCmds[0]
	// If run as root or windows, no sudo prefix.
	// But in tests, usually run as non-root on Linux/macOS, so sudo might be added.
	// Let's strip "sudo" if it exists for easier assertion.
	if cmd1[0] == "sudo" {
		cmd1 = cmd1[1:]
	}
	if len(cmd1) < 3 || cmd1[0] != "cp" || !strings.HasSuffix(cmd1[2], "/etc/yum.repos.d/test.repo") {
		t.Errorf("Unexpected first command (expected cp to /etc/yum.repos.d/test.repo): %v", cmd1)
	}

	// Second command: dnf copr enable -y user/project
	cmd2 := executedCmds[1]
	if cmd2[0] == "sudo" {
		cmd2 = cmd2[1:]
	}
	if len(cmd2) < 5 || cmd2[0] != "dnf" || cmd2[1] != "copr" || cmd2[2] != "enable" || cmd2[3] != "-y" || cmd2[4] != "user/project" {
		t.Errorf("Unexpected second command (expected dnf copr enable -y user/project): %v", cmd2)
	}

	// Third command: dnf config-manager --add-repo http://example.com/normal-repo
	cmd3 := executedCmds[2]
	if cmd3[0] == "sudo" {
		cmd3 = cmd3[1:]
	}
	if len(cmd3) < 4 || cmd3[0] != "dnf" || cmd3[1] != "config-manager" || cmd3[2] != "--add-repo" || cmd3[3] != "http://example.com/normal-repo" {
		t.Errorf("Unexpected third command (expected dnf config-manager --add-repo http://example.com/normal-repo): %v", cmd3)
	}
}
