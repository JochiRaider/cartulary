//go:build linux

package harnessplan

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

const executableInputPolicyPath = "tools/executable_input_policy.json"

// documentationName matches path components that are human documentation.
var documentationName = regexp.MustCompile(`(?i)(\.(md|markdown)$|^readme(\.|$))`)

// SourceSnapshot is the executable source a plan is prepared from: every
// tracked or untracked, non-ignored repository file outside the restricted
// documentation roots, with its mode and content digest.
type SourceSnapshot struct {
	Inputs []SourceInput
	Commit string
	// State is "clean" when the executable boundary has no changes.
	State string
}

// CaptureSource enumerates and hashes the executable source under root.
func CaptureSource(root string) (SourceSnapshot, error) {
	fail := func(format string, args ...any) (SourceSnapshot, error) {
		return SourceSnapshot{}, fmt.Errorf("%w: source snapshot: %s", ErrPlan, fmt.Sprintf(format, args...))
	}
	var policy struct {
		RestrictedRoots []string `json:"restricted_roots"`
	}
	data, err := readBounded(filepath.Join(root, executableInputPolicyPath))
	if err != nil || json.Unmarshal(data, &policy) != nil || len(policy.RestrictedRoots) == 0 {
		return fail("cannot read %s", executableInputPolicyPath)
	}
	pathspec := []string{"."}
	for _, restricted := range policy.RestrictedRoots {
		pathspec = append(pathspec, ":(top,exclude,literal)"+restricted)
	}
	for _, pattern := range []string{"*.md", "*.markdown", "README", "README.*", "**/README", "**/README.*"} {
		pathspec = append(pathspec, ":(exclude,icase)"+pattern)
	}
	listing, err := git(root, append([]string{"ls-files", "--cached", "--others", "--exclude-standard", "-z", "--"}, pathspec...)...)
	if err != nil {
		return fail("cannot enumerate repository inputs: %v", err)
	}
	var paths []string
	for _, entry := range strings.Split(string(listing), "\x00") {
		switch {
		case entry == "":
		case strings.HasSuffix(entry, "/"):
			return fail("nested Git repository inside the checkout: %s; move it outside the checkout or ignore it", entry)
		case documentation(entry, policy.RestrictedRoots):
		default:
			paths = append(paths, entry)
		}
	}
	paths = dedupe(paths)
	inputs := make([]*SourceInput, len(paths))
	errs := make([]error, len(paths))
	work := make(chan int)
	var wait sync.WaitGroup
	for range runtime.GOMAXPROCS(0) {
		wait.Add(1)
		go func() {
			defer wait.Done()
			buffer := make([]byte, 64<<10)
			for i := range work {
				inputs[i], errs[i] = sourceInput(root, paths[i], buffer)
			}
		}()
	}
	for i := range paths {
		work <- i
	}
	close(work)
	wait.Wait()
	snapshot := SourceSnapshot{}
	for i, input := range inputs {
		if errs[i] != nil {
			return fail("%s: %v", paths[i], errs[i])
		}
		if input != nil {
			snapshot.Inputs = append(snapshot.Inputs, *input)
		}
	}
	commit, err := git(root, "rev-parse", "HEAD")
	if err != nil {
		return fail("cannot read HEAD: %v", err)
	}
	status, err := git(root, append([]string{"status", "--porcelain=v1", "--untracked-files=all", "--"}, pathspec...)...)
	if err != nil {
		return fail("cannot read worktree status: %v", err)
	}
	snapshot.Commit = strings.TrimSpace(string(commit))
	snapshot.State = "clean"
	if len(status) != 0 {
		snapshot.State = "dirty"
	}
	return snapshot, nil
}

func documentation(relative string, restricted []string) bool {
	for _, root := range restricted {
		if relative == root || strings.HasPrefix(relative, root+"/") {
			return true
		}
	}
	return slices.ContainsFunc(strings.Split(relative, "/"), documentationName.MatchString)
}

// sourceInput hashes one repository path. A path Git lists but the worktree
// no longer has is omitted; an ancestor that is not a directory, or a file
// that changes while it is hashed, fails the snapshot.
func sourceInput(root, relative string, buffer []byte) (*SourceInput, error) {
	if !safeRelative(relative) {
		return nil, errors.New("path is not normalized")
	}
	segments := strings.Split(relative, "/")
	absolute := root
	var stat unix.Stat_t
	for i, segment := range segments {
		absolute = filepath.Join(absolute, segment)
		if err := unix.Lstat(absolute, &stat); err != nil {
			if errors.Is(err, unix.ENOENT) {
				return nil, nil
			}
			return nil, err
		}
		if i < len(segments)-1 && stat.Mode&unix.S_IFMT != unix.S_IFDIR {
			return nil, errors.New("an ancestor is not a directory (symbolic links are forbidden)")
		}
	}
	input := &SourceInput{Path: relative, Mode: fmt.Sprintf("%04o", stat.Mode&0o7777)}
	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFLNK:
		target, err := os.Readlink(absolute)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256([]byte(target))
		input.Kind, input.ByteDigest = "symlink", "sha256:"+hex.EncodeToString(sum[:])
		return input, nil
	case unix.S_IFREG:
	default:
		return nil, errors.New("not a regular file or symbolic link")
	}
	descriptor, err := unix.Open(absolute, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(descriptor), absolute)
	defer file.Close()
	var opened unix.Stat_t
	if err := unix.Fstat(descriptor, &opened); err != nil || !sameFile(stat, opened) {
		return nil, errors.New("file changed before hashing")
	}
	hash := sha256.New()
	if _, err := io.CopyBuffer(hash, io.LimitReader(file, opened.Size+1), buffer); err != nil {
		return nil, err
	}
	var after, linked unix.Stat_t
	if unix.Fstat(descriptor, &after) != nil || unix.Lstat(absolute, &linked) != nil || !sameFile(opened, after) || !sameFile(opened, linked) {
		return nil, errors.New("file changed during hashing")
	}
	input.Kind, input.ByteDigest = "file", "sha256:"+hex.EncodeToString(hash.Sum(nil))
	return input, nil
}

func sameFile(a, b unix.Stat_t) bool {
	return a.Dev == b.Dev && a.Ino == b.Ino && a.Mode == b.Mode && a.Size == b.Size && a.Mtim == b.Mtim && a.Ctim == b.Ctim
}

// git runs one read-only Git command in root.
func git(root string, args ...string) ([]byte, error) {
	command := exec.Command("git", args...)
	command.Dir = root
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %v: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return output, nil
}
