package cli

import (
	"cloakpkg/internal/config"
	"cloakpkg/internal/installer"
	"cloakpkg/internal/runner"
	"flag"
	"fmt"
	"os"
	"strings"
)

func Run() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list-installers":
		runListInstallers()
	case "install", "uninstall", "update", "reinstall", "check":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: missing config file path argument.\n\n")
			printUsage()
			os.Exit(1)
		}
		configFile := os.Args[2]
		runBundleCommand(command, configFile)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("cloakpkg - A universal package installer wrapper")
	fmt.Println("\nUsage:")
	fmt.Println("  cloakpkg install <config> [bundle...]   [flags]")
	fmt.Println("  cloakpkg uninstall <config> [bundle...] [flags]")
	fmt.Println("  cloakpkg update <config> [bundle...]    [flags]")
	fmt.Println("  cloakpkg reinstall <config> [bundle...] [flags]")
	fmt.Println("  cloakpkg check <config> [bundle...]     [flags]")
	fmt.Println("  cloakpkg list-installers")
	fmt.Println("\nFlags (for install, uninstall, update, reinstall & check):")
	fmt.Println("  -t <tags>         Comma-separated list of tags to include")
	fmt.Println("  -e <tags>         Comma-separated list of tags to exclude")
	fmt.Println("  -n                Dry-run mode (print commands without executing)")
	fmt.Println("  -v                Verbose output")
}

func runListInstallers() {
	registry := installer.GetRegistry()
	fmt.Println("Checking installer availability on this system:")
	for name, inst := range registry {
		status := "NOT AVAILABLE"
		if inst.Available() {
			status = "AVAILABLE"
		}
		fmt.Printf("  %-10s : %s\n", name, status)
	}
	fmt.Printf("  %-10s : AVAILABLE (runs custom shell scripts)\n", "custom")
}

func runBundleCommand(command string, configFile string) {
	// Parse subcommand flags
	fs := flag.NewFlagSet(command, flag.ExitOnError)
	tagsFlag := fs.String("t", "", "Comma-separated list of tags to include")
	excludeTagsFlag := fs.String("e", "", "Comma-separated list of tags to exclude")
	dryRunFlag := fs.Bool("n", false, "Dry run mode")
	verboseFlag := fs.Bool("v", false, "Verbose output")

	// Parse starting from os.Args[3:] because Args[0]=cloakpkg, Args[1]=command, Args[2]=configFile
	if len(os.Args) > 3 {
		if err := fs.Parse(os.Args[3:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Build availability cache once at startup
	registry := installer.GetRegistry()
	availableCache := make(map[string]bool)
	for name, inst := range registry {
		availableCache[name] = inst.Available()
	}
	availableCache["custom"] = true

	// Parse tags filter
	includeTags := make(map[string]bool)
	if *tagsFlag != "" {
		for _, tag := range strings.Split(*tagsFlag, ",") {
			includeTags[strings.TrimSpace(tag)] = true
		}
	}
	excludeTags := make(map[string]bool)
	if *excludeTagsFlag != "" {
		for _, tag := range strings.Split(*excludeTagsFlag, ",") {
			excludeTags[strings.TrimSpace(tag)] = true
		}
	}

	// Determine bundles to process
	var bundleNames []string
	requestedBundles := fs.Args()
	if len(requestedBundles) > 0 {
		// Verify requested bundles exist
		for _, name := range requestedBundles {
			if _, ok := cfg.Bundles[name]; !ok {
				fmt.Fprintf(os.Stderr, "Error: bundle %q not found in config\n", name)
				os.Exit(1)
			}
			bundleNames = append(bundleNames, name)
		}
	} else {
		// Default: all bundles in config
		for name := range cfg.Bundles {
			bundleNames = append(bundleNames, name)
		}
	}

	if command == "check" {
		fmt.Printf("Evaluating bundles in %s:\n\n", configFile)
		for _, name := range bundleNames {
			bundle := cfg.Bundles[name]

			// Apply tag filters
			if !matchTags(bundle.Tags, includeTags, excludeTags) {
				continue
			}

			// Select provider(s) to run/check based on settings priority list
			var selectedProviders []string
			priorityList := cfg.Settings.ProviderPriority

			if len(priorityList) > 0 {
				// Find the first matching available provider
				for _, provName := range priorityList {
					provConfig, defined := bundle.Providers[provName]
					if !defined {
						continue
					}

					isAvail := false
					if provName == "custom" {
						if provConfig.Detect == "" || installer.CheckCustom(provConfig) {
							isAvail = true
						} else {
							isAvail = true
						}
					} else {
						isAvail = availableCache[provName]
					}

					if isAvail {
						selectedProviders = append(selectedProviders, provName)
						break
					}
				}
			} else {
				// Priority list empty: execute all defined providers that are available
				for provName := range bundle.Providers {
					isAvail := false
					if provName == "custom" {
						isAvail = true
					} else {
						isAvail = availableCache[provName]
					}

					if isAvail {
						selectedProviders = append(selectedProviders, provName)
					}
				}
			}

			fmt.Printf("Bundle: %s\n", name)
			if bundle.Description != "" {
				fmt.Printf("  Description: %s\n", bundle.Description)
			}
			if len(bundle.Tags) > 0 {
				fmt.Printf("  Tags:        %s\n", strings.Join(bundle.Tags, ", "))
			}
			if len(selectedProviders) == 0 {
				fmt.Println("  Selected:    NONE (no defined providers are available on this system)")
			} else {
				fmt.Printf("  Selected:    %s\n", strings.Join(selectedProviders, ", "))
				for _, provName := range selectedProviders {
					provConfig := bundle.Providers[provName]
					if provName == "custom" {
						installed := installer.CheckCustom(provConfig)
						status := "not installed"
						if installed {
							status = "already installed"
						}
						fmt.Printf("    - custom   (%s)\n", status)
					} else {
						inst := registry[provName]
						fmt.Printf("    - %s packages:\n", provName)
						for _, p := range provConfig.Packages {
							status := "not installed"
							if inst.Installed(p) {
								status = "installed"
							}
							fmt.Printf("        %-15s : %s\n", p.Name, status)
						}
					}
				}
			}
			fmt.Println()
		}
	} else {
		// Pass 1: Accumulate built-in packages, repositories, and custom jobs
		builtinPackages := make(map[string][]config.Package)
		builtinRepos := make(map[string][]config.Repository)
		providerBundles := make(map[string][]string)
		type customJob struct {
			bundleName string
			provider   config.Provider
		}
		var customJobs []customJob

		for _, name := range bundleNames {
			bundle := cfg.Bundles[name]

			// Apply tag filters
			if !matchTags(bundle.Tags, includeTags, excludeTags) {
				if *verboseFlag {
					fmt.Printf("Skipping bundle %q (tag filter mismatch)\n", name)
				}
				continue
			}

			// Select provider(s) based on settings priority list
			var selectedProviders []string
			priorityList := cfg.Settings.ProviderPriority

			if len(priorityList) > 0 {
				for _, provName := range priorityList {
					_, defined := bundle.Providers[provName]
					if !defined {
						continue
					}

					isAvail := false
					if provName == "custom" {
						isAvail = true
					} else {
						isAvail = availableCache[provName]
					}

					if isAvail {
						selectedProviders = append(selectedProviders, provName)
						break
					}
				}
			} else {
				for provName := range bundle.Providers {
					isAvail := false
					if provName == "custom" {
						isAvail = true
					} else {
						isAvail = availableCache[provName]
					}

					if isAvail {
						selectedProviders = append(selectedProviders, provName)
					}
				}
			}

			if len(selectedProviders) == 0 {
				fmt.Printf("Skipping bundle %q: no available providers defined\n", name)
				continue
			}

			for _, provName := range selectedProviders {
				provConfig := bundle.Providers[provName]
				if provName == "custom" {
					customJobs = append(customJobs, customJob{bundleName: name, provider: provConfig})
				} else {
					builtinPackages[provName] = append(builtinPackages[provName], provConfig.Packages...)
					builtinRepos[provName] = append(builtinRepos[provName], provConfig.Repositories...)

					// Add to providerBundles, keeping it unique
					alreadyAdded := false
					for _, bName := range providerBundles[provName] {
						if bName == name {
							alreadyAdded = true
							break
						}
					}
					if !alreadyAdded {
						providerBundles[provName] = append(providerBundles[provName], name)
					}
				}
			}
		}

		// Determine execution order for built-ins
		var executionOrder []string
		priorityList := cfg.Settings.ProviderPriority
		for _, provName := range priorityList {
			if len(builtinPackages[provName]) > 0 {
				executionOrder = append(executionOrder, provName)
			}
		}
		for provName := range builtinPackages {
			found := false
			for _, ordered := range executionOrder {
				if ordered == provName {
					found = true
					break
				}
			}
			if !found {
				executionOrder = append(executionOrder, provName)
			}
		}

		// Run built-in installers in collated execution order
		for _, provName := range executionOrder {
			pkgs := builtinPackages[provName]
			repos := deduplicateRepos(builtinRepos[provName])
			inst := registry[provName]
			bundles := providerBundles[provName]

			if len(repos) > 0 && (command == "install" || command == "update" || command == "reinstall") {
				if err := inst.AddRepositories(*verboseFlag, *dryRunFlag, repos); err != nil {
					fmt.Fprintf(os.Stderr, "  Error adding repositories for provider %s: %v\n", provName, err)
					os.Exit(1)
				}
			}

			switch command {
			case "install":
				executeBuiltinAction("install", *verboseFlag, *dryRunFlag, provName, pkgs, bundles, cfg, inst.Install)
			case "uninstall":
				executeBuiltinAction("uninstall", *verboseFlag, *dryRunFlag, provName, pkgs, bundles, cfg, inst.Uninstall)
			case "update":
				executeBuiltinAction("update", *verboseFlag, *dryRunFlag, provName, pkgs, bundles, cfg, inst.Update)
			case "reinstall":
				executeBuiltinAction("uninstall", *verboseFlag, *dryRunFlag, provName, pkgs, bundles, cfg, inst.Uninstall)
				executeBuiltinAction("install", *verboseFlag, *dryRunFlag, provName, pkgs, bundles, cfg, inst.Install)
			}
		}

		// Run custom jobs
		for _, job := range customJobs {
			b := cfg.Bundles[job.bundleName]
			switch command {
			case "install":
				executeCustomAction("install", *verboseFlag, *dryRunFlag, job.bundleName, b, job.provider, installer.InstallCustom)
			case "uninstall":
				executeCustomAction("uninstall", *verboseFlag, *dryRunFlag, job.bundleName, b, job.provider, installer.UninstallCustom)
			case "update":
				executeCustomAction("update", *verboseFlag, *dryRunFlag, job.bundleName, b, job.provider, installer.UpdateCustom)
			case "reinstall":
				executeCustomAction("uninstall", *verboseFlag, *dryRunFlag, job.bundleName, b, job.provider, installer.UninstallCustom)
				executeCustomAction("install", *verboseFlag, *dryRunFlag, job.bundleName, b, job.provider, installer.InstallCustom)
			}
		}
	}
}

func matchTags(bundleTags []string, includeTags, excludeTags map[string]bool) bool {
	// Filter out if not in include list (only if include list is non-empty)
	if len(includeTags) > 0 {
		matched := false
		for _, tag := range bundleTags {
			if includeTags[tag] {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Filter out if in exclude list
	if len(excludeTags) > 0 {
		for _, tag := range bundleTags {
			if excludeTags[tag] {
				return false
			}
		}
	}

	return true
}

func deduplicateRepos(repos []config.Repository) []config.Repository {
	seen := make(map[string]bool)
	var unique []config.Repository
	for _, repo := range repos {
		if repo.Source == "" {
			continue
		}
		if !seen[repo.Source] {
			seen[repo.Source] = true
			unique = append(unique, repo)
		}
	}
	return unique
}

func runHook(verbose bool, dryRun bool, hookType string, bundleName string, hookCmd string) error {
	if hookCmd == "" {
		return nil
	}
	fmt.Printf("Running %s hook for bundle %q...\n", hookType, bundleName)
	if err := runner.RunShell(verbose, dryRun, hookCmd); err != nil {
		return fmt.Errorf("%s hook failed for bundle %q: %w", hookType, bundleName, err)
	}
	return nil
}

func executeBuiltinAction(command string, verbose bool, dryRun bool, provName string, pkgs []config.Package, bundles []string, cfg *config.Config, actionFunc func(bool, bool, []config.Package) error) {
	for _, bName := range bundles {
		b := cfg.Bundles[bName]
		if err := runPreHooks(HookOptions{Verbose: verbose, DryRun: dryRun, Command: command, ProvName: provName, BundleName: bName, Bundle: b}); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
	}

	actionStr := ""
	switch command {
	case "install":
		actionStr = "Installing"
	case "uninstall":
		actionStr = "Uninstalling"
	case "update":
		actionStr = "Updating"
	default:
		actionStr = "Processing"
	}
	fmt.Printf("%s packages for provider %q...\n", actionStr, provName)

	if err := actionFunc(verbose, dryRun, pkgs); err != nil {
		fmt.Fprintf(os.Stderr, "  Error %s built-in provider %s: %v\n", strings.ToLower(actionStr), provName, err)
		os.Exit(1)
	}

	for _, bName := range bundles {
		b := cfg.Bundles[bName]
		if err := runPostHooks(HookOptions{Verbose: verbose, DryRun: dryRun, Command: command, ProvName: provName, BundleName: bName, Bundle: b}); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func executeCustomAction(command string, verbose bool, dryRun bool, bundleName string, b config.Bundle, provider config.Provider, actionFunc func(bool, bool, config.Provider) error) {
	if err := runPreHooks(HookOptions{Verbose: verbose, DryRun: dryRun, Command: command, ProvName: "custom", BundleName: bundleName, Bundle: b}); err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		os.Exit(1)
	}

	actionStr := ""
	switch command {
	case "install":
		actionStr = "Installing"
	case "uninstall":
		actionStr = "Uninstalling"
	case "update":
		actionStr = "Updating"
	default:
		actionStr = "Processing"
	}
	fmt.Printf("%s custom provider for bundle %q...\n", actionStr, bundleName)

	if err := actionFunc(verbose, dryRun, provider); err != nil {
		fmt.Fprintf(os.Stderr, "  Error %s custom provider: %v\n", strings.ToLower(actionStr), err)
		os.Exit(1)
	}

	if err := runPostHooks(HookOptions{Verbose: verbose, DryRun: dryRun, Command: command, ProvName: "custom", BundleName: bundleName, Bundle: b}); err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		os.Exit(1)
	}
}

type HookOptions struct {
	Verbose    bool
	DryRun     bool
	Command    string
	ProvName   string
	BundleName string
	Bundle     config.Bundle
}

func runPreHooks(opts HookOptions) error {
	var bundleHook, provHook string
	switch opts.Command {
	case "install":
		if opts.Bundle.Hooks != nil {
			bundleHook = opts.Bundle.Hooks.PreInstall
		}
		if p, ok := opts.Bundle.Providers[opts.ProvName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PreInstall
		}
	case "uninstall":
		if opts.Bundle.Hooks != nil {
			bundleHook = opts.Bundle.Hooks.PreUninstall
		}
		if p, ok := opts.Bundle.Providers[opts.ProvName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PreUninstall
		}
	case "update":
		if opts.Bundle.Hooks != nil {
			bundleHook = opts.Bundle.Hooks.PreUpdate
		}
		if p, ok := opts.Bundle.Providers[opts.ProvName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PreUpdate
		}
	}

	if bundleHook != "" {
		if err := runHook(opts.Verbose, opts.DryRun, "pre_"+opts.Command, opts.BundleName, bundleHook); err != nil {
			return err
		}
	}
	if provHook != "" {
		if err := runHook(opts.Verbose, opts.DryRun, "pre_"+opts.Command+" ("+opts.ProvName+")", opts.BundleName, provHook); err != nil {
			return err
		}
	}
	return nil
}

func runPostHooks(opts HookOptions) error {
	var bundleHook, provHook string
	switch opts.Command {
	case "install":
		if opts.Bundle.Hooks != nil {
			bundleHook = opts.Bundle.Hooks.PostInstall
		}
		if p, ok := opts.Bundle.Providers[opts.ProvName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PostInstall
		}
	case "uninstall":
		if opts.Bundle.Hooks != nil {
			bundleHook = opts.Bundle.Hooks.PostUninstall
		}
		if p, ok := opts.Bundle.Providers[opts.ProvName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PostUninstall
		}
	case "update":
		if opts.Bundle.Hooks != nil {
			bundleHook = opts.Bundle.Hooks.PostUpdate
		}
		if p, ok := opts.Bundle.Providers[opts.ProvName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PostUpdate
		}
	}

	if provHook != "" {
		if err := runHook(opts.Verbose, opts.DryRun, "post_"+opts.Command+" ("+opts.ProvName+")", opts.BundleName, provHook); err != nil {
			return err
		}
	}
	if bundleHook != "" {
		if err := runHook(opts.Verbose, opts.DryRun, "post_"+opts.Command, opts.BundleName, bundleHook); err != nil {
			return err
		}
	}
	return nil
}
