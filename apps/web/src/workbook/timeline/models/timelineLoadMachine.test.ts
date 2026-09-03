import { describe, expect, it } from "vitest";
import {
  createTimelineLoadState,
  type TimelineLoadIdentity,
  type TimelineLoadSubject,
  timelineFreshnessRetryLimit,
  timelineLoadSubjectsEqual,
  transitionTimelineLoad,
} from "./timelineLoadMachine";

const identity: TimelineLoadIdentity = {
  incidentId: "incident-1",
  queryIdentity: "query-1",
  surfaceIdentity: "view_schema:cartulary.view.timeline.v2",
};

function subject(
  overrides: Partial<TimelineLoadSubject> = {},
): TimelineLoadSubject {
  return {
    ...identity,
    mutationEpoch: 4,
    requestGeneration: 7,
    sourceVersionObligation: null,
    ...overrides,
  };
}

describe("Timeline load machine", () => {
  it("keys subjects by incident, surface, query, generation, mutation epoch, and source obligation", () => {
    const base = subject();
    expect(timelineLoadSubjectsEqual(base, { ...base })).toBe(true);
    for (const changed of [
      { incidentId: "incident-2" },
      { surfaceIdentity: "saved_view:view-1" },
      { queryIdentity: "query-2" },
      { requestGeneration: 8 },
      { mutationEpoch: 5 },
      {
        sourceVersionObligation: {
          minimumRowVersion: 9,
          recordId: "record-1",
        },
      },
    ]) {
      expect(timelineLoadSubjectsEqual(base, { ...base, ...changed })).toBe(
        false,
      );
    }
  });

  it("publishes initial and refresh starts and commits only the complete active subject", () => {
    const initial = createTimelineLoadState(identity, 4);
    const started = transitionTimelineLoad(initial, {
      hasLoadedRows: false,
      kind: "start",
      retryDepth: 0,
      showLoading: true,
      subject: subject(),
    });
    expect(started.state.status).toBe("initial_loading");
    expect(started.effects.map((effect) => effect.kind)).toEqual([
      "publish_status",
      "request",
    ]);
    expect(
      transitionTimelineLoad(started.state, {
        kind: "success",
        subject: subject({ requestGeneration: 8 }),
      }).effects,
    ).toEqual([]);
    expect(
      transitionTimelineLoad(started.state, {
        kind: "success",
        subject: subject(),
      }).effects.map((effect) => effect.kind),
    ).toEqual(["commit", "publish_status"]);

    const refresh = transitionTimelineLoad(initial, {
      hasLoadedRows: true,
      kind: "start",
      retryDepth: 0,
      showLoading: false,
      subject: subject(),
    });
    expect(refresh.state.status).toBe("refreshing");
  });

  it("tracks accepted mutations and refuses a pre-mutation success", () => {
    const started = transitionTimelineLoad(createTimelineLoadState(identity), {
      hasLoadedRows: true,
      kind: "start",
      retryDepth: 0,
      showLoading: false,
      subject: subject({ mutationEpoch: 4 }),
    });
    const mutated = transitionTimelineLoad(started.state, {
      kind: "accepted_mutation",
      mutationEpoch: 5,
    });
    expect(mutated.state.latestMutationEpoch).toBe(5);
    expect(
      transitionTimelineLoad(mutated.state, {
        kind: "success",
        subject: subject({ mutationEpoch: 4 }),
      }).effects,
    ).toEqual([]);
  });

  it("joins a satisfied obligation and otherwise retries exactly twice", () => {
    const obligation = {
      minimumRowVersion: 9,
      recordId: "record-1",
    };
    const active = transitionTimelineLoad(createTimelineLoadState(identity), {
      hasLoadedRows: true,
      kind: "start",
      retryDepth: 0,
      showLoading: false,
      subject: subject({ sourceVersionObligation: obligation }),
    }).state;
    expect(
      transitionTimelineLoad(active, {
        kind: "stale_result",
        obligationSatisfied: true,
        retryDepth: 0,
        retryable: true,
        subject: subject({ sourceVersionObligation: obligation }),
      }).effects.map((effect) => effect.kind),
    ).toEqual(["settle_obligation", "publish_status"]);

    let state = active;
    for (let depth = 1; depth <= timelineFreshnessRetryLimit; depth += 1) {
      const transition = transitionTimelineLoad(state, {
        kind: "stale_result",
        obligationSatisfied: false,
        retryDepth: depth - 1,
        retryable: true,
        subject: subject({ sourceVersionObligation: obligation }),
      });
      expect(transition.effects).toMatchObject([
        { kind: "retry", retryDepth: depth },
      ]);
      state = transition.state;
    }
    expect(
      transitionTimelineLoad(state, {
        kind: "stale_result",
        obligationSatisfied: false,
        retryDepth: timelineFreshnessRetryLimit,
        retryable: true,
        subject: subject({ sourceVersionObligation: obligation }),
      }).effects,
    ).toEqual([]);
  });

  it("ignores retry and exhaustion effects from a superseded complete subject", () => {
    const activeSubject = subject({ requestGeneration: 8 });
    const active = transitionTimelineLoad(createTimelineLoadState(identity), {
      hasLoadedRows: true,
      kind: "start",
      retryDepth: 0,
      showLoading: false,
      subject: activeSubject,
    }).state;
    const supersededSubject = subject({ requestGeneration: 7 });
    expect(
      transitionTimelineLoad(active, {
        kind: "stale_result",
        obligationSatisfied: false,
        retryDepth: 0,
        retryable: true,
        subject: supersededSubject,
      }).effects,
    ).toEqual([]);
    expect(
      transitionTimelineLoad(active, {
        hasLoadedRows: true,
        kind: "retry_exhaustion",
        message: "did not converge",
        subject: supersededSubject,
      }).effects,
    ).toEqual([]);
  });

  it("publishes failures, clears protected state on access loss, and terminates exhausted retries", () => {
    const active = transitionTimelineLoad(createTimelineLoadState(identity), {
      hasLoadedRows: true,
      kind: "start",
      retryDepth: 0,
      showLoading: false,
      subject: subject(),
    }).state;
    expect(
      transitionTimelineLoad(active, {
        hasLoadedRows: true,
        kind: "failure",
        message: "refresh failed",
        subject: subject(),
      }).effects,
    ).toMatchObject([
      { kind: "fail_continuity" },
      { kind: "publish_status", status: "stale_error" },
    ]);
    expect(
      transitionTimelineLoad(active, {
        kind: "access_loss",
        message: "access lost",
        subject: subject(),
      }).effects.map((effect) => effect.kind),
    ).toEqual(["fail_continuity", "clear_protected_rows", "publish_status"]);
    expect(
      transitionTimelineLoad(active, {
        hasLoadedRows: true,
        kind: "retry_exhaustion",
        message: "did not converge",
        subject: subject(),
      }).effects.map((effect) => effect.kind),
    ).toEqual(["fail_continuity", "publish_status"]);
  });

  it("resets a changed subject and preserves state for an equal subject", () => {
    const state = createTimelineLoadState(identity, 4);
    expect(
      transitionTimelineLoad(state, {
        hasLoadedRows: false,
        identity,
        kind: "subject_changed",
        mutationEpoch: 4,
      }).state,
    ).toBe(state);
    const changed = transitionTimelineLoad(state, {
      hasLoadedRows: false,
      identity: { ...identity, queryIdentity: "query-2" },
      kind: "subject_changed",
      mutationEpoch: 5,
    });
    expect(changed.state.identity.queryIdentity).toBe("query-2");
    expect(changed.effects).toMatchObject([
      { kind: "publish_status", status: "idle" },
    ]);
  });
});
