# Workbook save-status refactor handoff

## Control and baseline

This is the controlling record for the authorized save-status seam. No commit
is authorized. The digest remains read-only and the completed presence,
relationship/collection, Evidence, inspector, recovery, view-bar, density and
grid-state seams remain regression boundaries.

- Branch: `main`.
- Baseline: `66161eef7f22576c1c5a6b9bad504425bb57bf76`.
- Initial dirty state: clean, reconfirmed at implementation entry.
- Instructions: root `AGENTS.md`; no nested instructions.
- Stack: existing React/TypeScript/Vite, Vitest/Testing Library, Playwright.
- Allowed paths: status-related workbook source and producer adapters under
  `apps/web/src/workbook`, directly required shell integration under
  `apps/web/src/app` or `apps/web/src/networkFlow`, focused unit/browser support
  and tests, authored UI presentation/selector projections, source-ownership
  and verification routing, Make-generated derivatives, reviewed affected
  goldens, and this handoff. Narrow design clarification only if required.
- Preserve unrelated work; no backend, API, storage, dependency or migration
  changes. No executable dependency on Markdown.

## Authority and decisions

Core 00 §§1–2 and §5.1 establish precedence. Core 01 §3.3.5 owns mutation
identity and dispatch; §3.3.10.1 REQ-01-261/262 and §7.4 own exact surface
identity. Core 03 §§3.1–3.3, §4.2 REQ-03-089, §4.3 REQ-03-095/096 and
§4.4 REQ-03-099/100/301/302 own continuity, conflicts, save state and replay.
REQ-03-287/288 govern closed-incident presentation and retained rejected
drafts. Core 04 §§1–2 own authentication and current authorization; status
presentation must not change those boundaries.

Design §§7.4–7.5, §10.1, §§10.5/10.8 and §14.2 own responsive allocation,
secondary priority, recovery activation, safe copy and transition announcements.
The adopted Testing Harness NLSpec owns execution mechanics; verification
routing is downstream evidence, not product authority. `docs/domain.md` supplies
vocabulary and owner navigation only. Completed conflict-recovery and presence
handoffs are evidence only. No unresolved owner contradiction is identified.

The digest START_HERE, LOCAL_AGENT_PROMPT, REPO_MAP, OWNER_MAP, rules and
acceptance were read during planning and revalidated against the current
checkout. Its older localization and historical handoff pointers are advisory
drift, not new product authority. Generated boundaries follow
`tools/generated_artifact_policy.json`; visual review follows the current
visual golden maintenance guide.

## Tracker

Only one row may be `IN_PROGRESS`. Record its bounded work, evidence, failure
disposition and next action before beginning its successor. A blocked dependency
prevents advancement.

| Workstream | Status | Exit and next action |
| --- | --- | --- |
| SS-01 Baseline and characterization | DONE | Four expected failures demonstrate global explanation loss, overlap undercount, unmount false Saved and independent live labels. Next: SS-02. |
| SS-02 Typed projection | DONE | Typed scope, safe global blockers, active counts and coherent actions pass pure tests. Next: SS-03 adapters. |
| SS-03 Integration | DONE | Shared projection/rendering, completion handles and transition host integrated; frontend checks pass. Next: SS-04. |
| SS-04 Browser and visual evidence | DONE | Stateful, accessibility and measurement routes pass; 13 goldens reviewed and two fresh ordinary visual passes succeed. Next: SS-05. |
| SS-05 Final validation | DONE | Finalizer, terminal frontend/browser checks, Markdown and whitespace checks pass; scope and rollback audited. No successor seam. |

## Producer inventory and characterization

| Producer or consumer | Baseline responsibility and finding |
| --- | --- |
| Pending queue model | FIFO units, queued/in-flight counts, auth pause, overflow, halt and same-field anchors; existing derivation already normalizes duplicate anchors. Preserve queue algorithms. |
| Mutation runtime | Explicit operation counter, conflict store and surface-label map; global precedence is already Conflict before Syncing. |
| Timeline save presentation | Re-derives queue state plus pending operations and refresh counters, reports labels, clears its report on unmount. Reporting lifetime requires characterization. |
| Generic mutation controller | A single mutable completion ref settles the previous operation when another starts; all rejected operations and refresh failures become local Conflict. |
| Entity operations | Operation-local completion closures already exist; local labels override the runtime and local errors bypass secondary selection. |
| Assessment creation | Runtime operation closure and stale-generation guard; preserve these while integrating presentation scope. |
| Evidence, coordination, party links, Timeline explicit actions | Consume generic begin/finish or Timeline beginSave/finishSave adapters; migrate reporting only and retain owner feedback. |
| Status projector and selector | Candidates match schema ID, so saved views collapse and global blockers disappear from other schemas; selected copy and recovery priority can diverge. |
| Status renderers | Two primary renderers each use polite live status; secondary error renderer is another live region. Quiet presence is already separate. |
| Shell recovery focus/frame | Semantic refs and return fallback exist; every terminal blocker currently takes precedence over active same-field conflicts. |
| Network Analysis | Own table-loading status; regression-only unless held Base work demonstrates missing shared-shell save visibility. |

Proven source gaps are distinguished from browser risks. Passing baseline tests
do not establish coverage of overlapping operations, saved-view scope, transition
DOM mutations or responsive clipping. Characterization results follow below.

## Verification ledger

Planning inspection: `make help`, `make help-all`, and module-author task guides
for `web.workbook`, `module.workbook`, `module.timeline`, and `web.design` passed.
Start with exact owner rows via `test-slice`/`service-backed-test-slice`.

- Status baseline: `make test-slice OWNER=module.workbook` selecting the secondary
  selector and two save-state derivation/presentation rows passed at
  `.cartulary/test-results/20260905T053817Z-p89261`.
- Runtime baseline: `make test-slice OWNER=web.workbook` selecting responsibility
  and surface-continuity rows passed at
  `.cartulary/test-results/20260905T053818Z-p89493`.
- Implementation-entry `git status --short --branch`: clean; baseline unchanged.

### SS-01 exit

Before production edits, the new authored characterization row failed all four
assertions at `.cartulary/test-results/20260905T054733Z-p93746`:
missing global transaction explanation; two generic operations counted as one;
Timeline unmount reported Saved while its explicit work remained pending; and
the visible Saved label owned an independent live region. This is expected
failure evidence, not a waived check. The preceding routing attempt rejected
unsorted catalog titles before execution; sorting corrected it.

Changed paths are the controlling handoff, the new runtime characterization
test, its authored `web.workbook` family row and source-ownership entry.
Queue normalization is already tested and will be reused. Remaining browser
combinations are explicit SS-03/SS-04 verification obligations. No owner change
is required. Next action: typed primary/secondary projection and pure tests.

### SS-02 exit

The pure projector now derives primary state through the existing queue helper,
prefers current conflict-store entries over older duplicate queue reports, and
projects explicit scopes, counts and semantic action targets. Exact saved-view
identity is presentation metadata; queue identity and payload remain unchanged.
Typed blockers use the existing safe recovery summary. Empty auth/refresh activity
does not manufacture unsaved work. Runtime/producer and rendering cutover follows
in SS-03; the three integration characterization failures remain mandatory exits.

- Runtime responsibility row PASS: `.cartulary/test-results/20260905T055212Z-p95536`.
- Secondary selection and queue derivation rows PASS:
  `.cartulary/test-results/20260905T055213Z-p95768`.
- Pure cases include priority permutations, base/two saved-view isolation, global
  eligibility, refresh and explicit work, duplicate token/version producers,
  global/local conflict counts, safe summaries and no-work authentication.

### SS-03 ongoing evidence

Integration typechecks initially failed at
`.cartulary/test-results/20260905T060047Z-p98986` and
`.cartulary/test-results/20260905T060620Z-p1595` while retiring old consumer ports
and migrating test fixtures. These are change-related failures to correct,
not waived checks. Refresh-paused copy remains informational: no refresh retry
action is invented. Authentication activation reuses the existing shell
explicit session recovery port.

The first broad frontend unit run failed 66 rows at
`.cartulary/test-results/20260905T060827Z-p5030`, primarily a Timeline render
loop caused by publishing unchanged local conflict state. Retaining unchanged
state and subscribing its queue projection to runtime updates corrected the loop
and stale authentication notice. The next run passed 440/441 rows at
`.cartulary/test-results/20260905T061448Z-p56531`; remaining assertions depended
on shell child indices and the retired classification of row-version rejection
as Conflict. Both are corrected against semantic DOM and owned local error
feedback. Typecheck and Biome passed at `20260905T061448Z-p56501` and
`20260905T061448Z-p56581`; import boundaries passed at `20260905T061320Z-p54031`
after moving cross-surface characterization outside the common runtime directory.

Additional characterization found accepted managed writes could expose Saved
during required refresh: expected failure at `20260905T060924Z-p26208`, corrected
PASS at `20260905T061247Z-p52063`. Existing Entity merge and create-related
characterization failed at `20260905T061700Z-p99739` and
`20260905T061730Z-p712` because those operations supplied no pending contribution.
The adapter audit therefore includes Entity batch paste and shared inspector
history/create-related writes, using completion closures only. An initial merge
slice used an incorrect row ID and was rejected before execution; the catalog ID
was then used. No failed check was weakened.

### SS-03 exit

`make frontend-unit` PASS, 441/441 units, at
`.cartulary/test-results/20260905T062533Z-p52900`. Corrected shell, merge,
create-related and save-status focused rows PASS at `20260905T062430Z-p50415`.
Final integration typecheck PASS at `20260905T062737Z-p95868`, import boundaries
PASS at `20260905T062508Z-p51720`, and Biome PASS at `20260905T062508Z-p51750`
with final cleanup passing Biome at `20260905T062749Z-p96442`. Local control busy state is Boolean;
only the runtime projection supplies primary labels. Indicator inspector writes
also receive a reporting-only adapter; their action and feedback owners remain
unchanged. Transition acknowledgement stays with the runtime across shell
remounts. Timeline pending notices retain their detail/count but no longer own
nested save-related live regions. Required recovery/resolver announcements remain.

### SS-04 ongoing evidence

Production browser characterization at
`.cartulary/test-results/20260905T063018Z-p97874` demonstrated that switching to
Network Analysis removed `save-state` while a Base patch remained pending. The
existing extension continuity row reached the real Network Analysis workspace
before this assertion failed. This authorizes only a shared status presentation
slot in its existing status row; extension loading remains separate. Fixture
cleanup also exposed a held-route completion/disposal race; cleanup now waits
for that response before removing interception. No replacement markup is injected.

The new browser row observes actual live-region DOM mutations and covers saved
views, cross-surface pending work, same-field recovery and a global FIFO blocker
combined with a held local create. Its first routing attempt rejected unsorted
scenario IDs before execution; the authored IDs were sorted as required.

Further SS-04 characterization and corrections:

- `20260905T063508Z-p95002`: global FIFO/concurrent-create case PASS; saved-view
  case used a nonexistent Assessment tab and was corrected to the existing
  System views switcher. `20260905T064119Z-p42674` then demonstrated retained
  pending work after a successful detached Timeline response.
- Extension follow-ups `20260905T063340Z-p46325`, `20260905T064137Z-p82206`
  and `20260905T064255Z-p74867` retained Syncing correctly but could not settle.
  Production timing identified `Timeline accepted projection was not committed`:
  the old completion attempted a React state commit after unmount. The narrow
  coordinator guard skips detached UI effects, preserving FIFO settlement and
  accepted response handling. Added unit characterization failed at
  `20260905T064358Z-p21553`, then passed at `20260905T064437Z-p22273`.
- `20260905T064438Z-p22537`: cross-surface settlement and exact saved-view scope
  passed; explicit global resolver dismissal failed to return focus to its
  invoker. The existing shell focus owner now includes the activated resolver
  in its recovery target/return lifecycle, without reordering conflicts.
- Initial accessibility slice `20260905T064214Z-p29394` failed an obsolete
  visible-primary `role=status` assertion; assertions now inspect the separate
  host and require visible primary silence. Its shell scenario also failed
  startup with an asset 404 while a concurrent browser build rewrote shared
  assets, consistent with shared-output interference. Subsequent
  browser build/run routes are serialized; this infrastructure failure is not
  waived. Presence tests now observe actual save-host DOM silence through
  presence arrival, density reloads and responsive/zoom/text-spacing changes.

- `20260905T064722Z-p69541`: geometry/focus checks passed; final event assertion
  omitted the real Conflict → Syncing → Saved interval during resolver refresh.
  The expected sequence now includes that required work rather than altering
  runtime settlement. Browser-only helper type errors at
  `20260905T064723Z-p69837` were corrected; typecheck PASS at
  `20260905T064850Z-p16732`.
- `20260905T064849Z-p16440`: both new stateful scenarios and both existing
  Timeline public mutation/recovery scenarios PASS, including remount-safe
  retry/discard/resolver handoff. Shell accessibility PASS. The remaining
  accessibility assertion still required the retired pending-notice live region;
  it now requires quiet detail and the shared Syncing event. Corrected mutation
  accessibility PASS at `20260905T065205Z-p16574`.
- `20260905T065028Z-p66041`: collaboration accessibility, socket-revocation
  recovery and queued replay after re-authentication PASS (15/15 execution units).
  This includes presence geometry, all densities, zoom, spacing, resolver
  focus/continuity and silent shared save hosts during presence/layout updates.
- Biome at `20260905T065141Z-p15988` found only unformatted new accessibility
  assertions; Make-owned formatting corrects these. The existing presence
  renewal unit now additionally asserts unchanged mutation snapshot and no
  pending save announcement. Additional browser coverage exercises terminal
  recovery across surfaces and overflow through 65 real editor admissions.

- Expanded status browser row PASS, three scenarios, at
  `20260905T065451Z-p65658`: all transition counts, saved-view isolation,
  overlapping work, terminal blocker and full FIFO overflow activation passed.
  The overflow fixture uses the real 64-unit capacity plus one refused edit;
  no alternate status markup or queue is injected.
- Final extension continuity row PASS at `20260905T065625Z-p12700`, including
  pending Base status in Network Analysis and successful settlement after return.
- Focused presence renewal and save-status rows PASS at `20260905T065541Z-p11129`
  and `20260905T065543Z-p11376`. Typecheck PASS at `20260905T065453Z-p65930`.
  Make generation PASS at `20260905T065419Z-p62461` after the final authored
  selector additions; only generated browser batching and topology hashes changed.

- Timeline measurement PASS at `20260905T065734Z-p59022` (19/19 execution
  units): all four selected paint-qualified AC-043 typing, focus, selection and
  blank-row-creation routes. This is implementation evidence, not a new claim.
- Runtime replacement characterization failed at `20260905T070231Z-p54832`:
  a reused announcement host retained the old runtime's text. Delivered text now
  carries its runtime owner and clears when that owner changes, without creating
  an initial Saved event. This is presentation state only.

- Runtime-owned announcement replacement PASS at `20260905T070300Z-p59931`.
- Ordinary `make browser-e2e-visual` at `20260905T070155Z-p12951` failed only
  screenshot comparisons: 13 changed PNGs across five semantic rows. Functional
  assertions completed. Its `browser-e2e-visual/frontend-visual-reconciliation.json`
  accounts for 94/94 active captures/goldens, 26 registered fixtures, zero
  missing goldens, zero ambiguous mappings and zero orphans. Renderer and golden
  manifest validation pass; reconciliation's sole error records the failed
  screenshot target. No capture/viewport/zoom/mask/crop/scroll policy changed.
  Accepted refresh trigger: existing goldens are stale relative to validated
  shared status copy, narrow/compact secondary allocation and the demonstrated
  Network Analysis workbook-status slot. Images remain unmodified pending the
  Make-owned updater and individual review.

- Network Analysis accessibility and its two directly relevant grid measurement
  rows PASS at `20260905T070503Z-p61459` (14/14 execution units). The graph
  algorithm/DOM-capacity certification remains outside this status-only change.
- Global design accessibility matrix PASS at `20260905T070710Z-p11994`.

### SS-04 visual refresh and review

`make browser-e2e-visual-update` PASS at `20260905T070826Z-p57642`, with
94/94 active captures/goldens and no reconciliation errors. The updater also
rewrote two unrelated images with small raster differences:
`entity-mention-chip-states-linux.png` and
`timeline-mutation-pending-replay-status-linux.png`. Their exact baseline bytes
were restored. `make generate` PASS at `20260905T071321Z-p5359` regenerated the
golden manifest from the retained images; no generated manifest was hand-edited.

Every retained changed image below was opened and visually reviewed. Accepted
changes are safe syncing copy, narrow truncation, compact accessible-only detail
with presence visible, and workbook status in Network Analysis. Primary labels,
recovery controls, modal/inspector boundaries and existing focus cues remain
visible. One affected Network Analysis image also has 15 isolated pixels outside
the status row differing by only one or two channel levels, with no geometry or
content change; no pixels were manually edited. No unrelated golden is retained.
Viewport, zoom, theme, density, masks, scroll normalization and crop are unchanged.

Semantic row aliases (from ordinary reconciliation and the authored catalog):

- V1: `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c`.
- V2: `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc`.
- V3: `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`.
- V4: `module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf`.
- V5: `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67`.

All filenames are under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.

| Reviewed golden | Row | Registered fixture (or active nonregistry capture) |
| --- | --- | --- |
| `collaboration-grid-conflict-resolver-compact-linux.png` | V1 | Active nonregistry capture |
| `collaboration-grid-syncing-strip-linux.png` | V2 | `visual.fixture.save_state_strip` |
| `network-flow-analysis-accepted-inspector-linux.png` | V3 | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-compact-saved-graphs-linux.png` | V3 | `visual.fixture.claimed_network_analysis_compact_workspace` |
| `network-flow-analysis-delete-dialog-linux.png` | V3 | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-graph-contributors-linux.png` | V3 | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-mapping-dialog-linux.png` | V3 | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-narrow-query-controls-linux.png` | V3 | `visual.fixture.claimed_network_analysis_narrow_workspace` |
| `network-flow-analysis-rejected-diagnostics-linux.png` | V3 | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-saved-graph-result-linux.png` | V3 | `visual.fixture.claimed_network_analysis_workspace_states` |
| `timeline-grid-syncing-strip-linux.png` | V4 | Active nonregistry capture |
| `timeline-mutation-transaction-recovery-panel-compact-linux.png` | V5 | Active nonregistry capture |
| `timeline-mutation-transaction-recovery-panel-narrow-linux.png` | V5 | Active nonregistry capture |

First fresh ordinary `make browser-e2e-visual` PASS at
`20260905T071423Z-p8631` (12/12 execution units, all 94 captures). Second fresh ordinary `make browser-e2e-visual` PASS at
`20260905T071702Z-p56391` (12/12 execution units, all 94 captures). Network Analysis retains the completed
presence seam's header-only allocation; its narrow shared slot carries save
status and does not reopen extension presence behavior.

### SS-04 exit

All bounded browser, accessibility, measurement and visual exits pass. The
changes demonstrated by browser failures were repaired and rerun; no failed
functional or visual assertion was waived. All 13 retained changed PNGs were
individually reviewed, with no unrelated image retained. The two final ordinary
visual runs reconcile 94 active captures, all registered fixtures and the final
manifest. Next action: SS-05 finalizer, terminal checks and atomic scope audit.

### SS-05 ongoing evidence

`make agent-finalize` PASS at `20260905T072223Z-p4136` before terminal
verification. Its `unit-artifacts/finalize-summary.json` records successful
JSON-shape, catalog/tier coverage and generated-structure drift checks, with
zero updated files. `RESULTS_DIR` was unset: canonical retained-run evidence,
scheduler event/timing drift and retained performance maintenance were skipped
because no qualifying successful full warm `check` evidence exists. These are
reported skips, not passes. Frontend and final production status browser checks
follow this finalizer.

Terminal frontend verification after finalization:

| Command | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make frontend-typecheck` | PASS | `20260905T072247Z-p7567` |
| `make frontend-unit` | PASS, 441/441 execution units | `20260905T072247Z-p7583` |
| `make frontend-import-boundary-check` | PASS | `20260905T072247Z-p7600` |
| `make lint-biome` | PASS | `20260905T072247Z-p7637` |
| `make generated-artifact-policy-check` | PASS | `20260905T072325Z-p50225` |
| `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` | PASS, `check=pass` | Console diagnostic; no run root emitted. |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.workbook_save_status_preserves_authoritative_tra_0fa8b0b16c` | PASS, all three production status scenarios; 11/11 execution units | `20260905T072248Z-p8653` |
| `make lint-markdown` | PASS | `20260905T072625Z-p95655` |
| `git diff --check` | PASS | Terminal scope audit; no whitespace errors. |

### SS-05 exit

All required seam checks passed. The final browser rerun confirms the completed
production transition, scoped recovery and overflow scenarios after finalization.
Both fresh ordinary visual passes remain valid: only this handoff changed after
those passes. The final audit includes all tracked and untracked seam paths,
confirms the baseline HEAD and an empty index, and finds no unrelated changes.
The closed handoff is read back and checked again with Markdown lint and Git
whitespace validation after this closure update; that final validation result is
reported with the delivery. No further implementation or seam is started.

## Digest dispositions

| IDs | Disposition in this seam | Evidence or boundary |
| --- | --- | --- |
| R001–R015 | ADOPT | Shared semantic status, explicit recovery, local error loci, quiet presence, owned tokens and navigation; focused unit/stateful/accessibility routes. |
| R016, R019 | ADAPT | Retain current desktop targets and typography; no touch profile or row-size inflation. |
| R017–R018, R035 | ADAPT | Current viewport bands, protected primary, full accessible secondary, internal grid scrolling, zoom and text-spacing checks. |
| R020 | ADAPT | Loading remains with its existing owners; no fake rows or save contribution from loading. |
| R021 | ADAPT | Retain existing motion tokens; reduced-motion status/recovery remains usable. |
| R022–R023 | ADAPT | Existing primary activation and semantic markers; no Save button, icon registry or dependency. |
| R024–R025 | ADAPT | Editor/inspector feedback and current recovery controls retain their existing owners and confirmation semantics. |
| R026–R034 | REJECT | Preserve all rejected cybersecurity prescriptions: no cyberpunk palette, decorative motion, dashboard/marketing redesign, generated visual authority or incidental runtime selectors. |

A001–A005 and A024–A027 map to authority, token, scope, generated-policy and
handoff audits. A008–A009/A011–A017/A019–A023 map to the status, recovery,
continuity, accessibility, measurement and visual routes in this ledger.
A006/A007/A010/A018 remain density, creation, inspector and Evidence regression
boundaries: reporting adapters change no authority, payload or owned workflow.
This record claims only the affected seam and its regression evidence, not
specification completeness or new conformance/release authority.

## Changed-file responsibilities

| Paths | Responsibility |
| --- | --- |
| `runtime/workbookMutationStatusProjector.ts`, `utils/workbookStatusSecondary.ts` | Pure typed primary/secondary projection, exact SheetRef eligibility, conflict normalization, safe copy and coherent action. |
| `runtime/WorkbookMutationRuntime.ts`, `runtime/useWorkbookMutationRuntime.ts` | Existing mutation counters and queue authority, scoped refresh reporting, transition identity/acknowledgement, active-surface adapter. |
| `runtime/WorkbookConflictStore.ts`, `runtime/workbookConflictModel.ts`, `runtime/WorkbookManagedPatchDriver.ts`, `utils/workbookPendingQueue.ts` | Capture/copy presentation origin and retain existing conflict anchors; required accepted-write refresh reporting. No identity or queue algorithm change. |
| `components/WorkbookStatusStrip.tsx`, `components/WorkbookSaveAnnouncements.tsx` | One visible renderer and one shell announcement host; Unicode truncation/full accessible detail and protected primary. Quiet presence retained. |
| `WorkbookShell.tsx`, `components/WorkbookActiveSurface*.tsx`, `hooks/useWorkbookRecoveryFocus.ts`, `components/WorkbookSameFieldConflictResolver.tsx` | Shared slot, coherent recovery-layer/focus dispatch, stable resolver activation and invoker return. |
| `components/{Generic,Entity,Assessment}WorkbookSurface.tsx`, `surfaces/WorkbookSurfacesFacade.tsx`, `timeline/presentation/*`, `timeline/models/timelineWorkbookSurfaceRuntime.ts` | Small surface adapters carry exact sheet context and shared presentation; local labels/raw-error priority bypasses removed. |
| `hooks/useGenericSurfaceMutationController.ts`, `features/{generic,entities,assessments,coordination,evidence,parties,indicators}/*`, `inspector/*` | Per-operation reporting handles through existing actions/refresh/stale completion; local busy and owned feedback remain separate. |
| `timeline/{hooks,mutations,composition,bulk,models}/*` | Replace label-based finish reporting, retain queued-work lifetimes and origin context; skip detached UI projection effects after acceptance. Existing dispatch, replay and transaction semantics retained. |
| `apps/web/src/networkFlow/NetworkAnalysisWorkspace.tsx` | Narrow shared workbook-status slot in the existing row, following demonstrated Base-status loss on navigation. Extension activity remains separate. |
| Affected unit tests; `apps/web/e2e/workbook-save-status.spec.ts`, `apps/web/e2e/support/workbook/saveStatus.ts`, `apps/web/e2e/{workbook.a11y,extensions.stateful}.spec.ts` | Characterization and production UI evidence, actual announcement DOM observation, safe scope/action/focus, responsive geometry and continuity. |
| `tools/test_families/{module.timeline,module.workbook,web.workbook}.json`, `tools/frontend_source_ownership.json` | Authored semantic selectors and ownership for new source/tests. |
| `tools/browser_e2e_batch_manifest.json`, `tools/execution_topology_render_index.json` | Make-generated routing derivatives; no manual edits. |

Workbook-relative paths above are rooted at `apps/web/src/workbook`. Exact final
source/golden inventory and reviewed visual changes are recorded at closure.

## Compatibility and rollback

Save-status projection, interaction and accessibility corrections only. No
mutation protocol, recovery algorithm, API, storage, dependency or data migration
change. Queue/coalescing/transaction identity, payloads, dispatch and replay
policies remain with their existing owners. The detached-completion guard skips
unmounted UI effects so the existing accepted-response settlement can finish.

The rollback unit is the complete reviewed seam diff against
`66161eef7f22576c1c5a6b9bad504425bb57bf76`, including source/producer interfaces,
projection and scope metadata, tests, authored routing/ownership, generated
browser topology, the generated visual manifest and all 13 reviewed goldens.
Restore those existing paths together and remove the five new files listed below
when reversing the entire seam. Do not roll back only the renderer, a producer
adapter, or the golden manifest: these private interfaces are coupled. Protect
any later unrelated edits before restoring individual audited paths. Re-run Make
generation/drift, frontend checks and the affected browser/visual routes against
the restored source. No database rollback or data migration is involved.

New files in the atomic unit:

- `apps/web/src/workbook/components/WorkbookSaveAnnouncements.tsx`
- `apps/web/src/workbook/workbookSaveStatus.test.tsx`
- `apps/web/e2e/workbook-save-status.spec.ts`
- `apps/web/e2e/support/workbook/saveStatus.ts`
- `docs/handoffs/workbook-save-status-refactor-handoff.md`

The current scope audit contains 102 paths: 86 authored files (including this
handoff), three generated metadata files and 13 goldens. No path lies outside
the allowed seam. The digest, prior handoffs, adopted owners, backend, contracts,
packages, APIs, SQL, storage and dependency manifests are unchanged. Branch and
commit remain at the recorded baseline; nothing is staged or committed.

Full backend, release and full warm `make check` suites were not run because
this is a frontend presentation seam; retained-run maintenance is explicitly
skipped as recorded above. No required seam check is waived. SS-01 through SS-05
are complete. No commit or subsequent seam is authorized.
