package cli

import (
	"cloakpkg/internal/runner"
	"os"
	"path/filepath"
	"testing"
)

type mockEnv struct {
	availableCmds []string
	checkFunc     func(bin string, args ...string) error
	checkOutput   func(bin string, args ...string) ([]byte, error)
}

func runTestConfig(t *testing.T, configJSON string, env mockEnv) [][]string {
	origExecutor := runner.DefaultExecutor
	origExists := runner.CommandExists
	origCheck := runner.DefaultCheckExecutor
	origCheckOutput := runner.DefaultCheckOutputExecutor
	defer func() {
		runner.DefaultExecutor = origExecutor
		runner.CommandExists = origExists
		runner.DefaultCheckExecutor = origCheck
		runner.DefaultCheckOutputExecutor = origCheckOutput
	}()

	var executedCmds [][]string
	runner.DefaultExecutor = func(verbose bool, bin string, args ...string) error {
		executedCmds = append(executedCmds, append([]string{bin}, args...))
		return nil
	}
	runner.CommandExists = func(name string) bool {
		for _, cmd := range env.availableCmds {
			if cmd == name {
				return true
			}
		}
		return false
	}
	if env.checkFunc != nil {
		runner.DefaultCheckExecutor = env.checkFunc
	} else {
		runner.DefaultCheckExecutor = func(bin string, args ...string) error {
			return os.ErrNotExist
		}
	}
	if env.checkOutput != nil {
		runner.DefaultCheckOutputExecutor = env.checkOutput
	} else {
		runner.DefaultCheckOutputExecutor = func(bin string, args ...string) ([]byte, error) {
			return []byte(""), nil
		}
	}

	tmpDir, err := os.MkdirTemp("", "cloakpkg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"cloakpkg", "install", configPath}

	runBundleCommand("install", configPath)

	return executedCmds
}

func stripSudo(cmd []string) []string {
	if len(cmd) > 0 && cmd[0] == "sudo" {
		return cmd[1:]
	}
	return cmd
}

func runTestFile(t *testing.T, relativePath string, env mockEnv) [][]string {
	content, err := os.ReadFile(filepath.Join("../../testdata", relativePath))
	if err != nil {
		t.Fatalf("Failed to read test config file %s: %v", relativePath, err)
	}
	return runTestConfig(t, string(content), env)
}

func findCommand(executed [][]string, bin string, args ...string) []string {
	for _, cmd := range executed {
		clean := stripSudo(cmd)
		if clean[0] != bin {
			continue
		}
		if len(args) == 0 {
			return clean
		}
		match := true
		for _, arg := range args {
			found := false
			for _, a := range clean[1:] {
				if a == arg {
					found = true
					break
				}
			}
			if !found {
				match = false
				break
			}
		}
		if match {
			return clean
		}
	}
	return nil
}
