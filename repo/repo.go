package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Repo struct {
	RootDir string
}

func Open(path string) (*Repo, error) {
	root, err := findRoot(path)
	if err != nil {
		if errors.Is(err, ErrRepoNotFound) {
			return nil, fmt.Errorf("no repository found in '%s' (.hg not found)", path)
		}
		return nil, err
	}

	return &Repo{
		RootDir: root,
	}, nil
}

var ErrRepoNotFound = errors.New(".hg not found")

// findRoot looks up the directory tree from given path to find a repo root (".hg" directory).
func findRoot(path string) (string, error) {
	for {
		info, err := os.Stat(filepath.Join(path, ".hg"))
		if err == nil && info.IsDir() {
			return path, nil
		}

		prev := path
		path = filepath.Dir(path)
		if prev == path || path == "." {
			return "", ErrRepoNotFound
		}
	}
}
