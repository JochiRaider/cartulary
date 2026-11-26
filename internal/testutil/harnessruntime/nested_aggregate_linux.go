//go:build linux

package harnessruntime

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// maximumAncestry bounds the walk to the init process. Real ancestries are a
// few dozen processes deep; a longer chain is not a process tree.
const maximumAncestry = 256

// processTable reads the kernel's view of a process: its parent, its argument
// vector and its working directory. Tests substitute a fixed table.
type processTable interface {
	parent(pid int) (int, error)
	arguments(pid int) ([]string, error)
	directory(pid int) (string, error)
}

type procTable struct{ root string }

func (p procTable) parent(pid int) (int, error) {
	data, err := os.ReadFile(filepath.Join(p.root, strconv.Itoa(pid), "stat"))
	if err != nil {
		return 0, err
	}
	// The command name is parenthesized and may hold spaces or parentheses;
	// the fields after its last closing parenthesis are state, then parent.
	end := bytes.LastIndexByte(data, ')')
	if end < 0 {
		return 0, errors.New("malformed process status")
	}
	fields := strings.Fields(string(data[end+1:]))
	if len(fields) < 2 {
		return 0, errors.New("malformed process status")
	}
	return strconv.Atoi(fields[1])
}

func (p procTable) arguments(pid int) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(p.root, strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00"), nil
}

func (p procTable) directory(pid int) (string, error) {
	return os.Readlink(filepath.Join(p.root, strconv.Itoa(pid), "cwd"))
}

// enclosingAggregate returns the target of an aggregate graph (test-fast,
// check or ci) that encloses process pid and works in repository, or "" when
// none does (TH-HARNESS Section 3, nested graphs; P5-2). A graph started
// beneath an aggregate against the same repository would rebuild its shared
// outputs, borrow its host state and hide its own rows from the enclosing
// run, so nested harness probes run only from explicit selections. The answer
// comes from the process ancestry, not the environment: launchers of nested
// commands deliberately clear their children's environments. Only an
// ancestor graph binary named runtime counts, and a process that leaves the
// table during the walk ends it, since its descendants were re-parented.
func enclosingAggregate(table processTable, pid int, runtime, repository string) (string, error) {
	for depth := 0; depth < maximumAncestry; depth++ {
		parent, err := table.parent(pid)
		if errors.Is(err, os.ErrNotExist) || err == nil && parent <= 1 {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		pid = parent
		arguments, err := table.arguments(pid)
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		target, aggregate := aggregateGraph(arguments, runtime)
		if !aggregate {
			continue
		}
		directory, err := table.directory(pid)
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if directory == repository {
			return target, nil
		}
	}
	return "", errors.New("process ancestry exceeds its bound")
}

// aggregateGraph reports whether arguments start the graph runtime named
// runtime for an aggregate selection, and that selection's target.
func aggregateGraph(arguments []string, runtime string) (string, bool) {
	if len(arguments) < 2 || filepath.Base(arguments[0]) != runtime || arguments[1] != "graph" {
		return "", false
	}
	selection, target := "", ""
	for i := 2; i+1 < len(arguments); i++ {
		switch arguments[i] {
		case "--selection":
			i++
			selection = arguments[i]
		case "--target":
			i++
			target = arguments[i]
		}
	}
	return target, selection == "aggregate"
}
