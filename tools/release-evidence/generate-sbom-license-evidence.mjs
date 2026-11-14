#!/usr/bin/env node
import {
  chmodSync,
  existsSync,
  lstatSync,
  mkdtempSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import crypto from "node:crypto";
import { licenseEvidence } from "./license-inputs.mjs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import {
  canonicalJSONString,
  parseStrictJSON,
  publicExitCodeForFailure,
  semanticJSONDigest,
  validateSchemaSync,
} from "../harness/contract/index.mjs";
import { independentGoModules, validateIndependentGoModule } from "../harness/test-catalog/go-modules.mjs";
import { borrowRuntime } from "../harness/runtime/session-client.mjs";
import { runPrivateCapturedProcess } from "../harness/runtime/private-child-process.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile, removePrivateTree } from "../harness/runtime/secure-local-files.mjs";
import { CommandFailure, reportCommandFailure } from "../harness/runtime/command-failure.mjs";
import { enforcePrivateProcessUmask } from "../harness/runtime/private-process.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../..");
const cyclonedxSpecVersion = "1.7";
const firstPartyGoModule = "github.com/JochiRaider/cartulary";
const firstPartyNpmScope = "@cartulary/";
const goTestPackagePatterns = Object.freeze([
  "./cmd/...",
  "./db/...",
  "./internal/...",
  "./tools/...",
]);

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function rel(file) {
  return path.relative(repoRoot, file).replaceAll(path.sep, "/") || ".";
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function writePrivateJSON(file, value) {
  atomicLocalFile(file, `${canonicalJSONString(value)}\n`);
}

function ensureDir(dir) {
  privateDirectory(dir);
}

function ensureSafeDirectory(directory) {
  const absolute = path.resolve(directory);
  const parsed = path.parse(absolute);
  let current = parsed.root;
  for (const segment of absolute.slice(parsed.root.length).split(path.sep).filter(Boolean)) {
    current = path.join(current, segment);
    if (!existsSync(current)) mkdirSync(current, { mode: 0o755 });
    const info = lstatSync(current);
    if (!info.isDirectory() || info.isSymbolicLink()) {
      throw new Error(`release inventory directory ancestry is unsafe: ${current}`);
    }
  }
}

function deterministicUUID(digest) {
  const hex = digest.replace(/^sha256:/u, "").slice(0, 32).split("");
  hex[12] = "5";
  hex[16] = "8";
  return `${hex.slice(0, 8).join("")}-${hex.slice(8, 12).join("")}-${hex.slice(12, 16).join("")}-${hex.slice(16, 20).join("")}-${hex.slice(20).join("")}`;
}

function publishCanonicalPair(outputs) {
  const transactionID = crypto.randomUUID();
  const staged = [];
  const published = [];
  const backups = [];
  try {
    for (const { destination, value } of outputs) {
      ensureSafeDirectory(path.dirname(destination));
      if (existsSync(destination)) {
        const info = lstatSync(destination);
        if (!info.isFile() || info.isSymbolicLink() || info.nlink !== 1) {
          throw new Error(`release inventory destination is not a regular file: ${destination}`);
        }
      }
      const stage = path.join(
        path.dirname(destination),
        `.cartulary-release-stage-${path.basename(destination)}-${transactionID}`,
      );
      writeFileSync(stage, `${canonicalJSONString(value)}\n`, { flag: "wx", mode: 0o644 });
      chmodSync(stage, 0o644);
      staged.push({ destination, stage });
    }
    for (const entry of staged) {
      if (existsSync(entry.destination)) {
        const backup = `${entry.destination}.cartulary-release-backup-${transactionID}`;
        renameSync(entry.destination, backup);
        backups.push({ destination: entry.destination, backup });
      }
      renameSync(entry.stage, entry.destination);
      published.push(entry.destination);
    }
    for (const { backup } of backups) {
      rmSync(backup, { force: true });
    }
  } catch (error) {
    for (const destination of [...published].reverse()) {
      rmSync(destination, { force: true });
    }
    for (const { destination, backup } of [...backups].reverse()) {
      if (existsSync(backup)) renameSync(backup, destination);
    }
    for (const { stage } of staged) {
      rmSync(stage, { force: true });
    }
    throw error;
  }
}

export async function runInventoryCommand(ctx, label, command, args, options = {}) {
  const result = await runPrivateCapturedProcess(command, args, {
    repoRoot, runRoot: ctx.runRoot, cwd: options.cwd ?? repoRoot,
    env: { ...ctx.environment, ...(options.env ?? {}) }, dependencies: [], signal: options.signal,
  });
  let primary;
  try {
    if (result.signal) throw new CommandFailure(`${label} interrupted`, { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" });
    if (result.status !== 0) {
      const diagnostic = result.stderr.trim().split(/\r?\n/u).slice(-8).join(" | ").slice(-4096);
      throw new CommandFailure(`${label} failed with exit code ${result.status}${diagnostic ? `: ${diagnostic}` : ""}`,
        { failure_class: "harness", failure_reason: "tool_diagnostic_failure" });
    }
    // The capture's stdout property is a diagnostic tail, never a complete scanner stream.
    const stdout = new TextDecoder("utf-8", { fatal: true }).decode(readLocalFile(result.stdoutPath, { maximum: 256 * 1024 * 1024 }));
    return { stdout };
  } catch (error) { primary = error; throw error; }
  finally {
    try { result.cleanup(); }
    catch (error) { if (!primary) throw error; (primary.cleanupFailures ??= []).push(error); }
  }
}

async function runOutput(ctx, label, command, args, options = {}) {
  const result = await runInventoryCommand(ctx, label, command, args, options);
  return result.stdout.trim();
}

function goReadOnlyEnv() {
  return Object.fromEntries(
    Object.entries({
      GOFLAGS: "-mod=readonly",
      GOWORK: "off",
      GOCACHE: process.env.GO_CACHE_DIR,
      GOMODCACHE: process.env.GO_MOD_CACHE_DIR,
      GOTMPDIR: process.env.GO_TMP_DIR,
    }).filter(([, value]) => value),
  );
}

function goDownloadModuleDir(ctx, root, directory) {
  const dir = path.join(ctx.workspace, "go-download-modules", directory === "." ? "application" : directory);
  ensureDir(dir);
  atomicLocalFile(path.join(dir, "go.mod"), readFileSync(path.join(root, directory, "go.mod")));
  atomicLocalFile(path.join(dir, "go.sum"), readFileSync(path.join(root, directory, "go.sum")));
  return dir;
}

function parseJSONStream(text) {
  const objects = [];
  let depth = 0;
  let start = -1;
  let inString = false;
  let escaped = false;
  for (let index = 0; index < text.length; index += 1) {
    const char = text[index];
    if (start === -1) {
      if (/[\t\n\r ]/u.test(char)) continue;
      if (char !== "{") throw scannerOutputFailure("JSON stream", index);
      start = index;
      depth = 1;
      continue;
    }
    if (inString) {
      if (escaped) {
        escaped = false;
      } else if (char === "\\") {
        escaped = true;
      } else if (char === '"') {
        inString = false;
      }
      continue;
    }
    if (char === '"') {
      inString = true;
      continue;
    }
    if (char === "{") {
      depth += 1;
      continue;
    }
    if (char === "}") {
      depth -= 1;
      if (depth === 0 && start >= 0) {
        try { objects.push(parseStrictJSON(text.slice(start, index + 1))); }
        catch { throw scannerOutputFailure("JSON stream", start); }
        start = -1;
      }
    }
  }
  if (start !== -1) throw scannerOutputFailure("JSON stream", start);
  return objects;
}

function scannerOutputFailure(format, offset) {
  return new CommandFailure(`invalid inventory ${format} at offset ${offset}`, {
    failure_class: "harness", failure_reason: "tool_diagnostic_failure",
  });
}

function licenseReviewFlags(licenseExpression, rawLicenseMetadata) {
  const text = String(licenseExpression ?? rawLicenseMetadata ?? "").toUpperCase();
  const flags = [];
  if (/\b(AGPL|GPL|LGPL)\b/.test(text)) {
    flags.push("copyleft_review");
    flags.push("legal_review");
  }
  if (/\b(MPL|EPL|CDDL|CPL)\b/.test(text)) {
    flags.push("weak_copyleft_review");
  }
  if (/\b(BUSL|PROPRIETARY|UNLICENSED|COMMERCIAL)\b/.test(text)) {
    flags.push("commercial_or_legal_review");
    flags.push("legal_review");
  }
  if (/\b(APACHE|BSD|MIT|ISC)\b/.test(text)) {
    flags.push("notice_or_attribution_review");
  }
  return [...new Set(flags)].sort();
}

function npmPurl(name, version) {
  const encodedName = name.startsWith("@")
    ? `%40${encodeURIComponent(name.slice(1).split("/")[0])}/${encodeURIComponent(name.split("/")[1])}`
    : encodeURIComponent(name);
  return version ? `pkg:npm/${encodedName}@${encodeURIComponent(version)}` : `pkg:npm/${encodedName}`;
}

function goPurl(modulePath, version) {
  return version ? `pkg:golang/${modulePath}@${version}` : `pkg:golang/${modulePath}`;
}

function dockerPurl(image) {
  const [name, version = "latest"] = image.split(":");
  return `pkg:docker/${name}@${encodeURIComponent(version)}`;
}

function bomRef(prefix, name, version) {
  return `${prefix}:${name}@${version || "unversioned"}`;
}

function componentProperty(name, value) {
  return { name, value: String(value) };
}

function makeCycloneDxBom({
  name,
  version,
  bomRefValue,
  components,
  dependencies,
  tools,
  semanticInputDigest,
  completeness,
}) {
  const bom = {
    bomFormat: "CycloneDX",
    specVersion: cyclonedxSpecVersion,
    serialNumber: `urn:uuid:${deterministicUUID(semanticInputDigest)}`,
    version: 1,
    metadata: {
      tools: {
        components: tools,
      },
      component: {
        type: "application",
        "bom-ref": bomRefValue,
        name,
        version,
        purl: goPurl(firstPartyGoModule, version),
        properties: [
          componentProperty("cartulary:first_party", true),
          componentProperty("cartulary:license_evidence", "LICENSE"),
          componentProperty("cartulary:semantic_input_digest", semanticInputDigest),
        ],
      },
    },
    components,
    dependencies,
    compositions: [
      {
        aggregate: completeness,
        assemblies: [bomRefValue],
      },
    ],
  };
  return bom;
}

function readPackageManifests() {
  const files = [path.join(repoRoot, "package.json")];
  const appsWeb = path.join(repoRoot, "apps", "web", "package.json");
  if (existsSync(appsWeb)) {
    files.push(appsWeb);
  }
  const packagesDir = path.join(repoRoot, "packages");
  if (existsSync(packagesDir)) {
    for (const name of readdirSync(packagesDir).sort()) {
      const packageJSON = path.join(packagesDir, name, "package.json");
      if (existsSync(packageJSON)) {
        files.push(packageJSON);
      }
    }
  }
  return files.map((file) => ({ file, manifest: readJSON(file) }));
}

function directNodeDependencyIndex(packageManifests) {
  const index = new Map();
  const sections = [
    ["dependencies", "runtime dependency"],
    ["devDependencies", "development-only dependency"],
    ["optionalDependencies", "optional dependency"],
    ["peerDependencies", "peer dependency"],
  ];
  for (const { file, manifest } of packageManifests) {
    for (const [section, classification] of sections) {
      for (const [name, specifier] of Object.entries(manifest[section] ?? {})) {
        if (!index.has(name)) {
          index.set(name, []);
        }
        index.get(name).push({
          file,
          section,
          specifier,
          classification,
          workspace: String(specifier).startsWith("workspace:"),
        });
      }
    }
  }
  return index;
}

function classifyNodeDependency(name, directEntries) {
  if (name.startsWith(firstPartyNpmScope)) {
    return "first-party";
  }
  if (!directEntries || directEntries.length === 0) {
    return "transitive dependency";
  }
  if (directEntries.some((entry) => entry.section === "dependencies")) {
    return "runtime dependency";
  }
  if (directEntries.some((entry) => entry.section === "optionalDependencies")) {
    return "optional dependency";
  }
  if (directEntries.some((entry) => entry.section === "devDependencies")) {
    if (/^(vitest|jsdom|@testing-library\/|@playwright\/)/.test(name)) {
      return "test dependency";
    }
    if (/^(typescript|vite|@vitejs\/|@biomejs\/|@types\/)/.test(name)) {
      return "build-time dependency";
    }
    return "development-only dependency";
  }
  if (directEntries.some((entry) => entry.section === "peerDependencies")) {
    return "peer dependency";
  }
  return "unknown";
}

function flattenPnpmProjects(projects, directIndex) {
  const components = new Map();
  const dependencies = new Map();

  function addDependency(parentRef, childRef) {
    if (!dependencies.has(parentRef)) {
      dependencies.set(parentRef, new Set());
    }
    dependencies.get(parentRef).add(childRef);
  }

  function visitPackage(pkg, parentRef, section) {
    if (!pkg?.name) {
      return null;
    }
    const sourcePath = pkg.path?.startsWith(repoRoot) ? pkg.path : null;
    const firstParty =
      pkg.name.startsWith(firstPartyNpmScope) ||
      (sourcePath !== null &&
        !sourcePath.includes(`${path.sep}node_modules${path.sep}`) &&
        (sourcePath === repoRoot ||
          sourcePath.startsWith(path.join(repoRoot, "apps") + path.sep) ||
          sourcePath.startsWith(path.join(repoRoot, "packages") + path.sep)));
    const version = String(pkg.version ?? "");
    const ref = firstParty
      ? bomRef("npm-workspace", pkg.name, version || "0.0.0")
      : bomRef("npm", pkg.name, version);
    const directEntries = directIndex.get(pkg.name) ?? [];
    const classification = firstParty ? "first-party" : classifyNodeDependency(pkg.name, directEntries);
    const evidence = [
      "pnpm list --recursive --depth Infinity --json",
      directEntries.length > 0 ? "workspace package.json" : null,
      pkg.resolved ? "pnpm registry resolution" : null,
      "pnpm-lock.yaml",
    ];
    const existing = components.get(ref);
    const record = {
      ref,
      ecosystem: firstParty ? "node-workspace" : "npm",
      name: pkg.name,
      version: version || "0.0.0",
      classification,
      direct: directEntries.length > 0 || firstParty,
      transitive: directEntries.length === 0 && !firstParty,
      optional: section === "optionalDependencies" || directEntries.some((entry) => entry.section === "optionalDependencies"),
      resolved: pkg.resolved ?? null,
      path: sourcePath,
      evidence: [...new Set(evidence.filter(Boolean))].sort(),
    };
    components.set(ref, existing ? { ...existing, ...record, evidence: [...new Set([...existing.evidence, ...record.evidence])].sort() } : record);
    if (parentRef) {
      addDependency(parentRef, ref);
    }
    for (const [childSection, values] of Object.entries({
      dependencies: pkg.dependencies,
      devDependencies: pkg.devDependencies,
      optionalDependencies: pkg.optionalDependencies,
    })) {
      for (const child of Object.values(values ?? {})) {
        visitPackage(child, ref, childSection);
      }
    }
    return ref;
  }

  for (const project of projects) {
    const ref = visitPackage(project, null, "workspace");
    for (const [section, values] of Object.entries({
      dependencies: project.dependencies,
      devDependencies: project.devDependencies,
      optionalDependencies: project.optionalDependencies,
    })) {
      for (const child of Object.values(values ?? {})) {
        visitPackage(child, ref, section);
      }
    }
  }
  return { components, dependencies };
}

function parseNpmLicenseOutput(text) {
  const byNameVersion = new Map();
  const raw = text.trim() ? JSON.parse(text) : {};
  for (const [license, entries] of Object.entries(raw)) {
    for (const entry of entries ?? []) {
      for (const version of entry.versions ?? []) {
        byNameVersion.set(`${entry.name}@${version}`, {
          license,
          raw: entry.license ?? license,
          paths: entry.paths ?? [],
          homepage: entry.homepage ?? null,
          author: entry.author ?? null,
        });
      }
    }
  }
  return byNameVersion;
}

function parseGoModuleGraph(text) {
  const edges = [];
  let offset = 0;
  for (const raw of text.split("\n")) {
    const line = raw.replace(/\r$/u, "");
    if (!/^[\t ]*$/u.test(line)) {
      const edge = /^[\t ]*([^\s\x00-\x1f\x7f]+)[\t ]+([^\s\x00-\x1f\x7f]+)[\t ]*$/u.exec(line);
      if (!edge) throw scannerOutputFailure("Go module graph", offset);
      edges.push([edge[1], edge[2]]);
    }
    offset += raw.length + 1;
  }
  return edges;
}

function parseGoModuleSet(text) {
  const modules = new Set();
  for (const line of text.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (trimmed) {
      modules.add(trimmed);
    }
  }
  return modules;
}

function goTestDependencyListArgs(patterns = goTestPackagePatterns) {
  return [
    "list",
    "-deps",
    "-test",
    "-f",
    "{{if and (not .Standard) .Module}}{{.Module.Path}}@{{.Module.Version}}{{end}}",
    ...patterns,
  ];
}

function parseContainerImages() {
  const images = new Map();
  const compose = path.join(repoRoot, "docker-compose.dev.yml");
  if (existsSync(compose)) {
    const text = readFileSync(compose, "utf8");
    for (const match of text.matchAll(/^\s*image:\s*["']?([^"'\s]+)["']?\s*$/gm)) {
      images.set(match[1], {
        image: match[1],
        source: "docker-compose.dev.yml",
        classification: "local-service or container-image dependency",
      });
    }
  }
  for (const file of [
    path.join(repoRoot, "internal", "testutil", "pgtest", "pgtest.go"),
    path.join(repoRoot, "internal", "testutil", "s3test", "s3test.go"),
  ]) {
    if (!existsSync(file)) {
      continue;
    }
    const text = readFileSync(file, "utf8");
    for (const match of text.matchAll(/=\s*"((?:postgres|minio\/minio):[^"]+)"/g)) {
      const existing = images.get(match[1]);
      images.set(match[1], {
        image: match[1],
        source: existing ? `${existing.source}, ${rel(file)}` : rel(file),
        classification: "local-service or container-image dependency",
      });
    }
  }
  return [...images.values()].sort((a, b) => compareASCII(a.image, b.image));
}

function makeDependencyBomComponents(records, licenseIndex) {
  return [...records.values()]
    .sort((a, b) => compareASCII(`${a.ecosystem}:${a.name}:${a.version}`, `${b.ecosystem}:${b.name}:${b.version}`))
    .map((record) => {
      const license = licenseIndex.get(`${record.ecosystem}:${record.name}@${record.version}`);
      const component = {
        type:
          record.ecosystem === "container"
            ? "container"
            : record.ecosystem === "node-workspace"
              ? "application"
              : record.ecosystem === "font"
                ? "file"
                : "library",
        "bom-ref": record.ref,
        name: record.name,
        version: record.version,
        purl:
          record.ecosystem === "npm"
            ? npmPurl(record.name, record.version)
            : record.ecosystem === "go"
              ? goPurl(record.name, record.version)
              : record.ecosystem === "container"
                ? dockerPurl(record.name)
                : undefined,
        properties: [
          componentProperty("cartulary:ecosystem", record.ecosystem),
          componentProperty("cartulary:classification", record.classification),
          componentProperty("cartulary:direct", record.direct),
          componentProperty("cartulary:transitive", record.transitive),
          componentProperty("cartulary:optional", record.optional),
          componentProperty("cartulary:evidence", record.evidence.join("; ")),
        ],
      };
      if (record.resolved) {
        component.externalReferences = [{ type: "distribution", url: record.resolved }];
      }
      if (license?.license_expression) {
        component.licenses = [{ expression: license.license_expression }];
      }
      return Object.fromEntries(Object.entries(component).filter(([, value]) => value !== undefined));
    });
}

function dependencyGraphToCycloneDx(dependencies) {
  return [...dependencies.entries()]
    .sort(([a], [b]) => compareASCII(a, b))
    .map(([ref, children]) => ({
      ref,
      dependsOn: [...children].sort(),
    }));
}

function loadFontEvidence() {
  const manifestFile = path.join(repoRoot, "apps", "web", "public", "assets", "fonts", "FONT_MANIFEST.json");
  if (!existsSync(manifestFile)) {
    return { manifestFile, manifest: null, records: new Map(), licenseEntries: [] };
  }
  const manifest = readJSON(manifestFile);
  const records = new Map();
  const licenseEntries = [];
  for (const family of manifest.families ?? []) {
    const version = family.version_or_ref ?? "unversioned";
    const ref = bomRef("font", family.family, version);
    const fontDir = path.join(repoRoot, "apps", "web", "public", "assets", "fonts", family.directory);
    const license = licenseEvidence(fontDir, { metadata: family.license, kind: "font" });
    const record = {
      ref,
      ecosystem: "font",
      name: family.family,
      version,
      classification: family.active_by_default ? "runtime vendored font" : "optional vendored font",
      direct: true,
      transitive: false,
      optional: family.active_by_default !== true,
      path: fontDir,
      resolved: family.source,
      evidence: [
        "apps/web/public/assets/fonts/FONT_MANIFEST.json",
        `apps/web/public/assets/fonts/${family.directory}/SOURCE.json`,
      ],
    };
    records.set(ref, record);
    licenseEntries.push({
      package: family.family,
      version,
      ecosystem: "font",
      classification: record.classification,
      direct: true,
      transitive: false,
      license_expression: family.license,
      raw_license_metadata: family.license,
      evidence_source: license.evidenceDigest ? "vendored font license file" : "missing",
      evidence_digest: license.evidenceDigest,
      issue_flags: license.evidenceDigest ? [] : ["license_text_missing"],
      review_flags: licenseReviewFlags(family.license, family.license),
      font_files: family.files ?? [],
    });
  }
  return { manifestFile, manifest, records, licenseEntries };
}

function buildToolsMetadata(ctx) {
  const tools = [
    { type: "application", name: "cartulary-release-inventory-generator", version: "2" },
  ];
  if (ctx.toolVersions.node) {
    tools.push({ type: "application", name: "node", version: ctx.toolVersions.node });
  }
  if (ctx.toolVersions.pnpm) {
    tools.push({ type: "application", name: "pnpm", version: ctx.toolVersions.pnpm });
  }
  if (ctx.toolVersions.go) {
    tools.push({ type: "application", name: "go", version: ctx.toolVersions.go });
  }
  return tools;
}

export async function collectGoEvidence(ctx, root = repoRoot) {
  const independent = independentGoModules(root).sort((a, b) => compareASCII(a.root, b.root));
  for (const module of independent) validateIndependentGoModule(root, module.root);
  const selections = [
    { directory: ".", patterns: goTestPackagePatterns },
    ...independent.map(module => ({ directory: module.root, patterns: ["./..."] })),
  ];
  const go = ctx.config.go;
  const goVersion = await runOutput(ctx, "go version", go, ["version"], { cwd: root, env: goReadOnlyEnv() });
  ctx.toolVersions.go = goVersion;
  const downloads = new Map(), modules = new Map(), testSet = new Set(), edges = new Set(), mainPaths = new Set();
  for (const [ordinal, selection] of selections.entries()) {
    const { directory, patterns } = selection;
    const options = { cwd: path.join(root, directory), env: goReadOnlyEnv() };
    const label = `go module ${ordinal} ${directory}`;
    const download = await runInventoryCommand(ctx, `${label} download`, go, ["mod", "download", "-json", "all"], {
      ...options, cwd: goDownloadModuleDir(ctx, root, directory),
    });
    const listed = await runInventoryCommand(ctx, `${label} modules`, go, ["list", "-m", "-json", "all"], options);
    const graph = await runInventoryCommand(ctx, `${label} graph`, go, ["mod", "graph"], options);
    const test = await runInventoryCommand(ctx, `${label} test deps`, go, goTestDependencyListArgs(patterns), options);
    const moduleTestSet = parseGoModuleSet(test.stdout);
    for (const entry of parseJSONStream(download.stdout)) downloads.set(`${entry.Path}@${entry.Version}`, entry);
    for (const entry of parseJSONStream(listed.stdout)) {
      if (entry.Main) { mainPaths.add(entry.Path); continue; }
      const key = `${entry.Path}@${entry.Version}`, previous = modules.get(key);
      const inventoryEvidence = new Set(previous?.inventoryEvidence ?? []);
      for (const item of [path.posix.join(directory, "go.mod"), path.posix.join(directory, "go.sum"),
        `go -C ${directory} list -m -json all`, `go -C ${directory} mod graph`]) inventoryEvidence.add(item);
      if (moduleTestSet.has(key)) inventoryEvidence.add(`go -C ${directory} list -deps -test ${patterns.join(" ")}`);
      modules.set(key, {
        ...entry,
        Indirect: entry.Indirect === true && (previous === undefined || previous.Indirect === true),
        inventoryEvidence: [...inventoryEvidence].sort(compareASCII),
      });
    }
    for (const key of moduleTestSet) testSet.add(key);
    for (const edge of parseGoModuleGraph(graph.stdout)) edges.add(JSON.stringify(edge));
  }
  const runtime = await runInventoryCommand(ctx, "go list runtime deps", go, [
    "list",
    "-deps",
    "-f",
    "{{if and (not .Standard) .Module}}{{.Module.Path}}@{{.Module.Version}}{{end}}",
    "./cmd/server",
    "./cmd/migrate",
    "./cmd/operator",
  ], { cwd: root, env: goReadOnlyEnv() });
  const runtimeSet = parseGoModuleSet(runtime.stdout);
  const graphEdges = [...edges].sort(compareASCII).map(edge => JSON.parse(edge));
  return { modules: [...modules.entries()].sort(([a], [b]) => compareASCII(a, b)).map(([, value]) => value),
    downloads, runtimeSet, testSet, graphEdges, mainPaths };
}

export function goInventoryDependencies(evidence, records) {
  const dependencies = new Map();
  for (const [from, to] of evidence.graphEdges) {
    const fromRef = evidence.mainPaths.has(from) ? "cartulary:go-backend" : bomRef("go", from.slice(0, from.lastIndexOf("@")), from.slice(from.lastIndexOf("@") + 1));
    const toRef = bomRef("go", to.slice(0, to.lastIndexOf("@")), to.slice(to.lastIndexOf("@") + 1));
    if ((fromRef !== "cartulary:go-backend" && !records.has(fromRef)) || !records.has(toRef)) continue;
    if (!dependencies.has(fromRef)) dependencies.set(fromRef, new Set());
    dependencies.get(fromRef).add(toRef);
  }
  return dependencies;
}

async function collectNodeEvidence(ctx, packageManifests) {
  const pnpm = ctx.config.pnpm;
  const list = await runInventoryCommand(ctx, "pnpm list recursive json", pnpm, ["list", "--recursive", "--depth", "Infinity", "--json"]);
  const licenses = await runInventoryCommand(ctx, "pnpm licenses recursive json long", pnpm, [
    "licenses",
    "list",
    "--recursive",
    "--json",
    "--long",
  ]);
  const directIndex = directNodeDependencyIndex(packageManifests);
  const projects = JSON.parse(list.stdout);
  const licenseIndex = parseNpmLicenseOutput(licenses.stdout);
  return { directIndex, projects, licenseIndex };
}

function collectContainerEvidence() {
  const images = parseContainerImages();
  return { images };
}

export async function withInventoryWorkspace(work, environment = process.env) {
  const runID = environment.CARTULARY_TEST_RUN_ID;
  if (!/^[A-Za-z0-9_.-]+$/u.test(runID ?? "")) throw new Error("release inventory requires Make run identity");
  const runRoot = path.resolve(repoRoot, environment.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results", runID);
  let runtime;
  try { runtime = borrowRuntime({ repoRoot, runRoot, environment }); }
  catch (error) {
    throw new CommandFailure("release inventory requires a bound private runtime", { failure_class: "artifact", failure_reason: "artifact_error" }, { cause: error });
  }
  let workspace, primary;
  try {
    workspace = mkdtempSync(runtime.producerPath("release-inventory-"));
    privateDirectory(workspace);
    return await work({ workspace, runRoot,
      environment: { ...environment, ...runtime.producerEnvironment() }, toolVersions: {},
      config: { go: environment.GO || "go", pnpm: environment.PNPM || "pnpm", node: environment.NODE_BIN || process.execPath },
    });
  } catch (error) { primary = error; throw error; }
  finally {
    const failures = [];
    if (workspace && primary?.processSettled !== false) {
      try { removePrivateTree(workspace); } catch (error) { failures.push(error); }
    }
    try { runtime.close(); } catch (error) { failures.push(error); }
    if (failures.length) {
      if (primary) (primary.cleanupFailures ??= []).push(...failures);
      else throw new CommandFailure("release inventory workspace cleanup failed", { failure_class: "harness", failure_reason: "cleanup_error" }, { cause: new AggregateError(failures) });
    }
  }
}

async function produceInventory(ctx, canonicalSbom, canonicalLicenseReport) {
  const outputDir = ctx.workspace;
  const packageManifests = readPackageManifests();
  const containerRecords = new Map();
  const licenseReport = [];

  const nodeVersion = await runOutput(ctx, "node version", ctx.config.node, ["--version"]);
  const pnpmVersion = await runOutput(ctx, "pnpm version", ctx.config.pnpm, ["--version"]);
  ctx.toolVersions.node = nodeVersion;
  ctx.toolVersions.pnpm = pnpmVersion;
  const goEvidence = await collectGoEvidence(ctx);
  const nodeEvidence = await collectNodeEvidence(ctx, packageManifests);
  const containerEvidence = collectContainerEvidence();
  const fontEvidence = loadFontEvidence();

  const firstPartyLicenseDigest = licenseEvidence(repoRoot, { kind: "first_party" }).evidenceDigest;
  if (!firstPartyLicenseDigest) throw new Error("first-party license input missing");

  const goRecords = new Map();
  const goDownloadByKey = goEvidence.downloads;
  for (const module of goEvidence.modules) {
    const key = `${module.Path}@${module.Version}`;
    const direct = module.Indirect !== true;
    const classification = goEvidence.runtimeSet.has(key)
      ? "runtime dependency"
      : goEvidence.testSet.has(key)
        ? "test dependency"
        : direct
          ? "build-time dependency"
          : "transitive dependency";
    const downloadInfo = goDownloadByKey.get(key) ?? {};
    const sourceDir = module.Dir ?? downloadInfo.Dir ?? null;
    const ref = bomRef("go", module.Path, module.Version);
    const record = {
      ref,
      ecosystem: "go",
      name: module.Path,
      version: module.Version,
      classification,
      direct,
      transitive: !direct,
      optional: false,
      path: sourceDir,
      evidence: [...module.inventoryEvidence,
        goEvidence.runtimeSet.has(key) ? "go list -deps ./cmd/server ./cmd/migrate ./cmd/operator" : null,
      ].filter(Boolean).sort(compareASCII),
    };
    goRecords.set(ref, record);
    const evidence = licenseEvidence(sourceDir);
    const issues = [...evidence.issues];
    const licenseEntry = {
      package: module.Path,
      version: module.Version,
      ecosystem: "go",
      classification,
      direct,
      transitive: !direct,
      license_expression: null,
      raw_license_metadata: null,
      evidence_source: evidence.evidenceSource,
      evidence_digest: evidence.evidenceDigest,
      issue_flags: issues,
      review_flags: [],
    };
    licenseReport.push(licenseEntry);
  }

  const goDependencies = goInventoryDependencies(goEvidence, goRecords);

  const { components: nodeRecords, dependencies: nodeDependencies } = flattenPnpmProjects(
    nodeEvidence.projects,
    nodeEvidence.directIndex,
  );
  for (const record of nodeRecords.values()) {
    const licenseMetadata = nodeEvidence.licenseIndex.get(`${record.name}@${record.version}`);
    const packageJSON = record.path ? path.join(record.path, "package.json") : null;
    let rawLicenseMetadata = licenseMetadata?.raw ?? null;
    if (!rawLicenseMetadata && packageJSON && existsSync(packageJSON)) {
      rawLicenseMetadata = readJSON(packageJSON).license ?? null;
    }
    let licenseExpression = typeof rawLicenseMetadata === "string" ? rawLicenseMetadata : null;
    let evidence;
    if (record.ecosystem === "node-workspace") {
      rawLicenseMetadata = rawLicenseMetadata ?? "Apache-2.0";
      licenseExpression = licenseExpression ?? "Apache-2.0";
      evidence = {
        evidenceSource: "repository root LICENSE",
        evidenceDigest: firstPartyLicenseDigest,
        issues: [],
      };
    } else {
      evidence = licenseEvidence(record.path, { metadata: rawLicenseMetadata });
    }
    const issues = [...evidence.issues];
    if (!licenseExpression) {
      issues.push("license_expression_missing");
    }
    const reviewFlags = licenseReviewFlags(licenseExpression, rawLicenseMetadata);
    const licenseEntry = {
      package: record.name,
      version: record.version,
      ecosystem: record.ecosystem,
      classification: record.classification,
      direct: record.direct,
      transitive: record.transitive,
      license_expression: licenseExpression,
      raw_license_metadata: rawLicenseMetadata,
      evidence_source: evidence.evidenceSource,
      evidence_digest: evidence.evidenceDigest,
      issue_flags: [...new Set(issues)].sort(),
      review_flags: reviewFlags,
    };
    licenseReport.push(licenseEntry);
  }

  for (const image of containerEvidence.images) {
    const ref = bomRef("container", image.image, "observed-tag");
    const record = {
      ref,
      ecosystem: "container",
      name: image.image,
      version: image.image.includes(":") ? image.image.split(":").slice(1).join(":") : "latest",
      classification: image.classification,
      direct: true,
      transitive: false,
      optional: false,
      path: null,
      evidence: [image.source, "container image reference scan"],
    };
    containerRecords.set(ref, record);
    const issueFlags = ["container_image_sbom_incomplete", "license_text_missing", "license_metadata_missing"];
    licenseReport.push({
      package: image.image,
      version: record.version,
      ecosystem: "container",
      classification: record.classification,
      direct: true,
      transitive: false,
      license_expression: null,
      raw_license_metadata: null,
      evidence_source: "container image reference",
      evidence_digest: null,
      issue_flags: issueFlags,
      review_flags: ["legal_review"],
    });
  }

  for (const entry of fontEvidence.licenseEntries) {
    licenseReport.push(entry);
  }

  for (const entry of licenseReport) {
    entry.issue_flags = [...new Set(entry.issue_flags)].sort(compareASCII);
    entry.review_flags = [...new Set(entry.review_flags)].sort(compareASCII);
    if (entry.font_files) {
      entry.font_files = [...entry.font_files].sort((left, right) => compareASCII(left.path, right.path));
    }
  }
  licenseReport.sort((a, b) =>
    compareASCII(
      `${a.ecosystem}:${a.package}:${a.version}`,
      `${b.ecosystem}:${b.package}:${b.version}`,
    ),
  );
  const licenseIndexForBom = new Map(
    licenseReport.map((entry) => [`${entry.ecosystem}:${entry.package}@${entry.version}`, entry]),
  );
  const tools = buildToolsMetadata(ctx);
  const cartularyRef = "cartulary:application";

  const goComponents = makeDependencyBomComponents(goRecords, licenseIndexForBom);
  const goCycloneDependencies = dependencyGraphToCycloneDx(goDependencies);
  const nodeComponents = makeDependencyBomComponents(nodeRecords, licenseIndexForBom);
  const nodeCycloneDependencies = dependencyGraphToCycloneDx(nodeDependencies);
  const shippedRecords = new Map();
  for (const record of [...goRecords.values(), ...nodeRecords.values()]) {
    if (record.classification === "runtime dependency" || record.classification === "first-party") {
      shippedRecords.set(record.ref, record);
    }
  }
  for (const record of fontEvidence.records.values()) {
    shippedRecords.set(record.ref, record);
  }
  for (const record of containerRecords.values()) {
    shippedRecords.set(record.ref, record);
  }
  const combinedDependencies = new Map([[cartularyRef, new Set(shippedRecords.keys())]]);
  const combinedComponents = makeDependencyBomComponents(shippedRecords, licenseIndexForBom);
  const combinedCycloneDependencies = dependencyGraphToCycloneDx(combinedDependencies);
  const semanticInputDigest = semanticJSONDigest({
    schema_id: "cartulary.release_inventory_semantic_inputs.v1",
    tools,
    go_components: goComponents,
    go_dependencies: goCycloneDependencies,
    node_components: nodeComponents,
    node_dependencies: nodeCycloneDependencies,
    combined_components: combinedComponents,
    combined_dependencies: combinedCycloneDependencies,
    license_entries: licenseReport,
  });
  const goBom = makeCycloneDxBom({
    name: "Cartulary Go backend",
    version: "0.0.0",
    bomRefValue: "cartulary:go-backend",
    components: goComponents,
    dependencies: goCycloneDependencies,
    tools,
    semanticInputDigest,
    completeness: "complete",
  });
  const nodeBom = makeCycloneDxBom({
    name: "Cartulary pnpm workspace",
    version: "0.0.0",
    bomRefValue: "cartulary:node-workspace",
    components: nodeComponents,
    dependencies: nodeCycloneDependencies,
    tools,
    semanticInputDigest,
    completeness: "complete",
  });
  const combinedBom = makeCycloneDxBom({
    name: "Cartulary shipped application artifact set",
    version: "0.0.0",
    bomRefValue: cartularyRef,
    components: combinedComponents,
    dependencies: combinedCycloneDependencies,
    tools,
    semanticInputDigest,
    completeness: containerRecords.size > 0 ? "incomplete" : "complete",
  });

  const sbomGo = path.join(outputDir, "sbom.go.cyclonedx.json");
  const sbomNode = path.join(outputDir, "sbom.node.cyclonedx.json");
  const sbomCombined = path.join(outputDir, "sbom.cyclonedx.json");
  const licenseReportPath = path.join(outputDir, "license-report.json");
  const licenseReportDocument = {
    schema_id: "cartulary.license_report.v2",
    semantic_input_digest: semanticInputDigest,
    entries: licenseReport,
  };
  validateSchemaSync(licenseReportDocument.schema_id, licenseReportDocument);
  writePrivateJSON(sbomGo, goBom);
  writePrivateJSON(sbomNode, nodeBom);
  writePrivateJSON(sbomCombined, combinedBom);
  writePrivateJSON(licenseReportPath, licenseReportDocument);

  const validator = path.join(repoRoot, "tools", "release-evidence", "validate-release-sbom.mjs");
  await runInventoryCommand(ctx, "validate combined sbom", ctx.config.node, [validator, sbomCombined]);
  await runInventoryCommand(ctx, "validate go sbom", ctx.config.node, [validator, sbomGo]);
  await runInventoryCommand(ctx, "validate node sbom", ctx.config.node, [validator, sbomNode]);
  publishCanonicalPair([
    { destination: canonicalLicenseReport, value: licenseReportDocument },
    { destination: canonicalSbom, value: combinedBom },
  ]);
  await runInventoryCommand(ctx, "validate canonical sbom", ctx.config.node, [validator, canonicalSbom]);

  console.log(`Release inventory semantic input: ${semanticInputDigest}`);
  console.log(`Canonical SBOM: ${rel(canonicalSbom)}`);
  console.log(`Canonical license report: ${rel(canonicalLicenseReport)}`);
}

async function main() {
  const releaseDir = path.resolve(process.env.RELEASE_ARTIFACT_DIR ?? path.join(repoRoot, ".cartulary", "release-artifacts"));
  const canonicalSbom = path.resolve(process.env.SBOM_ARTIFACT ?? path.join(releaseDir, "sbom.cyclonedx.json"));
  const canonicalLicenseReport = path.resolve(
    process.env.LICENSE_REPORT_ARTIFACT ?? path.join(releaseDir, "license-report.json"),
  );
  if (
    canonicalSbom !== path.join(releaseDir, "sbom.cyclonedx.json") ||
    canonicalLicenseReport !== path.join(releaseDir, "license-report.json")
  ) {
    throw new Error("release inventory artifacts must use the canonical paired paths beneath RELEASE_ARTIFACT_DIR");
  }
  await withInventoryWorkspace(ctx => produceInventory(ctx, canonicalSbom, canonicalLicenseReport));
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    enforcePrivateProcessUmask();
    await main();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(`SBOM/license evidence generation failed: ${message}`);
    if (error.cleanupFailures?.length) console.error(`Inventory cleanup also failed (${error.cleanupFailures.length} failures)`);
    const failure = reportCommandFailure(repoRoot, error, { failure_class: "artifact", failure_reason: "artifact_error" });
    process.exitCode = publicExitCodeForFailure(failure);
  }
}

export {
  goTestDependencyListArgs,
  licenseReviewFlags,
  makeCycloneDxBom,
  parseGoModuleGraph,
  parseJSONStream,
};
