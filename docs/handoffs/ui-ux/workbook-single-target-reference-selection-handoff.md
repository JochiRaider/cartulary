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

The original focused regression command (nine rows, retained for the historical ledger below) was:

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

## Reference-test, Go-security, and harness remediation (2026-10-09)

This remediation starts at `a546f96b31ce1a4862cd4aad3044eb12ea424a18`.
The earlier failed full run, `single-reference-full-check-20261009a`, records a
different dirty source snapshot (`9ab6a79a530469e0044edd21325fd2c8e87533cd`).
Its 971/974 passed work units and 1,099 completed cleanup operations are
historical diagnosis, not qualification of this candidate.

The consumer audit adds ordinary-create reference controls and Timeline
related-Evidence Party authoring to focused verification. Both use the existing
scalar radio picker through their authoring owners. Lifecycle enumeration in
the Evidence form and independent ordinary direct-reference selects retain
legitimate select semantics. No picker compatibility adapter, selector registry,
request schema, stored draft format, or database migration is introduced.

The complete focused regression selection now includes all eleven rows:

```sh
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=web.workbook ROWS=web.workbook.regression.party_link_controls,web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.assessment_discovery,web.workbook.regression.reference_controls,web.workbook.regression.authoring_candidate_discovery,web.workbook.regression.note_create_authoring,web.workbook.regression.contextual_task_decision_authoring,web.workbook.regression.coordination_create_authoring,web.workbook.regression.workbookshell_assessments_suite_f2e4a841a2,web.workbook.regression.ordinary_create_controls,web.workbook.regression.timeline_related_evidence_authoring
```

### Implementation disposition

| Gap | Change and intended completion evidence |
| --- | --- |
| G1 | Harness defaults and bootstrap recovery guidance defer volatile values to the typed pin and readiness output. Human review preserves specification authority; Markdown is not an executable input. |
| G2 | Ordinary-reference tests select exact accessible radio identities, preserve staged selection across paging, and retain Apply, Escape, Retry, and suspension assertions. |
| G3 | Timeline Party tests use radio clicks and verify cancellation leaves the original identity unchanged; applying a candidate preserves independent collector text. Lifecycle assertions remain selects. |
| G4 | Go minimum and exact Make-selected compiler are 1.27.2; runtime admission agrees. x/net 0.60.0 requires crypto 0.57.0, sys 0.48.0, and text 0.42.0; text requires sync 0.23.0. Go manages checksums via bootstrap. |
| G5 | Successful fake build metadata reads the typed compiler and FIPS module selector; negative admission fixtures retain distinct wrong identities. |
| G6 | The scanner wrapper publishes invocation-bound diagnostics through the existing command-failure channel. Make-wrapped graph tests cover blocking, diagnostic-only, clean, malformed/missing output, setup and execution failures, conflicting diagnostics, independent aggregate work, and channel cleanup. |

Go removes an explicit `toolchain` directive when it equals the module minimum.
The specification and drift checker recognize the equivalent implied toolchain;
explicit mismatches still fail. This avoids bootstrap repeatedly causing drift.
The exact Make compiler and executable admission checks remain mandatory.
Bootstrap now downloads `all` through Go so that archive checksums are recorded;
argument-free download populated the cache but left new archive sums absent.
The additional `go.sum` entries are tool-managed closure checksums, not new
selected application requirements.
The cryptographic archive selector and digest remain unchanged because Go 1.27.2
contains the same reviewed archive bytes.

Contributors and builders must acquire Go 1.27.2 and rebuild all Go executables.
Deployment is a separate action; existing deployed binaries remain vulnerable
until replaced. On regression, repair forward or rebuild the prior application
revision with patched dependencies and compiler instead of restoring vulnerable
pins. Correct security failures now use the existing `security/security_finding`
classification with normalized exit 1; Make may still exit 2. Historical artifacts
are preserved.

### Qualification ledger

Every run below uses `.cartulary/test-results/<run-id>` and
`CARTULARY_OUTPUT_MODE=ci`. Named graph checks use
`CARTULARY_HARNESS_CACHE_MODE=off` for the frontend, harness, vulnerability,
cryptographic, and release runs. Each invocation has a fresh run ID. Other
readiness/build caches remain independently verified by their owners.

| Run ID (`reference-remediation-` prefix) | Command/selection | Outer exit; normalized result; cleanup |
| --- | --- | --- |
| `bootstrap-20261009a` | `make bootstrap` | 0; pass; non-graph tool summary, no graph cleanup ledger. |
| `generate-20261009a` | `make generate` | 0; pass; non-graph tool summary, no graph cleanup ledger. |
| `drift-20261009a` | `make toolchain-drift` | 0; pass 2/2 units; cache 0 hits; cleanup 4/4 completed. |
| `frontend-20261009a` | `make test-slice OWNER=web.workbook ROWS=` ordinary-create and Timeline related-Evidence rows | 0; pass 3/3 units; cache 0 hits; cleanup 5/5 completed. |
| `harness-20261009a` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity` | 0; pass 1/1 units; cache 0 hits; cleanup 3/3 completed. |
| `format-20261009a` | `make format` | 0; pass 2/2 units; cache 0 hits; cleanup 4/4 completed. |
| `vuln-20261009a` | `make go-vulncheck` before checksum repair | 2; fail 3/4 units; harness/tool_diagnostic_failure; cache 0 hits; cleanup 6/6 completed. |
| `crypto-20261009a` | `make cryptographic-policy-assessment` before checksum repair | 2; fail 2/3 units; harness/child_target_failure; cache 0 hits; cleanup 5/5 completed. |
| `bootstrap-20261009b` | `make bootstrap` with explicit module closure download | 0; pass; non-graph tool summary, no graph cleanup ledger. |
| `vuln-20261009b` | `make go-vulncheck` after checksum repair | 0; pass 4/4 units; cache 0 hits; cleanup 6/6 completed. |
| `crypto-20261009b` | `make cryptographic-policy-assessment` after checksum repair | 0; pass 3/3 units; cache 0 hits; cleanup 5/5 completed. |
| `finalize-20261009a` | `make agent-finalize` without `RESULTS_DIR` | 0; pass 1/1 units; cache 0 hits; cleanup 3/3 completed. |
| `markdown-20261009a` | `make lint-markdown` | 0; pass; non-graph tool summary, no graph cleanup ledger. |
| `scripts-20261009a` | `make lint-scripts` | 0; pass 2/2 units; cache 0 hits; cleanup 4/4 completed. |

The first vulnerability run correctly preserved `harness/tool_diagnostic_failure`
for failed package discovery, instead of misreporting a security verdict. Both
failed checks were followed by a source repair to bootstrap and fresh successful
runs. No failure was suppressed and no deadline was changed. Goldens were
unchanged at this stage; the later authorized refresh is recorded below.

`agent-finalize` passed with retained-run maintenance skipped because
`RESULTS_DIR` was unset; the failed historical run was not used as successful
evidence. Read-only investigations succeeded after correcting two invalid
lookup names (`harness-smoke-fast` and `harness.readiness`); those lookup errors
did not execute tests. An early receipt lookup used a nonexistent filename;
the retained assessment was then read at the actual path below.

The cryptographic assessment is retained at
`reference-remediation-crypto-20261009b/cryptographic-policy-assessment/assessment.json`.
It records Go 1.27.2 and matching FIPS identity for server, migrate, and operator,
15 application rejection cases, pinned execution success, disabled/ordinary
execution rejection, and strict-mode diagnostic success. Archive selector
`v1.0.0-c2097c7c` and SHA-256
`daf3614e0406f67ae6323c902db3f953a1effb199142362a039e7526dfb9368b`
remain unchanged. This assessment is implementation support, not deployment or
a separate cryptographic certification claim.

### Advisory closure

The fresh scan is
`reference-remediation-vuln-20261009b/unit-artifacts/target-go-vulncheck/govulncheck-findings.json`:
Go 1.27.2, Govulncheck 1.3.0, symbol/source mode, database revision
`2026-10-08T22:31:09Z`, zero blocking findings. Original-to-current mapping:

| Original blocking ID | Remediated baseline | Fresh scan |
| --- | --- | --- |
| GO-2026-6599 | Go 1.27.2 | Absent; closed. |
| GO-2026-6600 | Go 1.27.2 | Absent; closed. |
| GO-2026-6603 | Go 1.27.2 + x/net 0.60.0 | Absent; closed. |
| GO-2026-6604 | Go 1.27.2 | Absent; closed. |
| GO-2026-6605 | Go 1.27.2 | Absent; closed. |
| GO-2026-6607 | Go 1.27.2 | Absent; closed. |
| GO-2026-6608 | Go 1.27.2 | Absent; closed. |
| GO-2026-6609 | Go 1.27.2 | Absent; closed. |
| GO-2026-6610 | Go 1.27.2 + x/net 0.60.0 | Absent; closed. |
| GO-2026-6611 | Go 1.27.2 + x/net 0.60.0 | Absent; closed. |
| GO-2026-6612 | Go 1.27.2 + x/net 0.60.0 | Absent; closed. |
| GO-2026-6613 | Go 1.27.2 | Absent; closed. |
| GO-2026-6617 | Go 1.27.2 + x/net 0.60.0 | Absent; closed. |

GO-2026-5932 remains a module-only diagnostic for unmaintained
`golang.org/x/crypto/openpgp`; there is no package or symbol reachability in the
fresh result. Its metadata is retained without an advisory suppression.

### Changed-file inventory

Paths are repository-relative. Generated files were refreshed through Make.

- `AGENTS.md`
- `Makefile`
- `apps/web/e2e/contextual-create.spec.ts`
- `apps/web/e2e/sentinel.spec.ts`
- `apps/web/e2e/support/visual/metadataNormalization.ts`
- `apps/web/e2e/visual-harness.spec.ts`
- `apps/web/e2e/workbook.generic.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/contextual-decision-authoring-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/contextual-decision-authoring-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/contextual-decision-references-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/contextual-task-request-authoring-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/contextual-task-request-authoring-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/contextual-task-request-references-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/coordination-handoff-authoring-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/coordination-source-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/evidence-affordance-states-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/linked-note-authoring-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/linked-note-authoring-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/linked-note-source-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/ordinary-reference-authoring-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/timeline-related-evidence-party-narrow-linux.png`
- `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/components/SystemViewSwitcher.tsx`
- `apps/web/src/workbook/components/WorkbookAuthoringReferenceControl.tsx`
- `apps/web/src/workbook/components/WorkbookCandidateSelection.tsx`
- `apps/web/src/workbook/components/WorkbookShellTopBar.test.tsx`
- `apps/web/src/workbook/components/WorkbookShellTopBar.tsx`
- `apps/web/src/workbook/features/coordination/ContextualReferenceControl.tsx`
- `apps/web/src/workbook/features/evidence/timelineRelatedEvidenceAuthoring.test.tsx`
- `apps/web/src/workbook/features/notes/NoteSourceControl.tsx`
- `apps/web/src/workbook/features/ordinary/ordinaryCreateControls.test.tsx`
- `apps/web/src/workbook/layout/workbookShellStyles.ts`
- `docs/guides/cartulary_repository_bootstrap_guide.md`
- `docs/handoffs/ui-ux/workbook-single-target-reference-selection-handoff.md`
- `docs/testing-harness-nlspec.md`
- `go.mod`
- `go.sum`
- `internal/platform/cryptography/execution.go`
- `tools/execution_topology_render_index.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/frontend_visual_normalization.json`
- `tools/go-analysis/go.mod`
- `tools/go-analysis/go.sum`
- `tools/harness/backend/tests/test-build-go-artifact.sh`
- `tools/harness/readiness/bootstrap-go-tool.sh`
- `tools/harness/readiness/install-go-dependencies.sh`
- `tools/harness/readiness/tests/test-check-toolchain-pins.sh`
- `tools/harness/readiness/tests/test-go-toolchain-readiness.sh`
- `tools/harness/readiness/toolchain-pin-check-cli.mjs`
- `tools/harness/static-analysis/go-gosec-audit.sh`
- `tools/harness/static-analysis/go-gosec-targeted.sh`
- `tools/harness/static-analysis/go-govulncheck.sh`
- `tools/harness/static-analysis/go-staticcheck.sh`
- `tools/harness/static-analysis/tests/test-run-go-govulncheck.sh`
- `tools/harness/tests/command-failure-fixture.mjs`
- `tools/harness/tests/test-command-failure.mjs`
- `tools/task_surface.generated.mk`
- `tools/task_surface_manifest.json`
- `tools/task_surface_owner.json`
- `tools/toolchain_pins.json`

### Release findings and bounded follow-up

The first full release attempt, `reference-remediation-release-20261009a`,
exited 2 after 3,818,704 ms: 1,370/1,388 units passed, 17 failed, one was
skipped, and all 1,570 cleanup operations completed. Graph cache reuse was
zero. Its normalized aggregate cause is `product/test_assertion_failure`.
Its source was dirty at the baseline commit with digest
`sha256:9a253295d3f4fc0b1b481ec62dc292521b49fa2ff50a2802698d7b9caaf94b98`.
It is failed-candidate evidence, not final qualification.

The attempt exposed three additional maintenance issues:

- The six new Testing Library radio queries incorrectly supplied Playwright's
  `exact` option. That unsupported option was removed; string accessible names
  still match the complete label and stable ID. Frontend typecheck now passes.
- Contextual-create, generic-workbook, and sentinel browser helpers still used
  select events for migrated authoring controls. They now activate the native
  radio/checkbox groups. Shared ordinary editing and enum selection remain
  distinct legitimate select consumers. All five previously failing functional
  and stateful scenarios passed fresh focused verification.
- The old analyzers could not decode Go 1.27.2 export-data version 5. Updating
  only their release versions reproduced that same incompatibility. Staticcheck
  0.8.1 and gosec 2.29.0 now build from the isolated `tools/go-analysis` module
  with x/tools 0.51.0 and x/net 0.60.0. The application dependency graph does not
  include these tool dependencies. Bootstrap reconciles the Go-managed module
  closure; tool builds are read-only, cache identity includes both module files,
  and drift rejects mismatched tool versions or Go minimums. Analyzer configuration and
  suppressions are unchanged. Embedded binary metadata confirms the selected
  tool versions and x/tools 0.51.0 (plus x/net 0.60.0 in gosec). Lint and both
  gosec profiles pass.

The upstream [Staticcheck release](https://github.com/dominikh/go-tools/releases/tag/2026.2.1)
and [gosec release](https://github.com/securego/gosec/releases/tag/v2.29.0) supply the
selected tools. The [x/tools versioned reader](https://github.com/golang/tools/blob/v0.51.0/internal/pkgbits/version.go)
contains version-5 method-order support. The locked tool module resolves the
observed incompatibility without maintaining a fork or weakening compiler admission.

Additional fresh runs use the same prefix and retained-root convention as above:

| Run suffix | Command/selection | Outer exit; result; cleanup |
| --- | --- | --- |
| `format-20261009b` | `make format` | 0; 2/2 passed; 4/4 completed. |
| `bootstrap-20261009c` | `make bootstrap`, analyzer release upgrade | 0; non-graph pass. |
| `browser-workbook-20261009a` | Workbook Communications Log, generic workbook, stateful Timeline workflow rows | 0; 14/14 passed; 19/19 completed. |
| `browser-entities-20261009a` | Contextual task replay row | 0; 11/11 passed; 14/14 completed. |
| `browser-tasks-20261009a` | Task Requests and Decisions browser row | 0; 11/11 passed; 14/14 completed. |
| `lint-20261009a` | `make lint`, release-only analyzer upgrade | 2; 10/11 passed; harness/child_target_failure; 13/13 completed. |
| `gosec-targeted-20261009a` | `make go-gosec-targeted`, release-only upgrade | 2; 3/4 passed; harness/child_target_failure; 6/6 completed. |
| `lint-diagnostic-20261009a` | Same lint selection with verbose output for diagnosis | 2; same failure; 13/13 completed. |
| `gosec-diagnostic-20261009a` | Same targeted scan with verbose output for diagnosis | 2; same failure; 6/6 completed. |
| `generate-20261009b` | `make generate`, authored bootstrap projection | 0; non-graph pass. |
| `bootstrap-20261009d` | `make bootstrap`, isolated analyzer module | 0; non-graph pass. |
| `bootstrap-20261009e` | `make bootstrap`, patched tool-module x/net closure | 0; non-graph pass. |
| `lint-20261009b` | `make lint`, locked tool module | 0; 11/11 passed; 13/13 completed. |
| `gosec-targeted-20261009b` | `make go-gosec-targeted`, locked tool module | 0; 4/4 passed; 6/6 completed. |
| `gosec-audit-20261009a` | `make go-gosec-audit`, locked tool module | 0; 4/4 passed; 6/6 completed. |
| `finalize-20261009b` | `make agent-finalize`, without `RESULTS_DIR` | 0; 1/1 passed; 3/3 completed; retained-run maintenance skipped. |
| `frontend-20261009b` | Complete eleven-row picker regression selection | 0; 12/12 passed; 14/14 completed. |
| `harness-20261009b` | `make harness-contract` | 0; 2/2 passed; 4/4 completed. |
| `format-20261009c` | `make format` after helper-name clarification | 0; 2/2 passed; 4/4 completed. |

All graph runs in this follow-up used disabled graph caches. The two verbose
reproductions were diagnostic failures, never evidence of success. Their bounded
unit logs confirmed the same export-reader error. No deadline was increased.

### Authorized visual refresh review

The user subsequently authorized review and updating the affected goldens,
replacing the plan's original prohibition for this bounded refresh. The ordinary
release-run reconciliation accounts for 255 active captures and 255 committed
PNGs, with zero orphans, missing goldens, ambiguous mappings, or unresolved
registered fixtures. Renderer and golden-manifest identities passed; the
reconciliation's failure is the failed visual target attempt.

Artifact-mode review imported the exact 14 mismatched canonical capture IDs.
Actual and expected image bytes and SHA-256 values were verified before viewing.
The native radio controls and persistent stable IDs match the already-validated
reference behavior. Smaller text-rendering differences in the existing collection
and Evidence states preserve content, controls, and framing; visual inspection
found no missing data, unexpected clipping, or altered state semantics. These are
stale regression images eligible for a reviewed refresh, not new design authority. Source-scrolling and
explicit focus behavior remain under their existing assertions. Viewport, browser
zoom, masks, scroll normalization, renderer/font pins, and screenshot scope are
unchanged.

The review session closed successfully with 42 imported images and complete
cleanup. Image review did not run axe or establish a separate accessibility or
release claim. Private review bundles were consumed before deletion; durable
comparison evidence remains in the canonical failed run. The first transactional
update passed, but the first ordinary validation and a promoted-image review exposed
unstable generated UUID labels in eleven captures. That candidate is not accepted.
The existing declared metadata normalization will cover those ID leaves before
the next transactional update; two fresh ordinary passes remain required.

All filenames below are under
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`. Owner rows were checked
against the authored catalog; registry claims come from reconciliation.
Every capture retains 100% browser zoom.

| Golden filename | Semantic owner row | Registered fixture | Viewport; density |
| --- | --- | --- | --- |
| `contextual-decision-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | No registry claim | 1280x720; default |
| `contextual-decision-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | No registry claim | 768x640; default |
| `contextual-decision-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | No registry claim | 768x640; default |
| `contextual-task-request-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | No registry claim | 1280x720; default |
| `contextual-task-request-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | No registry claim | 768x640; default |
| `contextual-task-request-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | No registry claim | 768x640; default |
| `coordination-handoff-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` | 1280x720; compact |
| `coordination-source-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` | 768x640; compact |
| `evidence-affordance-states-linux.png` | `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `visual.fixture.evidence_affordance` | 1280x720; default |
| `linked-note-authoring-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | No registry claim | 1280x720; compact |
| `linked-note-authoring-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | No registry claim | 768x640; compact |
| `linked-note-source-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | No registry claim | 768x640; compact |
| `ordinary-reference-authoring-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | No registry claim | 1280x720; default |
| `timeline-related-evidence-party-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | No registry claim | 768x640; compact |

### Generated reference metadata stabilization

The first update, `reference-remediation-visual-update-20261009a`, exited 0:
96/96 units passed, no graph cache hits, and 100/100 cleanup operations completed.
It changed exactly the fourteen reviewed goldens and their manifest. A second
artifact session inspected the promoted source picker and closed with complete
cleanup. Eleven promoted images differed from the previous captures because
newly visible generated UUID labels had no capture normalization; the two
collection captures and Evidence capture were byte-identical.

The subsequent ordinary run, `reference-remediation-visual-20261009a`, exited 2
after 930,378 ms: 90/96 units passed, five visual groups and their aggregate
failed with `product/test_assertion_failure`, no graph cache hits, and all
105 cleanup operations completed. This is retained failure evidence, not a
successful golden qualification.

Four authoring components now mark only the selected generated ID text leaf as
`selected-reference-id`. Existing candidate ID markers are reused. The authored
normalization policy adds exact capture-specific rules for eleven captures:
13 candidates in each source picker, one in the ordinary owner and Timeline
Party pickers, and the exact mounted selected-ID counts for each form. Counts
include staged and committed presentations. Rules use the existing fixed UUID
replacement and validation/receipt mechanism; capture cleanup restores text.
Input values, accessible identity labels, authored source text, selection
behavior, viewport, zoom, masks, scroll anchors, screenshot scope, renderer,
fonts, tolerances, and functional assertions remain unchanged. No new selector
registry or schema is introduced. The narrowly scoped draft-chooser admission
correction below retains the existing normalization mechanism.

The second update, `reference-remediation-visual-update-20261009b`, exited 2
with 94/96 work units passed and 101/101 cleanup operations completed. Only the
ordinary-reference visual group and its aggregate failed. Its root error was
`Metadata source overlap candidate-reference-id`: the native picker is hosted in
a grid draft cell, which the guard treated as a saved source value. Transactional
failure preserved all committed goldens and manifest SHA-256
`93cd70bd1e6a19f3357fb35e3982fdd8e6a59bca0b178f4022b1d910ee0cbd1b`.

TH-HARNESS-REQ-819 now clarifies the distinction between generated identity
labels in a draft chooser and protected source values. The normalizer admits
only the existing candidate-ID leaf marker inside a native single/multi chooser
within a marked grid draft row. Committed cells, labels marked as source,
input/editor values, and non-chooser draft content remain rejected. The existing
Timeline capture-support row verifies both picker modes, unchanged input values
and checked state, restoration, and five rejection cases, including an empty
committed field key. Its row identity and
scenario title remain unchanged.

Additional validation after metadata markers:

| Run suffix | Command | Outer exit; result; cleanup |
| --- | --- | --- |
| `format-20261009d` | `make format` | 0; 2/2 passed; 4/4 completed. |
| `finalize-20261009c` | `make agent-finalize`, without `RESULTS_DIR` | 0; 1/1 passed; 3/3 completed; retained-run maintenance skipped. |
| `json-20261009a` | `make json-shape-check` | 0; 3/3 passed; 5/5 completed. |
| `markdown-20261009b` | `make lint-markdown` | 0; non-graph pass. |
| `frontend-20261009c` | Eleven picker rows after metadata markers | 0; 12/12 passed; 14/14 completed. |
| `format-20261009e` | `make format` after normalization guard and tests | 0; 2/2 passed; 4/4 completed. |
| `metadata-20261009a` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.timeline_investigation_capture` | 0; 11/11 passed; 14/14 completed. |
| `biome-20261009a` | `make lint-biome` | 2; 1/2 passed; harness/child_target_failure; 4/4 completed. Five non-null assertions in the new test fixture were replaced with an explicit presence check. |
| `finalize-20261009d` | `make agent-finalize`, without `RESULTS_DIR` | 0; 1/1 passed; 3/3 completed; retained-run maintenance skipped. |
| `biome-20261009b` | `make lint-biome` after fixture repair | 0; 2/2 passed; 4/4 completed. |

The third transactional update, `reference-remediation-visual-update-20261009c`,
exited 0 after 889,492 ms: 96/96 work units passed, graph cache hits were zero,
and 100/100 cleanup operations completed. It replaced exactly the eleven unstable
images; the other three reviewed images remained byte-identical. The final golden
manifest SHA-256 is
`66151fa1fbc36219d0c6bcc14441128aeb375fc29ef7cf18a2f353c92f1b685c`.
All 255 capture/golden mappings reconcile, including all registered fixtures.

A final artifact-mode review imported those eleven promoted PNGs, verified each
bundle's byte count and SHA-256 against the manifest, and inspected every image.
Reference labels, native controls, focus outlines, source text and framing are
preserved; only declared generated ID presentation is stabilized. No unexpected
clipping or missing content was observed. The other three images retain their
prior review through exact byte identity. The review closed successfully at
`reference-remediation-golden-review-20261009c/ui-review/terminal.json`, with
11 images, no DOM/axe observations, and complete cleanup. Private images and
caller request scratch were deleted after consumption.

The grid guard also retains rejection when a committed field marker has an empty
value; it tests marker presence instead of relying on a truthy field key. This
admission hardening does not change any reviewed capture. After that final change,
`reference-remediation-finalize-20261009e` passed (1/1 units, 3/3 cleanup, outer 0,
`RESULTS_DIR` unset), and `reference-remediation-metadata-20261009b` passed the
focused browser contract (11/11 units, 14/14 cleanup, outer 0, cache disabled).

### Final qualification

The first fresh ordinary validation, `reference-remediation-visual-20261009b`,
exited 0 after 871,356 ms: 96/96 work units passed, no harness result-cache reuse,
and 100/100 cleanup operations completed. Reconciliation passed against the
final manifest SHA-256 recorded above. The candidate source is dirty at
`a546f96b31ce1a4862cd4aad3044eb12ea424a18`, with executable-source digest
`sha256:faa3e25dba16480002572431fdca6e72499c4c8e09e7f039b2b63ee9a3c16774`.
`reference-remediation-markdown-20261009c` also passed `make lint-markdown`.

The `reference-remediation-release-20261009b` aggregate exited 2 after
3,701,156 ms: 1,386/1,388 work units passed, two failed, none skipped, and all
1,563 cleanup operations completed. Harness cache reuse was disabled. The
grid-provenance browser row and its target-summary rollup both report
`product/test_assertion_failure`; this run cannot qualify the release. Its
ordinary visual stage passed: all 44 groups completed, and
`browser-e2e-visual/frontend-visual-reconciliation.json` reconciles 255 captures
against the same final manifest with no errors. Together with the standalone
ordinary run above, this supplies the two fresh visual passes required after
the authorized golden update. It does not turn the failed aggregate into release
qualification. No redundant full `make check` was run immediately before it.

The failing row is
`module.entities.browser.verify_hosts_identities_and_notes_grids_render_c_7531b57a50`,
under `browser-e2e-webserver-backed/browser-groups/functional-support-default-grid-provenance`.
Its retained trace shows a completed Hosts-tab click after reload while the
shell's incident/session reads are still pending. No Hosts query follows; Notes
remains selected and navigation remains idle until the row's existing deadline.
The session navigation owner correctly rejects requests while unreadable, but
the shell presents its navigation buttons as enabled during that interval.
This is a separate product admission/presentation defect, not another reference
picker selector migration. The bounded correction derives native control
availability from the same existing read-authority predicate and retains owner
admission. Desktop tabs, the compact Surfaces menu, and More views are disabled
until reads are admitted; open navigation menus close on uncertainty and do not
reopen on recovery. Existing tokens distinguish the unavailable controls.
Source edits began only after the aggregate exited. Deterministic regression
coverage extends the existing component and shell-authority rows without
changing their titles or the failing browser test's interaction or deadline.

Core 03 §2.5 (especially REQ-03-314/315), Core 04 REQ-04-169, and design §§7.2/8.4
govern the correction. The shell composes authority, the navigation owner admits
destinations, and the top bar presents availability; no new authorization cache,
queued intent, startup override, compatibility layer, route or stored-data
migration is introduced. Closed incidents remain readable. Leaving this mismatch
would continue to lose deliberate navigation during startup or session recovery.
Advisory disabled-state and keyboard guidance is adapted through those owners
and existing tokens; no upstream visual prescription becomes a new authority.

Focused navigation validation passed with harness cache reuse disabled:

| Run ID (`reference-remediation-` prefix) | Command/selection | Outer exit; result; cleanup |
| --- | --- | --- |
| `format-20261009f` | `make format` | 0; 2/2 units; 4/4 cleanup. |
| `navigation-20261009a` | `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.desktop_built_in_surface_selector_keyboard_navig_ebad53c0ea,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e` | 0; 3/3 units; 5/5 cleanup. |
| `grid-provenance-20261009a` | `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.verify_hosts_identities_and_notes_grids_render_c_7531b57a50` | 0; 11/11 units; 14/14 cleanup. |
| `biome-20261009c` | `make lint-biome` | 0; 2/2 units; 4/4 cleanup. |
| `typecheck-20261009a` | `make frontend-typecheck` | 0; 2/2 units; 4/4 cleanup. |
| `finalize-20261009f` | `make agent-finalize` with `RESULTS_DIR` unset | 0; 1/1 units; 3/3 cleanup; retained-run maintenance skipped. |

The component cases prove unavailable pointer activation does nothing, recovery
restores keyboard/pointer selection, and open desktop/compact menus retire on
uncertainty without reopening. The shell case proves the same wiring through
the real authority owner and successful navigation after recovery in a closed
incident. Existing delayed-startup selection coverage also passes. The original
browser row passes with its unchanged reload, tab click, provenance assertions
and deadline. No new timing wait or golden refresh was needed for this correction.

The completed candidate remains dirty at
`a546f96b31ce1a4862cd4aad3044eb12ea424a18`, with executable-source digest
`sha256:6edfa748429e86456e2c0a1ee4f7c4041e58381e4c1205242e61b54e5a61cdcc`.
Final qualification is blocked in `reference-remediation-release-20261009c`,
which exited 2 after 3,720,325 ms (62 minutes). It passed 1,387/1,388 work units,
with one failure, none skipped or cancelled, zero cache hits and 1,388 cache
bypasses. All 1,562 terminal cleanup operations completed successfully.
Unit `go:bc72395d4fcf3e4c:b3f543dc9303` failed row
`platform.jobs.integration.expiry_and_compaction`; the harness preserves
`product/test_assertion_failure`, normalized exit 10. Its exact retained log is
`unit-logs/go-bc72395d4fcf3e4c-b3f543dc9303/stderr.log`.
`TestCompactExpiredJobs_Integration` completed its behavior assertions and
operator-attribution subtest, then failed at `internal/testutil/pgtest/pgtest.go:677`
while deleting its isolated database:

```text
force drop postgres database ct_4c3a4cde_fdfe65bf_000001_jobs_compact_expired: timeout: context deadline exceeded
```

The failed row's published result is
`rows/platform.jobs.integration.expiry_and_compaction.json`. This is teardown
failure evidence, not a Jobs behavior assertion failure, despite the current
Go-test normalization. The source was not changed in this cleanup path.
Inspection confirms `t.Cleanup(pool.Close)` runs before database deletion;
the helper uses separate five-second normal and fifteen-second forced operation
budgets with exact-owner proof and target-specific advisory admission. The
ordinary failure is not retained in the final error when forced cleanup also
fails, and the published evidence contains no PostgreSQL wait/blocker sample.
Host and temporary filesystems have ample free space; that observation does not
establish the database's wait cause. The underlying stall remains undiagnosed.

No deadline, capacity, cleanup policy or Jobs behavior was changed to conceal
this failure, and this unchanged failure has not been retried. Qualification
requires bounded, redacted diagnostics for the exact-owned drop to establish
the PostgreSQL wait cause, an evidence-backed correction, focused cleanup/Jobs
validation, and a successful release aggregate. The original remediation and
navigation fixes remain independently evidenced; the successful terminal resource
sweep does not erase this failed eager-cleanup result.

For the additional navigation correction, the digest acceptance assessment is
bounded to the read-admission presentation seam:

| Acceptance rows | Disposition and evidence |
| --- | --- |
| A001–A003 | PASS: the owner clauses, existing `web.workbook` source boundary, independent browser owner `module.entities`, baseline and observed defect are recorded above. |
| A004–A005 | PASS: existing graphite and disabled-command tokens; no new theme or token registry. |
| A006–A007 | N/A: density, creation capability and source authoring are unchanged. |
| A008–A009 | PASS within scope: existing responsive thresholds and footer hosts remain; focused desktop/compact keyboard and pointer cases pass. |
| A010 | N/A: inspector dispatch and review state are unchanged. |
| A011 | PASS: existing delayed-startup selection coverage and the unchanged reload/provenance browser row pass; navigation state remains with its owner. |
| A012–A015 | N/A: transaction identity, replay, acknowledgement, editor acceptance and conflict behavior are unchanged. |
| A016–A017 | PASS: native control availability uses current read authority, uncertainty closes menus, recovery re-enables navigation, and closure alone preserves reads. Focused shell-owner tests pass. |
| A018 | N/A: Evidence lifecycle and preview are unchanged. |
| A019–A020 | PASS within scope: native disabled semantics, existing tokens, desktop/compact keyboard activation, safe access-check reason and menu recovery are covered by focused tests. The final candidate's accessibility and visual targets also pass. |
| A021 | N/A: virtualized row identity, grid coordinates and query paging are unchanged. |
| A022 | PASS: the final candidate passes all 44 ordinary visual groups; 255 captures and 255 goldens reconcile with no errors against the unchanged reviewed manifest. |
| A023–A026 | PASS within scope: existing semantic selectors and row titles; no Markdown execution dependency, generated edits, compatibility adapter, authorization policy or persisted format change. |
| A027 | BLOCKED for overall release qualification by the documented PostgreSQL eager-cleanup failure; implementation evidence and the remaining work are retained here. |

The final candidate's `browser-e2e-visual/browser-target-result.json` and
`browser-e2e-visual/frontend-visual-reconciliation.json` pass against manifest
`66151fa1fbc36219d0c6bcc14441128aeb375fc29ef7cf18a2f353c92f1b685c`.
All six browser target summaries (accessibility, measurement, stateful, support,
visual and webserver-backed) pass. This also verifies the navigation correction
without another golden update. The earlier two fresh visual passes remain
retained evidence for the authorized golden refresh.
All eleven picker regression rows listed above also pass in this final candidate,
including the two omitted authoring rows; none used harness result-cache reuse.

The completed candidate's fresh symbol scan at
`reference-remediation-release-20261009c/unit-artifacts/target-go-vulncheck/govulncheck-findings.json`
passed with zero blocking IDs, reconfirming closure of every ID in the mapping
above. It reports Go 1.27.2, Govulncheck 1.3.0 and database revision
`2026-10-08T22:31:09Z`. Only GO-2026-5932 remains, at module reachability; package
and symbol counts are zero. Both gosec profiles, Go lint, frontend typecheck,
and toolchain drift also passed within this candidate's aggregate. The repaired
grid-provenance browser group passed here as well. These passes do not clear the
independent PostgreSQL teardown blocker.

The current executable admission evidence is
`reference-remediation-release-20261009c/cryptographic-policy-assessment/assessment.json`.
All three receipts agree on Go 1.27.2, x/net v0.60.0, selector
`v1.0.0-c2097c7c`, and the unchanged archive digest. The 15 application rejection
cases and pinned/disabled/ordinary/strict execution checks pass their expected
outcomes. Verified build-receipt identities (release qualification remains blocked):

| Executable | Binary SHA-256 |
| --- | --- |
| server | `69cb5e07cb732714afbe72721105c9605d527cc7adf676fa04152ccf75b9e38e` |
| migrate | `6997d79ae484b180da9795ecdbac94f37b74ddb4cc2bd25248f202388f29ce47` |
| operator | `486e56760f479d10411a10955d203b94d6216151bfaf8167d1486182829d833a` |

### Phase exits and remaining work

| Phase | Disposition |
| --- | --- |
| 1. Authority and baseline | Complete: reviewed specification corrections, raised Go minimum, owner boundaries and historical/current evidence distinction recorded. |
| 2. Implementation | Complete for G1–G6: native-radio tests, patched dependency/compiler graph, typed security attribution, pin-derived fixture metadata and documentation are implemented. Analyzer compatibility, the authorized visual refresh and the observed navigation correction are also implemented and documented. |
| 3. Focused convergence | Complete: generated projections, focused owner checks, security scans, cryptographic admission, normalization guards, navigation regression, type checking, lint and required finalization pass. Retained-run maintenance was skipped because `RESULTS_DIR` was unset. |
| 4. Release qualification and handoff | Blocked: the final candidate's release aggregate has the exact PostgreSQL eager-cleanup failure recorded above. All browser targets and the remediation-specific gates pass; that does not satisfy the aggregate exit criterion. The terminal result, complete cleanup and remaining diagnosis are retained here. |

The final aggregate command was:

```sh
CARTULARY_TEST_RUN_ID=reference-remediation-release-20261009c CARTULARY_OUTPUT_MODE=ci CARTULARY_HARNESS_CACHE_MODE=off make release-check
```

Its root is `.cartulary/test-results/reference-remediation-release-20261009c`.
`run-summary.json` and `target-summaries/release-check.json` both retain
`product/test_assertion_failure`; the failed row records normalized exit 10,
and the outer Make process exited 2. `cleanup-results.json` records all 1,562
operations as completed. `run-manifest.json` records the final dirty source
identity above. The passing release-evidence work unit does not override the
failed aggregate or authorize deployment. No unchanged rerun followed this
failure, and no failed run was supplied as successful retained evidence.

Final documentation validation passed: `reference-remediation-markdown-20261009d`
ran `make lint-markdown` with outer exit 0 in 16,220 ms; its retained summary is
`adhoc/lint-markdown/tool-run-summary.json`. `git diff --check` passed, and the
changed-file inventory covers all 61 paths, including the two new analyzer module
files. These documentation checks do not replace the blocked release gate.

The remaining work is to obtain bounded PostgreSQL wait/blocker evidence for
the exact-owned failing cleanup, correct the established cause without restoring
vulnerable pins or weakening cleanup, validate the affected fixture/Jobs rows,
then run the finalizer and a fresh release aggregate. Preserve this failed run
as diagnosis. Deployment remains a separate release action after qualification.
