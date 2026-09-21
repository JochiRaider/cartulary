import type {
  InspectorPanel,
  InspectorPanelId,
} from "@cartulary/view-contracts";
import {
  createContext,
  type RefObject,
  useContext,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

// One bounded presentation selection survives closing the layout's inspector.
// It contains no record content, authoring, request or permission state.
export type WorkbookInspectorNavigationSelection = {
  readonly scope: string;
  readonly panelId: InspectorPanelId;
};
export const WorkbookInspectorNavigationContext =
  createContext<RefObject<WorkbookInspectorNavigationSelection | null> | null>(
    null,
  );

type NavigationSection = {
  readonly panel: InspectorPanel;
  readonly focusDestination: (section: HTMLElement) => HTMLElement;
};

export function workbookInspectorSectionFocusDestination(
  section: HTMLElement,
): HTMLElement {
  return (
    section.querySelector<HTMLElement>(
      "[data-inspector-section-entry]:not(:disabled)",
    ) ?? section
  );
}

export function useWorkbookInspectorNavigation(
  scope: string,
  sections: readonly NavigationSection[],
) {
  const retainedNavigation = useContext(WorkbookInspectorNavigationContext);
  const [selection, setSelection] = useState(
    retainedNavigation?.current ?? null,
  );
  const [menuScope, setMenuScope] = useState<string | null>(null);
  const bodyRef = useRef<HTMLDivElement>(null);
  const navigationRef = useRef<HTMLFieldSetElement>(null);
  const measurementRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const closeRef = useRef<HTMLButtonElement>(null);
  const elements = useRef(new Map<InspectorPanelId, HTMLElement>());
  const displacedFocus = useRef(false);
  const reconciledBoundary = useRef<string | null>(null);
  const reconciledScope = useRef(scope);
  const positionedScroll = useRef<number | null>(null);
  const [fits, setFits] = useState(false);
  const navigationFocus = useRef<string | null>(null);
  const previousDirect = useRef(false);
  const currentNavigation = useRef({ scope, sections });
  currentNavigation.current = { scope, sections };
  const active =
    sections.find(
      ({ panel }) =>
        selection?.scope === scope && panel.panelId === selection.panelId,
    ) ?? sections[0];
  const menuOpen = menuScope === scope && sections.length > 0;
  const direct = fits && !menuOpen;
  const labels = JSON.stringify(sections.map(({ panel }) => panel.label));
  // biome-ignore lint/correctness/useExhaustiveDependencies: The DOM labels and subject boundary invalidate geometry even though measurement reads the DOM rather than their values.
  useLayoutEffect(() => {
    const container = navigationRef.current;
    const measurement = measurementRef.current;
    if (!container || !measurement) return;
    const measure = () => {
      const available = container.clientWidth;
      const required = measurement.getBoundingClientRect().width;
      setFits(
        document.fonts?.status !== "loading" &&
          available > 0 &&
          required > 0 &&
          required <= available,
      );
    };
    measure();
    const resize =
      typeof ResizeObserver === "undefined"
        ? null
        : new ResizeObserver(measure);
    resize?.observe(container);
    resize?.observe(measurement);
    const spacing = new MutationObserver(measure);
    for (const element of [document.documentElement, document.body, container])
      spacing.observe(element, {
        attributes: true,
        attributeFilter: ["style", "class"],
      });
    document.fonts?.addEventListener("loadingdone", measure);
    window.addEventListener("resize", measure);
    return () => {
      resize?.disconnect();
      spacing.disconnect();
      document.fonts?.removeEventListener("loadingdone", measure);
      window.removeEventListener("resize", measure);
    };
  }, [labels, scope]);
  const remember = (panelId: InspectorPanelId) => {
    const next = { scope, panelId };
    if (retainedNavigation) retainedNavigation.current = next;
    setSelection((previous) =>
      previous?.scope === scope && previous.panelId === panelId
        ? previous
        : next,
    );
  };
  const scrollToSection = (element: HTMLElement) => {
    const body = bodyRef.current;
    if (!body) return;
    const bounds = body.getBoundingClientRect();
    const scale = body.offsetHeight > 0 ? bounds.height / body.offsetHeight : 1;
    body.scrollTop = Math.max(
      0,
      Math.min(
        body.scrollHeight - body.clientHeight,
        body.scrollTop +
          (element.getBoundingClientRect().top - bounds.top) / (scale || 1),
      ),
    );
    positionedScroll.current = body.scrollTop;
  };
  const choose = (
    section: NavigationSection,
    destination?: (element: HTMLElement) => HTMLElement | null,
  ) => {
    if (
      currentNavigation.current.scope !== scope ||
      !currentNavigation.current.sections.includes(section)
    )
      return;
    setMenuScope(null);
    remember(section.panel.panelId);
    const element = elements.current.get(section.panel.panelId);
    if (!element) return;
    const target = destination?.(element) ?? section.focusDestination(element);
    scrollToSection(destination ? target : element);
    target.focus({ preventScroll: true });
  };
  const observeScroll = () => {
    const body = bodyRef.current;
    if (!body || !sections.length) return;
    if (positionedScroll.current === body.scrollTop) return;
    positionedScroll.current = null;
    if (body.scrollHeight <= body.clientHeight) {
      if (sections[0]) remember(sections[0].panel.panelId);
      return;
    }
    const top = body.getBoundingClientRect().top;
    const preceding = sections.filter(
      ({ panel }) =>
        (elements.current.get(panel.panelId)?.getBoundingClientRect().top ??
          Infinity) <=
        top + 1,
    );
    const current =
      body.scrollTop + body.clientHeight >= body.scrollHeight - 1
        ? sections.at(-1)
        : (preceding.at(-1) ?? sections[0]);
    if (current) remember(current.panel.panelId);
  };
  const admittedIds = sections.map(({ panel }) => panel.panelId).join(",");
  useLayoutEffect(() => {
    const navigationControl = () =>
      direct
        ? navigationRef.current?.querySelector<HTMLElement>(
            `[data-inspector-navigation-panel="${active?.panel.panelId}"]`,
          )
        : triggerRef.current;
    if (
      previousDirect.current !== direct &&
      navigationFocus.current &&
      document.activeElement === document.body
    )
      navigationControl()?.focus({ preventScroll: true });
    previousDirect.current = direct;
    // A subject retarget belongs to its opener (often a grid pointer gesture).
    // Only recover focus removed by concealment within the same subject.
    if (
      reconciledScope.current === scope &&
      displacedFocus.current &&
      document.activeElement === document.body
    )
      (
        navigationRef.current?.querySelector<HTMLElement>("button") ??
        closeRef.current
      )?.focus({ preventScroll: true });
    reconciledScope.current = scope;
    displacedFocus.current = false;
    const boundary = JSON.stringify([scope, admittedIds]);
    if (reconciledBoundary.current === boundary) return;
    reconciledBoundary.current = boundary;
    setMenuScope(null);
    const focused = sections.find(({ panel }) =>
      elements.current.get(panel.panelId)?.contains(document.activeElement),
    );
    if (focused) remember(focused.panel.panelId);
    else if (active) {
      remember(active.panel.panelId);
      const element = elements.current.get(active.panel.panelId);
      if (element) scrollToSection(element);
    }
    // Scope/admission changes reset presentation; ordinary content refresh does not.
  });
  useEffect(() => {
    if (!menuOpen) return;
    const dismiss = (event: PointerEvent) => {
      if (
        event.target instanceof Node &&
        !navigationRef.current?.contains(event.target)
      )
        setMenuScope(null);
    };
    document.addEventListener("pointerdown", dismiss);
    return () => document.removeEventListener("pointerdown", dismiss);
  }, [menuOpen]);
  return {
    active,
    menuOpen,
    direct,
    measurementRef,
    onNavigationFocus: (target: EventTarget) => {
      navigationFocus.current =
        target instanceof HTMLElement
          ? (target.dataset.inspectorNavigationPanel ?? "chooser")
          : null;
    },
    onNavigationBlur: (target: EventTarget | null) => {
      if (target instanceof Node && !navigationRef.current?.contains(target))
        navigationFocus.current = null;
    },
    bodyRef,
    navigationRef,
    triggerRef,
    closeRef,
    choose,
    reveal: scrollToSection,
    observeScroll,
    toggleMenu: () => setMenuScope(menuOpen ? null : scope),
    dismissMenu: () => {
      setMenuScope(null);
      triggerRef.current?.focus({ preventScroll: true });
    },
    remember,
    registerSection: (
      panelId: InspectorPanelId,
      element: HTMLElement | null,
    ) => {
      const previous = elements.current.get(panelId);
      if (!element && previous?.contains(document.activeElement))
        displacedFocus.current = true;
      if (element) elements.current.set(panelId, element);
      else elements.current.delete(panelId);
    },
  };
}
