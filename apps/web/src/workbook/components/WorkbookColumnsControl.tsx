import {
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import { useEffect, useId, useLayoutEffect, useRef, useState } from "react";
import type { WorkbookColumnSizingControls } from "../layout/WorkbookColumnLayoutController";
import {
  parseWorkbookColumnWidth,
  workbookColumnSizing,
} from "../models/workbookColumnSizing";
import type {
  WorkbookGridQueryCommand,
  WorkbookGridQueryControlProjection,
} from "../models/workbookGridQueryControls";
import {
  controlButtonStyle,
  fixedMenuFrameStyle,
  inputStyle,
  menuStyle,
} from "./workbookGridControlStyles";

export function WorkbookColumnsControl({
  isOpen,
  onClose,
  onCommand,
  onToggle,
  projection,
  sizing,
  surface,
}: {
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly sizing: WorkbookColumnSizingControls;
  readonly surface: string;
}) {
  const root = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const panel = useRef<HTMLDivElement>(null);
  const widthButtons = useRef(new Map<string, HTMLButtonElement>());
  const [field, setField] = useState<string | null>(null);
  const returnField = useRef<string | null>(null);
  useLayoutEffect(() => {
    if (field === null && returnField.current) {
      widthButtons.current
        .get(returnField.current)
        ?.focus({ preventScroll: true });
      returnField.current = null;
    }
  }, [field]);
  const [notice, setNotice] = useState("");
  const cancel = sizing.cancel;
  useEffect(() => () => cancel(), [cancel]);
  useLayoutEffect(() => {
    if (isOpen)
      panel.current
        ?.querySelector<HTMLInputElement>('input[type="checkbox"]')
        ?.focus({ preventScroll: true });
    else {
      setField(null);
      cancel();
    }
  }, [isOpen, cancel]);
  useEffect(() => {
    if (!isOpen) return;
    const outside = (event: PointerEvent) => {
      if (
        event.target instanceof Node &&
        !root.current?.contains(event.target)
      ) {
        cancel();
        onClose();
      }
    };
    document.addEventListener("pointerdown", outside);
    return () => document.removeEventListener("pointerdown", outside);
  }, [isOpen, onClose, cancel]);
  const close = () => {
    cancel();
    onClose();
    trigger.current?.focus({ preventScroll: true });
  };
  const closeWidth = () => {
    cancel();
    returnField.current = field;
    setField(null);
  };
  const selected = projection.columns.find(
    (column) => column.fieldKey === field,
  );
  return (
    <div
      ref={root}
      style={fixedMenuFrameStyle}
      data-grid-editor-external-action="true"
    >
      <button
        style={controlButtonStyle}
        ref={trigger}
        aria-controls={isOpen ? workbookColumnsMenuTestId(surface) : undefined}
        aria-expanded={isOpen}
        aria-haspopup="dialog"
        data-testid={workbookColumnsMenuTriggerTestId(surface)}
        type="button"
        onClick={() => {
          if (isOpen) close();
          else onToggle();
        }}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown") {
            event.preventDefault();
            if (!isOpen) onToggle();
          }
        }}
      >
        Columns
      </button>
      {isOpen ? (
        <div
          ref={panel}
          aria-label="Column controls"
          data-testid={workbookColumnsMenuTestId(surface)}
          id={workbookColumnsMenuTestId(surface)}
          role="dialog"
          style={columnsPanelStyle}
          tabIndex={-1}
          onBlur={(event) => {
            if (
              event.relatedTarget instanceof Node &&
              !root.current?.contains(event.relatedTarget)
            ) {
              cancel();
              onClose();
            }
          }}
          onKeyDown={(event) => {
            if (event.key !== "Escape" || event.defaultPrevented) return;
            event.preventDefault();
            event.stopPropagation();
            if (field !== null) closeWidth();
            else close();
          }}
        >
          {(sizing.notice ?? notice) ? (
            <p aria-hidden="true">{sizing.notice ?? notice}</p>
          ) : null}
          {selected ? (
            <ColumnWidthPanel
              key={selected.fieldKey}
              fieldKey={selected.fieldKey}
              label={selected.label}
              onCancel={closeWidth}
              onNotice={setNotice}
              sizing={sizing}
            />
          ) : (
            <>
              {projection.columns.map((column) => (
                <div key={column.fieldKey} style={columnRowStyle}>
                  <label style={{ minInlineSize: 0, overflowWrap: "anywhere" }}>
                    <input
                      type="checkbox"
                      checked={!column.hidden}
                      onChange={(event) =>
                        onCommand({
                          kind: "column_set_hidden",
                          fieldKey: column.fieldKey,
                          hidden: !event.currentTarget.checked,
                        })
                      }
                    />{" "}
                    {column.label}
                  </label>
                  <button
                    style={controlButtonStyle}
                    type="button"
                    aria-label={`Move ${column.label} earlier`}
                    disabled={column.position === 0}
                    onClick={() =>
                      onCommand({
                        kind: "column_move",
                        fieldKey: column.fieldKey,
                        direction: "earlier",
                      })
                    }
                  >
                    ↑
                  </button>
                  <button
                    style={controlButtonStyle}
                    type="button"
                    aria-label={`Move ${column.label} later`}
                    disabled={column.position === projection.columns.length - 1}
                    onClick={() =>
                      onCommand({
                        kind: "column_move",
                        fieldKey: column.fieldKey,
                        direction: "later",
                      })
                    }
                  >
                    ↓
                  </button>
                  <button
                    style={controlButtonStyle}
                    ref={(node) => {
                      if (node) widthButtons.current.set(column.fieldKey, node);
                      else widthButtons.current.delete(column.fieldKey);
                    }}
                    type="button"
                    aria-label={`Width for ${column.label}`}
                    onClick={() => {
                      setNotice("");
                      setField(column.fieldKey);
                    }}
                  >
                    Width
                  </button>
                </div>
              ))}
              <div style={actionsStyle}>
                <button
                  style={controlButtonStyle}
                  type="button"
                  onClick={() => {
                    onCommand({ kind: "columns_reset" });
                    setNotice("Column order, visibility and widths reset.");
                  }}
                >
                  Reset columns
                </button>
                <button
                  style={controlButtonStyle}
                  type="button"
                  onClick={close}
                >
                  Close columns
                </button>
              </div>
            </>
          )}
        </div>
      ) : null}
      <span role="status" style={statusStyle}>
        {sizing.notice ?? notice}
      </span>
    </div>
  );
}

function ColumnWidthPanel({
  fieldKey,
  label,
  onCancel,
  onNotice,
  sizing,
}: {
  readonly fieldKey: string;
  readonly label: string;
  readonly onCancel: () => void;
  readonly onNotice: (message: string) => void;
  readonly sizing: WorkbookColumnSizingControls;
}) {
  const descriptor = sizing.read(fieldKey);
  const [text, setText] = useState(String(descriptor.width ?? ""));
  const [error, setError] = useState<string | null>(null);
  const input = useRef<HTMLInputElement>(null);
  const id = useId();
  useLayoutEffect(() => {
    input.current?.focus({ preventScroll: true });
    input.current?.select();
  }, []);
  return (
    <form
      aria-label={`Width for ${label}`}
      onSubmit={(event) => {
        event.preventDefault();
        const widthPx = parseWorkbookColumnWidth(text);
        if (widthPx === null) {
          setError("Enter a whole number from 40 to 4096.");
          return;
        }
        setError(null);
        sizing.onIntent({ kind: "set_width", fieldKey, widthPx });
        onNotice(`${label} width set to ${widthPx} px.`);
      }}
    >
      <strong style={{ overflowWrap: "anywhere" }}>{label}</strong>
      <p>
        Current: {descriptor.width ?? "unavailable"} px · Default:{" "}
        {descriptor.defaultWidth ?? "unavailable"} px
        {descriptor.overridden ? " · Custom width" : ""}
      </p>
      <label htmlFor={id}>Width in CSS pixels</label>
      <input
        ref={input}
        id={id}
        inputMode="numeric"
        type="text"
        value={text}
        aria-invalid={error !== null}
        aria-describedby={error ? `${id}-error` : `${id}-range`}
        style={{ ...inputStyle, inlineSize: "100%", minInlineSize: 0 }}
        onChange={(event) => {
          setText(event.currentTarget.value);
          setError(null);
        }}
      />
      <p id={`${id}-range`}>
        {workbookColumnSizing.minimumWidthPx}–
        {workbookColumnSizing.maximumWidthPx}, whole pixels.
      </p>
      {error ? (
        <p id={`${id}-error`} role="alert">
          {error}
        </p>
      ) : null}
      <div style={actionsStyle}>
        <button style={controlButtonStyle} type="submit">
          Apply width
        </button>
        <button
          style={controlButtonStyle}
          type="button"
          disabled={
            descriptor.unavailableReason !== null ||
            sizing.pendingField === fieldKey
          }
          onClick={() => {
            setError(null);
            sizing.onIntent({ kind: "fit_visible", fieldKey });
          }}
        >
          Fit visible content
        </button>
        <button
          style={controlButtonStyle}
          type="button"
          onClick={() => {
            setError(null);
            sizing.restoreDefault(fieldKey);
            setText(String(descriptor.defaultWidth ?? ""));
          }}
        >
          Restore default
        </button>
        <button style={controlButtonStyle} type="button" onClick={onCancel}>
          Cancel
        </button>
      </div>
      <p>
        Fit measures the header and committed cells currently visible in the
        viewport. Drafts and unloaded content are excluded.
      </p>
      {descriptor.unavailableReason ? (
        <p>{descriptor.unavailableReason}</p>
      ) : null}
      {sizing.pendingField === fieldKey ? (
        <p>Measuring visible content…</p>
      ) : null}
    </form>
  );
}
const columnRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) 2rem 2rem auto",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
};
const actionsStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "var(--ct-spacing-xs)",
  marginBlockStart: "var(--ct-spacing-sm)",
};
const columnsPanelStyle = {
  ...menuStyle,
  boxSizing: "border-box" as const,
  insetInlineStart: "auto",
  insetInlineEnd: 0,
  inlineSize: "min(var(--ct-layout-viewBarOverlayMaxInlineSize), 92vw)",
  maxBlockSize: "min(32rem, 70vh)",
  overflowY: "auto" as const,
  padding: "var(--ct-spacing-sm)",
};
const statusStyle = {
  position: "absolute" as const,
  inlineSize: "1px",
  blockSize: "1px",
  overflow: "hidden",
  clipPath: "inset(50%)",
};
