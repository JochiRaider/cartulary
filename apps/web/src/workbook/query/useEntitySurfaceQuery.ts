import {
  normalizeViewRowPatchV1,
  requireViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import type { WorkbookSurfaceRecordChangeResult } from "../collaboration/workbookSurfacePort";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import type { EntityRow } from "../models/entityWorkbookModel";
import { entityRowFromApi } from "../models/entityWorkbookModel";
import {
  initialWorkbookQueryLoadState,
  type WorkbookQueryLoadState,
} from "../models/workbookGridState";
import type { WorkbookQueryState } from "../models/workbookQuery";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { reconcileWorkbookRecordRows } from "../utils/workbookRowReconciliation";
import {
  type WorkbookViewQueryPort,
  workbookViewQueryFailureIsAccessLoss,
} from "./WorkbookViewQueryPort";
import {
  abortLatestQuery,
  beginLatestQuery,
  type LatestQueryRuntime,
} from "./workbookLatestRequest";
import { applyWorkbookQueryRowPatch } from "./workbookQueryRowPatch";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

export type EntitySurfaceQueryInput = {
  readonly hostQueryState: WorkbookQueryState;
  readonly identityQueryState: WorkbookQueryState;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly viewQuery: WorkbookViewQueryPort;
};

export function useEntitySurfaceQuery({
  hostQueryState,
  identityQueryState,
  onIncidentAccessLost,
  viewQuery,
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

  const refresh = useCallback(async () => {
    const request = beginLatestQuery(queryRuntimeRef);
    setLoadState(
      acceptedRowCountRef.current > 0
        ? { kind: "refreshing" }
        : { generationKey: request.generationKey, kind: "initial_loading" },
    );
    const [hostsResult, identitiesResult] = await Promise.all([
      viewQuery.query({
        contract: hostsContract,
        queryState: hostQueryState,
        signal: request.signal,
      }),
      viewQuery.query({
        contract: identitiesContract,
        queryState: identityQueryState,
        signal: request.signal,
      }),
    ]);
    if (
      !request.isCurrent() ||
      hostsResult.kind === "aborted" ||
      identitiesResult.kind === "aborted"
    ) {
      return;
    }
    const rejected = [hostsResult, identitiesResult].find(
      (result) => result.kind === "rejected",
    );
    if (rejected?.kind === "rejected") {
      const message = rejected.failure.message;
      if (workbookViewQueryFailureIsAccessLoss(rejected.failure)) {
        onIncidentAccessLost?.();
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
      return;
    }
    if (
      hostsResult.kind !== "accepted" ||
      identitiesResult.kind !== "accepted"
    ) {
      return;
    }
    const nextHosts = hostsResult.value.rows.map((row) =>
      entityRowFromApi(row, "host"),
    );
    const nextIdentities = identitiesResult.value.rows.map((row) =>
      entityRowFromApi(row, "identity"),
    );
    setHostRows((current) => [
      ...reconcileWorkbookRecordRows(current, nextHosts),
    ]);
    setIdentityRows((current) => [
      ...reconcileWorkbookRecordRows(current, nextIdentities),
    ]);
    acceptedRowCountRef.current = nextHosts.length + nextIdentities.length;
    setLoadState({ kind: "ready" });
  }, [hostQueryState, identityQueryState, onIncidentAccessLost, viewQuery]);

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
