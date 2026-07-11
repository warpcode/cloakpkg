package cli

import (
	"testing"
)

func TestAppsPackages(t *testing.T) {
	// Test Case: Flatpak and Snap and Apt available
	t.Run("FlatpakSnapApt", func(t *testing.T) {
		executed := runTestFile(t, "apps.json", mockEnv{
			availableCmds: []string{"flatpak", "snap", "apt-get"},
		})
		// We expect 3 main install commands:
		// 1. flatpak: com.discordapp.Discord, org.keepassxc.KeePassXC
		// 2. snap: kontena-lens (with --classic)
		// 3. apt: ffmpeg, code, cursor

		flatpakCmd := findCommand(executed, "flatpak", "install")
		snapCmd := findCommand(executed, "snap", "install")
		aptCmd := findCommand(executed, "apt-get", "install")

		if flatpakCmd == nil {
			t.Errorf("Missing flatpak command")
		} else {
			expected := map[string]bool{
				"com.discordapp.Discord": true, "org.keepassxc.KeePassXC": true,
			}
			for _, arg := range flatpakCmd[2:] {
				delete(expected, arg)
			}
			if len(expected) > 0 {
				t.Errorf("Flatpak command missing apps: %v", expected)
			}
		}

		if len(snapCmd) == 0 {
			t.Errorf("Missing snap command")
		} else {
			// Expect snap install --classic kontena-lens
			hasClassic := false
			hasLens := false
			for _, arg := range snapCmd {
				if arg == "--classic" {
					hasClassic = true
				}
				if arg == "kontena-lens" {
					hasLens = true
				}
			}
			if !hasClassic || !hasLens {
				t.Errorf("Expected snap command to install kontena-lens classic, got: %v", snapCmd)
			}
		}

		if len(aptCmd) == 0 {
			t.Errorf("Missing apt command")
		} else {
			expected := map[string]bool{
				"ffmpeg": true, "code": true, "cursor": true,
			}
			for _, arg := range aptCmd[3:] {
				delete(expected, arg)
			}
			if len(expected) > 0 {
				t.Errorf("Apt command missing apps: %v", expected)
			}
		}
	})

	// Test Case 2: Brew is available
	t.Run("Brew", func(t *testing.T) {
		executed := runTestFile(t, "apps.json", mockEnv{
			availableCmds: []string{"brew"},
		})
		// We expect 2 commands:
		// 1. brew install ffmpeg (standard)
		// 2. brew install --cask discord lens keepassxc visual-studio-code cursor
		if len(executed) != 2 {
			t.Fatalf("Expected 2 commands executed, got %d: %v", len(executed), executed)
		}

		var standardCmd, caskCmd []string
		for _, cmd := range executed {
			cleanCmd := stripSudo(cmd)
			isCask := false
			for _, arg := range cleanCmd {
				if arg == "--cask" {
					isCask = true
					break
				}
			}
			if isCask {
				caskCmd = cleanCmd
			} else {
				standardCmd = cleanCmd
			}
		}

		if len(standardCmd) == 0 {
			t.Errorf("Missing brew standard command")
		} else {
			if standardCmd[len(standardCmd)-1] != "ffmpeg" {
				t.Errorf("Expected standard brew command to install ffmpeg, got: %v", standardCmd)
			}
		}

		if len(caskCmd) == 0 {
			t.Errorf("Missing brew cask command")
		} else {
			expected := map[string]bool{
				"discord": true, "lens": true, "keepassxc": true,
				"visual-studio-code": true, "cursor": true,
			}
			for _, arg := range caskCmd[2:] {
				delete(expected, arg)
			}
			if len(expected) > 0 {
				t.Errorf("Brew cask missing packages: %v", expected)
			}
		}
	})
}
