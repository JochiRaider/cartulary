import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import test from "node:test";
import { buildMeasurementFixture, measurementContext, measurementObservation, measurementReport } from "./measurement-fixtures.mjs";
test("measurement fixture support imports without acquiring resources or launching work", () => {
    const source = `
    import fs from "node:fs";
    import childProcess from "node:child_process";
    import { syncBuiltinESMExports } from "node:module";
    const reject = () => { throw new Error("fixture import performed work"); };
    for (const key of ["mkdtempSync", "mkdirSync", "writeFileSync", "renameSync", "rmSync"]) fs[key] = reject;
    for (const key of ["spawn", "spawnSync", "exec", "execSync", "execFile", "execFileSync"]) childProcess[key] = reject;
    syncBuiltinESMExports();
    await import(${JSON.stringify(new URL("./measurement-fixtures.mjs", import.meta.url).href)});
  `;
    const child = spawnSync(process.execPath, ["--input-type=module", "-e", source], { encoding: "utf8" });
    assert.equal(child.status, 0, child.stderr);
    assert.equal(child.stdout, "");
});
test("measurement construction owns fresh data and never validates away corruption", (t) => {
    const { fixtureProfile } = measurementContext(t);
    const first = buildMeasurementFixture(fixtureProfile);
    const second = buildMeasurementFixture(fixtureProfile);
    first.summary.rollup.sample_count = 0;
    first.files["snapshot-lease.json"].fill(0);
    assert.equal(second.summary.rollup.sample_count, 100);
    assert.equal(JSON.parse(second.files["snapshot-lease.json"]).cleanup_state, "complete");
    const observation = measurementObservation();
    observation.samples[0].total_ms = -1;
    assert.equal(measurementObservation().samples[0].total_ms, 10);
    const attachment = measurementReport(observation).suites[0].specs[0].tests[0].results[0].attachments[0];
    assert.equal(JSON.parse(Buffer.from(attachment.body, "base64")).samples[0].total_ms, -1);
});
