import {
  type AuditAuthority,
  type AuditField,
  type AuditInputs,
  type AuditPageDescriptor,
  type AuditProblem,
  type AuditQuery,
  type AuditReadKind,
  type AuditState,
  auditBinding,
  auditHistoryLimit,
  auditPageLimit,
  auditPageSummary,
  auditReadTimeout,
  emptyAuditQuery,
  initialAuditState,
  sameAuditQuery,
  validateAuditInputs,
} from "./administrativeAuditModel";
import {
  type AuditReadResult,
  auditFailureMessage,
  listAdministrativeAuditPage,
} from "./api/administrativeAuditClient";

export type AuditPorts = {
  readonly list: typeof listAdministrativeAuditPage;
  readonly isCurrent: (authority: AuditAuthority) => boolean;
  readonly confirmAccess: (
    authority: AuditAuthority,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<{
    kind:
      | "authorized"
      | "unavailable"
      | "cancelled"
      | "access_lost"
      | "session_lost";
  }>;
  readonly authorizationFailed: (
    status: 401 | 403,
    authority: AuditAuthority,
  ) => void;
};
type Intent = {
  readonly kind: AuditReadKind;
  readonly query: AuditQuery;
  readonly descriptor: AuditPageDescriptor;
  readonly history: readonly AuditPageDescriptor[];
  readonly confirm: boolean;
  readonly accessOnly?: boolean;
};
type Read = {
  readonly generation: number;
  readonly authority: AuditAuthority;
  readonly intent: Intent;
  readonly controller: AbortController;
  readonly inputs: AuditInputs;
  timer?: ReturnType<typeof setTimeout>;
};

export class AdministrativeAuditController {
  private state = initialAuditState();
  private listeners = new Set<() => void>();
  private generation = 0;
  private read: Read | null = null;
  private failed: Intent | null = null;
  private cursorRecoveryRequired = false;
  private visited = false;
  private disposed = false;
  constructor(private readonly ports: AuditPorts) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(
    change: Partial<AuditState>,
    announcement?: string,
    announcementRole: AuditState["announcementRole"] = "status",
  ) {
    this.state = {
      ...this.state,
      ...change,
      ...(announcement === undefined
        ? {}
        : {
            announcement,
            announcementRole,
            announcementSequence: this.state.announcementSequence + 1,
          }),
    };
    for (const listener of this.listeners) listener();
  }
  setAuthority(authority: AuditAuthority | null) {
    if (
      authority?.lifetime === this.state.authority?.lifetime &&
      authority?.actorId === this.state.authority?.actorId
    )
      return;
    this.retire();
    this.publish({ authority });
  }
  setActive(active: boolean) {
    if (this.disposed || this.state.active === active) return;
    this.cancel();
    this.publish({ active, activity: null });
    if (!active || !this.state.authority) return;
    const page = this.state.page;
    const descriptor =
      page && this.state.continuationValid
        ? page.descriptor
        : this.firstDescriptor();
    const confirm = this.visited;
    this.visited = true;
    // Recheck authority on return without retrying a rejected continuation.
    const accessOnly = this.cursorRecoveryRequired;
    this.start({
      kind: page ? "reactivate" : "initial",
      query: this.state.applied,
      descriptor,
      history: this.state.history,
      confirm: confirm || accessOnly,
      accessOnly,
    });
  }
  retire = () => {
    this.cancel();
    this.failed = null;
    this.cursorRecoveryRequired = false;
    this.visited = false;
    this.state = initialAuditState();
    for (const listener of this.listeners) listener();
  };
  dispose = () => {
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  };
  private cancel() {
    ++this.generation;
    const read = this.read;
    this.read = null;
    clearTimeout(read?.timer);
    read?.controller.abort();
  }
  edit = (field: AuditField, value: string) => {
    if (!this.available()) return;
    const errors = { ...this.state.fieldErrors };
    delete errors[field];
    this.publish({
      inputs: { ...this.state.inputs, [field]: value },
      fieldErrors: errors,
    });
  };
  apply = () => {
    if (!this.available()) return;
    const { query, errors } = validateAuditInputs(this.state.inputs);
    this.publish({ fieldErrors: errors });
    if (Object.keys(errors).length) {
      this.publish(
        {},
        "Filters were not applied. Check the highlighted fields.",
        "alert",
      );
      return;
    }
    if (sameAuditQuery(query, this.state.applied)) {
      this.refresh();
      return;
    }
    this.cancel();
    this.failed = null;
    this.cursorRecoveryRequired = false;
    this.publish({
      applied: query,
      page: null,
      history: [],
      expandedId: null,
      continuationValid: false,
      problem: null,
      scrollTop: 0,
      tableScrollLeft: 0,
    });
    this.start({
      kind: "initial",
      query,
      descriptor: this.firstDescriptor(),
      history: [],
      confirm: false,
    });
  };
  clearFilters = () => {
    if (!this.available()) return;
    this.publish({ inputs: emptyAuditQuery, fieldErrors: {} });
    this.apply();
  };
  refresh = () => {
    if (!this.available() || this.read?.intent.kind === "refresh") return;
    this.cursorRecoveryRequired = false;
    this.publish({ history: [], continuationValid: false });
    this.start({
      kind: "refresh",
      query: this.state.applied,
      descriptor: this.firstDescriptor(),
      history: [],
      confirm: false,
    });
  };
  next = () => {
    const { page, history, continuationValid } = this.state;
    if (
      !this.available() ||
      this.read ||
      !page ||
      !continuationValid ||
      !page.paging.has_more ||
      !page.paging.next_cursor
    )
      return;
    const descriptor = {
      ...page.descriptor,
      cursor: page.paging.next_cursor,
      number: page.descriptor.number + 1,
    };
    if (
      [...history, page.descriptor].some(
        (entry) => entry.cursor === descriptor.cursor,
      )
    ) {
      this.cursorFailure();
      return;
    }
    this.start({
      kind: "page",
      query: this.state.applied,
      descriptor,
      history: [...history, page.descriptor].slice(-auditHistoryLimit),
      confirm: false,
    });
  };
  previous = () => {
    const descriptor = this.state.history.at(-1);
    if (
      !this.available() ||
      this.read ||
      !this.state.continuationValid ||
      !descriptor
    )
      return;
    this.start({
      kind: "page",
      query: this.state.applied,
      descriptor,
      history: this.state.history.slice(0, -1),
      confirm: false,
    });
  };
  retry = () => {
    if (!this.available() || this.read || !this.failed) return;
    this.start(this.failed);
  };
  toggleExpanded = (id: string) => {
    if (
      !this.available() ||
      !this.state.page?.rows.some((event) => event.audit_event_id === id)
    )
      return;
    this.publish({ expandedId: this.state.expandedId === id ? null : id });
  };
  rememberScroll = (scrollTop: number, tableScrollLeft: number) => {
    if (this.state.active) this.publish({ scrollTop, tableScrollLeft });
  };
  private available() {
    return (
      !this.disposed &&
      this.state.active &&
      this.state.authority !== null &&
      this.ports.isCurrent(this.state.authority)
    );
  }
  private firstDescriptor(): AuditPageDescriptor {
    return {
      binding: this.state.authority
        ? auditBinding(this.state.authority, this.state.applied)
        : "",
      cursor: null,
      number: 1,
    };
  }
  private current(read: Read) {
    return (
      this.read === read &&
      read.generation === this.generation &&
      !read.controller.signal.aborted &&
      this.available() &&
      this.state.authority?.lifetime === read.authority.lifetime &&
      this.state.authority.actorId === read.authority.actorId
    );
  }
  private start(intent: Intent) {
    const authority = this.state.authority;
    if (!authority || !this.available()) return;
    if (
      intent.descriptor.binding !== auditBinding(authority, intent.query) ||
      !sameAuditQuery(intent.query, this.state.applied)
    ) {
      this.cursorFailure();
      return;
    }
    this.cancel();
    this.failed = null;
    const read: Read = {
      generation: this.generation,
      authority,
      intent,
      controller: new AbortController(),
      inputs: this.state.inputs,
    };
    this.read = read;
    read.timer = setTimeout(() => {
      if (!this.current(read)) return;
      this.fail(read, {
        kind: this.problemKind(read),
        message: "The audit read timed out. Try the read again.",
      });
    }, auditReadTimeout);
    const retained = this.state.page !== null;
    this.publish(
      {
        activity: intent.kind,
        problem: intent.accessOnly ? this.state.problem : null,
      },
      intent.accessOnly
        ? "Checking current audit access."
        : intent.kind === "page"
          ? `Loading audit page ${intent.descriptor.number}.`
          : retained
            ? "Refreshing administrative audit."
            : "Loading administrative audit.",
    );
    void this.execute(read).catch(() => {
      if (this.current(read))
        this.fail(read, {
          kind: this.problemKind(read),
          message:
            read.intent.kind === "page"
              ? "The requested audit page could not be loaded. Displayed events were kept. Try the page again."
              : this.state.page
                ? "Audit refresh failed. Displayed events may be stale. Try the read again."
                : "Administrative audit is unavailable. Try the read again.",
        });
    });
  }
  private async execute(read: Read) {
    if (read.intent.confirm) {
      const access = await this.ports.confirmAccess(
        read.authority,
        read.controller.signal,
        () => this.current(read),
      );
      if (!this.current(read)) return;
      if (access.kind === "access_lost" || access.kind === "session_lost") {
        this.deny(read, access.kind === "session_lost" ? 401 : 403);
        return;
      }
      if (access.kind !== "authorized") {
        this.fail(read, {
          kind: "access",
          message:
            "Current audit access could not be checked. Try the read again.",
        });
        return;
      }
    }
    if (!this.current(read)) return;
    if (read.intent.accessOnly) {
      this.cursorFailure();
      return;
    }
    const result = await this.ports.list({
      query: read.intent.query,
      limit: auditPageLimit,
      cursor: read.intent.descriptor.cursor,
      signal: read.controller.signal,
    });
    if (!this.current(read)) return;
    this.accept(read, result);
  }
  private accept(read: Read, result: AuditReadResult) {
    if (!result.ok) {
      if (result.status === 401 && result.error.code === "session_required") {
        this.deny(read, 401);
        return;
      }
      if (
        result.status === 403 &&
        result.error.code === "authorization_denied"
      ) {
        this.deny(read, 403);
        return;
      }
      if (result.error.code === "invalid_pagination_request") {
        this.cursorFailure();
        return;
      }
      this.fail(read, {
        kind:
          result.error.code === "invalid_list_query"
            ? "query"
            : this.problemKind(read),
        message: auditFailureMessage(
          result.error,
          result.status,
          this.state.page !== null,
          read.intent.kind === "page",
        ),
      });
      return;
    }
    if (
      result.paging.has_more &&
      read.intent.history.some(
        (entry) => entry.cursor === result.paging.next_cursor,
      )
    ) {
      this.cursorFailure();
      return;
    }
    const page = {
      descriptor: read.intent.descriptor,
      rows: result.rows,
      paging: result.paging,
    };
    this.cancel();
    this.failed = null;
    this.publish(
      {
        page,
        history: read.intent.history,
        continuationValid: true,
        activity: null,
        problem: null,
        expandedId: result.rows.some(
          (event) => event.audit_event_id === this.state.expandedId,
        )
          ? this.state.expandedId
          : null,
      },
      auditPageSummary(page),
    );
  }
  private problemKind(read: Read): "page" | "refresh" | "initial" {
    return read.intent.kind === "page"
      ? "page"
      : this.state.page
        ? "refresh"
        : "initial";
  }
  private fail(read: Read, problem: AuditProblem) {
    this.cancel();
    this.failed = read.intent;
    this.publish({ activity: null, problem }, problem.message);
  }
  private cursorFailure() {
    this.cancel();
    this.failed = null;
    this.cursorRecoveryRequired = true;
    const message =
      "This audit continuation is no longer usable. Reload the first page with the applied filters.";
    this.publish(
      {
        activity: null,
        problem: { kind: "cursor", message },
        continuationValid: false,
        history: [],
      },
      message,
    );
  }
  private deny(read: Read, status: 401 | 403) {
    this.retire();
    this.ports.authorizationFailed(status, read.authority);
  }
}

export const auditTransportPorts = { list: listAdministrativeAuditPage };
