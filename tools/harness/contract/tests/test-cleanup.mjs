#!/usr/bin/env node
import assert from "node:assert/strict";
import { spawn, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { chmodSync, cpSync, existsSync, lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, readlinkSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { once } from "node:events";
import test from "node:test";
import { renderLayout, validateLayout } from "../../workspace/layout.mjs";

const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../../..");
const fixtureEnvironment = (overrides = {}) => ({ PATH: process.env.PATH, HOME: process.env.HOME, LC_ALL: "C", GIT_OPTIONAL_LOCKS: "0", ...overrides });
function fixture(t) {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-cleanup-"));
  t.after(() => {
    const repair = (directory) => { if (!existsSync(directory) || lstatSync(directory).isSymbolicLink()) return; chmodSync(directory, 0o700); for (const item of readdirSync(directory, { withFileTypes: true })) if (item.isDirectory()) repair(path.join(directory, item.name)); };
    repair(root); rmSync(root, { recursive: true, force: true });
  });
  for (const file of ["Makefile", "tools/task_surface.generated.mk", "tools/task_surface.runtime.generated.mk", "tools/workspace_layout.json", "tools/workspace_layout.generated.sh", "tools/workspace_layout.generated.mk", "tools/schemas/cartulary.workspace_layout.v1.schema.json", "tools/harness/workspace"]) {
    mkdirSync(path.dirname(path.join(root, file)), { recursive: true }); cpSync(path.join(repo, file), path.join(root, file), { recursive: true });
  }
  const git = (...args) => { const result = spawnSync("git", ["-C", root, ...args], { encoding: "utf8" }); assert.equal(result.status, 0, result.stderr); return result.stdout; };
  git("init", "--quiet"); git("config", "user.email", "fixture@example.invalid"); git("config", "user.name", "Fixture");
  writeFileSync(path.join(root, "source"), "source\n"); git("add", "source"); git("-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "fixture");
  const write = (file, content = "fixture") => { mkdirSync(path.dirname(path.join(root, file)), { recursive: true }); writeFileSync(path.join(root, file), content); };
  const run = (args, env = {}) => spawnSync("make", ["--no-print-directory", ...args], { cwd: root, encoding: "utf8", timeout: 15000, env: fixtureEnvironment(env) });
  return { root, git, write, run };
}
function snapshot(root) {
  const records = [];
  function walk(relative) {
    const absolute = path.join(root, relative), stat = lstatSync(absolute);
    records.push([relative, stat.mode, stat.isSymbolicLink() ? readlinkSync(absolute) : stat.isFile() ? createHash("sha256").update(readFileSync(absolute)).digest("hex") : "directory"]);
    if (stat.isDirectory()) for (const name of readdirSync(absolute).sort()) walk(path.join(relative, name));
  }
  walk(""); return records;
}
function success(result) { assert.equal(result.status, 0, result.stderr || result.stdout); }
function rejected(result, reason) { assert.notEqual(result.status, 0); assert.match(result.stderr, new RegExp(reason)); }
function utility(f, name, body) {
  f.write(`utilities/${name}`, `#!/bin/bash\nset -e\n${body}\n`);
  chmodSync(path.join(f.root, `utilities/${name}`), 0o755);
  return { PATH: `${f.root}/utilities:${process.env.PATH}` };
}
function outputBudget(result) {
  assert.ok(Buffer.byteLength(result.stdout) <= 20000, "stdout byte budget");
  assert.ok(result.stdout.trimEnd().split("\n").length <= 200, "stdout line budget");
  assert.ok(Buffer.byteLength(result.stderr) <= 6000, "stderr byte budget");
  assert.ok(result.stderr.trimEnd().split("\n").length <= 35, "stderr line budget");
}

test("ordinary tree growth does not grow coordinator memory or utility launches", (t) => {
  const measurements = [];
  for (const count of [128, 8192]) {
    const f = fixture(t);
    for (let index = 0; index < count; index++) f.write(`tmp/d${index}/file`);
    let environment;
    for (const command of ["stat", "find", "chmod"]) {
      environment = utility(f, command, `
printf '${command}\\n' >> trace
while read -r key value rest; do
  if [[ "$key" == VmRSS: ]]; then printf '%s\\n' "$value" >> memory; fi
done < "/proc/$PPID/status"
exec /usr/bin/${command} "$@"`);
    }
    const result = f.run(["clean"], environment); success(result); outputBudget(result);
    assert.equal(result.stderr, "");
    assert.match(result.stdout, /phase=validating[\s\S]*phase=repairing[\s\S]*phase=deleting[\s\S]*CLEANED[\s\S]*phase=complete/u);
    const calls = readFileSync(path.join(f.root, "trace"), "utf8").trim().split("\n");
    measurements.push({
      stat: calls.filter((name) => name === "stat").length,
      find: calls.filter((name) => name === "find").length,
      chmod: calls.filter((name) => name === "chmod").length,
      memoryKiB: Math.max(...readFileSync(path.join(f.root, "memory"), "utf8").trim().split("\n").map(Number)),
    });
  }
  assert.equal(measurements[0].stat, measurements[1].stat);
  assert.ok(measurements[1].stat <= 8, JSON.stringify(measurements));
  assert.equal(measurements[0].find, measurements[1].find);
  assert.equal(measurements[1].chmod, 0);
  assert.ok(measurements[1].memoryKiB < measurements[0].memoryKiB + 16384, JSON.stringify(measurements));
  t.diagnostic(`128 and 8192 ordinary directories: ${JSON.stringify(measurements)}`);
});

test("permission repair uses bounded count and byte batches with exact identity checks", (t) => {
  for (const longNames of [false, true]) {
    const f = fixture(t), count = 300;
    for (let index = 0; index < count; index++) {
      const directory = `tmp/${longNames ? "x".repeat(200) + "/" : ""}d${index}${longNames ? "y".repeat(100) : ""}`;
      f.write(`${directory}/file`);
      chmodSync(path.join(f.root, directory), 0o500);
    }
    const environment = utility(f, "chmod", `
count=0; bytes=0
for item in "\${@:3}"; do count=$((count+1)); bytes=$((bytes+\${#item}+1)); done
printf '%s %s\\n' "$count" "$bytes" >> batches
exec /usr/bin/chmod "$@"`);
    const result = f.run(["clean"], environment); success(result);
    const batches = readFileSync(path.join(f.root, "batches"), "utf8").trim().split("\n").map((line) => line.split(" ").map(Number));
    assert.equal(batches.reduce((sum, batch) => sum + batch[0], 0), count);
    assert.ok(batches.every(([size, bytes]) => size > 0 && size <= 128 && bytes <= 32768), JSON.stringify(batches));
    assert.equal(batches.length, longNames ? 3 : Math.ceil(count / 128));
    assert.equal(existsSync(path.join(f.root, "tmp")), false);
  }
});

test("partial or failed stream producers cannot authorize mutation", (t) => {
  for (const scenario of ["failed-find", "truncated-find", "failed-stat", "truncated-stat", "surplus-stat"]) {
    const f = fixture(t);
    f.write("tmp/readonly/file"); chmodSync(path.join(f.root, "tmp/readonly"), 0o500);
    const scripts = {
      "failed-find": ["find", `printf 'repair\\0'; exit 7`],
      "truncated-find": ["find", `printf 'repair'; exit 0`],
      "failed-stat": ["stat", `/usr/bin/stat "$@"; exit 7`],
      "truncated-stat": ["stat", `printf '1:2:700:1000:41c0'; exit 0`],
      "surplus-stat": ["stat", `/usr/bin/stat "$@"; printf '1:2:700:1000:41c0\\0'`],
    };
    const environment = utility(f, ...scripts[scenario]);
    const before = snapshot(path.join(f.root, "tmp"));
    const result = f.run(["clean"], environment); rejected(result, "cleanup_error"); outputBudget(result);
    assert.deepEqual(snapshot(path.join(f.root, "tmp")), before, scenario);
    assert.doesNotMatch(result.stdout, /CLEANED/u);
  }
});

test("root replacement and repair symlink substitution fail closed", (t) => {
  for (const scenario of ["root", "repair"]) {
    const f = fixture(t);
    f.write("tmp/readonly/file"); chmodSync(path.join(f.root, "tmp/readonly"), 0o500);
    f.write("outside/keep"); chmodSync(path.join(f.root, "outside"), 0o500);
    const environment = utility(f, "stat", scenario === "root" ? `
for item in "$@"; do
  if [[ "$item" == tmp ]]; then
    if [[ -e inspected && ! -e replaced ]]; then
      /usr/bin/mv tmp original; /usr/bin/mkdir tmp; : > replaced
    fi
    : > inspected
  fi
done
exec /usr/bin/stat "$@"` : `
for item in "$@"; do
  if [[ "$item" == tmp/readonly ]]; then
    /usr/bin/chmod u+w tmp/readonly
    /usr/bin/mv tmp/readonly original
    /usr/bin/chmod 500 original
    /usr/bin/ln -s "$PWD/outside" tmp/readonly
  fi
done
exec /usr/bin/stat "$@"`);
    const result = f.run(["clean"], environment); rejected(result, `${scenario} identity changed`);
    assert.doesNotMatch(result.stdout, /CLEANED/u);
    assert.equal(readFileSync(path.join(f.root, "outside/keep"), "utf8"), "fixture");
    assert.equal(lstatSync(path.join(f.root, "outside")).mode & 0o777, 0o500);
    assert.equal(lstatSync(path.join(f.root, scenario === "root" ? "original/readonly" : "original")).mode & 0o777, 0o500);
  }
});

test("failed permission repair stops removal and reports incomplete work", (t) => {
  const f = fixture(t);
  f.write("tmp/readonly/file"); chmodSync(path.join(f.root, "tmp/readonly"), 0o500);
  const result = f.run(["clean"], utility(f, "chmod", "exit 7"));
  rejected(result, "permission repair failed");
  assert.match(result.stderr, /completed=0.*phase=repairing/u); outputBudget(result);
  assert.equal(existsSync(path.join(f.root, "tmp/readonly/file")), true);
  assert.doesNotMatch(result.stdout, /CLEANED/u);
});

test("repair producer failure after complete records remains an execution failure", (t) => {
  const f = fixture(t);
  for (let index = 0; index < 300; index++) {
    f.write(`tmp/d${index}/file`); chmodSync(path.join(f.root, `tmp/d${index}`), 0o500);
  }
  const environment = utility(f, "find", `
/usr/bin/find "$@"
for item in "$@"; do
  if [[ "$item" == *'%D:'* ]]; then exit 7; fi
done`);
  const result = f.run(["clean"], environment);
  rejected(result, "repair enumeration failed"); outputBudget(result);
  assert.match(result.stderr, /completed=0.*phase=repairing/u);
  assert.doesNotMatch(result.stdout, /CLEANED/u);
  for (let index = 0; index < 300; index++) assert.equal(existsSync(path.join(f.root, `tmp/d${index}/file`)), true);
});

test("cancellation settles owned utility groups during validation, repair, and deletion", async (t) => {
  for (const signal of ["SIGINT", "SIGTERM"]) {
    for (const [command, phase] of [["find", "validating"], ["chmod", "repairing"], ["rm", "deleting"]]) {
      const f = fixture(t);
      f.write("tmp/readonly/file");
      if (command === "chmod") chmodSync(path.join(f.root, "tmp/readonly"), 0o500);
      const environment = utility(f, command, `
trap '' INT TERM
${command === "rm" ? "/usr/bin/rm -f tmp/readonly/file" : ""}
sleep 30 &
descendant=$!
printf '%s %s %s\\n' "$PPID" "$$" "$descendant" > barrier
wait "$descendant"`);
      const child = spawn("make", ["--no-print-directory", "clean"], { cwd: f.root, env: fixtureEnvironment(environment), stdio: ["ignore", "pipe", "pipe"] });
      let stdout = "", stderr = "", identities = [];
      child.stdout.on("data", (chunk) => { stdout += chunk; });
      child.stderr.on("data", (chunk) => { stderr += chunk; });
      const exited = once(child, "exit");
      t.after(() => {
        child.kill("SIGKILL");
        for (const pid of identities) { try { process.kill(pid, "SIGKILL"); } catch {} }
      });
      const barrier = path.join(f.root, "barrier");
      for (let attempt = 0; !existsSync(barrier) && attempt < 500; attempt++) await new Promise((resolve) => setTimeout(resolve, 10));
      assert.ok(existsSync(barrier), `${command} did not reach the cancellation barrier: ${stderr}`);
      identities = readFileSync(barrier, "utf8").trim().split(" ").map(Number);
      assert.equal(identities.length, 3);
      process.kill(identities[0], signal);
      const [code] = await exited;
      assert.notEqual(code, 0);
      assert.match(stderr, new RegExp(`cancelled_or_interrupted.*phase=${phase}`));
      assert.match(stderr, new RegExp(`Error ${signal === "SIGINT" ? 130 : 143}`));
      assert.doesNotMatch(stdout, /CLEANED/u);
      assert.equal(existsSync(path.join(f.root, "tmp")), true);
      assert.equal(existsSync(path.join(f.root, "tmp/readonly/file")), command !== "rm");
      outputBudget({ stdout, stderr });
      for (const pid of identities.slice(1)) {
        for (let attempt = 0; existsSync(`/proc/${pid}/fd/8`) && attempt < 100; attempt++) await new Promise((resolve) => setTimeout(resolve, 10));
        assert.equal(existsSync(`/proc/${pid}/fd/8`), false, `${pid} retained admission`);
        if (existsSync(`/proc/${pid}/status`)) assert.match(readFileSync(`/proc/${pid}/status`, "utf8"), /^State:\s+Z/mu);
      }
      identities = [];
      // Prove the settled children no longer hold admission without deleting.
      success(f.run(["clean", "CARTULARY_CLEANUP_DRY_RUN=1"]));
    }
  }
});

test("default Make identity is stable across child environments and honors caller identity", (t) => {
  const f = fixture(t);
  f.write("Makefile", `${readFileSync(path.join(f.root, "Makefile"), "utf8")}
ifeq ($(workspace_admitted),yes)
.PHONY: identity-probe
identity-probe:
	@printf '%s\\n' '$(CARTULARY_TEST_RUN_ID)' "$$CARTULARY_TEST_RUN_ID"
	@printf '%s\\n' '$(CARTULARY_TEST_RUN_ID)' "$$CARTULARY_TEST_RUN_ID"
endif
`);
  for (const [args, environment, expected] of [
    [[], {}, null], [[], { CARTULARY_TEST_RUN_ID: "inherited-identity" }, "inherited-identity"],
    [["CARTULARY_TEST_RUN_ID=explicit-identity"], { CARTULARY_TEST_RUN_ID: "inherited-identity" }, "explicit-identity"],
  ]) {
    const result = f.run(["identity-probe", ...args], environment); success(result);
    const identities = result.stdout.trim().split("\n");
    assert.equal(identities.length, 4, result.stdout);
    assert.equal(new Set(identities).size, 1, result.stdout);
    if (expected) assert.equal(identities[0], expected);
    else assert.match(identities[0], /^\d{8}T\d{6}Z-p\d+$/u);
  }
});

test("cleanup input parity and inert preview without runtimes", (t) => {
  const f = fixture(t), utilities = path.join(f.root, "utilities"); mkdirSync(utilities);
  for (const command of ["bash", "env", "make", "git", "flock", "stat", "find", "sha256sum", "readlink", "rm", "chmod", "dirname", "setsid"]) symlinkSync(`/usr/bin/${command}`, path.join(utilities, command));
  f.write("tmp/file"); f.write(".cache/cartulary/cache");
  const before = snapshot(f.root);
  for (const value of ["1", " 1 ", "\t1\n", "\u00a0\ufeff1\u3000", "\u2028 1\u2029"]) {
    const result = f.run(["distclean", `CARTULARY_CLEANUP_DRY_RUN=${value}`], { PATH: utilities }); success(result); assert.match(result.stdout, /DRY-RUN/); assert.deepEqual(snapshot(f.root), before);
  }
  for (const args of [["clean", "CARTULARY_CLEANUP_DRY_RUN=0"], ["clean", "CARTULARY_CLEANUP_DRY_RUN=true"], ["clean", "CARTULARY_CLEANUP_DRY_RUN=$(shell touch injected)"], ["clean", "CARTULARY_OUTPUT_MODE=machine"], ["clean", "UNDECLARED=1"], ["clean", "build-server"]]) { rejected(f.run(args, { PATH: utilities }), "usage_error|cannot be combined"); assert.deepEqual(snapshot(f.root), before); }
  success(f.run(["clean", "CARTULARY_CLEANUP_DRY_RUN="], { PATH: utilities, CARTULARY_CLEANUP_DRY_RUN: "1" })); assert.equal(existsSync(path.join(f.root, "tmp")), false);
  success(f.run(["distclean"], { PATH: utilities })); success(f.run(["distclean"], { PATH: utilities }));
});

test("directory ownership tiers, legacy adoption, symlinks, and custom-root precedence", (t) => {
  const f = fixture(t);
  const clean = ["tmp/scratch", "tmp/space and\nnewline", "build/bin/server", "apps/web/dist/index.html", "apps/web/dist-measurement/index.html", ".cartulary/test-results/run", ".cartulary/release-artifacts/report", "apps/web/.cartulary/test-results/manual", "node_modules/.vite/state", "apps/web/node_modules/.vite-temp/state", "packages/sample/node_modules/.vite/state", "packages/line\nbreak/node_modules/.vite/state", "internal/platform/httpapi/webassets/assets/generated/archive", "internal/platform/httpapi/webassets/dist/old"];
  const dist = [".cache/cartulary/node-runtime/bin/node", ".cache/cartulary/frontend-install/proof", ".cache/cartulary/typescript/web.tsbuildinfo", ".pnpm-store/content", "node_modules/dependency/index.js", "apps/web/node_modules/dependency/index.js", "packages/sample/node_modules/dependency/index.js", ...["go-build", "go-cache", "go-mod", "go-mod-cache", "cache", "tmp"].map((name) => `.cartulary/${name}/old`), ".fallow/cache"];
  const preserved = ["temp/keep", "build/unrelated/keep", ".cartulary/retained-work/keep", ".cartulary/runtime/state", ".cartulary/web-e2e-port-leases/lease", ".cache/other/keep", "custom-results/keep", "internal/platform/httpapi/webassets/assets/fallback/index.html"];
  for (const file of [...clean, ...dist, ...preserved]) f.write(file);
  symlinkSync(path.join(f.root, "temp"), path.join(f.root, "tmp/external"));
  symlinkSync("/missing-target", path.join(f.root, "tmp/dangling"));
  chmodSync(path.join(f.root, ".cartulary/go-build"), 0o500);
  success(f.run(["clean"], { CARTULARY_TEST_RESULTS_DIR: path.join(f.root, "tmp/custom"), SERVER_BIN: path.join(f.root, "temp/keep") }));
  for (const file of clean) assert.equal(existsSync(path.join(f.root, file)), false, file);
  for (const file of [...dist, ...preserved]) assert.equal(existsSync(path.join(f.root, file)), true, file);
  symlinkSync(path.join(f.root, "temp"), path.join(f.root, "tmp"));
  success(f.run(["distclean"])); assert.equal(existsSync(path.join(f.root, "tmp")), false);
  for (const file of dist) assert.equal(existsSync(path.join(f.root, file)), false, file);
  for (const file of preserved) assert.equal(existsSync(path.join(f.root, file)), true, file);
});

test("complete-plan validation rejects late unsafe state without content or permission changes", (t) => {
  for (const scenario of ["nested-git", "wrong-type", "unreadable", "parent-permission", "source", "ancestor", "stale", "enumeration"]) {
    const f = fixture(t); f.write("tmp/early-safe/file"); chmodSync(path.join(f.root, "tmp/early-safe"), 0o500);
    if (scenario === "nested-git") f.write("internal/platform/httpapi/webassets/dist/nested/.git/config");
    if (scenario === "wrong-type") f.write(".fallow");
    if (scenario === "unreadable") { f.write(".fallow/hidden/file"); chmodSync(path.join(f.root, ".fallow/hidden"), 0); }
    if (scenario === "parent-permission") {
      if (process.getuid() === 0) continue;
      f.write("build/bin/file"); chmodSync(path.join(f.root, "build"), 0o500);
    }
    if (scenario === "source") { f.write("apps/web/dist/source.js"); f.git("add", "apps/web/dist/source.js"); }
    if (scenario === "ancestor") { f.write("external/bin/keep"); symlinkSync(path.join(f.root, "external"), path.join(f.root, "build")); }
    if (scenario === "stale") f.write("tools/workspace_layout.json", "{}\n");
    let environment = {};
    if (scenario === "enumeration") { f.write("utilities/find", "#!/bin/bash\nexit 1\n"); chmodSync(path.join(f.root, "utilities/find"), 0o755); environment = { PATH: `${f.root}/utilities:${process.env.PATH}` }; }
    // Unreadable-content snapshots use known candidates, not an elevated traversal.
    const before = scenario === "unreadable" ? snapshot(path.join(f.root, "tmp")) : snapshot(f.root);
    rejected(f.run(["distclean", "CARTULARY_CLEANUP_DRY_RUN=1"], environment), "cleanup_error|configuration_error");
    assert.deepEqual(scenario === "unreadable" ? snapshot(path.join(f.root, "tmp")) : snapshot(f.root), before);
    rejected(f.run(["distclean"], environment), "cleanup_error|configuration_error");
    assert.equal(readFileSync(path.join(f.root, "tmp/early-safe/file"), "utf8"), "fixture");
    assert.equal(lstatSync(path.join(f.root, "tmp/early-safe")).mode & 0o777, 0o500);
  }
});

test("registered and locked worktrees reject cleanup", (t) => {
  const f = fixture(t); f.write("build/bin/safe");
  f.git("worktree", "add", "--detach", "tmp/checkout", "HEAD"); f.git("worktree", "lock", "tmp/checkout");
  for (const args of [["clean", "CARTULARY_CLEANUP_DRY_RUN=1"], ["clean"]]) rejected(f.run(args), "registered worktree");
  assert.equal(readFileSync(path.join(f.root, "build/bin/safe"), "utf8"), "fixture");
});

test("execution failure reports progress and permits idempotent retry", (t) => {
  const f = fixture(t); f.write("tmp/scratch"); f.write("build/bin/server");
  f.write("utilities/rm", '#!/bin/bash\nfor arg in "$@"; do if [[ "$arg" == build/bin ]]; then exit 1; fi; done\nexec /usr/bin/rm "$@"\n'); chmodSync(path.join(f.root, "utilities/rm"), 0o755);
  const result = f.run(["clean"], { PATH: `${f.root}/utilities:${process.env.PATH}` }); rejected(result, "cleanup_error.*completed=1.*remaining=");
  assert.equal(existsSync(path.join(f.root, "tmp")), false); assert.equal(existsSync(path.join(f.root, "build/bin/server")), true);
  success(f.run(["clean"]));
});

async function ready(child) {
  let output = "";
  await new Promise((resolve, reject) => { child.stdout.on("data", (data) => { output += data; if (output.includes("ready")) resolve(); }); child.once("error", reject); child.once("exit", (code) => reject(new Error(`admission fixture exited ${code}: ${output}`))); });
  return output;
}
test("shared admission excludes cleanup, recurses, and survives a parent exit", async (t) => {
  const f = fixture(t); f.write("tmp/scratch");
  // A child retains the same kernel lock when its launching shell exits.
  const child = spawn("bash", ["tools/harness/workspace/admission.sh", "run", "--", "bash", "tools/harness/workspace/admission.sh", "run", "--", "bash", "-c", "echo ready; read -r done"], { cwd: f.root, stdio: ["pipe", "pipe", "pipe"] });
  t.after(() => child.kill()); await ready(child);
  const before = snapshot(f.root);
  rejected(f.run(["clean", "CARTULARY_CLEANUP_DRY_RUN=1"]), "resource_conflict"); assert.deepEqual(snapshot(f.root), before);
  rejected(f.run(["clean"]), "resource_conflict");
  child.stdin.end("done\n"); await once(child, "exit"); success(f.run(["clean"]));
  // Environment assertions cannot stand in for an open, locked descriptor.
  f.write("tmp/again"); success(f.run(["clean"], { CARTULARY_WORKSPACE_ADMITTED: "1", CARTULARY_WORKSPACE_ADMISSION_FD: "9" }));
});

test("managed Node descendants retain admission after parent death", async (t) => {
  const f = fixture(t); f.write("tmp/scratch");
  const program = 'import {spawn} from "./tools/harness/workspace/child-process.mjs"; const child=spawn("sleep",["30"],{stdio:"ignore",detached:true}); child.unref(); console.log(`ready ${child.pid}`); setInterval(()=>{},1000);';
  const parent = spawn("bash", ["tools/harness/workspace/admission.sh", "run", "--", process.execPath, "--input-type=module", "-e", program], { cwd: f.root, stdio: ["ignore", "pipe", "pipe"] });
  const output = await ready(parent), descendant = Number(output.match(/ready (\d+)/u)[1]);
  t.after(() => { parent.kill("SIGKILL"); try { process.kill(descendant, "SIGKILL"); } catch {} });
  const exited = once(parent, "exit"); parent.kill("SIGKILL"); await exited;
  rejected(f.run(["clean"]), "resource_conflict");
  process.kill(descendant, "SIGKILL");
  for (let attempt = 0; existsSync(`/proc/${descendant}/fd/9`) && attempt < 100; attempt++) await new Promise((resolve) => setTimeout(resolve, 10));
  success(f.run(["clean"]));
});

test("Make prerequisite work and concurrent cleanup participate in admission", async (t) => {
  const f = fixture(t);
  const makefile = readFileSync(path.join(f.root, "Makefile"), "utf8");
  f.write("Makefile", `${makefile}\nifeq ($(workspace_admitted),yes)\n.PHONY: admission-fixture\nadmission-fixture: tmp/prepared\ntmp/prepared:\n\t@mkdir -p tmp; echo ready; while ! test -e release; do sleep 0.05; done; touch tmp/prepared\nendif\n`);
  const builder = spawn("make", ["--no-print-directory", "admission-fixture"], { cwd: f.root, stdio: ["ignore", "pipe", "pipe"] }); t.after(() => builder.kill()); await ready(builder);
  rejected(f.run(["clean"]), "resource_conflict");
  const exited = once(builder, "exit"); f.write("release"); await exited;
  const holder = spawn("bash", ["-c", 'exec 8<.git/cartulary-workspace.lock; flock -n -x 8; echo ready; read -r done'], { cwd: f.root, stdio: ["pipe", "pipe", "pipe"] }); t.after(() => holder.kill()); await ready(holder);
  rejected(f.run(["clean"]), "resource_conflict");
  const denied = spawnSync("bash", ["tools/harness/workspace/admission.sh", "run", "--", "true"], { cwd: f.root, encoding: "utf8" }); rejected(denied, "resource_conflict");
  const released = once(holder, "exit"); holder.stdin.end("done\n"); await released; success(f.run(["clean"]));
});

test("compiler configurations keep build metadata inside distinct owned cache paths", () => {
  const configs = JSON.parse(readFileSync(path.join(repo, "tsconfig.json"))).references.map((entry) => entry.path.endsWith(".json") ? entry.path : `${entry.path}/tsconfig.json`);
  const paths = new Set();
  for (const config of configs) {
    const value = JSON.parse(readFileSync(path.join(repo, config))).compilerOptions.tsBuildInfoFile;
    const destination = path.resolve(repo, path.dirname(config), value);
    assert.ok(destination.startsWith(path.join(repo, ".cache/cartulary/typescript/")), config);
    assert.ok(!paths.has(destination), config); paths.add(destination);
  }
});

test("mount boundaries reject before mutation", (t) => {
  const probe = spawnSync("unshare", ["--user", "--map-root-user", "--mount", "true"], { encoding: "utf8" });
  if (probe.status !== 0) { t.skip("this host does not allow an unprivileged mount namespace"); return; }
  const f = fixture(t); f.write("tmp/safe"); f.write("outside/keep"); mkdirSync(path.join(f.root, "tmp/mount"));
  const result = spawnSync("unshare", ["--user", "--map-root-user", "--mount", "bash", "-c", 'mount --bind "$1/outside" "$1/tmp/mount" && cd "$1" && make --no-print-directory clean', "--", f.root], { encoding: "utf8", env: fixtureEnvironment() });
  rejected(result, "mount boundary"); assert.equal(readFileSync(path.join(f.root, "tmp/safe"), "utf8"), "fixture"); assert.equal(readFileSync(path.join(f.root, "outside/keep"), "utf8"), "fixture");
});

test("workspace projections reject ambiguity and stay independent of Markdown", () => {
  const owner = JSON.parse(readFileSync(path.join(repo, "tools/workspace_layout.json"))), schema = JSON.parse(readFileSync(path.join(repo, "tools/schemas/cartulary.workspace_layout.v1.schema.json")));
  for (const badPath of ["../outside", "tools", "apps/web", "tmp/../../source", "tmp/$(touch-pwned)", "packages/*/src", "tmp/'quote"]) {
    const changed = structuredClone(owner); changed.directories[0].path = badPath; assert.throws(() => validateLayout(changed, schema));
  }
  const changed = structuredClone(owner); changed.directories.push({ ...changed.directories[0], id: "overlap", path: "tmp/nested" }); assert.throws(() => validateLayout(changed, schema), /overlap/);
  for (const [file, content] of renderLayout(repo)) { assert.equal(readFileSync(path.join(repo, file), "utf8"), content); assert.doesNotMatch(content, /(?:docs\/|README|\.md\b)/u); }
});
