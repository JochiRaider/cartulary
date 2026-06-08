#!/usr/bin/env node

import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

import { redactValue } from "./harness-contract.mjs";

const schemaID = "cartulary.govulncheck_findings.v1";

function usage() {
  process.stderr.write(
    "usage: govulncheck-findings.mjs --input <json-stream> --output <findings-json>\n",
  );
}

function parseArgs(argv) {
  const options = {};
  for (let index = 0; index < argv.length; index += 1) {
    const token = argv[index];
    if (token === "--input" || token === "--output") {
      const value = argv[index + 1];
      if (!value) {
        throw new Error(`${token} requires a value`);
      }
      options[token.slice(2)] = value;
      index += 1;
      continue;
    }
    throw new Error(`unknown argument ${token}`);
  }
  if (!options.input || !options.output) {
    throw new Error("missing required input or output");
  }
  return options;
}

function parseJSONStream(source) {
  const values = [];
  let index = 0;
  while (index < source.length) {
    while (index < source.length && /\s/u.test(source[index])) {
      index += 1;
    }
    if (index >= source.length) {
      break;
    }

    const start = index;
    let depth = 0;
    let inString = false;
    let escaped = false;
    for (; index < source.length; index += 1) {
      const char = source[index];
      if (inString) {
        if (escaped) {
          escaped = false;
        } else if (char === "\\") {
          escaped = true;
        } else if (char === "\"") {
          inString = false;
        }
        continue;
      }
      if (char === "\"") {
        inString = true;
      } else if (char === "{") {
        depth += 1;
      } else if (char === "}") {
        depth -= 1;
        if (depth === 0) {
          index += 1;
          break;
        }
      }
    }
    if (depth !== 0 || inString) {
      throw new Error("unterminated JSON object in Govulncheck output");
    }
    values.push(JSON.parse(source.slice(start, index)));
  }
  return values;
}

function uniqueStrings(values) {
  return [...new Set(values.filter((value) => typeof value === "string" && value !== ""))].sort();
}

function fixedVersions(osv = {}) {
  const versions = [];
  for (const affected of osv.affected ?? []) {
    for (const range of affected.ranges ?? []) {
      for (const event of range.events ?? []) {
        if (typeof event.fixed === "string" && event.fixed !== "") {
          versions.push(event.fixed.startsWith("v") ? event.fixed : `v${event.fixed}`);
        }
      }
    }
  }
  return uniqueStrings(versions);
}

function affectedPackages(osv = {}) {
  return uniqueStrings(
    (osv.affected ?? []).map((affected) => affected.package?.name ?? ""),
  );
}

function traceReachability(trace = []) {
  if (trace.some((entry) => typeof entry.function === "string" && entry.function !== "")) {
    return "symbol";
  }
  if (trace.some((entry) => typeof entry.package === "string" && entry.package !== "")) {
    return "package";
  }
  return "module";
}

function normalizeTrace(trace = []) {
  return trace.map((entry) => ({
    module: entry.module ?? "",
    version: entry.version ?? "",
    package: entry.package ?? "",
    receiver: entry.receiver ?? "",
    function: entry.function ?? "",
    position: entry.position
      ? {
          filename: entry.position.filename ?? "",
          line: entry.position.line ?? null,
          column: entry.position.column ?? null,
        }
      : null,
  }));
}

function normalizeFinding(finding, osvByID) {
  const id = finding.osv ?? "";
  const osv = osvByID.get(id) ?? {};
  const trace = normalizeTrace(finding.trace ?? []);
  const reachability = traceReachability(trace);
  return {
    id,
    aliases: uniqueStrings(osv.aliases ?? []),
    summary: osv.summary ?? "",
    fixed_version: finding.fixed_version ?? "",
    fixed_versions: fixedVersions(osv),
    affected_packages: affectedPackages(osv),
    reachability,
    blocking: reachability === "symbol",
    modules: uniqueStrings(trace.map((entry) => entry.module)),
    packages: uniqueStrings(trace.map((entry) => entry.package)),
    symbols: trace
      .filter((entry) => entry.function)
      .map((entry) => ({
        package: entry.package,
        receiver: entry.receiver,
        function: entry.function,
        position: entry.position,
      })),
    trace,
  };
}

function summarize(values) {
  const config = values.find((value) => value.config)?.config ?? null;
  const osvByID = new Map();
  for (const value of values) {
    if (value.osv?.id && !osvByID.has(value.osv.id)) {
      osvByID.set(value.osv.id, value.osv);
    }
  }
  const findings = values
    .filter((value) => value.finding)
    .map((value) => normalizeFinding(value.finding, osvByID))
    .sort((left, right) =>
      [
        left.blocking ? "0" : "1",
        left.id,
        left.reachability,
        left.modules.join(","),
        left.packages.join(","),
      ]
        .join("\0")
        .localeCompare(
          [
            right.blocking ? "0" : "1",
            right.id,
            right.reachability,
            right.modules.join(","),
            right.packages.join(","),
          ].join("\0"),
        ),
    );
  const blockingFindings = findings.filter((finding) => finding.blocking);
  const reachabilityCounts = {
    module: findings.filter((finding) => finding.reachability === "module").length,
    package: findings.filter((finding) => finding.reachability === "package").length,
    symbol: findings.filter((finding) => finding.reachability === "symbol").length,
  };
  return redactValue({
    schema_id: schemaID,
    tool: "govulncheck",
    status: blockingFindings.length > 0 ? "fail" : "pass",
    config,
    counts: {
      raw_event_count: values.length,
      osv_count: osvByID.size,
      finding_count: findings.length,
      blocking_count: blockingFindings.length,
      reachability: reachabilityCounts,
    },
    vulnerability_ids: uniqueStrings(findings.map((finding) => finding.id)),
    blocking_vulnerability_ids: uniqueStrings(
      blockingFindings.map((finding) => finding.id),
    ),
    findings,
  });
}

function main(argv) {
  const options = parseArgs(argv);
  const source = readFileSync(options.input, "utf8");
  const values = parseJSONStream(source);
  if (values.length === 0) {
    throw new Error("Govulncheck JSON output was empty");
  }
  const summary = summarize(values);
  mkdirSync(path.dirname(options.output), { recursive: true, mode: 0o700 });
  writeFileSync(options.output, `${JSON.stringify(summary, null, 2)}\n`, {
    mode: 0o600,
  });
  return summary.counts.blocking_count > 0 ? 1 : 0;
}

try {
  process.exitCode = main(process.argv.slice(2));
} catch (error) {
  usage();
  process.stderr.write(
    `govulncheck JSON parse failed: ${error instanceof Error ? error.message : String(error)}\n`,
  );
  process.exitCode = 2;
}
