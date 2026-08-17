#!/usr/bin/env node
import { readFileSync } from "node:fs";

import { validateSchemaSync } from "../harness/contract/index.mjs";

function main() {
  const [file] = process.argv.slice(2);
  if (!file) {
    throw new Error("usage: validate-license-report.mjs <license-report-json>");
  }
  const report = JSON.parse(readFileSync(file, "utf8"));
  if (report.schema_id !== "cartulary.license_report.v2") {
    throw new Error(`${file}: unsupported license report schema ${report.schema_id ?? "missing"}`);
  }
  validateSchemaSync(report.schema_id, report);
  const identities = report.entries.map(
    (entry) => `${entry.ecosystem}:${entry.package}:${entry.version}`,
  );
  const sortedIdentities = [...identities].sort();
  if (JSON.stringify(identities) !== JSON.stringify(sortedIdentities)) {
    throw new Error(`${file}: license report entries are not in canonical order`);
  }
  if (new Set(identities).size !== identities.length) {
    throw new Error(`${file}: license report contains duplicate package identities`);
  }
  for (const entry of report.entries) {
    if (entry.direct === entry.transitive) {
      throw new Error(`${file}: ${entry.package} must be exactly one of direct or transitive`);
    }
    for (const field of ["issue_flags", "review_flags"]) {
      if (JSON.stringify(entry[field]) !== JSON.stringify([...entry[field]].sort())) {
        throw new Error(`${file}: ${entry.package}.${field} is not in canonical order`);
      }
    }
  }
  console.log(`${file}: valid ${report.schema_id}`);
}

try {
  main();
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
}
