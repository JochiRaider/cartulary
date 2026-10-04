package referenceassembly

import (
	"context"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/google/uuid"
)

var workspaceParent = rootedfs.MustParseReference("reference-packs/verification")
var workspaceMember = regexp.MustCompile(`^(?:[a-f0-9]{64}|\.cartulary-tmp-[a-f0-9]{32})$`)

func lockWorkspaceParent(ctx context.Context, root *rootedfs.Root) (io.Closer, error) {
	if err := root.MakePrivateDir(workspaceParent); err != nil {
		return nil, err
	}
	return root.LockDirectory(ctx, workspaceParent)
}

// ReconcileWorkspaces collects only abandoned private verification scratch.
// A directory inode lease protects live work in every process sharing this
// root. Creation, ownership transfer and collection share the parent lease,
// so an unleased creation interval can never be mistaken for a crashed job.
// Authoritative objects, frozen inputs and incoming uploads are outside this
// namespace and cannot be addressed by the sweep.
func (s *RootStorage) ReconcileWorkspaces(ctx context.Context) (resultErr error) {
	if s == nil || s.temporary == nil {
		return errors.New("reference pack storage unavailable")
	}
	parent, err := lockWorkspaceParent(ctx, s.temporary)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, parent.Close()) }()
	return s.temporary.VisitDirectory(ctx, workspaceParent, func(entry rootedfs.DirectoryEntry) error {
		name := strings.TrimPrefix(entry.Reference.String(), workspaceParent.String()+"/")
		id, err := uuid.Parse(name)
		if !entry.Directory || err != nil || id == uuid.Nil || id.String() != name {
			return errors.New("reference pack: unexpected workspace inventory")
		}
		lease, acquired, err := s.temporary.TryLockDirectory(entry.Reference)
		if err != nil || !acquired {
			return err
		}
		return s.collectWorkspace(ctx, entry.Reference, lease)
	})
}

func (s *RootStorage) collectWorkspace(ctx context.Context, directory rootedfs.Reference, lease io.Closer) (resultErr error) {
	defer func() { resultErr = errors.Join(resultErr, lease.Close()) }()
	if err := s.temporary.VisitDirectory(ctx, directory, func(entry rootedfs.DirectoryEntry) error {
		name := strings.TrimPrefix(entry.Reference.String(), directory.String()+"/")
		if entry.Directory || !workspaceMember.MatchString(name) {
			return errors.New("reference pack: unexpected scratch member")
		}
		return s.temporary.RemoveRegular(entry.Reference)
	}); err != nil {
		return err
	}
	return s.temporary.RemoveEmptyDir(directory)
}
