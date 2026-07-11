package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
)

type Apk struct{}

func (a *Apk) Name() string    { return "apk" }
func (a *Apk) Available() bool { return runner.CommandExists("apk") }
func (a *Apk) Installed(pkg config.Package) bool {
	return runner.RunCheck("apk", "info", "-e", pkg.Name) == nil
}

func (a *Apk) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (a *Apk) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && a.Installed(pkg) {
				if verbose {
					fmt.Printf("apk: package %s is already installed, skipping\n", pkg.Name)
				}
				continue
			}
			toInstall = append(toInstall, pkg.Name)
		}
		if len(toInstall) == 0 {
			continue
		}
		args := []string{"add"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, "--")
		args = append(args, toInstall...)
		if err := runner.RunSudo(verbose, dryRun, "apk", args...); err != nil {
			return fmt.Errorf("apk: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (a *Apk) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !a.Installed(pkg) {
				if verbose {
					fmt.Printf("apk: package %s is not installed, skipping\n", pkg.Name)
				}
				continue
			}
			toUninstall = append(toUninstall, pkg.Name)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"del"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, "--")
		args = append(args, toUninstall...)
		if err := runner.RunSudo(verbose, dryRun, "apk", args...); err != nil {
			return fmt.Errorf("apk: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (a *Apk) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		args := []string{"add", "--upgrade"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, "--")
		args = append(args, toUpdate...)
		if err := runner.RunSudo(verbose, dryRun, "apk", args...); err != nil {
			return fmt.Errorf("apk: failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
