package runner

import (
	"runtime"
	"testing"
)

func TestCommandExists(t *testing.T) {
	// "go" should always exist in our testing environment
	if !CommandExists("go") {
		t.Errorf("expected 'go' to exist in PATH")
	}

	// a non-existent command
	if CommandExists("this-command-should-not-exist-12345") {
		t.Errorf("expected 'this-command-should-not-exist-12345' to NOT exist in PATH")
	}
}

func TestDefaultExecutor(t *testing.T) {
	// Test a simple command that should succeed
	err := DefaultExecutor(false, "echo", "hello")
	if err != nil {
		t.Errorf("expected DefaultExecutor to succeed, got %v", err)
	}

	// Test with verbose = true
	err = DefaultExecutor(true, "echo", "hello")
	if err != nil {
		t.Errorf("expected DefaultExecutor verbose to succeed, got %v", err)
	}

	// Test a command that should fail
	err = DefaultExecutor(false, "this-command-should-not-exist-12345")
	if err == nil {
		t.Errorf("expected DefaultExecutor to fail for non-existent command")
	}
}

func TestDefaultShellExecutor(t *testing.T) {
	cmdStr := "echo hello"
	if runtime.GOOS == "windows" {
		cmdStr = "echo hello"
	}

	err := DefaultShellExecutor(false, cmdStr)
	if err != nil {
		t.Errorf("expected DefaultShellExecutor to succeed, got %v", err)
	}

	err = DefaultShellExecutor(true, cmdStr)
	if err != nil {
		t.Errorf("expected DefaultShellExecutor verbose to succeed, got %v", err)
	}
}

func TestDefaultShellCheckExecutor(t *testing.T) {
	cmdStr := "echo hello"
	if runtime.GOOS == "windows" {
		cmdStr = "echo hello"
	}

	err := DefaultShellCheckExecutor(cmdStr)
	if err != nil {
		t.Errorf("expected DefaultShellCheckExecutor to succeed, got %v", err)
	}
}

func TestRun(t *testing.T) {
	// Save the original executor and restore it after the test
	origExecutor := DefaultExecutor
	defer func() { DefaultExecutor = origExecutor }()

	var lastBin string
	var lastArgs []string
	var lastVerbose bool

	// Mock the executor
	DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		lastVerbose = verbose
		lastBin = bin
		lastArgs = args
		return nil
	}

	// Test dryRun = false
	err := Run(true, false, "echo", "hello", "world")
	if err != nil {
		t.Errorf("expected Run to succeed, got %v", err)
	}
	if !lastVerbose {
		t.Errorf("expected verbose to be true")
	}
	if lastBin != "echo" {
		t.Errorf("expected bin to be 'echo', got %v", lastBin)
	}
	if len(lastArgs) != 2 || lastArgs[0] != "hello" || lastArgs[1] != "world" {
		t.Errorf("expected args to be ['hello', 'world'], got %v", lastArgs)
	}

	// Test dryRun = true
	lastBin = "" // Reset to verify it doesn't get called
	err = Run(true, true, "echo", "hello", "world")
	if err != nil {
		t.Errorf("expected Run to succeed, got %v", err)
	}
	if lastBin != "" {
		t.Errorf("expected DefaultExecutor to not be called on dryRun, but it was")
	}
}

func TestRunSudo(t *testing.T) {
	// Save and restore global variables
	originalExecutor := DefaultExecutor
	originalGeteuid := osGeteuid
	defer func() {
		DefaultExecutor = originalExecutor
		osGeteuid = originalGeteuid
	}()

	tests := []struct {
		name         string
		geteuid      int
		dryRun       bool
		verbose      bool
		bin          string
		args         []string
		wantExecuted bool
		wantBin      string
		wantArgs     []string
	}{
		{
			name:         "non-root user prepends sudo on non-windows",
			geteuid:      1000,
			dryRun:       false,
			verbose:      false,
			bin:          "mycmd",
			args:         []string{"arg1", "arg2"},
			wantExecuted: true,
			wantBin:      "sudo",
			wantArgs:     []string{"mycmd", "arg1", "arg2"},
		},
		{
			name:         "root user runs directly",
			geteuid:      0,
			dryRun:       false,
			verbose:      false,
			bin:          "mycmd",
			args:         []string{"arg1", "arg2"},
			wantExecuted: true,
			wantBin:      "mycmd",
			wantArgs:     []string{"arg1", "arg2"},
		},
		{
			name:         "dryRun mode does not execute",
			geteuid:      1000,
			dryRun:       true,
			verbose:      false,
			bin:          "mycmd",
			args:         []string{"arg1"},
			wantExecuted: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var executedBin string
			var executedArgs []string
			executed := false

			DefaultExecutor = func(verbose bool, bin string, args ...string) error {
				executed = true
				executedBin = bin
				executedArgs = append([]string{}, args...)
				return nil
			}

			osGeteuid = func() int { return tc.geteuid }

			err := RunSudo(tc.verbose, tc.dryRun, tc.bin, tc.args...)
			if err != nil {
				t.Fatalf("RunSudo(%v, %v, %q, %v) unexpected error: %v", tc.verbose, tc.dryRun, tc.bin, tc.args, err)
			}

			if executed != tc.wantExecuted {
				t.Errorf("RunSudo(%v, %v, %q, %v) executed = %v, want %v", tc.verbose, tc.dryRun, tc.bin, tc.args, executed, tc.wantExecuted)
			}

			if !executed {
				return
			}

			// Adjust expectations for Windows
			expectedBin := tc.wantBin
			expectedArgs := tc.wantArgs
			if runtime.GOOS == "windows" {
				expectedBin = tc.bin
				expectedArgs = tc.args
			}

			if executedBin != expectedBin {
				t.Errorf("RunSudo(%v, %v, %q, %v) bin = %q, want %q", tc.verbose, tc.dryRun, tc.bin, tc.args, executedBin, expectedBin)
			}

			if len(executedArgs) != len(expectedArgs) {
				t.Errorf("RunSudo(%v, %v, %q, %v) args length = %d, want %d (got %v, want %v)", tc.verbose, tc.dryRun, tc.bin, tc.args, len(executedArgs), len(expectedArgs), executedArgs, expectedArgs)
				return
			}

			for i := range executedArgs {
				if executedArgs[i] != expectedArgs[i] {
					t.Errorf("RunSudo(%v, %v, %q, %v) args[%d] = %q, want %q (got %v, want %v)", tc.verbose, tc.dryRun, tc.bin, tc.args, i, executedArgs[i], expectedArgs[i], executedArgs, expectedArgs)
				}
			}
		})
	}
}

func TestRunShell(t *testing.T) {
	origShellExecutor := DefaultShellExecutor
	defer func() { DefaultShellExecutor = origShellExecutor }()

	var lastCmdStr string
	var lastVerbose bool

	DefaultShellExecutor = func(verbose bool, cmdStr string) error {
		lastVerbose = verbose
		lastCmdStr = cmdStr
		return nil
	}

	// Test dryRun = false
	err := RunShell(true, false, "echo hello")
	if err != nil {
		t.Errorf("expected RunShell to succeed, got %v", err)
	}
	if !lastVerbose {
		t.Errorf("expected verbose to be true")
	}
	if lastCmdStr != "echo hello" {
		t.Errorf("expected cmdStr to be 'echo hello', got %v", lastCmdStr)
	}

	// Test dryRun = true
	lastCmdStr = "" // Reset to verify it doesn't get called
	err = RunShell(true, true, "echo hello")
	if err != nil {
		t.Errorf("expected RunShell to succeed, got %v", err)
	}
	if lastCmdStr != "" {
		t.Errorf("expected DefaultShellExecutor to not be called on dryRun, but it was")
	}
}

func TestRunShellCheck(t *testing.T) {
	origShellCheckExecutor := DefaultShellCheckExecutor
	defer func() { DefaultShellCheckExecutor = origShellCheckExecutor }()

	var lastCmdStr string

	DefaultShellCheckExecutor = func(cmdStr string) error {
		lastCmdStr = cmdStr
		return nil
	}

	err := RunShellCheck("echo hello")
	if err != nil {
		t.Errorf("expected RunShellCheck to succeed, got %v", err)
	}
	if lastCmdStr != "echo hello" {
		t.Errorf("expected cmdStr to be 'echo hello', got %v", lastCmdStr)
	}
}

func TestRunCheck(t *testing.T) {
	origCheckExecutor := DefaultCheckExecutor
	defer func() { DefaultCheckExecutor = origCheckExecutor }()

	var lastBin string
	var lastArgs []string

	DefaultCheckExecutor = func(bin string, args ...string) error {
		lastBin = bin
		lastArgs = args
		return nil
	}

	err := RunCheck("echo", "hello")
	if err != nil {
		t.Errorf("expected RunCheck to succeed, got %v", err)
	}
	if lastBin != "echo" {
		t.Errorf("expected bin to be 'echo', got %v", lastBin)
	}
	if len(lastArgs) != 1 || lastArgs[0] != "hello" {
		t.Errorf("expected args to be ['hello'], got %v", lastArgs)
	}
}

func TestRunCheckOutput(t *testing.T) {
	origCheckOutputExecutor := DefaultCheckOutputExecutor
	defer func() { DefaultCheckOutputExecutor = origCheckOutputExecutor }()

	var lastBin string
	var lastArgs []string

	DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
		lastBin = bin
		lastArgs = args
		return []byte("output"), nil
	}

	out, err := RunCheckOutput("echo", "hello")
	if err != nil {
		t.Errorf("expected RunCheckOutput to succeed, got %v", err)
	}
	if out != "output" {
		t.Errorf("expected output to be 'output', got %v", out)
	}
	if lastBin != "echo" {
		t.Errorf("expected bin to be 'echo', got %v", lastBin)
	}
	if len(lastArgs) != 1 || lastArgs[0] != "hello" {
		t.Errorf("expected args to be ['hello'], got %v", lastArgs)
	}
}

func TestExecuteCommand(t *testing.T) {
	origExecutor := DefaultExecutor
	defer func() { DefaultExecutor = origExecutor }()

	var lastBin string
	var lastArgs []string

	DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		lastBin = bin
		lastArgs = args
		return nil
	}

	err := ExecuteCommand(false, false, "echo", []string{"hello"}, nil)
	if err != nil {
		t.Errorf("expected ExecuteCommand to succeed, got %v", err)
	}
	if lastBin != "echo" {
		t.Errorf("expected bin to be 'echo', got %v", lastBin)
	}
	if len(lastArgs) != 1 || lastArgs[0] != "hello" {
		t.Errorf("expected args to be ['hello'], got %v", lastArgs)
	}
}
