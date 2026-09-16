import {
  createContext,
  forwardRef,
  type ReactNode,
  useContext,
  useMemo,
  useState,
} from "react";
import { createPortal } from "react-dom";

const OverlayContext = createContext<{
  readonly host: HTMLDivElement | null;
  readonly register: (host: HTMLDivElement | null) => void;
} | null>(null);

export function WorkbookWorkAreaOverlayProvider({
  children,
}: {
  readonly children: ReactNode;
}) {
  const [host, register] = useState<HTMLDivElement | null>(null);
  const value = useMemo(() => ({ host, register }), [host]);
  return (
    <OverlayContext.Provider value={value}>{children}</OverlayContext.Provider>
  );
}

export function WorkbookWorkAreaOverlayHost() {
  const context = useContext(OverlayContext);
  return <div ref={context?.register} style={hostStyle} />;
}

/** Placement and scrolling belong to the work-area boundary. */
export const WorkbookWorkAreaOverlay = forwardRef<
  HTMLElement,
  { readonly children: ReactNode; readonly label: string }
>(function WorkbookWorkAreaOverlay({ children, label }, ref) {
  const context = useContext(OverlayContext);
  return context?.host
    ? createPortal(
        <section
          aria-label={label}
          tabIndex={-1}
          ref={(node) => {
            if (typeof ref === "function") ref(node);
            else if (ref !== null) ref.current = node;
          }}
          style={panelStyle}
        >
          {children}
        </section>,
        context.host,
      )
    : null;
});

const hostStyle = {
  position: "absolute" as const,
  inset: 0,
  pointerEvents: "none" as const,
  // Work-area panels sit below the existing view-bar navigation layer (9).
  zIndex: 8,
};
const panelStyle = {
  position: "absolute" as const,
  insetBlock: "var(--ct-spacing-sm)",
  insetInlineEnd: "var(--ct-spacing-sm)",
  inlineSize: "min(var(--ct-layout-recoveryMaxWidth), 100%)",
  maxInlineSize: "calc(100% - var(--ct-spacing-sm) - var(--ct-spacing-sm))",
  minBlockSize: 0,
  overflowY: "auto" as const,
  overflowX: "hidden" as const,
  pointerEvents: "auto" as const,
  boxSizing: "border-box" as const,
  padding: "var(--ct-spacing-panel-padding)",
  background: "var(--ct-colors-surface-1)",
  border: "var(--ct-border-hairline)",
  boxShadow: "var(--ct-elevation-popover)",
};
