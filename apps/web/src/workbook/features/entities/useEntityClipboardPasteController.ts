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
  grouped,
  rows,
  setActionFeedback,
  setMutationError,
  viewSchemaId,
  writableFieldKeys,
}: {
  readonly canCreateRows: boolean;
  readonly clipboardPaste: WorkbookClipboardPastePort;
  readonly commitGridEdit: CommitEntityGridEdit;
  readonly grouped: boolean;
  readonly rows: readonly EntityRow[];
  readonly setActionFeedback: (value: WorkbookInspectorFeedback | null) => void;
  readonly setMutationError: (
    value: WorkbookInspectorErrorPresentation | null,
  ) => void;
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
          setActionFeedback(null);
          clipboardPaste.paste(plan.input, { delivery: intent });
      }
    },
    [
      canCreateRows,
      clipboardPaste,
      setActionFeedback,
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
