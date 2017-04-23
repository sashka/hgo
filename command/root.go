package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/sashka/hgo/repo"
)

// RootCommand is a Command that prints root directory for the repo at a given path.
type RootCommand struct {
}

func (c *RootCommand) Run(args []string) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("abort: error getting current working directory: %s", err)
		return 255
	}

	repo, err := repo.Open(wd)
	if err != nil {
		fmt.Printf("abort: %s!\n", err)
		return 255
	}

	fmt.Println(repo.RootDir)
	return 0
}

func (c *RootCommand) Synopsis() string {
	return "print the root (top) of the current working directory"
}

func (c *RootCommand) Help() string {
	helpText := `
Print the root directory of the current repository.

Returns 0 on success.
	`
	return strings.TrimSpace(helpText)
}
