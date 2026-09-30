# Readable retained authoring selections

Implementation date: 2026-09-30. The bounded correction and focused product checks
pass. Completion remains blocked by the inherited contextual visual comparison
described below. No commit, push, deployment, migration, or golden update was made.

## Baseline and authority

The checkout started clean on `main` at
`25a41ee5c9beda9db480d2925a4940e43d1c1c44`, four commits ahead of `origin/main`.
HEAD still matches that reviewed baseline; there were no intervening source
changes or unrelated local edits to preserve. Root `AGENTS.md` was the only
applicable repository agent file. Current Make guidance, source guides, authored
ownership/routing inputs, and generated-artifact policy were inspected.

The supplied production reproduction was editor → populated review incident →
Timeline record → Inspector → Create task request → default Owner → Choose Owner.
The native option became “Review editor,” but the selected list and removal name
showed the actor UUID. Unchanged-identity Apply degraded the parent summary to the
UUID; keep/close and Recovery preserved that presentation and the authored Title.
This established contradictory labels, not an incorrect saved assignee or lost
draft; no creation was submitted.

Human authority review covered the digest README, START_HERE, LOCAL_AGENT_PROMPT,
REPO_MAP, OWNER_MAP, QUERY_RECIPES, rules, and acceptance material, plus the actual
adopted owners. Research, historical handoffs, and bundled prompts remained source
material and did not expand the task. No adopted-owner contradiction or normative
specification change was identified.

| Owner | Application to this correction |
| --- | --- |
| Core 00 | Adopted precedence; downstream code and projections do not define requirements. |
| Core 01 §§7.4, 18B, 19 | Existing creation/reference identities, capability and explicit-write boundaries remain owned by their features. |
| Core 03 §16.4, REQ-03-256 | Retain original contextual source, raw fields, verification and readiness independently of readable labels. |
| Core 03 REQ-03-299/100; Core 04 §§1–2 | Keep concealment, suspension, incident/account retirement and authorization lifetime fences. |
| Design §§6.2, 12, 14 | Ordinary readable reference context, staged authoring, keyboard continuity and usable accessible names. |
| Domain | Vocabulary and owner navigation; no new reference-service concept. |

Source/import ownership is `web.workbook`. Independent browser routing is
`module.workbook`. Routing, runner evidence and advisory material do not establish
specification completeness. No test, runtime, generator or routing dependency on
Markdown was introduced.

## Selected boundary and consumers

`WorkbookAuthoringReferencePicker.tsx` already owns local staged selections. Its
freshness transition cleared selected `displayText` while preserving identity,
but accepted pages did not restore presentation. The correction stays there:
a private pure function decides whether a current authorized candidate can supply
presentation for an existing selection. This encapsulates a real decision shared
by authoring consumers without adding a cache, store, broker or workflow owner.

The helper matches exact `recordId`, distinguishes `incident_members` from record
references, and copies only nonempty `displayText`. Record labels can update
across record-reference surfaces without rewriting retained `viewSchemaId`.
Unchanged labels preserve object and array identity. Freshness clearing and
reconciliation feed one guarded staged update before rendering, so the selected
summary, removal action and explicit Apply use the same selections.

Reconciliation requires an available, enabled, settled accepted page under
readable authority. Missing entries are not deletions. Pending, failed and empty
reads do not manufacture new labels. Selected labels survive permitted page
eviction; full visited candidate rows are not retained. Existing discovery
generation, current-controller, target, reader, revision and authority fences
remain responsible for accepting observations. Concealment still suppresses
protected presentation.

| Inspected consumer/boundary | Result and coverage |
| --- | --- |
| ContextualReferenceControl / ContextualCreateForm / ContextualCreateRecovery / WorkbookContextualTaskDecisionCreateOwner | Ordinary chooser supplies labels only on explicit Apply. Owner/revision/source-readiness unit tests, Timeline Owner/Party browser tests, and accessibility recovery journey pass. Retained owner code is unchanged. |
| OrdinaryCreateControl / WorkbookAuthoringReferenceControl | Shared staging and ordinary raw-ID input remain available. Existing Party browser scenario exercises ordinary authoring and keyboard/removal continuity. |
| CoordinationCreateForm | Ordinary references use the shared correction; Source retains `captureRowVersion`. No feature write or source replacement behavior changes. |
| RelatedEvidencePartyControl | Shared ordinary Party presentation; existing cross-page/failed-continuation browser coverage passes. |
| NoteAssociationPanel | Shared ordinary record selections. Browsing stays local; its existing explicit “Link selected records” Apply callback still dispatches that feature's mutation. |
| NoteSourceControl and Coordination Source | Reconciliation is bypassed under `captureRowVersion`; reviewed label/version, cancellation and explicit replacement semantics are preserved. Extended picker compatibility assertions and the Note-source browser row pass. |
| WorkbookCandidateSelection / WorkbookRecordCandidatePicker / AssessmentDiscovery | Inspected but unchanged. No generic reconciliation helper changes Assessment behavior; the expanded authoring browser batch includes its existing scenario. |
| WorkbookReferenceControl | Separate existing-record editing path; unchanged. |
| useWorkbookCandidateDiscovery / WorkbookCandidateDiscovery / createWorkbookAuthoringReader / authoring and candidate read ports | Inspected but unchanged. Bounded requests, read admission, authorization and source payload ownership stay here. |

Ordering, cardinality, membership, captured row versions and other metadata remain
unchanged. Labels grant no eligibility, permission, verification or readiness.
Opening/loading never calls the parent or submits. Cancel/Escape preserves parent
values and metadata. Only explicit Apply reaches the existing consumer callback;
only an existing feature submission saves the contextual Task. DOM keys and focus
owners are unchanged.

The benefit is consistent readable selected context and an applied summary that
survives retained-draft recovery. A future caller can use the existing ordinary
picker contract without acquiring a label service. No speculative extension was
implemented. Remaining risk is limited to observed selected presentation and the
inherited visual gate; this is not a claim that every consumer/collaboration
combination has been reproduced.

## Characterization and regression evidence

Before production edits, routed assertions failed at
`.cartulary/test-results/20260930T192103Z-p4576`. Both the picker member test and
the contextual Owner test reported the UUID where “Review editor” was expected.
The contextual revision test could not find the removal name using the refreshed
source label after the new page was accepted. Exact diagnostics are retained in
each row's `unit-logs/row-web.workbook.regression.<row>/vitest-failure-details.json`.
These establish the diagnosed missing reconciliation, including the explicit
freshness-clear → accepted-page transition.

One new duplicate-label fixture also failed because its retained Timeline schema
selected an unavailable initial surface. The first post-fix run
`20260930T192245Z-p6856` exposed the same fixture mistake; explicitly selecting the
Party discovery surface repaired that fixture. It is not evidence of the label
defect. New TypeScript checks also rejected four test-only Testing Library `exact`
options; those were removed without changing product behavior.

Coverage now includes default member fallback, accepted record renaming, held
revision invalidation and recovery, exact IDs across record surfaces, namespace
separation, duplicate/long labels, selected-only metadata, multiple/off-page
selection and removal, twelve-page eviction, pending/failure/empty pages, Apply
versus cancellation, obsolete/authority/disposed completions, suspension and
account replacement. Source verification/readiness, reviewed source label/version
and explicit replacement remain separately asserted. Same controls remain
connected and focused while labels change. Parent-readable keep/close/resume
assertions preserve the Title and actor identity and observe zero Task creations.

Exact new titles were added to `tools/test_families/web.workbook.json` and
`tools/test_families/module.workbook.json` before Make-owned generation. The latter
adds the retained Owner and retained Party-label scenarios. Only the generated
execution-topology render index changed downstream. No generated root was edited
by hand.

All commands below ran from the repository root. Run IDs resolve under
`.cartulary/test-results/`; graph work units are not test counts.

| Command / evidence | Outcome and exact run root |
| --- | --- |
| `make help`; `make help-all`; `make task-guide ROLE=module-author OWNER=web.workbook`; matching `OWNER=module.workbook` | Active rows and public routing revalidated, including both visual rows. |
| `make doctor` | PASS, `20260930T192103Z-p4401`. |
| Initial `make agent-finalize` / `make json-shape-check` | Expected stale generated routing failure after authored titles, `20260930T192005Z-p3749` / `20260930T192103Z-p4522`; repaired through generation. |
| First `make generate` | Failed on new scenario IDs not being ASCII-sorted, `20260930T192552Z-p8415`; authored ordering corrected. |
| `make generate` | PASS, `20260930T192804Z-p11837`. |
| Focused unit slice (command below) | RED `20260930T192103Z-p4576`; fixture-only failure `20260930T192245Z-p6856`; PASS `20260930T192805Z-p12333`, `20260930T193055Z-p24545`, and final `20260930T195513Z-p86632`. Final runner results: 20 picker tests and 8 contextual authoring tests pass; 3/3 graph units. |
| Requested browser slice (command below) | PASS, `20260930T193056Z-p24767`; all three semantic rows pass. Harness expands the authoring-candidates batch to 6 functional scenarios, plus 1 accessibility scenario. No skipped/flaky scenarios; 13/13 graph units. |
| Focused visual slice (command below) | Ordinary row PASS; contextual row FAIL, `20260930T193218Z-p62624`. Two semantic scenarios and 14 capture intents; 9/11 graph units. See inherited blocker below. |
| Baseline-picker visual comparison, same focused command | Same ordinary PASS/contextual FAIL, `20260930T195334Z-p53445`. Only production picker was temporarily restored to reviewed bytes; final bytes restored in `finally`. All six actual images are byte-identical to the changed-picker run. |
| `make format` | PASS, `20260930T192955Z-p15877`, `20260930T193231Z-p84033`; formatter diffs inspected and confined to added source/test formatting. Final formatting/check roots recorded below. |
| `make agent-finalize` | PASS, `20260930T193008Z-p20418`; retained-run maintenance skipped because `RESULTS_DIR` was unset. Final finalization recorded below. |
| `make frontend-typecheck` | Test-query typing failure `20260930T193058Z-p25196`, repaired; PASS `20260930T193336Z-p816`. |
| `make lint-biome` | PASS, `20260930T193059Z-p25453`; final check recorded below. |
| `make generated-artifact-policy-check` | PASS, `20260930T193100Z-p26342`; final check recorded below. |
| `make frontend-import-boundary-check` | PASS, `20260930T193101Z-p29193`; final check recorded below. |

```sh
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.contextual_task_decision_authoring CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.contextual_task_decision_creation,module.workbook.browser.authoring_candidates_parties,module.workbook.browser.authoring_candidates_note_source CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.visual.contextual_task_decision_creation,module.workbook.visual.ordinary_create_authoring_recovery CARTULARY_HARNESS_CACHE_MODE=off
```

Recovery/exact-replay operation rows were not expanded: retained operation owners,
captured attempts and acknowledgement/replay behavior are unchanged. Broad backend,
release and unrestricted browser suites were not justified by this presentation
slice. No successful full warm check is claimed; `RESULTS_DIR` stays unset.

## Production rendered review and cleanup

Used `make ui-review UI_MODE=seeded`, with editor authentication through the review
harness and an isolated synthetic populated incident. Session run
`20260930T193912Z-p3179` served sealed source digest
`2fa4211d9cb43efc8a63f1d9b5fdb115aedac718fd48f2f7b4227850794416d9`.
The reviewed production picker has SHA-256
`175d10ddcc9e43af38e9b6a674a29a49d62c9ae61823f5de9ff84a6e3fc15cc8`;
it was restored byte-for-byte after the baseline visual investigation.

Finite interactions used `make ui-browser`; captures used `make ui-capture`, with
the exact printed session locator, returned epochs and observed targets. Request
files were private caller-owned scratch. The returned bundle manifests supplied
the image and observation paths; every component length/hash and viewport image
dimension was checked before viewing all eight images. No private images,
observations or capability URLs are copied into this handoff.

| Journey | 1440×900 evidence | 1024×720 evidence | Observation |
| --- | --- | --- | --- |
| Timeline → Task default Owner → settled picker | bundle-17, epoch 11 | bundle-68, epoch 49 | Native selected “Review editor,” selected list and removal name agree. Parent remains “Current actor” before Apply. Narrow journey additionally checks Escape before reopening/Apply. |
| Explicit Apply → keep/close → Recovery → resume | bundle-25, epoch 17 | bundle-78, epoch 57 | Parent removal name/summary is “Review editor”; authored Title is intact. Product browser assertions separately prove exact actor ID and no Task creation. |
| Contextual Requester Party, Apply/reopen/accepted refresh | bundle-35, epoch 24 | bundle-89, epoch 64 | Native selection, selected summary and removal name retain “Response coordination team.” Seeded review has one Party page; changed label and actual next/previous replacement are covered in the isolated product browser scenario. |
| Captured Note source chooser | bundle-47, epoch 32 | bundle-98, epoch 71 | Reviewed source stays selected/readable; Cancel retains its parent source, and unchanged-source explicit Apply at narrow width remains readable. Exact captured version/replacement is asserted by focused tests. |

Owner and Party summaries are contained and readable. The long reviewed Note
source wraps in its selected summary; its native single-line option clips within
the existing control, retaining its complete accessible name. No form or shell
layout was changed. Controls below the viewport remain in the existing scrollport
and are revealed by keyboard/actions.

Axe completed on all eight captures. Advisory findings were `page-has-heading-one`
(moderate) and, where the grid was included, `empty-table-header` (minor).
`color-contrast` remained incomplete for overlapping/grid/control nodes. These are
private rendered observations, distinct from the passing product accessibility
row. Actual assistive-technology testing was unavailable and is not claimed.
Separate heading/source-label findings remain outside this task. Some exploratory
requests used absent targets (including “Title” instead of observed “Title value”
and an action after returning from Recovery); fresh snapshots and observed targets
resolved them. No product write was submitted during this review.

The exact session was stopped with `make ui-review-stop`; startup foreground
process exited successfully. Its `ui-review/terminal.json` reports `closed`,
`status=ok`, `cleanup=complete`, 8 images and no failed requests. Receipt SHA-256:
`3d1043d98de6ecf8843263847207060fa94fdcb075517fe26540cfdfd67725ea`.
An earlier startup `20260930T193233Z-p87351` failed in frontend build artifact
preparation (`invalid_artifact`, `child_failed`). Exact stop retained its original
failed outcome and a `cleanup=complete` terminal receipt; the fresh session above
succeeded. Caller-owned request scratch was removed after review.

## Inherited visual blocker

Required target: the focused service-backed visual slice above, specifically
`module.workbook.visual.contextual_task_decision_creation`.
Changed-picker run: `.cartulary/test-results/20260930T193218Z-p62624`.
Baseline-picker comparison: `.cartulary/test-results/20260930T195334Z-p53445`.
The two `frontend-visual-reconciliation.json` artifacts and the Playwright reports
were inspected, as were all six failed actual/diff pairs.

The same six committed comparisons fail with identical actual PNG SHA-256 values
under the pinned renderer in both runs: contextual Task/Decision authoring wide,
authoring narrow, and reference chooser narrow. Pixel counts respectively are
4,675 / 1,592 / 1,047 for Task and 3,146 / 101 / 6,684 for Decision. Differences
are in text/control rendering already present with the reviewed picker. There is
no observed pixel change attributable to this correction in these canonical
captures. Both runs pass ordinary authoring/recovery. Renderer attestation and
golden manifest verification pass; reconciliation reports the unsuccessful visual
attempt rather than a missing or ambiguous golden.

No capture settings, masks, normalization, tolerance, fixture, renderer or golden
were changed. A golden refresh is not justified by readable selected labels when
the actual images are unchanged. Resolving this inherited comparison requires a
separate explanation and scoped visual maintenance; the current public update
target expands to the visual catalog, so it was not used to refresh unrelated
differences. This applicable gate remains failed and prevents claiming completion.

## Digest acceptance assessment

Assessment is scoped to this correction; passing rows do not certify untouched
workbook behavior or a broader conformance profile.

| Row | Result | Evidence or scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | Owner map above; human review only; source and browser ownership separated. |
| A002 Scope | PASS | One staged-selection boundary, real presentation decision, bounded benefit/extension/risk and retirement stated; scoped production diff. |
| A003 Repository state | PASS | Clean reviewed main; current source/import/generation/routing inspection and Make guidance; HEAD unchanged. |
| A004 Tokens | PASS | No design literal, token, theme or density registry added. |
| A005 Theme | PASS | Existing dark_graphite observed in all eight manifests; theme code unchanged. |
| A006 Density | N/A | No density selection, geometry, typography or editor sizing algorithm changes. Existing compact rendering observed, not re-certified. |
| A007 Creation | PASS | Exact actor/source IDs and metadata, Apply-only parent change, no contextual submission, reviewed-source compatibility and existing consumer browser assertions. |
| A008 Responsive | N/A | Inline-size/clamp/ARIA algorithms untouched; representative 1440×900 and 1024×720 reviewed. Unchanged threshold/below-minimum/resize algorithms do not require a new sweep. |
| A009 Overflow | PASS | Eight inspected viewport images retain shell status/navigation and owned grid/Inspector/Recovery scrolling; selected summaries contained. |
| A010 Inspector | PASS | Existing contextual route/entry-point and source invalidation authoring tests pass; no dispatcher change or new route. |
| A011 Continuity | PASS | Same control identity/focus, retained selection/raw Title/source, page/revision/authority fences, cancellation and Recovery tests and rendered journeys. |
| A012 Transactions | N/A | No transaction IDs, captured requests, attempt admission, replay or dispatch ownership changed. |
| A013 Acknowledgement/recovery | N/A | This is unsubmitted-draft close/resume. Acknowledged/uncertain operation and write replay machinery are unchanged. |
| A014 Editing | PASS | Local staging, Escape/Cancel, explicit Apply, metadata/raw text retention and authority lifetime tests pass. |
| A015 Conflict | N/A | No saved-value conflict presentation, mutation response or toast behavior changed. |
| A016 Query/interaction states | PASS | Existing pending/failure/empty/authorization tests plus guarded accepted reconciliation; labels do not admit reads/writes or readiness. No producer/Grid Adapter state changes. |
| A017 Refresh/authorization | PASS | Revision invalidation, concealment, obsolete/disposed completion, suspension/account retirement tests pass in requested owner rows. |
| A018 Evidence | N/A | Evidence lifecycle, overlays and preview presentation unchanged; related Party chooser coverage is ordinary selection coverage. |
| A019 Accessibility | PASS | Required accessibility row and keyboard/focus/name assertions pass; rendered Axe advisories and unavailable actual AT explicitly separated. No new motion/color behavior. |
| A020 Components | PASS | Existing controls/keys retained; duplicate/long-name and off-page/removal tests plus inspected narrow containment. Zoom/text-spacing algorithms untouched. |
| A021 Virtualization | N/A | Grid Adapter rows, virtualization and authoritative result production unchanged; candidate/selection boundedness tested separately. |
| A022 Visual fixtures | BLOCKED | Ordinary row passes, but required contextual comparison fails identically with baseline; exact roots and acceptance gap above. No unrelated golden refresh. |
| A023 Selectors | PASS | Existing test IDs, fields, schemas and exact candidate IDs target identity; readable names asserted separately. No selector/API change requiring package.ui rerouting. |
| A024 Test authority | PASS | Tests/routing/generation consume source and machine inputs only; no Markdown dependency added. |
| A025 Generated artifacts | PASS | Authored routing precedes Make generation; only render-index input hashes change; policy and finalizer shape/catalog/drift validation pass. |
| A026 Compatibility | PASS | No API, schema, migration, cache, policy, persistence or operation-owner change. Captured-source/generic selector/Assessment/existing-record boundaries retained. |
| A027 Handoff | BLOCKED | Implementation and scoped evidence delivered, with cleanup and rollback; applicable A022 prevents completion. |

## Retirement, rollback and remaining work

Retired only ordinary authoring's stale selected-label preference when a current
authorized candidate supplies presentation. Existing selection, discovery,
reviewed-source and mutation owners remain. No compatibility adapter, feature
flag, endpoint, new request or data migration is needed.

Rollback reverts the picker correction, its tests, authored routing, generated
render-index projection and this documentation together. It neither changes saved
records nor undoes the preceding status-strip, bulk-selection, mention/focus,
Columns-continuity or surface-navigation corrections.

The separately reproduced contextual source-label pruning issue and heading
findings are unchanged. Next action is to explain and resolve the inherited
contextual visual comparison under its visual owner, then rerun the applicable
gate before claiming completion. All required label product assertions and seeded
review journeys already pass; no broad UX audit or unrelated product repair was
performed.

## Final maintenance

| Final command | Result / run root |
| --- | --- |
| `make format` | PASS, `20260930T200427Z-p90626`; no additional formatter diff. |
| `make agent-finalize` | PASS, `20260930T200442Z-p95126`; shape/catalog/tier/generation/drift validation, zero updated files. |
| `make frontend-typecheck` | PASS, `20260930T200526Z-p99616`. |
| `make lint-biome` | PASS, `20260930T200526Z-p99681`. |
| `make generated-artifact-policy-check` | PASS, `20260930T200526Z-p99364`. |
| `make frontend-import-boundary-check` | PASS, `20260930T200526Z-p99677`. |
| `make lint-markdown` | PASS, `20260930T200526Z-p99754`; summary at `adhoc/lint-markdown/tool-run-summary.json`. |
| `git diff --check`; final status/diff inspection | PASS; five authored source/test files, two authored routing inputs, one generated render index, and this handoff only. HEAD remains reviewed main. |

Retained-run maintenance was skipped because `RESULTS_DIR` was unset. The scoped
diff contains no golden, normative specification, dependency or operation-owner
change. Applicable visual acceptance remains blocked as documented above.
