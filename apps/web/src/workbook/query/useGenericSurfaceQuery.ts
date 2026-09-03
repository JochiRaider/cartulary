import {
  normalizeViewRowPatchV1,
  type ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useRef, useState } from "react";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import type { WorkbookSurfaceRecordChangeResult } from "../collaboration/workbookSurfacePort";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import {
  initialWorkbookQueryLoadState,
  type WorkbookQueryLoadState,
} from "../models/workbookGridState";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { workbookOperationFailureIsAccessLoss } from "../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "./WorkbookViewQueryPort";
import {
  abortLatestQuery,
  beginLatestQuery,
  type LatestQueryRuntime,
} from "./workbookLatestRequest";
import { applyWorkbookQueryRowPatch } from "./workbookQueryRowPatch";

export type GenericSurfaceQueryInput = {
  readonly active: boolean;
  readonly contract: ViewContract;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly queryState: WorkbookQueryState;
  readonly viewQuery: WorkbookViewQueryPort;
  readonly viewSchemaId: string;
};

export function useGenericSurfaceQuery({
  active,
  contract,
  onIncidentAccessLost,
  queryState,
  viewQuery,
  viewSchemaId,
}: GenericSurfaceQueryInput) {
  const [rows, setRows] = useState<WorkbookQueryRow[]>([]);
  const [loadState, setLoadState] = useState<WorkbookQueryLoadState>(
    initialWorkbookQueryLoadState,
  );
  const acceptedRowCountRef = useRef(0);
  const rowsRef = useRef(rows);
  const activeViewSchemaIdRef = useRef(viewSchemaId);
  const queryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  rowsRef.current = rows;

  const clearRows = useCallback(() => {
    rowsRef.current = [];
    acceptedRowCountRef.current = 0;
    setRows([]);
  }, []);

  const refresh = useCallback(async () => {
    if (!active) {
      abortLatestQuery(queryRuntimeRef);
      activeViewSchemaIdRef.current = viewSchemaId;
      clearRows();
      setLoadState(initialWorkbookQueryLoadState);
      return;
    }
    const requestedViewSchemaId = viewSchemaId;
    if (activeViewSchemaIdRef.current !== requestedViewSchemaId) {
      activeViewSchemaIdRef.current = requestedViewSchemaId;
      clearRows();
    }
    const request = beginLatestQuery(queryRuntimeRef);
    setLoadState(
      acceptedRowCountRef.current > 0
        ? { kind: "refreshing" }
        : { generationKey: request.generationKey, kind: "initial_loading" },
    );
    const result = await viewQuery.query({
      contract,
      queryState,
      signal: request.signal,
    });
    if (!request.isCurrent() || result.kind === "aborted") {
      return;
    }
    if (result.kind === "rejected") {
      const message = result.failure.message;
      if (workbookOperationFailureIsAccessLoss(result.failure)) {
        onIncidentAccessLost?.();
        clearRows();
        setLoadState({ kind: "permission_denied", message });
      } else if (acceptedRowCountRef.current > 0) {
        setLoadState({ kind: "stale_error", message });
      } else {
        setLoadState({ kind: "unavailable", message });
      }
      return;
    }
    const nextRows = [...result.value.rows];
    rowsRef.current = nextRows;
    setRows(nextRows);
    acceptedRowCountRef.current = nextRows.length;
    setLoadState({ kind: "ready" });
  }, [
    active,
    clearRows,
    contract,
    onIncidentAccessLost,
    queryState,
    viewQuery,
    viewSchemaId,
  ]);

  const applyRecordChanged = useCallback(
    (payload: RecordChangedPayload): WorkbookSurfaceRecordChangeResult => {
      const affected = payload.affected_views.find(
        (view) => view.view_schema_id === viewSchemaId,
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
          contract,
          affected.patch_cells,
          "record_changed patch_cells",
        );
      } catch {
        return { kind: "refresh_required" };
      }
      if (patch.recordId !== payload.record_id) {
        return { kind: "refresh_required" };
      }

      const current = rowsRef.current;
      const existing = current.find((row) => row.record_id === patch.recordId);
      if (existing === undefined) return { kind: "refresh_required" };
      if (existing.row_version >= patch.rowVersion) return { kind: "stale" };
      const next = current.map((row) =>
        row.record_id === patch.recordId
          ? applyWorkbookQueryRowPatch(row, patch)
          : row,
      );
      rowsRef.current = next;
      setRows(next);
      return { kind: "applied" };
    },
    [contract, viewSchemaId],
  );

  const invalidate = useCallback(
    (reason: WorkbookQueryInvalidationReason) => {
      abortLatestQuery(queryRuntimeRef);
      if (
        reason.kind === "collaboration_reset_required" ||
        reason.kind === "incident_closed"
      ) {
        return;
      }
      clearRows();
    },
    [clearRows],
  );

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
