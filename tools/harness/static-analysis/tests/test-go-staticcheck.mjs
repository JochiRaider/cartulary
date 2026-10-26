import assert from "node:assert/strict";
import test from "node:test";
import path from "node:path";
import { scannerFixture, authoredPackages, discoveryArgs, assertToolEnvironment } from "./go-scanner-fixtures.mjs";

for (const checks of [null, "SA*"]) {
  test(`Staticcheck scans authored roots with ${checks ?? "repository defaults"}`, (t) => {
    const fixture = scannerFixture(t);
    const child = fixture.invoke("go-staticcheck", checks === null ? {} : { STATICCHECK_CHECKS: checks });
    assert.equal(child.status, 0, child.stderr);
    assert.deepEqual(fixture.calls("go").map((call) => call.args), [discoveryArgs]);
    assert.deepEqual(fixture.calls().map((call) => call.args), [[...(checks ? [`-checks=${checks}`] : []), ...authoredPackages]]);
    assertToolEnvironment(fixture, fixture.calls("go")[0], false);
    assertToolEnvironment(fixture, fixture.calls()[0], false);
  });
}
test("Staticcheck requires its tool before discovery", (t) => {
  const fixture = scannerFixture(t);
  const child = fixture.invoke("go-staticcheck", { STATICCHECK_BIN: path.join(fixture.root, "absent") });
  assert.notEqual(child.status, 0);
  assert.match(child.stderr, /requires an executable STATICCHECK_BIN/u);
  assert.match(child.stderr, /make bootstrap/u);
  assert.deepEqual(fixture.calls("go"), []);
  assert.deepEqual(fixture.calls(), []);
});
test("Staticcheck propagates scanner failure", (t) => {
  const fixture = scannerFixture(t, { scannerExit: 9 });
  assert.notEqual(fixture.invoke("go-staticcheck").status, 0);
  assert.equal(fixture.calls().length, 1);
});

for (const [name, packages, goExit] of [
  ["failed discovery without output", [], 9],
  ["failed discovery with partial output", [authoredPackages[0]], 9],
  ["successful empty discovery", [], 0],
  ["generated-only discovery", ["github.com/JochiRaider/cartulary/internal/gen/contracts"], 0],
]) {
  test(`Staticcheck stops before scanning after ${name}`, (t) => {
    const fixture = scannerFixture(t, { packages, goExit });
    const child = fixture.invoke("go-staticcheck");
    assert.notEqual(child.status, 0, "discovery cannot produce a successful partial scan");
    assert.match(child.stderr, /package discovery/u);
    assert.deepEqual(fixture.calls("go").map((call) => call.args), [discoveryArgs]);
    assert.deepEqual(fixture.calls(), [], "the scanner must never run with incomplete or empty discovery");
  });
}
