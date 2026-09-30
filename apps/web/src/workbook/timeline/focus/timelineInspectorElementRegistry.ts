import type { InspectorPanelId } from "@cartulary/view-contracts";
import { useRef } from "react";
import { workbookInspectorSubjectsEqual } from "../../inspector/workbookInspectorSubject";
import type { WorkbookInspectorExplicitNavigation } from "../../layout/workbookInspectorNavigation";
import type { WorkbookInspectorState } from "../../models/workbookInspectorModel";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";

type TimelineInspectorElement = HTMLElement;

type TimelineInspectorElementScope = {
  readonly reviewGeneration: number;
  readonly lifecycleKey: string;
  readonly authorityKey?: string;
  readonly subject: WorkbookRecordSubject | null;
};

type TimelineInspectorFocusIdentity = {
  readonly recordId: string;
  readonly rowVersion: number;
  readonly viewSchemaId: string;
};

type MentionElementRegistration = {
  readonly element: HTMLButtonElement;
  readonly itemRef: string;
  readonly sourceRecordId: string;
};

export type TimelineInspectorElementRegistry = ReturnType<
  typeof createTimelineInspectorElementRegistry
>;

export function createTimelineInspectorElementRegistry(
  initialScope: TimelineInspectorElementScope,
) {
  let scope = initialScope;
  let root: HTMLElement | null = null;
  let destinationNavigator: WorkbookInspectorExplicitNavigation | null = null;
  const panels = new Map<InspectorPanelId, TimelineInspectorElement>();
  let evidenceList: { recordId: string; element: HTMLElement } | null = null;
  let pendingFocus: {
    identity: TimelineInspectorFocusIdentity;
    target: InspectorPanelId | "evidence_list";
  } | null = null;
  const attemptPendingFocus = () => {
    if (
      pendingFocus === null ||
      !scopeMatchesIdentity(scope, pendingFocus.identity)
    )
      return false;
    const panelId =
      pendingFocus.target === "evidence_list"
        ? "evidence"
        : pendingFocus.target;
    const element =
      pendingFocus.target === "evidence_list"
        ? evidenceList?.recordId === pendingFocus.identity.recordId
          ? evidenceList.element
          : undefined
        : panels.get(pendingFocus.target);
    if (!isUsableInspectorElement(element)) {
      if (destinationNavigator?.(panelId, null) === "unavailable")
        pendingFocus = null;
      return false;
    }
    const result = destinationNavigator?.(
      panelId,
      panelId === "history" ? undefined : element,
    );
    if (result === "applied" || result === "unavailable") pendingFocus = null;
    return result === "applied";
  };
  const triggers = new Map<string, HTMLElement>();
  let returnTarget: { readonly recordId: string; readonly key: string } | null =
    null;
  const collections = new Map<string, HTMLElement>();
  const collectionKey = (
    recordId: string,
    fieldKey: string,
    itemRef: string | null,
  ) => JSON.stringify([recordId, fieldKey, itemRef]);
  const mentions = new Map<string, MentionElementRegistration>();

  const clear = () => {
    panels.clear();
    evidenceList = null;
    pendingFocus = null;
    mentions.clear();
    collections.clear();
    root = null;
  };

  return {
    isInspectionControlTarget(target: EventTarget | null) {
      if (!(target instanceof Node)) return false;
      return (
        [...triggers.values()].some((element) => element.contains(target)) ||
        [...mentions.values()].some(({ element }) => element.contains(target))
      );
    },
    registerCollectionTrigger(
      recordId: string,
      fieldKey: string,
      itemRef: string | null,
      element: HTMLElement | null,
    ) {
      const key = collectionKey(recordId, fieldKey, itemRef);
      if (element === null) triggers.delete(key);
      else triggers.set(key, element);
    },
    rememberCollectionReturnFocus(
      recordId: string,
      fieldKey: string,
      itemRef: string | null,
    ) {
      returnTarget = {
        recordId,
        key: collectionKey(recordId, fieldKey, itemRef),
      };
    },
    restoreCollectionReturnFocus() {
      const element =
        returnTarget === null ? undefined : triggers.get(returnTarget.key);
      returnTarget = null;
      if (!isUsableInspectorElement(element)) return false;
      element.focus({ preventScroll: true });
      return document.activeElement === element;
    },
    containsActiveElement() {
      const activeElement = document.activeElement;
      if (!(activeElement instanceof HTMLElement)) return false;
      return (
        root?.contains(activeElement) === true ||
        [...panels.values()].some((panel) => panel.contains(activeElement)) ||
        [...mentions.values()].some(({ element }) =>
          element.contains(activeElement),
        )
      );
    },
    registerCollectionItem(
      recordId: string,
      fieldKey: string,
      itemRef: string,
      element: HTMLElement | null,
    ) {
      const key = collectionKey(recordId, fieldKey, itemRef);
      if (element === null) {
        collections.delete(key);
        return;
      }
      if (scope.subject?.kind === "live" && scope.subject.recordId === recordId)
        collections.set(key, element);
    },
    focusCollectionItem(
      identity: TimelineInspectorFocusIdentity,
      fieldKey: string,
      itemRef: string,
    ) {
      if (!scopeMatchesIdentity(scope, identity)) return false;
      const element = collections.get(
        collectionKey(identity.recordId, fieldKey, itemRef),
      );
      if (!isUsableInspectorElement(element)) return false;
      return destinationNavigator?.("relationships", element) === "applied";
    },
    focusMention(
      identity: TimelineInspectorFocusIdentity,
      sourceRecordId: string,
      itemRef: string,
      placement: Parameters<WorkbookInspectorExplicitNavigation>[2] = "start",
    ) {
      if (!scopeMatchesIdentity(scope, identity)) return false;
      const registration = mentions.get(itemRef);
      if (
        registration === undefined ||
        registration.sourceRecordId !== sourceRecordId ||
        !isUsableInspectorElement(registration.element)
      ) {
        return false;
      }
      return (
        destinationNavigator?.(
          "relationships",
          registration.element,
          placement,
        ) === "applied"
      );
    },
    focusPanel(
      identity: TimelineInspectorFocusIdentity,
      panelId: InspectorPanelId,
    ) {
      pendingFocus = { identity, target: panelId };
      return attemptPendingFocus();
    },
    focusEvidenceList(identity: TimelineInspectorFocusIdentity) {
      pendingFocus = { identity, target: "evidence_list" };
      return attemptPendingFocus();
    },
    registerDestinationNavigator(
      navigate: WorkbookInspectorExplicitNavigation | null,
    ) {
      destinationNavigator = navigate;
      attemptPendingFocus();
    },
    registerEvidenceList(recordId: string, element: HTMLElement | null) {
      if (element === null) {
        if (evidenceList?.recordId === recordId) evidenceList = null;
      } else evidenceList = { recordId, element };
      attemptPendingFocus();
    },
    cancelPendingFocus() {
      pendingFocus = null;
    },
    registerMention(
      sourceRecordId: string,
      itemRef: string,
      element: HTMLButtonElement | null,
    ) {
      const normalizedItemRef = itemRef.trim();
      const subject = scope.subject;
      if (
        element === null ||
        subject === null ||
        subject.kind !== "live" ||
        subject.recordId !== sourceRecordId ||
        normalizedItemRef === ""
      ) {
        if (element === null) mentions.delete(normalizedItemRef);
        return;
      }
      mentions.set(normalizedItemRef, {
        element,
        itemRef: normalizedItemRef,
        sourceRecordId,
      });
    },
    registerPanel(panelId: InspectorPanelId, element: HTMLElement | null) {
      if (element === null) {
        panels.delete(panelId);
        return;
      }
      panels.set(panelId, element);
      attemptPendingFocus();
    },
    registerRoot(element: HTMLElement | null) {
      root = scope.subject === null ? null : element;
    },
    updateScope(nextScope: TimelineInspectorElementScope) {
      if (scope.authorityKey !== nextScope.authorityKey) pendingFocus = null;
      if (scope.lifecycleKey !== nextScope.lifecycleKey) {
        triggers.clear();
        returnTarget = null;
      }
      if (
        nextScope.subject !== null &&
        returnTarget !== null &&
        nextScope.subject.recordId !== returnTarget.recordId
      )
        returnTarget = null;
      const firstOpen =
        scope.subject === null &&
        nextScope.subject?.kind === "live" &&
        scope.lifecycleKey === nextScope.lifecycleKey;
      if (
        scope.lifecycleKey !== nextScope.lifecycleKey ||
        scope.reviewGeneration !== nextScope.reviewGeneration ||
        !workbookInspectorSubjectsEqual(scope.subject, nextScope.subject)
      ) {
        if (!firstOpen) {
          const requested = pendingFocus;
          clear();
          if (
            requested &&
            scope.lifecycleKey === nextScope.lifecycleKey &&
            scopeMatchesIdentity(nextScope, requested.identity)
          )
            pendingFocus = requested;
        }
      }
      scope = nextScope;
      attemptPendingFocus();
    },
  };
}

export function useTimelineInspectorElementRegistry(
  lifecycle: WorkbookInspectorState,
  authorityKey: string,
): TimelineInspectorElementRegistry {
  const registryRef = useRef<TimelineInspectorElementRegistry | null>(null);
  registryRef.current ??= createTimelineInspectorElementRegistry({
    reviewGeneration: lifecycle.reviewGeneration,
    lifecycleKey: lifecycle.lifecycleKey,
    authorityKey,
    subject: lifecycle.phase === "open_ready" ? lifecycle.subject : null,
  });
  registryRef.current.updateScope({
    reviewGeneration: lifecycle.reviewGeneration,
    lifecycleKey: lifecycle.lifecycleKey,
    authorityKey,
    subject: lifecycle.phase === "open_ready" ? lifecycle.subject : null,
  });
  return registryRef.current;
}

function scopeMatchesIdentity(
  scope: TimelineInspectorElementScope,
  identity: TimelineInspectorFocusIdentity,
) {
  const subject = scope.subject;
  return (
    subject !== null &&
    subject.kind === "live" &&
    subject.viewSchemaId === identity.viewSchemaId &&
    subject.recordId === identity.recordId &&
    subject.rowVersion === identity.rowVersion
  );
}

function isUsableInspectorElement<T extends TimelineInspectorElement>(
  element: T | undefined,
): element is T {
  if (
    element === undefined ||
    !element.isConnected ||
    element.hidden ||
    element.closest("[hidden], [aria-hidden='true']") !== null ||
    ("disabled" in element && element.disabled === true)
  ) {
    return false;
  }
  const style = window.getComputedStyle(element);
  return style.display !== "none" && style.visibility !== "hidden";
}
