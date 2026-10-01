import path from "node:path";
import { freeze } from "./immutable.mjs";
import { artifact } from "./bundles.mjs";
import { parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
import { digest, inputPath, readInput } from "./session-files.mjs";
import { limits, ReviewFailure } from "./contract.mjs";
import { repoRoot } from "./policy.mjs";

export function one(values) { if (values.length !== 1) throw new ReviewFailure("invalid_artifact"); return values[0]; }
export function readJSON(file, schema, expectedDigest, maximum = limits.component) {
  const bytes = readInput(file, { extension: ".json", maximum, privateFile: false });
  if (expectedDigest && digest(bytes) !== expectedDigest.replace(/^sha256:/u, "")) throw new ReviewFailure("invalid_artifact");
  try {
    const value = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes));
    if (schema) validateSchemaSync(schema, value);
    return { value, bytes, path: inputPath(file, ".json") };
  } catch (cause) { throw new ReviewFailure("invalid_artifact", { cause }); }
}
export function containedFile(root, ref) {
  if (typeof ref !== "string" || !ref || path.isAbsolute(ref) || ref.split("/").some((part) => !part || part === "." || part === "..")) throw new ReviewFailure("invalid_artifact");
  const file = inputPath(path.join(root, ref));
  if (path.relative(root, file).startsWith("..")) throw new ReviewFailure("invalid_artifact");
  return file;
}
export function catalogRow(binding, read = readJSON) {
  const catalog = read(path.join(repoRoot, "tools/test_catalog_owner.json"), "cartulary.test_owner_registry.v1").value;
  const owner = one(catalog.owners.filter((entry) => entry.owner_id === binding.owner_id && entry.status === "active"));
  const family = read(containedFile(repoRoot, owner.manifest_path), "cartulary.test_family_manifest.v6").value;
  const row = one(family.rows.filter((entry) => entry.row_id === binding.row_id && entry.owner_id === binding.owner_id && entry.status === "active" && entry.runner === "playwright" && entry.selector.stage === "visual"));
  const index = row.selector.scenario_ids.indexOf(binding.scenario_id);
  if (index < 0) throw new ReviewFailure("invalid_artifact");
  return { row, title: row.selector.titles[index] };
}
function pageBinding(mode, profile, binding) {
  if (binding === null) return null;
  if (mode !== "seeded") throw new ReviewFailure("invalid_request");
  const { row, title } = catalogRow(binding);
  if (row.runtime_profile_id !== profile) throw new ReviewFailure("invalid_artifact");
  const goldens = readJSON(path.join(repoRoot, "tools/frontend_visual_golden_manifest.json"), "cartulary.frontend_visual_golden_manifest.v1").value;
  const golden = one(goldens.goldens.filter((entry) => `visual.capture.${digest(Buffer.from(JSON.stringify([row.selector.project_id, title, entry.path]))).slice(0, 20)}` === binding.capture_id));
  const registry = readJSON(path.join(repoRoot, "tools/frontend_visual_fixture_registry.json"), "cartulary.frontend_visual_fixture_registry.v6").value;
  const fixtures = registry.fixtures.filter((entry) => entry.golden_artifacts.includes(golden.path));
  if (fixtures.length > 1 || fixtures.some((fixture) => !fixture.catalog_row_ids.includes(binding.row_id))) throw new ReviewFailure("invalid_artifact");
  if (fixtures.length && !fixtures[0].capture_profiles[golden.path]) throw new ReviewFailure("invalid_artifact");
  return { binding: { ...binding, fixture_ids: fixtures.map((fixture) => fixture.fixture_id).sort() }, profile: fixtures[0]?.capture_profiles[golden.path] ?? null };
}

export function pageSource({ mode, profile, workspaceDigest, runID, prepared, browserVersion }, requestedBinding) {
  const binding = pageBinding(mode, profile, requestedBinding);
  const fonts = binding ? readJSON(path.join(repoRoot, "apps/web/public/assets/fonts/FONT_MANIFEST.json")).value.families.filter((font) => font.active_by_default).map((font) => font.family) : [];
  return freeze({ binding, fonts, attest: async () => {
    if (mode === "dev") return { source: freeze({ kind: "live_unattested", workspace_digest: workspaceDigest, browser_version: browserVersion }) };
    await prepared.check();
    const stack = readJSON(prepared.attached.CARTULARY_WEB_E2E_STACK_JSON_FILE, "cartulary.web_e2e_stack.v7").value;
    const receipt = readJSON(containedFile(prepared.runRoot, stack.frontend.build_artifact_ref), "cartulary.frontend_build_artifact.v1", stack.frontend.build_receipt_sha256);
    if (receipt.value.run_id !== runID || receipt.value.source_digest.replace(/^sha256:/u, "") !== workspaceDigest) throw new ReviewFailure("invalid_artifact");
    return { receipt: receipt.bytes, source: freeze({ kind: "sealed_review", workspace_digest: workspaceDigest, served_source_digest: receipt.value.source_digest.slice(7), frontend_receipt: artifact("frontend-receipt.json", receipt.bytes, "application/json"), browser_version: browserVersion, runtime_profile_id: profile }) };
  } });
}
