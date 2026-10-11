//go:build linux

package harnessruntime

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

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
var jsonSecret = regexp.MustCompile(`(?i)"([^"]*(?:password|secret|token|cookie|dsn|access_key|private_key)[^"]*)"\s*:\s*"([^"]+)"`)
var assignmentSecret = regexp.MustCompile(`(?:PASSWORD|SECRET|TOKEN|COOKIE|DSN|ACCESS_KEY|PRIVATE_KEY)=([^\s]+)`)
var postgresSecret = regexp.MustCompile(`(?i)(?:postgres|postgresql)://[^\s/:]+:([^\s/@]+)@`)

func sensitiveJSONKey(key string) bool {
	k := strings.ToLower(key)
	for _, s := range []string{"password", "secret", "cookie", "dsn", "access_key", "private_key"} {
		if strings.Contains(k, s) {
			return true
		}
	}
	for {
		i := strings.Index(k, "token")
		if i < 0 {
			return false
		}
		k = k[i+5:]
		if !strings.HasPrefix(k, "s") {
			return true
		}
	}
}
func secretSyntax(text []byte) bool {
	for _, m := range jsonSecret.FindAllSubmatch(text, -1) {
		if sensitiveJSONKey(string(m[1])) && !bytes.HasPrefix(m[2], []byte("<redacted>")) {
			return true
		}
	}
	for _, re := range []*regexp.Regexp{assignmentSecret, postgresSecret} {
		for _, m := range re.FindAllSubmatch(text, -1) {
			if !bytes.HasPrefix(m[1], []byte("<redacted>")) {
				return true
			}
		}
	}
	return false
}
func scanContent(reader io.Reader, secrets []string) error {
	maximum := 256
	for _, s := range secrets {
		if len(s) >= 8 {
			maximum = max(maximum, len(s))
		}
	}
	buffer := make([]byte, 64*1024)
	overlap := []byte{}
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			b := append(overlap, buffer[:n]...)
			if secretSyntax(b) {
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
			return nil
		}
		if err != nil {
			return err
		}
	}
}
func scanRetained(root string, secrets []string) (map[string]any, error) {
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
	return map[string]any{"schema_id": "cartulary.harness_retained_secret_scan.v1", "status": "pass", "scanned_files": count, "policy_digest": bytesDigest([]byte(scanNames + scanSyntax))}, nil
}

// The scan is required even when optional evidence is absent. Failed scans
// never write a success record and remain secondary to an existing failure.
func publishScan(root string, secrets []string) error {
	v, err := scanRetained(root, secrets)
	if err != nil {
		return errors.Join(failed("artifact", "artifact_error"), err)
	}
	return writeEvidence(root, filepath.Base("retained-secret-scan.json"), "cartulary.harness_retained_secret_scan.v1", v, false)
}
