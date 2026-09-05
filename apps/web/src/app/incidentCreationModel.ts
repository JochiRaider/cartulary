import {
  type APIError,
  clientTxnID,
  type HTTPOperationResult,
} from "../services/browserApi";
import type {
  CreateIncidentRequest,
  CreateIncidentResponse,
} from "./api/publicHttpTypes";

export type CreatedIncident = CreateIncidentResponse["data"];
export type IncidentCreationDraft = {
  incident_key: string;
  title: string;
  description: string;
  severity: string;
  tlp: NonNullable<CreateIncidentRequest["tlp"]> | "";
  current_phase: string;
  primary_external_case_ref: string;
};
export type IncidentCreationField = keyof IncidentCreationDraft;
export type IncidentCreationErrors = Partial<
  Record<IncidentCreationField, string>
>;
type Attempt = Readonly<CreateIncidentRequest>;
type CreationOperation =
  | { kind: "editing" }
  | { kind: "submitting"; attempt: Attempt; replay: boolean }
  | { kind: "rejected"; transactionConflict: boolean }
  | { kind: "uncertain"; attempt: Attempt }
  | {
      kind: "created";
      incident: CreatedIncident;
      handoff: "idle" | "opening" | "failed" | "opened";
    };
export type IncidentCreationState = {
  open: boolean;
  detailsOpen: boolean;
  draft: IncidentCreationDraft;
  errors: IncidentCreationErrors;
  operation: CreationOperation;
  message: string;
  announcement: "polite" | "assertive";
};
export type IncidentCreationBinding = {
  controller: IncidentCreationController;
  state: IncidentCreationState;
};
export type CreationHandoffResult =
  | "opened"
  | "cancelled"
  | "unavailable"
  | "access_lost";
export type IncidentCreationPorts = {
  create: (
    request: Attempt,
    signal: AbortSignal,
  ) => Promise<HTTPOperationResult<CreateIncidentResponse>>;
  newTransactionId?: () => string;
  isCurrentSession: (identity: string) => boolean;
  sessionLost: () => void;
  openIncident: (
    incident: CreatedIncident,
    signal: AbortSignal,
    canNavigate: () => boolean,
  ) => Promise<CreationHandoffResult>;
};

const emptyDraft = (): IncidentCreationDraft => ({
  incident_key: "",
  title: "",
  description: "",
  severity: "",
  tlp: "",
  current_phase: "",
  primary_external_case_ref: "",
});
const initialState = (): IncidentCreationState => ({
  open: false,
  detailsOpen: false,
  draft: emptyDraft(),
  errors: {},
  operation: { kind: "editing" },
  message: "",
  announcement: "polite",
});
const blank = (value: string) => /^\p{White_Space}*$/u.test(value);
export function buildIncidentCreationRequest(
  draft: IncidentCreationDraft,
  transactionId: string,
): Attempt {
  return Object.freeze({
    client_txn_id: transactionId,
    incident_key: draft.incident_key,
    title: draft.title,
    ...(!blank(draft.description) ? { description: draft.description } : {}),
    ...(!blank(draft.severity) ? { severity: draft.severity } : {}),
    ...(draft.tlp !== "" ? { tlp: draft.tlp } : {}),
    ...(!blank(draft.current_phase)
      ? { current_phase: draft.current_phase }
      : {}),
    ...(!blank(draft.primary_external_case_ref)
      ? { primary_external_case_ref: draft.primary_external_case_ref }
      : {}),
  });
}

function fieldError(error: APIError): IncidentCreationErrors {
  if (error.code === "incident_key_conflict")
    return {
      incident_key: "This incident key is already in use. Choose another key.",
    };
  const field = error.details?.field;
  if (
    error.code !== "invalid_incident_create" ||
    typeof field !== "string" ||
    !Object.hasOwn(emptyDraft(), field)
  )
    return {};
  const reason = error.details?.reason_code;
  const message =
    reason === "missing_required_field" ||
    reason === "field_empty_after_normalization" ||
    reason === "field_not_nullable"
      ? "This field is required."
      : reason === "field_too_long"
        ? "Shorten this value and try again."
        : reason === "control_character_not_allowed"
          ? "Remove control characters from this value."
          : field === "tlp"
            ? "Select one of the listed TLP values, or Unset."
            : "Use a shorter value without control characters and try again.";
  return { [field]: message };
}

// Owns one account-scoped operation. UI rendering is never the dispatch lock.
export class IncidentCreationController {
  private state = initialState();
  private listeners = new Set<() => void>();
  private identity: string | null = null;
  private epoch = 0;
  private navigation = 0;
  private observation: AbortController | null = null;
  private handoff: AbortController | null = null;
  private handoffTimeout: ReturnType<typeof setTimeout> | undefined;
  private timeout: ReturnType<typeof setTimeout> | undefined;

  constructor(private readonly ports: IncidentCreationPorts) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(next: IncidentCreationState) {
    this.state = next;
    for (const listener of this.listeners) listener();
  }
  setSession(identity: string | null) {
    if (this.identity === identity) return;
    this.dispose();
    this.identity = identity;
  }
  dispose() {
    this.epoch += 1;
    this.navigation += 1;
    this.identity = null;
    this.observation?.abort();
    this.handoff?.abort();
    this.observation = null;
    this.handoff = null;
    clearTimeout(this.timeout);
    clearTimeout(this.handoffTimeout);
    this.publish(initialState());
  }
  private current(epoch = this.epoch) {
    return (
      epoch === this.epoch &&
      this.identity !== null &&
      this.ports.isCurrentSession(this.identity)
    );
  }
  open = () => {
    if (this.current()) this.publish({ ...this.state, open: true });
  };
  leaveSurface = () => {
    this.navigation += 1;
    this.handoff?.abort();
  };
  close = () => {
    this.leaveSurface();
    this.publish({ ...this.state, open: false });
  };
  setDetailsOpen = (detailsOpen: boolean) =>
    this.publish({ ...this.state, detailsOpen });
  change = <K extends IncidentCreationField>(
    field: K,
    value: IncidentCreationDraft[K],
  ) => {
    if (
      !this.current() ||
      !["editing", "rejected"].includes(this.state.operation.kind)
    )
      return;
    if (
      this.state.operation.kind === "rejected" &&
      this.state.operation.transactionConflict
    )
      return;
    const errors = { ...this.state.errors };
    delete errors[field];
    this.publish({
      ...this.state,
      draft: { ...this.state.draft, [field]: value },
      errors,
      operation: { kind: "editing" },
      message: "",
    });
  };
  startNewAttempt = () => {
    if (
      !this.current() ||
      !["rejected", "created"].includes(this.state.operation.kind) ||
      this.handoff !== null
    )
      return;
    this.navigation += 1;
    this.publish({
      ...initialState(),
      open: true,
      draft:
        this.state.operation.kind === "created"
          ? emptyDraft()
          : this.state.draft,
    });
  };
  submit = () => {
    if (
      !this.current() ||
      !this.state.open ||
      !["editing", "rejected"].includes(this.state.operation.kind)
    )
      return;
    if (
      this.state.operation.kind === "rejected" &&
      this.state.operation.transactionConflict
    )
      return;
    const errors: IncidentCreationErrors = {};
    if (blank(this.state.draft.incident_key))
      errors.incident_key = "Incident key is required.";
    if (blank(this.state.draft.title)) errors.title = "Title is required.";
    if (Object.keys(errors).length > 0) {
      this.publish({
        ...this.state,
        errors,
        message: "Complete the required fields.",
        announcement: "assertive",
      });
      return;
    }
    let attempt: Attempt;
    try {
      attempt = buildIncidentCreationRequest(
        this.state.draft,
        this.ports.newTransactionId?.() ?? clientTxnID("incident-create"),
      );
    } catch {
      this.publish({
        ...this.state,
        message: "The request could not be prepared. Try again.",
        announcement: "assertive",
      });
      return;
    }
    void this.dispatch(attempt, false);
  };
  retry = () => {
    if (!this.current() || this.state.operation.kind !== "uncertain") return;
    void this.dispatch(this.state.operation.attempt, true);
  };

  private async dispatch(attempt: Attempt, replay: boolean) {
    if (this.observation !== null || !this.current()) return;
    const observation = new AbortController();
    this.observation = observation;
    const epoch = this.epoch;
    const navigation = this.navigation;
    const canAccept = () =>
      this.current(epoch) && this.observation === observation;
    const uncertain = (
      message = "Creation could not be confirmed. Retry the same request to check its outcome.",
    ) => {
      this.publish({
        ...this.state,
        operation: { kind: "uncertain", attempt },
        message,
        announcement: "polite",
      });
    };
    this.publish({
      ...this.state,
      errors: {},
      operation: { kind: "submitting", attempt, replay },
      message: replay
        ? "Checking the previous creation request…"
        : "Creating incident…",
      announcement: "polite",
    });
    this.timeout = setTimeout(() => {
      if (!canAccept()) return;
      this.observation = null;
      observation.abort();
      uncertain();
    }, 30_000);
    const timeout = this.timeout;
    try {
      const result = await this.ports.create(attempt, observation.signal);
      if (!canAccept()) return;
      if (result.ok) {
        const incident = result.payload.data;
        this.publish({
          ...this.state,
          draft: emptyDraft(),
          errors: {},
          operation: { kind: "created", incident, handoff: "idle" },
          message: "Incident created.",
          announcement: "polite",
        });
        if (this.navigation === navigation && this.state.open)
          void this.openCreated();
        return;
      }
      const error = result.payload.error;
      if (result.status === 401) {
        this.dispose();
        this.ports.sessionLost();
        return;
      }
      const definitive =
        error !== undefined &&
        ((result.status === 400 && error.code === "invalid_incident_create") ||
          (result.status === 409 &&
            ["incident_key_conflict", "client_txn_conflict"].includes(
              error.code,
            )) ||
          result.status === 403);
      if (!definitive || replay) {
        uncertain(
          result.status === 403
            ? "Access to creation is currently denied. The earlier request may still have created an incident. Retry only this request when access is restored."
            : undefined,
        );
        return;
      }
      const errors = fieldError(error);
      const transactionConflict = error.code === "client_txn_conflict";
      this.publish({
        ...this.state,
        errors,
        detailsOpen:
          this.state.detailsOpen ||
          Object.keys(errors).some(
            (field) => field !== "incident_key" && field !== "title",
          ),
        operation: { kind: "rejected", transactionConflict },
        message: transactionConflict
          ? "This request could not be used. Start a new attempt to submit again."
          : result.status === 403
            ? "Creation is not permitted for this request. Check your session before trying again."
            : "The incident was not created. Correct the indicated fields and try again.",
        announcement: transactionConflict ? "polite" : "assertive",
      });
    } catch {
      if (canAccept()) uncertain();
    } finally {
      clearTimeout(timeout);
      if (this.observation === observation) this.observation = null;
    }
  }

  openCreated = async () => {
    const operation = this.state.operation;
    if (
      !this.current() ||
      operation.kind !== "created" ||
      this.handoff !== null
    )
      return;
    const handoff = new AbortController();
    this.handoff = handoff;
    const epoch = this.epoch;
    const navigation = this.navigation;
    const canNavigate = () =>
      this.current(epoch) &&
      this.navigation === navigation &&
      !handoff.signal.aborted;
    const canAccept = () =>
      this.current(epoch) &&
      this.handoff === handoff &&
      this.state.operation.kind === "created";
    this.publish({
      ...this.state,
      operation: { ...operation, handoff: "opening" },
      message: "Incident created. Opening workbook…",
      announcement: "polite",
    });
    const timeout = setTimeout(() => {
      if (!canAccept()) return;
      const stillRequested = canNavigate();
      this.handoff = null;
      handoff.abort();
      this.publish({
        ...this.state,
        operation: {
          ...operation,
          handoff: stillRequested ? "failed" : "idle",
        },
        message: stillRequested
          ? "Incident created, but the workbook could not be opened. Try opening it again."
          : "Incident created.",
        announcement: "polite",
      });
    }, 30_000);
    this.handoffTimeout = timeout;
    try {
      const result = await this.ports.openIncident(
        operation.incident,
        handoff.signal,
        canNavigate,
      );
      if (!canAccept()) return;
      const opened = result === "opened";
      this.publish({
        ...this.state,
        operation: {
          ...operation,
          handoff: opened
            ? "opened"
            : result === "cancelled"
              ? "idle"
              : "failed",
        },
        message:
          opened || result === "cancelled"
            ? "Incident created."
            : result === "access_lost"
              ? "Incident created. Your session no longer has access to this workbook."
              : "Incident created, but the workbook could not be opened. Try opening it again.",
        announcement: "polite",
      });
    } catch {
      if (canAccept())
        this.publish({
          ...this.state,
          operation: {
            ...operation,
            handoff: canNavigate() ? "failed" : "idle",
          },
          message: canNavigate()
            ? "Incident created, but the workbook could not be opened. Try opening it again."
            : "Incident created.",
          announcement: "polite",
        });
    } finally {
      clearTimeout(timeout);
      if (this.handoff === handoff) this.handoff = null;
    }
  };
}
