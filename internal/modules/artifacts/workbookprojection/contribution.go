package workbookprojection

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/surfacecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type Rows interface {
	RefreshArtifactTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadArtifactTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
	RebuildArtifactsTx(context.Context, pgx.Tx, uuid.UUID) error
}

type SourceReader interface {
	LoadProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (ProjectionInput, bool, error)
	ListProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (ProjectionInputPage, error)
}

// ProjectionInput is the typed source-owner materialization contract shared
// with private Projections storage. Nested optional details keep subtype-only
// facts explicit without exposing physical projection columns or SQL.
type ProjectionInput struct {
	RecordID                     uuid.UUID  `json:"record_id"`
	IncidentID                   uuid.UUID  `json:"incident_id"`
	RowVersion                   int64      `json:"row_version"`
	ArtifactType                 string     `json:"artifact_type"`
	Title                        *string    `json:"title"`
	Body                         *string    `json:"body"`
	TimestampUTC                 *time.Time `json:"timestamp_utc"`
	UpdatedAt                    time.Time  `json:"updated_at"`
	CreatedAt                    time.Time  `json:"created_at"`
	CreatedByUserID              *uuid.UUID `json:"created_by_user_id"`
	CommID                       *string    `json:"comm_id"`
	CommType                     *string    `json:"comm_type"`
	Audience                     *string    `json:"audience"`
	ChannelOrMeeting             *string    `json:"channel_or_meeting"`
	Summary                      *string    `json:"summary"`
	NextReportAt                 *time.Time `json:"next_report_at"`
	PrivilegeTag                 *string    `json:"privilege_tag"`
	HandoffID                    *string    `json:"handoff_id"`
	OutgoingOwnerUserID          *uuid.UUID `json:"outgoing_owner_user_id"`
	IncomingOwnerUserID          *uuid.UUID `json:"incoming_owner_user_id"`
	CurrentStateSummary          *string    `json:"current_state_summary"`
	NextChecks                   *string    `json:"next_checks"`
	AcknowledgedAt               *time.Time `json:"acknowledged_at"`
	StatusReviewID               *string    `json:"status_review_id"`
	ReviewOwnerUserID            *uuid.UUID `json:"review_owner_user_id"`
	ActiveRisksSummary           *string    `json:"active_risks_summary"`
	LessonID                     *string    `json:"lesson_id"`
	OwnerUserID                  *uuid.UUID `json:"owner_user_id"`
	ClosureState                 *string    `json:"closure_state"`
	FindingStatement             *string    `json:"finding_statement"`
	FindingKind                  *string    `json:"finding_kind"`
	FindingState                 *string    `json:"finding_state"`
	FindingOwnerUserID           *uuid.UUID `json:"finding_owner_user_id"`
	FindingConfidenceScore       *int       `json:"finding_confidence_score"`
	FindingClosedAt              *time.Time `json:"finding_closed_at"`
	FindingUpdatedAt             *time.Time `json:"finding_updated_at"`
	FindingConfidenceBand        *string    `json:"finding_confidence_band"`
	InvestigativeQueryQueryID    *string    `json:"investigative_query_query_id"`
	InvestigativeQueryPlatform   *string    `json:"investigative_query_platform"`
	InvestigativeQueryPurpose    *string    `json:"investigative_query_purpose"`
	InvestigativeQueryQueryText  *string    `json:"investigative_query_query_text"`
	InvestigativeQueryCreatedBy  *uuid.UUID `json:"investigative_query_created_by_user_id"`
	InvestigativeQueryCreatedAt  *time.Time `json:"investigative_query_created_at"`
	InvestigativeQueryCreatedDay *string    `json:"investigative_query_created_day"`
	ForensicKeywordKeywordID     *string    `json:"forensic_keyword_keyword_id"`
	ForensicKeywordPattern       *string    `json:"forensic_keyword_pattern"`
	ForensicKeywordReason        *string    `json:"forensic_keyword_reason"`
	ForensicKeywordMatchMode     *string    `json:"forensic_keyword_match_mode"`
	ForensicKeywordCaseSensitive *bool      `json:"forensic_keyword_case_sensitive"`
	ForensicKeywordCreatedAt     *time.Time `json:"forensic_keyword_created_at"`
	ForensicKeywordCreatedDay    *string    `json:"forensic_keyword_created_day"`
	TimestampDay                 *string    `json:"timestamp_day"`
	NextReportDay                *string    `json:"next_report_day"`
	AckState                     string     `json:"ack_state"`
	LinkedRecordCount            int        `json:"linked_record_count"`
}

type ProjectionInputPage struct {
	Inputs       []ProjectionInput
	NextRecordID *uuid.UUID
}

type DerivedFact struct {
	RecordID     uuid.UUID
	ArtifactType string
	FindingKind  *string
	Value        map[string]any
}

type Reader interface {
	CollectDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]DerivedFact, error)
}

type Rebuilder interface {
	RebuildArtifacts(context.Context, uuid.UUID) error
}

type Ports struct {
	Rows      Rows
	Rebuilder Rebuilder
	Reader    Reader
}

func (ports Ports) Ready() bool {
	return ports.Rows != nil && ports.Rebuilder != nil && ports.Reader != nil
}

type Contribution struct {
	contract providercontract.Contribution
	source   SourceReader
}

func NewContribution(
	descriptors []providercontract.ProviderDescriptor,
	intents []providercontract.SurfaceIntent,
	sources ...SourceReader,
) (Contribution, error) {
	if len(sources) > 1 {
		return Contribution{}, fmt.Errorf("artifacts projection contribution accepts at most one source")
	}
	contract, err := providercontract.NewContribution("artifacts", descriptors, intents)
	if err != nil {
		return Contribution{}, err
	}
	var source SourceReader
	if len(sources) == 1 {
		source = sources[0]
	}
	return Contribution{contract: contract, source: source}, nil
}

func NewRuntimeContribution(source SourceReader) (Contribution, error) {
	if source == nil {
		return Contribution{}, fmt.Errorf("artifacts projection source is required")
	}
	return NewContribution(
		[]providercontract.ProviderDescriptor{Descriptor()},
		SurfaceIntents(),
		source,
	)
}

func (contribution Contribution) ProjectionContribution() providercontract.Contribution {
	return contribution.contract
}

func (contribution Contribution) Source() SourceReader {
	return contribution.source
}

func Descriptor() providercontract.ProviderDescriptor {
	return providercontract.ProviderDescriptor{
		SchemaVersion:     providercontract.DescriptorSchemaVersion,
		Status:            providercontract.ProviderStatusActive,
		ProviderID:        "artifact",
		SourceOwnerModule: "artifacts",
		ViewSchemaIDs: []string{
			surfacecatalog.NotesViewSchemaID,
			surfacecatalog.CommLogViewSchemaID,
			surfacecatalog.HandoffViewSchemaID,
			surfacecatalog.StatusReviewViewSchemaID,
			surfacecatalog.LessonViewSchemaID,
			surfacecatalog.FindingsViewSchemaID,
			surfacecatalog.InvestigativeQueriesViewSchemaID,
			surfacecatalog.ForensicKeywordsViewSchemaID,
		},
		SourceRecordTypes:            []string{"artifact"},
		SourceAuthorityModules:       []string{"artifacts", "links", "parties", "records"},
		ProjectionTableIDs:           []string{"artifact_grid_projection"},
		ProjectionStorageOwnerModule: "projections",
		Capabilities: providercontract.ProviderCapabilities{
			Query:           true,
			RefreshRow:      true,
			RestoreRebuild:  true,
			IncidentRebuild: true,
		},
		RestoreRebuild: providercontract.RestoreRebuildRequired,
		FacadePackages: []string{"internal/modules/artifacts/workbookprojection"},
		RebuildAfter:   []string{"assessment"},
		CharacterizationRefs: []string{
			"internal/modules/workbook/coordination_surfaces_test.go",
			"internal/modules/projections/internal/runtime/query_test.go",
		},
	}
}

func SurfaceIntents() []providercontract.SurfaceIntent {
	surfaces := surfacecatalog.All()
	intents := make([]providercontract.SurfaceIntent, 0, len(surfaces))
	for _, surface := range surfaces {
		schema, ok := viewschema.Lookup(surface.ViewSchemaID)
		if !ok {
			panic("artifact projection surface has no view-schema contract: " + surface.ViewSchemaID)
		}
		fields := schema.Fields()
		fieldKeys := make([]string, 0, len(fields))
		for fieldKey := range fields {
			fieldKeys = append(fieldKeys, fieldKey)
		}
		slices.Sort(fieldKeys)
		intents = append(intents, providercontract.SurfaceIntent{
			ViewSchemaID: surface.ViewSchemaID,
			FieldKeys:    fieldKeys,
			CanonicalSourceFilter: &providercontract.SourceFilterIntent{
				Kind:  "artifact_type",
				Field: "artifact_type",
				Value: surface.ArtifactType,
			},
		})
	}
	return intents
}
