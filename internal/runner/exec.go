package runner

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// CommandExists checks if a binary is available in the user's PATH.
var CommandExists = func(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// DefaultExecutor runs a command synchronously. Expose as a variable to allow mocking in tests.
var DefaultExecutor = func(verbose bool, bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// DefaultShellExecutor runs a command string via /bin/sh (or cmd /C on Windows).
var DefaultShellExecutor = func(verbose bool, cmdStr string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("/bin/sh", "-c", cmdStr)
	}
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// DefaultShellCheckExecutor runs a check command silently.
var DefaultShellCheckExecutor = func(cmdStr string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("/bin/sh", "-c", cmdStr)
	}
	return cmd.Run()
}

// Run executes a command with standard stdout/stderr redirection.
func Run(verbose bool, dryRun bool, bin string, args ...string) error {
	if dryRun {
		fmt.Printf("[dry-run] %s %v\n", bin, args)
		return nil
	}

	if verbose {
		fmt.Printf("Executing: %s %v\n", bin, args)
	}

	if err := DefaultExecutor(verbose, bin, args...); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

// RunSudo executes a command prefixed with sudo if not running as root.
func RunSudo(verbose bool, dryRun bool, bin string, args ...string) error {
	if os.Geteuid() == 0 {
		return Run(verbose, dryRun, bin, args...)
	}

	sudoArgs := append([]string{bin}, args...)
	return Run(verbose, dryRun, "sudo", sudoArgs...)
}

// RunShell runs a custom script/command string.
func RunShell(verbose bool, dryRun bool, cmdStr string) error {
	if dryRun {
		shell := "/bin/sh -c"
		if runtime.GOOS == "windows" {
			shell = "cmd /C"
		}
		fmt.Printf("[dry-run] %s %q\n", shell, cmdStr)
		return nil
	}

	if verbose {
		fmt.Printf("Executing shell command: %s\n", cmdStr)
	}

	if err := DefaultShellExecutor(verbose, cmdStr); err != nil {
		return fmt.Errorf("shell command failed: %w", err)
	}
	return nil
}

// RunShellCheck runs a check command silently and returns nil only if exit code is 0.
func RunShellCheck(cmdStr string) error {
	return DefaultShellCheckExecutor(cmdStr)
}

// DefaultCheckExecutor runs a command silently. Expose as a variable to allow mocking.
var DefaultCheckExecutor = func(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	return cmd.Run()
}

// DefaultCheckOutputExecutor runs a command silently and returns stdout. Expose as a variable.
var DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
	cmd := exec.Command(bin, args...)
	return cmd.Output()
}

// RunCheck runs a command silently and returns nil only if exit code is 0.
func RunCheck(bin string, args ...string) error {
	return DefaultCheckExecutor(bin, args...)
}

// RunCheckOutput runs a command silently and returns its stdout.
func RunCheckOutput(bin string, args ...string) (string, error) {
	out, err := DefaultCheckOutputExecutor(bin, args...)
	return string(out), err
}
