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
	origExecutor := DefaultExecutor
	defer func() { DefaultExecutor = origExecutor }()

	var lastBin string
	var lastArgs []string

	DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		lastBin = bin
		lastArgs = args
		return nil
	}

	// In test environments, we might be root or not root, and we might be on Windows or not.
	// Since RunSudo checks OS and uid to decide if it should prepend sudo,
	// we will just execute it and check if it did prepend sudo or not.
	// We can't strictly enforce one outcome without mocking os.Geteuid, which isn't mocked in exec.go.
	// But we can check that it either executes the original command or prepends sudo.
	err := RunSudo(false, false, "echo", "hello")
	if err != nil {
		t.Errorf("expected RunSudo to succeed, got %v", err)
	}

	// It either ran "echo hello" or "sudo echo hello"
	if lastBin != "echo" && lastBin != "sudo" {
		t.Errorf("expected bin to be 'echo' or 'sudo', got %v", lastBin)
	}

	if lastBin == "sudo" {
		if len(lastArgs) != 2 || lastArgs[0] != "echo" || lastArgs[1] != "hello" {
			t.Errorf("expected sudo args to be ['echo', 'hello'], got %v", lastArgs)
		}
	} else if lastBin == "echo" {
		if len(lastArgs) != 1 || lastArgs[0] != "hello" {
			t.Errorf("expected args to be ['hello'], got %v", lastArgs)
		}
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
