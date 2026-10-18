import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, rmSync, existsSync, symlinkSync } from "node:fs";
import path from "node:path";
import os from "node:os";
import { resolvePlaywrightPackages } from "../../readiness/playwright-packages.mjs";
import { loadVisualRendererProfile, verifyRendererFonts, validateRendererAttestation, sha256 } from "../visual-renderer-profile.mjs";
import { rendererRelease, verifyRendererImage, startVisualRendererLease, rendererPublishedPorts, rendererLifecycleFailureReason } from "../visual-renderer-lease.mjs";
import { removePlaywrightWorkingTraces } from "../playwright-output-cleanup.mjs";
import { removeOwnedRenderer, validateRendererProof } from "../../runtime/owned-renderer.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const profile = loadVisualRendererProfile(root);
const installed = resolvePlaywrightPackages(root);
assert.equal(installed.version, profile.playwright_version);
assert.equal(installed.chromium.revision, profile.chromium_revision);
verifyRendererFonts(root, profile);
const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-renderer-identity-"));
const write = (file, value) => { const full = path.join(scratch, file); mkdirSync(path.dirname(full), { recursive: true }); writeFileSync(full, JSON.stringify(value)); };
try {
  const output = path.join(scratch, "playwright-output");
  mkdirSync(output, { mode: 0o700 });
  for (const name of [".playwright-artifacts-0", ".playwright-artifacts-12", ".playwright-artifacts-other", "completed-test"]) {
    mkdirSync(path.join(output, name), { mode: 0o700 });
    writeFileSync(path.join(output, name, "trace.zip"), "private fixture", { mode: 0o600 });
  }
  removePlaywrightWorkingTraces(output);
  removePlaywrightWorkingTraces(output);
  assert.equal(existsSync(path.join(output, ".playwright-artifacts-0")), false);
  assert.equal(existsSync(path.join(output, ".playwright-artifacts-12")), false);
  assert.equal(existsSync(path.join(output, ".playwright-artifacts-other/trace.zip")), true);
  assert.equal(existsSync(path.join(output, "completed-test/trace.zip")), true);
  symlinkSync(path.join(output, "completed-test"), path.join(output, ".playwright-artifacts-1"));
  assert.throws(() => removePlaywrightWorkingTraces(output));
  assert.equal(existsSync(path.join(output, "completed-test/trace.zip")), true);
  write("tools/toolchain_pins.json", { ui_review: { playwright: installed.version, "playwright-core": installed.version } });
  write("package.json", { devDependencies: { playwright: installed.version, "playwright-core": installed.version } });
  write("apps/web/package.json", { devDependencies: { "@playwright/test": installed.version } });
  const descriptor = { browsers: [installed.chromium] };
  for (const dir of ["node_modules/playwright", "node_modules/playwright-core", "apps/web/node_modules/@playwright/test", "apps/web/node_modules/@playwright/test/node_modules/playwright", "apps/web/node_modules/@playwright/test/node_modules/playwright/node_modules/playwright-core"]) {
    write(`${dir}/package.json`, { name: dir.split("/").at(-1), version: installed.version });
    if (dir.endsWith("playwright-core")) write(`${dir}/browsers.json`, descriptor);
  }
  assert.equal(resolvePlaywrightPackages(scratch).version, installed.version, "ordinary and nested layouts resolve without pnpm paths");
  for (const manifest of ["node_modules/playwright/package.json", "node_modules/playwright-core/package.json", "apps/web/node_modules/@playwright/test/package.json", "apps/web/node_modules/@playwright/test/node_modules/playwright/package.json"]) {
    write(manifest, { version: "0.0.0" }); assert.throws(() => resolvePlaywrightPackages(scratch), /version mismatch/u);
    write(manifest, { version: installed.version });
  }
  write("tools/frontend_visual_renderer_profile.json", { ...profile, profile_id: "visual.renderer.future", playwright_version: "2.0.0" });
  assert.equal(loadVisualRendererProfile(scratch).playwright_version, "2.0.0", "structural schema is independent of the active pin");
  write("tools/frontend_visual_renderer_profile.json", { ...profile, container_image: "mutable:latest" });
  assert.throws(() => loadVisualRendererProfile(scratch), /validation failed/u);
  const nested = "apps/web/node_modules/@playwright/test/node_modules/playwright/node_modules/playwright-core";
  write(`${nested}/package.json`, { version: "0.0.0" });
  assert.throws(() => resolvePlaywrightPackages(scratch), /version mismatch/u);
  write(`${nested}/package.json`, { version: installed.version });
  write(`${nested}/browsers.json`, { browsers: [{ ...installed.chromium, revision: "0" }] });
  assert.throws(() => resolvePlaywrightPackages(scratch), /descriptors disagree/u);
  write(`${nested}/browsers.json`, descriptor);
  rmSync(path.join(scratch, "node_modules/playwright-core"), { recursive: true });
  assert.throws(() => resolvePlaywrightPackages(scratch), /Cannot find module|ENOENT/u);
  const manifest = { families: [{ files: [{ path: "sample.woff2", bytes: 4, sha256: sha256("font") }] }] };
  write("public/assets/fonts/FONT_MANIFEST.json", manifest);
  writeFileSync(path.join(scratch, "public/assets/fonts/sample.woff2"), "font");
  const fontProfile = { ...profile, font_manifest_sha256: sha256(JSON.stringify(manifest)) };
  verifyRendererFonts(scratch, fontProfile, path.join(scratch, "public"));
  writeFileSync(path.join(scratch, "public/assets/fonts/sample.woff2"), "oops");
  assert.throws(() => verifyRendererFonts(scratch, fontProfile, path.join(scratch, "public")), /font bytes mismatch/u);
  assert.throws(() => verifyRendererFonts(scratch, profile, path.join(scratch, "public")), /manifest mismatch/u);
} finally { rmSync(scratch, { recursive: true, force: true }); }

for (const image of [{ Os: "linux", Architecture: "arm64", RepoDigests: [profile.container_image] }, { Os: "linux", Architecture: "amd64", RepoDigests: [] }]) {
  assert.throws(() => verifyRendererImage(profile, () => JSON.stringify([image])), /identity mismatch/u);
}
let calls = [], present = true;
const release = rendererRelease("owned-id", (args) => {
  calls.push(args);
  if (args[0] === "rm" && present) throw new Error("removal failed");
  return present ? "owned-id" : "";
});
assert.throws(release, /removal failed/u);
present = false; release(); const count = calls.length; release(); assert.equal(calls.length, count);
assert.ok(calls.every((args) => args.includes("owned-id") || args.includes("id=owned-id")));

const proof = { daemon_id: "fixture-daemon", image_id: `sha256:${"a".repeat(64)}`, token: "b".repeat(32), name: `cartulary-visual-123-${"b".repeat(12)}` };
for (const changed of [{ token: "../escape" }, { name: "borrowed-container" }, { image_id: "mutable:latest" }, { extra: true }]) assert.throws(() => validateRendererProof({ ...proof, ...changed }));
for (const fault of [null, "absent", "daemon", "ambiguous", "image", "label", "name", "removal"]) {
  let live = fault !== "absent", removals = 0;
  const id = "c".repeat(64);
  const run = (args) => {
    if (args[0] === "info") return fault === "daemon" ? "unrelated-daemon" : proof.daemon_id;
    if (args[0] === "ps") {
      assert.ok(args.includes(`name=^/${proof.name}$`) && args.includes(`label=cartulary.visual-owner=${proof.token}`));
      return live ? fault === "ambiguous" ? `${id}\n${"d".repeat(64)}` : id : "";
    }
    if (args[0] === "inspect") return JSON.stringify([{ Id: id, Name: fault === "name" ? "/borrowed" : `/${proof.name}`, Image: fault === "image" ? "wrong-image" : proof.image_id, Config: { Labels: { "cartulary.visual-owner": fault === "label" ? "wrong-owner" : proof.token } } }]);
    if (args[0] === "rm") { assert.equal(args.at(-1), id); removals++; if (fault === "removal") throw new Error("injected removal failure"); live = false; return ""; }
    assert.fail("unexpected Docker operation");
  };
  if (fault && fault !== "absent") assert.throws(() => removeOwnedRenderer(proof, run));
  else { removeOwnedRenderer(proof, run); removeOwnedRenderer(proof, run); assert.equal(live, false); }
  assert.equal(removals, fault === null || fault === "removal" ? 1 : 0);
}

const publishedPorts = { "3000/tcp": [{ HostIp: "127.0.0.1", HostPort: "43121" }] };
assert.deepEqual(rendererPublishedPorts(publishedPorts, false), { renderer: { containerPort: 3000, hostPort: 43121 }, diagnostic: null });
for (const invalid of [null, [], {}, { ...publishedPorts, "3001/tcp": publishedPorts["3000/tcp"] }, { "3000/tcp": null }, { "3000/tcp": [] }, { "3000/tcp": [...publishedPorts["3000/tcp"], ...publishedPorts["3000/tcp"]] }]) assert.throws(() => rendererPublishedPorts(invalid, false));
for (const binding of [{ HostIp: "0.0.0.0", HostPort: "43121" }, { HostIp: "::1", HostPort: "43121" }, { HostIp: "127.0.0.1", HostPort: 43121 }, ...["0", "01", "65536", "-1", "12.5", "123/private"].map((HostPort) => ({ HostIp: "127.0.0.1", HostPort }))]) assert.throws(() => rendererPublishedPorts({ "3000/tcp": [binding] }, false));
assert.throws(() => rendererPublishedPorts({ ...publishedPorts, "3001/tcp": publishedPorts["3000/tcp"] }, true), /duplicate/u);
for (const [result, reason] of [
  [{ error: { code: "ETIMEDOUT" } }, "command_timeout"],
  [{ error: { code: "ENOENT" } }, "command_unavailable"],
  [{ signal: "SIGTERM" }, "command_signalled"],
  [{ stderr: "secret payload: port is already allocated" }, "port_unavailable"],
  [{ stderr: "secret payload: Cannot connect to the Docker daemon" }, "daemon_unavailable"],
  [{ stderr: "secret payload: permission denied" }, "permission_denied"],
  [{ stderr: "secret payload: no space left on device" }, "storage_exhausted"],
  [{ stderr: "secret payload: unknown error" }, "command_failed"],
]) assert.equal(rendererLifecycleFailureReason(result), reason);

for (const diagnostics of [false, true]) for (const failAt of [null, "create", "cp", "start", "exec", "inspect", "ports", "browser", "package", "revision", "connect", "state_unavailable", "invalid_state", "cleanup"]) {
  calls = []; present = false;
  const run = (args) => {
    calls.push(args);
    if (args[0] === "create" && failAt === "create") { present = true; throw new Error("injected acquisition failure"); }
    if (args[0] === failAt) throw new Error("injected acquisition failure");
    if (args[0] === "info") return "fixture-daemon";
    if (args[0] === "image") return JSON.stringify([{ Id: `sha256:${"a".repeat(64)}`, Os: "linux", Architecture: "amd64", RepoDigests: [profile.container_image] }]);
    if (args[0] === "create") { present = true; return "owned-id"; }
    if (args[0] === "exec") return JSON.stringify({ playwright: profile.playwright_version, core: failAt === "package" ? "0.0.0" : profile.playwright_version, revision: failAt === "revision" ? "0" : profile.chromium_revision, version: profile.chromium_version });
    if (args[0] === "inspect" && args.at(-1) === "{{json .NetworkSettings.Ports}}") return JSON.stringify(failAt === "ports" ? { "3000/tcp": [{ HostIp: "0.0.0.0", HostPort: "43121" }] } : { ...publishedPorts, ...(diagnostics ? { "3001/tcp": [{ HostIp: "127.0.0.1", HostPort: "43122" }] } : {}) });
    if (args[0] === "inspect" && args.at(-1) === "{{json .State}}") {
      assert.equal(present, true, "observe exact owned container before removal");
      if (failAt === "state_unavailable") throw new Error("private daemon diagnostic");
      return JSON.stringify({ Status: failAt === "invalid_state" ? "private status" : "running", Running: true, OOMKilled: false, ExitCode: 0, Pid: 987654, Error: "private daemon diagnostic" });
    }
    if (args[0] === "rm" && failAt === "cleanup") throw new Error("private cleanup diagnostic");
    if (args[0] === "inspect") return `sha256:${"a".repeat(64)}`;
    if (args[0] === "rm") present = false;
    if (args[0] === "ps" && present && ["create", "cleanup"].includes(failAt)) return "a".repeat(64);
    return "";
  };
  const resources = [];
  const start = () => startVisualRendererLease({ root, environment: {}, diagnostics, onOwnedResource: (resource) => resources.push(resource), run, trust: () => ({ attestation: { schema_id: "cartulary.browser_tls_trust.v1" }, initialize() {} }), endpointReady: async (_containerID, binding) => { assert.deepEqual(binding, { containerPort: 3000, hostPort: 43121 }); return "ws://127.0.0.1:43121/private"; }, connect: async (_packages, endpoint) => { assert.equal(endpoint, "ws://127.0.0.1:43121/private"); if (["connect", "state_unavailable", "invalid_state", "cleanup"].includes(failAt)) throw new Error("browserType.connect: private ws://127.0.0.1:43121/private"); return failAt === "browser" ? "0.0.0.0" : profile.chromium_version; } });
  if (["connect", "state_unavailable", "invalid_state", "cleanup"].includes(failAt)) {
    await assert.rejects(start, (error) => {
      assert.equal(error.failure_class, "infra");
      assert.equal(error.failure_reason, "service_start_error");
      assert.match(error.message, /renderer connection failed/u);
      assert.ok(!/private|ws:\/\/|987654/u.test(error.message), "startup diagnostics exclude endpoint, raw error and process identity");
      assert.deepEqual(error.rendererState, ["state_unavailable", "invalid_state"].includes(failAt)
        ? { status: "unavailable" }
        : { status: "running", running: true, oom_killed: false, exit_code: 0 });
      assert.equal(error.cleanupFailures?.length ?? 0, failAt === "cleanup" ? 1 : 0);
      return true;
    });
    assert.equal(calls.filter(args => args[0] === "inspect" && args.at(-1) === "{{json .State}}").length, 1);
  }
  else if (failAt) await assert.rejects(start, /injected acquisition failure|mismatch|renderer.*port/u);
  else {
    const lease = await start(); assert.equal(lease.attestation.observed.chromium_version, profile.chromium_version); assert.ok(lease.attestation.observed.font_files.length);
    assert.deepEqual(lease.diagnosticBinding, diagnostics ? { containerPort: 3001, hostPort: 43122 } : null);
    const reordered = structuredClone(lease.attestation);
    reordered.profile = Object.fromEntries(Object.entries(reordered.profile).reverse());
    reordered.observed.font_files = reordered.observed.font_files.map((file) => Object.fromEntries(Object.entries(file).reverse()));
    assert.equal(validateRendererAttestation(reordered, profile, lease.attestation.observed.font_files), reordered);
    const tampered = structuredClone(lease.attestation); tampered.observed.font_files[0].sha256 = "0".repeat(64);
    assert.throws(() => validateRendererAttestation(tampered, profile, lease.attestation.observed.font_files), /font files mismatch/u);
    assert.throws(() => validateRendererAttestation(lease.attestation, { ...profile, profile_id: "visual.renderer.stale" }, lease.attestation.observed.font_files), /profile mismatch/u);
    lease.cleanup(); lease.cleanup();
  }
  assert.equal(present, failAt === "cleanup");
  assert.equal(resources[0].state, "pending");
  assert.equal(resources.at(-1).state, failAt === "cleanup" ? "acquired" : "released");
  const create = calls.find((args) => args[0] === "create");
  assert.equal(create[create.indexOf("--publish") + 1], "127.0.0.1::3000");
  assert.equal(create.includes("127.0.0.1::3001"), diagnostics);
  assert.equal(create[create.indexOf("--port") + 1], "3000");
  assert.equal(create[create.indexOf("--user") + 1], "pwuser");
  assert.ok(!create.includes("--volume") && !create.includes("--network"));
}
