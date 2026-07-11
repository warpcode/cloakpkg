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

		if findCommand(executed, "flatpak", "install", "com.discordapp.Discord") == nil {
			t.Errorf("Missing flatpak install command")
		}
		if findCommand(executed, "snap", "install", "kontena-lens") == nil {
			t.Errorf("Missing snap install command")
		}
		if findCommand(executed, "apt-get", "install", "ffmpeg") == nil {
			t.Errorf("Missing apt install command")
		}
	})

	// Test Case 2: Brew is available
	t.Run("Brew", func(t *testing.T) {
		executed := runTestFile(t, "apps.json", mockEnv{
			availableCmds: []string{"brew"},
		})

		if findCommand(executed, "brew", "install", "ffmpeg") == nil {
			t.Errorf("Missing brew standard install command")
		}
		if findCommand(executed, "brew", "install", "--cask", "discord") == nil {
			t.Errorf("Missing brew cask install command")
		}
	})
}
