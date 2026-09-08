import { clientTxnID } from "../services/browserApi";
import type { AuthorizationRecoveryResult } from "../shared/authorizationRecovery";
import { validateAccountEmail } from "./accountInputValidation";
import { observeAccountOperation } from "./accountOperation";
import type {
  listIncidentMembershipPage,
  mutateIncidentMembership,
} from "./api/incidentMembershipManagementClient";
import {
  initialMembershipManagementState,
  type Membership,
  type MembershipAttempt,
  type MembershipAuthority,
  type MembershipEditor,
  type MembershipInput,
  type MembershipManagementState,
  type MembershipPosition,
  type MembershipRefresh,
  type MembershipRole,
  membershipBinding,
  membershipEditorDirty,
  membershipHistoryLimit,
} from "./incidentMembershipManagementModel";

export type MembershipManagementPorts = {
  readonly list: typeof listIncidentMembershipPage;
  readonly mutate: typeof mutateIncidentMembership;
  readonly isCurrent: (authority: MembershipAuthority) => boolean;
  readonly recover: (
    authority: MembershipAuthority,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<AuthorizationRecoveryResult>;
  readonly lost: (
    reason: "incident" | "session",
    authority: MembershipAuthority,
  ) => void;
};
type ReadAdmission = {
  readonly token: number;
  readonly authority: MembershipAuthority;
  readonly position: MembershipPosition;
  readonly history: readonly MembershipPosition[];
  readonly recoveryId: number | null;
  readonly confirmationId: number | null;
  cancel: () => void;
};

/** One incident's membership workflow; App remains the only session and route owner. */
export class IncidentMembershipManagementController {
  private state = initialMembershipManagementState();
  private listeners = new Set<() => void>();
  private epoch = 0;
  private sequence = 0;
  private read: ReadAdmission | null = null;
  private write: { readonly token: number; cancel: () => void } | null = null;
  private transport: object | null = null;
  private accessObservation: (() => void) | null = null;
  private leave: ((accepted: boolean) => void) | null = null;
  private disposed = false;
  constructor(private readonly ports: MembershipManagementPorts) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(change: Partial<MembershipManagementState>) {
    if (this.disposed) return;
    this.state = { ...this.state, ...change };
    for (const listener of this.listeners) listener();
  }
  private current(authority = this.state.authority, epoch = this.epoch) {
    return (
      !this.disposed &&
      authority !== null &&
      epoch === this.epoch &&
      this.state.authority !== null &&
      membershipBinding(authority) ===
        membershipBinding(this.state.authority) &&
      this.ports.isCurrent(authority)
    );
  }
  private cancelRead() {
    const read = this.read;
    this.read = null;
    read?.cancel();
  }
  retire = () => {
    ++this.epoch;
    this.cancelRead();
    this.write?.cancel();
    this.write = null;
    this.accessObservation?.();
    this.accessObservation = null;
    const leave = this.leave;
    this.leave = null;
    leave?.(false);
    // Only a transport settlement token survives retirement, never incident data.
    this.state = {
      ...initialMembershipManagementState(),
      transportPending: this.transport !== null,
    };
    for (const listener of this.listeners) listener();
  };
  dispose = () => {
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  };
  setAuthority = (authority: MembershipAuthority | null) => {
    const previous = this.state.authority;
    if (!authority && !previous) return;
    if (
      !authority ||
      !previous ||
      membershipBinding(authority) !== membershipBinding(previous)
    ) {
      const active = this.state.active;
      this.retire();
      this.publish({ authority, active });
      if (authority && active) this.refresh();
      return;
    }
    if (authority.role === previous.role) return;
    this.publish({ authority });
    if (authority.role !== "admin") {
      const operation = this.state.operation;
      const leave = this.leave;
      this.leave = null;
      leave?.(false);
      this.publish({
        editor: null,
        departure: null,
        fieldError: null,
        operation:
          operation.kind === "uncertain" ||
          operation.kind === "conflicted" ||
          operation.kind === "rejected"
            ? {
                kind: "retired",
                message:
                  operation.kind === "uncertain"
                    ? "The previous action remains unconfirmed. Administrator access was lost; replay is unavailable."
                    : "Administrator access was lost. The previous request was rejected.",
              }
            : operation,
      });
    }
  };
  setActive = (active: boolean) => {
    if (this.disposed || this.state.active === active) return;
    this.cancelRead();
    const operation = this.state.operation;
    this.publish({
      ...(operation.kind === "uncertain" || operation.kind === "conflicted"
        ? { operation: { ...operation, observed: null } }
        : {}),
      active,
      access: "checking",
      page: null,
      history: [],
      read: { kind: "idle" },
    });
    if (active && this.state.authority) this.refresh();
  };
  private acceptAccess(
    result: AuthorizationRecoveryResult,
    authority: MembershipAuthority,
  ): boolean {
    if (result.kind === "session_lost" || result.kind === "access_lost") {
      this.retire();
      this.ports.lost(
        result.kind === "session_lost" ? "session" : "incident",
        authority,
      );
      return false;
    }
    if (result.kind !== "authorized") {
      this.publish({ access: "unavailable", page: null, history: [] });
      return false;
    }
    if (result.userId !== authority.actorId) {
      this.retire();
      return false;
    }
    this.setAuthority({ ...authority, role: result.role });
    this.publish({ access: "ready" });
    return true;
  }
  canPage = () =>
    this.current() &&
    this.state.active &&
    this.state.access === "ready" &&
    this.state.page !== null &&
    this.state.read.kind === "idle";
  refresh = () => this.startRead({ cursor: null, number: 1 }, []);
  next = () => {
    const { page, history } = this.state;
    if (!this.canPage() || !page?.paging.has_more || !page.paging.next_cursor)
      return;
    const cursor = page.paging.next_cursor;
    if ([...history, page.position].some((p) => p.cursor === cursor)) {
      this.rejectCursor();
      return;
    }
    this.startRead(
      { cursor, number: page.position.number + 1 },
      [...history, page.position].slice(-membershipHistoryLimit),
    );
  };
  previous = () => {
    const position = this.state.history.at(-1);
    if (this.canPage() && position)
      this.startRead(position, this.state.history.slice(0, -1));
  };
  retryRead = () => {
    const read = this.state.read;
    if (read.kind === "failed") this.startRead(read.position, read.history);
    else this.refresh();
  };
  private rejectCursor() {
    this.cancelRead();
    this.publish({ read: { kind: "cursor_rejected" }, history: [] });
  }
  private startRead(
    position: MembershipPosition,
    history: readonly MembershipPosition[],
  ) {
    const authority = this.state.authority;
    if (
      !authority ||
      !this.current() ||
      !this.state.active ||
      this.read?.position.cursor === position.cursor
    )
      return;
    this.cancelRead();
    const operation = this.state.operation;
    const admission: ReadAdmission = {
      token: ++this.sequence,
      authority,
      position,
      history,
      recoveryId:
        operation.kind === "uncertain" || operation.kind === "conflicted"
          ? operation.attempt.id
          : null,
      confirmationId:
        operation.kind === "confirmed" ? operation.attempt.id : null,
      cancel: () => {},
    };
    const epoch = this.epoch;
    const current = () =>
      this.read === admission &&
      this.state.active &&
      this.current(authority, epoch);
    this.read = admission;
    this.publish({ read: { kind: "pending", position } });
    this.refreshState(admission.confirmationId, "listRefresh", "pending");
    let accessAccepted = false;
    const observation = observeAccountOperation(async (signal) => {
      const access = await this.ports.recover(authority, signal, current);
      if (!current()) return null;
      if (!this.acceptAccess(access, authority)) return null;
      accessAccepted = true;
      if (!current()) return null;
      return this.ports.list({ authority, cursor: position.cursor, signal });
    });
    admission.cancel = observation.cancel;
    void observation.result.then((outcome) => {
      if (!current()) return;
      this.read = null;
      if (outcome.kind !== "completed" || !outcome.value) {
        if (!accessAccepted)
          this.publish({ access: "unavailable", page: null, history: [] });
        this.publish({
          read: {
            kind: "failed",
            position,
            history,
            reason: this.state.access === "unavailable" ? "access" : "read",
          },
        });
        this.refreshState(admission.confirmationId, "listRefresh", "failed");
        return;
      }
      const result = outcome.value;
      if (!result.ok) {
        if (this.handleDenial(result.status, result.problem.code, authority))
          return;
        if (
          result.problem.code === "invalid_pagination_request" &&
          position.cursor !== null
        )
          this.rejectCursor();
        else
          this.publish({
            read: { kind: "failed", position, history, reason: "read" },
          });
        this.refreshState(admission.confirmationId, "listRefresh", "failed");
        return;
      }
      if (
        result.paging.next_cursor &&
        [...history, position].some(
          (p) => p.cursor === result.paging.next_cursor,
        )
      ) {
        this.rejectCursor();
        this.refreshState(admission.confirmationId, "listRefresh", "failed");
        return;
      }
      let editor = this.state.editor;
      const observed = editor?.base
        ? result.rows.find((r) => r.user_id === editor?.base?.user_id)
        : null;
      if (
        editor?.base &&
        observed &&
        (observed.membership_version !== editor.base.membership_version ||
          observed.joined_at !== editor.base.joined_at)
      )
        editor = { ...editor, reviewRequired: true };
      this.publish({
        page: { rows: result.rows, paging: result.paging, position },
        history,
        read: { kind: "idle" },
        editor,
      });
      const op = this.state.operation;
      if (
        (op.kind === "uncertain" || op.kind === "conflicted") &&
        op.attempt.id === admission.recoveryId &&
        op.attempt.input.kind !== "add"
      ) {
        const userId = op.attempt.input.userId;
        this.publish({
          operation: {
            ...op,
            observed:
              result.rows.find((row) => row.user_id === userId) ?? op.observed,
          },
        });
      }
      this.refreshState(admission.confirmationId, "listRefresh", "done");
    });
  }
  private handleDenial(
    status: number,
    code: string,
    authority: MembershipAuthority,
  ) {
    if (
      (status === 401 && code === "session_required") ||
      (status === 404 && code === "incident_not_found")
    ) {
      this.retire();
      this.ports.lost(status === 401 ? "session" : "incident", authority);
      return true;
    }
    if (status === 403 && code === "authorization_denied") {
      // The denial invalidates write authority, but a non-admin may still read members.
      this.publish({ editor: null, fieldError: null, access: "unavailable" });
      void this.recoverAccess();
      return false;
    }
    return false;
  }
  private refreshState(
    id: number | null,
    field: "listRefresh" | "accessRefresh",
    value: MembershipRefresh,
  ) {
    const op = this.state.operation;
    if (op.kind === "confirmed" && op.attempt.id === id)
      this.publish({ operation: { ...op, [field]: value } });
  }
  recoverAccess = async () => {
    const authority = this.state.authority;
    if (!authority || !this.current() || this.accessObservation) return;
    const epoch = this.epoch;
    const id =
      this.state.operation.kind === "confirmed"
        ? this.state.operation.attempt.id
        : null;
    const current = () => this.current(authority, epoch);
    this.refreshState(id, "accessRefresh", "pending");
    const observation = observeAccountOperation((signal) =>
      this.ports.recover(authority, signal, current),
    );
    this.accessObservation = observation.cancel;
    const result = await observation.result;
    if (this.accessObservation === observation.cancel)
      this.accessObservation = null;
    if (!current()) return;
    if (result.kind !== "completed") {
      this.publish({ access: "unavailable", page: null, history: [] });
      this.refreshState(id, "accessRefresh", "failed");
      return;
    }
    const accepted = this.acceptAccess(result.value, authority);
    this.refreshState(id, "accessRefresh", accepted ? "done" : "failed");
    if (accepted && this.state.active) this.refresh();
  };
  private editorAllowed() {
    return (
      this.state.active &&
      this.current() &&
      this.state.authority?.role === "admin" &&
      this.state.access === "ready"
    );
  }
  openAdd = () =>
    this.chooseEditor({
      id: ++this.sequence,
      revision: 0,
      kind: "add",
      base: null,
      email: "",
      role: "viewer",
      reviewRequired: false,
    });
  openMember = (member: Membership, kind: "role" | "remove") => {
    if (
      member.incident_id !== this.state.authority?.incidentId ||
      !this.state.page?.rows.some(
        (r) =>
          r.user_id === member.user_id &&
          r.membership_version === member.membership_version,
      )
    )
      return;
    this.chooseEditor({
      id: ++this.sequence,
      revision: 0,
      kind,
      base: { ...member },
      email: "",
      role: member.role,
      reviewRequired: false,
    });
  };
  private chooseEditor(next: MembershipEditor | null) {
    if (
      !this.editorAllowed() ||
      this.write ||
      this.state.operation.kind === "uncertain" ||
      this.state.operation.kind === "conflicted"
    )
      return;
    if (membershipEditorDirty(this.state.editor)) {
      this.publish({ departure: { kind: "editor", next } });
      return;
    }
    this.publish({ editor: next, fieldError: null });
  }
  cancelEditor = () => this.chooseEditor(null);
  editEmail = (email: string) => {
    const editor = this.state.editor;
    if (this.editorAllowed() && editor?.kind === "add")
      this.publish({
        editor: { ...editor, email, revision: editor.revision + 1 },
        fieldError: null,
      });
  };
  editRole = (role: MembershipRole) => {
    const editor = this.state.editor;
    if (this.editorAllowed() && editor && editor.kind !== "remove")
      this.publish({
        editor: { ...editor, role, revision: editor.revision + 1 },
        fieldError: null,
      });
  };
  canSubmit = () =>
    this.editorAllowed() &&
    this.state.editor !== null &&
    !this.state.editor.reviewRequired &&
    !this.write &&
    !this.transport &&
    this.state.operation.kind !== "uncertain" &&
    this.state.operation.kind !== "conflicted";
  submit = () => {
    const editor = this.state.editor;
    const authority = this.state.authority;
    if (!this.canSubmit() || !editor || !authority) return;
    let input: MembershipInput;
    if (editor.kind === "add") {
      const email = validateAccountEmail(editor.email);
      if (email.error) {
        this.publish({ fieldError: email.error });
        return;
      }
      try {
        input = {
          kind: "add",
          payload: {
            email: email.value,
            role: editor.role,
            client_txn_id: clientTxnID("incident-membership"),
          },
        };
      } catch {
        this.publish({
          fieldError:
            "A secure request identity could not be created. Try again when secure browser randomness is available.",
        });
        return;
      }
    } else {
      if (!editor.base) return;
      input =
        editor.kind === "role"
          ? {
              kind: "role",
              userId: editor.base.user_id,
              payload: {
                base_membership_version: editor.base.membership_version,
                role: editor.role,
              },
            }
          : {
              kind: "remove",
              userId: editor.base.user_id,
              payload: {
                base_membership_version: editor.base.membership_version,
              },
            };
    }
    const attempt: MembershipAttempt = Object.freeze({
      id: ++this.sequence,
      authority: Object.freeze({ ...authority }),
      input: Object.freeze({
        ...input,
        payload: Object.freeze({ ...input.payload }),
      }) as MembershipInput,
      editorId: editor.id,
      draftRevision: editor.revision,
      base: editor.base && Object.freeze({ ...editor.base }),
      label: editor.base?.display_name ?? editor.email,
    });
    this.dispatch(attempt);
  };
  replay = () => {
    const op = this.state.operation;
    if (
      this.editorAllowed() &&
      !this.write &&
      !this.transport &&
      op.kind === "uncertain" &&
      op.attempt.input.kind === "add"
    )
      this.dispatch(op.attempt);
  };
  private dispatch(attempt: MembershipAttempt) {
    if (
      !this.editorAllowed() ||
      this.write ||
      this.transport ||
      !this.current(attempt.authority)
    )
      return;
    const epoch = this.epoch;
    const admission = { token: ++this.sequence, cancel: () => {} };
    this.write = admission;
    this.cancelRead();
    this.publish({
      read: { kind: "idle" },
      fieldError: null,
      operation: { kind: "pending", attempt, stage: "authorization" },
    });
    const current = () =>
      this.write === admission && this.current(attempt.authority, epoch);
    let sent = false;
    const observation = observeAccountOperation(async (signal) => {
      const access = await this.ports.recover(
        attempt.authority,
        signal,
        current,
      );
      if (!current()) return null;
      if (
        !this.acceptAccess(access, attempt.authority) ||
        !current() ||
        this.state.authority?.role !== "admin" ||
        !this.state.active
      )
        return null;
      const token = {};
      this.transport = token;
      sent = true;
      this.publish({
        transportPending: true,
        operation: { kind: "pending", attempt, stage: "write" },
      });
      try {
        return await this.ports.mutate({
          authority: attempt.authority,
          input: attempt.input,
          signal,
        });
      } finally {
        if (this.transport === token) {
          this.transport = null;
          this.publish({ transportPending: false });
        }
      }
    });
    admission.cancel = observation.cancel;
    void observation.result.then((outcome) => {
      if (!current()) return;
      this.write = null;
      if (outcome.kind !== "completed" || outcome.value === null) {
        if (!sent && outcome.kind !== "completed")
          this.publish({ access: "unavailable", page: null, history: [] });
        this.publish({
          operation: {
            kind: sent ? "uncertain" : "rejected",
            attempt,
            problem: {
              code: sent ? "transport_uncertain" : "access_unavailable",
            },
            observed: null,
          },
        });
        if (this.state.authority?.role !== "admin") this.retireRecovery();
        return;
      }
      const result = outcome.value;
      if (!result.ok) {
        this.publish({
          operation: {
            kind:
              result.status >= 500 || result.status === 0
                ? "uncertain"
                : result.problem.code === "membership_version_conflict" ||
                    result.problem.code === "client_txn_conflict"
                  ? "conflicted"
                  : "rejected",
            attempt,
            problem: result.problem,
            observed: null,
          },
        });
        this.handleDenial(
          result.status,
          result.problem.code,
          attempt.authority,
        );
        if (this.state.authority?.role !== "admin") this.retireRecovery();
        return;
      }
      let editor = this.state.editor;
      if (editor?.id === attempt.editorId) {
        editor =
          editor.revision === attempt.draftRevision
            ? null
            : { ...editor, reviewRequired: editor.kind !== "add" };
      }
      let page = this.state.page;
      if (page && attempt.input.kind !== "add") {
        const userId = attempt.input.userId;
        page = {
          ...page,
          rows: result.resource
            ? page.rows.map((r) =>
                r.user_id === userId ? (result.resource as Membership) : r,
              )
            : page.rows.filter((r) => r.user_id !== userId),
        };
      }
      this.publish({
        editor,
        page,
        operation: {
          kind: "confirmed",
          attempt,
          status: result.status,
          receipt: result.resource,
          listRefresh: "idle",
          accessRefresh: "idle",
        },
      });
      // Acknowledgement is already durable UI state. Follow-up work has no resend path.
      void this.recoverAccess();
    });
  }
  private retireRecovery() {
    const operation = this.state.operation;
    if (
      operation.kind === "uncertain" ||
      operation.kind === "conflicted" ||
      operation.kind === "rejected"
    )
      this.publish({
        editor: null,
        operation: {
          kind: "retired",
          message:
            operation.kind === "uncertain"
              ? "The previous action remains unconfirmed. Administrator access is unavailable."
              : "The previous request was rejected. Administrator access is unavailable.",
        },
      });
  }
  observeCurrent = () => {
    const op = this.state.operation;
    if (op.kind === "uncertain" || op.kind === "conflicted")
      this.publish({ operation: { ...op, observed: null } });
    this.refresh();
  };
  reviewObserved = () => {
    const op = this.state.operation;
    if (!this.editorAllowed() || this.write || this.transport) return;
    if (
      (op.kind === "uncertain" || op.kind === "conflicted") &&
      op.observed &&
      op.attempt.input.kind !== "add"
    ) {
      const role =
        this.state.editor?.role ??
        (op.attempt.input.kind === "role"
          ? op.attempt.input.payload.role
          : op.observed.role);
      this.publish({
        editor: {
          id: ++this.sequence,
          revision: 0,
          kind: op.attempt.input.kind,
          base: { ...op.observed },
          role,
          email: "",
          reviewRequired: false,
        },
        operation: {
          kind: "retired",
          message:
            op.kind === "uncertain"
              ? "The previous attempt remains unconfirmed. Review this new action and save explicitly."
              : "The previous request was rejected. Review this new action and save explicitly.",
        },
        fieldError: null,
      });
      return;
    }
    const editor = this.state.editor;
    const current = this.state.page?.rows.find(
      (r) => r.user_id === editor?.base?.user_id,
    );
    if (editor?.reviewRequired && current)
      this.publish({
        editor: {
          ...editor,
          base: { ...current },
          reviewRequired: false,
          revision: editor.revision + 1,
        },
      });
  };
  forgetRecovery = () => {
    if (!this.editorAllowed() || this.write || this.transport) return;
    const op = this.state.operation;
    if (op.kind === "uncertain" || op.kind === "conflicted") {
      this.cancelRead();
      this.publish({
        operation: {
          kind: "retired",
          message:
            op.kind === "uncertain"
              ? "Local recovery was dismissed. The previous attempt remains unconfirmed."
              : "The rejected action was dismissed.",
        },
        editor: null,
        fieldError: null,
        page: null,
        history: [],
        read: { kind: "idle" },
      });
      // Dismissal cannot make a pre-attempt row a fresh base for another write.
      this.refresh();
    }
  };
  hasDepartureWork = () =>
    this.current() &&
    (membershipEditorDirty(this.state.editor) ||
      this.state.operation.kind === "pending" ||
      this.state.operation.kind === "uncertain" ||
      this.state.operation.kind === "conflicted");
  requestLeave = (): Promise<boolean> => {
    if (!this.hasDepartureWork()) return Promise.resolve(true);
    if (this.leave) return Promise.resolve(false);
    this.publish({ departure: { kind: "route" } });
    return new Promise((resolve) => {
      this.leave = resolve;
    });
  };
  resolveDeparture = (choice: "stay" | "discard") => {
    const departure = this.state.departure;
    if (!departure) return;
    if (departure.kind === "editor") {
      this.publish({
        departure: null,
        ...(choice === "discard"
          ? { editor: departure.next, fieldError: null }
          : {}),
      });
      return;
    }
    const leave = this.leave;
    this.leave = null;
    this.publish({ departure: null });
    if (choice === "discard") this.retire();
    leave?.(choice === "discard");
  };
}
