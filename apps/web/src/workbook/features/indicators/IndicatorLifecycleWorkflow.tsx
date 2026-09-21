import { indicatorLifecycleTestId } from "@cartulary/ui-contracts";
import {
  useContext,
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import {
  type IndicatorLifecycleInterval,
  indicatorLifecycleConstraints,
} from "../../adapters/indicatorLifecycleProtocol";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import {
  inspectorReadData,
  WorkbookInspectorRegionContent,
} from "../../inspector/presentation/WorkbookInspectorPanelContent";
import type { WorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import { IndicatorLifecycleContext } from "./IndicatorLifecycleContext";
import { IndicatorLifecycleOperationStatus } from "./IndicatorLifecycleOperationStatus";
import {
  IndicatorLifecycleSupportPicker,
  LifecyclePagingFeedback,
} from "./IndicatorLifecycleSupportPicker";
import {
  type LifecycleDraftValues,
  type LifecycleFieldError,
  lifecycleTimeKey,
  validateLifecycleDraft,
} from "./indicatorLifecycleModel";
import type { IndicatorLifecycleOwnerPort } from "./indicatorLifecycleOperation";
import { IndicatorLifecyclePaging } from "./indicatorLifecyclePaging";
import {
  lifecycleField,
  lifecycleInput,
  lifecycleStack,
  lifecycleText,
} from "./indicatorLifecycleStyles";

export function IndicatorLifecycleWorkflow({
  action,
  subject,
}: {
  action: "indicator.lifecycle.read" | "indicator.lifecycle.manage";
  subject: WorkbookInspectorSubject;
}) {
  const owner = useContext(IndicatorLifecycleContext);
  if (!owner)
    return (
      <p role="status">
        Indicator lifecycle is unavailable while the workbook reconnects.
      </p>
    );
  return (
    <LifecycleContent
      key={`${subject.viewSchemaId}:${subject.recordId}`}
      owner={owner}
      action={action}
      subject={subject}
    />
  );
}

function LifecycleContent({
  owner,
  action,
  subject,
}: {
  owner: IndicatorLifecycleOwnerPort;
  action: "indicator.lifecycle.read" | "indicator.lifecycle.manage";
  subject: WorkbookInspectorSubject;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const identity = `${snapshot.authority?.actorId}:${snapshot.authority?.incidentId}:${subject.recordId}`;
  const current = useRef({
    identity,
    version: subject.rowVersion,
    mounted: true,
  });
  current.current = { identity, version: subject.rowVersion, mounted: true };
  useLayoutEffect(
    () => () => {
      current.current.mounted = false;
    },
    [],
  );
  const generation = snapshot.generation;
  const pages = useMemo(() => {
    void generation;
    return new IndicatorLifecyclePaging(
      (cursor, signal) => owner.intervals(subject.recordId, cursor, signal),
      (item) => item.interval_id,
      (previous, next) =>
        lifecycleTimeKey(previous.valid_from) >
          lifecycleTimeKey(next.valid_from) ||
        (previous.valid_from === next.valid_from &&
          previous.interval_id > next.interval_id),
    );
  }, [owner, subject.recordId, generation]);
  const collection = useSyncExternalStore(pages.subscribe, pages.getSnapshot);
  useEffect(() => {
    void pages.load();
    return () => pages.dispose();
  }, [pages]);
  const knownVersion =
    owner.latestVersion(subject.recordId) ?? subject.rowVersion;
  const observedVersion = useRef(knownVersion);
  useEffect(() => {
    if (knownVersion > observedVersion.current) void pages.refresh();
    observedVersion.current = knownVersion;
  }, [knownVersion, pages]);
  useLayoutEffect(() => {
    if (action === "indicator.lifecycle.manage" && snapshot.authority)
      owner.drafts.open(subject.recordId, subject.label, subject.rowVersion);
  }, [
    action,
    owner,
    subject.recordId,
    subject.label,
    subject.rowVersion,
    snapshot.authority,
  ]);
  const draft = owner.drafts.get(subject.recordId);
  const [errors, setErrors] = useState<readonly LifecycleFieldError[]>([]);
  const [message, setMessage] = useState<string | null>(null);
  const [reviewing, setReviewing] = useState(false);
  const id = useId();
  const busy = owner.blocksRecord(subject.recordId);
  const stale = draft !== null && knownVersion > draft.baseRowVersion;
  const update = (values: LifecycleDraftValues) => {
    owner.drafts.update(subject.recordId, values);
    setErrors([]);
    setMessage(null);
  };
  const submit = () => {
    if (!draft) return;
    const validated = validateLifecycleDraft(draft.values);
    setErrors(validated.errors);
    if (!validated.values) return;
    const capturedIdentity = identity,
      version = draft.baseRowVersion;
    const attempt = owner.admit(draft, {
      isCurrent: () =>
        current.current.mounted &&
        current.current.identity === capturedIdentity,
      matchesDraft: () => current.current.version === version,
      reconcile: async () => {
        if (!(await pages.refresh()))
          throw new Error("Lifecycle collection refresh required");
      },
    });
    if (!attempt) {
      setMessage(
        "The interval could not be submitted. Check access, secure browser support, and the current Indicator version. Your draft is retained.",
      );
      return;
    }
    setMessage(null);
    void owner.execute(attempt);
  };
  if (!snapshot.authority) return null;
  return (
    <section
      data-testid={indicatorLifecycleTestId("intervals")}
      aria-label="Indicator lifecycle intervals"
      style={lifecycleStack}
    >
      <h3>Lifecycle intervals</h3>
      <p style={lifecycleText}>
        Effective times are UTC. Appending retains every earlier interval,
        including overlapping intervals. Reverse an interval through record
        History.
      </p>
      {action === "indicator.lifecycle.manage" && draft ? (
        <form
          data-testid={indicatorLifecycleTestId("editor")}
          noValidate
          style={lifecycleStack}
          onSubmit={(event) => {
            event.preventDefault();
            submit();
          }}
        >
          <fieldset
            disabled={busy || !owner.canSubmit() || reviewing}
            style={lifecycleStack}
          >
            <legend>Append an interval for {subject.label}</legend>
            <label style={lifecycleField}>
              State
              <select
                style={lifecycleInput}
                value={draft.values.state}
                onChange={(event) =>
                  update({ ...draft.values, state: event.target.value })
                }
              >
                {indicatorLifecycleConstraints.states.map((state) => (
                  <option key={state} value={state}>
                    {state === "false_positive"
                      ? "False positive"
                      : state.charAt(0).toUpperCase() + state.slice(1)}
                  </option>
                ))}
              </select>
            </label>
            {(
              [
                ["validFrom", "Effective from (UTC)", "datetime-local"],
                ["validTo", "Effective to (UTC, optional)", "datetime-local"],
                ["confidence", "Confidence (optional)", "text"],
                ["rationale", "Rationale (optional)", "textarea"],
                [
                  "assessor",
                  "Assessor text (optional, analyst entered)",
                  "text",
                ],
              ] as const
            ).map(([field, label, type]) => {
              const error = errors.find((error) => error.field === field),
                errorId = `${id}-${field}`;
              return (
                <div key={field} style={lifecycleField}>
                  <label
                    style={lifecycleField}
                    htmlFor={`${id}-${field}-input`}
                  >
                    {label}
                    {type === "textarea" ? (
                      <textarea
                        id={`${id}-${field}-input`}
                        style={lifecycleInput}
                        value={draft.values[field]}
                        aria-invalid={!!error}
                        aria-describedby={error ? errorId : undefined}
                        onChange={(event) =>
                          update({
                            ...draft.values,
                            [field]: event.target.value,
                          })
                        }
                      />
                    ) : (
                      <input
                        id={`${id}-${field}-input`}
                        style={lifecycleInput}
                        type={type}
                        step={type === "datetime-local" ? "any" : undefined}
                        value={draft.values[field]}
                        aria-invalid={!!error}
                        aria-describedby={error ? errorId : undefined}
                        onChange={(event) =>
                          update({
                            ...draft.values,
                            [field]: event.target.value,
                          })
                        }
                      />
                    )}
                  </label>
                  {error ? (
                    <p id={errorId} role="alert" style={lifecycleText}>
                      {error.message}
                    </p>
                  ) : null}
                </div>
              );
            })}
            <IndicatorLifecycleSupportPicker
              owner={owner}
              disabled={busy || !owner.canSubmit()}
              selected={draft.values.support}
              onChange={(support) => update({ ...draft.values, support })}
            />
            {errors
              .filter(
                (error) => error.field === "support" || error.field === "state",
              )
              .map((error) => (
                <p key={error.field} role="alert">
                  {error.message}
                </p>
              ))}
          </fieldset>
          {!owner.canSubmit() ? (
            <p>
              Appending requires current editor access to an open incident. Your
              draft is retained.
            </p>
          ) : null}
          {stale ? (
            <p role="status">
              The Indicator changed. Refresh and review this retained draft
              before appending.
            </p>
          ) : null}
          <WorkbookInspectorActionButton
            disabled={busy || reviewing || !owner.canSubmit()}
            onClick={() => {
              if (reviewing) return;
              setReviewing(true);
              void owner
                .review(subject.recordId)
                .then((ok) => {
                  if (
                    current.current.mounted &&
                    current.current.identity === identity
                  )
                    setMessage(
                      ok
                        ? "Current Indicator loaded. Review your retained values, then append."
                        : "Current Indicator could not be verified. Your draft is retained.",
                    );
                })
                .finally(() => {
                  if (current.current.mounted) setReviewing(false);
                });
            }}
          >
            Refresh and review Indicator
          </WorkbookInspectorActionButton>
          <WorkbookInspectorActionButton
            type="submit"
            disabled={busy || stale || reviewing || !owner.canSubmit()}
          >
            Append lifecycle interval
          </WorkbookInspectorActionButton>
          <WorkbookInspectorActionButton
            disabled={busy}
            onClick={() => {
              owner.drafts.discard(subject.recordId);
              owner.drafts.open(subject.recordId, subject.label, knownVersion);
              setErrors([]);
              setMessage(null);
            }}
          >
            Discard interval draft
          </WorkbookInspectorActionButton>
        </form>
      ) : null}
      {message ? <p role="status">{message}</p> : null}
      {snapshot.entries
        .filter((entry) => entry.attempt.draft.recordId === subject.recordId)
        .map((entry) => (
          <IndicatorLifecycleOperationStatus
            key={entry.attempt.id}
            owner={owner}
            entry={entry}
          />
        ))}
      <div data-inspector-region="indicator-lifecycle">
        <WorkbookInspectorRegionContent
          model={{
            access: "readable",
            data: inspectorReadData({
              requested: collection.request > 0,
              pending: [
                "initial_loading",
                "loading_more",
                "refreshing",
              ].includes(collection.phase),
              accepted: collection.hasAccepted
                ? collection.items.length
                  ? {
                      kind: "populated",
                      content: collection.items.map((interval) => (
                        <LifecycleIntervalDetails
                          key={interval.interval_id}
                          interval={interval}
                        />
                      )),
                    }
                  : {
                      kind: "empty",
                      message: collection.hasMore
                        ? "No lifecycle intervals in the loaded pages; more pages are available."
                        : "No lifecycle intervals.",
                    }
                : null,
              failure: collection.failure?.message ?? null,
              notLoadedMessage: "Lifecycle intervals have not been loaded.",
            }),
            commands: (
              <LifecyclePagingFeedback
                controlsOnly
                pages={pages}
                state={collection}
                label="lifecycle intervals"
              />
            ),
          }}
        />
      </div>
    </section>
  );
}

export function LifecycleIntervalDetails({
  interval,
}: {
  interval: IndicatorLifecycleInterval;
}) {
  return (
    <article
      aria-label={`${interval.lifecycle_state.replaceAll("_", " ")} interval from ${interval.valid_from}`}
      style={{
        ...lifecycleStack,
        borderBlockStart: "var(--ct-border-hairline)",
        paddingBlock: "var(--ct-spacing-sm)",
      }}
    >
      <strong>{interval.lifecycle_state.replaceAll("_", " ")}</strong>
      <dl style={{ margin: 0 }}>
        <dt>Effective from (UTC)</dt>
        <dd>{interval.valid_from}</dd>
        <dt>Effective to (UTC)</dt>
        <dd>{interval.valid_to ?? "No end specified"}</dd>
        <dt>Confidence</dt>
        <dd>{interval.confidence ?? "Not supplied"}</dd>
        <dt>Rationale</dt>
        <dd style={{ whiteSpace: "pre-wrap" }}>
          {interval.rationale ?? "Not supplied"}
        </dd>
        <dt>Supporting records</dt>
        <dd>
          {interval.support_refs.length
            ? interval.support_refs.join(", ")
            : "None"}
        </dd>
        <dt>Assessor text (analyst entered)</dt>
        <dd style={{ whiteSpace: "pre-wrap" }}>
          {interval.assessor ?? "Not supplied"}
        </dd>
        <dt>Recorded by (server)</dt>
        <dd>{interval.created_by_user_id}</dd>
        <dt>Assessed at (server, UTC)</dt>
        <dd>{interval.assessed_at}</dd>
        <dt>Created at (server, UTC)</dt>
        <dd>{interval.created_at}</dd>
      </dl>
    </article>
  );
}
