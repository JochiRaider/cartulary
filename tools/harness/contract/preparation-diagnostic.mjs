import { validateSchemaSync } from "./harness-contract.mjs";

// The closed record a rejected graph preparation publishes instead of a plan
// (TH-HARNESS-REQ-854). It carries typed facts only, never exception text.
export const preparationDiagnosticSchemaID = "cartulary.harness_preparation_diagnostic.v1";
export const maximumPreparationDiagnosticBytes = 4096;

const reasons = new Set(["configuration_error", "usage_error"]);
const causes = new Set(["catalog_shape", "selector_resolution", "catalog_coverage", "command_selection", "unclassified"]);
const maximumSubjects = 20;
const maximumSubjectCharacters = 256;
const controlCharacter = /[\x00-\x1f\x7f]/u;

function admissibleSubject(subject) {
  return typeof subject === "string" && subject.length > 0 && [...subject].length <= maximumSubjectCharacters && !controlCharacter.test(subject);
}

export function preparationDiagnostic(error) {
  const reason = reasons.has(error?.failure_reason) ? error.failure_reason : "configuration_error";
  const cause = causes.has(error?.diagnostic_cause) ? error.diagnostic_cause : "unclassified";
  const offered = cause === "unclassified" || !Array.isArray(error?.diagnostic_subjects) ? [] : error.diagnostic_subjects;
  const admissible = offered.filter(admissibleSubject);
  const subjects = [...new Set(admissible)];
  const unrepresentable = offered.length - admissible.length;
  const record = (kept) => ({
    schema_id: preparationDiagnosticSchemaID,
    failure_reason: reason,
    cause,
    subjects: kept,
    omitted_subject_count: unrepresentable + subjects.length - kept.length,
  });
  let kept = subjects.slice(0, maximumSubjects);
  while (kept.length > 0 && Buffer.byteLength(JSON.stringify(record(kept))) > maximumPreparationDiagnosticBytes) kept = kept.slice(0, -1);
  const diagnostic = record(kept);
  validateSchemaSync(preparationDiagnosticSchemaID, diagnostic);
  return diagnostic;
}
