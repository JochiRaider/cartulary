// Implementation mirror of Testing Harness NLSpec Section 4.1A helper ownership.
// Public harness compatibility remains defined by Make targets, command IDs,
// schemas, output/failure contracts, retained paths, cleanup, and public inputs.
export const ownerFacadePathLists = Object.freeze({
  backend: Object.freeze([
    "tools/harness/backend/backend-duration-accounting.mjs",
    "tools/harness/backend/backend-shard-plan.mjs",
    "tools/harness/backend/backend-target-execution.mjs",
    "tools/harness/backend/backend-target-plan.mjs",
  ]),
  contract: Object.freeze([
    "tools/harness/contract/index.mjs",
  ]),
  browser: Object.freeze([
    "tools/harness/browser/accessibility-summary-cli.mjs",
    "tools/harness/browser/browser-batch-manifest.mjs",
    "tools/harness/browser/browser-duration-accounting.mjs",
    "tools/harness/browser/browser-shard-plan.mjs",
    "tools/harness/output/test-output/playwright-artifacts.mjs",
    "tools/harness/scheduler/adapters/browser.mjs",
  ]),
  command_surface: Object.freeze([
    "tools/harness/command-surface/make-node-tools.mjs",
  ]),
  duration_accounting: Object.freeze([
    "tools/harness/duration-accounting/index.mjs",
    "tools/harness/duration-accounting/duration-drift.mjs",
    "tools/harness/duration-accounting/duration-baseline-cli.mjs",
    "tools/harness/duration-accounting/target-duration-baselines.mjs",
  ]),
  execution_runtime: Object.freeze([
    "tools/harness/execution/phase-runtime.sh",
  ]),
  output: Object.freeze([
    "tools/harness/output/index.mjs",
    "tools/harness/output/test-output.mjs",
  ]),
  frontend: Object.freeze([
    "tools/harness/browser/accessibility-summary-cli.mjs",
    "tools/harness/execution/run-frontend-unit.sh",
    "tools/harness/execution/run-vitest-manifest-phase.sh",
    "tools/harness/execution/run-vitest-phase.sh",
    "tools/harness/generated-artifacts/design-tokens/index.mjs",
    "tools/harness/phase-accounting/frontend/index.mjs",
    "tools/harness/phase-accounting/frontend-readiness.mjs",
    "tools/harness/readiness/build-web-artifact.sh",
    "tools/harness/readiness/embed-web-assets.sh",
    "tools/harness/readiness/frontend-install.sh",
    "tools/harness/readiness/frontend-toolchain.sh",
    "tools/harness/static-analysis/font-bundle-check-cli.mjs",
  ]),
  phase_accounting: Object.freeze([
    "tools/harness/phase-accounting/frontend/index.mjs",
    "tools/harness/phase-accounting/frontend-phase-manifest.mjs",
    "tools/harness/phase-accounting/frontend-readiness.mjs",
    "tools/harness/phase-accounting/frontend-row-accounting.mjs",
    "tools/harness/phase-accounting/phase-manifest.mjs",
    "tools/harness/phase-accounting/phase-registry.mjs",
    "tools/harness/phase-accounting/phase-slice-plan.mjs",
  ]),
  scheduler: Object.freeze([
    "tools/harness/scheduler/phase-slice-execution.mjs",
    "tools/harness/scheduler/scheduler-runner.mjs",
    "tools/harness/scheduler/scheduler-family-contract.mjs",
    "tools/harness/scheduler/scheduler-manifest.mjs",
    "tools/harness/scheduler/scheduler-resource-policy.mjs",
    "tools/harness/scheduler/scheduler-reporting.mjs",
    "tools/harness/scheduler/scheduler-resources.mjs",
    "tools/harness/scheduler/process-executor.mjs",
  ]),
  scheduler_diagnostics: Object.freeze([
    "tools/harness/scheduler/scheduler/event-order.mjs",
    "tools/harness/scheduler/scheduler/summary-timing-drift.mjs",
  ]),
  service_backed_execution: Object.freeze([
    "tools/harness/execution/service-backed/schedule-planning.mjs",
  ]),
  test_output: Object.freeze([
    "tools/harness/output/test-output/frontend-indexes.mjs",
    "tools/harness/output/test-output/frontend-row-evidence.mjs",
    "tools/harness/output/test-output/playwright-artifacts.mjs",
  ]),
});

export const browserPrivateImportAllowedSourcePaths = Object.freeze([
  "tools/harness/output/test-output/playwright-artifacts.mjs",
  "tools/harness/scheduler/adapters/browser.mjs",
]);

export const unsupportedPrivateHelperRules = Object.freeze([
  Object.freeze({
    id: "legacy_backend_database_contract_drift",
    prefixes: Object.freeze(["tools/harness/backend/drift/"]),
    exact: Object.freeze([
      "tools/harness/backend/migration-history-cli.mjs",
      "tools/harness/backend/migration-history.mjs",
      "tools/harness/backend/schema-object-ownership-cli.mjs",
      "tools/harness/backend/schema-object-ownership.mjs",
    ]),
  }),
  Object.freeze({
    id: "legacy_backend_duration_and_shard_helpers",
    prefixes: Object.freeze([
      "tools/harness/backend/duration/",
      "tools/harness/backend/runner/",
    ]),
    exact: Object.freeze([]),
  }),
  Object.freeze({
    id: "legacy_backend_security_findings_helper",
    prefixes: Object.freeze([]),
    exact: Object.freeze(["tools/harness/backend/govulncheck-findings.mjs"]),
  }),
  Object.freeze({
    id: "legacy_frontend_catch_all_directory",
    prefixes: Object.freeze(["tools/harness/frontend/"]),
    exact: Object.freeze([]),
  }),
  Object.freeze({
    id: "legacy_scheduler_backend_adapters",
    prefixes: Object.freeze([]),
    exact: Object.freeze([
      "tools/harness/scheduler/adapters/backend.mjs",
      "tools/harness/scheduler/adapters/schedule-context.mjs",
    ]),
  }),
  Object.freeze({
    id: "legacy_scheduler_phase_slice_and_service_backed_helpers",
    prefixes: Object.freeze([]),
    exact: Object.freeze([
      "tools/harness/scheduler/check-service-backed-expansion.mjs",
      "tools/harness/scheduler/execution-dependencies.mjs",
      "tools/harness/scheduler/phase-slice-cli.mjs",
      "tools/harness/scheduler/phase-slice-plan.mjs",
      "tools/harness/scheduler/service-backed-schedule-manifest.mjs",
      "tools/harness/scheduler/service-backed-schedule-topology.mjs",
    ]),
  }),
  Object.freeze({
    id: "legacy_scheduler_duration_helpers",
    prefixes: Object.freeze([]),
    exact: Object.freeze([
      "tools/harness/scheduler/duration-baseline-cli.mjs",
      "tools/harness/scheduler/duration-baseline-drift-suite.sh",
      "tools/harness/scheduler/duration-drift.mjs",
      "tools/harness/scheduler/harness-smoke-durations-cli.mjs",
      "tools/harness/scheduler/service-backed-make-target-durations-cli.mjs",
      "tools/harness/scheduler/target-duration-baselines.mjs",
    ]),
  }),
  Object.freeze({
    id: "legacy_scheduler_process_and_evidence_drift_helpers",
    prefixes: Object.freeze([]),
    exact: Object.freeze([
      "tools/harness/scheduler/scheduler/process-executor.mjs",
      "tools/harness/scheduler/scheduler-event-order-drift-cli.mjs",
      "tools/harness/scheduler/scheduler-summary-timing-drift-cli.mjs",
    ]),
  }),
  Object.freeze({
    id: "legacy_execution_phase_runtime_and_node_registry",
    prefixes: Object.freeze([]),
    exact: Object.freeze([
      "tools/harness/execution/run-phase-common.sh",
      "tools/harness/execution/make-node-tools.mjs",
    ]),
  }),
]);
