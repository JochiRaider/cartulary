#!/usr/bin/env node

import { createHash } from "node:crypto";
import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../..",
);
const removedGeneratedModules = [
  "packages/protocol-ts/src/generated/audit-artifacts.ts",
  "packages/protocol-ts/src/generated/errors-artifacts.ts",
  "packages/protocol-ts/src/generated/extensions-artifacts.ts",
  "packages/protocol-ts/src/generated/index.ts",
  "packages/protocol-ts/src/generated/network-flow-artifacts.ts",
  "packages/protocol-ts/src/generated/protocol-validators.ts",
  "packages/protocol-ts/src/generated/revisions-artifacts.ts",
  "packages/protocol-ts/src/generated/ws-artifacts.ts",
];
const networkFlowRuntimeModules = [
  "packages/protocol-ts/src/entrypoints/network-flow.ts",
  "packages/protocol-ts/src/generated/network-flow-descriptor.ts",
  "packages/protocol-ts/src/generated/network-flow-error-registry.ts",
  "packages/protocol-ts/src/generated/network-flow-mapping-registry.ts",
  "packages/protocol-ts/src/generated/network-flow-presentation.ts",
  "packages/protocol-ts/src/generated/network-flow-validators.ts",
];
const requiredNetworkFlowRuntimeModules = [
  "packages/protocol-ts/src/entrypoints/network-flow.ts",
  "packages/protocol-ts/src/generated/network-flow-validators.ts",
];

function parseArguments(argv) {
  const result = {
    dist: path.join(repositoryRoot, "apps/web/dist"),
  };
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === "--dist") {
      const value = argv[index + 1];
      if (!value) {
        throw new Error("--dist requires a path");
      }
      result.dist = path.resolve(repositoryRoot, value);
      index += 1;
      continue;
    }
    throw new Error(`unexpected argument ${argument}`);
  }
  return result;
}

async function emittedFiles(root) {
  const result = [];
  async function visit(current) {
    for (const entry of await readdir(current, { withFileTypes: true })) {
      const entryPath = path.join(current, entry.name);
      if (entry.isDirectory()) {
        await visit(entryPath);
      } else if (entry.isFile() && (entry.name.endsWith(".js") || entry.name.endsWith(".map"))) {
        result.push(entryPath);
      }
    }
  }
  await visit(root);
  return result.sort((left, right) => left.localeCompare(right));
}

function canonicalJSON(value) {
  if (Array.isArray(value)) {
    return value.map(canonicalJSON);
  }
  if (!value || typeof value !== "object") {
    return value;
  }
  return Object.fromEntries(
    Object.keys(value)
      .sort((left, right) => left.localeCompare(right))
      .map((key) => [key, canonicalJSON(value[key])]),
  );
}

async function contractFiles(contractRoot) {
  const absoluteRoot = path.join(repositoryRoot, contractRoot);
  const result = [];
  async function visit(current) {
    for (const entry of await readdir(current, { withFileTypes: true })) {
      const entryPath = path.join(current, entry.name);
      if (entry.isDirectory()) {
        await visit(entryPath);
        continue;
      }
      if (
        entry.isFile() &&
        [".json", ".yaml", ".yml"].includes(path.extname(entry.name).toLowerCase())
      ) {
        result.push(entryPath);
      }
    }
  }
  await visit(absoluteRoot);
  return result.sort((left, right) => left.localeCompare(right));
}

function schemaIdentifiers(value, result = new Set()) {
  if (Array.isArray(value)) {
    for (const item of value) {
      schemaIdentifiers(item, result);
    }
    return result;
  }
  if (!value || typeof value !== "object") {
    return result;
  }
  for (const [key, member] of Object.entries(value)) {
    if ((key === "$id" || key === "schema_id") && typeof member === "string") {
      result.add(member);
    }
    schemaIdentifiers(member, result);
  }
  return result;
}

async function forbiddenEvidence() {
  const registry = JSON.parse(
    await readFile(path.join(repositoryRoot, "contracts/index.json"), "utf8"),
  );
  if (registry.schema_id !== "cartulary.contract_family_registry.v4") {
    throw new Error("contracts/index.json has an unexpected schema_id");
  }
  const protectedFamilyIDs = ["audit", "revisions"];
  const families = protectedFamilyIDs.map((familyID) => {
    const family = registry.families?.find(
      (candidate) => candidate?.family_id === familyID,
    );
    if (!family) {
      throw new Error(`contracts/index.json is missing protected family ${familyID}`);
    }
    if (
      family.generation_status !== "active" ||
      typeof family.contract_root !== "string"
    ) {
      throw new Error(`protected family ${familyID} must remain active with a contract root`);
    }
    if (
      family.typescript_projections?.length !== 0 ||
      family.generated_outputs?.some((output) =>
        output.startsWith("packages/protocol-ts/src/generated/"),
      )
    ) {
      throw new Error(`protected family ${familyID} must not declare browser projections`);
    }
    return family;
  });
  const protectedModules = protectedFamilyIDs.map(
    (familyID) =>
      `packages/protocol-ts/src/generated/${familyID}-artifacts.ts`,
  );
  const records = [];
  for (const family of families) {
    for (const absolutePath of await contractFiles(family.contract_root)) {
      const artifactPath = path
        .relative(repositoryRoot, absolutePath)
        .split(path.sep)
        .join("/");
      const decoded = JSON.parse(await readFile(absolutePath, "utf8"));
      const json = JSON.stringify(canonicalJSON(decoded));
      records.push({
        artifactPath,
        decoded,
        json,
        sha256: createHash("sha256").update(json).digest("hex"),
      });
    }
  }

  const signals = [];
  for (const record of records) {
    signals.push(
      { kind: "artifact_path", value: record.artifactPath },
      { kind: "artifact_digest", value: record.sha256 },
    );
    for (const schemaIdentifier of schemaIdentifiers(record.decoded)) {
      signals.push({ kind: "schema_identifier", value: schemaIdentifier });
    }
    const signature = record.json.slice(0, 96);
    signals.push(
      { kind: "embedded_json_signature", value: signature },
      {
        kind: "embedded_json_signature",
        value: JSON.stringify(signature).slice(1, -1),
      },
    );
  }
  return { protectedModules, signals };
}

function scanSignals(text, emittedPath, signals, findings) {
  for (const signal of signals) {
    if (signal.value.length > 0 && text.includes(signal.value)) {
      findings.push({
        emittedPath,
        kind: signal.kind,
        value: signal.value,
      });
    }
  }
}

async function main() {
  const options = parseArguments(process.argv.slice(2));
  const files = await emittedFiles(options.dist);
  if (files.length === 0) {
    throw new Error(`no emitted JavaScript or source maps found under ${options.dist}`);
  }

  const { protectedModules, signals } = await forbiddenEvidence();
  const forbiddenModules = [
    ...new Set([...protectedModules, ...removedGeneratedModules]),
  ].sort((left, right) => left.localeCompare(right));
  const findings = [];
  const networkFlowRuntimeModulesSeen = new Set();
  let javascriptCount = 0;
  let sourceMapCount = 0;
  for (const emittedPath of files) {
    const relativePath = path.relative(repositoryRoot, emittedPath);
    const text = await readFile(emittedPath, "utf8");
    scanSignals(text, relativePath, signals, findings);
    if (emittedPath.endsWith(".js")) {
      javascriptCount += 1;
      continue;
    }

    sourceMapCount += 1;
    let sourceMap;
    try {
      sourceMap = JSON.parse(text);
    } catch (error) {
      throw new Error(`invalid source map ${relativePath}: ${error.message}`);
    }
    const sources = Array.isArray(sourceMap.sources) ? sourceMap.sources : [];
    for (const source of sources) {
      if (typeof source !== "string") {
        continue;
      }
      const normalized = source.replaceAll("\\", "/");
      for (const modulePath of forbiddenModules) {
        if (normalized.endsWith(modulePath)) {
          findings.push({
            emittedPath: relativePath,
            kind: "source_map_module",
            value: modulePath,
          });
        }
      }
      for (const modulePath of networkFlowRuntimeModules) {
        if (!normalized.endsWith(modulePath)) {
          continue;
        }
        networkFlowRuntimeModulesSeen.add(modulePath);
        if (!path.basename(emittedPath).startsWith("NetworkFlowFeature-")) {
          findings.push({
            emittedPath: relativePath,
            kind: "network_flow_graph_escape",
            value: modulePath,
          });
        }
      }
    }
    const sourcesContent = Array.isArray(sourceMap.sourcesContent)
      ? sourceMap.sourcesContent
      : [];
    for (const sourceContent of sourcesContent) {
      if (typeof sourceContent === "string") {
        scanSignals(sourceContent, relativePath, signals, findings);
      }
    }
  }

  for (const modulePath of requiredNetworkFlowRuntimeModules) {
    if (!networkFlowRuntimeModulesSeen.has(modulePath)) {
      findings.push({
        emittedPath: path.relative(repositoryRoot, options.dist),
        kind: "network_flow_graph_missing",
        value: modulePath,
      });
    }
  }

  const uniqueFindings = [...new Map(
    findings.map((finding) => [
      `${finding.emittedPath}\u0000${finding.kind}\u0000${finding.value}`,
      finding,
    ]),
  ).values()].sort((left, right) =>
    `${left.emittedPath}:${left.kind}:${left.value}`.localeCompare(
      `${right.emittedPath}:${right.kind}:${right.value}`,
    ),
  );

  if (uniqueFindings.length > 0) {
    console.error(
      `protocol-ts browser artifact reachability failed: ${uniqueFindings.length} forbidden matches`,
    );
    for (const finding of uniqueFindings.slice(0, 30)) {
      console.error(
        `- ${finding.emittedPath}: ${finding.kind} ${JSON.stringify(finding.value)}`,
      );
    }
    if (uniqueFindings.length > 30) {
      console.error(`- ${uniqueFindings.length - 30} additional matches omitted`);
    }
    process.exitCode = 1;
    return;
  }

  console.log(
    `protocol-ts browser artifacts verified: js=${javascriptCount} maps=${sourceMapCount}`,
  );
}

await main();
