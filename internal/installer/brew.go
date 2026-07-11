package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"strings"
)

type Brew struct{}

func (b *Brew) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	var tapped []string
	if !dryRun {
		output, err := runner.RunCheckOutput("brew", "tap")
		if err == nil {
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					tapped = append(tapped, strings.ToLower(line))
				}
			}
		}
	}

	for _, repo := range repos {
		if repo.Source == "" {
			continue
		}
		tapName := strings.ToLower(repo.Source)
		alreadyTapped := false
		for _, t := range tapped {
			if t == tapName {
				alreadyTapped = true
				break
			}
		}

		if !alreadyTapped {
			if verbose || dryRun {
				fmt.Printf("brew: tapping repository %s\n", repo.Source)
			}
			if !dryRun {
				if err := runner.Run(verbose, dryRun, "brew", "tap", repo.Source); err != nil {
					return fmt.Errorf("brew: failed to tap repository %s: %w", repo.Source, err)
				}
			}
		}
	}
	return nil
}

func (b *Brew) Name() string    { return "brew" }
func (b *Brew) Available() bool { return runner.CommandExists("brew") }
func (b *Brew) Installed(pkg config.Package) bool {
	return runner.RunCheck("brew", "list", pkg.Name) == nil
}

func (b *Brew) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && b.Installed(pkg) {
				if verbose {
					fmt.Printf("brew: package %s is already installed, skipping\n", pkg.Name)
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
		if err := runner.Run(verbose, dryRun, "brew", args...); err != nil {
			return fmt.Errorf("brew: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (b *Brew) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !b.Installed(pkg) {
				if verbose {
					fmt.Printf("brew: package %s is not installed, skipping\n", pkg.Name)
				}
				continue
			}
			toUninstall = append(toUninstall, pkg.Name)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"uninstall"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUninstall...)
		if err := runner.Run(verbose, dryRun, "brew", args...); err != nil {
			return fmt.Errorf("brew: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (b *Brew) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		args := []string{"upgrade"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUpdate...)
		if err := runner.Run(verbose, dryRun, "brew", args...); err != nil {
			return fmt.Errorf("brew: failed to update/upgrade packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
