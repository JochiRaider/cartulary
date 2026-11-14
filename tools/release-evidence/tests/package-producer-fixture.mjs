// Synthetic image/tool boundaries for the real package producer and native owner.
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const repoRoot = path.resolve(import.meta.dirname, "../../..");
const hash = bytes => createHash("sha256").update(bytes).digest("hex");
export const imageIDs = ["1", "2", "3"].map(value => `sha256:${value.repeat(64)}`);
export const borrowedContainer = "f".repeat(64);
export const borrowedTags = { "borrowed:stable": `sha256:${"b".repeat(64)}`, [`cartulary/local:sha256-${"b".repeat(64)}`]: `sha256:${"b".repeat(64)}` };
function write(file, value, mode = 0o600) {
 fs.mkdirSync(path.dirname(file), { recursive: true });
 fs.writeFileSync(file, typeof value === "object" && !Buffer.isBuffer(value) ? JSON.stringify(value) : value, { mode });
}
export function createPackageProducerFixture(directory, scenario) {
 const source = path.join(directory, "source"), bin = path.join(directory, "bin"), data = path.join(directory, "data");
 for (const name of [source, bin, data]) fs.mkdirSync(name, { recursive: true });
 for (const name of ["server", "migrate", "operator"]) {
  const bytes = Buffer.from(`independent ${name} binary fixture\n`);
  write(path.join(source, "build/bin", name), bytes);
  write(path.join(source, "build/bin", `${name}.crypto.json`), { schema_id: "cartulary.cryptographic_build_receipt.v1",
   binary_sha256: hash(bytes), toolchain: "go1.27.3", selector: "v1.0.0-c2097c7c", service_version: "v1.0.0", archive_sha256: "a".repeat(64),
   build_info: { GoVersion: "go1.27.3", Path: `example.test/${name}`, Main: { Path: "example.test/module", Version: "(devel)" }, Settings: [{ Key: "GOFIPS140", Value: "v1.0.0-c2097c7c" }] } });
 }
 const renderer = path.join(source, "build/report-renderer"), bytes = Buffer.from("renderer fixture");
 const manifest = { schema_id: "cartulary.reporting_runtime_installation.v1", platform: "linux/amd64", node_version: "1.0.0", package_manager: "pnpm@1.0.0", chromium_version: "synthetic", bubblewrap_version: "1.0.0",
  lockfile_sha256: hash(bytes), package_store_digest: hash(bytes), font_manifest_sha256: hash(bytes), chromium_executable_sha256: hash(bytes), chromium_executable_path: "runtime/browser/chrome-headless-shell",
  packages: Object.fromEntries(["@slidev/cli", "@slidev/theme-default", "markdown-it", "mermaid", "pdf-lib", "playwright-chromium"].map(name => [name, "1.0.0"])), files: [{ path: "runtime/executable", sha256: hash(bytes) }] };
 const encoded = Buffer.from(JSON.stringify(manifest)), digest = hash(encoded);
 write(path.join(renderer, `${digest}.json`), encoded);
 write(path.join(renderer, `${digest}.root/runtime/executable`), bytes, 0o555);
 write(path.join(renderer, "installation.json"), { schema_id: "cartulary.reporting_runtime_receipt.v1", manifest_sha256: digest, root: `${digest}.root`, manifest: `${digest}.json` });
 const deploy = path.join(source, "deploy/mvp");
 for (const name of ["Containerfile", "application-seccomp.v1.json", ".env.example", "config.toml.example", "bootstrap-admin.json.example", "revisions-conflict-token-key-ring.json.example", "restore-verification-target.toml.example", "reference-pack-administration.toml.example", "postgres-entrypoint.sh", "postgres-hba.conf", "postgres-ident.conf", "postgres-provision.sh", "postgres-provision.sql", "seaweed-entrypoint.sh"]) write(path.join(deploy, name), "fixture\n");
 for (const name of ["docker-compose.yml", "docker-compose.reference-packs.yml"]) write(path.join(deploy, name), "services:\n  app:\n    image: ${CARTULARY_PACKAGE_APPLICATION_IMAGE:?immutable release required}\n  postgres:\n    image: docker.io/library/postgres:fixture\n  seaweed:\n    image: docker.io/seaweedfs/seaweedfs:fixture\n");
 for (const name of ["release-verify.sh", "install.sh"]) write(path.join(deploy, "scripts", name), fs.readFileSync(path.join(repoRoot, "deploy/mvp/scripts", name)), 0o755);
 fs.mkdirSync(path.join(deploy, "systemd"));
 const tar = (directory, archive, names) => {
  const result = spawnSync("tar", ["-cf", archive, "-C", directory, ...names], { encoding: "utf8", timeout: 10000 });
  assert.ifError(result.error); assert.equal(result.status, 0, result.stderr);
 };
 const configs = {};
 for (const [index, role] of ["application", "postgres", "seaweedfs"].entries()) {
  const config = JSON.stringify({ os: "linux", architecture: "amd64", rootfs: { diff_ids: [`sha256:${String(index + 4).repeat(64)}`] } });
  const configID = hash(config), staging = path.join(data, role);
  write(path.join(staging, `${configID}.json`), config);
  write(path.join(staging, "manifest.json"), [{ Config: `${configID}.json`, Layers: [] }]);
  tar(staging, path.join(data, `${role}.tar`), ["manifest.json", `${configID}.json`]);
  configs[role] = configID;
 }
 write(path.join(data, "filesystem/LICENSE"), "fixture copyright and permission\n");
 tar(path.join(data, "filesystem"), path.join(data, "filesystem.tar"), ["LICENSE"]);
 const state = path.join(directory, "state.json"), control = path.join(directory, "control.json"), log = path.join(directory, "calls.jsonl");
 write(state, { tags: { ...borrowedTags }, containers: { [borrowedContainer]: { Id: borrowedContainer, Name: "/borrowed", Config: { Labels: {} } } }, source, data, configs, stages: [] });
 write(control, { scenario }); write(log, "");
 const program = fileURLToPath(import.meta.url);
 for (const name of ["docker", "syft"]) write(path.join(bin, name), `#!${process.execPath}\nimport { runPackageFixtureTool } from ${JSON.stringify(`file://${program}`)};\nrunPackageFixtureTool(${JSON.stringify(name)});\n`, 0o700);
 write(path.join(bin, "go"), "#!/bin/sh\nprintf 'mod github.com/anchore/syft v1.44.0 fixture\\n'\n", 0o700);
 return { source, state, control, log, archiveRoot: path.join(directory, "archives"), environment: { PATH: `${bin}:${process.env.PATH}`, GO: path.join(bin, "go"), SYFT_BIN: path.join(bin, "syft"), PACKAGE_FIXTURE_STATE: state, PACKAGE_FIXTURE_CONTROL: control, PACKAGE_FIXTURE_LOG: log } };
}
export function runPackageFixtureTool(tool) {
 const args = process.argv.slice(2), state = JSON.parse(fs.readFileSync(process.env.PACKAGE_FIXTURE_STATE)), { scenario } = JSON.parse(fs.readFileSync(process.env.PACKAGE_FIXTURE_CONTROL));
 fs.appendFileSync(process.env.PACKAGE_FIXTURE_LOG, JSON.stringify([tool, ...args]) + "\n");
 const save = () => write(process.env.PACKAGE_FIXTURE_STATE, state);
 const output = value => process.stdout.write(typeof value === "string" ? value + "\n" : JSON.stringify(value) + "\n");
 const die = () => { process.kill(process.ppid, "SIGKILL"); process.exit(0); };
 const fail = () => { process.stderr.write("private fixture failure\n"); process.exit(1); };
 const referenceID = reference => state.tags[reference] ?? (imageIDs.includes(reference) ? reference : reference.includes("postgres") ? imageIDs[1] : reference.includes("seaweedfs") ? imageIDs[2] : null);
 if (tool === "syft") {
  const role = path.basename(args[1], ".tar");
  output({ descriptor: { name: "syft", version: "1.44.0" }, source: { type: "image", metadata: { imageID: `sha256:${state.configs[role]}` } }, artifacts: [{ id: role, name: role, version: "1.0.0", type: "binary", licenses: [] }] }); return;
 }
 if (args[0] === "info") { output(args.at(-1) === "{{.ID}}" ? "fixture-package-engine" : "linux/amd64"); return; }
 if (args[0] === "ps") {
  const expression = args[args.indexOf("--filter") + 1];
  const name = expression.slice("name=^/".length, -1);
  output(Object.values(state.containers).filter(value => value.Name === `/${name}`).map(value => value.Id).join("\n")); return;
 }
 if (args[0] === "image" && args[1] === "ls") { output(state.tags[args.at(-1).slice("reference=".length)] ?? ""); return; }
 if (args[0] === "build") {
  state.stages.push(path.dirname(args.at(-1))); save();
  if (scenario === "death_pending") die();
  state.tags[args[args.indexOf("--tag") + 1]] = imageIDs[0]; save();
  if (scenario === "death_build") die();
  if (scenario === "lost_build") fail();
  return;
 }
 if (args[0] === "tag") { assert.equal(referenceID(args[1]), imageIDs[0]); state.tags[args[2]] = args[1]; save(); return; }
 if (args[0] === "pull" || args[0] === "image" && args[1] === "load") return;
 if (args[0] === "image" && args[1] === "inspect") {
  const id = referenceID(args.at(-1)); if (!id) fail();
  if (args.includes("--format")) { output(args.includes("{{.Id}}") ? id : `${id} linux/amd64`); return; }
  output([{ Id: id, Os: "linux", Architecture: "amd64", RootFS: { Layers: [`sha256:${String(imageIDs.indexOf(id) + 4).repeat(64)}`] } }]); return;
 }
 if (args[0] === "image" && args[1] === "save") {
  const role = ["application", "postgres", "seaweedfs"][imageIDs.indexOf(referenceID(args.at(-1)))];
  fs.copyFileSync(path.join(state.data, `${role}.tar`), args[args.indexOf("--output") + 1]); return;
 }
 if (args[0] === "create") {
  const labels = {};
  for (let index = 0; index < args.length; index++) if (args[index] === "--label") { const [key, value] = args[++index].split("="); labels[key] = value; }
  const role = labels["cartulary.package.role"], id = String(["application", "postgres", "seaweedfs"].indexOf(role) + 7).repeat(64);
  assert.match(id, /^[789]{64}$/u);
  state.containers[id] = { Id: id, Name: `/${args[args.indexOf("--name") + 1]}`, Config: { Labels: labels } }; save();
  if (scenario === "death_create") die();
  if (["lost_create", "all_cleanup_fail"].includes(scenario)) fail();
  output(scenario === "malformed_create" ? "--all" : id); return;
 }
 if (args[0] === "inspect") { output(state.containers[args.at(-1)]); return; }
 if (args[0] === "cp") { fs.copyFileSync(path.join(state.source, "build/bin", args[1].split("cartulary-").at(-1)), args[2]); return; }
 if (args[0] === "export") { fs.copyFileSync(path.join(state.data, "filesystem.tar"), args[args.indexOf("--output") + 1]); return; }
 if (args[0] === "rm") {
  const id = args.at(-1); assert.notEqual(id, borrowedContainer);
  if (["cleanup_fail", "all_cleanup_fail"].includes(scenario)) fail();
  delete state.containers[id]; save(); return;
 }
 if (args[0] === "image" && args[1] === "rm") {
  assert.match(args[2], /^cartulary\/package-build:/u);
  if (scenario === "all_cleanup_fail") fail();
  delete state.tags[args[2]]; save(); return;
 }
 throw new Error("unhandled fixture command");
}

// The worker owns the actual native runtime; forced death bypasses all JS cleanup.
export const packageWorkerScript = `
import fs from "node:fs";
import { createRuntime } from ${JSON.stringify(new URL("../../harness/runtime/session-client.mjs", import.meta.url).href)};
import { produce, inspect, unpack } from ${JSON.stringify(new URL("../package-release.mjs", import.meta.url).href)};
const options = JSON.parse(process.env.PACKAGE_FIXTURE_OPTIONS);
const runtime = createRuntime(options.runtime);
fs.writeFileSync(options.identity, JSON.stringify({root:runtime.root,leaseID:runtime.leaseID}));
Object.assign(process.env, runtime.producerEnvironment());
let failure;
try {
 if(options.operation === "inspect") inspect(options.producer);
 else if(options.operation === "unpack") unpack(options.destination, options.producer);
 else produce(options.producer);
}
catch(error) { failure = error; }
try { runtime.close(); } catch(error) { if(!failure) failure=error; }
fs.writeFileSync(options.result, JSON.stringify({failed:!!failure,reason:failure?.failure_reason??null,
 cleanup:failure?.cleanupFailures?.map(value=>value.cleanup_step)??[]}));
process.exitCode=failure?1:0;
`;
