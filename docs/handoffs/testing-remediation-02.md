# Testing Remediation 02: Incident Completion and Auth Simplification Plan

## 1. Scope and Source Posture

### Current Auth implementation — E-017 onward

The 2026-10-28 request to implement the proposed plan authorizes the bounded
Auth workstreams WF-06–12, including their tests, authored routing, Make
generation/finalization and checkpoint updates. Earlier document-only limits
below describe their historical sessions and do not restrict this implementation.
No staging, commit or push is requested. WF-12 must finish last.

Starting HEAD is `53bdd5341304edd005c451b05178861eb02c51fc`; the only pre-existing
working-tree change is this tracker. Its existing Auth plan is preserved.
Content/file-identity evidence for 7,664 tracked/non-ignored paths, the index,
HEAD and the original tracker is retained at
`/tmp/cartulary-tr02-auth-impl-yeuvuwz3/baseline.json` and `tracker-before.md`.
Ignored build/cache/service outputs are outside that preservation comparison.

The bounded Auth effort is **complete**. WF-06–12 / S-006–012 are DONE at
E-017–023, with a saved, validated checkpoint before every successor. WF-12
finished last. The final 22-row portfolio, both lints, four source checks,
eight retained-root audits and final document validation pass.
Current-source characterization qualified. No production defect
or specification amendment is established. Sections 6–9 define the sequence;
Section 10 records actual execution and supersedes historical outcome statements.

### Auth planning/document-save iteration — historical baseline

Document date: **2026-10-27, America/New_York**. The current request saves the
next iteration's plan only. No Auth implementation, test execution, generation,
formatting, finalization, staging, commit or push is authorized by this document
step. The earlier incident implementation request and E-009–014 remain historical.

The supplied repository target remains `/home/jochi/code/cartulary`, canonical
repository-relative identity `.`, a directory. The controlling artifact remains
`docs/handoffs/testing-remediation-02.md`; its twelve numbered sections and
existing IDs are retained. Artifact mode: **normal**. The selected next seam is
**Auth's route-inventory framework and its consumers**. This selection supersedes
Auth's earlier deferral only for this new iteration; it does not reopen completed
incident slices. Baseline commit `53bdd5341304edd005c451b05178861eb02c51fc`
contains the completed incident work, which finished with WF-04.

Consolidate the eight Auth inventory-driven integration suites into the existing
behavioral tests, adding missing observations before deletion. Do not replace
one route interpreter with another comprehensive matrix. Remove incidental test
constraints rather than turning them into production or harness requirements.
No public API, schema, credential policy, role vocabulary, runtime mode or Make
interface change is planned. Any demonstrated production defect requires a
separate owner-backed correction workstream before dependent movement.

Current exclusions: enterprise-auth redesign, current-account features, incident
authorization redesign, broad HTTP-helper cleanup, shared harness lifecycle
changes, UI, performance and release qualification. Shared helpers with real
audit or enterprise consumers remain protected; retaining a mixed selector does
not expand the redesign scope.

The framework, skill format, root/backend/tooling instructions, Core 00 owner
allocation, Core 01 credential/user/transport owners, Core 04 security owners and
adopted/current Testing Harness NLSpec were inspected during planning. Inspection
has established no primary-owner contradiction or need to amend specifications.
Typed projections and tests remain evidence rather than requirements. Human
review confirms projection fidelity. No executable component may read, stat,
hash or otherwise consume this tracker or other Markdown as product evidence.

**Planning completeness:** complete for this bounded seam. **Slice readiness:**
S-006 is ready for a later authorized characterization task; S-007–012 are blocked
on their explicit characterization/predecessor gates. **Implementation
authorization:** not granted by this planning/document invocation. All Auth
workstreams start TODO. The sequence is WF-06 through WF-12, with a saved,
validated checkpoint after every workstream and before its successor.

| Baseline item | Starting evidence / identity | Comparison obligation | Limits |
| --- | --- | --- | --- |
| Checkout | Clean `main`; HEAD `53bdd5341304edd005c451b05178861eb02c51fc` | E-016 compares HEAD, index and working files after the save | Incident execution evidence stays historical; no current Auth pass inferred |
| Index | SHA-256 `fd8eb474144eb4cd5b382d4995f31db8ffbac5c95badab60606056a9be796b2c` | Must remain unchanged | Git inspection uses `--no-optional-locks` |
| Working files | 7,664 tracked/non-ignored paths; clean staged/unstaged/untracked status | Content and file-identity comparison excludes only this tracker | Enumeration is not whole-repository behavioral inspection |
| Existing output | Regular single-link file, inode `107764`, 85,196 bytes; SHA-256 `0c2049dd2353f2426336b514d18975514205cb5657b3104d02ec07394db474af` | Identity, target and content rechecked immediately before writing | No symlinked output/parent component; same canonical target |
| Comparison material | `/tmp/cartulary-tr02-auth-doc-j5vkqn7x/baseline.json` and `snapshot.py` | Final comparison retained beside baseline | Hashes/file metadata only; no copied credentials |
| Ignored output trees | Dependencies, caches and retained runs excluded from source enumeration | Documentation command artifacts redirected outside the repository | No exhaustive ignored-tree audit; no dependency installation or service work |

The incident scope, rationale and baseline below are preserved as historical
records, not authorization for the current document-only invocation.

### Incident iteration — completed record

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

### Auth baseline movement and preservation inventory — E-015/E-016

The declared movement universe is three tracked, authored regular Go files:
the Auth inventory provider and its two consumers. All three were directly read
at every declaration/assertion proposed for replacement; the provider and support
integration file were read in full. The handler file is mixed: its inventory loop,
payload/dispatch helpers and relevant retained handler tests were inspected;
unrelated bodies are outside this bounded behavioral audit. Supporting files
below are explicitly accounted for. No generated, vendor, binary, symlink or
submodule entry is proposed for code movement.

| Path / kind | Inspection and responsibility | Callers / dependencies | Disposition and risk |
| --- | --- | --- | --- |
| `internal/modules/auth/testsupport/routetest/routes.go` | Direct, full: 15 HTTP operations plus socket metadata, eight harness classes, validation and selection | Exactly two known repository consumers below | Delete after qualified migration; F-009/012 |
| `internal/modules/auth/support_integration_test.go` | Direct, full: eight generic suites, `SupportRouteContext`, request builders, broad denial hash; two independent user-list tests and helpers | Existing application/HTTP/socket support and Auth lifecycle helpers | Split meaningful observations into owners; retain live pagination/search and shared helper consumers; F-010/011/013 |
| `internal/modules/auth/handlers_test.go` | Direct at inventory CSRF loop, payload/handler switches, store/hub stubs and retained handler contracts | Auth inventory; concrete handlers | Add direct-case successor before deleting old loop/helpers; preserve unrelated tests; F-009/012 |
| `internal/modules/auth/integration_test.go` | Direct at the fourteen lifecycle, bootstrap, administration, audit and replay tests, plus fixture/audit/receipt helpers | Existing application runtime, HTTP, semantic sockets and real database | Primary consolidation destination; returned versions and explicit prerequisites; F-011/013 |
| `internal/modules/auth/admin_audit_integration_test.go`; `admin_audit_contract_test.go` in the same directory | Direct at route authorization, envelope and OpenAPI witnesses | Audit fixtures, list helpers and published route contract | Preserve malformed-query authorization ordering, pagination and exact OpenAPI evidence |
| `internal/modules/auth/secret_envelopes_integration_test.go`; `internal/platform/authn/secrets_test.go` | Direct: TOTP/enterprise substitution, activation re-sealing, purpose/record binding and cryptographic framing rejection | Real credential flows and Auth primitive | Keep storage-format assurance here; existing mixed enterprise selector remains |
| Auth `enterprise_auth_integration_test.go`, `testsupport/flowtest`, existing application/HTTP/socket helpers | Direct at relevant shared-helper references and established fixture/transport boundaries; other bodies excluded | `lookupUserAuditEvents`, actor/incident setup, runtime ownership and semantic socket client | Preserve real consumers; no broad helper or enterprise rewrite |
| `tools/test_families/module.auth.json`; `contracts/verification/owners/module.auth.json`; `tools/test_catalog_row_migrations.json` | Relevant rows, references and migration shape directly inspected; unrelated rows metadata-only | Authored selectors feed generated projections | Update authored inputs only in future slices; F-014 |
| Generated policy/topology, task-surface owner/projections, Make recipes, pin registry, Markdown configuration | Relevant policy/command inputs inspected; generated files metadata-only | Existing public Make surface | Discover/execute through Make; no manual projection edits or document-fed selectors |
| Other tracked/non-ignored files | Enumerated for preservation, excluded from behavioral audit | Outside the selected seam | No whole-Auth or whole-repository completeness claim |
| Ignored caches, dependencies, build/service outputs | Excluded from source inventory with scope-based reason | No implementation or runtime operation in this session | No ignored-tree preservation claim beyond admitted command effects |

Bounded searches used `rg` for inventory symbols/imports, direct source reads,
JSON selector inspection and Git identity checks. Two known inventory consumers
were found; text search is not proof about arbitrary external/dynamic consumers.
WF-06 revalidated callers and source at E-017. E-015/016 remain static planning
evidence only; current execution results are E-017–023.

Final Auth implementation surfaces:

| Path | Final responsibility / disposition |
| --- | --- |
| `internal/modules/auth/csrf_test.go` | Direct cookie-mutation CSRF handler cases; no incidental session-read ordering constraint |
| `internal/modules/auth/authorization_integration_test.go` | Eight current-capability scenarios and scoped administrative state observations |
| `internal/modules/auth/integration_test.go` | Consolidated lifecycle, bootstrap, replay, normalization, audit and revocation guarantees |
| `internal/modules/auth/handlers_test.go` | Independent handler tests retained; old inventory loop and dispatch helpers removed |
| `internal/modules/auth/support_integration_test.go` | Both user-list tests and four live helpers retained; generic suites/builders/hash removed |
| `internal/modules/auth/testsupport/routetest/routes.go` | Deleted after qualification; no surviving callers or compatibility facade |
| `tools/test_families/module.auth.json`; `tools/test_catalog_row_migrations.json` | Two real successors, one unit-row retirement, pagination selector/reference cleanup |
| `tools/execution_topology_render_index.json` | Make-regenerated input digest and Auth manifest hash |

Audit/OpenAPI, dedicated secrets and shared enterprise helper sources remain
unchanged. No production source was edited. See E-023 for final scope comparison.

### Incident iteration — completed record

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

### Auth responsibility decisions

| Responsibility | Current location | Durable owner | Decision | Trace |
| --- | --- | --- | --- | --- |
| Route inventory and category-driven execution | Auth routetest and two consumers | No production responsibility | Remove; direct requests/handler cases in behavioral tests | F-009/012; C-006/007/011 |
| Local credential/bootstrap/session behavior | Generic suites plus existing lifecycle tests | Auth lifecycle tests | Consolidate into existing tests with missing observations added | F-010/011; C-008/010 |
| User administration and current deployment capability | Generic suites plus existing administration/audit tests | Auth administrative tests | Keep distinct lifecycle/audit evidence; add one explicit current-authorization test | F-011/013; C-006/009 |
| Storage cryptography | Generic HTTP byte-format checks and dedicated secret tests | Auth secret integration and authn primitive tests | Keep qualified binding/framing/re-sealing witnesses at their source owner | F-012; C-008 |
| Fixture composition, HTTP and socket mechanics | Existing support packages plus universal context | Existing concrete helpers and operation-local arrangement | Delete context/switch machinery; retain demonstrated shared consumers | F-010/013; C-007/010 |
| Exact selection and evidence routing | Mixed support row and inventory-named unit row | Auth verification/catalog owner | Preserve live identities, retire only obsolete identity, record real crosswalks | F-014; C-011 |

No domain behavior moves into platform transport or application composition.
Use existing helpers for concrete setup/observation; a local literal case table
must not become a route registry, policy interpreter or generic snapshot engine.
The historical Auth deferral below was superseded by E-015's scope selection.

### Incident iteration — completed record

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

### Auth contract and behavior map

| Contract / surface | Observed evidence | Adopted owner and required behavior | Disposition / prerequisites | Trace |
| --- | --- | --- | --- | --- |
| C-006: current deployment authorization | Generic success/demote/deny loop; independent static admin/incident-admin denials | Core 04 REQ-04-028/030 and administrative-audit authorization: current `deployment_admin`, distinct 401/403, no incident-role substitute | Preserve through eight explicit operations using the same subject session and real demotion; qualify before deletion | F-009/011/013; S-009; V-010/011 |
| C-007: transport and request protection | Fifteen success envelopes; nine CSRF operations; four invalid vocabulary cases; normalization cases | Core 01 common HTTP envelopes, local login, §§3.3.2.2/3.3.5.1; Core 04 REQ-04-003: exact wire contracts, safe request IDs, fail-closed cookie mutations | Preserve direct wire and handler observations; do not freeze internal lookup counts; literal expected values | F-009/012; S-007–009; V-008–011 |
| C-008: credential, replay and secret integrity | Generic replay suite plus stronger existing replay, lifecycle and bound-secret witnesses | Core 01 REQ-01-122/123/522–529; Core 04 REQ-04-084/165: original replay payloads, conflicts, scoped TOTP activation, safe payloads and protected bound storage | Consolidate with single-receipt/pending-state evidence; dedicated secret tests own framing/binding; no format/profile change | F-010–012; S-008/009; V-009–011 |
| C-009: durable administration and attribution | Existing user/admin audit tests, version/last-admin tests; broad denial hash omits audit effects | Core 01 user/credential mutation contracts including REQ-01-124/125/527–529; Core 04 administrative audit and credential guards: authorized material writes, safe attribution, no unauthorized changes | Strengthen scoped denied-state and complete before/after audit observations; carry returned versions | F-010/011/013; S-009; V-010/011 |
| C-010: sessions and sockets | Real logout/password/admin-action revocation; existing concurrency tests; one generic socket branch only connects | Core 04 REQ-04-005–016; Core 01 §3.3.10: session lifecycle, immediate revocation and semantic socket outcomes | Preserve distinct lifecycle witnesses; add attached-socket self-password observation; retire redundant admission-only branch | F-010/011/013; S-008/009; V-009–011 |
| C-011: truthful evidence accounting | Mixed full-placement support row, inventory unit row and focused standard rows | Adopted/current Testing Harness REQ-012–016: immutable identities, exact selectors, actual dependencies and real crosswalks without aliases | Author new rows before execution, reconcile surviving references and regenerate through Make; no Markdown inputs | F-009/014; S-006–012; V-007–014 |

Core 00–04 and the adopted/current harness owner retain their existing authority.
Normative behavior, observed assertions and missing execution evidence are separate
columns. There is no proposed specification amendment. Existing DB-read ordering,
route taxonomy and universal fixture shape are not public compatibility contracts.
The old row ID containing `argon2id` is an immutable identity, not authority for
password algorithms; the current cryptographic owner governs actual behavior.
UI, saved-view, revision, schema and release-interface changes are not applicable.

### Incident iteration — completed record

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

### Auth findings and remediation decisions — all scoped must_fix

All six findings are `must_fix` for the named Auth slices only. They concern tests,
private support, verification routing and documentation; no specification or
production change is planned. E-015 supplies static evidence; E-016 revalidates
its unchanged source. F-009–014 are resolved by E-018–023; the criteria below
remain the acceptance rationale, and E-023 records their actual closure.

| Finding / remediation | Rationale and long-term benefit | Compatibility / migration impact | Risk if unresolved | Completion criterion / trace |
| --- | --- | --- | --- | --- |
| F-009: parallel route model — remove inventory, category maps, path replacement, self-validation and dispatch interpreter | One understandable owner per guarantee; future routes extend behavior tests rather than a second execution model | Delete private interfaces without wrappers/aliases | Metadata can drift while tests validate their own model | Both consumers migrate and provider/callers disappear; C-006/007/011; S-007–011; V-008–012 |
| F-010: universal setup — replace context and operation-name switches with concrete helpers and local arrangement | Visible prerequisites, smaller fixtures and fewer unrelated failures | Test arrangement only; actual services and cleanup ownership remain | Shared setup, generated identity sequence and hidden target repairs obscure failures | No universal two-admin fixture, framework secret pool or route-name repair; C-008–010; S-008–010; V-009–011 |
| F-011: redundant execution — move unique assertions into existing behavioral tests before deleting generic suites | Fewer overlapping suites and clearer failure ownership | Preserve focused test identities where responsibility continues | Duplication drifts and retirement can lose unique prerequisites | Every old assertion has a qualified successor or incidental disposition; C-006–010; S-008–010; V-009–011 |
| F-012: implementation-derived expectations — use literal expected normalization and direct handler cases; retire no-session-lookup constraint; keep format details with secret tests | Independent expectations protect meaningful behavior without freezing internal order | Private constraints may disappear; no wire or security relaxation | Shared normalization can reproduce mistakes; internal constraints block clean changes | CSRF rejection/no harmful effects and owner-local crypto witnesses qualify; C-007/008; S-007–010; V-008–011 |
| F-013: coarse state evidence — replace global hash with scoped user/credential/session/receipt/audit observations; close responses | Stronger integrity evidence, useful diagnostics and less housekeeping sensitivity | No generic DB snapshot API or production hooks | Correct statuses can conceal unauthorized writes, revocation or audit effects | Denials preserve relevant state; material writes persist with versions/attribution; C-006/009/010; S-008–010; V-009–011 |
| F-014: mixed routing — remove retired selectors, retain pagination identity, reconcile verification references and author two real successors | Executable evidence matches its catalog claims | Immutable surviving IDs; one actual unit-row retirement; no fabricated successors or aliases | Dangling selectors and overstated evidence survive cleanup | Authored/generated selector, owner, dependency and identity checks pass; C-011; S-007/009–012; V-008/010–014 |

### Auth assertion migration ledger

D-009–018 are **DONE at E-021**, following successor qualification at E-018–020.
E-021 records the actual successor tests, exact rows, source-bound passing root
and native exit. The destination table below preserves the original obligations.

| ID | Existing suite / infrastructure | Required destination and retained observations |
| --- | --- | --- |
| D-009 | `TestSurfaceEnvelope` | Existing login/session, credential-state, user lifecycle, admin-action and deployment-audit tests: all fifteen exact success statuses, envelopes and request-ID relationships |
| D-010 | `TestBootstrapBoundaries` | Expand `TestBootstrapTokenRouteBoundaries_Integration` to twelve selected HTTP operations plus socket rejection; retain existing incident and privileged-test-route cases |
| D-011 | `TestCSRFProtection` | Real-route rejections immediately before valid operations in owning lifecycle tests; all nine cookie mutations, including malformed password/TOTP requests |
| D-012 | `TestReplayAndStoredPayloadSafety` | Existing user-create/password/admin-reset/revoke-all replay and TOTP lifecycle tests; dedicated bound-secret tests retain storage assurance |
| D-013 | `TestAuditAttribution` | Existing user/admin credential audit tests: actual transaction IDs, request attribution, timestamps and full forbidden-secret checks in before/after payloads |
| D-014 | `TestSessionRevocation` | Existing logout/concurrency and admin-action socket tests; self-password attached-socket observation added; connect/hello-only branch retired as redundant admission evidence |
| D-015 | `TestAuthorizationReDerivation` | New `TestDeploymentAdminAuthorizationRefresh_Integration`: eight direct success/demotion/denial scenarios with scoped integrity observations |
| D-016 | `TestRequestContracts` | Literal invalid vocabulary and writable-string outcomes in owning login/credential/user/replay tests; no production normalizer constructs expected values |
| D-017 | `TestCSRFProtectionRoutes_Unit` | New `TestCookieMutationCSRF_Unit`: local method/path/body/direct-handler cases; missing/invalid headers plus three malformed-body cases; no mutation/revocation requirement retained, exact lookup-count constraint retired |
| D-018 | Provider, builders, fixture switches, global snapshot and mixed helpers | Delete framework-only machinery; preserve both user-list tests and helpers with audit/enterprise consumers; retain pagination row and correct its references |

The framework's metadata self-checks, categories, path substitution and handler
switches have no independent production guarantee and need no replacement test.
Secret framing stays with the dedicated primitive/binding witnesses, not repeated
HTTP byte counts. A shorter suite alone is not evidence of successful migration.

### Incident iteration — completed record

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

### Auth workstreams — execution sequence

The following prerequisite column is authoritative for the Auth iteration.
Section 9 records current execution statuses and Section 10 records checkpoints.
Each slice is a separate sequential workstream. The completed incident sequence
below remains historical and is not a fresh prerequisite execution.

| Workflow / slice | Class / prerequisite | Goal / likely files | Risk and control | Validation / binary exit |
| --- | --- | --- | --- | --- |
| WF-06 / S-006: characterize | root; later implementation authorization and source revalidation | Reconcile owner clauses, callers and D-009–018; baseline exact Section 8 rows | Historical results are not current; classify failures against owners | V-007; exact commands, exits, roots, source identity and dispositions saved |
| WF-07 / S-007: direct CSRF unit cases | chain; WF-06 | Add explicit handler successor and authored row; predecessor remains | Losing guard cases; enumerate nine mutations and three malformed bodies | V-008; successor rejection/no harmful effects pass and checkpoint saved |
| WF-08 / S-008: credentials and sessions | chain; WF-07 | Strengthen existing login/bootstrap/TOTP/password replay/socket tests | Invalid later prerequisites; explicit lifecycle arrangement and legitimate reauthentication | V-009; credential/replay/secret/bootstrap/CSRF/revocation successors pass |
| WF-09 / S-009: administration | chain; WF-08 | Strengthen user/action/audit tests; add current-capability test and row | False denials from stale targets; valid targets, real demotion, scoped effects | V-010; literal outcomes and durable integrity qualify |
| WF-10 / S-010: retire framework | chain; WF-09 | Delete eight suites, old unit loop, provider and unused machinery; selectors updated with removals | Lost shared consumers; closed dispositions and caller search before deletion | V-011; D-009–018 close, preserved portfolio passes, no dead callers |
| WF-11 / S-011: reconcile routing | chain; WF-10 | Owners/references/services/crosswalks and generated projections | Remaining rows overstate evidence; review actual selectors against claims | V-012; generated/schema/catalog/identity checks pass without aliases/orphans |
| WF-12 / S-012: qualify and hand off | chain; WF-11 and every applicable correction workstream | Finalize, inspect generated diff, freeze source, final portfolio/audits and handoff | Stale-source or premature completion; bind all evidence to final source | V-013/014; all gates and checkpoints recorded; this workstream finishes last |

After each workstream update status, slice readiness, assertion dispositions,
commands/native session handles/outer exits/result roots/source identity, failures,
blockers and next action. Validate the saved document before starting its successor.
Never mark DONE from test presence or an unqualified passing summary.

A confirmed production defect conditionally adds one new workflow and slice after
the reserved IDs above. Before implementation record its owner clause, reproducer,
affected interfaces, correction, compatibility impact and targeted validation.
Keep correction in the existing production owner/layer. Complete it before resuming
dependent work; do not normalize failures in fixtures. This document step grants
no correction authorization. Unresolved primary-owner contradictions block only
dependent work until resolved and adopted; a failed test alone cannot amend an owner.

### Incident iteration — completed record

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

### Auth slices and readiness

Section 9 records slice execution status. The readiness column below describes
the original entry gates; Section 10 records their actual satisfaction.
RB-004's current-characterization gate closed at E-017. Successor qualification
and each predecessor checkpoint remain required before retirement.

| Slice / kind | Findings / contracts | Entry / readiness | Required change and exit | Validation / rollback |
| --- | --- | --- | --- | --- |
| S-006: characterization | F-009–014; C-006–011 | ready for later authorized baseline discovery; E-015/016; no code movement | Revalidate baseline/owners/callers; record exact current portfolio and failure classifications | V-007; no code to roll back; retain evidence and stop dependent movement on failure |
| S-007: structural unit simplification | F-009/012/014; C-007/011 | blocked: RB-004 and WF-06 | Add direct CSRF successor alongside old test; qualify all cases | V-008; revert successor and its authored routing together if needed |
| S-008: structural credential consolidation | F-010–013; C-007/008/010 | blocked: WF-07 and RB-004 | Strengthen focused credential/session tests; generic predecessors remain | V-009; restore changed tests/helpers together without weakening owner outcomes |
| S-009: structural administration consolidation | F-009–014; C-006–011 | blocked: WF-08 | Strengthen admin/audit/replay tests and add current-authorization successor | V-010; restore affected scenarios/observers and authored new row together |
| S-010: consolidation and incidental retirement | F-009–014; C-006–011 | blocked: WF-09 and qualified D-009–017 successors | Delete old framework only when every assertion has disposition; retain pagination and shared helpers | V-011; restore predecessor code and selectors together; no compatibility layer |
| S-011: routing reconciliation | F-014; C-011 | blocked: WF-10 | Reconcile metadata/crosswalks, regenerate, verify source checks | V-012; revert authored inputs and regenerate through Make, never patch projections |
| S-012: final validation/handoff | F-009–014; C-006–011 | blocked: WF-11 and applicable correction exits | Finalize/freeze/qualify/audit; finish tracker and handoff last | V-013/014; reopen only affected obligations after repair; no premature completion |

#### S-007: direct unit scenarios

Introduce `TestCookieMutationCSRF_Unit` using a local literal table with method,
path, valid body and direct handler function. Cover logout, password change, TOTP
begin/complete, user create/patch, admin password reset, TOTP reset and revoke-all.
For each, missing and invalid CSRF headers return `403 csrf_verification_failed`.
Retain malformed-body cases for password change and TOTP begin/complete. Permit
legitimate read-only session lookup; use existing stubs to reject mutations and
assert no revocation. Delete neither predecessor nor inventory until S-010.

#### S-008: credential and session changes

- Place login envelope/invalid `second_factor.kind` observations in the existing login tests; retain password exactness, session timing and concurrency semantics.
- Expand `TestBootstrapTokenRouteBoundaries_Integration` with explicit wrong-route cases for session, logout, credential-state, password change, administrative audit, users list/create/get/patch, password reset, TOTP reset and revoke-all, plus semantic socket rejection. Expected HTTP result is `409 credential_bootstrap_rejected` with `reason_code=not_allowed_for_route`. Keep existing incident/test-route cases.
- Add cookie CSRF rejection before valid logout, password change and session-authenticated TOTP begin/complete. Preserve bearer/session/bootstrap distinctions and malformed-body cases; no new global setup.
- Preserve TOTP pending replay and seed identity, one pending enrollment/activation, safe complete/read payloads, consumed-token rejection including repeat complete, no automatic session cookie, first enrollment versus replacement revocation, and current factor login behavior.
- Add a real attached socket before successful self-password change and assert semantic revocation. Retain existing logout/concurrency isolation and admin-action socket evidence rather than duplicating it.
- Keep password replay's required fresh login after revocation, original response, changed-request conflict, one receipt and no cleartext credential storage.
- Use dedicated `TestTOTPEnvelopeSubstitutionAndActivation_Integration` and bound-secret primitive tests for storage binding, framing and re-sealing; strengthen those owner-local observations if needed before deleting duplicate HTTP format checks.

#### S-009: administrative changes

Use existing user lifecycle, user/admin audit and four replay witnesses as the
owning scenarios. Check literal trimmed email/display-name results for create
and patch. Keep initial create `201`, exact replay `200`, later-state original
replay payloads, divergent conflicts, stale versions and last-active-admin guards.
For password reset, TOTP reset and revoke-all, send padded reasons initially and
literal normalized reasons on replay; preserve distinct divergent-reason tests.
Add real-route missing-CSRF checks before valid user create/patch and the three
admin actions. The four closed-vocabulary cases use literal `400` outcomes:
`second_factor.kind=webauthn` on login, password change and TOTP begin returns
`invalid_auth_request`; user creation with `auth_kind=ldap` returns
`invalid_mutation_payload`. Retain the attributable error field for each case.
Assert actual transaction IDs, request-ID relationships, actor,
source, time and forbidden-secret fields in audit before/after payloads. Close
all response bodies, including revoked-session checks.

`TestDeploymentAdminAuthorizationRefresh_Integration` has eight direct operation
scenarios: deployment-audit read, users list/create/get/patch, password reset,
TOTP reset and revoke-all. Use an isolated application/database per operation,
reused across its success/demotion/denial stages. Arrange only required actors,
targets, credentials and sessions. A separate administrator changes the subject
through real user PATCH with its current returned version. Reuse the subject's
same session immediately; every denied operation returns `403 authorization_denied`
without relogin, sleeps or retry acceptance. Reads/successes use their exact status;
create uses `201`, other listed successes `200`.

Fresh valid targets and returned versions must distinguish authorization from
stale-version, consumed-factor, inactive-user or no-op outcomes. Snapshot after
arrangement and demotion. Observe target user fields/versions, credentials/pending
enrollment, session revocation, route receipts, audit rows and attributable
administrative projections as applicable. For denied creation, scope absence to
the requested identity and transaction. Exclude legitimate session bookkeeping.
Use concrete local observations, not a generalized DB snapshot framework.

#### S-010/S-011: actual interface and row migration

Remove the eight generic integration entrypoints, old CSRF unit entrypoint,
`SupportRouteContext`, `SupportRequest`, operation-name builders/dispatch, global
denial hash and Auth inventory provider after successors qualify. Keep both
user-list tests and helpers with live audit/enterprise consumers. The existing
application runtime, HTTP helpers, semantic socket client and resource cleanup
ownership remain. There are no new production hooks or test runtime modes.

| Current row / selector | Planned migration |
| --- | --- |
| `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` | Same identity and `full` placement; remove eight retired selectors, retain `TestUserListContinuationUsesLiveRows`; remove unsupported openapi/audit verification reference after confirming dedicated witnesses retain that evidence |
| `module.auth.unit.an_inventory_driven_loop_covers_every_cookie_aut_0b6e53ff68` | Retire only after new `TestCookieMutationCSRF_Unit` row qualifies; record old-to-real-new row crosswalk |
| Existing focused integration/handler/secret/audit/list rows | Keep identities and meaningful selectors; add observations inside the owning tests, not copied duplicate suites |
| New deployment authorization test | New semantic integration row for `TestDeploymentAdminAuthorizationRefresh_Integration`; D-015 records selector migration from still-live mixed support row, not a fictional whole-row retirement |

Allocate the two new immutable IDs through public authoring, not guessed hashes:

```sh
make author-test-row-id FAMILY_ID=module.auth.unit CLAIM='Cookie mutations reject missing or invalid CSRF without mutation or revocation' SELECTOR_KEY='TestCookieMutationCSRF_Unit'
make author-test-row-id FAMILY_ID=module.auth.integration CLAIM='Administrative routes require current deployment capability after same-session demotion' SELECTOR_KEY='TestDeploymentAdminAuthorizationRefresh_Integration'
```

Author selectors/references before Make generation and canonical execution. New
unit placement is `fast` with no service fixture; new integration placement is
`standard`, runtime `default`, resource `standard`, `postgres_dedicated`, actual
`object_store` and `postgres` dependencies and `exclusive` PostgreSQL isolation.
Use Auth-owned verification references; preserve actual collaborators. No aliases,
wrapper interfaces, orphaned references or invented replacement checks are allowed.

### Incident iteration — completed record

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

### Auth validation portfolio — obligations and execution evidence

V-007–013 define the execution obligations qualified in E-017–023. V-014 covers
document save/review and the final handoff. All commands run from `/home/jochi/code/cartulary` through
public Make targets. Refresh discovery and relevant guidance before execution;
use a fresh explicit `CARTULARY_TEST_RUN_ID` per invocation. Capture native session
handle, outer exit, actual owning summary/root, source identity and failures.
A row identifier is a selector, not a statement of specification completeness.

| Validation | Stage / trace | Command / exact selection rule | Required result and failure consequence |
| --- | --- | --- | --- |
| V-007 | Pre-move; S-006; C-006–011 | `make help ALL=1`; `make task-guide OWNER=module.auth DETAIL=coverage JSON=1`; `make test-slice OWNER=module.auth ROWS=<comma-separated current rows below>` | Current predecessor and preservation baseline; classify failures before retry or movement; closes RB-004 only on qualified evidence |
| V-008 | Post-slice; S-007; C-007/011 | New unit row via `make test-slice OWNER=module.auth ROWS=<allocated CSRF row ID>`; retain predecessor for comparison | Nine missing/invalid cases and three malformed cases pass without mutation/revocation |
| V-009 | Post-slice; S-008; C-007/008/010 | `make test-slice OWNER=module.auth ROWS=<credential, bootstrap, login, socket and secret rows below>` | All migrated credential/session observations qualify; no predecessor retired |
| V-010 | Post-slice; S-009; C-006–011 | Same command selecting user/admin/audit rows below plus allocated authorization row | Eight current-capability scenarios, meaningful persistence, denial integrity and audit/replay guarantees pass |
| V-011 | Retirement; S-010; D-009–018 | Same command selecting all surviving rows below plus both allocated rows; old unit row excluded only after actual retirement | Every disposition closes with exact successor evidence; pagination, OpenAPI and shared consumers preserved |
| V-012 | Routing; S-011; C-011 | `make generate`; `make source-check CHECK=generated`; `make source-check CHECK=schema`; `make source-check CHECK=catalog`; `make source-check CHECK=identity` | Authored/generated fidelity, exact selector resolution, no aliases/orphans; investigate routing failure before continuation |
| V-013 | Final regression; S-012; C-006–011 | `make agent-finalize`; review changes and freeze source; V-011 portfolio; `make lint CHECK=go`; `make lint CHECK=backend-boundary`; `make harness-observability-check RESULTS_DIR=<each exact passing root> CHECK=all` | Source-compatible final portfolio, cleanup and retained evidence; any repair reopens affected obligations |
| V-014 | Planning-session document check and S-012 final handoff | `make lint-markdown`; direct table/link/ID/dependency/history/diff review | Actual document validity with lint coverage limit; document success does not complete Auth implementation |

Exact current rows were read from the authored Auth manifest. Select all listed
rows for V-007; V-011/V-013 replace only the retired unit row with its real successor
and add the authorization row. Keep all other selectors belonging to a selected
row, including the secret-envelope row's enterprise test. This is a bounded
portfolio, not complete-owner or release qualification.

| Group / purpose | Current row ID |
| --- | --- |
| Admin actions/audit/replay | `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c` |
| Bootstrap route boundaries | `module.auth.integration.bootstrap_tokens_work_only_on_totp_begin_and_tot_36751ad438` |
| Credential/TOTP/password replay | `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb` |
| Deployment-audit route | `module.auth.integration.deployment_audit_route_is_scope_safe_and_keyset_4f8db0c261` |
| Login/session lifecycle | `module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af` |
| Attached socket/concurrency | `module.auth.integration.real_sockets_on_canonical_ws_v1_incidents_incide_8e29eb6263` |
| Secret-envelope integration, enterprise selector retained | `module.auth.integration.secret_envelope_substitution` |
| User lifecycle/audit/create replay | `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6` |
| User-list search/cursor | `module.auth.integration.user_list_search_tokenizes_identifiers_and_binds_45c3e062a8` |
| Predecessor mixed support / retained pagination | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| Administrative-audit OpenAPI | `module.auth.unit.administrative_audit_openapi_is_complete_and_exact_7281a6f4d2` |
| Predecessor unit — retire only at S-010 | `module.auth.unit.an_inventory_driven_loop_covers_every_cookie_aut_0b6e53ff68` |
| Bound-secret primitive | `module.auth.unit.bound_secret_envelopes` |
| Administrative credential guards | `module.auth.unit.deployment_admin_only_credential_routes_fail_clo_5cfa2f8dce` |
| Password-change handler | `module.auth.unit.the_password_change_route_proves_exact_current_p_338ff542a2` |
| Safe credential-state handler | `module.auth.unit.the_real_credential_state_route_returns_the_safe_db26b28985` |
| Session inspection handler | `module.auth.unit.the_real_get_auth_session_route_rejects_paginati_14b12305f4` |
| Login resource handler | `module.auth.unit.the_real_login_handler_creates_exactly_one_sessi_3b913a16a6` |
| Login exactness handler | `module.auth.unit.the_real_login_handler_trims_and_case_folds_the_9ed31374cf` |
| TOTP handlers | `module.auth.unit.the_totp_begin_and_complete_handlers_enforce_exa_59570a1bac` |
| User-create defaults/closed vocabulary | `module.auth.unit.user_create_rejects_undeclared_is_active_rejects_839f6b57c5` |

At execution entry, copy the exact selected IDs into each literal invocation and
record commands/results below the handoff checkpoint. The table is human review
material; no test, generator, runtime or evidence tool may parse it for selection.
New row IDs are recorded only after the public authoring command returns them.

Finalizer retained-run maintenance accepts `RESULTS_DIR` only when an eligible
successful full warm-check root already exists. Otherwise leave it unset and
report maintenance skipped because `RESULTS_DIR` was unset; do not run a redundant
warm check. During this document-only save, finalization is entirely out of scope.
Browser/UI, full-owner/full-project, performance and release gates are excluded:
no corresponding production surface changes or qualification claim are planned.

| Session operation | Invocation-path evidence / side effects | Admission / result |
| --- | --- | --- |
| Static source and preservation inspection | Read-only source/Git inspection and temporary comparison files outside checkout | E-015/016; no test result |
| V-007–013 and row authoring | Would execute tests/services, generate projections or finalize source | Not run in this document step; future authorization and entry gates required |
| V-014 Markdown lint | Public Make target, installed-only Node/frontend readiness, non-fixing lint; step/cache outputs redirected to `/tmp/cartulary-tr02-auth-doc-j5vkqn7x` | Only documentation check admitted; actual command/exit recorded in E-016 |
| Direct tracker review | `.markdownlint-cli2.jsonc` omits this handoff | Required independently; lint pass alone cannot validate it; no configuration change for this save |

### Incident iteration — completed record

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

### Auth top-level tracker

| Item | Workstream | Status | Dependency / evidence | Exit |
| --- | --- | --- | --- | --- |
| T-010: save and review Auth iteration plan | Documentation | DONE | E-015 static planning; E-016 source/save revalidation and validation | Tracker-only save, document validation and preservation comparison complete |
| T-011: current Auth characterization | WF-06 | DONE | E-017; S-006; V-007 | Exact baseline and disposition ledger recorded |
| T-012: direct CSRF unit successor | WF-07 | DONE | E-018; S-007; V-008 | All explicit unit cases qualify before retirement |
| T-013: credential/session consolidation | WF-08 | DONE | E-019; S-008; V-009 | Unique credential, replay, secret and revocation observations qualify |
| T-014: administrative consolidation | WF-09 | DONE | E-020; S-009; V-010 | Current capability and scoped integrity qualify |
| T-015: remove redundant framework | WF-10 | DONE | E-021; S-010; V-011 | D-009–018 closed; framework absent; real helpers preserved |
| T-016: reconcile Auth routing | WF-11 | DONE | E-022; S-011; V-012 | Metadata/projections/identities pass without aliases |
| T-017: final Auth validation and handoff | WF-12 | DONE | E-023; S-012; V-013/014 | Source-bound portfolio, audits, exclusions and all checkpoints recorded last |

Planning/document completion does not complete any Auth implementation workstream.

### Incident iteration — completed record

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

### Auth implementation checkpoint — WF-06 / E-017

The current implementation request supersedes the historical document-only
authorization. HEAD and all product sources still match the planned baseline;
the tracker is the sole pre-existing edit. Root/backend/tooling guidance and
the test-operations skill were reloaded. Prior owner/source inspection was
revalidated against unchanged inputs: no normative amendment or production
correction is indicated. The two inventory consumers remain the support
integration suites and handler unit suite. Existing dedicated secret-binding,
audit, lifecycle and pagination witnesses remain protected.

`make help ALL=1` and `make task-guide OWNER=module.auth DETAIL=coverage JSON=1`
passed. V-007 ran the following literal selection (native session 91839):

```sh
CARTULARY_TEST_RUN_ID=tr02-wf06-auth-20261028-01 make test-slice OWNER=module.auth ROWS=module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c,module.auth.integration.bootstrap_tokens_work_only_on_totp_begin_and_tot_36751ad438,module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb,module.auth.integration.deployment_audit_route_is_scope_safe_and_keyset_4f8db0c261,module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af,module.auth.integration.real_sockets_on_canonical_ws_v1_incidents_incide_8e29eb6263,module.auth.integration.secret_envelope_substitution,module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6,module.auth.integration.user_list_search_tokenizes_identifiers_and_binds_45c3e062a8,module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063,module.auth.unit.administrative_audit_openapi_is_complete_and_exact_7281a6f4d2,module.auth.unit.an_inventory_driven_loop_covers_every_cookie_aut_0b6e53ff68,module.auth.unit.bound_secret_envelopes,module.auth.unit.deployment_admin_only_credential_routes_fail_clo_5cfa2f8dce,module.auth.unit.the_password_change_route_proves_exact_current_p_338ff542a2,module.auth.unit.the_real_credential_state_route_returns_the_safe_db26b28985,module.auth.unit.the_real_get_auth_session_route_rejects_paginati_14b12305f4,module.auth.unit.the_real_login_handler_creates_exactly_one_sessi_3b913a16a6,module.auth.unit.the_real_login_handler_trims_and_case_folds_the_9ed31374cf,module.auth.unit.the_totp_begin_and_complete_handlers_enforce_exa_59570a1bac,module.auth.unit.user_create_rejects_undeclared_is_active_rejects_839f6b57c5
```

Native exit **0**, graph `status=pass`; root:
`/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf06-auth-20261028-01`.
Source digest: `sha256:a3febe51f08f1c1eb1f6332a460c776a6977ed0314d04d7a307e7673f92c79aa`.
All 21 selected rows qualified; this is bounded characterization, not full-owner
or release acceptance. D-009–018 retain their Section 5 destinations; none is
retired or closed merely by this baseline pass. RB-004's current-characterization
gate is satisfied; successor qualification remains mandatory.

The first summary-reader invocation with `JSON=1` exited 2 because that reader
allows JSON only for performance detail. It did not execute or fail a test;
the ordinary summary reader was used instead. No characterization rerun occurred.
`CARTULARY_TEST_RUN_ID=tr02-wf06-document-20261028-01 make lint-markdown`
passed (session 35963, native exit 0); owning summary is
`.cartulary/test-results/tr02-wf06-document-20261028-01/adhoc/lint-markdown/tool-run-summary.json`.
Configured lint excludes this tracker. Direct review of the saved checkpoint,
table structure, links, IDs and unchanged dependency chain completed separately;
`git diff --check` passed. WF-06 is DONE and WF-07 may begin; all later
predecessors remain required.

### Auth implementation checkpoint — WF-07 / E-018

Added `internal/modules/auth/csrf_test.go` with direct handler cases for all nine
cookie mutations, missing and invalid headers, and the three malformed-body
cases. Session lookup may succeed; mutation callbacks fail the test immediately
and hub revocation publication is checked. The predecessor remains intact.
D-017 has a qualified successor but retirement waits for WF-10.

The public row-authoring command in Section 7 returned
`module.auth.unit.cookie_mutations_reject_missing_or_invalid_csrf_dfaa827393`.
Its authored row is fast/unit, service-free, with Auth behavior-contract routing.
`make generate` updated only the expected execution-topology render index.

| Command | Native session / outer exit | Exact result root |
| --- | --- | --- |
| `CARTULARY_TEST_RUN_ID=tr02-wf07-generate-20261028-01 make generate` | 25674 / 0 | `.cartulary/test-results/tr02-wf07-generate-20261028-01` |
| `CARTULARY_TEST_RUN_ID=tr02-wf07-unit-20261028-01 make test-slice OWNER=module.auth ROWS=module.auth.unit.cookie_mutations_reject_missing_or_invalid_csrf_dfaa827393` | 42674 / 0, pass | `.cartulary/test-results/tr02-wf07-unit-20261028-01` |

Source digest for V-008 is
`sha256:e330c7abb9d8a637a1fbe231b1a3c76d97e592e534694795b9f07213ed48034f`;
the summary reader reports all ten cleanup steps complete without failure.
`CARTULARY_TEST_RUN_ID=tr02-wf07-document-20261028-01 make lint-markdown`
passed (session 57081, native exit 0; owning summary under that root's
`adhoc/lint-markdown/tool-run-summary.json`). Direct tracker review and
`git diff --check` passed; the configured lint exclusion still applies.
No production change, normative amendment, failed assertion or compatibility
shim was needed. The saved checkpoint is validated and WF-08 may begin.

### Auth implementation checkpoint — WF-08 / E-019

Strengthened the existing login, bootstrap boundary/consumption and credential
transition tests. Bootstrap rejection now covers twelve Auth HTTP operations,
the two retained incident/test-route cases and the semantic socket witness.
Cookie CSRF guards cover logout, TOTP begin/complete and password change,
including malformed credential bodies. Enrollment replay retains its seed and
one pending record; activation consumes it and completion payloads are safe.
Consumed bootstrap tokens also reject repeat completion. A real attached socket
now witnesses successful self-password revocation. Existing password replay,
replacement-factor, logout/concurrency and secret-binding tests remain intact.
Closed factor vocabulary uses literal wire errors and attributable fields.

`CARTULARY_TEST_RUN_ID=tr02-wf08-format-20261028-01 make format` passed
(session 91686, native exit 0). Its only source formatting was within the
authorized Auth test changes. V-009 then ran (session 51202, native exit 0):

```sh
CARTULARY_TEST_RUN_ID=tr02-wf08-credentials-20261028-01 make test-slice OWNER=module.auth ROWS=module.auth.integration.bootstrap_tokens_work_only_on_totp_begin_and_tot_36751ad438,module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb,module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af,module.auth.integration.real_sockets_on_canonical_ws_v1_incidents_incide_8e29eb6263,module.auth.integration.secret_envelope_substitution,module.auth.unit.bound_secret_envelopes,module.auth.unit.the_password_change_route_proves_exact_current_p_338ff542a2,module.auth.unit.the_real_credential_state_route_returns_the_safe_db26b28985,module.auth.unit.the_real_get_auth_session_route_rejects_paginati_14b12305f4,module.auth.unit.the_real_login_handler_creates_exactly_one_sessi_3b913a16a6,module.auth.unit.the_real_login_handler_trims_and_case_folds_the_9ed31374cf,module.auth.unit.the_totp_begin_and_complete_handlers_enforce_exa_59570a1bac
```

Root: `.cartulary/test-results/tr02-wf08-credentials-20261028-01`; status pass.
Source digest: `sha256:78e4fec91df56a7c777b0755806ed687ca7114a91594ef92be23fa85e8452242`.
D-010 and the credential portions of D-009/011/012/014/016 have qualified
successors. The generic predecessors remain until WF-10. No production defect
or specification change was needed. The summary reader reports all sixteen
cleanup steps complete. `CARTULARY_TEST_RUN_ID=tr02-wf08-document-20261028-01
make lint-markdown` passed (session 30215, native exit 0; owning summary at
`adhoc/lint-markdown/tool-run-summary.json` within that exact run root).
Direct review of the saved checkpoint, tables, links, IDs and dependency chain
and `git diff --check` passed. WF-09 may begin.

### Auth implementation checkpoint — WF-09 / E-020

Strengthened user lifecycle, credential-action, audit and replay tests with
literal normalized values, returned versions, persisted writes, cookie CSRF
guards, stale/last-admin integrity and complete before/after secret checks.
Audit request IDs match their actual responses and credential audit events
carry exact transaction IDs. Padded and normalized reasons replay the same
committed result. The eight direct current-capability scenarios use real
demotion and the same subject session; fresh credential targets retain active
factors, sessions and pending replacements. Scoped observations cover users,
revocations, enrollments, bootstrap tokens, receipts, raw audits and projections,
excluding incidental session activity timestamps.

The row authoring command in Section 7 allocated
`module.auth.integration.administrative_routes_require_current_deployment_71dd666a37`.
It uses Auth behavior routing, standard placement, dedicated exclusive PostgreSQL
and the existing object-store dependency. Shared audit/enterprise helpers remain.

| Command | Native session / exit | Exact root suffix under `.cartulary/test-results/` |
| --- | --- | --- |
| `make generate` | 20387 / 0 | `tr02-wf09-generate-20261028-01` |
| `make format` | 48039 / 0 | `tr02-wf09-format-20261028-01` |
| Initial V-010 portfolio | 38896 / 0, pass | `tr02-wf09-admin-20261028-01` |
| `make format` after final ledger additions | 79114 / 0 | `tr02-wf09-format-20261028-02` |
| Final V-010 portfolio below | 71668 / 0, pass | `tr02-wf09-admin-20261028-02` |

Each command used the root suffix above as its explicit
`CARTULARY_TEST_RUN_ID`. Initial V-010 selected the final command below except
the credential-transition row. It passed at source
`sha256:051f5d4404372d1571762de1b41841971b6c01fe214ecaea4c8a8c4d292df870`.
Ledger review then added scoped denied-read observations, the predecessor's
password-reset HTTP session rejection and self-password active-session count.
Those substantive additions justified the final expanded run:

```sh
CARTULARY_TEST_RUN_ID=tr02-wf09-admin-20261028-02 make test-slice OWNER=module.auth ROWS=module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c,module.auth.integration.deployment_audit_route_is_scope_safe_and_keyset_4f8db0c261,module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6,module.auth.integration.user_list_search_tokenizes_identifiers_and_binds_45c3e062a8,module.auth.unit.administrative_audit_openapi_is_complete_and_exact_7281a6f4d2,module.auth.unit.deployment_admin_only_credential_routes_fail_clo_5cfa2f8dce,module.auth.unit.user_create_rejects_undeclared_is_active_rejects_839f6b57c5,module.auth.integration.administrative_routes_require_current_deployment_71dd666a37,module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb
```

Final source digest:
`sha256:5bc303f139f19dc36a2f2c8fceda387c04826f9087234183588c5b8c369b9fbb`.
Both summaries report fifteen cleanup steps without failure. D-009–017 now
have qualified successors across E-018–020; no generic predecessor is removed
yet. No failing assertion, production correction or specification amendment
was needed. `CARTULARY_TEST_RUN_ID=tr02-wf09-document-20261028-01 make
lint-markdown` passed (session 93007, native exit 0); owning summary is that
root's `adhoc/lint-markdown/tool-run-summary.json`. The tracker remains outside
the configured lint globs. Direct review of the saved text, tables, links,
IDs and dependency chain and `git diff --check` passed. WF-10 may begin.

### Auth implementation checkpoint — WF-10 / E-021

Removed the eight generic integration suites, their universal context and
request switches, secret pool, path/category interpreter and global denial hash.
Removed the old CSRF unit loop and its payload/handler switches, then deleted
`internal/modules/auth/testsupport/routetest/routes.go`. Caller search found
no remaining framework references in executable source or selectors. Both
user-list tests and their four live helpers remain; audit/enterprise consumers
of `lookupUserAuditEvents` remain unchanged.

The pagination row retains its immutable identity and now selects only
`TestUserListContinuationUsesLiveRows`. The retired unit row maps to the
actual CSRF successor in `tools/test_catalog_row_migrations.json`; no executable
alias or compatibility facade was added. Verification-reference reconciliation
remains WF-11's distinct responsibility.

| Command | Native session / exit | Exact root suffix under `.cartulary/test-results/` |
| --- | --- | --- |
| `make format` | 56735 / 0 | `tr02-wf10-format-20261028-01` |
| `make generate` | 66317 / 0 | `tr02-wf10-generate-20261028-01` |
| V-011 portfolio below | 36039 / 0, pass | `tr02-wf10-portfolio-20261028-01` |

Each suffix is the explicit `CARTULARY_TEST_RUN_ID` for its invocation.

```sh
CARTULARY_TEST_RUN_ID=tr02-wf10-portfolio-20261028-01 make test-slice OWNER=module.auth ROWS=module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c,module.auth.integration.bootstrap_tokens_work_only_on_totp_begin_and_tot_36751ad438,module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb,module.auth.integration.deployment_audit_route_is_scope_safe_and_keyset_4f8db0c261,module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af,module.auth.integration.real_sockets_on_canonical_ws_v1_incidents_incide_8e29eb6263,module.auth.integration.secret_envelope_substitution,module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6,module.auth.integration.user_list_search_tokenizes_identifiers_and_binds_45c3e062a8,module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063,module.auth.unit.administrative_audit_openapi_is_complete_and_exact_7281a6f4d2,module.auth.unit.bound_secret_envelopes,module.auth.unit.deployment_admin_only_credential_routes_fail_clo_5cfa2f8dce,module.auth.unit.the_password_change_route_proves_exact_current_p_338ff542a2,module.auth.unit.the_real_credential_state_route_returns_the_safe_db26b28985,module.auth.unit.the_real_get_auth_session_route_rejects_paginati_14b12305f4,module.auth.unit.the_real_login_handler_creates_exactly_one_sessi_3b913a16a6,module.auth.unit.the_real_login_handler_trims_and_case_folds_the_9ed31374cf,module.auth.unit.the_totp_begin_and_complete_handlers_enforce_exa_59570a1bac,module.auth.unit.user_create_rejects_undeclared_is_active_rejects_839f6b57c5,module.auth.unit.cookie_mutations_reject_missing_or_invalid_csrf_dfaa827393,module.auth.integration.administrative_routes_require_current_deployment_71dd666a37
```

All 22 selected rows pass at source
`sha256:e823512d55b4537930eec2d75d205268d70102856605a54fe3f6649a4edc68f6`.
The summary reports all sixteen cleanup steps complete. D-009–018 close here;
all product successors first qualified before retirement at E-018–020 and
qualify together at this exact passing root. The following map records exact
row ownership; all selectors within those rows were retained and executed.

| Disposition | Qualified successor / retirement | Exact row ownership |
| --- | --- | --- |
| D-009 | `TestLoginSessionLifecycle_Integration`, credential lifecycle, `TestUserAdminLifecycle_Integration`, `TestAdminCredentialActions_Integration`, `TestDeploymentAdministrativeAuditRoute_Integration` preserve all fifteen success statuses and envelopes/request-ID relationships | `module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af`; `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb`; `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6`; `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c`; `module.auth.integration.deployment_audit_route_is_scope_safe_and_keyset_4f8db0c261` |
| D-010 | `TestBootstrapTokenRouteBoundaries_Integration`: twelve Auth HTTP operations, socket, incident and privileged test-route boundaries | `module.auth.integration.bootstrap_tokens_work_only_on_totp_begin_and_tot_36751ad438` |
| D-011 | Real-route CSRF failures before valid logout, credential, user and admin actions, including malformed credential bodies | `module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af`; `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb`; `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6`; `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c` |
| D-012 | Existing committed-resource replay tests and TOTP lifecycle; cryptographic storage/binding expectations owned by dedicated secret tests | `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6`; `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb`; `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c`; `module.auth.integration.secret_envelope_substitution`; `module.auth.unit.bound_secret_envelopes` |
| D-013 | `TestUserAdminAudit_Integration` and `TestAdminCredentialAuditAndScope_Integration`: attribution, actual transaction/request IDs, timestamps and full forbidden-key checks | `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6`; `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c` |
| D-014 | Logout/concurrency, self-password and administrative attached-socket revocation; reset target HTTP denial and zero active self-password sessions; redundant connect/hello-only branch retired | `module.auth.integration.real_sockets_on_canonical_ws_v1_incidents_incide_8e29eb6263`; `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb`; `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c` |
| D-015 | `TestDeploymentAdminAuthorizationRefresh_Integration`: eight concrete success/demotion/denial scenarios, scoped state including audits | `module.auth.integration.administrative_routes_require_current_deployment_71dd666a37` |
| D-016 | Literal closed vocabulary, normalized user strings and normalized-reason replay; production normalizer removed from expectations | `module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af`; `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb`; `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6`; `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c` |
| D-017 | `TestCookieMutationCSRF_Unit`: nine missing/invalid pairs and three malformed cases; no-mutation/revocation retained; incidental no-session-read constraint retired | `module.auth.unit.cookie_mutations_reject_missing_or_invalid_csrf_dfaa827393` |
| D-018 | Framework-only metadata/builders/hash removed without artificial replacement; live pagination and audit/enterprise helpers preserved | `module.auth.integration.user_list_search_tokenizes_identifiers_and_binds_45c3e062a8`; `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063`; `module.auth.integration.deployment_audit_route_is_scope_safe_and_keyset_4f8db0c261`; `module.auth.integration.secret_envelope_substitution` |

No owner amendment, production defect, public API migration or harness lifecycle
change was required. `CARTULARY_TEST_RUN_ID=tr02-wf10-document-20261028-01
make lint-markdown` passed (session 21981, native exit 0; owning summary under
that exact root at `adhoc/lint-markdown/tool-run-summary.json`). Direct saved
tracker review and `git diff --check` passed; configured lint excludes this
tracker. The checkpoint is validated and WF-11 may begin.

### Auth implementation checkpoint — WF-11 / E-022

Reconciled routing without changing test behavior. The surviving pagination row
keeps its identity, full aggregate placement, fixture, isolation and services;
its sole verification reference is now Auth's behavior contract. Dedicated
deployment-audit, audit OpenAPI and Auth OpenAPI rows retain the audit/OpenAPI
reference. Both new rows select their real test functions. The one unit-row
retirement has an actual successor; surviving rows have no migrations or aliases.
Generated changes are limited to the execution-topology input digest and Auth
manifest hash, produced through Make.

| Exact command | Native session / exit | Result |
| --- | --- | --- |
| `CARTULARY_TEST_RUN_ID=tr02-wf11-generate-20261028-01 make generate` | 81983 / 0 | pass |
| `CARTULARY_TEST_RUN_ID=tr02-wf11-generated-20261028-01 make source-check CHECK=generated` | 18825 / 0 | pass |
| `CARTULARY_TEST_RUN_ID=tr02-wf11-schema-20261028-01 make source-check CHECK=schema` | 64787 / 0 | pass |
| `CARTULARY_TEST_RUN_ID=tr02-wf11-catalog-20261028-01 make source-check CHECK=catalog` | 73218 / 0 | pass |
| `CARTULARY_TEST_RUN_ID=tr02-wf11-identity-20261028-01 make source-check CHECK=identity` | 75483 / 0 | pass |

Each exact root is `.cartulary/test-results/<CARTULARY_TEST_RUN_ID>`.
Generation owns `generate/tool-run-summary.json`; source checks own
`run-summary.json`. All four source checks share source digest
`sha256:918d8a8530983a0d954c3f8e06e4569204c9e768cb8e3cf6520a8e610aa50086`.
Their summary readers report complete cleanup (14/12/11/12 steps respectively),
without failures or blocked steps. No routing failure or repair occurred.
V-012 is complete; final-source regression and evidence audits remain WF-12.
`CARTULARY_TEST_RUN_ID=tr02-wf11-document-20261028-01 make lint-markdown`
passed (session 32431, native exit 0; owning summary under that exact root at
`adhoc/lint-markdown/tool-run-summary.json`). Direct saved tracker review and
`git diff --check` passed; the configured lint exclusion remains explicit.
WF-11 is complete and WF-12 may begin.

### Auth final validation and handoff — WF-12 / E-023

`CARTULARY_TEST_RUN_ID=tr02-wf12-finalize-20261028-01 make agent-finalize`
passed (native session 14892, outer exit 0). Its changes were reviewed before
the broader final portfolio. No further source change was required.
`RESULTS_DIR` was unset: retained-run maintenance was skipped because no
compatible successful full warm-check root was available. No redundant full
check was run merely to obtain one.

The frozen executable-source digest for finalization, final tests, both lints
and all four WF-11 source checks is
`sha256:918d8a8530983a0d954c3f8e06e4569204c9e768cb8e3cf6520a8e610aa50086`.
Final tracker edits are human documentation and are excluded from runtime
source identity. Earlier-source characterization and slice results remain
historical qualification; the following final-source portfolio closes C-006–011.

```sh
CARTULARY_TEST_RUN_ID=tr02-wf12-auth-20261028-01 make test-slice OWNER=module.auth ROWS=module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c,module.auth.integration.bootstrap_tokens_work_only_on_totp_begin_and_tot_36751ad438,module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb,module.auth.integration.deployment_audit_route_is_scope_safe_and_keyset_4f8db0c261,module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af,module.auth.integration.real_sockets_on_canonical_ws_v1_incidents_incide_8e29eb6263,module.auth.integration.secret_envelope_substitution,module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6,module.auth.integration.user_list_search_tokenizes_identifiers_and_binds_45c3e062a8,module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063,module.auth.unit.administrative_audit_openapi_is_complete_and_exact_7281a6f4d2,module.auth.unit.bound_secret_envelopes,module.auth.unit.deployment_admin_only_credential_routes_fail_clo_5cfa2f8dce,module.auth.unit.the_password_change_route_proves_exact_current_p_338ff542a2,module.auth.unit.the_real_credential_state_route_returns_the_safe_db26b28985,module.auth.unit.the_real_get_auth_session_route_rejects_paginati_14b12305f4,module.auth.unit.the_real_login_handler_creates_exactly_one_sessi_3b913a16a6,module.auth.unit.the_real_login_handler_trims_and_case_folds_the_9ed31374cf,module.auth.unit.the_totp_begin_and_complete_handlers_enforce_exa_59570a1bac,module.auth.unit.user_create_rejects_undeclared_is_active_rejects_839f6b57c5,module.auth.unit.cookie_mutations_reject_missing_or_invalid_csrf_dfaa827393,module.auth.integration.administrative_routes_require_current_deployment_71dd666a37
```

| Exact invocation / obligation | Native session / exit | Owning summary and result |
| --- | --- | --- |
| Final Auth selection above: 22 rows, including every surviving predecessor row and both real successors | 31855 / 0 | `.cartulary/test-results/tr02-wf12-auth-20261028-01/run-summary.json`; pass, 5/5 units; no failed/skipped/cancelled units |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-lint-go-20261028-01 make lint CHECK=go` | 6941 / 0 | `.cartulary/test-results/tr02-wf12-lint-go-20261028-01/run-summary.json`; pass |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-lint-backend-boundary-20261028-01 make lint CHECK=backend-boundary` | 82021 / 0 | `.cartulary/test-results/tr02-wf12-lint-backend-boundary-20261028-01/run-summary.json`; pass |
| WF-11 generated/schema/catalog/identity source checks | E-022 exact commands and native exits | Same source digest; retained and audited without redundant reruns |
| Required finalizer | 14892 / 0 | `.cartulary/test-results/tr02-wf12-finalize-20261028-01/run-summary.json`; pass |

The final Auth summary records zero cache hits (two misses, three bypasses).
Cleanup summaries report all steps complete without failure or blockage:
finalizer 10, Auth 16, Go lint 15, backend-boundary lint 13; WF-11 counts are
in E-022. The explicit selection is bounded Auth assurance, not complete-owner,
UI, performance or release acceptance.

All eight exact passing roots were admitted by the public read-only audit:

| Exact audit command | Observed result |
| --- | --- |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-audit-01-20261028-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf12-finalize-20261028-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-audit-02-20261028-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf12-auth-20261028-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-audit-03-20261028-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf12-lint-go-20261028-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-audit-04-20261028-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf12-lint-backend-boundary-20261028-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-audit-05-20261028-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf11-generated-20261028-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-audit-06-20261028-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf11-schema-20261028-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-audit-07-20261028-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf11-catalog-20261028-01 CHECK=all` | Native exit 0; `full_validity=1` |
| `CARTULARY_TEST_RUN_ID=tr02-wf12-audit-08-20261028-01 make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr02-wf11-identity-20261028-01 CHECK=all` | Native exit 0; `full_validity=1` |

These audits validate compatible retained evidence; they do not execute missing
tests. Optional resource/diagnostic instrumentation was not enabled, so no
resource-accounting or performance conclusion is claimed. No final validation
failure, unresolved cleanup issue or production correction remains.

| Finding | Final remediation and evidence |
| --- | --- |
| F-009 | Private route model and both interpreters removed; direct successors E-018/020, retirement E-021, final portfolio above |
| F-010 | Universal fixture, identity sequence, secret pool and operation-name repairs removed; concrete scenario arrangements remain |
| F-011 | All distinct assertions consolidated before deletion; D-009–018 exact test/row ledger at E-021 and final-source qualification above |
| F-012 | Literal expectations, permitted session reads and owner-local cryptographic witnesses; incidental internal-order/framing duplication retired |
| F-013 | Scoped user/credential/session/receipt/audit integrity and consumed response bodies; material writes and actual attribution qualify |
| F-014 | Two authored successors, one real unit-row migration, retained pagination identity and truthful references; E-022 source checks and final audits |

The specification reconciliation found no adopted-owner conflict or product
defect. Tests, private support, routing projections and this handoff required
changes; public APIs, database schema, credential policy and runtime behavior
need no migration. The private inventory has no wrapper or alias. New behavior
coverage belongs with the owning lifecycle or security scenario, using literal
expectations and existing fixture/transport ownership.

A final human preservation comparison is retained at
`/tmp/cartulary-tr02-auth-impl-yeuvuwz3/implementation-comparison.json`.
It identifies exactly the planned ten changed/added/deleted repository paths,
including this pre-existing tracker edit. HEAD and the staged tree are unchanged;
the index file's bytes changed, so byte-for-byte index preservation is not claimed.
Seventy regenerated files have new file identities but identical content.
Other retained tracked/non-ignored content is unchanged. Ignored build, cache,
service and result outputs remain outside that comparison. No staging, commit
or push occurred.

#### Final saved checkpoint — E-023

`CARTULARY_TEST_RUN_ID=tr02-wf12-document-20261028-01 make lint-markdown`
passed (native session 27032, outer exit 0), with owning summary
`.cartulary/test-results/tr02-wf12-document-20261028-01/adhoc/lint-markdown/tool-run-summary.json`.
Configured lint excludes this tracker. Separate direct review verified all
12 numbered sections, stable IDs, table shapes, local links, code fences,
sequential checkpoints, actual row/disposition mappings, historical separation
and final outcomes; `git diff --check` passed. Human-only document review
material is retained at
`/tmp/cartulary-tr02-auth-impl-yeuvuwz3/document-review.json`.

WF-12/S-012 is DONE and finishes the Auth effort last. F-009–014 and D-009–018
are closed, no dependent implementation work remains, and no unresolved
owner decision or validation failure is handed forward. Future changes should
extend owning behavioral tests and preserve current authorization, meaningful
state integrity, replay and independent cryptographic witnesses. This handoff
is review material, never a source for executable evidence. Human review of
projection fidelity remains distinct from test success.

### Auth planning and document-save session — E-015/E-016


All dates in this entry are 2026-10-27, America/New_York. Current source is clean
`main` at `53bdd5341304edd005c451b05178861eb02c51fc` before the tracker-only edit.
Earlier source identity, authorization and passing results below remain historical.

#### Auth scope and authority

| Evidence | Current state / inspected source | Supersession / next action |
| --- | --- | --- |
| E-015 | User selected Auth route tests during static planning; inspected inventory, consumers, existing successors and adopted owners; no tests ran | Supersedes Auth exclusion for the next iteration only; current task saves documentation |
| E-016 | Same HEAD and clean source at save entry; root/backend/tooling instructions, refactor skill/format/framework and tracker identity revalidated | Preserve existing incident records and all IDs; no implementation authority carried forward |

#### Auth backend boundary

| Evidence | Current state / files | Next action |
| --- | --- | --- |
| E-015/016 | Three-file movement universe and supporting sources in Section 2; no production layer move or public interface change | Future S-007–010 consolidate into owning tests using existing support; protect audit/enterprise consumers |

#### Auth frontend boundary

Not applicable: no browser, UI, frontend source or golden change is included.

#### Auth contracts and codegen

| Evidence | Current state / files | Next action |
| --- | --- | --- |
| E-015/016 | Current Auth authored row selectors and owner references inspected; no normative amendment or source contradiction found | Author two semantic rows and regenerate during future slices; keep projections independent of Markdown |

#### Auth tests and harness

| Evidence | Current state / commands | Next action |
| --- | --- | --- |
| E-015 | Direct assertion and catalog inspection only; zero tests/builds/generators/services/finalizer | Future V-007 supplies current characterization; old incident results do not qualify Auth |
| E-016 | Document-only save; installed-only Markdown lint exit 0, direct review and preservation comparison passed below | T-010 DONE; WF-06–12 remain TODO |

#### Auth security and authorization

| Evidence | Decision | Required future witness |
| --- | --- | --- |
| E-015/016 | Preserve current deployment capability, real cookie CSRF, bootstrap restrictions, credential replay, safe audit/storage and real socket revocation | Literal outcomes, same subject session after real demotion, valid targets, scoped effects and qualified bound-secret tests |
| E-015 | No-session-lookup constraint and metadata taxonomy are incidental; byte-format knowledge belongs to cryptographic owners | Retain rejection/no harmful effects and owner-local security assurance; do not weaken requirements |

#### Auth open risks and next session

| Evidence | Risk / blocker | Next action |
| --- | --- | --- |
| E-015/016 | RB-004: current Auth characterization missing; no successor execution; implementation not authorized by document task | Revalidate changed source on a later implementation task, enter WF-06, then save each workstream checkpoint before its successor |

#### Auth evidence register

| Evidence ID | Source / command | Baseline / observation | Applicability |
| --- | --- | --- | --- |
| E-015 | Prior-turn bounded `rg`, direct source/owner reads, authored JSON inspection and `git --no-optional-locks` identity/status reads | HEAD `53bdd5341304edd005c451b05178861eb02c51fc`; clean source; user selected Auth; complete static plan; no tests executed | Static evidence only; no current runtime pass |
| E-016 | Current save baseline, tracker identity recheck, source equality, public documentation check and direct document review | `/tmp/cartulary-tr02-auth-doc-j5vkqn7x/baseline.json`; actual command and comparison results recorded below | Sole-file documentation completion only; future work stays TODO |

#### Auth document validation and preservation — E-016

The pre-write check confirmed the exact baseline tracker SHA-256, inode, regular
single-link type, canonical target and unsymlinked parent chain before editing.
Prior planning source remains applicable because HEAD and clean working files
match its baseline. Relevant command paths and exact current authored rows were
re-read. No unseen concurrent tracker changes were overwritten.

Executed from `/home/jochi/code/cartulary` with installed-only prerequisites and
external artifact/cache roots:

```sh
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-tr02-auth-doc-j5vkqn7x/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-tr02-auth-doc-j5vkqn7x/static-analysis \
CARTULARY_TEST_RUN_ID=tr02-auth-document-help-20261027-01 \
COREPACK_ENABLE_NETWORK=0 make help ALL=1 > /tmp/cartulary-tr02-auth-doc-j5vkqn7x/help.txt

CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-tr02-auth-doc-j5vkqn7x/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-tr02-auth-doc-j5vkqn7x/static-analysis \
CARTULARY_TEST_RUN_ID=tr02-auth-document-lint-20261027-01 \
COREPACK_ENABLE_NETWORK=0 make lint-markdown
```

Command discovery returned native exit **0** without yielding a session. Markdown
lint used native session **6112** and returned outer/native exit **0**; the owning
summary reports `status=pass`, `exit_code=0`, no failure class/reason. Exact root:
`/tmp/cartulary-tr02-auth-doc-j5vkqn7x/results/tr02-auth-document-lint-20261027-01`.
Summary: `adhoc/lint-markdown/tool-run-summary.json` under that root. Installed-only
readiness inspected existing dependencies; no installation, build, generator,
service operation, formatter, product test or finalizer ran.

The lint configuration omits this handoff, so its pass qualifies configured
documents only. Supplemental direct review checked all twelve numbered sections,
58 tables, six local links, stable IDs, the sequential dependency chain, current
row references and complete preservation of each original section body. No issues
were found; `/tmp/cartulary-tr02-auth-doc-j5vkqn7x/document-review.json` records the
review. `git --no-optional-locks diff --check` also passed. This review is human
document maintenance, not a product test, generator or conformance mechanism.

The final content/file-identity comparison is retained in
`/tmp/cartulary-tr02-auth-doc-j5vkqn7x/comparison.json`, with the final snapshot
beside it. It verifies unchanged HEAD/branch/index, no added or removed paths,
7,663 unchanged tracked/non-ignored paths and exactly one modified path: this
tracker. Its inode and regular single-link identity remain unchanged. Ignored
trees were not exhaustively audited. Historical section contents remain intact;
new current-state material explicitly supersedes only the Auth exclusion and
invocation posture. No commit, staging or push was performed.

T-010 is DONE. All Auth implementation workstreams/slices remain TODO; RB-004
remains blocked on future characterization. V-007–013, generation, formatting,
product tests and finalization were skipped because this is the document-only
step. Finalizer retained-run maintenance is not applicable to this save; a future
S-012 must report it skipped if `RESULTS_DIR` is unset. Next action: a later
implementation task revalidates source and enters WF-06, then honors every saved
checkpoint gate through WF-12.

### Incident iteration — preserved execution and planning history

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

### Auth blockers and readiness

No unresolved design choice or primary-owner contradiction remains for the
bounded Auth plan. The following are execution gates, not missing plan decisions.

| ID | Gate / affected work | Status | Resolution requirement |
| --- | --- | --- | --- |
| RB-003 | Earlier Plan mode prevented saving the Auth plan | DONE | E-016 saved the plan; the later implementation request authorizes WF-06–12 |
| RB-004 | Current Auth characterization and successor qualification | DONE | E-017 baseline; E-018–020 successors before E-021 retirement; E-023 final-source portfolio |

No unresolved owner decision, production defect, test failure or cleanup issue
remains. WF-06–12 are complete at E-017–023; WF-12 finished last. Human
projection-fidelity review remains a separate review activity, and the scope
exclusions in Section 1 still apply.

### Incident iteration — completed record

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

### Current Auth iteration outcomes

| Outcome | Current statement |
| --- | --- |
| Planning completeness | complete for the bounded Auth framework/consumer seam; not a whole-suite audit |
| Per-slice readiness | S-006–012 complete; every predecessor checkpoint saved and validated; S-012 finished last |
| Implementation authorization | Granted by the subsequent user request to implement the plan; historical document-only limits are superseded |
| Document status | Complete; E-017–023 record every implementation checkpoint; E-023 final Markdown lint and direct tracker review pass |
| Session write-boundary verification | E-023 comparison records the ten authorized changed paths, unchanged HEAD/staged tree and content-identical regenerated files; ignored outputs excluded |

Historical Auth document-save completion — E-016:

- [x] Same canonical target and output admitted; current HEAD and content/file identities recorded.
- [x] Bounded inventory and inspection limits distinguish direct source reading from enumeration.
- [x] Adopted production requirements are separate from incidental test constraints and missing execution evidence.
- [x] F-009–014, C-006–011, D-009–018, S-006–012 and V-007–014 are connected.
- [x] Every sequential workstream has prerequisites, risk controls, validation, rollback and binary exit.
- [x] Exact current rows, real future migration rules and two public row-authoring commands are documented.
- [x] Existing twelve sections, IDs, historical incident results and prior decisions are preserved with explicit supersession.
- [x] Public documentation check and direct review completed with actual coverage limits.
- [x] Final preservation comparison verifies no persistent repository changes outside this tracker within the recorded comparison scope.

Auth implementation completion — E-017–023:

- [x] WF-06 current characterization qualifies and each successor begins only after a saved checkpoint.
- [x] Explicit CSRF unit and direct current-capability integration successors qualify.
- [x] Credential/bootstrap/replay/audit/session guarantees qualify in their owning tests.
- [x] Every denied administrative mutation preserves scoped state and attributable effects; material successful writes persist.
- [x] D-009–018 close with qualified successors or explicit incidental-retirement rationale.
- [x] Eight generic suites, old unit loop, Auth inventory and framework-only helpers are removed without wrappers.
- [x] Pagination, shared audit/enterprise helpers and independent security/lifecycle witnesses remain intact.
- [x] Authored/generated routing and actual migration crosswalks pass without aliases or orphaned references.
- [x] No production correction is necessary; none was demonstrated against the adopted owners.
- [x] Finalizer changes are reviewed; final source portfolio, lint and retained-root audits pass.
- [x] Actual gap/slice/evidence and row maps, failures, compatibility, exclusions and maintenance skips are recorded.
- [x] WF-12 completes the final handoff last; no broader owner/performance/UI/release qualification is implied.

The following incident outcomes remain completed historical records. Their older
phrases about the current implementation request and unchanged HEAD describe
that invocation, not the present Auth implementation.

### Incident iteration — completed record

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
