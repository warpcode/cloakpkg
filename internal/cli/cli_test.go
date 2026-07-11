package cli

import (
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
		// Let's assert we ran exactly 2 commands
		if len(executed) != 2 {
			t.Fatalf("Expected 2 commands executed, got %d: %v", len(executed), executed)
		}
		
		var miseCmd, brewCmd []string
		for _, cmd := range executed {
			if cmd[0] == "mise" {
				miseCmd = cmd
			} else if cmd[0] == "brew" {
				brewCmd = cmd
			}
		}

		if len(miseCmd) == 0 {
			t.Errorf("Missing mise installation command")
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
			"tmux": true,
			"screen": true,
			"rsync": true,
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
			"tmux": true,
			"screen": true,
			"rsync": true,
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
			"tmux": true,
			"screen": true,
			"rsync": true,
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
			"tmux": true,
			"screen": true,
			"rsync": true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Termux missing CLI packages: %v", expectedPkgs)
		}
	})
}
