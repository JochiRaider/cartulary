import { timelineCaptureActionTestId } from "@cartulary/ui-contracts";
import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type { TimelineCaptureOperation } from "../ports/TimelineRecordActionPort";
import type { WorkbookTimelineCaptureActionOwner } from "./WorkbookTimelineCaptureActionOwner";

export function TimelineCaptureRecovery({
  owner,
}: {
  readonly owner: WorkbookTimelineCaptureActionOwner;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const panel = useRef<HTMLElement>(null);
  const requestedFocus = useRef(false);
  const focusedControl = useRef<HTMLElement | null>(null);
  useLayoutEffect(() => {
    if (!snapshot.authority) setOpen(false);
    if (
      open &&
      (requestedFocus.current ||
        (focusedControl.current &&
          !focusedControl.current.isConnected &&
          document.activeElement === document.body))
    ) {
      requestedFocus.current = false;
      panel.current?.focus({ preventScroll: true });
    }
  }, [open, snapshot]);
  if (!snapshot.authority || !snapshot.entries.length) return null;
  const close = () => {
    const restore = panel.current?.contains(document.activeElement);
    setOpen(false);
    if (restore) trigger.current?.focus({ preventScroll: true });
  };
  return (
    <div
      style={{
        position: "relative",
        display: "inline-flex",
        alignItems: "center",
        minWidth: 0,
      }}
    >
      <WorkbookInspectorActionButton
        ref={trigger}
        style={{ whiteSpace: "nowrap", flexShrink: 0 }}
        aria-expanded={open}
        onClick={() => {
          if (open) close();
          else {
            requestedFocus.current = true;
            setOpen(true);
          }
        }}
      >
        Timeline actions ({snapshot.entries.length})
      </WorkbookInspectorActionButton>
      <span
        role="status"
        style={{
          marginInlineStart: "var(--ct-spacing-xs)",
          minWidth: 0,
          overflow: "hidden",
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
          fontSize: "var(--ct-typography-compact-metadata-fontSize)",
        }}
      >
        {snapshot.entries.length === 1
          ? status(snapshot.entries[0] as TimelineCaptureOperation)
          : `${snapshot.entries.filter((entry) => entry.receipt).length} Timeline actions completed. ${snapshot.entries.filter((entry) => entry.phase === "uncertain").length} outcomes unknown.`}
      </span>
      {open ? (
        <section
          ref={panel}
          tabIndex={-1}
          aria-label="Timeline action recovery"
          onFocusCapture={(event) => {
            focusedControl.current = event.target as HTMLElement;
          }}
          onBlurCapture={() => {
            focusedControl.current = null;
          }}
          data-testid={timelineCaptureActionTestId("recovery")}
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.stopPropagation();
              close();
            }
          }}
          style={{
            position: "fixed",
            zIndex: 30,
            insetBlockStart:
              "calc(var(--ct-layout-topBarHeight) + var(--ct-spacing-xs))",
            insetInlineEnd: "var(--ct-spacing-sm)",
            inlineSize: "min(38rem, 90vw)",
            maxBlockSize: "75vh",
            overflow: "auto",
            overflowWrap: "anywhere",
            padding: "var(--ct-spacing-md)",
            background: "var(--ct-colors-surface-1)",
            border: "var(--ct-border-hairline)",
            boxShadow: "var(--ct-elevation-popover)",
          }}
        >
          <strong>Timeline actions</strong>
          <p>
            Admitted actions remain here after the inspector closes. A lost
            response does not mean the server rejected the action.
          </p>
          <WorkbookInspectorActionButton onClick={close}>
            Close Timeline actions
          </WorkbookInspectorActionButton>
          {snapshot.entries.map((entry) => (
            <article
              key={entry.key}
              aria-label={`Timeline action for ${entry.review.target.label}`}
              style={{ marginBlock: "var(--ct-spacing-md)" }}
            >
              <strong>
                {entry.review.action === "mark-reviewed"
                  ? "Mark reviewed"
                  : "Supersede"}
                : {entry.review.target.label}
              </strong>
              <p>
                {entry.review.target.recordId} · reviewed version{" "}
                {entry.review.target.rowVersion}
              </p>
              {entry.review.reason ? (
                <p style={{ whiteSpace: "pre-wrap" }}>{entry.review.reason}</p>
              ) : null}
              {entry.review.action === "supersede" ? (
                <p>
                  Replacement:{" "}
                  {entry.review.replacement
                    ? `${entry.review.replacement.label} · ${entry.review.replacement.context} · ${entry.review.replacement.recordId}`
                    : "No replacement"}
                </p>
              ) : null}
              <p
                data-testid={timelineCaptureActionTestId(
                  "result",
                  entry.review.target.recordId,
                )}
              >
                {status(entry)}
              </p>
              {entry.failure ? (
                <p role="alert">{entry.failure.message}</p>
              ) : null}
              {entry.receipt ? (
                <p>
                  Saved version {entry.receipt.data.row_version} · change{" "}
                  {entry.receipt.data.change_set_id}. Your current selection and
                  filters are preserved; the completed row may leave the current
                  results.
                </p>
              ) : null}
              {entry.phase === "uncertain" ? (
                <>
                  <p>
                    The outcome remains unknown. Retry sends the exact original
                    request and transaction identity.
                  </p>
                  <WorkbookInspectorActionButton
                    disabled={!owner.canSubmit()}
                    title={owner.unavailableReason() ?? undefined}
                    data-testid={timelineCaptureActionTestId(
                      "retry",
                      entry.review.target.recordId,
                    )}
                    onClick={() => {
                      void owner.replay(entry.key);
                    }}
                  >
                    Retry exact action
                  </WorkbookInspectorActionButton>
                </>
              ) : null}
              {entry.receipt && entry.reconciliation !== "complete" ? (
                <WorkbookInspectorActionButton
                  disabled={entry.reconciliation === "refreshing"}
                  data-testid={timelineCaptureActionTestId(
                    "refresh",
                    entry.review.target.recordId,
                  )}
                  onClick={() => {
                    void owner.refresh(entry.key);
                  }}
                >
                  Refresh result
                </WorkbookInspectorActionButton>
              ) : null}
              {["preparation_failed", "rejected"].includes(entry.phase) ||
              entry.reconciliation === "complete" ? (
                <WorkbookInspectorActionButton
                  disabled={entry.transportPending}
                  onClick={() => owner.dismiss(entry.key)}
                >
                  Dismiss action
                </WorkbookInspectorActionButton>
              ) : null}
            </article>
          ))}
        </section>
      ) : null}
    </div>
  );
}
function status(entry: TimelineCaptureOperation) {
  if (entry.receipt)
    return `Timeline ${entry.review.action === "mark-reviewed" ? "review" : "supersession"} completed.${entry.reconciliation === "complete" ? "" : " Refresh still required."}`;
  if (entry.phase === "uncertain") return "Timeline action outcome unknown.";
  if (entry.phase === "preparation_failed")
    return "Timeline action needs another review. No action was sent.";
  if (entry.phase === "rejected") return "Timeline action rejected.";
  return "Timeline action in progress.";
}
