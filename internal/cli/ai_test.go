package cli

import (
	"testing"
)

func TestAiPackages(t *testing.T) {
	// Test Case: Mise and Npm available
	t.Run("MiseNpm", func(t *testing.T) {
		executed := runTestFile(t, "ai.json", mockEnv{
			availableCmds: []string{"mise", "npm"},
		})
		// We expect 2 commands:
		// 1. mise: ollama
		// 2. npm: skills@latest, @github/copilot, @google/gemini-cli, opencode-ai@latest, @kilocode/cli, @mariozechner/pi-coding-agent, @qwen-code/qwen-code@latest, @openai/codex, @anthropic-ai/claude-code
		if len(executed) != 2 {
			t.Fatalf("Expected 2 commands executed, got %d: %v", len(executed), executed)
		}

		var miseCmd, npmCmd []string
		for _, cmd := range executed {
			if cmd[0] == "mise" {
				miseCmd = cmd
			} else if cmd[0] == "npm" {
				npmCmd = cmd
			}
		}

		if len(miseCmd) == 0 {
			t.Errorf("Missing mise command")
		} else {
			lastArg := miseCmd[len(miseCmd)-1]
			if lastArg != "ollama" {
				t.Errorf("Expected mise to install ollama, got: %v", miseCmd)
			}
		}

		if len(npmCmd) == 0 {
			t.Errorf("Missing npm command")
		} else {
			expected := map[string]bool{
				"skills@latest":                 true,
				"@github/copilot":               true,
				"@google/gemini-cli":            true,
				"opencode-ai@latest":            true,
				"@kilocode/cli":                 true,
				"@mariozechner/pi-coding-agent": true,
				"@qwen-code/qwen-code@latest":   true,
				"@openai/codex":                 true,
				"@anthropic-ai/claude-code":     true,
			}
			for _, arg := range npmCmd[2:] {
				delete(expected, arg)
			}
			if len(expected) > 0 {
				t.Errorf("Npm command missing AI packages: %v", expected)
			}
		}
	})
}
