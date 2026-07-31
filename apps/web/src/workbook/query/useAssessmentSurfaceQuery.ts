import {
  normalizeViewRowPatchV1,
  requireViewContract,
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
import {
  initialWorkbookQueryLoadState,
  type WorkbookQueryLoadState,
} from "../models/workbookGridState";
import {
  buildQueryRequest,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { assessmentsViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";
import { applyWorkbookQueryRowPatch } from "./workbookQueryRowPatch";

const assessmentsContract = requireViewContract(assessmentsViewSchemaId);

type ViewQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: WorkbookQueryRow[];
  };
};

export type AssessmentSurfaceQueryInput = {
  readonly active: boolean;
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly queryState: WorkbookQueryState;
};

export function useAssessmentSurfaceQuery({
  active,
  apiBase,
  incidentId,
  onIncidentAccessLost,
  queryState,
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

  const refresh = useCallback(async () => {
    if (!active) {
      abortLatestQuery(queryRuntimeRef);
      return;
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
          `/api/v1/incidents/${incidentId}/views/${assessmentsViewSchemaId}/query`,
        ),
        {
          method: "POST",
          signal: request.signal,
          body: JSON.stringify(
            buildQueryRequest(assessmentsContract, queryState),
          ),
        },
      );
      if (!request.isCurrent()) {
        return;
      }
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      if (envelope.data.view_schema_id !== assessmentsViewSchemaId) {
        throw new Error(
          `Assessment load returned ${envelope.data.view_schema_id} for ${assessmentsViewSchemaId}.`,
        );
      }
      rowsRef.current = envelope.data.rows;
      setRows(envelope.data.rows);
      acceptedRowCountRef.current = envelope.data.rows.length;
      setLoadState({ kind: "ready" });
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Assessment load failed.",
        onIncidentAccessLost,
      );
      if (workbookLoadFailureIsAccessLoss(message)) {
        rowsRef.current = [];
        acceptedRowCountRef.current = 0;
        setRows([]);
        setLoadState({ kind: "permission_denied", message });
      } else if (acceptedRowCountRef.current > 0) {
        setLoadState({ kind: "stale_error", message });
      } else {
        setLoadState({ kind: "unavailable", message });
      }
    }
  }, [active, apiBase, incidentId, onIncidentAccessLost, queryState]);

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
    [],
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
