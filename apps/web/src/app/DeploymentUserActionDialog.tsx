import {
  type CSSProperties,
  type ReactNode,
  useLayoutEffect,
  useRef,
} from "react";
import { useAccountDialogFocus } from "./useAccountDialogFocus";

/** Native modal containment with the existing overlay focus/retirement owner. */
export function DeploymentUserActionDialog({
  label,
  onClose,
  onSubmit,
  style,
  children,
}: {
  label: string;
  onClose: () => void;
  onSubmit: () => void;
  style: CSSProperties;
  children: ReactNode;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const navigation = useAccountDialogFocus(() => {
    dialog.current?.close?.();
    onClose();
  });
  useLayoutEffect(() => {
    const element = dialog.current;
    element?.showModal?.();
    // jsdom has no modal dialog implementation; browser evidence covers containment.
    if (element && !element.open) element.setAttribute("open", "");
    element?.querySelector<HTMLButtonElement>("button")?.focus();
    return () => element?.close?.();
  }, []);
  return (
    <dialog
      ref={(element) => {
        dialog.current = element;
        navigation.register(element);
      }}
      aria-label={label}
      aria-modal="true"
      tabIndex={-1}
      style={{
        ...style,
        position: "fixed",
        inset: 0,
        margin: "auto",
        color: "var(--ct-colors-ink)",
        maxHeight: "calc(100% - 2rem)",
        maxWidth: "calc(100% - 2rem)",
        boxSizing: "border-box",
        overflow: "auto",
      }}
      onCancel={(event) => {
        event.preventDefault();
        navigation.close();
      }}
    >
      <form
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          onSubmit();
        }}
        onKeyDown={navigation.onKeyDown}
        onClick={(event) => {
          const button =
            event.target instanceof Element
              ? event.target.closest("button")
              : null;
          if (button?.dataset.dialogClose !== undefined) navigation.close();
        }}
        style={{ display: "grid", gap: "var(--ct-spacing-md)" }}
      >
        {children}
      </form>
    </dialog>
  );
}
