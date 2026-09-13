import { useLayoutEffect, useRef } from "react";
import type { WorkbookTimelineRelatedEvidenceOwner } from "../../features/evidence/WorkbookTimelineRelatedEvidenceOwner";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelineCommittedRecordIdleResult } from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/timelineRowModel";

/** Earlier local edits participate in coordination; explicit writes never enter autosave capacity. */
export function useTimelineRelatedEvidenceSource(options: {
  readonly owner: WorkbookTimelineRelatedEvidenceOwner;
  readonly rows: { readonly current: WorkbookRow[] };
  readonly drafts: TimelineEditorDraftRegistry;
  readonly available: boolean;
  readonly waitForIdle: (
    id: string,
    options: { signal: AbortSignal; refreshIfMissing: boolean },
  ) => Promise<TimelineCommittedRecordIdleResult | null>;
}) {
  const current = useRef(options);
  current.current = options;
  useLayoutEffect(
    () =>
      options.owner.registerSourceCoordinator(async (recordId, signal) => {
        const clean = () => {
          if (!current.current.available) return false;
          const row = current.current.rows.current.find(
            (row) => row.recordId === recordId,
          );
          if (!row) return true;
          const materialized = current.current.drafts.materializeRow(row);
          return (
            !Object.entries(materialized.values).some(
              ([key, value]) =>
                value !==
                row.committedValues[key as keyof typeof row.committedValues],
            ) &&
            !Object.values(materialized.collectionDrafts).some(
              (value) => value.trim() !== "",
            )
          );
        };
        if (!clean()) return false;
        if (
          current.current.rows.current.some((row) => row.recordId === recordId)
        ) {
          const idle = await current.current.waitForIdle(recordId, {
            signal,
            refreshIfMissing: false,
          });
          if (!idle) return false;
        }
        return !signal.aborted && clean();
      }),
    [options.owner],
  );
}
