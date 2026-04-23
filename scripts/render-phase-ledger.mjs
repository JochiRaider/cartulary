import { collectEntries, collectSupportGoEntries, loadManifest } from "./lib/phase-manifest.mjs";

const supportTargetDisplay = new Map([
  ["backend_unit", "backend-unit"],
  ["backend_integration_support", "backend-integration-support"],
]);
const supportTargetOrder = new Map([
  ["backend_unit", 0],
  ["backend_integration_support", 1],
]);

const phaseConfigs = {
  phase0: {
    title: "Phase 0 Coverage Ledger",
    introduction: [
      "This ledger is generated from `tools/phase0_test_map.json`. Update the manifest row metadata first, then regenerate this file.",
      "",
      "- Scope: infrastructure, deployment configuration, runtime roots, schema bootstrap, bootstrap-admin preflight, object-store reachability, and fail-closed startup only.",
      "- Normative owners: Core 01 `§1`; Core 01 `§3.3.5.1`; Core 04 `§5–§8`; Core 04 `§12`; Core 04 `§9.0.1`.",
      "- Authority: `tools/phase0_test_map.json` is the enforced Phase 0 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.",
      "- Browser E2E note: no Phase 0 browser-visible surface exists under `apps/web`, so authoritative `E-*` evidence lives on the real `cmd/server` process boundary.",
      "",
      "## Authoritative Execution",
      "",
      "- `backend-unit` selects authoritative `U-0-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase0 unit authoritative backend_unit`.",
      "- `backend-integration` selects authoritative `I-0-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase0 integration authoritative backend_integration`.",
      "- `backend-process` and `phase0-process-e2e` select authoritative `E-0-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase0 e2e authoritative backend_process`.",
    ],
    supportExecutionExtras: [],
    sections: [
      ["unit", "Unit"],
      ["integration", "Integration"],
      ["e2e", "Process E2E"],
    ],
    sharedHarness: [
      "| Harness | Phase 0 evidence |",
      "| --- | --- |",
      "| Startup diagnostics and real process boundary | `internal/testutil/diagnosticstest`, `internal/testutil/configtest`, and `internal/testutil/processtest` keep unit, integration, and process startup diagnostics on shared whole-payload goldens. |",
      "| Fail-closed HTTP readiness and health gating | `internal/testutil/processtest.WaitForReady`, `RequireStatus`, and `RequireConnectionRefused` prove `/healthz` and `/readyz` behavior across success and failure flows. |",
      "| Fail-closed WebSocket boundary | `internal/testutil/processtest.RequireWebsocketConnectionRefused`, built on `internal/testutil/wstest`, proves Phase 0 startup failures expose no WebSocket surface. |",
      "| Mutation attribution, secret-safe payloads, and bootstrap auth-binding shape | `internal/testutil/crosscutting` plus `internal/testutil/phase0test.RequireBootstrapUserLocalAuthOnly` cover startup audit attribution, secret-safe payloads, and bootstrap-created-user auth-binding shape. |",
      "| Real Postgres bootstrap harness | `internal/platform/postgres/postgres_phase0_test.go::TestPhase0_SchemaBootstrap_I_0_01` and the runtime integration suite prove migration and bootstrap state against real PostgreSQL. |",
      "| Real object-store binding harness | `internal/platform/objectstore/objectstore_phase0_test.go::TestPhase0_ObjectStoreInitialization_I_0_02` and the real process suite prove disconnected `filesystem_root` resolution through the generic `CARTULARY_S3_*` contract and live MinIO reachability. |",
    ],
    supportOnly: [
      "- `internal/platform/postgres/postgres_phase0_support_test.go` keeps the migration-text regression guard; authoritative schema-bootstrap evidence stays `I-0-01`.",
      "- `internal/platform/objectstore/objectstore_phase0_support_test.go` keeps managed-service object-store binding coverage; authoritative object-store reachability stays `I-0-02`.",
      "- `tools/testservices/integration_test.go` remains harness-development noise and is intentionally outside Phase 0 traceability.",
    ],
  },
  phase3: {
    title: "Phase 3 Coverage Ledger",
    introduction: [
      "This ledger is generated from `tools/phase3_test_map.json`. Update the manifest row metadata first, then regenerate this file.",
      "",
      "- Scope: Timeline workbook create, patch, query, lifecycle actions, replay stability, projection rebuild, save-state UI, and browser-visible collaboration behavior.",
      "- Normative owners: Core 03 `§6`, `§7`, `§15`; Core 01 `§3.3.5`, `§7.4.1`; Core 04 `AC-043`, `AC-191` through `AC-199`, `AC-329` through `AC-331`.",
      "- Authority: `tools/phase3_test_map.json` is the enforced Phase 3 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.",
      "- Timeline zero-field create traceability: cite Core 01 `REQ-01-057` plus Core 04 `AC-191` and `AC-192` for the owner rule. `contracts/view-schemas/cartulary.view.timeline.v1.json` is derived evidence only and is not the behavior source.",
      "- Grid note: the workbook currently renders through the RDG-backed `@cartulary/grid-adapter`. `U-3-GRID-01/02/03` own workbook binding behavior, while vendor-specific RDG semantics stay with adapter tests.",
      "",
      "## Authoritative Execution",
      "",
      "- `backend-unit` selects authoritative `U-3-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase3 unit authoritative backend_unit`.",
      "- `backend-store` selects store-backed authoritative `U-3-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase3 unit authoritative backend_store`.",
      "- `backend-integration` selects authoritative `I-3-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase3 integration authoritative backend_integration`.",
      "- `frontend-unit` selects authoritative `U-3-*` workbook rows only through the Phase 3 Vitest manifest for `frontend_unit`.",
      "- `browser-e2e-webserver-backed` and delegated `browser-e2e-functional` select authoritative functional `E-3-*` rows only through the Phase 3 Playwright manifest for `browser_functional`.",
      "- `browser-e2e-measurement` selects authoritative measurement `E-3-*` rows only through the Phase 3 Playwright manifest for `browser_measurement`.",
    ],
    supportExecutionExtras: [],
    sections: [
      ["unit", "Unit"],
      ["integration", "Integration"],
      ["e2e", "Browser E2E"],
    ],
    sharedHarness: [
      "| Harness | Phase 3 evidence |",
      "| --- | --- |",
      "| Real runtime, store, and socket harness | `internal/testutil/phase3test` centralizes the Postgres + MinIO runtime boot path, HTTP session helpers, incident or membership seeding, and Timeline WebSocket assertions shared by authoritative and support-only Phase 3 tests. |",
      "| Cross-cutting HTTP and replay helpers | `internal/testutil/httptestx` owns success or error envelope checks, replay scaffolding, mutation attribution helpers, and closed-vocabulary assertions used across the Phase 3 backend suite. |",
      "| Timeline substrate inspection helpers | `internal/testutil/timelinetest` owns projection-row, change-set, mutation-row, revision-count, and supersede-link inspection used by the Phase 3 store and integration slices. |",
      "| Browser timing and replay helpers | `apps/web/e2e/helpers.ts` provides the Phase 3 timing predicates, tracked-user browser auth helpers, and substrate snapshot accessors shared by `E-3-02`, `E-3-03`, and `E-3-04`. |",
    ],
    supportOnly: [
      "- `internal/modules/timeline/phase3_support_test.go` keeps helper-level regression coverage for request-shape helpers, vocabulary helpers, hash normalization, payload builders, and supersede guards. These tests run under `TestSupportPhase3Unit_` and are intentionally forbidden from carrying authoritative Phase 3 IDs.",
      "- `internal/modules/timeline/phase3_support_integration_test.go::TestSupportPhase3Integration_AuthorizationMatrix` table-drives create, query, patch, review, and supersede authorization across no-membership, editor, reviewer, and admin states. It strengthens route inventory confidence and does not replace `I-3-03`.",
    ],
  },
  phase4: {
    title: "Phase 4 Coverage Ledger",
    introduction: [
      "This ledger is generated from `tools/phase4_test_map.json`. Update the manifest row metadata first, then regenerate this file.",
      "",
      "- Scope: entity mentions, resolve and merge routes, entity-origin Host and Identity creation, indicator create and query surfaces, Timeline auto-resolution and manual relationship HTTP behavior, and the owned Phase 4 browser inspector flows.",
      "- Normative owners: Core 01 `§3.3.5`; Core 02 `§6` through `§10`; Core 03 `§9`, `§16`; Core 04 `AC-017`, `AC-019` through `AC-023`, `AC-077` through `AC-079`, `AC-186`, `AC-188` through `AC-190`, `AC-205`, `AC-209`, `AC-221` through `AC-225`, `AC-388` through `AC-397`.",
      "- Authority: `tools/phase4_test_map.json` is the enforced Phase 4 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.",
      "- Binding-mode note: the authoritative Phase 4 row set stays fixed. Contract-derived binding-mode, row-field, and projection-determinism assertions are absorbed into the existing authoritative rows instead of creating new Phase 4 IDs.",
      "",
      "## Authoritative Execution",
      "",
      "- `backend-unit` selects authoritative `U-4-*` decoder rows only through `RUN_GO_MANIFEST_PHASE ... phase4 unit authoritative backend_unit`.",
      "- `backend-store` selects authoritative store-backed `U-4-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase4 unit authoritative backend_store`.",
      "- `backend-integration` selects authoritative `I-4-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase4 integration authoritative backend_integration`.",
      "- `browser-e2e-webserver-backed` and delegated `browser-e2e-functional` select authoritative `E-4-*` rows only through the Phase 4 Playwright manifest for `browser_functional`.",
    ],
    supportExecutionExtras: [
      "- `apps/web/src/WorkbookShell.phase4.support.test.tsx` runs through `frontend-unit` and is forbidden from claiming `U-4-*`, `I-4-*`, or `E-4-*` identifiers.",
    ],
    sections: [
      ["unit", "Unit"],
      ["integration", "Integration"],
      ["e2e", "Browser E2E"],
    ],
    sharedHarness: [
      "| Harness | Phase 4 evidence |",
      "| --- | --- |",
      "| Real runtime and route helpers | `internal/testutil/phase4test` centralizes the Postgres + MinIO runtime boot path, bootstrap-admin login helpers, entity or mention seed helpers, contract-derived field-surface helpers, and the inventory-driven support matrix metadata for every owned Phase 4 HTTP route. |",
      "| Cross-cutting HTTP and projection helpers | `internal/testutil/httptestx` owns success or error envelope checks, replay scaffolding, authorization re-derivation assertions, closed-vocabulary checks, field-key conformance, and projection determinism assertions shared across the Phase 4 backend suite. |",
      "| Entity and Timeline substrate inspection | `internal/testutil/assertx`, `internal/testutil/timelinetest`, and Phase 4 package-local lookup helpers inspect durable mention, link, change-set, projection, and observation state that the authoritative Phase 4 rows rely on. |",
      "| Browser helper fixtures | `apps/web/src/timelineWorkbookTestSupport.tsx` and `apps/web/src/workbookShellPhase4.ts` provide the mocked workbook row, websocket, and mention helper scaffolding used by the support-only Phase 4 workbook tests. |",
    ],
    supportOnly: [
      "- `internal/modules/entities/phase4_support_integration_test.go` consumes the centralized Phase 4 route inventory through `TestSupportPhase4Integration_SurfaceEnvelope`, `..._CSRFProtection`, `..._ReplayAndDivergentConflict`, `..._AuthorizationReDerivation`, `..._DefaultQueryMetaAndFieldKeyConformance`, and `..._ProjectionAndWebsocketConsequences`. Together they keep the resolve, merge, entity-origin, indicator, and Timeline routes on one enforced support-only matrix without replacing any authoritative `I-4-*` row.",
      "- `apps/web/src/WorkbookShell.phase4.support.test.tsx` keeps mocked helper and component regression coverage for Phase 4 workbook chips, payload builders, inspector mention derivation, and auto-resolution notices. It remains support-only and is not completion evidence for Phase 4.",
    ],
  },
};

function renderEvidence(entry) {
  if (entry.runner === "go_test") {
    return `\`${entry.file}::${entry.symbol}\``;
  }
  return `\`${entry.file}::${entry.title}\``;
}

function renderExecution(entry) {
  return `\`${entry.execution_dependency}\``;
}

function phaseNumberFromKey(phase) {
  const match = /^phase(\d+)$/.exec(phase);
  if (!match) {
    throw new Error(`unsupported phase key ${phase}`);
  }
  return match[1];
}

function sectionPrefix(section) {
  switch (section) {
    case "unit":
      return "U";
    case "integration":
      return "I";
    case "e2e":
      return "E";
    default:
      throw new Error(`unsupported support section ${section}`);
  }
}

function renderSupportExecutionLines(phase, manifest, extras = []) {
  const phaseNumber = phaseNumberFromKey(phase);
  const supportLines = collectSupportGoEntries(manifest)
    .sort((left, right) => {
      const orderDiff =
        (supportTargetOrder.get(left.target) ?? Number.MAX_SAFE_INTEGER) -
        (supportTargetOrder.get(right.target) ?? Number.MAX_SAFE_INTEGER);
      if (orderDiff !== 0) {
        return orderDiff;
      }
      return left.file.localeCompare(right.file);
    })
    .map((entry) => {
      const target = supportTargetDisplay.get(entry.target);
      if (!target) {
        throw new Error(`unsupported support target ${entry.target}`);
      }
      return `- \`${entry.file}\` runs through \`${target}\` with \`${entry.selection_pattern}\` and is forbidden from claiming \`${sectionPrefix(entry.section)}-${phaseNumber}-*\` identifiers.`;
    });
  return [...supportLines, ...extras];
}

function renderSection(title, entries) {
  const lines = [
    `## ${title}`,
    "",
    "| Row | Evidence | Execution | Claim | Out of scope |",
    "| --- | --- | --- | --- | --- |",
  ];

  for (const entry of entries) {
    lines.push(
      `| \`${entry.id}\` | ${renderEvidence(entry)} | ${renderExecution(entry)} | ${entry.claim} | ${entry.out_of_scope} |`,
    );
  }

  return lines;
}

export function renderPhaseLedger(root, phase) {
  const config = phaseConfigs[phase];
  if (!config) {
    throw new Error(`unsupported phase ledger render: ${phase}`);
  }

  const { manifest } = loadManifest(root, phase);
  const entries = collectEntries(manifest).filter(
    (entry) => entry.coverage === "authoritative",
  );

  const lines = [`# ${config.title}`, "", ...config.introduction];
  const supportExecutionLines = renderSupportExecutionLines(
    phase,
    manifest,
    config.supportExecutionExtras,
  );
  if (supportExecutionLines.length > 0) {
    lines.push("", "## Support-Only Execution", "", ...supportExecutionLines);
  }

  for (const [sectionKey, title] of config.sections) {
    const sectionEntries = entries.filter((entry) => entry.section === sectionKey);
    lines.push("", ...renderSection(title, sectionEntries));
  }

  lines.push("", "## Shared Harness Coverage", "", ...config.sharedHarness);
  lines.push("", "## Support-Only Evidence", "", ...config.supportOnly);

  return `${lines.join("\n")}\n`;
}

function main(argv) {
  const [phase] = argv;
  if (!phase) {
    throw new Error("usage: render-phase-ledger.mjs <phase>");
  }
  process.stdout.write(renderPhaseLedger(process.cwd(), phase));
}

if (import.meta.url === `file://${process.argv[1]}`) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exit(1);
  }
}
