import assert from "node:assert/strict";
import test from "node:test";
import path from "node:path";
import { scannerFixture, assertToolEnvironment, assertPrivateArtifacts } from "./go-scanner-fixtures.mjs";

const rules = "G118,G122,G301,G302,G303,G304,G305,G306,G307";
test("advisory Gosec merges rules and roots once, bounds CPU and publishes private metadata", (t) => {
  const fixture = scannerFixture(t, { output: "fixture advisory finding\n" });
  const child = fixture.invoke("go-gosec-audit");
  assert.equal(child.status, 0, child.stderr);
  assert.match(child.stdout, /fixture advisory finding/u);
  const calls = fixture.calls();
  assert.equal(calls.length, 1);
  assert.deepEqual(calls[0].args, [`-include=${rules}`, "-exclude-generated", "-no-fail", "-quiet", "./cmd/...", "./internal/...", "./tools/..."]);
  assertToolEnvironment(fixture, calls[0]);
  assert.equal(calls[0].env.GOMAXPROCS, "4");
  assert.deepEqual(fixture.calls("go"), []);
  assert.deepEqual(fixture.profiles(), [{ tool: "gosec", target: "go-gosec-audit", profile: "repository", rules,
    flags: "-exclude-generated -no-fail -quiet", patterns: "./cmd/... ./internal/... ./tools/..." }]);
  assertPrivateArtifacts(fixture, ["security-profiles.jsonl"]);
});
test("advisory Gosec honors explicit flags and deduplicates covered support patterns", (t) => {
  const fixture = scannerFixture(t);
  const child = fixture.invoke("go-gosec-audit", { GOSEC_AUDIT_SUPPORT_FLAGS: "-exclude-generated -no-fail -terse",
    GOSEC_AUDIT_SUPPORT_PATTERNS: "./internal/testutil/... ./tools/..." });
  assert.equal(child.status, 0, child.stderr);
  assert.deepEqual(fixture.calls()[0].args, [`-include=${rules}`, "-exclude-generated", "-no-fail", "-terse", "./cmd/...", "./internal/...", "./tools/..."]);
});
test("advisory Gosec still fails on tool execution failure", (t) => {
  const fixture = scannerFixture(t, { scannerExit: 9 });
  assert.notEqual(fixture.invoke("go-gosec-audit").status, 0);
  assert.equal(fixture.calls().length, 1);
});
for (const problem of ["missing tool", "invalid CPU limit"]) {
  test(`advisory Gosec rejects ${problem} before scanning`, (t) => {
    const fixture = scannerFixture(t);
    const child = fixture.invoke("go-gosec-audit", problem === "missing tool" ? { GOSEC_BIN: path.join(fixture.root, "absent") } : { CARTULARY_SEQUENCE_HOST_CPU_LIMIT: "0" });
    assert.notEqual(child.status, 0);
    assert.match(child.stderr, problem === "missing tool" ? /requires an executable GOSEC_BIN/u : /positive scheduler CPU limit/u);
    assert.deepEqual(fixture.calls(), []);
  });
}
