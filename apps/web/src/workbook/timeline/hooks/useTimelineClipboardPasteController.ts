import type {
  GridCellPasteIntent,
  GridClipboardInput,
  GridDraftPasteIntent,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import { useCallback, useEffect, useRef } from "react";
import type { WorkbookClipboardPastePort } from "../../adapters/WorkbookClipboardPastePort";
import {
  workbookPasteColumns,
  workbookPasteTargets,
} from "../../models/workbookClipboardPaste";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  timelinePastePlanAdmission,
  timelinePasteTargetPlansMatch,
} from "../models/timelineClipboardPastePlan";
import type { TimelinePasteTargetResolution } from "../models/timelineControllerPorts";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";

export function useTimelineClipboardPasteController(input: {
  readonly canCreateRows: boolean;
  readonly editable: boolean;
  readonly grouped: boolean;
  readonly clipboardPaste: WorkbookClipboardPastePort;
  readonly captureCreatedBatchPresentation: () => (
    recordIds: readonly string[],
  ) => boolean;
  readonly captureCreatedPasteReveal: (
    fieldKey: string,
  ) => (recordId: string) => void;
  readonly pendingSavesRefs: Pick<TimelinePendingSavesRefs, "saveQueueRef">;
  readonly resolveTimelinePasteTargetResolution: (
    rowKey: string,
    fieldKey: string,
    input: GridClipboardInput,
  ) => TimelinePasteTargetResolution | null;
  readonly resolveTimelineDraftPasteTargetResolution: (
    fieldKey: string,
    input: GridClipboardInput,
  ) => TimelinePasteTargetResolution | null;
  readonly setError: (message: string | null) => void;
}) {
  const observations = useRef(new Set<() => void>());
  const currentAccess = useRef(input.editable);
  currentAccess.current = input.editable;
  useEffect(
    () => () => {
      for (const cancel of observations.current) cancel();
      observations.current.clear();
    },
    [],
  );
  const admitPaste = useCallback(
    (
      intent: {
        readonly input: Extract<GridClipboardInput, { readonly kind: "table" }>;
        readonly targetResolution: GridPasteTargetResolution;
      },
      fieldKey: string,
      current: TimelinePasteTargetResolution | null,
    ) => {
      if (
        !current ||
        !timelinePasteTargetPlansMatch(
          intent.targetResolution,
          current.targetResolution,
        ) ||
        timelinePastePlanAdmission(current.targetResolution, intent.input, {
          canCreateRows: input.canCreateRows,
          editable: input.editable,
          grouped: input.grouped,
        }).kind !== "accepted"
      ) {
        input.setError(
          "Paste targets changed or are unavailable for this Timeline.",
        );
        return false;
      }
      const columns = workbookPasteColumns(current.targetResolution.columns);
      const targets = workbookPasteTargets(
        current.targetResolution.rowTargets.map((target) =>
          target.kind === "create"
            ? { kind: "create" as const }
            : {
                kind: "record" as const,
                record_id: target.rowIdentity.recordId,
                base_row_version: target.mutationIdentity.baseRowVersion,
              },
        ),
      );
      if (!columns || !targets) {
        input.setError("The selected paste range is unavailable.");
        return false;
      }
      input.setError(null);
      const creates = targets.some((target) => target.kind === "create");
      const revealCreated = creates
        ? input.captureCreatedPasteReveal(fieldKey)
        : null;
      const acceptedAt = performance.now();
      const receipt = input.clipboardPaste.paste(
        {
          clipboard_text: intent.input.rawText,
          format: intent.input.format,
          header_mode: intent.input.headerMode ?? "auto",
          start_field_key: intent.input.fieldKeys?.[0] ?? fieldKey,
          columns,
          targets,
          view_schema_id: timelineViewSchemaId,
        },
        {
          delivery: intent,
          ready: input.pendingSavesRefs.saveQueueRef.current,
        },
      );
      if (receipt === null) return false;
      const presentCreated = creates
        ? input.captureCreatedBatchPresentation()
        : null;
      performance.mark("cartulary.workbook.timeline_paste_accepted", {
        startTime: acceptedAt,
      });
      if (presentCreated && revealCreated) {
        const existingIds = new Set(
          targets.flatMap((target) =>
            target.kind === "record" ? [target.record_id] : [],
          ),
        );
        const cancel = input.clipboardPaste.observeReceipt(
          receipt,
          (accepted) => {
            observations.current.delete(cancel);
            if (!accepted || !currentAccess.current) return;
            const createdIds = accepted.rows
              .map((row) => row.record_id)
              .filter((id) => !existingIds.has(id));
            if (presentCreated(createdIds)) {
              const last = createdIds.at(-1);
              if (last !== undefined) revealCreated(last);
            }
          },
        );
        observations.current.add(cancel);
      }
      return true;
    },
    [
      input.canCreateRows,
      input.editable,
      input.grouped,
      input.clipboardPaste,
      input.captureCreatedBatchPresentation,
      input.captureCreatedPasteReveal,
      input.pendingSavesRefs,
      input.setError,
    ],
  );
  const handleGridPaste = useCallback(
    (intent: GridCellPasteIntent) => {
      if (
        intent.input.kind !== "table" ||
        intent.target.rowIdentity.kind !== "core_record"
      )
        return false;
      return admitPaste(
        { ...intent, input: intent.input },
        intent.target.fieldKey,
        input.resolveTimelinePasteTargetResolution(
          intent.target.rowIdentity.recordId,
          intent.target.fieldKey,
          intent.input,
        ),
      );
    },
    [admitPaste, input.resolveTimelinePasteTargetResolution],
  );
  const handleDraftPaste = useCallback(
    (intent: GridDraftPasteIntent) => {
      const current = input.resolveTimelineDraftPasteTargetResolution(
        intent.fieldKey,
        intent.input,
      );
      if (
        current?.targetResolution.rowTargets.some(
          (target) => target.kind !== "create",
        )
      )
        return false;
      return admitPaste(intent, intent.fieldKey, current);
    },
    [admitPaste, input.resolveTimelineDraftPasteTargetResolution],
  );
  return { commands: { handleGridPaste, handleDraftPaste } };
}
