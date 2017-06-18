package command

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/armon/go-radix"
	"github.com/sashka/hgo/repo"
)

// This file will be `package "distate"` once.
// Original Hg wiki page on DirState: https://www.mercurial-scm.org/wiki/DirState

// FileState is used to quickly determine what files in the working directory have changed.
type FileState byte

// The states that are tracked are:
//   * n - normal
//   * a - added
//   * r - removed
//   * m - 3-way merged
const (
	Normal  FileState = 110
	Added   FileState = 97
	Removed FileState = 114
	Merged  FileState = 109
)

type DebugDirStateCommand struct {
}

// https://www.mercurial-scm.org/wiki/DirState

// 3.7. dirstate

// This file contains information on the current state of the working directory in a binary format.
// It begins with two 20-byte hashes, for first and second parent, followed by an entry for each file.
//
// Each file entry is of the following form:
// 	<1-byte state><4-byte mode><4-byte size><4-byte mtime><4-byte name length><n-byte name>
//
// If the name contains a null character, it is split into two strings,
// with the second being the copy source for move and copy operations.
//
// If the dirstate file is not present, parents are assumed to be (null, null) with no files tracked.
//
// Source: mercurial/parsers.py:parse_dirstate()

func readNextBytes(file *os.File, len int) []byte {
	bytes := make([]byte, len)
	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal("!!!", err)
	}
	return bytes
}

// Header for dirstate file record:
// <1-byte state><4-byte mode><4-byte size><4-byte mtime><4-byte name length>
type Header struct {
	State   byte
	Mode    uint32
	Size    int32
	Mtime   int32
	Namelen uint32
}

type DirStateFileInfo struct {
	state FileState
	mode  uint32
	size  int64
	mtime int32
}

func (c *DebugDirStateCommand) Run(args []string) int {
	wd, err := os.Getwd()
	if err != nil {
		return Abort("error getting current working directory: %s", err)
	}

	repo, err := repo.Open(wd)
	if err != nil {
		return Abort("%s!\n", err)
	}

	// All the previous code ^^^ to be removed completely on stage 1.

	path := filepath.Join(repo.RootDir, ".hg", "dirstate")
	f, err := os.Open(path)
	if err != nil {
		return Abort("%s!\n", err)
	}
	defer f.Close()

	finfo, err := f.Stat()
	if err != nil {
		return Abort("%s!\n", err)
	}

	// read parents
	_ = readNextBytes(f, 20)
	_ = readNextBytes(f, 20)
	// fmt.Printf("parent1: %x\n", parent1)
	// fmt.Printf("parent2: %x\n", parent2)

	fileTree := radix.New()
	copyTree := radix.New()

	// read filenames
	var offset int64 = 40 // we're now right after parent2
	for offset < finfo.Size() {
		header := Header{}
		data := readNextBytes(f, 17)
		offset += 17

		buffer := bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.BigEndian, &header)
		if err != nil {
			log.Fatal("binary.Read failed", err)
		}

		name := readNextBytes(f, int(header.Namelen))
		offset += int64(header.Namelen)

		filename := string(name)

		// If the name contains a null character, it is split into two strings,
		// with the second being the copy source for move and copy operations.
		for i := 0; i < int(header.Namelen); i++ {
			if name[i] == 0 {
				copysource := string(name[i+1 : header.Namelen])
				filename = string(name[0:i])
				copyTree.Insert(filename, copysource)
				break
			}
		}

		fileTree.Insert(filename, DirStateFileInfo{
			state: FileState(header.State),
			mode:  header.Mode,
			size:  int64(header.Size),
			mtime: header.Mtime,
		})
	}

	var walkFn radix.WalkFn = func(k string, raw interface{}) bool {
		info := raw.(DirStateFileInfo)

		var mtimestr string
		if info.mtime < 0 {
			mtimestr = "unset"
		} else {
			mtimestr = time.Unix(int64(info.mtime), 0).Format("2006-01-02 15:04:05")
		}

		fmt.Printf("%s %3o %10d %-19s %s\n", string(info.state), info.mode&0x0fff, info.size, mtimestr, k)
		return false
	}

	// Walk and print!
	fileTree.Walk(walkFn)

	if copyTree.Len() > 0 {
		var copyWalkFn radix.WalkFn = func(k string, raw interface{}) bool {
			copysource := raw.(string)
			fmt.Printf("copy: %s -> %s\n", copysource, k)
			return false
		}
		copyTree.Walk(copyWalkFn)
	}

	return 0
}

func (c *DebugDirStateCommand) Synopsis() string {
	return "show the contents of the current dirstate"
}

func (c *DebugDirStateCommand) Help() string {
	helpText := `
Show the content of the current dirstate.

Returns 0 on success.
	`
	return strings.TrimSpace(helpText)
}
