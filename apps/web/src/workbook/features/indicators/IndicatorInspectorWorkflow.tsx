import { indicatorObservationTestId } from "@cartulary/ui-contracts";
import {
  useContext,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import type { IndicatorInspectorAction } from "./indicatorInspectorHandlers";
import { ObservationCaptureEditor } from "./ObservationCaptureEditor";
import { ObservationCollection } from "./ObservationCollection";
import { ObservationContext } from "./ObservationContext";
import { ObservationDetails } from "./ObservationDetails";
import { ObservationOperationStatus } from "./ObservationOperationStatus";
import { ObservationPagingFeedback } from "./ObservationPagingFeedback";
import {
  type ObservationDraft,
  observationOrder,
  sameObservationSource,
} from "./observationModel";
import {
  type ObservationIntent,
  type ObservationOwnerPort,
  type ObservationSourcePort,
  type ObservationSubject,
  observationIntentSource,
} from "./observationOperation";
import {
  observationField,
  observationInput,
  observationStack,
} from "./observationStyles";
import { useObservationTargetNames } from "./useObservationTargetNames";

type Props = {
  action: IndicatorInspectorAction | null;
  indicatorRecordId?: string | undefined;
  sourceRecordId?: string | undefined;
  source?: ObservationSourcePort | undefined;
  onMutationCommitted?: (() => Promise<void> | void) | undefined;
};
export function IndicatorInspectorWorkflow(props: Props) {
  const owner = useContext(ObservationContext);
  if (!owner) return null;
  return <ObservationWorkflowOwner {...props} owner={owner} />;
}
function ObservationWorkflowOwner(
  props: Props & { owner: ObservationOwnerPort },
) {
  const snapshot = useSyncExternalStore(
    props.owner.subscribe,
    props.owner.getSnapshot,
  );
  const subject: ObservationSubject | null =
    props.action === "indicator.observations.manage" && props.sourceRecordId
      ? { kind: "source", recordId: props.sourceRecordId }
      : props.action === "indicator.observations.pivot" &&
          props.indicatorRecordId
        ? { kind: "indicator", recordId: props.indicatorRecordId }
        : null;
  if (!snapshot.authority || !subject) return null;
  return (
    <ObservationWorkflow
      key={`${snapshot.generation}:${subject.kind}:${subject.recordId}`}
      {...props}
      subject={subject}
      generation={snapshot.generation}
    />
  );
}
const noSourceSubscription = () => () => {};
const sourceNotReady = () => false;
function ObservationWorkflow({
  owner,
  subject,
  generation,
  ...props
}: Props & {
  owner: ObservationOwnerPort;
  subject: ObservationSubject;
  generation: number;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const sourceReady = useSyncExternalStore(
    props.source?.subscribe ?? noSourceSubscription,
    props.source?.ready ?? sourceNotReady,
  );
  const { kind, recordId } = subject;
  const sourceFieldId = useId();
  const collection = useMemo(
    () =>
      new ObservationCollection<IndicatorObservation>(
        (cursor, signal) =>
          owner.observations({ kind, recordId }, cursor, signal),
        (item) => item.observation_id,
        (item) => item.row_version,
        observationOrder,
      ),
    [owner, kind, recordId],
  );
  const state = useSyncExternalStore(
    collection.subscribe,
    collection.getSnapshot,
  );
  const targets = useObservationTargetNames(
    owner,
    generation,
    state.items.flatMap((item) =>
      item.resolved_indicator_record_id
        ? [item.resolved_indicator_record_id]
        : [],
    ),
  );
  const current = useRef(props);
  current.current = props;
  const alive = useRef(true);
  const [field, setField] = useState(props.source?.fields[0]?.fieldKey ?? "");
  const [error, setError] = useState<string | null>(null);
  useEffect(() => {
    alive.current = true;
    void collection.load();
    return () => {
      alive.current = false;
      collection.dispose();
    };
  }, [collection]);
  const entries = snapshot.entries.filter((entry) =>
    subject.kind === "source"
      ? observationIntentSource(entry.attempt.intent) === subject.recordId
      : entry.receipt?.affected_records.some(
          (record) => record.record_id === subject.recordId,
        ),
  );
  const stamp = entries
    .filter((entry) => entry.receipt)
    .map((entry) => entry.attempt.id)
    .join(":");
  const priorStamp = useRef(stamp);
  useEffect(() => {
    if (stamp !== priorStamp.current) {
      priorStamp.current = stamp;
      void collection.refresh();
    }
  }, [stamp, collection]);
  const isCurrent = () =>
    alive.current && owner.getSnapshot().generation === generation;
  function submit(intent: ObservationIntent, draft: ObservationDraft) {
    const matches = () =>
      owner.drafts.get(draft.key) === draft &&
      (intent.action === "create"
        ? current.current.source?.ready() === true &&
          sameObservationSource(
            current.current.source.source(intent.source.fieldKey),
            intent.source,
          )
        : collection
            .getSnapshot()
            .items.some(
              (item) =>
                item.observation_id === intent.observation.observation_id &&
                item.row_version === intent.observation.row_version,
            ));
    const attempt = owner.admit(intent, {
      isCurrent,
      matchesDraft: matches,
      prepare: async (signal) =>
        intent.action !== "create" ||
        (await current.current.source?.prepare(intent.source, signal)) === true,
      reconcile: async () => {
        if (!isCurrent()) return;
        if (!(await collection.refresh()))
          throw new Error("Observation collection needs refresh");
        if (!isCurrent()) return;
        await current.current.onMutationCommitted?.();
        if (isCurrent() && owner.drafts.get(draft.key) === draft)
          owner.drafts.update(draft.key, { source: null, selection: null });
      },
    });
    setError(
      attempt
        ? null
        : "The operation is unavailable or changed. Review the current source, observation and target.",
    );
    if (attempt) void owner.execute(attempt);
    return attempt;
  }
  const manage =
    subject.kind === "source" &&
    props.action === "indicator.observations.manage";
  const source = props.source?.source(field),
    ready = sourceReady;
  const createBusy = entries.some(
    (entry) =>
      entry.attempt.intent.action === "create" &&
      (entry.transportPending ||
        ["preparing", "submitting", "uncertain"].includes(entry.phase)),
  );
  return (
    <section
      data-testid={indicatorObservationTestId("editor")}
      aria-label="Indicator observations"
      style={observationStack}
    >
      <h3 tabIndex={-1}>Indicator observations</h3>
      {manage ? (
        <>
          <div style={observationField}>
            <label htmlFor={sourceFieldId}>Source field</label>
            <select
              id={sourceFieldId}
              style={observationInput}
              value={field}
              disabled={createBusy}
              onChange={(event) => setField(event.target.value)}
            >
              {props.source?.fields.map((item) => (
                <option key={item.fieldKey} value={item.fieldKey}>
                  {item.label}
                </option>
              ))}
            </select>
          </div>
          {source ? (
            <ObservationCaptureEditor
              key={field}
              source={source}
              ready={ready}
              draft={owner.drafts.ensure(
                `capture:${subject.recordId}:${field}`,
              )}
              drafts={owner.drafts}
              reader={owner}
              generation={generation}
              disabled={!owner.canSubmit() || createBusy}
              onSubmit={submit}
            />
          ) : (
            <p>No saved source text is available for capture.</p>
          )}
        </>
      ) : (
        <p>
          Observations currently resolved to this Indicator. Manage an
          observation from its source record’s Relationships panel.
        </p>
      )}
      {error ? <p role="alert">{error}</p> : null}
      {entries.map((entry) => (
        <ObservationOperationStatus
          key={entry.attempt.id}
          owner={owner}
          entry={entry}
        />
      ))}
      {state.phase === "ready" && state.items.length === 0 ? (
        <p>No observations.</p>
      ) : null}
      {state.items.map((item) => (
        <ObservationDetails
          key={item.observation_id}
          item={item}
          targetLabel={
            item.resolved_indicator_record_id
              ? targets.labels.get(item.resolved_indicator_record_id)
              : undefined
          }
          reader={owner}
          generation={generation}
          draft={owner.drafts.ensure(`resolve:${item.observation_id}`)}
          drafts={owner.drafts}
          manage={manage}
          disabled={
            !owner.canSubmit() ||
            owner.busy({ action: "dismiss", observation: item })
          }
          onSubmit={submit}
        />
      ))}
      {targets.state.failure ? (
        <ObservationPagingFeedback
          pages={targets.pages}
          state={targets.state}
          label="resolution details"
        />
      ) : null}
      <ObservationPagingFeedback
        pages={collection}
        state={state}
        label="observations"
      />
    </section>
  );
}
