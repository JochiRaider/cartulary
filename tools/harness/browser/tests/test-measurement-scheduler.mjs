import assert from "node:assert/strict";
import { chmodSync, closeSync, lstatSync, linkSync, mkdtempSync, openSync, readFileSync, rmSync, symlinkSync, writeFileSync, writeSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { currentUnitEventSource, readMeasurementSchedulerEvidence } from "../frontend-measurement-evidence.mjs";
import test from "node:test";
const root = path.resolve(import.meta.dirname, "../../../..");
test("scheduler evidence admits bounded private prefixes and rejects overlap and incomplete intervals", async (t) => {
    const measurementGroupResult = {
        group_id: "measurement-timeline-grid",
        stage_id: "measurement",
    };
    const measurementUnitID = "browser_group:measurement:measurement-timeline-grid";
    const schedulerStreamRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-measurement-scheduler-stream-"));
    try {
        const eventFile = path.join(schedulerStreamRoot, "unit-events.ndjson");
        const file = openSync(eventFile, "w", 0o600);
        let seq = 0;
        const writeEvent = (event) => {
            seq += 1;
            writeSync(file, `${JSON.stringify({
                schema_id: "cartulary.harness_unit_event.v4",
                seq,
                monotonic_ms: seq,
                resource_claims: {},
                needs: [],
                service_dependencies: [],
                status: "running",
                ...(event.event === "queued" ? { unit_kind: "runner", row_ids: [] } : {}),
                ...event,
            })}\n`);
        };
        for (let index = 0; index < 20000; index += 1) {
            writeEvent({ event: "queued", unit_id: `row:release-prefix-${index}` });
        }
        writeEvent({ event: "started", unit_id: measurementUnitID });
        for (let index = 0; index < 20000; index += 1) {
            writeEvent({ event: "completed", status: "passed", unit_id: `row:release-prefix-${index}` });
        }
        writeEvent({ event: "completed", unit_id: measurementUnitID, status: "passed" });
        closeSync(file);
        const streamedEvidence = await readMeasurementSchedulerEvidence({ file: eventFile }, measurementGroupResult);
        const previousPrefix = process.env.CARTULARY_HARNESS_UNIT_EVENT_PREFIX;
        const previousRunID = process.env.CARTULARY_TEST_RUN_ID;
        const liveEventFile = path.join(schedulerStreamRoot, ".opaque-publisher-stream");
        writeFileSync(liveEventFile, readFileSync(eventFile), { mode: 0o600 });
        const stat = lstatSync(liveEventFile, { bigint: true });
        const prefix = { schema_id: "cartulary.harness_unit_event_prefix.v1", run_id: "prefix-fixture",
            file: liveEventFile, device: stat.dev.toString(), inode: stat.ino.toString(),
            byte_length: Number(stat.size), last_seq: seq };
        const configure = (change = {}) => { process.env.CARTULARY_HARNESS_UNIT_EVENT_PREFIX = JSON.stringify({ ...prefix, ...change }); };
        try {
            process.env.CARTULARY_TEST_RUN_ID = prefix.run_id;
            configure();
            assert.deepEqual(currentUnitEventSource(schedulerStreamRoot), { file: liveEventFile, prefix });
            writeFileSync(liveEventFile, '{"partial":', { flag: "a" });
            assert.deepEqual(await readMeasurementSchedulerEvidence(currentUnitEventSource(schedulerStreamRoot), measurementGroupResult), streamedEvidence);
            await assert.rejects(readMeasurementSchedulerEvidence({ file: liveEventFile }, measurementGroupResult), /invalid JSON/u);
            configure({ run_id: "foreign" });
            assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /another run/u);
            configure();
            chmodSync(liveEventFile, 0o644);
            assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /private regular file/u);
            chmodSync(liveEventFile, 0o600);
            const alias = path.join(schedulerStreamRoot, "stream-alias");
            linkSync(liveEventFile, alias);
            assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /private regular file/u);
            rmSync(alias);
            symlinkSync(liveEventFile, alias);
            configure({ file: alias });
            assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /private regular file/u);
            rmSync(alias);
            configure({ file: path.join(schedulerStreamRoot, "nested", "stream") });
            assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /escapes its run/u);
            configure({ file: "relative-stream" });
            assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /must be absolute/u);
            configure({ file: path.join(os.tmpdir(), ".foreign-stream") });
            assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /escapes its run/u);
            delete process.env.CARTULARY_HARNESS_UNIT_EVENT_PREFIX;
            assert.deepEqual(currentUnitEventSource(schedulerStreamRoot), { file: eventFile });
        }
        finally {
            for (const [key, value] of [["CARTULARY_HARNESS_UNIT_EVENT_PREFIX", previousPrefix], ["CARTULARY_TEST_RUN_ID", previousRunID]]) {
                if (value === undefined)
                    delete process.env[key];
                else
                    process.env[key] = value;
            }
        }
        assert.equal(streamedEvidence.dependency_skipped, false);
        assert.equal(streamedEvidence.overlap_count, 0);
        assert.equal(streamedEvidence.start_seq, 20001);
        assert.equal(streamedEvidence.end_seq, 40002);
        const overlapFile = path.join(schedulerStreamRoot, "overlap-events.ndjson");
        const overlapEvents = [
            {
                event: "started",
                status: "running",
                unit_id: measurementUnitID,
            },
            { event: "started", status: "running", unit_id: "row:ordinary" },
            { event: "completed", status: "passed", unit_id: "row:ordinary" },
            { event: "completed", status: "passed", unit_id: measurementUnitID },
        ].map((entry, index) => ({
            schema_id: "cartulary.harness_unit_event.v4",
            seq: index + 1,
            monotonic_ms: index + 1,
            resource_claims: {},
            needs: [],
            service_dependencies: [],
            ...entry,
        }));
        writeFileSync(overlapFile, `${overlapEvents.map((entry) => JSON.stringify(entry)).join("\n")}\n`);
        assert.equal((await readMeasurementSchedulerEvidence({ file: overlapFile }, measurementGroupResult))
            .overlap_count, 1, "ordinary work beginning inside a measurement session must disqualify it");
        const incompleteFile = path.join(schedulerStreamRoot, "incomplete-events.ndjson");
        writeFileSync(incompleteFile, `${JSON.stringify(overlapEvents[0])}\n`);
        await assert.rejects(readMeasurementSchedulerEvidence({ file: incompleteFile }, measurementGroupResult), /lacks a closed scheduler interval/u, "incomplete scheduler proof must fail closed");
        const skippedFile = path.join(schedulerStreamRoot, "skipped-events.ndjson");
        writeFileSync(skippedFile, `${JSON.stringify({
            schema_id: "cartulary.harness_unit_event.v4",
            event: "skipped",
            failure_reason: "dependency_failure",
            monotonic_ms: 1,
            resource_claims: {},
            needs: [],
            seq: 1,
            service_dependencies: [],
            status: "skipped",
            unit_id: measurementUnitID,
        })}\n`);
        const skippedEvidence = await readMeasurementSchedulerEvidence({ file: skippedFile }, measurementGroupResult);
        assert.equal(skippedEvidence.dependency_skipped, true);
        const invalidSequenceFile = path.join(schedulerStreamRoot, "invalid-sequence-events.ndjson");
        writeFileSync(invalidSequenceFile, `${JSON.stringify({
            schema_id: "cartulary.harness_unit_event.v4",
            event: "started",
            monotonic_ms: 1,
            resource_claims: {},
            needs: [],
            seq: 2,
            service_dependencies: [],
            status: "running",
            unit_id: measurementUnitID,
        })}\n`);
        await assert.rejects(readMeasurementSchedulerEvidence({ file: invalidSequenceFile }, measurementGroupResult), /sequence 2 is not contiguous/u);
    }
    finally {
        rmSync(schedulerStreamRoot, { force: true, recursive: true });
    }
});
