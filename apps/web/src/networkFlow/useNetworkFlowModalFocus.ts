import {
  networkAnalysisTestId,
  workbookIncidentIdentityTestId,
} from "@cartulary/ui-contracts";
import {
  type KeyboardEvent as ReactKeyboardEvent,
  useCallback,
  useEffect,
  useRef,
} from "react";
import { useWorkbookSecondaryPanel } from "../shared/WorkbookRecoveryBoundary";

const focusableSelector = [
  "a[href]",
  "summary",
  "button:not([disabled])",
  "input:not([disabled])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  "[tabindex]:not([tabindex='-1'])",
].join(",");

export function useNetworkFlowModalFocus<Element extends HTMLElement>(options: {
  readonly dismissDisabled?: boolean | undefined;
  readonly initialFocusTestId?: string | undefined;
  readonly fallbackFocusTestId?: string | undefined;
  readonly restoreFallbackFocus?:
    | (() => boolean | Promise<boolean>)
    | undefined;
  readonly onDismiss: () => void;
}) {
  const coordinated = useWorkbookSecondaryPanel(true, options.onDismiss);
  const dialogRef = useRef<Element | null>(null);
  const dismissDisabledRef = useRef(options.dismissDisabled ?? false);
  const onDismissRef = useRef(options.onDismiss);
  const restoreFallbackRef = useRef(options.restoreFallbackFocus);
  restoreFallbackRef.current = options.restoreFallbackFocus;
  dismissDisabledRef.current = options.dismissDisabled ?? false;
  onDismissRef.current = options.onDismiss;

  useEffect(() => {
    const previouslyFocused =
      document.activeElement instanceof HTMLElement
        ? document.activeElement
        : null;
    const dialog = dialogRef.current;
    if (dialog === null) return;
    let mounted = true;
    queueMicrotask(() => {
      if (!mounted) return;
      const preferred =
        options.initialFocusTestId === undefined
          ? null
          : (Array.from(dialog.getElementsByTagName("*")).find(
              (element): element is HTMLElement =>
                element instanceof HTMLElement &&
                element.dataset.testid === options.initialFocusTestId,
            ) ?? null);
      const target = preferred ?? modalFocusableElements(dialog)[0] ?? dialog;
      if (!target.hasAttribute("tabindex") && target === dialog) {
        target.tabIndex = -1;
      }
      target.focus({ preventScroll: true });
    });
    return () => {
      mounted = false;
      queueMicrotask(async () => {
        if (coordinated.current) return;
        const focusBeforeRestore = document.activeElement;
        if (
          document.querySelector(
            '[role="dialog"][aria-modal="true"], [role="alertdialog"][aria-modal="true"]',
          ) !== null
        )
          return;
        if (
          (previouslyFocused?.isConnected !== true ||
            previouslyFocused.hasAttribute("disabled")) &&
          (await restoreFallbackRef.current?.()) === true
        )
          return;
        if (
          document.activeElement !== focusBeforeRestore ||
          document.querySelector(
            '[role="dialog"][aria-modal="true"], [role="alertdialog"][aria-modal="true"]',
          ) !== null
        )
          return;
        const target =
          previouslyFocused?.isConnected === true &&
          !previouslyFocused.hasAttribute("disabled")
            ? previouslyFocused
            : ([
                options.fallbackFocusTestId ??
                  networkAnalysisTestId("workspace"),
                networkAnalysisTestId("tab"),
                workbookIncidentIdentityTestId(),
              ].flatMap((testId) =>
                Array.from(document.getElementsByTagName("*")).filter(
                  (element): element is HTMLElement =>
                    element instanceof HTMLElement &&
                    element.dataset.testid === testId &&
                    !element.hidden &&
                    !element.hasAttribute("disabled"),
                ),
              )[0] ?? null);
        if (target === null) return;
        if (!target.matches(focusableSelector)) target.tabIndex = -1;
        target.focus({ preventScroll: true });
      });
    };
  }, [options.initialFocusTestId, options.fallbackFocusTestId, coordinated]);

  const onKeyDown = useCallback((event: ReactKeyboardEvent<Element>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      event.stopPropagation();
      if (!dismissDisabledRef.current) onDismissRef.current();
      return;
    }
    if (event.key !== "Tab") return;
    const dialog = dialogRef.current;
    if (dialog === null) return;
    const focusable = modalFocusableElements(dialog);
    if (focusable.length === 0) {
      event.preventDefault();
      dialog.focus({ preventScroll: true });
      return;
    }
    const first = focusable[0] as HTMLElement;
    const last = focusable.at(-1) as HTMLElement;
    if (
      event.shiftKey &&
      (document.activeElement === first ||
        !focusable.includes(document.activeElement as HTMLElement))
    ) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }, []);

  return { dialogRef, onKeyDown };
}

function modalFocusableElements(dialog: HTMLElement): HTMLElement[] {
  return Array.from(
    dialog.querySelectorAll<HTMLElement>(focusableSelector),
  ).filter(
    (element) =>
      !element.hidden &&
      element.closest('[hidden], [aria-hidden="true"]') === null &&
      (element.closest("details:not([open])") === null ||
        element.tagName === "SUMMARY") &&
      element.getAttribute("aria-hidden") !== "true" &&
      element.tabIndex >= 0,
  );
}
