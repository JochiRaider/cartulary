import type {
  GridCellAnchor,
  GridFocusResult,
  GridFocusTarget,
  GridHandle,
} from "@cartulary/grid-adapter";
import {
  workbookIncidentIdentityTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import { useCallback, useLayoutEffect, useRef } from "react";
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type ScrollPosition,
  type ViewportSnapshot,
} from "../../continuity/gridViewportContinuity";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelineAcceptedContinuity } from "../models/timelineAcceptedMutationEffects";
import {
  timelineCollectionBindings,
  timelineScalarBindings,
} from "../models/timelineFieldRegistry";
import {
  advanceTimelineContinuityRender,
  beginTimelineContinuityLifecycle,
  requireTimelineSourceRecord,
  settleTimelineContinuityRequirement,
  type TimelineContinuityLifecycle,
  type TimelineContinuityRequirementName,
  type TimelineContinuitySemanticTarget,
  type TimelineSourceRecordEvidence,
  type TimelineSourceRecordRequirement,
  timelineContinuityRequirementsSettled,
  transitionTimelineContinuity,
} from "../models/timelineViewportContinuityModel";

type TimelineMutableRef<T> = { current: T };
export type TimelineViewportContinuityTarget = TimelineContinuitySemanticTarget;
export type TimelineViewportContinuityRequest = {
  token: number;
  lifecycle: TimelineContinuityLifecycle;
  preservedViewport: ViewportSnapshot | null;
};

/** Accepted presentation identity, with live authority independently observable. */
export type TimelineViewportContinuityScope = {
  readonly getSnapshot: () => {
    readonly key: string;
    readonly readable: boolean;
  };
  readonly subscribe: (listener: () => void) => () => void;
};

type Restoration = {
  request: TimelineViewportContinuityRequest;
  readonly scopeKey: string;
  readonly controller: AbortController;
  cancelPass: (() => void) | null;
  destination: GridFocusTarget | HTMLElement | null;
};

const synopsis = "timeline.activity_synopsis_text";
function timelineAnchor(recordId: string, fieldKey: string): GridCellAnchor {
  return {
    fieldKey,
    rowIdentity: { kind: "core_record", recordId },
    surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
  };
}

export function useTimelineViewportContinuityController({
  gridHandleRef,
  gridShellRef,
  editorDraftRegistry,
  scope,
  setViewportContinuityRequest,
  viewportContinuityRequest,
  viewportContinuityTokenRef,
}: {
  readonly gridHandleRef: TimelineMutableRef<GridHandle | null>;
  readonly gridShellRef: TimelineMutableRef<HTMLDivElement | null>;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly scope: TimelineViewportContinuityScope;
  readonly setViewportContinuityRequest: (
    value:
      | TimelineViewportContinuityRequest
      | null
      | ((
          current: TimelineViewportContinuityRequest | null,
        ) => TimelineViewportContinuityRequest | null),
  ) => void;
  readonly viewportContinuityRequest: TimelineViewportContinuityRequest | null;
  readonly viewportContinuityTokenRef: TimelineMutableRef<number>;
}) {
  const active = useRef<Restoration | null>(null);
  const interactionGeneration = useRef(0);
  const mounted = useRef(true);
  const isCurrent = useCallback(
    (restoration: Restoration) => {
      const currentScope = scope.getSnapshot();
      return (
        mounted.current &&
        active.current === restoration &&
        !restoration.controller.signal.aborted &&
        currentScope.readable &&
        currentScope.key === restoration.scopeKey &&
        restoration.request.lifecycle.userInterruptionGeneration ===
          interactionGeneration.current
      );
    },
    [scope],
  );

  const clearViewportContinuity = useCallback(
    (token: number) => {
      const restoration = active.current;
      if (restoration?.request.token !== token) return;
      active.current = null;
      restoration.controller.abort();
      restoration.cancelPass?.();
      if (mounted.current)
        setViewportContinuityRequest((current) =>
          current?.token === token ? null : current,
        );
    },
    [setViewportContinuityRequest],
  );

  const interruptViewportContinuity = useCallback(() => {
    interactionGeneration.current += 1;
    const restoration = active.current;
    if (restoration) clearViewportContinuity(restoration.request.token);
  }, [clearViewportContinuity]);

  const resolveInputElement = useCallback(
    (focusKey: string) => {
      const input = editorDraftRegistry.inputElementForFocusKey(focusKey);
      return input?.closest("[inert]") ? null : input;
    },
    [editorDraftRegistry],
  );

  useLayoutEffect(() => {
    mounted.current = true;
    const focusChanged = (event: FocusEvent) => {
      const restoration = active.current;
      if (!restoration) return;
      const target = restoration.request.lifecycle.semanticFocusTarget;
      if (
        target.kind === "input" &&
        event.target === resolveInputElement(target.focusKey)
      )
        return;
      if (
        isRestorationFocus(
          event.target,
          restoration.destination,
          gridHandleRef.current,
        )
      )
        return;
      // Removing an editor naturally leaves body active without a focusin event.
      // An actual focus transition to another control is always a new owner.
      interruptViewportContinuity();
    };
    const invalidateScope = () => {
      const restoration = active.current;
      if (restoration && !isCurrent(restoration))
        clearViewportContinuity(restoration.request.token);
    };
    const unsubscribe = scope.subscribe(invalidateScope);
    const events = ["pointerdown", "keydown", "wheel", "input"] as const;
    for (const event of events)
      document.addEventListener(event, interruptViewportContinuity, true);
    document.addEventListener("focusin", focusChanged, true);
    return () => {
      unsubscribe();
      for (const event of events)
        document.removeEventListener(event, interruptViewportContinuity, true);
      document.removeEventListener("focusin", focusChanged, true);
      mounted.current = false;
      const restoration = active.current;
      if (restoration) clearViewportContinuity(restoration.request.token);
    };
  }, [
    clearViewportContinuity,
    gridHandleRef,
    interruptViewportContinuity,
    isCurrent,
    resolveInputElement,
    scope,
  ]);

  // The scope getter is updated by composition during render, before any layout
  // effect can start an old destination under the newly accepted presentation.
  useLayoutEffect(() => {
    const restoration = active.current;
    if (restoration && !isCurrent(restoration))
      clearViewportContinuity(restoration.request.token);
  });

  const currentGridScrollSnapshot = useCallback((): ScrollPosition | null => {
    const element = gridHandleRef.current?.getScrollElement();
    return element
      ? { top: element.scrollTop, left: element.scrollLeft }
      : null;
  }, [gridHandleRef]);
  const resolveRect = useCallback(
    (target: TimelineViewportContinuityTarget) => {
      if (target.kind === "input")
        return (
          resolveInputElement(target.focusKey)?.getBoundingClientRect() ?? null
        );
      if (target.kind === "row-inspect")
        return (
          gridHandleRef.current?.getAnchorRect(
            timelineAnchor(target.recordId, synopsis),
          ) ?? null
        );
      return null;
    },
    [gridHandleRef, resolveInputElement],
  );
  const currentGridViewportSnapshot = useCallback(
    (rect: DOMRectReadOnly | null = null): ViewportSnapshot | null => {
      const element = gridHandleRef.current?.getScrollElement();
      const scroll = currentGridScrollSnapshot();
      if (!gridShellRef.current || !element || !scroll) return null;
      const container = element.getBoundingClientRect();
      return {
        scroll,
        anchor:
          rect && isRectFullyVisibleWithinContainer(container, rect)
            ? captureViewportAnchor(container, rect)
            : null,
      };
    },
    [currentGridScrollSnapshot, gridHandleRef, gridShellRef],
  );

  const beginViewportContinuity = useCallback(
    (
      target: TimelineViewportContinuityTarget,
      options: {
        requirements?: readonly TimelineContinuityRequirementName[];
      } = {},
    ) => {
      const previous = active.current;
      if (target.kind === "scroll-only" && previous && isCurrent(previous))
        return previous.request.token;
      if (previous) clearViewportContinuity(previous.request.token);
      const token = viewportContinuityTokenRef.current++;
      const currentScope = scope.getSnapshot();
      if (!mounted.current || !currentScope.readable) return token;
      const request: TimelineViewportContinuityRequest = {
        token,
        lifecycle: beginTimelineContinuityLifecycle({
          semanticFocusTarget: target,
          userInterruptionGeneration: interactionGeneration.current,
          ...options,
        }),
        preservedViewport: currentGridViewportSnapshot(resolveRect(target)),
      };
      active.current = {
        request,
        scopeKey: currentScope.key,
        controller: new AbortController(),
        cancelPass: null,
        destination: null,
      };
      setViewportContinuityRequest(request);
      return token;
    },
    [
      clearViewportContinuity,
      currentGridViewportSnapshot,
      isCurrent,
      resolveRect,
      scope,
      setViewportContinuityRequest,
      viewportContinuityTokenRef,
    ],
  );

  const updateRequest = useCallback(
    (
      token: number | undefined,
      update: (
        request: TimelineViewportContinuityRequest,
      ) => TimelineViewportContinuityRequest,
    ) => {
      const restoration = active.current;
      if (!restoration || restoration.request.token !== token) return;
      if (!isCurrent(restoration)) {
        clearViewportContinuity(restoration.request.token);
        return;
      }
      restoration.cancelPass?.();
      restoration.cancelPass = null;
      restoration.request = update(restoration.request);
      setViewportContinuityRequest(restoration.request);
    },
    [clearViewportContinuity, isCurrent, setViewportContinuityRequest],
  );

  const advanceViewportContinuity = useCallback(
    (
      token: number | undefined,
      options: {
        sourceRecord?: TimelineSourceRecordEvidence;
        target?: TimelineViewportContinuityTarget | null;
      } = {},
    ) => {
      updateRequest(token, (request) => ({
        ...request,
        lifecycle: advanceTimelineContinuityRender(
          {
            ...request.lifecycle,
            semanticFocusTarget:
              options.target ?? request.lifecycle.semanticFocusTarget,
          },
          { sourceRecord: options.sourceRecord },
        ),
      }));
    },
    [updateRequest],
  );
  const completeAcceptedViewportContinuity = useCallback(
    (token: number | undefined, continuity: TimelineAcceptedContinuity) => {
      if (continuity.kind === "advance") {
        advanceViewportContinuity(token, { target: continuity.target });
      } else {
        updateRequest(token, (request) => ({
          ...request,
          // Creation continues at the trailing draft, never at pre-create scroll.
          preservedViewport: null,
          lifecycle: advanceTimelineContinuityRender({
            ...request.lifecycle,
            semanticFocusTarget: {
              kind: "input",
              focusKey: continuity.focusKey,
            },
          }),
        }));
      }
    },
    [advanceViewportContinuity, updateRequest],
  );
  const failViewportContinuity = useCallback(
    (token: number) => {
      updateRequest(token, (request) => ({
        ...request,
        lifecycle: transitionTimelineContinuity(request.lifecycle, "failed"),
      }));
    },
    [updateRequest],
  );
  const requireViewportContinuitySourceRecord = useCallback(
    (token: number, requirement: TimelineSourceRecordRequirement) => {
      updateRequest(token, (request) => ({
        ...request,
        lifecycle: requireTimelineSourceRecord(request.lifecycle, requirement),
      }));
    },
    [updateRequest],
  );
  const settleViewportContinuityFollowUp = useCallback(
    (
      token: number,
      requirement: TimelineContinuityRequirementName,
      state: "settled" | "terminal",
    ) => {
      updateRequest(token, (request) => ({
        ...request,
        lifecycle: settleTimelineContinuityRequirement(
          request.lifecycle,
          requirement,
          state,
        ),
      }));
    },
    [updateRequest],
  );

  useLayoutEffect(() => {
    const restoration = active.current;
    const request = viewportContinuityRequest;
    if (
      !restoration ||
      restoration.request !== request ||
      !request ||
      request.lifecycle.renderGeneration < 1 ||
      !isCurrent(restoration)
    )
      return;
    const controller = new AbortController();
    let timer: number | undefined;
    let frame: number | undefined;
    const cancelPass = () => {
      controller.abort();
      if (timer !== undefined) window.clearTimeout(timer);
      if (frame !== undefined) window.cancelAnimationFrame(frame);
    };
    restoration.cancelPass = cancelPass;
    restoration.controller.signal.addEventListener("abort", cancelPass, {
      once: true,
    });
    const current = () =>
      isCurrent(restoration) &&
      restoration.request === request &&
      !controller.signal.aborted;
    const pendingRequirements = () =>
      !timelineContinuityRequirementsSettled(request.lifecycle);
    const setScroll = (scroll: ScrollPosition | null) => {
      const element = gridHandleRef.current?.getScrollElement();
      if (!current() || !element || !scroll) return;
      element.scrollLeft = scroll.left;
      if (current()) element.scrollTop = scroll.top;
    };
    const semanticFocus = async (
      destination: GridFocusTarget,
    ): Promise<GridFocusResult> => {
      if (!current()) return "cancelled";
      restoration.destination = destination;
      const result =
        (await gridHandleRef.current?.requestFocus(destination, {
          signal: controller.signal,
          preserveSelection: true,
        })) ?? "unavailable";
      return current() ? result : "cancelled";
    };
    const target = request.lifecycle.semanticFocusTarget;
    let rect = () => resolveRect(target);
    let needsTargetGeometry = target.kind !== "scroll-only";
    const fallback = async (
      recordId?: string,
      fieldKey = synopsis,
    ): Promise<GridFocusResult> => {
      if (!current()) return "cancelled";
      const grid = gridHandleRef.current;
      const presentation = grid?.presentation?.getSnapshot();
      const rowExists =
        recordId &&
        (!presentation ||
          presentation.rowIdentities.some(
            (row) => row.kind === "core_record" && row.recordId === recordId,
          ));
      const field =
        !presentation || presentation.fieldKeys.includes(fieldKey)
          ? fieldKey
          : presentation.fieldKeys[0];
      if (rowExists && field) {
        const anchor = timelineAnchor(recordId, field);
        const result = await semanticFocus({ kind: "cell", anchor });
        if (result !== "unavailable") {
          rect = () => gridHandleRef.current?.getAnchorRect(anchor) ?? null;
          return result;
        }
      }
      if (!current()) return "cancelled";
      // Source projection may still be arriving. Do not turn temporary absence
      // into permanent removal or poll an adapter-declared unavailable target.
      if (pendingRequirements()) return "unavailable";
      const result = await semanticFocus({ kind: "root" });
      rect = () => null;
      needsTargetGeometry = false;
      if (result !== "unavailable" || !current()) return result;
      for (const element of shellFocusCandidates()) {
        if (!current()) return "cancelled";
        restoration.destination = element;
        element.focus({ preventScroll: true });
        if (current() && document.activeElement === element) return "focused";
      }
      return "unavailable";
    };
    const inputIdentity =
      target.kind === "input"
        ? resolveInputIdentity(target.focusKey, editorDraftRegistry)
        : null;
    // A detached registered input has no mount notification. Only this path and
    // transient layout geometry retain the prior bounded 50ms / 60 retry budget.
    const focus = async (
      attempt: number,
    ): Promise<GridFocusResult | "retry"> => {
      if (!current()) return "cancelled";
      if (target.kind === "scroll-only") return "focused";
      if (target.kind === "row-inspect") return fallback(target.recordId);
      if (inputIdentity?.draft && inputIdentity.surface === "grid") {
        const result = await semanticFocus({
          kind: "draft",
          fieldKey: inputIdentity.fieldKey,
        });
        return result === "unavailable" ? fallback() : result;
      }
      const input = resolveInputElement(target.focusKey);
      if (input) {
        restoration.destination = input;
        if (document.activeElement !== input)
          input.focus({ preventScroll: true });
        return current() && document.activeElement === input
          ? "focused"
          : "cancelled";
      }
      if (attempt < 60) return "retry";
      return fallback(inputIdentity?.recordId, inputIdentity?.fieldKey);
    };
    const viewport = request.preservedViewport?.anchor
      ? request.preservedViewport
      : { scroll: currentGridScrollSnapshot(), anchor: null };
    const geometry = () => {
      if (!current()) return false;
      if (target.kind === "scroll-only") {
        setScroll(request.preservedViewport?.scroll ?? null);
        return true;
      }
      const element = gridHandleRef.current?.getScrollElement();
      if (!needsTargetGeometry) return true;
      if (!element || !viewport.scroll) return false;
      // Inspector inputs own their own geometry; their focus must not make the
      // workbook scroll to a rectangle outside its scrollport.
      const input =
        target.kind === "input" ? resolveInputElement(target.focusKey) : null;
      if (input && !element.contains(input)) {
        setScroll(viewport.scroll);
        return current();
      }
      element.scrollLeft = viewport.scroll.left;
      if (!current()) return false;
      const targetRect = rect();
      if (!targetRect) return false;
      setScroll(
        computeRestoredViewportScroll({
          preservedScroll: viewport.scroll,
          currentScroll: { top: element.scrollTop, left: element.scrollLeft },
          preservedAnchor: viewport.anchor,
          containerRect: element.getBoundingClientRect(),
          elementRect: targetRect,
        }),
      );
      const updatedRect = rect();
      return (
        current() &&
        !!updatedRect &&
        isRectFullyVisibleWithinContainer(
          element.getBoundingClientRect(),
          updatedRect,
        )
      );
    };
    const finish = () => {
      if (current() && !pendingRequirements())
        clearViewportContinuity(request.token);
    };
    const retry = (attempt: number) => {
      if (!current()) return;
      timer = window.setTimeout(() => {
        void restore(attempt + 1);
      }, 50);
    };
    let focusedElement: Element | null = null;
    const restore = async (attempt: number) => {
      // Geometry can lag a successful semantic focus. Reuse that result while
      // its element stays mounted; only actual detachment needs focus recovery.
      if (!focusedElement?.isConnected) {
        const result = await focus(attempt);
        if (!current()) return;
        if (result === "cancelled") {
          clearViewportContinuity(request.token);
          return;
        }
        if (result === "retry") return retry(attempt);
        if (result === "unavailable") return finish();
        focusedElement = document.activeElement;
      }
      if (!geometry() && attempt < 60) return retry(attempt);
      if (!current()) return;
      // One stabilization frame per rendered projection. Readiness is already
      // owned by the adapter; do not issue another reveal/focus on every frame.
      frame = window.requestAnimationFrame(() => {
        if (!current()) return;
        if (!geometry() && attempt < 60) return retry(attempt);
        finish();
      });
    };
    void restore(0);
    return () => {
      cancelPass();
      restoration.controller.signal.removeEventListener("abort", cancelPass);
      if (restoration.cancelPass === cancelPass) restoration.cancelPass = null;
    };
  }, [
    currentGridScrollSnapshot,
    clearViewportContinuity,
    editorDraftRegistry,
    gridHandleRef,
    isCurrent,
    resolveInputElement,
    resolveRect,
    viewportContinuityRequest,
  ]);

  return {
    commands: {
      advanceViewportContinuity,
      beginViewportContinuity,
      clearViewportContinuity,
      completeAcceptedViewportContinuity,
      interruptViewportContinuity,
      failViewportContinuity,
      requireViewportContinuitySourceRecord,
      settleViewportContinuityFollowUp,
    },
  };
}

function resolveInputIdentity(
  focusKey: string,
  registry: TimelineEditorDraftRegistry,
) {
  const [localRowKey, field, surface] = focusKey.split(":");
  if (!localRowKey) return null;
  const rowKey = registry.resolveRowKey(localRowKey);
  const binding = [
    ...timelineScalarBindings,
    ...timelineCollectionBindings,
  ].find(
    (binding) => ("key" in binding ? binding.key : binding.draftKey) === field,
  );
  if (!binding) return null;
  return {
    draft: rowKey.startsWith("draft-"),
    recordId: rowKey.startsWith("draft-") ? undefined : rowKey,
    fieldKey: binding.fieldKey,
    surface: surface ?? "grid",
  };
}

function isRestorationFocus(
  target: EventTarget | null,
  destination: Restoration["destination"],
  grid: GridHandle | null,
) {
  if (destination instanceof HTMLElement) return target === destination;
  if (!(target instanceof Element) || !destination) return false;
  if (destination.kind === "root") return target === grid?.getScrollElement();
  if (
    destination.kind !== "cell" ||
    !grid?.getScrollElement()?.contains(target) ||
    target.closest("[role='gridcell']") !== target
  )
    return false;
  const cell =
    target.closest<HTMLElement>("[data-grid-field-key]") ??
    target
      .closest("[role='gridcell']")
      ?.querySelector<HTMLElement>("[data-grid-field-key]");
  return (
    cell?.dataset.gridFieldKey === destination.anchor.fieldKey &&
    destination.anchor.rowIdentity.kind === "core_record" &&
    target.closest<HTMLElement>("[data-grid-record-id]")?.dataset
      .gridRecordId === destination.anchor.rowIdentity.recordId
  );
}

function shellFocusCandidates() {
  const candidates = [
    document.querySelector<HTMLElement>(
      '[aria-label="Built-in workbook surfaces"] button[aria-current="page"]',
    ),
  ];
  for (const id of [
    workbookSurfacesMenuTriggerTestId(),
    workbookIncidentIdentityTestId(),
  ]) {
    const container = document.querySelector<HTMLElement>(
      `[data-testid="${id}"]`,
    );
    candidates.push(
      container?.matches("button, [tabindex]")
        ? container
        : (container?.querySelector<HTMLElement>("button, [tabindex]") ?? null),
    );
  }
  return candidates.filter(
    (element): element is HTMLElement =>
      !!element &&
      element.getClientRects().length > 0 &&
      !element.matches(":disabled, [aria-disabled='true']") &&
      !element.closest("[inert], [hidden], [aria-hidden='true']"),
  );
}
