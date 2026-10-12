//go:build linux

package harnesssession

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"golang.org/x/sys/unix"
)

// Preserve the scan-policy identity across the port. These are the existing
// owner patterns in their original ECMAScript notation; Go predicates below
// implement its look-ahead exclusions without adding a second schema reader.
const scanNames = `/^(?:stack|suite|test-services).*\.env$/iu
/^postgres\.recovery\.dsn$/u
/^performance-fixture-runtime\.json$/u
/^suite-environment\.json$/u
/^suite-lease\.json$/u
/^stack\.lease$/u
/^test-route-token$/u
/key[-_.]?ring/iu`
const scanSyntax = `/"[^"]*(?:password|secret|token(?!s)|cookie|dsn|access_key|private_key)[^"]*"\s*:\s*"(?!<redacted>)[^"]+"/iu
/(?:PASSWORD|SECRET|TOKEN(?!S)|COOKIE|DSN|ACCESS_KEY|PRIVATE_KEY)=(?!<redacted>)[^\s]+/u
/(?:postgres|postgresql):\/\/[^\s/:]+:(?!<redacted>)[^\s/@]+@/iu`

var forbiddenName = regexp.MustCompile(`(?i:^(?:stack|suite|test-services).*\.env$)|^postgres\.recovery\.dsn$|^performance-fixture-runtime\.json$|^suite-environment\.json$|^suite-lease\.json$|^stack\.lease$|^test-route-token$|(?i:key[-_.]?ring)`)

func scanContent(reader io.Reader, secrets []string) error {
	maximum := 256
	for _, s := range secrets {
		if len(s) >= 8 {
			maximum = max(maximum, len(s))
		}
	}
	buffer := make([]byte, 64*1024)
	overlap := []byte{}
	syntax := newSecretMatcher()
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			b := append(overlap, buffer[:n]...)
			if syntax.feed(buffer[:n], false) {
				return harnesslocal.ErrUnsafe
			}
			for _, s := range secrets {
				if len(s) >= 8 && bytes.Contains(b, []byte(s)) {
					return harnesslocal.ErrUnsafe
				}
			}
			overlap = append([]byte{}, b[max(0, len(b)-maximum):]...)
		}
		if err == io.EOF {
			if syntax.feed(nil, true) {
				return harnesslocal.ErrUnsafe
			}
			return nil
		}
		if err != nil {
			return err
		}
	}
}
func ScanRetained(root string, secrets []string) (map[string]any, error) {
	dir, err := harnesslocal.OpenDirectory(root, true)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	count := 0
	var visit func(*os.File) error
	visit = func(parent *os.File) error {
		names, err := parent.Readdirnames(-1)
		if err != nil {
			return err
		}
		slices.Sort(names)
		for _, name := range names {
			var before unix.Stat_t
			if err := unix.Fstatat(int(parent.Fd()), name, &before, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				return err
			}
			kind := before.Mode & unix.S_IFMT
			if before.Uid != uint32(os.Getuid()) {
				return harnesslocal.ErrUnsafe
			}
			if kind != unix.S_IFDIR && kind != unix.S_IFREG {
				_ = unix.Unlinkat(int(parent.Fd()), name, 0)
				return harnesslocal.ErrUnsafe
			}
			if kind == unix.S_IFREG && forbiddenName.MatchString(name) {
				_ = unix.Unlinkat(int(parent.Fd()), name, 0)
				return harnesslocal.ErrUnsafe
			}
			flags := unix.O_RDONLY | unix.O_NOFOLLOW | unix.O_NONBLOCK | unix.O_CLOEXEC
			if kind == unix.S_IFDIR {
				flags |= unix.O_DIRECTORY
			}
			fd, err := unix.Openat(int(parent.Fd()), name, flags, 0)
			if err != nil {
				return err
			}
			f := os.NewFile(uintptr(fd), name)
			var opened unix.Stat_t
			err = unix.Fstat(fd, &opened)
			if err != nil || opened.Dev != before.Dev || opened.Ino != before.Ino || opened.Uid != before.Uid || opened.Mode != before.Mode {
				f.Close()
				return harnesslocal.ErrUnsafe
			}
			mode := uint32(0600)
			if kind == unix.S_IFDIR {
				mode = 0700
			}
			if before.Mode&0777 != mode {
				_ = unix.Fchmod(fd, mode)
				f.Close()
				return harnesslocal.ErrUnsafe
			}
			if kind == unix.S_IFDIR {
				err = visit(f)
			} else if before.Nlink != 1 {
				err = harnesslocal.ErrUnsafe
			} else {
				count++
				err = scanContent(f, secrets)
			}
			var after, current unix.Stat_t
			statErr := unix.Fstat(fd, &after)
			pathErr := unix.Fstatat(int(parent.Fd()), name, &current, unix.AT_SYMLINK_NOFOLLOW)
			stable := statErr == nil && pathErr == nil && before.Ino == after.Ino && before.Dev == after.Dev && before.Ino == current.Ino && before.Dev == current.Dev && before.Mode == after.Mode && before.Uid == after.Uid && before.Nlink == after.Nlink && before.Size == after.Size && before.Mtim == after.Mtim && before.Ctim == after.Ctim
			// Only remove the inspected unsafe file, never a replacement path.
			if err != nil && kind == unix.S_IFREG && stable {
				_ = unix.Unlinkat(int(parent.Fd()), name, 0)
			}
			closeErr := f.Close()
			if err != nil {
				return err
			}
			if !stable {
				return harnesslocal.ErrUnsafe
			}
			if closeErr != nil {
				return closeErr
			}
		}
		return nil
	}
	if err = visit(dir); err != nil {
		return nil, err
	}
	if err = harnesslocal.CheckDirectory(dir, root); err != nil {
		return nil, err
	}
	return map[string]any{"schema_id": "cartulary.harness_retained_secret_scan.v1", "status": "pass", "scanned_files": count, "policy_digest": fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(scanNames+scanSyntax)))}, nil
}
