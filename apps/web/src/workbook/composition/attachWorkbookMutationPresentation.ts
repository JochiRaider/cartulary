import type { AuthorizationRecoveryResult } from "../../shared/authorizationRecovery";
import { decisionViewId } from "../features/coordination/decisionSupersessionModel";
import { reconcileDecisionReceipt } from "../features/coordination/reconcileDecisionReceipt";
import { indicatorLifecycleViewId } from "../features/indicators/indicatorLifecycleModel";
import { reconcileIndicatorCreateReceipt } from "../features/indicators/reconcileIndicatorCreateReceipt";
import { reconcileIndicatorLifecycleReceipt } from "../features/indicators/reconcileIndicatorLifecycleReceipt";
import { reconcileObservationReceipt } from "../features/indicators/reconcileObservationReceipt";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { reconcileTimelineCaptureReceipt } from "../timeline/actions/reconcileTimelineCaptureReceipt";
import { reconcileTimelineMentionReceipt } from "../timeline/actions/reconcileTimelineMentionReceipt";
import { timelineCaptureOwnerFor } from "../timeline/actions/timelineCaptureOwnerFor";
import { timelineMentionOwnerFor } from "../timeline/actions/timelineMentionOwnerFor";
import { createTimelineMentionSourceReader } from "../timeline/adapters/createTimelineMentionSourceReader";

type AcceptedAuthorization = Extract<
  AuthorizationRecoveryResult,
  { kind: "authorized" }
>;
type Refresh = (options: { readonly requireAcceptance: true }) => Promise<void>;

export type WorkbookMutationPresentation = {
  readonly runtime: WorkbookMutationRuntime;
  readonly apiBase: string | undefined;
  readonly actorId: string | null;
  readonly sessionIdentity: string | null;
  readonly surface: string;
  readonly extensionWorkspace: boolean;
  readonly refresh: {
    readonly entities: Refresh;
    readonly assessment: Refresh;
    readonly generic: Refresh;
  };
  readonly authorityUncertain: () => void;
  readonly authorizationRecovered: (result: AcceptedAuthorization) => void;
};

/** One committed presentation lease; retained attempts and dispatch never depend on it. */
export function attachWorkbookMutationPresentation(
  input: WorkbookMutationPresentation,
): () => void {
  const { runtime, surface, refresh } = input;
  const lease = runtime.attachPresentationCallbacks(input);
  const assertCurrent = () => {
    if (!lease.isCurrent()) throw new Error("Workbook presentation detached");
  };
  const currentScope = <T extends { isCurrent(): boolean }>(scope: T): T => ({
    ...scope,
    isCurrent: () => lease.isCurrent() && scope.isCurrent(),
  });
  const capture = timelineCaptureOwnerFor(runtime);
  const mentions = timelineMentionOwnerFor(runtime);
  const cleanups = [
    runtime.history.registerRelatedProjectionRefresh(async (receipt) => {
      assertCurrent();
      const origin = runtime.history
        .getSnapshot()
        .find((entry) => entry.receipt === receipt)?.attempt
        .subject.viewSchemaId;
      if (origin !== hostsViewSchemaId && origin !== identitiesViewSchemaId)
        await refresh.entities({ requireAcceptance: true });
      assertCurrent();
    }),
    runtime.entityMerge.registerProjectionRefresh(async () => {
      assertCurrent();
      const owner = runtime.entityMerge;
      const scope = owner.getSnapshot();
      if (
        scope.authority?.actorId !== input.actorId ||
        scope.authority?.sessionIdentity !== input.sessionIdentity
      )
        throw new Error("Merge projection authority changed");
      await refresh.entities({ requireAcceptance: true });
      assertCurrent();
      if (owner.getSnapshot().generation !== scope.generation)
        throw new Error("Merge projection scope changed");
      if (surface === timelineViewSchemaId) await owner.refreshTimeline();
      else if (surface === assessmentsViewSchemaId)
        await refresh.assessment({ requireAcceptance: true });
      else if (
        !input.extensionWorkspace &&
        surface !== hostsViewSchemaId &&
        surface !== identitiesViewSchemaId
      )
        await refresh.generic({ requireAcceptance: true });
      assertCurrent();
    }),
    runtime.decisionSupersession.registerReconciliation(
      async (receipt, originalScope) => {
        assertCurrent();
        const scope = currentScope(originalScope);
        await reconcileDecisionReceipt(
          runtime.decisionSupersession,
          runtime.history,
          receipt,
          scope,
        );
        if (!scope.isCurrent())
          throw new Error("Decision reconciliation detached");
        if (surface === decisionViewId)
          await runtime.history.refreshSurface(decisionViewId);
        assertCurrent();
      },
    ),
    runtime.indicatorLifecycle.registerReconciliation(
      async (receipt, originalScope) => {
        assertCurrent();
        const scope = currentScope(originalScope);
        await reconcileIndicatorLifecycleReceipt(
          runtime.indicatorLifecycle,
          runtime.history,
          receipt,
          scope,
        );
        if (!scope.isCurrent())
          throw new Error("Indicator reconciliation detached");
        if (surface === indicatorLifecycleViewId)
          await runtime.history.refreshSurface(indicatorLifecycleViewId);
        assertCurrent();
      },
    ),
    runtime.indicatorCreate.registerReconciliation(
      async (_attempt, receipt, originalScope) => {
        assertCurrent();
        const scope = currentScope(originalScope);
        await reconcileIndicatorCreateReceipt(
          runtime.indicatorRecords,
          runtime.indicatorObservations,
          runtime.history,
          runtime.scope.incidentId,
          receipt,
          scope,
          () => runtime.indicatorCreate.suspendForAuthorityRecovery(),
        );
        if (!scope.isCurrent())
          throw new Error("Canonical reconciliation detached");
        if (surface === indicatorLifecycleViewId)
          await runtime.history.refreshSurface(indicatorLifecycleViewId);
        assertCurrent();
      },
    ),
    runtime.indicatorObservations.registerReconciliation(
      async (attempt, receipt, originalScope) => {
        assertCurrent();
        const scope = currentScope(originalScope);
        await reconcileObservationReceipt(
          runtime.indicatorObservations,
          runtime.history,
          attempt,
          receipt,
          scope,
        );
        if (!scope.isCurrent())
          throw new Error("Observation reconciliation detached");
        if (
          surface === indicatorLifecycleViewId ||
          surface === timelineViewSchemaId
        )
          await runtime.history.refreshSurface(surface);
        assertCurrent();
      },
    ),
    capture.registerReconciliation(async (receipt, originalScope) => {
      assertCurrent();
      const scope = currentScope(originalScope);
      await reconcileTimelineCaptureReceipt(
        capture,
        runtime.history,
        receipt,
        scope,
      );
      if (!scope.isCurrent())
        throw new Error("Timeline reconciliation detached");
      if (surface === timelineViewSchemaId)
        await runtime.history.refreshSurface(timelineViewSchemaId);
      assertCurrent();
    }),
    mentions.registerReconciliation(async (receipt, originalScope) => {
      assertCurrent();
      const scope = currentScope(originalScope);
      await reconcileTimelineMentionReceipt(
        mentions,
        createTimelineMentionSourceReader({
          apiBase: input.apiBase,
          incidentId: runtime.scope.incidentId,
          readScope: () => runtime.recordReadScope,
        }),
        receipt,
        scope,
      );
      if (!scope.isCurrent())
        throw new Error("Mention reconciliation detached");
    }),
    mentions.registerCreationReconciliation(async (scope) => {
      assertCurrent();
      if (!scope.isCurrent()) throw new Error("Entity refresh detached");
      await refresh.entities({ requireAcceptance: true });
      assertCurrent();
      if (!scope.isCurrent()) throw new Error("Entity refresh detached");
    }),
  ];
  return () => {
    lease.release();
    for (const cleanup of cleanups) cleanup();
  };
}
