package networkflow

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// Portable records deliberately omit deployment-keyed correlation and live job state.
// These closed wire shapes are owned here, independently of storage row codecs.
type portableTable struct {
	TableID                    string     `json:"network_flow_table_id"`
	IncidentID                 uuid.UUID  `json:"incident_id"`
	DisplayName                string     `json:"display_name"`
	TableVersion               int64      `json:"table_version"`
	TableStatus                string     `json:"table_status"`
	CommittedSourceID          uuid.UUID  `json:"committed_source_id"`
	SourceInterpretationSHA256 string     `json:"source_interpretation_sha256"`
	SourceRawCellScalarLimit   int64      `json:"source_raw_cell_scalar_limit"`
	SourceContentSHA256        string     `json:"source_content_sha256"`
	SourceFilenameDisplay      string     `json:"source_filename_display"`
	MappingFingerprint         string     `json:"mapping_fingerprint"`
	SourceProfileID            string     `json:"source_profile_id"`
	ParserProfileID            string     `json:"parser_profile_id"`
	RowCountAccepted           int64      `json:"row_count_accepted"`
	RowCountRejected           int64      `json:"row_count_rejected"`
	DiagnosticsTruncated       bool       `json:"diagnostics_truncated"`
	CreatedByUserID            uuid.UUID  `json:"created_by_user_id"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
	DeletedAt                  *time.Time `json:"deleted_at"`
}

func portableTableRecord(value tableRecord) portableTable {
	return portableTable{
		TableID:                    value.TableID,
		IncidentID:                 value.IncidentID,
		DisplayName:                value.DisplayName,
		TableVersion:               value.TableVersion,
		TableStatus:                value.TableStatus,
		CommittedSourceID:          value.CommittedSourceID,
		SourceInterpretationSHA256: value.SourceInterpretationSHA256,
		SourceRawCellScalarLimit:   value.SourceRawCellScalarLimit,
		SourceContentSHA256:        value.SourceContentSHA256,
		SourceFilenameDisplay:      value.SourceFilenameDisplay,
		MappingFingerprint:         value.MappingFingerprint,
		SourceProfileID:            value.SourceProfileID,
		ParserProfileID:            value.ParserProfileID,
		RowCountAccepted:           value.RowCountAccepted,
		RowCountRejected:           value.RowCountRejected,
		DiagnosticsTruncated:       value.DiagnosticsTruncated,
		CreatedByUserID:            value.CreatedByUserID,
		CreatedAt:                  value.CreatedAt,
		UpdatedAt:                  value.UpdatedAt,
		DeletedAt:                  value.DeletedAt,
	}
}

func (value portableTable) retained() tableRecord {
	return tableRecord{
		TableID:                    value.TableID,
		IncidentID:                 value.IncidentID,
		DisplayName:                value.DisplayName,
		TableVersion:               value.TableVersion,
		TableStatus:                value.TableStatus,
		CommittedSourceID:          value.CommittedSourceID,
		SourceInterpretationSHA256: value.SourceInterpretationSHA256,
		SourceRawCellScalarLimit:   value.SourceRawCellScalarLimit,
		SourceContentSHA256:        value.SourceContentSHA256,
		SourceFilenameDisplay:      value.SourceFilenameDisplay,
		MappingFingerprint:         value.MappingFingerprint,
		SourceProfileID:            value.SourceProfileID,
		ParserProfileID:            value.ParserProfileID,
		RowCountAccepted:           value.RowCountAccepted,
		RowCountRejected:           value.RowCountRejected,
		DiagnosticsTruncated:       value.DiagnosticsTruncated,
		CreatedByUserID:            value.CreatedByUserID,
		CreatedAt:                  value.CreatedAt,
		UpdatedAt:                  value.UpdatedAt,
		DeletedAt:                  value.DeletedAt,
	}
}

type portableObservation struct {
	Direction                 flowDirection   `json:"direction"`
	DirectionEvidence         json.RawMessage `json:"direction_evidence"`
	RowID                     string          `json:"network_flow_row_id"`
	NetworkFlowTableID        string          `json:"network_flow_table_id"`
	IncidentID                uuid.UUID       `json:"incident_id"`
	SourceRowNumber           int64           `json:"source_row_number"`
	SourceRowDigestSHA256     string          `json:"source_row_digest_sha256"`
	NormalizedRowDigestSHA256 string          `json:"normalized_row_digest_sha256"`
	MappingFingerprint        string          `json:"mapping_fingerprint"`
	FlowStartUTC              time.Time       `json:"flow_start_utc"`
	FlowEndUTC                time.Time       `json:"flow_end_utc"`
	EndpointAIP               string          `json:"endpoint_a_ip"`
	EndpointBIP               string          `json:"endpoint_b_ip"`
	EndpointAPort             *int32          `json:"endpoint_a_port"`
	EndpointBPort             *int32          `json:"endpoint_b_port"`
	IPProtocol                int32           `json:"ip_protocol"`
	BytesCount                string          `json:"bytes_count"`
	PacketsCount              string          `json:"packets_count"`
	ExporterID                *string         `json:"exporter_id"`
	InputInterface            *string         `json:"input_interface"`
	OutputInterface           *string         `json:"output_interface"`
	TCPFlags                  *int32          `json:"tcp_flags"`
	ApplicationLabel          *string         `json:"application_label"`
	UnmappedRaw               json.RawMessage `json:"unmapped_raw"`
	ObservationSourceRef      json.RawMessage `json:"observation_source_ref"`
	CreatedAt                 time.Time       `json:"created_at"`
	CreatedByUserID           uuid.UUID       `json:"created_by_user_id"`
}

func portableFlowRow(value flowRow) portableObservation {
	return portableObservation(value)
}

func (value portableObservation) retained() flowRow {
	return flowRow(value)
}

type portableGraph struct {
	GraphViewID               string                          `json:"graph_view_id"`
	IncidentID                uuid.UUID                       `json:"incident_id"`
	DisplayName               string                          `json:"display_name"`
	NormalizedDisplayName     string                          `json:"normalized_display_name"`
	DeclarationState          string                          `json:"declaration_state"`
	RefreshIntentJSON         json.RawMessage                 `json:"refresh_intent"`
	SelectedSourceBindingJSON json.RawMessage                 `json:"selected_source_binding"`
	SemanticQueryJSON         json.RawMessage                 `json:"semantic_query"`
	SemanticQuerySHA256       string                          `json:"semantic_query_sha256"`
	DesiredSourceSnapshotID   string                          `json:"desired_source_snapshot_id"`
	SelectedResult            *graphViewSelectedResultBinding `json:"selected_result"`
	GraphViewVersion          int64                           `json:"graph_view_version"`
	MaterializationGeneration int64                           `json:"materialization_generation"`
	CreatedByUserID           uuid.UUID                       `json:"created_by_user_id"`
	CreatedAt                 time.Time                       `json:"created_at"`
	UpdatedAt                 time.Time                       `json:"updated_at"`
	RetiredAt                 *time.Time                      `json:"retired_at"`
}

func portableGraphViewDeclaration(value graphViewDeclaration) portableGraph {
	return portableGraph{
		GraphViewID:               value.GraphViewID,
		IncidentID:                value.IncidentID,
		DisplayName:               value.DisplayName,
		NormalizedDisplayName:     value.NormalizedDisplayName,
		DeclarationState:          value.DeclarationState,
		RefreshIntentJSON:         value.RefreshIntentJSON,
		SelectedSourceBindingJSON: value.SelectedSourceBindingJSON,
		SemanticQueryJSON:         value.SemanticQueryJSON,
		SemanticQuerySHA256:       value.SemanticQuerySHA256,
		DesiredSourceSnapshotID:   value.DesiredSourceSnapshotID,
		SelectedResult:            value.SelectedResult,
		GraphViewVersion:          value.GraphViewVersion,
		MaterializationGeneration: value.MaterializationGeneration,
		CreatedByUserID:           value.CreatedByUserID,
		CreatedAt:                 value.CreatedAt,
		UpdatedAt:                 value.UpdatedAt,
		RetiredAt:                 value.RetiredAt,
	}
}

func (value portableGraph) retained() graphViewDeclaration {
	return graphViewDeclaration{
		GraphViewID:               value.GraphViewID,
		IncidentID:                value.IncidentID,
		DisplayName:               value.DisplayName,
		NormalizedDisplayName:     value.NormalizedDisplayName,
		DeclarationState:          value.DeclarationState,
		RefreshIntentJSON:         value.RefreshIntentJSON,
		SelectedSourceBindingJSON: value.SelectedSourceBindingJSON,
		SemanticQueryJSON:         value.SemanticQueryJSON,
		SemanticQuerySHA256:       value.SemanticQuerySHA256,
		DesiredSourceSnapshotID:   value.DesiredSourceSnapshotID,
		SelectedResult:            value.SelectedResult,
		GraphViewVersion:          value.GraphViewVersion,
		MaterializationGeneration: value.MaterializationGeneration,
		CreatedByUserID:           value.CreatedByUserID,
		CreatedAt:                 value.CreatedAt,
		UpdatedAt:                 value.UpdatedAt,
		RetiredAt:                 value.RetiredAt,
	}
}

type portableDiagnostic struct {
	TableID    string                `json:"network_flow_table_id"`
	IncidentID uuid.UUID             `json:"incident_id"`
	CreatedAt  time.Time             `json:"created_at"`
	Diagnostic rejectedRowDiagnostic `json:"diagnostic"`
}

type portableBinding struct {
	BindingID               string              `json:"network_flow_indicator_binding_id"`
	IncidentID              uuid.UUID           `json:"incident_id"`
	IndicatorID             uuid.UUID           `json:"target_indicator_record_id"`
	IndicatorType           string              `json:"target_indicator_type"`
	CandidateValue          string              `json:"candidate_value"`
	SelectorKind            string              `json:"selector_kind"`
	SourceRowRefs           []networkFlowRowRef `json:"source_row_refs"`
	SourceRowRefsTruncated  bool                `json:"source_row_refs_truncated"`
	SourceRowRefsTotalCount int64               `json:"source_row_refs_total_count"`
	CreatedByUserID         uuid.UUID           `json:"created_by_user_id"`
	CreatedAt               time.Time           `json:"created_at"`
}

const portableTableMembers = "network_flow_table_id incident_id display_name table_version table_status committed_source_id source_interpretation_sha256 source_raw_cell_scalar_limit source_content_sha256 source_filename_display mapping_fingerprint source_profile_id parser_profile_id row_count_accepted row_count_rejected diagnostics_truncated created_by_user_id created_at updated_at deleted_at"

const portableObservationMembers = "direction direction_evidence network_flow_row_id network_flow_table_id incident_id source_row_number source_row_digest_sha256 normalized_row_digest_sha256 mapping_fingerprint flow_start_utc flow_end_utc endpoint_a_ip endpoint_b_ip endpoint_a_port endpoint_b_port ip_protocol bytes_count packets_count exporter_id input_interface output_interface tcp_flags application_label unmapped_raw observation_source_ref created_at created_by_user_id"

const portableDiagnosticMembers = "network_flow_table_id incident_id created_at diagnostic"

const portableBindingMembers = "network_flow_indicator_binding_id incident_id target_indicator_record_id target_indicator_type candidate_value selector_kind source_row_refs source_row_refs_truncated source_row_refs_total_count created_by_user_id created_at"

const portableGraphMembers = "graph_view_id incident_id display_name normalized_display_name declaration_state refresh_intent selected_source_binding semantic_query semantic_query_sha256 desired_source_snapshot_id selected_result graph_view_version materialization_generation created_by_user_id created_at updated_at retired_at"

const portableObservationNullable = "endpoint_a_port endpoint_b_port exporter_id input_interface output_interface tcp_flags application_label"
