import { requireViewContract } from "@cartulary/view-contracts";
import { useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookOrdinaryCreateOwner } from "./WorkbookOrdinaryCreateOwner";

/** Local acknowledgement and recovery; technical identities remain internal. */
export function OrdinaryCreateNotice({
  owner,
  view,
}: {
  owner: WorkbookOrdinaryCreateOwner;
  view: string;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const schema = snapshot.schemas[view];
  if (!schema) return null;
  const latest = schema.entries.at(-1);
  const entries = schema.entries.filter(
    (entry) =>
      entry === latest ||
      entry.phase === "uncertain" ||
      entry.phase === "submitting" ||
      entry.phase === "preparing" ||
      (entry.receipt && entry.refresh !== "complete"),
  );
  const errors = schema.showErrors ? Object.values(schema.errors) : [];
  const retained =
    !owner.canAuthor() && Object.keys(schema.draft.values).length > 0;
  if (!entries.length && !errors.length && !schema.message && !retained)
    return null;
  return (
    <section
      aria-label="Row creation"
      style={{
        padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
        fontSize: "var(--ct-typography-ui-fontSize)",
        maxBlockSize:
          "calc((100dvh - var(--ct-layout-topBarHeight) - var(--ct-layout-viewBarHeight) - var(--ct-layout-statusStripHeight)) / 2)",
        overflow: "auto",
      }}
    >
      {errors.length ? (
        <p role="alert" style={{ margin: 0 }}>
          {errors.join(" ")}
        </p>
      ) : null}
      {schema.message ? (
        <p role="status" style={{ margin: 0 }}>
          {schema.message}
        </p>
      ) : null}
      {entries.map((entry) => (
        <div key={entry.attempt.clientTxnId}>
          <span role={entry.phase === "rejected" ? "alert" : "status"}>
            {entry.message}
          </span>{" "}
          {entry.phase === "uncertain" ? (
            <Button
              type="button"
              tone="secondary"
              disabled={!owner.canReplay() || entry.transportPending}
              onClick={() => void owner.replay(entry.attempt.clientTxnId)}
            >
              Recover submission
            </Button>
          ) : null}
          {entry.failure?.kind === "client_txn_conflict" && !entry.receipt ? (
            <Button
              type="button"
              tone="secondary"
              disabled={!owner.canAuthor() || entry.transportPending}
              onClick={() =>
                void owner.submitFreshAfterConflict(entry.attempt.clientTxnId)
              }
            >
              Submit as a new request
            </Button>
          ) : null}
          {entry.receipt && entry.refresh !== "complete" ? (
            <Button
              type="button"
              tone="secondary"
              disabled={entry.refresh === "refreshing"}
              onClick={() =>
                void owner.refreshAccepted(entry.attempt.clientTxnId)
              }
            >
              Refresh accepted result
            </Button>
          ) : null}
        </div>
      ))}
      {retained ? (
        <details open>
          <summary>Unfinished row (read only)</summary>
          <p style={{ margin: 0 }}>
            Your authoring is retained and can be copied. Creating rows is
            currently unavailable.
          </p>
          {requireViewContract(view)
            .fields.filter((field) =>
              Object.hasOwn(schema.draft.values, field.fieldKey),
            )
            .map((field) => (
              <label
                key={field.fieldKey}
                style={{
                  display: "grid",
                  gap: "var(--ct-spacing-xs)",
                  marginBlock: "var(--ct-spacing-xs)",
                }}
              >
                {field.label}
                <textarea
                  aria-label={`${field.label} retained authoring`}
                  readOnly
                  rows={2}
                  value={
                    schema.draft.references[field.fieldKey]
                      ?.map((item) => item.displayText)
                      .join("\n") ??
                    schema.draft.values[field.fieldKey] ??
                    ""
                  }
                  style={{
                    boxSizing: "border-box",
                    width: "100%",
                    minWidth: 0,
                    font: "inherit",
                    color: "var(--ct-component-text-input-textColor)",
                    background:
                      "var(--ct-component-text-input-backgroundColor)",
                    border: "var(--ct-component-text-input-border)",
                  }}
                />
              </label>
            ))}
        </details>
      ) : null}
    </section>
  );
}
