import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { createReadStream } from "node:fs";
import { chmod, copyFile, lstat, mkdir, mkdtemp, open, readFile, readdir, realpath, rename, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import { createRequire } from "node:module";
import { pathToFileURL } from "node:url";

import { restrictedExecutableInputRoots } from "../harness/test-catalog/restricted-input-boundary.mjs";

// Packaging is an explicit build operation. Rendering never invokes a package
// manager, discovers host libraries, or copies files from the application tree.
const repository = process.cwd();
if (process.platform !== "linux" || process.arch !== "x64") throw new Error("renderer installation requires linux/amd64");
const pins = JSON.parse(await readFile("tools/toolchain_pins.json", "utf8"));
if (process.versions.node !== pins.node_version) throw new Error("renderer Node pin mismatch");
const excluded = new Set([...restrictedExecutableInputRoots(repository), ".git", ".bin", ".modules.yaml", ".pnpm-workspace-state-v1.json"]);
const omit = (name) => excluded.has(name) || /(?:\.md|\.markdown)$/iu.test(name) || /^readme(?:\.|$)/iu.test(name);
const output = path.join(repository, "build/report-renderer");
await mkdir(output, { recursive: true });
const work = await mkdtemp(path.join(output, ".build-"));
const root = path.join(work, "root");
const nativeFiles = new Set();
const run = (program, args, options = {}) => execFileSync(program, args, { encoding: "utf8", maxBuffer: 8 << 20, ...options });
const sha = (bytes) => createHash("sha256").update(bytes).digest("hex");

async function copyRegular(source, destination) {
  const info = await lstat(source);
  if (!info.isFile()) throw new Error("runtime source must be a regular file");
  await mkdir(path.dirname(destination), { recursive: true });
  await copyFile(source, destination); // Never carry package-store hard links.
  await chmod(destination, info.mode & 0o111 ? 0o555 : 0o444);
  const file = await open(destination, "r");
  try {
    const prefix = Buffer.alloc(4);
    await file.read(prefix, 0, 4, 0);
    if (prefix.equals(Buffer.from([0x7f, 0x45, 0x4c, 0x46]))) nativeFiles.add(destination);
  } finally { await file.close(); }
}

async function copyTree(source, destination, omitModules = false) {
  await mkdir(destination, { recursive: true });
  for (const entry of await readdir(source, { withFileTypes: true })) {
    // Exclude before stat, traversal or content access, including dangling links.
    if (omit(entry.name) || (omitModules && (entry.name === "node_modules" || /\.(?:test|spec)\.[cm]?[jt]sx?$/u.test(entry.name)))) continue;
    if (entry.isDirectory()) await copyTree(path.join(source, entry.name), path.join(destination, entry.name), omitModules);
    else if (entry.isFile()) await copyRegular(path.join(source, entry.name), path.join(destination, entry.name));
    else throw new Error("runtime tree contains a link or special entry");
  }
}

async function copyHostFile(source, destination) {
  // This is confined to explicitly selected installed binaries/libraries/fonts.
  // Host linker aliases become independent regular files in the installation.
  await copyRegular(await realpath(source), path.join(root, destination));
}

async function fileDigest(file) {
  const digest = createHash("sha256");
  for await (const bytes of createReadStream(file)) digest.update(bytes);
  return digest.digest("hex");
}

async function inventory(directory, prefix = "") {
  const result = [];
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const relative = prefix ? `${prefix}/${entry.name}` : entry.name;
    if (entry.name !== entry.name.normalize("NFC") || /[\x00-\x1f\x7f\\]/u.test(entry.name)) throw new Error("invalid runtime inventory path");
    if (entry.isDirectory()) result.push(...await inventory(path.join(directory, entry.name), relative));
    else if (entry.isFile()) {
      const absolute = path.join(directory, entry.name);
      const info = await lstat(absolute);
      if (info.nlink !== 1) throw new Error("runtime inventory hard link");
      result.push({ path: relative, sha256: await fileDigest(absolute) });
    } else throw new Error("invalid runtime inventory entry");
  }
  return result.sort((a, b) => Buffer.compare(Buffer.from(a.path), Buffer.from(b.path)));
}

try {
  await mkdir(root);
  const binary = path.join(work, "render-sandbox");
  run("bash", ["tools/harness/backend/build-go-artifact.sh"], { env: { ...process.env, BUILD_OUTPUT: binary, BUILD_PACKAGE: "./cmd/render-sandbox", BUILD_LABEL: "build renderer confinement launcher", CARTULARY_TEST_TARGET: "build-report-renderer" } });
  await copyRegular(binary, path.join(root, "bin/cartulary-render-sandbox"));

  // Stage only executable package inputs before pnpm inspects its workspace.
  // New deploy mode derives a frozen dedicated lock; legacy resolution is not used.
  const workspace = path.join(work, "workspace");
  await mkdir(path.join(workspace, "packages/report-renderer"), { recursive: true });
  await copyTree("packages/report-renderer", path.join(workspace, "packages/report-renderer"), true);
  // Deploy rewrites the copied manifest to the frozen dependency projection.
  // Only this disposable workspace is writable; final installed files are not.
  await chmod(path.join(workspace, "packages/report-renderer/package.json"), 0o644);
  await copyRegular("pnpm-lock.yaml", path.join(workspace, "pnpm-lock.yaml"));
  await writeFile(path.join(workspace, "package.json"), JSON.stringify({ private: true, packageManager: `pnpm@${pins.pnpm_version}` }));
  await writeFile(path.join(workspace, "pnpm-workspace.yaml"), "packages:\n  - packages/report-renderer\npatchedDependencies:\n  mermaid@12.2.0: packages/report-renderer/patches/mermaid@12.2.0.patch\n");
  const deployed = path.join(work, "deployed");
  run(process.env.PNPM, ["--filter", "@cartulary/report-renderer", "deploy", "--prod", "--frozen-lockfile", "--ignore-scripts", "--config.inject-workspace-packages=true", "--config.node-linker=hoisted", `--store-dir=${path.join(repository, ".pnpm-store")}`, deployed], {
    cwd: workspace, env: { ...process.env, CI: "true", PATH: `${path.dirname(process.execPath)}:${process.env.PATH}` },
  });
  for (const directory of ["src", "node_modules", "patches"]) {
    await copyTree(path.join(deployed, directory), path.join(root, "runtime/renderer", directory));
  }
  // Package-manager deployment metadata contains disposable source paths and
  // is not a runtime input. Retain the authored package identity and full source
  // lock separately; runtime resolution uses only the copied installed tree.
  await copyRegular("packages/report-renderer/package.json", path.join(root, "runtime/renderer/package.json"));
  const installedPackages = {};
  for (const [name, pin] of Object.entries(pins.report_renderer)) {
    const identity = JSON.parse(await readFile(path.join(root, "runtime/renderer/node_modules", name, "package.json"), "utf8"));
    if (identity.name !== name || identity.version !== pin) throw new Error("installed renderer package identity mismatch");
    installedPackages[name] = identity.version;
  }
  // Check packaged module loading on trusted build inputs before publication.
  // This does not execute a stage or establish native sandbox qualification.
  run(process.execPath, ["--input-type=module", "--eval", "await import(process.argv[1])", pathToFileURL(path.join(root, "runtime/renderer/src/stage.ts")).href]);
  await copyHostFile(process.execPath, "runtime/node/bin/node");
  await copyRegular("pnpm-lock.yaml", path.join(root, "runtime/workspace-lock.yaml"));

  const require = createRequire(path.join(repository, "packages/report-renderer/package.json"));
  // The pinned Playwright headless profile uses its separate headless shell.
  // chromium.executablePath() selects full Chromium, which adds profile IPC
  // and a different rendering implementation. Use the pinned exported registry
  // at build time; the installed renderer receives only the bound binary path.
  const { registry: { registry } } = require("playwright-core/lib/coreBundle");
  const chromiumBinary = registry.findExecutable("chromium-headless-shell").executablePathOrDie("javascript");
  await copyTree(path.dirname(chromiumBinary), path.join(root, "runtime/browser"));
  const browserPath = `runtime/browser/${path.basename(chromiumBinary)}`;
  const bubblewrap = run("which", ["bwrap"]).trim();
  if (run(bubblewrap, ["--version"]).trim() !== `bubblewrap ${pins.bubblewrap_version}`) throw new Error("renderer bubblewrap pin mismatch");
  await copyHostFile(bubblewrap, "bin/bwrap");

  // The installed profile declares a closed local font set and configuration.
  for (const family of ["Sans", "Serif", "Mono"]) {
    for (const style of ["Regular", "Bold", "Italic", "BoldItalic"]) {
      const name = `Liberation${family}-${style}.ttf`;
      await copyHostFile(`/usr/share/fonts/truetype/liberation/${name}`, `runtime/fonts/${name}`);
    }
  }
  await writeFile(path.join(root, "runtime/fonts/fonts.conf"), '<?xml version="1.0"?><!DOCTYPE fontconfig SYSTEM "fonts.dtd"><fontconfig><dir>/runtime/fonts</dir><cachedir>/work/font-cache</cachedir><alias><family>sans-serif</family><prefer><family>Liberation Sans</family></prefer></alias><alias><family>serif</family><prefer><family>Liberation Serif</family></prefer></alias><alias><family>monospace</family><prefer><family>Liberation Mono</family></prefer></alias></fontconfig>\n');

  // Read ELF metadata; never execute ldd or the selected binary to discover its
  // dependencies. Bind the complete copied library closure by content below.
  const copiedLibraries = new Set();
  for (const file of nativeFiles) {
    const programHeaders = run("readelf", ["--program-headers", "--wide", file]);
    const interpreter = programHeaders.match(/Requesting program interpreter: ([^\]]+)\]/u)?.[1];
    if (interpreter && interpreter !== "/lib64/ld-linux-x86-64.so.2") throw new Error("unsupported renderer ELF interpreter");
    if (interpreter && !copiedLibraries.has(interpreter)) {
      copiedLibraries.add(interpreter);
      await copyHostFile(interpreter, interpreter.slice(1));
    }
    const dynamic = run("readelf", ["--dynamic", "--wide", file]);
    for (const match of dynamic.matchAll(/\(NEEDED\).*Shared library: \[([^\]]+)\]/gu)) {
      const library = match[1];
      if (!/^[A-Za-z0-9_.+-]+$/u.test(library)) throw new Error("invalid ELF dependency");
      if (copiedLibraries.has(library)) continue;
      let source;
      for (const directory of ["/lib/x86_64-linux-gnu", "/usr/lib/x86_64-linux-gnu", "/lib64"]) {
        try { source = await realpath(path.join(directory, library)); break; }
        catch (error) { if (error.code !== "ENOENT") throw error; }
      }
      if (!source) throw new Error(`missing renderer ELF dependency: ${library}`);
      copiedLibraries.add(library);
      await copyRegular(source, path.join(root, "lib/x86_64-linux-gnu", library));
    }
  }
  for (const mountpoint of ["proc", "dev", "tmp", "work", "diagnostics", "etc"]) await mkdir(path.join(root, mountpoint), { recursive: true });
  await writeFile(path.join(root, "etc/hosts"), "127.0.0.1 localhost\n::1 localhost\n");
  await writeFile(path.join(root, "etc/nsswitch.conf"), "hosts: files\n");
  const files = await inventory(root);
  const contentManifestDigest = (entries) => sha(JSON.stringify(entries));
  const manifest = {
    schema_id: "cartulary.reporting_runtime_installation.v1",
    platform: "linux/amd64",
    node_version: pins.node_version,
    package_manager: `pnpm@${pins.pnpm_version}`,
    lockfile_sha256: await fileDigest("pnpm-lock.yaml"),
    package_store_digest: contentManifestDigest(files.filter((file) => file.path.startsWith("runtime/renderer/")).map((file) => ({ ...file, path: file.path.slice("runtime/renderer/".length) }))),
    font_manifest_sha256: contentManifestDigest(files.filter((file) => file.path.startsWith("runtime/fonts/")).map((file) => ({ ...file, path: file.path.slice("runtime/fonts/".length) }))),
    chromium_executable_path: browserPath,
    chromium_executable_sha256: await fileDigest(path.join(root, browserPath)),
    chromium_version: run(chromiumBinary, ["--version"]).trim(),
    bubblewrap_version: pins.bubblewrap_version,
    packages: installedPackages,
    files,
  };
  const bytes = JSON.stringify(manifest);
  const identity = sha(bytes);
  await writeFile(path.join(work, "manifest.json"), bytes);
  const destination = path.join(output, identity);
  try { await rename(root, path.join(output, `${identity}.root`)); }
  catch (error) {
    if (!["EEXIST", "ENOTEMPTY"].includes(error.code)) throw error;
    if (JSON.stringify(await inventory(path.join(output, `${identity}.root`))) !== JSON.stringify(files)) throw new Error("existing runtime installation was altered");
  }
  // Immutable content-addressed roots; the receipt is the sole mutable pointer.
  try { await writeFile(`${destination}.json`, bytes, { flag: "wx" }); }
  catch (error) { if (error.code !== "EEXIST" || await readFile(`${destination}.json`, "utf8") !== bytes) throw error; }
  const receipt = path.join(work, "installation.json");
  await writeFile(receipt, JSON.stringify({ schema_id: "cartulary.reporting_runtime_receipt.v1", manifest_sha256: identity, root: `${identity}.root`, manifest: `${identity}.json` }) + "\n");
  await rename(receipt, path.join(output, "installation.json"));
  process.stdout.write(`report renderer installation: ${identity}\n`);
} finally {
  await rm(work, { recursive: true, force: true });
}
