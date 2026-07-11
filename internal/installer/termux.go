package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"os"
	"strings"
)

type Termux struct{}

func (t *Termux) Name() string    { return "termux" }
func (t *Termux) Available() bool {
	if !runner.CommandExists("pkg") {
		return false
	}
	return os.Getenv("TERMUX_VERSION") != "" || strings.Contains(os.Getenv("PREFIX"), "com.termux")
}
func (t *Termux) Installed(pkg config.Package) bool {
	return runner.RunCheck("dpkg", "-s", pkg.Name) == nil
}

func (t *Termux) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (t *Termux) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && t.Installed(pkg) {
				if verbose {
					fmt.Printf("termux: package %s is already installed, skipping\n", pkg.Name)
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
		if err := runner.Run(verbose, dryRun, "pkg", args...); err != nil {
			return fmt.Errorf("termux (pkg): failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (t *Termux) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !t.Installed(pkg) {
				if verbose {
					fmt.Printf("termux: package %s is not installed, skipping\n", pkg.Name)
				}
				continue
			}
			toUninstall = append(toUninstall, pkg.Name)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"uninstall", "-y"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUninstall...)
		if err := runner.Run(verbose, dryRun, "pkg", args...); err != nil {
			return fmt.Errorf("termux (pkg): failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (t *Termux) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		args := []string{"install", "-y"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUpdate...)
		if err := runner.Run(verbose, dryRun, "pkg", args...); err != nil {
			return fmt.Errorf("termux (pkg): failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
