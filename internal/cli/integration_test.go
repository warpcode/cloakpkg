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
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		expectedPkgs := map[string]bool{
			"flatpak": true,
			"docker-ce": true,
			"docker-ce-cli": true,
			"containerd.io": true,
			"docker-buildx-plugin": true,
			"docker-compose-plugin": true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Apt missing integration packages: %v", expectedPkgs)
		}
	})

	// Test Case 2: Pacman is available
	t.Run("Pacman", func(t *testing.T) {
		executed := runTestFile(t, "integration.json", mockEnv{
			availableCmds: []string{"pacman"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		expectedPkgs := map[string]bool{
			"flatpak": true,
			"docker": true,
			"docker-compose": true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Pacman missing integration packages: %v", expectedPkgs)
		}
	})

	// Test Case 3: Dnf is available
	t.Run("Dnf", func(t *testing.T) {
		executed := runTestFile(t, "integration.json", mockEnv{
			availableCmds: []string{"dnf"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		expectedPkgs := map[string]bool{
			"flatpak": true,
			"docker-ce": true,
			"docker-ce-cli": true,
			"containerd.io": true,
			"docker-buildx-plugin": true,
			"docker-compose-plugin": true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Dnf missing integration packages: %v", expectedPkgs)
		}
	})

	// Test Case 4: Brew is available
	t.Run("Brew", func(t *testing.T) {
		executed := runTestFile(t, "integration.json", mockEnv{
			availableCmds: []string{"brew"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		if cmd[0] != "brew" || cmd[1] != "install" || cmd[2] != "--cask" || cmd[3] != "docker-desktop" {
			t.Errorf("Unexpected brew docker-desktop installation: %v", cmd)
		}
	})
}
