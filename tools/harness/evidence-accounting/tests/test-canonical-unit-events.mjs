import assert from "node:assert/strict";
import { appendFileSync, chmodSync, lstatSync, mkdtempSync, renameSync, rmSync, truncateSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import {
  readCanonicalUnitEvents,
  reduceCanonicalUnitIntervals,
} from "../canonical-unit-events.mjs";

function event(seq, monotonicMs, kind, unitID, status, extra = {}) {
  return {
    schema_id: "cartulary.harness_unit_event.v4",
    seq,
    monotonic_ms: monotonicMs,
    event: kind,
    unit_id: unitID,
    status,
    needs: [],
    resource_claims: {},
    service_dependencies: [],
    ...(kind === "queued" ? { unit_kind: "runner", row_ids: [] } : {}),
    ...extra,
  };
}

const wait = {
  wait_reason: "rank",
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

test("native event prefixes bind a complete inode range without following appends", async () => {
  const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-canonical-prefix-"));
  try {
    const events = [event(1, 0, "run_started", "run", "running"), event(2, 1, "queued", "unit:a", "pending")];
    const file = writeEvents(directory, "events", events);
    chmodSync(file, 0o600);
    const stat = lstatSync(file, { bigint: true });
    const prefix = { schema_id: "cartulary.harness_unit_event_prefix.v1", run_id: "fixture", file,
      device: stat.dev.toString(), inode: stat.ino.toString(), byte_length: Number(stat.size), last_seq: 2 };
    const collect = async (authority = prefix) => {
      const records = [];
      for await (const record of readCanonicalUnitEvents(file, { prefix: authority })) records.push(record);
      return records;
    };
    const records = [];
    for await (const record of readCanonicalUnitEvents(file, { prefix })) {
      records.push(record);
      if (records.length === 1) appendFileSync(file, '{"unfinished":');
    }
    assert.deepEqual(records, events);
    assert.deepEqual(await collect(), events);
    await assert.rejects(collect({ ...prefix, last_seq: 3 }), /final sequence/u);
    await assert.rejects(collect({ ...prefix, byte_length: prefix.byte_length - 1 }), /record boundary/u);
    await assert.rejects(collect({ ...prefix, inode: "0" }), /inode authority/u);
    await assert.rejects(collect({ ...prefix, file: path.join(directory, "foreign") }), /inode authority/u);
    await assert.rejects(collect({ ...prefix, unknown: true }));
    await assert.rejects(async () => { for await (const _event of readCanonicalUnitEvents(file)) continue; }, /invalid JSON/u);
    truncateSync(file, prefix.byte_length - 1);
    await assert.rejects(collect(), /inode authority/u);
    renameSync(file, `${file}-original`);
    writeFileSync(file, `${events.map((e) => JSON.stringify(e)).join("\n")}\n`, { mode: 0o600 });
    await assert.rejects(collect(), /inode authority/u);
  } finally {
    rmSync(directory, { recursive: true, force: true });
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
    assert.equal(Object.hasOwn(state, "projectedEvents"), false);
    assert.equal(state.waits.size, 1);
    assert.equal(state.intervals.size, 1);
    assert.equal(state.intervals.get("selected").end, 40_002);
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

test("wait segments reject a gap after a closed boundary", async () => {
  const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-canonical-waits-"));
  try {
    const file = writeEvents(directory, "duplicate.ndjson", [
      event(1, 0, "eligible", "a", "pending"),
      event(2, 0, "wait_started", "a", "pending", wait),
      event(3, 0, "wait_ended", "a", "pending", wait),
      event(4, 1, "wait_started", "a", "pending", wait),
    ]);
    await assert.rejects(reduceCanonicalUnitIntervals(file), /invalid wait start/u);
  } finally { rmSync(directory, { force: true, recursive: true }); }
});

test("normalized facts retain cache outcomes, zero-duration hits, cancellation and primary failure", async (t) => {
  const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-normalized-facts-"));
  t.after(() => rmSync(directory, { force: true, recursive: true }));
  const events = [];
  const add = (time, kind, id, status, extra = {}) => events.push(event(events.length + 1, time, kind, id, status, extra));
  for (const [id, cache, terminal, status] of [["hit", "hit", "completed", "passed"], ["failure", "miss", "failed", "failed"], ["cancel", "bypass", "cancelled", "cancelled"]]) {
    const time = events.length;
    add(time, "queued", id, "pending"); add(time, "eligible", id, "pending");
    add(time, "wait_started", id, "pending", wait); add(time, "wait_ended", id, "pending", wait);
    add(time, `cache_${cache}`, id, "pending", { cache_profile_id: "fixture", cache_reason: "fixture" });
    if (cache !== "hit") { add(time, "admitted", id, "running"); add(time, "started", id, "running"); }
    add(time, terminal, id, status, status === "failed" ? { failure_class: "test", failure_reason: "assertion_failure" } : {});
  }
  const facts = await reduceCanonicalUnitIntervals(writeEvents(directory, "facts.ndjson", events));
  assert.deepEqual(facts.cache, { hit: 1, miss: 1, bypass: 1 });
  assert.deepEqual(facts.intervals.get("hit"), { start: 0, end: 0, queue_ms: 0 });
  assert.equal(facts.terminals.get("failure").failure_reason, "assertion_failure");
  assert.equal(facts.terminals.get("failure").failure_class, "test");
  assert.equal(facts.terminals.get("cancel").event, "cancelled");
});

// Many holder transitions must not retain an array of every historical segment.
test("wait segments compact changing holders and preserve exact reason totals", async (t) => {
  const dir = mkdtempSync(path.join(os.tmpdir(), "cartulary-wait-segments-"));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  const events = [];
  const add = (time, kind, extra = {}) => events.push(event(events.length + 1, time, kind, "a", "pending", extra));
  for (const id of ["a", "b", "c"]) events.push(event(events.length + 1, 0, "queued", id, "pending"));
  add(0, "eligible"); add(0, "wait_started", wait); add(3, "wait_ended", wait);
  for (let i = 0; i < 2000; i++) {
    const detail = { wait_reason: "resources", blocking_resources: ["cpu"], blocking_unit_ids: [i % 2 ? "b" : "c"] };
    add(3 + i, "wait_started", detail); add(4 + i, "wait_ended", detail);
  }
  add(2003, "wait_started", wait); add(2005, "wait_ended", wait);
  add(2005, "admitted"); add(2005, "started");
  events.push(event(events.length + 1, 2010, "completed", "a", "passed"));
  const facts = await reduceCanonicalUnitIntervals(writeEvents(dir, "segments.ndjson", events));
  const value = facts.waits.get("a");
  assert.equal(value.count, 2002);
  assert.deepEqual(value.resource_intervals, [{ start: 3, end: 2003 }]);
  assert.deepEqual(value.by_reason, { resources: 2000, exclusive_barrier: 0, rank: 5, scheduler_stop: 0 });
  assert.equal(facts.intervals.get("a").queue_ms, 2005);
  for (const mutation of ["overlap", "after_admission", "old_schema", "unchanged"]) {
    const altered = structuredClone(events);
    if (mutation === "overlap") altered[5].event = "wait_started";
    if (mutation === "old_schema") altered[0].schema_id = "cartulary.harness_unit_event.v3";
    if (mutation === "after_admission") Object.assign(altered.at(-1), { event: "wait_started", ...wait });
    if (mutation === "unchanged") Object.assign(altered[8], { blocking_unit_ids: ["c"] });
    await assert.rejects(reduceCanonicalUnitIntervals(writeEvents(dir, `${mutation}.ndjson`, altered)));
  }
});
