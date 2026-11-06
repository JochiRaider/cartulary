package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PortabilityAttributionFactory func(context.Context, uuid.UUID, uuid.UUID, sourceport.ActorCatalog) (incidentportability.AttributionRecorder, error)

type portabilityWriteAccess struct {
	tx           pgx.Tx
	module       *Module
	attributions PortabilityAttributionFactory
}

func (*portabilityWriteAccess) ParticipantScope() string { return PortabilityTransactionParticipantID }

// Application composition supplies both the final transaction and the separate
// attribution owner's adapter. Network Flow never reaches into its relations.
func (m *Module) PortabilityTransactionCapabilities(tx pgx.Tx, attributions PortabilityAttributionFactory) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error) {
	if m == nil || tx == nil || attributions == nil || m.safeDigester == nil {
		return nil, nil, crossownertransaction.ErrUnavailable
	}
	access := &portabilityWriteAccess{tx: tx, module: m, attributions: attributions}
	return access, access, nil
}
func (a *portabilityWriteAccess) ValidatePortableState(ctx context.Context, p *preparedNetworkFlow) error {
	if a == nil || p == nil || a.module.graphStateReady == nil || !a.module.graphStateReady() {
		return crossownertransaction.ErrUnavailable
	}
	var present bool
	err := a.tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM network_flow_tables WHERE incident_id=$1 UNION ALL SELECT 1 FROM network_flow_graph_views WHERE incident_id=$1)`, p.incidentID).Scan(&present)
	if err != nil {
		return err
	}
	if present {
		return errPortableNetworkFlow
	}
	return ctx.Err()
}
func (a *portabilityWriteAccess) WritePortableState(ctx context.Context, p *preparedNetworkFlow) error {
	if a == nil || p == nil || a.tx == nil {
		return crossownertransaction.ErrUnavailable
	}
	attribution, err := a.attributions(ctx, p.incidentID, p.actorID, p.actors)
	if err != nil {
		return err
	}
	record := func(table, id string, actor uuid.UUID) error {
		return attribution.RecordImportedAttribution(table, id, "created_by_user_id", actor.String())
	}
	err = walkPortable(ctx, p.members, "tables.ndjson", portableTableMembers, "deleted_at", func(v portableTable) error {
		digest, key, err := a.module.safeDigester.Digest("source_filename", v.SourceFilenameDisplay)
		if err != nil {
			return err
		}
		_, err = insertTableTx(ctx, a.tx, v.TableID, insertTableParams{IncidentID: p.incidentID, DisplayName: v.DisplayName, CommittedSourceID: v.CommittedSourceID, SourceInterpretationSHA256: v.SourceInterpretationSHA256, SourceRawCellScalarLimit: v.SourceRawCellScalarLimit, SourceContentSHA256: v.SourceContentSHA256, SourceFilenameDisplay: v.SourceFilenameDisplay, SourceFilenameDigest: digest, SourceFilenameDigestKeyID: key, MappingFingerprint: v.MappingFingerprint, SourceProfileID: v.SourceProfileID, ParserProfileID: v.ParserProfileID, RowCountAccepted: v.RowCountAccepted, RowCountRejected: v.RowCountRejected, DiagnosticsTruncated: v.DiagnosticsTruncated, CreatedByUserID: p.actorID, CreatedAt: v.CreatedAt})
		if err != nil {
			return err
		}
		_, err = a.tx.Exec(ctx, `UPDATE network_flow_tables SET table_version=$3,table_status=$4,updated_at=$5,deleted_at=$6 WHERE incident_id=$1 AND network_flow_table_id=$2`, p.incidentID, v.TableID, v.TableVersion, v.TableStatus, v.UpdatedAt, v.DeletedAt)
		if err != nil {
			return err
		}
		return record("network_flow_tables", v.TableID, v.CreatedByUserID)
	})
	if err != nil {
		return err
	}
	err = insertRetainedRowsTx(ctx, a.tx, func(visit func(flowRow) error) error {
		return walkPortable(ctx, p.members, "observations.ndjson", portableObservationMembers, portableObservationNullable, func(v portableObservation) error {
			if err := record("network_flow_rows", v.RowID, v.CreatedByUserID); err != nil {
				return err
			}
			row := v.retained()
			row.CreatedByUserID = p.actorID
			return visit(row)
		})
	})
	if err != nil {
		return err
	}
	err = walkPortable(ctx, p.members, "diagnostics.ndjson", portableDiagnosticMembers, "", func(v portableDiagnostic) error {
		return insertDiagnosticsTx(ctx, a.tx, v.TableID, createTableParams{IncidentID: p.incidentID, Diagnostics: []rejectedRowDiagnostic{v.Diagnostic}}, v.CreatedAt)
	})
	if err != nil {
		return err
	}
	err = walkPortable(ctx, p.members, "indicator_bindings.ndjson", portableBindingMembers, "", func(v portableBinding) error {
		raw, err := json.Marshal(v.SourceRowRefs)
		if err != nil {
			return err
		}
		_, inserted, err := insertIndicatorBindingTx(ctx, a.tx, v.BindingID, createIndicatorBindingParams{IncidentID: p.incidentID, ActorUserID: p.actorID, TargetIndicator: indicators.IndicatorReference{RecordID: v.IndicatorID, IncidentID: p.incidentID, IndicatorType: v.IndicatorType, ValueKind: "atomic", NormalizedValue: &v.CandidateValue}, SelectorKind: v.SelectorKind, CandidateValue: v.CandidateValue, SourceRowRefs: v.SourceRowRefs, SourceRowRefsTruncated: v.SourceRowRefsTruncated, SourceRowRefsTotalCount: v.SourceRowRefsTotalCount}, raw, bindingSourceRowIDs(v.SourceRowRefs), v.CreatedAt)
		if err != nil {
			return err
		}
		if !inserted {
			return errPortableNetworkFlow
		}
		return record("network_flow_indicator_bindings", v.BindingID, v.CreatedByUserID)
	})
	if err != nil {
		return err
	}
	publisher, err := postgresresult.NewPublisher(a.tx)
	if err != nil {
		return err
	}
	for _, prepared := range p.graphs {
		declaration := prepared.declaration
		if declaration.IncidentID != p.incidentID {
			return errPortableNetworkFlow
		}
		if declaration.SelectedResult != nil {
			reader, _, err := prepared.members.OpenMember(ctx, prepared.path)
			if err != nil {
				return err
			}
			var result graphprojection.CompletedResultV2
			decoder := json.NewDecoder(reader)
			decoder.DisallowUnknownFields()
			err = decoder.Decode(&result)
			if err == nil {
				var extra any
				if decodeErr := decoder.Decode(&extra); decodeErr != io.EOF {
					err = errPortableNetworkFlow
				}
			}
			err = errors.Join(err, reader.Close())
			if err != nil {
				return err
			}
			if result.Binding != declaration.SelectedResult.resultBinding(declaration.GraphViewID) {
				return errPortableNetworkFlow
			}
			// Publish only these exact incident declarations. No deployment-wide clear,
			// enumeration, reconciliation or Recovery transaction is involved.
			if err := publisher.PublishResult(ctx, result); err != nil {
				return err
			}
		}
		original := declaration.CreatedByUserID
		declaration.CreatedByUserID = p.actorID
		if err := insertGraphViewDeclarationTx(ctx, a.tx, declaration); err != nil {
			return err
		}
		if err := record("network_flow_graph_views", declaration.GraphViewID, original); err != nil {
			return err
		}
	}
	return nil
}
