# Saved Timeline tag removal handoff

## Scope and authority

This slice adds one saved-tag removal action to the existing Timeline
Relationships collection. The source owner is `web.workbook`; independent
verification routes through `module.timeline`, `web.workbook`, and
`module.links`. Core 01 §7.4.1, especially REQ-01-321/322, owns the
`collection_actions_v1` `remove_tag` request and PATCH transaction. Core 03
§2.3A, §§3–4, and REQ-03-219/298/299/100 own authoring, source-write
coordination, conflict, recovery, and lifetime. Core 04 owns current incident
authorization. Design §§4, 8, 9.3, 10, 12, and 14 supply presentation and
keyboard direction. The digest and this handoff are navigation and review
material. No owner contradiction was found.

The execution recheck found clean `main` at
`cdd2fefde597131911f2f20bd326797dc42925ff`, with only the root
`AGENTS.md` applicable. No saved-tag removal frontend path existed. The
generated-artifact policy and current source/import manifests were inspected
before changes. No vendor code, dependency, API, schema, migration, or
specification changed.

## Behavior and retirement

- Writable Relationships tags now show a named Remove button beside each
  complete saved item. Read-only users retain complete inspection without the
  action. Grid chips and `+N` remain inspection entries; the Details Manage
  Tags entry reaches the same collection. The prior authorized inspector
  inspection-only tag branch is retired.
- Timeline builds one `remove_tag` change from the returned `item_ref` and
  captures its source record, saved baseline, sheet, and readable operation
  label. The retained explicit PATCH owner reserves the record synchronously,
  coordinates earlier writes, reads current values, enforces the collection
  baseline and row version, and retains one immutable request for uncertain
  replay. Timeline also checks that the exact item still belongs to the
  original record. No label-derived identity or global tag mutation is used.
- The accepted row is projected through the existing Timeline mutation path.
  The existing recovery component sits beside Tags in Relationships and shows
  pending, rejected, conflict, uncertain, saved, and saved-needs-refresh
  states. Its retry reuses the original transaction and request; refresh after
  acknowledgement performs reads. Attention and notices carry the captured
  tag label to the Relationships collection.
- The initiating button stays focusable while work is pending. On accepted
  removal, focus moves to the next surviving local Remove action, then the
  previous one, then the collection input/group. A newer focus, keyboard,
  pointer, or scroll intent cancels restoration. Inspector-only reveal does
  not scroll the grid. Grid and inspector raw drafts remain independent.
- No second operation store, retry engine, selected-tag authority, duplicate
  recovery list, or generalized collection-management API was introduced.
  Existing token addition and bulk-tag ownership remain in place.

## Changed paths

Authored source and tests:

- `apps/web/src/workbook/timeline/components/TimelineCollectionCell.tsx`
- `apps/web/src/workbook/timeline/components/TimelineCollectionCell.test.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookInspector.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookRenderers.tsx`
- `apps/web/src/workbook/timeline/components/useTimelineCollectionRenderer.tsx`
- `apps/web/src/workbook/timeline/models/timelineMutationIntents.ts`
- `apps/web/src/workbook/timeline/models/timelineModelBoundaries.test.ts`
- `apps/web/src/workbook/timeline/models/timelineRowModel.ts`
- `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- `apps/web/src/workbook/runtime/WorkbookExplicitPatchOwner.ts`
- `apps/web/src/workbook/inspector/WorkbookExplicitPatchRecovery.tsx`
- `apps/web/src/workbook/inspector/workbookInspectorErrorModel.ts`
- `apps/web/src/workbook/inspector/workbookInspectorOrdinaryAttention.ts`
- `apps/web/src/workbook/inspector/presentation/workbookInspectorPresentationModel.ts`
- `apps/web/e2e/timeline-collection-input.spec.ts`
- `tools/test_families/module.timeline.json`
- `docs/handoffs/ui-ux/workbook-timeline-saved-tag-removal-handoff.md`

Generated outputs: `tools/browser_e2e_batch_manifest.json` and
`tools/execution_topology_render_index.json`.

## Characterization and verification

The first seeded UI-review session used isolated synthetic tags. Details
Manage Tags reached the complete collection and found only read-only notes.
The UI-review grid scroll target was unavailable for chip and `+N` navigation;
the later routed Chromium scenarios exercised those entries. The private
review session was stopped. A focused red
unit run at `.cartulary/test-results/20260927T195829Z-p7527` failed on the
missing Remove button before production edits.

| Command | Result and run root |
| --- | --- |
| `make help`, `make help-all`, three owner `make task-guide` routes | Current public targets and owner routes inspected |
| `make format` | Passed, `.cartulary/test-results/20260927T212438Z-p10720` |
| `make generate` | Passed, `.cartulary/test-results/20260927T212448Z-p15335` |
| `make agent-finalize` | Passed, `.cartulary/test-results/20260927T212501Z-p18204`; `RESULTS_DIR` unset because no qualifying full warm run was supplied |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.collection_inspection_preserves_draft_text_and_s_f5c8c4c211` | Passed, `.cartulary/test-results/20260927T212425Z-p10084` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell__inspector_renders_timeline_collec_062be27af9,web.workbook.regression.workbooktimelinemodel_builds_scalar_collection_e_13cb3751a8` | Passed, `.cartulary/test-results/20260927T211758Z-p3952` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.explicit_task_patch_recovery` | Passed, `.cartulary/test-results/20260927T211933Z-p5558` |
| `make service-backed-test-slice OWNER=module.links ROWS=module.links.integration.link_and_tag_mutations_atomically_update_project_03f8283906` | Passed, `.cartulary/test-results/20260927T204415Z-p16307` |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.saved_timeline_tag_removal_preserves_exact_ident_42963ccf0e` | Passed, `.cartulary/test-results/20260927T212730Z-p33091` |
| `make frontend-typecheck`, `make lint-biome`, `make frontend-import-boundary-check` | Passed, `.cartulary/test-results/20260927T212602Z-p22978`, `...-p23071`, `...-p23088` |
| `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check` | Passed, `.cartulary/test-results/20260927T212602Z-p22750`, `...-p22840`, `...-p22933` |
| `make lint-markdown` | Passed, `.cartulary/test-results/20260927T212602Z-p23511` |

The browser run covers editor removal, another record with the same label,
viewer inspection through `+N`, preserved unsent input, keyboard focus,
exact uncertain replay, and a failed post-acknowledgement refresh followed
by read-only recovery. It uses synthetic records at 1440×900 in comfortable
density and 1024×720 in compact density. Screenshots named
`saved-tag-removal-1440`, `saved-tag-removal-1024`, and
`saved-tag-refresh-recovery` are attached to the browser report. They were
manually inspected as advisory UI evidence, separate from the routed product
test results. The long label wraps and leaves its Remove control visible.

Controlled unit promises cover duplicate pointer/keyboard activation, held
response, exact request and record, first/middle/last/empty focus, independent
drafts, stale collection, conflict registration, rejection, uncertain replay,
refresh debt, viewer authority, inspector navigation, and newer
keyboard/scroll intent. Existing
explicit PATCH owner tests cover detachment, session concealment and account
retirement. The Links contract row confirms persisted projection/history
semantics without backend edits.

An intermediate browser run failed because its synthetic long tag exceeded
the adopted 64-rune input limit; shortening the fixture resolved that
test-data error. Another intermediate run selected a hidden chip as though
it were visible; selecting the actual visible chip resolved that test error.
No current verification target is blocked. The full `make check` and release
suite were not run because these owner slices, service-backed scenarios, and
boundary checks cover the changed consumer. No visual golden was refreshed.

## Digest acceptance assessment

Applicable rows A001, A002, A003, A004, A010, A011, A012, A013, A014,
A015, A017, A019, A020, A023, A024, A025, A026, and A027 are PASS for this
slice based on the owner map, exact request/race tests, authorized browser
paths, inspected density/viewport screenshots, and generated/boundary checks
above. A005–A009, A016, A018, A021, and A022 concern unchanged theme,
global density/chrome, query/Evidence, virtualization, or golden registries;
they are N/A to this focused correction. The viewport/density and overflow
aspects of A020 were still exercised through the production renderer.

Rollback is the focused source, test, authored routing, and regenerated
topology change. Existing stored tags and history remain valid. No data
migration or compatibility rewrite is required.
