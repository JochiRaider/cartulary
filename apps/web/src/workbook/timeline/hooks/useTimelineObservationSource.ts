import { useCallback, useRef } from "react";
import {
  type ObservationSource,
  observationSourceFields,
  sameObservationSource,
} from "../../features/indicators/observationModel";
import type { ObservationSourcePort } from "../../features/indicators/observationOperation";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelineCommittedRecordIdleResult } from "../models/timelineControllerPorts";
import {
  timelineCollectionBindings,
  timelineScalarBindings,
} from "../models/timelineFieldRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";

export function useTimelineObservationSource(options: {
  runtime: WorkbookMutationRuntime;
  selectedRow: WorkbookRow | null;
  available: boolean;
  rowsRef: { readonly current: WorkbookRow[] };
  drafts: TimelineEditorDraftRegistry;
  waitForIdle: (
    id: string,
    options: { signal: AbortSignal; refreshIfMissing: boolean },
  ) => Promise<TimelineCommittedRecordIdleResult | null>;
}): ObservationSourcePort {
  const rowKey = options.available ? options.selectedRow?.key : undefined,
    drafts = options.drafts;
  const subscribe = useCallback(
    (listener: () => void) =>
      rowKey ? drafts.subscribeRow(rowKey, listener) : () => {},
    [drafts, rowKey],
  );
  const current = useRef(options);
  current.current = options;
  const recordId = options.selectedRow?.recordId;
  const view = timelineViewSchemaId;
  const live = () =>
    current.current.available &&
    current.current.selectedRow?.recordId === recordId
      ? (current.current.rowsRef.current.find(
          (row) => row.recordId === recordId,
        ) ?? null)
      : null;
  const source = (fieldKey: string): ObservationSource | null => {
    const row = live(),
      raw = row?.rawRow,
      text = raw?.cells[fieldKey]?.value;
    if (
      !raw ||
      typeof text !== "string" ||
      !observationSourceFields(view, raw).some(
        (field) => field.fieldKey === fieldKey,
      )
    )
      return null;
    return {
      incidentId: current.current.runtime.scope.incidentId,
      viewSchemaId: view,
      recordId: raw.record_id,
      rowVersion: raw.row_version,
      fieldKey,
      text,
    };
  };
  const ready = () => {
    const row = live();
    if (!row?.rawRow || !row.recordId) return false;
    const materialized = current.current.drafts.materializeRow(row);
    return (
      !Object.entries(materialized.values).some(
        ([key, value]) =>
          value !==
          row.committedValues[key as keyof typeof row.committedValues],
      ) &&
      !timelineScalarBindings.some((binding) => {
        const draft = current.current.drafts.draftValue({
          rowKey: row.key,
          field: binding.key,
          surface: "inspector",
        });
        return (
          draft !== undefined && draft !== row.committedValues[binding.key]
        );
      }) &&
      !timelineCollectionBindings.some(
        (binding) =>
          (
            current.current.drafts.draftValue({
              rowKey: row.key,
              field: binding.draftKey,
              surface: "inspector",
            }) ?? ""
          ).trim() !== "",
      ) &&
      !Object.values(materialized.collectionDrafts).some(
        (value) => value.trim() !== "",
      ) &&
      (current.current.runtime.history.latestVersion(row.recordId) ?? 0) <=
        row.rawRow.row_version &&
      !current.current.runtime.history
        .getSnapshot()
        .some(
          (entry) =>
            entry.attempt.subject.recordId === row.recordId &&
            ["preparing", "submitting", "uncertain"].includes(entry.phase),
        )
    );
  };
  return {
    subscribe,
    fields:
      options.available && options.selectedRow?.rawRow
        ? observationSourceFields(view, options.selectedRow.rawRow)
        : [],
    source,
    ready,
    async prepare(expected, signal) {
      if (
        !ready() ||
        !sameObservationSource(source(expected.fieldKey), expected)
      )
        return false;
      const committed = await current.current.waitForIdle(expected.recordId, {
        signal,
        refreshIfMissing: false,
      });
      if (signal.aborted || !committed?.row?.rawRow || !ready()) return false;
      return (
        committed.rowVersion === expected.rowVersion &&
        committed.row.rawRow.row_version === expected.rowVersion &&
        committed.row.rawRow.cells[expected.fieldKey]?.value ===
          expected.text &&
        sameObservationSource(source(expected.fieldKey), expected)
      );
    },
  };
}
