# Readable retained contextual original source

Implementation date: 2026-09-30. This bounded slice implements independently
retained original-source presentation for contextual Task Request and Decision
creation. Completion remains blocked by the contextual visual comparisons below.
No commit, push, deployment, specification edit, dependency change, migration,
endpoint change, or golden update was made.

## Baseline, authority, and ownership

The checkout started clean on `main` at
`0f31a850d943c1d646888308d3d2accf265c0fa8`, “Readable retained authoring selections”.
HEAD did not advance during implementation; there was no unrelated local work.
Root `AGENTS.md` was the applicable repository agent file. Current Make help,
help-all, and module-author task guides for `web.workbook`, `module.workbook`, and
`module.entities` revalidated the requested active catalog rows. Authored source
ownership/import boundaries, generated policy, local source guides, and digest
README, START_HERE, LOCAL_AGENT_PROMPT, REPO_MAP, OWNER_MAP, QUERY_RECIPES, rules,
and acceptance material were consulted.

The reproduced defect was saved Timeline row → Inspector → Create task request →
Title → remove seeded Linked Records → scalar edit → keep/close → Recovery →
actual retained item → Resume. The source paragraph degraded to the UUID because
`update()` pruned `draft.labels` according to editable target membership. Recovery
used a separate captured origin. Focused assertions reproduced both targets before
production changes. Research, bundled prompts, and historical contextual-create
and retained-selection handoffs were supporting evidence, not authorization.
No adopted-owner conflict was found.

| Adopted/design owner | Application |
| --- | --- |
| Core 00 document status and precedence | Repair implementation/projections without changing normative owners. |
| Core 01 §7.4 contextual feature registry, target seeds, reference contracts, and capability; §§18B/19 | Preserve declared source identity, editable target references, create-writable fields, minima, explicit activation, and existing read/write routes. |
| Core 02 §§10.4.1/10.4.2 | Task/Decision fields, relationship meaning, defaults, and lifecycle remain unchanged. |
| Core 03 §16.4, REQ-03-256 | Source continuity, retained raw authoring, explicit review, detach/navigation/recovery, and readiness remain independent of presentation. |
| Core 03 REQ-03-299/100; Core 04 §§1–2 | Concealment, same-account retention, authority changes, account replacement, incident retirement, and stale completions preserve their separate lifetimes. |
| Design §§6.2, 7.3, 12, 14 | Readable primary context, Inspector authoring, authorization-sensitive labels, wrapping, keyboard/removal and recovery focus. |
| Domain | Existing Task Request, Decision, original source, reference, and Recovery vocabulary/navigation. |

Source ownership is `web.workbook`. Independent browser verification owners are
`module.workbook` for accessibility/visual presentation and `module.entities` for
the existing cross-surface contextual replay scenario. Catalog routing is not
product authority or proof of specification completeness. Tests, generators,
runtime metadata, and routing gained no dependency on Markdown.

## Boundary, implementation, and lifetime decisions

The common decision is whether the retained original source currently has a
permitted readable label. It belongs to the existing contextual owner, independently
of editable target selection. One nullable source label is exposed in its snapshot;
existing operation entries also carry nullable presentation outside immutable
captured attempts. Neither is a new cache or workflow owner. Full rows and visited
candidate pages are not retained for display.

`draft.labels` remains selected-target metadata, pruned by the unchanged shared
`retainWorkbookReferenceLabels`. One local `contextualCreateSourceLabel` projection
supplies Form, Recovery origin, and attached workflow context. It uses “Original
source needs review” when presentation is unavailable. It never falls back to
captured `draft.presentation.label` or makes the UUID the primary label. The
historical captured presentation remains inside existing immutable drafts/attempts;
its live display use is retired. Identity, versions, seeds, raw values, attachment
tokens, and captured request bytes are not rewritten by label changes.

| Event | Label decision and preserved invariants |
| --- | --- |
| Target Apply/removal, scalar edit, detach/resume, different-source navigation | Retain original source label and identity. Target removal really removes the raw target link; no forced seed retention. |
| Other selected target changes | Retain source label; existing review invalidation and target label pruning still apply. |
| Original source change/removal or unscoped invalidation | Withdraw source presentation, including operation entries. Preserve authored values and identity. First removal at an already-observed version invalidates presentation/review; duplicate removal and lower versions remain fenced. |
| Authority change/suspension | Withdraw source and target labels. Conceal owner state while authority is absent; retain same-account work according to existing policy. |
| Current successful review | Derive from `genericInspectorRowLabel` on the existing authoritative full-row read after identity, version floor, removal, generation, originating draft, and review checks. Existing explicit review owns version acceptance/readiness and minima. |
| Failed review | Never install returned presentation. Authority/capability failure before the source read retains still-permitted presentation. Unsuccessful source reads conservatively withdraw it. Existing values and review-required state remain. |
| Obsolete success/failure | No presentation/readiness changes. Failure handling now also checks originating draft and review revision. |
| Acknowledged refresh | Matching source rows already admitted by existing read-only refresh may recover entry presentation after existing authority/lifetime/version/removal fences. Absent rows cause no traversal or additional read. |
| Account/incident replacement | Existing retirement clears drafts, entries, and protected presentation; obsolete work cannot restore them. |

Label-only review updates preserve draft object identity. This is necessary because
submission checks its captured draft reference. Accepted version changes still use
the existing explicit source-version update. Label changes cannot grant authority,
accept a version, satisfy target minima, clear `needsReview`, dispatch, capture a
new transaction, or replay a write.

Form source text wraps inside its existing scroll owner. Rendered inspection found
that Recovery's shared selected-origin paragraph omitted the wrapping already used
by its list items/detail host. The only shared presentation change applies that
same `overflowWrap: anywhere` rule to that paragraph. It changes no registration,
navigation, focus, scrolling, or owner semantics. A browser geometry assertion
first reproduced the clipping and now covers the contextual long-token origin.

The benefit is consistent readable original context without weakening review or
authorization. A later declared contextual source can use the same bounded owner
projection; no speculative surface was implemented. Leaving the coupling would
continue degrading identification after legitimate target edits; falling back to
captured text would restore protected labels after invalidation. The remaining
delivery risk is the unresolved visual gate.

### Changed authored files and generated projection

Paths below are relative to the repository root.

- `apps/web/src/workbook/features/coordination/WorkbookContextualTaskDecisionCreateOwner.ts`: source-label lifetime, guarded review recovery, operation-entry projection, removal invalidation and existing refresh integration.
- `apps/web/src/workbook/features/coordination/contextualCreateOperation.ts`: nullable entry presentation outside captured attempt.
- `apps/web/src/workbook/features/coordination/contextualCreateModel.ts`: one safe local label projection.
- `apps/web/src/workbook/features/coordination/ContextualCreateForm.tsx`, `ContextualCreateRecovery.tsx`, `useContextualCreateAttachment.ts`: consume safe source presentation.
- `apps/web/src/workbook/components/WorkbookRecoveryPanel.tsx`: wrap selected-origin text in the existing scrollport.
- `apps/web/src/workbook/features/coordination/contextualCreateAuthoring.test.tsx`, `contextualCreateRecovery.test.tsx`: focused source/target continuity, review, concealment, stale completions, operation presentation, and unchanged mutation assertions.
- `apps/web/e2e/support/workbook/contextualCreate.ts`: existing isolated fixture optionally uses Timeline; keyboard removal, exact scalar retention, actual Recovery selection/resume and zero target creation requests.
- `apps/web/e2e/workbook.a11y.spec.ts`, `workbook.visual.spec.ts`: extend existing scenarios with Timeline Task/Decision long-label journeys. Existing Evidence golden identities/settings remain intact; new Timeline captures are diagnostic attachments.
- `tools/test_families/web.workbook.json`: route six new literal unit titles. Existing browser titles/scenario IDs are unchanged, so browser family routing inputs need no new rows.
- `apps/web/src/workbook/features/coordination/README.md` and this handoff: human documentation only.
- `tools/execution_topology_render_index.json`: Make-generated downstream index; no generated output was hand-edited.

Inspected compatibility boundaries include `ContextualReferenceControl`, the
ordinary `WorkbookAuthoringReferencePicker`, generic Inspector labels, authoring
reader/adapters/ports, operation capture/transport, and retained reference pruning.
The preceding ordinary-picker reconciliation and Timeline fixes remain intact.
Captured-source pickers, Coordination, Notes, Evidence, Assessments, and ordinary
record editing gained no new owner, read path, or mutation semantics. The shared
Recovery origin style uses its existing wrapping policy, including for those
owners' long origins. No compatibility alias or migration is required.

## Characterization and product evidence

All commands ran from the repository root through public Make targets. Run IDs
below resolve under `.cartulary/test-results/`; graph units are not test counts.

The valid pre-production RED run is `20260930T225931Z-p28574`: both Task/Decision
journeys lost their readable source after pruning; review displayed a UUID;
suspension exposed old captured Recovery origin; obsolete failed review changed
the current snapshot. Exact failures are in
`unit-logs/row-web.workbook.regression.contextual_task_decision_authoring/vitest-failure-details.json`.
The first implementation GREEN is `20260930T230108Z-p29646`; expanded operation
coverage GREEN is `20260930T230530Z-p39623`.

Additional RED `20260930T230927Z-p41918` reproduced retained operation text after
a source removal at the known version. GREEN `20260930T231058Z-p75577` covers the
removal fix. Browser RED `20260930T232247Z-p2501` failed the new Recovery origin
`scrollWidth <= clientWidth` assertion, independently characterizing long-token
clipping before the wrapping change.

Coverage includes both targets, removing source from every target field while
retaining other references, subsequent exact raw scalar edits, actual Recovery
selection/resume, different-source navigation, bounded target labels and zero
creation requests. Separate observations/read variants cover other-target versus
source invalidation, incomplete/missing/wrong/stale/changed rows, failed/current
review, target minima, candidate-label reconciliation, obsolete successes and
failures, uncertainty/suspension, role changes, same-account recovery, account
replacement and incident retirement. Existing tests still protect settlement of
prior source writes, high-water marks, synchronous activation reservation,
immutable attempts, exact uncertain replay, late receipts, and acknowledged-write
refresh with no resend.

| Exact command | Outcome / run root |
| --- | --- |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.contextual_task_decision_authoring,web.workbook.regression.contextual_task_decision_recovery CARTULARY_HARNESS_CACHE_MODE=off` | Final PASS `20260930T232822Z-p48097` (3/3 units); earlier PASS `20260930T231058Z-p75577`. |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.contextual_task_decision_creation CARTULARY_HARNESS_CACHE_MODE=off` | Final wrapping GREEN `20260930T232822Z-p48124` (11/11); earlier PASS `20260930T231058Z-p75648`, then overflow assertion RED `20260930T232247Z-p2501`. |
| `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.contextual_task_replay CARTULARY_HARNESS_CACHE_MODE=off` | PASS `20260930T231058Z-p75649` (11/11), after final owner changes; later edits only affect wrapping/test geometry. |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.visual.contextual_task_decision_creation CARTULARY_HARNESS_CACHE_MODE=off` | Final FAIL `20260930T232822Z-p48147` (9/11); exactly six inherited comparison errors. All twelve actual/diff PNGs remain byte-identical to baseline, as in the first implementation run `20260930T231058Z-p75594`. All four new Timeline diagnostic attachments completed. |
| `make format` | PASS `20260930T230404Z-p31632`, `20260930T230957Z-p42903`, `20260930T232234Z-p97625`, `20260930T232415Z-p35247`; no unrelated authored edits. |
| `make generate` | Final PASS `20260930T232441Z-p39797`; earlier PASS `20260930T225651Z-p18148`, `20260930T225808Z-p22232`, `20260930T225919Z-p25567`, `20260930T230514Z-p36518`, `20260930T231005Z-p47496`. |
| `make agent-finalize` | Final PASS `20260930T232748Z-p43445` before final broader checks; earlier PASS `20260930T231021Z-p50393`. `RESULTS_DIR` unset: retained-run maintenance skipped; no legitimate successful full warm check exists. |
| `make frontend-typecheck` | Final PASS `20260930T232822Z-p48509`; earlier PASS `20260930T231058Z-p75774`. |
| `make lint-biome` | Final PASS `20260930T232822Z-p48625` after warning correction. |
| `make frontend-import-boundary-check` | Final PASS `20260930T232822Z-p48572`; earlier PASS `20260930T231120Z-p55911`. |
| `make generated-artifact-policy-check` | Final PASS `20260930T232822Z-p48086`; earlier PASS `20260930T231120Z-p55503`. |
| `make generate-drift` | Final PASS `20260930T232822Z-p48102`; earlier PASS `20260930T231120Z-p55509`. |
| `make lint-markdown` | Final PASS `20260930T233247Z-p24400` after evidence update; earlier PASS `20260930T232822Z-p48819`. Documentation maintenance only. |

Other intermediate failures were authored-fixture/tooling issues, not defect RED
evidence: `make generate` at `20260930T225614Z-p14982` rejected dynamically templated
test titles; literal titles repaired routing. The unit run
`20260930T225701Z-p21158` expected an unseeded actor label; corrected to the actual
selected-only label set. `make frontend-typecheck` at `20260930T230530Z-p39699`
rejected the new fixture's inferred Evidence-only literal type; an explicit
Evidence/Timeline union repaired it. `make lint-biome` at
`20260930T231120Z-p55958` rejected four new non-null assertions; explicit nullable
bounds handling repaired them. Final lint passes as recorded in the table.

### Final verification

Final source checks pass except for the exact contextual visual gate. `git diff
--check` passes. Final branch remains `main` at the baseline commit, with only this
slice's fifteen tracked changed files and this new handoff; no unrelated edits,
normative specs, goldens or lockfiles changed. Final Markdown lint passes as
recorded in the table.

## Rendered review and visual blocker

The supplied baseline visual run `20260930T221344Z-p63275` matched reviewed HEAD and
rendering inputs. Before implementation, all six actual/diff pairs were inspected
with `cartulary-ui-review` in artifact mode at `20260930T223301Z-p5440`. Terminal
receipt: `closed`, `status=ok`, `cleanup=complete`. No new product pass was inferred.

Current seeded review used `make ui-review UI_MODE=seeded` at
`20260930T231029Z-p50889`. Editor actions reproduced Timeline Task and Decision
source-reference removal, scalar editing, keep/close, actual Recovery selection,
and Enter resume. Task removal after explicit focus returned focus to Choose Linked
Records. Decision used targeted Enter removal of Support Refs and Affected Records;
its removal-focus guarantee comes from the product browser assertions, not an
inferred focus observation. Wide/narrow resumed screenshots showed readable source
context and focused retained scalar values. No contextual creation was activated.

Page captures at 1280×720 and 768×640 completed Axe 4.13.0 scans: advisory
`empty-table-header` (minor), `page-has-heading-one` (moderate), and incomplete
`color-contrast`; these shell/grid observations are not a contextual accessibility
product failure or a complete accessibility audit. Console/failed-request counts
in the session terminal receipt were zero.

Four diagnostic Timeline long-label PNGs from the current visual scenario were
imported by exact attachment identity and inspected through the skill. Original
bundle byte counts and SHA-256 were verified before opening. They established
readable Form wrapping and the Recovery origin clipping subsequently fixed and
characterized above. Standalone PNG imports provide no DOM/Axe/canonical comparison
claim. Canonical import of the new run returned `unsafe_artifact` because its
9,758,440-byte trace exceeds the tool's 8 MiB component limit. No metadata or limits
were modified. Separately comparing each exact report-associated current actual
and diff file against baseline established byte identity for all twelve PNGs.

`make ui-review-stop` used the exact seeded session locator; the original foreground
process exited. Terminal receipt: `closed`, `status=ok`, `cleanup=complete`, eight
images. Only owned request/diagnostic scratch was removed. Private observations,
credentials and images are not committed.

Final wrapping inspection used artifact mode at `20260930T232920Z-p21370`, importing
the exact four new Timeline PNG attachments from final visual run
`20260930T232822Z-p48147`. Bundle byte counts/digests were verified before viewing.
Task and Decision original context wraps completely in Form and Recovery origin
at both widths; controls remain in the existing vertical scrollport. The diagnostic
anchor favors source context, so a lower input can be partly below its viewport;
the independent accessibility scenario scrolls/focuses controls and verifies
reachability at 1280, 768 and 390 pixels. These PNGs have no DOM or Axe channel.
The exact artifact session was stopped; foreground exited and terminal receipt
confirmed `closed`, `status=ok`, `cleanup=complete`, four images. Owned scratch was
removed.

| Capture | Baseline and first implementation differing pixels |
| --- | ---: |
| contextual-task_request-authoring | 4675 |
| contextual-task_request-authoring-narrow | 1592 |
| contextual-task_request-references-narrow | 1047 |
| contextual-decision-authoring | 3146 |
| contextual-decision-authoring-narrow | 101 |
| contextual-decision-references-narrow | 6684 |

The exact unresolved gate is
`module.workbook.visual.contextual_task_decision_creation`. Reconciliation at
`20260930T232822Z-p48147/browser-e2e-visual/frontend-visual-reconciliation.json`
accounts for ten active contextual captures, no missing golden or ambiguous
mapping, and the six comparison failures. The partial selection's other goldens
are blocked from mutation, not deletion candidates. The reviewed differences
include text/control outlines and predate this slice; their golden-maintenance
trigger is not established here. No tolerances, masks, anchors, identities, goldens,
or manifests were changed to hide them.

The public `browser-e2e-visual-update` explanation selects 44 rows and exposes no
owner/row-scoped input. It was not run. Follow the visual golden maintenance guide:
the smallest next action is an explicitly supported scoped update path for the
affected contextual captures, with a justified trigger and ordinary validation.
This slice must not claim full completion while that gate fails.

## Acceptance assessment

The following assesses the applicable rows of the digest's `acceptance.tsv` for
this bounded slice; unrelated owner obligations are not re-audited.

| Row | Status | Evidence / bounded rationale |
| --- | --- | --- |
| A001 Authority | PASS | Owner map above; no normative edits; independent source/catalog owners. |
| A002 Scope | PASS | One label-lifetime decision in existing owner; coupling/display fallbacks retired; compatibility and extension rationale above. |
| A003 Repository state | PASS | Clean baseline, unchanged HEAD, applicable instructions/guides and current authored boundaries inspected. No direct grid-vendor imports. |
| A004 Tokens | PASS | No token/theme/density registry or design literal introduced; only existing wrapping behavior. |
| A005 Theme | N/A | No theme behavior changed; existing dark_graphite browser fixtures retained. |
| A006 Density | N/A | No density selection/geometry contract changed. |
| A007 Creation | PASS | Both target authoring/recovery unit slice; exact values/source/minima/no-write assertions; unchanged capability/route/payload boundaries. |
| A008 Responsive | PASS | Final accessibility GREEN: existing authoring journey plus 1280×720, 768×640, 390×480 retained long context and control/focus checks; no chrome thresholds changed. |
| A009 Overflow | PASS | Final accessibility GREEN and inspected corrected screenshots: existing Inspector/Recovery scroll ownership; long source Form and origin assertions. |
| A010 Inspector | PASS | Existing semantic actions, disabled authority state, review invalidation, detach/resume and late completion coverage. |
| A011 Continuity | PASS | Both targets, every source target field removed, other references retained, exact scalar edits, different-source navigation and actual Recovery resume. |
| A012 Transactions | PASS | Owner unit slice and module.entities contextual replay: captured identity/body, duplicate activation, settlement, dispatch and late receipt fences preserved. |
| A013 Acknowledgement/recovery | PASS | Exact replay and acknowledged-write/failed-refresh unit/browser evidence; source presentation refresh reads add no resend. Queue retry/discard owner untouched. |
| A014 Editing | PASS | Raw field retention, explicit Apply/Cancel, removal focus and keyboard browser journeys; no persistence added. |
| A015 Conflict | N/A | No cell conflict presentation or conflict resolution behavior changed. |
| A016 Data/interaction states | PASS | Missing/incomplete/wrong/stale source and authority/suspension/role scenarios remain independent of minima/readiness; no query state owner changed. |
| A017 Authorization scope | PASS | Concealment, role changes, same-account retention/review, account/incident retirement and obsolete completions; no captured-text fallback. |
| A018 Evidence | N/A | Evidence lifecycle/preview/overlay unchanged; existing Evidence fixture preserved. |
| A019 Accessibility | PASS | Final owner-selected accessibility GREEN and manual keyboard/readable-context observations; advisory Axe limitations stated separately. |
| A020 Components | PASS | Final geometry/focus GREEN: unavailable/current states and long/unbroken content at supported/narrow/below-minimum widths. No re-theme or form rearrangement. |
| A021 Virtualization | N/A | No grid adapter, virtual row identity, producer, or performance boundary change. |
| A022 Visual fixtures | BLOCKED | Current contextual visual gate fails six inherited comparisons; diagnostic/manual inspection is not a product pass. |
| A023 Selectors | PASS | Existing semantic roles, field IDs, view/record IDs and actual Recovery navigation; no new selector contract. package.ui unchanged. |
| A024 Test authority | PASS | Authored tests/routing/generation do not read or depend on Markdown. |
| A025 Generated artifacts | PASS | Authored unit-title routing before Make generation; final generated policy/drift passes. |
| A026 Compatibility | PASS | No policy, schema, persistence, request or replay redesign; shared pruning and preceding picker/Timeline fixes retained. No migration. |
| A027 Handoff | BLOCKED | This evidence-backed record preserves the unresolved A022 exit; completion cannot be certified until its gate passes. |

## Rollback and next action

Revert only this slice's authored files/routing and corresponding Make-generated
index projection. For the shared Recovery panel, revert only the selected-origin
wrapping line. Preserve preceding picker reconciliation and Timeline fixes. No
data migration, endpoint, dependency, lockfile, or persistent-draft cleanup is
needed. No unrelated source or scratch should be removed.

Picker/discovery rows were not added because those shared implementations/read
ports did not change. Broad backend, release, benchmark and full warm suites were
not run: this slice changes frontend retained presentation and focused browser
evidence, not backend/release boundaries. Golden update was deliberately skipped
because the public target cannot select this bounded scope. `RESULTS_DIR` remained
unset; retained-run maintenance was skipped rather than misusing narrow evidence.

Resolve the exact contextual visual gate through supported scoped maintenance and
its required reviewed trigger/validation. Keep implementation and passing product
evidence separate from that blocked exit.
