import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { createServer } from "node:http";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { crc32, decodePNG, encodePNG, pngHeader } from "../ui-review/png.mjs";
import { artifact, bundleBase, loadBundle, publishBundle, validateFiles } from "../ui-review/bundles.mjs";
import { execute, scopeRectangle } from "../ui-review/capture.mjs";
import { importCanonical, selectCaptureResult } from "../ui-review/canonical-import.mjs";
import { importFixture } from "./ui-review-import-fixture.mjs";
import { ReviewSession } from "../ui-review/session.mjs";
import { ReviewBrowser } from "../ui-review/browser.mjs";
import { limits, parseRequest, schemaID } from "../ui-review/contract.mjs";
import { toolProfile, repoRoot } from "../ui-review/toolchain.mjs";
import { acquireHostAdmission } from "../../runtime/host-admission.mjs";
import { captureCapabilitySnapshot } from "../../scheduler/work-graph/capability.mjs";

const sessionAt = (root) => Object.assign(new ReviewSession({ UI_MODE: "artifacts" }, toolProfile()), { runtime: { privatePath: (...parts) => path.join(root, ...parts) }, workspaceDigest: "a".repeat(64) });
const png = () => encodePNG({ width: 2, height: 1, data: Buffer.from([11, 22, 33, 0, 44, 55, 66, 255]) });
function insertChunk(bytes, kind, payload) {
  const chunk = Buffer.alloc(payload.length + 12); chunk.writeUInt32BE(payload.length); chunk.write(kind, 4); payload.copy(chunk, 8); chunk.writeUInt32BE(crc32(chunk.subarray(4, -4)), chunk.length - 4);
  return Buffer.concat([bytes.subarray(0, 33), chunk, bytes.subarray(33)]);
}
test("PNG decoding preserves transparent channels and rejects unsupported, corrupt and oversized inputs", async () => {
  const bytes = await png(), image = await decodePNG(bytes);
  assert.deepEqual([...image.data], [11, 22, 33, 0, 44, 55, 66, 255]);
  for (const value of [bytes.subarray(0, -1), Buffer.concat([bytes, Buffer.from([0])]), insertChunk(bytes, "acTL", Buffer.alloc(8)), insertChunk(bytes, "iCCP", Buffer.from("unsupported"))]) await assert.rejects(decodePNG(value), /invalid_artifact/u);
  const oversized = Buffer.from(bytes); oversized.writeUInt32BE(limits.dimension + 1, 16); oversized.writeUInt32BE(crc32(oversized.subarray(12, 29)), 29); assert.throws(() => pngHeader(oversized));
  const edge = Buffer.from(bytes); edge.writeUInt32BE(8192, 16); edge.writeUInt32BE(2048, 20); edge.writeUInt32BE(crc32(edge.subarray(12, 29)), 29); assert.deepEqual(pngHeader(edge), { width: 8192, height: 2048 });
  edge.writeUInt32BE(2049, 20); edge.writeUInt32BE(crc32(edge.subarray(12, 29)), 29); assert.throws(() => pngHeader(edge));
  assert.throws(() => pngHeader(Buffer.alloc(limits.png + 1)));
  const corrupt = Buffer.from(bytes); corrupt[40] ^= 1; await assert.rejects(decodePNG(corrupt));
});
test("encoded PNG and logical bundle byte limits accept the exact boundary and reject one byte over", async () => {
  const bytes = await png(), boundary = insertChunk(bytes, "ruSt", Buffer.alloc(limits.png - bytes.length - 12));
  assert.equal(boundary.length, limits.png); assert.deepEqual(pngHeader(boundary), { width: 2, height: 1 });
  assert.throws(() => pngHeader(Buffer.concat([boundary, Buffer.from([0])])));
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-byte-bounds-"));
  try {
    const session = sessionAt(root), bundle = bundleBase(session, 1), files = new Map();
    bundle.source = { kind: "reference_image", import_ref: { input_path: "/private/reference.png", input_sha256: "a".repeat(64), metadata: null } };
    bundle.components.original = artifact("original.png", boundary, "image/png"); files.set("original.png", boundary);
    for (let i = 0; i < 3; i++) { const name = `copy-${i}.png`; files.set(name, boundary); bundle.derived.push({ kind: "overlay", ref: artifact(name, boundary, "image/png"), source_refs: [bundle.components.original], rectangle: null }); }
    assert.equal(await validateFiles(bundle, files), limits.bundle);
    const excess = Buffer.from([1]); files.set("extra.zip", excess); bundle.components.trace = artifact("extra.zip", excess, "application/zip");
    await assert.rejects(validateFiles(bundle, files), /observation_limit/u);
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("native scope transforms use actual screenshot scale and outward rounding without applying CSS zoom again", () => {
  assert.deepEqual(scopeRectangle({ kind: "element" }, { x: 0.99, y: 2.01, width: 3.02, height: 4 }, { width: 10, height: 10 }, { width: 20, height: 30 }), { rectangle: { x: 1, y: 6, width: 8, height: 13 }, transform: { origin_x: 0.5, origin_y: 2, scale_x: 2, scale_y: 3 } });
  assert.throws(() => scopeRectangle({ kind: "region", x: 8, y: 0, width: 3, height: 1 }, null, { width: 10, height: 10 }, { width: 10, height: 10 }));
  assert.throws(() => scopeRectangle({ kind: "element" }, null, { width: 10, height: 10 }, { width: 10, height: 10 }));
});
test("reference imports freeze borrowed bytes, enforce reservations and never overwrite published bundles", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-bundles-")), session = sessionAt(root), file = path.join(root, "input.png");
  try {
    const bytes = await png(); writeFileSync(file, bytes, { mode: 0o600 });
    const result = await execute(session, { source: "image", path: file }, 1);
    assert.equal(result.bundle_id, "bundle-1"); assert.equal(result.counts.images, 1);
    const first = await loadBundle(session, "bundle-1"); assert.equal(first.bundle.source.kind, "reference_image");
    writeFileSync(file, "changed borrowed input");
    assert.deepEqual((await loadBundle(session, "bundle-1")).files.get("original.png"), bytes);
    await assert.rejects(publishBundle(session, first.bundle, first.files), /capacity_exceeded/u);
    const manifest = readFileSync(path.join(first.root, "bundle.json"));
    session.privateBytes = limits.storage;
    const next = { ...first.bundle, bundle_id: "bundle-2" };
    await assert.rejects(publishBundle(session, next, first.files), /capacity_exceeded/u);
    assert.deepEqual(readFileSync(path.join(first.root, "bundle.json")), manifest);
    session.privateBytes = 0; mkdirSync(path.join(root, "bundles/bundle-2"));
    await assert.rejects(publishBundle(session, next, first.files));
    writeFileSync(path.join(first.root, "original.png"), Buffer.from("tampered"));
    await assert.rejects(loadBundle(session, "bundle-1"), /invalid_artifact/u);
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("bundle publication rejects missing, mismatched and excessive components before exposure", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-bundle-limits-")), session = sessionAt(root);
  try {
    const bytes = await png(), bundle = bundleBase(session, 1);
    bundle.source = { kind: "reference_image", import_ref: { input_path: "/private/input.png", input_sha256: "a".repeat(64), metadata: null } };
    bundle.components.original = artifact("original.png", bytes, "image/png");
    await assert.rejects(publishBundle(session, bundle, new Map()), /observation_limit/u);
    await assert.rejects(publishBundle(session, bundle, new Map([["original.png", Buffer.from("bad")]])), /invalid_artifact/u);
    assert.equal(session.bundles?.size ?? 0, 0); assert.equal(session.privateBytes, 0);
    const files = new Map([["original.png", bytes]]);
    for (let index = 0; index < 62; index++) {
      const name = `derived-${index}.png`; files.set(name, bytes);
      bundle.derived.push({ kind: "overlay", ref: artifact(name, bytes, "image/png"), source_refs: [bundle.components.original], rectangle: null });
    }
    await publishBundle(session, bundle, files); // 63 components plus manifest.
    files.set("excess.png", bytes); bundle.derived.push({ kind: "overlay", ref: artifact("excess.png", bytes, "image/png"), source_refs: [bundle.components.original], rectangle: null });
    await assert.rejects(publishBundle(session, { ...bundle, bundle_id: "bundle-2" }, files), /observation_limit/u);
    session.bundles = new Map(Array.from({ length: limits.bundles }, (_, index) => [`bundle-${index + 1}`, {}]));
    await assert.rejects(publishBundle(session, { ...bundle, bundle_id: "bundle-101" }, files), /capacity_exceeded/u);
    session.privateBytes = 0; const release = session.reserveBytes(limits.storage); assert.throws(() => session.reserveBytes(1), /capacity_exceeded/u); release(); assert.equal(session.privateBytes, 0);
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("capture joins select only the exact pinned report result and reject duplicate associations", () => {
  const profile = { browser_zoom_percent: 100, color_scheme: "light", density_id: null, device_scale_factor: 1, reduced_motion: true, theme_id: "dark_graphite", project_id: "visual", snapshot_path_template: "{snapshotDir}/{testFileDir}/{testFileName}-snapshots/{arg}{-snapshotSuffix}{ext}", snapshot_suffix: "linux", viewport_css_px: "1440x900", expected_density_id: null, expected_theme_id: "dark_graphite", surface_kind: "application_shell" };
  const capture = { capture_id: `visual.capture.${"a".repeat(20)}`, capture_intent: "selected", expected_golden_path: "apps/web/e2e/workbook.visual.spec.ts-snapshots/selected-linux.png", project_id: "visual", renderer_profile_id: "visual.renderer.playwright_1_59_1_chromium_1217_linux_amd64", screenshot_assertion_location: "capture:1", assertion_file: "apps/web/e2e/workbook.visual.spec.ts", test_title: "Exact selected test", capture_profile: profile };
  const { assertion_file, ...fields } = capture;
  const payload = { schema_id: "cartulary.frontend_visual_capture_intent.v2", ...fields, test_file: assertion_file };
  const attachment = { name: `cartulary-visual-capture-intent-${capture.capture_id}.json`, contentType: "application/json", body: Buffer.from(JSON.stringify(payload)).toString("base64") };
  const result = { attachments: [attachment] }, report = { config: { version: "1.59.1", rootDir: path.join(repoRoot, "apps/web/e2e") }, suites: [{ specs: [{ title: capture.test_title, file: "workbook.visual.spec.ts", tests: [{ projectName: "visual", results: [result] }] }] }] };
  assert.equal(selectCaptureResult(report, capture, "/private/run"), result);
  result.attachments.push(attachment); assert.throws(() => selectCaptureResult(report, capture, "/private/run")); result.attachments.pop();
  report.config.version = "1.58.0"; assert.throws(() => selectCaptureResult(report, capture, "/private/run")); report.config.version = "1.59.1";
  report.suites[0].specs[0].title = "Different test"; assert.throws(() => selectCaptureResult(report, capture, "/private/run"));
});

test("canonical imports accept failed expected-only evidence, freeze metadata, and reject mixed or tampered identities", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-import-"));
  try {
    const fixture = importFixture(root);
    const first = await importCanonical(fixture.request);
    assert.equal(first.metadata.components.actual, null); assert.ok(first.metadata.limitations.includes("no_actual"));
    assert.equal(first.metadata.source.import_ref.metadata.reconciliation.status, "fail");
    const original = first.files.get("expected.png");
    writeFileSync(path.join(root, "actual.png"), original, { mode: 0o600 });
    fixture.result.attachments.push({ name: "auth-initial-actual.png", contentType: "image/png", path: path.join(root, "actual.png") });
    fixture.result.attachments.push({ name: "auth-initial-expected.png", contentType: "image/png", path: path.join(repoRoot, fixture.goldenPath) });
    fixture.publish(); const withActual = await importCanonical(fixture.request);
    assert.deepEqual(withActual.files.get("actual.png"), original);
    assert.equal(withActual.metadata.limitations.includes("no_actual"), false);
    fixture.result.attachments.push(fixture.result.attachments[1]); fixture.publish();
    await assert.rejects(importCanonical(fixture.request), /invalid_artifact/u); fixture.result.attachments.pop();
    const previous = fixture.group.runtime_profile_id; fixture.group.runtime_profile_id = "wrong_profile"; fixture.publish();
    await assert.rejects(importCanonical(fixture.request), /invalid_artifact/u); fixture.group.runtime_profile_id = previous;
    const schema = fixture.reconciliation.schema_id; fixture.reconciliation.schema_id = "cartulary.frontend_visual_reconciliation.v2"; fixture.publish();
    await assert.rejects(importCanonical(fixture.request), /invalid_artifact/u); fixture.reconciliation.schema_id = schema;
    fixture.reconciliation.source_refs.find((entry) => entry.kind === "fixture_registry").sha256 = "0".repeat(64); fixture.publish();
    await assert.rejects(importCanonical(fixture.request), /invalid_artifact/u);
    assert.deepEqual(first.files.get("expected.png"), original);
    assert.equal(first.metadata.source.import_ref.metadata.reconciliation.schema_id, "cartulary.frontend_visual_reconciliation.v3");
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("page capture preserves focus and scroll, crops exactly, and rejects post-screenshot geometry changes", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-capture-")), session = sessionAt(root);
  const server = createServer((_request, response) => { response.setHeader("content-type", "text/html"); response.end('<html lang="en"><title>Capture fixture</title><body style="margin:0"><button data-testid="box" style="position:absolute;left:10.25px;top:20.5px;width:31px;height:22px">Capture</button></body></html>'); });
  let admission, browser;
  try {
    await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
    admission = await acquireHostAdmission({ browsers: 1, browserCapacity: captureCapabilitySnapshot({ root: repoRoot }).port_lanes });
    browser = await new ReviewBrowser({ origin: `http://127.0.0.1:${server.address().port}`, mode: "dev" }).start();
    session.mode = "dev"; session.browser = browser;
    await browser.page.getByTestId("box").focus();
    await browser.page.evaluate(() => {
      Object.defineProperty(document.fonts, "ready", { value: new Promise((resolve) => setTimeout(resolve, 120)) });
      const node = document.querySelector("button"); node.style.fontFamily = "monospace";
      let count = 0; const update = () => { node.textContent = count++ % 2 ? "ABCD" : "WXYZ"; requestAnimationFrame(update); }; requestAnimationFrame(update);
    });
    const target = { kind: "test_id", value: "box" };
    const request = parseRequest(Buffer.from(JSON.stringify({ schema_id: schemaID("capture_request"), source: "page", expected_epoch: 0, scope: { kind: "element", target }, targets: [target], include_axe: false })), "capture_request");
    const start = performance.now();
    const produced = await execute(session, request, 1), { bundle } = await loadBundle(session, produced.bundle_id);
    assert.ok(performance.now() - start >= 100);
    assert.equal(bundle.source.kind, "live_unattested"); assert.deepEqual(bundle.observation.image_dimensions, { width: 32, height: 23 });
    assert.deepEqual(bundle.observation.coordinate_transform, { origin_x: 10, origin_y: 20, scale_x: 1, scale_y: 1 });
    assert.equal(await browser.page.getByTestId("box").evaluate((node) => document.activeElement === node), true);
    const screenshot = browser.page.screenshot.bind(browser.page);
    browser.page.screenshot = async (options) => { const bytes = await screenshot(options); await browser.page.getByTestId("box").evaluate((node) => { node.style.left = "30px"; }); return bytes; };
    await assert.rejects(execute(session, request, 2), /unstable_capture/u); assert.equal(session.bundles.size, 1);
    browser.page.screenshot = screenshot;
    await browser.page.getByTestId("box").evaluate((node) => {
      let count = 0; const move = () => { node.style.left = `${20 + count++ * 0.25}px`; requestAnimationFrame(move); }; requestAnimationFrame(move);
    });
    await assert.rejects(execute(session, request, 3), /operation_expired/u); assert.equal(session.bundles.size, 1);
  } finally { await browser?.close(); await admission?.release(); await new Promise((resolve) => server.close(resolve)); rmSync(root, { recursive: true, force: true }); }
});
