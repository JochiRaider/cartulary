import { type KeyboardEvent, useRef } from "react";
import { useRegisteredOverlayNavigation } from "../shared/useRegisteredOverlayNavigation";

/** Account dialog controls retain native form keys; Tab/Escape use the shared overlay owner. */
export function useAccountDialogFocus(onClose: () => void) {
  const triggerRef = useRef(
    document.activeElement instanceof HTMLElement
      ? document.activeElement
      : null,
  );
  const keys = useRef<string[]>([]);
  const controls = useRef<HTMLElement[]>([]);
  const navigation = useRegisteredOverlayNavigation({
    isOpen: true,
    initialItemKey: null,
    itemKeys: keys.current,
    triggerRef,
    subjectKey: "account-dialog",
    trapTab: true,
    reconcileItems: true,
    restoreFocusOnSubjectChange: false,
    onRequestClose: onClose,
  });
  const register = (node: HTMLElement | null) => {
    for (const key of keys.current) navigation.registerItem(key)(null);
    keys.current.length = 0;
    controls.current = node
      ? [
          ...node.querySelectorAll<HTMLElement>(
            'button, input, select, textarea, a[href], [tabindex="0"]',
          ),
        ]
      : [];
    controls.current = controls.current.filter(
      (control) =>
        control.tabIndex >= 0 &&
        !(
          control instanceof HTMLInputElement &&
          control.type === "radio" &&
          !control.checked
        ),
    );
    for (const [index, control] of controls.current.entries()) {
      const key = String(index);
      keys.current.push(key);
      navigation.registerItem(key)(control);
    }
  };
  const onKeyDown = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key !== "Tab" && event.key !== "Escape") return;
    const key = String(
      controls.current.indexOf(document.activeElement as HTMLElement),
    );
    navigation.onItemKeyDown(event, key);
  };
  return {
    register,
    onKeyDown,
    close: () => navigation.close({ restoreTriggerFocus: true }),
  };
}
