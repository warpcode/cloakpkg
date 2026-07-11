package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"strings"
)

type Uvx struct{}

func (u *Uvx) Name() string    { return "uvx" }
func (u *Uvx) Available() bool { return runner.CommandExists("uv") }
func (u *Uvx) Installed(pkg config.Package) bool {
	out, err := runner.RunCheckOutput("uv", "tool", "list")
	return err == nil && strings.Contains(out, pkg.Name+" v")
}

func (u *Uvx) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (u *Uvx) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		if !dryRun && u.Installed(pkg) {
			if verbose {
				fmt.Printf("uv: tool %s is already installed, skipping\n", pkg.Name)
			}
			continue
		}
		args := []string{"tool", "install"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "uv", args...); err != nil {
			return fmt.Errorf("uv: failed to install tool %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func (u *Uvx) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		if !dryRun && !u.Installed(pkg) {
			if verbose {
				fmt.Printf("uv: tool %s is not installed, skipping\n", pkg.Name)
			}
			continue
		}
		args := []string{"tool", "uninstall"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "uv", args...); err != nil {
			return fmt.Errorf("uv: failed to uninstall tool %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func (u *Uvx) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		args := []string{"tool", "upgrade"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "uv", args...); err != nil {
			return fmt.Errorf("uv: failed to update/upgrade tool %s: %w", pkg.Name, err)
		}
	}
	return nil
}
