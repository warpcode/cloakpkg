package cli

import (
	"testing"
)

func TestIntegrationPackages(t *testing.T) {
	// Test Case 1: Apt is available
	t.Run("Apt", func(t *testing.T) {
		executed := runTestFile(t, "integration.json", mockEnv{
			availableCmds: []string{"apt-get"},
		})

		if findCommand(executed, "apt-get", "install", "docker-ce") == nil {
			t.Errorf("Missing apt install command for docker-ce")
		}
		if findCommand(executed, "apt-get", "update") == nil {
			t.Errorf("Missing apt-get update command")
		}
		if findCommand(executed, "cp") == nil {
			t.Errorf("Missing cp command for repo setup")
		}
	})

	// Test Case 2: Pacman is available
	t.Run("Pacman", func(t *testing.T) {
		executed := runTestFile(t, "integration.json", mockEnv{
			availableCmds: []string{"pacman"},
		})
		if findCommand(executed, "pacman", "-S", "--noconfirm", "docker") == nil {
			t.Errorf("Missing pacman install command for docker")
		}
	})

	// Test Case 3: Dnf is available
	t.Run("Dnf", func(t *testing.T) {
		executed := runTestFile(t, "integration.json", mockEnv{
			availableCmds: []string{"dnf"},
		})
		if findCommand(executed, "dnf", "install", "docker-ce") == nil {
			t.Errorf("Missing dnf install command for docker-ce")
		}
		if findCommand(executed, "dnf", "config-manager", "--add-repo") == nil {
			t.Errorf("Missing dnf repo setup command")
		}
	})

	// Test Case 4: Brew is available
	t.Run("Brew", func(t *testing.T) {
		executed := runTestFile(t, "integration.json", mockEnv{
			availableCmds: []string{"brew"},
		})
		if findCommand(executed, "brew", "install", "--cask", "docker-desktop") == nil {
			t.Errorf("Missing brew install command for docker-desktop")
		}
	})
}
