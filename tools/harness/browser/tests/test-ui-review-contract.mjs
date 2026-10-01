import "./test-renderer-identity.mjs";
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { commandID, emptyCounts, parseRequest, result, schemaID, validate } from "../ui-review/contract.mjs";
import { emitResult } from "../ui-review/output.mjs";
import { resolveInputs } from "../ui-review/inputs.mjs";
import { controlProfile, toolProfile } from "../ui-review/toolchain.mjs";
import { repoRoot } from "../ui-review/policy.mjs";
import { reviewPins } from "../ui-review/policy.mjs";
import { workClaims } from "../ui-review/work-policy.mjs";
import { sourceVariants, validateObservationChannels } from "../ui-review/bundle-semantics.mjs";
import { importFixture } from "./ui-review-import-fixture.mjs";
import { importCanonical } from "../ui-review/canonical-import.mjs";

test("review work reserves explicit typed CPU, memory and process envelopes", () => {
  assert.deepEqual(workClaims.raster, { cpu: 1, io: 1, memory_mb: 1024, process: 1 });
  assert.deepEqual(workClaims.report, { cpu: 1, io: 1, memory_mb: 512, process: 1 });
  assert.equal(reviewPins().playwright, reviewPins()["playwright-core"]);
  assert.equal(reviewPins()["axe-core"], reviewPins()["@axe-core/playwright"]);
  assert.ok(Object.isFrozen(workClaims.raster));
});

const action = (action, parameters = {}) => ({ schema_id: schemaID("action"), expected_epoch: 0, action, parameters });
const parse = (value, type = "action") => parseRequest(Buffer.from(JSON.stringify(value)), type);
test("review requests reject malformed encodings and unsupported operations before execution", () => {
  assert.equal(parse(action("snapshot")).action, "snapshot");
  assert.equal(parse(action("zoom", { percent: 125 })).parameters.percent, 125);
  assert.throws(() => parse(action("zoom", { percent: 124 })), /invalid_request/u);
  assert.throws(() => parse({ ...action("snapshot"), schema_id: "cartulary.ui_review_action.v1" }), /invalid_request/u);
  const request = Buffer.from(JSON.stringify(action("snapshot")));
  assert.equal(parseRequest(Buffer.concat([request, Buffer.alloc(65536 - request.length, 32)]), "action").action, "snapshot");
  for (const value of [action("eval", { code: "document.cookie" }), { ...action("snapshot"), extra: true }, action("click", { target: { kind: "css", value: "body" } }), action("resize", { width: 3840, height: 3840 }), action("navigate", { path: "//example.com" })]) assert.throws(() => parse(value), /invalid_request/u);
  for (const bytes of [Buffer.from('{"schema_id":1,"schema_id":2}'), Buffer.from([0xff]), Buffer.from("{}"), Buffer.alloc(65537, 32)]) assert.throws(() => parseRequest(bytes, "action"), /invalid_request/u);
  assert.throws(() => parse(action("fill", { target: { kind: "test_id", value: "field" }, text: "🙂".repeat(4097) })), /invalid_request/u);
  assert.throws(() => parse(action("click", { target: { kind: "test_id", value: "🙂".repeat(257) } })), /invalid_request/u);
  assert.equal(parse(action("fill", { target: { kind: "test_id", value: "field" }, text: "" })).parameters.text, "");
});
test("review structural variants keep unavailable observations and terminal evidence distinct", () => {
  const session_id = `uireview-${"a".repeat(32)}`;
  const timestamp = "2026-09-27T00:00:00.000Z";
  const locator = { schema_id: schemaID("session"), session_id, run_id: "review-fixture", mode: "artifacts", state: "ready", created_at: timestamp, updated_at: timestamp, terminal_receipt: null };
  validate("session", locator);
  const receipt = { schema_id: schemaID("receipt"), command_id: commandID("ui-review"), session_id, operation_id: null, mode: "artifacts", state: "closed", status: "ok", exit_code: 0, started_at: timestamp, finished_at: timestamp, duration_ms: 0, failures: [], counts: emptyCounts(), bundle_id: null, cleanup: "complete" };
  validate("receipt", receipt);
  assert.throws(() => validate("receipt", { ...receipt, state: "ready" }));
  const observations = { schema_id: schemaID("observations"), elements: [], fonts: [], accessibility_snapshot: null, axe: { status: "unavailable", engine_version: null, scope: "main_document", violations: [], incomplete: [], unassessed_frames: null }, console: { records: [], truncated: false }, network: { records: [], truncated: false } };
  validate("observations", observations);
  const bundle = { schema_id: schemaID("bundle"), session_id, bundle_id: "bundle-1", classification: "private_diagnostic", tool_profile: toolProfile(), parents: [], source: { kind: "reference_image", import_ref: { input_path: "/private/image.png", input_sha256: "b".repeat(64), metadata: null } }, observation: null, binding: null, components: { original: null, expected: null, actual: null, diff: null, observations: null, trace: null }, derived: [], analysis: null, limitations: ["no_axe", "no_dom", "no_trace", "reference_only"] };
  bundle.components.original = { path: "original.png", sha256: "b".repeat(64), bytes: 100, media_type: "image/png" };
  validate("bundle", bundle);
  for (const source of [{ ...bundle.source, workspace_digest: "a".repeat(64) }, { ...bundle.source, kind: "canonical_visual" }]) assert.throws(() => validate("bundle", { ...bundle, source }));
  assert.throws(() => validate("bundle", { ...bundle, limitations: ["no_dom", "no_axe"] }));
  assert.throws(() => validate("command_result", { ...result("ui-browser"), status: "error" }));
});
test("every source variant rejects impossible channels, limitations and parent identities", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-review-variants-"));
  try {
  const canonical = await importCanonical(importFixture(root).request);
  const sha = "a".repeat(64), time = "2026-09-28T00:00:00.000Z";
  const ref = (name, media_type = "image/png") => ({ path: name, sha256: sha, bytes: 1, media_type });
  const binding = { owner_id: "fixture.owner", row_id: "fixture.row", scenario_id: "fixture.scenario", capture_id: "fixture.capture", fixture_ids: [] };
  const observed = { epoch: 0, started_at: time, finished_at: time, duration_ms: 0, viewport: { width: 1, height: 1 }, image_dimensions: { width: 1, height: 1 }, device_scale_factor: 1, visual_viewport_scale: 1, css_zoom: 1, theme: null, density: null, document_generation: 0, scope: { kind: "viewport" }, coordinate_transform: { origin_x: 0, origin_y: 0, scale_x: 1, scale_y: 1 } };
  for (const [kind, rule] of Object.entries(sourceVariants)) {
    const bundle = { schema_id: schemaID("bundle"), session_id: `uireview-${"a".repeat(32)}`, bundle_id: "bundle-1", classification: "private_diagnostic", tool_profile: toolProfile(), parents: [], source: { kind }, observation: rule.observation ? observed : null, binding: kind === "canonical_visual" ? binding : null, components: { original: null, expected: null, actual: null, diff: null, observations: null, trace: null }, derived: [], analysis: null, limitations: [] };
    if (kind === "reference_image") Object.assign(bundle.source, { import_ref: { input_path: "/private/reference.png", input_sha256: sha, metadata: null } });
    else Object.assign(bundle.source, { workspace_digest: sha, browser_version: "fixture-engine" });
    if (kind === "sealed_review") Object.assign(bundle.source, { served_source_digest: sha, frontend_receipt: ref("frontend.json", "application/json"), runtime_profile_id: "default" });
    for (const name of rule.required) bundle.components[name] = ref(`${name}.${name === "observations" ? "json" : "png"}`, name === "observations" ? "application/json" : "image/png");
    if (kind === "canonical_visual") Object.assign(bundle, canonical.metadata);
    bundle.limitations = ["no_trace", ...(rule.observation ? ["rendered_nodes_only", "no_axe"] : ["no_dom", "no_axe"]), ...(kind === "reference_image" ? ["reference_only"] : []), ...(kind === "live_unattested" ? ["live_unattested"] : []), ...(kind === "canonical_visual" ? ["no_actual"] : [])].sort();
    validate("bundle", bundle);
    for (const name of rule.required) assert.throws(() => validate("bundle", { ...bundle, components: { ...bundle.components, [name]: null } }), /invalid_artifact/u);
    for (const name of rule.forbidden) assert.throws(() => validate("bundle", { ...bundle, components: { ...bundle.components, [name]: ref(`${name}.png`) } }), /invalid_artifact/u);
    for (const token of bundle.limitations.filter((value) => value !== "no_axe" || !rule.observation)) assert.throws(() => validate("bundle", { ...bundle, limitations: bundle.limitations.filter((value) => value !== token) }), /invalid_artifact/u);
    assert.throws(() => validate("bundle", { ...bundle, parents: [{ session_id: bundle.session_id, bundle_id: bundle.bundle_id, sha256: sha }] }), /invalid_artifact/u);
    if (rule.observation) {
      const channels = { axe: { status: "disabled" }, console: { truncated: false }, network: { truncated: false } };
      validateObservationChannels(bundle, channels);
      for (const channel of ["console", "network"]) assert.throws(() => validateObservationChannels(bundle, { ...channels, [channel]: { truncated: true } }));
      assert.throws(() => validateObservationChannels(bundle, { ...channels, axe: { status: "completed" } }));
    }
  }
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("review analysis and capture omission rules remain explicit", () => {
  const capture = parse({ schema_id: schemaID("capture_request"), source: "page", expected_epoch: 0 }, "capture_request");
  assert.deepEqual(capture.scope, { kind: "viewport" }); assert.equal(capture.include_axe, true);
  const analysis = parse({ schema_id: schemaID("analysis_request"), bundle_id: "bundle-1" }, "analysis_request");
  assert.deepEqual(analysis.operations, ["contact_sheet"]);
  assert.throws(() => parse({ ...analysis, operations: ["exact_diff"] }, "analysis_request"));
  assert.throws(() => parse({ ...analysis, crops: [{ x: 0, y: 0, width: 1, height: 1 }] }, "analysis_request"));
});
test("review input modes ignore inherited UI inputs and reject cross-mode combinations", () => {
  assert.equal(resolveInputs("ui-review", { UI_MODE: "dev", CARTULARY_MAKE_INPUT_SOURCES: "UI_MODE=env" }).UI_MODE, "seeded");
  for (const env of [
    { UI_MODE: "dev", CARTULARY_MAKE_INPUT_SOURCES: "UI_MODE=cli" },
    { UI_MODE: "artifacts", REVIEW_PROFILE: "default", CARTULARY_MAKE_INPUT_SOURCES: "UI_MODE=cli REVIEW_PROFILE=cli" },
    { UI_MODE: "dev", UI_ORIGIN: "http://localhost:3000", CARTULARY_MAKE_INPUT_SOURCES: "UI_MODE=cli UI_ORIGIN=cli" },
    { UI_MODE: "", CARTULARY_MAKE_INPUT_SOURCES: "UI_MODE=cli" },
    { CARTULARY_MAKE_INPUT_SOURCES: "UNDECLARED=cli" },
    { CARTULARY_OUTPUT_MODE: "machine" },
  ]) assert.throws(() => resolveInputs("ui-review", env), /invalid_request/u);
  const dev = resolveInputs("ui-review", { UI_MODE: "dev", UI_ORIGIN: "http://[::1]:9000/", CARTULARY_MAKE_INPUT_SOURCES: "UI_MODE=cli UI_ORIGIN=cli" });
  assert.equal(dev.UI_ORIGIN, "http://[::1]:9000");
});
test("review output is one transient object and cannot add arbitrary fields", () => {
  const value = result("ui-capture"); let output = "";
  emitResult(value, "machine", { write: (part) => { output += part; } });
  assert.equal(output, `${JSON.stringify(value)}\n`);
  assert.throws(() => validate("command_result", { ...value, page_text: "private" }));
  assert.throws(() => validate("receipt", { schema_id: schemaID("receipt"), private_refs: [] }));
});
test("review core readiness rejects missing packages without an ambient fallback", () => {
  assert.equal(toolProfile().playwright_version, reviewPins().playwright);
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-ui-pins-"));
  try {
    mkdirSync(path.join(root, "tools"));
    for (const name of ["package.json", "pnpm-lock.yaml", "tools/toolchain_pins.json"]) writeFileSync(path.join(root, name), readFileSync(path.join(repoRoot, name)));
    assert.throws(() => toolProfile(root), /tool_configuration/u);
    assert.equal(controlProfile(root).node_version, process.version.slice(1));
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("public Make review rejections contain no private input or retained output", () => {
  for (const args of [["ui-review", "UI_MODE=invalid-private-marker"], ["ui-capture", "UI_SESSION=private-path-marker", "UNDECLARED=1"]]) {
    const env = { ...process.env, CARTULARY_OUTPUT_MODE: "machine" };
    for (const name of Object.keys(env)) if (/^(?:MAKE|MFLAGS|CARTULARY_HARNESS_GRAPH|CARTULARY_MAKE_INPUT|CARTULARY_TEST_TARGET)/u.test(name)) delete env[name];
    const child = spawnSync("make", ["--silent", ...args], { cwd: repoRoot, env, encoding: "utf8" });
    assert.equal(child.status, 2);
    const response = JSON.parse(child.stdout);
    validate("command_result", response);
    assert.equal(response.failures[0].diagnostic_code, "invalid_request");
    assert.ok(!child.stdout.includes("private-marker") && !child.stdout.includes("private-path-marker"));
    assert.equal(response.receipt, null);
  }
});

test("diagnostic action is closed and superseded requests have no reader", () => {
  const action = { schema_id: schemaID("action"), expected_epoch: 0, action: "diagnostic_snapshot", parameters: {} };
  assert.deepEqual(parse(action, "action"), action);
  for (const schema_id of ["cartulary.ui_review_action.v1", "cartulary.ui_review_action.v2"]) assert.throws(() => parse({ ...action, schema_id }, "action"));
  for (const parameters of [{ endpoint: "ws://127.0.0.1:1" }, { code: "1+1" }, { args: [] }, { path: "/tmp/out" }]) assert.throws(() => parse({ ...action, parameters }, "action"));
});
