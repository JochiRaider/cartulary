package producer

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/publication"
	"path/filepath"
	"strings"
)

// ReleaseApproval is a recorded licensing authority decision about exact output.
// Possession of a signing key and successful source acquisition are not approval.
type ReleaseApproval struct {
	SchemaID                 string `json:"schema_id"`
	Status                   string `json:"status"`
	DecisionRef              string `json:"decision_ref"`
	Authority                string `json:"authority"`
	ManifestSHA256           string `json:"manifest_sha256"`
	SourceProfileSHA256      string `json:"source_profile_sha256"`
	LicenseDecisionSHA256    string `json:"license_decision_sha256"`
	LicenseDecisionSourceRef string `json:"license_decision_source_ref"`
}
type ReleaseEnvelopeInput struct {
	SchemaID string                    `json:"schema_id"`
	Envelope publication.EnvelopeInput `json:"envelope"`
	Approval string                    `json:"approval"`
}
type FixtureEnvelopeInput struct {
	SchemaID string                    `json:"schema_id"`
	Envelope publication.EnvelopeInput `json:"envelope"`
}

func PublishRelease(ctx context.Context, input []byte, clock publication.Clock) (publication.EnvelopeReport, error) {
	var in ReleaseEnvelopeInput
	if err := DecodeClosed(input, &in); err != nil {
		return publication.EnvelopeReport{}, err
	}
	if in.SchemaID != "cartulary.windows_event_release_envelope_input.v3" {
		return publication.EnvelopeReport{}, errors.New("explicit current release input required")
	}
	raw, err := publication.ReadRegular(in.Approval, 65536)
	if err != nil {
		return publication.EnvelopeReport{}, err
	}
	var approval ReleaseApproval
	if err = DecodeClosed(raw, &approval); err != nil {
		return publication.EnvelopeReport{}, err
	}
	if approval.SchemaID != "cartulary.reference_pack_release_approval.v1" || approval.Status != "approved" || strings.TrimSpace(approval.DecisionRef) == "" || strings.TrimSpace(approval.Authority) == "" {
		return publication.EnvelopeReport{}, errors.New("licensing authority approval of exact output required")
	}
	raw, err = publication.ReadRegular(filepath.Join(in.Envelope.LogicalRoot, "manifest.json"), 1048576)
	if err != nil {
		return publication.EnvelopeReport{}, err
	}
	m, err := envelope.DecodeManifest(raw, true)
	if err != nil {
		return publication.EnvelopeReport{}, err
	}
	if Digest(raw) != approval.ManifestSHA256 || m.SourceProfileSHA256 != approval.SourceProfileSHA256 || m.SourceProfileID != WindowsProfile {
		return publication.EnvelopeReport{}, errors.New("approval does not cover current source profile and exact output")
	}
	scope, err := publication.ReleaseScope(ReleaseRepository, PackKey)
	if err != nil {
		return publication.EnvelopeReport{}, err
	}
	return publication.PublishRelease(ctx, publication.ReleaseInput{Envelope: in.Envelope, Approval: publication.ApprovedOutput{ManifestSHA256: approval.ManifestSHA256, DecisionSHA256: approval.LicenseDecisionSHA256, DecisionSourceRef: approval.LicenseDecisionSourceRef}}, scope, clock)
}
func PublishFixture(ctx context.Context, input []byte, clock publication.Clock) (publication.EnvelopeReport, error) {
	var in FixtureEnvelopeInput
	if err := DecodeClosed(input, &in); err != nil {
		return publication.EnvelopeReport{}, err
	}
	if in.SchemaID != "cartulary.windows_event_fixture_envelope_input.v3" {
		return publication.EnvelopeReport{}, errors.New("explicit current fixture input required")
	}
	scope, err := publication.FixtureScope(DatasetFixtureRepository, PackKey)
	if err != nil {
		return publication.EnvelopeReport{}, err
	}
	return publication.PublishFixture(ctx, publication.FixtureInput{Envelope: in.Envelope}, scope, clock)
}
