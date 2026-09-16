import {
  boundedRead,
  browserObservationClock,
  type ObservationClock,
} from "../services/asyncObservation";
import type {
  NetworkFlowIndicatorLinkResult,
  NetworkFlowIndicatorTarget,
} from "../services/networkFlowContractAdapter";
import {
  coreAtomicIPType,
  type IndicatorLinkTargetPort,
} from "../services/networkFlowIndicatorAdapter";
import { IndicatorLinkTargetDiscovery } from "./IndicatorLinkTargetDiscovery";
import type { NetworkFlowExtensionResourceChange } from "./networkFlowCollaborationInterpreter";
import {
  canLinkIndicator,
  captureIndicatorLinkAttempt,
  type IndicatorLinkAttempt,
  type IndicatorLinkAuthority,
  type IndicatorLinkFeedback,
  IndicatorLinkWriteError,
  type NetworkFlowIndicatorLinkCandidate,
  sameIndicatorLinkAuthority,
} from "./networkFlowIndicatorLinkOperation";

export type IndicatorLinkDraft = {
  readonly workId: number;
  readonly candidate: NetworkFlowIndicatorLinkCandidate;
  readonly targetMode: "create_indicator" | "existing_indicator";
  readonly existingId: string;
  readonly confirmation: string;
  readonly revision: number;
  readonly applicable: boolean;
  readonly feedback: IndicatorLinkFeedback | null;
};

export type IndicatorLinkSettlement =
  | { readonly kind: "queued" | "pending"; readonly replay: boolean }
  | {
      readonly kind: "not_dispatched" | "rejected" | "uncertain";
      readonly replay: boolean;
      readonly feedback: IndicatorLinkFeedback;
    }
  | {
      readonly kind: "confirmed" | "reused";
      readonly replay: boolean;
      readonly receipt: NetworkFlowIndicatorLinkResult;
    };

export type IndicatorLinkSnapshot = {
  readonly draft: IndicatorLinkDraft | null;
  readonly attempt: IndicatorLinkAttempt | null;
  readonly settlement: IndicatorLinkSettlement | null;
  readonly presentation: "draft" | "attempt" | null;
  readonly hidden: boolean;
  readonly writable: boolean;
  readonly sourceLimit: number;
  readonly limitError: string | null;
  readonly status: "link_pending" | "link_committed" | null;
};

export type IndicatorLinkTransport = {
  readonly discoverTargets: IndicatorLinkTargetPort;
  readonly sourceLimit: (signal: AbortSignal) => Promise<number>;
  readonly submit: (
    attempt: IndicatorLinkAttempt,
    signal: AbortSignal,
    authorizeDispatch: () => void,
  ) => Promise<NetworkFlowIndicatorLinkResult>;
};

type LinkDispatch = {
  readonly attempt: IndicatorLinkAttempt;
  readonly authority: IndicatorLinkAuthority;
  readonly request: AbortController;
  readonly replay: boolean;
  readonly wasUncertain: boolean;
  dispatched: boolean;
  cancelDeadline: () => void;
};

const initial = (): IndicatorLinkSnapshot => ({
  draft: null,
  attempt: null,
  settlement: null,
  presentation: null,
  hidden: false,
  writable: false,
  sourceLimit: 0,
  limitError: null,
  status: null,
});
const recoveryMessage =
  "The link may have committed. Replay this exact request to recover its receipt; a new transaction ID would be a different request.";

/** One workbook-lifetime owner. Presentation never owns a dispatched write. */
export class NetworkFlowIndicatorLinkController {
  readonly targets = new IndicatorLinkTargetDiscovery();
  private state = initial();
  private listeners = new Set<() => void>();
  private transport: IndicatorLinkTransport | null = null;
  private authorityReader: (() => IndicatorLinkAuthority) | null = null;
  private authority: IndicatorLinkAuthority | null = null;
  private retainedScope: {
    readonly incidentId: string;
    readonly actorId: string;
  } | null = null;
  private blockedAuthority: IndicatorLinkAuthority | null = null;
  private dispatch: LinkDispatch | null = null;
  private limitRequest: AbortController | null = null;
  private cancelStatus = () => {};
  private selectionContext = "";
  private selectionRevision = 0;
  private draftRevision = 0;
  private workSequence = 0;
  private admission = false;
  private active = false;
  private focusRestorer: (() => Promise<boolean>) | null = null;
  bindFocusRestoration(restore: () => Promise<boolean>): () => void {
    this.focusRestorer = restore;
    return () => {
      if (this.focusRestorer === restore) this.focusRestorer = null;
    };
  }
  readonly restoreFocus = async (): Promise<boolean> =>
    this.focusRestorer?.() ?? false;
  private removedSources = new Set<string>();

  constructor(
    private readonly clock: ObservationClock = browserObservationClock,
  ) {}
  readonly getSnapshot = (): IndicatorLinkSnapshot => this.state;
  readonly subscribe = (listener: () => void): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };
  private update(patch: Partial<IndicatorLinkSnapshot>): void {
    this.state = { ...this.state, ...patch };
    this.targets.revalidate();
    for (const listener of this.listeners) listener();
  }

  bind(
    transport: IndicatorLinkTransport,
    authority: () => IndicatorLinkAuthority,
  ): void {
    this.transport = transport;
    this.authorityReader = authority;
    this.targets.bind(transport.discoverTargets, () => {
      const current = this.authorityReader?.();
      const draft = this.state.draft;
      if (
        current === undefined ||
        this.state.hidden ||
        !current.sessionResolved ||
        current.actorId === null ||
        !current.available ||
        draft === null ||
        !draft.applicable ||
        draft.targetMode !== "existing_indicator"
      )
        return null;
      return {
        incidentId: current.incidentId,
        candidateValue: draft.candidate.candidateValue,
        key: JSON.stringify([
          current.incidentId,
          current.actorId,
          current.session,
          current.role,
          current.open,
          current.availabilityTag?.epochId,
          current.availabilityTag?.generation.toString(),
          draft.candidate.key,
          this.selectionRevision,
        ]),
      };
    });
    this.revalidateAuthority();
  }
  readonly activate = (): (() => void) => {
    this.active = true;
    this.loadLimit();
    return () => {
      this.active = false;
    };
  };
  readonly revalidateAuthority = (): void => {
    const next = this.authorityReader?.();
    if (next === undefined) return;
    const previous = this.authority;
    if (previous !== null && sameIndicatorLinkAuthority(previous, next)) return;
    if (
      this.retainedScope !== null &&
      (this.retainedScope.incidentId !== next.incidentId ||
        (next.actorId !== null && this.retainedScope.actorId !== next.actorId))
    )
      this.purge();
    if (
      previous !== null &&
      (previous.incidentId !== next.incidentId ||
        (next.actorId !== null &&
          previous.actorId !== null &&
          previous.actorId !== next.actorId))
    )
      this.purge();
    this.authority = next;
    if (
      this.blockedAuthority !== null &&
      !sameIndicatorLinkAuthority(this.blockedAuthority, next)
    )
      this.blockedAuthority = null;
    if (next.actorId !== null && next.sessionResolved)
      this.retainedScope = {
        incidentId: next.incidentId,
        actorId: next.actorId,
      };
    this.limitRequest?.abort();
    this.limitRequest = null;
    this.stopDispatch(
      "Authority changed. Review current access before continuing.",
    );
    this.clearStatus();
    if (!next.sessionResolved || next.actorId === null) {
      this.update({ hidden: true, writable: false });
      this.invalidateDraft(
        "Session recovery is required before linking.",
        "denied",
      );
      return;
    }
    const capturedActor = this.state.attempt?.authority.actorId;
    if (
      (capturedActor !== undefined && capturedActor !== next.actorId) ||
      !next.profileAvailable ||
      !["viewer", "editor", "reviewer", "admin"].includes(next.role ?? "")
    ) {
      this.purge();
      return;
    }
    this.update({
      hidden: false,
      writable: canLinkIndicator(next),
      sourceLimit: 0,
    });
    if (previous !== null)
      this.invalidateDraft(
        canLinkIndicator(next)
          ? "Authority changed. Select the endpoint again and confirm its exact value."
          : "Linking is unavailable with the current role or incident state. The draft is available to copy.",
        "denied",
      );
    this.loadLimit();
  };

  readonly loadLimit = (): void => {
    const authority = this.authorityReader?.();
    const transport = this.transport;
    if (
      !this.active ||
      authority === undefined ||
      transport === null ||
      !canLinkIndicator(authority) ||
      this.limitRequest !== null ||
      this.state.sourceLimit > 0
    )
      return;
    const request = new AbortController();
    this.limitRequest = request;
    void boundedRead(transport.sourceLimit, request.signal)
      .then((limit) => {
        if (
          request.signal.aborted ||
          this.limitRequest !== request ||
          !this.currentAuthority(authority)
        )
          return;
        if (!Number.isInteger(limit) || limit < 1 || limit > 1000)
          throw new Error("invalid_link_limit");
        this.update({ sourceLimit: limit, limitError: null });
      })
      .catch(() => {
        if (!request.signal.aborted && this.limitRequest === request)
          this.update({
            sourceLimit: 0,
            limitError:
              "Link limits could not be loaded. Retry loading limits before selecting an endpoint.",
          });
      })
      .finally(() => {
        if (this.limitRequest === request) this.limitRequest = null;
      });
  };

  readonly setSelectionContext = (context: string): void => {
    if (context === this.selectionContext) return;
    this.selectionContext = context;
    this.selectionRevision++;
    this.clearStatus();
    this.invalidateDraft(
      "The source selection or query changed. Select the endpoint again before confirming.",
      "stale",
    );
  };

  readonly openDraft = (candidate: NetworkFlowIndicatorLinkCandidate): void => {
    this.revalidateAuthority();
    const authority = this.authorityReader?.();
    if (
      this.state.hidden ||
      authority === undefined ||
      !authority.sessionResolved ||
      authority.actorId === null ||
      !authority.profileAvailable ||
      !["viewer", "editor", "reviewer", "admin"].includes(
        authority.role ?? "",
      ) ||
      (this.blockedAuthority !== null &&
        sameIndicatorLinkAuthority(authority, this.blockedAuthority))
    )
      return;
    this.selectionRevision++;
    this.clearStatus();
    this.update({
      draft: {
        candidate: structuredClone(candidate),
        workId: ++this.workSequence,
        targetMode: "create_indicator",
        existingId: "",
        confirmation: "",
        revision: ++this.draftRevision,
        applicable: true,
        feedback: null,
      },
      presentation: "draft",
    });
  };
  readonly editDraft = (
    patch: Partial<
      Pick<IndicatorLinkDraft, "targetMode" | "existingId" | "confirmation">
    >,
  ): void => {
    const draft = this.state.draft;
    if (draft === null) return;
    const targetChanged =
      (patch.targetMode !== undefined &&
        patch.targetMode !== draft.targetMode) ||
      (patch.existingId !== undefined && patch.existingId !== draft.existingId);
    if (targetChanged || this.unresolved()) this.clearStatus();
    this.update({
      draft: {
        ...draft,
        ...patch,
        confirmation: targetChanged
          ? ""
          : (patch.confirmation ?? draft.confirmation),
        revision: ++this.draftRevision,
        feedback: draft.applicable ? null : draft.feedback,
      },
    });
  };
  readonly dismiss = (): void => {
    this.update({ presentation: null });
  };
  readonly reopen = (): void => {
    if (this.state.attempt !== null && !this.state.hidden)
      this.update({ presentation: "attempt" });
  };
  readonly showDraft = (): void => {
    const { draft, attempt, settlement } = this.state;
    if (draft === null || this.state.hidden) return;
    const feedback =
      attempt?.draftRevision === draft.revision &&
      settlement !== null &&
      "feedback" in settlement
        ? settlement.feedback
        : draft.feedback;
    this.update({ presentation: "draft", draft: { ...draft, feedback } });
  };
  readonly done = (): void => {
    if (
      this.state.settlement?.kind !== "confirmed" &&
      this.state.settlement?.kind !== "reused"
    )
      return;
    this.stopDispatch();
    this.update({
      draft:
        this.state.draft?.revision === this.state.attempt?.draftRevision
          ? null
          : this.state.draft,
      attempt: null,
      settlement: null,
      presentation: null,
    });
  };
  /** Explicit local abandonment does not imply a server rollback. */
  readonly abandonRecovery = (): void => {
    this.stopDispatch();
    this.admission = false;
    this.update({
      attempt: null,
      settlement: null,
      presentation: this.state.draft === null ? null : "draft",
    });
  };

  readonly submit = (): boolean => {
    this.revalidateAuthority();
    const draft = this.state.draft;
    const authority = this.authorityReader?.();
    if (draft === null || authority === undefined || this.admission)
      return false;
    if (this.unresolved())
      return this.rejectDraft(
        "Recover or explicitly forget the previous link request before submitting another link.",
        null,
        "recovery",
      );
    if (!this.canWrite(authority))
      return this.rejectDraft(
        "The current role or incident state does not permit linking.",
        null,
        "denied",
      );
    if (!draft.applicable || this.sourcesRemoved(draft.candidate))
      return this.rejectDraft(
        "The captured source is stale. Select the endpoint again.",
        null,
        "stale",
      );
    if (draft.confirmation !== draft.candidate.candidateValue)
      return this.rejectDraft(
        "Enter the canonical candidate exactly as shown, without changing spacing or spelling.",
        "confirmation",
      );
    const indicatorType = coreAtomicIPType(draft.candidate.candidateValue);
    if (indicatorType === null)
      return this.rejectDraft(
        "Core cannot represent this endpoint as an atomic IP indicator.",
        "target",
      );
    if (
      this.state.sourceLimit < 1 ||
      draft.candidate.sourceRefs.length > this.state.sourceLimit
    )
      return this.rejectDraft(
        "The source-reference limit is unavailable or exceeded.",
        null,
      );
    if (
      draft.targetMode === "existing_indicator" &&
      !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/iu.test(
        draft.existingId,
      )
    )
      return this.rejectDraft(
        "Select a compatible indicator or enter its complete indicator ID.",
        "target",
      );
    const target: NetworkFlowIndicatorTarget =
      draft.targetMode === "create_indicator"
        ? { mode: "create_indicator", indicator_type: indicatorType }
        : { mode: "existing_indicator", indicator_id: draft.existingId };
    // Acquire before ID generation, subscribers, promises or a queued callback.
    this.admission = true;
    let attempt: IndicatorLinkAttempt;
    try {
      attempt = captureIndicatorLinkAttempt({
        authority,
        candidate: draft.candidate,
        selectionRevision: this.selectionRevision,
        draftRevision: draft.revision,
        workId: draft.workId,
        sourceLimit: this.state.sourceLimit,
        target,
        confirmation: draft.confirmation,
      });
    } catch {
      this.admission = false;
      return this.rejectDraft(
        "A secure request identifier could not be generated. Nothing was submitted.",
        null,
      );
    }
    this.clearStatus();
    this.update({ attempt, presentation: "attempt" });
    this.send(attempt, false, false);
    return true;
  };

  readonly replay = (): boolean => {
    this.revalidateAuthority();
    const attempt = this.state.attempt;
    const authority = this.authorityReader?.();
    const settlement = this.state.settlement;
    if (
      this.admission ||
      attempt === null ||
      authority === undefined ||
      settlement === null ||
      !["uncertain", "not_dispatched", "rejected"].includes(settlement.kind)
    )
      return false;
    if (
      !this.canWrite(authority) ||
      authority.actorId !== attempt.authority.actorId ||
      authority.incidentId !== attempt.authority.incidentId ||
      this.sourcesRemoved(attempt.candidate)
    )
      return false;
    this.admission = true;
    this.send(attempt, true, settlement.kind === "uncertain");
    return true;
  };

  private send(
    attempt: IndicatorLinkAttempt,
    replay: boolean,
    wasUncertain: boolean,
  ): void {
    const authority = this.authorityReader?.();
    const transport = this.transport;
    if (authority === undefined || transport === null) {
      this.admission = false;
      return;
    }
    this.stopDispatch();
    this.admission = true;
    const run: LinkDispatch = {
      attempt,
      authority: structuredClone(authority),
      request: new AbortController(),
      replay,
      wasUncertain,
      dispatched: false,
      cancelDeadline: () => {},
    };
    this.dispatch = run;
    this.update({
      settlement: { kind: "queued", replay },
      status:
        this.selectionRevision === attempt.selectionRevision &&
        this.state.draft?.revision === attempt.draftRevision
          ? "link_pending"
          : null,
    });
    run.cancelDeadline = this.clock.schedule(() => {
      if (this.dispatch !== run) return;
      run.request.abort();
      this.admission = false;
      this.update({
        settlement: {
          kind: run.dispatched || wasUncertain ? "uncertain" : "not_dispatched",
          replay,
          feedback: {
            kind: "recovery",
            field: null,
            message:
              run.dispatched || wasUncertain
                ? recoveryMessage
                : "The request was not dispatched before the observation window ended. Review access and retry this captured request.",
          },
        },
        status: null,
      });
    }, 30_000);
    const authorizeDispatch = () => {
      this.revalidateAuthority();
      const current = this.authorityReader?.();
      if (
        this.dispatch !== run ||
        this.state.attempt !== attempt ||
        run.request.signal.aborted ||
        current === undefined ||
        !sameIndicatorLinkAuthority(run.authority, current) ||
        !this.canWrite(current) ||
        this.sourcesRemoved(attempt.candidate) ||
        (!replay &&
          (this.selectionRevision !== attempt.selectionRevision ||
            this.state.draft?.revision !== attempt.draftRevision ||
            !this.state.draft.applicable))
      )
        throw new IndicatorLinkWriteError("not_dispatched", {
          kind: "stale",
          field: null,
          message:
            "The source, confirmation or authority changed before dispatch. Nothing was sent by this invocation.",
        });
      run.dispatched = true;
      this.update({
        settlement: { kind: "pending", replay },
        status:
          this.selectionRevision === attempt.selectionRevision
            ? "link_pending"
            : null,
      });
    };
    void Promise.resolve()
      .then(() =>
        transport.submit(attempt, run.request.signal, authorizeDispatch),
      )
      .then((receipt) => {
        if (!this.accepts(run)) return;
        run.cancelDeadline();
        this.admission = false;
        this.dispatch = null;
        this.update({
          settlement: {
            kind: receipt.duplicate ? "reused" : "confirmed",
            replay,
            receipt,
          },
        });
        if (
          this.selectionRevision === attempt.selectionRevision &&
          this.state.draft?.revision === attempt.draftRevision
        ) {
          this.update({ status: "link_committed" });
          this.cancelStatus = this.clock.schedule(
            () => this.update({ status: null }),
            5_000,
          );
        }
      })
      .catch((caught: unknown) => {
        if (!this.accepts(run)) return;
        run.cancelDeadline();
        this.admission = false;
        this.dispatch = null;
        const error =
          caught instanceof IndicatorLinkWriteError
            ? caught
            : new IndicatorLinkWriteError(
                run.dispatched ? "uncertain" : "not_dispatched",
                {
                  kind: "recovery",
                  field: null,
                  message: run.dispatched
                    ? recoveryMessage
                    : "The request was not dispatched. Review current access before retrying.",
                },
              );
        const uncertain = wasUncertain || error.certainty === "uncertain";
        this.update({
          settlement: {
            kind: uncertain ? "uncertain" : error.certainty,
            replay,
            feedback:
              uncertain && error.certainty !== "uncertain"
                ? {
                    ...error.feedback,
                    message: `${error.feedback.message} The original request remains uncertain. ${recoveryMessage}`,
                  }
                : error.feedback,
          },
          status: null,
        });
      });
  }

  private accepts(run: LinkDispatch): boolean {
    return (
      this.dispatch === run &&
      this.state.attempt === run.attempt &&
      this.currentAuthority(run.authority)
    );
  }
  private canWrite(authority: IndicatorLinkAuthority): boolean {
    return (
      canLinkIndicator(authority) &&
      (this.blockedAuthority === null ||
        !sameIndicatorLinkAuthority(authority, this.blockedAuthority))
    );
  }
  private currentAuthority(authority: IndicatorLinkAuthority): boolean {
    const current = this.authorityReader?.();
    return (
      current !== undefined && sameIndicatorLinkAuthority(authority, current)
    );
  }
  private unresolved(): boolean {
    return ["queued", "pending", "uncertain"].includes(
      this.state.settlement?.kind ?? "",
    );
  }
  private sourcesRemoved(
    candidate: NetworkFlowIndicatorLinkCandidate,
  ): boolean {
    return candidate.sourceTableIds.some((id) => this.removedSources.has(id));
  }
  private rejectDraft(
    message: string,
    field: IndicatorLinkFeedback["field"],
    kind: IndicatorLinkFeedback["kind"] = "validation",
  ): false {
    if (this.state.draft !== null)
      this.update({
        draft: { ...this.state.draft, feedback: { kind, message, field } },
      });
    return false;
  }
  private invalidateDraft(
    message: string,
    kind: IndicatorLinkFeedback["kind"],
  ): void {
    if (this.state.draft !== null)
      this.update({
        draft: {
          ...this.state.draft,
          applicable: false,
          confirmation: "",
          revision: ++this.draftRevision,
          feedback: { kind, message, field: null },
        },
      });
  }
  readonly onMutationAdmitted = (): void => {
    if (this.state.status === "link_committed") this.clearStatus();
  };
  private clearStatus(): void {
    this.cancelStatus();
    this.cancelStatus = () => {};
    if (this.state.status !== null) this.update({ status: null });
  }
  private stopDispatch(message?: string): void {
    const run = this.dispatch;
    if (run === null) return;
    run.cancelDeadline();
    run.request.abort();
    this.dispatch = null;
    this.admission = false;
    if (message !== undefined)
      this.update({
        settlement: {
          kind:
            run.dispatched || run.wasUncertain ? "uncertain" : "not_dispatched",
          replay: run.replay,
          feedback: {
            kind: "denied",
            field: null,
            message:
              run.dispatched || run.wasUncertain
                ? `${message} ${recoveryMessage}`
                : `${message} This invocation was not dispatched.`,
          },
        },
        status: null,
      });
  }
  readonly onResourceChange = (
    change: NetworkFlowExtensionResourceChange,
  ): void => {
    if (
      [
        "authorization_lost",
        "claim_withdrawn",
        "session_revoked",
        "incident_closed",
      ].includes(change.reasonCode)
    )
      this.blockedAuthority = this.authorityReader?.() ?? null;
    if (
      change.reasonCode === "authorization_lost" ||
      change.reasonCode === "claim_withdrawn"
    ) {
      this.purge();
      return;
    }
    if (change.reasonCode === "session_revoked") {
      this.stopDispatch("Session recovery is required.");
      this.clearStatus();
      this.update({ hidden: true, writable: false });
      this.invalidateDraft("Session recovery is required.", "denied");
      return;
    }
    if (change.reasonCode === "incident_closed") {
      this.stopDispatch("The incident is closed.");
      this.clearStatus();
      this.invalidateDraft(
        "The incident is closed. This draft is available to copy.",
        "denied",
      );
      this.update({ writable: false });
      return;
    }
    if (
      change.resourceKind === "network_flow_graph_view" ||
      (change.resourceKind === "network_flow_table" &&
        change.reasonCode === "renamed")
    )
      return;
    if (
      change.changeKind === "remove" &&
      change.resourceKind === "network_flow_table"
    ) {
      this.removedSources.add(change.resourceId);
      const attempt = this.state.attempt;
      const draft = this.state.draft;
      if (
        (attempt === null || !this.sourcesRemoved(attempt.candidate)) &&
        (draft === null || !this.sourcesRemoved(draft.candidate))
      )
        return;
      if (attempt !== null && this.sourcesRemoved(attempt.candidate)) {
        this.stopDispatch();
        this.update({
          attempt: null,
          settlement: null,
          presentation:
            this.state.presentation === "attempt"
              ? null
              : this.state.presentation,
        });
      }
      if (draft !== null && this.sourcesRemoved(draft.candidate))
        this.update({
          draft: null,
          presentation:
            this.state.presentation === "draft"
              ? null
              : this.state.presentation,
        });
    } else {
      this.selectionRevision++;
      this.stopDispatch("Source information changed.");
      this.invalidateDraft(
        "Source information changed. Select the endpoint again.",
        "stale",
      );
    }
    this.clearStatus();
  };
  private purge(): void {
    this.stopDispatch();
    this.limitRequest?.abort();
    this.limitRequest = null;
    this.clearStatus();
    this.removedSources.clear();
    this.selectionRevision++;
    this.admission = false;
    this.state = initial();
    this.targets.clear();
    for (const listener of this.listeners) listener();
  }
  readonly onTargetChange = (change: {
    readonly id: string;
    readonly removed: boolean;
    readonly clientTxnId: string;
  }): void => {
    this.targets.clear();
    const { attempt, draft, settlement } = this.state;
    if (attempt?.request.client_txn_id === change.clientTxnId) return;
    const capturedTarget =
      attempt?.request.target.mode === "existing_indicator"
        ? attempt.request.target.indicator_id
        : settlement !== null && "receipt" in settlement
          ? settlement.receipt.binding.target_indicator_ref.indicator_id
          : null;
    if (capturedTarget === change.id && change.removed) {
      this.stopDispatch();
      this.clearStatus();
      this.update({
        attempt: null,
        settlement: null,
        presentation:
          this.state.presentation === "attempt"
            ? null
            : this.state.presentation,
      });
    }
    if (
      draft?.targetMode === "existing_indicator" &&
      draft.existingId === change.id
    ) {
      this.clearStatus();
      if (change.removed)
        this.update({
          draft: null,
          presentation:
            this.state.presentation === "draft"
              ? null
              : this.state.presentation,
        });
      else
        this.invalidateDraft(
          "The selected indicator changed. Select the endpoint and review its target again.",
          "stale",
        );
    }
  };
  readonly dispose = (): void => {
    this.purge();
    this.active = false;
  };
}
