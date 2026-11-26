// suiteAttachmentEnvironment is the only Node writer of the suite services
// attachment names (TH-HARNESS Section 5.5, single writer): which test-services
// suite a process belongs to and how it relates to it. The Go half of the
// services component (internal/testutil/suiteservices, tools/testservices)
// reads and writes the same names.
export function suiteAttachmentEnvironment({ suiteID, callMode, lifecycleMode, persistentBorrower = false }) {
  return {
    CARTULARY_TEST_SUITE_ID: suiteID,
    ...(callMode ? { CARTULARY_TEST_SERVICES_CALL_MODE: callMode } : {}),
    ...(lifecycleMode ? { CARTULARY_TEST_SERVICES_LIFECYCLE_MODE: lifecycleMode } : {}),
    ...(persistentBorrower ? { CARTULARY_TEST_SERVICES_PERSISTENT_BORROWER: "1" } : {}),
  };
}

// postgresAllocationEnvironment is the only Node writer of a suite process's
// PostgreSQL allocation: its isolation and allocation reference. It always
// writes both; suiteservices.PostgresAllocation is its Go form.
export function postgresAllocationEnvironment({ isolation, ref }) {
  return { CARTULARY_PGTEST_ISOLATION: isolation, CARTULARY_PGTEST_ALLOCATION_REF: ref };
}
