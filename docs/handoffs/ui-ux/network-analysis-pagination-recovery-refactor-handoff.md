# Network Analysis pagination and cursor recovery

## Baseline and authority

- Branch: `main`; HEAD: `aac7a517a71aa1dae5501f48dce0c61c9bf9afb9`.
- Initial working tree: clean. No commits, resets, pushes or deployments.
- Adopted owners: Network Flow Activity 6.0.0 §§7, 13, 14.6, 17.8, 19,
  21, 26–28; Core 03 availability lifecycle; Core 04 cursor protection and
  authorization; Extensions client cleanup; design §§8, 10, 13.1 and 14.
- The localized digest was read in its prescribed order during planning and
  paths revalidated against this HEAD. Historical QA/TL handoffs are evidence.
- Allowed paths: Network Flow frontend sources and tests, its service adapter,
  relevant browser/support tests, Network Flow backend route tests, authored
  verification catalogs and their Make-generated projections, required visual
  artifacts, and this handoff. Production backend/contracts require a demonstrated
  owner mismatch. Saved-graph pagination and other completed seams stay bounded.
- Both requested owner task guides passed. Existing pager characterization
  passed at `.cartulary/test-results/20260910T023210Z-p53987` (2/2 units).

## Workstreams

| Workstream | Status | Exit |
| --- | --- | --- |
| PR-01 Baseline, owner/error matrix, gaps, characterization | DONE | Reproduced defects and routed tests |
| PR-02 Committed pages and captured navigation | DONE | Atomic navigation, admission and cancellation tests |
| PR-03 Recovery, contributors and invalidation | DONE | Recovery, correlation and real route tests |
| PR-04 Presentation, accessibility and browsers | DONE | Browser, focus, announcement and visual evidence |
| PR-05 Final verification and scope audit | DONE | Finalizer, owner checks, scope audit and completed handoff |

All workstreams are DONE. No dependency or owner decision remains blocked.

## Decisions and gap ledger

| Gap | Governing boundary | Repair |
| --- | --- | --- |
| Navigation changes page/history before success | NF §7 semantic identity; requested atomic transitions | Commit items, metadata, producing request and position together |
| Render-state loading guard admits same-turn commands | Requested deterministic navigation | Synchronous first-command admission; explicit Restart supersedes |
| Recovery reads mutable initial request | NF-REQ-054a | Capture initial request, context, route and callbacks per attempt |
| Error decoding ignores returned retry action | NF §21 | Validate and honor structured action; protected-state precedence |
| Contributor paging duplicates premature transitions | NF §§13–14, 26–28 | Share narrow paging mechanics, preserve server selectors |
| Contributor stale failure erases graph without recovery | NF §19 | Enter graph-owned stale/recompute boundary |
| Reads and callbacks have incomplete authority fences | NF-REQ-054a; Core 03/Extensions | Current read identity and queued-dispatch/acceptance checks |
| Query adapters discard metadata and lack correlation | NF §17.8 | Validate before publishing, preserve complete page metadata |
| Page feedback and Retry conflate attempts | Design local feedback; requested recovery behavior | Separate committed page, destination and typed recovery action |
| Replacement selection can retarget by row position | NF §7.3; requested pagination boundary | Retain visible semantic identity, otherwise clear pagination selection |

Expiry permits one automatic captured-query restart. Failed restart retains only
authorized committed exposure and never reuses the invalid cursor chain. Actor,
session or authorization loss clears protected exposure before owner recovery;
source invalidation refreshes its resource owner; graph staleness recomputes the
graph; request correction, scope reduction and no-retry remain distinct.

Public major 6 and wire/storage interfaces remain unchanged. No migration is
planned. Read recovery has no mutation transaction or receipt machinery.

## Execution record

### PR-01

Baseline revalidated unchanged and clean. Added two routed characterizations in
the existing pager suite. `make test-slice OWNER=web.networkflow
ROWS=web.networkflow.regression.usenetworkflowpagedquery_suite_7dc8b8d06a`
failed as expected (1/2 execution units) at
`.cartulary/test-results/20260910T023915Z-p57101`: failed navigation relabels the
committed page and same-turn commands dispatch repeatedly. These failures are
directly related to this seam. No compatibility or owner decision is blocked.
PR-02 now implements committed snapshots and synchronous admission; additional
recovery and boundary characterization belongs to its respective workstream.

### PR-02

Replaced scattered pager state with one committed snapshot, request-only history,
pending and failed attempt descriptors. Admission is synchronous and first-command
wins, with explicit Restart cancellation and duplicate restart coalescing.
Previous reads the producing request; page/history updates occur only on accepted
responses. Bounded observation uses the existing 30-second helper. Added distinct
retry/refresh tests, explicit restart races and empty-page/deadline coverage.

The pager slice passed at `.cartulary/test-results/20260910T024145Z-p58225`
and, after expanded tests, `.cartulary/test-results/20260910T024232Z-p58903`
(2/2 execution units each). Frontend interface integration is intentionally
completed in PR-03/04: page number is now absent before an accepted response.
Next: route-bound authority, response correlation and contributor recovery.

### PR-03

Rows, diagnostics and exploration contributors now share the committed pager.
Table authority supplies an independent live read identity; queued dispatch,
accepted callbacks and recovery check it. Graph composition still owns stale
recomputation and scoped source invalidation. Contract adapters reject mismatched
resources, selectors, digests, paging counts and changed continuation metadata.
Malformed envelopes are distinct from transport failures. Unapplied drafts retain
their existing ownership.

The complete web owner slice passed 60/60 units at
`.cartulary/test-results/20260910T030306Z-p49502`. Focused recovery/authoring and
lifecycle reruns also passed. Interim failures caught graph effect ordering,
a duplicated access-loss callback, and a stale callback test counting a legitimate
new read; these were repaired. Frontend typecheck passed at
`.cartulary/test-results/20260910T025347Z-p64968`. Formatting exposed a missing
effect dependency, repaired by reading the captured initial request from the
existing latest-context ref; subsequent type and lint checks passed as recorded below.

Real service-backed pagination, contributor and source-lifecycle rows passed
3/3 units at `.cartulary/test-results/20260910T030422Z-p56673`. The new route test
asserts actual ordered resources, producing-request replay, terminal pages,
validity at 899 seconds and expiry at 900 seconds for all three routes.
Initial fixture failures (duplicate diagnostic IDs and the wrong test clock,
then clock movement before two-factor login) were test setup defects, corrected
using the harness-owned server clock after authentication.

The real expiry run at `.cartulary/test-results/20260910T030244Z-p32313`
demonstrated an owner mismatch: cursor errors omitted `details.retry_action`.
NF Table 21-B1 requires `restart_query`; the bounded production repair is in
`internal/modules/networkflow/query.go`, adding that field in the shared cursor
error constructor. This additive correction also reaches existing cursor consumers;
no request schema, public major, storage or saved navigation changes are needed.
Allowed production scope is extended to that constructor only.

Next: truthful local page/recovery controls, semantic selection and browser proof.

### PR-04

Added shared page feedback and compact Previous, Next, Refresh page and recovery
controls. The committed position remains visible during pending navigation;
failures name the destination separately. Recovery distinguishes retry,
restart, resource refresh and correction guidance. Graph staleness exposes
Recompute graph. Pagination retains visible semantic selection and clears vanished
resources without moving focus from its initiating control. Saved navigation keeps
its existing grid behavior.

Browser evidence exposed a second owner mismatch at
`.cartulary/test-results/20260910T031259Z-p97147`: omitted query limits returned
100 rather than NF Table 13-A's `min(200, effective maximum)`. The bounded repair
extends the permitted `query.go` change to `defaultQueryLimit`, with assertions
for maxima 50, 150 and 1000 in the existing contract-boundary test. Its routed row
passed at `.cartulary/test-results/20260910T031648Z-p70046` (1/1).

Presentation and selection rows passed at
`.cartulary/test-results/20260910T031553Z-p36961` (3/3); the web sweep passed
62/62 at `.cartulary/test-results/20260910T031944Z-p45094`. An expanded matrix
now fences late row/diagnostic success, expiry and denial after incident, actor,
session, role and availability changes; the focused row passed at
`.cartulary/test-results/20260910T032302Z-p20477` (2/2). Source-loss handling in
the graph refreshes the authoritative table list without assuming the active
table was the failed contributor source.

Accessibility and both relevant measurement rows passed 15/15 execution units
at `.cartulary/test-results/20260910T031712Z-p71149`. The first pagination
browser attempts exposed test alert ambiguity, native disabled-button focus loss,
and a fixture advancing time into session expiry. The alert is scoped to its
grid; pagination now uses focusable `aria-disabled` controls backed by synchronous
command admission; independent cursor chains are reissued within session lifetime.
These observations retain actual response and resource assertions. A batch at
`.cartulary/test-results/20260910T032338Z-p23457` stopped before browser execution
because formatting changed captured frontend build inputs; reruns freeze sources.

The ordinary visual row at `.cartulary/test-results/20260910T031944Z-p45129`
completed functional assertions and found four intended screenshot differences:
accepted inspector, rejected diagnostics, graph contributors and narrow query
controls. Each actual image was inspected; changes follow the added local page
controls, including wrapping inside the contributor drawer. Reconciliation
accounts for 210 committed goldens, nine selected active captures, zero missing
or ambiguous mappings, and all registered fixtures. No unselected goldens are
being removed. Promotion and both ordinary post-promotion passes are recorded below.

`make generate` refreshed the authored-catalog projections. An initial
`make agent-finalize` failed because those projections were stale; generation
resolved it and the finalizer passed at
`.cartulary/test-results/20260910T031259Z-p97125`. `RESULTS_DIR` is unset:
retained-run and performance-baseline maintenance are explicitly skipped because
no exact-source successful full warm run qualifies.

The combined service/browser batch passed 20/20 execution units at
`.cartulary/test-results/20260910T032658Z-p67759`: real expiry in both directions
on all three surfaces, returned resource order, rename continuity, deletion,
authoring, table lifecycle, accessibility and rendering bounds. A focused browser
extension now exercises the visible Recompute graph action and fresh contributors;
it passed 13/13 units at `.cartulary/test-results/20260910T032939Z-p15752` and
again at `.cartulary/test-results/20260910T033512Z-p21203` after error-path review.
HTTP 401/403 withdraw protected data even when an error body is malformed;
malformed cursor-error bodies cannot authorize automatic restart. Pagination
failure alerts retain priority without a duplicate polite failure notice.

The final queued-graph characterization failed as expected at
`.cartulary/test-results/20260910T033954Z-p89533`: when live authority changed
before dispatch without a new render, the stopped graph remained loading.
The graph owner now settles only that still-current stopped read to idle;
newer graph reads, callbacks and errors remain fenced. Its full web-owner rerun passed 62/62 units at
`.cartulary/test-results/20260910T034100Z-p94779`.

#### Visual refresh record

The accepted trigger is the requested local pagination/recovery controls and
semantic focus continuity. The full ordinary visual run
`.cartulary/test-results/20260910T033001Z-p46358` accounted for all 210 active
captures and committed goldens, with no orphan, missing or ambiguous mapping;
only the four expected Network Analysis screenshots differed. All other visual
rows passed. The public update target has no narrow row-selection input, so the
complete visual inventory was exercised.

`make browser-e2e-visual-update` passed 12/12 units at
`.cartulary/test-results/20260910T033511Z-p19868`. Only these four files under
`apps/web/e2e/workbook.visual.spec.ts-snapshots/` and the Make-owned
`tools/frontend_visual_golden_manifest.json` changed:

- `network-flow-analysis-accepted-inspector-linux.png`
- `network-flow-analysis-rejected-diagnostics-linux.png`
- `network-flow-analysis-graph-contributors-linux.png`
- `network-flow-analysis-narrow-query-controls-linux.png`

Owner row: `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`;
scenario: `scenario_252dd2b0ff55`; fixtures:
`visual.fixture.claimed_network_analysis_workspace_states` and
`visual.fixture.claimed_network_analysis_narrow_workspace`. Every promoted image
was inspected. Controls stay within the existing footer, wrapping in the narrow
contributor drawer; grid identity, content, typography and surrounding chrome
remain intact. Viewport, zoom, masks, scroll normalization, screenshot scope,
fixtures, assertions and tolerances were not changed. Both required ordinary visual passes are recorded below.

After the last source change, `make frontend-typecheck` and `make lint-biome`
passed at `.cartulary/test-results/20260910T034100Z-p94909` and
`.cartulary/test-results/20260910T034100Z-p94939`; `make agent-finalize` passed at
`.cartulary/test-results/20260910T034100Z-p94774`. The final combined
service/browser/accessibility/measurement selection passed 20/20 units at
`.cartulary/test-results/20260910T034242Z-p6445`.

The first fresh ordinary post-promotion `make browser-e2e-visual` passed all
12 execution units (35 rows, 210 captures) at
`.cartulary/test-results/20260910T034242Z-p6557`. The second fresh ordinary pass also passed 12/12 units at
`.cartulary/test-results/20260910T034711Z-p83601`. PR-04 is complete.

### PR-05

Final state ownership remains local and explicit: the shared pager owns one
committed page, producing-request history and captured attempts; the table owner
supplies read authority and scoped source lifecycle; the graph owner owns stale
composition and recomputation. Route clients and the existing adapter validate
transport/response correlation. Query authoring retains its draft and Apply
revisions; semantic grids retain visible resources and fields and clear vanished
pagination selection. Saved graphs keep their existing navigation owner.

All gap-ledger repairs are implemented and verified. Public major 6, route shapes,
storage and dependencies are unchanged. The only backend production changes are
the two demonstrated owner repairs in `query.go`: omitted limits use the adopted
200-row default bounded by the configured maximum, and cursor errors include the
required retry action. Existing cursor tokens retain their signed query and limit;
no migration, key rotation or cursor reset is required. Rollback concerns are
frontend page/recovery behavior and associated tests/goldens; reverting the two
backend corrections would restore their documented owner mismatches.

Final boundary/catalog/drift checks passed. There is no unresolved owner
contradiction or missing product decision. The implementation is complete;
post-write Markdown, whitespace and scope checks are repeated after this record.

#### Final verification ledger

All commands ran from the repository root through public Make targets. Run roots
below are relative to `.cartulary/test-results/`; earlier characterization and
interim failures are retained in their workstream records.

| Command or routed selection | Result | Run root |
| --- | --- | --- |
| `make agent-finalize` with `RESULTS_DIR` unset | PASS 1/1; generated structure unchanged | `20260910T035238Z-p19083` |
| `make test-slice OWNER=web.networkflow` | PASS 62/62 | `20260910T034100Z-p94779` |
| `make frontend-typecheck` | PASS 2/2 | `20260910T034100Z-p94909` |
| `make lint-biome` | PASS 2/2 | `20260910T034100Z-p94939` |
| Combined service/browser selection below | PASS 20/20 | `20260910T034242Z-p6445` |
| `make browser-e2e-visual`, first ordinary post-promotion pass | PASS 12/12; 35 rows, 210 captures | `20260910T034242Z-p6557` |
| `make browser-e2e-visual`, second ordinary post-promotion pass | PASS 12/12; 35 rows, 210 captures | `20260910T034711Z-p83601` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260910T035553Z-p23261` |
| `make backend-module-boundary-check` | PASS 3/3 | `20260910T035553Z-p23267` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260910T035553Z-p23082` |
| `make json-shape-check` | PASS 3/3 | `20260910T035553Z-p23088` |
| `make test-catalog-check` | PASS, exit 0 | Same final Make invocation as JSON/drift checks |
| `make generate-drift` | PASS 4/4 | `20260910T035553Z-p23077` |
| Protocol boundary selection below | PASS 2/2 | `20260910T035553Z-p23153` |
| Backend query boundary selection below | PASS 1/1 | `20260910T035553Z-p23174` |
| `make lint-go-format lint-go-vet` | PASS, exit 0 | Direct public targets; no separate run root |
| `make lint-markdown` | PASS; repeated after final tracker write | `20260910T034259Z-p65422` |
| `git diff --check` and branch/HEAD/status/scope inspection | PASS; repeated after final tracker write | Working tree |

The exact combined service-backed command was:

```sh
make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.browser_stateful.pagination_recovery,module.networkflow.browser.query_authoring_integrity,module.networkflow.browser.table_lifecycle_continuity,module.networkflow.integration.pagination_recovery,module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.accessibility.verify_claimed_network_analysis_keyboard_focus_a_c81856fdc1,module.networkflow.measurement.measure_bounded_dom_growth_and_semantic_reachabi_297b6a7322,module.networkflow.measurement.network_flow_selector_covers_the_actual_accepted_f68f4c7a57
```

The final narrow protocol/backend selections were:

```sh
make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.boundary_support.index_decodes_exact_network_flow_contracts_witho_8fee657f09
make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_selector_covers_a_table_query_uses_6caf43b23c
```

Additional interim static failures were related to implementation in progress:
optional/null interface mismatches in typecheck at `20260910T024659Z-p60544`,
an effect dependency in Biome at `20260910T030222Z-p31781`, and three browser
test non-null assertions at `20260910T032658Z-p67920`. Each was repaired without
weakening assertions; final checks above passed. Stale catalog projections made
the finalizer fail at `20260910T031124Z-p82132`; Make-owned generation repaired
them, and the final finalizer made no additional changes.

Retained-run, scheduler-timing and performance-baseline maintenance were skipped
because `RESULTS_DIR` was unset and no successful exact-source full warm run
qualifies. Full `make check`, CI and release checks were not selected: the final
coverage uses the touched owners, protocol/backend boundaries and required real
browser, accessibility, measurement and full visual rows. No release, conformance
or performance-baseline claim is made. Tests, generators and runtime evidence
remain independent of Markdown.

#### Changed paths and scope

All 33 changed paths belong to this seam. The initial tree was clean, and no
unrelated changes were modified. The two authored verification catalogs generated
the browser/topology projections through Make; the visual update target generated
the golden manifest. No generated protocol code or lockfile was edited.

```text
apps/web/e2e/network-flow-pagination.spec.ts
apps/web/e2e/network-flow.spec.ts
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-accepted-inspector-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-graph-contributors-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-narrow-query-controls-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-rejected-diagnostics-linux.png
apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx
apps/web/src/networkFlow/NetworkAnalysisWorkspace.tsx
apps/web/src/networkFlow/NetworkFlowControls.tsx
apps/web/src/networkFlow/NetworkFlowQueryPagination.test.tsx
apps/web/src/networkFlow/NetworkFlowQueryPagination.tsx
apps/web/src/networkFlow/NetworkFlowSemanticGrid.test.tsx
apps/web/src/networkFlow/NetworkFlowSemanticGrid.tsx
apps/web/src/networkFlow/NetworkFlowTableController.ts
apps/web/src/networkFlow/networkFlowClient.ts
apps/web/src/networkFlow/networkFlowErrors.test.ts
apps/web/src/networkFlow/networkFlowErrors.ts
apps/web/src/networkFlow/networkFlowPageRecovery.test.tsx
apps/web/src/networkFlow/useNetworkFlowGraphController.ts
apps/web/src/networkFlow/useNetworkFlowPagedQuery.test.tsx
apps/web/src/networkFlow/useNetworkFlowPagedQuery.ts
apps/web/src/networkFlow/useNetworkFlowRejectedRowsController.ts
apps/web/src/networkFlow/useNetworkFlowRowsController.ts
apps/web/src/services/networkFlowContractAdapter.ts
docs/handoffs/network-analysis-pagination-recovery-refactor-handoff.md
internal/modules/networkflow/network_flow_contract_test.go
internal/modules/networkflow/query.go
internal/modules/networkflow/routes_integration_test.go
tools/browser_e2e_batch_manifest.json
tools/execution_topology_render_index.json
tools/frontend_visual_golden_manifest.json
tools/test_families/module.networkflow.json
tools/test_families/web.networkflow.json
```

Final branch and HEAD remain `main` at
`aac7a517a71aa1dae5501f48dce0c61c9bf9afb9`. Changes are uncommitted.
No routes, persistence, dependencies, migrations, commits, resets, pushes or
deployments were introduced. No further implementation action is pending.
