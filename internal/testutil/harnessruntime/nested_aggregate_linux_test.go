//go:build linux

package harnessruntime

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

type fixedProcess struct {
	parent    int
	arguments []string
	directory string
}

type fixedTable map[int]fixedProcess

func (t fixedTable) parent(pid int) (int, error) {
	if p, ok := t[pid]; ok {
		return p.parent, nil
	}
	return 0, os.ErrNotExist
}

func (t fixedTable) arguments(pid int) ([]string, error) {
	if p, ok := t[pid]; ok {
		return p.arguments, nil
	}
	return nil, os.ErrNotExist
}

func (t fixedTable) directory(pid int) (string, error) {
	if p, ok := t[pid]; ok {
		return p.directory, nil
	}
	return "", os.ErrNotExist
}

// A graph learns from its process ancestry, not its environment, whether an
// aggregate run of the same repository encloses it.
func TestEnclosingAggregate(t *testing.T) {
	const repository, runtime = "/work/cartulary", "cartulary-harness"
	graph := func(selection, target string) []string {
		return []string{"/work/cartulary/.tools/" + runtime, "graph", "--selection", selection, "--target", target}
	}
	// 40 is the graph asking; 30 a cleared-environment launcher; 20 the unit;
	// 10 the enclosing graph; 5 the shell that ran make.
	chain := func(outer []string, directory string) fixedTable {
		return fixedTable{
			40: {parent: 30, arguments: graph("command", "test-slice"), directory: repository},
			30: {parent: 20, arguments: []string{"bash", "tools/harness/tests/test-tool-output-real-targets.sh", "owner-json"}, directory: repository},
			20: {parent: 10, arguments: []string{"/usr/bin/go", "test"}, directory: repository},
			10: {parent: 5, arguments: outer, directory: directory},
			5:  {parent: 1, arguments: []string{"make", "ci"}, directory: repository},
		}
	}
	for _, c := range []struct {
		name  string
		table fixedTable
		want  string
	}{
		{"inside ci", chain(graph("aggregate", "ci"), repository), "ci"},
		{"inside check", chain(graph("aggregate", "check"), repository), "check"},
		{"inside an explicit suite", chain(graph("command", "harness-contract"), repository), ""},
		{"an aggregate of another repository", chain(graph("aggregate", "ci"), "/scratch/fixture-repository"), ""},
		{"another program's graph", chain([]string{"/usr/bin/other", "graph", "--selection", "aggregate", "--target", "ci"}, repository), ""},
		{"top level", fixedTable{40: {parent: 5, arguments: graph("aggregate", "check"), directory: repository}, 5: {parent: 1, arguments: []string{"make", "check"}}}, ""},
		{"an ancestor that exited", fixedTable{40: {parent: 30, arguments: graph("command", "build"), directory: repository}}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := enclosingAggregate(c.table, 40, runtime, repository)
			if err != nil || got != c.want {
				t.Fatalf("enclosing aggregate %q, %v; want %q", got, err, c.want)
			}
		})
	}
	t.Run("an unbounded ancestry", func(t *testing.T) {
		loop := fixedTable{40: {parent: 41, arguments: []string{"sh"}}, 41: {parent: 40, arguments: []string{"sh"}}}
		if _, err := enclosingAggregate(loop, 40, runtime, repository); err == nil {
			t.Fatal("a cyclic ancestry must not be accepted as a process tree")
		}
	})
}

// The kernel table parses a command name holding spaces and parentheses, and
// an argument vector without its trailing separator.
func TestProcTable(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "42")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"stat":    "42 (go test (x) y) S 17 42 42 0 -1",
		"cmdline": "cartulary-harness\x00graph\x00--selection\x00aggregate\x00",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("/work/cartulary", filepath.Join(directory, "cwd")); err != nil {
		t.Fatal(err)
	}
	table := procTable{root: root}
	if parent, err := table.parent(42); err != nil || parent != 17 {
		t.Fatalf("parent %d, %v", parent, err)
	}
	if arguments, err := table.arguments(42); err != nil || !slices.Equal(arguments, []string{"cartulary-harness", "graph", "--selection", "aggregate"}) {
		t.Fatalf("arguments %q, %v", arguments, err)
	}
	if cwd, err := table.directory(42); err != nil || cwd != "/work/cartulary" {
		t.Fatalf("directory %q, %v", cwd, err)
	}
	if _, err := table.parent(43); !os.IsNotExist(err) {
		t.Fatalf("a missing process must read as absent, got %v", err)
	}
}
