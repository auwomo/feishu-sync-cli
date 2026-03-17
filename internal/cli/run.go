package cli

import (
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
		force := fs.Bool("force", false, "overwrite existing .feishu-sync directory")
		out := fs.String("out", "backup", "default output directory (relative to workspace root)")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if err := runInit(*force, *out); err != nil {
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
	fmt.Fprintln(os.Stderr, "commands: init, pull, validate")
}
