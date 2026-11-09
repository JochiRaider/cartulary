package producer

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/publication"
)

const PackKey = "framework.veris"
const ReleaseRepository = VERISRepository
const DatasetFixtureRepository = "fixture.veris.dataset.v1"

func publisherScope(repository string) (publication.Scope, error) {
	switch repository {
	case ReleaseRepository:
		return publication.ReleaseScope(ReleaseRepository, PackKey)
	case DatasetFixtureRepository:
		return publication.FixtureScope(DatasetFixtureRepository, PackKey)
	default:
		return publication.Scope{}, errors.New("publisher identity outside this producer")
	}
}
func finishLogical(ctx context.Context, m envelope.Manifest, members packageSource, ledger, output string) (publication.LogicalReport, error) {
	scope, err := publisherScope(*m.Repository)
	if err != nil {
		return publication.LogicalReport{}, err
	}
	return publication.PublishLogical(ctx, publication.LogicalInput{Scope: scope, Manifest: m, Members: members, LedgerRoot: ledger, OutputRoot: output})
}

type PackageReport struct {
	SchemaID            string          `json:"schema_id"`
	ManifestSHA256      string          `json:"manifest_sha256"`
	PayloadSHA256       string          `json:"payload_sha256"`
	SourceProfileSHA256 string          `json:"source_profile_sha256"`
	SourceAuditSHA256   string          `json:"source_audit_sha256"`
	Builder             Builder         `json:"builder"`
	Objects             int64           `json:"object_count"`
	Relationships       int64           `json:"relationship_count"`
	Files               []envelope.File `json:"files"`
	LogicalPath         string          `json:"logical_path"`
}
