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

type customJob struct {
	bundleName string
	provider   config.Provider
}

type executionPlan struct {
	builtinPackages map[string][]config.Package
	builtinRepos    map[string][]config.Repository
	providerBundles map[string][]string
	executionOrder  []string
	customJobs      []customJob
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
		handleCheckCommand(configFile, bundleNames, cfg, includeTags, excludeTags, registry, availableCache)
	} else {
		handleActionCommand(command, cfg, bundleNames, includeTags, excludeTags, registry, availableCache, *verboseFlag, *dryRunFlag)
	}
}

func selectProvidersForBundle(bundle config.Bundle, priorityList []string, availableCache map[string]bool) []string {
	var selectedProviders []string

	if len(priorityList) > 0 {
		// Find the first matching available provider
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

	return selectedProviders
}

func printBundleStatus(name string, bundle config.Bundle, selectedProviders []string, registry map[string]installer.Installer) {
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

func handleCheckCommand(configFile string, bundleNames []string, cfg *config.Config, includeTags, excludeTags map[string]bool, registry map[string]installer.Installer, availableCache map[string]bool) {
	fmt.Printf("Evaluating bundles in %s:\n\n", configFile)
	for _, name := range bundleNames {
		bundle := cfg.Bundles[name]

		// Apply tag filters
		if !matchTags(bundle.Tags, includeTags, excludeTags) {
			continue
		}

		// Select provider(s) to run/check based on settings priority list
		selectedProviders := selectProvidersForBundle(bundle, cfg.Settings.ProviderPriority, availableCache)
		printBundleStatus(name, bundle, selectedProviders, registry)
	}
}

func buildExecutionPlan(cfg *config.Config, bundleNames []string, includeTags, excludeTags map[string]bool, availableCache map[string]bool, verbose bool) executionPlan {
	plan := executionPlan{
		builtinPackages: make(map[string][]config.Package),
		builtinRepos:    make(map[string][]config.Repository),
		providerBundles: make(map[string][]string),
	}

	for _, name := range bundleNames {
		bundle := cfg.Bundles[name]

		// Apply tag filters
		if !matchTags(bundle.Tags, includeTags, excludeTags) {
			if verbose {
				fmt.Printf("Skipping bundle %q (tag filter mismatch)\n", name)
			}
			continue
		}

		// Select provider(s) based on settings priority list
		selectedProviders := selectProvidersForBundle(bundle, cfg.Settings.ProviderPriority, availableCache)

		if len(selectedProviders) == 0 {
			fmt.Printf("Skipping bundle %q: no available providers defined\n", name)
			continue
		}

		for _, provName := range selectedProviders {
			provConfig := bundle.Providers[provName]
			if provName == "custom" {
				// Avoid duplicate custom jobs for same provider definition reference
				// Need to ensure we only append if not already checking
				alreadyAdded := false
				for _, c := range plan.customJobs {
					if c.bundleName == name {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					plan.customJobs = append(plan.customJobs, customJob{bundleName: name, provider: provConfig})
				}
			} else {
				plan.builtinPackages[provName] = append(plan.builtinPackages[provName], provConfig.Packages...)
				plan.builtinRepos[provName] = append(plan.builtinRepos[provName], provConfig.Repositories...)

				// Add to providerBundles, keeping it unique
				alreadyAdded := false
				for _, bName := range plan.providerBundles[provName] {
					if bName == name {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					plan.providerBundles[provName] = append(plan.providerBundles[provName], name)
				}
			}
		}
	}

	// Determine execution order for built-ins
	for _, provName := range cfg.Settings.ProviderPriority {
		if len(plan.builtinPackages[provName]) > 0 {
			plan.executionOrder = append(plan.executionOrder, provName)
		}
	}
	for provName := range plan.builtinPackages {
		found := false
		for _, ordered := range plan.executionOrder {
			if ordered == provName {
				found = true
				break
			}
		}
		if !found {
			plan.executionOrder = append(plan.executionOrder, provName)
		}
	}

	return plan
}

func runBuiltinAction(command string, cfg *config.Config, provName string, bundles []string, pkgs []config.Package, inst installer.Installer, verbose, dryRun bool) {
	if command == "reinstall" {
		for _, bName := range bundles {
			b := cfg.Bundles[bName]
			if err := runPreHooks(verbose, dryRun, "uninstall", provName, bName, b); err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Reinstalling packages for provider %q...\n", provName)
		if err := inst.Uninstall(verbose, dryRun, pkgs); err != nil {
			fmt.Fprintf(os.Stderr, "  Error uninstalling built-in provider %s: %v\n", provName, err)
			os.Exit(1)
		}

		for _, bName := range bundles {
			b := cfg.Bundles[bName]
			if err := runPostHooks(verbose, dryRun, "uninstall", provName, bName, b); err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				os.Exit(1)
			}
		}

		for _, bName := range bundles {
			b := cfg.Bundles[bName]
			if err := runPreHooks(verbose, dryRun, "install", provName, bName, b); err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				os.Exit(1)
			}
		}

		if err := inst.Install(verbose, dryRun, pkgs); err != nil {
			fmt.Fprintf(os.Stderr, "  Error installing built-in provider %s: %v\n", provName, err)
			os.Exit(1)
		}

		for _, bName := range bundles {
			b := cfg.Bundles[bName]
			if err := runPostHooks(verbose, dryRun, "install", provName, bName, b); err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				os.Exit(1)
			}
		}
		return
	}

	for _, bName := range bundles {
		b := cfg.Bundles[bName]
		if err := runPreHooks(verbose, dryRun, command, provName, bName, b); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
	}

	switch command {
	case "install":
		fmt.Printf("Installing packages for provider %q...\n", provName)
		if err := inst.Install(verbose, dryRun, pkgs); err != nil {
			fmt.Fprintf(os.Stderr, "  Error installing built-in provider %s: %v\n", provName, err)
			os.Exit(1)
		}
	case "uninstall":
		fmt.Printf("Uninstalling packages for provider %q...\n", provName)
		if err := inst.Uninstall(verbose, dryRun, pkgs); err != nil {
			fmt.Fprintf(os.Stderr, "  Error uninstalling built-in provider %s: %v\n", provName, err)
			os.Exit(1)
		}
	case "update":
		fmt.Printf("Updating packages for provider %q...\n", provName)
		if err := inst.Update(verbose, dryRun, pkgs); err != nil {
			fmt.Fprintf(os.Stderr, "  Error updating built-in provider %s: %v\n", provName, err)
			os.Exit(1)
		}
	}

	for _, bName := range bundles {
		b := cfg.Bundles[bName]
		if err := runPostHooks(verbose, dryRun, command, provName, bName, b); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runCustomAction(command string, cfg *config.Config, job customJob, verbose, dryRun bool) {
	b := cfg.Bundles[job.bundleName]

	if command == "reinstall" {
		if err := runPreHooks(verbose, dryRun, "uninstall", "custom", job.bundleName, b); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Reinstalling custom provider for bundle %q...\n", job.bundleName)
		if err := installer.UninstallCustom(verbose, dryRun, job.provider); err != nil {
			fmt.Fprintf(os.Stderr, "  Error uninstalling custom provider: %v\n", err)
			os.Exit(1)
		}

		if err := runPostHooks(verbose, dryRun, "uninstall", "custom", job.bundleName, b); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}

		if err := runPreHooks(verbose, dryRun, "install", "custom", job.bundleName, b); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}

		if err := installer.InstallCustom(verbose, dryRun, job.provider); err != nil {
			fmt.Fprintf(os.Stderr, "  Error installing custom provider: %v\n", err)
			os.Exit(1)
		}

		if err := runPostHooks(verbose, dryRun, "install", "custom", job.bundleName, b); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := runPreHooks(verbose, dryRun, command, "custom", job.bundleName, b); err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "install":
		fmt.Printf("Installing custom provider for bundle %q...\n", job.bundleName)
		if err := installer.InstallCustom(verbose, dryRun, job.provider); err != nil {
			fmt.Fprintf(os.Stderr, "  Error installing custom provider: %v\n", err)
			os.Exit(1)
		}
	case "uninstall":
		fmt.Printf("Uninstalling custom provider for bundle %q...\n", job.bundleName)
		if err := installer.UninstallCustom(verbose, dryRun, job.provider); err != nil {
			fmt.Fprintf(os.Stderr, "  Error uninstalling custom provider: %v\n", err)
			os.Exit(1)
		}
	case "update":
		fmt.Printf("Updating custom provider for bundle %q...\n", job.bundleName)
		if err := installer.UpdateCustom(verbose, dryRun, job.provider); err != nil {
			fmt.Fprintf(os.Stderr, "  Error updating custom provider: %v\n", err)
			os.Exit(1)
		}
	}

	if err := runPostHooks(verbose, dryRun, command, "custom", job.bundleName, b); err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		os.Exit(1)
	}
}

func handleActionCommand(command string, cfg *config.Config, bundleNames []string, includeTags, excludeTags map[string]bool, registry map[string]installer.Installer, availableCache map[string]bool, verbose, dryRun bool) {
	plan := buildExecutionPlan(cfg, bundleNames, includeTags, excludeTags, availableCache, verbose)

	// Run built-in installers in collated execution order
	for _, provName := range plan.executionOrder {
		pkgs := plan.builtinPackages[provName]
		repos := deduplicateRepos(plan.builtinRepos[provName])
		inst := registry[provName]
		bundles := plan.providerBundles[provName]

		if len(repos) > 0 && (command == "install" || command == "update" || command == "reinstall") {
			if err := inst.AddRepositories(verbose, dryRun, repos); err != nil {
				fmt.Fprintf(os.Stderr, "  Error adding repositories for provider %s: %v\n", provName, err)
				os.Exit(1)
			}
		}

		runBuiltinAction(command, cfg, provName, bundles, pkgs, inst, verbose, dryRun)
	}

	// Run custom jobs
	for _, job := range plan.customJobs {
		runCustomAction(command, cfg, job, verbose, dryRun)
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

func runPreHooks(verbose bool, dryRun bool, command string, provName string, bundleName string, b config.Bundle) error {
	var bundleHook, provHook string
	switch command {
	case "install":
		if b.Hooks != nil {
			bundleHook = b.Hooks.PreInstall
		}
		if p, ok := b.Providers[provName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PreInstall
		}
	case "uninstall":
		if b.Hooks != nil {
			bundleHook = b.Hooks.PreUninstall
		}
		if p, ok := b.Providers[provName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PreUninstall
		}
	case "update":
		if b.Hooks != nil {
			bundleHook = b.Hooks.PreUpdate
		}
		if p, ok := b.Providers[provName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PreUpdate
		}
	}

	if bundleHook != "" {
		if err := runHook(verbose, dryRun, "pre_" + command, bundleName, bundleHook); err != nil {
			return err
		}
	}
	if provHook != "" {
		if err := runHook(verbose, dryRun, "pre_" + command + " (" + provName + ")", bundleName, provHook); err != nil {
			return err
		}
	}
	return nil
}

func runPostHooks(verbose bool, dryRun bool, command string, provName string, bundleName string, b config.Bundle) error {
	var bundleHook, provHook string
	switch command {
	case "install":
		if b.Hooks != nil {
			bundleHook = b.Hooks.PostInstall
		}
		if p, ok := b.Providers[provName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PostInstall
		}
	case "uninstall":
		if b.Hooks != nil {
			bundleHook = b.Hooks.PostUninstall
		}
		if p, ok := b.Providers[provName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PostUninstall
		}
	case "update":
		if b.Hooks != nil {
			bundleHook = b.Hooks.PostUpdate
		}
		if p, ok := b.Providers[provName]; ok && p.Hooks != nil {
			provHook = p.Hooks.PostUpdate
		}
	}

	if provHook != "" {
		if err := runHook(verbose, dryRun, "post_" + command + " (" + provName + ")", bundleName, provHook); err != nil {
			return err
		}
	}
	if bundleHook != "" {
		if err := runHook(verbose, dryRun, "post_" + command, bundleName, bundleHook); err != nil {
			return err
		}
	}
	return nil
}
