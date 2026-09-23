import { createHash, randomUUID } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

const hash = (bytes) => createHash("sha256").update(bytes).digest("hex");
const json = (value) => Buffer.from(`${JSON.stringify(value)}\n`);

export function reviewTar(entries) {
  const chunks = [];
  for (const [name, bytes] of entries) {
    if (Buffer.byteLength(name) >= 100) throw new Error("Review TAR path exceeds header capacity");
    const header = Buffer.alloc(512);
    header.write(name, 0, 100);
    const octal = (value, offset, width) => header.write(`${value.toString(8).padStart(width - 1, "0")}\0`, offset, width);
    octal(0o644, 100, 8); octal(0, 108, 8); octal(0, 116, 8);
    octal(bytes.length, 124, 12); octal(0, 136, 12);
    header.fill(32, 148, 156); header.write("0", 156);
    header.write("ustar\0", 257); header.write("00", 263);
    header.write(`${header.reduce((sum, byte) => sum + byte, 0).toString(8).padStart(6, "0")}\0 `, 148, 8);
    chunks.push(header, bytes, Buffer.alloc((512 - bytes.length % 512) % 512));
  }
  return Buffer.concat([...chunks, Buffer.alloc(1024)]);
}

export function writeReviewSamples(root, directory) {
  mkdirSync(directory, { recursive: true, mode: 0o700 });
  const write = (name, bytes) => writeFileSync(path.join(directory, name), bytes, { mode: 0o600 });
  write("timeline.csv", "Activity Synopsis,Unmapped source note\nImported browser review activity,Retained source text\nSecond imported activity,Review mapping and outcomes\n");
  write("evidence.txt", "Browser review evidence: synthetic acquisition notes.\n");
  for (const [source, name] of [
    ["apps/web/e2e/testdata/import-assistant-partial.xlsx", "workbook-partial.xlsx"],
    ["fixtures/network-flow/NF-FIX-001-cisco-sna-minimal/source/cisco-sna-minimal.csv", "network-flow.csv"],
  ]) write(name, readFileSync(path.join(root, source)));
  const payload = json({ items: [{ key: "host", label: "Host" }] });
  write("reference-pack.tar", reviewTar([
    ["manifest.json", json({ pack_key: "type_registry.browser_review", pack_kind: "type_registry", pack_version: "review-1", pack_contract_version: "cartulary.reference_pack.v1", verification_method: "manifest_sha256_v1", payloads: [{ path: "payload/data.json", sha256: hash(payload) }] })],
    ["payload/data.json", payload],
  ]));

  // Same current-format, empty-workbook input as the real incident-import
  // browser fixture. The machine source catalog owns required logical paths.
  const catalog = JSON.parse(readFileSync(path.join(root, "contracts/incident-bundles/source_catalog.json"), "utf8"));
  const sources = new Map();
  for (const source of [...catalog.families.flatMap((family) => family.paths), ...catalog.special_consumers]) {
    if (source.versions.includes(4) && !source.logical_path.includes("*")) sources.set(source.logical_path, Buffer.alloc(0));
  }
  const incidentId = randomUUID();
  const actorId = randomUUID();
  const key = `REVIEW-IMPORT-${incidentId}`;
  const timestamp = "2026-01-02T03:04:05+00:00";
  sources.set("data/incident.json", json({ id: incidentId, incident_key: key, incident_key_canonical: key, title: "Imported browser review investigation", description: null, status: "active", severity: null, tlp: null, current_phase: null, primary_external_case_ref: null, created_by_user_id: actorId, created_at: timestamp, updated_at: timestamp, updated_by_user_id: actorId, incident_version: 1, closed_at: null }));
  sources.set("data/actors.ndjson", json({ actor_id: actorId, display_name: "Review source operator" }));
  sources.set("data/reference_pack_refs.json", json([]));
  sources.set("data/timeline_time_profiles.ndjson", json({ incident_id: incidentId, enabled: false, local_offset_minutes: null, local_label: null, profile_version: 1, updated_at: timestamp, updated_by_user_id: actorId }));
  const files = [...sources].sort(([a], [b]) => a.localeCompare(b)).map(([name, bytes]) => ({ path: name, sha256: `sha256:${hash(bytes)}`, size_bytes: bytes.length, required: true }));
  sources.set("manifest.json", json({ bundle_format: "cartulary.incident_bundle", bundle_version: 4, bundle_id: randomUUID(), incident_id: incidentId, incident_key: key, exported_at: timestamp, source_change_set_high_watermark: `cartulary.source_boundary.v1:${hash(Buffer.from(JSON.stringify(files)))}`, history_mode: "full", blob_mode: "full", reference_pack_mode: "refs_only", optional_sections: [], required_capabilities: [], files }));
  sources.set("integrity/checksums.sha256", Buffer.from(files.map((file) => `${file.sha256.slice(7)}  ${file.path}\n`).join("")));
  write("incident.tar", reviewTar(sources));
  return directory;
}
