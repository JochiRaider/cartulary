import { type MutableRefObject, useCallback, useEffect, useRef } from "react";
import type { WorkbookInspectorSubject } from "./workbookInspectorSubject";
import {
  type RecordHistoryRollbackAction,
  type WorkbookRecordHistoryState,
  workbookRecordHistoryFeedback,
  workbookRecordHistoryLoadError,
} from "./workbookRecordHistoryModel";

type HistoryFocusKind = "delete" | "restore" | "rollback";

type HistoryFocusRequest = {
  readonly historyItemRef?: string;
  readonly actionIdentity: string;
  readonly kind: HistoryFocusKind;
  readonly stage: "preview" | "submitted";
  readonly subjectIdentity: string;
};

export function useWorkbookRecordHistoryFocus({
  canMutate,
  state,
  onCancelPendingAction,
  onConfirmPendingAction,
}: {
  readonly canMutate: boolean;
  readonly state: WorkbookRecordHistoryState;
  readonly onCancelPendingAction: () => void;
  readonly onConfirmPendingAction: () => void;
}) {
  const actionElementsRef = useRef(new Map<string, HTMLButtonElement>());
  const interactionGeneration = useRef(0);
  const cancelRef = useRef(onCancelPendingAction);
  cancelRef.current = onCancelPendingAction;
  const focusRequestRef = useRef<HistoryFocusRequest | null>(null);
  const panelRef = useRef<HTMLElement>(null);
  const currentSubjectIdentity = historySubjectIdentity(state.subject);
  const feedback = workbookRecordHistoryFeedback(state);
  const loadError = workbookRecordHistoryLoadError(state);

  const queueActionFocus = useCallback((actionIdentity: string) => {
    const generation = interactionGeneration.current;
    queueMicrotask(() => {
      if (generation !== interactionGeneration.current) return;
      const action = actionElementsRef.current.get(actionIdentity);
      if (historyFocusElementIsAvailable(action)) {
        action.focus({ preventScroll: true });
        action.scrollIntoView?.({ block: "nearest", inline: "nearest" });
        return;
      }
      focusHistoryPanel(panelRef.current);
    });
  }, []);

  const queuePanelFocus = useCallback(() => {
    const generation = interactionGeneration.current;
    queueMicrotask(() => {
      if (generation === interactionGeneration.current)
        focusHistoryPanel(panelRef.current);
    });
  }, []);

  useEffect(() => {
    const interact = (event: Event) => {
      const inside =
        event.target instanceof Node &&
        panelRef.current?.contains(event.target);
      if (
        inside &&
        (event.type === "focusin" ||
          focusRequestRef.current?.stage !== "submitted")
      )
        return;
      interactionGeneration.current += 1;
      if (focusRequestRef.current?.stage === "preview") cancelRef.current();
      focusRequestRef.current = null;
    };
    document.addEventListener("pointerdown", interact, true);
    document.addEventListener("keydown", interact, true);
    document.addEventListener("focusin", interact, true);
    return () => {
      interactionGeneration.current += 1;
      document.removeEventListener("pointerdown", interact, true);
      document.removeEventListener("keydown", interact, true);
      document.removeEventListener("focusin", interact, true);
    };
  }, []);

  useEffect(() => {
    completeHistoryFocusRequest({
      canMutate,
      currentSubjectIdentity,
      feedback,
      focusRequestRef,
      loadError,
      phase: state.phase,
      queueActionFocus,
      queuePanelFocus,
    });
  }, [
    canMutate,
    currentSubjectIdentity,
    feedback,
    loadError,
    queueActionFocus,
    queuePanelFocus,
    state.phase,
  ]);

  const captureFocusRequest = useCallback(
    (
      actionIdentity: string,
      kind: HistoryFocusKind,
      element: HTMLButtonElement,
      historyItemRef?: string,
    ) => {
      if (currentSubjectIdentity === null) return;
      actionElementsRef.current.set(actionIdentity, element);
      focusRequestRef.current = {
        ...(historyItemRef ? { historyItemRef } : {}),
        actionIdentity,
        kind,
        stage: "preview",
        subjectIdentity: currentSubjectIdentity,
      };
    },
    [currentSubjectIdentity],
  );

  const cancelPendingAction = useCallback(() => {
    const request = focusRequestRef.current;
    focusRequestRef.current = null;
    onCancelPendingAction();
    if (
      request !== null &&
      canMutate &&
      request.subjectIdentity === currentSubjectIdentity
    ) {
      queueActionFocus(request.actionIdentity);
    }
  }, [
    canMutate,
    currentSubjectIdentity,
    onCancelPendingAction,
    queueActionFocus,
  ]);

  const confirmPendingAction = useCallback(() => {
    const request = focusRequestRef.current;
    if (request !== null) {
      focusRequestRef.current = { ...request, stage: "submitted" };
    }
    onConfirmPendingAction();
  }, [onConfirmPendingAction]);

  return {
    cancelEventReview: (historyItemRef: string) => {
      const request = focusRequestRef.current;
      if (
        request?.stage === "preview" &&
        request.historyItemRef === historyItemRef
      )
        cancelPendingAction();
    },
    cancelPendingAction,
    captureFocusRequest,
    confirmPendingAction,
    panelRef,
    registerActionElement: (
      identity: string,
      element: HTMLButtonElement | null,
    ) =>
      registerHistoryActionElement(
        actionElementsRef.current,
        identity,
        element,
      ),
  };
}

export function historyActionIdentity(
  subject: WorkbookInspectorSubject,
  action: "delete" | "restore",
): string {
  return `${historySubjectIdentity(subject)}:${action}`;
}

export function historyRollbackActionIdentity(
  subject: WorkbookInspectorSubject,
  historyItemRef: string,
  action: RecordHistoryRollbackAction,
): string {
  return `${historySubjectIdentity(subject)}:rollback:${action}:${historyItemRef}`;
}

function completeHistoryFocusRequest({
  canMutate,
  currentSubjectIdentity,
  feedback,
  focusRequestRef,
  loadError,
  phase,
  queueActionFocus,
  queuePanelFocus,
}: {
  readonly canMutate: boolean;
  readonly currentSubjectIdentity: string | null;
  readonly feedback: ReturnType<typeof workbookRecordHistoryFeedback>;
  readonly focusRequestRef: MutableRefObject<HistoryFocusRequest | null>;
  readonly loadError: ReturnType<typeof workbookRecordHistoryLoadError>;
  readonly phase: WorkbookRecordHistoryState["phase"];
  readonly queueActionFocus: (identity: string) => void;
  readonly queuePanelFocus: () => void;
}): void {
  const request = focusRequestRef.current;
  if (request === null) return;
  const sameSubject = request.subjectIdentity === currentSubjectIdentity;
  if (historyFocusRequestWasInvalidated(canMutate, sameSubject, feedback)) {
    focusRequestRef.current = null;
    return;
  }
  if (request.stage === "preview") return;
  if (historySubmissionWasRejected(phase, loadError, feedback)) {
    focusRequestRef.current = null;
    if (sameSubject) queueActionFocus(request.actionIdentity);
    return;
  }
  if (!historySubmissionCompletionIsReady(request, phase, feedback)) {
    return;
  }
  focusRequestRef.current = null;
  if (request.kind === "rollback" && sameSubject) {
    queueActionFocus(request.actionIdentity);
  } else {
    queuePanelFocus();
  }
}

function historyFocusRequestWasInvalidated(
  canMutate: boolean,
  sameSubject: boolean,
  feedback: ReturnType<typeof workbookRecordHistoryFeedback>,
): boolean {
  return !canMutate || (!sameSubject && feedback === null);
}

function historySubmissionWasRejected(
  phase: WorkbookRecordHistoryState["phase"],
  loadError: ReturnType<typeof workbookRecordHistoryLoadError>,
  feedback: ReturnType<typeof workbookRecordHistoryFeedback>,
): boolean {
  return (
    phase === "ready" && (loadError !== null || feedback?.kind === "error")
  );
}

function historySubmissionCompletionIsReady(
  request: HistoryFocusRequest,
  phase: WorkbookRecordHistoryState["phase"],
  feedback: ReturnType<typeof workbookRecordHistoryFeedback>,
): boolean {
  return (
    feedback !== null && (request.kind !== "rollback" || phase === "ready")
  );
}

function historySubjectIdentity(
  subject: WorkbookInspectorSubject | null,
): string | null {
  return subject === null
    ? null
    : `${subject.kind}:${subject.recordId}:${subject.rowVersion}`;
}

function registerHistoryActionElement(
  elements: Map<string, HTMLButtonElement>,
  identity: string,
  element: HTMLButtonElement | null,
): void {
  if (element === null) {
    elements.delete(identity);
  } else {
    elements.set(identity, element);
  }
}

function focusHistoryPanel(panel: HTMLElement | null): void {
  if (historyFocusElementIsAvailable(panel)) {
    panel.focus({ preventScroll: true });
    panel.scrollIntoView?.({ block: "nearest", inline: "nearest" });
  }
}

function historyFocusElementIsAvailable(
  element: HTMLElement | null | undefined,
): element is HTMLElement {
  if (element === null || element === undefined || !element.isConnected) {
    return false;
  }
  if (element instanceof HTMLButtonElement && element.disabled) return false;
  for (let current: HTMLElement | null = element; current !== null; ) {
    if (
      current instanceof HTMLDetailsElement &&
      !current.open &&
      !current.querySelector(":scope > summary")?.contains(element)
    )
      return false;
    const style = globalThis.getComputedStyle(current);
    if (
      current.hidden ||
      current.getAttribute("aria-hidden") === "true" ||
      style.display === "none" ||
      style.visibility === "hidden"
    ) {
      return false;
    }
    current = current.parentElement;
  }
  return true;
}
