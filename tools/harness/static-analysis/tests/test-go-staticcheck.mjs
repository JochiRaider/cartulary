import assert from "node:assert/strict";
import test from "node:test";
import path from "node:path";
import { repository } from "./fixtures.mjs";
import { scannerFixture, analyzerDirectory, analyzerPackage, authoredPackages, discoveryArgs, assertToolEnvironment } from "./go-scanner-fixtures.mjs";

for (const [tool, checks] of [["go-vet", null], ["go-staticcheck", null], ["go-staticcheck", "SA*"]]) {
  test(`${tool} scans both modules with ${checks ?? "repository defaults"}`, (t) => {
    const fixture = scannerFixture(t);
    const child = fixture.invoke(tool, checks ? { STATICCHECK_CHECKS: checks } : {});
    assert.equal(child.status, 0, child.stderr);
    const go = fixture.calls("go");
    const discovery = go.filter((call) => call.args[0] === "list");
    const scans = tool === "go-vet" ? go.filter((call) => call.args[0] === "vet") : fixture.calls();
    const prefix = tool === "go-vet" ? ["vet"] : checks ? [`-checks=${checks}`] : [];
    assert.deepEqual(discovery.map((call) => call.args), [discoveryArgs, ["list", "./..."]]);
    assert.deepEqual(scans.map((call) => call.args), [[...prefix, ...authoredPackages], [...prefix, analyzerPackage]]);
    for (const pair of [discovery, scans]) {
      assert.deepEqual(pair.map((call) => call.cwd), [repository, analyzerDirectory]);
      for (const call of pair) assertToolEnvironment(fixture, call, false);
      assert.equal(pair[1].env.GOWORK, "off");
      assert.equal(pair[1].env.GOPROXY, "off");
      assert.equal(pair[1].env.GOFLAGS, "-mod=readonly");
    }
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

for (const tool of ["go-vet", "go-staticcheck"]) {
  for (const module of ["root", "analysis"]) {
    for (const [name, packages, goExit] of [
      ["failed empty", [], 9],
      ["failed partial", [analyzerPackage], 9],
      ["empty", [], 0],
      ["generated-only", ["github.com/JochiRaider/cartulary/internal/gen/contracts"], 0],
    ]) {
      test(`${tool} rejects ${name} discovery in ${module}`, (t) => {
        const config = { packages, goExit };
        const fixture = scannerFixture(t, module === "root" ? config : { analysis: config });
        const child = fixture.invoke(tool);
        assert.notEqual(child.status, 0);
        assert.match(child.stderr, /package discovery/u);
        const go = fixture.calls("go");
        const scans = tool === "go-vet" ? go.filter((call) => call.args[0] === "vet") : fixture.calls();
        assert.deepEqual(scans.map((call) => call.cwd), module === "root" ? [] : [repository]);
      });
    }
    test(`${tool} propagates a finding in ${module}`, (t) => {
      const config = { vetExit: 9, scannerExit: 9 };
      const fixture = scannerFixture(t, module === "root" ? config : { analysis: config });
      assert.equal(fixture.invoke(tool).status, 9);
      const scans = tool === "go-vet" ? fixture.calls("go").filter((call) => call.args[0] === "vet") : fixture.calls();
      assert.deepEqual(scans.map((call) => call.cwd), module === "root" ? [repository] : [repository, analyzerDirectory]);
    });
  }
}
