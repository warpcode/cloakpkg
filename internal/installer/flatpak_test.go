package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"testing"
)

func TestFlatpakAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() { runner.CommandExists = origExists }()

	runner.CommandExists = func(name string) bool {
		return name == "flatpak"
	}

	f := &Flatpak{}
	if !f.Available() {
		t.Error("Expected Flatpak to be available")
	}

	runner.CommandExists = func(name string) bool {
		return false
	}

	if f.Available() {
		t.Error("Expected Flatpak to be unavailable")
	}
}

func TestFlatpakInstalled(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	defer func() { runner.DefaultCheckExecutor = origCheck }()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "flatpak" && len(args) == 2 && args[0] == "info" && args[1] == "installed-pkg" {
			return nil
		}
		return fmt.Errorf("not installed")
	}

	f := &Flatpak{}
	if !f.Installed(config.Package{Name: "installed-pkg"}) {
		t.Error("Expected package to be reported as installed")
	}
	if f.Installed(config.Package{Name: "missing-pkg"}) {
		t.Error("Expected package to be reported as not installed")
	}
}

func TestFlatpakInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if len(args) > 1 && args[1] == "already-installed" {
			return nil
		}
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	f := &Flatpak{}
	pkgs := []config.Package{
		{Name: "new-pkg", ExtraParams: []string{"--system"}},
		{Name: "already-installed", ExtraParams: []string{"--system"}}, // Should be skipped
		{Name: "another-new-pkg"},
	}

	err := f.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// We expect two groups based on ExtraParams
	// group 1: --system (contains new-pkg and already-installed)
	// group 2: none (contains another-new-pkg)
	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	// Validate the commands
	foundSystemCmd := false
	foundNormalCmd := false
	for _, cmd := range executedCmds {
		if len(cmd) > 3 && cmd[3] == "--system" {
			foundSystemCmd = true
			if cmd[4] != "new-pkg" {
				t.Errorf("Unexpected package in --system command: %v", cmd)
			}
		} else if len(cmd) > 3 && cmd[3] == "another-new-pkg" {
			foundNormalCmd = true
		}
	}

	if !foundSystemCmd {
		t.Error("Did not find expected --system install command")
	}
	if !foundNormalCmd {
		t.Error("Did not find expected normal install command")
	}
}

func TestFlatpakUninstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if len(args) > 1 && args[1] == "installed-pkg" {
			return nil
		}
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	f := &Flatpak{}
	pkgs := []config.Package{
		{Name: "installed-pkg"},
		{Name: "missing-pkg"}, // Should be skipped
	}

	err := f.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "flatpak" || cmd[1] != "uninstall" || cmd[2] != "-y" || cmd[3] != "installed-pkg" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}

func TestFlatpakUpdate(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	f := &Flatpak{}
	pkgs := []config.Package{
		{Name: "pkg1"},
		{Name: "pkg2", ExtraParams: []string{"--user"}},
	}

	err := f.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	// Commands can be in any order because map iteration is random
	foundNormalCmd := false
	foundUserCmd := false
	for _, cmd := range executedCmds {
		if len(cmd) == 4 && cmd[0] == "flatpak" && cmd[1] == "update" && cmd[2] == "-y" && cmd[3] == "pkg1" {
			foundNormalCmd = true
		} else if len(cmd) == 5 && cmd[0] == "flatpak" && cmd[1] == "update" && cmd[2] == "-y" && cmd[3] == "--user" && cmd[4] == "pkg2" {
			foundUserCmd = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundNormalCmd {
		t.Error("Did not find expected normal update command")
	}
	if !foundUserCmd {
		t.Error("Did not find expected --user update command")
	}
}

func TestFlatpakAddRepositories(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	f := &Flatpak{}
	repos := []config.Repository{
		{Source: "flathub https://dl.flathub.org/repo/flathub.flatpakrepo"}, // space separated
		{Remote: "custom", Source: "https://custom.org/repo"}, // Explicit Remote and Source
		{Source: "invalid"}, // invalid source, should return error
	}

	// Test with valid repos
	err := f.AddRepositories(false, false, repos[:2])
	if err != nil {
		t.Fatalf("AddRepositories failed: %v", err)
	}

	if len(executedCmds) != 3 {
		t.Fatalf("Expected 3 commands executed (2 add, 1 update appstream), got %d: %v", len(executedCmds), executedCmds)
	}

	if executedCmds[0][0] != "flatpak" || executedCmds[0][1] != "remote-add" || executedCmds[0][2] != "--if-not-exists" || executedCmds[0][3] != "flathub" || executedCmds[0][4] != "https://dl.flathub.org/repo/flathub.flatpakrepo" {
		t.Errorf("Unexpected command 1 executed: %v", executedCmds[0])
	}
	if executedCmds[1][0] != "flatpak" || executedCmds[1][1] != "remote-add" || executedCmds[1][2] != "--if-not-exists" || executedCmds[1][3] != "custom" || executedCmds[1][4] != "https://custom.org/repo" {
		t.Errorf("Unexpected command 2 executed: %v", executedCmds[1])
	}
	if executedCmds[2][0] != "flatpak" || executedCmds[2][1] != "update" || executedCmds[2][2] != "--appstream" {
		t.Errorf("Unexpected command 3 executed: %v", executedCmds[2])
	}

	// Reset executedCmds
	executedCmds = nil

	// Test with invalid repo
	err = f.AddRepositories(false, false, repos[2:])
	if err == nil {
		t.Fatalf("Expected AddRepositories to fail with invalid repo format")
	}
	if len(executedCmds) != 0 {
		t.Fatalf("Expected no commands executed on error, got %d", len(executedCmds))
	}
}
