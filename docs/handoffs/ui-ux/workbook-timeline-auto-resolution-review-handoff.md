# Timeline auto-resolution Review handoff

## Baseline and authority

Baseline: clean `main` at `cc34c69311af04710b5880be1e61f4cc58fa55a3`. The changed paths are the Timeline Review owner, its immediate Inspector and notice consumers, focused tests, authored test routing, and generated routing projections. No unrelated working-tree change was present at entry.

Core 03 §12.5, especially REQ-03-210, owns the immediate disclosure, direct Review, and retention until an actual correction or authoritative source change. REQ-03-298/299/100 keep raw authoring, authority, and retained mutations with their existing owners. Design §§9.4, 12.5, 12.7, and 14 guide local status, Inspector navigation, semantic focus, and accessible feedback. `docs/domain.md` supplies vocabulary. The UI/UX digest, historical handoffs, and `docs/research/nlspec-spec.md` were navigation and review aids, not additional product authority. Executable tests and generated inputs do not depend on Markdown.

The baseline implementation created an unretained AbortController in `prepareDisclosureReview`, checked only mount, authority generation, and presentation key after the read, then selected the mention in its caller. The notice leaf owned a separate promise rejection message and had no pending admission state. These are confirmed code defects: a later read could reach Inspector after newer editing, repeated activation could read twice, and cancellation could appear as a source failure. During browser regression development, an off-query source read was also observed not to start when the new implementation reused a retained committed row; the correction preserves the existing read requirement for a source outside visible query rows. Initial browser setup attempts failed on a nonfilterable field and then an Inspector overlay in the narrow viewport; neither is a product baseline pass.

The production browser path creates an authorized auto-resolution through the UI, filters its source out of the query window using a tag, holds Review's paged Timeline query, and resumes native grid authoring. Its assertions use the actual input element, value, caret, selection, focus, Inspector target, disclosure, and request count. Existing authority, accepted-receipt, Undo, and Find/Timeline cancellation owners were inspected before the change.

## Selection rubric and decision

| Decision | Review result |
| --- | --- |
| Observed gap | Review navigation had no retained request identity, admission, or cancellation. Off-query Inspector focus looked only in visible rows. The affected action is direct Review of a retained auto-resolution. |
| Remediation and scope | Behavior correction in the Timeline mention-action hook, Inspector selection, and notice presentation. Focused tests, authored routing, projections, and this handoff follow. No specification or schema edit is needed. |
| Owner boundary | `useTimelineMentionActions` owns one read/navigation request and local feedback. The mention operation owner still owns disclosure, authority, Undo, accepted receipts, replay, and source-read freshness. The committed-row ledger owns accepted off-query rows; Inspector selection owns semantic navigation and focus. Presentation remains passive. Verification is independently routed by `web.workbook`, `module.timeline`, and the existing `module.entities` row. |
| Rationale and extension | One request owner fences every publication with semantic disclosure identity, explicit request ID, authority, surface, and Inspector context. A future local navigation action can use the same ownership seam without a second draft or notification store; no speculative surface was added. |
| Capability and retirement | Retain direct Review, Undo, disclosure, source paging, semantic focus, native editing, and owner mutation recovery. Remove `prepareDisclosureReview`, the caller's post-await selection, and the notice leaf's promise/catch error state together. |
| Compatibility and migration | No endpoint, public route, schema, storage, durable state, dependency, or data migration. The internal notice callback changes from an async result to synchronous admission; all production callers and test fixtures migrate together. Rollback is the Timeline hook, Inspector selection, notice, presentation, and routing changes as one unit. |
| Risk and exit | Without fencing, a late read can move focus or Inspector over a newer draft, and duplicate reads or false errors obscure recovery. Exit requires one logical read for repeated activation; only the latest owned Review focuses its original mention; stale success/failure is silent; genuine failure retries; cancellation stops paging; disclosure, query, and mutation behavior remain correct. |

Material digest advice was classified through existing `rules.tsv`: R001/R002/R006/R008/R013/R014 `ADOPT` for keyboard continuity, semantic focus, local pending/failure and retry, semantic controls, and stable navigation; R010 `ADOPT` for compact feedback without a layout jump; R012 `ADAPT` so React state remains a projection of the Timeline request owner; R015 `ADOPT` for evidence-based exit. No upstream advice creates a second authority or a generic notification system.

## Final interaction policy

- Admission captures disclosure source, field, mention, item, version, and target plus a monotonic request ID and AbortController. The same pending disclosure coalesces; a different Review supersedes it.
- Reading requires the same still-authorized disclosure, sheet, Inspector context, and request. A source outside the visible query uses the existing paged reader. Its accepted committed row permits off-query Inspector focus. Stale source acceptance blocks navigation.
- A later user pointer, keyboard, wheel, focus, input, or composition interaction cancels read/navigation intent. The initiating Review gesture and Review-owned semantic focus are exempt. Input-origin cancellation defers feedback rendering until native input handling completes. Sheet departure, unmount, disclosure/authority retirement, and context changes also fence late results.
- After a short delay, the source Review control exposes `aria-busy` and its notice shows a sibling polite status while Review remains keyboard enabled. A current genuine failure shows a local alert and the same Review control retries. Cancellation and supersession are silent; an older result cannot replace newer feedback.
- Success selects and focuses the original source-bound mention in Inspector. It leaves the query and disclosure in place and sends no mutation. Undo, exact replay, accepted receipts, drafts, and authority state retain their preexisting owners.

## Verification and limitations

Commands below ran from the repository root with the pinned `tmp/node-runtime/bin` on `PATH`. The user-requested owner task guides for `web.workbook`, `module.timeline`, and `module.entities` were checked. Earlier attempts without that `PATH` failed at service startup because a child launched literal `node`; they are environment failures, not product passes. The first browser setup used a nonfilterable field and failed at filter selection; a later narrow test required closing the overlaid Inspector before the real collection edit. Those test setup failures were repaired and rerun.

| Command or row | Result and artifact |
| --- | --- |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.cancelled_review_source_read_stops_paging_0de2c50ecb,web.workbook.regression.timeline_review_notice_pending_and_retry_feedbac_8ccfa10010,web.workbook.regression.timeline_mention_auto_resolution_undo_ownership_9e834f7b55` | PASS; final `.cartulary/test-results/20260926T022910Z-p92701` |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.timeline_review_read_preserves_newer_editing_449509d125` | PASS, three production browser scenarios; final `.cartulary/test-results/20260926T022910Z-p92687` |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.auto_resolution_feedback_characterization` | PASS, five retained disclosure/Undo, density, composition, and authority scenarios; `.cartulary/test-results/20260926T022358Z-p5392` |
| `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.the_browser_workbook_shows_auto_resolution_only_758d41a16f,module.entities.browser_stateful.verify_manual_mention_resolution_dismissal_auto_dfa355e592` | PASS, existing auto-resolution and disclosure/Undo browser rows; `.cartulary/test-results/20260926T022613Z-p44050` |
| `make test-slice OWNER=package.ui` | PASS, existing selector contracts; `.cartulary/test-results/20260926T022722Z-p79426` |
| `make frontend-typecheck`, `make lint-biome` | PASS; `.cartulary/test-results/20260926T022917Z-p2422`, `.cartulary/test-results/20260926T022917Z-p2479` |
| `make frontend-import-boundary-check` | Initial FAIL from a new sibling-hook type import at `.cartulary/test-results/20260926T022345Z-p98681`; moved the scope and feedback types to Timeline models. Final PASS at `.cartulary/test-results/20260926T022917Z-p2493`. |
| `make generate`, `make generate-drift`, `make generated-artifact-policy-check` | PASS; `.cartulary/test-results/20260926T021913Z-p50921`, `.cartulary/test-results/20260926T022917Z-p1941`, `.cartulary/test-results/20260926T022917Z-p2058` |
| `make json-shape-check`, `make harness-contract` | PASS, including authored test catalog/routing validation; `.cartulary/test-results/20260926T022917Z-p2198`, `.cartulary/test-results/20260926T022918Z-p3030` |
| `make lint-markdown` | PASS after the final ledger update; `.cartulary/test-results/20260926T023106Z-p37926`. |
| `make agent-finalize` | PASS with `RESULTS_DIR` unset at `.cartulary/test-results/20260926T022844Z-p88824`; retained-run maintenance skipped because no qualifying successful full warm check was supplied. |

Browser coverage holds one query response while unit coverage proves that an ignored abort cannot publish and cancellation prevents a second paging request. No visual golden was changed. The browser test uses a supported query filter and direct gestures; it does not insert a product test hook or repair focus after asserting continuity. A pre-fix browser red for the original unretained path was not captured because the first setup attempt stopped before Review; the original failure is code-confirmed, while the corrected path has current green browser evidence. No full warm `make check` was run because the owner-selected focused, service-backed, and static gates covered the bounded change.

## Digest acceptance assessment

| Rows | Assessment and current evidence |
| --- | --- |
| A001–A003 | PASS. Exact Core/design owner map, baseline, rubric, changed-path and authored-routing review above. |
| A004–A009 | N/A. No token, theme, density geometry, creation, responsive chrome, or shell overflow implementation changed. The 768px Review interaction is exercised by the new browser row. |
| A010–A011 | PASS. Inspector scope and off-query focus tests, scalar/collection browser continuity, and same-row retargeting hook tests. |
| A012–A013 | N/A. Transaction IDs, uncertain replay, queue acknowledgement, and queue recovery remain with their existing mutation owners; Review issues no write. Existing Undo characterization passes. |
| A014 | PASS. Held-read browser cases assert native text, element identity, caret, selection, and focus after later scalar and collection authoring. |
| A015–A016 | N/A. Conflict and query data-state producers were not changed; Review uses source-local status. |
| A017 | PASS. Authority retirement, sheet departure, and unmount hook tests fence late results; existing session/access-loss disclosure browser row passes. |
| A018 | N/A. Evidence presentation/lifecycle was not changed. |
| A019–A020 | PASS. Keyboard-enabled busy Review and local alert/retry tests, 768px collection case, and current compact/default/comfortable disclosure characterization. No additional conformance claim is made. |
| A021–A022 | N/A. Grid virtualization and visual fixtures/goldens were not changed; the off-query Inspector focus path uses the committed ledger. |
| A023–A026 | PASS. Existing semantic UI selectors are used and `package.ui` tests pass; tests have no Markdown dependency; authored routing, catalog contract, and generated drift pass; no route, schema, storage, or data migration exists. |
| A027 | PASS. This handoff records owner boundaries, retirement, compatibility, limitations, rollback, and each current pass/failure. |

No historical handoff is counted as a current pass. There is no remaining applicable blocker.
