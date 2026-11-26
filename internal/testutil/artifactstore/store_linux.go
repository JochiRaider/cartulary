//go:build linux

// Package artifactstore is the harness build-artifact store (TH-HARNESS
// build-artifact profiles, P5-3b). A producer's outputs live in an immutable
// root under objects/, named by the digest of the root's manifest. A key
// record under keys/<producer>/ names the root a producer key built. The
// conventional paths consumers read, such as build/bin/server, are published
// views: copies of the root's files, each swapped into place by rename under
// a lock per view. A view is never a hard link, because product code opens
// its own files only when they have a single link.
package artifactstore

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const (
	manifestSchemaID = "cartulary.artifact_store_root.v1"
	recordSchemaID   = "cartulary.artifact_store_key.v1"
)

// unrecordedGrace is how long a root no key record names is kept after it was
// sealed.
const unrecordedGrace = time.Hour

// Store is a build-artifact store rooted at Root.
type Store struct{ Root string }

// File is one file of a root: its slash-separated path, permission bits and
// SHA-256 digest.
type File struct {
	Path   string `json:"path"`
	Mode   uint32 `json:"mode"`
	SHA256 string `json:"sha256"`
}

type manifest struct {
	SchemaID string `json:"schema_id"`
	Files    []File `json:"files"`
}

type record struct {
	SchemaID string `json:"schema_id"`
	Producer string `json:"producer"`
	Key      string `json:"key"`
	Root     string `json:"root"`
}

// Producer builds one root. ID names the producer in key records; Key covers
// every input its outputs depend on; Build writes the outputs into staging.
type Producer interface {
	ID() string
	Key() (string, error)
	Build(staging string) error
}

// Controls are the public cache controls (D28): Disable neither reads nor
// records keys, and Force rebuilds and records.
type Controls struct{ Disable, Force bool }

// Result reports the root a Produce call published and whether it was reused.
type Result struct {
	Root   string
	Reused bool
}

func (s Store) path(parts ...string) string {
	return filepath.Join(append([]string{s.Root}, parts...)...)
}

// RootPath is the directory of root id.
func (s Store) RootPath(id string) string { return s.path("objects", id) }

// Produce publishes the root for p's current key: the recorded root when it
// still verifies, otherwise a root p builds and the store seals. views maps a
// root file's path to the path it is published at.
func (s Store) Produce(p Producer, views map[string]string, controls Controls) (Result, error) {
	key, err := p.Key()
	if err != nil {
		return Result{}, fmt.Errorf("artifact key for %s: %w", p.ID(), err)
	}
	unlock, err := s.lock("key", p.ID(), key)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	if !controls.Disable && !controls.Force {
		if id, ok := s.lookup(p.ID(), key); ok {
			return Result{Root: id, Reused: true}, s.publish(id, views)
		}
	}
	staging, err := s.stage()
	if err != nil {
		return Result{}, err
	}
	defer removeTree(staging)
	if err := p.Build(staging); err != nil {
		return Result{}, err
	}
	id, err := s.seal(staging)
	if err != nil {
		return Result{}, err
	}
	if !controls.Disable {
		if err := s.record(p.ID(), key, id); err != nil {
			return Result{}, err
		}
	}
	return Result{Root: id}, s.publish(id, views)
}

func (s Store) stage() (string, error) {
	if err := os.MkdirAll(s.path("staging"), 0o700); err != nil {
		return "", err
	}
	return os.MkdirTemp(s.path("staging"), "root-")
}

// inventory lists every regular file under root in path order. A link or any
// other kind of entry makes the root unsealable.
func inventory(root string) ([]File, error) {
	files := []File{}
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || name == root {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("artifact root entry %s is not a regular file", name)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		digest, err := fileDigest(name)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		files = append(files, File{Path: filepath.ToSlash(relative), Mode: uint32(info.Mode().Perm()), SHA256: digest})
		return nil
	})
	slices.SortFunc(files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	return files, err
}

func fileDigest(name string) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

// seal makes staging read-only, names it by its manifest's digest and moves it
// into objects/. An existing root with that name is reused when it still
// verifies and quarantined when it does not.
func (s Store) seal(staging string) (string, error) {
	if err := filepath.WalkDir(staging, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return os.Chmod(name, info.Mode().Perm()&^0o222)
	}); err != nil {
		return "", err
	}
	files, err := inventory(staging)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", errors.New("an artifact root holds no file")
	}
	data, err := json.Marshal(manifest{SchemaID: manifestSchemaID, Files: files})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	id := hex.EncodeToString(sum[:])
	if err := os.MkdirAll(s.path("objects"), 0o700); err != nil {
		return "", err
	}
	if err := writeAtomic(s.path("objects", id+".json"), data); err != nil {
		return "", err
	}
	for attempt := 0; attempt < 2; attempt++ {
		// Moving a directory rewrites its parent link, which needs write
		// permission on it; its directories become read-only once in place.
		err = unix.Renameat2(unix.AT_FDCWD, staging, unix.AT_FDCWD, s.RootPath(id), unix.RENAME_NOREPLACE)
		if err == nil {
			return id, setDirectoryModes(s.RootPath(id), 0o555)
		}
		if !errors.Is(err, unix.EEXIST) && !errors.Is(err, unix.ENOTEMPTY) {
			return "", err
		}
		if s.verify(id) {
			return id, nil
		}
		// A root altered after sealing never supplies an artifact again.
		if err := os.Rename(s.RootPath(id), s.path("objects", ".quarantine-"+id+"-"+token())); err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("artifact root %s could not be sealed", id)
}

// manifest reads root id's manifest, which its name digests.
func (s Store) manifest(id string) (manifest, error) {
	var m manifest
	data, err := os.ReadFile(s.path("objects", id+".json"))
	if err != nil {
		return m, err
	}
	if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) != id {
		return m, fmt.Errorf("artifact manifest %s does not match its name", id)
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, err
	}
	if m.SchemaID != manifestSchemaID {
		return m, fmt.Errorf("artifact manifest %s has schema %q", id, m.SchemaID)
	}
	return m, nil
}

// verify reports whether root id still holds exactly its manifest's files.
func (s Store) verify(id string) bool {
	m, err := s.manifest(id)
	if err != nil {
		return false
	}
	files, err := inventory(s.RootPath(id))
	return err == nil && slices.Equal(files, m.Files)
}

func (s Store) recordPath(producer, key string) string {
	return s.path("keys", producer, key+".json")
}

// lookup returns the root recorded for key when it still verifies. A reused
// record is touched, so collection keeps what is in use.
func (s Store) lookup(producer, key string) (string, bool) {
	data, err := os.ReadFile(s.recordPath(producer, key))
	if err != nil {
		return "", false
	}
	var r record
	if json.Unmarshal(data, &r) != nil || r.SchemaID != recordSchemaID || r.Producer != producer || r.Key != key || !s.verify(r.Root) {
		return "", false
	}
	now := time.Now()
	_ = os.Chtimes(s.recordPath(producer, key), now, now)
	return r.Root, true
}

func (s Store) record(producer, key, id string) error {
	if err := os.MkdirAll(filepath.Dir(s.recordPath(producer, key)), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(record{SchemaID: recordSchemaID, Producer: producer, Key: key, Root: id})
	if err != nil {
		return err
	}
	return writeAtomic(s.recordPath(producer, key), append(data, '\n'))
}

// publish copies each root file to its view, in root file order, so a
// receipt named after its binary is published after the binary. A view that
// already holds its root file's manifest digest and view mode is left alone.
func (s Store) publish(id string, views map[string]string) error {
	m, err := s.manifest(id)
	if err != nil {
		return err
	}
	published := 0
	for _, file := range m.Files {
		view, ok := views[file.Path]
		if !ok {
			continue
		}
		published++
		if err := s.publishView(filepath.Join(s.RootPath(id), filepath.FromSlash(file.Path)), view, file); err != nil {
			return err
		}
	}
	if published != len(views) {
		return fmt.Errorf("artifact root %s lacks a file its views name", id)
	}
	return nil
}

func (s Store) publishView(source, view string, file File) error {
	view, err := filepath.Abs(view)
	if err != nil {
		return err
	}
	unlock, err := s.lock("view", view)
	if err != nil {
		return err
	}
	defer unlock()
	// A view is the owner's writable copy; only the root is sealed.
	mode := fs.FileMode(file.Mode).Perm() | 0o200
	if current(view, mode, file.SHA256) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(view), 0o755); err != nil {
		return err
	}
	temporary := filepath.Join(filepath.Dir(view), "."+filepath.Base(view)+".view-"+token())
	if err := copyFile(source, temporary, mode, file.SHA256); err != nil {
		return err
	}
	if err := os.Rename(temporary, view); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}

// current reports whether view is a single-link regular file with the view
// mode and the root file's digest.
func current(view string, mode fs.FileMode, digest string) bool {
	info, err := os.Lstat(view)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != mode {
		return false
	}
	var stat unix.Stat_t
	if unix.Lstat(view, &stat) != nil || stat.Nlink != 1 {
		return false
	}
	got, err := fileDigest(view)
	return err == nil && got == digest
}

// Collect removes key records unused for maxAge, then roots no record names,
// and staging directories older than maxAge. Views are copies, so a removed
// root never takes a published file with it. It removes
// at most limit entries in all, so a build never waits on a long sweep.
func (s Store) Collect(now time.Time, maxAge time.Duration, limit int) (int, error) {
	removed := 0
	referenced := map[string]bool{}
	records, _ := filepath.Glob(s.path("keys", "*", "*.json"))
	for _, name := range records {
		info, err := os.Stat(name)
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > maxAge && removed < limit {
			if os.Remove(name) == nil {
				removed++
				continue
			}
		}
		var r record
		if data, err := os.ReadFile(name); err == nil && json.Unmarshal(data, &r) == nil {
			referenced[r.Root] = true
		}
	}
	roots, _ := os.ReadDir(s.path("objects"))
	for _, entry := range roots {
		id := entry.Name()
		if !entry.IsDir() || referenced[id] || removed >= limit {
			continue
		}
		// A quarantined root goes after the record age. An unrecorded root,
		// such as one built with the cache disabled, stays for a grace period
		// after sealing, so its caller can still copy it.
		age := unrecordedGrace
		manifest := s.path("objects", id+".json")
		if strings.HasPrefix(id, ".quarantine-") {
			age, manifest = maxAge, s.RootPath(id)
		}
		if info, err := os.Stat(manifest); err == nil && now.Sub(info.ModTime()) <= age {
			continue
		}
		if removeTree(s.RootPath(id)) == nil {
			if !strings.HasPrefix(id, ".") {
				_ = os.Remove(s.path("objects", id+".json"))
			}
			removed++
		}
	}
	staged, _ := os.ReadDir(s.path("staging"))
	for _, entry := range staged {
		info, err := entry.Info()
		if err != nil || now.Sub(info.ModTime()) <= maxAge || removed >= limit {
			continue
		}
		if removeTree(s.path("staging", entry.Name())) == nil {
			removed++
		}
	}
	return removed, nil
}

func (s Store) lock(kind string, parts ...string) (func(), error) {
	if err := os.MkdirAll(s.path("locks"), 0o700); err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(kind + "\x00" + strings.Join(parts, "\x00")))
	return lockFile(s.path("locks", kind+"-"+hex.EncodeToString(sum[:16])+".lock"))
}

// lockFile holds an exclusive lock on the file name until the returned
// function is called.
func lockFile(name string) (func(), error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX); err != nil {
		file.Close()
		return nil, err
	}
	return func() {
		_ = unix.Flock(int(file.Fd()), unix.LOCK_UN)
		file.Close()
	}, nil
}

func writeAtomic(name string, data []byte) error {
	if existing, err := os.ReadFile(name); err == nil && bytes.Equal(existing, data) {
		return nil
	}
	temporary := name + ".tmp-" + token()
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(temporary, name); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}

// copyFile copies source to a new target and refuses a copy whose bytes do
// not have digest, so a root altered after its verification never reaches a
// view.
func copyFile(source, target string, mode fs.FileMode, digest string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	sum := sha256.New()
	_, err = io.Copy(io.MultiWriter(out, sum), in)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err == nil && hex.EncodeToString(sum.Sum(nil)) != digest {
		err = fmt.Errorf("artifact file %s changed after verification", source)
	}
	if err != nil {
		_ = os.Remove(target)
	}
	return err
}

func setDirectoryModes(root string, mode fs.FileMode) error {
	directories := []string{}
	if err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err == nil && entry.IsDir() {
			directories = append(directories, name)
		}
		return err
	}); err != nil {
		return err
	}
	// Children first, so a read-only parent never blocks its child.
	slices.Reverse(directories)
	for _, name := range directories {
		if err := os.Chmod(name, mode); err != nil {
			return err
		}
	}
	return nil
}

// removeTree removes a sealed or staged tree, restoring write permission
// first, since sealed directories are read-only.
func removeTree(root string) error {
	_ = filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err == nil && entry.IsDir() {
			_ = os.Chmod(name, 0o700)
		}
		return nil
	})
	return os.RemoveAll(root)
}

func token() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
