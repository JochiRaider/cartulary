package reporting

import (
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"time"
)

// renderInput is the immutable input shared by release and preview rendering.
// Public request/idempotency and durable lifecycle identities belong to their
// respective admission codecs, not to the rendering interface.
type renderInput struct {
	preparation                  *renderPreparation
	SourceFacts                  exportprovider.SourceFacts `json:"source_facts"`
	SourceFactsSHA256            string                     `json:"source_facts_sha256"`
	ActorUserID                  string                     `json:"actor_user_id"`
	SnapshotID                   string                     `json:"snapshot_id"`
	IncidentID                   string                     `json:"incident_id"`
	SnapshotAt                   time.Time                  `json:"snapshot_at"`
	SourceChangeSetHighWatermark string                     `json:"source_change_set_high_watermark"`
	DerivationVersion            string                     `json:"derivation_version"`
	ExportModelSHA256            string                     `json:"export_model_sha256"`
	ExportModel                  ExportModel                `json:"export_model"`
	TemplateID                   string                     `json:"template_id"`
	TemplateVersion              string                     `json:"template_version"`
	TemplateContract             TemplateContract           `json:"template_contract"`
	TemplateFiles                []TemplatePackFile         `json:"template_files"`
	RedactionProfileID           string                     `json:"redaction_profile_id"`
	RedactionProfileVersion      string                     `json:"redaction_profile_version"`
	RedactionProfile             RedactionProfile           `json:"redaction_profile"`
	RedactionProfileSHA256       string                     `json:"redaction_profile_sha256"`
	OutputKind                   string                     `json:"output_kind"`
	OutputOptions                json.RawMessage            `json:"output_options"`
	ReleaseScope                 string                     `json:"release_scope"`
	RecipientPartitionRefs       []string                   `json:"recipient_partition_refs"`
	GraphProjectionRefs          json.RawMessage            `json:"graph_projection_refs"`
	CompositionJSON              json.RawMessage            `json:"composition_json,omitempty"`
	CompositionID                *string                    `json:"composition_id"`
	CompositionVersion           *string                    `json:"composition_version"`
	CompositionSHA256            *string                    `json:"composition_sha256"`
	RenderAdmittedAt             time.Time                  `json:"render_admitted_at"`
	GraphSources                 frozenGraphSources         `json:"graph_sources"`
}
