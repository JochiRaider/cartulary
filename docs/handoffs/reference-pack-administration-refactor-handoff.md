# Reference Pack Administration Refactor Handoff

Reference Pack administration now has one application-lifetime owner for catalog,
selection, captured operations and known-job recovery. RP-01 through RP-05 are DONE. Production,
contract, backend/browser, accessibility, visual and final documentation/scope
checks are complete, with retained-run maintenance skips recorded below.

## Baseline, authority, and scope

Execution starts on `main`, HEAD `d115e800414fc8663fdcabf07a151796087249fd`,
with a clean working tree. The same state was rechecked after planning. Root
AGENTS.md is the only applicable instruction file. No reset, commit, push, or
deployment is authorized by this seam.

Core 01 §§3.3.7, 3.3.9, 11, 17.1.1, and 17.4 own pagination, jobs, pack
lifecycle, upload admission, queries, mutations, and replay. REQ-01-610/610A
own the browser query generation contract. Core 04 REQ-04-023/028/126 own
current authorization and administrative retirement. Design §§4.5, 12, and 14
own composition, feedback, forms, and keyboard behavior. Domain supplies
vocabulary/navigation. Current precedence, the digest's localized read order,
owner maps, rules, acceptance, and completed Incident import and Account
handoffs were inspected. The Reference Pack NLSpec remains draft; research and
the digest supply no product authority. No adopted-owner contradiction found.

Allowed changes: Reference Pack presentation, pure state/executor/binding and
service adapter; necessary App activation/lifetime integration; narrow authored
OpenAPI/protocol generation corrections; focused tests, source ownership,
verification routing, reviewed visual evidence, and this record. Backend
verification/storage/activation, pack formats, migrations, dependencies,
lockfiles, new routes, generic frameworks, and unrelated seams are excluded.
Incident import, Account/Security/Users, session/bootstrap, directory/creation,
application menu, and workbook work remain regression baselines.

## Tracker

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| RP-01 — Baseline, projections, characterization | DONE | Six failures reproduced; complete response variants generated and verified. |
| RP-02 — Feature, query, selection ownership | DONE | Generation, cursor, selection, retention and characterization cases pass. |
| RP-03 — Attempts, jobs, lifetime, presentation | DONE | Captured replay, bounded observation/cancellation, lifetime and local feedback pass focused verification. |
| RP-04 — Browser, accessibility, visual evidence | DONE | Real lifecycle, keyboard/race checks, reviewed images and two fresh ordinary visual passes complete. |
| RP-05 — Final verification and handoff | DONE | Final broad/owner checks, compatibility, drift, accessibility/visual evidence, acceptance, documentation and scope closure complete. |

Only the current row is IN_PROGRESS. Record its terminal evidence before
advancing. An adopted-owner contradiction means BLOCKED: owner contradiction.

## Inspection and owner-to-change map

Inspected ReferencePackAdminPanel, referencePackAdminModel, services/referencePacks,
App, appSessionController/useAppSession, panel/browser tests, shared transport,
public error presentation, focus/overlay infrastructure, generated HTTP bindings,
authored Reference Pack/Job OpenAPI, protocol selection/generator, source ownership,
verification families, and the visual maintenance guide.

| Behavior | Adopted owner | Verification owners |
| --- | --- | --- |
| Query generations, exact filters, cursor binding | Core 01 REQ-01-610/610A, §§3.3.7 and 17.4 | module.reference_data, web.application |
| Selection and refresh omission | Core 01 REQ-01-481 | module.reference_data |
| Import, exact actions, replay and response variants | Core 01 §§11, 17.1.1 and 17.4 | module.reference_data, package.protocol_ts |
| Job identity, observation, cancellation and concealment | Core 01 §3.3.9; Core 04 REQ-04-023/028 | module.jobapi, platform.jobs, module.reference_data |
| App activation, retirement and recovery | Core 04 REQ-04-126; Core 01 REQ-01-610A | web.application |
| Forms, focus, feedback and resilient layout | Design §§4.5, 12, 14 | web.design, package.ui |

Caller inventory: App is the sole production ReferencePackAdminPanel caller;
panel tests call it directly. The imperative handle has zero consumers. All
Reference Pack service mutation adapters are consumed only by the panel.
App's activeJob/onJobChange pair mirrors panel internalJob; no separate consumer
requires that interface. Generated operation bindings have wider package
consumers and must remain backward compatible outside the corrected variants.

Initial source observations, subsequently characterized below: filter edits do not admit reads;
duplicate page/action clicks issue competing requests; one mirrored job replaces
another; failed job reads stop polling; callbacks couple polling to query edits;
adapters generate new IDs on each call; hidden panels load/poll; mutation/job
continuations lack full retirement fencing. Existing stale list success/error
suppression and list clearing on explicit admin loss are retained protections.

Response decision: add existing owner-permitted 202 JobEnvelope variants for
Activate/Disable; generate unions of distinct successful bodies and optional
status-specific validation. Preserve inline 200 ReferencePackActionEnvelope.
Backend behavior remains inline. JobResourceRef.kind already admits future
strings, so additive refs require no relaxed envelope schema.

## Lifecycle and recovery decisions

| Event | Retention / execution |
| --- | --- |
| Ordinary panel/route departure or hidden document | Retain input, authorized rows, selection, file, captured attempt, jobs and scroll; pause automatic reads. |
| Dispatched mutation finishes while inactive | Accept only under the same current authority/attempt; no focus or automatic catalog/job request. |
| Reactivation | Confirm access through AppSessionController; reconcile catalog/jobs before stale actions. |
| Read failure | Retain validated authorized state; local read recovery does not repeat mutation. |
| Ambiguous job 404 | Mark observation unavailable and confirm access; do not infer terminal job state or global demotion. |
| Session retirement/account change/admin loss/profile unclaim/disposal | Invalidate continuations; clear rows, selection, jobs, attempts, errors and file references. |
| Access confirmation temporarily fails | Preserve useful state, block protected actions, expose Retry access. |

User decisions: one unresolved operation at a time; selection survives query,
paging, and ordinary navigation changes with visible hidden-selection scope.
Explicit deselection/clear and retirement remove selection. Refresh all omits
pack_keys; Refresh selected captures nonempty unique exact keys. History is local
to this application/session lifetime; reload loses recovery and server work may
continue. No job discovery, persistence, or implicit latest version exists.

Implementation timing defaults: immediate material search/filter admission;
one-second automatic job polling; 30-second read/cancel bounds; 120-second
admission bound. These are implementation choices, not adopted product timing.

## Verification log

Planning resolved make help/help-all and task-guide for web.application,
web.design, module.reference_data, module.jobapi, platform.jobs,
package.protocol_ts, and package.ui. Current frontend source owner web.app is
distinct from verification owner web.application. Existing panel rows are owned
by module.reference_data, with a stale-search regression also routed by
web.application; the existing browser row belongs to module.reference_data.

Before edits, cache-off `make test-slice OWNER=module.reference_data ROWS=...`
passed the two Reference Pack frontend rows, 3/3 harness units, at
`.cartulary/test-results/20260907T204400Z-p1139856`. The earlier cached run at
`20260907T204231Z-p1138783` is not new execution evidence. Branch/HEAD/dirty-state
recheck at implementation entry and git diff --check passed.

RP-01 exit: characterization at `20260907T205205Z-p1145406` reproduced four
failures; expanded characterization at `20260907T205340Z-p1149820` reproduced
six, with five preserved cases passing (non-admin hiding is a separate row).
Failures concern automatic query admission, inactive-panel loading, duplicate
paging, duplicate refresh, missing exact replay and missing observation recovery.
They are product assertions against the original panel/execution code.

`make generate` initially failed at `20260907T205339Z-p1149255` because the
release change set lacked the two additive 202 responses. Only those two
reviewed deltas were added to the authored pending release change set; historical
release snapshots remain untouched. Generation passed at
`20260907T205659Z-p1155414`. The protocol test initially exposed an unexported
per-status validator; the generator now includes each response variant in its
public validator roots. Cache-off generated HTTP binding verification passed
2/2 units at `20260907T205742Z-p1158609`, covering complete inline/job envelopes,
status mismatches, unexpected statuses, incomplete jobs and legacy two-argument
validation. Next action: RP-02 pure state/query ownership.

## Acceptance, compatibility, and rollback

Final A001–A027 dispositions, changed paths, evidence and limits follow below.
Advisory posture: ADOPT keyboard/focus, semantic controls,
local errors and explicit async recovery; ADAPT responsive density/overflow and
compact long-token layout through existing tokens; REJECT dashboards, decorative
animation, replacement palettes and advice as product authority. Workbook-only
criteria remain regression boundaries rather than permission to redesign them.

No pack-format, API-route, database or stored-data migration. Rollback must move
implementation, App callers, authored projections, generated bindings, routing,
selectors and reviewed visual evidence together. Runtime and executable checks
must never consume Markdown or digest content.

## RP-02 exit

The pure model now separates raw input, normalized query, accepted rows,
query-bound pagination, retained pack-key selection, operation attempts, known
jobs, and availability. The ReferencePackAdminController executes bounded
transport work behind lifetime/generation fences; useReferencePackAdmin binds
activation, document visibility and native focus/input state. App retains only
activation/lifetime/access wiring. The panel consumes that single owner.

Removed ReferencePackAdminPanelHandle, forwardRef/useImperativeHandle,
internalJob, activeJob, onJobChange and App's referencePackJob mirror; repository
caller search confirms no remaining consumers. Service adapters are now consumed
by the feature executor. Selection is retained by key, with hidden selection
and explicit clear controls; exact row identities are coalesced without sorting.
Refresh invalidation cannot replace the newest effective query or retry a failed
query behind the user's back.

The first integration run (`20260907T211514Z-p1170636`) passed all six
characterization gaps and the new pure model cases, but exposed an old immediate
fetch assertion that no longer awaited dispatch. Its request gate was corrected.
Type errors in optional shared pagination and test literal inference were fixed;
missing route paging is explicitly rejected. Cache-off narrow owner verification
passed 4/4 harness units at `20260907T212033Z-p1177312` (all selected panel and
model cases). `make frontend-typecheck` passed 2/2 at
`20260907T212032Z-p1176948`. Initial formatting caught an obsolete App effect
dependency after mirror removal; that dependency was removed. Next action:
RP-03 exact replay, job/cancellation, lifetime and presentation hardening.

## RP-03 exit

Captured immutable attempts own command, exact pair/key set, transaction ID and
File/Blob bytes. Adapters consume that capture and preserve the received HTTP
status through malformed bodies. Valid inline responses commit immediately;
validated job receipts remain accepted until an authoritative terminal read.
Fresh version actions re-read eligibility; replay never changes the request.
Only one unresolved admission/job is allowed. Completed jobs remain keyed by
job_id; dismissal removes only local terminal records. Uncertain cancellation
cannot be replaced by a newly keyed cancel. A fresh terminal observation resolves
uncertain cancellation without claiming rollback.

Job observation uses bounded non-overlapping reads, pauses while hidden, and
exposes local retry after read failure. Pure job transitions reject identity,
progress, timestamp and terminal regressions. Unknown reference kinds remain
additive; canonical known references and complete required envelopes are checked.
A concealed 404 invokes App's existing access-confirmation boundary. No browser
clock is used to infer session/job expiry. Retirement synchronously drops file,
rows, selection, attempts, jobs and errors and fences ignored-abort continuations.

The upload dialog is replaced with a native required file form. Selection has
native key-scoped checkboxes and visible scope outside loaded rows. Catalog
reload, pack refresh, exact version actions, observation and cancellation have
separate labels and feedback. Public error identity and shared presentation
classification are used without exposing wire messages or private diagnostics.
A stable live region reports meaningful transitions, not each poll; native input,
scroll and lost-origin focus are isolated in the binding. Nonexistent style
variables found during review were replaced with existing tokens.

`make format` passed at `20260907T213218Z-p1200776`. Cache-off five frontend
owner rows passed 6/6 harness units at `20260907T213234Z-p1205095`, including
13 executor recovery cases and five adapter cases in addition to model/panel
regressions. Typecheck and import-boundary checks passed at
`20260907T213234Z-p1205191` and `20260907T213234Z-p1205203`. An earlier adapter
fixture lacked JSON Content-Type; the complete fixture was corrected rather than
loosening parsing. The initial import-boundary failure identified route building
in the panel; it now lives in the service adapter. Next action: real backend,
keyboard, accessibility and reviewed visual evidence in RP-04.

## RP-04 validation record

The complete service-backed Reference Pack owner slice at
`20260907T214105Z-p1219219` passed every backend integration row and the
existing race/browser row plus async-action fixtures. Its new real-workflow row
first failed because its helper incorrectly required a populated initial catalog.
The real empty catalog is valid; the helper now accepts it. The next attempt at
`20260907T214344Z-p1369389` proved safe verification failure for a test bundle with
the wrong contract token. The fixture now uses the existing
`cartulary.reference_pack.v1` contract. A third attempt reached successful
import/replay and found a test assertion expecting “Inactive” instead of the
existing Active-column value “No”. The test now checks that exact column.
The complete real import/replay/activate/disable/reverify/selected-refresh/all-
refresh workflow passed 11/11 units at `20260907T215019Z-p1524885`. Returned
inline status is 200 for Activate/Disable and 202 for Reverify; upload metadata,
bytes and admission job identity match across lost-response replay. Import
remains inactive until explicit activation. Evidence is attached to that row.

Application regression coverage passed 76/76 units at
`20260907T214108Z-p1220229`. Reference Pack keyboard/accessibility coverage passed
11/11 at `20260907T214107Z-p1219447` and again at
`20260907T215219Z-p1614863`, including native file chooser, repeated recovery,
key selection shared by versions, observation, cancellation, dismissal, visible
focus, named controls, reduced motion, 390/640/768/1024/1280 widths, short
viewports, 200% zoom and text spacing. The row retains an accessibility tree.
Focused frontend regressions passed 6/6 at `20260907T214345Z-p1369631`.

The first ordinary visual run (`20260907T214245Z-p1321360`) exposed the new
scenario ID's stricter capture-schema format. Authored IDs were corrected to
stable hexadecimal IDs and regenerated; no harness/schema rule was weakened.
The next ordinary run (`20260907T214755Z-p1420422`) passed all existing visual
rows and reconciled 152 capture intents, 139 existing active goldens, zero
orphans/ambiguous mappings, and 26 resolved registered fixtures. The only 13
missing goldens are the newly introduced Reference Pack capture intents, all
owned by `web.design.visual.reference_pack_administration`, scenario
`scenario_034f30c7dcac`, project chromium. No existing missing golden or orphan
is being silently repaired or deleted. The initial captures are authorized by
this seam; existing golden mutation remains blocked. New images are generated
only through `make browser-e2e-visual-update`, then reviewed and followed by two
fresh ordinary passes. Subsequent review and final run roots are recorded below.

The visual trigger is the intentionally changed Reference Pack administration
workflow. The 13 stable capture names are reference-pack-catalog, selection,
submitted, recovery-short, recovery-zoom, recovery-spacing, queued, running,
observation-recovery, cancel-requested, canceled, succeeded and failed, all with
the reference-pack- prefix. There are no new D-VFIX registry claims: emitted
capture intents provide exact row/scenario/project accounting. Existing viewport
capture mechanics, font checks and renderer remain unchanged. New fixtures use
1280×900, 640×480, 1280×720 at 200% zoom, and 768×640 with text spacing; no
new masks or selector crops are introduced.


Visual candidate review found and fixed two presentation problems. A preceding
Account fixture had left an intentionally oversized display name in shared worker
state. The Reference Pack presentation fixture now normalizes only that display
label on the real authenticated session, preserving real identity and authority;
this follows the existing Incident import fixture boundary. No shared header or
Account behavior changed. Native indeterminate progress animation also prevented
stable captures; token-styled native progress now uses static stripes with a
textual unknown-total label. The failed update at `20260907T215200Z-p1572184`
left the committed snapshot directory unchanged.

Update runs `20260907T220258Z-p1727848` and
`20260907T221311Z-p1896874` each passed 12/12 units. All unrelated tracked PNGs
rewritten by update mode were restored byte-for-byte from the clean baseline
(14 and 11 respectively); Make regenerated the manifest. Review improved the
exact-identity column allocation and compacted Deselect while retaining its full
accessible key name. Ordinary reconciliation at `20260907T220944Z-p1793275`
accounted for 152 active goldens with no missing, orphan or ambiguous mappings;
only the intentionally changed Reference Pack row differed. Subsequent review
of every candidate found squeezed table headings in short/zoom/text-spacing
states. A local container layout now stacks labeled native table cells at narrow
widths, retaining the same controls and semantic identities. This is a current
presentation correction, not a new catalog or grid capability. Final promotion
and two ordinary passes remain required before RP-04 exits.

Additional bounded race review added authoritative terminal receipt scheduling
behind an existing known-job read, post-dismissal response suppression, and
post-disposal input rejection. The executor slice passed 2/2 units with 14 cases
at `20260907T221242Z-p1892175`. An actual ancestor-scroll retention regression
passed 2/2 units at `20260907T220145Z-p1726966`; the binding now captures/restores
the real scrolling ancestor instead of assuming the panel itself scrolls.
Latest pre-responsive accessibility evidence passed 11/11 units at
`20260907T220945Z-p1794750`. These refinements preserve the RP-02/RP-03 ownership
and recovery decisions and are validated as RP-04 race/presentation findings.


Ordinary run `20260907T222354Z-p1955762` reconciled all 152 active goldens and
26 registered fixtures, with zero missing/orphan/ambiguous entries. Only the
three responsive Reference Pack captures differed; all pre-existing visual rows
passed. Review fixed the stacked caption's intrinsic width. The expanded a11y
row initially failed an overly broad 100px label-width assertion at the existing
390px shell boundary (`20260907T222550Z-p2054685`): the unchanged 16rem admin
navigation leaves a narrower content column. The final check verifies readable
bounded label wrapping and reachable controls inside the actual available width;
it does not claim a shared navigation redesign. It passed 11/11 units at
`20260907T222759Z-p2105434`, including the added narrow version-action checks.
Browser runs are now sequential after concurrent Make builds emitted missing
shared-dist hash warnings; no harness or shared build change is in this seam.


### Required-case evidence map

| Cases | Focused evidence path / owner route |
| --- | --- |
| Stale success/error, accepted-row retention, auth clearing, effective filter edits, Enter, duplicate paging/actions | `apps/web/src/app/ReferencePackAdminPanel.test.tsx`; existing module.reference_data frontend panel rows |
| NFC/Unicode whitespace normalization, query-bound cursors, duplicate exact identities, retained key selection | `apps/web/src/app/referencePackAdminModel.test.ts`; module.reference_data.frontend.administration_query_ownership |
| Query/job completion races, failed query not silently retried, committed mutation plus failed catalog read | `apps/web/src/app/referencePackAdminController.test.ts`; module.reference_data.frontend.administration_recovery |
| Exact selected set/all omission, eligibility preflight, duplicate commands, lost admission/upload replay, deadlines | Same controller route plus module.reference_data.frontend.administration_transport |
| Failed observation/retry, multiple known jobs, terminal receipt behind another read, late dismissal responses, monotonic snapshots | Controller recovery route |
| Cancellation preflight, rejection/completion race, uncertain exact cancel replay, concealed 404 access confirmation | Controller recovery route and existing module.jobapi/platform.jobs owner evidence |
| Hidden-panel admission, no hidden automatic reads, retirement/disposal with ignored abort, actual scroll ancestor retention | Controller recovery and existing panel routes; web.application session/application baselines |
| Complete 200/202 envelopes, status mismatch, malformed success, transport uncertainty, safe error categories, additive refs | `apps/web/src/services/referencePacks.test.ts` and `packages/protocol-ts/src/index.test.ts` |
| Real import/lost-response replay/activate/disable/reverify/selected refresh/all refresh | `apps/web/e2e/reference-pack.spec.ts`; module.reference_data.browser.administration_real_lifecycle |
| Permitted async Activate/Disable, shared key selection across versions and searches, multiple local jobs | Same browser path; module.reference_data.browser.administration_async_recovery |
| Keyboard upload, selection, actions, recovery, cancellation, dismissal, narrow/short/zoom/text-spacing focus | `apps/web/e2e/workbook.a11y.spec.ts`; web.design.accessibility.reference_pack_administration |
| Catalog, hidden selection, submitted/uncertain outcomes, all six job states and observation recovery | `apps/web/e2e/workbook.visual.spec.ts`; web.design.visual.reference_pack_administration |

All transport fixtures include complete required envelopes and use deterministic
request gates. The real archive helper lives in
`apps/web/e2e/support/referencePacks.ts`, separate from production adapters.
No optional job discovery, durable history, latest-version action, fuzzy search,
backend long-running Activate/Disable implementation or generalized job framework
is treated as a missing current capability.


Final candidate generation passed 12/12 units at
`20260907T222918Z-p2152088`. The 12 unrelated tracked PNG rewrites from update
mode were rejected and restored from HEAD; `make generate` then passed at
`20260907T223233Z-p2201250`. Every one of the final 13 Reference Pack images was
visually inspected. Review accepts exact long identities, multiple versions,
hidden selection scope, submitted/uncertain feedback, local recovery, all six
job states, determinate/indeterminate progress, and stacked narrow table cells.
Short, zoom and spacing captures keep recovery visible without squeezed column
headings. Ordinary viewport cropping is unchanged; controls below the viewport
remain reachable through native scroll and keyboard checks. Existing theme,
fonts, shell navigation and prior goldens remain baseline inputs.

Current-source typecheck passed 2/2 units at `20260907T223208Z-p2200137`;
import-boundary verification passed 2/2 at `20260907T223208Z-p2200144`.
Repository searches found no remaining deleted imperative/mirror consumers,
private Incident/Account controller imports, or Markdown/digest dependency in
changed runtime, generator or feature test sources. Git diff --check passed.


The first post-promotion ordinary visual run passed 12/12 units at
`20260907T223250Z-p2204392`. During final route-retention review, a further
native-form gap was reproduced at `20260907T223450Z-p2252979`: a remounted file
picker is empty even though the application owner retains the selected File;
unconditional native required validation blocked import. The regression passed
all other panel cases and failed exactly the retained-file validity assertion.
The picker is now required only when the owner has no selected file. A real
backend scenario now leaves the administration route before initial submission,
returns to the empty native picker with the retained filename, then submits the
original bytes. This fixes ordinary route retention without fabricating a native
file-picker value. Two final-source ordinary visual passes are still required.


The retained-upload panel regression passed 2/2 units at
`20260907T223556Z-p2258805`; regeneration passed at
`20260907T223554Z-p2258240`. The expanded real backend workflow passed 11/11
units at `20260907T223617Z-p2261848`, proving initial upload from the retained
File after route remount, then lost-response exact replay and all existing pack
lifecycle actions. No screenshot changes follow from the conditional native
required attribute; the reviewed 13 goldens remain the accepted candidates.


Ordinary visual validation passed 12/12 units at
`20260907T223727Z-p2308356`. A combined known-job reactivation test then
reproduced a recovery stall at `20260907T224112Z-p2356528`: access confirmation
for an unavailable old terminal job interrupted reconciliation while the current
job remained paused. The scheduler now resumes a paused current observation once
current access is confirmed; failed/unavailable observations still require local
recovery. This is a one-condition change in the existing bounded scheduler, not
new job infrastructure. The deterministic multiple-job test retains the old
unavailable snapshot and checks resumed current observation independently of an
ignored-abort response. Final-source ordinary passes follow this last correction.


The combined known-job recovery regression passed 2/2 units at
`20260907T224154Z-p2361509`. The complete five-row frontend feature slice passed
6/6 units at `20260907T224223Z-p2398407`, including all preserved and newly
reproduced cases. Source changes are frozen for the remaining acceptance runs.


The next ordinary run (`20260907T224208Z-p2362200`) passed all other captures
but found two cancellation snapshots at inconsistent vertical scroll offsets.
Actual-image review showed correct state and text, with the viewport depending
on observation recovery/cancellation timing. Reconciliation still accounts for
152 active goldens, 26 registered fixtures and zero missing/orphan/ambiguous
entries. The two Reference Pack cancellation captures now await observation
recovery and cancellation acknowledgment, then explicitly align the operations
region. This is a declared fixture scroll-normalization correction for
reference-pack-cancel-requested and reference-pack-canceled only; it changes no
production focus behavior, viewport, mask, crop, renderer or pre-existing fixture.
The candidate update and fresh ordinary validations will be reviewed again.


The scroll-normalized update passed 12/12 units at
`20260907T224717Z-p2417144`. All 13 resulting Reference Pack PNGs were inspected
again. The intended cancellation anchors also move the subsequent success
capture's viewport; that composition was reviewed and accepted with current
scope, terminal history and result feedback visible. The 11 unrelated tracked
PNG rewrites were restored from baseline, and Make regenerated the manifest at
`20260907T225110Z-p2466059`. No unrelated image is included in the change.


The first fresh ordinary run after the final capture correction passed 12/12
units at `20260907T225150Z-p2469316`. Its
`browser-e2e-visual/frontend-visual-reconciliation.json` reports PASS with all
152 capture intents/goldens active, all 26 registry fixtures resolved, and zero
missing, orphan or ambiguous mappings. The subsequent run and RP-04 exit below record the second successful pass.


The next ordinary run (`20260907T225441Z-p2517993`) passed the Reference Pack
row and its screenshots but failed an unchanged Entity Linking visual fixture
before comparison: its dismissed-mention element was absent during an accessible
name assertion. This is outside the Reference Pack diff and occurs before its
fixture runs; no entity behavior, assertion, selector or golden was changed.
The owning row is
`module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`.
The failure is retained and reported rather than absorbed into a golden update.
A further fresh ordinary run is required for the second successful full pass.


## RP-04 exit

The second successful fresh ordinary visual run passed 12/12 units at
`20260907T225813Z-p2566317`. Together with `20260907T225150Z-p2469316`, this
satisfies the two fresh ordinary passes after the final promotion. Both retain
PASS reconciliation for 152 active capture intents/goldens, 26 resolved registry
fixtures, and zero missing/orphan/ambiguous mappings. The intervening unchanged
Entity Linking assertion failure passed on rerun without source or golden edits.
All 139 baseline PNGs were independently compared to HEAD and remain
byte-identical; exactly 13 new Reference Pack PNGs are included.

RP-04 exits DONE with the real backend, async fixtures, deterministic races,
keyboard/accessibility, responsive layout and reviewed visual evidence above.
No blocked dependency remains. Next action: mark RP-05 IN_PROGRESS, run
`make agent-finalize` with RESULTS_DIR unset, then broader owner and projection
verification and final acceptance/scope/Markdown checks.


## Final architecture and behavior

The bounded owner is `ReferencePackAdminController`, instantiated by the App
lifetime binding. Pure model transitions own accepted state and identity; the
executor owns bounded transport work and continuation fences; the React binding
owns visibility, native file input, actual scroll ancestor and explicit-action
focus recovery. The panel renders that state. App supplies activation, current
session identity, admin/profile authority, retirement and existing access recovery
ports. It owns no Reference Pack job snapshot or callback mirror.

Current behavior fixes are immediate normalized query admission, complete
pagination generation fencing, retained key selection and visible scope,
serialized captured attempts, exact request replay, local observation/cancellation
recovery, lifecycle clearing, native upload retention and keyboard feedback.
Structural changes remove competing job ownership, imperative handles and the
upload dialog. These are distinct from forward-compatible response support:
authored Activate/Disable projections now expose the already adopted 202 job
variant alongside current backend 200 inline completion. Complete contract
fixtures verify asynchronous variants; no backend long-running algorithm changed.
The optional status argument on generated response validation preserves existing
two-argument callers while allowing Reference Pack adapters to validate the
actual received status. No successful-response validation was weakened.

| Changed area | Source / behavior owner | Verification and review |
| --- | --- | --- |
| Panel, model, executor, binding, adapter and test fixtures | web.app source boundary; Core 01 Reference Packs/jobs | module.reference_data; web.application integration |
| App activation/lifetime ports | Existing AppSessionController/Core 04 boundary | web.application and focused retirement tests |
| Authored 202 response declarations and reviewed additive release deltas | module.reference_data OpenAPI owner | package.protocol_ts; platform.openapi compatibility |
| Successful response variant generator and generated projections | package.protocol_ts; authored OpenAPI source | protocol tests, typecheck, generation/drift and artifact policy |
| Source ownership, family rows and browser batching | Authored harness/verification catalogs | catalog/JSON-shape/generation checks |
| Forms, focus, tokens, layout and capture intents | Design direction within Core behavior | web.design accessibility/visual; package.ui selector compatibility |

Deleted interfaces: ReferencePackAdminPanelHandle, forwardRef,
useImperativeHandle, internalJob, activeJob, onJobChange, and App's
referencePackJob/setReferencePackJob mirror. App was the only production panel
caller; no imperative consumer existed. Old independently dispatched mutation
adapters were replaced by captured-attempt submission consumed by the feature
executor. No private Incident import or Account controller was reused.

## Acceptance dispositions

These are localized advisory dispositions, not new product requirements or a
Core 05 conformance claim. Existing owner behavior remains authoritative.

| ID | Disposition | Evidence / boundary |
| --- | --- | --- |
| A001 | PASS | Adopted-owner map above; draft NLSpec and digest remain advisory. |
| A002 | PASS | One Reference Pack seam; current fixes, structural work and response projection support distinguished. |
| A003 | PASS | main/d115e800 clean baseline, current owners/callers/routing/generated roots inspected; no grid imports. |
| A004 | ADAPT | Existing color, spacing, border, typography and focus tokens; no second registry. Font-relative intrinsic form bases, column percentages and a local 42rem container layout threshold are geometry decisions, not new theme/density tokens. |
| A005 | PASS | Existing dark_graphite presentation and renderer retained; no palette/theme expansion. |
| A006 | NOT APPLICABLE | Workbook density remains a regression baseline; the native paginated catalog does not create a density registry. |
| A007 | NOT APPLICABLE | Workbook creation rules unchanged; upload uses its adopted deployment-admin admission contract. |
| A008 | ADAPT | Local container layout stacks labeled table cells; narrow/short/zoom/text-spacing controls verified. Shared chrome thresholds, visualViewport fallback and inspector geometry unchanged. |
| A009 | PASS IN SCOPE | Actual scroll ancestor retained; local controls fit available content width. Existing shell/navigation remain baseline, including the narrow fixed admin navigation column. |
| A010 | NOT APPLICABLE | No inspector or semantic dispatcher changes; exact pack version actions retain current eligibility. |
| A011 | PASS IN SCOPE | Query rows, key selection, file, attempts, known jobs and scroll retained across ordinary navigation. Browser may clamp scroll when reconciled content becomes shorter. Forbidden lifetime transitions clear protected state. |
| A012 | PASS | Existing Web Crypto transaction generation; only exact uncertain replay reuses captured IDs and payload/bytes. |
| A013 | NOT APPLICABLE | Workbook blocked-edit queue policy unchanged. Applying its new-ID retry to Reference Pack uncertainty is rejected; adopted route replay preserves the original ID. |
| A014 | ADAPT | Native form, checkbox, action and recovery keyboard parity; no cell editor, paste or Escape-ladder redesign. |
| A015 | ADAPT | Operation-local conflict/error/recovery rather than a toast; captured target stays distinct from later query or selection edits. |
| A016 | PASS | Initial unavailable, loading/searching, empty, filtered empty, retained stale failure, authorization loss and operation outcomes are distinct. |
| A017 | PASS | Authorized accepted rows survive search/refresh/read failure; explicit auth/lifetime retirement clears them. |
| A018 | NOT APPLICABLE | Evidence lifecycle/preview matrix unchanged and retained as a regression boundary. |
| A019 | PASS IN SCOPE | Native keyboard workflow, accessible names/tree, visible token focus, non-color status/error text, stable announcements and reduced-motion checks. |
| A020 | PASS IN SCOPE | Focused state/race matrix plus 13 reviewed captures cover declared Reference Pack states and pending/recovery precedence. |
| A021 | NOT APPLICABLE | Workbook virtualization unchanged; no grid framework introduced into the native server-paginated catalog. |
| A022 | PASS | Exact capture intents/owner/scenario/project reconcile; existing 26 registered fixtures retained. New captures make no D-VFIX registry claim. |
| A023 | PASS IN SCOPE | Existing UI-contract selectors reused; exact pair/key identities and semantic native roles/names drive tests. |
| A024 | PASS | No changed runtime, generator, test or release input depends on Markdown/digest contents. This handoff is human evidence only. |
| A025 | PASS | Authored corrections regenerated through Make; final generation drift 4/4 and generated-artifact policy 3/3 passed. |
| A026 | PASS | Existing routes, authorization, lifecycle and storage behavior retained; no format, migration, pack kind, auto-activation or latest-version feature. |
| A027 | PASS | Final behavior, owners, exact paths, commands/results, reviewed evidence, limits, rollback and tracker closure recorded; no next seam authorized. |

Advisory classifications: ADOPT semantic controls, focus continuity, explicit
async outcomes, local safe errors, deterministic request gates and actual image
review. ADAPT layout guidance to a compact deployment-local catalog using current
tokens and shell boundaries. REJECT dashboards, decorative animation, generated
palettes, generic job/mutation frameworks, client ranking and advice that
contradicts exact replay or existing ownership. No deferred optional capability
is presented as implemented.

## Compatibility, limits, and rollback

Existing Reference Pack behavior and APIs remain authoritative. There are no new
routes, dependencies, lockfile changes, backend algorithms, pack/archive formats,
database migrations or stored-data migrations. Import stays staged_only with
separate activation. Exact version identity and server ordering are preserved.
Refresh all omits pack_keys and resolves scope at admission; refresh references
are not treated as an exhaustive changed-version list.

Only one unresolved operation is allowed across the feature. Completed known jobs
may remain in local history and can be dismissed; no job list API or durable
operations history exists. A browser reload loses the local recovery record even
when server work continues. Ordinary navigation/hidden documents pause reads;
already-dispatched mutations may settle in the same valid lifetime. Reactivation
confirms access and reconciles current state. Temporary observation failure
retains recoverable state; confirmed session/account/admin/profile retirement
clears protected rows, keys, files, attempts, jobs and errors. Admission/read/cancel
timeouts fence late continuations independently of transport abort.

Uncertain admission/cancellation recovery resends only the captured authorized
request. Conflict recovery observes current state and never substitutes another
mutation. A committed action plus failed catalog refresh remains committed;
recovery repeats the read. Cancellation acknowledgment establishes neither
cancellation nor rollback. Unknown additive job references have no invented
actions; inconsistent required envelopes remain observation/admission errors.

Rollback is one cohesive implementation unit: panel/model/executor/binding/service
adapter and App callers; authored OpenAPI owner/release deltas and protocol
generator; regenerated contracts/bindings; source ownership and verification
families/batching/topology; the 13 reviewed Reference Pack goldens and generated
manifest. Restore these together to the recorded baseline or a reviewed follow-up
change and regenerate through Make. Do not hand-edit generated roots, retain a
caller against a removed interface, or roll back unrelated goldens. No data
rollback or migration is required. No commit, push or deployment was performed.

## Changed paths

- `apps/web/e2e/reference-pack.spec.ts`
- `apps/web/e2e/support/referencePacks.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-cancel-requested-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-canceled-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-catalog-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-failed-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-observation-recovery-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-queued-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-recovery-short-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-recovery-spacing-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-recovery-zoom-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-running-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-selection-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-submitted-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/reference-pack-succeeded-linux.png`
- `apps/web/src/app/App.tsx`
- `apps/web/src/app/ReferencePackAdminPanel.test.tsx`
- `apps/web/src/app/ReferencePackAdminPanel.tsx`
- `apps/web/src/app/referencePackAdminController.test.ts`
- `apps/web/src/app/referencePackAdminController.ts`
- `apps/web/src/app/referencePackAdminModel.test.ts`
- `apps/web/src/app/referencePackAdminModel.ts`
- `apps/web/src/app/useReferencePackAdmin.ts`
- `apps/web/src/services/referencePacks.test.ts`
- `apps/web/src/services/referencePacks.ts`
- `apps/web/src/testing/referencePackTestSupport.ts`
- `contracts/openapi-releases/2.0.0.change-set.json`
- `contracts/openapi-source/owners/module.reference_data/openapi.json`
- `contracts/openapi/cartulary.openapi.yaml`
- `docs/handoffs/reference-pack-administration-refactor-handoff.md`
- `internal/gen/contractopenapi/artifacts_gen.go`
- `internal/gen/openapioperations/catalog_gen.go`
- `packages/protocol-ts/src/generated/core-http-types.ts`
- `packages/protocol-ts/src/generated/core-http-validators.ts`
- `packages/protocol-ts/src/generated/http-operation-bindings.ts`
- `packages/protocol-ts/src/index.test.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/protocol-ts/generate-protocol-types.mjs`
- `tools/test_families/module.reference_data.json`
- `tools/test_families/web.design.json`


## RP-05 final verification

`make agent-finalize` passed 1/1 unit at `20260907T230119Z-p2614325` before
broader verification. Its `unit-artifacts/finalize-summary.json` records zero
updated files, passing schema/catalog/tier validation and generated-structure
refresh/drift. RESULTS_DIR was unset because no qualifying exact-source full
warm evidence was supplied. Retained-run selection, canonical evidence checks,
scheduler event-order/timing drift and performance evidence maintenance were
therefore skipped with `results-dir-not-provided`; these skips are not product
check passes. The broader checks below execute independently with caching off.


The first broad `make check CARTULARY_HARNESS_CACHE_MODE=off` run at
`20260907T230315Z-p2617929` exposed two additional authored-test boundary issues:
three fixture exports had no external importer, and the scroll-host unit fixture
used an unowned data-testid literal. Caller inventory confirmed the exports are
private helpers; export modifiers were removed. The unit fixture now uses a named
native section/region instead of introducing a selector contract. No production
behavior, rendered browser fixture or golden changed. The harness.browser policy
slice passed 2/2 units at `20260907T230527Z-p2684596`; remaining policy and broad
reverification results follow.


The initial broad check completed 748/750 units; the only failures were the two
owned test-support issues just corrected. The selector policy slice passed 2/2
units at `20260907T230637Z-p2708349`, and formatting passed at
`20260907T230658Z-p2718637`. A fresh complete cache-off check is running after
both corrections. Read-only final checks already passed: OpenAPI compatibility
4/4 units (`20260907T230855Z-p2804762`), generated-artifact policy 3/3
(`20260907T230855Z-p2804751`), and frontend import boundaries 2/2
(`20260907T230855Z-p2804957`). The protocol operation selection inputs
`contracts/protocol-ts/http-operations.v2.json` and
`contracts/protocol-ts/frontend-entrypoints.v2.json` were inspected; existing
Reference Pack/job selection was sufficient and remains unchanged.


The fresh full `make check CARTULARY_HARNESS_CACHE_MODE=off` passed all 750/750
units at `20260907T231007Z-p2807840`. No units failed or were skipped, and all
750 bypassed cache. This includes current typecheck, frontend/script/Go lint,
JSON shape/catalog policy, the corrected support policies, affected package
protocol/UI checks, Reference Pack/Job API/platform jobs backend ownership, and
application/workbook/auth/session/bootstrap/directory/creation/menu/Account/
Incident-import regressions routed by the check graph. The first 748/750 run
remains failed evidence; this complete rerun is the final broad result.


Full `make generate-drift CARTULARY_HARNESS_CACHE_MODE=off` passed 4/4 units at
`20260907T231644Z-p2981629`, confirming the complete generated projections remain
reproducible after final source/test corrections. No generated artifact was
hand-edited. The earlier Markdown lint pass at `20260907T231103Z-p2858210`
covered the expanded handoff; final closure edits will be linted again.


Full `make browser-e2e-a11y CARTULARY_HARNESS_CACHE_MODE=off` passed 12/12 units
at `20260907T231722Z-p2985054`, covering all 19 routed accessibility rows and
their selected scenarios. This includes the Reference Pack keyboard/recovery
workflow plus the existing authentication, Account/navigation, workbook,
collaboration, evidence, entity, saved-view and Network Analysis baselines.


Final `make service-backed-test-slice OWNER=module.reference_data
CARTULARY_HARNESS_CACHE_MODE=off` passed all 13/13 units at
`20260907T231937Z-p3033159`. It reruns the existing real backend import,
activation/disable/reverify/refresh, admission/replay, authorization, upload
failure, cancellation and disconnected-bundle protections plus all three owned
Reference Pack browser rows. The real route-remount/upload/exact-replay workflow
and complete asynchronous Activate/Disable fixtures pass on final source.


`make protocol-ts-browser-artifact-reachability` passed at
`20260907T232128Z-p3083696` (exit 0), verifying 12 JavaScript bundles and 12 source
maps. Its retained tool summary lives under
`protocol-ts-browser-artifact-reachability/tool-run-summary.json` in that run.

### Final result index

Run IDs below are relative to `.cartulary/test-results/`. Graph targets retain
`run-summary.json` and `target-summaries/<target>.json`; individual row evidence
is retained beneath `rows/`, browser groups and unit artifacts. This is execution
and implementation/design evidence, not a release or Core 05 publication claim.

| Make command | Final result / run |
| --- | --- |
| `make agent-finalize` | PASS 1/1; 20260907T230119Z-p2614325; RESULTS_DIR unset, retained-run maintenance skips recorded above. |
| `make check CARTULARY_HARNESS_CACHE_MODE=off` | PASS 750/750, zero skips/cache hits; 20260907T231007Z-p2807840. Includes typecheck, lint, JSON/catalog, affected owner/package and regression slices. |
| `make frontend-import-boundary-check CARTULARY_HARNESS_CACHE_MODE=off` | PASS 2/2; 20260907T230855Z-p2804957. |
| `make service-backed-test-slice OWNER=module.reference_data CARTULARY_HARNESS_CACHE_MODE=off` | PASS 13/13; 20260907T231937Z-p3033159. |
| `make browser-e2e-a11y CARTULARY_HARNESS_CACHE_MODE=off` | PASS 12/12; 20260907T231722Z-p2985054. |
| `make browser-e2e-visual CARTULARY_HARNESS_CACHE_MODE=off` | Two fresh post-promotion PASS runs, 12/12 each: 20260907T225150Z-p2469316 and 20260907T225813Z-p2566317; 152 active goldens, zero missing/orphan/ambiguous entries. |
| `make openapi-compatibility-check CARTULARY_HARNESS_CACHE_MODE=off` | PASS 4/4; 20260907T230855Z-p2804762. |
| `make generated-artifact-policy-check CARTULARY_HARNESS_CACHE_MODE=off` | PASS 3/3; 20260907T230855Z-p2804751. |
| `make generate-drift CARTULARY_HARNESS_CACHE_MODE=off` | PASS 4/4; 20260907T231644Z-p2981629. |
| `make protocol-ts-browser-artifact-reachability` | PASS; 20260907T232128Z-p3083696; 12 bundles and maps verified. |

Earlier product characterizations, invalid test fixtures, policy failures,
responsive/capture corrections and the intermittent Entity Linking fixture
failure are explicitly retained above with their successful correction/rerun
routes. There is no unresolved failing required check. Full CI/release-check were not needed for the selected owner-routed implementation
acceptance and were not run; no release-readiness claim is made. Persistent
recovery, new job discovery and broader administration remain excluded scope.

### Final scope review

Final inspection still reports branch `main`, HEAD
`d115e800414fc8663fdcabf07a151796087249fd`. The initial tree was clean; the final
tree contains 45 changed paths listed above, all within the authorized seam.
There are 13 new Reference Pack PNGs; all 139 existing PNGs are byte-identical
to HEAD. No authored backend, migration, lockfile, dependency, digest or other
handoff changed. Generated Go/TypeScript/OpenAPI/topology/manifest differences
come from the recorded authored corrections and Make generation. No unrelated
changes were overwritten; update-mode image noise was restored to the clean
baseline. No commit, push, deployment or subsequent seam was started.

Final documentation lint, whitespace and path-inventory audit are the RP-05 exit
checks. After recording closure, Markdown lint, git diff --check and the scope
audit are rerun once more against the final handoff text.


## RP-05 exit and closure

RP-05 exits DONE. The pre-closure `make lint-markdown` passed at
`20260907T232333Z-p3084494`; git diff --check passed, and the explicit allowed-path
inventory audit confirmed exactly 45 paths, unchanged main/HEAD, 139 unchanged
baseline PNGs and 13 new Reference Pack PNGs. The final closure edit is followed
by the requested repeat of Markdown lint, whitespace and scope checks; delivery
reports those final results without modifying the handoff again.

All five workstreams are DONE. Required acceptance has no unresolved failure or
blocked dependency. Historical failed attempts and the successful fixes/reruns
are retained above; retained-run maintenance skips remain explicitly disclosed.
The implementation is ready for review as one uncommitted seam. No commit, push,
deployment, broader administration/job infrastructure work or subsequent seam
has been performed or authorized by this handoff.
