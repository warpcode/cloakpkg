package installer

import (
	"fmt"
	"strings"
	"testing"

	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
)

func TestCargoName(t *testing.T) {
	c := &Cargo{}
	if name := c.Name(); name != "cargo" {
		t.Errorf("Expected name 'cargo', got '%s'", name)
	}
}

func TestCargoAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() { runner.CommandExists = origExists }()

	runner.CommandExists = func(name string) bool {
		return name == "cargo"
	}

	c := &Cargo{}
	if !c.Available() {
		t.Error("Cargo should be available when command exists")
	}

	runner.CommandExists = func(name string) bool {
		return false
	}
	if c.Available() {
		t.Error("Cargo should not be available when command doesn't exist")
	}
}

func TestCargoInstalled(t *testing.T) {
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() { runner.DefaultCheckOutputExecutor = origCheckOutput }()

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		if bin == "cargo" && len(args) == 2 && args[0] == "install" && args[1] == "--list" {
			output := "bat v0.18.3:\n    bat\nripgrep v13.0.0:\n    rg\n"
			return []byte(output), nil
		}
		return nil, fmt.Errorf("command failed")
	}

	c := &Cargo{}
	pkgInstalled := config.Package{Name: "bat"}
	pkgNotInstalled := config.Package{Name: "fd-find"}

	if !c.Installed(pkgInstalled) {
		t.Error("bat should be installed")
	}
	if c.Installed(pkgNotInstalled) {
		t.Error("fd-find should not be installed")
	}
}

func TestCargoAddRepositories(t *testing.T) {
	c := &Cargo{}
	err := c.AddRepositories(false, false, []config.Repository{})
	if err != nil {
		t.Errorf("AddRepositories returned error: %v", err)
	}
}

func TestCargoInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		if bin == "cargo" && len(args) == 2 && args[0] == "install" && args[1] == "--list" {
			output := "already-installed v1.0.0:\n    already-installed\n"
			return []byte(output), nil
		}
		return nil, fmt.Errorf("command failed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	c := &Cargo{}
	pkgs := []config.Package{
		{Name: "bat"},
		{Name: "ripgrep", ExtraParams: []string{"--locked"}},
		{Name: "already-installed"},
	}

	err := c.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	var hasBat, hasRipgrep bool
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "bat") {
			if cmd[0] != "cargo" || cmd[1] != "install" || cmd[2] != "bat" {
				t.Errorf("Unexpected bat install command: %v", cmd)
			}
			hasBat = true
		} else if strings.Contains(cmdStr, "ripgrep") {
			if cmd[0] != "cargo" || cmd[1] != "install" || cmd[2] != "--locked" || cmd[3] != "ripgrep" {
				t.Errorf("Unexpected ripgrep install command: %v", cmd)
			}
			hasRipgrep = true
		}
	}

	if !hasBat {
		t.Error("Missing bat install command")
	}
	if !hasRipgrep {
		t.Error("Missing ripgrep install command")
	}

	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "already-installed") {
			t.Errorf("already-installed should have been skipped, but got: %s", cmdStr)
		}
	}
}

func TestCargoUninstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		if bin == "cargo" && len(args) == 2 && args[0] == "install" && args[1] == "--list" {
			output := "bat v0.18.3:\n    bat\nripgrep v13.0.0:\n    rg\n"
			return []byte(output), nil
		}
		return nil, fmt.Errorf("command failed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	c := &Cargo{}
	pkgs := []config.Package{
		{Name: "bat"},
		{Name: "ripgrep", ExtraParams: []string{"--locked"}},
		{Name: "not-installed"},
	}

	err := c.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	var hasBat, hasRipgrep bool
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "bat") {
			if cmd[0] != "cargo" || cmd[1] != "uninstall" || cmd[2] != "bat" {
				t.Errorf("Unexpected bat uninstall command: %v", cmd)
			}
			hasBat = true
		} else if strings.Contains(cmdStr, "ripgrep") {
			if cmd[0] != "cargo" || cmd[1] != "uninstall" || cmd[2] != "--locked" || cmd[3] != "ripgrep" {
				t.Errorf("Unexpected ripgrep uninstall command: %v", cmd)
			}
			hasRipgrep = true
		}
	}

	if !hasBat {
		t.Error("Missing bat uninstall command")
	}
	if !hasRipgrep {
		t.Error("Missing ripgrep uninstall command")
	}

	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "not-installed") {
			t.Errorf("not-installed should have been skipped, but got: %s", cmdStr)
		}
	}
}

func TestCargoUpdate(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	c := &Cargo{}
	pkgs := []config.Package{
		{Name: "bat"},
		{Name: "ripgrep", ExtraParams: []string{"--locked"}},
	}

	err := c.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	var hasBat, hasRipgrep bool
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "bat") {
			if cmd[0] != "cargo" || cmd[1] != "install" || cmd[2] != "--force" || cmd[3] != "bat" {
				t.Errorf("Unexpected bat update command: %v", cmd)
			}
			hasBat = true
		} else if strings.Contains(cmdStr, "ripgrep") {
			if cmd[0] != "cargo" || cmd[1] != "install" || cmd[2] != "--force" || cmd[3] != "--locked" || cmd[4] != "ripgrep" {
				t.Errorf("Unexpected ripgrep update command: %v", cmd)
			}
			hasRipgrep = true
		}
	}

	if !hasBat {
		t.Error("Missing bat update command")
	}
	if !hasRipgrep {
		t.Error("Missing ripgrep update command")
	}
}
