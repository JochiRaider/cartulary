import {
  savedViewOptionTestId,
  savedViewSelectorTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type RefObject,
  useId,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { ActiveSurfaceSavedViewProjection } from "../models/workbookSavedViewControl";
import type { SavedViewDiscoveryAction } from "../savedviews/SavedViewDiscovery";
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
  const options = useRef(new Map<string, HTMLButtonElement>());
  const actions = useRef(
    new Map<SavedViewDiscoveryAction, HTMLButtonElement>(),
  );
  const [focusedCandidateId, setFocusedCandidateId] = useState("base");
  const focusedCandidateIndex = useRef(0);
  const focusOwner = useRef<"candidate" | "action" | null>(null);
  const focusedAction = useRef<SavedViewDiscoveryAction | null>(null);
  const priorPendingAction = useRef<SavedViewDiscoveryAction | null>(null);
  const statusId = useId();
  const open = page.open && page.viewSchemaId === schema;
  const close = (returnFocus: boolean) => {
    controller.closeDiscovery();
    if (returnFocus) triggerRef.current?.focus({ preventScroll: true });
  };
  useLayoutEffect(() => {
    if (!open || !panel.current || !triggerRef.current) {
      focusOwner.current = null;
      focusedAction.current = null;
      priorPendingAction.current = null;
      return;
    }
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
    setFocusedCandidateId("base");
    focusedCandidateIndex.current = 0;
    options.current.get("base")?.focus({ preventScroll: true });
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
    if (!open || focusedCandidateId === "base") return;
    const retainedIndex = page.candidates.findIndex(
      (candidate) => candidate.saved_view_id === focusedCandidateId,
    );
    if (retainedIndex >= 0) {
      focusedCandidateIndex.current = retainedIndex + 1;
      return;
    }
    const fallbackIndex = Math.min(
      focusedCandidateIndex.current,
      page.candidates.length,
    );
    const fallbackId =
      page.candidates[fallbackIndex - 1]?.saved_view_id ?? "base";
    setFocusedCandidateId(fallbackId);
    focusedCandidateIndex.current = fallbackIndex;
    if (
      focusOwner.current === "candidate" &&
      document.activeElement === document.body
    )
      options.current.get(fallbackId)?.focus({ preventScroll: true });
  }, [page.candidates, focusedCandidateId, open]);
  useLayoutEffect(() => {
    const settledAction = priorPendingAction.current;
    priorPendingAction.current = page.pendingAction;
    if (!open || page.pending || !settledAction) return;
    if (
      focusOwner.current !== "action" ||
      focusedAction.current !== settledAction
    )
      return;
    const source = actions.current.get(settledAction);
    if (source?.isConnected && !source.disabled) return;
    if (
      document.activeElement !== document.body &&
      document.activeElement !== source
    )
      return;
    const fallbackId = page.candidates[0]?.saved_view_id ?? "base";
    setFocusedCandidateId(fallbackId);
    focusedCandidateIndex.current = fallbackId === "base" ? 0 : 1;
    options.current.get(fallbackId)?.focus({ preventScroll: true });
  }, [page.pending, page.pendingAction, page.candidates, open]);
  const pagingButton = (
    label: string,
    action: SavedViewDiscoveryAction,
    available: boolean,
    activate: () => void,
  ) => {
    const busy = page.pendingAction === action;
    const unavailable = !busy && (page.pending || !available);
    const stateStyle = busy
      ? {
          background: "var(--ct-colors-surface-2)",
          color: "var(--ct-colors-ink-muted)",
          border: "var(--ct-border-hairline)",
          cursor: "progress",
        }
      : unavailable
        ? {
            background: "var(--ct-colors-surface-2)",
            color: "var(--ct-colors-ink-tertiary)",
            border: "var(--ct-border-hairline)",
            cursor: "not-allowed",
          }
        : {};
    return (
      <button
        ref={(node) => {
          if (node) actions.current.set(action, node);
          else actions.current.delete(action);
        }}
        type="button"
        data-discovery-action={action}
        style={{ ...controlButtonStyle, ...stateStyle }}
        disabled={unavailable}
        aria-disabled={unavailable || undefined}
        aria-busy={busy || undefined}
        aria-describedby={busy ? statusId : undefined}
        onClick={() => {
          if (!page.pending && available) activate();
        }}
      >
        {label}
        {busy ? <span aria-hidden="true">…</span> : null}
      </button>
    );
  };
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
            const target = event.target as HTMLElement;
            const action = target.getAttribute("data-discovery-action");
            if (target.getAttribute("data-saved-view-choice-id") !== null) {
              focusOwner.current = "candidate";
              focusedAction.current = null;
            } else if (action) {
              focusOwner.current = "action";
              focusedAction.current = action as SavedViewDiscoveryAction;
            } else {
              focusOwner.current = null;
              focusedAction.current = null;
            }
          }}
          onBlurCapture={(event) => {
            if (
              event.relatedTarget instanceof Node &&
              !panel.current?.contains(event.relatedTarget)
            ) {
              focusOwner.current = null;
              focusedAction.current = null;
            }
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
              const ids = [
                "base",
                ...page.candidates.map((candidate) => candidate.saved_view_id),
              ];
              const currentIndex = Math.max(0, ids.indexOf(focusedCandidateId));
              let index: number;
              if (event.key === "ArrowDown")
                index = Math.min(page.candidates.length, currentIndex + 1);
              else if (event.key === "ArrowUp")
                index = Math.max(0, currentIndex - 1);
              else if (event.key === "Home") index = 0;
              else if (event.key === "End") index = page.candidates.length;
              else return;
              event.preventDefault();
              event.stopPropagation();
              const id = ids[index];
              if (!id) return;
              setFocusedCandidateId(id);
              focusedCandidateIndex.current = index;
              options.current.get(id)?.focus();
            }}
          >
            <button
              ref={(node) => {
                if (node) options.current.set("base", node);
                else options.current.delete("base");
              }}
              type="button"
              role="option"
              data-saved-view-choice-id="base"
              data-testid={savedViewOptionTestId(schema, "base")}
              aria-selected={!projection.selectedSavedViewId}
              tabIndex={focusedCandidateId === "base" ? 0 : -1}
              style={{
                ...optionStyle,
                ...(!projection.selectedSavedViewId ? selectedOptionStyle : {}),
              }}
              onFocus={() => {
                setFocusedCandidateId("base");
                focusedCandidateIndex.current = 0;
              }}
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
                  if (node) options.current.set(candidate.saved_view_id, node);
                  else options.current.delete(candidate.saved_view_id);
                }}
                type="button"
                role="option"
                data-saved-view-choice-id={candidate.saved_view_id}
                aria-selected={
                  projection.selectedSavedViewId === candidate.saved_view_id
                }
                tabIndex={
                  focusedCandidateId === candidate.saved_view_id ? 0 : -1
                }
                style={{
                  ...optionStyle,
                  ...(projection.selectedSavedViewId === candidate.saved_view_id
                    ? selectedOptionStyle
                    : {}),
                }}
                onFocus={() => {
                  setFocusedCandidateId(candidate.saved_view_id);
                  focusedCandidateIndex.current = index + 1;
                }}
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
            id={statusId}
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
            {pagingButton("First", "first", true, () => {
              void controller.discovery.first();
            })}
            {pagingButton(
              "Previous",
              "previous",
              !!page.previous.length,
              () => {
                void controller.discovery.previous();
              },
            )}
            {pagingButton("Next", "next", page.nextCursor !== null, () => {
              void controller.discovery.next();
            })}
            {pagingButton("Refresh", "refresh", true, () => {
              void controller.discovery.refresh();
            })}
            {page.problem || page.pendingAction === "retry"
              ? pagingButton("Retry page", "retry", !!page.problem, () => {
                  void controller.discovery.retry();
                })
              : null}
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
