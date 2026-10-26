import assert from "node:assert/strict";
import test from "node:test";
import path from "node:path";
import { scannerFixture, assertToolEnvironment, assertPrivateArtifacts } from "./go-scanner-fixtures.mjs";

const general = "G602,G124,G112,G114";
const runtime = "G122,G301,G302,G303,G304,G305,G306,G307";
test("targeted Gosec preserves blocking profiles, authored scope and private metadata", (t) => {
  const fixture = scannerFixture(t);
  const child = fixture.invoke("go-gosec-targeted");
  assert.equal(child.status, 0, child.stderr);
  const calls = fixture.calls();
  assert.equal(calls.length, 2);
  assert.deepEqual(calls[0].args, [`-include=${general}`, "-exclude-generated", "./cmd/...", "./internal/...", "./db/...", "./tools/..."]);
  assert.equal(calls[1].args[0], `-include=${runtime}`);
  assert.deepEqual(calls[1].args.slice(-2), ["./cmd/...", "./internal/..."]);
  for (const flag of ["-exclude-generated", "-quiet", ...["internal/testutil", "internal/modules/auth/testsupport", "internal/modules/workbook/testsupport", "internal/modules/records/testsupport"].map((root) => `-exclude-dir=${root}`)]) assert.ok(calls[1].args.includes(flag), flag);
  assert.ok(!calls[1].args.includes("-exclude-dir=internal/modules/networkflow/harnesscontrol"));
  for (const call of calls) { assert.ok(!call.args.includes("-no-fail")); assertToolEnvironment(fixture, call); }
  assert.deepEqual(fixture.calls("go"), [], "Gosec owns its package expansion");
  const profiles = fixture.profiles();
  assert.deepEqual(profiles.map(({ tool, target, profile, rules, patterns }) => ({ tool, target, profile, rules, patterns })), [
    { tool: "gosec", target: "go-gosec-targeted", profile: "general", rules: general, patterns: "./cmd/... ./internal/... ./db/... ./tools/..." },
    { tool: "gosec", target: "go-gosec-targeted", profile: "runtime", rules: runtime, patterns: "./cmd/... ./internal/..." },
  ]);
  for (const [index, profile] of profiles.entries()) assert.deepEqual(profile.flags.split(" "), calls[index].args.slice(1, -profile.patterns.split(" ").length));
  assertPrivateArtifacts(fixture, ["security-profiles.jsonl"]);
});
test("targeted Gosec forwards explicit flags and fails on runtime findings", (t) => {
  const fixture = scannerFixture(t, { scannerExit: 42, failArgument: `-include=${runtime}`, output: "fixture finding\n" });
  const child = fixture.invoke("go-gosec-targeted", { GOSEC_FLAGS: "-exclude-generated -quiet" });
  assert.notEqual(child.status, 0);
  assert.equal(fixture.calls().length, 2);
  assert.ok(fixture.calls()[0].args.includes("-quiet"));
  assert.match(child.stdout, /fixture finding/u);
});
test("targeted Gosec rejects a missing scanner", (t) => {
  const fixture = scannerFixture(t);
  const child = fixture.invoke("go-gosec-targeted", { GOSEC_BIN: path.join(fixture.root, "absent") });
  assert.notEqual(child.status, 0);
  assert.match(child.stderr, /requires an executable GOSEC_BIN/u);
  assert.match(child.stderr, /make bootstrap/u);
  assert.deepEqual(fixture.calls(), []);
});
