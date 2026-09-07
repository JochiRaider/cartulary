import {
  type CSSProperties,
  type ReactNode,
  type RefObject,
  useLayoutEffect,
  useRef,
} from "react";
import { useRegisteredOverlayNavigation } from "../shared/useRegisteredOverlayNavigation";

/** Native containment and cancellation; the shared overlay owner restores focus. */
export function AccountDialog({
  label,
  labelledBy,
  describedBy,
  fallbackFocusRef,
  onClose,
  style,
  children,
}: {
  label?: string;
  labelledBy?: string;
  describedBy?: string;
  fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
  onClose: () => void;
  style: CSSProperties;
  children: (dismiss: () => void) => ReactNode;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const triggerRef = useRef(
    document.activeElement instanceof HTMLElement &&
      document.activeElement !== document.body
      ? document.activeElement
      : null,
  );
  const closing = useRef(false);
  const controlKeys = useRef<string[]>([]);
  const navigation = useRegisteredOverlayNavigation({
    isOpen: true,
    initialItemKey: null,
    itemKeys: controlKeys.current,
    trapTab: true,
    triggerRef,
    fallbackFocusRef,
    subjectKey: "account-dialog",
    restoreFocusOnSubjectChange: false,
    onRequestClose: () => {
      if (closing.current) return;
      closing.current = true;
      dialog.current?.close();
      onClose();
    },
  });
  useLayoutEffect(() => {
    const element = dialog.current;
    if (!element) return;
    closing.current = false;
    element.showModal();
    element.querySelector<HTMLButtonElement>("button:not(:disabled)")?.focus();
    return () => {
      closing.current = true;
      if (element.open) element.close();
      if (
        !triggerRef.current?.isConnected &&
        fallbackFocusRef?.current?.isConnected
      )
        fallbackFocusRef.current.focus({ preventScroll: true });
    };
  }, [fallbackFocusRef]);
  const dismiss = () => {
    if (!closing.current) navigation.close({ restoreTriggerFocus: true });
  };
  return (
    <>
      <style>{`.cartulary-account-dialog:not([open]) { display: none; }
      .cartulary-account-dialog::backdrop { background: var(--ct-colors-overlay-backdrop, rgba(12,16,24,0.42)); }`}</style>
      <dialog
        ref={dialog}
        className="cartulary-account-dialog"
        aria-label={label}
        aria-labelledby={labelledBy}
        aria-describedby={describedBy}
        aria-modal="true"
        tabIndex={-1}
        style={{
          position: "fixed",
          inset: 0,
          margin: "auto",
          color: "var(--ct-colors-ink)",
          maxHeight: "calc(100% - 2rem)",
          maxWidth: "calc(100% - 2rem)",
          boxSizing: "border-box",
          overflow: "auto",
          ...style,
        }}
        onKeyDown={(event) => {
          if (event.key !== "Tab" || event.defaultPrevented) return;
          const controls = [
            ...event.currentTarget.querySelectorAll<HTMLElement>(
              "button, input, select, textarea, a[href], [tabindex]",
            ),
          ].filter(
            (element) =>
              element.tabIndex >= 0 &&
              !element.hasAttribute("disabled") &&
              !element.closest("[hidden]") &&
              getComputedStyle(element).visibility !== "hidden" &&
              getComputedStyle(element).display !== "none" &&
              !(
                element instanceof HTMLInputElement &&
                element.type === "radio" &&
                !element.checked
              ),
          );
          for (const key of controlKeys.current)
            navigation.registerItem(key)(null);
          controlKeys.current.length = 0;
          for (const [index, control] of controls.entries()) {
            const key = String(index);
            controlKeys.current.push(key);
            navigation.registerItem(key)(control);
          }
          if (controls.length === 0) {
            event.preventDefault();
            event.currentTarget.focus();
            return;
          }
          navigation.onItemKeyDown(
            event,
            String(controls.indexOf(document.activeElement as HTMLElement)),
          );
        }}
        onCancel={(event) => {
          event.preventDefault();
          dismiss();
        }}
        onClose={() => {
          if (!dialog.current?.open) dismiss();
        }}
      >
        {children(dismiss)}
      </dialog>
    </>
  );
}
