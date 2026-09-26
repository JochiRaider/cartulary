import {
  type RefObject,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import type {
  HistoryBrowsingState,
  HistoryReadKind,
} from "../history/workbookHistoryBrowsing";
import {
  type HistoryPageProvenance,
  sameHistoryReadScope,
} from "../history/workbookHistoryPage";
import type { WorkbookRecordHistoryState } from "./workbookRecordHistoryModel";

type ReadAnchor = {
  readonly request: HistoryPageProvenance;
  readonly kind: HistoryReadKind;
  readonly panel: HTMLElement;
  readonly trigger: HTMLButtonElement;
  readonly recovery: boolean;
  readonly top: number;
  readonly scrollers: readonly {
    element: HTMLElement;
    overflowAnchor: string;
  }[];
};
type RecoveryFailure = NonNullable<HistoryBrowsingState["failure"]>;
type RetainedRecovery = {
  readonly failure: RecoveryFailure;
  readonly request: HistoryPageProvenance;
  readonly trigger: HTMLButtonElement;
};

/** Presentation-only continuity; accepted pages and request authority stay in the read owner. */
export function useWorkbookHistoryReadContinuity(
  state: WorkbookRecordHistoryState,
  panelRef: RefObject<HTMLElement | null>,
  olderButton: RefObject<HTMLButtonElement | null>,
  refreshButton: RefObject<HTMLButtonElement | null>,
  readable = true,
) {
  const anchor = useRef<ReadAnchor | null>(null);
  const [retainedRecovery, setRetainedRecovery] =
    useState<RetainedRecovery | null>(null);
  const clear = useCallback(() => {
    for (const scroller of anchor.current?.scrollers ?? [])
      scroller.element.style.overflowAnchor = scroller.overflowAnchor;
    anchor.current = null;
  }, []);
  const retire = useCallback(() => {
    clear();
    setRetainedRecovery((current) => (current === null ? current : null));
  }, [clear]);
  useEffect(() => {
    const cancel = (event: Event) => {
      // Native Tab movement determines the destination. If the analyst tabs
      // back before settlement, the initiating control still owns focus.
      if (
        event instanceof KeyboardEvent &&
        ["Tab", "Shift", "Control", "Alt", "Meta"].includes(event.key)
      )
        return;
      const trigger = anchor.current?.trigger;
      if (
        trigger &&
        event.target === trigger &&
        trigger.getAttribute("aria-disabled") === "true" &&
        (event.type === "pointerdown" ||
          (event instanceof KeyboardEvent &&
            ["Enter", " "].includes(event.key)))
      )
        return;
      clear();
    };
    const events = ["pointerdown", "keydown", "wheel", "touchstart"] as const;
    for (const event of events) document.addEventListener(event, cancel, true);
    return () => {
      for (const event of events)
        document.removeEventListener(event, cancel, true);
      clear();
    };
  }, [clear]);
  useLayoutEffect(() => {
    const saved = anchor.current;
    const browsing = state.browsing;
    if (!saved) {
      if (
        retainedRecovery &&
        (!readable ||
          !browsing ||
          browsing.generation !== retainedRecovery.request.generation ||
          state.subject?.recordId !== retainedRecovery.request.recordId ||
          state.subject.viewSchemaId !==
            retainedRecovery.request.viewSchemaId ||
          !sameHistoryReadScope(
            browsing.scope,
            retainedRecovery.request.scope,
          ) ||
          browsing.failure ||
          (!browsing.pending &&
            document.activeElement !== retainedRecovery.trigger))
      )
        setRetainedRecovery(null);
      return;
    }
    if (
      !readable ||
      !browsing ||
      !saved.panel.isConnected ||
      panelRef.current !== saved.panel ||
      state.subject?.recordId !== saved.request.recordId ||
      state.subject.viewSchemaId !== saved.request.viewSchemaId ||
      !sameHistoryReadScope(browsing.scope, saved.request.scope) ||
      browsing.generation > saved.request.generation
    ) {
      retire();
      return;
    }
    if (browsing.generation < saved.request.generation) return;
    const observation =
      browsing.pending ??
      browsing.failure?.request ??
      browsing.accepted?.pages.find(
        (page) => page.generation === saved.request.generation,
      );
    if (
      !observation ||
      observation.chainId !== saved.request.chainId ||
      observation.generation !== saved.request.generation ||
      observation.request.cursorToken !== saved.request.request.cursorToken
    ) {
      retire();
      return;
    }
    const active = document.activeElement;
    const ownsFocus = active === saved.trigger && saved.trigger.isConnected;
    if (!ownsFocus) {
      retire();
      return;
    }
    if (saved.kind === "continuation") {
      const target = olderButton.current;
      if (!target) {
        retire();
        return;
      }
      for (const { element } of saved.scrollers)
        element.scrollTop += target.getBoundingClientRect().top - saved.top;
    } else if (
      saved.kind === "refresh" &&
      !browsing.pending &&
      !browsing.failure
    ) {
      const scroller = saved.scrollers[0]?.element;
      if (scroller)
        scroller.scrollTop +=
          saved.panel.getBoundingClientRect().top -
          scroller.getBoundingClientRect().top;
    }
    if (!browsing.pending) {
      if (!browsing.failure || !saved.recovery)
        (saved.kind === "continuation"
          ? olderButton.current
          : refreshButton.current
        )?.focus({ preventScroll: true });
      retire();
    }
  }, [
    state.browsing,
    state.subject,
    panelRef,
    olderButton,
    refreshButton,
    readable,
    retainedRecovery,
    retire,
  ]);

  const captureRead = (
    trigger: HTMLButtonElement,
    kind: HistoryReadKind,
    retry = false,
    recovery = false,
  ) => {
    clear();
    const browsing = state.browsing;
    const panel = panelRef.current;
    const target =
      kind === "continuation" ? olderButton.current : refreshButton.current;
    if (
      !browsing ||
      !panel ||
      !target ||
      (browsing.pending &&
        (kind !== "refresh" || browsing.pending.kind === "refresh"))
    )
      return;
    const scrollers: { element: HTMLElement; overflowAnchor: string }[] = [];
    for (
      let parent = panel.parentElement;
      parent;
      parent = parent.parentElement
    ) {
      if (
        /(auto|scroll)/.test(getComputedStyle(parent).overflowY) ||
        parent === document.scrollingElement
      ) {
        scrollers.push({
          element: parent,
          overflowAnchor: parent.style.overflowAnchor,
        });
        parent.style.overflowAnchor = "none";
      }
    }
    const request: HistoryPageProvenance = {
      scope: browsing.scope,
      recordId: browsing.recordId,
      viewSchemaId: browsing.viewSchemaId,
      generation: browsing.generation + 1,
      chainId: browsing.chainId + Number(kind !== "continuation"),
      effectiveLimit: browsing.accepted?.data.paging.limit ?? null,
      request:
        retry && browsing.failure
          ? browsing.failure.request.request
          : kind === "continuation"
            ? {
                cursorToken: browsing.accepted?.data.paging
                  .next_cursor as string,
              }
            : {},
    };
    anchor.current = {
      panel,
      trigger,
      kind,
      recovery,
      top: target.getBoundingClientRect().top,
      scrollers,
      request,
    };
    if (recovery && browsing.failure)
      setRetainedRecovery({ failure: browsing.failure, request, trigger });
  };
  const browsing = state.browsing;
  const recovery =
    readable && browsing
      ? (browsing.failure ??
        (retainedRecovery &&
        browsing.generation === retainedRecovery.request.generation &&
        state.subject?.recordId === retainedRecovery.request.recordId &&
        state.subject.viewSchemaId === retainedRecovery.request.viewSchemaId &&
        sameHistoryReadScope(browsing.scope, retainedRecovery.request.scope)
          ? retainedRecovery.failure
          : null))
      : null;
  return {
    captureRead,
    recovery,
    releaseRecovery: () => {
      if (!browsing?.pending && !browsing?.failure) setRetainedRecovery(null);
    },
  };
}
