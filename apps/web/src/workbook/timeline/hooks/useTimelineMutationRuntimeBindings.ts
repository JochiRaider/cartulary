import { useEffect, useLayoutEffect, useRef } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelineRowMutationEditorPort } from "../models/timelineControllerPorts";
import { timelineScalarBindingForField } from "../models/timelineFieldRegistry";
import { normalizeTimelineFullRow } from "../models/timelineRowModel";

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
    discardBlockedEdit,
    editorDraftRegistry,
    editorPort,
    loadRows,
  });
  current.current = {
    applyAcceptedRowMutation,
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
          if (binding !== null && !conflict.batchOperationId) {
            current.current.editorDraftRegistry.clearScalarDraftsForField(
              recordId,
              binding.key,
            );
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
          if (binding !== null && !conflict.batchOperationId) {
            window.setTimeout(() => {
              current.current.editorPort.focus({
                fieldKey: binding.fieldKey,
                recordId,
              });
            }, 0);
          }
        },
        (conflict) => {
          window.setTimeout(() => {
            const existingEditor =
              conflict.focusKey === null
                ? null
                : current.current.editorDraftRegistry.inputElementForFocusKey(
                    conflict.focusKey,
                  );
            if (existingEditor !== null) {
              existingEditor.focus({ preventScroll: true });
              return;
            }
            current.current.editorPort.activateEdit({
              fieldKey: conflict.conflict.field_key,
              recordId: conflict.conflict.record_id,
              value: conflict.localValue,
            });
          }, 0);
        },
        (unitId) => current.current.discardBlockedEdit(unitId),
        (rows) => {
          for (const row of rows)
            current.current.applyAcceptedRowMutation(row.record_id, {
              row: normalizeTimelineFullRow(row, "batch receipt"),
              viewSchemaId: timelineViewSchemaId,
            });
        },
      ),
    [mutationRuntime],
  );
}
