import {
  createContext,
  forwardRef,
  type ReactNode,
  useCallback,
  useContext,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";

type Attachment = {
  readonly key: symbol;
  readonly close: (reason?: "navigation") => void;
  readonly label: string;
};
const Context = createContext<{
  readonly attachment: Attachment | null;
  readonly host: HTMLDivElement | null;
  readonly setHost: (node: HTMLDivElement | null) => void;
  readonly attach: (attachment: Attachment) => void;
  readonly detach: (key: symbol) => void;
} | null>(null);
const FooterContext = createContext<{
  readonly host: HTMLDivElement | null;
  readonly setHost: (node: HTMLDivElement | null) => void;
} | null>(null);

/** Presentation placement only; Recovery and source owners retain operation state. */
export function WorkbookAuxiliaryDockProvider({
  children,
}: {
  readonly children: ReactNode;
}) {
  const [attachment, setAttachment] = useState<Attachment | null>(null);
  const attach = useCallback((next: Attachment) => setAttachment(next), []);
  const detach = useCallback(
    (key: symbol) =>
      setAttachment((current) => (current?.key === key ? null : current)),
    [],
  );
  const [host, setHost] = useState<HTMLDivElement | null>(null);
  const [footer, setFooter] = useState<HTMLDivElement | null>(null);
  const value = useMemo(
    () => ({ attachment, host, setHost, attach, detach }),
    [attachment, host, attach, detach],
  );
  const footerValue = useMemo(
    () => ({ host: footer, setHost: setFooter }),
    [footer],
  );
  return (
    <Context value={value}>
      <FooterContext value={footerValue}>{children}</FooterContext>
    </Context>
  );
}
export const useWorkbookAuxiliaryDock = () => useContext(Context);
export function WorkbookFooterNavigationHost() {
  const context = useContext(FooterContext);
  return context ? (
    <div
      ref={context.setHost}
      style={{
        display: "flex",
        alignItems: "center",
        minWidth: 0,
        flex: "1 1 auto",
        gap: "var(--ct-spacing-xs)",
      }}
    />
  ) : null;
}
export function WorkbookFooterNavigation({
  children,
}: {
  readonly children: ReactNode;
}) {
  const context = useContext(FooterContext);
  return context
    ? context.host
      ? createPortal(children, context.host)
      : null
    : children;
}
export const WorkbookAuxiliaryDock = forwardRef<
  HTMLElement,
  {
    readonly children: ReactNode;
    readonly label: string;
    readonly onClose: () => void;
    readonly onNavigationClose?: (() => void) | undefined;
  }
>(function WorkbookAuxiliaryDock(
  { children, label, onClose, onNavigationClose },
  ref,
) {
  const context = useContext(Context);
  if (!context) throw new Error("WorkbookAuxiliaryDockProvider is required");
  const key = useRef(Symbol("auxiliary destination")).current;
  const close = useRef(onClose);
  close.current = onClose;
  const navigationClose = useRef(onNavigationClose);
  navigationClose.current = onNavigationClose;
  const { attach, detach } = context;
  useLayoutEffect(() => {
    attach({
      key,
      label,
      close: (reason) => {
        if (reason === "navigation" && navigationClose.current)
          navigationClose.current();
        else close.current();
      },
    });
    return () => detach(key);
  }, [attach, detach, key, label]);
  return context.host
    ? createPortal(
        <section
          ref={ref}
          aria-label={label}
          onKeyDown={(event) => {
            if (event.key !== "Escape" || event.defaultPrevented) return;
            event.preventDefault();
            event.stopPropagation();
            close.current();
          }}
          style={{
            minWidth: 0,
            minHeight: 0,
            height: "100%",
            overflowX: "hidden",
            overflowY: "auto",
            background: "var(--ct-colors-surface-1)",
            padding: "var(--ct-spacing-panel-padding)",
            boxSizing: "border-box",
          }}
        >
          {children}
        </section>,
        context.host,
      )
    : null;
});
