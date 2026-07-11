package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"strings"
)

type Pipx struct{}

func (p *Pipx) Name() string    { return "pipx" }
func (p *Pipx) Available() bool { return runner.CommandExists("pipx") }
func (p *Pipx) Installed(pkg config.Package) bool {
	out, err := runner.RunCheckOutput("pipx", "list")
	return err == nil && strings.Contains(out, "package "+pkg.Name)
}

func (p *Pipx) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (p *Pipx) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		if !dryRun && p.Installed(pkg) {
			if verbose {
				fmt.Printf("pipx: package %s is already installed, skipping\n", pkg.Name)
			}
			continue
		}
		args := []string{"install"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "pipx", args...); err != nil {
			return fmt.Errorf("pipx: failed to install %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func (p *Pipx) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		if !dryRun && !p.Installed(pkg) {
			if verbose {
				fmt.Printf("pipx: package %s is not installed, skipping\n", pkg.Name)
			}
			continue
		}
		args := []string{"uninstall"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "pipx", args...); err != nil {
			return fmt.Errorf("pipx: failed to uninstall %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func (p *Pipx) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		args := []string{"upgrade"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "pipx", args...); err != nil {
			return fmt.Errorf("pipx: failed to update/upgrade %s: %w", pkg.Name, err)
		}
	}
	return nil
}
