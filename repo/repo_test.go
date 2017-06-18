package repo

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var join = filepath.Join

func preparePlayground(t *testing.T) string {
	tempdir, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	root, err := ioutil.TempDir(tempdir, "path-test")
	if err != nil {
		t.Fatal(err)
	}
	mkdir := func(args ...string) string {
		path := join(args...)
		if err := os.MkdirAll(path, 0777); err != nil {
			t.Fatal(err)
		}
		return path
	}
	mkfile := func(path string, content string) {
		if err := ioutil.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatal(err)
		}
	}

	a := mkdir(root, "a")
	mkfile(join(mkdir(a, ".hg"), "dirstate"), "dirstate")
	mkfile(join(mkdir(a, "a"), "a.go"), "package a")

	b := mkdir(root, "b")
	mkfile(join(mkdir(b, "b"), "b.go"), "package b")

	c := mkdir(root, "c")
	mkfile(join(c, ".hg"), ".hg")
	mkfile(join(mkdir(c, "c"), "c.go"), "package c")

	return root
}

func TestFindRoot(t *testing.T) {
	root := preparePlayground(t)
	defer os.RemoveAll(root)

	tests := []struct {
		path string
		want string
		err  error
	}{
		{path: "", err: ErrRepoNotFound},
		{path: root, err: ErrRepoNotFound},
		{path: join(root, "a"), want: join(root, "a")},
		{path: join(root, "a", "a"), want: join(root, "a")},
		{path: join(root, "b"), err: ErrRepoNotFound},
		{path: join(join(root, "b"), "b"), err: ErrRepoNotFound},
		{path: join(root, "c"), err: ErrRepoNotFound},
		{path: join(join(root, "c"), "c"), err: ErrRepoNotFound},
	}

	for _, tt := range tests {
		got, err := findRoot(tt.path)
		if got != tt.want || !errors.Is(err, tt.err) {
			t.Errorf(`findRoot(%s) = ("%s", "%v"), want ("%s", "%v")`, tt.path, got, err, tt.want, tt.err)
		}
	}
}
