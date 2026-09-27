import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  type CSSProperties,
  type ReactNode,
  useId,
  useLayoutEffect,
  useRef,
} from "react";
import { workbookTypography } from "../components/workbookFormStyles";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { WorkbookInspectorActionButton as Button } from "./presentation/WorkbookInspectorActions";
import {
  type WorkbookInspectorDisabledReason,
  workbookInspectorDisabledReasonKey,
  workbookInspectorDisabledReasonText,
} from "./presentation/workbookInspectorPresentationModel";
import { WorkbookInspectorSavedDetails } from "./WorkbookInspectorSavedDetails";

type WorkbookInspectorEditorSlots = {
  readonly content: ReactNode;
  readonly actions: ReactNode;
  readonly feedback: ReactNode;
  readonly retainedDraft: ReactNode;
};

/** Explicit edit controls, commands and feedback, independent of any draft owner. */
type WorkbookInspectorEditPresentation = {
  readonly contract: ViewContract;
  readonly row: WorkbookQueryRow;
  readonly editableFields: readonly ViewFieldContract[];
  readonly activeField: string;
  readonly onEdit: (fieldKey: string) => void;
  readonly onDetach: () => void;
  readonly onSubmit: () => void;
  readonly canSubmit: boolean;
  readonly editor: WorkbookInspectorEditorSlots;
  readonly retainedWork: readonly {
    readonly identity: { readonly fieldKey: string; readonly action: string };
    readonly command: {
      readonly kind: "resume" | "review";
      readonly invoke: (trigger?: HTMLElement) => void;
    } | null;
    readonly discard: () => void;
  }[];
  readonly collectionDestinations?:
    | Readonly<Record<string, (() => void) | undefined>>
    | undefined;
  readonly disabledReason?: WorkbookInspectorDisabledReason | null | undefined;
};

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
  retainedWork,
  collectionDestinations,
}: WorkbookInspectorEditPresentation) {
  const reasonId = useId();
  const reasonText = disabledReason
    ? workbookInspectorDisabledReasonText(disabledReason)
    : null;
  const editButtons = useRef(new Map<string, HTMLButtonElement>());
  const attachment = useRef<HTMLFieldSetElement>(null);
  const previousField = useRef("");
  const closingField = useRef<{ recordId: string; fieldKey: string } | null>(
    null,
  );
  useLayoutEffect(() => {
    const closing = closingField.current;
    closingField.current = null;
    if (closing && !activeField && closing.recordId === row.record_id)
      editButtons.current.get(closing.fieldKey)?.focus({ preventScroll: true });
    if (activeField && activeField !== previousField.current) {
      attachment.current
        ?.querySelector<HTMLElement>(
          "input:not(:disabled), textarea:not(:disabled), select:not(:disabled), [data-inspector-review-control]:not(:disabled)",
        )
        ?.focus({ preventScroll: true });
    }
    previousField.current = activeField;
  }, [activeField, row.record_id]);
  const detach = () => {
    closingField.current = { recordId: row.record_id, fieldKey: activeField };
    onDetach();
  };
  const fields = new Map(
    contract.fields.map((field) => {
      const cell = row.cells[field.fieldKey];
      const editable = editableFields.some(
        (candidate) => candidate.fieldKey === field.fieldKey,
      );
      const retained = retainedWork.filter(
        (work) => work.identity.fieldKey === field.fieldKey,
      );
      return [
        field.fieldKey,
        {
          controls: (
            <>
              {editable && field.fieldKey !== activeField && retained.length ? (
                retained.map((work) => (
                  <Button
                    key={work.identity.action}
                    aria-label={`${work.command?.kind === "review" ? "Review" : "Resume"} ${retained.length > 1 ? `${work.identity.action} ` : ""}draft for ${field.label}`}
                    tone="quiet"
                    data-inspector-edit-field={field.fieldKey}
                    ref={(element) => {
                      if (element)
                        editButtons.current.set(field.fieldKey, element);
                      else editButtons.current.delete(field.fieldKey);
                    }}
                    aria-describedby={disabledReason ? reasonId : undefined}
                    disabled={!cell || !!disabledReason || !work.command}
                    onClick={(event) =>
                      work.command?.invoke(event.currentTarget)
                    }
                  >
                    {work.command?.kind === "review" ? "Review" : "Resume"}{" "}
                    {retained.length > 1 ? `${work.identity.action} ` : ""}draft
                  </Button>
                ))
              ) : editable && field.fieldKey !== activeField ? (
                <Button
                  tone="quiet"
                  aria-label={`${field.readKind === "collection" ? "Manage" : "Edit"} ${field.label}`}
                  data-inspector-edit-field={field.fieldKey}
                  ref={(element) => {
                    if (element)
                      editButtons.current.set(field.fieldKey, element);
                    else editButtons.current.delete(field.fieldKey);
                  }}
                  disabled={!cell || !!disabledReason}
                  aria-describedby={disabledReason ? reasonId : undefined}
                  title={reasonText ?? undefined}
                  onClick={() => onEdit(field.fieldKey)}
                >
                  {field.readKind === "collection" ? "Manage" : "Edit"}
                </Button>
              ) : collectionDestinations?.[field.fieldKey] ? (
                <Button
                  tone="quiet"
                  aria-label={`Manage ${field.label}`}
                  onClick={collectionDestinations[field.fieldKey]}
                >
                  Manage
                </Button>
              ) : null}
            </>
          ),
          attachment: (
            <>
              {field.fieldKey !== activeField && retained.length ? (
                <dd style={fullRowStyle}>
                  {retained.map((work) => (
                    <span key={work.identity.action}>
                      <Button
                        tone="quiet"
                        aria-label={`Discard ${retained.length > 1 ? `${work.identity.action} ` : ""}draft for ${field.label}`}
                        onClick={work.discard}
                      >
                        Discard{" "}
                        {retained.length > 1 ? `${work.identity.action} ` : ""}
                        draft
                      </Button>
                    </span>
                  ))}
                </dd>
              ) : null}
              {field.fieldKey === activeField ? (
                <dd style={fullRowStyle}>
                  <fieldset
                    aria-label={`Unsaved change: ${field.label}`}
                    ref={attachment}
                    style={editorStyle}
                    data-inspector-editor-field={field.fieldKey}
                    aria-describedby={disabledReason ? reasonId : undefined}
                    onKeyDown={(event) => {
                      if (
                        event.defaultPrevented ||
                        event.nativeEvent.isComposing
                      )
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
                    <legend style={labelStyle}>Unsaved change</legend>
                    {editor.content}
                    <div style={actionsStyle}>
                      {editor.actions}
                      <Button tone="quiet" onClick={detach}>
                        Close editor
                      </Button>
                    </div>
                    <p style={retentionStyle}>
                      Update saves this field. Closing keeps unfinished work in
                      this session.
                    </p>
                    {editor.feedback}
                    {editor.retainedDraft}
                  </fieldset>
                </dd>
              ) : null}
            </>
          ),
        },
      ] as const;
    }),
  );
  return (
    <>
      {disabledReason ? (
        <p
          id={reasonId}
          data-inspector-read-only-reason={workbookInspectorDisabledReasonKey(
            disabledReason,
          )}
          style={retentionStyle}
        >
          {reasonText}
        </p>
      ) : null}
      <WorkbookInspectorSavedDetails
        contract={contract}
        row={row}
        fields={fields}
        describedBy={disabledReason ? reasonId : undefined}
      />
    </>
  );
}

const labelStyle = {
  ...workbookTypography("metadata"),
  color: "var(--ct-colors-ink-muted)",
  minInlineSize: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const fullRowStyle = {
  margin: 0,
  gridColumn: "1 / -1",
  minInlineSize: 0,
} satisfies CSSProperties;
const editorStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  margin: 0,
  padding: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  minInlineSize: 0,
} satisfies CSSProperties;
const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-xs)",
  alignItems: "center",
} satisfies CSSProperties;
const retentionStyle = {
  ...workbookTypography("metadata"),
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
} satisfies CSSProperties;
