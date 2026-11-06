// Package fixtureartifacts admits confined, immutable test inputs. It knows
// nothing about source owners, manifest schemas, scenarios, or product results.
package fixtureartifacts

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type File struct {
	Path   string `json:"logical_path"`
	Size   int64  `json:"size_bytes"`
	SHA256 string `json:"sha256"`
}

func SafePath(path string) bool {
	return path != "" && fs.ValidPath(path) && path != "." && !strings.ContainsAny(path, "\\:\x00") && filepath.IsLocal(path)
}

func RejectSymlinkComponents(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	current := string(filepath.Separator)
	for _, part := range strings.Split(strings.TrimPrefix(abs, current), string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("unsafe fixture directory")
		}
	}
	return nil
}

func SecureDirectory(path string) error {
	if !filepath.IsAbs(path) {
		return errors.New("fixture run root must be absolute")
	}
	current := string(filepath.Separator)
	for _, part := range strings.Split(strings.TrimPrefix(filepath.Clean(path), current), string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(current, 0700); err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
			info, err = os.Lstat(current)
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("unsafe run directory")
		}
	}
	return nil
}

func OpenRegular(root *os.Root, path string) (*os.File, error) {
	if !SafePath(path) {
		return nil, errors.New("unsafe fixture path")
	}
	parts := strings.Split(path, "/")
	for i := range parts {
		info, err := root.Lstat(strings.Join(parts[:i+1], "/"))
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 || i < len(parts)-1 && !info.IsDir() || i == len(parts)-1 && !info.Mode().IsRegular() {
			return nil, errors.New("fixture path must contain only directories and a regular file")
		}
	}
	file, err := root.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, errors.New("nonregular fixture file")
	}
	return file, nil
}

func ReadRegular(root *os.Root, path string, limit int64) ([]byte, error) {
	if limit < 0 || limit == 1<<63-1 {
		return nil, errors.New("invalid fixture read bound")
	}
	file, err := OpenRegular(root, path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("fixture document exceeds admission bound")
	}
	return data, nil
}

func CopyVerified(root *os.Root, expected File, inputs string) error {
	if expected.Size < 0 || expected.Size == 1<<63-1 || len(expected.SHA256) != 64 {
		return errors.New("invalid fixture descriptor")
	}
	source, err := OpenRegular(root, expected.Path)
	if err != nil {
		return err
	}
	defer source.Close()
	destination := filepath.Join(inputs, expected.Path)
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	hash := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(out, hash), io.LimitReader(source, expected.Size+1))
	if err := errors.Join(copyErr, out.Close()); err != nil {
		return err
	}
	if n != expected.Size || hex.EncodeToString(hash.Sum(nil)) != expected.SHA256 {
		return errors.New("fixture size or hash mismatch")
	}
	return nil
}
