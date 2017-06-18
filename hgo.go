//
//  88
//  88
//  88
//  88,dPPYba,   ,adPPYb,d8  ,adPPYba,
//  88P'    "8a a8"    `Y88 a8"     "8a
//  88       88 8b       88 8b       d8
//  88       88 "8a,   ,d88 "8a,   ,a8"
//  88       88  `"YbbdP"Y8  `"YbbdP"'
//               aa,    ,88
//                "Y8bbdP"
//
// (c) 2017-2019, Alexander Saltanov <asd@mokote.com>
//

package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/sashka/hgo/command"
)

func Commands() map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"root": func() (cli.Command, error) {
			return &command.RootCommand{}, nil
		},

		"status": func() (cli.Command, error) {
			return &command.StatusCommand{}, nil
		},

		"branch": func() (cli.Command, error) {
			return &command.BranchCommand{}, nil
		},

		"debugdirstate": func() (cli.Command, error) {
			return &command.DebugDirStateCommand{}, nil
		},
	}
}

// Run is inspired by https://github.com/hashicorp/vault/blob/master/cli/main.go
func Run(args []string) int {
	cli := &cli.CLI{
		Name:     "hgo",
		Args:     args,
		Commands: Commands(),
	}

	exitCode, err := cli.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}

func main() {
	// Get the command line args.
	args := os.Args[1:]

	os.Exit(Run(args))
}
