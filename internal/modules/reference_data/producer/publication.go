package producer

import (
	"bytes"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sys/unix"
)

// publishTree publishes an entire immutable release in one no-replace rename.
// Matching replay verifies the complete tree; an incomplete or changed existing
// release is never repaired in place or treated as a successful publication.
func publishTree(directory, relative string, members map[string][]byte) error {
	if !validRelative(relative) {
		return errors.New("invalid publication path")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	if err = noSymlinks(root, relative, true); err != nil {
		return err
	}
	if err = root.MkdirAll(filepath.Dir(relative), 0700); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(directory, ".partial-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	staged, err := os.OpenRoot(stage)
	if err != nil {
		return err
	}
	for name, b := range members {
		if err = Retain(staged, name, b, 0600); err != nil {
			staged.Close()
			return err
		}
	}
	staged.Close()
	destination := filepath.Join(directory, relative)
	err = unix.Renameat2(unix.AT_FDCWD, stage, unix.AT_FDCWD, destination, unix.RENAME_NOREPLACE)
	if !errors.Is(err, unix.EEXIST) {
		return err
	}
	existing, err := os.OpenRoot(destination)
	if err != nil {
		return err
	}
	defer existing.Close()
	seen := map[string]bool{}
	err = fs.WalkDir(existing.FS(), ".", func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return errors.New("nonregular publication member")
		}
		want, ok := members[name]
		if !ok {
			return errors.New("unexpected publication member")
		}
		got, e := existing.ReadFile(name)
		if e != nil {
			return e
		}
		if !bytes.Equal(want, got) {
			return errors.New("immutable publication collision")
		}
		seen[name] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(members) {
		return errors.New("incomplete immutable publication")
	}
	return nil
}

func logicalMemberInventory(members map[string][]byte) []packformat.File {
	files := make([]packformat.File, 0, len(members))
	for name, b := range members {
		role, media := "payload", "application/x-ndjson"
		if strings.HasPrefix(name, "notices/") {
			role, media = "notice", "text/plain"
		}
		files = append(files, packformat.File{Path: name, Role: role, MediaType: media, Size: int64(len(b)), SHA256: Digest(b)})
	}
	slices.SortFunc(files, func(a, b packformat.File) int { return strings.Compare(a.Path, b.Path) })
	return files
}
