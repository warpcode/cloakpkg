package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
)

type Mise struct{}

func (m *Mise) Name() string    { return "mise" }
func (m *Mise) Available() bool { return runner.CommandExists("mise") }
func (m *Mise) Installed(pkg config.Package) bool {
	toolName := pkg.Name
	if pkg.Version != "" {
		toolName = pkg.Name + "@" + pkg.Version
	}
	return runner.RunCheck("mise", "ls", toolName) == nil
}

func (m *Mise) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (m *Mise) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			toolName := pkg.Name
			if pkg.Version != "" {
				toolName = pkg.Name + "@" + pkg.Version
			}
			if !dryRun && m.Installed(pkg) {
				if verbose {
					fmt.Printf("mise: tool %s is already installed, skipping\n", toolName)
				}
				continue
			}
			toInstall = append(toInstall, toolName)
		}
		if len(toInstall) == 0 {
			continue
		}
		args := []string{"install"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toInstall...)
		if err := runner.Run(verbose, dryRun, "mise", args...); err != nil {
			return fmt.Errorf("mise: failed to install tools %v: %w", toInstall, err)
		}
	}
	return nil
}

func (m *Mise) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			toolName := pkg.Name
			if pkg.Version != "" {
				toolName = pkg.Name + "@" + pkg.Version
			}
			if !dryRun && !m.Installed(pkg) {
				if verbose {
					fmt.Printf("mise: tool %s is not installed, skipping\n", toolName)
				}
				continue
			}
			toUninstall = append(toUninstall, toolName)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"uninstall"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUninstall...)
		if err := runner.Run(verbose, dryRun, "mise", args...); err != nil {
			return fmt.Errorf("mise: failed to uninstall tools %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (m *Mise) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUpdate []string
		for _, pkg := range group {
			toolName := pkg.Name
			if pkg.Version != "" {
				toolName = pkg.Name + "@" + pkg.Version
			}
			toUpdate = append(toUpdate, toolName)
		}
		if len(toUpdate) == 0 {
			continue
		}
		args := []string{"upgrade"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUpdate...)
		if err := runner.Run(verbose, dryRun, "mise", args...); err != nil {
			return fmt.Errorf("mise: failed to update/upgrade tools %v: %w", toUpdate, err)
		}
	}
	return nil
}
