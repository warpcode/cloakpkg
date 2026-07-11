# cloakpkg

A declarative, cross-platform universal package manager wrapper written in Go.

`cloakpkg` consolidates system-level package managers and custom tool chains into a single, cohesive, tag-filtered declarative configuration file. It matches available package managers dynamically, applies fallback rules based on priority, supports custom install script scripts, and executes pre/post hooks.

---

## Features

- **Declarative Package Management**: Define your development environment, infra tools, and settings inside a single YAML or JSON file.
- **Smart Fallback Priority**: Execute installs dynamically using system-specific package managers (e.g., attempt installation via `brew`, fallback to `apt`, then run a `custom` script).
- **16+ Built-in Adapters**: Built-in support for `apt`, `brew`, `snap`, `flatpak`, `dnf`, `pacman`, `apk`, `zypper`, `npm`, `cargo`, `go`, `pipx`, `uvx`, `gem`, `mise`, and `termux`.
- **Tag-Based Filtering**: Filter package groups (bundles) selectively using tag inclusion (`-t`) or tag exclusion (`-e`).
- **Flexible Custom Providers**: Define arbitrary shell commands to `detect`, `install`, `update`, and `uninstall` tools not distributed via standard package managers.
- **Hook Lifecycle Execution**: Run bundle-level and provider-level pre/post hooks for `install`, `uninstall`, and `update` commands.
- **Dry-run Mode (`-n`)**: Preview shell commands and package managers invoked without modifying system state.
- **Verbose Mode (`-v`)**: Real-time output and diagnostic logging for command execution.
- **Installation Verification (`check`)**: Inspect the current status of all defined packages to verify whether they are installed.

---

## Prerequisites

- [Go](https://go.dev/) 1.21 or higher.
- Package managers utilized in your configuration (e.g., `apt`, `brew`, `npm`, `cargo`) must be installed on your system path.

---

## Project Structure

```
├── main.go                     # CLI Entry point
├── internal/
│   ├── cli/                    # CLI logic (cli.go, command routing, formatting)
│   ├── config/                 # YAML and JSON config parser and schemas
│   ├── installer/              # Adaptor layers for each supported package manager
│   └── runner/                 # Sudo and shell-safe command execution helpers
├── testdata/                   # Comprehensive JSON & YAML testing configurations
├── go.mod                      # Go module dependencies
└── README.md                   # Project documentation
```

---

## Installation & Local Development

### Direct Execution

To run `cloakpkg` directly without compilation:
```bash
go run . [command] <config_file> [bundle...] [flags]
```

### Installation

To install `cloakpkg` into your `$GOPATH/bin` or `$GOBIN`:
```bash
go install .
```

### Local Testing

To run unit tests across all packages:
```bash
go test -v ./...
```

---

## CLI Command Reference

`cloakpkg` supports the following subcommands:

### `list-installers`
Lists the availability of built-in package managers on the current operating system.
```bash
cloakpkg list-installers
```

### `check`
Evaluates the status of defined packages (e.g., "installed" or "not installed") without modifying the system.
```bash
cloakpkg check <config_file> [bundle...] [flags]
```

### `install`
Installs target package bundles.
```bash
cloakpkg install <config_file> [bundle...] [flags]
```

### `uninstall`
Removes package bundles from the system.
```bash
cloakpkg uninstall <config_file> [bundle...] [flags]
```

### `update`
Updates packages to their latest versions or upgrades dependencies.
```bash
cloakpkg update <config_file> [bundle...] [flags]
```

### `reinstall`
Uninstalls and then fresh-installs package bundles.
```bash
cloakpkg reinstall <config_file> [bundle...] [flags]
```

### Options & Flags
For `install`, `uninstall`, `update`, `reinstall`, and `check`:
- `-t <tags>`: Comma-separated list of tags to include. Only bundles matching one or more tags will be processed.
- `-e <tags>`: Comma-separated list of tags to exclude. Bundles matching one or more of these tags are skipped.
- `-n`: Dry-run mode. Displays what commands would be run without executing them.
- `-v`: Verbose output. Redirection of process stdout/stderr and diagnostic details.

---

## Configuration Schema Guide

The configuration file can be written in either **YAML** or **JSON**.

### Schema Elements

1. **`settings`**:
   - `provider_priority` (array of strings): Defines the search order for package managers. The first available provider in this list defined under a bundle will be selected. If this list is empty, `cloakpkg` executes all available package managers defined on a bundle.

2. **`bundles`**: A map of bundle names to their attributes:
   - `description` (string, optional): A description of the package bundle.
   - `tags` (array of strings, optional): Labels for tag-based filtering (`-t` / `-e`).
   - `providers` (map of provider names to provider configurations):
     - `packages` (array): A list of package definitions:
       - `name` (string): The package name expected by the manager.
       - `version` (string, optional): Specific version constraint.
       - `extra_params` (array of strings, optional): Command flags passed directly to the installer (e.g., `["--cask"]` for Homebrew).
     - `repositories` (array, optional): A list of sources or ppas that must be added/tapped prior to installation. Can be string format or map format:
       - `source` (string): Repository URL, tap source, or PPA.
       - `key_url` (string, optional): GPG keyring public key URL (APT only).
       - `keyring` (string, optional): File path/name where the keyring or repo sources file should be stored.
       - `type` (string, optional): Repository type definition (e.g., `list` or `deb822`).
     - `detect` (string, optional): Shell script command that returns exit code `0` if the custom tool is already installed.
     - `install` (string, optional): Installation command for custom providers.
     - `update` (string, optional): Upgrade command for custom providers.
     - `uninstall` (string, optional): Cleanup/removal command for custom providers.
     - `hooks` (Hooks object, optional): Local pre/post execution scripts.
   - `hooks` (Hooks object, optional): Bundle-level pre/post execution scripts.

3. **`hooks` Structure**:
   - `pre_install` / `post_install` (string)
   - `pre_uninstall` / `post_uninstall` (string)
   - `pre_update` / `post_update` (string)

---

### Configuration Examples

#### YAML Example (`config.yaml`)

```yaml
settings:
  provider_priority: [brew, apt, custom]

bundles:
  lazygit:
    description: "Terminal UI for git"
    tags: [core, cli]
    providers:
      apt:
        repositories:
          - source: "ppa:git-core/ppa"
        packages:
          - name: lazygit
      brew:
        repositories:
          - source: "jesseduffield/lazygit"
        packages:
          - name: lazygit
      custom:
        detect: "command -v lazygit"
        install: "echo 'Installing lazygit custom script'"

  docker:
    description: "Container runtime"
    tags: [core, infra]
    providers:
      apt:
        repositories:
          - source: "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu noble stable"
            key_url: "https://download.docker.com/linux/ubuntu/gpg"
            keyring: "docker.asc"
        packages:
          - name: docker-ce
          - name: docker-ce-cli
      brew:
        packages:
          - name: docker
            extra_params: ["--cask"]

  mise:
    description: "Developer environment manager"
    tags: [dev]
    providers:
      custom:
        detect: "command -v mise"
        install: "curl -fsSL https://mise.run | sh"
        update: "mise self-update"
        uninstall: "rm -rf ~/.local/share/mise ~/.local/bin/mise"
```

#### JSON Example (`config.json`)

```json
{
  "settings": {
    "provider_priority": ["brew", "apt", "custom"]
  },
  "bundles": {
    "tmux": {
      "description": "Terminal multiplexer",
      "tags": ["core", "cli"],
      "providers": {
        "brew": {
          "packages": [
            {
              "name": "tmux",
              "extra_params": ["--formula"]
            }
          ]
        },
        "apt": {
          "packages": [
            {
              "name": "tmux"
            }
          ]
        }
      }
    }
  }
}
```

---

## Lifecycle Hook Execution Sequence

Hooks are executed in the following strict order for `install`, `uninstall`, and `update` commands:

```
[Bundle Pre-Hook] ──> [Provider Pre-Hook] ──> [Installer Action (Install/Uninstall/Update)] ──> [Provider Post-Hook] ──> [Bundle Post-Hook]
```

1. **Bundle Pre-hook**: Executed first before any provider action inside that bundle.
2. **Provider Pre-hook**: Executed directly before the selected provider's action is performed.
3. **Execution**: The provider installs, updates, or uninstalls target packages.
4. **Provider Post-hook**: Executed directly after the selected provider succeeds.
5. **Bundle Post-hook**: Executed after all providers in the bundle complete.

If any pre-hook or package installation command returns a non-zero exit status, execution terminates immediately with an error, preventing downstream post-hooks from firing.

---

## Supported Built-in Installers

| Name | Command Checked | Installation Check Method | Notes |
| :--- | :--- | :--- | :--- |
| `apt` | `apt-get` | `dpkg -s <pkg>` | Supports GPG key downloads and APT source lists |
| `brew` | `brew` | `brew list <pkg>` | Supports formula, cask checks, and taps |
| `snap` | `snap` | `snap list <pkg>` | Linux snap packages |
| `flatpak` | `flatpak` | `flatpak info <pkg>` | Linux Flatpak package check |
| `dnf` | `dnf` | `rpm -q <pkg>` | Red Hat / Fedora package manager |
| `pacman` | `pacman` | `pacman -Qi <pkg>` | Arch Linux package manager |
| `apk` | `apk` | `apk info -e <pkg>` | Alpine Linux package manager |
| `zypper` | `zypper` | `rpm -q <pkg>` | openSUSE package manager |
| `npm` | `npm` | `npm list -g <pkg>` | Node Package Manager global installations |
| `cargo` | `cargo` | `cargo install --list` | Rust build tool and package manager |
| `go` | `go` | checks `go env GOPATH`/bin | Go installer |
| `pipx` | `pipx` | `pipx list` | Python isolated app installer |
| `uvx` | `uv` | `uv tool list` | Fast python tool runner |
| `gem` | `gem` | `gem list -i <pkg>` | Ruby gems installer |
| `mise` | `mise` | `mise ls <pkg>` | Developer tool and env orchestrator |
| `termux` | `pkg` | `dpkg -s <pkg>` | Android Termux environment |
| `custom` | *always available* | runs `detect` script | Executes user-defined shell command templates |
