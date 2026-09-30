# Workbook status-strip accessibility correction

Date: 2026-09-30. Implementation and focused product verification are complete;
closing documentation checks are recorded below.

## Baseline and authority

The starting checkout was clean `main` at
`ff795007602fd3e6651b5ff979a1bac8b99b5efa`, three commits ahead of `origin/main`.
HEAD had not advanced from the reviewed baseline. The preceding Timeline
bulk-selection, mention readability, focus-reveal, Columns, and surface-navigation
work remains intact. No commit, push, deployment, reset, or golden refresh was made.

| Owner or guidance | Application to this slice |
| --- | --- |
| [Core 00 §§1–2](../../spec/00_document_set_status_and_precedence.md) | Adopted owners govern; projections, code, tests, digest material, research and historical handoffs do not become requirements. |
| [Core 03 §4.2, REQ-03-089](../../spec/03_workbook_interaction_collaboration_and_workflows.md) | Preserve exactly Saved, Syncing and Conflict, authoritative mutation settlement, acknowledgement independent of refresh, and conflict recovery. Its prohibition on identifiers in ordinary visible conflict summaries is narrower than the hidden diagnostic finding. |
| [Design §§6.2, 7.1, 7.5, 10.1, 14.1–14.2](../../design.md) | Ordinary product language, status-strip allocation, full accessible secondary text, recovery access and single-owner announcements. Below-minimum secondary omission remains permitted. |
| [Domain §§1, 6, 8.1, 9](../../domain.md) | Vocabulary and semantic surface/record/field identity, without introducing behavior or treating implementation diagnostics as product concepts. |
| [Digest START_HERE](../../cartulary-ui-ux-refactor-digest/cartulary/START_HERE.md), maps, rules and acceptance | Advisory scope selection and evidence assessment; embedded prompts and research were treated as source material. |

The applicable repository AGENTS.md, digest README, LOCAL_AGENT_PROMPT,
REPO_MAP, OWNER_MAP, QUERY_RECIPES, rules and acceptance material were inspected.
The continuity, runtime, components and parent source guides were followed.
No adopted-owner contradiction or normative specification change was needed.

Source ownership is **web.workbook**, verified against the authored source and
import manifests. Browser verification is independently owned by
**module.workbook**. The generated-artifact policy and active catalog rows were
inspected. Direct grid-vendor integration remains in Grid Adapter.

## Boundary, edits and retained behavior

The confirmed defect was internal continuity text in the Status strip accessible
tree. It was not a changed save-state mapping, security incident, demonstrated
screen-reader interruption, or whole-page accessibility-conformance finding.

The shared `WorkbookContinuityAnchorStatus` span now has `aria-hidden="true"`.
It remains nonfocusable, visually hidden and queryable by
`workbookFocusAnchorTestId()`, with the exact null sentinel or serialized semantic
anchor. Only its accidental accessibility-tree exposure is retired.

The common decision is whether diagnostic continuity metadata belongs in
accessible product status. Keeping that decision at the existing shared boundary
corrects all its consumers and lets future consumers use the same status path
without another diagnostic implementation or state owner. Leaving the defect
would continue exposing internal identifiers as status text; actual assistive
technology disruption remains unmeasured.

Changed paths:

- `apps/web/src/workbook/continuity/useWorkbookGridContinuity.tsx`: exclusion and
  diagnostic-purpose comment; the only production change.
- `apps/web/src/workbook/workbookSaveStatus.test.tsx`: three focused tests for
  inaccessible diagnostics with exact observability, responsive status/recovery
  and presence semantics, and silent anchor-only rerenders.
- `apps/web/e2e/workbook.a11y.spec.ts`: existing shell scenario characterizes both
  anchors, keyboard movement and Hosts navigation; the existing mutation-lifecycle
  fixture checks Syncing/Conflict content, complete secondary text and keyboard
  recovery entry across the supported responsive modes.
- `tools/test_families/web.workbook.json`: exact new unit titles added to the
  existing save-status row; browser scenario titles remain unchanged.
- `tools/execution_topology_render_index.json`: generator-updated catalog input
  hash and aggregate digest, with no manual generated-file edits.
- Workbook and continuity READMEs: bounded test/diagnostic descriptions.
- This handoff: authority, evidence, dispositions and rollback.

Inspected retained paths include `WorkbookStatusStrip`, `WorkbookObservedStatusStrip`,
`WorkbookSaveAnnouncements`, `workbookContinuityPort`,
`workbookMutationStatusProjector`, `WorkbookMutationRuntime`, `workbookStyles`,
the selector owner, TimelineWorkbookView, EntityWorkbookSurface,
AssessmentWorkbookSurface, GenericWorkbookSurface and WorkbookShell. Save
projection, intentional visually hidden secondary text, action descriptions,
presence and the shell-lifetime announcement host are unchanged.

There is no new state, subscription, operation owner or lifetime. Queue admission,
captured requests, uncertain replay, acknowledged writes, refresh recovery,
authority concealment/retirement, raw drafts, selection and viewport mechanics
are untouched. No parallel diagnostic, narration or announcement path was added.

## Characterization and focused verification

All commands ran from the repository root through public Make targets. Discovery
used `make help`, `make help-all`, both module-author task guides, both owner
explanations and target summaries. Exact run roots below are relative to that root.

The pre-correction production-browser run failed only the four newly added
diagnostic-negative assertions: Timeline null, UTC anchor, keyboard-moved Local
Time anchor, and Hosts null. The corresponding trees retained Saved. Other shell,
focus and no-save-announcement assertions passed.

```bash
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159 CARTULARY_HARNESS_CACHE_MODE=off

make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbook_save_status_preserves_global_blockers_a_8d590e0883,web.workbook.regression.workbookcontinuityport_suite_b9660fc217,web.workbook.regression.workbookshell__sentinel_grid_anchor_shell_suppor_09dddf7118 CARTULARY_HARNESS_CACHE_MODE=off

make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_save_state_presentation_derives_one_prima_cad2a59be9 CARTULARY_HARNESS_CACHE_MODE=off

make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159,module.workbook.accessibility.verify_grid_navigation_edit_entry_exit_paste_fee_3de2b0afac CARTULARY_HARNESS_CACHE_MODE=off
```

| Command/result | Exact run root |
| --- | --- |
| Shell browser characterization: expected failure before production correction | `.cartulary/test-results/20260930T180504Z-p96622` |
| `make format`: pass; authored diff inspected | `.cartulary/test-results/20260930T181123Z-p31344` |
| Initial `make agent-finalize`: stopped before regeneration on stale routing inputs | `.cartulary/test-results/20260930T181213Z-p36178` |
| Diagnostic `make json-shape-check`: stale `web.workbook.json` topology input; directed `make generate` | `.cartulary/test-results/20260930T181310Z-p36873` |
| `make generate`: pass; regenerated only the topology render index | `.cartulary/test-results/20260930T181354Z-p37527` |
| `make agent-finalize`: pass, including shape/catalog/drift validation | `.cartulary/test-results/20260930T181423Z-p40643` |
| Initial web slice: continuity and sentinel passed; new responsive unit assertion needed to distinguish two DOM copies of short narrow-mode text | `.cartulary/test-results/20260930T181503Z-p44836` |
| Module shell save-state slice: pass | `.cartulary/test-results/20260930T181503Z-p44839` |
| Both browser rows: pass, 2 expected scenarios, 0 unexpected, skipped or flaky | `.cartulary/test-results/20260930T181503Z-p44858` |
| `make format` after the unit-query correction: pass; diff inspected | `.cartulary/test-results/20260930T181627Z-p80498` |
| Final web slice: pass, all 3 rows / 17 tests | `.cartulary/test-results/20260930T181713Z-p85454` |
| Final `make frontend-typecheck`: pass | `.cartulary/test-results/20260930T183335Z-p19824` |
| Final `make lint-biome`: pass | `.cartulary/test-results/20260930T183335Z-p19849` |
| `make lint-markdown`: pass, including the handoff and source guides | `.cartulary/test-results/20260930T183920Z-p22084` |
| `make generated-artifact-policy-check`: pass | `.cartulary/test-results/20260930T181530Z-p76407` |

The unit-query correction requires exactly one accessible full-text copy; it does
not change product code. Both Enter and Space must independently activate status
actions. Tests also retain same-field counts, Saved-with-refresh-debt detail,
presence names, transition priorities and shell-remount announcement deduplication.

Retained-run maintenance was skipped because `RESULTS_DIR` was unset; no eligible
successful full warm check was supplied. No goldens, broad visual/backend/release
suite, or unrestricted browser audit was needed. No source/import boundary or
selector API changed, so additional boundary suites were not selected.
The final diff and checkout status were inspected; `git diff --check` passed.
Only the eight implementation, test, routing, generated-index and documentation
paths listed above are changed.

## Seeded rendered review and limitations

Session command: `make ui-review UI_MODE=seeded REVIEW_PROFILE=default`, using the
isolated synthetic Review editor and populated review incident. Safe structural
run root: `.cartulary/test-results/20260930T181714Z-p85759`.

| Live-consumed evidence | Assessment |
| --- | --- |
| `bundle-6`, Timeline 1440×900, no active anchor | Visible strip shows Saved; accessibility tree contains Saved without cleared; diagnostic selector still contains cleared. |
| `bundle-9`, Timeline 1024×720, UTC cell activated | Editor/cell focus treatment is visible; Saved remains accessible; UTC diagnostic is observable in DOM but absent from accessible status. |
| `bundle-12`, same viewport after Escape and ArrowRight | Active outline moves to Local Time; the diagnostic field changes semantically; accessible status still contains only Saved. |
| `bundle-15`, Hosts 1440×900 after keyboard surface activation | Hosts is active with its seeded row; visible/accessibility status is Saved and excludes cleared. |
| Return to Timeline | Selected Timeline and Saved were observed again without an accessible diagnostic. |

All four original screenshots and observation components were resolved through
their returned manifests, verified for byte length and SHA-256, and consumed
before stopping. The visible strip matches the legitimate save-label presentation
described in baseline evidence; no pixel comparison or golden claim was made.
Automated browser focus assertions independently passed.

Axe completed in all four captures, reporting `empty-table-header` (minor) and
`page-has-heading-one` (moderate), plus one incomplete observation per capture.
These advisory observations are outside the status-strip correction and were not
treated as product-test failures or whole-page conformance certification. The
terminal receipt also counts two console errors; their origin was not assessed
in this bounded review. No actual assistive-technology testing or measurement of
screen-reader interruption was performed.

`make ui-review-stop` succeeded. Its verified `ui-review/terminal.json` reports
`status=ok`, `state=closed`, `cleanup=complete`; the foreground process exited 0.
Caller-owned request files were removed. Private screenshots, observations and
links expired with cleanup and are not retained or linked from this handoff.

## Digest acceptance assessment

The dispositions below apply to this presentation seam, not to unrelated product
workflows or whole-page conformance. No applicable row remains blocked.

| ID | Disposition | Evidence or scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | Exact owner/design mapping above; source and verification owners kept separate. |
| A002 Scope | PASS | One existing shared diagnostic boundary; accidental exposure retired, useful semantic marker retained; no new abstraction or owner. |
| A003 Repository state | PASS | Clean baseline, manifests, local guides, policy, consumer paths and active routing revalidated. |
| A004 Tokens | PASS | No token, styling literal or registry introduced. |
| A005 Theme | N/A | Theme behavior is unchanged; no new theme fixture required. |
| A006 Density | N/A | No geometry, density selection or editor-sizing change. |
| A007 Creation | N/A | No creation capability, route, payload or retained-authoring change. |
| A008 Responsive | PASS | Full secondary text checked in base/narrow/compact; below-minimum omission retained; production browser recovery checked at 1440×900, 1024×720 and 768×640. Responsive classification/inspector geometry remain untouched. |
| A009 Overflow | PASS | Visible bottom strip remains reachable in consumed captures; existing shell/recovery geometry assertions pass. |
| A010 Inspector | N/A | No inspector dispatch, configuration or owner-lifetime change. |
| A011 Continuity | PASS | Exact diagnostic identities remain observable; port capture/restore/clear/disposal and shell keyboard-anchor rows pass; production activation/movement/navigation passes. No draft or viewport implementation changed. |
| A012 Transactions | N/A | Request capture, randomness, admission and replay are unchanged. |
| A013 Acknowledgement and recovery | PASS | Existing acknowledged-write/read-recovery and pending-work tests pass; Saved refresh detail and recovery entry remain accessible. Transport/lifetime algorithms are unchanged. |
| A014 Editing | N/A | No editing algorithm or raw-draft lifetime change; existing browser edit/Escape checks provide regression support. |
| A015 Conflict | PASS | Conflict labels/count descriptions and recovery actions remain accessible; blocked-edit production fixture passes. Cell-level resolver implementation is unchanged. |
| A016 Query data and interaction | N/A | No query producer, authority inference or interaction-state composition change. |
| A017 Refresh and authorization scope | N/A | No refresh, concealment, retirement or account/incident-lifetime change. Retained acknowledged-read test passes as supporting evidence. |
| A018 Evidence | N/A | Evidence lifecycle, access and overlays are untouched. |
| A019 Accessibility | PASS | Scoped production trees exclude both diagnostics; labels, descriptions, keyboard actions, visible focus and announcement ownership/deduplication pass. Advisory Axe and absent AT testing are qualified above. |
| A020 Components | PASS | Save, conflict, refresh-detail and presence representations checked across responsive modes; recovery geometry/names/focus pass. Unchanged density/zoom/text-spacing algorithms are not redesigned. |
| A021 Virtualization | N/A | No virtualization, row identity or result-size implementation change; sampled production-grid continuity is assessed under A011. No performance claim made. |
| A022 Visual fixtures | PASS | Existing seeded harness captured the production renderers at explicit viewports; original images were integrity-checked and manually inspected. No canonical golden/publication claim. |
| A023 Selectors | PASS | Typed schema IDs, fixture-returned record IDs, field keys, existing selectors and semantic roles/names; no whole-page identifier ban or broad snapshot. |
| A024 Test authority | PASS | Added tests use source fixtures/contracts and renderer behavior, with no Markdown reads/stats/hashes/dependency. |
| A025 Generated artifacts | PASS | Authored routing changed first; Make generation and finalization/policy checks pass; generated diff is only the catalog input digest projection. |
| A026 Compatibility | PASS | No schema/API/data migration, adapter or flag; retained diagnostics have active semantic test consumers; rollback is bounded below. |
| A027 Handoff | PASS | Owners, edits, commands/roots, actual failures and recovery, consumed review, cleanup, limitations, dispositions and rollback are recorded. |

Relevant advisory rules were applied through existing owners: R001/R002 keyboard
and focus, R003/R008 meaningful state and recovery, R013 semantic controls,
R014 continuity and R015 bounded completion. R033/R034 reject using generic advice
as behavior authority or incidental selectors. No upstream recommendation
required a new dependency, design system or broader audit.

## Compatibility and rollback

No data migration, public API/type change, feature flag, persistence or compatibility
adapter is needed. The existing semantic continuity and announcement systems are
retained; only the diagnostic span's accessible exposure is removed.

Rollback reverts the presentation correction, its associated tests and authored
routing, regenerates the matching topology index, and removes this handoff/source
guide updates together. It must preserve preceding UI commits and user data.
No further implementation work is planned for this seam.
