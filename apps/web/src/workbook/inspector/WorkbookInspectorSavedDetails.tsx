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
  feedback,
}: {
  readonly contract: ViewContract;
  readonly row: WorkbookQueryRow;
  readonly fields?: ReadonlyMap<string, FieldPresentation>;
  readonly feedback?: ReactNode;
}) {
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
        const kind = inspectorSavedValueKind(field);
        const slots = fields?.get(field.fieldKey);
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
              {slots?.controls}
            </dd>
            {slots?.attachment}
          </div>
        );
      })}
      {feedback}
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
