import { useState } from "react";
import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookCandidate } from "../ports/WorkbookCandidateReadPort";
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
  readonly onChange: (selected: readonly T[]) => void;
}) {
  const [error, setError] = useState<string | null>(null);
  return (
    <div style={{ display: "grid", gap: "var(--ct-spacing-xs)", minWidth: 0 }}>
      <WorkbookRecordCandidatePicker
        candidates={concealed ? [] : candidates}
        label={label}
        testId={testId}
        selection={multiple ? "multiple" : "single"}
        disabled={disabled || concealed || !candidates.length}
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
                type="button"
                tone="secondary"
                disabled={disabled || concealed}
                aria-label={`Remove selected ${label} ${concealed ? "reference" : item.displayText || item.recordId}`}
                onClick={() => {
                  setError(null);
                  onChange(
                    selected.filter(
                      (value) => value.recordId !== item.recordId,
                    ),
                  );
                }}
              >
                Remove
              </Button>
            </li>
          ))}
        </ul>
      ) : null}
      {error ? <p role="alert">{error}</p> : null}
    </div>
  );
}
