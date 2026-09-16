import { assessmentCreateControlTestId } from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import {
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { WorkbookCandidateBrowsing } from "../../components/WorkbookCandidateBrowsing";
import { WorkbookCandidateQueryControl } from "../../components/WorkbookCandidateQueryControl";
import { WorkbookCandidateSelection } from "../../components/WorkbookCandidateSelection";
import { secondaryButtonStyle } from "../../components/workbookGridControlStyles";
import {
  useWorkbookCandidateDiscovery,
  WorkbookCandidateAuthorityContext,
} from "../../hooks/useWorkbookCandidateDiscovery";
import type { AssessmentCreateDraft } from "../../models/assessmentWorkbookModel";
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
  const authority = useContext(WorkbookCandidateAuthorityContext);
  const [concealed, setConcealed] = useState(false);
  const freshness = `${authority.identity}:${revision}`;
  const priorFreshness = useRef(freshness);
  if (priorFreshness.current !== freshness) {
    priorFreshness.current = freshness;
    if (staged) setStaged({ ...staged, labels: {} });
  }
  useEffect(() => {
    if (disabled) setStaged(null);
  }, [disabled]);
  const trigger = useRef<HTMLButtonElement>(null);
  const read = useCallback(
    (input: AssessmentCandidateQuery) => reader.support(input),
    [reader],
  );
  function close() {
    setStaged(null);
    trigger.current?.focus({ preventScroll: true });
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
              {concealed || !authority.canRead
                ? "Selected reference"
                : (draft.supportDisplayText?.[id] ?? id)}
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
      {staged && !disabled && (
        <fieldset
          style={{ border: 0, padding: 0, margin: 0, minWidth: 0 }}
          aria-label="Choose assessment support"
          onKeyDown={(event) => {
            event.stopPropagation();
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
            onConcealedChange={setConcealed}
            onChange={(ids, labels) => setStaged({ ids, labels })}
          />
          {staged.ids.length > 64 && (
            <p role="alert">Choose at most 64 supporting records.</p>
          )}
          <button
            style={secondaryButtonStyle}
            type="button"
            disabled={
              disabled ||
              concealed ||
              !authority.canRead ||
              staged.ids.length > 64
            }
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
  onConcealedChange,
}: {
  readonly onConcealedChange?: (concealed: boolean) => void;
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
  const [query, setQuery] = useState<WorkbookQueryState>(
    emptyWorkbookQueryState,
  );
  const page = useWorkbookCandidateDiscovery(
    read,
    query,
    `assessment:${view}:${multiple ? "support" : "subject"}`,
    revision,
    !disabled,
  );
  useLayoutEffect(() => {
    onConcealedChange?.(page.concealed || !page.canRead);
  }, [onConcealedChange, page.concealed, page.canRead]);
  return (
    <div style={{ display: "grid", gap: "var(--ct-spacing-xs)", minWidth: 0 }}>
      <WorkbookCandidateQueryControl
        key={view}
        view={view}
        label={label}
        query={query}
        onApply={setQuery}
      />
      <WorkbookCandidateSelection
        candidates={page.page?.candidates ?? []}
        selected={selected.map((recordId) => ({
          recordId,
          displayText: labels[recordId] ?? recordId,
        }))}
        label={label}
        testId={testId}
        multiple={multiple}
        maximum={multiple ? 64 : 1}
        disabled={disabled}
        concealed={page.concealed || !page.canRead}
        onChange={(items) =>
          onChange(
            items.map((item) => item.recordId),
            Object.fromEntries(
              items.map((item) => [item.recordId, item.displayText]),
            ),
          )
        }
      />
      <WorkbookCandidateBrowsing discovery={page} />
    </div>
  );
}
