import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useMemo,
} from "react";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import {
  parseSameFieldConflict,
  workbookConflictQueueKey,
} from "../../runtime/workbookConflictModel";
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
  activeConflictKey,
  conflictQueue,
  mutationRuntime,
  rowsRef,
  scalarDraftValuesRef,
  setActiveConflictKey,
  setConflictQueueState,
  setRows,
}: {
  readonly activeConflictKey: string | null;
  readonly conflictQueue: Record<string, LocalConflictState>;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly scalarDraftValuesRef: TimelineMutableRef<Map<string, string>>;
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
    ) => {
      const queueKey = workbookConflictQueueKey(conflict);
      const binding = timelineScalarBindingForField(conflict.field_key);
      mutationRuntime.registerConflict({
        conflict,
        focusKey,
        rowLabel: conflict.record_id,
        surfaceLabel: "Timeline",
        viewSchemaId: "cartulary.view.timeline.v2",
      });
      if (binding !== null && typeof conflict.client_value === "string") {
        scalarDraftValuesRef.current.set(
          inputFocusKey(conflict.record_id, binding.key, surface),
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
            return {
              ...row,
              rowVersion: conflict.current_row_version,
              values: { ...row.values, [binding.key]: serverText },
              committedValues: {
                ...row.committedValues,
                [binding.key]: serverText,
              },
              pendingSignature: null,
            };
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
      mutationRuntime,
      rowsRef,
      scalarDraftValuesRef,
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
