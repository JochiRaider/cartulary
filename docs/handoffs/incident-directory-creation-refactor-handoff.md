# Incident-directory creation refactor handoff

## Completed iteration — incident creation and directory cleanup

Status: IC-06–IC-10 DONE.
The user approved the revised remediation plan, including the specification and
projection repairs below. The existing staged handoff edits are preserved; the
historical IC-01–IC-05 record remains unchanged.

This completed refactoring iteration covers the frontend incident-creation and
incident-directory lifecycle, including queries, pagination, and App integration.
Its objective is one coherent directory lifecycle with fewer handwritten
contracts, obsolete state paths, and compatibility fallbacks, while preserving
useful creation recovery and navigation behavior. Future growth should fit clear
ownership and typed boundaries without adding more coordination to App.

### Planning baseline, authority, and boundaries

- Inspected branch: `main`; HEAD:
  `1aa5bd152e97be4dffcf80e89ae2aba55ffb4a70` (`Incident-directory creation`).
- Execution baseline: HEAD remains the inspected commit. The only pre-existing
  change is the staged update to this handoff. IC-01–IC-05 are committed and their
  original implementation record is retained below as history.
- Allowed in this iteration: directory and creation frontend models,
  bindings, presentation, narrow HTTP adapters, immediate App integration,
  affected fixtures/tests/selectors, authored verification routing, and this
  handoff; adopted root-navigation owners and current restatements; authored
  paging projections and the unpublished OpenAPI compatibility change set.
  Generated evidence changes follow their existing Make-owned workflow
  only when required by implementation changes.
- Excluded: backend features, new public routes, authorization-rule changes,
  persistence, dependencies, immutable release baselines, digest sources, and unrelated
  workbook or account/admin features. Existing browser routes and ordinary
  workbook startup remain the integration boundary. Internal TypeScript
  interfaces may change to remove legacy coupling.
- This is implementation-support guidance. Core owners govern behavior;
  `docs/domain.md` owns vocabulary and owner navigation; `docs/design.md` governs
  presentation within its stated boundary. `docs/research/nlspec-spec.md` informs
  plan quality and does not make this handoff an adopted NLSpec or runtime input.

The prior authority map remains applicable. Directory cleanup additionally
follows Core 01 §3.3.5.3.1, REQ-01-168 for membership-derived visibility, query
members, ordering, and paging; REQ-01-581 for search semantics; and REQ-01-580
with Design §4.6 for authenticated-root selection. Core 03 §2.4 retains ordinary
workbook startup ownership. No copied implementation default overrides an owner.

### Confirmed findings and initial cleanup inventory

These are source observations at the planning baseline, not newly reproduced
product failures. IC-06 must distinguish its characterization results from the
prior iteration's retained evidence.

| Disposition | Evidence and planned treatment |
| --- | --- |
| REPLACE — directory wire shadows | App declares `IncidentListEnvelope`, builds incident-list URLs, casts successful payloads, and borrows user-list `PagingMeta`. `landingAdminTypes.ts` permits missing lifecycle fields, and IncidentLanding defaults absent status to `active`. Use the generated `listVisibleIncidents` operation, validated responses, and incident-list-derived resource/paging types. |
| REPLACE — permissive fixtures | `appShellTestSupport.ts` duplicates a permissive incident resource type and defaults missing status during filtering. Its builder already supplies lifecycle and attribution values. Make the fixture contract complete and update affected callers; do not repair malformed production responses with fixture defaults. |
| DELETE — unread credential storage | App discards the credential state's read value while retaining `setCredentialState` and repeated writes. Remove that unused storage and its casts; retain credential loading and error handling still consumed by account security. |
| REPLACE — shared query coordination | App resets `incidentDirectoryControlsTouchedRef` on shell refresh and pagination. Query drafts, the last accepted scope, rows, paging, and refresh status are distributed across state and refs. Move these responsibilities to one directory model. The historical handoff records suppressed debounced input; reproduce it before changing behavior. |
| REPLACE — pagination lifecycle | `handleLoadMoreIncidents` has no request abort signal, generation acceptance check, or synchronous duplicate-dispatch guard. Characterize stale, repeated, session, and navigation responses before replacing this path. |
| REPLACE — root discovery coupling | The two-row bootstrap list also checks whether the requested incident is visible. Delete count discovery under the approved stable-root contract; reproduce an explicit incident outside that probe. |
| DELETE — directory snapshot upserts | App's `upsertIncident` receives creation results and partial workbook snapshots through `handleIncidentSnapshot`. Directory rows should come from accepted list responses, with authoritative refresh on re-entry; remove the directory-only callback wiring and local upserts. |
| REMOVE AFTER PROOF — fixture workaround | The account-menu scenario in `workbook.visual.spec.ts` forces Enter because initial directory refresh may suppress debounce. Remove that workaround only after the ordinary debounced path passes deterministic regression coverage. |
| RETAIN — useful current behavior | Keep the bounded creation controller, same-ID replay, opening-only retry, memory-only recovery, local errors, focus rules, and session/navigation guards. Keep shared dialog styles: IncidentImportPanel still imports and uses them. |
| DEFER — unrelated cleanup | Broader shell/account/client modernization, backend incidents work, and unrelated workbook/admin behavior remain outside this iteration. Delete other styles, selectors, helpers, or exports only after confirming no remaining consumers. |

Inspected sources include App, IncidentLanding, the creation model/hook/form,
landing types/styles/layout, appShellClient and publicHttpTypes, browserApi,
generated HTTP operations/resources, workbook snapshot contracts, shared test
fixtures, App and creation tests, the IncidentDirectory browser helper, visual
sources, import boundaries, and authored owner catalogs. Generated sources were
read as projections and were not edited.

### IC-06–IC-10 tracker

| Workstream | Status | Depends on | Planned changes | Exit criteria |
| --- | --- | --- | --- | --- |
| IC-06 — Baseline and specification cleanup | DONE | Inspected baseline | Complete the delete/replace/retain/defer inventory with caller evidence. Characterize refresh-versus-edit, stale and duplicate pagination, session/navigation races, and explicit incident access outside the count probe. Resolve focused owner routing and amend stable-root owners/restatements. | Current root-navigation owners agree; each removal has a demonstrated reason; lifecycle defects have focused reproductions and owner-aligned expected outcomes. |
| IC-07 — Contract projections and directory adapter | DONE | IC-06 | Repair authored paging projections and regenerate; reconcile the candidate compatibility change set; add a narrow `listVisibleIncidents` adapter using generated operation/response validation. Derive resource and paging types from public contracts. Replace handwritten envelopes, raw list URL assembly, unchecked success casts, permissive lifecycle fields, and missing-status fallbacks. Update affected fixtures and callers directly. | Only validated directory responses enter state. Malformed lifecycle or paging data produces an explicit error. Fixtures supply complete resources without production compatibility defaults. |
| IC-08 — Establish directory lifecycle ownership | DONE | IC-07 | Introduce a dedicated directory model and React binding. Move query drafts, accepted query, rows, paging, loading/error state, cancellation, and response acceptance out of App. | Initial load, search, refresh, and pagination share one ownership model; obsolete responses cannot update current state. |
| IC-09 — Simplify application integration | DONE | IC-08 | Separate directory refresh from session/preferences/extensions bootstrap. Remove superseded query refs, unread credential storage, and directory upserts from partial workbook snapshots. Preserve creation's independent lifecycle and ordinary workbook startup; make root navigation consistently display the directory. | App composes directory and creation owners through explicit session/navigation boundaries. Directory contents come from accepted list responses, and query changes cannot redirect the application. |
| IC-10 — Verify and close | DONE | IC-09 | Update focused tests and authored routing; remove the forced-Enter visual workaround after debounce is proven. Complete deletion/caller audits and record validation, compatibility, and remaining limitations. | Retained behavior passes, new races are covered, and removed implementation paths have no remaining consumers. Evidence is for the implemented source, not inherited from IC-01–IC-05. |

### Implementation decisions and internal interfaces

Use a dedicated directory model and React binding with explicit operations for
query changes, immediate submission, refresh, pagination, and lifetime
invalidation. The model owns directory state and request acceptance; the binding
connects it to React; App supplies authentication and navigation boundaries.
Creation remains a separate operation owner. The narrow adapter consumes the
existing generated `listVisibleIncidents` operation through the established
public HTTP adapter boundary; it adds no public route or dependency. Shared pageable metadata makes existing
owner-required paging facts explicit in the generated types and validators.

- Preserve the existing 180 ms search debounce, explicit Enter submission, page size of
  100, and default status filter of `all`. Continue using server search, ordering,
  and cursor semantics rather than filtering or sorting a partial client list.
- A changed search or status immediately invalidates older page requests.
  Disable Load more while a replacement query is pending. Explicit Enter
  submits the current query and cancels its scheduled debounce. Status selection
  submits immediately. Duplicate submission of an identical in-flight query is
  a no-op. Bound directory request observation to 30 seconds without automatic retry.
- Retain previous authorized rows during refresh with clear status. Accept
  responses only for the current authentication lifetime and query/request
  generation. Bind each page request to its accepted query and cursor, and guard
  pagination synchronously against duplicate activation.
- Ordinary load failures remain local to the directory and offer retry. They
  must not become successful empty results. Session loss clears account-scoped
  state; access denial clears protected directory rows. Late responses after
  either transition cannot restore them.
- Authenticated `/` always displays the visible-incident directory for zero, one,
  or multiple incidents. Delete count probes, auto-open suppression, and the
  manual-directory history flag. Explicit incident routes use existing workbook
  authorization/startup; explicit directory selection and successful creation
  supply no launch `sheet_ref`. Validated authentication return targets remain valid.
- Leaving the directory cancels reads and clears materialized rows, cursors, and
  errors while retaining query drafts in the current session. Re-entry loads a
  fresh first page. Session replacement, termination, disposal, and reload clear
  account-scoped state. Reject stale failures before invoking session callbacks.
- Accumulate accepted live pages by incident identity, keeping first-seen position
  and the latest accepted payload. Do not sort partial client results. Invalid
  continuation requires explicit first-page refresh; transient page failures can
  retry their accepted query and cursor. Previous rows remain clearly labeled
  during replacement/failure; only accepted successful empty responses show an
  empty state.
- Query changes do not rerun account bootstrap or trigger automatic workbook
  navigation. Directory re-entry refreshes authoritative contents. Creation
  success and workbook snapshots do not insert synthetic directory rows or
  bypass the current query's server ordering and membership boundary.
- Preserve creation's immutable request, same-ID replay, observation timeout,
  collapse-and-retain behavior, session invalidation, and opening-only retry
  after confirmation. A directory refresh must not replace the creation draft,
  alter an unresolved attempt, or restore revoked automatic navigation.
- Remove App's unread credential storage while retaining credential loading and
  error handling used by account security. Update affected internal callers
  directly; add no compatibility aliases or duplicate execution paths.
- Keep shared dialog styles used by incident import. Do not delete a live
  feature because a historical name looks obsolete. Add no generic mutation
  framework, persistent recovery, or speculative extension registry.

The structural goal is to let later directory work extend its owner without
growing App's query/ref coordination or coupling creation to read requests.
The iteration deliberately changes root navigation and rejects malformed list
responses previously tolerated by the client. It makes no release-readiness claim
beyond its verified scope.

### Verification and completion boundaries

Update this tracker after completing every workstream and before starting the
next one. Record changed paths, gap dispositions, commands and run roots, failures,
skips, and exit evidence. Expected characterization failures are baseline evidence,
not successful product verification. IC-10 is the final validation/handoff slice.

Planning diagnostics passed `make help`, relevant owner task guides, the OpenAPI
target explanation, and working/staged whitespace checks. No planning product
verification or generation ran. Historical runs do not validate this iteration.

### Remediation and ownership register

| Gap | Remediation / areas | Benefit and unresolved risk | Compatibility / validation |
| --- | --- | --- | --- |
| G1 | Core 00/01/04, Design and current restatements: stable authenticated root; implementation and tests follow. | Stable entry supports directory growth; count-dependent routing otherwise changes context with membership. | Sole-incident users select explicitly; validate zero/one/many and provider parity. |
| G2 | Authored pageable metadata and paging constraints; generated validators/types and contract tests. | One executable projection prevents missing or contradictory continuation metadata. | Conforming wire responses unchanged; reject malformed metadata and reconcile candidate 2.0.0 change set. |
| G3 | Generated incident-list adapter/resources; complete fixtures and presentation. | Remove contract drift and lifecycle defaults that hide malformed resources. | Internal callers migrate directly; malformed data is an explicit local error. |
| G4 | Directory model/binding owns queries, request generations, rows and errors. | Atomic lifecycle prevents suppressed debounce and obsolete response commits. | Preserve search timing/defaults; test initial-load typing and duplicate submit. |
| G5 | Synchronous page lock, accepted-query/cursor binding and identity accumulation. | Prevent duplicate requests and mixed-page/account results while tolerating live overlap. | Disable continuation during replacement; test stale pages, overlaps and cursor recovery. |
| G6 | Local failure states, 30-second observation, access/session invalidation. | Avoid false empty results and restoration of protected rows by late responses. | Retry ordinary failures; test late success/error across session replacement. |
| G7 | Separate bootstrap and directory reads; remove count and history exceptions. | Explicit links no longer depend on partial list authorization; queries cannot navigate. | Routes unchanged; test deep links beyond page one and back/forward. |
| G8 | Remove creation/workbook directory upserts; refresh on re-entry. | Server membership, filtering and order remain authoritative. | Retained query can omit a newly created incident; workbook updates remain local. |
| G9 | Delete unread credential storage and proven dead consumers only. | Clarifies ownership without removing live account or import behavior. | Retain credential errors and shared import dialog styles; caller and regression audits. |
| G10 | Authored routing, deterministic debounce/browser evidence and final handoff. | Avoid certifying historical behavior or forced-Enter workarounds. | Update retired expectations without weakening assertions; retain final-source runs. |

### Execution record

IC-06 completed before IC-07 began. Changed Core 00/01/03/04, Design §4.6,
the current UI/UX guide and Appendix D illustrations, this handoff, the App
landing characterization suite, and its authored `web.application` routing.
Core 01 now requires the stable directory; Core 03 workbook startup is unchanged.
Domain vocabulary needs no amendment. G1 owner amendments are complete; G2–G10
remain implementation/validation obligations.

`make test-slice OWNER=web.application
ROWS=web.application.regression.incident_directory_lifecycle` retained
`.cartulary/test-results/20260905T175101Z-p26819`: five expected baseline failures
(duplicate pagination; stale pages after query, navigation/re-entry, and session
replacement; unnecessary directory requests on explicit incident routes). The
refresh-versus-edit case passes at baseline. The historical initial-refresh
suppression is not claimed reproduced by that case; final initial-load and browser
debounce tests still gate removal of the forced-Enter workaround. Earlier routing
attempts failed ASCII row ordering and exact-title selection; both authoring
errors were corrected. An ambiguous replacement-user test selector was corrected
before the retained baseline. These are characterization results, not conformance.

`make lint-markdown` passed at
`.cartulary/test-results/20260905T175014Z-p25767`; working and staged
`git diff --check` passed. The caller inventory confirms that App alone owns the
unread credential storage and directory upserts; WorkbookShell transports the
outbound callback through identity loading and incident controls. Incident import
still consumes `dialogBackdropStyle` and `dialogHeaderStyle`, so those remain.
Generated/protocol sources were inspected without editing. No release or broad
product verification was selected for this baseline/specification slice.

IC-07 completed before IC-08 began. G2/G3 now use authored `PagedEnvelopeMeta`
and bounded, consistent `PagingMeta`, regenerated OpenAPI/backend/protocol
projections, and the existing generated `listVisibleIncidents` operation. The
candidate 2.0.0 change set includes the five new projection differences; immutable
release history is unchanged. App's list casts/URL builder and permissive landing
type are removed. Fixtures now include request IDs, required resource fields and
valid incident UUIDs; presentation no longer invents lifecycle state. One G8
integration deletion moved forward: App's outbound workbook-snapshot receiver was
removed because a partial workbook identity cannot populate the complete list
resource. The remaining callback transport is still scheduled for IC-09.

Validation:

- `make generate`: PASS, `20260905T175603Z-p30603`. The first attempt stopped
  at the expected missing compatibility entries (`20260905T175231Z-p27823`);
  `make openapi-compatibility-check` supplied the exact report, and the current
  candidate change set was reconciled without changing the immutable baseline.
- `make openapi-compatibility-check`: PASS, `20260905T175820Z-p47756`.
- `make frontend-typecheck`: PASS, `20260905T175655Z-p38250`.
- `make test-slice OWNER=package.protocol_ts`: PASS, `20260905T175655Z-p38114`.
- `make test-slice OWNER=platform.openapi`: PASS, `20260905T175655Z-p38146`.
- Incident OpenAPI/lifecycle contract rows: PASS, `20260905T175734Z-p46333`.
- Directory adapter and existing search/paging App rows: PASS,
  `20260905T175752Z-p46831`. The adapter's initial fixture reused a consumed
  Response body; returning a fresh Response corrected that test-only failure.
- `make format`: PASS, `20260905T175632Z-p33681`; `git diff --check`: PASS.

Run roots above are under `.cartulary/test-results/`. New paging negative cases
reject absent/contradictory metadata; complete terminal and nonterminal responses
pass. Existing query and page scope behavior remains until IC-08/IC-09. Browser
and broader final verification are deferred to IC-10; no conformance completion
or release claim is made at this boundary.

IC-08 completed before IC-09 began. Added `incidentDirectoryModel.ts`,
`useIncidentDirectory.ts`, their operation/binding tests, and authored
`web.application` rows. The controller owns query drafts and accepted queries,
all request/timer lifetimes, synchronous continuation locking, identity-based
live-page accumulation, local retries, invalid-cursor refresh, access denial,
session loss, and disposal. The binding uses the existing HTTP adapter and React
external-store pattern; it is not yet connected to App.

Seventeen operation cases pass in `20260905T180635Z-p56101`; the two binding
cases pass in `20260905T180654Z-p57239`. The initial type check exposed a test
helper whose default narrowed paging to terminal pages; the helper now explicitly
accepts the generated paging union. The Strict Mode test now enables React's
root-level strict replay through Testing Library's `reactStrictMode` option.
`make frontend-typecheck` passes at `20260905T180644Z-p56821`, and
`make frontend-import-boundary-check` passes at `20260905T180541Z-p54657`.
`make generate` passes at `20260905T180716Z-p57852`, including authored routing
projection updates; `git diff --check` passes. G4–G6 controller behavior is covered
independently; App acceptance of those changes and the IC-06 failures remain
IC-09 obligations. Browser/visual and broad final checks remain IC-10 work.

### Required implementation verification

- Model and integration cases for typing during initial refresh, rapid query
  changes, Enter/debounce deduplication, duplicate Load more, late page
  responses, navigation, disposal, session replacement, malformed responses,
  ordinary failure/retry, and clearing protected rows on access loss.
- Root behavior with zero, one, and multiple visible incidents; explicit
  directory navigation; and an explicit incident outside the count probe.
- Existing creation keyboard, validation, replay, membership-refresh failure,
  account navigation, focus, accessibility, and visual regressions. Confirm
  directory activity cannot disturb creation's draft or operation identity.
- Focused owner slices through `web.application` and `module.incidents`, with
  relevant `web.design` and `package.ui` rows when affected. Choose rows through
  the current task guides and owner catalogs. Update authored routing and use
  Make generation only when implementation changes require it; never hand-edit
  generated projections or weaken assertions to preserve old behavior.
- `make frontend-typecheck`, `make frontend-import-boundary-check`,
  `make lint-biome`, and `make build-web`. Run `make agent-finalize` before broader
  implementation verification and record retained-run skips when `RESULTS_DIR`
  is unset. Visual maintenance follows the existing guide when needed.

IC-10 must record actual changed paths, deletion/caller-audit results, commands
and retained run roots, related or unrelated failures, and skipped checks with
reasons. Record compatibility and rollback for the implemented frontend unit;
the plan requires no backend or stored-data migration. Do not mark the iteration
complete or production-ready solely from this plan or the prior successful UI
evidence.

### IC-09 completion — application integration and deletion

IC-09 completed after the IC-08 tracker update. App now composes directory and
creation controllers at explicit authentication/navigation boundaries. Session,
credential, preferences, and extension bootstrap no longer reads the incident
list. Directory query/refresh/re-entry does not rerun bootstrap. Root is stable
for every membership count; local sign-in preserves an explicit workbook URL,
and explicit workbook access is delegated to workbook authorization/startup.
Known sign-out or credential replacement clears both controllers synchronously.
Workbook access loss removes its protected materialization, replaces the route
with `/`, refreshes the retained directory query, and focuses its heading.

Changed paths: `apps/web/src/app/{App.tsx,IncidentLanding.tsx,landingAdminTypes.ts,
routeState.ts,routeState.test.ts,App.landing.test.tsx,App.auth.test.tsx,
IncidentAdminPanel.tsx,IncidentAdminPanel.test.tsx}`;
`apps/web/src/shared/workbookShellContracts.ts`;
`apps/web/src/workbook/{WorkbookShell.tsx,components/WorkbookIncidentControlsPresentation.tsx,
hooks/useWorkbookIncidentIdentity.ts}`; the directory model test import;
`tools/test_families/{web.application,module.incidents}.json`; generated topology
render index; and this tracker.

G1/G4–G9 implementation dispositions are closed: removed the count probe,
cardinality routing, manual-directory history state, suppression ref, distributed
query/paging handlers, creation upsert, unused credential storage, and the entire
directory-only workbook snapshot callback chain/type. Caller searches found no
remaining production consumers of these deleted interfaces. The independent
debug authentication harness retains its *read* credential state. IncidentImportPanel
still consumes shared dialog styles. IncidentAdminPanel's lifecycle regression
checks its own updated presentation after callback deletion. Directory failures
have local retry and no successful-empty fallback; previous results are labeled.

Validation and exit evidence (run roots under `.cartulary/test-results/`):

- App landing/route focused rows: PASS 35/35 units,
  `20260905T182222Z-p83066`. Final lifecycle/model/binding rows: PASS 4/4,
  `20260905T182450Z-p98834`, including 11 App lifecycle cases, 17 model cases,
  and 2 binding cases. The original six characterization cases all pass.
- Auth/account focused rows: PASS 7/7, `20260905T182450Z-p98845`.
- Incident creation/controls focused rows: PASS 7/7,
  `20260905T182222Z-p83076`.
- Workbook startup adapter and shell surfaces rows: PASS 3/3,
  `20260905T182237Z-p87554`.
- `make frontend-typecheck`: PASS `20260905T182450Z-p98958`.
  `make format`: PASS `20260905T182440Z-p94506`.
  `make generate`: PASS `20260905T182451Z-p99560`.
  Working/staged `git diff --check`: PASS.
- Intermediate failures: unused test imports and an established session-loader
  union required restoring its existing typed extraction (first typecheck
  `20260905T181725Z-p67852`); obsolete exact-title routing and unsorted appended
  titles blocked catalog validation until corrected. Focused runs
  `20260905T181926Z-p69925`, `20260905T181927Z-p70181`,
  `20260905T181928Z-p70715`, and `20260905T182222Z-p83067` identified old count
  expectations, a creation test incorrectly using directory refresh to replace
  authentication, an auth assertion preceding refresh completion, and historical
  shared-landing creation-error expectations. Tests now exercise explicit account
  replacement, wait for completed auth state, and verify associated creation field
  errors plus private-detail non-disclosure. These were corrected, not waived.

No IC-09 exit failures remain. Browser, accessibility, visual, broader static and
contract gates, final-source verification, and the workaround removal belong to
IC-10 and have not been claimed complete. No persistence migration, new public
route, dependency, or release-baseline edit was introduced.

### IC-10 validation findings and corrections

The first broad frontend run (`20260905T182821Z-p4136`, 445/448 units)
identified three maintenance gaps: the source ownership manifest omitted the
four new directory files and four previously committed creation files; the
shared import test expected a downstream paging error for envelopes now rejected
by generated validation; and historical focus tests used prohibited heading-name
selectors. The manifest now accounts for all eight paths, import validation
expects the generated boundary error while retaining cursor-cycle coverage, and
a shared heading selector preserves both accessible-name and focus assertions.
Focused architecture, App/import, and UI selector checks passed at
`20260905T183431Z-p65762`, `20260905T183432Z-p67302`, and
`20260905T183433Z-p69556`. Broad frontend units subsequently passed 448/448 at
`20260905T183548Z-p24694` before the late navigation fix below.

Incident browser coverage passed all eight scenarios (13/13 execution units),
`20260905T182728Z-p9139`. Auth browser runs
`20260905T182729Z-p9376` and `20260905T183453Z-p75590` exposed an obsolete
expectation that directory Refresh requests auth/session, an incomplete mocked
public denial, and then a real late-workbook navigation race. The first two test
expectations were corrected. The retiring workbook could restore incident_id
after App had navigated to the directory. Workbook sheet-URL writes now require
the currently selected incident route; stale access-loss callbacks cannot replace
a newer route. App context changes clear prior incident sheet parameters, while
local sign-in preserves actual explicit view-schema URL parameters. These are
immediate navigation-boundary fixes under G7/G8, not changes to the startup
selection algorithm. Tests prove that retiring workbook effects cannot change
root, another incident, or deployment-administration URLs. Workbook fixtures now
establish their selected incident route instead of relying on a child workbook
to select App context. Focused navigation and final-source rerun results are recorded below.

The initial accessibility run (`20260905T182821Z-p4480`) found two old sole-incident
auto-opening assertions. The revised interaction explicitly selects a workbook
and checks directory heading/focus after access loss. All accessibility checks
then passed 12/12 units at `20260905T183548Z-p25013` before the late navigation fix.

The initial visual run (`20260905T182821Z-p4280`) passed all functional interactions
but found 756 changed pixels in account-menu-directory-root: an obsolete return
notice obscured the accepted query status. The duplicate return handler was
removed. Ordinary debounce now drives the visual scenario without Enter. A fresh
ordinary visual run passed 12/12 units at `20260905T183548Z-p24814` before the late
navigation fix. Reconciliation accounts for 110 active captures/goldens, zero
orphans or missing/ambiguous mappings, and all 26 registered fixtures. No golden,
mask, viewport, renderer, or screenshot threshold has changed.

`make agent-finalize` passed at `20260905T182730Z-p10083` and
`20260905T183452Z-p75386`. Retained-run maintenance was explicitly skipped because
RESULTS_DIR was unset. Contract compatibility, generation drift, generated policy,
frontend typecheck/import boundaries, Biome, build-web, and Markdown have passed;
Final-source results and the full changed-path/compatibility/rollback record
follow below.

### Final evidence and handoff

IC-10 completed on 2026-09-05. All required in-scope gates passed; no exit
failure remains. The final tracker update closes this effort.

The final integration retains the incident-specific controllers, existing API,
and server-owned membership/search/order/cursor semantics. G1–G10 have implemented
remediations and final verification evidence below. The only additional changes
found necessary during validation were source/selector accounting, affected
fixture expectations, and the immediate workbook/App navigation boundary fixes.
The root/count, lifecycle, error, creation recovery, import, account security,
workbook controls, explicit startup, and browser focus scenarios are covered.
No product test, generator, conformance, or release input was made dependent on
Markdown. `docs/domain.md` was reviewed for vocabulary/ownership and needs no edit.

Final-source runs are under `.cartulary/test-results/`:

| Command / selected scope | Result | Run ID |
| --- | --- | --- |
| `make agent-finalize` before final broad gates | PASS; retained maintenance skipped, RESULTS_DIR unset | `20260905T184308Z-p93469` |
| `make frontend-unit` (all frontend owners, including protocol, UI, App, workbook, architecture and import regressions) | PASS 448/448 units | `20260905T184403Z-p42646` |
| `make service-backed-test-slice OWNER=module.auth` (11 authored webserver-backed scenarios) | PASS 13/13 units | `20260905T184306Z-p93241` |
| `make service-backed-test-slice OWNER=module.incidents` (8 authored webserver-backed scenarios) | PASS 13/13 units | `20260905T184403Z-p42497` |
| `make browser-e2e-a11y` (full applicable scope) | PASS 12/12 units | `20260905T184403Z-p43048` |
| `make browser-e2e-visual` (full applicable scope) | PASS 12/12 units | `20260905T184403Z-p42820` |
| `make test-slice OWNER=platform.openapi` | PASS 4/4 units | `20260905T184421Z-p91826` |
| `make test-slice OWNER=module.incidents` (lifecycle/request/OpenAPI and response/security parity rows) | PASS 1/1 combined unit | `20260905T184603Z-p26978` |
| `make openapi-compatibility-check` | PASS 4/4 units | `20260905T184403Z-p42438` |
| `make generate-drift` | PASS 4/4 units | `20260905T184403Z-p42359` |
| `make generated-artifact-policy-check` | PASS 3/3 units | `20260905T184403Z-p42380` |
| `make frontend-typecheck` | PASS 2/2 units | `20260905T184403Z-p42683` |
| `make frontend-import-boundary-check` | PASS 2/2 units | `20260905T184403Z-p42706` |
| `make lint-biome` | PASS 2/2 units | `20260905T184403Z-p42744` |
| `make build-web` | PASS | `20260905T184403Z-p43095` |
| `make lint-markdown` after handoff assembly | PASS | `20260905T184915Z-p35317` |
| Working/staged `git diff --check` | PASS; repeated at final handoff | Local Git check |

Earlier failures are fully described above. The late workbook-boundary focused
run `20260905T183829Z-p70636` additionally exposed workbook tests whose fixtures
had no selected incident URL and a missing test cleanup. Those fixtures now
establish real navigation context, cleanup is explicit, and the same focused
workbook rows passed 3/3 at `20260905T184204Z-p88324`. Final frontend verification
includes those fixes. No failing product assertion was waived or weakened.

Compatibility and rollout: deliver this entire integration unit together:
specification amendments, authored projections, candidate compatibility metadata,
generated artifacts, frontend models/bindings/integration, tests, catalogs, and
handoff. Authenticated root no longer auto-opens a sole visible incident; users
select it. Explicit incident URLs (including actual startup query parameters)
survive sign-in. Directory selection and creation start a fresh workbook context.
Malformed incident responses and inconsistent common paging are rejected at the
generated boundary. Ordinary singleton metadata remains independently usable.
Query and creation recovery remain memory-only and disappear on account lifetime
end/disposal/reload. There is no stored-data or endpoint migration, dependency
change, backend feature, release publication, or immutable baseline alteration.
The existing unpublished `2.0.0` change set includes the projection correction.

Rollback: restore the integration unit as a whole, including its specifications,
projection sources, generated outputs, candidate compatibility metadata, frontend,
tests, source/verification catalogs, and handoff changes. Preserve unrelated work
and the original staged handoff/historical record. No database rollback or golden
restoration is needed: this iteration changes neither stored data nor goldens.

Deletion audit: no production caller remains for `manualIncidentDirectory`,
`cartularyIncidentDirectory`, the two-row directory probe, auto-open suppression,
`upsertIncident`, `onIncidentSnapshot`, or `WorkbookIncidentSnapshot`. Only a route
test retains an old history flag to prove it is ignored. App's unread credential
storage is gone; the debug harness retains its live read state. Incident import
retains shared dialog styles and passed browser/frontend regressions. Workbook
controls retain their local updated resource presentation. The original staged
IC-01–IC-05 historical record is byte-identical; no new changes were staged.

Check omissions and limits: retained-run cache/evidence maintenance was skipped
because RESULTS_DIR was unset. No new goldens were accepted, so visual-update and
its two post-update runs were unnecessary. Full backend/CI/release/security-scan,
stateful-browser, and performance/measurement suites were not run: this bounded
frontend/projection change is covered by the selected API/browser and full
frontend/a11y/visual/static gates. This is no new release-readiness claim. No
in-scope issue is deferred; future backend features and persistence remain out
of scope.

Actual changed paths (including new files; no generated file was hand-edited):

- `apps/web/e2e/auth-and-incident-directory.spec.ts`
- `apps/web/e2e/enterprise-auth.spec.ts`
- `apps/web/e2e/incident-administration.spec.ts`
- `apps/web/e2e/pages/incidentDirectory.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/src/app/App.auth.test.tsx`
- `apps/web/src/app/App.landing.test.tsx`
- `apps/web/src/app/App.tsx`
- `apps/web/src/app/IncidentAdminPanel.test.tsx`
- `apps/web/src/app/IncidentAdminPanel.tsx`
- `apps/web/src/app/IncidentLanding.tsx`
- `apps/web/src/app/LandingAdminLayout.tsx`
- `apps/web/src/app/api/appShellClient.ts`
- `apps/web/src/app/api/incidentDirectoryClient.test.ts`
- `apps/web/src/app/api/publicHttpTypes.ts`
- `apps/web/src/app/incidentDirectoryModel.test.ts`
- `apps/web/src/app/incidentDirectoryModel.ts`
- `apps/web/src/app/landingAdminTypes.ts`
- `apps/web/src/app/routeState.test.ts`
- `apps/web/src/app/routeState.ts`
- `apps/web/src/app/useIncidentDirectory.ts`
- `apps/web/src/imports/importCoordinator.test.ts`
- `apps/web/src/shared/workbookShellContracts.ts`
- `apps/web/src/testing/appShellTestSupport.ts`
- `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/components/WorkbookIncidentControlsPresentation.tsx`
- `apps/web/src/workbook/hooks/useWorkbookIncidentIdentity.ts`
- `apps/web/src/workbook/hooks/useWorkbookStartupController.test.tsx`
- `apps/web/src/workbook/hooks/useWorkbookStartupController.ts`
- `contracts/openapi-releases/2.0.0.change-set.json`
- `contracts/openapi-source/owners/module.incidents/openapi.json`
- `contracts/openapi-source/owners/platform.openapi/openapi.json`
- `contracts/openapi/cartulary.openapi.yaml`
- `docs/design.md`
- `docs/guides/cartulary-ui-ux-design-guide.md`
- `docs/handoffs/incident-directory-creation-refactor-handoff.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/spec/D_workflow_and_ui_illustrations_source_extract.md`
- `internal/gen/contractopenapi/artifacts_gen.go`
- `internal/gen/openapioperations/catalog_gen.go`
- `packages/protocol-ts/src/generated/core-http-types.ts`
- `packages/protocol-ts/src/generated/core-http-validators.ts`
- `packages/protocol-ts/src/index.test.ts`
- `packages/ui-contracts/src/application-selectors.test.ts`
- `packages/ui-contracts/src/applicationSelectors.ts`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/test_families/module.incidents.json`
- `tools/test_families/package.protocol_ts.json`
- `tools/test_families/web.application.json`
- `tools/test_families/web.workbook.json`

## Historical record — completed IC-01–IC-05

Everything below records the prior creation iteration as it stood before its
commit at the planning baseline above. Its branch/HEAD references, working-tree
statements, scope exclusions, next-action text, command results, rollback unit,
and completion/authorization statements apply only to that historical iteration.
In particular, its exclusion of directory query redesign does not constrain the
new IC-06–IC-10 plan. Its retained runs are historical evidence and do not certify
the cleanup iteration. The record is preserved with subsection headings nested
under this historical boundary.

### Baseline and scope

- Branch: `main`; HEAD: `3972662b1ca2f994eec4ae453cc3f231de4cb6e1`.
- Working tree clean at implementation start; no reset, commit, or push.
- One seam: directory creation through `POST /api/v1/incidents`.
- Allowed: creation model/controller, presentation, narrow typed client, App
  integration, related styles/selectors/tests/routing/visual evidence, this handoff.
- Excluded: backend, public contracts, authorization, routes, dependencies,
  persistence, directory query redesign, account/admin forms, adopted owners,
  digest sources, and completed workbook/grid-row/menu/inspector work.

### Tracker

| Workstream | Status | Exit evidence / next action |
| --- | --- | --- |
| IC-01 — Baseline, authority, characterization | DONE | Five focused characterization failures reproduce keyboard, duplicate, local-error, replay, and navigation gaps. |
| IC-02 — Creation model and typed operation lifecycle | DONE | Fourteen focused operation/client tests and frontend typecheck pass. |
| IC-03 — Same-surface form and application handoff | DONE | Inline form, nine App lifecycle cases, baseline metadata/bootstrap checks, typecheck/import boundaries, and real browser creation pass. |
| IC-04 — Browser, accessibility, and visual evidence | DONE | Four creation/browser rows, full a11y, reviewed eight goldens, and two final-source ordinary visual passes. |
| IC-05 — Final validation and completed handoff | DONE | A001–A027 dispositions, compatibility/atomic rollback, final Make checks, Markdown lint and 32-path scope audit complete. |

### Authority and inspected evidence

| Behavior | Owner |
| --- | --- |
| Root visibility and ordinary workbook selection | Core 01 §3.3.2.1A, REQ-01-580; Core 03 §2.4 |
| Active authenticated creator; cookie CSRF | Core 01 REQ-01-154–155; Core 04 REQ-04-006, 008–012, 015–016, 023 |
| Resource and closed create payload | Core 01 §3.3.5.3, REQ-01-152–159 |
| Normalization, errors, actor-scoped replay | Core 01 REQ-01-157, 160–162, 166–167; §3.3.6.2 |
| Atomic incident/admin/preferences bootstrap | Core 01 REQ-01-163–165 |
| Optional string semantics and canonical TLP | Core 01 §18, REQ-01-488, 491, 491.1; Core 02 §18, REQ-02-222/254 |
| Local feedback, keyboard/focus, dialog restrictions | Design §§5, 8.4–8.5, 10.8, 12.4–12.8, 14 |
| Responsive composition and visual evidence | Design §§7, 15; visual golden maintenance guide |

Read root AGENTS.md (no nested instructions), digest README and localized read
order including maps/rules/acceptance, domain vocabulary boundary, relevant Core
and design owners, and completed account-application-menu handoff as evidence.
No adopted-owner contradiction found. Digest guidance is advisory; executable
code and evidence do not consume Markdown.

Inspected IncidentLanding, App, landingAdminTypes/Styles, LandingAdminDisplay,
appShellClient, browserApi/httpTransport, generated createIncident binding and
request/response types, shared authorization/registered focus, App.landing tests,
IncidentDirectory browser helper, auth-and-incident-directory browser checks,
accessibility/visual sources, and authored owner catalogs.

Current stack revalidated: React 19, TypeScript 6, Vite 8, pnpm workspace,
Vitest/Testing Library, Playwright, existing tokens and Lucide. Generated roots
remain governed by tools/generated_artifact_policy.json.

### Interaction decisions

| State / action | Draft and operation | Focus and navigation |
| --- | --- | --- |
| Open / edit | Required key/title; optional existing disclosure | Explicit open focuses key; ordinary Tab order. |
| Local or definitive field rejection | Retain exact draft; reveal affected optional field | Local associated errors; no response-driven focus move. |
| Submit | Capture declared payload and Web Crypto ID once; synchronous dispatch lock | Native submit; textarea Enter remains editing. |
| Pending / uncertain | Read-only captured request; manual replay uses same ID/body | Account navigation stays reachable. |
| Close / Escape | Collapse, retain memory-only draft/attempt, expose resume status | Return to trigger; revoke automatic handoff. |
| Reopen | Resume existing state, not a new attempt | Focus relevant error/replay/open action; reopening does not restore automatic handoff. |
| Navigate away / Account settings | Retain same-session controller; revoke automatic handoff | Later responses do not replace chosen destination. |
| Confirmed create / handoff failure | Creation stays committed; retry opens only | Ordinary incident_id startup; no explicit sheet. |
| Session termination / account replacement | Clear account-scoped draft/attempt and invalidate callbacks | No stale focus or navigation. |

User selected collapse-and-retain and replay-before-editing. No persistent
recovery, background replay, global mutation framework, or workbook queue reuse.

### Verification routing and initial evidence

Resolved `make help`, `make help-all`, and `make task-guide ROLE=module-author`
for `module.incidents`, `web.application`, and `web.design`.

- Incident creation behavior: module.incidents frontend/browser rows; App and
  narrow integration regressions retain web.application routing.
- Presentation/accessibility/visual readiness: web.design; selectors: package.ui.
- Optional metadata baseline passed: `20260905T143008Z-p1009`.
- App creation/membership/incident_id baseline passed: `20260905T143008Z-p1005`.
- Service-backed directory creation/admin/workbook baseline passed:
  `20260905T143029Z-p2102`.

These are exact-source baseline checks, not evidence that remediation is done.
Previous menu handoff failures are historical evidence only; this seam's three
selected baseline checks pass at current HEAD.

### IC-01 activity

Initial inspection finds a non-form modal, seven App draft variables, component
open state, raw fetch/success cast, no synchronous dispatch guard, a new ID per
call, global creation errors, and unguarded response navigation. Characterization
tests will distinguish reproduced gaps from these source observations.

Characterization command: `make test-slice OWNER=module.incidents
ROWS=module.incidents.frontend.incident_creation_lifecycle_keyboard_replay_and_e99c1bedb1`.
Final baseline run: `20260905T150136Z-p55051`, expected product assertion failure.
Native Enter sent zero requests; same-tick activation sent two; required errors
had no accessible association; uncertain recovery had no replay action; delayed
success changed the route despite chosen Account settings. The last test now
waits for response parsing and checks the published URL, removing an initially
premature assertion hidden by lazy workbook rendering. No production
source changed before these failures were captured. Exact-title/sort catalog
authoring errors were corrected before the retained test runs.

IC-01 exit: owner/scope/state matrix, successful baseline, reproduced gaps, and
owner routes recorded. Next action: implement and test the bounded controller.

### IC-02 activity and exit

Added incidentCreationModel.ts and focused tests; added the narrow generated
createIncident operation wrapper to appShellClient. The controller captures one
immutable allowed request and ID before dispatch, owns a synchronous observation
guard and 30-second timeout, preserves uncertain replay, and separates confirmed
creation from membership/workbook handoff. Session and navigation generations
guard callbacks; no queue or persistent storage is used.

`make test-slice OWNER=module.incidents
ROWS=module.incidents.frontend.incident_creation_operation_payload_outcomes_and_2ecdf8c632`
passes all 14 cases at `20260905T151131Z-p59594`. `make frontend-typecheck`
passes at `20260905T151047Z-p59028`. Its earlier failure at
`20260905T150819Z-p57928` caught exact-optional-property typing and one unused
import; both were fixed. Generated output remains untouched. Next action:
wire the controller and same-surface form to App.

### IC-03 activity and exit

IncidentCreationForm and useIncidentCreation now replace the creation modal,
seven App draft variables, and global creation notices. Native Enter submission,
textarea editing, local associated errors, optional-error disclosure, explicit
focus return, retained close/reopen, and read-only replay use the controller.
App owns current-session replacement, membership refresh, and incident_id startup;
navigation and Account settings immediately invalidate automatic handoff.
Confirmed handoff observation is also bounded, without another create request.

The generated boundary correctly rejected incomplete historical test fixtures;
appShellTestSupport now supplies complete attribution/lifecycle resource fields.
Form errors were moved outside labels so error text does not change control names.
An import-boundary failure (`20260905T152323Z-p73648`) caught a direct protocol
import and a type cycle; model types now use publicHttpTypes and presentation
bindings live in the model, preserving the existing adapter boundary.

Passing evidence:

- Nine App lifecycle cases plus existing create/bootstrap row:
  `20260905T152202Z-p72206`.
- Existing optional metadata row: `20260905T152349Z-p74108`.
- Selector contract: `20260905T152354Z-p74642`.
- Typecheck: `20260905T152312Z-p73156`.
- Import boundaries: `20260905T152442Z-p75447`.
- Format: `20260905T152446Z-p75876`; only seam files were rewritten.
- Existing real-browser directory create/admin/workbook row:
  `20260905T152518Z-p80212`.

Next action: add deterministic browser recovery, accessibility, geometry, and
visual capture evidence; retain completed-seam goldens unless a reviewed change
is directly caused by this creation seam.

### IC-04 activity

Added three service-backed browser scenarios covering native keyboard creation,
real optional/key rejection, exact payload and server normalization, duplicate
activation, server commit followed by a lost client response, identical 200 replay,
confirmed creation followed by membership-refresh failure, opening-only retry,
and delayed success after Account settings navigation. All pass at
`20260905T153821Z-p37684` via the three new module.incidents browser rows.
The existing IncidentDirectory helper now opens/resumes the inline form and only
starts another draft explicitly after a confirmed result.

The focused web.design accessibility row passes at
`20260905T154530Z-p35302`. It checks keyboard/name/error associations, optional
error disclosure, restrained announcements, pending/recovery reachability at
1280×720, 1024×720, 768×640, and 640×480, 200% CSS zoom, text spacing, and safe
account navigation. Its first run (`20260905T154024Z-p83887`) failed at a test
focus operation after simulated zoom: focusing an already-active control did not
scroll it back into view. The scenario now uses actual Shift+Tab/Tab to re-enter
that control; all visibility and overflow assertions remain. Directory
aria-busy now belongs to the independent query region, so query refresh cannot
defer creation announcements.

Eight deterministic visual captures cover initial, required error, expanded
details, pending, short recovery, zoom, text spacing, and confirmed handoff
failure. Ordinary reconciliation and image review precede promotion. Routing
inputs are authored in module.incidents.json and web.design.json; `make generate`
passed at `20260905T153724Z-p34459`, updating only browser batch and topology
render projections. No stable digest or registry fixture IDs were invented.

`make agent-finalize` passed before broad visual verification at
`20260905T155150Z-p82048`; generated maintenance was unchanged. RESULTS_DIR was
unset because no qualifying exact-source full warm check exists. Canonical
retained-run evidence, scheduler ordering/timing, and performance maintenance
were therefore skipped, as recorded in unit-artifacts/finalize-summary.json.

First full ordinary visual run `20260905T155217Z-p85214`: all existing visual
tests passed and the new scenario reached all eight captures. It failed on the
eight expected missing initial goldens, plus reconciliation rejected the new
authored visual scenario ID because that artifact requires a hexadecimal ID.
Corrected only the new visual routing input and regenerated through Make
(`20260905T155608Z-p33293`). This was a fixture/routing defect, not a product or
unrelated baseline failure; no images were promoted by the failed run.

Ordinary run `20260905T155640Z-p36382` now produces the required v2
reconciliation: 110 capture intents, 102 active existing goldens, eight new
missing goldens, 26 resolved registered fixtures, zero orphans and zero ambiguous
mappings. All functional assertions and existing comparisons pass. The accepted
trigger is the authorized same-surface form and operation-state presentation;
the canonical updater adds eight initial goldens, without moving, deleting, or
re-accounting existing captures. The new row is
`web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde`,
scenario `scenario_46ea092e7f2b`, project chromium. It is an active nonregistry
capture with no invented D-VFIX identity. Existing renderer, fonts, masks,
registered fixture IDs, and screenshot scope remain unchanged.

Update attempt `20260905T155940Z-p83937` did not promote files. Its creation
scenario passed and produced all eight candidates; an unchanged menu fixture
timed out waiting for two filtered rows (31 remained), and the retained-artifact
scanner rejected a copied existing PNG's permissions. The menu fixture passed
both preceding ordinary runs; it remains unchanged. The permission failure is
also documented in the completed menu handoff. Restricted all 102 source PNGs to
0600 with SHA-256 equality verified, then reran the canonical updater; no harness
check was bypassed and no tracked file content changed for this maintenance.

Individually reviewed all eight creation candidate images from that run:

| Filename under workbook.visual.spec.ts-snapshots | Viewport / zoom | Review |
| --- | --- | --- |
| incident-create-initial-linux.png | 1280×720 / 100% | Compact inline form, key focus visible, account trigger reachable. |
| incident-create-required-errors-linux.png | 1280×720 / 100% | Associated errors beneath fields; single operation message. |
| incident-create-expanded-details-linux.png | 1024×720 / 100% | All optional fields and action fit; long inputs retain native text scrolling. |
| incident-create-pending-linux.png | 768×640 / 100% | Pending label, close semantics, and read-only fields remain visible. |
| incident-create-recovery-short-linux.png | 640×480 / 100% | Recovery text and focused retry fit above directory controls. |
| incident-create-recovery-zoom-linux.png | 1280×720 / 200% CSS zoom | Native focus scroll reveals recovery and retry; surrounding content remains scrollable. |
| incident-create-recovery-text-spacing-linux.png | 768×640 / 100% | Expanded spacing wraps feedback; retry remains contained and focused. |
| incident-create-confirmed-handoff-failure-linux.png | 1280×720 / 100% | Committed outcome is explicit with opening-only retry. |

Decision for each: accept, classified as accepted intentional state under Design
§§7, 12.4–12.8, 14–15. These are implementation/design evidence, not Core 05
claim-publication or release-readiness evidence. No unrelated golden change is
accepted. Two fresh ordinary passes remain required after promotion.

The second update (`20260905T160444Z-p32056`) passed private-artifact checks but
reproduced the same existing menu filter assertion. Its trace contains only
unfiltered directory requests: initial refresh completes after the fixture fills
search and before debounced search dispatch. The directory query/touched-state
implementation is unchanged by this seam. Added an explicit native Enter and
response barrier to that fixture's existing filter step, retaining its two-row
assertion, inputs, menu behavior, and all screenshot settings. This is a bounded
fixture synchronization repair needed by the shared visual gate, not a menu or
directory-query refactor. No failed update promoted tracked evidence.

Canonical refresh passes at `20260905T160935Z-p83842` (12/12 units). All eight
creation goldens were reviewed; six are byte-identical to the earlier individual
review and the two with small border-rasterization differences were inspected
again. The updater also changed four unrelated existing images (presence marker
ordering, entity chip edge, network inspector input edge, and pending replay
toolbar). Their old/new images were inspected and rejected from this seam;
restored their exact HEAD bytes. All 102 pre-existing PNGs now match HEAD by
byte comparison, including every menu golden. `make generate` regenerated the
manifest at `20260905T161402Z-p32475`. No renderer, dependency, threshold, or
registered fixture changed.

Ordinary visual pass `20260905T161437Z-p38568` succeeds (12/12 units, 110
captures). Final-source typecheck (`20260905T161507Z-p81324`), import boundaries
(`20260905T161507Z-p81341`), and Biome (`20260905T161507Z-p81367`) pass. Final
accessibility review then removed aria-hidden from the visible pending suffix
and added an accessible-name assertion: sighted and screen-reader users now
receive the same pending button label. No pixels changed, but two fresh ordinary
visual runs will still be retained after this final presentation-source edit.

After that edit, `make agent-finalize` passes again at
`20260905T161807Z-p91893` with RESULTS_DIR unset and the same documented
retained-run skips. Full `make browser-e2e-a11y` passes at
`20260905T161851Z-p95175` (12/12 units), including the new pending-name assertion.
Creation/controller/bootstrap frontend rows pass at `20260905T161912Z-p38572`;
the complete web.application owner slice passes at `20260905T161912Z-p38580`
(58/58 units), covering unchanged menu/session/route integration; selector
contracts pass at `20260905T161912Z-p38593`.

Two fresh ordinary visual passes on the final presentation source:
`20260905T162101Z-p51590` and `20260905T162445Z-p98917`, each 12/12 units.
Both reconcile all 110 active goldens and all 26 registered fixtures, with zero
missing, orphaned, ambiguous, or unresolved fixture mappings. Final service-backed
verification passes at `20260905T162914Z-p46655` (11/11 units): the three new
creation scenarios plus the unchanged ordinary create/admin/bootstrap scenario.

IC-04 exit: keyboard, actual public validation, committed-response loss and 200
replay, no duplicate recreation, chosen navigation, full accessibility, and
reviewed constrained visual evidence pass. No blocked owner dependency remains.
Next action: finish IC-05 acceptance, compatibility, rollback, and final audits.

### IC-05 architecture and changed paths

Before: modal ceremony, click-only dispatch, scattered draft/open state, global
creation errors, unchecked success, and fresh transaction IDs after ambiguous
failure. After: one native inline form renders one account-scoped controller;
the controller owns draft/disclosure, explicit outcome state, captured request,
synchronous dispatch lock, observation timeout, and lifetime guards. The typed
adapter accepts the generated public request/response boundary. App continues to
own session integration, directory queries/pagination, and ordinary incident_id
workbook startup. Confirmed creation is recorded before handoff; failed handoff
cannot re-enter create dispatch.

| Paths | Purpose |
| --- | --- |
| apps/web/src/app/incidentCreationModel.ts; incidentCreationModel.test.ts | Operation owner, payload/replay/lifetime tests. |
| apps/web/src/app/useIncidentCreation.ts; IncidentCreationForm.tsx | React subscription and native same-surface presentation/focus. |
| apps/web/src/app/App.tsx; IncidentLanding.tsx; landingAdminTypes.ts | Session/route handoff and removal of modal/scattered draft wiring. |
| apps/web/src/app/api/appShellClient.ts; publicHttpTypes.ts | Existing generated createIncident binding, exact request allowlist and validated response. |
| apps/web/src/app/App.landing.test.tsx; apps/web/src/testing/appShellTestSupport.ts | App lifecycle regressions and complete public incident fixtures. |
| packages/ui-contracts/src/applicationSelectors.ts; application-selectors.test.ts | Stable creation form/status selectors. |
| apps/web/e2e/auth-and-incident-directory.spec.ts; pages/incidentDirectory.ts; support/incidents/creation.ts | Real creation/replay/bootstrap and deterministic response/presentation helpers. |
| apps/web/e2e/workbook.a11y.spec.ts; workbook.visual.spec.ts | Keyboard/geometry/accessibility, eight visual states, existing filter fixture synchronization. |
| apps/web/e2e/workbook.visual.spec.ts-snapshots/incident-create-*.png | Eight individually reviewed new goldens listed above. |
| tools/test_families/module.incidents.json; web.design.json | Authored creation and design verification routing. |
| tools/browser_e2e_batch_manifest.json; execution_topology_render_index.json; frontend_visual_golden_manifest.json | Make-generated projections and reviewed golden manifest. |
| docs/handoffs/incident-directory-creation-refactor-handoff.md | This implementation/support record. |

No local user, membership, preference, saved view, initial sheet, server default,
capability discovery, or new public route is synthesized by this workflow. The
existing server transaction remains the sole creator/admin bootstrap owner.
Session identity uses the public actor/authentication/provider values plus local
invalidation generations; membership refresh within the same authentication
lifetime retains the draft. The 30-second limit bounds client observation and
does not assert server cancellation or rollback.

### Acceptance dispositions

These A001–A027 IDs are advisory review routing. Core/design owners above govern
their applicable scope; incident/workspace creation is distinct from completed
view-scoped grid-row creation.

| ID | Disposition and evidence |
| --- | --- |
| A001 | PASS: owner-to-change map above; no advisory source becomes authority. |
| A002 | PASS: one incident-directory creation seam; fixture synchronization and generated evidence are explicitly identified. |
| A003 | PASS: actual branch/HEAD, initially clean state, stack, owners, generated policy, initial files and routing inspected. No new direct grid imports. |
| A004 | PASS: existing landing styles and tokens; no second token, theme, or density registry. |
| A005 | PRESERVED: existing dark_graphite appearance; no new theme or palette. |
| A006 | PRESERVED BASELINE: shared workbook density/editor geometry untouched; full application/a11y/visual regressions pass. |
| A007 | NOT AN INCIDENT-CREATE REQUIREMENT: create_capable, inline_create and fields[].create_writable belong to view-scoped row creation. That completed baseline is unchanged; Core 01 REQ-01-154–159 instead governs this payload. |
| A008 | PASS IN SCOPE: four viewport sizes, short height, zoom, and text-spacing form reachability pass. Shell/inspector breakpoint and clamp owners remain unchanged. |
| A009 | PASS IN SCOPE: native directory scrolling, no fixed-height nested form scroller or horizontal document overflow; safe account navigation remains reachable. Workbook scroll ownership is unchanged. |
| A010 | PRESERVED BASELINE: inspector registry/routes/features untouched; existing visual and accessibility checks pass. |
| A011 | PASS IN SCOPE: explicit opening/return focus, no background error focus jump, and chosen-route preservation pass. Workbook continuity remains its existing owner's responsibility. |
| A012 | PASS: Web Crypto IDs, synchronous duplicate guard, immutable unresolved request, same-ID replay, explicit new attempt after definitive transaction rejection. |
| A013 | NOT AN INCIDENT-CREATE REQUIREMENT: workbook blocked-edit queue/new-ID retry/discard remains unchanged. This workflow uses actor-scoped incident replay and no workbook/global queue. |
| A014 | PASS IN SCOPE: native Enter, multiline description, ordinary Tab, Escape, draft retention, read-only uncertainty, explicit reopen. Grid click/paste/edit behavior remains unchanged. |
| A015 | PASS IN SCOPE: key conflict and recognized public field errors are beside their controls; operation feedback is local. Workbook cell-conflict presentation is unchanged. |
| A016 | PASS: editing/submitting/rejected/uncertain/created and handoff failure are distinct; independent directory refresh/loading is preserved. |
| A017 | PRESERVED: authorized directory rows and query/pagination ownership remain outside the controller; App query regression slice passes. Initial-refresh/debounce fixture limitation is recorded separately. |
| A018 | PRESERVED BASELINE: evidence lifecycle/overlay/preview combinations unchanged; full visual/a11y checks pass. |
| A019 | PASS: real keyboard assertions, names/error associations, visible focus, contrast, restrained live regions, reduced motion, zoom and text-spacing evidence; full a11y gate passes. |
| A020 | PASS IN SCOPE: deterministic form editing/error/pending/recovery/confirmed states use existing component styles. Existing component-matrix evidence remains unchanged. |
| A021 | PRESERVED BASELINE: no workbook virtualization changes or new performance claim. |
| A022 | PASS: 26 existing registered fixtures and all 110 capture intents reconcile; eight new nonregistry captures use exact authored row/scenario/project identities and reviewed settings. Two fresh ordinary passes succeed. |
| A023 | PASS: semantic native controls plus shared stable form/status selectors; package.ui selector tests pass. |
| A024 | PASS: executable code/tests/generation consume no Markdown/digest paths, contents, hashes, or structure. Markdown is human review/support only. |
| A025 | PASS: authored catalog inputs feed Make generation; generated protocol/backend/UI roots untouched. Drift, JSON shape, catalog, and generated-policy checks cover projections. |
| A026 | PASS: existing API and authorization boundaries; no route/schema/storage/lifecycle migration or new incident defaults. Client interaction states are local, not wire/stored contracts. |
| A027 | PASS: owners, architecture, changed paths, behavior, evidence, failures, limitations, compatibility and rollback are recorded here. No next seam is authorized or started. |

### Compatibility, rollback, and limitations

Compatibility: frontend interaction remediation over the existing typed API;
no stored-data migration and no backend, API/schema, authorization, route,
dependency, lockfile, persistent-storage, adopted-spec, or digest change.
Both 201 creation and 200 actor-scoped replay keep the existing atomic bootstrap
and ordinary workbook launch policy. The destination workbook owns its own
loading/error/focus behavior after route handoff.

Rollback must restore the controller/hook/form, App and IncidentLanding wiring,
typed adapter/type exports, selectors, test fixtures, authored routing inputs,
generated projections, eight new goldens and their manifest as one unit. Restore
the baseline interaction and evidence together; do not leave new replay tests
bound to the old dispatch behavior or a manifest pointing at removed images.
There is no data rollback or server migration. This handoff remains the audit
record of that compatibility boundary.

Drafts and recovery are memory-only: close/reopen and same-session application
navigation retain them; session termination, account/authentication replacement,
controller disposal, or reloading/closing the tab clears them. Observation abort
never proves the server cancelled creation. No background replay, unload promise,
or persistent recovery was added. A new attempt cannot silently replace an
unresolved one. Initial directory refresh can suppress a concurrently typed
debounced query under unchanged query code; the visual fixture explicitly
submits its filter, and no directory query redesign is included.

The previously completed menu handoff's broader browser failures were historical
context, not assumed current failures. This seam's selected baseline and final
creation tests pass. Actual failures encountered here are recorded above,
including characterization, fixture/type/import authoring fixes, initial golden
absence, unchanged filter synchronization, and private-artifact permissions.
No test expectation or screenshot threshold was weakened.

Full release/check, backend-wide suites, measurement gates, and owner evidence
publication audit were not selected for this frontend seam. No qualifying full
warm check/support evidence packet exists for the exact source, so RESULTS_DIR
remains unset; retained canonical-run, scheduler ordering/timing, and performance
maintenance are skipped by agent-finalize. These UI artifacts make no release,
Core 05, or new performance claim.

### Final commands and evidence

All repository verification ran through public Make targets from the root.
Focused creation selections used:

```bash
make test-slice OWNER=module.incidents \
  ROWS=module.incidents.frontend.incident_creation_lifecycle_keyboard_replay_and_e99c1bedb1,module.incidents.frontend.incident_creation_operation_payload_outcomes_and_2ecdf8c632,module.incidents.frontend.the_ordinary_landing_shell_creates_an_incident_r_b64ef5986b
make service-backed-test-slice OWNER=module.incidents \
  ROWS=module.incidents.browser.incident_creation_native_keyboard_payload_and_fi_b6b4631f5f,module.incidents.browser.incident_creation_uncertain_replay_and_committed_985931d51a,module.incidents.browser.incident_creation_preserves_chosen_navigation_af_24cb9bd4de,module.incidents.browser.the_ordinary_authenticated_landing_screen_create_fa80fbfef5
make service-backed-test-slice OWNER=web.design \
  ROWS=web.design.accessibility.incident_creation_accessible_keyboard_feedback_a_f2a19fbfdb
make test-slice OWNER=package.ui \
  ROWS=package.ui.frontend_unit.application_selector_contracts_675995f2ab
```

| Command / check | Result and retained run |
| --- | --- |
| Creation frontend slice above | PASS, 20260905T161912Z-p38572. |
| Creation browser slice above | PASS, 20260905T162914Z-p46655. |
| Focused accessibility slice above | PASS, 20260905T154530Z-p35302; superseded by full final-source gate below. |
| Selector slice above | PASS, 20260905T161912Z-p38593. |
| make test-slice OWNER=web.application | PASS, 20260905T161912Z-p38580, 58/58 units. |
| make browser-e2e-a11y | PASS, 20260905T161851Z-p95175, 12/12 units. |
| make browser-e2e-visual-update | PASS, 20260905T160935Z-p83842; reviewed/restored unrelated candidate changes as above. |
| make browser-e2e-visual, twice after final presentation edit | PASS, 20260905T162101Z-p51590 and 20260905T162445Z-p98917, each 12/12 units and 110 reconciled captures. |
| make generate | PASS, 20260905T161402Z-p32475 after golden scope restoration. |
| make agent-finalize before final static checks | PASS, 20260905T163043Z-p91096; generated unchanged; JSON shape/catalog/tier/drift pass; retained-run maintenance skipped with RESULTS_DIR unset. |
| make frontend-typecheck | PASS, 20260905T163130Z-p94574. |
| make frontend-import-boundary-check | PASS, 20260905T163130Z-p94768. |
| make lint-biome | PASS, 20260905T163130Z-p94710. |
| make generated-artifact-policy-check | PASS, 20260905T163130Z-p94244. |
| make lint-markdown | PASS, 20260905T163522Z-p97004; rerun after this final tracker edit. |
| git diff --check and explicit allowed-path/old-golden audit | PASS; rerun after this final tracker edit. |

Retained roots are under `.cartulary/test-results/<run-id>/`. Browser targets
retain per-row results, Playwright reports/traces, accessibility summaries, and
visual reconciliation/renderer attestations. Consult `make explain-run
RESULTS_DIR=<run-root>` for execution details. A queried, undeclared standalone
fixture-check target returned a usage error; registry validation instead uses
the authored target surface, JSON shape, and full visual reconciliation. No
unsupported tool command substituted for required Make evidence.

Final scope audit: exactly 32 changed/new allowed paths, comprising this seam's
source/tests, authored routing, three generated evidence projections, eight new
goldens and this handoff. All 102 pre-existing goldens are byte-identical to HEAD.
Branch remains main at `3972662b1ca2f994eec4ae453cc3f231de4cb6e1`; the dirty tree
contains only this authorized work. No excluded path changed. `git diff --check`
passes. No commit, push, deployment, or next seam occurred.

IC-05 exit: all required seam evidence and acceptance dispositions are complete;
compatibility and atomic rollback are documented, and no owner contradiction or
unresolved creation defect remains. Remaining limitations and intentionally
unselected broader checks are explicit above. The work is ready for user review;
no additional seam or publication action is authorized.
