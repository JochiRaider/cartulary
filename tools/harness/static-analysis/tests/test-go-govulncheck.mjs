import assert from "node:assert/strict";
import test from "node:test";
import path from "node:path";
import { validateSchemaSync } from "../../contract/index.mjs";
import { scannerFixture, authoredPackages, discoveryArgs, assertToolEnvironment, assertPrivateArtifacts } from "./go-scanner-fixtures.mjs";

const config = { config: { protocol_version: "v1.0.0", scanner_name: "govulncheck", scanner_version: "fixture", scan_level: "symbol", scan_mode: "source" } };
const finding = { finding: { osv: "GO-2099-0001", fixed_version: "v1.2.3", trace: [{ module: "example.com/vulnerable", version: "v1.2.2", package: "example.com/vulnerable", function: "Explode" }] } };
const events = (...values) => values.map((value) => JSON.stringify(value)).join("\n") + "\n";

test("Govulncheck uses authored discovery roots and forwards the database as one argument", (t) => {
  const fixture = scannerFixture(t, { output: events(config) });
  const database = `file://${fixture.root}/database with spaces`;
  const child = fixture.invoke("go-govulncheck", { GOVULNCHECK_DB: database });
  assert.equal(child.status, 0, child.stderr);
  assert.deepEqual(fixture.calls("go").map((call) => call.args), [discoveryArgs]);
  assert.deepEqual(fixture.calls().map((call) => call.args), [["-db", database, "-test", "-json", ...authoredPackages]]);
  assertToolEnvironment(fixture, fixture.calls("go")[0], false);
  assertToolEnvironment(fixture, fixture.calls()[0]);
  const summary = fixture.readJSON("artifacts/govulncheck-findings.json");
  validateSchemaSync("cartulary.govulncheck_findings.v1", summary);
  assert.equal(summary.status, "pass");
  assert.equal(summary.counts.finding_count, 0);
  assertPrivateArtifacts(fixture, ["govulncheck-output.jsonstream", "govulncheck-findings.json"]);
});
test("Govulncheck publishes schema-valid symbol findings with private modes", (t) => {
  const fixture = scannerFixture(t, { output: events(config, finding) });
  const child = fixture.invoke("go-govulncheck");
  assert.equal(child.status, 1, child.stderr);
  const summary = fixture.readJSON("artifacts/govulncheck-findings.json");
  validateSchemaSync("cartulary.govulncheck_findings.v1", summary);
  assert.equal(summary.status, "fail");
  assert.equal(summary.counts.blocking_count, 1);
  assert.deepEqual(summary.counts.reachability, { module: 0, package: 0, symbol: 1 });
  assert.deepEqual(summary.blocking_vulnerability_ids, ["GO-2099-0001"]);
  assert.equal(summary.findings[0].reachability, "symbol");
  assert.match(child.stderr, /failure_class=security failure_reason=security_finding/u);
  assertPrivateArtifacts(fixture, ["govulncheck-output.jsonstream", "govulncheck-findings.json"]);
});
test("Govulncheck treats malformed scanner output as artifact failure", (t) => {
  const fixture = scannerFixture(t, { output: "not-json\n" });
  const child = fixture.invoke("go-govulncheck");
  assert.equal(child.status, 11);
  assert.match(child.stderr, /failure_class=artifact failure_reason=artifact_error/u);
  assertPrivateArtifacts(fixture, ["govulncheck-output.jsonstream"]);
});
test("Govulncheck reports missing tooling before discovery", (t) => {
  const fixture = scannerFixture(t);
  const child = fixture.invoke("go-govulncheck", { GOVULNCHECK_BIN: path.join(fixture.root, "absent") });
  assert.equal(child.status, 2);
  assert.match(child.stderr, /failure_class=config failure_reason=configuration_error/u);
  assert.match(child.stderr, /make bootstrap/u);
  assert.deepEqual(fixture.calls("go"), []);
  assert.deepEqual(fixture.calls(), []);
});
// The Make/runtime diagnostic matrix remains in test-command-failure.mjs;
// output-publication tests remain separate consumers of that evidence.
