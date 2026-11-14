import assert from "node:assert/strict";
import path from "node:path";
import test from "node:test";
import { spawn, spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { EventEmitter, once } from "node:events";
import { createRuntime } from "../../runtime/session-client.mjs";
import { ownedProcess, stopOwnedProcess } from "../../runtime/owned-process.mjs";
test("bootstrap artifact cache preserves content identity and failure handling", () => {
  const result = spawnSync("bash", ["tools/harness/readiness/tests/test-cache-artifact.sh"], {
    cwd: path.resolve(import.meta.dirname, "../../../.."),
    env: { ...process.env, NODE_BIN: process.execPath }, encoding: "utf8",
  });
  assert.equal(result.status, 0, result.stdout + result.stderr);
});

test("bootstrap cache owns inspection through publication across competing invocations", { timeout: 30000 }, async (t) => {
  const root = path.resolve(import.meta.dirname, "../../../..");
  const scratch = mkdtempSync(path.join(tmpdir(), "bootstrap-cache-transaction-"));
  const runRoot = path.join(scratch, "run");
  const children = [];
  let runtime;
  t.after(async () => {
    const failures = [];
    for (const entry of children) {
      try {
        if (entry.resource) {
          await stopOwnedProcess(entry.resource.target, { environment: {
            ...process.env, ...runtime.producerEnvironment(), CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT: runRoot,
          } });
          if (entry.registered) runtime.record({ ...entry.resource, state: "released" });
        } else if (entry.child.exitCode === null && entry.child.signalCode === null) entry.child.kill("SIGKILL");
        await entry.closed;
      } catch (error) { failures.push(error); }
    }
    if (failures.length) throw new AggregateError(failures, "cache process fixture cleanup failed");
    runtime?.close();
    rmSync(scratch, { recursive: true, force: true });
  });
  mkdirSync(runRoot, { mode: 0o700 });
  runtime = createRuntime({ repoRoot: root, runRoot, runID: "cache-transaction", scratchRoot: path.join(scratch, "private") });
  const bin = path.join(scratch, "bin"), cache = path.join(scratch, "cache"), output = path.join(scratch, "output");
  mkdirSync(bin);
  const producer = path.join(scratch, "producer");
  writeFileSync(producer, '#!/usr/bin/env bash\nset -euo pipefail\nprintf "producer\\n" >>"$CACHE_FIXTURE_ROOT/produced"\nprintf "complete\\n" >"$CACHE_FIXTURE_ROOT/output"\n', { mode: 0o700 });
  writeFileSync(path.join(bin, "stat"), `#!/usr/bin/env bash
set -euo pipefail
if [[ "$CACHE_FIXTURE_ID" == first && "\${!#}" == "$CACHE_FIXTURE_ROOT/output" ]]; then
  printf 'ready-to-publish\\n' >&3
  IFS= read -r release
  [[ "$release" == release ]]
fi
exec /usr/bin/stat "$@"
`, { mode: 0o700 });
  writeFileSync(path.join(bin, "flock"), `#!/usr/bin/env bash
set -euo pipefail
if [[ "$CACHE_FIXTURE_ID" == second ]]; then printf 'lock-attempt\\n' >&3; fi
exec /usr/bin/flock "$@"
`, { mode: 0o700 });
  const args = [path.join(root, "tools/harness/readiness/cache-artifact.sh"), "--scope", "build-artifact", "--profile", "transaction", "--cache-dir", cache, "--output", output, "--", producer];
  const start = (id) => {
    const child = spawn("bash", args, { detached: true, stdio: ["pipe", "ignore", "pipe", "pipe"], env: {
      ...process.env, PATH: `${bin}:${process.env.PATH}`, CACHE_FIXTURE_ROOT: scratch, CACHE_FIXTURE_ID: id,
    } });
    const entry = { child, lines: [], events: new EventEmitter(), errors: "", resource: null };
    children.push(entry);
    let spawnError, pending = "";
    child.once("error", (error) => { spawnError = error; });
    child.stdin.on("error", () => {});
    entry.closed = new Promise((resolve) => child.once("close", (code, signal) => resolve({ code, signal, spawnError })));
    child.stderr.on("data", (bytes) => { entry.errors += bytes; });
    child.stdio[3].on("data", (bytes) => {
      pending += bytes;
      while (pending.includes("\n")) {
        const index = pending.indexOf("\n"), line = pending.slice(0, index);
        pending = pending.slice(index + 1);
        entry.lines.push(line);
        entry.events.emit("line");
      }
    });
    if (child.pid) {
      const resource = { kind: "helper_process", target: ownedProcess(child.pid), state: "acquired", run_id: runtime.runID, lease_id: runtime.leaseID, depends_on: [], recovery_files: [] };
      entry.resource = resource;
      runtime.record(resource);
      entry.registered = true;
    }
    return entry;
  };
  const waitLine = async (entry, line) => {
    while (!entry.lines.includes(line)) {
      await Promise.race([
        once(entry.events, "line", { signal: AbortSignal.timeout(10000) }),
        entry.closed.then(() => { throw new Error(`fixture exited before ${line}: ${entry.errors}`); }),
      ]);
    }
  };
  const first = start("first");
  await waitLine(first, "ready-to-publish");
  const probe = spawnSync("/usr/bin/flock", ["--nonblock", path.join(cache, "transaction"), "true"], { encoding: "utf8", timeout: 1000 });
  assert.equal(probe.error, undefined);
  assert.equal(probe.status, 1, "output proof must remain under the transaction lock after producer exit");
  const second = start("second");
  await waitLine(second, "lock-attempt");
  first.child.stdin.end("release\n");
  for (const entry of [first, second]) {
    const result = await entry.closed;
    assert.equal(result.spawnError, undefined);
    assert.equal(result.signal, null, entry.errors);
    assert.equal(result.code, 0, entry.errors);
  }
  assert.equal(readFileSync(path.join(scratch, "produced"), "utf8"), "producer\n", "shared cold key executes only one producer");
  assert.equal(readFileSync(output, "utf8"), "complete\n");
});
