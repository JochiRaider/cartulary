package publication

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"golang.org/x/sys/unix"
)

// PublishTree publishes an entire immutable release in one no-replace rename.
// Matching replay verifies the complete tree; an incomplete or changed existing
// release is never repaired in place or treated as a successful publication.
func PublishTree(directory, relative string, members map[string][]byte) (result error) {
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
	id, err := randomSuffix()
	if err != nil {
		return err
	}
	stage := ".partial-" + id
	if err = root.Mkdir(stage, 0700); err != nil {
		return err
	}
	defer func() { result = errors.Join(result, root.RemoveAll(stage), syncParents(root, ".")) }()
	staged, err := root.OpenRoot(stage)
	if err != nil {
		return err
	}
	for name, b := range members {
		if err = Retain(staged, name, b, 0600); err != nil {
			staged.Close()
			return err
		}
	}
	if err = staged.Close(); err != nil {
		return err
	}
	// Pin both directory handles. Publication never re-resolves a host path
	// after checking confinement, and its no-replace rename is crash durable.
	sourceDirectory, err := root.Open(".")
	if err != nil {
		return err
	}
	defer sourceDirectory.Close()
	destinationDirectory, err := root.Open(filepath.Dir(relative))
	if err != nil {
		return err
	}
	defer destinationDirectory.Close()
	err = unix.Renameat2(int(sourceDirectory.Fd()), stage, int(destinationDirectory.Fd()), filepath.Base(relative), unix.RENAME_NOREPLACE)
	if err == nil {
		return syncParents(root, filepath.Dir(relative))
	}
	if !errors.Is(err, unix.EEXIST) {
		return err
	}
	existing, err := root.OpenRoot(relative)
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
		got, e := readRootRegular(existing, name, int64(len(want)))
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
	return syncParents(root, filepath.Dir(relative))
}

func LogicalMemberInventory(members map[string][]byte) []envelope.File {
	files := make([]envelope.File, 0, len(members))
	for name, b := range members {
		role, media := "payload", "application/x-ndjson"
		if name == "payload/descriptor.json" {
			media = "application/json"
		}
		if strings.HasPrefix(name, "notices/") {
			role, media = "notice", "text/plain"
		}
		files = append(files, envelope.File{Path: name, Role: role, MediaType: media, Size: int64(len(b)), SHA256: Digest(b)})
	}
	slices.SortFunc(files, func(a, b envelope.File) int { return strings.Compare(a.Path, b.Path) })
	return files
}
