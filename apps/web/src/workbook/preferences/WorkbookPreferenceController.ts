import type { AuthorizationRecoveryResult } from "../../shared/authorizationRecovery";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookPreferencePort } from "../ports/WorkbookPreferencePort";
import {
  emptyPreferenceSlot,
  frozenPreferencePointer,
  type PreferenceAttempt,
  type PreferenceAuthority,
  type PreferenceKind,
  type PreferenceResources,
  type PreferenceResult,
  type PreferenceSlot,
  type PreferenceSnapshot,
  type PreferenceSurface,
  preferencePointer,
  preferencePointersEqual,
  preferenceProblemText,
  validPreferenceResource,
} from "./workbookPreferenceModel";

type Observation = { cancel: () => void };
export type PreferenceObserver = <T>(
  request: (signal: AbortSignal) => Promise<T>,
) => {
  result: Promise<
    | { kind: "completed"; value: T }
    | { kind: "timeout" | "transport" | "cancelled" }
  >;
  settled: Promise<void>;
  cancel: () => void;
};
type ResourceWork = {
  read: Observation | null;
  write: Observation | null;
  readGeneration: number;
  transport: object | null;
  followup: number;
};
const emptyWork = (): ResourceWork => ({
  read: null,
  write: null,
  readGeneration: 0,
  transport: null,
  followup: 0,
});
const initial = (): PreferenceSnapshot => ({
  authority: null,
  access: "checking",
  home: emptyPreferenceSlot<"home">(),
  default: emptyPreferenceSlot<"default">(),
  surface: null,
  departure: false,
  announcement: null,
});

/** Two independent resources in one current incident/session. No startup or saved-view policy. */
export class WorkbookPreferenceController {
  private state = initial();
  private epoch = 0;
  private sequence = 0;
  private announcement = 0;
  private inspectionActive = false;
  private listeners = new Set<() => void>();
  private work = { home: emptyWork(), default: emptyWork() };
  private departure: ((leave: boolean) => void) | null = null;
  constructor(
    private readonly ports: {
      port: (authority: PreferenceAuthority) => WorkbookPreferencePort;
      isCurrent: (authority: PreferenceAuthority) => boolean;
      recover: (
        authority: PreferenceAuthority,
        signal: AbortSignal,
        current: () => boolean,
      ) => Promise<AuthorizationRecoveryResult>;
      lost: (
        reason: "session" | "incident",
        authority: PreferenceAuthority,
      ) => void;
      observe: PreferenceObserver;
    },
  ) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(patch: Partial<PreferenceSnapshot>) {
    this.state = { ...this.state, ...patch };
    for (const listener of this.listeners) listener();
  }
  private slot<K extends PreferenceKind>(
    kind: K,
    patch: Partial<PreferenceSlot<K>>,
  ) {
    this.publish({ [kind]: { ...this.state[kind], ...patch } });
  }
  private announce(text: string) {
    this.publish({ announcement: { id: ++this.announcement, text } });
  }
  private current(
    authority = this.state.authority,
    epoch = this.epoch,
  ): authority is PreferenceAuthority {
    const now = this.state.authority;
    return (
      authority !== null &&
      now !== null &&
      epoch === this.epoch &&
      authority.incidentId === now.incidentId &&
      authority.actorId === now.actorId &&
      authority.lifetime === now.lifetime &&
      this.ports.isCurrent(authority)
    );
  }
  setAuthority = (authority: PreferenceAuthority | null) => {
    const previous = this.state.authority;
    if (!authority) {
      if (previous) this.retire();
      return;
    }
    if (
      !previous ||
      previous.incidentId !== authority.incidentId ||
      previous.actorId !== authority.actorId ||
      previous.lifetime !== authority.lifetime ||
      previous.apiBase !== authority.apiBase
    ) {
      this.retire();
      this.publish({
        authority: Object.freeze({ ...authority }),
        access: "ready",
      });
    } else if (previous.role !== authority.role) {
      this.publish({
        authority: Object.freeze({ ...authority }),
        access: "ready",
      });
    }
  };
  setSurface = (surface: PreferenceSurface | null) => {
    const old = this.state.surface;
    if (
      old === surface ||
      (old &&
        surface &&
        preferencePointersEqual(old.sheetRef, surface.sheetRef) &&
        old.label === surface.label &&
        old.available === surface.available &&
        (old.savedViewLabels?.length ?? 0) ===
          (surface.savedViewLabels?.length ?? 0) &&
        (old.savedViewLabels ?? []).every(
          (label, index) =>
            label.id === surface.savedViewLabels?.[index]?.id &&
            label.label === surface.savedViewLabels?.[index]?.label,
        ))
    )
      return;
    this.publish({
      surface: surface
        ? Object.freeze({
            ...surface,
            savedViewLabels: Object.freeze(
              (surface.savedViewLabels ?? []).map((label) =>
                Object.freeze({ ...label }),
              ),
            ),
            sheetRef: Object.freeze({ ...surface.sheetRef }),
          })
        : null,
    });
  };
  setInspectionActive = (active: boolean) => {
    if (this.inspectionActive === active) return;
    this.inspectionActive = active;
    if (active) {
      this.refresh("home");
      this.refresh("default");
    }
  };
  private cancelRead(kind: PreferenceKind) {
    const work = this.work[kind];
    ++work.readGeneration;
    work.read?.cancel();
    work.read = null;
  }
  retire = () => {
    ++this.epoch;
    for (const kind of ["home", "default"] as const) {
      this.cancelRead(kind);
      this.work[kind].write?.cancel();
    }
    this.work = { home: emptyWork(), default: emptyWork() };
    this.inspectionActive = false;
    this.departure?.(false);
    this.departure = null;
    this.state = initial();
    this.publish({});
  };
  dispose = () => {
    this.retire();
    this.listeners.clear();
  };
  private acceptAccess(
    result: AuthorizationRecoveryResult,
    authority: PreferenceAuthority,
  ): boolean {
    if (!this.current(authority)) return false;
    if (result.kind === "session_lost" || result.kind === "access_lost") {
      this.retire();
      this.ports.lost(
        result.kind === "session_lost" ? "session" : "incident",
        authority,
      );
      return false;
    }
    if (
      result.kind !== "authorized" ||
      result.userId !== authority.actorId ||
      result.role === ""
    ) {
      this.publish({ access: "unavailable" });
      return false;
    }
    this.publish({
      authority: Object.freeze({ ...authority, role: result.role }),
      access: "ready",
    });
    return true;
  }
  refresh = (kind: PreferenceKind) => {
    const authority = this.state.authority;
    const work = this.work[kind];
    if (!this.current(authority) || work.read || work.write) return;
    const epoch = this.epoch;
    const generation = ++work.readGeneration;
    const admission: Observation = { cancel: () => {} };
    work.read = admission;
    const current = () =>
      this.current(authority, epoch) &&
      this.work[kind] === work &&
      work.read === admission &&
      work.readGeneration === generation;
    this.slot(kind, {
      read: this.state[kind].resource ? "refreshing" : "loading",
      problem: null,
      observation: null,
    });
    const observation = this.ports.observe(async (signal) => {
      const op = this.state[kind].operation;
      if (
        this.state.access !== "ready" ||
        (op.kind === "rejected" &&
          [
            "authentication_required",
            "authorization_denied",
            "incident_not_found",
            "csrf_failed",
          ].includes(op.problem.code))
      ) {
        const access = await this.ports.recover(authority, signal, current);
        if (!current() || !this.acceptAccess(access, authority) || !current()) {
          if (current())
            this.slot(kind, {
              read: "failed",
              problem: { code: "access_unavailable" },
            });
          return;
        }
      }
      // Reads are independent. Their route checks current authority; recovery is only needed on access failure.
      const port = this.ports.port(authority);
      const result = await (kind === "home"
        ? port.readHome({ signal })
        : port.readDefault({ signal }));
      if (!current()) return;
      if (
        result.kind === "accepted" &&
        validPreferenceResource(result.value, kind, authority)
      ) {
        this.slot(kind, {
          resource: result.value,
          read: "ready",
          problem: null,
          observation: work.transport === null ? generation : null,
        });
      } else {
        const problem =
          result.kind === "accepted"
            ? { code: "invalid_public_contract_response" as const }
            : result.problem;
        this.slot(kind, { read: "failed", problem, observation: null });
        if (
          result.kind !== "accepted" &&
          [401, 403, 404].includes(result.status)
        ) {
          const access = await this.ports.recover(authority, signal, current);
          if (current()) this.acceptAccess(access, authority);
        }
      }
    });
    admission.cancel = observation.cancel;
    void observation.result.then((outcome) => {
      if (!current()) return;
      work.read = null;
      if (outcome.kind !== "completed")
        this.slot(kind, {
          read: "failed",
          problem: {
            code: outcome.kind === "timeout" ? "timeout" : "transport",
          },
          observation: null,
        });
    });
  };
  canWrite = (kind: PreferenceKind): boolean => {
    const slot = this.state[kind];
    return (
      this.current() &&
      this.state.authority?.role !== "" &&
      (kind === "home" || this.state.authority?.role === "admin") &&
      this.work[kind].write === null &&
      this.work[kind].transport === null &&
      slot.operation.kind !== "uncertain" &&
      !this.state.departure
    );
  };
  canSetCurrent = (kind: PreferenceKind): boolean =>
    this.canWrite(kind) && this.state.surface?.available === true;
  setCurrent = (kind: PreferenceKind) => {
    if (this.canSetCurrent(kind) && this.state.surface)
      this.write(kind, this.state.surface.sheetRef);
  };
  clear = (kind: PreferenceKind) => {
    if (this.canWrite(kind)) this.write(kind, null);
  };
  canResolve = (
    kind: PreferenceKind,
    attemptId: number,
    observation: number | null,
  ): boolean => {
    const slot = this.state[kind];
    return (
      this.current() &&
      slot.operation.kind === "uncertain" &&
      slot.operation.attempt.id === attemptId &&
      !slot.transportPending &&
      slot.read === "ready" &&
      observation !== null &&
      slot.observation === observation &&
      !this.state.departure
    );
  };
  resolve = (
    kind: PreferenceKind,
    attemptId: number,
    observation: number | null,
    choice: "write" | "keep",
  ) => {
    if (!this.canResolve(kind, attemptId, observation)) return;
    const operation = this.state[kind].operation;
    if (operation.kind !== "uncertain") return;
    if (
      choice === "write" &&
      kind === "default" &&
      this.state.authority?.role !== "admin"
    )
      return;
    this.slot(kind, {
      operation: { kind: "reviewed", attempt: operation.attempt },
    });
    if (choice === "write") this.write(kind, operation.attempt.target);
    else
      this.announce(
        `${kind === "home" ? "Home" : "Incident default"} recovery ended. Keeping the observed value; no write was sent.`,
      );
  };
  private write(kind: PreferenceKind, target: SheetRef | null) {
    if (!this.canWrite(kind)) return;
    const authority = this.state.authority;
    if (!authority) return;
    const attempt: PreferenceAttempt = Object.freeze({
      id: ++this.sequence,
      kind,
      authority: Object.freeze({ ...authority }),
      target: frozenPreferencePointer(target),
    });
    const epoch = this.epoch;
    const work = this.work[kind];
    this.cancelRead(kind);
    const admission: Observation = { cancel: () => {} };
    work.write = admission;
    const current = () =>
      this.current(authority, epoch) && this.work[kind] === work;
    const attemptCurrent = () => {
      const op = this.state[kind].operation;
      return (
        current() &&
        "attempt" in op &&
        op.attempt.id === attempt.id &&
        (op.kind === "pending" || op.kind === "uncertain")
      );
    };
    this.slot(kind, {
      observation: null,
      read: ["loading", "refreshing"].includes(this.state[kind].read)
        ? this.state[kind].resource
          ? "ready"
          : "idle"
        : this.state[kind].read,
      operation: { kind: "pending", stage: "authorization", attempt },
    });
    this.announce(
      `Checking current access for ${kind === "home" ? "my home" : "the incident default"}.`,
    );
    let sent = false;
    const observation = this.ports.observe(async (signal) => {
      const admitted = () =>
        current() && work.write === admission && !signal.aborted;
      const access = await this.ports.recover(authority, signal, admitted);
      if (!admitted() || !this.acceptAccess(access, authority) || !admitted())
        return;
      if (kind === "default" && this.state.authority?.role !== "admin") {
        this.slot(kind, {
          operation: {
            kind: "rejected",
            attempt,
            problem: { code: "authorization_denied" },
          },
        });
        return;
      }
      sent = true;
      const transport = {};
      work.transport = transport;
      this.slot(kind, {
        transportPending: true,
        operation: { kind: "pending", stage: "write", attempt },
      });
      try {
        const port = this.ports.port(authority);
        const result = await (kind === "home"
          ? port.setHomeSheet({ sheetRef: attempt.target, signal })
          : port.setDefaultSheet({ sheetRef: attempt.target, signal }));
        if (attemptCurrent()) this.complete(attempt, result);
      } finally {
        if (current() && work.transport === transport) {
          work.transport = null;
          this.slot(kind, { transportPending: false });
        }
      }
    });
    admission.cancel = observation.cancel;
    void observation.result.then((outcome) => {
      if (!current() || work.write !== admission) return;
      work.write = null;
      if (outcome.kind !== "completed" && sent && attemptCurrent()) {
        this.slot(kind, {
          operation: {
            kind: "uncertain",
            attempt,
            problem: {
              code: outcome.kind === "timeout" ? "timeout" : "transport",
            },
          },
        });
        this.announce(
          `${kind === "home" ? "Home" : "Incident default"} update is uncertain. Observe the current value before deciding whether to write again.`,
        );
      } else if (!sent && this.state[kind].operation.kind === "pending") {
        this.slot(kind, {
          operation: {
            kind: "rejected",
            attempt,
            problem: { code: "access_unavailable" },
          },
        });
        this.publish({ access: "unavailable" });
      }
      if (work.transport === null && sent) this.followup(attempt);
    });
    // An ignored abort may settle after the deadline. One fresh observation then becomes eligible for recovery.
    void observation.settled.then(() => {
      if (!sent || !current() || work.write !== null || work.transport !== null)
        return;
      this.followup(attempt);
    });
  }
  private followup(attempt: PreferenceAttempt) {
    const work = this.work[attempt.kind];
    if (work.followup >= attempt.id) return;
    work.followup = attempt.id;
    this.cancelRead(attempt.kind);
    this.refresh(attempt.kind);
  }
  private complete(
    attempt: PreferenceAttempt,
    result: PreferenceResult<PreferenceResources[PreferenceKind]>,
  ) {
    const kind = attempt.kind;
    this.cancelRead(kind);
    if (
      result.kind === "accepted" &&
      validPreferenceResource(result.value, kind, attempt.authority) &&
      preferencePointersEqual(preferencePointer(result.value), attempt.target)
    ) {
      this.slot(kind, {
        resource: result.value,
        read: "ready",
        problem: null,
        observation: null,
        operation: { kind: "confirmed", attempt, resource: result.value },
      });
      this.announce(
        `${kind === "home" ? "Home" : "Incident default"} ${attempt.target === null ? "clear" : "update"} confirmed.`,
      );
    } else {
      const problem =
        result.kind === "accepted"
          ? { code: "invalid_public_contract_response" as const }
          : result.problem;
      const outcome = result.kind === "accepted" ? "uncertain" : result.kind;
      this.slot(kind, {
        operation: { kind: outcome, attempt, problem },
        observation: null,
      });
      this.announce(
        outcome === "uncertain"
          ? `${kind === "home" ? "Home" : "Incident default"} update is uncertain. Observe the current value before deciding whether to write again.`
          : preferenceProblemText(problem),
      );
    }
  }
  hasDepartureWork = () =>
    (["home", "default"] as const).some(
      (kind) =>
        this.state[kind].transportPending ||
        ["pending", "uncertain"].includes(this.state[kind].operation.kind),
    );
  requestLeave = (): Promise<boolean> => {
    if (!this.hasDepartureWork()) return Promise.resolve(true);
    if (this.departure) return Promise.resolve(false);
    this.publish({ departure: true });
    return new Promise((resolve) => {
      this.departure = resolve;
    });
  };
  resolveDeparture = (choice: "stay" | "discard") => {
    const resolve = this.departure;
    this.departure = null;
    if (choice === "discard") this.retire();
    else this.publish({ departure: false });
    resolve?.(choice === "discard");
  };
}
