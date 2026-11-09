package publication

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"maps"
)

type LogicalInput struct {
	Scope                  Scope
	Manifest               envelope.Manifest
	Members                map[string][]byte
	LedgerRoot, OutputRoot string
}
type LogicalReport struct {
	Manifest                                   envelope.Manifest
	ManifestSHA256, PayloadSHA256, LogicalPath string
}

// PublishLogical validates and reserves an immutable unsigned candidate. It
// does not grant distribution approval or create any signing authority.
func PublishLogical(ctx context.Context, in LogicalInput) (LogicalReport, error) {
	var result LogicalReport
	m := in.Manifest
	if !in.Scope.valid() || m.Repository == nil || *m.Repository != in.Scope.repository || m.Key != in.Scope.key || m.ProfileID != dataset.ProfileID || m.Summary.Objects == nil || m.Summary.Relationships == nil {
		return result, errors.New("explicit generic publisher identity required")
	}
	members := maps.Clone(in.Members)
	if _, exists := members["manifest.json"]; exists {
		return result, errors.New("manifest is composed by publication")
	}
	m.SchemaID = "cartulary.reference_pack_manifest.v2"
	m.Contract = envelope.ContractID
	m.Sequence = 1
	m.Files = LogicalMemberInventory(members)
	raw, err := canonicaljson.Marshal(m)
	if err != nil {
		return result, err
	}
	if _, err = envelope.DecodeManifest(raw, true); err != nil {
		return result, err
	}
	if _, err = dataset.Validate(ctx, memoryWorkspace(members), dataset.Expected{Objects: *m.Summary.Objects, Relationships: *m.Summary.Relationships, SourceArtifacts: len(m.Artifacts)}, nil); err != nil {
		return result, err
	}
	for _, file := range m.Files {
		if file.Role == "notice" {
			if err = envelope.ValidateNotice(members[file.Path]); err != nil {
				return result, err
			}
		}
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	reservation, err := ReserveRelease(in.LedgerRoot, m.Version, Digest(raw), in.Scope)
	if err != nil {
		return result, err
	}
	m.Sequence = reservation.Sequence
	raw, err = canonicaljson.Marshal(m)
	if err != nil {
		return result, err
	}
	payload, err := envelope.PayloadDigest(m.Files)
	if err != nil {
		return result, err
	}
	result = LogicalReport{Manifest: m, ManifestSHA256: Digest(raw), PayloadSHA256: payload, LogicalPath: "logical/" + Digest(raw)}
	members["manifest.json"] = raw
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if err = PublishTree(in.OutputRoot, result.LogicalPath, members); err != nil {
		return result, err
	}
	return result, nil
}
