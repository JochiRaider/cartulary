import {
  dataTestIdSelector,
  draftCellTestId,
  gridRowGutterTestId,
  gridScrollportSelector,
  rowCellTestId,
} from "@cartulary/ui-contracts";
import { useCallback, useLayoutEffect, useRef } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type ScrollPosition,
  type ViewportSnapshot,
} from "../../utils/workbookContinuity";
import {
  settleTimelineViewportContinuityBarrier,
  type TimelineEntityCatalogInput,
  type TimelineEntityRefreshSettleState,
  type TimelineViewportContinuityBarrier,
  timelineViewportContinuityBarrierSatisfied,
} from "../models/timelineViewportContinuityModel";
import { timelineScalarBindings } from "../models/workbookTimelineModel";

type TimelineMutableRef<T> = {
  current: T;
};

export type TimelineViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

export type TimelineViewportContinuityRequest = {
  token: number;
  attemptVersion: number;
  target: TimelineViewportContinuityTarget;
  preservedViewport: ViewportSnapshot | null;
  barrier: TimelineViewportContinuityBarrier;
};

function resolveTimelineGridScrollElement(
  element: HTMLElement,
  surface: string,
): HTMLElement {
  const selector = gridScrollportSelector();
  const scrollports = Array.from(
    element.querySelectorAll<HTMLElement>(selector),
  );
  if (scrollports.length !== 1) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${selector} scrollport, received ${scrollports.length}`,
    );
  }
  const scrollport = scrollports[0];
  if (scrollport === undefined) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${selector} scrollport, received 0`,
    );
  }
  return scrollport;
}

export function useTimelineViewportContinuityController({
  entityCatalogInput,
  gridShellRef,
  rowInputRefs,
  rowInputTestIdsRef,
  setViewportContinuityRequest,
  viewportContinuityRequest,
  viewportContinuityTokenRef,
}: {
  readonly entityCatalogInput: TimelineEntityCatalogInput;
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
  const currentGridScrollSnapshot = useCallback(() => {
    const element = gridShellRef.current;
    if (!element) {
      return null;
    }
    const scrollElement = resolveTimelineGridScrollElement(element, "timeline");
    return {
      top: scrollElement.scrollTop,
      left: scrollElement.scrollLeft,
    };
  }, [gridShellRef]);

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
        case "row-inspect":
          return (
            document.querySelector<HTMLElement>(
              dataTestIdSelector(
                rowCellTestId(
                  target.recordId,
                  "timeline.activity_synopsis_text",
                ),
              ),
            ) ??
            document.querySelector<HTMLElement>(
              dataTestIdSelector(
                gridRowGutterTestId(timelineViewSchemaId, target.recordId),
              ),
            )
          );
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
      if (gridShell === null || scroll === null) {
        return null;
      }
      return {
        scroll,
        anchor:
          target === null
            ? null
            : captureViewportAnchor(
                resolveTimelineGridScrollElement(
                  gridShell,
                  "timeline",
                ).getBoundingClientRect(),
                target.getBoundingClientRect(),
              ),
      };
    },
    [currentGridScrollSnapshot, gridShellRef],
  );

  const restoreGridScroll = useCallback(
    (preservedScroll: ScrollPosition | null) => {
      const gridShell = gridShellRef.current;
      if (gridShell === null || preservedScroll === null) {
        return;
      }
      const scrollElement = resolveTimelineGridScrollElement(
        gridShell,
        "timeline",
      );
      scrollElement.scrollTop = preservedScroll.top;
      scrollElement.scrollLeft = preservedScroll.left;
      window.requestAnimationFrame(() => {
        const currentGridShell = gridShellRef.current;
        if (currentGridShell === null) {
          return;
        }
        const currentScrollElement = resolveTimelineGridScrollElement(
          currentGridShell,
          "timeline",
        );
        currentScrollElement.scrollTop = preservedScroll.top;
        currentScrollElement.scrollLeft = preservedScroll.left;
      });
    },
    [gridShellRef],
  );

  const restoreGridViewportForElement = useCallback(
    (
      resolveElement: () => HTMLElement | null,
      preservedViewport: ViewportSnapshot | null,
    ) => {
      const currentViewport =
        preservedViewport ??
        ({
          scroll: currentGridScrollSnapshot(),
          anchor: null,
        } satisfies ViewportSnapshot);
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
        const currentGridShell = gridShellRef.current;
        const currentElement = resolveElement();
        if (
          currentGridShell === null ||
          preservedScroll === null ||
          currentElement === null ||
          !currentElement.isConnected
        ) {
          return false;
        }
        const scrollElement = resolveTimelineGridScrollElement(
          currentGridShell,
          "timeline",
        );
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
        const updatedGridShell = gridShellRef.current;
        const updatedElement = resolveElement();
        if (
          updatedGridShell === null ||
          updatedElement === null ||
          !updatedElement.isConnected
        ) {
          return false;
        }
        const fullyVisible = isRectFullyVisibleWithinContainer(
          resolveTimelineGridScrollElement(
            updatedGridShell,
            "timeline",
          ).getBoundingClientRect(),
          updatedElement.getBoundingClientRect(),
        );
        return focusResolvedElement() && fullyVisible;
      };
      const restoredNow = restoreViewportGeometryNow();
      const restoreViewportGeometry = (attempt: number) => {
        window.requestAnimationFrame(() => {
          if (restoreViewportGeometryNow()) {
            return;
          }
          if (attempt < 6) {
            restoreViewportGeometry(attempt + 1);
          }
        });
      };
      restoreViewportGeometry(0);
      return focusedNow && restoredNow;
    },
    [currentGridScrollSnapshot, gridShellRef, restoreGridScroll],
  );

  const beginViewportContinuity = useCallback(
    (
      target: TimelineViewportContinuityTarget,
      options: { barrier?: TimelineViewportContinuityBarrier } = {},
    ) => {
      const token = viewportContinuityTokenRef.current;
      viewportContinuityTokenRef.current += 1;
      setViewportContinuityRequest({
        token,
        attemptVersion: 0,
        target,
        preservedViewport: currentGridViewportSnapshot(
          resolveViewportContinuityElement(target),
        ),
        barrier: options.barrier ?? null,
      });
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

  const settleViewportContinuityBarrier = useCallback(
    (token: number, refreshState: TimelineEntityRefreshSettleState) => {
      setViewportContinuityRequest((current) => {
        if (!current || current.token !== token) {
          return current;
        }
        return {
          ...current,
          barrier: settleTimelineViewportContinuityBarrier(
            current.barrier,
            refreshState,
          ),
          attemptVersion: current.attemptVersion + 1,
        };
      });
    },
    [setViewportContinuityRequest],
  );

  const clearViewportContinuity = useCallback(
    (token: number) => {
      setViewportContinuityRequest((current) =>
        current?.token === token ? null : current,
      );
    },
    [setViewportContinuityRequest],
  );

  const advanceViewportContinuity = useCallback(
    (
      token: number | undefined,
      options: {
        barrier?: TimelineViewportContinuityBarrier;
        target?: TimelineViewportContinuityTarget | null;
      } = {},
    ) => {
      if (token === undefined) {
        return;
      }
      setViewportContinuityRequest((current) => {
        if (current === null || current.token !== token) {
          return current;
        }
        return {
          ...current,
          attemptVersion: current.attemptVersion + 1,
          barrier:
            options.barrier === undefined ? current.barrier : options.barrier,
          target: options.target ?? current.target,
        };
      });
    },
    [setViewportContinuityRequest],
  );
  const advanceViewportContinuityRef = useRef(advanceViewportContinuity);
  advanceViewportContinuityRef.current = advanceViewportContinuity;

  const tryRestoreViewportContinuity = useCallback(
    (continuity: TimelineViewportContinuityRequest) => {
      if (continuity.target.kind === "scroll-only") {
        restoreGridScroll(continuity.preservedViewport?.scroll ?? null);
        return true;
      }
      return restoreGridViewportForElement(
        () => resolveViewportContinuityElement(continuity.target),
        continuity.preservedViewport,
      );
    },
    [
      resolveViewportContinuityElement,
      restoreGridScroll,
      restoreGridViewportForElement,
    ],
  );

  const shouldHoldViewportContinuity = useCallback(
    (continuity: TimelineViewportContinuityRequest) => {
      return !timelineViewportContinuityBarrierSatisfied(
        continuity.barrier,
        entityCatalogInput,
      );
    },
    [entityCatalogInput],
  );

  useLayoutEffect(() => {
    if (
      viewportContinuityRequest === null ||
      viewportContinuityRequest.attemptVersion < 1
    ) {
      return;
    }
    let cancelled = false;
    const restoreTarget = (attempt: number) => {
      if (cancelled) {
        return;
      }
      if (!tryRestoreViewportContinuity(viewportContinuityRequest)) {
        if (attempt < 60) {
          window.setTimeout(() => {
            restoreTarget(attempt + 1);
          }, 50);
        }
        return;
      }
      if (shouldHoldViewportContinuity(viewportContinuityRequest)) {
        return;
      }
      window.requestAnimationFrame(() => {
        if (cancelled) {
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
          }
          return;
        }
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
    viewportContinuityRequest,
  ]);

  return {
    commands: {
      advanceViewportContinuity,
      beginViewportContinuity,
      clearViewportContinuity,
      currentGridScrollSnapshot,
      currentGridViewportSnapshot,
      resolveInputElement,
      resolveViewportContinuityElement,
      restoreGridScroll,
      restoreGridViewportForElement,
      settleViewportContinuityBarrier,
    },
    refs: {
      advanceViewportContinuityRef,
      beginViewportContinuityRef,
    },
  };
}
