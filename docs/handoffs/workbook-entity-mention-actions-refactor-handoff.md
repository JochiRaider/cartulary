# Workbook entity-mention actions refactor handoff

## Baseline and authority

- Execution baseline: `main`, `629361bf78f96b761a0da7aab2b1a4a67a51ce3e`;
  clean worktree, 36 commits ahead of `origin/main`.
- The repository procedure, digest read order, Core 01/02/03/04 owners,
  design/domain boundaries, Testing Harness NLSpec, and visual maintenance
  guide were inspected during planning and the execution baseline rechecked.
- The digest and completed Indicator handoff are advisory implementation evidence.
  `docs/research/nlspec-spec.md` is research, not product authority.
- The user authorized the plan, including adopting explicit collection mention
  identity/version, correcting the receipt projection, and a projection-only
  backfill migration tested exclusively on scratch databases.
- Scope excludes new routes, dependencies, Indicator semantics, bulk resolution,
  merge, other entry points, browser persistence, commits, deployment, and analyst
  database maintenance. No source data, receipts, or history may be deleted.

## Ordered execution tracker

Only the active row may be `IN_PROGRESS`. A blocked dependency stops successors.

| Workstream | Depends on | Status | Binary exit |
| --- | --- | --- | --- |
| EMA-01 baseline, owners, characterization, prerequisite | None | DONE | Public identity, receipts, upgrade and rebuild checks pass. |
| EMA-02 complete receipts and retained operations | EMA-01 | DONE | Immutable attempts, duplicate protection, replay and receipt retention pass. |
| EMA-03 action UI, candidates and reviewed creation | EMA-02 | DONE | Transition, paging, authoring and separate create/link recovery pass. |
| EMA-04 reconciliation, security, accessibility and evidence | EMA-03 | DONE | Ordering, navigation, security, browser and reviewed visual evidence pass. |
| EMA-05 final validation and completed handoff | EMA-04 | DONE | Finalize, final gates, Markdown and scope review pass. |

## Gap ledger

| Gap | Authority | Remediation and affected areas | Rationale and long-term benefit | Compatibility | Unresolved risk | Binary validation |
| --- | --- | --- | --- | --- | --- | --- |
| EMA-G01 opaque selector parsing | Core 01 REQ-01-196–210, collection identity and REQ-01-313 | Explicit mention ID in Timeline projection, schema and chip models; derived-data backfill. | Separates mention, source, target and collection identities. | Additive response metadata; coordinated backend/frontend upgrade. | Existing projection bindings must be verified. | Opaque-selector, upgrade and rebuild tests pass. |
| EMA-G02 incomplete receipts | Core 01 REQ-01-222–227 | Repair OpenAPI and decode/retain complete authoritative receipts. | Preserves committed evidence independently of current presentation. | Schema tightens to adopted behavior; no receipt rewrite. | Malformed success cannot imply rejection. | Presence, nullability, identity and uncertain-success tests pass. |
| EMA-G03 unresolved dismissal absent | Core 01 REQ-01-213–219; Core 03 REQ-03-129–134 | State-derived controls and action-discriminated requests. | One legal transition model for all mention actions. | Adds missing Dismiss; restore never relinks. | Stale state at activation. | Complete legal/illegal matrix passes. |
| EMA-G04 transient attempts and receipts | Core 03 REQ-03-095–100, REQ-03-301; authorized runtime lifetime | Timeline owner retained by workbook runtime. | Exact replay and recovery survive presentation changes. | Memory-only; explicit operations remain outside autosave capacity. | Prior-save ordering and late acceptance. | Duplicate, exact-replay, detached and refresh-only tests pass. |
| EMA-G05 candidates use sheet rows | Core 01 view-query contract | Independent authorized paged reader and shared picker. | Sheet filtering cannot hide all eligible targets. | No new query predicate or endpoint. | Paging is live, not a snapshot. | Independent filtering, pagination and honest states pass. |
| EMA-G06 identity guesses and lost partial creation | Core 01 REQ-01-206; Core 02 §§6–8; Core 03 §9 | Reviewed ordinary create followed by independently retained resolution. | No punctuation-defined identity or repeated accepted creation. | Keeps Host/Identity upsert semantics. | Stale subject after create. | Real-service partial-success recovery creates once. |
| EMA-G07 stale metadata, undo and access handling | Core 02 REQ-02-039–042; Core 03 collaboration/undo; Core 04 authorization | Monotonic source/mention reconciliation, scoped failures and current focus bindings. | Cleared metadata stays cleared; local failures do not retire incident recovery. | Dismissed entries are explicitly session-observed. | No durable dismissed-resource read exists. | Both orderings, role/local failures and stale undo tests pass. |
| EMA-G08 integrated evidence absent | Applicable product owners; Testing Harness NLSpec | Authored routing, focused browser/a11y tests and reviewed affected goldens. | Reproducible evidence for actual committed operations. | Existing semantic selectors and renderer profile. | Visual promotion must follow review. | Required matrix and final gates pass. |

## EMA-01 evidence and disposition

- Planning characterization passed: hook ownership and chip/schema slices at
  `.cartulary/test-results/20260912T035518Z-p17631` and
  `.cartulary/test-results/20260912T035518Z-p17637`; action planning at
  `.cartulary/test-results/20260912T040008Z-p20999`; mention route and ordinary
  Host/Identity create/upsert service slice at
  `.cartulary/test-results/20260912T040008Z-p21000`.
- `make task-guide ROLE=module-author OWNER=web.workbook`, `module.entities`,
  `module.timeline`, `module.projections`, and `module.database_migrations`
  passed during planning. Execution work begins from the same source.
- Authorized prerequisite implemented in Core 01 REQ-01-313, authored Entities
  receipt and Timeline collection schemas, Timeline `collection_facts.go`, and
  `db/migrations/00042_timeline_mention_public_identity.sql`. The migration
  validates source/field/type/version/selector bindings and preserves JSON order
  and metadata. Down preserves the additive projection data.
- Scratch upgrade evidence: valid upgrade equals the current projection owner;
  selector, version, identity and cross-source corruption reject atomically;
  scratch rollback/reapplication preserves source, mentions, links, receipts and
  history. Fresh current databases and a service-backed rebuild also pass.
- `make generate` initially failed at
  `.cartulary/test-results/20260912T041102Z-p41838` because the migration needed
  source ownership. Added Timeline ownership to the authored catalog generator.
  The next run at `20260912T041254Z-p43347` stopped for 17 unreviewed OpenAPI
  differences. `make openapi-compatibility-check` recorded them at
  `.cartulary/test-results/20260912T041309Z-p44867`: 14 additive and 3 breaking
  schema classifications. Reviewed entries were added to the evolving 2.0.0
  change set; the frozen release snapshot was not edited.
- `make generate` passed at
  `.cartulary/test-results/20260912T041355Z-p49873`, including compatibility and
  generated protocol/topology/catalog derivatives. `make format` passed at
  `.cartulary/test-results/20260912T041309Z-p45007`.
- `make test-slice OWNER=module.entities
  ROWS=module.entities.unit.openapi_mutation_contract_d0fe61dbba` passed (1/1)
  at `.cartulary/test-results/20260912T041420Z-p53162`.
- `make test-slice OWNER=module.timeline
  ROWS=module.timeline.unit.timeline_projection_contract_shaping_exposes_the_0a2232ebe0`
  passed (1/1) at `.cartulary/test-results/20260912T041420Z-p53170`.
- `make service-backed-test-slice OWNER=module.entities
  ROWS=module.entities.support_integration.timeline_mention_projection_upgrade_preserves_so_dabe8cb065,module.entities.integration.the_mention_resolve_route_persists_durable_resol_d68c1befd6,module.entities.integration.host_and_identity_create_routes_reuse_exact_matc_f443e2591f`
  passed (4/4) at `.cartulary/test-results/20260912T041420Z-p53193`.
- `make service-backed-test-slice OWNER=module.timeline
  ROWS=module.timeline.integration.timeline_patch_and_query_routes_auto_resolve_onl_8ab2448829`
  passed (3/3) at `.cartulary/test-results/20260912T041428Z-p64773`.
- `make migration-drift` passed (5/5) at
  `.cartulary/test-results/20260912T041420Z-p53143`.
- Disposition: EMA-01 DONE. No additional normative/backend prerequisite found.
  Coordinated upgrade remains required; no analyst database was migrated.
  Next action: EMA-02 immutable receipt/attempt owner and runtime integration.

## EMA-02 evidence and disposition

- Added `timelineMentionOperationModel.ts`, `timelineMentionProtocol.ts`,
  `WorkbookTimelineMentionOperationOwner.ts`, and `timelineMentionOwnerFor.ts`.
  Timeline constructs the concrete owner; `WorkbookMutationRuntime` retains its
  narrow lifecycle bridge, observes source versions, accounts explicit status,
  coordinates source action blocking and retires/suspends recovery.
- Immutable route/body/key/actor/intent capture precedes the save barrier. Receipt
  acceptance precedes presentation/refresh handling. Exact replay preserves the
  original request and uncertainty survives replay rejection. Accepted metadata
  and source/mention high-water marks are retained independently.
- Added authored frontend ownership and two routed regression rows covering 13
  tests: transitions/roles, duplicates, preparation changes, exact replay,
  late acceptance, both orderings, detached refresh, scoped failures, account
  retirement, runtime accounting and complete receipt validation.
- `make format` first rejected unsorted authored test titles; corrected their
  ordering. A subsequent format/typecheck reported one implicit attempt type;
  made it explicit. Format passed at
  `.cartulary/test-results/20260912T042520Z-p10753`.
- `make generate` passed at
  `.cartulary/test-results/20260912T042424Z-p6307` after the authored routing fix.
- `make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.timeline_mention_operation_recovery,web.workbook.regression.timeline_mention_complete_receipts`
  passed (3/3 units) at `.cartulary/test-results/20260912T042448Z-p9490`.
- `make frontend-typecheck` passed (2/2) at
  `.cartulary/test-results/20260912T042531Z-p15067`; the preceding failed run was
  `.cartulary/test-results/20260912T042448Z-p9562`.
- Disposition: EMA-02 DONE. Recovery is memory-only and bounded by the existing
  incident/account runtime. Next action: EMA-03 wires this owner into mention
  presentation, independent candidates and reviewed ordinary create/link.

## EMA-03 evidence and disposition

- Replaced the component-local action path with retained owner operations and explicit
  public identity. Added unresolved Dismiss, cleared restore presentation, independent
  100-row candidate pages, shared single picker, and reviewed ordinary create/link.
- Added `timelineMentionCreationModel.ts`, candidate port/reader/hook, source reader,
  action controls and shell recovery. Creation receipt is retained before linking;
  replay/navigation tests prove accepted creation is never repeated.
- Authored ownership and routing updated; `make generate` passed at
  `.cartulary/test-results/20260912T045716Z-p34734` and
  `.cartulary/test-results/20260912T045847Z-p43157` (see retained run manifests).
- `make test-slice OWNER=web.workbook` with mention operation, complete receipt and
  hook rows passed 4/4 at `20260912T045743Z-p42184`. Candidate/creation review and
  operation rows passed 3/3 at `20260912T050021Z-p52988`.
- `make test-slice OWNER=module.timeline` with collection inspection and action-port
  rows passed 3/3 at `20260912T050155Z-p60500`; action-plan row passed 2/2 at
  `20260912T045903Z-p46500`. Entities chip row passed 2/2 at
  `20260912T050037Z-p57879`.
- `make frontend-typecheck` passed 2/2 at `20260912T050037Z-p58041`;
  `make format` passed at `20260912T050154Z-p60299`.
- Initial lint/typecheck failures were corrected. Candidate default-sort assertion
  and stale dismissed-metadata assertions were corrected to adopted behavior;
  failed runs: `20260912T045903Z-p46498`, `20260912T045903Z-p46526`,
  `20260912T050037Z-p57861`. An unsupported Biome flag override was rejected by
  public input validation; ordinary Make targets were used afterward.
- Disposition: EMA-03 DONE. Auto-resolution Undo integration still needs the
  EMA-04 refresh/focus reconciliation checks; its failed characterization roots
  are `20260912T050037Z-p57872` and `20260912T050155Z-p60498`. Next action:
  monotonic reconciliation, continuity, authorization and real-service evidence.

## EMA-04 evidence and disposition

- Reconciliation now combines complete HTTP receipts, source high-water marks,
  projected mention versions and socket-driven reads. A cancelled presentation
  refresh is accepted only when the actual visible source already meets the
  required version. Other read failures retain “completed; refresh required”.
- Removed stale cached target/provenance fallbacks; session-dismissed entries are
  based on accepted cleared metadata. Auto-resolution notices bind the current
  mention ID, version, method and target. Same-source version changes preserve
  selection while changed mention intent invalidates creation/target review.
- Focus restoration requires the original presentation, selected mention and
  authorization generation. Local target/mention failures remain local; current
  session/incident authorization is re-read for authorization failures. Runtime
  suspension hides protected presentation; access loss/account retirement clears it.
- Added real-service `mentions.recovery.spec.ts`, routed as
  `module.entities.browser.timeline_mention_creation_recovery`: creation commits,
  then resolution receives a real version rejection or commits with a lost HTTP
  response. Navigation retains the original operation. Recovery verifies one
  entity/create, exact resolution replay, and accepted action plus failed refresh
  followed by read-only recovery. It also checks 768px bounds and keyboard access.
- Extended existing accessibility evidence with unresolved Dismiss and explicit
  Display-name-only creation review. Existing resolve/create, dismissal/restore,
  automatic Undo, stateful lifecycle and responsive collection checks passed.
- `make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.timeline_mention_operation_recovery,web.workbook.regression.timeline_mention_complete_receipts,web.workbook.regression.timeline_mention_candidate_creation_review,web.workbook.regression.timeline_mention_reconciliation,web.workbook.regression.timeline_mention_auto_resolution_undo_ownership_9e834f7b55,web.workbook.regression.mention_undo_current_committed_identity_ddf69011e7`
  passed 7/7 at `.cartulary/test-results/20260912T053338Z-p93878`.
- `make service-backed-test-slice OWNER=module.entities
  ROWS=module.entities.browser.timeline_mention_creation_recovery` passed 11/11
  at `.cartulary/test-results/20260912T053012Z-p14845`. The retained screenshots
  exposed narrow recovery clipping, which was corrected with viewport positioning;
  final bounds/Escape revalidation is recorded below.
- `make service-backed-test-slice OWNER=module.entities` with the existing resolve
  and dismiss/restore rows passed those rows at `20260912T051548Z-p49104`;
  automatic Undo passed at `20260912T051156Z-p32056`. Stateful lifecycle and
  accessibility rows passed at `20260912T052128Z-p62731`; accessibility passed
  again at `20260912T053338Z-p93877`. Those aggregate runs failed only the new
  recovery row during development. The retained reports identify each row.
- `make service-backed-test-slice OWNER=module.timeline
  ROWS=module.timeline.browser.collection_chips_disclose_exact_members_without_9dfb945466`
  passed 11/11 at `.cartulary/test-results/20260912T051327Z-p74320`.
- Initial browser failures identified stale source-selection reset, cleared restore
  presentation, concurrent socket-refresh cancellation, form-label fixture drift,
  premature test teardown, and use of the System views menu for built-in Hosts.
  Each was corrected and rerun. Narrow recovery Escape needed a focusable region.
  Failed roots include `20260912T051156Z-p32056`, `20260912T051327Z-p74325`,
  `20260912T051548Z-p49104`, `20260912T052128Z-p62731`,
  `20260912T052421Z-p48179`, and `20260912T053338Z-p93877`.
- Import-boundary checks identified protocol imports outside adapters and a
  type-import cycle. Fixed by using the existing Workbook protocol type facade
  and direct transport-port dependencies. `make frontend-import-boundary-check`
  passed at `20260912T053012Z-p14935`; typecheck passed at
  `20260912T053012Z-p14927`.
- Broad characterization found migration catalog fingerprints still naming the
  prior catalog. Updated only expected hashes in
  `internal/modules/database_migrations/catalog_characterization_test.go`,
  `internal/testutil/pgtest/pgtest_test.go`, and
  `internal/app/operator/operator_migration_evidence_test.go` for authorized
  migration 42. Narrow database-migration rows passed 2/2 at
  `20260912T053130Z-p78850`; operator transport passed 1/1 at
  `20260912T053314Z-p89259`. No new product prerequisite was introduced.

### Visual review and refresh record

- Accepted trigger: adopted mention transitions and cleared resolution metadata
  now have explicit controls and retained recovery presentation. Inspector captures
  are refreshed against the validated integrated layout. Reviewed adjacent
  positioning changes preserve typography, focus, scroll framing and content.
- Ordinary reconciliation at
  `.cartulary/test-results/20260912T052009Z-p25969/browser-e2e-visual/frontend-visual-reconciliation.json`
  accounted for 221 active captures/goldens and 28 registered fixtures, with zero
  missing, orphan, ambiguous or unresolved mappings. Only screenshot comparisons
  failed. All seven actual/diff pairs were inspected before promotion.
- Owner row `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`:
  `visual.fixture.mention_chip_state_matrix`,
  `entity-mention-chip-states-linux.png`.
- Owner row `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`:
  `visual.fixture.inspector_compact_actions` →
  `workbook-inspector-compact-actions-linux.png`;
  `visual.fixture.destructive_actions` →
  `workbook-inspector-destructive-confirmation-linux.png`;
  `visual.fixture.inspector_narrow_technical_details` →
  `workbook-inspector-narrow-technical-details-linux.png`;
  `visual.fixture.base_inspector` → `workbook-inspector-history-linux.png`,
  `workbook-inspector-rollback-preview-linux.png`, and
  `workbook-inspector-public-error-linux.png`.
- Viewports, zoom, masks, scroll normalization, screenshot scopes, renderer and
  fonts were unchanged. `make browser-e2e-visual-update` passed 12/12 at
  `.cartulary/test-results/20260912T053057Z-p48762`. Its generated manifest and
  exactly seven changed PNGs were reviewed after promotion. The six adjacent
  inspector PNGs match the already reviewed ordinary actual bytes; the promoted
  mention PNG was inspected separately. No clipped controls or stale metadata
  were accepted. Two fresh ordinary validations follow in EMA-05.
- Final recovery slice passed 11/11 at
  `.cartulary/test-results/20260912T053541Z-p40976`, including viewport bounds,
  Escape/trigger focus, real rejection, lost response, exact replay and read-only
  refresh. Corrected recovery screenshots were reviewed.
- Disposition: EMA-04 DONE. No blocker remains. Next action: EMA-05 completes
  broad validation, the two ordinary visual runs and the final scope/handoff.

### Final-validation follow-up

A stronger navigation assertion at `20260912T053838Z-p48016` found that the
created entity receipt was retained but the separately loaded Hosts query could
remain empty. EMA-04 was reopened before final completion. Added retained creation
projection refresh and read-only recovery, independently of resolution, through
the existing entity-query owner. This is frontend integration within the
authorized seam, not a new backend prerequisite.

- `WorkbookTimelineMentionOperationOwner` retains creation refresh state separately
  from its receipt and the link attempt. The shell registers the existing entity
  projection refresh callback; failure exposes **Refresh created entity** and
  never repeats creation or resolution. Fixture composition uses its existing
  entity refresh port. Focus restoration waits for both entity and mention
  refresh work to settle and remains fenced by presentation/authorization.
- The real-service recovery row now aborts entity-sheet refresh, retries reads,
  verifies the entity appears in Hosts, then independently replays the uncertain
  mention and retries source refresh. All mutation counts remain exact.
- Creation/owner tests passed 3/3 at `20260912T054256Z-p11302`; typecheck passed
  at `20260912T054258Z-p15040`. The initial integrated focus check failed after
  adding the entity read (`20260912T054331Z-p30863`); recovery and accessibility
  rows passed in that run. The focus ordering correction and both service rows
  passed 13/13 at `20260912T054626Z-p95422`.
- Hook/owner tests including the new dual-refresh focus regression passed 3/3 at
  `20260912T054723Z-p32333`. Authored catalog titles were updated; `make format`
  passed at `20260912T054722Z-p32133` and `make generate` passed at
  `20260912T054722Z-p32053`.
- Projection query/rebuild owner checks passed 5/5:
  `make service-backed-test-slice OWNER=module.projections
  ROWS=module.projections.rebuild.atomic_catalog_iteration3,module.projections.rebuild.restore_contract_206379110d,module.projections.query.contract_shape_and_keyset_59528aa56d`
  at `.cartulary/test-results/20260912T054436Z-p72461`.
- `make browser-e2e-visual` passed 12/12 twice against the promoted manifest:
  `.cartulary/test-results/20260912T053540Z-p40816` and
  `.cartulary/test-results/20260912T054043Z-p9204`. Neither run updated goldens.
  The final creation-refresh controls are covered by the real-service screenshots;
  these controls do not appear in the seven changed golden states.
- Reopened EMA-04 final aggregate exit passed 21/21 at
  `.cartulary/test-results/20260912T054754Z-p39967`:
  `make service-backed-test-slice OWNER=module.entities
  ROWS=module.entities.browser.the_browser_inspector_dismisses_a_mention_and_re_6150fd43cd,module.entities.browser.the_browser_inspector_resolves_existing_entities_325548131d,module.entities.browser.the_browser_workbook_shows_auto_resolution_only_758d41a16f,module.entities.browser.timeline_mention_creation_recovery,module.entities.browser_stateful.verify_manual_mention_resolution_dismissal_auto_dfa355e592,module.entities.accessibility.verify_mention_chip_states_and_manual_resolution_e5964739d3`.
  Disposition: EMA-04 DONE again, with entity-sheet refresh included. EMA-05
  resumes final broad validation and documentation; no dependency is blocked.

### Test ownership cleanup

Final scope review retired five unrouted component-local mention action fixtures
from `WorkbookShell.support.test.tsx` and their unused helpers. They depended on
incomplete receipts, displayed-sheet candidates, one-click inferred creation,
and component-local dismissed retention. Their continuity and ordering cases are
covered by the authored retained-owner/hook rows and real-service resolve,
lifecycle and recovery rows above. The routed integrated auto-resolution Undo
fixture remains and uses complete receipts. No active catalog row was removed.

## Compatibility and rollback

Future rollout must stop old projection writers, apply the derived-data backfill,
then serve matching backend/frontend versions. This task does not deploy or apply
migrations to analyst databases. Existing clients may ignore additive identity
metadata; new clients must not parse opaque selectors when identity is missing.

Frontend rollback changes available controls only. It does not reverse committed
entities, mentions, links, receipts or history. Preserve those records and use
ordinary owner-authorized recovery actions for any requested data correction.

## EMA-05 final verification

- `make agent-finalize` ran before broad final checks and passed at
  `.cartulary/test-results/20260912T054755Z-p40178` (earlier successful roots:
  `20260912T052315Z-p3768`, `20260912T053428Z-p28727`, and
  `20260912T054438Z-p72841`). `RESULTS_DIR` remained unset. Retained-run
  selection, performance evidence, canonical run checks and scheduler maintenance
  requiring full warm evidence were skipped, as recorded in
  `unit-artifacts/finalize-summary.json`. No retained-run validation is claimed.
- The first `make check` failed 861/867 at `20260912T052344Z-p8306`: protocol
  boundaries and migration catalog fingerprints were corrected as recorded above.
  The next run failed 866/867 at `20260912T053743Z-p7338`, solely on unrelated
  `web.networkflow.regression.exploration_focus` (15.3s test duration). Its
  isolated `make test-slice OWNER=web.networkflow
  ROWS=web.networkflow.regression.exploration_focus` passed 2/2 at
  `20260912T054547Z-p94569`; no Network Flow source was changed.
- Final frontend typecheck/import-boundary/Biome checks passed at
  `20260912T054901Z-p79867`, `20260912T054901Z-p79871`, and
  `20260912T054901Z-p79877`. Removing obsolete fixtures subsequently exposed two
  unused declarations (`20260912T055138Z-p28466`); both were removed and the
  final rerun is recorded below.
- Final focused mention rows passed 7/7 at `20260912T055139Z-p29056` using the
  six web.workbook row IDs listed in EMA-04. New creation refresh/focus coverage
  is included. The successful final broad check is recorded below.
- No deployment, analyst-database migration, commit, push, dependency installation
  or retained-output deletion occurred. `main` remains at
  `629361bf78f96b761a0da7aab2b1a4a67a51ce3e`, 36 commits ahead of origin/main.
- Skipped intentionally: release/deployment targets, the full browser measurement
  suite and unrelated optional browser scenarios. Required affected functional,
  stateful, accessibility, responsive and full visual rows were executed. These
  artifacts remain implementation evidence, not new conformance claims.

The repeated ordinary `make check` at `20260912T054952Z-p81803` also finished
866/867 with only the same unrelated Network Flow focus failure. The public
`make explain-target TARGET=check DETAIL=summary` identifies
`CARTULARY_HARNESS_CAPACITY_OVERRIDE` as a supported input. The final retry uses
`make check CARTULARY_HARNESS_CAPACITY_OVERRIDE=.cartulary/mention-final-capacity.json`,
where the retained local input is exactly
`{"schema_id":"cartulary.harness_capacity_override.v1","cpu_tokens":8}`.
It reduces scheduling concurrency without changing selected rows, assertions,
time limits or production source. The full retry passed 867/867 at
`.cartulary/test-results/20260912T055634Z-p49873` (540815ms). This run is not
used as canonical performance evidence.

### Final verification commands

| Command | Result and run root under `.cartulary/test-results/` |
| --- | --- |
| `make agent-finalize` (RESULTS_DIR unset) | PASS 1/1, `20260912T055308Z-p64288`; retained-run maintenance skipped. |
| `make frontend-typecheck` | PASS 2/2, `20260912T055426Z-p43299`. |
| `make frontend-import-boundary-check` | PASS 2/2, `20260912T054901Z-p79871`. |
| `make lint-biome` | PASS 2/2, `20260912T055426Z-p43310`. |
| `make generate-drift` | PASS 4/4, `20260912T055234Z-p49774`. |
| `make generated-artifact-policy-check` | PASS 3/3, `20260912T055234Z-p49776`. |
| `make json-shape-check` | PASS 3/3, `20260912T055234Z-p49778`. |
| `make openapi-compatibility-check` | PASS 4/4, `20260912T055234Z-p49784`. |
| `make migration-drift` | PASS 5/5, `20260912T055234Z-p49782`. |
| `make lint-markdown` | PASS on the completed handoff, `20260912T060629Z-p97703`. |
| `git diff --check` | PASS after the completed handoff and scope review. |

The two post-cleanup typecheck attempts at `20260912T055138Z-p28466` and
`20260912T055250Z-p58626` read the now-removed unused declarations. The final
frontend typecheck and Biome pass above use the cleaned source.

### Final gap dispositions

| Gap | Disposition | Remaining boundary |
| --- | --- | --- |
| EMA-G01 | Resolved: public identity, verified backfill and rebuild. | Coordinated writer/backend/frontend upgrade. |
| EMA-G02 | Resolved: complete receipt schemas and strict decoding. | Malformed success remains uncertain until exact replay reconciles it. |
| EMA-G03 | Resolved: complete state-derived transition matrix. | Server authorization/version checks remain decisive. |
| EMA-G04 | Resolved: retained immutable operations and exact replay. | Incident/account runtime lifetime; no browser persistence. |
| EMA-G05 | Resolved: independent contract-filtered cursor pages. | Label filtering covers loaded candidates; paging is live. |
| EMA-G06 | Resolved: reviewed ordinary create, separate link and independent entity/source refresh. | A saved entity survives link rejection or uncertainty. |
| EMA-G07 | Resolved: monotonic reconciliation, cleared metadata, scoped authorization and focus. | Session-dismissed observations are not a durable inventory; History remains available. |
| EMA-G08 | Resolved: focused/service checks, reviewed visuals and full broad validation pass. | Default concurrency exposed an unrelated Network Flow focus timeout; the same full target passed with 8 CPU tokens. |

### Changed paths

The following inventory includes authored source/tests/contracts, generated
projections and the reviewed golden updates. Generated files were produced only
through public Make targets.


- `apps/web/e2e/autoresolve.spec.ts`
- `apps/web/e2e/mention-lifecycle.spec.ts`
- `apps/web/e2e/mentions.lifecycle.spec.ts`
- `apps/web/e2e/mentions.recovery.spec.ts`
- `apps/web/e2e/mentions.resolve.spec.ts`
- `apps/web/e2e/support/entities/mentions.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/entity-mention-chip-states-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-compact-actions-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-destructive-confirmation-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-history-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-narrow-technical-details-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-public-error-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-rollback-preview-linux.png`
- `apps/web/src/testing/TimelineWorkbookRuntimeFixture.tsx`
- `apps/web/src/testing/timelineMentionTestSupport.ts`
- `apps/web/src/workbook/WorkbookShell.mentionChips.test.ts`
- `apps/web/src/workbook/WorkbookShell.support.test.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/adapters/workbookProtocolTypes.ts`
- `apps/web/src/workbook/components/WorkbookRecordCandidatePicker.tsx`
- `apps/web/src/workbook/components/WorkbookRelationshipChip.test.tsx`
- `apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts`
- `apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts`
- `apps/web/src/workbook/timeline/actions/TimelineMentionCandidatePort.ts`
- `apps/web/src/workbook/timeline/actions/TimelineMentionRecovery.tsx`
- `apps/web/src/workbook/timeline/actions/WorkbookTimelineMentionOperationOwner.test.ts`
- `apps/web/src/workbook/timeline/actions/WorkbookTimelineMentionOperationOwner.ts`
- `apps/web/src/workbook/timeline/actions/reconcileTimelineMentionReceipt.test.ts`
- `apps/web/src/workbook/timeline/actions/reconcileTimelineMentionReceipt.ts`
- `apps/web/src/workbook/timeline/actions/timelineMentionAuthority.ts`
- `apps/web/src/workbook/timeline/actions/timelineMentionCandidates.test.tsx`
- `apps/web/src/workbook/timeline/actions/timelineMentionCreationModel.ts`
- `apps/web/src/workbook/timeline/actions/timelineMentionOperationModel.ts`
- `apps/web/src/workbook/timeline/actions/timelineMentionOwnerFor.ts`
- `apps/web/src/workbook/timeline/actions/useTimelineMentionCandidates.ts`
- `apps/web/src/workbook/timeline/adapters/createTimelineActionAdapters.test.ts`
- `apps/web/src/workbook/timeline/adapters/createTimelineMentionCandidateReader.ts`
- `apps/web/src/workbook/timeline/adapters/createTimelineMentionEntityCreationAdapter.ts`
- `apps/web/src/workbook/timeline/adapters/createTimelineMentionResolutionAdapter.ts`
- `apps/web/src/workbook/timeline/adapters/createTimelineMentionSourceReader.ts`
- `apps/web/src/workbook/timeline/adapters/timelineMentionProtocol.test.ts`
- `apps/web/src/workbook/timeline/adapters/timelineMentionProtocol.ts`
- `apps/web/src/workbook/timeline/components/TimelineCollectionCell.test.tsx`
- `apps/web/src/workbook/timeline/components/TimelineMentionActionControls.tsx`
- `apps/web/src/workbook/timeline/components/TimelineMentionsPanel.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookInspector.tsx`
- `apps/web/src/workbook/timeline/composition/useTimelineInspectorStateComposition.ts`
- `apps/web/src/workbook/timeline/composition/useTimelineInspectorWorkflowComposition.ts`
- `apps/web/src/workbook/timeline/composition/useTimelineMutationComposition.ts`
- `apps/web/src/workbook/timeline/composition/useTimelineSurfaceFoundation.ts`
- `apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineCommittedRows.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineMentionActions.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineMentions.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineRowsLoader.ts`
- `apps/web/src/workbook/timeline/models/timelineAcceptedMutationEffects.ts`
- `apps/web/src/workbook/timeline/models/timelineMentionActionPlan.test.ts`
- `apps/web/src/workbook/timeline/models/timelineMentionActionPlan.ts`
- `apps/web/src/workbook/timeline/models/timelineMutationModels.test.ts`
- `apps/web/src/workbook/timeline/models/workbookMentionChips.ts`
- `apps/web/src/workbook/timeline/mutations/useTimelineRowMutationCoordinator.test.tsx`
- `apps/web/src/workbook/timeline/mutations/useTimelineRowMutationCoordinator.ts`
- `apps/web/src/workbook/timeline/ports/TimelineMentionPort.ts`
- `apps/web/src/workbook/timeline/presentation/TimelineWorkbookInspectorRegion.tsx`
- `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- `apps/web/src/workbook/timeline/useTimelineCompositionLifecycle.test.tsx`
- `apps/web/src/workbook/timeline/useTimelineMentionActions.test.tsx`
- `contracts/openapi-releases/2.0.0.change-set.json`
- `contracts/openapi-source/owners/module.entities/openapi.json`
- `contracts/openapi-source/owners/module.timeline/openapi.json`
- `contracts/openapi/cartulary.openapi.yaml`
- `db/migrations/00042_timeline_mention_public_identity.sql`
- `docs/handoffs/workbook-entity-mention-actions-refactor-handoff.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `internal/app/operator/operator_migration_evidence_test.go`
- `internal/gen/contractopenapi/artifacts_gen.go`
- `internal/gen/openapioperations/catalog_gen.go`
- `internal/modules/database_migrations/catalog_characterization_test.go`
- `internal/modules/entities/mention_projection_migration_test.go`
- `internal/modules/entities/openapi_contract_test.go`
- `internal/modules/timeline/derivation_equivalence_test.go`
- `internal/modules/timeline/resolution_integration_test.go`
- `internal/modules/timeline/workbookprojection/collection_facts.go`
- `internal/testutil/pgtest/pgtest_test.go`
- `packages/protocol-ts/src/generated/core-http-types.ts`
- `packages/protocol-ts/src/generated/core-http-validators.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/database-migrations/generate-catalog-projections.mjs`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/migration_history_manifest.json`
- `tools/test_families/module.entities.json`
- `tools/test_families/module.timeline.json`
- `tools/test_families/web.workbook.json`

### Completion and final scope review

EMA-05 DONE. All eight gaps are resolved, including independently retained entity
and source refresh obligations. There is no remaining implementation prerequisite
or seam blocker. The default-concurrency Network Flow focus timeout remains a
verification caveat outside this seam; the unmodified full target passed with
its supported CPU-capacity input. No Network Flow fix or next refactor was begun.

The final inventory is 98 changed/new paths. Scope review confirmed only the
approved specification/projection/receipt prerequisite, Timeline mention and
narrow workbook integration, focused tests/routing, generated derivatives,
reviewed goldens and this handoff. The digest, release snapshots, dependency locks,
branch and HEAD are unchanged. The local capacity file is a retained verification
input under `.cartulary/`, not a production or tracked change.

Final handoff Markdown and `git diff --check` passed after completion. The final
turn repeats these checks after this evidence note. Rollback preserves
all committed entities, mentions, links, receipts and history as described above.
