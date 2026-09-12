import {
  normalizeViewRowPatchV1,
  requireViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useRef, useState } from "react";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import {
  requireWorkbookSurfaceAcceptance,
  type WorkbookSurfaceRecordChangeResult,
} from "../collaboration/workbookSurfacePort";
import type { AssessmentCommittedRecordPort } from "../features/assessments/assessmentOperation";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import {
  initialWorkbookQueryLoadState,
  type WorkbookQueryLoadState,
} from "../models/workbookGridState";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { assessmentsViewSchemaId } from "../models/workbookSurfaceRegistry";
import { workbookOperationFailureIsAccessLoss } from "../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "./WorkbookViewQueryPort";
import {
  abortLatestQuery,
  beginLatestQuery,
  type LatestQueryRuntime,
} from "./workbookLatestRequest";
import { applyWorkbookQueryRowPatch } from "./workbookQueryRowPatch";

const assessmentsContract = requireViewContract(assessmentsViewSchemaId);

export type AssessmentSurfaceQueryInput = {
  readonly committedRecords?: AssessmentCommittedRecordPort | undefined;
  readonly active: boolean;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly queryState: WorkbookQueryState;
  readonly viewQuery: WorkbookViewQueryPort;
};

export function useAssessmentSurfaceQuery({
  committedRecords,
  active,
  onIncidentAccessLost,
  queryState,
  viewQuery,
}: AssessmentSurfaceQueryInput) {
  const [rows, setRows] = useState<WorkbookQueryRow[]>([]);
  const [loadState, setLoadState] = useState<WorkbookQueryLoadState>(
    initialWorkbookQueryLoadState,
  );
  const acceptedRowCountRef = useRef(0);
  const rowsRef = useRef(rows);
  const queryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  rowsRef.current = rows;

  const refresh = useCallback(
    async (options?: { readonly requireAcceptance?: boolean }) => {
      if (!active) {
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        abortLatestQuery(queryRuntimeRef);
        return;
      }
      const request = beginLatestQuery(queryRuntimeRef);
      setLoadState(
        acceptedRowCountRef.current > 0
          ? { kind: "refreshing" }
          : { generationKey: request.generationKey, kind: "initial_loading" },
      );
      const result = await viewQuery.query({
        contract: assessmentsContract,
        queryState,
        signal: request.signal,
      });
      if (!request.isCurrent() || result.kind === "aborted") {
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        return;
      }
      if (result.kind === "rejected") {
        const message = result.failure.message;
        if (workbookOperationFailureIsAccessLoss(result.failure)) {
          onIncidentAccessLost?.();
          rowsRef.current = [];
          acceptedRowCountRef.current = 0;
          setRows([]);
          setLoadState({ kind: "permission_denied", message });
        } else if (acceptedRowCountRef.current > 0) {
          setLoadState({ kind: "stale_error", message });
        } else {
          setLoadState({ kind: "unavailable", message });
        }
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance(result);
        return;
      }
      if (
        committedRecords &&
        result.value.rows.some(
          (row) =>
            row.row_version <
              (committedRecords.latestVersion(row.record_id) ?? 0) ||
            committedRecords.wasRemoved(row.record_id, row.row_version),
        )
      ) {
        const failure = {
          kind: "rejected" as const,
          failure: {
            kind: "stale_target" as const,
            message:
              "The query is older than an accepted change. Refresh current assessments.",
          },
        };
        setLoadState({ kind: "stale_error", message: failure.failure.message });
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance(failure);
        return;
      }
      // Only the current query establishes membership. Receipts never insert filtered-out rows.
      const nextRows = result.value.rows.map(
        (row) => committedRecords?.acceptRow(row) ?? row,
      );
      rowsRef.current = nextRows;
      setRows(nextRows);
      acceptedRowCountRef.current = nextRows.length;
      setLoadState({ kind: "ready" });
    },
    [active, committedRecords, onIncidentAccessLost, queryState, viewQuery],
  );

  const applyRecordChanged = useCallback(
    (payload: RecordChangedPayload): WorkbookSurfaceRecordChangeResult => {
      const affected = payload.affected_views.find(
        (view) => view.view_schema_id === assessmentsViewSchemaId,
      );
      if (
        affected?.change_kind !== "patch" ||
        affected.patch_cells === undefined
      ) {
        return { kind: "refresh_required" };
      }
      let patch: ReturnType<typeof normalizeViewRowPatchV1>;
      try {
        patch = normalizeViewRowPatchV1(
          assessmentsContract,
          affected.patch_cells,
          "record_changed patch_cells",
        );
      } catch {
        return { kind: "refresh_required" };
      }
      if (patch.recordId !== payload.record_id) {
        return { kind: "refresh_required" };
      }
      if (
        patch.rowVersion <
          (committedRecords?.latestVersion(patch.recordId) ?? 0) ||
        committedRecords?.wasRemoved(patch.recordId, patch.rowVersion)
      )
        return { kind: "stale" };
      const current = rowsRef.current;
      const existing = current.find((row) => row.record_id === patch.recordId);
      if (existing === undefined) return { kind: "refresh_required" };
      if (existing.row_version >= patch.rowVersion) return { kind: "stale" };
      const next = current.map((row) =>
        row.record_id === patch.recordId
          ? (committedRecords?.acceptRow(
              applyWorkbookQueryRowPatch(row, patch),
            ) ?? applyWorkbookQueryRowPatch(row, patch))
          : row,
      );
      rowsRef.current = next;
      setRows(next);
      return { kind: "applied" };
    },
    [committedRecords],
  );

  const invalidate = useCallback((reason: WorkbookQueryInvalidationReason) => {
    abortLatestQuery(queryRuntimeRef);
    if (
      reason.kind === "collaboration_reset_required" ||
      reason.kind === "incident_closed"
    ) {
      return;
    }
    rowsRef.current = [];
    acceptedRowCountRef.current = 0;
    setRows([]);
  }, []);

  useEffect(() => {
    if (!committedRecords) return;
    return committedRecords.subscribe(() => {
      if (!committedRecords.getSnapshot().authority) {
        abortLatestQuery(queryRuntimeRef);
        rowsRef.current = [];
        acceptedRowCountRef.current = 0;
        setRows([]);
        return;
      }
      const current = rowsRef.current;
      const next = current
        .filter(
          (row) => !committedRecords.wasRemoved(row.record_id, row.row_version),
        )
        .map((row) => committedRecords.latestRow(row.record_id) ?? row);
      if (
        next.length !== current.length ||
        next.some((row, index) => row !== current[index])
      ) {
        rowsRef.current = next;
        acceptedRowCountRef.current = next.length;
        setRows(next);
      }
    });
  }, [committedRecords]);

  useEffect(
    () => () => {
      abortLatestQuery(queryRuntimeRef);
    },
    [],
  );

  return {
    applyRecordChanged,
    invalidate,
    loadState,
    refresh,
    rows,
  };
}
