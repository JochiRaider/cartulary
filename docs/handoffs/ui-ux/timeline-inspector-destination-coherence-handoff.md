# Timeline inspector destination coherence

## Baseline and authority

Implementation began on clean `main` at `352b763d085c0c9b5f689dd3174ca4b9633b652b`. The adopted owners are Core 01 §6 (`inspector_config_v1`, including REQ-01-615/616) and Core 03 REQ-03-220, REQ-03-299, and REQ-03-100. Design §§8.4–8.5, 12.7, and 14 govern the presentation and accessibility behavior; `docs/domain.md` supplies vocabulary. The UI/UX digest, NLSpec research essay, and prior handoffs were navigation and review material. The typed panel descriptors and subject identity remain projections of the owners; test catalog rows route evidence independently.

Baseline browser assertions were added before runtime changes. They failed in production composition after focus and scroll settled:

- Space on a committed Timeline row with one linked text Evidence item, then preview Escape: the Evidence list had focus, while Details retained `aria-current`. `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser.timeline_linked_text_review_3f18dc6414`, root `.cartulary/test-results/20260928T152516Z-p12387`.
- Details → Manage Tags: Relationships had focus, while Details retained `aria-current`. `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.timeline_manage_tags_makes_relationships_current_3cb930689d`, root `.cartulary/test-results/20260928T152406Z-p79521`.
- Direct Relationships navigation was already coherent. The defect was the split between Timeline's semantic focus registry and shared inspector section selection.

## Ownership and change

`workbookInspectorNavigation.ts` now exposes an explicit destination operation over an admitted panel ID and optional registered semantic element. It returns applied, pending, or unavailable. Shared presentation selects the section, reveals inside the inspector body, opens presentation disclosures needed for the target, and focuses it. It suppresses the matching programmatic scroll observation, then resumes passive section selection on user scroll. Ordinary buttons and the narrow chooser continue through the same presentation coordination. No second selected-section store was introduced.

The Timeline element registry still owns registration, canonical subject/row-version checks, first-open pending intent, collection/mention identity, and return anchors. It now forwards valid explicit destinations to shared presentation. Space entry and Evidence preview Escape/Close return, Alt+H, Details Manage Hosts/Identities/Tags/attached Evidence, and the traced collection/mention jumps use this path. Their former direct positioning/focus calls were removed; direct focus remains only for local return anchors. Alt+H focuses `Open history` while History is unrequested and does not start its read.

Pending focus is retired on superseding intent, deliberate section choice or user body scroll, immediate closure, subject/version/lifecycle change, authority change, and unavailable admission. A matching first-open request may complete when its semantic target registers. Concealment removes both the panel and destination. Existing authoring and source-owned Evidence eligibility/freshness/recovery, mutation state, and uncertain-write identity remain untouched. Compatibility requires no route, schema, persistence, dependency, or data migration. Rollback is one source/test/routing diff reverting the shared operation and migrated callers.

## Evidence and acceptance

Focused routed unit tests cover shared explicit/passive scroll ordering, late registration, supersession, invalidation, concealment/unavailable admission, and first-open empty Evidence focus. Routed browser checks cover the two reproduced failures, linked preview Escape and Close, blocked/multiple Evidence outcomes, Alt+H with zero History reads, generic Hosts authoring retention with zero navigation writes, and 320px/200%-text accessibility fixtures. Wide and narrow seeded review captures showed Relationships current and revealed, with the record header and Close available; the narrow chooser exposed all five admitted destinations. The seeded review is visual and observational evidence, not product-test, accessibility-conformance, or golden evidence. Its wide-page axe scan reported four advisory findings (aria-required-children, color-contrast, empty-table-header, and page-has-heading-one); the narrow chooser capture reported none. Those adjacent findings were not changed in this bounded slice.

Digest acceptance: A001 (owner map and distinct routing), A002 (one shared decision and retired positioning), A003 (clean baseline/current source and generated-root review), A009 (body-only overflow), A010 (the affected admitted-descriptor and concealment subset), and A011 (semantic focus and retained authoring) apply and have the evidence above. A004 (no new token), A005 (no theme change), A006 (no density change), A007 (no creation change), and A008 (no breakpoint or responsive-threshold implementation change) are N/A; narrow rendering was still checked for this destination behavior. Design D-AC-082 is the primary navigation acceptance row; D-AC-083/084/085 cover retained authoring, Evidence, and unrequested History in the selected regressions. D-AC-086 has seeded rendered observations only; no visual-golden or publication claim is made.

## Verification record

Repository commands were run from the root through public Make targets. Initial discovery used `make help`, `make help-all`, the four requested `make task-guide ROLE=module-author OWNER=...` routes, and targeted `make explain-test-owner`/`make explain-target` inspection. Authored catalog changes were projected with `make generate`; no generated output was hand-edited.

| Command / row | Outcome and run root |
| --- | --- |
| `make generate` | Pass, `.cartulary/test-results/20260928T153829Z-p54274` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_persistent_context,web.workbook.regression.timeline_first_open_evidence_focus_78b3aba0fd` | Pass, `.cartulary/test-results/20260928T154620Z-p7766` |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.collection_inspection_focuses_only_live_semantic_40e7d15084` | Pass, `.cartulary/test-results/20260928T154423Z-p93526` |
| `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser.timeline_linked_text_review_3f18dc6414,module.evidence.browser.timeline_linked_blocked_preview_87b7079b89` | Pass, `.cartulary/test-results/20260928T154032Z-p92111` |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.timeline_manage_tags_makes_relationships_current_3cb930689d` | Pass, `.cartulary/test-results/20260928T154140Z-p25287` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.required_keyboard_shortcuts_operate_on_the_live_3ff4a5c6a9` | Pass, `.cartulary/test-results/20260928T153844Z-p57315` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_subject_retention,module.workbook.accessibility.inspector_edit_recovery` | Pass, `.cartulary/test-results/20260928T154249Z-p58541` |
| `make format` | Pass, `.cartulary/test-results/20260928T154450Z-p94558` and `.cartulary/test-results/20260928T154611Z-p3328`; diff reviewed |
| `make agent-finalize` | Pass, `.cartulary/test-results/20260928T154507Z-p98989`; `RESULTS_DIR` unset because no eligible successful full warm check was supplied, so retained-run maintenance was skipped |
| `make frontend-typecheck` | Pass, `.cartulary/test-results/20260928T154423Z-p93604` |
| `make frontend-import-boundary-check` | Pass, `.cartulary/test-results/20260928T154232Z-p57776` |
| `make ui-review UI_MODE=seeded`, `make ui-capture`, `make ui-review-stop` | Seeded isolated review, root `.cartulary/test-results/20260928T154625Z-p8554`; captures `bundle-9` (1440px), `bundle-11` (320px), `bundle-13` (320px chooser); images and observations inspected before stop; terminal receipt and status confirmed closed |

Failures during implementation were resolved: the first Manage Tags test targeted an inner nonfocusable wrapper (`.cartulary/test-results/20260928T152241Z-p46812`); the initial new catalog row needed generation (`.cartulary/test-results/20260928T152145Z-p42618` and `.cartulary/test-results/20260928T152208Z-p43202`); an unsorted authored title list failed `make generate` (`.cartulary/test-results/20260928T153647Z-p51026`); the new shared test had incorrect mocked geometry (`.cartulary/test-results/20260928T153844Z-p57283` and `.cartulary/test-results/20260928T153930Z-p90281`); and an old Space assertion expected the whole section instead of its empty-state semantic target (`.cartulary/test-results/20260928T153224Z-p14541`). Each was test/routing setup or an assertion corrected against the observed UI; the two baseline product failures are listed above.

Final-source verification after formatting and the last focus-lifetime adjustment:

| Command / row | Outcome and run root |
| --- | --- |
| `make agent-finalize` | Pass, `.cartulary/test-results/20260928T155150Z-p16285`; `RESULTS_DIR` unset, so retained-run maintenance was skipped |
| `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser.timeline_linked_text_review_3f18dc6414` | Pass, `.cartulary/test-results/20260928T155245Z-p20307` |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.timeline_manage_tags_makes_relationships_current_3cb930689d` | Pass, `.cartulary/test-results/20260928T155355Z-p61016` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.required_keyboard_shortcuts_operate_on_the_live_3ff4a5c6a9` | Pass, `.cartulary/test-results/20260928T155453Z-p93520` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_subject_retention,module.workbook.accessibility.inspector_edit_recovery` | Pass, `.cartulary/test-results/20260928T155559Z-p26596` |
| `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser.timeline_linked_blocked_preview_87b7079b89` | Pass, `.cartulary/test-results/20260928T155707Z-p63145` |
| `make generate-drift` | Pass, `.cartulary/test-results/20260928T155251Z-p26785` |
| `make generated-artifact-policy-check` | Pass, `.cartulary/test-results/20260928T155251Z-p26775` |
| `make json-shape-check` | Pass, `.cartulary/test-results/20260928T155251Z-p26855` |
| `make frontend-typecheck` | Pass, `.cartulary/test-results/20260928T155327Z-p57455` |
| `make frontend-import-boundary-check` | Pass, `.cartulary/test-results/20260928T155327Z-p57453` |
| `make lint-biome` | Pass, `.cartulary/test-results/20260928T155327Z-p57501` |
| `make lint-markdown` | Pass, `.cartulary/test-results/20260928T155251Z-p27325` and `.cartulary/test-results/20260928T155618Z-p55211` |

No commit, push, deploy, golden refresh, or unrelated UX change was made.
