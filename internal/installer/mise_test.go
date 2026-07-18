package installer

import (
	"fmt"
	"strings"
	"testing"

	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
)

func TestMiseName(t *testing.T) {
	mise := &Mise{}
	if name := mise.Name(); name != "mise" {
		t.Errorf("Expected name 'mise', got '%s'", name)
	}
}

func TestMiseAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() { runner.CommandExists = origExists }()

	runner.CommandExists = func(name string) bool {
		return name == "mise"
	}

	mise := &Mise{}
	if !mise.Available() {
		t.Error("Mise should be available when command exists")
	}

	runner.CommandExists = func(name string) bool {
		return false
	}
	if mise.Available() {
		t.Error("Mise should not be available when command doesn't exist")
	}
}

func TestMiseInstalled(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	defer func() { runner.DefaultCheckExecutor = origCheck }()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "mise" && len(args) == 2 && args[0] == "ls" && args[1] == "node@20" {
			return nil
		}
		if bin == "mise" && len(args) == 2 && args[0] == "ls" && args[1] == "python" {
			return nil
		}
		return fmt.Errorf("not installed")
	}

	mise := &Mise{}
	pkgInstalledWithVersion := config.Package{Name: "node", Version: "20"}
	pkgInstalledNoVersion := config.Package{Name: "python"}
	pkgNotInstalled := config.Package{Name: "ruby"}

	if !mise.Installed(pkgInstalledWithVersion) {
		t.Error("node@20 should be installed")
	}
	if !mise.Installed(pkgInstalledNoVersion) {
		t.Error("python should be installed")
	}
	if mise.Installed(pkgNotInstalled) {
		t.Error("ruby should not be installed")
	}
}

func TestMiseAddRepositories(t *testing.T) {
	mise := &Mise{}
	repos := []config.Repository{
		{URL: "https://example.com"},
	}

	err := mise.AddRepositories(false, false, repos)
	if err != nil {
		t.Errorf("AddRepositories should always return nil, got: %v", err)
	}
}

func TestMiseInstallGrouping(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if len(args) == 2 && args[0] == "ls" && args[1] == "already-installed" {
			return nil // installed
		}
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	mise := &Mise{}
	pkgs := []config.Package{
		{Name: "node", Version: "20"},
		{Name: "python", ExtraParams: []string{"--yes"}},
		{Name: "already-installed"},
	}

	err := mise.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	var hasNode, hasPython bool
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "node@20") {
			if cmd[0] != "mise" || cmd[1] != "install" || cmd[2] != "node@20" {
				t.Errorf("Unexpected node install command: %v", cmd)
			}
			hasNode = true
		} else if strings.Contains(cmdStr, "python") {
			if cmd[0] != "mise" || cmd[1] != "install" || cmd[2] != "--yes" || cmd[3] != "python" {
				t.Errorf("Unexpected python install command: %v", cmd)
			}
			hasPython = true
		}
	}

	if !hasNode {
		t.Error("Missing node install command")
	}
	if !hasPython {
		t.Error("Missing python install command")
	}

	// already-installed should be skipped because it's not a dry-run and it's already installed.
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "already-installed") {
			t.Errorf("already-installed should have been skipped, but got: %s", cmdStr)
		}
	}
}

func TestMiseUninstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if len(args) == 2 && args[0] == "ls" && args[1] == "not-installed" {
			return fmt.Errorf("not installed")
		}
		return nil // installed
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	mise := &Mise{}
	pkgs := []config.Package{
		{Name: "node", Version: "20"},
		{Name: "python", ExtraParams: []string{"--yes"}},
		{Name: "not-installed"},
	}

	err := mise.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	var hasNode, hasPython bool
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "node@20") {
			if cmd[0] != "mise" || cmd[1] != "uninstall" || cmd[2] != "node@20" {
				t.Errorf("Unexpected node uninstall command: %v", cmd)
			}
			hasNode = true
		} else if strings.Contains(cmdStr, "python") {
			if cmd[0] != "mise" || cmd[1] != "uninstall" || cmd[2] != "--yes" || cmd[3] != "python" {
				t.Errorf("Unexpected python uninstall command: %v", cmd)
			}
			hasPython = true
		}
	}

	if !hasNode {
		t.Error("Missing node uninstall command")
	}
	if !hasPython {
		t.Error("Missing python uninstall command")
	}

	// not-installed should be skipped because it's not a dry-run and it's not installed.
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "not-installed") {
			t.Errorf("not-installed should have been skipped, but got: %s", cmdStr)
		}
	}
}

func TestMiseUpdate(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	mise := &Mise{}
	pkgs := []config.Package{
		{Name: "node", Version: "20"},
		{Name: "python", ExtraParams: []string{"--yes"}},
	}

	err := mise.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	var hasNode, hasPython bool
	for _, cmd := range executedCmds {
		cmdStr := strings.Join(cmd, " ")
		if strings.Contains(cmdStr, "node@20") {
			if cmd[0] != "mise" || cmd[1] != "upgrade" || cmd[2] != "node@20" {
				t.Errorf("Unexpected node update command: %v", cmd)
			}
			hasNode = true
		} else if strings.Contains(cmdStr, "python") {
			if cmd[0] != "mise" || cmd[1] != "upgrade" || cmd[2] != "--yes" || cmd[3] != "python" {
				t.Errorf("Unexpected python update command: %v", cmd)
			}
			hasPython = true
		}
	}

	if !hasNode {
		t.Error("Missing node update command")
	}
	if !hasPython {
		t.Error("Missing python update command")
	}
}
