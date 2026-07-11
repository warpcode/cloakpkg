package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
)

type Pacman struct{}

func (p *Pacman) Name() string    { return "pacman" }
func (p *Pacman) Available() bool { return runner.CommandExists("pacman") }
func (p *Pacman) Installed(pkg config.Package) bool {
	return runner.RunCheck("pacman", "-Qq", pkg.Name) == nil
}

func (p *Pacman) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (p *Pacman) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && p.Installed(pkg) {
				if verbose {
					fmt.Printf("pacman: package %s is already installed, skipping\n", pkg.Name)
				}
				continue
			}
			toInstall = append(toInstall, pkg.Name)
		}
		if len(toInstall) == 0 {
			continue
		}
		args := []string{"-S", "--noconfirm"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toInstall...)
		if err := runner.RunSudo(verbose, dryRun, "pacman", args...); err != nil {
			return fmt.Errorf("pacman: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (p *Pacman) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !p.Installed(pkg) {
				if verbose {
					fmt.Printf("pacman: package %s is not installed, skipping\n", pkg.Name)
				}
				continue
			}
			toUninstall = append(toUninstall, pkg.Name)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"-R", "--noconfirm"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUninstall...)
		if err := runner.RunSudo(verbose, dryRun, "pacman", args...); err != nil {
			return fmt.Errorf("pacman: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (p *Pacman) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUpdate []string
		for _, pkg := range group {
			toUpdate = append(toUpdate, pkg.Name)
		}
		if len(toUpdate) == 0 {
			continue
		}
		args := []string{"-S", "--noconfirm"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUpdate...)
		if err := runner.RunSudo(verbose, dryRun, "pacman", args...); err != nil {
			return fmt.Errorf("pacman: failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
