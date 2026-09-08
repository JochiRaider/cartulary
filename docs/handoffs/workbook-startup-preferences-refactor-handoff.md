# Workbook Startup Preferences Refactor Handoff

## Baseline, authority, and scope

Execution entry: main, HEAD d4670bb956c9719f6d0e085529d47d63108616fe,
clean tree. Root AGENTS.md is the only applicable agent instruction file.
React 19, TypeScript 6, Vite 8, pnpm 10.33, Vitest/Testing Library,
Playwright, Biome and Go 1.27.1 remain the current stack.
No reset, commit, push, deployment, dependencies or stored-data migration.

Core 01 REQ-01-148–151.1 and REQ-01-261–262 own preference resources and
sheet_ref identity; Core 03 §2.4 / REQ-03-027–032 owns startup selection and
personal/default authority. Core 01 REQ-01-590 allows preferences on closed
incidents. Core 04 REQ-04-003/017/023/026/029 owns CSRF, current session,
membership and deployment-admin separation. Design §§3/7/8/12/14–15 and domain
vocabulary apply within their boundaries. No owner contradiction found.
The NLSpec research essay is background, not product authority.

The localized digest read order and stale map were revalidated during planning;
the digest remains advisory and unchanged. Completed lifecycle, metadata,
membership and view-bar handoffs are implementation evidence only.

Allowed paths: preference model/controller/port/adapter/presentation and tests;
bounded App, shared contracts, summary and workbook/view-bar integration;
module.workbook OpenAPI/projection compatibility inputs and generated outputs;
workbook/incident backend evidence; authored source ownership/test routing and
Make-generated derivatives; semantic selectors, browser/accessibility/visual
tests and reviewed goldens; this handoff. Startup fallback/repair, saved-view
CRUD, lifecycle/metadata/membership policy, account settings, storage, new routes
and digest edits are excluded. Generated roots and files follow
tools/generated_artifact_policy.json and are never hand-edited.

## Tracker

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| WP-01 — Baseline, authority, characterization | DONE | Six frontend defects and missing error projection reproduced. |
| WP-02 — Typed transport and operation ownership | DONE | Twenty-two focused controller/transport cases and response projection pass. |
| WP-03 — Shared UI and old-path retirement | DONE | Both presentations delegate; focused integration and typecheck pass. |
| WP-04 — Service/browser/accessibility/visual evidence | DONE | Required journeys pass; 17 images reviewed and two fresh full visual passes. |
| WP-05 — Final verification and completed handoff | DONE | Final gates passed; completed handoff and post-completion audit. |

Only the current row is IN_PROGRESS. Record paths, commands, results, decisions,
risks and next action at every DONE/BLOCKED exit. Never cross a blocked dependency.

## Baseline reader/writer map and planning evidence

Baseline writers: WorkbookShell infrastructure creates the preference adapter;
useWorkbookShellRuntime passes it to useWorkbookStartupController; view-bar
binding delegates through ActiveSurfaceSavedViewSelector and the saved-view
action dispatcher. IncidentAdminPanel separately reads both preference routes
and incident summary with Promise.all. Startup admission consumes startup identity
and availability through its existing port and must retain that ownership.

Target ownership: module.workbook owns preference contracts/API evidence;
module.incidents owns incident-summary retrieval; web.workbook owns the narrow
preference workflow; web.application retains it for current incident/session,
authorization and departure. package.protocol_ts, package.ui and web.design
retain their existing contract, selector and readiness ownership.

Planning public task guides passed for those seven owners. Baseline narrow
commands passed under .cartulary/test-results/:

- web.workbook startup/adapter test-slice: 3/3 units,
  20260908T160052Z-p4173288.
- web.application summary/preferences test-slice: 4/4 units,
  20260908T160052Z-p4173289.
- module.workbook service-backed preference/fallback/upsert/extension/repair
  test-slice: 4/4 units, 20260908T160404Z-p4175503.
- git diff --check passed and execution entry remains clean.

These are existing baseline checks, not reproductions of the new race matrix.
Tests, generators and runtime evidence must not depend on Markdown.

## Decisions

One memory-only workflow per current incident/session, with independent resources,
reads, synchronous write admission and recovery. Thirty-second bounded observation;
transport settlement is separate from its deadline. Preserve captured targets across
surface changes and panel closure. Confirmed PUT remains confirmed after read failure.
User chose captured-target versus current-value review for uncertain writes: after
transport settlement and fresh observation, explicitly write the captured target
again or keep the observed value. No automatic replay or concurrency-token invention.

## Execution log

WP-01 entry: baseline revalidated; production source unchanged. Next action:
add and run routed characterization, retaining reproduced defects separately
from source risks and correct existing behavior.

### WP-01 exit

Added workbookPreferenceCharacterization.test.tsx using the actual startup,
saved-view shortcut and summary seams, plus the workbook-owned response projection
assertion in openapi_contract_test.go. Authored routing and source ownership
updated; make generate passed at 20260908T161415Z-p1338 and
20260908T161553Z-p5968.

CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=web.workbook
ROWS=web.workbook.regression.preference_characterization intentionally failed
at 20260908T161514Z-p5277: six failures, two passes. Reproduced absent clear
controls, summary/home withheld by delayed default, wrong incident/user admission,
two writes after surface switch, lost acknowledgement feedback after switch, and
stale inspected value after Set. Captured request identity itself was correct.
Confirmed-write feedback survived a later failed summary refresh; loading and
missing data were not rendered as null. The first run at 161434-p4496 had a
shared Response-body fixture error; clone-per-request corrected that fixture.
These are component/transport reproductions, not browser or security claims.

make test-slice OWNER=module.workbook
ROWS=module.workbook.support_unit.preference_response_projection failed at
20260908T161611Z-p9213 on existing 400/401/404/500 GET and PUT errors and PUT
403 responses omitted from the machine projection. Production remained unchanged.
Next: WP-02 typed resources, independent controller and narrow projection repair.

### WP-02 exit

Added typed preference model/controller and independent GET/PUT transport. Both
resources retain complete correlated resources, explicit null and exact captured
identities. Per-resource locks survive panel/surface changes and transport deadlines;
reads are fenced and bounded. Confirmation is separate from follow-up observation;
uncertainty permits only explicit observed-value review and a new authorized write.

Added 17 controller and five adapter cases and dedicated routing. Preference adapter
cases moved out of the mixed startup/incident adapter file. The workbook response
projection now includes only the 18 previously omitted existing error responses.
make generate initially required compatibility accounting at 20260908T162654Z-p11995;
all 18 additions were reviewed as additive preference response projections and
recorded in the existing change set. Generation passed at 20260908T162746Z-p13539.

- make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.preference_owner,web.workbook.regression.preference_transport:
  PASS 3/3 units at 20260908T162808Z-p16680.
- make test-slice OWNER=module.workbook
  ROWS=module.workbook.support_unit.preference_response_projection:
  PASS 1/1 at 20260908T162808Z-p16686.

No blocked dependency. Next: WP-03 App retention, both presentations, summary
separation and retirement of startup/saved-view preference ownership.

### WP-03 exit

App now retains the workflow, binds canonical workbook identity and current-session
recovery, and joins the existing departure review after lifecycle. Startup and
saved-view CRUD no longer own preference commands. Summary retrieval has its own
typed incidents GET, independent of both preference reads, and publishes through
the neutral incident-resource owner. Missing role is unresolved, not viewer.
Summary offers two independent inspect/set/clear/refresh/recovery rows. View-bar
shortcuts close with existing focus return and retain local preference feedback;
only the workbook-lifetime announcer publishes preference live updates.

Reworked existing component fixtures for complete resources and added three App
integration cases. Initial type/format failures were retired interface fixtures,
missing fixture envelope types, and new lint errors; corrected. Existing
nonblocking useTemplate informational diagnostics remain unrelated.

- make test-slice OWNER=web.application with preference_integration, the three
  summary rows, and metadata_departure: PASS 6/6 at 20260908T164439Z-p32514.
- make test-slice OWNER=web.workbook characterization and shell surfaces:
  PASS 3/3 at 20260908T164624Z-p40080. All eight characterization cases now pass.
  Earlier 164439-p32522 also passed selector focus and startup controller rows;
  shell fixture then needed the preserved shortcut menu-close behavior.
- make frontend-typecheck: PASS 2/2 at 20260908T164623Z-p39721.
- make frontend-import-boundary-check: PASS 2/2 at 20260908T164439Z-p32656.
- make format: PASS 2/2 at 20260908T164614Z-p35448.
- make generate: PASS at 20260908T164346Z-p24825.

No blocked dependency. Next: WP-04 service-backed clear/fallback/authorization,
transport timing, accessibility, visual review and adjacent regressions.

### WP-04 progress

Real service-backed journeys pass for exact base/saved identities, repeat/no-op,
both explicit clears, future-open fallback, unchanged URL/query configuration,
independent reads, cross-entry admission, changed surfaces, committed lost responses,
explicit matching/divergent recovery, closed-incident viewer isolation and admin
bounds. Claimed Network Analysis set/clear uses the declared extension identity.
Backend preference evidence now also checks closed set/clear/no-op and nonmember
GET/PUT denial even for deployment admins.

- module.workbook service-backed-test-slice: preferences_1–4,
  browser_stateful.preferences_extension and the existing preference pointer
  catalog row: PASS 14/14 at 20260908T165415Z-p51206.
- module.workbook accessibility.preferences: PASS 11/11 at
  20260908T165601Z-p138557.
- module.incidents lifecycle_departure, lifecycle_workbook, metadata_live,
  membership_management_self_actions and creation/summary composition:
  PASS 17/17 at 20260908T170023Z-p267280.
- module.savedviews CRUD, persistence/replay and query accessibility initially
  failed at 20260908T170023Z-p267270: fixed-width preference feedback overlapped
  Sort and obscured the action trigger. Moved feedback into the existing bounded
  allocation and preserved compact status presentation. PASS 15/15 at
  20260908T170601Z-p438421.
- module.timeline ArrowDown selection and Enter-focus measurement slice:
  PASS 16/16 at 20260908T170152Z-p368533.
- frontend-unit: 490/491 at 20260908T165416Z-p51483. Only new ownership-map
  duplicate entries failed; corrected, and the exact source-ownership row passed
  at 20260908T165828Z-p253190. Full final unit gate remains pending.
- Added and passed preflight-failure/read-interruption recovery; controller now
  exposes read retry after that rejected unsent attempt. Explicit Keep has
  persistent local reviewed feedback without claiming a receipt or cancellation.
- Focused controller/adapter: PASS 3/3 at 20260908T165939Z-p265901.
  Typecheck: PASS at 20260908T170605Z-p445405.

First ordinary visual run 20260908T165748Z-p207281: 33/37 default scenarios and
claimed Network Analysis passed. Failures: the now-corrected view-bar overlap;
expected lifecycle pixels where the replaced preferences section enters the
viewport; four intentionally new missing preference captures; and the theme
scenario's document reload returning HTTP 404 (trace has no subsequent startup
request). The last failure occurred during concurrent builds; it is an observation,
not an established product defect. Serial ordinary rerun will resolve it.
Reconciliation retained: 185 capture intents, 204 committed PNGs, 181 active,
23 unconsumed in this failed attempt, four missing, zero ambiguous mappings,
26 registered fixtures, none unresolved. No goldens changed or deleted.
New preference capture images and compact/comfortable density views require review
before public update; active absent captures are deliberate new evidence, not
permission to delete existing goldens.

Next: complete ordinary visual reconciliation, review candidate pixels, public
golden update and two fresh ordinary passes, then WP-05 finalization.

Serial ordinary runs at 20260908T170801Z-p499916 and
20260908T172246Z-p779217 resolved the transient theme failure. The latter completed
all functional assertions: 34/37 default scenarios and 1/1 claimed Network Analysis
passed outright; three scenarios failed only on screenshot comparison or deliberately
new captures. Reconciliation: 210 intents, all 204 committed goldens active, zero
orphans, six new missing captures, zero ambiguous mappings, 26 registered fixtures,
none unresolved. Every new capture was inspected from retained trace attachments.

Final presentation decision supersedes the intermediate WP-03 menu-close behavior:
preference Set keeps the existing action menu open with persistent per-resource
outcomes. Escape returns focus to its trigger. Inspect and recover opens Summary
with the same return target. Saved-view CRUD retains its existing close behavior.
This avoids crowding the finite view-bar allocation, which visual review found in
both earlier inline-feedback arrangements. No asynchronous result moves focus.
The two obsolete fixture assumptions (automatic close and a status title tooltip)
were replaced with visible outcome and focused-control assertions. The narrow
accessibility fixture restores the desktop viewport before accessing desktop chrome.

- Final saved-view query accessibility: PASS 11/11 at 20260908T171844Z-p674156.
- Final preference accessibility, single announcer and Inspect focus return:
  PASS 11/11 at 20260908T172112Z-p730076.
- Preference persistence/timing journeys and claimed extension journey passed at
  20260908T171722Z-p621852; only the then-unfixed narrow accessibility fixture
  failed that combined run (subsequently passed above).
- App integration now includes wrong-incident/actor workbook publications; four
  cases pass at 20260908T171121Z-p555096. App also fences the binding callback by
  exact session lifetime. Already-authorized saved-view labels are supplied from
  the current workbook list without additional retrieval or resource caching.
- Final characterization (nine cases), controller (18 cases), shell surfaces:
  PASS 4/4 units at 20260908T172604Z-p835145.
- package.ui owner slice: PASS 10/10 at 20260908T172247Z-p780672.
- Import boundary: PASS 2/2 at 20260908T172252Z-p795241.
- Typecheck: PASS 2/2 at 20260908T172552Z-p834499.
- Format: PASS 2/2 at 20260908T172547Z-p830276.
- Generation after final routing update: PASS at 20260908T173224Z-p837193.

Golden acceptance trigger: the old summary/preference and saved-view feedback
pixels are stale against the now-validated cohesive workflow. Six new preference
captures were explicitly requested and are created through the public candidate
update; the ordinary run's deliberate absent captures do not authorize deleting or
re-accounting any existing golden. No existing viewport, zoom, mask, scroll
normalization, crop, browser/font pin or fixture registry changed. New captures
use the existing full-viewport helper and pinned renderer, including explicit
compact/comfortable geometry assertions and narrow recovery reachability.
Affected authored rows: web.design.visual.lifecycle;
module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc;
module.workbook.visual.preferences. These changed captures are nonregistry intents
with exact catalog/scenario/project/path mappings; no registered D-VFIX identity
is replaced. Public update and post-update inspection remain in progress.

Public make browser-e2e-visual-update passed 12/12 units at
20260908T173318Z-p840516. Its reconciliation passed: 210/210 active goldens,
zero missing/orphan/ambiguous, all 26 registered fixtures resolved. Post-update
review inspected every accepted changed PNG individually: full-width preference
rows, read/operation separation, retained focus, readable recovery wrapping and
density passed. Lifecycle controls remain reachable; their changed pixels come
from the replaced summary preference section and existing scroll positioning.

The full update also proposed 15 incidental changes outside this accepted seam
(most were 1–23 anti-alias pixels; the appearance-zoom image had low-level raster
variation with the same inspected content/geometry). Those candidate replacements
were declined by restoring only this run's unrelated PNG outputs to their baseline
bytes. No pre-existing user work was involved. make generate then refreshed the
manifest from the accepted set (PASS at 20260908T173838Z-p899077). No golden was
hand-created, edited, cropped, deleted or re-accounted; all accepted new bytes came
from the public update. First post-update ordinary run 20260908T173910Z-p902292 completed 36/37
default scenarios and 1/1 claimed extension; only lifecycle-compact differed by
a six-pixel scroll position. The broad web.design slice at
20260908T173958Z-p967351 passed its seven browser and 12 accessibility scenarios;
its visual selection reproduced the same sole mismatch (17/19 scheduler units).
That slice unexpectedly included browser builds, so further visual runs are serial.
All 210 captures remained reconciled, with no missing or ambiguous mappings.

Inspection found that lifecycle's density fixture started focus/scroll before the
now-independent summary/home/default reads settled. Added explicit observation
readiness before fixture interaction, without changing product scroll behavior,
viewport, crop or masks. This repairs the demonstrated incomplete visual fixture;
its ordinary validation precedes the two required fresh full visual passes.
The focused lifecycle run 20260908T174342Z-p1030750 confirmed the corrected
compact geometry exactly matches the accepted golden. It failed only on the
pre-existing full-suite actor-label fixture dependency (the focused run shows
Playwright instead of the full suite's AnalystAna; inspected difference bounds
for compact were x=464–570, y=19–31). No goldens were changed for that unrelated
identity. Full ordinary validation retains the suite's established fixture order. Format and generation passed at
20260908T174315Z-p1023471 and 20260908T174319Z-p1027663.

Accepted changed filenames under apps/web/e2e/workbook.visual.spec.ts-snapshots:

- lifecycle-comfortable-linux.png
- lifecycle-compact-linux.png
- lifecycle-confirmed-refresh-failure-linux.png
- lifecycle-pending-linux.png
- lifecycle-review-linux.png
- lifecycle-review-narrow-linux.png
- lifecycle-review-spacing-linux.png
- lifecycle-review-zoom-linux.png
- lifecycle-uncertain-linux.png
- workbook-preferences-comfortable-linux.png
- workbook-preferences-compact-linux.png
- workbook-preferences-confirmed-stale-linux.png
- workbook-preferences-uncertain-linux.png
- workbook-preferences-uncertain-narrow-linux.png
- workbook-preferences-unset-linux.png
- workbook-view-bar-saved-view-actions-linux.png
- workbook-view-bar-saved-view-clean-linux.png

## Final owner-to-change map

| Adopted owner / implementation owner | Resulting change and evidence paths |
| --- | --- |
| Core 01 REQ-01-148–151.1, 261–262; Core 03 REQ-03-027–032 / module.workbook | Complete nullable resource GET/PUT support; exact incident, actor and pointer validation in WorkbookPreferencePort, createWorkbookPreferenceAdapter and workbookPreferenceModel. Existing startup fallback/repair algorithm remains unchanged. Preference response projection and backend authorization/no-op evidence are workbook-owned. |
| Core 04 session, CSRF, incident membership / web.application | useWorkbookPreferences binds AppSessionController and existing authorization recovery; App retains and retires the owner and joins appDepartureReview. WorkbookPreferenceDepartureDialog reuses AccountDialog. Integration tests cover panel closure, role loss, same-actor session replacement and wrong workbook publications. |
| Core 01 / module.incidents | incidentResourceClient retrieves the typed incident summary. IncidentAdminPanel publishes through the existing neutral IncidentResourceController callback. Preference reads no longer delay incident observation or lifecycle controls. |
| Core 03 startup and saved-view separation / web.workbook | WorkbookPreferenceController owns independent resource reads, admissions, immutable attempts, settlement locks, acknowledgement and observation. WorkbookShell supplies canonical identity and already-authorized labels. Startup hooks retain selection/URL/focus; saved-view dispatcher retains CRUD. |
| Design, keyboard, focus and vocabulary / web.workbook, package.ui, web.design | WorkbookPreferencesPanel, saved-view menu delegation and the single workbook announcer provide complete accessible inspection, set/clear and recovery. Semantic selectors live in ui-contracts. Existing token/density/overflow owners are reused. Browser, accessibility, measurement and reviewed visual evidence remain implementation support. |
| Existing API projections / package.protocol_ts | Four existing generated operations are reused with normal browser CSRF. Eighteen existing error responses added to the workbook authored OpenAPI and additive compatibility change set; public generation updates assembled OpenAPI and Go projection artifacts. No new route, payload field, concurrency token or public runtime behavior. |
| Verification owners | tools/test_families/module.workbook.json, web.workbook.json, web.application.json and package.ui.json route new evidence; frontend_source_ownership.json assigns source paths. Public generation updates browser batch and topology render accounting. Tests and generators do not depend on this handoff or other Markdown. |

Final resource readers/writers: only createWorkbookPreferenceAdapter transports
preference GET/PUT for the retained WorkbookPreferenceController. Summary rows and
view-bar Set commands share that controller. The backend startup route retains
its pre-existing fallback and pointer-repair behavior. No display invokes it.

## Digest acceptance applicability

The localized digest remains advisory. A001–A006, A008–A009, A011,
A016–A017, A019–A020 and A022–A027 apply to the changed seam: exact owner map,
bounded scope, verified baseline, existing tokens/dark_graphite/density, reachable
responsive controls, continuity, truthful asynchronous states, keyboard and one
live announcer, deterministic component states, reviewed captures, semantic
selectors, Markdown-independent verification, generated boundaries and this record.
A006 uses existing shared density plus computed preference padding/font assertions;
this work does not redefine grid geometry. A008 reuses current responsive chrome
rather than adding viewport parsing or breakpoints.

A007, A010, A014–A015, A018 and A021 are preserved adjacent creation, inspector,
editing/conflict, evidence and virtualization behavior; their owners are not
redesigned. Existing visual/keyboard/measurement/saved-view regression evidence
covers relevant continuity. A012–A013 do not impose transaction IDs or queue replay
on these preference APIs: their adopted contract has neither. Explicit uncertain
preference recovery is a new authorized write, never a transaction replay.

## Compatibility, limits and rollback

Compatibility is complete frontend support for existing APIs and stored values.
No schema/storage migration, dependency, new field/route, write authorization rule,
startup policy or deployment change is introduced. created_at, updated_at and
default attribution are retained as resource data, never used as versions. A
no-op acknowledgement may retain another actor's attribution. Inspection shows
the stored pointer, not an independently resolved next startup choice.

The workflow is bounded and memory-only. A 30-second deadline limits observation,
not server execution; an ignored abort retains its settlement lock. An uncertain
attempt remains unresolved after a matching GET and requires explicit review;
no automatic replay or indefinite observation loop is present. Ordinary App
navigation reviews unsettled work, and deliberate departure forgets local recovery
without canceling or undoing a server write. Already-authorized metadata improves
labels; unavailable labels retain precise typed identity. No target browser or
cross-surface saved-view fetch is added.

Rollback must restore the coherent App retention/departure binding, workbook
preference port/controller/presentation, summary separation, old-path wiring,
projection compatibility inputs, generated derivatives, selectors and routed test/
visual evidence together. Do not roll back users' legitimately saved preference
values. No commits, pushes or deployments are part of this work.

## Changed path inventory

Paths below are relative to the repository root. PNG filenames are listed in the
visual acceptance record above. Generated entries were produced by public Make
targets; authored files supply their inputs.

```text
apps/web/e2e/extensions.stateful.spec.ts
apps/web/e2e/incident-administration.spec.ts
apps/web/e2e/support/workbook/savedViews.ts
apps/web/e2e/workbook-preferences.spec.ts
apps/web/e2e/workbook.a11y.spec.ts
apps/web/e2e/workbook.visual.spec.ts
apps/web/src/app/App.tsx
apps/web/src/app/IncidentAdminPanel.test.tsx
apps/web/src/app/IncidentAdminPanel.tsx
apps/web/src/app/IncidentLifecyclePanel.tsx
apps/web/src/app/WorkbookPreferenceDepartureDialog.tsx
apps/web/src/app/api/incidentResourceClient.ts
apps/web/src/app/appDepartureReview.test.ts
apps/web/src/app/appDepartureReview.ts
apps/web/src/app/incidentLifecycleTestSurface.tsx
apps/web/src/app/useWorkbookPreferences.ts
apps/web/src/app/workbookPreferenceIntegration.test.tsx
apps/web/src/testing/workbookPreferenceTestSupport.ts
apps/web/src/workbook/WorkbookShell.surfaces.test.tsx
apps/web/src/workbook/WorkbookShell.tsx
apps/web/src/workbook/adapters/createWorkbookPreferenceAdapter.test.ts
apps/web/src/workbook/adapters/createWorkbookPreferenceAdapter.ts
apps/web/src/workbook/adapters/createWorkbookStartupAndIncidentAdapters.test.ts
apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.test.tsx
apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.tsx
apps/web/src/workbook/components/SavedViewActionPanel.tsx
apps/web/src/workbook/components/WorkbookShellViewBarControls.tsx
apps/web/src/workbook/hooks/useActiveSurfaceSavedViewActions.ts
apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts
apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts
apps/web/src/workbook/hooks/useWorkbookStartupController.test.tsx
apps/web/src/workbook/hooks/useWorkbookStartupController.ts
apps/web/src/workbook/models/workbookSavedViewControl.ts
apps/web/src/workbook/ports/WorkbookPreferencePort.ts
apps/web/src/workbook/preferences/WorkbookPreferenceController.test.ts
apps/web/src/workbook/preferences/WorkbookPreferenceController.ts
apps/web/src/workbook/preferences/WorkbookPreferencesPanel.tsx
apps/web/src/workbook/preferences/workbookPreferenceCharacterization.test.tsx
apps/web/src/workbook/preferences/workbookPreferenceModel.ts
contracts/openapi-releases/2.0.0.change-set.json
contracts/openapi-source/owners/module.workbook/openapi.json
contracts/openapi/cartulary.openapi.yaml
docs/handoffs/workbook-startup-preferences-refactor-handoff.md
internal/gen/contractopenapi/artifacts_gen.go
internal/gen/openapioperations/catalog_gen.go
internal/modules/workbook/openapi_contract_test.go
internal/modules/workbook/workbook_startup_test.go
packages/ui-contracts/src/index.ts
packages/ui-contracts/src/workbook-shell-grid-selectors.test.ts
packages/ui-contracts/src/workbookShellSelectors.ts
tools/browser_e2e_batch_manifest.json
tools/execution_topology_render_index.json
tools/frontend_source_ownership.json
tools/frontend_visual_golden_manifest.json
tools/test_families/module.workbook.json
tools/test_families/package.ui.json
tools/test_families/web.application.json
tools/test_families/web.workbook.json
```

## Remaining WP-04 validation log

The complete run 20260908T174603Z-p1079887 passed every preference, lifecycle,
density and other screenshot comparison after the readiness fix. Its sole failure
was an adjacent entity-linking functional assertion: a dismissed mention briefly
rendered as resolved then disappeared before the expected accessible name was
observed. No mention/entity production code changed. No golden update was attempted
for that failed assertion. Fresh complete runs are required to establish stability.

First fresh full ordinary visual pass: make browser-e2e-visual, 12/12 units at
20260908T174950Z-p1130336. All 37 default scenarios plus claimed Network Analysis
passed; the mention assertion did not recur. A second fresh complete pass is
running against the same accepted manifest and source.

Final narrow module.workbook service slice at 20260908T173936Z-p947562:
PASS 4/4 units (preference routes, startup fallback and bootstrap/upsert/repair).
The service-backed selector filters pure unit rows, so response projection,
nullable PUT decoding and startup admission were also run explicitly with
make test-slice OWNER=module.workbook: PASS 2/2 at
20260908T174706Z-p1128894. package.protocol_ts owner slice passed 7/7 at
20260908T173518Z-p891819. Fixture-only readiness change then passed typecheck
and Biome at 20260908T174418Z-p1076055 and 20260908T174431Z-p1078589.

### WP-04 exit

Second fresh make browser-e2e-visual passed 12/12 at
20260908T175251Z-p1180020. Together with 20260908T174950Z-p1130336, both complete
runs passed 37 default scenarios and one claimed Network Analysis scenario;
reconciliation passed all 210 active captures, zero missing/orphan/ambiguous,
all 26 registered fixtures resolved. The adjacent mention assertion did not recur.
All 17 accepted changed images were inspected; no further update was necessary.

Required preference persistence, authorization, extension identity, independent
recovery, failure transitions, accessibility, continuity and adjacent regressions
have passing evidence above. No blocked dependency or product-policy expansion.
Next: WP-05, run agent-finalize with RESULTS_DIR unset before broader final gates,
then complete the handoff and final audit.

### WP-05 progress

Before broader final verification, env -u RESULTS_DIR make agent-finalize passed
1/1 at 20260908T175637Z-p1229800. Schema shape, tier coverage and generated
structure refresh passed. Canonical retained-evidence and scheduler drift actions
were skipped with results-dir-not-provided: no qualifying successful full warm
check from this exact source exists. This is not a successful retained-run audit.
No additional source changes resulted from finalization.

Final gates are running on the completed implementation. Public command selections
and exact ROWS arguments are retained in each run-manifest.json declared_inputs;
counts below distinguish scheduler units from individual test cases. All run roots
in this handoff are relative to .cartulary/test-results/.

| Final public gate | Result | Run root |
| --- | --- | --- |
| make frontend-typecheck | PASS 2/2 units | 20260908T175725Z-p1233356 |
| make frontend-import-boundary-check | PASS 2/2 | 20260908T175800Z-p1278844 |
| make lint-biome | PASS 2/2 | 20260908T175807Z-p1286152 |
| make json-shape-check | PASS 3/3 | 20260908T175812Z-p1290148 |
| make generated-artifact-policy-check | PASS 3/3 | 20260908T175820Z-p1296967 |
| make generate-drift | PASS 4/4 | 20260908T175823Z-p1299617 |
| make openapi-compatibility-check | PASS 4/4 | 20260908T175846Z-p1311212 |
| make frontend-unit | PASS 491/491 | 20260908T175725Z-p1233352 |
| make browser-e2e-a11y | PASS 12/12 | 20260908T175725Z-p1233583 |
| Final module.workbook service-backed browser/startup/extension slice | PASS 16/16 | 20260908T180052Z-p1339433 |

The full accessibility pass contains 32 default scenarios and one claimed extension
scenario. Final workbook ROWS are preferences_1–4, preferences_extension,
browser_startup_honors_explicit_sheet_home_defau_ee3b02ee01 and
verify_continuous_workbook_shell_composition_top_96ea2f5084; full IDs and exact
command are in that run's declared_inputs. The final unit gate includes every
routed frontend test, including characterization, transport, controller, App,
summary, startup, saved-view, source ownership, selectors and adjacent owners.

Verification limits: the complete repository check/release/security/performance
corpus was not run; narrow owner evidence, complete frontend unit/accessibility/
visual gates and the listed static/generated checks cover this frontend refactor.
No backend production implementation, storage or dependency changed. Stateful
preference extension and saved-view regressions, and selection/focus measurement,
used their applicable owner slices rather than unrelated full stateful/measurement
corpora. Retained successful full-warm-run maintenance was skipped as recorded.
The isolated lifecycle visual row has a pre-existing actor-label fixture dependency;
complete visual runs supply the established fixture context and both passed. The
one observed adjacent mention failure did not recur in either final visual pass.

### WP-05 exit

All workstream exit conditions pass. The final owner-to-change map, digest
applicability, path inventory, compatibility, limitations, failures/corrections,
reviewed visual evidence and coherent rollback are retained above. Scope audit:
main and HEAD d4670bb956c9719f6d0e085529d47d63108616fe are unchanged; all dirty
paths belong to this authorized implementation/evidence/handoff. No unrelated
work was present or overwritten. No normative owner, digest, startup resolver,
backend route implementation, storage, dependencies, lockfiles, lifecycle/
metadata/membership policy or saved-view CRUD algorithm changed. Preference
commands are absent from startup and saved-view mutation ownership. Authored
verification remains independent of Markdown; public generators own every changed
generated artifact. No commit, push or deployment occurred.

WP-05 is DONE. After marking it DONE, make lint-markdown passed at
20260908T180259Z-p1392746 (the earlier documentation pass was
20260908T180126Z-p1383945). git diff --check passed. The final manual scope audit
matched all 75 changed paths against this completed handoff, found no excluded
changes, confirmed all five tracker rows DONE, and confirmed unchanged main/HEAD.
This closure record is included in the final repeated Markdown/diff/scope check.

Next action: **none**.
