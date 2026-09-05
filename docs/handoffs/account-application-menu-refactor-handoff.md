# Account/Application Menu Refactor Handoff

## Baseline and scope

- Branch `main`; HEAD `f1d06b667c8ed693665128dd62a59144c837722f`; initially clean.
- Inspection recommendation and current source match. No reset, commit, or push.
- Scope: account/application menu, its nested Controls navigation, necessary
  destination focus integration, shared registered-focus mechanics, and evidence.
- Existing save-status, presence, relationships, evidence, grid operational state,
  inspector, density, and responsive behavior are regression baselines.
- Stack: React 19, TypeScript 6, Vite 8, pnpm 10.33, Vitest, Playwright, Biome.
- Only root `AGENTS.md` applies. Digest read in documented order; advisory only.

## Tracker

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| AM-01 — Baseline, authority, characterization | DONE | Five component failures and seven browser assertions reproduce the bounded defects. |
| AM-02 — Menu state and shared focus | DONE | Six menu tests and seven shared-focus tests pass; typecheck and imports pass. |
| AM-03 — Integration and containment | DONE | Browser navigation, destination focus, long labels, viewport/zoom/text-spacing geometry pass. |
| AM-04 — Regression, accessibility, visual | DONE | Focused behavioral/a11y checks pass; eight reviewed menu goldens; two ordinary visual passes. |
| AM-05 — Final validation and handoff | DONE | All in-scope exits satisfied; acceptance, terminal validation, and 44-path audit complete. |

Only the current row may be IN_PROGRESS. Complete its evidence and terminal
status before advancing. Owner contradictions block dependent work.

## Authority and architecture

| Behavior | Governing owner | Implementation boundary |
| --- | --- | --- |
| Application contexts and destinations | Core 01 §§3.3.2.1A–B; design §4.5 | App routing and account menu |
| Keyboard ownership | Core 03 §13.2; design §8.4 | Menu handlers; existing workbook handlers remain owners |
| Escape and restoration | Design §§8.5, 12.6 | Registered live refs and explicit dismissal reasons |
| Controls destinations | Core 01 incident/account/workbook routes; Core 03 §2.4; Core 04 §2 | Existing semantic section callback and drawer |
| Availability and authorization | Core 04 §2, REQ-04-023/028/029/114 | Current session/context; no new authorization policy |
| Responsive navigation | Design §§7.4–7.5 | Menu containment; workbook chrome policy unchanged |
| Names, focus, announcements | Design §§14.1–14.2 | Accurate ARIA and token focus; no menu live announcements |
| Stable identity | Domain §8.1 | Semantic action and section keys, independent of labels |

No inspected owner contradiction. User selected an indented Controls panel and
current-item opening focus (first available fallback). Shared React mechanics
belong under `apps/web/src/shared`, not a contracts package. `web.shared` is a
source owner only; existing shared tests route through `web.application`.

## Interaction matrix

| Input / transition | Before (reproduced against baseline) | Required after |
| --- | --- | --- |
| Open root | Click toggles; focus remains at trigger | Current action; Down first, Up last |
| Up/Down, Home/End | No account menu handlers | Wrap/move inside active menu |
| Open Controls | Click toggles inline content | Current section; Right/Enter/Space/click entry |
| Nested Left/Escape | No nested handlers | Close nested only; focus Controls |
| Root Escape | No handler | Close root; focus trigger |
| Tab/outside focus | Blur closes | Native departure; no restoration |
| Outside nonfocusable pointer | No outside-pointer handler | Dismiss without stealing focus |
| Context/items change | Independent booleans remain | Close/reconcile semantic state |
| Action handoff | Existing callbacks and Controls return target | Exactly once; destination owns focus |
| Expanded short viewport | Unbounded absolute panel | Anchored, bounded internal scrolling |

## Verification routes and baseline evidence

- `make help`, `make help-all`, and task guides for `web.application`,
  `web.workbook`, `web.design`, `module.workbook`, and `module.auth` succeeded.
- `make task-guide ROLE=module-author OWNER=web.shared` failed: no active test
  owner. Use `web.application` for shared tests, following current routing.
- Selected application baseline rows passed from valid cache:
  `.cartulary/test-results/20260905T121910Z-p64757`.
- Existing registered-overlay row passed from valid cache:
  `.cartulary/test-results/20260905T121911Z-p64999`.
- System views real-browser keyboard regression passed:
  `.cartulary/test-results/20260905T122025Z-p66555`.
- These passing baselines do not prove account-menu keyboard behavior.
- At AM-01 entry, `git diff --check` passed before production edits.

## Decisions, evidence, and work log

AM-01 completed before production edits. New authored test rows and source
ownership were registered, then `make generate` passed at
`.cartulary/test-results/20260905T123230Z-p14286`. Generated browser batching and
render-index changes are necessary downstream routing for the new scenario.

`make test-slice OWNER=web.application
ROWS=web.application.regression.account_application_menu_keyboard_focus_and_life_e79d61a853`
failed as expected at `.cartulary/test-results/20260905T123254Z-p17337`:
five of six tests failed (current opening focus, Right-to-Controls, outside
nonfocusable pointer, stale context, removed focused capability). Native Tab
departure and exactly-once ordinary Enter dispatch already pass and are retained.
Assertion details are in that root's `unit-logs/row-*/stderr.log`.

`make service-backed-test-slice OWNER=web.design
ROWS=web.design.browser.account_application_menu_keyboard_navigation_and_8cba739220`
failed as expected at `.cartulary/test-results/20260905T123256Z-p17570`.
The real browser independently reproduced root opening-focus and End movement,
nested opening-focus, nested Escape dismissal/return, and root Escape failures.
At 640×480 the expanded menu top reached −166.90625 CSS px after native focus
scrolling. The Playwright report and trace retain the open-menu state. These are
product failures against unchanged production source, not harness failures.

No existing completion of this seam was found. Next: begin AM-02 using the
approved semantic interaction model and domain-neutral registered-focus seam.

AM-02 completed: extracted the menu from the multi-purpose layout, moved the
registered hook and tests into shared, updated all seven consumers, and removed
the old implementation paths. Semantic root/Controls state, native activation,
item registration, nested return, context invalidation, and item reconciliation
pass at `.cartulary/test-results/20260905T124155Z-p78396`. Informational text is
outside the actionable collection. Controls keeps `role=menuitem` and adds
`aria-current` for its selected section, preserving existing semantic helpers.

Validation: `make format` passed at `20260905T124225Z-p79709`;
`make frontend-typecheck` passed at `20260905T124235Z-p83954`;
`make frontend-import-boundary-check` passed at `20260905T124246Z-p84474`.
These run IDs are under `.cartulary/test-results/`.

Resolved intermediate failures: initial format preflight referenced the moved
test; the authored row migration also referenced the old row. Generation failed
at `20260905T123957Z-p64861` and `20260905T124030Z-p68109` until that migration
was updated, then passed at `20260905T124051Z-p71058`. Format at
`20260905T124105Z-p74027` identified a nonsemantic separator (replaced with `hr`).
Typecheck at `20260905T124156Z-p78669` rejected a Playwright-only `exact` option
in Testing Library (removed). No checks were weakened.

Next: AM-03 destination focus, lifetime identity, and bounded menu placement.

AM-03 completed. App now provides an account/incident/context identity and a
live trigger ref. Directory/admin headings own navigation focus; Account
settings owns opening and explicit-dismissal focus. Context changes discard
stale Account settings state without restoration. The Controls callback and
drawer opening focus remain unchanged. No route, permission, or form behavior
was added.

The panel uses fixed placement while retaining DOM order; registered anchor
geometry, ResizeObserver, viewport events, and root-style changes maintain
bounds. Rect/layout ratios normalize CSS zoom without importing workbook
layout policy or changing responsive algorithms. Focus scrolling updates only
the panel scroll offset. Long header metadata truncates without displacing the
menu. The account label remains fully available in the open panel.

The new authored `accountTestId("application-menu")` selector identifies the
scroll viewport, distinct from its semantic menu content. No generated token
or API projection changed. Existing trigger styling is retained to preserve
closed-trigger visual baselines.

Passing evidence under `.cartulary/test-results/`:

- Typecheck: `20260905T124831Z-p90524`.
- Menu/shared/App integration slice: `20260905T124832Z-p90911`.
- Initial repaired browser and reviewed 640×480 PNG: `20260905T124833Z-p91146`.
- Expanded menu/shared/App tests: `20260905T125543Z-p46091`.
- Expanded browser: `20260905T125545Z-p46330`; root/nested keyboard, Tab,
  inspector Escape isolation, Controls drawer return, directory/admin heading
  focus, Account settings opening/Escape return, outside pointer, 252-character
  account label, six viewports including short/below-minimum, 200% CSS zoom,
  and text spacing all pass with visible focused-item geometry.
- `make generate`: `20260905T125445Z-p38777`; `make format`:
  `20260905T125512Z-p41819`.

Next: AM-04 consumer regressions and dedicated accessibility/visual evidence.

AM-04 progress: added dedicated accessibility and eight open-menu capture
intents, each attached to an authored semantic web.design row. New visual
captures are active nonregistry fixtures, not invented D-VFIX identities.
Existing registered fixtures remain unchanged.

Consumer regression exposed a real extraction policy change: the Saved view
panel temporarily disables its fields while saving. Automatic reconciliation
altered that existing consumer. Its untouched baseline passed in an isolated
worktree at `/tmp/cartulary-account-menu-baseline`, run
`.cartulary/test-results/20260905T130849Z-p33914`. The first worktree command
failed before tests because bootstrap could not resolve Ajv; using
`NODE_PATH=/home/jochi/code/cartulary/node_modules` for the Make invocation
allowed its own frontend install and test execution. No baseline source changed.
Reconciliation is now opt-in for the account menu; existing consumers retain
prior behavior. Grid controls, Surfaces/saved views, and Timeline actions pass
at `20260905T131024Z-p43676`, replacing the failure at
`20260905T130542Z-p3026`. Menu/shared tests pass at `20260905T131127Z-p51382`.

Other passing evidence: selector contracts `20260905T130543Z-p3329`;
System views browser `20260905T130558Z-p7285`; ordinary nonadmin application
navigation `20260905T130600Z-p8085`; new accessibility plus existing responsive
geometry `20260905T130557Z-p7047`. Full accessibility passes at
`20260905T131156Z-p60994`. The preceding full run at
`20260905T130812Z-p90977` found a trigger helper that also matched the newly
named semantic menu; the helper now requests the button role explicitly.

Typecheck caught Node.remove in the two new fixture cleanup calls at
`20260905T130613Z-p94379`; cleanup now uses the Node parent API. Typecheck and
imports subsequently pass at `20260905T131158Z-p73947` and
`20260905T131214Z-p70686`.

Ordinary visual attempt `20260905T130612Z-p85911` reached eight missing new
goldens, but failed reconciliation because authored scenario IDs lacked the
required hexadecimal shape. IDs were repaired in authored rows and regenerated.
That attempt also had a relationship fixture failure; the next ordinary run
passed that unchanged fixture. Attempt `20260905T131154Z-p59133` reconciled all
94 existing goldens with no orphans or ambiguous mappings, but exposed a profile
loading race in the new long-label setup before its last two captures. The
functional row reproduced the same fixture issue at `20260905T131157Z-p65983`.
The fixtures now await profile readiness before editing. The corrected
functional row passes at `20260905T131529Z-p4576`. No profile implementation
changed; visual promotion remains pending a complete ordinary reconciliation.

`make agent-finalize` passed before broad checks at `20260905T130540Z-p2825`,
`20260905T131130Z-p55824`, and `20260905T131528Z-p4379`.
RESULTS_DIR is unset: retained performance evidence, run health, and scheduler
maintenance are explicitly skipped (`results-dir-not-provided`), not passed.

AM-04 focus-lifetime audit also made explicit unmount restoration configurable:
account navigation cancels it, while the seven existing consumers retain the
original default. The new test covers a requested close that unmounts immediately;
ordinary unmount and destination handoff tests remain. Route focus requests now
cancel when the effective destination no longer matches. Menu/shared tests pass
at `20260905T132038Z-p95816`; consumer slices pass at `20260905T132129Z-p16980`.
Typecheck/imports/Biome pass at `20260905T132127Z-p12172`,
`20260905T132142Z-p45263`, and `20260905T132145Z-p50782`.
Generation at `20260905T131851Z-p64772` and the following catalog check rejected
an unsorted newly appended test title. Sorting the authored titles restored
passing generation at `20260905T132037Z-p95217`, format at
`20260905T132046Z-p98955`, and finalization at `20260905T132050Z-p4623`.

Ordinary visual reconciliation at `20260905T131530Z-p4948` is complete:
102 capture intents, all 94 existing goldens active, eight new menu goldens
missing, zero orphans/ambiguous mappings/unresolved registered fixtures.
Every functional assertion completed. Only the new capture assertions fail for
missing initial goldens; no existing PNG needs refresh. The accepted trigger is
the validated menu interaction/containment remediation and its new open-state
regression coverage. This creates the eight authorized new goldens; it does not
move, delete, or re-account an existing golden. Renderer, masks, old viewports,
and existing fixture registry remain unchanged. The new capture declarations
explicitly include 200% CSS zoom and text spacing.

Canonical update attempt `20260905T132126Z-p10627` failed the private-evidence
permission check on an existing PNG copied into snapshot-scratch. No tracked PNG
or manifest changed. Existing golden permissions were restricted to 0600, with
SHA-256 equality checked for all 94 files, so the canonical copy retains private
permissions. The updater is being rerun; no harness check was bypassed.

The full `make browser-e2e-webserver-backed` attempt at
`20260905T131531Z-p7142` failed (42/60 work units). Three more existing role-status
helpers matched both the new menu and its trigger; they now request the button
role. The incident creation row passes at `20260905T132318Z-p53153`, Timeline
reviewer navigation at `20260905T132321Z-p54032`, and the three affected Auth rows
at `20260905T132442Z-p98558`. Broader failures in creation, relationship display,
and mention continuity are tracked separately from those helper corrections.
Untouched baseline runs reproduce the Auto/auto mismatch at
`/tmp/cartulary-account-menu-baseline/.cartulary/test-results/20260905T132244Z-p60482`
and the Display Name/Incoming Owner creation-error expectations at
`/tmp/cartulary-account-menu-baseline/.cartulary/test-results/20260905T132245Z-p62034`.
The former run's mention-resolution row passes; a current-source narrow rerun is
pending to resolve its earlier continuity failure. Merge navigation now passes
its account-role check but reaches an unrelated resolved-label expectation at
`20260905T132319Z-p53406`; a baseline comparison is pending. These findings do
not authorize reopening the completed relationship or creation workstreams.

Current-source mention continuity passes in the narrow row at
`20260905T132629Z-p44736`; the broader occurrence was transient. The stale
incident/membership navigation row passes at `20260905T132629Z-p44823`.
The merge resolved-label failure also reproduces at untouched baseline run
`/tmp/cartulary-account-menu-baseline/.cartulary/test-results/20260905T132629Z-p44695`.
No relationship/creation implementation or expectation was changed.

The private-permission repair allowed the next canonical update to execute,
but its route-focus assertion found an in-scope race at
`20260905T132628Z-p43362`. The early cancellation added during the lifetime audit
ran before asynchronous route publication. Requests now carry the originating
context and account identity: a menu-close render preserves the request, while a
different context/account/capability cancels it. The corrected real-browser row,
including held-Enter repeat suppression, passes at `20260905T133054Z-p27197`.
Menu/shared tests pass at `20260905T133055Z-p28738`; typecheck/imports/finalize
pass at `20260905T133114Z-p72493`, `20260905T133125Z-p75178`, and
`20260905T133129Z-p75533`. No unsuccessful update promoted tracked files.

Canonical promotion passed at `20260905T133220Z-p81501` (102/102 reconciled
captures). All eight new images were inspected individually. The workbook root,
Controls at 1024×720/768×640/640×480, long-label 200% zoom, and text-spacing
captures show contained panels and unobscured focused items. Clipped inactive
content in short panels is reachable through internal scrolling.

Review rejected the two context backgrounds as fixture-dependent: the directory
included earlier tests' incidents and administration included random actor emails.
The visual fixture now filters to its own named incident and current worker user
before opening each root menu. This changes fixture inputs, not runtime search
behavior or masks. Its ordinary comparison and follow-up canonical refresh are
pending. Four pre-existing goldens also received byte changes during the full
updater (presence marker ordering and small rendering/timing differences in
entity chips, network inspector, and pending replay). Their old/new images were
inspected; these are outside this seam and will be restored to their baseline
bytes. No unrelated visual refresh is accepted.

Fresh full accessibility passes at `20260905T133504Z-p34009`. App integration
rows, including heading focus after route publication and Account settings focus
after profile loading, pass at `20260905T133220Z-p81374`. Fixture typecheck passes
at `20260905T133808Z-p29297`.

Ordinary comparison `20260905T133748Z-p82504` completed every functional
assertion and reconciled all 102 active goldens with no missing/orphan/ambiguous
mapping. Only the intentionally filtered directory/admin backgrounds differ
(18,303 and 7,088 pixels). The remaining six menu images and all existing images
match under ordinary comparison. This reviewed ordinary artifact authorizes the
follow-up canonical refresh of those two new fixtures. No masking, viewport,
zoom, renderer, registry identity, or screenshot scope changed.

The follow-up refresh at `20260905T134118Z-p35642` stopped before the first
new menu capture: the initial document request returned HTTP 404, confirmed in
its retained Playwright trace. This happened while another source-triggered
browser build was starting; a shared preview/build race is the likely cause,
not a menu assertion failure. No files were promoted. The unchanged-source
updater is rerunning without competing builds. The expanded functional row now
also verifies long-label directory/admin roots at 640×480 and passes at
`20260905T134348Z-p82301`.

Final canonical refresh passes at `20260905T134540Z-p32270`. The two filtered
root captures were inspected again and accepted: one fixture incident and one
worker user provide deterministic backgrounds, with current-item focus visible.
The other six menu image hashes are unchanged from their individual review.
Unrelated updater changes were rolled back to HEAD. SHA-256 comparisons confirm
all 94 pre-existing PNGs are byte-identical to HEAD; only eight new menu PNGs
remain in scope. `make generate` regenerated the manifest after that rollback
at `20260905T134908Z-p79206`; finalization passes at `20260905T134917Z-p82118`.
The first fresh ordinary visual pass succeeds at `20260905T135012Z-p85494`
(12/12 work units, all 102 captures reconciled). The second fresh pass succeeds
at `20260905T135259Z-p33624` with the same 12/12 result and reviewed manifest.

Accepted menu capture filenames under
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`:

- `account-menu-workbook-root-linux.png` — 1280×720, first root action focused.
- `account-menu-directory-root-linux.png` — 1280×720, current Incidents action.
- `account-menu-deployment-root-linux.png` — 1280×720, current administration.
- `account-menu-controls-narrow-linux.png` — 1024×720, last Controls action.
- `account-menu-controls-compact-linux.png` — 768×640, expanded Controls.
- `account-menu-controls-short-linux.png` — 640×480, internally scrolled focus.
- `account-menu-long-label-zoom-linux.png` — 1280×720, 200% CSS zoom.
- `account-menu-long-label-text-spacing-linux.png` — 768×640, text spacing.

All eight belong to
`web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc`,
scenario `scenario_d840806c01cd`, Chromium, the pinned renderer profile recorded
by reconciliation. They are active nonregistry captures, so stable D-VFIX IDs
are not applicable. Existing D-VFIX fixtures were exercised unchanged.

AM-04 completed. Both fresh ordinary passes satisfy the visual-maintenance exit.
Full accessibility, focused menu/shared/App checks, browser keyboard/geometry,
nonadmin navigation, System views, and extracted-consumer regressions pass.
All menu-related helper failures are resolved. Remaining broader creation and
relationship failures reproduce at baseline; the transient mention-continuity
failure passes in a narrow current-source run. They are documented limitations
of the broad run, not unresolved menu dependencies. Next: AM-05 acceptance,
terminal validation, and scope audit. No further seam is authorized.

## Advisory classifications

- ADOPT R001/R002/R013/R014/R015: keyboard parity, visible focus, semantic
  controls, safe navigation, and evidence-backed completion.
- ADAPT R017/R018/R035: desktop viewport bands, owned scrolling, zoom and long
  text resilience. Preserve dense graphite tokens and existing grid geometry.
- REJECT R026–R034 where relevant: alternate theme/framework, invented product
  behavior, incidental selector authority, and generated design systems.
- ADOPT R003–R005/R011: current cues, names, reduced motion, and existing tokens.
- ADAPT R019/R023: retain desktop text tokens and existing Lucide icon mapping.
- Classifications refer to the existing `rules.tsv` ledger; no external query
  or new design system was needed for this seam.

## Acceptance and finalization

AM-05 entered after AM-04 was recorded DONE. `make agent-finalize` passes at
`20260905T135613Z-p80485`, before final static/generation checks. RESULTS_DIR
remains unset; retained-run-only maintenance is skipped for the documented
absence of an exact-source successful full warm check.

The following dispositions apply to this seam, not repository-wide completion.
Evidence IDs below refer to the detailed retained-run log above.

| Acceptance | Disposition | Reason and evidence |
| --- | --- | --- |
| A001 | PASS | Adopted owner-to-change map above; digest remains advisory. |
| A002 | PASS | One menu seam; extraction, behavior remediation, and destination integration are identified separately. |
| A003 | PASS | Clean main/HEAD recorded; source inventory, seven consumers, generated paths, and owner routing inspected and revalidated. |
| A004 | PASS | Existing 18rem menu width retained; spacing, elevation, colors, and focus use current tokens; no new registry. |
| A005 | PASS | Graphite presentation preserved; full visual passes with all 94 existing PNG bytes unchanged. |
| A006 | UNCHANGED SCOPE | No density selection, row geometry, editor geometry, or density projection changed; existing visual/a11y fixtures pass. |
| A007 | UNCHANGED SCOPE | No creation capability, payload, or form behavior changed. Broader stale creation-error expectations reproduce at baseline. |
| A008 | PASS, BOUNDED | Menu viewport clamping and viewport fallback use current measurements; six workbook viewport/height cases, zoom/text spacing, and narrow directory/admin roots pass. Existing responsive geometry row passes unchanged. |
| A009 | PASS | Panel owns scrolling; trigger and focused items stay inside viewport. Browser asserts no workbook/document scrolling from menu movement. |
| A010 | UNCHANGED SCOPE | Inspector dispatcher and feature-group behavior unchanged; open inspector remains present through nested/root Escape. |
| A011 | PASS, BOUNDED | Menu/drawer/route focus continuity passes; System views regression and narrow mention continuity pass; no grid continuity algorithm changed. |
| A012 | UNCHANGED SCOPE | No transaction ID, mutation, or replay implementation changed. |
| A013 | UNCHANGED SCOPE | No blocked queue recovery changed; pending-replay golden is byte-identical to baseline. |
| A014 | PASS, BOUNDED | Handled menu keys stop propagation; unsupported modified keys remain available; inspector Escape isolation and existing overlay regressions pass. Editing/paste behavior is unchanged. |
| A015 | UNCHANGED SCOPE | No conflict presentation or draft/saved-value behavior changed; existing visual/a11y evidence passes. |
| A016 | UNCHANGED SCOPE | No asynchronous resource-state contract changed. Menu invalidation/availability is tested separately; existing state fixtures pass. |
| A017 | UNCHANGED SCOPE | No background-refresh retention policy changed; existing refresh visuals pass. |
| A018 | UNCHANGED SCOPE | No evidence lifecycle, overlay, or preview behavior changed; existing visual/a11y fixtures pass. |
| A019 | PASS | Full accessibility `20260905T133504Z-p34009`; native keyboard/focus, accessible names/descriptions/current state, contrast, no extra live announcements. |
| A020 | PASS, BOUNDED | Closed/root/Controls, selected root/section, capability/context changes, removed items, pointer/keyboard, and dismissal paths are covered by 10 menu and 8 shared-hook tests. |
| A021 | UNCHANGED SCOPE | No virtualized rows, result-size policies, or measurement code changed. Grid measurement target is outside this menu seam. |
| A022 | PASS | Eight individually reviewed new captures; two fresh ordinary visual passes at `20260905T135012Z-p85494` and `20260905T135259Z-p33624`; 102 captures reconcile, 26 registered fixtures remain unchanged. |
| A023 | PASS | Stable action/section keys and UI-contract panel selector; selector package slice `20260905T130543Z-p3329`; semantic button targeting fixes ambiguous helpers. |
| A024 | PASS | New executable code and routing have no Markdown/digest dependency; human review supplies owner traceability. |
| A025 | PASS | Authored owner/routing inputs updated; Make generators/updater own downstream routing and golden manifest. Policy, shape, and drift checks recorded below. |
| A026 | PASS | Routes, callbacks, capabilities, API/schema, storage, and backend authorization remain in existing owners. |
| A027 | PASS | Final scope, commands, results, retained evidence, baseline limitations, skips, and rollback boundaries are recorded; AM-05 is complete. |

## Final architecture and scope

Source ownership is `web.app` for application components and `web.shared` for
the extracted React hook. Verification ownership remains `web.application` for
those focused tests. Workbook consumers keep their own owner policies.
`packages/ui-contracts` receives only a semantic test-selector contract.

The menu owns `closed | root | controls` and semantic action dispatch. The shared
hook owns registered item movement and live-reference restoration. Reconciliation
is opt-in; the menu suppresses subject-change and unmount restoration. Existing
consumers retain their declared Tab policy and explicit unmount restoration.
Outside/focus departure dismiss without restoration. Native button click is the
single activation path and repeated Enter is suppressed.

App owns account/incident identity, Account settings opening/explicit-close focus,
and heading handoffs. Pending route focus carries account and originating-context
identity, survives delayed route publication, and cancels on a different owner or
destination. Controls retains `onSelectSection(section, returnFocusTarget)` and
receives the live account trigger; its existing drawer owns opening/return focus.
The new menu/handoff mechanics use no focus timers, document-wide target
searches, compatibility wrappers, new packages, or workbook-private imports.

Fixed menu positioning preserves adjacent DOM/Tab order and escapes clipping
ancestors. Visual viewport bounds, anchor/panel measurements, CSS zoom ratios,
resize/scroll observers, and token padding contain it; focused-item scrolling
changes only the panel. Header metadata can shrink so the trigger remains usable.

The final diff contains these bounded paths (44 status paths, including both
sides of the two file moves):

- Application: `apps/web/src/app/AccountApplicationMenu.tsx`,
  `AccountApplicationMenu.test.tsx`, `useAccountMenuLayout.ts`, `App.tsx`,
  `App.landing.test.tsx`, `LandingAdminLayout.tsx`, `landingAdminStyles.ts`,
  and `landingAdminTypes.ts` in that same directory.
- Shared extraction: `apps/web/src/shared/useRegisteredOverlayNavigation.ts`
  and `.test.tsx`; both old matching files under
  `apps/web/src/workbook/focus/` are removed.
- Direct consumer imports: `SavedViewActionPanel.tsx`, `SystemViewSwitcher.tsx`,
  `WorkbookColumnsControl.tsx`, `WorkbookFiltersControl.tsx`,
  `WorkbookShellTopBar.tsx`, `WorkbookSortControl.tsx` under
  `apps/web/src/workbook/components/`, and
  `apps/web/src/workbook/timeline/components/TimelineRowActions.tsx`.
- Browser fixtures/helpers: `auth-and-incident-directory.spec.ts`,
  `incident-administration.spec.ts`, `timeline-workbook.spec.ts`,
  `workbook.a11y.spec.ts`, `workbook.visual.spec.ts`, and
  `support/entities/merge.ts` under `apps/web/e2e/`.
- Selector projection: `packages/ui-contracts/src/applicationSelectors.ts`
  and `application-selectors.test.ts`.
- Authored routing: `tools/frontend_source_ownership.json`,
  `tools/test_catalog_row_migrations.json`, and
  `tools/test_families/{web.application,web.workbook,web.design}.json`.
- Generated routing/evidence inventory: `tools/browser_e2e_batch_manifest.json`,
  `tools/execution_topology_render_index.json`, and
  `tools/frontend_visual_golden_manifest.json`, regenerated through Make.
- The eight new PNGs listed in the refresh record and this handoff.

No adopted owner, digest, dependency, lockfile, API/schema, backend, stored data,
transaction, grid editing, inspector dispatcher, or existing golden changed.
Branch and HEAD remain `main` / `f1d06b667c8ed693665128dd62a59144c837722f`.
Nothing is staged, committed, or pushed.

## Reproduction and verification commands

Run from the repository root. Resolve future routing through `make help`,
`make help-all`, and `make task-guide ROLE=module-author OWNER=<owner>`.
The semantic IDs below are recorded execution inputs, not new authority.

```sh
make test-slice OWNER=web.application ROWS=web.application.regression.account_application_menu_keyboard_focus_and_life_e79d61a853,web.application.regression.registered_overlay_semantic_focus_navigation_42dc7fdfee
make test-slice OWNER=web.application ROWS=web.application.regression.app_landing_keeps_explicit_incident_directory_na_c4fd880324,web.application.regression.app_landing_opens_account_settings_from_the_acco_12ecd21c1b,web.application.regression.app_landing_passes_workbook_incident_controls_th_f0ff8ac3c4
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.grid_controls_component_a104000002,web.workbook.regression.timeline_row_action_overlay_opens_from_pointer_a_d6ad456872,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e
make test-slice OWNER=package.ui ROWS=package.ui.frontend_unit.application_selector_contracts_675995f2ab
make service-backed-test-slice OWNER=web.design ROWS=web.design.browser.account_application_menu_keyboard_navigation_and_8cba739220
make service-backed-test-slice OWNER=web.design ROWS=web.design.accessibility.account_menu_accessible_keyboard_and_focus_state_3c31f28451,web.design.browser.workbook_responsive_frame_geometry_e78e1d1ac3
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.verify_system_views_switcher_keyboard_entry_rovi_90a2f62956
make service-backed-test-slice OWNER=module.auth ROWS=module.auth.browser.the_ordinary_shell_keeps_deployment_user_adminis_8a57567508
make browser-e2e-a11y
make browser-e2e-visual
```

The corrected navigation/helper rows and the passing current mention row
are reproduced with these exact inputs:

```sh
make service-backed-test-slice OWNER=module.incidents ROWS=module.incidents.browser.the_ordinary_authenticated_landing_screen_create_fa80fbfef5,module.incidents.browser.a_selected_incident_whose_membership_is_removed_21054ae899
make service-backed-test-slice OWNER=module.auth ROWS=module.auth.browser.a_reviewer_visible_incident_edit_control_fails_w_b6248bf4c8,module.auth.browser.deployment_admin_revoke_all_invalidates_a_target_a09ed5d06f,module.auth.browser.verify_ordinary_login_incident_directory_entry_o_d66ea23dc1
make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.a_reviewer_session_browser_flow_visibly_performs_c3aaf62d89
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.the_browser_inspector_resolves_existing_entities_325548131d
```

`make browser-e2e-webserver-backed` is the broader failed run, not a claimed
pass. Its remaining baseline failures are reproduced by the corresponding
public slices in the clean detached baseline worktree:

```sh
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.the_browser_inspector_merges_duplicate_entities_d0cb95b021,module.entities.browser.the_browser_workbook_shows_auto_resolution_only_758d41a16f
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.the_browser_workbook_exercises_parties_notes_tas_8ca985fff2,module.workbook.browser.the_browser_workbook_opens_communications_log_ha_e14517982d
```

Use the retained logs above to distinguish menu/helper corrections, fixture
setup failures, baseline product/test mismatches, and transient infrastructure.
The scratch worktree remains at `/tmp/cartulary-account-menu-baseline` with its
baseline evidence. Its Make invocations used the documented NODE_PATH bootstrap
fallback; no baseline source was edited.

## Compatibility, rollback, and limits

Compatibility is UI interaction remediation only: no runtime-data or API
migration. Roll back the application/menu integration and hook extraction together
with the seven import updates, selector/routing changes, generated routing,
eight new goldens, and their generated manifest entries. Existing golden bytes
are unchanged and need no replacement.

There is no unresolved in-scope implementation work. Known broad-run limitations
are the baseline creation-error and relationship-label expectations documented
above; they must not be presented as passed. No separate product repair is
implied. The earlier mention-continuity and document-404 failures were transient
and later evidence passes without changing those owners.

Skipped: full warm `make check`, `test`, release/CI gates, stateful backend suites,
and the grid-only measurement target. No backend, grid algorithm, virtualization,
or performance claim changed; focused consumer/geometry evidence plus full
accessibility/visual gates cover this seam. `make explain-target
TARGET=browser-e2e-measurement DETAIL=summary` resolved its seven Timeline/Network
Flow measurement rows and confirmed that ownership. No full frontend unit sweep
was substituted for the narrow affected consumers. Retained-run performance,
run-health, and scheduler maintenance remain explicitly skipped because
RESULTS_DIR is unset. No skipped check is claimed as passed.

Final terminal checks and AM-05 completion are recorded below.


AM-05 final static checks pass after finalization:

| Command | Retained run under `.cartulary/test-results/` |
| --- | --- |
| `make frontend-typecheck` | `20260905T135755Z-p84214` |
| `make frontend-import-boundary-check` | `20260905T135755Z-p84222` |
| `make lint-biome` | `20260905T135755Z-p84235` |
| `make generated-artifact-policy-check` | `20260905T135755Z-p84098` |
| `make generate-drift` | `20260905T135755Z-p84089` |
| `make json-shape-check` | `20260905T135755Z-p84108` |
| `make test-catalog-check` | Exit 0 in the same final Make invocation; also passed inside finalization. |

The final Markdown/scope command log is retained at
`.cartulary/account-menu-final-validation.log`, and the explicit scope inventory
at `.cartulary/account-menu-final-scope.json`. These human review artifacts do
not provide product/conformance authority. Initial terminal Markdown lint passed at `20260905T140354Z-p90134`;
`git diff --check`, the 44-path allowlist audit, index review, and 94-PNG
baseline comparison also passed. After this final handoff edit, the same terminal
checks are rerun and retained at the stable log/inventory paths above.

AM-05 DONE. All in-scope exits are satisfied; no owner contradiction or unresolved
menu work remains. No commit, push, migration, or next seam was performed.
The broader baseline failures and retained-run maintenance skips remain explicit.
Next action: none within this authorized seam.
