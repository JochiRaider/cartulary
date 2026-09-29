import path from "node:path";
import { freeze } from "./immutable.mjs";
import { reviewPins } from "./policy.mjs";
import { validateSchemaSync, parseStrictJSON } from "../../contract/index.mjs";
import { artifact, components } from "./bundles.mjs";
import { digest, inputPath, readInput } from "./session-files.mjs";
import { catalogRow, containedFile, one, readJSON } from "./source.mjs";
import { repoRoot } from "./policy.mjs";
import { ReviewFailure, limits } from "./contract.mjs";

function exactAttachmentFile(runRoot, file) {
  const absolute = inputPath(path.isAbsolute(file) ? file : path.resolve(repoRoot, file));
  if (path.relative(runRoot, absolute).startsWith("..") || absolute === runRoot) throw new ReviewFailure("invalid_artifact");
  return absolute;
}
function attachmentBytes(runRoot, attachment, extension, maximum) {
  if (attachment.path !== undefined && attachment.body !== undefined) throw new ReviewFailure("invalid_artifact");
  if (typeof attachment.path === "string") return readInput(exactAttachmentFile(runRoot, attachment.path), { extension, maximum, privateFile: false });
  if (typeof attachment.body !== "string" || attachment.body.length > Math.ceil(maximum / 3) * 4 || !/^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/u.test(attachment.body)) throw new ReviewFailure("invalid_artifact");
  const bytes = Buffer.from(attachment.body, "base64");
  if (bytes.length > maximum) throw new ReviewFailure("observation_limit");
  return bytes;
}
function specs(suites, depth = 0) {
  if (!Array.isArray(suites) || depth > 32 || suites.length > 10000) throw new ReviewFailure("invalid_artifact");
  return suites.flatMap((suite) => [...(suite.specs ?? []), ...specs(suite.suites ?? [], depth + 1)]);
}
export function selectCaptureResult(report, capture, runRoot) {
  if (report.config?.version !== reviewPins().playwright) throw new ReviewFailure("invalid_artifact");
  const matches = [];
  for (const spec of specs(report.suites)) for (const test of spec.tests ?? []) for (const result of test.results ?? []) {
    if (test.projectName !== capture.project_id || spec.title !== capture.test_title || path.resolve(report.config.rootDir, spec.file) !== path.resolve(repoRoot, capture.assertion_file)) continue;
    const attachments = result.attachments;
    if (!Array.isArray(attachments) || attachments.length > 10000) throw new ReviewFailure("invalid_artifact");
    for (const attachment of attachments.filter((entry) => entry.name === `cartulary-visual-capture-intent-${capture.capture_id}.json`)) {
      if (attachment.contentType !== "application/json") throw new ReviewFailure("invalid_artifact");
      const payload = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(attachmentBytes(runRoot, attachment, ".json", limits.component)));
      validateSchemaSync("cartulary.frontend_visual_capture_intent.v2", payload);
      const mapping = { test_file: "assertion_file", test_title: "test_title", capture_id: "capture_id", capture_intent: "capture_intent", expected_golden_path: "expected_golden_path", project_id: "project_id", renderer_profile_id: "renderer_profile_id", screenshot_assertion_location: "screenshot_assertion_location", capture_profile: "capture_profile" };
      for (const [field, source] of Object.entries(mapping)) if (JSON.stringify(payload[field]) !== JSON.stringify(capture[source])) throw new ReviewFailure("invalid_artifact");
      matches.push(result);
    }
  }
  return one(matches);
}
export async function importCanonical(request) {
  try {
    const root = inputPath(request.run_root);
    const manifest = readJSON(containedFile(root, "run-manifest.json"), "cartulary.harness_run_manifest.v1").value;
    if (manifest.run_id !== path.basename(root)) throw new ReviewFailure("invalid_artifact");
    const target = readJSON(containedFile(root, "browser-e2e-visual/browser-target-result.json"), "cartulary.browser_target_result.v4").value;
    if (target.target_id !== "browser-e2e-visual") throw new ReviewFailure("invalid_artifact");
    const reference = one((target.artifacts ?? []).filter((entry) => entry.kind === "frontend_visual_reconciliation_v3"));
    const reconciliation = readJSON(containedFile(root, reference.ref), "cartulary.frontend_visual_reconciliation.v3", reference.sha256);
    const capture = one(reconciliation.value.capture_intents.filter((entry) => entry.capture_id === request.capture_id));
    if (capture.capture_id !== `visual.capture.${digest(Buffer.from(JSON.stringify([capture.project_id, capture.test_title, capture.expected_golden_path]))).slice(0, 20)}`) throw new ReviewFailure("invalid_artifact");
    if (capture.screenshot_name_source !== capture.capture_intent || !/^[A-Za-z0-9_-]{1,200}$/u.test(capture.capture_intent)) throw new ReviewFailure("invalid_artifact");
    const sourceRefs = reconciliation.value.source_refs;
    const readSource = (file, schema) => {
      const relative = path.relative(repoRoot, inputPath(file)).split(path.sep).join("/");
      const source = one(sourceRefs.filter((entry) => entry.path === relative));
      return readJSON(containedFile(repoRoot, source.path), schema, source.sha256);
    };
    const { row, title } = catalogRow(capture, readSource);
    if (title !== capture.test_title || row.selector.file !== capture.assertion_file || row.selector.project_id !== capture.project_id) throw new ReviewFailure("invalid_artifact");
    const renderer = readSource(path.join(repoRoot, "tools/frontend_visual_renderer_profile.json"), "cartulary.frontend_visual_renderer_profile.v1").value;
    for (const [key, value] of Object.entries(renderer)) if (key !== "schema_id" && reconciliation.value.renderer[key] !== value) throw new ReviewFailure("invalid_artifact");
    if (capture.renderer_profile_id !== renderer.profile_id) throw new ReviewFailure("invalid_artifact");
    if (reconciliation.value.renderer.attestation_count < 1) throw new ReviewFailure("invalid_artifact");
    const registry = readSource(path.join(repoRoot, "tools/frontend_visual_fixture_registry.json"), "cartulary.frontend_visual_fixture_registry.v6").value;
    const fixtures = registry.fixtures.filter((entry) => entry.golden_artifacts.includes(capture.expected_golden_path));
    if (fixtures.length > 1 || fixtures.some((entry) => !entry.catalog_row_ids.includes(capture.row_id))) throw new ReviewFailure("invalid_artifact");
    const fixture = fixtures[0] ?? null;
    if (fixture && !fixture.capture_profiles[capture.expected_golden_path]) throw new ReviewFailure("invalid_artifact");
    if (fixture) for (const [key, value] of Object.entries(fixture.capture_profiles[capture.expected_golden_path] ?? {})) if (capture.capture_profile[key] !== value) throw new ReviewFailure("invalid_artifact");
    const groups = target.group_results.map((ref) => {
      const group = readJSON(containedFile(root, ref.ref), "cartulary.browser_group_result.v6", ref.sha256).value;
      if (group.target_id !== target.target_id || group.group_id !== ref.group_id || group.browser_session_id !== ref.browser_session_id || group.stage_id !== "visual") throw new ReviewFailure("invalid_artifact");
      return group;
    });
    const group = one(groups.filter((entry) => entry.selected_rows.includes(capture.row_id)));
    if (group.runtime_profile_id !== row.runtime_profile_id || group.row_results.filter((entry) => entry.row_id === capture.row_id).length !== 1) throw new ReviewFailure("invalid_artifact");
    const session = one(target.sessions.filter((entry) => entry.browser_session_id === group.browser_session_id));
    if (session.runtime_profile_id !== group.runtime_profile_id || session.service_requirement !== group.service_requirement || JSON.stringify(session.artifacts) !== JSON.stringify(group.session_artifacts)) throw new ReviewFailure("invalid_artifact");
    const stackRef = one(group.session_artifacts.filter((entry) => entry.kind === "stack_v7"));
    const stack = readJSON(containedFile(root, stackRef.ref), "cartulary.web_e2e_stack.v7", stackRef.sha256).value;
    if (stack.browser_session_id !== group.browser_session_id || stack.runtime_profile_id !== group.runtime_profile_id) throw new ReviewFailure("invalid_artifact");
    const receipt = readJSON(containedFile(root, stack.frontend.build_artifact_ref), "cartulary.frontend_build_artifact.v1", stack.frontend.build_receipt_sha256).value;
    if (receipt.run_id !== manifest.run_id || receipt.source_digest !== manifest.source_digest || receipt.toolchain_digest !== manifest.toolchain_digest || receipt.content_digest !== stack.frontend.build_artifact_sha256) throw new ReviewFailure("invalid_artifact");
    const report = readJSON(containedFile(root, group.artifacts.playwright_report)).value;
    const result = selectCaptureResult(report, capture, root);
    const files = new Map(), refs = components();
    const put = (kind, bytes, media = "image/png") => { const name = `${kind}.${media === "application/zip" ? "zip" : "png"}`; files.set(name, bytes); refs[kind] = artifact(name, bytes, media); };
    const golden = one(reconciliation.value.goldens.filter((entry) => entry.golden_path === capture.expected_golden_path && entry.consumer_capture_ids.includes(capture.capture_id)));
    if (golden.sha256) {
      const expected = readInput(containedFile(repoRoot, capture.expected_golden_path), { extension: ".png", maximum: limits.png, privateFile: false });
      if (digest(expected) !== golden.sha256) throw new ReviewFailure("invalid_artifact"); put("expected", expected);
    }
    for (const kind of ["actual", "diff", "expected"]) {
      const matches = result.attachments.filter((entry) => entry.name === `${capture.capture_intent}-${kind}.png`);
      if (matches.length > 1) throw new ReviewFailure("invalid_artifact");
      if (matches.length) {
        const attachment = matches[0]; if (attachment.contentType !== "image/png") throw new ReviewFailure("invalid_artifact");
        let bytes;
        if (kind === "expected" && typeof attachment.path === "string") {
          const selected = inputPath(attachment.path, ".png");
          if (selected !== containedFile(repoRoot, capture.expected_golden_path) || attachment.body !== undefined) throw new ReviewFailure("invalid_artifact");
          bytes = readInput(selected, { extension: ".png", maximum: limits.png, privateFile: false });
        } else bytes = attachmentBytes(root, attachment, ".png", limits.png);
        if (kind === "expected") { if (!golden.sha256 || digest(bytes) !== golden.sha256) throw new ReviewFailure("invalid_artifact"); }
        else put(kind, bytes);
      }
    }
    const traces = result.attachments.filter((entry) => entry.name === "trace");
    if (traces.length > 1) throw new ReviewFailure("invalid_artifact");
    if (traces.length) {
      if (traces[0].contentType !== "application/zip") throw new ReviewFailure("invalid_artifact");
      const bytes = attachmentBytes(root, traces[0], ".zip", limits.component);
      if (bytes.length < 4 || bytes.readUInt32LE() !== 0x04034b50) throw new ReviewFailure("invalid_artifact");
      put("trace", bytes, "application/zip");
    }
    if (!refs.expected && !refs.actual) throw new ReviewFailure("invalid_artifact");
    return { files, metadata: freeze({
      source: { kind: "canonical_visual", workspace_digest: manifest.source_digest.slice(7), served_source_digest: receipt.source_digest.slice(7), renderer_profile_id: renderer.profile_id, browser_version: renderer.chromium_version, runtime_profile_id: group.runtime_profile_id, import_ref: { input_path: reconciliation.path, input_sha256: digest(reconciliation.bytes), metadata: { reconciliation: reconciliation.value, capture_intent: capture, source_identity: manifest, fixture } } },
      binding: { owner_id: capture.owner_id, row_id: capture.row_id, scenario_id: capture.scenario_id, capture_id: capture.capture_id, fixture_ids: fixtures.map((entry) => entry.fixture_id).sort() },
      components: refs, limitations: ["no_axe", "no_dom", ...(!refs.actual ? ["no_actual"] : []), ...(!refs.trace ? ["no_trace"] : [])].sort(),
    }) };
  } catch (cause) { if (cause instanceof ReviewFailure) throw cause; throw new ReviewFailure("invalid_artifact", { cause }); }
}

/** Producer-specific joins end here. Missing scope is not comparable evidence;
 * it does not prevent an expected-only diagnostic import or report.
 */
export function canonicalComparison(source) {
  if (source.kind !== "canonical_visual") return null;
  const metadata = source.import_ref?.metadata, fixture = metadata?.fixture;
  if (!metadata || !fixture?.capture_scope || !Array.isArray(fixture.dynamic_masks) || typeof fixture.no_dynamic_regions !== "boolean" || (fixture.no_dynamic_regions ? fixture.dynamic_masks.length !== 0 : fixture.dynamic_masks.length === 0)) return null;
  return freeze({ renderer: source.renderer_profile_id, source_kind: source.kind, runtime: source.runtime_profile_id, profile: structuredClone(metadata.capture_intent.capture_profile), capture_id: metadata.capture_intent.capture_id, scenario: metadata.capture_intent.scenario_id, scope: structuredClone(fixture.capture_scope), masks: [...fixture.dynamic_masks].sort(), no_dynamic_regions: fixture.no_dynamic_regions });
}
