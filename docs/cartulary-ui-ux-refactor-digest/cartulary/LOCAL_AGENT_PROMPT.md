# Local Agent Prompt

Perform a future Cartulary UI/UX refactor only when the user separately
authorizes implementation. Treat it as contract-preserving remediation, not a
re-theme or product redesign.

Work from the Cartulary repository root. Read, in order:

1. `AGENTS.md` (the only currently applicable agent-instruction file);
2. `docs/cartulary-ui-ux-refactor-digest/cartulary/START_HERE.md`;
3. `docs/cartulary-ui-ux-refactor-digest/cartulary/REPO_MAP.tsv`;
4. relevant rows from
   `docs/cartulary-ui-ux-refactor-digest/cartulary/OWNER_MAP.tsv`,
   `rules.tsv`, and `acceptance.tsv`.

The bundled files under
`docs/cartulary-ui-ux-refactor-digest/upstream/` are immutable offline advisory
material. Do not treat their original plugin-context commands as Cartulary
commands.

## Authority

Adopted subsystem authority applies only within each named scope:
`docs/extension-subsystem-nlspec.md`, `docs/graph_projection_nlspec.md`,
`docs/network-flow-activity-nlspec.md`,
`docs/opentelemetry-instrumentation-nlspec.md`,
`docs/report-composition-nlspec.md`, `docs/reporting-subsystem-nlspec.md`, and
`docs/testing-harness-nlspec.md`. The draft
`docs/reference-pack-subsystem-nlspec.md` is not adopted authority.

The exact Core paths are:

1. `docs/spec/00_document_set_status_and_precedence.md`;
2. `docs/spec/01_architecture_storage_and_view_contracts.md`;
3. `docs/spec/02_domain_model_schema_and_history.md`;
4. `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`;
5. `docs/spec/04_security_deployment_and_conformance.md`;
6. `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`.

Core 05 applies only to claim-bearing timed or fixture-sensitive publication.
`docs/design.md` governs observable design behavior inside its scope, and
`docs/domain.md` governs vocabulary and owner navigation inside its scope.
`docs/handoffs/cartulary_modular_refactor_planning_framework.md` is
subordinate planning support; its path examples are not code truth.
Current code/tests are implementation evidence, not automatic authority.

If owner documents conflict, report `BLOCKED: owner contradiction`. Do not
create routes, schemas, authorization rules, lifecycle transitions, field
membership, write behavior, compatibility policy, or a second design system
from generic UI advice.

## Repository boundaries

Use `REPO_MAP.tsv` as the detailed map. The verified headline boundaries are:

- `apps/web/src/app` / `web.application`: application shell and route state;
- `apps/web/src/workbook` / `web.workbook`: workbook shell and browser runtime;
- `packages/grid-adapter` / `package.grid_adapter`: sole direct
  `react-data-grid` integration;
- `contracts/view-schemas`: authored view-schema projections;
- `packages/protocol-ts/src/generated`: generated protocol/view-schema output;
- `packages/view-contracts/src/generated`: generated projection of authored
  view-schema inputs;
- `packages/view-contracts` / `package.view_contracts`: authored view adapter;
- `packages/ui-contracts` / `package.ui`: authored stable selectors and token
  facade;
- `packages/test-utils` / `package.test_utils`: reusable semantic test helpers;
- `apps/web/e2e` and `apps/web/e2e/support`: application-specific browser
  suites, fixtures, and choreography;
- `contracts/design/tokens.v1.json`: authored token/theme/density projection;
- `packages/ui-contracts/src/generated/design-tokens.ts`: generated token
  output.

Do not infer a functional package from owner ID `package.ui`; `packages/ui` is
absent and the owner routes to `packages/ui-contracts`. `docs/design.md` §3.11
owns semantic icon IDs, but a standalone implementation icon registry is not
present. Preserve that conditional boundary rather than inventing one.

Before editing, inspect `tools/generated_artifact_policy.json`. Do not hand-edit
`internal/gen`, `packages/protocol-ts/src/generated`,
`packages/view-contracts/src/generated`, `packages/ui-contracts/src/generated`,
or generated harness/task files. Update authored owner inputs and use:

```bash
make generate
make generate-drift
make generated-artifact-policy-check
```

## First inspection

Record branch, commit, dirty state, allowed scope, and whether behavior changes
are authorized. Revalidate the stack from `package.json`,
`pnpm-workspace.yaml`, `apps/web/package.json`, and package imports. The
localization snapshot found TypeScript/React/Vite, Vitest/Testing Library/jsdom,
Playwright, Biome, Lucide, and `react-data-grid`; it found no Tailwind or shadcn
dependency.

Inventory only the proposed seam: shell/work-area ownership, grid adapter,
tokens/density, active view and inspector contracts, creation capability,
focus/keyboard/paste/edit/conflict behavior, transaction/pending recovery,
tests, and authored/generated boundaries.

Treat the six concerns in `START_HERE.md` as current regression baselines.
Verify current behavior and exact owner clauses before calling any concern a
defect. The remediation handoff is implementation evidence only; it is not an
owner or an authorization for another product slice.

## Work one coherent slice

Map observable behavior to an exact owner and characterize risky behavior
before movement. Separate normative corrections from behavior-preserving
refactors. Keep vendor semantics in `packages/grid-adapter`, app/workbook state
under its verified `apps/web/src` owner, generated contract output read-only,
stable selectors in `packages/ui-contracts`, reusable semantic helpers in
`packages/test-utils`, and application browser choreography in
`apps/web/e2e/support`.

Classify every material upstream recommendation `ADOPT`, `ADAPT`, or `REJECT`.
Query only a targeted concern:

```bash
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "keyboard focus color only error feedback" --domain ux --json

PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "virtualized grid rerender focus async state" --stack react -n 8 --json
```

Never use upstream `--design-system`, `--persist`, or `--force`. Reject the
bundled Cybersecurity Platform visual defaults.

## Verification

Choose an exact owner ID from `REPO_MAP.tsv`, then run its current guidance:

```bash
make task-guide ROLE=module-author OWNER=web.application
make task-guide ROLE=module-author OWNER=web.workbook
make task-guide ROLE=module-author OWNER=package.grid_adapter
make task-guide ROLE=module-author OWNER=package.view_contracts
make task-guide ROLE=module-author OWNER=package.ui
make task-guide ROLE=module-author OWNER=package.test_utils
make task-guide ROLE=module-author OWNER=web.design
make task-guide ROLE=module-author OWNER=harness.browser
```

Use the returned focused slice command. Add only the public readiness targets
required by the seam:

```bash
make frontend-typecheck
make frontend-unit
make frontend-import-boundary-check
make browser-e2e-webserver-backed
make browser-e2e-a11y
make browser-e2e-visual
```

`harness.visual` is a verification/fixture owner but is not accepted by
`make task-guide`; its public entry point is `make browser-e2e-visual`.

Test stable semantic identities and production-relevant outcomes. Never bind
tests to literal specification prose, documentation paths, line positions,
hashes, formatting, DOM hierarchy, CSS classes, vendor coordinates, or package
icon names.

## Required handoff

For every changed behavior, report the exact owner clause, before/after
behavior, implementation paths, validation commands/results, retained evidence,
material upstream classification, deferred issues, and next safe slice. Do not
claim completion until every applicable
`docs/cartulary-ui-ux-refactor-digest/cartulary/acceptance.tsv` row is `PASS`,
`N/A` with rationale, or explicitly `BLOCKED`.
