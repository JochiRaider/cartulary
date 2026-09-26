import type { GridHandle } from "@cartulary/grid-adapter";
import type { ViewContract } from "@cartulary/view-contracts";
import {
  type RefObject,
  useCallback,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import type { WorkbookGridDraftStore } from "../models/WorkbookGridDraftStore";
import { focusWorkbookShellFallback } from "./workbookFocusFallback";

export type ParkedGridDraft = {
  readonly key: string;
  readonly recordId: string;
  readonly label: string;
  readonly value: string | null;
  readonly reason: string;
  readonly discard: () => void;
};

type ParkedDraftFocusPort = {
  readonly focusScopeKey: string;
  readonly gridHandleRef: RefObject<GridHandle | null>;
};

type OwnedDisclosureFocus =
  | { readonly kind: "summary"; readonly element: HTMLElement }
  | {
      readonly kind: "draft";
      readonly key: string;
      readonly element: HTMLElement;
    };

function disclosureFocusOwner(
  target: Element | null,
  details: HTMLDetailsElement | null,
  summary: HTMLElement | null,
): OwnedDisclosureFocus | null {
  if (!(target instanceof HTMLElement) || !details?.contains(target))
    return null;
  if (target === summary) return { kind: "summary", element: target };
  const section = target.closest<HTMLElement>("[data-parked-draft-key]");
  const key = section?.dataset.parkedDraftKey;
  return key === undefined ? null : { kind: "draft", key, element: target };
}

function revealRetainedText(
  control: HTMLTextAreaElement,
  scrollport: HTMLDivElement | null,
) {
  if (!scrollport) return;
  const controlRect = control.getBoundingClientRect();
  const viewport = scrollport.getBoundingClientRect();
  if (controlRect.top < viewport.top)
    scrollport.scrollTop += controlRect.top - viewport.top;
  else if (controlRect.bottom > viewport.bottom)
    scrollport.scrollTop += controlRect.bottom - viewport.bottom;
  if (controlRect.left < viewport.left)
    scrollport.scrollLeft += controlRect.left - viewport.left;
  else if (controlRect.right > viewport.right)
    scrollport.scrollLeft += controlRect.right - viewport.right;
}

/** Readable local work whose original cell is currently unavailable for editing. */
export function WorkbookParkedGridDrafts({
  drafts,
  focusScopeKey,
  gridHandleRef,
}: {
  readonly drafts: readonly ParkedGridDraft[];
} & ParkedDraftFocusPort) {
  const detailsRef = useRef<HTMLDetailsElement>(null);
  const summaryRef = useRef<HTMLElement>(null);
  const scrollportRef = useRef<HTMLDivElement>(null);
  const retainedTextRef = useRef(new Map<string, HTMLTextAreaElement>());
  const ownedFocusRef = useRef<OwnedDisclosureFocus | null>(null);
  const previousKeysRef = useRef(drafts.map((draft) => draft.key));
  const scopeRef = useRef(focusScopeKey);
  const pendingFocusRef = useRef<AbortController | null>(null);
  const cancelPendingFocus = useCallback(() => {
    pendingFocusRef.current?.abort();
    pendingFocusRef.current = null;
  }, []);

  useLayoutEffect(() => {
    const onFocusIn = (event: FocusEvent) => {
      if (
        event.target instanceof Node &&
        detailsRef.current?.contains(event.target)
      )
        return;
      ownedFocusRef.current = null;
      if (event.target !== gridHandleRef.current?.getScrollElement())
        cancelPendingFocus();
    };
    const onPointerDown = (event: PointerEvent) => {
      if (
        event.target instanceof Node &&
        detailsRef.current?.contains(event.target)
      )
        return;
      ownedFocusRef.current = null;
      cancelPendingFocus();
    };
    document.addEventListener("focusin", onFocusIn, true);
    document.addEventListener("pointerdown", onPointerDown, true);
    document.addEventListener("keydown", cancelPendingFocus, true);
    document.addEventListener("wheel", cancelPendingFocus, true);
    return () => {
      cancelPendingFocus();
      document.removeEventListener("focusin", onFocusIn, true);
      document.removeEventListener("pointerdown", onPointerDown, true);
      document.removeEventListener("keydown", cancelPendingFocus, true);
      document.removeEventListener("wheel", cancelPendingFocus, true);
    };
  }, [cancelPendingFocus, gridHandleRef]);

  useLayoutEffect(() => {
    const keys = drafts.map((draft) => draft.key);
    const previousKeys = previousKeysRef.current;
    previousKeysRef.current = keys;
    if (scopeRef.current !== focusScopeKey) {
      scopeRef.current = focusScopeKey;
      cancelPendingFocus();
      ownedFocusRef.current = disclosureFocusOwner(
        document.activeElement,
        detailsRef.current,
        summaryRef.current,
      );
      return;
    }
    const owned = ownedFocusRef.current;
    if (!owned || owned.element.isConnected) return;
    if (
      document.activeElement !== document.body &&
      document.activeElement !== document.documentElement
    ) {
      ownedFocusRef.current = null;
      return;
    }
    if (drafts.length && owned.kind === "draft") {
      const oldIndex = previousKeys.indexOf(owned.key);
      const survivors = new Set(keys);
      const candidates = [
        ...previousKeys.slice(oldIndex + 1).filter((key) => survivors.has(key)),
        ...previousKeys
          .slice(0, oldIndex < 0 ? 0 : oldIndex)
          .reverse()
          .filter((key) => survivors.has(key)),
      ];
      for (const key of candidates) {
        const control = retainedTextRef.current.get(key);
        if (!detailsRef.current?.open || !control?.isConnected) continue;
        control.focus({ preventScroll: true });
        if (document.activeElement === control) {
          revealRetainedText(control, scrollportRef.current);
          return;
        }
      }
    }
    if (drafts.length && summaryRef.current?.isConnected) {
      summaryRef.current.focus({ preventScroll: true });
      if (document.activeElement === summaryRef.current) return;
    }
    if (drafts.length) return;

    cancelPendingFocus();
    const controller = new AbortController();
    pendingFocusRef.current = controller;
    const grid = gridHandleRef.current;
    const gridRoot = grid?.getScrollElement();
    const scope = focusScopeKey;
    void (async () => {
      if (grid && gridRoot?.isConnected) {
        let focused = false;
        try {
          const result = await grid.requestFocus(
            { kind: "root" },
            { signal: controller.signal, preserveSelection: true },
          );
          focused =
            result === "focused" &&
            document.activeElement instanceof Node &&
            gridRoot.contains(document.activeElement);
        } catch {
          // The current shell destination is still available if the grid retires.
        }
        if (controller.signal.aborted || scopeRef.current !== scope) return;
        if (focused) return;
      }
      if (
        controller.signal.aborted ||
        scopeRef.current !== scope ||
        (document.activeElement !== document.body &&
          document.activeElement !== document.documentElement)
      )
        return;
      pendingFocusRef.current = null;
      focusWorkbookShellFallback();
    })().finally(() => {
      if (pendingFocusRef.current === controller)
        pendingFocusRef.current = null;
    });
  }, [cancelPendingFocus, drafts, focusScopeKey, gridHandleRef]);

  if (!drafts.length) return null;
  return (
    <details
      data-grid-editor-external-action="true"
      data-grid-parked-drafts="true"
      ref={detailsRef}
      onFocusCapture={(event) => {
        ownedFocusRef.current = disclosureFocusOwner(
          event.target,
          detailsRef.current,
          summaryRef.current,
        );
      }}
    >
      <summary ref={summaryRef}>Unsaved cells ({drafts.length})</summary>
      <style>{`[data-grid-parked-drafts] textarea:focus { outline: var(--ct-component-focus-ring-border); outline-offset: var(--ct-component-focus-ring-offset); }`}</style>
      <div
        ref={scrollportRef}
        style={{ maxBlockSize: "12rem", overflow: "auto" }}
      >
        {drafts.map((draft) => (
          <section
            key={draft.key}
            data-parked-draft-key={draft.key}
            aria-label={`Unsaved ${draft.label}`}
          >
            <span>
              {draft.label} · {draft.recordId} · {draft.reason}
            </span>
            <textarea
              aria-label={`Retained ${draft.label}`}
              ref={(element) => {
                if (element) retainedTextRef.current.set(draft.key, element);
                else retainedTextRef.current.delete(draft.key);
              }}
              readOnly
              value={draft.value ?? ""}
              rows={2}
              style={{
                boxSizing: "border-box",
                display: "block",
                width: "100%",
              }}
            />
            {draft.value === null ? <span>Explicit clear</span> : null}
            <button type="button" onClick={draft.discard}>
              Discard {draft.label} draft
            </button>
          </section>
        ))}
      </div>
    </details>
  );
}

export function WorkbookUnavailableGridDrafts({
  store,
  contract,
  recordIds,
  fieldKeys,
  focusScopeKey,
  gridHandleRef,
}: {
  readonly store: WorkbookGridDraftStore;
  readonly contract: ViewContract;
  readonly recordIds: readonly string[];
  readonly fieldKeys: readonly string[];
} & ParkedDraftFocusPort) {
  useSyncExternalStore(store.subscribe, store.getSnapshot);
  const drafts = store.list(contract.viewSchemaId).flatMap((draft) => {
    const { identity } = draft;
    const field = contract.fieldMap[identity.fieldKey];
    const reason = !store.canAuthor()
      ? "Editing is unavailable."
      : !recordIds.includes(identity.recordId)
        ? "Original row is outside this result."
        : !fieldKeys.includes(identity.fieldKey) || !field?.gridEditable
          ? "Original field is unavailable."
          : null;
    return reason
      ? [
          {
            key: `${identity.recordId}:${identity.fieldKey}`,
            recordId: identity.recordId,
            label: field?.label ?? identity.fieldKey,
            value: draft.value,
            reason,
            discard: () => store.discard(identity),
          },
        ]
      : [];
  });
  return (
    <WorkbookParkedGridDrafts
      drafts={drafts}
      focusScopeKey={focusScopeKey}
      gridHandleRef={gridHandleRef}
    />
  );
}
