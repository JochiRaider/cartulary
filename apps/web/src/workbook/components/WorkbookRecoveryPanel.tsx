import {
  type RefObject,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import { WorkbookWorkAreaOverlay } from "../../shared/WorkbookWorkAreaOverlay";
import type { WorkbookRecoveryNavigation } from "../../shared/workbookRecoveryNavigation";
import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";

export function WorkbookRecoveryEntry({
  navigation,
  invokerRef,
}: {
  readonly navigation: WorkbookRecoveryNavigation;
  readonly invokerRef: RefObject<HTMLElement | null>;
}) {
  const state = useSyncExternalStore(
    navigation.subscribe,
    navigation.getSnapshot,
  );
  return (
    <Button
      style={{ whiteSpace: "nowrap", flexShrink: 0 }}
      aria-expanded={state.open}
      aria-controls="workbook-recovery-panel"
      data-grid-editor-external-action="true"
      onClick={(event) => {
        invokerRef.current = event.currentTarget;
        if (state.open) navigation.close();
        else navigation.openList();
      }}
    >
      Recovery ({state.count})
    </Button>
  );
}

/** One non-modal shell attachment. Background publications never request focus. */
export function WorkbookRecoveryPanel({
  navigation,
  registerDetailHost,
  invokerRef,
  fallbackRef,
}: {
  readonly navigation: WorkbookRecoveryNavigation;
  readonly registerDetailHost: (node: HTMLDivElement | null) => void;
  readonly invokerRef: RefObject<HTMLElement | null>;
  readonly fallbackRef: RefObject<HTMLElement | null>;
}) {
  const state = useSyncExternalStore(
    navigation.subscribe,
    navigation.getSnapshot,
  );
  const panel = useRef<HTMLElement>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  const focus = useRef<HTMLElement | null>(null);
  const previous = useRef({
    open: false,
    activation: -1,
    selected: state.selected,
  });
  const hadPanelFocus =
    panel.current?.contains(document.activeElement) === true;
  useLayoutEffect(() => {
    const before = previous.current;
    if (state.open && !heading.current) return;
    previous.current = {
      open: state.open,
      activation: state.activation,
      selected: state.selected,
    };
    if (
      state.open &&
      (before.activation !== state.activation ||
        (before.selected !== state.selected && hadPanelFocus) ||
        (focus.current &&
          !focus.current.isConnected &&
          document.activeElement === document.body))
    ) {
      if (panel.current) panel.current.scrollTop = 0;
      heading.current?.focus({ preventScroll: true });
      focus.current = heading.current;
    } else if (
      before.open &&
      !state.open &&
      focus.current &&
      !focus.current.isConnected &&
      document.activeElement === document.body
    ) {
      const invoker = invokerRef.current;
      (invoker?.isConnected && !invoker.closest("[hidden], [inert], [disabled]")
        ? invoker
        : fallbackRef.current
      )?.focus({
        preventScroll: true,
      });
      focus.current = null;
    }
    // Owner portals can detach in a later layout update in the same commit.
    // Repair only removed focus, after those updates, without overriding new work.
    let current = true;
    queueMicrotask(() => {
      if (
        current &&
        navigation.getSnapshot().open &&
        focus.current &&
        !focus.current.isConnected &&
        document.activeElement === document.body
      ) {
        if (panel.current) panel.current.scrollTop = 0;
        heading.current?.focus({ preventScroll: true });
        focus.current = heading.current;
      }
    });
    return () => {
      current = false;
    };
  });
  const focusWorkbook = () => {
    navigation.close();
    const target =
      fallbackRef.current?.querySelector<HTMLElement>(
        '[role="grid"] [tabindex="0"], [role="grid"][tabindex="0"]',
      ) ?? fallbackRef.current;
    target?.focus({ preventScroll: true });
  };
  const selected = state.entries.find((entry) => entry.key === state.selected);
  if (!state.open) return null;
  const group = (
    attention: "attention" | "progress" | "draft" | "completed",
    title: string,
  ) => {
    const entries = state.entries.filter(
      (entry) => entry.attention === attention,
    );
    if (!entries.length) return null;
    return (
      <section aria-label={title}>
        <h3>{title}</h3>
        <ul style={{ paddingInlineStart: "var(--ct-spacing-md)" }}>
          {entries.map((entry) => (
            <li
              key={entry.key}
              style={{
                marginBlock: "var(--ct-spacing-sm)",
                overflowWrap: "anywhere",
              }}
            >
              <Button
                aria-describedby={`recovery-summary-${encodeURIComponent(entry.key)}`}
                onClick={() => navigation.activate(entry.key)}
              >
                {entry.label} · {entry.origin}
              </Button>
              <p
                id={`recovery-summary-${encodeURIComponent(entry.key)}`}
                style={{ marginBlock: "var(--ct-spacing-xs)" }}
              >
                {entry.summary}
              </p>
              {attention === "completed" ? (
                <Button onClick={() => navigation.dismissCompleted(entry.key)}>
                  Dismiss notice for {entry.label}
                </Button>
              ) : null}
            </li>
          ))}
        </ul>
      </section>
    );
  };
  return (
    <WorkbookWorkAreaOverlay ref={panel} label="Workbook recovery">
      <section
        aria-label="Recovery navigation"
        id="workbook-recovery-panel"
        data-grid-editor-external-action="true"
        onFocusCapture={(event) => {
          focus.current = event.target as HTMLElement;
        }}
        onBlurCapture={(event) => {
          // Keep a live focus anchor when an invoked action disables itself.
          if (
            (!event.relatedTarget || event.relatedTarget === document.body) &&
            event.target instanceof HTMLElement &&
            event.target.matches(":disabled")
          ) {
            heading.current?.focus({ preventScroll: true });
            return;
          }
          if (
            event.relatedTarget &&
            event.relatedTarget !== document.body &&
            !event.currentTarget.contains(event.relatedTarget)
          )
            focus.current = null;
        }}
        onKeyDown={(event) => {
          if (event.key !== "Escape" || event.defaultPrevented) return;
          const target = event.target instanceof Element ? event.target : null;
          if (
            target?.closest(
              '[role="dialog"], [role="alertdialog"], [role="listbox"]',
            )
          )
            return;
          const disclosure = target?.closest("details[open]");
          if (disclosure instanceof HTMLDetailsElement) {
            disclosure.open = false;
            disclosure.querySelector("summary")?.focus();
          } else navigation.close();
          event.preventDefault();
          event.stopPropagation();
        }}
      >
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "var(--ct-spacing-sm)",
          }}
        >
          {selected ? (
            <Button onClick={navigation.openList}>All recovery</Button>
          ) : null}
          <Button onClick={navigation.close}>Close recovery</Button>
          <Button onClick={focusWorkbook}>Return to workbook</Button>
        </div>
        <style>{`.workbook-recovery-heading:focus { outline: var(--ct-component-focus-ring-border); outline-offset: var(--ct-component-focus-ring-offset); }`}</style>
        <h2 className="workbook-recovery-heading" ref={heading} tabIndex={-1}>
          {selected?.label ?? "Recovery"}
        </h2>
        {state.message ? <p role="status">{state.message}</p> : null}
        {selected ? (
          <p>
            {selected.origin} · {selected.summary}
          </p>
        ) : (
          <>
            {state.count === 0 ? <p>No unfinished recovery work.</p> : null}
            {group("attention", "Needs attention")}
            {group("progress", "In progress")}
            {group("draft", "Retained drafts")}
            {state.entries.some((entry) => entry.attention === "completed") ? (
              <details>
                <summary>Completed</summary>
                {group("completed", "Completed notices")}
              </details>
            ) : null}
          </>
        )}
        <div
          ref={registerDetailHost}
          style={{ overflowWrap: "anywhere", minInlineSize: 0 }}
        />
      </section>
    </WorkbookWorkAreaOverlay>
  );
}
