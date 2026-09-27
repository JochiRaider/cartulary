import path from "node:path";
import { parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
import { digest, inputPath, readInput } from "./session-files.mjs";
import { limits, ReviewFailure } from "./contract.mjs";
import { repoRoot } from "./toolchain.mjs";

export function one(values) { if (values.length !== 1) throw new ReviewFailure("invalid_artifact"); return values[0]; }
export function readJSON(file, schema, expectedDigest) {
  const bytes = readInput(file, { extension: ".json", maximum: limits.component, privateFile: false });
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
export function pageBinding(session, binding) {
  if (binding === null) return null;
  if (session.mode !== "seeded") throw new ReviewFailure("invalid_request");
  const { row, title } = catalogRow(binding);
  if (row.runtime_profile_id !== session.input.REVIEW_PROFILE) throw new ReviewFailure("invalid_artifact");
  const goldens = readJSON(path.join(repoRoot, "tools/frontend_visual_golden_manifest.json"), "cartulary.frontend_visual_golden_manifest.v1").value;
  const golden = one(goldens.goldens.filter((entry) => `visual.capture.${digest(Buffer.from(JSON.stringify([row.selector.project_id, title, entry.path]))).slice(0, 20)}` === binding.capture_id));
  const registry = readJSON(path.join(repoRoot, "tools/frontend_visual_fixture_registry.json"), "cartulary.frontend_visual_fixture_registry.v6").value;
  const fixtures = registry.fixtures.filter((entry) => entry.golden_artifacts.includes(golden.path));
  if (fixtures.length > 1 || fixtures.some((fixture) => !fixture.catalog_row_ids.includes(binding.row_id))) throw new ReviewFailure("invalid_artifact");
  if (fixtures.length && !fixtures[0].capture_profiles[golden.path]) throw new ReviewFailure("invalid_artifact");
  return { binding: { ...binding, fixture_ids: fixtures.map((fixture) => fixture.fixture_id).sort() }, profile: fixtures[0]?.capture_profiles[golden.path] ?? null };
}
