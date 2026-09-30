# Built-in workbook selector keyboard handoff

The desktop selector now has one sequential Tab stop. Left/Right move focus in
canonical registry order, wrapping at either end; Home/End focus Timeline/Notes.
Enter and Space use native button activation and the existing explicit selection
callback exactly once. Pointer selection remains available. Escape does nothing;
Tab and Shift+Tab leave normally. Arrow movement changes only presentation focus.
Implementation and focused product evidence are ready for review; private seeded
rendered review remains blocked as recorded below. No commit or deployment occurred.

## Boundary and authority

The checkout started clean on `main` at
`1083a006325671b3f722ef9a7bd3683815d1b351`; branch and HEAD remain unchanged.
Applicable AGENTS instructions, current source guides, authored source ownership,
import boundaries, routing families, and generated policy were inspected. All
working changes belong to this slice. The digest's older snapshot and historical
handoffs supplied navigation, not fresh execution evidence or product authority.
The authored frontend manifest uses React 19.2.5, TypeScript 6.0.2, Vite 8.0.8,
Vitest 4.1.4 and Playwright 1.59.1. No direct grid-vendor import was added; the
existing Grid Adapter remains the boundary. No localization drift changed this seam.

| Owner | Decision |
| --- | --- |
| Core 03 §2, §§2.1–2.2 | Canonical built-ins and System views; explicit selection enters the first visible authorized writable creation control, then an eligible committed navigation cell, then the grid root. Selection creates no record. Actual focus receipt completes entry; replacement, unmount, authority loss and newer deliberate navigation invalidate old requests. |
| Core 03 retained-authoring requirements, including REQ-03-299/100 | Existing draft, operation, acknowledgement, authorization and account/incident lifetimes remain with their owners. Selector focus adds no persistence or write admission. |
| Design §§7.4, 8.4–8.5, 14 | Existing responsive bands, manual horizontal tab navigation, explicit activation, Escape ownership, truthful accessible selection and visible focus. Wrapping and Home/End are the user's approved interaction choices, not additional adopted requirements. |
| Domain vocabulary; Core 00 precedence | `view_schema_id` remains canonical. Adopted owners govern; design governs its declared presentation scope. Digest rules R001/R002/R013/R014 guide this bounded implementation. |
| Source `web.workbook` | Top bar roving focus, panel relationships and Workbook fallback consumers. No shared helper, grid-vendor integration, backend or new identity owner. |
| Verification `web.workbook`, `module.workbook` | Frontend regressions and service-backed browser scenarios are separately routed. `module.entities`, `module.timeline` and `module.imports` checks cover their directly migrated test consumers only. |

The observed gap was five sequentially tabbable controls without an arrow handler.
The small top-bar mechanism now hides one real decision: which canonical control
receives focus while selection remains elsewhere. Keeping it local avoids coupling
horizontal navigation to overlay opening/closing. Registry-driven refs and IDs
allow a future owner-approved registry change without a second identity list.
No speculative surfaces, framework, shortcuts or layout thresholds were added.

## Changes and focus lifetimes

`WorkbookShellTopBar` derives outside entry from the selected built-in, otherwise
Timeline. System views and Network Analysis leave every built-in unselected.
Actual focus receipt, including pointer and fallback focus, updates the local
roving key. Leaving resets the next entry; passive selection rerenders never move
actual focus. The existing external-action marker preserves authoring focus borrowing.

The labelled horizontal tablist uses tabs with `aria-selected` and stable
schema-keyed controlled-panel IDs. The existing active frame becomes the selected
desktop panel; inactive panels are hidden empty shells. Only active content mounts.
Compact, System-view and extension active frames retain their region semantics.
The inset focus ring uses the existing border and twice the existing offset token
to keep its painted bounds inside the strip and viewport. Tab geometry and the
selected underline retain their prior presentation. The superseded all-buttons-tabbable path and built-in
`aria-current` semantics are retired without a compatibility shim.

Responsive characterization reproduced focus falling to `body` when the focused
branch was removed. The local layout effect synchronously focuses its replacement
only while selector focus is owned and no newer external control owns focus.
A retiring compact menu closes without restoring its removed trigger. Explicit
compact selection retires menu-item ownership before destination entry. This adds
no request owner: startup, semantic grid focus, entry generations, cancellation,
drafts and mutation recovery remain unchanged.

Changed paths, all relative to the repository root:

- Production: `apps/web/src/workbook/WorkbookShell.tsx`,
  `components/WorkbookShellTopBar.tsx`, `components/WorkbookActiveSurfaceFrame.tsx`,
  `layout/workbookShellStyles.ts`, `models/workbookSurfaceRegistry.ts` beneath
  `apps/web/src/workbook/`.
- Fallback consumers: `components/workbookFocusFallback.ts` and
  `timeline/hooks/useTimelineViewportContinuityController.ts` now find the selected
  tab. Coordinated fixtures: `components/WorkbookParkedGridDrafts.test.tsx` and
  `timeline/useTimelineViewportContinuityController.test.tsx`.
- Focused tests: new `components/WorkbookShellTopBar.test.tsx` and
  `WorkbookShell.surfaces.test.tsx`; source guide `components/README.md`.
- Browser coverage/consumers in `apps/web/e2e/`: `incident-administration.spec.ts`,
  `merge-recovery.spec.ts`, `import-assistant.spec.ts`,
  `indicator-canonical-create.spec.ts`, `keyboard.spec.ts`,
  `timeline-auto-resolution-feedback.spec.ts`, `timeline-deferred-continuity.spec.ts`,
  `timeline-grid-entry.spec.ts`, `workbook.a11y.spec.ts`, `workbook.visual.spec.ts`.
  Existing test identities remain intact; selected-state and role consumers migrate
  together. `data-workbook-tab-index` remains for its concrete registry-order test.
- Authored routing: `tools/frontend_source_ownership.json`,
  `tools/test_families/web.workbook.json`, `tools/test_families/module.workbook.json`.
  `make generate` produced `tools/browser_e2e_batch_manifest.json` and
  `tools/execution_topology_render_index.json`. No generated source or lockfile was
  hand edited. This handoff is documentation only; no executable input reads it.

## Verification

All commands ran from the repository root through public Make targets. Discovery
used `make help`, `make help-all` and `make task-guide ROLE=module-author OWNER=...`
for `web.workbook`, `module.workbook`, `module.entities`, `module.timeline` and
`module.imports`. The following exact row sets preserve the existing identities:

```bash
unit_rows=web.workbook.regression.desktop_built_in_surface_selector_keyboard_navig_ebad53c0ea,web.workbook.regression.explicit_built_in_grid_entry_92c0a52a3d,web.workbook.regression.parked_grid_draft_focus_survives_removal_and_pre_bee6364d8c,web.workbook.regression.timeline_deferred_continuity,web.workbook.regression.conflict_editor_accessibility,web.workbook.regression.conflict_completion_preserves_newer_selection_444f6007dd
browser_rows=module.workbook.browser.desktop_built_in_surface_selector_keyboard_journ_3475c2c6d2,module.workbook.browser.built_in_selector_entry_from_network_analysis_8f5bebcd07,module.workbook.browser.built_in_grid_entry_editor_59f8176b4a,module.workbook.browser.built_in_grid_entry_supersession_5f9a05b19f,module.workbook.browser.built_in_grid_entry_viewer_c3bbf9714f,module.workbook.browser.verify_system_views_switcher_keyboard_entry_rovi_90a2f62956
compat_rows=module.workbook.browser.desktop_built_in_surface_selector_keyboard_journ_3475c2c6d2,module.workbook.browser.canonical_create_late_closed_replay,module.workbook.browser.host_paste_recovery,module.workbook.browser.identity_paste_recovery,module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159,module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0
timeline_rows=module.timeline.browser.auto_resolution_feedback_characterization,module.timeline.browser.clear_recovery,module.timeline.browser.deferred_continuity,module.timeline.browser.spreadsheet_batch_replay
import_rows=module.imports.browser.claimed_import_assistant_production_workflow_986c2117ee,module.imports.browser.keyboard_recovery,module.imports.browser.partial_outcomes
make test-slice OWNER=web.workbook ROWS="$unit_rows" CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=module.workbook ROWS="$browser_rows" CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.entity_merge_exact_recovery CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=module.workbook ROWS="$compat_rows" CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=module.timeline ROWS="$timeline_rows" CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=module.imports ROWS="$import_rows" CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.host_paste_recovery,module.workbook.browser.identity_paste_recovery CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=module.workbook ROWS="$browser_rows",module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159,module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0 CARTULARY_HARNESS_CACHE_MODE=off
```

Run roots below are under `.cartulary/test-results/`; `run-summary.json` and each
`rows/<row-id>.json` retain terminal evidence. Graph unit counts include setup units.

| Command | Result and run root |
| --- | --- |
| Focused units above | PASS, six rows; `20260930T003809Z-p33683` |
| Six workbook browser rows above | PASS, default and `network_flow_claimed`; `20260930T002208Z-p67325` |
| Entity merge compatibility row | PASS, four scenarios; `20260930T002208Z-p67333` |
| Timeline compatibility rows | PASS, four rows; `20260930T002757Z-p55042` |
| Import Assistant compatibility rows | PASS, three rows; `20260930T002757Z-p55049` |
| Additional workbook compatibility rows | Journey, canonical late replay and shell accessibility PASS; paste and visual failures below; `20260930T002757Z-p55027` |
| Paste compatibility retry | PASS, Hosts and Identities; `20260930T003355Z-p89990` |
| Final workbook browser/accessibility/visual slice | PASS, eight rows; `20260930T003809Z-p33681`; goldens unchanged |
| `make generate` | PASS; `20260930T001136Z-p34812` |
| `make format` | PASS; `20260930T003604Z-p24045` |
| `make agent-finalize` | PASS; `20260930T003632Z-p28616`; `RESULTS_DIR` unset |
| `make frontend-typecheck` | PASS; `20260930T003809Z-p33871` |
| `make frontend-import-boundary-check` | PASS; `20260930T003809Z-p33908` |
| `make lint-biome` | PASS; `20260930T003809Z-p33950` |
| `make generate-drift` | PASS; `20260930T001645Z-p12411` |
| `make generated-artifact-policy-check` | PASS; `20260930T001645Z-p12430` |
| `make json-shape-check` | PASS; `20260930T002128Z-p64746` |
| `make lint-markdown` | PASS; `20260930T004046Z-p76566` |

The production-top-bar red characterization failed five desktop cases, including
responsive focus loss, at `20260930T000138Z-p23322`; compact dismissal already passed.
Green units then passed at `20260930T000340Z-p24628` and
`20260930T000505Z-p25792`. Initial routing attempts were rejected for unsorted
titles and an unsupported interpolated title; sorted explicit titles fixed them.

Earlier `make agent-finalize` failed at `20260930T000924Z-p32768` because routing
projections were stale. `make json-shape-check` confirmed this at
`20260930T001014Z-p33316` and `20260930T001056Z-p34148`; `make generate` resolved it.
An attempted Make input `CARTULARY_OUTPUT_MODE=verbose` was rejected; the diagnostic
rerun correctly used an environment assignment. Typecheck at
`20260930T001250Z-p41995` found a nullable registry lookup; guarding it resolved the
failure. The browser slices at `20260930T001300Z-p43123` and
`20260930T001626Z-p83539` caught the half-pixel tab boundary. A content-height
adjustment passed at `20260930T002042Z-p28444` but shifted the selected underline:
the visual row in `20260930T002757Z-p55027` failed by 67 pixels. Manual artifact
review localized that change. The final implementation preserves prior geometry
and places the painted focus ring farther inside using existing tokens; the browser
assertion measures the ring's outer edge against both clipping boundaries.
The same compatibility run timed out on a remaining computed Hosts/Identities
button-role lookup; its coordinated tab-role migration passed at
`20260930T003355Z-p89990`. These failures were investigated, not accepted as passes.
The final eight-row run `20260930T003809Z-p33681` passed the revised focus-ring
checks, shell accessibility and the default/narrow/compact visual fixture without
changing goldens. No visual failure remains in the selected fixture.

Product browser observations: keyboard entry used an outside incident-details
anchor and actual Tab/arrow sequences, without manually focusing destinations.
Editor selection focused creation without creating records; viewer destinations
used populated-cell and empty-root fallbacks. Held destination mounting yielded
to a newer selection. Compact dismissal returned to Surfaces, explicit selection
entered its grid, and System views/Network Analysis retained their entry policy.
Timeline raw inspector authoring survived selection away and back. Focused controls
and token outlines were checked at 1440×900; responsive replacement was exercised
at 1024×720 and 125% CSS zoom at 1440×900. These checks cover those scenarios only,
not every zoom mechanism, whole-page accessibility or measured productivity.

## Rendered review and remaining work

`make ui-review UI_MODE=seeded REVIEW_PROFILE=default` failed before browser startup
at `20260930T002033Z-p28086` with
`infra/service_start_error/startup_failed`. The exact locator was
`/home/jochi/code/cartulary/.cartulary/test-results/20260930T002033Z-p28086/ui-review/session.json`.
`CARTULARY_OUTPUT_MODE=machine make ui-review-stop UI_SESSION=<that exact locator>`
preserved the primary failure; `ui-review/terminal.json` records `cleanup: complete`.
The foreground process exited. No request scratch, image, DOM snapshot or Axe
result was produced. The earlier planning failure at `20260929T234638Z-p15349`
also had exact-session cleanup; its historical result was not reused as fresh evidence.

An artifact-mode session at `20260930T003220Z-p80491` imported the exact failed
visual row's actual and diff PNGs through `make ui-capture`. Their bundle component
bytes and SHA-256 were verified before native-resolution inspection. The default
shell and selector labels were visible; the diff localized the 67 changed pixels
to the selected Timeline underline. Canonical import of capture
`visual.capture.7314c63e882d78cd3cf8` was rejected with
`artifact/artifact_error/unsafe_artifact`; explicit image imports supplied image-only
evidence, with no DOM or Axe channel. Exact-session stop confirmed cleanup complete,
the foreground exited and caller-owned request scratch was removed. Transient image
links are expired and are not published here. This observation drove preservation
of the prior underline geometry, not a golden refresh.

Seeded interaction review remains unavailable even though the product browser harness
started successfully. No current manual selector focus/contrast inspection or private
Axe pass is claimed. No golden was refreshed. Retained-run maintenance was skipped
because `RESULTS_DIR` was unset; no qualifying full warm check run was supplied.
A full repository suite, shared-source owner tests and unrelated UX repairs were
skipped because this diff is local presentation plus coordinated test consumers.

| Digest acceptance | Assessment and bounded evidence |
| --- | --- |
| A001–A004 | PASS: authority map, local mechanism/rubric, clean baseline/current manifests, existing focus token and no second registry. |
| A007 | PASS for selector entry: editor creation, viewer fallback and no selector-created record; source/capability authoring policies unchanged. |
| A008–A009 | PASS for this seam: supported compact replacement, tested zoom, Tab exits and external-focus guard; thresholds and independent panel scrolling unchanged. |
| A011, A014 | PASS: raw authoring retention, selected-tab fallbacks and delayed-entry supersession; no new operation or authoring lifetime. |
| A016 | PASS for entry: editable and viewer populated/empty destinations; data/permission producers unchanged. |
| A019 | BLOCKED for private rendered focus review; production keyboard/semantic/geometry product checks pass, but seeded rendered inspection is unavailable. |
| A020 | PASS for selector variants and exercised width/zoom; no whole-page/text-spacing claim. |
| A022 | PASS, bounded: production default/narrow/compact fixture passed against unchanged goldens; actual/diff image review localized the earlier underline change and led to its retirement. Image-only review supplies no DOM/Axe or manual keyboard focus claim. |
| A023–A026 | PASS: canonical IDs and migrated consumers; no Markdown dependency, generated projections via Make, no API/data/saved-view/preference migration. |
| A027 | BLOCKED for overall completion because A019 remains blocked; this handoff records the available evidence and prerequisite. |
| A005–A006, A010, A012–A013, A015, A017–A018, A021 | N/A: theme/density, inspector dispatch, transaction/recovery implementation, conflict presentation, authorization lifetimes, Evidence lifecycle and virtualization are not changed. Existing owners remain intact; compatibility test results do not certify those subsystems. |

Next action: resolve the seeded-review startup prerequisite, start a newly sealed
session, inspect selector focus and relationships through the UI-review skill,
then stop that exact session with confirmed cleanup and update blocked acceptance.
Rollback reverts the frontend implementation, coordinated tests, authored routing,
generated projections and this handoff together. No data rollback is required.
