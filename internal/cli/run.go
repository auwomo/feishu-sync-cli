package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

func Run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}

	switch args[0] {
	case "init":
		fs := flag.NewFlagSet("init", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		chdir := fs.String("C", "", "initialize workspace in this directory")
		force := fs.Bool("force", false, "overwrite existing .feishu-sync directory")
		appID := fs.String("app-id", "", "set app.id in generated config.yaml (e.g. cli_xxx)")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if err := runInit(*chdir, *force, initOptions{AppID: *appID}); err != nil {
			fmt.Fprintln(os.Stderr, "FAIL:", err)
			return 1
		}
		fmt.Fprintln(os.Stdout, "OK")
		fmt.Fprintln(os.Stderr, "Next steps:")
		fmt.Fprintln(os.Stderr, "  1) feishu-sync config")
		fmt.Fprintln(os.Stderr, "  2) feishu-sync login")
		fmt.Fprintln(os.Stderr, "  3) feishu-sync pull --dry-run   # preview")
		fmt.Fprintln(os.Stderr, "  4) feishu-sync pull            # export")
		return 0

	case "config":
		fs := flag.NewFlagSet("config", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		chdir := fs.String("C", "", "run as if started in this directory")
		configPath := fs.String("c", "", "explicit config file path (advanced)")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if err := runConfigWizard(*chdir, *configPath, os.Stdin, os.Stdout, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, "FAIL:", err)
			return 1
		}
		return 0

	case "login":
		fs := flag.NewFlagSet("login", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		chdir := fs.String("C", "", "run as if started in this directory")
		configPath := fs.String("c", "", "explicit config file path (advanced)")
		noBrowser := fs.Bool("no-browser", false, "do not auto-open the browser")
		verbose := fs.Bool("verbose", false, "verbose output")
		timeout := fs.Duration("timeout", 2*time.Minute, "timeout for local callback flow")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if err := runAuthLogin(context.Background(), *chdir, *configPath, authLoginOptions{NoBrowser: *noBrowser, Verbose: *verbose, Timeout: *timeout}, os.Stdout, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, "FAIL:", err)
			return 1
		}
		fmt.Fprintln(os.Stdout, "OK")
		return 0

	case "pull":
		fs := flag.NewFlagSet("pull", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		chdir := fs.String("C", "", "run as if started in this directory")
		configPath := fs.String("c", "", "explicit config file path (advanced)")
		dryRun := fs.Bool("dry-run", false, "discover scope and print a manifest, without downloading")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if err := runPull(*chdir, *configPath, *dryRun); err != nil {
			fmt.Fprintln(os.Stderr, "FAIL:", err)
			return 1
		}
		return 0

	case "wiki":
		fs := flag.NewFlagSet("wiki", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		chdir := fs.String("C", "", "run as if started in this directory")
		configPath := fs.String("c", "", "explicit config file path (advanced)")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "wiki requires subcommand: ls")
			return 2
		}
		sub := fs.Arg(0)
		switch sub {
		case "ls":
			opt, _, err := parseWikiLsFlags(fs.Args()[1:])
			if err != nil {
				return 2
			}
			if err := runWikiLs(context.Background(), *chdir, *configPath, opt, os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, "FAIL:", err)
				return 1
			}
			return 0
		default:
			fmt.Fprintln(os.Stderr, "unknown wiki subcommand:", sub)
			return 2
		}

	case "drive":
		fs := flag.NewFlagSet("drive", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		chdir := fs.String("C", "", "run as if started in this directory")
		configPath := fs.String("c", "", "explicit config file path (advanced)")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "drive requires subcommand: roots | ls")
			return 2
		}
		sub := fs.Arg(0)
		switch sub {
		case "roots":
			if err := runDriveRoots(context.Background(), *chdir, *configPath, os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, "FAIL:", err)
				return 1
			}
			return 0
		case "ls":
			lsfs := flag.NewFlagSet("drive ls", flag.ContinueOnError)
			lsfs.SetOutput(os.Stderr)
			folder := lsfs.String("folder", "", "folder token to list")
			depth := lsfs.Int("depth", 1, "recursion depth (0=only this folder)")
			if err := lsfs.Parse(fs.Args()[1:]); err != nil {
				return 2
			}
			if err := runDriveLs(context.Background(), *chdir, *configPath, driveLsOptions{FolderToken: *folder, Depth: *depth}, os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, "FAIL:", err)
				return 1
			}
			return 0
		default:
			fmt.Fprintln(os.Stderr, "unknown drive subcommand:", sub)
			return 2
		}

	case "validate":
		fs := flag.NewFlagSet("validate", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		chdir := fs.String("C", "", "run as if started in this directory")
		configPath := fs.String("c", "", "explicit config file path (advanced)")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if err := runValidate(*chdir, *configPath); err != nil {
			fmt.Fprintln(os.Stderr, "FAIL:", err)
			return 1
		}
		fmt.Fprintln(os.Stdout, "OK")
		return 0

	case "-h", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "feishu-sync <command> [args]")
	fmt.Fprintln(os.Stderr, "commands: init, config, login, pull, drive, wiki, validate")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "drive subcommands: roots, ls")
	fmt.Fprintln(os.Stderr, "wiki subcommands: ls")
}
