import {
  createContext,
  forwardRef,
  type ReactNode,
  useContext,
  useLayoutEffect,
  useMemo,
  useRef,
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

/** Placement, entry focus and scrolling belong to the work-area boundary. */
export const WorkbookWorkAreaOverlay = forwardRef<
  HTMLElement,
  { readonly children: ReactNode; readonly label: string }
>(function WorkbookWorkAreaOverlay({ children, label }, ref) {
  const context = useContext(OverlayContext);
  const panel = useRef<HTMLElement>(null);
  const entered = useRef(false);
  useLayoutEffect(() => {
    if (context?.host === null || context === null || entered.current) return;
    entered.current = true;
    panel.current?.focus({ preventScroll: true });
  }, [context]);
  return context?.host
    ? createPortal(
        <section
          aria-label={label}
          tabIndex={-1}
          ref={(node) => {
            panel.current = node;
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
  zIndex: 12,
};
const panelStyle = {
  position: "absolute" as const,
  insetBlock: "var(--ct-spacing-sm)",
  insetInlineEnd: "var(--ct-spacing-sm)",
  inlineSize: "min(28rem, 100%)",
  maxInlineSize: "calc(100% - var(--ct-spacing-sm) - var(--ct-spacing-sm))",
  minBlockSize: 0,
  overflow: "auto",
  pointerEvents: "auto" as const,
  boxSizing: "border-box" as const,
  padding: "var(--ct-spacing-panel-padding)",
  background: "var(--ct-colors-surface-1)",
  border: "var(--ct-border-hairline)",
  boxShadow: "var(--ct-elevation-popover)",
};
