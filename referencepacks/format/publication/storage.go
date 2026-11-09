// Package publication owns offline signing and immutable release output. Runtime
// consumers must not import this package.
package publication

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func Digest(b []byte) string { return envelope.Digest(b) }
func validBuildTime(value string) bool {
	at, err := time.Parse("2006-01-02T15:04:05Z", value)
	return err == nil && at.Year() >= 1 && at.Format("2006-01-02T15:04:05Z") == value
}

func DecodeClosed(b []byte, target any) error {
	value, err := canonicaljson.DecodeStrict(b)
	if err != nil {
		return errors.New("invalid JSON structure")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return errors.New("invalid closed input shape")
	}
	// Every producer contract member is required, including explicit nulls.
	// Re-encoding detects omitted struct fields without materializing defaults.
	original, err := canonicaljson.Marshal(value)
	if err != nil {
		return err
	}
	decoded, err := canonicaljson.Marshal(target)
	if err != nil {
		return err
	}
	if !bytes.Equal(original, decoded) {
		return errors.New("missing required producer member")
	}
	return nil
}

func randomSuffix() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Retain uses no-replace publication. Existing differing bytes and symlinks are
// rejected. os.Root confines every relative operation to the selected store.
func Retain(root *os.Root, name string, b []byte, mode os.FileMode) (result error) {
	if !validRelative(name) {
		return errors.New("invalid evidence path")
	}
	if err := noSymlinks(root, name, true); err != nil {
		return err
	}
	if st, err := root.Lstat(name); err == nil {
		if !st.Mode().IsRegular() {
			return errors.New("evidence is not a regular file")
		}
		old, err := readRootRegular(root, name, int64(len(b)))
		if err != nil || !bytes.Equal(old, b) {
			return errors.New("immutable evidence collision")
		}
		return syncParents(root, path.Dir(name))
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := root.MkdirAll(path.Dir(name), 0700); err != nil {
		return err
	}
	if err := noSymlinks(root, path.Dir(name), false); err != nil {
		return err
	}
	// The private temporary file is not an admitted artifact. Link publishes
	// atomically without replacing an existing name, including across producers.
	id, err := randomSuffix()
	if err != nil {
		return err
	}
	temp := path.Join(path.Dir(name), ".partial-"+id)
	f, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer func() {
		removeErr := root.Remove(temp)
		if errors.Is(removeErr, os.ErrNotExist) {
			removeErr = nil
		}
		result = errors.Join(result, removeErr, syncParents(root, path.Dir(name)))
	}()
	_, writeErr := f.Write(b)
	syncErr := f.Sync()
	closeErr := f.Close()
	if err = errors.Join(writeErr, syncErr, closeErr); err != nil {
		return err
	}
	if err = root.Link(temp, name); errors.Is(err, os.ErrExist) {
		return Retain(root, name, b, mode)
	}
	return err
}

// Persist the name and every newly created parent before reporting success.
// Handles remain confined to the selected store while directories are synced.
func syncParents(root *os.Root, directory string) error {
	for {
		file, err := root.Open(directory)
		if err != nil {
			return err
		}
		err = errors.Join(file.Sync(), file.Close())
		if err != nil {
			return err
		}
		if directory == "." {
			return nil
		}
		directory = path.Dir(directory)
	}
}

func readRootDirectory(root *os.Root) ([]os.DirEntry, error) {
	file, err := root.Open(".")
	if err != nil {
		return nil, err
	}
	entries, readErr := file.ReadDir(-1)
	return entries, errors.Join(readErr, file.Close())
}

func validRelative(p string) bool {
	return p != "" && p == path.Clean(p) && p != "." && !strings.HasPrefix(p, "/") && !strings.Contains(p, "\\") && p != ".." && !strings.HasPrefix(p, "../")
}
func noSymlinks(root *os.Root, name string, allowMissing bool) error {
	parts := strings.Split(name, "/")
	for i := range parts {
		st, err := root.Lstat(strings.Join(parts[:i+1], "/"))
		if errors.Is(err, os.ErrNotExist) && allowMissing {
			return nil
		}
		if err != nil {
			return err
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return errors.New("symlinked evidence path")
		}
	}
	return nil
}

// ReadRegular bounds explicit inputs before allocation. It rejects symlinked
// final components and reads through the checked open handle.
func ReadRegular(name string, limit int64) ([]byte, error) {
	if limit < 0 || limit > 536870912 {
		return nil, errors.New("invalid input limit")
	}
	root, err := os.OpenRoot(filepath.Dir(name))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return readRootRegular(root, filepath.Base(name), limit)
}

func readRootRegular(root *os.Root, name string, limit int64) ([]byte, error) {
	if !validRelative(name) || limit < 0 || limit > 536870912 {
		return nil, errors.New("invalid bounded relative input")
	}
	if err := noSymlinks(root, name, false); err != nil {
		return nil, err
	}
	// O_NONBLOCK prevents a substituted FIFO from blocking before fstat rejects it.
	file, err := root.OpenFile(name, os.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("bounded regular input required")
	}
	b, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit || int64(len(b)) != info.Size() {
		return nil, errors.New("input size changed or exceeded")
	}
	return b, nil
}
