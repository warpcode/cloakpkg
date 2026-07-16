package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
)

type Snap struct{}

func (s *Snap) Name() string    { return "snap" }
func (s *Snap) Available() bool { return runner.CommandExists("snap") }
func (s *Snap) Installed(pkg config.Package) bool {
	return runner.RunCheck("snap", "info", pkg.Name) == nil
}

func (s *Snap) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (s *Snap) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && s.Installed(pkg) {
				if verbose {
					fmt.Printf("snap: package %s is already installed, skipping\n", pkg.Name)
				}
				continue
			}
			toInstall = append(toInstall, pkg.Name)
		}
		if len(toInstall) == 0 {
			continue
		}
		args := []string{"install"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toInstall...)
		if err := runner.RunSudo(verbose, dryRun, "snap", args...); err != nil {
			return fmt.Errorf("snap: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (s *Snap) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !s.Installed(pkg) {
				if verbose {
					fmt.Printf("snap: package %s is not installed, skipping\n", pkg.Name)
				}
				continue
			}
			toUninstall = append(toUninstall, pkg.Name)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"remove"}
		// Filter out flags that are not valid for 'remove' but might be in ExtraParams (like --classic)
		for _, param := range group[0].ExtraParams {
			if param != "--classic" {
				args = append(args, param)
			}
		}
		args = append(args, toUninstall...)
		if err := runner.RunSudo(verbose, dryRun, "snap", args...); err != nil {
			return fmt.Errorf("snap: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (s *Snap) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		args := []string{"refresh"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUpdate...)
		if err := runner.RunSudo(verbose, dryRun, "snap", args...); err != nil {
			return fmt.Errorf("snap: failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
