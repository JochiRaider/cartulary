import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import { type ReactNode, useLayoutEffect, useRef } from "react";
import { genericCellLabelForField } from "../models/genericWorkbookModel";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { WorkbookInspectorActionButton as Button } from "./presentation/WorkbookInspectorActions";

/** Saved values always come from the accepted row, never from editor text. */
export function WorkbookInspectorDetails({
  contract,
  row,
  editableFields,
  activeField,
  onEdit,
  onDetach,
  onSubmit,
  canSubmit,
  editor,
  disabledReason,
}: {
  readonly contract: ViewContract;
  readonly row: WorkbookQueryRow;
  readonly editableFields: readonly ViewFieldContract[];
  readonly activeField: string;
  readonly onEdit: (fieldKey: string) => void;
  readonly onDetach: () => void;
  readonly onSubmit: () => void;
  readonly canSubmit: boolean;
  readonly editor: ReactNode;
  readonly disabledReason?: string | null | undefined;
}) {
  const editButtons = useRef(new Map<string, HTMLButtonElement>());
  const attachment = useRef<HTMLFieldSetElement>(null);
  const previousField = useRef("");
  useLayoutEffect(() => {
    if (activeField && activeField !== previousField.current) {
      attachment.current
        ?.querySelector<HTMLElement>(
          "input:not(:disabled), textarea:not(:disabled), select:not(:disabled), button:not(:disabled)",
        )
        ?.focus({ preventScroll: true });
    }
    previousField.current = activeField;
  }, [activeField]);
  const detach = () => {
    const trigger = editButtons.current.get(activeField);
    onDetach();
    trigger?.focus({ preventScroll: true });
  };
  return (
    <dl
      style={{
        display: "grid",
        gap: "var(--ct-spacing-md)",
        margin: 0,
        minInlineSize: 0,
      }}
    >
      {contract.fields.map((field) => {
        const cell = row.cells[field.fieldKey];
        const value = cell?.value;
        const saved = !cell
          ? "Not loaded"
          : value === null
            ? "Not set"
            : value === undefined
              ? "Not loaded"
              : value === ""
                ? "Empty text"
                : genericCellLabelForField(
                    contract.viewSchemaId,
                    field.fieldKey,
                    value,
                  );
        const editable = editableFields.some(
          (candidate) => candidate.fieldKey === field.fieldKey,
        );
        return (
          <div
            key={field.fieldKey}
            data-inspector-saved-field={field.fieldKey}
            style={{ minInlineSize: 0 }}
          >
            <dt
              style={{
                display: "flex",
                flexWrap: "wrap",
                alignItems: "center",
                justifyContent: "space-between",
                gap: "var(--ct-spacing-xs)",
              }}
            >
              <span>{field.label}</span>
              {editable ? (
                <Button
                  tone="secondary"
                  aria-label={`Edit ${field.label}`}
                  data-inspector-edit-field={field.fieldKey}
                  ref={(element) => {
                    if (element)
                      editButtons.current.set(field.fieldKey, element);
                    else editButtons.current.delete(field.fieldKey);
                  }}
                  disabled={!cell || !!disabledReason}
                  title={disabledReason ?? undefined}
                  onClick={() => onEdit(field.fieldKey)}
                >
                  Edit
                </Button>
              ) : null}
            </dt>
            <dd
              style={{
                margin: 0,
                whiteSpace: "pre-wrap",
                overflowWrap: "anywhere",
                minInlineSize: 0,
              }}
            >
              {saved}
            </dd>
            {field.fieldKey === activeField ? (
              <dd style={{ margin: 0 }}>
                <fieldset
                  ref={attachment}
                  style={{ margin: 0, padding: 0, border: 0, minInlineSize: 0 }}
                  data-inspector-editor-field={field.fieldKey}
                  onKeyDown={(event) => {
                    if (event.defaultPrevented || event.nativeEvent.isComposing)
                      return;
                    if (event.key === "Escape") {
                      event.preventDefault();
                      event.stopPropagation();
                      detach();
                    } else if (
                      event.key === "Enter" &&
                      (event.ctrlKey || event.metaKey)
                    ) {
                      event.preventDefault();
                      event.stopPropagation();
                      if (!event.repeat && canSubmit) onSubmit();
                    }
                  }}
                >
                  <legend style={{ color: "var(--ct-colors-ink-muted)" }}>
                    Unsaved change — {field.label}
                  </legend>
                  {editor}
                  <Button tone="secondary" onClick={detach}>
                    Done editing
                  </Button>
                </fieldset>
              </dd>
            ) : null}
          </div>
        );
      })}
      {disabledReason ? (
        <div>
          <dt>Editing unavailable</dt>
          <dd style={{ margin: 0 }}>{disabledReason}</dd>
        </div>
      ) : null}
    </dl>
  );
}
