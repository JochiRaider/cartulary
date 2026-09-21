import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
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
  useState,
} from "react";
import { workbookTypography } from "../components/workbookFormStyles";
import { genericCellLabelForField } from "../models/genericWorkbookModel";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { WorkbookInspectorActionButton as Button } from "./presentation/WorkbookInspectorActions";
import type { WorkbookInspectorEditDraft } from "./useWorkbookInspectorEditDraft";

export type WorkbookInspectorEditorSlots = {
  readonly content: ReactNode;
  readonly actions: ReactNode;
  readonly feedback: ReactNode;
  readonly retainedDraft: ReactNode;
};

/** Append-only source families reuse the saved-value renderer without an editor. */
export function WorkbookInspectorReadOnlyDetails({
  contract,
  row,
}: {
  readonly contract: ViewContract;
  readonly row: WorkbookQueryRow;
}) {
  return (
    <WorkbookInspectorDetails
      contract={contract}
      row={row}
      editableFields={[]}
      activeField=""
      onEdit={noEdit}
      onDetach={noEdit}
      onSubmit={noEdit}
      canSubmit={false}
      retainedWork={[]}
      onReviewDraft={noEdit}
      editor={{
        content: null,
        actions: null,
        feedback: null,
        retainedDraft: null,
      }}
    />
  );
}
const noEdit = () => {};

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
}: {
  readonly contract: ViewContract;
  readonly row: WorkbookQueryRow;
  readonly editableFields: readonly ViewFieldContract[];
  readonly activeField: string;
  readonly onEdit: (fieldKey: string) => void;
  readonly onDetach: () => void;
  readonly onSubmit: () => void;
  readonly canSubmit: boolean;
  readonly editor: WorkbookInspectorEditorSlots;
  readonly retainedWork: WorkbookInspectorEditDraft["retainedWork"];
  readonly onReviewDraft: (
    identity: WorkbookInspectorEditDraft["identity"],
  ) => void;
  readonly collectionDestinations?:
    | Readonly<Record<string, (() => void) | undefined>>
    | undefined;
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
    <dl style={detailsStyle}>
      {contract.fields.map((field) => {
        const cell = row.cells[field.fieldKey];
        const value = cell?.value;
        const saved =
          !cell || value === undefined
            ? "Not loaded"
            : value === null
              ? "Not set"
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
        const kind = inspectorSavedValueKind(field);
        const retained = retainedWork.filter(
          (work) => work.identity.fieldKey === field.fieldKey,
        );
        return (
          <div
            key={field.fieldKey}
            data-inspector-saved-field={field.fieldKey}
            data-inspector-value-kind={kind}
            style={{
              ...fieldStyle,
              gridTemplateColumns:
                kind === "scalar"
                  ? "minmax(0, 2fr) minmax(0, 3fr) auto"
                  : "minmax(0, 1fr) auto",
            }}
          >
            <dt style={labelStyle}>{field.label}</dt>
            <dd
              style={{
                ...valueStyle,
                ...(kind !== "scalar"
                  ? { gridColumn: "1 / -1", gridRow: 2 }
                  : {}),
              }}
            >
              <SavedValue
                key={JSON.stringify([row.record_id, field.fieldKey])}
                value={saved}
                label={field.label}
              />
            </dd>
            <dd
              style={{
                margin: 0,
                gridColumn: kind === "scalar" ? 3 : 2,
                gridRow: 1,
              }}
            >
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
            </dd>
            {field.fieldKey !== activeField && retained.length ? (
              <dd style={fullRowStyle}>
                <span style={workbookTypography("metadata")}>
                  Unfinished work retained for {field.label}.{" "}
                </span>
                {retained.map((work) => (
                  <span key={work.identity.action}>
                    {editable ? (
                      <Button
                        disabled={!cell || !!disabledReason || !work.canResume}
                        onClick={() => onReviewDraft(work.identity)}
                      >
                        Review{" "}
                        {retained.length > 1 ? `${work.identity.action} ` : ""}
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

/** Consume semantic source metadata; never infer field meaning from its label. */
export function inspectorSavedValueKind(
  field: ViewFieldContract,
): "scalar" | "narrative" | "stacked" {
  if (
    field.stringContractId === "multiline_body_v1" ||
    field.stringContractId === "reason_note_v1" ||
    field.stringContractId === "timeline_visible_text_v1"
  )
    return "narrative";
  if (
    ["text", "timestamp", "date", "number", "boolean", "enum"].includes(
      field.readKind,
    )
  )
    return "scalar";
  return "stacked";
}

function SavedValue({
  value,
  label,
}: {
  readonly value: string;
  readonly label: string;
}) {
  const id = useId();
  const text = useRef<HTMLDivElement>(null);
  const [expandedValue, setExpandedValue] = useState<string | null>(null);
  const [overflow, setOverflow] = useState(false);
  const expanded = expandedValue === value;
  const lines = cartularyDesignPresentation.inspector.narrativePreviewLines;
  useLayoutEffect(() => {
    const element = text.current;
    if (!element || element.textContent !== value) return;
    const measure = () => {
      const height = Number.parseFloat(getComputedStyle(element).lineHeight);
      setOverflow(
        Number.isFinite(height) && element.scrollHeight > height * lines + 1,
      );
    };
    measure();
    if (typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver(measure);
    observer.observe(element);
    return () => observer.disconnect();
  }, [lines, value]);
  return (
    <>
      <div
        id={id}
        ref={text}
        onCopy={(event) => {
          const selection = event.currentTarget.ownerDocument.getSelection();
          if (!selection || selection.isCollapsed || selection.rangeCount !== 1)
            return;
          const range = selection.getRangeAt(0);
          if (
            !event.currentTarget.contains(range.startContainer) ||
            !event.currentTarget.contains(range.endContainer)
          )
            return;
          // Native rendered-text serialization can drop trailing newlines.
          // A range wholly within this saved value copies its exact source text.
          event.clipboardData.setData("text/plain", range.toString());
          event.preventDefault();
          event.stopPropagation();
        }}
        style={{
          whiteSpace: "pre-wrap",
          overflowWrap: "anywhere",
          display: "-webkit-box",
          WebkitBoxOrient: "vertical",
          WebkitLineClamp: expanded ? "unset" : lines,
          overflow: expanded ? "visible" : "hidden",
        }}
      >
        {value}
      </div>
      {overflow || expanded ? (
        <Button
          aria-expanded={expanded}
          aria-controls={id}
          onClick={() => setExpandedValue(expanded ? null : value)}
        >
          {expanded ? "Show less" : "Show full value"} for {label}
        </Button>
      ) : null}
    </>
  );
}

const detailsStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  margin: 0,
  minInlineSize: 0,
} satisfies CSSProperties;
const fieldStyle = {
  display: "grid",
  alignItems: "baseline",
  columnGap: "var(--ct-spacing-sm)",
  rowGap: "var(--ct-spacing-xs)",
  minInlineSize: 0,
  paddingBlock: "var(--ct-spacing-xs)",
  borderBlockEnd: "var(--ct-border-hairline)",
} satisfies CSSProperties;
const labelStyle = {
  ...workbookTypography("metadata"),
  color: "var(--ct-colors-ink-muted)",
  minInlineSize: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const valueStyle = {
  ...workbookTypography("ui"),
  margin: 0,
  minInlineSize: 0,
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
