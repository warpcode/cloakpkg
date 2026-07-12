package installer

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/runner"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Dnf struct{}

var (
	dnfOnce      sync.Once
	dnfDistro    = "fedora"
	dnfVersionID = "40"
)

func initDnfVariables() {
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
				dnfDistro = v
			} else if k == "VERSION_ID" {
				dnfVersionID = v
			}
		}
	}
}

func expandRepoVariablesDnf(s string) string {
	dnfOnce.Do(initDnfVariables)

	s = strings.ReplaceAll(s, "${distro}", dnfDistro)
	s = strings.ReplaceAll(s, "${DISTRO}", dnfDistro)
	s = strings.ReplaceAll(s, "${version_id}", dnfVersionID)
	s = strings.ReplaceAll(s, "${VERSION_ID}", dnfVersionID)
	return s
}

func (d *Dnf) AddRepositories(verbose bool, dryRun bool, repos []config.Repository) error {
	for _, r := range repos {
		repo := r
		repo.Source = expandRepoVariablesDnf(repo.Source)
		repo.Keyring = expandRepoVariablesDnf(repo.Keyring)

		if repo.Source == "" {
			continue
		}

		isCopr := repo.Type == "copr" || strings.HasPrefix(repo.Source, "copr:")
		cleanSource := strings.TrimPrefix(repo.Source, "copr:")

		isRepoFile := repo.Type == "repo" || (repo.Keyring != "" && strings.HasSuffix(repo.Keyring, ".repo"))

		alreadyAdded := false
		if !dryRun {
			if isRepoFile {
				repoPath := filepath.Join("/etc/yum.repos.d", repo.Keyring)
				if _, err := os.Stat(repoPath); err == nil {
					alreadyAdded = true
				}
			} else {
				files, err := os.ReadDir("/etc/yum.repos.d")
				if err == nil {
					for _, file := range files {
						if file.IsDir() || !strings.HasSuffix(file.Name(), ".repo") {
							continue
						}
						filePath := filepath.Join("/etc/yum.repos.d", file.Name())

						if isCopr {
							parts := strings.Split(cleanSource, "/")
							match := true
							for _, part := range parts {
								part = strings.TrimSpace(part)
								if part != "" && !strings.Contains(strings.ToLower(file.Name()), strings.ToLower(part)) {
									match = false
									break
								}
							}
							if match {
								alreadyAdded = true
								break
							}
						} else {
							content, err := os.ReadFile(filePath)
							if err == nil && strings.Contains(string(content), cleanSource) {
								alreadyAdded = true
								break
							}
						}
					}
				}
			}
		}

		if !alreadyAdded {
			if verbose || dryRun {
				if isRepoFile {
					fmt.Printf("dnf: writing repo file to /etc/yum.repos.d/%s\n", repo.Keyring)
				} else if isCopr {
					fmt.Printf("dnf: enabling COPR repository %s\n", cleanSource)
				} else {
					fmt.Printf("dnf: adding repository %s\n", repo.Source)
				}
			}
			if !dryRun {
				if isRepoFile {
					repoPath := filepath.Join("/etc/yum.repos.d", repo.Keyring)
					tmpFile, err := os.CreateTemp("", "cloakpkg-dnf-repo-*")
					if err != nil {
						return fmt.Errorf("dnf: failed to create temp file: %w", err)
					}
					tmpPath := tmpFile.Name()
					defer os.Remove(tmpPath)

					if _, err := tmpFile.WriteString(repo.Source); err != nil {
						tmpFile.Close()
						return fmt.Errorf("dnf: failed to write repo file content: %w", err)
					}
					tmpFile.Close()

					if err := runner.RunSudo(verbose, dryRun, "cp", tmpPath, repoPath); err != nil {
						return fmt.Errorf("dnf: failed to copy repo file to %s: %w", repoPath, err)
					}
				} else if isCopr {
					if err := runner.RunSudo(verbose, dryRun, "dnf", "copr", "enable", "-y", cleanSource); err != nil {
						return fmt.Errorf("dnf: failed to enable COPR repository %s: %w", cleanSource, err)
					}
				} else {
					if err := runner.RunSudo(verbose, dryRun, "dnf", "config-manager", "--add-repo", repo.Source); err != nil {
						return fmt.Errorf("dnf: failed to add repository %s: %w", repo.Source, err)
					}
				}
			}
		}
	}
	return nil
}

func (d *Dnf) Name() string    { return "dnf" }
func (d *Dnf) Available() bool { return runner.CommandExists("dnf") }
func (d *Dnf) Installed(pkg config.Package) bool {
	return runner.RunCheck("rpm", "-q", pkg.Name) == nil
}

func (d *Dnf) Install(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toInstall []string
		for _, pkg := range group {
			if !dryRun && d.Installed(pkg) {
				if verbose {
					fmt.Printf("dnf: package %s is already installed, skipping\n", pkg.Name)
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
		if err := runner.RunSudo(verbose, dryRun, "dnf", args...); err != nil {
			return fmt.Errorf("dnf: failed to install packages %v: %w", toInstall, err)
		}
	}
	return nil
}

func (d *Dnf) Uninstall(verbose bool, dryRun bool, pkgs []config.Package) error {
	keys, groups := GroupPackagesByExtraParams(pkgs)
	for _, key := range keys {
		group := groups[key]
		var toUninstall []string
		for _, pkg := range group {
			if !dryRun && !d.Installed(pkg) {
				if verbose {
					fmt.Printf("dnf: package %s is not installed, skipping\n", pkg.Name)
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
		args = append(args, toUninstall...)
		if err := runner.RunSudo(verbose, dryRun, "dnf", args...); err != nil {
			return fmt.Errorf("dnf: failed to uninstall packages %v: %w", toUninstall, err)
		}
	}
	return nil
}

func (d *Dnf) Update(verbose bool, dryRun bool, pkgs []config.Package) error {
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
		args := []string{"upgrade", "-y"}
		args = append(args, group[0].ExtraParams...)
		args = append(args, toUpdate...)
		if err := runner.RunSudo(verbose, dryRun, "dnf", args...); err != nil {
			return fmt.Errorf("dnf: failed to update packages %v: %w", toUpdate, err)
		}
	}
	return nil
}
