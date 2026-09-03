import type {
  GridCellAnchor,
  GridFillIntent,
  GridInteractionMode,
} from "@cartulary/grid-adapter";
import type { ViewContract } from "@cartulary/view-contracts";
import { useCallback } from "react";
import type { TimelineBulkMutationPort } from "../../mutations/workbookMutationCommandPorts";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import { timelineScalarBindingForField } from "../models/timelineFieldRegistry";
import {
  readTimelineCellValue,
  type WorkbookRow,
} from "../models/timelineRowModel";

export const timelineFillRejectedMessage =
  "Fill was rejected because one or more targets are unavailable or stale.";

type TimelineFillCommand = {
  readonly fieldKey: string;
  readonly sourceAnchor: GridCellAnchor;
  readonly targets: readonly {
    readonly recordId: string;
    readonly baseRowVersion: number;
  }[];
  readonly value: string;
};

type TimelineFillPlan =
  | { readonly kind: "accepted"; readonly command: TimelineFillCommand }
  | { readonly kind: "rejected"; readonly message: string };

export function planTimelineFill({
  contract,
  groupBy,
  interactionMode,
  intent,
  rows,
  visibleFieldKeys,
}: {
  readonly contract: ViewContract;
  readonly groupBy: string | null;
  readonly interactionMode: GridInteractionMode;
  readonly intent: GridFillIntent;
  readonly rows: readonly WorkbookRow[];
  readonly visibleFieldKeys: ReadonlySet<string>;
}): TimelineFillPlan {
  const fieldKey = intent.source.fieldKey;
  const sourceRecordId = coreRecordId(intent.source);
  const field = contract.fieldMap[fieldKey];
  const binding = timelineScalarBindingForField(fieldKey);
  const sourceRow = rows.find((row) => row.recordId === sourceRecordId);
  if (
    interactionMode.kind !== "editable" ||
    groupBy !== null ||
    !isTimelineSurface(intent.source, contract.viewSchemaId) ||
    sourceRecordId === null ||
    !visibleFieldKeys.has(fieldKey) ||
    field?.gridEditable !== true ||
    binding === null ||
    sourceRow?.rowVersion === null ||
    sourceRow === undefined ||
    sourceRow.rowVersion !== intent.source.mutationIdentity.baseRowVersion ||
    sourceRow.pendingSignature !== null
  ) {
    return rejectTimelineFill();
  }

  const targets = intent.targets.filter(
    (target) => coreRecordId(target) !== sourceRecordId,
  );
  const seenRecordIds = new Set<string>();
  const commandTargets: TimelineFillCommand["targets"][number][] = [];
  for (const target of targets) {
    const recordId = coreRecordId(target);
    const row = rows.find((candidate) => candidate.recordId === recordId);
    if (
      recordId === null ||
      seenRecordIds.has(recordId) ||
      !isTimelineSurface(target, contract.viewSchemaId) ||
      target.fieldKey !== fieldKey ||
      row?.rowVersion === null ||
      row === undefined ||
      row.rowVersion !== target.mutationIdentity.baseRowVersion ||
      row.pendingSignature !== null
    ) {
      return rejectTimelineFill();
    }
    seenRecordIds.add(recordId);
    commandTargets.push({
      recordId,
      baseRowVersion: target.mutationIdentity.baseRowVersion,
    });
  }
  if (commandTargets.length === 0) {
    return rejectTimelineFill();
  }

  return {
    kind: "accepted",
    command: {
      fieldKey,
      sourceAnchor: intent.source,
      targets: commandTargets,
      value: stringifyGridValue(
        readTimelineCellValue(sourceRow.rawRow, binding.fieldKey),
      ),
    },
  };
}

export function useTimelineFillController({
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  contract,
  enqueueSaveWork,
  finishSave,
  groupBy,
  getVisibleFieldKeys,
  interactionMode,
  loadRows,
  port,
  resolvePendingSocketTxn,
  restoreFocusAnchor,
  rowsRef,
  setError,
  trackPendingSocketTxn,
}: {
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (request: {
    readonly kind: "scroll-only";
  }) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly contract: ViewContract;
  readonly enqueueSaveWork: (work: () => Promise<void>) => void;
  readonly finishSave: (result: "Conflict" | "Saved" | "Syncing") => void;
  readonly getVisibleFieldKeys: () => ReadonlySet<string>;
  readonly groupBy: string | null;
  readonly interactionMode: GridInteractionMode;
  readonly loadRows: (options: {
    readonly showLoading: false;
    readonly viewportContinuityToken: number;
  }) => Promise<void>;
  readonly port: Pick<TimelineBulkMutationPort, "fillDown">;
  readonly resolvePendingSocketTxn: (
    clientTxnId: string | null | undefined,
  ) => unknown;
  readonly restoreFocusAnchor: (anchor: GridCellAnchor) => unknown;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly setError: (message: string | null) => void;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
}) {
  const onFillCells = useCallback(
    (intent: GridFillIntent) => {
      const plan = planTimelineFill({
        contract,
        groupBy,
        interactionMode,
        intent,
        rows: rowsRef.current,
        visibleFieldKeys: getVisibleFieldKeys(),
      });
      if (plan.kind === "rejected") {
        setError(plan.message);
        return;
      }

      const viewportContinuityToken = beginViewportContinuity({
        kind: "scroll-only",
      });
      beginSave();
      enqueueSaveWork(async () => {
        const result = await port.fillDown({
          fieldKey: plan.command.fieldKey,
          onClientTxnId: trackPendingSocketTxn,
          targets: plan.command.targets,
          value: plan.command.value,
        });
        resolvePendingSocketTxn(result.clientTxnId);
        if (result.outcome.kind === "rejected") {
          clearViewportContinuity(viewportContinuityToken);
          setError(result.outcome.failure.message);
          finishSave("Conflict");
          return;
        }
        await loadRows({
          showLoading: false,
          viewportContinuityToken,
        });
        restoreFocusAnchor(plan.command.sourceAnchor);
        finishSave("Saved");
      });
    },
    [
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      contract,
      enqueueSaveWork,
      finishSave,
      getVisibleFieldKeys,
      groupBy,
      interactionMode,
      loadRows,
      port,
      resolvePendingSocketTxn,
      restoreFocusAnchor,
      rowsRef,
      setError,
      trackPendingSocketTxn,
    ],
  );

  return { commands: { onFillCells } };
}

function coreRecordId(anchor: GridCellAnchor): string | null {
  return anchor.rowIdentity.kind === "core_record" &&
    anchor.rowIdentity.recordId !== ""
    ? anchor.rowIdentity.recordId
    : null;
}

function isTimelineSurface(
  anchor: GridCellAnchor,
  viewSchemaId: string,
): boolean {
  return (
    anchor.surface.kind === "view_schema" &&
    anchor.surface.viewSchemaId === viewSchemaId
  );
}

function rejectTimelineFill(): TimelineFillPlan {
  return { kind: "rejected", message: timelineFillRejectedMessage };
}
