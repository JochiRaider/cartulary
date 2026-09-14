import { requireViewContract } from "@cartulary/view-contracts";
import { useSyncExternalStore } from "react";
import { genericInspectorRowLabel } from "../models/genericWorkbookModel";
import type { WorkbookExplicitPatchOwner } from "../runtime/WorkbookExplicitPatchOwner";
import { WorkbookInspectorActionButton as Button } from "./presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "./presentation/WorkbookInspectorFeedback";
import { workbookInspectorErrorPresentation } from "./workbookInspectorErrorModel";

/** Surface-local recovery remains reachable after the originating inspector closes. */
export function WorkbookExplicitPatchRecovery({
  owner,
  viewSchemaId,
}: {
  owner: WorkbookExplicitPatchOwner;
  viewSchemaId: string;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const entries = snapshot.entries.filter(
    (entry) =>
      entry.intent.viewSchemaId === viewSchemaId &&
      entry.intent.owner !== "party_link" &&
      entry.intent.purpose !== "task-lifecycle",
  );
  if (!snapshot.authority || !entries.length) return null;
  const contract = requireViewContract(viewSchemaId);
  return (
    <section
      aria-label="Inspector changes"
      style={{
        display: "grid",
        gap: "var(--ct-spacing-xs)",
        maxBlockSize: "calc(var(--ct-layout-statusStripHeight) * 5)",
        overflowY: "auto",
        padding: "var(--ct-spacing-xs)",
      }}
    >
      {entries.map((entry) => (
        <div key={entry.id}>
          <p role="status" style={{ margin: 0 }}>
            {genericInspectorRowLabel(contract, entry.intent.baseline)} —{" "}
            {entry.intent.changes
              .map(
                (change) =>
                  contract.fieldMap[change.field_key]?.label ?? "Field",
              )
              .join(", ")}
            :{" "}
            {entry.receipt
              ? `Saved, version ${entry.receipt.row.row_version}.`
              : entry.phase === "uncertain"
                ? "The outcome could not be confirmed. The original request is retained."
                : entry.phase === "coordinating"
                  ? "Waiting for earlier writes and current saved values."
                  : entry.phase === "submitting"
                    ? "Saving changes."
                    : entry.phase === "conflict"
                      ? "Review the saved-field conflict. Your draft is retained."
                      : "The change was not accepted. Your draft is retained."}
          </p>
          {entry.failure ? (
            <WorkbookInspectorPublicError
              error={workbookInspectorErrorPresentation(entry.failure)}
            />
          ) : null}
          {entry.receipt && entry.reconciliation !== "complete" ? (
            <p role="status">
              The change was saved. The view still needs a refresh.
            </p>
          ) : null}
          {entry.phase === "uncertain" ? (
            <Button
              type="button"
              disabled={!owner.canSubmit()}
              onClick={() => void owner.replay(entry.id)}
            >
              Retry original change
            </Button>
          ) : null}
          {entry.receipt && entry.reconciliation === "required" ? (
            <Button type="button" onClick={() => void owner.refresh(entry.id)}>
              Refresh saved change
            </Button>
          ) : null}
          {entry.phase === "preparation_failed" ||
          entry.phase === "rejected" ||
          (entry.receipt && entry.reconciliation === "complete") ? (
            <Button type="button" onClick={() => owner.dismiss(entry.id)}>
              Dismiss change notice
            </Button>
          ) : null}
        </div>
      ))}
    </section>
  );
}
