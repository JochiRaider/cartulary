package recoveryassembly

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/google/uuid"
)

type transferDirectory struct {
	*FilesystemStorage
	parent      *rootedfs.Root
	name        string
	destination string
	retained    bool
}

func NewTransferDirectory(output string) (recovery.TransferDirectory, error) {
	if !filepath.IsAbs(output) || filepath.Clean(output) != output {
		return nil, recovery.ErrInvalidTransfer
	}
	parent, err := rootedfs.Open(filepath.Dir(output))
	if err != nil {
		return nil, err
	}
	destination, err := rootedfs.ParseReference(filepath.Base(output))
	if err != nil {
		parent.Close()
		return nil, err
	}
	occupied, err := parent.Exists(destination)
	if err != nil || occupied {
		parent.Close()
		return nil, recovery.ErrInvalidTransfer
	}
	name := "transfer-stage-" + uuid.NewString()
	if err := parent.MakePrivateDir(rootedfs.MustParseReference(name)); err != nil {
		parent.Close()
		return nil, err
	}
	root, err := rootedfs.Open(filepath.Join(filepath.Dir(output), name))
	if err != nil {
		_ = parent.RemoveEmptyDir(rootedfs.MustParseReference(name))
		parent.Close()
		return nil, err
	}
	return &transferDirectory{FilesystemStorage: &FilesystemStorage{root: root}, parent: parent, name: name, destination: destination.String()}, nil
}

func OpenTransferDirectory(input string) (recovery.TransferDirectory, error) {
	if !filepath.IsAbs(input) || filepath.Clean(input) != input {
		return nil, recovery.ErrInvalidTransfer
	}
	if _, err := rootedfs.ParseReference(filepath.Base(input)); err != nil {
		return nil, err
	}
	parent, err := rootedfs.Open(filepath.Dir(input))
	if err != nil {
		return nil, err
	}
	root, err := rootedfs.Open(input)
	if err != nil {
		parent.Close()
		return nil, err
	}
	return &transferDirectory{FilesystemStorage: &FilesystemStorage{root: root}, parent: parent, name: filepath.Base(input), retained: true}, nil
}

func (d *transferDirectory) ListArtifacts(ctx context.Context) ([]string, error) {
	if err := d.root.Check(); err != nil {
		return nil, err
	}
	names := []string{}
	directories := []string{}
	count := 0
	var visit func(rootedfs.Reference) error
	visit = func(ref rootedfs.Reference) error {
		return d.parent.VisitDirectory(ctx, ref, func(entry rootedfs.DirectoryEntry) error {
			count++
			if len(strings.TrimPrefix(entry.Reference.String(), d.name+"/")) > 512 {
				return recovery.ErrInvalidTransfer
			}
			if count > recovery.TransferMaximumArtifacts*512+1 {
				return recovery.ErrInvalidTransfer
			}
			if entry.Directory {
				directories = append(directories, strings.TrimPrefix(entry.Reference.String(), d.name+"/"))
				return visit(entry.Reference)
			}
			name := strings.TrimPrefix(entry.Reference.String(), d.name+"/")
			if len(name) > 512 || len(names) >= recovery.TransferMaximumArtifacts+1 {
				return recovery.ErrInvalidTransfer
			}
			names = append(names, name)
			return nil
		})
	}
	if err := visit(rootedfs.MustParseReference(d.name)); err != nil {
		return nil, err
	}
	if err := d.root.Check(); err != nil {
		return nil, err
	}
	prefixes := map[string]bool{}
	for _, name := range names {
		for dir := filepath.Dir(name); dir != "."; dir = filepath.Dir(dir) {
			prefixes[dir] = true
		}
	}
	for _, dir := range directories {
		if !prefixes[dir] {
			return nil, recovery.ErrInvalidTransfer
		}
	}
	sort.Strings(names)
	return names, nil
}
func (d *transferDirectory) Publish(ctx context.Context) error {
	if d.retained || d.destination == "" {
		return recovery.ErrInvalidTransfer
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := d.root.PublishSiblingExclusive(rootedfs.MustParseReference(d.destination)); err != nil {
		return err
	}
	d.name = d.destination
	d.retained = true
	return nil
}
func (d *transferDirectory) Close() error {
	if d.parent == nil {
		return nil
	}
	var cleanupErr error
	if !d.retained {
		var remove func(rootedfs.Reference) error
		remove = func(ref rootedfs.Reference) error {
			err := d.parent.VisitDirectory(context.Background(), ref, func(entry rootedfs.DirectoryEntry) error {
				if entry.Directory {
					return remove(entry.Reference)
				}
				return d.parent.RemoveRegular(entry.Reference)
			})
			if err != nil {
				return err
			}
			return d.parent.RemoveEmptyDir(ref)
		}
		if err := d.root.Check(); err != nil {
			cleanupErr = err
		} else {
			cleanupErr = remove(rootedfs.MustParseReference(d.name))
		}
	}
	err := errors.Join(cleanupErr, d.FilesystemStorage.Close(), d.parent.Close())
	d.parent = nil
	return err
}

func OpenPortableJournal(rootPath string) (application.PortableJournal, error) {
	return NewFilesystemStorage(filepath.Join(rootPath, "portable-restore-journal"))
}
