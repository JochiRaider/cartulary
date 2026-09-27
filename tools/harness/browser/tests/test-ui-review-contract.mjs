import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { commandID, emptyCounts, parseRequest, result, schemaID, validate } from "../ui-review/contract.mjs";
import { emitResult } from "../ui-review/output.mjs";
import { resolveInputs } from "../ui-review/inputs.mjs";
import { repoRoot, toolProfile } from "../ui-review/toolchain.mjs";

const action = (action, parameters = {}) => ({ schema_id: schemaID("action"), expected_epoch: 0, action, parameters });
const parse = (value, type = "action") => parseRequest(Buffer.from(JSON.stringify(value)), type);
test("review requests reject malformed encodings and unsupported operations before execution", () => {
  assert.equal(parse(action("snapshot")).action, "snapshot");
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
  validate("bundle", bundle);
  for (const source of [{ ...bundle.source, workspace_digest: "a".repeat(64) }, { ...bundle.source, kind: "canonical_visual" }]) assert.throws(() => validate("bundle", { ...bundle, source }));
  assert.throws(() => validate("bundle", { ...bundle, limitations: ["no_dom", "no_axe"] }));
  assert.throws(() => validate("command_result", { ...result("ui-browser"), status: "error" }));
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
  assert.equal(toolProfile().playwright_version, "1.59.1");
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-ui-pins-"));
  try {
    mkdirSync(path.join(root, "tools"));
    for (const name of ["package.json", "pnpm-lock.yaml", "tools/toolchain_pins.json"]) writeFileSync(path.join(root, name), readFileSync(path.join(repoRoot, name)));
    assert.throws(() => toolProfile(root), /tool_configuration/u);
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
