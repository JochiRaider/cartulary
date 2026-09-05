import type { InspectorPanelId } from "@cartulary/view-contracts";
import { useRef } from "react";
import type { WorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import { workbookInspectorSubjectsEqual } from "../../inspector/workbookInspectorSubject";
import type { WorkbookInspectorState } from "../../models/workbookInspectorModel";

type TimelineInspectorElement = HTMLElement;

type TimelineInspectorElementScope = {
  readonly invalidationGeneration: number;
  readonly lifecycleKey: string;
  readonly subject: WorkbookInspectorSubject | null;
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
  const panels = new Map<InspectorPanelId, TimelineInspectorElement>();
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
      element.focus({ preventScroll: true });
      element.scrollIntoView?.({ block: "nearest", inline: "nearest" });
      return document.activeElement === element;
    },
    focusMention(
      identity: TimelineInspectorFocusIdentity,
      sourceRecordId: string,
      itemRef: string,
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
      registration.element.focus({ preventScroll: true });
      return document.activeElement === registration.element;
    },
    focusPanel(
      identity: TimelineInspectorFocusIdentity,
      panelId: InspectorPanelId,
    ) {
      if (!scopeMatchesIdentity(scope, identity)) return false;
      const element = panels.get(panelId);
      if (!isUsableInspectorElement(element)) return false;
      element.focus({ preventScroll: true });
      return document.activeElement === element;
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
      if (scope.subject !== null) panels.set(panelId, element);
    },
    registerRoot(element: HTMLElement | null) {
      root = scope.subject === null ? null : element;
    },
    updateScope(nextScope: TimelineInspectorElementScope) {
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
      if (
        scope.lifecycleKey !== nextScope.lifecycleKey ||
        scope.invalidationGeneration !== nextScope.invalidationGeneration ||
        !workbookInspectorSubjectsEqual(scope.subject, nextScope.subject)
      ) {
        clear();
      }
      scope = nextScope;
    },
  };
}

export function useTimelineInspectorElementRegistry(
  lifecycle: WorkbookInspectorState,
): TimelineInspectorElementRegistry {
  const registryRef = useRef<TimelineInspectorElementRegistry | null>(null);
  registryRef.current ??= createTimelineInspectorElementRegistry({
    invalidationGeneration: lifecycle.invalidationGeneration,
    lifecycleKey: lifecycle.lifecycleKey,
    subject: lifecycle.phase === "open_ready" ? lifecycle.subject : null,
  });
  registryRef.current.updateScope({
    invalidationGeneration: lifecycle.invalidationGeneration,
    lifecycleKey: lifecycle.lifecycleKey,
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
