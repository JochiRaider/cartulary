package performancefixture

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ProductionApplication struct {
	runtime *projectionassembly.Runtime
}

func NewProductionApplication(runtime *projectionassembly.Runtime) (*ProductionApplication, error) {
	if runtime == nil {
		return nil, fmt.Errorf("projections performance fixture runtime is required")
	}
	return &ProductionApplication{runtime: runtime}, nil
}

func (a *ProductionApplication) ValidateFixtureProjectionSets(ctx context.Context, incidentID string, expectations []SetExpectation) error {
	incidentUUID, err := uuid.Parse(incidentID)
	if err != nil {
		return err
	}
	for _, expectation := range expectations {
		viewSchemaID := expectation.ViewSchemaID
		want := expectation.ExactRows
		schema, ok := viewschema.Lookup(viewSchemaID)
		if !ok {
			return fmt.Errorf("fixture projection schema %s is unavailable", viewSchemaID)
		}
		window := querypage.Window{Limit: want + 1}
		switch viewSchemaID {
		case "cartulary.view.hosts.v1":
			reader := a.runtime.EntityPorts().Reader
			if reader == nil {
				return fmt.Errorf("fixture projection provider %s is unavailable", viewSchemaID)
			}
			rows, err := reader.SelectHostQueryProjections(ctx, incidentUUID, schema.DefaultQueryMeta(), window)
			if err != nil {
				return fmt.Errorf("query fixture projection %s: %w", viewSchemaID, err)
			}
			if len(rows) != want {
				return fmt.Errorf("fixture projection %s rows=%d want=%d", viewSchemaID, len(rows), want)
			}
		case "cartulary.view.identities.v1":
			reader := a.runtime.EntityPorts().Reader
			if reader == nil {
				return fmt.Errorf("fixture projection provider %s is unavailable", viewSchemaID)
			}
			rows, err := reader.SelectIdentityQueryProjections(ctx, incidentUUID, schema.DefaultQueryMeta(), window)
			if err != nil {
				return fmt.Errorf("query fixture projection %s: %w", viewSchemaID, err)
			}
			if len(rows) != want {
				return fmt.Errorf("fixture projection %s rows=%d want=%d", viewSchemaID, len(rows), want)
			}
		default:
			provider, ok := a.runtime.WorkbookQueryProvider(viewSchemaID)
			if !ok {
				return fmt.Errorf("fixture projection provider %s is unavailable", viewSchemaID)
			}
			result, err := provider.QueryRowsPage(ctx, workbook.QueryCommand{
				IncidentID:   incidentUUID,
				ViewSchemaID: viewSchemaID,
				Query:        schema.DefaultQueryMeta(),
				Window:       window,
			})
			if err != nil {
				return fmt.Errorf("query fixture projection %s: %w", viewSchemaID, err)
			}
			if len(result.Rows) != want || result.HasMore {
				return fmt.Errorf("fixture projection %s rows=%d has_more=%t want=%d", viewSchemaID, len(result.Rows), result.HasMore, want)
			}
		}
	}
	return nil
}
