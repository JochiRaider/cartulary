package performancefixture

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type ProductionApplication struct {
	actor authn.UserRecord
	owner *timeline.PerformanceFixtureContribution
	now   func() time.Time
}

func NewProductionApplication(owner *timeline.PerformanceFixtureContribution, actor authn.UserRecord) (*ProductionApplication, error) {
	if owner == nil {
		return nil, fmt.Errorf("timeline performance fixture facade is required")
	}
	if actor.ID == uuid.Nil {
		return nil, fmt.Errorf("timeline performance fixture actor is required")
	}
	return &ProductionApplication{actor: actor, owner: owner, now: time.Now}, nil
}

func (a *ProductionApplication) CreateFixtureTimelineRows(ctx context.Context, incidentID string, rows []Row) error {
	incidentUUID, err := uuid.Parse(incidentID)
	if err != nil {
		return err
	}
	ownerRows := make([]timeline.PerformanceFixtureRow, len(rows))
	relationshipRows := 0
	for index := range rows {
		ownerRows[index] = timeline.PerformanceFixtureRow{
			Summary: rows[index].Summary, HostRef: rows[index].HostRef,
			IdentityRef: rows[index].IdentityRef, Tag: rows[index].Tag,
			DataSource: rows[index].DataSource,
		}
		if rows[index].HostRef != "" {
			relationshipRows++
		}
	}
	expected := timeline.PerformanceFixtureResult{RowCount: len(rows), RelationshipRows: relationshipRows}
	result, err := a.owner.CreatePerformanceFixtureRows(ctx, timeline.PerformanceFixtureCommand{
		Actor: a.actor, IncidentID: incidentUUID, Rows: ownerRows, Now: a.now().UTC(),
	})
	if err != nil {
		return err
	}
	if result != expected {
		return fmt.Errorf("timeline performance fixture created %#v, want %#v", result, expected)
	}
	return a.owner.ValidatePerformanceFixtureRows(ctx, incidentUUID, expected)
}
