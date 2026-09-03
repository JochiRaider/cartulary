import type {
  GridCellPasteIntent,
  GridEditCommitOutcome,
} from "@cartulary/grid-adapter";
import { useCallback } from "react";
import type { WorkbookClipboardPastePort } from "../../adapters/WorkbookClipboardPastePort";
import {
  type WorkbookInspectorErrorPresentation,
  type WorkbookInspectorFeedback,
  workbookInspectorLocalErrorPresentation,
  workbookInspectorMessageFeedback,
  workbookInspectorOperationFailureFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  type EntityClipboardPastePlan,
  entityClipboardPastePlan,
} from "../../models/entityClipboardPastePlan";
import type { EntityRow } from "../../models/entityWorkbookModel";

type CommitEntityGridEdit = (
  fieldKey: string,
  value: string,
  target: Extract<
    EntityClipboardPastePlan,
    { readonly kind: "scalar" }
  >["target"],
) => Promise<GridEditCommitOutcome>;

export function useEntityClipboardPasteController({
  canCreateRows,
  clipboardPaste,
  commitGridEdit,
  entityType,
  grouped,
  onRefreshEntities,
  rows,
  setActionFeedback,
  setMutationError,
  setSelectedRecordId,
  viewSchemaId,
  writableFieldKeys,
}: {
  readonly canCreateRows: boolean;
  readonly clipboardPaste: WorkbookClipboardPastePort;
  readonly commitGridEdit: CommitEntityGridEdit;
  readonly entityType: EntityRow["entityType"];
  readonly grouped: boolean;
  readonly onRefreshEntities: () => Promise<void>;
  readonly rows: readonly EntityRow[];
  readonly setActionFeedback: (value: WorkbookInspectorFeedback | null) => void;
  readonly setMutationError: (
    value: WorkbookInspectorErrorPresentation | null,
  ) => void;
  readonly setSelectedRecordId: (recordId: string) => void;
  readonly viewSchemaId: string;
  readonly writableFieldKeys: ReadonlySet<string>;
}) {
  const rejectLocally = useCallback(
    (message: string) => {
      setMutationError(workbookInspectorLocalErrorPresentation(message));
    },
    [setMutationError],
  );

  const executeScalarPlan = useCallback(
    async (plan: Extract<EntityClipboardPastePlan, { kind: "scalar" }>) => {
      const outcome = await commitGridEdit(
        plan.fieldKey,
        plan.value,
        plan.target,
      );
      if (outcome.kind !== "accepted") {
        rejectLocally(outcome.message ?? "Paste could not be applied.");
      }
    },
    [commitGridEdit, rejectLocally],
  );

  const executeBatchPlan = useCallback(
    async (plan: Extract<EntityClipboardPastePlan, { kind: "batch" }>) => {
      setActionFeedback(null);
      const { outcome } = await clipboardPaste.paste(plan.input);
      if (outcome.kind === "rejected") {
        setActionFeedback(
          workbookInspectorOperationFailureFeedback(outcome.failure),
        );
        return;
      }
      const firstRow = outcome.value.rows[0];
      await onRefreshEntities();
      if (firstRow !== undefined) setSelectedRecordId(firstRow.record_id);
      const count = outcome.value.rows.length;
      setActionFeedback(
        workbookInspectorMessageFeedback(
          `Paste applied to ${count} ${entityType === "host" ? "host" : "identity"} row${count === 1 ? "" : "s"}.`,
          "none",
        ),
      );
    },
    [
      clipboardPaste,
      entityType,
      onRefreshEntities,
      setActionFeedback,
      setSelectedRecordId,
    ],
  );

  const handlePaste = useCallback(
    (intent: GridCellPasteIntent) => {
      const plan = entityClipboardPastePlan(intent, {
        canCreateRows,
        grouped,
        rows,
        viewSchemaId,
        writableFieldKeys,
      });
      switch (plan.kind) {
        case "rejected":
          rejectLocally(plan.message);
          return;
        case "scalar":
          void executeScalarPlan(plan);
          return;
        case "batch":
          void executeBatchPlan(plan);
      }
    },
    [
      canCreateRows,
      executeBatchPlan,
      executeScalarPlan,
      grouped,
      rejectLocally,
      rows,
      viewSchemaId,
      writableFieldKeys,
    ],
  );

  return { handlePaste };
}
