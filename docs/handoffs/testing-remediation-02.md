# Testing Remediation 02: Simplify Incident Authorization Tests

## 1. Scope and Source Posture

Document date: 2026-10-27, America/New_York. **Implementation authorized by the
current user request.** This tracker controls the sequential remediation effort.
The earlier document-only invocation and its evidence remain historical below.
Current sequence: **WF-00 -> WF-01 -> WF-05 -> WF-02 -> WF-03 -> WF-04**.
Each workstream requires a saved, validated checkpoint before its successor.
**Complete:** WF-00, WF-01, WF-05, WF-02, WF-03 and WF-04 are DONE. All eighteen selected test rows, lint/source checks, final evidence audits and document validation pass. WF-04 finished last.

The repository target is `/home/jochi/code/cartulary`, canonical relative identity
`.`, a directory. The bounded seam and normalized label are
`incident-authorization-tests`. The controlling artifact is
`docs/handoffs/testing-remediation-02.md`; its parent resolves inside the
repository without a symlink component. It was first created in the historical
document-save session; implementation now changes the bounded test/routing seam. Paths in this document are repository-relative unless absolute.

Artifact mode: **normal for this bounded seam**. The selected scope is incident
authorization tests, their shared route-inventory framework, overlapping
Timeline permission tests, and affected verification routing. The user selected
this scope over including Auth's separate route framework. Auth framework
redesign, broad HTTP-helper cleanup, browser/UI work, service lifecycle redesign,
performance and release qualification are excluded. Confirmed production defects
inside this incident-authorization seam are included as separate owner-backed
correction workstreams; this supersedes the earlier production-change exclusion.

This plan follows the [modular refactor planning framework](cartulary_modular_refactor_planning_framework.md)
and the Refactor Tracker twelve-section format. The root, backend and tooling
instructions apply to the planned changes. The [test-operations skill](../../.agents/skills/cartulary-test-ops/SKILL.md)
governs verification selection, execution, diagnosis and reporting.

[Core 00](../spec/00_document_set_status_and_precedence.md) allocates authority;
Core 00–04 are the current implementation-conformance corpus. The
[Testing Harness NLSpec](../testing-harness-nlspec.md) is `adopted/current` and
owns harness mechanics only. Domain vocabulary supplies owner navigation.
No production or harness-mechanics amendment is proposed. Human review of
projection fidelity remains separate from passing tests.

[Testing Remediation](testing-remediation.md) and
[Testing Remediation 01](testing-remediation-01.md) are historical context. The
latter records completion of WF-55–61 for WebSocket infrastructure. Their
execution permissions, historical source identities and passing results do not
carry forward. Their unrelated open work is not closed by this plan.

Tests, generators, runtime metadata, conformance and release evidence must not
read, stat, hash or otherwise consume this handoff or any other Markdown.
The assertion ledger below is human review material, never executable routing.

**Historical document-save baseline (E-001–008); implementation evidence starts at E-009.**

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Checkout | Clean `main`; HEAD `396a6d364ec0ce9807a6b9ad143964abc8a04363` | E-008 records final comparison | No commit, branch, push or production change requested |
| Index | SHA-256 `4b95d53f457739625bd2e818417ff34de8c5acdaf35a90b339b150b4aac834dc` | E-008: index bytes unchanged | Git inspection disables optional locks |
| Tracked/non-ignored files | 7,666 paths; content/file-identity digest `964c8644fb907376c1838ebf31b69eb46fe08708a044f4220aba4b66eeb3bdab`; no staged, unstaged or untracked changes | E-008: all original paths unchanged; only this handoff added | Enumeration is not behavioral inspection |
| Output | Absent; parent admitted; no symlink components | Absence rechecked before creation; regular single-link output verified | No existing tracker replaced |
| Comparison material | `/tmp/cartulary-tr02-doc-chx2ukk1/baseline.json` | Final comparison retained beside it | Human document-session preservation evidence only |
| Ignored locations | Caches, installed dependencies, build outputs and retained test results excluded from source inventory | Documentation outputs directed outside the repository if validation runs | No exhaustive ignored-tree preservation claim |

## 2. Current-State Repository Inventory

The initial movement inventory was the following twelve tracked, authored, regular Go files.
All twelve were directly inspected at the declarations/assertion bodies proposed
for removal or replacement. The two mixed incident test files and extension
test file retain unrelated bodies; this is not a whole-file behavioral audit of
those retained tests. No generated, vendor, binary, symlink or submodule entry is
in this bounded source universe.

| Path | Inspection / current responsibility | Callers and dependencies | Owner / disposition / risk | Evidence / contracts |
| --- | --- | --- | --- | --- |
| `internal/testutil/routeinventory/routeinventory.go` | Direct: entry/fixture types, transport and role categories, path substitution, selection | Six providers below; incident test interpreter | Shared test mechanism; delete after caller cutover; avoid recreating a framework | E-003; C-001/002 |
| `internal/modules/incidents/inventory_helpers_test.go` | Direct: universal fixture, role interpreter, request dispatch and name-specific repairs | Incident public/control tests; app, Auth, Timeline and socket support | Incident test composition; replace with minimal private fixtures | E-003; C-001–004 |
| `internal/modules/incidents/integration_test.go` | Direct at five control tests and progression helper | Incident, membership, preference, query, record and socket inventories | Preserve current-session evidence; leave independent durability tests intact | E-003; C-001/003/004 |
| `internal/modules/incidents/http_conformance_test.go` | Direct at three public-envelope tests, five deployment-admin tests, their helpers and preservation witnesses | Shared interpreter and provider declarations | Consolidate overlapping observations; preserve bootstrap/version/last-admin guarantees | E-003; C-001–003 |
| `internal/modules/incidents/testsupport/routetest/routes.go` | Direct: duplicate public/control incident and membership descriptions | Incident public/control suites | Remove inventory provider; retain real scenario helpers elsewhere | E-003; C-001/002 |
| `internal/modules/timeline/testsupport/routetest/inventory.go` | Direct: query/create/record/socket descriptions | Incident control suites | Remove metadata; keep sibling `timeline.go` and its real `CreateRow` operation | E-003; C-001/004 |
| `internal/modules/entities/hostidentity/testsupport/routetest/inventory.go` | Direct: Hosts/Identities query descriptions | Incident control suites; exact boundary-policy exception | Remove provider and stale policy allowance without broadening imports | E-003; C-001/005 |
| `internal/modules/indicators/testsupport/routetest/inventory.go` | Direct: Indicators query description | Incident control suites | Remove provider; preserve this view's executed authorization coverage | E-003; C-001 |
| `internal/modules/workbook/testsupport/routetest/preferences.go` | Direct: public/control posture parameter and four preference descriptions | Incident public/control suites | Remove provider; preserve personal versus incident-default permissions | E-003; C-001/002 |
| `internal/modules/timeline/support_integration_test.go` | Direct: query/create/patch/review/supersede permission matrix | App/Auth/incident helpers and real Timeline row creation | Consolidate after all outcomes, including reviewed-target supersede, have successors | E-003; C-001/003 |
| `internal/platform/httpapi/extensiondiscovery/testsupport/routetest/routes.go` | Direct: singleton discovery route description | Inventory self-check only in inspected references | Delete incidental declaration | E-003; C-005 |
| `internal/platform/httpapi/extensiondiscovery/routes_test.go` | Direct: self-check and neighboring handler/response/OpenAPI assertions | Discovery provider, HTTP handler and generated-registry fixture | Delete only self-validating metadata assertion; keep product witnesses | E-003; C-002/005 |

Current replacement files, directly reviewed during implementation:

| Path | Responsibility / final disposition |
| --- | --- |
| `internal/modules/incidents/authorization_helpers_test.go` | Private account/incident fixture, real membership transitions, explicit transport assertions and response closure |
| `internal/modules/incidents/authorization_incident_test.go` | Collection and resource scenarios; ordinary-account creation and metadata persistence |
| `internal/modules/incidents/authorization_membership_test.go` | List/audit/create/patch/delete scenarios; target-specific prerequisites, versions, durable attribution |
| `internal/modules/incidents/authorization_workbook_test.go` | Preference and four query scenarios; route-based preference setup and query POST without CSRF |
| `internal/modules/incidents/authorization_timeline_test.go` | Create/patch/review/rough and reviewed supersede scenarios; semantic socket admission |
| `internal/modules/incidents/authorization_observations_test.go` | Concrete protected-state observations and existing history/publication helpers; explicit incident fields |

Nine initial files were deleted at WF-02; the two mixed incident files and the
extension test file retain their independent assertions. Six replacement files
are private to `incidents_test`. Authored routing changed in three family JSON
files and the existing row-migration crosswalk; the Entities boundary allowance
was deleted. The only generated change is the execution-topology render index.
No production Go, schema, public API, role vocabulary or specification changed.

Supporting evidence, outside the twelve-file movement inventory:

| Surface | Accounting and depth | Use / boundary |
| --- | --- | --- |
| `internal/testutil/appsupport/runtime.go`; `internal/testutil/httptestx/httptestx.go` | Direct at composition, service admission, request and envelope helpers | Reuse established isolation and explicit disabled test-route mode; no shared runtime redesign |
| Incident `testsupport/scenariotest/incidents.go`, `testsupport/mutationtest/mutations.go`, `admission/admission_test.go` | Direct | Existing real membership operations, scoped audit observations and non-ordinal production role tests |
| Workbook `routes_startup.go`, startup PostgreSQL preference reader and `route_guard_characterization_test.go` | Direct at preference authorization and independent guard matrix | Prepare personal preferences only where needed; retain operation identity/guard evidence |
| Collaboration `integration_test.go`; Timeline lifecycle test | Direct at targeted socket membership loss; Timeline lifecycle selector and relevant boundary inspection | Existing established-socket/other-incident witness is not replaced by fresh dial tests; lifecycle coverage remains separate |
| `tools/test_families/{module.incidents,module.timeline,app.server,module.workbook,module.collaboration}.json`; verification owners | Relevant rows/contracts directly inspected; other rows metadata-only | Exact selected-row ledger in Section 8; no whole-owner qualification claim |
| `tools/test_catalog_row_migrations.json` and its schema; `tools/backend_module_boundaries.json` | Migration shape and relevant policy entry directly inspected | Record real successors and remove the deleted Entities filename allowance |
| Generated-artifact policy, execution topology, public Make recipes and Markdown lint configuration | Relevant authored rules/recipes inspected; generated products used only for command discovery | Update authored routing before future Make generation; do not hand-edit projections |
| Other repository files; ignored outputs, dependencies and vendor material | Excluded from behavioral audit with scope-based reason | No whole-repository test-quality or external-consumer claim |

Discovery used bounded `rg` file/symbol/import searches, direct source reads,
JSON row inspection and Git inspection. Text searches establish the known
repository caller set, not absence of arbitrary dynamic or external consumers.
WF-00 revalidated this bounded set and exact assertion ledger before movement.
No unrelated suite discovery was added.

## 3. Module Boundary Diagnosis

| Responsibility | Current location | Owner candidate | Decision | Findings / contracts |
| --- | --- | --- | --- | --- |
| Current membership and role propagation across real application routes | Incident control suites and shared interpreter | `incidents_test` | Keep composed evidence; replace interpreter with explicit scenarios | F-001–005; C-001/002 |
| HTTP, application and socket mechanics | Established test-support packages | Existing support owners | Keep; no new runner, registry, scenario DSL or generic fixture engine | C-002/004/005 |
| Source mutation semantics, history and lifecycle | Source-owner suites | Incidents, Timeline, Workbook, Collaboration | Keep independent meaningful witnesses | C-003/004 |
| Route-description bookkeeping | Shared package and six providers | No durable production responsibility | Remove | F-001/006/007 |
| Auth's separate route framework | Auth support suites | Auth | Defer, explicitly outside this iteration | User-selected boundary, E-002 |

New helpers remain private to the incident test package and perform concrete
setup or observation. A local test table may hold explicit expected outcomes;
it must not grow an interpreter for transport, role tiers, route names or
mutation-owner categories.

## 4. Public Contract and Behavior Freeze Map

| Contract / surface | Observed evidence | Adopted owner and required behavior | Disposition / prerequisites | Risks / slices / validation |
| --- | --- | --- | --- | --- |
| C-001: current authorization | Same-session progression; separate deployment-admin denial; overlapping Timeline matrix | Core 04 REQ-04-021–023, REQ-04-026, REQ-04-028–030; Core 01 REQ-01-172/180; Core 03 REQ-03-104/106: exact permissions, current membership, no deployment-admin bypass | Preserve and strengthen editor/demotion coverage; independent literal expectations; V-001 baseline before movement | F-001–005; S-001/002/003; V-001–003 |
| C-002: transport outcomes | Public inventory executes actual routes; HTTP helpers check status/request ID/envelopes | Core 01 §§3.3.4, 3.3.5.1–3.3.5.3 and common envelopes; Core 04 REQ-04-003: authentication, CSRF, concealment, exact success/error outcomes and no-content deletion | Preserve executed observations; no requirement to preserve test metadata; qualify before deletion | F-001/003/006; S-001/003; V-002/003 |
| C-003: mutation integrity | Independent bootstrap/versioning/atomicity tests; inventory `MutationOwners` itself proves no effects | Core 01 REQ-01-160–166/176–180 and membership contracts; Core 02 history; Core 03 REQ-03-104–107: authorized durable changes and rejected-operation atomicity | Preserve independent witnesses; add scoped protected-state observations to migrated denials | F-003/004/007; S-002/003; V-002/003 |
| C-004: socket admission | Fresh hello/dial denial plus independent established-subscription revocation | Core 01 §3.3.10; Core 04 REQ-04-017/023: current admission and committed, targeted membership-loss effects | Keep fresh admission and existing revocation/isolation as distinct witnesses | F-001–003; S-001/003; V-002/003 |
| C-005: evidence integrity | Three incident rows, Timeline matrix row and mixed extension row; service metadata and generated routing | Testing Harness REQ-012–016: exact live selectors, truthful ownership/services, immutable identities, migration crosswalks, no aliases | Authored-input-first cutover and Make regeneration; no document-fed executable accounting | F-006/007; S-004; V-004–006 |

Owner status was inspected at E-003: Core 00–04 are current normative owners and
the Testing Harness owner is adopted/current. No primary-owner contradiction or
required amendment was found. If later execution exposes a production defect,
record it against its owner and separate its correction from structural movement;
do not normalize the defect in fixtures or weaken expectations.

No production API/type, storage schema, view identity, saved-view behavior,
projection contract, public Make command or runtime mode changes. Browser/UI
selectors are not applicable. Core 05 publication qualification is not claimed.

## 5. Coupling and Boundary Findings

| Finding | Evidence / consequence | Classification / owner | Required action | Trace |
| --- | --- | --- | --- | --- |
| F-001: parallel descriptions and role interpretation | E-003: public/control route metadata repeats operations; `AllowedRole` determines expectations | must_fix; incident test composition | Explicit operation/outcome cases; no expectations derived from production policy or another inventory | C-001/002; S-001/003; V-001/002 |
| F-002: excessive setup | E-003: every route fixture seeds candidate/member accounts and two Timeline records | must_fix; incident fixtures | Arrange only the operation's prerequisites | C-001–004; S-002; V-002 |
| F-003: hidden repair | E-003: successful routes update fixture versions or reset records by route-name strings | must_fix; incident fixtures | Handle versions and fresh targets explicitly in each scenario | C-001–004; S-001/002; V-002 |
| F-004: overlapping permission suites | E-003: Timeline's five groups repeat query/create/patch/review/supersede role outcomes | must_fix; Incidents/Timeline tests | Consolidate with every distinct outcome and reviewed-target prerequisite retained | C-001/003; S-003; V-002/003 |
| F-005: incomplete progression | E-003: current sequence skips editor and returns no downgraded-role observation before removal | must_fix; incident authorization evidence | Exercise all four roles and admin-to-viewer downgrade using one session | C-001; S-001; V-002 |
| F-006: self-validating metadata | E-003: extension test compares its own singleton inventory to constants | must_fix; extension test support | Delete that test/provider, preserving handler/response/OpenAPI witnesses | C-002/005; S-003/004; V-003/004 |
| F-007: misleading effect metadata | E-003: inventory carries `MutationOwners` without using it to assert durable effects | must_fix; test observations | Remove unused categories; assert actual scoped state at the operation boundary | C-003/005; S-002/004; V-002/004 |
| F-008: tracker sequencing and validation omissions | Current request and inspected authored rows | must_fix; documentation/execution | Separate S-002 as WF-05, include owner-required production corrections, and add omitted preservation rows | All slices; V-001–006 |

Decision rationale, compatibility and gap-to-slice evidence:

| Gap / slices | Remediation and affected areas | Rationale and long-term benefit | Compatibility / migration | Risk if unresolved | Validation evidence |
| --- | --- | --- | --- | --- | --- |
| F-001 / S-001, S-003, S-004 | Tests/support/routing/docs: seven direct operation-family groups replace the route interpreter | One readable operation boundary avoids a second policy model; future operations extend their owning group | Private inventory interfaces deleted without aliases; four row migrations preserve meaningful evidence | Policy-mirroring tests and drift can conceal missing cases | E-010/012/013; all retained operations have literal independent outcomes; E-014 final portfolio |
| F-002 / S-002 | Tests/support/docs: isolated per-operation fixture reused across role stages; only required actors and targets | Visible prerequisites reduce unrelated failure modes and setup coupling | Test arrangement only; PostgreSQL/object-store and cleanup ownership retained | Unrelated fixture failures obscure authorization and impede maintenance | E-011/012; query/socket fixtures have no Timeline records; membership targets only for writes; E-014 cleanup audits |
| F-003 / S-001, S-002 | Tests/support: returned versions, fresh terminal/destructive targets, material writes | Valid preconditions separate authorization from version/lifecycle/no-op failures and keep scenarios extensible | No lifecycle exception, production hook or compatibility wrapper | Invalid repeated operations can produce misleading status or persistence evidence | E-011/012; material state and versions persist; rough/reviewed supersede both qualify; E-014 final source |
| F-004 / S-003 | Tests/routing/docs: consolidate composed permissions while retaining independent owner witnesses | One clear owner reduces duplicate maintenance without weakening lifecycle assurance | Timeline matrix row migrates to query and mutation groups; lifecycle row remains | Duplicate matrices drift; premature retirement loses reviewed-target/editor outcomes | D-004–006 close at E-012; E-014 preserves Timeline lifecycle and all operation outcomes |
| F-005 / S-001 | Tests/docs: full same-session progression, downgrade/removal and separate deployment-admin denial | Current authorization and revocation are explicit without ordinal role assumptions | No permission or role-vocabulary change; production correction only for demonstrated owner violations | Editor access, stale privilege or admin bypass can go undetected | E-010/011/012; exact role outcomes without relogin/sleep/retry; E-014 final portfolio |
| F-006 / S-003, S-004 | Tests/support/routing/docs: delete singleton inventory/self-check; retain four product assertions | Removes bookkeeping with no independent assurance and avoids a needless compatibility surface | Extension row identity retained; only obsolete selector removed; no artificial successor | Apparent coverage adds coupling without protecting behavior | E-012/013; provider/self-check absent; retained extension row passes; E-014 final evidence |
| F-007 / S-002, S-003 | Tests/observations/docs: remove unused inventory mutation categories; verify scoped source and attributable durable effects | State observations establish integrity and atomicity instead of descriptive metadata | Existing meaningful owner observers remain; no generic database snapshot framework | Correct HTTP status can conceal writes, audit/history/receipt or publication effects | E-011/012; denials preserve state, successes persist fields/versions/attribution; E-014 boundary repair and qualification |
| F-008 / S-005 and every checkpoint | Docs/execution: separate WF-05, add five omitted preservation rows, admit scoped production corrections, save each gate | Sequential evidence makes completion auditable and keeps future handoffs accurate | Existing finding/slice/workflow/disposition IDs and historical evidence retained | Combined checkpoints and omitted guarantees can overstate completion | E-009–014; every workstream checkpoint saved before successor; all eighteen selected rows covered |

No normative specification change was justified: governing requirements,
observed baseline behavior and previously missing evidence remain distinct in
Section 4 and E-009. The specification reconciliation is complete; these gaps
required test composition/evidence and routing changes. Public APIs, schemas,
runtime modes and Make command contracts retain their owner-defined meaning.

`must_fix` applies only to the named slices. No new analyzer, suppression,
compatibility facade, alias, fixture DSL, runner or production test hook is
admitted. Existing test counts, names, metadata shapes and seed arrangements
are not compatibility requirements.

Assertion dispositions were classified at WF-00, qualified before retirement at
WF-01/WF-05 and closed at WF-02 (E-012); final-source admission is E-014:

| Disposition ID | Baseline assertion / mechanism | Classification | Successor or retirement rationale |
| --- | --- | --- | --- |
| D-001 | Thirteen public HTTP operations: four incident, five membership, four preference operations | Production assurance | Collection/resource/membership/preference scenarios retain actual success, request-ID, envelope, authentication and applicable CSRF observations |
| D-002 | Twenty control operations: two incident, five membership, four preferences, four queries, four Timeline mutations and socket admission | Production assurance | Six incident-scoped scenario groups retain each operation and current-session progression; counts are discovery accounting, never executable acceptance constants |
| D-003 | Deployment administrator without membership denied on those control operations | Security assurance | Denial case within each corresponding operation scenario, using a distinct deployment administrator |
| D-004 | Timeline query/create/patch/review/supersede permission matrix, including editor rejection and reviewed supersede targets | Production assurance | Query and Timeline-mutation successors before the old matrix is removed |
| D-005 | Actual bootstrap, replay, version, last-admin, closed-incident, atomicity and source lifecycle tests | Distinct production assurance | Retain independent tests; shared status assertions do not make their effects redundant |
| D-006 | Workbook operation identity/guard matrix and Collaboration established-socket revocation/isolation | Distinct architecture/security assurance | Retain; role progression and fresh connections cannot replace them |
| D-007 | Extension inventory self-check, transport discriminator, role tiers, path replacer, name repair and unused mutation-owner metadata | Incidental test behavior | Retire without artificial replacement checks; real transport/product witnesses remain |
| D-008 | Isolated fixtures, disabled privileged test routes, cleanup and exact catalog selection | Necessary harness assurance | Existing providers and routing retained; no new lifecycle or accounting mechanism |

## 6. Refactor Workstreams

All workstreams are included and sequential. This table is the authoritative
dependency list. Update and validate this tracker after each exit and before
entering its successor; WF-04 finishes last. The current request authorizes execution.
WF-05 separates S-002 from WF-01, retaining existing workflow and slice identities.

| Workflow | Class / prerequisite | Goal / files | Validation | Checkpoint and binary exit |
| --- | --- | --- | --- | --- |
| WF-00: characterize | root; current authorization and baseline revalidation | Close D-001–008 and caller accounting; selected current tests | V-001 | Save exact baseline roots/native exits and assertion ledger; no removal lacks a disposition |
| WF-01: replace framework | chain; WF-00 | S-001: explicit scenarios, literal outcomes and complete progression | V-002 | All seven groups pass; predecessors remain; save checkpoint |
| WF-05: fixtures and observations | chain; WF-01 | S-002: minimal fixtures, valid material writes and scoped durable observations | V-002 | Success and denial effects qualify; save checkpoint before retirement |
| WF-02: consolidate | chain; WF-05 | S-003: consolidate Timeline/incident duplicates and remove all inventory consumers | V-003 | Every old assertion has its qualified successor or incidental disposition |
| WF-03: reconcile routing | chain; WF-02 | S-004: complete authored rows, migration crosswalks and stale boundary allowance removal | V-004 | Exact selectors/services/owners and generated projections agree; no stale imports or aliases |
| WF-04: qualify and hand off | chain; WF-03 | Finalize, freeze, qualify affected portfolio and finish documentation | V-005/006 | Compatible final results, evidence audits, exclusions and all workstream exits recorded |

Register new tests through authored rows and generation as they are introduced
so WF-01 successors can run canonically. WF-03 is the final reconciliation gate,
not permission to defer executable routing until after qualification.

## 7. Refactor Slice Plan and Execution Status

| Slice / kind | Findings / contracts | Entry / readiness | Required change and exit | Validation / rollback |
| --- | --- | --- | --- | --- |
| S-001: structural preservation with stronger authorization evidence | F-001/003/005; C-001/002/004 | DONE at E-010; WF-00 characterization qualified | Seven explicit groups below; literal outcomes and same-session changes qualify | V-002; restore scenarios/callers/rows together |
| S-002: structural fixture and observation simplification | F-002/003/007; C-001–004 | DONE at E-011; separate WF-05 after S-001 | Minimal setup, real versions, scoped no-mutation evidence and bounded cleanup | V-002; restore fixture and affected scenarios together |
| S-003: structural consolidation and incidental retirement | F-001/004/006/007; C-001–005 | DONE at E-012; successors qualified before retirement | All D dispositions close; duplicate suites and shared inventory disappear | V-003; restore tests/helpers and associated routing together |
| S-004: routing projection cleanup | F-006/007; C-005 | DONE at E-013; incremental additions in WF-01, final routing gate after WF-02 | Seven semantic groups, real crosswalks and accurate generated selection | V-004/005; restore authored inputs and regenerate through Make |
| S-005: final validation and handoff | F-001–008; C-001–005 | DONE at E-014; all predecessor gates passed | Finalize, freeze source, qualify portfolio, audit evidence and complete tracker | V-005/006; no completion with unmet gates |

### S-001 — Explicit authorization scenarios

New authored row IDs allocated through `make author-test-row-id` at WF-01 entry:

| Test | Exact row ID |
| --- | --- |
| `TestIncidentCollectionAuthorization_Integration` | `module.incidents.integration.incident_collection_authenticated_authorization_dab8552d8f` |
| `TestIncidentResourceAuthorization_Integration` | `module.incidents.integration.incident_resource_current_authorization_b010ae008a` |
| `TestIncidentMembershipAuthorization_Integration` | `module.incidents.integration.incident_membership_current_authorization_5018f673df` |
| `TestIncidentWorkbookPreferencesAuthorization_Integration` | `module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6` |
| `TestIncidentWorkbookQueriesAuthorization_Integration` | `module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2` |
| `TestIncidentTimelineMutationsAuthorization_Integration` | `module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da` |
| `TestIncidentSocketAdmissionAuthorization_Integration` | `module.incidents.integration.incident_socket_current_admission_697be7989c` |

Create seven independently selectable groups in the incident test package:
incident collection, incident resource, memberships, Workbook preferences,
Workbook queries, Timeline mutations and socket admission. Keep direct test
logic together by operation family, with small private setup/observation helpers.
Use ordinary Go tests and local tables with literal expected status/code values.
Do not derive expected permissions from production role sets, registered routes,
another inventory or a role-ordering interpreter.

For every incident-scoped operation, retain the same subject session through:

`no membership -> viewer -> editor -> reviewer -> admin -> viewer -> removed`

A separate incident administrator performs membership changes through real
routes, using returned membership versions. No relogin, clock advance, polling
or eventual retry substitutes for immediate reauthorization. The second viewer
stage proves downgrade after privilege gain. An independent deployment
administrator without membership must also receive the concealed-resource denial.

| Operation family | Explicit permitted roles |
| --- | --- |
| Incident read; membership list; both preference reads; personal preference update; four view queries; socket admission | viewer, editor, reviewer, admin |
| Timeline create and record patch | editor, reviewer, admin |
| Incident metadata patch; Timeline review and supersede | reviewer, admin |
| Membership audit; membership create/patch/delete; incident default preference update | admin |

Preserve Timeline, Hosts, Identities and Indicators query coverage. Incident
collection list/create remain authenticated operations without a pre-existing
membership requirement, separate from incident-scoped permission progression.

Assert exact success status, envelope/request-ID relationships, `404
incident_not_found` for non-members and `403 authorization_denied` for visible
but insufficiently privileged members. Preserve applicable `401 session_required`
and `403 csrf_verification_failed` observations. Query POSTs remain reads; HTTP
method alone must not decide CSRF policy. Membership deletion asserts `204` and
no response body rather than requiring a JSON envelope. Socket admission uses
the established semantic client and real hello/upgrade failure observations.

### S-002 — Minimal fixtures and meaningful observations

- Allocate an isolated application/database fixture per operation scenario and
  reuse it across that operation's authorization stages. Keep explicit disabled
  privileged test routes and existing provider cleanup.
- Seed membership targets only for membership operations and Timeline records
  only for record operations. Remove unconditional candidate/member/record seeds.
- Prepare personal preferences through the existing PUT route only for
  preference scenarios; remove the interpreter's unconditional SQL insertion.
  Do not change production handling of absent preference rows to fit test setup.
- Read versions from real responses. Allocate fresh valid targets for repeated
  review/supersede attempts and preserve reviewed-target supersede coverage.
  Valid setup prevents a stale-version or terminal-state failure from masking
  the authorization outcome under examination.
- Observe authorized durable outcomes and compare scoped protected state before
  and after denied mutations. Include the operation's affected resource fields,
  versions, preferences or membership rows and relevant audit/history/idempotency/
  publication-intent effects. Use existing owner observations or small local
  queries, not a generic database snapshot framework. Arrange prerequisites
  before the snapshot; exclude legitimate session-activity bookkeeping.
- Close response bodies, including `204` bodies, and socket clients explicitly
  using existing lifecycle support.

### S-003 — Consolidate and delete

After passing successors, remove the thirteen inventory-driven incident test
entrypoints and interpreter. Remove Timeline's `TestAuthorizationMatrix`, mapping
its query and four mutation groups to the new cases. Delete the shared
`routeinventory` package and all six inventory providers listed in Section 2.
Retain Timeline's sibling `timeline.go` with real `CreateRow` support.

Delete only `TestPublicRouteInventoryExtensionDiscovery_Unit` from the mixed
extension-discovery test file. Keep its handler, authentication, response-shape
and OpenAPI tests. Remove the boundary-policy allowance naming the deleted
Entities inventory. Do not widen allowed imports to make moved tests compile.

Retain independent bootstrap state, idempotency, version-conflict, last-admin,
closed-incident membership administration, atomicity, Workbook operation-guard,
Timeline lifecycle and targeted socket-revocation witnesses. No old test is
removed solely because another test calls the same route.

### S-004 — Routing and interfaces

Allocate semantic rows for the seven groups under `module.incidents.integration`.
Use appropriate collaborators and actual verification references, including the
membership-audit reference where exercised. Preserve PostgreSQL and object-store
service declarations, `postgres_dedicated` fixture capability and exclusive
PostgreSQL isolation; simpler test data does not remove application dependencies.
Use standard aggregate placement for the consolidated authorization groups.

Retire the three old incident rows and the Timeline matrix row in Section 8,
recording their actual replacement rows in the existing migration crosswalk.
Update the extension-discovery row selector without retiring its meaningful
remaining tests. Do not invent a successor for its incidental self-check.

Actual authored row migration map (qualified at E-012; reconciled at E-013):

| Retired row | Exact successor rows |
| --- | --- |
| `module.incidents.integration.one_same_session_role_progression_no_membership_6df5c1b9fb` | `module.incidents.integration.incident_membership_current_authorization_5018f673df`; `module.incidents.integration.incident_resource_current_authorization_b010ae008a`; `module.incidents.integration.incident_socket_current_admission_697be7989c`; `module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da`; `module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6`; `module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2` |
| `module.incidents.support_integration.control_boundary_incident_core_deployment_admin_9ba3db3d5f` | `module.incidents.integration.incident_membership_current_authorization_5018f673df`; `module.incidents.integration.incident_resource_current_authorization_b010ae008a`; `module.incidents.integration.incident_socket_current_admission_697be7989c`; `module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da`; `module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6`; `module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2` |
| `module.incidents.support_integration.public_route_inventory_incident_surfaces_d792370039` | `module.incidents.integration.incident_collection_authenticated_authorization_dab8552d8f`; `module.incidents.integration.incident_membership_current_authorization_5018f673df`; `module.incidents.integration.incident_resource_current_authorization_b010ae008a`; `module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6` |
| `module.timeline.support_integration.integration_authorization_matrix_2699762a78` | `module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da`; `module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2` |

`app.server.unit.extension_discovery_transport_7f6ad80724` keeps its identity and
selects `TestExtensionDiscoveryRejectsSingletonQueryBeforeAuthentication_Unit`,
`TestExtensionDiscoveryRequiresAuthentication_Unit`,
`TestExtensionDiscoveryResponseShapeAndOrder_Unit` and
`TestOpenAPIExtensionDiscoveryExposesClosedContract_Unit`. The deleted inventory
self-check has no replacement because it provided no independent product assurance.

Only private test interfaces disappear. No production API/type, compatibility
wrapper, alias, feature flag, runner, schema version or production test hook is
introduced. Authored ownership/routing changes precede Make generation; generated
artifacts are never hand-edited. A production behavior correction, if discovered,
requires a separate owner-backed correction workstream within this authorized effort.

## 8. Validation Plan

**V-001–005 are required implementation gates. Current execution evidence is recorded below; historical document-save evidence is unchanged.** Commands execute from `/home/jochi/code/cartulary` through public Make
targets. Read current tool versions from the pin registry, not this handoff.

| Validation | Stage / trace | Canonical command and selection | Required result / gate |
| --- | --- | --- | --- |
| V-001 | Pre-move; WF-00; D-001–008 | `make help ALL=1`; `make task-guide OWNER=<owner> DETAIL=coverage JSON=1` for `module.incidents`, `module.timeline`, `app.server`; `make test-slice OWNER=<owner> ROWS=<exact baseline IDs below>` | Current passing baseline with native exits and complete assertion ledger; investigate failure before movement |
| V-002 | Post-slice; S-001/002; C-001–004 | Exact new rows through `make test-slice OWNER=module.incidents ROWS=<new semantic row IDs>` | All roles, downgrade/removal, no admin bypass, exact wire outcomes, durable effects and valid prerequisites qualify before retirement |
| V-003 | Consolidation; S-003; D-004–008 | New rows plus exact affected preservation rows below; extension-discovery transport row after selector update | Independent guarantees survive; no incidental test receives a fabricated successor |
| V-004 | Routing; S-004; C-005 | `make generate`; `make source-check CHECK=generated`; `make source-check CHECK=schema`; `make source-check CHECK=catalog`; `make source-check CHECK=identity` | Authored/generated consistency and exact live selectors; no orphaned imports or rows |
| V-005 | Final source; WF-04 | `make agent-finalize`, then frozen-source affected row portfolio, `make lint CHECK=go`, `make lint CHECK=backend-boundary`; `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all` | Compatible final evidence, native exits, cleanup and retained validity; re-evaluate only affected obligations after repair |
| V-006 | Document save and final handoff | `make lint-markdown`; direct structure/table/link/diff review | Documentation validity with actual coverage limits; never product qualification |

Historical baseline rows, executed at WF-00 before retirement:

| Owner | Row ID |
| --- | --- |
| `module.incidents` | `module.incidents.integration.one_same_session_role_progression_no_membership_6df5c1b9fb` |
| `module.incidents` | `module.incidents.support_integration.control_boundary_incident_core_deployment_admin_9ba3db3d5f` |
| `module.incidents` | `module.incidents.support_integration.public_route_inventory_incident_surfaces_d792370039` |
| `module.timeline` | `module.timeline.support_integration.integration_authorization_matrix_2699762a78` |
| `app.server` | `app.server.unit.extension_discovery_transport_7f6ad80724` |

Preservation portfolio, discovered from current authored rows:

| Owner | Row ID / distinct assurance |
| --- | --- |
| `module.incidents` | `module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972`: creator/preferences, Location, membership admission, version/last-admin and closed-incident administration |
| `module.incidents` | `module.incidents.unit.incident_admission_role_sets_and_denials_are_closed_0ed8398f44`: non-ordinal role sets and typed denial precedence |
| `module.workbook` | `module.workbook.catalog.keep_saved_clears_the_conflict_without_creating_c65a9c3099`: contains independent `TestWorkbookRouteGuardsFailBeforeMutation`; its other selected conflict witnesses remain part of this row's accounting |
| `module.timeline` | `module.timeline.integration.incident_role_authorization_review_and_supersede_62f3d384f6`: source lifecycle authorization and transitions |
| `module.collaboration` | `module.collaboration.support_integration.incident_socket_revocation_sources_00f908473e`: established-subscription termination and unaffected-user/incident isolation |

Affected owner guides were refreshed at WF-00. New IDs were allocated at WF-01
and used literally in canonical selection; no executable tool reads this
document to discover them. Every command gets a fresh explicit
`CARTULARY_TEST_RUN_ID`; retain the native session handle, outer exit and exact
result root. Do not infer complete-owner acceptance from these bounded rows.

Run retained finalizer maintenance only with an eligible successful full warm
check root supplied as `RESULTS_DIR`; otherwise report it skipped because
`RESULTS_DIR` was unset. Do not launch a redundant warm check for maintenance.
Browser/visual, performance, whole-project, full-owner and release qualification
are excluded because the selected seam changes test composition, not those
production surfaces. No timing improvement or Core 05 publication claim follows.

| Session operation | Invocation-path / side effects | Admission / actual result |
| --- | --- | --- |
| Source and Git inspection | File reads, bounded searches and external comparison material | Executed, E-001–003/E-007; no test qualification |
| V-001–005 | Tests/services, generation, finalization or product checks | Current implementation authorized; E-009 onward records execution; original save session did not run them |
| V-006 | Public Markdown target uses frontend readiness, step artifacts and static-analysis cache | E-008 used installed-only preparation and external roots for the historical document-only save; E-014 records final implementation handoff validation |
| Markdown coverage | `.markdownlint-cli2.jsonc` excludes this handoff filename | A target pass qualifies configured documents only; direct review must cover this file; do not change lint scope for this save |

## 9. Top-Level Work Tracker

| Item | Workstream | Status | Dependency / evidence | Exit |
| --- | --- | --- | --- | --- |
| T-001: select scope and inspect current sources | Planning | DONE | E-001–003 | Bounded target and findings grounded |
| T-002: prepare decision-complete content | Planning | DONE | T-001; E-003 | Contracts, dispositions, slices and validation specified |
| T-003: save and review this handoff | Documentation | DONE | T-002; E-007/008 | Sole-file save and validation recorded |
| T-004: current characterization | WF-00 | DONE | E-009; V-001 | D-001–008 revalidated; five baseline rows pass |
| T-005: explicit scenarios | WF-01 | DONE | E-010; S-001 | Seven groups pass; predecessors retained |
| T-009: minimal fixtures and durable observations | WF-05 | DONE | E-011; S-002 | Seven groups pass with durable observations |
| T-006: consolidation and retirement | WF-02 | DONE | E-012; S-003 | V-003 passed; D-001–008 closed |
| T-007: routing/projection reconciliation | WF-03 | DONE | E-013; S-004 | V-004 passed |
| T-008: final qualification and handoff | WF-04 | DONE | E-014; S-005 | V-005/006 passed; final handoff complete |

Planning completion does not complete any implementation workstream.

## 10. Session Handoff Log

### Implementation entry and checkpoint ledger — E-009 onward

Current user authorization covers all planned slices and separate corrections for
confirmed in-scope production defects. The starting HEAD is
`396a6d364ec0ce9807a6b9ad143964abc8a04363`; tracked/index changes are absent;
this controlling tracker is the sole pre-existing untracked file. No normative
owner contradiction was established in planning. Core 01 REQ-01-148–151,
REQ-01-160–180, Core 03 REQ-03-104–107, Core 04 REQ-04-021–030 and
TH-HARNESS-REQ-012–016 govern the selected contracts. Expected values remain
literal test assertions; no executable tool consumes this Markdown.

| Checkpoint | Status | Evidence / native exit / exact run root | Next action |
| --- | --- | --- | --- |
| E-009: implementation entry / WF-00 | DONE | `make help ALL=1` and five owner coverage guides exit 0; five baseline rows passed with native exit 0 in three runs below; document lint exit 0 | Start WF-01 only after this saved checkpoint |
| E-010: explicit scenarios / WF-01 | DONE | Seven groups passed in `tr02-wf01-scenarios-20261027-01`, native session 61719, exit 0; corrected generation/format pass below | Predecessors retained; checkpoint reviewed; enter WF-05 |
| E-011: fixtures and observations / WF-05 | DONE | `tr02-wf05-effects-20261027-01`, native session 30388, exit 0; all seven groups pass | Scoped success/denial effects qualify; checkpoint reviewed; enter WF-02 |
| E-012: consolidation / WF-02 | DONE | Successors and preservation portfolio pass with native exit 0 in five runs below | D-001–008 closed; checkpoint reviewed; enter WF-03 |
| E-013: routing / WF-03 | DONE | Generation and generated/schema/catalog/identity source checks pass below | Checkpoint reviewed; enter WF-04 |
| E-014: final qualification / WF-04 | DONE | Final source portfolio and twelve retained-root audits pass; Markdown lint and direct tracker review pass | Bounded effort complete; no dependent work or blocker remains |

WF-00 native execution records (all exit 0; run parent
`/home/jochi/code/cartulary/.cartulary/test-results`):

| Run ID | Exact target and selection | Native session / result |
| --- | --- | --- |
| `tr02-wf00-incidents-20261027-01` | `make test-slice OWNER=module.incidents ROWS=module.incidents.integration.one_same_session_role_progression_no_membership_6df5c1b9fb,module.incidents.support_integration.control_boundary_incident_core_deployment_admin_9ba3db3d5f,module.incidents.support_integration.public_route_inventory_incident_surfaces_d792370039` | 57370; pass, exit 0 |
| `tr02-wf00-timeline-20261027-01` | `make test-slice OWNER=module.timeline ROWS=module.timeline.support_integration.integration_authorization_matrix_2699762a78` | 42736; pass, exit 0 |
| `tr02-wf00-extension-20261027-01` | `make test-slice OWNER=app.server ROWS=app.server.unit.extension_discovery_transport_7f6ad80724` | 89467; pass, exit 0 |
| `tr02-wf00-document-20261027-01` | `make lint-markdown` | 82733; pass, exit 0; tracker outside configured lint globs, direct review required |

Every invocation above used `CARTULARY_TEST_RUN_ID=<Run ID>`; its exact root is
the run parent joined with that ID. Coverage guides for `module.incidents`,
`module.timeline`, `app.server`, `module.workbook`, and `module.collaboration`
used `make task-guide OWNER=<owner> DETAIL=coverage JSON=1`, all exit 0, retained
at `/tmp/tr02-<owner>-guide.json`. Source HEAD remains the entry HEAD; only this
tracker changed before baseline execution. D-001–008 match the inspected source:
13 public operations, 20 scoped operations, deployment-admin isolation,
reviewed-target supersede, independent preservation tests, incidental metadata,
and existing fixture/routing mechanics. Counts are human discovery accounting.
No baseline defect or owner contradiction was observed. The checkpoint was
reviewed for twelve sections, table structure, stable IDs and dependency order.

Additional preservation rows admitted by the current plan:

| Owner | Exact row ID / guarantee |
| --- | --- |
| `module.incidents` | `module.incidents.integration.incident_create_persists_the_incident_row_bootst_0bd8d02f12`: bootstrap persistence and atomic rollback |
| `module.incidents` | `module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6`: normalized replay and duplicate conflict integrity |
| `module.incidents` | `module.incidents.integration.real_incident_patch_persists_only_patchable_meta_9fc8fc207f`: material metadata persistence/versioning |
| `module.incidents` | `module.incidents.integration.live_membership_patch_to_the_current_role_return_7a14cba87a`: same-role no-op integrity |
| `module.incidents` | `module.incidents.support_integration.authenticated_incident_creation_and_normalized_r_e2cd9c99b5`: HTTP replay/idempotency |

For F-001–007, remediation belongs to tests, test-support implementation, routing
where affected, and this documentation. Removing the parallel policy interpreter,
excess setup, hidden repair, duplicate execution, incomplete progression,
self-validating metadata and unobserved effects improves cohesion, clarity and
security evidence. Private interfaces retire without aliases; production APIs and
schemas have no planned change. Unresolved gaps risk false authorization assurance,
missed revocation, hidden mutations and maintenance drift. Exits in Sections 6–8
are the per-gap validation criteria. F-008 belongs to documentation and execution
planning and closes only with separate checkpoints and the complete portfolio.

Each confirmed production defect receives a new slice/workstream, owner clause,
reproducer, explicit correction and compatibility impact before repair. Dependent
work cannot resume until its correction qualifies. Final handoff includes exact
row crosswalks, final-source results, maintenance skips and all bounded exclusions.


### WF-01 checkpoint — E-010

WF-01 interim failures: `tr02-wf01-format-20261027-01` (native session 38527, exit 2, runtime-plan configuration error) and `tr02-wf01-generate-20261027-01` (63089, outer exit 2, owning generate exit 1 / harness unknown_failure). `make target-plan TARGET=format JSON=1` identified the invalid cross-owner verification reference. Corrected new incident rows to reference only incident-owned verifications; collaborating owners remain explicit in `collaborator_ids`. These are authored test-routing defects introduced in this work, not production failures.

WF-01 checkpoint: `CARTULARY_TEST_RUN_ID=tr02-wf01-generate-20261027-02 make generate`
(native session 74441, exit 0) and `CARTULARY_TEST_RUN_ID=tr02-wf01-format-20261027-02 make format`
(92844, exit 0) passed after repairing authored verification ownership.
`CARTULARY_TEST_RUN_ID=tr02-wf01-scenarios-20261027-01 make test-slice OWNER=module.incidents ROWS=module.incidents.integration.incident_collection_authenticated_authorization_dab8552d8f,module.incidents.integration.incident_resource_current_authorization_b010ae008a,module.incidents.integration.incident_membership_current_authorization_5018f673df,module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6,module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2,module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da,module.incidents.integration.incident_socket_current_admission_697be7989c`
passed with native exit 0 (61719). Exact roots use the previously stated absolute
run parent and IDs. These observations cover literal status/code/envelopes,
request IDs, authenticated collection behavior, all four roles, immediate downgrade
and removal, deployment-admin isolation, query POST without CSRF, empty 204 bodies,
and rough/reviewed supersede targets. All predecessor witnesses remain. Snapshot
and durable-success strengthening remains WF-05, not claimed complete here.
The saved checkpoint was directly reviewed for twelve sections, stable IDs,
row/selector identities, table closure and the WF-05 dependency before entry.

### WF-05 checkpoint — E-011

`CARTULARY_TEST_RUN_ID=tr02-wf05-format-20261027-01 make format` passed
(native session 46276, exit 0).
`CARTULARY_TEST_RUN_ID=tr02-wf05-effects-20261027-01 make test-slice OWNER=module.incidents ROWS=module.incidents.integration.incident_collection_authenticated_authorization_dab8552d8f,module.incidents.integration.incident_resource_current_authorization_b010ae008a,module.incidents.integration.incident_membership_current_authorization_5018f673df,module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6,module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2,module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da,module.incidents.integration.incident_socket_current_admission_697be7989c`
passed (native session 30388, exit 0). Exact roots use the absolute run parent
above. HEAD remains the entry HEAD; authored tests and routing are working-tree
changes. Six private test files now contain operation-specific fixtures and
observations. No production source changed.

D-001–004 successors now qualify status, envelope, immediate current-session
permissions, reviewed-target supersede and durable effects. Snapshots are taken
after prerequisite and membership setup; they cover protected incident/member/
preference/Timeline state and relevant audit, history, idempotency and immutable
publication intents, excluding session/dispatcher bookkeeping. Successful writes
assert material values, returned/persisted versions and attribution. Fresh targets
and real returned versions protect repeated transitions. Account collection,
query and socket fixtures create no unnecessary Timeline targets; personal GET
is prepared by PUT. Existing application cleanup ownership is unchanged.
D-005/006 remain independent preservation obligations; D-007 is incidental and
D-008 retains existing fixtures/routing. No owner contradiction or production
defect was observed. The saved checkpoint was directly reviewed for twelve
sections, stable IDs, exact commands, table closure and dependencies before WF-02.

### WF-02 checkpoint — E-012

All successors and ten preserved rows plus the retained extension row passed
after removal. The following commands used `CARTULARY_TEST_RUN_ID=<Run ID>`;
exact roots use the absolute run parent above. HEAD remains unchanged and the
implementation is in the working tree.

| Run ID | Exact target and selection | Native session / result |
| --- | --- | --- |
| `tr02-wf02-generate-20261027-01` | `make generate` | 28676; pass, exit 0 |
| `tr02-wf02-format-20261027-01` | `make format` | 99777; pass, exit 0 |
| `tr02-wf02-incidents-20261027-01` | `make test-slice OWNER=module.incidents ROWS=module.incidents.integration.incident_collection_authenticated_authorization_dab8552d8f,module.incidents.integration.incident_resource_current_authorization_b010ae008a,module.incidents.integration.incident_membership_current_authorization_5018f673df,module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6,module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2,module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da,module.incidents.integration.incident_socket_current_admission_697be7989c,module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972,module.incidents.unit.incident_admission_role_sets_and_denials_are_closed_0ed8398f44,module.incidents.integration.incident_create_persists_the_incident_row_bootst_0bd8d02f12,module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6,module.incidents.integration.real_incident_patch_persists_only_patchable_meta_9fc8fc207f,module.incidents.integration.live_membership_patch_to_the_current_role_return_7a14cba87a,module.incidents.support_integration.authenticated_incident_creation_and_normalized_r_e2cd9c99b5` | 90367; pass, exit 0 |
| `tr02-wf02-timeline-20261027-01` | `make test-slice OWNER=module.timeline ROWS=module.timeline.integration.incident_role_authorization_review_and_supersede_62f3d384f6` | 43257; pass, exit 0 |
| `tr02-wf02-workbook-20261027-01` | `make test-slice OWNER=module.workbook ROWS=module.workbook.catalog.keep_saved_clears_the_conflict_without_creating_c65a9c3099` | 86464; pass, exit 0 |
| `tr02-wf02-collaboration-20261027-01` | `make test-slice OWNER=module.collaboration ROWS=module.collaboration.support_integration.incident_socket_revocation_sources_00f908473e` | 88558; pass, exit 0 |
| `tr02-wf02-extension-20261027-01` | `make test-slice OWNER=app.server ROWS=app.server.unit.extension_discovery_transport_7f6ad80724` | 58512; pass, exit 0 |

D-001–004 close through the new groups, qualified before deletion at E-010/011
and after deletion here. D-005/006 close through the preservation portfolio,
including all five additional incident rows. D-007 closes by intentional removal
without fabricated successors; four meaningful extension selectors remain.
D-008 closes through the existing isolated application fixtures, explicit disabled
test routes, response/socket closure and exact canonical selection.

Removed thirteen incident entrypoints, the Timeline permission matrix, the
private interpreter, the shared inventory package and its six providers.
Timeline's real `CreateRow` helper and unrelated mutation-support helpers remain.
Bounded source searches find no remaining shared-inventory imports or callers.
The obsolete Entities path allowance was removed without widening imports.
Four retired rows have exact crosswalks; the extension row retains its identity.
No production API/schema or owner behavior changed; no correction slice was needed.
The saved checkpoint was reviewed for all twelve sections, stable IDs,
dispositions, literal commands, table closure and dependency order before WF-03.

### WF-03 checkpoint — E-013

The following commands used `CARTULARY_TEST_RUN_ID=<Run ID>`; exact roots use
the absolute run parent above. All native command exits are 0.

| Run ID | Exact target | Native session / result |
| --- | --- | --- |
| `tr02-wf03-generate-20261027-01` | `make generate` | 56371; pass, exit 0 |
| `tr02-wf03-generated-20261027-01` | `make source-check CHECK=generated` | 12067; pass, exit 0 |
| `tr02-wf03-schema-20261027-01` | `make source-check CHECK=schema` | 77775; pass, exit 0 |
| `tr02-wf03-catalog-20261027-01` | `make source-check CHECK=catalog` | 45679; pass, exit 0 |
| `tr02-wf03-identity-20261027-01` | `make source-check CHECK=identity` | 43939; pass, exit 0 |

Reviewed all seven rows: incident ownership and verification references, concrete
collaborators, standard aggregate placement, PostgreSQL/object-store dependencies,
dedicated PostgreSQL fixtures and exclusive isolation agree with actual execution.
The four retired rows map only to real successors. The extension row keeps four
meaningful selectors. No alias or orphaned reference remains. The generated diff
contains only the execution-topology render input digest and the three changed
test-family hashes; generation did not alter production contracts.
The saved checkpoint was directly reviewed for twelve sections, stable IDs,
closed assertion dispositions, exact commands and dependency order before WF-04.

### WF-04 source-freeze repair — E-014

Initial finalizer `tr02-wf04-finalize-20261027-01` passed (native session 24834,
exit 0), changed zero generated files, and explicitly skipped retained maintenance
because `RESULTS_DIR` was unset. The initial frozen source digest was
`sha256:8e91df08967ee846418c376d5d022934b2102a776e5e4990a21be88a6f87ad5a`.
All five final test invocations and Go lint passed on that source. Backend-boundary
lint `tr02-wf04-lint-backend-boundary-20261027-01` failed (native session 27023,
outer exit 2; unit exit 10 / row exit 1; normalized `product/test_assertion_failure`).
`make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-lint-backend-boundary-20261027-01`
returned exit 0 and reported cleanup: 13 steps, zero failed/blocked.

The bounded unit log identifies `forbidden_source_token` in
`authorization_observations_test.go`: the incident source-codec boundary forbids
whole-row `to_jsonb(i)` within the incident module. This was introduced in the new
test helper, not a production behavior defect. Replaced all three occurrences in
three observer functions with explicit incident identity, metadata, lifecycle,
version and attribution fields. No boundary exception was added or broadened.
Re-finalization and source-compatible requalification are required; earlier
passes remain historical rather than silently qualifying the repaired source.

### Final-source qualification — E-014

The source repair was formatted with
`CARTULARY_TEST_RUN_ID=tr02-wf04-repair-format-20261027-01 make format`
(native session 53913, exit 0). Re-finalization:
`CARTULARY_TEST_RUN_ID=tr02-wf04-finalize-20261027-02 make agent-finalize`
(native session 62579, exit 0). The finalizer reports zero changed generated
files and `results-dir-not-provided` for retained maintenance. No eligible full
warm-check root existed; `RESULTS_DIR` was unset and no redundant warm check ran.

Final executable source was frozen after that review at HEAD
`396a6d364ec0ce9807a6b9ad143964abc8a04363`, working-tree source digest
`sha256:a3febe51f08f1c1eb1f6332a460c776a6977ed0314d04d7a307e7673f92c79aa`.
All twelve finalizer/test/lint/source-check manifests share that digest.
Subsequent changes affect only this human tracker. No test, generator, runtime
metadata, conformance or release evidence reads or hashes Markdown.

Final commands below all used `CARTULARY_TEST_RUN_ID=<Run ID>`. Each exact root
is `/home/jochi/code/cartulary/.cartulary/test-results/<Run ID>`; owning summaries
are `run-summary.json` and `target-summaries/<target>.json` beneath that root.
All native exits and reported outcomes are pass/0, with no normalized failure.
The final run summaries report zero cache hits. Cleanup and retained evidence
were independently audited with the public read-only check below.

| Run ID | Exact target and selection | Native session / result |
| --- | --- | --- |
| `tr02-wf04-incidents-20261027-02` | `make test-slice OWNER=module.incidents ROWS=module.incidents.integration.incident_collection_authenticated_authorization_dab8552d8f,module.incidents.integration.incident_resource_current_authorization_b010ae008a,module.incidents.integration.incident_membership_current_authorization_5018f673df,module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6,module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2,module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da,module.incidents.integration.incident_socket_current_admission_697be7989c,module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972,module.incidents.unit.incident_admission_role_sets_and_denials_are_closed_0ed8398f44,module.incidents.integration.incident_create_persists_the_incident_row_bootst_0bd8d02f12,module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6,module.incidents.integration.real_incident_patch_persists_only_patchable_meta_9fc8fc207f,module.incidents.integration.live_membership_patch_to_the_current_role_return_7a14cba87a,module.incidents.support_integration.authenticated_incident_creation_and_normalized_r_e2cd9c99b5` | 24957; pass, exit 0 |
| `tr02-wf04-timeline-20261027-02` | `make test-slice OWNER=module.timeline ROWS=module.timeline.integration.incident_role_authorization_review_and_supersede_62f3d384f6` | 50105; pass, exit 0 |
| `tr02-wf04-workbook-20261027-02` | `make test-slice OWNER=module.workbook ROWS=module.workbook.catalog.keep_saved_clears_the_conflict_without_creating_c65a9c3099` | 37666; pass, exit 0 |
| `tr02-wf04-collaboration-20261027-02` | `make test-slice OWNER=module.collaboration ROWS=module.collaboration.support_integration.incident_socket_revocation_sources_00f908473e` | 81207; pass, exit 0 |
| `tr02-wf04-extension-20261027-02` | `make test-slice OWNER=app.server ROWS=app.server.unit.extension_discovery_transport_7f6ad80724` | 2340; pass, exit 0 |
| `tr02-wf04-lint-go-20261027-02` | `make lint CHECK=go` | 91522; pass, exit 0 |
| `tr02-wf04-lint-backend-boundary-20261027-02` | `make lint CHECK=backend-boundary` | 49366; pass, exit 0 |
| `tr02-wf04-generated-20261027-01` | `make source-check CHECK=generated` | 7463; pass, exit 0 |
| `tr02-wf04-schema-20261027-01` | `make source-check CHECK=schema` | 85219; pass, exit 0 |
| `tr02-wf04-catalog-20261027-01` | `make source-check CHECK=catalog` | 82111; pass, exit 0 |
| `tr02-wf04-identity-20261027-01` | `make source-check CHECK=identity` | 28421; pass, exit 0 |

Portfolio obligations: the fourteen incident rows cover seven replacement groups
plus seven independent incident guarantees; Timeline adds source lifecycle,
Workbook adds route guards and its retained conflict assertions, Collaboration
adds established-socket revocation/isolation, and Extension Discovery adds its
four meaningful transport/contract selectors. This is eighteen selected test rows.
The four source checks add routing/schema/identity/generated consistency, and
the lint checks add Go and module-boundary assurance. Final runs repeat earlier
qualification only to bind the complete required portfolio to the repaired source.

Retained-run audits (native commands returned immediately; no outstanding session):

| Exact command | Native result |
| --- | --- |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-01-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-finalize-20261027-02 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-02-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-incidents-20261027-02 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-03-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-timeline-20261027-02 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-04-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-workbook-20261027-02 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-05-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-collaboration-20261027-02 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-06-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-extension-20261027-02 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-07-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-lint-go-20261027-02 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-08-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-lint-backend-boundary-20261027-02 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-09-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-generated-20261027-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-10-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-schema-20261027-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-11-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-catalog-20261027-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf04-audit-12-20261027-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-identity-20261027-01 CHECK=all` | Native exit 0; `full_validity=1` |

Each audit reports `CHECK=all`, `PASS`, `full_validity=1`, `read_only=1`.
Manifest/source agreement, published outcomes and same-run cleanup are admitted.
Optional execution/resource diagnostics were not collected and remain unavailable;
they are not performance evidence. The audits establish validity for these exact
roots, not complete-owner, full-project, UI, performance or release acceptance.

For completeness, the initial final portfolio below ran on the earlier source
digest recorded in the repair entry. It is historical and is superseded for final
qualification by the compatible runs above. The backend-boundary failure was
investigated before repair; it was neither ignored nor relabeled as a production
defect. WF-01's authored cross-owner verification failure and this test-observer
failure are both resolved. No product behavior defect or primary-owner
contradiction was demonstrated; no conditional correction workstream was needed.

| Earlier run ID | Exact target and selection | Native session / result |
| --- | --- | --- |
| `tr02-wf04-incidents-20261027-01` | `make test-slice OWNER=module.incidents ROWS=module.incidents.integration.incident_collection_authenticated_authorization_dab8552d8f,module.incidents.integration.incident_resource_current_authorization_b010ae008a,module.incidents.integration.incident_membership_current_authorization_5018f673df,module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6,module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2,module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da,module.incidents.integration.incident_socket_current_admission_697be7989c,module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972,module.incidents.unit.incident_admission_role_sets_and_denials_are_closed_0ed8398f44,module.incidents.integration.incident_create_persists_the_incident_row_bootst_0bd8d02f12,module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6,module.incidents.integration.real_incident_patch_persists_only_patchable_meta_9fc8fc207f,module.incidents.integration.live_membership_patch_to_the_current_role_return_7a14cba87a,module.incidents.support_integration.authenticated_incident_creation_and_normalized_r_e2cd9c99b5` | 31173; pass, exit 0 |
| `tr02-wf04-timeline-20261027-01` | `make test-slice OWNER=module.timeline ROWS=module.timeline.integration.incident_role_authorization_review_and_supersede_62f3d384f6` | 40292; pass, exit 0 |
| `tr02-wf04-workbook-20261027-01` | `make test-slice OWNER=module.workbook ROWS=module.workbook.catalog.keep_saved_clears_the_conflict_without_creating_c65a9c3099` | 85917; pass, exit 0 |
| `tr02-wf04-collaboration-20261027-01` | `make test-slice OWNER=module.collaboration ROWS=module.collaboration.support_integration.incident_socket_revocation_sources_00f908473e` | 74245; pass, exit 0 |
| `tr02-wf04-extension-20261027-01` | `make test-slice OWNER=app.server ROWS=app.server.unit.extension_discovery_transport_7f6ad80724` | 59093; pass, exit 0 |
| `tr02-wf04-lint-go-20261027-01` | `make lint CHECK=go` | 99934; pass, exit 0 |
| `tr02-wf04-lint-backend-boundary-20261027-01` | `make lint CHECK=backend-boundary` | 27023; fail, exit 2 |

Final handoff compatibility: only private inventory interfaces and obsolete
selectors were removed. The actual row migration map is in Section 7; public
API, schema, roles, runtime modes and Make commands did not change. Independent
source-owner guarantees remain. All F-001–008 and D-001–008 map to their slices
and evidence in Sections 4–8. No required gate for this bounded remediation remains open. Document lint and
direct final tracker review completed V-006 before WF-04 closed.

### Validation and handoff completion — E-014 final checkpoint

`CARTULARY_TEST_RUN_ID=tr02-wf04-document-20261027-01 make lint-markdown`
passed (native session 82065, outer exit 0; owning `status=pass`, `exit_code=0`).
Exact summary:
`/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf04-document-20261027-01/adhoc/lint-markdown/tool-run-summary.json`.
The configured globs omit this handoff; the pass qualifies configured documents.
Direct final review separately checked twelve numbered sections, every table's
column count, all Markdown links, stable F/S/WF/D/T/E identities, acyclic sequence,
row crosswalks against authored JSON, exact commands/results and source identity,
assertion closure, current versus historical posture, exclusions and whitespace.
`git diff --check` passed; the untracked handoff was checked directly as well.

All required workstreams have saved checkpoints and complete slice statuses.
WF-04 finished last. The bounded effort is complete with no unresolved failure,
production defect, owner contradiction or pending implementation work. Findings
F-001–008 and assertion dispositions D-001–008 are closed by the evidence above.

Remaining limits: no full-owner/full-project, UI/visual, performance or release
qualification; Auth's separate framework and unrelated route families remain
out of scope. Optional runtime diagnostics were not collected. Retained full
warm-check maintenance was skipped because `RESULTS_DIR` was unset. These limits
do not replace any required bounded gate. Human review still confirms normative
projection fidelity. The changes remain uncommitted in the working tree; no
public API/schema migration or production consumer action is required.

### Historical document-only session — E-001–008

The following subsections describe only the original planning/save invocation.
Their authorization limits and future-work statements are superseded by E-009 onward.

### Scope and authority

| Session / baseline | State and evidence | Next action |
| --- | --- | --- |
| 2026-10-27 planning and document save; HEAD `396a6d364ec0ce9807a6b9ad143964abc8a04363` | E-001–003 established scope; E-007 matches prior complete content/file-identity digest; current request saves documentation only | Later actor revalidates source and authorization before WF-00 |

### Backend boundary

| Evidence | State / inspected surface | Next action |
| --- | --- | --- |
| E-003 | Twelve-file test seam and supporting application/admission/owner observations; no production changes | Use existing providers and keep private test helpers local |

### Frontend boundary

Not applicable: no frontend, browser, rendering, accessibility or golden change.

### Contract and codegen

| Evidence | State | Next action |
| --- | --- | --- |
| E-003 | Owners and exact authored rows inspected; no owner amendment needed; generated outputs untouched | Future authored rows and real crosswalks before Make regeneration |

### Tests and harness

| Evidence | Commands / state | Next action |
| --- | --- | --- |
| E-004/E-007 | No tests, builds, generators, services or finalizer executed; public command paths inspected | Future V-001 supplies characterization; old passes are not substitutes |
| E-008 | Markdown lint passed; direct handoff review and sole-file comparison passed | Documentation complete; implementation remains future work |

### Security and authorization

| Evidence | Required distinction | Next action |
| --- | --- | --- |
| E-003 | Literal expected permissions remain independent of implementation; editor/demotion gaps identified; real socket revocation and operation-guard evidence remain distinct | Qualify every security successor before retirement |

### Open risks and next session

| Evidence | Current limit | Next action |
| --- | --- | --- |
| E-007/008 | Structural plan is complete; no current test characterization; implementation permission absent | Begin WF-00 only under a later implementation task; preserve independent witnesses |

| Evidence ID | Source / observation | Applicability |
| --- | --- | --- |
| E-001 | Prior planning: clean branch/HEAD and output absence | Same source as save session |
| E-002 | User chose incident authorization rather than adding Auth's framework | Scope decision retained |
| E-003 | Static source, owner, caller, assertion and routing inspection described in Sections 2–8 | No execution claim; source unchanged at E-007 |
| E-004 | Prior planning ran no tests, builds, generators, services or finalizer | Historical session boundary only |
| E-005 | Prior final comparison: 7,666 unchanged tracked/non-ignored paths, index and HEAD | No ignored-tree audit |
| E-006 | User requested saving the proposed documentation plan after Plan mode ended | Supersedes the file-save blocker only, not future implementation restrictions |
| E-007 | Save-session baseline in `/tmp/cartulary-tr02-doc-chx2ukk1/baseline.json`; same content/file identities and index as E-005; destination absence rechecked | Current sole-file write baseline |
| E-008 | Public Markdown lint native exit 0; direct document review; `/tmp/cartulary-tr02-doc-chx2ukk1/comparison.json` | Sole-file save verified; no product qualification |

### Document validation and preservation — E-008

The sole repository write is this new handoff. Prior handoffs remain unchanged
and their lengthy historical execution logs are linked rather than copied.

Executed from the repository root:

```sh
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-tr02-doc-chx2ukk1/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-tr02-doc-chx2ukk1/static-analysis \
CARTULARY_TEST_RUN_ID=tr02-document-20261027-01 \
COREPACK_ENABLE_NETWORK=0 make lint-markdown
```

Native exit: **0**. The owning summary reports `status=pass`, `exit_code=0`,
with no failure class or reason. Summary:
`/tmp/cartulary-tr02-doc-chx2ukk1/results/tr02-document-20261027-01/adhoc/lint-markdown/tool-run-summary.json`.
Installed-only preparation checks existing dependencies; no install, build,
generation, service operation or test execution was performed.

The lint configuration excludes this handoff, so that pass qualifies configured
documents only. Supplemental review covered all twelve numbered sections, table
widths, local links, whitespace, stable IDs, workflow dependencies, assertion
dispositions and separation of completed planning from future work. Results are
retained in `/tmp/cartulary-tr02-doc-chx2ukk1/document-review.json`.
`git diff --check` passed for tracked changes; because this handoff is new and
untracked, its whitespace was checked separately during direct document review.
These are document checks, not product verification or test qualification.

The final comparison in `/tmp/cartulary-tr02-doc-chx2ukk1/comparison.json` verifies
unchanged content and file identity for all 7,666 original tracked/non-ignored
paths, unchanged HEAD/index, no removed paths, and exactly one added path: this
regular single-link handoff. Ignored trees were not exhaustively audited.
Product tests, formatting, generation, finalization and broader verification
were skipped as required by the document-only scope. Finalizer retained-run
maintenance is not applicable to this save; no full warm-check root was supplied.

## 11. Open Questions and Blockers

No unresolved implementation-design choice remains for the bounded plan.

| ID | Gate / affected work | Status / resolution |
| --- | --- | --- |
| RB-001 | Prior active Plan mode prevented saving | DONE; E-006 permits this document save; no longer an active blocker |
| RB-002 | Current pre-move characterization and qualified successors | DONE at E-009–011; V-001 and V-002 qualified before retirement; remaining workstream gates still apply |

WF-00 and S-001–005 qualified (E-009–014). S-005 finished last after repairing
the observer boundary issue, qualifying compatible source and completing the
handoff. No production correction, unresolved failure or owner-resolution blocker
is open. Human projection-fidelity review remains a separate review activity.

## 12. Binary Completion Criteria

| Outcome | Current statement |
| --- | --- |
| Planning completeness | complete for this bounded incident-authorization seam; not a whole-suite audit |
| Per-slice readiness | S-001–005 complete; every workstream exit qualified and checkpoint saved |
| Implementation authorization | Granted by the current implementation request; earlier planning-only restrictions are historical |
| Document status | Complete; E-014 records final Markdown lint, direct review and bounded handoff |
| Session write boundary | Historical sole-file save verified at E-008; current implementation changes the Section 2 test/routing seam; HEAD unchanged |

Historical documentation-save completion (E-008):

- [x] Target/output and current-source identity established.
- [x] Bounded file inventory and inspection limits explicit.
- [x] Production requirements separated from incidental test mechanisms.
- [x] Findings, contracts, dispositions, slices and validation connected.
- [x] Acyclic workflow order, entry gates, rollback and exits specified.
- [x] Prior authorization/results remain historical; no tests executed.
- [x] Documentation checks and actual coverage limits recorded.
- [x] Final comparison verifies only the new handoff changed.

Implementation completion:

- [x] Every removed assertion has a qualified successor or incidental retirement rationale.
- [x] Retained operations cover all roles, immediate changes and deployment-admin isolation.
- [x] Denied mutations leave protected state unchanged; authorized outcomes persist.
- [x] Shared inventories, role interpretation and route-name fixture repairs are gone.
- [x] Duplicate Timeline permission execution is removed without losing reviewed-target coverage.
- [x] Independent production witnesses remain intact.
- [x] Authored/generated routing accurately describes executed evidence without aliases.
- [x] Final results, all workstream checkpoints, exclusions and maintenance skips are recorded.

Historical planning/document-save session (E-001–008) executed **no tests**.
Current implementation execution, routing and generated changes are recorded at
E-009 onward; no production behavior or normative owner amendment was needed.
