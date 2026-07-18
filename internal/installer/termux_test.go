package installer

import (
	"errors"
	"reflect"
	"testing"

	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
)

func TestTermux_Name(t *testing.T) {
	termux := &Termux{}
	if name := termux.Name(); name != "termux" {
		t.Errorf("Expected 'termux', got '%s'", name)
	}
}

func TestTermux_Available(t *testing.T) {
	origExists := runner.CommandExists
	defer func() {
		runner.CommandExists = origExists
	}()

	termux := &Termux{}

	tests := []struct {
		name          string
		commandExists bool
		termuxVersion string
		prefix        string
		expected      bool
	}{
		{"pkg not exists", false, "", "", false},
		{"pkg exists, TERMUX_VERSION set", true, "0.118", "", true},
		{"pkg exists, PREFIX com.termux", true, "", "/data/data/com.termux/files/usr", true},
		{"pkg exists, none set", true, "", "/usr/local", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner.CommandExists = func(name string) bool {
				return tt.commandExists && name == "pkg"
			}

			if tt.termuxVersion != "" {
				t.Setenv("TERMUX_VERSION", tt.termuxVersion)
			} else {
				t.Setenv("TERMUX_VERSION", "")
			}

			if tt.prefix != "" {
				t.Setenv("PREFIX", tt.prefix)
			} else {
				t.Setenv("PREFIX", "")
			}

			if got := termux.Available(); got != tt.expected {
				t.Errorf("Expected Available() to be %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestTermux_Installed(t *testing.T) {
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultCheckExecutor = origCheck
	}()

	termux := &Termux{}

	tests := []struct {
		name       string
		pkg        config.Package
		checkError error
		expected   bool
	}{
		{"installed", config.Package{Name: "git"}, nil, true},
		{"not installed", config.Package{Name: "git"}, errors.New("not installed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var executedBin string
			var executedArgs []string
			runner.DefaultCheckExecutor = func(bin string, args ...string) error {
				executedBin = bin
				executedArgs = args
				return tt.checkError
			}

			got := termux.Installed(tt.pkg)
			if got != tt.expected {
				t.Errorf("Expected Installed() to be %v, got %v", tt.expected, got)
			}

			if executedBin != "dpkg" {
				t.Errorf("Expected bin 'dpkg', got '%s'", executedBin)
			}

			expectedArgs := []string{"-s", tt.pkg.Name}
			if !reflect.DeepEqual(executedArgs, expectedArgs) {
				t.Errorf("Expected args %v, got %v", expectedArgs, executedArgs)
			}
		})
	}
}

func TestTermux_AddRepositories(t *testing.T) {
	termux := &Termux{}
	err := termux.AddRepositories(false, false, []config.Repository{})
	if err != nil {
		t.Errorf("Expected AddRepositories to return nil, got %v", err)
	}
}

func TestTermux_Install(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	termux := &Termux{}

	// Mock `dpkg -s <pkg>` (checking if installed)
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "dpkg" && len(args) == 2 && args[0] == "-s" {
			if args[1] == "installed-pkg" {
				return nil // Installed
			}
			return errors.New("not installed")
		}
		return errors.New("unexpected command")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	pkgs := []config.Package{
		{Name: "installed-pkg"},
		{Name: "new-pkg"},
		{Name: "quiet-pkg", ExtraParams: []string{"--quiet"}},
	}

	err := termux.Install(false, false, pkgs)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundNewPkg := false
	foundQuietPkg := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, []string{"pkg", "install", "-y", "new-pkg"}) {
			foundNewPkg = true
		} else if reflect.DeepEqual(cmd, []string{"pkg", "install", "-y", "--quiet", "quiet-pkg"}) {
			foundQuietPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundNewPkg || !foundQuietPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}

func TestTermux_Uninstall(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	origCheck := runner.DefaultCheckExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.DefaultCheckExecutor = origCheck
	}()

	termux := &Termux{}

	// Mock `dpkg -s <pkg>` (checking if installed)
	runner.DefaultCheckExecutor = func(bin string, args ...string) error {
		if bin == "dpkg" && len(args) == 2 && args[0] == "-s" {
			if args[1] == "installed-pkg" || args[1] == "quiet-pkg" {
				return nil // Installed
			}
			return errors.New("not installed")
		}
		return errors.New("unexpected command")
	}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	pkgs := []config.Package{
		{Name: "installed-pkg"},
		{Name: "not-installed-pkg"},
		{Name: "quiet-pkg", ExtraParams: []string{"--quiet"}},
	}

	err := termux.Uninstall(false, false, pkgs)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundInstalledPkg := false
	foundQuietPkg := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, []string{"pkg", "uninstall", "-y", "installed-pkg"}) {
			foundInstalledPkg = true
		} else if reflect.DeepEqual(cmd, []string{"pkg", "uninstall", "-y", "--quiet", "quiet-pkg"}) {
			foundQuietPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundInstalledPkg || !foundQuietPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}

func TestTermux_Update(t *testing.T) {
	origExecutor := runner.DefaultExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
	}()

	termux := &Termux{}

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}

	pkgs := []config.Package{
		{Name: "pkg1"},
		{Name: "pkg2"},
		{Name: "quiet-pkg", ExtraParams: []string{"--quiet"}},
	}

	err := termux.Update(false, false, pkgs)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Fatalf("Expected 2 commands to be executed, got %d", len(executedCmds))
	}

	foundPkgs := false
	foundQuietPkg := false

	for _, cmd := range executedCmds {
		if reflect.DeepEqual(cmd, []string{"pkg", "install", "-y", "pkg1", "pkg2"}) || reflect.DeepEqual(cmd, []string{"pkg", "install", "-y", "pkg2", "pkg1"}) {
			foundPkgs = true
		} else if reflect.DeepEqual(cmd, []string{"pkg", "install", "-y", "--quiet", "quiet-pkg"}) {
			foundQuietPkg = true
		} else {
			t.Errorf("Unexpected command executed: %v", cmd)
		}
	}

	if !foundPkgs || !foundQuietPkg {
		t.Errorf("Missing expected commands. Executed: %v", executedCmds)
	}
}
