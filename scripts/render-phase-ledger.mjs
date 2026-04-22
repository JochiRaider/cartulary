import { collectEntries, loadManifest } from "./lib/phase-manifest.mjs";

const phaseConfigs = {
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
      "",
      "## Support-Only Execution",
      "",
      "- `internal/modules/timeline/phase3_support_test.go` runs through `backend-unit` with `TestSupportPhase3Unit_` and is forbidden from claiming `U-3-*` identifiers.",
      "- `internal/modules/timeline/phase3_support_integration_test.go` runs through `backend-integration-support` with `TestSupportPhase3Integration_` and is forbidden from claiming `I-3-*` identifiers.",
    ],
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
