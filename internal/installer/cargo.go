package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"strings"
)

type Cargo struct{}

func (c *Cargo) Name() string    { return "cargo" }
func (c *Cargo) Available() bool { return runner.CommandExists("cargo") }
func (c *Cargo) Installed(pkg config.Package) bool {
	out, err := runner.RunCheckOutput("cargo", "install", "--list")
	return err == nil && strings.Contains(out, pkg.Name+" v")
}

func (c *Cargo) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (c *Cargo) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		if !dryRun && c.Installed(pkg) {
			if verbose {
				fmt.Printf("cargo: package %s is already installed, skipping\n", pkg.Name)
			}
			continue
		}
		args := []string{"install"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "cargo", args...); err != nil {
			return fmt.Errorf("cargo: failed to install %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func (c *Cargo) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		if !dryRun && !c.Installed(pkg) {
			if verbose {
				fmt.Printf("cargo: package %s is not installed, skipping\n", pkg.Name)
			}
			continue
		}
		args := []string{"uninstall"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "cargo", args...); err != nil {
			return fmt.Errorf("cargo: failed to uninstall %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func (c *Cargo) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		args := []string{"install", "--force"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "cargo", args...); err != nil {
			return fmt.Errorf("cargo: failed to update %s: %w", pkg.Name, err)
		}
	}
	return nil
}
