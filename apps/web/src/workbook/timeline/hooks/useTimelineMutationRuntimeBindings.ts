import { useEffect } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelineRowMutationEditorPort } from "../models/timelineControllerPorts";
import {
  normalizeTimelineFullRow,
  timelineScalarBindingForField,
} from "../models/workbookTimelineModel";

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
  }) => Promise<void>;
  readonly mutationRuntime: WorkbookMutationRuntime;
}) {
  useEffect(
    () =>
      mutationRuntime.registerSurface(
        timelineViewSchemaId,
        () => loadRows({ showLoading: false }),
        async (mutation, conflict) => {
          const recordId = conflict.conflict.record_id;
          const binding = timelineScalarBindingForField(
            conflict.conflict.field_key,
          );
          if (binding !== null) {
            editorDraftRegistry.clearScalarDraftsForField(
              recordId,
              binding.key,
            );
            editorPort.cancelEdit({
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
            applyAcceptedRowMutation(recordId, outcome);
          } else {
            await loadRows({ showLoading: false });
          }
          if (binding !== null) {
            window.setTimeout(() => {
              editorPort.focus({ fieldKey: binding.fieldKey, recordId });
            }, 0);
          }
        },
        (conflict) => {
          window.setTimeout(() => {
            const existingEditor =
              conflict.focusKey === null
                ? null
                : editorDraftRegistry.inputElementForFocusKey(
                    conflict.focusKey,
                  );
            if (existingEditor !== null) {
              existingEditor.focus({ preventScroll: true });
              return;
            }
            editorPort.activateEdit({
              fieldKey: conflict.conflict.field_key,
              recordId: conflict.conflict.record_id,
              value: conflict.localValue,
            });
          }, 0);
        },
        discardBlockedEdit,
      ),
    [
      applyAcceptedRowMutation,
      discardBlockedEdit,
      editorDraftRegistry,
      editorPort,
      loadRows,
      mutationRuntime,
    ],
  );
}
