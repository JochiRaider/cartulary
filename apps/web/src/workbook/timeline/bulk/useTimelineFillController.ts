import type {
  GridCellAnchor,
  GridFillIntent,
  GridInteractionMode,
} from "@cartulary/grid-adapter";
import type { ViewContract } from "@cartulary/view-contracts";
import { useCallback } from "react";
import type { TimelineFillMutationPort } from "../../mutations/workbookMutationCommandPorts";
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

export function useTimelineFillController(input: {
  readonly contract: ViewContract;
  readonly getVisibleFieldKeys: () => ReadonlySet<string>;
  readonly groupBy: string | null;
  readonly interactionMode: GridInteractionMode;
  readonly port: TimelineFillMutationPort;
  readonly precedingSaves: () => Promise<void>;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly setError: (message: string | null) => void;
}) {
  const onFillCells = useCallback(
    (intent: GridFillIntent) => {
      const plan = planTimelineFill({
        ...input,
        intent,
        rows: input.rowsRef.current,
        visibleFieldKeys: input.getVisibleFieldKeys(),
      });
      if (plan.kind === "rejected") {
        input.setError(plan.message);
        return;
      }
      input.setError(null);
      input.port.fillDown(
        {
          fieldKey: plan.command.fieldKey,
          targets: plan.command.targets,
          value: plan.command.value,
        },
        { delivery: intent, ready: input.precedingSaves() },
      );
    },
    [input],
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
