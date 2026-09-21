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

type FieldPresentation = {
  readonly controls: ReactNode;
  readonly attachment: ReactNode;
};
/** Reads only accepted values. Optional presentation slots do not own editing state. */
export function WorkbookInspectorSavedDetails({
  contract,
  row,
  fields,
  describedBy,
}: {
  readonly contract: ViewContract;
  readonly row: WorkbookQueryRow;
  readonly fields?: ReadonlyMap<string, FieldPresentation>;
  readonly describedBy?: string | undefined;
}) {
  return (
    <dl style={detailsStyle} aria-describedby={describedBy}>
      <style>{`
        [data-inspector-field-layout="property"] { grid-template-columns: minmax(0, 2fr) minmax(0, 3fr) auto; }
        [data-inspector-field-layout="property"] > [data-inspector-field-actions] { grid-column: 3; grid-row: 1; }
        @container inspector-fields (width < ${cartularyDesignPresentation.inspector.propertyStackBelowPx}px) {
          [data-inspector-field-layout="property"] { grid-template-columns: minmax(0, 1fr) auto; }
          [data-inspector-field-layout="property"] > dt { grid-column: 1 / -1; }
          [data-inspector-field-layout="property"] > [data-inspector-field-value] { grid-column: 1; grid-row: 2; }
          [data-inspector-field-layout="property"] > [data-inspector-field-actions] { grid-column: 2; grid-row: 2; }
        }
      `}</style>
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
                : typeof value === "boolean"
                  ? value
                    ? "True"
                    : "False"
                  : typeof value === "string" && field.readKind === "text"
                    ? value
                    : genericCellLabelForField(
                        contract.viewSchemaId,
                        field.fieldKey,
                        value,
                      );
        const kind = inspectorSavedValueKind(field);
        const override =
          cartularyDesignPresentation.inspector.fieldLayoutOverrides.find(
            (layout) =>
              layout.viewSchemaId === contract.viewSchemaId &&
              layout.fieldKey === field.fieldKey,
          );
        const property = override
          ? override.layout === "property"
          : kind === "scalar";
        const slots = fields?.get(field.fieldKey);
        return (
          <div
            key={field.fieldKey}
            tabIndex={-1}
            data-inspector-saved-field={field.fieldKey}
            data-inspector-value-kind={kind}
            data-inspector-field-layout={property ? "property" : "narrative"}
            style={{
              ...fieldStyle,
              ...(!property
                ? { gridTemplateColumns: "minmax(0, 1fr) auto" }
                : {}),
            }}
          >
            <dt style={labelStyle}>{field.label}</dt>
            <dd
              data-inspector-field-value
              style={{
                ...valueStyle,
                ...(!property ? { gridColumn: "1 / -1", gridRow: 2 } : {}),
              }}
            >
              <SavedValue
                key={JSON.stringify([row.record_id, field.fieldKey])}
                value={saved}
                label={field.label}
              />
            </dd>
            <dd
              data-inspector-field-actions
              style={{
                margin: 0,
                ...(!property ? { gridColumn: 2, gridRow: 1 } : {}),
              }}
            >
              {slots?.controls}
            </dd>
            {slots?.attachment}
          </div>
        );
      })}
    </dl>
  );
}

/** Consume semantic source metadata; never infer field meaning from its label. */
function inspectorSavedValueKind(
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
      {value !== "" && value.trim() === "" ? (
        <span style={workbookTypography("metadata")}>Whitespace only</span>
      ) : null}
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
      {value !== "" && value.trim() === "" ? (
        <details>
          <summary>Inspect source whitespace</summary>
          <code style={{ whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}>
            {JSON.stringify(value)}
          </code>
        </details>
      ) : null}
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
  containerType: "inline-size",
  containerName: "inspector-fields",
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
