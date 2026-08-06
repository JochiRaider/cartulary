#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import {
  copyFileSync,
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import crypto from "node:crypto";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../..");
const cyclonedxSpecVersion = "1.7";
const cyclonedxGomodSpecVersion = "1.6";
const firstPartyGoModule = "github.com/JochiRaider/cartulary";
const firstPartyNpmScope = "@cartulary/";

function rel(file) {
  return path.relative(repoRoot, file).replaceAll(path.sep, "/") || ".";
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function writeJSON(file, value) {
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function ensureDir(dir) {
  mkdirSync(dir, { recursive: true });
}

function sha256File(file) {
  return crypto.createHash("sha256").update(readFileSync(file)).digest("hex");
}

function nowRunID() {
  const compact = new Date().toISOString().replaceAll("-", "").replace("T", "-").replaceAll(":", "").replace(/\.\d+Z$/, "Z");
  return `${compact}-p${process.pid}`;
}

function shellQuote(value) {
  if (/^[A-Za-z0-9_@%+=:,./-]+$/.test(value)) {
    return value;
  }
  return `'${value.replaceAll("'", "'\"'\"'")}'`;
}

function commandText(command, args) {
  return [command, ...args].map(shellQuote).join(" ");
}

function rawName(label) {
  return label.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "command";
}

function runCommand(ctx, label, command, args, options = {}) {
  const cwd = options.cwd ?? repoRoot;
  const stdoutFile = path.join(ctx.rawDir, `${rawName(label)}.stdout`);
  const stderrFile = path.join(ctx.rawDir, `${rawName(label)}.stderr`);
  const startedAt = new Date().toISOString();
  const result = spawnSync(command, args, {
    cwd,
    env: { ...process.env, ...(options.env ?? {}) },
    encoding: "utf8",
    maxBuffer: 256 * 1024 * 1024,
  });
  const stdout = result.stdout ?? "";
  const stderr = result.stderr ?? "";
  writeFileSync(stdoutFile, stdout);
  writeFileSync(stderrFile, stderr);
  const endedAt = new Date().toISOString();
  const record = {
    label,
    command: commandText(command, args),
    cwd: rel(cwd),
    started_at: startedAt,
    ended_at: endedAt,
    exit_code: result.status,
    stdout: rel(stdoutFile),
    stderr: rel(stderrFile),
  };
  ctx.commands.push(record);
  if (result.error) {
    record.error = result.error.message;
  }
  if ((options.required ?? true) && result.status !== 0) {
    throw new Error(`${label} failed with exit code ${result.status}; see ${rel(stderrFile)}`);
  }
  return { ...record, stdout, stderr };
}

function runOutput(ctx, label, command, args, options = {}) {
  const result = runCommand(ctx, label, command, args, options);
  return result.stdout.trim();
}

function goReadOnlyEnv() {
  return Object.fromEntries(
    Object.entries({
      GOFLAGS: "-mod=readonly",
      GOCACHE: process.env.GO_CACHE_DIR,
      GOMODCACHE: process.env.GO_MOD_CACHE_DIR,
      GOTMPDIR: process.env.GO_TMP_DIR,
    }).filter(([, value]) => value),
  );
}

function goDownloadModuleDir(ctx) {
  const dir = path.join(ctx.rawDir, "go-download-module");
  ensureDir(dir);
  copyFileSync(path.join(repoRoot, "go.mod"), path.join(dir, "go.mod"));
  copyFileSync(path.join(repoRoot, "go.sum"), path.join(dir, "go.sum"));
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
      if (depth === 0) {
        start = index;
      }
      depth += 1;
      continue;
    }
    if (char === "}") {
      depth -= 1;
      if (depth === 0 && start >= 0) {
        objects.push(JSON.parse(text.slice(start, index + 1)));
        start = -1;
      }
    }
  }
  return objects;
}

function normalizePackageName(value) {
  return value.replace(/^@/, "").replace(/[^A-Za-z0-9._-]+/g, "_").replace(/^_+|_+$/g, "") || "unnamed";
}

export function deterministicLicenseFilename(ecosystem, name, version) {
  const safeVersion = normalizePackageName(version || "unversioned");
  return `${normalizePackageName(ecosystem)}__${normalizePackageName(name)}__${safeVersion}__LICENSE.txt`;
}

function findLicenseFiles(dir) {
  if (!dir || !existsSync(dir)) {
    return [];
  }
  const candidates = [];
  for (const name of readdirSync(dir)) {
    const full = path.join(dir, name);
    if (!statSync(full).isFile()) {
      continue;
    }
    if (/^(license|licence|copying|notice)([._-].*)?$/i.test(name)) {
      candidates.push(full);
    }
  }
  return candidates.sort((a, b) => a.localeCompare(b));
}

function copyLicenseEvidence(ctx, dependency, sourceDir, rawLicenseMetadata) {
  const licenseFiles = findLicenseFiles(sourceDir);
  const issues = [];
  let evidencePath = null;
  let evidenceSource = null;
  if (licenseFiles.length > 0) {
    const out = path.join(
      ctx.licensesDir,
      deterministicLicenseFilename(dependency.ecosystem, dependency.name, dependency.version),
    );
    const chunks = [];
    for (const file of licenseFiles) {
      chunks.push(`===== ${rel(file)} =====\n${readFileSync(file, "utf8").trimEnd()}\n`);
    }
    writeFileSync(out, `${chunks.join("\n")}\n`);
    evidencePath = rel(out);
    evidenceSource = "package-distributed license file";
  } else if (rawLicenseMetadata) {
    evidenceSource = "package metadata license field";
    issues.push("license_text_missing");
  } else {
    evidenceSource = "missing";
    issues.push("license_text_missing");
    issues.push("license_metadata_missing");
  }
  return {
    evidencePath,
    evidenceSource,
    issues,
  };
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

function npmPackageNameFromSpecifier(specifier) {
  if (specifier.startsWith("node:") || specifier.startsWith(".") || specifier.startsWith("/")) {
    return null;
  }
  const parts = specifier.split("/");
  return specifier.startsWith("@") ? `${parts[0]}/${parts[1]}` : parts[0];
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

function makeCycloneDxBom({ name, version, bomRefValue, components, dependencies, tools, timestamp, completeness }) {
  const bom = {
    bomFormat: "CycloneDX",
    specVersion: cyclonedxSpecVersion,
    serialNumber: `urn:uuid:${crypto.randomUUID()}`,
    version: 1,
    metadata: {
      timestamp,
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

function collectNodeImportScan() {
  const roots = [path.join(repoRoot, "apps"), path.join(repoRoot, "packages")];
  const specifiers = new Map();
  const importPattern = /(?:^|\s)(?:import\s+(?:[^'"]+\s+from\s+)?|export\s+[^'"]+\s+from\s+|import\()\s*["']([^"']+)["']/gm;
  function walk(dir) {
    if (!existsSync(dir)) {
      return;
    }
    for (const name of readdirSync(dir)) {
      const full = path.join(dir, name);
      const stat = statSync(full);
      if (stat.isDirectory()) {
        if (name === "node_modules" || name === "dist" || name === "coverage") {
          continue;
        }
        walk(full);
        continue;
      }
      if (!/\.(ts|tsx|js|jsx|mjs|cjs)$/.test(name)) {
        continue;
      }
      const text = readFileSync(full, "utf8");
      for (const match of text.matchAll(importPattern)) {
        const specifier = match[1];
        const pkg = npmPackageNameFromSpecifier(specifier);
        if (!pkg) {
          continue;
        }
        if (!specifiers.has(pkg)) {
          specifiers.set(pkg, []);
        }
        specifiers.get(pkg).push(rel(full));
      }
    }
  }
  for (const root of roots) {
    walk(root);
  }
  return specifiers;
}

function parseGoModuleGraph(text) {
  const edges = [];
  for (const line of text.split(/\r?\n/)) {
    if (!line.trim()) {
      continue;
    }
    const [from, to] = line.trim().split(/\s+/);
    if (from && to) {
      edges.push([from, to]);
    }
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

function parseMakeToolPins() {
  const makefile = readFileSync(path.join(repoRoot, "Makefile"), "utf8");
  const pins = [];
  for (const [name, regex, classification] of [
    ["sqlc", /^SQLC_TOOL\s*:=\s*(\S+)/m, "code-generation tool dependency"],
    ["goose", /^GOOSE_TOOL\s*:=\s*(\S+)/m, "build-time dependency"],
    ["staticcheck", /^STATICCHECK_TOOL\s*:=\s*(\S+)/m, "development-only dependency"],
    ["govulncheck", /^GOVULNCHECK_TOOL\s*:=\s*(\S+)/m, "development-only dependency"],
    ["gosec", /^GOSEC_TOOL\s*:=\s*(\S+)/m, "development-only dependency"],
    ["cyclonedx-gomod", /^CYCLONEDX_GOMOD_TOOL\s*:=\s*(\S+)/m, "code-generation tool dependency"],
    ["syft", /^SYFT_TOOL\s*:=\s*(\S+)/m, "code-generation tool dependency"],
  ]) {
    const match = makefile.match(regex);
    if (match) {
      const spec = match[1];
      const at = spec.lastIndexOf("@");
      pins.push({
        name,
        module: at > 0 ? spec.slice(0, at) : spec,
        version: at > 0 ? spec.slice(at + 1) : "",
        spec,
        classification,
      });
    }
  }
  return pins;
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
  return [...images.values()].sort((a, b) => a.image.localeCompare(b.image));
}

function makeDependencyBomComponents(records, licenseIndex) {
  return [...records.values()]
    .sort((a, b) => `${a.ecosystem}:${a.name}:${a.version}`.localeCompare(`${b.ecosystem}:${b.name}:${b.version}`))
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
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([ref, children]) => ({
      ref,
      dependsOn: [...children].sort(),
    }));
}

function addProvenance(provenance, key, evidence) {
  if (!provenance[key]) {
    provenance[key] = [];
  }
  provenance[key].push(evidence);
}

function artifactRows(files) {
  return files
    .filter((file) => existsSync(file))
    .map((file) => ({
      path: rel(file),
      sha256: sha256File(file),
      bytes: statSync(file).size,
    }))
    .sort((a, b) => a.path.localeCompare(b.path));
}

function loadFontEvidence(ctx) {
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
    const licenseFile =
      readdirSync(fontDir)
        .filter((name) => /^(?:OFL|LICENSE)(?:\.[^.]+)?$/iu.test(name))
        .sort()[0] ?? null;
    const licenseEvidencePath = path.join(ctx.licensesDir, deterministicLicenseFilename("font", family.family, version));
    if (licenseFile) {
      copyFileSync(path.join(fontDir, licenseFile), licenseEvidencePath);
    }
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
      evidence_source: licenseFile ? "vendored font license file" : "missing",
      evidence_path: licenseFile ? rel(licenseEvidencePath) : null,
      issue_flags: licenseFile ? [] : ["license_text_missing"],
      review_flags: licenseReviewFlags(family.license, family.license),
      font_files: family.files ?? [],
    });
  }
  return { manifestFile, manifest, records, licenseEntries };
}

function writeMarkdownReport(file, lines) {
  writeFileSync(file, `${lines.join("\n")}\n`);
}

function commandsMarkdown(ctx, { timestamp, commit, status, nodeVersion, pnpmVersion }) {
  return [
    "# Commands Used",
    "",
    `- Working directory: ${repoRoot}`,
    `- Generated at: ${timestamp}`,
    `- Commit: ${commit || "unavailable"}`,
    `- Dirty worktree at generation: ${status ? "yes" : "no"}`,
    `- Node: ${nodeVersion || "unavailable"}`,
    `- pnpm: ${pnpmVersion || "unavailable"}`,
    `- CycloneDX primary spec version: ${cyclonedxSpecVersion}`,
    `- cyclonedx-gomod raw Go app spec version: ${cyclonedxGomodSpecVersion}`,
    "- Network status: generator does not pull container images implicitly; package-manager commands may use configured caches/registries.",
    "",
    "## Commands",
    "",
    ...ctx.commands.map(
      (command) =>
        `- \`${command.command}\`\n  - cwd: ${command.cwd}\n  - exit: ${command.exit_code}\n  - stdout: ${command.stdout}\n  - stderr: ${command.stderr}`,
    ),
  ];
}

function buildToolsMetadata(ctx) {
  const tools = [
    { type: "application", name: "cartulary-sbom-license-evidence-generator", version: "1" },
  ];
  if (ctx.toolVersions.cdxgen) {
    tools.push({ type: "application", name: "@cyclonedx/cdxgen", version: ctx.toolVersions.cdxgen });
  }
  if (ctx.toolVersions.cyclonedxGomod) {
    tools.push({ type: "application", name: "cyclonedx-gomod", version: ctx.toolVersions.cyclonedxGomod });
  }
  if (ctx.toolVersions.syft) {
    tools.push({ type: "application", name: "syft", version: ctx.toolVersions.syft });
  }
  return tools;
}

function collectGoEvidence(ctx) {
  const go = ctx.config.go;
  const goVersion = runOutput(ctx, "go version", go, ["version"], { env: goReadOnlyEnv() });
  ctx.toolVersions.go = goVersion;
  const download = runCommand(ctx, "go mod download json all", go, ["mod", "download", "-json", "all"], {
    cwd: goDownloadModuleDir(ctx),
    env: goReadOnlyEnv(),
  });
  const goList = runCommand(ctx, "go list modules json", go, ["list", "-m", "-json", "all"], {
    env: goReadOnlyEnv(),
  });
  const goGraph = runCommand(ctx, "go mod graph", go, ["mod", "graph"], { env: goReadOnlyEnv() });
  const runtime = runCommand(ctx, "go list runtime deps", go, [
    "list",
    "-deps",
    "-f",
    "{{if and (not .Standard) .Module}}{{.Module.Path}}@{{.Module.Version}}{{end}}",
    "./cmd/server",
    "./cmd/migrate",
    "./cmd/operator",
  ], { env: goReadOnlyEnv() });
  const test = runCommand(ctx, "go list test deps", go, [
    "list",
    "-deps",
    "-test",
    "-f",
    "{{if and (not .Standard) .Module}}{{.Module.Path}}@{{.Module.Version}}{{end}}",
    "./...",
  ], { env: goReadOnlyEnv() });

  for (const [name, main] of [
    ["server", "cmd/server"],
    ["migrate", "cmd/migrate"],
    ["operator", "cmd/operator"],
  ]) {
    runCommand(
      ctx,
      `cyclonedx gomod ${name}`,
      ctx.config.cyclonedxGomod,
      [
        "app",
        "-json",
        "-output-version",
        cyclonedxGomodSpecVersion,
        "-licenses",
        "-packages",
        "-main",
        main,
        "-output",
        path.join(ctx.rawDir, `sbom.go.${name}.raw.cyclonedx.json`),
        repoRoot,
      ],
      { env: goReadOnlyEnv(), required: false },
    );
  }

  const downloads = new Map(parseJSONStream(download.stdout).map((entry) => [`${entry.Path}@${entry.Version}`, entry]));
  const modules = parseJSONStream(goList.stdout);
  const runtimeSet = parseGoModuleSet(runtime.stdout);
  const testSet = parseGoModuleSet(test.stdout);
  const graphEdges = parseGoModuleGraph(goGraph.stdout);
  const toolPins = parseMakeToolPins();
  return { modules, downloads, runtimeSet, testSet, graphEdges, toolPins };
}

function collectNodeEvidence(ctx, packageManifests) {
  const pnpm = ctx.config.pnpm;
  const cdxgenVersion = runOutput(ctx, "cdxgen version", pnpm, ["exec", "cdxgen", "--version"], { required: false });
  const versionMatch = cdxgenVersion.match(/CycloneDX Generator\s+([^\s]+)/);
  ctx.toolVersions.cdxgen = versionMatch?.[1] ?? cdxgenVersion.split(/\s+/).find(Boolean) ?? "";
  const list = runCommand(ctx, "pnpm list recursive json", pnpm, ["list", "--recursive", "--depth", "Infinity", "--json"]);
  const licenses = runCommand(ctx, "pnpm licenses recursive json long", pnpm, [
    "licenses",
    "list",
    "--recursive",
    "--json",
    "--long",
  ]);
  runCommand(
    ctx,
    "cdxgen node raw",
    pnpm,
    [
      "exec",
      "cdxgen",
      "-t",
      "js",
      "--spec-version",
      cyclonedxSpecVersion,
      "--json-pretty",
      "--no-install-deps",
      "--validate",
      "-o",
      path.join(ctx.rawDir, "sbom.node.raw.cyclonedx.json"),
      repoRoot,
    ],
    { required: false },
  );
  const directIndex = directNodeDependencyIndex(packageManifests);
  const projects = JSON.parse(list.stdout);
  const licenseIndex = parseNpmLicenseOutput(licenses.stdout);
  const importScan = collectNodeImportScan();
  return { directIndex, projects, licenseIndex, importScan };
}

function collectContainerEvidence(ctx) {
  const images = parseContainerImages();
  const docker = spawnSync("docker", ["--version"], { encoding: "utf8" });
  const dockerAvailable = docker.status === 0;
  const scanned = [];
  const skipped = [];
  if (ctx.config.syft && existsSync(ctx.config.syft)) {
    const version = runOutput(ctx, "syft version", ctx.config.syft, ["version"], { required: false });
    ctx.toolVersions.syft = version.split(/\r?\n/).find((line) => line.includes("Version:"))?.replace(/^.*Version:\s*/, "") ?? version;
  }
  for (const entry of images) {
    if (!dockerAvailable || !ctx.config.syft || !existsSync(ctx.config.syft)) {
      skipped.push({
        image: entry.image,
        reason: dockerAvailable ? "syft binary unavailable" : "docker unavailable",
      });
      continue;
    }
    const inspect = spawnSync("docker", ["image", "inspect", entry.image], {
      cwd: repoRoot,
      encoding: "utf8",
      maxBuffer: 32 * 1024 * 1024,
    });
    if (inspect.status !== 0) {
      skipped.push({
        image: entry.image,
        reason: "image not present locally; generator does not pull images implicitly",
      });
      continue;
    }
    const safe = normalizePackageName(entry.image);
    const output = path.join(ctx.outputDir, `sbom.container.${safe}.cyclonedx.json`);
    const result = runCommand(
      ctx,
      `syft ${entry.image}`,
      ctx.config.syft,
      [entry.image, "-o", `cyclonedx-json=${output}`],
      { required: false },
    );
    if (result.exit_code === 0 && existsSync(output)) {
      scanned.push({ image: entry.image, output });
    } else {
      skipped.push({ image: entry.image, reason: "syft scan failed" });
    }
  }
  return { images, scanned, skipped };
}

function main() {
  const releaseDir = path.resolve(process.env.RELEASE_ARTIFACT_DIR ?? path.join(repoRoot, ".cartulary", "release-artifacts"));
  const canonicalSbom = path.resolve(process.env.SBOM_ARTIFACT ?? path.join(releaseDir, "sbom.cyclonedx.json"));
  const canonicalLicenseReport = path.resolve(
    process.env.LICENSE_REPORT_ARTIFACT ?? path.join(releaseDir, "license-report.json"),
  );
  const outputDir = path.join(releaseDir, "sbom", process.env.CARTULARY_SBOM_RUN_ID || nowRunID());
  const rawDir = path.join(outputDir, "raw");
  const licensesDir = path.join(outputDir, "licenses");
  rmSync(outputDir, { recursive: true, force: true });
  ensureDir(rawDir);
  ensureDir(licensesDir);

  const ctx = {
    outputDir,
    rawDir,
    licensesDir,
    commands: [],
    toolVersions: {},
    config: {
      go: process.env.GO || "go",
      pnpm: process.env.PNPM || "pnpm",
      node: process.env.NODE_BIN || process.execPath,
      cyclonedxGomod: process.env.CYCLONEDX_GOMOD_BIN || "cyclonedx-gomod",
      syft: process.env.SYFT_BIN || "syft",
    },
  };

  const timestamp = new Date().toISOString();
  const packageManifests = readPackageManifests();
  const provenance = {};
  const dependencyRecords = new Map();
  const licenseReport = [];
  const unresolvedIssues = [];

  const commit = runOutput(ctx, "git rev-parse HEAD", "git", ["rev-parse", "HEAD"], { required: false });
  const status = runOutput(ctx, "git status short", "git", ["status", "--short"], { required: false });
  const nodeVersion = runOutput(ctx, "node version", ctx.config.node, ["--version"], { required: false });
  const pnpmVersion = runOutput(ctx, "pnpm version", ctx.config.pnpm, ["--version"], { required: false });
  ctx.toolVersions.node = nodeVersion;
  ctx.toolVersions.pnpm = pnpmVersion;
  const gomodVersion = runOutput(ctx, "cyclonedx gomod version", ctx.config.cyclonedxGomod, ["version"], {
    required: false,
  });
  ctx.toolVersions.cyclonedxGomod = gomodVersion.match(/Version:\s*(\S+)/)?.[1] ?? gomodVersion.split(/\s+/).find(Boolean) ?? "";

  const goEvidence = collectGoEvidence(ctx);
  const nodeEvidence = collectNodeEvidence(ctx, packageManifests);
  const containerEvidence = collectContainerEvidence(ctx);
  const fontEvidence = loadFontEvidence(ctx);

  const rootLicense = path.join(repoRoot, "LICENSE");
  const firstPartyLicenseOut = path.join(licensesDir, deterministicLicenseFilename("first-party", "cartulary", "0.0.0"));
  copyFileSync(rootLicense, firstPartyLicenseOut);

  const goRecords = new Map();
  const goDependencies = new Map();
  const goDownloadByKey = goEvidence.downloads;
  for (const module of goEvidence.modules) {
    if (module.Main) {
      continue;
    }
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
      evidence: [
        "go.mod",
        "go.sum",
        "go list -m -json all",
        goEvidence.runtimeSet.has(key) ? "go list -deps ./cmd/server ./cmd/migrate ./cmd/operator" : null,
        goEvidence.testSet.has(key) ? "go list -deps -test ./..." : null,
        "go mod graph",
      ].filter(Boolean),
    };
    goRecords.set(ref, record);
    dependencyRecords.set(ref, record);
    addProvenance(provenance, ref, {
      dependency: key,
      evidence: record.evidence,
      source_path: sourceDir ? rel(sourceDir) : null,
    });
    const evidence = copyLicenseEvidence(ctx, record, sourceDir, null);
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
      evidence_path: evidence.evidencePath,
      issue_flags: issues,
      review_flags: [],
    };
    licenseReport.push(licenseEntry);
    if (issues.length > 0) {
      unresolvedIssues.push(`${module.Path}@${module.Version}: ${issues.join(", ")}`);
    }
  }

  for (const [from, to] of goEvidence.graphEdges) {
    const fromRef = from.startsWith(firstPartyGoModule) ? "cartulary:go-module" : bomRef("go", from.slice(0, from.lastIndexOf("@")), from.slice(from.lastIndexOf("@") + 1));
    const toRef = bomRef("go", to.slice(0, to.lastIndexOf("@")), to.slice(to.lastIndexOf("@") + 1));
    if (!goDependencies.has(fromRef)) {
      goDependencies.set(fromRef, new Set());
    }
    if (goRecords.has(toRef)) {
      goDependencies.get(fromRef).add(toRef);
    }
  }

  for (const pin of goEvidence.toolPins) {
    const ref = bomRef("go-tool", pin.module, pin.version);
    const record = {
      ref,
      ecosystem: "go",
      name: pin.module,
      version: pin.version,
      classification: pin.classification,
      direct: true,
      transitive: false,
      optional: false,
      path: null,
      evidence: ["Makefile pinned tool variable"],
    };
    dependencyRecords.set(ref, record);
    addProvenance(provenance, ref, {
      dependency: pin.spec,
      evidence: record.evidence,
      source_path: "Makefile",
    });
  }

  const { components: nodeRecords, dependencies: nodeDependencies } = flattenPnpmProjects(
    nodeEvidence.projects,
    nodeEvidence.directIndex,
  );
  const importScan = nodeEvidence.importScan;
  for (const record of nodeRecords.values()) {
    dependencyRecords.set(record.ref, record);
    addProvenance(provenance, record.ref, {
      dependency: `${record.name}@${record.version}`,
      evidence: record.evidence,
      source_path: record.path ? rel(record.path) : null,
      import_scan_paths: importScan.get(record.name) ?? [],
    });
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
        evidencePath: rel(firstPartyLicenseOut),
        issues: [],
      };
    } else {
      evidence = copyLicenseEvidence(ctx, record, record.path, rawLicenseMetadata);
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
      evidence_path: evidence.evidencePath,
      issue_flags: [...new Set(issues)].sort(),
      review_flags: reviewFlags,
    };
    licenseReport.push(licenseEntry);
    if (licenseEntry.issue_flags.length > 0) {
      unresolvedIssues.push(`${record.name}@${record.version}: ${licenseEntry.issue_flags.join(", ")}`);
    }
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
    dependencyRecords.set(ref, record);
    addProvenance(provenance, ref, {
      dependency: image.image,
      evidence: record.evidence,
      source_path: image.source,
      scan_status: containerEvidence.scanned.find((scan) => scan.image === image.image)
        ? "scanned"
        : "not_scanned",
    });
    const issueFlags = containerEvidence.scanned.find((scan) => scan.image === image.image)
      ? ["container_license_review_required"]
      : ["container_image_sbom_incomplete", "license_text_missing", "license_metadata_missing"];
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
      evidence_path: null,
      issue_flags: issueFlags,
      review_flags: ["legal_review"],
    });
    unresolvedIssues.push(`${image.image}: ${issueFlags.join(", ")}`);
  }

  for (const record of fontEvidence.records.values()) {
    dependencyRecords.set(record.ref, record);
    addProvenance(provenance, record.ref, {
      dependency: `${record.name}@${record.version}`,
      evidence: record.evidence,
      source_path: record.path ? rel(record.path) : null,
      font_manifest: fontEvidence.manifestFile ? rel(fontEvidence.manifestFile) : null,
    });
  }
  for (const entry of fontEvidence.licenseEntries) {
    licenseReport.push(entry);
    if (entry.issue_flags.length > 0) {
      unresolvedIssues.push(`${entry.package}@${entry.version}: ${entry.issue_flags.join(", ")}`);
    }
  }

  const declaredNodeNames = new Set(nodeEvidence.directIndex.keys());
  const usedUndeclared = [...importScan.keys()].filter(
    (name) => !declaredNodeNames.has(name) && !name.startsWith(firstPartyNpmScope),
  );
  const declaredUnused = [...declaredNodeNames].filter((name) => !importScan.has(name) && !name.startsWith(firstPartyNpmScope));

  const licenseIndexForBom = new Map(
    licenseReport.map((entry) => [`${entry.ecosystem}:${entry.package}@${entry.version}`, entry]),
  );
  const tools = buildToolsMetadata(ctx);
  const cartularyRef = "cartulary:application";

  const goComponents = makeDependencyBomComponents(goRecords, licenseIndexForBom);
  const goBom = makeCycloneDxBom({
    name: "Cartulary Go backend",
    version: "0.0.0",
    bomRefValue: "cartulary:go-backend",
    components: goComponents,
    dependencies: dependencyGraphToCycloneDx(goDependencies),
    tools,
    timestamp,
    completeness: "complete",
  });

  const nodeComponents = makeDependencyBomComponents(nodeRecords, licenseIndexForBom);
  const nodeBom = makeCycloneDxBom({
    name: "Cartulary pnpm workspace",
    version: "0.0.0",
    bomRefValue: "cartulary:node-workspace",
    components: nodeComponents,
    dependencies: dependencyGraphToCycloneDx(nodeDependencies),
    tools,
    timestamp,
    completeness: "complete",
  });

  const shippedRecords = new Map();
  for (const record of [...goRecords.values(), ...nodeRecords.values()]) {
    if (record.classification === "runtime dependency" || record.classification === "first-party") {
      shippedRecords.set(record.ref, record);
    }
  }
  for (const record of fontEvidence.records.values()) {
    shippedRecords.set(record.ref, record);
  }
  const combinedDependencies = new Map([[cartularyRef, new Set(shippedRecords.keys())]]);
  const combinedBom = makeCycloneDxBom({
    name: "Cartulary shipped application artifact set",
    version: "0.0.0",
    bomRefValue: cartularyRef,
    components: makeDependencyBomComponents(shippedRecords, licenseIndexForBom),
    dependencies: dependencyGraphToCycloneDx(combinedDependencies),
    tools,
    timestamp,
    completeness: containerEvidence.skipped.length > 0 ? "incomplete" : "complete",
  });

  const sbomGo = path.join(outputDir, "sbom.go.cyclonedx.json");
  const sbomNode = path.join(outputDir, "sbom.node.cyclonedx.json");
  const sbomCombined = path.join(outputDir, "sbom.cyclonedx.json");
  writeJSON(sbomGo, goBom);
  writeJSON(sbomNode, nodeBom);
  writeJSON(sbomCombined, combinedBom);

  const licenseReportPath = path.join(outputDir, "license-report.json");
  const provenancePath = path.join(outputDir, "dependency-provenance.json");
  const commandsPath = path.join(outputDir, "commands-used.md");
  const commandsJsonPath = path.join(outputDir, "commands-used.json");
  const readinessPath = path.join(outputDir, "sbom-readiness-report.md");
  const licenseReportMarkdown = path.join(outputDir, "license-report.md");
  const toolingPlanPath = path.join(outputDir, "tooling-plan.md");
  const artifactDigestsPath = path.join(outputDir, "artifact-digests.json");

  licenseReport.sort((a, b) =>
    `${a.ecosystem}:${a.package}:${a.version}`.localeCompare(`${b.ecosystem}:${b.package}:${b.version}`),
  );
  writeJSON(licenseReportPath, {
    schema_id: "cartulary.license_report.v1",
    generated_at: timestamp,
    repository: repoRoot,
    commit,
    entries: licenseReport,
  });
  writeJSON(provenancePath, {
    schema_id: "cartulary.dependency_provenance.v1",
    generated_at: timestamp,
    repository: repoRoot,
    commit,
    provenance,
    import_scan: {
      node_used_external_packages: Object.fromEntries([...importScan.entries()].sort()),
      node_used_undeclared_packages: usedUndeclared.sort(),
      node_declared_not_observed_in_import_scan: declaredUnused.sort(),
      method: "regex import/export scan of apps/ and packages/ TypeScript/JavaScript sources; source imports corroborate use but do not establish transitive versions",
    },
    container_images: containerEvidence,
  });

  const artifactFiles = [
    sbomCombined,
    sbomGo,
    sbomNode,
    licenseReportPath,
    provenancePath,
    commandsPath,
    commandsJsonPath,
    readinessPath,
    licenseReportMarkdown,
    toolingPlanPath,
    ...containerEvidence.scanned.map((entry) => entry.output),
  ];

  const licenseSummary = new Map();
  for (const entry of licenseReport) {
    const key = entry.license_expression ?? "NOASSERTION";
    licenseSummary.set(key, (licenseSummary.get(key) ?? 0) + 1);
  }
  const textEvidenceCount = licenseReport.filter((entry) => entry.evidence_path).length;
  const referenceOnlyCount = licenseReport.filter((entry) => !entry.evidence_path && entry.raw_license_metadata).length;
  const missingEvidence = licenseReport.filter((entry) => entry.issue_flags.length > 0);

  writeMarkdownReport(licenseReportMarkdown, [
    "# Cartulary License Evidence Report",
    "",
    "This is an engineering evidence report, not a legal compliance certification.",
    "",
    "## Summary By License Expression",
    "",
    ...[...licenseSummary.entries()].sort().map(([license, count]) => `- ${license}: ${count}`),
    "",
    "## Evidence Coverage",
    "",
    `- Entries with collected package-distributed license text: ${textEvidenceCount}`,
    `- Entries with metadata/reference-only evidence: ${referenceOnlyCount}`,
    `- Entries with unresolved issue flags: ${missingEvidence.length}`,
    "",
    "## Unresolved License Issues",
    "",
    ...(missingEvidence.length > 0
      ? missingEvidence.map((entry) => `- ${entry.ecosystem}:${entry.package}@${entry.version}: ${entry.issue_flags.join(", ")}`)
      : ["- None observed by this generator."]),
  ]);

  writeMarkdownReport(toolingPlanPath, [
    "# SBOM Tooling Plan And Patch Summary",
    "",
    "- `make sbom` and `make license-report` generate `.cartulary/release-artifacts/sbom/<run-id>/` through `tools/release-evidence/generate-sbom-license-evidence.mjs`, then validate canonical artifacts.",
    "- Pinned Node tooling: `@cyclonedx/cdxgen@12.3.1` and `ajv@8.20.0` in the pnpm workspace.",
    "- Pinned Go tooling: `cyclonedx-gomod@v1.10.0` and `syft@v1.44.0` installed into `tmp/toolbin` by Make prerequisites.",
    "- Generated release artifacts remain ignored under `.cartulary/release-artifacts/`.",
    "- `release-check` continues to fail when the configured SBOM or license report is missing or empty.",
  ]);

  ensureDir(path.dirname(canonicalSbom));
  ensureDir(path.dirname(canonicalLicenseReport));
  copyFileSync(sbomCombined, canonicalSbom);
  copyFileSync(licenseReportPath, canonicalLicenseReport);

  const validator = path.join(repoRoot, "tools", "release-evidence", "validate-cyclonedx.mjs");
  runCommand(ctx, "validate combined sbom", ctx.config.node, [validator, sbomCombined]);
  runCommand(ctx, "validate go sbom", ctx.config.node, [validator, sbomGo]);
  runCommand(ctx, "validate node sbom", ctx.config.node, [validator, sbomNode]);
  for (const entry of containerEvidence.scanned) {
    runCommand(ctx, `validate container sbom ${entry.image}`, ctx.config.node, [validator, entry.output]);
  }
  runCommand(ctx, "validate canonical sbom", ctx.config.node, [validator, canonicalSbom]);

  writeJSON(commandsJsonPath, {
    schema_id: "cartulary.sbom_commands.v1",
    generated_at: timestamp,
    commands: ctx.commands,
  });
  writeMarkdownReport(commandsPath, commandsMarkdown(ctx, { timestamp, commit, status, nodeVersion, pnpmVersion }));

  const artifactDigestRows = artifactRows([
    ...artifactFiles.filter((file) => file !== readinessPath),
    canonicalSbom,
    canonicalLicenseReport,
  ]);
  const licenseEvidenceCount = existsSync(licensesDir)
    ? readdirSync(licensesDir).filter((name) => statSync(path.join(licensesDir, name)).isFile()).length
    : 0;
  const readinessLines = [
    "# Cartulary SBOM Readiness Report",
    "",
    "This package is an engineering evidence package for CRA-readiness work. It is not a legal compliance certification.",
    "",
    "## Scope",
    "",
    `- Repository path: ${repoRoot}`,
    `- Commit/hash: ${commit || "unavailable"}`,
    `- Generation time: ${timestamp}`,
    "- Ecosystems inspected: Go modules, pnpm/Node workspace, local container image references, generated-code surfaces, scripts/tools.",
    "- Artifact set covered: shipped Go server/migrate dependency graph, pnpm workspace dependency graph, local-service container references, and tool/generator dependency evidence.",
    "- Network access used: package-manager and Go commands used configured caches/registries; container images were not pulled implicitly.",
    "",
    "## Generated Artifacts",
    "",
    ...artifactDigestRows.map((artifact) => `- ${artifact.path} (${artifact.bytes} bytes, sha256=${artifact.sha256})`),
    `- ${rel(licensesDir)}/ (${licenseEvidenceCount} collected license evidence files)`,
    "",
    "## Coverage",
    "",
    `- Direct dependencies covered: ${licenseReport.filter((entry) => entry.direct).length}`,
    `- Transitive dependencies covered: ${licenseReport.filter((entry) => entry.transitive).length}`,
    "- Runtime/build/test/dev/generated-code/container classifications are recorded when manifest, package-manager, or import evidence supports them.",
    "- Explicit exclusions: container OS package details are incomplete unless the referenced image already exists locally and Syft successfully scans it.",
    "",
    "## License Evidence",
    "",
    ...[...licenseSummary.entries()].sort().map(([license, count]) => `- ${license}: ${count}`),
    `- Dependencies with collected full license text: ${textEvidenceCount}`,
    `- Dependencies with only license references/metadata: ${referenceOnlyCount}`,
    `- Dependencies with missing, ambiguous, custom, or conflicting evidence flags: ${missingEvidence.length}`,
    "",
    "## Dependency Provenance",
    "",
    "- Go provenance comes from `go.mod`, `go.sum`, `go list -m -json all`, `go mod graph`, and `go list -deps` scans.",
    "- Node provenance comes from workspace `package.json` files, `pnpm-lock.yaml`, `pnpm list`, `pnpm licenses`, and import scans.",
    "- Workspace packages are first-party components and use the repository root `LICENSE` as first-party license evidence.",
    "- No `vendor/` directory or checked-in third-party license directory was observed by this generator.",
    "",
    "## Toolchain Integration",
    "",
    "- Existing release commands found: `make release-check`, `make license-report`, and `make sbom`.",
    "- Added/proposed commands: existing `make sbom` and `make license-report` now generate evidence artifacts through the repo-local generator before checking non-empty outputs.",
    "- Regenerate with: `make sbom` or `make license-report`.",
    "- Release verification checks canonical non-empty SBOM/license artifacts through `tools/release-evidence/check-release-artifact.sh`.",
    "",
    "## Unresolved Issues",
    "",
    ...(unresolvedIssues.length > 0 ? unresolvedIssues.slice(0, 200).map((issue) => `- ${issue}`) : ["- None observed."]),
    unresolvedIssues.length > 200 ? `- Additional unresolved entries omitted from Markdown summary: ${unresolvedIssues.length - 200}` : "",
    usedUndeclared.length > 0 ? `- Node imports observed without direct manifest declarations: ${usedUndeclared.join(", ")}` : "- Node import scan found no strong used-but-undeclared direct dependency evidence.",
    declaredUnused.length > 0
      ? `- Declared Node dependencies not observed by import scan: ${declaredUnused.join(", ")}. This is informational because config files, type-only use, transitive plugins, and test/runtime entrypoints can hide usage from a simple import scan.`
      : "- Node import scan found no declared dependency gaps.",
    containerEvidence.skipped.length > 0
      ? `- Container/image scan gaps: ${containerEvidence.skipped.map((entry) => `${entry.image} (${entry.reason})`).join("; ")}`
      : "- Container/image scan gaps: none observed.",
    "",
    "## Assumptions And Limitations",
    "",
    "- License expressions are not inferred from package names, popularity, or license-file conventions.",
    "- Go module license expressions are reported as unresolved unless package metadata or tool output provides support.",
    "- pnpm license metadata is registry/package metadata and may require legal review when full license text is missing.",
    "- Generated code is attributed to repo-local generators and generator/tool pins; generated outputs are not treated as third-party source by themselves.",
    "- Container image package/license completeness depends on local image availability and Syft scan success.",
    "",
    "## Follow-Up Work Before CRA-Ready Treatment",
    "",
    "- Review every `license_text_missing`, `license_metadata_missing`, and `license_expression_missing` entry in `license-report.json`.",
    "- Decide whether release artifacts should remain generated-only or be retained per release.",
    "- Run container image scans from a controlled environment with pinned image digests and retain the resulting container SBOMs.",
    "- Add legal-owner review for entries with copyleft, commercial, custom, missing, or ambiguous license flags.",
    "- Re-run `make sbom` from a fresh clone using the same manifests and lockfiles and compare artifact shape plus unresolved issue counts.",
  ].filter(Boolean);
  writeMarkdownReport(readinessPath, readinessLines);

  writeJSON(artifactDigestsPath, {
    schema_id: "cartulary.sbom_artifact_digests.v1",
    generated_at: timestamp,
    artifacts: artifactRows([...artifactFiles, canonicalSbom, canonicalLicenseReport]),
  });

  console.log(`SBOM/license evidence generated: ${rel(outputDir)}`);
  console.log(`Canonical SBOM: ${rel(canonicalSbom)}`);
  console.log(`Canonical license report: ${rel(canonicalLicenseReport)}`);
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    main();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(`SBOM/license evidence generation failed: ${message}`);
    process.exit(1);
  }
}

export {
  artifactRows,
  findLicenseFiles,
  licenseReviewFlags,
  makeCycloneDxBom,
  normalizePackageName,
  parseGoModuleGraph,
  parseJSONStream,
};
