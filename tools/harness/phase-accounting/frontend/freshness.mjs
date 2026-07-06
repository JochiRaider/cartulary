import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

const frontendPhaseRegistrySchemaID = "cartulary.frontend_phase_registry.v5";
const frontendPhaseTestMapSchemaID = "cartulary.frontend_phase_test_map.v4";
const frontendRowAccountingSchemaID = "cartulary.frontend_row_accounting.v4";
const frontendVisualFixtureRegistrySchemaID =
  "cartulary.frontend_visual_fixture_registry.v3";

export function sha256File(root, relativePath) {
  const absolute = path.join(root, relativePath);
  if (!existsSync(absolute)) {
    return "";
  }
  return createHash("sha256").update(readFileSync(absolute)).digest("hex");
}

export function frontendEvidenceFreshnessDigest(root, registry, entry) {
  const payload = {
    schema_id: frontendPhaseRegistrySchemaID,
    map_schema_id: frontendPhaseTestMapSchemaID,
    row_accounting_schema_id: frontendRowAccountingSchemaID,
    phase_id: entry.phase_id,
    base_phase_join: entry.base_phase_join,
    guide_digest: registry.guide_digest,
    manifest_digest: entry.manifest_digest,
    ledger_digest: entry.ledger_digest,
    visual_fixture_registry_digest: sha256File(
      root,
      "tools/frontend_visual_fixture_registry.json",
    ),
    visual_fixture_registry_schema_digest: sha256File(
      root,
      `tools/schemas/${frontendVisualFixtureRegistrySchemaID}.schema.json`,
    ),
    registry_schema_digest: sha256File(
      root,
      `tools/schemas/${frontendPhaseRegistrySchemaID}.schema.json`,
    ),
    map_schema_digest: sha256File(
      root,
      `tools/schemas/${frontendPhaseTestMapSchemaID}.schema.json`,
    ),
    row_accounting_schema_digest: sha256File(
      root,
      `tools/schemas/${frontendRowAccountingSchemaID}.schema.json`,
    ),
  };
  return createHash("sha256").update(JSON.stringify(payload)).digest("hex");
}
