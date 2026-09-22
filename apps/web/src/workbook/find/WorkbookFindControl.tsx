import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import { Search } from "lucide-react";
import { type RefObject, useId, useLayoutEffect } from "react";
import { workbookQuietCommandStyle } from "../components/workbookFormStyles";
import { menuStyle } from "../components/workbookGridControlStyles";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import {
  type WorkbookFindSnapshot,
  workbookFindStatus,
} from "./WorkbookFindController";

const presentation = cartularyDesignPresentation.workbookFind;

export type WorkbookFindControlBinding = {
  snapshot: WorkbookFindSnapshot;
  available: boolean;
  hostRef: RefObject<HTMLDivElement | null>;
  inputRef: RefObject<HTMLTextAreaElement | null>;
  capture: (target?: EventTarget | null) => void;
  open: () => void;
  close: () => unknown;
  changeTerm: (value: string) => void;
  changeCase: (value: boolean) => void;
  navigate: (direction: 1 | -1) => unknown;
};

export function WorkbookFindControl({
  binding,
  chromeMode,
}: {
  binding: WorkbookFindControlBinding;
  chromeMode: WorkbookChromeMode;
}) {
  const { snapshot, inputRef } = binding;
  const id = useId();
  useLayoutEffect(() => {
    if (snapshot.open) inputRef.current?.focus();
  }, [snapshot.open, inputRef]);
  const compact =
    chromeMode === "compact_desktop" ||
    chromeMode === "below_supported_minimum";
  const status = workbookFindStatus(snapshot);
  const canNavigate =
    snapshot.status === "ready" &&
    snapshot.matches.length > 0 &&
    !snapshot.navigating;
  return (
    <div
      ref={binding.hostRef}
      data-grid-editor-external-action="true"
      style={{ position: "relative" }}
    >
      <button
        aria-label={presentation.scopeLabel}
        aria-expanded={snapshot.open}
        aria-controls={`${id}-panel`}
        type="button"
        title={presentation.scopeLabel}
        disabled={!binding.available}
        style={buttonStyle}
        onPointerDown={() => binding.capture()}
        onFocus={(event) => binding.capture(event.relatedTarget)}
        onClick={binding.open}
      >
        <Search aria-hidden="true" size={16} />
        {compact ? null : "Find"}
      </button>
      {snapshot.open ? (
        <section
          id={`${id}-panel`}
          aria-label={presentation.scopeLabel}
          style={{
            position: "absolute",
            insetBlockStart: "calc(100% + 0.5rem)",
            insetInlineEnd: 0,
            inlineSize: "min(25rem, calc(100vw - 2rem))",
            padding: "var(--ct-spacing-sm)",
            maxBlockSize: menuStyle.maxBlockSize,
            overflowY: menuStyle.overflowY,
            background: "var(--ct-colors-surface-1)",
            color: "var(--ct-colors-ink)",
            border: "var(--ct-border-hairline)",
            borderRadius: "var(--ct-rounded-xs)",
            boxShadow: "0 4px 12px #0003",
            zIndex: 30,
            display: "grid",
            gap: "0.5rem",
          }}
          onKeyDown={(event) => {
            if (event.nativeEvent.isComposing) return;
            if (event.key === "Escape") {
              event.preventDefault();
              event.stopPropagation();
              void binding.close();
            }
            if (event.key === "Enter" && event.target === inputRef.current) {
              event.preventDefault();
              event.stopPropagation();
              if (canNavigate) void binding.navigate(event.shiftKey ? -1 : 1);
            }
          }}
        >
          <label htmlFor={`${id}-input`}>Find in loaded rows</label>
          <textarea
            rows={1}
            ref={inputRef}
            id={`${id}-input`}
            autoComplete="off"
            spellCheck={false}
            aria-describedby={`${id}-help ${id}-status`}
            value={snapshot.term}
            onChange={(event) => binding.changeTerm(event.currentTarget.value)}
            style={{
              ...buttonStyle,
              boxSizing: "border-box",
              resize: "vertical",
              inlineSize: "100%",
              minInlineSize: 0,
            }}
          />
          <div
            style={{
              display: "flex",
              gap: "0.5rem",
              alignItems: "center",
              flexWrap: "wrap",
            }}
          >
            <label>
              <input
                type="checkbox"
                checked={snapshot.matchCase}
                onChange={(event) =>
                  binding.changeCase(event.currentTarget.checked)
                }
              />{" "}
              Match case
            </label>
            <button
              type="button"
              style={buttonStyle}
              disabled={!canNavigate}
              onClick={() => void binding.navigate(-1)}
            >
              Previous
            </button>
            <button
              type="button"
              style={buttonStyle}
              disabled={!canNavigate}
              onClick={() => void binding.navigate(1)}
            >
              Next
            </button>
            <button
              type="button"
              style={buttonStyle}
              aria-label="Close Find"
              onClick={() => void binding.close()}
            >
              Close
            </button>
          </div>
          <p id={`${id}-status`} style={{ margin: 0 }}>
            {status}
          </p>
          <p id={`${id}-help`} style={{ margin: 0, fontSize: "0.85em" }}>
            {presentation.scopeHelp}
            {snapshot.stale
              ? " Showing retained rows; refresh has not completed."
              : ""}
          </p>
        </section>
      ) : null}
      <span
        role="status"
        aria-live={presentation.live}
        aria-atomic="true"
        style={{
          position: "absolute",
          width: 1,
          height: 1,
          padding: 0,
          margin: -1,
          overflow: "hidden",
          clipPath: "inset(50%)",
          whiteSpace: "nowrap",
        }}
      >
        {snapshot.active ? status : ""}
      </span>
    </div>
  );
}
const buttonStyle = workbookQuietCommandStyle;
