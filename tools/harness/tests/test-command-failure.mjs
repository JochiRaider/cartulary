import { createLaunchContext } from "../runtime/launch-context.mjs";
import { spawnSync } from "node:child_process";
import { WorkGraphCompiler } from "../scheduler/work-graph/index.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync, symlinkSync, chmodSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createRuntime } from "../runtime/session-client.mjs";
import { createCommandFailureContext, publishCommandFailure, readCommandFailure } from "../runtime/command-failure.mjs";
import { adaptShellInvocation } from "../execution/runners/shell.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const fixture = path.join(root, "tools/harness/tests/command-failure-fixture.mjs");
const artifactFailure = { failure_class: "artifact", failure_reason: "artifact_error" };
const accounting = { failure_class: "harness", failure_reason: "scheduler_accounting_error" };
function setup(t) {
  const scratch = mkdtempSync(path.join(os.tmpdir(), "command-failure-test-"));
  const runRoot = path.join(scratch, "run"); mkdirSync(runRoot, { mode: 0o700 });
  const runtime = createRuntime({ repoRoot: root, runRoot, runID: "run", scratchRoot: path.join(scratch, "private") });
  const environment = { CARTULARY_TEST_RESULTS_DIR: scratch, CARTULARY_TEST_RUN_ID: "run", CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "run" };
  const contexts=[];
  t.after(()=>{for(const context of contexts)context.close();runtime.close();rmSync(scratch,{recursive:true,force:true});});
  const create = () => {
    const context = createCommandFailureContext({ ...{ repoRoot: root, environment, unitID: "target:build-web", commandID: "cartulary.harness.command.build_web.v2" }, launch: createLaunchContext({ repoRoot: root, environment, unitID: "target:build-web", commandID: "cartulary.harness.command.build_web.v2" }) });
    contexts.push(context); return context;
  };
  return { scratch, runRoot, runtime, environment, create };
}
export function registerCommandFailureTests() {
  test("repeated command failure contexts allocate independent invocation identities", (t) => {
    const { create } = setup(t);
    const first = create(), second = create();
    const key = "CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT";
    const a = JSON.parse(first.environment[key]), b = JSON.parse(second.environment[key]);
    assert.notEqual(a.invocation_id, b.invocation_id);
    assert.equal(a.run_id, b.run_id);
    assert.equal(a.unit_id, b.unit_id);
    first.close();
    assert.equal(second.read(), null);
  });
  test("command failure envelopes bind identity and reject corruption", (t) => {
    const { scratch, runtime, environment, create } = setup(t);
    for (const mutation of ["valid", "identity", "malformed", "conflict", "permission", "symlink", "partial"]) {
      const context = create();
      const env = { ...environment, ...context.environment };
      assert.equal(context.read(), null);
      publishCommandFailure(root, artifactFailure, env);
      const identity = JSON.parse(env.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT);
      const file = runtime.privatePath(`command-failure-${identity.invocation_id}`, "failure.json");
      if (mutation === "identity") writeFileSync(file, JSON.stringify({ ...JSON.parse(readFileSync(file)), unit_id: "another-unit" }));
      if (mutation === "malformed") writeFileSync(file, '{"schema_id":');
      if (mutation === "conflict") assert.throws(() => publishCommandFailure(root, accounting, env), /conflicting/u);
      if (mutation === "permission") chmodSync(file, 0o644);
      if (mutation === "symlink") { rmSync(file); symlinkSync(path.join(scratch, "absent"), file); }
      if (mutation === "partial") rmSync(file);
      assert.deepEqual(context.read(), mutation === "valid" ? artifactFailure : accounting, mutation);
      assert.deepEqual(readCommandFailure(root, env), context.read());
      if (mutation === "valid") {
        publishCommandFailure(root, artifactFailure, env);
        assert.deepEqual(context.read(), artifactFailure, "identical diagnostic is idempotent");
        assert.equal(lstatSync(file).mode & 0o777, 0o600);
      }
      context.close();
    }
  });
  test("shell classification preserves diagnostics and accounts for contradictory success", () => {
    const invocation = { rows: [{ row_id: "example" }] };
    assert.equal(adaptShellInvocation(invocation, { status: 2, commandFailure: artifactFailure })[0].failure_reason, "artifact_error");
    assert.equal(adaptShellInvocation(invocation, { status: 0, commandFailure: artifactFailure })[0].failure_reason, "scheduler_accounting_error");
    assert.equal(adaptShellInvocation(invocation, { status: 1 })[0].failure_reason, "test_assertion_failure");

  });
  test("Govulncheck producer preserves scanner causes through its private diagnostic channel", { timeout: 60000 }, async (t) => {
    const { scratch, runtime, environment } = setup(t);
    const compiler = new WorkGraphCompiler(root);
    const canonical = compiler.compile({ kind: "target", target: "go-vulncheck" }).units.find((unit) => unit.unit_id === "target:go-vulncheck");
    assert.deepEqual(compiler.compile({ kind: "aggregate", target: "check" }).units.find((unit) => unit.unit_id === canonical.unit_id), canonical, "direct and aggregate selection use the same scanner unit");
    const fakeGo = path.join(scratch, "go");
    const scanner = path.join(scratch, "govulncheck");
    const payload = path.join(scratch, "scanner-output");
    const invoked = path.join(scratch, "scanner-invoked");
    const makefile = path.join(scratch, "Makefile");
    writeFileSync(fakeGo, '#!/bin/sh\nprintf "%s\\n" github.com/JochiRaider/cartulary/cmd/server\nexit "${FAKE_GO_EXIT:-0}"\n', { mode: 0o700 });
    writeFileSync(scanner, '#!/bin/sh\nprintf invoked >"$FAKE_SCANNER_INVOKED"\nif [ "$FAKE_SCANNER_CONFLICT" = 1 ]; then "$NODE_BIN" "$FAKE_DIAGNOSTIC_FIXTURE" conflict >&2; fi\ncat "$FAKE_SCANNER_OUTPUT"\nexit "$FAKE_SCANNER_EXIT"\n', { mode: 0o700 });
    writeFileSync(makefile, `go-vulncheck:\n\t@bash "${root}/tools/harness/static-analysis/go-govulncheck.sh"\n`);
    const config = { config: { protocol_version: "v1.0.0", scanner_name: "govulncheck", scanner_version: "v1.3.0", scan_level: "symbol", scan_mode: "source" } };
    const finding = (reachability) => ({ finding: { osv: "GO-2099-0001", fixed_version: "v1.2.3", trace: [{ module: "example.com/vulnerable", version: "v1.2.2", ...(reachability !== "module" ? { package: "example.com/vulnerable" } : {}), ...(reachability === "symbol" ? { function: "Explode" } : {}) }] } });
    const events = (...values) => values.map((value) => JSON.stringify(value)).join("\n");
    const scenarios = [
      { name: "clean", output: events(config) },
      { name: "diagnostic", output: events(config, finding("module"), finding("package")), diagnostic: true },
      { name: "blocking", output: events(config, finding("symbol")), failure: ["security", "security_finding", 1] },
      { name: "blocking-nonzero", output: events(config, finding("symbol")), scannerExit: 3, failure: ["security", "security_finding", 1] },
      { name: "malformed", output: "not-json", failure: ["artifact", "artifact_error", 11] },
      { name: "missing-findings", output: "", failure: ["artifact", "artifact_error", 11] },
      { name: "scanner-failure", output: events(config), scannerExit: 9, failure: ["harness", "tool_diagnostic_failure", 1] },
      { name: "scanner-failure-no-output", output: "", scannerExit: 9, failure: ["harness", "tool_diagnostic_failure", 1] },
      { name: "missing-tool", output: "", missingTool: true, failure: ["config", "configuration_error", 2] },
      { name: "package-discovery-failure", output: "", goExit: 9, failure: ["harness", "tool_diagnostic_failure", 1] },
      { name: "conflicting-diagnostics", output: events(config, finding("symbol")), conflict: true, failure: ["harness", "scheduler_accounting_error", 11] },
    ];
    for (const scenario of scenarios) {
      writeFileSync(payload, scenario.output);
      rmSync(invoked, { force: true });
      const artifacts = path.join(scratch, scenario.name);
      const unit = {
        ...canonical, needs: [], cache_policy: "none", resource_claims: { cpu: 1 }, timeout_ms: 10000,
        command: { executable: "make", args: ["--silent", "--no-print-directory", "-f", makefile, "go-vulncheck"], environment: {
          CARTULARY_TEST_TARGET: "go-vulncheck", CARTULARY_HARNESS_CACHE_MODE: "off", NODE_BIN: process.execPath,
          GO: fakeGo, GOVULNCHECK_BIN: scenario.missingTool ? path.join(scratch, "absent") : scanner,
          GO_CACHE_DIR: path.join(scratch, "cache"), GO_MOD_CACHE_DIR: path.join(scratch, "modules"), GO_TMP_DIR: scratch,
          CARTULARY_STEP_ARTIFACT_DIR: artifacts, FAKE_SCANNER_OUTPUT: payload, FAKE_SCANNER_INVOKED: invoked,
          FAKE_SCANNER_EXIT: String(scenario.scannerExit ?? 0), FAKE_GO_EXIT: String(scenario.goExit ?? 0),
          FAKE_SCANNER_CONFLICT: scenario.conflict ? "1" : "0", FAKE_DIAGNOSTIC_FIXTURE: fixture,
        } },
      };
      const launch = createLaunchContext({repoRoot: root, environment, unitID: unit.unit_id, commandID: "cartulary.harness.command.go_vulncheck.v2", producer: "unit"});
      const diagnostic = createCommandFailureContext({repoRoot: root, environment, launch});
      try {
        const env = {...process.env, ...environment, ...unit.command.environment, ...launch.environment, ...diagnostic.environment};
        for (const key of ["MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS"]) delete env[key];
        const outcome = spawnSync(unit.command.executable, unit.command.args, {cwd: root, env, encoding: "utf8", timeout: 10000});
        assert.equal(outcome.error, undefined);
        if (scenario.failure) {
          const failure = diagnostic.read();
          assert.deepEqual([failure.failure_class, failure.failure_reason], scenario.failure.slice(0, 2), scenario.name);
          assert.notEqual(outcome.status, 0);
          assert.match(outcome.stderr, /Error (1|2|11)/u, "GNU Make wrapped the underlying failure");
        } else {
          assert.equal(outcome.status, 0, outcome.stderr);
          assert.equal(diagnostic.read(), null);
        }
      } finally { diagnostic.close(); }
      assert.equal(Boolean(lstatSync(invoked, { throwIfNoEntry: false })), !scenario.missingTool && !scenario.goExit, "fixtures execute freshly; discovery failures stop before scanning");
      if (scenario.diagnostic) {
        const findings = JSON.parse(readFileSync(path.join(artifacts, "govulncheck-findings.json")));
        assert.equal(findings.counts.blocking_count, 0);
        assert.deepEqual(findings.counts.reachability, { module: 1, package: 1, symbol: 0 });
      }
      assert.equal(readdirSync(path.join(runtime.root, "detail")).some((name) => name.startsWith("command-failure-")), false, "each scanner invocation releases its private diagnostic channel");
    }
  });

  test("canonical shell row preserves artifact diagnostics and removes private captures", (t) => {
    const { scratch, runtime, runRoot, environment } = setup(t);
    const shim = path.join(scratch, "make-probe");
    writeFileSync(shim, `#!/bin/sh\ncase "$*" in\n  *protocol-ts-browser-artifact-reachability*) exec ${process.execPath} ${fixture} diagnostic ;;\n  *) exec env MAKE=make make "$@" ;;\nesac\n`, { mode: 0o700 });
    const rowID = "package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b";
    const env = {...process.env, ...environment, CARTULARY_TEST_TARGET: "protocol-ts-browser-artifact-reachability", MAKE: shim};
    for (const key of ["MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS"]) delete env[key];
    const result = spawnSync(process.execPath, ["tools/harness/execution/runners/row-runner-cli.mjs", "--row-id", rowID], {cwd: root, env, encoding: "utf8", timeout: 10000});
    assert.equal(result.error, undefined);
    assert.equal(result.status, 11, result.stderr);
    const row = JSON.parse(readFileSync(path.join(runRoot, "rows", `${rowID}.json`)));
    assert.equal(row.failure_reason, "artifact_error");
    assert.deepEqual(readdirSync(runtime.privatePath("child-captures")), []);
    assert.equal(readdirSync(path.join(runtime.root, "detail")).some((name) => name.startsWith("command-failure-")), false);
  });

}
