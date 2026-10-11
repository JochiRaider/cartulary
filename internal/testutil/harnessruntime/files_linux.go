//go:build linux

package harnessruntime

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"golang.org/x/sys/unix"
)

// outputParent pins a non-symlink parent for a complete file transaction. Public
// artifacts can have their declared mode; private evidence uses harnesslocal.
type outputParent struct {
	file       *os.File
	path, leaf string
}

func openOutputParent(name string, create bool) (*outputParent, error) {
	parent := filepath.Dir(name)
	if create {
		root, err := rootedfs.OpenOrCreate(parent)
		if err != nil {
			return nil, err
		}
		if err := root.Close(); err != nil {
			return nil, err
		}
	}
	f, err := harnesslocal.OpenDirectory(parent, false)
	if err != nil {
		return nil, err
	}
	return &outputParent{file: f, path: parent, leaf: filepath.Base(name)}, nil
}
func (p *outputParent) check() error {
	current, err := harnesslocal.OpenDirectory(p.path, false)
	if err != nil {
		return err
	}
	defer current.Close()
	a, err := p.file.Stat()
	if err != nil {
		return err
	}
	b, err := current.Stat()
	if err != nil || !os.SameFile(a, b) {
		return harnesslocal.ErrUnsafe
	}
	return nil
}
func (p *outputParent) close() { _ = p.file.Close() }
func stableStat(a, b unix.Stat_t) bool {
	return a.Dev == b.Dev && a.Ino == b.Ino && a.Mode == b.Mode && a.Size == b.Size && a.Mtim == b.Mtim && a.Ctim == b.Ctim && a.Nlink == b.Nlink
}
func checkRegularEvidence(name string) error {
	p, err := openOutputParent(name, false)
	if err != nil {
		return err
	}
	defer p.close()
	fd, err := unix.Openat(int(p.file.Fd()), p.leaf, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return err
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Nlink != 1 || stat.Uid != uint32(os.Getuid()) {
		return harnesslocal.ErrUnsafe
	}
	return p.check()
}
func streamRegular(ctx context.Context, name string, write io.Writer) (string, os.FileMode, error) {
	p, err := openOutputParent(name, false)
	if err != nil {
		return "", 0, err
	}
	defer p.close()
	fd, err := unix.Openat(int(p.file.Fd()), p.leaf, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return "", 0, err
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	var before, after, current unix.Stat_t
	if err := unix.Fstat(fd, &before); err != nil {
		return "", 0, err
	}
	if before.Mode&unix.S_IFMT != unix.S_IFREG || before.Nlink != 1 || before.Uid != uint32(os.Getuid()) || before.Size < 0 {
		return "", 0, harnesslocal.ErrUnsafe
	}
	hash := sha256.New()
	buffer := make([]byte, 64*1024)
	var count int64
	for {
		if err := ctx.Err(); err != nil {
			return "", 0, err
		}
		n, readErr := f.Read(buffer)
		if n > 0 {
			count += int64(n)
			if count > before.Size {
				return "", 0, harnesslocal.ErrUnsafe
			}
			_, _ = hash.Write(buffer[:n])
			if write != nil {
				if _, err := write.Write(buffer[:n]); err != nil {
					return "", 0, err
				}
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return "", 0, readErr
		}
	}
	if unix.Fstat(fd, &after) != nil || unix.Fstatat(int(p.file.Fd()), p.leaf, &current, unix.AT_SYMLINK_NOFOLLOW) != nil || !stableStat(before, after) || !stableStat(before, current) || count != before.Size || p.check() != nil {
		return "", 0, harnesslocal.ErrUnsafe
	}
	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), os.FileMode(before.Mode & 07777), nil
}
func modeString(mode os.FileMode) string { return fmt.Sprintf("%04o", uint32(mode)&07777) }

type artifactMember struct {
	Type   string `json:"artifact_type"`
	Path   string `json:"relative_path"`
	Mode   string `json:"mode"`
	Digest string `json:"digest"`
}

func directoryDigest(m artifactMember) (string, error) {
	return semanticDigest(map[string]any{"artifact_type": m.Type, "relative_path": m.Path, "mode": m.Mode})
}
func walkArtifact(ctx context.Context, name string) ([]artifactMember, error) {
	members := []artifactMember{}
	var visit func(string, string) error
	visit = func(directory, prefix string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		f, err := harnesslocal.OpenDirectory(directory, false)
		if err != nil {
			return err
		}
		defer f.Close()
		entries, err := f.ReadDir(-1)
		if err != nil {
			return err
		}
		slices.SortFunc(entries, func(a, b os.DirEntry) int {
			if a.Name() < b.Name() {
				return -1
			}
			if a.Name() > b.Name() {
				return 1
			}
			return 0
		})
		for _, entry := range entries {
			relative := entry.Name()
			if prefix != "" {
				relative = prefix + "/" + relative
			}
			if !safeRelative(relative) {
				return harnesslocal.ErrUnsafe
			}
			var stat unix.Stat_t
			if err := unix.Fstatat(int(f.Fd()), entry.Name(), &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				return err
			}
			member := artifactMember{Path: relative, Mode: fmt.Sprintf("%04o", stat.Mode&0777)}
			switch stat.Mode & unix.S_IFMT {
			case unix.S_IFDIR:
				member.Type = "directory"
				member.Digest, err = directoryDigest(member)
				if err != nil {
					return err
				}
				members = append(members, member)
				if err := visit(filepath.Join(directory, entry.Name()), relative); err != nil {
					return err
				}
			case unix.S_IFREG:
				member.Type = "file"
				var mode os.FileMode
				member.Digest, mode, err = streamRegular(ctx, filepath.Join(directory, entry.Name()), nil)
				if err != nil || modeString(mode) != member.Mode {
					return harnesslocal.ErrUnsafe
				}
				members = append(members, member)
			default:
				return harnesslocal.ErrUnsafe
			}
		}
		parent := outputParent{file: f, path: directory}
		return parent.check()
	}
	if err := visit(name, ""); err != nil {
		return nil, err
	}
	slices.SortFunc(members, func(a, b artifactMember) int {
		if a.Path < b.Path {
			return -1
		}
		if a.Path > b.Path {
			return 1
		}
		return 0
	})
	return members, nil
}

type publishItem struct {
	Path, Destination, Mode, Type, Digest, Source string
	Members                                       []artifactMember
	Payloads                                      map[string]string
	Bytes                                         []byte
}
type stagedOutput struct {
	parent                                       *outputParent
	stage, extraStage, backup                    string
	published, backed, owned                     bool
	stageIdentity, extraIdentity, backupIdentity unix.Stat_t
}

var errIncompletePublication = errors.New("cache publication cleanup is incomplete")

func sameInodeAt(parent *os.File, name string, expected unix.Stat_t) error {
	var actual unix.Stat_t
	if err := unix.Fstatat(int(parent.Fd()), name, &actual, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err
	}
	if actual.Dev != expected.Dev || actual.Ino != expected.Ino || actual.Mode&unix.S_IFMT != expected.Mode&unix.S_IFMT {
		return harnesslocal.ErrUnsafe
	}
	return nil
}
func removeOwnedOutput(parent *os.File, name string, expected unix.Stat_t) error {
	err := sameInodeAt(parent, name, expected)
	if errors.Is(err, unix.ENOENT) {
		return nil
	}
	if err != nil {
		return err
	}
	return os.RemoveAll(fmt.Sprintf("/proc/self/fd/%d/%s", parent.Fd(), name))
}

func materialize(ctx context.Context, root *os.Root, item publishItem) error {
	if item.Type == "file" {
		f, err := root.OpenFile("payload", os.O_CREATE|os.O_EXCL|os.O_WRONLY|syscall.O_NOFOLLOW, 0600)
		if err != nil {
			return err
		}
		var copyErr error
		if item.Source != "" {
			digest, _, err := streamRegular(ctx, item.Source, f)
			copyErr = err
			if err == nil && digest != item.Digest {
				copyErr = harnesslocal.ErrUnsafe
			}
		} else {
			_, copyErr = f.Write(item.Bytes)
		}
		mode, err := parseMode(item.Mode)
		if copyErr == nil {
			copyErr = err
		}
		if copyErr == nil {
			copyErr = f.Chmod(mode)
		}
		if copyErr == nil {
			copyErr = f.Sync()
		}
		return errors.Join(copyErr, f.Close())
	}
	for _, m := range item.Members {
		if !safeRelative(m.Path) {
			return harnesslocal.ErrUnsafe
		}
		if m.Type == "directory" {
			if err := root.MkdirAll(m.Path, 0700); err != nil {
				return err
			}
			continue
		}
		parent := filepath.Dir(m.Path)
		if parent != "." {
			if err := root.MkdirAll(parent, 0700); err != nil {
				return err
			}
		}
		f, err := root.OpenFile(m.Path, os.O_CREATE|os.O_EXCL|os.O_WRONLY|syscall.O_NOFOLLOW, 0600)
		if err != nil {
			return err
		}
		digest, _, copyErr := streamRegular(ctx, item.Payloads[m.Path], f)
		if copyErr == nil && digest != m.Digest {
			copyErr = harnesslocal.ErrUnsafe
		}
		mode, err := parseMode(m.Mode)
		if copyErr == nil {
			copyErr = err
		}
		if copyErr == nil {
			copyErr = f.Chmod(mode)
		}
		if copyErr == nil {
			copyErr = f.Sync()
		}
		if err := errors.Join(copyErr, f.Close()); err != nil {
			return err
		}
	}
	for i := len(item.Members) - 1; i >= 0; i-- {
		m := item.Members[i]
		if m.Type == "directory" {
			mode, err := parseMode(m.Mode)
			if err != nil {
				return err
			}
			if err := root.Chmod(m.Path, mode); err != nil {
				return err
			}
		}
	}
	mode, err := parseMode(item.Mode)
	if err != nil {
		return err
	}
	return root.Chmod(".", mode)
}
func parseMode(text string) (os.FileMode, error) {
	var mode uint32
	if len(text) != 4 || text[0] != '0' {
		return 0, harnesslocal.ErrUnsafe
	}
	for _, c := range text {
		if c < '0' || c > '7' {
			return 0, harnesslocal.ErrUnsafe
		}
		mode = mode*8 + uint32(c-'0')
	}
	return os.FileMode(mode), nil
}
func safeExistingOutput(ctx context.Context, name string) error {
	info, err := os.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		_, err = walkArtifact(ctx, name)
		return err
	}
	_, _, err = streamRegular(ctx, name, nil)
	return err
}

// Stage the complete set before any destination changes. Pinned parents and
// exclusive renames prevent following a replacement; rollback restores every
// previous output if any publication fails. Only transaction-owned names are
// removed. The caller serializes cache transactions with its own mutex.
func publishOutputs(ctx context.Context, repository, run string, items []publishItem, beforeRename func(int) error) (result error) {
	id := uuid()
	staged := []stagedOutput{}
	destinations := []string{}
	defer func() {
		cleanupErrors := []error{}
		rollback := result != nil
		for i := len(staged) - 1; i >= 0; i-- {
			s := &staged[i]
			if rollback {
				if s.published {
					cleanupErrors = append(cleanupErrors, removeOwnedOutput(s.parent.file, s.parent.leaf, s.stageIdentity))
				}
				if s.backed {
					if err := sameInodeAt(s.parent.file, s.backup, s.backupIdentity); err != nil {
						cleanupErrors = append(cleanupErrors, err)
					} else {
						cleanupErrors = append(cleanupErrors, unix.Renameat2(int(s.parent.file.Fd()), s.backup, int(s.parent.file.Fd()), s.parent.leaf, unix.RENAME_NOREPLACE))
					}
				}
			} else if s.backed {
				cleanupErrors = append(cleanupErrors, removeOwnedOutput(s.parent.file, s.backup, s.backupIdentity))
			}
			if s.owned {
				cleanupErrors = append(cleanupErrors, removeOwnedOutput(s.parent.file, s.stage, s.stageIdentity))
				if s.extraStage != "" {
					cleanupErrors = append(cleanupErrors, removeOwnedOutput(s.parent.file, s.extraStage, s.extraIdentity))
				}
			}
			s.parent.close()
		}
		if err := errors.Join(cleanupErrors...); err != nil {
			result = errors.Join(result, errIncompletePublication, err)
		}
	}()
	for _, item := range items {
		if !safeRelative(item.Path) {
			return harnesslocal.ErrUnsafe
		}
		base := repository
		if item.Destination == "run_root" {
			base = run
		}
		name := filepath.Join(base, item.Path)
		for _, prior := range destinations {
			if name == prior || stringsUnder(name, prior) || stringsUnder(prior, name) {
				return harnesslocal.ErrUnsafe
			}
		}
		destinations = append(destinations, name)
		parent, err := openOutputParent(name, true)
		if err != nil {
			return err
		}
		stage := ".cartulary-cache-stage-" + parent.leaf + "-" + id
		backup := ".cartulary-cache-backup-" + parent.leaf + "-" + id
		staged = append(staged, stagedOutput{parent: parent, stage: stage, backup: backup})
		prefix := fmt.Sprintf("/proc/self/fd/%d/", parent.file.Fd())
		if err := safeExistingOutput(ctx, name); err != nil {
			return err
		}
		if err := unix.Mkdirat(int(parent.file.Fd()), stage, 0700); err != nil {
			return err
		}
		staged[len(staged)-1].owned = true
		if err := unix.Fstatat(int(parent.file.Fd()), stage, &staged[len(staged)-1].stageIdentity, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		root, err := os.OpenRoot(prefix + stage)
		if err != nil {
			return err
		}
		err = materialize(ctx, root, item)
		closeErr := root.Close()
		if err != nil || closeErr != nil {
			return errors.Join(err, closeErr)
		}
		if item.Type == "file" {
			staged[len(staged)-1].extraIdentity = staged[len(staged)-1].stageIdentity
			var identity unix.Stat_t
			if err := unix.Fstatat(int(parent.file.Fd()), stage+"/payload", &identity, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				return err
			}
			temporary := stage + "-file"
			if err := unix.Renameat2(int(parent.file.Fd()), stage+"/payload", int(parent.file.Fd()), temporary, unix.RENAME_NOREPLACE); err != nil {
				return err
			}
			staged[len(staged)-1].stage = temporary
			staged[len(staged)-1].stageIdentity = identity
			staged[len(staged)-1].extraStage = stage
			if err := unix.Unlinkat(int(parent.file.Fd()), stage, unix.AT_REMOVEDIR); err != nil {
				return err
			}
		}
	}
	for i := range staged {
		s := &staged[i]
		if err := ctx.Err(); err != nil {
			return err
		}
		if beforeRename != nil {
			if err := beforeRename(i); err != nil {
				return err
			}
		}
		if err := s.parent.check(); err != nil {
			return err
		}
		if err := sameInodeAt(s.parent.file, s.stage, s.stageIdentity); err != nil {
			return err
		}
		var stat unix.Stat_t
		if err := unix.Fstatat(int(s.parent.file.Fd()), s.parent.leaf, &stat, unix.AT_SYMLINK_NOFOLLOW); err == nil {
			if stat.Mode&unix.S_IFMT != unix.S_IFREG && stat.Mode&unix.S_IFMT != unix.S_IFDIR || stat.Mode&unix.S_IFMT == unix.S_IFREG && stat.Nlink != 1 {
				return harnesslocal.ErrUnsafe
			}
			if err := unix.Renameat2(int(s.parent.file.Fd()), s.parent.leaf, int(s.parent.file.Fd()), s.backup, unix.RENAME_NOREPLACE); err != nil {
				return err
			}
			s.backed = true
			s.backupIdentity = stat
		} else if !errors.Is(err, unix.ENOENT) {
			return err
		}
		if err := unix.Renameat2(int(s.parent.file.Fd()), s.stage, int(s.parent.file.Fd()), s.parent.leaf, unix.RENAME_NOREPLACE); err != nil {
			return err
		}
		s.published = true
		if err := sameInodeAt(s.parent.file, s.parent.leaf, s.stageIdentity); err != nil {
			return err
		}
		if err := s.parent.file.Sync(); err != nil {
			return err
		}
		if err := s.parent.check(); err != nil {
			return err
		}
	}
	return nil
}
func stringsUnder(child, parent string) bool {
	return len(child) > len(parent) && child[:len(parent)] == parent && child[len(parent)] == os.PathSeparator
}
