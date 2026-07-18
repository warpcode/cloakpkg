package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"testing"
)

func TestPipxName(t *testing.T) {
	pipx := &Pipx{}
	if pipx.Name() != "pipx" {
		t.Errorf("Expected name 'pipx', got '%s'", pipx.Name())
	}
}

func TestPipxAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() {
		runner.CommandExists = origExists
	}()

	runner.CommandExists = func(name string) bool {
		return name == "pipx"
	}

	pipx := &Pipx{}
	if !pipx.Available() {
		t.Error("Pipx should be available")
	}

	runner.CommandExists = func(name string) bool {
		return false
	}
	if pipx.Available() {
		t.Error("Pipx should not be available")
	}
}

func TestPipxInstalled(t *testing.T) {
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	pipx := &Pipx{}
	pkg := config.Package{Name: "cowsay"}

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		return []byte("package cowsay"), nil
	}

	if !pipx.Installed(pkg) {
		t.Error("Package should be reported as installed")
	}

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		return []byte("package otherpkg"), nil
	}

	if pipx.Installed(pkg) {
		t.Error("Package should not be reported as installed")
	}
}

func TestPipxInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		// Mock that packages are not installed
		return []byte(""), nil
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	pipx := &Pipx{}

	pkgs := []config.Package{
		{Name: "cowsay", ExtraParams: []string{"--force"}},
	}

	err := pipx.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "pipx" || cmd[1] != "install" || cmd[2] != "--force" || cmd[3] != "cowsay" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}

func TestPipxUpdate(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	pipx := &Pipx{}

	pkgs := []config.Package{
		{Name: "cowsay"},
	}

	err := pipx.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "pipx" || cmd[1] != "upgrade" || cmd[2] != "cowsay" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}

func TestPipxAddRepositories(t *testing.T) {
	pipx := &Pipx{}
	repos := []config.Repository{
		{URL: "https://example.com/repo"},
	}

	err := pipx.AddRepositories(false, false, repos)
	if err != nil {
		t.Fatalf("AddRepositories failed: %v", err)
	}
}

func TestPipxUninstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		// Mock that packages ARE installed so they can be uninstalled
		return []byte("package cowsay"), nil
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	pipx := &Pipx{}

	pkgs := []config.Package{
		{Name: "cowsay"},
	}

	err := pipx.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "pipx" || cmd[1] != "uninstall" || cmd[2] != "cowsay" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}
