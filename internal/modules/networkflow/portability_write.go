package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// PortabilityAttributionFactory binds the attribution owner's adapter to the
// final transaction. The returned recorder must also be an
// incidentportability.AttributionSink, which flushes precomputed streams.
type PortabilityAttributionFactory func(context.Context, uuid.UUID, uuid.UUID) (incidentportability.AttributionRecorder, error)

type portabilityWriteAccess struct {
	tx           pgx.Tx
	module       *Module
	attributions PortabilityAttributionFactory
}

func (*portabilityWriteAccess) ParticipantScope() string { return PortabilityTransactionParticipantID }

// Application composition supplies both the final transaction and the separate
// attribution owner's adapter. Network Flow never reaches into its relations.
func (m *Module) PortabilityTransactionCapabilities(tx pgx.Tx, attributions PortabilityAttributionFactory) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error) {
	if m == nil || tx == nil || attributions == nil {
		return nil, nil, crossownertransaction.ErrUnavailable
	}
	access := &portabilityWriteAccess{tx: tx, module: m, attributions: attributions}
	return access, access, nil
}
func (a *portabilityWriteAccess) ValidatePortableState(ctx context.Context, p *portableImportPlan) error {
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

// WritePortableState applies only staged write-ready streams (Core 01
// REQ-01-641): one COPY per family, the bounded selected graph results, and
// one attribution COPY. It opens no bundle member and re-proves no bundle fact.
func (a *portabilityWriteAccess) WritePortableState(ctx context.Context, p *portableImportPlan) error {
	if a == nil || p == nil || a.tx == nil || p.preparation == nil {
		return crossownertransaction.ErrUnavailable
	}
	recorder, err := a.attributions(ctx, p.incidentID, p.actorID)
	if err != nil {
		return err
	}
	sink, ok := recorder.(incidentportability.AttributionSink)
	if !ok {
		return crossownertransaction.ErrUnavailable
	}
	if err := copyPreparedFamilies(ctx, a.tx, p); err != nil {
		return err
	}
	if err := a.publishPreparedGraphs(ctx, p); err != nil {
		return err
	}
	return sink.FlushImportedAttributions(ctx, p.preparation, p.attribution)
}

// publishPreparedGraphs reads the graph results reconstructed during
// preparation (staged outputs, not bundle members) and publishes them with
// their declarations under the importing actor.
func (a *portabilityWriteAccess) publishPreparedGraphs(ctx context.Context, p *portableImportPlan) error {
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
		declaration.CreatedByUserID = p.actorID
		if err := insertGraphViewDeclarationTx(ctx, a.tx, declaration); err != nil {
			return err
		}
	}
	return nil
}
