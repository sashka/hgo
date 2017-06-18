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
func findRepositoryRoot(path string) (string, error) {
	start := path
	prev := path
	for {
		info, err := os.Stat(filepath.Join(path, ".hg"))
		if err == nil && info.IsDir() {
			return path, nil
		}

		prev = path
		path = filepath.Dir(path)
		if prev == path || path == "." {
			return "", fmt.Errorf("no repository found in '%s' (.hg not found)", start)
		}
	}
}

func Open(path string) (*Repo, error) {
	root, err := findRepositoryRoot(path)
	if err != nil {
		return nil, err
	}

	return &Repo{
		RootDir: root,
	}, nil
}
