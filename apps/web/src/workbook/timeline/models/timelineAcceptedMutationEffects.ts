import type { TimelineAcceptedProjection } from "./timelineAcceptedProjection";
import type { WorkbookRow } from "./timelineRowModel";
import type { AutoResolutionNotice } from "./workbookMentionChips";
import { buildAutoResolutionNotices } from "./workbookMentionChips";

export type TimelineAcceptedContinuity =
  | {
      readonly kind: "fresh_draft";
      readonly focusKey: string;
      readonly recordId: string;
    }
  | {
      readonly kind: "advance";
      readonly target:
        | { readonly kind: "input"; readonly focusKey: string }
        | { readonly kind: "row-inspect"; readonly recordId: string }
        | null;
    };

export type TimelineAcceptedMutationEffects = {
  readonly autoResolutionNotices: readonly AutoResolutionNotice[];
  readonly continuity: TimelineAcceptedContinuity;
  readonly createdRecordId: string | null;
  readonly reconcileDismissedMentions: boolean;
  readonly selectionUpdate: { readonly recordId: string | null } | null;
};

export function planTimelineAcceptedMutationEffects({
  committed,
  continueOnFreshDraft,
  detectAutoResolution,
  projection,
  promoteToCommittedRowInspect,
  selectedRowId,
}: {
  readonly committed: WorkbookRow;
  readonly continueOnFreshDraft: boolean;
  readonly detectAutoResolution: boolean;
  readonly projection: TimelineAcceptedProjection;
  readonly promoteToCommittedRowInspect: boolean;
  readonly selectedRowId: string | null;
}): TimelineAcceptedMutationEffects {
  const recordId = committed.recordId;
  const selectionUpdate =
    selectedRowId !== null && projection.previousRow?.recordId === selectedRowId
      ? { recordId }
      : null;
  const continuity: TimelineAcceptedContinuity =
    projection.createdFromDraft &&
    continueOnFreshDraft &&
    projection.draftFocusKey !== null &&
    recordId !== null
      ? {
          kind: "fresh_draft",
          focusKey: projection.draftFocusKey,
          recordId,
        }
      : {
          kind: "advance",
          target:
            continueOnFreshDraft && projection.draftFocusKey !== null
              ? { kind: "input", focusKey: projection.draftFocusKey }
              : promoteToCommittedRowInspect && recordId !== null
                ? { kind: "row-inspect", recordId }
                : null,
        };
  return {
    autoResolutionNotices: detectAutoResolution
      ? buildAutoResolutionNotices(projection.previousRow, committed)
      : [],
    continuity,
    createdRecordId: projection.createdFromDraft ? recordId : null,
    reconcileDismissedMentions: recordId !== null,
    selectionUpdate,
  };
}
