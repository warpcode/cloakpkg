package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Apt struct{}

var (
	aptOnce     sync.Once
	aptArch     = "amd64"
	aptDistro   = "ubuntu"
	aptCodename = "noble"
)

func initAptVariables() {
	if runner.CommandExists("dpkg") {
		if out, err := runner.RunCheckOutput("dpkg", "--print-architecture"); err == nil {
			aptArch = strings.TrimSpace(out)
		}
	}

	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if k == "ID" {
				aptDistro = v
			} else if k == "VERSION_CODENAME" {
				aptCodename = v
			}
		}
	}
}

func expandRepoVariables(s string) string {
	aptOnce.Do(initAptVariables)

	s = strings.ReplaceAll(s, "${ARCH}", aptArch)
	s = strings.ReplaceAll(s, "${arch}", aptArch)
	s = strings.ReplaceAll(s, "${DISTRO}", aptDistro)
	s = strings.ReplaceAll(s, "${distro}", aptDistro)
	s = strings.ReplaceAll(s, "${VERSION_CODENAME}", aptCodename)
	s = strings.ReplaceAll(s, "${version_codename}", aptCodename)
	return s
}

func downloadKey(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (a *Apt) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	anyAdded := false
	for _, r := range repos {
		repo := r
		repo.KeyURL = expandRepoVariables(repo.KeyURL)
		repo.Source = expandRepoVariables(repo.Source)
		repo.Keyring = expandRepoVariables(repo.Keyring)

		// 1. Download and handle keyring GPG key if specified
		if repo.KeyURL != "" && repo.Keyring != "" {
			keyringPath := filepath.Join("/etc/apt/keyrings", repo.Keyring)
			exists := false
			if !dryRun {
				if _, err := os.Stat(keyringPath); err == nil {
					exists = true
				}
			}
			if !exists {
				if verbose || dryRun {
					fmt.Printf("apt: downloading and installing GPG keyring %s from %s\n", repo.Keyring, repo.KeyURL)
				}
				if !dryRun {
					keyBytes, err := downloadKey(repo.KeyURL)
					if err != nil {
						return fmt.Errorf("apt: failed to download keyring from %s: %w", repo.KeyURL, err)
					}

					// Auto-detect if ASCII armored
					isArmored := strings.HasPrefix(string(keyBytes), "-----BEGIN PGP PUBLIC KEY BLOCK-----")

					tmpFile, err := os.CreateTemp("", "cloakpkg-key-*")
					if err != nil {
						return fmt.Errorf("apt: failed to create temp file: %w", err)
					}
					tmpPath := tmpFile.Name()
					defer os.Remove(tmpPath)

					if _, err := tmpFile.Write(keyBytes); err != nil {
						tmpFile.Close()
						return fmt.Errorf("apt: failed to write key to temp file: %w", err)
					}
					tmpFile.Close()

					if err := runner.RunSudo(verbose, dryRun, "mkdir", "-p", "/etc/apt/keyrings"); err != nil {
						return fmt.Errorf("apt: failed to create /etc/apt/keyrings: %w", err)
					}

					if isArmored {
						if err := runner.RunSudo(verbose, dryRun, "gpg", "--dearmor", "-o", keyringPath, tmpPath); err != nil {
							return fmt.Errorf("apt: failed to dearmor keyring using gpg: %w", err)
						}
					} else {
						if err := runner.RunSudo(verbose, dryRun, "cp", tmpPath, keyringPath); err != nil {
							return fmt.Errorf("apt: failed to copy keyring to %s: %w", keyringPath, err)
						}
					}
				}
			}
		}

		// 2. Add repository source
		isList := repo.Type == "list" || (repo.Keyring != "" && strings.HasSuffix(repo.Keyring, ".list"))
		isDeb822 := repo.Type == "deb822" || (repo.Keyring != "" && strings.HasSuffix(repo.Keyring, ".sources"))

		if isList || isDeb822 {
			sourcePath := ""
			if isList {
				sourcePath = filepath.Join("/etc/apt/sources.list.d", repo.Keyring)
			} else {
				sourceName := "custom"
				if repo.Keyring != "" {
					sourceName = strings.TrimSuffix(repo.Keyring, filepath.Ext(repo.Keyring))
				}
				sourcePath = filepath.Join("/etc/apt/sources.list.d", sourceName+".sources")
			}

			exists := false
			if !dryRun {
				if _, err := os.Stat(sourcePath); err == nil {
					exists = true
				}
			}

			if !exists {
				if verbose || dryRun {
					fmt.Printf("apt: adding repository file %s\n", sourcePath)
				}
				if !dryRun {
					tmpFile, err := os.CreateTemp("", "cloakpkg-apt-source-*")
					if err != nil {
						return fmt.Errorf("apt: failed to create temp file: %w", err)
					}
					tmpPath := tmpFile.Name()
					defer os.Remove(tmpPath)

					if _, err := tmpFile.WriteString(repo.Source); err != nil {
						tmpFile.Close()
						return fmt.Errorf("apt: failed to write source file content: %w", err)
					}
					tmpFile.Close()

					if err := runner.RunSudo(verbose, dryRun, "cp", tmpPath, sourcePath); err != nil {
						return fmt.Errorf("apt: failed to copy source file to %s: %w", sourcePath, err)
					}
					anyAdded = true
				}
			}
		} else if repo.Source != "" {
			alreadyAdded := false
			if !dryRun {
				searchTerm := strings.TrimPrefix(repo.Source, "ppa:")

				args := []string{"-q", "-r", searchTerm, "/etc/apt/sources.list"}
				if _, err := os.Stat("/etc/apt/sources.list.d"); err == nil {
					args = append(args, "/etc/apt/sources.list.d")
				}
				if runner.RunCheck("grep", args...) == nil {
					alreadyAdded = true
				}
			}

			if !alreadyAdded {
				if verbose || dryRun {
					fmt.Printf("apt: adding repository %s\n", repo.Source)
				}
				if !dryRun {
					if err := runner.RunSudo(verbose, dryRun, "add-apt-repository", "-y", repo.Source); err != nil {
						return fmt.Errorf("apt: failed to add repository %s: %w", repo.Source, err)
					}
					anyAdded = true
				}
			}
		}
	}

	if anyAdded {
		if verbose || dryRun {
			fmt.Println("apt: updating package lists")
		}
		if err := runner.RunSudo(verbose, dryRun, "apt-get", "update"); err != nil {
			return fmt.Errorf("apt: failed to update package lists: %w", err)
		}
	}
	return nil
}

func (a *Apt) Name() string    { return "apt" }
func (a *Apt) Available() bool { return runner.CommandExists("apt-get") }
func (a *Apt) Installed(pkg config.Package) bool {
	return runner.RunCheck("dpkg", "-s", pkg.Name) == nil
}

func (a *Apt) bulkInstalled(pkgs []config.Package) map[string]bool {
	installedMap := make(map[string]bool)
	if len(pkgs) == 0 {
		return installedMap
	}

	var pkgNames []string
	for _, pkg := range pkgs {
		pkgNames = append(pkgNames, pkg.Name)
	}

	args := append([]string{"-W", "-f=${Package} ${Status}\n"}, pkgNames...)
	out, _ := runner.RunCheckOutput("dpkg-query", args...)

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			pkgName := parts[0]
			status := parts[1]
			if strings.Contains(status, "install ok installed") {
				installedMap[pkgName] = true
			}
		}
	}
	return installedMap
}

func (a *Apt) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string

		var installedMap map[string]bool
		if !dryRun {
			installedMap = a.bulkInstalled(group)
		}

		for _, pkg := range group {
			if !dryRun && installedMap[pkg.Name] {
				if verbose {
					fmt.Printf("apt: package %s is already installed, skipping\n", pkg.Name)
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
		args = append(args, "--")
		args = append(args, toInstall...)
		if err := runner.RunSudo(verbose, dryRun, "apt-get", args...); err != nil {
			return fmt.Errorf("apt: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (a *Apt) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string

		var installedMap map[string]bool
		if !dryRun {
			installedMap = a.bulkInstalled(group)
		}

		for _, pkg := range group {
			if !dryRun && !installedMap[pkg.Name] {
				if verbose {
					fmt.Printf("apt: package %s is not installed, skipping\n", pkg.Name)
				}
				continue
			}
			toUninstall = append(toUninstall, pkg.Name)
		}
		if len(toUninstall) == 0 {
			continue
		}
		args := []string{"remove", "-y"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, "--")
		args = append(args, toUninstall...)
		if err := runner.RunSudo(verbose, dryRun, "apt-get", args...); err != nil {
			return fmt.Errorf("apt: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (a *Apt) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		args := []string{"install", "-y", "--only-upgrade"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, "--")
		args = append(args, toUpdate...)
		if err := runner.RunSudo(verbose, dryRun, "apt-get", args...); err != nil {
			return fmt.Errorf("apt: failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
