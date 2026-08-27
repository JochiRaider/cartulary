package workbookprojection

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const evidenceViewSchemaID = "cartulary.view.evidence.v1"

type Rows interface {
	RefreshEvidenceTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadEvidenceTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
}

type ProjectionInput struct {
	RecordID           uuid.UUID
	IncidentID         uuid.UUID
	RowVersion         int64
	Title              *string
	LifecycleState     string
	RequestedAt        *time.Time
	ReceivedAt         *time.Time
	StorageRef         *string
	BlobHash           *string
	CollectorPartyText *string
	CollectorPartyID   *uuid.UUID
	SourcePartyText    *string
	SourcePartyID      *uuid.UUID
	UploadState        string
	LinkedRecordCount  int
	EditedAt           time.Time
}

// ViewRow is the canonical Evidence projection-input-to-workbook-row mapper.
// Source mutations, revision live values, and provider transaction reads use
// this boundary so new Evidence fields cannot drift across parallel serializers.
func ViewRow(input ProjectionInput) map[string]any {
	return map[string]any{
		"record_id":   input.RecordID.String(),
		"row_version": input.RowVersion,
		"cells": map[string]any{
			"evidence.title":                viewCell(stringValue(input.Title)),
			"evidence.lifecycle_state":      viewCell(input.LifecycleState),
			"evidence.requested_at":         viewCell(timeValue(input.RequestedAt)),
			"evidence.received_at":          viewCell(timeValue(input.ReceivedAt)),
			"evidence.storage_ref":          viewCell(stringValue(input.StorageRef)),
			"evidence.blob_hash":            viewCell(stringValue(input.BlobHash)),
			"evidence.collector_party_text": viewCell(stringValue(input.CollectorPartyText)),
			"evidence.collector_party_id":   viewCell(uuidValue(input.CollectorPartyID)),
			"evidence.source_party_text":    viewCell(stringValue(input.SourcePartyText)),
			"evidence.source_party_id":      viewCell(uuidValue(input.SourcePartyID)),
			"evidence.upload_state":         viewCell(input.UploadState),
			"evidence.linked_record_count":  viewCell(int32(input.LinkedRecordCount)),
			"evidence.edited_at":            viewCell(input.EditedAt.UTC().Format(time.RFC3339Nano)),
		},
	}
}

func viewCell(value any) map[string]any {
	return map[string]any{"value": value}
}

func stringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func timeValue(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func uuidValue(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

type ProjectionInputPage struct {
	Inputs       []ProjectionInput
	NextRecordID *uuid.UUID
}

type SourceReader interface {
	LoadProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (ProjectionInput, bool, error)
	ListProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (ProjectionInputPage, error)
}

type Ports struct {
	Rows           Rows
	SupportEffects SupportProjectionEffectsTx
}

type Contribution struct {
	contract providercontract.Contribution
	source   SourceReader
}

func NewContribution(source SourceReader) (Contribution, error) {
	if source == nil {
		return Contribution{}, fmt.Errorf("evidence projection source is required")
	}
	intent, err := SurfaceIntent()
	if err != nil {
		return Contribution{}, err
	}
	contract, err := providercontract.NewContribution(
		"evidence",
		[]providercontract.ProviderDescriptor{Descriptor()},
		[]providercontract.SurfaceIntent{intent},
	)
	if err != nil {
		return Contribution{}, err
	}
	return Contribution{contract: contract, source: source}, nil
}

func (contribution Contribution) ProjectionContribution() providercontract.Contribution {
	return contribution.contract
}

func (contribution Contribution) Source() SourceReader {
	return contribution.source
}

func Descriptor() providercontract.ProviderDescriptor {
	return providercontract.ProviderDescriptor{
		SchemaVersion:                providercontract.DescriptorSchemaVersion,
		Status:                       providercontract.ProviderStatusActive,
		ProviderID:                   "evidence",
		SourceOwnerModule:            "evidence",
		ViewSchemaIDs:                []string{evidenceViewSchemaID},
		SourceRecordTypes:            []string{"evidence"},
		SourceAuthorityModules:       []string{"evidence", "links", "records"},
		ProjectionTableIDs:           []string{"evidence_grid_projection"},
		ProjectionStorageOwnerModule: "projections",
		Capabilities: providercontract.ProviderCapabilities{
			Query:           true,
			RefreshRow:      true,
			RestoreRebuild:  true,
			IncidentRebuild: true,
		},
		RestoreRebuild: providercontract.RestoreRebuildRequired,
		FacadePackages: []string{"internal/modules/evidence/workbookprojection"},
		RebuildAfter:   []string{"artifact"},
		CharacterizationRefs: []string{
			"internal/modules/evidence/projection_collaboration_integration_test.go",
			"internal/modules/projections/internal/runtime/query_test.go",
		},
	}
}

func SurfaceIntent() (providercontract.SurfaceIntent, error) {
	return surfaceIntent(evidenceViewSchemaID)
}

func surfaceIntent(viewSchemaID string) (providercontract.SurfaceIntent, error) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		return providercontract.SurfaceIntent{}, fmt.Errorf("evidence projection surface %q has no view-schema contract", viewSchemaID)
	}
	fields := schema.Fields()
	fieldKeys := make([]string, 0, len(fields))
	for fieldKey := range fields {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	slices.Sort(fieldKeys)
	return providercontract.SurfaceIntent{
		ViewSchemaID: viewSchemaID,
		FieldKeys:    fieldKeys,
	}, nil
}
