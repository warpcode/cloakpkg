package installer

import (
	"errors"
	"reflect"
	"testing"
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
)

func TestBrew_NameAndAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() {
		runner.CommandExists = origExists
	}()

	runner.CommandExists = func(name string) bool {
		return name == "brew"
	}

	brew := &Brew{}
	if brew.Name() != "brew" {
		t.Errorf("Expected name 'brew', got '%s'", brew.Name())
	}
	if !brew.Available() {
		t.Error("Expected brew to be available")
	}
}

func TestBrew_AddRepositories(t *testing.T) {
	origCheckOutput := runner.DefaultCheckOutputExecutor
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultCheckOutputExecutor = origCheckOutput
		runner.DefaultExecutor = origExecutor
	}()

	// Mock `brew tap` output
	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		if bin == "brew" && len(args) > 0 && args[0] == "tap" {
			return []byte("homebrew/core\nhomebrew/cask\n"), nil
		}
		return nil, errors.New("unexpected command")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	brew := &Brew{}
	repos := []config.Repository{
		{Source: "homebrew/core"},        // Already tapped
		{Source: "homebrew/cask-fonts"},  // New tap
		{Source: ""},                     // Empty source should be skipped
	}

	err := brew.AddRepositories(false, false, repos)
	if err != nil {
		t.Fatalf("AddRepositories failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command to be executed, got %d", len(executedCmds))
	}

	expectedCmd := []string{"brew", "tap", "homebrew/cask-fonts"}
	if !reflect.DeepEqual(executedCmds[0], expectedCmd) {
		t.Errorf("Expected command %v, got %v", expectedCmd, executedCmds[0])
	}
}

func TestBrew_Install(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultExecutor = origExecutor
	}()

	// Mock `brew list <pkg>` (checking if installed)
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "brew" && len(args) == 2 && args[0] == "list" {
			if args[1] == "installed-pkg" {
				return nil // Installed
			}
			return errors.New("not installed")
		}
		return errors.New("unexpected command")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	brew := &Brew{}
	pkgs := []config.Package{
		{Name: "installed-pkg"},
		{Name: "new-pkg"},
		{Name: "cask-pkg", ExtraParams: []string{"--cask"}},
	}

	err := brew.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	// Commands can be executed in any group order, but usually map iteration order is unpredictable.
	// We'll check if both commands are present.
	foundNewPkg := false
	foundCaskPkg := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, []string{"brew", "install", "new-pkg"}) {
			foundNewPkg = true
		} else if reflect.DeepEqual(cmd, []string{"brew", "install", "--cask", "cask-pkg"}) {
			foundCaskPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundNewPkg || !foundCaskPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}

func TestBrew_Uninstall(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultExecutor = origExecutor
	}()

	// Mock `brew list <pkg>`
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "brew" && len(args) == 2 && args[0] == "list" {
			if args[1] == "installed-pkg" || args[1] == "cask-pkg" {
				return nil // Installed
			}
			return errors.New("not installed")
		}
		return errors.New("unexpected command")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	brew := &Brew{}
	pkgs := []config.Package{
		{Name: "installed-pkg"},
		{Name: "not-installed-pkg"},
		{Name: "cask-pkg", ExtraParams: []string{"--cask"}},
	}

	err := brew.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundInstalledPkg := false
	foundCaskPkg := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, []string{"brew", "uninstall", "installed-pkg"}) {
			foundInstalledPkg = true
		} else if reflect.DeepEqual(cmd, []string{"brew", "uninstall", "--cask", "cask-pkg"}) {
			foundCaskPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundInstalledPkg || !foundCaskPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}

func TestBrew_Update(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	brew := &Brew{}
	pkgs := []config.Package{
		{Name: "pkg1"},
		{Name: "pkg2"},
		{Name: "cask-pkg", ExtraParams: []string{"--cask"}},
	}

	err := brew.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundPkgs := false
	foundCaskPkg := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, []string{"brew", "upgrade", "pkg1", "pkg2"}) || reflect.DeepEqual(cmd, []string{"brew", "upgrade", "pkg2", "pkg1"}) {
			foundPkgs = true
		} else if reflect.DeepEqual(cmd, []string{"brew", "upgrade", "--cask", "cask-pkg"}) {
			foundCaskPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundPkgs || !foundCaskPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}
