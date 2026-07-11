package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
)

type Npm struct{}

func (n *Npm) Name() string    { return "npm" }
func (n *Npm) Available() bool { return runner.CommandExists("npm") }
func (n *Npm) Installed(pkg config.Package) bool {
	return runner.RunCheck("npm", "list", "-g", pkg.Name) == nil
}

func (n *Npm) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (n *Npm) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && n.Installed(pkg) {
				if verbose {
					fmt.Printf("npm: package %s is already installed globally, skipping\n", pkg.Name)
				}
				continue
			}
			toInstall = append(toInstall, pkg.Name)
		}
		if len(toInstall) == 0 {
			continue
		}
		args := []string{"install", "-g"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toInstall...)
		if err := runner.Run(verbose, dryRun, "npm", args...); err != nil {
			return fmt.Errorf("npm: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (n *Npm) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !n.Installed(pkg) {
				if verbose {
					fmt.Printf("npm: package %s is not installed globally, skipping\n", pkg.Name)
				}
				continue
			}
			toUninstall = append(toUninstall, pkg.Name)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"uninstall", "-g"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUninstall...)
		if err := runner.Run(verbose, dryRun, "npm", args...); err != nil {
			return fmt.Errorf("npm: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (n *Npm) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		args := []string{"update", "-g"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUpdate...)
		if err := runner.Run(verbose, dryRun, "npm", args...); err != nil {
			return fmt.Errorf("npm: failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
