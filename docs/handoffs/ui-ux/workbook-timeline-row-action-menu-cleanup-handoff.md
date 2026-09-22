# Timeline row-action menu cleanup

## State, authority and scope

Working branch: `main`. Baseline and unchanged HEAD:
`dab4560febe1c6462f5304c067691a374cbb2734`. The initial working tree was clean.
The resulting uncommitted changes belong to this slice; no user changes were
removed. No commit, push, deployment, dependencies, endpoints, persistence or
migration were introduced. Implementation, production verification and handoff are complete. All applicable
acceptance rows pass; scoped N/A decisions and platform limits are recorded below.

Read `AGENTS.md`, the digest's `START_HERE.md`, `LOCAL_AGENT_PROMPT.md`, maps,
rules, acceptance and query recipes. Embedded upstream instructions remain
reference material. Current source guides, authored source/import manifests,
verification families and generated-artifact policy were revalidated. The digest
is dated navigation, not behavior authority. Current stack includes React 19.2.5,
RDG 7.0.0-beta.59, TypeScript 6.0.2, Vite 8.0.8, Vitest 4.1.4, Node 24.15.0 and
pnpm 10.33.0. Vendor imports remain inside Grid Adapter. The new hook's tests
live at the Timeline root, matching the current controller isolation boundary;
the earlier hooks-directory test placement was corrected, not exempted.
Localization metadata dates to 2026-09-13 at `59fa79e04035a4f29251fcb2337376f29c40f1f4`;
the inspected stack versions still match, while current owner/source/routing
inputs and the new semantic focus capabilities supersede that navigation snapshot.

| Behavior | Adopted owner | Source and verification ownership |
| --- | --- | --- |
| Native clipboard, editors, composition, local keys and accepted departure | Core 03 §11.1 REQ-03-147; §13 REQ-03-217/218/219/300 | Source `web.workbook` Timeline editing/keyboard and Grid Adapter; verification `web.workbook`, `module.timeline`, `package.grid_adapter` |
| Menu parity, one-layer Escape, semantic fallback, retained range | Core 03 §13, including §13.4; design §§8.4–8.6, §12.6, §14 | Source Timeline row-menu hook and presentation, `web.shared` overlay navigation, Grid Adapter; verification `web.application`, `web.workbook`, `module.workbook`, `package.grid_adapter` |
| Explicit Inspector and History destinations | Core 03 §13 and design §§8.5, 12.6–12.7 | Existing Timeline Inspector registry, selection and History owners; menu activation requests their destinations |
| Mark reviewed and Supersede | Core 03 REQ-03-104/106; existing Timeline capture owner | `timeline/actions`, `captureActions.start`, existing eligibility/preparation/transport; regression and existing browser accessibility rows retained |
| Scroll, refresh and authority invalidation | Selected user lifecycle policy within design §§8.5–8.6 and existing authorization owners | One Timeline invocation scoped to accepted query, incident lifetime, role, authorization epoch and closed state; existing presentation membership and shell authorization remain authoritative |

No adopted-owner contradiction was found. Existing tests expecting focused menus
to survive scrolling were implementation evidence and were corrected under the
user's explicit policy. Draft rows remain outside committed-record actions.

## Advisory query and selection rubric

Completed offline query from `QUERY_RECIPES.md`:
`keyboard focus context menu --domain ux -n 4 --json`, using the bundled search
script with Python bytecode disabled. Four results were returned. ADOPT R001
for keyboard parity and no traps; ADOPT R002 for visible, unobscured semantic
focus. REJECT R033 as a source of new behavior authority and R034 incidental
selectors. No upstream design system, palette, persistence or extra AAA claim
was generated. Semantic IDs, field/record identities, roles and existing grid
helpers select tests; Markdown never enters product inputs.

| Rubric decision | Assessment |
| --- | --- |
| Observed weakness | Confirmed: editor-insensitive pointer invocation, keyboard dispatch before interactive ownership checks, missing focus-exit wiring, focused-menu scroll survival and menu-induced Inspector selection. Captured DOM return targets were a structural weakness. Production later confirmed 200% zoom clipping, delayed pre-invocation scroll dismissing a new menu, and secondary pointer focus accepting a dirty collection departure before contextmenu. Native OS menu usability remains a manual-check limitation. |
| Remediation and change areas | Correct menu admission/lifetime in implementation and tests; separate semantic invocation from Inspector selection. Update source guides, authored routing and generated topology. No owner specification changes. |
| Responsible boundary | Timeline decides whether a gesture belongs to row actions and whether its invocation remains valid. Overlay owns registered item navigation; Grid Adapter owns semantic focus/reveal. Source owners retain action meaning and mutations. |
| Rationale and benefit | One admission decision prevents pointer/keyboard divergence. One semantic invocation eliminates detached-element return targets and accidental Inspector retargeting. Cause-specific dismissal yields to newer destinations. |
| Future extension | Another Timeline entry point can call the same admission owner; a semantic overlay consumer can opt into the shared restore callback without a second menu store or generic router. No speculative surface was implemented. |
| Capability value | Keep Inspect/History for explicit investigation and capture actions for existing review workflows. Keep semantic presentation/focus ports for virtualization and authority checks; preserve independent range selection. |
| Retirement | Remove menu state and setter, captured invoking/fallback element refs, duplicate row-menu resolution, Inspector selection on open, and focused-menu scroll exception. Remove the estimated-height clamp after measured zoom failure. |
| Retention and compatibility | Keep existing scalar/collection authoring, paste settlement, autosave, range/bulk selection, History and capture transport. Other overlay consumers retain DOM restoration defaults. Grid focus selection preservation is opt-in; existing callers keep their behavior. |
| Risk of leaving gaps | Native interaction interception, menu-induced writes or draft displacement, stale action subjects and focus theft. Zoom clipping was observed; no unmeasured performance benefit is claimed. |
| Validation and delivery | Characterize, correct, verify production, finalize and assess in that order. Failures feed bounded corrections. Commands/artifacts and remaining limits are below. Rollback reverts this slice's source/routing/generated outputs together; no data migration. |

## Implementation and lifecycle

`useTimelineRowActionMenu.ts` holds the single invocation: committed semantic
cell, position, accepted scope and grid offset at invocation. Both event paths
reject handled events, native controls/editors, editable content, embedded
controls, pickers, overlays, composition and drafts. Only admitted gestures
prevent default and open the menu. Current rows and presentation membership,
not saved DOM nodes, supply current action subjects. An admitted secondary
pointer-down prevents the browser's intermediate cell focus so the eventual menu
can borrow focus directly through the existing external-action marker. This
prevents a dirty collection input from accepting departure before contextmenu.
Native targets and already-handled pointer gestures retain their ownership;
primary interaction and accepted departure remain with existing owners.

`TimelineRowActions.tsx` connects item focus, focus exit, explicit destination
activation and dismissal causes to the shared registered overlay. Its external
action marker reuses editor focus borrowing. Inspect/History request their
existing semantic Inspector panels; capture actions still call their existing
owner. Accessible descriptions state which review/destination opens.

`GridHandle.requestFocus` adds optional `preserveSelection`, implemented in both
production and support adapters. Escape restoration uses cancellable requests;
new pointer, keyboard, focus or authority intent cancels the entire fallback
chain, including between requests. Support presentation now exposes the same
membership capability used in production.

| Cause | Implemented result |
| --- | --- |
| Escape | Close one menu; resolve current invoking cell, then row's first presented field, grid root, active surface selector or incident identity. No editor entry or completed-range replacement. |
| Outside pointer/focus; Tab or Shift+Tab | Close without restoring or preventing native destination interaction. No trap. |
| Action activation | Close; only the activated existing action dispatches. Inspector/History destination focus wins. |
| External scrolling | Close even while an item is focused; recover owned focus only to the grid container, without restoring the previous viewport. A queued scroll whose offset already existed at invocation does not represent a new movement; native input-internal scroll on blur does not move the invoking cell and is ignored. |
| Menu-internal scrolling | Keep open; item navigation reveals within the overlay. |
| Resize | Close with the same conditional non-scrolling fallback. Window and visual-viewport listeners are installed during layout. |
| Compatible refresh | Retain the same row/field invocation, resolve current row version and eligibility, reconcile disabled focused items. |
| Hidden/collapsed membership, removal, scope or authority change | Invalidate menu and pending return requests. Preserve outside focus; a hidden invoking field falls back to the same presented row, while absent rows and changed scopes recover only within the current readable grid/shell. |

Production geometry showed a focused item ending at x=1860 in a 1280px viewport
at 200% zoom. Measurement now converts viewport coordinates to the menu's CSS
coordinate space and clamps measured width/height with scrollable bounds. Existing
tokens, theme, action order and action labels remain. No general positioning or
event-routing framework was introduced.

Changed source groups: Timeline menu hook; Inspector/keyboard hooks; workflow and
workbook composition; presentation; row-menu component; optional work-area
pointer callback in the shared layout; shared overlay hook;
Grid Adapter core/production/support focus. Tests change the existing workbook
support, shared overlay and adapter tests, add the Timeline lifetime test, and
extend production Timeline/browser accessibility files. Guides change alongside
those owners. Authored ownership and five test-family files precede generator
changes to `browser_e2e_batch_manifest.json` and
`execution_topology_render_index.json`.

### Changed files

```text
apps/web/e2e/timeline-grid-entry.spec.ts
apps/web/e2e/workbook.a11y.spec.ts
apps/web/src/shared/README.md
apps/web/src/shared/useRegisteredOverlayNavigation.test.tsx
apps/web/src/shared/useRegisteredOverlayNavigation.ts
apps/web/src/workbook/WorkbookShell.support.test.tsx
apps/web/src/workbook/layout/README.md
apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx
apps/web/src/workbook/timeline/README.md
apps/web/src/workbook/timeline/components/README.md
apps/web/src/workbook/timeline/components/TimelineRowActions.tsx
apps/web/src/workbook/timeline/composition/README.md
apps/web/src/workbook/timeline/composition/useTimelineInspectorWorkflowComposition.ts
apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts
apps/web/src/workbook/timeline/hooks/README.md
apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts
apps/web/src/workbook/timeline/hooks/useTimelineKeyboardController.ts
apps/web/src/workbook/timeline/hooks/useTimelineRowActionMenu.ts
apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx
apps/web/src/workbook/timeline/useTimelineRowActionMenu.test.tsx
docs/handoffs/ui-ux/workbook-timeline-row-action-menu-cleanup-handoff.md
packages/grid-adapter/README.md
packages/grid-adapter/src/SemanticDataGrid.tsx
packages/grid-adapter/src/core.ts
packages/grid-adapter/src/index.test.tsx
packages/grid-adapter/src/test-support.tsx
tools/browser_e2e_batch_manifest.json
tools/execution_topology_render_index.json
tools/frontend_source_ownership.json
tools/test_families/module.timeline.json
tools/test_families/module.workbook.json
tools/test_families/package.grid_adapter.json
tools/test_families/web.application.json
tools/test_families/web.workbook.json
```

## Sequential exits and verification

All commands run from the repository root through public Make targets. This
shell requires the existing pinned runtime on PATH:
`PATH="$PWD/tmp/node-runtime/bin:$PATH"`. Artifact paths below are relative to
`.cartulary/test-results/`; each graph run contains `run-summary.json`.

1. **Characterize — exited.** Revalidated clean baseline and owners. The original
   two-row baseline `20260922T005743Z-p16810` failed before tests with
   `infra/service_start_error`. `make explain-run RESULTS_DIR=...` identified
   absent `node` on PATH. No harness code was changed. Red characterization at
   `20260922T010233Z-p19904` passed keyboard ownership and failed the required
   row-menu assertions. This is defect evidence, not acceptance.
2. **Correct — exited, then refined against production evidence.** Focused
   row-menu/keyboard tests passed at `20260922T011432Z-p31273`. Expanded tests
   and shared/adapter boundaries subsequently passed below. Production exposed
   the real grid's outer focused cell/inner field markup, zoom clipping and
   queued scroll ordering; each was corrected within this boundary. The final
   dirty collection cross-cell characterization also exposed intermediate native
   focus departure and text-input internal scrolling. The corrected scenario
   passes at `20260922T021443Z-p12797`; no authoring owner was modified. Inspector
   editing stayed explicit and its independent authoring survived menu opening.
3. **Production — exited.** The combined scenario passed at
   `20260922T021658Z-p55730`: native scalar/collection gestures, dirty collection
   cross-cell menu focus borrowing, undo/redo,
   menu invocation, Tab/Shift+Tab, destination focus, range/bulk preservation,
   dirty Inspector authoring, compatible refresh and disabled-item reconciliation,
   real virtualization/DOM replacement, removal, owner-observed authorization
   denial and live membership revocation. The corrected geometry/resize scenario
   passed at `20260922T015110Z-p46667`. Additional final text-spacing coverage
   is recorded below; full accessibility also passed.
4. **Finalize — exited.**
   `make agent-finalize` passed at `20260922T011929Z-p38997` and again at
   `20260922T013634Z-p28085`, `20260922T015629Z-p49254` and
   `20260922T021608Z-p51475` after the final collection-boundary correction,
   before final broad verification. `RESULTS_DIR` was unset: retained-run maintenance
   was skipped because no eligible successful full warm check was used.

| Command / selected owner rows | Result and run |
| --- | --- |
| `make task-guide ROLE=module-author OWNER=web.workbook` and changed shared/grid/browser owners | Current routing selected; also consulted `module.timeline`, `module.workbook`, `web.application`, `package.grid_adapter`, `package.ui` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_row_menu_lifetime,web.workbook.regression.timeline_row_action_overlay_opens_from_pointer_a_d6ad456872,web.workbook.regression.timeline_keyboard_event_and_focus_ownership_97dc4e9a17` | PASS 4/4 units, `20260922T021557Z-p50693` |
| `make test-slice OWNER=web.workbook` with selected committed autosave, capture authoring/characterization/recovery/transport, paste completion and duplicate pending-save rows | PASS 8/8 units, `20260922T021719Z-p11013`; earlier combined menu/compatibility run also passed 10/10 at `20260922T013109Z-p81727`; exact row selections in run manifests |
| `make test-slice OWNER=web.application ROWS=web.application.regression.registered_overlay_semantic_focus_navigation_42dc7fdfee` | PASS 2/2, `20260922T012254Z-p61959` |
| `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.overlay_focus_preserves_selection,package.grid_adapter.regression.semantic_focus_requests` | PASS 3/3, `20260922T015441Z-p46988`; production and support binding covered |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.collection_inspection_preserves_draft_text_and_s_f5c8c4c211` | PASS 2/2, `20260922T021719Z-p11049` |
| `make test-slice OWNER=package.ui` with application and selector-core rows | PASS 3/3, `20260922T014115Z-p43027` |
| `make browser-e2e-a11y` | PASS 20/20, `20260922T021658Z-p56333` (also `20260922T015734Z-p86111` and `20260922T013712Z-p62918`); production renderers and existing capture/Inspector/accessibility coverage |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95` | PASS 14/14, `20260922T014239Z-p76734`; existing budget retained, no latency-improvement claim |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.row_action_menu` | PASS 11/11, `20260922T021658Z-p55730` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.timeline_row_action_menu` | PASS 11/11, `20260922T020038Z-p59128` including text spacing (also `20260922T015110Z-p46667`) |
| `make json-shape-check` | PASS 3/3, `20260922T021658Z-p55837`; catalog/topology validation |
| `make frontend-typecheck` | PASS 2/2, `20260922T021658Z-p55996` |
| `make frontend-import-boundary-check` | PASS 2/2, `20260922T021658Z-p56071` after test relocation |
| `make lint-biome` | PASS 2/2, `20260922T021658Z-p56149` |
| `make lint-markdown` | PASS, `20260922T022153Z-p33468`; summary `adhoc/lint-markdown/tool-run-summary.json` |
| `git diff --check` | PASS; final branch and HEAD unchanged |
| `make generate` | PASS, `20260922T013845Z-p25741`; authored routing first, generated outputs never hand-edited |

Reviewed final production screenshots are attached to `20260922T021658Z-p56333`
in the full accessibility report under `browser-e2e-a11y/browser-groups/a11y-workbook-a11y/playwright-report.json`.
Decoded copies for convenient review are in that run's `unit-artifacts/`:
`row-menu-1280-720-zoom-2.png` and `row-menu-390-720-zoom-1.png`, alongside the
1440×900, 1024×720 and 768×640 captures. These are implementation support, not
new conformance claims or updated visual goldens. The final text-spacing captures
are also retained under `20260922T020038Z-p59128/unit-artifacts/`, with the same
names. The 200% zoom capture with increased line, letter and word spacing was
manually reviewed, including the final full-run 390px and 200% captures: all
actions and focus outlines remain readable and reachable.

### Failed attempts and disposition

- Typecheck failures at `20260922T010759Z-p23897`, `20260922T012241Z-p43903`
  and `20260922T012514Z-p79929`: unused removed prop, nullable version and missed
  support-adapter argument respectively; corrected.
- Focused intermediate support failures (`20260922T011000Z-p25533`,
  `20260922T011059Z-p26943`, `20260922T011154Z-p28545`) exposed missing
  presentation and unstable handle callbacks; corrected. Adapter range tests
  failed at `20260922T012254Z-p61918` and `20260922T012514Z-p79681`; corrected.
- Finalize/json-shape at `20260922T011615Z-p32993` and
  `20260922T011704Z-p33675` required regenerating changed routing. Generate
  attempts `20260922T013448Z-p17351` and `20260922T013546Z-p21367` required
  ASCII-sorted authored titles; corrected, then generated successfully.
- Formatter/lint at `20260922T011918Z-p38473` and `20260922T012254Z-p62187`
  found two new exhaustive-dependency errors; corrected. Existing informational
  template-string advice outside this slice was left untouched.
- Production `20260922T012241Z-p43779` exposed outer-cell keyboard admission;
  `20260922T012514Z-p79758` exposed zoom clipping. Both corrected.
- Scenario setup at `20260922T012514Z-p79725` and `20260922T013109Z-p81704`
  waited on a hidden column and a Details editor not explicitly opened; corrected.
- `20260922T013634Z-p28084` / `20260922T013830Z-p2599` established delayed
  pre-invocation scroll dismissal. The latter includes a bounded event-order
  attachment. Corrected by invocation-offset comparison; diagnostic code removed.
- `20260922T014115Z-p42980` / `20260922T014332Z-p8644` required the collection
  owner's semantic summary selector and virtualization helper. Corrected.
- `20260922T014334Z-p8869` failed the immediate resize assertion; listeners now
  install during layout with stable callbacks and the scenario establishes an
  open menu before resize. Follow-up `20260922T014536Z-p76703` showed callback
  replacement during resize; stable callback identity corrected it.
- Final dirty collection cross-cell characterization at `20260922T020457Z-p97648`
  showed a saved token after a secondary click moved focus before contextmenu.
  A shared-admission secondary pointer guard now prevents that intermediate
  departure; no collection save or settlement owner was modified. Follow-up
  `20260922T021033Z-p37239` preserved the draft but still closed the menu.
  Diagnostic `20260922T021238Z-p74817` established the sequence: secondary
  pointer, contextmenu, menu focus, native input-internal scroll, grid focus.
  Grid-descendant scrolling does not move the invoking cell and is now ignored;
  grid/page scrolling still dismisses. Temporary event instrumentation removed.
- `20260922T013711Z-p62612` import check rejected hook-test placement. Relocated
  to the existing Timeline test boundary; rules were not weakened.

Additional characterization: `20260922T014536Z-p76686` assumed a live remote role
update; the current authorization owner uses action revalidation. The scenario
now uses that existing path. `20260922T015110Z-p46625` attempted Mark reviewed
as an editor; Core 03 requires reviewer/admin, so the fixture was corrected.
`20260922T015348Z-p14247` assumed scrolling immediately unmounts RDG's selected
cell; RDG retains it until selection moves. The final real navigation scenario
verifies replacement without weakening the product focus assertion.

## Acceptance assessment

Assessment is scoped to menu admission/lifetime and the explicitly changed shared
capabilities; PASS does not assert untouched subsystem completeness.

| Row | Assessment | Evidence or scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | Exact owner map above; user-selected lifecycle corrections distinguished from advisory navigation and historic tests. |
| A002 Scope | PASS | Selection rubric, migrated composition ports, one invocation/admission owner, removed duplicate state and handlers; no generic router. |
| A003 Repository state | PASS | Clean baseline, unchanged HEAD, current manifests/guides and stack audit; localization drift recorded. No direct vendor imports added outside Adapter. |
| A004 Tokens | PASS | Existing margin/width and action styles retained; estimated height removed. Measured dimensions are geometry, not a second design registry. |
| A005 Theme | PASS | Existing dark_graphite theme and focus tokens retained; reviewed production captures. |
| A006 Density | N/A | No density selection, row/header height, gutters, typography tokens or editor geometry changed. Menu fixtures use the production default. |
| A007 Creation | N/A | No creation capability, route or payload change. Draft admission exclusion is verified without altering draft controls. |
| A008 Responsive | PASS | Real production menu at 1440×900, 1024×720, 768×640, below-minimum 390px width and 200% zoom; resize closes. Existing shell thresholds/Inspector clamps unchanged. |
| A009 Overflow | PASS | External grid scroll preserves its new viewport; menu-internal scroll does not dismiss; screenshots retain visible status/navigation outside menu geometry. |
| A010 Inspector | PASS | Explicit Inspect/History activation and destination focus; opening/dismissing a different row menu preserves dirty Inspector subject. Existing capture/feature owner regression and full accessibility coverage. |
| A011 Continuity | PASS | Native raw scalar/collection drafts, independent Inspector authoring, completed range/bulk selection, compatible version refresh, semantic return and new outside focus tested. |
| A012 Transactions | PASS | Menu open/dismiss produces zero record writes; explicit activation uses unchanged capture owner. Existing capture transport/recovery and duplicate-autosave dispatch tests pass. No transaction identity algorithm changed. |
| A013 Acknowledgement and recovery | N/A | No queue/replay/acknowledgement/recovery implementation changed. Existing autosave and capture recovery regression rows pass as compatibility evidence. |
| A014 Editing | PASS | Production right-click/Shift+F10/ContextMenu retain scalar focus/value/undo; collection native gestures and cross-cell menu focus borrowing retain token drafts without submission. Existing paste, acceptance, duplicate dispatch and collection settlement rows pass. |
| A015 Conflict | N/A | No conflict presentation, retained local/saved values or resolver behavior changed. Autosave regressions retained. |
| A016 Data/interaction states | PASS | Scope-limited admission requires current readable committed presentation. Current action eligibility resolves from refreshed rows; draft, native, handled and composition exclusions tested. No query data-state policy changed. |
| A017 Refresh/authorization | PASS | Compatible refresh and disabled-item reconciliation, hidden/collapsed membership invalidation, scope cancellation, server-denied downgraded action and live revocation tested. Existing account/session recovery owners untouched; remote-role observation limitation below. |
| A018 Evidence | N/A | No Evidence lifecycle, preview, transport, access-state or attachment change; its native/embedded controls keep ownership. |
| A019 Accessibility | PASS | Full `browser-e2e-a11y`, keyboard parity, one-layer Escape, Tab exit, destination focus, accessible action descriptions and viewport/zoom geometry. No new conformance claim. |
| A020 Components | PASS | Existing four actions and disabled reasons retained; item reconciliation and measured/clamped menu bounds tested. Text-spacing extension passed at `20260922T020038Z-p59128`. |
| A021 Virtualization | PASS | Production scroll dismisses without jumping; RDG's selected-cell retention is released by a newer real cell interaction, old invoking node disconnects, and semantic navigation/remenuing succeeds on replacement DOM. Existing focus-paint budget passes. |
| A022 Visual fixtures | PASS | Production accessibility geometry captures use semantic fixture IDs, declared viewports, default theme/density and 200% CSS zoom. Narrow/zoom screenshots manually reviewed. No golden changed; visual registry/maintenance guide inspected. |
| A023 Selectors | PASS | Existing UI-contract record/field/overlay IDs, roles and semantic grid helpers; selector contract rows pass. No new incidental product IDs or vendor coordinates. |
| A024 Test authority | PASS | Source/test dependency audit: no runtime, tests, generators or product checks read/stat/hash Markdown. Handoff and Markdown lint are human/document maintenance only. |
| A025 Generated artifacts | PASS | Authored families/ownership updated before `make generate`; json-shape and finalization pass. Only generator-owned topology projections changed. |
| A026 Compatibility | PASS | Existing mutation/authorization owners and supported overlay callers preserved; opt-in grid option. No schema, endpoint or persistence changes; retirement/rollback explicit. |
| A027 Handoff | PASS | Owners, changes, sequential exits, commands/results/artifacts, corrected failures, manual limitation, skipped checks, retirement and rollback are recorded. All applicable rows pass. |


## Limitations and rollback

Automated Chromium assertions establish non-interception, retained focus/draft,
clipboard/range behavior and native undo/redo. They cannot demonstrate the
operating system/browser context menu itself is usable. No interactive OS-menu
manual check was available; manually check right-click, Shift+F10/ContextMenu,
selection/copy/paste/undo inside scalar and collection inputs before claiming
that platform-specific usability. CSS zoom evidence is not an OS zoom-menu check.

Authorization invalidation follows the existing owner observations. A remote
role downgrade is not a new live role-update stream: the existing explicit action
receives authoritative rejection and suspends capture actions. The production
scenario verifies that path and separately verifies live membership revocation.
Known role/epoch/scope changes invalidate the menu and pending return requests in
focused lifecycle tests. This slice adds no authorization polling or mutation
bypass. Mark reviewed retains its existing immediate action semantics; Supersede
retains its existing explicit review workflow.

Next action: review the uncommitted diff and, where platform usability evidence
is required, perform the native-menu manual check above.

No broad visual golden update was needed; screenshots were inspected as support
evidence. No extra performance claim, backend/full release verification or data
migration is asserted. Rollback reverts this slice's source, tests, guides,
authored routing and generated topology together, preserving later user changes.
