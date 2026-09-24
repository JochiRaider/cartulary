# Requested Filters during query replacement

## Decision and owners

This bounded correction began on clean `main` at `986689839d89a879bb4c4b47faeee94f314e6b63` (the local branch was three commits ahead of `origin/main`). Core 03 §14.1/REQ-03-224 and §14.9 own filter query behavior, accepted rows, and replacement browsing. Design §8.3 directs the accepted chip strip, chip activation, and compact Unapplied presentation. `docs/domain.md` supplies vocabulary; the UI/UX digest is advisory. Current frontend source and import ownership place the implementation in Workbook. `web.workbook` routes the focused unit rows; `module.workbook` routes service-backed browser rows. The verification catalog does not define product behavior.

At baseline, applying a Timeline filter closed and cleared the form. With the read held, reopening Filters exposed only accepted filters; a failed edit likewise hid the latest requested value. The new production-grid regression failed as expected at the missing Unapplied cue: `.cartulary/test-results/20260924T221342Z-p76799`. The accepted rows and chips were retained, but the requested criterion was not discoverable for editing.

The shared decision is a derived comparison of accepted and requested filter criteria by field key. `WorkbookQueryBrowser` remains the accepted/canonical and request-lifecycle owner; the existing surface query state remains the requested owner; the Filters reducer remains a temporary form owner. The gap risk was repeated input or accidental editing of an older accepted value during pending/failure. The correction benefits other registered surfaces through the existing shell binding. No persistent draft, toolbar, request owner, or new query store was introduced. This is a behavior correction with a small projection, not a broader query refactor.

## Changes and compatibility

- `workbookGridQueryControls.ts` derives requested additions, edits, and removals from accepted/requested filters. `WorkbookGridControls.tsx` keeps accepted chips and accepted-list activation tied to retained rows, while projecting the requested criteria into Filters.
- `WorkbookFiltersControl.tsx` shows a compact Unapplied cue, editable requested entries, pending-removal Restore, and pending-addition Remove. Clear acts on requested filters; accepted entries remain separately labeled. The same overlay navigation owns keyboard order, outside dismissal, and invoking-control focus return.
- The Timeline presentation and shared query controller pass `canonicalIntent(...).filters` beside the existing requested Sort and Group projections. Apply, Remove, and Clear update the latest requested query through functional setters, preserving unrelated filters, Sort, and Group. Canonical acceptance clears the cue; open form input and focus survive that response.
- The former Timeline runtime `applyQueryFilter` helper and the accepted-only Filters-list assumption were retired. Retry and Revert remain in the browsing owner. No route, schema, saved-view format, persistence, record mutation, authorization, or dependency changed. All supported consumers use the migrated internal control prop; no compatibility adapter or data migration is needed.
- Authored test catalog rows route the unit and browser cases. `make generate` produced the corresponding browser batch and topology index changes; generated outputs were not edited by hand. Executable checks do not read Markdown.

Rollback restores the authored Workbook, test, and catalog changes together and reruns `make generate`; no data rollback is required. The local form has only its existing overlay lifetime. Reload or cross-tab persistence is outside this correction.

## Verification

| Evidence | Result |
| --- | --- |
| Expected-red Timeline browser reproduction | Failed at missing Unapplied cue, `.cartulary/test-results/20260924T221342Z-p76799`. |
| New Timeline and Notes service-backed row | PASS 11/11, `.cartulary/test-results/20260924T224435Z-p77080`, including held/failed addition and edit, consecutive edits, pending Clear/Restore, Retry/Revert, canonical acceptance, stale response, accepted rows/chips, exact requests, focus, and no record writes. |
| Focused `web.workbook` model, component, controller, and browsing rows | PASS 6/6, `.cartulary/test-results/20260924T224128Z-p99176`. |
| Existing filter keyboard, Group replacement, recovery focus, and canonical continuation browser rows | PASS 13/13, `.cartulary/test-results/20260924T223919Z-p57591`. |
| `make agent-finalize` | PASS, `.cartulary/test-results/20260924T223305Z-p66043`. `RESULTS_DIR` was unset, so retained-run maintenance was skipped. |
| Broader typecheck, import boundary, Biome, catalog, generation drift, artifact policy, JSON shape, and Markdown checks | Final results below. |

The first existing filter-keyboard rerun found an unordered two-row expectation; comparing record identities without order kept its keyboard/focus assertions intact. A compact-case selector initially missed the applied overflow entry's accessible prefix; its snapshot confirmed both applied and Unapplied entries and the selector was corrected. `make format` stopped after formatting touched frontend files because Biome flagged a Promise truthiness condition in the new browser route gate; the explicit null comparison resolved it and `make lint-biome` passed. These are test/format corrections, not remaining product failures.

### Final run evidence

- `make generate`: PASS, `.cartulary/test-results/20260924T222859Z-p88980`.
- `make format`: PASS, `.cartulary/test-results/20260924T224425Z-p72675`.
- `make frontend-typecheck`: PASS, `.cartulary/test-results/20260924T224514Z-p7934`.
- `make frontend-import-boundary-check`: PASS, `.cartulary/test-results/20260924T223635Z-p95999`.
- `make lint-biome`: PASS, `.cartulary/test-results/20260924T224446Z-p94728`.
- `make test-catalog-check`: PASS.
- `make generate-drift`: PASS, `.cartulary/test-results/20260924T223634Z-p95685`.
- `make generated-artifact-policy-check`: PASS, `.cartulary/test-results/20260924T223634Z-p95756`.
- `make json-shape-check`: PASS, `.cartulary/test-results/20260924T223827Z-p35304`.
- `make lint-markdown`: PASS, `.cartulary/test-results/20260924T223828Z-p35580`.

## Digest acceptance disposition

| Rows | Disposition and evidence |
| --- | --- |
| A001–A004 | PASS: exact Core/design owners and separate source/verification routing above; clean baseline and current source checked; one derived comparison boundary; no token/theme change. |
| A011, A014 | PASS: held-response input/focus, explicit Apply, Escape/Cancel, accepted chip activation, and stale-response browser coverage. |
| A016–A017 | PASS for the affected query state: accepted rows/chips remain through pending/failure, canonical success and recovery reconcile through the browsing owner; authorization-loss behavior was not changed. |
| A019–A020 | PASS: native filter keys, focus return, requested and accepted entries, and compact overflow are covered by production-grid/browser evidence. |
| A023–A027 | PASS: semantic selectors and owner-routed rows; no Markdown dependency; generated outputs from `make generate`; no external compatibility or data change; this handoff records evidence, rollback, limitations, and skipped maintenance. |
| A005–A010, A012–A013, A015, A018, A021–A022 | N/A: theme, density, creation, shell geometry, inspector, write replay, conflict, Evidence, virtualization policy, and visual fixtures were outside this Filters correction. Existing grid/renderer owners were unchanged. |

Digest rules R001–R004 and R007–R008 were adopted through the existing keyboard, accessible text, retained-query feedback, and owner recovery patterns. R012 and R035 were adapted to the current React query lifetimes and compact Workbook controls. No upstream palette, component system, or separate persistence guidance was applied.
