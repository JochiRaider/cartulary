import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { existsSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import { createServer } from "node:http";
import os from "node:os";
import path from "node:path";
import { repoRoot } from "../ui-review/toolchain.mjs";
import { schemaID, validate } from "../ui-review/contract.mjs";
import { digest } from "../ui-review/session-files.mjs";

const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const cleanEnvironment = () => {
  const env = { ...process.env };
  for (const key of Object.keys(env)) if (/^(?:MAKE|MFLAGS|CARTULARY_(?:HARNESS|MAKE|TEST|OUTPUT)|OWNER$|ROWS$|UI_|REVIEW_PROFILE$)/u.test(key)) delete env[key];
  return env;
};
function make(args, env) {
  const child = spawn("make", ["--silent", ...args], { cwd: repoRoot, env, stdio: ["ignore", "pipe", "pipe"] });
  let output = "", error = "";
  child.stdout.on("data", (part) => { output += part; }); child.stderr.on("data", (part) => { error += part; });
  return { child, output: () => output, ended: new Promise((resolve, reject) => { child.once("error", reject); child.once("close", (code) => resolve({ code, output, error })); }) };
}
export async function publicWorkflow({ seeded = false, resultsRoot, runID = "workflow" } = {}) {
  const privateRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-public-workflow-"));
  const env = cleanEnvironment(), privateRefs = [], manifests = [], sentinel = "review-private-sentinel-392884";
  let server, running, locator, requestIndex = 0;
  const invoke = async (command, value, extra = []) => {
    let inputs = [];
    if (value) { const file = path.join(privateRoot, `request-${++requestIndex}.json`); writeFileSync(file, JSON.stringify(value), { mode: 0o600 }); inputs = [`UI_REQUEST=${file}`]; }
    const response = await make([command, `UI_SESSION=${locator}`, ...inputs, ...extra], { ...env, CARTULARY_OUTPUT_MODE: "machine" }).ended;
    const result = validate("command_result", JSON.parse(response.output));
    assert.equal(result.status, "ok", `${command}: ${JSON.stringify(result.failures)}`); assert.equal(response.code, 0);
    privateRefs.push(...result.private_refs.map((ref) => ref.absolute_path));
    return result;
  };
  const start = async (mode, extra = [], id = runID) => {
    running = make(["ui-review", ...(mode === "seeded" ? [] : [`UI_MODE=${mode}`]), ...extra], { ...env, CARTULARY_TEST_RESULTS_DIR: resultsRoot ?? privateRoot, CARTULARY_TEST_RUN_ID: id, ...(mode === "artifacts" ? { PLAYWRIGHT_BROWSERS_PATH: path.join(privateRoot, "absent-browsers"), DOCKER_HOST: "unix:///no-docker", PATH: "/usr/bin:/bin" } : {}), OTEL_EXPORTER_OTLP_ENDPOINT: "http://127.0.0.1:1/private-telemetry" });
    locator = path.join(resultsRoot ?? privateRoot, id, "ui-review/session.json");
    for (let attempt = 0; attempt < 3000 && !running.output().includes("UI review ready"); attempt++) {
      if (running.child.exitCode !== null) break; await pause(100);
    }
    assert.ok(running.output().includes("UI review ready"), `review startup failed: ${running.output()}`);
  };
  const stop = async () => {
    const result = await invoke("ui-review-stop"), ended = await running.ended;
    assert.equal(ended.code, 0); assert.deepEqual((await invoke("ui-review-stop")).receipt, result.receipt); assert.equal((await invoke("ui-review-status")).state, "closed");
    running = null;
    return { run_root: path.dirname(path.dirname(locator)), terminal_sha256: result.receipt.sha256 };
  };
  try {
    if (!seeded) {
      server = createServer((_request, response) => { response.setHeader("content-type", "text/html"); response.end(`<!doctype html><html lang="en"><title>Private review</title><body><main><h1>${sentinel}</h1><input aria-label="Private field" data-testid="field"><button data-testid="apply" onclick="document.querySelector('h1').textContent=document.querySelector('input').value">Apply</button><script>console.error('${sentinel}')</script></main></body></html>`); });
      await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
    }
    await start(seeded ? "seeded" : "dev", seeded ? ["REVIEW_PROFILE=default"] : [`UI_ORIGIN=http://127.0.0.1:${server.address().port}`]);
    let epoch = (await invoke("ui-review-status")).epoch;
    const action = async (name, parameters = {}) => { const result = await invoke("ui-browser", { schema_id: schemaID("action"), expected_epoch: epoch, action: name, parameters }); epoch = result.epoch; return result; };
    const review = async () => {
      await action("snapshot");
      const captured = await invoke("ui-capture", { schema_id: schemaID("capture_request"), source: "page", expected_epoch: epoch });
      const file = captured.private_refs[0].absolute_path, bytes = readFileSync(file), bundle = JSON.parse(bytes); manifests.push([file, digest(bytes)]);
      assert.equal(bundle.source.kind, seeded ? "sealed_review" : "live_unattested");
      const analyzed = await invoke("ui-analyze", { schema_id: schemaID("analysis_request"), bundle_id: captured.bundle_id, operations: ["contact_sheet", "crop"], crops: [{ x: 0, y: 0, width: 100, height: 100 }] });
      const rendered = await invoke("ui-review-report", null, [`UI_BUNDLE=${analyzed.bundle_id}`]);
      assert.match(readFileSync(rendered.private_refs[0].absolute_path, "utf8"), /Private UI review/u);
      return { image: path.join(path.dirname(file), bundle.components.original.path), manifest: file };
    };
    let capture;
    if (seeded) for (const actor of ["editor", "viewer"]) { await action("authenticate", { actor }); await action("navigate", { path: "/" }); capture = await review(); }
    else {
      await action("navigate", { path: `/?private=${sentinel}` });
      await action("fill", { target: { kind: "test_id", value: "field" }, text: sentinel });
      await action("click", { target: { kind: "test_id", value: "apply" } }); capture = await review();
    }
    const reference = path.join(privateRoot, "reference.png"), original = readFileSync(capture.image); writeFileSync(reference, original, { mode: 0o600 });
    for (const [file, sha] of manifests) assert.equal(digest(readFileSync(file)), sha);
    const browserResult = await stop();
    assert.ok(privateRefs.every((file) => !existsSync(file)));
    if (server) assert.equal((await fetch(`http://127.0.0.1:${server.address().port}`)).status, 200);
    await start("artifacts", [], `${runID}-artifacts`);
    const imported = [];
    for (let index = 0; index < 2; index++) imported.push(await invoke("ui-capture", { schema_id: schemaID("capture_request"), source: "image", path: reference }));
    const analyzed = await invoke("ui-analyze", { schema_id: schemaID("analysis_request"), bundle_id: imported[0].bundle_id, operations: ["exact_diff", "contact_sheet"], comparison: { kind: "reference", bundle_id: imported[1].bundle_id } });
    const bundle = JSON.parse(readFileSync(analyzed.private_refs[0].absolute_path)); assert.equal(bundle.analysis.comparison.different_pixels, 0);
    await invoke("ui-review-report", null, [`UI_BUNDLE=${analyzed.bundle_id}`]);
    const artifactResult = await stop(); assert.deepEqual(readFileSync(reference), original);
    assert.ok(privateRefs.every((file) => !existsSync(file)));
    const inspect = (root) => { for (const entry of readdirSync(root, { withFileTypes: true })) { const file = path.join(root, entry.name); if (entry.isDirectory()) inspect(file); else { assert.ok(["session.json", "receipt.json", "terminal.json"].includes(entry.name)); const text = readFileSync(file, "utf8"); for (const secret of [sentinel, privateRoot, "private-telemetry", "private_refs"]) assert.ok(!text.includes(secret)); } } };
    inspect(browserResult.run_root); inspect(artifactResult.run_root);
    return { browser: browserResult, artifacts: artifactResult, private_cleanup: "complete", retained_projection: "structural_only" };
  } finally {
    if (running) { try { await invoke("ui-review-stop"); } catch { running.child.kill("SIGTERM"); } await running.ended; }
    if (server) await new Promise((resolve) => server.close(resolve));
    rmSync(privateRoot, { recursive: true, force: true });
  }
}
