import { requireViewContract } from "@cartulary/view-contracts";
import {
  useContext,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import {
  workbookFormFieldsStyle as observationStack,
  workbookFormPreservedTextStyle as observationText,
} from "../../components/workbookFormStyles";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import { IndicatorCanonicalAuthoring } from "./IndicatorCanonicalAuthoring";
import { IndicatorCreateContext } from "./IndicatorCreateContext";
import { IndicatorCreateOperationStatus } from "./IndicatorCreateOperationStatus";
import { indicatorCreateAvailable } from "./indicatorCreateModel";
import type { IndicatorCreateOwnerPort } from "./indicatorCreateOperation";
import { ObservationContext } from "./ObservationContext";
import {
  type ObservationDraft,
  observationIndicatorView,
} from "./observationModel";
import type {
  ObservationAttempt,
  ObservationIntent,
  ObservationOwnerPort,
} from "./observationOperation";

type Props = {
  observation: IndicatorObservation;
  disabled: boolean;
  onResolve: (
    intent: ObservationIntent,
    draft: ObservationDraft,
  ) => ObservationAttempt | null | undefined;
};
export function IndicatorCreateFromObservation(props: Props) {
  const owner = useContext(IndicatorCreateContext),
    observations = useContext(ObservationContext);
  return owner && observations ? (
    <CanonicalWorkflow {...props} owner={owner} observations={observations} />
  ) : null;
}
function CanonicalWorkflow({
  observation,
  disabled,
  onResolve,
  owner,
  observations,
}: Props & {
  owner: IndicatorCreateOwnerPort;
  observations: ObservationOwnerPort;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot),
    observationSnapshot = useSyncExternalStore(
      observations.subscribe,
      observations.getSnapshot,
    );
  const [open, setOpen] = useState(false),
    [error, setError] = useState<string | null>(null);
  const alive = useRef(true),
    current = useRef(observation),
    container = useRef<HTMLDivElement>(null),
    heading = useRef<HTMLParagraphElement>(null);
  current.current = observation;
  useEffect(() => {
    alive.current = true;
    return () => {
      alive.current = false;
    };
  }, []);
  const entries = snapshot.entries.filter(
      (entry) =>
        entry.attempt.observation.observation_id === observation.observation_id,
    ),
    entry = entries.at(-1);
  const contract = requireViewContract(observationIndicatorView);
  const pending = owner.busy(observation.observation_id);
  const retainFocus = container.current?.contains(document.activeElement);
  useEffect(() => {
    if (retainFocus && document.activeElement === document.body)
      heading.current?.focus();
  });
  if (
    !snapshot.authority ||
    !indicatorCreateAvailable(contract) ||
    (observation.resolution_status === "dismissed" && !entry)
  )
    return null;
  const draft = owner.drafts.ensure(observation);
  const related = observationSnapshot.entries.filter((item) =>
    entry?.resolutionAttemptIds.includes(item.attempt.id),
  );
  const linkUnknown = related.some(
    (item) =>
      item.phase === "uncertain" || (item.transportPending && !item.receipt),
  );
  const latest = observationSnapshot.entries.reduce(
    (latest, operation) =>
      operation.receipt?.observation.observation_id === latest.observation_id &&
      operation.receipt.observation.row_version > latest.row_version
        ? operation.receipt.observation
        : latest,
    observation,
  );
  const target = entry?.receipt?.row.record_id;
  const linked =
    !!target &&
    latest.resolution_status === "resolved" &&
    latest.resolved_indicator_record_id === target;
  const elsewhere =
    latest.resolution_status === "resolved" &&
    latest.resolved_indicator_record_id !== target;
  return (
    <div ref={container} style={observationStack}>
      <p ref={heading} tabIndex={-1} style={observationText}>
        Canonical Indicator
      </p>
      {entry ? (
        <IndicatorCreateOperationStatus owner={owner} entry={entry} />
      ) : null}
      {entry?.receipt ? (
        <>
          <p role="status" style={observationText}>
            {linkUnknown
              ? "Link outcome unknown. Recover the observation operation separately."
              : linked
                ? "Observation linked to this Indicator."
                : elsewhere
                  ? "Linked elsewhere. This Indicator remains available."
                  : "Not linked to this Indicator."}
          </p>
          {latest.resolution_status === "dismissed" ? (
            <p>Restore the observation explicitly before resolving it.</p>
          ) : !linked ? (
            <WorkbookInspectorActionButton
              disabled={
                disabled ||
                !observations.canSubmit() ||
                observations.busy({ action: "dismiss", observation: latest }) ||
                linkUnknown ||
                latest.row_version !== observation.row_version
              }
              onClick={() => {
                if (!target) return;
                const resolution = onResolve(
                  { action: "resolve", observation: latest, targetId: target },
                  observations.drafts.ensure(
                    `resolve:${observation.observation_id}`,
                  ),
                );
                if (resolution)
                  owner.associateResolution(entry.attempt.id, resolution);
              }}
            >
              {elsewhere
                ? "Reassign observation to this Indicator"
                : "Resolve observation to this Indicator"}
            </WorkbookInspectorActionButton>
          ) : null}
          {!linked && !linkUnknown ? (
            <p style={observationText}>
              The canonical operation is complete. Review or retry observation
              resolution; canonical creation will not run again.
            </p>
          ) : null}
        </>
      ) : (
        <>
          <WorkbookInspectorActionButton
            disabled={disabled || pending || !owner.canSubmit()}
            aria-expanded={open}
            onClick={() => setOpen((value) => !value)}
          >
            {open ? "Close canonical proposal" : "Create canonical Indicator…"}
          </WorkbookInspectorActionButton>
          {open ? (
            <IndicatorCanonicalAuthoring
              observation={observation}
              contract={contract}
              draft={draft}
              disabled={disabled || pending || !owner.canSubmit()}
              onChange={(key, value) =>
                owner.drafts.update(observation.observation_id, key, value)
              }
              onSubmit={() => {
                const result = owner.admit(
                  observation,
                  contract,
                  draft.values,
                  {
                    isCurrent: () =>
                      alive.current &&
                      owner.getSnapshot().generation === snapshot.generation,
                    matchesDraft: () =>
                      owner.drafts.get(observation.observation_id) === draft &&
                      current.current.row_version === observation.row_version,
                  },
                );
                setError(
                  result.kind === "invalid"
                    ? Object.values(result.errors).join(" ")
                    : result.kind === "unavailable"
                      ? result.message
                      : null,
                );
                if (result.kind === "admitted")
                  void owner.execute(result.attempt);
              }}
            />
          ) : null}
        </>
      )}
      {error ? <p role="alert">{error}</p> : null}
    </div>
  );
}
