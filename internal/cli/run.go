package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
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
		out := fs.String("out", "backup", "default output directory (relative to workspace root)")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if err := runInit(*chdir, *force, *out); err != nil {
			fmt.Fprintln(os.Stderr, "FAIL:", err)
			return 1
		}
		fmt.Fprintln(os.Stdout, "OK")
		return 0

	case "secret":
		fs := flag.NewFlagSet("secret", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		chdir := fs.String("C", "", "run as if started in this directory")
		reveal := fs.Bool("reveal", false, "print the secret value (unsafe)")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "secret requires subcommand: set | show")
			return 2
		}
		sub := fs.Arg(0)
		switch sub {
		case "set":
			if err := runSecretSet(*chdir, os.Stdin); err != nil {
				fmt.Fprintln(os.Stderr, "FAIL:", err)
				return 1
			}
			fmt.Fprintln(os.Stdout, "OK")
			return 0
		case "show":
			if err := runSecretShow(*chdir, *reveal, os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, "FAIL:", err)
				return 1
			}
			return 0
		default:
			fmt.Fprintln(os.Stderr, "unknown secret subcommand:", sub)
			return 2
		}

	case "auth":
		fs := flag.NewFlagSet("auth", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		chdir := fs.String("C", "", "run as if started in this directory")
		configPath := fs.String("c", "", "explicit config file path (advanced)")
		noBrowser := fs.Bool("no-browser", false, "do not auto-open the browser")
		host := fs.String("host", "127.0.0.1", "callback listen host")
		port := fs.Int("port", 18900, "callback listen port")
		callbackPath := fs.String("callback-path", "/callback", "callback path")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "auth requires subcommand: login")
			return 2
		}
		sub := fs.Arg(0)
		switch sub {
		case "login":
			if err := runAuthLogin(context.Background(), *chdir, *configPath, authLoginOptions{ListenHost: *host, Port: *port, CallbackPath: *callbackPath, NoBrowser: *noBrowser}, os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, "FAIL:", err)
				return 1
			}
			fmt.Fprintln(os.Stdout, "OK")
			return 0
		default:
			fmt.Fprintln(os.Stderr, "unknown auth subcommand:", sub)
			return 2
		}

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
	fmt.Fprintln(os.Stderr, "commands: init, secret, auth, pull, drive, validate")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "drive subcommands: roots, ls")
}
