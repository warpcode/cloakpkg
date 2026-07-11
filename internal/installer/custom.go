package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
)

// CheckCustom detects if a custom provider is already installed by running its detect script.
// Returns true if installed (detect exit code is 0), false otherwise.
func CheckCustom(cp config.Provider) bool {
	if cp.Detect == "" {
		return false
	}
	err := runner.RunShellCheck(cp.Detect)
	return err == nil
}

// InstallCustom runs the install script of a custom provider if its detect script fails.
func InstallCustom(verbose bool, dryRun bool, cp config.Provider) error {
	if cp.Detect != "" {
		if err := runner.RunShellCheck(cp.Detect); err == nil {
			if verbose {
				fmt.Printf("Custom provider already installed (detect passed: %q)\n", cp.Detect)
			}
			return nil // Already installed, skip
		}
	}

	if cp.InstallCmd == "" {
		return fmt.Errorf("custom provider missing install command")
	}

	if err := runner.RunShell(verbose, dryRun, cp.InstallCmd); err != nil {
		return fmt.Errorf("custom install failed: %w", err)
	}
	return nil
}

// UninstallCustom runs the uninstall script of a custom provider.
func UninstallCustom(verbose bool, dryRun bool, cp config.Provider) error {
	if cp.Uninstall == "" {
		if verbose {
			fmt.Println("Custom provider missing uninstall command, skipping")
		}
		return nil
	}

	if err := runner.RunShell(verbose, dryRun, cp.Uninstall); err != nil {
		return fmt.Errorf("custom uninstall failed: %w", err)
	}
	return nil
}

// UpdateCustom runs the update script of a custom provider if defined.
func UpdateCustom(verbose bool, dryRun bool, cp config.Provider) error {
	if cp.Update == "" {
		if verbose {
			fmt.Println("Custom provider missing update command, falling back to install logic")
		}
		return InstallCustom(verbose, dryRun, cp)
	}

	if err := runner.RunShell(verbose, dryRun, cp.Update); err != nil {
		return fmt.Errorf("custom update failed: %w", err)
	}
	return nil
}
