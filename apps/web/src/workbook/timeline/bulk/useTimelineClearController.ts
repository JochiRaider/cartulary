import type { GridClearIntent } from "@cartulary/grid-adapter";
import type { ViewContract } from "@cartulary/view-contracts";
import { useCallback } from "react";
import type { TimelineClearMutationPort } from "../../mutations/workbookMutationCommandPorts";
import { timelineScalarBindingForField } from "../models/timelineFieldRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";

type ClearCommand = Parameters<TimelineClearMutationPort["clearCells"]>[0];
type ClearPlan =
  | { readonly kind: "accepted"; readonly command: ClearCommand }
  | { readonly kind: "rejected"; readonly message: string };
const unavailable =
  "Clear contents requires available, editable Timeline cells. Review the selection and try again.";
const drafts =
  "Clear contents is blocked by unsaved work in a selected cell. Review or discard that draft before clearing.";

export function planTimelineClear(input: {
  readonly authorized: boolean;
  readonly contract: ViewContract;
  readonly intent: GridClearIntent;
  readonly rows: readonly WorkbookRow[];
  readonly visibleFieldKeys: ReadonlySet<string>;
  readonly hasUnsubmittedDraft: (row: WorkbookRow, fieldKey: string) => boolean;
}): ClearPlan {
  const reject = (message = unavailable): ClearPlan => ({
    kind: "rejected",
    message,
  });
  const { expandedRange, targets } = input.intent;
  const fields = expandedRange.fieldKeys;
  if (
    !input.authorized ||
    fields.length < 1 ||
    fields.length > 10 ||
    expandedRange.rowIdentities.length < 1 ||
    expandedRange.rowIdentities.length > 500 ||
    new Set(fields).size !== fields.length ||
    targets.length !== fields.length * expandedRange.rowIdentities.length
  )
    return reject();
  for (const key of fields) {
    const field = input.contract.fieldMap[key];
    if (
      !input.visibleFieldKeys.has(key) ||
      !field?.patchWritable ||
      !field.gridEditable ||
      !field.clearable ||
      field.writeKind !== "direct_value" ||
      timelineScalarBindingForField(key) === null
    )
      return reject();
  }
  const commandTargets: ClearCommand["targets"][number][] = [];
  const seen = new Set<string>();
  for (const [rowIndex, identity] of expandedRange.rowIdentities.entries()) {
    if (identity.kind !== "core_record" || seen.has(identity.recordId))
      return reject();
    seen.add(identity.recordId);
    const row = input.rows.find(
      (candidate) => candidate.recordId === identity.recordId,
    );
    if (
      !row ||
      row.rowVersion === null ||
      row.rowVersion < 1 ||
      row.captureState === "superseded"
    )
      return reject();
    for (const [columnIndex, fieldKey] of fields.entries()) {
      const target = targets[rowIndex * fields.length + columnIndex];
      if (
        !target ||
        target.surface.kind !== "view_schema" ||
        target.surface.viewSchemaId !== input.contract.viewSchemaId ||
        target.fieldKey !== fieldKey ||
        target.rowIdentity.kind !== "core_record" ||
        target.rowIdentity.recordId !== identity.recordId ||
        target.mutationIdentity.kind !== "core_row_version" ||
        target.mutationIdentity.baseRowVersion !== row.rowVersion
      )
        return reject();
      if (input.hasUnsubmittedDraft(row, fieldKey)) return reject(drafts);
    }
    commandTargets.push({
      recordId: identity.recordId,
      baseRowVersion: row.rowVersion,
    });
  }
  return {
    kind: "accepted",
    command: { fieldKeys: [...fields], targets: commandTargets },
  };
}

export function useTimelineClearController(input: {
  readonly authorized: boolean;
  readonly contract: ViewContract;
  readonly getVisibleFieldKeys: () => ReadonlySet<string>;
  readonly hasUnsubmittedDraft: (row: WorkbookRow, fieldKey: string) => boolean;
  readonly port: TimelineClearMutationPort;
  readonly precedingSaves: () => Promise<void>;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly setError: (message: string | null) => void;
}) {
  return useCallback(
    (intent: GridClearIntent | null) => {
      if (!intent) {
        input.setError(unavailable);
        return;
      }
      const plan = planTimelineClear({
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
      input.port.clearCells(plan.command, {
        delivery: intent.delivery,
        ready: input.precedingSaves(),
      });
    },
    [input],
  );
}
