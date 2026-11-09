package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/agentic-workflow/awm/internal/awm"
)

type console struct{ in *bufio.Reader }

func newConsole() *console { return &console{in: bufio.NewReader(os.Stdin)} }

func (c *console) Confirm(prompt string, defaultYes bool) bool {
	choice := "y/N"
	if defaultYes {
		choice = "Y/n"
	}
	fmt.Printf("%s [%s]: ", prompt, choice)
	value, _ := c.in.ReadString('\n')
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return defaultYes
	}
	return value == "y" || value == "yes"
}

func (c *console) SelectSkills(candidates []awm.BundledSkill) ([]string, error) {
	fmt.Println("Choose skills: Y = all, N = none, S = select IDs.")
	fmt.Print("Install missing AWM skills? [Y/n/s]: ")
	value, err := c.in.ReadString('\n')
	if err != nil && !errors.Is(err, os.ErrClosed) {
		return nil, err
	}
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "y" || value == "yes" {
		return allSkillIDs(candidates), nil
	}
	if value == "n" || value == "no" {
		return nil, nil
	}
	if value != "s" && value != "select" {
		return nil, errors.New("enter Y, N, or S")
	}
	fmt.Print("Enter comma-separated skill IDs: ")
	value, err = c.in.ReadString('\n')
	if err != nil && !errors.Is(err, os.ErrClosed) {
		return nil, err
	}
	allowed := map[string]bool{}
	for _, item := range candidates {
		allowed[item.ID] = true
	}
	var selected []string
	for _, part := range strings.Split(value, ",") {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		if !allowed[id] {
			return nil, fmt.Errorf("unknown or unavailable skill: %s", id)
		}
		selected = append(selected, id)
	}
	return selected, nil
}

func allSkillIDs(items []awm.BundledSkill) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.ID)
	}
	return out
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"install"}
	}
	var err error
	switch args[0] {
	case "install":
		err = runInstall(args[1:])
	case "detect":
		err = runDetect(args[1:])
	case "doctor":
		err = runDoctor(args[1:])
	case "status":
		err = runStatus(args[1:])
	case "skills":
		err = runSkills(args[1:])
	case "plugins":
		err = runPlugins(args[1:])
	case "memory":
		err = runMemory(args[1:])
	case "workspace":
		err = runWorkspace(args[1:])
	case "builder":
		err = runBuilder(args[1:])
	case "rollback":
		err = runRollback(args[1:])
	case "update":
		err = runUpdate(args[1:])
	case "uninstall":
		err = runUninstall(args[1:])
	case "version", "--version", "-v":
		fmt.Printf("AWM %s (%s)\n", awm.Version, awm.RuntimePlatform())
	case "help", "--help", "-h":
		usage()
	default:
		usage()
		err = fmt.Errorf("unknown command %q", args[0])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "awm:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf(`Agentic Workflow Manager %s

Commands:
  awm install                         Interactive installation wizard
  awm detect                          Detect supported harnesses
  awm doctor                          Read-only environment audit
  awm status                          Show current AWM installation
  awm skills audit                    Audit bundled and external skills
  awm plugins audit                   Audit plugin markers
  awm memory doctor                   Inspect memory integration paths
  awm workspace init                  Create a repository .ai-memory.toml marker
  awm builder render                  Create an approved task-scoped Builder profile
  awm builder rollback                Restore a task-scoped Builder profile
  awm rollback                        Restore the previous AWM-managed files
  awm update                          Re-apply the installed AWM payload
  awm uninstall                       Remove AWM-managed files
  awm version                         Show version and platform

Install flags:
  --harness <ids>                     Comma-separated harness IDs
  --allow-undetected                  Allow an explicit harness without its binary
  --yes                               Select all detected harnesses and missing bundled skills
  --skip-skills                       Do not offer bundled skill installation
  --force                             Replace unmanaged target files after explicit flag
  --dry-run                           Preview writes without changing files
  --home <path>                       Use another home directory for isolated testing

Update flags:
  --harness <ids>                     Override the previous harness selection
  --force                             Replace files changed after installation
  --skip-skills                       Do not update bundled skills
  --dry-run                           Preview writes and removals
  --home <path>                       Use another home directory

Uninstall flags:
  --force                             Remove files changed after installation
  --remove-cli                        Also remove the awm binary
  --yes                               Skip confirmation
  --dry-run                           Preview removals
  --home <path>                       Use another home directory

Harness IDs: opencode, claude-code, codex-cli, omp
`, awm.Version)
}

func runInstall(args []string) error {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	harnessArg := flags.String("harness", "", "comma-separated harness IDs")
	allowUndetected := flags.Bool("allow-undetected", false, "allow explicitly selected undetected harnesses")
	yes := flags.Bool("yes", false, "select all detected harnesses and all missing bundled skills")
	skipSkills := flags.Bool("skip-skills", false, "skip bundled skill installation")
	force := flags.Bool("force", false, "replace unmanaged files")
	dryRun := flags.Bool("dry-run", false, "preview only")
	home := flags.String("home", "", "override home directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	console := newConsole()
	baseHome := *home
	if baseHome == "" {
		baseHome = awm.HomeDir()
	}
	detected, err := awm.DetectHarnesses(baseHome)
	if err != nil {
		return err
	}
	var ids []string
	if strings.TrimSpace(*harnessArg) != "" {
		for _, item := range strings.Split(*harnessArg, ",") {
			if strings.TrimSpace(item) != "" {
				ids = append(ids, strings.TrimSpace(item))
			}
		}
	} else if *yes {
		for _, item := range detected {
			if item.Detected {
				ids = append(ids, item.ID)
			}
		}
	} else {
		ids, err = chooseHarnesses(console, detected)
		if err != nil {
			return err
		}
	}
	if len(ids) == 0 {
		return errors.New("no harness selected")
	}
	fmt.Println("\nSelected harnesses:", strings.Join(ids, ", "))
	for _, id := range ids {
		for _, item := range detected {
			if item.ID == id {
				fmt.Printf("  %s: %s\n", id, item.Capabilities["builder_exact_write_scope"])
			}
		}
	}
	if !*yes && !console.Confirm("Install AWM agents, Core, policies, and registries?", true) {
		return errors.New("installation cancelled")
	}
	result, err := awm.RunInstall(awm.InstallOptions{
		HarnessIDs: ids, Force: *force, Yes: *yes, DryRun: *dryRun,
		SkipSkills: *skipSkills, AllowUndetected: *allowUndetected, Home: *home,
	}, console)
	if err != nil {
		return err
	}
	if *dryRun {
		fmt.Println("\nDry run complete. No files were changed.")
	} else {
		fmt.Println("\nAWM installation complete.")
	}
	fmt.Printf("Harnesses: %s\nFiles written or planned: %d\nFiles skipped: %d\n", strings.Join(result.State.Selected, ", "), len(result.Written), len(result.Skipped))
	if !*dryRun {
		fmt.Println("Run `awm doctor` to review the installation.")
	}
	return nil
}

func chooseHarnesses(c *console, items []awm.DetectedHarness) ([]string, error) {
	fmt.Println("Supported harnesses:")
	for index, item := range items {
		status := "not detected"
		if item.Detected {
			status = "detected"
		}
		fmt.Printf("  %d. %-12s %s", index+1, item.ID, status)
		if item.Version != "" {
			fmt.Printf(" — %s", item.Version)
		}
		fmt.Println()
	}
	fmt.Print("Choose one or more numbers, or `all`: ")
	value, err := c.in.ReadString('\n')
	if err != nil {
		return nil, err
	}
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "all") {
		var ids []string
		for _, item := range items {
			if item.Detected {
				ids = append(ids, item.ID)
			}
		}
		return ids, nil
	}
	seen := map[string]bool{}
	var ids []string
	for _, part := range strings.Split(value, ",") {
		number, parseErr := strconv.Atoi(strings.TrimSpace(part))
		if parseErr != nil || number < 1 || number > len(items) {
			return nil, fmt.Errorf("invalid harness selection %q", part)
		}
		item := items[number-1]
		if !item.Detected {
			return nil, fmt.Errorf("%s is not detected; use --harness %s --allow-undetected to continue", item.ID, item.ID)
		}
		if !seen[item.ID] {
			seen[item.ID] = true
			ids = append(ids, item.ID)
		}
	}
	return ids, nil
}

func runDetect(args []string) error {
	flags := flag.NewFlagSet("detect", flag.ContinueOnError)
	home := flags.String("home", "", "home override")
	if err := flags.Parse(args); err != nil {
		return err
	}
	items, err := awm.DetectHarnesses(homeOr(*home))
	if err != nil {
		return err
	}
	for _, item := range items {
		status := "NOT DETECTED"
		if item.Detected {
			status = "DETECTED"
		}
		fmt.Printf("%-12s %-12s %s\n", item.ID, status, item.Version)
		if item.Evidence != "" {
			fmt.Printf("  %s\n", item.Evidence)
		}
	}
	return nil
}

func runDoctor(args []string) error {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	home := flags.String("home", "", "home override")
	if err := flags.Parse(args); err != nil {
		return err
	}
	report, err := awm.AuditReport(homeOr(*home))
	if err != nil {
		return err
	}
	fmt.Print(report)
	return nil
}

func runStatus(args []string) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	home := flags.String("home", "", "home override")
	if err := flags.Parse(args); err != nil {
		return err
	}
	state, err := awm.LoadInstallState(homeOr(*home))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("AWM is not installed.")
			return nil
		}
		return err
	}
	fmt.Printf("AWM %s installed at %s\nHarnesses: %s\nManaged files: %d\nBackup: %s\n", state.Version, state.InstalledAt.Format("2006-01-02 15:04:05 MST"), strings.Join(state.Selected, ", "), len(state.ManagedFiles), state.BackupRoot)
	if state.Version != awm.Version {
		fmt.Printf("CLI version %s differs from installed config %s. Run `awm update`.\n", awm.Version, state.Version)
	}
	return nil
}

func runSkills(args []string) error {
	if len(args) == 0 || args[0] != "audit" {
		return errors.New("usage: awm skills audit [--home path]")
	}
	flags := flag.NewFlagSet("skills audit", flag.ContinueOnError)
	home := flags.String("home", "", "home override")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	harnesses, err := awm.DetectHarnesses(homeOr(*home))
	if err != nil {
		return err
	}
	bundled, external, err := awm.AuditSkills(harnesses, homeOr(*home))
	if err != nil {
		return err
	}
	printSkillGroup("Bundled", bundled)
	printSkillGroup("External", external)
	fmt.Println("External skills are never installed without a trusted immutable source and Human selection.")
	return nil
}

func printSkillGroup(title string, items []awm.SkillFinding) {
	fmt.Println(title + " skills:")
	for _, item := range items {
		status := "MISSING"
		if item.Installed {
			status = "FOUND"
		}
		fmt.Printf("  %-28s %-8s", item.ID, status)
		if item.Version != "" {
			fmt.Printf(" version=%s", item.Version)
		} else if item.Installed {
			fmt.Print(" version=unknown")
		}
		if item.ExpectedVersion != "" && item.VersionStatus != "match" {
			fmt.Printf(" expected=%s (%s)", item.ExpectedVersion, item.VersionStatus)
		}
		if len(item.MissingDependencies) > 0 {
			fmt.Printf(" missing-deps=%s", strings.Join(item.MissingDependencies, ","))
		}
		if len(item.Compatibility) > 0 {
			fmt.Printf(" expected-compatible=%s", strings.Join(item.Compatibility, ","))
		}
		if item.CompatibilityStatus != "" {
			fmt.Printf(" compatibility=%s", item.CompatibilityStatus)
		}
		if len(item.ActualCompatibility) > 0 {
			fmt.Printf(" declared-compatible=%s", strings.Join(item.ActualCompatibility, ","))
		}
		if len(item.UnsupportedHarnesses) > 0 {
			fmt.Printf(" unsupported-detected=%s", strings.Join(item.UnsupportedHarnesses, ","))
		}
		if item.Path != "" {
			fmt.Printf(" path=%s", item.Path)
		}
		if item.Reason != "" && !item.Installed {
			fmt.Printf(" — %s", item.Reason)
		}
		fmt.Println()
	}
}

func runPlugins(args []string) error {
	if len(args) == 0 || args[0] != "audit" {
		return errors.New("usage: awm plugins audit [--home path]")
	}
	flags := flag.NewFlagSet("plugins audit", flag.ContinueOnError)
	home := flags.String("home", "", "home override")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	items, err := awm.AuditPlugins(homeOr(*home))
	if err != nil {
		return err
	}
	awm.PrintPluginFindings(items)
	fmt.Println("Plugin installation and configuration remain under Human control.")
	return nil
}

func runMemory(args []string) error {
	if len(args) == 0 || args[0] != "doctor" {
		return errors.New("usage: awm memory doctor [--home path]")
	}
	flags := flag.NewFlagSet("memory doctor", flag.ContinueOnError)
	home := flags.String("home", "", "home override")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	fmt.Println("Memory integration paths:")
	for _, item := range awm.MemoryStatus(homeOr(*home)) {
		fmt.Println("  " + item)
	}
	fmt.Println("AWM does not change ai-memory capture mode, auto-improve, scheduler, or stored memories.")
	return nil
}

func runWorkspace(args []string) error {
	if len(args) == 0 || args[0] != "init" {
		return errors.New("usage: awm workspace init --repo . --workspace <id> [--force]")
	}
	flags := flag.NewFlagSet("workspace init", flag.ContinueOnError)
	repo := flags.String("repo", ".", "repository path")
	workspace := flags.String("workspace", "", "workspace identity")
	force := flags.Bool("force", false, "replace an existing marker")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	path, err := awm.InitRepositoryMemory(*repo, *workspace, *force)
	if err != nil {
		return err
	}
	fmt.Println("Created repository memory marker:", path)
	fmt.Println("Review the workspace ID before committing this file.")
	return nil
}

func runBuilder(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: awm builder render|rollback")
	}
	if args[0] == "rollback" {
		flags := flag.NewFlagSet("builder rollback", flag.ContinueOnError)
		taskID := flags.String("task-id", "", "Builder task ID")
		home := flags.String("home", "", "home override")
		force := flags.Bool("force", false, "restore files changed after task profile creation")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if *taskID == "" {
			return errors.New("--task-id is required")
		}
		restored, skipped, err := awm.RollbackBuilderTask(homeOr(*home), *taskID, *force)
		if err != nil {
			return err
		}
		fmt.Printf("Restored: %d\nSkipped: %d\n", len(restored), len(skipped))
		for _, item := range skipped {
			fmt.Println("  SKIPPED:", item)
		}
		return nil
	}
	if args[0] != "render" {
		return errors.New("usage: awm builder render|rollback")
	}
	flags := flag.NewFlagSet("builder render", flag.ContinueOnError)
	task := flags.String("task", "", "Task Envelope JSON file")
	harness := flags.String("harness", "", "opencode or claude-code")
	approved := flags.Bool("human-approved", false, "confirm Human approval for Builder activation")
	force := flags.Bool("force", false, "replace existing task profile and allowlist")
	home := flags.String("home", "", "home override")
	out := flags.String("out", "", "optional Builder profile output path")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if *task == "" || *harness == "" {
		return errors.New("--task and --harness are required")
	}
	path, err := awm.RenderApprovedBuilderTask(*task, *harness, homeOr(*home), *out, *approved, *force)
	if err != nil {
		return err
	}
	fmt.Println("Task-scoped Builder profile created:", path)
	fmt.Println("The selected adapter still requires isolated worktree and Human-approved shell commands.")
	return nil
}

func runRollback(args []string) error {
	flags := flag.NewFlagSet("rollback", flag.ContinueOnError)
	home := flags.String("home", "", "home override")
	force := flags.Bool("force", false, "restore changed files anyway")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if !*force {
		fmt.Println("Rollback checks file hashes and skips files changed after installation.")
	}
	restored, skipped, err := awm.Rollback(homeOr(*home), *force)
	if err != nil {
		return err
	}
	fmt.Printf("Restored: %d\nSkipped: %d\n", len(restored), len(skipped))
	for _, item := range skipped {
		fmt.Println("  SKIPPED:", item)
	}
	return nil
}

func runUpdate(args []string) error {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	harnessArg := flags.String("harness", "", "comma-separated harness IDs")
	force := flags.Bool("force", false, "replace files changed after installation")
	dryRun := flags.Bool("dry-run", false, "preview only")
	skipSkills := flags.Bool("skip-skills", false, "skip bundled skill updates")
	home := flags.String("home", "", "home override")
	if err := flags.Parse(args); err != nil {
		return err
	}
	baseHome := homeOr(*home)
	previous, err := awm.LoadInstallState(baseHome)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("AWM is not installed; run `awm install`")
		}
		return err
	}
	ids := previous.Selected
	if strings.TrimSpace(*harnessArg) != "" {
		ids = nil
		for _, item := range strings.Split(*harnessArg, ",") {
			id := strings.TrimSpace(item)
			if id != "" {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return errors.New("previous installation has no selected harnesses; pass --harness")
	}
	result, err := awm.RunInstall(awm.InstallOptions{
		HarnessIDs: ids, AllowUndetected: true, Force: *force, Yes: true,
		DryRun: *dryRun, SkipSkills: *skipSkills, PruneStale: true, Home: *home,
	}, newConsole())
	if err != nil {
		return err
	}
	if *dryRun {
		fmt.Println("\nDry run complete. No files were changed.")
	} else {
		fmt.Printf("\nAWM updated: %s -> %s\n", previous.Version, awm.Version)
	}
	fmt.Printf("Harnesses: %s\nFiles written or planned: %d\nFiles removed: %d\nFiles skipped: %d\n", strings.Join(result.State.Selected, ", "), len(result.Written), len(result.Removed), len(result.Skipped))
	if !*dryRun {
		fmt.Println("The awm CLI binary is unchanged. Re-run the installer to update it.")
		fmt.Println("Run `awm doctor` to review the installation.")
	}
	return nil
}

func runUninstall(args []string) error {
	flags := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	home := flags.String("home", "", "home override")
	force := flags.Bool("force", false, "remove files changed after installation")
	dryRun := flags.Bool("dry-run", false, "preview only")
	removeCLI := flags.Bool("remove-cli", false, "also remove the awm binary")
	yes := flags.Bool("yes", false, "skip confirmation")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if !*yes && !*dryRun {
		prompt := "Remove AWM-managed files and " + filepath.Join(homeOr(*home), ".agentic-workflow") + "?"
		if !newConsole().Confirm(prompt, false) {
			return errors.New("uninstall cancelled")
		}
	}
	result, err := awm.Uninstall(awm.UninstallOptions{Home: *home, Force: *force, DryRun: *dryRun, RemoveCLI: *removeCLI})
	if err != nil {
		return err
	}
	if len(result.Removed) == 0 && len(result.Skipped) == 0 {
		fmt.Println("AWM is not installed.")
		return nil
	}
	if *dryRun {
		fmt.Println("\nDry run complete. No files were changed.")
	}
	fmt.Printf("Removed: %d\nSkipped: %d\n", len(result.Removed), len(result.Skipped))
	for _, item := range result.Skipped {
		fmt.Println("  SKIPPED:", item)
	}
	if len(result.Skipped) > 0 {
		fmt.Println("Some files were skipped. Re-run with --force after reviewing them.")
	}
	return nil
}

func homeOr(value string) string {
	if strings.TrimSpace(value) == "" {
		return awm.HomeDir()
	}
	return filepath.Clean(value)
}
