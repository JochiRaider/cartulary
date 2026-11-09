// Package verification provides standalone generic-pack qualification. Trust is
// explicit input; successful verification neither installs trust nor activates a pack.
package verification

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"io"
	"slices"
	"strings"
	"time"
)

type Workspace interface {
	envelope.Destination
	dataset.Source
}
type Result struct {
	Manifest                      envelope.Manifest
	ManifestSHA256, PayloadSHA256 string
	Trust                         envelope.TrustProposal
	Dataset                       dataset.Result
}

// Dataset extracts to a caller-owned private workspace, verifies offline trust
// at the supplied trusted instant, and admits the complete generic dataset.
// The caller removes the workspace on every outcome. No index is published.
func Dataset(ctx context.Context, archive io.ReaderAt, size int64, roots map[string]envelope.TrustSnapshot, at time.Time, workspace Workspace) (Result, error) {
	var result Result
	inventory, err := envelope.Extract(ctx, archive, size, envelope.DefaultArchiveLimits(), workspace)
	if err != nil {
		return result, err
	}
	read := func(name string, limit int64) ([]byte, error) {
		member, ok := inventory[name]
		if !ok || member.Size > limit {
			return nil, errors.New("missing or oversized verification member")
		}
		stream, err := workspace.Open(ctx, name)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(io.LimitReader(stream, limit+1))
		err = errors.Join(err, stream.Close())
		if err != nil {
			return nil, err
		}
		if int64(len(b)) != member.Size || envelope.Digest(b) != member.SHA256 {
			return nil, errors.New("verification member changed")
		}
		return b, nil
	}
	hint, err := read("bundle.json", 16384)
	if err != nil {
		return result, err
	}
	repository, err := envelope.DecodeHint(hint)
	if err != nil {
		return result, err
	}
	root, ok := roots[repository]
	if !ok {
		return result, errors.New("publisher bootstrap unavailable")
	}
	metadata := map[string][]byte{}
	paths := []string{}
	for name := range inventory {
		if strings.HasPrefix(name, "metadata/") {
			paths = append(paths, name)
		}
	}
	slices.Sort(paths)
	var total int64
	for _, name := range paths {
		size := inventory[name].Size
		if size > 2097152 || size > 8388608-total {
			return result, errors.New("trust metadata resource limit")
		}
		total += size
		b, err := read(name, 2097152)
		if err != nil {
			return result, err
		}
		metadata[name] = b
	}
	result.Trust, err = envelope.VerifyTrust(ctx, metadata, inventory, repository, at, root)
	if err != nil {
		return result, err
	}
	raw, err := read("manifest.json", 1048576)
	if err != nil {
		return result, err
	}
	result.Manifest, err = envelope.DecodeManifest(raw, true)
	if err != nil {
		return result, err
	}
	manifest := result.Manifest
	if manifest.ProfileID != dataset.ProfileID || manifest.Summary.Objects == nil || manifest.Summary.Relationships == nil {
		return result, errors.New("generic dataset profile required")
	}
	if err = envelope.ValidateInventory(manifest, inventory, true); err != nil {
		return result, err
	}
	result.ManifestSHA256 = envelope.Digest(raw)
	result.PayloadSHA256, err = envelope.PayloadDigest(manifest.Files)
	if err != nil {
		return result, err
	}
	if err = envelope.TrustBindingFindings(manifest, result.ManifestSHA256, result.PayloadSHA256, result.Trust, func(envelope.Finding) error { return errors.New("manifest trust binding mismatch") }); err != nil {
		return result, err
	}
	if err = envelope.DisallowedContentFindings(manifest, func(envelope.Finding) error { return errors.New("disallowed inventory content") }); err != nil {
		return result, err
	}
	for _, file := range manifest.Files {
		if file.Role != "notice" {
			continue
		}
		stream, err := workspace.Open(ctx, file.Path)
		if err != nil {
			return result, err
		}
		err = errors.Join(envelope.ValidateNoticeStream(ctx, stream), stream.Close())
		if err != nil {
			return result, err
		}
	}
	result.Dataset, err = dataset.Validate(ctx, workspace, dataset.Expected{Objects: *manifest.Summary.Objects, Relationships: *manifest.Summary.Relationships, SourceArtifacts: len(manifest.Artifacts)}, nil)
	return result, err
}
