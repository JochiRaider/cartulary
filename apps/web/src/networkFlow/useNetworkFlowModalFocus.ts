import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import {
  type KeyboardEvent as ReactKeyboardEvent,
  useCallback,
  useEffect,
  useRef,
} from "react";

const focusableSelector = [
  "a[href]",
  "button:not([disabled])",
  "input:not([disabled])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  "[tabindex]:not([tabindex='-1'])",
].join(",");

export function useNetworkFlowModalFocus<Element extends HTMLElement>(options: {
  readonly dismissDisabled?: boolean | undefined;
  readonly initialFocusTestId?: string | undefined;
  readonly onDismiss: () => void;
}) {
  const dialogRef = useRef<Element | null>(null);
  const dismissDisabledRef = useRef(options.dismissDisabled ?? false);
  const onDismissRef = useRef(options.onDismiss);
  dismissDisabledRef.current = options.dismissDisabled ?? false;
  onDismissRef.current = options.onDismiss;

  useEffect(() => {
    const previouslyFocused =
      document.activeElement instanceof HTMLElement
        ? document.activeElement
        : null;
    const dialog = dialogRef.current;
    if (dialog === null) return;
    queueMicrotask(() => {
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
      queueMicrotask(() => {
        const target =
          previouslyFocused?.isConnected === true
            ? previouslyFocused
            : (Array.from(document.getElementsByTagName("*")).find(
                (element): element is HTMLElement =>
                  element instanceof HTMLElement &&
                  element.dataset.testid === networkAnalysisTestId("workspace"),
              ) ?? null);
        if (target === null) return;
        if (!target.hasAttribute("tabindex")) target.tabIndex = -1;
        target.focus({ preventScroll: true });
      });
    };
  }, [options.initialFocusTestId]);

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
    if (event.shiftKey && document.activeElement === first) {
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
      element.getAttribute("aria-hidden") !== "true" &&
      element.tabIndex >= 0,
  );
}
