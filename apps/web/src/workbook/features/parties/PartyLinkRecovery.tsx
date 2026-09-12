import { useRef, useState, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "../../inspector/presentation/WorkbookInspectorFeedback";
import { workbookInspectorErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import type { ExplicitPatchOperation } from "../../runtime/WorkbookExplicitPatchOwner";
import type { WorkbookPartyLinkOperationOwner } from "./WorkbookPartyLinkOperationOwner";

export function PartyPatchFeedback({
  owner,
  entry,
}: {
  owner: WorkbookPartyLinkOperationOwner;
  entry: ExplicitPatchOperation;
}) {
  return (
    <section aria-label="Party source operation">
      <p role="status">
        {entry.receipt
          ? entry.reconciliation === "complete"
            ? "Source change saved."
            : "Source change saved; refresh is still required."
          : entry.phase === "uncertain"
            ? "Source change outcome is uncertain. The original request is retained."
            : entry.phase === "conflict"
              ? "The source field changed. Resolve its conflict, then review the complete Party action again."
              : entry.phase === "preparation_failed"
                ? "The source change was not sent. Review the current source and try again."
                : entry.phase === "rejected"
                  ? "The source change was rejected. Review current source values before another attempt."
                  : "Saving source change…"}
      </p>
      {entry.failure ? (
        <WorkbookInspectorPublicError
          error={workbookInspectorErrorPresentation(entry.failure)}
        />
      ) : null}
      {entry.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          disabled={!owner.canSubmit()}
          onClick={() => void owner.patches.replay(entry.id)}
        >
          Replay original source change
        </WorkbookInspectorActionButton>
      ) : null}
      {entry.receipt && entry.reconciliation === "required" ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          onClick={() => void owner.patches.refresh(entry.id)}
        >
          Refresh source result
        </WorkbookInspectorActionButton>
      ) : null}
    </section>
  );
}
export function PartyLinkRecovery({
  owner,
}: {
  owner: WorkbookPartyLinkOperationOwner;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const disclosure = useRef<HTMLDetailsElement>(null),
    trigger = useRef<HTMLElement>(null);
  const [open, setOpen] = useState(false);
  const total = snapshot.creations.length + snapshot.patches.length;
  if (!snapshot.authority || !total) return null;
  return (
    <details
      ref={disclosure}
      onToggle={(event) => setOpen(event.currentTarget.open)}
      style={{ position: "relative", minWidth: 0 }}
    >
      <summary ref={trigger}>Party operations ({total})</summary>
      {open ? (
        <section
          aria-label="Retained Party operations"
          tabIndex={-1}
          onKeyDown={(event) => {
            if (event.key === "Escape" && disclosure.current) {
              event.preventDefault();
              event.stopPropagation();
              disclosure.current.open = false;
              trigger.current?.focus({ preventScroll: true });
            }
          }}
          style={{
            position: "fixed",
            zIndex: 20,
            insetInlineEnd: "var(--ct-spacing-md)",
            boxSizing: "border-box",
            background: "var(--ct-colors-surface-1)",
            border: "var(--ct-border-hairline)",
            padding: "var(--ct-spacing-md)",
            width: "min(28rem, calc(100vw - 2rem))",
            maxHeight: "65vh",
            overflow: "auto",
            overflowWrap: "anywhere",
          }}
        >
          {snapshot.creations.map((entry) => (
            <section key={entry.id}>
              <p>
                {entry.attempt.review.sourceLabel} —{" "}
                {entry.attempt.review.pair.label}
              </p>
              <p role="status">
                {entry.receipt
                  ? `Party saved: ${String(entry.receipt.data.row.cells["party.display_name"]?.value ?? "Party")}. ${entry.linkId ? "Source linking has its own result." : "Return to this source and pair to review linking the saved Party."}`
                  : entry.phase === "uncertain"
                    ? "Party creation outcome is uncertain."
                    : (entry.failure?.message ?? "Saving Party…")}
              </p>
              {entry.failure ? (
                <WorkbookInspectorPublicError
                  error={workbookInspectorErrorPresentation(entry.failure)}
                />
              ) : null}
              {entry.phase === "uncertain" ? (
                <WorkbookInspectorActionButton
                  tone="secondary"
                  disabled={!owner.canSubmit()}
                  onClick={() => void owner.replayCreation(entry.id)}
                >
                  Replay Party creation
                </WorkbookInspectorActionButton>
              ) : null}
              {entry.receipt && entry.refresh === "required" ? (
                <>
                  <p>Party saved; its view still needs refreshing.</p>
                  <WorkbookInspectorActionButton
                    tone="secondary"
                    onClick={() => void owner.refreshCreation(entry.id)}
                  >
                    Refresh saved Party
                  </WorkbookInspectorActionButton>
                </>
              ) : null}
            </section>
          ))}
          {snapshot.patches.map((entry) => (
            <section key={entry.id}>
              <p>
                {entry.intent.partyReview?.sourceLabel} —{" "}
                {entry.intent.partyReview?.pair.label}
              </p>
              <PartyPatchFeedback owner={owner} entry={entry} />
            </section>
          ))}
        </section>
      ) : null}
    </details>
  );
}
