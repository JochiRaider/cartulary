import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, rmSync, existsSync, symlinkSync } from "node:fs";
import path from "node:path";
import os from "node:os";
import { resolvePlaywrightPackages } from "../../readiness/playwright-packages.mjs";
import { loadVisualRendererProfile, verifyRendererFonts, validateRendererAttestation, sha256 } from "../visual-renderer-profile.mjs";
import { rendererRelease, verifyRendererImage, startVisualRendererLease } from "../visual-renderer-lease.mjs";
import { removePlaywrightWorkingTraces } from "../playwright-output-cleanup.mjs";

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

for (const failAt of [null, "create", "cp", "start", "exec", "inspect", "browser", "package", "revision"]) {
  calls = []; present = false;
  const run = (args) => {
    calls.push(args);
    if (args[0] === "create" && failAt === "create") { present = true; throw new Error("injected acquisition failure"); }
    if (args[0] === failAt) throw new Error("injected acquisition failure");
    if (args[0] === "image") return JSON.stringify([{ Id: `sha256:${"a".repeat(64)}`, Os: "linux", Architecture: "amd64", RepoDigests: [profile.container_image] }]);
    if (args[0] === "create") { present = true; return "owned-id"; }
    if (args[0] === "exec") return JSON.stringify({ playwright: profile.playwright_version, core: failAt === "package" ? "0.0.0" : profile.playwright_version, revision: failAt === "revision" ? "0" : profile.chromium_revision, version: profile.chromium_version });
    if (args[0] === "inspect") return `sha256:${"a".repeat(64)}`;
    if (args[0] === "rm") present = false;
    if (args[0] === "ps" && present && failAt === "create") return "a".repeat(64);
    return "";
  };
  const start = () => startVisualRendererLease({ root, environment: {}, run, endpointReady: async () => "ws://127.0.0.1:1/private", connect: async () => failAt === "browser" ? "0.0.0.0" : profile.chromium_version });
  if (failAt) await assert.rejects(start, /injected acquisition failure|mismatch/u);
  else {
    const lease = await start(); assert.equal(lease.attestation.observed.chromium_version, profile.chromium_version); assert.ok(lease.attestation.observed.font_files.length);
    const reordered = structuredClone(lease.attestation);
    reordered.profile = Object.fromEntries(Object.entries(reordered.profile).reverse());
    reordered.observed.font_files = reordered.observed.font_files.map((file) => Object.fromEntries(Object.entries(file).reverse()));
    assert.equal(validateRendererAttestation(reordered, profile, lease.attestation.observed.font_files), reordered);
    const tampered = structuredClone(lease.attestation); tampered.observed.font_files[0].sha256 = "0".repeat(64);
    assert.throws(() => validateRendererAttestation(tampered, profile, lease.attestation.observed.font_files), /font files mismatch/u);
    assert.throws(() => validateRendererAttestation(lease.attestation, { ...profile, profile_id: "visual.renderer.stale" }, lease.attestation.observed.font_files), /profile mismatch/u);
    lease.cleanup(); lease.cleanup();
  }
  assert.equal(present, false);
  const create = calls.find((args) => args[0] === "create");
  assert.match(create[create.indexOf("--publish") + 1], /^127\.0\.0\.1:/u);
  assert.equal(create[create.indexOf("--user") + 1], "pwuser");
  assert.ok(!create.includes("--volume") && !create.includes("--network"));
}
