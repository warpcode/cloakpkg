package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
)

type Zypper struct{}

func (z *Zypper) Name() string    { return "zypper" }
func (z *Zypper) Available() bool { return runner.CommandExists("zypper") }
func (z *Zypper) Installed(pkg config.Package) bool {
	return runner.RunCheck("rpm", "-q", pkg.Name) == nil
}

func (z *Zypper) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (z *Zypper) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && z.Installed(pkg) {
				if verbose {
					fmt.Printf("zypper: package %s is already installed, skipping\n", pkg.Name)
				}
				continue
			}
			toInstall = append(toInstall, pkg.Name)
		}
		if len(toInstall) == 0 {
			continue
		}
		args := []string{"install", "-y"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toInstall...)
		if err := runner.RunSudo(verbose, dryRun, "zypper", args...); err != nil {
			return fmt.Errorf("zypper: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (z *Zypper) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !z.Installed(pkg) {
				if verbose {
					fmt.Printf("zypper: package %s is not installed, skipping\n", pkg.Name)
				}
				continue
			}
			toUninstall = append(toUninstall, pkg.Name)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"remove", "-y"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUninstall...)
		if err := runner.RunSudo(verbose, dryRun, "zypper", args...); err != nil {
			return fmt.Errorf("zypper: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (z *Zypper) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		args := []string{"update", "-y"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUpdate...)
		if err := runner.RunSudo(verbose, dryRun, "zypper", args...); err != nil {
			return fmt.Errorf("zypper: failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
