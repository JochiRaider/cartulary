import {
  normalizeViewRowPatchV1,
  requireViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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
import type { EntityRow } from "../models/entityWorkbookModel";
import { entityRowFromApi } from "../models/entityWorkbookModel";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import {
  initialWorkbookQueryLoadState,
  type WorkbookQueryLoadState,
} from "../models/workbookGridState";
import {
  buildQueryRequest,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { reconcileWorkbookRecordRows } from "../utils/workbookRowReconciliation";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";
import { applyWorkbookQueryRowPatch } from "./workbookQueryRowPatch";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

type ViewQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: WorkbookQueryRow[];
  };
};

export type EntitySurfaceQueryInput = {
  readonly apiBase: string | undefined;
  readonly hostQueryState: WorkbookQueryState;
  readonly identityQueryState: WorkbookQueryState;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
};

export function useEntitySurfaceQuery({
  apiBase,
  hostQueryState,
  identityQueryState,
  incidentId,
  onIncidentAccessLost,
}: EntitySurfaceQueryInput) {
  const [hostRows, setHostRows] = useState<EntityRow[]>([]);
  const [identityRows, setIdentityRows] = useState<EntityRow[]>([]);
  const [loadState, setLoadState] = useState<WorkbookQueryLoadState>(
    initialWorkbookQueryLoadState,
  );
  const acceptedRowCountRef = useRef(0);
  const hostRowsRef = useRef(hostRows);
  const identityRowsRef = useRef(identityRows);
  const queryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  hostRowsRef.current = hostRows;
  identityRowsRef.current = identityRows;

  const entityIndex = useMemo(() => {
    const index: Record<string, EntityRow> = {};
    for (const row of [...hostRows, ...identityRows]) {
      index[row.recordId] = row;
    }
    return index;
  }, [hostRows, identityRows]);

  const queryEntityView = useCallback(
    async (
      viewSchemaId: string,
      entityType: EntityRow["entityType"],
      queryState: WorkbookQueryState,
      signal: AbortSignal,
    ) => {
      const contract =
        viewSchemaId === hostsViewSchemaId ? hostsContract : identitiesContract;
      const result = await fetchWorkbookJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
        ),
        {
          method: "POST",
          signal,
          body: JSON.stringify(buildQueryRequest(contract, queryState)),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      if (envelope.data.view_schema_id !== viewSchemaId) {
        throw new Error(
          `Entity surface load returned ${envelope.data.view_schema_id} for ${viewSchemaId}.`,
        );
      }
      return normalizeWorkbookViewRows(
        contract,
        envelope.data.rows,
        `${viewSchemaId} query response`,
      ).map((row) => entityRowFromApi(row, entityType));
    },
    [apiBase, incidentId],
  );

  const refresh = useCallback(async () => {
    const request = beginLatestQuery(queryRuntimeRef);
    setLoadState(
      acceptedRowCountRef.current > 0
        ? { kind: "refreshing" }
        : initialWorkbookQueryLoadState,
    );
    try {
      const [nextHosts, nextIdentities] = await Promise.all([
        queryEntityView(
          hostsViewSchemaId,
          "host",
          hostQueryState,
          request.signal,
        ),
        queryEntityView(
          identitiesViewSchemaId,
          "identity",
          identityQueryState,
          request.signal,
        ),
      ]);
      if (!request.isCurrent()) {
        return;
      }
      setHostRows((current) => [
        ...reconcileWorkbookRecordRows(current, nextHosts),
      ]);
      setIdentityRows((current) => [
        ...reconcileWorkbookRecordRows(current, nextIdentities),
      ]);
      acceptedRowCountRef.current = nextHosts.length + nextIdentities.length;
      setLoadState({ kind: "ready" });
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Entity load failed.",
        onIncidentAccessLost,
      );
      if (workbookLoadFailureIsAccessLoss(message)) {
        hostRowsRef.current = [];
        identityRowsRef.current = [];
        acceptedRowCountRef.current = 0;
        setHostRows([]);
        setIdentityRows([]);
        setLoadState({ kind: "permission_denied", message });
      } else if (acceptedRowCountRef.current > 0) {
        setLoadState({ kind: "stale_error", message });
      } else {
        setLoadState({ kind: "unavailable", message });
      }
    }
  }, [
    hostQueryState,
    identityQueryState,
    onIncidentAccessLost,
    queryEntityView,
  ]);

  const applyRecordChanged = useCallback(
    (
      payload: RecordChangedPayload,
      viewSchemaId: string,
    ): WorkbookSurfaceRecordChangeResult => {
      if (
        viewSchemaId !== hostsViewSchemaId &&
        viewSchemaId !== identitiesViewSchemaId
      ) {
        return { kind: "refresh_required" };
      }
      const affected = payload.affected_views.find(
        (view) => view.view_schema_id === viewSchemaId,
      );
      if (
        affected?.change_kind !== "patch" ||
        affected.patch_cells === undefined
      ) {
        return { kind: "refresh_required" };
      }
      const contract =
        viewSchemaId === hostsViewSchemaId ? hostsContract : identitiesContract;
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

      const entityType =
        viewSchemaId === hostsViewSchemaId ? "host" : "identity";
      const rowsRef = entityType === "host" ? hostRowsRef : identityRowsRef;
      const current = rowsRef.current;
      const existing = current.find((row) => row.recordId === patch.recordId);
      if (existing === undefined) return { kind: "refresh_required" };
      if (existing.rowVersion >= patch.rowVersion) return { kind: "stale" };
      const next = current.map((row) =>
        row.recordId === patch.recordId
          ? entityRowFromApi(
              applyWorkbookQueryRowPatch(row.rawRow, patch),
              entityType,
            )
          : row,
      );
      rowsRef.current = next;
      if (entityType === "host") setHostRows(next);
      else setIdentityRows(next);
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
    hostRowsRef.current = [];
    identityRowsRef.current = [];
    acceptedRowCountRef.current = 0;
    setHostRows([]);
    setIdentityRows([]);
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
    entityIndex,
    hostRows,
    identityRows,
    loadState,
    refresh,
  };
}
