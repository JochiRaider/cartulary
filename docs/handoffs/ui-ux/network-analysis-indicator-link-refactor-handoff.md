# Network Analysis indicator-link workflow handoff

## Baseline and authority

Implementation begins on clean `main` at
`e6d66bdb319f5d7f1986e7828c94de2b2ac137a7` (20 commits ahead of origin),
revalidated on 2026-09-09. The recommendation baseline was `1e85da38`;
the intervening saved-graph reassessment is retained. No reset, commit, push,
deployment, dependency addition, or stored-data rewrite is authorized.

AGENTS.md and the digest's localized read order were inspected during planning
and the root instructions and worktree were revalidated before implementation.
Core and adopted subsystem owners govern behavior. Domain vocabulary, scoped
design direction, NLSpec research, and completed chrome/import/saved-graph/visual
handoffs retain their distinct authority and evidence boundaries. The digest
remains unchanged. No runtime, test, generator, routing, conformance, or release
input may depend on this document or other Markdown.

## Ordered tracker

| Workstream | Status | Exit |
| --- | --- | --- |
| IL-01 Baseline, owner map, characterization | DONE | Mapped gaps, reproductions and compatibility assessment |
| IL-02 Typed transport and captured operation | DONE | Admission, fencing, receipt, atomicity and replay tests |
| IL-03 Confirmation, targets and recovery | DONE | Deterministic local recovery and stale-context behavior |
| IL-04 Browser, accessibility and regressions | DONE | Real workflows and preserved analytical baselines |
| IL-05 Final validation and handoff | DONE | Required checks and final scope audit |

Only the current row may be IN_PROGRESS. Complete its exit, append paths,
commands, results, decisions and next action, and then start its successor.
An unresolved adopted-owner conflict is `BLOCKED: owner contradiction`.

## Owner-to-change map and characterization

Exact owner navigation: `docs/network-flow-activity-nlspec.md` §§5–7, 15–17,
20–21 and 25–28; `docs/spec/00_document_set_status_and_precedence.md`;
`docs/spec/01_architecture_storage_and_view_contracts.md` (view queries,
continuations and transaction boundary); `docs/spec/02_domain_model_schema_and_history.md`
(indicator identity, canonical creation/deduplication and REQ-02-074A–C);
`docs/spec/03_workbook_interaction_collaboration_and_workflows.md` (workbook
interaction/invalidation); `docs/spec/04_security_deployment_and_conformance.md`
(authorization, CSRF, hidden resources, lifecycle and audit);
`docs/extension-subsystem-nlspec.md` EXT-REQ-201/212 (generation, availability
intersection and typed cleanup); and `docs/design.md` within its design scope.
`docs/domain.md` and `docs/research/nlspec-spec.md` retain vocabulary/navigation
and authoring boundaries. The digest remains advisory review evidence.

| Gap | Owner | Reproduction and bounded correction |
| --- | --- | --- |
| Result identifier | NF-REQ-165, Table 17-E | Backend constant, schemas/routes/index and browser decoder use the dotted identifier instead of exact `cartulary.network_flow_indicator_link_result.v1`. Correct projections and emission together. |
| Mutation digest | NF-REQ-024–027, Table 5-B | `indicatorLinkRequestHash` hashes a JSON wrapper without the prescribed domain/path transcript. Correct only indicator linking; preserve exact bytes for replay. |
| Selector boundary | NF-REQ-142–143a, Tables 15-A1/B, §27 | Indicator selector projection admits bucket-edge IDs. Restrict linking to authorized endpoint/default-flow IDs while retaining semantic-query v2 and temporal contributor behavior. |
| Admission/errors | NF-REQ-020–023, NF-REQ-175–176, Tables 21-A1/B/C/D; Core 04 auth/CSRF | Decoder emits unregistered reasons, misses closed row-ref validation and required contextual details, and mixes schema/freshness/semantic order. Repair link-local admission and exact errors. |
| Fresh write checks | NF-REQ-143/144b; Core incident lifecycle and indicator transaction owner | Resolution precedes transaction acquisition; participant checks target only. Revalidate current write authority and active sources in the existing unit of work. |
| Operation lifetime | NF-REQ-052–054a; Extensions EXT-REQ-201/212 | React pending state admits same-turn repeats; transport generates a new ID for each call; no immutable recovery exists. Capture one attempt under workbook lifetime with independent presentation. |
| Permission | NF Table 5-A; Core 04 route authorization | UI sets `canLink = canImport`; explicit link role admission is editor/admin. Separate the link gate and test every role. |
| Targets and receipts | NF Tables 15-C–F; Core 02 REQ-02-074A–C; Core 01 view query/pagination | Dialog guesses type and accepts only a raw ID; structural decoder does not validate binding applicability or truncation relationships. Use Core registry facts, bounded authorized discovery and semantic validation. |
| Recovery and status | NF-REQ-053–054a; Extensions §18; design §§8.5, 10, 12–14 | Failure is outside the modal, closing is disabled while pending, and link status never enters owner precedence. Keep local recovery and focus behavior separate from transport effects. |

The existing focused frontend model/transport slice passed 6/6 execution units:
`make test-slice OWNER=web.networkflow ROWS=<five recorded model/transport rows>`
at `.cartulary/test-results/20260909T195910Z-p89317`.
The service-backed graph/contributor/link slice passed 3/3:
`make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline`
at `.cartulary/test-results/20260909T200007Z-p90454`.
These characterize the implementation; passing tests do not bless owner drift.
`make help`, `make help-all`, and guides for web.networkflow,
module.networkflow, module.indicators, web.workbook, package.protocol_ts and
app.server succeeded. Initial probes of nonexistent owner IDs package.protocol
and application.server failed usage validation; corrected IDs succeeded.

## Decisions and compatibility

ADOPT semantic identity, local recovery, non-color state and keyboard access.
ADAPT compact dialog, asynchronous feedback and responsive containment through
existing tokens and controls. REJECT generic workflow machinery, invented
contracts, automatic re-keying and visual redesign.

The user selected workbook-lifetime volatile recovery: one unresolved attempt
and a separate editable draft. Dialog dismissal does not cancel server work.
Explicit recovery reuses exact bytes/ID under current authority. A 30-second
browser observation window is not a server deadline or proof of rollback.

The user selected a read-only admission check for incompatible retained link
receipts. Correcting the result identifier and digest affects public clients and
replay compatibility, although it introduces no route, API version, schema
migration or stored-data rewrite. Reject nonconforming retained receipts before
claim admission; preserve their bytes for separately authorized remediation.
No alias, decoder fallback, receipt normalization or deletion is permitted.

Rollback must restore a coherent server/browser/projection set whose admission
accepts retained receipts. After corrected writes exist, prefer a forward fix;
do not delete indicators/bindings, rewrite receipts/history, or assume an older
binary can read the corrected contract. Backup restoration is not authorized.

## Binary acceptance

| Criterion | Status |
| --- | --- |
| Public boundary and retained-receipt admission match adopted owners | PASS |
| Single synchronous admission, immutable bytes and exact replay | PASS |
| Dispatch and response fencing cover authority/source/presentation changes | PASS |
| Core identity, binding-only atomicity, reuse and audit semantics preserved | PASS |
| Compatible target discovery and field-local recovery are keyboard complete | PASS |
| Cleanup, uncertainty and status lifetime are deterministic | PASS |
| Browser/a11y/visual workflows and analytical regressions pass | PASS |
| Finalize, selected checks and final Markdown/diff/scope checks pass | PASS |

## Execution log

IL-01 opened after exact branch/commit/dirty-state revalidation. Inspected the
controller/model/dialog/client/adapter, authored contracts, indicator participant
and binding storage, extension admission, workbook lifetime, current routing and
visual guide. Next: complete link-local characterization, then IL-02.

IL-01 exit: `indicator_link_boundary_test.go` reproduces the incorrect result
identifier, missing domain transcript and accepted unknown source-ref members.
Its owner-routed unit slice fails as expected at
`.cartulary/test-results/20260909T201133Z-p18192`. The existing binding evidence
still passes. Remaining reproductions are the bounded source paths in the map.
No adopted-owner contradiction was found. Semantic-query v2 supersedes the old
query reference; bucket contributor support does not authorize bucket linking.
Core-unrepresentable flow IP values remain inspectable without link admission.

`make generate` initially rejected unsorted authored test selectors at
`.cartulary/test-results/20260909T201034Z-p12091`; sorting that input resolved it.
The rerun passed at `.cartulary/test-results/20260909T201113Z-p15215`.
The compatibility assessment above precedes all projection changes.
Next: IL-02 contract corrections, transaction checks and captured operation.

IL-02 implementation separates owner-backed boundary corrections from movement:

- Corrections: result identifier, domain-separated comparison digest, flow-edge
  selector restriction, UUID target/actor fields, absolute source-ref bound,
  closed deterministic link admission, contextual errors, fresh transaction
  authority/source checks, and read-only binding/receipt admission. Core exposes
  exact atomic-IP classification without changing canonicalization or dedupe.
  The transaction coordinator exposes its effective timeout for required public
  diagnostics; its commit and rollback protocol is unchanged.
- Structure: the workbook mounts `NetworkFlowIndicatorLinkController` separately
  from lazy workspace presentation. A thin hook subscribes. Captured operation
  types own immutable request bytes, one secure ID, dispatch authority, semantic
  source identities, observation and exact replay. Core's existing OpenAPI type
  registry generates browser facts through the HTTP entrypoint; the focused
  adapter handles exact IP compatibility, outside the dialog.
- Compatibility: well-formed UUIDs and source-ref counts already conform to the
  adopted boundary. Placeholder IDs and bucket-edge link requests now fail.
  Generated Imports registry digests change transitively because they bind the
  Network Flow schema bundle; import behavior is unchanged. No schema migration
  or alias was introduced. Old dotted result receipts block read-only admission.

Verification so far:

| Command / scope | Result and retained root |
| --- | --- |
| `make generate` after contract correction | PASS `.cartulary/test-results/20260909T201936Z-p25794` |
| Link unit / service slice, initial correction | Compile failure in private API-error propagation at `20260909T201954Z-p28800` and `20260909T201954Z-p28810`; wrapped the transport value in a private participant error. |
| Link unit rerun | New characterization passed; legacy fixture used invalid placeholder IDs at `20260909T202041Z-p38593`. Corrected fixture IDs and deferred field-policy expectation. |
| Graph/contributor/link service slice | PASS 3/3 `.cartulary/test-results/20260909T202042Z-p38847` |
| Same service slice with receipt admission and source-removal replay | PASS 3/3 `.cartulary/test-results/20260909T202411Z-p62254` |
| Link boundary unit including selector/order checks | PASS `.cartulary/test-results/20260909T204214Z-p96523` |
| New captured-operation frontend slice | PASS `.cartulary/test-results/20260909T203746Z-p89361` |
| `make frontend-typecheck` | Nullable availability tag found at `20260909T203946Z-p90459`; corrected representation. PASS `.cartulary/test-results/20260909T204040Z-p91243` |
| Core Indicators exported-surface and vocabulary rows | PASS `.cartulary/test-results/20260909T204214Z-p96495` |
| `make test-slice OWNER=module.crossownertransaction` | PASS 4/4 `.cartulary/test-results/20260909T204214Z-p96500` |
| `make frontend-import-boundary-check` | PASS 2/2 `.cartulary/test-results/20260909T204214Z-p96685` |

Authored test routing is regenerated through `make generate`; latest successful
routing generation is `.cartulary/test-results/20260909T203736Z-p86509`.
`make format` caught the thin subscription hook's dependency expression at
`20260909T204200Z-p92040`; changed it to an explicit effect callback.
Next: finish the combined IL-02 slice, then local dialog/discovery integration.

IL-02 exit: the combined operation/model/transport slice passes 7/7 execution
units at `.cartulary/test-results/20260909T204329Z-p16928`; formatting passes at
`.cartulary/test-results/20260909T204328Z-p16777`, and `git diff --check` passes.
Admission, queued dispatch, immutable replay, late receipt, cleanup, digest,
receipt validation, Core boundary and transaction protocol evidence passes.
The workspace is structurally wired to the workbook owner; its legacy dialog
presentation is the next bounded replacement. Next: IL-03 target discovery,
field-local feedback, recovery entry, focus behavior and status projection.


IL-03 exit: `IndicatorLinkDialog.tsx` now renders local fields, captured pending
and uncertain intent, explicit replay/abandonment, and persistent binding
receipts. The workbook mounts presentation and recovery independently of the
lazy analytical workspace. Native radios/select/ID entry preserve keyboard
access; semantic grid focus restoration uses the existing `GridHandle` port.
Target discovery uses one current Indicators query page (100 rows), exact local
atomic-IP compatibility, stable record IDs and explicit opaque continuation.
Discovery generations and read failures are separate from write settlement.
Core target changes invalidate discovery/confirmation and removal purges the
related volatile state. No browser storage or additional routes were added.

`networkFlowIndicatorLinkModel.ts` owns row and graph candidate construction;
graph candidates capture semantic selectors and source-table identity without
using labels or a contributor page as completeness evidence. The focused status
projection follows Table 7-B and retains receipts after the five-second success
token expires. Source-limit reads have bounded observation and local retry.

A follow-up boundary audit found inherited graph-query admission issues on the
link route. `indicator_link_graph_admission.go` closes the nested query shape,
provides filter context, and defers current effective query limits until after
replay. JSON member categories now precede canonical member ordering. The link
route also rejects undeclared URL parameters. These are owner-backed corrections
under NF-REQ-020–027 and Tables 21-B/C/D; other query routes and graph algorithms
are unchanged. Current contract-shaped clients remain compatible; malformed
nested members and undeclared parameters are now rejected.

| Command / scope | Result and retained root |
| --- | --- |
| Operation/discovery/status, command matrix and focus rows | PASS 4/4 `.cartulary/test-results/20260909T210524Z-p49640` |
| Full production row/graph/indicator workspace integration | PASS 2/2 `.cartulary/test-results/20260909T210524Z-p49647` |
| `make test-slice OWNER=web.networkflow` | PASS 51/51 `.cartulary/test-results/20260909T211548Z-p76842` |
| Production workspace plus backend link boundary | PASS 3/3 `.cartulary/test-results/20260909T211548Z-p76848` |
| Graph/contributor/link service-backed slice | PASS 3/3 `.cartulary/test-results/20260909T211548Z-p76874` |
| `make frontend-typecheck` | PASS `.cartulary/test-results/20260909T211646Z-p2169` |
| `make frontend-import-boundary-check` | PASS `.cartulary/test-results/20260909T211646Z-p2173` |
| Latest `make format` | PASS `.cartulary/test-results/20260909T211515Z-p72298` |

Resolved intermediate failures: unused moved helpers (`205956Z-p30724`), grid
focus-ref effect dependencies (`210200Z-p32075`), missing Core query metadata in
a new test fixture (`210409Z-p44015`), and an unsupported Testing Library option
in two assertions (`211548Z-p77025`). No tolerance or product contract was relaxed.
Authored routing generation passed at `20260909T210346Z-p41018`.
Next: IL-04 expand real linking, transaction, accessibility and visual evidence,
then audit existing Network Analysis browser baselines.


IL-04 expanded the real service matrix to source/destination rows, same-value
ranges, graph vertices/default-flow edges, IPv6, different source sets, original
reuse metadata, truncated graph references, incompatible/hidden/cross-incident
Core targets, and Core create-or-reuse. Transaction races hold the existing
incident lock, observe the request blocked on that lock, then publish closure,
role loss or source removal; fresh writes reject without an indicator or receipt.
Audit assertions distinguish new-key reuse from exact replay and reject candidate
leakage or missing digest-key pairing. No graph algorithm or Core identity changed.

New owner-routed browser and accessibility rows exercise native target selection,
field-local exact confirmation, a response lost after real server commit, replay
of unchanged bytes from another workbook surface, persistent receipts, Escape,
focus return, narrow/short layouts, 200% zoom, text spacing and long IPv6 values.
The first combined run passed 13/13 execution units at
`.cartulary/test-results/20260909T212739Z-p26059`. Its inline image attachments
were reviewed: containment and wrapping passed; excessive default paragraph
spacing prompted a token-based dialog spacing correction and another ordinary
run. No screenshot tolerance or fixture timing was changed.

Resolved expanded-test failures: the existing destination fixture was IPv6,
not the assumed IPv4 (`212739Z-p26069`), and Core's current transaction role
failure is `authorization_denied` (`212957Z-p83059`). The corrected service matrix
passed 3/3 at `213649Z-p7268`. Browser typechecking caught `Node.remove` on a
broad DOM type (`212739Z-p26168`); typed parent removal resolved it.

The final link-error audit added registry-based status/reason/retry validation
and required detail presence before treating a Network Flow rejection as definite.
An incomplete error remains uncertain. It also corrected the missing source-limit
`row_limit_exceeded` reason and the query-validation retry action. These are
owner-backed Table 21-A1/B corrections, with no additional route or schema version.
Tests now cover a queued observation deadline before bytes, verified owner timeout,
malformed rejection, failed exact replay and a late abandoned attempt after a new
write. A failed replay never erases the original uncertainty.

| Command / scope | Result and retained root |
| --- | --- |
| `make generate` for new browser routing | PASS `.cartulary/test-results/20260909T212658Z-p22841` |
| `make test-slice OWNER=web.networkflow` including added uncertainty cases | PASS 51/51 `.cartulary/test-results/20260909T214022Z-p31002` |
| Expanded graph/contributor/link service slice with audit privacy | PASS 3/3 `.cartulary/test-results/20260909T214022Z-p31017` |
| `make frontend-typecheck` after final error validation | PASS `.cartulary/test-results/20260909T214052Z-p81274` |
| `make frontend-import-boundary-check` | PASS `.cartulary/test-results/20260909T213649Z-p7344` |

Next: finish the complete Network Analysis browser regression set, review ordinary
visual reconciliation and any justified refresh, then close IL-04.


The complete selected Network Analysis functional/stateful/accessibility/
measurement slice passed 21/21 execution units at
`.cartulary/test-results/20260909T214052Z-p80975`. Selection covered all active
nonvisual Playwright rows for module.networkflow, including the new link rows,
imports, saved-graph recovery/lifecycle/deferred navigation, graph contributors,
protected-state cleanup, virtualization and the saved-graph DOM ceiling. Fresh
link receipt and IPv6-spacing attachments were reviewed from that run's
`browser-e2e-webserver-backed` and `browser-e2e-a11y` Playwright reports.
Core target/participant/rollback service rows passed 3/3 at
`.cartulary/test-results/20260909T214052Z-p81004`.

### Visual refresh review

Ordinary `make browser-e2e-visual` at
`.cartulary/test-results/20260909T214022Z-p31225` failed only the Network Analysis
screenshot comparisons; the other workbook visual group passed. The retained
`browser-e2e-visual/frontend-visual-reconciliation.json` accounts for all 210
capture intents and committed goldens, 26 registered fixtures, zero orphans,
zero missing goldens, zero ambiguous mappings and zero unresolved fixtures.
Both renderer attestations and the committed manifest passed. The sole
reconciliation error is the unsuccessful comparison attempt.

Accepted trigger: the NF Table 7-B status projection replaces stale import
success text with the current analytical status, including graph availability
and loaded-with-rejections. All eight actual images and the delete-dialog diff
were inspected; the comparison differences are status text at the footer.
Layout, typography, focus, framing and containment remain as before. No viewport,
zoom, masks, scroll normalization, screenshot scope, fixture, renderer or tolerance
changes are included. The new dialog has separately reviewed ordinary browser
attachments; it introduces no golden fixture or screenshot harness change.

Affected authored row:
`module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`,
scenario `scenario_252dd2b0ff55`. All refreshed paths are under
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`:

| Golden filename | Stable fixture |
| --- | --- |
| `network-flow-analysis-accepted-inspector-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-rejected-diagnostics-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-graph-contributors-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-mapping-dialog-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-saved-graph-result-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-delete-dialog-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-compact-saved-graphs-linux.png` | `visual.fixture.claimed_network_analysis_compact_workspace` |
| `network-flow-analysis-narrow-query-controls-linux.png` | `visual.fixture.claimed_network_analysis_narrow_workspace` |

The active nonregistry filtered-empty-grid golden remains retained. Refresh uses
only `make browser-e2e-visual-update`; its promotion and two subsequent fresh
ordinary passes remain required before IL-04 can close.


`make browser-e2e-visual-update` passed 12/12 execution units at
`.cartulary/test-results/20260909T214531Z-p58702`, promoting exactly the eight
listed goldens and `tools/frontend_visual_golden_manifest.json`. Seven promoted
images are pixel-identical to the reviewed ordinary actuals. The promoted delete
dialog was inspected separately: its difference from the original golden is
confined to the footer as well. No unrelated golden changed.

Additional regression evidence:

| Command / scope | Result and retained root |
| --- | --- |
| `make test-slice OWNER=web.workbook` | PASS 169/169 `.cartulary/test-results/20260909T214734Z-p10618` |
| `make test-slice OWNER=package.protocol_ts` | PASS 7/7 `.cartulary/test-results/20260909T214734Z-p10629` |
| Network Flow production frontend, link boundary, error details, claim configuration and transaction lifecycle rows | PASS 3/3 `.cartulary/test-results/20260909T214734Z-p10638` |
| Link service matrix with forced audit-insert failure | PASS 3/3 `.cartulary/test-results/20260909T214657Z-p92670` |
| Expanded destination row/range/vertex and temporal vertex/bucket-edge matrix | PASS 3/3 `.cartulary/test-results/20260909T215035Z-p35842` |
| `make format` | PASS `.cartulary/test-results/20260909T215005Z-p31432` |

The audit-failure test installs a constraint only in its disposable test database,
fails the real audit write, checks rollback of the Core indicator, binding,
idempotency receipt and audit occurrence, and removes the constraint. It adds
no production fault hook, route or migration. Temporal endpoint vertices remain
linkable; a bucket-edge ID remains unsupported even though its contributor route
continues to pass.


First fresh ordinary visual pass: `make browser-e2e-visual` passed 12/12 at
`.cartulary/test-results/20260909T215035Z-p35962`, including both the Network
Analysis group and all 37 wider workbook visual scenarios. No image changed.
The second fresh ordinary pass is the remaining IL-04 dependency.


Retained extension-state, saved-graph cutover/lifecycle and table-store lifecycle
service rows passed 5/5 at `.cartulary/test-results/20260909T215500Z-p18169`.
The final feedback audit corrected target incompatibility to local validation
rather than denied access, and maps target/confirmation schema fields to their
inputs. This bounded transport presentation correction does not change any
captured golden state. Its focused test runs again before IL-04 closes.


Second fresh ordinary visual pass: `make browser-e2e-visual` passed 12/12 at
`.cartulary/test-results/20260909T215441Z-p88516`. Both required passes use the
promoted manifest without update mode or tolerance changes. Their reconciliations
pass, and all 210 goldens remain accounted for.

The feedback fixture initially omitted the JSON content type and correctly
produced uncertainty (`215631Z-p44385`); supplying the real route's content type
repairs the test. Final receipt review also found that UUID spelling must compare
by UUID identity: uppercase hexadecimal in a known target ID now matches the
canonical returned UUID, while the captured body and IP confirmation remain
unchanged. Link admission and retained-receipt checks reject non-UUID URN,
braced and unhyphenated spellings that Go's general UUID parser also accepts.
This is a correction to the already-adopted UUID wire format, not a new identity
policy or compatibility alias. Focused boundary and service evidence is rerun.


IL-04 exit: final link unit admission passes at
`.cartulary/test-results/20260909T215853Z-p50126`; the full service matrix passes
3/3 at `215853Z-p50140`. The operation/transport slice passes 2/2 at
`220016Z-p68988`. Its last intermediate failure (`215853Z-p50117`) was a test
reference-equality assertion against the validator's intentionally copied receipt;
value equality fixes the assertion without changing product behavior. Complete
workflows, local recovery, cleanup, keyboard behavior, retained-state baselines,
Core atomicity and both required fresh ordinary visual passes are evidenced above.
Next: IL-05 agent finalization, broader applicable checks and final scope audit.

IL-05 starts with `make agent-finalize`. `RESULTS_DIR` is deliberately unset:
there is no qualifying exact-source successful full warm check run. Retained-run
maintenance will therefore be reported as skipped, not represented as current
verification evidence. No commit, push or deployment is part of finalization.


`env -u RESULTS_DIR make agent-finalize` passed at
`.cartulary/test-results/20260909T220047Z-p69617` before broader final checks.
Its `unit-artifacts/finalize-summary.json` reports generated artifacts unchanged,
zero updated files, no rollback needed, no failures, and retained-run selection,
run checks and performance evidence skipped because no results directory was
provided. No full warm check, retained-run duration refresh or performance claim
is asserted by this handoff.


### Final verification

| Command / scope | Result and retained root |
| --- | --- |
| `make generate-drift` (includes toolchain drift) | PASS 4/4 `.cartulary/test-results/20260909T220140Z-p73491` |
| `make json-shape-check` | PASS 3/3 `.cartulary/test-results/20260909T220140Z-p73508` |
| `make generated-artifact-policy-check` | PASS 3/3 `.cartulary/test-results/20260909T220140Z-p73499` |
| `make frontend-typecheck` | PASS 2/2 `.cartulary/test-results/20260909T220140Z-p73724` |
| `make frontend-import-boundary-check` | PASS 2/2 `.cartulary/test-results/20260909T220140Z-p73739` |
| `make test-slice OWNER=web.networkflow` | PASS 51/51 `.cartulary/test-results/20260909T220140Z-p73598` |
| Server logical extension-state adapter, readiness admission and policy catalog rows | PASS 3/3 `.cartulary/test-results/20260909T220140Z-p73620` |
| `make go-gosec-targeted` | PASS 4/4 `.cartulary/test-results/20260909T220216Z-p30125` |
| `make lint` (Go, Biome, scripts, shell and backend/frontend boundaries) | PASS 11/11 `.cartulary/test-results/20260909T220354Z-p81033` |
| `make build-web protocol-ts-browser-artifact-reachability` | PASS; current production build `.cartulary/test-results/20260909T220517Z-p94109`; reachability wrapper exit 0 |

Final lint initially found an unused superseded IP-type wrapper and formatting
in the last receipt test (`220216Z-p30116`, with the same formatting failure at
`220140Z-p73764`). Removing the unused wrapper and `make format` resolved both;
formatting passed at `220339Z-p76545`. The sequential check command stopped at
that initial Biome failure, so protocol browser-artifact reachability was run
explicitly afterward and again against the freshly built production output.

### Scope, limitations and rollback

The change is limited to the link owner/model/dialog/transport and Core registry
adapter, workbook lifetime/focus bindings, link-local backend admission/receipt/
transaction checks, authored contract and test routing, public generated outputs,
focused unit/service/browser tests, the eight reviewed goldens/manifest, and this
handoff. Core's exported classifier and the coordinator's read-only timeout
accessor expose existing facts; neither changes identity or transaction policy.
The digest and all adopted owners are unchanged. No dependency, route, observation
mode, browser persistence, compatibility alias, graph algorithm, canonical identity,
import/saved-graph lifecycle or database migration was added.

The schema identifier, transcript, selector, UUID/member and error corrections
require coordinated server/browser/projection delivery. Current contract-shaped
requests remain supported; previously admitted unsupported input and incompatible
retained receipts do not. Existing incompatible receipts block admission and
remain byte-preserved pending separately authorized remediation. No such
remediation, commit, push or deployment was performed.

Recovery is intentionally memory-only. Reloading the workbook/browser or losing
incident/profile access removes local recovery under the adopted cleanup rules.
Observation timeout or local abandonment cannot prove that server effects were
absent. The existing exact-replay contract is the recovery mechanism; there is
no binding-status endpoint, browser history ledger or compensating deletion.
Target discovery examines one bounded page at a time and does not claim global
absence. A binding receipt never implies that Core created a new indicator.

Retained-run maintenance, duration/performance refresh and full warm check evidence
were skipped because no qualifying exact-source successful full warm run exists.
Full `make check`, CI, release, deployment, migration and vulnerability-update
workflows were not run: owner-routed regression, real browser workflows, current
production build, lint/security and generation checks cover the touched scope;
no dependency, migration or release action is included. There is no performance
or release-conformance claim.

Rollback is the coherent set described in Decisions and compatibility: restore
matching browser, server, authored projections and their generated outputs, routing
and goldens together only when that set accepts retained receipts. Prefer forward
repair after corrected writes exist. Never remove committed indicators/bindings,
rewrite receipts/audit history, or apply an old binary solely to bypass admission.


IL-05 exit: the additional final ordinary `make browser-e2e-visual` passed 12/12
at `.cartulary/test-results/20260909T220140Z-p73892`. The first final
`make lint-markdown` passed at `220651Z-p95816`; `git diff --check` and the
tracked/untracked scope review passed. The 65 changed/new paths form the bounded
implementation, tests/projections/routing, reviewed goldens and handoff set.
Untracked source whitespace and the absence of new executable Markdown
references were also reviewed. Adopted owners and the advisory digest remain
unchanged. Branch `main` still points to
`e6d66bdb319f5d7f1986e7828c94de2b2ac137a7`, ahead of origin by 20 commits.
All work remains uncommitted; no reset, push or deployment occurred.

All five workstreams are DONE and every binary criterion is PASS. The final
handoff bytes receive the required repeated Markdown, diff and scope checks.

Next action: **none**.
