# Workbook Grid Operational State Plane Refactor Handoff

## Baseline

- Started: `2026-09-04T08:04:30-04:00`.
- Branch: `main`.
- Commit: `2494ab3db5afd2d7fc91fd00b9b2f234a4869f1b`.
- `origin/main`: `2494ab3db5afd2d7fc91fd00b9b2f234a4869f1b`.
- Initial Git state: clean.
- Toolchain: Node `v24.15.0`, pnpm `10.33.0`, Go `go1.27.1
  linux/amd64`, GNU Make `4.4.1`.
- No Git commit is authorized.

## Authority and owner reconciliation

Core 03 REQ-03-286 and REQ-03-299 own fixed work-area geometry and the
closed query-data/interaction-state vocabularies. Core 04 AC-484 restates the
authorization, lifecycle, ARIA, retention, and no-synthetic-record acceptance
boundary. `docs/design.md` sections 7, 8, 10.6, 10.8, 12.3, and 12.8 own the
presentation projection. `docs/domain.md` already supplies the applicable
vocabulary and owner navigation and requires no change.

No contradiction was found among the adopted owners. The UI/UX digest and its
upstream material remain advisory.

The completed view-bar, Inspector, density/renderer, recovery, and Network
Analysis handoffs were reviewed. This slice preserves their closed seams:
fixed view-bar and status-strip geometry, semantic grid-entry focus, Inspector
ownership and scrolling, adapter-owned density and vendor containment,
single-locus recovery, and Network Analysis workflow/virtualization behavior.

## Allowed scope

- Narrow design-owner clarification, authored presentation JSON/schema and
  generator, generated UI-contract projection, and its typed facade.
- Grid Adapter semantic data-state policy, production binding, support
  binding, styles, and focused tests.
- Workbook and Network Analysis state producers/consumers and focused tests.
- Applicable browser accessibility, stateful, support, webserver-backed,
  measurement, and visual evidence; authored harness catalogs/registries and
  Make-generated topology/visual manifests.
- Reviewed visual goldens and this handoff.

Excluded: routes and payloads; query grammar, sorting, grouping, filters, and
saved-view semantics; authorization and incident lifecycle; persistence and
database schemas; row-creation/mutation/conflict algorithms; Inspector or
view-bar information architecture; virtualization and vendor coordinates;
dependencies, themes, and stored data.

## Workstream ledger

| Workstream | Status | Exit condition |
| --- | --- | --- |
| GS-01 — Authority, state inventory, and expected-red characterization | DONE | Baseline, complete producer/consumer/state inventory, advisory dispositions, and user-observable production failure are recorded. |
| GS-02 — Authored presentation policy and shared semantic model | DONE | The design owner, authored/generated projection, facade, and exhaustive resolver agree on one closed policy. |
| GS-03 — Production renderer, support binding, and consumer migration | DONE | Production and support share one state plane; Workbook and Network Analysis consumers preserve owner behavior. |
| GS-04 — Responsive, accessibility, lifecycle, and visual evidence | DONE | Required behavioral, accessibility, geometry, and production-backed visual evidence passes and every retained PNG is reviewed. |
| GS-05 — Final verification, scope audit, and handoff | DONE | All applicable routed/terminal checks pass, scope is clean, and this record is complete. |

Only one row may be `IN_PROGRESS`. A workstream is logged and marked `DONE`
before its successor is opened. An adopted-owner conflict is recorded as
`BLOCKED: owner contradiction` and stops work.

## Current architecture and inventory

The public `GridDataState` union already defines `initial_loading`, `ready`,
`refreshing`, `empty`, `filtered_empty`, `stale_error`, `unavailable`, and
`permission_denied`; `GridInteractionMode` independently defines `editable`
and `read_only`. The public union, props, handle, and package exports remain
stable.

| Producer or consumer | Current responsibility | Required disposition |
| --- | --- | --- |
| `workbookGridDataState` | Maps Workbook query and owner action state to the shared union. | Retain the mapping and owner-owned messages/actions; prove all branches. |
| Timeline presentation | Supplies Timeline copy, draft focus, retry, filter clear, and generation. | Keep draft/create and query behavior unchanged; make empty visibly explanatory. |
| Generic surfaces | Covers Evidence and other registered system views, owner copy, optional draft creation, and query state. | Preserve surface copy/authority and use the shared plane. |
| Entity surfaces | Covers Hosts and Identities with owner copy and draft focus. | Preserve entity behavior and use the shared plane. |
| Assessments | Supplies assessment copy and its existing standalone draft action. | Preserve that creation path; do not manufacture an inline draft. |
| Workbook query controllers | Own initial/refresh/stale/unavailable/access-loss transitions and authorized row clearing/retention. | Retain query/lifecycle semantics; strengthen access-loss evidence. |
| Workbook semantic focus | Owns generation-keyed surface entry through `GridHandle`. | Preserve semantic focus; block hidden state content and restore only after explicit state action replacement. |
| Network Analysis grids | Adapt accepted, rejected, and contributor query state to `GridDataState`. | Compute once, map typed protected loss to `permission_denied`, and remove duplicate raw-kind announcements. |
| Production Grid Adapter | Resolves state and renders an absolute state layer plus a separate read-only layer around RDG. | Replace with one body-owned, density-aware plane outside vendor containment. |
| Grid Adapter test support | Re-resolves state and renders separate support markup before its table. | Consume the same component, timer, decisions, actions, roles, and retention without claiming RDG behavior. |
| Browser and visual fixtures | Real empty Workbook captures coexist with injected empty/loading/error HTML facsimiles. | Keep real service-backed evidence and remove synthetic state-presentation substitutes. |

## Closed state inventory

| State | Stable identity | Authorized rows | Blocking | Visible/action posture | Focus/live/transition posture |
| --- | --- | --- | --- | --- | --- |
| `ready` | State kind | Current owner materialization | No | No data message; read-only may co-display | Preserve; no data announcement |
| `initial_loading` | Surface identity plus loader generation | None presented | Yes | Immediate owner-label loading copy; exact delayed copy; no action | No automatic move; polite; reset/cancel by generation/state |
| `refreshing` | State kind on current surface | Retain prior authorized rows and draft | No | Compact progress; no action | Preserve selection, scroll, focus; polite |
| `empty` | State kind on current surface | No committed rows; retain owner draft | No | Surface-owned neutral copy; optional owner create | No automatic move; polite; create targets existing owner path |
| `filtered_empty` | State kind plus retained query | No matches; retain query/draft | No | Neutral filter copy; `Clear filters` | No automatic move; polite; action replacement returns to grid root |
| `stale_error` | State kind on current surface | Retain prior authorized rows and draft | No | Caution; optional owner Retry | Preserve until activation; assertive |
| `unavailable` | State kind on current surface | None presented | Yes | Safe owner error; optional owner Retry | Never autofocus Retry; assertive |
| `permission_denied` | Typed access-loss state | Suppress immediately and clear in owner | Yes | Restrained safe error; no local action | Assertive; authenticated-root recovery owns final focus |

Read-only is independent from the table above. It co-displays through the same
plane, never weakens a data state, preserves read/copy, and blocks all mutation
entry. `Closed, read-only` remains exact.

## Current routing

- `package.grid_adapter`: focused and service-backed slices; broader
  accessibility, visual, Fallow, and `test-fast`.
- `web.workbook`: focused slice; broader `test-fast`.
- `web.design`: focused and service-backed slices; broader accessibility,
  visual, and webserver-backed.
- `module.workbook`: focused and service-backed slices; broader
  accessibility, stateful, support, visual, webserver-backed, and `test-fast`.
- `web.networkflow`: focused slice; broader `test-fast`.
- `module.networkflow`: focused and service-backed slices; broader
  accessibility, measurement, stateful, visual, webserver-backed, targeted
  Go security, and `test-fast`.
- `package.ui` and `web.architecture`: focused projection/boundary slices.

## Advisory disposition

- ADOPT R001-R015.
- ADAPT R016-R025 and R035 only through Cartulary's current desktop bands,
  grid-owned horizontal scrolling, token/density rules, semantic state/actions,
  and owner-defined workflows.
- REJECT R026-R034, including cyberpunk/matrix styling, alert theater,
  generated palettes, marketing or dashboard composition, decorative motion,
  advisory behavior authority, and incidental selectors.
- The required offline searches and result-by-result dispositions are recorded
  in the GS-01 checkpoint below.

## Checkpoint log

| Timestamp | Workstream | Changes and findings | Commands and results | Compatibility, rollback, next action |
| --- | --- | --- | --- | --- |
| `2026-09-04T08:04:30-04:00` | GS-01 start | Created this handoff before product edits; recorded the clean baseline, authority, scope, predecessor seams, state/consumer inventory, and refreshed owner routing. | Git/tool checks passed; task guides and owner manifests were inspected during planning. | Documentation only. No product behavior changed. Rollback this file with the complete slice. Next: run advisory searches and capture the real empty-state expected red. |
| `2026-09-04T08:09:10-04:00` | GS-01 completion / GS-02 start | The UX search returned explanatory empty state, mobile pull-to-refresh, and generic bulk actions: ADOPT the local explanatory state; REJECT mobile refresh and bulk-workflow changes as out of scope. The React search returned focus return, live updates, Actions, async errors, and memoization: ADAPT semantic focus return and live updates; ADOPT existing typed async-error handling; REJECT a transition rewrite and unmeasured memoization. The production Timeline empty state is present in the accessibility tree, but the existing golden does not paint it and activating its `Add row` action leaves focus on the action instead of the existing draft. | The first characterization run passed before action behavior was asserted: `.cartulary/test-results/20260904T120552Z-p69747`. The strengthened user-observable expected-red failed as intended at `.cartulary/test-results/20260904T120738Z-p15475`: the visible owner action did not transfer focus to the draft. The retained trace and accessibility snapshot show the semantic state and draft coexist while the committed golden remains blank. | No owner contradiction. The failure is classified as a product presentation/focus defect, not accepted behavior. GS-01 is `DONE`; only GS-02 is open. Next: project and generate the closed semantic policy before changing either renderer. |
| `2026-09-04T08:18:40-04:00` | GS-02 completion / GS-03 start | Added the closed eight-row data-state policy, the two-row interaction policy, and their precedence/live composition to the design owner and authored JSON/schema. The generator now validates exact cardinality and emits typed camel-case rows; the UI facade performs typed lookup. The Grid Adapter resolver is total, resolves copy/actions from typed strategy fields, and exposes fail-closed row/draft materialization helpers. | `make generate` passed at `.cartulary/test-results/20260904T121315Z-p61417`; generated diffs were inspected. `make test-slice OWNER=package.ui` passed at `.cartulary/test-results/20260904T121552Z-p65365`. The combined `web.design` slice reached 13/15 before an existing responsive keyboard measurement reported 23 px shell overflow at `.cartulary/test-results/20260904T121556Z-p66731`; no rendered source had changed, so this is retained for GS-04 geometry reconciliation. The Grid Adapter slice reached 43/44 at `.cartulary/test-results/20260904T121716Z-p17098`; its only failure is the expected transitional delayed-loader call site, which GS-03 now replaces with the shared phase API. The new exhaustive semantic-kernel row passed. | Public state, action, interaction, props, handle, and export surfaces remain unchanged. Generated output came only from the authored projection/generator. GS-02 is `DONE`; only GS-03 is open. Next: replace both renderers' separate layers with one package-private production state plane. |
| `2026-09-04T08:44:05-04:00` | GS-03 completion / GS-04 start | Production and support now share one package-private operational state plane, resolver, delayed timer, action guard, row/draft retention policy, roles, live priority, and interaction-state composition. Workbook creation actions focus existing drafts. Network Analysis computes each state once, binds typed errors, maps protected-state loss to `permission_denied`, and no longer emits a duplicate raw-kind announcement. The strengthened production Timeline scenario now paints its owner message and action; its remaining expected mismatch is the intentionally stale golden. A Workbook fixture had used a non-UUID Indicator record ID: the new required suppression for an unavailable response exposed that invalid fixture, which was corrected to a valid stable ID rather than weakening the blocking state. | `make format` initially failed at `.cartulary/test-results/20260904T122408Z-p68542` because the new semantic section redundantly declared its implicit role; removing that declaration corrected it. `make frontend-typecheck` initially failed at `.cartulary/test-results/20260904T122441Z-p77176` because the projected `none` live priority required an explicit DOM `off` mapping; corrected and rerun. Final format passed at `.cartulary/test-results/20260904T123418Z-p36897`, typecheck at `.cartulary/test-results/20260904T123422Z-p41063`, focused Grid Adapter at `.cartulary/test-results/20260904T123433Z-p41515`, and `web.networkflow` at `.cartulary/test-results/20260904T123438Z-p42041`. The service-backed expected-red now differs only from the old blank PNG at `.cartulary/test-results/20260904T122915Z-p85306`; its actual was manually inspected and visibly contains headers, owner copy, action, and the existing draft. The full Workbook slice exposed the invalid Indicator fixture at `.cartulary/test-results/20260904T123730Z-p47655`; after correction its exact row passed at `.cartulary/test-results/20260904T124325Z-p67776`. | Query, authorization, lifecycle, persistence, virtualization, and public Grid Adapter interfaces remain unchanged. GS-03 is `DONE`; only GS-04 is open. Next: replace injected state facsimiles, broaden lifecycle/accessibility/geometry coverage, and update/review only production-backed goldens. |
| `2026-09-04T09:51:33-04:00` | GS-04 completion / GS-05 start | Removed injected empty/loading/error presentation facsimiles and replaced them with production Workbook/Grid Adapter scenarios for immediate and delayed loading, successful and filtered empty, retained-row background refresh and stale failure, unavailable initial load, editable/read-only/closed interaction, responsive/density/zoom/text-spacing variants, and Network Analysis filtered empty. Final reconciliation covers 94 active captures with zero missing, ambiguous, or unresolved fixtures. All 22 changed or added PNGs were manually reviewed; the one confirmed orphan was the deleted synthetic `design-error-presentation-loci` mosaic, and one unrelated nondeterministic entity rewrite was restored. The real Timeline action focus now selects the first presented draft column instead of a horizontally virtualized hard-coded field. | Failed update attempts at `.cartulary/test-results/20260904T125329Z-p24335`, `.cartulary/test-results/20260904T130006Z-p21209`, and `.cartulary/test-results/20260904T131242Z-p21476` exposed and corrected the draft-focus target, closed-state action scope, test helper identity, offscreen trigger, and projected stale-copy expectation. Final refresh reconciliation passed at `.cartulary/test-results/20260904T144608Z-p60052`; its manifest was regenerated at `.cartulary/test-results/20260904T144908Z-p8021`. Final-set ordinary visual runs passed at `.cartulary/test-results/20260904T144921Z-p11187` and `.cartulary/test-results/20260904T145126Z-p58401`. Accessibility passed at `.cartulary/test-results/20260904T132544Z-p11125`, measurement at `.cartulary/test-results/20260904T132714Z-p57580`, stateful at `.cartulary/test-results/20260904T133147Z-p16895`, and support at `.cartulary/test-results/20260904T133410Z-p68090`. The first webserver-backed run failed one of 60 units at `.cartulary/test-results/20260904T133535Z-p14195`: horizontal virtualization deferred the semantic cell registry beyond the two-task focus restore. Bounded paint-frame retries using the existing semantic registry corrected it; the exact row passed at `.cartulary/test-results/20260904T134441Z-p20216` and the complete 60-unit rerun passed at `.cartulary/test-results/20260904T134538Z-p65253`. | No synthetic records, presentation-only HTML fixture, selector-driven focus path, or vendor-coordinate behavior was introduced. Geometry, read-only behavior, authorized-row retention, and access-loss suppression are green. GS-04 is `DONE`; only GS-05 is open. Next: refresh all owner routes, run terminal verification, perform closure/scope audits, and finalize this record. |
| `2026-09-04T10:42:32-04:00` | GS-05 completion | Refreshed all eight task guides; ran every focused and service-backed route; completed generation, frontend, browser, security, fast, Markdown, and Git checks. Closure searches found no duplicate presentation switch/copy, synthetic state markup, message-selected runtime behavior, production selector focus, support RDG import, fake record, compatibility shim, Markdown runtime dependency, temporary marker, or vendor import outside the adapter. The final Git audit contains only the authorized design/projection, Grid Adapter, Workbook/Network Analysis, test/harness, reviewed golden, and handoff paths; the index is unstaged and no commit was created. | Final `make agent-finalize` passed with `RESULTS_DIR` unset at `.cartulary/test-results/20260904T145418Z-p5413`; final substantive Markdown lint passed at `.cartulary/test-results/20260904T145806Z-p20142`; `git diff --check` and `git diff --cached --check` passed. Detailed terminal roots and all failure/correction pairs are recorded below. | Public Grid Adapter types/exports, API/routes, query/saved-view semantics, authorization/lifecycle/persistence, virtualization, Network Analysis workflow, dependencies, and stored data are compatible. Rollback is source-only with no data migration. GS-01 through GS-05 are `DONE`; no workstream remains open. |

## Implemented paths and behavior

- `contracts/design/presentation.v1.json`, its schema and generator, and the
  generated `packages/ui-contracts` projection now own exactly eight data-state
  rows, two independent interaction rows, and one precedence/live composition
  rule. The generated facade is exhaustive and runtime code reads no Markdown.
- `packages/grid-adapter/src/GridOperationalStatePlane.tsx`,
  `SemanticDataGrid.tsx`, `semanticDataState.ts`, `test-support.tsx`, and
  `styles.css` provide one density-aware production/support state plane. The
  plane owns typed resolution, loading timing, roles/live priority, action
  guarding, action-origin focus recovery, and fail-closed row/draft
  materialization without changing the public adapter API.
- Workbook owners still supply surface copy, create/filter/retry authority,
  query state, and draft callbacks. Timeline creation now resolves its first
  presented draft field semantically, and semantic focus restoration tolerates
  delayed horizontal virtualization without reclaiming focus from another live
  control.
- Network Analysis passes typed request errors to each grid, maps protected
  loss to `permission_denied`, retains authorized accepted/rejected/contributor
  rows for ordinary refresh failure, clears them for protected or graph-stale
  loss, and removes its duplicate raw-kind announcement.
- `apps/web/e2e/workbook.visual.spec.ts` now drives the production Workbook or
  Network Analysis component for every operational-state capture. The injected
  empty/loading/error presentation HTML and synthetic error mosaic were
  removed. The reviewed goldens and fixture registry cover immediate/delayed
  load, successful/filtered empty, stale/unavailable, read-only/closed,
  responsive/density/zoom/text-spacing, and Network Analysis regression.

Before this slice, the production Timeline successful-empty state existed in
the semantic tree but painted as an unexplained blank grid, its create action
did not reach a currently presented draft control, read-only was a competing
absolute layer, support duplicated presentation, and synthetic HTML supplied
part of the visual evidence. After this slice, owner copy and permitted actions
are visible inside the grid body, all operational states resolve through the
closed projection and one shared plane, and the production visual path proves
the result.

## Final verification

- Final authored generation: `make generate` passed at
  `.cartulary/test-results/20260904T144908Z-p8021`; generated diffs were
  inspected. `make generate-drift`,
  `make generated-artifact-policy-check`, and `make json-shape-check` passed at
  `.cartulary/test-results/20260904T143919Z-p44136`,
  `.cartulary/test-results/20260904T143928Z-p47219`, and
  `.cartulary/test-results/20260904T143929Z-p47629`.
- Final focused owner evidence passed for `package.grid_adapter` (44/44,
  `.cartulary/test-results/20260904T135235Z-p26599`), `web.workbook` (160/160,
  `.cartulary/test-results/20260904T135335Z-p76120`), `web.design` (15/15,
  `.cartulary/test-results/20260904T135420Z-p93344`), `module.workbook` (69/69,
  `.cartulary/test-results/20260904T135524Z-p42190`), `module.networkflow`
  (34/34, `.cartulary/test-results/20260904T142804Z-p60753`), `package.ui`
  (10/10, `.cartulary/test-results/20260904T143858Z-p38266`),
  `web.networkflow` including its added operational-state row (39/39,
  `.cartulary/test-results/20260904T143903Z-p39659`), and `web.architecture`
  (12/12, `.cartulary/test-results/20260904T140007Z-p69273`).
- Service-backed owner evidence passed for `package.grid_adapter` (13/13,
  `.cartulary/test-results/20260904T140018Z-p70802`), `web.design` (15/15,
  `.cartulary/test-results/20260904T140114Z-p17253`), `module.workbook` (39/39,
  `.cartulary/test-results/20260904T140218Z-p65458`), and final
  `module.networkflow` (28/28,
  `.cartulary/test-results/20260904T143013Z-p23378`).
- Final frontend and boundary evidence passed: typecheck at
  `.cartulary/test-results/20260904T145525Z-p17188`, 432/432 frontend units at
  `.cartulary/test-results/20260904T143305Z-p84958`, import boundaries at
  `.cartulary/test-results/20260904T140833Z-p15756`, Biome at
  `.cartulary/test-results/20260904T145535Z-p17658`, and Fallow at
  `.cartulary/test-results/20260904T140921Z-p21489`.
- Browser evidence passed: accessibility 12/12 at
  `.cartulary/test-results/20260904T140946Z-p22396`, measurement 22/22 at
  `.cartulary/test-results/20260904T141117Z-p70151`, stateful 34/34 at
  `.cartulary/test-results/20260904T141544Z-p29539`, support 19/19 at
  `.cartulary/test-results/20260904T141804Z-p80941`, and webserver-backed 60/60
  at `.cartulary/test-results/20260904T134538Z-p65253`. The exact focus row and
  final Network Flow service-backed route passed after their respective fixes.
- The final visual update passed at
  `.cartulary/test-results/20260904T144608Z-p60052` with 94 active captures and
  zero orphan, missing, ambiguous, or unresolved artifacts. All 22 changed or
  added PNGs were manually reviewed; one unrelated command-produced entity
  rewrite was restored before manifest regeneration. Ordinary final-set visual
  comparisons passed twice at
  `.cartulary/test-results/20260904T144921Z-p11187` and
  `.cartulary/test-results/20260904T145126Z-p58401`.
- `make agent-finalize` passed with `RESULTS_DIR` unset at
  `.cartulary/test-results/20260904T145418Z-p5413`. Targeted Go security passed
  at `.cartulary/test-results/20260904T142142Z-p73546`; final `test-fast` passed
  484/484 at `.cartulary/test-results/20260904T143457Z-p26873`. Final Markdown
  and Git whitespace checks are recorded in the closing checkpoint.

## Failure and correction log

| Target and run root | Classification | Correction and successful evidence |
| --- | --- | --- |
| Expected-red service-backed Timeline at `.cartulary/test-results/20260904T120738Z-p15475` | Expected product defect | Shared production plane plus semantic draft focus; production scenario and refreshed visual passed. |
| `web.design` slice at `.cartulary/test-results/20260904T121556Z-p66731` | Pre-renderer geometry failure | Reconciled through adapter-owned work-area geometry; final 22/22 measurement suite passed. |
| Grid Adapter slice at `.cartulary/test-results/20260904T121716Z-p17098` | Expected transitional call site | Replaced arbitrary delayed copy with the shared loading phase; final adapter slice passed 44/44. |
| `make format` at `.cartulary/test-results/20260904T122408Z-p68542` | Authored-code lint | Removed a redundant implicit role; later format and Biome passed. |
| `make frontend-typecheck` at `.cartulary/test-results/20260904T122441Z-p77176` | Authored DOM mapping | Mapped projected live `none` to DOM `off`; final typecheck passed. |
| Workbook slice at `.cartulary/test-results/20260904T123730Z-p47655` | Invalid pre-existing test fixture exposed by fail-closed suppression | Replaced the non-UUID Indicator identity with a valid stable fixture ID; exact row passed at `.cartulary/test-results/20260904T124325Z-p67776`. |
| Visual updates at `.cartulary/test-results/20260904T125329Z-p24335`, `.cartulary/test-results/20260904T130006Z-p21209`, and `.cartulary/test-results/20260904T131242Z-p21476` | Product/test evidence | Corrected draft focus, closed action scope, helper identity, offscreen trigger, and stale-copy expectation; final update passed. |
| `browser-e2e-webserver-backed` at `.cartulary/test-results/20260904T133535Z-p14195` | Product focus regression, 59/60 units | Added bounded semantic-registry paint-frame restoration; exact row passed at `.cartulary/test-results/20260904T134441Z-p20216`, full rerun 60/60. |
| `make lint-biome` at `.cartulary/test-results/20260904T140836Z-p16158` | Authored-code lint | Applied optional chaining to two connectivity guards; Biome passed at `.cartulary/test-results/20260904T140920Z-p21057`. |
| `make frontend-typecheck` at `.cartulary/test-results/20260904T143225Z-p83789` | Test-fixture enum typo | Used typed `retry_with_backoff`; final typecheck and frontend units passed. |
| `make generate` at `.cartulary/test-results/20260904T143700Z-p28360` and `.cartulary/test-results/20260904T143748Z-p31799` | Authored harness ordering | ASCII-sorted projection titles and the new Network Flow row; final generation and drift passed. |
| `make frontend-typecheck` at `.cartulary/test-results/20260904T145444Z-p12272` | Visual-fixture callback inference | Explicitly typed the held-refresh release callbacks as void functions; final typecheck passed at `.cartulary/test-results/20260904T145525Z-p17188`. |

No check was weakened and no unrelated golden was accepted.

## Acceptance ledger

| Row | Result | Rationale |
| --- | --- | --- |
| A001 | PASS | Every behavior maps to Core 03/Core 04/design or a current application owner; advisory results remain classified only. |
| A002 | PASS | One grid operational-state seam changed; structural consolidation and behavior corrections are recorded separately. |
| A003 | PASS | Baseline, owner map, generated policy, direct imports, producer/consumer inventory, fixtures, and dirty state were inspected. |
| A004 | PASS | State geometry uses existing design/density tokens; no local theme or token registry was added. |
| A005 | PASS | Existing `dark_graphite` remains the only theme and bundled cybersecurity visual defaults remain rejected. |
| A006 | PASS | Adapter density metrics drive state/header/draft insets; compact/default/comfortable captures and measurement pass. |
| A007 | PASS | Create remains owner-supplied and gated by existing capabilities; no row contract or payload changed. |
| A008 | PASS | Base, narrow, compact, below-minimum, vertical resize, zoom, text spacing, and inspector geometry pass. |
| A009 | PASS | State planes remain in grid ownership; shell overflow and anchored status-strip measurement pass. |
| A010 | PASS | Inspector ownership/dispatch remains unchanged and its routed regression evidence passes. |
| A011 | PASS | Selection, scroll, rows, and focus survive retained refresh/error and inspector transitions; the virtualization focus regression is fixed. |
| A012 | N/A | Transaction identifiers and uncertain replay were not changed; existing regression evidence remained green. |
| A013 | N/A | Blocked-edit queue recovery behavior was not changed; existing stateful/browser evidence remained green. |
| A014 | PASS | Existing edit, keyboard, paste, draft, commit/cancel, and Escape behavior passes adapter/workbook/browser evidence. |
| A015 | N/A | Same-field conflict behavior was not changed; existing conflict evidence remained green. |
| A016 | PASS | All eight closed data states plus independent interaction state are exhaustive in projection, resolver, production, and support. |
| A017 | PASS | Workbook and all Network Analysis semantic grids retain prior authorized rows for ordinary refresh/failure and clear on protected loss. |
| A018 | N/A | Evidence lifecycle semantics were not changed; existing Evidence regression evidence remained green. |
| A019 | PASS | Keyboard focus, names, live priority, non-color cues, and reduced-motion accessibility pass. |
| A020 | PASS | Every data/interaction combination follows one generated precedence and is exhaustively unit tested. |
| A021 | PASS | RDG virtualization remains adapter-private; many-row, focus, identity, scroll, and service-backed continuity pass. |
| A022 | PASS | D-VFIX-012 through D-VFIX-014 use production captures with exact registry/manifest reconciliation and two ordinary visual passes. |
| A023 | PASS | Runtime focus uses semantic handles/registries; fixture tests use stable owner selectors, never copy for behavior. |
| A024 | PASS | Runtime/tests do not read Markdown; the generator rejects design inputs under `docs/`. |
| A025 | PASS | Generated outputs came only from authored JSON/generator/catalog inputs; generation, policy, and drift pass. |
| A026 | PASS | No route, payload, schema, auth, lifecycle, persistence, virtualization, workflow, dependency, or compatibility behavior was invented. |
| A027 | PASS | This handoff records owners, behavior, paths, commands, failures, visual review, rollback, risks, deferrals, and next seam. |

## Rollback, risks, deferrals, and next safe seam

Rollback is source-only and atomic: revert the authored design projection and
generator, generated UI/topology artifacts, Grid Adapter/Workbook/Network
Analysis presentation changes, test catalogs/fixtures, reviewed goldens, and
this handoff together. No database, stored-data, service, or external rollback
is needed.

No known product risk or temporary adapter remains. The bounded semantic focus
retry is intentionally limited to eight paint frames and stops as soon as the
user focuses another connected control. Transaction recovery, same-field
conflict, and Evidence lifecycle changes remain deferred because they were
outside this seam. The next safe seam is additive adoption by any future
semantic-grid surface through the existing `GridDataState` union and owner
actions, without another renderer or policy registry.

## Compatibility statement

“Workbook and shared semantic-grid operational-state presentation was
corrected and structurally consolidated. Query, saved-view, route,
authorization, lifecycle, persistence, virtualization, Network Analysis
workflow, and stored-data behavior remain unchanged. No data migration is
required.”
