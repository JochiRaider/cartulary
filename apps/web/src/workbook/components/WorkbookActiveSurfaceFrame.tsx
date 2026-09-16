import { workbookActiveSurfaceFocusTargetTestId } from "@cartulary/ui-contracts";
import {
  type ReactNode,
  type RefObject,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import {
  useWorkbookRecoveryActivation,
  useWorkbookRecoveryNavigation,
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../shared/WorkbookRecoveryBoundary";
import {
  type WorkbookRecoveryItem,
  workbookConflictRecoveryKey,
} from "../../shared/workbookRecoveryNavigation";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import { shellActiveSurfaceStyle } from "../layout/workbookShellStyles";
import type {
  WorkbookMutationRuntime,
  WorkbookStatusPresentation,
} from "../runtime/WorkbookMutationRuntime";
import { WorkbookEditRecoveryPanel } from "./WorkbookEditRecoveryPanel";
import { WorkbookQueueOverflowNotice } from "./WorkbookQueueOverflowNotice";
import { WorkbookSameFieldConflictResolver } from "./WorkbookSameFieldConflictResolver";

export function WorkbookActiveSurfaceFrame({
  activeContent,
  activeSurfaceRef,
  apiBase,
  focus,
  mutationRuntime,
  mutationSnapshot,
  onActivateOrigin,
}: {
  readonly activeContent: ReactNode;
  readonly activeSurfaceRef: RefObject<HTMLElement | null>;
  readonly apiBase: string | undefined;
  readonly focus: {
    readonly resolverActivation: {
      readonly conflictKey: string;
      readonly sequence: number;
    } | null;
    readonly sameFieldSummaryRef: RefObject<HTMLDivElement | null>;
  };
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly mutationSnapshot: WorkbookStatusPresentation;
  readonly onActivateOrigin: (viewSchemaId: string) => void;
}) {
  const navigation = useWorkbookRecoveryNavigation();
  const activateRecovery = useWorkbookRecoveryActivation();
  // Presentation continuation only: no request, draft, receipt or replay state.
  const retry = useRef<{
    unitId: string;
    conflictKeys: readonly string[];
    activation: number;
  } | null>(null);
  const retryUnit =
    retry.current === null
      ? undefined
      : mutationRuntime
          .pendingQueue()
          .model.snapshot()
          .units.find((unit) => unit.id === retry.current?.unitId);
  const items: WorkbookRecoveryItem[] = mutationSnapshot.authPaused
    ? []
    : mutationSnapshot.conflicts
        .filter((entry) => !entry.batchOperationId)
        .map((entry, order) => ({
          id: `conflict:${entry.key}`,
          conflictKey: entry.key,
          label: "Same-field conflict",
          summary: "Review saved and unsaved values",
          origin: `${entry.origin.surfaceLabel}: ${entry.origin.rowLabel}`,
          ...(entry.compoundOperationId
            ? { conflictOperationId: entry.compoundOperationId }
            : {}),
          sheetRef: entry.origin.sheetRef ?? null,
          order,
          priority: 2,
          attention: "attention",
        }));
  const blocked = mutationSnapshot.blockedEdit;
  if (!mutationSnapshot.authPaused && blocked)
    items.push({
      id: `fifo:${blocked.unitId}`,
      label: "Queued edit recovery",
      summary: blocked.message,
      origin: "Workbook",
      sheetRef: null,
      order: 0,
      priority: blocked.kind === "client_txn_conflict" ? 0 : 3,
      attention: "attention",
    });
  if (!mutationSnapshot.authPaused && mutationSnapshot.overflowMessage)
    items.push({
      id: "overflow",
      label: "Pending queue full",
      summary: mutationSnapshot.overflowMessage,
      origin: "Workbook",
      sheetRef: null,
      order: 0,
      priority: 1,
      attention: "attention",
    });
  if (
    !mutationSnapshot.authPaused &&
    retryUnit &&
    blocked?.unitId !== retryUnit.id
  )
    items.push({
      id: `fifo:${retryUnit.id}`,
      label: "Queued edit recovery",
      summary: "Retried edit awaiting settlement",
      origin: "Workbook",
      sheetRef: null,
      order: retryUnit.enqueueOrder,
      attention: "progress",
    });
  const selected = useWorkbookRecoverySource("core", items);
  const nav = useSyncExternalStore(
    navigation?.subscribe ?? subscribeEmpty,
    navigation?.getSnapshot ?? snapshotEmpty,
  );
  useLayoutEffect(() => {
    const continuation = retry.current;
    if (!continuation || retryUnit) return;
    retry.current = null;
    const current = navigation?.getSnapshot();
    const conflict = mutationSnapshot.conflicts.find((entry) =>
      continuation.conflictKeys.includes(entry.key),
    );
    if (
      conflict &&
      current?.open &&
      current.activation === continuation.activation &&
      document.activeElement?.closest("#workbook-recovery-panel")
    )
      navigation?.activate(
        workbookConflictRecoveryKey(current.entries, conflict),
      );
  }, [mutationSnapshot.conflicts, navigation, retryUnit]);
  const retryBlocked = async () => {
    const unit = mutationRuntime
      .pendingQueue()
      .model.snapshot()
      .units.find((unit) => unit.id === blocked?.unitId);
    if (unit && navigation)
      retry.current = {
        unitId: unit.id,
        conflictKeys:
          unit.identity.kind === "patch"
            ? unit.identity.changes.map(
                (change) => `${unit.recordId}:${change.field_key}`,
              )
            : [],
        activation: navigation.getSnapshot().activation,
      };
    const result = await mutationRuntime.retryBlockedEdit();
    if (!result.ok) retry.current = null;
    return result;
  };
  const selectedEntry = nav?.entries.find(
    (entry) => entry.key === nav.selected,
  );
  const conflicts = mutationSnapshot.conflicts.filter((entry) =>
    selectedEntry?.source === "batch"
      ? entry.batchOperationId === selectedEntry.id
      : selectedEntry?.conflictKeys?.includes(entry.key) ||
        (entry.compoundOperationId &&
          selectedEntry?.operationIds?.includes(entry.compoundOperationId)) ||
        `conflict:${entry.key}` === selected,
  );
  const notice = mutationSnapshot.authPaused
    ? null
    : mutationSnapshot.secondary?.kind === "same_field_conflict" ||
        mutationSnapshot.secondary?.kind === "queue_overflow" ||
        blocked
      ? mutationSnapshot.secondary?.message
      : null;
  return (
    <section
      aria-label="Active workbook surface focus target"
      data-testid={workbookActiveSurfaceFocusTargetTestId()}
      ref={activeSurfaceRef}
      style={{
        ...shellActiveSurfaceStyle,
        gridTemplateRows: notice ? "auto minmax(0, 1fr)" : "minmax(0, 1fr)",
      }}
      tabIndex={-1}
    >
      {notice ? (
        <section
          aria-label="Recovery attention"
          style={{
            padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
            overflowWrap: "anywhere",
          }}
        >
          {notice}{" "}
          <WorkbookInspectorActionButton
            tone="secondary"
            onClick={() => activateRecovery()}
          >
            Review recovery
          </WorkbookInspectorActionButton>
        </section>
      ) : null}
      {activeContent}
      <WorkbookRecoveryDetail source="core" item={selected}>
        {blocked && selected === `fifo:${blocked.unitId}` ? (
          <WorkbookEditRecoveryPanel
            key={blocked.unitId}
            blockedEdit={blocked}
            onDiscard={() => mutationRuntime.discardBlockedEdit()}
            onRetry={retryBlocked}
          />
        ) : null}
        {retryUnit && selected === `fifo:${retryUnit.id}` && !blocked ? (
          <p>Retried edit awaiting settlement.</p>
        ) : null}
        {selected === "overflow" && mutationSnapshot.overflowMessage ? (
          <WorkbookQueueOverflowNotice
            message={mutationSnapshot.overflowMessage}
            onClose={() => navigation?.close()}
          />
        ) : null}
      </WorkbookRecoveryDetail>
      {conflicts.length && selectedEntry ? (
        <WorkbookRecoveryDetail
          source={selectedEntry.source}
          item={selectedEntry.id}
        >
          <WorkbookSameFieldConflictResolver
            key={selectedEntry.key}
            activation={
              focus.resolverActivation &&
              conflicts.some(
                (entry) => entry.key === focus.resolverActivation?.conflictKey,
              )
                ? focus.resolverActivation
                : null
            }
            apiBase={apiBase}
            onClose={() => navigation?.close()}
            mutationRuntime={mutationRuntime}
            onActivateOrigin={onActivateOrigin}
            snapshot={{ conflicts }}
            summaryRef={focus.sameFieldSummaryRef}
          />
        </WorkbookRecoveryDetail>
      ) : null}
    </section>
  );
}
const subscribeEmpty = () => () => {};
const snapshotEmpty = () => null;
