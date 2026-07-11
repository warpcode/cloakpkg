package installer

import (
	"cloakpkg/internal/config"
	"strings"
)

// GroupPackagesByExtraParams groups a list of packages by their ExtraParams.
// It returns the slice of unique key strings in insertion order, and the map of keys to packages.
func GroupPackagesByExtraParams(pkgs []config.Package) ([]string, map[string][]config.Package) {
	groups := make(map[string][]config.Package)
	var keys []string
	for _, pkg := range pkgs {
		key := strings.Join(pkg.ExtraParams, "\x00")
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], pkg)
	}
	return keys, groups
}

// Installer defines the interface that all built-in package manager adapters must implement.
type Installer interface {
	Name() string
	Available() bool
	Installed(pkg config.Package) bool
	Install(verbose bool, dryRun bool, packages []config.Package) error
	Uninstall(verbose bool, dryRun bool, packages []config.Package) error
	Update(verbose bool, dryRun bool, packages []config.Package) error
	AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error
}

// GetRegistry returns a map of all supported built-in package managers.
func GetRegistry() map[string]Installer {
	return map[string]Installer{
		"apt":     &Apt{},
		"brew":    &Brew{},
		"snap":    &Snap{},
		"flatpak": &Flatpak{},
		"dnf":     &Dnf{},
		"pacman":  &Pacman{},
		"apk":     &Apk{},
		"zypper":  &Zypper{},
		"npm":     &Npm{},
		"cargo":   &Cargo{},
		"go":      &Go{},
		"pipx":    &Pipx{},
		"uvx":     &Uvx{},
		"gem":     &Gem{},
		"mise":    &Mise{},
		"termux":  &Termux{},
	}
}
