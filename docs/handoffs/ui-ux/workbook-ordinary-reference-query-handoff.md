# Ordinary reference-picker queries

Status: complete.

## Baseline, scope and owners

Implementation started on clean `main` at
`54f674bee579568f9be6d9c0198dc5e99278223b`, one commit ahead of `origin/main`.
That commit and the landed destination-handoff work were preserved. The change
reuses candidate-query presentation in existing-record reference pickers. It
does not change membership discovery, query capabilities, mutations or paging.

Behavior remains governed by Core 01 §§3.3.4, 3.3.7, 7.4.8, 18B and applicable
field/action contracts; Core 03 §§13, 14.1, 16.2, 16.4 and REQ-03-299/100;
and design §§7–8 and 14. The UX capability improvement is authorized by the
requested journey; Core 03's sheet-query clauses do not independently require
identical controls in every picker. Domain vocabulary, digest guidance and
historical handoffs are navigation and supporting evidence.

Source ownership is `web.workbook`; verification independently routes through
`web.workbook` and `module.workbook`. No specification, wire or data migration
was needed. No executable consumer depends on this handoff or other Markdown.

The common decision is contract-aware typed candidate-query editing. Reusing
`WorkbookCandidateQueryControl` removes the ordinary text-only path and lets
future declared source capabilities use the same presentation without another
query owner. Ordinary selection and contextual discovery remain separate.
The principal risks are source/controller replacement and disclosure focus in
the compact popup; dedicated characterization and rendered checks address them.

| Workstream | Status | Exit |
| --- | --- | --- |
| Characterization | PASS | Seeded existing Task picker: 66 Timeline candidates, only Capture State, no ordering. Registered boundary test failed at missing query disclosure before production edits. |
| Integration | PASS | Shared query presentation, legacy retirement and lifecycle/focus coverage; focused rows pass. |
| Verification and review | PASS | Selected product checks, desktop/compact review, cleanup and acceptance completed. |

## Evidence

Planning revalidated `make help`, `make help-all` and both module-author task
guides. The three requested regression rows passed freshly with
`CARTULARY_HARNESS_CACHE_MODE=off` in
`.cartulary/test-results/20261007T200234Z-p51632`. The earlier planning run
`.cartulary/test-results/20261007T200048Z-p50278` reused cached row results and
is not fresh product execution. Neither is a full warm check.

The seeded baseline session at
`.cartulary/test-results/20261007T201258Z-p84761` reproduced the restriction;
its terminal receipt confirms cleanup complete. An earlier session at
`.cartulary/test-results/20261007T201125Z-p55131` failed sealed-source admission
because characterization edits overlapped build preparation; cleanup completed.
Neither session produced a visual or accessibility assessment.

Characterization first ran unregistered and was skipped at
`.cartulary/test-results/20261007T201230Z-p83573`; it is not red-test evidence.
After authored title registration and `make generate`, it failed as intended at
`.cartulary/test-results/20261007T202112Z-p27620` (15 passed, one failed).
The integration's first run at `.cartulary/test-results/20261007T202258Z-p31682`
caught fixture assumptions about canonical filter ordering and declared
Indicator fields, plus a real Party-identity integration omission. Those were
corrected. All three focused rows now pass at
`.cartulary/test-results/20261007T202429Z-p41132`.

The ordinary picker uses the unchanged shared query control for all record-view
sources, including Party identities. Incident members retain their unfiltered
membership path. Source keys reset incompatible query drafts; replacing the
reader, parent record or field closes the existing picker/controller lifetime.
Applying a query preserves the disclosure and initiating focus. The popup's
focus enumeration includes summaries and skips closed disclosure descendants;
Enter in a text operand does not implicitly submit an enclosing form.

The former text field/value state, chosen-operator logic and manual filter
payload were removed. No hidden compatibility branch remains. The existing
recovery browser scenario now stages through the shared control. The new browser
row comes from the authored row-ID workflow and verifies existing-Task linking;
generated routing was refreshed through `make generate`.

## Verification and rendered review

All commands ran from `/home/jochi/code/cartulary`. The combined focused command was:

```sh
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=web.workbook ROWS=web.workbook.regression.reference_controls,web.workbook.regression.reference_selection,web.workbook.regression.authoring_candidate_presentation
```

It passed freshly at `.cartulary/test-results/20261007T202713Z-p11663`:
19 ordinary-control, 9 selection/reader and 22 contextual-presentation tests,
50 total, none skipped. Shared operand permutations remain in their existing
coverage; new boundary cases cover boolean/order, corrected date range, enum,
exact literal sets, local staging/removal, unfinished unadded operands, cursor
reset, retained selections, source/reader replacement and enclosing-form safety.
Existing controlled failure/race, authority and focus cases still run.
After removing an obsolete guard from the migrated test, the ordinary-control
row passed freshly again at `.cartulary/test-results/20261007T203532Z-p46289`
using `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.reference_controls`.

The browser command selected these rows through
`make service-backed-test-slice OWNER=module.workbook ROWS=<comma-separated rows>`:

- `module.workbook.browser.existing_task_reference_queries_stage_typed_time_f92c8c967a`
- `module.workbook.browser.reference_selection_recovery`
- `module.workbook.browser.reference_grid_commit`
- `module.workbook.browser.reference_session_recovery`
- `module.workbook.browser.reference_target_eligibility`
- `module.workbook.accessibility.reference_picker`
- `module.workbook.browser.literal_contextual_candidate_staging_1c1261f5df`

At `.cartulary/test-results/20261007T202435Z-p41788`, all six existing rows
passed freshly. The new row failed because its Evidence fixture lacked an
available uploaded object. Has Evidence correctly excludes that state. The
fixture now uses `createUploadedEvidenceFixture`. Its isolated rerun at
`.cartulary/test-results/20261007T202702Z-p96335` reached the saved result but
failed an incorrect assumption about persisted collection presentation order.
The test now asserts ordered submitted actions and persisted membership
independently. The final isolated rerun passed freshly at
`.cartulary/test-results/20261007T203313Z-p98491` (11/11 execution units).
All seven selected browser rows now have fresh passing evidence.

These browser rows were selected for concrete integration risks: preserving
grid acceptance, read retry and acknowledged-write recovery, session retirement,
direct-reference eligibility/clear, keyboard popup behavior, and compatibility
of the existing contextual consumer. No unrestricted product suite was run.

| Command | Result and exact run root |
| --- | --- |
| `make generate` | PASS; final authored routing generation `.cartulary/test-results/20261007T202411Z-p32868`. Only routing batch/index projections changed. |
| `make format` | PASS; final `.cartulary/test-results/20261007T203519Z-p39139`. Diff contains only intended source/test formatting. |
| `make agent-finalize` | Initial PASS `.cartulary/test-results/20261007T202507Z-p75743`, before broader verification, with retained-run maintenance skipped because `RESULTS_DIR` was unset. Follow-up maintenance PASS `.cartulary/test-results/20261007T210557Z-p56767` using the full check below. |
| `make frontend-typecheck` | PASS `.cartulary/test-results/20261007T202714Z-p13579`. Earlier `.cartulary/test-results/20261007T202603Z-p94422` found two new test-only `exact` options unsupported by Testing Library; removed. |
| `make frontend-import-boundary-check` | PASS `.cartulary/test-results/20261007T202603Z-p94448`. |
| `make lint-biome` | PASS; final `.cartulary/test-results/20261007T203614Z-p47231`. |
| `make lint-markdown` | PASS; final content review `.cartulary/test-results/20261007T203614Z-p47239`. |

`cartulary-ui-review` seeded review at
`.cartulary/test-results/20261007T202744Z-p39874` used repository-owned synthetic
data and the default `network_flow_claimed` profile. Three captured images were
verified against their bundle digests and visually inspected. At 1440×900 the
expanded query, candidates, selection summary and actions fit. At 1024×768,
keyboard Tab scrolls the bounded popup to expose the retained selection summary
and Use selection/Cancel actions; no horizontal popup overflow was measured.
Keyboard Apply retained visible focus and reduced 66 candidates to two only
after explicit application. An off-filter identity survived; native keyboard
selection added another. Keyboard cancellation left the parent field empty.
The contextual Task picker independently returned the same two candidates and
retained its source identity. No review write was submitted.

Axe completed with no violations in those captures and incomplete contrast
results, including native option content. This is advisory review evidence,
not product assertions or accessibility certification. Long native candidate
labels retain their existing clipping; full selected labels wrap in the summary.
No shared-control styling or golden refresh was needed. The exact session was
stopped, its terminal receipt says cleanup complete, the foreground process
exited, and only caller-owned scratch requests were removed. Startup attempts
with machine output were rejected before session creation; startup requires
the standard `make ui-review` output mode. A few ambiguous/unavailable targets
were resolved using fresh observations and existing keyboard navigation.

## Retained-run maintenance follow-up

The user subsequently requested retained-run maintenance. The earlier successful
full check at `.cartulary/test-results/20261007T162949Z-p11436` had a different
source digest and was ineligible for the current implementation. The following
commands ran from `/home/jochi/code/cartulary`:

```sh
make target-plan TARGET=check
make check
make agent-finalize RESULTS_DIR=.cartulary/test-results/20261007T204755Z-p53701
```

The full warm-cache check passed at
`.cartulary/test-results/20261007T204755Z-p53701`: all 1,014 work units passed,
with 899 executed and 115 cache hits, no failures, skips or cancellations.
Its source digest matches the completed implementation. `make explain-run`
confirmed 1,032 cleanup steps with no failed or blocked steps.

The finalizer passed at `.cartulary/test-results/20261007T210557Z-p56767`.
`unit-artifacts/finalize-summary.json` records `results_dir_status: valid`,
selection of the latest compatible successful check without an older-run
override, and PASS for scheduler drift, schema shape, tier coverage, generated
structure refresh and canonical evidence validation. Generated files were
unchanged. Retained-run maintenance is now complete; the earlier skipped result
remains historical evidence. This follow-up changed only this handoff.
`make lint-markdown` passed at
`.cartulary/test-results/20261007T210702Z-p61948`; `git diff --check` also passed.

## Digest acceptance within this seam

| Rows | Assessment |
| --- | --- |
| A001–A003 | PASS: adopted owners and separate source/routing ownership, baseline, source guides, bounded reuse decision and legacy retirement recorded above. |
| A004–A005 | PASS: no new design literals, theme or density registry; existing graphite presentation reused. |
| A007 | PASS for contextual compatibility only: unchanged authoring owner and selected source; contextual regression and rendered comparison. |
| A008–A011 | PASS for popup/inspector scope: both requested viewports, bounded scrolling, local staging, explicit acceptance and replacement/dismissal fencing. Shell geometry and feature routing unchanged. |
| A013–A014 | PASS: existing read recovery versus write ownership, explicit Update, raw cancellation, input correction and keyboard cases. |
| A016–A017 | PASS: candidate data does not confer authority; selection/controller concealment and session retirement scenarios pass. |
| A019–A020, A022 | PASS for selected keyboard/component review: public accessibility row plus desktop/compact captures, stated axe limitations and no golden changes. |
| A023–A026 | PASS: semantic view/record/field selectors, no Markdown execution dependency, authored routing generated by Make, unchanged interfaces and no migration. |
| A006, A012, A015, A018, A021 | N/A: no density, transaction identity/replay, conflict UI, Evidence lifecycle or virtualization implementation changed. Existing owners remain authoritative. |
| A027 | PASS: completed exits, checks, failures and corrections, review limits, cleanup and rollback recorded. |

## Compatibility and rollback

Applied queries, cursor chains, retries and ordered selections remain in the
ordinary selection controller. Query-edit state stays with the shared local
presentation. Parent drafts, explicit inspector Update and grid acceptance
remain with their existing owners. Rollback is a frontend integration and test
routing revert; persisted data requires no migration.

Production changes are confined to `WorkbookReferenceControl.tsx`. Its component
tests, `workbook-inspector-edit.spec.ts`, two authored test-family manifests and
their generated routing outputs supply coverage. The shared query control,
query model, selection controller, reader/port, inspector/grid/coordination write
owners and Timeline contract were inspected and left unchanged. The one-page,
100-candidate, ten-checkpoint and 64-selection/direct-single limits remain intact.

Final working-tree review confirms only this slice's eight files are changed
(including this handoff), with no unrelated formatter rewrites. `git diff --check`
passes. Branch remains `main` at the original HEAD, zero behind and one ahead of
`origin/main`. Existing local commit and destination-handoff work are preserved;
no commits were reset, amended or created. No applicable verification is blocked.
The initial implementation used bounded verification; the explicitly requested
maintenance follow-up also completed the full warm check above. No further
product or retained-run maintenance action is required.
