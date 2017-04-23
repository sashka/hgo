package command

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"time"

	"regexp"

	radix "github.com/armon/go-radix"
	"github.com/sashka/hgo/repo"
)

// StatusCommand is a Command that show status of all files.
type StatusCommand struct {
}

func printSlice(print bool, prefix string, slice []string, nostatus bool) {
	if !print || len(slice) == 0 {
		return
	}

	for _, s := range slice {
		if nostatus {
			fmt.Println(s)
		} else {
			fmt.Println(prefix, s)
		}
	}
}

// Syntax xxx.
type Syntax int

// Syntax types
const (
	Glob Syntax = iota
	Regexp
)

type pattern struct {
	syntax   Syntax
	pattern  string
	lineno   int
	original string
	re       *regexp.Regexp
}

// Mechanical translation of match._globre:
func globre(s string) string {
	i := 0
	n := len(s)
	group := 0
	var res bytes.Buffer

	for i < n {
		c := s[i : i+1]
		i++

		if c != "*" && c != "?" && c != "[" && c != "{" && c != "}" && c != "," && c != "\\" {
			res.WriteString(regexp.QuoteMeta(c))
		} else if c == "*" {
			if i < n && s[i:i+1] == "*" {
				i++
				if i < n && s[i:i+1] == "/" {
					i++
					res.WriteString("(?:.*/)?")
				} else {
					res.WriteString(".*")
				}
			} else {
				res.WriteString("[^/]*")
			}
		} else if c == "?" {
			res.WriteString(".")
		} else if c == "[" {
			j := i
			if j < n && (s[j:j+1] == "!" || s[j:j+1] == "]") {
				j++
			}
			for j < n && s[j:j+1] != "]" {
				j++
			}
			if j >= n {
				res.WriteString("\\[")
			} else {
				stuff := strings.Replace(s[i:j], "\\", "\\\\", -1)
				i = j + 1
				if stuff[0:1] == "!" {
					stuff = "^" + stuff[1:]
				} else if stuff[0:1] == "^" {
					stuff = "\\" + stuff
				}
				res.WriteString("[" + stuff + "]")
			}
		} else if c == "{" {
			group++
			res.WriteString("(?:")
		} else if c == "}" && group > 0 {
			res.WriteString(")")
			group--
		} else if c == "," && group > 0 {
			res.WriteString("|")
		} else if c == "\\" {
			if i < n {
				i++
				res.WriteString(regexp.QuoteMeta(s[i : i+1]))
			} else {
				res.WriteString(regexp.QuoteMeta(c))
			}
		} else {
			res.WriteString(regexp.QuoteMeta(c))
		}
	}

	return res.String()
}

// parse a pattern file, returning a list of
// patterns. These patterns should be given to compile()
// to be validated and converted into a match function.
//
// trailing white space is dropped.
// the escape character is backslash.
// comments start with #.
// empty lines are skipped.
//
// lines can be of the following formats:
//
// syntax: regexp # defaults following lines to non-rooted regexps
// syntax: glob   # defaults following lines to non-rooted globs
// re:pattern     # non-rooted regular expression
// glob:pattern   # non-rooted glob
// pattern        # pattern of the current default type
//
// if sourceinfo is set, returns a list of tuples:
// (pattern, lineno, originalline). This is useful to debug ignore patterns.
func parseHgIgnore(filepath string) ([]pattern, error) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	currentSyntax := Glob
	commentre := regexp.MustCompile("((?:^|[^\\\\])(?:\\\\\\\\)*)#.*")

	// read line by line
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	lineno := 1
	pats := make([]pattern, 0)

	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		lineno++

		if s == "" {
			continue
		}

		original := s

		// Strip comment
		if strings.Contains(s, "#") {
			loc := commentre.FindStringIndex(s)
			if len(loc) > 0 {
				s = strings.TrimSpace(s[:loc[0]])
			}
			s = strings.Replace(s, "\\#", "#", -1)
		}

		if s == "" {
			continue
		}

		if strings.HasPrefix(s, "syntax:") {
			s := strings.TrimSpace(s[7:])
			if s == "" {
				continue
			}

			switch s {
			case "glob":
				currentSyntax = Glob
			case "regexp":
				currentSyntax = Regexp
			default:
				return nil, fmt.Errorf("%s: invalid syntax '%s'", filepath, s)
			}
			continue
		}

		pat := pattern{
			syntax:   currentSyntax,
			lineno:   lineno,
			original: original,
			pattern:  s,
		}

		pats = append(pats, pat)
	}

	if err := scanner.Err(); err != nil {
		// Handle the error
		return nil, err
	}

	// Compile regexps
	for i := range pats {
		switch pats[i].syntax {
		case Glob:
			// Convert an extended glob string to a regexp string.
			pats[i].re = regexp.MustCompile(globre(pats[i].pattern))
		case Regexp:
			pats[i].re = regexp.MustCompile(pats[i].pattern)
		}
	}

	return pats, nil
}

func shouldIgnore(s string, pats []pattern) bool {
	if len(pats) == 0 {
		return false
	}
	for i := range pats {
		if pats[i].re.MatchString(s) {
			return true
		}
	}
	return false
}

func (c *StatusCommand) Run(args []string) int {
	listdeleted := true
	listmodified := true
	listunknown := true
	var listadded bool
	var listclean bool
	var listignored bool
	var listremoved bool
	var nostatus bool

	var added []string
	var clean []string
	var deleted []string
	var ignored []string
	var lookup []string
	var modified []string
	var removed []string
	var unknown []string

	for i := range args {
		switch args[i] {
		case "-A", "--all":
			listignored = true
			listclean = true
			listunknown = true
			listmodified = true
			listadded = true
			listremoved = true
			listdeleted = true
		case "-m", "--modified":
			listmodified = true
		case "-a", "--added":
			listadded = true
		case "-r", "--removed":
			listremoved = true
		case "-d", "--deleted":
			listdeleted = true
		case "-c", "--clean":
			listclean = true
		case "-u", "--unknown":
			listunknown = true
		case "-i", "--ignored ":
			listignored = true
		case "-n", "--no-status":
			nostatus = true
		}
	}

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

	path := filepath.Join(repo.RootDir, ".hg", "dirstate")
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("abort: %s!\n", err)
		return 255
	}
	defer f.Close()

	finfo, err := f.Stat()
	if err != nil {
		fmt.Printf("abort: %s!\n", err)
		return 255
	}

	// skip parents
	readNextBytes(f, 40)

	// dirstate content
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
			log.Fatal("binary.Read failed:", err)
			return 255
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

	// step 0: read .hgignore
	ignoreMatchers, err := parseHgIgnore(filepath.Join(repo.RootDir, ".hgignore"))
	if err != nil {
		fmt.Printf("abort: %s!\n", err)
		return 255
	}

	// step 1: find all explicit files
	filesFound := radix.New()

	// Prepare traverse function...
	var fsWalkFn filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
		filename := strings.TrimPrefix(path, repo.RootDir+"/")

		// Skip .hg and all its files and subdirs unconditionnaly.
		if filename == ".hg" && info.IsDir() {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			filesFound.Insert(filename, info)
		}
		return nil
	}
	// ... and walk.
	filepath.Walk(repo.RootDir, fsWalkFn)

	// Traverse the dirstate fileTree too
	var dirstateWalkFn radix.WalkFn = func(k string, raw interface{}) bool {
		info, err := os.Stat(filepath.Join(repo.RootDir, k))
		if err != nil {
			filesFound.Insert(k, nil)
		}
		filesFound.Insert(k, info)
		return false
	}
	fileTree.Walk(dirstateWalkFn)

	// Walk all found files and put the into respective tree:
	var tossWalkFn radix.WalkFn = func(k string, raw interface{}) bool {
		dirstateraw, found := fileTree.Get(k)

		if !found {
			if shouldIgnore(k, ignoreMatchers) {
				ignored = append(ignored, k)
			} else {
				unknown = append(unknown, k)
			}
			return false
		}

		dirstatefileinfo := dirstateraw.(DirStateFileInfo)

		if raw == nil && (dirstatefileinfo.state == Normal || dirstatefileinfo.state == Merged || dirstatefileinfo.state == Added) {
			deleted = append(deleted, k)
			return false
		}

		stat := raw.(os.FileInfo)
		_, foundCopy := copyTree.Get(k)
		if foundCopy {
			added = append(added, k)
			return false
		}

		switch dirstatefileinfo.state {
		case Normal:
			if dirstatefileinfo.size > 0 && (dirstatefileinfo.size != stat.Size() || dirstatefileinfo.mode&0x0fff != uint32(stat.Mode())) || dirstatefileinfo.size == -2 || foundCopy {
				modified = append(modified, k)
			} else if int64(dirstatefileinfo.mtime) != stat.ModTime().UnixNano()/int64(time.Second) {
				lookup = append(lookup, k)
			} else {
				clean = append(clean, k)
			}

		case Merged:
			modified = append(modified, k)

		case Added:
			added = append(added, k)

		case Removed:
			removed = append(removed, k)

		}

		return false
	}
	filesFound.Walk(tossWalkFn)

	// Print respective slices to show the status():
	printSlice(listmodified, "M", modified, nostatus)
	printSlice(listadded, "A", added, nostatus)
	printSlice(listremoved, "R", removed, nostatus)
	printSlice(listdeleted, "!", deleted, nostatus)
	printSlice(listunknown, "?", unknown, nostatus)
	printSlice(listignored, "I", ignored, nostatus)
	printSlice(listclean, "C", clean, nostatus)

	return 0
}

func (c *StatusCommand) Synopsis() string {
	return "show changed files in the working directory"
}

func (c *StatusCommand) Help() string {
	helpText := `
Show status of files in the repository.
If names are given, only files that match are shown.

The codes used to show the status of files are:

	M = modified
	A = added
	R = removed
	C = clean
	! = missing (deleted by non-hg command, but still tracked)
	? = not tracked
	I = ignored
	= origin of the previous file (with --copies)

Returns 0 on success.
	`
	return strings.TrimSpace(helpText)
}
