import {
  savedViewOptionTestId,
  savedViewSelectorTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type RefObject,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { ActiveSurfaceSavedViewProjection } from "../models/workbookSavedViewControl";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";
import { visuallyHiddenStyle } from "../utils/workbookStyles";
import { controlButtonStyle, menuStyle } from "./workbookGridControlStyles";

/** Focus moves through candidates; only activation resolves and applies a resource. */
export function SavedViewBrowser({
  controller,
  schema,
  projection,
  triggerRef,
  triggerStyle,
  onBase,
}: {
  controller: WorkbookSavedViewController;
  schema: string;
  projection: ActiveSurfaceSavedViewProjection;
  triggerRef: RefObject<HTMLButtonElement | null>;
  triggerStyle: CSSProperties;
  onBase: () => void;
}) {
  const snapshot = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const page = snapshot.discovery;
  const panel = useRef<HTMLDivElement>(null);
  const options = useRef(new Map<number, HTMLButtonElement>());
  const [focus, setFocus] = useState(0);
  const candidateHadFocus = useRef(false);
  const open = page.open && page.viewSchemaId === schema;
  const close = (returnFocus: boolean) => {
    controller.closeDiscovery();
    if (returnFocus) triggerRef.current?.focus({ preventScroll: true });
  };
  useLayoutEffect(() => {
    if (!open || !panel.current || !triggerRef.current) return;
    const popup = panel.current;
    const trigger = triggerRef.current;
    popup.showPopover?.();
    const position = () => {
      const viewport = window.visualViewport;
      const left = viewport?.offsetLeft ?? 0;
      const top = viewport?.offsetTop ?? 0;
      const width = viewport?.width ?? window.innerWidth;
      const height = viewport?.height ?? window.innerHeight;
      const anchor = trigger.getBoundingClientRect();
      // Rectangles use viewport pixels, while fixed offsets inherit CSS zoom.
      const zoom = popup.offsetWidth
        ? popup.getBoundingClientRect().width / popup.offsetWidth
        : 1;
      popup.style.maxInlineSize = `${(width * 0.96) / zoom}px`;
      popup.style.maxBlockSize = `${(height * 0.8) / zoom}px`;
      const bounds = popup.getBoundingClientRect();
      popup.style.left = `${Math.max(left, Math.min(anchor.left, left + width - bounds.width)) / zoom}px`;
      popup.style.top = `${Math.max(top, Math.min(anchor.bottom, top + height - bounds.height)) / zoom}px`;
    };
    position();
    setFocus(0);
    options.current.get(0)?.focus({ preventScroll: true });
    const observer = new ResizeObserver(position);
    observer.observe(popup);
    window.addEventListener("resize", position);
    window.addEventListener("scroll", position, true);
    window.visualViewport?.addEventListener("resize", position);
    return () => {
      observer.disconnect();
      window.removeEventListener("resize", position);
      window.removeEventListener("scroll", position, true);
      window.visualViewport?.removeEventListener("resize", position);
    };
  }, [open, triggerRef]);
  useLayoutEffect(() => {
    const index = Math.min(focus, page.candidates.length);
    if (index !== focus) setFocus(index);
    if (
      open &&
      candidateHadFocus.current &&
      document.activeElement === document.body
    )
      options.current.get(index)?.focus();
  }, [page.candidates, focus, open]);
  const label =
    projection.selectedSavedView?.display_name ??
    (projection.selectedSavedViewId ? "Selected saved view" : "Unsaved view");
  const status = page.problem
    ? `${page.accepted ? "Showing the previously accepted page. " : ""}${page.problem.message}`
    : page.pending
      ? page.accepted
        ? "Loading another page. Current choices remain available."
        : "Loading saved views…"
      : page.accepted
        ? page.candidates.length
          ? `${page.candidates.length} saved views on this page.${page.nextCursor ? " More pages are available." : " End of results."}`
          : page.cursor
            ? "No saved views on this page. First refreshes discovery."
            : "No saved views for this surface. Unsaved view remains available."
        : "";
  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        aria-label="Saved view"
        aria-describedby={`${savedViewSelectorTestId(schema)}-description`}
        aria-haspopup="dialog"
        aria-expanded={open}
        data-grid-editor-external-action="true"
        data-testid={savedViewSelectorTestId(schema)}
        data-active-view-schema-id={schema}
        data-resource-kind={projection.resourceKind}
        data-selected-saved-view-id={projection.selectedSavedViewId}
        data-selected-sheet-ref-kind={
          projection.selectedSavedViewId ? "saved_view" : "view_schema"
        }
        title={label}
        style={{
          ...triggerStyle,
          display: "flex",
          alignItems: "center",
          gap: "0.25rem",
          overflow: "hidden",
          whiteSpace: "nowrap",
          textAlign: "start",
        }}
        onClick={() => (open ? close(false) : controller.openDiscovery())}
      >
        <span
          style={{
            minInlineSize: 0,
            flex: "1 1 auto",
            overflow: "hidden",
            textOverflow: "ellipsis",
          }}
        >
          {label}
        </span>
        {projection.selectedSavedView ? (
          <span style={{ ...scopeStyle, fontSize: "0.7rem" }}>
            {projection.selectedSavedView.scope}
          </span>
        ) : null}
        <span aria-hidden="true" style={{ flexShrink: 0 }}>
          ▾
        </span>
      </button>
      <span
        id={`${savedViewSelectorTestId(schema)}-description`}
        style={visuallyHiddenStyle}
      >
        {label}
        {projection.selectedSavedView
          ? `, ${projection.selectedSavedView.scope} scope`
          : ""}
      </span>
      {open ? (
        <div
          ref={panel}
          popover="auto"
          role="dialog"
          aria-label="Saved views"
          onFocusCapture={(event) => {
            candidateHadFocus.current =
              (event.target as HTMLElement).getAttribute("role") === "option";
          }}
          data-grid-editor-external-action="true"
          style={{
            ...menuStyle,
            position: "fixed",
            inset: "auto",
            margin: 0,
            boxSizing: "border-box",
            color: "var(--ct-colors-ink)",
            inlineSize: "min(26rem, 96vw)",
            maxInlineSize: "96vw",
            maxBlockSize: "80vh",
            overflow: "auto",
          }}
          onToggle={(event) => {
            if (event.newState === "closed") close(false);
          }}
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.preventDefault();
              event.stopPropagation();
              close(true);
            }
          }}
        >
          <div
            role="listbox"
            aria-label="Saved-view choices"
            onKeyDown={(event) => {
              let index: number;
              if (event.key === "ArrowDown")
                index = Math.min(page.candidates.length, focus + 1);
              else if (event.key === "ArrowUp") index = Math.max(0, focus - 1);
              else if (event.key === "Home") index = 0;
              else if (event.key === "End") index = page.candidates.length;
              else return;
              event.preventDefault();
              event.stopPropagation();
              setFocus(index);
              options.current.get(index)?.focus();
            }}
          >
            <button
              ref={(node) => {
                if (node) options.current.set(0, node);
                else options.current.delete(0);
              }}
              type="button"
              role="option"
              data-testid={savedViewOptionTestId(schema, "base")}
              aria-selected={!projection.selectedSavedViewId}
              tabIndex={focus === 0 ? 0 : -1}
              style={{
                ...optionStyle,
                ...(!projection.selectedSavedViewId ? selectedOptionStyle : {}),
              }}
              onFocus={() => setFocus(0)}
              onClick={() => {
                onBase();
                close(true);
              }}
            >
              Unsaved view
            </button>
            {page.candidates.map((candidate, index) => (
              <button
                key={candidate.saved_view_id}
                ref={(node) => {
                  if (node) options.current.set(index + 1, node);
                  else options.current.delete(index + 1);
                }}
                type="button"
                role="option"
                aria-selected={
                  projection.selectedSavedViewId === candidate.saved_view_id
                }
                tabIndex={focus === index + 1 ? 0 : -1}
                style={{
                  ...optionStyle,
                  ...(projection.selectedSavedViewId === candidate.saved_view_id
                    ? selectedOptionStyle
                    : {}),
                }}
                onFocus={() => setFocus(index + 1)}
                data-testid={savedViewOptionTestId(
                  schema,
                  candidate.saved_view_id,
                )}
                data-saved-view-id={candidate.saved_view_id}
                data-view-schema-id={schema}
                onClick={() => {
                  void controller
                    .activateResource(candidate.saved_view_id, schema)
                    .then((applied) => {
                      if (applied) close(true);
                    });
                }}
              >
                <span style={{ overflowWrap: "anywhere" }}>
                  {candidate.display_name}
                </span>
                <span style={scopeStyle}>{candidate.scope}</span>
              </button>
            ))}
          </div>
          <p
            role="status"
            aria-live="polite"
            style={{ fontSize: "0.8rem", margin: "0.4rem 0" }}
          >
            {snapshot.activationId
              ? "Loading selected saved view. Your current configuration remains active."
              : status}
          </p>
          {snapshot.notice ? (
            <p style={{ fontSize: "0.8rem" }}>{snapshot.notice}</p>
          ) : null}
          <div style={{ display: "flex", flexWrap: "wrap", gap: "0.35rem" }}>
            <button
              type="button"
              style={controlButtonStyle}
              disabled={page.pending}
              onClick={() => void controller.discovery.first()}
            >
              First
            </button>
            <button
              type="button"
              style={controlButtonStyle}
              disabled={page.pending || !page.previous.length}
              onClick={() => void controller.discovery.previous()}
            >
              Previous
            </button>
            <button
              type="button"
              style={controlButtonStyle}
              disabled={page.pending || !page.nextCursor}
              onClick={() => void controller.discovery.next()}
            >
              Next
            </button>
            <button
              type="button"
              style={controlButtonStyle}
              disabled={page.pending}
              onClick={() => void controller.discovery.first()}
            >
              Refresh
            </button>
            {page.problem ? (
              <button
                type="button"
                style={controlButtonStyle}
                disabled={page.pending}
                onClick={() => void controller.discovery.retry()}
              >
                Retry page
              </button>
            ) : null}
            <button
              type="button"
              style={controlButtonStyle}
              onClick={() => close(true)}
            >
              Close
            </button>
          </div>
        </div>
      ) : null}
    </>
  );
}
const optionStyle = {
  display: "flex",
  alignItems: "baseline",
  justifyContent: "space-between",
  gap: "0.6rem",
  width: "100%",
  textAlign: "start" as const,
  padding: "0.3rem 0.45rem",
  border: "none",
  borderRadius: "var(--ct-rounded-xs)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  font: "inherit",
  cursor: "pointer",
};
const scopeStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
  flexShrink: 0,
};

const selectedOptionStyle = {
  background: "var(--ct-colors-surface-3)",
  fontWeight: 700,
};
