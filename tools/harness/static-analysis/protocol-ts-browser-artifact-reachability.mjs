#!/usr/bin/env node

import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../..",
);

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

function artifactRecords(source, sourcePath) {
  const records = [];
  const recordPattern =
    /path:\s*("(?:\\.|[^"\\])*")\s*,\s*json:\s*("(?:\\.|[^"\\])*")\s*,\s*sha256:\s*("[a-f0-9]{64}")/gu;
  for (const match of source.matchAll(recordPattern)) {
    const artifactPath = JSON.parse(match[1]);
    const json = JSON.parse(match[2]);
    const sha256 = JSON.parse(match[3]);
    records.push({ artifactPath, json, sha256, sourcePath });
  }
  if (records.length === 0) {
    throw new Error(`no artifact records found in ${sourcePath}`);
  }
  return records;
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
  const modules = [
    "packages/protocol-ts/src/generated/audit-artifacts.ts",
    "packages/protocol-ts/src/generated/revisions-artifacts.ts",
  ];
  const records = [];
  for (const modulePath of modules) {
    const absolutePath = path.join(repositoryRoot, modulePath);
    records.push(
      ...artifactRecords(await readFile(absolutePath, "utf8"), modulePath),
    );
  }

  const signals = [];
  for (const record of records) {
    signals.push(
      { kind: "artifact_path", value: record.artifactPath },
      { kind: "artifact_digest", value: record.sha256 },
    );
    for (const schemaIdentifier of schemaIdentifiers(JSON.parse(record.json))) {
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
  return { modules, signals };
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

  const { modules, signals } = await forbiddenEvidence();
  const findings = [];
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
      for (const modulePath of modules) {
        if (normalized.endsWith(modulePath)) {
          findings.push({
            emittedPath: relativePath,
            kind: "source_map_module",
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
