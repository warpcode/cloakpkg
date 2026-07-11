package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Settings Settings          `yaml:"settings" json:"settings"`
	Bundles  map[string]Bundle `yaml:"bundles" json:"bundles"`
}

type Settings struct {
	ProviderPriority []string `yaml:"provider_priority" json:"provider_priority"`
}

type Bundle struct {
	Description string              `yaml:"description" json:"description"`
	Tags        []string            `yaml:"tags" json:"tags"`
	Providers   map[string]Provider `yaml:"providers" json:"providers"`
	Hooks       *Hooks              `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

type Provider struct {
	Packages     []Package    `yaml:"packages" json:"packages"`
	Repositories []Repository `yaml:"repositories,omitempty" json:"repositories,omitempty"`
	Detect       string       `yaml:"detect" json:"detect"`
	InstallCmd   string       `yaml:"install" json:"install"`
	Update       string       `yaml:"update" json:"update"`
	Uninstall    string       `yaml:"uninstall" json:"uninstall"`
	Hooks        *Hooks       `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

type Hooks struct {
	PreInstall    string `yaml:"pre_install,omitempty" json:"pre_install,omitempty"`
	PostInstall   string `yaml:"post_install,omitempty" json:"post_install,omitempty"`
	PreUninstall  string `yaml:"pre_uninstall,omitempty" json:"pre_uninstall,omitempty"`
	PostUninstall string `yaml:"post_uninstall,omitempty" json:"post_uninstall,omitempty"`
	PreUpdate     string `yaml:"pre_update,omitempty" json:"pre_update,omitempty"`
	PostUpdate    string `yaml:"post_update,omitempty" json:"post_update,omitempty"`
}

type Repository struct {
	Source  string `yaml:"source" json:"source"`
	KeyURL  string `yaml:"key_url,omitempty" json:"key_url,omitempty"`
	Keyring string `yaml:"keyring,omitempty" json:"keyring,omitempty"`
	Type    string `yaml:"type,omitempty" json:"type,omitempty"`
	Remote  string `yaml:"remote,omitempty" json:"remote,omitempty"`
}

func (r *Repository) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err == nil {
		r.Source = str
		return nil
	}
	type rawRepo Repository
	var raw rawRepo
	if err := value.Decode(&raw); err != nil {
		return err
	}
	*r = Repository(raw)
	return nil
}

func (r *Repository) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		r.Source = str
		return nil
	}
	type rawRepo Repository
	var raw rawRepo
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*r = Repository(raw)
	return nil
}

type Package struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	ExtraParams []string `yaml:"extra_params,omitempty" json:"extra_params,omitempty"`
}

func (p *Package) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err == nil {
		p.Name = str
		return nil
	}
	type rawPkg Package
	var raw rawPkg
	if err := value.Decode(&raw); err != nil {
		return err
	}
	*p = Package(raw)
	return nil
}

func (p *Package) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		p.Name = str
		return nil
	}
	type rawPkg Package
	var raw rawPkg
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*p = Package(raw)
	return nil
}

// LoadConfig reads and parses a JSON or YAML config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if strings.HasSuffix(path, ".json") {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	} else {
		// Fallback: try YAML (which is a superset of JSON)
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("unsupported file extension and failed to parse as YAML: %w", err)
		}
	}

	return &cfg, nil
}
