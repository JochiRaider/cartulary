import { useContext, useRef, useState } from "react";
import { WorkbookCandidateAuthorityContext } from "../hooks/useWorkbookCandidateDiscovery";
import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookCandidate } from "../ports/WorkbookCandidateReadPort";
import { useSelectedReferenceRemovalFocus } from "./useSelectedReferenceRemovalFocus";
import { WorkbookRecordCandidatePicker } from "./WorkbookRecordCandidatePicker";

/** Only current-page membership changes on page selection. Removal is explicit. */
export function WorkbookCandidateSelection<T extends WorkbookCandidate>({
  candidates,
  selected,
  label,
  testId,
  multiple,
  maximum,
  disabled,
  concealed = false,
  scopeKey = testId,
  onChange,
}: {
  readonly candidates: readonly T[];
  readonly selected: readonly T[];
  readonly label: string;
  readonly testId: string;
  readonly multiple: boolean;
  readonly maximum: number;
  readonly disabled: boolean;
  readonly concealed?: boolean;
  readonly scopeKey?: string;
  readonly onChange: (selected: readonly T[]) => void;
}) {
  const [error, setError] = useState<string | null>(null);
  const authority = useContext(WorkbookCandidateAuthorityContext);
  const selector = useRef<HTMLSelectElement>(null);
  const group = useRef<HTMLFieldSetElement>(null);
  const removalFocus = useSelectedReferenceRemovalFocus({
    ids: selected.map((item) => item.recordId),
    scopeKey: `${scopeKey}:${authority.identity}`,
    disabled: disabled || concealed || !authority.canRead,
    fallback: () => selector.current,
    groupRef: group,
  });
  return (
    <fieldset
      ref={group}
      aria-label={`${label} selected references`}
      data-reference-focus-fallback
      tabIndex={-1}
      style={{
        display: "grid",
        gap: "var(--ct-spacing-xs)",
        minWidth: 0,
        border: 0,
        margin: 0,
        padding: 0,
      }}
    >
      <WorkbookRecordCandidatePicker
        selectorRef={selector}
        candidates={concealed ? [] : candidates}
        label={label}
        testId={testId}
        selection={multiple ? "multiple" : "single"}
        disabled={
          disabled || concealed || !authority.canRead || !candidates.length
        }
        selectedRecordIds={selected
          .filter((item) =>
            candidates.some(
              (candidate) => candidate.recordId === item.recordId,
            ),
          )
          .map((item) => item.recordId)}
        onSelectedRecordIdsChange={(ids) => {
          const retained = multiple
            ? selected.filter(
                (item) =>
                  !candidates.some(
                    (candidate) => candidate.recordId === item.recordId,
                  ),
              )
            : [];
          const next = [
            ...retained,
            ...ids.flatMap((id) => {
              const item =
                selected.find((item) => item.recordId === id) ??
                candidates.find((item) => item.recordId === id);
              return item ? [item] : [];
            }),
          ];
          if (next.length > maximum) {
            setError(
              `Choose at most ${maximum} references. Existing selections are retained.`,
            );
            return;
          }
          setError(null);
          onChange(next);
        }}
      />
      <span>
        {selected.length} selected{multiple ? ` (maximum ${maximum})` : ""}.
        Selections outside this page are retained.
      </span>
      {selected.length ? (
        <ul style={{ margin: 0, paddingInlineStart: "var(--ct-spacing-lg)" }}>
          {selected.map((item) => (
            <li key={item.recordId} style={{ overflowWrap: "anywhere" }}>
              {concealed
                ? "Selected reference"
                : item.displayText || item.recordId}{" "}
              <Button
                ref={removalFocus.buttonRef(item.recordId)}
                type="button"
                tone="secondary"
                disabled={disabled || concealed || !authority.canRead}
                aria-label={`Remove selected ${label} ${concealed ? "reference" : item.displayText || item.recordId}${!concealed && selected.filter((value) => value.displayText === item.displayText).length > 1 ? ` (${item.recordId})` : ""}`}
                onClick={(event) =>
                  removalFocus.remove(
                    item.recordId,
                    event.currentTarget,
                    () => {
                      setError(null);
                      onChange(
                        selected.filter(
                          (value) => value.recordId !== item.recordId,
                        ),
                      );
                    },
                  )
                }
              >
                Remove
              </Button>
            </li>
          ))}
        </ul>
      ) : null}
      {error ? <p role="alert">{error}</p> : null}
    </fieldset>
  );
}
