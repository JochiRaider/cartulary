import { readFileSync } from "node:fs";
import path from "node:path";

const root = process.cwd();
const manifestPath = path.join(root, "tools", "phase2_test_map.json");
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));

const expected = [
  ...Array.from({ length: 10 }, (_, index) => `U-2-${String(index + 1).padStart(2, "0")}`),
  ...Array.from({ length: 6 }, (_, index) => `I-2-${String(index + 1).padStart(2, "0")}`),
  ...Array.from({ length: 6 }, (_, index) => `E-2-${String(index + 1).padStart(2, "0")}`),
];

const entries = [
  ...(manifest.unit ?? []),
  ...(manifest.integration ?? []),
  ...(manifest.e2e ?? []),
];

const ids = entries.map((entry) => entry.id);
const uniqueIds = new Set(ids);
if (uniqueIds.size !== ids.length) {
  throw new Error(`duplicate Phase 2 ids in ${manifestPath}`);
}

const missing = expected.filter((id) => !uniqueIds.has(id));
const unexpected = ids.filter((id) => !expected.includes(id));
if (missing.length > 0 || unexpected.length > 0) {
  throw new Error(
    `phase2 manifest mismatch: missing=${missing.join(",") || "none"} unexpected=${unexpected.join(",") || "none"}`,
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

const smokeSource = readFileSync(
  path.join(root, "cmd", "server", "main_phase2_smoke_test.go"),
  "utf8",
);
for (const id of expected) {
  if (smokeSource.includes(id) || smokeSource.includes(id.replaceAll("-", "_"))) {
    throw new Error(`process smoke file must not claim Phase 2 guide id ${id}`);
  }
}

console.log("phase2 traceability map verified");
