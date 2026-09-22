import { useState } from "react";
import {
  workbookQuietCommandStyle,
  workbookTypography,
} from "../components/workbookFormStyles";
import type { WorkbookBrowseAction } from "./WorkbookQueryBrowser";
import { useWorkbookQueryPresentation } from "./WorkbookQueryBrowsingContext";

/** Explicit navigation within the workbook's owned work area. */
export function WorkbookQueryBrowsingControls({
  viewSchemaId,
}: {
  readonly viewSchemaId: string;
}) {
  const [focusedAction, setFocusedAction] =
    useState<WorkbookBrowseAction | null>(null);
  const registry = useWorkbookQueryPresentation();
  const browser = registry?.find(viewSchemaId);
  if (!browser || !registry) return null;
  const state = browser.getSnapshot();
  const pending = state.pending !== null;
  const unapplied =
    state.requested !== null && browser.hasUnapplied(state.requested);
  const activate = (action: WorkbookBrowseAction) => {
    void registry.activate(viewSchemaId, action);
  };
  const button = (
    label: string,
    action: WorkbookBrowseAction,
    unavailable = false,
  ) => (
    <button
      data-grid-editor-external-action="true"
      onFocus={() => setFocusedAction(action)}
      onBlur={() => setFocusedAction(null)}
      aria-busy={
        state.pending === action ||
        (action === "restart" && state.pending === "replace") ||
        undefined
      }
      type="button"
      aria-disabled={pending || unavailable}
      onPointerDown={() => {
        if (!pending && !unavailable) registry.prepareBrowse(viewSchemaId);
      }}
      onClick={() => {
        if (!pending && !unavailable) activate(action);
      }}
      style={workbookQuietCommandStyle}
    >
      {label}
    </button>
  );
  return (
    <fieldset
      aria-label="Workbook browsing"
      style={{
        ...workbookTypography("metadata"),
        boxSizing: "border-box",
        margin: 0,
        minWidth: 0,
        flex: "0 0 auto",
        minBlockSize: "var(--ct-layout-queryFooterHeight)",
        border: 0,
        display: "flex",
        alignItems: "center",
        gap: "var(--ct-spacing-xs)",
        flexWrap: "wrap",
        padding: "0 var(--ct-spacing-sm)",
        color: "var(--ct-colors-ink-muted)",
        background: "var(--ct-colors-surface-1)",
        borderTop: "var(--ct-border-hairline)",
      }}
    >
      {state.hasEarlier ||
      state.pending === "earlier" ||
      focusedAction === "earlier"
        ? button("Earlier rows", "earlier", !state.hasEarlier || unapplied)
        : null}
      {state.canLoadMore || state.pending === "more" || focusedAction === "more"
        ? button("Load more", "more", !state.canLoadMore || unapplied)
        : null}
      {button("Refresh", "restart")}
      <span
        role="status"
        aria-live={pending ? "off" : "polite"}
        aria-atomic="true"
      >
        {pending
          ? `${state.accepted ? `${state.accepted.rows.length} records loaded. ` : ""}${state.pending === "more" ? "Loading more records…" : "Loading records…"}`
          : state.accepted
            ? `${state.accepted.rows.length} records loaded${state.canLoadMore ? "; more available" : state.accepted.paging.hasMore ? "; refresh required" : "; end of current results"}.`
            : "No accepted results."}
      </span>
      {state.failure ? (
        <>
          <span>
            {unapplied
              ? "Query changes are unapplied."
              : "Read failed. Retry or refresh."}
          </span>
          {button("Retry", "retry")}
          {unapplied ? (
            <button
              data-grid-editor-external-action="true"
              style={workbookQuietCommandStyle}
              type="button"
              onClick={() => registry.revert(viewSchemaId)}
            >
              Revert
            </button>
          ) : null}
        </>
      ) : null}
      {!state.hasEarlier && state.earlierEvicted ? (
        <span>
          Earlier history limit reached. Refresh to return to the beginning.
        </span>
      ) : null}
    </fieldset>
  );
}
