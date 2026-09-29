import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { createServer } from "node:http";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { crc32, decodePNG, encodePNG, pngHeader } from "../ui-review/png.mjs";
import { artifact, bundleBase, validateBundleBudget } from "../ui-review/bundles.mjs";
import { scopeRectangle } from "../ui-review/capture.mjs";
import { importCanonical, selectCaptureResult, canonicalComparison } from "../ui-review/canonical-import.mjs";
import { capture as execute } from "./ui-review-work-fixture.mjs";
import { importFixture } from "./ui-review-import-fixture.mjs";
import { ArtifactStore } from "../ui-review/artifact-store.mjs";
import { atomicLocalFile, removePrivateTree } from "../../runtime/secure-local-files.mjs";
import { ReviewSession } from "../ui-review/session.mjs";
import { ReviewBrowser } from "../ui-review/browser.mjs";
import { limits, parseRequest, schemaID } from "../ui-review/contract.mjs";
import { toolProfile } from "../ui-review/toolchain.mjs";
import { repoRoot, reviewPins } from "../ui-review/policy.mjs";
import { acquireHostAdmission, inheritedHostLease } from "../../runtime/host-admission.mjs";
import { captureCapabilitySnapshot, resourceCapacities } from "../../scheduler/work-graph/capability.mjs";

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
  const files = new Map(Array.from({ length: 4 }, (_, index) => [`image-${index}.png`, boundary]));
  assert.equal(validateBundleBudget(files), limits.bundle);
  files.set("extra.zip", Buffer.from([1])); assert.throws(() => validateBundleBudget(files), /observation_limit/u);
  const many = new Map(Array.from({ length: 63 }, (_, index) => [`image-${index}.png`, bytes]));
  validateBundleBudget(many); many.set("excess.png", bytes); assert.throws(() => validateBundleBudget(many), /observation_limit/u);
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
    const first = await session.store.loadBundle("bundle-1"); assert.equal(first.bundle.source.kind, "reference_image");
    writeFileSync(file, "changed borrowed input");
    assert.deepEqual((await session.store.loadBundle("bundle-1")).files.get("original.png"), bytes);
    await assert.rejects(session.store.publishBundle(first.bundle, first.files), /capacity_exceeded/u);
    const manifest = readFileSync(path.join(first.root, "bundle.json"));
    const constrained = new ArtifactStore({ sessionID: session.sessionID, profile: session.profile, signal: session.abort.signal, privatePath: (...parts) => path.join(root, "bounded", ...parts), capacity: session.store.usage.bytes });
    await constrained.publishBundle(first.bundle, first.files);
    const next = { ...first.bundle, bundle_id: "bundle-2" };
    await assert.rejects(constrained.publishBundle(next, first.files), /capacity_exceeded/u);
    assert.deepEqual(readFileSync(path.join(first.root, "bundle.json")), manifest);
    constrained.close(); assert.equal(constrained.usage.bytes, 0);
    mkdirSync(path.join(root, "artifacts/bundles/bundle-2"));
    await assert.rejects(session.store.publishBundle(next, first.files));
    writeFileSync(path.join(first.root, "original.png"), Buffer.from("tampered"));
    await assert.rejects(session.store.loadBundle("bundle-1"), /invalid_artifact/u);
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("bundle publication rejects missing, mismatched and excessive components before exposure", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-bundle-limits-")), session = sessionAt(root);
  try {
    const bytes = await png(), bundle = bundleBase(session, 1);
    bundle.source = { kind: "reference_image", import_ref: { input_path: "/private/input.png", input_sha256: "a".repeat(64), metadata: null } };
    bundle.components.original = artifact("original.png", bytes, "image/png");
    bundle.source.import_ref.input_sha256 = bundle.components.original.sha256;
    bundle.limitations = ["no_axe", "no_dom", "no_trace", "reference_only"];
    await assert.rejects(session.store.publishBundle(bundle, new Map()), /observation_limit/u);
    await assert.rejects(session.store.publishBundle(bundle, new Map([["original.png", Buffer.from("bad")]])), /invalid_artifact/u);
    assert.equal(session.store.usage.bundles, 0); assert.equal(session.store.usage.bytes, 0);
    const files = new Map([["original.png", bytes]]);
    for (let index = 1; index <= limits.bundles; index++) await session.store.publishBundle({ ...bundle, bundle_id: `bundle-${index}` }, files);
    await assert.rejects(session.store.publishBundle({ ...bundle, bundle_id: "bundle-101" }, files), /capacity_exceeded/u);
    assert.equal(session.store.usage.bundles, limits.bundles);
    session.store.close(); assert.equal(session.store.usage.bytes, 0);
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("artifact transactions retain primary and rollback failures and charge surviving staging", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-artifact-faults-"));
  const session = sessionAt(root), bytes = await png(), bundle = bundleBase(session, 1);
  bundle.components.original = artifact("original.png", bytes, "image/png");
  bundle.source = { kind: "reference_image", import_ref: { input_path: "/private/input.png", input_sha256: bundle.components.original.sha256, metadata: null } };
  bundle.limitations = ["no_axe", "no_dom", "no_trace", "reference_only"];
  const files = new Map([["original.png", bytes]]);
  try {
    for (const boundary of [1, 2]) for (const afterWrite of [false, true]) for (const rollbackFails of [false, true]) {
      let writes = 0, failRemove = rollbackFails;
      const primary = new Error("injected initiating write failure");
      const store = new ArtifactStore({ sessionID: session.sessionID, profile: session.profile, signal: session.abort.signal,
        privatePath: (...parts) => path.join(root, `${boundary}-${afterWrite}-${rollbackFails}`, ...parts),
        io: { write(file, content) { writes++; if (writes === boundary && !afterWrite) throw primary; atomicLocalFile(file, content); if (writes === boundary) throw primary; },
          remove(directory) { if (failRemove) throw new Error("injected rollback failure"); removePrivateTree(directory); } } });
      await assert.rejects(store.publishBundle(bundle, files), (error) => error === primary);
      assert.equal(store.usage.bundles, 0); await assert.rejects(store.loadBundle("bundle-1"), /invalid_request/u);
      assert.equal(store.usage.bytes > 0, rollbackFails);
      assert.equal(primary.cleanupFailures?.length ?? 0, rollbackFails ? 1 : 0);
      failRemove = false; store.close(); assert.equal(store.usage.bytes, 0); assert.equal(store.usage.reservations, 0);
    }
    const abort = new AbortController();
    const cancelled = new ArtifactStore({ sessionID: session.sessionID, profile: session.profile, signal: abort.signal,
      privatePath: (...parts) => path.join(root, "cancelled", ...parts), io: { write(file, content) { atomicLocalFile(file, content); abort.abort(new Error("cancel before commit")); } } });
    await assert.rejects(cancelled.publishBundle(bundle, files), /cancel before commit/u);
    assert.deepEqual(cancelled.usage, { bytes: 0, bundles: 0, reservations: 0 });
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("report assets and browser observations share reservations and publication rollback", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-store-channels-")), session = sessionAt(root);
  try {
    const bytes = await png(), file = path.join(root, "input.png"); writeFileSync(file, bytes, { mode: 0o600 });
    await execute(session, { source: "image", path: file }, 1);
    const before = session.store.usage.bytes;
    const observed = { schema_id: schemaID("observations"), elements: [], fonts: [], accessibility_snapshot: null, axe: { status: "disabled", engine_version: null, scope: "main_document", violations: [], incomplete: [], unassessed_frames: 0 }, console: { records: [], truncated: false }, network: { records: [], truncated: false } };
    const refs = await session.store.observations(2, observed);
    assert.equal(session.store.usage.bytes - before, readFileSync(refs[0].absolute_path).length);
    await assert.rejects(session.store.report("bundle-1", "oversized", () => new Map([["index.html", Buffer.alloc(limits.report + 1)]])), /observation_limit/u);
    const after = session.store.usage.bytes;
    await session.store.report("bundle-1", "size", () => new Map([["image.png", bytes], ["index.html", Buffer.from("local report")]]));
    assert.equal(session.store.usage.bytes - after, bytes.length + Buffer.byteLength("local report"));
    session.store.close(); assert.deepEqual(session.store.usage, { bytes: 0, bundles: 0, reservations: 0 });
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("capture joins select only the exact pinned report result and reject duplicate associations", () => {
  const profile = { browser_zoom_percent: 100, color_scheme: "light", density_id: null, device_scale_factor: 1, reduced_motion: true, theme_id: "dark_graphite", project_id: "visual", snapshot_path_template: "{snapshotDir}/{testFileDir}/{testFileName}-snapshots/{arg}{-snapshotSuffix}{ext}", snapshot_suffix: "linux", viewport_css_px: "1440x900", expected_density_id: null, expected_theme_id: "dark_graphite", surface_kind: "application_shell" };
  const capture = { capture_id: `visual.capture.${"a".repeat(20)}`, capture_intent: "selected", expected_golden_path: "apps/web/e2e/workbook.visual.spec.ts-snapshots/selected-linux.png", project_id: "visual", renderer_profile_id: "visual.renderer.playwright_1_59_1_chromium_1217_linux_amd64", screenshot_assertion_location: "capture:1", assertion_file: "apps/web/e2e/workbook.visual.spec.ts", test_title: "Exact selected test", capture_profile: profile };
  const { assertion_file, ...fields } = capture;
  const payload = { schema_id: "cartulary.frontend_visual_capture_intent.v2", ...fields, test_file: assertion_file };
  const attachment = { name: `cartulary-visual-capture-intent-${capture.capture_id}.json`, contentType: "application/json", body: Buffer.from(JSON.stringify(payload)).toString("base64") };
  const result = { attachments: [attachment] }, report = { config: { version: reviewPins().playwright, rootDir: path.join(repoRoot, "apps/web/e2e") }, suites: [{ specs: [{ title: capture.test_title, file: "workbook.visual.spec.ts", tests: [{ projectName: "visual", results: [result] }] }] }] };
  assert.equal(selectCaptureResult(report, capture, "/private/run"), result);
  result.attachments.push(attachment); assert.throws(() => selectCaptureResult(report, capture, "/private/run")); result.attachments.pop();
  report.config.version = "0.0.0"; assert.throws(() => selectCaptureResult(report, capture, "/private/run")); report.config.version = reviewPins().playwright;
  report.suites[0].specs[0].title = "Different test"; assert.throws(() => selectCaptureResult(report, capture, "/private/run"));
});

test("canonical imports accept failed expected-only evidence, freeze metadata, and reject mixed or tampered identities", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-import-"));
  try {
    const fixture = importFixture(root);
    const first = await importCanonical(fixture.request);
    assert.ok(Object.isFrozen(first.metadata) && Object.isFrozen(first.metadata.source.import_ref.metadata));
    assert.equal(first.metadata.components.actual, null); assert.ok(first.metadata.limitations.includes("no_actual"));
    assert.equal(first.metadata.source.import_ref.metadata.reconciliation.status, "fail");
    const original = first.files.get("expected.png");
    const identity = canonicalComparison(first.metadata.source);
    fixture.manifest.source_digest = `sha256:${"b".repeat(64)}`; fixture.receipt.source_digest = fixture.manifest.source_digest;
    fixture.publish(); const revised = await importCanonical(fixture.request);
    assert.equal(revised.metadata.source.served_source_digest, "b".repeat(64));
    assert.deepEqual(canonicalComparison(revised.metadata.source), identity);
    assert.equal(first.metadata.source.served_source_digest, "a".repeat(64));
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
test("multiple captures in one failed producer result keep exact images and comparison identities separate", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-multiple-import-"));
  try {
    const first = importFixture(path.join(root, "run"), { captureName: "design-immediate-initial-loading", ownerID: "web.design" });
    const second = importFixture(path.join(root, "source"), { captureName: "design-delayed-initial-loading", ownerID: "web.design" });
    first.reconciliation.capture_intents.push(second.reconciliation.capture_intents[0]);
    first.reconciliation.goldens.push(second.reconciliation.goldens[0]);
    Object.assign(first.reconciliation.counts, { capture_intents: 2, committed_goldens: 2, active: 2 });
    first.result.attachments.push(...second.result.attachments);
    for (const [name, golden] of [["design-immediate-initial-loading", first.goldenPath], ["design-delayed-initial-loading", second.goldenPath]]) {
      const file = path.join(first.request.run_root, `${name}.png`); writeFileSync(file, readFileSync(path.join(repoRoot, golden)), { mode: 0o600 });
      first.result.attachments.push({ name: `${name}-actual.png`, contentType: "image/png", path: file });
    }
    first.publish();
    const selected = await importCanonical(first.request), other = await importCanonical({ ...first.request, capture_id: second.request.capture_id });
    assert.deepEqual(selected.files.get("actual.png"), readFileSync(path.join(repoRoot, first.goldenPath)));
    assert.deepEqual(other.files.get("actual.png"), readFileSync(path.join(repoRoot, second.goldenPath)));
    assert.notDeepEqual(canonicalComparison(selected.metadata.source), canonicalComparison(other.metadata.source));
    first.manifest.source_digest = `sha256:${"b".repeat(64)}`; first.receipt.source_digest = first.manifest.source_digest; first.publish();
    const revised = await importCanonical(first.request);
    assert.deepEqual(canonicalComparison(revised.metadata.source), canonicalComparison(selected.metadata.source));
    assert.notEqual(revised.metadata.source.served_source_digest, selected.metadata.source.served_source_digest);
    const results = first.report.suites[0].specs[0].tests[0].results;
    results.push(structuredClone(first.result)); first.publish();
    await assert.rejects(importCanonical(first.request), /invalid_artifact/u);
    results.pop(); first.publish();
    assert.ok(Object.isFrozen(canonicalComparison(selected.metadata.source)));
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("page capture preserves focus and scroll, crops exactly, and rejects post-screenshot geometry changes", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-capture-")), session = sessionAt(root);
  const server = createServer((_request, response) => { response.setHeader("content-type", "text/html"); response.end('<html lang="en"><title>Capture fixture</title><body style="margin:0"><button data-testid="box" style="position:absolute;left:10.25px;top:20.5px;width:31px;height:22px">Capture</button></body></html>'); });
  let admission, browser;
  try {
    await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
    admission = await acquireHostAdmission({ parent: inheritedHostLease(), claims: { browser_stack: 1 }, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))) });
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
    const produced = await execute(session, request, 1), { bundle } = await session.store.loadBundle(produced.bundle_id);
    assert.ok(performance.now() - start >= 100);
    assert.equal(bundle.source.kind, "live_unattested"); assert.deepEqual(bundle.observation.image_dimensions, { width: 32, height: 23 });
    assert.deepEqual(bundle.observation.coordinate_transform, { origin_x: 10, origin_y: 20, scale_x: 1, scale_y: 1 });
    assert.equal(await browser.page.getByTestId("box").evaluate((node) => document.activeElement === node), true);
    const screenshot = browser.page.screenshot.bind(browser.page);
    browser.page.screenshot = async (options) => { const bytes = await screenshot(options); await browser.page.getByTestId("box").evaluate((node) => { node.style.left = "30px"; }); return bytes; };
    await assert.rejects(execute(session, request, 2), /unstable_capture/u); assert.equal(session.store.usage.bundles, 1);
    browser.page.screenshot = screenshot;
    await browser.page.getByTestId("box").evaluate((node) => {
      let count = 0; const move = () => { node.style.left = `${20 + count++ * 0.25}px`; requestAnimationFrame(move); }; requestAnimationFrame(move);
    });
    await assert.rejects(execute(session, request, 3), /operation_expired/u); assert.equal(session.store.usage.bundles, 1);
  } finally { await browser?.close(); await admission?.release(); await new Promise((resolve) => server.close(resolve)); rmSync(root, { recursive: true, force: true }); }
});
