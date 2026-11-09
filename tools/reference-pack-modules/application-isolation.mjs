import { readFileSync, writeFileSync, mkdirSync, cpSync, lstatSync, readdirSync, existsSync } from "node:fs";
import path from "node:path";
import { generationProfile, generationCommands } from "./product-generation.mjs";
import { nodeInputs } from "./node-inputs.mjs";
import { isolated, executableInputs, privateDirectories, retainIsolation } from "./isolation.mjs";
import { goInputs } from "./isolated-go.mjs";
import { decodeGoPackages, validateApplicationClosure } from "./closure.mjs";

const [packagesPath, work, go, moduleCache, profilePath] = process.argv.slice(2);
if (!profilePath) throw new Error("explicit application qualification inputs required");
const root = process.cwd();
const packages = decodeGoPackages(readFileSync(packagesPath, "utf8")); validateApplicationClosure(packages);
const [source, output] = privateDirectories(work, ["application", "binaries"]);
const files = new Set(["go.mod", "go.sum"]);
const generatorPackages = decodeGoPackages(readFileSync(path.join(work, "generator-packages.json"), "utf8"));
validateApplicationClosure(generatorPackages);
for (const pkg of [...packages, ...generatorPackages]) {
  if (!pkg.Module?.Main) continue;
  for (const group of ["GoFiles", "CgoFiles", "CFiles", "CXXFiles", "HFiles", "SFiles", "SysoFiles", "EmbedFiles"]) for (const name of pkg[group] ?? []) files.add(path.relative(root, path.resolve(pkg.Dir, name)));
}
const profile = generationProfile(profilePath);
const projections = JSON.parse(readFileSync(profile.format_projections)).projections;
const families = JSON.parse(readFileSync(profile.contract_registry)).families;
const active = families.filter(family => family.generation_status === "active");
const fixtureOutputs = JSON.parse(readFileSync(profile.fixture_outputs)).generated_paths.filter(file => file.startsWith("contracts/reference-pack-fixtures/"));
const outputs = new Set([...projections.map(p => p.output), ...fixtureOutputs, ...active.flatMap(f => f.generated_outputs), ...profile.protocol_outputs]);
const admit = file => {
  if (path.isAbsolute(file) || file.split(path.sep).some(part => !part || part === ".." || part === "docs" || /^readme(?:\.|$)/iu.test(part)) || /\.md$/iu.test(file) || file.startsWith("fixtures/") || file.startsWith("contracts/reference-pack-producers/") || /^referencepacks\/(?!format\/(?:envelope|dataset)\/schemas\/)/u.test(file)) throw new Error(`forbidden application input: ${file}`);
};
const add = file => {
  admit(file);
  const info = lstatSync(file);
  if (info.isSymbolicLink()) throw new Error(`symlinked application input: ${file}`);
  if (info.isDirectory()) {
    for (const name of readdirSync(file).sort()) {
      // Repository guidance/prose is never an executable generator input.
      if (/\.md$/iu.test(name) || /^readme(?:\.|$)/iu.test(name)) continue;
      add(path.join(file, name));
    }
  } else if (info.isFile()) files.add(file);
  else throw new Error(`nonregular application input: ${file}`);
};
for (const file of [profilePath, profile.format_projections, profile.fixture_outputs, ...profile.copy_paths, ...projections.map(p => p.source), ...active.map(f => f.contract_root)]) add(file);
for (const file of outputs) { admit(file); files.delete(file); }
// Missing generated packages are intentional: the actual generator must create
// every declared output from authored inputs before the application can build.
for (const file of [...files].sort()) {
  admit(file);
  if (lstatSync(file).isSymbolicLink()) throw new Error(`symlinked application input: ${file}`);
  const target = path.join(source, file); mkdirSync(path.dirname(target), { recursive: true }); cpSync(file, target);
}
const outputDirectories = [...new Set([...outputs].map(file => path.dirname(file)))].sort().filter((dir, i, all) => !all.some((other, j) => j !== i && dir.startsWith(other + "/")));
for (const dir of outputDirectories) mkdirSync(path.join(source, dir), {recursive: true});
for (const file of outputs) if (existsSync(path.join(source, file))) throw new Error(`preexisting generation answer: ${file}`);
const nodeDependencies = nodeInputs(root, path.join(source, "node_modules"), profile.node_dependencies);
const inputs = goInputs(go, source, work, moduleCache);
const manifest = { ...inputs, readonly: [...inputs.readonly, ...executableInputs(process.execPath), source], writable: [...inputs.writable, ...outputDirectories.map(dir => path.join(source, dir)), output], readonly_overlays: [...files].filter(file => outputDirectories.some(dir => file.startsWith(dir + "/"))).map(file => path.join(source, file)) };
const commands = generationCommands(profile, inputs.go);
const generationManifests = commands.map((command, index) => {
  const step = profile.steps[index];
  if (!step.readonly_inputs) return {...manifest, writable: manifest.writable.filter(dir => dir !== output), command};
  for (const file of step.readonly_inputs) admit(file);
  const writable = outputDirectories.filter(dir => dir.startsWith(step.output_prefix)).map(dir => path.join(source, dir));
  const readonly = [...executableInputs(process.execPath), ...step.readonly_inputs.map(file => path.join(source, file))];
  return {readonly, writable, readonly_overlays: manifest.readonly_overlays.filter(file => writable.some(dir => file.startsWith(dir + "/"))), cwd: source, command};
});
for (const [index, capability] of generationManifests.entries()) {
  isolated(capability);
  if (profile.steps[index].readonly_inputs) {
    const denied = [path.join(source, "internal/modules/reference_data/internal/packformat/manifest.go"), path.join(root, "referencepacks/format/envelope/manifest.go"), path.join(root, "fixtures/reference-packs/predecessor-contract-v2-vectors/manifest.v1.json")];
    const probe = `import fs from "node:fs";for(const file of ${JSON.stringify(denied)}){try{fs.readFileSync(file);throw new Error("undeclared generation input readable")}catch(error){if(!["ENOENT","EACCES"].includes(error.code))throw error}}`;
    isolated({...capability, command: [process.execPath, "--input-type=module", "-e", probe]});
  }
}
const inventoryPath = path.join(work, "generation-output-inventory.json");
const inventory = {inputs: [...files].sort(), outputs: [...outputs].sort(), output_directories: outputDirectories};
writeFileSync(inventoryPath, JSON.stringify(inventory));
const validator = path.join(root, "tools/reference-pack-modules/generation-outputs.mjs");
// The generator never sees checked-in answers. Only this trusted comparison
// receives those exact files, and it cannot write to either tree. Host code
// consumes hashes from stdout, never follows a generator-created pathname.
const outputValidation = {readonly: [...executableInputs(process.execPath), validator, inventoryPath, source, ...[...files, ...outputs].map(file => path.join(root, file))], cwd: source, command: [process.execPath, validator, source, root, inventoryPath]};
const outputHashes = JSON.parse(isolated(outputValidation));
const run = args => isolated({ ...manifest, command: [inputs.go, "-C", source, ...args] });
run(["build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-o", output + "/", "./cmd/..."]);
validateApplicationClosure(decodeGoPackages(run(["list", "-mod=readonly", "-deps", "-json", "./cmd/..."])));
// The operator's command surface can execute with only its completed binary;
// neither sources, corpora, signing dependencies nor build caches are mounted.
const usage = isolated({ readonly: [output], cwd: output, environment: { GODEBUG: "fips140=on" }, command: [path.join(output, "operator"), "--help"], expected_exit: 2 }, { includeStderr: true });
if (!usage.startsWith("usage:\n") || !usage.includes("operator reference-pack import <bundle_name>")) throw new Error("isolated operator usage probe failed");
writeFileSync(path.join(work, "isolation.json"), JSON.stringify({ schema_id: "cartulary.reference_pack_application_isolation.v1", status: "pass", inputs: [...files].sort(), product_generation: generationManifests, generated_outputs: outputHashes, output_validation: outputValidation, node_dependencies: nodeDependencies, application_build: "./cmd/...", runtime_probe: "operator --help", network: false, shared_cache: false }));
retainIsolation("application", manifest, { inputs: [...files].sort(), generation: generationManifests, generated_outputs: outputHashes, output_validation: outputValidation, node_dependencies: nodeDependencies, runtime_probe: "operator --help" });
process.stdout.write("Application generation, build and operator probe qualified with producer inputs inaccessible.\n");
