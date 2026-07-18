package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"errors"
	"reflect"
	"testing"
)

func TestPacman_NameAndAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() {
		runner.CommandExists = origExists
	}()

	runner.CommandExists = func(name string) bool {
		return name == "pacman"
	}

	pacman := &Pacman{}
	if pacman.Name() != "pacman" {
		t.Errorf("Expected name 'pacman', got '%s'", pacman.Name())
	}
	if !pacman.Available() {
		t.Error("Expected pacman to be available")
	}
}

func TestPacman_Installed(t *testing.T) {
	origCheckExecutor := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultCheckExecutor = origCheckExecutor
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "pacman" && len(args) == 3 && args[0] == "-Qq" && args[1] == "--" && args[2] == "installed-pkg" {
			return nil
		}
		return errors.New("not installed")
	}

	pacman := &Pacman{}
	if !pacman.Installed(config.Package{Name: "installed-pkg"}) {
		t.Error("Expected installed-pkg to be installed")
	}
	if pacman.Installed(config.Package{Name: "not-installed-pkg"}) {
		t.Error("Expected not-installed-pkg to not be installed")
	}
}

func TestPacman_AddRepositories(t *testing.T) {
	pacman := &Pacman{}
	repos := []config.Repository{
		{Source: "some-repo"},
	}

	err := pacman.AddRepositories(false, false, repos)
	if err != nil {
		t.Fatalf("AddRepositories failed: %v", err)
	}
}

func TestPacman_Install(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultExecutor = origExecutor
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "pacman" && len(args) == 3 && args[0] == "-Qq" && args[1] == "--" {
			if args[2] == "installed-pkg" {
				return nil
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

	pacman := &Pacman{}
	pkgs := []config.Package{
		{Name: "installed-pkg"},
		{Name: "new-pkg"},
		{Name: "extra-pkg", ExtraParams: []string{"--asdeps"}},
	}

	err := pacman.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundNewPkg := false
	foundExtraPkg := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, []string{"sudo", "pacman", "-S", "--noconfirm", "--", "new-pkg"}) {
			foundNewPkg = true
		} else if reflect.DeepEqual(cmd, []string{"sudo", "pacman", "-S", "--noconfirm", "--asdeps", "--", "extra-pkg"}) {
			foundExtraPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundNewPkg || !foundExtraPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}

func TestPacman_Uninstall(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultExecutor = origExecutor
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "pacman" && len(args) == 3 && args[0] == "-Qq" && args[1] == "--" {
			if args[2] == "installed-pkg" || args[2] == "extra-pkg" {
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

	pacman := &Pacman{}
	pkgs := []config.Package{
		{Name: "installed-pkg"},
		{Name: "not-installed-pkg"},
		{Name: "extra-pkg", ExtraParams: []string{"--nosave"}},
	}

	err := pacman.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundInstalledPkg := false
	foundExtraPkg := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, []string{"sudo", "pacman", "-R", "--noconfirm", "--", "installed-pkg"}) {
			foundInstalledPkg = true
		} else if reflect.DeepEqual(cmd, []string{"sudo", "pacman", "-R", "--noconfirm", "--nosave", "--", "extra-pkg"}) {
			foundExtraPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundInstalledPkg || !foundExtraPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}

func TestPacman_Update(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	pacman := &Pacman{}
	pkgs := []config.Package{
		{Name: "pkg1"},
		{Name: "pkg2"},
		{Name: "extra-pkg", ExtraParams: []string{"--asdeps"}},
	}

	err := pacman.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundPkgs := false
	foundExtraPkg := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, []string{"sudo", "pacman", "-S", "--noconfirm", "--", "pkg1", "pkg2"}) || reflect.DeepEqual(cmd, []string{"sudo", "pacman", "-S", "--noconfirm", "--", "pkg2", "pkg1"}) {
			foundPkgs = true
		} else if reflect.DeepEqual(cmd, []string{"sudo", "pacman", "-S", "--noconfirm", "--asdeps", "--", "extra-pkg"}) {
			foundExtraPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundPkgs || !foundExtraPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}
