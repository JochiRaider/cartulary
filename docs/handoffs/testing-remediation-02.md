# Testing Remediation 02: HTTP Test Support Simplification

## 1. Scope and Source Posture

### HTTP support implementation — E-024 onward

The 2026-10-28 request to implement the HTTP Test Support Simplification plan
authorizes WF-13–19 / S-013–019 and their required verification, generation,
finalization and tracker checkpoints. This supersedes the earlier document-only
restriction for this effort. No staging, commit or push is requested.

The repository target remains `/home/jochi/code/cartulary`, relative identity `.`,
a directory. The controlling artifact is this regular single-link file; its
initial SHA-256 is `d7a08b8ed4ecc8d218c1dcc9b9c6cfb8e2b2a69c183810ee5904ca46b46ca876`.
Artifact mode: **normal**. The earlier planning session made no changes or test
runs. Implementation starts from clean HEAD
`157b9673b65101a80bada3ec749fe9c518279350`.

Consolidate generic HTTP test requests and response ownership in
`internal/testutil/httptestx`. Preserve domain fixtures, cookie policy, process
configuration, application startup, clocks and service lifecycles. Workbook
conformance, UI, performance and release qualification remain outside scope.
Core 01 REQ-01-228–232 already owns the required envelope behavior; no normative
amendment is needed to preserve incidental helper assumptions. Human review of
projection fidelity remains separate from executable evidence. Executable tests,
generators, runtime metadata and release evidence must never consume this tracker
or other Markdown.

Content and identity evidence for 7,665 tracked/non-ignored files, HEAD, index and
the original tracker is retained in `/tmp/cartulary-tr02-http-ojbvn4pn/baseline.json`
and `tracker-before.md`. Ignored caches/build/service results are excluded.
Existing Incident and Auth completion records remain historical and unchanged.
The twelve numbered sections and existing IDs are retained.

**Current outcome:** WF-13–19 / S-013–019 are complete at E-025–031;
592 final rows, both lints, four source checks, 38 audits and document validation
pass. WF-19 finished last. Retained finalizer maintenance was skipped because
RESULTS_DIR was unset. See E-031 for scope and preservation limits.

**Initial posture (historical):** planning complete; implementation authorized; WF-13–19 TODO.
S-013 may enter characterization; later slices depend on the qualified predecessor
and a saved, reviewed checkpoint. WF-19 must finish last. Current execution and
any superseding status are recorded in Sections 9–12.


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

### HTTP movement and preservation inventory — E-024

The generic helper bodies and their local callers were directly inspected during
planning and revalidated at the unchanged HEAD. A conservative static import
closure identifies 40 packages and 590 Go catalog rows; this is dependency
inventory, not whole-repository behavioral inspection. The complete package/row
inventory is retained in `/tmp/cartulary-tr02-http-ojbvn4pn/portfolio.json` and the
exact row portfolio is recorded in Section 8. Other files were enumerated for
preservation, not read as production-assurance evidence.

| Surface / accounting | Responsibility and intended disposition | Inspection limits / risk |
| --- | --- | --- |
| `internal/testutil/httptestx/httptestx.go`, `httptestx_test.go`, `routes.go` — directly inspected | Canonical HTTP primitives; retain server/clock capabilities and real assembly witnesses; replace response ownership and envelope assumptions | No lifecycle or cookie-policy redesign |
| `internal/testutil/appsupport/scenario_fixtures.go`; Auth `testsupport/flowtest/flow.go` — generic bodies directly inspected | Remove JSON/options/envelope forwarding; preserve semantic auth and domain setup | All generic callers require mechanical migration |
| Auth integration and enterprise tests; Timeline integration; Entities resolution support; Evidence routes; Reference Data reference-pack tests — helper bodies directly inspected | Delete duplicated JSON/raw/options/envelope helpers | Retain forms, redirect clients, route-specific assertions and storage effects |
| `internal/app/serverprocess` HTTP support — helper bodies directly inspected | Keep origin plus trusted-client adapter; remove generic options forwarding | Preserve TLS, timeouts and process launch behavior |
| Status/raw-body consumers in Evidence, Incidents, Network Flow, Saved Views and serverprocess — relevant blocks directly inspected | Choose a single terminal body owner, preserving bytes/headers/empty-body guarantees | Avoid double reads and closing before an existing body assertion |
| Remaining transitive caller packages — changed call patterns mechanically inspected; unrelated bodies metadata-only | Mechanical API update and selected regression only | No claim that all test bodies were audited |
| Authored family manifests and verification owners — relevant metadata inspected | Add exact helper tests; retain existing claims/IDs | Generated projections refreshed only through Make |
| Other tracked/non-ignored files — enumerated, excluded from movement | Preserve outside authorized scope | Ignored output trees excluded from content comparison |


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

### HTTP responsibility decisions

| Responsibility | Owner | Decision | Finding / contract |
| --- | --- | --- | --- |
| Request encoding, options, dispatch | Existing `httptestx` | One explicit-client API; no compatibility aliases | F-015; C-012 |
| Response read/close/status and JSON parsing | Existing `httptestx` | Terminal readers; private error-returning functions for failure tests | F-016; C-013 |
| Generic envelope requirements | Core 01, asserted in `httptestx` | Correct mandatory fields; tolerate additive members and optional details | F-017; C-014 |
| Route policy, fixtures and effects | Existing feature/process test owners | Preserve semantic helpers and assertions | C-015 |
| Exact selection and evidence | Adopted Testing Harness NLSpec and authored routing | Add unit witnesses without replacing assembly assurance | F-018; C-016 |


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

### HTTP contract map

| Contract | Required behavior / owner | Disposition and validation |
| --- | --- | --- |
| C-012 | Requests preserve nil/JSON/raw bodies, ordered options and caller-selected TLS/redirect/timeout policy | Preserve; S-015–017; V-017–019 |
| C-013 | Each response has one terminal body owner; raw/binary/204 responses are not coerced through JSON | Correct helper lifecycle; S-014; V-016 |
| C-014 | Core 01 REQ-01-228–232: data presence, correlated meta; error code/message/status/request_id/retryable; optional details and additive members | Remove incidental object-data/mandatory-details assumptions; preserve route-required details; S-014; V-016 |
| C-015 | Adopted route/security/storage owners continue governing literal outcomes, authorization ordering, cookies, replay, persistence and process behavior | Preserve inspected assertions; no product policy change; S-013–019; V-015–021 |
| C-016 | Adopted harness requires unique exact selectors, authored ownership and compatible retained evidence | Preserve current claim identities; S-014–019; V-016–021 |

Tests are downstream witnesses, not specification authority. A production defect
exposed by corrected assertions requires a distinct owner-backed correction
workstream before dependent movement; never weaken assertions to hide it.


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

### HTTP findings and disposition ledger

| Finding | Remediation / areas | Rationale and benefit | Migration impact / unresolved risk | Completion |
| --- | --- | --- | --- | --- |
| F-015 — must_fix: duplicated request builders/options/aliases | One `httptestx` API; implementation/tests/docs | Prevent drift and reveal transport choices | Internal callers migrate atomically; no aliases. Leaving copies risks policy loss and divergent fixes | S-015–017; V-017–019; no generic duplicates remain |
| F-016 — must_fix: ambiguous response ownership and fatal-path rewind | Terminal body readers; implementation/tests/docs | Reliable cleanup and one understandable lifecycle | Replace status-only helper and redundant closes; otherwise leaks/double reads remain | S-014; V-016; closure on all paths, raw/body semantics preserved |
| F-017 — must_fix: missing required error assertions and unsupported generic constraints | Correct shared assertions; tests/docs; no normative amendment | Enforce owner requirements without blocking valid evolution | Newly detected defects get separate correction; otherwise false positives and false negatives persist | S-014; V-016; required/optional/additive cases qualified |
| F-018 — must_fix: missing isolated helper failure witnesses | Service-free tests and authored routing; tests/docs | Fast diagnosis plus retained assembly assurance | No replacement runner or fixture framework; otherwise expensive suites hide helper defects | S-014–019; V-016–021; exact selectors and preserved assembly witness |

| Disposition | Decision | Completion witness |
| --- | --- | --- |
| D-019 | Delete generic shared and feature-local request/options/envelope facades; no compatibility aliases | F-015; S-015–017 |
| D-020 | Retire `RequireStatus`, fatal body rewind and `ReadBodyString`; explicit terminal readers replace them | F-016; S-014 |
| D-021 | Retire globally mandatory error details and object-only success data; enforce required message/status | F-017; S-014 |
| D-022 | Preserve domain fixtures, process trusted-client/origin adapters, forms, redirect policy and cookie policy | C-012/015; S-015–017 |
| D-023 | Preserve unique route/storage/security assertions and real application assembly tests | C-015/016; S-013–019 |


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

### HTTP workstreams — authoritative prerequisite chain

| Workflow / slice | Prerequisite | Goal / risk | Validation and binary exit |
| --- | --- | --- | --- |
| WF-13 / S-013 — characterize | root: authorized request and revalidated baseline | Freeze owners/callers/rows; avoid missing consumers | V-015; current representative characterization qualifies and exact portfolio recorded |
| WF-14 / S-014 — responses | WF-13 saved checkpoint | Terminal ownership and correct envelope assertions; avoid double consumption or false constraints | V-016; negative witnesses and affected response consumers pass |
| WF-15 / S-015 — request API | WF-14 saved checkpoint | Explicit clients and atomic direct-call migration; preserve transport policy | V-017; request and process witnesses pass |
| WF-16 / S-016 — shared facades | WF-15 saved checkpoint | Remove application/Auth forwarding; retain domain setup | V-018; all shared facade callers migrated and checks pass |
| WF-17 / S-017 — local facades | WF-16 saved checkpoint | Remove feature duplicates; keep semantic adapters | V-019; disposition ledger complete and checks pass |
| WF-18 / S-018 — routing | WF-17 saved checkpoint | Audit authored selectors/projections; avoid orphan/overlap/identity churn | V-020; generated/schema/catalog/identity checks pass |
| WF-19 / S-019 — final handoff | WF-18 and all applicable correction checkpoints | Finalize, freeze, qualify, audit; avoid stale evidence | V-021/022; final-source portfolio and handoff complete; finishes last |

Every workstream is a separate execution slice. Save and review its tracker update
before starting its successor. Conditional production corrections receive their
own IDs and dependencies; no correction is implied by planning alone.


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

### HTTP implementation decisions and slice readiness

- S-013–019 are complete at E-025–031. Their entry gates required current
  characterization and each predecessor's saved checkpoint; RB-005 is closed.
- S-014 adds `ReadResponseBody(t, response, expectedStatus) []byte` and
  `RequireNoContent(t, response)` (204 and empty bytes). Terminal readers read and
  close before reporting errors; private error-returning readers/validators expose
  failures to unit tests. Strict JSON decoding accepts one complete non-null
  object with existing map/float64 representation. Envelope checks allow additive
  members and optional details. Success data shape is route-specific. Error
  message is nonempty, status equals transport, retryable is boolean and request
  IDs correlate. Replace all status/raw reader consumers together; preserve raw
  downloads, headers and no-content assertions. No body rewind or fake TB framework.
- S-015 makes `DoJSON` and `DoRawJSON` take `*http.Client` after `testing.TB`.
  Migrate every direct caller in the same slice. Keep `NewJSONRequest`, `Do`,
  `WithCookies`, `WithHeader` and literal malformed payloads; delete unused
  `NewAuthenticatedJSONRequest`. Do not mutate global clients or add retry policy.
- S-016 removes generic application/Auth-flow request/options exports, relocates
  the object-shaped `RequireSuccessData` assertion to `httptestx`, and removes
  the redundant error-envelope alias. Preserve semantic login/fixture behavior.
- S-017 removes generic Auth/Timeline/Entities/Evidence/Reference Data/process
  wrappers while keeping explicit form, redirect, origin and trusted-client policy.
- S-018 updates authored inputs and regenerates; new top-level tests receive
  routing in their introducing slice, not only at final routing review.
- S-019 runs finalization before final-source regression, then completes handoff.

Rollback restores a slice's callers, helpers and authored routing together; it
never adds a permanent compatibility adapter. Stricter helper assertions are
intentional test corrections, not production API migrations.


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

### HTTP validation obligations

All commands run through public Make from the repository root. New test row IDs
are authored through `make author-test-row-id`, then generated through Make.
Exact selections and retained roots are recorded at each checkpoint. Source
changes invalidate evidence according to harness compatibility, not informal
claims. Initial representative characterization preserves the existing helper
and application-assembly rows; final regression covers the conservative affected
Go row union below plus added witnesses. No UI/performance/release claim is made.

| ID | Stage / command | Acceptance / failure gate |
| --- | --- | --- |
| V-015 | `make help ALL=1`; per-owner `make task-guide OWNER=<owner> DETAIL=coverage JSON=1`; exact existing helper and assembly `make test-slice` rows | Freeze row union; qualify current baseline before code movement |
| V-016 | New response unit row and affected response-consumer rows through `make test-slice` | Read/close/status/malformed/trailing/null JSON and envelope fields; optional/additive cases; bytes/headers/204 retained |
| V-017 | New request unit row and selected process/custom-client rows through `make test-slice` | Nil/ordinary/raw body, options, TLS/redirect/client behavior; migrated direct callers compile |
| V-018 | Shared facade consumer selections through `make test-slice` | Auth/domain fixtures still preserve literal outcomes and effects |
| V-019 | Feature-local consumer selections through `make test-slice` | No generic forwarding remains; meaningful local adapters preserved |
| V-020 | `make generate`; `make source-check CHECK=generated`; `CHECK=schema`; `CHECK=catalog`; `CHECK=identity` | Unique complete routing and current generated projections |
| V-021 | `make agent-finalize`; `make lint CHECK=go`; `make lint CHECK=backend-boundary`; exact final row union through `make test-slice`; `make harness-observability-check RESULTS_DIR=<exact passing root> CHECK=all` | Final-source compatible passing evidence; no hidden missing obligations |
| V-022 | `make lint-markdown`; direct ID/link/table/dependency/history/diff review | Saved coherent handoff; Markdown lint alone does not cover this tracker |

Pass RESULTS_DIR to finalization only for a compatible successful full warm check;
otherwise report retained maintenance skipped. Do not launch a redundant check
just to obtain that root. Baseline and focused checks have distinct purposes;
final portfolio is selected once after finalization, extending only for repairs
or proven missing obligations.

#### Frozen conservative HTTP regression row union

This table is human handoff material; runtime selection is supplied from authored
JSON catalog inputs, never read from this document.

| Owner | Exact baseline row IDs |
| --- | --- |
| `app.operator` | `app.operator.process.collaboration_requeue_v2_process_contract_4ce8fef4eb`, `app.operator.process.object_store_init_process_transport_400591b90d` |
| `app.server` | `app.server.integration.application_fixtures_own_independent_databases_a_04455c6902`, `app.server.integration.borrowed_application_fixtures_and_deployment_byt_f2fa662581`, `app.server.integration.extension_discovery_session_sliding_7f6ad80724`, `app.server.integration.http_harness_restart_storage`, `app.server.process.a_fresh_process_reaches_health_and_ready_only_af_c93818219a`, `app.server.process.a_real_process_requiring_bootstrap_becomes_ready_41afb1f3e5`, `app.server.process.a_running_server_exposes_no_public_http_or_web_s_2934cde8e0`, `app.server.process.effective_deployment_configuration_fails_the_rea_c17cbafdfc`, `app.server.process.invalid_config_exits_non_zero_emits_structured_s_6c0922ea28`, `app.server.process.process_smoke_exercises_incident_creation_valida_541e2263d1`, `app.server.process.process_smoke_exercises_login_csrf_session_revoc_c6fd6eb90f`, `app.server.process.required_bootstrap_failures_on_the_real_process_9dbfe1d620`, `app.server.process.stale_or_invalid_manifests_are_ignored_when_an_a_1d42fdeb31`, `app.server.process.the_packaged_standalone_server_serves_its_embedd_9d0f816da3`, `app.server.process.the_standalone_server_keeps_harness_routes_disab_4dc8b1f7d9`, `app.server.process.the_standalone_server_process_can_create_evidenc_4ca93ab308`, `app.server.support_unit.application_fixture_admission_validates_complete_2e28540399`, `app.server.support_unit.incident_create_commit_fault_requires_validated_efdc344831`, `app.server.unit.extension_discovery_transport_7f6ad80724` |
| `harness.browser` | `harness.browser.integration.evidence_inspector_fixture`, `harness.browser.integration.evidence_object_fixture`, `harness.browser.integration.managed_service_assertions_for_runtime_reset_rec_79940f98ab`, `harness.browser.integration.performance_fixture_snapshot_lifecycle`, `harness.browser.integration.reset_checkpoint_characterization`, `harness.browser.integration.service_free_contract_assertions_for_runtime_res_04b3f3a218`, `harness.browser.unit.httptestx_helper_contract`, `harness.browser.unit.performance_fixture_lifecycle`, `harness.browser.unit.performance_fixture_objects`, `harness.browser.unit.retired_public_error_control`, `harness.browser.unit.source_owner_contribution_assembler` |
| `module.artifacts` | `module.artifacts.collection_semantics.artifact_collection_mutations_preserve_typed_col_ab3546b03e`, `module.artifacts.collection_semantics.artifact_risk_references_use_canonical_parent_sc_b69b618649`, `module.artifacts.conflict_revalidation.artifact_conflict_resolution_revalidates_source_07b695d48b`, `module.artifacts.import_facade.artifact_import_facade_preserves_registered_targ_a54c97551b`, `module.artifacts.linked_notes.artifact_linked_note_creation_is_atomic_across_f_bce1ac123b`, `module.artifacts.linked_notes.note_association_resource`, `module.artifacts.rollback_provider.artifact_rollback_service_executes_all_eight_sour_0dcd05ed0a`, `module.artifacts.source_mutations.artifact_admission_and_replay_hashing_are_owner_local`, `module.artifacts.source_mutations.coordination_source_atomic_creation`, `module.artifacts.source_mutations.managed_service_assertions_for_artifact_source_c_a67a87fbf9`, `module.artifacts.source_mutations.note_association_admission`, `module.artifacts.source_mutations.service_free_contract_assertions_for_artifact_so_0c20f86ece`, `module.artifacts.surface_contracts.artifact_surface_registry_recognizes_all_eight_c_778e06f594`, `module.artifacts.unit.artifact_import_contribution_rejects_missing_dep_312e585529` |
| `module.assessments` | `module.assessments.store.assessment_assessor_membership_contract_743da311fe`, `module.assessments.store.assessment_import_create_facade_contract_7282c4a7e3`, `module.assessments.store.assessment_projection_query_remains_owner_backed_172051cb0d`, `module.assessments.store.assessment_support_refs_task_linked_record_ids_d_9c08a8f2f2`, `module.assessments.store.facade_transaction_participant_rollback_and_replay`, `module.assessments.store.http_create_replay_projection_publication_characterization`, `module.assessments.store.http_filter_validation_characterization`, `module.assessments.store.import_binding_identity_characterization`, `module.assessments.store.managed_service_assertions_for_compromise_assess_98ae87986a`, `module.assessments.store.service_free_contract_assertions_for_compromise_d48d9308ef`, `module.assessments.unit.facade_and_import_contracts` |
| `module.auth` | `module.auth.integration.actor_fixtures_share_preparation_while_owning_in_64d66bfd0d`, `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c`, `module.auth.integration.administrative_routes_require_current_deployment_71dd666a37`, `module.auth.integration.bootstrap_tokens_work_only_on_totp_begin_and_tot_36751ad438`, `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb`, `module.auth.integration.deployment_admin_binding_create_replay_divergent_85cb20830a`, `module.auth.integration.deployment_audit_route_is_scope_safe_and_keyset_4f8db0c261`, `module.auth.integration.enterprise_provider_discovery_proves_sorted_exac_059973abc3`, `module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af`, `module.auth.integration.real_sockets_on_canonical_ws_v1_incidents_incide_8e29eb6263`, `module.auth.integration.saml_sp_initiated_begin_acs_verifies_determinist_3b5547023b`, `module.auth.integration.secret_envelope_substitution`, `module.auth.integration.slide_session_does_not_regress_persisted_timing_0a4df22db3`, `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6`, `module.auth.integration.user_list_search_tokenizes_identifiers_and_binds_45c3e062a8`, `module.auth.store.one_route_matrix_proves_logout_password_change_t_728b4fb4df`, `module.auth.store.the_real_login_handler_creates_the_sixth_session_b3a9603ee1`, `module.auth.store.the_real_patch_route_rejects_missing_base_user_v_e16c180611`, `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063`, `module.auth.support_unit.session_inspection_helpers_da94c8032f`, `module.auth.unit.admin_list_queries_use_core_list_query_errors_ee8aae9fd6`, `module.auth.unit.administrative_audit_openapi_is_complete_and_exact_7281a6f4d2`, `module.auth.unit.auth_openapi_responses_and_security_match_runtime_7f0e6a2c91`, `module.auth.unit.cookie_mutations_reject_missing_or_invalid_csrf_dfaa827393`, `module.auth.unit.credential_capacity_public_boundary`, `module.auth.unit.deployment_admin_only_credential_routes_fail_clo_5cfa2f8dce`, `module.auth.unit.enterprise_auth_error_registries_closed_8f9cca353a`, `module.auth.unit.enterprise_route_registration_requires_exact_adm_1f99782c9f`, `module.auth.unit.login_request_shape_validation_rejects_unknown_t_3a3eb47266`, `module.auth.unit.the_password_change_route_proves_exact_current_p_338ff542a2`, `module.auth.unit.the_real_credential_state_route_returns_the_safe_db26b28985`, `module.auth.unit.the_real_get_auth_session_route_rejects_paginati_14b12305f4`, `module.auth.unit.the_real_login_handler_creates_exactly_one_sessi_3b913a16a6`, `module.auth.unit.the_real_login_handler_trims_and_case_folds_the_9ed31374cf`, `module.auth.unit.the_totp_begin_and_complete_handlers_enforce_exa_59570a1bac`, `module.auth.unit.user_create_rejects_undeclared_is_active_rejects_839f6b57c5` |
| `module.collaboration` | `module.collaboration.integration.closed_incident_socket_terminates_before_writabl_8057ea1833`, `module.collaboration.integration.cookie_authenticated_browser_socket_upgrades_rej_73faec66da`, `module.collaboration.integration.durable_incident_stream_commits_intents_atomically_s_80128fe153`, `module.collaboration.integration.entity_candidate_discovery_does_not_publish`, `module.collaboration.integration.resume_with_a_valid_replay_token_replays_replaya_1826b38095`, `module.collaboration.integration.two_real_clients_connect_to_the_same_incident_so_257bcc588c`, `module.collaboration.store.presence_payloads_are_incident_scoped_and_epheme_185de72cf2`, `module.collaboration.store.strict_frame_and_failure_contract_a3b4c5d6e7`, `module.collaboration.store.the_first_application_message_on_the_incident_so_11b91f6bd8`, `module.collaboration.support_integration.incident_socket_revocation_sources_00f908473e`, `module.collaboration.support_unit.incident_test_client_preserves_ordered_observati_6defc64846`, `module.collaboration.support_unit.lifecycle_telemetry_vocabulary_b1c2d3e4f5`, `module.collaboration.support_unit.semantic_hub_and_payload_ownership_d1e2f3a4b5`, `module.collaboration.support_unit.semantic_telemetry_ownership_e1f2a3b4c5`, `module.collaboration.unit.event_intent_exports_no_mutable_fields_unit_1d9b83aca1`, `module.collaboration.unit.publication_catalog_rejects_incomplete_and_ambig_858f5dab4f`, `module.collaboration.unit.replayable_intent_validation_admits_owner_additi_39345aa09a`, `module.collaboration.unit.runtime_construction_rejects_every_missing_depen_eb9cd9aca2` |
| `module.entities` | `module.entities.integration.candidate_discovery`, `module.entities.integration.entity_party_assignment_writes_preserve_authorit_9fb842a70d`, `module.entities.integration.host_and_identity_create_routes_reuse_exact_matc_f443e2591f`, `module.entities.integration.the_explicit_merge_route_repoints_live_fan_out_p_0ec76e8044`, `module.entities.integration.the_mention_resolve_route_persists_durable_resol_d68c1befd6`, `module.entities.store.create_from_mention_store_paths_promote_unresolv_11902ddb10`, `module.entities.store.dismiss_and_restore_store_operations_preserve_ra_df842d8c6d`, `module.entities.store.exact_match_precedence_resolves_reuse_decisions_949743e789`, `module.entities.store.explicit_merge_store_behavior_repoints_live_ment_0560122310`, `module.entities.store.mention_commands_apply_injected_timeline_effects_4764d2ea7c`, `module.entities.support_integration.active_identifier_claims_8f302d2a40`, `module.entities.support_integration.entity_alias_migration_31`, `module.entities.support_integration.entity_identifier_claims_37_3ba3e2aaaf`, `module.entities.support_integration.entity_incident_bundle_invariant_closure_92db429a94`, `module.entities.support_integration.entity_relationship_route_contracts_and_conseque_44a6032c5e`, `module.entities.support_integration.entity_source_integrity_migration_36_e6d97d1e54`, `module.entities.support_integration.integration_entity_create_idempotency_is_actor_s_3c8ed1fe5c`, `module.entities.support_integration.integration_record_envelope_head_schema_3b559686a9`, `module.entities.support_integration.timeline_mention_projection_upgrade_preserves_so_dabe8cb065`, `module.entities.unit.entity_revision_contribution_owns_current_snapsh_4a2385fd4a`, `module.entities.unit.openapi_mutation_contract_d0fe61dbba`, `module.entities.unit.root_boundary_guards_e292ec1361` |
| `module.evidence` | `module.evidence.integration.a_real_object_store_upload_can_be_created_throug_9924ede0b0`, `module.evidence.integration.a_redeemed_preview_or_download_handle_fails_clos_cbb006eb24`, `module.evidence.integration.a_successful_evidence_attach_publishes_canonical_d63a42e346`, `module.evidence.integration.attach_route_validation_authorization_route_owne_7567e19cd9`, `module.evidence.integration.attached_evidence_projection_storage_can_be_corr_933714dcfe`, `module.evidence.integration.background_verification_cancellation`, `module.evidence.integration.blob_association_concurrent_race_has_one_winner_9d6a5c2e1b`, `module.evidence.integration.blob_uniqueness_migration_round_trip_6de21f480c`, `module.evidence.integration.dedicated_database_assertions_for_when_quarantin_f5c765cb0c`, `module.evidence.integration.durable_cleanup_claim_engine_5be2d5d937`, `module.evidence.integration.evidence_blob_attachment_and_handle_route_contra_9ce6c25e48`, `module.evidence.integration.evidence_handle_issuance_reports_registered_unav_e4474a6987`, `module.evidence.integration.expired_slot_replay_returns_the_same_expired_slo_6d29ac9f3a`, `module.evidence.integration.initial_blob_create_atomic_replay_concealment_73c81fe329`, `module.evidence.integration.preview_and_download_handle_issuance_accepts_onl_84f55ae52b`, `module.evidence.integration.public_evidence_responses_do_not_leak_object_sto_9e5c545bd7`, `module.evidence.integration.single_connection_rollback_assertions_for_when_q_7d5c25603e`, `module.evidence.integration.the_object_blob_create_route_enforces_request_sh_17b6e1b674`, `module.evidence.integration.verification_cutover`, `module.evidence.store.attached_evidence_create_and_patch_aa33ea0168`, `module.evidence.store.blob_association_reuse_is_concealed_8b3e27a14f`, `module.evidence.store.blob_create_idempotency_is_keyed_by_actor_user_i_f161d12732`, `module.evidence.store.blob_create_size_ceilings_reject_oversize_byte_s_9ba60e8bdd`, `module.evidence.store.download_responses_use_authoritative_metadata_fo_a55dbb9aac`, `module.evidence.store.evidence_lifecycle_state_remains_separate_from_b_df4f302819`, `module.evidence.store.handle_access_revalidates_current_evidence_blob_7f02f9ebce`, `module.evidence.store.post_api_v1_evidence_records_record_id_attach_bl_14293e0ccf`, `module.evidence.store.post_api_v1_object_blobs_enforces_required_field_6c09c04c66`, `module.evidence.store.preview_handle_and_download_handle_issuance_rout_e9c8e5a3cd`, `module.evidence.unit.bounded_metadata_admission_f99b626064`, `module.evidence.unit.create_effect_order_and_fault_boundary_5d70a3c812`, `module.evidence.unit.create_signal_and_initial_lifecycle_matrix_4f9a8b671c`, `module.evidence.unit.evidence_and_timeline_evidence_derived_field_key_57b5d6b63a`, `module.evidence.unit.evidence_create_attach_preview_issuance_and_rede_cef38b411a`, `module.evidence.unit.evidence_direct_party_references_admit_exact_sta_d51415ee2a`, `module.evidence.unit.evidence_handle_request_rejects_unknown_members_e3ebd53914`, `module.evidence.unit.evidence_mutation_admission_and_replay_hashing_are_o_48d36a71bc`, `module.evidence.unit.evidence_open_apiincludes_route_family_866f0d0acf`, `module.evidence.unit.owner_private_cleanup_dispatcher_and_typed_storage_a1b2c3d4e5`, `module.evidence.unit.provider_contribution_closure_f247686138` |
| `module.imports` | `module.imports.integration.apply_transaction_current_state_bc8ad0ac32`, `module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88`, `module.imports.integration.evidence_create_parity_ac652dc3b3`, `module.imports.integration.fresh_session_explicit_reimport_workflow_18fc21c7f5`, `module.imports.integration.import_upload_metadata_that_is_syntactically_val_6c1083ba3e`, `module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873`, `module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`, `module.imports.integration.partial_cancellation_outcomes_a14b66f830`, `module.imports.integration.selection_overlap_reselection_lifecycle_f8525366fa`, `module.imports.integration.timeline_unit_atomic_rollback_61c5e452e8`, `module.imports.integration.uncommitted_unit_cancellation`, `module.imports.integration.unit_commit_crash_recovery_108cd054a6`, `module.imports.integration.upload_admission_replay_complete_b8684935f2`, `module.imports.integration.view_owner_write_null_application_rs08_8bb5e55b95`, `module.imports.integration.xlsx_import_discovery_accepts_a_bounded_open_xml_d7c25ef6c8`, `module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912`, `module.imports.unit.behavior_contract`, `module.imports.unit.canonical_write_null_and_safe_owner_translation_rs08_3f58f0c338`, `module.imports.unit.indicator_owner_catalog_identity`, `module.imports.unit.module_composition_contract`, `module.imports.unit.xlsx_locator_security_and_formula_index_41d0b577ac` |
| `module.incidentbundles` | `module.incidentbundles.integration.actor_export_combines_native_users_and_imported_f4f3ed2ecc`, `module.incidentbundles.integration.extension_portability_matrix_and_atomic_import_1b27fc2d91`, `module.incidentbundles.integration.incident_bundle_descriptor_reads_reject_singleto_ea18aa4fe5`, `module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8`, `module.incidentbundles.integration.incident_bundle_export_jobs_use_the_incident_sco_d94a8de941`, `module.incidentbundles.integration.incident_bundle_export_uses_the_common_job_resou_a1e51b82d7`, `module.incidentbundles.integration.incident_bundle_import_final_publication_recheck_b7b07b3e93`, `module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755`, `module.incidentbundles.integration.incident_bundle_import_upload_envelope_failures_bf5449e3d4`, `module.incidentbundles.integration.incident_bundle_import_z_evidence_cleanup_a3d619b413`, `module.incidentbundles.integration.incident_bundle_retired_version_is_rejected_with_1ec7e171fe`, `module.incidentbundles.integration.incident_bundle_storage_reference_cutover_is_cle_7e4d8743ff`, `module.incidentbundles.integration.layout_legacy_conversion_replay`, `module.incidentbundles.integration.network_flow_retained_state_blocks_bundle_export_7dcd6f85c2`, `module.incidentbundles.unit.compression_ratio_boundary_uses_threshold_compar_9c4165a1f9`, `module.incidentbundles.unit.cross_owner_import_descriptor_is_closed_9f6d848f63`, `module.incidentbundles.unit.incident_bundle_archive_construction_emits_deter_d848a55dc0`, `module.incidentbundles.unit.incident_bundle_contract_registries_use_exactly_1233d2b3f3`, `module.incidentbundles.unit.incident_bundle_export_requests_canonicalize_sel_b7c820b069`, `module.incidentbundles.unit.incident_bundle_logical_refs_are_root_free_228bd44f05`, `module.incidentbundles.unit.incident_bundle_route_registration_remains_exten_943ad7d722`, `module.incidentbundles.unit.incident_bundle_verification_accepts_safe_zip_an_892d158923`, `module.incidentbundles.unit.incident_bundle_worker_terminal_success_transiti_66f6ace21e`, `module.incidentbundles.unit.named_job_runner_is_required_for_worker_dispatch_09c1f60805`, `module.incidentbundles.unit.open_api_and_error_registries_expose_incident_bu_25db0a5cae` |
| `module.incidents` | `module.incidents.integration.concurrent_cross_admin_deletion_preserves_an_ad_9f25d8ef61`, `module.incidents.integration.incident_collection_authenticated_authorization_dab8552d8f`, `module.incidents.integration.incident_create_persists_the_incident_row_bootst_0bd8d02f12`, `module.incidents.integration.incident_lifecycle_close_reopen_authorization_re_6ac1f84d20`, `module.incidents.integration.incident_list_search_status_and_cursor_scope_80dd540eb3`, `module.incidents.integration.incident_membership_current_authorization_5018f673df`, `module.incidents.integration.incident_resource_current_authorization_b010ae008a`, `module.incidents.integration.incident_socket_current_admission_697be7989c`, `module.incidents.integration.incident_timeline_mutations_current_authorizatio_ecdb3731da`, `module.incidents.integration.incident_workbook_preferences_current_authorizat_9ddc0742c6`, `module.incidents.integration.incident_workbook_queries_current_authorization_581c7192f2`, `module.incidents.integration.live_membership_patch_to_the_current_role_return_7a14cba87a`, `module.incidents.integration.membership_audit_route_is_scope_safe_and_keyset_8e31c7f24a`, `module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6`, `module.incidents.integration.real_incident_patch_persists_only_patchable_meta_9fc8fc207f`, `module.incidents.integration.reserved_but_unclaimed_family_roots_and_descenda_3cbf9bf6e6`, `module.incidents.store.concurrent_mutation_admission`, `module.incidents.store.incident_create_idempotency_scopes_by_actor_plus_b2c23d5e25`, `module.incidents.store.managed_service_assertions_for_store_create_inci_a391a222d4`, `module.incidents.store.service_free_contract_assertions_for_store_creat_f476b0f36b`, `module.incidents.store.store_backed_membership_patch_and_delete_reject_17d0f4ff28`, `module.incidents.store.store_update_incident_returns_a_typed_incident_v_0669e48fa6`, `module.incidents.store.terminal_mutation_results_distinguish_fresh_comm_2101097f8d`, `module.incidents.support_integration.authenticated_incident_creation_and_normalized_r_e2cd9c99b5`, `module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972`, `module.incidents.support_integration.incident_list_uses_live_first_page_independently_f2ad5f1106`, `module.incidents.support_integration.membership_create_replay_returns_original_and_di_98e85ad975`, `module.incidents.unit.admitted_mutation_hashes_match_compatibility_dig_6c66c494bf`, `module.incidents.unit.application_constructor_guards_i2`, `module.incidents.unit.idempotency_and_private_helper_contracts_i2`, `module.incidents.unit.incident_16_source_codec_i3`, `module.incidents.unit.metadata_field_intent` |
| `module.indicators` | `module.indicators.identity.active_identity_claims_follow_records_and_rebuild`, `module.indicators.identity.create_command_projects_identity_validation`, `module.indicators.identity.direct_create_serializes_transaction_identity_an_8f843da442`, `module.indicators.identity.z_concurrent_find_or_create_convergence`, `module.indicators.integration.indicator_create_and_query_routes_persist_canoni_e3a29b7aaf`, `module.indicators.integration.production_observation_and_lifecycle_routes`, `module.indicators.lifecycle.exact_production_vocabulary`, `module.indicators.observation.closed_transition_state_machine`, `module.indicators.observation.source_span_utf8_admission`, `module.indicators.origin.observation_origin_registry_and_producer_context_40d43e15e1`, `module.indicators.origin.observation_origin_storage_constraint_and_system_93f5dcc952`, `module.indicators.portability.indicator_source_rows_are_deterministic_complete_68b42b02f6`, `module.indicators.portability.partition_deterministic`, `module.indicators.portability.state_invariants`, `module.indicators.projection.application_mutations_delegate_row_refresh_and__026301a7c8`, `module.indicators.projection.owner_contribution_identity_and_surface_parity`, `module.indicators.replay.stored_create_replay_preserves_bytes_and_identit_6c548da7ff`, `module.indicators.source_state.projection_parity`, `module.indicators.storage.envelope_contract_migration_down_up`, `module.indicators.storage.lifecycle_support_reference_integrity`, `module.indicators.store.indicator_storage_keeps_source_bound_observation_07c064324e`, `module.indicators.store.indicators_persist_as_canonical_indicator_record_3998750e91`, `module.indicators.structure.indicator_application_requires_complete_composition_6ecc4edf41`, `module.indicators.target_resolution.role_classification_and_storage_propagation`, `module.indicators.target_resolution.role_failures_are_atomic`, `module.indicators.transaction.indicator_repository_writes_roll_back_when_revis_c37659989d`, `module.indicators.unit.indicator_revision_contribution_owns_current_sna_b8894bd1ae`, `module.indicators.vocabulary.contract_parity_and_portable_admission` |
| `module.jobapi` | `module.jobapi.integration.deployment_scoped_common_job_reads_and_cancel_re_85147de44e`, `module.jobapi.integration.expired_job_read_and_cancel_masking`, `module.jobapi.integration.incident_scoped_common_job_read_and_cancel_autho_821120b13c`, `module.jobapi.integration.the_reusable_deployment_admin_incident_membershi_b1cbd96a5a`, `module.jobapi.unit.internal_error_safety` |
| `module.links` | `module.links.integration.active_links_and_tags_views_v1_contract_eaf4aac3b7`, `module.links.integration.link_and_tag_mutations_atomically_update_project_03f8283906`, `module.links.integration.party_assignment_source_facts_preserve_snapshot_963c86bbbb`, `module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1`, `module.links.unit.links_item_refs_are_canonical_ade3870b93`, `module.links.unit.links_revision_contribution_owns_current_target_08116fb575`, `module.links.unit.source_state_contributions_remain_exact_8d3f67c521` |
| `module.networkflow` | `module.networkflow.integration.bounded_graph_contributor_pipeline`, `module.networkflow.integration.effective_resource_limit_discovery`, `module.networkflow.integration.extension_state_v4_v1_rejection`, `module.networkflow.integration.graph_result_cleanup`, `module.networkflow.integration.graph_result_cleanup_races`, `module.networkflow.integration.network_flow_routes_remain_unclaimed_by_default_e8dabf0ee1`, `module.networkflow.integration.pagination_recovery`, `module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5`, `module.networkflow.integration.resource_intent_failure_rolls_back_source_mutation_8a4c3d2e1f`, `module.networkflow.integration.saved_graph_cutover_v6`, `module.networkflow.integration.saved_graph_lifecycle_v2`, `module.networkflow.integration.table_lifecycle_admission_replay`, `module.networkflow.integration.time_bucket_saved_graph_lifecycle`, `module.networkflow.measurement.backend_graph_capacity_certification`, `module.networkflow.process.the_packaged_standalone_server_composes_the_netw_400a31ad27`, `module.networkflow.store.graph_view_declaration_persistence`, `module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232`, `module.networkflow.unit.bounded_graph_aggregation_and_selectors`, `module.networkflow.unit.cleanup_dispatcher_lifecycle`, `module.networkflow.unit.configuration_claim_and_manifest_admission`, `module.networkflow.unit.current_cursor_digest_lifecycle`, `module.networkflow.unit.graph_telemetry_boundary`, `module.networkflow.unit.import_owner_error_translation_rs08_3d3f02e050`, `module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700`, `module.networkflow.unit.network_flow_selector_covers_a_deployment_admin_af7fc1917b`, `module.networkflow.unit.network_flow_selector_covers_a_table_query_uses_6caf43b23c`, `module.networkflow.unit.network_flow_selector_covers_cidr_filtering_reje_b5a9752753`, `module.networkflow.unit.network_flow_selector_covers_cisco_sna_required_66f0e756eb`, `module.networkflow.unit.network_flow_selector_covers_cisco_sna_v1_permit_a2c32abbbf`, `module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99`, `module.networkflow.unit.network_flow_selector_covers_cursor_continuation_0f6e2c662c`, `module.networkflow.unit.network_flow_selector_covers_diagnostic_discover_6f2d284ad8`, `module.networkflow.unit.network_flow_selector_covers_duplicate_csv_heade_61bc2b1f54`, `module.networkflow.unit.network_flow_selector_covers_effective_limits_ar_2cc8742d08`, `module.networkflow.unit.network_flow_selector_covers_empty_csv_forbidden_421ddebc7c`, `module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847`, `module.networkflow.unit.network_flow_selector_covers_every_network_flow_13fc533c41`, `module.networkflow.unit.network_flow_selector_covers_every_network_flow_43bd8f4992`, `module.networkflow.unit.network_flow_selector_covers_every_route_local_e_5288e318ae`, `module.networkflow.unit.network_flow_selector_covers_import_apply_replay_8b8c7de503`, `module.networkflow.unit.network_flow_selector_covers_import_facade_previ_2371e1f671`, `module.networkflow.unit.network_flow_selector_covers_interface_fields_st_6333a351b3`, `module.networkflow.unit.network_flow_selector_covers_ipv6_outputs_follow_e1193869aa`, `module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37`, `module.networkflow.unit.network_flow_selector_covers_logs_telemetry_audi_61f27d1e10`, `module.networkflow.unit.network_flow_selector_covers_no_network_flow_imp_8e4a194dd0`, `module.networkflow.unit.network_flow_selector_covers_reserved_source_pro_527d462af8`, `module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58`, `module.networkflow.unit.network_flow_selector_covers_source_filename_dis_208f945599`, `module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a`, `module.networkflow.unit.network_flow_selector_covers_the_created_table_t_27dec58147`, `module.networkflow.unit.network_flow_selector_covers_the_table_lifecycle_2abb92e0a0`, `module.networkflow.unit.network_flow_selector_covers_timestamp_profile_p_bdf924da13`, `module.networkflow.unit.network_flow_selector_covers_trim_ascii_space_v1_3a1dbda5f2`, `module.networkflow.unit.network_flow_selector_covers_uint64_decimal_stri_bb2238c22e`, `module.networkflow.unit.network_flow_selector_covers_unmapped_raw_values_ad3df69663`, `module.networkflow.unit.query_authoring_admission`, `module.networkflow.unit.saved_graph_receipt_integrity`, `module.networkflow.unit.table_lifecycle_contract`, `module.networkflow.unit.time_bucket_graph_backend`, `module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b`, `module.networkflow.unit.transaction_lifecycle`, `module.networkflow.unit.unicode17_text_canonicalization_3de608c3dc` |
| `module.parties` | `module.parties.claims.party_active_claims_are_migration_safe_concurren_9ad2db3764`, `module.parties.composition.party_mutation_and_import_composition_exposes_on_8a69574087`, `module.parties.portability.party_portable_public_failures_are_closed_and_at_cc2b089315`, `module.parties.references.party_reference_targets_must_be_active_parties_i_6ca84ebe53`, `module.parties.store.party_create_reuses_exactly_one_active_same_inci_f7f18f1c5c`, `module.parties.unit.party_create_admission_defaults_are_shared_unit_fdc0b6d64d`, `module.parties.unit.party_mutation_admission_and_replay_hashing_are_o_7e9a31d4c2`, `module.parties.unit.party_mutation_admission_public_identity_unit_27b3ca8c40`, `module.parties.unit.party_mutation_hash_normalization_equivalence_un_c73a730e10`, `module.parties.unit.row_version_consumption_requires_exact_int64_uni_2f197503f1` |
| `module.projections` | `module.projections.foundation.private_catalog_validation_iteration2`, `module.projections.provider.catalog_validation_684d752cb4`, `module.projections.query.contract_shape_and_keyset_59528aa56d`, `module.projections.query.timeline_keyset_bounds_e67384caff`, `module.projections.rebuild.managed_service_assertions_for_atomic_catalog_a9d64032f0`, `module.projections.rebuild.managed_service_assertions_for_restore_contract_f5f84e9227`, `module.projections.rebuild.service_free_contract_assertions_for_atomic_cata_73ede45d54`, `module.projections.rebuild.service_free_contract_assertions_for_restore_con_663c31db5c`, `module.projections.source.party_projection_source_enumeration_is_determini_69ccf7df7c`, `module.projections.storage.dedicated_database_assertions_for_query_row_pari_2cb81df082`, `module.projections.storage.single_connection_rollback_assertions_for_query_028ff27904`, `module.projections.telemetry.safe_boundary_8e0e774b19` |
| `module.recovery` | `module.recovery.integration.evidence_inventory_closure_9f2714ac10`, `module.recovery.integration.real_postgres_backed_runtime_metadata_persistenc_d652f0998b`, `module.recovery.integration.reset_preserves_crypto_format_identity`, `module.recovery.integration.restore_candidate_uses_object_store_backup_lates_c2cf211fbd`, `module.recovery.integration.restore_readiness_selects_the_latest_retained_su_e71d710085`, `module.recovery.integration.restore_verification_atomic_completion_449dbe5d91`, `module.recovery.integration.selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a`, `module.recovery.integration.selecting_a_retained_backup_set_missing_a_requir_d6af667fba`, `module.recovery.process.canonical_operator_process_evidence_maps_the_imp_9808fdd4e9`, `module.recovery.process.due_verification_continues_after_a_determinate_f_2af5cf65e1`, `module.recovery.process.portable_source_loss_restore`, `module.recovery.process.the_packaged_server_serves_the_coherently_restor_6b1731590e`, `module.recovery.store.durable_backup_set_creation_requires_artifact_ba_d68130215e`, `module.recovery.store.managed_service_assertions_for_incident_bearing_f5fc1b4643`, `module.recovery.store.service_free_contract_assertions_for_incident_be_c37e0a4dad`, `module.recovery.unit.authored_public_route_inventory_exposes_no_api_v_bb82dec1de`, `module.recovery.unit.graph_restore_v4_contract_projection`, `module.recovery.unit.object_store_snapshot_carries_restorable_bytes_3d67535266`, `module.recovery.unit.owner_registered_workbook_probe_and_v2_artifact_a881cb6053`, `module.recovery.unit.recovery_generation_selection_drives_verificatio_2a8734f884`, `module.recovery.unit.streaming_authenticated_envelope_v2_75ff1a51f9`, `module.recovery.unit.support_object_store_backup_object_store_backup_445727db3c`, `module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c` |
| `module.reference_data` | `module.reference_data.integration.a_fresh_disconnected_runtime_seeds_exactly_the_t_aa72222238`, `module.reference_data.integration.absent_disabled_failed_and_missing_optional_refe_e5bec4b7f9`, `module.reference_data.integration.all_profiles_consumer_boundary`, `module.reference_data.integration.canonical_persistence`, `module.reference_data.integration.canonical_reference_data_cutovers_reject_incompa_3682e9346f`, `module.reference_data.integration.disconnected_reference_pack_failure_modes_produc_d1bffa4cca`, `module.reference_data.integration.omitted_reference_pack_refresh_selectors_replay_a29a9f95dc`, `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a`, `module.reference_data.integration.reference_pack_import_admission_returns_a_queued_9940d5be50`, `module.reference_data.integration.reference_pack_import_jobs_re_derive_route_famil_dd456a1690`, `module.reference_data.integration.reference_pack_import_uses_the_shared_upload_env_95574f69b8`, `module.reference_data.integration.reference_pack_storage_reference_cutover_is_reset_dc28423946`, `module.reference_data.integration.reference_pack_upload_envelope_failures_create_n_710dd1d70f`, `module.reference_data.unit.canonical_consumers`, `module.reference_data.unit.generated_open_api_and_error_registry_artifacts_0843af138f`, `module.reference_data.unit.reference_pack_logical_refs_are_root_free_96bb19c696`, `module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda`, `module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2`, `module.reference_data.unit.reference_pack_verifier_fails_closed_for_extract_ae36003205`, `module.reference_data.unit.support_reference_pack_list_query_uses_shared_li_ccbac49024` |
| `module.reportcomposition` | `module.reportcomposition.integration.preview_reporting_delegation_ecdd2bc64a`, `module.reportcomposition.integration.public_lifecycle_authority`, `module.reportcomposition.integration.resolve_release_tuple_tx_reason_matrix_ad31a02455`, `module.reportcomposition.unit.canonical_digests` |
| `module.reporting` | `module.reporting.integration.exact_graph_result_lease_lifecycle_8f1c5c43a2`, `module.reporting.integration.execution_payload_boundary`, `module.reporting.integration.external_releases_require_distinct_reviewer_and_0df12b0a47`, `module.reporting.integration.managed_service_assertions_for_evidence_provider_0b8acd1609`, `module.reporting.integration.renderer_installed_sandbox`, `module.reporting.integration.renderer_stage_boundary`, `module.reporting.integration.reporting_identity_cutover_preserves_incompatibl_62c39cd14e`, `module.reporting.integration.service_free_contract_assertions_for_evidence_pr_80ebcf69f9`, `module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c`, `module.reporting.integration.snapshot_reporting_route_idempotency_treats_an_o_1d2451a56f`, `module.reporting.integration.snapshot_reporting_singleton_reads_and_release_a_943c339eea`, `module.reporting.unit.admitted_render_export_participant_c0da40986c`, `module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393`, `module.reporting.unit.frozen_output_options`, `module.reporting.unit.frozen_release_payload`, `module.reporting.unit.frozen_snapshot_payload`, `module.reporting.unit.graph_v2_exact_selection_and_rendering_f54d8b1a62`, `module.reporting.unit.owner_boundaries`, `module.reporting.unit.renderer_stage_result_admission`, `module.reporting.unit.reporting_job_worker_cancels_before_work_68d5024f75`, `module.reporting.unit.snapshot_record_materialization`, `module.reporting.unit.snapshot_reporting_open_api_uses_reusable_closed_0d1a4531ba`, `module.reporting.unit.snapshot_reporting_post_redaction_validation_fai_9df5c768cd`, `module.reporting.unit.snapshot_reporting_recipient_redaction_profiles_75c820718c`, `module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d`, `module.reporting.unit.snapshot_reporting_rejects_duplicate_same_preced_7daa9a6a07`, `module.reporting.unit.snapshot_reporting_route_decoders_reject_legacy_9d9f64130f`, `module.reporting.unit.source_owner_orchestration` |
| `module.revisions` | `module.revisions.delete_restore.managed_service_assertions_for_source_adapter_70c51a4a80`, `module.revisions.delete_restore.service_free_contract_assertions_for_source_adap_a647a96f3a`, `module.revisions.delete_restore.source_port_catalog`, `module.revisions.facade.construction_guards`, `module.revisions.integration.covers_delete_restore_single_entry_rollback_dire_b348439e6d`, `module.revisions.integration.covers_stale_restore_plus_stale_or_non_reversibl_637111e642`, `module.revisions.integration.history_logical_page_boundaries`, `module.revisions.integration.history_pagination_remains_bound_to_record_id_re_a92fcdb263`, `module.revisions.integration.history_source_family_paging`, `module.revisions.integration.indicator_observation_create_and_resolution_targ_6249c556e2`, `module.revisions.integration.merge_aware_change_set_rollback_restores_survivo_130bd55b07`, `module.revisions.integration.migration_61_admits_fresh_schema_and_rejects_leg_20baec4d72`, `module.revisions.integration.projection_dependencies`, `module.revisions.integration.retained_history_evidence_for_an_extant_record_r_b590f9c4b1`, `module.revisions.portability.atomicity`, `module.revisions.portability.attribution_boundary`, `module.revisions.portability.contract_and_attribution`, `module.revisions.portability.managed_service_assertions_for_allocator_1e7086ab36`, `module.revisions.portability.service_free_contract_assertions_for_allocator_b4f5c965d4`, `module.revisions.providers.catalog_guards`, `module.revisions.rollback.changed_field_completeness`, `module.revisions.rollback.decomposition_8c5e3f2a10`, `module.revisions.rollback.transaction_failure_injection_8c5e3f2a11`, `module.revisions.store.covers_rollback_route_admission_derived_rollback_c6c889199c`, `module.revisions.store.covers_shared_destructive_operation_lock_precede_02f3fc18c4`, `module.revisions.store.covers_the_first_class_record_soft_delete_route_30cb39b83e`, `module.revisions.store.covers_the_restore_route_contract_reviewer_admin_a2df5bc94d`, `module.revisions.store.destructive_commands_authorize_the_current_actor_ad6fddd708`, `module.revisions.store.history_bounded_projection`, `module.revisions.store.history_entry_ref_is_present_only_for_a_logical_5f0060122a`, `module.revisions.store.history_lookup_uses_stored_associations_independ_c8be17a439`, `module.revisions.store.history_query_repository_mapping_7b4d2e9c11`, `module.revisions.store.http_security_precedence_4e7a9c2d13`, `module.revisions.store.managed_service_assertions_for_get_api_v1_record_049278bce8`, `module.revisions.store.retained_history_invariants_for_extant_records_r_b40fb75124`, `module.revisions.store.revision_window_utc_ordering_51f2a6b3c5`, `module.revisions.store.rollback_projection_and_publication_participants_f7381e09d3`, `module.revisions.store.service_free_contract_assertions_for_get_api_v1_a2f18fa208`, `module.revisions.support_unit.history_materialization_attribution_and_page_7b4d2e9c10`, `module.revisions.support_unit.http_application_boundary_4e7a9c2d12`, `module.revisions.unit.generic_rollback_dispatch_executes_a_newly_contr_19b83ad4e2`, `module.revisions.unit.projection_scope` |
| `module.savedviews` | `module.savedviews.integration.imported_saved_views_preserve_portable_owners_se_45009255f0`, `module.savedviews.integration.independent_read_contract`, `module.savedviews.integration.saved_view_create_update_duplicate_and_delete_pe_5f93d857d3`, `module.savedviews.integration.saved_view_export_emits_the_exact_canonical_elev_fb78b01a43`, `module.savedviews.integration.saved_views_storage_hardening_rejects_invalid_le_debf877cb1`, `module.savedviews.integration.timestamp_saved_queries_preserve_precision_and_r_993e5032bf`, `module.savedviews.store.managed_service_assertions_for_ordinary_saved_vi_fc464a05e5`, `module.savedviews.store.ordinary_saved_view_patch_enforces_the_mutable_f_9554a5d90f`, `module.savedviews.store.service_free_contract_assertions_for_ordinary_sa_27bac8bf90`, `module.savedviews.unit.independent_read_query_contract`, `module.savedviews.unit.saved_view_application_policy_preserves_visibili_5af44d4ba8`, `module.savedviews.unit.saved_view_scope_vocabulary_is_exactly_private_s_aa8991ad84`, `module.savedviews.unit.strict_saved_views_preparation_rejects_illegal_i_d1e9382c70`, `module.savedviews.unit.strict_saved_views_preparation_rejects_unbounded_0ac27d4f10` |
| `module.tasksdecisions` | `module.tasksdecisions.composition.mutation_and_import_composition_fails_closed_and_9add8849f4`, `module.tasksdecisions.conflict_composition.application_workbook_conflict_resolution_contrib_10e9976746`, `module.tasksdecisions.import_create.owner_reference_contract_and_atomicity`, `module.tasksdecisions.incident_portability.source_port_invariants_and_atomicity`, `module.tasksdecisions.source_mutations.task_and_decision_mutation_admission_and_replay_6d4ea5900a`, `module.tasksdecisions.source_workbook.managed_service_assertions_for_task_requests_and_9063e4f330`, `module.tasksdecisions.source_workbook.service_free_contract_assertions_for_task_reques_6b1b390c49`, `module.tasksdecisions.unit.task_decision_direct_references_admit_exact_stab_7bf2beff80` |
| `module.timeline` | `module.timeline.integration.create_patch_mark_reviewed_supersede_query_and_w_90276912d7`, `module.timeline.integration.incident_role_authorization_review_and_supersede_62f3d384f6`, `module.timeline.integration.real_http_create_patch_review_supersede_replay_m_37e5c78ca1`, `module.timeline.integration.rough_and_uncertain_timeline_capture_preserves_n_a868f4aa14`, `module.timeline.integration.same_field_timeline_patch_conflicts_are_transpor_808ccc2585`, `module.timeline.integration.the_query_route_reads_projection_rows_instead_of_2edde6eb1c`, `module.timeline.integration.timeline_create_and_patch_routes_persist_manual_ac8c9abb5b`, `module.timeline.integration.timeline_patch_and_query_routes_auto_resolve_onl_8ab2448829`, `module.timeline.store.a_committed_timeline_create_initializes_capture_e04b19cb10`, `module.timeline.store.a_material_edit_demotes_reviewed_enriched_legal_29bb8697de`, `module.timeline.store.a_one_value_timeline_create_commits_a_real_row_a_1964bc0936`, `module.timeline.store.closed_incidents_remain_readable_and_permit_exac_1af610054a`, `module.timeline.store.patch_replay_returns_the_original_committed_resu_b93ddcbc77`, `module.timeline.store.repeated_identical_mention_tokens_persist_as_sep_d79faec8e6`, `module.timeline.store.stale_timeline_patches_derive_committed_writable_b89a3c27fa`, `module.timeline.store.successful_timeline_create_and_patch_each_write_686ca358f8`, `module.timeline.store.supersede_with_replacement_rejects_illegal_targe_e4d81b8a06`, `module.timeline.store.the_first_material_patch_transitions_rough_enric_c36dc39675`, `module.timeline.store.timeline_mention_fields_prove_mention_origin_bin_46b2abfeb9`, `module.timeline.store.timeline_query_accepts_timeline_evidence_count_a_22cc245694`, `module.timeline.support_integration.integration_route_idempotency_is_actor_scoped_ab7992bb7d`, `module.timeline.support_unit.clipboard_paste_parsing_mapping_provenance_and_b_43e8f0fd4f`, `module.timeline.support_unit.timeline_import_normalization_preserves_owner_po_0e92bc5ae9`, `module.timeline.support_unit.unit_capture_state_and_payload_helpers_772e7f2e60`, `module.timeline.unit.restore_workbook_probe_matches_canonical_fixture_4fbe5d13e2`, `module.timeline.unit.timeline_create_and_patch_decoding_accepts_manua_d80720f76e`, `module.timeline.unit.timeline_projection_contract_shaping_exposes_the_0a2232ebe0`, `module.timeline.unit.timeline_request_decoding_preserves_authoritativ_aa7608a640`, `module.timeline.unit.workbook_projection_contribution_owns_typed_runt_729aeb2b34` |
| `module.workbook` | `module.workbook.catalog.collection_review_fields_use_collection_value_v1_4efe5fcdae`, `module.workbook.catalog.dedicated_database_assertions_for_managed_servic_7d059c99ab`, `module.workbook.catalog.every_grid_write_is_anchored_to_the_route_bound_3926ec3df4`, `module.workbook.catalog.findings_investigative_queries_and_forensic_keyw_1ca6136d7c`, `module.workbook.catalog.keep_saved_clears_the_conflict_without_creating_c65a9c3099`, `module.workbook.catalog.linked_notes_create_contextual_artifact_links_9a09b2f951`, `module.workbook.catalog.locator_consistent_read_c33982976c`, `module.workbook.catalog.locator_public_contract`, `module.workbook.catalog.locator_validation`, `module.workbook.catalog.same_field_conflict_transport_uses_409_error_cod_4447ac22bf`, `module.workbook.catalog.service_free_contract_assertions_for_communicati_53927f649e`, `module.workbook.catalog.single_connection_rollback_assertions_for_manage_3c34c59fa8`, `module.workbook.catalog.text_compare_merge_treats_the_field_as_plain_tex_977f9eb56f`, `module.workbook.catalog.workbook_preference_routes_keep_home_sheet_ref_a_72ab108a31`, `module.workbook.integration.communications_log_handoff_status_review_and_les_94a6f7d0c2`, `module.workbook.integration.concurrent_edits_to_different_fields_succeed_wit_c3f74ce361`, `module.workbook.integration.coordination_collection_fields_round_trip_typed_9c30c3f538`, `module.workbook.integration.cursor_pagination_re_derives_live_authorization_2d2231384b`, `module.workbook.integration.evidence_row_create_and_patch_through_workbook_p_d074e288e1`, `module.workbook.integration.full_row_and_sparse_patch_wire_families_preserve_32a4ab3be1`, `module.workbook.integration.generic_record_patch_collection_fields_use_stabl_92304f55b9`, `module.workbook.integration.note_associations_public_contract`, `module.workbook.integration.notes_task_requests_and_decisions_mutations_thro_c325a978ef`, `module.workbook.integration.parties_and_coordination_system_views_query_rout_e0808abf39`, `module.workbook.integration.party_link_helper_field_writes_update_hidden_sam_014303b350`, `module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120`, `module.workbook.integration.required_surface_persistence_queries_through_wor_fc52555181`, `module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770`, `module.workbook.integration.startup_selection_fallback_behaves_correctly_acr_081544c048`, `module.workbook.integration.the_parties_system_view_is_discoverable_as_a_bas_7c96f60552`, `module.workbook.integration.timeline_clear_cells`, `module.workbook.support_integration.integration_notes_full_text_exact_search_ac185_ed48fdeae3`, `module.workbook.support_integration.managed_service_assertions_for_shared_harness_wo_ffc6e20715`, `module.workbook.support_integration.workbook_surface_create_query_and_patch_contract_ae5c34d894`, `module.workbook.support_unit.preference_response_projection`, `module.workbook.support_unit.startup_construction_failure_stops_workbook_rout_dcb8180f37`, `module.workbook.unit.note_association_query_admission`, `module.workbook.unit.openapi_record_mutation_contract_12a2ffaaf6`, `module.workbook.unit.registry_derived_writable_conflict_capabilities_c0614db52a`, `module.workbook.unit.the_owner_mutation_admission_enforces_closed_co_572104489b`, `module.workbook.unit.the_owner_mutation_admission_enforces_direct_re_b3fec12cf8`, `module.workbook.unit.timeline_grouping_permits_only_the_declared_whit_6db1021961`, `module.workbook.unit.timestamp_instant_v1_accepts_rfc_3339_json_strin_ba11c9c407`, `module.workbook.unit.workbook_telemetry_remains_bounded_content_safe_b88c7be5d3` |
| `platform.objectstore` | `platform.objectstore.support_integration.object_store_init_durable_effect_578bf55fdf` |
| `platform.viewschema` | `platform.viewschema.integration.view_schemas_discovery_http_036651fc36` |


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

### HTTP top-level tracker

| ID | Work item | Workstream | Status | Evidence / gate |
| --- | --- | --- | --- | --- |
| T-018 | Save accepted HTTP plan and implementation entry | planning | DONE | E-024; authorized scope and twelve-section plan recorded |
| T-019 | Characterize and freeze portfolio | WF-13 | DONE | E-025; V-015; saved checkpoint |
| T-020 | Correct responses | WF-14 | DONE | E-026; V-016; saved checkpoint |
| T-021 | Explicit-client request API | WF-15 | DONE | E-027; V-017; saved checkpoint |
| T-022 | Remove shared facades | WF-16 | DONE | E-028; V-018; saved checkpoint |
| T-023 | Remove local facades | WF-17 | DONE | E-029; V-019; saved checkpoint |
| T-024 | Qualify routing | WF-18 | DONE | E-030; V-020; saved checkpoint |
| T-025 | Final validation and handoff | WF-19 | DONE | E-031; V-021/022; saved checkpoint |


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

### HTTP checkpoint — WF-19 / E-031

2026-10-28, E-031: **the HTTP support effort is complete**. WF-19 finishes
last, after the saved E-030 checkpoint and final-source qualification. F-015–018
and D-019–023 are closed; C-012–016 are preserved or corrected as specified.
No production API/schema/policy source or adopted specification was changed.

#### Final-source qualification

All **592 required Go catalog rows across 29 owners** pass in **31 bounded
invocations**, with native outer exits 0. The dependency union has 589 retained
baseline rows, two new helper rows, and the additional production internal-error
boundary witness. The one excluded release-capacity row is explicitly justified
at E-026. Network Flow's two commands preserve its entire 62-row selection.

Each invocation uses `CARTULARY_OUTPUT_MODE=machine`,
`CARTULARY_TEST_RUN_ID=<run ID>` and `make test-slice OWNER=<owner> ROWS=<literal
selection>`. The exact literal selections are retained in each run-manifest's
`declared_inputs`, Section 8's baseline row table plus the named new/added rows,
and `/tmp/cartulary-tr02-http-ojbvn4pn/final-commands.json`. Runtime commands were
constructed from authored JSON catalog inputs, never from this document.
Run roots below are under `.cartulary/test-results/`.

| Run ID / retained root suffix | Owner | Required rows | Native outcome and audit |
| --- | --- | --- | --- |
| `tr02-http-final-app-operator` | `app.operator` | 2 | PASS / exit 0 / audit PASS |
| `tr02-http-final-app-server` | `app.server` | 19 | PASS / exit 0 / audit PASS |
| `tr02-http-final-error-boundary` | `app.server` | 1 | PASS / exit 0 / audit PASS |
| `tr02-http-final-harness-browser` | `harness.browser` | 13 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-artifacts` | `module.artifacts` | 14 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-assessments` | `module.assessments` | 11 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-auth` | `module.auth` | 36 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-collaboration` | `module.collaboration` | 18 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-entities` | `module.entities` | 22 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-evidence` | `module.evidence` | 40 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-imports` | `module.imports` | 21 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-incidentbundles` | `module.incidentbundles` | 25 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-incidents` | `module.incidents` | 32 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-indicators` | `module.indicators` | 28 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-jobapi` | `module.jobapi` | 5 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-links` | `module.links` | 7 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-networkflow-part1` | `module.networkflow` | 46 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-networkflow-part2` | `module.networkflow` | 16 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-parties` | `module.parties` | 10 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-projections` | `module.projections` | 12 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-recovery` | `module.recovery` | 23 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-reference_data` | `module.reference_data` | 20 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-reportcomposition` | `module.reportcomposition` | 4 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-reporting` | `module.reporting` | 28 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-revisions` | `module.revisions` | 42 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-savedviews` | `module.savedviews` | 14 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-tasksdecisions` | `module.tasksdecisions` | 8 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-timeline` | `module.timeline` | 29 | PASS / exit 0 / audit PASS |
| `tr02-http-final-module-workbook` | `module.workbook` | 44 | PASS / exit 0 / audit PASS |
| `tr02-http-final-platform-objectstore` | `platform.objectstore` | 1 | PASS / exit 0 / audit PASS |
| `tr02-http-final-platform-viewschema` | `platform.viewschema` | 1 | PASS / exit 0 / audit PASS |

The native exit/session ledger is retained in `final-outcomes.json` beside the
commands. Comparison of all manifest selections proves exact union membership:
592 unique required rows, none missing and none duplicated. Every final manifest
has source digest
`sha256:173e9b18994ba9829d3929d0c668c155d89d6fb37f07a9dc1e7bc3bb08285244`,
HEAD `157b9673b65101a80bada3ec749fe9c518279350`, and dirty source state reflecting
these uncommitted changes. Passing historical Incident/Auth or intermediate HTTP
runs were not substituted for final-source evidence.

V-020/021 also pass: source generated/schema/catalog/identity checks at the four
E-030 roots; final `make lint CHECK=go` and `CHECK=backend-boundary` at
`tr02-http-final-lint-go` and `tr02-http-final-lint-backend-boundary`. Finalizer
`tr02-http-wf19-finalize` passed before regression and reported zero generated
updates. Its retained maintenance was **skipped because RESULTS_DIR was unset**;
no successful compatible full warm-check root was supplied and no redundant full
check was run merely to obtain one.

All **38** exact graph roots (31 regression, two final lints, four source checks,
and finalizer) passed `make harness-observability-check RESULTS_DIR=<exact root>
CHECK=all`, each with native exit 0. Audit logs are retained beside
`final-outcomes.json`. These audits establish the scoped compatible evidence;
they do not claim complete mixed-runner owner acceptance, full check/CI, UI,
performance or release qualification.

#### Structural and production-assurance disposition

Request encoding/options live in request.go; response ownership/assertions live
in response.go, both in the existing httptestx package. Default clients are now
explicit. Generic application, Auth-flow and feature-local facades are gone;
no compatibility aliases, new runner, policy inventory or client registry remain.
The unused authenticated request shortcut, status-only response helper, fatal
body rewind, string reader, mandatory generic details and object-only generic
data assumptions are retired. Route object data is asserted explicitly where
needed. Terminal readers close before status/JSON/envelope failures, preserve
raw bytes and header access, and expose read/close failures without replaying or
printing full response bodies.

Seven new service-free helper tests in two canonical rows exercise malformed and
trailing JSON, required envelope members, optional/additive fields, raw/binary and
204 responses, failure cleanup, request bytes/options, explicit TLS trust and
redirect behavior. Real application assembly tests and all existing product
claim identities remain. Process TLS/origin binding, enterprise form/redirect
adapters, cookie policy, semantic domain setup, literal authorization outcomes,
replay, persistence and unique state observations are preserved. Production
internal-error sanitization remains asserted at its production owner, including
required empty details and private-diagnostic non-disclosure.

#### Failure history and preservation

Intermediate failures are retained, not rewritten: E-026 records the initial
unsorted-selector admission/generation failures and their repair; E-028 records
the missed authflow alias causing lint/process compilation failure and its
qualified repair. The oversized Network Flow ROWS input was rejected before
execution with usage_error and created no run root; bounded commands subsequently
qualified the complete selection. No unresolved product, compilation, cleanup,
source, or evidence failure remains.

Baseline comparison accounts for **125 modified existing paths and four new Go
files** (request.go, request_test.go, response.go, response_test.go). Every changed
path is test code/support, this tracker, the authored harness family manifest or
the Make-generated execution-topology render index. No production policy source,
dependency lockfile or unrelated content changed. HEAD and staged contents are
unchanged; Git index bytes changed during Git inspection/refresh, so byte-for-byte
index preservation is explicitly **not** claimed. Ignored build/cache/service
outputs are outside the comparison. After source finalization, only this tracker
changed. Original Incident/Auth text in all twelve section bodies is preserved
verbatim; the document title now identifies the current HTTP effort.

Inspection depth remains bounded: helper bodies and selected ownership/client
consumers were directly inspected; the remaining changed calls were mechanically
migrated, statically checked and regression-tested, not subjected to a manual
whole-test-body audit. This does not imply whole-repository specification
completeness or human adoption review.

#### Document and handoff completion

V-022: `CARTULARY_TEST_RUN_ID=tr02-http-final-markdown make lint-markdown`
exited 0 with a passing tool summary at
`.cartulary/test-results/tr02-http-final-markdown/adhoc/lint-markdown/tool-run-summary.json`.
Its configured globs do not include this tracker. Direct review separately
confirms twelve ordered sections, valid local references, stable IDs, acyclic
WF-13→19 dependencies, historical preservation and source/evidence separation.
`git diff --check` passes. Every predecessor checkpoint was saved and reviewed
before its successor began. There are no remaining implementation blockers,
required production corrections or scoped validation obligations. No commit,
staging or push was performed. The final saved checkpoint is rechecked before
handoff; no continuing background monitor is implied.

WF-19/S-019 is DONE. This checkpoint is saved and reviewed before
final handoff completion.

### HTTP final validation progress record — E-031 (superseded)

Final qualification refinement: the production-owner witness
`app.server.unit.internal_error_boundary` is added to the 591-row dependency union
for **592 required rows**. Direct inspection of `internal_error_test.go` confirms
it requires the complete safe 500 payload, empty details, no conflict, correlation
and no private diagnostic leakage. This distinct guarantee stays at its owner;
the generic helper must not make details mandatory for every code. No source or
routing changes are required. Earlier 591-row counts describe the dependency
union before this qualification addition.

The original Network Flow 62-row command was rejected before execution because
ROWS exceeded the public 4,096-character bound; no run root was created. Its exact
selection is split into part1 (46 rows) and part2 (16 rows), with no dropped or
duplicated rows. This command-admission correction does not change the source or
weaken a test. The final commands file records the bounded invocations.

WF-19 entered after the saved E-030 checkpoint. Finalization
`CARTULARY_TEST_RUN_ID=tr02-http-wf19-finalize make agent-finalize` passed and
reported zero generated updates. RESULTS_DIR was unset: retained-run maintenance
was explicitly skipped, not silently qualified. Final-source Go and backend
boundary lints pass at `tr02-http-final-lint-go` and
`tr02-http-final-lint-backend-boundary` under `.cartulary/test-results/`.

Final source digest is
`sha256:173e9b18994ba9829d3929d0c668c155d89d6fb37f07a9dc1e7bc3bb08285244`.
The 591-row exact portfolio is executing under `tr02-http-final-<owner-with-dashes>`
roots. Exact native commands are retained in
`/tmp/cartulary-tr02-http-ojbvn4pn/final-commands.json`; source fingerprints are in
`final-source.json`. No source edits are allowed during qualification; a repair
requires re-evaluating compatibility and recording separate evidence. Completion
is not yet claimed. The final completion checkpoint supersedes this progress entry.

### HTTP checkpoint — WF-18 / E-030

2026-10-28, E-030: V-020 is complete. The two new service-free helper rows retain
harness.browser support-contract ownership, no fixture dependencies, and exact
unique selectors. Existing helper/assembly and all product row IDs remain intact;
no row migration or production verification-contract edit is needed.

`CARTULARY_TEST_RUN_ID=tr02-http-wf18-generate make generate` exited 0. Reviewed
tracked generated diff is limited to the execution-topology render index input
digest and authored harness family hash. No generated file was hand-edited.

All four canonical checks exited 0 with passing summaries. Each command used
`CARTULARY_OUTPUT_MODE=machine` and its explicit run ID below:

- `tr02-http-wf18-generated`: `make source-check CHECK=generated`
- `tr02-http-wf18-schema`: `make source-check CHECK=schema`
- `tr02-http-wf18-catalog`: `make source-check CHECK=catalog`
- `tr02-http-wf18-identity`: `make source-check CHECK=identity`

Exact retained roots are `.cartulary/test-results/<run ID>`. Current routing has
no overlapping/orphaned selectors; new test names were authored in their introducing
slices. Final portfolio is 589 baseline rows plus the two new witnesses = 591,
across 29 owners, in `final-portfolio.json`. The one release-capacity exclusion
and its direct-source rationale remain recorded at E-026. Finalization, final
regression, evidence audits and handoff completion remain WF-19.

WF-18/S-018 is DONE. This checkpoint is saved and reviewed before
WF-19 begins.

### HTTP checkpoint — WF-17 / E-029

2026-10-28, E-029: generic Auth, Timeline, Entities, Evidence, Reference Data
and process option/envelope/request wrappers are removed with their callers.
The additionally inspected Timeline requireSuccessEnvelopeWithBody was another
status-only forwarding layer and is retired under F-015. Enterprise form and
non-following redirect adapters, process origin/trusted client binding, and
semantic recovery/domain fixtures remain. No generic aliases were added.

V-019 passes: native exits 0 for the focused commands below. Each retained root
is `.cartulary/test-results/<CARTULARY_TEST_RUN_ID>`.

- `CARTULARY_TEST_RUN_ID=tr02-http-wf17-auth CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=module.auth ROWS=module.auth.integration.deployment_admin_binding_create_replay_divergent_85cb20830a,module.auth.integration.enterprise_provider_discovery_proves_sorted_exac_059973abc3,module.auth.integration.saml_sp_initiated_begin_acs_verifies_determinist_3b5547023b`
- `CARTULARY_TEST_RUN_ID=tr02-http-wf17-timeline CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=module.timeline ROWS=module.timeline.integration.create_patch_mark_reviewed_supersede_query_and_w_90276912d7,module.timeline.integration.real_http_create_patch_review_supersede_replay_m_37e5c78ca1`
- `CARTULARY_TEST_RUN_ID=tr02-http-wf17-entities CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=module.entities ROWS=module.entities.integration.the_mention_resolve_route_persists_durable_resol_d68c1befd6`
- `CARTULARY_TEST_RUN_ID=tr02-http-wf17-evidence CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=module.evidence ROWS=module.evidence.integration.the_object_blob_create_route_enforces_request_sh_17b6e1b674`
- `CARTULARY_TEST_RUN_ID=tr02-http-wf17-reference CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.reference_pack_import_uses_the_shared_upload_env_95574f69b8`
- `CARTULARY_TEST_RUN_ID=tr02-http-wf17-go CARTULARY_OUTPUT_MODE=machine make lint CHECK=go`

Canonical formatting passed at tr02-http-wf17-format. The remaining doJSON is
the intentional process origin/client adapter, not a default-client copy. Search
and direct inspection find no retired generic exports or their alias callers.
Other NewJSONRequest/options uses are individual tests with meaningful request
customization. Diff whitespace review passes. No production changes or defects
were established. Final-source regression remains WF-19.

WF-17/S-017 is DONE. This checkpoint is saved and reviewed before
WF-18 begins.

### HTTP checkpoint — WF-16 / E-028

2026-10-28, E-028: generic application-support and Auth-flow request/options
exports are deleted, their callers use httptestx directly, and the redundant
error alias is removed. RequireSuccessData now lives with response assertions;
it asserts an object only for routes asking for that shape and has positive and
negative ownership witnesses. Semantic auth/fixture operations remain in their
owners; process origin plus trusted-client binding is retained. The final alias
review includes imports named authflow/authflowtest, not only flowtest.

V-018 focused checks passed (native exit 0), with exact roots below under
`.cartulary/test-results/`. Commands use `CARTULARY_OUTPUT_MODE=machine` and
`CARTULARY_TEST_RUN_ID=<root suffix>`:

- `tr02-http-wf16-unit`: `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.http_response_readers_own_cleanup_and_enforce_re_d6b5f11be6`
- `tr02-http-wf16-auth`: `make test-slice OWNER=module.auth ROWS=module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af`
- `tr02-http-wf16-entities`: `make test-slice OWNER=module.entities ROWS=module.entities.integration.candidate_discovery`
- `tr02-http-wf16-process-fixed`: `make test-slice OWNER=app.server ROWS=app.server.integration.application_fixtures_own_independent_databases_a_04455c6902,app.server.process.process_smoke_exercises_login_csrf_session_revoc_c6fd6eb90f`
- `tr02-http-wf16-go-fixed`: `make lint CHECK=go`

Initial lint `tr02-http-wf16-go` failed vet/staticcheck and initial
`tr02-http-wf16-process` failed process preflight compilation because Evidence's
verification fixture still called the removed Auth facade through alias authflow.
The exact retained logs establish the compile failure, not a product assertion
failure or infrastructure outage. The two calls now preserve their client/context
through httptestx directly; the unused Auth import is removed. Both failed
obligations passed after this repair. Earlier focused successes remain slice
observations; final-source compatibility is re-established in WF-19, not inferred.
Format and diff whitespace checks pass. No production source changed.

WF-16/S-016 is DONE. This checkpoint is saved and reviewed before
WF-17 begins.

### HTTP checkpoint — WF-15 / E-027

2026-10-28, E-027: request primitives now live in request.go within the existing
httptestx package. Both JSON dispatch helpers take the caller's client explicitly;
471 direct calls in 49 files now pass their prior default client explicitly.
The unused authenticated-request shortcut is removed. Nil versus explicit JSON
null, raw malformed/duplicate-member bytes, option order, cookies and headers are
preserved. No global client mutation or compatibility overload was introduced.

V-017 passed, all native command exits 0 and summaries pass:

- `CARTULARY_TEST_RUN_ID=tr02-http-wf15-unit CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=harness.browser ROWS=harness.browser.unit.http_request_helpers_preserve_encoding_options_a_e755d11c07,harness.browser.unit.http_response_readers_own_cleanup_and_enforce_re_d6b5f11be6`
- `CARTULARY_TEST_RUN_ID=tr02-http-wf15-process CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=app.server ROWS=app.server.process.a_fresh_process_reaches_health_and_ready_only_af_c93818219a`
- `CARTULARY_TEST_RUN_ID=tr02-http-wf15-go CARTULARY_OUTPUT_MODE=machine make lint CHECK=go`

Retained roots use these exact run IDs under `.cartulary/test-results/`. The new
request witness proves supplied TLS trust and non-following redirect policy for
both helpers and the raw dispatcher; the process row preserves HTTPS/readiness
admission. Go format/vet/staticcheck qualify the changed call signatures across
authored packages. Make generation and formatting also passed at explicit roots
`tr02-http-wf15-generate` and `tr02-http-wf15-format`. Diff whitespace review passes.
No production changes or current production defect were identified.

WF-15/S-015 is DONE. This checkpoint is saved and reviewed before
WF-16 begins.

### HTTP checkpoint — WF-14 / E-026

2026-10-28, E-026: terminal raw/JSON/no-content readers now close before
reporting failures, preserve read/close errors, and never restore consumed bodies.
Shared envelopes verify required message/status/correlation and permit optional
details/additive members and route-owned data shapes. All former shared status
and string readers migrated; the distinct process readiness method remains.
Evidence bytes and disposition headers, Incident correlation, concurrent 200/201
creation and Network Flow empty retirement responses retain their observations.
Five new service-free tests are routed under
`harness.browser.unit.http_response_readers_own_cleanup_and_enforce_re_d6b5f11be6`.

V-016 passed: the new/existing helper rows and 27 focused consumer rows all exited
0 with passing summaries. Exact commands use `CARTULARY_OUTPUT_MODE=machine`,
`CARTULARY_TEST_RUN_ID=<root suffix>` and the selections below. Retained roots are
under `.cartulary/test-results/`:

- `tr02-http-wf14-unit`: `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.http_response_readers_own_cleanup_and_enforce_re_d6b5f11be6,harness.browser.unit.httptestx_helper_contract`
- `tr02-http-wf14-module-savedviews`: `make test-slice OWNER=module.savedviews ROWS=module.savedviews.integration.saved_view_create_update_duplicate_and_delete_pe_5f93d857d3,module.savedviews.store.managed_service_assertions_for_ordinary_saved_vi_fc464a05e5,module.savedviews.store.ordinary_saved_view_patch_enforces_the_mutable_f_9554a5d90f,module.savedviews.store.service_free_contract_assertions_for_ordinary_sa_27bac8bf90,module.savedviews.unit.saved_view_scope_vocabulary_is_exactly_private_s_aa8991ad84`
- `tr02-http-wf14-module-networkflow`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.integration.effective_resource_limit_discovery,module.networkflow.integration.network_flow_routes_remain_unclaimed_by_default_e8dabf0ee1,module.networkflow.integration.pagination_recovery,module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5,module.networkflow.integration.saved_graph_cutover_v6,module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.time_bucket_saved_graph_lifecycle`
- `tr02-http-wf14-module-incidents`: `make test-slice OWNER=module.incidents ROWS=module.incidents.integration.incident_create_persists_the_incident_row_bootst_0bd8d02f12,module.incidents.integration.live_membership_patch_to_the_current_role_return_7a14cba87a,module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6,module.incidents.integration.real_incident_patch_persists_only_patchable_meta_9fc8fc207f,module.incidents.integration.reserved_but_unclaimed_family_roots_and_descenda_3cbf9bf6e6,module.incidents.support_integration.membership_create_replay_returns_original_and_di_98e85ad975`
- `tr02-http-wf14-app-server`: `make test-slice OWNER=app.server ROWS=app.server.process.process_smoke_exercises_incident_creation_valida_541e2263d1`
- `tr02-http-wf14-module-evidence`: `make test-slice OWNER=module.evidence ROWS=module.evidence.integration.attach_route_validation_authorization_route_owne_7567e19cd9,module.evidence.integration.expired_slot_replay_returns_the_same_expired_slo_6d29ac9f3a,module.evidence.integration.initial_blob_create_atomic_replay_concealment_73c81fe329,module.evidence.integration.the_object_blob_create_route_enforces_request_sh_17b6e1b674,module.evidence.store.download_responses_use_authoritative_metadata_fo_a55dbb9aac,module.evidence.store.preview_handle_and_download_handle_issuance_rout_e9c8e5a3cd,module.evidence.unit.evidence_open_apiincludes_route_family_866f0d0acf`

Initial `make format` was rejected before execution because the newly authored
selector names were not ASCII-sorted. `make generate` likewise failed at
`.cartulary/test-results/20261028T132247Z-p30372`. Public task-guide diagnosed the
exact authored-row defect; sorting its names repaired it. Subsequent generation
and formatting passed (`tr02-http-wf14-generate`, `tr02-http-wf14-format`). No
production defect was found. `git diff --check` passes.

Final portfolio refinement: exclude
`module.networkflow.measurement.backend_graph_capacity_certification` from the
conservative 590-row inventory: its inspected graph-capacity source does not call
changed HTTP support and release capacity qualification is outside scope. The
589-row base is retained in `final-portfolio.json`, extended by new helper rows.
This is a human scope disposition, not a new executable exemption or selector.

WF-14/S-014 is DONE. This checkpoint is saved and reviewed before
WF-15 begins.

### HTTP checkpoint — WF-13 / E-025

2026-10-28, E-025: current-source representative characterization qualifies.
`make task-guide OWNER=<owner> DETAIL=coverage JSON=1` succeeded for all 29
potentially affected owners; exact JSON outputs are retained beside portfolio.json.
Both canonical baseline runs exited 0 with status pass:

- `CARTULARY_TEST_RUN_ID=tr02-http-wf13-unit CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=harness.browser ROWS=harness.browser.unit.httptestx_helper_contract`
- `CARTULARY_TEST_RUN_ID=tr02-http-wf13-assembly CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=app.server ROWS=app.server.integration.http_harness_restart_storage`

Exact retained roots are `.cartulary/test-results/tr02-http-wf13-unit` and
`.cartulary/test-results/tr02-http-wf13-assembly`. The first qualifies existing
route-mode/cookie helper witnesses; the second qualifies real Postgres/object-store
assembly and envelope behavior. This is representative characterization, not a
590-row baseline pass. Final regression remains due after finalization.

Relevant response consumers were inspected for body reuse: Evidence redemption
must return raw bytes; disposition tests retain headers; concurrent exact creates
accept only 200/201 before consuming once; Incident 204 checks retain correlation
and empty-body requirements. No normative contradiction or production defect is
established. V-015 is satisfied and RB-005 is closed. No production source changed.

WF-13/S-013 is DONE. This checkpoint is saved and reviewed before
WF-14 begins.

### HTTP implementation entry — E-024

2026-10-28: clean HEAD `157b9673b65101a80bada3ec749fe9c518279350` revalidated.
Root/backend/tooling instructions, test-operations skill/recipes, relevant Core 01
and adopted Testing Harness clauses inspected. Public Make help discovered.
Baseline content/index/identity snapshot retained at the Section 1 location.
This accepted plan is saved before implementation. No new test success claimed.

| Boundary | Evidence / current disposition |
| --- | --- |
| Scope and authority | Explicit implementation request authorizes WF-13–19; historical document-only restriction superseded |
| Backend | Existing HTTP support owner retained; request/response implementation and downstream test calls only |
| Frontend | Not applicable: no frontend changes or rendered UI work |
| Contract/codegen | Authored catalog rows change only for actual helper witnesses; generated outputs through Make |
| Tests/harness | Conservative 40-package/590-row inventory retained; current qualification pending |
| Security | Preserve credential policy, custom clients, literal authorization outcomes and safe error boundaries |
| Risks/next action | RB-005: qualify representative current helper/assembly rows, save WF-13 checkpoint |

| Evidence | Source / baseline | Observation and limits |
| --- | --- | --- |
| E-024 | Clean HEAD above; direct helper/owner inspection; `portfolio.json`; `baseline.json`; Make help | Planning/source evidence only; no execution pass inferred from old Auth results |


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

### HTTP blockers and readiness

| ID | Blocker | Status | Resolution gate |
| --- | --- | --- | --- |
| RB-005 | Current HTTP characterization qualified at E-025 | DONE | WF-13 / V-015 passing evidence and saved checkpoint |

No ownership contradiction or unresolved blocker remains. All predecessor
qualifications and checkpoints are complete at E-025–031. No production defect
requiring a separate correction workstream was established.


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

### HTTP completion criteria

| Outcome | Current statement |
| --- | --- |
| Planning | Complete for the bounded HTTP support iteration |
| Authorization | Implementation granted by current request |
| Readiness | S-013–019 qualified; every predecessor checkpoint saved before successor |
| Implementation | WF-13–19 DONE; F-015–018 and D-019–023 closed at E-031 |
| Preservation | E-031: authorized content changes only; HEAD/staged contents unchanged; index bytes refreshed; historical section bodies preserved |

- [x] WF-13 qualifies characterization and freezes the exact portfolio.
- [x] WF-14 closes F-016/017 with terminal ownership and accurate envelope checks.
- [x] WF-15–17 close F-015; no generic compatibility aliases remain.
- [x] F-018 has isolated failure witnesses and retained application assurance.
- [x] WF-18 qualifies authored routing and generated projections.
- [x] Every workstream has a saved reviewed checkpoint before its successor.
- [x] WF-19 finalizes before final regression, qualifies compatible evidence and
  records validation, limitations, preservation and handoff last.


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
