# Workbook Timeline supersession Review handoff cleanup

## Baseline, authority, and reproduced gap

The slice began on clean `main` at
`712039be207c16ae4dad5c1234ea4d9e241bd7c8`, the cited assessment
baseline. The checkout had no unrelated changes. `AGENTS.md`, the localized
UI/UX digest's entry documents, owner/rule/acceptance maps and query recipes,
the Timeline actions and Inspector source guides, the Timeline capture-actions
handoff, and the current History read-continuity handoff and implementation were
read. The digest was navigation and advisory evidence. The narrow offline UX
query returned eight focus, loading, keyboard, and recovery entries. R001,
R002, R006, R008, R014, and R025 were adopted within the Core-owned flow;
R012 was adapted to the existing Timeline lifetimes. The bundled suggestion to
disable a loading button was not applied as native disabling because it would
break this Inspector's pending keyboard continuity. No design system was
generated, and no executable input depends on Markdown.
The current frontend manifest confirms React 19.2.5, TypeScript 6.0.2, and
Vite 8.0.8. Current source/import ownership and generated-artifact policy were
inspected. Direct `react-data-grid` imports remain in `packages/grid-adapter`;
the changed Timeline paths introduce none.

Core 01 REQ-01-083–087 and §7.4.1A own the dedicated supersession route,
optional replacement, reason, and required destructive confirmation. Core 03
§2.3A, REQ-03-106/107/282/292 own Inspector attachment, review invalidation,
committed-version admission, and terminal supersession behavior. Design
§§12.4, 12.7–12.8, and 14 own loading-button names, pending state, Inspector
scroll/navigation, local feedback, and keyboard/focus presentation. Source
ownership is `web.workbook` under `tools/frontend_source_ownership.json`;
`web.workbook`, `module.timeline`, and `module.workbook` route the selected tests
independently. No owner contradiction was found.

The old editor made Review natively disabled while `onPrepare` awaited the
replacement-validation read. It then set an unconditional focus flag and
replaced authoring controls with confirmation. A new held-preparation component
characterization failed at `.cartulary/test-results/20260926T175943Z-p40249`:
the connected Review button's `disabled` property was `true`. The same new
scenario was run through the production Timeline Inspector with the baseline
editor temporarily restored and a deterministic replacement-query gate. It
failed because the Review button became disabled and lost focus while the read
was held (`.cartulary/test-results/20260926T180828Z-p53531`; Playwright report
and trace in that run root). The working editor was restored immediately after
that run. This is a confirmed pending-interval defect, distinct from the
expected rejection when a material autosave advances the reviewed row version.

## Correction and ownership decisions

`TimelineSupersessionEditor` keeps the initiating Review button connected,
focusable, and named `Review supersession` throughout preparation. It shows
`aria-busy`, `aria-disabled`, and local neutral progress; a synchronous ref
rejects repeat Enter, Space, and click activation. Its editor-local intent
captures the initiating control, reviewed key, and starting focus. Newer
keyboard, focus, pointer, wheel, touch, or scroll interaction retires automatic
transition. Settlement checks the current preparation, key, attachment,
authority, control and interaction before mounting confirmation; the layout
focus step checks again before focusing Confirm. A retired transition leaves
reason and replacement mounted and requires fresh explicit Review. Input,
subject, version, scope, and authority changes abort stale authoring preparation;
Cancel and unmount do likewise. Failure copy covers committed-row and optional
replacement validation without assuming every wait concerns earlier edits.
Back preserves raw reason and selected replacement.

The Timeline action owner still reserves and validates preparation. Its
authoring-only reservation is now excluded from admitted mutation counts and
retained action recovery and is removed after either outcome. This prevents a
pending review read or cancelled authoring from being reported as an admitted
write. Owner-controlled `submit` is unchanged: only explicit Confirm enters
captured transaction, transport, uncertain replay, receipt, and read-only
refresh lifetimes. Cancellation of this handoff cannot cancel or replay a
previously admitted write. The existing History read-continuity hook was used
as an interaction precedent; its cursor, accepted pages, and shared recovery
lease were not imported into Timeline. No shared abstraction was warranted by
this editor-local presentation decision.

The retired paths are native pending disabling of Review, unconditional
`focusReview`, and global recovery/status reporting of authoring-only
preparation. Public routes, schemas, fields, permissions, reason normalization,
optional replacement semantics, immutable requests, storage, dependencies, and
goldens did not change. There is no migration or compatibility adapter. A
future matching workflow can use the same separation of owner admission and
presentation intent without importing this editor's state. The remaining risk
is browser behavior outside the selected Timeline Inspector profiles; the
focused checks do not assert whole-page accessibility conformance.

Changed authored code and tests are the Timeline supersession editor and
colocated test, Timeline capture owner and colocated test, and
`apps/web/e2e/timeline-workbook.spec.ts`. The authored family manifests are
`tools/test_families/web.workbook.json` and `module.timeline.json`; Make
regenerated `tools/browser_e2e_batch_manifest.json` and
`tools/execution_topology_render_index.json`. The Timeline actions guide and
this handoff are documentation-only outputs. No other source or generated root
was changed.

## Verification and artifacts

| Check | Result |
| --- | --- |
| Baseline `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_capture_authoring` | PASS 2/2 at `.cartulary/test-results/20260926T175026Z-p33781`; old tests lacked held-focus assertions. |
| New failing component and production browser characterizations | Expected FAIL at `.cartulary/test-results/20260926T175943Z-p40249` and `.cartulary/test-results/20260926T180828Z-p53531`; both failed on pending Review reachability. |
| Authored generation | `make generate` PASS after catalog edits, latest `.cartulary/test-results/20260926T181754Z-p33151`. |
| Focused editor and owner rows | `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_capture_authoring,web.workbook.regression.timeline_capture_recovery` PASS 3/3 on the final product tree at `.cartulary/test-results/20260926T181840Z-p40029`. |
| Existing sequencing row | `make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_8992eb8931` PASS 2/2 on the final product tree at `.cartulary/test-results/20260926T182022Z-p48496`; a held material autosave still requires fresh Review after its version changes. |
| Gated production and exact recovery browser rows | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.supersession_review_handoff,module.timeline.browser.capture_action_exact_recovery` PASS 11/11 at `.cartulary/test-results/20260926T181200Z-p24252`; the expanded gated scenario including a failed read and retry PASS 11/11 at `.cartulary/test-results/20260926T181353Z-p62359`. |
| Existing Timeline accessibility browser row | `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.timeline_capture_actions` PASS 11/11 at `.cartulary/test-results/20260926T181450Z-p94888`. |
| Typecheck and import boundaries | `make frontend-typecheck` PASS 2/2 at `.cartulary/test-results/20260926T180752Z-p52925`; `make frontend-import-boundary-check` PASS 2/2 at `.cartulary/test-results/20260926T181309Z-p56869`. |
| Formatting | `make format` PASS 2/2, latest `.cartulary/test-results/20260926T181744Z-p28762`. |

The first typecheck at `.cartulary/test-results/20260926T180329Z-p46665`
failed on a new test mock's argument types; the mock was typed and the rerun
passed. The first `make lint-biome` at
`.cartulary/test-results/20260926T181309Z-p56878` reported formatting in
three edited frontend files only; `make format` corrected it. The baseline
browser and component failures above were intentional characterizations, not
waived product checks. The final checks and any repair results are recorded
below before handoff completion.

## Digest acceptance and handoff

| Result | Acceptance rows and evidence |
| --- | --- |
| PASS | A001–A004: exact Core/design owner mapping, bounded editor decision and retirement, clean branch/HEAD and current source/stack/import boundaries, and no new token literal or registry. |
| PASS | A009–A011: 1024×720 Inspector scroll and controls, canonical row/review invalidation, and held keyboard/pointer/scroll continuity in the gated production browser. |
| PASS | A012–A014: duplicate admission, zero preparation writes, exact recovery regression, retained raw reason/replacement, and owner-only cancellation of authoring preparation. |
| PASS | A017: closure, subject/scope replacement, authority loss and late settlement fencing; no protected presentation revival. |
| PASS | A019–A020: stable accessible name, busy/unavailable cues, native Tab/Shift+Tab, local recovery, visible focus, 200% enlargement and narrow Inspector checks; these are not whole-page conformance claims. |
| PASS | A023–A027: semantic selectors, Markdown-independent tests, authored catalogs and Make-generated outputs, unchanged public compatibility, and evidence-backed handoff. |
| N/A | A005–A008: theme, density, creation and responsive threshold algorithms were not changed. |
| N/A | A015–A016, A018, A021–A022: conflict, grid data-state, Evidence, virtualization and visual-golden owners are outside this Inspector Review presentation change. |

No applicable acceptance row is blocked. No digest row was used as product
authority.

Rollback is a coordinated revert of the editor, Timeline owner, focused tests,
authored catalogs, generated topology, source guide, and this handoff. There is
no stored-data or route migration to reverse. The work remains uncommitted for
review. Full `make check`, `make ci`, release, broad visual-golden validation,
and publication checks are skipped because the changed boundary is the focused
Timeline handoff and no golden or public contract changed.

### Final checks

`make agent-finalize` PASS 1/1 at
`.cartulary/test-results/20260926T181809Z-p36094` before the final checks.
`RESULTS_DIR` was unset, so retained-run maintenance was skipped: no qualifying
successful full warm check run was supplied. On the final product tree,
`make frontend-typecheck` PASS 2/2 at
`.cartulary/test-results/20260926T181840Z-p40203`, `make lint-biome` PASS 2/2
at `.cartulary/test-results/20260926T181840Z-p40277`, and
`make frontend-import-boundary-check` PASS 2/2 at
`.cartulary/test-results/20260926T181953Z-p43703`.
`make test-catalog-check` PASS; `make generate-drift` PASS 4/4 at
`.cartulary/test-results/20260926T181953Z-p43373` and
`make generated-artifact-policy-check` PASS 3/3 at
`.cartulary/test-results/20260926T181953Z-p43424`.
`make lint-markdown` PASS after the evidence update at
`.cartulary/test-results/20260926T182150Z-p49681`.
