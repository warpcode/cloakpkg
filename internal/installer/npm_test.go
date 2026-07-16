package installer

import (
	"fmt"
	"strings"
	"testing"

	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
)

func TestNpmName(t *testing.T) {
	npm := &Npm{}
	if name := npm.Name(); name != "npm" {
		t.Errorf("Expected name 'npm', got '%s'", name)
	}
}

func TestNpmAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() { runner.CommandExists = origExists }()

	runner.CommandExists = func(name string) bool {
		return name == "npm"
	}

	npm := &Npm{}
	if !npm.Available() {
		t.Error("Npm should be available when command exists")
	}

	runner.CommandExists = func(name string) bool {
		return false
	}
	if npm.Available() {
		t.Error("Npm should not be available when command doesn't exist")
	}
}

func TestNpmInstalled(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	defer func() { runner.DefaultCheckExecutor = origCheck }()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "npm" && len(args) == 3 && args[0] == "list" && args[1] == "-g" && args[2] == "typescript" {
			return nil
		}
		return fmt.Errorf("not installed")
	}

	npm := &Npm{}
	pkgInstalled := config.Package{Name: "typescript"}
	pkgNotInstalled := config.Package{Name: "eslint"}

	if !npm.Installed(pkgInstalled) {
		t.Error("typescript should be installed")
	}
	if npm.Installed(pkgNotInstalled) {
		t.Error("eslint should not be installed")
	}
}

func TestNpmAddRepositories(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() { runner.DefaultExecutor = origExecutor }()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	npm := &Npm{}
	repos := []config.Repository{
		{URL: "https://registry.npmjs.org/"},
	}

	err := npm.AddRepositories(false, false, repos)
	if err != nil {
		t.Errorf("AddRepositories returned error: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "npm" || cmd[1] != "config" || cmd[2] != "set" || cmd[3] != "registry" || cmd[4] != "https://registry.npmjs.org/" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}

func TestNpmInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if len(args) > 2 && args[2] == "already-installed" {
			return nil // installed
		}
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	npm := &Npm{}
	pkgs := []config.Package{
		{Name: "typescript"},
		{Name: "eslint", ExtraParams: []string{"--no-fund"}},
		{Name: "already-installed"},
	}

	err := npm.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	// Because order of groups is not guaranteed, check that both commands are present
	var hasTypescript, hasEslint bool
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "typescript") {
			if cmd[0] != "npm" || cmd[1] != "install" || cmd[2] != "-g" || cmd[3] != "typescript" {
				t.Errorf("Unexpected typescript install command: %v", cmd)
			}
			hasTypescript = true
		} else if strings.Contains(cmdStr, "eslint") {
			if cmd[0] != "npm" || cmd[1] != "install" || cmd[2] != "-g" || cmd[3] != "--no-fund" || cmd[4] != "eslint" {
				t.Errorf("Unexpected eslint install command: %v", cmd)
			}
			hasEslint = true
		}
	}

	if !hasTypescript {
		t.Error("Missing typescript install command")
	}
	if !hasEslint {
		t.Error("Missing eslint install command")
	}

	// already-installed should be skipped because it's not a dry-run and it's already installed.
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "already-installed") {
			t.Errorf("already-installed should have been skipped, but got: %s", cmdStr)
		}
	}
}

func TestNpmUninstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if len(args) > 2 && args[2] == "not-installed" {
			return fmt.Errorf("not installed")
		}
		return nil // installed
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	npm := &Npm{}
	pkgs := []config.Package{
		{Name: "typescript"},
		{Name: "eslint", ExtraParams: []string{"--no-fund"}},
		{Name: "not-installed"},
	}

	err := npm.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	var hasTypescript, hasEslint bool
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "typescript") {
			if cmd[0] != "npm" || cmd[1] != "uninstall" || cmd[2] != "-g" || cmd[3] != "typescript" {
				t.Errorf("Unexpected typescript uninstall command: %v", cmd)
			}
			hasTypescript = true
		} else if strings.Contains(cmdStr, "eslint") {
			if cmd[0] != "npm" || cmd[1] != "uninstall" || cmd[2] != "-g" || cmd[3] != "--no-fund" || cmd[4] != "eslint" {
				t.Errorf("Unexpected eslint uninstall command: %v", cmd)
			}
			hasEslint = true
		}
	}

	if !hasTypescript {
		t.Error("Missing typescript uninstall command")
	}
	if !hasEslint {
		t.Error("Missing eslint uninstall command")
	}

	// not-installed should be skipped because it's not a dry-run and it's not installed.
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "not-installed") {
			t.Errorf("not-installed should have been skipped, but got: %s", cmdStr)
		}
	}
}

func TestNpmUpdate(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	npm := &Npm{}
	pkgs := []config.Package{
		{Name: "typescript"},
		{Name: "eslint", ExtraParams: []string{"--no-fund"}},
	}

	err := npm.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	var hasTypescript, hasEslint bool
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "typescript") {
			if cmd[0] != "npm" || cmd[1] != "update" || cmd[2] != "-g" || cmd[3] != "typescript" {
				t.Errorf("Unexpected typescript update command: %v", cmd)
			}
			hasTypescript = true
		} else if strings.Contains(cmdStr, "eslint") {
			if cmd[0] != "npm" || cmd[1] != "update" || cmd[2] != "-g" || cmd[3] != "--no-fund" || cmd[4] != "eslint" {
				t.Errorf("Unexpected eslint update command: %v", cmd)
			}
			hasEslint = true
		}
	}

	if !hasTypescript {
		t.Error("Missing typescript update command")
	}
	if !hasEslint {
		t.Error("Missing eslint update command")
	}
}
