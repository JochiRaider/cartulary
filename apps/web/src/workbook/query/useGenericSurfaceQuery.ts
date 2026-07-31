import {
  normalizeViewRowPatchV1,
  type ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useRef, useState } from "react";
import { apiPath } from "../../services/browserApi";
import {
  abortLatestQuery,
  beginLatestQuery,
  fetchWorkbookJSON,
  handleWorkbookLoadFailure,
  isAbortError,
  type LatestQueryRuntime,
  parseErrorMessage,
  readEnvelope,
  workbookLoadFailureIsAccessLoss,
} from "../../services/workbookApi";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import type { WorkbookSurfaceRecordChangeResult } from "../collaboration/workbookSurfacePort";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import {
  initialWorkbookQueryLoadState,
  type WorkbookQueryLoadState,
} from "../models/workbookGridState";
import {
  buildQueryRequest,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { notesViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";
import { applyWorkbookQueryRowPatch } from "./workbookQueryRowPatch";

type ViewQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: WorkbookQueryRow[];
  };
};

export type GenericSurfaceQueryInput = {
  readonly active: boolean;
  readonly apiBase: string | undefined;
  readonly contract: ViewContract;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly queryState: WorkbookQueryState;
  readonly viewSchemaId: string;
};

export function useGenericSurfaceQuery({
  active,
  apiBase,
  contract,
  incidentId,
  onIncidentAccessLost,
  queryState,
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
        : initialWorkbookQueryLoadState,
    );
    try {
      const result = await fetchWorkbookJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${requestedViewSchemaId}/query`,
        ),
        {
          method: "POST",
          signal: request.signal,
          body: JSON.stringify(buildQueryRequest(contract, queryState)),
        },
      );
      if (!request.isCurrent()) {
        return;
      }
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      if (envelope.data.view_schema_id !== requestedViewSchemaId) {
        throw new Error(
          `Surface load returned ${envelope.data.view_schema_id} for ${requestedViewSchemaId}.`,
        );
      }
      const nextRows =
        requestedViewSchemaId === notesViewSchemaId
          ? normalizeWorkbookViewRows(
              contract,
              envelope.data.rows,
              `${requestedViewSchemaId} query response`,
            )
          : envelope.data.rows;
      rowsRef.current = nextRows;
      setRows(nextRows);
      acceptedRowCountRef.current = nextRows.length;
      setLoadState({ kind: "ready" });
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Surface load failed.",
        onIncidentAccessLost,
      );
      if (workbookLoadFailureIsAccessLoss(message)) {
        clearRows();
        setLoadState({ kind: "permission_denied", message });
      } else if (acceptedRowCountRef.current > 0) {
        setLoadState({ kind: "stale_error", message });
      } else {
        setLoadState({ kind: "unavailable", message });
      }
    }
  }, [
    active,
    apiBase,
    clearRows,
    contract,
    incidentId,
    onIncidentAccessLost,
    queryState,
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
