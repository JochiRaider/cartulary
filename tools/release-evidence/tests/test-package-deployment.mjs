import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { cleanupPackageDeployment, packageDeploymentIdentity } from "../package-deployment.mjs";

const token = "11111111-2222-4333-8444-555555555555", lease = "00000000-0000-4000-8000-000000000001";
const proof = target => ({ kind: "package_deployment", target: { daemon_id: "fixture-engine", token, ...target }, state: "pending", run_id: "fixture", lease_id: lease, depends_on: [] });
const idA = "a".repeat(64), idB = "b".repeat(64), helperImage = `sha256:${"c".repeat(64)}`;

test("deployment families remove only individually inspected project members and prove absence", () => {
 for (const scope of ["source", "destination"]) for (const family of ["container", "volume", "network"]) {
  const resource = proof({ scope, family }), project = `cartulary-qualification-${token}-${scope}`;
  assert.equal(packageDeploymentIdentity(resource).project, project);
  const ids = family === "volume" ? [`${project}_one`, `${project}_two`] : [idA, idB];
  const query = [family, "ls", ...(family === "container" ? ["--all", "--no-trunc"] : family === "network" ? ["--no-trunc"] : []), "--filter", `label=com.docker.compose.project=${project}`, "--format", family === "volume" ? "{{.Name}}" : "{{.ID}}"];
  for (const scenario of ["success", "absent", "daemon", "lookup", "duplicate", "unsafe_id", "oversize", "invalid_json", "wrong_id", "wrong_project", "remove_failure", "remaining", "post_lookup"]) {
   const calls = [], removed = [], present = new Set(scenario === "absent" ? [] : ids); let queries = 0;
   const run = args => {
    calls.push(args);
    if (args[0] === "info") { assert.deepEqual(args, ["info", "--format", "{{.ID}}"]); return scenario === "daemon" ? "other-engine" : "fixture-engine"; }
    assert.equal(args[0], family);
    if (args[1] === "ls") {
     assert.deepEqual(args, query); queries++;
     if (scenario === "lookup" || scenario === "post_lookup" && queries > 1) throw new Error("private provider output");
     if (scenario === "duplicate") return `${ids[0]}\n${ids[0]}`;
     if (scenario === "unsafe_id") return "--all";
     if (scenario === "oversize") return Array.from({ length: 129 }, (_, i) => family === "volume" ? `volume-${i}` : i.toString(16).padStart(64, "0")).join("\n");
     return [...present].join("\n");
    }
    if (args[1] === "inspect") {
     const id = args[2]; assert.ok(ids.includes(id));
     if (id === ids[0] && scenario === "invalid_json") return "{";
     const labels = { "com.docker.compose.project": id === ids[0] && scenario === "wrong_project" ? "borrowed" : project };
     const identity = id === ids[0] && scenario === "wrong_id" ? "foreign" : id;
     return JSON.stringify([{ ...(family === "volume" ? { Name: identity } : { Id: identity }), ...(family === "container" ? { Config: { Labels: labels } } : { Labels: labels }) }]);
    }
    const id = args.at(-1); assert.deepEqual(args, [family, "rm", ...(family === "container" ? ["--force"] : []), id]);
    removed.push(id);
    if (id === ids[0] && scenario === "remove_failure") throw new Error("private provider output");
    if (!(id === ids[0] && scenario === "remaining")) present.delete(id);
    return "";
   };
   if (["success", "absent"].includes(scenario)) cleanupPackageDeployment(resource, run);
   else assert.throws(() => cleanupPackageDeployment(resource, run), error => error.failure_reason === "cleanup_error" && error.message === "package deployment cleanup failed");
   if (["daemon", "lookup", "duplicate", "unsafe_id", "oversize", "absent"].includes(scenario)) assert.deepEqual(removed, []);
   else if (["invalid_json", "wrong_id", "wrong_project"].includes(scenario)) assert.deepEqual(removed, [ids[1]], "foreign member is untouched while an independent owned member is attempted");
   else assert.deepEqual(removed, ids);
   assert.ok(calls.every(args => !args.includes("borrowed")));
  }
 }
 for (const mutate of [r => { r.kind = "unknown"; }, r => { r.target.scope = "operator"; }, r => { r.target.token = "--all"; }, r => { r.target.family = "image"; }, r => { r.target.extra = true; }]) {
  const selected = proof({ scope: "source", family: "container" }); mutate(selected); let calls = 0;
  assert.throws(() => cleanupPackageDeployment(selected, () => { calls++; })); assert.equal(calls, 0);
 }
});

test("workspace repair owns its helper before effects and preserves unrelated paths across failures", t => {
 const root = createScratch(t, "cartulary-deployment-workspace-");
 for (const scenario of ["success", "absent", "prior_helper", "absent_with_helper", "wrong_helper_name", "wrong_helper_label", "wrong_helper_image", "helper_lookup", "lost_create", "invalid_create", "repair_failure", "remove_failure", "helper_remaining", "linked_workspace", "nonprivate_workspace", "workspace_replaced"]) {
  const parent = path.join(root, scenario); fs.mkdirSync(parent, { mode: 0o700 });
  const workspace = path.join(parent, 'owned,"workspace'), borrowed = path.join(parent, "borrowed"); fs.mkdirSync(borrowed, { mode: 0o700 }); fs.writeFileSync(path.join(borrowed, "sentinel"), "borrowed");
  if (scenario === "linked_workspace") fs.symlinkSync(borrowed, workspace);
  else if (!["absent", "absent_with_helper"].includes(scenario)) { fs.mkdirSync(workspace, { mode: scenario === "nonprivate_workspace" ? 0o755 : 0o700 }); fs.chmodSync(workspace, scenario === "nonprivate_workspace" ? 0o755 : 0o700); fs.writeFileSync(path.join(workspace, "owned"), "owned"); }
  const resource = proof({ scope: "shared", family: "workspace", workspace, helper_image: helperImage });
  const helper = `cartulary-qualification-${token}-cleanup`, labels = { "cartulary.package.allocation": token, "cartulary.package.run": "fixture", "cartulary.package.lease": lease, "cartulary.package.role": "qualification_cleanup" };
  let present = ["prior_helper", "absent_with_helper", "wrong_helper_name", "wrong_helper_label", "wrong_helper_image"].includes(scenario), created = false, started = false;
  const removals = [];
  const run = (args, options) => {
   if (args[0] === "info") return "fixture-engine";
   if (args[0] === "container" && args[1] === "ls") {
    assert.deepEqual(args, ["container", "ls", "--all", "--no-trunc", "--filter", `name=^/${helper}$`, "--format", "{{.ID}}"]);
    if (scenario === "helper_lookup") throw new Error("private lookup failure");
    return present ? idA : "";
   }
   if (args[0] === "container" && args[1] === "inspect") {
    assert.equal(args[2], idA);
    return JSON.stringify([{ Id: idA, Name: scenario === "wrong_helper_name" ? "/borrowed" : `/${helper}`, Image: scenario === "wrong_helper_image" ? `sha256:${idB}` : helperImage,
     Config: { Labels: { ...labels, ...(scenario === "wrong_helper_label" ? { "cartulary.package.run": "foreign" } : {}) } } }]);
   }
   if (args[0] === "container" && args[1] === "rm") {
    assert.deepEqual(args, ["container", "rm", "--force", idA]); removals.push(idA);
    if (created && scenario === "remove_failure") throw new Error("private remove failure");
    if (!(created && scenario === "helper_remaining")) present = false;
    return "";
   }
   if (args[0] === "create") {
    assert.equal(present, false, "old helper must settle before another acquisition");
    created = true; present = true;
    assert.equal(args[args.indexOf("--name") + 1], helper);
    const observed = Object.fromEntries(args.flatMap((value, index) => value === "--label" ? [args[index + 1].split("=")] : [])); assert.deepEqual(observed, labels);
    assert.equal(args[args.indexOf("--mount") + 1], `"type=bind","source=${parent}/owned,""workspace","target=/cleanup"`);
    assert.equal(args[args.indexOf("--network") + 1], "none"); assert.ok(args.includes("--read-only"));
    assert.deepEqual(args.slice(-5), [helperImage, "-c", 'chown -Rh "$1" /cleanup', "sh", `${process.getuid()}:${process.getgid()}`]);
    if (scenario === "lost_create") throw new Error("lost create reply");
    return scenario === "invalid_create" ? "--all" : idA;
   }
   assert.deepEqual(args, ["start", "--attach", idA]); assert.deepEqual(options, { timeout: 110000 }); started = true;
   if (scenario === "repair_failure") throw new Error("repair failed");
   if (scenario === "workspace_replaced") { fs.renameSync(workspace, path.join(parent, "original")); fs.mkdirSync(workspace, { mode: 0o700 }); fs.writeFileSync(path.join(workspace, "replacement"), "preserve"); }
   return "";
  };
  const succeeds = ["success", "absent", "prior_helper", "absent_with_helper"].includes(scenario);
  if (succeeds) cleanupPackageDeployment(resource, run);
  else assert.throws(() => cleanupPackageDeployment(resource, run), error => error.failure_reason === "cleanup_error" && error.message === "package deployment cleanup failed");
  assert.equal(fs.readFileSync(path.join(borrowed, "sentinel"), "utf8"), "borrowed");
  if (succeeds) { assert.equal(fs.existsSync(workspace), false); assert.equal(present, false); }
  else assert.equal(fs.existsSync(workspace), true);
  if (["lost_create", "invalid_create", "repair_failure"].includes(scenario)) { assert.equal(present, false); assert.deepEqual(removals, [idA]); }
  if (scenario.startsWith("wrong_helper") || scenario === "helper_lookup") { assert.equal(created, false); assert.deepEqual(removals, []); }
  if (["linked_workspace", "nonprivate_workspace"].includes(scenario)) assert.equal(created, false);
  if (["absent", "absent_with_helper"].includes(scenario)) assert.equal(started, false);
  if (scenario === "workspace_replaced") { assert.equal(fs.readFileSync(path.join(workspace, "replacement"), "utf8"), "preserve"); assert.equal(fs.readFileSync(path.join(parent, "original/owned"), "utf8"), "owned"); }
 }
});
