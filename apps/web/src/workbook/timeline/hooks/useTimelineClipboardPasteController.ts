import type {
  GridCellPasteIntent,
  GridClipboardInput,
} from "@cartulary/grid-adapter";
import { useCallback } from "react";
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
import type {
  TimelinePasteTargetResolution,
  TimelineScalarSaveOptions,
} from "../models/timelineControllerPorts";
import type {
  RowValues,
  TimelineScalarEditorSurface,
} from "../models/timelineFieldRegistry";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";

export function useTimelineClipboardPasteController(input: {
  readonly canCreateRows: boolean;
  readonly editable: boolean;
  readonly grouped: boolean;
  readonly clipboardPaste: WorkbookClipboardPastePort;
  readonly pendingSavesRefs: TimelinePendingSavesRefs;
  readonly queueScalarSave: (
    rowKey: string,
    focusField: keyof RowValues,
    options: TimelineScalarSaveOptions,
    currentValue?: string,
  ) => void;
  readonly resolveTimelinePasteTargetResolution: (
    rowKey: string,
    fieldKey: string,
    input: GridClipboardInput,
  ) => TimelinePasteTargetResolution | null;
  readonly setError: (message: string | null) => void;
}) {
  const handlePaste = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
      value: string,
    ) => {
      input.queueScalarSave(
        rowKey,
        focusField,
        { continueOnFreshDraft: false, preserveInputFocus: true, surface },
        value,
      );
    },
    [input.queueScalarSave],
  );
  const handleGridPaste = useCallback(
    (intent: GridCellPasteIntent) => {
      if (intent.input.kind !== "table") return false;
      const rowKey =
        intent.target.rowIdentity.kind === "core_record"
          ? intent.target.rowIdentity.recordId
          : "";
      const current = input.resolveTimelinePasteTargetResolution(
        rowKey,
        intent.target.fieldKey,
        intent.input,
      );
      if (
        !current ||
        !timelinePasteTargetPlansMatch(
          intent.targetResolution,
          current.targetResolution,
        ) ||
        timelinePastePlanAdmission(
          current.targetResolution,
          intent.input,
          input,
        ).kind !== "accepted"
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
      return (
        input.clipboardPaste.paste(
          {
            clipboard_text: intent.input.rawText,
            format: intent.input.format,
            header_mode: intent.input.headerMode ?? "auto",
            start_field_key:
              intent.input.fieldKeys?.[0] ?? intent.target.fieldKey,
            columns,
            targets,
            view_schema_id: timelineViewSchemaId,
          },
          {
            delivery: intent,
            ready: input.pendingSavesRefs.saveQueueRef.current,
          },
        ) !== null
      );
    },
    [input],
  );
  return { commands: { handlePaste, handleGridPaste } };
}
