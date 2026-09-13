# Network Analysis query authoring refactor handoff

## Baseline and authority

- Branch: `main`; HEAD: `e5d01e1f66158ac51d351564cf2c5f099c30938d`.
- Initial tracked and untracked state: clean. Only root `AGENTS.md` applies.
- Adopted owners: Network Flow Activity 6.0.0 §§6–7, 13–14, 17.7, 19,
  26–28; Core 03 REQ-03-011A and REQ-03-303; design §§12–14.
- Digest localized read order was followed and mappings checked against HEAD.
  Digest, research, and completed handoffs are advisory evidence, not authority.
- Verification owners: `web.networkflow`, `module.networkflow`,
  `package.protocol_ts`, with UI, design, and harness collaborators as touched.
- Permitted paths: Network Flow frontend and service adapters; narrow workspace
  context wiring; Network Flow authored contracts and their generators; backend
  query admission/comparison/error boundaries; relevant UI selectors, tests,
  routing, reviewed goldens, and this handoff. Generated outputs change only by
  Make-owned generation. No unrelated changes existed to preserve.
- Excluded: pagination/retry redesign, graph algorithms, saved-graph authoring,
  import workflows, new persistence/routes/capabilities, migrations, dependency
  changes, digest/DU edits, commits, reset, push, or deployment.

## Ordered tracker

| Workstream | Status | Exit |
| --- | --- | --- |
| QA-01 Baseline, owner audit, ledger, characterization | DONE | Owner mapping and reproducible gaps |
| QA-02 Typed draft, reconstruction, compilation | DONE | Pure model and protocol verification |
| QA-03 Controls, scope, application continuity | DONE | Workspace/controller verification |
| QA-04 Service, browser, accessibility, visual evidence | DONE | Real result and regression verification |
| QA-05 Final verification and handoff | DONE | Finalization, evidence, scope audit |

Only one row may be IN_PROGRESS. A BLOCKED dependency prevents advancement.

## Decisions

One Apply action stages graph scope/aggregation with accepted filters and time.
Accepted and diagnostic drafts remain separate and survive mode/table changes
within the incident. Incident/actor/session replacement and authorization loss
purge protected state. A submitted attempt is separate from the last successful
applied query. Rejection retains exact draft values. Clear is immediate and
restores empty accepted filters/time/sort, active-table scope and default graph
aggregation; diagnostic Clear affects diagnostics only. No persistence is added.

## Bounded field and request ledger

NF Table 13-D/E and §6.4 govern accepted values; §13.3 governs conjunction and
normalization. Presentation v2 governs the offered subset, not public grammar.
All accepted errors route through `network_flow_invalid_filter` with safe
field/operator/index context; local draft errors precede dispatch.

| Field(s) | Adopted operators/value | Initial control and gap | Required wire and reconstruction | Verification |
| --- | --- | --- | --- | --- |
| src_ip, dst_ip | eq/in/cidr_contains; IPv4/IPv6 or CIDR | Advanced universal text parser | Scalar string or individual string list; preserve IP family and every predicate | Model, protocol, real rows |
| endpoint_ip | eq/in/cidr_contains; either endpoint | Basic scalar absorbs lists | Basic only for complete scalar predicate; lists and extras remain advanced | Round trip and real rows |
| src_port, dst_port | eq/in/range/is_null/not_null; 0..65535 | Comma parser and Number coercion | Integer scalar/list or inclusive gte/lte; null tests omit value | Model, protocol, service |
| ip_protocol | eq/in/range; 0..255 | Basic scalar absorbs in and erases it | Basic eq only; retain lists/ranges explicitly; offer existing eq/in subset | Add/apply/reapply browser |
| flow_start_utc, flow_end_utc | range only; UTC timestamps | eq unsupported; lt escapes range compilation | Half-open gte/lt, at least one bound; preserve microseconds | Boundary and overlap tests |
| bytes_count, packets_count | eq/range; uint64 decimal strings | Basic lower bound absorbs equality/upper bound | Exact decimal scalar or inclusive gte/lte; only lower-only range fits basic | uint64 extremes, real results |
| exporter_id, input_interface, output_interface | eq/in/prefix/contains/is_null/not_null; bounded text | Current offered eq/in/null subset uses ambiguous parsing | Preserve scalar text/list members; never numeric-compare text | Typed lists, canonical comparison |
| source_row_number | eq/in/range; positive integer | Not offered in add menu | Preserve valid reconstructed predicate; reject unsafe numeric conversion | Model and protocol |
| tcp_flags, application_label | Not filterable or sortable under §13 | Presentation incorrectly offers filtering/sorting | Remove unsupported query eligibility, retain columns | Projection and grid regression |
| Row overlap window | §13 timestamp filters implementing §14.2 overlap | Independent draft strings without validation | flow_end gte start AND flow_start lt end; no epsilon | Zero-duration and boundary rows |
| Graph scope/time/aggregation | §13.2 and §§26–28 semantic query v2 | Disconnected select; graph settings apply independently | One captured scope/time/aggregation; use server selectors/digest | Atomic dispatch, temporal results |
| Diagnostic error_codes/field_keys | §17.7 and Table 9-D/§21 token arrays | Comma split drops empties; only first field retained | Separate typed lists, no duplicates/unknown tokens | Diagnostic model/service/browser |
| Diagnostic source_row_range | §17.7 positive integer gte/lte | Number coercion; server accepts empty range | Inclusive nonempty range; retain raw rejected bounds | Model and service |

At baseline, advanced entries suppressed exact duplicates using raw JSON equality.
Compilation now rejects canonical duplicates visibly, retaining every entry. All conjunctive
same-field predicates survive unrelated edits. Server admission now includes
field-aware text comparison and complete diagnostic token/range checks.

## Execution evidence

| Stage | Command or inspection | Result |
| --- | --- | --- |
| Planning baseline | Both requested owner task guides; protocol task guide | PASS; narrow slices identified |
| Planning baseline | Query-model routed slice | PASS at `.cartulary/test-results/20260910T002541Z-p54315` |
| Implementation start | Branch, HEAD, status, tracker absence | Baseline unchanged; clean; tracker created |
| QA-01 | Routed query-control regression | Expected failure at `.cartulary/test-results/20260910T003204Z-p57042`: unchanged Apply loses protocol list/counter equality/upper bound |
| QA-02 | `make generate` | PASS at `.cartulary/test-results/20260910T003841Z-p62633`; typed filter variants and query metadata generated |
| QA-02 | Routed compilation and existing query-model slices | PASS at `.cartulary/test-results/20260910T004131Z-p66122` |
| QA-02 | Backend query-admission slice | PASS at `.cartulary/test-results/20260910T004308Z-p67138` |
| QA-02 | Protocol Network Flow decoder slice | PASS at `.cartulary/test-results/20260910T004329Z-p68113` |
| QA-02 | `make frontend-typecheck` during transition | Failed at `.cartulary/test-results/20260910T003733Z-p61833`: generated closed types expose old control casts; intermediate model range/token issues repaired. Full typecheck repeats after QA-03 replaces controls. |

## Implementation decisions and compatibility

No owner contradiction has been found. Projection repairs do not expand the
public API or change schema IDs, graph identity, persistence, or migrations.
Rollback will revert coordinated source, projections, routing, and reviewed
visual evidence. Browser and visual evidence is recorded below; no data migration is involved.

QA-02 owns exact scalar parsing in `networkFlowQueryValues.ts` and draft/wire
translation in `networkFlowQueryModel.ts`. Schema generation also updates
import contract digests because the analytical binding depends on the Network
Flow schema bundle; these are required derived compatibility artifacts.

QA-03 uses one volatile workspace authoring hook. Row, diagnostic, and graph
controllers execute captured queries with revision fences and report initial
success/rejection without changing attempt identity. Captured drafts attribute
errors without moving them onto newly edited predicates. Navigation restores
committed requests while retaining separate drafts; authority replacement purges
them. Scope and aggregation now live in the query band and Apply atomically.
Saved graph declarations remain independent. Clear resets immediately.

QA-03 evidence: frontend owner slice passed 59/59 units at
`.cartulary/test-results/20260910T010055Z-p97090`; typecheck passed at
`.cartulary/test-results/20260910T010057Z-p97345`. The final focused query,
application, and temporal interaction slice passed 5/5 at
`.cartulary/test-results/20260910T010308Z-p15111`.
Earlier owner runs at `20260910T005638Z-p72686` and
`20260910T005932Z-p82777` exposed old immediate-application assertions and a
graph refresh on unrelated removal; interactions and stable scope retention
were repaired. Formatting initially exposed semantic-label, list-key, and scalar
validation lint issues; `make format` passed at `20260910T010241Z-p9887`.
Retained owner-valid empty text list members remain explicit; newly added blank
members require correction/removal.

QA-03 exited into QA-04 with no blocked dependency.

## QA-04 validation record

- Real query-authoring browser scenario PASS, 11/11 units:
  `.cartulary/test-results/20260910T012212Z-p3967`. UI-authored requests return
  exact protocol-list/counter/range/null results, both diagnostic field tokens,
  inclusive source ranges, row overlap with zero-duration flows, temporal
  start-time selection, and explicit graph table IDs.
- Service-backed graph/contributor, table/cursor lifecycle, and temporal saved
  graph slices PASS, 3/3: `20260910T010703Z-p22677`.
- Existing accessibility row PASS, 11/11: `20260910T010859Z-p79823`.
- Scope and protected-state/deletion browser rows PASS, 13/13:
  `20260910T012054Z-p58293`. Separate graph application revision preserves
  stale-on-deletion until explicit recomputation.
- Whole module slice at `20260910T011225Z-p55863` failed 8/35 units, exposing
  the stale-on-deletion boundary, old immediate-scope assertions, old error-detail
  assertion, query-editor toggle, and expected golden changes. Narrow repaired
  error-detail/workspace rows PASS 3/3 at `20260910T011847Z-p13136`.
- Browser development failures at `20260910T010733Z-p46919`,
  `20260910T010916Z-p13868`, `20260910T011224Z-p55627`,
  `20260910T011531Z-p44797`, and `20260910T011951Z-p21966` identified
  source-row/header and exact field-key fixture expectations, editor toggle,
  value-error focus, and the Cisco profile's invalid-empty-port policy.
  The successful final scenario above uses the existing nullable interface
  mapping and preserves all result assertions.
- Latest model/scope unit repair PASS 3/3 at
  `20260910T012232Z-p33454`; a stricter generated graph decoder also caught
  a test-only 64-digit table ID, corrected to current-major 32 digits.
- Full ordinary visual run `20260910T011532Z-p45086` accounted for all 210
  capture intents/goldens: 210 active, zero orphan/missing/ambiguous, all 26
  registered fixtures resolved. Its only failing group was Network Analysis.
  Earlier scoped visual runs informed query-band compaction before promotion.
- Visual refresh trigger: intentional owner-aligned query controls, explicit
  applied status/scope, and independent saved-graph declarations. No viewport,
  zoom, mask, scroll-anchor, crop, tolerance, or fixture geometry change.
  Owner row: `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`.
  The Make-owned refresh and image review are complete; ordinary
  post-promotion evidence follows.

### Visual promotion review

`make browser-e2e-visual-update` passed 12/12 units at
`.cartulary/test-results/20260910T012811Z-p85990`. All eight promoted images
were opened and reviewed: compact controls, readable status/scope, preserved
inspector/drawer/dialog geometry, and independent saved declarations. No
unexpected clipping, theme, focus, or fixture changes were found.

Changed golden stems below all use `network-flow-analysis-` and
`-linux.png` in `apps/web/e2e/workbook.visual.spec.ts-snapshots/`:

- `accepted-inspector`
- `compact-saved-graphs`
- `delete-dialog`
- `filtered-empty-grid`
- `graph-contributors`
- `narrow-query-controls`
- `rejected-diagnostics`
- `saved-graph-result`

These are the same stable capture fixture IDs (without the platform suffix);
the mapping-dialog golden remained unchanged. The generated golden manifest
was promoted by Make. The first update attempt at `20260910T012412Z-p37413`
failed closed because frontend inputs changed during its build, promoting
nothing; the successful retry used frozen inputs.

Final ownership review keeps scalar parsing, reconstruction, filter/window and
graph request compilation, graph validation, and semantic comparison in the
pure query model/value modules. The workspace hook owns only volatile draft,
attempt, successful-snapshot, navigation, and authorization state. The frontend
source registry spells this source owner `web.network_flow`; verification
routing uses `web.networkflow`. Both new hook files are registered there.

The new late-failure test initially lacked a rejection method in its deferred
helper (`20260910T012648Z-p74497` and typecheck
`20260910T012649Z-p75986`); the helper is repaired. The complete frontend owner slice passed 59/59 at
`20260910T013521Z-p30867`; typecheck, import-boundary, and Biome checks passed
at `20260910T013521Z-p31097`, `20260910T013521Z-p31121`, and
`20260910T013521Z-p31159`.

The first ordinary post-promotion visual run passed 12/12 at
`20260910T013520Z-p29562`. A final bounded readability repair displays compact
Edit/Remove text with complete accessible predicate names, and prevents long
values from widening the advanced editor. The real browser scenario also checks
256-character text and explicit removal waiting for Apply. Frontend owner
verification after that repair passed 59/59 at `20260910T014258Z-p77606`.

The final real query-authoring browser slice passed 11/11 at
`20260910T014258Z-p77595`, including the long-value geometry and explicit
predicate removal checks. Functional assertions inspect successful service
responses and returned rows/contributors, not only outbound JSON.

The second fresh ordinary visual run passed 12/12 at
`20260910T014259Z-p77800`. Both required post-promotion ordinary runs accepted
the promoted manifest, with no tolerance or assertion weakening. QA-04 is DONE.

## QA-05 final verification

`env -u RESULTS_DIR make agent-finalize` ran before broader final checks.
No successful full warm check exists for the exact final source, so retained-run
maintenance is deliberately skipped with `RESULTS_DIR` unset. Owner slices and
ordinary visual evidence above do not qualify as a full warm retained run.

Finalization passed at `20260910T014708Z-p52771`; generated outputs were
unchanged (zero rewritten files). Its JSON-shape, catalog/tier, and generation
transaction checks passed. Retained canonical evidence, scheduler/timing, and
performance maintenance were skipped with reason `results-dir-not-provided`.

| Final command | Result/run under `.cartulary/test-results/` |
| --- | --- |
| `make test-slice OWNER=package.protocol_ts` | PASS 7/7, `20260910T014732Z-p56333` |
| `make frontend-typecheck` | PASS 2/2, `20260910T014732Z-p56519` |
| `make frontend-import-boundary-check` | PASS 2/2, `20260910T014732Z-p56538` |
| `make backend-module-boundary-check` | PASS 3/3, `20260910T014732Z-p56548` |
| `make lint-biome` | PASS 2/2, `20260910T014732Z-p56568` |
| `make json-shape-check` | PASS 3/3, `20260910T014732Z-p56269` |
| `make generated-artifact-policy-check` | PASS 3/3, `20260910T014732Z-p56260` |
| `make openapi-compatibility-check` | PASS 4/4, `20260910T014732Z-p56300` |
| `make lint-scripts` | PASS 2/2, `20260910T014732Z-p56597` |
| `make go-gosec-targeted` | PASS 4/4, `20260910T014732Z-p56781` |
| `make protocol-ts-browser-artifact-reachability test-catalog-check` | PASS, public checks exited zero (quiet success) |
| `make test-slice OWNER=module.networkflow` | PASS 35/35, `20260910T014732Z-p56309` |
| `make generate-drift` | PASS 4/4, `20260910T015033Z-p37406` |
| `make lint-markdown` | PASS, `20260910T014840Z-p21149` |
| `git diff --check`; branch/HEAD/status and permitted-path audit | PASS; same branch/HEAD, 52 intended changed paths, no unrelated changes |

The final module slice includes real service query/contributor and temporal
results, browser query authoring, import recovery, rename/deletion continuity,
saved-graph and indicator-link integration, protected-state cleanup,
accessibility, measurement, and ordinary visual verification. The frontend owner
slice passed 59/59 on the final authored source before finalization; finalization
changed no source or generated files. Protocol reachability retained its tool
summary under `20260910T014738Z-p67942`, and the catalog check under
`20260910T014756Z-p9874`.

### Completed behavior and ownership

- Protocol/endpoint lists and counter equality/upper/two-sided ranges remain
  explicit advanced predicates. Basic controls own only complete scalar or
  lower-only meanings. Repeated same-field predicates retain identity and order.
- Timestamp controls compile only adopted half-open ranges. Numeric bounds are
  inclusive; counters remain exact decimal strings through uint64 maximum.
  Typed list controls preserve every member, with canonical duplicate and
  invalid-member errors instead of coercion, truncation, or deduplication.
- Presentation eligibility matches the adopted grammar. Unsupported timestamp
  equality and TCP-flags/application-label filtering/sorting are removed. Typed
  schema variants and generated metadata drive admission and offered controls.
- The pure model owns reconstruction, validation, canonical comparison, row
  overlap compilation, and graph request construction. The volatile workspace
  hook owns raw drafts, immutable attempts, and successful applied snapshots.
  Controllers execute snapshots with existing paging/cursor mechanics.
- Graph scope, stable selected IDs, time, aggregation, and bucket width Apply
  together. Rows remain on their active-table route. Draft edits do not fetch or
  invalidate selections; submitted changes fence obsolete responses and actions.
- Rejections preserve raw local input and the last successful applied snapshot.
  Current response errors retain safe field/operator/index context. Clear,
  navigation, and authority replacement follow the explicit lifecycle above.
- Diagnostic token arrays/ranges remain separate and complete. Backend admission
  rejects unknown/duplicate/empty invalid inputs, and comparisons preserve text
  that resembles numbers. Server normalization, selectors, and digests remain
  authoritative.
- Native compact controls provide complete predicate summaries, associated local
  errors, keyboard focus, announcements, and distinct draft/applied/pending/failed
  text. Long values wrap without widening the advanced editor.

### Scope, compatibility, and rollback

No adopted owner decision remains unresolved. There are no new API capabilities,
routes, dependencies, persistence, or data migrations. Existing schema IDs remain;
closed variants repair their owner-governed grammar. Clients that relied on
unsupported presentation combinations or invalid diagnostic admission must
correct their requests. Unsafe browser integers receive explicit errors; uint64
counters remain exact strings. No client-generated graph digest is introduced.

The only paging change rejects a superseded failure using the existing generation
fence, matching existing success handling. No pagination/retry-controller redesign
was undertaken. Table lifecycle, import recovery, saved declarations, indicator
linking, grid semantics, density, and responsive behavior retain passing evidence.
Derived Imports/import-target digests changed solely because their existing
analytical binding consumes the Network Flow schema bundle.

Rollback is a coordinated source, authored projection, generated adapter,
routing, and visual-evidence revert. It requires no data migration. Generated
files and goldens were updated exclusively through Make-owned generation and
promotion; no lockfiles were edited. The digest, completed DU tracker, and other
historical handoffs are unchanged. No commits, resets, pushes, or deployments
were performed.

Repository-wide `make check`, CI/release publication, broad vulnerability audits,
and unrelated browser families were not requested or run: owner-routed suites
cover this seam, and the full ordinary visual runs cover the promoted manifest.
Retained-run maintenance remains explicitly skipped, as described above. All
implementation failures found during this seam were repaired and the affected
checks passed; no failed or blocked dependency is marked complete.

QA-05 is DONE. After this final tracker update, repeat Markdown lint,
`git diff --check`, and branch/HEAD/status/scope inspection. Next action is review
of the completed uncommitted patch; no implementation work remains open.

### Changed paths

```text
apps/web/e2e/network-flow.spec.ts
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-accepted-inspector-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-compact-saved-graphs-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-delete-dialog-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-filtered-empty-grid-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-graph-contributors-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-narrow-query-controls-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-rejected-diagnostics-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-saved-graph-result-linux.png
apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx
apps/web/src/networkFlow/NetworkAnalysisWorkspace.tsx
apps/web/src/networkFlow/NetworkFlowControls.tsx
apps/web/src/networkFlow/NetworkFlowQueryControls.test.tsx
apps/web/src/networkFlow/NetworkFlowQueryControls.tsx
apps/web/src/networkFlow/NetworkFlowTableController.ts
apps/web/src/networkFlow/networkFlowClient.ts
apps/web/src/networkFlow/networkFlowErrors.test.ts
apps/web/src/networkFlow/networkFlowErrors.ts
apps/web/src/networkFlow/networkFlowQueryAuthoring.test.ts
apps/web/src/networkFlow/networkFlowQueryModel.ts
apps/web/src/networkFlow/networkFlowQueryValues.ts
apps/web/src/networkFlow/tableConsumerContinuity.test.tsx
apps/web/src/networkFlow/useNetworkFlowGraphController.ts
apps/web/src/networkFlow/useNetworkFlowPagedQuery.test.tsx
apps/web/src/networkFlow/useNetworkFlowPagedQuery.ts
apps/web/src/networkFlow/useNetworkFlowQueryAuthoring.test.ts
apps/web/src/networkFlow/useNetworkFlowQueryAuthoring.ts
apps/web/src/networkFlow/useNetworkFlowRejectedRowsController.ts
apps/web/src/networkFlow/useNetworkFlowRowsController.ts
apps/web/src/services/networkFlowContractAdapter.ts
contracts/network-flow/presentation.v2.json
contracts/network-flow/schemas.v3.json
docs/handoffs/network-analysis-query-authoring-refactor-handoff.md
internal/gen/contractimports/artifacts_gen.go
internal/gen/contractnetworkflow/artifacts_gen.go
internal/gen/importtargetregistry/registry_gen.go
internal/modules/networkflow/network_flow_unit_test.go
internal/modules/networkflow/query.go
internal/modules/networkflow/query_authoring_test.go
internal/modules/networkflow/query_filter.go
packages/protocol-ts/src/entrypoints/network-flow.ts
packages/protocol-ts/src/generated/import-target-registry.ts
packages/protocol-ts/src/generated/network-flow-presentation.ts
packages/protocol-ts/src/generated/network-flow-types.ts
packages/protocol-ts/src/generated/network-flow-validators.ts
tools/browser_e2e_batch_manifest.json
tools/execution_topology_render_index.json
tools/frontend_source_ownership.json
tools/frontend_visual_golden_manifest.json
tools/protocol-ts/generate-protocol-types.mjs
tools/test_families/module.networkflow.json
tools/test_families/web.networkflow.json
```
