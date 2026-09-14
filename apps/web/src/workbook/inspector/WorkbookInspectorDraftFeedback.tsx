import type { ViewContract } from "@cartulary/view-contracts";
import { genericCellLabel } from "../models/genericWorkbookModel";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { WorkbookInspectorActionButton as Button } from "./presentation/WorkbookInspectorActions";
import type { WorkbookInspectorEditDraft } from "./useWorkbookInspectorEditDraft";

export function WorkbookInspectorDraftFeedback({
  edit,
  contract,
  row,
}: {
  edit: WorkbookInspectorEditDraft;
  contract: ViewContract;
  row: WorkbookQueryRow | null;
}) {
  if (edit.needsResume)
    return (
      <div role="status">
        Unfinished work for this record is retained.{" "}
        <Button
          type="button"
          disabled={!edit.canResume}
          onClick={(event) => edit.resume(event.currentTarget)}
        >
          Resume draft
        </Button>{" "}
        <Button
          type="button"
          onClick={(event) => edit.discard(event.currentTarget)}
        >
          Discard draft
        </Button>
      </div>
    );
  return (
    <>
      {edit.staleFields.map((key) => (
        <div role="status" key={key}>
          Saved {contract.fieldMap[key]?.label ?? "field"} changed to{" "}
          {genericCellLabel(row?.cells[key]?.value)}. Your draft is retained.{" "}
          <Button
            type="button"
            onClick={(event) => edit.review(key, false, event.currentTarget)}
          >
            Use saved {contract.fieldMap[key]?.label ?? "field"}
          </Button>{" "}
          <Button
            type="button"
            onClick={(event) => edit.review(key, true, event.currentTarget)}
          >
            {contract.fieldMap[key]?.writeKind === "action_payload"
              ? "Review retained actions for"
              : "Keep draft"}{" "}
            {contract.fieldMap[key]?.label ?? "field"}
          </Button>
        </div>
      ))}
      {edit.draft ? (
        <Button
          type="button"
          onClick={(event) => edit.discard(event.currentTarget)}
        >
          Discard draft
        </Button>
      ) : null}
    </>
  );
}
