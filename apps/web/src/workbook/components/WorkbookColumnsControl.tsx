import {
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  useCallback,
  useEffect,
  useId,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import type {
  WorkbookColumnSizingControls,
  WorkbookFrozenColumnControls,
} from "../layout/WorkbookColumnLayoutController";
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

type ColumnFocusAction = "earlier" | "later" | "freeze" | "unfreeze";

export function WorkbookColumnsControl({
  isOpen,
  onClose,
  onCommand,
  onToggle,
  projection,
  sizing,
  freezing,
  surface,
}: {
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly sizing: WorkbookColumnSizingControls;
  readonly freezing: WorkbookFrozenColumnControls;
  readonly surface: string;
}) {
  const root = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const panel = useRef<HTMLDivElement>(null);
  const widthButtons = useRef(new Map<string, HTMLButtonElement>());
  const earlierButtons = useRef(new Map<string, HTMLButtonElement>());
  const laterButtons = useRef(new Map<string, HTMLButtonElement>());
  const freezeButtons = useRef(new Map<string, HTMLButtonElement>());
  const pendingActionFocus = useRef<{
    readonly source: HTMLButtonElement;
    readonly fieldKey: string;
    readonly action: ColumnFocusAction;
    readonly panel: HTMLDivElement;
  } | null>(null);
  const [, setActionRevision] = useState(0);
  const [field, setField] = useState<string | null>(null);
  const widthContext = useRef<{
    readonly fieldKey: string;
    readonly panel: HTMLDivElement;
    readonly top: number;
    readonly left: number;
  } | null>(null);
  const widthReturn = useRef<{
    readonly context: NonNullable<typeof widthContext.current>;
    readonly source: Element;
  } | null>(null);
  const focusAndReveal = useCallback(
    (target: HTMLElement | null | undefined) => {
      if (isOpen) focusAndRevealColumnsControl(panel.current, target);
    },
    [isOpen],
  );
  const clearFocusWork = useCallback(() => {
    pendingActionFocus.current = null;
    widthReturn.current = null;
    widthContext.current = null;
  }, []);
  useLayoutEffect(() => {
    const request = widthReturn.current;
    widthReturn.current = null;
    if (
      !request ||
      !isOpen ||
      field !== null ||
      panel.current !== request.context.panel
    )
      return;
    const active = document.activeElement;
    if (active !== request.source && active !== document.body) return;
    const target = widthButtons.current.get(request.context.fieldKey);
    if (!target?.isConnected || target.disabled) return;
    panel.current.scrollTop = request.context.top;
    panel.current.scrollLeft = request.context.left;
    focusAndReveal(target);
  }, [field, isOpen, focusAndReveal]);
  const [notice, setNotice] = useState("");
  const cancel = sizing.cancel;
  useEffect(() => () => cancel(), [cancel]);
  useLayoutEffect(() => {
    if (isOpen)
      panel.current
        ?.querySelector<HTMLInputElement>('input[type="checkbox"]')
        ?.focus({ preventScroll: true });
    else {
      pendingActionFocus.current = null;
      widthReturn.current = null;
      widthContext.current = null;
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
        clearFocusWork();
        cancel();
        onClose();
      }
    };
    document.addEventListener("pointerdown", outside);
    return () => document.removeEventListener("pointerdown", outside);
  }, [isOpen, onClose, cancel, clearFocusWork]);
  const close = () => {
    clearFocusWork();
    cancel();
    onClose();
    trigger.current?.focus({ preventScroll: true });
  };
  const closeWidth = () => {
    const source = document.activeElement;
    const context = widthContext.current;
    widthReturn.current =
      context && source && panel.current?.contains(source)
        ? { context, source }
        : null;
    widthContext.current = null;
    cancel();
    setField(null);
  };
  const selected = projection.columns.find(
    (column) => column.fieldKey === field,
  );
  const boundaryIndex = projection.columns.findIndex(
    (column) => column.fieldKey === projection.frozenThroughFieldKey,
  );
  const boundary = projection.columns[boundaryIndex];
  useLayoutEffect(() => {
    const request = pendingActionFocus.current;
    pendingActionFocus.current = null;
    if (
      !request ||
      !isOpen ||
      field !== null ||
      request.panel !== panel.current
    )
      return;
    if (
      !projection.columns.some((column) => column.fieldKey === request.fieldKey)
    )
      return;
    const active = document.activeElement;
    if (active !== request.source && active !== document.body) return;
    const eligible = (button: HTMLButtonElement | undefined) =>
      button?.isConnected && !button.disabled ? button : undefined;
    const target =
      request.action === "earlier"
        ? (eligible(earlierButtons.current.get(request.fieldKey)) ??
          eligible(laterButtons.current.get(request.fieldKey)))
        : request.action === "later"
          ? (eligible(laterButtons.current.get(request.fieldKey)) ??
            eligible(earlierButtons.current.get(request.fieldKey)))
          : request.action === "freeze"
            ? (eligible(freezeButtons.current.get(request.fieldKey)) ??
              eligible(widthButtons.current.get(request.fieldKey)))
            : eligible(freezeButtons.current.get(request.fieldKey));
    focusAndReveal(target);
  });
  const commandWithFocus = (
    source: HTMLButtonElement,
    fieldKey: string,
    action: ColumnFocusAction,
    command: WorkbookGridQueryCommand,
  ) => {
    if (document.activeElement === source && panel.current) {
      pendingActionFocus.current = {
        source,
        fieldKey,
        action,
        panel: panel.current,
      };
      // Consume even a command that leaves the layout unchanged in this commit.
      setActionRevision((revision) => revision + 1);
    }
    onCommand(command);
  };
  const visibleCount = projection.columns
    .slice(0, boundaryIndex + 1)
    .filter((column) => !column.hidden).length;
  const freezeMessage = boundary
    ? `Freeze through ${boundary.label}${boundary.hidden ? " (hidden)" : ""}. ${visibleCount} visible data ${visibleCount === 1 ? "column" : "columns"}.` +
      (visibleCount === 0
        ? " Show a column in this prefix to freeze it."
        : freezing.status?.kind === "suspended"
          ? freezing.status.reason === "insufficient_space"
            ? " Freezing paused: more scrollable space is needed."
            : " Freezing paused while grid geometry is unavailable."
          : "")
    : "";
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
        title={freezeMessage || undefined}
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
              pendingActionFocus.current &&
              (event.relatedTarget === null ||
                event.relatedTarget === document.body)
            )
              return;
            if (
              event.relatedTarget instanceof Node &&
              !root.current?.contains(event.relatedTarget)
            ) {
              clearFocusWork();
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
          {freezeMessage ? (
            <p>{freezeMessage}</p>
          ) : (
            <p>No frozen data columns.</p>
          )}
          {(sizing.notice ?? notice) ? (
            <p aria-hidden="true">{sizing.notice ?? notice}</p>
          ) : null}
          {selected ? (
            <ColumnWidthPanel
              key={selected.fieldKey}
              fieldKey={selected.fieldKey}
              label={selected.label}
              onCancel={closeWidth}
              onFocusReturn={focusAndReveal}
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
                    ref={(node) => {
                      if (node)
                        earlierButtons.current.set(column.fieldKey, node);
                      else earlierButtons.current.delete(column.fieldKey);
                    }}
                    type="button"
                    aria-label={`Move ${column.label} earlier`}
                    disabled={column.position === 0}
                    onClick={(event) =>
                      commandWithFocus(
                        event.currentTarget,
                        column.fieldKey,
                        "earlier",
                        {
                          kind: "column_move",
                          fieldKey: column.fieldKey,
                          direction: "earlier",
                        },
                      )
                    }
                  >
                    ↑
                  </button>
                  <button
                    style={controlButtonStyle}
                    ref={(node) => {
                      if (node) laterButtons.current.set(column.fieldKey, node);
                      else laterButtons.current.delete(column.fieldKey);
                    }}
                    type="button"
                    aria-label={`Move ${column.label} later`}
                    disabled={column.position === projection.columns.length - 1}
                    onClick={(event) =>
                      commandWithFocus(
                        event.currentTarget,
                        column.fieldKey,
                        "later",
                        {
                          kind: "column_move",
                          fieldKey: column.fieldKey,
                          direction: "later",
                        },
                      )
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
                      if (panel.current)
                        widthContext.current = {
                          fieldKey: column.fieldKey,
                          panel: panel.current,
                          top: panel.current.scrollTop,
                          left: panel.current.scrollLeft,
                        };
                      setNotice("");
                      setField(column.fieldKey);
                    }}
                  >
                    Width
                  </button>
                  <button
                    style={{
                      ...controlButtonStyle,
                      gridColumn: "1 / -1",
                      justifySelf: "start",
                    }}
                    ref={(node) => {
                      if (node)
                        freezeButtons.current.set(column.fieldKey, node);
                      else freezeButtons.current.delete(column.fieldKey);
                    }}
                    type="button"
                    aria-label={`Freeze through ${column.label}`}
                    disabled={
                      projection.frozenThroughFieldKey === column.fieldKey
                    }
                    onClick={(event) =>
                      commandWithFocus(
                        event.currentTarget,
                        column.fieldKey,
                        "freeze",
                        {
                          kind: "columns_freeze",
                          fieldKey: column.fieldKey,
                        },
                      )
                    }
                  >
                    Freeze through this column
                  </button>
                </div>
              ))}
              <div style={actionsStyle}>
                <button
                  style={controlButtonStyle}
                  type="button"
                  disabled={!boundary}
                  onClick={(event) => {
                    if (boundary)
                      commandWithFocus(
                        event.currentTarget,
                        boundary.fieldKey,
                        "unfreeze",
                        {
                          kind: "columns_freeze",
                          fieldKey: null,
                        },
                      );
                    setNotice("Columns unfrozen.");
                  }}
                >
                  Unfreeze columns
                </button>
                <button
                  style={controlButtonStyle}
                  type="button"
                  onClick={() => {
                    onCommand({ kind: "columns_reset" });
                    setNotice(
                      "Column order, visibility, widths and freezing reset.",
                    );
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
        {[freezeMessage, sizing.notice ?? notice].filter(Boolean).join(" ")}
      </span>
    </div>
  );
}

// Columns owns this scrollport, never the page or another workbook region.
function focusAndRevealColumnsControl(
  panel: HTMLDivElement | null,
  target: HTMLElement | null | undefined,
) {
  if (
    !panel?.isConnected ||
    !target?.isConnected ||
    !panel.contains(target) ||
    target.matches(":disabled")
  )
    return;
  if (document.activeElement !== target) target.focus({ preventScroll: true });
  if (
    document.activeElement !== target ||
    !panel.isConnected ||
    !panel.contains(target)
  )
    return;
  const viewport = panel.ownerDocument.defaultView;
  if (!viewport || panel.offsetWidth === 0 || panel.offsetHeight === 0) return;
  const outer = panel.getBoundingClientRect();
  const scaleX = outer.width / panel.offsetWidth;
  const scaleY = outer.height / panel.offsetHeight;
  if (scaleX <= 0 || scaleY <= 0) return;
  const box = target.getBoundingClientRect();
  const style = viewport.getComputedStyle(target);
  const ring = Math.max(
    0,
    (Number.parseFloat(style.outlineWidth) || 0) +
      (Number.parseFloat(style.outlineOffset) || 0),
  );
  const top = Math.max(0, outer.top + panel.clientTop * scaleY);
  const bottom = Math.min(
    viewport.innerHeight,
    outer.top + (panel.clientTop + panel.clientHeight) * scaleY,
  );
  const left = Math.max(0, outer.left + panel.clientLeft * scaleX);
  const right = Math.min(
    viewport.innerWidth,
    outer.left + (panel.clientLeft + panel.clientWidth) * scaleX,
  );
  if (bottom <= top || right <= left) return;
  const delta = (
    start: number,
    end: number,
    minimum: number,
    maximum: number,
  ) =>
    start < minimum || end - start > maximum - minimum
      ? start - minimum
      : end > maximum
        ? end - maximum
        : 0;
  panel.scrollTop +=
    delta(box.top - ring * scaleY, box.bottom + ring * scaleY, top, bottom) /
    scaleY;
  panel.scrollLeft +=
    delta(box.left - ring * scaleX, box.right + ring * scaleX, left, right) /
    scaleX;
}

function ColumnWidthPanel({
  fieldKey,
  label,
  onCancel,
  onFocusReturn,
  sizing,
}: {
  readonly fieldKey: string;
  readonly label: string;
  readonly onCancel: () => void;
  readonly onFocusReturn: (target: HTMLElement | null) => void;
  readonly sizing: WorkbookColumnSizingControls;
}) {
  const descriptor = sizing.read(fieldKey);
  const [text, setText] = useState(String(descriptor.width ?? ""));
  const [error, setError] = useState<string | null>(null);
  const draft = useRef(false);
  const editRevision = useRef(0);
  const input = useRef<HTMLInputElement>(null);
  const fitButton = useRef<HTMLButtonElement>(null);
  const restoreButton = useRef<HTMLButtonElement>(null);
  const fitOwnedFocus = useRef(false);
  const id = useId();
  const fitting = sizing.pendingField === fieldKey;
  useLayoutEffect(() => {
    if (
      fitting ||
      descriptor.unavailableReason === null ||
      !fitOwnedFocus.current
    )
      return;
    const active = document.activeElement;
    if (active === fitButton.current || active === document.body)
      onFocusReturn(restoreButton.current);
    fitOwnedFocus.current = false;
  }, [fitting, descriptor.unavailableReason, onFocusReturn]);
  useLayoutEffect(() => {
    if (!draft.current) setText(String(descriptor.width ?? ""));
  }, [descriptor.width]);
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
          sizing.cancel();
          setError("Enter a whole number from 40 to 4096.");
          return;
        }
        const outcome = sizing.applyWidth(fieldKey, widthPx);
        if (outcome.kind !== "completed") {
          if (outcome.kind === "unavailable") setError(outcome.reason);
          return;
        }
        draft.current = false;
        setText(String(outcome.widthPx ?? ""));
        setError(null);
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
          draft.current = true;
          editRevision.current += 1;
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
          ref={fitButton}
          style={controlButtonStyle}
          type="button"
          disabled={!fitting && descriptor.unavailableReason !== null}
          aria-busy={fitting || undefined}
          aria-describedby={
            fitting
              ? `${id}-fit-pending`
              : descriptor.unavailableReason
                ? `${id}-fit-unavailable`
                : undefined
          }
          onFocus={() => {
            fitOwnedFocus.current = true;
          }}
          onBlur={(event) => {
            if (event.relatedTarget && event.relatedTarget !== document.body)
              fitOwnedFocus.current = false;
          }}
          onClick={async () => {
            if (fitting) return;
            const startedAtRevision = editRevision.current;
            const outcome = await sizing.fitVisible(fieldKey);
            if (
              outcome.kind === "completed" &&
              editRevision.current === startedAtRevision
            ) {
              draft.current = false;
              setText(String(outcome.widthPx ?? ""));
              setError(null);
            }
          }}
        >
          Fit visible content
        </button>
        <button
          ref={restoreButton}
          style={controlButtonStyle}
          type="button"
          onClick={() => {
            const outcome = sizing.restoreDefault(fieldKey);
            if (outcome.kind === "completed") {
              draft.current = false;
              setText(String(outcome.widthPx ?? ""));
              setError(null);
            } else if (outcome.kind === "unavailable") setError(outcome.reason);
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
        <p id={`${id}-fit-unavailable`}>{descriptor.unavailableReason}</p>
      ) : null}
      {fitting ? (
        <p id={`${id}-fit-pending`}>Measuring visible content…</p>
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
