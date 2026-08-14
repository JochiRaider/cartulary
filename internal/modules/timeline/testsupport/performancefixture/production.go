package performancefixture

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type ProductionApplication struct {
	actor  authn.UserRecord
	facade *timeline.Facade
	now    func() time.Time
	batch  int
}

func NewProductionApplication(facade *timeline.Facade, actor authn.UserRecord) (*ProductionApplication, error) {
	if facade == nil {
		return nil, fmt.Errorf("timeline performance fixture facade is required")
	}
	if actor.ID == uuid.Nil {
		return nil, fmt.Errorf("timeline performance fixture actor is required")
	}
	return &ProductionApplication{actor: actor, facade: facade, now: time.Now}, nil
}

func (a *ProductionApplication) CreateFixtureTimelineRows(ctx context.Context, incidentID string, rows []Row) error {
	incidentUUID, err := uuid.Parse(incidentID)
	if err != nil {
		return err
	}
	a.batch++
	clientTxnID := fmt.Sprintf("ac043-timeline-batch-%02d", a.batch)
	targets := make([]workbook.TimelineBatchTarget, len(rows))
	ownerTargets := make([]timeline.OwnerBatchTargetV1, len(rows))
	for index := range rows {
		targets[index] = workbook.TimelineBatchTarget{Kind: "create"}
		ownerTargets[index] = timeline.OwnerBatchTargetV1{Kind: "create"}
	}
	request := workbook.TimelineClipboardPasteRequest{
		ViewSchemaID:  timeline.TimelineViewSchemaID,
		ClientTxnID:   clientTxnID,
		ClipboardText: timelineTSV(rows),
		Format:        "tsv",
		StartFieldKey: "timeline.activity_synopsis_text",
		Columns: []string{
			"timeline.activity_synopsis_text",
			"timeline.host_refs",
			"timeline.identity_refs",
			"timeline.tags",
			"timeline.data_source_text",
		},
		Targets: targets,
	}
	plan, err := workbook.BuildTimelineClipboardPlan(request)
	if err != nil {
		return err
	}
	result, err := a.facade.ApplyClipboardPaste(ctx, timeline.ClipboardPasteCommand{
		Actor: a.actor, IncidentID: incidentUUID, ClientTxnID: clientTxnID,
		Plan: plan, Targets: ownerTargets, RequestHash: workbook.TimelineClipboardPasteRequestHash(request),
		RequestID: "req-" + clientTxnID, Now: a.now().UTC(),
	})
	if err != nil {
		return err
	}
	if len(result.Rows) != len(rows) {
		return fmt.Errorf("timeline performance fixture created %d rows, want %d", len(result.Rows), len(rows))
	}
	return nil
}

func timelineTSV(rows []Row) string {
	lines := make([]string, len(rows))
	for index, row := range rows {
		lines[index] = strings.Join([]string{row.Summary, row.HostRef, row.IdentityRef, row.Tag, row.DataSource}, "\t")
	}
	return strings.Join(lines, "\n")
}
