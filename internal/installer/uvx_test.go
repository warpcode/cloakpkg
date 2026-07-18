package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"errors"
	"reflect"
	"testing"
)

func TestUvx_NameAndAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() {
		runner.CommandExists = origExists
	}()

	runner.CommandExists = func(name string) bool {
		return name == "uv"
	}

	uvx := &Uvx{}
	if uvx.Name() != "uvx" {
		t.Errorf("Expected name 'uvx', got '%s'", uvx.Name())
	}
	if !uvx.Available() {
		t.Error("Expected uvx to be available")
	}
}

func TestUvx_Installed(t *testing.T) {
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		if bin == "uv" && len(args) == 2 && args[0] == "tool" && args[1] == "list" {
			return []byte("pkg1 v1.0.0\npkg2 v2.0.0\n"), nil
		}
		return nil, errors.New("unexpected command")
	}

	uvx := &Uvx{}

	if !uvx.Installed(config.Package{Name: "pkg1"}) {
		t.Error("Expected pkg1 to be installed")
	}

	if uvx.Installed(config.Package{Name: "pkg3"}) {
		t.Error("Expected pkg3 not to be installed")
	}
}

func TestUvx_AddRepositories(t *testing.T) {
	uvx := &Uvx{}
	err := uvx.AddRepositories(false, false, nil)
	if err != nil {
		t.Errorf("Expected nil error from AddRepositories, got %v", err)
	}
}

func TestUvx_Install(t *testing.T) {
	origCheckOutput := runner.DefaultCheckOutputExecutor
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultCheckOutputExecutor = origCheckOutput
		runner.DefaultExecutor = origExecutor
	}()

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		if bin == "uv" && len(args) == 2 && args[0] == "tool" && args[1] == "list" {
			return []byte("installed-pkg v1.0.0\n"), nil
		}
		return nil, errors.New("unexpected command")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	uvx := &Uvx{}
	pkgs := []config.Package{
		{Name: "installed-pkg"},
		{Name: "new-pkg"},
		{Name: "with-args", ExtraParams: []string{"--with", "requests"}},
	}

	err := uvx.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	expectedCmd1 := []string{"uv", "tool", "install", "new-pkg"}
	if !reflect.DeepEqual(executedCmds[0], expectedCmd1) {
		t.Errorf("Expected command %v, got %v", expectedCmd1, executedCmds[0])
	}

	expectedCmd2 := []string{"uv", "tool", "install", "--with", "requests", "with-args"}
	if !reflect.DeepEqual(executedCmds[1], expectedCmd2) {
		t.Errorf("Expected command %v, got %v", expectedCmd2, executedCmds[1])
	}
}

func TestUvx_Uninstall(t *testing.T) {
	origCheckOutput := runner.DefaultCheckOutputExecutor
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultCheckOutputExecutor = origCheckOutput
		runner.DefaultExecutor = origExecutor
	}()

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		if bin == "uv" && len(args) == 2 && args[0] == "tool" && args[1] == "list" {
			return []byte("installed-pkg v1.0.0\n"), nil
		}
		return nil, errors.New("unexpected command")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	uvx := &Uvx{}
	pkgs := []config.Package{
		{Name: "not-installed-pkg"},
		{Name: "installed-pkg"},
		{Name: "installed-pkg", ExtraParams: []string{"--force"}},
	}

	err := uvx.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	expectedCmd1 := []string{"uv", "tool", "uninstall", "installed-pkg"}
	if !reflect.DeepEqual(executedCmds[0], expectedCmd1) {
		t.Errorf("Expected command %v, got %v", expectedCmd1, executedCmds[0])
	}

	expectedCmd2 := []string{"uv", "tool", "uninstall", "--force", "installed-pkg"}
	if !reflect.DeepEqual(executedCmds[1], expectedCmd2) {
		t.Errorf("Expected command %v, got %v", expectedCmd2, executedCmds[1])
	}
}

func TestUvx_Update(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	uvx := &Uvx{}
	pkgs := []config.Package{
		{Name: "pkg1"},
		{Name: "pkg2", ExtraParams: []string{"--all"}},
	}

	err := uvx.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	expectedCmd1 := []string{"uv", "tool", "upgrade", "pkg1"}
	if !reflect.DeepEqual(executedCmds[0], expectedCmd1) {
		t.Errorf("Expected command %v, got %v", expectedCmd1, executedCmds[0])
	}

	expectedCmd2 := []string{"uv", "tool", "upgrade", "--all", "pkg2"}
	if !reflect.DeepEqual(executedCmds[1], expectedCmd2) {
		t.Errorf("Expected command %v, got %v", expectedCmd2, executedCmds[1])
	}
}
