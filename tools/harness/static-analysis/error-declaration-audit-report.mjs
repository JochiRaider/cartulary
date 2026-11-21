import { readJsonObject, assertObjectKeys, assertRequiredKeys, requireSchemaID, requireArray, requireInteger, requireString, requireStringArray, requireSorted } from "../contract/json-shape.mjs";

export const errorDeclarationAuditSchemaID = "cartulary.harness_error_declaration_audit.v1";

const reportKeys = ["schema_id", "operation_count", "rooted_operation_count", "unrooted_operations", "finding_count", "findings"];
const findingKeys = ["operation_id", "code", "functions", "functions_omitted"];

// validateErrorDeclarationAudit checks one advisory audit report's closed shape
// (TH-HARNESS-REQ-867). It validates structure only; findings are never errors.
export function validateErrorDeclarationAudit(report, label = "error declaration audit") {
  assertObjectKeys(report, new Set(reportKeys), label);
  assertRequiredKeys(report, reportKeys, label);
  requireSchemaID(report, errorDeclarationAuditSchemaID, label);
  const operations = requireInteger(report.operation_count, `${label}.operation_count`, { min: 1 });
  const rooted = requireInteger(report.rooted_operation_count, `${label}.rooted_operation_count`, { min: 0 });
  const unrooted = requireStringArray(report.unrooted_operations, `${label}.unrooted_operations`);
  requireSorted(unrooted, `${label}.unrooted_operations`, (operation) => operation, "operation ID");
  if (rooted + unrooted.length !== operations) throw new Error(`${label}: rooted and unrooted operations must total operation_count`);
  const findings = requireArray(report.findings, `${label}.findings`);
  if (requireInteger(report.finding_count, `${label}.finding_count`, { min: 0 }) !== findings.length) throw new Error(`${label}.finding_count must equal the findings`);
  const keys = [];
  for (const [index, finding] of findings.entries()) {
    const item = `${label}.findings[${index}]`;
    assertObjectKeys(finding, new Set(findingKeys), item);
    assertRequiredKeys(finding, findingKeys, item);
    requireString(finding.operation_id, `${item}.operation_id`);
    requireString(finding.code, `${item}.code`, { pattern: /^[a-z][a-z0-9_]*$/u });
    const functions = requireStringArray(finding.functions, `${item}.functions`, { nonEmpty: true });
    if (functions.length > 5) throw new Error(`${item}.functions is bounded at 5`);
    requireInteger(finding.functions_omitted, `${item}.functions_omitted`, { min: 0 });
    keys.push(`${finding.operation_id}\u0000${finding.code}`);
  }
  requireSorted(keys, `${label}.findings`, (key) => key, "operation ID and code");
  return report;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const path = process.argv[2];
  if (!path) {
    process.stderr.write("usage: error-declaration-audit-report.mjs <report.json>\n");
    process.exit(2);
  }
  const report = validateErrorDeclarationAudit(readJsonObject(path));
  process.stdout.write(`error-declaration-audit report valid operations=${report.operation_count} findings=${report.finding_count} (advisory)\n`);
}
