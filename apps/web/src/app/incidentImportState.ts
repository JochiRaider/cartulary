import {
  type ImportAdmissionOutcome,
  type ImportAttempt,
  type ImportCancelAttempt,
  type ImportCancellationOutcome,
  type ImportObservationOutcome,
  type ImportProblem,
  type IncidentImportJob,
  importedIncidentTarget,
  importJobAdvances,
  terminalImportJob,
} from "./api/incidentImportClient";

type Admission =
  | { readonly kind: "idle" }
  | { readonly kind: "pending" | "uncertain"; readonly attempt: ImportAttempt }
  | { readonly kind: "rejected"; readonly problem: ImportProblem }
  | { readonly kind: "accepted"; readonly jobId: string };
type Observation =
  | { readonly kind: "stale" | "ready" | "unavailable" }
  | { readonly kind: "failed"; readonly problem: ImportProblem };
type Cancellation =
  | { readonly kind: "idle" | "observed" }
  | {
      readonly kind: "pending" | "uncertain";
      readonly attempt: ImportCancelAttempt;
    }
  | { readonly kind: "rejected"; readonly problem: ImportProblem };
export type ImportAction =
  | { readonly kind: "idle" }
  | {
      readonly kind: "checking" | "failed";
      readonly action: "open" | "cancel";
      readonly jobId: string;
    };
type Navigation =
  | { readonly kind: "idle" }
  | {
      readonly kind: "opening" | "opened" | "unavailable" | "access_lost";
      readonly jobId: string;
    };
export type KnownImport = {
  readonly job: IncidentImportJob;
  readonly filename: string;
  readonly observation: Observation;
  readonly cancellation: Cancellation;
  readonly reading: boolean;
  readonly availability: "unchecked" | "available" | "unavailable";
};
export type IncidentImportState = {
  readonly access: "checking" | "ready" | "unavailable" | "lost";
  readonly selectedFile: File | null;
  readonly fieldError: "required" | null;
  readonly admission: Admission;
  readonly jobs: Readonly<Record<string, KnownImport>>;
  readonly order: readonly string[];
  readonly selectedJobId: string | null;
  readonly paused: boolean;
  readonly navigation: Navigation;
  readonly action: ImportAction;
  readonly announcement: {
    readonly sequence: number;
    readonly text: string;
    readonly priority: "polite" | "assertive";
  };
};
export type ImportAuthority = {
  readonly lifetime: string;
  readonly actorId: string;
};

export const initialImportState = (): IncidentImportState => ({
  access: "checking",
  selectedFile: null,
  fieldError: null,
  admission: { kind: "idle" },
  jobs: {},
  order: [],
  selectedJobId: null,
  paused: false,
  navigation: { kind: "idle" },
  action: { kind: "idle" },
  announcement: { sequence: 0, text: "", priority: "polite" },
});
export const importStatusLabel: Record<IncidentImportJob["status"], string> = {
  queued: "Queued",
  running: "Processing",
  cancel_requested: "Cancellation requested",
  succeeded: "Import succeeded",
  failed: "Import failed",
  canceled: "Import canceled",
};
export const admissionUnresolved = (state: IncidentImportState) =>
  state.admission.kind === "pending" || state.admission.kind === "uncertain";
const availableImport = (entry: KnownImport) =>
  entry.availability !== "unavailable";
export const cancelableImport = (entry: KnownImport) =>
  availableImport(entry) &&
  entry.job.cancelable &&
  (entry.job.status === "queued" || entry.job.status === "running") &&
  entry.cancellation.kind !== "pending";
export const openableImport = (entry: KnownImport) =>
  availableImport(entry) ? importedIncidentTarget(entry.job) : null;

export type ImportEvent =
  | { type: "reset" }
  | { type: "access"; access: IncidentImportState["access"] }
  | { type: "file"; file: File | null }
  | { type: "required" }
  | { type: "admission_started"; attempt: ImportAttempt }
  | {
      type: "admission_finished";
      attempt: ImportAttempt;
      outcome: Exclude<ImportAdmissionOutcome, { kind: "access_failed" }>;
      actorId: string;
    }
  | { type: "selected"; id: string }
  | { type: "paused"; paused: boolean }
  | { type: "inactive" | "reads_stopped" }
  | { type: "read_started"; id: string }
  | {
      type: "read_finished";
      id: string;
      outcome: Exclude<ImportObservationOutcome, { kind: "access_failed" }>;
      active: boolean;
    }
  | { type: "cancel_started"; id: string; attempt: ImportCancelAttempt }
  | {
      type: "cancel_finished";
      id: string;
      attempt: ImportCancelAttempt;
      outcome: Exclude<ImportCancellationOutcome, { kind: "access_failed" }>;
      active: boolean;
    }
  | { type: "navigation"; navigation: Navigation }
  | { type: "action"; action: ImportAction }
  | { type: "announce"; text: string; priority: "polite" | "assertive" };

function updateJob(
  state: IncidentImportState,
  id: string,
  update: (entry: KnownImport) => KnownImport,
): IncidentImportState {
  const entry = state.jobs[id];
  return entry
    ? { ...state, jobs: { ...state.jobs, [id]: update(entry) } }
    : state;
}
function acceptSnapshot(
  entry: KnownImport,
  job: IncidentImportJob,
  active: boolean,
): KnownImport | null {
  if (!importJobAdvances(entry.job, job)) return null;
  return {
    ...entry,
    job,
    reading: false,
    availability: "available",
    observation: { kind: active ? "ready" : "stale" },
    cancellation:
      job.status === "cancel_requested" || terminalImportJob(job)
        ? { kind: "observed" }
        : entry.cancellation,
  };
}
/** Pure transitions: operation/lifetime fencing and all effects belong to the executor. */
export function transitionImport(
  state: IncidentImportState,
  event: ImportEvent,
): IncidentImportState {
  switch (event.type) {
    case "reset":
      return initialImportState();
    case "access":
      return { ...state, access: event.access };
    case "file":
      return admissionUnresolved(state)
        ? state
        : {
            ...state,
            selectedFile: event.file,
            fieldError: null,
            admission: { kind: "idle" },
          };
    case "required":
      return { ...state, fieldError: "required" };
    case "admission_started":
      return {
        ...state,
        admission: { kind: "pending", attempt: event.attempt },
        fieldError: null,
      };
    case "admission_finished": {
      if (
        state.admission.kind !== "pending" ||
        state.admission.attempt !== event.attempt
      )
        return state;
      const outcome = event.outcome;
      if (
        outcome.kind !== "accepted" ||
        outcome.job.submitted_by_user_id !== event.actorId
      )
        return {
          ...state,
          admission:
            outcome.kind === "rejected"
              ? { kind: "rejected", problem: outcome.problem }
              : { kind: "uncertain", attempt: event.attempt },
        };
      const job = outcome.job;
      const existing = state.jobs[job.job_id];
      return {
        ...state,
        selectedFile: null,
        admission: { kind: "accepted", jobId: job.job_id },
        selectedJobId: job.job_id,
        navigation: { kind: "idle" },
        order: existing ? state.order : [...state.order, job.job_id],
        jobs: {
          ...state.jobs,
          [job.job_id]: existing ?? {
            job,
            filename: event.attempt.filename,
            observation: { kind: "stale" },
            reading: false,
            availability: "unchecked",
            cancellation: { kind: "idle" },
          },
        },
      };
    }
    case "selected":
      return state.jobs[event.id]
        ? { ...state, selectedJobId: event.id, navigation: { kind: "idle" } }
        : state;
    case "paused":
      return { ...state, paused: event.paused };
    case "inactive":
    case "reads_stopped": {
      const inactive = event.type === "inactive";
      return {
        ...state,
        jobs: Object.fromEntries(
          Object.entries(state.jobs).map(([id, entry]) => [
            id,
            {
              ...entry,
              reading: false,
              observation:
                inactive && entry.observation.kind === "ready"
                  ? { kind: "stale" as const }
                  : entry.observation,
            },
          ]),
        ),
        announcement: inactive
          ? { ...state.announcement, text: "" }
          : state.announcement,
      };
    }
    case "read_started":
      return updateJob(state, event.id, (entry) => ({
        ...entry,
        reading: true,
      }));
    case "read_finished":
      return updateJob(state, event.id, (entry) => {
        const result = event.outcome;
        if (result.kind === "observed") {
          const accepted = acceptSnapshot(entry, result.job, event.active);
          if (accepted) return accepted;
        }
        return {
          ...entry,
          reading: false,
          availability:
            result.kind === "unavailable" ? "unavailable" : entry.availability,
          observation:
            result.kind === "unavailable"
              ? { kind: "unavailable" }
              : {
                  kind: "failed",
                  problem:
                    result.kind === "failed" ? result.problem : "contract",
                },
        };
      });
    case "cancel_started":
      return updateJob(state, event.id, (entry) => ({
        ...entry,
        cancellation: { kind: "pending", attempt: event.attempt },
      }));
    case "cancel_finished":
      return updateJob(state, event.id, (entry) => {
        if (
          entry.cancellation.kind !== "pending" ||
          entry.cancellation.attempt !== event.attempt
        )
          return entry;
        const result = event.outcome;
        if (result.kind === "acknowledged") {
          const accepted = acceptSnapshot(entry, result.job, event.active);
          if (accepted)
            return { ...accepted, cancellation: { kind: "observed" } };
        }
        const uncertain =
          result.kind === "uncertain" || result.kind === "acknowledged";
        return {
          ...entry,
          availability:
            result.kind === "unavailable" ? "unavailable" : entry.availability,
          cancellation: uncertain
            ? { kind: "uncertain", attempt: event.attempt }
            : {
                kind: "rejected",
                problem:
                  result.kind === "rejected" ? result.problem : "unavailable",
              },
          observation: {
            kind: result.kind === "unavailable" ? "unavailable" : "stale",
          },
        };
      });
    case "navigation":
      return { ...state, navigation: event.navigation };
    case "action":
      return { ...state, action: event.action };
    case "announce":
      return {
        ...state,
        announcement: {
          sequence: state.announcement.sequence + 1,
          text: event.text,
          priority: event.priority,
        },
      };
  }
}
