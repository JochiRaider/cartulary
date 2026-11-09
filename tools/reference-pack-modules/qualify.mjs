import { createHash } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { decodeGoPackages, validateModuleClosure } from "./closure.mjs";
const [descriptorPath, packagesPath, sourceRoot, moduleCache, goRoot, buildCache, output, isolationPath] = process.argv.slice(2);
if (!output || process.argv.length !== 10) throw new Error("explicit module qualification inputs required");
const descriptor = JSON.parse(readFileSync(descriptorPath, "utf8"));
const packages = decodeGoPackages(readFileSync(packagesPath, "utf8"));
validateModuleClosure(packages, { modulePath: descriptor.module_path, sourceRoot, moduleCache, goRoot, buildCache });
const sha256 = (value) => createHash("sha256").update(value).digest("hex");
if (sha256(readFileSync(descriptor.archive_path)) !== descriptor.archive_sha256) throw new Error("module archive changed during qualification");
for (const file of descriptor.files) {
  const bytes = readFileSync(path.join(sourceRoot, file.path));
  if (bytes.length !== file.size_bytes || sha256(bytes) !== file.sha256) throw new Error("exported module source changed during qualification");
}
const isolation = JSON.parse(readFileSync(isolationPath, "utf8"));
if (isolation.status !== "pass" || isolation.network !== false || isolation.shared_cache !== false) throw new Error("filesystem isolation evidence missing");
const recipe = sha256(Buffer.concat(["tools/reference-pack-modules/check.sh", "tools/reference-pack-modules/closure.mjs", "tools/reference-pack-modules/isolation.mjs", "tools/reference-pack-modules/isolated-go.mjs", "tools/reference-pack-modules/qualify.mjs", "tools/toolchain_pins.json"].map(file => readFileSync(file))));
const result = {
  schema_id: "cartulary.reference_pack_module_qualification.v1",
  module_path: descriptor.module_path, version: descriptor.version,
  source_sha256: descriptor.source_sha256, archive_sha256: descriptor.archive_sha256,
  go_workspace: "off", dependency_resolution: "readonly", network_resolution: "off",
  packages: [...new Set(packages.map(pkg => pkg.ImportPath))].sort(),
  filesystem_isolation: "linux_bubblewrap",
  pinned_dependencies: isolation.dependencies,
  qualification_recipe_sha256: recipe,
  status: "pass",
};
const directory = path.join(output, "qualified-isolated-v1", descriptor.source_sha256, recipe);
mkdirSync(directory, { recursive: true, mode: 0o700 });
for (const [name, value] of [["archive.json", descriptor], ["qualification.json", result]]) {
  const bytes = Buffer.from(`${JSON.stringify(value)}\n`); const file = path.join(directory, name);
  try { writeFileSync(file, bytes, { flag: "wx", mode: 0o600 }); }
  catch (error) { if (error.code !== "EEXIST" || !readFileSync(file).equals(bytes)) throw error; }
}
