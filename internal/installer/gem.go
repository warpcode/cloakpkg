package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
)

type Gem struct{}

func (g *Gem) Name() string    { return "gem" }
func (g *Gem) Available() bool { return runner.CommandExists("gem") }
func (g *Gem) Installed(pkg config.Package) bool {
	return runner.RunCheck("gem", "list", "-i", pkg.Name) == nil
}

func (g *Gem) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	for _, repo := range repos {
		if repo.URL != "" {
			fmt.Printf("Adding Gem repository %s...\n", repo.URL)
			err := runner.ExecuteCommand(verbose, dryRun, "gem", []string{"sources", "-a", repo.URL}, []string{})
			if err != nil {
				return fmt.Errorf("gem: failed to add repository %s: %w", repo.URL, err)
			}
		}
	}
	return nil
}

func (g *Gem) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && g.Installed(pkg) {
				if verbose {
					fmt.Printf("gem: package %s is already installed, skipping\n", pkg.Name)
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
		if err := runner.Run(verbose, dryRun, "gem", args...); err != nil {
			return fmt.Errorf("gem: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (g *Gem) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !g.Installed(pkg) {
				if verbose {
					fmt.Printf("gem: package %s is not installed, skipping\n", pkg.Name)
				}
				continue
			}
			toUninstall = append(toUninstall, pkg.Name)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"uninstall", "-a", "-x"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUninstall...)
		if err := runner.Run(verbose, dryRun, "gem", args...); err != nil {
			return fmt.Errorf("gem: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (g *Gem) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		args := []string{"update"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUpdate...)
		if err := runner.Run(verbose, dryRun, "gem", args...); err != nil {
			return fmt.Errorf("gem: failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
