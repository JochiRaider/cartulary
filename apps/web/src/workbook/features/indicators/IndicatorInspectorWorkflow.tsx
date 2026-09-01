import type { CSSProperties, FormEvent } from "react";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  WorkbookInspectorFeedbackView,
  WorkbookInspectorPublicError,
} from "../../inspector/presentation/WorkbookInspectorFeedback";
import {
  type WorkbookInspectorErrorPresentation,
  type WorkbookInspectorFeedback,
  workbookInspectorErrorPresentation,
  workbookInspectorMessageFeedback,
  workbookInspectorOperationFailureFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type {
  IndicatorLifecycleState,
  IndicatorMutationAccepted,
  IndicatorObservation,
  IndicatorPage,
  IndicatorPaging,
  IndicatorStateInterval,
  IndicatorWorkflowPort,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type { IndicatorInspectorAction } from "./indicatorInspectorHandlers";

const lifecycleStates = [
  "active",
  "benign",
  "false_positive",
  "retired",
] as const satisfies readonly IndicatorLifecycleState[];
const indicatorTypes = [
  "ipv4_addr",
  "ipv6_addr",
  "domain_name",
  "url",
  "sha256",
  "email_addr",
  "registry_key",
  "process_name",
  "text",
] as const;

type IndicatorParsedType = Exclude<
  IndicatorObservation["parsed_indicator_type"],
  null
>;
type IndicatorCommittedMutation =
  | IndicatorMutationAccepted<IndicatorObservation>
  | IndicatorMutationAccepted<IndicatorStateInterval>;

export function IndicatorInspectorWorkflow({
  action,
  indicatorRecordId,
  onMutationCommitted,
  port,
  rowVersion,
  sourceFields = [],
  sourceRecordId,
}: {
  readonly action: IndicatorInspectorAction | null;
  readonly indicatorRecordId?: string | undefined;
  readonly onMutationCommitted?:
    | ((mutation: IndicatorCommittedMutation) => Promise<void> | void)
    | undefined;
  readonly port: IndicatorWorkflowPort;
  readonly rowVersion: number;
  readonly sourceFields?:
    | readonly {
        readonly fieldKey: string;
        readonly label: string;
        readonly value?: string | undefined;
      }[]
    | undefined;
  readonly sourceRecordId?: string | undefined;
}) {
  const [observations, setObservations] = useState<
    readonly IndicatorObservation[]
  >([]);
  const [intervals, setIntervals] = useState<readonly IndicatorStateInterval[]>(
    [],
  );
  const [message, setMessage] = useState<WorkbookInspectorFeedback | null>(
    null,
  );
  const [loadError, setLoadError] =
    useState<WorkbookInspectorErrorPresentation | null>(null);
  const [paging, setPaging] = useState<IndicatorPaging | null>(null);
  const [busy, setBusy] = useState(false);
  const loadRequestId = useRef(0);
  const [sourceFieldKey, setSourceFieldKey] = useState(
    sourceFields[0]?.fieldKey ?? "",
  );
  const [spanStart, setSpanStart] = useState("0");
  const [spanEnd, setSpanEnd] = useState("1");
  const [parsedType, setParsedType] = useState<"" | IndicatorParsedType>("");
  const [resolvedIndicatorID, setResolvedIndicatorID] = useState("");
  const [lifecycleState, setLifecycleState] =
    useState<IndicatorLifecycleState>("active");
  const [validFrom, setValidFrom] = useState("");
  const [validTo, setValidTo] = useState("");
  const [confidence, setConfidence] = useState("");
  const [rationale, setRationale] = useState("");
  const [assessor, setAssessor] = useState("");

  useEffect(() => {
    setSourceFieldKey((current) => current || sourceFields[0]?.fieldKey || "");
  }, [sourceFields]);

  const loadResources = useCallback(
    async (cursorToken?: string) => {
      const requestId = ++loadRequestId.current;
      setBusy(true);
      setLoadError(null);
      let outcome: WorkbookOperationOutcome<
        IndicatorPage<IndicatorObservation>
      >;
      if (action === "indicator.observations.pivot" && indicatorRecordId) {
        outcome = await port.listObservations({
          cursorToken,
          indicatorRecordId,
        });
      } else if (action === "indicator.observations.manage" && sourceRecordId) {
        outcome = await port.listSourceObservations({
          cursorToken,
          sourceRecordId,
        });
      } else if (
        (action === "indicator.lifecycle.read" ||
          action === "indicator.lifecycle.manage") &&
        indicatorRecordId
      ) {
        const lifecycleOutcome = await port.listStateIntervals({
          cursorToken,
          indicatorRecordId,
        });
        if (requestId !== loadRequestId.current) return;
        setBusy(false);
        if (lifecycleOutcome.kind === "rejected") {
          setLoadError(
            workbookInspectorErrorPresentation(lifecycleOutcome.failure),
          );
          return;
        }
        setIntervals((current) =>
          cursorToken === undefined
            ? lifecycleOutcome.value.items
            : [...current, ...lifecycleOutcome.value.items],
        );
        setPaging(lifecycleOutcome.value.paging);
        return;
      } else {
        setBusy(false);
        return;
      }
      if (requestId !== loadRequestId.current) return;
      setBusy(false);
      if (outcome.kind === "rejected") {
        setLoadError(workbookInspectorErrorPresentation(outcome.failure));
        return;
      }
      setObservations((current) =>
        cursorToken === undefined
          ? outcome.value.items
          : [...current, ...outcome.value.items],
      );
      setPaging(outcome.value.paging);
    },
    [action, indicatorRecordId, port, sourceRecordId],
  );

  useEffect(() => {
    setMessage(null);
    setLoadError(null);
    setPaging(null);
    setObservations([]);
    setIntervals([]);
    if (
      (action === "indicator.observations.pivot" ||
        action === "indicator.lifecycle.read" ||
        action === "indicator.lifecycle.manage") &&
      !indicatorRecordId
    ) {
      setMessage(
        workbookInspectorMessageFeedback(
          "The selected row is not an Indicator record.",
          "polite",
        ),
      );
      return () => {
        loadRequestId.current += 1;
      };
    }
    void loadResources();
    return () => {
      loadRequestId.current += 1;
    };
  }, [action, indicatorRecordId, loadResources]);

  if (action === null) return null;

  const submitObservation = async (event: FormEvent) => {
    event.preventDefault();
    if (!sourceRecordId || sourceFieldKey === "") {
      setMessage(
        workbookInspectorMessageFeedback("Select a source field.", "polite"),
      );
      return;
    }
    const start = Number(spanStart);
    const end = Number(spanEnd);
    if (
      !Number.isInteger(start) ||
      !Number.isInteger(end) ||
      start < 0 ||
      end <= start
    ) {
      setMessage(
        workbookInspectorMessageFeedback(
          "Enter a valid UTF-8 byte span.",
          "polite",
        ),
      );
      return;
    }
    setBusy(true);
    setMessage(null);
    const outcome = await port.createManualObservation({
      baseRowVersion: rowVersion,
      sourceFieldKey,
      sourceRecordId,
      spanStartByte: start,
      spanEndByte: end,
      ...(parsedType ? { parsedIndicatorType: parsedType } : {}),
      ...(resolvedIndicatorID
        ? { resolvedIndicatorRecordId: resolvedIndicatorID.trim() }
        : {}),
    });
    setBusy(false);
    if (outcome.kind === "rejected") {
      setMessage(workbookInspectorOperationFailureFeedback(outcome.failure));
      return;
    }
    const observation = outcome.value.resource;
    setMessage(
      workbookInspectorMessageFeedback(
        `Observation ${observation.observation_id} created.`,
        "polite",
      ),
    );
    setObservations((current) => [observation, ...current]);
    await onMutationCommitted?.(outcome.value);
  };

  const transitionObservation = async (
    observation: IndicatorObservation,
    transition: "dismiss" | "resolve" | "restore",
  ) => {
    const target = resolvedIndicatorID.trim();
    if (transition === "resolve" && target === "") {
      setMessage(
        workbookInspectorMessageFeedback(
          "Enter the Indicator ID to resolve this observation.",
          "polite",
        ),
      );
      return;
    }
    setBusy(true);
    setMessage(null);
    const outcome = await port.transitionObservation({
      action: transition,
      baseRowVersion: observation.row_version,
      observationId: observation.observation_id,
      ...(transition === "resolve"
        ? { resolvedIndicatorRecordId: target }
        : {}),
    });
    setBusy(false);
    if (outcome.kind === "rejected") {
      setMessage(workbookInspectorOperationFailureFeedback(outcome.failure));
      return;
    }
    setObservations((current) =>
      current.map((candidate) =>
        candidate.observation_id === outcome.value.resource.observation_id
          ? outcome.value.resource
          : candidate,
      ),
    );
    setMessage(
      workbookInspectorMessageFeedback(
        `Observation ${outcome.value.resource.observation_id} updated.`,
        "polite",
      ),
    );
    await onMutationCommitted?.(outcome.value);
  };

  const submitLifecycle = async (event: FormEvent) => {
    event.preventDefault();
    if (!indicatorRecordId || validFrom.trim() === "") {
      setMessage(
        workbookInspectorMessageFeedback(
          "Enter a valid-from timestamp.",
          "polite",
        ),
      );
      return;
    }
    const parsedConfidence =
      confidence.trim() === "" ? null : Number(confidence);
    if (
      parsedConfidence !== null &&
      (!Number.isInteger(parsedConfidence) ||
        parsedConfidence < 0 ||
        parsedConfidence > 100)
    ) {
      setMessage(
        workbookInspectorMessageFeedback(
          "Confidence must be an integer from 0 through 100.",
          "polite",
        ),
      );
      return;
    }
    setBusy(true);
    setMessage(null);
    const outcome = await port.appendStateInterval({
      assessor: assessor.trim() || null,
      baseRowVersion: rowVersion,
      confidence: parsedConfidence,
      indicatorRecordId,
      lifecycleState,
      rationale: rationale.trim() || null,
      supportRefs: [],
      validFrom: new Date(validFrom).toISOString(),
      validTo: validTo.trim() ? new Date(validTo).toISOString() : null,
    });
    setBusy(false);
    if (outcome.kind === "rejected") {
      setMessage(workbookInspectorOperationFailureFeedback(outcome.failure));
      return;
    }
    setIntervals((current) => [outcome.value.resource, ...current]);
    setMessage(
      workbookInspectorMessageFeedback(
        `Lifecycle interval ${outcome.value.resource.interval_id} created.`,
        "polite",
      ),
    );
    await onMutationCommitted?.(outcome.value);
  };

  return (
    <div style={shellStyle}>
      {action === "indicator.observations.manage" ? (
        <form
          style={formStyle}
          onSubmit={(event) => void submitObservation(event)}
        >
          <label style={labelStyle}>
            Source field
            <select
              disabled={busy}
              value={sourceFieldKey}
              onChange={(event) => setSourceFieldKey(event.target.value)}
            >
              {sourceFields.map((field) => (
                <option key={field.fieldKey} value={field.fieldKey}>
                  {field.label}
                </option>
              ))}
            </select>
          </label>
          <label style={labelStyle}>
            Span start byte
            <input
              min={0}
              step={1}
              type="number"
              value={spanStart}
              onChange={(event) => setSpanStart(event.target.value)}
            />
          </label>
          <label style={labelStyle}>
            Span end byte
            <input
              min={1}
              step={1}
              type="number"
              value={spanEnd}
              onChange={(event) => setSpanEnd(event.target.value)}
            />
          </label>
          <label style={labelStyle}>
            Parsed type (optional)
            <select
              value={parsedType}
              onChange={(event) =>
                setParsedType(event.target.value as "" | IndicatorParsedType)
              }
            >
              <option value="">Infer from selected text</option>
              {indicatorTypes.map((indicatorType) => (
                <option key={indicatorType} value={indicatorType}>
                  {indicatorType}
                </option>
              ))}
            </select>
          </label>
          <label style={labelStyle}>
            Resolve to Indicator ID (optional)
            <input
              value={resolvedIndicatorID}
              onChange={(event) => setResolvedIndicatorID(event.target.value)}
            />
          </label>
          <button disabled={busy} type="submit">
            Create observation
          </button>
        </form>
      ) : null}

      {action === "indicator.observations.manage" ? (
        <ObservationActions
          busy={busy}
          observations={observations}
          onTransition={transitionObservation}
        />
      ) : null}

      {action === "indicator.lifecycle.manage" ? (
        <form
          style={formStyle}
          onSubmit={(event) => void submitLifecycle(event)}
        >
          <label style={labelStyle}>
            State
            <select
              value={lifecycleState}
              onChange={(event) =>
                setLifecycleState(event.target.value as IndicatorLifecycleState)
              }
            >
              {lifecycleStates.map((state) => (
                <option key={state} value={state}>
                  {state}
                </option>
              ))}
            </select>
          </label>
          <label style={labelStyle}>
            Valid from
            <input
              required
              type="datetime-local"
              value={validFrom}
              onChange={(event) => setValidFrom(event.target.value)}
            />
          </label>
          <label style={labelStyle}>
            Valid to (optional)
            <input
              type="datetime-local"
              value={validTo}
              onChange={(event) => setValidTo(event.target.value)}
            />
          </label>
          <label style={labelStyle}>
            Confidence
            <input
              max={100}
              min={0}
              step={1}
              type="number"
              value={confidence}
              onChange={(event) => setConfidence(event.target.value)}
            />
          </label>
          <label style={labelStyle}>
            Rationale
            <textarea
              value={rationale}
              onChange={(event) => setRationale(event.target.value)}
            />
          </label>
          <label style={labelStyle}>
            Assessor
            <input
              value={assessor}
              onChange={(event) => setAssessor(event.target.value)}
            />
          </label>
          <button disabled={busy} type="submit">
            Append lifecycle interval
          </button>
        </form>
      ) : null}

      {action === "indicator.observations.pivot" ? (
        <ResourceList
          empty="No active observations resolve to this Indicator."
          items={observations.map((observation) => ({
            id: observation.observation_id,
            primary: observation.observed_text,
            secondary: `${observation.resolution_status} · ${observation.source_field_key}`,
          }))}
        />
      ) : null}
      {action === "indicator.lifecycle.read" ||
      action === "indicator.lifecycle.manage" ? (
        <ResourceList
          empty="No active lifecycle intervals."
          items={intervals.map((interval) => ({
            id: interval.interval_id,
            primary: interval.lifecycle_state,
            secondary: interval.valid_to
              ? `${interval.valid_from} – ${interval.valid_to}`
              : `from ${interval.valid_from}`,
          }))}
        />
      ) : null}
      {paging?.has_more && paging.next_cursor !== null ? (
        <button
          disabled={busy}
          type="button"
          onClick={() => void loadResources(paging.next_cursor ?? undefined)}
        >
          Load more
        </button>
      ) : null}
      {busy ? (
        <p aria-live="polite" role="status" style={messageStyle}>
          Loading…
        </p>
      ) : null}
      {loadError ? (
        <div style={shellStyle}>
          <WorkbookInspectorPublicError error={loadError} />
          <button
            disabled={busy}
            type="button"
            onClick={() => void loadResources()}
          >
            Retry
          </button>
        </div>
      ) : null}
      <WorkbookInspectorFeedbackView
        feedback={message}
        neutralStyle={messageStyle}
      />
    </div>
  );
}

function ResourceList({
  empty,
  items,
}: {
  readonly empty: string;
  readonly items: readonly {
    readonly id: string;
    readonly primary: string;
    readonly secondary: string;
  }[];
}) {
  if (items.length === 0) return <p style={messageStyle}>{empty}</p>;
  return (
    <ul style={listStyle}>
      {items.map((item) => (
        <li key={item.id}>
          <strong>{item.primary}</strong>
          <br />
          <span>{item.secondary}</span>
        </li>
      ))}
    </ul>
  );
}

function ObservationActions({
  busy,
  observations,
  onTransition,
}: {
  readonly busy: boolean;
  readonly observations: readonly IndicatorObservation[];
  readonly onTransition: (
    observation: IndicatorObservation,
    action: "dismiss" | "resolve" | "restore",
  ) => Promise<void>;
}) {
  if (observations.length === 0) {
    return <p style={messageStyle}>No active observations for this source.</p>;
  }
  return (
    <ul style={listStyle}>
      {observations.map((observation) => (
        <li key={observation.observation_id}>
          <strong>{observation.observed_text}</strong>
          <br />
          <span>{observation.resolution_status}</span>{" "}
          {observation.resolution_status !== "dismissed" ? (
            <>
              <button
                disabled={busy}
                type="button"
                onClick={() => void onTransition(observation, "resolve")}
              >
                Resolve
              </button>{" "}
              <button
                disabled={busy}
                type="button"
                onClick={() => void onTransition(observation, "dismiss")}
              >
                Dismiss
              </button>
            </>
          ) : (
            <button
              disabled={busy}
              type="button"
              onClick={() => void onTransition(observation, "restore")}
            >
              Restore
            </button>
          )}
        </li>
      ))}
    </ul>
  );
}

const shellStyle = { display: "grid", gap: "0.75rem" } satisfies CSSProperties;
const formStyle = { display: "grid", gap: "0.6rem" } satisfies CSSProperties;
const labelStyle = { display: "grid", gap: "0.25rem" } satisfies CSSProperties;
const listStyle = {
  margin: 0,
  paddingInlineStart: "1.25rem",
} satisfies CSSProperties;
const messageStyle = { margin: 0 } satisfies CSSProperties;
