import assert from "node:assert/strict";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import {
  readCanonicalUnitEvents,
  reduceCanonicalUnitIntervals,
} from "../canonical-unit-events.mjs";

function event(seq, monotonicMs, kind, unitID, status, extra = {}) {
  return {
    schema_id: "cartulary.harness_unit_event.v2",
    seq,
    monotonic_ms: monotonicMs,
    event: kind,
    unit_id: unitID,
    status,
    needs: [],
    resource_claims: {},
    service_dependencies: [],
    ...extra,
  };
}

const wait = {
  wait_reason: "capacity",
  blocking_resources: [],
  blocking_unit_ids: [],
};

function writeEvents(directory, name, events) {
  const file = path.join(directory, name);
  writeFileSync(file, `${events.map((entry) => JSON.stringify(entry)).join("\n")}\n`);
  return file;
}

test("canonical event iterator validates JSON, sequence, time, and line bounds", async () => {
  const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-canonical-events-"));
  try {
    const malformed = path.join(directory, "malformed.ndjson");
    writeFileSync(malformed, "{not-json}\n");
    await assert.rejects(async () => {
      for await (const _event of readCanonicalUnitEvents(malformed)) continue;
    }, /line 1 is invalid JSON/u);

    const gap = writeEvents(directory, "gap.ndjson", [
      event(2, 0, "run_started", "run", "running"),
    ]);
    await assert.rejects(async () => {
      for await (const _event of readCanonicalUnitEvents(gap)) continue;
    }, /sequence 2 is not contiguous/u);

    const reversed = writeEvents(directory, "reversed.ndjson", [
      event(1, 2, "run_started", "run", "running"),
      event(2, 1, "run_completed", "run", "passed"),
    ]);
    await assert.rejects(async () => {
      for await (const _event of readCanonicalUnitEvents(reversed)) continue;
    }, /monotonic time regresses/u);

    const oversized = path.join(directory, "oversized.ndjson");
    writeFileSync(
      oversized,
      `${JSON.stringify(event(1, 0, "run_started", "run", "running"))}${" ".repeat(4096)}\n`,
    );
    await assert.rejects(async () => {
      for await (const _event of readCanonicalUnitEvents(oversized, { maxLineBytes: 1024 })) continue;
    }, /exceeds 1024 bytes/u);
  } finally {
    rmSync(directory, { force: true, recursive: true });
  }
});

test("canonical event iterator releases early and honors cancellation", async () => {
  const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-canonical-events-"));
  try {
    const file = writeEvents(directory, "events.ndjson", [
      event(1, 0, "run_started", "run", "running"),
      event(2, 1, "eligible", "unit:a", "pending"),
      event(3, 1, "wait_started", "unit:a", "pending", wait),
      event(4, 1, "wait_ended", "unit:a", "pending", wait),
      event(5, 1, "admitted", "unit:a", "running"),
      event(6, 1, "started", "unit:a", "running"),
      event(7, 2, "cancelled", "unit:a", "cancelled"),
      event(8, 3, "run_completed", "run", "cancelled"),
    ]);
    let count = 0;
    for await (const _event of readCanonicalUnitEvents(file)) {
      count += 1;
      break;
    }
    assert.equal(count, 1);
    const controller = new AbortController();
    controller.abort();
    await assert.rejects(async () => {
      for await (const _event of readCanonicalUnitEvents(file, { signal: controller.signal })) continue;
    }, /cancelled/u);
    const state = await reduceCanonicalUnitIntervals(file);
    assert.equal(state.eventCount, 8);
    assert.equal(state.terminals.get("unit:a").event, "cancelled");
    assert.equal(state.runCompleted.status, "cancelled");
  } finally {
    rmSync(directory, { force: true, recursive: true });
  }
});

test("unit reducer retains bounded selected state while validating the full stream", async () => {
  const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-canonical-events-"));
  try {
    const events = [event(1, 0, "run_started", "run", "running")];
    for (let index = 0; index < 20_000; index += 1) {
      events.push(event(events.length + 1, index + 1, "queued", `unit:${index}`, "pending"));
    }
    events.push(event(events.length + 1, 20_001, "eligible", "selected", "pending"));
    events.push(event(events.length + 1, 20_001, "wait_started", "selected", "pending", wait));
    events.push(event(events.length + 1, 20_001, "wait_ended", "selected", "pending", wait));
    events.push(event(events.length + 1, 20_001, "admitted", "selected", "running"));
    events.push(event(events.length + 1, 20_001, "started", "selected", "running"));
    for (let index = 0; index < 20_000; index += 1) {
      events.push(event(events.length + 1, 20_002 + index, "skipped", `unit:${index}`, "skipped", {
        failure_reason: "dependency_failure",
      }));
    }
    events.push(event(events.length + 1, 40_002, "completed", "selected", "passed"));
    events.push(event(events.length + 1, 40_003, "run_completed", "run", "passed"));
    const file = writeEvents(directory, "large.ndjson", events);
    const state = await reduceCanonicalUnitIntervals(file, { unitIDs: ["selected"] });
    assert.equal(state.eventCount, 40_008);
    assert.deepEqual([...state.starts.keys()], ["selected"]);
    assert.deepEqual([...state.terminals.keys()], ["selected"]);
  } finally {
    rmSync(directory, { force: true, recursive: true });
  }
});

test("unit reducer rejects unmatched and mismatched wait boundaries", async () => {
  const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-canonical-events-"));
  try {
    const unmatched = writeEvents(directory, "unmatched.ndjson", [
      event(1, 0, "eligible", "unit:a", "pending"),
      event(2, 0, "wait_started", "unit:a", "pending", wait),
    ]);
    await assert.rejects(
      reduceCanonicalUnitIntervals(unmatched, { unitIDs: ["unit:a"] }),
      /unmatched wait start/u,
    );

    const mismatched = writeEvents(directory, "mismatched.ndjson", [
      event(1, 0, "eligible", "unit:a", "pending"),
      event(2, 0, "wait_started", "unit:a", "pending", wait),
      event(3, 0, "wait_ended", "unit:a", "pending", {
        ...wait,
        wait_reason: "scheduler_stop",
      }),
    ]);
    await assert.rejects(
      reduceCanonicalUnitIntervals(mismatched, { unitIDs: ["unit:a"] }),
      /mismatched wait end/u,
    );
  } finally {
    rmSync(directory, { force: true, recursive: true });
  }
});
