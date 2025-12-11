package main

import (
	"fmt"
	"os"

	"udpie/cmd/producer/commands"
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
	case "register":
		cmd = commands.NewRegisterCommand(cfg)
	case "register-file":
		cmd = commands.NewRegisterFileCommand(cfg)
	case "listen":
		cmd = commands.NewListenCommand(cfg)
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
  register          Register a producer and get ProducerId
  register-file     Register a file
  listen            Start websocket listener

Use '%s <command> -help' for command-specific help.
`, os.Args[0], os.Args[0])
}
