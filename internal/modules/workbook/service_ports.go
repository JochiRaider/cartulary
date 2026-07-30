package workbook

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type workbookMutationPort interface {
	CreateLinkedNote(ctx context.Context, actor authn.UserRecord, sourceRecordID uuid.UUID, request LinkedNoteCreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error)
	LinkedNoteSourceIncident(ctx context.Context, sourceRecordID uuid.UUID) (uuid.UUID, error)
	SupersedeDecision(ctx context.Context, actor authn.UserRecord, targetRecordID uuid.UUID, request timeline.SupersedeRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error)
}

type workbookConflictTokenClaims = conflicttokens.ConflictTokenClaims

type workbookRecordTargetPort interface {
	Resolve(ctx context.Context, recordID uuid.UUID) (recordRouteTarget, error)
}

type recordRouteTarget = records.RouteTarget

type workbookTimelineMutationPort interface {
	ApplyClipboardPaste(ctx context.Context, command timeline.ClipboardPasteCommand) (timeline.ClipboardPasteResult, error)
	ApplyFillDown(ctx context.Context, command timeline.FillDownCommand) (timeline.ClipboardPasteResult, error)
	ApplyMultiRowTagAssignment(ctx context.Context, command timeline.MultiRowTagAssignmentCommand) (timeline.ClipboardPasteResult, error)
	CreateRow(ctx context.Context, command timeline.CreateRowCommand) (timeline.MutationResult, error)
	PatchRow(ctx context.Context, command timeline.PatchRowCommand) (timeline.MutationResult, error)
	SupersedeRow(ctx context.Context, command timeline.SupersedeCommand) (timeline.MutationResult, error)
	ResolveConflict(ctx context.Context, command timeline.ConflictResolveCommand) (timeline.MutationResult, error)
}

type workbookEntityMutationPort interface {
	CreateHostRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request hostidentity.CreateRequest, requestHash []byte, requestID string, now time.Time) (hostidentity.MutationResult, error)
	CreateIdentityRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request hostidentity.CreateRequest, requestHash []byte, requestID string, now time.Time) (hostidentity.MutationResult, error)
	PatchEntityRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request hostidentity.PatchRequest, requestHash []byte, requestID string, now time.Time, routeKey string) (hostidentity.PatchMutationResult, error)
	ApplyClipboardPastePlan(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, viewSchemaID string, plan tabularingest.TabularRowPlanV1, requestHash []byte, requestID string, now time.Time) (hostidentity.ClipboardPasteResult, error)
}

type workbookConflictTokenPort interface {
	Parse(token string) (conflicttokens.ConflictTokenClaims, bool)
}
