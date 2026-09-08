import { validatedPublicErrorReason } from "../services/publicErrorIdentity";
import type { AuthorizationRecoveryResult } from "../shared/authorizationRecovery";
import type {
  listIncidentMembershipAuditPage,
  MembershipAuditResult,
} from "./api/incidentMembershipAuditClient";
import {
  type AuditField,
  emptyMembershipAuditQuery,
  initialMembershipAuditState,
  type MembershipAuditAuthority,
  type MembershipAuditIntent,
  type MembershipAuditProblem,
  type MembershipAuditState,
  membershipAuditBinding,
  membershipAuditDeadline,
  membershipAuditHistoryLimit,
  membershipAuditLimit,
  membershipAuditPageSummary,
  sameMembershipAuditQuery,
  validateMembershipAuditInputs,
} from "./incidentMembershipAuditModel";

export type MembershipAuditPorts = {
  readonly list: typeof listIncidentMembershipAuditPage;
  readonly isCurrent: (authority: MembershipAuditAuthority) => boolean;
  readonly recover: (
    authority: MembershipAuditAuthority,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<AuthorizationRecoveryResult>;
  readonly lost: (
    reason: "role" | "incident" | "session",
    authority: MembershipAuditAuthority,
  ) => void;
};
type Admission = {
  readonly generation: number;
  readonly authority: MembershipAuditAuthority;
  readonly intent: MembershipAuditIntent;
  readonly abort: AbortController;
  timer?: ReturnType<typeof setTimeout>;
};
const cursorMessage =
  "This audit continuation is no longer usable. Reload the first page with the applied filters.";

/** One incident's read workflow. Session acceptance and workbook navigation are external owners. */
export class IncidentMembershipAuditController {
  private state = initialMembershipAuditState();
  private listeners = new Set<() => void>();
  private generation = 0;
  private admission: Admission | null = null;
  private cursorRejected = false;
  private disposed = false;
  constructor(private readonly ports: MembershipAuditPorts) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(
    change: Partial<MembershipAuditState>,
    announcement?: string,
    announcementRole: "status" | "alert" = "status",
  ) {
    this.state = {
      ...this.state,
      ...change,
      ...(announcement === undefined ? {} : { announcement, announcementRole }),
    };
    for (const listener of this.listeners) listener();
  }
  private cancel() {
    ++this.generation;
    const previous = this.admission;
    this.admission = null;
    clearTimeout(previous?.timer);
    previous?.abort.abort();
  }
  retire = () => {
    this.cancel();
    this.cursorRejected = false;
    this.state = initialMembershipAuditState();
    for (const listener of this.listeners) listener();
  };
  dispose = () => {
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  };
  setAuthority(authority: MembershipAuditAuthority | null) {
    const prior = this.state.authority;
    if (authority === null && prior === null) return;
    if (
      authority &&
      prior &&
      membershipAuditBinding(authority, emptyMembershipAuditQuery) ===
        membershipAuditBinding(prior, emptyMembershipAuditQuery)
    )
      return;
    const active = this.state.active;
    this.retire();
    this.publish({
      authority,
      active,
      read: authority ? { kind: "idle" } : { kind: "denied", reason: "role" },
    });
    if (active && authority) this.first("initial");
  }
  setActive = (active: boolean) => {
    if (this.disposed || this.state.active === active) return;
    this.cancel();
    this.publish({
      active,
      page: null,
      history: [],
      expandedId: null,
      errors: {},
      read: this.cursorRejected
        ? { kind: "cursor_rejected" }
        : { kind: "idle" },
      announcement: "",
    });
    if (active && this.state.authority)
      this.first(this.cursorRejected ? "access" : "initial");
  };
  private available() {
    return (
      !this.disposed &&
      this.state.active &&
      this.state.authority !== null &&
      this.ports.isCurrent(this.state.authority)
    );
  }
  private current(admission: Admission) {
    return (
      this.admission === admission &&
      admission.generation === this.generation &&
      this.available() &&
      this.state.authority !== null &&
      membershipAuditBinding(admission.authority, admission.intent.query) ===
        membershipAuditBinding(this.state.authority, admission.intent.query)
    );
  }
  edit = (field: AuditField, value: string) => {
    if (!this.available() || this.state.read.kind === "denied") return;
    const errors = { ...this.state.errors };
    delete errors[field];
    this.publish({ inputs: { ...this.state.inputs, [field]: value }, errors });
  };
  apply = () => {
    if (!this.available() || this.state.read.kind === "denied") return;
    const { query, errors } = validateMembershipAuditInputs(this.state.inputs);
    this.publish({ errors });
    if (Object.keys(errors).length) {
      this.publish(
        {},
        "Filters were not applied. Check the highlighted fields.",
        "alert",
      );
      return;
    }
    const unchanged = sameMembershipAuditQuery(query, this.state.applied);
    this.publish({ applied: query });
    this.first(unchanged ? "refresh" : "replace");
  };
  clearFilters = () => {
    if (!this.available() || this.state.read.kind === "denied") return;
    this.publish({ inputs: emptyMembershipAuditQuery, errors: {} });
    this.apply();
  };
  refresh = () => {
    if (this.available() && this.state.read.kind !== "denied")
      this.first("refresh");
  };
  private first(kind: MembershipAuditIntent["kind"]) {
    this.start({
      kind,
      query: this.state.applied,
      position: { cursor: null, number: 1 },
      history: [],
    });
  }
  canContinue = () => {
    const { page, read, applied } = this.state;
    return (
      this.available() &&
      !this.cursorRejected &&
      read.kind !== "pending" &&
      read.kind !== "denied" &&
      page !== null &&
      sameMembershipAuditQuery(page.query, applied)
    );
  };
  next = () => {
    const { page, history } = this.state;
    if (
      !this.canContinue() ||
      !page?.paging.has_more ||
      !page.paging.next_cursor
    )
      return;
    const cursor = page.paging.next_cursor;
    if (
      [...history, page.position].some((position) => position.cursor === cursor)
    ) {
      this.rejectCursor();
      return;
    }
    this.start({
      kind: "page",
      query: page.query,
      position: { cursor, number: page.position.number + 1 },
      history: [...history, page.position].slice(-membershipAuditHistoryLimit),
    });
  };
  previous = () => {
    const { page, history } = this.state;
    const position = history.at(-1);
    if (!this.canContinue() || !page || !position) return;
    this.start({
      kind: "page",
      query: page.query,
      position,
      history: history.slice(0, -1),
    });
  };
  retry = () => {
    if (!this.available()) return;
    const read = this.state.read;
    if (read.kind === "failed") this.start(read.intent);
    if (read.kind === "denied") this.first("initial");
  };
  toggleExpanded = (id: string) => {
    if (
      !this.available() ||
      this.state.read.kind === "denied" ||
      !this.state.page?.rows.some((event) => event.audit_event_id === id)
    )
      return;
    this.publish({ expandedId: this.state.expandedId === id ? null : id });
  };
  private start(intent: MembershipAuditIntent) {
    if (!this.available() || !this.state.authority) return;
    const pending = this.admission?.intent;
    if (
      pending &&
      pending.kind !== "access" &&
      intent.kind !== "access" &&
      pending.position.cursor === intent.position.cursor &&
      sameMembershipAuditQuery(pending.query, intent.query)
    )
      return;
    this.cancel();
    if (intent.kind !== "access" && intent.position.cursor === null)
      this.cursorRejected = false;
    const admission: Admission = {
      generation: this.generation,
      authority: this.state.authority,
      intent,
      abort: new AbortController(),
    };
    this.admission = admission;
    admission.timer = setTimeout(() => {
      if (this.current(admission))
        this.fail(admission, {
          kind: "unavailable",
          message: "The membership audit read timed out. Try the read again.",
        });
    }, membershipAuditDeadline);
    this.publish(
      { read: { kind: "pending", intent } },
      intent.kind === "access"
        ? "Checking membership audit access."
        : intent.kind === "page"
          ? `Loading membership audit page ${intent.position.number}.`
          : this.state.page
            ? "Refreshing membership audit. Displayed events are previous results."
            : "Loading membership audit…",
    );
    void this.execute(admission).catch(() => {
      if (this.current(admission))
        this.fail(admission, {
          kind: "unavailable",
          message: this.unavailableMessage(admission),
        });
    });
  }
  private async execute(admission: Admission) {
    const access = await this.ports.recover(
      admission.authority,
      admission.abort.signal,
      () => this.current(admission),
    );
    if (!this.current(admission)) return;
    if (access.kind === "session_lost") {
      this.deny(admission, "session");
      return;
    }
    if (access.kind === "access_lost") {
      this.deny(admission, "incident");
      return;
    }
    if (
      access.kind === "authorized" &&
      (access.role !== "admin" || access.userId !== admission.authority.actorId)
    ) {
      this.deny(
        admission,
        access.userId === admission.authority.actorId ? "role" : "session",
      );
      return;
    }
    if (access.kind !== "authorized") {
      this.fail(admission, {
        kind: "access",
        message:
          "Current membership audit access could not be checked. Try the read again.",
      });
      return;
    }
    if (admission.intent.kind === "access") {
      this.rejectCursor();
      return;
    }
    const result = await this.ports.list({
      authority: admission.authority,
      query: admission.intent.query,
      limit: membershipAuditLimit,
      cursor: admission.intent.position.cursor,
      signal: admission.abort.signal,
    });
    if (this.current(admission)) this.accept(admission, result);
  }
  private accept(admission: Admission, result: MembershipAuditResult) {
    if (!result.ok) {
      if (result.status === 401 && result.error.code === "session_required") {
        this.deny(admission, "session");
        return;
      }
      if (
        result.status === 403 &&
        result.error.code === "authorization_denied"
      ) {
        this.deny(admission, "role");
        return;
      }
      if (result.status === 404 && result.error.code === "incident_not_found") {
        this.deny(admission, "incident");
        return;
      }
      if (
        result.error.code === "invalid_pagination_request" &&
        admission.intent.position.cursor !== null
      ) {
        this.rejectCursor();
        return;
      }
      const reason = validatedPublicErrorReason(
        result.error.code,
        result.error.details?.reason_code,
      );
      this.fail(
        admission,
        result.error.code === "invalid_list_query"
          ? {
              kind: "query",
              message:
                reason === "invalid_filter_range"
                  ? "The applied time range was rejected. Check the explicit timezones and inclusive lower/exclusive upper bounds."
                  : "The applied filters were rejected. Review the exact filter values and apply them again.",
            }
          : {
              kind: "unavailable",
              message: this.unavailableMessage(admission),
            },
      );
      return;
    }
    if (
      result.paging.has_more &&
      admission.intent.history.some(
        (position) => position.cursor === result.paging.next_cursor,
      )
    ) {
      this.rejectCursor();
      return;
    }
    const page = {
      binding: membershipAuditBinding(
        admission.authority,
        admission.intent.query,
      ),
      query: admission.intent.query,
      position: admission.intent.position,
      rows: result.rows,
      paging: result.paging,
    };
    this.cancel();
    this.publish(
      {
        page,
        history: admission.intent.history,
        read: { kind: "idle" },
        expandedId: result.rows.some(
          (event) => event.audit_event_id === this.state.expandedId,
        )
          ? this.state.expandedId
          : null,
      },
      membershipAuditPageSummary(page),
    );
  }
  private unavailableMessage(admission: Admission) {
    return this.state.page
      ? admission.intent.kind === "page"
        ? "The requested audit page could not be loaded. Displayed events were kept. Try the page again."
        : "Membership audit refresh failed. Displayed events may be stale. Try the read again."
      : "Membership audit is unavailable. Try the read again.";
  }
  private fail(admission: Admission, problem: MembershipAuditProblem) {
    this.cancel();
    this.publish(
      { read: { kind: "failed", intent: admission.intent, problem } },
      problem.message,
    );
  }
  private rejectCursor() {
    this.cancel();
    this.cursorRejected = true;
    this.publish(
      { read: { kind: "cursor_rejected" }, history: [] },
      cursorMessage,
    );
  }
  private deny(admission: Admission, reason: "role" | "incident" | "session") {
    const authority = admission.authority;
    this.retire();
    this.publish(
      { authority, active: true, read: { kind: "denied", reason } },
      reason === "role"
        ? "Only incident admins can review membership audit."
        : reason === "incident"
          ? "Incident access has ended."
          : "Your session has ended.",
      "alert",
    );
    this.ports.lost(reason, authority);
  }
}
