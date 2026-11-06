package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const portabilityContentSchema = "cartulary.network_flow.incident_members.v1"
const portabilityMemberPrefix = "ext/extensions/network_flow_activity/members/"

var portabilityFamilyPaths = []string{"tables.ndjson", "observations.ndjson", "diagnostics.ndjson", "indicator_bindings.ndjson", "graph_views.ndjson"}

// Only application assembly can refine the caller's export snapshot. Participants
// receive a logical, incident-scoped stream operation, never its SQL capability.
type portabilityExportAccess struct {
	query        incidentbundles.ExportSnapshotQuery
	incidentID   uuid.UUID
	attributions sourceport.PortableAttributionResolver
}

func (r *portabilityExportAccess) ProfileScope() string     { return ProfileID }
func (r *portabilityExportAccess) IncidentScope() uuid.UUID { return r.incidentID }
func NewPortabilityExportAccess(query incidentbundles.ExportSnapshotQuery, incidentID uuid.UUID, attributions sourceport.PortableAttributionResolver) (incidentbundles.ExportReadCapability, error) {
	if query == nil || incidentID == uuid.Nil || attributions == nil {
		return nil, incidentbundles.ErrPortabilityUnavailable
	}
	return &portabilityExportAccess{query: query, incidentID: incidentID, attributions: attributions}, nil
}

// Each page releases its query before attribution lookup, so all reads can share
// one physical repeatable-read snapshot without accumulating the incident.
func exportPortabilityFamily[T any](ctx context.Context, access *portabilityExportAccess, members incidentportability.MemberWriter, actors incidentportability.ActorReferences, path, table string, page func(*T) ([]T, error), actor func(*T) (string, *uuid.UUID)) error {
	return incidentportability.WriteMemberFrom(ctx, members, portabilityMemberPrefix+path, func(ctx context.Context, out io.Writer) error {
		var after *T
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			values, err := page(after)
			if err != nil {
				return err
			}
			if len(values) == 0 {
				return nil
			}
			var historical map[string]string
			if actor != nil {
				ids := make([]string, 0, len(values))
				for i := range values {
					id, _ := actor(&values[i])
					ids = append(ids, id)
				}
				historical, err = access.attributions.ResolvePortableSourceActors(ctx, access.query, access.incidentID, table, "created_by_user_id", ids)
				if err != nil {
					return err
				}
			}
			// Capture the physical cursor before replacing attribution.
			last := values[len(values)-1]
			after = &last
			for i := range values {
				if actor != nil {
					id, target := actor(&values[i])
					if original, ok := historical[id]; ok {
						parsed, err := uuid.Parse(original)
						if err != nil {
							return err
						}
						*target = parsed
					}
					if err := actors.RecordActorReference(*target); err != nil {
						return err
					}
				}
				raw, err := json.Marshal(values[i])
				if err != nil {
					return err
				}
				canonical, err := incidentportability.CanonicalRawJSON(raw)
				if err != nil {
					return err
				}
				if len(canonical) > incidentportability.MaxNDJSONLogicalLineBytes {
					return incidentbundles.ErrPortabilityLimit
				}
				n, err := out.Write(canonical)
				if err != nil {
					return err
				}
				if n != len(canonical) {
					return io.ErrShortWrite
				}
			}
		}
	})
}

func portabilityPage[T any](rows pgx.Rows, scan func(pgx.Row) (T, error)) ([]T, error) {
	defer rows.Close()
	values := []T{}
	for rows.Next() {
		value, err := scan(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func (r *portabilityExportAccess) ExportRetained(ctx context.Context, members incidentportability.MemberWriter, actors incidentportability.ActorReferences) error {
	if r == nil || members == nil || actors == nil {
		return incidentbundles.ErrPortabilityUnavailable
	}
	err := exportPortabilityFamily(ctx, r, members, actors, "tables.ndjson", "network_flow_tables", func(after *portableTable) ([]portableTable, error) {
		cursor := ""
		if after != nil {
			cursor = after.TableID
		}
		rows, err := r.query.Query(ctx, tableSelectColumns()+` FROM network_flow_tables WHERE incident_id=$1 AND network_flow_table_id>$2 ORDER BY network_flow_table_id LIMIT 64`, r.incidentID, cursor)
		if err != nil {
			return nil, err
		}
		return portabilityPage(rows, func(row pgx.Row) (portableTable, error) {
			value, err := scanTable(row)
			return portableTableRecord(value), err
		})
	}, func(value *portableTable) (string, *uuid.UUID) { return value.TableID, &value.CreatedByUserID })
	if err != nil {
		return err
	}
	err = exportPortabilityFamily(ctx, r, members, actors, "observations.ndjson", "network_flow_rows", func(after *portableObservation) ([]portableObservation, error) {
		query := `SELECT ` + flowRowColumnList() + ` FROM network_flow_rows WHERE incident_id=$1`
		args := []any{r.incidentID}
		if after != nil {
			query += ` AND (network_flow_table_id,source_row_number)>($2,$3)`
			args = append(args, after.NetworkFlowTableID, after.SourceRowNumber)
		}
		query += ` ORDER BY network_flow_table_id,source_row_number LIMIT 64`
		rows, err := r.query.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		return portabilityPage(rows, func(row pgx.Row) (portableObservation, error) {
			value, err := scanFlowRow(row)
			return portableFlowRow(value), err
		})
	}, func(value *portableObservation) (string, *uuid.UUID) { return value.RowID, &value.CreatedByUserID })
	if err != nil {
		return err
	}
	err = exportPortabilityFamily(ctx, r, members, actors, "diagnostics.ndjson", "network_flow_rejected_row_diagnostics", func(after *portableDiagnostic) ([]portableDiagnostic, error) {
		query := `SELECT network_flow_table_id,incident_id,created_at,diagnostic_id,source_row_number,source_column_ordinal,raw_header_sha256,field_key,error_code,reason_code,safe_sample,raw_value_sha256,message_key,message_args,message,limit_name,limit_value,actual_value FROM network_flow_rejected_row_diagnostics WHERE incident_id=$1`
		args := []any{r.incidentID}
		if after != nil {
			query += ` AND (network_flow_table_id,diagnostic_id)>($2,$3)`
			args = append(args, after.TableID, after.Diagnostic.DiagnosticID)
		}
		query += ` ORDER BY network_flow_table_id,diagnostic_id LIMIT 64`
		rows, err := r.query.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		return portabilityPage(rows, func(row pgx.Row) (portableDiagnostic, error) {
			var v portableDiagnostic
			d := &v.Diagnostic
			err := row.Scan(&v.TableID, &v.IncidentID, &v.CreatedAt, &d.DiagnosticID, &d.SourceRowNumber, &d.SourceColumnOrdinal, &d.RawHeaderSHA256, &d.FieldKey, &d.ErrorCode, &d.ReasonCode, &d.SafeSample, &d.RawValueSHA256, &d.MessageKey, &d.MessageArgs, &d.Message, &d.LimitName, &d.LimitValue, &d.ActualValue)
			v.CreatedAt = v.CreatedAt.UTC()
			return v, err
		})
	}, nil)
	if err != nil {
		return err
	}
	err = exportPortabilityFamily(ctx, r, members, actors, "indicator_bindings.ndjson", "network_flow_indicator_bindings", func(after *portableBinding) ([]portableBinding, error) {
		cursor := ""
		if after != nil {
			cursor = after.BindingID
		}
		rows, err := r.query.Query(ctx, `SELECT network_flow_indicator_binding_id,incident_id,target_indicator_record_id,target_indicator_type,candidate_value,selector_kind,source_row_refs,source_row_refs_truncated,source_row_refs_total_count,created_by_user_id,created_at FROM network_flow_indicator_bindings WHERE incident_id=$1 AND network_flow_indicator_binding_id>$2 ORDER BY network_flow_indicator_binding_id LIMIT 64`, r.incidentID, cursor)
		if err != nil {
			return nil, err
		}
		return portabilityPage(rows, func(row pgx.Row) (portableBinding, error) {
			var v portableBinding
			err := row.Scan(&v.BindingID, &v.IncidentID, &v.IndicatorID, &v.IndicatorType, &v.CandidateValue, &v.SelectorKind, &v.SourceRowRefs, &v.SourceRowRefsTruncated, &v.SourceRowRefsTotalCount, &v.CreatedByUserID, &v.CreatedAt)
			v.CreatedAt = v.CreatedAt.UTC()
			return v, err
		})
	}, func(value *portableBinding) (string, *uuid.UUID) { return value.BindingID, &value.CreatedByUserID })
	if err != nil {
		return err
	}
	return exportPortabilityFamily(ctx, r, members, actors, "graph_views.ndjson", "network_flow_graph_views", func(after *portableGraph) ([]portableGraph, error) {
		cursor := ""
		if after != nil {
			cursor = after.GraphViewID
		}
		rows, err := r.query.Query(ctx, graphViewDeclarationSelect+` WHERE incident_id=$1 AND graph_view_id>$2 ORDER BY graph_view_id LIMIT 16`, r.incidentID, cursor)
		if err != nil {
			return nil, err
		}
		return portabilityPage(rows, func(row pgx.Row) (portableGraph, error) {
			v, err := scanGraphViewDeclaration(row)
			return portableGraphViewDeclaration(v), err
		})
	}, func(value *portableGraph) (string, *uuid.UUID) { return value.GraphViewID, &value.CreatedByUserID })
}

func requirePortabilityExportAccess(invocation incidentbundles.ExportInvocation) (interface {
	ExportRetained(context.Context, incidentportability.MemberWriter, incidentportability.ActorReferences) error
}, error) {
	access, ok := invocation.ReadAccess.(interface {
		ExportRetained(context.Context, incidentportability.MemberWriter, incidentportability.ActorReferences) error
	})
	if !ok || invocation.ProfileID != ProfileID || invocation.ContractMajor != 10 || invocation.ClaimState != incidentbundles.ClaimStateClaimed || invocation.ReadAccess.IncidentScope() != invocation.IncidentID {
		return nil, fmt.Errorf("%w: Network Flow snapshot scope", incidentbundles.ErrPortabilityUnavailable)
	}
	return access, nil
}
