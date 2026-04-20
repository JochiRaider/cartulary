import { readFileSync } from "node:fs";
import path from "node:path";

const phase = process.argv[2];
if (!phase) {
  throw new Error("usage: check-phase-map.mjs <phase>");
}

const root = process.cwd();
const manifestPath = path.join(root, "tools", `${phase}_test_map.json`);
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));

if (!Array.isArray(manifest.expected_ids) || manifest.expected_ids.length === 0) {
  throw new Error(`manifest ${manifestPath} must define a non-empty expected_ids array`);
}

const sections = [
  ["unit", "U-"],
  ["integration", "I-"],
  ["e2e", "E-"],
];

const entries = [];
for (const [section, prefix] of sections) {
  for (const entry of manifest[section] ?? []) {
    if (typeof entry.id !== "string" || !entry.id.startsWith(prefix)) {
      throw new Error(`manifest entry in ${section} has invalid id: ${JSON.stringify(entry)}`);
    }
    entries.push(entry);
  }
}

const ids = entries.map((entry) => entry.id);
const uniqueIds = new Set(ids);
if (uniqueIds.size !== ids.length) {
  throw new Error(`duplicate ids in ${manifestPath}`);
}

const expected = manifest.expected_ids;
const missing = expected.filter((id) => !uniqueIds.has(id));
const unexpected = ids.filter((id) => !expected.includes(id));
if (missing.length > 0 || unexpected.length > 0) {
  throw new Error(
    `${phase} manifest mismatch: missing=${missing.join(",") || "none"} unexpected=${unexpected.join(",") || "none"}`,
  );
}

for (const entry of entries) {
  const targetPath = path.join(root, entry.file);
  const source = readFileSync(targetPath, "utf8");
  const needle = entry.symbol ?? entry.title;
  if (!needle) {
    throw new Error(`manifest entry ${entry.id} is missing symbol/title`);
  }
  if (!source.includes(needle)) {
    throw new Error(`manifest entry ${entry.id} not found in ${entry.file}: ${needle}`);
  }
}

for (const target of manifest.forbidden_id_files ?? []) {
  const source = readFileSync(path.join(root, target), "utf8");
  for (const id of expected) {
    if (source.includes(id) || source.includes(id.replaceAll("-", "_"))) {
      throw new Error(`${target} must not claim phase id ${id}`);
    }
  }
}

console.log(`${phase} traceability map verified`);
