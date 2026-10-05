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
  const run = (args, env = {}) => spawnSync("make", ["--no-print-directory", ...args], { cwd: root, encoding: "utf8", timeout: 15000, env: { PATH: process.env.PATH, HOME: process.env.HOME, LC_ALL: "C", GIT_OPTIONAL_LOCKS: "0", ...env } });
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

test("cleanup input parity and inert preview without runtimes", (t) => {
  const f = fixture(t), utilities = path.join(f.root, "utilities"); mkdirSync(utilities);
  for (const command of ["bash", "env", "make", "git", "flock", "stat", "find", "sort", "sha256sum", "readlink", "rm", "chmod", "dirname"]) symlinkSync(`/usr/bin/${command}`, path.join(utilities, command));
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
  const result = spawnSync("unshare", ["--user", "--map-root-user", "--mount", "bash", "-c", 'mount --bind "$1/outside" "$1/tmp/mount" && cd "$1" && make --no-print-directory clean', "--", f.root], { encoding: "utf8" });
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
