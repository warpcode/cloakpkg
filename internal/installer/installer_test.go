package installer

import (
	"errors"
	"fmt"
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"testing"
)

func TestAptInstall(t *testing.T) {
	// Save originals to restore at the end
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "apt-get"
	}

	apt := &Apt{}
	if !apt.Available() {
		t.Error("Apt should be available")
	}

	pkgs := []config.Package{
		{Name: "git", ExtraParams: []string{"--no-install-recommends"}},
	}

	err := apt.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	// Since os.Geteuid() is likely not root (0), RunSudo will execute "sudo apt-get ..."
	if cmd[0] == "sudo" {
		if cmd[1] != "apt-get" || cmd[2] != "install" || cmd[3] != "-y" || cmd[4] != "--no-install-recommends" || cmd[5] != "--" || cmd[6] != "git" {
			t.Errorf("Unexpected sudo command: %v", cmd)
		}
	} else {
		if cmd[0] != "apt-get" || cmd[1] != "install" || cmd[2] != "-y" || cmd[3] != "--no-install-recommends" || cmd[4] != "--" || cmd[5] != "git" {
			t.Errorf("Unexpected non-sudo command: %v", cmd)
		}
	}
}

func TestMiseInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "mise"
	}

	mise := &Mise{}
	if !mise.Available() {
		t.Error("Mise should be available")
	}

	pkgs := []config.Package{
		{Name: "node", Version: "20", ExtraParams: []string{"--yes"}},
	}

	err := mise.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "mise" || cmd[1] != "install" || cmd[2] != "--yes" || cmd[3] != "node@20" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}

func TestCustomInstall(t *testing.T) {
	origShellExecutor := runner.DefaultShellExecutor
	origShellCheck := runner.DefaultShellCheckExecutor
	defer func() {
		runner.DefaultShellExecutor = origShellExecutor
		runner.DefaultShellCheckExecutor = origShellCheck
	}()

	var executedShellCmds []string
	runner.DefaultShellExecutor = func(verbose bool, cmdStr string) error {
		executedShellCmds = append(executedShellCmds, cmdStr)
		return nil
	}

	// Case 1: Already installed (detect returns 0)
	runner.DefaultShellCheckExecutor = func(cmdStr string) error {
		return nil // success (exit 0)
	}

	cp := config.Provider{
		Detect:     "command -v test-cmd",
		InstallCmd: "curl -fsSL test-url | sh",
	}

	err := InstallCustom(false, false, cp)
	if err != nil {
		t.Fatalf("InstallCustom failed: %v", err)
	}
	if len(executedShellCmds) != 0 {
		t.Errorf("Should not run install command if already detected, ran: %v", executedShellCmds)
	}

	// Case 2: Not installed (detect fails)
	runner.DefaultShellCheckExecutor = func(cmdStr string) error {
		return errors.New(t.Name()) // fail (non-zero exit)
	}

	err = InstallCustom(false, false, cp)
	if err != nil {
		t.Fatalf("InstallCustom failed: %v", err)
	}
	if len(executedShellCmds) != 1 || executedShellCmds[0] != "curl -fsSL test-url | sh" {
		t.Errorf("Expected install command to be executed, got: %v", executedShellCmds)
	}
}

func TestTermuxInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "pkg"
	}

	t.Setenv("TERMUX_VERSION", "0.118.0")

	termux := &Termux{}
	if !termux.Available() {
		t.Error("Termux should be available")
	}

	pkgs := []config.Package{
		{Name: "git", ExtraParams: []string{"--quiet"}},
	}

	err := termux.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	// Should run "pkg install -y --quiet git" without sudo
	if cmd[0] != "pkg" || cmd[1] != "install" || cmd[2] != "-y" || cmd[3] != "--quiet" || cmd[4] != "git" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}

func TestCargoInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "cargo"
	}

	cargo := &Cargo{}

	pkgs := []config.Package{
		{Name: "ripgrep"},
		{Name: "fd-find"},
	}

	// Mock cargo list output to show ripgrep already installed, but fd-find not
	runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		return []byte("ripgrep v13.0.0:\n    rg\n"), nil
	}

	err := cargo.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Should skip ripgrep and only execute fd-find!
	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "cargo" || cmd[1] != "install" || cmd[2] != "fd-find" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}

func TestPacmanInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
	}()

	var checkedCmds [][]string
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		checkedCmds = append(checkedCmds, append([]string{bin}, args...))
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "pacman"
	}

	pacman := &Pacman{}

	pkgs := []config.Package{
		{Name: "git", ExtraParams: []string{"--quiet"}},
	}

	err := pacman.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(checkedCmds) == 0 {
		t.Fatalf("Expected check command to be executed, but none was")
	}
	checkCmd := checkedCmds[0]
	if len(checkCmd) < 4 {
		t.Fatalf("Expected checked command to have at least 4 arguments, got %d", len(checkCmd))
	}
	if checkCmd[0] != "pacman" || checkCmd[1] != "-Qq" || checkCmd[2] != "--" || checkCmd[3] != "git" {
		t.Errorf("Unexpected check command executed: %v", checkCmd)
	}

	if len(executedCmds) == 0 {
		t.Fatalf("Expected command to be executed, but none was")
	}
	cmd := executedCmds[0]
	if len(cmd) < 7 {
		t.Fatalf("Expected executed command to have at least 7 arguments, got %d", len(cmd))
	}
	if cmd[0] != "sudo" || cmd[1] != "pacman" || cmd[2] != "-S" || cmd[3] != "--noconfirm" || cmd[4] != "--quiet" || cmd[5] != "--" || cmd[6] != "git" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}
