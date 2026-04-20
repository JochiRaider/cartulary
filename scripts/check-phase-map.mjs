import { readFileSync } from "node:fs";
import path from "node:path";

const phase = process.argv[2];
if (!phase) {
  throw new Error("usage: check-phase-map.mjs <phase>");
}
const phaseMatch = /^phase(\d+)$/.exec(phase);
if (!phaseMatch) {
  throw new Error(`invalid phase name ${phase}; expected phase<number>`);
}
const phaseNumber = phaseMatch[1];

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
    if (!new RegExp(`^${prefix}${phaseNumber}-\\d{2}$`).test(entry.id)) {
      throw new Error(`manifest entry ${entry.id} does not belong to ${phase}`);
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
  const hyphenMatches = source.match(new RegExp(String.raw`\b[UIE]-${phaseNumber}-\d{2}\b`, "g")) ?? [];
  const underscoreMatches = source.match(new RegExp(String.raw`\b[UIE]_${phaseNumber}_\d{2}\b`, "g")) ?? [];
  const claimedIDs = new Set([
    ...hyphenMatches,
    ...underscoreMatches.map((value) => value.replaceAll("_", "-")),
  ]);
  if (claimedIDs.size > 0) {
    throw new Error(`${target} must not claim ${phase} authoritative ids: ${Array.from(claimedIDs).sort().join(", ")}`);
  }
}

console.log(`${phase} traceability map verified`);
