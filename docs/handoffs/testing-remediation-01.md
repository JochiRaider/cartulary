# Testing Remediation 01: Frontend Test-Policy Simplification

## 1. Scope and Source Posture

Document date: 2026-10-26. **Current action: completed implementation and handoff.** The user
approved the revised WF-06–WF-11 plan and instructed its implementation after
selecting semantic readiness, simplified E2E owner rules, and retention of the
existing bounded Fallow gate. This supersedes the earlier document-only scope.
WF-06–WF-11 are DONE, including final validation and handoff. Historical WF-00–WF-05 remain complete with their original
qualification limits in Section 10. The preceding
[remediation handoff](testing-remediation.md) and its separate open WF-33 effort
are unchanged.

Repository: `/home/jochi/code/cartulary`; supplied and established canonical target
`.`; kind: repository directory; label: `frontend-test-policy-assurance`. This
continues the existing tracker's repository identity with a newly bounded seam.
The controlling plan and handoff is `docs/handoffs/testing-remediation-01.md`.
Authorized edits also include the specification, implementation, tests, authored
routing and Make-generated projections listed below. The output is a regular, singly linked file; neither it
nor its parent components is symlinked. Artifact mode: **normal**. Planning
completeness, workstream completion and implementation authorization are separate
outcomes; see Section 12.

The [modular refactor framework](cartulary_modular_refactor_planning_framework.md)
and Refactor Tracker format supply planning structure. They do not make current
test arrangements requirements. This iteration covers the two bespoke frontend
policy scanners, their useful successor assurance, E2E transport ownership,
Fallow E2E roots, and directly affected routing and support guidance. It does not
redesign production behavior, browser workflows, service fixtures, scheduling,
runner protocols, security thresholds, visual goldens or performance policy.

| Authority | Applicable boundary and disposition |
| --- | --- |
| [Core 00](../spec/00_document_set_status_and_precedence.md), §§2–5 | Adopted primary-owner precedence and the documentation-free executable boundary. Tests, typed projections and this handoff cannot define product requirements. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md), adopted/current | REQ-001/003/004: scope and projections; REQ-070: existing Fallow policy; REQ-075/082: discovery and canonical assurance; REQ-103: semantic frontend readiness; REQ-665: stable post-mutation identity and observation-only continuity; REQ-676/677: E2E discovery/reachability and private raw transport; Sections 3, 8–10: exact selection, evidence and execution. |
| Core 01–04 and adopted product owners | Production HTTP, authorization, storage, history, collaboration and UI behavior remain outside the change. Core 03 REQ-03-283 remains the product owner for row-local focus restoration. |
| [Domain vocabulary](../domain.md) and [frontend testing guide](../guides/cartulary_frontend_implementation_testing_guide.md), §7 | Owner navigation and implementation-support boundaries; the guide is not a second requirements authority. |
| [Core 05](../spec/05_claim_publication_and_benchmark_reproducibility.md) | Not applicable: no claim-bearing timed or fixture-sensitive publication is proposed. |

Executable tests, generators, runtime metadata, conformance and release evidence
must never read, stat, hash or otherwise depend on this document or other Markdown.
Human inspection and documentation lint are not product evidence consumers.

The following table preserves the **prior document-only baseline** (E-010–E-017).
The implementation baseline, including the pre-existing tracker changes, is
E-018; implementation preservation and final evidence are recorded in E-029.

| Historical baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Checkout | `main`; HEAD `971d6423c1bfc5f1f3ac6569fc524b596cf33f61` | E-017 records the final comparison | No commit, push or PR requested |
| Index | Clean; entry fingerprint `005bc70a5d18167d1c962ef3029fb6530fab6a16d8746efb8e367b54ddf104f2` | Must remain identical | Optional Git writes disabled |
| Tracked/non-ignored content | 7,638 paths; combined fingerprint `7ed6019c9a137bac0f549c20351d540cf11fbac8c0775d09e82945df30d91ad3`; no unstaged or non-ignored untracked changes | Excluding this output, fingerprint must remain `a7df1bc81868dab69454c7e23a14d67e0ced15e124fc1cab6dfb76033db16215` | Per-path hashes and file identities retained outside repository |
| Output | Inode `348826`, one link; initial SHA-256 `70dc2e851a16e6c07ed52e483ccbc853b5cb15dc9589396fe271bddcca3058b0` | Identity and contents rechecked immediately before replacement | Historical canonical target `.` established from existing text |
| Comparison material | `/tmp/cartulary-tr01-doc-hla6fi0g/baseline.json` and `prior-handoff.md` | E-017 | Temporary local resumption aids, not executable evidence inputs |
| Ignored state | Installed dependencies, caches, builds and retained results excluded from refactor inputs | Documentation validation must use installed dependencies and external temporary outputs, or be skipped with its reason | Ignored trees are not exhaustively content-audited; no broad preservation claim for concurrent external activity |

Those document-only fingerprints matched the preceding planning turn. They do
not qualify implementation results. The subsequent explicit implementation
instruction authorizes WF-06–WF-11; E-018 preserves its distinct starting state.

## 2. Inspected Baseline Repository Inventory

The enumeration universe is tracked and non-ignored untracked paths under `.`.
Every path is metadata-only for this bounded analysis unless explicitly identified
below as directly inspected. Generated, binary and vendored entries remain
metadata-only; ignored dependency installs, service data and result directories
are excluded because they are not authored refactor inputs. No submodule or
external symlink target is traversed as source. Enumeration is not behavioral
inspection. Exploratory Go fixture and remaining shell-suite reads did not select
those subsystems for change.

| Path / kind | Disposition and inspection depth | Responsibility, consumers and dependencies | Owner / contracts / risk / evidence |
| --- | --- | --- | --- |
| `apps/web/e2e/support/runtime/architecturePolicy.test.ts` | Directly inspected: complete 655-line authored scanner and nine assertion groups | Own source walker, TypeScript import/reachability/export analyzer, literal and naming restrictions; one `harness.browser` row, project `harness-node` | Split across existing owners; C-007/C-009/C-010; loss of unique checks if simply deleted; E-012 |
| `apps/web/src/testing/selectorContractPolicy.node.test.ts` | Directly inspected: complete 671-line authored scanner and five assertion groups | Source walker, selector-prefix inventory, mock exceptions, regex and AST scans, prescribed heading grammar; five `web.architecture` rows, project `node-unit` | UI Contracts and behavioral support; C-008/C-010; REQ-103 amendment needed before retiring grammar; E-013 |
| `tools/harness/static-analysis/frontend-import-boundary-check.mjs` and `tests/test-frontend-policy.mjs` | Directly inspected at parser, resolver, evaluator, policy fixture and result boundaries | Existing explicit-root checker with one source enumeration and TypeScript parsing; real CLI and canonical assurance consumers | Frontend static analysis; C-007/C-009; retain independent positive/negative expectations; E-012 |
| `tools/frontend_import_boundaries.json` | Directly inspected relevant authored rules and scan scopes | Already checks grid/vendor, generated artifacts, package facades and E2E/private-app imports; scanner-specific core-support exception | Authored projection, not requirements authority; C-007/C-010; E-012/E-015 |
| `.fallowrc.json`, `tools/fallow/reachability_owner.json` | Directly inspected configuration and relevant owner fields | Blanket `apps/web/e2e/**/*.ts` root hides helper orphaning; existing Vitest entry patterns omit E2E support tests | Fallow; C-007; root correction precedes analyzer deletion; E-012 |
| `tools/harness/static-analysis/fallow-reachability.mjs`, `tests/fallow-fixtures.mjs`, `tests/test-fallow-integration.mjs` | Directly inspected root construction and real analyzer fixtures/assertions | Effective config merges authored roots; fixtures cover live/dead peers, but not the required E2E root distinction | Fallow; C-007/C-010; pinned analyzer qualification is future work; E-012 |
| `apps/web/vite.config.ts`, `apps/web/playwright.config.ts`, `apps/web/playwright.shared.config.ts` | Directly inspected test discovery and setup bindings | `harness-node` owns E2E `*.test.ts`; Playwright owns specs and explicit global setup/teardown; no config behavior change planned | Existing runtime-discovery owners; C-007/C-010; E-012 |
| `apps/web/e2e/support/transport/publicJsonClient.ts`, `publicHttpOperationClient.ts`, `publicJsonClient.test.ts` | Directly inspected complete execution and test boundaries; import search across E2E | Origin/type consumers span E2E; `requestPublicJson` is used only by direct transport tests; the operation client uses `requestPublicJsonObserved` | Transport; C-009; guard names the wrong seam; E-014 |
| `apps/web/src/testing/timelineWorkbookTestSupport.ts`, `timelineWorkbookTestSupport.test.tsx`, `workbookInspectorTestSupport.ts`, `workbookInspectorTestSupport.test.tsx` | Directly inspected row/inspector waits and existing assertions | Finite configured waits; stable row IDs and exact inspector view/record/version; existing delayed hydration and diagnostic tests | Application test support; C-008; add negative identity and observation witnesses; E-014 |
| `packages/ui-contracts/src/selector-core.node.test.ts`, `application-selectors.node.test.ts` | Directly inspected selector-identity, escaping, registry and authentication assertions | Public package facade tests already cover deterministic stable identities and malformed inputs | `package.ui`; C-008; preserve existing runtime selector output; E-014 |
| `tools/test_families/{web.architecture,harness.browser,package.ui}.json`, `contracts/verification/owners/web.architecture.json`, `tools/frontend_source_ownership.json` | Directly inspected affected rows, verification kinds and path membership | Six scanner rows; existing frontend/Fallow operations; readiness, transport and selector successors; authored source ownership | Routing owners; C-010; atomic source/selector/projection migration; E-015 |
| Makefile, generated task-surface includes, generated-artifact policy, documentation lint and readiness wrappers | Directly inspected relevant invocation paths and mutation boundaries | Canonical commands, installed-only admission, external result/cache outputs, generated ownership | Harness plumbing; C-010; documentation-only invocation boundary; E-015/E-016 |
| Remaining enumerated paths | Metadata-only; not proposed for movement | Not a claim of assertion completeness or exhaustive callers | Product/backend/browser workflow audit deferred outside this seam |

The primary inventory is two scanners, fourteen top-level assertion groups and six
catalog rows. These counts identify the inspected baseline; they are not future
acceptance constraints. Supporting files are inspected only to the depth stated.
Direct import and routing searches establish the known consumers, not the absence
of dynamic external consumers. E-018/E-019 revalidated this baseline. WF-10 also
inspected and updated the graph assurance tier sentinel discovered during cutover.

## 3. Module Boundary Diagnosis

| Responsibility found | Baseline location | Required owner / disposition | Evidence / finding / contract |
| --- | --- | --- | --- |
| Facades, private imports and vendor boundaries | E2E scanner plus existing frontend checker | **Keep** existing checker; **move** useful residual checks into its existing source pass | E-012; F-007; C-007 |
| Reachability and unused exports/types | Bespoke E2E analyzer plus Fallow | **Move** E2E assurance to Fallow after root and real-analyzer qualification | E-012; F-007/F-008; C-007 |
| Selector identity and encoding | Selector scanner plus UI Contracts tests | **Keep** package behavioral assertions; retire duplicate syntax/family inventories | E-013/E-014; F-009; C-008 |
| Bounded readiness and observation-only continuity | Source-expression grammar plus actual helpers | **Keep/extend** helper behavioral tests; separately amend incidental grammar | E-013/E-014; F-009; C-008 |
| Raw versus validated HTTP execution | Symbol-specific scanner guard and mixed transport file | **Split** raw execution from the reusable context/origin boundary; enforce actual module ownership | E-014; F-010; C-009 |
| Catalog ownership and execution | Browser/architecture scanner rows and existing assurance operations | **Move/consolidate** obligations into existing owners; retire six redundant rows when qualified | E-015; F-011; C-010 |
| Production HTTP, WebSocket, storage, revision and UI semantics | Existing production owners | **Keep**; not proposed for movement or correction | C-006; no production slice |

No new public production API, runner, schema, registry, scheduler, compatibility
mode or general fixture language is needed. Private support exports can be
removed or narrowed; forwarding wrappers are not a completion strategy.

## 4. Public Contract and Behavior Freeze Map

Historical C-001–C-005 remain reserved for the completed graph/scanner iteration.

| ID / surface | Observed behavior and owner | Required disposition / successor | Prerequisite / risk / slice / validation |
| --- | --- | --- | --- |
| C-006 — production behavior | Core 01–04 and applicable adopted subsystem owners retain HTTP, authorization, storage/history, collaboration and UI requirements; this iteration changes support only | Preserve routes, wire shapes, authentication, persistence, revision semantics, DOM identity and actual focus/accessibility behavior | No product correction or product-data migration; all slices must remain within support scope |
| C-007 — architectural integrity | Existing checker enforces meaningful import boundaries; E2E scanner adds its own graph and source-shape rules; REQ-070/082 govern analyzer and consolidated assurance mechanics | Preserve facade/vendor separation, route-neutral shared helpers and meaningful reachability; retire closed-root/catch-all naming and parser-specific restrictions | F-007/F-008; S-007/S-008/S-010; V-009/V-010/V-012 |
| C-008 — selectors/readiness | Amended REQ-103 requires finite waits, stable identity and observation-only readiness; REQ-665 requires committed identity/version convergence | Preserve semantic guarantees and runtime selector values; the authorized amendment removes matcher grammar before scanner retirement | RB-002; F-009; S-006/S-009; V-008/V-011 |
| C-009 — support HTTP transport | Origin binding, body/header omission, non-2xx results, generated method/path/query, success validation, observed headers and private diagnostics; guide §7 and Harness evidence integrity apply | Preserve all meaningful operation/transport assertions; replace ineffective wrapper-name guard with module ownership | RB-001; F-010; S-007; V-009 |
| C-010 — execution/evidence | REQ-003/004/075/082, Section 3 and Sections 8–10 govern authored routing, exact selectors, canonical operations and truthful publication | Preserve exact resolution and distinct contexts; remove obsolete rows with explicit successor/retirement mapping; no historical-ID aliases | F-011; S-010/S-011; V-012/V-013 |

Source literals, helper filenames, test titles, directory placement, exact test
populations and the old scanners' parser limitations are not production
requirements. Actual requirements and necessary safety/evidence invariants need
successor assurance. Do not invent normative text to preserve incidental test
arrangements. Test presence is not execution success, and no result proves owner
projection fidelity without human review.

## 5. Coupling and Boundary Findings

Findings describe the inspected baseline. Their completed remediation and
validation are recorded in E-019–E-029; the risk column below describes what
retaining that baseline would have left unresolved.

| ID / classification | Finding and evidence | Required action / scoped acceptance | Contract / slice / validation |
| --- | --- | --- | --- |
| F-007 / must_fix | E-012: the E2E test duplicates graph traversal, import/export interpretation and existing boundary enforcement | One established owner per meaningful rule; no second analyzer or new generic policy engine | C-007; S-007/S-008/S-010; V-009/V-010/V-012 |
| F-008 / must_fix | E-012: every E2E TypeScript file is currently a Fallow root, so existing Fallow coverage is not equivalent to orphan detection | Admit true test/setup roots, prove imported helpers live and dead peers dead using the real pinned analyzer before deletion | C-007; S-008; V-010 |
| F-009 / must_fix | E-013/E-014: selector prefixes, mock allowlists and heading-expression grammar duplicate identity knowledge and constrain legitimate assertions | Adopt semantic REQ-103 replacement; qualify readiness failures and unchanged focus directly; remove syntax inventories | C-008; S-006/S-009; V-008/V-011 |
| F-010 / must_fix | E-014: name guard bans test-only `requestPublicJson`, not actual raw execution through `requestPublicJsonObserved` | Establish a transport-private execution module and one validator path; remove the unused wrapper without a shim | C-009; S-007; V-009 |
| F-011 / must_fix | E-015: six scanner rows and a scanner-specific boundary exception preserve redundant work and split ownership | Migrate authored routing, source ownership and generated projections atomically; preserve successor qualification contexts | C-010; S-010/S-011; V-012/V-013 |
| F-012 / intentional/no_action | Existing real checker/analyzer, selector-package, transport and helper tests observe different boundaries | Keep those distinct witnesses; matching topic names do not make them duplicate coverage | C-007/C-008/C-009; no separate removal slice; V-009–V-011 |
| F-013 / must_fix | Revised planning found that REQ-676 owns the closed E2E layout/importer accounting and REQ-677 names the wrapper slated for removal | Amend these owners before dependent implementation; retain semantic discovery and validated transport authority | C-007/C-009; S-006; V-008 |
| F-014 / must_fix | REQ-070/help describe advisory execution ambiguously although the existing implementation blocks bounded web/package findings | Specify the existing bounded gate, correct authored help, and qualify E2E finding normalization without expanding local checks | C-007/C-010; S-006/S-008/S-010; V-010/V-012 |

`must_fix` applies only to these slices and their completion conditions. No
production defect or unrelated cleanup is authorized by a finding. Lower test
counts alone are not evidence of improvement; the result must remove machinery
and duplicated decisions while retaining meaningful assurance.

### Decision, migration and unresolved-risk ledger

The action and validation columns above define each fix. This ledger records why
that disposition is preferable over preserving the baseline, its affected areas,
and its long-term implications.

| Gap | Areas changed | Rationale and long-term benefit | Compatibility / migration impact | Risk if left unresolved |
| --- | --- | --- | --- | --- |
| F-007 | Implementation, tests, routing, guidance | Consolidate useful residual checks into one parsed source pass and established import policy; eliminate duplicate traversal and parser semantics | Retire the E2E scanner after successor qualification; no replacement analyzer or rule language | Divergent enforcement, false positives and multiple maintenance sites for every syntax change |
| F-008 | Analyzer configuration, tests, specification | True runtime roots expose dead support files; real-analyzer positive and negative controls establish capabilities instead of assuming parity | Specs/setup/support tests remain roots; imported helpers remain ordinary dependencies; dead code becomes visible to the existing bounded gate | Blanket entries silently hide orphans and make removal of the old scanner unsafe |
| F-009 | Specification, tests, routing, documentation | Stable identity, finite deadlines and observation-only behavior directly test the needed postconditions; removing source grammar permits future legitimate assertion styles | Remove five scanner rows and private syntax/fixture exceptions; preserve selector output and production behavior | Brittle parser restrictions grow while wrong-record, wrong-version or action-retrying helpers can still pass |
| F-010 | Specification, implementation, tests, import policy | Restrict the actual execution module and share one same-origin validator; callers have a clear validated-operation boundary | Direct private imports migrate; remove the unused result wrapper without alias; no public API or data migration | The ineffective wrapper-name ban permits bypass of operation-derived method/path/status/schema validation |
| F-011 | Authored routing, generated projections, graph/source ownership tests, documentation | One explicit successor map and atomic cutover keep discovery, tiers and evidence coherent | Six obsolete row IDs disappear; no historical aliases; standard readiness/full analyzer placement retained | Dangling selectors, duplicated execution, orphan exceptions and misleading coverage claims |
| F-012 | Retained tests and evidence accounting | Keep checks that observe different boundaries, including real repository analysis and isolated fixtures; this retains independent failure detection | No compatibility change; each retained context has a distinct qualification purpose | Over-consolidation could discard runtime behavior, integration or failure-normalization evidence |
| F-013 | Specification and downstream projections | Amend actual adopted owners before implementation, removing accidental filename/importer rules and naming the real transport boundary | Owner text changes with its private projections; canonical discovery and identity remain | Implementation would conflict with adopted requirements or keep obsolete behavior solely to satisfy a scanner |
| F-014 | Specification, authored/generated help, tests | Describe the existing bounded blocking policy accurately and test normalized findings despite analyzer exit 0 | No new gate or local broad scan; advisory findings outside current scopes remain advisory | Operators misread a failing analysis as advisory, and future work may weaken or accidentally broaden enforcement |

## 6. Refactor Workstreams

The prerequisite column is authoritative. All workstreams are included and
sequential; none conditionally skips a predecessor. Each handoff checkpoint
records source identity, assertion dispositions, commands/exits/artifacts,
remaining gates and next action before the successor begins.

| ID / slice | Class / authoritative prerequisite | Goal and likely changes | Validation / handoff checkpoint / binary exit |
| --- | --- | --- | --- |
| WF-06 / S-006 | root; baseline revalidation and explicit implementation authorization | Final assertion ledger, focused characterization and concrete owner review packet | V-008; record baseline and REQ-103 disposition; no ambiguity about preserved obligations or retired restrictions |
| WF-07 / S-007 | chain; WF-06 | Existing frontend checker, transport-private execution and direct tests | V-009; boundary/transport checkpoint; actual raw execution is confined and transport assertions pass |
| WF-08 / S-008 | chain; WF-07 | Authored Fallow roots and existing real-analyzer fixtures | V-010; analyzer checkpoint; legitimate E2E roots/helpers and dead peers are distinguished |
| WF-09 / S-009 | chain; WF-08 plus adoption of the concrete REQ-103 amendment | Readiness witnesses and selector scanner/row retirement | V-011; behavioral checkpoint; semantic readiness and observation guarantees qualify without syntax policing |
| WF-10 / S-010 | chain; WF-09 | E2E scanner/row retirement, remaining routing, source ownership and guidance cleanup | V-012; cutover checkpoint; successors account for every obligation and no obsolete binding remains |
| WF-11 / S-011 | chain; WF-10 | Finalization, compatible final portfolio and completion handoff | V-013; final checkpoint; Section 12 implementation criteria satisfied with explicit exclusions |

Phase risks follow the same dependency chain: WF-06 can omit a normative owner
(mitigated by explicit REQ-103/676/677/070 review); WF-07 can expose raw execution
or duplicate validation (module allow/deny and transport witnesses); WF-08 can
mistake analyzer success for coverage (independent live/dead and normalized-failure
controls); WF-09 can lose readiness semantics (negative identity, timeout and
no-action witnesses); WF-10 can leave routing drift (generation, catalog,
identity, source ownership and tier assurance); WF-11 can accept incompatible or
incomplete evidence (freeze, exact-root audits, cleanup/scan review and explicit
coverage limits). A failed exit blocks the next phase; rollback boundaries are
defined in Section 7.

## 7. Authorized Refactor Slice Plan

Implementation of all six slices is authorized. Readiness describes prerequisite
evidence, not completion. Later slices inherit their workstream prerequisites
from Section 6; each checkpoint must be saved before the successor begins.

| Slice / kind | Findings / contracts | Entry evidence and readiness | Intended change / tests | Completion and rollback |
| --- | --- | --- | --- | --- |
| S-006 / owner proposal and characterization | F-007–F-011; C-006–C-010 | **DONE**, E-018/E-019; baseline passed and owner amendments adopted | V-008 baseline; adopt REQ-103/676/677/070 and AC-021 with support guidance | Record exact baseline and adoption status; roll back owners with dependent projections if the coordinated change is reverted |
| S-007 / structural simplification | F-007/F-010; C-007/C-009 | **DONE**, E-020/E-021; boundary and transport witnesses passed | Existing checker source pass and private transport split; V-009 | Positive/negative boundary controls and transport witnesses pass; rollback source/tests/authored binding changes together |
| S-008 / analyzer configuration correction | F-007/F-008; C-007/C-010 | **DONE**, E-022/E-023; real analyzer and repository scan passed | Narrow existing roots and extend real pinned-analyzer fixtures; V-010 | Orphan and unused-export controls fail correctly without suppressions; restore configuration and fixtures together on rollback |
| S-009 / adopted harness-policy change and test simplification | F-009/F-011; C-008/C-010 | **DONE**, E-024/E-025; semantic witnesses and selector retirement passed | Add negative readiness witnesses; remove selector scanner and five rows; V-011 | No grammar/selector-family inventory remains; meaningful behaviors qualify; rollback owner/source/routing/projections together |
| S-010 / structural cutover | F-007/F-011; C-007/C-010 | **DONE**, E-026/E-027; scanner retirement and routing qualified | Remove E2E scanner, row and exception; regenerate and reconcile inventories; V-012 | Every assertion group has a successor or justified retirement, with no orphan selectors or compatibility aliases |
| S-011 / final qualification | F-007–F-014; C-006–C-010 | **DONE**, E-028/E-029; compatible final qualification and handoff complete | Finalize, freeze source and execute V-013 once per required context | Compatible results and limits recorded; rollback is coordinated, slice-scoped and preserves unrelated work |

### S-006 — requirement amendment and baseline

Use the assertion successor ledger below as the initial coverage map. Revalidate
its source evidence and execute V-008 before implementation movement. Add characterization only for missing consequential behavior; do not
freeze accidental source shapes or failures.

Adopted replacement for the named-heading paragraphs of Harness REQ-103:

> Readiness helpers MUST observe the semantic state needed by the next operation
> within the existing finite, configuration-backed deadline. When identity
> matters, the predicate MUST use the required stable surface, record and version
> identity rather than visible text or count alone. Accessible role/name queries
> MAY assert accessible content or actual focus. Enforcement MUST validate the
> helper's behavior and postcondition; it MUST NOT prescribe a particular matcher
> grammar or prohibit aliases merely because a source scanner cannot interpret
> them. Helpers MUST remain observation-only after the invoking action and MUST
> NOT move focus or repeat an action to manufacture a passing result.

Retain REQ-103's existing diagnostic fields, payload exclusion and product-failure
classification, and retain REQ-665's response/version convergence. Preserve exact
inspector-version matching separately from mutation convergence's accepted-version
floor; do not collapse these different postconditions. Update AC-021's references
to selector-policy assurance to name the corresponding behavioral assurance.
The user selected this concrete semantic amendment and authorized its adoption
with implementation. WF-06 applies it to the NLSpec and AC-021. Human projection
review remains distinct from passing tests. REQ-676 now retires the closed root
layout/custom importer rule in favor of qualified Fallow analysis; REQ-677 names
the actual private execution boundary. REQ-070 explicitly retains bounded web
and package enforcement, with advisory findings outside those scopes.

### S-007 — one boundary implementation and a real transport seam

Reuse the existing frontend checker for package facades, private application
imports, generated-source ownership and grid-vendor imports. Add only two narrow
owner-local checks: grid-vendor selector leakage in E2E source and application
route/control-token leakage into shared test-utils. Consume the same enumerated
and parsed file records; do not add a second walker, parser, configurable syntax
rule language or copied rule registry. Preserve the narrow ownership restrictions
with allowed/forbidden examples; comments or fixture descriptions are not runtime
dependencies. Existing authored import rules remain independently tested.

Move raw JSON execution into `apps/web/e2e/support/transport/rawPublicJson.ts`.
Keep origin binding and request-context types in `publicJsonClient.ts` for actual
consumers. Maintain one same-origin path validator shared inside the transport
owner; do not duplicate validation or introduce a circular runtime dependency.
Remove the result-only `requestPublicJson` wrapper. The operation client uses the
raw module, and direct transport tests observe that same implementation.

Declare the raw module restricted through the existing path-prefix import policy;
allow its operation client and direct transport test only. Public context/origin
consumers do not gain raw execution through a re-export. Preserve omission versus
supplied body/headers, method/path/query generation, non-2xx payloads, success
status/schema validation, repeated-header availability after validated success,
and diagnostics that omit capabilities, payload values and private URLs. Update
private imports directly; no deprecated facade or wrapper remains.

### S-008 — qualify E2E reachability through Fallow

Replace `.fallowrc.json`'s `apps/web/e2e/**/*.ts` entry with
`apps/web/e2e/**/*.spec.ts`, `apps/web/e2e/**/*.spec.tsx`, and the existing global
setup/teardown files. Add `apps/web/e2e/**/*.test.ts` to the current Vitest owner
entry patterns, matching `harness-node` discovery. `fixtures.ts`, support and page
helpers are ordinary imports, not unconditional roots. Use existing config/owner
shapes; no new dynamic-reachability family or schema is needed.

Extend the existing real-analyzer fixture with a live spec, setup/teardown,
support test, imported helper, orphan helper, unused exported value and type,
namespace import, re-export and literal dynamic import. Include a negative control
where the blanket E2E root would hide an orphan. Expected live/dead identities are
independent assertions; fixture configuration projects the actual authored E2E
root policy rather than an unrelated copied policy. Keep the real repository scan
as a separate configuration-integration witness.

Do not assume Fallow is equivalent before these witnesses run. If the pinned
analyzer cannot supply required assurance, block deletion of the affected bespoke
check and revise this plan. Do not introduce a replacement analyzer, suppression,
compatibility import rule or blanket helper entrypoint. Unexpected unrelated
findings require scoped diagnosis rather than silent cleanup or weaker policy.

### S-009 — behavior instead of selector-expression policing

Preserve UI Contracts tests for deterministic stable IDs, escaping, collision
resistance, invalid identifier rejection and view-registry validation. Preserve
runtime selector outputs and existing production tests; removing a source ban
does not require rewriting their selectors.

Extend the existing row/inspector helper suites with controlled transitions that
prove same text on the wrong record, wrong surface, wrong exact inspector version
and a not-ready inspector cannot satisfy identity readiness. Prove delayed correct
identity succeeds, persistent mismatch times out with bounded payload-free
diagnostics, and observation does not change focus or dispatch an action. Retain
configuration-backed deadlines; do not increase them or introduce arbitrary
settling sleeps, retry wrappers or test-only production APIs.

After amendment adoption and qualification, delete the selector scanner and its
five rows. Remove prefix lists, blanket raw-literal prohibitions, fixture-name
exceptions and heading-expression grammar. The package facade and actual semantic
postconditions remain the assurance owners. Update source ownership and guidance
with the deletion, regenerating through Make as needed.

### S-010 — retire the E2E scanner and complete routing

Delete the E2E scanner only after S-007/S-008 successors and S-009 behavior qualify.
Remove its catalog row and the scanner-specific core-support exception. Retire
closed E2E root placement, catch-all filename bans, mandatory external-importer
counting and parser-specific namespace/dynamic-expression restrictions. Dead-code
assurance uses Fallow's actual semantics; semantic catalog identities remain with
the existing identity owner.

Update affected authored family manifests, frontend source ownership, backing-file
declarations and support README/guide links. Reuse current canonical frontend and
Fallow assurance operations. Preserve standard placement for behavioral readiness
and full placement for analyzer assurance; consolidation must not promote broad
analyzer work into local gates. Regenerate downstream artifacts through Make and
remove obsolete active references without aliases. Human history may name retired
paths; executable consumers may not depend on it.

### Assertion successor ledger

| Baseline assertion group | Classification / successor | Retirement condition |
| --- | --- | --- |
| E2E current-view-schema literal declarations | Incidental syntax restriction; UI Contracts registry validation and public facade usage remain | Delete the literal ban without adding a new schema-ID source scanner |
| E2E raw grid-vendor selectors | Useful architectural boundary; existing frontend source pass | Positive allowed owner and forbidden consumer witnesses qualify |
| E2E phase/catch-all helper filenames | Retire catch-all naming restrictions; retain canonical semantic identity rules where applicable | Do not move the old filename blacklist into another specification |
| E2E closed root and reachable support | Root layout is incidental; reachability moves to Fallow | True roots and live/dead helper controls qualify |
| Every support/page export needs an external importer | Replace custom importer counting with actual unused export/type assurance | Real Fallow detects independently seeded unused exports/types |
| Namespace and ambiguous dynamic-import bans | Parser limitations, not production requirements | Remove extra grammar restrictions; preserve existing canonical catalog discovery requirements |
| `requestPublicJson` named import/call ban | Defective proxy for transport ownership | Actual raw execution module is confined; validated operation behavior qualifies |
| Routes/payload/control-token literals in test-utils | Retain route-neutral shared support boundary in current frontend source pass | Allowed neutral helper and forbidden route/control witnesses qualify |
| E2E contract/package facade string checks | Duplicate existing checker/package exports | Retain canonical positive/negative import witnesses; delete duplicate string searches |
| Direct heading focus versus other heading expressions | Former owner grammar; explicitly amended in WF-06 | Adopt REQ-103 replacement and qualify semantic readiness/focus witnesses |
| Delivery-shaped selector helper names in production source | Retire duplicate private pattern scan; retain applicable canonical semantic identity validation | No second naming registry or compatibility alias |
| Hard-coded authentication selector-family classification | Incidental duplicate inventory | UI Contracts application-selector behavior remains qualified |
| Blanket selector template/raw-literal scan and fixture allowlists | Incidental construction constraints | Deterministic identity, registry validation and helper behavior remain qualified |
| Authentication helper-level literal sinks | Incidental spelling constraints | Preserve actual operation/selector behavior, not a parallel list of accepted source spellings |

S-011 rolls back only the coordinated authored, test, routing, documentation and
generated changes for this iteration. No production data, dependency, golden or
public artifact migration is required. Do not broaden a rollback into unrelated
work or recreate private aliases to accommodate old test arrangements.

## 8. Validation Plan

All commands run from `/home/jochi/code/cartulary`. Product/harness execution below
is authorized; actual execution and results are recorded in Section 10. Installed pinned Node/pnpm/Fallow and ordinary native
harness prerequisites are required; use the pin registry, not versions copied
into this document. No product browser workflow is selected. The WF-10 execution
contract preset includes its declared build and service-backed harness checks;
these are part of that canonical qualification, not a product browser suite.
Use fresh explicit run IDs and retain exact roots and native outer exits.

| ID / layer | Stage / slice / contracts | Exact command or defined portfolio | Expected result / failure gate |
| --- | --- | --- | --- |
| V-008 / focused characterization | Pre-move; S-006; C-007–C-010 | `make test-slice OWNER=web.architecture`; the pre-move browser selection below; the UI selector selection below | Current owner-aligned baseline plus known incidental checks; failures diagnosed before movement; not current qualification from historical runs |
| V-009 / boundaries and transport | Post-slice; S-007; C-007/C-009 | `make test-slice OWNER=web.architecture ROWS=web.architecture.support.frontend_import_boundary_checker_assurance_393c407a70`; `make lint CHECK=frontend-boundary`; the transport selection below | Real checker integration, fresh positive/negative fixtures, transport semantics and privacy pass; a failed required witness blocks S-008 |
| V-010 / reachability | Post-slice; S-008; C-007/C-010 | `make test-slice OWNER=web.architecture ROWS=web.architecture.support.fallow_static_and_protocol_analyzer_assurance_0794130901`; `make frontend-fallow-static` | Pinned analyzer detects E2E dead peers, keeps true roots live and accepts repository configuration; no suppression growth |
| V-011 / behavioral support | Post-slice; S-009; C-008 | The readiness and UI selector selections below | Negative identity, finite timeout, payload-safe diagnostics and unchanged focus qualify; no syntax scanner needed |
| V-012 / projection/cutover | Post-slice; S-009/S-010; C-010 | `make generate`; `make source-check CHECK=generated`; `make source-check CHECK=schema`; `make source-check CHECK=catalog`; `make source-check CHECK=identity`; `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19` | Authored inputs agree with generated outputs; exact selection and source ownership reconcile; obsolete references absent |
| V-013 / final portfolio | Final regression gate; S-011; C-006–C-010 | `make agent-finalize`, then the frozen-source portfolio below | Review any finalizer mutations before freezing; all selected obligations pass on compatible source identities |
| V-014 / documentation | Document-save check; current invocation | `make lint-markdown` with installed-only preparation and external temporary result/cache paths when safe; manual handoff/diff review | No product execution, installation, generation or authored auto-fix; document actual lint coverage and any skip reason |

### Exact focused selections

Pre-move browser selection:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131,harness.browser.boundary_support.publicjsonclient_suite_a3aedb76f9,harness.browser.boundary_support.public_http_diagnostic_privacy,harness.browser.boundary_support.source_publicjsonclient_test_ts_a6c5c0d17c
```

Transport selection, retained after scanner deletion:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.publicjsonclient_suite_a3aedb76f9,harness.browser.boundary_support.public_http_diagnostic_privacy,harness.browser.boundary_support.source_publicjsonclient_test_ts_a6c5c0d17c
```

Readiness selection; add new named cases to the same behavioral rows' exact title
lists, keeping independent cases rather than introducing a scenario DSL:

```sh
make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.timelineworkbooktestsupport_waits_for_delayed_wo_8de9145b31,web.architecture.boundary_support.workbookinspectortestsupport_suite_b85a10b67a
```

UI selector selection:

```sh
make test-slice OWNER=package.ui ROWS=package.ui.frontend_unit.selector_core_contracts_375732dc80,package.ui.frontend_unit.application_selector_contracts_675995f2ab
```

### Final portfolio and evidence limits

After focused exits, `make agent-finalize` precedes the broader final portfolio.
Supply `RESULTS_DIR` only for an existing compatible successful full warm-check
root; otherwise record retained-run maintenance skipped. Do not launch a redundant
check solely to produce that root.

Freeze source after reviewing finalizer changes, then qualify this portfolio.
Compatible complete focused evidence may be retained after the final validity
audit; do not rerun a passing context solely to change its run ID:

- `make test-slice OWNER=web.architecture` — includes actual frontend boundary
  enforcement, its assurance, Fallow assurance, source ownership and readiness.
- The exact transport and UI selector selections above.
- `make frontend-fallow-static` — actual repository reachability, distinct from
  isolated analyzer fixtures.
- `make source-check CHECK=all`.
- `make harness-contract SUITE=execution` — canonical graph assurance for the
  retired-row sentinel and retained standard/full tier placement; the existing
  execution preset is the narrowest public preset containing that assurance.
- `make lint CHECK=types`, `make lint CHECK=frontend`, and `make lint CHECK=scripts`.
- `make lint-markdown` with its actual configured coverage recorded.

Do not separately replay a checker operation already covered by this owner union.
For each accepted root retain the native exit, selected obligations, source and
toolchain identities, summary, cleanup outcome and retained-secret scan. Use
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all` for required
canonical validity, not as a substitute for test execution or full-owner audit.
Use `make test-evidence-audit OWNER=web.architecture EVIDENCE_ROOTS_FILE=<path>`
with an explicit manifest of exact compatible roots before claiming complete
architecture-owner acceptance. The transport and UI selections remain subsets.
After a repair, reassess compatibility and missing obligations before rerunning.
No blanket `check`, CI, product browser/visual/accessibility/UI-review, benchmark
or release qualification is required by this bounded change. Service-backed
checks are limited to the existing execution-contract preset's declared work.

| Validation / invocation admission | Side-effect inspection | Current execution posture |
| --- | --- | --- |
| Discovery: `make help ALL=1`, `make task-guide OWNER=web.architecture`, `make task-guide OWNER=harness.browser` | Public help/diagnostic paths and existing workspace admission inspected; no readiness installation or service work | Revalidated for implementation in E-018; final public preset confirmed in E-026/E-027 |
| V-008–V-013 | Generate/finalizer can rewrite projections; tests and lint create execution artifacts | Authorized and executed through workstream checkpoints; final coverage and limitations in E-029 |
| V-014 | `lint-markdown` can ordinarily install dependencies and write cache/results; installed-only readiness is read-only; supported environment paths permit external temporary outputs | E-016 is historical; implementation lint and manual handoff review are recorded in E-029. Configured lint globs exclude this handoff |

## 9. Top-Level Work Tracker

| ID | Work item / phase | Workstream | Status | Prerequisite / references | Exit condition |
| --- | --- | --- | --- | --- | --- |
| T-001 | Save and review the document-only plan | Historical planning session | DONE | E-010–E-015; V-014 | Only admitted handoff changed; validation and limits recorded |
| T-002 | Baseline, assertion mapping and concrete owner amendments | WF-06 | DONE | S-006; V-008 | Baseline and amendment disposition recorded before code movement |
| T-003 | Consolidate boundaries and raw execution ownership | WF-07 | DONE | Section 6; F-007/F-010; V-009 | Transport and boundary checkpoint qualifies |
| T-004 | Qualify Fallow E2E reachability | WF-08 | DONE | Section 6; F-008; V-010 | Real live/dead E2E controls qualify |
| T-005 | Replace selector grammar with semantic behavior assurance | WF-09 | DONE | Section 6; F-009; V-011 | Amendment adopted and scanner/five rows removed with qualified successors |
| T-006 | Remove E2E scanner and reconcile routing | WF-10 | DONE | Section 6; F-011; V-012 | No obsolete active scanner/row/exception; generated consistency |
| T-007 | Finalize and qualify compatible final portfolio | WF-11 | DONE | Section 6; V-013 | Section 12 implementation criteria met |

Historical WF-00–WF-05 / S-000–S-005 remain DONE. Their F-001–F-006,
C-001–C-005 and V-001–V-007 identifiers are reserved and are not reassigned to this
iteration. Their evidence is historical and does not qualify the current
implementation.

## 10. Session Handoff Log

### Authorized implementation — current checkpoints

The current implementation follows WF-06 → WF-07 → WF-08 → WF-09 → WF-10 → WF-11.
Each completed checkpoint is saved before its successor starts. The earlier
document-only authorization statements and preservation evidence below are
historical; they do not restrict the explicitly authorized implementation.

| Evidence / workstream | State, changes and qualification | Next action |
| --- | --- | --- |
| E-018 / WF-06 entry | IN_PROGRESS. HEAD `971d6423c1bfc5f1f3ac6569fc524b596cf33f61`, branch `main`; only this tracker was dirty. All 7,638 tracked/non-ignored paths and index captured in `/tmp/cartulary-tr01-implementation-r5ea9f4a`; pre-existing tracker preserved as `prior-handoff.md`. Make help and guides for web.architecture, harness.browser and package.ui passed. | Execute V-008, amend REQ-103/676/677/070 and acceptance/support text; retain baseline results before code movement. |

WF-06 exit — E-019: DONE. The user-selected REQ-103 semantic amendment,
REQ-676 discovery/liveness simplification, REQ-677 real transport boundary, and
REQ-070 bounded enforcement are adopted in the NLSpec; AC-021 and the frontend
support guide agree. The fourteen assertion dispositions were revalidated.
Projection fidelity remains subject to human review; execution is not a substitute.

All commands used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run ID>` from the
repository root. Native outer exits were 0:

| Run ID | Command / exact selection | Result / artifact |
| --- | --- | --- |
| `tr01-wf06-architecture-20261026-01` | `make test-slice OWNER=web.architecture` | PASS; `run-summary.json`, 6/6 units |
| `tr01-wf06-transport-20261026-01` | V-008 exact four-row browser/scanner selection | PASS; `run-summary.json` |
| `tr01-wf06-selectors-20261026-01` | V-008 exact two-row package.ui selection | PASS; `run-summary.json` |
| `tr01-wf06-markdown-20261026-01` | `make lint-markdown` | PASS; `adhoc/lint-markdown/tool-run-summary.json`; handoff excluded by lint globs and reviewed manually |

Run roots are `.cartulary/test-results/<run ID>`. Baseline executable source:
`sha256:b296423ca57d61784fda96b379452ee93e05f609ad2fda8d08a584a56b1d8efc`;
toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
`git diff --check` passed. No executable source moved before these baselines.
This checkpoint closes RB-001/RB-002 and is saved before WF-07 starts.

WF-07 entry — E-020: IN_PROGRESS after saved E-019. Boundary checker source
records, transport imports/tests and authored path restrictions revalidated;
V-009 is the exit gate.

WF-07 exit — E-021: DONE. The canonical checker shares source text, AST and
import records, including existing design-token evaluation. Two owner-local
executable-syntax checks replace residual vendor-selector/route-control scans;
comments are ignored. Raw execution moved to `rawPublicJson.ts`, with only the
operation client/direct tests admitted; context retains the sole path validator
and actual shared types. The result-only wrapper is removed. Public/control
separation now covers all three public transport modules. No aliases were added.

V-009 native exits were 0 (same environment/root convention as E-019):

| Run ID | Command / selection | Result |
| --- | --- | --- |
| `tr01-wf07-boundary-20261026-01` | Exact frontend boundary checker assurance row | PASS; positive/negative syntax, raw import/re-export and privileged-control fixtures |
| `tr01-wf07-transport-20261026-01` | Exact three transport rows | PASS; actual raw execution, omission/supplied values, absolute/protocol-relative/backslash escape rejection, operation validation and diagnostic privacy |
| `tr01-wf07-integration-20261026-01` | `make lint CHECK=frontend-boundary` | PASS; actual repository source integration |

Each root retains `run-summary.json`; focused source identities are recorded in
its manifest. These are slice witnesses, not the frozen final portfolio. This
checkpoint is saved before WF-08 begins.

WF-08 entry — E-022: IN_PROGRESS after saved E-021. Existing Fallow owner,
real-analyzer fixture and publication tests revalidated; V-010 is the exit gate.

WF-08 exit — E-023: DONE. Authored roots now select E2E specs and explicit
setup/teardown, and the Vitest owner includes E2E support tests. Ordinary helpers
are not unconditional roots. Real pinned-analyzer fixtures qualified live/dead
helpers, unused values/types, type-only/namespace/re-export/literal dynamic imports,
and the negative blanket-root control. Publication rejects seeded E2E findings
even with analyzer exit 0; existing advisory behavior remains covered.

V-010 native exits were 0:

| Run ID | Command / selection | Result |
| --- | --- | --- |
| `tr01-wf08-analyzer-20261026-01` | Exact Fallow analyzer assurance row | PASS; real analyzer and publication controls |
| `tr01-wf08-repository-20261026-01` | `make frontend-fallow-static` | PASS; repository integration with zero bounded web/package findings |

Roots use the E-019 convention and retain run/target summaries, resolved config
and Fallow reports. Advisory findings remain diagnostic evidence; no suppressions
or unrelated cleanup were introduced. `git diff --check` passed. RB-003 is closed.
This checkpoint is saved before WF-09 begins.

WF-09 entry — E-024: IN_PROGRESS after saved E-023 and adopted semantic
readiness amendments. V-011 plus projection consistency gate scanner retirement.

WF-09 exit — E-025: DONE. Eight new controlled behavioral cases cover wrong
record/surface, lower and higher incorrect inspector versions, persistent mismatch,
finite configured timeouts, payload exclusion and observation without focus,
click or event dispatch. Existing helpers needed no behavior correction. Fake
clocks drive existing polling/deadlines; no production timeout was increased.
The selector scanner and its five rows/source-ownership entry are removed after
successor qualification; README guidance now describes semantic assurance.

The initial `make generate` root `tr01-wf09-generate-20261026-01` failed with
native exit 2 (normalized harness/unknown_failure) because newly authored title
lists were not ASCII-sorted. Its `generate/tool-run-summary.json` and
`generate/generate-task-surface-and-execution-topology/stderr.log` retain the
failure. Sorting the two affected row title lists repaired the authored input;
no generator or generated output was hand-edited.

Subsequent native exits were all 0:

| Run ID | Command / selection | Result |
| --- | --- | --- |
| `tr01-wf09-generate-20261026-02` | `make generate` | PASS after authored sorting repair |
| `tr01-wf09-format-20261026-01` | `make format` | PASS; only intended authored TypeScript formatting changed |
| `tr01-wf09-readiness-20261026-01` | Exact two readiness rows with new title inventories | PASS; 11 selected cases |
| `tr01-wf09-selectors-20261026-01` | Exact two package.ui selector rows | PASS |
| `tr01-wf09-cutover-generate-20261026-01` | `make generate` | PASS after scanner/row removal |
| `tr01-wf09-generated-20261026-01` | `make source-check CHECK=generated` | PASS |
| `tr01-wf09-schema-20261026-01` | `make source-check CHECK=schema` | PASS |
| `tr01-wf09-catalog-20261026-01` | `make source-check CHECK=catalog` | PASS |
| `tr01-wf09-identity-20261026-01` | `make source-check CHECK=identity` | PASS |
| `tr01-wf09-ownership-20261026-01` | Exact source-ownership row | PASS |

Roots use the E-019 convention. Generate roots retain `generate/tool-run-summary.json`;
graph roots retain `run-summary.json`. Active source/routing references to the
selector scanner are absent; `git diff --check` passed. V-011 and this cutover's
V-012 gates are satisfied. This checkpoint is saved before WF-10 begins.

WF-10 entry — E-026: IN_PROGRESS after saved E-025. All scanner successors
qualified; V-012 gates final removal and authored/generated routing reconciliation.

Cutover inspection also found the retired E2E row in the graph assurance's
independent full-tier sentinel list. Remove that obsolete sentinel and assert
standard placement for both retained readiness rows; keep both existing analyzer
assurance sentinels in full. V-012 therefore additionally requires
`make harness-contract SUITE=execution`, the existing public graph-assurance
route. No scheduler implementation, preset, or tier policy changes are needed.

WF-10 exit — E-027: DONE. Removed the 655-line E2E scanner, its catalog row,
core-support exception and graph sentinel. Both readiness rows are asserted as
standard tier and both existing analyzer assurance rows remain full tier.
Authored Fallow help now describes bounded web/package gates; Make regenerated
the task-surface projections and topology render index. No executable reference
to either retired scanner or its rows remains under `apps`, `tools` or `contracts`.

V-012 native exits were 0, under the E-019 environment/root convention:

| Run ID | Selection | Result / duration ms |
| --- | --- | --- |
| `tr01-wf10-generate-20261026-01` | `make generate` | PASS; 11170 |
| `tr01-wf10-generated-20261026-01` | `make source-check CHECK=generated` | PASS; 75219 |
| `tr01-wf10-schema-20261026-01` | `make source-check CHECK=schema` | PASS; 55972 |
| `tr01-wf10-catalog-20261026-01` | `make source-check CHECK=catalog` | PASS; 46176 |
| `tr01-wf10-identity-20261026-01` | `make source-check CHECK=identity` | PASS; 49530 |
| `tr01-wf10-ownership-20261026-01` | Exact source-ownership row | PASS; 45726 |
| `tr01-wf10-graph-20261026-01` | `make harness-contract SUITE=execution` | PASS; 20/20 units; 174681 |

These graph roots identify source
`sha256:daaea890979e0349dc4392aeffd167b41b172981b29d13dca99008fbad112eb9`.
Durations include concurrent scheduling and make no performance claim.
An exploratory `make help TARGET=harness-contract` was rejected with exit 2
because `TARGET` is not a help input; `make help ALL=1` supplied the public
selection instead. This discovery rejection did not launch verification.
`git diff --check` passed. This completed checkpoint is saved before WF-11 starts.

WF-11 entry — E-028: IN_PROGRESS after saved E-027. Run `make agent-finalize`
without `RESULTS_DIR`: no compatible successful full warm-check root was supplied,
so retained-run maintenance is intentionally skipped. Review generated mutations,
then freeze executable source and qualify V-013. Reuse E-027's complete execution
contract only if finalizer review and canonical validity establish compatibility;
otherwise rerun it on the final source. Documentation cleanup cannot supply
executable evidence and is reviewed separately.

WF-11 exit — E-029: DONE. Finalizer, compatible executable qualification,
documentation review, preservation comparison and handoff are complete. The finalizer
reported `generated.status=unchanged`, zero updated files and no rollback.
Executable source is frozen at
`sha256:daaea890979e0349dc4392aeffd167b41b172981b29d13dca99008fbad112eb9`;
all roots below use toolchain
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
The finalizer and final portfolio used `CGO_ENABLED=0`, fresh explicit run IDs and
the default `.cartulary/test-results/<run-id>` roots. Native outer exits were 0.

| Run ID | Exact selection / distinct obligation | Passed work units / duration ms |
| --- | --- | --- |
| `tr01-wf11-finalize-20261026-01` | `make agent-finalize`; required maintenance before final verification | 1/1; 29097 |
| `tr01-wf11-architecture-20261026-01` | `make test-slice OWNER=web.architecture`; complete owner, including real checker, analyzer assurance, readiness and ownership | 6/6; 8078 |
| `tr01-wf11-transport-20261026-01` | Section 8's exact three-row `harness.browser` transport selection; raw/validated semantics and privacy | 2/2; 3949 |
| `tr01-wf11-selectors-20261026-01` | Section 8's exact two-row `package.ui` selection; stable selector identity | 2/2; 2338 |
| `tr01-wf11-fallow-20261026-01` | `make frontend-fallow-static`; actual repository reachability, separate from fixtures | 2/2; 21623 |
| `tr01-wf11-source-20261026-01` | `make source-check CHECK=all`; complete generated/schema/catalog/surface/identity/toolchain/migration-input consistency | 9/9; 25696 |
| `tr01-wf11-types-20261026-01` | `make lint CHECK=types`; frontend compile/type compatibility | 2/2; 27089 |
| `tr01-wf11-frontend-20261026-01` | `make lint CHECK=frontend`; authored frontend lint | 2/2; 3273 |
| `tr01-wf11-scripts-20261026-01` | `make lint CHECK=scripts`; authored harness syntax/lint | 2/2; 749 |
| `tr01-wf10-graph-20261026-01` | Retained compatible `make harness-contract SUITE=execution`; graph/tier cutover and preset's declared supporting checks | 20/20; 174681 |
| `tr01-wf11-owner-audit-20261026-01` | `make test-evidence-audit OWNER=web.architecture EVIDENCE_ROOTS_FILE=/tmp/cartulary-tr01-implementation-r5ea9f4a/web-architecture-evidence.json`; complete-owner evidence admission | 1/1; 2868 |

Each graph root retains `run-manifest.json`, `run-summary.json`, target/row
projections, `cleanup-results.json` and `retained-secret-scan.json`. All eleven
passed `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all` with
native exit 0 and `full_validity=1`. All selected units passed, no failed/skipped/
cancelled units remain, cleanup outcomes are `completed`, and retained-secret
scans passed. The ten execution/finalization roots had zero action-cache hits;
this is functional evidence, not performance qualification. Optional resource
diagnostics were not collected.

The owner audit's `test-evidence-audit/audit.json` reports `complete_owner`, all
nine required `web.architecture` rows accepted, and no rejected or unused inputs.
Its manifest names only the exact final architecture run. Transport and UI
selector runs qualify their explicit subsets, not their complete owners. This
portfolio closes the bounded remediation scope without claiming whole-project
acceptance.

The final Fallow report has zero blocking web or package findings. Its raw health
report exits 1, retained as the existing nonblocking warning; findings outside
the bounded policies remain advisory. See
`tr01-wf11-fallow-20261026-01/frontend-fallow-static/fallow-static-summary.json`
and `tool-run-summary.json` under the result parent. No suppression, baseline,
dependency or analyzer policy was weakened to obtain the pass.

Finalizer retained-evidence and scheduler-drift maintenance were skipped because
`RESULTS_DIR` was unset; the detailed reason is `results-dir-not-provided` in
`tr01-wf11-finalize-20261026-01/unit-artifacts/finalize-summary.json`. No compatible
full warm check was supplied or manufactured. The per-root validity and complete
architecture-owner audit above are separate completed obligations.

Full `check`/CI, product browser workflows, visual/accessibility/UI-review,
dedicated race testing, benchmarks and release qualification were not selected:
no behavior in those areas changed. The execution-contract preset did run its
declared build and service-backed harness checks; their cleanup passed. The
previously diagnosed WF-09 title-order generation failure remains recorded in
E-025; its repaired generation and final source checks passed.

Manual preservation review against E-018 found only the scoped 26 changed
existing files, the new raw transport module and the two planned scanner
deletions. The index and HEAD are unchanged. Generated outputs also received
content-identical file replacements during Make generation; these are file
identity changes, not additional source changes. Comparison details are in
`/tmp/cartulary-tr01-implementation-r5ea9f4a/implementation-comparison.json`.
The pre-existing tracker is retained as `prior-handoff.md`; the separate
`testing-remediation.md` is unchanged. Ignored builds, dependencies and retained
results were not exhaustively content-audited.

Documentation validation: `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=tr01-wf11-markdown-20261026-01 make lint-markdown` passed
with native exit 0 in 8080 ms. Its leaf summary is
`.cartulary/test-results/tr01-wf11-markdown-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The configured globs cover the changed NLSpec, frontend guide and source READMEs,
but exclude this handoff. Manual review additionally checked all twelve handoff
sections, table widths, fences, local links, whitespace, the fourteen assertion
dispositions and historical/current status separation; no issues were found.
`git diff --check` passed. This leaf documentation result is not a graph result
or executable acceptance evidence.

Handoff: all six authorized workstreams are complete. Review the adopted owner
amendments and downstream implementation together; passing checks do not replace
human projection-fidelity review. There is no pending implementation gate in
this bounded effort. No public API, production data, dependency, golden or
release migration is required. Private callers and exact test routing were
migrated without shims. Continue any separate WF-33 work from its own preceding
handoff; this completion does not close or alter that effort.

Confirmed specification decisions: replace heading matcher grammar with bounded
identity-based, observation-only readiness; retire the closed E2E filename list
and custom external-importer counting; describe the real raw execution boundary;
retain the existing bounded Fallow web and package gates and advisory findings
elsewhere. Fallow stays out of local `check` report execution. No new analyzer,
schema, runtime mode, compatibility wrapper or suppression growth is admitted.

### Historical document-only planning session — E-010–E-017

The following subsections retain that session’s observations and limits. Phrases
such as “current” and “next action” below refer to that earlier session; the
implementation checkpoints above supersede its pending work.

#### Scope and authority

| Time / baseline | State and evidence | Next action |
| --- | --- | --- |
| 2026-10-26; `971d6423c1bfc5f1f3ac6569fc524b596cf33f61` | E-010/E-011: user selected in-place update, frontend policies and simplification at source. Prior Plan Mode allowed inspection only. Current task authorizes saving that document-only plan. | Revalidate baseline before later implementation; do not inherit historical implementation permission |

#### Backend boundary

Not applicable: no backend, Go fixture, production assembly, storage or service
change is part of this iteration. Initial exploratory fixture reads selected no
backend workstream.

#### Frontend boundary

| Evidence | Inspected state / affected files | Next action |
| --- | --- | --- |
| E-012/E-013 | Two bespoke scanners duplicate established analysis and encode incidental constraints; Fallow roots hide a real successor-coverage gap | S-007/S-008 establish successors before S-009/S-010 removals |

#### Contract and codegen

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-013/E-015 | REQ-103 grammar is currently adopted; authored row/source inputs and generated outputs are identified; no owner/projection edits occurred | Propose/adopt only the concrete amendment; regenerate from authored inputs during separately authorized implementation |

#### Tests and harness

| Evidence | Commands/results and applicability | Next action |
| --- | --- | --- |
| E-014/E-015 | Assertions and command bindings inspected; read-only help and two task guides passed; no tests ran | Use V-008 as the future baseline, not historical passes |
| E-016 | `make lint-markdown` passed, outer exit 0; installed-only preparation and external result/cache directories. Handoff structure, links and diff reviewed separately because lint globs exclude this file. | No tests, installation, generation or finalizer run; detailed command and coverage below |

#### Security and authorization

| Evidence | Required distinction | Next action |
| --- | --- | --- |
| E-014 | Preserve actual HTTP validation, same-origin boundaries and private diagnostics; current wrapper-name guard is incomplete | Enforce the module that actually performs raw execution; keep success/error/observed-header witnesses |

#### Open risks and next session

| Evidence | Current state / limits | Next action |
| --- | --- | --- |
| E-017 | All 7,638 tracked/non-ignored paths compared: only this handoff content changed; no added/removed paths or file-identity changes; HEAD and index unchanged | Later actor starts with baseline/adoption/characterization revalidation, not broad rediscovery |

#### Historical planning evidence register

| ID | Source / observation | Baseline and applicability |
| --- | --- | --- |
| E-010 | Git status/HEAD/index and per-path content/file-identity capture; 7,638 paths; exact comparison files in Section 1 | Current planning/save baseline; matches prior-turn fingerprints; no current test evidence |
| E-011 | Root/tooling guidance, Refactor Tracker format and modular framework; target/output identity and link checks | Applies to this sole-file write; prior completion authorization is historical |
| E-012 | E2E scanner, frontend checker/rules, Fallow configuration/fixtures and actual runner configuration | Static source evidence for F-007/F-008; real analyzer successor qualification remains future |
| E-013 | Selector scanner and adopted REQ-103/AC-021; proposed semantic amendment | Current requirement distinguished from desired future requirement; no adoption inferred |
| E-014 | Transport modules/tests, row/inspector waits/tests, selector package tests and direct import search | Test assertions inspected, not executed; no exhaustive external-consumer claim |
| E-015 | Relevant catalog, source ownership, task-surface and generated policy inputs; Make help and owner guides | Commands discovered and routing known; generated products not edited |
| E-016 | Document write, successful Markdown lint (outer exit 0), manual structural/link review and clean `git diff --check` | Lint covers configured globs, which exclude this handoff; supplemental inspection found twelve required sections, fourteen assertion dispositions and no structural/link issues |
| E-017 | Final index/content/file-identity comparison; `/tmp/cartulary-tr01-doc-hla6fi0g/comparison.json` | Only admitted output content changed; no added/removed paths or identity changes; HEAD/index unchanged. Ignored trees were not exhaustively audited |

#### Historical document validation — E-016/E-017

Executed from the repository root:

```sh
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-tr01-doc-hla6fi0g/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-tr01-doc-hla6fi0g/static-analysis \
CARTULARY_TEST_RUN_ID=tr01-frontend-plan-doc-20261026-01 \
COREPACK_ENABLE_NETWORK=0 make lint-markdown
```

Outer exit: **0**. Step summary:
`/tmp/cartulary-tr01-doc-hla6fi0g/results/tr01-frontend-plan-doc-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The existing lint configuration excludes this handoff; the pass qualifies only
its configured documents. Supplemental document inspection covered section
structure, table widths, fenced blocks, whitespace, local links, the fourteen
assertion dispositions and preservation of historical evidence. `git diff
--check` passed. These are documentation checks, not test execution or product
qualification. Review notes and the content/index comparison are retained in
`document-review.json` and `comparison.json` beside the temporary baseline.
No tests, format, generation, finalizer or product verification ran, as required
by the current document-only scope.

### Historical graph/scanner iteration — E-001–E-009

The previous iteration completed WF-00–WF-05 / S-000–S-005: canonical graph
assurance replaced duplicate modes; graph policy, diagnostics and execution
fixtures gained clear owners; four scanner shell suites became directly registered
Node tests; Staticcheck now fails before scanning when package discovery fails or
is empty. Public commands, security thresholds and distinct diagnostic/publication
witnesses were preserved. No aliases or production test-only APIs were retained.

The full prior tracker is retained in Git at
`971d6423c1bfc5f1f3ac6569fc524b596cf33f61:docs/handoffs/testing-remediation-01.md`
and in the temporary pre-edit copy identified in Section 1. Its statements about
uncommitted work describe that historical session, not this clean starting
checkout. Artifact contents have not been re-audited in this document session.

| Evidence | Historical outcome and retained reference | Qualification limit |
| --- | --- | --- |
| E-001 | Prior implementation resumed on HEAD `3d192eea610867bac59f4eaf79ba54f4dd05dd68` with preceding dirty work present | Historical authorization only |
| E-002 | Baseline `tr01-baseline-20261026-01`: regression passed, 55 graph units; source `sha256:aa432d498946013285d924d51e7f754a7d1bfc37a40492be0cb5c09e413ca9d0` | Pre-move historical baseline |
| E-003 | Graph policy/wiring split; first expectation failure repaired; `tr01-wf01-graph-20261026-02` passed, 60 tests | Slice evidence, earlier source |
| E-004 | Canonical matrix and explicit published-unit fixtures; generation and invalid fixture-ID failures repaired; `tr01-wf02-graph-20261026-02` passed, 74 tests | Earlier-source generation/schema roots remain in prior tracker |
| E-005 | Four scanner suites consolidated; `tr01-wf03-scanner-20261026-01`, `tr01-wf03-graph-20261026-01`, `tr01-wf03-source-20261026-01` passed | Wrapper refactor before discovery correction |
| E-006 | `tr01-wf04-negative-20261026-01` exposed partial-output discovery failure; corrected scanner and actual Go lint passed at `tr01-wf04-scanner-20261026-01` / `tr01-wf04-go-20261026-01` | Intentional fail-closed correction; earlier source |
| E-007 | Initial final contract failed because graph umbrella imported a backend-private test; direct operation registration repaired ownership | Initial portfolio superseded |
| E-008 | Regeneration, exact boundary row and second finalizer passed; final compatible portfolio below accepted | Retained maintenance excluded because RESULTS_DIR unset |
| E-009 | Historical `tr01-final-markdown-20261026-01` lint passed; prior bounded iteration closed | No current-document or current-source test qualification |

### Historical final acceptance portfolio

The following retained portfolio is reproduced from E-008/E-009. Its statements
report the earlier session; they do not certify this iteration or a new run.


Every command below used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run ID>`. Every
native command exited 0. Roots are `.cartulary/test-results/<run ID>`; each contains
`run-manifest.json`, `run-summary.json`, target/unit results, `cleanup-results.json`
and `retained-secret-scan.json`.

Executable source: `sha256:b296423ca57d61784fda96b379452ee93e05f609ad2fda8d08a584a56b1d8efc`.
Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
Earlier-source passes and failed characterization runs are historical evidence,
not part of this final acceptance union.

| Run ID | Command | Duration ms | Added assurance |
| --- | --- | --- | --- |
| `tr01-final-contract-20261026-02` | `make harness-contract` | 114368 | Default contracts, graph matrix, owner boundaries and real Make/private diagnostic channel; 11/11 units |
| `tr01-final-scanner-20261026-02` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.go_scanner_wrappers` | 14457 | All 20 wrapper cases, including discovery failure before scan |
| `tr01-final-output-20261026-02` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.output_accounting` | 27139 | Distinct step/target/tool publication, including valid and malformed Govulncheck findings |
| `tr01-final-source-20261026-02` | `make source-check CHECK=all` | 35700 | Generated, schema, catalog, identity, surface, toolchain and migration-input consistency |
| `tr01-final-scripts-20261026-02` | `make lint CHECK=scripts` | 625 | Authored script validation |
| `tr01-final-shell-20261026-02` | `make lint CHECK=shell` | 23784 | Shell analysis including the changed Staticcheck wrapper |
| `tr01-final-go-20261026-02` | `make lint CHECK=go` | 14032 | Actual Go formatting, vet and Staticcheck invocation |

For every listed root, `make harness-observability-check RESULTS_DIR=<exact-root>
CHECK=all` passed with outer exit 0 and `full_validity=1`. All cleanup outcomes
were `completed`; retained-secret scans passed; no failed, skipped or cancelled
units remain. This qualifies the selected obligations, not complete acceptance
of every owner or the whole project. No performance claim follows from these
functional timings; optional resource diagnostics were not collected.

Full CI, browser/visual/UI-review, dedicated race-suite execution, benchmarks,
release qualification and complete audits of other owners were not run: no
production behavior in those areas changed. Their selection/isolation guarantees
remain covered by the graph assertions. Finalizer canonical retained-evidence and
scheduler-drift maintenance were skipped because `RESULTS_DIR` was unset; no
compatible full warm check was supplied and no redundant check was created for
maintenance. This does not replace the per-run validity checks recorded above.

## 11. Open Questions and Blockers

No unresolved scope or implementation-design choice blocks this implementation.
The following entry gates are closed by actual workstream evidence, not planning
completion or historical passes.

| ID | Blocker / affected slice | Needed evidence or authority | Status / resolution condition |
| --- | --- | --- | --- |
| RB-001 | Current pre-move characterization for S-007 | V-008 on a revalidated source baseline | DONE; E-019 retains all three passing baseline roots |
| RB-002 | Owner amendments for S-007–S-010 | User selected concrete semantic/simplification/bounded-gate decisions and authorized implementation | DONE; E-019 records adopted REQ-103/676/677/070 and AC-021 changes |
| RB-003 | Real Fallow E2E successor assurance | V-010 real analyzer, publication and repository integration | DONE; E-023 |
| RB-004 | Successor and final portfolio gates; S-009–S-011 | Predecessor exits and V-011–V-013 | DONE; E-025/E-027/E-029 close all gates |

The authorized owner amendments are adopted in E-019. F-014 resolved the
ambiguous Fallow specification/help wording while retaining the existing bounded
implementation policy. Human projection review remains distinct from passing tests.

## 12. Binary Completion Criteria

| Outcome | Current statement |
| --- | --- |
| Planning completeness | **complete** for the bounded frontend test-policy iteration; not an exhaustive suite audit |
| Workstream completion | WF-06–WF-11 **DONE**; each predecessor checkpoint was saved before the next entry; validation and handoff completed last |
| Implementation authorization | **granted by the current implementation instruction** for WF-06–WF-11 and their scoped verification |
| Implementation write boundary | **Verified**, E-029: scoped source/specification/test/routing edits and generated projections only; index/HEAD unchanged. E-017’s sole-file result belongs to historical planning |

Historical document-only completion criteria (E-016/E-017):

- [x] Existing canonical target and requested output identity established.
- [x] Current observations distinguish direct inspection from metadata accounting.
- [x] Each finding maps to a contract, slice and validation obligation or justified no-action disposition.
- [x] All fourteen baseline assertion groups have a successor or reasoned retirement.
- [x] Dependency order, entry gates, rollback and exact validation selections are defined.
- [x] Incidental rule changes are separated from unchanged production requirements.
- [x] Historical completion IDs, final portfolio and exclusions are retained.
- [x] No tests, production changes, owner amendments or generated edits were executed.
- [x] Documentation validation and its actual coverage/limitations recorded.
- [x] Final content/index comparison verifies the admitted sole-file write.

Implementation completion criteria (updated at workstream exits):

- [x] Both bespoke scanners and obsolete routing are removed without aliases.
- [x] Every meaningful obligation has qualified successor coverage at its appropriate boundary.
- [x] Incidental naming, literal and matcher restrictions are retired explicitly.
- [x] The concrete REQ-103 amendment is adopted before dependent removals.
- [x] Actual raw transport execution is confined and all retained transport semantics qualify.
- [x] Fallow detects dead E2E peers without blanket roots or suppression growth.
- [x] Selector identity and bounded observation-only readiness remain qualified.
- [x] Authored/generated state and exact selected-row accounting agree.
- [x] No new analyzer, compatibility mechanism or generic policy framework is introduced.
- [x] Final compatible evidence, exclusions and all six checkpoints are recorded.

T-001 records the historical document update. Current implementation progress is
recorded by T-002–T-007 and the current checkpoints in Section 10. The separate
WF-33 remains open in the preceding handoff; historical passes do not qualify
the current implementation.
