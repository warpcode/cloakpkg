package cli

import (
	"testing"
)

func TestSystemPackages(t *testing.T) {
	// Test case 1: APT is available
	t.Run("Apt", func(t *testing.T) {
		executed := runTestFile(t, "system.json", mockEnv{
			availableCmds: []string{"apt-get"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		if cmd[0] != "apt-get" || cmd[1] != "install" {
			t.Errorf("Unexpected command: %v", cmd)
		}
		expectedPkgs := map[string]bool{
			"git": true, "zsh": true, "curl": true, "gcc": true, "make": true,
			"ca-certificates": true, "gnupg2": true, "libatomic1": true,
			"openssl": true, "pkg-config": true, "unzip": true,
			"libsecret-tools": true, "bc": true, "dialog": true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Some expected packages were not installed: %v", expectedPkgs)
		}
	})

	// Test case 2: Brew is available
	t.Run("Brew", func(t *testing.T) {
		executed := runTestFile(t, "system.json", mockEnv{
			availableCmds: []string{"brew"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		if cmd[0] != "brew" || cmd[1] != "install" {
			t.Errorf("Unexpected command: %v", cmd)
		}
		expectedPkgs := map[string]bool{
			"git": true, "zsh": true, "curl": true, "gcc": true, "make": true,
			"gnupg": true, "openssl": true, "pkg-config": true, "unzip": true,
			"bc": true, "dialog": true,
		}
		for _, arg := range cmd[2:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Some expected packages were not installed: %v", expectedPkgs)
		}
	})

	// Test case 3: DNF is available
	t.Run("Dnf", func(t *testing.T) {
		executed := runTestFile(t, "system.json", mockEnv{
			availableCmds: []string{"dnf"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		if cmd[0] != "dnf" || cmd[1] != "install" {
			t.Errorf("Unexpected command: %v", cmd)
		}
		expectedPkgs := map[string]bool{
			"git": true, "zsh": true, "curl": true, "gcc": true, "make": true,
			"ca-certificates": true, "gnupg2": true, "libatomic": true,
			"openssl": true, "pkgconf-pkg-config": true, "unzip": true,
			"libsecret": true, "bc": true, "dialog": true, "dnf-plugins-core": true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Some expected packages were not installed: %v", expectedPkgs)
		}
	})

	// Test case 4: Pacman is available
	t.Run("Pacman", func(t *testing.T) {
		executed := runTestFile(t, "system.json", mockEnv{
			availableCmds: []string{"pacman"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		if cmd[0] != "pacman" || cmd[1] != "-S" {
			t.Errorf("Unexpected command: %v", cmd)
		}
		expectedPkgs := map[string]bool{
			"git": true, "zsh": true, "curl": true, "gcc": true, "make": true,
			"ca-certificates": true, "gcc-libs": true, "bc": true, "dialog": true,
		}
		for _, arg := range cmd[2:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Some expected packages were not installed: %v", expectedPkgs)
		}
	})

	// Test case 5: Termux is available
	t.Run("Termux", func(t *testing.T) {
		origEnv := func() {
			t.Setenv("TERMUX_VERSION", "0.118.0")
		}
		origEnv()

		executed := runTestFile(t, "system.json", mockEnv{
			availableCmds: []string{"pkg"},
		})
		if len(executed) != 1 {
			t.Fatalf("Expected 1 command executed, got %d: %v", len(executed), executed)
		}
		cmd := stripSudo(executed[0])
		if cmd[0] != "pkg" || cmd[1] != "install" {
			t.Errorf("Unexpected command: %v", cmd)
		}
		expectedPkgs := map[string]bool{
			"git": true, "zsh": true, "curl": true, "clang": true, "make": true,
			"gnupg": true, "openssl": true, "pkg-config": true, "unzip": true,
			"bc": true, "dialog": true,
		}
		for _, arg := range cmd[3:] {
			delete(expectedPkgs, arg)
		}
		if len(expectedPkgs) > 0 {
			t.Errorf("Some expected packages were not installed: %v", expectedPkgs)
		}
	})
}
