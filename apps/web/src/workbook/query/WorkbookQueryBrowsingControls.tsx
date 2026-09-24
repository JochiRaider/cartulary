import { useEffect, useId, useState } from "react";
import {
  workbookQuietCommandStyle,
  workbookTypography,
} from "../components/workbookFormStyles";
import type {
  WorkbookBrowseAction,
  WorkbookQueryBrowser,
} from "./WorkbookQueryBrowser";
import { useWorkbookQueryPresentation } from "./WorkbookQueryBrowsingContext";

type FooterAction = WorkbookBrowseAction | "revert";
type FocusedAction = {
  readonly action: FooterAction;
  readonly browser: WorkbookQueryBrowser;
  readonly focusEpoch: number;
  readonly viewSchemaId: string;
};

/** Explicit navigation within the workbook's owned work area. */
export function WorkbookQueryBrowsingControls({
  viewSchemaId,
}: {
  readonly viewSchemaId: string;
}) {
  const [focusedAction, setFocusedAction] = useState<FocusedAction | null>(
    null,
  );
  useEffect(
    () =>
      setFocusedAction((current) =>
        current?.viewSchemaId === viewSchemaId ? current : null,
      ),
    [viewSchemaId],
  );
  const statusId = useId();
  const registry = useWorkbookQueryPresentation();
  const browser = registry.find(viewSchemaId);
  if (!browser) return null;
  const state = browser.getSnapshot();
  const pending = state.pending !== null;
  const pendingDescription =
    state.pendingAction === "retry"
      ? "Retrying records…"
      : state.pending === "restart"
        ? "Refreshing records…"
        : state.pendingAction === "more"
          ? "Loading more records…"
          : state.pendingAction === "earlier"
            ? "Loading earlier records…"
            : "Loading records…";
  const unapplied =
    state.requested !== null && browser.hasUnapplied(state.requested);
  const focused = (action: FooterAction) =>
    focusedAction?.action === action &&
    focusedAction.browser === browser &&
    focusedAction.focusEpoch === state.focusEpoch &&
    focusedAction.viewSchemaId === viewSchemaId;
  const onFocus = (action: FooterAction) =>
    setFocusedAction({
      action,
      browser,
      focusEpoch: state.focusEpoch,
      viewSchemaId,
    });
  const onBlur = (action: FooterAction) =>
    setFocusedAction((current) =>
      current?.action === action ? null : current,
    );
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
      onFocus={() => onFocus(action)}
      onBlur={() => onBlur(action)}
      aria-busy={state.pendingAction === action || undefined}
      aria-describedby={state.pendingAction === action ? statusId : undefined}
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
      state.pendingAction === "earlier" ||
      focused("earlier")
        ? button("Earlier rows", "earlier", !state.hasEarlier || unapplied)
        : null}
      {state.canLoadMore || state.pendingAction === "more" || focused("more")
        ? button("Load more", "more", !state.canLoadMore || unapplied)
        : null}
      {button("Refresh", "restart")}
      <span
        id={statusId}
        role="status"
        aria-live={pending || state.failure ? "off" : "polite"}
        aria-atomic="true"
      >
        {pending
          ? `${state.accepted ? `${state.accepted.rows.length} records loaded. ` : ""}${pendingDescription}`
          : state.accepted
            ? `${state.accepted.rows.length} records loaded${state.canLoadMore ? "; more available" : state.accepted.paging.hasMore ? "; refresh required" : "; end of current results"}.`
            : "No accepted results."}
      </span>
      {state.failure ? (
        <span>
          {unapplied
            ? "Query changes are unapplied."
            : "Read failed. Retry or refresh."}
        </span>
      ) : null}
      {state.failure || state.pendingAction === "retry" || focused("retry")
        ? button("Retry", "retry", state.failure === null)
        : null}
      {(state.failure && unapplied) || focused("revert") ? (
        <button
          data-grid-editor-external-action="true"
          style={workbookQuietCommandStyle}
          type="button"
          aria-disabled={pending || !state.failure || !unapplied}
          onFocus={() => onFocus("revert")}
          onBlur={() => onBlur("revert")}
          onClick={() => {
            if (!pending && state.failure && unapplied)
              registry.revert(viewSchemaId);
          }}
        >
          Revert
        </button>
      ) : null}
      {!state.hasEarlier && state.earlierEvicted ? (
        <span>
          Earlier history limit reached. Refresh to return to the beginning.
        </span>
      ) : null}
    </fieldset>
  );
}
