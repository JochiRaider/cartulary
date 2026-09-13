import type { TimelineSourceRecordRequirement } from "./timelineViewportContinuityModel";

export const timelineFreshnessRetryLimit = 2;

export type TimelineLoadIdentity = {
  readonly incidentId: string;
  readonly queryIdentity: string;
  readonly surfaceIdentity: string;
};

export type TimelineLoadSubject = TimelineLoadIdentity & {
  readonly mutationEpoch: number;
  readonly requestGeneration: number;
  readonly sourceVersionObligation: TimelineSourceRecordRequirement | null;
};

export type TimelineLoadStatus =
  | "idle"
  | "initial_loading"
  | "refreshing"
  | "ready"
  | "stale_error"
  | "unavailable";

export type TimelineLoadState = {
  readonly activeSubject: TimelineLoadSubject | null;
  readonly identity: TimelineLoadIdentity;
  readonly latestMutationEpoch: number;
  readonly retryDepth: number;
  readonly status: TimelineLoadStatus;
};

export type TimelineLoadEvent =
  | {
      readonly hasLoadedRows: boolean;
      readonly identity: TimelineLoadIdentity;
      readonly kind: "subject_changed";
      readonly mutationEpoch: number;
    }
  | {
      readonly hasLoadedRows: boolean;
      readonly kind: "start";
      readonly retryDepth: number;
      readonly showLoading: boolean;
      readonly subject: TimelineLoadSubject;
    }
  | {
      readonly kind: "accepted_mutation";
      readonly mutationEpoch: number;
    }
  | {
      readonly kind: "success";
      readonly subject: TimelineLoadSubject;
    }
  | {
      readonly kind: "stale_result";
      readonly obligationSatisfied: boolean;
      readonly retryDepth: number;
      readonly retryable: boolean;
      readonly subject: TimelineLoadSubject;
    }
  | {
      readonly hasLoadedRows: boolean;
      readonly kind: "failure";
      readonly message: string;
      readonly subject: TimelineLoadSubject;
    }
  | {
      readonly kind: "access_loss";
      readonly message: string;
      readonly subject: TimelineLoadSubject;
    }
  | {
      readonly hasLoadedRows: boolean;
      readonly kind: "retry_exhaustion";
      readonly message: string;
      readonly subject: TimelineLoadSubject;
    };

export type TimelineLoadEffect =
  | { readonly kind: "request"; readonly subject: TimelineLoadSubject }
  | {
      readonly kind: "retry";
      readonly retryDepth: number;
      readonly subject: TimelineLoadSubject;
    }
  | { readonly kind: "commit"; readonly subject: TimelineLoadSubject }
  | {
      readonly hasLoadedRows: boolean;
      readonly kind: "publish_status";
      readonly message: string | null;
      readonly requestGeneration: number | null;
      readonly status: TimelineLoadStatus;
    }
  | {
      readonly kind: "clear_protected_rows";
      readonly subject: TimelineLoadSubject;
    }
  | {
      readonly kind: "settle_obligation";
      readonly subject: TimelineLoadSubject;
    }
  | {
      readonly kind: "fail_continuity";
      readonly subject: TimelineLoadSubject;
    };

export type TimelineLoadTransition = {
  readonly effects: readonly TimelineLoadEffect[];
  readonly state: TimelineLoadState;
};

function sourceObligationsEqual(
  left: TimelineSourceRecordRequirement | null,
  right: TimelineSourceRecordRequirement | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      left.recordId === right.recordId &&
      left.minimumRowVersion === right.minimumRowVersion)
  );
}

export function timelineLoadIdentitiesEqual(
  left: TimelineLoadIdentity,
  right: TimelineLoadIdentity,
): boolean {
  return (
    left.incidentId === right.incidentId &&
    left.surfaceIdentity === right.surfaceIdentity &&
    left.queryIdentity === right.queryIdentity
  );
}

export function timelineLoadSubjectsEqual(
  left: TimelineLoadSubject,
  right: TimelineLoadSubject,
): boolean {
  return (
    timelineLoadIdentitiesEqual(left, right) &&
    left.requestGeneration === right.requestGeneration &&
    left.mutationEpoch === right.mutationEpoch &&
    sourceObligationsEqual(
      left.sourceVersionObligation,
      right.sourceVersionObligation,
    )
  );
}

export function createTimelineLoadState(
  identity: TimelineLoadIdentity,
  mutationEpoch = 0,
): TimelineLoadState {
  return {
    activeSubject: null,
    identity,
    latestMutationEpoch: mutationEpoch,
    retryDepth: 0,
    status: "idle",
  };
}

function publishStatus(
  status: TimelineLoadStatus,
  hasLoadedRows: boolean,
  message: string | null = null,
  requestGeneration: number | null = null,
): TimelineLoadEffect {
  return {
    hasLoadedRows,
    kind: "publish_status",
    message,
    requestGeneration,
    status,
  };
}

function transitionSubjectChange(
  state: TimelineLoadState,
  event: Extract<TimelineLoadEvent, { readonly kind: "subject_changed" }>,
): TimelineLoadTransition {
  if (
    timelineLoadIdentitiesEqual(state.identity, event.identity) &&
    state.latestMutationEpoch === event.mutationEpoch
  ) {
    return { effects: [], state };
  }
  const next = createTimelineLoadState(event.identity, event.mutationEpoch);
  return {
    effects: [publishStatus("idle", event.hasLoadedRows)],
    state: next,
  };
}

function transitionStart(
  state: TimelineLoadState,
  event: Extract<TimelineLoadEvent, { readonly kind: "start" }>,
): TimelineLoadTransition {
  if (!timelineLoadIdentitiesEqual(state.identity, event.subject)) {
    return { effects: [], state };
  }
  const status = event.hasLoadedRows
    ? "refreshing"
    : event.showLoading
      ? "initial_loading"
      : "idle";
  const next: TimelineLoadState = {
    ...state,
    activeSubject: event.subject,
    latestMutationEpoch: Math.max(
      state.latestMutationEpoch,
      event.subject.mutationEpoch,
    ),
    retryDepth: event.retryDepth,
    status,
  };
  return {
    effects: [
      publishStatus(
        status,
        event.hasLoadedRows,
        null,
        event.subject.requestGeneration,
      ),
      { kind: "request", subject: event.subject },
    ],
    state: next,
  };
}

function subjectIsActive(
  state: TimelineLoadState,
  subject: TimelineLoadSubject,
): boolean {
  return (
    state.activeSubject !== null &&
    timelineLoadSubjectsEqual(state.activeSubject, subject)
  );
}

function transitionStaleResult(
  state: TimelineLoadState,
  event: Extract<TimelineLoadEvent, { readonly kind: "stale_result" }>,
): TimelineLoadTransition {
  if (event.obligationSatisfied) {
    const active = subjectIsActive(state, event.subject);
    return {
      effects: active
        ? [
            { kind: "settle_obligation", subject: event.subject },
            publishStatus("ready", true),
          ]
        : [{ kind: "settle_obligation", subject: event.subject }],
      state: active ? { ...state, status: "ready" } : state,
    };
  }
  if (!subjectIsActive(state, event.subject)) {
    return { effects: [], state };
  }
  if (!event.retryable || event.retryDepth >= timelineFreshnessRetryLimit) {
    return { effects: [], state };
  }
  const retryDepth = event.retryDepth + 1;
  return {
    effects: [{ kind: "retry", retryDepth, subject: event.subject }],
    state: { ...state, retryDepth },
  };
}

export function transitionTimelineLoad(
  state: TimelineLoadState,
  event: TimelineLoadEvent,
): TimelineLoadTransition {
  switch (event.kind) {
    case "subject_changed":
      return transitionSubjectChange(state, event);
    case "start":
      return transitionStart(state, event);
    case "accepted_mutation":
      if (event.mutationEpoch <= state.latestMutationEpoch) {
        return { effects: [], state };
      }
      return {
        effects: [],
        state: { ...state, latestMutationEpoch: event.mutationEpoch },
      };
    case "success":
      if (
        !subjectIsActive(state, event.subject) ||
        event.subject.mutationEpoch < state.latestMutationEpoch
      ) {
        return { effects: [], state };
      }
      return {
        effects: [
          { kind: "commit", subject: event.subject },
          publishStatus("ready", true),
        ],
        state: { ...state, status: "ready" },
      };
    case "stale_result":
      return transitionStaleResult(state, event);
    case "failure":
      if (!subjectIsActive(state, event.subject)) {
        return { effects: [], state };
      }
      return {
        effects: [
          { kind: "fail_continuity", subject: event.subject },
          publishStatus("stale_error", event.hasLoadedRows, event.message),
        ],
        state: { ...state, status: "stale_error" },
      };
    case "access_loss":
      if (!subjectIsActive(state, event.subject)) {
        return { effects: [], state };
      }
      return {
        effects: [
          { kind: "fail_continuity", subject: event.subject },
          { kind: "clear_protected_rows", subject: event.subject },
          publishStatus("unavailable", false, event.message),
        ],
        state: { ...state, status: "unavailable" },
      };
    case "retry_exhaustion":
      if (!subjectIsActive(state, event.subject)) {
        return { effects: [], state };
      }
      return {
        effects: [
          {
            kind: "fail_continuity",
            subject: event.subject,
          },
          publishStatus("stale_error", event.hasLoadedRows, event.message),
        ],
        state: { ...state, status: "stale_error" },
      };
  }
}
