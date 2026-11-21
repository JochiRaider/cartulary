import type {
  CDPSession,
  Page,
  Request,
  TestInfo,
  WebSocket,
} from "@playwright/test";

// Non-gating evidence for the AC-043 blank-row measurement. It never feeds the
// frontend measurement observation, its samples, statistics, or threshold.
export const blankRowDiagnosticSchemaId =
  "cartulary.diagnostic.timeline_blank_row_main_thread.v2";

const inPageKey = "__cartularyMainThreadDiagnostics";
const capacity = 20_000;
const timelineQueryPath = /\/views\/cartulary\.view\.timeline\.v2\/query$/u;
const timelineCreatePath = /\/views\/cartulary\.view\.timeline\.v2\/rows$/u;
const decodeSpikeMs = 10;
const preWindowMs = 250;
const triggerWindowMs = 50;

export type DiagnosticInterval = { start: number; duration: number };

export type DiagnosticLoafScript = {
  invokerType: string;
  invoker: string;
  duration: number;
};

export type DiagnosticLoaf = DiagnosticInterval & {
  blockingDuration: number;
  scripts: DiagnosticLoafScript[];
};

export type DiagnosticDataState = { at: number; value: string };

/**
 * `record_changed` frames are foreign; `record_changed_local` frames echo this
 * page's own creates. Other frame types keep their protocol message type.
 */
export type DiagnosticFrame = { at: number; type: string };

export type DiagnosticRequest = {
  start: number;
  end: number | null;
  ttfbMs: number | null;
  failed: boolean;
};

/** Page resource timing; `null` when the browser withholds a phase. */
export type DiagnosticResource = {
  kind: "query" | "create";
  start: number;
  responseStart: number | null;
  responseEnd: number | null;
  /** The connection protocol, such as "h2" or "http/1.1", when exposed. */
  nextHopProtocol: string | null;
};

/** Cumulative renderer task time read between samples, never inside a window. */
export type DiagnosticTaskClock = {
  at: number;
  taskMs: number;
  scriptMs: number;
};

export type BlankRowSampleWindow = {
  acceptedAt: number;
  stoppedAt: number;
  marks: Record<string, number>;
};

export type DiagnosticSample = {
  sampleIndex: number;
  warmup: boolean;
  totalMs: number;
  window: BlankRowSampleWindow;
  before?: DiagnosticTaskClock | undefined;
  after?: DiagnosticTaskClock | undefined;
};

export type BlankRowDiagnosticInput = {
  supported: {
    longtask: boolean;
    loaf: boolean;
    dataState: boolean;
    resource: boolean;
    taskClock: boolean;
  };
  droppedEntries: number;
  longTasks: DiagnosticInterval[];
  loafs: DiagnosticLoaf[];
  dataStates: DiagnosticDataState[];
  frames: DiagnosticFrame[];
  timelineQueries: DiagnosticRequest[];
  creates: DiagnosticRequest[];
  resources: DiagnosticResource[];
  samples: DiagnosticSample[];
};

export type LoafInvokerClass =
  | "websocket_message"
  | "scheduler_message"
  | "promise_resolution"
  | "animation_frame"
  | "timer"
  | "event_listener"
  | "script"
  | "other";

const invokerClasses: readonly LoafInvokerClass[] = [
  "websocket_message",
  "scheduler_message",
  "promise_resolution",
  "animation_frame",
  "timer",
  "event_listener",
  "script",
  "other",
];

/** Reduces LoAF script attribution to a closed class; raw invoker text is never retained. */
export function classifyLoafInvoker(
  invokerType: string,
  invoker: string,
): LoafInvokerClass {
  if (invoker.startsWith("WebSocket.")) return "websocket_message";
  if (invoker.startsWith("MessagePort.")) return "scheduler_message";
  if (invokerType === "resolve-promise" || invokerType === "reject-promise")
    return "promise_resolution";
  if (invoker.startsWith("FrameRequestCallback")) return "animation_frame";
  if (invoker.startsWith("TimerHandler")) return "timer";
  if (invokerType === "event-listener") return "event_listener";
  if (invokerType === "classic-script" || invokerType === "module-script")
    return "script";
  return "other";
}

function overlap(start: number, end: number, interval: DiagnosticInterval) {
  return Math.max(
    0,
    Math.min(end, interval.start + interval.duration) -
      Math.max(start, interval.start),
  );
}

function round(value: number) {
  return Math.round(value * 10) / 10;
}

function ratio(numerator: number, denominator: number) {
  return denominator === 0
    ? null
    : Math.round((numerator / denominator) * 1000) / 1000;
}

function nearestRank(values: readonly number[], percentile: number) {
  if (values.length === 0) return null;
  const sorted = [...values].sort((left, right) => left - right);
  return round(
    sorted[Math.ceil((percentile / 100) * sorted.length) - 1] as number,
  );
}

function emptyInvokerTotals(): Record<LoafInvokerClass, number> {
  return Object.fromEntries(invokerClasses.map((name) => [name, 0])) as Record<
    LoafInvokerClass,
    number
  >;
}

function addScripts(
  totals: Record<LoafInvokerClass, number>,
  frame: DiagnosticLoaf,
) {
  for (const script of frame.scripts) {
    const name = classifyLoafInvoker(script.invokerType, script.invoker);
    totals[name] += script.duration;
  }
}

function roundedTotals(totals: Record<LoafInvokerClass, number>) {
  return Object.fromEntries(
    invokerClasses.map((name) => [name, round(totals[name])]),
  ) as Record<LoafInvokerClass, number>;
}

function intervalsIn(
  intervals: readonly DiagnosticInterval[],
  start: number,
  end: number,
) {
  let count = 0;
  let overlapMs = 0;
  let maxMs = 0;
  for (const interval of intervals) {
    const covered = overlap(start, end, interval);
    if (covered <= 0) continue;
    count += 1;
    overlapMs += covered;
    maxMs = Math.max(maxMs, interval.duration);
  }
  return { count, overlapMs, maxMs };
}

function countIn<T extends { at: number }>(
  entries: readonly T[],
  start: number,
  end: number,
  accept: (entry: T) => boolean = () => true,
) {
  return entries.filter(
    (entry) => entry.at >= start && entry.at < end && accept(entry),
  ).length;
}

function mark(window: BlankRowSampleWindow, name: string) {
  const value = window.marks[name];
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

type QueryTrigger =
  | "foreign_record_changed"
  | "local_record_changed"
  | "after_query"
  | "other";

/** Attributes a Timeline query to the evidence that immediately preceded it. */
function queryTrigger(
  query: DiagnosticRequest,
  frames: readonly DiagnosticFrame[],
  queries: readonly DiagnosticRequest[],
): QueryTrigger {
  let frame: DiagnosticFrame | undefined;
  for (const candidate of frames)
    if (
      candidate.at <= query.start &&
      query.start - candidate.at <= triggerWindowMs &&
      (candidate.type === "record_changed" ||
        candidate.type === "record_changed_local") &&
      (frame === undefined || candidate.at > frame.at)
    )
      frame = candidate;
  const previous = queries.some(
    (candidate) =>
      candidate !== query &&
      candidate.end !== null &&
      candidate.end <= query.start &&
      query.start - candidate.end <= triggerWindowMs &&
      (frame === undefined || candidate.end > frame.at),
  );
  if (previous) return "after_query";
  if (frame?.type === "record_changed") return "foreign_record_changed";
  if (frame?.type === "record_changed_local") return "local_record_changed";
  return "other";
}

/**
 * Pure attribution of passive main-thread, socket, request, and renderer task
 * evidence to blank-row sample windows and the gaps between them. All times
 * share the page performance clock.
 */
export function summarizeBlankRowDiagnostics(input: BlankRowDiagnosticInput) {
  const samples = [...input.samples].sort(
    (left, right) => left.sampleIndex - right.sampleIndex,
  );
  const first = samples[0];
  const last = samples.at(-1);
  const spanStart = first?.window.acceptedAt ?? 0;
  const spanEnd = last?.window.stoppedAt ?? 0;
  const foreign = (frame: DiagnosticFrame) => frame.type === "record_changed";
  const local = (frame: DiagnosticFrame) =>
    frame.type === "record_changed_local";
  const refreshing = (state: DiagnosticDataState) =>
    state.value === "refreshing";
  const queryStarts = input.timelineQueries.map((query) => ({
    at: query.start,
    trigger: queryTrigger(query, input.frames, input.timelineQueries),
  }));

  const perSample = samples.map((sample) => {
    const { acceptedAt, stoppedAt } = sample.window;
    const requestAt = mark(sample.window, "pending_fetch_request");
    const responseAt = mark(sample.window, "pending_fetch_response");
    const parsedAt = mark(sample.window, "pending_fetch_json_parsed");
    const decodeMs =
      responseAt !== null && parsedAt !== null ? parsedAt - responseAt : null;
    const longTask = intervalsIn(input.longTasks, acceptedAt, stoppedAt);
    const preWindow = intervalsIn(
      input.longTasks,
      acceptedAt - preWindowMs,
      acceptedAt,
    );
    const frames = input.loafs.filter(
      (frame) => overlap(acceptedAt, stoppedAt, frame) > 0,
    );
    const invokerTotals = emptyInvokerTotals();
    for (const frame of frames) addScripts(invokerTotals, frame);
    const decodeLongTaskMs =
      responseAt !== null && parsedAt !== null
        ? intervalsIn(input.longTasks, responseAt, parsedAt).overlapMs
        : 0;
    const decodeLoafMs =
      responseAt !== null && parsedAt !== null
        ? intervalsIn(input.loafs, responseAt, parsedAt).overlapMs
        : 0;
    const create =
      requestAt === null
        ? undefined
        : input.creates.find(
            (candidate) =>
              candidate.start >= requestAt - 50 && candidate.start <= stoppedAt,
          );
    const resource =
      requestAt === null
        ? undefined
        : input.resources.find(
            (candidate) =>
              candidate.kind === "create" &&
              candidate.start >= requestAt - 5 &&
              candidate.start <= stoppedAt,
          );
    const queriesInFlight =
      create === undefined
        ? null
        : input.timelineQueries.filter(
            (query) =>
              query.start < create.start &&
              (query.end === null || query.end > create.start),
          ).length;
    const resourceStart = resource?.responseStart ?? null;
    const resourceEnd = resource?.responseEnd ?? null;
    const bodyMs =
      resourceStart === null || resourceEnd === null
        ? null
        : resourceEnd - resourceStart;
    const parseWaitMs =
      resourceEnd === null || parsedAt === null
        ? null
        : parsedAt - Math.max(resourceEnd, responseAt ?? 0);
    const statusDelayMs =
      resourceStart === null || responseAt === null
        ? null
        : responseAt - resourceStart;
    const sampleTaskMs =
      sample.before === undefined || sample.after === undefined
        ? null
        : sample.after.taskMs - sample.before.taskMs;
    return {
      sample_index: sample.sampleIndex,
      warmup: sample.warmup,
      total_ms: round(sample.totalMs),
      response_decode_ms: decodeMs === null ? null : round(decodeMs),
      decode_spike: decodeMs !== null && decodeMs > decodeSpikeMs,
      create_resource_ms: {
        status_callback_delay:
          statusDelayMs === null ? null : round(statusDelayMs),
        body_transfer: bodyMs === null ? null : round(bodyMs),
        parse_wait_after_body: parseWaitMs === null ? null : round(parseWaitMs),
      },
      create_next_hop_protocol: resource?.nextHopProtocol ?? null,
      long_task: {
        count: longTask.count,
        overlap_ms: round(longTask.overlapMs),
        max_ms: round(longTask.maxMs),
      },
      long_task_pre_window: {
        count: preWindow.count,
        overlap_ms: round(preWindow.overlapMs),
      },
      long_animation_frame: {
        count: frames.length,
        blocking_ms: round(
          frames.reduce((total, frame) => total + frame.blockingDuration, 0),
        ),
        script_ms_by_invoker: roundedTotals(invokerTotals),
      },
      decode_blocked_ms: {
        long_task: round(decodeLongTaskMs),
        long_animation_frame: round(decodeLoafMs),
      },
      record_changed_frames: countIn(
        input.frames,
        acceptedAt,
        stoppedAt,
        foreign,
      ),
      timeline_queries_started: countIn(queryStarts, acceptedAt, stoppedAt),
      timeline_queries_in_flight_at_create: queriesInFlight,
      create_ttfb_ms:
        create === undefined || create.ttfbMs === null
          ? null
          : round(create.ttfbMs),
      renderer_task_ms_around_sample:
        sampleTaskMs === null ? null : round(sampleTaskMs),
      refreshing_entries: countIn(
        input.dataStates,
        acceptedAt,
        stoppedAt,
        refreshing,
      ),
    };
  });

  const gaps = samples.slice(1).flatMap((sample, index) => {
    const previous = samples[index];
    if (previous === undefined) return [];
    const start = previous.window.stoppedAt;
    const end = sample.window.acceptedAt;
    return end > start ? [{ start, end }] : [];
  });
  let gapMs = 0;
  let gapFrames = 0;
  let gapQueries = 0;
  let gapLongTaskCount = 0;
  let gapLongTaskMs = 0;
  let gapLoafMs = 0;
  const gapInvokers = emptyInvokerTotals();
  for (const gap of gaps) {
    gapMs += gap.end - gap.start;
    gapFrames += countIn(input.frames, gap.start, gap.end, foreign);
    gapQueries += countIn(queryStarts, gap.start, gap.end);
    const tasks = intervalsIn(input.longTasks, gap.start, gap.end);
    gapLongTaskCount += tasks.count;
    gapLongTaskMs += tasks.overlapMs;
    for (const frame of input.loafs) {
      const covered = overlap(gap.start, gap.end, frame);
      if (covered <= 0) continue;
      gapLoafMs += covered;
      addScripts(gapInvokers, frame);
    }
  }

  // Renderer task time between one sample's end and the next sample's start.
  let clockGapMs = 0;
  let clockGapTaskMs = 0;
  let clockGapScriptMs = 0;
  let clockGapFrames = 0;
  let clockGapLocalFrames = 0;
  for (const [index, sample] of samples.slice(1).entries()) {
    const previous = samples[index];
    const from = previous?.after;
    const to = sample.before;
    if (from === undefined || to === undefined || to.at <= from.at) continue;
    clockGapMs += to.at - from.at;
    clockGapTaskMs += to.taskMs - from.taskMs;
    clockGapScriptMs += to.scriptMs - from.scriptMs;
    clockGapFrames += countIn(input.frames, from.at, to.at, foreign);
    clockGapLocalFrames += countIn(input.frames, from.at, to.at, local);
  }

  const measured = perSample.filter((sample) => !sample.warmup);
  const spikes = measured.filter((sample) => sample.decode_spike);
  const spikesBlocked = spikes.filter(
    (sample) =>
      sample.decode_blocked_ms.long_task > 0 ||
      sample.decode_blocked_ms.long_animation_frame > 0,
  );
  const spikeParseWait = spikes.filter(
    (sample) =>
      (sample.create_resource_ms.parse_wait_after_body ?? 0) > decodeSpikeMs,
  );
  const spikeBody = spikes.filter(
    (sample) => (sample.create_resource_ms.body_transfer ?? 0) > decodeSpikeMs,
  );
  const spanFrames = countIn(input.frames, spanStart, spanEnd, foreign);
  const spanLocalFrames = countIn(input.frames, spanStart, spanEnd, local);
  const spanQueries = queryStarts.filter(
    (query) => query.at >= spanStart && query.at < spanEnd,
  );
  const triggers = {
    foreign_record_changed: 0,
    local_record_changed: 0,
    after_query: 0,
    other: 0,
  } satisfies Record<QueryTrigger, number>;
  for (const query of spanQueries) triggers[query.trigger] += 1;
  const ttfb = (inFlight: (count: number) => boolean) =>
    measured.flatMap((sample) =>
      sample.create_ttfb_ms !== null &&
      sample.timeline_queries_in_flight_at_create !== null &&
      inFlight(sample.timeline_queries_in_flight_at_create)
        ? [sample.create_ttfb_ms]
        : [],
    );
  const protocols: Record<string, number> = {};
  for (const sample of measured) {
    const protocol = sample.create_next_hop_protocol ?? "unreported";
    protocols[protocol] = (protocols[protocol] ?? 0) + 1;
  }
  const bodySpikes = (inFlight: (count: number) => boolean) => {
    const selected = measured.filter(
      (sample) =>
        sample.create_resource_ms.body_transfer !== null &&
        sample.timeline_queries_in_flight_at_create !== null &&
        inFlight(sample.timeline_queries_in_flight_at_create),
    );
    const spiked = selected.filter(
      (sample) =>
        (sample.create_resource_ms.body_transfer ?? 0) > decodeSpikeMs,
    ).length;
    return {
      count: selected.length,
      spikes: spiked,
      rate: ratio(spiked, selected.length),
    };
  };
  const allTtfb = ttfb(() => true);
  const busyTtfb = ttfb((count) => count > 0);
  const quietTtfb = ttfb((count) => count === 0);

  return {
    schema_id: blankRowDiagnosticSchemaId,
    supported: {
      long_task: input.supported.longtask,
      long_animation_frame: input.supported.loaf,
      data_state: input.supported.dataState,
      resource_timing: input.supported.resource,
      renderer_task_clock: input.supported.taskClock,
    },
    dropped_entries: input.droppedEntries,
    run: {
      sampling_span_ms: round(Math.max(0, spanEnd - spanStart)),
      record_changed_frames: spanFrames,
      record_changed_local_frames: spanLocalFrames,
      timeline_queries_started: spanQueries.length,
      timeline_queries_by_trigger: triggers,
      timeline_queries_failed: input.timelineQueries.filter(
        (query) =>
          query.failed && query.start >= spanStart && query.start < spanEnd,
      ).length,
      timeline_queries_per_record_changed: ratio(
        spanQueries.length,
        spanFrames,
      ),
      foreign_triggered_queries_per_record_changed: ratio(
        triggers.foreign_record_changed,
        spanFrames,
      ),
      refreshing_entries: countIn(
        input.dataStates,
        spanStart,
        spanEnd,
        refreshing,
      ),
      gaps: {
        count: gaps.length,
        duration_ms: round(gapMs),
        record_changed_frames: gapFrames,
        timeline_queries_started: gapQueries,
        long_task_count: gapLongTaskCount,
        long_task_ms: round(gapLongTaskMs),
        long_animation_frame_ms: round(gapLoafMs),
        long_task_ms_per_record_changed:
          gapFrames === 0 ? null : round(gapLongTaskMs / gapFrames),
        long_animation_frame_ms_per_record_changed:
          gapFrames === 0 ? null : round(gapLoafMs / gapFrames),
        script_ms_by_invoker: roundedTotals(gapInvokers),
      },
      renderer_task_clock_gaps: {
        elapsed_ms: round(clockGapMs),
        task_ms: round(clockGapTaskMs),
        script_ms: round(clockGapScriptMs),
        busy_fraction: ratio(clockGapTaskMs, clockGapMs),
        record_changed_frames: clockGapFrames,
        record_changed_local_frames: clockGapLocalFrames,
        task_ms_per_record_changed:
          clockGapFrames === 0 ? null : round(clockGapTaskMs / clockGapFrames),
      },
      decode_spikes: {
        threshold_ms: decodeSpikeMs,
        measured_samples: measured.length,
        count: spikes.length,
        rate: ratio(spikes.length, measured.length),
        overlapped_by_long_work: spikesBlocked.length,
        parse_wait_after_body_over_threshold: spikeParseWait.length,
        body_transfer_over_threshold: spikeBody.length,
      },
      create_next_hop_protocols: protocols,
      create_body_transfer_spikes: {
        threshold_ms: decodeSpikeMs,
        with_timeline_query_in_flight: bodySpikes((count) => count > 0),
        without_timeline_query_in_flight: bodySpikes((count) => count === 0),
      },
      create_ttfb_ms: {
        p50: nearestRank(allTtfb, 50),
        p95: nearestRank(allTtfb, 95),
        with_timeline_query_in_flight: {
          count: busyTtfb.length,
          p50: nearestRank(busyTtfb, 50),
          p95: nearestRank(busyTtfb, 95),
        },
        without_timeline_query_in_flight: {
          count: quietTtfb.length,
          p50: nearestRank(quietTtfb, 50),
          p95: nearestRank(quietTtfb, 95),
        },
      },
    },
    samples: perSample,
  };
}

export type BlankRowDiagnosticSummary = ReturnType<
  typeof summarizeBlankRowDiagnostics
>;

type InPageStore = {
  supported: {
    longtask: boolean;
    loaf: boolean;
    dataState: boolean;
    resource: boolean;
  };
  dropped: number;
  longTasks: DiagnosticInterval[];
  loafs: DiagnosticLoaf[];
  dataStates: DiagnosticDataState[];
  resources: DiagnosticResource[];
};

/** Runs before any page script; observers are passive and buffer bounded entries. */
export function installMainThreadObserversInBrowser(options: {
  key: string;
  capacity: number;
  queryPath: string;
  createPath: string;
}) {
  const store: InPageStore = {
    supported: {
      longtask: false,
      loaf: false,
      dataState: false,
      resource: false,
    },
    dropped: 0,
    longTasks: [],
    loafs: [],
    dataStates: [],
    resources: [],
  };
  Object.defineProperty(window, options.key, {
    configurable: false,
    enumerable: false,
    value: store,
    writable: false,
  });
  const push = <T>(list: T[], entry: T) => {
    if (list.length >= options.capacity) store.dropped += 1;
    else list.push(entry);
  };
  const supported =
    typeof PerformanceObserver === "undefined"
      ? []
      : (PerformanceObserver.supportedEntryTypes ?? []);
  if (supported.includes("longtask")) {
    try {
      new PerformanceObserver((list) => {
        for (const entry of list.getEntries())
          push(store.longTasks, {
            start: entry.startTime,
            duration: entry.duration,
          });
      }).observe({ type: "longtask", buffered: true });
      store.supported.longtask = true;
    } catch {
      store.supported.longtask = false;
    }
  }
  if (supported.includes("long-animation-frame")) {
    try {
      new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          const frame = entry as PerformanceEntry & {
            blockingDuration?: number;
            scripts?: readonly {
              duration?: number;
              invoker?: string;
              invokerType?: string;
            }[];
          };
          push(store.loafs, {
            start: frame.startTime,
            duration: frame.duration,
            blockingDuration: frame.blockingDuration ?? 0,
            scripts: Array.from(frame.scripts ?? [], (script) => {
              const invoker = String(script.invoker ?? "");
              return {
                invokerType: String(script.invokerType ?? ""),
                // Script URLs and source positions never leave the page.
                invoker: invoker.includes("/") ? "" : invoker.slice(0, 64),
                duration: script.duration ?? 0,
              };
            }),
          });
        }
      }).observe({ type: "long-animation-frame", buffered: true });
      store.supported.loaf = true;
    } catch {
      store.supported.loaf = false;
    }
  }
  if (supported.includes("resource")) {
    try {
      const query = new RegExp(options.queryPath, "u");
      const create = new RegExp(options.createPath, "u");
      new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          const timing = entry as PerformanceResourceTiming;
          let pathname: string;
          try {
            pathname = new URL(timing.name).pathname;
          } catch {
            continue;
          }
          const kind = query.test(pathname)
            ? "query"
            : create.test(pathname)
              ? "create"
              : null;
          if (kind === null) continue;
          // Only phase times leave the page; URLs and identifiers do not.
          push(store.resources, {
            kind,
            start: timing.startTime,
            responseStart:
              timing.responseStart > 0 ? timing.responseStart : null,
            responseEnd: timing.responseEnd > 0 ? timing.responseEnd : null,
            nextHopProtocol: timing.nextHopProtocol || null,
          });
        }
      }).observe({ type: "resource", buffered: true });
      store.supported.resource = true;
    } catch {
      store.supported.resource = false;
    }
  }
  try {
    new MutationObserver((records) => {
      for (const record of records) {
        const target = record.target;
        if (!(target instanceof Element)) continue;
        push(store.dataStates, {
          at: performance.now(),
          value: target.getAttribute("data-grid-data-state") ?? "",
        });
      }
    }).observe(document, {
      attributeFilter: ["data-grid-data-state"],
      attributes: true,
      subtree: true,
    });
    store.supported.dataState = true;
  } catch {
    store.supported.dataState = false;
  }
}

type RecordedRequest = {
  kind: "query" | "create";
  startEpochMs: number;
  endEpochMs: number | null;
  ttfbMs: number | null;
  failed: boolean;
};

type EpochTaskClock = { epochMs: number; taskMs: number; scriptMs: number };

export type BlankRowDiagnosticsRecorder = {
  /** Reads the renderer task clock before a sample starts. */
  readonly beforeSample: () => Promise<void>;
  /** Records a completed sample and reads the renderer task clock after it. */
  readonly recordSample: (
    sample: Omit<DiagnosticSample, "before" | "after">,
  ) => Promise<void>;
  readonly attach: (testInfo: TestInfo) => Promise<void>;
};

function parsedJSON(payload: string | Buffer | null): unknown {
  if (payload === null) return null;
  try {
    return JSON.parse(
      typeof payload === "string" ? payload : payload.toString("utf8"),
    );
  } catch {
    return null;
  }
}

function stringMember(value: unknown, key: string): string | null {
  if (value === null || typeof value !== "object" || !(key in value))
    return null;
  const member = (value as Record<string, unknown>)[key];
  return typeof member === "string" ? member : null;
}

function objectMember(value: unknown, key: string): unknown {
  return value !== null && typeof value === "object" && key in value
    ? (value as Record<string, unknown>)[key]
    : null;
}

/**
 * Installs passive observers before navigation. Socket frames and requests are
 * observed through Playwright events; the renderer task clock is read through
 * CDP only between samples.
 */
export async function installBlankRowDiagnostics(
  page: Page,
): Promise<BlankRowDiagnosticsRecorder> {
  await page.addInitScript(installMainThreadObserversInBrowser, {
    key: inPageKey,
    capacity,
    queryPath: timelineQueryPath.source,
    createPath: timelineCreatePath.source,
  });
  const frames: { epochMs: number; type: string }[] = [];
  const requests = new Map<Request, RecordedRequest>();
  // Local create transaction identities stay in this process and are never attached.
  const localTransactions = new Set<string>();
  const samples: DiagnosticSample[] = [];
  const clocks = new Map<
    number,
    { before?: EpochTaskClock; after?: EpochTaskClock }
  >();
  let session: CDPSession | null = null;
  try {
    session = await page.context().newCDPSession(page);
    await session.send("Performance.enable");
  } catch {
    session = null;
  }
  page.on("websocket", (socket: WebSocket) => {
    socket.on("framereceived", (frame) => {
      if (frames.length >= capacity) return;
      const message = parsedJSON(frame.payload);
      const type = stringMember(message, "type")?.slice(0, 48) ?? "unparsed";
      const txn = stringMember(
        objectMember(message, "payload"),
        "client_txn_id",
      );
      frames.push({
        epochMs: Date.now(),
        type:
          type === "record_changed" &&
          txn !== null &&
          localTransactions.has(txn)
            ? "record_changed_local"
            : type,
      });
    });
  });
  page.on("request", (request) => {
    if (request.method() !== "POST") return;
    let pathname: string;
    try {
      pathname = new URL(request.url()).pathname;
    } catch {
      return;
    }
    const kind = timelineQueryPath.test(pathname)
      ? "query"
      : timelineCreatePath.test(pathname)
        ? "create"
        : null;
    if (kind === null || requests.size >= capacity) return;
    if (kind === "create") {
      const txn = stringMember(parsedJSON(request.postData()), "client_txn_id");
      if (txn !== null) localTransactions.add(txn);
    }
    requests.set(request, {
      kind,
      startEpochMs: Date.now(),
      endEpochMs: null,
      ttfbMs: null,
      failed: false,
    });
  });
  page.on("requestfinished", (request) => {
    const entry = requests.get(request);
    if (entry === undefined) return;
    const timing = request.timing();
    if (Number.isFinite(timing.startTime) && timing.startTime > 0) {
      entry.startEpochMs = timing.startTime;
      if (timing.responseEnd >= 0)
        entry.endEpochMs = timing.startTime + timing.responseEnd;
    }
    entry.endEpochMs ??= Date.now();
    if (timing.requestStart >= 0 && timing.responseStart >= 0)
      entry.ttfbMs = timing.responseStart - timing.requestStart;
  });
  page.on("requestfailed", (request) => {
    const entry = requests.get(request);
    if (entry === undefined) return;
    entry.failed = true;
    entry.endEpochMs = Date.now();
  });
  const readClock = async (): Promise<EpochTaskClock | undefined> => {
    if (session === null) return undefined;
    try {
      const { metrics } = await session.send("Performance.getMetrics");
      const value = (name: string) =>
        metrics.find((metric) => metric.name === name)?.value;
      const taskSeconds = value("TaskDuration");
      const scriptSeconds = value("ScriptDuration");
      if (taskSeconds === undefined || scriptSeconds === undefined)
        return undefined;
      return {
        epochMs: Date.now(),
        taskMs: taskSeconds * 1000,
        scriptMs: scriptSeconds * 1000,
      };
    } catch {
      return undefined;
    }
  };

  return {
    async beforeSample() {
      const before = await readClock();
      if (before !== undefined) clocks.set(samples.length, { before });
    },
    async recordSample(sample) {
      const after = await readClock();
      const entry = clocks.get(samples.length) ?? {};
      if (after !== undefined) entry.after = after;
      clocks.set(samples.length, entry);
      samples.push(sample);
    },
    async attach(testInfo) {
      let body: unknown;
      try {
        const inPage = await page.evaluate((key) => {
          const store = (window as unknown as Record<string, unknown>)[key] as
            | InPageStore
            | undefined;
          return store === undefined
            ? null
            : { store, timeOrigin: performance.timeOrigin };
        }, inPageKey);
        if (inPage === null) throw new Error("diagnostic observers missing");
        const toPage = (epochMs: number) => epochMs - inPage.timeOrigin;
        const recorded = [...requests.values()];
        const asRequest = (entry: RecordedRequest): DiagnosticRequest => ({
          start: toPage(entry.startEpochMs),
          end: entry.endEpochMs === null ? null : toPage(entry.endEpochMs),
          ttfbMs: entry.ttfbMs,
          failed: entry.failed,
        });
        const asClock = (clock: EpochTaskClock | undefined) =>
          clock === undefined
            ? undefined
            : {
                at: toPage(clock.epochMs),
                taskMs: clock.taskMs,
                scriptMs: clock.scriptMs,
              };
        body = summarizeBlankRowDiagnostics({
          supported: { ...inPage.store.supported, taskClock: session !== null },
          droppedEntries: inPage.store.dropped,
          longTasks: inPage.store.longTasks,
          loafs: inPage.store.loafs,
          dataStates: inPage.store.dataStates,
          resources: inPage.store.resources,
          frames: frames.map((frame) => ({
            at: toPage(frame.epochMs),
            type: frame.type,
          })),
          timelineQueries: recorded
            .filter((entry) => entry.kind === "query")
            .map(asRequest),
          creates: recorded
            .filter((entry) => entry.kind === "create")
            .map(asRequest),
          samples: samples.map((sample, index) => ({
            ...sample,
            before: asClock(clocks.get(index)?.before),
            after: asClock(clocks.get(index)?.after),
          })),
        });
      } catch (error) {
        body = {
          schema_id: blankRowDiagnosticSchemaId,
          collection_failure: (error instanceof Error
            ? error.message
            : String(error)
          ).slice(0, 256),
        };
      }
      try {
        await testInfo.attach(blankRowDiagnosticSchemaId, {
          body: Buffer.from(`${JSON.stringify(body)}\n`, "utf8"),
          contentType: "application/json",
        });
      } catch {
        // Diagnostics are evidence only and never decide the measurement.
      }
      try {
        await session?.detach();
      } catch {
        // The page may already be closed.
      }
    },
  };
}
