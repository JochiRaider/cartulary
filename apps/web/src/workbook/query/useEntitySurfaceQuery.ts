import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import {
  requireWorkbookSurfaceAcceptance,
  type WorkbookSurfaceRecordChangeResult,
} from "../collaboration/workbookSurfacePort";
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
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import { reconcileWorkbookRecordRows } from "../utils/workbookRowReconciliation";
import { planEntityLiveEventPatch } from "./entityLiveEventPatchPlanner";
import type { WorkbookCommittedRecordPort } from "./WorkbookCommittedRecordPort";
import type { WorkbookViewQueryPort } from "./WorkbookViewQueryPort";
import {
  abortLatestQuery,
  beginLatestQuery,
  type LatestQueryRuntime,
} from "./workbookLatestRequest";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

export type EntitySurfaceQueryInput = {
  readonly ordinaryCreateOwner?: WorkbookCommittedRecordPort | undefined;
  readonly editOwner?: WorkbookCommittedRecordPort | undefined;
  readonly hostQueryState: WorkbookQueryState;
  readonly identityQueryState: WorkbookQueryState;
  readonly onAuthorityUncertain: (() => void) | undefined;
  readonly viewQuery: WorkbookViewQueryPort;
};

export function useEntitySurfaceQuery({
  ordinaryCreateOwner,
  editOwner,
  hostQueryState,
  identityQueryState,
  onAuthorityUncertain,
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

  const refresh = useCallback(
    async (options?: { readonly requireAcceptance?: boolean }) => {
      if (ordinaryCreateOwner && !ordinaryCreateOwner.getSnapshot().authority) {
        abortLatestQuery(queryRuntimeRef);
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        return;
      }
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
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        return;
      }
      const rejected = [hostsResult, identitiesResult].find(
        (result) => result.kind === "rejected",
      );
      if (rejected?.kind === "rejected") {
        const message = rejected.failure.message;
        if (
          workbookFailureLifecycle(rejected.failure).kind ===
          "authority_unavailable"
        ) {
          onAuthorityUncertain?.();
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
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance(rejected);
        return;
      }
      if (
        hostsResult.kind !== "accepted" ||
        identitiesResult.kind !== "accepted"
      ) {
        return;
      }
      if (
        [...hostsResult.value.rows, ...identitiesResult.value.rows].some(
          (row) =>
            row.row_version <
            Math.max(
              ordinaryCreateOwner?.latestVersion(row.record_id) ?? 0,
              editOwner?.latestVersion(row.record_id) ?? 0,
            ),
        )
      ) {
        const failure = {
          kind: "rejected" as const,
          failure: {
            kind: "stale_target" as const,
            message:
              "The query is older than an accepted change. Refresh current rows.",
          },
        };
        setLoadState({ kind: "stale_error", message: failure.failure.message });
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance(failure);
        return;
      }
      for (const row of [
        ...hostsResult.value.rows,
        ...identitiesResult.value.rows,
      ]) {
        ordinaryCreateOwner?.acceptRow(row);
        editOwner?.acceptRow(row);
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
    },
    [
      hostQueryState,
      identityQueryState,
      onAuthorityUncertain,
      viewQuery,
      ordinaryCreateOwner,
      editOwner,
    ],
  );

  const applyRecordChanged = useCallback(
    (
      payload: RecordChangedPayload,
      viewSchemaId: string,
    ): WorkbookSurfaceRecordChangeResult => {
      if (
        payload.row_version <
        Math.max(
          ordinaryCreateOwner?.latestVersion(payload.record_id) ?? 0,
          editOwner?.latestVersion(payload.record_id) ?? 0,
        )
      )
        return { kind: "stale" };
      const plan = planEntityLiveEventPatch({
        hostRows: hostRowsRef.current,
        identityRows: identityRowsRef.current,
        payload,
        viewSchemaId,
      });
      if (plan.kind === "refresh_required") return plan;
      if (plan.kind === "stale_noop") return { kind: "stale" };
      const next = [...plan.rows];
      if (plan.entityType === "host") {
        hostRowsRef.current = next;
        setHostRows(next);
      } else {
        identityRowsRef.current = next;
        setIdentityRows(next);
      }
      return { kind: "applied" };
    },
    [ordinaryCreateOwner, editOwner],
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

  useEffect(() => {
    const subscribe = (owner: WorkbookCommittedRecordPort) => {
      let authorized = !!owner.getSnapshot().authority;
      return owner.subscribe(() => {
        const previouslyAuthorized = authorized;
        authorized = !!owner.getSnapshot().authority;
        if (!authorized) {
          abortLatestQuery(queryRuntimeRef);
          hostRowsRef.current = [];
          identityRowsRef.current = [];
          acceptedRowCountRef.current = 0;
          setHostRows([]);
          setIdentityRows([]);
          return;
        }
        if (!previouslyAuthorized) void refresh();
        const project = (rows: EntityRow[], type: "host" | "identity") =>
          rows.map((row) => {
            const accepted = owner.latestRow(row.recordId);
            return accepted && accepted.row_version > row.rowVersion
              ? entityRowFromApi(accepted, type)
              : row;
          });
        const hosts = project(hostRowsRef.current, "host"),
          identities = project(identityRowsRef.current, "identity");
        if (hosts.some((row, index) => row !== hostRowsRef.current[index])) {
          hostRowsRef.current = hosts;
          setHostRows(hosts);
        }
        if (
          identities.some(
            (row, index) => row !== identityRowsRef.current[index],
          )
        ) {
          identityRowsRef.current = identities;
          setIdentityRows(identities);
        }
      });
    };
    const unsubscribe = [ordinaryCreateOwner, editOwner].flatMap((owner) =>
      owner ? [subscribe(owner)] : [],
    );
    return () => {
      for (const stop of unsubscribe) stop();
    };
  }, [ordinaryCreateOwner, editOwner, refresh]);

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
