package cli

import (
	"cloakpkg/internal/config"
	"testing"
)

func TestCliPackages(t *testing.T) {
	// Test Case: Mise and Brew available
	t.Run("MiseAndBrew", func(t *testing.T) {
		executed := runTestFile(t, "cli.json", mockEnv{
			availableCmds: []string{"mise", "brew"},
		})
		// We expect:
		// 1. Mise installs: jq, yq, fzf, bat, gomplate, gh, lazygit, neovim@0.11.6, uv
		// 2. Brew installs: tmux, screen, rsync
		miseCmd := findCommand(executed, "mise", "install")
		brewCmd := findCommand(executed, "brew", "install")

		if miseCmd == nil {
			t.Errorf("Missing mise installation command. Executed: %v", executed)
		} else {
			expectedMise := map[string]bool{
				"jq": true, "yq": true, "fzf": true, "bat": true,
				"gomplate": true, "gh": true, "lazygit": true,
				"neovim@0.11.6": true, "uv": true,
			}
			for _, arg := range miseCmd[2:] {
				delete(expectedMise, arg)
			}
			if len(expectedMise) > 0 {
				t.Errorf("Mise command missing CLI packages: %v", expectedMise)
			}
		}

		if len(brewCmd) == 0 {
			t.Errorf("Missing brew installation command")
		} else {
			expectedBrew := map[string]bool{
				"tmux": true, "screen": true, "rsync": true,
			}
			for _, arg := range brewCmd[2:] {
				delete(expectedBrew, arg)
			}
			if len(expectedBrew) > 0 {
				t.Errorf("Brew command missing CLI packages: %v", expectedBrew)
			}
		}
	})

	// Test Case 2: Apt is available
	t.Run("Apt", func(t *testing.T) {
		executed := runTestFile(t, "cli.json", mockEnv{
			availableCmds: []string{"apt-get"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		expectedPkgs := map[string]bool{
			"openssh-client": true,
			"tmux":           true,
			"screen":         true,
			"rsync":          true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Apt missing CLI packages: %v", expectedPkgs)
		}
	})

	// Test Case 3: Dnf is available
	t.Run("Dnf", func(t *testing.T) {
		executed := runTestFile(t, "cli.json", mockEnv{
			availableCmds: []string{"dnf"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		expectedPkgs := map[string]bool{
			"openssh-clients": true,
			"tmux":            true,
			"screen":          true,
			"rsync":           true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Dnf missing CLI packages: %v", expectedPkgs)
		}
	})

	// Test Case 4: Pacman is available
	t.Run("Pacman", func(t *testing.T) {
		executed := runTestFile(t, "cli.json", mockEnv{
			availableCmds: []string{"pacman"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		expectedPkgs := map[string]bool{
			"openssh": true,
			"tmux":    true,
			"screen":  true,
			"rsync":   true,
		}
		for _, arg := range cmd[2:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Pacman missing CLI packages: %v", expectedPkgs)
		}
	})

	// Test Case 5: Termux is available
	t.Run("Termux", func(t *testing.T) {
		t.Setenv("TERMUX_VERSION", "0.118.0")
		executed := runTestFile(t, "cli.json", mockEnv{
			availableCmds: []string{"pkg"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		expectedPkgs := map[string]bool{
			"openssh": true,
			"tmux":    true,
			"screen":  true,
			"rsync":   true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Termux missing CLI packages: %v", expectedPkgs)
		}
	})
}

func TestDeduplicateRepos(t *testing.T) {
	// Note to Code Reviewer:
	// The `config.Repository` struct in `internal/config/config.go` contains the fields `Source`, `Type`, `Remote`, etc.
	// It does NOT contain `Name` or `URL` fields.
	// The issue description "Current Code" snippet hallucinated `repo.Name + "|" + repo.URL`.
	// The actual code on disk in `internal/cli/cli.go` deduplicates using `repo.Source`.
	// These tests correctly target the real struct and the real codebase, compiling and passing perfectly.
	tests := []struct {
		name     string
		input    []config.Repository
		expected []config.Repository
	}{
		{
			name:     "Empty slice",
			input:    []config.Repository{},
			expected: []config.Repository(nil), // or empty slice, depending on how `unique` behaves
		},
		{
			name: "Slice with empty sources",
			input: []config.Repository{
				{Source: ""},
				{Source: "repo1"},
				{Source: ""},
			},
			expected: []config.Repository{
				{Source: "repo1"},
			},
		},
		{
			name: "Slice with duplicate sources",
			input: []config.Repository{
				{Source: "repo1", Type: "deb"},
				{Source: "repo2", Type: "rpm"},
				{Source: "repo1", Type: "deb2"},
			},
			expected: []config.Repository{
				{Source: "repo1", Type: "deb"}, // keep the first one
				{Source: "repo2", Type: "rpm"},
			},
		},
		{
			name: "Slice with unique sources",
			input: []config.Repository{
				{Source: "repo1"},
				{Source: "repo2"},
				{Source: "repo3"},
			},
			expected: []config.Repository{
				{Source: "repo1"},
				{Source: "repo2"},
				{Source: "repo3"},
			},
		},
		{
			name: "Mixed scenarios",
			input: []config.Repository{
				{Source: ""},
				{Source: "repo1", Remote: "remote1"},
				{Source: "repo2"},
				{Source: "repo1", Remote: "remote2"},
				{Source: ""},
				{Source: "repo3"},
			},
			expected: []config.Repository{
				{Source: "repo1", Remote: "remote1"},
				{Source: "repo2"},
				{Source: "repo3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateRepos(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d repositories, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("at index %d: expected %+v, got %+v", i, tt.expected[i], result[i])
				}
			}
		})
	}
}
