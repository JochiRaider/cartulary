package performancefixture

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type ProductionApplication struct {
	actor authn.UserRecord
	store *hostidentity.Store
	now   func() time.Time
}

func NewProductionApplication(store *hostidentity.Store, actor authn.UserRecord) (*ProductionApplication, error) {
	if store == nil {
		return nil, fmt.Errorf("entities performance fixture store is required")
	}
	if actor.ID == uuid.Nil {
		return nil, fmt.Errorf("entities performance fixture actor is required")
	}
	return &ProductionApplication{actor: actor, store: store, now: time.Now}, nil
}

func (a *ProductionApplication) CreateFixtureHosts(ctx context.Context, incidentID string, hosts []Host) error {
	rows := make([][]string, len(hosts))
	for index, host := range hosts {
		rows[index] = []string{host.DisplayName, host.Hostname, host.Hostname}
	}
	return a.apply(ctx, incidentID, hostidentity.HostsViewSchemaID, "host.display_name", []string{"host.display_name", "host.hostname", "host.aliases"}, rows)
}

func (a *ProductionApplication) CreateFixtureIdentities(ctx context.Context, incidentID string, identities []Identity) error {
	rows := make([][]string, len(identities))
	for index, identity := range identities {
		rows[index] = []string{identity.DisplayName, identity.UPN, identity.UPN}
	}
	return a.apply(ctx, incidentID, hostidentity.IdentitiesViewSchemaID, "identity.display_name", []string{"identity.display_name", "identity.upn", "identity.aliases"}, rows)
}

func (a *ProductionApplication) apply(ctx context.Context, incidentID string, viewSchemaID string, startField string, columns []string, rows [][]string) error {
	incidentUUID, err := uuid.Parse(incidentID)
	if err != nil {
		return err
	}
	clientTxnID := fmt.Sprintf("ac043-%s-%s", strings.TrimPrefix(viewSchemaID, "cartulary.view."), rows[0][1])
	request := hostidentity.ClipboardPasteRequest{
		ViewSchemaID: viewSchemaID, ClientTxnID: clientTxnID, ClipboardText: encodeTSV(rows),
		Format: "tsv", StartFieldKey: startField, Columns: columns, CreateOnlyRows: len(rows),
	}
	plan, err := hostidentity.BuildClipboardPastePlan(request)
	if err != nil {
		return err
	}
	result, err := a.store.ApplyClipboardPastePlan(ctx, a.actor, incidentUUID, viewSchemaID, plan, request.RequestHash(), "req-"+clientTxnID, a.now().UTC())
	if err != nil {
		return err
	}
	if len(result.Rows) != len(rows) {
		return fmt.Errorf("entities performance fixture created %d rows, want %d", len(result.Rows), len(rows))
	}
	return nil
}

func encodeTSV(rows [][]string) string {
	lines := make([]string, len(rows))
	for index, row := range rows {
		lines[index] = strings.Join(row, "\t")
	}
	return strings.Join(lines, "\n")
}
