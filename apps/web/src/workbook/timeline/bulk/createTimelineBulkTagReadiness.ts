import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import {
  timelineCollectionBindings,
  timelineScalarBindings,
  timelineScalarEditorSurfaces,
} from "../models/timelineFieldRegistry";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { TimelineBulkTagReadiness } from "./useTimelineBulkTagController";

/** Reads existing authoring and queue ownership; never mirrors or flushes it. */
export function createTimelineBulkTagReadiness(input: {
  runtime: WorkbookMutationRuntime;
  drafts: TimelineEditorDraftRegistry;
  pending: TimelinePendingSavesRefs;
  rows: { readonly current: readonly WorkbookRow[] };
}): TimelineBulkTagReadiness {
  return {
    subscribe(listener) {
      const offRuntime = input.runtime.subscribe(listener);
      const offDrafts = input.drafts.subscribe(listener);
      return () => {
        offRuntime();
        offDrafts();
      };
    },
    blockingReason(ids) {
      const queue = input.pending.pendingQueueRef.current.model.snapshot();
      const halted = queue.units.find(
        (unit) => unit.id === queue.halted?.unit_id,
      );
      if (
        (halted?.recordId && ids.has(halted.recordId)) ||
        input.runtime
          .getSnapshot()
          .conflicts.some((entry) => ids.has(entry.conflict.record_id))
      )
        return "A selected record has an edit that needs recovery. Resolve or discard that edit before assigning the tag.";
      const admitted = queue.units.flatMap((unit) => {
        const revisions = input.pending.replayContextByUnitId.get(
          unit.id,
        )?.draftRevisions;
        return revisions ? [revisions] : [];
      });
      for (const row of input.rows.current) {
        if (!row.recordId || !ids.has(row.recordId)) continue;
        for (const surface of timelineScalarEditorSurfaces) {
          for (const binding of [
            ...timelineScalarBindings,
            ...timelineCollectionBindings,
          ]) {
            const scalar = "key" in binding;
            const field = scalar ? binding.key : binding.draftKey;
            const value = input.drafts.draftValue({
              rowKey: row.key,
              field,
              surface,
            });
            if (
              value === undefined ||
              (scalar
                ? value === row.committedValues[binding.key]
                : value.trim() === "")
            )
              continue;
            const revisions = input.drafts.captureRow(
              row.key,
              surface,
              new Set([binding.fieldKey]),
            );
            if (
              [...revisions].some(
                ([key, revision]) =>
                  !admitted.some((entry) => entry.get(key) === revision),
              )
            )
              return "A selected record has an unsaved edit. Finish or discard that edit before assigning the tag.";
          }
        }
      }
      return null;
    },
  };
}
