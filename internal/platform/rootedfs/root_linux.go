//go:build linux

package rootedfs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	pathpkg "path"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	"golang.org/x/text/unicode/norm"
)

const (
	directoryOpenFlags = unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC | unix.O_NOFOLLOW
	regularReadFlags   = unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK
	regularWriteFlags  = unix.O_WRONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK
)

type objectIdentity struct {
	device uint64
	inode  uint64
}

type Root struct {
	mu       sync.RWMutex
	fd       int
	parentFD int
	name     string
	identity objectIdentity
	closed   bool
}

func Open(rootPath string) (*Root, error) {
	return open(rootPath, false)
}

// OpenOrCreate opens a canonical absolute root, securely creating missing
// directory components with private permissions. Existing components are
// always opened with no-follow semantics.
func OpenOrCreate(rootPath string) (*Root, error) {
	return open(rootPath, true)
}

func open(rootPath string, create bool) (*Root, error) {
	if err := validateAbsoluteRootPath(rootPath); err != nil {
		return nil, operationError("open-root", Reference{}, "root path is not a canonical absolute POSIX directory", err)
	}
	rootFD, parentFD, name, err := openRootPath(rootPath, create)
	if err != nil {
		return nil, operationError("open-root", Reference{}, "root directory is unavailable or unsafe", err)
	}
	var stat unix.Stat_t
	if err := unix.Fstat(rootFD, &stat); err != nil || stat.Mode&unix.S_IFMT != unix.S_IFDIR {
		unix.Close(rootFD)
		unix.Close(parentFD)
		return nil, operationError("open-root", Reference{}, "root object is not a directory", err)
	}
	root := &Root{
		fd:       rootFD,
		parentFD: parentFD,
		name:     name,
		identity: identityFromStat(stat),
	}
	if err := root.checkRootIdentity(); err != nil {
		root.Close()
		return nil, err
	}
	return root, nil
}

func (root *Root) Close() error {
	root.mu.Lock()
	defer root.mu.Unlock()
	if root.closed {
		return nil
	}
	root.closed = true
	rootErr := unix.Close(root.fd)
	parentErr := unix.Close(root.parentFD)
	if rootErr != nil {
		return operationError("close-root", Reference{}, "close failed", rootErr)
	}
	if parentErr != nil {
		return operationError("close-root", Reference{}, "close failed", parentErr)
	}
	return nil
}

func (root *Root) Check() error {
	root.mu.RLock()
	defer root.mu.RUnlock()
	return root.checkReady("check")
}

// Exists reports whether a rooted object exists without opening or reading the
// object. The final component is inspected with no-follow semantics, so callers
// can safely perform presence-only policy checks for regular files, symlinks,
// directories, and other filesystem objects.
func (root *Root) Exists(reference Reference) (bool, error) {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if err := root.checkReady("exists"); err != nil {
		return false, err
	}
	chain, finalName, err := root.openParent(reference)
	if err != nil {
		return false, operationError("exists", reference, "parent traversal failed", err)
	}
	defer chain.close()
	var stat unix.Stat_t
	if err := unix.Fstatat(chain.lastFD(), finalName, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		if errors.Is(err, unix.ENOENT) {
			return false, nil
		}
		return false, operationError("exists", reference, "object presence is unavailable", err)
	}
	if err := root.validateChain(chain); err != nil {
		return false, operationError("exists", reference, "directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		return false, err
	}
	return true, nil
}

func (root *Root) MakePrivateDir(reference Reference) error {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if err := root.checkReady("mkdir"); err != nil {
		return err
	}
	if err := validateReference(reference.value); err != nil {
		return operationError("mkdir", reference, "invalid reference", err)
	}

	chain, err := root.openDirectoryChain(reference, true)
	if err != nil {
		return operationError("mkdir", reference, "directory creation or traversal failed", err)
	}
	defer chain.close()
	if err := root.validateChain(chain); err != nil {
		chain.rollbackCreatedDirs()
		return operationError("mkdir", reference, "directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		chain.rollbackCreatedDirs()
		return err
	}
	return nil
}

func (root *Root) ReadRegular(reference Reference, maxBytes int64) ([]byte, Metadata, error) {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if maxBytes <= 0 {
		return nil, Metadata{}, operationError("read", reference, "maximum size must be positive", nil)
	}
	if err := root.checkReady("read"); err != nil {
		return nil, Metadata{}, err
	}
	chain, finalName, err := root.openParent(reference)
	if err != nil {
		return nil, Metadata{}, operationError("read", reference, "parent traversal failed", err)
	}
	defer chain.close()
	fd, err := unix.Openat(chain.lastFD(), finalName, regularReadFlags, 0)
	if err != nil {
		return nil, Metadata{}, operationError("read", reference, "regular file is unavailable", err)
	}
	file := os.NewFile(uintptr(fd), reference.value)
	defer file.Close()

	before, err := requireBoundedRegularFile(fd, maxBytes, true)
	if err != nil {
		return nil, Metadata{}, operationError("read", reference, "object is not an allowed bounded regular file", err)
	}
	payload, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, Metadata{}, operationError("read", reference, "file read failed", err)
	}
	if int64(len(payload)) > maxBytes {
		return nil, Metadata{}, operationError("read", reference, "file exceeds maximum size", fs.ErrInvalid)
	}
	after, err := requireBoundedRegularFile(fd, maxBytes, true)
	if err != nil || before.Dev != after.Dev || before.Ino != after.Ino ||
		before.Size != after.Size || before.Mtim != after.Mtim || before.Ctim != after.Ctim {
		return nil, Metadata{}, operationError("read", reference, "file identity changed during read", err)
	}
	if err := root.validateChain(chain); err != nil {
		return nil, Metadata{}, operationError("read", reference, "directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		return nil, Metadata{}, err
	}
	return append([]byte(nil), payload...), metadataFromStat(after), nil
}

func (root *Root) OpenRegular(reference Reference) (ReadSeekCloser, Metadata, error) {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if err := root.checkReady("open"); err != nil {
		return nil, Metadata{}, err
	}
	chain, finalName, err := root.openParent(reference)
	if err != nil {
		return nil, Metadata{}, operationError("open", reference, "parent traversal failed", err)
	}
	defer chain.close()
	fd, err := unix.Openat(chain.lastFD(), finalName, regularReadFlags, 0)
	if err != nil {
		return nil, Metadata{}, operationError("open", reference, "regular file is unavailable", err)
	}
	file := os.NewFile(uintptr(fd), reference.value)
	stat, err := requireRegularFile(fd, true)
	if err != nil {
		file.Close()
		return nil, Metadata{}, operationError("open", reference, "object is not an allowed regular file", err)
	}
	if err := root.validateChain(chain); err != nil {
		file.Close()
		return nil, Metadata{}, operationError("open", reference, "directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		file.Close()
		return nil, Metadata{}, err
	}
	return file, metadataFromStat(stat), nil
}

func (root *Root) ListRegular() ([]RegularEntry, error) {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if err := root.checkReady("list"); err != nil {
		return nil, err
	}
	rootFD, err := unix.Dup(root.fd)
	if err != nil {
		return nil, operationError("list", Reference{}, "root traversal failed", err)
	}
	defer unix.Close(rootFD)
	entries := make([]RegularEntry, 0)
	if err := root.listRegularDirectory(rootFD, "", &entries); err != nil {
		return nil, operationError("list", Reference{}, "directory traversal found an unsafe or changed object", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		return nil, err
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].Reference.value < entries[right].Reference.value
	})
	return entries, nil
}

func (root *Root) listRegularDirectory(directoryFD int, prefix string, result *[]RegularEntry) error {
	readFD, err := unix.Openat(directoryFD, ".", directoryOpenFlags, 0)
	if err != nil {
		return err
	}
	directory := os.NewFile(uintptr(readFD), "rooted-directory")
	entries, err := directory.ReadDir(-1)
	closeErr := directory.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].Name() < entries[right].Name()
	})
	for _, entry := range entries {
		name := entry.Name()
		if name == "." || name == ".." || strings.HasPrefix(name, ".cartulary-tmp-") {
			continue
		}
		rawReference := name
		if prefix != "" {
			rawReference = prefix + "/" + name
		}
		reference, err := ParseReference(rawReference)
		if err != nil {
			return err
		}
		var before unix.Stat_t
		if err := unix.Fstatat(directoryFD, name, &before, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		switch before.Mode & unix.S_IFMT {
		case unix.S_IFREG:
			if err := requireAllowedRegularStat(before, true); err != nil {
				return err
			}
			*result = append(*result, RegularEntry{
				Reference: reference,
				Metadata:  metadataFromStat(before),
			})
		case unix.S_IFDIR:
			childFD, err := unix.Openat(directoryFD, name, directoryOpenFlags, 0)
			if err != nil {
				return err
			}
			var opened unix.Stat_t
			if err := unix.Fstat(childFD, &opened); err != nil ||
				opened.Mode&unix.S_IFMT != unix.S_IFDIR ||
				identityFromStat(opened) != identityFromStat(before) {
				unix.Close(childFD)
				return ErrRootIdentityChanged
			}
			walkErr := root.listRegularDirectory(childFD, reference.value, result)
			var after unix.Stat_t
			afterErr := unix.Fstatat(directoryFD, name, &after, unix.AT_SYMLINK_NOFOLLOW)
			unix.Close(childFD)
			if walkErr != nil {
				return walkErr
			}
			if afterErr != nil || after.Mode&unix.S_IFMT != unix.S_IFDIR ||
				identityFromStat(after) != identityFromStat(before) {
				return ErrRootIdentityChanged
			}
		default:
			return fs.ErrInvalid
		}
	}
	return nil
}

func (root *Root) CreateExclusive(ctx context.Context, reference Reference, write WriteFunc) error {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if write == nil {
		return operationError("create", reference, "write callback is required", nil)
	}
	if err := root.checkReady("create"); err != nil {
		return err
	}
	chain, finalName, err := root.openParent(reference)
	if err != nil {
		return operationError("create", reference, "parent traversal failed", err)
	}
	defer chain.close()
	if err := root.validateChain(chain); err != nil {
		return operationError("create", reference, "directory identity changed", err)
	}
	if err := contextCause(ctx); err != nil {
		return operationError("create", reference, "operation canceled", err)
	}

	var destination unix.Stat_t
	destinationErr := unix.Fstatat(chain.lastFD(), finalName, &destination, unix.AT_SYMLINK_NOFOLLOW)
	switch {
	case destinationErr == nil:
		return operationError("create", reference, "exclusive destination already exists", fs.ErrExist)
	case !errors.Is(destinationErr, unix.ENOENT):
		return operationError("create", reference, "destination inspection failed", destinationErr)
	}

	tempName, tempFD, err := createTemporaryFile(chain.lastFD())
	if err != nil {
		return operationError("create", reference, "temporary file creation failed", err)
	}
	tempExists := true
	defer func() {
		if tempExists {
			_ = unix.Unlinkat(chain.lastFD(), tempName, 0)
		}
	}()
	if err := writeAndSeal(ctx, tempFD, reference, write); err != nil {
		return err
	}
	if err := root.validateChain(chain); err != nil {
		return operationError("create", reference, "directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		return err
	}
	if err := unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_NOREPLACE); err != nil {
		return operationError("create", reference, "atomic exclusive publication failed", err)
	}
	tempExists = false
	if err := root.validateChain(chain); err != nil {
		_ = unix.Unlinkat(chain.lastFD(), finalName, 0)
		return operationError("create", reference, "directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		_ = unix.Unlinkat(chain.lastFD(), finalName, 0)
		return err
	}
	if err := unix.Fsync(chain.lastFD()); err != nil {
		return operationError("create", reference, "parent synchronization failed", err)
	}
	return nil
}

func (root *Root) AtomicReplace(ctx context.Context, reference Reference, write WriteFunc) error {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if write == nil {
		return operationError("replace", reference, "write callback is required", nil)
	}
	if err := root.checkReady("replace"); err != nil {
		return err
	}
	chain, finalName, err := root.openParent(reference)
	if err != nil {
		return operationError("replace", reference, "parent traversal failed", err)
	}
	defer chain.close()
	if err := root.validateChain(chain); err != nil {
		return operationError("replace", reference, "directory identity changed", err)
	}

	tempName, tempFD, err := createTemporaryFile(chain.lastFD())
	if err != nil {
		return operationError("replace", reference, "temporary file creation failed", err)
	}
	tempExists := true
	defer func() {
		if tempExists {
			_ = unix.Unlinkat(chain.lastFD(), tempName, 0)
		}
	}()
	if err := writeAndSeal(ctx, tempFD, reference, write); err != nil {
		return err
	}
	if err := root.validateChain(chain); err != nil {
		return operationError("replace", reference, "directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		return err
	}

	var destination unix.Stat_t
	destinationErr := unix.Fstatat(chain.lastFD(), finalName, &destination, unix.AT_SYMLINK_NOFOLLOW)
	switch {
	case destinationErr == nil:
		if err := requireAllowedRegularStat(destination, true); err != nil {
			return operationError("replace", reference, "existing destination is not an allowed regular file", err)
		}
		if err := unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_EXCHANGE); err != nil {
			return operationError("replace", reference, "atomic exchange failed", err)
		}
		if err := root.validateChain(chain); err != nil {
			_ = unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_EXCHANGE)
			return operationError("replace", reference, "directory identity changed", err)
		}
		if err := root.checkRootIdentity(); err != nil {
			_ = unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_EXCHANGE)
			return err
		}
		var displaced unix.Stat_t
		if err := unix.Fstatat(chain.lastFD(), tempName, &displaced, unix.AT_SYMLINK_NOFOLLOW); err != nil ||
			requireAllowedRegularStat(displaced, true) != nil ||
			identityFromStat(displaced) != identityFromStat(destination) {
			_ = unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_EXCHANGE)
			return operationError("replace", reference, "destination changed during replacement", err)
		}
		if err := unix.Unlinkat(chain.lastFD(), tempName, 0); err != nil {
			if rollbackErr := unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_EXCHANGE); rollbackErr == nil {
				_ = unix.Unlinkat(chain.lastFD(), tempName, 0)
			}
			return operationError("replace", reference, "displaced file cleanup failed", err)
		}
		tempExists = false
	case errors.Is(destinationErr, unix.ENOENT):
		if err := unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_NOREPLACE); err != nil {
			return operationError("replace", reference, "atomic publication failed", err)
		}
		tempExists = false
		if err := root.validateChain(chain); err != nil {
			_ = unix.Unlinkat(chain.lastFD(), finalName, 0)
			return operationError("replace", reference, "directory identity changed", err)
		}
		if err := root.checkRootIdentity(); err != nil {
			_ = unix.Unlinkat(chain.lastFD(), finalName, 0)
			return err
		}
	default:
		return operationError("replace", reference, "destination inspection failed", destinationErr)
	}
	if err := root.checkRootIdentity(); err != nil {
		return err
	}
	if err := unix.Fsync(chain.lastFD()); err != nil {
		return operationError("replace", reference, "parent synchronization failed", err)
	}
	return nil
}

func (root *Root) RenameExclusive(source Reference, destination Reference) error {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if err := root.checkReady("rename"); err != nil {
		return err
	}
	sourceChain, sourceName, err := root.openParent(source)
	if err != nil {
		return operationError("rename", source, "source traversal failed", err)
	}
	defer sourceChain.close()
	destinationChain, destinationName, err := root.openParent(destination)
	if err != nil {
		return operationError("rename", destination, "destination traversal failed", err)
	}
	defer destinationChain.close()
	var sourceStat unix.Stat_t
	if err := unix.Fstatat(sourceChain.lastFD(), sourceName, &sourceStat, unix.AT_SYMLINK_NOFOLLOW); err != nil ||
		requireAllowedRegularStat(sourceStat, true) != nil {
		return operationError("rename", source, "source is not an allowed regular file", err)
	}
	if err := root.validateChain(sourceChain); err != nil {
		return operationError("rename", source, "source directory identity changed", err)
	}
	if err := root.validateChain(destinationChain); err != nil {
		return operationError("rename", destination, "destination directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		return err
	}
	if err := unix.Renameat2(sourceChain.lastFD(), sourceName, destinationChain.lastFD(), destinationName, unix.RENAME_NOREPLACE); err != nil {
		return operationError("rename", destination, "exclusive rename failed", err)
	}
	if err := root.validateChain(sourceChain); err != nil {
		_ = unix.Renameat2(destinationChain.lastFD(), destinationName, sourceChain.lastFD(), sourceName, unix.RENAME_NOREPLACE)
		return operationError("rename", source, "source directory identity changed", err)
	}
	if err := root.validateChain(destinationChain); err != nil {
		_ = unix.Renameat2(destinationChain.lastFD(), destinationName, sourceChain.lastFD(), sourceName, unix.RENAME_NOREPLACE)
		return operationError("rename", destination, "destination directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		_ = unix.Renameat2(destinationChain.lastFD(), destinationName, sourceChain.lastFD(), sourceName, unix.RENAME_NOREPLACE)
		return err
	}
	if err := unix.Fsync(sourceChain.lastFD()); err != nil {
		return operationError("rename", source, "source parent synchronization failed", err)
	}
	if destinationChain.lastFD() != sourceChain.lastFD() {
		if err := unix.Fsync(destinationChain.lastFD()); err != nil {
			return operationError("rename", destination, "destination parent synchronization failed", err)
		}
	}
	return nil
}

func (root *Root) RemoveRegular(reference Reference) error {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if err := root.checkReady("remove"); err != nil {
		return err
	}
	chain, finalName, err := root.openParent(reference)
	if err != nil {
		return operationError("remove", reference, "parent traversal failed", err)
	}
	defer chain.close()
	var stat unix.Stat_t
	if err := unix.Fstatat(chain.lastFD(), finalName, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil ||
		requireAllowedRegularStat(stat, true) != nil {
		return operationError("remove", reference, "object is not an allowed regular file", err)
	}
	if err := root.validateChain(chain); err != nil {
		return operationError("remove", reference, "directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		return err
	}
	tempName, err := temporaryName()
	if err != nil {
		return operationError("remove", reference, "cleanup name allocation failed", err)
	}
	if err := unix.Renameat2(chain.lastFD(), finalName, chain.lastFD(), tempName, unix.RENAME_NOREPLACE); err != nil {
		return operationError("remove", reference, "quarantine rename failed", err)
	}
	if err := root.validateChain(chain); err != nil {
		_ = unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_NOREPLACE)
		return operationError("remove", reference, "directory identity changed", err)
	}
	if err := root.checkRootIdentity(); err != nil {
		_ = unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_NOREPLACE)
		return err
	}
	if err := unix.Unlinkat(chain.lastFD(), tempName, 0); err != nil {
		_ = unix.Renameat2(chain.lastFD(), tempName, chain.lastFD(), finalName, unix.RENAME_NOREPLACE)
		return operationError("remove", reference, "quarantined file cleanup failed", err)
	}
	if err := unix.Fsync(chain.lastFD()); err != nil {
		return operationError("remove", reference, "parent synchronization failed", err)
	}
	return nil
}

type directoryChain struct {
	fds        []int
	names      []string
	identities []objectIdentity
	created    []bool
}

func (chain *directoryChain) lastFD() int {
	return chain.fds[len(chain.fds)-1]
}

func (chain *directoryChain) close() {
	for index := len(chain.fds) - 1; index >= 0; index-- {
		_ = unix.Close(chain.fds[index])
	}
}

func (chain *directoryChain) rollbackCreatedDirs() {
	for index := len(chain.names) - 1; index >= 0; index-- {
		if index >= len(chain.created) || !chain.created[index] {
			continue
		}
		_ = unix.Unlinkat(chain.fds[index], chain.names[index], unix.AT_REMOVEDIR)
	}
}

func (root *Root) openParent(reference Reference) (*directoryChain, string, error) {
	if err := validateReference(reference.value); err != nil {
		return nil, "", err
	}
	components := strings.Split(reference.value, "/")
	parent := Reference{value: strings.Join(components[:len(components)-1], "/")}
	chain, err := root.openDirectoryChain(parent, false)
	if err != nil {
		return nil, "", err
	}
	return chain, components[len(components)-1], nil
}

func (root *Root) openDirectoryChain(reference Reference, create bool) (*directoryChain, error) {
	rootFD, err := unix.Dup(root.fd)
	if err != nil {
		return nil, err
	}
	var rootStat unix.Stat_t
	if err := unix.Fstat(rootFD, &rootStat); err != nil {
		unix.Close(rootFD)
		return nil, err
	}
	chain := &directoryChain{
		fds:        []int{rootFD},
		identities: []objectIdentity{identityFromStat(rootStat)},
	}
	if reference.value == "" {
		return chain, nil
	}
	if err := validateReference(reference.value); err != nil {
		chain.close()
		return nil, err
	}
	for _, component := range strings.Split(reference.value, "/") {
		parentFD := chain.lastFD()
		childFD, openErr := unix.Openat(parentFD, component, directoryOpenFlags, 0)
		created := false
		if errors.Is(openErr, unix.ENOENT) && create {
			if err := unix.Mkdirat(parentFD, component, 0o700); err != nil {
				chain.rollbackCreatedDirs()
				chain.close()
				return nil, err
			}
			created = true
			childFD, openErr = unix.Openat(parentFD, component, directoryOpenFlags, 0)
		}
		if openErr != nil {
			if created {
				_ = unix.Unlinkat(parentFD, component, unix.AT_REMOVEDIR)
			}
			chain.rollbackCreatedDirs()
			chain.close()
			return nil, openErr
		}
		var childStat unix.Stat_t
		if err := unix.Fstat(childFD, &childStat); err != nil || childStat.Mode&unix.S_IFMT != unix.S_IFDIR {
			unix.Close(childFD)
			if created {
				_ = unix.Unlinkat(parentFD, component, unix.AT_REMOVEDIR)
			}
			chain.rollbackCreatedDirs()
			chain.close()
			return nil, err
		}
		if create && childStat.Mode&0o077 != 0 {
			unix.Close(childFD)
			if created {
				_ = unix.Unlinkat(parentFD, component, unix.AT_REMOVEDIR)
			}
			chain.rollbackCreatedDirs()
			chain.close()
			return nil, fs.ErrPermission
		}
		chain.names = append(chain.names, component)
		chain.created = append(chain.created, created)
		chain.fds = append(chain.fds, childFD)
		chain.identities = append(chain.identities, identityFromStat(childStat))
	}
	return chain, nil
}

func (root *Root) validateChain(chain *directoryChain) error {
	if len(chain.fds) == 0 || chain.identities[0] != root.identity {
		return ErrRootIdentityChanged
	}
	for index, fd := range chain.fds {
		var actual unix.Stat_t
		if err := unix.Fstat(fd, &actual); err != nil ||
			actual.Mode&unix.S_IFMT != unix.S_IFDIR ||
			identityFromStat(actual) != chain.identities[index] {
			return ErrRootIdentityChanged
		}
		if index == 0 {
			continue
		}
		var linked unix.Stat_t
		if err := unix.Fstatat(chain.fds[index-1], chain.names[index-1], &linked, unix.AT_SYMLINK_NOFOLLOW); err != nil ||
			linked.Mode&unix.S_IFMT != unix.S_IFDIR ||
			identityFromStat(linked) != chain.identities[index] {
			return ErrRootIdentityChanged
		}
	}
	return nil
}

func (root *Root) checkReady(operation string) error {
	if root.closed {
		return operationError(operation, Reference{}, "capability is closed", ErrClosed)
	}
	return root.checkRootIdentity()
}

func (root *Root) checkRootIdentity() error {
	var opened unix.Stat_t
	if err := unix.Fstat(root.fd, &opened); err != nil ||
		opened.Mode&unix.S_IFMT != unix.S_IFDIR ||
		identityFromStat(opened) != root.identity {
		return operationError("check-root", Reference{}, "root identity changed", ErrRootIdentityChanged)
	}
	var linked unix.Stat_t
	if err := unix.Fstatat(root.parentFD, root.name, &linked, unix.AT_SYMLINK_NOFOLLOW); err != nil ||
		linked.Mode&unix.S_IFMT != unix.S_IFDIR ||
		identityFromStat(linked) != root.identity {
		return operationError("check-root", Reference{}, "root identity changed", ErrRootIdentityChanged)
	}
	return nil
}

func openRootPath(rootPath string, create bool) (int, int, string, error) {
	current, err := unix.Open("/", directoryOpenFlags, 0)
	if err != nil {
		return -1, -1, "", err
	}
	if rootPath == "/" {
		rootFD, err := unix.Dup(current)
		if err != nil {
			unix.Close(current)
			return -1, -1, "", err
		}
		return rootFD, current, ".", nil
	}
	components := strings.Split(strings.TrimPrefix(rootPath, "/"), "/")
	for index, component := range components {
		next, openErr := unix.Openat(current, component, directoryOpenFlags, 0)
		if errors.Is(openErr, unix.ENOENT) && create {
			if mkdirErr := unix.Mkdirat(current, component, 0o700); mkdirErr != nil && !errors.Is(mkdirErr, unix.EEXIST) {
				unix.Close(current)
				return -1, -1, "", mkdirErr
			}
			if syncErr := unix.Fsync(current); syncErr != nil {
				unix.Close(current)
				return -1, -1, "", syncErr
			}
			next, openErr = unix.Openat(current, component, directoryOpenFlags, 0)
		}
		if openErr != nil {
			unix.Close(current)
			return -1, -1, "", openErr
		}
		if index == len(components)-1 {
			return next, current, component, nil
		}
		unix.Close(current)
		current = next
	}
	unix.Close(current)
	return -1, -1, "", fs.ErrInvalid
}

func validateAbsoluteRootPath(rootPath string) error {
	switch {
	case rootPath == "" || !strings.HasPrefix(rootPath, "/"):
		return fs.ErrInvalid
	case !utf8.ValidString(rootPath) || strings.IndexByte(rootPath, 0) >= 0:
		return fs.ErrInvalid
	case strings.Contains(rootPath, `\`):
		return fs.ErrInvalid
	case norm.NFC.String(rootPath) != rootPath:
		return fs.ErrInvalid
	case pathpkg.Clean(rootPath) != rootPath:
		return fs.ErrInvalid
	default:
		return nil
	}
}

func writeAndSeal(ctx context.Context, fd int, reference Reference, write WriteFunc) error {
	file := os.NewFile(uintptr(fd), reference.value)
	if err := contextCause(ctx); err != nil {
		file.Close()
		return operationError("write", reference, "operation canceled", err)
	}
	writeErr := write(file)
	contextErr := contextCause(ctx)
	syncErr := file.Sync()
	var stat unix.Stat_t
	statErr := unix.Fstat(fd, &stat)
	closeErr := file.Close()
	switch {
	case writeErr != nil:
		return operationError("write", reference, "write callback failed", writeErr)
	case contextErr != nil:
		return operationError("write", reference, "operation canceled", contextErr)
	case syncErr != nil:
		return operationError("write", reference, "file synchronization failed", syncErr)
	case statErr != nil:
		return operationError("write", reference, "file inspection failed", statErr)
	case requireAllowedRegularStat(stat, true) != nil:
		return operationError("write", reference, "created object is not an allowed regular file", fs.ErrInvalid)
	case closeErr != nil:
		return operationError("write", reference, "file close failed", closeErr)
	default:
		return nil
	}
}

func contextCause(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return context.Cause(ctx)
}

func createTemporaryFile(parentFD int) (string, int, error) {
	for range 32 {
		name, err := temporaryName()
		if err != nil {
			return "", -1, err
		}
		fd, err := unix.Openat(parentFD, name, regularWriteFlags|unix.O_CREAT|unix.O_EXCL, 0o600)
		if err == nil {
			return name, fd, nil
		}
		if !errors.Is(err, unix.EEXIST) {
			return "", -1, err
		}
	}
	return "", -1, fs.ErrExist
}

func temporaryName() (string, error) {
	var suffix [16]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return ".cartulary-tmp-" + hex.EncodeToString(suffix[:]), nil
}

func requireBoundedRegularFile(fd int, maxBytes int64, requireSingleLink bool) (unix.Stat_t, error) {
	stat, err := requireRegularFile(fd, requireSingleLink)
	if err != nil {
		return unix.Stat_t{}, err
	}
	if stat.Size < 0 || stat.Size > maxBytes {
		return unix.Stat_t{}, fs.ErrInvalid
	}
	return stat, nil
}

func requireRegularFile(fd int, requireSingleLink bool) (unix.Stat_t, error) {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return unix.Stat_t{}, err
	}
	if err := requireAllowedRegularStat(stat, requireSingleLink); err != nil {
		return unix.Stat_t{}, err
	}
	if stat.Size < 0 {
		return unix.Stat_t{}, fs.ErrInvalid
	}
	return stat, nil
}

func requireAllowedRegularStat(stat unix.Stat_t, requireSingleLink bool) error {
	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return fs.ErrInvalid
	}
	if requireSingleLink && stat.Nlink != 1 {
		return fs.ErrInvalid
	}
	return nil
}

func identityFromStat(stat unix.Stat_t) objectIdentity {
	return objectIdentity{device: uint64(stat.Dev), inode: stat.Ino}
}

func metadataFromStat(stat unix.Stat_t) Metadata {
	return Metadata{
		Size:    stat.Size,
		Mode:    fs.FileMode(stat.Mode & 0o777),
		ModTime: time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec),
	}
}
