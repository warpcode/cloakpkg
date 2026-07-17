package installer

import (
	"fmt"
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"testing"
)

func TestGemInstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		// Mock that packages are not installed
		return fmt.Errorf("not installed")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "gem"
	}

	gem := &Gem{}
	if !gem.Available() {
		t.Error("Gem should be available")
	}

	pkgs := []config.Package{
		{Name: "bundler", ExtraParams: []string{"--no-document"}},
	}

	err := gem.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "gem" || cmd[1] != "install" || cmd[2] != "--no-document" || cmd[3] != "bundler" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}

func TestGemUninstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
	}()

	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		// Mock that packages ARE installed so they can be uninstalled
		return nil
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "gem"
	}

	gem := &Gem{}
	if !gem.Available() {
		t.Error("Gem should be available")
	}

	pkgs := []config.Package{
		{Name: "bundler", ExtraParams: []string{"--force"}},
	}

	err := gem.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "gem" || cmd[1] != "uninstall" || cmd[2] != "-a" || cmd[3] != "-x" || cmd[4] != "--force" || cmd[5] != "bundler" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}

func TestGemUpdate(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		return name == "gem"
	}

	gem := &Gem{}
	if !gem.Available() {
		t.Error("Gem should be available")
	}

	pkgs := []config.Package{
		{Name: "bundler", ExtraParams: []string{"--system"}},
	}

	err := gem.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 1 {
		t.Fatalf("Expected 1 command executed, got %d", len(executedCmds))
	}

	cmd := executedCmds[0]
	if cmd[0] != "gem" || cmd[1] != "update" || cmd[2] != "--system" || cmd[3] != "bundler" {
		t.Errorf("Unexpected command executed: %v", cmd)
	}
}


func TestGemAddRepositoriesSuccess(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	gem := &Gem{}
	repos := []config.Repository{
		{URL: "https://rubygems.org"},
		{URL: "https://custom.gem.server"},
	}

	err := gem.AddRepositories(false, false, repos)
	if err != nil {
		t.Fatalf("AddRepositories failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands executed, got %d", len(executedCmds))
	}

	cmd1 := executedCmds[0]
	if cmd1[0] != "gem" || cmd1[1] != "sources" || cmd1[2] != "-a" || cmd1[3] != "https://rubygems.org" {
		t.Errorf("Unexpected command executed: %v", cmd1)
	}

	cmd2 := executedCmds[1]
	if cmd2[0] != "gem" || cmd2[1] != "sources" || cmd2[2] != "-a" || cmd2[3] != "https://custom.gem.server" {
		t.Errorf("Unexpected command executed: %v", cmd2)
	}
}

func TestGemAddRepositoriesFailure(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		return fmt.Errorf("mock error")
	}

	gem := &Gem{}
	repos := []config.Repository{
		{URL: "https://rubygems.org"},
	}

	err := gem.AddRepositories(false, false, repos)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}
