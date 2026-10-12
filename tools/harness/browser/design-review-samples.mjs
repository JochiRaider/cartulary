import { fixtureTar } from "../fixtures/ustar.mjs";
import { incidentBundleFixture } from "../fixtures/incident-bundle.mjs";
import { createHash } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

const hash = (bytes) => createHash("sha256").update(bytes).digest("hex");
const json = (value) => Buffer.from(`${JSON.stringify(value)}\n`);


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
  write("reference-pack.tar", fixtureTar([
    ["manifest.json", json({ pack_key: "type_registry.browser_review", pack_kind: "type_registry", pack_version: "review-1", pack_contract_version: "cartulary.reference_pack.v1", verification_method: "manifest_sha256_v1", payloads: [{ path: "payload/data.json", sha256: hash(payload) }] })],
    ["payload/data.json", payload],
  ]));

  write("incident.tar", incidentBundleFixture("active", root).upload.buffer);
  return directory;
}
