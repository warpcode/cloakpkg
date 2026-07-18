package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"testing"
)

func TestApkName(t *testing.T) {
	apk := &Apk{}
	if apk.Name() != "apk" {
		t.Errorf("Expected Name to return 'apk', got '%s'", apk.Name())
	}
}

func TestApkAvailable(t *testing.T) {
	origExists := runner.CommandExists
	defer func() { runner.CommandExists = origExists }()

	apk := &Apk{}

	runner.CommandExists = func(name string) bool {
		return name == "apk"
	}
	if !apk.Available() {
		t.Error("Expected Apk to be available")
	}

	runner.CommandExists = func(name string) bool {
		return false
	}
	if apk.Available() {
		t.Error("Expected Apk to not be available")
	}
}

func TestApkInstalled(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	defer func() { runner.DefaultCheckExecutor = origCheck }()

	apk := &Apk{}
	pkg := config.Package{Name: "testpkg"}

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "apk" && args[0] == "info" && args[1] == "-e" && args[2] == "testpkg" {
			return nil
		}
		return fmt.Errorf("not installed")
	}
	if !apk.Installed(pkg) {
		t.Error("Expected package to be installed")
	}

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		return fmt.Errorf("not installed")
	}
	if apk.Installed(pkg) {
		t.Error("Expected package to not be installed")
	}
}

func TestApkInstall(t *testing.T) {
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
		return fmt.Errorf("mock not installed")
	}

	apk := &Apk{}
	pkgs := []config.Package{
		{Name: "pkg1", ExtraParams: []string{"--quiet"}},
		{Name: "pkg2", ExtraParams: []string{"--quiet"}},
	}

	err := apk.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] == "sudo" {
		if len(cmd) < 6 || cmd[1] != "apk" || cmd[2] != "add" || cmd[3] != "--quiet" || cmd[4] != "pkg1" || cmd[5] != "pkg2" {
			t.Errorf("Unexpected sudo command executed: %v", cmd)
		}
	} else {
		if len(cmd) < 5 || cmd[0] != "apk" || cmd[1] != "add" || cmd[2] != "--quiet" || cmd[3] != "pkg1" || cmd[4] != "pkg2" {
			t.Errorf("Unexpected non-sudo command executed: %v", cmd)
		}
	}
}

func TestApkUninstall(t *testing.T) {
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
		return nil // Mock installed
	}

	apk := &Apk{}
	pkgs := []config.Package{
		{Name: "pkg1"},
		{Name: "pkg2"},
	}

	err := apk.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] == "sudo" {
		if len(cmd) < 5 || cmd[1] != "apk" || cmd[2] != "del" || cmd[3] != "pkg1" || cmd[4] != "pkg2" {
			t.Errorf("Unexpected sudo command executed: %v", cmd)
		}
	} else {
		if len(cmd) < 4 || cmd[0] != "apk" || cmd[1] != "del" || cmd[2] != "pkg1" || cmd[3] != "pkg2" {
			t.Errorf("Unexpected non-sudo command executed: %v", cmd)
		}
	}
}

func TestApkUpdate(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	apk := &Apk{}
	pkgs := []config.Package{
		{Name: "pkg1"},
	}

	err := apk.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] == "sudo" {
		if len(cmd) < 5 || cmd[1] != "apk" || cmd[2] != "add" || cmd[3] != "--upgrade" || cmd[4] != "pkg1" {
			t.Errorf("Unexpected sudo command executed: %v", cmd)
		}
	} else {
		if len(cmd) < 4 || cmd[0] != "apk" || cmd[1] != "add" || cmd[2] != "--upgrade" || cmd[3] != "pkg1" {
			t.Errorf("Unexpected non-sudo command executed: %v", cmd)
		}
	}
}

func TestApkAddRepositories(t *testing.T) {
	apk := &Apk{}
	repos := []config.Repository{
		{URL: "https://example.com/repo"},
	}
	err := apk.AddRepositories(false, false, repos)
	if err != nil {
		t.Fatalf("Expected AddRepositories to return nil, got %v", err)
	}
}
