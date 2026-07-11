package cli

import (
	"cloakpkg/internal/runner"
	"os"
	"path/filepath"
	"testing"
)

func TestMatchTags(t *testing.T) {
	tests := []struct {
		name        string
		bundleTags  []string
		includeTags map[string]bool
		excludeTags map[string]bool
		expected    bool
	}{
		{
			name:       "No filters",
			bundleTags: []string{"core", "cli"},
			expected:   true,
		},
		{
			name:        "Match include filter",
			bundleTags:  []string{"core", "cli"},
			includeTags: map[string]bool{"core": true},
			expected:    true,
		},
		{
			name:        "Mismatch include filter",
			bundleTags:  []string{"core", "cli"},
			includeTags: map[string]bool{"dev": true},
			expected:    false,
		},
		{
			name:        "Exclude match",
			bundleTags:  []string{"core", "cli"},
			excludeTags: map[string]bool{"cli": true},
			expected:    false,
		},
		{
			name:        "Exclude mismatch",
			bundleTags:  []string{"core", "cli"},
			excludeTags: map[string]bool{"dev": true},
			expected:    true,
		},
		{
			name:        "Both filters match include exclude doesn't",
			bundleTags:  []string{"core", "cli"},
			includeTags: map[string]bool{"core": true},
			excludeTags: map[string]bool{"dev": true},
			expected:    true,
		},
		{
			name:        "Both filters match both",
			bundleTags:  []string{"core", "cli"},
			includeTags: map[string]bool{"core": true},
			excludeTags: map[string]bool{"cli": true},
			expected:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := matchTags(tc.bundleTags, tc.includeTags, tc.excludeTags)
			if result != tc.expected {
				t.Errorf("matchTags(%v, %v, %v) = %v; expected %v",
					tc.bundleTags, tc.includeTags, tc.excludeTags, result, tc.expected)
			}
		})
	}
}

func TestBundleCommandLifecycle(t *testing.T) {
	// Mock environment
	origExecutor := runner.DefaultExecutor
	origShellExecutor := runner.DefaultShellExecutor
	origShellCheck := runner.DefaultShellCheckExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultShellExecutor = origShellExecutor
		runner.DefaultShellCheckExecutor = origShellCheck
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.DefaultShellExecutor = func(verbose bool, cmdStr string) error {
		executedCmds = append(executedCmds, []string{"/bin/sh", "-c", cmdStr})
		return nil
	}
	runner.DefaultShellCheckExecutor = func(cmdStr string) error {
		executedCmds = append(executedCmds, []string{"/bin/sh", "-c", cmdStr})
		return os.ErrNotExist
	}
	runner.CommandExists = func(name string) bool {
		return name == "npm" || name == "brew" || name == "apt-get"
	}
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		return os.ErrNotExist
	}
	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		return []byte(""), nil
	}

	tmpDir, err := os.MkdirTemp("", "cloakpkg-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "test_config.json")
	configJSON := `{
		"settings": {
			"provider_priority": ["npm", "brew", "apt", "custom"]
		},
		"bundles": {
			"git": {
				"tags": ["core"],
				"providers": {
					"brew": {
						"repositories": [{"source": "jesseduffield/lazygit"}],
						"packages": [{"name": "git", "extra_params": ["--formula"]}]
					},
					"apt": {
						"packages": [{"name": "git"}]
					}
				}
			},
			"tmux": {
				"tags": ["core"],
				"providers": {
					"brew": {
						"repositories": [{"source": "jesseduffield/lazygit"}],
						"packages": [{"name": "tmux", "extra_params": ["--formula"]}]
					},
					"apt": {
						"packages": [{"name": "tmux"}]
					}
				}
			},
			"htop": {
				"tags": ["extra"],
				"providers": {
					"brew": {
						"packages": [{"name": "htop"}]
					}
				}
			},
			"neovim": {
				"tags": ["core"],
				"providers": {
					"apt": {
						"packages": [{"name": "neovim"}]
					}
				}
			},
			"eslint": {
				"tags": ["core"],
				"providers": {
					"npm": {
						"packages": [{"name": "eslint", "extra_params": ["--save-dev"]}]
					}
				}
			},
			"prettier": {
				"tags": ["core"],
				"providers": {
					"npm": {
						"packages": [{"name": "prettier", "extra_params": ["--save-dev"]}]
					}
				}
			},
			"typescript": {
				"tags": ["core"],
				"providers": {
					"npm": {
						"packages": [{"name": "typescript"}]
					}
				}
			},
			"dotfiles": {
				"tags": ["core"],
				"providers": {
					"custom": {
						"detect": "echo 'checking custom'",
						"install": "echo 'installing custom'"
					}
				}
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"cloakpkg", "install", configPath, "-t", "core"}

	runBundleCommand("install", configPath)

	if findCommand(executedCmds, "npm", "install", "-g", "eslint", "prettier") == nil {
		t.Errorf("Missing collated npm command for eslint and prettier")
	}

	if findCommand(executedCmds, "npm", "install", "-g", "typescript") == nil {
		t.Errorf("Missing npm command for typescript")
	}

	if findCommand(executedCmds, "brew", "tap", "jesseduffield/lazygit") == nil {
		t.Errorf("Missing brew tap command")
	}

	if findCommand(executedCmds, "brew", "install", "--formula", "git", "tmux") == nil {
		t.Errorf("Missing collated brew install command for git and tmux")
	}

	if findCommand(executedCmds, "apt-get", "install", "neovim") == nil {
		t.Errorf("Missing apt install command for neovim")
	}

	if findCommand(executedCmds, "/bin/sh", "-c", "echo 'checking custom'") == nil {
		t.Errorf("Missing custom detect command")
	}

	if findCommand(executedCmds, "/bin/sh", "-c", "echo 'installing custom'") == nil {
		t.Errorf("Missing custom install command")
	}

	for _, cmd := range executedCmds {
		for _, arg := range cmd {
			if arg == "htop" {
				t.Errorf("htop package should have been excluded by tag filter")
			}
		}
	}
}

func TestBundleCommandIdempotency(t *testing.T) {
	// Mock environment
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "brew"
	}
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		for _, arg := range args {
			if arg == "git" {
				return nil
			}
		}
		return os.ErrNotExist
	}

	tmpDir, err := os.MkdirTemp("", "cloakpkg-test-idempotency")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "test_config_idempotency.json")
	configJSON := `{
		"settings": {
			"provider_priority": ["brew"]
		},
		"bundles": {
			"tools": {
				"providers": {
					"brew": {
						"packages": [
							{"name": "git", "extra_params": ["--formula"]},
							{"name": "tmux", "extra_params": ["--formula"]}
						]
					}
				}
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"cloakpkg", "install", configPath}

	runBundleCommand("install", configPath)

	if findCommand(executedCmds, "brew", "install", "--formula", "tmux") == nil {
		t.Errorf("Expected brew install command for tmux")
	}
	if findCommand(executedCmds, "brew", "install", "--formula", "git") != nil {
		t.Errorf("git should not have been installed")
	}
}

func TestBundleCommandHooks(t *testing.T) {
	// Mock environment
	origExecutor := runner.DefaultExecutor
	origShellExecutor := runner.DefaultShellExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultShellExecutor = origShellExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.DefaultShellExecutor = func(verbose bool, cmdStr string) error {
		executedCmds = append(executedCmds, []string{"/bin/sh", "-c", cmdStr})
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "apt-get"
	}
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		return os.ErrNotExist
	}
	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		return []byte(""), nil
	}

	tmpDir, err := os.MkdirTemp("", "cloakpkg-test-hooks")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "test_config_hooks.json")
	configJSON := `{
		"settings": {
			"provider_priority": ["apt", "custom"]
		},
		"bundles": {
			"flatpak": {
				"hooks": {
					"pre_install": "echo 'pre-install bundle'",
					"post_install": "echo 'post-install bundle'"
				},
				"providers": {
					"apt": {
						"packages": [{"name": "flatpak"}],
						"hooks": {
							"pre_install": "echo 'pre-install apt'",
							"post_install": "echo 'post-install apt'"
						}
					}
				}
			},
			"custom-pkg": {
				"hooks": {
					"pre_install": "echo 'pre-install custom bundle'",
					"post_install": "echo 'post-install custom bundle'"
				},
				"providers": {
					"custom": {
						"detect": "false",
						"install": "echo 'install custom'",
						"hooks": {
							"pre_install": "echo 'pre-install custom provider'",
							"post_install": "echo 'post-install custom provider'"
						}
					}
				}
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"cloakpkg", "install", configPath}

	runBundleCommand("install", configPath)

	expectedHooks := []string{
		"echo 'pre-install bundle'",
		"echo 'pre-install apt'",
		"echo 'post-install apt'",
		"echo 'post-install bundle'",
		"echo 'pre-install custom bundle'",
		"echo 'pre-install custom provider'",
		"echo 'install custom'",
		"echo 'post-install custom provider'",
		"echo 'post-install custom bundle'",
	}

	for _, hook := range expectedHooks {
		if findCommand(executedCmds, "/bin/sh", "-c", hook) == nil {
			t.Errorf("Missing hook command: %s", hook)
		}
	}
}
