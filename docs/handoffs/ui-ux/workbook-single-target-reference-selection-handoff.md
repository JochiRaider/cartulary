# Readable, identity-aware single-target reference selection

## Boundary and authority

Baseline: clean `main`, HEAD and local `origin/main` both
`9ab6a79a530469e0044edd21325fd2c8e87533cd`. Only root `AGENTS.md` applied.
Work remains uncommitted on that branch; no publishing, merging or deployment.
The maintained digest, current source guides, generated-artifact policy and
verification catalogs were inspected. Historical prompts and handoffs supplied
navigation/evidence, not implementation authority. No localization changes were
needed for this bounded seam.

The governing owners are Core 01 §§3.3.4/3.3.7 (query and paging), §18B (direct
references), applicable Party/subject contracts; Core 03 §13 (keyboard/editing),
§16.2 (direct references), §16.4 (retained contextual authoring), REQ-03-267–274
(Party wording/link separation and continuity), and authorization retirement.
Design §§7–8/14 and §15.2A supply compact, accessible long/equal-label direction.
Domain vocabulary and owner navigation remain unchanged. No specification was
modified and no executable artifact depends on Markdown.

The common presentation decision is inspecting one candidate identity before
acceptance. The native select hid equal-label identity and clipped long text in
its closed control, while the separate Note summary already wrapped. No wrong-link
incident, frequency measurement or work loss is claimed. The replacement exposes
wrapping labels and stable IDs without introducing a new state owner or interaction
framework. This makes existing single-target consumers distinguishable; future
consumers can use the same scalar boundary without speculative capabilities.

## Implementation and consumer inventory

Paths in this section are relative to `apps/web/src/workbook/`.

- `components/WorkbookRecordCandidatePicker.tsx`: bounded native radio list,
  unique native group name, full wrapping label/ID, visible focus tokens, disabled
  candidates, no default choice, and Enter suppression within parent forms.
- Exactly two direct production callers: `components/WorkbookCandidateSelection.tsx`
  and `features/parties/PartyLinkControls.tsx`. The former preserves the multiple
  branch and retains owner-supplied metadata independently of page membership.
- Transitive consumers: `WorkbookAuthoringReferencePicker`, noncompact
  `WorkbookAuthoringReferenceControl`, `NoteSourceControl`, `AssessmentDiscovery`,
  `ContextualReferenceControl`, and `CoordinationCreateForm`. Single selected
  summaries expose full IDs after acceptance; compact grid triggers keep their
  geometry and expose identity when opened.
- `WorkbookCandidateQueryControl` accepts the Assessment disclosure focus ref.
  `useSelectedReferenceRemovalFocus` remains the focus owner without code changes:
  obsolete intentions, external focus protection and local reveal are preserved.
  Staged removal uses the first enabled radio, then authoring Cancel, Assessment
  ordering/filter disclosure, or Party loaded filter. These new empty-page fallbacks apply only to single
  selection; the multiple-selection fallback is unchanged. Parent removal retains
  its Choose trigger.
- Party adds Clear selected party, which only clears the staged target. Link
  existing party still performs acceptance. Existing Clear party link/text/both
  payloads and source wording are unchanged. The current link also exposes its ID
  and wraps unbroken labels inside the inspector.
- Full IDs and labels are concealed together when the existing authority requires
  concealment, including accessible names and fallback summaries.

The retired select API (`selectedRecordIds`, `onSelectedRecordIdsChange` array callback and
`selectorRef: Ref<HTMLSelectElement>`) has no compatibility alias. Its replacement is
`selectedRecordId: string | null`, `onSelect(recordId)` and an optional
`focusTargetRef` for an input. Shared browser helpers support radios while retaining
actual select consumers. The independent ordinary single direct-reference branch,
specialized supersession selectors, enum/query selects and multiple picker remain.

Queries, eligibility, maximum counts, authoring Apply/Cancel, Assessment draft
updates, explicit Party Link, mutation payloads and draft lifetime stay with their
owners. The shared discovery service still owns one current page and ten checkpoints;
Party discovery still accumulates pages. No detail/enrichment reads, cache, store,
API, schema, dependency or migration was added. Re-selecting a retained candidate
uses its already-reviewed metadata. Note source ID, schema, row version and label
remain associated through paging, surface change and cancellation.

Tests and affected browser helpers were migrated in the existing authored files.
New titles extend existing `web.workbook` and `module.workbook` rows; no new row ID
was invented. `make generate` refreshed only the topology input index for those
catalogs. Components and the four affected feature source guides were updated.

## Verification ledger

All commands ran from `/home/jochi/code/cartulary` through public Make targets.
Run roots below are relative to `.cartulary/test-results/`. Fresh named runs used
`CARTULARY_TEST_RUN_ID=<root> CARTULARY_OUTPUT_MODE=machine` before the shown Make
command. Outer exits are recorded separately from harness status. These are focused
implementation checks, not a full-owner, release, golden or conformance claim.

Discovery: `make help`, `make help-all`, `make task-guide ROLE=module-author
OWNER=web.workbook`, the corresponding `module.workbook`, `module.parties`,
`module.assessments`, `web.design` and `module.links` guides, and relevant `explain-target`/
`explain-test-owner` reads. Discovery commands did not use global machine mode.

The pre-production characterization commands were:

```sh
# single-reference-characterization-20261009a
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.party_link_controls,web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.assessment_discovery,web.workbook.regression.note_create_authoring,web.workbook.regression.contextual_task_decision_authoring
# single-reference-characterization-20261009b and c
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.contextual_task_decision_authoring
```

The full focused regression command is:

```sh
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.party_link_controls,web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.assessment_discovery,web.workbook.regression.reference_controls,web.workbook.regression.authoring_candidate_discovery,web.workbook.regression.note_create_authoring,web.workbook.regression.contextual_task_decision_authoring,web.workbook.regression.coordination_create_authoring,web.workbook.regression.workbookshell_assessments_suite_f2e4a841a2
```

| Run root | Command/selection | Outer exit and result |
| --- | --- | --- |
| `single-reference-characterization-20261009a` | `make test-slice OWNER=web.workbook ROWS=` Party controls, candidate presentation, Assessment discovery, Note create, contextual Task/Decision authoring rows | 2; contextual synthetic fixture omitted the Party surface. Other four rows passed before production changes. |
| `single-reference-characterization-20261009b` | Same contextual row only | 2; fixture still omitted Party surface. |
| `single-reference-characterization-20261009c` | Same contextual row only, corrected surface inventory/raw field key | 0; pass before production changes. |
| `single-reference-regression-20261009a` | Full focused regression command above | 2; three migrated-test failures: obsolete option lookups in contextual/shell tests and an unstable Assessment test reader. Corrected test fixtures. |
| `single-reference-regression-20261009b` | Full focused regression command | 2 before run publication; catalog title/scenario cardinality was temporarily incomplete. Corrected authored registration. |
| `single-reference-regression-20261009c` | Full focused regression command | 0; all nine rows passed (10 harness units including install). |
| `single-reference-shell-20261009a` | `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.recovery_production_shell` | 0; pass. |
| `single-reference-final-regression-20261009a` | `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.workbookshell_assessments_suite_f2e4a841a2` | 0; pass after all-disabled and exact-radio assertion updates. |
| `20261009T133021Z-p58796`, `20261009T133659Z-p67809`, `20261009T134008Z-p6456`, `20261009T134141Z-p27418`, `20261009T134309Z-p94592` | `make format` | 0 each; authored frontend diffs inspected. One intervening attempt exited 2 before publication while catalog cardinality was incomplete. |
| `single-reference-finalize-20261009a` | `make agent-finalize` | 2; schema check reported stale generated topology inputs. |
| `20261009T133821Z-p74929` | `make json-shape-check` | 2; same stale catalog projection diagnostic, inspected via `make explain-run RESULTS_DIR=.cartulary/test-results/20261009T133821Z-p74929 TARGET=json-shape-check DETAIL=logs`. |
| `20261009T133837Z-p75771` | `make generate` | 0; generated topology index refreshed. |
| `single-reference-finalize-20261009b` | `make agent-finalize` | 0; schema/catalog validation and generated refresh/drift passed. Retained-run maintenance skipped because RESULTS_DIR was unset; no appropriate full warm check was supplied. |
| `single-reference-frontend-20261009a` | `make frontend-typecheck frontend-import-boundary-check lint-biome` | 2 overall: typecheck passed, second target refused reuse of the nonempty explicit run root. Remaining targets run separately below. |
| `single-reference-imports-20261009a` | `make frontend-import-boundary-check` | 0; pass. |
| `single-reference-build-20261009a` | `make build-web` | 0; pass while diagnosing private review startup. |
| `single-reference-biome-20261009a` | `make lint-biome` | 2; two new non-null assertions in shell tests. Replaced with exact accessible radio lookups. |
| `single-reference-biome-20261009b` | `make lint-biome` | 0; pass. |

Browser commands (same run prefix convention as above):

```sh
# single-reference-browser-workbook-20261009a: outer 2, Note row passed;
# Assessment off-page Tab assertion and Task stored-wording assertion failed.
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.authoring_candidates_assessments,module.workbook.browser.authoring_candidates_note_source,module.workbook.browser.authoring_candidates_parties
# single-reference-browser-workbook-20261009b: outer 0, both rows passed.
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.authoring_candidates_assessments,module.workbook.browser.authoring_candidates_parties
# single-reference-browser-parties-20261009a: outer 0, both rows passed.
make service-backed-test-slice OWNER=module.parties ROWS=module.parties.browser.source_link_recovery,module.parties.browser.the_browser_workbook_surface_keeps_ordinary_evid_687c58e86e
# single-reference-browser-assessments-20261009a: outer 2;
# both workbook-assessments rows passed, recovery/sentinel still asserted select values.
make service-backed-test-slice OWNER=module.assessments ROWS=module.assessments.browser.append_recovery,module.assessments.browser.the_browser_appends_a_subject_only_assessment_fo_2a82260393,module.assessments.browser.the_browser_workbook_assessment_workflow_appends_aefa86174b,module.assessments.browser.the_browser_workbook_drives_the_assessments_surf_4f277414a9
# single-reference-browser-assessments-20261009b: outer 0, both corrected rows passed.
make service-backed-test-slice OWNER=module.assessments ROWS=module.assessments.browser.append_recovery,module.assessments.browser.the_browser_workbook_assessment_workflow_appends_aefa86174b
# single-reference-browser-note-20261009a: outer 0, passed.
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.note_create_source_edit
# single-reference-browser-links-20261009a: outer 0, passed.
make service-backed-test-slice OWNER=module.links ROWS=module.links.browser.the_notes_tab_supports_browser_visible_creation_ae6bdcc2da
# single-reference-accessibility-20261009a: outer 2, helper focused retired select wrapper.
# single-reference-accessibility-20261009b: outer 0, passed after radio target correction.
make service-backed-test-slice OWNER=web.design ROWS=web.design.accessibility.assessment_authoring
# single-reference-browser-authoring-20261009a: outer 0, both rows passed;
# selected because both browser helpers migrated their actual reference interactions.
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.coordination_create_matrix,module.workbook.browser.ordinary_create_matrix
# single-reference-browser-wrap-20261009a: outer 0, passed after adding
# an unbroken-label fixture and saved Party-link overflow assertion.
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.authoring_candidates_parties
# single-reference-accessibility-workbook-20261009a: outer 0, all three passed.
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.timeline_related_evidence,module.workbook.accessibility.note_create_authoring_recovery,module.workbook.accessibility.ordinary_create_authoring_recovery
# single-reference-accessibility-contextual-20261009a: outer 0, all three passed.
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.contextual_task_decision_creation,module.workbook.accessibility.coordination_create_authoring_recovery,module.workbook.browser.coordination_create_retained_source
```

The failed candidate browser assertions were updated for native radio behavior:
an off-page unchecked group exits with Tab from its last radio. The new Task case
keeps literal whitespace through Cancel/Apply and checks the existing server
normalization only after explicit Create. Existing-Party linking preserves that
stored wording. Recovery assertions now inspect checked radios rather than asking
a fieldset for a select value. These corrections did not change production behavior.

The scenarios verify equal labels/full IDs, pointer and keyboard choice, no implicit
choice or Enter submission, exact accepted IDs, no writes during staging or staged
clear, retained off-page labels/metadata, explicit append/link/create, stale-response
fencing and authority concealment. Party continuation/retry still uses its accumulated
pages. Assessment and Party recovery distinguish accepted writes from failed refresh.
Source replacement tests retain ID/schema/reviewed row version and raw fields.

Additional final checks:

- `make format` also passed (outer 0) at `20261009T135635Z-p20053`,
  `20261009T140003Z-p9600`, `20261009T140120Z-p15495`, `20261009T140142Z-p20711`,
  `20261009T140629Z-p66762`, `20261009T140801Z-p12998`,
  `20261009T140819Z-p18326`, `20261009T140840Z-p23555`, and
  `20261009T141040Z-p71430`.
  Each resulting authored diff was inspected.
- `single-reference-typecheck-20261009b`: `make frontend-typecheck`, outer 2,
  two test-only unsupported Testing Library `exact` properties. Removing them
  retains exact string role-name matching. Runs `single-reference-typecheck-20261009c`,
  `single-reference-typecheck-20261009d` and `single-reference-typecheck-20261009e`
  passed, outer 0.
- `single-reference-biome-20261009c`, `single-reference-biome-20261009d`, and
  `single-reference-biome-20261009e`: `make lint-biome`, outer 0, pass.
- Final post-helper checks use `single-reference-typecheck-20261009f`
  (`make frontend-typecheck`) and `single-reference-biome-20261009f`
  (`make lint-biome`); both passed, outer 0.
- `single-reference-markdown-20261009a`: `make lint-markdown`, outer 0, pass;
  final documentation check `single-reference-markdown-20261009b` uses the same
  command (outer 0, pass).

Product summaries are at each exact run's `run-summary.json`, `target-summaries/`,
`rows/` and the referenced runner reports; initial failures are retained, not erased
by later passes. The visual scenario's retired select readiness locator was migrated
alongside the keyboard scenarios. Canonical visual/golden suites were not run or
refreshed; private rendered review supplies the scoped visual evidence.

## Rendered review

Private seeded review is separate from automated product and accessibility results.
The first startup failed building the frontend; `make build-web` then passed. The
second startup rejected a changing source snapshot while test edits were still
being formatted. Both exact sessions were stopped; terminal receipts recorded
`cleanup: complete` and their foreground processes exited 2, preserving the primary
failure. No private observations or images were obtained from these attempts.
The final sealed-source seeded session completed, with eight images inspected after
verifying each bundle component's byte length and SHA-256. No goldens were refreshed.
Its safe run root is `20261009T134336Z-p1153`; startup and exact stop both exited 0,
and `ui-review/terminal.json` recorded `state: closed`, `cleanup: complete`.
The two failed startup roots were `20261009T134010Z-p8689` and
`20261009T134048Z-p57888`, with cleanup complete and primary failures retained.

Commands were `make ui-review UI_MODE=seeded`, and serial
`CARTULARY_OUTPUT_MODE=machine make ui-review-status`, `ui-browser`, `ui-capture`
and `ui-review-stop` with each exact printed `UI_SESSION` and caller-owned
`UI_REQUEST`. Scratch requests were mode 0600 inside a 0700 directory and removed
after cleanup. Private screenshot paths, raw observations and synthetic IDs are
intentionally not committed. Production files were unchanged throughout that
inspection. Its source seal predates the final saved Party-link unbroken-label
wrapping correction; that correction was verified by the additional browser
assertion and image review described below.

Observed states:

- Timeline contextual Task requester, reopened through its retained Recovery owner:
  two synthetic Parties with identical wording showed different full IDs before
  selection. Space and arrow keys selected exactly one target. Escape cancelled
  staging and returned to Choose; reopening and Apply exposed the chosen ID in
  the parent summary. A duplicate Party was created through the ordinary UI in
  the isolated seeded incident solely for this fixture.
- Equal-label choices and summaries were inspected at 1440×900, 1024×768 and
  768×640, with an additional 125% zoom capture. Native keyboard navigation
  revealed the focused radio within the inspector. A resize alone can leave the
  previously focused item below the visible inspector portion until it is revealed;
  keyboard navigation restored the visible ring. No horizontal chooser overflow
  was observed; grid and inspector ownership remained separate.
- Note source: a complete multi-sentence label and ID wrapped inside the bounded
  chooser at all three sizes. The selected summary retained the full label and ID;
  Apply exposed them in the parent source summary. The chooser's internal scrolling
  preserved access to all 66 authorized candidates without a tall unbounded list.
- Existing-Party link presentation: both equal-label identities and the staged
  target were readable. Clearing by keyboard left a usable candidate (Space then
  selected the first identity). Arrow navigation chose the other identity; explicit
  Link displayed that exact full ID in the saved link. Staged clear left the saved
  source/link state unchanged. Product tests separately assert nonempty source
  wording preservation and exact mutation payloads.

An additional artifacts-mode review imported the exact
`single-target-linked-long-identity` screenshot attachment from the passing
`single-reference-browser-wrap-20261009a` report. The bundle's original image was
verified by byte length and SHA-256 and inspected at 1440×900. The saved link's
complete long label, unbroken segment and full ID wrap inside the inspector, as do
the bounded radio candidates. This standalone image supplies no DOM or axe evidence.
Commands were `make ui-review UI_MODE=artifacts`, `make ui-capture` with the exact
printed session/request, and `make ui-review-stop` with that same session. Safe run
root `20261009T140642Z-p72052` recorded outer exits 0, terminal `state: closed`,
`cleanup: complete`, and one inspected image. Caller-owned scratch was removed.

Axe completed on all eight captures. Seven reported no violations; the 1440×900
empty Notes grid capture reported one minor `empty-table-header` on the unchanged
grid action header. All had an incomplete contrast result, mostly for workbook
chrome/grid content. These are advisory observations, not a conformance or global
contrast claim. No screen-reader session or cross-engine review was performed.
Several review actions needed a fresh snapshot and a more precise target (retained
Recovery opening and duplicate Add row names); no uncertain mutation was replayed.

The final diff review limited the newly supplied empty-page focus fallback to
single selection, retaining the multiple branch's existing fallback. Focused
follow-up `single-reference-final-regression-20261009b` passed both rows, outer 0:

```sh
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.assessment_discovery
```

The last `make frontend-typecheck` and `make lint-biome` passed, outer 0, at
`single-reference-typecheck-20261009g` and `single-reference-biome-20261009g`.
`git diff --check` passed, and final working-tree review found only the described
slice. HEAD remains the stated baseline on `main`. No unrelated work was removed.

## Acceptance and rollback

The acceptance dispositions below apply only to this single-target seam, not the
whole workbook. Required scoped checks are complete; no blocked check remains.

| Digest rows | Disposition and evidence |
| --- | --- |
| A001–A003 | PASS: adopted clause/source/verification maps above; clean baseline and current ownership checked; direct callers migrated together. Scope encapsulates the existing single-choice presentation decision. |
| A004–A005 | PASS: existing typography, spacing, input/focus tokens and chooser bound; reviewed current graphite renderer; no theme/density registry. |
| A006 | N/A: no density selection, row height or grid-editor geometry changes. |
| A007 | PASS within scope: deliberate requester/source/subject targets and explicit create/link/append; capability and source continuity unit/browser checks. |
| A008–A009 | PASS within scope: 1440×900, 1024×768, 768×640 and enlarged-text captures; bounded chooser and owned inspector scroll; existing chrome thresholds unchanged. |
| A010–A011 | PASS within scope: obsolete scope/authority reads fenced, retained drafts/source metadata and off-page identity preserved, Cancel/Escape/focus and recovery scenarios pass. Feature dispatch is untouched. |
| A012 | N/A to implementation change: no transaction identity/replay construction changed. Existing affected recovery rows nevertheless verify captured replay behavior. |
| A013–A014 | PASS within scope: explicit acceptance/no selection writes; Party/Assessment accepted-write versus refresh recovery, raw Cancel/Apply, and native keyboard paths verified. |
| A015 | N/A: no conflict presentation changes; existing Party conflict recovery remains covered. |
| A016–A017 | PASS within scope: empty/disabled/loading/failure/retained-page controls, shared stale updates and authorization concealment checked at existing boundaries. No new permission policy. |
| A018 | N/A: Evidence lifecycle/preview presentation untouched. |
| A019 | PASS within scope: routed Assessment, Note, Timeline Evidence, contextual, coordination and ordinary authoring accessibility checks, keyboard/removal/Enter/Escape product checks and private focus observations passed. Private axe limitations are listed separately. |
| A020 | PASS within scope: equal/long/unbroken labels, full stable IDs, independent radio groups, empty/all-disabled input, viewport/zoom wrapping and bounded scrolling. |
| A021 | N/A: grid virtualization and performance model untouched. |
| A022 | PASS as implementation support only: representative production renders inspected; no golden refresh or canonical visual/release claim. |
| A023–A024 | PASS: tests use stable record/field IDs and accessible role/name identity; no Markdown dependencies or new selector registry. |
| A025 | PASS: authored test catalogs updated; Make generation and finalizer schema/catalog/drift checks passed; generated index was not hand-edited. |
| A026 | PASS: no new server/data/authority behavior; retired select API, preserved independent branches and no-migration rollback documented. |
| A027 | PASS: final focused regression, frontend typecheck, import boundary, Biome, Markdown and diff checks passed; final working-tree review found only this slice. This handoff records exact failures, evidence limits and rollback. |

Rollback: revert this cohesive frontend, tests/helpers, authored routing, generated
index and documentation slice. No data migration or server rollout is necessary.
