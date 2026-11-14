// One Make-owned producer. Qualification consumes its immutable archive verbatim.
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import { chmodSync, copyFileSync, lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";

import { publicExitCodeForFailure, validateSchemaSync } from "../harness/contract/index.mjs";
import { CommandFailure, reportCommandFailure } from "../harness/runtime/command-failure.mjs";
import { borrowRuntime } from "../harness/runtime/session-client.mjs";
import { reservePackageResources } from "./package-resources.mjs";
import { stageRendererInstallation } from "../report-renderer/stage-installation.mjs";
import { packageRunContext, publishPackageEvidence, readPackageJSON, readPackageReceipt } from "./package-identity.mjs";
import { hashPackageFile as hashFile, preparePackageArchiveDirectory, publishPackageArchive, withVerifiedPackageArchive } from "./package-archive.mjs";

const root = path.resolve(import.meta.dirname, "../..");
const packageArchiveRoot = path.join(root, ".cartulary/release-artifacts/packages");
const sha = (bytes) => createHash("sha256").update(bytes).digest("hex");
function run(command, args, options = {}) {
  const result = spawnSync(command, args, { cwd: root, encoding: "utf8", timeout: 600000, maxBuffer: 64 * 1024 * 1024, ...options });
  assert.equal(result.status, 0, `${path.basename(command)} failed: ${String(result.stderr || result.error || "").slice(-2000)}`);
  return result.stdout.trim();
}
function write(file, value) { if(value.schema_id?.startsWith("cartulary.")) validateSchemaSync(value.schema_id,value); mkdirSync(path.dirname(file), { recursive: true }); writeFileSync(file, JSON.stringify(value, null, 2) + "\n", { mode: 0o644 }); chmodSync(file,0o644); }
function copy(source, destination) { mkdirSync(path.dirname(destination), { recursive: true }); copyFileSync(source, destination); chmodSync(destination, source.endsWith(".sh") ? 0o755 : 0o644); }
function files(directory, prefix = "") {
  return readdirSync(directory).sort().flatMap((name) => {
    const rel = prefix ? `${prefix}/${name}` : name;
    const file = path.join(directory, name), info = lstatSync(file);
    assert.ok(!info.isSymbolicLink() && info.nlink === (info.isDirectory() ? info.nlink : 1));
    return info.isDirectory() ? files(file, rel) : [{ path: rel, size_bytes: info.size, sha256: hashFile(file), mode: (info.mode & 0o777).toString(8) }];
  });
}
function imageInfo(reference) {
  const [image] = JSON.parse(run("docker", ["image", "inspect", reference]));
  assert.equal(`${image.Os}/${image.Architecture}`, "linux/amd64");
  assert.match(image.Id, /^sha256:[a-f0-9]{64}$/);
  return image;
}

export function finalizePackageResources({ allocation, stage, runtime }, primary) {
  const failures = allocation?.releaseAll() ?? [];
  const attempt = (step, action) => {
    try { action(); }
    catch (cause) {
      failures.push(Object.assign(new CommandFailure("package producer cleanup failed", {
        failure_class: "harness", failure_reason: "cleanup_error",
      }, { cause }), { cleanup_step: step }));
    }
  };
  if (stage) attempt("stage", () => rmSync(stage, { recursive: true, force: true }));
  if (runtime) attempt("runtime_handle", () => runtime.close());
  if (failures.length) {
    if (primary) (primary.cleanupFailures ??= []).push(...failures);
    else throw Object.assign(new CommandFailure("package producer cleanup failed", {
      failure_class: "harness", failure_reason: "cleanup_error",
    }, { cause: new AggregateError(failures) }), { cleanupFailures: failures });
  }
}

export function produce({ sourceRoot = root, archiveRoot = packageArchiveRoot } = {}) {
  const { runRoot } = packageRunContext();
  const runtime = borrowRuntime({ repoRoot: root, runRoot });
  let allocation, stage, primary;
  try {
    preparePackageArchiveDirectory(archiveRoot);
    const workspace = runtime.producerPath("package-production");
    mkdirSync(workspace, { recursive: true, mode: 0o700 });
    stage = mkdtempSync(path.join(workspace, "stage-"));
    const payload = path.join(stage, "release"); mkdirSync(payload);
    const scanner = process.env.SYFT_BIN;
    assert.ok(scanner && path.isAbsolute(scanner), "Make must supply the pinned inventory tool");
    const scannerBuild = run(process.env.GO || "go", ["version", "-m", scanner]);
    assert.match(scannerBuild, /mod\s+github\.com\/anchore\/syft\s+v1\.44\.0\s/);
    write(path.join(payload,"inventories/scanner.json"), {schema_id:"cartulary.image_inventory_tool.v1", module:"github.com/anchore/syft", version:"v1.44.0", binary_sha256:hashFile(scanner)});
    const context = path.join(stage, "image-context");
    mkdirSync(path.join(context, "build/bin"), { recursive: true });
    for (const binary of ["server", "migrate", "operator"]) copy(path.join(sourceRoot, "build/bin", binary), path.join(context, "build/bin", binary));
    copy(path.join(sourceRoot, "deploy/mvp/Containerfile"), path.join(context, "Containerfile"));
    for (const directory of ["backups", "exports", "reference-packs", "tmp"]) mkdirSync(path.join(context, "deploy/mvp/runtime-roots", directory), { recursive: true });
    stageRendererInstallation(path.join(sourceRoot, "build/report-renderer"), path.join(context, "build/report-renderer"));
    allocation = reservePackageResources(runtime);
    const appTag = allocation.identity("build_tag").name;
    run("docker", ["build", "--platform", "linux/amd64", "--file", path.join(context, "Containerfile"), "--tag", appTag, context]);
    const application = imageInfo(appTag);
    const immutableTag = `cartulary/local:sha256-${application.Id.slice(7)}`;
    const prior = spawnSync("docker", ["image", "inspect", "--format", "{{.Id}}", immutableTag], { encoding: "utf8" });
    assert.ok(prior.status !== 0 || prior.stdout.trim() === application.Id, "immutable application tag contaminated");
    run("docker", ["tag", application.Id, immutableTag]);
    const source = path.join(sourceRoot, "deploy/mvp");
    const compose = readFileSync(path.join(source, "docker-compose.yml"), "utf8");
    const references = [...compose.matchAll(/^    image: (docker\.io\/[^\n]+)$/gm)].map((m) => m[1]);
    assert.equal(references.length, 2);
    const images = [];
    const binaries = {};
    const notices = { schema_id: "cartulary.distribution_notices.v1", review_disposition: "unestablished", images: [] };
    for (const [role, reference] of [["application", immutableTag], ["postgres", references[0]], ["seaweedfs", references[1]]]) {
      if (role !== "application") run("docker", ["pull", "--platform", "linux/amd64", reference]);
      const info = imageInfo(reference);
      const archive = `images/${role}.tar`, inventory = `inventories/${role}.syft.json`;
      mkdirSync(path.join(payload, "images"), { recursive: true });
      mkdirSync(path.join(payload, "inventories"), { recursive: true });
      run("docker", ["image", "save", "--output", path.join(payload, archive), reference]);
      chmodSync(path.join(payload, archive), 0o644);
      const entries=JSON.parse(run("tar",["-xOf",path.join(payload,archive),"manifest.json"]));
      assert.equal(entries.length,1,"image archive must contain one selected image");
      assert.match(entries[0].Config,/^(?:blobs\/sha256\/)?[a-f0-9]{64}(?:\.json)?$/);
      const configBytes=spawnSync("tar",["-xOf",path.join(payload,archive),entries[0].Config],{maxBuffer:16*1024*1024,timeout:30000});
      assert.equal(configBytes.status,0);
      const configSHA256=sha(configBytes.stdout), config=JSON.parse(configBytes.stdout);
      assert.equal(`${config.os}/${config.architecture}`,"linux/amd64");
      assert.deepEqual(config.rootfs.diff_ids,info.RootFS.Layers);
      // Scan the exact shipped archive. Docker's manifest ID and the OCI config
      // digest are distinct identities; Syft reports the latter as imageID.
      const scan = run(scanner, ["scan", `docker-archive:${archive}`, "--output", "syft-json", "--quiet"], { cwd: payload, env: { ...process.env, SYFT_CHECK_FOR_APP_UPDATE: "false" } });
      const inventoryData = JSON.parse(scan);
      assert.equal(inventoryData.descriptor.name, "syft");
      assert.ok(["1.44.0", "[not provided]"].includes(inventoryData.descriptor.version));
      assert.equal(inventoryData.source.metadata.imageID, `sha256:${configSHA256}`);
      assert.ok(inventoryData.artifacts.length > 0, "empty shipped-content inventory");
      write(path.join(payload, inventory), inventoryData);
      const identity = allocation.identity(role);
      const temporaryContainer = run("docker", ["create", "--name", identity.name,
        ...Object.entries(identity.labels).flatMap(([key, value]) => ["--label", `${key}=${value}`]), reference]);
      assert.match(temporaryContainer, /^[a-f0-9]{64}$/u);
      if (role === "application") for (const binary of ["migrate", "operator", "server"]) {
        const output = path.join(stage, binary);
        run("docker", ["cp", `${temporaryContainer}:/usr/local/bin/cartulary-${binary}`, output]);
        const receipt = readPackageJSON(sourceRoot, `build/bin/${binary}.crypto.json`, "cartulary.cryptographic_build_receipt.v1");
        assert.equal(receipt.binary_sha256, hashFile(output), "shipped binary differs from admitted receipt");
        write(path.join(payload, "receipts", `${binary}.crypto.json`), receipt);
        binaries[binary] = { sha256: receipt.binary_sha256, receipt: `receipts/${binary}.crypto.json` };
      }
      const exported = path.join(stage, `${role}-filesystem.tar`);
      run("docker", ["export", "--output", exported, temporaryContainer]);
      allocation.release(role);
      const candidates = run("tar", ["-tf", exported]).split("\n").filter((entry) => /(^|\/)(copyright|licen[cs]e[^/]*|notice[^/]*)$/i.test(entry) && /^[A-Za-z0-9_./+@ -]+$/.test(entry) && !entry.split("/").includes("..") && !entry.startsWith("/"));
      const noticeFiles = [];
      for (const [index, member] of candidates.entries()) {
        const content = spawnSync("tar", ["-xOf", exported, member], { maxBuffer: 1048576, timeout: 30000 });
        if (content.status !== 0 || !content.stdout?.length) continue;
        const relative = `notices/${role}/${index}.txt`;
        mkdirSync(path.dirname(path.join(payload, relative)), { recursive: true });
        writeFileSync(path.join(payload, relative), content.stdout, { mode: 0o644 });
        chmodSync(path.join(payload,relative),0o644);
        noticeFiles.push({ source_path: member, path: relative, sha256: sha(content.stdout) });
      }
      rmSync(exported);
      notices.images.push({ image_id: info.Id, role, notice_files: noticeFiles, components: inventoryData.artifacts.map((component) => ({ id: component.id, name: component.name, version: component.version || null, type: component.type, licenses: component.licenses || [], review_disposition: "unestablished", findings: [...(!component.version ? ["version_metadata_missing"] : []), ...(!component.licenses?.length ? ["license_metadata_missing"] : []), "component_notice_association_requires_review"] })) });
      images.push({ role, reference, id: info.Id, config_sha256: configSHA256, archive, inventory });
    }
    write(path.join(payload, "inventories/notices.json"), notices);
    const assets = ["application-seccomp.v1.json", ".env.example", "config.toml.example", "bootstrap-admin.json.example", "revisions-conflict-token-key-ring.json.example", "restore-verification-target.toml.example", "reference-pack-administration.toml.example", "docker-compose.yml", "docker-compose.reference-packs.yml", "postgres-entrypoint.sh", "postgres-hba.conf", "postgres-ident.conf", "postgres-provision.sh", "postgres-provision.sql", "seaweed-entrypoint.sh"];
    for (const name of assets) copy(path.join(source, name), path.join(payload, "assets", name));
    for (const directory of ["scripts", "systemd"]) for (const name of readdirSync(path.join(source, directory))) copy(path.join(source, directory, name), path.join(payload, "assets", directory, name));
    for (const name of ["docker-compose.yml","docker-compose.reference-packs.yml"]) {
      let shippedCompose = readFileSync(path.join(payload, "assets",name), "utf8").replaceAll("${CARTULARY_PACKAGE_APPLICATION_IMAGE:?immutable release required}", application.Id);
      for (const image of images.slice(1)) shippedCompose = shippedCompose.replaceAll(image.reference, image.id);
      assert.ok(!/^\s+build:|context:|cartulary\/mvp:local/m.test(shippedCompose));
      for (const match of shippedCompose.matchAll(/^    image: (.+)$/gm)) assert.ok(images.some(image=>image.id===match[1]),"Compose image is outside the immutable release");
      writeFileSync(path.join(payload, "assets",name), shippedCompose);
    }
    const manifest = { schema_id: "cartulary.local_release_manifest.v1", platform: "linux/amd64", binaries, images, assets: files(payload).sort((a, b) => a.path.localeCompare(b.path, "en", { sensitivity: "variant" })) };
    manifest.assets.sort((a,b) => a.path < b.path ? -1 : a.path > b.path ? 1 : 0);
    validateSchemaSync(manifest.schema_id, manifest);
    write(path.join(payload, "release-manifest.json"), manifest);
    const digest = run(path.join(payload, "assets/scripts/release-verify.sh"), [payload, "images"]);
    const archive = path.join(stage, "package.tar");
    run("tar", ["--sort=name", "--mtime=@0", "--owner=0", "--group=0", "--numeric-owner", "-cf", archive, "-C", stage, "release"]);
    const published = publishPackageArchive(archiveRoot, digest, archive);
    const receipt = { schema_id: "cartulary.local_release_receipt.v1", manifest_sha256: digest, ...published, images: images.map(({role,id}) => ({role,id})), operator_sha256: binaries.operator.sha256, inventory_disposition: "content_complete_review_unestablished" };
    publishPackageEvidence(runRoot, "package", receipt);
    return receipt;
  } catch (error) { primary = error; throw error; }
  finally { finalizePackageResources({ allocation, stage, runtime }, primary); }
}

function extractPackage(destination, receipt, runtime, archiveRoot) {
  // This proves receipt-bound bytes, not the publisher's authenticity.
  withVerifiedPackageArchive(archiveRoot, receipt, runtime.producerPath("package-snapshots"), archive =>
    run("tar", ["--no-same-owner", "--same-permissions", "-xf", archive, "-C", destination]));
  const release = path.join(destination, "release");
  assert.equal(run(path.join(release, "assets/scripts/release-verify.sh"), [release, "images"]), receipt.manifest_sha256);
  return { release, receipt };
}
export function unpack(destination, { archiveRoot = packageArchiveRoot } = {}) {
  const { runRoot } = packageRunContext(), receipt = readPackageReceipt(runRoot);
  const runtime = borrowRuntime({ repoRoot: root, runRoot });
  let primary;
  try { return extractPackage(destination, receipt, runtime, archiveRoot); }
  catch (error) { primary = error; throw error; }
  finally { finalizePackageResources({ runtime }, primary); }
}
export function inspect({ archiveRoot = packageArchiveRoot } = {}) {
  const { runRoot } = packageRunContext();
  const runtime = borrowRuntime({ repoRoot: root, runRoot });
  let scratch, primary;
  try {
    const workspace = runtime.producerPath("package-inspection");
    mkdirSync(workspace, { recursive: true, mode: 0o700 });
    scratch = mkdtempSync(path.join(workspace, "inspect-"));
    const { release, receipt } = extractPackage(scratch, readPackageReceipt(runRoot), runtime, archiveRoot);
    const installation = path.join(scratch, "installation");
    run(path.join(release, "assets/scripts/install.sh"), [installation]);
    assert.equal(run(path.join(installation, "release/assets/scripts/release-verify.sh"), [path.join(installation, "release"), "images"]), receipt.manifest_sha256);
    assert.notEqual(spawnSync(path.join(release, "assets/scripts/install.sh"), [installation], { encoding:"utf8" }).status, 0);
    publishPackageEvidence(runRoot, "inspection", { ...receipt, schema_id: "cartulary.local_release_inspection.v1", installed_without_build_toolchain: true, occupied_destination_rejected: true });
  } catch (error) { primary = error; throw error; }
  finally { finalizePackageResources({ stage: scratch, runtime }, primary); }
}
if (process.argv[1] && path.resolve(process.argv[1]) === path.resolve(import.meta.filename)) {
  try {
    const command = process.argv[2];
    if (command === "produce") produce();
    else if (command === "inspect") inspect();
    else if (command === "unpack") { const {receipt} = unpack(process.argv[3]); process.stdout.write(JSON.stringify(receipt)+"\n"); }
    else throw new CommandFailure("unsupported package operation", { failure_class: "config", failure_reason: "usage_error" });
  } catch (error) {
    const failure = reportCommandFailure(root, error, { failure_class: "artifact", failure_reason: "artifact_error" });
    console.error(`Package release failed: ${failure.failure_class}/${failure.failure_reason}`);
    if (error.cleanupFailures?.length) console.error(`Package producer cleanup failed: ${error.cleanupFailures.length} steps`);
    process.exitCode = publicExitCodeForFailure(failure);
  }
}
