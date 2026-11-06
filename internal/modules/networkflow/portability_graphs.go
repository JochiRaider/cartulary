package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

// This reader is limited to admitted immutable members. Graph reconstruction
// asks for retained sources only after admitting the selected-source binding.
func (p *preparedNetworkFlow) GetRetainedTable(_ context.Context, incidentID uuid.UUID, id string) (tableRecord, error) {
	if p != nil && incidentID == p.incidentID {
		if value, ok := p.tables[id]; ok {
			return value, nil
		}
	}
	return tableRecord{}, errTableNotFound
}
func (p *preparedNetworkFlow) GetActiveTable(ctx context.Context, incidentID uuid.UUID, id string) (tableRecord, error) {
	value, err := p.GetRetainedTable(ctx, incidentID, id)
	if err != nil {
		return value, err
	}
	if value.TableStatus != tableStatusActive {
		return tableRecord{}, errTableNotActive
	}
	return value, nil
}
func (p *preparedNetworkFlow) ListActiveTables(ctx context.Context, incidentID uuid.UUID) ([]tableRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p == nil || incidentID != p.incidentID {
		return nil, errTableNotFound
	}
	values := []tableRecord{}
	for _, value := range p.tables {
		if value.TableStatus == tableStatusActive {
			values = append(values, value)
		}
	}
	sort.Slice(values, func(i, j int) bool {
		if !values[i].CreatedAt.Equal(values[j].CreatedAt) {
			return values[i].CreatedAt.Before(values[j].CreatedAt)
		}
		return values[i].TableID < values[j].TableID
	})
	return values, nil
}
func (p *preparedNetworkFlow) IterateRowsForTables(ctx context.Context, incidentID uuid.UUID, tableIDs []string, visit func(flowRow) error) error {
	if p == nil || incidentID != p.incidentID || visit == nil {
		return errPortableNetworkFlow
	}
	// Each pass is a bounded sequential read in selected workspace order. Pure
	// composition selects canonical examples independently of source record order.
	for _, id := range tableIDs {
		if _, ok := p.tables[id]; !ok {
			return errTableNotFound
		}
		err := walkPortable(ctx, p.members, "observations.ndjson", portableObservationMembers, portableObservationNullable, func(v portableObservation) error {
			if v.NetworkFlowTableID != id {
				return nil
			}
			return visit(v.retained())
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *preparedNetworkFlow) prepareGraphs(ctx context.Context, preparation incidentbundles.PreparationMembers) error {
	previous := ""
	var active int64
	composer, err := newGraphSourceComposer(p, retainedSemanticLimits(), newGraphProjectionAdapter(), nil)
	if err != nil {
		return err
	}
	return walkPortable(ctx, p.members, "graph_views.ndjson", portableGraphMembers, "selected_source_binding selected_result retired_at", func(v portableGraph) error {
		d := v.retained()
		if bytes.Equal(bytes.TrimSpace(d.SelectedSourceBindingJSON), []byte("null")) {
			d.SelectedSourceBindingJSON = nil
		}
		if d.IncidentID != p.incidentID || d.GraphViewID <= previous || p.counts["graph_views"] >= p.limits.MaxRetainedGraphViewsPerIncident || !p.actor(d.CreatedByUserID) {
			return fmt.Errorf("graph identity, ordering or actor: %w", errPortableNetworkFlow)
		}
		if !validPortableTime(d.CreatedAt) || !validPortableTime(d.UpdatedAt) {
			return fmt.Errorf("graph timestamps: %w", errPortableNetworkFlow)
		}
		if !validGraphViewDeclaration(d) {
			return fmt.Errorf("graph declaration: %w", errPortableNetworkFlow)
		}
		if d.RetiredAt != nil && (!validPortableTime(*d.RetiredAt) || d.RetiredAt.After(d.UpdatedAt)) {
			return errPortableNetworkFlow
		}
		if d.DeclarationState == graphViewDeclarationStateActive {
			active++
			if active > p.limits.MaxActiveGraphViewsPerIncident {
				return errPortableNetworkFlow
			}
		}
		semantic, failure := decodeGraphSemanticRequest(d.SemanticQueryJSON, retainedSemanticLimits())
		if failure != nil {
			return failure
		}
		canonical, err := canonicalGraphSemanticJSON(semantic.Raw)
		if err != nil || graphViewSemanticQuerySHA256(canonical) != d.SemanticQuerySHA256 {
			return fmt.Errorf("graph semantic query digest: %w", errPortableNetworkFlow)
		}
		for _, id := range semantic.SelectedTableIDs {
			if _, ok := p.tables[id]; !ok {
				return errPortableNetworkFlow
			}
		}
		prepared := preparedPortableGraph{declaration: d}
		if d.SelectedResult != nil {
			if preparation == nil {
				return incidentbundles.ErrPortabilityUnavailable
			}
			composition, _, err := composer.composeSelectedSource(ctx, d)
			if err != nil {
				return err
			}
			encoded, err := json.Marshal(networkFlowProjectionInput(d.SelectedResult.SourceSnapshotID, composition))
			if err != nil {
				return err
			}
			result, err := graphprojection.ReconstructExactV2(ctx, d.SelectedResult.resultBinding(d.GraphViewID), encoded, d.UpdatedAt, nil)
			if err != nil {
				return err
			}
			prepared.path = "prepared/network_flow_activity/graphs/" + d.GraphViewID + ".json"
			err = incidentportability.WriteMemberFrom(ctx, preparation, prepared.path, func(ctx context.Context, out io.Writer) error { return json.NewEncoder(out).Encode(result) })
			if err != nil {
				return err
			}
			prepared.members, err = incidentportability.BindMembers(preparation, []string{prepared.path})
			if err != nil {
				return err
			}
		}
		p.graphs = append(p.graphs, prepared)
		p.counts["graph_views"]++
		previous = d.GraphViewID
		return nil
	})
}
