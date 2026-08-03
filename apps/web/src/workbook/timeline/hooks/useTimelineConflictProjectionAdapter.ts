import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useMemo,
} from "react";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import {
  parseSameFieldConflict,
  workbookConflictQueueKey,
} from "../../runtime/workbookConflictModel";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelineMutableRef } from "../models/timelineControllerPorts";
import {
  type FocusFieldKey,
  inputFocusKey,
  type LocalConflictState,
  type SameFieldConflictPayload,
  type TimelineScalarEditorSurface,
  timelineScalarBindingForField,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

/**
 * Timeline adapter for the shell-owned conflict queue.
 *
 * The local record is a render projection only: registration and resolution
 * authority lives in WorkbookMutationRuntime. The projection retains Timeline
 * cell-state and draft behavior until the common queue reports that a conflict
 * has been resolved.
 */
export function useTimelineConflictProjectionAdapter({
  acceptCommittedRow,
  activeConflictKey,
  conflictQueue,
  editorDraftRegistry,
  mutationRuntime,
  rowsRef,
  setActiveConflictKey,
  setConflictQueueState,
  setRows,
}: {
  readonly acceptCommittedRow: (row: WorkbookRow) => {
    readonly accepted: boolean;
    readonly row: WorkbookRow;
    readonly stale: boolean;
  };
  readonly activeConflictKey: string | null;
  readonly conflictQueue: Record<string, LocalConflictState>;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly setActiveConflictKey: Dispatch<SetStateAction<string | null>>;
  readonly setConflictQueueState: (
    updater: (
      current: Record<string, LocalConflictState>,
    ) => Record<string, LocalConflictState>,
  ) => void;
  readonly setRows: Dispatch<SetStateAction<WorkbookRow[]>>;
}) {
  const registerSameFieldConflict = useCallback(
    (
      conflict: SameFieldConflictPayload,
      focusKey: string,
      surface: TimelineScalarEditorSurface,
      refresh?: (() => Promise<WorkbookOperationOutcome<unknown>>) | undefined,
    ) => {
      const queueKey = workbookConflictQueueKey(conflict);
      const binding = timelineScalarBindingForField(conflict.field_key);
      mutationRuntime.registerConflict({
        conflict,
        focusKey,
        refresh,
        rowLabel: conflict.record_id,
        surfaceLabel: "Timeline",
        viewSchemaId: "cartulary.view.timeline.v2",
      });
      if (binding !== null && typeof conflict.client_value === "string") {
        editorDraftRegistry.setDraft(
          { field: binding.key, rowKey: conflict.record_id, surface },
          conflict.client_value,
        );
      }
      if (binding !== null) {
        setRows((current) => {
          const nextRows = current.map((row) => {
            if (row.recordId !== conflict.record_id) return row;
            const serverText =
              typeof conflict.server_value === "string"
                ? conflict.server_value
                : "";
            return acceptCommittedRow({
              ...row,
              rowVersion: conflict.current_row_version,
              values: { ...row.values, [binding.key]: serverText },
              committedValues: {
                ...row.committedValues,
                [binding.key]: serverText,
              },
              pendingSignature: null,
            }).row;
          });
          rowsRef.current = nextRows;
          return nextRows;
        });
      }
      setConflictQueueState((current) => ({
        ...current,
        [queueKey]: {
          key: queueKey,
          anchor: {
            record_id: conflict.record_id,
            field_key: conflict.field_key,
            base_row_version: conflict.base_row_version,
            current_row_version: conflict.current_row_version,
          },
          conflict,
          focusKey,
          localValue: conflict.client_value,
          mergedDraft:
            typeof conflict.server_value === "string"
              ? conflict.server_value
              : "",
        },
      }));
      setActiveConflictKey(queueKey);
    },
    [
      acceptCommittedRow,
      editorDraftRegistry,
      mutationRuntime,
      rowsRef,
      setActiveConflictKey,
      setConflictQueueState,
      setRows,
    ],
  );

  const handleMutationConflict = useCallback(
    (
      payload: unknown,
      rowKey: string,
      focusField: FocusFieldKey,
      surface: TimelineScalarEditorSurface,
    ) => {
      const conflict = parseSameFieldConflict(payload);
      if (conflict === null) return false;
      registerSameFieldConflict(
        conflict,
        inputFocusKey(rowKey, focusField, surface),
        surface,
      );
      return true;
    },
    [registerSameFieldConflict],
  );

  const activeConflict = useMemo(
    () =>
      activeConflictKey === null
        ? null
        : (conflictQueue[activeConflictKey] ?? null),
    [activeConflictKey, conflictQueue],
  );

  return {
    commands: {
      handleMutationConflict,
      registerSameFieldConflict,
    },
    snapshot: { activeConflict },
  };
}
