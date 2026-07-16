package installer

import (
	"fmt"
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"testing"
)

func TestNpmInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		// Mock everything as not installed
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "npm"
	}

	npm := &Npm{}
	if !npm.Available() {
		t.Error("Npm should be available")
	}

	pkgs := []config.Package{
		{Name: "typescript"},
		{Name: "eslint", ExtraParams: []string{"--no-fund"}},
	}

	err := npm.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Because of GroupPackagesByExtraParams, we should get 2 separate commands
	// One for packages without extra params, one for packages with --no-fund
	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	// The order of groups from GroupPackagesByExtraParams is not strictly guaranteed as it uses a map internally, Wait, GroupPackagesByExtraParams returns keys, groups! Let's check what it does.
}
