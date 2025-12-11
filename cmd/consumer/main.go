package main

import (
	"fmt"
	"os"

	"udpie/cmd/consumer/commands"
)

func main() {
	cfg := loadConfig()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	var cmd Command
	switch command {
	case "download":
		cmd = commands.NewDownloadCommand(cfg)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	cmd.Execute()
}

// Command interface for all commands
type Command interface {
	Execute()
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s <command> [options]

Commands:
  download        Download a file by file ID

Use '%s <command> -help' for command-specific help.
`, os.Args[0], os.Args[0])
}

