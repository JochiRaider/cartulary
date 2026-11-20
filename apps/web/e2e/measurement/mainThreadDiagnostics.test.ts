import { describe, expect, it } from "vitest";

import {
  type BlankRowDiagnosticInput,
  blankRowDiagnosticSchemaId,
  classifyLoafInvoker,
  summarizeBlankRowDiagnostics,
} from "./mainThreadDiagnostics";

function sample(
  sampleIndex: number,
  acceptedAt: number,
  stoppedAt: number,
  decode: readonly [number, number],
  clock: readonly [number, number, number, number, number, number],
) {
  return {
    sampleIndex,
    warmup: sampleIndex === 0,
    totalMs: stoppedAt - acceptedAt,
    window: {
      acceptedAt,
      stoppedAt,
      marks: {
        pending_fetch_request: acceptedAt + 5,
        pending_fetch_response: decode[0],
        pending_fetch_json_parsed: decode[1],
      },
    },
    before: { at: clock[0], taskMs: clock[1], scriptMs: clock[2] },
    after: { at: clock[3], taskMs: clock[4], scriptMs: clock[5] },
  };
}

function input(): BlankRowDiagnosticInput {
  return {
    supported: {
      longtask: true,
      loaf: true,
      dataState: true,
      resource: true,
      taskClock: true,
    },
    droppedEntries: 0,
    longTasks: [
      { start: 1200, duration: 60 },
      { start: 1335, duration: 55 },
      { start: 1500, duration: 60 },
    ],
    loafs: [
      {
        start: 1195,
        duration: 70,
        blockingDuration: 20,
        scripts: [
          {
            invokerType: "event-listener",
            invoker: "WebSocket.onmessage",
            duration: 30,
          },
          {
            invokerType: "event-listener",
            invoker: "MessagePort.onmessage",
            duration: 25,
          },
        ],
      },
    ],
    dataStates: [
      { at: 1152, value: "refreshing" },
      { at: 1225, value: "ready" },
      { at: 1452, value: "refreshing" },
    ],
    frames: [
      { at: 1150, type: "record_changed" },
      { at: 1160, type: "presence_delta" },
      { at: 1295, type: "record_changed_local" },
      { at: 1350, type: "record_changed" },
      { at: 1450, type: "record_changed" },
    ],
    timelineQueries: [
      { start: 1151, end: 1220, ttfbMs: 30, failed: false },
      { start: 1300, end: 1350, ttfbMs: 30, failed: false },
      { start: 1451, end: 1520, ttfbMs: 30, failed: true },
      { start: 1525, end: 1580, ttfbMs: 30, failed: false },
    ],
    creates: [
      { start: 1006, end: 1040, ttfbMs: 20, failed: false },
      { start: 1306, end: 1340, ttfbMs: 60, failed: false },
      { start: 1606, end: 1640, ttfbMs: 25, failed: false },
    ],
    resources: [
      { kind: "query", start: 1151, responseStart: 1180, responseEnd: 1181 },
      { kind: "create", start: 1305.5, responseStart: 1338, responseEnd: 1339 },
      { kind: "create", start: 1605.5, responseStart: 1630, responseEnd: 1641 },
    ],
    samples: [
      sample(2, 1600, 1690, [1640, 1641], [1590, 1230, 600, 1700, 1260, 610]),
      sample(0, 1000, 1100, [1040, 1041], [990, 1000, 500, 1110, 1030, 510]),
      sample(1, 1300, 1400, [1340, 1390], [1290, 1100, 540, 1410, 1150, 560]),
    ],
  };
}

const protectedKey =
  /^(?:bucket_name|cookie|credential|credentials|database_name|dsn|email|entered_text|password|payload|record_id|runtime_path|secret|session|token|transaction_id|txn_id|user_id)$/iu;

function keys(value: unknown): string[] {
  if (Array.isArray(value)) return value.flatMap(keys);
  if (value === null || typeof value !== "object") return [];
  return Object.entries(value).flatMap(([key, child]) => [key, ...keys(child)]);
}

describe("AC-043 blank-row main-thread diagnostics", () => {
  it("attributes long work, socket frames and Timeline queries to sample windows and gaps", () => {
    const summary = summarizeBlankRowDiagnostics(input());

    expect(summary.schema_id).toBe(blankRowDiagnosticSchemaId);
    expect(summary.samples.map((entry) => entry.sample_index)).toEqual([
      0, 1, 2,
    ]);
    expect(summary.run).toMatchObject({
      sampling_span_ms: 690,
      record_changed_frames: 3,
      record_changed_local_frames: 1,
      timeline_queries_started: 4,
      timeline_queries_by_trigger: {
        foreign_record_changed: 2,
        local_record_changed: 1,
        after_query: 1,
        other: 0,
      },
      timeline_queries_failed: 1,
      timeline_queries_per_record_changed: 1.333,
      foreign_triggered_queries_per_record_changed: 0.667,
      refreshing_entries: 2,
      gaps: {
        count: 2,
        duration_ms: 400,
        record_changed_frames: 2,
        timeline_queries_started: 3,
        long_task_count: 2,
        long_task_ms: 120,
        long_animation_frame_ms: 70,
        long_task_ms_per_record_changed: 60,
        long_animation_frame_ms_per_record_changed: 35,
      },
      renderer_task_clock_gaps: {
        elapsed_ms: 360,
        task_ms: 150,
        script_ms: 70,
        busy_fraction: 0.417,
        record_changed_frames: 2,
        record_changed_local_frames: 0,
        task_ms_per_record_changed: 75,
      },
      decode_spikes: {
        threshold_ms: 10,
        measured_samples: 2,
        count: 1,
        rate: 0.5,
        overlapped_by_long_work: 1,
        parse_wait_after_body_over_threshold: 1,
        body_transfer_over_threshold: 0,
      },
      create_ttfb_ms: {
        p50: 25,
        p95: 60,
        with_timeline_query_in_flight: { count: 1, p50: 60, p95: 60 },
        without_timeline_query_in_flight: { count: 1, p50: 25, p95: 25 },
      },
    });
    expect(summary.run.gaps.script_ms_by_invoker).toMatchObject({
      websocket_message: 30,
      scheduler_message: 25,
      promise_resolution: 0,
    });
    expect(summary.samples[1]).toMatchObject({
      warmup: false,
      total_ms: 100,
      response_decode_ms: 50,
      decode_spike: true,
      create_resource_ms: {
        status_callback_delay: 2,
        body_transfer: 1,
        parse_wait_after_body: 50,
      },
      long_task: { count: 1, overlap_ms: 55, max_ms: 55 },
      long_task_pre_window: { count: 1, overlap_ms: 60 },
      long_animation_frame: { count: 0, blocking_ms: 0 },
      decode_blocked_ms: { long_task: 50, long_animation_frame: 0 },
      record_changed_frames: 1,
      timeline_queries_started: 1,
      timeline_queries_in_flight_at_create: 1,
      create_ttfb_ms: 60,
      renderer_task_ms_around_sample: 50,
      refreshing_entries: 0,
    });
    expect(summary.samples[2]).toMatchObject({
      decode_spike: false,
      create_resource_ms: { body_transfer: 11, parse_wait_after_body: 0 },
      long_task: { count: 0, overlap_ms: 0 },
      timeline_queries_in_flight_at_create: 0,
      create_ttfb_ms: 25,
    });
    expect(summary.samples[0]).toMatchObject({
      warmup: true,
      create_resource_ms: {
        status_callback_delay: null,
        body_transfer: null,
        parse_wait_after_body: null,
      },
    });
  });

  it("keeps only closed invoker classes and no protected keys in the attachment", () => {
    expect(
      [
        ["event-listener", "WebSocket.onmessage"],
        ["event-listener", "MessagePort.onmessage"],
        ["resolve-promise", "Response.json.then"],
        ["user-callback", "FrameRequestCallback"],
        ["user-callback", "TimerHandler:setTimeout"],
        ["event-listener", "HTMLInputElement.oninput"],
        ["classic-script", ""],
        ["user-callback", "Window.requestIdleCallback"],
      ].map(([invokerType, invoker]) =>
        classifyLoafInvoker(invokerType ?? "", invoker ?? ""),
      ),
    ).toEqual([
      "websocket_message",
      "scheduler_message",
      "promise_resolution",
      "animation_frame",
      "timer",
      "event_listener",
      "script",
      "other",
    ]);

    const body = JSON.stringify(summarizeBlankRowDiagnostics(input()));
    expect(body).not.toContain("WebSocket.onmessage");
    expect(body).not.toContain("MessagePort.onmessage");
    expect(
      keys(JSON.parse(body)).filter((key) => protectedKey.test(key)),
    ).toEqual([]);

    const empty = summarizeBlankRowDiagnostics({
      ...input(),
      frames: [],
      samples: [],
    });
    expect(empty.run.timeline_queries_per_record_changed).toBeNull();
    expect(empty.run.decode_spikes.rate).toBeNull();
    expect(empty.samples).toEqual([]);
  });
});
