import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import childProcess from "node:child_process";
import { syncBuiltinESMExports } from "node:module";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { rendererProbe } from "../package-platform.mjs";

test("renderer platform probe preserves confinement observations and exact owned cleanup", t => {
 const root = createScratch(t, "cartulary-package-platform-");
 const work = path.join(root, "work"), artifacts = path.join(root, "artifacts");
 fs.mkdirSync(path.join(work, "release/assets"), { recursive: true }); fs.mkdirSync(artifacts);
 const profile = { defaultAction: "SCMP_ACT_ERRNO", syscalls: [{ names: ["read"], action: "SCMP_ACT_ALLOW" }] };
 fs.writeFileSync(path.join(work, "release/assets/application-seccomp.v1.json"), JSON.stringify(profile));
 const baseHost = { Privileged: false, CapDrop: ["ALL"], CapAdd: [], SecurityOpt: ["no-new-privileges", `seccomp=${JSON.stringify(profile)}`] };
 const lsm = { apparmor: "fixture-confined", process_label: "system_u:system_r:container_t:s0:c1,c2" };
 const id = "a".repeat(64), image = `sha256:${"b".repeat(64)}`, project = "cartulary-qualification-11111111-2222-4333-8444-555555555555-source";
 const original = childProcess.spawnSync;
 try {
  for (const scenario of ["success", "privileged", "cap_add", "unconfined", "cap_drop", "missing_seccomp", "wrong_seccomp",
   "create_failure", "invalid_id", "probe_failure", "cleanup", "both", "lsm_apparmor", "lsm_domain", "execution_error",
   "execution_exit", "container_exit", "oom", "schema", "status", "outer_policy", "operations", "attempts", "loopback", "inner_cleanup"]) {
   const host = structuredClone(baseHost), calls = [], primary = new Error("private primary probe diagnostic");
   if (scenario === "privileged") host.Privileged = true;
   if (scenario === "cap_add") host.CapAdd = ["SYS_ADMIN"];
   if (scenario === "unconfined") host.SecurityOpt.push("apparmor=unconfined");
   if (scenario === "cap_drop") host.CapDrop = [];
   if (scenario === "missing_seccomp") host.SecurityOpt.pop();
   if (scenario === "wrong_seccomp") host.SecurityOpt[1] = 'seccomp={"different":true}';
   childProcess.spawnSync = (binary, args) => {
    assert.equal(binary, "docker"); calls.push(args);
    if (args[0] === "create") {
     if (scenario === "create_failure") throw primary;
     assert.deepEqual(args, ["create", "--network", "none", "--read-only", "--label", `com.docker.compose.project=${project}`,
      "--tmpfs", "/tmp:rw,nosuid,nodev,size=1g,mode=1777", "--mount", `type=bind,source=${work}/render-probe,target=/render-probe,readonly`,
      "--mount", `type=bind,source=${work}/render-fixtures,target=/probe-fixtures,readonly`, "--entrypoint", "/render-probe",
      "--cap-drop", "ALL", "--security-opt", "no-new-privileges", "--security-opt", `seccomp=${work}/release/assets/application-seccomp.v1.json`, image]);
     return { status: 0, stdout: scenario === "invalid_id" ? "--all" : id };
    }
    if (args[0] === "rm") {
     assert.deepEqual(args, ["rm", "--force", id]);
     return { status: ["cleanup", "both"].includes(scenario) ? 1 : 0, stdout: "", stderr: "private cleanup diagnostic" };
    }
    if (args[0] === "inspect" && args[2] !== "{{json .State}}") {
     assert.equal(args.at(-1), id);
     if (["probe_failure", "both"].includes(scenario)) throw primary;
     return { status: 0, stdout: JSON.stringify({ apparmor: scenario === "lsm_apparmor" ? "unconfined" : lsm.apparmor,
      process_label: scenario === "lsm_domain" ? "other:domain:type:s0" : "system_u:system_r:container_t:s0:c9,c10" }) };
    }
    if (args[0] === "inspect") {
     assert.deepEqual(args, ["inspect", "--format", "{{json .State}}", id]);
     return { status: 0, stdout: JSON.stringify({ ExitCode: scenario === "container_exit" ? 7 : 0, OOMKilled: scenario === "oom" }) };
    }
    assert.deepEqual(args, ["start", "--attach", id]);
    const qualification = { schema_id: "cartulary.packaged_renderer_qualification.v1", status: "pass", outer_policy_verified: true,
     operations: ["mermaid_svg", "slidev_pdf"].map(operation => ({ operation, fresh_attempts: 2, native_loopback_observed: true, cleanup_verified: true })) };
    if (scenario === "schema") qualification.schema_id = "incorrect";
    if (scenario === "status") qualification.status = "fail";
    if (scenario === "outer_policy") qualification.outer_policy_verified = false;
    if (scenario === "operations") qualification.operations.reverse();
    if (scenario === "attempts") qualification.operations[0].fresh_attempts = 1;
    if (scenario === "loopback") qualification.operations[1].native_loopback_observed = false;
    if (scenario === "inner_cleanup") qualification.operations[0].cleanup_verified = false;
    return { status: scenario === "execution_exit" ? 7 : 0, error: scenario === "execution_error" ? primary : undefined, stdout: JSON.stringify(qualification) };
   };
   syncBuiltinESMExports();
   let caught;
   try { rendererProbe(work, artifacts, project, image, host, lsm); } catch (error) { caught = error; }
   assert.equal(caught === undefined, scenario === "success", scenario);
   const beforeCreate = ["privileged", "cap_add", "unconfined", "cap_drop", "missing_seccomp", "wrong_seccomp"].includes(scenario);
   assert.equal(calls.filter(args => args[0] === "create").length, beforeCreate ? 0 : 1, scenario);
   assert.equal(calls.filter(args => args[0] === "rm").length, beforeCreate || ["create_failure", "invalid_id"].includes(scenario) ? 0 : 1, scenario);
   if (["create_failure", "probe_failure", "both"].includes(scenario)) assert.equal(caught, primary);
   if (scenario === "both") {
    assert.equal(caught.cleanupFailures.length, 1); assert.equal(caught.cleanupFailures[0].failure_reason, "cleanup_error");
    assert.equal(caught.message, "private primary probe diagnostic");
   }
   if (scenario === "cleanup") { assert.equal(caught.failure_class, "harness"); assert.equal(caught.failure_reason, "cleanup_error"); }
   if (scenario === "success") {
    const evidence = JSON.parse(fs.readFileSync(path.join(artifacts, "renderer-container.json")));
    assert.equal(evidence.image, image); assert.equal(evidence.container_id, id);
    assert.equal(evidence.exit_code, 0); assert.equal(evidence.oom_killed, false);
    assert.deepEqual(evidence.cap_drop, ["ALL"]);
    assert.match(evidence.security_options[1], /^seccomp=sha256:[a-f0-9]{64}$/u);
    assert.equal(evidence.qualification.status, "pass");
   }
  }
 } finally { childProcess.spawnSync = original; syncBuiltinESMExports(); }
});

test("package platform CLI keeps private errors behind fixed typed diagnostics", t => {
 const scratch = createScratch(t, "cartulary-private-platform-cli-");
 const bin = path.join(scratch, "bin"); fs.mkdirSync(bin);
 // Successful fake copy creates no destination: the real CLI then encounters a
 // filesystem error containing this private path, before any genuine Docker work.
 fs.writeFileSync(path.join(bin, "docker"), "#!/bin/sh\nexit 0\n", { mode: 0o700 });
 const env = { ...process.env, PATH: `${bin}:${process.env.PATH}` };
 delete env.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
 for (const [operation, project, origin, status, diagnostic] of [
  ["unknown-private-operation", "cartulary-qualification-11111111-2222-4333-8444-555555555555-source", "", 2, "config/usage_error"],
  ["layout", "private-invalid-project", "a".repeat(64), 2, "config/usage_error"],
  ["layout", "cartulary-qualification-11111111-2222-4333-8444-555555555555-source", "private-invalid-id", 2, "config/usage_error"],
  ["layout", "cartulary-qualification-11111111-2222-4333-8444-555555555555-source", "a".repeat(64), 11, "artifact/artifact_error"],
 ]) {
  const result = childProcess.spawnSync(process.execPath, [path.resolve(import.meta.dirname, "../package-platform.mjs"), operation, scratch, scratch, project, origin], { env, encoding: "utf8", timeout: 10000 });
  assert.ifError(result.error); assert.equal(result.status, status);
  assert.equal(result.stdout, ""); assert.equal(result.stderr, `Package platform failed: ${diagnostic}\n`);
 }
});
