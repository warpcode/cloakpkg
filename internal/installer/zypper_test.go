package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"reflect"
	"testing"
)

func TestZypperName(t *testing.T) {
	z := &Zypper{}
	if name := z.Name(); name != "zypper" {
		t.Errorf("Expected name 'zypper', got '%s'", name)
	}
}

func TestZypperAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() { runner.CommandExists = origExists }()

	runner.CommandExists = func(name string) bool {
		return name == "zypper"
	}

	z := &Zypper{}
	if !z.Available() {
		t.Error("Expected Zypper to be available")
	}

	runner.CommandExists = func(name string) bool {
		return false
	}
	if z.Available() {
		t.Error("Expected Zypper to not be available")
	}
}

func TestZypperInstalled(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	defer func() { runner.DefaultCheckExecutor = origCheck }()

	z := &Zypper{}
	pkg := config.Package{Name: "test-package"}

	// Mock installed
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "rpm" && len(args) == 2 && args[0] == "-q" && args[1] == "test-package" {
			return nil
		}
		return fmt.Errorf("unexpected command")
	}

	if !z.Installed(pkg) {
		t.Error("Expected package to be considered installed")
	}

	// Mock not installed
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		return fmt.Errorf("not installed")
	}

	if z.Installed(pkg) {
		t.Error("Expected package to not be considered installed")
	}
}

func TestZypperAddRepositories(t *testing.T) {
	z := &Zypper{}
	err := z.AddRepositories(false, false, []config.Repository{})
	if err != nil {
		t.Errorf("Expected nil error from AddRepositories, got: %v", err)
	}
}

func TestZypperInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		return fmt.Errorf("not installed")
	}

	z := &Zypper{}
	pkgs := []config.Package{
		{Name: "curl"},
		{Name: "wget", ExtraParams: []string{"--no-recommends"}},
	}

	err := z.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	// Order of keys might vary since it relies on a map in GroupPackagesByExtraParams
	// We'll just check that both commands are present.
	expected1 := []string{"sudo", "zypper", "install", "-y", "curl"}
	expected2 := []string{"sudo", "zypper", "install", "-y", "--no-recommends", "wget"}

	found1 := false
	found2 := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, expected1) {
			found1 = true
		} else if reflect.DeepEqual(cmd, expected2) {
			found2 = true
		}
	}

	if !found1 {
		t.Errorf("Expected command not found: %v\nExecuted: %v", expected1, executedCmds)
	}
	if !found2 {
		t.Errorf("Expected command not found: %v\nExecuted: %v", expected2, executedCmds)
	}
}

func TestZypperUninstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		return nil // mock as installed
	}

	z := &Zypper{}
	pkgs := []config.Package{
		{Name: "curl"},
		{Name: "wget", ExtraParams: []string{"--clean-deps"}},
	}

	err := z.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	expected1 := []string{"sudo", "zypper", "remove", "-y", "curl"}
	expected2 := []string{"sudo", "zypper", "remove", "-y", "--clean-deps", "wget"}

	found1 := false
	found2 := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, expected1) {
			found1 = true
		} else if reflect.DeepEqual(cmd, expected2) {
			found2 = true
		}
	}

	if !found1 {
		t.Errorf("Expected command not found: %v\nExecuted: %v", expected1, executedCmds)
	}
	if !found2 {
		t.Errorf("Expected command not found: %v\nExecuted: %v", expected2, executedCmds)
	}
}

func TestZypperUpdate(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	z := &Zypper{}
	pkgs := []config.Package{
		{Name: "curl"},
		{Name: "wget", ExtraParams: []string{"--repo=updates"}},
	}

	err := z.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	expected1 := []string{"sudo", "zypper", "update", "-y", "curl"}
	expected2 := []string{"sudo", "zypper", "update", "-y", "--repo=updates", "wget"}

	found1 := false
	found2 := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, expected1) {
			found1 = true
		} else if reflect.DeepEqual(cmd, expected2) {
			found2 = true
		}
	}

	if !found1 {
		t.Errorf("Expected command not found: %v\nExecuted: %v", expected1, executedCmds)
	}
	if !found2 {
		t.Errorf("Expected command not found: %v\nExecuted: %v", expected2, executedCmds)
	}
}
