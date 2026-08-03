import type { CSSProperties, FormEvent } from "react";
import { useEffect, useState } from "react";
import type {
  IndicatorLifecycleState,
  IndicatorObservation,
  IndicatorStateInterval,
  IndicatorWorkflowPort,
} from "../../mutations/workbookMutationCommandPorts";

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

export type IndicatorInspectorAction =
  | "indicator.observations.manage"
  | "indicator.observations.pivot"
  | "indicator.lifecycle.read"
  | "indicator.lifecycle.manage";

export function isIndicatorInspectorAction(
  value: string | undefined,
): value is IndicatorInspectorAction {
  return (
    value === "indicator.observations.manage" ||
    value === "indicator.observations.pivot" ||
    value === "indicator.lifecycle.read" ||
    value === "indicator.lifecycle.manage"
  );
}

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
  readonly onMutationCommitted?: (() => Promise<void> | void) | undefined;
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
  const [message, setMessage] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [sourceFieldKey, setSourceFieldKey] = useState(
    sourceFields[0]?.fieldKey ?? "",
  );
  const [spanStart, setSpanStart] = useState("0");
  const [spanEnd, setSpanEnd] = useState("1");
  const [parsedType, setParsedType] = useState("");
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

  useEffect(() => {
    let active = true;
    setMessage(null);
    if (
      (action === "indicator.observations.pivot" ||
        action === "indicator.lifecycle.read" ||
        action === "indicator.lifecycle.manage") &&
      !indicatorRecordId
    ) {
      setMessage("The selected row is not an Indicator record.");
      return () => {
        active = false;
      };
    }
    if (action === "indicator.observations.pivot" && indicatorRecordId) {
      setBusy(true);
      void port
        .listObservations({ indicatorRecordId })
        .then((outcome) => {
          if (!active) return;
          if (outcome.kind === "accepted") setObservations(outcome.value);
          else setMessage(outcome.failure.message);
        })
        .finally(() => {
          if (active) setBusy(false);
        });
    }
    if (action === "indicator.observations.manage" && sourceRecordId) {
      setBusy(true);
      void port
        .listSourceObservations({ sourceRecordId })
        .then((outcome) => {
          if (!active) return;
          if (outcome.kind === "accepted") setObservations(outcome.value);
          else setMessage(outcome.failure.message);
        })
        .finally(() => {
          if (active) setBusy(false);
        });
    }
    if (
      (action === "indicator.lifecycle.read" ||
        action === "indicator.lifecycle.manage") &&
      indicatorRecordId
    ) {
      setBusy(true);
      void port
        .listStateIntervals({ indicatorRecordId })
        .then((outcome) => {
          if (!active) return;
          if (outcome.kind === "accepted") setIntervals(outcome.value);
          else setMessage(outcome.failure.message);
        })
        .finally(() => {
          if (active) setBusy(false);
        });
    }
    return () => {
      active = false;
    };
  }, [action, indicatorRecordId, port, sourceRecordId]);

  if (action === null) return null;

  const submitObservation = async (event: FormEvent) => {
    event.preventDefault();
    if (!sourceRecordId || sourceFieldKey === "") {
      setMessage("Select a source field.");
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
      setMessage("Enter a valid UTF-8 byte span.");
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
      setMessage(outcome.failure.message);
      return;
    }
    setMessage(`Observation ${outcome.value.observation_id} created.`);
    setObservations((current) => [outcome.value, ...current]);
    await onMutationCommitted?.();
  };

  const transitionObservation = async (
    observation: IndicatorObservation,
    transition: "dismiss" | "resolve" | "restore",
  ) => {
    const target = resolvedIndicatorID.trim();
    if (transition === "resolve" && target === "") {
      setMessage("Enter the Indicator ID to resolve this observation.");
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
      setMessage(outcome.failure.message);
      return;
    }
    setObservations((current) =>
      current.map((candidate) =>
        candidate.observation_id === outcome.value.observation_id
          ? outcome.value
          : candidate,
      ),
    );
    setMessage(`Observation ${outcome.value.observation_id} updated.`);
    await onMutationCommitted?.();
  };

  const submitLifecycle = async (event: FormEvent) => {
    event.preventDefault();
    if (!indicatorRecordId || validFrom.trim() === "") {
      setMessage("Enter a valid-from timestamp.");
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
      setMessage("Confidence must be an integer from 0 through 100.");
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
      setMessage(outcome.failure.message);
      return;
    }
    setIntervals((current) => [outcome.value, ...current]);
    setMessage(`Lifecycle interval ${outcome.value.interval_id} created.`);
    await onMutationCommitted?.();
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
              onChange={(event) => setParsedType(event.target.value)}
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
      {busy ? <p style={messageStyle}>Loading…</p> : null}
      {message ? <p style={messageStyle}>{message}</p> : null}
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
