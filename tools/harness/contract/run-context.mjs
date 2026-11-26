// The single Node writer of a harness command's invocation context (TH-HARNESS
// Section 5.5, single writer): where the run's results go, which run, target
// and work unit the command reports for, the accounted invocation and host
// admission it runs under, the operation it executes and where its public
// inputs came from. Its Go form is internal/testutil/runcontext. Readers keep
// reading the names; every writer outside the Make entry point calls this.
const contextNames = Object.freeze({
  resultsDir: "CARTULARY_TEST_RESULTS_DIR",
  runID: "CARTULARY_TEST_RUN_ID",
  target: "CARTULARY_TEST_TARGET",
  workUnit: "CARTULARY_WORK_UNIT_ID",
  physicalOutputRoot: "CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT",
  invocationRoot: "CARTULARY_HARNESS_INVOCATION_ROOT",
  admissionLease: "CARTULARY_HOST_ADMISSION_LEASE",
  operationDigest: "CARTULARY_OPERATION_DIGEST",
});

// runContextEnvironment writes only the fields that are set: an empty value is
// never a meaningful context, so a writer that means "none" removes the name.
export function runContextEnvironment(context) {
  const unknown = Object.keys(context).filter((field) => !Object.hasOwn(contextNames, field));
  if (unknown.length > 0) throw new Error(`unknown run context fields: ${unknown.join(", ")}`);
  return Object.fromEntries(Object.entries(contextNames)
    .filter(([field]) => context[field] !== undefined && context[field] !== null && context[field] !== "")
    .map(([field, name]) => [name, String(context[field])]));
}

// inputSourcesEnvironment records where each public input came from as
// NAME=source entries, in the given order. It always writes, so an empty list
// clears an inherited record.
export function inputSourcesEnvironment(sources) {
  return { CARTULARY_MAKE_INPUT_SOURCES: sources.join(" ") };
}
