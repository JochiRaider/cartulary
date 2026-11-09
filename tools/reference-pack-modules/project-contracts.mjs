import { readFileSync, mkdirSync, writeFileSync, renameSync } from "node:fs";
import { dirname } from "node:path";

// The standalone module owns neutral machine declarations. Application
// contracts are downstream copies; neither module generation nor execution
// discovers application schemas or specification prose.
const manifest = JSON.parse(readFileSync("tools/reference-pack-modules/projections.json", "utf8"));
if (manifest.schema_id !== "cartulary.reference_pack_format_projections.v1" || !Array.isArray(manifest.projections)) throw new Error("invalid format projection inventory");
const outputs = new Set();
const prepared = manifest.projections.map(({ source, output }) => {
  if (!/^referencepacks\/format\/(?:envelope|dataset)\/schemas\/[a-z0-9_.-]+\.json$/u.test(source) ||
    !/^contracts\/reference-packs\/(?:profiles\/)?[a-z0-9_.-]+\.json$/u.test(output) || outputs.has(output)) throw new Error("invalid format projection boundary");
  outputs.add(output);
  const raw = readFileSync(source);
  JSON.parse(raw);
  return { output, raw };
});
for (const { output, raw } of prepared) {
  mkdirSync(dirname(output), { recursive: true });
  try { if (readFileSync(output).equals(raw)) continue; } catch (error) { if (error.code !== "ENOENT") throw error; }
  const temporary = `${output}.projection-${process.pid}`;
  writeFileSync(temporary, raw, { flag: "wx", mode: 0o644 });
  renameSync(temporary, output);
}
