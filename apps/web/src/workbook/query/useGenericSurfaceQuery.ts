import {
  indicatorsViewSchemaId,
  normalizeViewRowPatchV1,
  type ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useRef, useState } from "react";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import {
  requireWorkbookSurfaceAcceptance,
  type WorkbookSurfaceRecordChangeResult,
} from "../collaboration/workbookSurfacePort";
import { decisionViewId } from "../features/coordination/decisionSupersessionModel";
import type { DecisionSupersessionOwnerPort } from "../features/coordination/decisionSupersessionOperation";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import {
  initialWorkbookQueryLoadState,
  type WorkbookQueryLoadState,
} from "../models/workbookGridState";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import type { WorkbookCommittedRecordPort } from "../query/WorkbookCommittedRecordPort";
import type { WorkbookExplicitPatchOwner } from "../runtime/WorkbookExplicitPatchOwner";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "./WorkbookViewQueryPort";
import {
  abortLatestQuery,
  beginLatestQuery,
  type LatestQueryRuntime,
} from "./workbookLatestRequest";
import { applyWorkbookQueryRowPatch } from "./workbookQueryRowPatch";

export type GenericSurfaceQueryInput = {
  readonly ordinaryCreateOwner?: WorkbookCommittedRecordPort | undefined;
  readonly indicatorOwner?: WorkbookCommittedRecordPort | undefined;
  readonly explicitPatchOwner?: WorkbookExplicitPatchOwner | undefined;
  readonly decisionOwner?: DecisionSupersessionOwnerPort | undefined;
  readonly active: boolean;
  readonly contract: ViewContract;
  readonly onAuthorityUncertain: (() => void) | undefined;
  readonly queryState: WorkbookQueryState;
  readonly viewQuery: WorkbookViewQueryPort;
  readonly viewSchemaId: string;
};

export function useGenericSurfaceQuery({
  ordinaryCreateOwner,
  decisionOwner,
  indicatorOwner,
  explicitPatchOwner,
  active,
  contract,
  onAuthorityUncertain,
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

  const refresh = useCallback(
    async (options?: { readonly requireAcceptance?: boolean }) => {
      if (
        !active ||
        (ordinaryCreateOwner && !ordinaryCreateOwner.getSnapshot().authority)
      ) {
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
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
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        return;
      }
      if (result.kind === "rejected") {
        const message = result.failure.message;
        if (
          workbookFailureLifecycle(result.failure).kind ===
          "authority_unavailable"
        ) {
          onAuthorityUncertain?.();
          clearRows();
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
      const decision =
        viewSchemaId === decisionViewId
          ? decisionOwner
          : viewSchemaId === indicatorsViewSchemaId
            ? indicatorOwner
            : undefined;
      if (
        result.value.rows.some(
          (row) =>
            row.row_version <
            Math.max(
              decision?.latestVersion(row.record_id) ?? 0,
              ordinaryCreateOwner?.latestVersion(row.record_id) ?? 0,
              explicitPatchOwner?.latestVersion(row.record_id) ?? 0,
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
      const nextRows = [...result.value.rows].map((row) => {
        const accepted = decision?.acceptRow(row) ?? row;
        return explicitPatchOwner?.observeQuery(accepted) ?? accepted;
      });
      for (const row of nextRows) ordinaryCreateOwner?.acceptRow(row);
      rowsRef.current = nextRows;
      setRows(nextRows);
      acceptedRowCountRef.current = nextRows.length;
      setLoadState({ kind: "ready" });
    },
    [
      active,
      clearRows,
      contract,
      onAuthorityUncertain,
      queryState,
      viewQuery,
      viewSchemaId,
      decisionOwner,
      indicatorOwner,
      ordinaryCreateOwner,
      explicitPatchOwner,
    ],
  );

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

      const decision =
        viewSchemaId === decisionViewId
          ? decisionOwner
          : viewSchemaId === indicatorsViewSchemaId
            ? indicatorOwner
            : undefined;
      if (
        patch.rowVersion <
        Math.max(
          decision?.latestVersion(patch.recordId) ?? 0,
          ordinaryCreateOwner?.latestVersion(patch.recordId) ?? 0,
          explicitPatchOwner?.latestVersion(patch.recordId) ?? 0,
        )
      )
        return { kind: "stale" };
      const current = rowsRef.current;
      const existing = current.find((row) => row.record_id === patch.recordId);
      if (existing === undefined) return { kind: "refresh_required" };
      if (existing.row_version >= patch.rowVersion) return { kind: "stale" };
      const next = current.map((row) =>
        row.record_id === patch.recordId
          ? (decision?.acceptRow(applyWorkbookQueryRowPatch(row, patch)) ??
            applyWorkbookQueryRowPatch(row, patch))
          : row,
      );
      for (const row of next) {
        ordinaryCreateOwner?.acceptRow(row);
        explicitPatchOwner?.acceptRow(row);
      }
      rowsRef.current = next;
      setRows(next);
      return { kind: "applied" };
    },
    [
      contract,
      viewSchemaId,
      decisionOwner,
      explicitPatchOwner,
      indicatorOwner,
      ordinaryCreateOwner,
    ],
  );

  useEffect(() => {
    if (!explicitPatchOwner || !active) return;
    return explicitPatchOwner.subscribe(() => {
      if (!explicitPatchOwner.getSnapshot().authority) {
        clearRows();
        return;
      }
      const current = rowsRef.current;
      let changed = false;
      const next = current.map((row) => {
        const accepted = explicitPatchOwner.latestRow(row.record_id);
        if (accepted && accepted.row_version > row.row_version) {
          changed = true;
          return accepted;
        }
        return row;
      });
      if (changed) {
        rowsRef.current = next;
        setRows(next);
      }
    });
  }, [active, clearRows, explicitPatchOwner]);

  useEffect(() => {
    if (!indicatorOwner || !active || viewSchemaId !== indicatorsViewSchemaId)
      return;
    return indicatorOwner.subscribe(() => {
      if (!indicatorOwner.getSnapshot().authority) {
        clearRows();
        return;
      }
      let changed = false;
      const next = rowsRef.current.map((row) => {
        const accepted = indicatorOwner.latestRow(row.record_id);
        if (accepted && accepted.row_version > row.row_version) {
          changed = true;
          return accepted;
        }
        return row;
      });
      if (changed) {
        rowsRef.current = next;
        setRows(next);
      }
    });
  }, [indicatorOwner, active, viewSchemaId, clearRows]);

  useEffect(() => {
    if (!ordinaryCreateOwner || !active) return;
    let authorized = !!ordinaryCreateOwner.getSnapshot().authority;
    return ordinaryCreateOwner.subscribe(() => {
      const previouslyAuthorized = authorized;
      authorized = !!ordinaryCreateOwner.getSnapshot().authority;
      if (!authorized) {
        abortLatestQuery(queryRuntimeRef);
        clearRows();
        return;
      }
      if (!previouslyAuthorized) void refresh();
      let changed = false;
      const next = rowsRef.current.map((row) => {
        const accepted = ordinaryCreateOwner.latestRow(row.record_id);
        if (accepted && accepted.row_version > row.row_version) {
          changed = true;
          return accepted;
        }
        return row;
      });
      // Only already-admitted query members can be projected from a receipt.
      if (changed) {
        rowsRef.current = next;
        setRows(next);
      }
    });
  }, [ordinaryCreateOwner, active, clearRows, refresh]);

  const invalidate = useCallback(
    (reason: WorkbookQueryInvalidationReason) => {
      // Closure disables source writes; it does not revoke an authorized read.
      if (reason.kind === "incident_closed") return;
      abortLatestQuery(queryRuntimeRef);
      if (reason.kind === "collaboration_reset_required") return;
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
