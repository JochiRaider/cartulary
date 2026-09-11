# Workbook entity merge review and recovery

## Baseline and authority

Implementation entry: clean `main` at
`6ac49305dac340059ac9e17f705bc2cb491c49d7`. Root AGENTS.md is the only applicable
instruction file. Go 1.27.1 (module 1.27.0), Node 24.15.0 and pnpm 10.33.0 match
the repository pins. All three requested task guides passed again on entry.
The digest was read in its README order during planning and its source mappings
revalidated. Embedded historical instructions are evidence, not authorization.

Core 01 REQ-01-104 and §3.3.5.4 own locks, authorization, request identity,
receipts and rollback. Core 02 §8.2, REQ-02-270 and §9 own identifier comparison,
claims, carry-forward and historical loser retention; §15.4 owns merge history.
Core 03 REQ-03-291/292 and §16.2 own confirmation, invalidation and continuity.
Core 04 owns current membership, session lifecycle, CSRF and concealment.
Design §§7.3, 10, 12 and 14 own inspector, local feedback and accessibility.
Domain vocabulary and research guidance introduce no additional behavior.

User-approved owner decision: clarify Core 02 §9 to establish loser canonical
first, then distinct secondary values in normalized-value order. Existing
unordered collection positions are not promotion order. The subsequent explicit
authorization also covers backend normalizer reconciliation and active-claim
compatibility, as detailed below. No public wire change is authorized. Any
further owner mismatch or material scope expansion blocks dependent work.

Allowed: the merge workflow, necessary workbook runtime/transport/query
composition, focused tests and browser support, owner-controlled fixtures,
authored selectors/routing/ownership inputs, Make-generated derivatives,
reviewed affected goldens, the approved Core 02 clarification and this handoff.
Preserve completed history, save-status, inspector and Network Analysis work,
the digest and existing handoffs. The authorized normalizer repair adds migration
41 without rewriting stored data. No dependency, authorization policy or route
change; no commit, reset, push or deployment.

Current verification owners: web.workbook for browser runtime/model/controller;
module.entities for merge source/HTTP and service/browser evidence;
module.revisions for existing rollback. package.ui owns authored selectors.
Routing remains under contracts/verification and tools/test_families;
new frontend paths require explicit tools/frontend_source_ownership entries.
Markdown is human evidence only, never a runtime, generator or test input.

## Tracker

Only the current row may be IN_PROGRESS. Close it before advancing.

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| EM-01 Baseline and characterization | DONE | Baseline, five gaps and failing promotion characterization retained. |
| EM-02 Planning and confirmation | DONE | Planner, captured review, normalization and active-claim compatibility pass. |
| EM-03 Retained attempts and recovery | DONE | Retained exact transport, scoped admission and regression checks pass. |
| EM-04 Reconciliation and browser evidence | DONE | Exact replay, receipt reconciliation, History rollback and accessibility pass. |
| EM-05 Final verification | DONE | Full check 826/826; final handoff and post-edit audit recorded below. |

## Gap ledger

| Gap / affected area | Owner / remedy and rationale | Compatibility / unresolved risk | Binary exit |
| --- | --- | --- | --- |
| G1 entityWorkbookModel promotes before duplicate checks, separates secondary candidates and lowercases aliases | Core 02 REQ-02-060–066/270; unified typed per-class plan and separate alias comparison | Loaded collections already provide required values; Unicode parity and within-class order must be proven | Corpus agrees across frontend and source-owner evidence; duplicate never promotes |
| G2 controller confirms mutable pair/version/reason | Core 03 REQ-03-291/292/249; immutable reviewed snapshot with synchronous invalidation | Internal UI change only; drafts must survive preparation | Changed record, version, reason, view or authority sends zero requests |
| G3 command merge creates a fresh ID on each invocation and retains no uncertainty | Core 01 REQ-01-104/187; runtime-owned attempt, synchronous record reservation, immutable body | Same public route; no lock API or autosave queue extension | Duplicate admission refused and replay bytes/ID identical, including after loser removal |
| G4 controller lifetime drops completion and refresh errors can be swallowed | Core 03 continuity and lifecycle; retain acknowledgement before acceptance-aware reconciliation | Existing query/continuity paths retained; stale callbacks and drafts need fences | Failed refresh preserves receipt; retry sends zero merges and takes no newer focus |
| G5 accepted mapping drops incident and merge_summary | Core 01 REQ-01-193–195; complete generated receipt and existing history route | Wire/data unchanged; malformed successes remain uncertain | Validate identities/versions/counts and retain complete summary through closure |

## Existing evidence

Latest check-remediation handoff records full make check PASS 820/820 at
`.cartulary/test-results/20260910T201249Z-p98482`; no older waived failure is
carried forward. This is historical baseline, not final-source verification.
Planning focused model/controller baseline passed 3/3 at
`.cartulary/test-results/20260910T203219Z-p63819` with:

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.entityworkbookmodel_builds_merge_plans_without_c_6b84a1bd7f,web.workbook.regression.workbookz_entity_merge_controller_lifecycle_7ad8b846c1
```

## Compatibility and rollback

Public interfaces, stored state, server authorization, destructive locks and
existing change-set rollback remain unchanged. Revert coordinated implementation,
fixtures, generated derivatives and the approved owner clarification to roll
back. Delete no entities, history, revisions, receipts or analyst data. No later
refactor is authorized.

## EM-01 exit

Changed this handoff, the authorized Core 02 §9 ordering paragraph, and the
existing entityWorkbookModel merge test. The latter now requires duplicate no-op
before promotion and case-preserving alias comparison. Narrow characterization:

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.entityworkbookmodel_builds_merge_plans_without_c_6b84a1bd7f
```

FAIL 1/2 units at `.cartulary/test-results/20260910T203747Z-p66281`, as expected:
actual `Promote device-existing`, required `Duplicate no-op device-existing`.
The run-summary and unit logs retain the exact failure. G2-G5 are source-backed
characterizations; dynamic regression evidence will accompany their owner work.
All requested entry task guides passed. No implementation, public contract,
storage or server semantic change yet. Remaining risk: normalization parity and
captured confirmation; next action is the machine corpus and faithful planner.

## EM-02 blocked exit

The approved plan explicitly requires cross-language normalization parity and
stops on an unresolved mismatch. Expanding the existing owner-controlled corpus
exposed a real mismatch in the unchanged backend normalizers:

| Input | Go fieldnorm.NormalizeIdentifier | PostgreSQL entities_normalize_identifier_v1 |
| --- | --- | --- |
| identifier_type=sid, raw_value=ß (U+00DF) | ß (U+00DF) | ẞ (U+1E9E) |

The added `sid_simple_sharp_s` case uses the existing Go simple-uppercase
result as its expectation. This fixture choice is characterization, not an
adopted decision that Go should supersede PostgreSQL. REQ-02-270 requires one
normalizer and Go/PostgreSQL equality; neither changing the expected value to
PostgreSQL's output nor using browser full uppercase (`SS`) resolves parity.
The other added cases cover context-independent sigma, dotted capital I,
Greek uppercase, Unicode next-line edge whitespace and BOM rejection.

Exact commands/results:

```bash
make generate
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.support_integration.entity_identifier_claims_37_3ba3e2aaaf
```

Generation PASS at `.cartulary/test-results/20260910T203906Z-p67159`;
the only generated modification is internal/gen/contractentities/artifacts_gen.go.
Service-backed normalization verification FAIL 2/3 units at
`.cartulary/test-results/20260910T204004Z-p70243`. The failing test is
TestEntityIdentifierNormalizationAndClaimsMigration_Integration /
valid_36_to_37_upgrade_projects_the_closed_corpus_and_supports_disposable_rollback /
sid_simple_sharp_s. The retained diagnostic reports Go="ß" SQL="ẞ".
All other normalization cases in that run passed. Evidence is in run-summary.json
and unit-logs/go-39abafc5fcf7c949-b2bc31f0d95c/stderr.log.

This is an uncovered backend projection mismatch, not a new failing frontend
implementation or a waived historical check failure. The service test uses an
isolated disposable database. No analyst database or stored claims were changed.

The incomplete frontend normalizer/planner and its integration were removed
after discovering the blocker. Only this session's unfinished edits were removed;
the EM-01 failing test and all retained owner/normalization evidence remain.
No runtime recovery work, captured confirmation, merge-planning corpus, browser
scenario or visual change has been completed. EM-03 through EM-05 remain pending.

Current changed paths:

- apps/web/src/workbook/models/entityWorkbookModel.test.ts
- contracts/entities/identifier-normalization-corpus.v1.json
- docs/spec/02_domain_model_schema_and_history.md
- internal/gen/contractentities/artifacts_gen.go (Make-generated)
- docs/handoffs/workbook-entity-merge-recovery-refactor-handoff.md

Compatibility: production frontend/backend code, HTTP interfaces and stored data
remain unchanged. The authorized ordering clarification and additional tests
expose requirements without implementing the remaining seam. Product verification
is deliberately red on the characterized planner bug and normalization mismatch.

Unresolved risk: choosing either casing result may affect matching and existing
active claims for affected identifiers. A repair must select the owner-backed
mapping, reconcile both projections, and establish migration/claim-validation
requirements before changing a backend normalizer. The existing authorization
covers ordering clarification only and does not authorize that repair.

Exit criterion: the versioned normalization corpus agrees across Go, PostgreSQL
and the future frontend projection without silently rewriting or reassigning
identifier claims. Next action requires authorization to expand scope for this
specific normalizer/claim issue. Do not bypass this dependency or mark EM-02 DONE.

Skipped because of this stop: make agent-finalize, coordinated full make check,
merge recovery/service/browser/a11y/visual validation and EM-05 completion.
RESULTS_DIR remains unset; no retained-run maintenance or final-source success
is claimed. Checkpoint Markdown, generated drift and scope validation follow.

## Blocked checkpoint verification

`make lint-markdown generate-drift`: PASS. Markdown retained root is
`.cartulary/test-results/20260910T204303Z-p88269`, with
adhoc/lint-markdown/tool-run-summary.json (68350ms). Generate drift passed 4/4
at `.cartulary/test-results/20260910T204303Z-p88197` (8674ms).
`git diff --check`: PASS. Branch/HEAD remain the entry main/HEAD, the index is
empty, and the working-tree audit contains exactly the five paths listed above.
No production implementation file, dependency, migration, existing handoff or
digest change remains. No service claim or analyst data was modified.

After this final handoff edit, repeat `make lint-markdown`, `git diff --check`
and the same branch/HEAD/index/changed-path audit. EM-01 evidence is retained;
EM-02 is BLOCKED and later workstreams have not started. This is an incomplete
implementation checkpoint requiring the explicit scope decision described above,
not a completed seam or successful full developer gate.

## EM-02 resumed scope

The user explicitly authorized reconciling the backend normalizers and assessing
active-claim compatibility. This supersedes the earlier scope block for that
specific issue. Preserve the existing PostgreSQL U+00DF to U+1E9E identity in Go
and the future frontend normalizer; do not redefine stored claims to match Go's
previous U+00DF identity. The source-integrity trigger already requires this SQL
mapping for preserved identifiers, and canonical claims already use it.
No migration or claim rewrite is intended. Corpus, scalar-casing parity and
existing-claim reuse tests must establish that this bounded repair is sufficient.
The remaining workflow and scope restrictions continue to apply.

### Expanded normalization investigation and implementation

The initial no-migration assumption was disproven. Service-backed scalar casing
characterization found 124 Go/PostgreSQL disagreements in each direction at
`.cartulary/test-results/20260910T205549Z-p4005`. The database reports Unicode
16.0; Go and x/text report 17.0. Expanded NFC characterization also found 34
mismatches: stream-safe CGJ insertion on long combining runs and Unicode 17
combining-class assignments. Evidence:
`.cartulary/test-results/20260910T210434Z-p24319` and
`.cartulary/test-results/20260910T210653Z-p42213` (expected failures).

Owner decision within the newly authorized reconciliation: preserve database
Unicode 16 NFC and SID sharp-S identity, freeze scalar casing to Unicode 16
plus that existing sharp-S mapping, and reject incompatible stored identities
transactionally. The machine substrate comes from
[Unicode 16 UnicodeData](https://www.unicode.org/Public/16.0.0/ucd/UnicodeData.txt),
with its digest recorded in `identifier-unicode.v1.json`. NFC composition uses
[Unicode normalization stability](https://www.unicode.org/reports/tr15/)
for the frozen assigned repertoire; unassigned scalars remain inert starters.
No network or Markdown access is part of runtime, tests or generation.

A forward migration is necessary because historical migrations may already be
applied. Core 01 §2.1A/REQ-01-661 head references advance to 41 with the new
Entities-owned source. Core 02 REQ-02-270 records the concrete normalization
identity and fail-closed compatibility boundary. The migration preflight covers
canonical values in active/inactive source and both revision snapshots,
preserved identifiers including historical mutation values, and exact active
claim-set validity. It performs no source, claim, history or receipt rewrite.
Existing create/patch request normalization stays unchanged to preserve retained
request hashes. Entity alias carry-forward, rollback and bundle validation use
the database NFC comparison without broadening C0/C1 rejection or changing case.

Additional allowed paths for this authorized repair: the Entities fieldnorm
substrate and source consumers, migration 41, the migration owner assignment,
Core 01/02 owner clarification, generated migration manifests, and focused
compatibility evidence. No historical migration, runtime authorization, route,
dependency, source schema, table allocation or production data mutation is
planned. Disposable migration Down is test cleanup, not production rollback.

`make format generate` first passed formatting then failed generation at
`.cartulary/test-results/20260910T211247Z-p60887`: the new migration used prohibited
`CREATE OR REPLACE`. The migration was corrected to explicit drop/create with
existing grants, triggers and the dependent CHECK restored transactionally;
no policy relaxation. `make generate` passed at
`.cartulary/test-results/20260910T211436Z-p66230` and, after restoring the explicit
CHECK dependency, at `.cartulary/test-results/20260910T211758Z-p99096`.
The intervening service startup failure at
`.cartulary/test-results/20260910T211614Z-p73960` and migration-drift failure at
`.cartulary/test-results/20260910T211733Z-p92628` were related to that dependency.
`make format service-backed-test-slice OWNER=...` was rejected as invalid target
usage (format does not accept OWNER); the separate `make format` passed at
`.cartulary/test-results/20260910T211605Z-p69734`.

EM-02 remains IN_PROGRESS. Required exit still includes passing normalization
and active-claim compatibility evidence, faithful planning, captured review and
synchronous invalidation; later workstreams remain unstarted.


### EM-02 implementation evidence

The typed planner now consumes canonical fields and validated reusable and alias
collections. It uses the Unicode machine projection, canonical-first then scalar
normalized secondary ordering, duplicate checks before promotion, and separate
case-preserving alias comparison. Invalid or incomplete inputs block review.
The historical loser/provenance explanation makes no invented dependency counts.
An immutable reviewed pair includes full IDs, labels, versions, reason, plan,
actor/session/incident and presentation context. Changes invalidate review
synchronously; confirmation checks again before dispatch. Affected drafts require
explicit completion or scoped discard; unrelated drafts survive.

New authored sources are entityIdentifierNormalization.ts, entityIdentifierClasses.ts,
entityMergePlan.ts and its fixture tests, entityMergeReview.ts, and the merge-owner
authority foundation. Model/controller/inspector/runtime/shell composition and
selectors are coordinated. The merge-planning corpus and schema are under
contracts/entities. The protocol entities entrypoint exposes only the generated
Unicode projection; its exact package/entrypoint/import owner inputs and generator
were updated. Existing HTTP-only import rules remain intact. Source ownership and
focused test selectors are authored, with derivatives generated through Make.

Passing evidence:

- `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.support_integration.entity_identifier_claims_37_3ba3e2aaaf`:
  PASS 3/3 at `.cartulary/test-results/20260910T212009Z-p21210`.
  Frozen scalar casing and broad canonical decomposition/combining-run parity
  agree with PostgreSQL; compatible upgrades preserve source/history/claim/receipt
  snapshots. Incompatible current or historical identifiers reject the migration
  transactionally. Disposable Down/reapply and exact grants are covered.
- `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.store.explicit_merge_store_behavior_repoints_live_ment_0560122310`:
  PASS 3/3 at `.cartulary/test-results/20260910T213323Z-p76211`.
  Both entity types promote scalar-ordered secondary values. No merge backend
  ordering change was necessary; the fixture uses actual NULL canonical fields.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.entity_merge_planning_contract,web.workbook.regression.entityworkbookmodel_builds_merge_plans_without_c_6b84a1bd7f,web.workbook.regression.workbookz_entity_merge_controller_lifecycle_7ad8b846c1`:
  PASS 4/4 at `.cartulary/test-results/20260910T215300Z-p31209`.
- `make frontend-typecheck frontend-import-boundary-check`: PASS 2/2 each at
  `.cartulary/test-results/20260910T215300Z-p31324` and
  `.cartulary/test-results/20260910T215300Z-p31336`.
- `make format generate`: PASS at
  `.cartulary/test-results/20260910T215455Z-p50571` (2/2 format) and
  `.cartulary/test-results/20260910T215455Z-p50491` (generation).

Intermediate related failures were fixed, not waived: Vite fixture path and NULL
seed issues (`213058Z-p55849`, `213058Z-p55859`); same-turn React review invalidation,
type-import cycle and missing declared entrypoint (`214441Z-p4248`,
`214441Z-p4378`); generator's closed entrypoint set (`214841Z-p24353`). Earlier
formatting at `214841Z-p24449` passed even though that generation failed.
The real HTTP claim-reuse assertion initially supplied a different display name,
which intentionally updates an entity under existing upsert behavior
(`214441Z-p4255`). Corrected same-name evidence exposed only create-response
nanoseconds versus PostgreSQL microsecond timestamp projection
(`215300Z-p31233`, `215515Z-p57477`); row version and revision count remained 1.
The test now compares that timestamp at storage precision while preserving all
other row and history checks. Its rerun is pending. The attempted
`make task-guide ROLE=module-author OWNER=platform.fieldnorm` was rejected because
that is not an active verification owner; fieldnorm parity is routed through the
Entities owner above.

Compatibility limitation: existing deployments with any affected canonical or
preserved historical normalized identity fail migration preflight and require
explicit owner disposition. No automatic rewrite, reassignment or claim deletion
is provided. The earlier statements of a blocked seam are historical checkpoints;
the resumed tracker above is current. EM-03 remains pending until this exit passes.


## EM-02 exit

`make service-backed-test-slice OWNER=module.entities ROWS=module.entities.support_integration.active_identifier_claims_8f302d2a40,module.entities.support_integration.entity_identifier_claims_37_3ba3e2aaaf`:
PASS 4/4 at `.cartulary/test-results/20260910T215746Z-p92932`. A preceding build
failure at `215641Z-p75387` was the new test's missing time import, now fixed.
The corpus, broad Go/PostgreSQL Unicode parity, compatible/no-rewrite migration,
incompatible fail-closed migration, both claim classes and immutable UI review
exit criteria pass. Changed paths are the normalization/fixture/generated,
planner/model/controller/inspector, authority foundation and authored routing
areas enumerated above. Transport retention and reconciliation are deliberately
still open G3-G5 work. Compatibility and migration limitations above remain.
EM-02 is DONE. Next action: EM-03's dedicated merge attempts and record-scoped
runtime admission, preserving existing history and autosave ownership.


## EM-03 exit

The dedicated WorkbookEntityMergeOwner retains immutable captured review, request
path/body and one secure transaction ID. Preparation, submission, uncertainty,
definitive rejection, acknowledgement and reconciliation are distinct. Exact
replay bypasses fresh loser eligibility while checking current authority. Later
replay rejection, including transaction conflict, never erases earlier uncertainty.
Bounded observation separates timeout from actual transport settlement; late
acknowledgement is retained before callbacks. Account/incident/runtime retirement
clears attempts; authentication suspension conceals them and fences callbacks.
No browser persistence, autosave merge unit or history workflow extension exists.

Runtime reservations cover both IDs synchronously. Earlier queued and direct
writes are coordinated and observed versions remain monotonic; changed versions
require review, never rebase. New conflicting queued/direct/clipboard/history
admissions fail locally, while unrelated record writes remain usable. Entity
create/upsert targets are server-resolved: an earlier same-type create is a
preparation barrier, and new same-type creates wait for pending merge recovery.
This deliberately does not guess which record an upsert will match. It adds no
global mutation queue or server lock API. The history owner receives only a narrow
record admission predicate; its existing operation and rollback behavior remains.

The new exact send adapter retains the entire generated receipt, validates pair,
incident/type, advancing versions, change set and ordered/count-checked summary,
and classifies thrown/malformed/indeterminate transport as uncertain. Safe long
normalized collision values survive presentation. The old per-call merge command,
lossy accepted projection and its types were removed; existing payload evidence
now calls the dedicated adapter. Controller confirmation delegates admitted work
to the runtime owner and preserves only fenced presentation callbacks.

Changed areas: merge owner/review/operation/controller, exact merge adapter and
protocol aliases, narrow entity write boundary, command/clipboard composition,
runtime and shell version observations, four history admission guard lines,
focused tests/test support and authored routing/source ownership. No public route,
authorization, backend merge semantic, dependency or stored-data change in EM-03.
Save-status consumes existing pending/recovery facts and labels.

Exact verification:

- `make format generate`: PASS at `221204Z-p15147` / `221204Z-p15063`, then
  `221349Z-p24078` / `221349Z-p23982` under `.cartulary/test-results`.
- `make frontend-typecheck frontend-import-boundary-check`: PASS 2/2 each at
  `221229Z-p22346` / `221229Z-p22353`, and
  `221427Z-p31358` / `221427Z-p31366`.
- Initial owner/controller/transport slice passed 3/4 at `221229Z-p22280`;
  the failing transport fixture omitted the required error status/request ID.
  It was corrected to the actual server error envelope, preserving the long-value
  assertion. Server record_locked is a definitive non-retryable rejection; no
  automatic retry was invented.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.entity_merge_admission,web.workbook.regression.entity_merge_owner,web.workbook.regression.entity_merge_transport,web.workbook.regression.workbookz_entity_merge_controller_lifecycle_7ad8b846c1`:
  PASS 5/5 at `221427Z-p31283`.
- The same command with rows
  `web.workbook.regression.history_operation_owner`,
  `web.workbook.regression.workbookz_mutation_runtime_semantic_command_ports_7b2f1d02a4`,
  `web.workbook.regression.workbookz_mutation_runtime_paste_adapter_a108000001`,
  `web.workbook.regression.workbookz_mutation_runtime_common_coalescing_ba2a71e53d`,
  `web.workbook.regression.workbookz_mutation_runtime_semantic_secure_a6d23fa13e`,
  `web.workbook.regression.workbookz_mutation_runtime_semantic_token_cutover_263d874aa1`,
  and `web.workbook.regression.workbookz_mutation_runtime_surface_continuity_97808fe8ae`:
  PASS 12/12 at `221657Z-p37742`. It covers both queued participants and direct
  writes, exact replay, timeout/late settlement, session/account concealment,
  complete receipts, failed refresh and no version regression.

EM-03 is DONE. Receipt/reconciliation primitives are exercised with controlled
ports; the production recovery panel and acceptance-aware dependent refresh are
EM-04's remaining exit. Real service-backed browser replay/rollback evidence and
final coordinated checks remain required, not waived.

## EM-04 progress

Added shell-level Merge actions, retained receipt details, explicit exact replay,
refresh-only recovery and the existing History/change-set rollback controller.
The recovery section is keyboard reachable, focuses its summary only on explicit
opening, restores its trigger on Escape and conceals protected entries when
access is unavailable. Confirmation focuses Cancel first; completion never
moves focus asynchronously. Both complete record IDs distinguish duplicate labels.

Entity, Assessment and Generic query acceptance now handles access-loss
concealment before throwing to a required-refresh caller. Merge reconciliation
uses the current entity projections plus the active dependent surface and, when
still current, the originating survivor Timeline preview. Missing, superseded or
failed refreshes retain the receipt and require explicit refresh retry. Receipt
history delegates to existing rollback ownership and requires accepted refresh.

Changed areas: WorkbookShell, WorkbookSurfacesFacade, EntityWorkbookSurface,
entity inspector/controller, WorkbookEntityMergeOwner, new recovery component
and tests, entity Timeline preview/query composition, Timeline runtime bindings,
existing shell fixtures, merge browser support and new merge-recovery.spec.ts.
Authored catalog and generated browser batch inputs route real service evidence.
The frontend guide's subpath count follows the approved machine projection.

Checks during composition:

- `make frontend-typecheck`: PASS 2/2 at
  `.cartulary/test-results/20260910T224536Z-p9913`.
- `make frontend-import-boundary-check`: PASS 2/2 at
  `.cartulary/test-results/20260910T224136Z-p91867`.
- `make lint-biome`: PASS 2/2 at
  `.cartulary/test-results/20260910T224151Z-p17511`.
- `make format generate`: PASS at
  `.cartulary/test-results/20260910T224432Z-p42472`
  (format root `20260910T224433Z-p42552`).

Earlier composition failures are retained: typecheck `223011Z-p48756` used
unsupported test matchers, fixed to repository assertions; typecheck
`222420Z-p41391` lacked a schema import and needed a string-set type. Format
`222851Z-p43516` and lint `223009Z-p48340` exposed effect-dependency and test
assertion cleanup, subsequently fixed. UI test run `223632Z-p58041` failed only
on those unsupported matchers. Shell runs `223752Z-p59359`, `224132Z-p82795`
and `224238Z-p40236` exposed incomplete old merge receipts; fixtures now supply
the full generated class entries and advancing loser version. The strict
receipt validation was retained. Shell suite PASS 2/2 units at
`.cartulary/test-results/20260910T224350Z-p41659` with:

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell_surfaces_suite_668e482b1e
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.store.explicit_merge_store_behavior_repoints_live_ment_0560122310,module.entities.integration.the_explicit_merge_route_repoints_live_fan_out_p_0ec76e8044
make service-backed-test-slice OWNER=module.revisions ROWS=module.revisions.integration.merge_aware_change_set_rollback_restores_survivo_130bd55b07
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.verify_inspector_tabs_relationship_links_evidenc_9ee9fd9ea2,module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea
```

Entities service checks PASS 4/4 at `224206Z-p18316`; Revisions PASS 3/3 at
`224118Z-p68216`; ordinary inspector accessibility/visual comparison PASS 13/13
at `224518Z-p77498`. No golden mutation was needed.

Initial browser attempt `224131Z-p81220` FAIL 10/13 units: the existing merge
helper used the retired relationship-chip accessible name; the new scenario
omitted the existing grid-scroll helper before accessing a virtualized action.
Both fixture issues were corrected without weakening row/receipt/effect
assertions. New catalog authoring briefly failed before execution on unsorted
and dynamically composed selectors; literal test titles and sorted authored
selectors repaired the routing. Exact recovery browser evidence remains the
current exit dependency. No failure is waived.

## EM-04 exit

Exact service-backed recovery PASS 11/11 units at
`.cartulary/test-results/20260910T224748Z-p18756`. The final compact presentation
and Entities accessibility rerun PASS 13/13 at
`.cartulary/test-results/20260910T225031Z-p68190`:

```bash
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.entity_merge_exact_recovery
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.entity_merge_exact_recovery,module.entities.accessibility.verify_mention_chip_states_and_manual_resolution_e5964739d3
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.entity_merge_owner,web.workbook.regression.entity_merge_recovery_presentation,web.workbook.regression.entity_merge_query_concealment,web.workbook.regression.entity_merge_admission,web.workbook.regression.entity_merge_transport,web.workbook.regression.workbookz_entity_merge_controller_lifecycle_7ad8b846c1
```

The focused owner/admission/transport/controller/recovery/query slice PASS 7/7
at `224825Z-p49892`. Receipt-refresh tests additionally reject a replaced
projection callback and propagate failed acceptance to the existing History
controller; retry still sends no merge. The earlier browser rerun
`224504Z-p49876` passed the existing merge row but failed the new scenario's
wide-screen navigation setup (11/13 units). It now uses the visible direct
surface buttons or the existing narrow-screen menu as appropriate.

The four real-server cases cover both entity types and both loss positions.
For post-commit loss, route.fetch reaches the actual server before the response
is discarded; replay occurs after inspector closure and surface navigation have
removed the loser from active rows. The two request strings are byte-identical;
receipts are identical; survivor/loser/dependent versions, rows and histories do
not change on replay; only one merge change set exists. Carry-forward and alias
copy are verified separately. A deliberate projection failure retains completion;
refresh-only retry leaves the merge request count at two. Existing History
change-set rollback restores the loser, identifiers and dependent reference,
with advancing versions. Pre-dispatch loss verifies the same recovery UI without
pretending the first request reached the server.

Evidence is embedded in the final root's
browser-e2e-webserver-backed/browser-groups/functional-support-default-merge-recovery/playwright-report.json:
each case retains exact request/receipt/effect JSON and a recovery screenshot.
Decoded copies under that group's reviewed-attachments directory were inspected:
full identities, readable non-color status, compact spacing, summary table and
scrollable receipt content. All four images were reviewed; no committed golden
changed. Keyboard checks cover Cancel initial focus, recovery opening and Escape
restoration. Existing inspector visual/accessibility and entity-linking
accessibility rows passed without weakening assertions.

Compatibility: public interfaces, authorization, server merge behavior and
existing rollback are unchanged. No browser persistence was introduced.
The only backend expansion remains the authorized normalizer/migration repair.
JSON-shape check exposed its second authored migration-owner allocation map at
`tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership.mjs`;
adding migration 41 as Entities repaired the projection. Failure roots
`224850Z-p59526` (missing allocation) and `225014Z-p60995` (check started before
generation settled) are retained. After generation, JSON-shape PASS 3/3 at
`225058Z-p97898`; frontend import boundaries PASS 2/2 at `225114Z-p2179`.
EM-04 criteria pass; G1-G5 are implemented and narrowly verified. Remaining risk
is coordinated repository verification and the stated deployment compatibility
preflight. Next action is EM-05, with RESULTS_DIR unset.

## EM-05 verification in progress

`env -u RESULTS_DIR make agent-finalize`: PASS 1/1 at
`.cartulary/test-results/20260910T225223Z-p4632` and again at
`225413Z-p15618`. Both report zero generated changes. Retained-run selection,
performance-evidence maintenance and retained-run checks were skipped because
RESULTS_DIR was unset; no historical warm run was supplied.

Final narrow strengthening verifies the originating survivor remains selected,
its merge section retains focus, and grid scroll coordinates are unchanged after
acknowledgement and refresh. The existing service-backed merge row PASS 11/11
at `.cartulary/test-results/20260910T225452Z-p19493`:

```bash
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.the_browser_inspector_merges_duplicate_entities_d0cb95b021
```

Final preparation commands and results:

```bash
make frontend-typecheck frontend-import-boundary-check backend-module-boundary-check lint-biome lint-scripts lint-shell test-catalog-check
make generated-artifact-policy-check json-shape-check generate-drift toolchain-drift migration-drift
```

The first group PASS at `225532Z-p55526` (type 2/2), `p55532` (frontend imports
2/2), `p55539` (backend boundaries 3/3), `p55543` (Biome 2/2), `p55546` (scripts
2/2), `p55552` (shell 4/4), followed by successful catalog validation. The second
group PASS at `225317Z-p8139` (generated policy 3/3), `p8143` (JSON shape 3/3),
`p8135` (generation drift 4/4), `p8147` (toolchain 2/2), `p8151` (migration drift
5/5). All shortened roots in this handoff are under .cartulary/test-results
with the same 20260910T prefix shown in their section.

Typecheck failures `225317Z-p8243` and `225453Z-p19779` were test-authoring
errors: Promise<void> versus the fixture's Promise<undefined>, and an extra
argument to the existing grid-scroll selector. Both were corrected; owner test
PASS 2/2 at `225414Z-p15848`, typecheck and browser checks pass above.
`make format` PASS 2/2 at `225502Z-p36891`. No product assertion was waived.
Full `make check` is now running against the coordinated implementation.


## Coordinated changed-path inventory

The final scope includes the following authored and Make-generated artifacts.
Only this dedicated handoff is new under docs/handoffs; completed handoffs and
the advisory digest are unchanged. No lockfile, dependency version or analyst
data is modified. The three catalog/evidence hash assertions are reviewed
migration-41 projections, with their strict comparisons retained.

```text
apps/web/e2e/merge-recovery.spec.ts
apps/web/e2e/support/entities/merge.ts
apps/web/src/testing/entityMergeTestSupport.ts
apps/web/src/workbook/WorkbookShell.surfaces.test.tsx
apps/web/src/workbook/WorkbookShell.tsx
apps/web/src/workbook/adapters/createWorkbookClipboardPasteAdapter.ts
apps/web/src/workbook/adapters/createWorkbookEntityMergeAdapter.test.ts
apps/web/src/workbook/adapters/createWorkbookEntityMergeAdapter.ts
apps/web/src/workbook/adapters/entityIdentifierNormalization.ts
apps/web/src/workbook/adapters/workbookOperationErrorPolicy.ts
apps/web/src/workbook/adapters/workbookProtocolTypes.ts
apps/web/src/workbook/components/EntityWorkbookSurface.tsx
apps/web/src/workbook/features/entities/WorkbookEntityMergeOwner.test.ts
apps/web/src/workbook/features/entities/WorkbookEntityMergeOwner.ts
apps/web/src/workbook/features/entities/WorkbookEntityMergeRecovery.test.tsx
apps/web/src/workbook/features/entities/WorkbookEntityMergeRecovery.tsx
apps/web/src/workbook/features/entities/entityMergeOperation.ts
apps/web/src/workbook/features/entities/entityMergeReview.ts
apps/web/src/workbook/features/entities/useEntityMergeController.test.tsx
apps/web/src/workbook/features/entities/useEntityMergeController.ts
apps/web/src/workbook/features/entities/useEntityWorkbookInspectorComposition.tsx
apps/web/src/workbook/history/WorkbookRecordHistoryOwner.ts
apps/web/src/workbook/hooks/useEntityTimelinePreview.ts
apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts
apps/web/src/workbook/hooks/useWorkbookSurfaceQueries.ts
apps/web/src/workbook/models/entityIdentifierClasses.ts
apps/web/src/workbook/models/entityMergePlan.test.ts
apps/web/src/workbook/models/entityMergePlan.ts
apps/web/src/workbook/models/entityWorkbookModel.test.ts
apps/web/src/workbook/models/entityWorkbookModel.ts
apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.test.ts
apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.ts
apps/web/src/workbook/mutations/entityRecordWriteBoundary.ts
apps/web/src/workbook/mutations/workbookMutationCommandPorts.ts
apps/web/src/workbook/query/useAssessmentSurfaceQuery.ts
apps/web/src/workbook/query/useEntitySurfaceQuery.test.tsx
apps/web/src/workbook/query/useEntitySurfaceQuery.ts
apps/web/src/workbook/query/useGenericSurfaceQuery.ts
apps/web/src/workbook/runtime/WorkbookEntityMergeAdmission.test.ts
apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts
apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx
apps/web/src/workbook/timeline/hooks/useTimelineMutationRuntimeBindings.ts
apps/web/src/workbook/timeline/useTimelineMutationRuntimeBindings.test.tsx
contracts/entities/identifier-normalization-corpus.v1.json
contracts/entities/identifier-unicode.v1.json
contracts/entities/identifier-unicode.v1.schema.json
contracts/entities/merge-planning-corpus.v1.json
contracts/entities/merge-planning-corpus.v1.schema.json
contracts/index.json
contracts/protocol-ts/frontend-entrypoints.v2.json
db/migrations/00041_entities_identifier_unicode_v1.sql
docs/guides/cartulary_frontend_implementation_testing_guide.md
docs/handoffs/workbook-entity-merge-recovery-refactor-handoff.md
docs/spec/01_architecture_storage_and_view_contracts.md
docs/spec/02_domain_model_schema_and_history.md
internal/app/operator/operator_migration_evidence_test.go
internal/gen/contractentities/artifacts_gen.go
internal/modules/database_migrations/catalog_characterization_test.go
internal/modules/entities/active_identifier_claims_integration_test.go
internal/modules/entities/hostidentity/alias_sync.go
internal/modules/entities/hostidentity/rollbackprovider/collections.go
internal/modules/entities/incident_bundle_portable_prepare.go
internal/modules/entities/merge_unit_test.go
internal/platform/fieldnorm/entity_unicode.go
internal/platform/fieldnorm/text.go
internal/testutil/pgtest/pgtest_test.go
packages/protocol-ts/package.json
packages/protocol-ts/src/entrypoints/entities.ts
packages/protocol-ts/src/generated/entity-identifier-unicode.ts
packages/ui-contracts/src/entityEvidenceSelectors.ts
packages/ui-contracts/src/workbook-interaction-selectors.test.ts
tools/browser_e2e_batch_manifest.json
tools/database-migrations/generate-catalog-projections.mjs
tools/execution_topology_render_index.json
tools/frontend_import_boundaries.json
tools/frontend_source_ownership.json
tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership.mjs
tools/migration_history_manifest.json
tools/protocol-ts/generate-protocol-types.mjs
tools/schema_object_ownership_manifest.json
tools/schemas/cartulary.protocol_ts_frontend_entrypoints.v2.schema.json
tools/test_families/module.entities.json
tools/test_families/web.workbook.json
```

### First coordinated check and remediation

`make check`: FAIL 819/826 at
`.cartulary/test-results/20260910T225632Z-p58460` (417095ms). This failure is not
waived and is not final-source evidence. The seven failing units were:

- Database Migrations canonical catalog hash and pgtest catalog hash: the
  migration-40 expected digest required its reviewed migration-41 replacement.
- Operator migration-evidence exact-byte digest: the manifest/source audit now
  includes migration 41; the strict golden comparison remains.
- Frontend source ownership: a new controller test asserted wire fields under
  the features directory. That assertion moved to the exact transport test;
  the controller verifies the full captured review. The ownership policy and
  exact serialized-byte checks are unchanged in strength.
- Timeline runtime binding fixture: its partial runtime lacked the new merge
  refresh registration. The fixture now checks registration, acceptance-required
  refresh and cleanup on dependency change and unmount alongside prior checks.
- Network Analysis exploration-focus test: the retained result reports a
  15437ms failure with a runner stack rather than a failed product assertion.
  Standalone `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.exploration_focus`
  PASS 2/2 at `230357Z-p43441`. No Network Analysis code, assertion, timeout or
  resource policy was changed. A full passing rerun remains required.
- PostgreSQL privilege parity: migration 41 mistakenly granted direct EXECUTE
  on the private normalizer. The metadata was correct. Both Up and disposable
  Down now retain migration 40's private privileges, with explicit before/after
  assertions for PUBLIC, runtime and recovery. No authorization-policy expansion
  is retained. The initial diagnosis of a metadata mismatch was corrected after
  reading the old grants; no privilege metadata was relaxed.

The catalog hash was independently recomputed from ordered SQL filenames and
bytes with the existing runner identity and NUL separators. Its final value is
`14d9c07efb0828cf3c9af2aa973ad34c9d6f5ae0db3fcc285f0b769dd8e7ab20`.
The reviewed operator evidence digest is
`c5dc7fe1c8419140edf241d116813ea6fe4d489ff35ad1f2c33d43e31988d83b`.
Intermediate digests containing the mistaken grant were replaced, not adopted.
The corresponding failing operator characterizations are `225836Z-p16115`
and `230545Z-p55687`. Exact redaction, relocation, evidence-only posture and
non-authorization assertions remain intact.

Focused remediation PASS: Database Migrations 1/1 at `230010Z-p47896`,
pgtest 1/1 at `230045Z-p61740`, Operator 1/1 at `230021Z-p52680`, frontend
ownership 2/2 at `230011Z-p48394`, controller/transport 3/3 at
`230022Z-p53465`, Timeline bindings 2/2 at `230550Z-p66484`.
The privilege correction has a further coordinated service/browser rerun below.

The privilege-preserving correction initially exposed the underlying new-trigger
mistake: its invoker body called the private normalizer. Combined Entities/browser
run `230545Z-p55679` failed 10/13 with runtime create errors, while the ACL row
passed 3/3 at `230545Z-p55682`. The final trigger retains its prior invoker
security and applies the same frozen owner casing map inline; disposable Down
restores the prior inline casing. No helper grant or SECURITY DEFINER expansion
is introduced. All unrelated trigger checks and source error codes are retained.

Final correction commands:

```bash
make generate
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.support_integration.active_identifier_claims_8f302d2a40,module.entities.support_integration.entity_identifier_claims_37_3ba3e2aaaf,module.entities.browser.entity_merge_exact_recovery
make service-backed-test-slice OWNER=platform.postgres ROWS=platform.postgres.integration.purpose_identity_roles_and_acl
```

PASS: generation at `230912Z-p11432`, Entities/compatibility/exact browser recovery
13/13 at `230945Z-p14512`, and PostgreSQL roles/ACL 3/3 at `230945Z-p14519`.
The last operator golden characterization `230945Z-p14527` identified the final
source digest change; the strict final digests above cover the corrected trigger.
The earlier post-private-grant Operator and catalog checks passed at
`230700Z-p7098` and `230704Z-p7530`, respectively; final full verification follows.

Receipt validation delegates UUID format to the existing generated validator,
removing a redundant local pattern. An exploratory v7 fixture failed because the
current generator accepts versions 1-5; no broader UUID change is part of this
seam. The fixture again uses the currently supported generated receipt format,
while malformed IDs remain covered. The failure at `231142Z-p69265` records this
incorrect test assumption, not a product compatibility change.

## Final compatibility, rollback and limitations

The public HTTP routes, payloads, receipts, authorization/CSRF policy, server
merge locking, idempotency semantics and change-set rollback interface remain
unchanged. Source/history/claim data is not rewritten by migration 41. Existing
history recovery, inspector ownership, save-state labels, autosave FIFO and
Network Analysis behavior remain in their owners. No browser persistence or
generic mutation framework was added. Merge-only per-call transaction plumbing
was removed in favor of the retained exact attempt.

Review explains currently loaded canonical, reusable and alias state. It does
not guarantee server collision or authorization acceptance or invent dependency
counts. Recovery is limited to the existing permitted in-memory workbook lifetime;
account/incident/runtime retirement clears it. Acknowledgement is permanent for
that lifetime even if presentation detaches or a projection refresh fails.

Deployment compatibility is assessed by the migration's transactional preflight,
not by an audit of any deployed analyst database in this task. Any existing
canonical or preserved historical identifier whose identity would change causes
a safe rejection requiring explicit owner disposition. No claim is reassigned
or deleted to make the migration pass. PostgreSQL Unicode 16 is checked explicitly.

Rollback reverts the coordinated implementation, fixtures, generated artifacts
and owner clarification. Delete or rewrite no entity, history, revision,
idempotency receipt, claim or analyst data. Disposable migration Down has its own
compatibility preflight. A live rollback must retain a compatible binary/database
pair: keep the applied migration or use a reviewed forward repair if newly
admitted identities make a return to the old normalizer incompatible. Existing
History change-set rollback remains the analyst route for reversing a merge.
No commit, reset, push, deployment or further refactor is authorized or performed.

### Final gate preparation

`env -u RESULTS_DIR make agent-finalize`: PASS 1/1 at
`231324Z-p74312` (13524ms), with retained-run maintenance skipped and no generated
updates. Earlier repeats at `231144Z-p69478` also passed. Current transport and
Timeline-binding slice PASS 3/3 at `231323Z-p74072`; final Operator golden PASS
1/1 at `231148Z-p70197`.

The complete preparation command groups listed above were rerun after the final
corrections. All PASS: typecheck `231357Z-p78180` (2/2), frontend imports `p78188`
(2/2), backend boundaries `p78192` (3/3), Biome `p78200` (2/2), scripts `p78204`
(2/2), shell `p78216` (4/4), and catalog validation (command exit 0). Generated
policy `231356Z-p78075` (3/3), JSON shape `p78079` (3/3), generation drift `p78071`
(4/4), toolchain `p78083` (2/2), and migration drift `p78087` (5/5) also PASS.
The final `make check` rerun starts after these checks, with no further product,
fixture or generated-source edits planned. Only final handoff evidence and its
required Markdown/scope audit remain after the gate.

## EM-05 exit

`make check`: PASS **826/826 units** at
`.cartulary/test-results/20260910T231454Z-p87451` (374630ms). The coordinated run
has no failed units and includes all previously failing rows, including Network
Analysis, without a waiver or a changed Network Analysis assertion/budget.
Authoritative local gate evidence is run-manifest.json, run-summary.json,
target-summaries/check.json, unit-results and unit-logs under that root. This
supersedes the first failed full run as final-source verification.

All EM-01 through EM-05 exits are DONE. The five gap criteria pass: owner-faithful
planning, captured review and admission, exact retained recovery, receipt-first
completion with refresh-only reconciliation, and complete response retention.
The real service/browser evidence and reviewed screenshots are recorded in EM-04
and the final privilege-correction run. There are no unresolved implementation
failures or waived checks. The deployment compatibility limitation and permitted
in-memory recovery lifetime remain as stated above.

Final scope review matches all 83 paths in the inventory. Branch remains `main`,
HEAD remains `6ac49305dac340059ac9e17f705bc2cb491c49d7`, and the index is empty.
Only the coordinated working-tree changes are present. Existing digest/handoffs,
lockfiles, dependencies and unrelated product owners remain unchanged. No commit,
reset, push, deployment or analyst-data change was performed.

Post-tracker commands are run after this final edit:

```bash
make lint-markdown
git diff --check
git branch --show-current
git rev-parse HEAD
git diff --cached --name-only
git status --short
```

Their exact output and scope comparison are retained alongside the passing full
run as final-scope-audit.txt and final-markdown-audit.txt. No product or generated
source changed after the successful full check; Markdown remains outside product
verification inputs. Retained-run maintenance was skipped because RESULTS_DIR
was unset when finalization ran. This seam grants no authorization for another
refactor.
