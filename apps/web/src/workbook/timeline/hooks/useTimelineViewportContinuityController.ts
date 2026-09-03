import type { GridCellAnchor, GridHandle } from "@cartulary/grid-adapter";
import { useCallback, useEffect, useLayoutEffect, useRef } from "react";
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type ScrollPosition,
  type ViewportSnapshot,
} from "../../continuity/gridViewportContinuity";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { timelineScalarBindings } from "../models/timelineFieldRegistry";
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

type TimelineMutableRef<T> = {
  current: T;
};

export type TimelineViewportContinuityTarget = TimelineContinuitySemanticTarget;

export type TimelineViewportContinuityRequest = {
  token: number;
  lifecycle: TimelineContinuityLifecycle;
  preservedViewport: ViewportSnapshot | null;
};

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
  setViewportContinuityRequest,
  viewportContinuityRequest,
  viewportContinuityTokenRef,
}: {
  readonly gridHandleRef: TimelineMutableRef<GridHandle | null>;
  readonly gridShellRef: TimelineMutableRef<HTMLDivElement | null>;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
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
    (focusKey: string) => editorDraftRegistry.inputElementForFocusKey(focusKey),
    [editorDraftRegistry],
  );

  const resolveViewportContinuityRect = useCallback(
    (target: TimelineViewportContinuityTarget) => {
      switch (target.kind) {
        case "row-inspect":
          return (
            gridHandleRef.current?.getAnchorRect(
              timelineAnchor(
                target.recordId,
                "timeline.activity_synopsis_text",
              ),
            ) ?? null
          );
        case "input":
          return (
            resolveInputElement(target.focusKey)?.getBoundingClientRect() ??
            null
          );
        case "scroll-only":
          return null;
      }
    },
    [gridHandleRef, resolveInputElement],
  );

  const currentGridViewportSnapshot = useCallback(
    (targetRect: DOMRectReadOnly | null = null): ViewportSnapshot | null => {
      const gridShell = gridShellRef.current;
      const scroll = currentGridScrollSnapshot();
      const scrollElement = currentGridScrollElement();
      if (gridShell === null || scroll === null || scrollElement === null) {
        return null;
      }
      const containerRect = scrollElement.getBoundingClientRect();
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

  const restoreGridViewportForTarget = useCallback(
    (
      focusTarget: () => boolean,
      resolveRect: () => DOMRectReadOnly | null,
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
      window.focus();
      const focusedNow = focusTarget();
      restoreGridScroll(preservedScroll);
      const restoreViewportGeometryNow = () => {
        const scrollElement = currentGridScrollElement();
        const currentRect = resolveRect();
        if (
          scrollElement === null ||
          preservedScroll === null ||
          currentRect === null
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
          elementRect: currentRect,
        });
        restoreGridScroll(restoredScroll);
        const updatedScrollElement = currentGridScrollElement();
        const updatedRect = resolveRect();
        if (updatedScrollElement === null || updatedRect === null) {
          return false;
        }
        const fullyVisible = isRectFullyVisibleWithinContainer(
          updatedScrollElement.getBoundingClientRect(),
          updatedRect,
        );
        return focusTarget() && fullyVisible;
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
          resolveViewportContinuityRect(target),
        ),
      };
      activeViewportContinuityRequestRef.current = request;
      setViewportContinuityRequest(request);
      return token;
    },
    [
      currentGridViewportSnapshot,
      resolveViewportContinuityRect,
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
        anchor = timelineAnchor(
          target.recordId,
          "timeline.activity_synopsis_text",
        );
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

  const focusViewportContinuityTarget = useCallback(
    (target: TimelineViewportContinuityTarget): boolean => {
      if (target.kind === "row-inspect") {
        return (
          gridHandleRef.current?.focusAnchor(
            timelineAnchor(target.recordId, "timeline.activity_synopsis_text"),
          ) ?? false
        );
      }
      if (target.kind === "input") {
        const element = resolveInputElement(target.focusKey);
        if (element === null) return false;
        element.focus({ preventScroll: true });
        return document.activeElement === element;
      }
      return false;
    },
    [gridHandleRef, resolveInputElement],
  );

  const tryRestoreViewportContinuity = useCallback(
    (continuity: TimelineViewportContinuityRequest) => {
      const target = continuity.lifecycle.semanticFocusTarget;
      if (target.kind === "scroll-only") {
        restoreGridScroll(continuity.preservedViewport?.scroll ?? null);
        return true;
      }
      if (resolveViewportContinuityRect(target) === null) {
        scrollToViewportContinuityTarget(target);
      }
      return restoreGridViewportForTarget(
        () => focusViewportContinuityTarget(target),
        () => resolveViewportContinuityRect(target),
        continuity.preservedViewport,
      );
    },
    [
      focusViewportContinuityTarget,
      resolveViewportContinuityRect,
      restoreGridScroll,
      restoreGridViewportForTarget,
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
      resolveViewportContinuityRect,
      restoreGridScroll,
      restoreGridViewportForTarget,
      scrollToViewportContinuityTarget,
      settleViewportContinuityFollowUp,
    },
    refs: {
      advanceViewportContinuityRef,
      beginViewportContinuityRef,
    },
  };
}
