import {
  type GridCellAnchor,
  type GridEditCommitOutcome,
  type GridHandle,
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
import { useWorkbookRecoveryNavigation } from "../../../shared/WorkbookRecoveryBoundary";
import {
  WorkbookFindController,
  type WorkbookFindSource,
} from "../../find/WorkbookFindController";
import { decideWorkbookApplicationShortcut } from "../../policies/workbookApplicationShortcuts";
import type { WorkbookQueryBrowser } from "../../query/WorkbookQueryBrowser";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type {
  TimelineQueueCollectionSave,
  TimelineQueueScalarSave,
} from "../models/timelineControllerPorts";
import {
  timelineCollectionBindings,
  timelineScalarBindings,
} from "../models/timelineFieldRegistry";
import { timelineFindText } from "../models/timelineFindText";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";

const subscribeEmpty = () => () => {};
const emptyPresentation = () => null;

/** Timeline supplies authority, accepted membership, text and source editor settlement. */
export function useTimelineFind(input: {
  runtime: {
    incident: Pick<
      TimelineWorkbookSurfaceRuntime["incident"],
      "id" | "continuityResetKey" | "sheetRef" | "currentRole"
    >;
    query: Pick<TimelineWorkbookSurfaceRuntime["query"], "state">;
    layout: {
      snapshot: Pick<
        TimelineWorkbookSurfaceRuntime["layout"]["snapshot"],
        "state"
      >;
    };
    entities: Pick<TimelineWorkbookSurfaceRuntime["entities"], "index">;
    collaborationProjection: Pick<
      TimelineWorkbookSurfaceRuntime["collaborationProjection"],
      "subscribe" | "getReadAuthorization"
    >;
  };
  browser: WorkbookQueryBrowser;
  rows: readonly WorkbookRow[];
  registry: TimelineEditorDraftRegistry;
  accessLost: boolean;
  stale: boolean;
  interruptViewportContinuity: () => void;
  queueScalarSave: TimelineQueueScalarSave;
  queueCollectionSave: TimelineQueueCollectionSave;
}) {
  const gridRef = useRef<GridHandle | null>(null);
  const [presentationPort, setPresentationPort] =
    useState<GridHandle["presentation"]>();
  const bindGrid = useCallback((handle: GridHandle | null) => {
    gridRef.current = handle;
    if (handle?.presentation) setPresentationPort(handle.presentation);
  }, []);
  const latest = useRef(input);
  latest.current = input;
  const borrowed =
    useRef<ReturnType<TimelineEditorDraftRegistry["activeInput"]>>(null);
  const origin = useRef<GridCellAnchor | null>(null);
  const lastAdmitted = useRef<GridCellAnchor | null>(null);
  const restoreRequest = useRef<AbortController | null>(null);
  const sourceSettlement = useRef<{
    key: string;
    promise: Promise<GridEditCommitOutcome>;
  } | null>(null);
  const hostRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const controller = useMemo(
    () =>
      new WorkbookFindController(async (anchor, options) => {
        const current = latest.current;
        const editor = borrowed.current;
        const row =
          editor &&
          current.rows.find(
            (row) => row.key === current.registry.resolveRowKey(editor.rowKey),
          );
        // Recordless authoring stays retained; Find must never be its creation trigger.
        if (editor && row?.recordId) {
          const collection = timelineCollectionBindings.find(
            (binding) => binding.draftKey === editor.field,
          );
          const scalar = timelineScalarBindings.find(
            (binding) => binding.key === editor.field,
          );
          if (collection || (scalar && editor.surface === "inspector")) {
            const element = current.registry.inputElementForFocusKey(
              editor.focusKey,
            );
            const key = `${editor.focusKey}:${current.registry.revisionForFocusKey(editor.focusKey)}`;
            if (sourceSettlement.current?.key !== key) {
              const promise = new Promise<GridEditCommitOutcome>((resolve) => {
                if (collection) {
                  const value =
                    current.registry.draftValue(editor) ?? element?.value ?? "";
                  if (value === "") {
                    resolve({ kind: "accepted" });
                    return;
                  }
                  current.queueCollectionSave(
                    editor.rowKey,
                    collection.fieldKey,
                    collection.draftKey,
                    value,
                    editor.surface,
                    resolve,
                  );
                } else if (scalar)
                  current.queueScalarSave(
                    editor.rowKey,
                    scalar.key,
                    {
                      continueOnFreshDraft: false,
                      preserveInputFocus: false,
                      surface: editor.surface,
                    },
                    current.registry.draftValue(editor) ?? element?.value,
                    resolve,
                  );
              }).finally(() => {
                if (sourceSettlement.current?.promise === promise)
                  sourceSettlement.current = null;
              });
              sourceSettlement.current = { key, promise };
            }
            const outcome = await new Promise<GridEditCommitOutcome | null>(
              (resolve) => {
                const abort = () => resolve(null);
                if (options.signal?.aborted) return abort();
                options.signal?.addEventListener("abort", abort, {
                  once: true,
                });
                void sourceSettlement.current?.promise.then((result) => {
                  options.signal?.removeEventListener("abort", abort);
                  resolve(result);
                });
              },
            );
            if (
              !outcome ||
              options.signal?.aborted ||
              options.isCurrent?.() === false
            )
              return "cancelled";
            if (outcome.kind !== "accepted") {
              current.registry
                .inputElementForFocusKey(editor.focusKey)
                ?.focus();
              return "rejected";
            }
          }
        }
        if (options.signal?.aborted || options.isCurrent?.() === false)
          return "cancelled";
        const result =
          (await gridRef.current?.navigateToCell?.(anchor, {
            ...options,
            beforeFocus: () => {
              latest.current.interruptViewportContinuity();
              options.beforeFocus?.();
            },
          })) ?? "unavailable";
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
    input.browser.subscribe,
    input.browser.getSnapshot,
  );
  const presentation = useSyncExternalStore(
    presentationPort?.subscribe ?? subscribeEmpty,
    presentationPort?.getSnapshot ?? emptyPresentation,
  );
  const authority = useSyncExternalStore(
    input.runtime.collaborationProjection.subscribe,
    input.runtime.collaborationProjection.getReadAuthorization,
  );
  const available =
    !input.accessLost && authority && !!input.runtime.incident.currentRole;
  const configuration = JSON.stringify([
    input.runtime.incident.id,
    input.runtime.incident.continuityResetKey,
    input.runtime.incident.sheetRef,
    input.runtime.query.state,
    input.runtime.layout.snapshot.state,
  ]);
  const source = useMemo<WorkbookFindSource | null>(() => {
    if (!available || !presentation) return null;
    const accepted = browsing.accepted;
    const members = new Set(accepted?.rows.map((row) => row.record_id) ?? []);
    return {
      lifetimeKey: input.runtime.incident.id,
      readable: true,
      navigationKey: JSON.stringify([
        configuration,
        accepted?.producingRequest.identity?.authorityGeneration,
        accepted?.producingRequest.cursorToken,
        accepted?.canonicalQuery,
        [...members],
        presentation.revision,
      ]),
      presentation: {
        ...presentation,
        rowIdentities: presentation.rowIdentities.filter(
          (row) => row.kind === "core_record" && members.has(row.recordId),
        ),
      },
      unavailableReason: accepted
        ? null
        : "Loaded Timeline rows are unavailable.",
      stale: input.stale || browsing.failure !== null,
      readText: (anchor) => {
        const identity = anchor.rowIdentity;
        if (identity.kind !== "core_record" || !members.has(identity.recordId))
          return [];
        const row = input.rows.find(
          (row) => row.recordId === identity.recordId,
        );
        return row
          ? timelineFindText(row, anchor.fieldKey, input.runtime.entities.index)
          : [];
      },
    };
  }, [
    available,
    presentation,
    browsing.accepted,
    browsing.failure,
    configuration,
    input.rows,
    input.runtime.entities.index,
    input.runtime.incident.id,
    input.stale,
  ]);
  useLayoutEffect(() => {
    controller.setSource(source);
    if (!controller.getSnapshot().active) lastAdmitted.current = null;
    if (source === null) {
      restoreRequest.current?.abort();
      origin.current = null;
      borrowed.current = null;
    }
  }, [controller, source]);
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
      input.runtime.collaborationProjection.subscribe(() => {
        if (!input.runtime.collaborationProjection.getReadAuthorization()) {
          restoreRequest.current?.abort();
          borrowed.current = null;
          origin.current = null;
          lastAdmitted.current = null;
          controller.setSource(null);
        }
      }),
    [controller, input.runtime.collaborationProjection],
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
      const editor = latest.current.registry.activeInput(target);
      if (editor) borrowed.current = editor;
      if (!controller.getSnapshot().active)
        origin.current = gridRef.current?.getActiveCell?.() ?? null;
    },
    [controller],
  );
  const open = useCallback(() => {
    capture();
    restoreRequest.current?.abort();
    controller.open(
      origin.current ?? gridRef.current?.getActiveCell?.() ?? null,
    );
    inputRef.current?.focus();
  }, [capture, controller]);
  useLayoutEffect(() => {
    const keyboard = (event: KeyboardEvent) => {
      if (event.isComposing) return;
      restoreRequest.current?.abort();
      if (
        event.target instanceof Node &&
        !hostRef.current?.contains(event.target)
      )
        controller.collapse();
      if (!available) return;
      const own =
        event.target instanceof Node && hostRef.current?.contains(event.target);
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
      )
        return;
      if (event.type === "pointerdown") restoreRequest.current?.abort();
      if (event.type === "focusin" && controller.getSnapshot().navigating)
        return;
      controller.collapse();
    };
    document.addEventListener("keydown", keyboard, true);
    document.addEventListener("pointerdown", outside, true);
    document.addEventListener("focusin", outside);
    return () => {
      document.removeEventListener("keydown", keyboard, true);
      document.removeEventListener("pointerdown", outside, true);
      document.removeEventListener("focusin", outside);
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
      const element = latest.current.registry.inputElementForFocusKey(
        editor.focusKey,
      );
      if (element) {
        element.focus();
        return;
      }
    }
    const handle = gridRef.current;
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
