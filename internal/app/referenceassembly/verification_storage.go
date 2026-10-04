package referenceassembly

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"slices"
	"sync"
	"syscall"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/google/uuid"
)

var _ reference_data.VerificationStorage = (*RootStorage)(nil)

func (s *RootStorage) OpenStaged(ctx context.Context, ref reference_data.StagingRef) (reference_data.ContainerReader, int64, error) {
	if s == nil {
		return nil, 0, errors.New("reference pack storage unavailable")
	}
	return openContainer(ctx, s.temporary, ref.String())
}
func (s *RootStorage) OpenPublished(ctx context.Context, ref reference_data.StorageRef) (reference_data.ContainerReader, int64, error) {
	if s == nil {
		return nil, 0, errors.New("reference pack storage unavailable")
	}
	return openContainer(ctx, s.published, ref.String())
}
func openContainer(ctx context.Context, root *rootedfs.Root, raw string) (reference_data.ContainerReader, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	if root == nil {
		return nil, 0, errors.New("reference pack storage unavailable")
	}
	ref, err := rootedfs.ParseReference(raw)
	if err != nil {
		return nil, 0, err
	}
	file, metadata, err := root.OpenRegular(ref)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrInvalid) || errors.Is(err, syscall.ELOOP) || errors.Is(err, syscall.ENOTDIR) {
			return nil, 0, reference_data.ErrArtifactUnavailable
		}
		return nil, 0, err
	}
	return file, metadata.Size, nil
}
func (s *RootStorage) NewWorkspace(ctx context.Context) (result reference_data.VerificationWorkspace, resultErr error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s == nil || s.temporary == nil {
		return nil, errors.New("reference pack storage unavailable")
	}
	parentLease, err := lockWorkspaceParent(ctx, s.temporary)
	if err != nil {
		return nil, err
	}
	var lease io.Closer
	defer func() {
		resultErr = errors.Join(resultErr, parentLease.Close())
		if resultErr != nil && lease != nil {
			resultErr = errors.Join(resultErr, lease.Close())
			result = nil
		}
	}()
	dir := rootedfs.MustParseReference("reference-packs/verification/" + uuid.NewString())
	if err := s.temporary.MakePrivateDir(dir); err != nil {
		return nil, err
	}
	var acquired bool
	lease, acquired, err = s.temporary.TryLockDirectory(dir)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return nil, errors.New("reference pack: new workspace already leased")
	}
	return &verificationWorkspace{root: s.temporary, directory: dir, lease: lease, members: map[string]rootedfs.Reference{}}, nil
}

type verificationWorkspace struct {
	mu        sync.Mutex
	root      *rootedfs.Root
	directory rootedfs.Reference
	members   map[string]rootedfs.Reference
	closed    bool
	sealed    bool
	lease     io.Closer
}

func (w *verificationWorkspace) Write(ctx context.Context, name string, write func(io.Writer) error) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.sealed {
		return fs.ErrClosed
	}
	if _, err := rootedfs.ParseReference(name); err != nil {
		return err
	}
	if _, exists := w.members[name]; exists {
		return fs.ErrExist
	}
	// The untrusted logical name never becomes a physical filesystem segment.
	// Full logical names remain in the private in-memory inventory only.
	digest := sha256.Sum256([]byte(name))
	ref := rootedfs.MustParseReference(w.directory.String() + "/" + hex.EncodeToString(digest[:]))
	if err := w.root.CreateExclusive(ctx, ref, write); err != nil {
		return err
	}
	w.members[name] = ref
	return nil
}
func (w *verificationWorkspace) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if w.sealed {
		return nil, fs.ErrClosed
	}
	ref, exists := w.members[name]
	if !exists {
		return nil, fs.ErrNotExist
	}
	file, _, err := w.root.OpenRegular(ref)
	return file, err
}

func (w *verificationWorkspace) Remove(ctx context.Context, name string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if w.sealed {
		return fs.ErrClosed
	}
	ref, exists := w.members[name]
	if !exists {
		return fs.ErrNotExist
	}
	if err := w.root.RemoveRegular(ref); err != nil {
		return err
	}
	delete(w.members, name)
	return nil
}
func (w *verificationWorkspace) Close() (result error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.sealed = true
	defer func() {
		if w.lease != nil {
			result = errors.Join(result, w.lease.Close())
			w.lease = nil
		}
	}()
	parentLease, err := lockWorkspaceParent(context.Background(), w.root)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, parentLease.Close()) }()
	names := make([]string, 0, len(w.members))
	for name := range w.members {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if err := w.root.RemoveRegular(w.members[name]); err != nil && !errors.Is(err, fs.ErrNotExist) {
			result = errors.Join(result, err)
		} else {
			delete(w.members, name)
		}
	}
	if result == nil {
		if err := w.root.RemoveEmptyDir(w.directory); err != nil && !errors.Is(err, fs.ErrNotExist) {
			result = err
		}
	}
	w.closed = result == nil
	return result
}
