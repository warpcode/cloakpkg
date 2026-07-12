package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"strings"
)

type Flatpak struct{}

func (f *Flatpak) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	for _, repo := range repos {
		if repo.Source == "" {
			continue
		}

		name := repo.Remote
		url := repo.Source
		if name == "" {
			parts := strings.Fields(repo.Source)
			if len(parts) >= 2 {
				name = parts[0]
				url = parts[1]
			} else {
				return fmt.Errorf("flatpak: invalid remote repository format (expected name and url): %q", repo.Source)
			}
		}

		if verbose || dryRun {
			fmt.Printf("flatpak: adding remote repository %s -> %s\n", name, url)
		}
		if !dryRun {
			if err := runner.Run(verbose, dryRun, "flatpak", "remote-add", "--if-not-exists", name, url); err != nil {
				return fmt.Errorf("flatpak: failed to add remote repository %s: %w", name, err)
			}
			// Refresh AppStream metadata so that newly added remote's packages are searchable
			if err := runner.Run(verbose, dryRun, "flatpak", "update", "--appstream"); err != nil {
				if verbose {
					fmt.Printf("flatpak: warning: failed to update appstream metadata: %v\n", err)
				}
			}
		}
	}
	return nil
}

func (f *Flatpak) Name() string    { return "flatpak" }
func (f *Flatpak) Available() bool { return runner.CommandExists("flatpak") }
func (f *Flatpak) Installed(pkg config.Package) bool {
	return runner.RunCheck("flatpak", "info", pkg.Name) == nil
}

func (f *Flatpak) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && f.Installed(pkg) {
				if verbose {
					fmt.Printf("flatpak: package %s is already installed, skipping\n", pkg.Name)
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
		if err := runner.Run(verbose, dryRun, "flatpak", args...); err != nil {
			return fmt.Errorf("flatpak: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (f *Flatpak) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !f.Installed(pkg) {
				if verbose {
					fmt.Printf("flatpak: package %s is not installed, skipping\n", pkg.Name)
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
		if err := runner.Run(verbose, dryRun, "flatpak", args...); err != nil {
			return fmt.Errorf("flatpak: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (f *Flatpak) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		if err := runner.Run(verbose, dryRun, "flatpak", args...); err != nil {
			return fmt.Errorf("flatpak: failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
