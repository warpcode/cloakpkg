package cli

import (
	"testing"
)

func TestDevelopmentPackages(t *testing.T) {
	// Test Case: Mise, Npm, and Apt available
	t.Run("MiseNpmApt", func(t *testing.T) {
		executed := runTestFile(t, "development.json", mockEnv{
			availableCmds: []string{"mise", "npm", "apt-get"},
		})
		// We expect 3 commands:
		// 1. mise: python@3.12, node@latest, rust, go, deno
		// 2. npm: prettier, eslint, typescript-language-server
		// 3. apt: php-cli
		if len(executed) != 3 {
			t.Fatalf("Expected 3 commands executed, got %d: %v", len(executed), executed)
		}

		var miseCmd, npmCmd, aptCmd []string
		for _, cmd := range executed {
			cleanCmd := stripSudo(cmd)
			switch cleanCmd[0] {
			case "mise":
				miseCmd = cleanCmd
			case "npm":
				npmCmd = cleanCmd
			case "apt-get":
				aptCmd = cleanCmd
			}
		}

		if len(miseCmd) == 0 {
			t.Errorf("Missing mise command")
		} else {
			expected := map[string]bool{
				"python@3.12": true, "node@latest": true, "rust": true, "go": true, "deno": true,
			}
			for _, arg := range miseCmd[2:] {
				delete(expected, arg)
			}
			if len(expected) > 0 {
				t.Errorf("Mise command missing development packages: %v", expected)
			}
		}

		if len(npmCmd) == 0 {
			t.Errorf("Missing npm command")
		} else {
			expected := map[string]bool{
				"prettier": true, "eslint": true, "typescript-language-server": true,
			}
			for _, arg := range npmCmd[2:] {
				delete(expected, arg)
			}
			if len(expected) > 0 {
				t.Errorf("Npm command missing development packages: %v", expected)
			}
		}

		if len(aptCmd) == 0 {
			t.Errorf("Missing apt command")
		} else {
			lastArg := aptCmd[len(aptCmd)-1]
			if lastArg != "php-cli" {
				t.Errorf("Expected apt command to install php-cli, got: %v", aptCmd)
			}
		}
	})

	// Test Case 2: Dnf is available
	t.Run("Dnf", func(t *testing.T) {
		executed := runTestFile(t, "development.json", mockEnv{
			availableCmds: []string{"dnf"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		lastArg := cmd[len(cmd)-1]
		if lastArg != "php-cli" {
			t.Errorf("Expected dnf command to install php-cli, got: %v", cmd)
		}
	})

	// Test Case 3: Pacman is available
	t.Run("Pacman", func(t *testing.T) {
		executed := runTestFile(t, "development.json", mockEnv{
			availableCmds: []string{"pacman"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		lastArg := cmd[len(cmd)-1]
		if lastArg != "php" {
			t.Errorf("Expected pacman command to install php, got: %v", cmd)
		}
	})

	// Test Case 4: Brew is available
	t.Run("Brew", func(t *testing.T) {
		executed := runTestFile(t, "development.json", mockEnv{
			availableCmds: []string{"brew"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		lastArg := cmd[len(cmd)-1]
		if lastArg != "php" {
			t.Errorf("Expected brew command to install php, got: %v", cmd)
		}
	})
}
