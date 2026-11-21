package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Staged write-ready streams (Core 01 REQ-01-641). Preparation proves every
// bundle-only fact and stages these rows; the final transaction only copies
// them. Paths stay below the participant's preparation prefix.
const (
	preparedRowsPrefix        = "prepared/network_flow_activity/rows/"
	preparedTablesPath        = preparedRowsPrefix + "tables.ndjson"
	preparedObservationsPath  = preparedRowsPrefix + "observations.ndjson"
	preparedDiagnosticsPath   = preparedRowsPrefix + "diagnostics.ndjson"
	preparedBindingsPath      = preparedRowsPrefix + "indicator_bindings.ndjson"
	preparedAttributionPath   = "prepared/network_flow_activity/attribution.ndjson"
	portableCreatedByColumn   = "created_by_user_id"
	attributionTableTables    = "network_flow_tables"
	attributionTableRows      = "network_flow_rows"
	attributionTableBindings  = "network_flow_indicator_bindings"
	attributionTableGraphView = "network_flow_graph_views"
)

// portableImportPlan is the only state the final transaction receives. It
// holds sealed preparation streams and bounded declarations, never a bundle
// member reader.
type portableImportPlan struct {
	incidentID   uuid.UUID
	operationID  string
	actorID      uuid.UUID
	actors       sourceport.ActorCatalog
	counts       map[string]int64
	graphs       []preparedPortableGraph
	preparation  incidentportability.MemberReader
	tables       incidentportability.PreparedRowStream
	observations incidentportability.PreparedRowStream
	diagnostics  incidentportability.PreparedRowStream
	bindings     incidentportability.PreparedRowStream
	attribution  incidentportability.AttributionStream
}

// preparedBinding is one write-ready indicator binding plus the historical
// actor its attribution preserves.
type preparedBinding struct {
	BindingID       string
	IndicatorID     uuid.UUID
	IndicatorType   string
	CandidateValue  string
	SelectorKind    string
	SourceRowRefs   json.RawMessage
	SourceRowIDs    []string
	Truncated       bool
	TotalCount      int64
	CreatedAt       time.Time
	HistoricalActor uuid.UUID
}

func encodePositional(values ...any) ([]byte, error) { return json.Marshal(values) }

func decodePositional(line []byte, targets ...any) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil || len(raw) != len(targets) {
		return errPortableNetworkFlow
	}
	for index, target := range targets {
		if err := json.Unmarshal(raw[index], target); err != nil {
			return errPortableNetworkFlow
		}
	}
	return nil
}

var portableTableColumns = []string{
	"network_flow_table_id", "incident_id", "display_name", "table_version", "table_status",
	"committed_source_id", "source_interpretation_sha256", "source_content_sha256",
	"source_filename_display", "source_filename_digest", "source_filename_digest_key_id",
	"mapping_fingerprint", "source_profile_id", "parser_profile_id", "row_count_accepted",
	"row_count_rejected", "diagnostics_truncated", "created_by_user_id", "created_at", "updated_at",
	"deleted_at", "source_raw_cell_scalar_limit",
}

// tableCodec stages a table with its final status, version and deployment-keyed
// filename digest, so COPY writes the final row (no insert-then-update).
type tableCodec struct{}

func (tableCodec) Encode(t tableRecord) ([]byte, error) {
	return encodePositional(t.TableID, t.IncidentID, t.DisplayName, t.TableVersion, t.TableStatus, t.CommittedSourceID,
		t.SourceInterpretationSHA256, t.SourceContentSHA256, t.SourceFilenameDisplay, t.SourceFilenameDigest,
		t.SourceFilenameDigestKeyID, t.MappingFingerprint, t.SourceProfileID, t.ParserProfileID, t.RowCountAccepted,
		t.RowCountRejected, t.DiagnosticsTruncated, t.CreatedByUserID, t.CreatedAt, t.UpdatedAt, t.DeletedAt, t.SourceRawCellScalarLimit)
}
func (tableCodec) Decode(line []byte) (tableRecord, error) {
	var t tableRecord
	err := decodePositional(line, &t.TableID, &t.IncidentID, &t.DisplayName, &t.TableVersion, &t.TableStatus, &t.CommittedSourceID,
		&t.SourceInterpretationSHA256, &t.SourceContentSHA256, &t.SourceFilenameDisplay, &t.SourceFilenameDigest,
		&t.SourceFilenameDigestKeyID, &t.MappingFingerprint, &t.SourceProfileID, &t.ParserProfileID, &t.RowCountAccepted,
		&t.RowCountRejected, &t.DiagnosticsTruncated, &t.CreatedByUserID, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &t.SourceRawCellScalarLimit)
	return t, err
}
func (tableCodec) Columns() []string { return portableTableColumns }
func (tableCodec) Values(t tableRecord) ([]any, error) {
	return []any{t.TableID, t.IncidentID, t.DisplayName, t.TableVersion, t.TableStatus, t.CommittedSourceID,
		t.SourceInterpretationSHA256, t.SourceContentSHA256, t.SourceFilenameDisplay, t.SourceFilenameDigest,
		t.SourceFilenameDigestKeyID, t.MappingFingerprint, t.SourceProfileID, t.ParserProfileID, t.RowCountAccepted,
		t.RowCountRejected, t.DiagnosticsTruncated, t.CreatedByUserID, t.CreatedAt, t.UpdatedAt, t.DeletedAt, t.SourceRawCellScalarLimit}, nil
}

// observationCodec stages an accepted observation exactly as the live import
// retains it (retainedFlowRowColumns), with the importing actor as creator.
type observationCodec struct{}

func (observationCodec) Encode(r flowRow) ([]byte, error) {
	return encodePositional(r.RowID, r.NetworkFlowTableID, r.IncidentID, r.SourceRowNumber, r.SourceRowDigestSHA256,
		r.NormalizedRowDigestSHA256, r.MappingFingerprint, r.FlowStartUTC, r.FlowEndUTC, r.EndpointAIP, r.EndpointBIP,
		r.EndpointAPort, r.EndpointBPort, r.IPProtocol, r.BytesCount, r.PacketsCount, r.ExporterID, r.InputInterface,
		r.OutputInterface, r.TCPFlags, r.ApplicationLabel, r.UnmappedRaw, r.ObservationSourceRef, r.CreatedAt,
		r.CreatedByUserID, r.Direction.Kind, r.Direction.FromEndpoint, r.Direction.Reason, r.DirectionEvidence)
}
func (observationCodec) Decode(line []byte) (flowRow, error) {
	var r flowRow
	err := decodePositional(line, &r.RowID, &r.NetworkFlowTableID, &r.IncidentID, &r.SourceRowNumber, &r.SourceRowDigestSHA256,
		&r.NormalizedRowDigestSHA256, &r.MappingFingerprint, &r.FlowStartUTC, &r.FlowEndUTC, &r.EndpointAIP, &r.EndpointBIP,
		&r.EndpointAPort, &r.EndpointBPort, &r.IPProtocol, &r.BytesCount, &r.PacketsCount, &r.ExporterID, &r.InputInterface,
		&r.OutputInterface, &r.TCPFlags, &r.ApplicationLabel, &r.UnmappedRaw, &r.ObservationSourceRef, &r.CreatedAt,
		&r.CreatedByUserID, &r.Direction.Kind, &r.Direction.FromEndpoint, &r.Direction.Reason, &r.DirectionEvidence)
	return r, err
}
func (observationCodec) Columns() []string               { return retainedFlowRowColumns }
func (observationCodec) Values(r flowRow) ([]any, error) { return retainedFlowRowValues(r), nil }

var portableDiagnosticColumns = []string{
	"diagnostic_id", "network_flow_table_id", "incident_id", "source_row_number",
	"source_column_ordinal", "raw_header_sha256", "field_key", "error_code", "reason_code",
	"safe_sample", "raw_value_sha256", "message_key", "message_args", "message",
	"limit_name", "limit_value", "actual_value", "created_at",
}

type diagnosticCodec struct{ incidentID uuid.UUID }

func (diagnosticCodec) Encode(d portableDiagnostic) ([]byte, error) {
	return encodePositional(d.TableID, d.CreatedAt, d.Diagnostic)
}
func (c diagnosticCodec) Decode(line []byte) (portableDiagnostic, error) {
	d := portableDiagnostic{IncidentID: c.incidentID}
	err := decodePositional(line, &d.TableID, &d.CreatedAt, &d.Diagnostic)
	return d, err
}
func (diagnosticCodec) Columns() []string { return portableDiagnosticColumns }
func (c diagnosticCodec) Values(v portableDiagnostic) ([]any, error) {
	d := v.Diagnostic
	messageArgs := d.MessageArgs
	if len(messageArgs) == 0 {
		messageArgs = json.RawMessage(`{}`)
	}
	return []any{d.DiagnosticID, v.TableID, c.incidentID, d.SourceRowNumber, d.SourceColumnOrdinal, d.RawHeaderSHA256,
		d.FieldKey, d.ErrorCode, d.ReasonCode, d.SafeSample, d.RawValueSHA256, d.MessageKey, string(messageArgs),
		d.Message, d.LimitName, d.LimitValue, d.ActualValue, v.CreatedAt}, nil
}

var portableBindingColumns = []string{
	"network_flow_indicator_binding_id", "incident_id", "target_indicator_record_id",
	"target_indicator_type", "target_indicator_value_kind", "target_indicator_normalized_value",
	"selector_kind", "candidate_value", "source_row_refs", "source_row_ref_row_ids",
	"source_row_refs_truncated", "source_row_refs_total_count", "created_by_user_id", "created_at",
}

type bindingCodec struct{ incidentID, actorID uuid.UUID }

func (bindingCodec) Encode(b preparedBinding) ([]byte, error) {
	return encodePositional(b.BindingID, b.IndicatorID, b.IndicatorType, b.CandidateValue, b.SelectorKind, b.SourceRowRefs,
		b.SourceRowIDs, b.Truncated, b.TotalCount, b.CreatedAt, b.HistoricalActor)
}
func (bindingCodec) Decode(line []byte) (preparedBinding, error) {
	var b preparedBinding
	err := decodePositional(line, &b.BindingID, &b.IndicatorID, &b.IndicatorType, &b.CandidateValue, &b.SelectorKind, &b.SourceRowRefs,
		&b.SourceRowIDs, &b.Truncated, &b.TotalCount, &b.CreatedAt, &b.HistoricalActor)
	return b, err
}
func (bindingCodec) Columns() []string { return portableBindingColumns }
func (c bindingCodec) Values(b preparedBinding) ([]any, error) {
	return []any{b.BindingID, c.incidentID, b.IndicatorID, b.IndicatorType, "atomic", b.CandidateValue, b.SelectorKind,
		b.CandidateValue, string(b.SourceRowRefs), b.SourceRowIDs, b.Truncated, b.TotalCount, c.actorID, b.CreatedAt}, nil
}

// stageTables stages the admitted tables (bounded by the retained-table limit)
// in identity order with the importing actor and a destination-keyed digest.
func (p *preparedNetworkFlow) stageTables(ctx context.Context, preparation incidentportability.MemberWriter, digester safeDigester) (incidentportability.PreparedRowStream, error) {
	if digester == nil {
		return incidentportability.PreparedRowStream{}, incidentbundles.ErrPortabilityUnavailable
	}
	ids := make([]string, 0, len(p.tables))
	for id := range p.tables {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return incidentportability.StagePreparedRows(ctx, preparation, preparedTablesPath, tableCodec{}, func(emit func(tableRecord) error) error {
		for _, id := range ids {
			table := p.tables[id]
			digest, key, err := digester.Digest("source_filename", table.SourceFilenameDisplay)
			if err != nil {
				return err
			}
			table.SourceFilenameDigest, table.SourceFilenameDigestKeyID, table.CreatedByUserID = digest, key, p.actorID
			if err := emit(table); err != nil {
				return err
			}
		}
		return nil
	})
}

// stageAttribution derives every historical-actor attribution from admitted
// state and the sealed preparation streams; it reads no bundle member.
func (p *preparedNetworkFlow) stageAttribution(ctx context.Context, preparation incidentbundles.PreparationMembers) (incidentportability.AttributionStream, error) {
	ids := make([]string, 0, len(p.tables))
	for id := range p.tables {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	admitted := func(actor string) bool { _, ok := p.actors.Lookup(actor); return ok }
	stream, err := incidentportability.StageAttributions(ctx, preparation, preparedAttributionPath, admitted, func(emit func(incidentportability.AttributionRow) error) error {
		row := func(table, id string, actor uuid.UUID) error {
			return emit(incidentportability.AttributionRow{SourceTable: table, SourceRowID: id, SourceColumn: portableCreatedByColumn, SourceActorID: actor.String()})
		}
		for _, id := range ids {
			if err := row(attributionTableTables, id, p.tables[id].CreatedByUserID); err != nil {
				return err
			}
		}
		// Source verification proved each observation's creator equals its table's.
		err := incidentportability.ReadPreparedRows(ctx, preparation, p.observations, observationCodec{}, func(observation flowRow) error {
			table, ok := p.tables[observation.NetworkFlowTableID]
			if !ok {
				return errPortableNetworkFlow
			}
			return row(attributionTableRows, observation.RowID, table.CreatedByUserID)
		})
		if err != nil {
			return err
		}
		err = incidentportability.ReadPreparedRows(ctx, preparation, p.bindings, bindingCodec{}, func(binding preparedBinding) error {
			return row(attributionTableBindings, binding.BindingID, binding.HistoricalActor)
		})
		if err != nil {
			return err
		}
		for _, graph := range p.graphs {
			if err := row(attributionTableGraphView, graph.declaration.GraphViewID, graph.declaration.CreatedByUserID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return incidentportability.AttributionStream{}, err
	}
	want := p.counts["tables"] + p.counts["observations"] + p.counts["indicator_bindings"] + p.counts["graph_views"]
	if stream.Rows != want {
		return incidentportability.AttributionStream{}, fmt.Errorf("prepared attribution count: %w", errPortableNetworkFlow)
	}
	return stream, nil
}

// copyPreparedFamilies applies the staged families in foreign-key order with
// one COPY each into Network Flow's own relations. The database acknowledgement
// is the aggregate check against the staged expectations.
func copyPreparedFamilies(ctx context.Context, tx pgx.Tx, p *portableImportPlan) error {
	tables, err := incidentportability.OpenPreparedCopy(ctx, p.preparation, p.tables, tableCodec{})
	if err != nil {
		return err
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"network_flow_tables"}, portableTableColumns, tables)
	if err := tables.Finish(written, err); err != nil {
		return fmt.Errorf("copy prepared network flow tables: %w", err)
	}
	observations, err := incidentportability.OpenPreparedCopy(ctx, p.preparation, p.observations, observationCodec{})
	if err != nil {
		return err
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"network_flow_rows"}, retainedFlowRowColumns, observations)
	if err := observations.Finish(written, err); err != nil {
		return fmt.Errorf("copy prepared network flow rows: %w", err)
	}
	diagnostics, err := incidentportability.OpenPreparedCopy(ctx, p.preparation, p.diagnostics, diagnosticCodec{incidentID: p.incidentID})
	if err != nil {
		return err
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"network_flow_rejected_row_diagnostics"}, portableDiagnosticColumns, diagnostics)
	if err := diagnostics.Finish(written, err); err != nil {
		return fmt.Errorf("copy prepared network flow diagnostics: %w", err)
	}
	bindings, err := incidentportability.OpenPreparedCopy(ctx, p.preparation, p.bindings, bindingCodec{incidentID: p.incidentID, actorID: p.actorID})
	if err != nil {
		return err
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"network_flow_indicator_bindings"}, portableBindingColumns, bindings)
	if err := bindings.Finish(written, err); err != nil {
		return fmt.Errorf("copy prepared network flow indicator bindings: %w", err)
	}
	return nil
}

// plan stages the remaining write-ready streams and returns the transaction
// state. Preparation-time bundle access ends here.
func (p *preparedNetworkFlow) plan(ctx context.Context, digester safeDigester) (*portableImportPlan, error) {
	if p == nil || p.preparation == nil {
		return nil, incidentbundles.ErrPortabilityUnavailable
	}
	tables, err := p.stageTables(ctx, p.preparation, digester)
	if err != nil {
		return nil, err
	}
	if tables.Rows != p.counts["tables"] {
		return nil, fmt.Errorf("prepared table count: %w", errPortableNetworkFlow)
	}
	attribution, err := p.stageAttribution(ctx, p.preparation)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int64, len(p.counts))
	for family, count := range p.counts {
		counts[family] = count
	}
	return &portableImportPlan{
		incidentID: p.incidentID, operationID: p.operationID, actorID: p.actorID, actors: p.actors,
		counts: counts, graphs: append([]preparedPortableGraph(nil), p.graphs...), preparation: p.preparation,
		tables: tables, observations: p.observations, diagnostics: p.diagnostics, bindings: p.bindings,
		attribution: attribution,
	}, nil
}
