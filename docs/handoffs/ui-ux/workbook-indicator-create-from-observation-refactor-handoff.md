# Workbook Indicator create-from-observation handoff

## Baseline, authority, and authorization

- Execution baseline: `main`, `6f98051c0d3f5097ee625c2894c6bc800d91fcb5`.
- Initial tracked and untracked status: clean. Existing analyst data is preserved.
- Read order: repository `AGENTS.md`, localized digest overlay, current Core
  owners, bounded design/domain guidance, then implementation and test evidence.
  `docs/research/nlspec-spec.md` is explanatory research, not product authority.
- The implementation request explicitly gates ICO-02 on owner adoption and
  verification of a separate backend prerequisite. Planning preferences select
  the proposal below; they do not themselves adopt normative owner changes.
  The user subsequently explicitly authorized adoption of the proposed Core
  clarifications and implementation of the scoped backend prerequisite.
- Allowed scope: this Timeline Relationships workflow, narrow workbook integration,
  typed projections, focused service/browser tests, authored ownership/routing
  inputs, Make-generated derivatives, reviewed visuals, this handoff, and the
  separately authorized Core clarifications and backend prerequisite. The
  prerequisite gate was completed before frontend implementation.
- Excluded: new public routes or response shapes, database migrations, dependencies,
  digest edits, Notes/Evidence entry points, capture/lifecycle/Network Analysis
  behavior changes, commits, resets, pushes, deployment, and analyst-data cleanup.
  Only the separately authorized Core 01/Core 03 clarifications and scoped
  Indicator backend prerequisite are exceptions to the original scope.

## Ordered execution tracker

Only the current row may be `IN_PROGRESS`. A blocker prevents all successor work.

| Workstream | Depends on | Status | Risk | Binary exit |
| --- | --- | --- | --- | --- |
| ICO-01 baseline, owner audit, characterization, prerequisite | None | DONE | Owner adoption, replay compatibility, concurrent history | Owner decisions adopted and backend correction authorized; status, duplicate, replay, concurrency, and history checks pass. |
| ICO-02 contextual canonical authoring | ICO-01 | DONE | Proposals mistaken for identity; forbidden request members | Contract-backed authoring, invalidation, and request tests pass. |
| ICO-03 retained creation and explicit resolution | ICO-02 | DONE | Duplicate dispatch, lost receipts, automatic linking | Independently recoverable stages pass; no implicit resolution. |
| ICO-04 reconciliation, accessibility, service/browser evidence | ICO-03 | DONE | Stale materialization, retargeting, focus and provenance loss | Required real-service/browser, accessibility, and reviewed visual scenarios pass. |
| ICO-05 final validation and handoff | ICO-04 | DONE | Incomplete evidence or scope drift | Finalize, check, final Markdown/diff/status audits pass against completed handoff bytes. |

## Confirmed gap ledger

| Gap | Owner | Remediation and affected areas | Rationale and long-term benefit | Compatibility impact | Unresolved risk | Binary completion criterion |
| --- | --- | --- | --- | --- | --- | --- |
| ICO-G01 fresh duplicate creates return HTTP 200 | Core 01 REQ-01-070 | Indicator create service and Workbook adapter return 201 for fresh success, 200 only for replay. | Makes public evidence reliable without implying record insertion. | Fresh reuse changes from 200 to 201; response body unchanged. | Existing consumers/tests may infer creation from status. | New, reused, replayed, divergent status matrix passes. |
| ICO-G02 comparison hashes use raw identity and transaction key | Core 01 REQ-01-057/061/070 | Normalize new create comparison through the Indicator owner; retain legacy exact digest reader. | Separates operation identity from semantic equality and preserves recovery. | New normalized equivalents replay; legacy exact replay remains supported; no receipt rewrite. | Legacy normalized requests cannot be reconstructed from a hash/result. | New equivalence/divergence and retained legacy fixtures pass. |
| ICO-G03 duplicate create fills absent metadata | Core 01 REQ-01-331; Core 03 REQ-03-137 | Adopt unchanged canonical reuse and remove duplicate metadata mutation. | Keeps create-only authoring from modifying a matched canonical record. | Explicit metadata on a canonical match is not applied; returned row is authoritative. | History for successful unchanged reuse must remain explicit. | Metadata/version unchanged; distinct operation receipt and target history verified. |
| ICO-G04 direct create reads replay outside its transaction and snapshots before identity ownership | Core 01 replay; Core 02 REQ-02-265 | Characterize direct-create races; serialize identical create keys and lock canonical identity before before-snapshot capture. | Gives replay and history one transactionally consistent decision. | No public shape change; competing exact requests replay. | Characterized races, lock ordering and failure atomicity now pass; unrelated transaction participants retain their existing contracts. | Exact/divergent same-key and distinct same-identity concurrency checks pass. |
| ICO-G05 persisted observations lack canonical authoring | Core 03 REQ-03-135/136/306; Core 02 REQ-02-075–078 | Add source-side contextual editor and typed constraints. | Keeps historical occurrence separate from canonical proposals. | Additive Timeline UI; pivot stays read-only. | Missing value kind, type changes, hash/IP rules. | Required identity and forbidden-field tests pass. |
| ICO-G06 generic create generates a fresh transaction per call | Core 03 REQ-03-301; Core 01 REQ-01-070 | Dedicated retained create owner/transport and narrow resolution handoff. | Exact recovery without a second mutation engine. | Additive runtime state; no persistence or FIFO insertion. | Late acceptance and authority/lifetime changes. | Receipt retention and separate replay tests pass. |
| ICO-G07 existing reconciliation covers observation/lifecycle only | Core 03 REQ-03-283/306 and collaboration version rules | Integrate create receipts into committed-record ports and identity refresh. | Prevents older receipts replacing current rows. | Additive owner contribution; no lifecycle behavior changes. | Filtered-out/deleted target and detached refresh. | HTTP/socket monotonicity and refresh-only recovery pass. |
| ICO-G08 no browser evidence for this two-stage seam | Core 01/02/03/04; Testing Harness NLSpec | Add routed real-service browser scenarios, selectors, accessible and visual states. | Proves actual committed operations and user recovery. | Harness inputs and reviewed affected goldens only. | Confusing two intended commits with duplicate execution. | Full required scenario matrix and final gates pass. |
| ICO-G09 child-version conflicts retire incident access | Core 03 recovery/continuity; Core 04 visibility | Narrow observation/create access-loss classification; keep child and target conflicts local. | Prevents unrelated receipt retirement and misleading navigation. | Actual authentication/authorization/incident visibility loss still clears protected state. | Distinguishing unavailable targets from lost incident visibility. | Real stale resolve retains canonical result; actual role loss removes protected UI. |
| ICO-G10 recovery popover extends beyond narrow viewport | Core 01 inspector composition; Core 03 accessibility | Bound canonical recovery with native inset/flex containment and use a bounded icon/count toolbar control with a full accessible name. | Keeps explicit recovery keyboard reachable without horizontal overflow. | Only the new recovery presentation changes. | Narrow-width overlap was characterized and corrected; reviewed phone recovery remains reachable. | Phone, tablet, desktop, zoom and painted-control checks pass; artifacts reviewed. |

## Authorized prerequisite decisions

The user explicitly authorized adopting this proposal and implementing its
narrowly bounded Core 01/Core 03 clarifications and Indicator backend correction.
The specification files own the adopted behavior; this section records the decision. No new public route, response
shape, dependency, receipt rewrite, or database migration was introduced.

### Decision and status matrix

1. Fresh canonical-create success returns 201 whether insertion or unchanged reuse;
   exact replay returns the original row/change set with 200.
2. New create comparisons use normalized admitted fields, excluding
   `client_txn_id`. Legacy receipts retain their original comparison; no inference
   from mutable or returned row state is permitted. Existing receipts are retained.
3. A fresh canonical match remains unchanged, including optional metadata and
   Records attribution/version. The operation still commits its own change set
   and Indicator target-history entry with equal before/after state. It adds no
   live row revision or record-change publication. Unchanged Records version and
   absence of those effects were explicitly adopted in Core 01; they are not
   inferred from tests. Exact replay adds none of these effects.
4. Current security checks precede replay. Fresh writes require an open incident;
   valid committed replay is evaluated before fresh state/concurrency checks.
5. Canonical creation and observation resolution remain two public operations with
   separate receipts, transaction identities, and explicit analyst activation.

| Request after current security checks | Adopted status | Canonical state and durable effects |
| --- | --- | --- |
| Fresh key, new canonical identity | 201 | Insert once; own change set, target history, live revision, publication, receipt. |
| Fresh key, existing active canonical identity | 201 | Return existing row unchanged, even with supplied metadata; own change set, equal-state target history, receipt; no live revision/publication. |
| Matching new-format normalized request with retained key | 200 | Original committed view, row and change set; no additional effects. |
| Matching historical exact admitted request with retained key | 200 | Original committed result, even if the old operation enriched metadata; no additional effects or receipt rewrite. |
| Divergent request with retained key | 409 `client_txn_conflict` | No committed effects. |
| Historical normalization-equivalent variant failing the legacy digest | 409 `client_txn_conflict` | No reconstruction from the result row; no committed effects. |

Fresh requests require an open incident. Exact committed replay remains eligible
under current session, membership, role and CSRF checks, including after incident
closure. Invalid or unauthorized requests must not use replay to bypass admission.

### Adopted owner text

For Core 01 REQ-01-061, append this bounded compatibility exception:

> Indicator row-create receipts written with the legacy raw admitted-request
> digest MUST retain their original exact admitted-request comparison, including
> the transaction key and submitted identity representation. New receipts MUST
> compare normalized admitted create fields, excluding `client_txn_id` and route
> scope members. Implementations MUST retain legacy comparison alongside the new
> comparison without rewriting receipts. They MUST NOT infer an original request
> from its returned Indicator or current canonical state. A historical request
> variant that fails the legacy comparison MUST fail with `client_txn_conflict`.
> This exception does not relax current request admission or security checks.

For Core 01 REQ-01-331, append the direct-create reuse rule:

> A fresh direct create matching an active same-incident canonical identity MUST
> return that Indicator unchanged, including optional metadata, attribution and
> Records version. Supplied metadata MUST NOT enrich a matched record. This is a
> distinct successful operation, not a read-only lookup: it MUST return `201` and
> commit its own change set, idempotency receipt and Indicator target-history entry
> with equal before/after snapshots and version identifiers. It MUST append no
> live record revision and publish no `record_changed` intent for that unchanged
> reuse. Exact replay MUST return `200` and the original committed result without
> additional effects. Concurrent matching transaction keys MUST commit once and
> replay that result; divergent reuse of a transaction key MUST commit no effects.
> Canonical identity ownership MUST precede before-snapshot capture. Distinct
> requests for the same canonical identity MUST converge on one Indicator while
> retaining independent operation receipts and consistent history.

For Core 03 REQ-03-137, append the user-visible distinction:

> Direct-create canonical reuse follows Core 01 REQ-01-331. The client MUST display
> the returned canonical value and MUST NOT infer record insertion from `201` or
> observation resolution from a canonical-create receipt. Creating from a
> persisted observation and subsequently resolving it use independent operations
> and receipts. Resolution requires explicit analyst activation after validated
> create acceptance.

Core 01 REQ-01-070 already requires fresh success `201`, exact replay `200`, and
divergent-key rejection. Its existing status rule was implemented. Human review
confirmed equal-state history is compatible with Core 02 REQ-02-265 before
adoption; Core 02 was not changed.

### Adopted implementation boundary and prerequisite verification

- `internal/modules/indicators/create_service.go`: retain the existing public
  operation; normalize through the Indicator owner; acquire the scoped transaction
  lock before receipt lookup, using the same database transaction; acquire the
  canonical identity lock before the before snapshot; reuse matches unchanged.
- `internal/modules/indicators/idempotency_hash.go`: preserve the legacy digest
  reader byte-for-byte and write normalized digests for new attempts. Never derive
  a request from mutable canonical metadata or silently upgrade old receipts.
- `internal/modules/indicators/application.go` and
  `internal/app/indicatorassembly/idempotency.go`: add only the narrow transaction
  receipt-read capability needed to avoid a lookup on a separate connection.
- `internal/app/workbookassembly/indicator_adapter.go`: base response status on
  replay evidence, not whether a new canonical record was inserted. Retain the
  existing response envelope and public route.
- Focused admission/hash, adapter, direct-create service and concurrency tests must
  cover the matrix, historical fixtures, metadata preservation, target snapshots,
  version/publication counts, failed-transaction atomicity and current security.
  Existing observation, lifecycle and internal participant behavior remain intact.
- Same-key races must commit once; distinct-key canonical matches must have one
  Indicator and separate receipts/history. A changed receipt format must not make
  previously committed operations unrecoverable. No frontend implementation begins
  until these exits pass following adoption and separate authorization.

## Inspection and characterization evidence

Inspected sources include:

- `apps/web/src/workbook/features/indicators/ObservationDetails.tsx`,
  `ObservationTargetPicker.tsx`, `IndicatorInspectorWorkflow.tsx`,
  `WorkbookObservationOwner.ts`, observation drafts/model/operation/transport,
  response validation, recovery and reconciliation.
- Workbook runtime, shell and shell-infrastructure integration; generic create
  controls, `genericCreateRequestBuilder.ts`,
  `createGenericMutationCommandPort.ts`, and workbook request/row adapters.
- `contracts/view-schemas/cartulary.view.indicators.v1.json`, authored workbook
  and Indicator OpenAPI inputs, generated protocol facade and type registry.
- `internal/modules/indicators/admission/create.go`, `create_service.go`,
  `idempotency_hash.go`, identity canonicalization and repository locking;
  `internal/app/workbookassembly/indicator_adapter.go`.
- Existing Indicator admission, hash, adapter, canonical identity, observation,
  lifecycle and concurrency tests; `apps/web/e2e/indicator-observations.spec.ts`
  and its service support; authored ownership and verification routing inputs.

Confirmed baseline behavior: the adapter used `Created` to select status, so a fresh
match returned 200. The replay digest included `client_txn_id`, view identity and
submitted representations. The matched-record path could fill missing optional
metadata and advance Records version; it was not necessarily read-only. Fresh
success appended a change set and target-history mutation; live revision and
publication were conditional on state change. Replay lookup preceded transaction
start, and before-snapshot capture preceded the canonical advisory lock. Those
orders established the concurrency audit concern. Direct-create race tests were
then added and run before correction; the execution log records their initial
failures and the passing prerequisite checks.

- Both initial task guides passed: `web.workbook` and `module.indicators`.
- Planning status characterization passed:
  `.cartulary/test-results/20260911T235234Z-p14060`.
- Planning admission/hash characterization passed:
  `.cartulary/test-results/20260912T000934Z-p18664`.
- These characterize existing behavior; they do not establish conformance.
- A temporary adapter assertion requiring fresh-match 201 failed as expected:
  `make test-slice OWNER=module.indicators ROWS=module.indicators.facade.consumer_required_create_signals`;
  `.cartulary/test-results/20260912T003613Z-p25240` (exit 2,
  `product/test_assertion_failure`, 0/1 execution units passed).
  `make explain-run RESULTS_DIR=.cartulary/test-results/20260912T003613Z-p25240`
  confirmed the classification. The assertion was removed with the tentative
  edits; this is characterization evidence, not a passing prerequisite check.
- After removing the tentative changes, the same routed adapter slice passed
  (1/1 execution units): `.cartulary/test-results/20260912T004445Z-p28174`.
  It confirms restoration of baseline behavior, not the proposed status rule.

## Execution log

- ICO-01: revalidated the clean recommended baseline and task guidance; created
  this tracker before implementation.
- ICO-01: initially misread implementation approval as owner adoption and drafted
  specification/backend changes. Corrected that interpretation and removed only
  these agent-authored changes, including the temporary test assertion. No user
  edits were present or removed; no reset or checkout command was used.
- ICO-01: marked `BLOCKED`. The plan explicitly requires adopted owner decisions
  and a verified prerequisite before ICO-02; the original scope separately
  excludes adopted-specification and backend-production changes without further
  authorization. The proposal above is ready for that decision. ICO-02 through
  ICO-05 remain `NOT_STARTED`.

- ICO-01 resumed: explicit user authorization received. Baseline branch/HEAD are
  unchanged, with only this prior handoff untracked. Re-ran both task guides.
  Reviewed Core 02 REQ-02-265: canonical equal-state snapshots and indexed target
  history are compatible with the authorized no-op reuse clarification.

- ICO-01 DONE: adopted the exact authorized clarifications in Core 01
  REQ-01-061/331 and Core 03 REQ-03-137. Corrected status selection, normalized
  comparison with retained legacy reader, unchanged reuse, transaction-key
  serialization and identity ownership before snapshot capture.
- Characterization before correction failed as expected at
  `.cartulary/test-results/20260912T010828Z-p36388`: equivalent concurrent create
  conflicted, distinct-key reuse lacked the equal before snapshot, and supplied
  metadata advanced the canonical version. These failures were not waived.
- Focused unit slice passed (3/3) at
  `.cartulary/test-results/20260912T011242Z-p64133`; focused service slice passed
  (5/5) at `.cartulary/test-results/20260912T011243Z-p64353`. They cover status,
  normalization, legacy exact/variant comparison, unchanged metadata, independent
  receipts, canonical concurrency, equal history and revision/publication counts.
- Rollback and internal participant regression checks passed in
  `.cartulary/test-results/20260912T011410Z-p87323`. That run's route test initially
  expected 403 after membership removal; Core 01's visibility rule requires 404
  `incident_not_found`. Corrected the assertion to the owner contract. The route
  rerun passed (3/3) at `.cartulary/test-results/20260912T011518Z-p5549`, including
  current session/CSRF/role/membership checks, closure and historical replay after
  newer materialization. `make format` passed.

- ICO-02 IN_PROGRESS: contextual editor and owner-backed typed create constraints.

- ICO-02 DONE: added `contracts/indicators/create-constraints.v1.json`, its
  protocol generator/export and workbook adapter; contextual authoring, retained
  draft store, field-level feedback and focused tests. Generic controls gained
  optional accessible labels/descriptions/invalid state without changing defaults.
  Requests reuse `buildGenericCreateRequest`; no browser normalization or dedupe
  algorithm was added. Optional presentation metadata remains blank.
- `make generate` passed at `.cartulary/test-results/20260912T011911Z-p24001`.
  The authoring slice passed (2/2) at
  `.cartulary/test-results/20260912T012203Z-p31992`, and frontend type checking
  passed (2/2) at `.cartulary/test-results/20260912T012204Z-p32260`.
  The form remains behind its pending retained-owner integration until ICO-03.

- ICO-03 IN_PROGRESS: retained operation owner, transport and explicit handoff.

- ICO-03 DONE: retained canonical owner, transport, explicit source-side UI and
  runtime lifecycle integration are implemented. Complete receipts precede
  refresh; create and observation replay retain independent attempts and bytes.
  Context closure, late acceptance, account retirement, authority loss and
  explicit-only resolution pass in the focused owner/transport/workflow slice
  (5/5), `.cartulary/test-results/20260912T014114Z-p54784`.
- Type checking passed at `.cartulary/test-results/20260912T014033Z-p53757`;
  import boundaries passed at `.cartulary/test-results/20260912T014115Z-p55049`.
  Earlier test-fixture type errors and an incorrect child-version assertion
  were corrected, not waived. Reconciliation remains ICO-04's dependent work.
- ICO-04 IN_PROGRESS: canonical identity refresh, browser recovery, accessibility
  and reviewed visuals.

- ICO-04 characterization found a prerequisite integration defect in the existing
  observation owner: `row_version_conflict` was treated as incident access loss,
  navigating to the directory and retiring the accepted canonical result. The
  real stale-resolution browser failed at
  `.cartulary/test-results/20260912T014740Z-p71590`; the independent replay case
  passed. The rejection was not waived.
- Gap ICO-G09: Core 03 transaction recovery/continuity and Core 04 incident
  visibility own this distinction. Remediation is a narrow observation-operation
  access-loss predicate shared by the new create owner and its reconciliation.
  Affected areas: observation owner, canonical owner, history refresh and focused
  unit/browser tests. Rationale: a child version conflict proves no loss of
  incident authority. Long-term benefit: local conflicts retain reviewable
  receipts. Compatibility: stale child/target failures remain local; current
  authentication, authorization and incident-not-found still retire access.
  Risk: misclassifying actual visibility loss; binary exit: real stale resolve
  retains canonical result without leaving Inspector, while genuine access loss
  still clears protected state. Observation dispatch and transition rules remain
  unchanged.

- ICO-04 evidence: reconciliation and monotonic query slices passed at
  `.cartulary/test-results/20260912T014404Z-p61561`; observation/current-access
  regression passed at `.cartulary/test-results/20260912T015410Z-p51721`.
  Full create/lost-response and stale resolution browser cases passed in their
  subsequent routed runs. The narrow recovery layout failure at
  `.cartulary/test-results/20260912T015044Z-p8620` was corrected; a11y and both
  affected ordinary visual rows passed at
  `.cartulary/test-results/20260912T015852Z-p34880`. No golden was changed.
  Retargeting test setup was corrected to use the explicit Inspector action and
  the authored System view selector, rather than assuming grid focus retargets
  an explicitly opened Inspector. No runtime continuity behavior was weakened.
- `.markdownlint-cli2.jsonc` now includes this handoff as a documentation lint
  input. This does not bind product verification, generation or runtime behavior
  to Markdown. The default globs previously omitted handoffs.

- ICO-04 DONE: late create response, retargeted draft/focus, sheet changes and
  exact replay after incident closure passed together with the new accessibility
  row (13/13 graph units) at `.cartulary/test-results/20260912T020512Z-p20173`.
  The remaining three canonical browser cases passed in
  `.cartulary/test-results/20260912T015852Z-p34880`; no failures are waived.
  Desktop/tablet/phone proposal and phone available-result images were reviewed
  under the `canonical-visual-review` folders in these roots. New recovery is
  viewport-bounded; its toolbar control passed painted reachability.
  The ordinary visual reconciliation artifact at
  `browser-e2e-visual/frontend-visual-reconciliation.json` in the earlier root
  passed and accounts for the selected observation/lifecycle goldens. No golden
  update was needed, so update mode and post-promotion repeat runs do not apply.
- ICO-05 IN_PROGRESS: finalization, broad checks and final source/scope audit.

- ICO-05 review corrected an overly shrinking recovery label visible in the
  retained phone image. The toolbar now uses a bounded fingerprint icon/count
  with a full canonical-recovery accessible name and title. Accessibility adds
  a minimum-width assertion and a dedicated recovery screenshot. This final
  presentation correction is revalidated before completion.
- `env -u RESULTS_DIR make agent-finalize` passed at
  `.cartulary/test-results/20260912T020649Z-p55055`. Retained-run-dependent actions
  were skipped because `RESULTS_DIR` was unset. No full warm retained run is
  claimed. Finalization was repeated after the presentation correction and passed at
  `.cartulary/test-results/20260912T020838Z-p63298`, before the full `make check`.

- First full `make check` failed (857/863) at
  `.cartulary/test-results/20260912T020912Z-p66897`. Related findings were five
  Biome warnings, viewport subtraction in the recovery layout, a heading-name
  readiness selector, a wire-shape assertion in the authoring test, and direct
  cross-owner publication SQL in two new service assertions. Corrections keep
  all assertions: use native inset/flex containment, separate heading identity
  assertions, place wire assertions in adapter tests, narrow return types and
  explicit fixture guards, and reuse `collaborationsupport` publication helpers.
  No policy check, assertion or limit was weakened.
- Focused reruns passed: lint/type at
  `.cartulary/test-results/20260912T021559Z-p76777` and
  `.cartulary/test-results/20260912T021604Z-p77947`; canonical unit slices at
  `.cartulary/test-results/20260912T021600Z-p77018`; architecture policies at
  `.cartulary/test-results/20260912T021621Z-p79576`; backend boundaries at
  `.cartulary/test-results/20260912T021801Z-p13617`; latest canonical accessibility
  at `.cartulary/test-results/20260912T021622Z-p79840`.
- The unchanged Network Analysis focus row also failed in the full run, then
  passed its narrow rerun at `.cartulary/test-results/20260912T021800Z-p13313`.
  Its code, test, timeout and routing are unchanged. Full-check success is still
  required; this rerun is not a waiver.
- The broader final browser slice passed (19/19) at
  `.cartulary/test-results/20260912T020913Z-p67052`: all four canonical scenarios,
  all existing Indicator lifecycle/observation browser scenarios, canonical and
  observation accessibility, and both affected ordinary visual rows. Final
  phone available/recovery images were reviewed from its `canonical-visual-review`
  folder. Canonical accessibility passed again after native layout containment.
- Protocol browser artifact reachability passed at
  `.cartulary/test-results/20260912T021312Z-p90086`; import boundaries passed at
  `.cartulary/test-results/20260912T021255Z-p85105`.

- Finalization detected stale generated topology after the test ownership/catalog
  corrections (`.cartulary/test-results/20260912T021942Z-p19397`). The direct
  diagnostic `make json-shape-check` at
  `.cartulary/test-results/20260912T022103Z-p36977` explicitly required
  `make generate`; generated inputs are refreshed through that target before
  repeating finalization. No generated output is edited by hand.
- The updated direct-create concurrency service slice passed (3/3) at
  `.cartulary/test-results/20260912T021940Z-p19158`, including the shared
  Collaboration publication assertion. New/expanded catalog claims name their
  Collaboration, Records and Revisions collaborators.
- `make generate` passed at `.cartulary/test-results/20260912T022237Z-p37689`.
  The subsequent `env -u RESULTS_DIR make agent-finalize` passed at
  `.cartulary/test-results/20260912T022327Z-p40884`, before the second full check.
  Its `unit-artifacts/finalize-summary.json` explicitly records the
  `results-dir-not-provided` skips for retained-run closure and warm-run health
  maintenance. No retained successful full warm run was supplied or claimed.
- ICO-05 DONE: the second `env -u RESULTS_DIR make check` passed all 863/863
  execution units at `.cartulary/test-results/20260912T022646Z-p45110`
  (384805 ms). The previously failing unchanged Network Analysis focus row
  also passed within this full run. No failures were waived, no policy checks
  were weakened, and no timeout increases or resource-capacity overrides were used.
  The final handoff is followed by the required Markdown, whitespace and scope
  recheck; completion is reported only after those checks pass.

## Verification and completion

All five workstreams are complete. `make agent-finalize` passed before the
unwaived full `make check`, which passed 863/863 execution units. Frontend type,
import and ownership boundaries, backend boundaries, generation/policy checks,
required browser/accessibility scenarios and affected ordinary visuals passed.
No adopted-owner contradiction or implementation blocker remains.

`RESULTS_DIR` remained unset during finalization. Retained-run closure and
warm-run performance/health maintenance were skipped with
`results-dir-not-provided`, as recorded in the finalizer artifact. No qualifying
full warm successful exact-source run was supplied to finalization. Visual
golden update mode and post-promotion repeat runs were not used because no
golden changed. Broader release/CI publication is outside this slice; no release
or performance claim is made from these local results.

The final evidence matrix below maps the required behavior to executable sources.
The execution log above records run roots; each root's `run-manifest.json`
retains the exact selected `OWNER` and `ROWS`, and its target summary records the
result. Counts in this handoff refer to harness graph execution units, including
setup and collation, rather than the number of individual assertions.

| Required behavior | Evidence source | Passing evidence root suffix |
| --- | --- | --- |
| Required identity, exact vocabularies, hash pairs, IP restrictions, omission/null handling, forbidden create members and complete response classification | `createIndicatorCreateTransport.test.ts`, Indicator admission/hash/service tests | `021600Z-p77018`, `011242Z-p64133`, `011243Z-p64353` |
| Original observation text/provenance, proposal seeding, type-change invalidation and read-only pivot | `indicatorCreateAuthoring.test.tsx`, `IndicatorCreateFromObservation.test.tsx`, canonical browser scenarios | `021600Z-p77018`, `020913Z-p67052` |
| New/reused/replayed status, unchanged explicit metadata on matches, normalized and legacy replay, divergent reuse and concurrent canonical identity | Indicator adapter/hash/route/concurrency tests | `011518Z-p5549`, `021940Z-p19158` |
| Synchronous duplicate reservation, immutable attempts, accepted receipt before refresh and separate explicit resolution | `WorkbookIndicatorCreateOwner.test.ts`, `IndicatorCreateFromObservation.test.tsx` | `021600Z-p77018` |
| Real post-commit create and resolution response loss, exact replay, two intended operations and zero duplicate effects | `indicator-canonical-create.spec.ts` independent-recovery scenario | `020913Z-p67052` |
| Real stale/rejected resolution, unavailable target, authority loss and preservation of canonical results while authorized | Canonical reuse/stale and target/authority browser scenarios | `020913Z-p67052` |
| Closing/reopening, retargeting, sheet changes, late responses, preserved draft/focus and create replay after incident closure | Canonical late/closed browser scenario and retained-owner tests | `020913Z-p67052`, `021600Z-p77018` |
| Unfiltered identity refresh, newer history/HTTP/socket fences, detached reads and mutation-free refresh retry | `indicatorCreateReconciliation.test.ts`, existing observation/query reconciliation slices | `014404Z-p61561` |
| Desktop/tablet/phone, zoom, named controls, field feedback, keyboard/focus and bounded recovery | Canonical accessibility row and reviewed attached screenshots | `021622Z-p79840` |
| Existing lifecycle, observation and affected visual regressions | Routed service browser/accessibility/visual slice and visual reconciliation artifact | `020913Z-p67052` |

All suffixes above expand under `.cartulary/test-results/20260912T`.
Browser and accessibility slices used `make service-backed-test-slice
OWNER=module.workbook` with the selected rows retained in their manifests.
Frontend slices used `make test-slice OWNER=web.workbook`; backend service slices
used `make service-backed-test-slice OWNER=module.indicators`, with the selected
rows likewise retained. No browser request was synthesized as a committed
receipt; response-loss tests allow the real service to commit before withholding
the response. Test fixture deletion/restoration and role changes affect only
isolated harness incidents and accounts.

## Implemented files and integration

- Canonical authoring under `apps/web/src/workbook/features/indicators/`:
  `IndicatorCanonicalAuthoring.tsx`, `indicatorCreateModel.ts`,
  `IndicatorCreateDraftStore.ts`, `IndicatorCreateFromObservation.tsx`,
  `IndicatorCreateContext.ts`, `IndicatorCreateOperationStatus.tsx`,
  `WorkbookIndicatorCreateRecovery.tsx`, `indicatorCreateOperation.ts`,
  `WorkbookIndicatorCreateOwner.ts`, `reconcileIndicatorCreateReceipt.ts`.
  `ObservationDetails.tsx` and `IndicatorInspectorWorkflow.tsx` expose the seam
  and return the existing admitted observation attempt for association.
- Workbook integration: `WorkbookShell.tsx`,
  `hooks/useWorkbookShellInfrastructure.ts`, `runtime/WorkbookMutationRuntime.ts`.
  Its committed-record port combines canonical, observation and lifecycle
  evidence with the shared history fence. The new workflow contains no lifecycle
  mutation logic and no observation transition dispatcher.
- Transport/protocol: `adapters/createIndicatorCreateTransport.ts`,
  `adapters/indicatorCreateProtocol.ts`, `components/GenericMutationControl.tsx`,
  `contracts/indicators/create-constraints.v1.json`,
  `tools/protocol-ts/generate-protocol-types.mjs`,
  `packages/protocol-ts/src/entrypoints/http.ts`, and its Make-generated
  `generated/core-indicator-registry.ts`. The existing generic request builder
  is reused; the generic command port cannot retain the exact original attempt
  and complete create response, so it is not used as the retained owner.
- Narrow local-conflict repair: `WorkbookObservationOwner.ts`,
  `observationOperation.ts`, `reconcileObservationReceipt.ts`. Child/target
  conflicts no longer masquerade as incident visibility loss.
- Backend prerequisite: `internal/modules/indicators/create_service.go`,
  `idempotency_hash.go`, `application.go`, `source_repository.go`,
  `internal/app/indicatorassembly/idempotency.go`, and
  `internal/app/workbookassembly/indicator_adapter.go`.
  Authorized owner text is in Core 01 REQ-01-061/331 and Core 03 REQ-03-137.
- New focused frontend evidence: `indicatorCreateAuthoring.test.tsx`,
  `WorkbookIndicatorCreateOwner.test.ts`,
  `IndicatorCreateFromObservation.test.tsx`,
  `indicatorCreateReconciliation.test.ts`,
  `adapters/createIndicatorCreateTransport.test.ts`, and
  `apps/web/src/testing/indicatorCreateTestSupport.ts`. Existing observation
  owner/reconciliation tests cover the narrow integration repair.
- Service characterization/regressions: Indicator `application_test.go`,
  `idempotency_hash_test.go`, `identity_concurrency_test.go`, `indicators_test.go`,
  `resolution_integration_test.go`, `transaction_atomicity_test.go`,
  `unit_test.go`, and Workbook assembly `indicator_adapter_test.go`.
- Browser/accessibility evidence: `apps/web/e2e/indicator-canonical-create.spec.ts`,
  `apps/web/e2e/support/workbook/indicatorCanonicalCreate.ts`, and
  `apps/web/e2e/workbook.a11y.spec.ts`.
- Authored selectors and routing: `packages/ui-contracts/src/index.ts`,
  `tools/frontend_source_ownership.json`,
  `tools/test_families/{web.workbook,module.workbook,module.indicators}.json`,
  `tools/browser_e2e_batch_manifest.json`. `make generate`/finalization maintain
  `tools/execution_topology_render_index.json`; it was not hand-edited.
  `.markdownlint-cli2.jsonc` includes the new handoff for documentation lint only.

## Compatibility and rollback

Public routes and response shapes are unchanged. Fresh canonical reuse now
returns 201, preserves all existing Indicator fields/version, and commits its
own equal-state target history. New normalized comparisons and the retained
legacy exact-request reader coexist without receipt rewrites. Legacy requests
whose submitted representation changes still conflict when the original
normalized request cannot be reconstructed.

Frontend drafts, admitted attempts, receipts and associations last for the
existing account/incident workbook runtime lifetime. They have no browser
persistence and do not enter the autosave FIFO. Closing or retargeting a panel
never dispatches resolution. Refresh retries issue reads only. Protected state
is retired with its account/incident lifetime.

Code rollback removes the contextual UI and its narrow runtime contribution;
canonical Indicators, observations, receipts, revisions, change sets, history
and analyst data must remain intact. Preserve a reader for both legacy and newly
committed normalized digests. Do not restore metadata enrichment or old status
classification without a new owner decision. No public API or stored-data
migration is required.

## Final scope and byte audit

Final checkout: `main`, HEAD
`6f98051c0d3f5097ee625c2894c6bc800d91fcb5`, unchanged from baseline.
There are 61 changed paths: 39 modified tracked files and 22 new untracked files,
with no staged changes. The implemented-file inventory above accounts for this
scope. No migration, dependency, digest, Notes/Evidence or Network Analysis path
changed. Existing capture and lifecycle behavior remain covered by regression
evidence. The observation access-loss predicate is the documented narrow
integration repair; its dispatcher and transition rules are unchanged.

The post-DONE byte audit passed: `make lint-markdown` at
`.cartulary/test-results/20260912T023421Z-p96208`
(`adhoc/lint-markdown/tool-run-summary.json`), `git diff --check`, and
`git status --porcelain=v1 --untracked-files=all` plus branch/HEAD/staged checks.
All 22 new files also passed a trailing-whitespace/final-newline audit.
The handoff is included explicitly in Markdown lint. The same checks are repeated
after recording this result, against the final handoff bytes, before reporting
completion.
No commit, reset, push, deployment or analyst-data cleanup was performed.

Next action: none unless separately authorized.
