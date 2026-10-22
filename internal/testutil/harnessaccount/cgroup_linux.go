//go:build linux

// Package harnessaccount measures one owned invocation using kernel lifetime
// counters. Fixture owners retain all service lifecycle and recovery authority.
package harnessaccount

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

var errAccounting = errors.New("invocation accounting unavailable or incomplete")

type counters struct {
	CPU, Peak, OOMKills uint64
	Populated           bool
}

// Counters must fit the exact integer range shared by native and JSON readers.
const maximumCounter = uint64(1<<53 - 1)

func unsigned(value string) (uint64, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.IndexFunc(value, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return 0, errAccounting
	}
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil || n > maximumCounter {
		return 0, errAccounting
	}
	return n, nil
}

func counterField(value, key string) (uint64, error) {
	var result uint64
	found := false
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return 0, errAccounting
		}
		if fields[0] != key {
			continue
		}
		if found {
			return 0, errAccounting
		}
		n, err := unsigned(fields[1])
		if err != nil {
			return 0, err
		}
		result, found = n, true
	}
	if !found {
		return 0, errAccounting
	}
	return result, nil
}

func readCounter(root *os.Root, name, key string) (uint64, error) {
	f, err := root.Open(name)
	if err != nil {
		return 0, errAccounting
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 16385))
	if err != nil || len(data) > 16384 {
		return 0, errAccounting
	}
	if key == "" {
		return unsigned(string(data))
	}
	return counterField(string(data), key)
}

type group struct {
	path        string
	root        *os.Root
	commands    *os.File
	initialPeak uint64
}

func ownMembership() (string, error) {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "", errAccounting
	}
	line := strings.TrimSpace(string(data))
	member, ok := strings.CutPrefix(line, "0::")
	if !ok || !filepath.IsAbs(member) || filepath.Clean(member) != member || strings.ContainsAny(member, "\n\r\x00") {
		return "", errAccounting
	}
	return member, nil
}

// The outer entry creates a uniquely named delegated scope. Only its sole
// initial process may prepare the hierarchy; arbitrary existing groups cannot
// be adopted. The controller moves to its own sibling before enabling memory.
func createGroup(scope string) (*group, error) {
	member, err := ownMembership()
	if err != nil || filepath.Base(member) != scope || !strings.HasPrefix(scope, "cartulary-account-") || !strings.HasSuffix(scope, ".scope") {
		return nil, errAccounting
	}
	base := filepath.Join("/sys/fs/cgroup", member)
	root, err := os.OpenRoot(base)
	if err != nil {
		return nil, errAccounting
	}
	defer root.Close()
	f, err := root.Open(".")
	if err != nil {
		return nil, errAccounting
	}
	var fs unix.Statfs_t
	err = unix.Fstatfs(int(f.Fd()), &fs)
	f.Close()
	if err != nil || fs.Type != unix.CGROUP2_SUPER_MAGIC {
		return nil, errAccounting
	}
	procs, err := root.ReadFile("cgroup.procs")
	if err != nil || strings.TrimSpace(string(procs)) != strconv.Itoa(os.Getpid()) {
		return nil, errAccounting
	}
	if root.Mkdir("controller", 0700) != nil || root.WriteFile("controller/cgroup.procs", []byte(strconv.Itoa(os.Getpid())), 0600) != nil || root.WriteFile("cgroup.subtree_control", []byte("+cpu +memory"), 0600) != nil {
		return nil, errAccounting
	}
	if root.Mkdir("measured", 0700) != nil || root.WriteFile("measured/cgroup.subtree_control", []byte("+cpu +memory"), 0600) != nil || root.Mkdir("measured/commands", 0700) != nil {
		return nil, errAccounting
	}
	measured, err := root.OpenRoot("measured")
	if err != nil {
		return nil, errAccounting
	}
	commands, err := measured.Open("commands")
	if err != nil {
		measured.Close()
		return nil, errAccounting
	}
	g := &group{path: filepath.Join(member, "measured"), root: measured, commands: commands}
	initial, err := g.snapshot()
	// Creating the empty descendant hierarchy itself can incur kernel memory
	// charges. Retain them in the peak; subtracting independent peaks is invalid.
	if err != nil || initial.CPU != 0 || initial.Populated || initial.OOMKills != 0 {
		g.close()
		return nil, errAccounting
	}
	g.initialPeak = initial.Peak
	return g, nil
}

func (g *group) snapshot() (counters, error) {
	var result counters
	var err error
	result.CPU, err = readCounter(g.root, "cpu.stat", "usage_usec")
	if err != nil {
		return counters{}, err
	}
	result.Peak, err = readCounter(g.root, "memory.peak", "")
	if err != nil {
		return counters{}, err
	}
	result.OOMKills, err = readCounter(g.root, "memory.events", "oom_kill")
	if err != nil {
		return counters{}, err
	}
	populated, err := readCounter(g.root, "cgroup.events", "populated")
	if err != nil || populated > 1 {
		return counters{}, errAccounting
	}
	result.Populated = populated == 1
	return result, nil
}

func (g *group) close() {
	if g.commands != nil {
		_ = g.commands.Close()
	}
	if g.root != nil {
		_ = g.root.Close()
	}
}

// release never kills, migrates, or recursively removes workload resources.
// A remaining container group is an unresolved lifecycle, even if unpopulated.
func (g *group) release() error {
	value, err := g.snapshot()
	if err != nil || value.Populated || value.OOMKills != 0 {
		return errAccounting
	}
	directory, err := g.root.Open(".")
	if err != nil {
		return errAccounting
	}
	entries, err := directory.ReadDir(-1)
	directory.Close()
	if err != nil {
		return errAccounting
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "commands" {
			return errAccounting
		}
	}
	if err = g.root.Remove("commands"); err != nil {
		return errAccounting
	}
	if err = os.Remove(filepath.Join("/sys/fs/cgroup", g.path)); err != nil {
		return fmt.Errorf("%w: owned hierarchy not settled", errAccounting)
	}
	return nil
}
