package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temp YAML config file
	tempFile, err := os.CreateTemp("", "cloakpkg_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := `
settings:
  provider_priority: [brew, apt]
bundles:
  testpkg:
    description: "A test bundle"
    tags: [core]
    providers:
      apt:
        packages:
          - name: "test-apt-package"
            extra_params: ["--no-install-recommends"]
`
	if _, err := tempFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	cfg, err := LoadConfig(tempFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Settings.ProviderPriority) != 2 || cfg.Settings.ProviderPriority[0] != "brew" {
		t.Errorf("Unexpected provider priority: %v", cfg.Settings.ProviderPriority)
	}

	bundle, exists := cfg.Bundles["testpkg"]
	if !exists {
		t.Fatalf("Expected 'testpkg' bundle to exist")
	}

	if bundle.Description != "A test bundle" {
		t.Errorf("Unexpected description: %q", bundle.Description)
	}

	if len(bundle.Tags) != 1 || bundle.Tags[0] != "core" {
		t.Errorf("Unexpected tags: %v", bundle.Tags)
	}

	aptProvider, ok := bundle.Providers["apt"]
	if !ok {
		t.Fatalf("Expected 'apt' provider to exist")
	}

	if len(aptProvider.Packages) != 1 || aptProvider.Packages[0].Name != "test-apt-package" {
		t.Errorf("Unexpected package info: %+v", aptProvider.Packages)
	}

	if len(aptProvider.Packages[0].ExtraParams) != 1 || aptProvider.Packages[0].ExtraParams[0] != "--no-install-recommends" {
		t.Errorf("Unexpected extra params: %v", aptProvider.Packages[0].ExtraParams)
	}
}

func TestLoadConfigFromTestData(t *testing.T) {
	// Test loading static YAML from testdata
	cfg, err := LoadConfig("../../testdata/config.yaml")
	if err != nil {
		t.Fatalf("failed to load testdata/config.yaml: %v", err)
	}
	if cfg.Bundles["lazygit"].Description != "Terminal UI for git" {
		t.Errorf("unexpected description: %q", cfg.Bundles["lazygit"].Description)
	}

	// Test loading static JSON from testdata
	cfgJSON, err := LoadConfig("../../testdata/config.json")
	if err != nil {
		t.Fatalf("failed to load testdata/config.json: %v", err)
	}
	if cfgJSON.Bundles["lazygit"].Description != "Terminal UI for git" {
		t.Errorf("unexpected description: %q", cfgJSON.Bundles["lazygit"].Description)
	}

	// Test loading invalid YAML from testdata
	_, err = LoadConfig("../../testdata/invalid.yaml")
	if err == nil {
		t.Error("expected error when loading invalid.yaml, got nil")
	}
}

func TestPackageUnmarshal(t *testing.T) {
	// Test YAML unmarshaling of package as string vs object
	content := `
bundles:
  testpkg:
    providers:
      apt:
        packages:
          - "simple-string-package"
          - name: "object-package"
            extra_params: ["--classic"]
`
	tempFile, err := os.CreateTemp("", "cloakpkg_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	cfg, err := LoadConfig(tempFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	bundle := cfg.Bundles["testpkg"]
	apt := bundle.Providers["apt"]
	if len(apt.Packages) != 2 {
		t.Fatalf("Expected 2 packages, got %d", len(apt.Packages))
	}

	if apt.Packages[0].Name != "simple-string-package" {
		t.Errorf("Expected first package name to be 'simple-string-package', got %q", apt.Packages[0].Name)
	}

	if apt.Packages[1].Name != "object-package" || len(apt.Packages[1].ExtraParams) != 1 || apt.Packages[1].ExtraParams[0] != "--classic" {
		t.Errorf("Unexpected second package: %+v", apt.Packages[1])
	}
}

func TestLoadConfigWithHooks(t *testing.T) {
	content := `
bundles:
  hooked-pkg:
    description: "A bundle with hooks"
    hooks:
      pre_install: "echo 'pre-install bundle'"
      post_install: "echo 'post-install bundle'"
    providers:
      apt:
        packages:
          - "test-package"
        hooks:
          pre_install: "echo 'pre-install apt'"
          post_install: "echo 'post-install apt'"
`
	tempFile, err := os.CreateTemp("", "cloakpkg_test_hooks_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	cfg, err := LoadConfig(tempFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	bundle, exists := cfg.Bundles["hooked-pkg"]
	if !exists {
		t.Fatalf("Expected 'hooked-pkg' bundle to exist")
	}

	if bundle.Hooks == nil {
		t.Fatalf("Expected bundle hooks to be populated, got nil")
	}
	if bundle.Hooks.PreInstall != "echo 'pre-install bundle'" {
		t.Errorf("Expected bundle pre_install hook %q, got %q", "echo 'pre-install bundle'", bundle.Hooks.PreInstall)
	}
	if bundle.Hooks.PostInstall != "echo 'post-install bundle'" {
		t.Errorf("Expected bundle post_install hook %q, got %q", "echo 'post-install bundle'", bundle.Hooks.PostInstall)
	}

	apt, ok := bundle.Providers["apt"]
	if !ok {
		t.Fatalf("Expected 'apt' provider to exist")
	}

	if apt.Hooks == nil {
		t.Fatalf("Expected provider hooks to be populated, got nil")
	}
	if apt.Hooks.PreInstall != "echo 'pre-install apt'" {
		t.Errorf("Expected provider pre_install hook %q, got %q", "echo 'pre-install apt'", apt.Hooks.PreInstall)
	}
	if apt.Hooks.PostInstall != "echo 'post-install apt'" {
		t.Errorf("Expected provider post_install hook %q, got %q", "echo 'post-install apt'", apt.Hooks.PostInstall)
	}
}
