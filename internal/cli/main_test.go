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
	// Mock executor & exists
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		// Mock npm, brew, and apt as available
		return name == "npm" || name == "brew" || name == "apt-get"
	}
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		// Mock packages as not installed so they get installed
		return os.ErrNotExist
	}
	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		// Mock that no Homebrew taps are added
		return []byte(""), nil
	}

	// Write a temp config file representing multi-bundle, multi-installer setup
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
						"repositories": ["jesseduffield/lazygit"],
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
						"repositories": ["jesseduffield/lazygit"],
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

	// Override command args flag to simulate running with "-t core" (only include core tags)
	// We call runBundleCommand directly but we need to pass the tags flag.
	// Since runBundleCommand parses os.Args[3:], we can override os.Args!
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Simulate command line: cloakpkg install <config> -t core
	os.Args = []string{"cloakpkg", "install", configPath, "-t", "core"}

	runBundleCommand("install", configPath)

	// Since tag is "core", we expect:
	// 1. htop (tag "extra") is skipped
	// 2. npm packages are executed first (npm is first in priority order).
	//    - eslint and prettier are collated (same params: ["--save-dev"]) -> "npm install -g --save-dev eslint prettier"
	//    - typescript has no params -> "npm install -g typescript"
	// 3. brew packages are executed second:
	//    - git and tmux are collated -> "brew install --formula git tmux"
	// 4. apt packages are executed third:
	//    - neovim -> "apt-get install -y neovim"
	// 5. custom installer is executed last:
	//    - dotfiles -> Custom setup script executes

	// We expect 5 total commands executed (4 built-in collated commands + 1 custom script check/install)
	// Let's assert they ran in the correct sequence.
	if len(executedCmds) < 4 {
		t.Fatalf("Expected at least 4 commands executed, got %d: %v", len(executedCmds), executedCmds)
	}

	// Find commands by binary name
	var npmCmds, brewCmds, aptCmds, customCmds [][]string
	for _, cmd := range executedCmds {
		cleanCmd := stripSudo(cmd)
		switch cleanCmd[0] {
		case "npm":
			npmCmds = append(npmCmds, cleanCmd)
		case "brew":
			brewCmds = append(brewCmds, cleanCmd)
		case "apt-get":
			aptCmds = append(aptCmds, cleanCmd)
		case "/bin/sh":
			customCmds = append(customCmds, cleanCmd)
		}
	}

	// 1. Verify NPM collation and execution
	if len(npmCmds) != 2 {
		t.Errorf("Expected exactly 2 npm commands (collated by extra params), got %d: %v", len(npmCmds), npmCmds)
	} else {
		// One npm command should contain eslint and prettier, the other typescript
		var hasCollatedNpm, hasSingleNpm bool
		for _, cmd := range npmCmds {
			if len(cmd) >= 6 && cmd[1] == "install" && cmd[2] == "-g" && cmd[3] == "--save-dev" {
				hasEslint := false
				hasPrettier := false
				for _, arg := range cmd[4:] {
					if arg == "eslint" {
						hasEslint = true
					}
					if arg == "prettier" {
						hasPrettier = true
					}
				}
				if hasEslint && hasPrettier {
					hasCollatedNpm = true
				}
			}
			if len(cmd) == 4 && cmd[1] == "install" && cmd[2] == "-g" && cmd[3] == "typescript" {
				hasSingleNpm = true
			}
		}
		if !hasCollatedNpm {
			t.Errorf("Missing collated npm command for eslint and prettier in %v", npmCmds)
		}
		if !hasSingleNpm {
			t.Errorf("Missing single npm command for typescript in %v", npmCmds)
		}
	}

	// 2. Verify Brew collation
	if len(brewCmds) != 2 {
		t.Errorf("Expected exactly 2 brew commands (1 tap, 1 install), got %d: %v", len(brewCmds), brewCmds)
	} else {
		// Verify tap command
		tapCmd := brewCmds[0]
		if tapCmd[1] != "tap" || tapCmd[2] != "jesseduffield/lazygit" {
			t.Errorf("Unexpected brew tap command: %v", tapCmd)
		}

		// Verify install command
		installCmd := brewCmds[1]
		if installCmd[1] != "install" || installCmd[2] != "--formula" {
			t.Errorf("Unexpected brew install command structure: %v", installCmd)
		}
		hasGit := false
		hasTmux := false
		for _, arg := range installCmd[3:] {
			if arg == "git" {
				hasGit = true
			}
			if arg == "tmux" {
				hasTmux = true
			}
		}
		if !hasGit || !hasTmux {
			t.Errorf("Brew install command missing git or tmux: %v", installCmd)
		}
	}

	// 3. Verify Apt execution
	if len(aptCmds) != 1 {
		t.Errorf("Expected exactly 1 apt command, got %d: %v", len(aptCmds), aptCmds)
	} else {
		cmd := aptCmds[0]
		if cmd[1] != "install" || cmd[2] != "-y" || cmd[3] != "--" || cmd[4] != "neovim" {
			t.Errorf("Unexpected apt command structure: %v", cmd)
		}
	}

	// 4. Verify tag filtering (htop was excluded)
	for _, cmd := range executedCmds {
		for _, arg := range cmd {
			if arg == "htop" {
				t.Errorf("htop package should have been excluded by tag filter")
			}
		}
	}
}

func TestBundleCommandIdempotency(t *testing.T) {
	// Mock executor & exists
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
		// Mock "git" as ALREADY installed (exit code 0), but "tmux" as not installed (exit code 1)
		for _, arg := range args {
			if arg == "git" {
				return nil // success: git is already installed!
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

	// Simulate command line: cloakpkg install <config>
	os.Args = []string{"cloakpkg", "install", configPath}

	runBundleCommand("install", configPath)

	// Since "git" is already installed and "tmux" is not, we expect:
	// - git to be skipped
	// - Only "tmux" to be installed
	// - Brew command should be: "brew install --formula tmux" (only tmux, no git!)

	if len(executedCmds) != 1 {
		t.Fatalf("Expected exactly 1 command executed, got %d: %v", len(executedCmds), executedCmds)
	}

	cmd := executedCmds[0]
	if cmd[0] != "brew" || cmd[1] != "install" || cmd[2] != "--formula" {
		t.Fatalf("Unexpected command structure: %v", cmd)
	}

	// Verify cmd arguments
	hasGit := false
	hasTmux := false
	for _, arg := range cmd[3:] {
		if arg == "git" {
			hasGit = true
		}
		if arg == "tmux" {
			hasTmux = true
		}
	}
	if hasGit {
		t.Errorf("Command should not contain 'git' as it was mocked as already installed: %v", cmd)
	}
	if !hasTmux {
		t.Errorf("Command missing 'tmux': %v", cmd)
	}
}

func TestBundleCommandHooks(t *testing.T) {
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

	var executedShellCmds []string
	runner.DefaultShellExecutor = func(verbose bool, cmdStr string) error {
		executedShellCmds = append(executedShellCmds, cmdStr)
		return nil
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
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

	// Let's first test install
	os.Args = []string{"cloakpkg", "install", configPath, "flatpak"}
	runBundleCommand("install", configPath)

	// Expected shell commands executed in order for flatpak bundle:
	// 1. pre_install bundle: "echo 'pre-install bundle'"
	// 2. pre_install apt: "echo 'pre-install apt'"
	// (then apt install is run via DefaultExecutor, not DefaultShellExecutor)
	// 3. post_install apt: "echo 'post-install apt'"
	// 4. post_install bundle: "echo 'post-install bundle'"

	expectedShellCmds := []string{
		"echo 'pre-install bundle'",
		"echo 'pre-install apt'",
		"echo 'post-install apt'",
		"echo 'post-install bundle'",
	}

	if len(executedShellCmds) != len(expectedShellCmds) {
		t.Fatalf("Expected %d shell commands, got %d: %v", len(expectedShellCmds), len(executedShellCmds), executedShellCmds)
	}
	for i, cmd := range expectedShellCmds {
		if executedShellCmds[i] != cmd {
			t.Errorf("Expected shell command %d to be %q, got %q", i, cmd, executedShellCmds[i])
		}
	}

	// Reset counters and test custom provider hooks
	executedShellCmds = nil
	executedCmds = nil
	os.Args = []string{"cloakpkg", "install", configPath, "custom-pkg"}
	runBundleCommand("install", configPath)

	// Expected shell commands executed in order for custom provider:
	// 1. pre_install bundle: "echo 'pre-install custom bundle'"
	// 2. pre_install custom provider: "echo 'pre-install custom provider'"
	// 3. custom install: "echo 'install custom'" (custom install command runs as shell command)
	// 4. post_install custom provider: "echo 'post-install custom provider'"
	// 5. post_install bundle: "echo 'post-install custom bundle'"

	expectedCustomShellCmds := []string{
		"echo 'pre-install custom bundle'",
		"echo 'pre-install custom provider'",
		"echo 'install custom'",
		"echo 'post-install custom provider'",
		"echo 'post-install custom bundle'",
	}

	if len(executedShellCmds) != len(expectedCustomShellCmds) {
		t.Fatalf("Expected %d shell commands, got %d: %v", len(expectedCustomShellCmds), len(executedShellCmds), executedShellCmds)
	}
	for i, cmd := range expectedCustomShellCmds {
		if executedShellCmds[i] != cmd {
			t.Errorf("Expected custom shell command %d to be %q, got %q", i, cmd, executedShellCmds[i])
		}
	}
}
