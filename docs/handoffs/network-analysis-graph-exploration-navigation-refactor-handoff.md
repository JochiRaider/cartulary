# Network Analysis graph exploration navigation refactor

## Baseline and authority

- Branch `main`; HEAD `767035c6e92f30bbfd391f825b2b115f994fa47b`.
- Planning and implementation entry working trees were clean. Root `AGENTS.md`
  is the only applicable agent instruction file. No commit, reset, push or deployment.
- The localized digest read order was followed and mappings revalidated against
  HEAD. Research and completed QA/PR/TL/SG/link handoffs are evidence, not new authority.
- Adopted owners: Network Flow 6.0.0 §§7, 14.6, 15, 19, 26–28;
  Core 03 REQ-03-011A/303; Core 04 REQ-04-021–023; Extensions
  EXT-REQ-101/104/201 and cleanup Table 18-C; design §§8.5, 13.1, 14.
- Verification owners: `web.networkflow`, `module.networkflow`, `package.ui`,
  `package.protocol_ts`, and frontend architecture/harness collaborators as touched.
- Permitted paths: exploration frontend and narrow pager/link/workspace adapters;
  relevant semantic selectors, tests/browser fixtures, source ownership and authored
  routing, Make-generated projections, reviewed affected goldens, design clarification,
  and this tracker. Backend production, public contracts, dependencies, routes,
  persistence, graph algorithms, imports and saved-graph lifecycle are excluded.
- Both requested task guides passed. Existing temporal and table-continuity slice
  passed 3/3 at `.cartulary/test-results/20260910T043252Z-p38609` during planning.

## Ordered workstreams

| Workstream | Status | Exit |
| --- | --- | --- |
| GN-01 Baseline, owner audit, policy, characterization | DONE | Characterized gaps and bounded decisions |
| GN-02 Result-bound model and presentation indexes | DONE | Pure transitions and bounded indexes |
| GN-03 Selection, contributor/link integration, continuity | DONE | Context and request fences |
| GN-04 Focus, accessibility, browser and rendering evidence | DONE | Semantic interactions and real results |
| GN-05 Final verification and completed handoff | DONE | Finalizer, evidence and scope audit |

Only the current row is IN_PROGRESS. A blocked dependency prevents advancement.

## UI decisions adopted for this seam

These are newly selected implementation UI policies, not claims that the prior
NLSpec required these exact transitions. They refine existing semantic identity,
bounded rendering, focus and supersession requirements without changing graph meaning.

| Event | Result |
| --- | --- |
| Change bucket | Reset vertex/edge pages and clear selection/contributors together. |
| Change object page | Retain valid selection and contributors; show off-page context and Reveal. |
| Close/Escape | Clear selection/contributors and return focus by semantic identity or visible fallback. |
| Explore/Saved switch | Retain one authorized accepted exploration result and coherent navigation in workspace memory. |
| Hide exploration | Cancel unfinished reads; retain committed data; resume missing initial reads on return, not abandoned page destinations. |
| Query/result replacement or dependent invalidation | Withdraw the entire affected navigation context; never expose stale bounds. |
| Table rename | Refresh labels without changing semantic result identity or navigation. |

Vertex contributors and counts cover the query range; temporal edges use their
exact bucket. Server bucket totals remain authoritative. Existing default-edge
indicator-link eligibility remains unchanged; no temporal-edge link contract is added.

## Characterization ledger

| Baseline transition | Observed implementation and gap |
| --- | --- |
| Select edge in A, move to empty B | B objects disappear, but controller selection and contributor request/data survive; drawer remains and endpoint labels use B's incomplete index. |
| Select object, page it away, close drawer | Selection is retained without off-page explanation; selected button ref becomes null, leaving no focus return target. |
| Later bucket/page → Saved → Explore | GraphPanel remount resets local bucket/pages while controller retains selection and contributors. |
| Result replacement/shrink | Passive page/bucket reset permits a render where stale bucket resolves to null and filters admit all buckets. |
| Superseded contributor read | Existing selector fences work when called, but bucket changes never call them. |
| Metadata-only refresh | Stable graph remains; table labels refresh. Preserve and extend this regression boundary. |

## Execution evidence

GN-01 implementation entry confirms the planning baseline unchanged. Bounded
design clarification and failing workspace characterization precede production edits.
The strengthened temporal characterization failed as expected at
`.cartulary/test-results/20260910T044003Z-p41136` (1/2 execution units), because
the empty bucket retains the contributor drawer. Existing source inspection
also establishes the off-page null ref, remount reset and transient null-bucket
fallback; GN-02/04 add deterministic model and mounted-boundary regressions.
No owner contradiction has been identified. GN-01 exits into GN-02.

GN-02: Added pure result-bound transitions, complete selector indexes, deterministic
ordering and bounded presentation. Model slice passed 2/2 execution units at
`20260910T044621Z-p44081`. Initial routing rejected unsorted test titles; sorted
the authored titles. Initial typecheck at `20260910T044518Z-p43384` exposed
fixture union/tuple typing; repaired the fixtures to current-major variants.
Production presentation extraction joins controller integration in GN-03 so no
intermediate duplicate state owner is introduced. Next: connect the model,
pause reads, and preserve link/contributor fences.

GN-03: The graph controller composes the pure navigation model and the extracted
exploration panel. Selection transitions synchronously fence contributors and
notify the existing indicator-link owner. Pager activity pauses reads without
resetting committed pages/history. Saved current-graph handoff retains only the
authorized accepted result. Focused model/controller/temporal/table/pager tests
passed 6/6 at `20260910T045553Z-p58411`; typecheck passed at
`20260910T045554Z-p58682`. The initial extraction exposed a shared row-action
style dependency, restored without changing row presentation. Formatting passed
at `20260910T045524Z-p54097`. Existing contributor race fixtures now select
objects actually present in the accepted graph. No cursor/replay algorithm,
public interface, persistence or saved lifecycle changed. Next: mounted focus,
real graph/contributor data, accessibility and rendering validation.

GN-04 review finding: the existing absolute contributor drawer covers the graph's
selection column. A Reveal can focus a mounted control behind that drawer.
The bounded presentation will reserve adjacent drawer space and keep the object
list scrollable so semantic focus is visible. This implements the selected focus
policy within existing tokens; saved presentation and the visual theme remain
separate. The first ordinary visual run's graph-contributor image was inspected;
golden promotion waits for reconciled evidence and this usability correction.

GN-04 exits with explicit semantic focus intents bound to authority, result and
interaction generation, consumed once. Close/Escape use a mounted semantic target
or navigation/region fallback; withdrawal restores the region only if it still
owns focus. Newer focus and Saved departure cancel obsolete intentions. The
adjacent drawer and minimum scrollable graph area prevent covered controls and
zero-height object lists when temporal/page controls wrap.

- Mounted focus/continuity slice: 4/4 at `20260910T051641Z-p94610`.
- Real temporal and accessibility: 13/13 at `20260910T050706Z-p15363`.
- Saved measurement, protected-state browser and existing keyboard regression:
  18/18 at `20260910T051300Z-p98619`.
- Final exploration measurement: 12/12 at `20260910T052014Z-p32703`.
  Actual server result has 501 vertices, 2,002 temporal edges, and 1,001 edges
  per populated bucket. Mutation observations verify 500/1,000 mounted maxima,
  no intermediate wrong-bucket edges, no local graph requests, actual contributor
  row references, and unobscured Reveal focus at 1280×720.
- Service-backed bounded graph/contributor pipeline, pagination recovery and
  temporal saved lifecycle: 3/3 at `20260910T051301Z-p99004`.
- UI selector slice 10/10 at `20260910T051836Z-p30212`; protocol slice 7/7 at
  `20260910T051837Z-p30436`.

GN-04 failures were resolved without reducing assertions or tolerances. The
initial browser fixture assumed fractional UTC serialization (fixed to compare
instants while checking exact selectors). A large mounted test timed out in the
full parallel owner run (`20260910T050635Z-p3945`, 64/65); scoped semantic queries
removed unnecessary whole-document traversal. New source-withdrawal fixtures
needed matching accepted source identities and awaited lifecycle observation.
The first unobscured-Reveal checks (`20260910T051541Z-p55145` and
`20260910T051744Z-p97357`) identified the zero-height list at constrained height;
the final measurement verifies the presentation correction. The original
focus-test duplicate-text assertion was scoped to the drawer.

Ordinary visual reconciliation at `20260910T051057Z-p56704` accounts for all
210 active goldens/capture intents and 26 registered fixtures: no orphan,
missing golden, ambiguous mapping or unresolved fixture. Its only reconciliation
error is the expected unsuccessful visual attempt; the graph-contributor
comparison differs. Next: run finalizer with no retained full-run input, inspect
the corrected image, promote only intentional goldens, then final owner and
harness checks and two ordinary visual passes.

## Final state ownership and scope

- `networkFlowExplorationNavigation.ts` owns pure context/result acceptance,
  complete selector indexes, server-ordered bucket membership, bounded pages,
  selection, announcements and semantic focus intent. Context replacement keeps
  interaction generations monotonic, including authority A → B → A.
- `useNetworkFlowGraphController.ts` composes that model and retains graph read
  execution, source observation and contributor dispatch fences. Accepted server
  source identities determine affected-result invalidation.
- `NetworkFlowExplorationPanel.tsx` realizes the bounded presentation and semantic
  focus targets. It reserves drawer space and a usable scrollable object area;
  it never retains a selected DOM node as navigation state.
- `NetworkAnalysisWorkspace.tsx` connects explicit transitions to the existing
  indicator-link selection context and the Explore/Saved activity boundary.
- `useNetworkFlowPagedQuery.ts` adds optional activity/pause behavior only;
  committed pages and history survive suspension. Cursor and recovery algorithms
  retain their existing owner.
- Supporting tests: `networkFlowExplorationNavigation.test.ts`,
  `explorationContinuity.test.tsx`, `explorationFocus.test.tsx`,
  `explorationTestFixtures.ts`, workspace, table-continuity and page-recovery
  tests; real application `e2e/network-flow-navigation.spec.ts`. Existing browser
  and visual selector queries now match semantic control names.
- Authored ownership/routing: `tools/frontend_source_ownership.json` and the
  `web.networkflow`/`module.networkflow` test-family manifests. `make generate`
  updated only the downstream browser batch manifest and topology render index.
- Documentation: the bounded design §13.1 clarification and this tracker only.
  The digest and completed trackers were not changed.

All source paths above are relative to `apps/web/src/networkFlow/` unless stated
otherwise. The implementation introduces no migration, route, dependency,
persistence, contract version or graph semantic change. Saved declarations,
results and their navigation remain separate. Rollback is a coordinated revert
of these source, test, routing/generated, documentation and reviewed golden
changes. There is no data migration or compatibility bridge to roll back.
No unresolved owner decision remains. No unrelated changes were present at entry or introduced.

## GN-05 verification record

`env -u RESULTS_DIR make agent-finalize` passed 1/1 at
`20260910T052222Z-p2200`. It changed no files. Retained full-run, performance,
canonical-evidence and scheduler maintenance were skipped because no qualifying
successful full warm run was supplied; `RESULTS_DIR` was unset.

- `make test-slice OWNER=web.networkflow`: 66/66 at
  `20260910T052306Z-p6428`, including query authoring, pagination recovery,
  indicator linking, saved graphs, semantic grids and table continuity.
- `make frontend-typecheck lint-biome frontend-import-boundary-check`: each
  passed 2/2 at `20260910T052409Z-p74888`, `-p74898`, and `-p74892` respectively.
  The preceding typecheck at `20260910T052306Z-p6544` found test-observer
  branded-ID/window typing; corrected without changing runtime assertions.
- Final generation-fence model/controller/focus slice: 4/4 at
  `20260910T052552Z-p92601`. Context replacement is now an explicit pure
  transition; a saved old focus intent cannot match a later return to the same
  authority/result/selector.
- Real browser compatibility slice: 21/21 at `20260910T052345Z-p43113`, covering
  query-authoring integrity, indicator recovery and keyboard use, pagination
  recovery, saved deferred/exact-result navigation, and exploration temporal
  and accessibility scenarios.
- `make test-catalog-check json-shape-check generated-artifact-policy-check`:
  passed; shape 3/3 at `20260910T052445Z-p89574`, policy 3/3 at `-p89570`.
- `make protocol-ts-browser-artifact-reachability`: passed.

The GN-01 characterization command selected the temporal workspace row with
`make test-slice OWNER=web.networkflow
ROWS=web.networkflow.regression.temporal_graph_controls_bucket_navigation`.
GN-02 selected `web.networkflow.regression.exploration_navigation`; GN-03 added
`exploration_continuity`, `table_consumer_continuity`, and the existing pager
suite to that routed slice. GN-04 browser rows are
`module.networkflow.browser_stateful.exploration_navigation`,
`module.networkflow.accessibility.exploration_navigation`, and
`module.networkflow.measurement.exploration_graph_dom_ceiling`, invoked with
`make test-slice OWNER=module.networkflow ROWS=...`. Service-backed rows were
`module.networkflow.integration.bounded_graph_contributor_pipeline`,
`module.networkflow.integration.pagination_recovery`, and
`module.networkflow.integration.time_bucket_saved_graph_lifecycle`, invoked
with `make service-backed-test-slice OWNER=module.networkflow ROWS=...`.

Visual promotion and post-promotion verification are complete; evidence follows.

Visual update passed 12/12 at `20260910T052306Z-p6567`. The only promoted image is
`apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-graph-contributors-linux.png`,
with its Make-owned manifest hash. The accepted trigger is explicit graph identity,
query-wide count/contributor labeling and adjacent drawer space that leaves
semantic selection controls visible. The authored owner row is
`module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`;
fixture `visual.fixture.claimed_network_analysis_workspace_states`, scenario
`scenario_252dd2b0ff55`. Reviewed both the ordinary actual image at
`20260910T052137Z-p69497` and the promoted image. No viewport, zoom, mask,
scroll normalization, screenshot scope, density or renderer pin changed.
Reconciliation passed with all 210 active goldens and 26 registered fixtures,
no missing/ambiguous/orphan mappings, and no errors. Saved-graph goldens did not
change.

Final type/Biome/import-boundary rerun passed 2/2 each at
`20260910T052941Z-p72247`, `-p72257`, and `-p72251`. The intermediate
`20260910T052828Z-p98628` typecheck identified a nullable focus fixture helper;
the helper now rejects both null and undefined. The authority replacement
fixture now includes enough actual bucket endpoints to exercise a noninitial
vertex page; its rerun passed at `20260910T052918Z-p70211`.
`make generate-drift` passed 4/4 at `20260910T052828Z-p98449`.
Post-promotion catalog/shape/policy checks passed; shape/policy roots are
`20260910T052957Z-p72912` and `-p72910`. Initial Markdown lint passed at
`20260910T052920Z-p70525`; it will run again after the final tracker edit.


Both fresh ordinary post-promotion `make browser-e2e-visual` runs passed 12/12:
`20260910T052828Z-p98712` and `20260910T052828Z-p98713`. Each independently
reconciled all 210 captures/goldens and 26 registered fixtures without errors.
The final `make test-slice OWNER=web.networkflow` rerun passed 66/66 at
`20260910T053057Z-p78655` after the monotonic-generation correction.

GN-05 is DONE. All planned behavior and scoped compatibility checks are
complete. Full repository CI/release, unrelated backend security suites and
backend capacity certification were not selected for this frontend seam;
service-backed semantic checks and frontend rendering measurement were executed.
Retained full warm-run maintenance remains explicitly skipped as recorded above.
No migration, commit, reset, push, deployment, lockfile or dependency edit was
performed. The only remaining delivery checks are the required post-record
Markdown lint, `git diff --check`, and final branch/HEAD/scope/status inspection;
their terminal results accompany delivery rather than prompting another tracker
edit.
