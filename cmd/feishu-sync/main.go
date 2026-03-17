package main

import (
	"os"

	"github.com/your-org/feishu-sync/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
