import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import { useEffect, useId, useState } from "react";
import {
  workbookFormInputStyle,
  workbookTypography,
} from "./workbookFormStyles";

/** Controlled discovery presentation; requests, selection and authority stay with the owner. */
export function WorkbookSearchableCandidateChooser({
  candidates,
  search,
  selected,
  disabled,
  busy,
  testId,
  onSearch,
  onSelect,
  onComposition,
}: {
  readonly candidates: readonly {
    readonly recordId: string;
    readonly displayText: string;
  }[];
  readonly search: string;
  readonly selected: {
    readonly recordId: string;
    readonly displayText: string;
  } | null;
  readonly disabled: boolean;
  readonly busy: boolean;
  readonly testId: string;
  readonly onSearch: (text: string) => void;
  readonly onSelect: (id: string) => void;
  readonly onComposition: (composing: boolean) => void;
}) {
  const id = useId();
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState<string | null>(null);
  const optionId = (recordId: string) => `${id}-${recordId}`;
  const activeCandidate = candidates.find(
    (candidate) => candidate.recordId === active,
  );
  useEffect(() => {
    if (open && activeCandidate)
      document
        .getElementById(`${id}-${activeCandidate.recordId}`)
        ?.scrollIntoView?.({ block: "nearest" });
  }, [open, activeCandidate, id]);
  return (
    <div>
      <label htmlFor={id}>Search targets</label>
      <input
        id={id}
        role="combobox"
        aria-autocomplete="list"
        aria-expanded={open && !disabled}
        aria-controls={open && !disabled ? `${id}-list` : undefined}
        aria-activedescendant={
          open && !disabled && activeCandidate
            ? optionId(activeCandidate.recordId)
            : undefined
        }
        aria-busy={busy}
        autoComplete="off"
        disabled={disabled}
        data-testid={testId}
        style={{ ...workbookFormInputStyle, inlineSize: "100%" }}
        value={search}
        onFocus={() => setOpen(true)}
        onBlur={() => setOpen(false)}
        onChange={(event) => {
          setActive(null);
          setOpen(true);
          onSearch(event.currentTarget.value);
        }}
        onCompositionStart={() => onComposition(true)}
        onCompositionEnd={() => onComposition(false)}
        onKeyDown={(event) => {
          if (event.nativeEvent.isComposing) return;
          if (event.key === "Escape" && open) {
            event.preventDefault();
            event.stopPropagation();
            setOpen(false);
            setActive(null);
            return;
          }
          if (["ArrowDown", "ArrowUp"].includes(event.key)) {
            event.preventDefault();
            setOpen(true);
            const index = candidates.findIndex(
              (candidate) => candidate.recordId === active,
            );
            const next =
              event.key === "ArrowDown"
                ? Math.min(index + 1, candidates.length - 1)
                : index < 0
                  ? candidates.length - 1
                  : Math.max(0, index - 1);
            setActive(candidates[next]?.recordId ?? null);
          } else if (event.key === "Enter" && open) {
            event.preventDefault();
            event.stopPropagation();
            if (activeCandidate) {
              onSelect(activeCandidate.recordId);
              setOpen(false);
            }
          }
        }}
      />
      {open && !disabled ? (
        <div
          id={`${id}-list`}
          role="listbox"
          aria-label="Matching targets"
          style={{
            maxBlockSize: `${cartularyDesignPresentation.inspector.candidateChooserMaxBlockSizeRem}rem`,
            overflow: "auto",
            border: "var(--ct-border-hairline)",
            background: "var(--ct-colors-surface-2)",
          }}
        >
          {candidates.map((candidate) => (
            <button
              type="button"
              tabIndex={-1}
              key={candidate.recordId}
              id={optionId(candidate.recordId)}
              role="option"
              aria-selected={active === candidate.recordId}
              onPointerDown={(event) => event.preventDefault()}
              onClick={() => {
                onSelect(candidate.recordId);
                setActive(candidate.recordId);
                setOpen(false);
              }}
              style={{
                display: "block",
                inlineSize: "100%",
                border: 0,
                color: "inherit",
                font: "inherit",
                textAlign: "start",
                padding: "var(--ct-spacing-xs)",
                minBlockSize: "var(--ct-component-button-quiet-minBlockSize)",
                overflowWrap: "anywhere",
                cursor: "pointer",
                background:
                  candidate.recordId === active
                    ? "var(--ct-colors-surface-3)"
                    : "transparent",
              }}
            >
              {candidate.displayText}
              {candidates.some(
                (other) =>
                  other !== candidate &&
                  other.displayText === candidate.displayText,
              ) ? (
                <small style={{ display: "block" }}>{candidate.recordId}</small>
              ) : null}
            </button>
          ))}
        </div>
      ) : null}
      {selected ? (
        <p
          style={{
            ...workbookTypography("metadata"),
            marginBlock: "var(--ct-spacing-xs)",
            overflowWrap: "anywhere",
          }}
        >
          Selected: {selected.displayText}
        </p>
      ) : null}
    </div>
  );
}
