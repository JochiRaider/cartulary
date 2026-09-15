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
import { workbookBrowsingBounds } from "./WorkbookQueryBrowser";
import {
  useWorkbookBrowsingRead,
  useWorkbookQueryBrowser,
} from "./WorkbookQueryBrowsingContext";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "./WorkbookViewQueryPort";
import {
  abortLatestQuery,
  beginLatestQuery,
  type LatestQueryRuntime,
} from "./workbookLatestRequest";
import { applyWorkbookQueryRowPatch } from "./workbookQueryRowPatch";

export type GenericSurfaceQueryInput = {
  readonly committedRecordOwner?: WorkbookCommittedRecordPort | undefined;
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
  committedRecordOwner,
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
  const { browser, snapshot: browsing } = useWorkbookQueryBrowser(
    viewQuery,
    viewSchemaId,
    active,
  );
  const [rows, setRows] = useState<WorkbookQueryRow[]>([]);
  const [loadState, setLoadState] = useState<WorkbookQueryLoadState>(
    initialWorkbookQueryLoadState,
  );
  const hasAcceptedResultRef = useRef(false);
  const rowsRef = useRef(rows);
  const activeViewSchemaIdRef = useRef(viewSchemaId);
  const queryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  rowsRef.current = rows;

  const clearRows = useCallback(() => {
    rowsRef.current = [];
    hasAcceptedResultRef.current = false;
    setRows([]);
  }, []);

  const refresh = useCallback(
    async function refreshQuery(options?: {
      readonly requireAcceptance?: boolean;
      readonly recoveryDepth?: number;
    }) {
      if (
        !active ||
        (ordinaryCreateOwner && !ordinaryCreateOwner.getSnapshot().authority)
      ) {
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        abortLatestQuery(queryRuntimeRef);
        browser.detach();
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
        hasAcceptedResultRef.current
          ? { kind: "refreshing" }
          : { generationKey: request.generationKey, kind: "initial_loading" },
      );
      const result = await browser.query(
        {
          contract,
          queryState,
          signal: request.signal,
        },
        { recoveryDepth: options?.recoveryDepth ?? 0 },
      );
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
          browser.invalidate();
          clearRows();
          setLoadState({ kind: "permission_denied", message });
        } else if (hasAcceptedResultRef.current) {
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
              committedRecordOwner?.latestVersion(row.record_id) ?? 0,
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
        browser.reject(result.value, failure.failure);
        const recoveryDepth = browser.recoveryAttemptsUsed();
        if (recoveryDepth < workbookBrowsingBounds.recoveryAttempts)
          return refreshQuery({ ...options, recoveryDepth: recoveryDepth + 1 });
        setLoadState({ kind: "stale_error", message: failure.failure.message });
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance(failure);
        return;
      }
      const nextRows = [...result.value.rows].map((row) => {
        // Reading a page does not make every row retained mutation work.
        for (const owner of [
          decision,
          ordinaryCreateOwner,
          explicitPatchOwner,
          committedRecordOwner,
        ]) {
          if (owner?.latestRow(row.record_id)) owner.acceptRow(row);
        }
        return row;
      });
      if (!browser.accept(result.value)) return;
      rowsRef.current = nextRows;
      setRows(nextRows);
      hasAcceptedResultRef.current = true;
      setLoadState({ kind: "ready" });
    },
    [
      active,
      clearRows,
      contract,
      onAuthorityUncertain,
      queryState,
      viewSchemaId,
      decisionOwner,
      indicatorOwner,
      ordinaryCreateOwner,
      explicitPatchOwner,
      browser,
      committedRecordOwner,
    ],
  );

  useWorkbookBrowsingRead(viewSchemaId, refresh, active);
  useEffect(() => browser.observeRows(rows), [browser, rows]);
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
          committedRecordOwner?.latestVersion(patch.recordId) ?? 0,
        )
      )
        return { kind: "stale" };
      const current = rowsRef.current;
      const existing = current.find((row) => row.record_id === patch.recordId);
      if (existing === undefined) return { kind: "refresh_required" };
      if (existing.row_version >= patch.rowVersion) return { kind: "stale" };
      const next = current.map((row) =>
        row.record_id === patch.recordId
          ? applyWorkbookQueryRowPatch(row, patch)
          : row,
      );
      for (const row of next) {
        for (const owner of [
          decision,
          ordinaryCreateOwner,
          explicitPatchOwner,
          committedRecordOwner,
        ]) {
          if (owner?.latestRow(row.record_id)) owner.acceptRow(row);
        }
      }
      rowsRef.current = next;
      setRows(next);
      const placementFields = browser.getSnapshot().canonicalQuery;
      return placementFields?.filters.some(
        (filter) => filter.op === "full_text",
      ) ||
        payload.changed_field_keys.some(
          (key) =>
            placementFields?.sort.some((sort) => sort.fieldKey === key) ||
            placementFields?.filters.some(
              (filter) => filter.fieldKey === key,
            ) ||
            placementFields?.groupBy === key,
        )
        ? { kind: "refresh_required" }
        : { kind: "applied" };
    },
    [
      contract,
      viewSchemaId,
      decisionOwner,
      explicitPatchOwner,
      indicatorOwner,
      ordinaryCreateOwner,
      committedRecordOwner,
      browser,
    ],
  );

  useEffect(() => {
    if (!committedRecordOwner || !active) return;
    return committedRecordOwner.subscribe(() => {
      if (!committedRecordOwner.getSnapshot().authority) {
        browser.invalidate();
        clearRows();
        return;
      }
      const current = rowsRef.current;
      const next = current.map((row) => {
        const accepted = committedRecordOwner.latestRow(row.record_id);
        return accepted && accepted.row_version > row.row_version
          ? accepted
          : row;
      });
      if (next.some((row, index) => row !== current[index])) {
        rowsRef.current = next;
        setRows(next);
      }
    });
  }, [active, browser, clearRows, committedRecordOwner]);

  useEffect(() => {
    if (!explicitPatchOwner || !active) return;
    return explicitPatchOwner.subscribe(() => {
      if (!explicitPatchOwner.getSnapshot().authority) {
        browser.invalidate();
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
  }, [active, browser, clearRows, explicitPatchOwner]);

  useEffect(() => {
    if (!indicatorOwner || !active || viewSchemaId !== indicatorsViewSchemaId)
      return;
    return indicatorOwner.subscribe(() => {
      if (!indicatorOwner.getSnapshot().authority) {
        browser.invalidate();
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
  }, [indicatorOwner, active, viewSchemaId, clearRows, browser]);

  useEffect(() => {
    if (!ordinaryCreateOwner || !active) return;
    let authorized = !!ordinaryCreateOwner.getSnapshot().authority;
    return ordinaryCreateOwner.subscribe(() => {
      const previouslyAuthorized = authorized;
      authorized = !!ordinaryCreateOwner.getSnapshot().authority;
      if (!authorized) {
        abortLatestQuery(queryRuntimeRef);
        browser.invalidate();
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
  }, [ordinaryCreateOwner, active, browser, clearRows, refresh]);

  useEffect(() => {
    if (active) return;
    abortLatestQuery(queryRuntimeRef);
    browser.detach();
    clearRows();
    setLoadState(initialWorkbookQueryLoadState);
  }, [active, browser, clearRows]);

  const invalidate = useCallback(
    (reason: WorkbookQueryInvalidationReason) => {
      // Closure disables source writes; it does not revoke an authorized read.
      if (reason.kind === "incident_closed") return;
      abortLatestQuery(queryRuntimeRef);
      if (reason.kind === "collaboration_reset_required") return;
      browser.invalidate();
      clearRows();
    },
    [clearRows, browser],
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
    browser,
    browsing,
    acceptedQueryState: browser.presentationQuery(queryState),
  };
}
