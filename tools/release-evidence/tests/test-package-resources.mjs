import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "node:child_process";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { createRuntime } from "../../harness/runtime/session-client.mjs";
import { cleanupPackageResource, packageResourceIdentity, reservePackageResources } from "../package-resources.mjs";

const token = "11111111-2222-4333-8444-555555555555";
const specimen = role => ({ kind: "package_resource", target: { daemon_id: "fixture-engine", token, role },
 state: "pending", run_id: "package-owner", lease_id: "00000000-0000-4000-8000-000000000001", depends_on: [] });

test("package cleanup admits only exact daemon and allocation ownership for every role", () => {
 for (const role of ["application", "postgres", "seaweedfs", "build_tag"]) {
  const resource = specimen(role), image = role === "build_tag";
  const name = image ? `cartulary/package-build:${token}` : `cartulary-package-${token}-${role}`;
  const labels = { "cartulary.package.allocation": token, "cartulary.package.run": "package-owner",
   "cartulary.package.lease": resource.lease_id, "cartulary.package.role": role };
  assert.deepEqual(packageResourceIdentity(resource), { name, labels });
  for (const scenario of ["success", "absent", "daemon", "lookup_error", "ambiguous", "short_id", "inspect_error", "wrong_id",
   "wrong_name", "wrong_token", "wrong_run", "wrong_lease", "wrong_role", "remove_error", "leftover", "post_lookup_error"]) {
   if (image && ["inspect_error", "wrong_id", "wrong_name", "wrong_token", "wrong_run", "wrong_lease", "wrong_role"].includes(scenario)) continue;
   const calls = []; let present = scenario !== "absent", removed = false;
   const id = image ? `sha256:${"a".repeat(64)}` : "a".repeat(64);
   const list = image ? ["image", "ls", "--no-trunc", "--quiet", "--filter", `reference=${name}`]
    : ["ps", "--all", "--no-trunc", "--filter", `name=^/${name}$`, "--format", "{{.ID}}"];
   const run = args => {
    calls.push(args);
    if (args[0] === "info") { assert.deepEqual(args, ["info", "--format", "{{.ID}}"]); return scenario === "daemon" ? "foreign-engine" : "fixture-engine"; }
    if (JSON.stringify(args) === JSON.stringify(list)) {
     if (scenario === "lookup_error" || removed && scenario === "post_lookup_error") throw new Error("private daemon diagnostic");
     if (scenario === "ambiguous") return `${id}\n${id}`;
     if (scenario === "short_id") return "short";
     return present ? id : "";
    }
    if (args[0] === "inspect") {
     assert.deepEqual(args, ["inspect", "--format", "{{json .}}", id]);
     if (scenario === "inspect_error") return "private invalid JSON";
     const value = { Id: id, Name: `/${name}`, Config: { Labels: { ...labels } } };
     if (scenario === "wrong_id") value.Id = "b".repeat(64);
     if (scenario === "wrong_name") value.Name = "/borrowed";
     for (const [mutation, key] of [["wrong_token", "allocation"], ["wrong_run", "run"], ["wrong_lease", "lease"], ["wrong_role", "role"]]) {
      if (scenario === mutation) value.Config.Labels[`cartulary.package.${key}`] = "foreign";
     }
     return JSON.stringify(value);
    }
    assert.deepEqual(args, image ? ["image", "rm", name] : ["rm", "--force", id]);
    removed = true;
    if (scenario === "remove_error") throw new Error("private remove diagnostic");
    if (scenario !== "leftover") present = false;
    return "";
   };
   if (["success", "absent"].includes(scenario)) cleanupPackageResource(resource, run);
   else assert.throws(() => cleanupPackageResource(resource, run), error => error.failure_class === "harness" && error.failure_reason === "cleanup_error" && error.message === "package resource cleanup failed");
   assert.equal(removed, ["success", "remove_error", "leftover", "post_lookup_error"].includes(scenario), `${role}/${scenario}`);
   assert.ok(calls.every(args => !args.some(value => value.startsWith("cartulary/local:"))));
  }
 }
 for (const mutate of [r => { r.kind = "foreign"; }, r => { r.target.role = "borrowed"; }, r => { r.target.token = "--all"; },
  r => { r.target.daemon_id = ""; }, r => { r.target.extra = true; }, r => { r.extra = true; }]) {
  const resource = specimen("application"); mutate(resource); let calls = 0;
  assert.throws(() => cleanupPackageResource(resource, () => { calls++; })); assert.equal(calls, 0);
 }
});

test("physical package cleanup uses the actual private runtime binding and rejects a foreign envelope", t => {
 const scratch = createScratch(t, "cartulary-package-physical-");
 const repoRoot = path.resolve(import.meta.dirname, "../../.."), runRoot = path.join(scratch, "run"), bin = path.join(scratch, "bin"), log = path.join(scratch, "calls");
 fs.mkdirSync(runRoot, { mode: 0o700 }); fs.mkdirSync(bin); fs.writeFileSync(log, "");
 fs.writeFileSync(path.join(bin, "docker"), '#!/bin/sh\nprintf "%s\\n" "$*" >> "$PACKAGE_FIXTURE_LOG"\nif [ "$1" = info ]; then printf "fixture-engine\\n"; fi\n', { mode: 0o700 });
 let runtime;
 const resources = [];
 try {
  runtime = createRuntime({ repoRoot, runRoot, runID: "package-owner", scratchRoot: path.join(scratch, "private") });
  const env = { ...process.env, ...runtime.producerEnvironment(), PATH: `${bin}:${process.env.PATH}`,
   CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT: runRoot, PACKAGE_FIXTURE_LOG: log };
  const invoke = resource => spawnSync(process.execPath, [path.join(repoRoot, "tools/harness/runtime/resource-stop-private.mjs")], { cwd: repoRoot, env, input: JSON.stringify(resource), encoding: "utf8", timeout: 10000 });
  for (const role of ["application", "postgres", "seaweedfs", "build_tag"]) {
   const resource = { ...specimen(role), lease_id: runtime.leaseID };
   runtime.record(resource); resources.push(resource);
   const admitted = runtime.inspectResource(resource);
   assert.deepEqual(admitted, resource);
   const result = invoke(resource);
   assert.ifError(result.error); assert.equal(result.status, 0, result.stderr); assert.equal(result.stdout, "");
   runtime.record({ ...resource, state: "released" });
  }
  const renderer = { kind: "browser_renderer", target: { daemon_id: "fixture-engine", image_id: `sha256:${"a".repeat(64)}`,
   name: "cartulary-visual-123-aaaaaaaaaaaa", token: "a".repeat(32) }, state: "pending", run_id: runtime.runID, lease_id: runtime.leaseID, depends_on: [] };
  runtime.record(renderer); resources.push(renderer);
  const rendererResult = invoke(renderer); assert.ifError(rendererResult.error); assert.equal(rendererResult.status, 0, rendererResult.stderr);
  assert.equal(rendererResult.stdout, ""); runtime.record({ ...renderer, state: "released" });
  const before = fs.readFileSync(log, "utf8");
  for (const resource of [resources[0], renderer, { ...specimen("application"), lease_id: runtime.leaseID, target: { ...specimen("application").target, token: "11111111-2222-4333-8444-666666666666" } }]) {
   const result = invoke(resource); assert.ifError(result.error); assert.equal(result.status, 12); assert.equal(result.stdout + result.stderr, "");
  }
  assert.equal(fs.readFileSync(log, "utf8"), before, "released and unregistered identities cannot reach Docker");
  for (const mutate of [r => { r.run_id = "foreign-run"; }, r => { r.lease_id = "00000000-0000-4000-8000-000000000000"; }, r => { r.target.role = "unknown"; }]) {
   const resource = { ...specimen("application"), lease_id: runtime.leaseID }; mutate(resource);
   const result = invoke(resource); assert.ifError(result.error); assert.equal(result.status, 12); assert.equal(result.stdout + result.stderr, "");
  }
  assert.equal(fs.readFileSync(log, "utf8"), before, "rejected binding cannot reach Docker");
  assert.equal(before.split("\n").filter(line => line.startsWith("info ")).length, 5);
 } finally {
  const errors = [];
  for (const resource of resources) { try { runtime.record({ ...resource, state: "released" }); } catch (error) { errors.push(error); } }
  try { runtime?.close(); } catch (error) { errors.push(error); }
  if (errors.length) throw new AggregateError(errors, "package resource fixture cleanup failed");
 }
});

test("allocation reserves only fresh names and independently settles journal or cleanup failures", () => {
 for (const scenario of ["success", "existing_container", "existing_tag", "lookup_error", "invalid_daemon", "pending_reply", "release_reply", "remove_error"]) {
  const observations = [], records = [], present = new Set(); let pending = 0, token;
  const runtime = { runID: "package-owner", leaseID: specimen("application").lease_id, record: resource => {
   records.push(resource);
   assert.equal(observations.filter(args => args[0] === "ps" || args[1] === "ls").length >= 4, true, "absence for every name precedes every journal write");
   if (resource.state === "pending") { pending++; if (scenario === "pending_reply" && pending === 2) throw new Error("pending reply lost"); }
   if (scenario === "release_reply" && resource.state === "released" && resource.target.role === "application") throw new Error("release reply lost");
  } };
  const run = args => {
   observations.push(args);
   if (args[0] === "info") return scenario === "invalid_daemon" ? "invalid daemon" : "fixture-engine";
   if (args[0] === "ps" || args[1] === "ls") {
    const role = args[0] === "ps" ? args[args.indexOf("--filter") + 1].match(/-(application|postgres|seaweedfs)\$$/u)[1] : "build_tag";
    if (scenario === "lookup_error") throw new Error("lookup failed");
    if (!pending && (scenario === "existing_container" && role === "postgres" || scenario === "existing_tag" && role === "build_tag")) return "foreign";
    return present.has(role) ? role === "build_tag" ? `sha256:${"a".repeat(64)}` : "a".repeat(64) : "";
   }
   if (args[0] === "inspect") return JSON.stringify({ Id: "a".repeat(64), Name: `/cartulary-package-${token}-application`, Config: { Labels: {
    "cartulary.package.allocation": token, "cartulary.package.run": runtime.runID, "cartulary.package.lease": runtime.leaseID, "cartulary.package.role": "application",
   } } });
   if (scenario === "remove_error") throw new Error("removal failed");
   present.delete(args[0] === "rm" ? "application" : "build_tag"); return "";
  };
  if (["existing_container", "existing_tag", "lookup_error", "invalid_daemon"].includes(scenario)) {
   assert.throws(() => reservePackageResources(runtime, run)); assert.equal(records.length, 0);
   assert.equal(observations.some(args => args[0] === "rm" || args[1] === "rm"), false); continue;
  }
  if (scenario === "pending_reply") {
   assert.throws(() => reservePackageResources(runtime, run), /pending reply lost/u);
   assert.deepEqual(records.filter(value => value.state === "released").map(value => value.target.role), ["application", "postgres"]); continue;
  }
  const allocation = reservePackageResources(runtime, run);
  assert.deepEqual(records.map(value => [value.target.role, value.state]), ["application", "postgres", "seaweedfs", "build_tag"].map(role => [role, "pending"]));
  token = records[0].target.token; assert.match(token, /^[a-f0-9-]{36}$/u);
  assert.equal(new Set(records.map(value => value.target.token)).size, 1);
  present.add("application"); present.add("build_tag");
  const failures = allocation.releaseAll();
  assert.deepEqual(failures.map(error => error.cleanup_step), scenario === "success" ? [] : scenario === "remove_error" ? ["application", "build_tag"] : ["application"]);
  assert.ok(failures.every(error => error.failure_reason === "cleanup_error"));
  assert.deepEqual(records.filter(value => value.state === "released").map(value => value.target.role), scenario === "remove_error" ? ["postgres", "seaweedfs"] : ["application", "postgres", "seaweedfs", "build_tag"]);
  if (scenario === "success") { const before = observations.length; allocation.releaseAll(); allocation.release("application"); assert.equal(observations.length, before); }
  assert.throws(() => allocation.identity("borrowed"));
 }
});
