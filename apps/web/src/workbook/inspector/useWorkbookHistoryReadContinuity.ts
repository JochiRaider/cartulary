import {
  type RefObject,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";
import type { HistoryReadKind } from "../history/workbookHistoryBrowsing";
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
  readonly top: number;
  readonly scrollers: readonly {
    element: HTMLElement;
    overflowAnchor: string;
  }[];
};

/** Presentation-only continuity; accepted pages and request authority stay in the read owner. */
export function useWorkbookHistoryReadContinuity(
  state: WorkbookRecordHistoryState,
  panelRef: RefObject<HTMLElement | null>,
  olderButton: RefObject<HTMLButtonElement | null>,
  refreshButton: RefObject<HTMLButtonElement | null>,
) {
  const anchor = useRef<ReadAnchor | null>(null);
  const clear = useCallback(() => {
    for (const scroller of anchor.current?.scrollers ?? [])
      scroller.element.style.overflowAnchor = scroller.overflowAnchor;
    anchor.current = null;
  }, []);
  useEffect(() => {
    const cancel = () => clear();
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
    if (!saved) return;
    const browsing = state.browsing;
    if (
      !browsing ||
      !saved.panel.isConnected ||
      panelRef.current !== saved.panel ||
      state.subject?.recordId !== saved.request.recordId ||
      state.subject.viewSchemaId !== saved.request.viewSchemaId ||
      !sameHistoryReadScope(browsing.scope, saved.request.scope) ||
      browsing.generation > saved.request.generation
    ) {
      clear();
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
      clear();
      return;
    }
    const active = document.activeElement;
    const ownsFocus =
      active === saved.trigger ||
      (!saved.trigger.isConnected && active === document.body);
    if (!ownsFocus) {
      clear();
      return;
    }
    if (saved.kind === "continuation") {
      const target = olderButton.current;
      if (!target) {
        clear();
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
      (saved.kind === "continuation"
        ? olderButton.current
        : refreshButton.current
      )?.focus({ preventScroll: true });
      clear();
    }
  }, [
    state.browsing,
    state.subject,
    panelRef,
    olderButton,
    refreshButton,
    clear,
  ]);

  return (trigger: HTMLButtonElement, kind: HistoryReadKind, retry = false) => {
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
    anchor.current = {
      panel,
      trigger,
      kind,
      top: target.getBoundingClientRect().top,
      scrollers,
      request: {
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
      },
    };
  };
}
