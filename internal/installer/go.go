package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Go struct{}

func (g *Go) Name() string    { return "go" }
func (g *Go) Available() bool { return runner.CommandExists("go") }
func (g *Go) Installed(pkg config.Package) bool {
	binName := getGoBinaryName(pkg.Name)
	return runner.CommandExists(binName)
}

func (g *Go) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	return nil
}

func (g *Go) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		if !dryRun && g.Installed(pkg) {
			binName := getGoBinaryName(pkg.Name)
			if verbose {
				fmt.Printf("go: binary %s is already in PATH, skipping\n", binName)
			}
			continue
		}
		args := []string{"install"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "go", args...); err != nil {
			return fmt.Errorf("go: failed to install %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func (g *Go) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	goBin := os.Getenv("GOBIN")
	if goBin == "" {
		goPath := os.Getenv("GOPATH")
		if goPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("go: failed to find user home dir: %w", err)
			}
			goPath = filepath.Join(home, "go")
		}
		goBin = filepath.Join(goPath, "bin")
	}

	for _, pkg := range pkgs {
		binName := getGoBinaryName(pkg.Name)
		binPath := filepath.Join(goBin, binName)
		if !dryRun {
			if _, err := os.Stat(binPath); err == nil {
				if verbose {
					fmt.Printf("go: removing binary %s\n", binPath)
				}
				if err := os.Remove(binPath); err != nil {
					return fmt.Errorf("go: failed to remove binary %s: %w", binPath, err)
				}
			} else {
				if verbose {
					fmt.Printf("go: binary %s not found in GOBIN (%s), skipping\n", binName, binPath)
				}
			}
		}
	}
	return nil
}

func (g *Go) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
	for _, pkg := range pkgs {
		args := []string{"install"}
		args = append(args, pkg.ExtraParams...)
		args = append(args, pkg.Name)
		if err := runner.Run(verbose, dryRun, "go", args...); err != nil {
			return fmt.Errorf("go: failed to update %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func getGoBinaryName(pkgName string) string {
	parts := strings.Split(pkgName, "/")
	last := parts[len(parts)-1]
	pkgParts := strings.Split(last, "@")
	return pkgParts[0]
}
