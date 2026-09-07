import type { CSSProperties, ReactNode, RefObject } from "react";
import { AccountDialog } from "./AccountDialog";

export function DeploymentUserActionDialog({
  label,
  onClose,
  onSubmit,
  style,
  fallbackFocusRef,
  children,
}: {
  label: string;
  onClose: () => void;
  onSubmit: () => void;
  style: CSSProperties;
  fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
  children: (dismiss: () => void) => ReactNode;
}) {
  return (
    <AccountDialog
      label={label}
      onClose={onClose}
      style={style}
      fallbackFocusRef={fallbackFocusRef}
    >
      {(dismiss) => (
        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault();
            onSubmit();
          }}
          style={{ display: "grid", gap: "var(--ct-spacing-md)" }}
        >
          {children(dismiss)}
        </form>
      )}
    </AccountDialog>
  );
}
