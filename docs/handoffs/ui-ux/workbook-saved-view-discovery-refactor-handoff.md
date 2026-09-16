# Workbook saved-view discovery and continuity handoff

## Execution control

Entry: clean `main`, HEAD `e72aa9a4180d663ddd885a90f2744447b91dc183`.
Root AGENTS.md is the only applicable repository instruction file. There are
no captured pre-existing edits. Revalidated branch, HEAD, index and worktree
before writing this file. No commit, push, deployment, dependency addition,
persistent browser storage, digest edit or analyst-data mutation is authorized.

This is the sole execution tracker. Save only the current row IN_PROGRESS;
record its actual evidence and save DONE before beginning its dependent.
An applicable BLOCKED row prevents completion and dependent execution.

| Workstream | Status | Dependency | Exit |
| --- | --- | --- | --- |
| SVD-01 Contract reconciliation | DONE | None | Consumers, adopted reads and browsing policy reconciled |
| SVD-02 Independent reads and bounded discovery | DONE | SVD-01 DONE | Authorized independent reads and bounded accepted pages |
| SVD-03 Integration and retirement | DONE | SVD-02 DONE | Selection, actions and recovery independent of discovery |
| SVD-04 Integrated evidence | DONE | SVD-03 DONE | All applicable behavior and lifecycle evidence passes |
| SVD-05 Final validation | DONE | SVD-04 DONE | Terminal checks and completed handoff verified |

## Authority and planning evidence

Core 01 REQ-01-138–147 and pagination own public resources and reads; Core 03
§2.3/REQ-03-295 and §2.4 own configuration and startup semantics; Core 04 owns
visibility/concealment. Design §8.3 supplies presentation and domain §9/§11
distinguishes base surfaces from saved configurations. The NLSpec research
essay and localized digest are advisory. The localized read order, relevant
historical authoring/startup/query/clipboard/reference/recovery handoffs and
Testing Harness generation/verification procedures were inspected in planning.
Historical evidence is not current acceptance. Product execution does not read
Markdown.

Planning commands: `make help`, `make help-all`, task guides for
module.savedviews, web.workbook, module.workbook, platform.openapi,
package.protocol_ts, web.architecture, web.design, web.application and
module.database_migrations passed. The baseline three selected web.workbook
controller/hook/adapter rows passed (4/4 harness units) at
`.cartulary/test-results/20260916T005602Z-p122153/run-summary.json`.

## Consumer and gap register

Every row below is a confirmed source limitation unless marked hypothesis.
Source placement and verification routing are separate: Saved Views owns its
SQL/application/API; web.workbook owns browser semantics; web.application binds
account/incident lifetime; module.workbook retains startup/preference ownership.

| Gap / consumers | Remediation and affected areas | Rationale / long-term benefit | Compatibility / migration | Unresolved risk | Binary validation |
| --- | --- | --- | --- | --- | --- |
| G01 Exhaustive loader withholds valid early pages; active-schema options filter incident-wide results locally | Core 01/03, API/SQL, adapter, discovery controller and tests: filter before paging and publish each page | Bounded work and first-page availability independent of catalog size | Additive optional filter; unfiltered clients/cursors retained; additive index | Live pages can shift under concurrent writes | A delayed or failed next page leaves first-page choices usable; memory/history stay bounded |
| G02 Selected configuration, Modified, Reset and view-bar identity derive from catalog membership | Independent addressed resource observation, hook/projection/selector tests | Resource identity survives discovery eviction and refresh | No persisted configuration change | Resource can become concealed after observation | Startup/write-selected resources survive unrelated page absence/failure; background reads never apply configuration |
| G03 All actions require list completion | Controller admission and action UI use current authority and relevant resource | Removes unrelated read dependency without weakening writes | CRUD/version/no-op contracts unchanged | Authority can change at dispatch | CRUD/Duplicate/Reset work through discovery failure with existing ownership/version/busy gates |
| G04 Version review and uncertain recovery require a complete list | Detail GET, separate recovery observation, operation/recovery UI and tests | Known identity is independently observable | No replay/idempotency change; uncertain create still may duplicate | Absence cannot establish a prior receipt | Recovery reads by captured ID; no automatic mutation; confirmed receipt survives read failure |
| G05 Startup seed is inserted into list; missing selection triggers list fallback | Seed independent resource; fallback only after authoritative detail unavailability and access classification | Preserves startup/preference authority and working edits | Startup endpoint and repair rules unchanged | Concealment must not be mistaken for incident revocation | Partial/transient reads never fallback or repair preferences; genuine unavailability uses existing base fallback |
| G06 Preference labels/availability use list, including home/default outside current page | Resolve inspected pointers by ID, bind selected observation independently | Removes a hidden catalog consumer | Pointer wire contracts unchanged | Stale names must be suppressed on authority loss | Labels and selected-target availability do not depend on discovery pages |
| G07 Client upsert alphabetizes and selector groups by scope | Keep page server order and show inline scope | Predictable opaque continuation without client ordering authority | Presentation changes only | Long names/narrow layouts are browser hypotheses | Production keyboard, scope, narrow, zoom and visual checks pass |
| G08 REQ-03-295 says Duplicate captures working layout while §2.3 requires saved source | Correct owner wording and acceptance; retain existing implementation semantics | One unambiguous duplication contract | No write behavior change | None after owner reconciliation | Modified working layout is not duplicated; selected saved layout is |
| G09 Existing Saved Views adapter recognized obsolete session/CSRF names rather than source-owned session_required/csrf_verification_failed | Saved Views adapter and adapter/service tests classify the current public codes into existing recovery categories | Correctly routes new read authentication failures to authority revalidation | Additive decoder recognition, no route or write/replay change | Other adapters retain their own owner boundaries | Real unauthenticated detail returns session_required; adapter maps it to authentication recovery and csrf_verification_failed to access recovery |
| G10 Browser image review confirmed CSS zoom doubled fixed popover offsets | SavedViewBrowser positioning and production a11y anchor assertion; this handoff records the finding | Correct anchoring and viewport bounds at supported zoomed widths | Presentation correction only; existing shell minimum-width policy remains | CSS zoom is synthetic browser evidence, not an OS screen-magnifier claim | At 200% CSS zoom and text spacing the popover remains within the viewport and within two pixels of the trigger bottom |
| G11 Final announcement audit found discovery failures copied into activation notice, creating duplicate live messages | Controller restricts activation notices to non-null resource IDs; browser leaves shared notices to the persistent announcement owner; selector regression asserts one failure message | One accessible message locus for each read failure | No read or write contract change | Actual assistive-technology pronunciation is outside automated Chromium evidence | Continuation failure renders one message and production accessibility/discovery rows pass |
| G12 Finalizer rejected the new index because the authored schema-ownership validator allocation stopped at migration 41 | Align schema-object-ownership.mjs allocation with the catalog generator for existing 42 and Saved Views 43; validate through JSON shape, finalizer and migration drift | Keeps independent generation and ownership validation consistent for additive migrations | No runtime or data change beyond the approved index | Allocation remains deliberately duplicated by existing harness design | JSON shape/finalizer and migration drift pass with migration 43 owned by savedviews |
| G13 Import checking found a type cycle between the read owners and operation model | Move the injected observation port type to WorkbookSavedViewPort; operation and read owners depend on that lower boundary; import and controller tests verify it | Maintains cohesive owners with an acyclic production graph | Type placement only, no runtime API change | None after import validation | Frontend import check and focused owner tests pass |
| G14 Broad selector-policy scan finds 24 pre-existing heading readiness waits | Deferred to web.architecture selector-maintenance work; preserve unrelated files and retain baseline comparison evidence | Semantic readiness selectors improve future test resilience | No change in this seam; policy and all offending expressions are byte-identical to HEAD | The broad pre-existing policy row remains failing | This seam introduces zero additional violations; owner-specific import/source and selector behavior checks pass |

## Adopted decisions

The authorized plan adopts a compact paged popover: 50 candidates, one accepted
page, at most ten previous page-start cursors plus current/next cursors. Previous
refetches; First restarts beyond retained history. No auto-prefetch or exhaustive
traversal. Close cancels pending browsing/activation but retains the committed
page/history; reopen revalidates it. Schema change clears discovery only.
Explicit candidate activation verifies by ID before applying query/layout and
identity; late activation cannot overwrite a newer selection or working revision.
Selected, operation and inspected preference resources are retained separately.

Public reads: optional exact view_schema_id collection filter; existing resource
and envelope/order/limit semantics; authorized non-locking detail GET; no-store;
strict query validation and cursor binding; same concealed 404 for unavailable
targets. Reads never repair preferences or mutate saved/source configuration.
Create/Update capture working configuration; Duplicate uses saved configuration;
Reset is local; uncertain create never gains automatic replay or receipt inference.

## Execution log

SVD-01 entry: baseline revalidated unchanged. Next: adopt exact owner wording,
review consumer dispositions and compatibility, then document the exit.

### SVD-01 exit — DONE

Changed Core 01 route inventory, REQ-01-144/144A/144B and read-error vocabulary;
Core 03 REQ-03-022A and corrected REQ-03-295; Core 04 AC-151A/480A; design §8.3.
All G01–G08 consumers have dispositions. The Duplicate contradiction is resolved
in adopted wording, with no write semantics changed. Browsing bounds, dismissal,
activation, independent resource retention and failure classification are explicit.
Compatibility: existing resource/CRUD envelopes and unfiltered cursors remain;
filtered reads/detail GET are additive; formerly ignored malformed query members
now fail closed. Index migration requires no data rewrite; backend precedes UI.

`make lint-markdown` initially failed on the new error table's missing three
cells (related change), run `20260916T010410Z-p126089`; corrected that row.
Rerun PASS at `20260916T010550Z-p128116/adhoc/lint-markdown/tool-run-summary.json`.
`git diff --check` PASS. No unresolved adopted-owner contradiction remains.
Browser geometry/keyboard risks remain hypotheses requiring SVD-04 evidence.
Next: SVD-02 independent reads, projections and bounded observation owners.

SVD-02 entry: SVD-01 saved DONE. Implementing adopted reads and independent owners.

### SVD-02 exit — DONE

Implemented source-owned non-locking visible-resource lookup and schema-filtered
keyset discovery in `internal/modules/savedviews/{application,store,routes,read_query}.go`
and authored `db/queries/savedviews.sql`. Migration 00043 adds the schema/order
index; existing unfiltered SQL/index remain. Both reads enforce current incident
membership and the same private/shared/system visibility predicate. Detail and
collection reads use no-store and never resolve/repair startup preferences.

Authored projections changed: Saved Views OpenAPI fragment, public error registry,
protocol-ts operation allowlist, migration catalog owner input and verification
rows. Make generation produced SQL/API/client/catalog outputs. Candidate OpenAPI
2.0.0 change-set records the three new reviewed fingerprints; immutable 1.0.0
release history was not changed. The compatibility checker conservatively calls
the changed parameter array breaking; the reviewed optional filter remains
compatible with existing unfiltered callers and legacy cursors. Backend and index
must precede the new frontend.

Frontend port/adapter now expose getResource and schema-scoped listPage.
`SavedViewDiscovery.ts` owns one accepted page and ten prior checkpoints;
`SavedViewResourceObserver.ts` retains only named consumer slots. Their tests
cover delayed/failing continuation and exact retry, bounded repeated browsing,
dismissal/schema fencing, retained selection, release, receipt/deletion fencing,
and unavailable versus transient observation. They do not own configuration or
mutations. The old UI loader remains temporarily until SVD-03 consumer migration.

Verification:

- Initial focused runs at `20260916T011552Z-p143915`, `-p143926`, `-p143952`
  and `-p144110` failed on the new codec pointer assignment and missing authored
  protocol operation allowlist entry. Both related implementation defects fixed.
- `make generate` initially required explicit compatibility review at
  `20260916T010817Z-p130156`; reviewed candidate changes and reran successfully.
  Latest generation PASS: `20260916T011900Z-p156093/generate/tool-run-summary.json`.
- Saved Views query unit slice PASS 1/1 at `20260916T012035Z-p183365`.
- Workbook independent-read and adapter slices PASS 3/3 at
  `20260916T012035Z-p183373`.
- Saved Views independent-read service slice PASS 3/3 at
  `20260916T012035Z-p183394`: 101 unrelated-schema resources before active-schema
  results, deterministic filtered continuation, private/system/shared visibility,
  detail/list agreement, concealed absence, membership revocation and untouched
  invalid home pointer. These use isolated harness data only.
- `make frontend-typecheck` PASS 2/2 at `20260916T011926Z-p177695`.
- `make format` PASS 2/2 at `20260916T012010Z-p178911`; `git diff --check` PASS.

Run roots are under `.cartulary/test-results/`; graph summaries are run-summary.json.
A transient generated-file checksum diagnostic appeared during the first service
build concurrent with generation; both that run and the subsequent isolated
repeat passed. Subsequent generation and dependent compilation are sequential.
No blocking read-contract issue remains. Browser integration, authority lifecycle
and production rendering are SVD-03/04 work, not claimed by these foundation tests.
Next: integrate selection, actions, preferences and recovery, then retire catalog
accumulation and invalid-selection inference.

SVD-03 entry: SVD-02 saved DONE. Integrating independent resource and discovery owners.

### SVD-03 exit — DONE

Integrated discovery and resource owners into WorkbookSavedViewController and its
hook. Selected identity/version, Modified and Reset now use a retained addressed
observation; the list never owns selection. Startup and confirmed writes seed
observations directly. Candidate activation reads by ID and fences authority,
selection intent and working revision before applying query/layout/identity.
Background version updates announce a new saved baseline without changing work.

Added `components/SavedViewBrowser.tsx`: compact popover, server-order candidates,
inline scope, explicit activation, page controls/retry, base choice, keyboard
focus and dismissal. Unrelated discovery failure does not disable CRUD, Duplicate
or local Reset. Known-target recovery reads the captured ID; apply requires a
fresh usable resource revision. Keep/new-create review no longer waits for a list
or unrelated read. The existing duplicate-risk acknowledgement remains in UI;
there is no inferred create receipt or replay. Confirmed receipts survive failed
follow-up reads. Authoritative selected-resource absence invokes the existing
base identity fallback only after incident-access classification, preserving work.

Preference labels resolve inspected home/default pointers independently via two
bounded resource slots; closing inspection releases them. Preference repair stays
with startup. Updated view-bar identity and source guides. Removed
`loadSavedViewList.ts`, exhaustive pagination accumulation, alphabetical catalog
upsert/removal, list-derived invalid-selection projection and complete-list gates.
The semantic frontend port now requires schema-scoped discovery. Browser read-by-ID
helpers use the detail route; native selector callers were migrated to activation.
Existing scope selectors remain unchanged.

Focused evidence (roots under `.cartulary/test-results/`):

- Controller, production selector component, hook and full shell surface suite:
  PASS 5/5 harness units, `20260916T013720Z-p228879`.
- Independent read owners and adapter: PASS 3/3, `20260916T013804Z-p234671`.
- Preference characterization/owner and startup admission: PASS 4/4,
  `20260916T013815Z-p235405`.
- Browser saved-view helper boundary: PASS 2/2, `20260916T013815Z-p235412`.
- Frontend typecheck: PASS 2/2, `20260916T013722Z-p230221`.
- Format PASS 2/2, `20260916T013719Z-p228715`; git diff --check PASS.

Intermediate failures were related test migrations: old catalog/native-select
assumptions, a Reset assertion after Duplicate had navigated away, a non-spy test
fixture, an unused selector variable and widened mock literal typing. Repaired
and covered by the passing slices above; no unexplained baseline failure.
These checks establish integration semantics, including all five actions through
unrelated discovery failure, candidate working/selection/dismissal races, selected
retention and deletion fencing. Production-browser geometry, a11y, real route
interception and visual evidence remain SVD-04 work. No dependent blocker remains.
Next: complete the contract/lifecycle matrix and real-browser regression evidence.

SVD-04 entry: SVD-03 saved DONE. Completing contract/lifecycle and production-browser evidence.

### SVD-04 evidence collected

Added public-route discovery coverage with 53 Timeline resources and newer Notes
resources. The browser verifies no list request before opening, delayed first
page, 50 immediately usable candidates, retained choices through failed/delayed
continuation, exact retry, Previous/Next refetch, startup identity outside the
page, explicit keyboard activation, destination identity, applied grouping and
hidden-column layout. No automatic traversal occurs.

Added production accessibility coverage across compact/default/comfortable
density, 1440/768 widths, long names, scope, roving candidate focus, Tab to page
controls, Escape/focus return, outside dismissal, text spacing and 200% CSS zoom.
The existing shell's below-supported-minimum rule still hides configuration
controls at 720 effective CSS pixels; the browser is exercised at 960 effective
pixels. This is not an operating-system screen-magnifier claim.

Expanded source-owner service evidence with current-role changes, foreign-incident
concealment, session loss, closed-incident reads and configuration writes.
Expanded controller authority replacement/disposal tests to include delayed list
and addressed reads, alongside captured write callbacks. All test data resides in
isolated harness databases/accounts.

Related failures and resolutions:

- Service run `20260916T014219Z-p276345` expected the obsolete authentication code;
  corrected to source-owned session_required and added adapter recognition (G09).
  Read/lifecycle repeat PASS 3/3 at `20260916T014442Z-p300535`.
- New a11y row initially lacked generated browser grouping; `make generate` PASS
  at `20260916T014851Z-p399638` supplied the authored routing projection.
- A11y run `20260916T014927Z-p402913` attempted the selector below the established
  minimum width. Corrected the scenario, then image review found and fixed G10.
- `20260916T015802Z-p556238` passed a11y, public collection, version/no-op and
  delayed-receipt rows; uncertain-create coverage failed because its old glob
  intercepted only the collection. It now fails the addressed read too. Repeat
  with a11y and visual comparison PASS 15/15 at `20260916T020049Z-p631272`.
- Initial visual comparison `20260916T014444Z-p301824` correctly rejected changed
  selector goldens; its existing saved-view a11y row passed. Image review led to
  visible scope/arrow allocation and removal of the redundant View prefix on a
  selected resource, avoiding overlap with Modified. No tolerance was changed.

Current passing focused evidence (all roots under .cartulary/test-results):

| Acceptance area | Result | Evidence |
| --- | --- | --- |
| Filtered discovery, strict validation, legacy cursors, ordering, visibility parity and concealment | PASS | SVD-02 query/read rows; expanded read run 20260916T014442Z-p300535 |
| Bounded page/history and resource retention, failed continuation, exact retry, stale reads/deletion fences | PASS | independent-read slice 20260916T013804Z-p234671; controller 20260916T015518Z-p480921 |
| Selection outside page, refresh/eviction, activation generation and working-configuration fences | PASS | controller/hook/shell 20260916T013720Z-p228879; production browser 20260916T015919Z-p594806 (11/11) |
| All five actions during discovery failure; saved Duplicate, local Reset, ownership/system/version gates | PASS | controller 20260916T014927Z-p402934 and 20260916T015518Z-p480921; lifecycle 20260916T014442Z-p300535 |
| Normalized no-op, version review, uncertain creation, acknowledged write then failed read | PASS | authoring rows 20260916T015802Z-p556238 and uncertain-create repeat 20260916T020049Z-p631272 |
| Session/account/incident replacement, disposal, role demotion and authoritative access loss | PASS | controller/adapter 20260916T014927Z-p402934; expanded controller 20260916T015518Z-p480921; service read above |
| Genuine selected unavailability versus transient/partial absence; startup and preference ownership | PASS | controller/hook and preference slices SVD-03; startup/preference/service + grid autosave 20260916T014927Z-p402924 (13/13) |
| Real persisted query/layout, home/default preferences and saved destination | PASS | stateful saved-view/preferences row 20260916T015518Z-p480919 (13/13 with a11y) |
| Timeline direct entry, Enter/Tab variants, Escape/caret, range, clipboard, fill-down and qualifying-input creation | PASS | four Timeline browser rows 20260916T014704Z-p361107 (13/13); grid autosave row above |
| Accessible browser keyboard, dismissal, focus, names/scopes, density, narrow width, zoom/spacing and nonoverlap | PASS | 20260916T020321Z-p672719 (11/11); selector component 20260916T020322Z-p672958 (2/2) |

Full suite/release/performance completeness is not claimed. Visual and accessibility
artifacts are implementation/design evidence, not Core 05 publication evidence.

Closed-incident configuration coverage now explicitly creates, updates, observes
and deletes a saved configuration after closing the incident. The first expanded
fixture omitted required query_json (related test failure at
`20260916T020523Z-p743058`); corrected the fixture. Final read service PASS 3/3
at `20260916T020755Z-p792788`. The real inspector/closed-incident saved-view switch
and Timeline authoring scenario PASS 11/11 at `20260916T020651Z-p761027`.

Visual refresh uses only `make browser-e2e-visual-update`; its public surface has
no row filter, so the full visual corpus is the necessary refresh boundary.
Initial full refresh PASS 12/12 at `20260916T015539Z-p510560`; reconciliation
accounts for all 252 captures/goldens, zero orphan/missing/ambiguous mappings.
Reviewed all 101 changed selector strips in before/after contact sheets and
outlying full-image diffs, retained as selector-review-0..5.png and
saved-view-golden-differences.json in that run root. Substantive changes are the
button label, scope/indicator, selected focus return and associated local chrome;
minor non-selector raster differences are edge/shadow quantization (mostly
1–3 color levels), with no changed geometry or content. No golden was hand-edited.

The later refresh at `20260916T020324Z-p673863` failed fixture readiness on an
aborted vendored-font request during reload, not a screenshot comparison. The subsequent full refresh and ordinary comparison below resolve this
related failure; neither the failed run nor historical goldens establish PASS.

### Visual refresh record

Accepted trigger: adopted paged selector replaces the native exhaustive selector.
Final refresh PASS 12/12 at `20260916T020831Z-p810607`; all 252 captures reconcile,
zero missing/orphan/ambiguous goldens. Renderer profile remains
`visual.renderer.playwright_1_59_1_chromium_1217_linux_amd64`. No viewport, zoom,
mask, scroll normalization, screenshot scope or comparison tolerance changed.
The new independent browser scenarios are additional evidence, not replacement
fixture geometry. All changed goldens contain the affected saved-view chrome.

Row aliases below are exact authored catalog IDs, not inferred from filenames.
A dash in the fixture column denotes an active nonregistry capture, whose exact
scenario/project/profile is recorded in the reconciliation artifact.

| Alias | Semantic owner row |
| --- | --- |
| V01 | module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1 |
| V02 | module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c |
| V03 | module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc |
| V04 | module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7 |
| V05 | module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7 |
| V06 | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc |
| V07 | module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206 |
| V08 | module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0 |
| V09 | module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea |
| V10 | module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67 |
| V11 | module.workbook.visual.contextual_task_decision_creation |
| V12 | module.workbook.visual.coordination_create_authoring_recovery |
| V13 | module.workbook.visual.decision_supersession_review_recovery |
| V14 | module.workbook.visual.indicator_lifecycle_authoring |
| V15 | module.workbook.visual.indicator_observations_authoring |
| V16 | module.workbook.visual.note_create_authoring_recovery |
| V17 | module.workbook.visual.ordinary_create_authoring_recovery |
| V18 | module.workbook.visual.preferences |
| V19 | module.workbook.visual.timeline_capture_actions |
| V20 | module.workbook.visual.timeline_related_evidence |
| V21 | web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc |
| V22 | web.design.visual.lifecycle |
| V23 | web.design.visual.membership_audit_browsing |
| V24 | web.design.visual.membership_management_visual |
| V25 | web.design.visual.metadata_editing |

All filenames below are under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
The Make-owned golden SHA-256 manifest was refreshed with the images.

| Golden filename | Row | Stable fixture IDs |
| --- | --- | --- |
| account-menu-controls-compact-linux.png | V21 | — |
| account-menu-controls-narrow-linux.png | V21 | — |
| account-menu-long-label-text-spacing-linux.png | V21 | — |
| account-menu-workbook-root-linux.png | V21 | — |
| collaboration-conflict-resolver-linux.png | V01 | visual.fixture.same_field_conflict |
| collaboration-grid-blocked-conflict-linux.png | V03 | — |
| collaboration-grid-conflict-resolver-compact-linux.png | V02 | — |
| collaboration-grid-conflict-resolver-linux.png | V02 | — |
| collaboration-grid-conflict-resolver-narrow-linux.png | V02 | — |
| collaboration-presence-markers-linux.png | V01 | visual.fixture.presence_overflow |
| contextual-decision-authoring-linux.png | V11 | — |
| contextual-decision-authoring-narrow-linux.png | V11 | — |
| contextual-decision-recovery-linux.png | V11 | — |
| contextual-decision-recovery-narrow-linux.png | V11 | — |
| contextual-decision-references-narrow-linux.png | V11 | — |
| contextual-task-request-authoring-linux.png | V11 | — |
| contextual-task-request-authoring-narrow-linux.png | V11 | — |
| contextual-task-request-recovery-linux.png | V11 | — |
| contextual-task-request-recovery-narrow-linux.png | V11 | — |
| contextual-task-request-references-narrow-linux.png | V11 | — |
| coordination-comm-log-authoring-linux.png | V12 | visual.fixture.contextual_coordination_creation |
| coordination-handoff-authoring-linux.png | V12 | visual.fixture.contextual_coordination_creation |
| coordination-lesson-authoring-linux.png | V12 | visual.fixture.contextual_coordination_creation |
| coordination-recovery-linux.png | V12 | visual.fixture.contextual_coordination_creation |
| coordination-source-narrow-linux.png | V12 | visual.fixture.contextual_coordination_creation |
| coordination-status-review-authoring-linux.png | V12 | visual.fixture.contextual_coordination_creation |
| decision-supersession-accepted-linux.png | V13 | — |
| decision-supersession-review-linux.png | V13 | — |
| decision-supersession-review-narrow-linux.png | V13 | — |
| entity-mention-chip-states-linux.png | V04 | visual.fixture.mention_chip_state_matrix |
| incident-directory-compact-desktop-workbook-shell-linux.png | V08 | visual.fixture.compact_desktop_workbook_shell |
| incident-directory-default-timeline-workbook-shell-linux.png | V08 | visual.fixture.default_timeline_workbook_shell |
| incident-directory-narrow-desktop-workbook-shell-linux.png | V08 | visual.fixture.narrow_desktop_workbook_shell |
| indicator-lifecycle-authoring-linux.png | V14 | visual.fixture.indicator_lifecycle_authoring |
| indicator-lifecycle-authoring-narrow-linux.png | V14 | visual.fixture.indicator_lifecycle_authoring |
| indicator-observation-authoring-linux.png | V15 | visual.fixture.indicator_observations_authoring |
| indicator-observation-authoring-narrow-linux.png | V15 | visual.fixture.indicator_observations_authoring |
| lifecycle-closed-linux.png | V22 | — |
| lifecycle-confirmed-refresh-failure-linux.png | V22 | — |
| lifecycle-pending-linux.png | V22 | — |
| lifecycle-reason-linux.png | V22 | — |
| lifecycle-review-linux.png | V22 | — |
| lifecycle-uncertain-linux.png | V22 | — |
| linked-note-authoring-linux.png | V16 | — |
| linked-note-authoring-narrow-linux.png | V16 | — |
| linked-note-recovery-linux.png | V16 | — |
| linked-note-recovery-narrow-linux.png | V16 | — |
| linked-note-source-narrow-linux.png | V16 | — |
| membership-audit-cursor-recovery-linux.png | V23 | — |
| membership-audit-empty-linux.png | V23 | — |
| membership-audit-inspected-linux.png | V23 | — |
| membership-audit-loading-linux.png | V23 | — |
| membership-audit-stale-linux.png | V23 | — |
| membership-management-confirmed-refresh-failure-linux.png | V24 | — |
| membership-management-loading-linux.png | V24 | — |
| membership-management-pending-linux.png | V24 | — |
| membership-management-role-linux.png | V24 | — |
| membership-management-uncertain-linux.png | V24 | — |
| metadata-conflict-linux.png | V25 | — |
| metadata-dirty-linux.png | V25 | — |
| metadata-loading-linux.png | V25 | — |
| metadata-saving-linux.png | V25 | — |
| metadata-uncertain-linux.png | V25 | — |
| ordinary-recovery-1280-linux.png | V17 | — |
| ordinary-reference-authoring-linux.png | V17 | — |
| record-relationships-mention-chips-linux.png | V05 | — |
| timeline-grid-timeline-default-linux.png | V07 | — |
| timeline-mutation-pending-replay-status-linux.png | V10 | — |
| timeline-mutation-transaction-recovery-panel-compact-linux.png | V10 | — |
| timeline-mutation-transaction-recovery-panel-linux.png | V10 | — |
| timeline-mutation-transaction-recovery-panel-narrow-linux.png | V10 | — |
| timeline-related-evidence-authoring-linux.png | V20 | — |
| timeline-related-evidence-authoring-narrow-linux.png | V20 | — |
| timeline-related-evidence-partial-linux.png | V20 | — |
| timeline-related-evidence-partial-narrow-linux.png | V20 | — |
| timeline-related-evidence-party-narrow-linux.png | V20 | — |
| timeline-supersession-accepted-linux.png | V19 | — |
| timeline-supersession-authoring-linux.png | V19 | — |
| timeline-supersession-review-linux.png | V19 | — |
| timeline-supersession-review-narrow-linux.png | V19 | — |
| workbook-inspector-compact-actions-linux.png | V09 | visual.fixture.inspector_compact_actions |
| workbook-inspector-history-linux.png | V09 | visual.fixture.base_inspector |
| workbook-inspector-narrow-technical-details-linux.png | V09 | visual.fixture.inspector_narrow_technical_details |
| workbook-inspector-public-error-linux.png | V09 | visual.fixture.base_inspector |
| workbook-inspector-relationships-linux.png | V09 | visual.fixture.base_inspector |
| workbook-inspector-rollback-preview-linux.png | V09 | visual.fixture.base_inspector |
| workbook-preferences-confirmed-stale-linux.png | V18 | — |
| workbook-preferences-uncertain-linux.png | V18 | — |
| workbook-preferences-unset-linux.png | V18 | — |
| workbook-query-empty-text-spacing-linux.png | V06 | visual.fixture.empty_successful_query |
| workbook-query-saved-view-query-controls-linux.png | V06 | visual.fixture.saved_view_query_controls_and_grouped_result |
| workbook-view-bar-filter-editing-overflow-linux.png | V06 | — |
| workbook-view-bar-long-columns-linux.png | V06 | — |
| workbook-view-bar-maximum-pressure-base-linux.png | V06 | — |
| workbook-view-bar-maximum-pressure-compact-linux.png | V06 | — |
| workbook-view-bar-maximum-pressure-narrow-linux.png | V06 | — |
| workbook-view-bar-ordered-maximum-sort-linux.png | V06 | — |
| workbook-view-bar-saved-view-actions-linux.png | V06 | — |
| workbook-view-bar-saved-view-clean-linux.png | V06 | — |
| workbook-view-bar-saved-view-modified-linux.png | V06 | — |
| workbook-view-bar-text-spacing-linux.png | V06 | — |

### SVD-04 exit — DONE

Final ordinary saved-view visual comparison PASS 11/11 at
`20260916T021327Z-p881427`. Final stateful saved query/layout/home/default helper
migration PASS 11/11 at `20260916T021020Z-p848982`. The G11 single-message regression
and controller slice PASS 3/3 at `20260916T021514Z-p918963`; production discovery
and accessibility repeat PASS 13/13 at `20260916T021515Z-p919230`. The zoomed
popover image was inspected again after its anchor correction; it is aligned to
the trigger and bounded within the viewport. Final selected Modified golden was
inspected after removing the redundant prefix; scope/arrow/badge do not overlap.

All applicable contract/lifecycle rows in the matrix have passing evidence.
Closed configuration writes and the real inspector/closed surface regression
pass as recorded above. No unresolved applicable failure remains. The font-load
failure was transient harness fixture readiness; the successful full repeat
accounts for every active capture. Markdown lint PASS at
`20260916T021344Z-p910287/adhoc/lint-markdown/tool-run-summary.json`;
`git diff --check` PASS. Final formatting will be checked again in SVD-05.
Next: owner-routed terminal verification and completed deployment/rollback handoff.

SVD-05 entry: SVD-04 saved DONE. Current branch and HEAD still match entry.
Running agent-finalize with RESULTS_DIR explicitly unset; no qualifying successful
full warm-check run was supplied for retained-run maintenance.

Initial agent-finalize failed JSON shape validation at `20260916T021640Z-p953769`;
direct diagnostic repeat at `20260916T021709Z-p954384` identified G12. Updated
the authored validator allocation, not generated manifest bytes.

Terminal verification found and fixed G13 (type-import cycle), the authored
canonical migration-catalog hash expectation for the additive migration, and
non-null assertion lint warnings in new browser assertions. Initial affected
runs: architecture `20260916T021936Z-p967509`, migration characterization
`20260916T021953Z-p972131`, Biome `20260916T021953Z-p972300`. The migration drift
itself already passed 5/5 at `20260916T021953Z-p972032`.

The architecture scan also found G14: 24 existing heading-name readiness waits
across unrelated scenarios. Read-only comparison against captured HEAD proves
all 24 offending expressions and the policy source are unchanged, with zero
additional violations. Evidence: `20260916T021936Z-p967509/selector-policy-baseline-comparison.json`.
This is a retained baseline FAIL, not a PASS or an unexplained applicable failure.
Migrating those existing headings is N/A to this saved-view seam (web.architecture
owns that separate maintenance). The changed seam continues to use shared saved-view
selector builders and semantic dialog/option roles; its applicable source/import
and behavioral checks must pass. No unrelated tests or policy were weakened.

## Final scope and behavior

The implemented seam adds active-schema collection paging and a stable-ID detail
read through Saved Views SQL/application/repository/API, generated protocol clients
and the workbook semantic port. Discovery has one 50-resource page and ten prior
checkpoints. Addressed observations have five named retention slots (selection,
activation, operation, inspected home and default), coalescing shared identities;
they do not form another catalog. The operation controller retains drafts,
captured attempts and receipts independently of discovery and working state.

Retired complete-catalog dependencies: selector candidates; selected label/version,
Modified and Reset; action admission; startup seeding; deletion/missing-selection
inference; version review and known-target recovery; preference labels/availability;
and view-bar identity. No exhaustive UI read remains. The existing Saved Views
incident-bundle export enumerates configuration source rows for its own export
contract; it does not participate in UI discovery or admission.

Browsing, focus, paging and refresh never apply a configuration. Explicit activation
resolves the candidate by ID and fences selection/working/authority generations.
Current authorized receipts and selected resources survive page eviction and
unrelated read failure. Partial absence cannot prove deletion; genuine unavailable
reads require access classification before base identity fallback, retaining work.
Create/Update use working configuration; Duplicate uses the accepted saved source;
Reset remains local. No row replay or idempotent-create semantics were added.

Current authority, version, scope, system immutability and mutation-busy rules remain.
The source read queries are non-locking and never repair preferences. Startup and
preference repair remain with their existing owner. Timeline editing/capture,
range, keyboard, clipboard, fill-down and recovery navigation evidence is recorded
above. The compact browser exposes page-local counts, useful partial results and
local retry, with server order and opaque cursors preserved.

## Compatibility, deployment and rollback

Existing CRUD envelopes, stored query/layout objects, versions and unfiltered
clients remain compatible. A missing legacy cursor schema binding means unfiltered
only. The optional exact schema filter and detail GET are additive. Previously
ignored unsupported/malformed queries now fail closed intentionally; no data rewrite
is required. The candidate OpenAPI release change-set records the conservative
parameter-array compatibility fingerprint; immutable release history is unchanged.

Deploy migration 00043 and the backend before the new frontend. Migration 00043
adds only the schema/order index and uses the existing migration transaction
mechanics; scheduling its index build remains a deployment-owner concern. No
migration was applied to analyst data during this work.

Rollback the frontend before rolling back the backend. The additive index may
remain. For source rollback, revert only the exact task paths listed here to
captured HEAD e72aa9a4180d663ddd885a90f2744447b91dc183 and remove task-added files
after review; do not reset unrelated subsequent work. Entry had no pre-existing
edits. No commit, push or deployment was performed.

## Changed-file inventory

The 101 PNG paths and their semantic mappings are listed in the visual record
above. The remaining exact paths follow (A added, M modified, D retired). Generated
SQL/API/client/catalog/topology/golden-manifest changes came from Make-owned targets.

| Change | Path |
| --- | --- |
| M | `apps/web/e2e/incident-administration.spec.ts` |
| M | `apps/web/e2e/inspector-actions.spec.ts` |
| M | `apps/web/e2e/keyboard.spec.ts` |
| M | `apps/web/e2e/support/workbook/savedViews.test.ts` |
| M | `apps/web/e2e/support/workbook/savedViews.ts` |
| M | `apps/web/e2e/workbook-preferences.spec.ts` |
| A | `apps/web/e2e/workbook-saved-view-discovery.spec.ts` |
| M | `apps/web/e2e/workbook.a11y.spec.ts` |
| M | `apps/web/e2e/workbook.spec.ts` |
| M | `apps/web/e2e/workbook.visual.spec.ts` |
| M | `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx` |
| M | `apps/web/src/workbook/WorkbookShell.tsx` |
| M | `apps/web/src/workbook/adapters/createWorkbookSavedViewAdapter.test.ts` |
| M | `apps/web/src/workbook/adapters/createWorkbookSavedViewAdapter.ts` |
| M | `apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.test.tsx` |
| M | `apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.tsx` |
| M | `apps/web/src/workbook/components/README.md` |
| A | `apps/web/src/workbook/components/SavedViewBrowser.tsx` |
| M | `apps/web/src/workbook/components/SavedViewRecovery.tsx` |
| M | `apps/web/src/workbook/components/WorkbookShellViewBarControls.tsx` |
| M | `apps/web/src/workbook/hooks/useWorkbookSavedViewController.test.tsx` |
| M | `apps/web/src/workbook/hooks/useWorkbookSavedViewController.ts` |
| M | `apps/web/src/workbook/models/workbookSavedViewControl.test.ts` |
| M | `apps/web/src/workbook/models/workbookSavedViewControl.ts` |
| M | `apps/web/src/workbook/models/workbookSavedViewPaginationMachine.test.ts` |
| M | `apps/web/src/workbook/models/workbookSavedViewPaginationMachine.ts` |
| M | `apps/web/src/workbook/models/workbookSavedViewRuntime.test.ts` |
| M | `apps/web/src/workbook/models/workbookSavedViewRuntime.ts` |
| M | `apps/web/src/workbook/models/workbookViewBarWorkingSet.ts` |
| M | `apps/web/src/workbook/ports/README.md` |
| M | `apps/web/src/workbook/ports/WorkbookSavedViewPort.ts` |
| M | `apps/web/src/workbook/preferences/WorkbookPreferenceController.ts` |
| M | `apps/web/src/workbook/preferences/workbookPreferenceCharacterization.test.tsx` |
| M | `apps/web/src/workbook/preferences/workbookPreferenceModel.ts` |
| M | `apps/web/src/workbook/savedviews/README.md` |
| A | `apps/web/src/workbook/savedviews/SavedViewDiscovery.ts` |
| A | `apps/web/src/workbook/savedviews/SavedViewResourceObserver.ts` |
| M | `apps/web/src/workbook/savedviews/WorkbookSavedViewController.test.ts` |
| M | `apps/web/src/workbook/savedviews/WorkbookSavedViewController.ts` |
| D | `apps/web/src/workbook/savedviews/loadSavedViewList.ts` |
| M | `apps/web/src/workbook/savedviews/savedViewOperationModel.ts` |
| A | `apps/web/src/workbook/savedviews/savedViewReads.test.ts` |
| M | `contracts/errors/index.json` |
| M | `contracts/openapi-releases/2.0.0.change-set.json` |
| M | `contracts/openapi-source/owners/module.savedviews/openapi.json` |
| M | `contracts/openapi/cartulary.openapi.yaml` |
| M | `contracts/protocol-ts/http-operations.v2.json` |
| A | `db/migrations/00043_saved_views_schema_discovery.sql` |
| M | `db/queries/savedviews.sql` |
| M | `docs/design.md` |
| A | `docs/handoffs/ui-ux/workbook-saved-view-discovery-refactor-handoff.md` |
| M | `docs/spec/01_architecture_storage_and_view_contracts.md` |
| M | `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` |
| M | `docs/spec/04_security_deployment_and_conformance.md` |
| M | `internal/gen/contracterrors/artifacts_gen.go` |
| M | `internal/gen/contractopenapi/artifacts_gen.go` |
| M | `internal/gen/openapioperations/catalog_gen.go` |
| M | `internal/gen/sql/savedviews.sql.go` |
| M | `internal/modules/database_migrations/catalog_characterization_test.go` |
| M | `internal/modules/savedviews/application.go` |
| M | `internal/modules/savedviews/application_test.go` |
| A | `internal/modules/savedviews/read_contract_test.go` |
| A | `internal/modules/savedviews/read_query.go` |
| A | `internal/modules/savedviews/read_query_test.go` |
| M | `internal/modules/savedviews/routes.go` |
| M | `internal/modules/savedviews/store.go` |
| M | `packages/protocol-ts/src/generated/error-registry.ts` |
| M | `packages/protocol-ts/src/generated/http-operation-bindings.ts` |
| M | `tools/browser_e2e_batch_manifest.json` |
| M | `tools/database-migrations/generate-catalog-projections.mjs` |
| M | `tools/execution_topology_render_index.json` |
| M | `tools/frontend_source_ownership.json` |
| M | `tools/frontend_visual_golden_manifest.json` |
| M | `tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership.mjs` |
| M | `tools/migration_history_manifest.json` |
| M | `tools/schema_object_ownership_manifest.json` |
| M | `tools/test_families/module.savedviews.json` |
| M | `tools/test_families/web.workbook.json` |

## Terminal verification ledger

All roots below are under `.cartulary/test-results/`. Each graph run retains
`run-manifest.json` (exact selected inputs), `run-summary.json`, per-row results,
unit logs and browser/report artifacts where applicable. Counts are harness
units, not assertions or independent acceptance claims. Repeated units across
runs are not added into a misleading total. Commands show the public Make target
and recorded owner/row selection; standard worker defaults are in the manifests.

`make generate` PASS at `20260916T021824Z-p956110/generate/tool-run-summary.json`.
It resolved the stale generated topology input after the schema-owner validator
fix (intermediate finalizer/diagnostic failures at `20260916T021754Z-p955130`
and `20260916T021810Z-p955594`). Final authored formatting:
`make format` PASS 2/2 at `20260916T022213Z-p979362`.

`env -u RESULTS_DIR make agent-finalize` ran before broader terminal verification
and passed at `20260916T021850Z-p959187`; the final repeat below also passes.
Its Make-owned maintenance includes JSON shape, catalog and generated drift
validation. Retained-run maintenance/performance evidence was **skipped** because
RESULTS_DIR was unset: no qualifying successful full warm-check evidence was
supplied. This does not claim a full warm check.

| Command | Result | Run root |
| --- | --- | --- |
| `env -u RESULTS_DIR make agent-finalize` | PASS 1/1 | `20260916T022338Z-p985330` |
| `make test-slice OWNER=platform.openapi` | PASS 4/4 | `20260916T021936Z-p967476` |
| `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.family_registries_and_account_types_i2,package.protocol_ts.frontend_unit.generated_http_operation_bindings` | PASS 3/3 | `20260916T021936Z-p967492` |
| `make test-slice OWNER=module.savedviews ROWS=module.savedviews.unit.independent_read_query_contract,module.savedviews.unit.saved_view_application_policy_preserves_visibili_5af44d4ba8,module.savedviews.frontend_unit.workbook_saved_view_pagination_machine_lr06` | PASS 3/3 | `20260916T021936Z-p967551` |
| `make test-slice OWNER=web.application ROWS=web.application.regression.preference_integration,web.application.regression.routestate_suite_375a237e22,web.application.regression.shared_public_error_normalization_and_sanitizati_80bdbb4daf,web.application.regression.public_error_presentation_typed_mapping_1ea7e57fb4` | PASS 5/5 | `20260916T021936Z-p967576` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.saved_view_independent_reads,web.workbook.regression.saved_view_operation_owner,web.workbook.regression.activesurfacesavedviewselector_resource_action_focus_a103000001,web.workbook.regression.useworkbooksavedviewcontroller_suite_d3f1a57e5d,web.workbook.regression.workbook_saved_view_adapter_boundary_b1c2d3e4f5,web.workbook.regression.workbooksavedviewcontrol_resource_reducer_parser_a103000002,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e,web.workbook.regression.preference_owner,web.workbook.regression.preference_characterization` | PASS 10/10 | `20260916T021953Z-p972158` |
| `make frontend-typecheck` | PASS 2/2 | `20260916T022420Z-p989488` |
| `make lint-biome` | PASS 2/2 | `20260916T022420Z-p989540` |
| `make lint-scripts` | PASS 2/2 | `20260916T022420Z-p989550` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260916T022420Z-p989514` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260916T021953Z-p971976` |
| `make openapi-compatibility-check` | PASS 4/4 | `20260916T021953Z-p972016` |
| `make backend-module-boundary-check` | PASS 3/3 | `20260916T021953Z-p972295` |
| `make migration-drift` | PASS 5/5 | `20260916T021953Z-p972032` |
| `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.production_ddl_v2_recurrence,module.database_migrations.unit.public_surface_and_reader_capability` | PASS 1/1 | `20260916T022420Z-p989329` |
| `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19,web.architecture.boundary_support.transportboundarypolicy_suite_aab5aaaedb,web.architecture.boundary_support.workbook_layout_policy_suite_5f8fee9a2f` | PASS 4/4 | `20260916T022420Z-p989358` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.saved_view_independent_reads,web.workbook.regression.saved_view_operation_owner,web.workbook.regression.activesurfacesavedviewselector_resource_action_focus_a103000001` | PASS 4/4 | `20260916T022420Z-p989382` |
| `make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.integration.independent_read_contract` | PASS 3/3 | `20260916T020755Z-p792788` |
| `make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.browser.independent_discovery_continuity,module.savedviews.accessibility.independent_discovery_continuity` | PASS 13/13 | `20260916T021515Z-p919230` |
| `make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | PASS 11/11 | `20260916T021327Z-p881427` |
| `make browser-e2e-visual-update` | PASS 12/12 | `20260916T020831Z-p810607` |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.spreadsheet_creation,module.timeline.browser.clipboard_fidelity,module.timeline.browser_support.timeline_grid_keyboard_navigation_edit_cancellat_6a838bf56c,module.timeline.browser_support.timeline_keyboard_fill_down_preserves_the_select_66fb142159` | PASS 13/13 | `20260916T014704Z-p361107` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.integration.startup_selection_fallback_behaves_correctly_acr_081544c048,module.workbook.support_integration.workbook_startup_preferences_bootstrap_and_upser_6ac935eac8,module.workbook.browser.grid_autosave_availability` | PASS 13/13 | `20260916T014927Z-p402924` |
| `make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.browser_stateful.browser_saved_view_query_layout_state_user_home_cb9c681674` | PASS 11/11 | `20260916T021020Z-p848982` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.verify_default_closed_inspector_state_no_row_sta_897d604d9e` | PASS 11/11 | `20260916T020651Z-p761027` |
| `make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.browser.authoring_recovery_5cd20f1b328e,module.savedviews.accessibility.independent_discovery_continuity,module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | PASS 15/15 | `20260916T020049Z-p631272` |

The authoring version/no-op, delayed-receipt and public collection rows passed
individually in `20260916T015802Z-p556238`; that aggregate run failed its older
uncertain-create interception and is not claimed as an overall PASS. Its fixed
uncertain-create row is in the passing repeat above. The retained SVD-04 matrix
and row artifacts supply the detailed acceptance mapping.

The broad `make test-slice OWNER=web.architecture` run at
`20260916T021936Z-p967509` remains FAIL because of the 24 baseline selector-policy
findings documented in G14. The related type cycle was fixed and both direct
import validation and the affected source/transport/layout rows pass above.
This seam has no unresolved applicable acceptance row or BLOCKED dependency.
The baseline failure is explicitly retained, not suppressed or relabeled PASS.

## Limits, skipped checks and next action

Full `make check`, release/CI gates, performance measurement and an unrestricted
whole-repository test run were not invoked: the changed boundary is covered by
selected owner slices, live service/browser checks, API compatibility, migration
and generated drift, source/import checks, typecheck and lint. The full visual
refresh was necessary because its Make-owned target has no narrower row filter.
No performance or Core 05 release-readiness claim is made.

OS screen-reader/screen-magnifier testing is N/A to the automated Chromium
acceptance boundary; accessible DOM, keyboard, live announcements, contrast,
geometry and synthetic zoom/text-spacing evidence are recorded. Unrelated
heading-selector migration is N/A to this implementation scope and remains
web.architecture maintenance. These scope statements do not exempt any changed
saved-view contract or lifecycle behavior from its passing evidence.

Remaining risks: keyset discovery reflects live concurrent writes, not a frozen
catalog; First/Refresh starts a new chain. Uncertain creation may duplicate an
unacknowledged earlier create under the unchanged owner contract. Actual assistive
technology pronunciation is not established by browser automation. Deployment
must schedule the additive index and respect backend-before-frontend ordering.
The pre-existing selector-policy failure remains a separate repository issue.

Next action: review the 179 changed paths (101 seam-affected PNGs and 78 other
paths) and this evidence, then deploy only through the normal authorized release
process with the sequencing above. No additional product implementation is
outstanding for this seam. No commit, push, deployment or analyst-data mutation
was performed. Source rollback must preserve any work added after the clean
captured baseline.

### SVD-05 terminal exit

`make lint-markdown` PASS at
`.cartulary/test-results/20260916T023112Z-p994191/adhoc/lint-markdown/tool-run-summary.json`.
`git diff --check` PASS. Final byte review verifies UTF-8, a terminal newline,
no CR bytes or trailing whitespace, balanced Markdown tables, and exact inventory
agreement with all 179 changed paths. Branch remains main at captured HEAD;
index is unstaged. Digest, dependency lockfiles and immutable OpenAPI release
history are unchanged. No pre-existing work was present or discarded.

All applicable contract/lifecycle acceptance rows have passing evidence. Related
implementation and verification defects are resolved; the separate unchanged
selector-policy baseline failure is retained explicitly. There is no outstanding
seam implementation, required verification or blocked dependency. The completed
handoff contains ownership decisions, compatibility, exact paths, command/run
records, limitations, deployment order and rollback. Next action is review; no
commit or deployment is authorized by completion of this work.
