import {
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryPanelTestId,
  rowHistoryReadControlTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import type { RecordHistoryItem } from "../adapters/workbookHistoryResponse";
import { HistoryLookupFeedback } from "../history/HistoryLookupFeedback";
import {
  useHistoryRecordPending,
  useWorkbookHistoryRuntime,
} from "../history/WorkbookHistoryContext";
import { WorkbookHistoryLocalStatus } from "../history/WorkbookHistoryLocalStatus";
import type { RecordHistoryRollbackAction } from "../history/workbookHistoryItem";

import type { WorkbookRecordSubject } from "../ports/WorkbookRecordSubject";
import type { InspectorRecordHistoryAction } from "./inspectorCapabilityResolver";
import { WorkbookInspectorActionButton } from "./presentation/WorkbookInspectorActions";
import {
  WorkbookInspectorFeedbackView,
  WorkbookInspectorTechnicalDetails,
} from "./presentation/WorkbookInspectorFeedback";
import {
  type PresentInspectorRegion,
  type WorkbookInspectorPanelContentModel,
  type WorkbookInspectorPanelData,
  WorkbookInspectorRegionContent,
} from "./presentation/WorkbookInspectorPanelContent";
import { useWorkbookRecordHistoryController } from "./useWorkbookRecordHistoryController";
import { useWorkbookRecordHistoryFocus } from "./useWorkbookRecordHistoryFocus";
import { useWorkbookRecordHistoryState } from "./useWorkbookRecordHistoryState";
import { WorkbookRecordHistoryLoadedPresentation } from "./WorkbookRecordHistoryPresentation";
import {
  type WorkbookRecordHistoryState,
  workbookRecordHistoryFeedback,
  workbookRecordHistoryLoadError,
  workbookRecordHistoryLoadedData,
  workbookRecordHistoryPendingAction,
} from "./workbookRecordHistoryModel";
import type { WorkbookRecordHistoryOwnerEffects } from "./workbookRecordHistoryOwnerEffects";

export function WorkbookInspectorRecordHistory({
  actions,
  canMutate,
  ownerEffects,
  subject,
  present,
}: {
  readonly actions: ReadonlySet<InspectorRecordHistoryAction>;
  readonly canMutate: boolean;
  readonly ownerEffects: WorkbookRecordHistoryOwnerEffects;
  readonly subject: WorkbookRecordSubject | null;
  readonly present?: PresentInspectorRegion;
}) {
  const presentation = useWorkbookRecordHistoryState(subject);
  const controller = useWorkbookRecordHistoryController({
    presentation,
    canMutate,
    ownerEffects,
    subject,
  });
  return (
    <WorkbookRecordHistoryPanel
      present={present}
      actions={actions}
      canMutate={canMutate}
      idleRecordId={subject?.recordId}
      state={controller.snapshot}
      browsingControls={controller.commands}
      onCancelPendingAction={controller.commands.cancel}
      onConfirmPendingAction={() => void controller.commands.confirm()}
      onOpenHistory={controller.commands.open}
      onPreviewDeleteRestore={controller.commands.previewDeleteRestore}
      onPreviewRollback={controller.commands.previewRollback}
    />
  );
}

export type HistoryBrowsingControls = {
  readonly open: () => void;
  readonly loadOlder: () => void;
  readonly retryRead: () => void;
  readonly continuePreview: () => void;
  readonly restartPreview: () => void;
};

export function WorkbookRecordHistoryPanel({
  present = (model) => <WorkbookInspectorRegionContent model={model} />,
  requestedChangeSetId,
  locatingChange = false,
  actions,
  canMutate,
  destructiveSubject = "this record",
  idleRecordId,
  openTestId,
  refreshLabel = "Open history",
  refreshControl,
  browsingControls,
  state,
  onCancelPendingAction,
  onConfirmPendingAction,
  onOpenHistory,
  onPreviewDeleteRestore,
  onPreviewRollback,
}: {
  readonly present?: PresentInspectorRegion | undefined;
  readonly requestedChangeSetId?: string | undefined;
  readonly locatingChange?: boolean | undefined;
  readonly actions: ReadonlySet<InspectorRecordHistoryAction>;
  readonly canMutate: boolean;
  readonly destructiveSubject?: string;
  readonly idleRecordId?: string | undefined;
  readonly openTestId?: string;
  readonly refreshLabel?: string;
  readonly refreshControl?:
    | {
        readonly label: string;
        readonly testId: string;
        readonly onRefresh: () => void;
      }
    | undefined;
  readonly browsingControls?: HistoryBrowsingControls;
  readonly state: WorkbookRecordHistoryState;
  readonly onCancelPendingAction: () => void;
  readonly onConfirmPendingAction: () => void;
  readonly onOpenHistory: () => void;
  readonly onPreviewDeleteRestore: (operation: "delete" | "restore") => void;
  readonly onPreviewRollback: (
    item: RecordHistoryItem,
    action: RecordHistoryRollbackAction,
  ) => void;
}) {
  const runtime = useWorkbookHistoryRuntime();
  const retainedPending = useHistoryRecordPending(
    state.subject?.recordId ?? idleRecordId ?? null,
  );
  const data = workbookRecordHistoryLoadedData(state);
  const error = workbookRecordHistoryLoadError(state);
  const feedback = workbookRecordHistoryFeedback(state);
  const pendingAction = workbookRecordHistoryPendingAction(state);
  const focus = useWorkbookRecordHistoryFocus({
    pending: retainedPending || Boolean(state.browsing?.pending),
    canMutate,
    state,
    onCancelPendingAction,
    onConfirmPendingAction,
  });
  const refreshFocus = useRef<HTMLButtonElement | null>(null);
  const refreshButton = useRef<HTMLButtonElement | null>(null);
  const olderButton = useRef<HTMLButtonElement | null>(null);
  const retryFocus = useRef<HTMLButtonElement | null>(null);
  const [delayedLoading, setDelayedLoading] = useState(false);
  const loadGeneration =
    state.phase === "loading" ? state.browsing?.pending?.generation : null;
  useEffect(() => {
    setDelayedLoading(false);
    if (loadGeneration === null) return;
    const timer = setTimeout(() => setDelayedLoading(true), 2_000);
    return () => clearTimeout(timer);
  }, [loadGeneration]);
  useEffect(() => {
    const cancelScroll = () => {
      refreshFocus.current = null;
      retryFocus.current = null;
    };
    document.addEventListener("pointerdown", cancelScroll, true);
    document.addEventListener("keydown", cancelScroll, true);
    document.addEventListener("wheel", cancelScroll, true);
    return () => {
      document.removeEventListener("pointerdown", cancelScroll, true);
      document.removeEventListener("keydown", cancelScroll, true);
      document.removeEventListener("wheel", cancelScroll, true);
    };
  }, []);
  useLayoutEffect(() => {
    if (state.browsing?.pending || !retryFocus.current) return;
    const trigger = retryFocus.current;
    retryFocus.current = null;
    if (
      document.activeElement === trigger ||
      (!trigger.isConnected && document.activeElement === document.body)
    )
      (olderButton.current ?? refreshButton.current)?.focus({
        preventScroll: true,
      });
  }, [state.browsing]);
  useLayoutEffect(() => {
    if (!refreshFocus.current || state.browsing?.pending) return;
    const trigger = refreshFocus.current;
    refreshFocus.current = null;
    const panel = focus.panelRef.current;
    if (
      document.activeElement !== trigger ||
      !panel ||
      state.browsing?.failure ||
      !state.browsing?.accepted
    )
      return;
    for (
      let parent = panel.parentElement;
      parent;
      parent = parent.parentElement
    ) {
      if (
        /(auto|scroll)/.test(getComputedStyle(parent).overflowY) &&
        parent.scrollHeight > parent.clientHeight
      ) {
        parent.scrollTop +=
          panel.getBoundingClientRect().top -
          parent.getBoundingClientRect().top;
        break;
      }
    }
  }, [state.browsing, focus.panelRef]);
  const presentedRecordId = state.subject?.recordId ?? idleRecordId ?? null;
  if (presentedRecordId === null) return null;
  if (runtime?.history.readable === false)
    return present({ access: "concealed" });
  const busy =
    retainedPending ||
    state.phase === "loading" ||
    state.phase === "submitting" ||
    Boolean(state.browsing?.pending) ||
    state.lookup?.phase === "checking" ||
    state.lookup?.phase === "paused";
  const browsing = state.browsing;
  const readBlocked =
    locatingChange || retainedPending || state.phase === "submitting";
  const reading = Boolean(browsing?.pending);
  const refreshBlocked =
    readBlocked ||
    state.phase === "loading" ||
    browsing?.pending?.kind === "refresh";
  const acceptedContent: WorkbookInspectorPanelContentModel = data?.items.length
    ? {
        kind: "populated",
        content:
          data === null || state.subject === null ? null : (
            <>
              <WorkbookInspectorTechnicalDetails
                fields={[
                  {
                    label: "Current row version",
                    value: String(data.row_version),
                  },
                  { label: "Deleted", value: data.deleted ? "yes" : "no" },
                ]}
              />
              <WorkbookRecordHistoryLoadedPresentation
                requestedChangeSetId={requestedChangeSetId}
                actions={actions}
                busy={busy}
                canMutate={canMutate}
                data={data}
                destructiveSubject={destructiveSubject}
                focus={{
                  cancelEventReview: focus.cancelEventReview,
                  capture: focus.captureFocusRequest,
                  register: focus.registerActionElement,
                }}
                pendingAction={pendingAction}
                subject={state.subject}
                onCancelPendingAction={focus.cancelPendingAction}
                onConfirmPendingAction={focus.confirmPendingAction}
                onPreviewDeleteRestore={onPreviewDeleteRestore}
                onPreviewRollback={onPreviewRollback}
              />
            </>
          ),
      }
    : {
        kind: "empty",
        message: browsing?.accepted?.data.paging.has_more
          ? "No history entries in the loaded page. Older entries are available."
          : "No retained history.",
      };
  const regionData: WorkbookInspectorPanelData = data
    ? browsing?.failure || error
      ? {
          state: "stale_failure",
          content: acceptedContent,
          message:
            error?.primaryMessage ??
            "Could not refresh history. Previously loaded history remains visible.",
        }
      : {
          state: browsing?.pending ? "refreshing" : "ready",
          content: acceptedContent,
        }
    : state.phase === "loading" || browsing?.pending
      ? {
          state: "initial_loading",
          message: delayedLoading
            ? "Still loading this surface"
            : "Loading history...",
        }
      : {
          state: "unavailable",
          cause: error || browsing?.failure ? "load_failed" : "not_requested",
          message:
            error?.primaryMessage ??
            (browsing?.failure
              ? "Could not load history."
              : "History has not been loaded."),
        };
  const noticeRequest = browsing?.pending ?? browsing?.failure?.request;
  const readNotice =
    runtime?.history && noticeRequest
      ? {
          consume: runtime.history.inspectorNotices.consume,
          value: runtime.history.readNotice(
            presentedRecordId,
            noticeRequest.viewSchemaId,
            noticeRequest,
            regionData.state,
            "message" in regionData
              ? (regionData.message ?? "Loading history…")
              : regionData.state === "initial_loading"
                ? "Loading history…"
                : "Refreshing history…",
          ),
        }
      : undefined;
  return (
    <section
      data-testid={rowHistoryPanelTestId()}
      ref={focus.panelRef}
      style={panelStyle}
      tabIndex={-1}
    >
      {present({
        access: "readable",
        data: regionData,
        commandsPlacement: "before_content",
        messageId:
          regionData.state === "initial_loading"
            ? rowHistoryLoadingTestId()
            : rowHistoryMessageTestId(),
        ...(readNotice ? { notice: readNotice } : {}),
        commands: (
          <>
            {refreshControl === undefined ||
            browsingControls !== undefined ? null : (
              <WorkbookInspectorActionButton
                data-testid={refreshControl.testId}
                tone="secondary"
                onClick={refreshControl.onRefresh}
              >
                {refreshControl.label}
              </WorkbookInspectorActionButton>
            )}
            <WorkbookInspectorTechnicalDetails
              fields={[
                { label: "Record ID", value: presentedRecordId },
                ...(requestedChangeSetId
                  ? [
                      {
                        label: "Requested change set ID",
                        value: requestedChangeSetId,
                      },
                    ]
                  : []),
              ]}
            />
            {state.phase === "idle" && !browsingControls ? (
              <WorkbookInspectorActionButton
                data-inspector-section-entry
                data-testid={openTestId}
                onClick={onOpenHistory}
              >
                {refreshLabel}
              </WorkbookInspectorActionButton>
            ) : null}
            {browsingControls ? (
              <div
                style={{
                  display: "flex",
                  flexWrap: "wrap",
                  gap: "var(--ct-spacing-sm)",
                }}
              >
                <WorkbookInspectorActionButton
                  data-inspector-section-entry={
                    state.phase === "idle" ? true : undefined
                  }
                  data-testid={
                    state.phase === "idle"
                      ? openTestId
                      : (refreshControl?.testId ??
                        rowHistoryReadControlTestId("refresh"))
                  }
                  ref={refreshButton}
                  tone="secondary"
                  aria-disabled={refreshBlocked}
                  onClick={(event) => {
                    if (refreshBlocked) return;
                    if (state.phase !== "idle")
                      refreshFocus.current = event.currentTarget;
                    browsingControls.open();
                  }}
                >
                  {state.phase === "idle" ? refreshLabel : "Refresh history"}
                </WorkbookInspectorActionButton>
                {browsing?.accepted &&
                (browsing.accepted.data.paging.has_more ||
                  browsing.accepted.pages.length > 1) ? (
                  <WorkbookInspectorActionButton
                    data-testid={rowHistoryReadControlTestId("load-older")}
                    ref={olderButton}
                    aria-disabled={
                      reading ||
                      readBlocked ||
                      !browsing.chainValid ||
                      Boolean(browsing.failure) ||
                      !browsing.accepted.data.paging.has_more
                    }
                    aria-busy={browsing.pending?.kind === "continuation"}
                    onClick={() => {
                      if (
                        !reading &&
                        !readBlocked &&
                        browsing.chainValid &&
                        !browsing.failure &&
                        browsing.accepted?.data.paging.has_more
                      )
                        browsingControls.loadOlder();
                    }}
                  >
                    Load older entries
                  </WorkbookInspectorActionButton>
                ) : null}
                {browsing?.accepted &&
                !browsing.accepted.data.paging.has_more ? (
                  <span>No older entries.</span>
                ) : null}
                {browsing?.failure ? (
                  <WorkbookInspectorActionButton
                    data-testid={rowHistoryReadControlTestId(
                      browsing.failure.restart ? "start-fresh" : "retry",
                    )}
                    disabled={reading || readBlocked}
                    onClick={(event) => {
                      retryFocus.current = event.currentTarget;
                      if (browsing.failure?.restart) browsingControls.open();
                      else browsingControls.retryRead();
                    }}
                  >
                    {browsing.failure.restart
                      ? "Start fresh history"
                      : browsing.failure.request.kind === "continuation"
                        ? "Retry older entries"
                        : browsing.failure.request.kind === "refresh"
                          ? "Retry refresh"
                          : "Retry history"}
                  </WorkbookInspectorActionButton>
                ) : null}
              </div>
            ) : null}
            <HistoryLookupFeedback
              state={state.lookup}
              onContinue={browsingControls?.continuePreview ?? onOpenHistory}
              onRestart={browsingControls?.restartPreview ?? onOpenHistory}
              onCancel={focus.cancelPendingAction}
            />
            {error ? (
              <WorkbookInspectorTechnicalDetails
                fields={error.technicalFields}
              />
            ) : null}
            <WorkbookInspectorFeedbackView
              feedback={feedback}
              neutralStyle={metadataStyle}
              testId={rowHistoryMessageTestId()}
            />
            <WorkbookHistoryLocalStatus recordId={presentedRecordId} />

            {data?.items.length === 0 ? (
              data === null || state.subject === null ? null : (
                <>
                  <WorkbookInspectorTechnicalDetails
                    fields={[
                      {
                        label: "Current row version",
                        value: String(data.row_version),
                      },
                      { label: "Deleted", value: data.deleted ? "yes" : "no" },
                    ]}
                  />
                  <WorkbookRecordHistoryLoadedPresentation
                    requestedChangeSetId={requestedChangeSetId}
                    actions={actions}
                    busy={busy}
                    canMutate={canMutate}
                    data={data}
                    destructiveSubject={destructiveSubject}
                    focus={{
                      cancelEventReview: focus.cancelEventReview,
                      capture: focus.captureFocusRequest,
                      register: focus.registerActionElement,
                    }}
                    pendingAction={pendingAction}
                    subject={state.subject}
                    onCancelPendingAction={focus.cancelPendingAction}
                    onConfirmPendingAction={focus.confirmPendingAction}
                    onPreviewDeleteRestore={onPreviewDeleteRestore}
                    onPreviewRollback={onPreviewRollback}
                  />
                </>
              )
            ) : null}
          </>
        ),
      })}
    </section>
  );
}

const panelStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const metadataStyle = {
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
