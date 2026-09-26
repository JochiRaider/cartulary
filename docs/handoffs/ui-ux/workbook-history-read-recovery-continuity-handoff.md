# Workbook History read-recovery continuity handoff

## Baseline, authority, and observed gap

The work began on clean `main` at `b6362ccf41d5ad6178170f930b82efb06180b18a`; no unrelated files were changed. `AGENTS.md`, the localized UI/UX digest navigation and rubric, the record-history browsing and recovery handoffs, and the current Inspector and History source guides were inspected. Historical passes supplied context only. The current narrow History unit baseline passed 4/4 execution units at `.cartulary/test-results/20260926T140521Z-p7131`.

Core 01 §§3.3.4.2 and 3.3.7 own row-centric reads, server order, opaque cursors, and live authorized continuation. Core 03 §2.3A, REQ-03-138/285, §10, and REQ-03-298/299/100 govern Inspector subject, review versus attachment lifetime, retained authoring, and authority. Design §§8.5, 10.8, 12.7–12.8, and 14 govern focus, local recovery, and accessibility. No owner contradiction or normative amendment was found. `web.workbook` owns the shared client implementation; `module.revisions` routes the service-backed browser evidence. Neither routing nor the advisory digest defines product behavior.

The defect was reproduced through the production Timeline Inspector. After a failed continuation, Enter on **Retry older entries** admitted the retry and held its response. The button disappeared before settlement: `toBeAttached` failed at `.cartulary/test-results/20260926T142012Z-p12098` (9/11 execution units). This is a confirmed pending-interval failure. The existing test's manual removal of the retry element had characterized settlement fallback, not pending reachability. Suspected effects on other surfaces were hypotheses until the shared-consumer browser runs passed after the correction.

## Decision and implementation

The bounded correction gives the presentation continuity hook one transient recovery-control lease keyed to the record/view, authority scope, chain/request generation, and initiating element. The existing browsing state remains the sole accepted-page representation and read-admission owner. The shared panel renders the initiating action across failure → pending even though `beginHistoryRead` correctly clears the failure. It stays focusable with `aria-busy`, `aria-disabled`, and guarded activation. No stale error is retained to mount it.

On repeat failure, the current retry or explicit fresh-chain action remains local and usable. On success, focus and viewport alignment move to Refresh or Load older only while the initiating interaction still owns focus. A terminal continuation leaves Load older focusable and unavailable. Native Tab departure and newer pointer or scroll input keep the user's destination and scroll; a focused control whose return was cancelled briefly shows a neutral completed label until blur. Subject replacement, authority loss, superseding reads, and detachment retire obsolete presentation. The disconnected-trigger/body focus fallback and failure-only recovery rendering were removed together.

This encapsulates one common decision already shared by Timeline, Generic, Entity, Assessment, and History review/recovery panels. It avoids surface-specific flags and permits another current consumer to use the same panel without a second data store. Exact failed-request retry still reuses its cursor; **Start fresh history** still issues a cursor-free initial or refresh read. Server order, page retention, exhaustion, representation checks, and stale-response fencing remain in the browsing/controller owners. Rollback, delete, restore, confirmations, captured mutations, receipts, and uncertain-write recovery were not moved. Read recovery sends no mutation.

Changed authored paths are the shared Inspector panel, continuity hook and focused test; the History browsing characterization test; `apps/web/e2e/history-browsing.spec.ts`; the Inspector source guide; and the authored `web.workbook` and `module.revisions` test-family manifests. Make regenerated only `tools/browser_e2e_batch_manifest.json` and `tools/execution_topology_render_index.json`. No public API, schema, route, storage, dependency, migration, or golden changed. The risk of leaving the gap was a disconnected keyboard position during an admitted read and possible focus theft at settlement. The binary exit was a connected, focused, busy recovery control while held; one admitted read per repeated activation; correct focus/scroll on settlement; all four consumers retaining accepted History and explicit rollback behavior; and zero mutations from recovery reads.

## Verification and limits

| Check | Result and evidence |
| --- | --- |
| Focused History and Inspector draft unit rows | PASS 6/6 on the final changed tree at `.cartulary/test-results/20260926T145307Z-p95769`. Covers all four held control kinds, repeat failure/success, exact cursor versus restart, terminal control, scope/subject/supersession/detachment/interaction invalidation, close/reopen retry, retained pages, and independent draft binding/lifetime. The draft owner was unchanged. |
| Five `module.revisions` History browsing browser rows | PASS 11/11 at `.cartulary/test-results/20260926T144623Z-p52311`. Real delayed responses cover Timeline, Generic, Entity, Assessment, keyboard pending reachability, repeat activation, refresh retry, cursor-free restart, server order, accepted entries, terminal paging, and the existing rollback. |
| Focused narrow Timeline initial retry and newer pointer/scroll row | PASS 11/11 at `.cartulary/test-results/20260926T144311Z-p1125`. A 1024×720 Inspector keeps the initial retry in view; native Tab/Shift+Tab works while held. A later held refresh preserves a clicked History disclosure and the observed Inspector scroll position. Full server History is unchanged and zero record mutations are observed. |
| `make agent-finalize` without `RESULTS_DIR` | PASS on the final changed tree at `.cartulary/test-results/20260926T145242Z-p91829`. Retained-run maintenance was skipped because no qualifying successful full warm check root was supplied. |
| Final static and generated checks | `make frontend-typecheck` PASS at `20260926T145307Z-p95980`; `make lint-biome` PASS at `20260926T145307Z-p96170`; `make frontend-import-boundary-check` PASS at `20260926T145307Z-p96045`; `make test-catalog-check` PASS; `make generate-drift` PASS at `20260926T144535Z-p41607`; `make generated-artifact-policy-check` PASS at `20260926T144535Z-p41678`. |

The first finalizer failed because the newly authored `web.workbook` title made topology inputs stale; `make generate` repaired that projection, and the next finalizer passed. An intermediate initial browser assertion used Generic's refresh selector for Timeline; another recorded the wheel position before scrolling completed. Both were test corrections, followed by passing isolated and combined runs. An intermediate Biome run found only formatting in the browser test; `make format` and the passing rerun resolved it. No product assertion was weakened to hide the confirmed pending failure.

Full `make check`/`make ci`, release, deployment, broad visual readiness, and two-pass visual golden validation were not run: this slice changes neither visual goldens nor backend/public contracts. The narrow browser layout and focus assertions are implementation evidence, not a whole-page accessibility or Core 05 publication claim. Tests, generated files, and runtime do not read Markdown. Rollback reverts this coordinated client, test-routing, generated, and guide/handoff diff; stored History, revisions, and analyst data require no rollback.

## Digest acceptance assessment

| Result | Criteria and scope rationale |
| --- | --- |
| PASS | A001–A004: owner/source/routing mapping, bounded shared decision, clean baseline and stack review, and no new token literal or registry. |
| PASS | A009–A011: narrow Inspector control/scroll reachability, canonical subject, and recovery focus continuity. |
| PASS | A013, A016–A017: read-only recovery after failure, retained authorized entries, and scope/authority invalidation; no write replay. |
| PASS | A019–A020: applicable pending keyboard, focus, busy/unavailable state, and 1024×720 layout checks; no whole-page conformance claim. |
| PASS | A023–A027: semantic selectors, Markdown-independent executable checks, Make-generated projections, unchanged public compatibility, and this evidence-backed handoff. |
| N/A | A005–A008: theme, density, creation, and responsive threshold algorithms were not changed. |
| N/A | A012, A014–A015, A018, A021–A022: transaction capture, editing/conflict, Evidence, grid virtualization, and visual goldens are outside this read-control lifetime change. Existing owners remain intact; no new golden was needed. |

No applicable criterion is blocked. `make lint-markdown` passed after the handoff edit at `.cartulary/test-results/20260926T145307Z-p96139`; `git diff --check` and changed-path/status review were clean. The branch remains `main` at the baseline HEAD, with only the coordinated paths above modified or added; no commit, push, or deployment occurred.
