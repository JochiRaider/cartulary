import { useEffect, useLayoutEffect, useRef } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelineRowMutationEditorPort } from "../models/timelineControllerPorts";
import { timelineScalarBindingForField } from "../models/timelineFieldRegistry";
import {
  normalizeTimelineFullRow,
  type TimelineApiRow,
} from "../models/timelineRowModel";

function normalizeResolvedTimelineMutation(input: {
  readonly expectedRecordId: string;
  readonly row: unknown;
  readonly viewSchemaId: string;
}): Pick<WorkbookPendingMutationAccepted, "row" | "viewSchemaId"> | null {
  if (input.viewSchemaId !== timelineViewSchemaId) return null;
  try {
    const row = normalizeTimelineFullRow(
      input.row,
      "conflict resolution response row",
    );
    return row.record_id === input.expectedRecordId
      ? { row, viewSchemaId: input.viewSchemaId }
      : null;
  } catch {
    return null;
  }
}

export function useTimelineMutationRuntimeBindings({
  applyAcceptedRowMutation,
  applyAcceptedBatchRows,
  discardBlockedEdit,
  editorDraftRegistry,
  editorPort,
  loadRows,
  mutationRuntime,
}: {
  readonly applyAcceptedRowMutation: (
    rowKey: string,
    mutation: Pick<WorkbookPendingMutationAccepted, "row" | "viewSchemaId">,
  ) => unknown;
  readonly applyAcceptedBatchRows: (rows: readonly TimelineApiRow[]) => void;
  readonly discardBlockedEdit: (unitId: string) => boolean;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly editorPort: TimelineRowMutationEditorPort;
  readonly loadRows: (options: {
    readonly showLoading: boolean;
    readonly requireAcceptance?: boolean;
  }) => Promise<void>;
  readonly mutationRuntime: WorkbookMutationRuntime;
}) {
  const current = useRef({
    applyAcceptedRowMutation,
    applyAcceptedBatchRows,
    discardBlockedEdit,
    editorDraftRegistry,
    editorPort,
    loadRows,
  });
  current.current = {
    applyAcceptedRowMutation,
    applyAcceptedBatchRows,
    discardBlockedEdit,
    editorDraftRegistry,
    editorPort,
    loadRows,
  };
  useLayoutEffect(
    () =>
      mutationRuntime.entityMerge.registerTimelineRefresh(() =>
        current.current.loadRows({
          showLoading: false,
          requireAcceptance: true,
        }),
      ),
    [mutationRuntime],
  );
  useEffect(
    () =>
      mutationRuntime.registerSurface(
        timelineViewSchemaId,
        () =>
          current.current.loadRows({
            showLoading: false,
            requireAcceptance: true,
          }),
        async (mutation, conflict) => {
          const recordId = conflict.conflict.record_id;
          const binding = timelineScalarBindingForField(
            conflict.conflict.field_key,
          );
          const cleared =
            binding !== null &&
            !conflict.batchOperationId &&
            current.current.editorDraftRegistry.clearCapturedScalarField(
              recordId,
              binding.key,
              conflict.draftRevisions,
            );
          if (cleared && binding !== null) {
            current.current.editorPort.cancelEdit({
              fieldKey: binding.fieldKey,
              recordId,
            });
          }
          const outcome = normalizeResolvedTimelineMutation({
            expectedRecordId: recordId,
            row: mutation.row,
            viewSchemaId: mutation.viewSchemaId,
          });
          if (outcome !== null) {
            current.current.applyAcceptedRowMutation(recordId, outcome);
          } else {
            await current.current.loadRows({ showLoading: false });
          }
        },
        (unitId) => current.current.discardBlockedEdit(unitId),
        (rows) =>
          current.current.applyAcceptedBatchRows(
            rows.map((row) => normalizeTimelineFullRow(row, "batch receipt")),
          ),
      ),
    [mutationRuntime],
  );
}
