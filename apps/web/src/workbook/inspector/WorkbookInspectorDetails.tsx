import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  type CSSProperties,
  type ReactNode,
  useLayoutEffect,
  useRef,
} from "react";
import { workbookTypography } from "../components/workbookFormStyles";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { WorkbookInspectorActionButton as Button } from "./presentation/WorkbookInspectorActions";
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
    readonly canResume: boolean;
    readonly discard: () => void;
  }[];
  readonly onReviewDraft: (identity: {
    readonly fieldKey: string;
    readonly action: string;
  }) => void;
  readonly collectionDestinations?:
    | Readonly<Record<string, (() => void) | undefined>>
    | undefined;
  readonly disabledReason?: string | null | undefined;
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
  onReviewDraft,
  collectionDestinations,
}: WorkbookInspectorEditPresentation) {
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
              {editable ? (
                <Button
                  tone="secondary"
                  aria-label={`${field.readKind === "collection" ? "Manage" : "Edit"} ${field.label}`}
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
                  {field.readKind === "collection" ? "Manage" : "Edit"}
                </Button>
              ) : collectionDestinations?.[field.fieldKey] ? (
                <Button onClick={collectionDestinations[field.fieldKey]}>
                  Manage {field.label}
                </Button>
              ) : null}
            </>
          ),
          attachment: (
            <>
              {field.fieldKey !== activeField && retained.length ? (
                <dd style={fullRowStyle}>
                  <span style={workbookTypography("metadata")}>
                    Unfinished work retained for {field.label}.{" "}
                  </span>
                  {retained.map((work) => (
                    <span key={work.identity.action}>
                      {editable ? (
                        <Button
                          disabled={
                            !cell || !!disabledReason || !work.canResume
                          }
                          onClick={() => onReviewDraft(work.identity)}
                        >
                          Review{" "}
                          {retained.length > 1
                            ? `${work.identity.action} `
                            : ""}
                          draft for {field.label}
                        </Button>
                      ) : null}
                      <Button onClick={work.discard}>
                        Discard{" "}
                        {retained.length > 1 ? `${work.identity.action} ` : ""}
                        draft for {field.label}
                      </Button>
                    </span>
                  ))}
                </dd>
              ) : null}
              {field.fieldKey === activeField ? (
                <dd style={fullRowStyle}>
                  <fieldset
                    ref={attachment}
                    style={editorStyle}
                    data-inspector-editor-field={field.fieldKey}
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
                    <legend style={labelStyle}>
                      Unsaved change: {field.label}
                    </legend>
                    {editor.content}
                    <div style={actionsStyle}>
                      {editor.actions}
                      <Button tone="secondary" onClick={detach}>
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
    <WorkbookInspectorSavedDetails
      contract={contract}
      row={row}
      fields={fields}
      feedback={
        disabledReason ? (
          <div>
            <dt>Editing unavailable</dt>
            <dd style={{ margin: 0 }}>{disabledReason}</dd>
          </div>
        ) : null
      }
    />
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
