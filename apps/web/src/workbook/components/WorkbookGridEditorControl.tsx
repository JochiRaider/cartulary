import type {
  GridCellTarget,
  GridEditCommitOutcome,
  GridEditorAdapter,
} from "@cartulary/grid-adapter";
import type { ViewFieldContract } from "@cartulary/view-contracts";
import type { KeyboardEvent } from "react";
import type { GenericReferenceOptions } from "../models/workbookReferenceOptions";
import { GenericMutationControl } from "./GenericMutationControl";

export function workbookGridEditorAdapter<Row>({
  commit,
  field,
  readValue,
  referenceOptions,
}: {
  readonly commit: (
    draftValue: string,
    target: GridCellTarget,
    row: Row,
  ) => Promise<GridEditCommitOutcome>;
  readonly field: ViewFieldContract;
  readonly readValue: (row: Row) => unknown;
  readonly referenceOptions: GenericReferenceOptions;
}): GridEditorAdapter<Row> {
  return {
    ...(field.clearable ? { clearDraftValue: "" } : {}),
    commit: (intent) =>
      commit(String(intent.draftValue ?? ""), intent.target, intent.row),
    initialDraftValue: (row) => {
      const value = readValue(row);
      return value === null || value === undefined ? "" : String(value);
    },
    renderEditor: (context) => {
      const draftValue = String(context.draftValue ?? "");
      const commitOnEnter = (event: KeyboardEvent<HTMLFieldSetElement>) => {
        if (event.key === "Escape") {
          event.preventDefault();
          context.cancel();
          return;
        }
        if (event.key === "Enter" && !event.shiftKey) {
          event.preventDefault();
          void context.commit();
        }
      };
      return (
        <fieldset
          aria-label={`Edit ${field.label}`}
          data-grid-editor-kind={workbookGridEditorKind(field)}
          onKeyDown={commitOnEnter}
        >
          <GenericMutationControl
            collectionMode="add"
            field={field}
            focusTargetRef={context.focusTargetRef}
            referenceOptions={referenceOptions}
            surface="grid"
            testId={`grid-editor-${
              context.target.rowIdentity.kind === "core_record"
                ? context.target.rowIdentity.recordId
                : "unsupported"
            }-${field.fieldKey}`}
            value={draftValue}
            onChange={(value) => context.setDraftValue(value)}
          />
          <button
            disabled={context.pending}
            type="button"
            onClick={() => void context.commit()}
          >
            Commit
          </button>
          <button
            disabled={context.pending}
            type="button"
            onClick={context.cancel}
          >
            Cancel
          </button>
        </fieldset>
      );
    },
  };
}

export type WorkbookGridEditorKind =
  | "boolean"
  | "enum"
  | "multiline"
  | "number"
  | "single-line"
  | "stable-reference"
  | "timestamp";

export function workbookGridEditorKind(
  field: ViewFieldContract,
): WorkbookGridEditorKind | null {
  if (!field.gridEditable || field.writeKind !== "direct_value") return null;
  if (field.enumValues && field.enumValues.length > 0) return "enum";
  if (field.readKind === "boolean") return "boolean";
  if (field.readKind === "number") return "number";
  if (field.directReferenceContractId) return "stable-reference";
  if (field.directScalarContractId === "timestamp_instant_v1") {
    return "timestamp";
  }
  if (field.stringContractId === "multiline_body_v1") return "multiline";
  return "single-line";
}
