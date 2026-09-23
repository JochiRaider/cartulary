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
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
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

const assessmentsContract = requireViewContract(assessmentsViewSchemaId);

export type AssessmentSurfaceQueryInput = {
  readonly committedRecords?: AssessmentCommittedRecordPort | undefined;
  readonly active: boolean;
  readonly onAuthorityUncertain: (() => void) | undefined;
  readonly queryState: WorkbookQueryState;
  readonly viewQuery: WorkbookViewQueryPort;
};

export function useAssessmentSurfaceQuery({
  committedRecords,
  active,
  onAuthorityUncertain,
  queryState,
  viewQuery,
}: AssessmentSurfaceQueryInput) {
  const {
    binding,
    browser,
    snapshot: browsing,
  } = useWorkbookQueryBrowser(viewQuery, assessmentsViewSchemaId, active);
  const [rows, setRows] = useState<WorkbookQueryRow[]>([]);
  const [loadState, setLoadState] = useState<WorkbookQueryLoadState>(
    initialWorkbookQueryLoadState,
  );
  const hasAcceptedResultRef = useRef(false);
  const rowsRef = useRef(rows);
  const queryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  rowsRef.current = rows;

  const refresh = useCallback(
    async function refreshQuery(options?: {
      readonly requireAcceptance?: boolean;
      readonly recoveryDepth?: number;
    }) {
      const browser = binding.currentBrowser();
      if (!browser || !active) {
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        abortLatestQuery(queryRuntimeRef);
        rowsRef.current = [];
        setRows([]);
        hasAcceptedResultRef.current = false;
        return;
      }
      const request = beginLatestQuery(queryRuntimeRef);
      setLoadState(
        hasAcceptedResultRef.current
          ? { kind: "refreshing" }
          : { generationKey: request.generationKey, kind: "initial_loading" },
      );
      const result = await browser.query(
        {
          contract: assessmentsContract,
          queryState,
          signal: request.signal,
        },
        { recoveryDepth: options?.recoveryDepth ?? 0 },
      );
      if (
        binding.currentBrowser() !== browser ||
        !request.isCurrent() ||
        result.kind === "aborted"
      ) {
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
          binding.currentBrowser()?.invalidate();
          rowsRef.current = [];
          hasAcceptedResultRef.current = false;
          setRows([]);
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
        browser.reject(result.value, failure.failure);
        const recoveryDepth = browser.recoveryAttemptsUsed();
        if (recoveryDepth < workbookBrowsingBounds.recoveryAttempts)
          return refreshQuery({ ...options, recoveryDepth: recoveryDepth + 1 });
        setLoadState({ kind: "stale_error", message: failure.failure.message });
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance(failure);
        return;
      }
      // Only the current query establishes membership. Receipts never insert filtered-out rows.
      const nextRows = result.value.rows.map((row) => {
        if (committedRecords?.latestRow(row.record_id))
          committedRecords.acceptRow(row);
        return row;
      });
      if (
        binding.currentBrowser() !== browser ||
        !browser.accept(result.value)
      ) {
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        return;
      }
      rowsRef.current = nextRows;
      setRows(nextRows);
      hasAcceptedResultRef.current = true;
      setLoadState({ kind: "ready" });
    },
    [active, binding, committedRecords, onAuthorityUncertain, queryState],
  );

  useWorkbookBrowsingRead(binding, refresh);
  useEffect(() => binding.currentBrowser()?.observeRows(rows), [binding, rows]);
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
          ? applyWorkbookQueryRowPatch(
              row,
              patch,
              viewQuery.readScope?.() ?? null,
            )
          : row,
      );
      rowsRef.current = next;
      for (const row of next)
        if (committedRecords?.latestRow(row.record_id))
          committedRecords.acceptRow(row);
      setRows(next);
      const placement = binding.currentBrowser()?.getSnapshot().canonicalQuery;
      return payload.changed_field_keys.some(
        (key) =>
          placement?.sort.some((sort) => sort.fieldKey === key) ||
          placement?.filters.some(
            (filter) => filter.fieldKey === key || filter.op === "full_text",
          ) ||
          placement?.groupBy === key,
      )
        ? { kind: "refresh_required" }
        : { kind: "applied" };
    },
    [binding, committedRecords, viewQuery.readScope],
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
      rowsRef.current = [];
      hasAcceptedResultRef.current = false;
      binding.currentBrowser()?.invalidate();
      setRows([]);
    },
    [binding],
  );

  useEffect(() => {
    if (!committedRecords) return;
    return committedRecords.subscribe(() => {
      if (!committedRecords.getSnapshot().authority) {
        abortLatestQuery(queryRuntimeRef);
        rowsRef.current = [];
        hasAcceptedResultRef.current = false;
        binding.currentBrowser()?.invalidate();
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
        setRows(next);
      }
    });
  }, [binding, committedRecords]);

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
    acceptedQueryState: browser?.presentationQuery(queryState) ?? queryState,
  };
}
