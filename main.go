package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)

type workType int

const (
	MK_FILE workType = 1 << iota
	MK_SYMLINK
	MK_HARDLINK

	MK_OPEN
)

type work struct {
	T     workType
	Paths []string
}

const usage = `Usage:
       mk <path> [-s- <target> | -h- <target>] ...
make handles creating files, directories and symbolic links.
It unifies the unix commands 'touch' 'mkdir' and 'ln'.

Paths ending on / (slash) indicate a directory, else it's a file.
  ~/ $ mk main.go   # create file main.go
  ~/ $ mk docs/	    # create docs folder	

Missing directories are created by default. 
  ~/ $ mk non/existent/parent/and/file.txt   # create missing dirs and file.txt

Create symbolic links with -s- and hard links with -h-. Think of labeled edges.
  ~/ $ mk symname -s- target/file   # create symlink symname -s-> target/file 
  ~/ $ mk alias -h- target/file     # create hardlink alias -h-> target/file
`

func main() {
	invokeDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Print(usage)
		return
	}

	tasks := Tasks(args)

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Paths[0] < tasks[j].Paths[0]
	})

	for _, task := range tasks {
		Mk(usr, invokeDir, task)
	}
}

const MK_DIR_EDITOR = "_MK_DIR_EDITOR"
const MK_FILE_EDITOR = "_MK_FILE_EDITOR"

func execute(exe string, args ...string) error {
	path, err := exec.LookPath(exe)
	if err != nil {
		return fmt.Errorf("could not find command %q", exe)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not start editor: %v", err)
	}
	return nil
}

func OpenEditor(path string, isDir bool) {
	envVar := MK_FILE_EDITOR
	if isDir {
		envVar = MK_DIR_EDITOR
	}

	name := os.Getenv(envVar)
	if name == "" {
		log.Printf("mk: don't know with editor to open, set $%s\n", envVar)
		return
	}
	if err := execute(name, path); err != nil {
		log.Printf("mk: -o: %v", err)
		return
	}
	return
}

// TODO(liamvdv): the UNDO COMMAND WOULD BE AWESOME

func Mk(user *user.User, invokeDir string, t work) {
	switch {
	case t.T & MK_FILE > 0:
		path := t.Paths[0]
		if err := MkFile(user, path); err != nil {
			log.Fatalf("mk: \n")
		}
	case t.T & MK_SYMLINK > 0:
		sym := t.Paths[0]
		target, err := filepath.Abs(ExpandPath(user, t.Paths[1]))
		if err != nil {
			log.Printf("mk: note: cannot find absolute of %q\n", t.Paths[1])
		}
		if does, _, _ := Exists(target); !does {
			log.Fatalf("mk: target file %q does not exist\n", target)
		}
		eSym, err := filepath.Abs(ExpandPath(user, sym))
		if err != nil {
			log.Printf("mk: note: failed to expand %q\n", sym)
		} else {
			dir := filepath.Dir(eSym)
			if does, _, _ := Exists(dir); !does {
				log.Fatalf("mk: symlink directory %q doesn't exist\n"+
					"\t%s && %s\n",
					dir, os.Args[0]+" "+filepath.Join(".", dir[len(invokeDir):]), strings.Join(os.Args, " "))
			}
		}

		if err := os.Symlink(target, sym); err != nil {
			log.Fatalf("mk: failed to create symlink %s -> %s: %v\nmk: does not create missing parents when creating ", sym, target, err)
		}
	case t.T & MK_HARDLINK > 0:
		hard := t.Paths[0]
		target := t.Paths[1]
		does, _, isSym := Exists(target)
		if !does {
			log.Fatalf("mk: target file %q does not exist\n", target)
		}

		if isSym {
			log.Printf("mk: note: target %q is a symlink\n", target)
		}
		if err := os.Link(target, hard); err != nil {
			log.Fatalf("mk: failed to create hardlink %s -> %s: %v\n", hard, target, err)
		}
	}
	if t.T & MK_OPEN > 0 {
		path := t.Paths[0]
		ePath, err := filepath.Abs(ExpandPath(user, path))
		if err != nil {
			log.Printf("mk: open: failed to expand path: %v\n", err) // how should this happen??
			return
		}
		does, isDir, _ := Exists(ePath)
		if !does {
			log.Printf("mk: open: target file %q does not exist\n", ePath)
		}
		OpenEditor(t.Paths[0], isDir)
	}
}

func ExpandPath(user *user.User, path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		path = filepath.Join(user.HomeDir, path[1:])
	}
	return path
}

// invokeDir must be abspath
// path may not be absolute
func MkFile(user *user.User, path string) error {
	ePath, err := filepath.Abs(ExpandPath(user, path))
	if err != nil {
		return err
	}
	// trailing / -> dir
	if strings.HasSuffix(path, "/") {
		return EnsureDir(ePath)
	}
	// else file
	if does, _, _ := Exists(ePath); does {
		return nil
	}
	dir := filepath.Dir(ePath)
	if err := EnsureDir(dir); err != nil {
		return err
	}
	return os.WriteFile(ePath, nil, 0644)
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func Exists(path string) (exists bool, isDir bool, isSym bool) {
	fi, err := os.Lstat(path)
	exists = !os.IsNotExist(err)
	if !exists {
		return
	}
	isDir = fi.IsDir()
	isSym = fi.Mode()&os.ModeSymlink != 0 // if true, access target with os.Readlink(path)
	return
}

func Tasks(args []string) []work {
	var tasks []work
	used := make(map[int]struct{})
	for i := 1; i < len(args); i++ {
		slug := args[i]

		// TODO(liamvdv): need to check that ternary args are not flags... fails on -h- -s- 
		switch slug {
		case "-s-":
			if i == 0 {
				log.Fatal("mk: invalid input: mk /specify/symlink/path -s-> /target/path")
			}
			if i == len(args)-1 {
				log.Fatal("mk: invalid input: mk /symlink/path -s-> /specify/target/path")
			}
			tasks = append(tasks, work{T: MK_SYMLINK, Paths: []string{args[i-1], args[i+1]}})
			used[i-1] = struct{}{}
			used[i] = struct{}{}
			used[i+1] = struct{}{}
			i++
		case "-h-":
			if i == 0 {
				log.Fatal("mk: invalid input: mk /specify/hardlink/path -h-> /target/path")
			}
			if i == len(args)-1 {
				log.Fatal("mk: invalid input: mk /hardlink/path -h-> /specify/target/path")
			}
			tasks = append(tasks, work{T: MK_HARDLINK, Paths: []string{args[i-1], args[i+1]}})
			used[i-1] = struct{}{}
			used[i] = struct{}{}
			used[i+1] = struct{}{}
			i++
		}
	}
	for i := 0; i < len(args); i++ {
		if _, have := used[i]; have {
			continue
		}
		t := work{}
		if args[i] == "-o" || args[i] == "--open" {
			// TODO(liamvdv): check that args[i+1] is not flag...
			if i + 1 < len(args) {
				t.T |= MK_OPEN
				i++
			}
		}
		t.T |= MK_FILE
		t.Paths = append(t.Paths, args[i])
		tasks = append(tasks, t)
	}
	return tasks
}
