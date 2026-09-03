import type {
  GridCellAnchor,
  GridCellPasteIntent,
  GridClipboardInput,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import { useCallback, useRef } from "react";
import type {
  WorkbookClipboardPasteAccepted,
  WorkbookClipboardPasteInput,
  WorkbookClipboardPastePort,
} from "../../adapters/WorkbookClipboardPastePort";
import {
  workbookPasteColumns,
  workbookPasteTargets,
} from "../../models/workbookClipboardPaste";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { parseSameFieldConflictPayload } from "../../runtime/workbookConflictModel";
import { sameFieldConflictQueueKey } from "../../utils/workbookPendingQueue";
import {
  type TimelinePasteAuthority,
  timelinePastePlanAdmission,
  timelinePasteRequestTargetsMatchResolution,
  timelinePasteTargetPlansMatch,
} from "../models/timelineClipboardPastePlan";
import type { SameFieldConflictPayload } from "../models/timelineConflictState";
import type {
  TimelinePasteTargetResolution,
  TimelineScalarSaveOptions,
} from "../models/timelineControllerPorts";
import {
  inputFocusKey,
  type RowValues,
  type TimelineScalarEditorSurface,
  timelineScalarBindingForField,
} from "../models/timelineFieldRegistry";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import {
  normalizeTimelineFullRow,
  type TimelineApiRow,
} from "../models/timelineRowModel";

type TimelineTableInput = Extract<
  GridClipboardInput,
  { readonly kind: "table" }
>;

type TimelineLoadRowsForPaste = (options: {
  readonly showLoading: boolean;
  readonly viewportContinuityToken?: number;
}) => Promise<void>;

type TimelineQueueScalarSave = (
  rowKey: string,
  focusField: keyof RowValues,
  options: TimelineScalarSaveOptions,
  currentValue?: string,
) => void;

type TimelineCommittedRecordIdle = {
  readonly rowVersion: number;
};

type ResolveTimelinePasteTargetResolution = (
  rowKey: string,
  fieldKey: string,
  input: GridClipboardInput,
) => TimelinePasteTargetResolution | null;

type DecodedTimelinePaste = {
  readonly conflicts: readonly SameFieldConflictPayload[];
  readonly rows: readonly TimelineApiRow[];
};

type AdmittedTimelinePaste = {
  readonly bindingKey: keyof RowValues;
  readonly fieldKey: string;
  readonly initial: TimelinePasteTargetResolution;
  readonly input: TimelineTableInput;
  readonly rowKey: string;
};

type TimelinePasteAdmission =
  | { readonly kind: "ignored" }
  | { readonly kind: "rejected"; readonly message: string }
  | { readonly kind: "admitted"; readonly value: AdmittedTimelinePaste };

type TimelinePasteExecutionPorts = {
  readonly applyResponseRows: (rows: readonly TimelineApiRow[]) => void;
  readonly authority: () => TimelinePasteAuthority;
  readonly clearViewportContinuity: (token: number) => void;
  readonly clipboardPaste: WorkbookClipboardPastePort;
  readonly finishSave: (nextState: "Conflict" | "Saved" | "Syncing") => void;
  readonly loadRows: TimelineLoadRowsForPaste;
  readonly registerSameFieldConflict: (
    conflict: SameFieldConflictPayload,
    focusKey: string,
    surface: TimelineScalarEditorSurface,
  ) => void;
  readonly resolvePendingSocketTxn: (
    clientTxnId: string | null | undefined,
  ) => boolean;
  readonly resolveTarget: ResolveTimelinePasteTargetResolution;
  readonly restoreFocus: (anchor: GridCellAnchor) => boolean;
  readonly setActiveConflictKey: (key: string | null) => void;
  readonly setError: (message: string | null) => void;
  readonly setPasteConflictGroup: (group: { keys: string[] } | null) => void;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<TimelineCommittedRecordIdle | null>;
};

function decodeTimelinePaste(
  accepted: WorkbookClipboardPasteAccepted,
): DecodedTimelinePaste | null {
  try {
    const rows = accepted.rows.map((row) =>
      normalizeTimelineFullRow(row, "clipboard paste response row"),
    );
    const conflicts: SameFieldConflictPayload[] = [];
    for (const value of accepted.conflicts) {
      const conflict = parseSameFieldConflictPayload(value);
      if (conflict === null) return null;
      conflicts.push(conflict);
    }
    return { conflicts, rows };
  } catch {
    return null;
  }
}

function resolutionIsCurrentAndAdmitted(
  reference: GridPasteTargetResolution,
  candidate: TimelinePasteTargetResolution | null,
  input: TimelineTableInput,
  authority: TimelinePasteAuthority,
): candidate is TimelinePasteTargetResolution {
  return (
    candidate !== null &&
    timelinePasteTargetPlansMatch(reference, candidate.targetResolution) &&
    timelinePastePlanAdmission(candidate.targetResolution, input, authority)
      .kind === "accepted"
  );
}

function admitTimelineGridPaste(
  intent: GridCellPasteIntent,
  resolveTarget: ResolveTimelinePasteTargetResolution,
  authority: TimelinePasteAuthority,
): TimelinePasteAdmission {
  if (intent.input.kind !== "table") return { kind: "ignored" };
  const binding = timelineScalarBindingForField(intent.target.fieldKey);
  const targetsTimeline =
    intent.target.rowIdentity.kind === "core_record" &&
    intent.target.surface.kind === "view_schema" &&
    intent.target.surface.viewSchemaId === timelineViewSchemaId;
  if (binding === null || !targetsTimeline) {
    return {
      kind: "rejected",
      message: "Paste targets are unavailable for the active Timeline.",
    };
  }
  const rowKey = intent.target.rowIdentity.recordId;
  const initial = resolveTarget(rowKey, intent.target.fieldKey, intent.input);
  if (
    !resolutionIsCurrentAndAdmitted(
      intent.targetResolution,
      initial,
      intent.input,
      authority,
    )
  ) {
    return {
      kind: "rejected",
      message: "Paste targets changed or are unavailable for this Timeline.",
    };
  }
  return {
    kind: "admitted",
    value: {
      bindingKey: binding.key,
      fieldKey: intent.target.fieldKey,
      initial,
      input: intent.input,
      rowKey,
    },
  };
}

async function buildTimelinePasteRequestTargets(
  resolution: GridPasteTargetResolution,
  waitForCommittedRecordIdle: TimelinePasteExecutionPorts["waitForCommittedRecordIdle"],
): Promise<readonly WorkbookClipboardPasteInput["targets"][number][] | null> {
  const values: WorkbookClipboardPasteInput["targets"][number][] = [];
  for (const target of resolution.rowTargets) {
    if (target.kind === "create") {
      values.push({ kind: "create" });
      continue;
    }
    const idleRecord = await waitForCommittedRecordIdle(
      target.rowIdentity.recordId,
    );
    if (idleRecord === null) return null;
    values.push({
      base_row_version: idleRecord.rowVersion,
      kind: "record",
      record_id: target.rowIdentity.recordId,
    });
  }
  return values;
}

function rejectTimelinePaste(
  paste: AdmittedTimelinePaste,
  viewportContinuityToken: number,
  message: string,
  ports: TimelinePasteExecutionPorts,
) {
  ports.clearViewportContinuity(viewportContinuityToken);
  if (paste.initial.anchor !== null) ports.restoreFocus(paste.initial.anchor);
  ports.setError(message);
  ports.finishSave("Conflict");
}

function projectTimelinePasteConflicts(
  conflicts: readonly SameFieldConflictPayload[],
  fallbackBindingKey: keyof RowValues,
  ports: TimelinePasteExecutionPorts,
) {
  const keys = conflicts.map((conflict) => {
    const binding = timelineScalarBindingForField(conflict.field_key);
    const key = sameFieldConflictQueueKey(conflict);
    ports.registerSameFieldConflict(
      conflict,
      inputFocusKey(
        conflict.record_id,
        binding?.key ?? fallbackBindingKey,
        "grid",
      ),
      "grid",
    );
    return key;
  });
  if (keys.length > 1) {
    ports.setPasteConflictGroup({ keys });
    ports.setActiveConflictKey(keys[0] ?? null);
  } else if (keys.length === 0) {
    ports.setPasteConflictGroup(null);
  }
}

function finalTimelinePasteRequest(
  paste: AdmittedTimelinePaste,
  current: TimelinePasteTargetResolution,
  requestTargetValues: readonly WorkbookClipboardPasteInput["targets"][number][],
  ports: TimelinePasteExecutionPorts,
): WorkbookClipboardPasteInput | null {
  const finalResolution = ports.resolveTarget(
    paste.rowKey,
    paste.fieldKey,
    paste.input,
  );
  const columns = workbookPasteColumns(current.targetResolution.columns);
  const targets = workbookPasteTargets(requestTargetValues);
  if (
    !resolutionIsCurrentAndAdmitted(
      current.targetResolution,
      finalResolution,
      paste.input,
      ports.authority(),
    ) ||
    columns === null ||
    targets === null ||
    !timelinePasteRequestTargetsMatchResolution(
      requestTargetValues,
      finalResolution.targetResolution,
    )
  ) {
    return null;
  }
  return {
    clipboard_text: paste.input.rawText,
    columns,
    format: paste.input.format,
    onClientTxnId: ports.trackPendingSocketTxn,
    start_field_key: paste.fieldKey,
    targets,
    view_schema_id: timelineViewSchemaId,
  };
}

async function executeTimelineGridPaste(
  paste: AdmittedTimelinePaste,
  viewportContinuityToken: number,
  ports: TimelinePasteExecutionPorts,
) {
  const changedMessage =
    "Paste targets changed or are unavailable for this Timeline.";
  const current = ports.resolveTarget(
    paste.rowKey,
    paste.fieldKey,
    paste.input,
  );
  if (
    !resolutionIsCurrentAndAdmitted(
      paste.initial.targetResolution,
      current,
      paste.input,
      ports.authority(),
    )
  ) {
    rejectTimelinePaste(paste, viewportContinuityToken, changedMessage, ports);
    return;
  }
  const targetValues = await buildTimelinePasteRequestTargets(
    current.targetResolution,
    ports.waitForCommittedRecordIdle,
  );
  if (targetValues === null) {
    rejectTimelinePaste(
      paste,
      viewportContinuityToken,
      "A Timeline record changed before paste could be submitted.",
      ports,
    );
    return;
  }
  const request = finalTimelinePasteRequest(
    paste,
    current,
    targetValues,
    ports,
  );
  if (request === null) {
    rejectTimelinePaste(paste, viewportContinuityToken, changedMessage, ports);
    return;
  }
  const result = await ports.clipboardPaste.paste(request);
  ports.resolvePendingSocketTxn(result.clientTxnId);
  if (result.outcome.kind === "rejected") {
    rejectTimelinePaste(
      paste,
      viewportContinuityToken,
      result.outcome.failure.message,
      ports,
    );
    return;
  }
  const decoded = decodeTimelinePaste(result.outcome.value);
  if (decoded === null) {
    rejectTimelinePaste(
      paste,
      viewportContinuityToken,
      "The Timeline paste response was invalid.",
      ports,
    );
    return;
  }
  projectTimelinePasteConflicts(decoded.conflicts, paste.bindingKey, ports);
  ports.applyResponseRows(decoded.rows);
  await ports.loadRows({
    showLoading: false,
    viewportContinuityToken,
  });
  if (paste.initial.anchor !== null) ports.restoreFocus(paste.initial.anchor);
  ports.finishSave(decoded.conflicts.length > 0 ? "Conflict" : "Saved");
}

export function useTimelineClipboardPasteController({
  applyResponseRows,
  beginSave,
  beginViewportContinuity,
  canCreateRows,
  clearViewportContinuity,
  clipboardPaste,
  editable,
  finishSave,
  grouped,
  loadRows,
  pendingSavesRefs,
  queueScalarSave,
  registerSameFieldConflict,
  resolvePendingSocketTxn,
  resolveTimelinePasteTargetResolution,
  restoreTimelineFocusAnchor,
  setActiveConflictKey,
  setError,
  setPasteConflictGroup,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly applyResponseRows: (rows: readonly TimelineApiRow[]) => void;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    target:
      | { readonly kind: "input"; readonly focusKey: string }
      | { readonly kind: "row-inspect"; readonly recordId: string }
      | { readonly kind: "scroll-only" },
  ) => number;
  readonly canCreateRows: boolean;
  readonly clearViewportContinuity: (token: number) => void;
  readonly clipboardPaste: WorkbookClipboardPastePort;
  readonly editable: boolean;
  readonly finishSave: (nextState: "Conflict" | "Saved" | "Syncing") => void;
  readonly grouped: boolean;
  readonly loadRows: TimelineLoadRowsForPaste;
  readonly pendingSavesRefs: TimelinePendingSavesRefs;
  readonly queueScalarSave: TimelineQueueScalarSave;
  readonly registerSameFieldConflict: TimelinePasteExecutionPorts["registerSameFieldConflict"];
  readonly resolvePendingSocketTxn: TimelinePasteExecutionPorts["resolvePendingSocketTxn"];
  readonly resolveTimelinePasteTargetResolution: ResolveTimelinePasteTargetResolution;
  readonly restoreTimelineFocusAnchor: (anchor: GridCellAnchor) => boolean;
  readonly setActiveConflictKey: TimelinePasteExecutionPorts["setActiveConflictKey"];
  readonly setError: TimelinePasteExecutionPorts["setError"];
  readonly setPasteConflictGroup: TimelinePasteExecutionPorts["setPasteConflictGroup"];
  readonly trackPendingSocketTxn: TimelinePasteExecutionPorts["trackPendingSocketTxn"];
  readonly waitForCommittedRecordIdle: TimelinePasteExecutionPorts["waitForCommittedRecordIdle"];
}) {
  const authorityRef = useRef<TimelinePasteAuthority>({
    canCreateRows,
    editable,
    grouped,
  });
  authorityRef.current = { canCreateRows, editable, grouped };

  const handlePaste = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
      value: string,
    ) => {
      queueScalarSave(
        rowKey,
        focusField,
        {
          continueOnFreshDraft: false,
          preserveInputFocus: true,
          surface,
        },
        value,
      );
    },
    [queueScalarSave],
  );

  const handleGridPaste = useCallback(
    (intent: GridCellPasteIntent) => {
      const admission = admitTimelineGridPaste(
        intent,
        resolveTimelinePasteTargetResolution,
        authorityRef.current,
      );
      if (admission.kind === "ignored") return;
      if (admission.kind === "rejected") {
        setError(admission.message);
        return;
      }
      const paste = admission.value;
      const viewportContinuityToken = beginViewportContinuity(
        paste.initial.anchor === null
          ? { kind: "scroll-only" }
          : {
              kind: "input",
              focusKey: inputFocusKey(paste.rowKey, paste.bindingKey, "grid"),
            },
      );
      beginSave();
      setError(null);
      const ports: TimelinePasteExecutionPorts = {
        applyResponseRows,
        authority: () => authorityRef.current,
        clearViewportContinuity,
        clipboardPaste,
        finishSave,
        loadRows,
        registerSameFieldConflict,
        resolvePendingSocketTxn,
        resolveTarget: resolveTimelinePasteTargetResolution,
        restoreFocus: restoreTimelineFocusAnchor,
        setActiveConflictKey,
        setError,
        setPasteConflictGroup,
        trackPendingSocketTxn,
        waitForCommittedRecordIdle,
      };
      pendingSavesRefs.saveQueueRef.current =
        pendingSavesRefs.saveQueueRef.current
          .catch(() => undefined)
          .then(() =>
            executeTimelineGridPaste(paste, viewportContinuityToken, ports),
          );
    },
    [
      applyResponseRows,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      clipboardPaste,
      finishSave,
      loadRows,
      pendingSavesRefs,
      registerSameFieldConflict,
      resolvePendingSocketTxn,
      resolveTimelinePasteTargetResolution,
      restoreTimelineFocusAnchor,
      setActiveConflictKey,
      setError,
      setPasteConflictGroup,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  return {
    commands: {
      handleGridPaste,
      handlePaste,
    },
  };
}
