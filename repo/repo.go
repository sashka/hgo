package repo

import (
	"fmt"
	"os"
	"path/filepath"
)

type Repo struct {
	RootDir string
}

// findRepositoryRoot looks up the directory tree from given path to find a repo root (".hg" directory).
func findRepositoryRoot(dir string) (found bool, root string) {
	prev := dir
	for {
		info, err := os.Stat(filepath.Join(dir, ".hg"))
		if err == nil && info.IsDir() {
			return true, dir
		}

		prev = dir
		dir = filepath.Dir(dir)
		if prev == dir || dir == "." {
			return false, ""
		}
	}
}

func Open(path string) (*Repo, error) {
	found, root := findRepositoryRoot(path)
	if !found {
		return nil, fmt.Errorf("no repository found in '%s' (.hg not found)", path)
	}

	return &Repo{
		RootDir: root,
	}, nil
}
