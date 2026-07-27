import type { GridCellAnchor, GridHandle } from "@cartulary/grid-adapter";
import {
  dataTestIdSelector,
  draftCellTestId,
  gridRowGutterTestId,
  rowCellTestId,
} from "@cartulary/ui-contracts";
import { useCallback, useEffect, useLayoutEffect, useRef } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type ScrollPosition,
  type ViewportSnapshot,
} from "../../utils/workbookContinuity";
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
import { timelineScalarBindings } from "../models/workbookTimelineModel";

type TimelineMutableRef<T> = {
  current: T;
};

export type TimelineViewportContinuityTarget = TimelineContinuitySemanticTarget;

export type TimelineViewportContinuityRequest = {
  token: number;
  lifecycle: TimelineContinuityLifecycle;
  preservedViewport: ViewportSnapshot | null;
};

export function useTimelineViewportContinuityController({
  gridHandleRef,
  gridShellRef,
  rowInputRefs,
  rowInputTestIdsRef,
  setViewportContinuityRequest,
  viewportContinuityRequest,
  viewportContinuityTokenRef,
}: {
  readonly gridHandleRef: TimelineMutableRef<GridHandle | null>;
  readonly gridShellRef: TimelineMutableRef<HTMLDivElement | null>;
  readonly rowInputRefs: TimelineMutableRef<
    Map<string, HTMLInputElement | HTMLTextAreaElement>
  >;
  readonly rowInputTestIdsRef: TimelineMutableRef<Map<string, string>>;
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
  const activeViewportContinuityRequestRef =
    useRef<TimelineViewportContinuityRequest | null>(viewportContinuityRequest);
  const userInteractionVersionRef = useRef(0);

  useEffect(() => {
    const recordUserInteraction = () => {
      userInteractionVersionRef.current += 1;
    };
    document.addEventListener("keydown", recordUserInteraction, true);
    document.addEventListener("pointerdown", recordUserInteraction, true);
    document.addEventListener("wheel", recordUserInteraction, true);
    return () => {
      document.removeEventListener("keydown", recordUserInteraction, true);
      document.removeEventListener("pointerdown", recordUserInteraction, true);
      document.removeEventListener("wheel", recordUserInteraction, true);
    };
  }, []);

  const currentGridScrollElement = useCallback(
    () => gridHandleRef.current?.getScrollElement() ?? null,
    [gridHandleRef],
  );

  const currentGridScrollSnapshot = useCallback(() => {
    const scrollElement = currentGridScrollElement();
    if (scrollElement === null) return null;
    return {
      top: scrollElement.scrollTop,
      left: scrollElement.scrollLeft,
    };
  }, [currentGridScrollElement]);

  const resolveInputElement = useCallback(
    (focusKey: string) => {
      const selectorTestId = rowInputTestIdsRef.current.get(focusKey) ?? null;
      const selector =
        selectorTestId === null
          ? null
          : document.querySelector<HTMLInputElement | HTMLTextAreaElement>(
              dataTestIdSelector(selectorTestId),
            );
      if (selector !== null) {
        return selector;
      }
      const [rowKey, fieldKey, surface] = focusKey.split(":");
      const scalarBinding = timelineScalarBindings.find(
        (binding) => binding.key === fieldKey,
      );
      if (
        rowKey !== undefined &&
        surface === "grid" &&
        scalarBinding !== undefined
      ) {
        const fallbackTestId = rowKey.startsWith("draft-")
          ? draftCellTestId(scalarBinding.fieldKey)
          : rowCellTestId(rowKey, scalarBinding.fieldKey);
        const fallback = document.querySelector<
          HTMLInputElement | HTMLTextAreaElement
        >(dataTestIdSelector(fallbackTestId));
        if (fallback !== null) {
          return fallback;
        }
      }
      return rowInputRefs.current.get(focusKey) ?? null;
    },
    [rowInputRefs, rowInputTestIdsRef],
  );

  const resolveViewportContinuityElement = useCallback(
    (target: TimelineViewportContinuityTarget) => {
      switch (target.kind) {
        case "row-inspect": {
          const content = document.querySelector<HTMLElement>(
            dataTestIdSelector(
              rowCellTestId(target.recordId, "timeline.activity_synopsis_text"),
            ),
          );
          return (
            content?.closest<HTMLElement>('[role="gridcell"]') ??
            content ??
            document.querySelector<HTMLElement>(
              dataTestIdSelector(
                gridRowGutterTestId(timelineViewSchemaId, target.recordId),
              ),
            )
          );
        }
        case "input":
          return resolveInputElement(target.focusKey);
        case "scroll-only":
          return null;
      }
    },
    [resolveInputElement],
  );

  const currentGridViewportSnapshot = useCallback(
    (target: HTMLElement | null = null): ViewportSnapshot | null => {
      const gridShell = gridShellRef.current;
      const scroll = currentGridScrollSnapshot();
      const scrollElement = currentGridScrollElement();
      if (gridShell === null || scroll === null || scrollElement === null) {
        return null;
      }
      const containerRect = scrollElement.getBoundingClientRect();
      const targetRect = target?.getBoundingClientRect() ?? null;
      return {
        scroll,
        anchor:
          targetRect === null ||
          !isRectFullyVisibleWithinContainer(containerRect, targetRect)
            ? null
            : captureViewportAnchor(containerRect, targetRect),
      };
    },
    [currentGridScrollElement, currentGridScrollSnapshot, gridShellRef],
  );

  const restoreGridScroll = useCallback(
    (preservedScroll: ScrollPosition | null) => {
      const scrollElement = currentGridScrollElement();
      if (scrollElement === null || preservedScroll === null) {
        return;
      }
      scrollElement.scrollTop = preservedScroll.top;
      scrollElement.scrollLeft = preservedScroll.left;
      window.requestAnimationFrame(() => {
        const currentScrollElement = currentGridScrollElement();
        if (currentScrollElement === null) return;
        currentScrollElement.scrollTop = preservedScroll.top;
        currentScrollElement.scrollLeft = preservedScroll.left;
      });
    },
    [currentGridScrollElement],
  );

  const restoreGridViewportForElement = useCallback(
    (
      resolveElement: () => HTMLElement | null,
      preservedViewport: ViewportSnapshot | null,
    ) => {
      const currentViewport =
        preservedViewport?.anchor === null
          ? ({
              scroll: currentGridScrollSnapshot(),
              anchor: null,
            } satisfies ViewportSnapshot)
          : (preservedViewport ??
            ({
              scroll: currentGridScrollSnapshot(),
              anchor: null,
            } satisfies ViewportSnapshot));
      const preservedScroll = currentViewport.scroll;
      const focusResolvedElement = () => {
        const element = resolveElement();
        if (element === null || !element.isConnected) {
          return false;
        }
        if (!element.hasAttribute("tabindex")) {
          element.tabIndex = -1;
        }
        element.focus({ preventScroll: true });
        return document.activeElement === element;
      };
      window.focus();
      const focusedNow = focusResolvedElement();
      restoreGridScroll(preservedScroll);
      const restoreViewportGeometryNow = () => {
        const scrollElement = currentGridScrollElement();
        const currentElement = resolveElement();
        if (
          scrollElement === null ||
          preservedScroll === null ||
          currentElement === null ||
          !currentElement.isConnected
        ) {
          return false;
        }
        const restoredScroll = computeRestoredViewportScroll({
          preservedScroll,
          currentScroll: {
            top: scrollElement.scrollTop,
            left: scrollElement.scrollLeft,
          },
          preservedAnchor: currentViewport.anchor,
          containerRect: scrollElement.getBoundingClientRect(),
          elementRect: currentElement.getBoundingClientRect(),
        });
        restoreGridScroll(restoredScroll);
        const updatedScrollElement = currentGridScrollElement();
        const updatedElement = resolveElement();
        if (
          updatedScrollElement === null ||
          updatedElement === null ||
          !updatedElement.isConnected
        ) {
          return false;
        }
        const fullyVisible = isRectFullyVisibleWithinContainer(
          updatedScrollElement.getBoundingClientRect(),
          updatedElement.getBoundingClientRect(),
        );
        return focusResolvedElement() && fullyVisible;
      };
      const restoredNow = restoreViewportGeometryNow();
      return focusedNow && restoredNow;
    },
    [currentGridScrollElement, currentGridScrollSnapshot, restoreGridScroll],
  );

  const beginViewportContinuity = useCallback(
    (
      target: TimelineViewportContinuityTarget,
      options: {
        requirements?: readonly TimelineContinuityRequirementName[];
      } = {},
    ) => {
      const activeRequest = activeViewportContinuityRequestRef.current;
      if (
        target.kind === "scroll-only" &&
        activeRequest !== null &&
        activeRequest.lifecycle.userInterruptionGeneration ===
          userInteractionVersionRef.current
      ) {
        return activeRequest.token;
      }
      const token = viewportContinuityTokenRef.current;
      viewportContinuityTokenRef.current += 1;
      const request: TimelineViewportContinuityRequest = {
        token,
        lifecycle: beginTimelineContinuityLifecycle({
          semanticFocusTarget: target,
          userInterruptionGeneration: userInteractionVersionRef.current,
          ...(options.requirements === undefined
            ? {}
            : { requirements: options.requirements }),
        }),
        preservedViewport: currentGridViewportSnapshot(
          resolveViewportContinuityElement(target),
        ),
      };
      activeViewportContinuityRequestRef.current = request;
      setViewportContinuityRequest(request);
      return token;
    },
    [
      currentGridViewportSnapshot,
      resolveViewportContinuityElement,
      setViewportContinuityRequest,
      viewportContinuityTokenRef,
    ],
  );
  const beginViewportContinuityRef = useRef(beginViewportContinuity);
  beginViewportContinuityRef.current = beginViewportContinuity;

  const settleViewportContinuityFollowUp = useCallback(
    (
      token: number,
      requirement: TimelineContinuityRequirementName,
      state: "settled" | "terminal",
    ) => {
      const activeRequest = activeViewportContinuityRequestRef.current;
      if (activeRequest?.token === token) {
        activeViewportContinuityRequestRef.current = {
          ...activeRequest,
          lifecycle: settleTimelineContinuityRequirement(
            activeRequest.lifecycle,
            requirement,
            state,
          ),
        };
      }
      setViewportContinuityRequest((current) => {
        if (!current || current.token !== token) {
          return current;
        }
        return {
          ...current,
          lifecycle: settleTimelineContinuityRequirement(
            current.lifecycle,
            requirement,
            state,
          ),
        };
      });
    },
    [setViewportContinuityRequest],
  );

  const clearViewportContinuity = useCallback(
    (token: number) => {
      const activeRequest = activeViewportContinuityRequestRef.current;
      if (activeRequest?.token === token) {
        activeViewportContinuityRequestRef.current = {
          ...activeRequest,
          lifecycle: transitionTimelineContinuity(
            activeRequest.lifecycle,
            "cancelled",
          ),
        };
        activeViewportContinuityRequestRef.current = null;
      }
      setViewportContinuityRequest((current) =>
        current?.token === token ? null : current,
      );
    },
    [setViewportContinuityRequest],
  );

  const failViewportContinuity = useCallback(
    (token: number) => {
      const activeRequest = activeViewportContinuityRequestRef.current;
      if (activeRequest?.token === token) {
        activeViewportContinuityRequestRef.current = {
          ...activeRequest,
          lifecycle: transitionTimelineContinuity(
            activeRequest.lifecycle,
            "failed",
          ),
        };
      }
      setViewportContinuityRequest((current) => {
        if (current === null || current.token !== token) {
          return current;
        }
        return {
          ...current,
          lifecycle: transitionTimelineContinuity(current.lifecycle, "failed"),
        };
      });
    },
    [setViewportContinuityRequest],
  );

  const requireViewportContinuitySourceRecord = useCallback(
    (token: number, requirement: TimelineSourceRecordRequirement) => {
      const activeRequest = activeViewportContinuityRequestRef.current;
      if (activeRequest?.token === token) {
        activeViewportContinuityRequestRef.current = {
          ...activeRequest,
          lifecycle: requireTimelineSourceRecord(
            activeRequest.lifecycle,
            requirement,
          ),
        };
      }
      setViewportContinuityRequest((current) => {
        if (current === null || current.token !== token) {
          return current;
        }
        return {
          ...current,
          lifecycle: requireTimelineSourceRecord(
            current.lifecycle,
            requirement,
          ),
        };
      });
    },
    [setViewportContinuityRequest],
  );

  const advanceViewportContinuity = useCallback(
    (
      token: number | undefined,
      options: {
        sourceRecord?: TimelineSourceRecordEvidence;
        target?: TimelineViewportContinuityTarget | null;
      } = {},
    ) => {
      if (token === undefined) {
        return;
      }
      const activeRequest = activeViewportContinuityRequestRef.current;
      if (activeRequest?.token === token) {
        activeViewportContinuityRequestRef.current = {
          ...activeRequest,
          lifecycle: advanceTimelineContinuityRender(
            {
              ...activeRequest.lifecycle,
              semanticFocusTarget:
                options.target ?? activeRequest.lifecycle.semanticFocusTarget,
            },
            { sourceRecord: options.sourceRecord },
          ),
        };
      }
      setViewportContinuityRequest((current) => {
        if (current === null || current.token !== token) {
          return current;
        }
        return {
          ...current,
          lifecycle: advanceTimelineContinuityRender(
            {
              ...current.lifecycle,
              semanticFocusTarget:
                options.target ?? current.lifecycle.semanticFocusTarget,
            },
            { sourceRecord: options.sourceRecord },
          ),
        };
      });
    },
    [setViewportContinuityRequest],
  );
  const advanceViewportContinuityRef = useRef(advanceViewportContinuity);
  advanceViewportContinuityRef.current = advanceViewportContinuity;

  const scrollToViewportContinuityTarget = useCallback(
    (target: TimelineViewportContinuityTarget) => {
      let anchor: GridCellAnchor | null = null;
      if (target.kind === "row-inspect") {
        anchor = {
          fieldKey: "timeline.activity_synopsis_text",
          rowIdentity: { kind: "core_record", recordId: target.recordId },
          surface: {
            kind: "view_schema",
            viewSchemaId: timelineViewSchemaId,
          },
        };
      } else if (target.kind === "input") {
        const [rowKey, fieldKey] = target.focusKey.split(":");
        const scalarBinding = timelineScalarBindings.find(
          (binding) => binding.key === fieldKey,
        );
        if (
          rowKey !== undefined &&
          !rowKey.startsWith("draft-") &&
          scalarBinding !== undefined
        ) {
          anchor = {
            fieldKey: scalarBinding.fieldKey,
            rowIdentity: { kind: "core_record", recordId: rowKey },
            surface: {
              kind: "view_schema",
              viewSchemaId: timelineViewSchemaId,
            },
          };
        }
      }
      return anchor === null
        ? false
        : (gridHandleRef.current?.scrollToAnchor(anchor) ?? false);
    },
    [gridHandleRef],
  );

  const tryRestoreViewportContinuity = useCallback(
    (continuity: TimelineViewportContinuityRequest) => {
      const target = continuity.lifecycle.semanticFocusTarget;
      if (target.kind === "scroll-only") {
        restoreGridScroll(continuity.preservedViewport?.scroll ?? null);
        return true;
      }
      if (resolveViewportContinuityElement(target) === null) {
        scrollToViewportContinuityTarget(target);
      }
      return restoreGridViewportForElement(
        () => resolveViewportContinuityElement(target),
        continuity.preservedViewport,
      );
    },
    [
      resolveViewportContinuityElement,
      restoreGridScroll,
      restoreGridViewportForElement,
      scrollToViewportContinuityTarget,
    ],
  );

  const shouldHoldViewportContinuity = useCallback(
    (continuity: TimelineViewportContinuityRequest) => {
      return !timelineContinuityRequirementsSettled(continuity.lifecycle);
    },
    [],
  );

  const userInterruptedViewportContinuity = useCallback(
    (continuity: TimelineViewportContinuityRequest) => {
      return (
        userInteractionVersionRef.current !==
        continuity.lifecycle.userInterruptionGeneration
      );
    },
    [],
  );

  useLayoutEffect(() => {
    if (
      viewportContinuityRequest === null ||
      viewportContinuityRequest.lifecycle.renderGeneration < 1
    ) {
      return;
    }
    let cancelled = false;
    const restoreTarget = (attempt: number) => {
      if (cancelled) {
        return;
      }
      if (userInterruptedViewportContinuity(viewportContinuityRequest)) {
        clearViewportContinuity(viewportContinuityRequest.token);
        return;
      }
      if (!tryRestoreViewportContinuity(viewportContinuityRequest)) {
        if (attempt < 60) {
          window.setTimeout(() => {
            restoreTarget(attempt + 1);
          }, 50);
        } else {
          clearViewportContinuity(viewportContinuityRequest.token);
        }
        return;
      }
      // A slow named follow-up must not leave focus on <body> after the
      // authoritative row has committed. Restore the deterministic fallback
      // provisionally, but keep the lifecycle open until every follow-up has
      // settled so a later render is revalidated before completion.
      if (shouldHoldViewportContinuity(viewportContinuityRequest)) {
        return;
      }
      const stableRenderGeneration =
        viewportContinuityRequest.lifecycle.renderGeneration;
      window.requestAnimationFrame(() => {
        if (cancelled) {
          return;
        }
        const activeRequest = activeViewportContinuityRequestRef.current;
        if (
          activeRequest?.token !== viewportContinuityRequest.token ||
          activeRequest.lifecycle.renderGeneration !== stableRenderGeneration
        ) {
          return;
        }
        if (userInterruptedViewportContinuity(activeRequest)) {
          clearViewportContinuity(activeRequest.token);
          return;
        }
        if (shouldHoldViewportContinuity(viewportContinuityRequest)) {
          return;
        }
        if (!tryRestoreViewportContinuity(viewportContinuityRequest)) {
          if (attempt < 60) {
            window.setTimeout(() => {
              restoreTarget(attempt + 1);
            }, 50);
          } else {
            clearViewportContinuity(viewportContinuityRequest.token);
          }
          return;
        }
        activeViewportContinuityRequestRef.current = {
          ...activeRequest,
          lifecycle: transitionTimelineContinuity(
            activeRequest.lifecycle,
            activeRequest.lifecycle.state === "failed" ? "failed" : "completed",
          ),
        };
        clearViewportContinuity(viewportContinuityRequest.token);
      });
    };
    restoreTarget(0);
    return () => {
      cancelled = true;
    };
  }, [
    clearViewportContinuity,
    shouldHoldViewportContinuity,
    tryRestoreViewportContinuity,
    userInterruptedViewportContinuity,
    viewportContinuityRequest,
  ]);

  return {
    commands: {
      advanceViewportContinuity,
      beginViewportContinuity,
      clearViewportContinuity,
      currentGridScrollSnapshot,
      currentGridViewportSnapshot,
      failViewportContinuity,
      requireViewportContinuitySourceRecord,
      resolveInputElement,
      resolveViewportContinuityElement,
      restoreGridScroll,
      restoreGridViewportForElement,
      scrollToViewportContinuityTarget,
      settleViewportContinuityFollowUp,
    },
    refs: {
      advanceViewportContinuityRef,
      beginViewportContinuityRef,
    },
  };
}
