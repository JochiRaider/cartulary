import { assessmentCreateControlTestId } from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { useCallback, useMemo, useRef, useState } from "react";
import { WorkbookRecordCandidatePicker } from "../../components/WorkbookRecordCandidatePicker";
import {
  inputStyle,
  secondaryButtonStyle,
  stackedLabelStyle,
} from "../../components/workbookGridControlStyles";
import { useWorkbookCandidates } from "../../hooks/useWorkbookCandidates";
import type {
  AssessmentCreateDraft,
  AssessmentSupportCandidate,
} from "../../models/assessmentWorkbookModel";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import type {
  AssessmentCandidateQuery,
  AssessmentCandidateReadPort,
} from "./assessmentCandidatePort";

type Props = {
  readonly reader: AssessmentCandidateReadPort;
  readonly draft: AssessmentCreateDraft;
  readonly disabled: boolean;
  readonly revision: number | string;
  readonly update: (
    update: (draft: AssessmentCreateDraft) => AssessmentCreateDraft,
  ) => void;
};
export function AssessmentSubjectPicker({
  reader,
  draft,
  disabled,
  revision,
  update,
}: Props) {
  const read = useCallback(
    (input: AssessmentCandidateQuery) =>
      reader.subjects(draft.subjectType, input),
    [reader, draft.subjectType],
  );
  return (
    <CandidateList
      read={read}
      revision={revision}
      view={
        draft.subjectType === "host"
          ? hostsViewSchemaId
          : identitiesViewSchemaId
      }
      label="Subject"
      testId={assessmentCreateControlTestId("subject")}
      disabled={disabled}
      multiple={false}
      selected={draft.subjectRecordId ? [draft.subjectRecordId] : []}
      labels={
        draft.subjectRecordId
          ? {
              [draft.subjectRecordId]:
                draft.subjectDisplayText ?? draft.subjectRecordId,
            }
          : {}
      }
      onChange={(ids, labels) =>
        update((current) => ({
          ...current,
          subjectRecordId: ids[0] ?? "",
          subjectDisplayText: labels[ids[0] ?? ""] ?? "",
        }))
      }
    />
  );
}

export function AssessmentSupportPicker({
  reader,
  draft,
  disabled,
  revision,
  update,
}: Props) {
  const [staged, setStaged] = useState<{
    ids: string[];
    labels: Readonly<Record<string, string>>;
  } | null>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const read = useCallback(
    (input: AssessmentCandidateQuery) => reader.support(input),
    [reader],
  );
  function close() {
    setStaged(null);
    trigger.current?.focus();
  }
  return (
    <section aria-label="Assessment supporting records">
      <p>Supporting records ({draft.supportRecordIds.length}/64)</p>
      {draft.supportRecordIds.length === 0 ? (
        <p>No supporting records selected.</p>
      ) : (
        <ul>
          {draft.supportRecordIds.map((id) => (
            <li key={id} style={{ overflowWrap: "anywhere" }}>
              {draft.supportDisplayText?.[id] ?? id}
            </li>
          ))}
        </ul>
      )}
      <button
        ref={trigger}
        style={secondaryButtonStyle}
        type="button"
        disabled={disabled}
        onClick={() =>
          setStaged({
            ids: [...draft.supportRecordIds],
            labels: draft.supportDisplayText ?? {},
          })
        }
      >
        Choose support
      </button>
      {staged && (
        <fieldset
          style={{ border: 0, padding: 0, margin: 0, minWidth: 0 }}
          aria-label="Choose assessment support"
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.preventDefault();
              event.stopPropagation();
              close();
            }
          }}
        >
          <CandidateList
            read={read}
            revision={revision}
            view={timelineViewSchemaId}
            label="Timeline support candidates"
            testId={assessmentCreateControlTestId("support-refs")}
            disabled={disabled}
            multiple
            selected={staged.ids}
            labels={staged.labels}
            onChange={(ids, labels) => setStaged({ ids, labels })}
          />
          {staged.ids.length > 64 && (
            <p role="alert">Choose at most 64 supporting records.</p>
          )}
          <button
            style={secondaryButtonStyle}
            type="button"
            disabled={disabled || staged.ids.length > 64}
            onClick={() => {
              update((current) => ({
                ...current,
                supportRecordIds: [...staged.ids],
                supportDisplayText: { ...staged.labels },
              }));
              close();
            }}
          >
            Apply support selection
          </button>
          <button style={secondaryButtonStyle} type="button" onClick={close}>
            Cancel support selection
          </button>
        </fieldset>
      )}
    </section>
  );
}

function CandidateList({
  read,
  revision,
  view,
  label,
  testId,
  disabled,
  multiple,
  selected,
  labels,
  onChange,
}: {
  readonly read: AssessmentCandidateReadPort["support"];
  readonly revision: number | string;
  readonly view: string;
  readonly label: string;
  readonly testId: string;
  readonly disabled: boolean;
  readonly multiple: boolean;
  readonly selected: readonly string[];
  readonly labels: Readonly<Record<string, string>>;
  readonly onChange: (
    ids: string[],
    labels: Readonly<Record<string, string>>,
  ) => void;
}) {
  const contract = requireViewContract(view);
  const [query, setQuery] = useState<WorkbookQueryState>(
    emptyWorkbookQueryState,
  );
  const page = useWorkbookCandidates(read, query, revision);
  const candidates = useMemo(() => {
    const rows = new Map<string, AssessmentSupportCandidate>();
    for (const id of selected)
      rows.set(id, { recordId: id, displayText: labels[id] ?? id });
    for (const candidate of page.candidates)
      rows.set(candidate.recordId, candidate);
    return [...rows.values()];
  }, [labels, page.candidates, selected]);
  const filterableFields = contract.fields.filter(
    (field) =>
      contract.filterFields.includes(field.fieldKey) &&
      field.filterOps.includes("eq") &&
      (field.enumValues || field.readKind === "text"),
  );
  return (
    <div style={{ display: "grid", gap: "var(--ct-spacing-xs)", minWidth: 0 }}>
      <details>
        <summary>{label} ordering and filters</summary>
        <label style={stackedLabelStyle}>
          Order
          <select
            style={inputStyle}
            aria-label={`${label} order`}
            value={
              query.sort[0]
                ? `${query.sort[0].fieldKey}:${query.sort[0].direction}`
                : ""
            }
            onChange={(event) => {
              const [fieldKey, direction] = event.target.value.split(":");
              setQuery((current) => ({
                ...current,
                sort: fieldKey
                  ? [
                      {
                        fieldKey,
                        direction: direction === "desc" ? "desc" : "asc",
                      },
                    ]
                  : [],
              }));
            }}
          >
            <option value="">Default order</option>
            {contract.sortFields.flatMap((key) =>
              ["asc", "desc"].map((direction) => (
                <option
                  key={`${key}:${direction}`}
                  value={`${key}:${direction}`}
                >
                  {contract.fieldMap[key]?.label ?? key} ·{" "}
                  {direction === "asc" ? "ascending" : "descending"}
                </option>
              )),
            )}
          </select>
        </label>
        {filterableFields.map((field) => (
          <label
            htmlFor={`${testId}-filter-${field.fieldKey}`}
            key={field.fieldKey}
            style={stackedLabelStyle}
          >
            {field.label}
            {field.enumValues ? (
              <select
                style={inputStyle}
                id={`${testId}-filter-${field.fieldKey}`}
                aria-label={`${label} filter ${field.label}`}
                value={String(
                  query.filters.find(
                    (filter) => filter.fieldKey === field.fieldKey,
                  )?.arg.value ?? "",
                )}
                onChange={(event) => {
                  const value = event.target.value;
                  setQuery((current) => ({
                    ...current,
                    filters: [
                      ...current.filters.filter(
                        (filter) => filter.fieldKey !== field.fieldKey,
                      ),
                      ...(value
                        ? [
                            {
                              fieldKey: field.fieldKey,
                              op: "eq" as const,
                              arg: { value },
                            },
                          ]
                        : []),
                    ],
                  }));
                }}
              >
                <option value="">All</option>
                {field.enumValues?.map((value) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </select>
            ) : (
              <input
                style={inputStyle}
                id={`${testId}-filter-${field.fieldKey}`}
                aria-label={`${label} filter ${field.label}`}
                type="text"
                placeholder="Equals…"
                value={String(
                  query.filters.find(
                    (filter) => filter.fieldKey === field.fieldKey,
                  )?.arg.value ?? "",
                )}
                onChange={(event) => {
                  const value = event.target.value;
                  setQuery((current) => ({
                    ...current,
                    filters: [
                      ...current.filters.filter(
                        (filter) => filter.fieldKey !== field.fieldKey,
                      ),
                      ...(value
                        ? [
                            {
                              fieldKey: field.fieldKey,
                              op: "eq" as const,
                              arg: { value },
                            },
                          ]
                        : []),
                    ],
                  }));
                }}
              />
            )}
          </label>
        ))}
      </details>
      <WorkbookRecordCandidatePicker
        candidates={candidates}
        disabled={disabled || page.stale || page.phase === "loading"}
        label={label}
        selection={multiple ? "multiple" : "single"}
        selectedRecordIds={selected}
        testId={testId}
        onSelectedRecordIdsChange={(ids) =>
          onChange(
            ids,
            Object.fromEntries(
              candidates.map((candidate) => [
                candidate.recordId,
                candidate.displayText,
              ]),
            ),
          )
        }
      />
      {page.loaded &&
      selected.some(
        (id) => !page.candidates.some((candidate) => candidate.recordId === id),
      ) ? (
        <p>
          Selected references outside these candidate pages remain selected.
          Their availability is checked when you submit.
        </p>
      ) : null}
      <p role={page.phase === "failed" ? "alert" : "status"}>
        {page.phase === "loading"
          ? page.loaded
            ? "Refreshing candidates…"
            : "Loading candidates…"
          : page.phase === "failed"
            ? page.error
            : page.candidates.length === 0
              ? query.filters.length
                ? "No candidates match these filters."
                : "No candidates found."
              : page.hasMore
                ? "More candidates are available."
                : "All matching candidates loaded."}
        {page.stale
          ? " Previously loaded candidates may be stale. Selected references are retained."
          : ""}
      </p>
      <div>
        {page.phase === "failed" && (
          <button
            type="button"
            style={secondaryButtonStyle}
            onClick={() => void page.retry()}
          >
            Retry candidates
          </button>
        )}
        <button
          type="button"
          style={secondaryButtonStyle}
          disabled={page.phase === "loading"}
          onClick={() => void page.reload()}
        >
          Reload candidates
        </button>
        {page.hasMore && (
          <button
            type="button"
            style={secondaryButtonStyle}
            disabled={page.phase !== "ready" || page.stale}
            onClick={() => void page.loadMore()}
          >
            Load more candidates
          </button>
        )}
      </div>
    </div>
  );
}
