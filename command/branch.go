package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/sashka/hgo/repo"
)

// RootCommand is a Command that prints root directory for the repo at a given path.
type BranchCommand struct {
}

func (c *BranchCommand) Run(args []string) int {
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

	// All the previous code ^^^ to be removed completely on stage 1.

	// dirstate.py:
	//
	// @repocache('branch')
	// def _branch(self):
	//     try:
	//         return self._opener.read("branch").strip() or "default"
	//     except IOError as inst:
	//         if inst.errno != errno.ENOENT:
	//             raise
	//         return "default"

	// https://www.mercurial-scm.org/wiki/FileFormats#line-59
	// This file contains a single line with the branch name for the branch in the working directory.
	// If it doesn't exist, the branch is '' (aka 'default').
	path := filepath.Join(repo.RootDir, ".hg", "branch")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path/to/branch does not exist
		fmt.Println("default")
		return 0
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("abort: %s!\n", err)
	}
	branch := strings.TrimSpace(string(b))

	fmt.Println(branch)
	return 0
}

func (c *BranchCommand) Synopsis() string {
	return "show the current branch name"
}

func (c *BranchCommand) Help() string {
	helpText := `
With no argument, show the current branch name.

Returns 0 on success.
	`
	return strings.TrimSpace(helpText)
}
