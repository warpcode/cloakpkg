package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"errors"
	"reflect"
	"testing"
)

func TestSnap_NameAndAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() {
		runner.CommandExists = origExists
	}()

	runner.CommandExists = func(name string) bool {
		return name == "snap"
	}

	snap := &Snap{}
	if snap.Name() != "snap" {
		t.Errorf("Expected name 'snap', got '%s'", snap.Name())
	}
	if !snap.Available() {
		t.Error("Expected snap to be available")
	}
}

func TestSnap_AddRepositories(t *testing.T) {
	snap := &Snap{}
	repos := []config.Repository{
		{Source: "some-repo"},
	}

	// AddRepositories should just return nil for snap
	err := snap.AddRepositories(false, false, repos)
	if err != nil {
		t.Fatalf("AddRepositories failed: %v", err)
	}
}

func TestSnap_Installed(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "snap" && len(args) == 2 && args[0] == "info" {
			if args[1] == "installed-pkg" {
				return nil
			}
			return errors.New("not installed")
		}
		return errors.New("unexpected command")
	}

	snap := &Snap{}
	if !snap.Installed(config.Package{Name: "installed-pkg"}) {
		t.Error("Expected package to be reported as installed")
	}
	if snap.Installed(config.Package{Name: "missing-pkg"}) {
		t.Error("Expected package to be reported as not installed")
	}
}

func TestSnap_Install(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultExecutor = origExecutor
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "snap" && len(args) == 2 && args[0] == "info" {
			if args[1] == "installed-pkg" {
				return nil
			}
			return errors.New("not installed")
		}
		return errors.New("unexpected command")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		// When RunSudo is called (e.g. for install, uninstall, update), it either
		// invokes 'sudo snap ...' or 'snap ...' depending on OS/uid.
		// For simplicity, we just collect whatever DefaultExecutor receives.
		// In our tests, RunSudo will pass 'sudo' as bin and 'snap', etc. as args,
		// or 'snap' as bin on Windows/root. Let's just flatten it for testing.
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	snap := &Snap{}
	pkgs := []config.Package{
		{Name: "installed-pkg"}, // Should be skipped
		{Name: "new-pkg"},
		{Name: "classic-pkg", ExtraParams: []string{"--classic"}},
	}

	err := snap.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundNewPkg := false
	foundClassicPkg := false

	for _, cmd := range executedCmds {
		// Because RunSudo may or may not prepend 'sudo', we will check if the command slice
		// matches the expected suffix of arguments.
		// A typical command slice here would be ["sudo", "snap", "install", "--", "new-pkg"]
		// or ["snap", "install", "--", "new-pkg"].
		// We'll extract the arguments ignoring 'sudo' if present.

		var effectiveCmd []string
		if cmd[0] == "sudo" {
			effectiveCmd = cmd[1:]
		} else {
			effectiveCmd = cmd
		}

		if reflect.DeepEqual(effectiveCmd, []string{"snap", "install", "--", "new-pkg"}) {
			foundNewPkg = true
		} else if reflect.DeepEqual(effectiveCmd, []string{"snap", "install", "--classic", "--", "classic-pkg"}) {
			foundClassicPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundNewPkg || !foundClassicPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}

func TestSnap_Uninstall(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultExecutor = origExecutor
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "snap" && len(args) == 2 && args[0] == "info" {
			if args[1] == "installed-pkg" || args[1] == "classic-pkg" {
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

	snap := &Snap{}
	pkgs := []config.Package{
		{Name: "installed-pkg"},
		{Name: "not-installed-pkg"}, // Should be skipped
		{Name: "classic-pkg", ExtraParams: []string{"--classic", "--purge"}},
	}

	err := snap.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundInstalledPkg := false
	foundClassicPkg := false

	for _, cmd := range executedCmds {
		var effectiveCmd []string
		if cmd[0] == "sudo" {
			effectiveCmd = cmd[1:]
		} else {
			effectiveCmd = cmd
		}

		if reflect.DeepEqual(effectiveCmd, []string{"snap", "remove", "--", "installed-pkg"}) {
			foundInstalledPkg = true
		} else if reflect.DeepEqual(effectiveCmd, []string{"snap", "remove", "--purge", "--", "classic-pkg"}) { // --classic should be filtered out
			foundClassicPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundInstalledPkg || !foundClassicPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}

func TestSnap_Update(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	snap := &Snap{}
	pkgs := []config.Package{
		{Name: "pkg1"},
		{Name: "pkg2"},
		{Name: "classic-pkg", ExtraParams: []string{"--classic"}},
	}

	err := snap.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundPkgs := false
	foundClassicPkg := false

	for _, cmd := range executedCmds {
		var effectiveCmd []string
		if cmd[0] == "sudo" {
			effectiveCmd = cmd[1:]
		} else {
			effectiveCmd = cmd
		}

		if reflect.DeepEqual(effectiveCmd, []string{"snap", "refresh", "--", "pkg1", "pkg2"}) || reflect.DeepEqual(effectiveCmd, []string{"snap", "refresh", "--", "pkg2", "pkg1"}) {
			foundPkgs = true
		} else if reflect.DeepEqual(effectiveCmd, []string{"snap", "refresh", "--classic", "--", "classic-pkg"}) {
			foundClassicPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundPkgs || !foundClassicPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}
