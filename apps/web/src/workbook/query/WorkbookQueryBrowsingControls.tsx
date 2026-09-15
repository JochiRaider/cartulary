import type { WorkbookBrowseAction } from "./WorkbookQueryBrowser";
import { useWorkbookQueryPresentation } from "./WorkbookQueryBrowsingContext";

/** Explicit navigation within the workbook's owned work area. */
export function WorkbookQueryBrowsingControls({
  viewSchemaId,
}: {
  readonly viewSchemaId: string;
}) {
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
      type="button"
      aria-disabled={pending || unavailable}
      onPointerDown={() => {
        if (!pending && !unavailable) registry.prepareBrowse(viewSchemaId);
      }}
      onClick={() => {
        if (!pending && !unavailable) activate(action);
      }}
      style={{
        color: "inherit",
        background: "transparent",
        border: "var(--ct-border-hairline)",
        borderRadius: 3,
        padding: "3px 8px",
        font: "inherit",
        opacity: pending || unavailable ? 0.55 : 1,
      }}
    >
      {label}
    </button>
  );
  return (
    <fieldset
      aria-label="Workbook browsing"
      style={{
        margin: 0,
        minWidth: 0,
        border: 0,
        display: "flex",
        alignItems: "center",
        gap: 8,
        flexWrap: "wrap",
        padding: "4px 8px",
        fontSize: 12,
        borderBottom: "var(--ct-border-hairline)",
      }}
    >
      {button("Earlier rows", "earlier", !state.hasEarlier || unapplied)}
      {button("Load more", "more", !state.canLoadMore || unapplied)}
      {button("Refresh", "restart")}
      <span
        role="status"
        aria-live={pending ? "off" : "polite"}
        aria-atomic="true"
      >
        {pending
          ? state.pending === "more"
            ? "Loading more records…"
            : "Loading records…"
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
            <button type="button" onClick={() => registry.revert(viewSchemaId)}>
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
