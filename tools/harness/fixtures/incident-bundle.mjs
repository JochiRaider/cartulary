import { createHash, randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fixtureTar } from "./ustar.mjs";

// Test-only current-format empty-workbook source, using the authored machine inventory.
// Actual admission, validation, publication and membership run on the server.
export function incidentBundleFixture(status = "active", root = path.resolve(import.meta.dirname, "../../..")) {
  if (!["active", "closed"].includes(status)) throw new Error("invalid fixture status");
  const incidentId = randomUUID();
  const actorId = randomUUID();
  const key = `IMPORT-${incidentId}`;
  const timestamp = "2026-01-02T03:04:05+00:00";
  const catalog = JSON.parse(readFileSync(path.join(root, "contracts/incident-bundles/source_catalog.json"), "utf8"));
  const sources = new Map();
  for (const source of [
    ...catalog.families.flatMap((family) => family.paths),
    ...catalog.special_consumers,
  ]) {
    if (source.versions.includes(5) && !source.logical_path.includes("*"))
      sources.set(source.logical_path, Buffer.alloc(0));
  }
  const json = (value) => Buffer.from(`${JSON.stringify(value)}\n`);
  sources.set(
    "data/incident.json",
    json({
      id: incidentId,
      incident_key: key,
      incident_key_canonical: key,
      title: `Imported ${status} investigation`,
      description: null,
      status,
      severity: null,
      tlp: null,
      current_phase: null,
      primary_external_case_ref: null,
      created_by_user_id: actorId,
      created_at: timestamp,
      updated_at: timestamp,
      updated_by_user_id: actorId,
      incident_version: 1,
      closed_at: status === "closed" ? timestamp : null,
    }),
  );
  sources.set(
    "data/actors.ndjson",
    json({ actor_id: actorId, display_name: "Portable operator" }),
  );
  sources.set(
    "data/reference_pack_refs.json",
    Buffer.from(
      JSON.stringify({
        schema_id: "reference_pack_refs.v1",
        sets: [],
        versions: [],
      }),
    ),
  );
  sources.set(
    "data/timeline_time_profiles.ndjson",
    json({
      incident_id: incidentId,
      enabled: false,
      local_offset_minutes: null,
      local_label: null,
      profile_version: 1,
      updated_at: timestamp,
      updated_by_user_id: actorId,
    }),
  );
  const hash = (bytes) =>
    createHash("sha256").update(bytes).digest("hex");
  const files = [...sources]
    .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0))
    .map(([path, bytes]) => ({
      path,
      sha256: `sha256:${hash(bytes)}`,
      size_bytes: bytes.length,
      required: true,
    }));
  sources.set(
    "manifest.json",
    json({
      bundle_format: "cartulary.incident_bundle",
      bundle_version: 5,
      bundle_id: randomUUID(),
      incident_id: incidentId,
      incident_key: key,
      exported_at: timestamp,
      source_change_set_high_watermark: `cartulary.source_boundary.v1:${hash(Buffer.from(JSON.stringify(files)))}`,
      history_mode: "full",
      blob_mode: "full",
      reference_pack_mode: "refs_only",
      optional_sections: [],
      required_capabilities: [],
      files,
    }),
  );
  sources.set(
    "integrity/checksums.sha256",
    Buffer.from(
      files.map((file) => `${file.sha256.slice(7)}  ${file.path}\n`).join(""),
    ),
  );
  return {
    incidentId,
    key,
    upload: {
      name: `incident-${status}.tar`,
      mimeType: "application/x-tar",
      buffer: fixtureTar(sources),
    },
  };
}
