package reportcomposition

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	ProfileID = "snapshot_reporting"

	CompositionSchemaID       = "cartulary.report_composition.v1"
	PreviewSourceSchemaID     = "cartulary.report_composition_preview_source.v1"
	DefaultSourceKind         = "draft"
	SourceKindDraft           = "draft"
	SourceKindVersion         = "version"
	SourceKindInline          = "inline"
	ReleaseScopeInternalDraft = "internal_draft"
	OutputKindSlidev          = "slidev"
	OutputKindMermaid         = "mermaid"
)

var emptyJSONArray = json.RawMessage(`[]`)
var emptyJSONObject = json.RawMessage(`{}`)

type ResourceRecord struct {
	CompositionID             uuid.UUID
	IncidentID                uuid.UUID
	CreatedByUserID           uuid.UUID
	ClientTxnID               string
	TemplateID                string
	TemplateVersion           string
	DraftVersion              int64
	AuthoredAgainstSnapshotID *string
	DeckOps                   json.RawMessage
	DiagramDecls              json.RawMessage
	AuthoredTexts             json.RawMessage
	LatestCompositionVersion  *int64
	RetiredAt                 *time.Time
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type VersionRecord struct {
	CompositionID        uuid.UUID
	CompositionVersion   int64
	CompositionSHA256    string
	CanonicalComposition json.RawMessage
	CreatedByUserID      uuid.UUID
	CreatedAt            time.Time
	ReleaseBound         bool
}

type CreateDraftRequest struct {
	ClientTxnID               string
	TemplateID                string
	TemplateVersion           string
	AuthoredAgainstSnapshotID *string
	DeckOps                   json.RawMessage
	DiagramDecls              json.RawMessage
	AuthoredTexts             json.RawMessage
	Normalized                []byte
}

type UpdateDraftRequest struct {
	ClientTxnID               string
	BaseDraftVersion          int64
	AuthoredAgainstSnapshotID *string
	AuthoredAgainstPresent    bool
	DeckOps                   *json.RawMessage
	DiagramDecls              *json.RawMessage
	AuthoredTexts             *json.RawMessage
	Normalized                []byte
}

type DraftVersionRequest struct {
	ClientTxnID      string
	BaseDraftVersion int64
	Normalized       []byte
}

type ValidateRequest struct {
	SourceKind          string
	CompositionVersion  *int64
	InlineComposition   *json.RawMessage
	ValidationContext   *json.RawMessage
	ValidationContextIs bool
}

type PreviewRequest struct {
	ClientTxnID              string
	SourceKind               string
	CompositionVersion       *int64
	SnapshotID               string
	DerivationVersion        string
	TemplateID               string
	TemplateVersion          string
	RedactionProfileID       string
	RedactionProfileVersion  string
	RedactionProfileSHA256   string
	RenderEnvironmentProfile string
	OutputKind               string
	OutputOptions            json.RawMessage
	RecipientPartitionRefs   json.RawMessage
	GraphProjectionRefs      json.RawMessage
	Normalized               []byte
}

type MutationResult struct {
	Payload    map[string]any
	StatusCode int
	IncidentID uuid.UUID
	Replayed   bool
}

type PreviewResult struct {
	Payload    map[string]any
	StatusCode int
	IncidentID uuid.UUID
	Replayed   bool
}

// PreviewRenderSource is the immutable, typed projection Report Composition
// exposes to the Reporting worker. It contains only the persisted source and
// render-selection facts needed for an internal-draft render.
type PreviewRenderSource struct {
	PreviewAttemptID         uuid.UUID
	RenderAttemptID          uuid.UUID
	IncidentID               uuid.UUID
	CompositionID            uuid.UUID
	SourceKind               string
	DraftVersion             *int64
	CompositionVersion       *int64
	PreviewSourceSHA256      string
	CompositionSHA256        *string
	PreviewSourceJSON        json.RawMessage
	SnapshotID               string
	DerivationVersion        string
	TemplateID               string
	TemplateVersion          string
	RedactionProfileID       string
	RedactionProfileVersion  string
	RedactionProfileSHA256   string
	RenderEnvironmentProfile string
	OutputKind               string
	OutputOptions            json.RawMessage
	RecipientPartitionRefs   json.RawMessage
	GraphProjectionRefs      json.RawMessage
	CreatedByUserID          uuid.UUID
	CreatedAt                time.Time
}
