import {
  type GridCellAnchor,
  type GridCellNavigationOptions,
  type GridCellNavigationResult,
  type GridHandle,
  type GridPresentationSnapshot,
  gridAnchorKey,
  gridRowIdentitiesEqual,
} from "@cartulary/grid-adapter";
import {
  useCallback,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { useWorkbookRecoveryNavigation } from "../../shared/WorkbookRecoveryBoundary";
import { decideWorkbookApplicationShortcut } from "../policies/workbookApplicationShortcuts";
import type { WorkbookQueryBrowser } from "../query/WorkbookQueryBrowser";
import type { WorkbookViewQueryAccepted } from "../query/WorkbookViewQueryPort";
import {
  WorkbookFindController,
  type WorkbookFindSource,
} from "./WorkbookFindController";

/** A source-owned focus loan contains no copied authoring or mutation state. */
export type WorkbookFindFocusLoan = {
  readonly restore: () => boolean;
  /** Recognizes focus returned to this source editor during its own settlement. */
  readonly ownsFocus: (target: EventTarget | null) => boolean;
  readonly settle?: (
    options: GridCellNavigationOptions,
  ) => Promise<GridCellNavigationResult | "accepted">;
};
export type WorkbookFindInput = {
  readonly lifetimeKey: string;
  readonly configurationKey: string;
  readonly browser: WorkbookQueryBrowser | undefined;
  readonly authorization: {
    readonly subscribe: (listener: () => void) => () => void;
    readonly getReadAuthorization: () => boolean;
  };
  readonly readable: boolean;
  readonly stale: boolean;
  readonly readText: (anchor: GridCellAnchor) => readonly string[];
  readonly captureFocus: (
    target?: EventTarget | null,
  ) => WorkbookFindFocusLoan | null;
  readonly beforeFocus?: () => void;
};
const subscribeEmpty = () => () => {};
const emptySnapshot = () => null;

function navigationKey(
  configuration: string,
  accepted: WorkbookViewQueryAccepted | null | undefined,
  presentation: GridPresentationSnapshot | null | undefined,
) {
  return JSON.stringify([
    configuration,
    accepted?.producingRequest.identity?.authorityGeneration,
    accepted?.producingRequest.cursorToken,
    accepted?.canonicalQuery,
    accepted?.rows.map((row) => row.record_id),
    presentation?.revision,
  ]);
}

/** Workbook owns Find attachment, accepted membership and intentional focus lifetime. */
export function useWorkbookFind(input: WorkbookFindInput) {
  const gridRef = useRef<GridHandle | null>(null);
  const [presentationPort, setPresentationPort] =
    useState<GridHandle["presentation"]>();
  const bindGrid = useCallback((handle: GridHandle | null) => {
    gridRef.current = handle;
    setPresentationPort(handle?.presentation);
  }, []);
  const latest = useRef(input);
  latest.current = input;
  const borrowed = useRef<WorkbookFindFocusLoan | null>(null);
  const applyingFocus = useRef<object | null>(null);
  const pendingTabFocus = useRef(false);
  const origin = useRef<GridCellAnchor | null>(null);
  const lastAdmitted = useRef<GridCellAnchor | null>(null);
  const restoreRequest = useRef<AbortController | null>(null);
  const hostRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const controller = useMemo(
    () =>
      new WorkbookFindController(async (anchor, options) => {
        const outcome = await borrowed.current?.settle?.(options);
        if (outcome && outcome !== "accepted") return outcome;
        if (options.signal?.aborted || options.isCurrent?.() === false)
          return "cancelled";
        const token = {};
        let result: GridCellNavigationResult;
        try {
          result =
            (await gridRef.current?.navigateToCell?.(anchor, {
              ...options,
              beforeFocus: () => {
                applyingFocus.current = token;
                latest.current.beforeFocus?.();
                options.beforeFocus?.();
              },
            })) ?? "unavailable";
        } finally {
          if (applyingFocus.current === token) applyingFocus.current = null;
        }
        if (result === "focused") {
          borrowed.current = null;
          lastAdmitted.current = anchor;
        }
        return result;
      }),
    [],
  );
  const snapshot = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const browsing = useSyncExternalStore(
    input.browser?.subscribe ?? subscribeEmpty,
    input.browser?.getSnapshot ?? emptySnapshot,
  );
  const presentation = useSyncExternalStore(
    presentationPort?.subscribe ?? subscribeEmpty,
    presentationPort?.getSnapshot ?? emptySnapshot,
  );
  const authority = useSyncExternalStore(
    input.authorization.subscribe,
    input.authorization.getReadAuthorization,
  );
  const available = input.readable && authority;
  const configuration = input.configurationKey;
  const source = useMemo<WorkbookFindSource | null>(() => {
    if (!available || !presentation) return null;
    const accepted = browsing?.accepted;
    const members = new Set(accepted?.rows.map((row) => row.record_id) ?? []);
    const key = navigationKey(configuration, accepted, presentation);
    const isNavigationCurrent = () =>
      latest.current.readable &&
      latest.current.authorization.getReadAuthorization() &&
      latest.current.lifetimeKey === input.lifetimeKey &&
      key ===
        navigationKey(
          latest.current.configurationKey,
          latest.current.browser?.getSnapshot().accepted,
          presentationPort?.getSnapshot(),
        );
    return {
      lifetimeKey: input.lifetimeKey,
      readable: true,
      navigationKey: key,
      isNavigationCurrent,
      isCurrent: () =>
        isNavigationCurrent() &&
        latest.current.browser?.getSnapshot().accepted === accepted &&
        latest.current.readText === input.readText,
      presentation: {
        ...presentation,
        rowIdentities: presentation.rowIdentities.filter(
          (row) => row.kind === "core_record" && members.has(row.recordId),
        ),
      },
      unavailableReason: accepted ? null : "Loaded rows are unavailable.",
      stale: input.stale || !!browsing?.failure,
      readText: (anchor) => {
        const identity = anchor.rowIdentity;
        if (identity.kind !== "core_record" || !members.has(identity.recordId))
          return [];
        return input.readText(anchor);
      },
    };
  }, [
    available,
    presentation,
    presentationPort,
    browsing?.accepted,
    browsing?.failure,
    configuration,
    input.readText,
    input.lifetimeKey,
    input.stale,
  ]);
  const lifetime = useRef(input.lifetimeKey);
  useLayoutEffect(() => {
    controller.setSource(source);
    if (!controller.getSnapshot().active) lastAdmitted.current = null;
    if (source === null || lifetime.current !== input.lifetimeKey) {
      restoreRequest.current?.abort();
      origin.current = null;
      borrowed.current = null;
      lastAdmitted.current = null;
      applyingFocus.current = null;
      pendingTabFocus.current = false;
    }
    lifetime.current = input.lifetimeKey;
  }, [controller, source, input.lifetimeKey]);
  // Query and Adapter owners invalidate destinations before React catches up.
  useLayoutEffect(() => {
    const invalidate = () => {
      if (source?.isNavigationCurrent?.() === false) {
        controller.cancelNavigation();
        restoreRequest.current?.abort();
      }
    };
    const unsubscribeQuery = input.browser?.subscribe(invalidate);
    const unsubscribePresentation = presentationPort?.subscribe(invalidate);
    return () => {
      unsubscribeQuery?.();
      unsubscribePresentation?.();
    };
  }, [controller, input.browser, presentationPort, source]);
  useLayoutEffect(
    () => () => {
      restoreRequest.current?.abort();
      controller.dispose();
      borrowed.current = null;
      origin.current = null;
      lastAdmitted.current = null;
    },
    [controller],
  );
  // Retire synchronously with the authority owner, before deferred React rendering.
  useLayoutEffect(
    () =>
      input.authorization.subscribe(() => {
        if (!input.authorization.getReadAuthorization()) {
          restoreRequest.current?.abort();
          borrowed.current = null;
          origin.current = null;
          lastAdmitted.current = null;
          controller.setSource(null);
        }
      }),
    [controller, input.authorization],
  );
  const recovery = useWorkbookRecoveryNavigation();
  useLayoutEffect(
    () =>
      recovery?.subscribe(() => {
        if (recovery.getSnapshot().open) {
          restoreRequest.current?.abort();
          controller.collapse();
        }
      }),
    [controller, recovery],
  );
  const capture = useCallback(
    (target?: EventTarget | null) => {
      const editor = latest.current.captureFocus(target);
      if (editor) borrowed.current = editor;
      if (!controller.getSnapshot().active)
        origin.current = gridRef.current?.getActiveCell?.() ?? null;
    },
    [controller],
  );
  const open = useCallback(() => {
    capture();
    restoreRequest.current?.abort();
    pendingTabFocus.current = false;
    controller.open(
      origin.current ?? gridRef.current?.getActiveCell?.() ?? null,
    );
    inputRef.current?.focus();
  }, [capture, controller]);
  useLayoutEffect(() => {
    const keyboard = (event: KeyboardEvent) => {
      if (event.isComposing) return;
      restoreRequest.current?.abort();
      applyingFocus.current = null;
      const own =
        event.target instanceof Node && hostRef.current?.contains(event.target);
      // A Tab inside Find is only departure when the resulting focus leaves it.
      pendingTabFocus.current = own === true && event.key === "Tab";
      if (
        event.target instanceof Node &&
        !hostRef.current?.contains(event.target)
      )
        controller.collapse();
      if (!available) return;
      const decision = decideWorkbookApplicationShortcut(event, {
        capabilities: {
          find: true,
          closeInspector: false,
          history: false,
          linkedEvidence: false,
          quickLink: false,
        },
        focusOwner: own
          ? "find"
          : gridRef.current?.ownsNavigationFocus?.(event.target)
            ? "grid_navigation"
            : "editor",
        rowKind: "none",
        selectionIdentity: null,
        previewableEvidenceCount: 0,
      });
      if (decision.kind !== "open_find") return;
      event.preventDefault();
      event.stopPropagation();
      open();
    };
    const outside = (event: Event) => {
      if (
        event.target instanceof Node &&
        hostRef.current?.contains(event.target)
      ) {
        if (event.type === "focusin") pendingTabFocus.current = false;
        return;
      }
      const userTabDeparture =
        event.type === "focusin" && pendingTabFocus.current;
      pendingTabFocus.current = false;
      if (event.type !== "focusin") {
        restoreRequest.current?.abort();
        applyingFocus.current = null;
      }
      // Source refocus and the admitted semantic destination belong to this move.
      if (
        event.type === "focusin" &&
        !userTabDeparture &&
        controller.getSnapshot().navigating &&
        ((applyingFocus.current !== null &&
          gridRef.current?.ownsNavigationFocus?.(event.target) === true) ||
          borrowed.current?.ownsFocus(event.target) === true)
      )
        return;
      controller.collapse();
    };
    document.addEventListener("keydown", keyboard, true);
    document.addEventListener("pointerdown", outside, true);
    // Let the editor receive its raw input before publishing a Find collapse.
    // This still cancels destinations synchronously within the same event.
    document.addEventListener("input", outside);
    document.addEventListener("compositionstart", outside, true);
    document.addEventListener("focusin", outside, true);
    return () => {
      document.removeEventListener("keydown", keyboard, true);
      document.removeEventListener("pointerdown", outside, true);
      document.removeEventListener("input", outside);
      document.removeEventListener("compositionstart", outside, true);
      document.removeEventListener("focusin", outside, true);
    };
  }, [available, controller, open]);
  const close = useCallback(async () => {
    const current = controller.getSnapshot().current;
    const editor = borrowed.current;
    const moved = current !== null || lastAdmitted.current !== null;
    const anchor = current ?? lastAdmitted.current ?? origin.current;
    controller.close();
    lastAdmitted.current = null;
    borrowed.current = null;
    restoreRequest.current?.abort();
    const abort = new AbortController();
    restoreRequest.current = abort;
    if (!moved && editor) {
      if (editor.restore()) return;
    }
    const handle = gridRef.current;
    const token = {};
    applyingFocus.current = token;
    try {
      if (
        anchor &&
        (await handle?.requestFocus(
          { kind: "cell", anchor },
          { signal: abort.signal },
        )) === "focused"
      )
        return;
      const model = handle?.presentation?.getSnapshot();
      if (
        !abort.signal.aborted &&
        anchor &&
        model &&
        !model.fieldKeys.includes(anchor.fieldKey) &&
        model.rowIdentities.some((row) =>
          gridRowIdentitiesEqual(row, anchor.rowIdentity),
        )
      ) {
        const fieldKey = model.fieldKeys[0];
        if (
          fieldKey &&
          (await handle?.requestFocus(
            { kind: "cell", anchor: { ...anchor, fieldKey } },
            { signal: abort.signal },
          )) === "focused"
        )
          return;
      }
      if (!abort.signal.aborted)
        await handle?.requestFocus({ kind: "root" }, { signal: abort.signal });
    } finally {
      if (applyingFocus.current === token) applyingFocus.current = null;
    }
  }, [controller]);
  const matches = useMemo(
    () => new Set(snapshot.matches.map(gridAnchorKey)),
    [snapshot.matches],
  );
  const cellMatch = useCallback(
    (anchor: GridCellAnchor): "current" | "match" | undefined =>
      snapshot.active && matches.has(gridAnchorKey(anchor))
        ? snapshot.current &&
          gridAnchorKey(snapshot.current) === gridAnchorKey(anchor)
          ? "current"
          : "match"
        : undefined,
    [snapshot.active, snapshot.current, matches],
  );
  return {
    bindGrid,
    isApplyingFocus: () => applyingFocus.current !== null,
    cellMatch,
    control: {
      snapshot,
      available,
      hostRef,
      inputRef,
      capture,
      open,
      close,
      changeTerm: (value: string) => controller.setTerm(value),
      changeCase: (value: boolean) => controller.setMatchCase(value),
      navigate: (direction: 1 | -1) =>
        controller.navigate(direction, gridRef.current?.getActiveCell?.()),
    },
  };
}
