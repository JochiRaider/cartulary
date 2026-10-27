# Testing Remediation 01: Complete Boundary Assurance Remediation

## 1. Scope and Source Posture

Implementation authorized by the user's instruction to implement the expanded
proposed plan on 2026-10-27. The approved scope includes previously deferred
F-057/F-058 and a shared Go parser for local and central import inspection.
This supersedes the earlier document-only scope, not its historical evidence.

Repository and canonical target: `/home/jochi/code/cartulary`, `.`. This file is
the controlling artifact. Preserve twelve sections, stable IDs and prior logs.
WF-44 remains the final validation and handoff workstream. Each workstream exit
must be saved here before its successor starts; no parallel agent work.

Baseline: `main`, HEAD `1f6158c2778d14483c8a801705214d1815625422`. The only
pre-existing tracked modification was this tracker (SHA-256
`5a513b607773531d7a9ee21a1717f46deab22d42ccb2256f9b0a7bbe1bb7d292`). Its content
and Git diff were preserved under `/tmp/cartulary-testing-remediation-implementation/`.
No production pass or whole-repository behavioral audit is inferred from inspection.

Authority remains Core 00–04, adopted Entities/Indicators/Revisions decisions and
Testing Harness REQ-012–016/062/063/066. Reporting 1.4.0 remains proposed over its
adopted 1.3.0 base. Owner prose is human input only: executable checks, generators,
runtime metadata and evidence must not inspect or depend on Markdown.

No public API, replay, database-lineage or wire migration is intended. No filename
census, compatibility wrapper, production test hook, new public runner or policy
engine is admitted. Newly exposed production defects block only dependent removal
and receive an explicit corrective disposition, never a permissive exception.

## 2. Current-State Repository Inventory

Revalidated selected source files, policies, owners, Make surface and routed rows
from the earlier inventory. Historical enumeration is not a new exhaustive audit.

| Surface | Current diagnosis / intended owner |
| --- | --- |
| Entities, Indicators, Reporting boundary guards | Duplicate Go parsers and mixed semantic/source-layout checks; module policies stay local |
| Revisions application/history/rollback guards | Substring import/authorization checks and filename/count inventories; qualify semantic successors before removal |
| `internal/testutil/sourcecheck` | New shared Go syntax mechanics, independent of module policy |
| Central backend checker and SQL relation classifier | Go regex misses syntax; relation classifier remains the SQL analysis mechanism |
| Indicator envelope schema fixture | Absence helper incorrectly accepts one to six forbidden mirror columns |
| Verification owners/families, work graph, task surface, cache policy | Authored routing/build dependencies first; generated outputs only through Make |
| Specifications and this handoff | Owner clarification, dispositions, serial checkpoints and evidence |

## 3. Module Boundary Diagnosis

Share parsing mechanics; keep dependency authority beside modules. Centralize
static relation ownership policy and retain read/write/lock distinctions. Security,
transactional effects, history lookup and provider dispatch require executable
behavioral evidence. File layout and private names are not replacement contracts.

## 4. Public Contract and Behavior Freeze Map

| Contract | Required behavior / validation |
| --- | --- |
| C-037 | Compiler-backed imports: aliases, raw/escaped strings, repeated groups, comments, build-tag-independent inspection and fail-closed errors; V-060/066 |
| C-038 | Existing module capability scopes, support exclusion and exact-file storage permissions; V-061/067 |
| C-039 | Source/route/effect ownership, merge atomicity, Indicator envelope authority, indexed history and generic dispatch; V-062/067–069 |
| C-040 | Current actor/incident/command authorized inside delete/restore/rollback transaction; denial commits nothing; V-062 |
| C-041 | Exact unique routing, actual service/isolation metadata, truthful claims and final-source evidence; V-063/064 |

The support API is `FileImports{Path string, Imports []string}`,
`ParseImports(path string, source []byte) ([]string, error)` and
`ReadImports(source fs.FS) ([]FileImports, error)`. ReadImports excludes test Go
files and non-Go entries, never follows links, sorts normalized relative paths,
and returns no partial success. Module callers choose package scope and require
nonempty production selection. The central private batch adapter receives source
bytes rather than rediscovering the repository; policy evaluation remains pure.

## 5. Coupling and Boundary Findings

| Finding | Remediation / areas | Benefit, compatibility, unresolved risk | Validation |
| --- | --- | --- | --- |
| F-054 / must_fix | Shared Go import parser; support implementation/tests/routing/docs | Consistent syntax without shared policy; private helper cutover only; duplication otherwise drifts | V-060/061 |
| F-055 / must_fix | Retire filename/operation-count assertions; tests/routing/docs | Harmless reorganization stops breaking tests; no aliases; inventories otherwise substitute for assurance | V-061/068 |
| F-056 / must_fix | Behavioral merge/effect/authorization witnesses and admission-package imports; tests/routing/docs | Observe actual guarantees without private spelling; no public migration; security/atomicity can otherwise falsely pass | V-061/062 |
| F-057 / must_fix | Shared parser via private central batch adapter; harness/build/tests/spec/docs | One Go parser; public output unchanged; missed imports become failures rather than fallback | V-065/066 |
| F-058 / must_fix | Central relation policy plus history/dispatch/publication witnesses; policy/tests/spec/docs | Remove SQL/layout coupling without losing guarantees; no runtime migration; substring blind spots otherwise persist | V-065/067/068 |
| F-059 / must_fix | Authored routing and real fixture metadata before generation | Exact claims and accounting; dead selectors removed without aliases | V-063/064 |
| F-060 / must_fix | Exact zero mirror columns and individual negative schema cases; tests/docs | Partial schema regression detected; no production DDL change | V-069 |

D-016–029 remain the assertion ledger in the archived original active plan below.
Superseding decisions: D-018/D-023 SQL mechanisms move in WF-47; D-021 mirror
scanning retires only after V-069 and application witnesses; its literal source
identity check remains. D-027 moves to behavioral/index/provider witnesses in
WF-48, with incidental file-purity assertions explicitly retired. All other
retirement prerequisites remain intact.

## 6. Refactor Workstreams

This table is the sole dependency definition. Every row is a separate slice.

| Workflow / slice | Prerequisite | Deliverable / binary exit |
| --- | --- | --- |
| WF-41 / S-041 | none | Revalidate ledger, owners and exact rows; retain baseline results and classify failures |
| WF-45 / S-045 | WF-41 | Clarify harness syntax/SQL evidence and Revisions coordination ownership before dependent changes |
| WF-42 / S-042 | WF-45 | Shared parser and independently routed mechanics tests pass before consumers move |
| WF-46 / S-046 | WF-42 | Central adapter, declared build/cache inputs and CLI contract qualify |
| WF-43A / S-043A | WF-46 | Entities imports plus merge/Mentions effects qualify before spelling retirement |
| WF-43B / S-043B | WF-43A | Indicators package admission, decoding and exact mirror absence qualify |
| WF-43C / S-043C | WF-43B | Reporting facade imports qualify; SQL retained until WF-47 |
| WF-43D / S-043D | WF-43C | Revisions imports, inventory retirement and real transaction denial matrix qualify |
| WF-47 / S-047 | WF-43D | Central SQL policy passes independent negative/positive fixtures and real sources before scanner removal |
| WF-48 / S-048 | WF-47 | Indexed associations, generic dispatch and transaction participants qualify before D-027 retirement |
| WF-44 / S-044 | WF-48 | Finalize/freeze, complete compatible portfolio and routing, close handoff |

WF-43/S-043 and T-050 remain rollup identities for WF-43A–D. Numeric order does
not override the dependency column. Revert an incomplete slice's implementation,
tests and authored routing together; regenerate projections, never add a shim.

## 7. Proposed Refactor Slice Plan

The approved implementation sequence is Section 6. All implementation workstreams
have exited; WF-44 final qualification and handoff are complete (E-124).
The user selected shared Go parsing and inclusion of F-057/F-058. No pending
architecture selection remains. Newly discovered defects require specific evidence
and corrective slices, not silently broadened permissions.

The central adapter batches `{path, source}` records and returns decoded imports
for every path. Missing/duplicate/mismatched results and process/parse failures
reject. Preparation uses pinned Go and existing Make-owned artifact machinery;
no import-time I/O or fallback parser. SQL classification remains bounded static
assurance and never substitutes for executed transaction witnesses.

Phase risks and controls: WF-41 avoids treating old passes as current evidence;
WF-45 keeps proposed Reporting prose separate from adopted authority; WF-42
rejects partial parsing and accidental non-source reads; WF-46 declares tool
preparation/cache inputs and rejects incomplete batches; WF-43A requires merge
and effect rollback witnesses before deleting spelling checks; WF-43B tests
partial schema regressions; WF-43C preserves the exact facade capability set;
WF-43D proves transaction authorization independently of routing; WF-47 retains
closed relation sets and exact-file external permissions; WF-48 separates
index usability from planner cost choice and tests independent participants;
WF-44 requires compatible final source evidence and explicit maintenance limits.
Each phase exits only at its Section 6 criterion and saved Section 10 checkpoint.

The implementation adds no product API, storage migration or compatibility alias.
Private test selectors removed from mixed rows have no replacement aliases;
new semantic rows are independently registered. The shared helper owns syntax,
module tests own import policy, and the existing central relation classifier owns
static SQL mechanics. Transaction tests remain responsible for committed effects.

## 8. Validation Plan

| ID | Obligation / canonical command |
| --- | --- |
| V-059 | `make help ALL=1`, five `make task-guide OWNER=... DETAIL=coverage JSON=1` selections, then exact existing `make test-slice` baseline rows |
| V-060 | New shared parser row: syntax, failures, exclusion, nested scope, nonregular entries, deterministic order |
| V-061 | Four module boundary/admission rows and independent allowed/forbidden reorganization fixtures |
| V-062 | Merge success/collision/failure, Mentions injected effects, three-command transaction authorization denial and successful controls |
| V-063 | `make generate`; `make source-check CHECK=generated`, `schema`, `catalog`, `identity`; `make lint CHECK=go`; `make lint CHECK=backend-boundary` |
| V-064 | `make agent-finalize`, review/freeze source, missing compatible portfolio, `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, documentation review |
| V-065 | Owner clarification review; `make lint-markdown`; direct tracker links/IDs/dependency review (default globs omit this tracker) |
| V-066 | Central/local parser parity, complete batches, explicit errors, cold/warm preparation, unchanged CLI summary/output |
| V-067 | SQL case/quoting/CTE/nesting/composition, separate read/write/lock, unresolved relation rejection, exact-file permission safety |
| V-068 | History membership independent of source JSON; valid usable association index; generic provider dispatch; same-transaction projection/publication success/failure |
| V-069 | Exact Indicator mirror absence; each forbidden column individually rejected; current application operations pass |

Use exact authored rows under module.entities/indicators/reporting/revisions and
harness.command_surface. Author new IDs with `make author-test-row-id`. Preserve
fresh run IDs, exact roots and native exits. Never infer execution from selection.
Finalizer maintenance uses an eligible successful full warm check root only;
otherwise explicitly record RESULTS_DIR unset. No redundant full check, browser,
benchmark or release claim is automatic. Diagnose failures before rerunning.

## 9. Top-Level Work Tracker

| ID | Workstream | Status | Evidence / exit |
| --- | --- | --- | --- |
| T-047 | Earlier document save | DONE | E-110–112, historical |
| T-048 | WF-41 baseline | DONE | E-113/114; five baseline runs passed |
| T-052 | WF-45 specifications | DONE | E-115; owner clarification passed |
| T-049 | WF-42 parser | DONE | E-116; parser row passed |
| T-053 | WF-46 central parser | DONE | E-117; central parser qualified |
| T-050 | WF-43A–D module rollup | DONE | E-118–121; module slices qualified |
| T-054 | WF-43A Entities | DONE | E-118; imports/effects/merge qualified |
| T-055 | WF-43B Indicators | DONE | E-119; admission/schema qualified |
| T-056 | WF-43C Reporting | DONE | E-120; facade boundaries qualified |
| T-057 | WF-43D Revisions | DONE | E-121; denial matrix qualified |
| T-058 | WF-47 SQL | DONE | E-122; central relation policy qualified |
| T-059 | WF-48 semantics | DONE | E-123; semantic successors qualified |
| T-051 | WF-44 completion | DONE | E-124; final portfolio and handoff complete |

## 10. Session Handoff Log

### WF-44 completion — E-124 / V-063/064/065

All eleven workstreams completed in the approved dependency order, with each
predecessor exit saved before its successor began. Final source identity is
`sha256:90449f35a3d942247eb403bed711d5e3efc8570347dd159006b00c1be4d091e7`;
HEAD remains `1f6158c2778d14483c8a801705214d1815625422`, with scoped uncommitted
changes. No public product API or database migration was introduced.

`make generate`, `make format`, and `make agent-finalize` passed before final
qualification. Finalizer root: `.cartulary/test-results/remediation-wf44-finalize-03`,
native exit 0, generated structure unchanged. Retained-run maintenance was
explicitly skipped because `RESULTS_DIR` was unset: no eligible full warm check
was supplied, and a redundant check was not created for maintenance.

The following final invocations all have native exit 0 and the identical source
digest above. Roots are under `.cartulary/test-results/`.

| Root suffix (prefix `remediation-wf44-`) | Command / selected obligation |
| --- | --- |
| `source-02` | `make source-check CHECK=all`: generated, schema, catalog, surface, identity, toolchain and migration-input integrity |
| `go-02` | `make lint CHECK=go`: Go checks on final sources |
| `boundary-02` | `make lint CHECK=backend-boundary`: real-source import/SQL/storage ownership |
| `harness-02` | Three exact rows: shared parser, central boundary fixtures, public registry parity |
| `entities-02` | Four exact rows: imports/routes, merge store/route outcomes and injected Mention effects |
| `indicators-02` | Four exact rows: structure/admission, exact mirror schema and transaction rollback |
| `reporting-02` | Exact facade-boundary row |
| `revisions-02` | Fourteen exact rows: construction/imports, history/planning, command controls/denial, catalogs/scope, association index, generic dispatch and transactional participants |

All eight roots passed `make harness-observability-check RESULTS_DIR=<exact-root>
CHECK=all` with `full_validity=1`. All cleanup outcomes are `completed`, retained
secret scans pass, and there are no failed/skipped/cancelled units or cache hits.
This qualifies the selected portfolio, not whole-owner, full-project or release
acceptance. Optional resource diagnostics were not collected; no timing claim
follows from functional results.

`make lint-markdown` exited 0 at `remediation-wf44-docs-01`. The default globs omit
this tracker and the decision file; direct review checked their links, IDs,
dependency order, disposition consistency and whitespace. `git diff --check`
passed. Markdown is not an executable test or evidence input.

Final qualification corrections: `finalize-01` stopped on stale generated
routing; regeneration repaired it. `finalize-02` passed. The first final module
portfolio, Go lint and source checks passed, but `boundary-01` found a comment-only
Go string falsely classified as SQL after adding nested-comment recognition.
The classifier now distinguishes comment-only data from statements, with explicit
fixtures. The finalizer and complete selected portfolio were then repeated on
one frozen source identity, producing the `-02` roots above. Earlier roots remain
historical qualification and are not mixed into final acceptance.

<details>
<summary>Exact final row selections, retained in each run-manifest.json</summary>

```sh
make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.behavior.public_registry_parity,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3
make test-slice OWNER=module.entities ROWS=module.entities.integration.the_explicit_merge_route_repoints_live_fan_out_p_0ec76e8044,module.entities.store.explicit_merge_store_behavior_repoints_live_ment_0560122310,module.entities.store.mention_commands_apply_injected_timeline_effects_4764d2ea7c,module.entities.unit.root_boundary_guards_e292ec1361
make test-slice OWNER=module.indicators ROWS=module.indicators.storage.envelope_contract_migration_down_up,module.indicators.structure.indicator_application_requires_complete_composition_6ecc4edf41,module.indicators.transaction.indicator_repository_writes_roll_back_when_revis_c37659989d,module.indicators.unit.create_admission_preserves_wire_and_replay_7ef2ad611c
make test-slice OWNER=module.reporting ROWS=module.reporting.unit.owner_boundaries
make test-slice OWNER=module.revisions ROWS=module.revisions.integration.history_source_family_paging,module.revisions.providers.catalog_guards,module.revisions.rollback.decomposition_8c5e3f2a10,module.revisions.store.covers_rollback_route_admission_derived_rollback_c6c889199c,module.revisions.store.covers_the_first_class_record_soft_delete_route_30cb39b83e,module.revisions.store.covers_the_restore_route_contract_reviewer_admin_a2df5bc94d,module.revisions.store.destructive_commands_authorize_the_current_actor_ad6fddd708,module.revisions.store.history_bounded_projection,module.revisions.store.history_lookup_uses_stored_associations_independ_c8be17a439,module.revisions.store.rollback_projection_and_publication_participants_f7381e09d3,module.revisions.support_unit.history_materialization_attribution_and_page_7b4d2e9c10,module.revisions.support_unit.http_application_boundary_4e7a9c2d12,module.revisions.unit.generic_rollback_dispatch_executes_a_newly_contr_19b83ad4e2,module.revisions.unit.projection_scope
```

</details>

Final assertion dispositions:

| Ledger | Final disposition / qualified successor |
| --- | --- |
| D-016, D-020, D-023 imports, D-024 | Shared Go syntax with local capabilities and independent policy fixtures |
| D-017 | Route and clipboard ownership assertions retained |
| D-018 | Closed central merge relation policy plus real success/collision atomicity; private planner-name heuristic retired |
| D-019 | Injected Timeline effects exercised through the public Mention command in the Entities integration package; no production test hook; constructor checks retained |
| D-021 | Literal source identity retained; schema exact-zero/individual-negative and current application witnesses replace mirror SQL spelling |
| D-022 | Admission package import policy plus actual decoder errors and replay behavior |
| D-023 SQL | Closed central Reporting relation set with exact-file, read-only external grants |
| D-025 | Incidental filenames and operation counts removed without aliases |
| D-026 | Three-command transaction denial matrix plus real route controls |
| D-027 | Association membership/index, generic provider dispatch and transaction participants; incidental file-purity/private-call restrictions retired |
| D-028 | Adjacent composition/provider/admission contracts retained |
| D-029 | Authored unique rows and truthful fixture approvals; generated projections and identity/catalog checks pass |

F-054–060 are resolved within this bounded scope. New tests do not read owner
prose. Import inspection fails closed on selected source errors and incomplete
batch/tool results. SQL analysis remains bounded static assurance; executed
transaction evidence independently establishes committed/rolled-back effects.
The existing root-package SQL constraints and real planner/paging tests remain.

No remaining implementation blocker is known in this plan. Human review of
specification/projection fidelity remains separate from test success. Reporting
1.4.0 remains proposed. Historical unrelated WF-33 is not closed by this work.
Full CI/check, browser/visual/UI review, performance benchmarks, dedicated race
qualification and release qualification were not selected: this change affects
backend assurance and harness inspection, covered by the recorded portfolio.


### WF-48 exit / WF-44 entry — E-123 / V-068

New independent witnesses cover history membership despite misleading source JSON,
a valid ready GIN association index usable for the containment predicate, generic
inverse dispatch of a newly contributed target kind, and real rollback command
transactions with injected projection/publication participants. Both participants
observe the same transaction; failures after their writes roll back durable
source, envelope, history, projection, publication and receipt effects. A committed
historical append changes retained history without live state or publication.
The filename/private-call/SQL-purity/source-kind token decomposition tests are
retired after these witnesses passed. Existing planner, ordering, projection-scope,
bounded history and source-family paging tests remain independently routed.
Three new rows exited 0 at `remediation-wf48-semantics-02`. The first attempt
`remediation-wf48-semantics-01` exposed fixture weaknesses: missing request hash
and an overly restrictive whole-query index expectation on a tiny dataset. The
hash now matches command admission requirements; index usability is checked for
the actual association predicate separately from whole-query membership, without
forcing a particular join plan. Final frozen-source qualification remains due.


### WF-47 exit / WF-48 entry — E-122 / V-067

Closed central relation policies now cover Entities merge and Reporting root.
Reporting external reads remain exact-file and read-only; existing Records and
projection grants are unchanged. Removed the obsolete Revisions `source` table
allowance. Independent policy fixtures cover case/quoting/comments, CTEs, unknown
and dynamic relations, and distinct read/write/lock failures. Local Entities and
Reporting scanners are retired; Indicator mirror scanning is retired based on
WF-43B's exact schema/application witnesses. Literal source identity stays tested.
A negative fixture exposed SQL recognition skipping commented UPDATE statements;
recognition now uses the existing relation lexer, including INSERT/DELETE comment
regressions and a non-SQL label case. No new SQL engine was introduced.
Qualification: production scan `remediation-wf47-sql-04` and central fixture row
`remediation-wf47-contract-03` exited 0. Earlier `sql-01` caught direct Collaboration
test SQL, replaced with existing owner test helpers; `sql-03` caught a bare-label
false positive, corrected with statement-shape recognition. `contract-01` exposed
the commented UPDATE defect; `contract-02` then passed. Final source validation
will rerun affected module rows after formatting. SQL evidence remains bounded
static assurance; transaction outcomes remain independent evidence.


### WF-43D exit / WF-47 entry — E-121 / V-061/062

Revisions root/conflicts imports use decoded shared syntax and package policy;
retired filenames, HTTP operation counts and private authorization-call spellings
are retired. A new real database transaction matrix checks delete, restore and
rollback authorization with exact current actor/incident/command/transaction,
no commit, rollback and unchanged durable source/envelope/projection/history,
intent and receipt state. Dependency completeness tests remain.
`remediation-wf43d-revisions-02` passed the boundary and all three successful route
controls but failed the new rollback case: its stub provider described no affected
records, so pre-admission correctly rejected it. The fixture now supplies the
actual target association. Only the denial row was repeated; it exited 0 at
`remediation-wf43d-denial-03`. Attempt `-01` stopped at owner ordering validation
before execution; ordering was corrected. Diagnosis used exact retained results.
All four module-local workstreams are complete; SQL replacement may begin.


### WF-43C exit / WF-43D entry — E-120 / V-061

Reporting root import inspection now uses the shared parser, retaining its exact
facade capability set, explicit nonempty scope and independent allowed/forbidden
fixtures. SQL policy stays in place until WF-47. The existing exact owner boundary
row exited 0 at `remediation-wf43c-reporting-01`. No permission was broadened and
Reporting specification adoption status remains unchanged.


### WF-43B exit / WF-43C entry — E-119 / V-061/069

Indicators uses shared parsed imports with its existing package capability map
and testsupport exclusion. Transport neutrality applies to all admission sources;
viewschema/strictjson remain admitted. Actual decode/error/replay assertions
replace the create.go token check. Exact mirror absence now rejects any nonzero
count, and seven transactional DDL cases each expose one forbidden column and
verify rejection, then roll back. Current create/transaction witnesses pass.
Four selected rows exited 0 at `remediation-wf43b-indicators-01`; no production
schema or public behavior changed. The local SQL mirror scanner remains pending
WF-47; the persisted source-prefix contract remains retained.


### WF-43A exit / WF-43B entry — E-118 / V-061/062

Entities now uses shared import syntax while retaining root capabilities and
merge-owned cross-owner ports, nonempty scopes and reorganization fixtures.
The merge collision route compares durable source/envelope/identifier/projection,
history and receipt state and intent count before/after rejection. A new real
transaction row exercises injected Timeline effect preparation/application with
exact transaction, state, actor, incident and time, proves source-write visibility,
successful history/intent/receipt publication and rollback on effect failure.
Existing host/identity merge success and lineage witnesses also pass.
The private method-prefix and Timeline interface-spelling assertions are retired;
route ownership guards and SQL scanner remain. New semantic routing and approved
transaction fixture metadata are authored together.
`make test-slice` exact four-row selection exited 0 at
`remediation-wf43a-entities-03`. Attempts `-01` and `-02` failed during admission
because the removed selector and new transaction-row approval were not yet
updated; neither executed tests. Those routing defects were corrected before
qualification. No production behavior changed.


### WF-46 exit / WF-43A entry — E-117 / V-066

Central import inspection now batches collected Go sources through the same Go
parser, validates complete ordered output, and fails closed on missing tools,
malformed sources/results and incomplete batches. Policy evaluation remains pure.
The private binary is built through existing Make artifact/cache machinery with
explicit work-graph dependencies and source inputs. Fixture coverage includes
all import declarations, build constraints, false comment/body imports and CLI
failure behavior. Inert-import coverage includes the adapter.
`make generate` reports pass at `remediation-wf46-generate-01` (native session
exit was unavailable after context recovery). Exact owner-boundaries row passed
with native exit 0 at `remediation-wf46-central-01`; the production boundary scan
passed with native exit 0 at `remediation-wf46-production-01`, exercising subsequent
artifact preparation. Final frozen-source checks will qualify generated inputs
and import initialization again. No fallback parser or new public command exists.


### WF-42 exit / WF-46 entry — E-116 / V-060/063

Added `internal/testutil/sourcecheck` with caller-rooted filesystem inspection and
Go-parser import decoding. Independent fixtures cover all declaration forms,
comments/body strings, build constraints, exclusions, nested paths, deterministic
results, empty scopes, malformed/read/walk failures and nonregular entries.
Registered row
`harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3`
and its semantic verification under the existing owner. No module parser removed.
`make generate`, the exact new `make test-slice`, and `make source-check CHECK=catalog`
all exited 0; roots `remediation-wf42-generate-01`, `remediation-wf42-parser-01`,
`remediation-wf42-catalog-01` under `.cartulary/test-results/`. Existing shared
support inventory covers the helper. Central integration may now begin.


### WF-45 exit / WF-42 entry — E-115 / V-065

Updated Testing Harness REQ-063 with compiler-backed import semantics, complete
results/fail-closed failures and bounded SQL-evidence meaning. Clarified the
Revisions decision against Core 01's destructive-command coordinator ownership:
injected participants retain projection/event ownership; history appenders and
source providers do not combine those effects. These are the user-approved
clarifications; Reporting's proposed status is unchanged. `make lint-markdown`
exited 0 at `.cartulary/test-results/remediation-wf45-docs-01`; direct Git whitespace
review passed. Default Markdown globs do not cover this tracker or the decision
file; those received direct review. No executable input reads owner prose.


### WF-41 exit / WF-45 entry — E-114 / V-059

All five baseline invocations passed with native exit 0. Exact selections are
in each retained invocation manifest; roots under `.cartulary/test-results/`:
`remediation-wf41-harness-01`, `remediation-wf41-entities-01`,
`remediation-wf41-indicators-01`, `remediation-wf41-reporting-01`, and
`remediation-wf41-revisions-01`. These include the original boundary rows,
merge store/route outcomes, Indicator envelope/transaction witnesses, and
Revisions command routes, bounded history, source-family paging, provider
catalogs and projection scope. Existing checks pass; F-060 remains a known
assertion weakness, not evidence of bad production schema. D-016–029 remain
bound to the inspected source and authored rows. RB-017 is resolved for this
baseline; RB-018 still gates each behavioral retirement. No code moved.


### Implementation entry — E-113 / WF-41

2026-10-27: user authorized the expanded proposed plan. Revalidated HEAD, current
instructions, source posture and test-operations procedure. Preserved the existing
tracker and diff externally. `make help ALL=1` and all five owner coverage guides
exited 0; outputs are under `/tmp/cartulary-testing-remediation-implementation/`.
Baseline execution is in progress. No old pass is claimed as current evidence.
F-057/F-058 deferrals are superseded by the selected workstreams; F-060 records the
exact-absence assertion defect. No frontend or public runtime change is planned.

### Superseded module-local active plan — preserved at implementation entry

<details>
<summary>Original active sections 1–9, before the approved scope expansion</summary>

### Testing Remediation 01: Simplify Module-Local Boundary Tests

#### 1. Scope and Source Posture

Document date: 2026-10-27. **Document-only plan: WF-41–44 are future work.**
The user selected module-local guards in Entities, Indicators, Reporting and
Revisions, then selected a shared Go-parser helper rather than a central-checker
cutover. The instruction to implement the proposed plan authorizes saving this
handoff. It does not authorize production/test changes or test execution: the
accepted proposal explicitly retained that document-only boundary.

Repository: `/home/jochi/code/cartulary`. Existing supplied continuation target
and canonical identity: `.` (repository directory; label `cartulary`). Controlling
output: `docs/handoffs/testing-remediation-01.md`, a regular single-link file under
non-symlinked parents. Artifact mode: `normal`. The available
[planning framework](cartulary_modular_refactor_planning_framework.md) supplies
planning doctrine; the refactor-tracker format supplies these twelve sections.
Neither supplies production authority. WF-36–40 are completed history archived
in Section 10; their authorization and passes do not transfer to WF-41–44.

The iteration consolidates Go import inspection, keeps semantic policies beside
their modules, and separates meaningful assurance from historical source layout.
A small helper replaces repeated mechanics; it does not become a policy engine.
Production APIs, dependency authority, HTTP/WebSocket contracts, replay bytes,
schemas, source ownership, authorization and transaction outcomes stay unchanged.
No compatibility wrappers, export censuses, replacement filename inventories,
new runner, production test hooks or owner-document inputs to executable checks
are admitted. Test reduction is justified by retained guarantees, not a count or
line-reduction target.

Excluded: broad SQL-checker redesign, central import-parser repair, other modules,
fixture/service lifecycle redesign, frontend work, production refactoring and
publication. Specialized mixed SQL, dispatch and publication assertions receive
explicit retained/deferred dispositions below; this iteration does not certify
or freeze all their implementation details as permanent requirements.

| Authority / status | Applicable requirement and treatment |
| --- | --- |
| [Core 00](../spec/00_document_set_status_and_precedence.md), current | Primary-owner allocation; REQ-00-071/073/074 recognize Revisions/Entities/Indicators topology decisions. Tests and typed contracts are downstream evidence, never requirement owners. No executable check may read, stat, hash or depend on Markdown. |
| [Core 01](../spec/01_architecture_storage_and_view_contracts.md), current | REQ-01-004/005 retain domain/application and platform boundaries; source/projection ownership and mutation/history semantics remain unchanged. Keep file-exact storage permissions and distinct behavioral witnesses. |
| Adopted [Entities](../decisions/entities-module-boundary.md), [Indicators](../decisions/indicators-module-boundary.md) and [Revisions](../decisions/revisions-module-boundary.md) decisions | Preserve capability/dependency direction, borrowed transactions, injected cross-owner effects, transport/application separation and source-owned provider coordination. Entities/Indicators explicitly distinguish capability ownership from exhaustive exports and file placement. |
| [Core 04](../spec/04_security_deployment_and_conformance.md), current | AC-529/558/560 retain meaningful Revisions/Entities/Indicators conformance. Authorization inside destructive transactions cannot be replaced by a check that a method name appears in source. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md), adopted/current | REQ-012–016: truthful identifiers and unique routing; REQ-062: no default private compatibility; REQ-063: central policy semantics and exact-file safety; REQ-066: authored routing before generation. No public harness interface change is planned. |
| Reporting owner boundary | Preserve the inspected facade/storage safeguards under Core ownership. The working Reporting NLSpec is `proposed-amendment` 1.4.0 over stated adopted base 1.3.0; do not treat its proposed clauses as newly adopted authority. No Reporting semantics or policy expansion is selected. |
| Domain / Core 05 | Existing vocabulary and owner navigation retained; no presentation, benchmark, release or publication claim. |

No owner amendment is proposed merely to legitimize a historical test. If later
characterization exposes a genuine production violation, record it separately
and block only dependent removal; do not add an allowance to make a test pass.

| Baseline item | Starting evidence / identity | Final comparison / limit |
| --- | --- | --- |
| Branch / commit | `main`; `1f6158c2778d14483c8a801705214d1815625422`; clean index/working tree, no nonignored untracked files | E-110 planning and E-111 save use the same source; E-112 records final comparison |
| Index | Raw SHA-256 `a9f65f99c74e0e853603ac4f0011dfb16d973ca2db68dd107575fe8b0256bc9c`; staged-entry SHA-256 `4fd3fc17d528b477ecf9297b50351dc36dcd25983a2d39cf7219730773182ca4` | Git reads use optional writes disabled; no staging/commit |
| Enumerated universe | 7,653 tracked/nonignored paths under `.`; content and lstat identity captured | Enumeration is not behavioral inspection; E-112 compares every outside-tracker entry |
| Tracker | 591,235 bytes; device/inode `2096:348826`; one link; SHA-256 `e880eb3104a95cbbb919503fdce5be23ed40b75758b9827e0aeade3f9503b095` | Unchanged since planning; rechecked immediately before write |
| External preservation material | `/tmp/cartulary-module-guards-plan-mclx85cr/baseline.json`, `original-handoff.md`, `inspection-accounting.json` | Planning/preservation material only; not product or release evidence |
| Ignored roots | Dependencies, caches, generated build/report outputs and retained-run directories excluded | No Make, service or test command executed; no exhaustive ignored-tree audit or historical-result requalification |

#### 2. Current-State Repository Inventory

The baseline accounts for every tracked/nonignored path. Direct inspection below
means the relevant implementation/assertion sections were opened, not that entire
modules were audited. `inspection-accounting.json` records 81 directly inspected entries at this
section-level depth and 7,572 metadata-only entries. Generated/vendor/binary files
receive no behavioral claim. Symlink entries are accounted for without following external targets.
Source has not changed between E-110 discovery and E-111 save. Text searches do
not prove absence of dynamic or external consumers.

| Path / group | Inspection depth and responsibility | Callers / disposition / dependencies | Risk / evidence |
| --- | --- | --- | --- |
| `internal/modules/entities/boundary_guard_test.go` | Direct: package import map, temporary import fixtures, merge import/SQL/planner checks, route/clipboard tokens, Timeline port spelling | Existing Entities root-boundary row; shared parsing, local policy, selective behavioral replacements | C-037–039; E-110; mixed assertions must not be deleted wholesale |
| Indicators `boundary_guard_test.go`, `admission/create_test.go`, `admission/create.go` | Direct: recursive package map, support exclusion, five-root-file transport inventory, admission source-token subtest and real decoding/error cases | Existing structure and admission rows; keep root capability map, share parsing, enforce transport restriction at admission package | C-037–039; E-110/111; view-schema use in admission is legitimate |
| `internal/modules/reporting/boundary_guard_test.go` | Direct: facade imports, raw-string import fixture and SQL fragment restrictions | Existing owner-boundaries row; migrate import mechanics only; retain SQL witness | C-037/038; E-110; central regex is not an equivalent successor |
| Revisions `application_boundary_test.go`, `history_components_test.go`, `rollback_components_test.go` | Direct: import bans, concrete type tokens, retired/required filenames, interface count, authorization spelling; mixed decomposition assertions | Existing boundary/history/rollback rows; retire pure inventories; preserve distinct persistence/dispatch/publication checks pending separate redesign | C-038–040; E-110/111 |
| Revisions `application_ports.go`, `command_service.go`, `command_service_test.go`, `delete_restore_store.go`, `rollback_coordinator.go`, `projection_scope_test.go`, `history_paging_test.go` | Direct at dependency, transaction, authorizer and reusable fixture seams | Existing explicit ports permit behavioral assurance without production exports/hooks; construction-only stubs are not execution evidence | C-040; E-110/111; denial must occur inside an actual command transaction |
| Entities `routes.go`, `merge/source_carry_forward.go`, `merge/merge_protected_set_test.go`, `mentions/ports.go`, `mentions/ports_test.go`; root merge unit/route tests | Direct at route binding, carry planning, injected effects, nil-dependency, collision and observable consequence sections | Retain route/storage checks; qualify actual merge/Timeline effects before deleting spelling checks | C-039; E-110; existing test presence is not a current pass |
| Indicators `source_state_contract_test.go`, `envelope_contract_migration_test.go`, `production_contract_test.go`, `composition_boundary_test.go`; Reporting `recovery_state.go` | Direct at semantic parity, envelope, composition and storage boundaries | Context and retained assurance; no blanket retirement of unrelated scanners or SQL assertions | C-039; E-110; no exhaustive behavior audit |
| Central backend checker, SQL classifier, policy and focused policy fixtures | Direct at Go extraction, SQL semantics, scope filtering and independent fixtures | Central import regex sees only first grouped block and double-quoted imports; defer repair and import cutover. SQL access modes are not interchangeable with import checks | F-057/058; E-110; no central evaluator/policy edits in this iteration |
| Four module test families and verification owners; `harness.command_surface` family, catalog, support inventory, helper ownership | Direct at affected rows, owner/support scopes and registration | Authored routing first; existing `internal/testutil` shared root already covers the proposed helper | C-041; E-110/111; no new support-root census |
| Applicable guidance, framework, owner clauses, generated policy, Make recipes/admission, Markdown configuration | Direct at scope, generated ownership and command effects | Human planning inputs only; public Make commands are future implementation interfaces | E-110/111; Markdown check omitted for explicit side-effect conflict |
| Proposed `internal/testutil/sourcecheck` | Absent; future authored test-only helper, not a current inspected implementation | Used by selected module tests; mechanics registered once under existing harness owner | S-042; helper characterization is future V-060 |

Next discovery boundary: revalidate these sources and exact rows at WF-41 entry;
resolve the specified characterization gaps before removals. Broader repository,
SQL-parser and specialized composition-scanner audits remain outside this bounded
plan rather than being silently counted as completed.

#### 3. Module Boundary Diagnosis

The selected tests mix source-inspection mechanics, module dependency policy,
behavioral assurance and historical layout snapshots. They are not a new runtime
subsystem. Share the small mechanism with actual reuse; leave ownership decisions
with the modules and transaction behavior with executable domain tests.

| Responsibility | Current location | Keep / move / split / defer | Finding / contract |
| --- | --- | --- | --- |
| Go import extraction and file reading | Repeated Go parsers and source-text readers | Move to one test-only Go helper; preserve syntax fidelity | F-054; C-037 |
| Allowed capabilities and prohibited dependencies | Module-local guards | Keep local, operating on shared inspection results | F-054/056; C-038 |
| File inventories, retired names and interface method counts | Revisions mixed structural guards | Retire standalone incidental assertions; no replacement inventory | F-055; C-039 |
| Transaction authorization and merge/Timeline effects | Source spelling plus domain tests | Split; qualify behavioral witnesses before retiring spelling assertions | F-056; C-039/040 |
| Generic SQL/source-dispatch/publication checks | Central policy and specialized local guards | Keep existing meaningful assurance; defer broader redesign | F-058; C-039 |
| Central Go import extraction | Backend checker regex | Defer; no cutover to weaker extraction | F-057; C-037 |
| Selection, support ownership and evidence attribution | Existing authored harness owners | Keep; update real selectors/claims atomically with movement | F-059; C-041 |

#### 4. Public Contract and Behavior Freeze Map

Preserve owner-aligned guarantees, not every observed test outcome. No HTTP,
WebSocket, entity/view schema, saved view, generated product contract, storage,
revision, authorization or replay interface is changed. Specific test-name/file
arrangements are not production contracts. Distinct domain integration witnesses
are not replaced by shared source inspection.

| Contract | Adopted owner / observed posture | Required treatment and assertion evidence | Slice / validation / gate |
| --- | --- | --- | --- |
| C-037 — Faithful import inspection | Harness REQ-062/063 and module ownership; local Go parsers exceed central regex syntax coverage | Decode real Go imports; aliases, raw strings and repeated groups work; comments/fixture strings do not count. Parse/read failures fail closed; tests remain excluded and build tags do not select host-specific subsets | S-042/043; V-059–061; RB-017 |
| C-038 — Module dependency authority | Core 01 REQ-01-004/005 and adopted module decisions | Preserve current allowed package capabilities, forbidden concrete/assembly dependencies and support-scope separation. No file inventory or broader permission masquerades as simplification | S-043; V-061; D-016/019/022/024 |
| C-039 — Source, route and effect ownership | Entities/Indicators/Revisions decisions; Core source/projection/history ownership | Keep route ownership, read/write boundaries, merge atomicity and caller-owned effects. Separate incidental private names from observable consequences. Retain real admission Field/ReasonCode and unchanged Indicator envelope/source assertions | S-043; V-061/062; RB-018 |
| C-040 — Authorization inside destructive transactions | Core security and Revisions application boundary; current source only searches for an authorization call | Delete/restore/rollback use the current actor, incident and command in the transaction. Denial returns the classified error, commits no mutation/history/receipt/publication and rolls back. Existing route precedence and successful paths remain | S-043; V-062; RB-018 |
| C-041 — Truthful ownership and routing | Harness REQ-012–016/066; mixed current rows contain several independent assertions | Keep rows with surviving semantics; add genuine helper/security rows under existing owners; register once, classify real service needs, regenerate authored changes through Make. No aliases or false successor claims | S-042–044; V-063/064 |

The shared helper is a test-support interface, not a production API. Public Make
commands and evidence schemas stay unchanged. Any newly required service-backed
witness must declare its real service/isolation metadata rather than inheriting
`unit` from a name or its predecessor row.

#### 5. Coupling and Boundary Findings

| Finding / classification | Evidence and consequence | Required action / benefit | Contracts / slices / validation |
| --- | --- | --- | --- |
| F-054 / `must_fix` for S-042/043 | Three local import parsers plus Revisions substring scans repeat mechanics; fixtures independently create files or inspect text | One small Go-native inspection helper with in-memory fixtures; local policies remain readable and syntax handling consistent | C-037/038; S-042/043; V-060/061 |
| F-055 / `must_fix` for S-043 | Revisions requires a component-file list, forbids retired filenames and counts four interface methods in a source substring | Delete pure layout/count assertions without aliases or replacements; retain actual methods, routes and meaningful decomposition assurance | C-039; S-043; V-061/062 |
| F-056 / `must_fix` for S-043 | Entities uses `plan*`/mutation-name prefixes and port spelling; Revisions searches for `AuthorizeCommandTx`; Indicator transport policy names five root files and separately checks admission source text | Move substantive assurance to typed capabilities and exercised outcomes; enforce actual admission-package imports. Do not infer a whole Indicator-root transport ban from a partial old list | C-038–040; S-043; V-061/062; RB-018 |
| F-057 / `defer` | Central `extractGoImports` recognizes double-quoted imports and only the first grouped block; Reporting already has a raw-string rejection example | Keep the selected Go-parser approach. Central parser repair is separate work; no compatibility/fallback route or duplicated central policy added here | C-037; explicit exclusion, not an S-042 dependency |
| F-058 / `defer` | Local SQL, source-dispatch and publication checks mix distinct guarantees with implementation-sensitive inspection; central relations do not cover every local invariant | Preserve distinct witnesses and document residual limitations. No generic SQL framework, text-rule migration, or blanket deletion of mixed tests in this iteration | C-039; D-018/021/023/027; separate future discovery |
| F-059 / `must_fix` for S-042–044 | Existing rows combine imports with admission, storage, composition or history assertions | Update selectors and descriptions to real surviving evidence. Separate new transaction evidence from static/unit rows; use existing catalog owner | C-041; S-042–044; V-063/064 |

P denotes a production requirement, H a necessary harness invariant and I an
incidental arrangement. A P/H removal needs a qualified surviving witness. An I
retirement needs a reason, not a manufactured successor. The following ledger is
the selected assertion boundary; WF-41 confirms exact source locations and row
bindings without reopening the chosen design.

| Disposition ID / assertion | Class and decision | Named survivor or removal reason / gate |
| --- | --- | --- |
| D-016 Entities root import allowlist and merge-to-Assessments prohibition | P/H; consolidate mechanics, preserve policy | Shared helper plus unchanged module-local capability checks; concrete allowed/forbidden paths including renamed files; V-060/061 |
| D-017 Entities root Workbook-route and clipboard-plan prohibitions | P; preserve | Existing route/storage-independent checks remain; no route ownership redesign or new caller inventory |
| D-018 Entities merge SQL check and nested carry-forward planner-name scan | P SQL / I name heuristic; retain SQL, replace nested spelling assurance | Preserve direct mention/projection write bans. Qualify host/identity merge success and collision/no-change behavior before dropping `plan*` and `Sync/Update/Insert/Delete/Exec` spelling tests; no normative private phase/name rule; V-062 |
| D-019 Entities `TimelineEffectsPort` source spelling and retired interface signatures | P effect boundary / I spelling; replace | Preserve typed command/fact ports and module import policy. Add focused injected-effect success/failure witness under `mentions`, using the caller transaction and existing public commands; errors remain classified and state rolls back. Retain constructor nil/typed-nil tests; V-062 |
| D-020 Indicators production imports and reorganization examples | P/H; consolidate | Retain per-package capability map, owner-local child permissions, app-assembly prohibition and explicit `testsupport` exclusion. Do not copy central support permissions into this policy |
| D-021 Indicators source-prefix and SQL envelope-mirror checks | P; preserve, mechanism redesign deferred | Keep `TestIndicatorsDoNotUseEntitiesSourcePrefixes` and `TestIndicatorSourceSQLDoesNotUseEnvelopeMirrors`, plus actual source-state/envelope witnesses. They are not parser-duplication retirements |
| D-022 Indicator five-root-file transport list and admission `create.go` token subtest | I file/token arrangement / P admission boundary; replace | One admission-package import guard rejects `net/http` and platform `httpapi`; view-schema and strict-JSON dependencies remain permitted. Retain actual `DecodeCreateRequest` cases and closed Field/ReasonCode results. Root imports are not globally transport-free; no production relocation or new ban; V-061 |
| D-023 Reporting import and SQL tests | P/H; consolidate imports only | Shared Go parser and existing facade policy replace local import helper/duplicate syntax cases. Preserve `TestReportingProductionDoesNotReadOwnerTablesDirectly`; SQL assurance remains distinct |
| D-024 Revisions root/conflicts import guards and concrete dependency text checks | P; consolidate semantic dependency restrictions | Use shared parser with root package bans on HTTP/auth/pagination adapters and concrete Postgres/Records roots; conflicts retain existing transport/platform bans. Existing command dependency/port construction tests remain. No source filename is a dependency boundary; V-061 |
| D-025 Revisions retired root files, required rollback files, retired `store.go`/`rollback_store.go`, HTTP interface operation counts | I; retire without replacement | Normal compilation and retained route/behavior tests exercise real interfaces. Remove standalone presence/absence/count assertions; no alternate filename map or fixed operation census |
| D-026 Revisions `.AuthorizeCommandTx(` source searches | P security / I source evidence; replace | New transaction authorization-denial matrix for soft delete, restore and rollback; existing route precedence and successful transaction witnesses remain; V-062 must pass before removal |
| D-027 Remaining history/rollback decomposition SQL, indexed-association, provider-dispatch and publication assertions | P; preserve/defer inspection redesign | Keep semantic assertions in mixed tests; consolidate duplicated transport-import check into D-024. Removing D-025 does not authorize deleting these bodies. Remaining source-file coupling is disclosed as residual debt, not an adopted compatibility contract |
| D-028 Adjacent Indicator composition/provider and entity mutation-admission tests | P; retain | No expansion into their independent AST construction scanners, source catalog parity, actor admission or borrowed-transaction assurance |
| D-029 Catalog and support ownership | H; preserve/update | Exact live selectors, single registration, genuine new evidence rows and real service metadata; no empty-success rows, selector aliases or synthetic successor crosswalks; V-063 |

#### 6. Refactor Workstreams

The prerequisite column is authoritative; the graph is serial and acyclic. Save
each exit here before its successor begins. All four workstreams are included;
there is no parallel agent work and no implicit production-fix branch.

| Workflow / slice | Class / prerequisites | Goal and likely changes | Validation / checkpoint / binary exit |
| --- | --- | --- | --- |
| WF-41 / S-041 | Root; later implementation authorization and source revalidation | Confirm D-016–029 locations, authority and exact rows; execute current baseline; record gaps separately from passes | V-059; ledger and baseline saved, with no unexamined selected assertion |
| WF-42 / S-042 | Chain; WF-41 exit | Introduce shared source inspection and independent mechanics tests; register new support evidence through existing owner | V-060/063; helper qualification passes before consumers move |
| WF-43 / S-043 | Chain; WF-42 exit | Migrate four module import guards; establish V-062 successors; remove only admitted duplicate/incidental assertions and update affected selectors | V-061/062/063; local policies and retained domain guarantees qualify; each removal has its ledger disposition |
| WF-44 / S-044 | Chain; WF-43 exit | Close authored/generated routing; finalize, freeze and qualify missing obligations; finish handoff | V-063/064; compatible final portfolio, native exits, retained validity and residual exclusions recorded |

A newly exposed genuine production defect blocks its dependent removal. Do not
relax module ownership, add exceptions, preserve a broken alternate reader or
change a normative requirement just to close this iteration.

#### 7. Proposed Refactor Slice Plan

Implementation authorization: **not granted by this planning invocation**.
Readiness below is separate from permission and from eventual completion.

| Slice / kind | Findings / contracts | Entry and readiness | Change / validation / rollback / binary completion |
| --- | --- | --- | --- |
| S-041 / characterization | F-054–059; C-037–041 | Ready for authorized entry; E-110/111 static discovery | Revalidate ledger and execute V-059. No code movement; baseline/failed evidence retained. Exit only with current classification and exact results |
| S-042 / structural preservation | F-054/059; C-037/041 | Blocked: S-041 and RB-017 current characterization | Add helper and register qualified tests; V-060/063. Revert helper/tests/authored routing together if needed. Exit before deleting any local parser |
| S-043 / test assurance correction | F-054–056/059; C-037–041 | Blocked: S-042 exit and RB-018 behavioral successor qualification | Migrate mechanics and apply D-016–029; V-061–063. Test-policy changes require later authorization; production unchanged. Roll back affected tests and authored selectors together, not a shim |
| S-044 / qualification | F-059; C-041 | Blocked: S-043 exit | V-063/064 final portfolio and handoff. Regenerate from authored inputs; retain failed artifacts and revise compatibility after repairs. Exit only on final-source evidence |

##### Shared helper and module policy

Create `internal/testutil/sourcecheck` with one small API:

```go
type FileImports struct {
    Path    string
    Imports []string
}

func ReadImports(source fs.FS) ([]FileImports, error)
```

The filesystem is explicitly rooted by its caller; normal consumers use an
`os.DirFS` for their module and test fixtures use `fstest.MapFS`. Walk regular
non-`_test.go` Go files recursively, parse with `go/parser` in `ImportsOnly` mode,
and decode every import using `strconv.Unquote`. Return normalized relative
paths in deterministic order and retain import declaration order. Non-Go files
are not opened or statted; directory enumeration is not document inspection.
Do not follow symlinked Go files/directories; report selected nonregular Go files,
read failures and malformed imports with the relative path. Do not silently
return partial success. Empty valid scopes return an empty collection; actual
module guards assert that their selected production package is nonempty.

The helper owns no package policy, support inventory, glob language, build-tag
resolver, repository-root discovery, subprocess, cache or assertion framework.
Module tests filter results by exact package directory/subtree before applying
local policies. All existing source build tags remain visible, independent of
host GOOS/GOARCH. It must remain concurrency-safe without process/global state.
The existing shared `internal/testutil` support-root classification already
covers it; do not add a redundant per-helper support registry.

Keep capability maps local and ordinary. Entities checks its root and the
specific merge dependency; Reporting checks its root; Revisions checks its root
and `conflicts`; Indicators checks its existing recursive package map excluding
`testsupport`, with a separate admission-package transport rule. Root/child scope
must not drift during extraction. Parse fixture syntax once in shared tests;
module fixtures test real allowed/forbidden dependency decisions independently of
the policy implementation. A harmless rename/split within an admitted package
must pass without editing a filename inventory.

##### Behavioral replacements and limits

Use existing test-local seams; never add production exports or generic fault
injection to make a guard removable.

- **Merge:** retain `TestExplicitEntityMerge_Unit` and
  `TestExplicitMergeRoute_Integration`. Qualify host and identity carry-forward,
  raw mention preservation, survivor/loser lineage, projection/Timeline effects
  and failure atomicity. Extend collision/error cases to compare source,
  envelope, history, projection and intent/receipt state before/after where the
  existing case lacks that assertion. Retire planner-name heuristics only after
  this meaningful outcome coverage passes; do not require private functions to
  keep a `plan` prefix or introduce a new normative planning phase.
- **Mentions/Timeline:** add focused package-local tests around the existing
  caller-transaction command/effect port. An injected effect records the same
  transaction and command/fact payload; success produces the expected effect,
  and a sentinel failure remains classifiable and cannot leave committed
  mention/history/projection state. Use the existing `pgtest` fixture where
  persistence is asserted. Retain nil/typed-nil composition tests and current
  import restrictions; this is not a port-method census.
- **Revisions authorization:** add a three-command matrix with the existing
  `TransactionRunner`, `CommandAuthorizer`, envelope/idempotency/provider ports
  and test-local dependency builders. Use a real PostgreSQL fixture for durable
  state assertions and a delegating transaction observer for identity and
  commit/rollback observation. Seed a valid target and use a fresh request so
  each operation reaches authorization rather than replay or earlier rejection.
  Inject a denying authorizer; assert incident, actor, command kind and active
  transaction identity; require the classified denial, rollback and no commit,
  and unchanged source/envelope/history/idempotency/publication state. Keep the
  existing route authorization/CSRF/precondition tests and successful transaction
  cases as independent evidence. Constructor stubs that return nil transactions
  are not execution fixtures. Do not add SQL-string expectation transcripts.

Standalone Revisions file/count assertions can retire without a fake replacement.
Mixed SQL/index/dispatch/publication assertions remain until separately designed
successors exist. This explicit residual debt does not make their filenames a
specification requirement or close the broader testing-remediation effort.

##### Routing decisions

Retain existing module row identities where substantive assertions survive.
Remove only deleted selectors; keep accurate claims for mixed rows. Register
helper mechanics once under `harness.command_surface`, using a semantic
source-import-inspection verification/row authored through the existing catalog
workflow. Register new Revisions transaction authorization and Mentions injected
effect evidence under their existing module owners, with distinct semantic
security/effect claims and the actual PostgreSQL fixture/isolation requirements.
New identities are authored via `make author-test-row-id`; save the resulting
literal IDs before selecting those rows. This is normal identity authoring, not
an unresolved choice of owner or behavior. No new owner, runner, task or evidence
schema is required. Do not reclassify new integration work as unit to preserve an
old aggregate or fabricate a successor for a purely incidental assertion.

#### 8. Validation Plan

**No tests or Make commands ran in this document session.** Commands below are
future implementation obligations, selected from inspected public recipes and
authored rows. `make task-guide` must revalidate selections at authorized entry;
its output is discovery, not passing evidence. Run from the repository root with
pinned tools, canonical runner context and fresh explicit run IDs.

| Validation / stage | Exact public command or bounded selection | Expected assurance / gate |
| --- | --- | --- |
| V-058 / current document save | Direct heading/table/link/ID/dependency/archive review; `git --no-optional-locks diff --check -- docs/handoffs/testing-remediation-01.md`; full baseline/index comparison | Only this handoff changes; history preserved; no product pass inferred; E-112 |
| V-059 / future pre-move | `make help ALL=1`; `make task-guide OWNER=<owner> DETAIL=coverage JSON=1` for the four modules and `harness.command_surface`; then `make test-slice OWNER=<owner> ROWS=<exact rows below>` | Current baseline for selected assertions and required surviving witnesses; RB-017. Existing defects are recorded distinctly, not normalized into expected behavior |
| V-060 / helper successor | `make test-slice OWNER=harness.command_surface ROWS=<newly authored source-import-inspection row>` | Real imports: aliases, blank/dot imports, raw/escaped strings, repeated groups, comments and strings containing fake imports; malformed import/read failure, nested packages, non-Go/test exclusion, build-tag visibility, nonregular entries and deterministic output. One focused fixture matrix, not a parser reimplementation |
| V-061 / module cutover | `make test-slice OWNER=<module> ROWS=<its boundary/admission/decomposition rows below>` | Local permitted/prohibited capabilities and package scopes; rename/split positive case; no root-wide Indicator HTTP ban; no leaked support permission or dead private helper |
| V-062 / semantic successors | Existing Entities merge rows and Revisions route/transaction rows below, plus new Mentions effect and Revisions transaction-authorization rows via `make test-slice` | Actual effects, errors, durable no-change/rollback and security precedence qualify before D-018/019/026 removals; RB-018 |
| V-063 / routing and source | `make generate`; `make source-check CHECK=generated`; `make source-check CHECK=schema`; `make source-check CHECK=catalog`; `make source-check CHECK=identity`; `make lint CHECK=go`; `make lint CHECK=backend-boundary` | Authored/generated consistency, unique selectors, actual fixture requirements, runtime/test-support separation and unaffected central policy. Generation only during separately authorized implementation |
| V-064 / final qualification | `make agent-finalize`, review and freeze source; execute only missing compatible obligations from V-060–063; `make harness-observability-check RESULTS_DIR=<exact-successful-root> CHECK=all` for each qualifying root; final documentation review | Native exits, source/toolchain identity, selected rows, cleanup, retained validity and exclusions recorded; no historical CI pass substitutes |

Exact existing selectors for V-059/061/062 (comma-join only rows needed by that
invocation; use separate owner selections):

| Owner | Existing row IDs / purpose |
| --- | --- |
| `module.entities` | `module.entities.unit.root_boundary_guards_e292ec1361` — imports and mixed guards; `module.entities.store.explicit_merge_store_behavior_repoints_live_ment_0560122310` and `module.entities.integration.the_explicit_merge_route_repoints_live_fan_out_p_0ec76e8044` — merge outcomes |
| `module.indicators` | `module.indicators.structure.indicator_application_requires_complete_composition_6ecc4edf41` — structure/composition; `module.indicators.unit.create_admission_preserves_wire_and_replay_7ef2ad611c` — actual admission |
| `module.reporting` | `module.reporting.unit.owner_boundaries` — facade imports and retained SQL check |
| `module.revisions` | `module.revisions.support_unit.http_application_boundary_4e7a9c2d12`; `module.revisions.support_unit.history_materialization_attribution_and_page_7b4d2e9c10`; `module.revisions.rollback.decomposition_8c5e3f2a10` — mixed architecture and retained semantics |
| `module.revisions` | `module.revisions.store.covers_rollback_route_admission_derived_rollback_c6c889199c`; `module.revisions.store.covers_the_first_class_record_soft_delete_route_30cb39b83e`; `module.revisions.store.covers_the_restore_route_contract_reviewer_admin_a2df5bc94d` — existing command/route outcomes |
| `harness.command_surface` | New helper row is authored in WF-42 and selected literally after catalog admission; no invented current selector or direct `go test`/Node invocation |

Choose a coverage portfolio, not repeated complete owners. New database tests
must use truthful fixture metadata; several existing tests named `_Unit` already
have integration/service routing. Keep existing service settings until authored
routing deliberately changes. Preserve distinct route, durable-state and static
boundaries; test count is not an equivalence criterion.

Finalizer retained-run maintenance requires a compatible successful full warm
`check` root. Otherwise report it skipped because `RESULTS_DIR` is unset; do not
launch a redundant full check for maintenance. Full CI, browser/visual/a11y,
benchmarks and release qualification are not automatic additions. Expand only
for a specific missing obligation, and reassess compatibility after each repair.

| Current session command / side effects | Admission and result |
| --- | --- |
| Targeted source reads, Git inspection with optional writes disabled, external content/identity snapshot | Executed; E-110/111. Planning evidence only; no services or test evaluation |
| `make lint-markdown` | Skipped: tracker skill permits only the admitted file; Make admission can write `.git/cartulary-workspace.lock`, and the recipe can prepare dependencies/write cache and run artifacts. No run-then-restore workaround. Configured globs omit this tracker, so a pass would not validate its contents |
| `make help ALL=1`, `make task-guide` | Deferred to implementation entry because Make admission does not meet this session's sole-file write boundary; recipes and authored rows inspected instead |
| Tests, `make format`, generation, finalizer and product checks | Not run: document-only user scope. All V-059–064 implementation checks remain future obligations |

#### 9. Top-Level Work Tracker

| ID | Work item / workstream | Status | Depends on / evidence | Exit condition |
| --- | --- | --- | --- | --- |
| T-047 | Save selected plan / document-only | DONE | E-110–112; V-058 | Concrete plan saved; direct validation and full sole-file/archive comparison passed |
| T-048 | Current characterization / WF-41 | TODO | Later implementation authorization; S-041; V-059 | Exact ledger and baseline recorded, no historical pass substituted |
| T-049 | Shared import inspection / WF-42 | BLOCKED | T-048; RB-017; S-042 | V-060/063 pass before old parser removal |
| T-050 | Module cutover and semantic successors / WF-43 | BLOCKED | T-049; RB-018; S-043 | D-016–029 dispositions implemented and qualified; no production change |
| T-051 | Final qualification and handoff / WF-44 | BLOCKED | T-050; S-044 | V-063/064 final-source portfolio and truthful exclusions saved |


</details>


### Module-local guard planning and save — E-110–E-112

#### Scope and authority

| Date / evidence | Decision / current state | Next boundary |
| --- | --- | --- |
| 2026-10-27 / E-110 | User chose four module-local guards and a shared Go helper; central cutover rejected because it weakens import parsing | WF-41–44 define later implementation, not this save |
| 2026-10-27 / E-111 | Same source/fingerprint revalidated; only this handoff admitted for write; prior current plan archived | Complete V-058; no code, tests, specifications, routing or generation executed |

#### Backend boundary

| Evidence | Inspected decision | Limit |
| --- | --- | --- |
| E-110/111 | Shared Go-native mechanics; local capability policy; specified behavioral replacements for mixed assertions | Production code unchanged; broad SQL/composition scans deferred |

#### Frontend boundary

Not applicable: no frontend, rendered UI, selector or browser change is planned.

#### Contract and codegen

| Evidence | Decision | Next boundary |
| --- | --- | --- |
| E-110/111 | No production or public harness interface amendment; existing shared support-root ownership suffices | Future authored catalog rows/claims first, then Make generation and schema/catalog/identity checks |

#### Tests and harness

| Evidence | Commands / result | Applicability |
| --- | --- | --- |
| E-110/111 | Source assertions, command recipes and exact family rows inspected; no tests or Make invocation | V-059–064 are future validation obligations, not passes |
| E-112 | Direct review and Git whitespace check passed: twelve sections, 22 active tables, nine local links and fourteen dispositions; history preserved | V-058 only; no product or release claim |

#### Security and authorization

| Evidence | Required distinction | Next boundary |
| --- | --- | --- |
| E-110/111 | A call-name substring is not proof of in-transaction authorization; actual denial/rollback requires separate evidence | RB-018 gates removal, with explicit actor/incident/command/transaction and durable-state assertions |

#### Open risks and next session

| Evidence | Current risk / readiness | Next action |
| --- | --- | --- |
| E-110/111 | No current characterization; mixed assertions can conceal distinct guarantees; central parser is weaker | Start WF-41 only after later authorization; keep deferred witnesses and production requirements, not accidental test arrangements |
| E-111 | Plan is bounded, not a suite-wide audit; Reporting proposal not promoted into authority | Revalidate only affected owners/source; no unrelated adoption or production repair |

#### Evidence register

| ID | Source / baseline | Observation and limit |
| --- | --- | --- |
| E-110 | Prior read-only planning on `1f6158c2778d14483c8a801705214d1815625422`; targeted file reads/searches, JSON policy/row inspection, owner clauses and two user selections | F-054–059 and D-016–029 grounded; no tests or mutations; source bodies/assertions inspected at stated depth |
| E-111 | Same HEAD; clean save baseline; `/tmp/cartulary-module-guards-plan-mclx85cr/` | 7,653 entries captured, index/HEAD/tracker admission revalidated; focused Revisions fixture/port and lint-path inspection refines the accepted document plan; sole-file save authorized |
| E-112 | V-058 final review; `/tmp/cartulary-module-guards-plan-mclx85cr/document-preservation-review.json` | Direct structure/table/link/ledger and Git whitespace checks passed. Prior Section 10 preserved verbatim and former active sections preserved with heading demotion. All 7,652 outside-tracker entries, file identities, index/staged entries and HEAD unchanged; only this handoff content changed. Ignored roots not exhaustively audited; no canonical product verification |

### Completed central-checker iteration — historical active-plan archive

The following preserves WF-36–40's former active sections. Their authorizations,
statuses, source posture and validation are historical. Only heading levels were
demoted for nesting. The preceding current sections supersede their active-scope
statements. The prior Section 10 log follows this archive verbatim.

<details>
<summary>Completed WF-36–40 active sections and completion criteria</summary>

#### Testing Remediation 01: Simplify Central Backend Boundary Assurance

##### 1. Scope and Source Posture

Document date: 2026-10-27. **Implementation complete: WF-36–40 are DONE.**
The user requested implementation of the proposed five-workstream plan.
Specification, checker, policy, schema, tests, routing and generated projections
are authorized within that boundary; production behavior remains excluded. WF-31–35 and
E-079–097 are completed history preserved in Section 10; their authorizations
and passes do not qualify this iteration.

Repository: `/home/jochi/code/cartulary`. Supplied continuation target and existing
canonical identity: `.` (repository directory; label `cartulary`). Primary seam:
central backend boundary enforcement and its assurance. Controlling tracker:
`docs/handoffs/testing-remediation-01.md`, a regular single-link file under the
non-symlinked `docs/handoffs` directory. Artifact mode: `normal`. The available
[planning framework](cartulary_modular_refactor_planning_framework.md) supplies
planning doctrine; the refactor-tracker format supplies these twelve sections.
Neither supplies production authority.

The user selected **central checker**, excluding broader module-local guard
consolidation. The implementation separates scanning from self-tests, corrects
policy admission and matching, and removes redundant assertions. Preserve substantive
production ownership guarantees. Do not preserve malformed rules, fake source
paths, fixed counts, or test organization for compatibility. No new runner,
compatibility mode, generic rule framework, dual reader, or export census is
planned. Production APIs, wire/event behavior, schemas, authorization, replay,
transaction effects, and fixture APIs are outside the change boundary.

Also excluded: frontend checks, service/process lifecycle redesign, module-local
scanner consolidation, general parser replacement, benchmarks and publication.
Other central policies remain unchanged unless the explicitly bounded correction
below affects their interpretation. This is not a mandate to retire every
historical rule or to broaden source traversal.

| Authority / status | Governing requirement and treatment |
| --- | --- |
| [Core 00](../spec/00_document_set_status_and_precedence.md), current | Primary owners and adopted decisions govern behavior; machine contracts and tests are projections. No executable evidence may read, stat, hash, or depend on Markdown. |
| [Core 01](../spec/01_architecture_storage_and_view_contracts.md), current | REQ-01-004/005 and REQ-01-649 preserve module boundaries, Records-owned envelope writes and exact approved read shapes. Section 8 and REQ-01-621/622/626 preserve source/projection separation and code-backed provider ownership. Section 12.3 governs Timeline portability/provenance, not a historical two-table test assumption. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md), adopted/current | REQ-012–016 govern truthful identifiers, selection and routing. REQ-062/063 govern private helper contraction, explicit import/token matching, v4 policy validation, exact-file-set safety and unchanged public checker output. S-038 implements the user-approved bounded clarification of ordinary path patterns in the owner document; E-105 records its adoption and qualification. |
| [Projections decision](../decisions/projections-module-boundary.md), adopted; Core 04 AC-539 | Preserve provider/storage ownership and production assembly. Semantic table-set equality is meaningful; the literal number ten is not independent assurance. |
| [Domain](../domain.md); Core 04; [Core 05](../spec/05_claim_publication_and_benchmark_reproducibility.md) | Vocabulary and security/conformance owners remain unchanged. No UI, benchmark, release or publication claim is made. |

The following table preserves the **planning-session baseline**. E-101 records
the implementation baseline and E-103 onward record actual workstream exits.

| Baseline item | Starting evidence / identity | Final comparison / limits |
| --- | --- | --- |
| Branch / commit | `main`; `3bd81c1b2dc9fb8e8424fa50ce0633f982840702`; clean index and working tree | E-098's inspected source remains current; E-100 records save verification |
| Index | Raw SHA-256 `274413c40a349611b90f8c28c610242ad799d52e8a7ac01d627177443327e293`; staged-entry SHA-256 `e29f5c6547de78b0fd0e9be9d378e26da79bb2d32387a053cee3548bb8c3a48e` | Use Git with optional writes disabled; no staging or commit |
| Enumerated universe | 7,646 tracked/nonignored entries; no nonignored untracked files | Per-entry content and lstat identity retained, including generated/binary entries; enumeration is not behavioral inspection |
| Tracker | 530,644 bytes; device/inode `2096:348826`; one link; SHA-256 `d243dfb60b38768f5c58cf5cbf256b96a110ea605881884f396b693b4d900421` | Recheck immediately before writing; existing historical Section 10 retained verbatim |
| External comparison material | `/tmp/cartulary-boundary-plan-sriebz8d/baseline.json` and `original-handoff.md` | Preservation records only, not product evidence; no secrets copied into this tracker |
| Ignored outputs | Caches, dependency installs, build outputs and retained-run directories excluded from the file universe | No Make/service/test command executed; no exhaustive ignored-tree fingerprint or historical-run requalification claimed |

##### 2. Current-State Repository Inventory

E-098 directly inspected the checker, relevant policy definitions, its callers,
existing assurance and adopted owner clauses. E-099 revalidates the unchanged
baseline and admission before saving. Entries not listed below receive
**metadata-only** accounting in the external baseline. Generated/vendor/binary
entries receive no claim of behavioral review. Symlinks are accounted for without
following external targets. No production source movement is planned.

The external `inspection-accounting.json` records all 7,646 entries: 46 directly
inspected at the section-level depth described here and 7,600 metadata-only.
Targeted search hits outside those 46 are not promoted to direct inspection.
These counts describe planning coverage, not an executable acceptance threshold.

| Path / group | Inspection depth; responsibility / dependencies | Caller, owner and planned disposition | Evidence / risk |
| --- | --- | --- | --- |
| `tools/harness/static-analysis/backend-module-boundary-check-cli.mjs` | Direct inspection of scanner, normalization, matching, all embedded fixture groups, main and output; 2,068 lines | Work-graph policy operation and Records architecture row; split CLI/evaluator/tests within `static_analysis` | E-098; F-050–053; C-032–036; high assurance risk |
| `tools/backend_module_boundaries.json` and `tools/schemas/cartulary.backend_module_boundaries.v4.schema.json` | Policy categories enumerated; affected patterns, malformed rules, schema and relevant permissions directly inspected | Authored inputs consumed by checker and assembly tests; repair bounded projections, preserve unrelated rule meaning | E-098; F-051/052; full per-assertion disposition is WF-36's specified entry deliverable |
| `sql-relations.mjs`, `exact-file-sets.mjs`, `support-inventory-profiles.mjs` under the same static-analysis directory | Parser, exact-set validation/resolution and support inventory boundaries inspected | Existing dependencies retained; no generic replacement or broader scan roots | C-032/034/036; parser and path safety must survive movement |
| Static-analysis `tests/test-exact-file-sets.mjs`; `tools/harness/tests/test-harness-boundary-contracts.mjs`, `contract-initialization-cases.mjs` | Independent exact-set fixtures, registration and inert-import/failure-isolation witnesses inspected | Reuse existing boundary-contract operation; consolidate only overlapping assurance | F-050/053; C-035/036 |
| `internal/app/projectionassembly/catalog_manifest_test.go`, `source_ownership_test.go` | Descriptor manifest and exact equality of boundary, descriptor, migration and Recovery table sets inspected | Keep existing semantic owner witnesses; they supersede a literal count assertion | F-053; C-032; no assembly source changes |
| Timeline `history_projection.go`, `import_create.go`, `incident_bundle_source_port.go`, `incident_bundle_subtype_presence.go`, `provenance.go`, `incident_bundle_portability.go` | Relevant imports and SQL/provenance paths inspected; unrelated behavior not audited | Four narrow contracts plus provenance and approved Records joins explain stale central policy | F-051; C-032/033; no production corrections authorized |
| Recovery import sites; affected Incidents, Links, Assessments, Workbook and Tasks/Decisions paths | Targeted import/declaration/token searches; search results are not complete behavior inspection | Bound newly matched rules; no unrelated module-local guard changes | F-051/052; characterize all affected policies before enforcement changes |
| Authored work graph, task surface, test families for `harness.command_surface`, `module.records`, `module.projections`; Records verification owner | Relevant operation, row and verification entries directly inspected | Retain existing selection; generated task/topology files are projections | C-035; S-039; no new owner or runner |
| Contract schema-validation facade, generated-artifact policy, topology, toolchain pins, Make admission and Markdown lint path | Relevant validation, generation ownership and command side effects inspected | Reuse compiled validators and public Make operations during later implementation | E-098/099; documentation-save checks restricted by Section 8 |

Remaining broad policy/source assertions are not silently certified by this
inventory. E-102 saves the ledger for assertions moved and rules whose
interpretation changes; new discoveries outside that boundary remain separate work.
Textual caller searches do not establish absence of external consumers. The
public Make surface is the maintained external interface; private helper paths
receive no compatibility promise from this plan.

##### 3. Module Boundary Diagnosis

The initial checker mixed enforcement and self-tests, making fixture mistakes
look like lint failures. E-104 implements the responsibility split below;
E-105 corrects policy admission and matching within the same static-analysis owner.

| Responsibility | Current location | Decision / owner | Finding / contract |
| --- | --- | --- | --- |
| Arguments, file loading, summary and exit mapping | Checker CLI and existing helper dependencies | Split: thin CLI plus one backend boundary module; retain static-analysis ownership | F-050; C-035 |
| Policy shape validation and evaluation | Ad hoc CLI normalization/check functions | Consolidate schema admission through existing compiled validator; retain pure evaluation over explicit source collections | F-050/052; C-034 |
| SQL parsing, exact-set containment, support-root permissions | Existing focused helpers | Keep; reuse rather than create another parser, traversal layer or fixture lifecycle | C-032/036 |
| Scanner and parser examples | Embedded assertions called by `main()` | Move to focused named tests; expected outcomes independent of production policy | F-050/053; C-035 |
| Production-policy wiring | Embedded owner fixtures and authored rules | Keep meaningful allowed/forbidden examples in assurance tests; isolate them from generic mechanics | F-051/052; C-032–034 |
| Fixed count and generated fake fixture paths | Embedded projection and migration assertions | Remove fixed count; replace path-derived fixtures with concrete source examples | F-053; C-032/033 |
| Unrelated module-local scanners and product behavior | Owning modules | Defer; selected iteration is central checker only | Explicit exclusion, no dependency on their consolidation |

##### 4. Public Contract and Behavior Freeze Map

Preserve owner-required guarantees, not observed test defects. The initial
inspection witnesses are now qualified by the exact execution evidence in
E-103–109; historical passes remain outside the final portfolio. Production authorization, replay,
transactions, HTTP/WebSocket contracts, saved views, UI selectors and database
schemas receive no changes; their functional tests are not replaced by static
source checks.

| Contract | Adopted requirement / observed posture | Required treatment and assurance | Slice / validation |
| --- | --- | --- | --- |
| C-032 — Production ownership | Core 01 REQ-01-004/005/649 and projection owners; embedded import, Records, Jobs, projection and migration witnesses; assembly table-set equality | Preserve dependency direction, Records-owned writes/exact reads, projection isolation, Jobs storage/locks, migration ownership and distinct support permissions. Keep existing assembly equality; remove redundant literal count | S-036–040; V-051/052/054 |
| C-033 — Effective path selection | Current matcher treats `*.go` literally; E-087 already exposed Timeline policy drift. Harness REQ-063 controls bounded traversal and file-exact permissions | Separate correction: explicit bounded pattern grammar, qualification of all eleven affected rules, and narrow Timeline policy repair. Proposed owner clarification must precede changed semantics; no broader storage permission or scan root | S-038; V-053; RB-016 |
| C-034 — Policy admission | Harness REQ-063 requires v4 validation and explicit match kinds; current normalization silently loses four import rules and schema declares an object as an array | Separate correction: compiled validation, strict rule fields, proper import rules, no malformed-input compatibility. Preserve explicit import/token semantics and failure mapping | S-038; V-053/055 |
| C-035 — Public operation and truthful assurance | Harness REQ-012–016/062/063; public lint operation also backs the Records architecture row; harness boundary operation owns self-tests | Keep command/operation/summary identities and sorting/exit semantics. Separate scanning from tests; register assurance once, retain independent failure reporting and inert imports | S-037/039/040; V-052/055/056 |
| C-036 — Exact sets and safe support access | Harness REQ-063 and support-inventory ownership; existing exact-file-set test plus production/support projection fixtures | Preserve rejection of stale/missing/omitted/escaping/symlink/non-regular exact paths, unknown sets and conflicting selectors. Preserve source-table file-exact restrictions; support permission never grants production permission | S-037/038; V-052/053 |

##### 5. Coupling and Boundary Findings

| Finding / classification | Evidence and consequence | Structural remedy / future benefit | Contracts / slices / validation |
| --- | --- | --- | --- |
| F-050 / `must_fix` for S-037 | `main()` calls embedded boundary, token and migration fixture groups before scanning results; SQL cache is module-global | One evaluator, invocation-local analysis, separate named tests; lint reports repository violations and tests report checker defects without repeated fixture execution | C-032/035/036; S-037; V-052 |
| F-051 / `must_fix` for S-038 | Eleven policy rules contain `*.go` or `*_adapters.go` unsupported by `matchesPattern`; synthetic paths can retain a literal `*` and falsely qualify the same bug | Correct the small path grammar and independently qualify every affected policy; reconcile narrow Timeline contracts and source tables rather than suppress failures | C-032/033; S-038; V-053 |
| F-052 / `must_fix` for S-038 | Four Recovery-related import entries are in `forbidden_source_tokens`; their missing `tokens` normalize to empty. `generated_root_writes` is an object but schema says array; two call rules contain ignored fields | Enforce one schema-admission boundary, fix authored input placement/shape and remove inert fields. Prevent silent no-op rules without widening accepted input | C-034/035; S-038/039; V-053/055 |
| F-053 / `must_fix` for S-037 | Literal ten-projection assertion duplicates stronger assembly equality; migration fixtures derive paths/symbols from rules and repeat helpers | Retire count; use concrete permitted/prohibited examples and existing semantic set equality. Consolidate overlap without deleting distinct read/write/lock or safety cases | C-032/033/036; S-036/037; V-051/052/054 |

Each disposition records one of: **P** production requirement; **H** necessary
harness safety/evidence invariant; **I** incidental implementation arrangement.
P/H assertions need a meaningful surviving witness; I assertions need a reason
for removal, not a fabricated successor. No automatic one-test-per-rule, fixed
count, filename census, or mirror of implementation branches is an exit criterion.


The findings describe the pre-remediation baseline; E-104–108 record their fixes.

| Gap | Changed areas | Rationale / long-term benefit | Compatibility / residual risk | Completion witness |
| --- | --- | --- | --- | --- |
| F-050 | Implementation, tests, registration | Separate process I/O, evaluation and assurance; invocation-local state makes repeated inspection deterministic | Public command/output retained; no private shim. Leaving the old design would couple lint to fixture failures | Independent successor pass before deletion, inert imports and final CLI/mechanics cases |
| F-051 | REQ-063, evaluator, policy, tests | Explicit bounded grammar makes selection predictable as files evolve | Invalid glob forms now reject; directory intent is explicit. Dormant policies would otherwise conceal violations | Eleven real-path positive/negative policies; narrow Timeline corrections; final repository scan |
| F-052 | Schema, policy, evaluator, generated validators, tests | One fail-closed admission boundary prevents silent no-op rules and typos | Malformed v4 input is intentionally rejected; no new reader/version or data migration | Real-manifest admission, invalid input rejection, explicit Recovery imports and schema qualification |
| F-053 | Tests, checker, registration | Independent examples and semantic set equality avoid count/fixture maintenance | Incidental counts and rule-derived examples retired. Retaining them would preserve false confidence | Existing provider/table equality and concrete import/declaration/access witnesses |

##### 6. Refactor Workstreams

The prerequisite column is authoritative. All workstreams are included and run
serially. Save each exit here before its successor starts. No parallel agent work
is planned. The current user instruction authorizes execution of WF-36–40.

| Workflow / slice | Class / prerequisites | Goal and likely changes | Validation / binary exit |
| --- | --- | --- | --- |
| WF-36 / S-036 | Root; E-101 authorization and source revalidation | Save P/H/I assertion/rule ledger; owner review and exact characterization portfolio. Handoff and focused existing assurance only | V-051; every moved assertion and changed policy has a disposition; baseline results distinguish current passes from intentional defect reproductions |
| WF-37 / S-037 | Chain; WF-36 exit | Extract backend evaluator, thin CLI, move/register focused tests, retire redundant assertions/cache state | V-052/054; successor assurance executes independently before embedded assertions disappear; public scan output preserved |
| WF-38 / S-038 | Chain; WF-37 exit; RB-016 owner review | Correct schema admission, malformed rules and bounded path matching; qualify affected policies | V-053/054; no silently ignored rule or wildcard; no widened exact storage permissions; owner clarification recorded |
| WF-39 / S-039 | Chain; WF-38 exit | Close authored routing/dependencies; regenerate projections and remove obsolete references | V-055; existing rows retain truthful claims, unique registration and consistent generated/schema/catalog state |
| WF-40 / S-040 | Chain; WF-39 exit | Finalize, freeze, qualify missing final obligations and hand off | V-056/057; compatible evidence, native exits, cleanup/retained validity and skips recorded; no out-of-scope changes |

No production correction is implicit. A newly exposed genuine production
violation blocks the dependent slice and is recorded as separate future work.
Do not add an exception, relax an owner, or restore the broken matcher to claim
completion. Independently completed extraction can remain qualified on its own
baseline, with the correction explicitly blocked.

##### 7. Proposed Refactor Slice Plan

Implementation authorization is established by E-101. Entry gates still apply;
completion requires the actual recorded exits, not historical planning evidence.

| Slice / kind | Status / exit evidence | Delivered change / assurance | Rollback boundary |
| --- | --- | --- | --- |
| S-036 / characterization | DONE; E-102/103 | P/H/I ledger and current baseline; RB-015 closed | Documentation and baseline only |
| S-037 / structural preservation | DONE; E-104; final E-108/109 | Evaluator/CLI separation, named independent tests, invocation-local state and count retirement | Revert evaluator, CLI, tests and registration together |
| S-038 / harness correction | DONE; E-105; final E-108/109 | Approved REQ-063 grammar, strict v4 admission, bounded policy corrections; RB-016 closed | Revert coordinated owner/input/code correction and regenerate |
| S-039 / routing and generation | DONE; E-106; final E-108/109 | Existing identities/rows retained, single registration and current generated outputs | Revert authored routing changes and regenerate; no aliases |
| S-040 / qualification | DONE; E-107–109 | Finalization, compatible final portfolio, native exits, retained validity and handoff | Preserve failed evidence if later repairs require requalification |

###### S-037 implementation decisions

Use one new backend boundary module, not a hierarchy of rule classes or public
facades. Its pure evaluator accepts the prepared policy, ordinary source files,
migration-boundary files, support files and resolved exact sets; it returns sorted
violations and dynamic-relation observations. Its filesystem entrypoint loads
those inputs for the CLI. Keep arguments, summary identity/digests/timestamp,
stdout/stderr and exit mapping at the process edge. Retain existing SQL,
exact-file-set and support-inventory helpers. Scope SQL memoization to one
inspection; imports must not scan, register tests, or terminate the process.

Place mechanics, authored-policy examples and SQL-parser cases in focused files
under `tools/harness/static-analysis/tests`. Register through the existing
boundary-contract entrypoint once; reuse existing inert-import/failure-isolation
assurance. Tests use minimal explicit policies and concrete source paths, with
expected outcomes authored independently. Preserve distinct CTE, dynamic-target,
read/write/lock, exact-set and support-scope cases; consolidate repeated setup and
identical assertions only. Move wiring in the same slice as the tests.

Delete the fixed ten-rule assertion without another count registry. Keep
`TestProjectionTableOwnershipSetsAreExactlyEqual` and its existing assembly row:
it compares active descriptors with policy, migration and Recovery table sets.
Keep exact-fixture permission assertions and code-backed provider-manifest parity.

###### S-038 policy decisions

Reuse `validateSchemaSync` through the existing contract facade before preparing
policy or scanning. Correct v4 schema projection errors in place; no new version,
dual reader or malformed-input fallback is needed. Close rule object fields to
those consumed by the evaluator; preserve established meaningful optional values
and use semantic checks only for cross-reference/path constraints the schema
cannot establish. Remove ignored `scan_paths`/`production_only` fields from the
two existing call rules rather than inventing another call-selection mode.

| Current input / defect | Planned disposition |
| --- | --- |
| `generated_root_writes` | Correct schema to its existing object shape and required consumed fields; validate the real manifest, not just arbitrary schema probes |
| `recovery-application-no-operator-transport-imports` | Move to `forbidden_go_imports`; one subtree ban on the operator package covers its child implementation too |
| `recovery-sqlc-confined-to-private-store-adapter` | Move to exact forbidden-import matching with the existing `store.go` permission; no wider directory permission |
| `recovery-application-no-audit-adapter-import` | Move to explicit subtree forbidden-import matching |
| `retired-operator-recovery-cli-import` | Retire obsolete path tombstone; preserve the substantive application/transport boundary through the rule above and ordinary compilation; no fake successor |
| Timeline root import allowlist | Add exact `revisions/historycontract`, `imports/ownerfacade`, `incidentbundles/sourceport`, `records/subtypepresence` imports under the repository module prefix; do not permit their entire parent implementation packages |
| Timeline canonical-source table allowlist | Admit `timeline_source_provenance` and necessary `records` joins; retain the independent exact Records read/write policy and source ownership. Do not admit arbitrary envelope access |

Implemented REQ-063 clarification (E-105): ordinary path selectors use normalized repository-
relative paths, exact path matching, `*` matching within one component, and a
terminal `/**` matching that subtree on a segment boundary. Unsupported glob
forms reject. Express directory selection explicitly instead of implicit prefix
matching. This does not widen traversal roots, change import matching, permit
globs inside exact-file sets, or widen file-exact source-table permissions. The owner text was amended under the approved implementation plan before
changing matching semantics; E-105 records that source amendment.

Qualify the eleven affected rules: `links-retired-production-declarations`,
`incidents-root-is-transport-neutral`, `timeline-no-tabular-parser-implementations`,
`timeline-approved-owner-imports`, `timeline-canonical-source-repository-tables-only`,
`retired-incidents-http-root-api`, `assessments-retired-incident-bundle-helpers`,
`retired-workbook-legacy-mutation-surfaces`,
`retired-workbook-provider-callback-and-restore-aliases`,
`retired-workbook-adapter-admission-dispatch-and-hash-override`, and
`tasks-decisions-retired-forwarders-and-dead-surfaces`. This list bounds the matcher
correction, not a new enduring census. Keep unrelated policy meanings unchanged;
reorganizing all historical tombstones is outside this slice.

##### 8. Validation Plan

**Implementation checks are recorded by exact run in E-103 onward.** Earlier
planning observations remain historical and are not canonical product verification.
The public selections were rediscovered with `make help ALL=1` and
`make task-guide OWNER=<owner> DETAIL=coverage JSON=1` at E-101; discovery supplies
selection information, not passing evidence.

| Validation / stage | Exact command or obligation from repository root | Assurance / failure gate |
| --- | --- | --- |
| V-050 / document save | Direct structure, links, table widths, IDs, dependencies and archive comparison; `git --no-optional-locks diff --check -- docs/handoffs/testing-remediation-01.md`; compare baseline/index | E-100; document/preservation checks only, no product claim |
| V-051 / pre-move | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries`; `make lint CHECK=backend-boundary`; V-054 assembly row | Save exact baseline and P/H/I ledger; separately record known expected defect reproductions. A green old scan cannot establish wildcard/admission correctness |
| V-052 / post-S-037 | Same existing harness boundary row; `make lint CHECK=backend-boundary` | Named independent mechanics/policy/parser tests, inert imports, invocation-local state, CLI success/violation/invalid-input output; no embedded self-tests or duplicate registration |
| V-053 / post-S-038 | Same harness boundary row and backend-boundary lint; `make source-check CHECK=schema` | Real manifest admitted; missing/unknown/misplaced fields rejected; eleven corrected rules have real-path positive/negative examples; C-033/034/036 preserved |
| V-054 / pre-move and affected post-slice | `make test-slice OWNER=module.projections ROWS=module.projections.assembly.catalog_and_source_ownership_927c3f08ce` | Existing provider/table/source-ownership equality and fixture permissions survive removal of literal count; no duplicate set registry |
| V-055 / post-S-039 | `make generate`; `make source-check CHECK=generated`; `make source-check CHECK=schema`; `make source-check CHECK=catalog`; `make source-check CHECK=identity`; `make lint CHECK=scripts` | Authored inputs and generated outputs agree; selectors/claims/operations remain truthful; generated files never hand-edited |
| V-056 / final gate | `make agent-finalize`, then freeze source and run missing compatible obligations from V-052–055; `make harness-observability-check RESULTS_DIR=<exact-successful-run-root> CHECK=all` for qualifying roots | Native exits, exact rows, source/toolchain identity, cleanup and retained validity recorded. No automatic replay of every suite or historical CI run |
| V-057 / final handoff | `make lint-markdown` during later implementation where admitted; direct handoff checks and final diff/preservation review | Final documentation truthful, old evidence historical, no out-of-scope changes; configured lint alone does not cover this handoff |

Future cases cover root/renamed/nested files, sibling-prefix exclusion, rejected
unsupported patterns, exact/subtree imports, production/test code separation,
source-text versus identifier matching, SQL CTE/read/write/lock/dynamic/unclassified
syntax, malformed policy, and exact-set missing/stale/omitted/escaping/symlink/
non-regular paths. Keep test-support permissions separate from production access.
One isolated CLI witness per distinct output/exit outcome is sufficient; do not
repeat complete subprocess suites to exercise pure evaluation.

Use installed pins from `tools/toolchain_pins.json`, canonical runner context and
fresh explicit run IDs. Retain actual native exits and exact roots, normally
`.cartulary/test-results/<run-id>`. Finalizer retained-run maintenance needs a
compatible successful full warm-check `RESULTS_DIR`; otherwise report it skipped
because `RESULTS_DIR` is unset. Do not launch a redundant full check for maintenance.
After repairs, reevaluate compatibility and rerun only missing/affected obligations.
Full CI, browser/visual/a11y, database integration and performance/release work are
not default additions to this central static-checker slice; expand only for a
specific uncovered production or harness obligation. No production behavior
changes are planned, so functional domain tests are preserved, not replaced.

| Historical planning-session command / side-effect review | Admission / result |
| --- | --- |
| Source reads, Git inspection with optional writes disabled, external baseline material | Executed for E-098/099; no test evaluation or service activity |
| `make lint-markdown` | Skipped. The refactor-tracker skill prohibits writes outside the tracker; Make admission can write `.git/cartulary-workspace.lock`, the recipe can prepare dependencies, and lint writes cache/run artifacts. No run-then-restore workaround. Configured globs also omit this handoff; direct checks are necessary |
| `make help ALL=1`, `make task-guide` | Not executed during save; Make admission is not read-only under this write boundary. Public recipe/owner inspection supplies the future command plan |
| `make format`, `make generate`, `make agent-finalize`, all tests/product checks | Not executed: document-only scope; implementation workstreams are future work |

##### 9. Top-Level Work Tracker

| ID | Work item | Workflow / status | Dependencies / exit |
| --- | --- | --- | --- |
| T-041 | Save selected central-checker plan | Planning; DONE | E-098–100; V-050 direct document checks and tracker-only preservation complete |
| T-042 | Guarantee ledger and current characterization | WF-36; DONE | E-101 implementation entry; V-051 closes RB-015 |
| T-043 | Separate enforcement and assurance | WF-37; DONE | WF-36; F-050/053; V-052/054 |
| T-044 | Repair policy admission and matching | WF-38; DONE | WF-37 and RB-016; F-051/052; V-053/054 |
| T-045 | Routing/generated closure | WF-39; DONE | WF-38; V-055 |
| T-046 | Final qualification and handoff | WF-40; DONE | WF-39; V-056/057; compatible results and limitations recorded |

##### 11. Open Questions and Blockers

| ID | Blocker / affected slice | Needed evidence or authority | Status / resolution |
| --- | --- | --- | --- |
| RB-015 | Missing current characterization for risky checker movement; S-037 and successors | WF-36 P/H/I ledger and V-051 current baseline, with distinct expected defect reproductions | DONE; E-102/103 ledger and current baseline recorded |
| RB-016 | Owner/projection review before changed path interpretation; S-038 and successors | Review/adopt the bounded REQ-063 clarification and qualify the eleven affected policies plus explicit Timeline/Recovery corrections | DONE; E-105 records the user-authorized REQ-063 amendment and qualification; product repair remains excluded |

Both implementation-entry gates are closed by E-103/E-105. No current
primary-owner contradiction has been established.
A newly discovered production violation must be recorded separately and blocks
only dependent work. Historical RB-013/014 resolutions remain in the archive.
The separate WF-33 in [the preceding handoff](testing-remediation.md) is neither
resumed nor closed by this document.

##### 12. Binary Completion Criteria

| Outcome | Current statement |
| --- | --- |
| Planning completeness | Complete for F-050–053 and the selected central-checker scope |
| Workstream completion | WF-36–40 and T-042–046 DONE; E-103–109 record sequential exits and the final handoff |
| Implementation authorization | Granted by the user's implementation instruction; E-101 |
| Implementation boundary | Authored checker, tests, policy/schema, bounded owner clarification and Make-generated validators; no production behavior changed |
| Final qualification | E-108 compatible portfolio passed; E-109 records native exits, retained validity, documentation and preservation review |
| Historical evidence | Earlier sessions remain historical; no prior CI pass substitutes for the final portfolio |
| Review responsibility | Human projection-fidelity review remains separate from executable validation; no release/publication claim |

- [x] All moved assertions and changed rules have a P/H/I disposition.
- [x] Independent successor tests passed before embedded assertions were removed.
- [x] Policy preparation, filesystem loading and pure evaluation have distinct responsibilities.
- [x] No import-time scanning/test registration or process termination occurs in the reusable modules or CLI.
- [x] SQL analysis belongs to one inspection; parser and access-mode guarantees survive.
- [x] Explicit path matching and strict compiled v4 admission are qualified.
- [x] All eleven affected policies and narrow Recovery/Timeline repairs have concrete witnesses.
- [x] Existing semantic projection equality replaces the redundant literal count.
- [x] Exact-set safety and support/production permission separation remain qualified.
- [x] Public operations, summary/exit behavior and row identities remain intact.
- [x] Generated outputs were regenerated through Make and verified.
- [x] Every workstream exit was saved before its successor began.
- [x] Finalization preceded final-source qualification; maintenance skips are explicit.
- [x] Final compatible results, cleanup, retained validity and limitations are recorded.
- [x] Documentation, historical preservation and final diff/scope review are complete.

</details>

### Implementation entry — E-101 / WF-36

2026-10-27: user authorized implementation of the proposed plan. HEAD remains
`3bd81c1b2dc9fb8e8424fa50ce0633f982840702`; the sole pre-existing change is
this tracker. Its contents and the original checker are preserved at
`/tmp/cartulary-boundary-implementation-z4bwfomz`. Historical evidence remains historical.

The accepted plan requires compiled-validator regeneration in WF-38 before
admission qualification, with WF-39 closing remaining routing/generation.
No production API/storage change, new runner or compatibility mode is admitted.
`make help ALL=1` and owner guidance establish current public selections.


### WF-36 guarantee and correction ledger — E-102 / V-051

The source inspection reconfirms F-050–053 at the E-101 baseline. The following
ledger covers all moved assertion groups. P = owner requirement, H = harness
safety/assurance, I = incidental arrangement. P/H witnesses survive; explicit
examples below are source evidence, not passing execution of the corrected checker.

| Assertion group | Class / authority | Disposition and successor |
| --- | --- | --- |
| SQL relation fixtures and invalid/unclassified syntax | H; REQ-063, C-032 | Move every named parser case into individually registered SQL tests; retain closed allowlist/dynamic denial cases in mechanics tests |
| Records owner write, foreign write and unapproved join | P; Core 01 REQ-649 | Preserve three allowed/denied policy examples |
| Projection query read, source-owner read, fixture read/write, production/fixture separation, unapproved fixture write | P/H; Core 01 REQ-621/626, C-036 | Preserve distinct access and scope examples |
| Exactly ten projection policies/tables | I | Delete; existing projectionassembly exact set equality is stronger |
| Jobs owner write, consumer read, foreign write/lock, auth adapter read, foreign auth/cancellation/Collaboration access, test and migration exclusions | P; C-032 | Preserve every access-mode and boundary example |
| Jobs narrow capability and peer persistence restrictions | P; C-032 | Preserve composition/consumer and production/test import/token examples |
| Generic delete/restore SQL metadata and approved contract imports | P; C-032 | Preserve forbidden metadata and allowed/denied contract witnesses |
| Config kernel import and registry boundaries | P; C-032 | Preserve neutral import and three forbidden owner directions plus registry witness |
| Contracted Config retired surfaces and supported decoder/clone seams | P; C-032 | Preserve six prohibited examples and three supported examples |
| Imports owner facade/peer implementation boundary | P; C-032 | Preserve capability acceptance and peer/Workbook/Projections denials |
| Imports exact-set transport rule | P/H; REQ-063, C-036 | Preserve admitted route binding and retired-call rejection; reuse independent filesystem safety tests |
| Artifacts HTTP/consumer/idempotency boundaries | P; C-032 | Preserve allowed and prohibited case per boundary |
| Migration allowed imports/declarations/single construction | P; C-032 | Preserve explicit four-source approved portfolio |
| Migration restricted imports, retired declarations and extra construction | P/H; C-032/035 | Replace rule-derived source strings/paths with concrete policy examples; Links wildcard correction belongs to WF-38 |
| Identifier token matching and malformed kinds | H; REQ-063 | Move whitespace/comment/string/rune/raw-string/longer-identifier and invalid kind cases into mechanics tests |
| Embedded fixture invocation and global SQL cache | I | Remove from scan; independently registered tests and per-inspection analysis replace them |

Legacy labeled examples accounted for above: `records owner write`; `revisions write`; `unapproved source-owner read join`; `private projection query read`; `source-owner projection read`; `Timeline test-support projection read`; `Projections test-support assessment projection write`; `Projections test-support path is not a production permission`; `unapproved projection fixture write`; `Jobs owner write`; `consumer Jobs read`; `non-owner Jobs write`; `non-owner Jobs transition lock`; `auth owner adapter storage access`; `Jobs auth storage access`; `Jobs extension cancellation storage access`; `Jobs Collaboration storage access`; `Jobs test storage access`; `Jobs migration write`; `application Jobs composition`; `consumer broad Jobs manager`; `consumer raw Jobs transition lock`; `consumer test broad Jobs manager`; `application Auth owner adapter`; `Jobs peer persistence import`; `source owner adapter contract import`; `unapproved contract import`; `platform-neutral external import`; `application import`; `module owner import`; `platform owner import`; `retired full-document source`; `retired reflective owner clone`; `retired in-memory TOML snapshot loading`; `retired configassembly admission definition`; `retired configassembly admission caller`; `wire-tagged application deployment projection`; `Imports capability imports`; `Imports peer source-owner implementation import`; `Imports Workbook implementation import`; `Imports Projections implementation import`; `Artifact transport-neutral source`; `Artifact future HTTP import`; `Artifact published projection contribution contract`; `Artifact future Workbook implementation import`; `Artifact centralized idempotency lookup`; `Artifact future direct idempotency lookup`.

| Changed policy / shape | Class / concrete defect | Planned correction / witness |
| --- | --- | --- |
| `links-retired-production-declarations` | P/H; literal wildcard in `internal/modules/links/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `incidents-root-is-transport-neutral` | P/H; literal wildcard in `internal/modules/incidents/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `timeline-no-tabular-parser-implementations` | P/H; literal wildcard in `internal/modules/timeline/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `timeline-approved-owner-imports` | P/H; literal wildcard in `internal/modules/timeline/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `timeline-canonical-source-repository-tables-only` | P/H; literal wildcard in `internal/modules/timeline/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `retired-incidents-http-root-api` | P/H; literal wildcard in `internal/modules/incidents/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `assessments-retired-incident-bundle-helpers` | P/H; literal wildcard in `internal/modules/assessments/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `retired-workbook-legacy-mutation-surfaces` | P/H; literal wildcard in `internal/modules/workbook/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `retired-workbook-provider-callback-and-restore-aliases` | P/H; literal wildcard in `internal/modules/workbook/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `retired-workbook-adapter-admission-dispatch-and-hash-override` | P/H; literal wildcard in `internal/app/workbookassembly/*_adapters.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `tasks-decisions-retired-forwarders-and-dead-surfaces` | P/H; literal wildcard in `internal/modules/tasksdecisions/*.go` does not select a concrete filename | Bounded matcher and explicit allowed/denied source examples in WF-38 |
| `retired-operator-recovery-cli-import` | P/I; `imports` ignored, missing `tokens` becomes empty | Retire obsolete private-path tombstone |
| `recovery-application-no-operator-transport-imports` | P/I; `imports` ignored, missing `tokens` becomes empty | Move to forbidden import rules with explicit kind and concrete importing source |
| `recovery-sqlc-confined-to-private-store-adapter` | P/I; `imports` ignored, missing `tokens` becomes empty | Move to forbidden import rules with explicit kind and concrete importing source |
| `recovery-application-no-audit-adapter-import` | P/I; `imports` ignored, missing `tokens` becomes empty | Move to forbidden import rules with explicit kind and concrete importing source |
| `generated_root_writes` | H; real object conflicts with schema array | Closed object schema; real-manifest admission and malformed-shape rejection |
| Two call rules with `scan_paths`/`production_only` | I; evaluator ignores these fields | Remove misleading fields; retain existing production call restriction |
| Ordinary implicit directory selectors | H; grammar ambiguity | Explicit subtree selectors; preserve bounded traversal |
| Open rule-object definitions | H; unknown fields silently tolerated | Close to consumed fields, preserving meaningful optional values |

Known defects are statically reproduced by the actual paths
`internal/modules/incidents/routes.go` versus `internal/modules/incidents/*.go`,
and the Recovery rules' absent `tokens`. The old matcher compares the wildcard
literally; the old normalizer creates empty token lists. These are intentional
future correction targets, not requirements to preserve and not passing
behavioral reproductions. Independent executable regressions are established
in WF-38 after the evaluator becomes directly testable in WF-37.

### WF-36 exit / WF-37 entry — E-103

V-051/V-054 completed with native Make exit 0 for:

| Command | Exact run root under `.cartulary/test-results/` | Result |
| --- | --- | --- |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries` | `boundary-wf36-harness-20261027` | PASS |
| `make lint CHECK=backend-boundary` | `boundary-wf36-lint-20261027` | PASS |
| `make test-slice OWNER=module.projections ROWS=module.projections.assembly.catalog_and_source_ownership_927c3f08ce` | `boundary-wf36-projections-20261027` | PASS |

These are pre-move checks; they do not qualify later edits. E-102 accounts for
all moved assertion groups and changed rules, with independent regression
expectations for known defects. RB-015 is closed and T-042 is DONE. WF-37 enters
only after this checkpoint is saved. Production files remain unchanged.

### WF-37 exit / WF-38 entry — E-104

The evaluator, loader and thin CLI are separated. SQL analysis is invocation-local.
Parser, mechanics, policy and CLI cases register once through the existing boundary
entrypoint; inert-import coverage includes every new support module. Existing
ownership witnesses survive, while the literal count and rule-derived migration
fixtures are removed. Concrete declaration examples replace per-symbol repetition;
the previously inert Links root wildcard receives its independent witness in WF-38.

| Qualification | Exact root under `.cartulary/test-results/` | Native exit / outcome |
| --- | --- | --- |
| Successor harness row, before embedded assertions were removed | `boundary-wf37-successors-20261027` | 0 / PASS |
| Same harness row, after CLI replacement, including process contract cases | `boundary-wf37-harness-20261027` | 0 / PASS |
| `make lint CHECK=backend-boundary` | `boundary-wf37-lint-20261027` | 0 / PASS |
| Existing projection assembly row | `boundary-wf37-projections-20261027` | 0 / PASS |

T-043 is DONE. WF-38 starts after this saved exit. User approval of the proposed
plan authorizes the bounded REQ-063 amendment; its source text is updated before
matcher semantics. Human review of projection fidelity remains a separate review
responsibility and is not claimed by executable checks.

### WF-38 exit / WF-39 entry — E-105

REQ-063 now states the approved normalized exact/component-star/terminal-subtree
grammar. The v4 schema admits only consumed rule fields, rejects blank rule
values, and models generated-root writes correctly. The evaluator validates
through the compiled contract facade before loading source/support inputs.
Recovery restrictions moved to explicit import rules; the obsolete private-path
tombstone and ignored call fields are gone. Timeline admits four exact contracts,
provenance and Records joins without widening the independent Records permissions.
The eleven affected policies have concrete positive/negative witnesses.

| Qualification | Exact root under `.cartulary/test-results/` | Native exit / outcome |
| --- | --- | --- |
| `make generate` | `boundary-wf38-generate-20261027`, `boundary-wf38-generate-final-20261027` | 0 / PASS; second includes whitespace-only-value rejection |
| Harness boundary row | `boundary-wf38-harness-20261027`, `boundary-wf38-harness-final-20261027` | 0 / PASS; final adds duplicate-set/symlink-parent and CLI inert-import coverage |
| `make lint CHECK=backend-boundary` | `boundary-wf38-lint-20261027`, `boundary-wf38-lint-final-20261027` | 0 / PASS; no production violation exposed |
| `make source-check CHECK=schema` | `boundary-wf38-schema-20261027`, `boundary-wf38-schema-final-20261027` | 0 / PASS |
| Existing projection assembly row | `boundary-wf38-projections-20261027` | 0 / PASS; final-source requalification remains WF-40 |

Generated outputs were produced only through Make. Inspection of compiler output
found 154 other validator files changed only in generated identifiers; the
backend-boundary validator is the sole substantive compiled-schema change.
RB-016 and T-044 are DONE. WF-39 starts after this checkpoint. Human
projection-fidelity review remains distinct from these machine checks.

### WF-39 exit / WF-40 entry — E-106

Current authored operation/row identities remain correct: the CLI is still the
Records architecture operation and successor tests register exactly once through
`harness.harness_boundary_contract`. Imports carry the new internal dependencies;
no additional process entrypoint, facade registry, task alias or row is needed.
Existing backing-script declarations satisfy their recipe contract without turning
all private imports into permanent entrypoint roots. No obsolete moved path remains.
The earlier WF-38 Make generation supplies the current compiled schema; WF-39
checks it rather than creating a duplicate generation step with unchanged inputs.

| Command | Exact root under `.cartulary/test-results/` | Native exit / outcome |
| --- | --- | --- |
| `make source-check CHECK=generated` | `boundary-wf39-generated-20261027` | 0 / PASS |
| `make source-check CHECK=catalog` | `boundary-wf39-catalog-20261027` | 0 / PASS |
| `make source-check CHECK=identity` | `boundary-wf39-identity-20261027` | 0 / PASS |
| `make lint CHECK=scripts` | `boundary-wf39-scripts-20261027` | 0 / PASS |

Schema validation is also current at `boundary-wf38-schema-final-20261027`;
only tracker text changed afterward. Script lint uses its configured scripts
scope; execution/import checks in the harness row qualify the changed tools.
T-045 is DONE; this checkpoint precedes WF-40. Finalization runs without
`RESULTS_DIR`: no eligible full warm check was supplied, so retained maintenance
will be explicitly skipped. No redundant full check is launched to obtain one.

### WF-40 final review — E-107

`make agent-finalize` passed with native exit 0 at
`.cartulary/test-results/boundary-wf40-finalize-20261027`.
Final review found that generic identifier mechanics preserved lexical cases but
needed the original Collaboration production-policy wiring witness as well.
That witness is restored using independently authored code-versus-string cases;
test registration imports are grouped at the entrypoint. No production or
policy semantics changed. Finalization is repeated after this test-only repair,
then final-source qualification begins. This is a review correction, not a
failed test or an assertion of earlier final-source acceptance.

### Final qualification portfolio — E-108 / V-056

Finalization passed with native exit 0 at
`.cartulary/test-results/boundary-wf40-finalize-final-20261027` after the E-107
review correction. The implementation source is frozen at
`sha256:25447c697bff27447ab97779bfb7cb7ebe9fc88e3a81514352757a6e9a0440e9`;
toolchain identity is
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
All eight final manifests agree on these identities. Each command below exited 0;
all run summaries report PASS, all cleanup outcomes are completed, retained-secret
scans pass, and no unit is failed, skipped or cancelled. Test-result caching is
disabled for the two selected test slices; no final graph unit reports a cache hit.

| Run root under `.cartulary/test-results/` | Command / distinct obligation |
| --- | --- |
| `boundary-final-harness-20261027` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries`; independent mechanics/policy/parser/CLI/admission and import-isolation assurance |
| `boundary-final-projections-20261027` | `make test-slice OWNER=module.projections ROWS=module.projections.assembly.catalog_and_source_ownership_927c3f08ce`; semantic set and provider/source ownership |
| `boundary-final-lint-20261027` | `make lint CHECK=backend-boundary`; real repository enforcement and Records architecture row |
| `boundary-final-generated-20261027` | `make source-check CHECK=generated`; authored/generated agreement |
| `boundary-final-schema-20261027` | `make source-check CHECK=schema`; schema closure |
| `boundary-final-catalog-20261027` | `make source-check CHECK=catalog`; routing/catalog closure |
| `boundary-final-identity-20261027` | `make source-check CHECK=identity`; semantic identity |
| `boundary-final-scripts-20261027` | `make lint CHECK=scripts`; configured script-check obligation |

Exact-root retained-validity audits and final documentation checks are recorded
at E-109. These results qualify the selected central-checker effort, not full
owner acceptance, whole-project CI, product functional conformance, or publication.

### WF-40 exit and completed handoff — E-109 / V-057

All eight E-108 roots passed the canonical command
`make harness-observability-check RESULTS_DIR=<that-exact-root> CHECK=all`, each
with native exit 0 and `full_validity=1`. Optional resource/diagnostic instrumentation
was not collected; functional acceptance makes no performance claim.

`make lint-markdown` passed with native exit 0 at
`.cartulary/test-results/boundary-final-markdown-20261027`; its summary is
`adhoc/lint-markdown/tool-run-summary.json`. Configured globs omit this tracker,
so final direct review additionally checks its twelve sections, active table
structure, local links, status/dependency consistency, historical preservation,
and Git whitespace. The pre-existing Section 10 history is preserved verbatim
against the E-101 snapshot. All changed paths are within the selected authored
checker/specification/test/policy/schema scope or Make-generated validator outputs;
HEAD remains unchanged and the index has no staged changes.

T-046 and WF-40 are DONE. Every workstream exit was saved before its successor
started; no in-scope implementation or validation blocker remains. The current
private evaluator interfaces are `prepareBoundaryPolicy`, `evaluateBoundaryPolicy`,
and `inspectBackendBoundaries`. Ordinary lint scans repository sources without
running checker fixtures; checker assurance remains in the existing boundary row.
There is no database, deployment, production API or fixture-API migration. Invalid
policy shapes and unsupported selector syntax intentionally reject; no fallback,
alias or private-path compatibility layer was added.

Finalizer retained-run and scheduler-maintenance steps were skipped because
`RESULTS_DIR` was unset: no eligible successful full warm-check run was supplied.
Full CI, browser/visual/a11y, database integration, benchmarks and publication were
not selected because this effort changes central static enforcement and its
assurance, not production behavior in those areas. Module-local scanner
consolidation and the separate preceding-handoff WF-33 remain excluded.
Human projection-fidelity review is a separate responsibility, not a machine-test
claim. The implementation is ready for that review with its exact evidence above.

### Central-checker planning and document save — E-098–E-100

#### Scope and authority

| Date / evidence | Decision / changed file | Next action |
| --- | --- | --- |
| 2026-10-27 / E-098/099 | User selected central checker after read-only inspection. Save only this handoff; WF-36–40 are future work. Prior implementation authorizations archived | Later implementation starts at WF-36 after source/owner revalidation |

#### Backend boundary

| Evidence | Inspected state / decision | Limit |
| --- | --- | --- |
| E-098 | Checker, affected policy and assembly witnesses establish F-050–053; C-032–036 freeze production assurance | No checker, test, module or production source changed; no complete module behavior audit |

#### Frontend boundary

Not applicable: frontend behavior, checks and rendered inspection are excluded.

#### Contract and codegen

| Evidence | Decision | Next action |
| --- | --- | --- |
| E-098/099 | v4 admission repairs and bounded REQ-063 clarification planned; existing command/summary/row identities retained | Adoption/review precedes S-038; authored inputs and Make generation occur only during later implementation |

#### Tests and harness

| Evidence | Commands / results | Limit |
| --- | --- | --- |
| E-098/099 | Read-only source/Git inspection; no test or Make execution. V-051–057 are future obligations | E-097 remains historical CI evidence, not current checker qualification |
| E-100 | Direct document review and `git --no-optional-locks diff --check -- docs/handoffs/testing-remediation-01.md` passed; 12 sections, 21 active tables and 8 active links checked | No Markdown-lint or product pass claimed |

#### Security and authorization

| Evidence | Guarantee / decision | Limit |
| --- | --- | --- |
| E-098/099 | Exact source ownership, support/production isolation and fail-closed policy admission preserved in plan | Static assurance is not a substitute for authorization, replay or transaction integration tests; those tests remain unchanged |

#### Open risks and next session

| Evidence | Risk / next boundary | Resolution |
| --- | --- | --- |
| E-099 | Dormant matcher rules expose stale projections; fixture-derived expectations can conceal defects | WF-36 ledger and independent examples; WF-38 owner-reviewed corrections; no policy exceptions to suppress production defects |
| E-100 | Save boundary verified: all 7,645 outside-tracker entries preserve bytes and lstat identity; raw index, staged entries and HEAD unchanged | Prior Section 10 content preserved verbatim; earlier active prose/evidence preserved with heading demotion |

#### Evidence register

| Evidence ID | Source / baseline | Observation / applicability |
| --- | --- | --- |
| E-098 | Prior read-only planning on `3bd81c1b2dc9fb8e8424fa50ce0633f982840702`; targeted `rg`, `sed`, JSON/path inspection and Git status | User selected central checker; findings, source limits, affected rules and surviving assembly witness identified. No checker or test executed |
| E-099 | Current save baseline, same HEAD and clean source; full content/identity snapshot at `/tmp/cartulary-boundary-plan-sriebz8d/` | Target/output identity admitted; proposed plan materialized with stable new IDs. Former active WF-31–35 sections archived below; their headings are demoted, all prior prose/evidence preserved |
| E-100 | Current documentation and preservation verification; same HEAD; `/tmp/cartulary-boundary-plan-sriebz8d/document-preservation-review.json` | Direct section/table/link review, historical-log/archive comparison and diff whitespace check passed. All 7,645 outside-tracker entries and index/HEAD preserved; T-041 DONE. No test or Make invocation; Markdown lint skipped for the documented side-effect conflict |

### Completed structural-guard iteration — historical active-section archive

The following was the active WF-31–35 plan and closure before E-099. Its source
posture, implementation authorizations, statuses and command results are
historical. Only heading levels have been adjusted for nesting. E-098/099 replace
its current-scope summary; its stable IDs and rationale remain available.

#### Testing Remediation 01: Simplify Structural Test Guards

##### 1. Scope and Source Posture

Document date: 2026-10-27. **Bounded remediation complete through WF-35
validation and handoff; E-095/E-096. Full CI follow-up passed at E-097.** The user's instruction to implement
the proposed plan authorizes the scoped specification amendments, tests, authored
boundary/routing projections, generation, and verification. The earlier document-
only authorization and evidence are historical. Planning and execution are complete; per-workstream exits and final
qualification are recorded separately below.

Repository: `/home/jochi/code/cartulary`; canonical target `.`. Implementation
baseline: `main` at `005f6dfa196022cfe6c7497cdc16953edfd1cf9b`, with only this
tracker already modified. Preserve that existing content and all historical
records. A content/file-identity snapshot of 7,653 tracked/nonignored entries,
HEAD and index, plus the original tracker, is retained at
`/tmp/cartulary-tr01-guards-9qhwwd5g/`. No staging or commit is part of this effort.
The prior save baseline table below remains historical E-079 evidence.

The controlling artifact is `docs/handoffs/testing-remediation-01.md`. Every
workstream exit must be recorded here before the next starts. WF-33/S-033 is a
rollup over four separate executable workstreams, WF-33A–D/S-033A–D. The planning
framework and tracker format organize records; adopted owners govern behavior.

Authorized scope: structural guards in **Artifacts, Assessments, Entities,
Evidence, Imports, Indicators, Parties, Tasks/Decisions and Projections**, shared
HTTP test support, and their authored verification routing. Retire incidental
export inventories, fixed counts, declaration-to-file maps, historical filename
prohibitions and exact helper caller lists. Consolidate meaningful checks under
existing owners. Do not introduce an export-snapshot framework, compatibility
registry, runner, blanket ban on new exports or substitute implementation census.
Production APIs, wire/event contracts, database schemas and fixture APIs remain
unchanged. Adopted requirement changes are recorded in Section 6 and E-085.

**Excluded:** process fixtures and service lifecycle redesign; Graph Projection's
separately adopted architecture; Network Flow; unrelated scanner redesign;
application-fixture follow-up implementation; UI, performance and publication work.
Graph Projection's NLSpec §8.1, architecture decision and Core 04 AC-568 require
separate owner review; this plan does not generalize away their boundaries.

WF-31–WF-35 below belong to **this document**. WF-26–WF-30 are completed
application-fixture history, archived in Section 10 with their original evidence.
Stable historical IDs remain unchanged. The separate open WF-33 in
[the preceding handoff](testing-remediation.md) is neither resumed nor closed here.
Historical implementation authorizations and passes are not current instructions
or qualification for this iteration.

| Authority / adoption | Governing requirement and treatment |
| --- | --- |
| [Core 00](../spec/00_document_set_status_and_precedence.md), current | Adopted owners govern requirements; tests and routing are projections. REQ-00-070/073/074 adopt the Projections/Entities/Indicators decisions. REQ-00-076 separately adopts Graph Projection. No test, generator, runtime metadata, conformance or release evidence may read, stat or hash Markdown. |
| [Core 01](../spec/01_architecture_storage_and_view_contracts.md), current | REQ-01-004/005: module responsibility and dependency boundaries. REQ-01-621/622/623/626: deterministic projection rebuilding, source/history preservation, code-backed descriptors, facade/query contracts and import boundaries. Preserve these semantic requirements. |
| [Core 04](../spec/04_security_deployment_and_conformance.md), current | AC-473/539 retain facade, storage and transaction guarantees. AC-558 and AC-560 were amended in E-085 to require semantic boundary assurance without inventory/count checks. Authorization and security obligations remain binding. |
| [Entities decision](../decisions/entities-module-boundary.md), [Indicators decision](../decisions/indicators-module-boundary.md), [Projections decision](../decisions/projections-module-boundary.md), adopted | Preserve capability separation, dependency direction, construction, transaction and storage ownership. E-085 adopts the approved removal of exhaustive implementation inventories while preserving these substantive boundaries. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md), adopted/current | §§3.1–3.3, REQ-012–016: exact catalog selection, coverage and routing accountability. Preserve one resolver and canonical Go build context; never recycle row IDs or fabricate evidence successors. |
| [Domain](../domain.md); [Core 05](../spec/05_claim_publication_and_benchmark_reproducibility.md) | Domain supplies vocabulary and navigation. No benchmark, release, timing or publication claim; functional checks do not satisfy separate publication requirements. |

| Baseline item | Document-save baseline / evidence | Preservation obligation |
| --- | --- | --- |
| Branch / commit | `main`; `005f6dfa196022cfe6c7497cdc16953edfd1cf9b`; clean at E-079 | Recheck HEAD and outside-tracker contents before reporting success |
| Earlier concurrent changes | Inspection began against `e6ed547531945b62d3006ae8c1d458746ddf4670`; existing working-tree contents stayed unchanged while staging changed the index; the earlier work was subsequently committed by another process | Preserve that work and new commit; no reset, staging, unstaging or attribution of its contents to this save |
| Index | Raw-index SHA-256 `62c219d6129f9d1daaef212cf9d808c80367421def15322a78b1f353a71ff8ce`; staged-entry SHA-256 `2d56f2a1f92f7540172830913c49b7a1adb77c51d14c7908bed4ac65884fa8a0`; no staged edits | Compare both at E-082; use Git with optional writes disabled |
| File universe | 7,653 tracked/nonignored entries; no nonignored untracked files | Per-entry content and lstat identity captured; only this handoff may change |
| Tracker before save | 452,488 bytes; device/inode `2096:348826`; one link; SHA-256 `89ff6fa697da48d4a77fe80b0654af35b36f62ef3465505db63cdf54433d2d4c` | Recheck identity and bytes immediately before edit; preserve historical evidence |
| External comparison material | `/tmp/cartulary-structural-guards-plan-4m_6t9_z/baseline.json` and `original-handoff.md` | Preservation material only, not product evidence; no exhaustive ignored-tree fingerprint |

##### 2. Baseline Repository Inventory

E-079 enumerates the repository file universe above. Entries outside the inspected
seam receive **metadata-only** accounting, including generated, vendor and binary
entries. Ignored caches, dependencies and outputs are excluded from that universe;
symlinks are accounted for as entries, without following external targets. E-080
records the bounded direct inspection. The following counts describe source size,
not acceptance thresholds or a requirement to preserve test counts.

| Path / group | Inspection depth and responsibility | Disposition / evidence |
| --- | --- | --- |
| `internal/modules/artifacts/exported_surface_test.go` (404 lines) | Inventory/parser, export roles, declaration placement and historical file bans inspected | Retire incidental machinery; D-001; F-044 |
| `internal/modules/assessments/export_surface_test.go` (423 lines); `facade_contract_test.go` | Inventory/scanner tests inspected; participant rollback/replay and constructor witnesses inspected | Retire census; preserve functional facade tests; D-002; C-027/028 |
| `internal/modules/entities/boundary_guard_test.go` (709 lines) | Mixed import/call/storage/route assertions and local selector/export reconcilers inspected | Split semantic guards from duplicate catalog and adopted inventory assertions; D-003/004; F-045/047 |
| `internal/modules/evidence/exported_surface_test.go` (375 lines); `projection_collaboration_integration_test.go` | Inventory and retired-directory checks inspected; rebuild/source-history and collaboration witnesses inspected | Retire census, retain domain behavior; D-005; C-029/030 |
| `internal/modules/imports/export_surface_contract_test.go` (541 lines) | Root/owner-facade inventory, exact roles and synthetic export checker inspected | Retire census and unused scanner helpers; D-006 |
| `internal/modules/indicators/exported_surface_test.go` (507 lines), `application_test.go`, `application_actor_test.go`, `provider_encapsulation_test.go`, `production_contract_test.go` | Inventory and mixed structural helpers inspected; constructor, actor, replay, provider and HTTP assertions inspected | Extract meaningful architecture/behavior before removing maps; D-007/008; F-046 |
| `internal/modules/parties/exported_surface_test.go` (188 lines) | Export census and reflected command/result field ordering inspected | Retire incidental export/layout contract; preserve behavior-bearing composition and command tests; D-009 |
| `internal/modules/tasksdecisions/exported_surface_test.go` (478 lines) | Export/file/provider declaration sets, old-name bans and test-helper naming checks inspected | Retire incidental census; isolate any distinct provider-boundary guarantee; D-010 |
| `internal/modules/projections/adapters/export_surface_guard_test.go` (257 lines), `boundary_guard_test.go`, `adapters_test.go` | Inventory scanner, facade import rules, testsupport assembly spelling and constructor/descriptor witnesses inspected | Adopted inventory removal gated; preserve semantic facade and provider isolation; D-011/012 |
| `internal/testutil/httptestx/httptestx_test.go`, `httptestx.go` | Exact projection caller/source-marker matrix, Server exported-field census, envelopes and route-mode assertions inspected | Delete caller/field census; retain service-backed behavior; D-013/014 |
| `internal/app/projectionassembly/catalog_manifest_test.go`, `source_ownership_test.go` | Code-backed registry/descriptors, table ownership and provider SQL boundary witnesses inspected | Retain meaningful exact sets of owned tables/providers; C-029; not symbol snapshots |
| `tools/harness/test-catalog/selector-resolution.mjs`, `test-catalog.mjs`; authored families/verification claims; migration schema | Full-catalog resolution, canonical build context, overlap/coverage checks and affected declarations inspected | Consolidate local selector reconciliation here; D-015; C-031 |
| `tools/backend_module_boundaries.json`; applicable instructions, adopted clauses, Make and lint path | Relevant enforcement entries, authority and documentation-command side effects inspected | Reuse semantic static rules; do not migrate obsolete snapshots into central JSON |
| Remaining enumerated entries | Metadata-only; no movement planned | No claim of corpus-wide assertion review; exact per-assertion ledger for affected tests is WF-31's implementation-entry deliverable |

Follow direct and transitive references before deleting private helpers, including
external test packages, scanner fixtures, authored selectors and verification
claims. Recheck any changed owner/test/routing source before relying on this plan.
A newly discovered semantic assertion remains in place until its disposition and
surviving evidence are recorded; discovery does not expand the authorized scope.

##### 3. Module Boundary Diagnosis

The seam is **test policy ownership**. Independent scanners have conflated useful
architecture boundaries with the historical shape of their implementation. Keep
domain behavior with domain tests, architecture enforcement with the existing
boundary owner, and catalog reconciliation with the existing catalog resolver.

| Responsibility | Decision / destination | Reason and constraint |
| --- | --- | --- |
| Authorization, replay, transaction rollback, wire/event behavior | Keep in owning module/assembly integration and unit tests | These validate production effects; count or placement of exported declarations does not substitute for them |
| Dependency, provider, storage and source ownership | Consolidate under existing backend rules where equivalent; retain a narrow owner test for distinct guarantees | Prohibited dependencies must fail even after files are renamed or split; avoid an additional scanner framework |
| Complete test selection / overlap detection | Solely the existing full-catalog resolver | It already enumerates active Go tests in the canonical build context and enforces exactly-once selection |
| Exact export inventories, fixed counts, filename histories, helper caller lists | Retire without replacement after applicable adoption gates | These encode accidental implementation structure and add maintenance burden without distinct production assurance |
| Tests mixing semantic and incidental assertions | Split by guarantee, then remove only incidental parts | Preserve every meaningful assertion and route it truthfully; deleting a whole mixed file/row is unsafe |
| Owner-prescribed inventories | Propose owner amendments; dependent deletion blocked until adoption | Treat genuine adopted requirements differently from unsupported historical conventions |
| Production/fixture APIs and unrelated scanners | Keep unchanged / defer | No aliases, compatibility mechanisms or broad architecture redesign are needed |

##### 4. Public Contract and Behavior Freeze Map

Freeze **owner-required production behavior**, not existing test implementation.
The witnesses below are the inspected semantic guarantees. E-084 records
pre-removal characterization; E-093 corrects the Harness execution mapping; E-095
records the final exact compatible portfolio. Unrelated passing tests are not
evidence successors.

| Contract | Adopted requirement / observable guarantee | Existing surviving evidence / replacement obligation | Validation |
| --- | --- | --- | --- |
| C-027 | Core 01 module boundaries; adopted owner construction/capability requirements; Core 04 AC-560: fail-closed admission, nil/typed-nil rejection, complete returned capabilities and actor admission | Projections `TestNewRejectsMissingDependencies`, `TestNewReturnsReadyPortsAndImmutableDescriptorSet`; Indicators `TestIndicatorApplicationCompositionAndRepositoryBoundaries`, `TestIndicatorApplicationRejectsNilActorBeforeDependencies`; existing Assessments/Parties/TasksDecisions composition rows. Retain all dependency cases; add a witness only if ledger identifies a real uncovered capability | V-042/043 |
| C-028 | Core 04 AC-539/558/560 and module owners: borrowed transaction ownership, participant failure rollback, replay identity and no repeated effects | Assessments `TestAssessmentFacadeParticipantFailuresRollbackAndReplayBypassesParticipants`; Indicators `TestIndicatorCreateReplayCodecPreservesStoredBytes` and replay subtests in `TestIndicatorApplicationCompositionAndRepositoryBoundaries`; Entities `TestResolveRoute_Integration`. Preserve byte/hash identity, persisted effects and failure-path assertions; an export map is not their evidence | V-042/044 |
| C-029 | Core 01 REQ-01-621/622/626; Core 04 AC-539; Projections decision: immutable descriptors, complete provider/table ownership, provider isolation, production assembly, rebuilding without source/history writes | `TestProjectionAssemblyPortsAreCompleteAndDescriptorsImmutable`, `TestProjectionProviderManifestMirrorsCodeBackedRegistry`, `TestProjectionTableOwnershipSetsAreExactlyEqual`, `TestProjectionProviderSQLSourceOwnership`, `TestValidateProjectionProviderSQLReference`; adapter constructor/descriptor tests; Evidence `TestAttachedEvidenceProjectionRebuild_Integration`. Keep exact semantic sets and mutation/rebuild assertions. If replacing literal Build-call counting, prove testsupport uses production assembly through the existing assembly contract or a focused missing witness first | V-045/047 |
| C-030 | Core 01 REQ-01-228/229/251, Core 04 REQ-04-022/023 and §2.0A, Harness TH-HARNESS-REQ-480: response envelopes, explicit test-route admission, request-time authorization and collaboration effects | HTTP `TestHarnessBootsServerAndAssertsEnvelopes`, `TestStartServerHonorsTestRouteMode`; Indicators production-contract tests; Entities `TestResolveRoute_Integration`; Evidence `TestObjectUploadAttachWorkbookProjection_Integration`. Retain actual requests, error/status/body assertions, replay/revision/event observations and route-mode behavior | V-044/046 |
| C-031 | Harness NLSpec §§3.1–3.3, REQ-012–016: one catalog owner, complete/unique active test selection, truthful row identities and claims | `createRowSelectorResolver().assertSourceCoverage()` plus full `resolveSelectors` overlap checks in existing catalog loader; canonical Go build context. Preserve rejection of missing, duplicate, stale and build-excluded selectors; audit existing harness fixtures and add only missing cases | V-048 |

Architectural requirements remain meaningful even when checked by source inspection.
Retain dependency direction, private storage, borrowed transactions, separate
mutation/query/reporting capabilities and no ambient concrete peer construction.
Exact lexical spelling, export count or declaration filename is not automatically
a requirement. For an adopted inventory rule, Section 6's amendment gate applies.

##### 5. Coupling and Boundary Findings

| Finding | Evidence and consequence | Structural remedy / classification and scope | Contracts / slices |
| --- | --- | --- | --- |
| F-044 | Nine owners independently inventory exports and several self-test injected exports; file placement/history is often part of the same machinery | Retire unsupported snapshots and now-unused scanners/fixtures; no shared replacement census. `must_fix` within D-001–012, subject to mixed-test/adoption gates | C-027–030; S-031/033 |
| F-045 | Entities locally reconciles test selectors using host `go/build.Default` and one family; full catalog already covers active source under canonical build settings | Remove local reconciliation and private helper; preserve completeness/overlap exclusively in catalog, including build-excluded rejection. `must_fix` D-004/015 | C-031; S-033/034 |
| F-046 | Indicators test freezes 56 exports while adopted AC-560 records historical 55→50→54 inventory; inventories also mix peer-construction, actor/provider and fixture checks | Do not repair the mismatch by adopting another count. Amend the requirement; preserve semantic construction, replay and isolation assertions. `must_fix` D-007/008 after adoption | C-027/028; S-032/033 |
| F-047 | Entities/Indicators/Projections tests combine real ownership with exact helper names, files and assembly spelling | Establish assertion-level dispositions; retain or consolidate each distinct semantic boundary before deletion. `must_fix` mixed rows, not just filenames | C-027–031; S-031/033/034 |
| F-048 | HTTP helper test pins eight projection caller filepaths/markers and exact capability/Server fields despite existing behavior and assembly witnesses | Retire caller and reflected-field inventories; preserve envelopes, route admission and semantic capabilities. `must_fix` D-013/014; fixture API unchanged | C-027/029/030; S-033 |
| F-049 | Standalone and mixed catalog rows currently make export-minimization/caller-matrix claims; migration schema requires at least one real successor | Delete obsolete claims with rows; preserve/reword mixed claims to match surviving evidence. Pure retirement has no fabricated successor or alias. `must_fix` per changed row | C-031; S-034 |

The following evaluation applies to each gap; D-001–015 below and E-084–090
record the individual assertion decisions. No production, wire, schema or fixture
API migration is required. Historical run records retain their original meaning;
retired rows do not acquire false evidence successors.

| Gap | Changed areas and remediation | Rationale and long-term benefit | Compatibility impact and unresolved risk | Completion validation |
| --- | --- | --- | --- | --- |
| F-044 | Tests, routing and owner specifications: retire exhaustive export/file histories and unused scanner fixtures | Capability boundaries have architectural meaning; symbol counts do not. Removes recurring bookkeeping and permits cohesive internal refactoring | Internal guard policy changes only; leaving it preserves nine coupled implementation snapshots and rewards misleading edits to satisfy counts | No dead helpers/selectors; all retained capability and integration witnesses pass; no replacement census |
| F-045 | Tests and routing: remove Entities local selector resolver; retain Harness catalog ownership | One canonical build context and resolver avoids divergent test discovery and duplicate maintenance | No test execution API migration; unresolved duplication can accept stale or omit active selectors under different build contexts | Full catalog plus existing missing/duplicate/stale/build-excluded negative fixtures pass |
| F-046 | Specifications, tests, routing and documentation: replace stale Indicator export count with injected-capability, actor and replay assurance | Requirements remain durable as implementation grows; exact stored identities and security admission stay explicit | No replay/wire migration; historical 54-versus-56 mismatch is retired, not normalized to another count. Leaving it makes both spec fidelity and passing inventory checks misleading | Adopted decision and AC-560 agree; constructor, origin, actor, borrowed transaction and stored-byte replay rows pass |
| F-047 | Tests and boundary policy: isolate semantic package/dependency, transaction, storage and assembly checks from file/name inventories | High cohesion and minimal coupling: harmless file splits pass while prohibited dependencies fail | New package guards replace incidental filename restrictions; inactive Entities central duplicate is removed. Unresolved mixed assertions make deletion unsafe and refactoring brittle | Positive rename/split and negative capability cases, backend lint, assembly/provider/table ownership and real rollback/rebuild checks pass |
| F-048 | Tests and routing: remove HTTP caller and reflected field matrices; keep real requests and production assembly delegation | Future consumers do not require a central test-file inventory; observable envelopes and admission remain the contract | Fixture API unchanged; leaving it couples unrelated modules through exact caller lists without stronger behavioral assurance | HTTP success/error envelopes and explicit route modes, assembly delegation and affected real caller integrations pass |
| F-049 | Authored routing, verification claims, migration crosswalk and generated consistency | Evidence remains attributable to real surviving semantics; no compatibility aliases or empty successor claims | Four pure rows retire; the mixed Indicator surface row migrates only to stored-create replay. Unresolved claims would overstate assurance or break exact selection | Generated/schema/catalog/identity checks pass; final exact selected rows resolve once; removed minimization claims have no dangling references |

The disposition ledger below records the approved implementation decisions. **Retain** keeps
the meaningful witness; **consolidate** moves duplicate responsibility to a named
existing owner; **replace** needs a specified surviving assurance before removal;
**retire without replacement** explicitly removes an accidental constraint. E-084
expands mixed entries to individual assertions and records exact clauses, source
locations, tests/static rules, rows and gates; E-087–094 record qualified adjustments. No removal proceeds on an unexamined
mixed entry.

| ID / assertion family | Disposition and exact intended treatment | Gate / surviving owner |
| --- | --- | --- |
| D-001 Artifacts `TestArtifactExportedSurfaceReachabilityLock` | **Retire without replacement:** exact symbol roles, declaration locations, retired-file checks, synthetic export rejection and unused private scanner tests | WF-31 confirms no independent semantic assertion; Artifacts behavior and architecture witnesses remain |
| D-002 Assessments `TestAssessmentExportSurface`, `TestAssessmentExportSurfaceRejectsInjectedExport`, `TestAssessmentExportSurfaceClassifiesReceiverVisibility` | **Retire without replacement:** inventory and scanner self-tests; remove helpers once unused. **Retain** constructor/import/transaction tests in the mixed family | C-027/028; preserve `TestAssessmentFacadeParticipantFailuresRollbackAndReplayBypassesParticipants` |
| D-003 Entities `TestEntitiesProductionImportBoundaries` | **Retire without replacement:** `production exports have exact dispositions` census after adoption. **Consolidate/retain:** meaningful dependency/merge checks via existing backend rules or narrow package/capability-based test; remove filename-specific allowances only after equivalent boundary protection is identified | WF-32 Entities/Core 04 adoption; C-027/028/029; V-047 |
| D-004 Entities `active tests have one exact owner selector` / `entitiesReconcileExactTestSelectors` | **Consolidate:** remove duplicate owner-family/host-build reconciliation and its unused helpers; catalog owns exactly-once selection | C-031; no removal of catalog completeness; any adopted wording revised to reference catalog ownership |
| D-005 Evidence `TestEvidenceExportedSurfaceReachabilityLock` | **Retire without replacement:** symbol census, retired provider-directory checks and scanner fixtures; keep rebuilding, authoritative-history and collaboration tests | C-029/030; no prohibition on harmless future directory organization |
| D-006 Imports `TestImportsExportSurfaceHasExactProductionRoles`, `TestImportsExportSurfaceGuardRejectsUnclassifiedSyntheticExport` | **Retire without replacement:** exact root/facade fields/methods/roles, synthetic extra-export rejection and unused helper/scanner code | WF-31 maps distinct capability/transaction requirements to their meaningful existing witnesses first |
| D-007 Indicators `TestIndicatorExportedSurfaceReachabilityLock` and inventory helpers | **Retire without replacement:** count/map, predecessor file/symbol history and exact private-helper names/locations. **Retain/consolidate:** no ambient concrete peer construction or fabricated auth actor; replace route-registration counting only if it protects an actual uncovered admission guarantee | WF-32 Indicators/Core 04 adoption; C-027/030; existing boundary owner preferred |
| D-008 Indicators mixed application/provider/fixture checks | **Retain:** replay-codec stored bytes, hash/replay behavior, nil/typed-nil construction, actor/security, contribution readiness and provider isolation. **Retire without replacement:** old-file absence or exact repository/helper/fixture field shape. **Replace** any uncovered fixture value-isolation guarantee with one focused semantic witness only if necessary | C-027/028/030; inspect `application_test.go` and `provider_encapsulation_test.go`, not only the inventory file |
| D-009 Parties `TestPartyExportedSurfaceLock_Unit` | **Retire without replacement:** export list and reflected command/result field ordering; Go struct field order is not a public wire contract. **Retain** actual command, mutation, import and reporting capability tests | C-027/028/030; preserve serialization/HTTP behavior where applicable |
| D-010 Tasks/Decisions `TestTasksDecisionsExportedSurfaceAllowlist_Unit` | **Retire without replacement:** declaration/file sets, historical filenames/symbols, test-helper spelling restrictions. **Retain/consolidate:** any distinct provider/import ownership assertion with the existing boundary owner | C-027/028/029; no wholesale deletion of mixed composition evidence |
| D-011 Projections `TestProjectionInternalExportSurface` and scanner fixtures | **Retire without replacement:** exhaustive production/test-fixture exports and synthetic extra-export tests after adoption | WF-32 Projections adoption; preserve separate mutation/query/reporting ports and private storage |
| D-012 Projections related boundary guards | **Retain/consolidate:** approved facade imports, provider separation and production-assembly semantics. **Replace** exact testsupport Build-call/import spelling if needed, after C-029 evidence proves real assembly use. Do not drop `TestProductionProjectionImportsUseApprovedFacades` or meaningful provider dependency rejection merely because they inspect source | C-029; V-045/047; no unrelated legacy scanner redesign |
| D-013 HTTP `TestProjectionCapabilityCallerMatrix` | **Retire without replacement:** exact eight-file caller/marker map and capability field census; remove now-unused AST/reflection helpers | C-027/029 already owned by adapter and assembly tests; no caller allowlist replacement |
| D-014 HTTP `TestHarnessBootsServerAndAssertsEnvelopes` | **Retire without replacement:** exact exported Server-field census. **Retain:** boot, actual envelope assertions and `TestStartServerHonorsTestRouteMode` | C-030; same fixture API and route-mode behavior |
| D-015 full catalog coverage and overlap | **Retain:** existing resolver source enumeration/overlap enforcement. **Replace** missing harness contract examples only if review finds a gap for missing, duplicate, stale or build-excluded selectors | C-031; no second source scanner, resolver or compatibility mode |

Named source anchors for surviving checks: [adapter construction and descriptors](../../internal/modules/projections/adapters/adapters_test.go),
[assembly manifests and immutable capabilities](../../internal/app/projectionassembly/catalog_manifest_test.go),
[provider/table source ownership](../../internal/app/projectionassembly/source_ownership_test.go),
[Assessment transaction/replay](../../internal/modules/assessments/facade_contract_test.go),
[Indicator replay codec](../../internal/modules/indicators/replay_codec_test.go),
[Evidence rebuild and collaboration](../../internal/modules/evidence/projection_collaboration_integration_test.go),
[HTTP boot and route admission](../../internal/testutil/httptestx/httptestx_test.go),
and [existing backend boundary rules](../../tools/backend_module_boundaries.json).
E-087 qualified the Entities enforcement destination: the central root import
rule was inactive, so E-088 removed it and retained one package-level owner-local
guard. Other central child-import, WebSocket and projection storage policies
remain. No filename exceptions or retired-symbol maps were moved into policy.

Entities' remaining named guards for workbook row routes, clipboard planning,
collaboration publication, host identity, merge writes to mention/projection tables
and command-level timeline effects require individual retain/consolidate decisions.
They are **not** implied retirements under D-003. Likewise provider/table manifests
are semantic ownership sets, unlike a census of every exported declaration.

##### 6. Refactor Workstreams

Execute serially: WF-31 → WF-32 → WF-33A → WF-33B → WF-33C → WF-33D → WF-34 → WF-35.
Each workstream is one slice. Save its exit before beginning the next. A failed
exit remains open; diagnose before retrying or broadening.

| Workstream / slice | Work | Risk / exit gate |
| --- | --- | --- |
| WF-31 / S-031 | Exact assertion/helper/row ledger and current characterization | No unexamined mixed assertion; classified baseline and exact surviving witnesses |
| WF-32 / S-032 | Adopt coordinated Entities/AC-558, Indicators/AC-560 and Projections decision amendments | Remove census mandates, preserve C-027–031; AC-539 unchanged; human-authorized adoption recorded |
| WF-33A / S-033A | Artifacts, Assessments, Evidence, Imports, Parties, Tasks/Decisions inventories and atomic routing | Retained behavior/architecture pass; no dead scanner helpers or stale selectors |
| WF-33B / S-033B | Entities census, file-specific import duplication and local selector reconciliation | Reconcile central policy first; preserve merge, Timeline effects, SQL and route boundaries; catalog assurance passes |
| WF-33C / S-033C | Indicators mixed guards, historical snapshots and replay routing | Actor, composition, provider, transaction and replay guarantees retained; dedicated replay successor |
| WF-33D / S-033D | Projections and HTTP inventories | Production assembly delegation, private storage, descriptors, rebuilds, envelopes and route modes qualify |
| WF-34 / S-034 | Complete authored routing/claims and generated consistency | Exactly-once closure; genuine successor crosswalks only; no source or fixture API migration |
| WF-35 / S-035 | Finalize, freeze, qualify bounded union and complete handoff | Compatible evidence, native exits, cleanup and retained validity; truthful skips and final preservation |

Amendments remove exact export counts/dispositions and per-export adoption, not
capability, security, storage or dependency ownership. Export justification remains
a review responsibility. Entities exact test completeness is Harness-owned.
Projections AC-539 contains semantic requirements and stays unchanged; Core 00
ownership stays unchanged. No domain vocabulary change is needed.

Mixed assertions retain their semantic witness until an equivalent replacement
passes. Exact provider/table ownership sets remain. For failed tranches, restore
tests and corresponding authored routing together, never a compatibility mode.

##### 7. Refactor Slice Plan

| Slice | Status / readiness | Output and validation |
| --- | --- | --- |
| S-031 | DONE; E-084 | Ledger, current baseline, V-042 |
| S-032 | DONE; E-085 | Adopted amendments, V-041 documentation review |
| S-033A | DONE; E-086 | D-001/002/005/006/009/010; V-043–047 |
| S-033B | DONE; E-088 | D-003/004; V-043–048 |
| S-033C | DONE; E-089 | D-007/008; V-043/044/046/047 |
| S-033D | DONE; E-090 | D-011–014; V-043/045–047 |
| S-034 | DONE; E-094 | D-015 and complete routing, V-048 |
| S-035 | DONE; E-095/E-096 | Finalization, frozen-source portfolio, V-049 |

Remove standalone Artifacts, Evidence, Imports and HTTP caller-matrix inventory
rows and their obsolete claims. Preserve mixed Assessments, Parties,
Tasks/Decisions, Entities, Projections, Indicators composition and HTTP integration
rows. Allocate one semantic Indicators stored-create-replay row through the public
row authoring procedure; crosswalk the old mixed surface row only to that real
successor. Remove obsolete minimization references from other surviving rows.
Never fabricate successors for pure retirements or weaken the migration schema.

##### 8. Validation Plan

**Historical save scope:** the prior save performed documentation validation
only. The implementation qualified V-042–049 below at E-084–096. Configuration excludes this
handoff from Markdown lint globs, so a green lint result alone cannot qualify it.
Directly review active links, tables, IDs, dependencies, historical preservation
and the authorized implementation write boundary in addition to `make lint-markdown`.

| Obligation | Verification / pass condition | Timing / scope |
| --- | --- | --- |
| V-041 | Configured `make lint-markdown`; direct handoff structure/link/table/ID/dependency review; `git --no-optional-locks diff --check -- docs/handoffs/testing-remediation-01.md`; content/identity/index comparison | Historical save E-081/082; implementation documentation and preservation closure E-096 |
| V-042 | WF-31 assertion/guarantee ledger complete; exact catalog rows derived with `make task-guide OWNER=<affected-owner> DETAIL=coverage JSON=1`; characterization through exact `make test-slice OWNER=<owner> ROWS=<comma-separated-row-ids>` | Completed before deletion at E-084; inspected assertions are distinguished from executed rows |
| V-043 | C-027 constructor and capability witnesses pass, including nil/typed-nil dependencies, complete ports and rejected actors before dependency work | Exact surviving constructor/application rows; add only a demonstrated missing case |
| V-044 | C-028 rollback/replay witnesses pass; participant failure leaves no partial state; replay preserves identity and avoids repeat effects | Exact Assessments/Indicators/Entities and other ledger-selected mutation rows; real transaction integration where required |
| V-045 | C-029 adapter/assembly, descriptor-mutation, provider/table ownership and corrupted-projection rebuild witnesses pass without changes to authoritative source/history | Exact Projections/assembly/Evidence rows; source manifests alone do not replace rebuild behavior |
| V-046 | C-030 envelope/authorization/route-mode/collaboration witnesses pass | Exact HTTP/module integration rows; preserve existing scenario and event assertions |
| V-047 | `make lint CHECK=backend-boundary`, `make lint CHECK=go`; affected narrow architecture rows | Reject prohibited dependencies/ownership violations; allow harmless file rename/split inside unchanged boundary. If an existing checker changes, extend its focused fixtures only for missing semantic or rename/split cases; no new generic scanner |
| V-048 | `make source-check CHECK=catalog`, `CHECK=schema`, `CHECK=generated`, `CHECK=identity`; focused existing harness contract checks if catalog/checker fixtures change | Full catalog rejects missing, duplicate, stale and build-excluded selectors. Use public `make harness-contract SUITE=contract` or discovered exact-row selection as appropriate; never treat private subset loader success as full coverage |
| V-049 | Regenerate through `make generate` during source/routing implementation; `make agent-finalize` before broader end-of-run verification, then freeze source and qualify the bounded portfolio | Resolve failures before rerunning or broadening; record targets, row IDs, run roots, source/toolchain compatibility and limitations. Supply `RESULTS_DIR=<successful full warm check run root>` for retained maintenance, otherwise report maintenance skipped because `RESULTS_DIR` is unset |

Read installed pins and current public Make help when implementing; do not invent
private runner invocations. Selection expands only for identified ownership/risk,
not to meet an arbitrary test count. After successful narrow checks, broaden only
where needed. Keep failed attempts and repairs in the handoff; a source-changing
repair requires requalification of affected evidence on compatible source.
Historical E-078 remains application-fixture evidence, not a pass for V-042–049.

##### 9. Top-Level Work Tracker

| ID | Work | Status | Evidence / disposition |
| --- | --- | --- | --- |
| T-034 | Historical plan save | DONE | E-079–082 |
| T-035 | WF-31 characterization | DONE | E-084 assertion ledger and baseline |
| T-036 | WF-32 amendments | DONE | E-085 adopted owner edits |
| T-037A | WF-33A straightforward inventories | DONE | E-086 |
| T-037B | WF-33B Entities | DONE | E-087/E-088 effective owner-local enforcement |
| T-037C | WF-33C Indicators | DONE | E-089 |
| T-037D | WF-33D Projections/HTTP | DONE | E-090 |
| T-037 | WF-33 rollup | DONE | All four individual workstream exits |
| T-038 | WF-34 routing reconciliation | DONE | E-091 initial exit; E-092/E-093 findings; E-094 repaired exit |
| T-039 | WF-35 validation and handoff | DONE | E-095 final compatible qualification; E-096 completion |
| T-040 | User-requested full CI and failure evaluation | DONE | E-097; full CI, cleanup and retained validity pass; no failures |

##### 11. Open Questions and Blockers

| ID | Scope | Status | Closure |
| --- | --- | --- | --- |
| RB-013 | Adopted inventory amendments | DONE | E-085 owner edits before protected deletion, authorized by the approved plan and implementation request |
| RB-014 | Mixed assertions/current characterization | DONE | E-084 exact ledger and pre-removal checks; E-087–094 adjustments and E-095 final qualification |

No unresolved bounded-task blocker remains. The failed general-wildcard trial
exposed unrelated Timeline policy drift; that observation remains a separately
scoped follow-up, with no broad scanner or Timeline changes retained (E-087).
A future general matcher change must qualify all affected policies rather than
silently changing their enforcement. The separate WF-33 in the preceding handoff
retains its prior status. Historical blockers retain their historical resolutions.

##### 12. Binary Completion Criteria

| Outcome | Final statement |
| --- | --- |
| Planning and execution | Complete for F-044–049 / D-001–015; remediation, affected areas, rationale, long-term benefit, migration, risk and validation recorded |
| Workstreams | WF-31, WF-32, WF-33A–D, WF-34 and WF-35 DONE; each exit saved before successor entry; WF-34 reopened and reclosed after final qualification exposed a stale fixture |
| Specification | Adopted semantic boundary assurance replaces inventories/counts; capability, security, transaction, replay and storage ownership remain explicit |
| Implementation and migration | Incidental inventories and duplicate selector resolution retired; meaningful package/assembly/transaction checks retained; one real replay successor, four pure row retirements; no production, wire, database or fixture API migration |
| Validation | E-095 bounded portfolio passes; E-097 adds full CI on the same source/toolchain: 422 work units, 2,304 catalog rows, 728 completed cleanup steps and retained validity pass |
| Preservation | Historical archive and unrelated source content preserved; HEAD/staged entries unchanged; generated inode rewrites and raw-index metadata refresh disclosed in E-096 |
| Handoff limits | Full CI passed; no separate full-owner audit, explicit visual screenshot suite, performance publication or release claim. Historical retained-maintenance skip and separate scanner/Timeline follow-up remain recorded |

- [x] Scope and implementation authorization are explicit; historical evidence preserved.
- [x] F-044–049 and D-001–015 have completed dispositions and qualification.
- [x] Owner amendments adopted before protected removal; C-027–031 preserved.
- [x] All workstream exits saved before successor entry, including the repaired WF-34 exit.
- [x] No replacement census, alias, compatibility runner or unnecessary API change.
- [x] Catalog and generated state agree; retired claims have truthful dispositions.
- [x] Finalizer precedes broader final qualification; compatible narrow passes validated against unchanged source.
- [x] Final evidence, cleanup, retained validity, skips and preservation recorded.
- [x] WF-35 validation and handoff completed last.


### Full CI follow-up — E-097

The user requested full CI and evaluation of any failures after WF-35. Executed
from the repository root:

```sh
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-guards-ci-20261027-01 make ci HARNESS_DIAGNOSTICS=basic
```

Exact root: `.cartulary/test-results/tr01-guards-ci-20261027-01`.
Native execution session `48235` returned **exit 0**. The owning
`run-summary.json` reports **pass**, with **422/422 work units passed**, zero
failed/skipped/cancelled units, and null failure class/reason. All **2,304 catalog
row results** passed. The recorded graph duration is 2,189,036 ms (36m 29s), a
functional execution observation rather than a performance acceptance claim.
Cache mode was the normal default: two hits, 198 misses and 222 bypasses; this is
not a claim that every preparation or result was freshly executed.

Source digest: `sha256:ed1f86514097a3d2f1585e5a93c6209c45ef217b2d812ee139be91228acd9ac6`.
Toolchain digest: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
These match the final remediation portfolio. HEAD remains
`005f6dfa196022cfe6c7497cdc16953edfd1cf9b`; the non-Markdown source freeze is
unchanged and no files are staged. The previous successful finalizer still
precedes this broader verification; diagnosis did not trigger another mutating
finalizer or convert the historical RESULTS_DIR-unset maintenance skip to a pass.

Evaluation inspected the public run/child summaries and their structured
artifacts: all 83 target projections report pass, including backend unit/store/
integration/process work, frontend unit checks, functional/stateful/support/
accessibility/measurement browser projections, build, lint, security scanner
operations, Harness contracts, migration drift and conformance. Projections share
units; their counts and durations are not additive. No failing unit or row exists
to diagnose, so no source repair, retry, baseline refresh or deadline change was
performed.

All **728 cleanup records** completed, with no blocked or failed cleanup.
`retained-secret-scan.json` passed (6,515 files). The exact-root command
`make harness-observability-check RESULTS_DIR=.cartulary/test-results/tr01-guards-ci-20261027-01 CHECK=all`
returned **exit 0**, **full_validity=1**. Optional execution diagnostics were
truncated and sampled resource diagnostics were partial (774 omitted
observations); invocation accounting and PostgreSQL checkpoints were unavailable.
Those limits do not invalidate the passing functional evidence and are not
represented as complete resource measurement.

Live publication entered `paused_measurement`; subsequent observation used only
the owning native session until command completion. The final report uses
terminal evidence, not the stale pre-measurement progress snapshot.

Full CI is now qualified, superseding the earlier *not run* limitation for that
specific gate. Explicit visual screenshot validation, release acceptance and
complete-owner evidence audits remain separate scopes. E-087's dormant general
matcher/Timeline policy observation is not resolved merely by this CI pass.
Only this controlling tracker was updated for the follow-up; no production or
test-source changes were made. `make lint-markdown` passed with native exit 0 at
`tr01-guards-ci-markdown-20261027-01`; its summary is
`.cartulary/test-results/tr01-guards-ci-markdown-20261027-01/adhoc/lint-markdown/tool-run-summary.json`.
Because configured Markdown lint excludes this handoff, direct review also
confirmed its section structure, tables, active local links, preserved historical
archive and the unchanged non-Markdown source freeze. `git diff --check` passed.
The full-CI follow-up and failure evaluation are complete.

### WF-35 / S-035 completion and handoff — E-096

All bounded workstreams are complete. E-095 supplies the final compatible evidence;
E-087, E-090 and E-092 retain failures, diagnoses and repairs rather than erasing
them. No production API, wire/event contract, schema, fixture API, compatibility
alias or replacement export-census framework was introduced. Specification,
assertion dispositions, authored routing, generated consistency and final handoff
now agree. Domain vocabulary needed no amendment.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-guards-final-markdown-20261027-01 make
lint-markdown` passed, native exit 0. Summary:
`.cartulary/test-results/tr01-guards-final-markdown-20261027-01/adhoc/lint-markdown/tool-run-summary.json`.
That configured lint excludes this handoff; direct review additionally verifies
its twelve top-level sections, active table shapes, local links, completed IDs,
sequencing, evidence references and historical preservation. `git diff --check`
is clean. Only this tracker changed after that documentation run; direct handoff
checks apply to the final saved version.

Comparison with the captured implementation baseline confirms **35 intended
tracked paths changed plus one new Indicators test file**. All other enumerated
source bytes remain unchanged. Seventy generated entries were recreated with
identical bytes by canonical generation; their inode identities changed, so this
is not a claim of untouched filesystem metadata. The archived historical block
is byte-for-byte preserved. The 7,335-file non-Markdown final source freeze is
unchanged. HEAD remains `005f6dfa196022cfe6c7497cdc16953edfd1cf9b` and staged-entry
digest remains `2d56f2a1f92f7540172830913c49b7a1adb77c51d14c7908bed4ac65884fa8a0`;
no files were staged or committed. Raw index bytes refreshed to
`ec157236c7eb7021d74baf34e37bd9ffd2604aa6c99d690b0b7bed492adfedcf`; staged content
is unchanged. Preservation manifests and native/audit records remain at
`/tmp/cartulary-tr01-guards-9qhwwd5g/`; runnable evidence remains at the exact
repository result roots in E-095.

Finalizer retained-run maintenance was skipped because RESULTS_DIR was unset;
full gate, UI, performance and release exclusions and the unrelated general
matcher/Timeline observation are explicit in E-095/Section 11. Review of normative
projection fidelity remains human governance, not a claim established by tests.
WF-35 validation and handoff is the final completed slice.

### WF-35 final qualification — E-095

The bounded final union passes: **23 exact rows across 12 owners**, three lint
selections, four source-integrity selections, and finalization. Every native
Make exit is 0. Source: `sha256:ed1f86514097a3d2f1585e5a93c6209c45ef217b2d812ee139be91228acd9ac6`.
Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.

The repaired WF-34 graph/selector-negative and script-lint roots remain in this
union because the second finalizer reported `generated.status=unchanged`, zero
updated files, and all roots have the identical source/toolchain identity. They
were not rerun merely to change chronology. The finalizer preceded the broader
remaining qualification; source byte comparison confirms no subsequent code or
routing edits. The first final attempt is superseded, not combined with this
snapshot. Commands use `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run> make ...`.
Roots below are under `.cartulary/test-results/`.

| Owner / exact run | Exact selected row IDs (all passed) |
| --- | --- |
| `app.server` / `tr01-guards-final-server-20261027-02` | `app.server.integration.http_harness_restart_storage` |
| `harness.test_catalog` / `tr01-guards-34-catalog-contract-20261027-02` | `harness.test_catalog.behavior.routing_contract` |
| `module.assessments` / `tr01-guards-final-assessments-20261027-02` | `module.assessments.store.facade_transaction_participant_rollback_and_replay`, `module.assessments.unit.facade_and_import_contracts` |
| `module.entities` / `tr01-guards-final-entities-20261027-02` | `module.entities.integration.the_mention_resolve_route_persists_durable_resol_d68c1befd6`, `module.entities.unit.root_boundary_guards_e292ec1361` |
| `module.evidence` / `tr01-guards-final-evidence-20261027-02` | `module.evidence.integration.a_real_object_store_upload_can_be_created_throug_9924ede0b0`, `module.evidence.integration.attached_evidence_projection_storage_can_be_corr_933714dcfe` |
| `module.imports` / `tr01-guards-final-imports-20261027-02` | `module.imports.unit.module_composition_contract` |
| `module.indicators` / `tr01-guards-final-indicators-20261027-02` | `module.indicators.integration.production_observation_and_lifecycle_routes`, `module.indicators.origin.observation_origin_registry_and_producer_context_40d43e15e1`, `module.indicators.projection.application_mutations_delegate_row_refresh_and__026301a7c8`, `module.indicators.structure.indicator_application_requires_complete_composition_6ecc4edf41`, `module.indicators.target_resolution.safe_http_error_mapping`, `module.indicators.transaction.indicator_repository_writes_roll_back_when_revis_c37659989d`, `module.indicators.replay.stored_create_replay_preserves_bytes_and_identit_6c548da7ff` |
| `module.parties` / `tr01-guards-final-parties-20261027-02` | `module.parties.composition.party_mutation_and_import_composition_exposes_on_8a69574087` |
| `module.projections` / `tr01-guards-final-projections-20261027-02` | `module.projections.architecture.boundaries_and_source_ownership_9b8de61f81`, `module.projections.assembly.catalog_and_source_ownership_927c3f08ce`, `module.projections.foundation.adapter_contract_48f57d6542` |
| `module.tasksdecisions` / `tr01-guards-final-tasksdecisions-20261027-02` | `module.tasksdecisions.composition.mutation_and_import_composition_fails_closed_and_9add8849f4` |
| `harness.command_surface` / `tr01-guards-34-command_surface-20261027-02` | `harness.command_surface.behavior.public_registry_parity` |
| `harness.evidence_accounting` / `tr01-guards-34-evidence_accounting-20261027-02` | `harness.evidence_accounting.behavior.current_epoch_evidence` |

| Check / exact run | Public command | Result |
| --- | --- | --- |
| `tr01-guards-final-go-20261027-02` | `make lint CHECK=go` | pass; native exit 0 |
| `tr01-guards-final-backend-boundary-20261027-02` | `make lint CHECK=backend-boundary` | pass; native exit 0 |
| `tr01-guards-final-generated-20261027-02` | `make source-check CHECK=generated` | pass; native exit 0 |
| `tr01-guards-final-schema-20261027-02` | `make source-check CHECK=schema` | pass; native exit 0 |
| `tr01-guards-final-catalog-20261027-02` | `make source-check CHECK=catalog` | pass; native exit 0 |
| `tr01-guards-final-identity-20261027-02` | `make source-check CHECK=identity` | pass; native exit 0 |
| `tr01-guards-34-scripts-20261027-02` | `make lint CHECK=scripts` | pass; native exit 0 |
| `tr01-guards-finalize-20261027-02` | `make agent-finalize` | pass; native exit 0 |

Every one of these **20 roots** passed `make harness-observability-check
RESULTS_DIR=<exact-root> CHECK=all` with native exit 0 and `full_validity=1`.
All cleanup records completed; retained-secret scans passed; no failed, skipped
or cancelled execution units remain. Every owner root contains exactly its
selected rows, all terminal passed with row exit 0. Optional resource diagnostics
were not enabled; no performance or publication claim is made.

Finalizer retained-run selection, retained-evidence maintenance and run checks
were **skipped because RESULTS_DIR was unset**. No compatible full warm check was
supplied and no redundant full check was created for maintenance. Per-run audits
above are separate. Full CI/check, all-owner acceptance, dedicated race suites,
UI/browser/visual, migration drift, benchmarks and release qualification were not
run: this scoped change alters assurance policy and routing, without production,
wire, fixture API, database or frontend changes. Human review of specification
projection fidelity remains separate from test success.

E-087's general wildcard matcher/Timeline policy drift is a separate follow-up
observation. No broad scanner change was retained. Entities now has its effective
owner-local dependency guard; this effort makes no claim that all unrelated
central patterns were repaired. Documentation review and preservation closure
remain the final WF-35 exit steps.

### WF-34 repaired exit — E-094

Removed only the retired caller-matrix sentinel from the independent tier fixture;
all remaining tier assertions and production routing remain unchanged. A complete
retired-row-reference search now finds only the intentional Indicators replay
crosswalk. The repaired graph/routing contract passed at
`tr01-guards-34-catalog-contract-20261027-02` (all 74 assertions), native exit 0.
The actual selector-negative owners passed at
`tr01-guards-34-command_surface-20261027-02` and
`tr01-guards-34-evidence_accounting-20261027-02`, native exits 0.
`make lint CHECK=scripts` passed at `tr01-guards-34-scripts-20261027-02`, exit 0.
Source digest: `sha256:ed1f86514097a3d2f1585e5a93c6209c45ef217b2d812ee139be91228acd9ac6`.
WF-34 repaired exit saved before resuming WF-35. Finalization and compatibility
checks determine whether these narrow passes remain usable; no retired claim is
reintroduced and no unrelated production change is authorized.

### WF-34 selector-negative qualification correction — E-093

Inspection of the catalog fixture call chain identified a reporting gap in E-084:
`harness.test_catalog.behavior.routing_contract` executes graph/routing closure,
but the existing missing/overlapping/private-selection fixtures execute under
`harness.command_surface.behavior.public_registry_parity`, and stale/duplicate/
build-excluded selector fixtures under
`harness.evidence_accounting.behavior.current_epoch_evidence`. Their source was
inspected earlier; their execution must not be attributed to the graph row.
Public task-guide coverage confirms both selections. Add these two existing rows
to V-048 and final qualification (23 rows, 12 owners); no resolver or new test is
needed. This expansion closes the removed Entities duplicate's assurance at its
actual Harness owners.

### WF-35 qualification failure; WF-34 reopened — E-092

The first final portfolio found one missed authored fixture reference:
`tools/harness/scheduler/work-graph/tests/tier-selection.mjs` still required the
retired HTTP caller-matrix row in its independent retained-row sentinels.
`tr01-guards-final-test_catalog-20261027-01` failed 1 of 74 assertions
(product/test_assertion_failure, unit 10, native Make 2); its cleanup completed.
The public `make explain-run` logs identify `assertHarnessTierRouting` line 119.
This qualifies E-091's closure claim: catalog loading passed, but an independent
routing expectation was stale. No successor row should be fabricated. Retire
that one obsolete sentinel while preserving all remaining independent tier,
projection, closure and negative-selector assertions.

WF-34 is reopened before this repair; WF-35 waits. The other fifteen first-final
roots passed with native exits 0 and retained validity 1, but are not final
compatible qualification after the fixture source changes. The first finalizer
passed at `tr01-guards-finalize-20261027-01`; retained-run maintenance skipped
because RESULTS_DIR was unset. Re-finalize and requalify the bounded union after
the routing repair; do not silently reuse the first snapshot.

### WF-34 / S-034 exit — E-091

Authored selection now retires four pure inventory rows and preserves the one
real Indicators replay successor; every mixed row retains its behavioral claim.
Removed all three obsolete export-minimization verification definitions and
references. No scanner helper/test references or retired selectors remain in
active source/routing. The baseline 25-row characterization becomes a 21-row
final union across ten owners; Artifacts' pure census has no invented successor.
The existing catalog contract fixtures already cover the canonical discovery,
coverage, overlap and build-context cases, so no second resolver was added.

`make format` at `tr01-guards-format-20261027-01` and `make generate` at
`tr01-guards-generate-20261027-01` passed with native exits 0. Generation changed
only `tools/execution_topology_render_index.json` hashes for affected authored
owners/families. `make source-check CHECK=generated|schema|catalog|identity`
(each a separate command) passed at `tr01-guards-34-<check>-20261027-01`, native
exits 0. All commands used `CGO_ENABLED=0`. Roots are under
`.cartulary/test-results`. No generated file was hand-edited. WF-34 exit saved
before WF-35 finalization and compatible final qualification.

### WF-33D / S-033D exit — E-090

Removed Projections internal export/scanner inventories and HTTP exact caller,
capability-field and Server-field inventories. Preserved all real HTTP envelope
and route-mode assertions, adapter constructors, descriptor/table ownership,
source isolation and Evidence rebuild/upload scenarios. Testsupport delegation
is checked across the package, with renamed/split-file fixtures, required actual
production Build calls (no exact count), forbidden direct adapter construction
and forbidden private provider imports. Adapter capability type references are
permitted; they do not reconstruct a graph.

The first Projections run (`tr01-guards-33d-projections-20261027-01`) failed
product/test_assertion_failure (unit 10, outer 2): the broadened import ban
incorrectly rejected testsupport's existing adapter interface field. Published
logs identified `testsupport/capability.go`; no product defect or production
change was needed. Replacing that import ban with a constructor-use check and
retaining provider import rejection fixed the test policy. The second run
(`tr01-guards-33d-projections-20261027-02`) passed architecture, assembly and
adapter-foundation rows, native exit 0. `tr01-guards-33d-evidence-20261027-01`
and `tr01-guards-33d-server-20261027-01` passed the selected rebuild/upload and
HTTP-envelope/route-mode rows, native exits 0. Roots remain under
`.cartulary/test-results`; final compatible requalification follows WF-34.
WF-33D exit saved before WF-34 generation/routing closure.

### WF-33C / S-033C exit — E-089

Retired the Indicators export inventory, predecessor names/files and the observation
constructor-name census. Preserved injected peer/actor composition checks in a
focused boundary test, replaced fixed repository-file transaction checks with a
package-wide borrowed-transaction parameter check, and made sibling imports
package-based. Positive rename/split and negative cross-capability cases qualify
those permissions. Existing constructor, actor, provider and replay behavior stays.
Stored-create byte replay now has its own truthful replay row, with the sole real
successor crosswalk from the former mixed surface row; obsolete minimization
claims are removed.

`tr01-guards-33c-indicators-20261027-01` passed the seven selected structure,
origin, replay, transaction, production-route, safe-error and row-refresh rows
(native exit 0). After the package-permission fixture addition,
`tr01-guards-33c-structure-20261027-02` reports terminal pass for the structure
row; its native process exit was not retained across context compaction, so it
is not represented as an observed outer exit. WF-35 requalifies the final union.
Roots are `.cartulary/test-results/<run>`.

Read-only recovery attempts with missing RESULTS_DIR and undeclared help TARGET
returned usage exit 2; an initial bare result selector found no published run.
Corrected exact-root `make test-run-status` confirmed the structure pass. These
reader mistakes did not execute or change tests. WF-33C exit saved before WF-33D.

### WF-33B / S-033B exit — E-088

Entities now has one effective package-level dependency guard with positive and
negative rename/split fixtures. Removed the exhaustive export ledger, owner-local
selector reconciler and dead helpers, inactive central import duplicate, and
redundant root HostIdentity/WebSocket tests (covered by the package guard).
Retained merge peer restrictions, Workbook/clipboard ownership, cross-owner SQL,
read-only planning and command-level Timeline effects. Route, planning and effect
checks enumerate their owning packages rather than fixed filenames.

`tr01-guards-33b-entities-20261027-01` passed the root-boundary and real resolve
integration rows; after extending filename-independent semantic checks,
`tr01-guards-33b-entities-20261027-02` passed the exact root row.
`tr01-guards-33b-boundary-20261027-02` passed `make lint CHECK=backend-boundary`.
All commands exited 0. The graph catalog fixture pass from E-084 protects the
removed reconciliation; final compatible catalog qualification remains WF-35.
WF-33B exit saved before WF-33C entry. E-087 records the failed scanner trial and
its bounded disposition; no scanner change remains.

### WF-33B enforcement qualification — E-087

The trial general wildcard matcher at `tr01-guards-33b-boundary-20261027-01`
failed (`product/test_assertion_failure`, unit exit 10, outer Make 2). Published
boundary diagnostics identified dormant Entities platform allowlist omissions
and unrelated Timeline import/SQL policy drift. Cleanup completed. The trial
matcher change was removed before retiring any guard; broad Timeline repair is
outside this effort. The failed root remains historical evidence.

The plan's equivalent-owner rule applies: keep one narrow Entities package import
guard and remove the inactive central duplicate. Its admitted package dependencies
include the existing candidate, admission, history and platform primitives, with
explicit negative cases for concrete peers, HostIdentity, WebSocket and application
assembly. File reorganization has no policy meaning. This avoids widening the
scanner or preserving a second ineffective policy. Other central boundary rules
remain unchanged. E-084's intended central destination is superseded by this
qualified owner-local consolidation.

Two diagnostic queries initially used unsupported LIMIT/internal TARGET inputs;
those readers exited 2 without executing tests. `explain-run ... TARGET=lint
DETAIL=logs` resolved the actual published violations.

### WF-33A / S-033A exit — E-086

Removed the six straightforward inventory/scanner files (D-001/002/005/006/009/010),
three standalone inventory rows and Artifacts/Evidence minimization claims. Mixed
composition/behavior rows remain. No production code or fixture API changed.
Focused owner runs `tr01-guards-33a-{assessments,imports,parties,tasksdecisions}-20261027-01`
and `make source-check CHECK=catalog` at `tr01-guards-33a-catalog-20261027-01`
all exited 0 and reported pass. Commands use the surviving E-084 row selections
(Imports excludes its retired census row); all roots are under `.cartulary/test-results`.
This checkpoint is saved before WF-33B starts. Generation consistency is qualified
in WF-34; no generated artifact was hand-edited.

### WF-32 / S-032 exit — E-085

Adopted the user-approved structural-assurance amendments in the Entities,
Indicators and Workbook Projections decisions and Core 04 AC-558/560. Removed
mandatory exhaustive export dispositions/counts and incidental implementation
snapshots; retained semantic capability, dependency, security, transaction,
replay and storage requirements. Core 00 and AC-539 are unchanged. Domain
vocabulary is unchanged. Entities exact selector completeness is explicitly
Harness-owned. No production/test/routing source changed in this workstream.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-guards-spec-markdown-20261027-01 make
lint-markdown` exited 0, summary pass at
`.cartulary/test-results/tr01-guards-spec-markdown-20261027-01/adhoc/lint-markdown/tool-run-summary.json`.
Direct diff review and `git diff --check` passed after removing an extra EOF
blank line. The configured lint excludes this handoff; direct review supplements
it. User approval of the concrete proposed amendments followed by the instruction
to implement is the adoption authority; tests do not adopt specifications.
RB-013 is closed. WF-32 exit is saved before WF-33A begins.

### WF-31 / S-031 exit — E-084

Assertion/helper inspection and pre-removal characterization are complete. All
eleven owner invocations exited 0; summaries pass, cleanup outcomes completed,
retained-secret scans pass. Source: `sha256:4bef6ffbe421cf8ccce5009e0412ada22a9da4bbc77615d7bec8d79446a5be30`.
Every command used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run> make test-slice
OWNER=<owner> ROWS=<listed rows>`. Roots are `.cartulary/test-results/<run>`.

| Owner / run suffix | Exact selected rows |
| --- | --- |
| `app.server` / `tr01-guards-base-server-20261027-01` | `app.server.integration.http_harness_restart_storage` |
| `harness.test_catalog` / `tr01-guards-base-test_catalog-20261027-01` | `harness.test_catalog.behavior.routing_contract` |
| `module.artifacts` / `tr01-guards-base-artifacts-20261027-01` | `module.artifacts.surface.iteration3_exported_surface_reachability` |
| `module.assessments` / `tr01-guards-base-assessments-20261027-01` | `module.assessments.store.facade_transaction_participant_rollback_and_replay`, `module.assessments.unit.facade_and_import_contracts` |
| `module.entities` / `tr01-guards-base-entities-20261027-01` | `module.entities.integration.the_mention_resolve_route_persists_durable_resol_d68c1befd6`, `module.entities.unit.root_boundary_guards_e292ec1361` |
| `module.evidence` / `tr01-guards-base-evidence-20261027-01` | `module.evidence.integration.a_real_object_store_upload_can_be_created_throug_9924ede0b0`, `module.evidence.integration.attached_evidence_projection_storage_can_be_corr_933714dcfe`, `module.evidence.surface.exported_surface_reachability` |
| `module.imports` / `tr01-guards-base-imports-20261027-01` | `module.imports.unit.export_surface_contract`, `module.imports.unit.module_composition_contract` |
| `module.indicators` / `tr01-guards-base-indicators-20261027-01` | `module.indicators.integration.production_observation_and_lifecycle_routes`, `module.indicators.origin.observation_origin_registry_and_producer_context_40d43e15e1`, `module.indicators.projection.application_mutations_delegate_row_refresh_and__026301a7c8`, `module.indicators.structure.indicator_application_requires_complete_composition_6ecc4edf41`, `module.indicators.surface.iteration2_exported_surface_reachability`, `module.indicators.target_resolution.safe_http_error_mapping`, `module.indicators.transaction.indicator_repository_writes_roll_back_when_revis_c37659989d` |
| `module.parties` / `tr01-guards-base-parties-20261027-01` | `module.parties.composition.party_mutation_and_import_composition_exposes_on_8a69574087` |
| `module.projections` / `tr01-guards-base-projections-20261027-01` | `module.projections.architecture.boundaries_and_source_ownership_9b8de61f81`, `module.projections.assembly.catalog_and_source_ownership_927c3f08ce`, `module.projections.foundation.adapter_contract_48f57d6542`, `module.projections.testsupport.projection_capability_caller_matrix_6f21615b92` |
| `module.tasksdecisions` / `tr01-guards-base-tasksdecisions-20261027-01` | `module.tasksdecisions.composition.mutation_and_import_composition_fails_closed_and_9add8849f4` |

The earlier `tr01-guards-baseline-20261027-01` ROWS-only invocation was rejected
before setup (`config/usage_error`, outer 2): the public target requires OWNER or
SUITE. It is not test evidence. Owner-scoped invocations above correct the input.

The backend boundary baseline also exited 0 at `tr01-guards-base-boundary-20261027-01`.
Inspection nevertheless found that the existing matcher treats `/*.go` literally:
the Entities central allowlist is inactive, in addition to omitting candidates,
mutationadmission and Revisions historycontract. WF-33B must qualify direct-file
pattern enforcement and rename/split fixtures before retiring the local guard.

Final assertion dispositions beyond D-001–015:

| Assertion / helper closure | Decision and surviving assurance |
| --- | --- |
| Six straightforward export files; parser/synthetic fixtures, Tasks/Decisions provider filename maps and historical symbol scans | Retire completely; no external helper consumers; retain constructor/import/provider and actual mutation tests in their mixed families |
| Entities root file-specific import map and root HostIdentity/ws restrictions | Central package boundary after matcher and permitted-import reconciliation; retain merge peer prohibition, clipboard/row-route ownership, SQL mutation/planning and Timeline effect port semantics |
| Entities export/selector reconcilers and private inventory maps | Retire; canonical catalog negative fixtures cover missing/stale/duplicate/build-excluded and source-coverage cases, including Markdown exclusion |
| Indicators predecessor scanner | Extract no peer construction and no fabricated actor guarantees; retire old symbols/files and registration call count, qualified by actual HTTP workflows |
| Indicators application repository filename checks | Retire old-file absence and lexical workflow-word prohibitions; retain borrowed transaction ownership as a function-level check independent of filenames. Existing transaction rollback and projection delegation rows qualify workflow effects |
| Indicators fixture minimal-field/helper checks | Retire; current examples contain only immutable strings and value times, with no mutable shared reference; no redundant new layout test |
| Indicators origin test using exportedRootDeclarations | Retire historical removed-export name list and now-unused scanner; retain closed origin parser and caller-controlled-field admission checks in existing origin row |
| Indicators provider_encapsulation_test | Retain imports and all contribution readiness assertions; these are semantic, not a symbol census |
| Projections testsupport Build assertion | Replace filename and exact-one-call constraints with package-wide dependency/delegation evidence and focused rename/split and forbidden-construction fixtures |
| Projections facade/provider and SQL/table manifests | Retain meaningful approved imports, physical ownership, descriptor identity/immutability and functional rebuild witnesses |
| HTTP caller matrix and Server field census | Retire; retain every boot/envelope/route-mode assertion and preserve fixture API |
| Shared helpers | Only extra export-helper consumer is the origin test above. All other deleted scanner helpers are file-local; imported package/global names are checked again during Go compilation |

No new catalog resolver or fixture framework is needed. WF-31 exit is saved;
WF-32 may now begin. RB-014 is closed for these concrete dispositions.

### Implementation baseline — E-083

The user authorized implementation of the proposed plan. HEAD and the only dirty
path match the planning baseline. Snapshot of 7,653 tracked/nonignored entries,
HEAD/index and original handoff: `/tmp/cartulary-tr01-guards-9qhwwd5g/`.
Read-only help and eleven task guides passed; no historical pass qualifies this
implementation. The refactor-tracker planning boundary has been superseded by
implementation authorization. Test operations follow the test-operations skill.
WF-31 is IN_PROGRESS; no owner/test/routing changes yet. Initial tracker editing
hit an overly broad heading-count assertion against historical nested headings;
anchoring the heading fixed the editor before any history was changed.

### Current planning save — E-079–E-082

**E-079 — Baseline and concurrency, 2026-10-27.** The prior read-only inspection
left working-tree contents unchanged while the index changed through concurrent
staging. During save preparation the earlier work was committed. The retained
save baseline is clean `main` at `005f6dfa196022cfe6c7497cdc16953edfd1cf9b`.
The handoff and scoped structural guards, catalog resolver and adopted owner
sources were unchanged across those commits; their inspected conclusions remain
applicable. Baseline fingerprints and the original handoff are retained externally
at the Section 1 path. No staging/reset/commit action is part of this save.

**E-080 — Inspection and design.** Read applicable root/backend/tooling guidance,
the planning framework and refactor/test-operations skills; inspected the source,
owner and routing groups in Section 2. Read-only `rg`, file reads and Git comparisons
established F-044–049 and D-001–015; public Make help was used to discover future
commands. No test command executed. Exact assertion-to-row closure and new gap
characterization are explicitly future WF-31 gates, not claimed completed tests.
The archive below supersedes the previous active summary without changing its
historical IDs, decisions or result log.

**E-081 — Documentation validation, PASS.** The following admitted command exited
0; its summary reports `status=pass`, `failed=0`, with no tests executed:

```bash
GIT_OPTIONAL_LOCKS=0 \
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-structural-guards-plan-4m_6t9_z/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-structural-guards-plan-4m_6t9_z/static-analysis \
CARTULARY_TEST_RUN_ID=tr01-structural-guards-plan-20261027-01 \
COREPACK_ENABLE_NETWORK=0 HARNESS_DIAGNOSTICS=off HARNESS_ACCOUNTING=off \
make lint-markdown
```

Retained run root:
`/tmp/cartulary-structural-guards-plan-4m_6t9_z/results/tr01-structural-guards-plan-20261027-01`;
summary: `adhoc/lint-markdown/tool-run-summary.json`. Installed-only readiness,
network-disabled Corepack, no autofix, external result/cache roots and disabled
diagnostics/accounting avoided bootstrap, generation and product verification.
Configured lint excludes this handoff. Direct review separately verified twelve
ordered sections, 19 active local links, 15 active tables, new identifier
definitions, 275 retained historical IDs and the acyclic conditional dependency
graph. `git --no-optional-locks diff --check -- docs/handoffs/testing-remediation-01.md`
also passed. External `direct-review.json` records the structural checks; an
initial review-script assertion counted literal HTML tags in historical prose
and was corrected to count actual markup lines. No repository fix was needed.
Evidence-record updates were re-reviewed directly; no product assurance is claimed.

**E-082 — Preservation comparison, PASS.** Immediately before writing, all 7,653
baseline entries, the admitted output's identity/bytes and the clean index/HEAD
matched E-079. The final comparison verifies unchanged content and lstat identity
for every entry outside the handoff, unchanged file universe, HEAD, branch, raw
index and staged entries, and only this handoff modified. Its inode/link/mode
identity remains unchanged. External `prewrite-preservation.json` and
`final-preservation.json` retain the comparisons. The previous session-log body
is byte-for-byte intact (SHA-256
`ee8dacc9122345459e6bdca7ef43143d187c2402583a61f843c87c94e2935119`);
the former active summary/completion text is preserved with heading demotion in
the historical snapshot. Ignored output trees were not exhaustively fingerprinted;
inspected command effects and external output locations bound that limitation.
No tests, generation, finalization, source checks or product verification ran,
as required by this document-only action. V-042–049 remain future obligations.

### Historical completed application-fixture iteration

The following snapshot and earlier log are historical. Their authorizations,
readiness states and completion claims apply only to their stated baselines.
They do not authorize implementation or adopt amendments in this iteration.

<details>
<summary>Completed WF-26–WF-30: historical active-plan snapshot</summary>

#### Testing Remediation 01: Application Fixture Simplification

##### 1. Scope and Source Posture

Document date: 2026-10-27. **Current action: WF-26–WF-30 implementation and handoff complete.**
The user approved the proposed application-fixture remediation plan and requested
implementation. The bounded source, caller, test and authored-routing changes below
are authorized. Preserve pre-existing handoff edits and historical evidence. Save
each workstream exit before entering the next; qualification and handoff finish last.

Repository: `/home/jochi/code/cartulary`; established supplied and canonical target
`.`; kind: repository directory; iteration label: `application-fixture-simplification`.
The controlling tracker is `docs/handoffs/testing-remediation-01.md`, a regular,
singly linked file with no symlinked parent component. The repository identity is
unchanged. The [modular refactor framework](cartulary_modular_refactor_planning_framework.md)
and Refactor Tracker format supply planning structure, not behavioral authority.
Artifact mode is **normal for the bounded iteration design**.

The implementation source boundary is application-fixture composition in
`internal/testutil/appsupport`, its constructor-only Auth/Timeline wrappers,
affected caller setup, and authored verification routing. Keep `httptestx` as the
HTTP/runtime composition boundary, `pgtest` and `s3test` as service-fixture owners,
and the existing broker as physical-service owner. Production behavior, provider
internals, service-broker/scheduler redesign, browser presentation, performance
policy and unrelated test-suite cleanup are excluded. Domain scenario helpers
remain with their owners. No generic fixture framework, public injection API,
compatibility alias, dual path or additional runner is planned.

WF-26 → WF-27 → WF-28 → WF-29 → WF-30 is the implementation sequence in **this** document.
WF-00–WF-25 and E-001–E-068 retain their historical source baselines and results.
E-069–E-072 record this refinement's baseline, inspection, documentation checks
and preservation comparison. E-073 onward record this authorized implementation.
The separate open WF-33 in [the preceding handoff](testing-remediation.md) is
neither a prerequisite nor resumed or closed here. Historical authorizations,
commands and completion statements in Section 10 are not current instructions.

| Authority / status | Applicable requirement and boundary |
| --- | --- |
| [Core 00](../spec/00_document_set_status_and_precedence.md), current normative allocation | Adopted owners govern requirements; tests and routing are projections. No executable test, generator, runtime metadata, conformance or release evidence may consume Markdown. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md), `adopted/current` | REQ-410: complete explicit, sorted dependency admission before service acquisition. REQ-804: explicit attachment, independent physical owners and recoverable cleanup. §11.0: dedicated databases, test-owned namespaces and retained explicit package reuse. §§3/8/9: exact selection, truthful evidence and failure classification. REQ-480/§12: explicit test-route modes selected by the behavior under test. |
| [Core 01](../spec/01_architecture_storage_and_view_contracts.md), current; §§11–13, §16 and §17.2 | Preserve Reference Pack state, portability/Recovery, background-job behavior, Evidence access and Imports semantics at their existing owner boundaries. Caller migration does not redefine those guarantees. |
| [Core 04](../spec/04_security_deployment_and_conformance.md), current; §1.3 and authorization owners | Preserve verified service transport, purpose-specific database admission, authorization and secret confinement. Tool versions remain in the pin registry. |
| [Reference Pack NLSpec](../reference-pack-subsystem-nlspec.md), `adopted/current`; [Domain](../domain.md) | Retained authoritative pack bytes and instance-scoped work remain meaningful witnesses. Domain supplies vocabulary/navigation, not fixture API requirements. |
| [Core 05](../spec/05_claim_publication_and_benchmark_reproducibility.md) | No timing, benchmark, release or publication claim. Functional evidence cannot stand in for distinct publication qualification. |

The harness permits unique test-owned namespaces while retaining explicit
package reuse. Changing appsupport's default does not remove that separate
capability. No inspected owner requires the current constructor family, mutable
global substitutions, unconditional S3 acquisition, or package-bucket serialization
for every application test. No normative amendment is proposed to preserve them.

This table preserves the planning baseline. The implementation baseline is E-073;
its authorized write boundary and final comparison are recorded at E-078.

| Baseline item | Starting evidence / identity | Final comparison / limits |
| --- | --- | --- |
| Branch / commit | `main`; `e6ed547531945b62d3006ae8c1d458746ddf4670`; handoff already modified | Same source commit as the preceding planning turn; E-069; final comparison E-072 |
| Index / staged state | No staged edits; raw-index SHA-256 `547731b07f82b3dfbb7035c53148d37819e29e4fa4bcbdf189aa1f073598abae`; staged-entry SHA-256 `e151b74207a9bae401c77cbc98dfba04b011e842a3acd77c3c63f87c68356ed2` | E-069; final raw-index and staged-entry comparison E-072 |
| Worktree / untracked | Only this handoff modified; no untracked files; 7,654 tracked/nonignored entries individually fingerprinted with path identities | Preserve the existing handoff delta; E-069 was documentation-only; E-073 establishes the authorized implementation baseline |
| Tracker | 319,808 bytes; device/inode `2096:348826`; one link; SHA-256 `7720584c4ba3cb520687093d1cad49cc03b7000a70ffbd64a1baa61c8ac1653f` | Matched the preceding read-only planning turn; identity/content rechecked before writing |
| Comparison material | `/tmp/cartulary-fixture-plan-refinement-lz9muvcb/baseline.json` and `original-handoff.md` | External preservation material; not executable product evidence; earlier E-065–E-068 material remains historical |
| Ignored roots | Build/cache/dependency/result trees excluded from behavioral enumeration | No exhaustive ignored-tree fingerprint; admitted documentation outputs use external paths |

##### 2. Baseline Repository Inventory

The universe is the 7,654 tracked and non-ignored untracked entries under `.`
recorded in E-069. Generated/vendor/binary and other unrelated entries receive
metadata-only accounting; symlink entries are not followed outside the repository.
Ignored outputs are excluded. The table distinguishes direct source inspection
from caller discovery; neither constitutes executed assurance.

| Path / group | Disposition / inspection depth | Responsibility, callers and dependencies | Owner / risk / evidence |
| --- | --- | --- | --- |
| `internal/testutil/appsupport/runtime.go`, `runtime_guard_test.go` | Directly inspected in full | Composite acquisition, constructor variants, per-database pack root, borrowed dependencies, global substitutions and two guard tests | appsupport; F-038–F-040/F-043; C-022–C-026; E-066 |
| `appsupport/commit_fault.go`, `commit_fault_test.go`, `store.go` | Directly inspected in full | Validated incident fault capability and separate store-only fixture path | Keep semantics and store-only path; C-022/C-025; E-066 |
| Auth `testsupport/flowtest/runtime.go`; Timeline `testsupport/scenariotest/harness.go` | Directly inspected in full | Embedded runtime, constructor forwarding, type aliases and explicit route defaults | Remove constructor-only layers; retain domain helpers; F-042; E-066 |
| `internal/testutil/httptestx/httptestx.go`, `httptestx_test.go`; `internal/app/server/runtime_test.go` | Directly inspected relevant setup, route mode, close and borrowed-store assertions; other bodies metadata-only | Lower composition boundary, production borrowed/owned lifetime, observation capabilities | Preserve production assembly and existing meaningful witnesses; C-022/C-025 |
| `internal/testutil/suiteservices/dependencies.go`, `dependencies_test.go` | Directly inspected in full | Shared declaration parser, typed failure and malformed-assignment tests | Reuse owner instead of copying validation matrices; C-023 |
| `internal/testutil/s3test/s3test.go`, attachment entrypoint; pgtest entrypoints | Directly inspected affected attachment, namespace, package-reuse and cleanup bodies; remainder metadata-only | Existing attachment cache, fresh bucket primitive, explicit package reuse and database preparation | Consumers change; provider mechanics do not; C-023/C-025 |
| Incident Bundles `routes_helpers_integration_test.go`, round-trip callers; Reference Data `worker_isolation_integration_test.go`, `reference_pack_test_helpers_test.go` | Affected setup and meaningful isolation/round-trip assertions directly inspected; unrelated bodies metadata-only | Duplicate fresh namespace/store composition to avoid package lease contention | Replace generic setup; preserve domain setup and independent workers/round trips; F-041 |
| Evidence `blob_verification_integration_test.go`, `attachment_authorization_integration_test.go`, `objectstore_dependency_test.go`; Imports `imports_integration_test.go` | Cancellation scenario inspected; relevant adapter replacement and crash-recovery setup/assertions directly inspected; unrelated bodies metadata-only | Filesystem stores, read cancellation/drain, dependency error mapping, restart with one database/store | Required production witnesses, not blanket permission to rewrite scenario assertions; C-022/C-024 |
| Recovery `backup_metadata_integration_test.go`; Reference Data restart/setup consumers | Affected direct service access, durable root and storage assertions inspected; remaining bodies metadata-only | Explicit real-service operations and independent backing storage | Migrate access without replacing real provider evidence; C-024/C-025 |
| Other caller search hits | Metadata-only discovery; individual migration/coverage ledger is WF-26 work | 192 direct appsupport import files across 35 packages; not all use constructors. Search found 73 `StartDefaultServer` calls in 30 files, three database variants, one store variant, three combined variants and four database aliases | Counts are inventory, not a removal target or complete transitive caller audit; E-066 |
| Authored families, verification owners, catalog/task/topology and generated policy | Relevant declarations/selectors directly inspected; remaining entries metadata-only | Existing support row combines fixture admission with incident commit-fault checks; affected real rows declare both services | Owner-first routing changes only during future implementation; F-043/C-026 |
| Instructions, framework, adopted clauses, Make/lint path and configuration | Applicable sections directly inspected | Planning authority and documentation-only command admission | E-065–E-067 |
| Refinement source subset | Reinspected construction/guards, Auth/Timeline forwarding, Auth route contributions, HTTP close/admission, suiteservices validation, fresh/package bucket lifecycle, second-server workarounds, filesystem-backed cancellation/recovery, Recovery service access and selected authored rows | E-070 adds the resource matrix, ownership distinctions and consolidation rules; prior direct inspection is retained at its stated depth, not expanded into a complete caller audit | C-022–C-026; E-070; instructions/Make/lint path revalidated for E-071 |
| Remaining enumerated entries | Metadata-only; no movement proposed | Outside the named composition seam; no behavioral-completeness claim | Full path/content accounting is retained in `baseline.json` |

WF-26 must trace direct and transitive constructor consumers, method promotion,
type aliases and direct service-field access, inspect each affected setup before
editing it, and map each affected test to exact current rows. Discovery is not
limited to import-text hits. This is a specified entry obligation, not a claim
that all 192 files have had their assertions audited. Newly encountered provider,
production or measurement changes require a separate bounded decision.

##### 3. Module Boundary Diagnosis

The seam is reusable application-test composition, not another service owner.
Retain the existing layers and give callers one options-based construction path.

| Responsibility | Current location | Disposition / destination | Finding / contract |
| --- | --- | --- | --- |
| Parent test lifetime and per-database Reference Pack root | appsupport.Runtime | Keep as a local composition scope; no managed-service acquisition in StartRuntime | F-039; C-024 |
| Constructor argument translation | appsupport variants and Auth/Timeline embedding wrappers | Consolidate into package and Runtime StartServer methods using ServerOptions; remove forwarding aliases | F-038/F-042; C-022 |
| Service declaration admission | appsupport guard plus suiteservices | Keep shared guard; select the complete sorted required set before either service is touched; distinguish service use from fixture allocation | F-039/F-040; C-023 |
| Physical services and fixture primitives | Existing broker, pgtest, s3test | Keep ownership; call existing attachment/database/namespace primitives | C-023/C-025 |
| Fresh application object namespace | Package bucket default plus caller workarounds | Use an independently owned bucket when appsupport creates a store; remove duplicate consumer setup | F-041; C-024/C-025 |
| Domain setup and production assertions | Module test support and scenarios | Keep with domain owners; preserve explicit route modes, failure injection and behavioral assertions | C-022/C-024 |
| Coverage/evidence organization | Authored families and verification owners | Separate fixture admission from incident fault admission; truthful service declarations | F-043; C-026 |

##### 4. Public Contract and Behavior Freeze Map

Production guarantees and necessary harness invariants are preserved. Existing
helper names, source layout, global patching, redundant setup and accidental
serialization are not compatibility contracts. No production route, wire schema,
WebSocket event, authorization policy, revision/projection behavior or UI selector
change is proposed; affected domain tests remain witnesses of their owners.

| Contract | Observed behavior and owner-required guarantee | Disposition / witnesses | Slice / validation |
| --- | --- | --- | --- |
| C-022 — application/security composition | Existing wrappers encode disabled or harness-owned routes and pass environment, dependencies and additional routes. Core 01/Core 04 own product behavior; Harness REQ-480/§12 owns test-route admission. | Preserve configuration precedence, bootstrap, capabilities and validated fault admission. Select explicit TestRouteMode by the exercised route/contribution; a former wrapper default is not an independent requirement. Preserve meaningful Auth route contributions. | S-027/S-029; V-035/V-037/V-038 |
| C-023 — dependency admission | StartRuntime currently requires both services, even for supplied stores. REQ-410 requires every actually needed service before acquisition; REQ-804 requires explicit attachment. | Separate harness behavior change: guard the complete sorted set before acquisition using the Section 7 matrix. Supplied databases avoid allocation/fixture attachment but still require PostgreSQL runtime access; supplied stores avoid appsupport object-store acquisition. | S-028; V-036/V-038 |
| C-024 — isolated and durable test state | Package bucket serialization causes custom simultaneous-server setup. Parent Runtime stores Reference Pack bytes by database. Core 01 §§11–13/17.2 and the Reference Pack owner govern durable effects and restart behavior. | Fresh default DB/bucket per application; explicit database/store reuse for restart. Preserve Imports exactly-once recovery, Evidence cancellation/drain, Reference Pack worker isolation, portability and Recovery storage witnesses. | S-027/S-028; V-035/V-036/V-038 |
| C-025 — ownership and cleanup | Supplied stores/databases are borrowed; created clients and runtime handles are test-owned; server Options borrows injected resources. Harness REQ-804/§11 and backend assembly ownership require correct owner lifetimes. | Never close a supplied store, drop a supplied database or terminate a physical owner. Stop app work, close appsupport-owned handles/clients, then delete only its owned fixtures. Preserve ownership on partial startup failure and retain cleanup uncertainty; keep TLS/role admission and explicit package reuse. | S-028; V-036/V-038 |
| C-026 — exact routing and truthful assurance | Support row mixes two responsibilities; two known filesystem-only scenarios still request S3. Harness §§3/8/9 own selection, evidence and classification. | Separate coherent rows, preserve independent real-provider checks, change authored inputs first and regenerate. A supplied S3-backed store still needs its row's service claim. | S-029/S-030; V-037/V-038 |

Pre-move execution and the complete caller/assertion disposition ledger are
recorded at E-073/E-074. Final qualification is recorded at E-078; selected coverage
does not establish full-owner acceptance or normative completeness. No primary-owner
contradiction was found in this bounded design scope.

##### 5. Coupling and Boundary Findings

`must_fix` applies only to the named iteration's acceptance conditions.

| Finding / classification | Evidence and consequence | Required action / retirement decision | Contract / slice / validation |
| --- | --- | --- | --- |
| F-038 / must_fix | runtime.go repeats ServerOptions construction through several exported variants; some have no external callers | Keep one options shape and two lifetime-appropriate entrypoints; remove variants and PrepareServerDatabase without forwarding stubs | C-022; S-027; V-035/V-038 |
| F-039 / must_fix | StartRuntime always acquires PostgreSQL/S3; Evidence cancellation and Imports recovery supply filesystem stores | Apply the four-case resource matrix; validate complete sorted dependencies before acquisition; avoid allocation of supplied resources and remove the two unnecessary S3 row claims | C-023/C-024; S-028/S-029; V-036–V-038 |
| F-040 / must_fix | runtime_guard_test.go saves/restores mutable package starter functions | Use private per-call starter dependencies and an error-returning admission boundary; consolidate assertions without global mutation | C-023/C-025; S-028; V-036/V-038 |
| F-041 / must_fix | PreparePackageBucketT lease lasts until test cleanup; Incident Bundles and Reference Pack worker tests hand-build a second store/bucket | Use existing fresh namespace primitive for appsupport defaults; remove workarounds and retain actual simultaneous-instance/isolation assertions | C-024/C-025; S-028; V-036/V-038 |
| F-042 / should_fix | Auth/Timeline runtime wrappers mostly embed/alias appsupport and translate constructor parameters | Migrate callers with explicit owner-aligned route modes and required contributions; delete constructor-only wrappers/aliases; keep semantic scenario/login helpers | C-022; S-027; V-035/V-038 |
| F-043 / must_fix | Fixture admission is selected by a row named for incident commit-fault admission; service declarations mirror old setup | Give fixture acquisition a semantic unit group; retain incident fault admission separately; exact coverage and capability declarations must agree | C-026; S-029; V-037/V-038 |

###### Remediation rationale, migration and regression risks

These dispositions are implemented; their final qualification is E-078. Normative
specifications need no amendment: the adopted requirements already govern the
chosen admission, isolation, route and ownership semantics. Tests and routing
remain projections subject to human fidelity review.

| Gap | Areas changed | Rationale and long-term benefit | Compatibility / migration impact | Risk prevented | Completion evidence |
| --- | --- | --- | --- | --- | --- |
| F-038 | Implementation, callers, tests, API documentation/handoff | One explicit options shape eliminates divergent forwarding paths and makes future options local to one constructor | Internal helper API deliberately breaks; all caller paths migrated, with no compatibility aliases; production API unchanged | Wrapper drift, lost options and growing overload burden | E-073 caller closure; compilation/lint and all affected rows at E-078 |
| F-039 | Implementation, tests, authored routing, handoff | Resource needs derive from supplied fixtures before any acquisition; filesystem scenarios no longer depend on an unused service | StartRuntime becomes local-only; missing fixtures acquire services at StartServer; two filesystem rows become PostgreSQL-only | Hidden infrastructure coupling, unnecessary failures and partial acquisition before rejection | Four-case admission matrix, each omitted dependency, and real filesystem restart/cancellation selections |
| F-040 | Implementation, unit tests, handoff | Private per-call starters make admission deterministic and permit parallel tests without global restoration | Removes private global substitution hooks; no new public injection API | Order dependence, races and test state leaking across cases | Parallel admission cases prove exact allocation counts, identity and rejection before acquisition |
| F-041 | Implementation, caller setup, integration tests, handoff | Existing independent namespace primitive makes simultaneous applications composable while preserving provider ownership | Defaults allocate a fresh bucket per app; persistence requires explicit borrowed fixtures and parent Runtime; explicit provider package reuse remains available | Package-lease deadlock, sibling contamination and custom duplicated fixture setup | Two live applications, database/object absence and isolation, child fixture retirement, sibling survival and durable restart bytes |
| F-042 | Implementation/callers, tests, handoff | Removing constructor-only wrappers keeps domain helpers focused on domain setup and makes security mode visible | Auth retains its route contribution and harness-owned mode; ordinary Timeline scenarios use disabled mode; wrapper types removed | Hidden defaults, redundant abstractions and accidental route admission changes | Auth/Timeline affected unions, HTTP route-mode and incident fault admission witnesses |
| F-043 | Tests, verification contracts, authored/generated routing, handoff | Cohesive rows make selection and assurance truthful; fixture mechanics use harness/support classification | Fixture admission receives a new semantic row; incident row retains its identity and guarantee; downstream index regenerated | Duplicate or missing execution, misleading assurance and excessive service claims | Generated/schema/catalog/identity checks, exact row-set accounting and retained-evidence validation |

Do not replace these shortcomings with a service-selection policy language,
public dependency-injection API, package-lock workaround or compatibility mode.
Standalone process tests, broker lifecycle tests, source-shape guards unrelated
to the affected seam and performance fixtures are deferred: this inspection does
not establish that they are redundant. No test-count or line-count quota applies.

##### 6. Refactor Workstreams

This table is the authoritative prerequisite graph. Current progress is recorded
in Section 9 and the saved implementation checkpoints in Section 10; historical
passing roots do not qualify a new source baseline.

| Workflow | Class / prerequisites | Goal / likely boundary | Validation | Binary exit and checkpoint |
| --- | --- | --- | --- | --- |
| WF-26 — characterization | root; user implementation authorization and E-069/E-070 revalidation | Complete affected caller/assertion dispositions; exact selection and current baseline | V-034 | Every proposed migration has an inspected setup, owner-required guarantee or retirement rationale, exact row mapping and classified baseline outcome |
| WF-27 — constructor consolidation | chain; WF-26 exit | Options-based API, appsupport callers and constructor-only Auth/Timeline wrappers | V-035 | Retired variants/aliases have no live callers; route modes, config and restart inputs preserved; focused caller checks pass |
| WF-28 — acquisition and isolation | chain; WF-27 exit | Per-call service admission/dependencies, local Runtime scope, independent namespaces and direct-field migration | V-036 | Required-service, ownership, simultaneous-server and restart witnesses pass; globals and custom generic store setup removed |
| WF-29 — tests and routing | chain; WF-28 exit | Consolidated admission assertions; authored families and generated projections | V-037 | Coherent exact selectors and truthful service sets; no dangling, duplicate or alias rows; generation/static checks pass |
| WF-30 — qualification and handoff | chain; WF-29 exit | Finalize, source freeze, affected coverage portfolio and evidence review | V-038 | Required compatible coverage, native exits and cleanup/evidence checks pass; scope and skips recorded |

Save each exit before entering its successor. Minimal selector updates required
to keep a renamed test executable accompany its owning slice; WF-29 owns the final
semantic grouping and dependency cleanup. This avoids unselectable intermediate
checkpoints without creating compatibility rows. There are no optional or parallel
workflows and no cyclic or external WF-33 dependency.

##### 7. Refactor Slice Plan

| Slice / kind | Findings / contracts | Depends on / readiness | Intended change and entry evidence | Validation / rollback / binary completion |
| --- | --- | --- | --- | --- |
| S-026 / characterization | F-038–F-043; C-022–C-026 | E-069/E-070; DONE, E-073/E-074 | Finish bounded ledger and exact selection; characterize guarantees, not accidental constructor behavior | V-034; no production move; complete WF-26 exit |
| S-027 / structural preservation | F-038/F-042; C-022/C-024 | S-026; DONE, E-075 | Migrate API/callers atomically with explicit options and owner-aligned route semantics; remove redundant constructors and wrappers | V-035; revert coordinated API/caller/selector changes if assurance gaps appear; complete WF-27 exit |
| S-028 / separate harness behavior change | F-039–F-041; C-023–C-025 | S-027; DONE, E-076 | Change acquisition timing, default namespace lifetime and direct service access; retain production guarantees and existing physical owners | V-036; revert the complete fixture/caller delta, not a legacy mode; complete WF-28 exit |
| S-029 / test/routing consolidation | F-043; C-022/C-026 | S-028; DONE, E-077 | Separate fixture admission from incident fault checks; correct dependencies and regenerate authored projections | V-037; revert authored inputs and regenerate together; complete WF-29 exit |
| S-030 / qualification | All | S-029; DONE, E-078 | Finalize then freeze and qualify the necessary affected portfolio | V-038; unresolved failure prevents acceptance; complete WF-30 exit |

###### Construction and lifetime decisions

- Keep `StartRuntime(t)` as a local parent-test scope. It acquires no managed
  service; its Reference Pack root remains tied to that parent and keyed by
  database identity, not a shorter-lived replacement server.
- Change package `StartServer(t, ServerOptions)` to delegate to
  `Runtime.StartServer(t, ServerOptions)`. Keep `PrepareIsolatedDatabase`; remove
  `PrepareServerDatabase`, `StartDefaultServer` and every `StartServerWith...`
  variant. Keep `ServerOptions` and `ServerHarness` capabilities. Callers supply
  an explicit `TestRouteMode` justified by the scenario under REQ-480; this plan
  adds no implicit route enablement or new public injection API.
- Before either managed service is touched, validate options and the complete
  sorted service set through `suiteservices.CheckServiceDependencies`. Use
  `object_store,postgres` when appsupport creates a store, and `postgres` when a
  store is supplied. A supplied database avoids database allocation and pgtest
  attachment, but opening its runtime handles still requires PostgreSQL admission.
  `PrepareIsolatedDatabase` explicitly guards/acquires PostgreSQL only.
- Use private per-call starter values at the acquisition boundary and retain an
  error-returning core for deterministic tests. Remove the two mutable starter
  globals and restore helper. Do not duplicate the shared declaration parser or
  create new public injection/configuration types.
- When Database is absent, prepare one isolated database through pgtest. When
  ObjectStore is absent, attach through s3test and use `BootstrapBucketT` for an
  independent test-owned bucket and appsupport-owned client. Keep the shared physical
  generation. Register cleanup under the test owning that allocation.
- Supplied stores and databases are borrowed. Their allocating caller owns store
  closure and database deletion; appsupport owns only the new handles it opens.
  Do not auto-close, drop or replace supplied fixtures, including on partial
  startup failure. Stop application work, close owned handles/clients, then delete
  owned fixture resources; cleanup uncertainty must remain visible. Early server
  close for a restart does not retire parent-owned database, store or pack bytes.
- Remove Runtime's exposed Postgres/S3 fields. Migrate database callers to
  `PrepareIsolatedDatabase`; explicit Recovery operations call existing service
  helpers directly. Do not replace the removed fields with forwarding accessors.
- Incident Bundle target setup and Reference Pack simultaneous-worker setup use
  the canonical constructor with independent defaults. Retain domain-specific
  trust, environment and fault configuration. Preserve the same database and
  explicit store in restart/adapter-replacement tests, closing the prior server
  before replacing it on that database.
- Remove the Auth/Timeline constructor-only wrappers and aliases. Migrate Auth
  callers with their meaningful owner route contributions and admission settings;
  keep Timeline ordinary product scenarios disabled. Domain login/scenario
  helpers remain. Select disabled mode for ordinary product-route tests,
  harness-owned mode for guarded route tests or owner contributions, and custom
  environment only for the relevant negative cases. The inspected Auth helper
  registers `auth.RegisterTestRoutes()`; retain that intentional contribution,
  rather than mechanically removing it with the forwarding wrapper.

###### Resource acquisition and ownership matrix

This matrix defines appsupport's required-service set and allocations. A service
declaration is not an instruction to allocate another fixture. `StartRuntime`
itself allocates only local parent-test state and acquires no managed service.

| Resources supplied to StartServer | Complete sorted helper-required services | Allocations owned by appsupport's calling test | Borrowed resources |
| --- | --- | --- | --- |
| Neither | `object_store,postgres` | Isolated database, fresh bucket, object client and runtime database handles | Physical service generations |
| Database only | `object_store,postgres` | Fresh bucket, object client and runtime database handles | Supplied database and physical service generations |
| Object store only | `postgres` | Isolated database and runtime database handles | Supplied object store and PostgreSQL generation |
| Database and object store | `postgres` | Runtime database handles only | Both supplied fixtures and PostgreSQL generation |

The catalog row declares the union required by its entire scenario, including
caller-owned setup. A supplied S3-backed store can therefore require the row to
declare `object_store` even though appsupport does not acquire it. Do not infer a
row's complete dependencies from `ServerOptions` alone. The inspected filesystem
Evidence cancellation and Imports recovery cases need no S3 acquisition at either
boundary. Retain their dedicated database capability and exclusive cluster scope.

Use existing cleanup registration and service ownership. No new fixture registry,
public resource-policy type, lock workaround or replacement cleanup coordinator is
needed. Supplied resources must outlive every server that borrows them; callers
keep restart resources in the parent scope and close the prior server before
starting its replacement.

###### Assertion successor and retirement decisions

| Existing arrangement / guarantee | Decision / successor assurance |
| --- | --- |
| Two StartRuntime guard tests with patched globals | Replace with one coherent acquisition/admission group using independent local starters: all four resource combinations, complete sorted required sets, no acquisition after rejection, no reacquisition of supplied fixtures and correct failure envelope |
| Shared malformed-declaration matrix | Keep suiteservices tests; appsupport proves the integration boundary and early rejection, not another full parser matrix |
| Every runtime must acquire both services | Retire as incidental; a supplied filesystem store succeeds with PostgreSQL-only assignment and no object-store startup |
| Package bucket serialization and custom second-server scaffolding | Retire from default application construction; prove independent live servers, sibling data/bytes and cleanup isolation |
| Constructor names, aliases and counts | Retire; ordinary compilation and behavioral successors establish migration correctness; add no exact-symbol/source-layout replacement test |
| Explicit route modes, injected dependencies, observation capabilities and fault admission | Preserve owner-required admission and meaningful route contributions through migrated callers and existing HTTP/security/commit-fault witnesses; historical wrapper defaults are not separate requirements |
| Owned versus borrowed fixture lifetime | Preserve supplied stores/databases after close or failed startup; close newly opened handles and retire only owned fixtures; use existing assembly failure witnesses plus missing composition-boundary coverage |
| Imports crash recovery and durable Reference Pack bytes | Preserve explicit database/store/parent-root reuse and exactly-once durable outcomes across restart |
| Evidence read cancellation and mapped object-store failures | Preserve drained read before terminal cancellation, effect-free cancellation and typed/redacted public mapping |
| Real provider, Recovery backing storage and explicit package reuse | Keep their distinct observation boundaries; filesystem fault fixtures do not replace real-service evidence |

WF-26 expanded this table into the E-073 affected assertion/caller dispositions.
A consolidation names its surviving witness; retirement of incidental behavior
requires a reason, not an artificial replacement assertion. Do not silently drop
negative or cleanup coverage while reducing setup or row duplication.

The ledger records caller/setup, owner-required guarantee, observation boundary,
existing assertion/subcase, resource ownership, explicit route contribution/mode,
exact current row, disposition, successor witness/row and qualification evidence.
Inspect the actual constructor consumer closure, including promoted methods and
type aliases; importing appsupport alone does not make all tests in a file part
of the migration. Enumeration and source inspection do not establish a pass.

Consolidate assertions only when both their guarantee and observation boundary
coincide. Shared setup can be removed while retaining independently meaningful
process, provider, restart, security and cleanup assertions. Keep domain failures
and negative cases explicit; table-driven tests must not obscure which production
effect or failure boundary is being checked. Add characterization only for an
identified assurance gap. Do not create source-shape/symbol-count guards or copy
provider/parser matrices merely to prove that obsolete helpers disappeared.

##### 8. Validation Plan

V-034–V-038 are complete at E-074–E-078.
Public help/task-guide refresh and the exact caller selection were completed in
WF-26. Commands run from `/home/jochi/code/cartulary` through Make, with pinned
prerequisites, explicit fresh run IDs and native exit/result-root capture.

| Validation | Stage / slice | Commands and required coverage | Gate / failure consequence |
| --- | --- | --- | --- |
| V-034 | Pre-move characterization; S-026 | `make help ALL=1`; `make task-guide OWNER=<affected-owner> DETAIL=rows`; exact `make test-slice OWNER=<owner> ROWS=<ledger-row-union>` using initial selections below plus all affected callers | Complete selection ledger and classified current baseline before movement; diagnose failures rather than broaden/retry by habit |
| V-035 | Post-structural slice; S-027 | Migrated constructor consumers and route/config/restart witnesses through exact test-slice selections; `make lint CHECK=go` and `make lint CHECK=backend-boundary` | No live retired API/type uses, no composition-boundary drift, unchanged production assertions |
| V-036 | Post-acquisition slice; S-028 | New fixture admission group covering all four resource combinations; no managed acquisition in StartRuntime; no acquisition after rejection; two independent application servers with sibling data/bytes and cleanup; borrowed store/database lifetime and startup failure; parent/child restart storage; named production witnesses below | No unnecessary allocation/attachment, no S3 acquisition for filesystem-only setup, no package lease deadlock, no premature resource deletion or lost durable effects; real composition evidence remains integration evidence |
| V-037 | Post-routing slice; S-029 | `make generate`; `make source-check CHECK=generated`, `CHECK=schema`, `CHECK=catalog`, `CHECK=identity`; exact new/changed row selections | Authored/generated selectors, service requirements, evidence class and fixture capabilities agree; no aliases/duplicate execution |
| V-038 | Final regression; S-030 | `make agent-finalize`, inspect changes, freeze source; qualify affected caller/guarantee union and missing static obligations; `make harness-observability-check RESULTS_DIR=<exact-compatible-root> CHECK=all` | Native exits, valid evidence and cleanup pass; repair changes require renewed compatibility assessment; no full-owner/release claim from a subset |
| V-039 | Historical initial planning save | Installed-only external-output `make lint-markdown`; direct structure/links/IDs/graph review; diff and baseline comparison | E-067/E-068 belong to the earlier save and are not current validation |
| V-040 | Current plan-refinement documentation check | Installed-only external-output `make lint-markdown`; direct structure/links/IDs/graph and plan review; `git --no-optional-locks diff --check -- docs/handoffs/testing-remediation-01.md`; current content/index/history comparison | E-071/E-072; configured lint does not include this handoff, so direct document review is required; no production qualification |

The source policy, services and expected artifacts remain those of the selected
rows. Do not invent a fallback, extend deadlines, refresh goldens, change resource
claims or make unavailable infrastructure a passing skip. A scoped portfolio is
not full-owner acceptance. `RESULTS_DIR` is supplied to future finalization only
for a compatible successful full warm-check; otherwise record retained maintenance
skipped because it is unset. Do not run an extra broad check merely to obtain one.

###### Exact initial witnesses

These current rows anchor WF-26 selection; they are not a complete migrated-caller
portfolio or permanent names. Preserve existing semantic identities when their
meaning is unchanged. Record exact successors before removing selectors.

| Owner | Current row ID | Required observation |
| --- | --- | --- |
| app.server | `app.server.support_unit.incident_create_commit_fault_requires_validated_efdc344831` | Initial admission/commit-fault assertions; split fixture acquisition into its own semantic unit row while keeping the incident guarantee |
| app.server | `app.server.integration.http_harness_restart_storage` | Lower HTTP composition and explicit route behavior |
| module.evidence | `module.evidence.integration.background_verification_cancellation` | Stored read drains before cancellation publication; no leaked verification/association effects |
| module.imports | `module.imports.integration.unit_commit_crash_recovery_108cd054a6` | Restart before/after unit commit preserves exactly one owner effect and applied outcome |
| module.reference_data | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` | Includes TestWorkerDatabaseBarriersAreInstanceScoped_Integration; independent worker progresses while sibling is blocked |
| module.recovery | `module.recovery.integration.real_postgres_backed_runtime_metadata_persistenc_d652f0998b` | Real backing-storage metadata and restored object evidence |

For example, the two filesystem-backed scenarios start with these exact commands:

```sh
make test-slice OWNER=module.evidence ROWS=module.evidence.integration.background_verification_cancellation
make test-slice OWNER=module.imports ROWS=module.imports.integration.unit_commit_crash_recovery_108cd054a6
```

After full selected-case inspection, remove `object_store` from those two authored
rows while retaining dedicated/exclusive PostgreSQL capability. Do not infer that
all rows supplying a store can lose S3: the store may wrap real S3, another case
may need it, or a helper may perform explicit object operations. Map Incident
Bundle round trips, Auth/Timeline wrapper consumers, Reference Pack restarts and
other changed callers before movement. Group compatible coverage once; retain
independently meaningful process/provider boundaries when affected, without
repeating an entire historical portfolio by suite name.

New real application isolation/cleanup evidence belongs in an integration row;
local admission tests remain service-free unit evidence. Keep incident commit-fault
admission separate. Update authored family/verification inputs as required,
regenerate projections through Make and keep Markdown outside every executable
selection, generator and evidence input.

| Invocation group | Side effects / current admission | Result / evidence |
| --- | --- | --- |
| Source/Git inspection and preservation material | Read repository; comparison files outside it; optional Git writes disabled | Completed, E-069/E-070; not a canonical product check |
| V-039 | Earlier document save | Historical E-067/E-068 only; not rerun under the old identity |
| V-040 | Installed prerequisites only, fixing disabled, existing workspace lock, external result/static-analysis cache | Passed, E-071; current sole-file/content/index/history comparison passed, E-072 |
| V-034–V-038 | Authorized tests, services, source changes, generation and finalization | V-034–V-038 passed; final compatible portfolio, audits and handoff at E-078 |
| Full CI, visual/UI-review/goldens, benchmarks, releases, preceding WF-33 | Unrelated qualification or unrequested execution | Not run; excluded from this bounded iteration |

##### 9. Top-Level Work Tracker

| Work item | Workstream | Status | Dependencies / evidence | Binary exit |
| --- | --- | --- | --- | --- |
| T-027 — save application-fixture plan | Historical document save | DONE | E-065–E-068; V-039 | Earlier reviewed sole-file save; no current source qualification |
| T-028 — characterize affected construction and guarantees | WF-26 | DONE | S-026/V-034; later implementation authorization | Complete affected ledger/current baseline, no incidental behavior frozen |
| T-029 — consolidate constructors and wrappers | WF-27 | DONE | T-028; S-027/V-035; RB-011 | One options path, migrated callers and qualified structural exit |
| T-030 — simplify acquisition and namespace ownership | WF-28 | DONE | T-029; S-028/V-036; RB-012 | Conditional services and independent fixtures qualified |
| T-031 — consolidate assertions and routing | WF-29 | DONE | T-030; S-029/V-037; RB-012 | Coherent rows and regenerated truthful declarations |
| T-032 — qualify and hand off | WF-30 | DONE | T-031; S-030/V-038; RB-012 | Frozen compatible affected portfolio and explicit acceptance limits |
| T-033 — refine application-fixture plan | Documentation only | DONE | E-069–E-072; V-040; agreed refinement plan | Resource matrix, ownership, route-mode and assertion decisions saved; current documentation review and preservation evidence recorded |

T-027/T-033 retain their historical documentation-only scope. The user subsequently
authorized T-028–T-032; their ordered checkpoints and final qualification are
recorded below. Earlier passes do not qualify later source revisions.

##### 11. Open Questions and Blockers

| ID | Question / affected scope | Needed evidence | Status | Resolution |
| --- | --- | --- | --- | --- |
| RB-011 | Current pre-move characterization; S-027 | E-073/E-074 caller ledger and V-034 baseline | DONE | WF-26 exit saved before constructor edits |
| RB-012 | Ordered successor qualification; S-028–S-030 | WF-27, WF-28 and WF-29 exits | DONE | E-075/E-076/E-077 saved before each dependent entry |

No unresolved bounded-task design or implementation blocker remains. Final
qualification and handoff are complete at E-078. Historical RB-001–RB-010 and the
separate WF-33 retain their previous dispositions.

##### 12. Binary Completion Criteria

| Outcome | Final statement |
| --- | --- |
| Planning completeness | Complete for F-038–F-043; remediation areas, rationale, benefit, migration, risks, validation and sequencing recorded |
| Workstream completion | WF-26–WF-30 DONE; each predecessor exit saved before the next entry; validation and handoff completed last |
| Authorization | User authorized implementation of the proposed plan; all writes and checks stayed within its bounded scope |
| Specification disposition | No normative amendment required; adopted harness requirements already govern the selected behavior; projections remain subject to human fidelity review |
| Implementation and migration | One options path, conditional acquisition, private starters, independent defaults, explicit borrowed lifetimes and migrated callers; redundant wrappers and compatibility burden removed |
| Final validation | 198 exact rows across 24 owners, Go lint, backend boundaries, source-check/all and 27 retained-validity checks passed on one compatible source/toolchain |
| Documentation and preservation | Configured Markdown lint plus direct handoff review; historical log preserved, source freeze unchanged, HEAD/staged entries unchanged; raw-index stat-cache refresh disclosed |
| Handoff scope | Bounded effort complete; no full-owner, full-CI, visual, performance or release claim; maintenance skip and separate WF-33 explicitly retained |

- [x] Every affected caller is mapped to an exact current row or an intentional retirement.
- [x] Complete dependency admission precedes conditional acquisition in all four cases.
- [x] Independent applications preserve sibling state and retire only owned fixtures.
- [x] Borrowed resources survive close/failure; owned handles close and parent restart bytes persist.
- [x] Auth/Timeline contributions, explicit route modes and incident-fault admission remain qualified.
- [x] Cancellation, crash recovery, real backing storage and other domain assertions remain qualified.
- [x] Authored/generated routing agrees; no duplicate, dangling or compatibility rows remain.
- [x] Compatible final evidence, cleanup, secret scans, failures/repairs, exclusions and handoff are recorded.

</details>


### WF-30 exit — E-078 / V-038 / T-032 DONE

Final review documented ServerOptions ownership and restored standard-library/project
import grouping in already changed files. `make agent-finalize` passed at
`tr01-wf30-finalize-20261027-01`, native exit 0, before the broader portfolio.
It does not replace `make format`. No compatible successful full warm-check was
supplied; retained-evidence/scheduler maintenance was skipped because `RESULTS_DIR`
was unset. The selected portfolio is 198 exact rows: the 192 E-073 caller rows plus
fixture admission, both lifecycle rows, incident fault admission, HTTP restart and
invalid-deployment/assembly-failure witnesses, across 24 owners.

First frozen attempt: `tr01-wf30-source-20261027-01` passed source-check/all.
`tr01-wf30-go-20261027-01` exited 2 with product/test_assertion_failure: the
import regrouping needed canonical formatting. `tr01-wf30-owner-00-20261027-01`
also exited 2 with product/test_assertion_failure: a one-shot server-statistics
query was an invalid synchronization boundary for client-pool closure. The corrected
witness will capture the actual owned pool via the existing runtime configuration
hook, reject route assembly with an identified error, prove the server leaves its
borrowed pool open, then prove fixture cleanup closes it. This keeps the real
startup/cleanup witness without adding sleeps, polling or timeout allowances.
The 16 active domain selections finished successfully on that unchanged source;
those roots remain preliminary, not the final compatible acceptance portfolio.
Applied the direct-pool correction after they settled. `make format` passed at
`tr01-wf30-format-20261027-01`. The exact borrowed-fixture row passed at
`tr01-wf30-cleanup-20261027-01`, and Go lint passed at
`tr01-wf30-go-20261027-02`, native exits 0. The second finalizer passed at `tr01-wf30-finalize-20261027-02`, native exit 0,
with zero generated-file changes. The source was frozen again; the successful
Go lint result has the identical executable-source/toolchain identity and is
retained in the final compatible portfolio.


Final executable source: `sha256:4bef6ffbe421cf8ccce5009e0412ada22a9da4bbc77615d7bec8d79446a5be30`.
Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
Every command below used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run ID>` and
returned native exit 0. Roots are `.cartulary/test-results/<run ID>`.
Owner commands are `make test-slice OWNER=<owner> ROWS=<exact union>`: E-073's
192 rows plus the six named app.server witnesses above. The exact final union is
retained in `final-portfolio.json` beside the implementation baseline and in each
run manifest's declared inputs; no executable selection reads this Markdown.
The three final batch sessions all exited 0; all constituent Make invocations
completed successfully. Static commands are `make source-check CHECK=all`,
`make lint CHECK=go` and `make lint CHECK=backend-boundary`.

| Owner / static obligation | Exact run ID | Selected rows | Duration ms |
| --- | --- | --- | --- |
| app.server | `tr01-wf30-owner-00-20261027-02` | 7 | 16944 |
| module.assessments | `tr01-wf30-owner-01-20261027-02` | 2 | 17263 |
| module.auth | `tr01-wf30-owner-02-20261027-02` | 13 | 64236 |
| module.collaboration | `tr01-wf30-owner-03-20261027-02` | 10 | 32869 |
| module.entities | `tr01-wf30-owner-04-20261027-02` | 8 | 92245 |
| module.evidence | `tr01-wf30-owner-05-20261027-02` | 17 | 48929 |
| module.imports | `tr01-wf30-owner-06-20261027-02` | 16 | 42594 |
| module.incidentbundles | `tr01-wf30-owner-07-20261027-02` | 11 | 80122 |
| module.incidents | `tr01-wf30-owner-08-20261027-02` | 15 | 94924 |
| module.indicators | `tr01-wf30-owner-09-20261027-02` | 2 | 17076 |
| module.jobapi | `tr01-wf30-owner-10-20261027-02` | 4 | 16816 |
| module.links | `tr01-wf30-owner-11-20261027-02` | 1 | 12596 |
| module.networkflow | `tr01-wf30-owner-12-20261027-02` | 9 | 89951 |
| module.parties | `tr01-wf30-owner-13-20261027-02` | 2 | 19542 |
| module.recovery | `tr01-wf30-owner-14-20261027-02` | 2 | 18910 |
| module.reference_data | `tr01-wf30-owner-15-20261027-02` | 10 | 68537 |
| module.reportcomposition | `tr01-wf30-owner-16-20261027-02` | 2 | 32148 |
| module.reporting | `tr01-wf30-owner-17-20261027-02` | 5 | 41850 |
| module.revisions | `tr01-wf30-owner-18-20261027-02` | 18 | 49103 |
| module.savedviews | `tr01-wf30-owner-19-20261027-02` | 4 | 24106 |
| module.tasksdecisions | `tr01-wf30-owner-20-20261027-02` | 1 | 17712 |
| module.timeline | `tr01-wf30-owner-21-20261027-02` | 11 | 49318 |
| module.workbook | `tr01-wf30-owner-22-20261027-02` | 27 | 74557 |
| platform.viewschema | `tr01-wf30-owner-23-20261027-02` | 1 | 24154 |
| source-check / all | `tr01-wf30-source-20261027-02` | static | 15527 |
| lint / go | `tr01-wf30-go-20261027-02` | static | 7630 |
| lint / backend-boundary | `tr01-wf30-boundary-20261027-02` | static | 2704 |

All **198 required rows passed**, with exact selected/observed row agreement and
no missing, extra, failed, skipped or cancelled rows. All 27 acceptance roots share
the source/toolchain identities above; no execution-cache hits were used. Earlier
failures, characterization and workstream roots remain retained, but are excluded
from this acceptance union. Functional durations make no performance claim.

For each listed root, `make harness-observability-check RESULTS_DIR=<exact-root>
CHECK=all` passed with native exit 0 and `full_validity=1`. Every cleanup outcome
was `completed`, with no cleanup failures; every retained-secret scan passed.
Optional diagnostics/resource accounting were not selected and are not claimed.
`final-result-review.json` and `final-evidence-roots.json` retain the explicit
selection/source/cleanup review alongside the external baseline. Canonical retained
artifacts remain the evidence authority; these review notes are not a new runner.

Documentation: `make lint-markdown` passed at `tr01-wf30-markdown-20261027-01`,
native exit 0; summary `adhoc/lint-markdown/tool-run-summary.json`. Its configured
scope excludes this handoff. Supplemental direct review therefore covers the final
handoff's structure, local links, table shape, dependency/checkpoint sequence,
completion statements and historical preservation. `git diff --check` is clean.

Final boundary: 93 existing files changed, one integration-test file added and the
two constructor-only wrapper files removed, relative to E-073. Changes are confined
to application fixture composition, its caller setup/imports, fixture tests, one
verification owner, three authored families, the generated topology index and this
handoff. Provider internals, production implementation and normative specifications
were not changed. The 7,342-file non-Markdown freeze comparison is identical after
qualification. E-069–E-072 and all preceding historical log content remain verbatim;
pre-existing handoff work is preserved. HEAD and staged entries are unchanged;
raw index bytes changed through stat-cache refresh, with no staged content change.
Ignored output/cache trees were not exhaustively fingerprinted. Boundary and
preservation notes are in `boundary-review.json` beside the E-073 baseline.

Migration/handoff: use `StartServer(t, ServerOptions)` or a parent Runtime's same
method, with an explicit route mode. Reuse database and store explicitly for a
restart and keep the Runtime in their parent scope for Reference Pack bytes.
Use provider APIs directly when the scenario needs provider-specific operations.
No retired aliases, global substitutions, second-store workarounds or forwarding
service accessors remain. F-038–F-043 are resolved; the per-gap rationale, migration
impact, risks and witnesses are recorded in Section 5. No remaining bounded-task
implementation work or design blocker is deferred to the next owner.

Full CI/full-owner acceptance, browser/visual/UI-review, dedicated race suites,
benchmarks and release qualification were not run: this scope changes Go test
composition and routing, with no production/UI behavior change. Finalizer canonical
retained-run and scheduler-drift maintenance were skipped because `RESULTS_DIR`
was unset; no compatible successful full warm-check was supplied. The per-root
validity checks above are complete. The separate WF-33 remains outside this task.
Normative completeness and human projection review are not inferred from test passes.

### WF-29 exit — E-077 / V-037 / T-031 DONE

Saved before WF-30 finalization. The old incident commit-fault row now selects only
its incident-admission witness. Fixture admission has semantic row
`app.server.support_unit.application_fixture_admission_validates_complete_2e28540399`,
derived through `make author-test-row-id`; its selector is
`TestApplicationFixtureAdmission`, with no service or runtime requirement. The two
real fixture lifecycle rows remain integration evidence. All three now project
`app.server.verification.application_fixtures`, behavior `harness`, profile
`support`, evidence `go_test`; they do not claim product conformance. No alias,
duplicate execution, copied declaration parser or provider-matrix replacement was
introduced. Evidence cancellation and Imports crash recovery retain dedicated,
exclusive PostgreSQL but no longer declare object storage. Their scenario-owned
stores are filesystem adapters; other real-service witnesses remain intact.

Authored verification/family inputs were updated first. `make generate` passed at
`tr01-wf29-generate-20261027-01`; generated topology index changes are reproducible.
`make source-check CHECK=generated|schema|catalog|identity` ran as four distinct
commands, each passed with native exit 0 at
`tr01-wf29-source-<check>-20261027-01`. Exact fixture admission, incident admission
and both lifecycle rows passed at `tr01-wf29-fixtures-20261027-01`. The corrected
Evidence and Imports rows passed at `tr01-wf29-evidence-20261027-01` and
`tr01-wf29-imports-20261027-01`; all native exits 0. Commands used `CGO_ENABLED=0`,
with roots under `.cartulary/test-results/`. No failure or skipped row occurred.

WF-30 now finalizes and freezes source, requalifies all E-073 affected rows plus
new fixture and independent assembly witnesses, validates retained evidence,
and records final boundary, migration and handoff limits.

### WF-28 exit — E-076 / V-036 / T-030 DONE

Saved before WF-29 routing edits. Runtime now owns only parent-scoped deployment
bytes. Each application validates its complete service declaration before private,
per-call acquisition, borrows supplied fixtures, and allocates a fresh database
and bucket for missing fixtures. Supplied stores remove appsupport's S3 acquisition;
supplied databases retain PostgreSQL admission for application connections.
Owned clients/handles report cleanup failures; borrowed stores/databases survive
server close, child cleanup and failed startup. Removed direct runtime service
fields; Recovery uses the existing explicit provider, while Incident Bundles and
Reference Pack worker tests use canonical independent application construction.
No provider, broker, production or normative-specification changes were needed.

The admission matrix covers each missing required service, rejection before any
factory call, local Runtime construction, invalid options and application-env
noninterference. Two new integration witnesses prove live database/object isolation,
sibling cleanup, borrowed-resource survival, failed-startup connection cleanup,
closed child handles and retained Reference Pack bytes across replacement servers.
Minimal authored row additions made these witnesses selectable; the semantic unit
split and final dependency/profile corrections remain WF-29 work.

All commands used `CGO_ENABLED=0`; roots are `.cartulary/test-results/<run ID>`.
Initial format/generate `tr01-wf28-format-20261027-01` and
`tr01-wf28-generate-20261027-01` exited 2: unsorted selector names were rejected
before execution. Sorted the authored list; generation `-02` and format `-02`
passed. Go lint passed at `tr01-wf28-go-20261027-01`.

Fixture selection: initial incident/admission row, both new integration rows,
HTTP restart and invalid-deployment/preflight rows. `tr01-wf28-fixtures-20261027-01`
exited 2 (product/test_assertion_failure): the test's store wrapper hid typed
capabilities. Embedding the concrete filesystem store and stopping after child
failure fixed it; `-02` passed. After strengthening negative-startup cleanup,
`-03` exited 2 because its observer confused the effective database role with the
login identity. The observer now counts other connections to the isolated database
relative to its pre-start baseline. `tr01-wf28-fixtures-20261027-04` passed all
five rows, native exit 0; earlier failures remain retained. Format `-03` passed.

Focused domain selections all passed, native exits 0: exact E-073 Incident Bundles
and Reference Data unions, Recovery union, Evidence cancellation and Imports crash
recovery, respectively `tr01-wf28-domain-0-20261027-01` through
`tr01-wf28-domain-4-20261027-01`. Earlier-source passes are workstream evidence,
not a compatible final acceptance union. WF-30 requalifies the complete affected
portfolio after finalization and source freeze.

### WF-27 exit — E-075 / V-035 / T-029 DONE

Saved before WF-28 acquisition edits. Migrated 86 files, removed constructor-only
Auth/Timeline wrappers and redundant appsupport constructors/aliases, retained the
options shape, explicit modes, Auth route contributions and domain helpers. Caller
assertions were retained; source changes are constructor arguments, imports and
helper types. No acquisition behavior changed in this slice. Searches found no
retired constructor/type consumers.

`CGO_ENABLED=0` was used for every following Make command; roots are
`.cartulary/test-results/<run ID>`. Format passed at `tr01-wf27-format-20261027-01`.
Focused commands are the E-073 exact owner row unions for Auth, Timeline and
Evidence; the Imports crash-recovery row; and the initial app.server admission/HTTP
rows. Backend boundary, app.server, Auth, Evidence and Imports passed at
`tr01-wf27-check-1-20261027-01`, `-2-`, `-3-`, `-5-` and `-6-` respectively.
Initial Go lint (`-0-`) and Timeline (`-4-`) failed because Timeline support retained
one unused scenariotest import. Lint classified the compile finding as
product/test_assertion_failure; Timeline reported infra/preflight_error with no
selector execution. Both native commands exited 2 and their roots are retained.
Removed the unused import. `make lint CHECK=go` and the exact Timeline selection
then passed at `tr01-wf27-check-0-20261027-02` and
`tr01-wf27-check-4-20261027-02`, native exits 0. Earlier focused successes describe
their earlier source; they are not claimed as a compatible final portfolio.
Diagnostic readers require public `TARGET=lint` or `TARGET=test-slice`; attempted
internal-target/log queries were rejected without product changes.

Next: WF-28 implements conditional acquisition, private per-call fixture starters,
fresh namespaces, explicit provider access and ownership witnesses. Final frozen
qualification will cover the complete affected union again.

### WF-26 exit — E-074 / V-034 / T-028 DONE

Saved before WF-27 source edits. Current-source characterization passed for all
192 affected rows across the 24 exact owner selections below. Each command was
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run ID> make test-slice OWNER=<owner>
ROWS=<exact row union from E-073>`. All commands returned native exit 0; the native
portfolio session returned 0. Roots are `.cartulary/test-results/<run ID>`.
The additional initial admission/HTTP selection passed at
`tr01-wf26-admission-20261026-01` (native exit 0). `make help ALL=1` and all 24
`make task-guide OWNER=<owner> DETAIL=rows JSON=1` discovery commands passed.
Source digest: `sha256:d7b53f17a7e34c2ec60fbdb5f8e64e484deefa482346c352c4cb6ddafa0b8be7`.
All retained-secret scans passed. No assertion failures or missing rows were found.
The baseline comparison confirms only the tracker changed among the 7,654 initial
entries. No product/source edits occurred before this exit; ignored build/cache
outputs were not fingerprinted. No full-owner, benchmark or release claim follows.

| Owner | Affected rows | Run ID | Result |
| --- | --- | --- | --- |
| app.server | 1 | `tr01-wf26-owner-00-20261026-01` | pass; native exit 0 |
| module.assessments | 2 | `tr01-wf26-owner-01-20261026-01` | pass; native exit 0 |
| module.auth | 13 | `tr01-wf26-owner-02-20261026-01` | pass; native exit 0 |
| module.collaboration | 10 | `tr01-wf26-owner-03-20261026-01` | pass; native exit 0 |
| module.entities | 8 | `tr01-wf26-owner-04-20261026-01` | pass; native exit 0 |
| module.evidence | 17 | `tr01-wf26-owner-05-20261026-01` | pass; native exit 0 |
| module.imports | 16 | `tr01-wf26-owner-06-20261026-01` | pass; native exit 0 |
| module.incidentbundles | 11 | `tr01-wf26-owner-07-20261026-01` | pass; native exit 0 |
| module.incidents | 15 | `tr01-wf26-owner-08-20261026-01` | pass; native exit 0 |
| module.indicators | 2 | `tr01-wf26-owner-09-20261026-01` | pass; native exit 0 |
| module.jobapi | 4 | `tr01-wf26-owner-10-20261026-01` | pass; native exit 0 |
| module.links | 1 | `tr01-wf26-owner-11-20261026-01` | pass; native exit 0 |
| module.networkflow | 9 | `tr01-wf26-owner-12-20261026-01` | pass; native exit 0 |
| module.parties | 2 | `tr01-wf26-owner-13-20261026-01` | pass; native exit 0 |
| module.recovery | 2 | `tr01-wf26-owner-14-20261026-01` | pass; native exit 0 |
| module.reference_data | 10 | `tr01-wf26-owner-15-20261026-01` | pass; native exit 0 |
| module.reportcomposition | 2 | `tr01-wf26-owner-16-20261026-01` | pass; native exit 0 |
| module.reporting | 5 | `tr01-wf26-owner-17-20261026-01` | pass; native exit 0 |
| module.revisions | 18 | `tr01-wf26-owner-18-20261026-01` | pass; native exit 0 |
| module.savedviews | 4 | `tr01-wf26-owner-19-20261026-01` | pass; native exit 0 |
| module.tasksdecisions | 1 | `tr01-wf26-owner-20-20261026-01` | pass; native exit 0 |
| module.timeline | 11 | `tr01-wf26-owner-21-20261026-01` | pass; native exit 0 |
| module.workbook | 27 | `tr01-wf26-owner-22-20261026-01` | pass; native exit 0 |
| platform.viewschema | 1 | `tr01-wf26-owner-23-20261026-01` | pass; native exit 0 |

The implementation retains domain assertion bodies and owner contributions.
Retired constructor names, unconditional service acquisition and package-bucket
serialization are explicitly not requirements. Existing production owned-startup
failure checks remain distinct from the new appsupport composition witnesses.
Next: WF-27 constructor/caller migration; F-039–F-041 acquisition changes wait
until its saved exit. RB-011 is resolved; later predecessor gates remain open.

### Implementation entry — E-073

Current source: `e6ed547531945b62d3006ae8c1d458746ddf4670`; only this handoff was
dirty on entry. Baseline and original handoff are retained in
`/tmp/cartulary-appsupport-implementation-den6agq8`. The user instruction authorizes
all five bounded workstreams. Earlier document-only restrictions describe their
historical sessions. No owner amendment is needed: REQ-410/804/480 and §11.0 already
support the target behavior. Provider internals and production assembly stay owned
by their existing boundaries. No frontend changes are planned.

The caller closure includes 321 functions (286 test functions), mapped to 192 rows
across 24 owners. The inventory traces local helper calls and constructor-wrapper
types; all discovered affected test functions have exact catalog rows. The
following ledger preserves domain assertions and records construction changes;
non-test helpers inherit qualification from their listed package consumers. Route
modes remain explicit: Auth's contributed routes retain harness-owned admission;
Timeline and ordinary product scenarios use disabled mode. Existing explicit modes,
custom environment negatives, dependency adapters, durable storage and domain fault
configuration are retained. F-039–F-041 separately change acquisition/ownership.

| Caller / source | Setup / disposition | Exact row(s) |
| --- | --- | --- |
| `internal/modules/assessments/assessments_integration_test.go:149` — `TestAssessmentsAppendOnlyFiltersAndValidation` | harness := appsupport.StartServer(t, "assessments-filters") | `module.assessments.store.http_filter_validation_characterization` |
| `internal/modules/assessments/assessments_integration_test.go:21` — `TestAssessmentsCreateAndProjection` | harness := appsupport.StartServer(t, "assessments-create") | `module.assessments.store.http_create_replay_projection_publication_characterization` |
| `internal/modules/auth/integration_test.go:597` — `TestAdminCredentialActions_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c` |
| `internal/modules/auth/integration_test.go:997` — `TestAdminCredentialAuditAndScope_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c` |
| `internal/modules/auth/integration_test.go:1317` — `TestAdminPasswordResetReplayReturnsOriginalCommittedResource_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c` |
| `internal/modules/auth/integration_test.go:1409` — `TestAdminTOTPResetAndRevokeAllReplay_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c` |
| `internal/modules/auth/support_integration_test.go:318` — `TestAuditAttribution` | runtime := flowtest.StartRuntime(t) | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| `internal/modules/auth/support_integration_test.go:472` — `TestAuthorizationReDerivation` | runtime := flowtest.StartRuntime(t) | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| `internal/modules/auth/support_integration_test.go:197` — `TestBootstrapBoundaries` | runtime := flowtest.StartRuntime(t) | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| `internal/modules/auth/integration_test.go:877` — `TestBootstrapEnrollmentConsumption_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb` |
| `internal/modules/auth/integration_test.go:409` — `TestBootstrapTokenRouteBoundaries_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.bootstrap_tokens_work_only_on_totp_begin_and_tot_36751ad438` |
| `internal/modules/auth/support_integration_test.go:227` — `TestCSRFProtection` | runtime := flowtest.StartRuntime(t) | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| `internal/modules/auth/integration_test.go:257` — `TestCredentialStateAndBootstrapFlows_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb` |
| `internal/modules/auth/integration_test.go:714` — `TestCredentialStateTransitions_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb` |
| `internal/modules/auth/admin_audit_integration_test.go:17` — `TestDeploymentAdministrativeAuditRoute_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.deployment_audit_route_is_scope_safe_and_keyset_4f8db0c261` |
| `internal/modules/auth/enterprise_auth_integration_test.go:468` — `TestEnterpriseAuthBindingLifecycle_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.deployment_admin_binding_create_replay_divergent_85cb20830a` |
| `internal/modules/auth/enterprise_auth_integration_test.go:20` — `TestEnterpriseAuthProviderOIDC_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.enterprise_provider_discovery_proves_sorted_exac_059973abc3` |
| `internal/modules/auth/secret_envelopes_integration_test.go:68` — `TestEnterprisePKCEEnvelopeSubstitution_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.secret_envelope_substitution` |
| `internal/modules/auth/enterprise_auth_integration_test.go:264` — `TestEnterpriseSAMLACS_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.saml_sp_initiated_begin_acs_verifies_determinist_3b5547023b` |
| `internal/modules/auth/integration_test.go:32` — `TestLoginSessionLifecycle_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.real_login_creates_one_durable_session_row_sessi_fffac0a9af` |
| `internal/modules/auth/integration_test.go:1237` — `TestPasswordChangeReplayAndStoredPayload_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.credential_state_transitions_prove_first_enrollm_850d3dd9cb` |
| `internal/modules/auth/support_integration_test.go:241` — `TestReplayAndStoredPayloadSafety` | runtime := flowtest.StartRuntime(t) | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| `internal/modules/auth/support_integration_test.go:499` — `TestRequestContracts` | runtime := flowtest.StartRuntime(t) | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| `internal/modules/auth/support_integration_test.go:347` — `TestSessionRevocation` | runtime := flowtest.StartRuntime(t) | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| `internal/modules/auth/integration_test.go:178` — `TestSessionRevocationClosesAttachedSocket_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.real_sockets_on_canonical_ws_v1_incidents_incide_8e29eb6263` |
| `internal/modules/auth/support_integration_test.go:183` — `TestSurfaceEnvelope` | runtime := flowtest.StartRuntime(t) | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| `internal/modules/auth/secret_envelopes_integration_test.go:13` — `TestTOTPEnvelopeSubstitutionAndActivation_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.secret_envelope_substitution` |
| `internal/modules/auth/integration_test.go:931` — `TestUserAdminAudit_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6` |
| `internal/modules/auth/integration_test.go:459` — `TestUserAdminLifecycle_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6` |
| `internal/modules/auth/integration_test.go:1139` — `TestUserCreateReplayReturnsOriginalCommittedResource_Integration` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6` |
| `internal/modules/auth/support_integration_test.go:28` — `TestUserListContinuationUsesLiveRows` | runtime := flowtest.StartRuntime(t) | `module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` |
| `internal/modules/auth/support_integration_test.go:109` — `TestUserListSearchTokenizesIdentifiersAndBindsCursorScope` | runtime := flowtest.StartRuntime(t) | `module.auth.integration.user_list_search_tokenizes_identifiers_and_binds_45c3e062a8` |
| `internal/modules/auth/support_integration_test.go:559` — `newSupportRouteContext` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/auth/integration_test.go:1554` — `startEnterpriseServer` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/auth/integration_test.go:1550` — `startServer` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/auth/integration_test.go:1558` — `startServerWithExtensionClaims` | harness := runtime.StartServerWithDependenciesAndEnv(t, prefix, deps, env, auth.RegisterTestRoutes()) | Helper; package consumer rows below/above |
| `internal/modules/auth/testsupport/flowtest/runtime.go:17` — `StartRuntime` | func StartRuntime(t testing.TB) *RuntimeHarness {; return &RuntimeHarness{Runtime: appsupport.StartRuntime(t)} | Helper; package consumer rows below/above |
| `internal/modules/collaboration/integration_test.go:231` — `TestClosedIncidentSocketTerminatesBeforeWritableAck` | runtime := appsupport.StartRuntime(t) | `module.collaboration.integration.closed_incident_socket_terminates_before_writabl_8057ea1833` |
| `internal/modules/collaboration/integration_test.go:377` — `TestCookieSocketRejectsUntrustedOrigin_Integration` | runtime := appsupport.StartRuntime(t) | `module.collaboration.integration.cookie_authenticated_browser_socket_upgrades_rej_73faec66da` |
| `internal/modules/collaboration/stream_integration_entrypoint_test.go:15` — `TestDurableIncidentStream_Integration` | runtime := appsupport.StartRuntime(t) | `module.collaboration.integration.durable_incident_stream_commits_intents_atomically_s_80128fe153` |
| `internal/modules/collaboration/candidate_discovery_integration_test.go:13` — `TestEntityCandidateDiscoveryDoesNotPublish_Integration` | h := appsupport.StartServer(t, "collaboration-candidate-discovery") | `module.collaboration.integration.entity_candidate_discovery_does_not_publish` |
| `internal/modules/collaboration/socket_transport_test.go:18` — `TestIncidentSocketFrameFailureContract_Unit` | runtime := appsupport.StartRuntime(t) | `module.collaboration.store.strict_frame_and_failure_contract_a3b4c5d6e7` |
| `internal/modules/collaboration/socket_session_test.go:25` — `TestIncidentSocketHandshakeResume_Unit` | runtime := appsupport.StartRuntime(t) | `module.collaboration.store.the_first_application_message_on_the_incident_so_11b91f6bd8` |
| `internal/modules/collaboration/socket_session_test.go:235` — `TestIncidentSocketHeartbeatIdleExpiry_Unit` | runtime := appsupport.StartRuntime(t) | `module.collaboration.store.presence_payloads_are_incident_scoped_and_epheme_185de72cf2` |
| `internal/modules/collaboration/socket_session_test.go:277` — `TestIncidentSocketPresenceScopeEphemeral_Unit` | runtime := appsupport.StartRuntime(t) | `module.collaboration.store.presence_payloads_are_incident_scoped_and_epheme_185de72cf2` |
| `internal/modules/collaboration/integration_test.go:79` — `TestIncidentSocketRevocationSources` | runtime := appsupport.StartRuntime(t) | `module.collaboration.support_integration.incident_socket_revocation_sources_00f908473e` |
| `internal/modules/collaboration/integration_test.go:329` — `TestResumeReplaysReplayableMessagesOnly_Integration` | runtime := appsupport.StartRuntime(t) | `module.collaboration.integration.resume_with_a_valid_replay_token_replays_replaya_1826b38095` |
| `internal/modules/collaboration/integration_test.go:27` — `TestTwoClientsPresenceReplay_Integration` | runtime := appsupport.StartRuntime(t) | `module.collaboration.integration.two_real_clients_connect_to_the_same_incident_so_257bcc588c` |
| `internal/modules/collaboration/socket_session_test.go:347` — `setupSocketIncident` | harness := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled, | Helper; package consumer rows below/above |
| `internal/modules/collaboration/integration_test.go:397` — `setupSocketIncidentWithAdminID` | harness := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled, | Helper; package consumer rows below/above |
| `internal/modules/entities/support_contract_integration_test.go:73` — `TestAuthorizationReDerivation` | Indirect composition consumer; retain domain assertions | `module.entities.support_integration.integration_surface_envelope_1366b4ce6c` |
| `internal/modules/entities/support_contract_integration_test.go:30` — `TestCSRFProtection` | Indirect composition consumer; retain domain assertions | `module.entities.support_integration.integration_surface_envelope_1366b4ce6c` |
| `internal/modules/entities/support_contract_integration_test.go:95` — `TestDefaultQueryMetaAndFieldKeyConformance` | Indirect composition consumer; retain domain assertions | `module.entities.support_integration.integration_surface_envelope_1366b4ce6c` |
| `internal/modules/entities/active_identifier_claims_integration_test.go:233` — `TestEntityActiveIdentifierClaimsAndConcurrentMatching_Integration` | runtime := appsupport.StartRuntime(t); database := runtime.PrepareServerDatabase(t, "entities-active-identifier-claims"); harness := runtime.StartServerWithDatabase(t, "entities-active-identifier-claims", database) | `module.entities.support_integration.active_identifier_claims_8f302d2a40` |
| `internal/modules/entities/candidate_discovery_integration_test.go:18` — `TestEntityCandidateDiscovery_Integration` | h := appsupport.StartServer(t, "entity-candidate-discovery") | `module.entities.integration.candidate_discovery` |
| `internal/modules/entities/create_security_integration_test.go:13` — `TestEntityCreateAuthAndCSRFFailBeforeMalformedBody_Integration` | harness := appsupport.StartServer(t, "entity_linking-entity-create-auth-csrf-order") | `module.entities.integration.host_and_identity_create_routes_reuse_exact_matc_f443e2591f` |
| `internal/modules/entities/create_idempotency_integration_test.go:14` — `TestEntityCreateIdempotencyIsActorScoped` | harness := appsupport.StartServer(t, "entity_linking-entity-create-actor-scope") | `module.entities.support_integration.integration_entity_create_idempotency_is_actor_s_3c8ed1fe5c` |
| `internal/modules/entities/origin_upsert_integration_test.go:24` — `TestEntityOriginUpsert_Integration` | harness := appsupport.StartServer(t, "entity_linking-i-4-02-host"); harness := appsupport.StartServer(t, "entity_linking-i-4-02-identity"); harness := appsupport.StartServer(t, "entity_linking-i-4-02-query-auth"); harness := appsupport.StartServer(t, "entity_linking-i-4-02-create-only-cutover") | `module.entities.integration.host_and_identity_create_routes_reuse_exact_matc_f443e2591f` |
| `internal/modules/entities/party_assignment_integration_test.go:19` — `TestEntityPartyAssignmentWrites_Integration` | h := appsupport.StartServer(t, "entity-party-assignment-writes") | `module.entities.integration.entity_party_assignment_writes_preserve_authorit_9fb842a70d` |
| `internal/modules/entities/merge_route_integration_test.go:28` — `TestExplicitMergeRoute_Integration` | harness := appsupport.StartServer(t, "entity_linking-entity-route-failure-precedence"); harness := appsupport.StartServer(t, "entity_linking-i-4-03"); harness := appsupport.StartServer(t, "entity_linking-i-4-03-version-conflict"); harness := appsupport.StartServer(t, "entity_linking-i-4-03-collision-detail"); harness := appsupport.StartServer(t, "entity_linking-i-4-03-authz"); harness := appsupport.StartServer(t, "entity_linking-i-4-03-identity") | `module.entities.integration.the_explicit_merge_route_repoints_live_fan_out_p_0ec76e8044` |
| `internal/modules/entities/support_contract_integration_test.go:118` — `TestProjectionAndWebsocketConsequences` | Indirect composition consumer; retain domain assertions | `module.entities.support_integration.integration_surface_envelope_1366b4ce6c` |
| `internal/modules/entities/support_contract_integration_test.go:45` — `TestReplayAndDivergentConflict` | Indirect composition consumer; retain domain assertions | `module.entities.support_integration.integration_surface_envelope_1366b4ce6c` |
| `internal/modules/entities/resolution_route_integration_test.go:22` — `TestResolveRoute_Integration` | harness := appsupport.StartServer(t, "entity_linking-i-4-01-resolve-conformance"); harness := appsupport.StartServer(t, "entity_linking-i-4-01-resolve"); harness := appsupport.StartServer(t, "entity_linking-i-4-01-dismiss"); harness := appsupport.StartServer(t, "entity_linking-i-4-01-revert"); harness := appsupport.StartServer(t, "entity_linking-i-4-01-access") | `module.entities.integration.the_mention_resolve_route_persists_durable_resol_d68c1befd6` |
| `internal/modules/entities/support_contract_integration_test.go:15` — `TestSurfaceEnvelope` | Indirect composition consumer; retain domain assertions | `module.entities.support_integration.integration_surface_envelope_1366b4ce6c` |
| `internal/modules/entities/support_scenario_test.go:69` — `newSupportSuite` | harness := appsupport.StartRuntime(t).StartDefaultServer( | Helper; package consumer rows below/above |
| `internal/modules/evidence/lifecycle_projection_collaboration_integration_test.go:19` — `TestAttachPublishesWorkbookWebSocketRefresh_Integration` | harness := appsupport.StartServer(t, "evidence_lifecycle-attach-websocket-refresh") | `module.evidence.integration.a_successful_evidence_attach_publishes_canonical_d63a42e346` |
| `internal/modules/evidence/attachment_authorization_integration_test.go:238` — `TestAttachRouteContract_Integration` | runtime := appsupport.StartRuntime(t); testDB := runtime.PrepareServerDatabase(t, "evidence_lifecycle-attach-route-contract"); harness := runtime.StartServerWithDatabase(t, "evidence_lifecycle-attach-route-contract", testDB); return runtime.StartServerWithDatabaseAndObjectStore(t, prefix, testDB, store) | `module.evidence.integration.attach_route_validation_authorization_route_owne_7567e19cd9` |
| `internal/modules/evidence/projection_collaboration_integration_test.go:176` — `TestAttachedEvidenceProjectionRebuild_Integration` | harness := appsupport.StartServer(t, "evidence_lifecycle-projection-rebuild") | `module.evidence.integration.attached_evidence_projection_storage_can_be_corr_933714dcfe` |
| `internal/modules/evidence/blob_create_test.go:207` — `TestBlobCreateSizeCeiling_Unit` | harness := appsupport.StartServer(t, "evidence_lifecycle-blob-size-ceiling") | `module.evidence.store.blob_create_size_ceilings_reject_oversize_byte_s_9ba60e8bdd` |
| `internal/modules/evidence/handles_test.go:84` — `TestDownloadDispositionFallback_Unit` | harness := appsupport.StartServer(t, "evidence_lifecycle-disposition") | `module.evidence.store.download_responses_use_authoritative_metadata_fo_a55dbb9aac` |
| `internal/modules/evidence/access_handle_integration_test.go:317` — `TestDownloadHandleBlobMissingDoesNotConsumeHandle_Integration` | harness := appsupport.StartServer(t, "entity_linking-download-handle-no-byte-failure") | `module.evidence.integration.preview_and_download_handle_issuance_accepts_onl_84f55ae52b` |
| `internal/modules/evidence/attachment_authorization_integration_test.go:18` — `TestEvidenceAuthorizationMatrixAC543_Integration` | harness := appsupport.StartServer(t, "evidence-authorization-matrix-ac543") | `module.evidence.integration.the_object_blob_create_route_enforces_request_sh_17b6e1b674` |
| `internal/modules/evidence/access_handle_integration_test.go:164` — `TestEvidenceHandleIssuanceReportsRegisteredUnavailableReasons` | harness := appsupport.StartServer(t, "entity_linking-evidence-handle-issue-reasons") | `module.evidence.integration.evidence_handle_issuance_reports_registered_unav_e4474a6987` |
| `internal/modules/evidence/access_handle_integration_test.go:239` — `TestEvidenceHandleRedemptionReportsRegisteredUnavailableReasons` | harness := appsupport.StartServer(t, "entity_linking-evidence-handle-redeem-reasons") | `module.evidence.integration.evidence_handle_issuance_reports_registered_unav_e4474a6987` |
| `internal/modules/evidence/access_handle_integration_test.go:14` — `TestEvidenceHandles_Integration` | harness := appsupport.StartServer(t, "entity_linking-evidence-handles") | `module.evidence.integration.preview_and_download_handle_issuance_accepts_onl_84f55ae52b` |
| `internal/modules/evidence/lifecycle_boundary_integration_test.go:14` — `TestEvidenceObjectBoundaryAndActiveContent_Integration` | harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-object-boundary"); harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-active-"+active.name) | `module.evidence.integration.dedicated_database_assertions_for_when_quarantin_f5c765cb0c` |
| `internal/modules/evidence/blob_verification_integration_test.go:18` — `TestEvidenceVerificationCancelsAndDrainsStoredRead_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartServerWithObjectStore(t, "evidence-verification-cancel", store) | `module.evidence.integration.background_verification_cancellation` |
| `internal/modules/evidence/upload_lifecycle_integration_test.go:341` — `TestExpiredSlotReplay_Integration` | harness := appsupport.StartServer(t, "evidence_lifecycle-expired-slot-replay") | `module.evidence.integration.expired_slot_replay_returns_the_same_expired_slo_6d29ac9f3a` |
| `internal/modules/evidence/handles_test.go:18` — `TestHandleIssueEmptyBodyNonIdempotent_Unit` | harness := appsupport.StartServer(t, "evidence_lifecycle-handle-issue") | `module.evidence.store.preview_handle_and_download_handle_issuance_rout_e9c8e5a3cd` |
| `internal/modules/evidence/access_handle_lifecycle_integration_test.go:15` — `TestHandleRedeemInvalidatesOnCurrentStateLoss_Integration` | harness := appsupport.StartServer(t, "evidence_lifecycle-handle-redeem-invalidates") | `module.evidence.integration.a_redeemed_preview_or_download_handle_fails_clos_cbb006eb24` |
| `internal/modules/evidence/upload_routes_integration_test.go:16` — `TestObjectBlobCreate_Integration` | harness := appsupport.StartServer(t, "evidence-blob-routes") | `module.evidence.integration.the_object_blob_create_route_enforces_request_sh_17b6e1b674` |
| `internal/modules/evidence/blob_create_test.go:20` — `TestObjectBlobCreate_Unit` | harness := appsupport.StartServer(t, "evidence_lifecycle-blob-create-route") | `module.evidence.store.post_api_v1_object_blobs_enforces_required_field_6c09c04c66` |
| `internal/modules/evidence/upload_lifecycle_integration_test.go:18` — `TestObjectUploadAtomicEvidenceCreate_Integration` | harness := appsupport.StartServer(t, "evidence-atomic-initial-blob-create") | `module.evidence.integration.initial_blob_create_atomic_replay_concealment_73c81fe329` |
| `internal/modules/evidence/projection_collaboration_integration_test.go:17` — `TestObjectUploadAttachWorkbookProjection_Integration` | harness := appsupport.StartServer(t, "evidence_lifecycle-upload-attach-projection") | `module.evidence.integration.a_real_object_store_upload_can_be_created_throug_9924ede0b0` |
| `internal/modules/evidence/upload_lifecycle_integration_test.go:268` — `TestObjectUploadCapabilityRoute_Integration` | harness := appsupport.StartServer(t, "evidence_lifecycle-object-upload-capability") | `module.evidence.integration.the_object_blob_create_route_enforces_request_sh_17b6e1b674` |
| `internal/modules/evidence/blob_create_test.go:252` — `TestPreviewPayloadCeiling_Unit` | harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled, | `module.evidence.store.blob_create_size_ceilings_reject_oversize_byte_s_9ba60e8bdd` |
| `internal/modules/evidence/objectstore_dependency_test.go:207` — `TestPublicEvidenceResponsesDoNotLeakObjectStoreIdentifiers` | harness := appsupport.StartServer(t, "evidence-redaction") | `module.evidence.integration.public_evidence_responses_do_not_leak_object_sto_9e5c545bd7` |
| `internal/modules/evidence/quarantine_integration_test.go:15` — `TestQuarantineBoundaryPreservesTwoStepAttach_Integration` | harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-quarantine") | `module.evidence.integration.dedicated_database_assertions_for_when_quarantin_f5c765cb0c` |
| `internal/modules/evidence/handles_test.go:126` — `TestSanitizeFilenameRemovesNUL_Unit` | harness := appsupport.StartServer(t, "evidence_lifecycle-disposition-nul") | `module.evidence.store.download_responses_use_authoritative_metadata_fo_a55dbb9aac` |
| `internal/modules/imports/imports_integration_test.go:1135` — `TestApplyRevalidatesTransactionCurrentState_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "imports-transaction-current-"+strings.ReplaceAll(testCase.name, " ", "-")) | `module.imports.integration.apply_transaction_current_state_bc8ad0ac32` |
| `internal/modules/imports/imports_integration_test.go:2128` — `TestCancellationAfterCommittedUnitDerivesPartialApplication_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "imports-partial-cancellation") | `module.imports.integration.partial_cancellation_outcomes_a14b66f830` |
| `internal/modules/imports/apply_cancellation_integration_test.go:19` — `TestCancellationInterruptsUncommittedImportUnit_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "imports-uncommitted-cancellation") | `module.imports.integration.uncommitted_unit_cancellation` |
| `internal/modules/imports/imports_integration_test.go:32` — `TestExtensionImportUploadEarlyFailCreatesNoDurableRows` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-import-early-fail") | `module.imports.integration.upload_admission_replay_complete_b8684935f2` |
| `internal/modules/imports/imports_integration_test.go:82` — `TestExtensionImportUploadExactReplayAndReadResources` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-import-replay") | `module.imports.integration.upload_admission_replay_complete_b8684935f2` |
| `internal/modules/imports/imports_integration_test.go:618` — `TestFreshSessionIsTheOnlyExplicitReimportWorkflow_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "imports-fresh-session-reimport") | `module.imports.integration.fresh_session_explicit_reimport_workflow_18fc21c7f5` |
| `internal/modules/imports/imports_integration_test.go:2516` — `TestImportsEvidenceCreateParity_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-import-evidence-owner-facade") | `module.imports.integration.evidence_create_parity_ac652dc3b3` |
| `internal/modules/imports/http_security_integration_test.go:19` — `TestImportsRouteCSRFSecurityMatrix_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "imports-csrf-security-matrix") | `module.imports.integration.upload_admission_replay_complete_b8684935f2` |
| `internal/modules/imports/imports_integration_test.go:703` — `TestMappingSelectApplyCreatesTimelineRows_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-import-apply") | `module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88` |
| `internal/modules/imports/imports_integration_test.go:1804` — `TestNetworkFlowImportMappingAndApplyCreatesOneAtomicTable` | runtime := appsupport.StartRuntime(t); harness := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled, | `module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873` |
| `internal/modules/imports/imports_integration_test.go:1989` — `TestNetworkFlowOwnerErrorsTranslateToSafeImportsFailures_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled, | `module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0` |
| `internal/modules/imports/imports_integration_test.go:339` — `TestSelectionLifecycleEnforcesOverlapAndRetainsSkippedMapping_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "imports-selection-lifecycle") | `module.imports.integration.selection_overlap_reselection_lifecycle_f8525366fa` |
| `internal/modules/imports/imports_integration_test.go:1682` — `TestTargetRegistryAndEntityOwnerFacade_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-import-target-registry-host") | `module.imports.integration.view_owner_write_null_application_rs08_8bb5e55b95` |
| `internal/modules/imports/imports_integration_test.go:949` — `TestTimelineOwnerUnitRollsBackWhenJournalFails_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-import-timeline-owner-rollback") | `module.imports.integration.timeline_unit_atomic_rollback_61c5e452e8` |
| `internal/modules/imports/imports_integration_test.go:1443` — `TestUnitCommitCrashRecoveryDoesNotReplayOwnerEffects_Integration` | runtime := appsupport.StartRuntime(t); first := runtime.StartServerWithDatabaseAndObjectStore(; second := runtime.StartServerWithDatabaseAndObjectStore( | `module.imports.integration.unit_commit_crash_recovery_108cd054a6` |
| `internal/modules/imports/imports_integration_test.go:67` — `TestUploadMetadataNonObjectCreatesNoDurableRows_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-import-metadata-non-object") | `module.imports.integration.import_upload_metadata_that_is_syntactically_val_6c1083ba3e` |
| `internal/modules/imports/imports_integration_test.go:174` — `TestXLSXDiscoveryUsesBoundedUsedRange_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-import-xlsx-discovery") | `module.imports.integration.xlsx_import_discovery_accepts_a_bounded_open_xml_d7c25ef6c8` |
| `internal/modules/imports/imports_integration_test.go:226` — `TestXLSXOperatorRegionCreatesDurableExactReplay_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "imports-xlsx-operator-region") | `module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912` |
| `internal/modules/imports/network_flow_harness_fault_test.go:22` — `assertNetworkFlowImportFaultConsumers` | runtime := appsupport.StartRuntime(t); h := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode:    httptestx.TestRouteModeHarnessOwned,; AdditionalRoutes: controls.Contribution().Routes, | Helper; package consumer rows below/above |
| `internal/modules/incidentbundles/routes_compatibility_integration_test.go:81` — `TestDescriptorPaginationAndCanonicalManifest_Integration` | harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-descriptor-canonical") | `module.incidentbundles.integration.incident_bundle_descriptor_reads_reject_singleto_ea18aa4fe5` |
| `internal/modules/incidentbundles/routes_admission_integration_test.go:262` — `TestExportJobAuthorizationReDerivesIncidentMembership_Integration` | harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-export-auth") | `module.incidentbundles.integration.incident_bundle_export_jobs_use_the_incident_sco_d94a8de941` |
| `internal/modules/incidentbundles/routes_admission_integration_test.go:30` — `TestExportJobIdempotencyAndDescriptor_Integration` | runtime := appsupport.StartRuntime(t); database := runtime.PrepareServerDatabase(t, "extension_profile-incident-bundle-export"); harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-incident-bundle-export", Database: database, Env: retainedRoots, TestRouteMode: httptestx.TestRouteModeDisabled}); return runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident-bundle-named-runner-recovery", Database: database, ObjectStore: harness.ObjectStore, Env: retainedRoots, TestRouteMode: httptestx.TestRouteModeDisabled}) | `module.incidentbundles.integration.incident_bundle_export_uses_the_common_job_resou_a1e51b82d7` |
| `internal/modules/incidentbundles/routes_atomicity_integration_test.go:102` — `TestImportEnvelopeFailuresCreateNoDurableState_Integration` | harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-envelope-failures") | `module.incidentbundles.integration.incident_bundle_import_upload_envelope_failures_bf5449e3d4` |
| `internal/modules/incidentbundles/routes_admission_integration_test.go:378` — `TestImportEnvelopeIdempotencyAndImportedIncidentOpen_Integration` | runtime := appsupport.StartRuntime(t); sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-source") | `module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755` |
| `internal/modules/incidentbundles/routes_atomicity_integration_test.go:16` — `TestImportFinalPublicationRechecksSubmitterAvailability_Integration` | runtime := appsupport.StartRuntime(t); sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-finalize-source") | `module.incidentbundles.integration.incident_bundle_import_final_publication_recheck_b7b07b3e93` |
| `internal/modules/incidentbundles/routes_compatibility_integration_test.go:22` — `TestIncidentBundleRetiredVersionIsRejectedWithoutEffects_Integration` | runtime := appsupport.StartRuntime(t); sourceHarness := runtime.StartDefaultServer(t, "incident-bundle-retired-version-source") | `module.incidentbundles.integration.incident_bundle_retired_version_is_rejected_with_1ec7e171fe` |
| `internal/modules/incidentbundles/routes_round_trip_integration_test.go:142` — `TestIncidentBundlesEvidenceBlobStagingCleanup_Integration` | runtime := appsupport.StartRuntime(t); sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-failure-source") | `module.incidentbundles.integration.incident_bundle_import_z_evidence_cleanup_a3d619b413` |
| `internal/modules/incidentbundles/routes_round_trip_integration_test.go:23` — `TestIncidentReferenceCatalogSurvivesUnavailablePackRoundTrip_Integration` | runtime := appsupport.StartRuntime(t); source := runtime.StartDefaultServer(t, "reference-catalog-source") | `module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` |
| `internal/modules/incidentbundles/routes_compatibility_integration_test.go:144` — `TestLegacyBundleLayoutRejectionAndHistoricalReplay_Integration` | runtime := appsupport.StartRuntime(t); source := runtime.StartDefaultServer(t, "frozen-layout-legacy-source") | `module.incidentbundles.integration.layout_legacy_conversion_replay` |
| `internal/modules/incidentbundles/routes_round_trip_integration_test.go:56` — `TestNativeSnapshotReferencesSurviveBundleRoundTrip_Integration` | runtime := appsupport.StartRuntime(t); source := runtime.StartDefaultServer(t, "native-reference-source") | `module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` |
| `internal/modules/incidentbundles/routes_extension_integration_test.go:23` — `TestNetworkFlowRetainedStateBlocksBundleExport_Integration` | harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-network-flow-block") | `module.incidentbundles.integration.network_flow_retained_state_blocks_bundle_export_7dcd6f85c2` |
| `internal/modules/incidentbundles/routes_party_portability_integration_test.go:15` — `TestPartyIncidentBundleFailuresAreClosedAndAtomic_Integration` | runtime := appsupport.StartRuntime(t); sourceHarness := runtime.StartDefaultServer(t, "party-incident-bundle-current-source") | `module.parties.portability.party_portable_public_failures_are_closed_and_at_cc2b089315` |
| `internal/modules/incidentbundles/routes_round_trip_integration_test.go:89` — `TestSupersededTimelineReplacementSurvivesImport_Integration` | runtime := appsupport.StartRuntime(t); sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-supersede-source") | `module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` |
| `internal/modules/incidentbundles/routes_helpers_integration_test.go:1179` — `startIsolatedIncidentBundleServer` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/incidentbundles/routes_helpers_integration_test.go:1188` — `startIsolatedIncidentBundleServerConfigured` | testDB := runtime.Postgres.PrepareIsolatedDatabaseT(t, prefix); bucket := runtime.S3.BootstrapBucketT(t, prefix); env := runtime.S3.Env(bucket); Endpoint:    runtime.S3.Endpoint,; AccessKey:   runtime.S3.AccessKey,; SecretKey:   runtime.S3.SecretKey,; Secure:      runtime.S3.Secure,; return runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode:    httptestx.TestRouteModeDisabled, | Helper; package consumer rows below/above |
| `internal/modules/incidentbundles/routes_helpers_integration_test.go:1184` — `startIsolatedIncidentBundleServerWithEnv` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/incidentbundles/routes_reference_verification_integration_test.go:21` — `testEmbeddedReferenceDestinationVerification` | fault = &portableCommitFaultPool{AdmittedPool: options.Postgres, rollback: commitRollback}; options.Postgres = fault | Helper; package consumer rows below/above |
| `internal/modules/incidentbundles/routes_reporting_artifacts_integration_test.go:26` — `testReportingArtifactRoundTrip` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/incidents/http_conformance_test.go:577` — `TestControlBoundaryIncidentCoreDeploymentAdminWithoutMembershipDenied` | Indirect composition consumer; retain domain assertions | `module.incidents.support_integration.control_boundary_incident_core_deployment_admin_9ba3db3d5f` |
| `internal/modules/incidents/integration_test.go:335` — `TestControlBoundaryIncidentCoreReDerivesAuthorizationImmediately_Integration` | Indirect composition consumer; retain domain assertions | `module.incidents.integration.one_same_session_role_progression_no_membership_6df5c1b9fb` |
| `internal/modules/incidents/http_conformance_test.go:581` — `TestControlBoundaryMembershipAdminDeploymentAdminWithoutMembershipDenied` | Indirect composition consumer; retain domain assertions | `module.incidents.support_integration.control_boundary_incident_core_deployment_admin_9ba3db3d5f` |
| `internal/modules/incidents/integration_test.go:339` — `TestControlBoundaryMembershipAdminReDerivesAuthorizationImmediately_Integration` | Indirect composition consumer; retain domain assertions | `module.incidents.integration.one_same_session_role_progression_no_membership_6df5c1b9fb` |
| `internal/modules/incidents/http_conformance_test.go:595` — `TestControlBoundaryTimelineRecordAndLiveDeploymentAdminWithoutMembershipDenied` | Indirect composition consumer; retain domain assertions | `module.incidents.support_integration.control_boundary_incident_core_deployment_admin_9ba3db3d5f` |
| `internal/modules/incidents/integration_test.go:353` — `TestControlBoundaryTimelineRecordAndLiveReDerivesAuthorizationImmediately_Integration` | Indirect composition consumer; retain domain assertions | `module.incidents.integration.one_same_session_role_progression_no_membership_6df5c1b9fb` |
| `internal/modules/incidents/http_conformance_test.go:585` — `TestControlBoundaryWorkbookPreferencesDeploymentAdminWithoutMembershipDenied` | Indirect composition consumer; retain domain assertions | `module.incidents.support_integration.control_boundary_incident_core_deployment_admin_9ba3db3d5f` |
| `internal/modules/incidents/integration_test.go:343` — `TestControlBoundaryWorkbookPreferencesReDerivesAuthorizationImmediately_Integration` | Indirect composition consumer; retain domain assertions | `module.incidents.integration.one_same_session_role_progression_no_membership_6df5c1b9fb` |
| `internal/modules/incidents/http_conformance_test.go:589` — `TestControlBoundaryWorkbookQueriesDeploymentAdminWithoutMembershipDenied` | Indirect composition consumer; retain domain assertions | `module.incidents.support_integration.control_boundary_incident_core_deployment_admin_9ba3db3d5f` |
| `internal/modules/incidents/integration_test.go:347` — `TestControlBoundaryWorkbookQueriesReDerivesAuthorizationImmediately_Integration` | Indirect composition consumer; retain domain assertions | `module.incidents.integration.one_same_session_role_progression_no_membership_6df5c1b9fb` |
| `internal/modules/incidents/http_conformance_test.go:261` — `TestIncidentCreateIdempotencyUsesActorAndNormalizedReplayHTTPConformance` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-u-2-04") | `module.incidents.support_integration.authenticated_incident_creation_and_normalized_r_e2cd9c99b5` |
| `internal/modules/incidents/integration_test.go:33` — `TestIncidentCreatePersistsBootstrapStateAndRollsBackAtomically_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-i-2-01-persist"); harness := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode:             httptestx.TestRouteModeHarnessOwned, | `module.incidents.integration.incident_create_persists_the_incident_row_bootst_0bd8d02f12` |
| `internal/modules/incidents/integration_test.go:185` — `TestIncidentCreateReplayAndDuplicateKeyConflictUseNormalizedState_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-i-2-02") | `module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6` |
| `internal/modules/incidents/http_conformance_test.go:237` — `TestIncidentCreateReturnsStableLocationHeaderHTTPConformance` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-u-2-03") | `module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972` |
| `internal/modules/incidents/http_conformance_test.go:24` — `TestIncidentCreationBootstrapsCreatorAndWorkbookPreferencesHTTPConformance` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-u-2-02") | `module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972` |
| `internal/modules/incidents/lifecycle_integration_test.go:19` — `TestIncidentLifecycleCloseReopenAuthorizationReplayAndTransitions_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident-lifecycle-close-reopen") | `module.incidents.integration.incident_lifecycle_close_reopen_authorization_re_6ac1f84d20` |
| `internal/modules/incidents/pagination_integration_test.go:204` — `TestIncidentListContinuationOmitsRevokedMembership` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-pagination-revoked-membership") | `module.incidents.support_integration.incident_list_uses_live_first_page_independently_f2ad5f1106` |
| `internal/modules/incidents/pagination_integration_test.go:121` — `TestIncidentListContinuationUsesLiveMembershipQuery` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-pagination-incidents") | `module.incidents.support_integration.incident_list_uses_live_first_page_independently_f2ad5f1106` |
| `internal/modules/incidents/pagination_integration_test.go:62` — `TestIncidentListSearchStatusAndCursorScope` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-pagination-incidents-search-status") | `module.incidents.integration.incident_list_search_status_and_cursor_scope_80dd540eb3` |
| `internal/modules/incidents/pagination_integration_test.go:16` — `TestIncidentListUsesLiveFirstPageIndependentlyOfTestClock` | runtime := appsupport.StartRuntime(t); harness := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeHarnessOwned, | `module.incidents.support_integration.incident_list_uses_live_first_page_independently_f2ad5f1106` |
| `internal/modules/incidents/membership_audit_integration_test.go:21` — `TestIncidentMembershipAuditRouteAuthorizationScopeFiltersAndKeyset_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident-membership-audit-route") | `module.incidents.integration.membership_audit_route_is_scope_safe_and_keyset_8e31c7f24a` |
| `internal/modules/incidents/integration_test.go:482` — `TestIncidentPatchPersistsOnlyPromotedFieldsAndAdvancesOnMaterialChange_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-i-2-04") | `module.incidents.integration.real_incident_patch_persists_only_patchable_meta_9fc8fc207f` |
| `internal/modules/incidents/extra_integration_test.go:120` — `TestIncidentPatchWritesAuditBeforeAfter` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-incident-audit") | `module.incidents.support_integration.membership_create_replay_returns_original_and_di_98e85ad975` |
| `internal/modules/incidents/extra_integration_test.go:19` — `TestMembershipCreateReplayReturnsOriginalAndDivergentConflict` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-membership-replay") | `module.incidents.support_integration.membership_create_replay_returns_original_and_di_98e85ad975` |
| `internal/modules/incidents/http_conformance_test.go:339` — `TestMembershipCreateRequiresOneSelectorClosedRolesAndNoInvitationFieldsHTTPConformance` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-u-2-06") | `module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972` |
| `internal/modules/incidents/pagination_integration_test.go:269` — `TestMembershipListContinuationUsesLiveRows` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-pagination-memberships") | `module.incidents.support_integration.incident_list_uses_live_first_page_independently_f2ad5f1106` |
| `internal/modules/incidents/extra_integration_test.go:176` — `TestMembershipMutationsWriteAuditBeforeAfter` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-membership-audit") | `module.incidents.support_integration.membership_create_replay_returns_original_and_di_98e85ad975` |
| `internal/modules/incidents/http_conformance_test.go:447` — `TestMembershipPatchAndDeleteEnforceBaseVersionAndLastAdminGuardHTTPConformance` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-u-2-07") | `module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972` |
| `internal/modules/incidents/integration_test.go:578` — `TestMembershipPatchSameRoleReturnsOKWithoutVersionOrMutationArtifact_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-i-2-07") | `module.incidents.integration.live_membership_patch_to_the_current_role_return_7a14cba87a` |
| `internal/modules/incidents/http_conformance_test.go:195` — `TestPublicRouteInventoryIncidentCoreEnvelopes` | Indirect composition consumer; retain domain assertions | `module.incidents.support_integration.public_route_inventory_incident_surfaces_d792370039` |
| `internal/modules/incidents/http_conformance_test.go:199` — `TestPublicRouteInventoryMembershipAdminEnvelopes` | Indirect composition consumer; retain domain assertions | `module.incidents.support_integration.public_route_inventory_incident_surfaces_d792370039` |
| `internal/modules/incidents/http_conformance_test.go:203` — `TestPublicRouteInventoryWorkbookPreferencesEnvelopes` | Indirect composition consumer; retain domain assertions | `module.incidents.support_integration.public_route_inventory_incident_surfaces_d792370039` |
| `internal/modules/incidents/integration_test.go:645` — `TestUnclaimedReservedFamiliesReturnCanonical404AndOutsidePathsDoNot_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "incident_membership-i-2-06") | `module.incidents.integration.reserved_but_unclaimed_family_roots_and_descenda_3cbf9bf6e6` |
| `internal/modules/incidents/inventory_helpers_test.go:46` — `newRouteFixture` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, slug) | Helper; package consumer rows below/above |
| `internal/modules/incidents/http_conformance_test.go:600` — `requireControlBoundaryInventoryDeploymentAdminWithoutMembershipDenied` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/incidents/integration_test.go:358` — `requireControlBoundaryInventoryReDerivesAuthorizationImmediately` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/incidents/http_conformance_test.go:207` — `requirePublicRouteInventoryEnvelopes` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/indicators/child_routes_integration_test.go:19` — `TestIndicatorProductionChildRoutes_Integration` | harness := appsupport.StartServer(t, "indicator-production-child-routes") | `module.indicators.integration.production_observation_and_lifecycle_routes` |
| `internal/modules/indicators/resolution_integration_test.go:23` — `TestIndicatorsRoute_Integration` | harness := appsupport.StartServer(t, "entity_linking-i-4-07-indicators") | `module.indicators.integration.indicator_create_and_query_routes_persist_canoni_e3a29b7aaf` |
| `internal/modules/jobapi/routes_integration_test.go:167` — `TestDeploymentAdminIncidentMembershipPolicy_Unit` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-jobapi-incident-admin-member-auth") | `module.jobapi.integration.the_reusable_deployment_admin_incident_membershi_b1cbd96a5a` |
| `internal/modules/jobapi/routes_integration_test.go:79` — `TestDeploymentJobAuthorizationReDerivedAtRequestTime_Unit` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-jobapi-deployment-auth") | `module.jobapi.integration.deployment_scoped_common_job_reads_and_cancel_re_85147de44e` |
| `internal/modules/jobapi/routes_integration_test.go:118` — `TestExpiredJobReadAndCancelAreMasked_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "jobapi-expired-masking") | `module.jobapi.integration.expired_job_read_and_cancel_masking` |
| `internal/modules/jobapi/routes_integration_test.go:18` — `TestIncidentJobAuthorizationReDerivedAtRequestTime_Unit` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-jobapi-incident-auth") | `module.jobapi.integration.incident_scoped_common_job_read_and_cancel_autho_821120b13c` |
| `internal/modules/links/route_projection_history_test.go:255` — `TestLinkTagProjectionHistoryQuery_Integration` | harness := appsupport.StartServer(t, "saved_view_query-i-8-03-link-tag-atomic") | `module.links.integration.link_and_tag_mutations_atomically_update_project_03f8283906` |
| `internal/modules/networkflow/routes_integration_test.go:192` — `TestNetworkFlowEffectiveResourceLimitsReachDiscovery_Integration` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.effective_resource_limit_discovery` |
| `internal/modules/networkflow/routes_integration_test.go:522` — `TestNetworkFlowGraphContributorsAndIndicatorLinkRoutes` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.bounded_graph_contributor_pipeline` |
| `internal/modules/networkflow/routes_integration_test.go:46` — `TestNetworkFlowPaginationRecovery_Integration` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.pagination_recovery` |
| `internal/modules/networkflow/routes_integration_test.go:270` — `TestNetworkFlowRoutesQueryPageAndInvalidateAfterSoftDelete` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5` |
| `internal/modules/networkflow/routes_integration_test.go:173` — `TestNetworkFlowRoutesRemainUnclaimedByDefault` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "network-flow-routes-unclaimed") | `module.networkflow.integration.network_flow_routes_remain_unclaimed_by_default_e8dabf0ee1` |
| `internal/modules/networkflow/routes_integration_test.go:1344` — `TestNetworkFlowSavedGraphLifecycleRoutes_Integration` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.saved_graph_lifecycle_v2` |
| `internal/modules/networkflow/routes_integration_test.go:229` — `TestNetworkFlowStreamingGraphContributingRowLimit_Integration` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.bounded_graph_contributor_pipeline` |
| `internal/modules/networkflow/routes_integration_test.go:1222` — `TestNetworkFlowTimeBucketSavedGraphLifecycle_Integration` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.time_bucket_saved_graph_lifecycle` |
| `internal/modules/networkflow/saved_graph_admission_integration_test.go:25` — `TestSavedGraphCutoverAdmission_Integration` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.saved_graph_cutover_v6` |
| `internal/modules/networkflow/graph_view_authority_integration_test.go:27` — `TestSavedGraphMaterializationRevalidatesSubmitterRole_Integration` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.saved_graph_lifecycle_v2` |
| `internal/modules/networkflow/table_lifecycle_integration_test.go:23` — `TestTableLifecycleAdmissionAndReplay_Integration` | runtime := appsupport.StartRuntime(t) | `module.networkflow.integration.table_lifecycle_admission_replay` |
| `internal/modules/networkflow/harness_authorization_integration_test.go:19` — `assertNetworkFlowAuthorizationConsumers` | h := claimedNetworkFlowServerWithControlsForRouteTest(t, appsupport.StartRuntime(t), "network-flow-auth-controls", "", controls) | Helper; package consumer rows below/above |
| `internal/modules/networkflow/harness_audit_integration_test.go:26` — `assertNetworkFlowCommittedAuditConsumers` | h := claimedNetworkFlowServerWithControlsForRouteTest(t, appsupport.StartRuntime(t), "network-flow-audit-controls", "", controls) | Helper; package consumer rows below/above |
| `internal/modules/networkflow/harness_worker_integration_test.go:20` — `assertNetworkFlowWorkerFaultConsumers` | h := claimedNetworkFlowServerWithControlsForRouteTest(t, appsupport.StartRuntime(t), "network-flow-worker-faults", "", controls) | Helper; package consumer rows below/above |
| `internal/modules/networkflow/routes_integration_test.go:2071` — `claimedNetworkFlowServerForRouteTest` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/networkflow/routes_integration_test.go:2087` — `claimedNetworkFlowServerWithControlsForRouteTest` | TestRouteMode: httptestx.TestRouteModeDisabled,; options.TestRouteMode = httptestx.TestRouteModeHarnessOwned; options.AdditionalRoutes = controls.Contribution().Routes; return runtime.StartServer(t, options) | Helper; package consumer rows below/above |
| `internal/modules/networkflow/routes_integration_test.go:2079` — `claimedNetworkFlowServerWithLimitsForRouteTest` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/parties/active_key_claims_integration_test.go:169` — `TestPartyConcurrentCreateAndRecoveryClaims_Integration` | runtime := appsupport.StartRuntime(t); database := runtime.PrepareServerDatabase(t, "parties-active-key-claims"); harness := runtime.StartServerWithDatabase(t, "parties-active-key-claims", database) | `module.parties.claims.party_active_claims_are_migration_safe_concurren_9ad2db3764` |
| `internal/modules/recovery/backup_metadata_integration_test.go:292` — `S3StoreForBucket` | env := runtimeHarness.S3.EnvForServiceRef(serviceRef, bucket) | Helper; package consumer rows below/above |
| `internal/modules/recovery/backup_metadata_integration_test.go:30` — `TestRealBackingStorageMetadataPersistsAndLatestLookup_Integration` | runtimeHarness := appsupport.StartRuntime(t); harness := runtimeHarness.StartDefaultServer(; if err := runtimeHarness.S3.CreateBucket(ctx, sourceBucket); err != nil {; targetDB := runtimeHarness.Postgres.PrepareIsolatedDatabaseT(t, "backup_restore-i-10-01-service-backed-target"); if err := runtimeHarness.S3.CreateBucket(ctx, targetBucket); err != nil {; reloaded.PostgresRestoreAnchor != latestCreated.PostgresRestoreAnchor \|\|; requireRetentionFloor(t, reloaded.CreatedAt, reloaded.PostgresRestoreAnchorRetainedUntil, "postgres_restore_anchor_retained_until") | `module.recovery.integration.real_postgres_backed_runtime_metadata_persistenc_d652f0998b` |
| `internal/modules/recovery/evidence_inventory_integration_test.go:24` — `TestRecoveryEvidenceInventoryClosure_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "recovery-evidence-inventory-closure") | `module.recovery.integration.evidence_inventory_closure_9f2714ac10` |
| `internal/modules/reference_data/reference_pack_integration_test.go:117` — `TestActivationDisableReverifyAndRefreshLifecycle_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` |
| `internal/modules/reference_data/reference_pack_integration_test.go:231` — `TestAdmissionQueuesBeforeVerificationAndCancelPreventsCommit_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_import_admission_returns_a_queued_9940d5be50` |
| `internal/modules/reference_data/profile_consumers_integration_test.go:21` — `TestAllCanonicalProfilesThroughProductionConsumers_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.all_profiles_consumer_boundary` |
| `internal/modules/reference_data/concurrent_publication_integration_test.go:16` — `TestConcurrentRootPublicationRejectsFrozenOtherKey_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` |
| `internal/modules/reference_data/consumer_integrity_integration_test.go:44` — `TestConsumerLossPublishesInvalidationAndAbortsSourceTransaction_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` |
| `internal/modules/reference_data/reference_pack_integration_test.go:193` — `TestFailuresRemainInactiveAndNoNetworkIsNeeded_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.disconnected_reference_pack_failure_modes_produc_d1bffa4cca` |
| `internal/modules/reference_data/reference_pack_integration_test.go:32` — `TestImportListReadReplayAndJobSummary_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_import_uses_the_shared_upload_env_95574f69b8` |
| `internal/modules/reference_data/reference_pack_integration_test.go:511` — `TestJobsRequireDeploymentAdminAtPollAndCancelTime_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_import_jobs_re_derive_route_famil_dd456a1690` |
| `internal/modules/reference_data/verification_precedence_integration_test.go:16` — `TestLiveVerificationIdentityPrecedesContentAndRuntimeChecks_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` |
| `internal/modules/reference_data/verification_precedence_integration_test.go:125` — `TestLiveVerificationResolvesExactDependenciesBeforeIndexing_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` |
| `internal/modules/reference_data/operator_import_integration_test.go:18` — `TestLocalOperatorImportUsesDeploymentWorker_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_import_uses_the_shared_upload_env_95574f69b8` |
| `internal/modules/reference_data/reference_pack_integration_test.go:303` — `TestMinimumDisconnectedBundleSeededExactly_Integration` | runtime := appsupport.StartRuntime(t); unclaimed := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled, | `module.reference_data.integration.a_fresh_disconnected_runtime_seeds_exactly_the_t_aa72222238` |
| `internal/modules/reference_data/mixed_refresh_integration_test.go:18` — `TestMixedRefreshCommitsCohortFallbackAndReplay_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` |
| `internal/modules/reference_data/reference_pack_integration_test.go:414` — `TestOptionalPackStatesDegradeOnlyOptionalSurfacesAndPreserveCoreWorkflows_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.absent_disabled_failed_and_missing_optional_refe_e5bec4b7f9` |
| `internal/modules/reference_data/reference_pack_integration_test.go:350` — `TestRefreshOmittedSelectorReplayUsesAdmittedSet_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.omitted_reference_pack_refresh_selectors_replay_a29a9f95dc` |
| `internal/modules/reference_data/consumer_integrity_integration_test.go:187` — `TestRequiredBaseLossClearsSelectionAndBlocksReadiness_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` |
| `internal/modules/reference_data/consumer_integrity_integration_test.go:272` — `TestUnavailableBasePreservesHealthyReplacement_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` |
| `internal/modules/reference_data/reference_pack_integration_test.go:373` — `TestUploadEnvelopeFailureCreatesNoDurableStateAndAdminIsRequired_Integration` | runtime := appsupport.StartRuntime(t) | `module.reference_data.integration.reference_pack_upload_envelope_failures_create_n_710dd1d70f` |
| `internal/modules/reference_data/worker_isolation_integration_test.go:17` — `TestWorkerDatabaseBarriersAreInstanceScoped_Integration` | runtime := appsupport.StartRuntime(t); bucket := runtime.S3.BootstrapBucketT(t, "reference-worker-second"); store, err := objectstore.Setup(context.Background(), objectstore.Settings{BindingKind: "managed_service", Endpoint: runtime.S3.Endpoint, AccessKey: runtime.S3.AccessKey, SecretKey: runtime.S3.SecretKey, Secure: runtime.S3.Secure, Bucket: bucket}, objectstore.Instrumentation{}); second := startReferencePackServerConfigured(t, runtime, appsupport.ServerOptions{Prefix: "reference-worker-second", ObjectStore: store, Env: runtime.S3.Env(bucket), TestRouteMode: httptestx.TestRouteModeDisabled, ConfigureRuntime: func(o *server.Options) { o.ReferenceDataComposition = secondBarrier }}) | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` |
| `internal/modules/reference_data/reference_pack_test_helpers_test.go:108` — `startReferencePackServer` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/reference_data/reference_pack_test_helpers_test.go:118` — `startReferencePackServerConfigured` | return runtime.StartServer(t, options) | Helper; package consumer rows below/above |
| `internal/modules/reference_data/reference_pack_test_helpers_test.go:111` — `startReferencePackServerWithEnv` | return startReferencePackServerConfigured(t, runtime, appsupport.ServerOptions{Prefix: prefix, TestRouteMode: httptestx.TestRouteModeDisabled, Env: env, ConfigureRuntime: func(o *server.Options) { | Helper; package consumer rows below/above |
| `internal/modules/reference_data/upload_limit_integration_test.go:18` — `testOversizedUploadAdmission` | runtime := appsupport.StartRuntime(t) | Helper; package consumer rows below/above |
| `internal/modules/reference_data/reference_pack_integration_test.go:960` — `testPersistedRootRotation` | runtime := appsupport.StartRuntime(t) | Helper; package consumer rows below/above |
| `internal/modules/reportcomposition/routes_integration_test.go:18` — `TestCompositionPublicLifecycleAndAuthority_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "composition-lifecycle") | `module.reportcomposition.integration.public_lifecycle_authority` |
| `internal/modules/reporting/reporting_integration_test.go:496` — `TestBoundaryReplayDefaultsAndActionIdempotency_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-reporting-idempotency") | `module.reporting.integration.snapshot_reporting_route_idempotency_treats_an_o_1d2451a56f` |
| `internal/modules/reporting/composition_preview_integration_test.go:22` — `TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension-profile-reporting-composition-preview", TestRouteMode: httptestx.TestRouteModeDisabled, ConfigureRuntime: func(options *server.Options) { | `module.reportcomposition.integration.preview_reporting_delegation_ecdd2bc64a` |
| `internal/modules/reporting/reporting_integration_test.go:728` — `TestExactShapesAndRouteScopedVisibility_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-reporting-shape-auth") | `module.reporting.integration.snapshot_reporting_singleton_reads_and_release_a_943c339eea` |
| `internal/modules/reporting/reporting_integration_test.go:349` — `TestExternalReleaseApprovalPublishAndStateConflicts_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-reporting-lifecycle") | `module.reporting.integration.external_releases_require_distinct_reviewer_and_0df12b0a47` |
| `internal/modules/reporting/evidence_provider_integration_test.go:23` — `TestReportingEvidenceProviderRedaction_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "reporting-evidence-provider-redaction") | `module.reporting.integration.managed_service_assertions_for_evidence_provider_0b8acd1609` |
| `internal/modules/reporting/snapshot_admission_integration_test.go:18` — `TestSnapshotAdmissionAndCancellationDoNotWaitForSourceCapture_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-reporting-async-capture") | `module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c` |
| `internal/modules/reporting/reference_binding_integration_test.go:32` — `TestSnapshotAdmissionAndRerenderRetainSetAcrossActivation_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "snapshot-reference-binding", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}}) | `module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c` |
| `internal/modules/reporting/reporting_integration_test.go:27` — `TestSnapshotReplayAndReleaseProvenanceAreStable_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension_profile-reporting-provenance") | `module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c` |
| `internal/modules/reporting/backup_recovery_integration_test.go:25` — `assertReportingPrivateBackupRestore` | Indirect composition consumer; retain domain assertions | Helper; package consumer rows below/above |
| `internal/modules/revisions/delete_restore_test.go:62` — `TestDeleteRestoreConcreteSourceAdapterMatrix_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled, | `module.revisions.delete_restore.managed_service_assertions_for_source_adapter_70c51a4a80` |
| `internal/modules/revisions/integration_test.go:73` — `TestDeleteRestoreRollbackAtomicConsequences_Integration` | harness := appsupport.StartServer(t, "history_revision-i-7-01-delete-restore") | `module.revisions.integration.covers_delete_restore_single_entry_rollback_dire_b348439e6d` |
| `internal/modules/revisions/locks_test.go:17` — `TestDestructiveOperationLocks_Unit` | harness := appsupport.StartServer(t, "history_revision-u-7-06-rollback-lock") | `module.revisions.store.covers_shared_destructive_operation_lock_precede_02f3fc18c4` |
| `internal/modules/revisions/history_test.go:117` — `TestHistoryEntryRefStability_Unit` | harness := appsupport.StartServer(t, "history_revision-u-7-02-history-ref") | `module.revisions.store.history_entry_ref_is_present_only_for_a_logical_5f0060122a` |
| `internal/modules/revisions/history_test.go:438` — `TestHistoryLogicalPageBoundaries_Integration` | harness := appsupport.StartServer(t, "history-logical-pages") | `module.revisions.integration.history_logical_page_boundaries` |
| `internal/modules/revisions/integration_test.go:338` — `TestHistoryPaginationRecordBinding_Integration` | harness := appsupport.StartServer(t, "history_revision-i-7-02-pagination") | `module.revisions.integration.history_pagination_remains_bound_to_record_id_re_a92fcdb263` |
| `internal/modules/revisions/integration_test.go:30` — `TestHistoryProjectionDependencies_Integration` | harness := appsupport.StartServer(t, "revisions-dependent-projection-refresh") | `module.revisions.integration.projection_dependencies` |
| `internal/modules/revisions/history_source_paging_test.go:19` — `TestHistorySourceFamilyPaging_Integration` | harness := appsupport.StartServer(t, "history-source-pages") | `module.revisions.integration.history_source_family_paging` |
| `internal/modules/revisions/indicator_children_test.go:29` — `TestIndicatorChildHistoryRollback_Integration` | harness := appsupport.StartServer(t, "history_revision-i-7-06-indicator-child-rollback") | `module.revisions.integration.indicator_observation_create_and_resolution_targ_6249c556e2` |
| `internal/modules/revisions/integration_test.go:437` — `TestMergeChangeSetRollback_Integration` | harness := appsupport.StartServer(t, "history_revision-i-7-05-merge-rollback") | `module.revisions.integration.merge_aware_change_set_rollback_restores_survivo_130bd55b07` |
| `internal/modules/revisions/history_test.go:21` — `TestRecordHistoryEnvelope_Unit` | harness := appsupport.StartServer(t, "history_revision-u-7-01-history-envelope") | `module.revisions.store.managed_service_assertions_for_get_api_v1_record_049278bce8` |
| `internal/modules/revisions/delete_restore_test.go:309` — `TestRestoreTombstonePreconditions_Unit` | harness := appsupport.StartServer(t, "history_revision-u-7-04-restore") | `module.revisions.store.covers_the_restore_route_contract_reviewer_admin_a2df5bc94d` |
| `internal/modules/revisions/history_test.go:175` — `TestRetainedHistoryInvariants_Unit` | harness := appsupport.StartServer(t, "history_revision-u-7-07-retained-history") | `module.revisions.store.retained_history_invariants_for_extant_records_r_b40fb75124` |
| `internal/modules/revisions/http_security_precedence_test.go:15` — `TestRevisionsMutationSecurityPrecedence_Integration` | harness := appsupport.StartServer(t, "revisions-security-precedence") | `module.revisions.store.http_security_precedence_4e7a9c2d13` |
| `internal/modules/revisions/rollback_test.go:25` — `TestRollbackSelectorUnion_Unit` | harness := appsupport.StartServer(t, "history_revision-u-7-05-rollback") | `module.revisions.store.covers_rollback_route_admission_derived_rollback_c6c889199c` |
| `internal/modules/revisions/rollback_failure_injection_test.go:17` — `TestRollbackTransactionalFailureInjection_Integration` | harness := appsupport.StartServer(t, "revisions-rollback-failure-injection") | `module.revisions.rollback.transaction_failure_injection_8c5e3f2a11` |
| `internal/modules/revisions/delete_restore_test.go:222` — `TestSoftDeleteRoutePreconditions_Unit` | harness := appsupport.StartServer(t, "history_revision-u-7-03-delete") | `module.revisions.store.covers_the_first_class_record_soft_delete_route_30cb39b83e` |
| `internal/modules/revisions/integration_test.go:634` — `TestStaleRestoreRollbackFailsClosed_Integration` | harness := appsupport.StartServer(t, "history_revision-i-7-03-stale-restore") | `module.revisions.integration.covers_stale_restore_plus_stale_or_non_reversibl_637111e642` |
| `internal/modules/savedviews/savedviews_test.go:34` — `TestSavedViewCreateDefaults_Unit` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "saved_view_query-savedviews-u-8-02") | `module.savedviews.store.managed_service_assertions_for_ordinary_saved_vi_fc464a05e5` |
| `internal/modules/savedviews/savedviews_test.go:347` — `TestSavedViewCreateEvolvesAdditiveHiddenFields_Unit` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "saved_view_query-savedviews-layout-evolution-u-8-02") | `module.savedviews.store.managed_service_assertions_for_ordinary_saved_vi_fc464a05e5` |
| `internal/modules/savedviews/read_contract_test.go:19` — `TestSavedViewIndependentReadContract_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "saved-view-independent-reads") | `module.savedviews.integration.independent_read_contract` |
| `internal/modules/savedviews/savedviews_test.go:664` — `TestSavedViewLifecyclePersistence_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "saved_view_query-savedviews-i-8-01") | `module.savedviews.integration.saved_view_create_update_duplicate_and_delete_pe_5f93d857d3` |
| `internal/modules/savedviews/savedviews_test.go:407` — `TestSavedViewSystemFixtureRouteUnavailableByDefault_Unit` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "savedviews-system-fixture-unregistered") | `module.savedviews.store.managed_service_assertions_for_ordinary_saved_vi_fc464a05e5` |
| `internal/modules/savedviews/savedviews_test.go:428` — `TestSavedViewSystemFixtureRoute_Unit` | runtime := appsupport.StartRuntime(t); harness := runtime.StartServer(t, appsupport.ServerOptions{; AdditionalRoutes: []httpapi.RouteRegistrar{savedviews.RegisterTestRoutes()},; TestRouteMode:    httptestx.TestRouteModeHarnessOwned, | `module.savedviews.store.managed_service_assertions_for_ordinary_saved_vi_fc464a05e5` |
| `internal/modules/savedviews/read_contract_test.go:224` — `TestSavedViewTimestampCompatibility_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "saved-view-timestamp-compatibility") | `module.savedviews.integration.timestamp_saved_queries_preserve_precision_and_r_993e5032bf` |
| `internal/modules/timeline/timeline_event_integration_test.go:1124` — `TestAuthorizationLifecycleAndSupersedeTransitions_Integration` | runtime := scenariotest.StartRuntime(t) | `module.timeline.integration.incident_role_authorization_review_and_supersede_62f3d384f6` |
| `internal/modules/timeline/support_integration_test.go:17` — `TestAuthorizationMatrix` | runtime := scenariotest.StartRuntime(t); harness := runtime.StartServer(t, "timeline_mutation-support-auth") | `module.timeline.support_integration.integration_authorization_matrix_2699762a78` |
| `internal/modules/timeline/resolution_integration_test.go:129` — `TestAutoResolutionEligibility_Integration` | harness := appsupport.StartServer(t, "entity_linking-u-4-08-host-auto-match"); harness := appsupport.StartServer(t, "entity_linking-u-4-08-identity-auto-match"); harness := appsupport.StartServer(t, "entity_linking-u-4-08-create-auto-match"); harness := appsupport.StartServer(t, "entity_linking-u-4-08-unresolved"); harness := appsupport.StartServer(t, "entity_linking-u-4-08-competing"); harness := appsupport.StartServer(t, "entity_linking-u-4-08-mixed"); harness := appsupport.StartServer(t, "entity_linking-u-4-08-security-order"); harness := appsupport.StartServer(t, "entity_linking-u-4-08-import-ineligible"); harness := appsupport.StartServer(t, "entity_linking-u-4-08-entities-port"); harness := appsupport.StartServer(t, "entity_linking-u-4-08-rollback"); harness := appsupport.StartServer(t, "entity_linking-i-4-08-rebuild") | `module.timeline.integration.timeline_patch_and_query_routes_auto_resolve_onl_8ab2448829` |
| `internal/modules/timeline/timeline_event_integration_test.go:1672` — `TestCanonicalIncidentWebSocket_Integration` | runtime := scenariotest.StartRuntime(t) | `module.timeline.integration.the_canonical_incident_web_socket_route_proves_h_aef38ee6e5` |
| `internal/modules/timeline/timeline_event_integration_test.go:38` — `TestCreatePatchReplayAndRollback_Integration` | runtime := scenariotest.StartRuntime(t) | `module.timeline.integration.real_http_create_patch_review_supersede_replay_m_37e5c78ca1` |
| `internal/modules/timeline/resolution_integration_test.go:1196` — `TestManualTimelineConfidenceNull_Integration` | harness := appsupport.StartServer(t, "entity_linking-i-4-09-create-add-resolved-ref"); harness := appsupport.StartServer(t, "entity_linking-u-4-09-add-resolved-ref"); harness := appsupport.StartServer(t, "entity_linking-u-4-09-resolve-item"); harness := appsupport.StartServer(t, "entity_linking-u-4-09-resolve-create"); harness := appsupport.StartServer(t, "entity_linking-u-4-09-reject") | `module.timeline.integration.timeline_create_and_patch_routes_persist_manual_ac8c9abb5b` |
| `internal/modules/timeline/timeline_event_integration_test.go:647` — `TestPatchSameFieldConflictEnvelope_Integration` | runtime := scenariotest.StartRuntime(t) | `module.timeline.integration.same_field_timeline_patch_conflicts_are_transpor_808ccc2585` |
| `internal/modules/timeline/timeline_event_integration_test.go:982` — `TestProjectionQueryUsesDeterministicRebuild_Integration` | runtime := scenariotest.StartRuntime(t) | `module.timeline.integration.the_query_route_reads_projection_rows_instead_of_2edde6eb1c` |
| `internal/modules/timeline/timeline_event_integration_test.go:894` — `TestRoughUncertainCapturePreservation_Integration` | runtime := scenariotest.StartRuntime(t) | `module.timeline.integration.rough_and_uncertain_timeline_capture_preserves_n_a868f4aa14` |
| `internal/modules/timeline/timeline_event_integration_test.go:733` — `TestRouteEnvelopeMatrix_Integration` | runtime := scenariotest.StartRuntime(t) | `module.timeline.integration.create_patch_mark_reviewed_supersede_query_and_w_90276912d7` |
| `internal/modules/timeline/timeline_event_integration_test.go:503` — `TestRouteIdempotencyIsActorScoped` | runtime := scenariotest.StartRuntime(t) | `module.timeline.support_integration.integration_route_idempotency_is_actor_scoped_ab7992bb7d` |
| `internal/modules/timeline/timeline_event_integration_test.go:1768` — `TestTimelineTimeConversionProfile` | runtime := scenariotest.StartRuntime(t) | `module.timeline.support_integration.integration_route_idempotency_is_actor_scoped_ab7992bb7d` |
| `internal/modules/timeline/timeline_event_integration_test.go:1972` — `startServer` | harness := runtime.StartServer(t, prefix) | Helper; package consumer rows below/above |
| `internal/modules/timeline/testsupport/scenariotest/harness.go:17` — `StartRuntime` | func StartRuntime(t testing.TB) *RuntimeHarness {; return &RuntimeHarness{Runtime: appsupport.StartRuntime(t)} | Helper; package consumer rows below/above |
| `internal/modules/viewschemas/viewschemas_integration_test.go:15` — `TestViewSchemasDiscoveryHTTP` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "view-schemas-discovery") | `platform.viewschema.integration.view_schemas_discovery_http_036651fc36` |
| `internal/modules/workbook/conflict_route_test.go:843` — `ConflictFixture` | harness := appsupport.StartServer(t, name) | Helper; package consumer rows below/above |
| `internal/modules/workbook/clipboard_paste_integration_test.go:392` — `TestBulkMutationsPersistOneVisibleBatch_Integration` | harness := appsupport.StartServer(t, "workbook_interaction-i-9-01-bulk-mutations") | `module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770` |
| `internal/modules/workbook/clipboard_paste_integration_test.go:441` — `TestClipboardPasteAndBulkRejectCrossIncidentTargets_Integration` | harness := appsupport.StartServer(t, "workbook_interaction-i-9-01-cross-incident-batch-targets") | `module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770` |
| `internal/modules/workbook/conflict_route_test.go:241` — `TestCollectionReviewRouteResolve_Unit` | Indirect composition consumer; retain domain assertions | `module.workbook.catalog.collection_review_fields_use_collection_value_v1_4efe5fcdae` |
| `internal/modules/workbook/integration_test.go:24` — `TestConcurrentEditsResolverPath_Integration` | harness := appsupport.StartServer(t, "collaboration-i-6-03-concurrent-resolver") | `module.workbook.integration.concurrent_edits_to_different_fields_succeed_wit_c3f74ce361` |
| `internal/modules/workbook/conflict_route_test.go:599` — `TestConflictResolutionRevalidatesStaleToken_Unit` | Indirect composition consumer; retain domain assertions | `module.workbook.catalog.keep_saved_clears_the_conflict_without_creating_c65a9c3099` |
| `internal/modules/workbook/conflict_route_test.go:413` — `TestConflictResolveDurability_Unit` | Indirect composition consumer; retain domain assertions | `module.workbook.catalog.keep_saved_clears_the_conflict_without_creating_c65a9c3099` |
| `internal/modules/workbook/workbook_mutation_integration_test.go:464` — `TestCoordinationCollections_Integration` | Indirect composition consumer; retain domain assertions | `module.workbook.integration.coordination_collection_fields_round_trip_typed_9c30c3f538` |
| `internal/modules/workbook/workbook_mutation_integration_test.go:283` — `TestCoordinationDefaults_Integration` | harness := appsupport.StartServer(t, "entity_linking-coordination-defaults") | `module.workbook.integration.communications_log_handoff_status_review_and_les_94a6f7d0c2` |
| `internal/modules/workbook/workbook_integration_test.go:515` — `TestCoordinationProjectionQueries_Integration` | harness := appsupport.StartServer(t, "workbook-seeded-surfaces") | `module.workbook.integration.parties_and_coordination_system_views_query_rout_e0808abf39` |
| `internal/modules/workbook/coordination_surfaces_test.go:231` — `TestCoordinationSavedViewsRemainAdditive_Unit` | harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled, | `module.workbook.catalog.dedicated_database_assertions_for_managed_servic_7d059c99ab` |
| `internal/modules/workbook/row_wire_test.go:451` — `TestCursorContinuationRechecksAuthorization_Integration` | harness := appsupport.StartServer(t, "saved_view_query-i-8-04-cursor-auth") | `module.workbook.integration.cursor_pagination_re_derives_live_authorization_2d2231384b` |
| `internal/modules/workbook/row_wire_test.go:483` — `TestCursorContinuationRechecksMembership_Integration` | harness := appsupport.StartServer(t, "saved_view_query-i-8-04-cursor-membership") | `module.workbook.integration.cursor_pagination_re_derives_live_authorization_2d2231384b` |
| `internal/modules/workbook/clipboard_paste_integration_test.go:261` — `TestEntityOriginClipboardPasteUsesSharedIngest_Integration` | harness := appsupport.StartServer(t, "workbook_interaction-i-9-01-entity-origin-paste") | `module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770` |
| `internal/modules/workbook/workbook_mutation_integration_test.go:460` — `TestGenericCollectionPatch_Integration` | Indirect composition consumer; retain domain assertions | `module.workbook.integration.generic_record_patch_collection_fields_use_stabl_92304f55b9` |
| `internal/modules/workbook/conflict_route_test.go:21` — `TestGridWriteConcurrencyRoute_Unit` | Indirect composition consumer; retain domain assertions | `module.workbook.catalog.every_grid_write_is_anchored_to_the_route_bound_3926ec3df4` |
| `internal/modules/workbook/row_wire_test.go:380` — `TestLiveAuthorizedCursorPagination_Integration` | harness := appsupport.StartServer(t, "saved_view_query-i-8-04-cursor") | `module.workbook.integration.cursor_pagination_re_derives_live_authorization_2d2231384b` |
| `internal/modules/workbook/note_association_guards_test.go:21` — `TestNoteAssociationHTTPGuardsAndEffects` | Indirect composition consumer; retain domain assertions | `module.workbook.integration.note_associations_public_contract` |
| `internal/modules/workbook/note_associations_integration_test.go:17` — `TestNoteAssociationsPublicPagingAndHistory` | Indirect composition consumer; retain domain assertions | `module.workbook.integration.note_associations_public_contract` |
| `internal/modules/workbook/row_wire_test.go:508` — `TestNotesFullTextExactSearch` | harness := appsupport.StartServer(t, "saved_view_query-ac185-notes") | `module.workbook.support_integration.integration_notes_full_text_exact_search_ac185_ed48fdeae3` |
| `internal/modules/workbook/workbook_mutation_integration_test.go:19` — `TestPartiesSurface_Integration` | harness := appsupport.StartServer(t, "entity_linking-parties-surface") | `module.workbook.integration.the_parties_system_view_is_discoverable_as_a_bas_7c96f60552` |
| `internal/modules/workbook/parties_integration_test.go:15` — `TestPartyLinkHelperFieldsPreserveTextIndependently_Integration` | harness := appsupport.StartServer(t, "workbook_interaction-i-9-03-party-links") | `module.workbook.integration.party_link_helper_field_writes_update_hidden_sam_014303b350` |
| `internal/modules/workbook/row_wire_test.go:575` — `TestPrefixAndNullLastOrdering` | harness := appsupport.StartServer(t, "saved_view_query-e-8-04-prefix-null-last") | `module.workbook.support_integration.integration_notes_full_text_exact_search_ac185_ed48fdeae3` |
| `internal/modules/workbook/row_wire_test.go:162` — `TestRecordChangedSparsePatchPayloads_Unit` | harness := appsupport.StartServer(t, "saved_view_query-ac368-sparse-patches") | `module.workbook.integration.full_row_and_sparse_patch_wire_families_preserve_32a4ab3be1` |
| `internal/modules/workbook/coordination_surfaces_test.go:180` — `TestRejectedCoordinationCreateEmitsNoRecordChanged_Unit` | harness := appsupport.StartServer(t, "workbook_interaction-collaboration-rejected-create-ws") | `module.workbook.catalog.dedicated_database_assertions_for_managed_servic_7d059c99ab` |
| `internal/modules/workbook/row_wire_test.go:20` — `TestRowWireFamilies_Unit` | harness := appsupport.StartServer(t, "saved_view_query-u-8-10-row-wire") | `module.workbook.integration.full_row_and_sparse_patch_wire_families_preserve_32a4ab3be1` |
| `internal/modules/workbook/conflict_route_test.go:100` — `TestSameFieldConflictHTTP_Unit` | Indirect composition consumer; retain domain assertions | `module.workbook.catalog.same_field_conflict_transport_uses_409_error_cod_4447ac22bf` |
| `internal/modules/workbook/conflict_route_test.go:497` — `TestSourceOwnerConflictContributions_Integration` | Indirect composition consumer; retain domain assertions | `module.workbook.catalog.keep_saved_clears_the_conflict_without_creating_c65a9c3099` |
| `internal/modules/workbook/tasksdecisions_conflict_contribution_test.go:17` — `TestTasksDecisionsConflictResolutionContribution_Integration` | Indirect composition consumer; retain domain assertions | `module.tasksdecisions.conflict_composition.application_workbook_conflict_resolution_contrib_10e9976746` |
| `internal/modules/workbook/conflict_route_test.go:141` — `TestTextCompareMergeDurability_Unit` | Indirect composition consumer; retain domain assertions | `module.workbook.catalog.text_compare_merge_treats_the_field_as_plain_tex_977f9eb56f` |
| `internal/modules/workbook/clear_cells_integration_test.go:17` — `TestTimelineClearCellsAuthoritativeBatch_Integration` | harness := appsupport.StartServer(t, "timeline-clear-cells") | `module.workbook.integration.timeline_clear_cells` |
| `internal/modules/workbook/clear_cells_integration_test.go:128` — `TestTimelineClearCellsConflictsRetainNull_Integration` | harness := appsupport.StartServer(t, "timeline-clear-conflicts") | `module.workbook.integration.timeline_clear_cells` |
| `internal/modules/workbook/clear_cells_integration_test.go:88` — `TestTimelineClearDateIntent_Integration` | harness := appsupport.StartServer(t, "timeline-clear-dates") | `module.workbook.integration.timeline_clear_cells` |
| `internal/modules/workbook/clear_cells_integration_test.go:159` — `TestTimelineClearPreservesRelationshipsAndHistory_Integration` | harness := appsupport.StartServer(t, "timeline-clear-history") | `module.workbook.integration.timeline_clear_cells` |
| `internal/modules/workbook/clipboard_paste_integration_test.go:122` — `TestTimelineClipboardPastePersistsOrderedMutationsAndConflicts_Integration` | harness := appsupport.StartServer(t, "workbook_interaction-i-9-01-clipboard-paste") | `module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770` |
| `internal/modules/workbook/clipboard_paste_integration_test.go:534` — `TestWorkbookBatchAdmissionHasNoPartialEffects_Integration` | harness := appsupport.StartServer(t, "workbook-batch-admission-recovery") | `module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770` |
| `internal/modules/workbook/clipboard_paste_integration_test.go:25` — `TestWorkbookClipboardFidelityAndLegacyReceipt_Integration` | harness := appsupport.StartServer(t, "workbook-clipboard-fidelity") | `module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770` |
| `internal/modules/workbook/locate_integration_test.go:17` — `TestWorkbookLocatorBoundedWindowAndOrdinaryCursor_Integration` | Indirect composition consumer; retain domain assertions | `module.workbook.catalog.locator_public_contract` |
| `internal/modules/workbook/locate_snapshot_integration_test.go:95` — `TestWorkbookLocatorConsistentReadAndTransactionLifetime_Integration` | Indirect composition consumer; retain domain assertions | `module.workbook.catalog.locator_consistent_read_c33982976c` |
| `internal/modules/workbook/workbook_startup_test.go:25` — `TestWorkbookPreferencePointers_Unit` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "saved_view_query-workbook-prefs-u-8-05") | `module.workbook.catalog.workbook_preference_routes_keep_home_sheet_ref_a_72ab108a31` |
| `internal/modules/workbook/shared_harness_test.go:61` — `TestWorkbookRouteConformance` | Indirect composition consumer; retain domain assertions | `module.workbook.support_integration.managed_service_assertions_for_shared_harness_wo_ffc6e20715` |
| `internal/modules/workbook/route_guard_characterization_test.go:28` — `TestWorkbookRouteGuardsFailBeforeMutation` | harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled,; AdditionalRoutes: []httpapi.RouteRegistrar{func(_ *http.ServeMux, deps httpapi.DependencySet) error { | `module.workbook.catalog.keep_saved_clears_the_conflict_without_creating_c65a9c3099` |
| `internal/modules/workbook/workbook_startup_test.go:318` — `TestWorkbookStartupBaseSurfaceDoesNotRequireSavedView_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "saved_view_query-workbook-startup-base-surface-i-8-02") | `module.workbook.integration.startup_selection_fallback_behaves_correctly_acr_081544c048` |
| `internal/modules/workbook/workbook_startup_test.go:154` — `TestWorkbookStartupFallback_Integration` | runtime := appsupport.StartRuntime(t); harness := runtime.StartServer(t, appsupport.ServerOptions{; TestRouteMode: httptestx.TestRouteModeDisabled, | `module.workbook.integration.startup_selection_fallback_behaves_correctly_acr_081544c048` |
| `internal/modules/workbook/workbook_integration_test.go:21` — `TestWorkbook_AllDiscoveredBaseSurfacesQueryEmptyIncident` | harness := appsupport.StartServer(t, "workbook-all-surfaces-empty") | `module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120` |
| `internal/modules/workbook/workbook_integration_test.go:83` — `TestWorkbook_AllRegisteredSurfacesHaveCreateAdmission_Integration` | harness := appsupport.StartServer(t, "workbook-all-surfaces-create-admission") | `module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120` |
| `internal/modules/workbook/workbook_integration_test.go:162` — `TestWorkbook_CoordinationDefaultQueryReturnsCreatedRows` | harness := appsupport.StartServer(t, "workbook-coordination-default-query") | `module.workbook.integration.parties_and_coordination_system_views_query_rout_e0808abf39` |
| `internal/modules/workbook/workbook_mutation_integration_test.go:152` — `TestWorkbook_EvidenceMutations` | harness := appsupport.StartServer(t, "workbook-evidence-mutations") | `module.workbook.integration.evidence_row_create_and_patch_through_workbook_p_d074e288e1` |
| `internal/modules/workbook/workbook_mutation_integration_test.go:721` — `TestWorkbook_NotesTasksAndDecisionsMutations` | harness := appsupport.StartServer(t, "workbook-notes-tasks-decisions-mutations") | `module.workbook.integration.notes_task_requests_and_decisions_mutations_thro_c325a978ef` |
| `internal/modules/workbook/workbook_integration_test.go:121` — `TestWorkbook_ProjectionBackedQueryRouteUsesCommonBoundaryBehavior` | harness := appsupport.StartServer(t, "workbook-projection-query-boundary") | `module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120` |
| `internal/modules/workbook/workbook_integration_test.go:400` — `TestWorkbook_QueryCursorContinuationUsesLiveRows` | harness := appsupport.StartServer(t, "workbook-query-live-cursor") | `module.workbook.integration.cursor_pagination_re_derives_live_authorization_2d2231384b` |
| `internal/modules/workbook/workbook_integration_test.go:477` — `TestWorkbook_QueryCursorRejectsTampering` | harness := appsupport.StartServer(t, "workbook-query-cursor-tamper") | `module.workbook.integration.cursor_pagination_re_derives_live_authorization_2d2231384b` |
| `internal/modules/workbook/workbook_integration_test.go:285` — `TestWorkbook_QueryPaginationContract` | harness := appsupport.StartServer(t, "workbook-query-pagination") | `module.workbook.integration.cursor_pagination_re_derives_live_authorization_2d2231384b` |
| `internal/modules/workbook/workbook_mutation_integration_test.go:468` — `testCoordinationCollections` | harness := appsupport.StartServer(t, "workbook-coordination-mutations") | Helper; package consumer rows below/above |
| `internal/platform/httpapi/extensiondiscovery/routes_service_test.go:12` — `TestExtensionDiscoverySessionSliding_ServiceBacked` | runtime := appsupport.StartRuntime(t); harness := runtime.StartDefaultServer(t, "extension-discovery-session") | `app.server.integration.extension_discovery_session_sliding_7f6ad80724` |

Current baseline execution is in progress; no pass is inferred from inventory.

### Application-fixture plan refinement — E-069–E-072

2026-10-26, America/New_York. The user selected refinement of the existing
application-fixture iteration, then authorized implementing the proposed
document update. Only this handoff is writable. The preceding read-only turn
changed nothing and executed no tests. This save retains the earlier handoff
delta, stable findings/contracts/slices/workflows, and historical evidence.

#### Scope and authority

| Session / evidence | Current state and supersession | Next boundary |
| --- | --- | --- |
| E-069/E-070 | Same source commit e6ed547; this handoff was already modified. Active baseline now describes this refinement rather than the clean starting state of E-065. Adopted owners remain unchanged. | No source-work or test-execution authorization; revalidate source and caller closure before WF-26 |

#### Backend boundary

| Session / evidence | Inspected decision | Next boundary |
| --- | --- | --- |
| E-070 | Four resource combinations separate required services from fixture allocation; borrowed database/store lifetime is explicit. Local Runtime retains restart bytes; existing service and HTTP owners remain. | Apply S-026–S-030 only in a later implementation task; remove forwarding layers and second-server workarounds without new abstractions |

#### Frontend boundary

Not applicable: no browser, presentation, selector, accessibility, visual or
frontend fixture change is included.

#### Contract and codegen

| Session / evidence | Inspected decision | Next boundary |
| --- | --- | --- |
| E-070 | Appsupport service requirements are not a substitute for the full scenario's catalog dependency union. Filesystem cancellation/recovery rows can shed S3; caller-owned real S3 still needs its claim. | Future authored-input-first cutover and Make regeneration; no generated or owner-specification edits in this save |

#### Tests and harness

| Session / evidence | Inspected decision | Next boundary |
| --- | --- | --- |
| E-070 | The ledger must identify guarantee, observation boundary, ownership, row and successor/retirement. Consolidate only coincident guarantees/boundaries; retain independent process/provider/restart/security/cleanup witnesses. | V-034–V-038 remain unexecuted; source inspection is not current characterization or a pass |
| E-071 | V-040 passed: canonical lint exited 0 and direct document review passed; V-039 remains historical. The configured lint population omits this handoff. | No product qualification inferred from documentation checks |

#### Security and authorization

| Session / evidence | Inspected decision | Next boundary |
| --- | --- | --- |
| E-070 | REQ-480 determines explicit route mode. The inspected Auth helper intentionally contributes auth.RegisterTestRoutes(); removing its forwarding wrapper does not remove that contribution. Shared dependency admission remains sorted and fail-closed. | Keep production authorization, TLS, purpose-specific database access and typed failure behavior; no compatibility defaults or weaker admission |

#### Open risks and next session

| Session / evidence | Bounded risk / readiness | Next action |
| --- | --- | --- |
| RB-011/RB-012 | Caller/subcase characterization and predecessor qualification are still unperformed entry gates; this refinement adds no competing workstream | Later-authorized WF-26, followed by the existing ordered chain |
| E-072 | Current-baseline comparison passed: only this handoff changed; earlier history is verbatim and index/HEAD unchanged | Recheck after final evidence annotations; later source work needs a new baseline |

#### Refinement evidence register

| Evidence | Source / baseline | Observation, artifact and applicability |
| --- | --- | --- |
| E-069 | `/tmp/cartulary-fixture-plan-refinement-lz9muvcb/baseline.json` and `original-handoff.md` | 7,654 per-path fingerprints; dirty handoff baseline, unchanged main/HEAD and unstaged-only status; raw/staged index hashes; admitted regular singly linked output |
| E-070 | Preceding read-only refinement inspection plus this save's source/owner/command revalidation | Section 2 names the source subset; Sections 4/7 record sorted admission, resource ownership, explicit route semantics and assertion dispositions. No complete transitive caller audit or test result claimed |
| E-071 | V-040; exact command below; `document-review.json` beside the baseline | Native session 81743 and outer Make exit 0; lint pass in 17,320 ms, zero tests/failures; direct structure/links/IDs/graph review passed; earlier E-067 is not reused |
| E-072 | `comparison.json` beside the baseline | Only the handoff changed among 7,654 entries; HEAD/branch/raw index/staged entries and output identity unchanged; earlier history preserved verbatim; ignored build/cache/dependency trees not exhaustively fingerprinted |

#### Documentation validation and preservation

The Make task projection, workspace admission, installed readiness, step runtime,
Markdown runner and cache path were inspected. Installed-only preparation rejects
missing prerequisites without installing them; fixing is disabled and the
workspace lock already exists. Results and static-analysis cache are external.
This command is documentation validation, not product-test execution:

```sh
GIT_OPTIONAL_LOCKS=0 \
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-fixture-plan-refinement-lz9muvcb/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-fixture-plan-refinement-lz9muvcb/static-analysis \
CARTULARY_TEST_RUN_ID=tr01-fixture-plan-refinement-20261026-01 \
COREPACK_ENABLE_NETWORK=0 HARNESS_DIAGNOSTICS=off HARNESS_ACCOUNTING=off \
make lint-markdown
```

E-071 passed: native session 81743, outer Make exit **0**, summary **pass**,
17,320 ms, zero tests and zero failures. Exact summary:
`/tmp/cartulary-fixture-plan-refinement-lz9muvcb/results/tr01-fixture-plan-refinement-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The lint configuration does not include this handoff. Supplemental direct review
in `document-review.json` verified twelve top-level sections, 24 current tables,
nine local links, stable IDs, the four-case resource matrix and the acyclic
WF-26–WF-30 chain. The focused `git --no-optional-locks diff --check` passed.
Final evidence annotations are directly reviewed; the configured linted file
population is unchanged, so they do not require another lint invocation.

E-072 passed: `comparison.json` records this handoff as the only changed entry
among 7,654 tracked/nonignored paths, with no added/deleted paths. Branch, HEAD,
raw index, staged entries and output identity are unchanged. Comparison is
rechecked after these annotations; the ignored-tree coverage limit remains.

No formatter, generator, finalizer, service operation, dependency installer or
product test ran. `agent-finalize` and retained-run maintenance are skipped
because this is documentation-only work; `RESULTS_DIR` is unset. Future
V-034–V-038 remain separate and unexecuted.

The pre-existing Section 10 history is 280,636 bytes with SHA-256
`908814228dead8324340ab598c878c3808a79f2c3746e07bfeee67a4e750f12b`.
E-072 verifies its byte-for-byte preservation. The next entry and all earlier
completion statements retain their original session meaning. T-033 and V-040
are complete; no future implementation slice is completed by this save.

### Application-fixture planning save — E-065–E-068

2026-10-26, America/New_York. The preceding read-only planning turn selected
application fixtures and acquisition cleanup with the user. This save implements
the agreed document update. Root/backend/tooling instructions, Refactor Tracker,
its format and framework, and the test-operations skill govern the bounded work.
Only this file may change; source implementation and tests remain future work.

#### Scope and authority

| Session / evidence | Current state and supersession | Next boundary |
| --- | --- | --- |
| E-065/E-066 | Clean e6ed547 baseline and established target identity retained. The active sections now describe application fixtures; object-store completion moves to explicit history. | Later authorization for WF-26; no inherited source-work permission |

#### Backend boundary

| Session / evidence | Inspected decision | Next boundary |
| --- | --- | --- |
| E-066 | One options-based appsupport path, private acquisition dependencies, fresh default namespaces and retained explicit restart inputs; existing physical and HTTP owners remain | Complete affected caller/assertion ledger before any migration |

#### Frontend boundary

Not applicable: this iteration changes no frontend presentation, browser helper,
selector, renderer, accessibility or visual baseline.

#### Contract and codegen

| Session / evidence | Inspected decision | Next boundary |
| --- | --- | --- |
| E-066 | Authored rows and owners remain projections; two known filesystem scenarios can shed unnecessary S3 claims; fixture admission and incident fault admission are distinct | Future owner-first selector/dependency cutover; no generated writes now |

#### Tests and harness

| Session / evidence | Inspected decision | Next boundary |
| --- | --- | --- |
| E-066 | Existing exact rows and source assertions support the plan, not a passing baseline. Incidental wrappers, global patching and package serialization are retired by design. | V-034–V-038 unexecuted; preserve cancellation, restart, isolation, cleanup and real-service witnesses |
| E-067 | V-039 lint and direct document review passed; documentation validation is separate from production assurance | No historical or documentation result qualifies future source work |

#### Security and authorization

| Session / evidence | Inspected decision | Next boundary |
| --- | --- | --- |
| E-066 | Preserve explicit route admission, fault guard, typed/redacted errors, purpose-specific PostgreSQL connections and verified object transport | No production API, policy, credential, trust or failure-class change |

#### Open risks and next session

| Session / evidence | Bounded risk / readiness | Next action |
| --- | --- | --- |
| RB-011/RB-012 | No unresolved design choice; current characterization and predecessor qualification remain unperformed entry gates | Revalidate baseline and relevant source, then execute only after later source-work authorization |
| E-068 | Sole-file/history/index comparison passed; no unexpected changes | Revalidate the new baseline before later source work |

#### Current evidence register

| Evidence | Source / baseline | Observation, artifact and applicability |
| --- | --- | --- |
| E-065 | `baseline.json`, `original-handoff.md` in `/tmp/cartulary-app-fixture-plan-wdyx9lju/` | 7,654 per-path content/identity records, clean main/HEAD, raw/staged-index hashes and admitted tracker identity; ignored trees excluded |
| E-066 | Prior read-only planning inspection; same clean source revalidated on save; Section 2 sources and current machine owners | Findings F-038–F-043 and decisions C-022–C-026; 192 direct importer files/35 packages is search inventory only; no test execution or comprehensive assertion audit |
| E-067 | V-039 documentation command and direct review below | Outer Make exit 0; lint passed in 16,990 ms with zero tests/failures. Direct review passed; configured Markdown lint excludes this handoff |
| E-068 | Pre-write identity/content recheck and final `comparison.json` | Pre-write matched E-065. Only this handoff changed among 7,654 entries; HEAD/branch/raw index/staged entries/output identity unchanged; historical log preserved |

#### Documentation validation and preservation

The admitted lint invocation uses installed prerequisites, fixing disabled,
network-disabled Corepack and external results/cache. Its wrapper path was read:
Make task projection, workspace admission, installed readiness, run-step and
Markdown lint/cache helpers. The workspace lock already exists. No formatter,
generator, finalizer, service command, dependency installer or product test runs.

```sh
GIT_OPTIONAL_LOCKS=0 \
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-app-fixture-plan-wdyx9lju/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-app-fixture-plan-wdyx9lju/static-analysis \
CARTULARY_TEST_RUN_ID=tr01-application-fixture-plan-20261026-01 \
COREPACK_ENABLE_NETWORK=0 HARNESS_DIAGNOSTICS=off HARNESS_ACCOUNTING=off \
make lint-markdown
```

E-067 passed: native session 62810, outer Make exit **0**, summary **pass**,
16,990 ms, zero tests and zero failures. Exact summary:
`/tmp/cartulary-app-fixture-plan-wdyx9lju/results/tr01-application-fixture-plan-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The configured lint population excludes this handoff. Supplemental direct review
in `document-review.json` beside the baseline verified twelve top-level sections,
23 current tables, nine local links, stable IDs and finding/contract/slice/validation
traceability, the five-workflow acyclic chain and six existing initial catalog rows.
`git --no-optional-locks diff --check -- docs/handoffs/testing-remediation-01.md`
passed. Final evidence annotations were directly reviewed; the linted file
population did not change.

E-068 passed: `comparison.json` records exactly one changed entry, this handoff,
among 7,654 tracked/nonignored entries, with no added/deleted paths. HEAD, branch,
raw index, staged entries and output identity are unchanged. The earlier
208,404-byte Section 10 log is preserved byte-for-byte; SHA-256
`ab7b2a12f6734023d15c8e72ee0478fec85365c9cc0cad9bc19aba97cf96d636`.
The earlier active sections below retain their content with headings demoted;
their authority and results remain historical. Preservation is rechecked after
the final evidence annotations. No exhaustive ignored-tree preservation or current
production qualification is claimed. T-027 and V-039 are DONE; every future source
workstream and test obligation remains unexecuted.

### Completed object-store iteration — historical active-plan snapshot

The following Sections 1–9 and 11–12 are the preceding iteration's saved active
text, retained for its assertion ledger, stable IDs and decision history. Their
statements of current authorization/completion apply only to E-057–E-064. The
application-fixture plan above supersedes them as the active plan. Heading levels
are demoted only; the original source is also retained in E-065's starting copy.

#### 1. Scope and Source Posture

Document date: 2026-10-26. **Current action: completed review and corrections**, authorized by the
user's follow-up request to review guarantees, regressions and complexity, fix
issues, run focused validation and update this handoff.
Execution order is WF-20 → WF-21 → WF-25 → WF-22 → WF-23 → WF-24.
The earlier document-only saves remain historical. This invocation admits the
bounded source, test, caller and routing changes below, their Make validation and
handoff updates; it does not authorize normative amendments or unrelated redesign.
Artifact mode remains **normal for this bounded iteration**.

Repository: `/home/jochi/code/cartulary`; established supplied and canonical target
`.`; kind: repository directory; iteration label: `object-store-fixture-simplification`.
The handoff is `docs/handoffs/testing-remediation-01.md`. Its parent
components are not symlinks; the output is a regular, singly linked file. The
existing tracker's repository identity is retained. The [modular refactor
framework](cartulary_modular_refactor_planning_framework.md) and Refactor Tracker
format supply planning structure, not behavioral authority.

The implementation change boundary is `internal/testutil/s3test`, object-store composition
in `tools/testservices`, directly affected process/application/recovery consumers,
and their authored routing and generated projections. Keep physical service
ownership distinct from borrowed fixture access. Remove accidental test
constraints at their source; do not add compatibility adapters, public injection
frameworks, another service manager, a fixture-policy language, or a new runner.
Production APIs/data schemas, provider replacement, PostgreSQL internals, scheduler
architecture, browser presentation, and performance-fixture policy are excluded.
Affected consumers remain regression witnesses, not invitations to redesign them.

WF-00–WF-19 remain completed at their historical baselines. E-047 records the
PostgreSQL closure's 22 compatible roots and 42 selected rows; those results do
not qualify this object-store iteration. The separate open WF-33 in [the
preceding handoff](testing-remediation.md) is neither resumed nor closed here.
Historical instructions inside this handoff describe their original sessions;
they do not enlarge the current bounded implementation request.

| Authority / status | Required distinction and applicable clauses |
| --- | --- |
| [Core 00](../spec/00_document_set_status_and_precedence.md), current normative owner allocation | Adopted owners govern requirements. Tests, typed projections and this handoff do not establish specification completeness. Tests, generators, runtime metadata, conformance and release evidence must not consume Markdown. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md), `adopted/current` | REQ-804: explicit attachment, independent physical owners, cancellation, exact recovery proof and quarantine. REQ-410 and §11: declared capabilities/dependencies, namespace ownership and cleanup. §11.4: shared mutation-readiness contract, deadlines and package admission. REQ-414 and §9: bounded diagnostics and primary/secondary failure classification. §§3/8: exact routing and truthful evidence. |
| [Core 04](../spec/04_security_deployment_and_conformance.md), §1.3 | Preserve verified transport, the shared S3 transport policy and secret confinement. Pins remain machine-owned in `tools/toolchain_pins.json`; no new pin inventory or qualification claim. |
| [Domain](../domain.md), §16 / §16.1 | Vocabulary and owner navigation: object storage backs binary evidence; buckets and keys are not public evidence identity. No domain-model change is planned. |
| [NLSpec guidance](../research/nlspec-spec.md) | Supporting guidance on requirements versus implementation choices; not adopted authority for fixture APIs, test names, layouts or mock behavior. |
| [Core 05](../spec/05_claim_publication_and_benchmark_reproducibility.md) | No timing, performance or release-publication claim. Functional fixtures do not replace distinct claim-bearing qualification. |

No inspected owner requires a mixed `Harness`, three owned-start wrappers,
package-global test substitutions, fake deletion that succeeds while reporting
failure, wait-strategy description matching, or one particular cleanup request
count. No specification amendment is proposed to preserve those arrangements.
Binding validation and quarantine corrections implement existing requirements;
they are explicitly identified separately from mechanical movement below.

| Baseline item | Implementation entry, E-057 | Current comparison | Limits |
| --- | --- | --- | --- |
| Branch / commit | `main`; `9e6fa4053e2de4367ce106606da02e74fafc72d5` | Same commit; no commits or branch changes | Historical evidence retains its own source scope |
| Index / staged state | No staged changes; entry SHA-256 `5370d11e07f02f067cde89dc1b15d83aa93adfb200c82034f6ab8db1a31b9f04`; raw-index SHA-256 `443632109419c6b351af3837dc3abbf68e1eed208ab4403c7bbfb01a285284c8` | Staged entries unchanged; raw index metadata differs during implementation, with no staged changes | Raw index byte identity is not claimed |
| Worktree / untracked | Pre-existing edits confined to this handoff; 7,649 tracked/nonignored entries | Exactly 20 changed/added/removed entries, including this handoff and the generated topology index; 7,654 current entries; E-063, with six existing entries updated during review at E-064 | Ignored build/cache/runtime trees excluded |
| Tracker | 229,205 bytes; inode 348826; SHA-256 `43960f81a5e8c28a8853e425330bb017d9b939f504aa2b3de02c8fd24b0cd7cc` | Existing edits and all prior historical log text preserved | E-053/E-056 describe the preceding save, not this entry |
| Ignored outputs / dependencies | Installed runtime/frontend prerequisites and cached pinned Go | Invocations use explicit result roots and pinned Go executable/PATH; no dependency or pin updates | No exhaustive ignored-tree fingerprint |

Implementation preservation material is in
`/tmp/cartulary-s3-implementation-p5bpj7gf/`: `baseline.json`,
`original-handoff.md`, `pre-finalize.json` and `frozen-source.json`.
E-048–E-056 remain historical document-save evidence. Source accounting includes
tracked/nonignored entries under `.`; generated products remain projections and
never behavioral owners. None of the executable changes consumes Markdown.

#### 2. Current-State Repository Inventory

This table records the inspected **pre-refactor state**, revalidated at WF-20
(E-057/E-058). The original five S3 Go files contained 30 top-level tests and
3,076 lines; those counts were inventory, not a reduction quota. Current owners,
source groups and qualified successors are recorded in E-059–E-063. The bounded
caller inspection does not claim an exhaustive audit of the repository.

| Path / kind | Disposition / inspection depth | Responsibility and inbound callers | Dependencies / tests / projections | Owner / risk / evidence |
| --- | --- | --- | --- | --- |
| `internal/testutil/s3test/s3test.go` | Directly inspected, authored | Container ownership, borrowed attachment, cached client, readiness, namespace creation/reset/cleanup, diagnostics and attribution; used by testservices, appsupport, process tests and performancefixturelifecycle | testcontainersx, suiteservices, S3 transport, private resource journal; 25 tests in companion file | Split physical owner from fixture operations; C-017–C-020; E-049 |
| `s3test/s3test_test.go` | Directly inspected, authored | Startup/attachment, mutation probe, package admission and real namespace tests | Global substitution helpers; fake readiness state; real service fixtures; five catalog rows across the package | Consolidate assertions, localize dependencies and strengthen absence proof; F-028–F-031/F-033; E-049/E-054 |
| `s3test/object_cleanup.go` | Directly inspected, authored | Exact-prefix pagination, bounded delete batches and independent emptiness proof | Narrow private client port; no new algorithm/framework needed | Keep useful boundary; C-019; E-049 |
| `s3test/object_cleanup_test.go` | Directly inspected, authored | Five tests cover cleanup protocol faults, package quarantine, real pagination and pre-client ownership reservation | Local cleanup fake already has a narrow port; package cases still patch globals | Move package cases to their owner and preserve cleanup faults; C-019; E-049 |
| `s3test/tls.go` | Directly inspected, authored | Owner TLS directory, server/proxy identities and root-certificate access | tlstest, suiteservices runtime root and container configuration | Move private material ownership only; verified client trust remains; C-017/C-020 |
| `tools/testservices/main.go`, `main_test.go`, `partial_ownership.go` | Directly inspected affected composition, callback, recovery and publication bodies/assertions | Sole external owned-start caller; broker probe bucket, lease publication, browser namespace reset/cleanup; two replaceable object-store globals | Existing dependency seams, partial startup journal, bounded failure summaries; lifecycle-contract row | Keep orchestration; use local object-store seam, no new manager; C-017/C-020 |
| `internal/app/serverprocess/process_fixture_harness_test.go`, relevant config/bootstrap callers | Directly inspected setup/teardown and affected assertions; remaining bodies metadata-only | TestMain borrows services, then calls misleading `StopShared`; evidence process uses fixture bucket | Existing app.server process selectors and borrowed owner lifetime | Migrate local transport cleanup; no product assembly changes; C-017/C-021 |
| `internal/testutil/appsupport/runtime.go`, runtime guards; `performancefixturelifecycle/lease.go` | Directly inspected service acquisition and consequential namespace cleanup; remaining bodies metadata-only | Composite fixture acquisition and snapshot consumer attachment | Existing support guard, managed-service and process regression witnesses | Keep consumer semantics and performance policy; C-017/C-019 |
| `internal/platform/objectstore/s3transport/connection.go`; suiteservices env/dependency/event boundaries | Directly inspected shared connection policy, local close and relevant binding/journal paths | Verified HTTPS, single-attempt product requests, captured suite inputs and cleanup authority | Local transport Close releases idle connections; it does not terminate a service | Keep production transport; bind support inputs in S-025; C-017/C-020 |
| Native boundary `fixtures_linux_test.go`, `selected_services_linux_test.go` | Directly inspected affected lifecycle/selection boundaries; remaining bodies metadata-only | Owned/borrowed physical integration and nested exact-row selection | Existing broker/runtime and service fixtures | Retain independent physical evidence; migrate changed nested row IDs; C-017/C-021 |
| Affected `tools/test_families` rows, verification owners, task/catalog/topology and generated policies | Directly inspected affected selectors and ownership/projection declarations | Authored routing; a real reset is currently classed `unit` | Make generation and generated/schema/catalog/identity checks are future work | Repair authored inputs atomically; F-032/C-021 |
| Instructions, framework, owner clauses, command recipes and lint paths | Directly inspected relevant sections | Authority, future verification selection and document-only check admission | Installed-only readiness, fixing disabled, external caches/results | E-050 historically; E-054/E-055 for this refinement; no Markdown added to executable inputs |
| Remaining entries / ignored roots | Metadata-only / excluded as declared in Section 1 | No movement proposed | Text searches do not establish absence of dynamic/external consumers | Next boundary: newly encountered owned-start, attachment, namespace cleanup or exact-selector consumer |

#### 3. Module Boundary Diagnosis

The problem is mixed ownership and test coupling inside an existing support
package, not a missing platform layer. Keep `s3test` as the package boundary.
Responsibility-oriented files are sufficient; do not create a generic fixture
manager or force PostgreSQL and S3 into a shared lifecycle abstraction.

| Responsibility | Current location | Recommended owner / disposition | Evidence / finding / contract |
| --- | --- | --- | --- |
| Container acquisition, partial settlement, TLS material and physical close | Mixed `Harness`, three start wrappers | **Split** into `OwnedStore`; testservices keeps composition | E-049; F-027; C-017/C-020 |
| Borrowed client, attachment reuse and local transport release | Shared singleton plus shared/attached flags | **Keep/simplify** in `Harness`; separate mechanical extraction (S-021) from bound-admission correction (S-025) | E-049/E-054; F-028; C-017 |
| Mutation-readiness algorithm and diagnostics | Existing shared probe plus repeated test scaffolding | **Keep** algorithm; **split** local test dependencies and owner-specific wiring assertions | E-049; F-029/F-030; C-018/C-020 |
| Namespace ownership, package admission, reset and cleanup | Harness plus existing cleanup module | **Keep** lifecycle; **move** package tests to coherent namespace group and require typed absence plus sibling payload proof | E-049/E-054; F-029–F-031/F-033; C-019 |
| Source assertions, exact selectors and evidence classification | Two mixed unit umbrellas, separate origins row and split real rows | **Consolidate** by observation boundary; integration class for real service work | E-049/E-050; F-032; C-021 |
| Provider, scheduler, PostgreSQL and performance policy | Independent existing owners | **Defer** redesign; retain affected regression boundaries | User-selected scope; no independent change slice |

#### 4. Public Contract and Behavior Freeze Map

Freeze owner-required guarantees, not test implementation details. No HTTP,
WebSocket, domain mutation/query, saved-view, projection, revision, public evidence
identity or UI-selector change is proposed. Public Make commands and evidence
schemas retain their current contracts. Internal support APIs and row identities
may change atomically without aliases.

| Contract / surface | Observed implementation and adopted owner | Required behavior / disposition | Inspected witnesses / prerequisites / trace |
| --- | --- | --- | --- |
| C-017 — ownership and attachment | Mixed Harness, process singleton and StopShared; REQ-804/410, §11 | Preserve explicit dependencies/attachment, independent physical owners, partial recovery authority and borrower nontermination. Separate types and bind cached admission to its inputs; changed bindings cannot bypass admission. | Startup/attachment tests, testservices callbacks, process TestMain and native boundaries; current characterization required by RB-008; F-027/F-028; S-020/S-021/S-025; V-025/V-026/V-031/V-029 |
| C-018 — mutation readiness | One probe used by owned, attached, broker and package paths; §11.4 and startup retry table | Preserve put, exact-size head, delete and not-found proof; bounded same-lane polling, terminal capability rejection, separate startup/readiness windows, parent cancellation and cleanup on every outcome. No timeout expansion or product retry. | Mutation probe, owned-start and package admission cases; add deterministic cancellation/failure witnesses as needed; F-029/F-030; S-021/S-022; V-026/V-027/V-029 |
| C-019 — namespace lifecycle | Existing reserved ownership, package lock/quarantine, stable reset and cleanup algorithm; REQ-804 and §11 | Preserve exact ownership before risky acquisition, unique scopes, package reuse/concurrency, failed-generation rejection, independent empty proof, sibling protection, stable reset and final deletion. No cleanup success inferred from a fake's error path or an arbitrary read error; absence requires typed not-found, and surviving siblings retain their payloads. | All cleanup/package tests and four real-service tests; F-029/F-031/F-033; S-022; V-025/V-027/V-029 |
| C-020 — security and failure evidence | TLS roots/owner keys, S3 transport, typed readiness projection and testservices summaries; Core 04 §1.3, REQ-414, §9 | Preserve verified TLS, exact secure origins, private keys/credentials, primary versus secondary errors, timeout exit 3, cancellation and capability classifications, bounded public/retained diagnostics and correct journal destination. | Origins/configuration, cleanup diagnostic, setup-envelope and composition publication assertions; F-027–F-030; S-021/S-025/S-022; V-026/V-031/V-027/V-029 |
| C-021 — selection and assurance | Five S3 rows; real ResetBucket case labelled unit; §§3/8 and REQ-410 | Preserve meaningful independent qualification contexts and exact service/capability declarations. Give each assertion one coherent selector; retired names need no compatibility routing. A subset is not complete-owner acceptance. | Authored selectors, native nested selections, app/platform consumer rows; F-032; S-020/S-023/S-024; V-025/V-028/V-029 |

#### 5. Coupling and Boundary Findings

`must_fix` is limited to the named slices and acceptance conditions, not unrelated
cleanup. Findings describe the pre-refactor defects. E-060 reproduced cached-binding
bypass; E-059–E-063 record their fixes and qualification without adopting the
old tests' incidental constraints. F-027 is resolved by E-059, F-028 by
E-059/E-060, F-029/F-030/F-031/F-033 by E-061, and F-032 by E-062; E-063
qualified all successors together at that checkpoint. E-064 records the
follow-up review, its missed failure paths and their focused requalification.
F-034–F-037 are resolved by E-064; these corrections do not rewrite the earlier
checkpoint outcomes.

| Finding / classification | Evidence and consequence | Required action / compatibility disposition | Contract / slice / validation |
| --- | --- | --- | --- |
| F-027 / must_fix | Harness holds Container, TLS directory, client, shared/attached flags and bucket state; owned entrypoints return that same type. StopShared describes owner shutdown although ordinary startup only attaches. | Give physical ownership one type and constructor; keep borrowed local transport cleanup separate. Migrate direct callers, remove wrappers and StopShared; no forwarding stubs. | C-017/C-020; S-021; V-026/V-029; E-049 |
| F-028 / must_fix | Startup/readiness/package operations and testservices object-store composition use mutable package globals. StartShared returns the cached handle without comparing the new binding; later guards/events reread ambient values. | S-021 introduces private instance dependencies and independent test admission objects. S-025 separately corrects cached-binding bypass, ambient event/dependency reads and terminal-failure retention under REQ-804/410. No public injection API or generic service registry. | C-017/C-020; S-021/S-025; V-026/V-031/V-029; E-049/E-054 |
| F-029 / must_fix | Fake readiness DeleteNamespace removes namespace/objects even when recordAndFail returns an error, allowing no-residue assertions to succeed on a purported failure. | Default failure leaves state unchanged; explicit cases model response loss after mutation. Deleting a nonempty namespace rejects without recursively erasing objects. Assert retained uncertainty/cleanup outcome rather than fabricate success; preserve the earlier readiness failure as primary. | C-018/C-019/C-020; S-022; V-027/V-029; E-049 |
| F-030 / must_fix | Wait-strategy test matches dependency description text; cleanup tests freeze exact batch/list counts; short real timers and repeated readiness setup mix protocol, controller and publication responsibilities. | Assert mapped-endpoint admission, deadline separation, bounded batches, complete traversal and postconditions. Use local deterministic fault seams and small readable case tables; retain true boundary constraints and distinct publication witnesses. | C-018/C-020; S-022; V-027/V-029; E-049 |
| F-031 / should_fix | Package quarantine sits in cleanup-algorithm tests; repeated real roundtrip/prefix setup and separate readiness/package scaffolds obscure ownership. | Group namespace lifecycle assertions; consolidate redundant real setup while preserving an independent stable-reset scenario. Keep narrow cleanup and readiness fakes separate: they model different protocols. | C-019; S-022; V-027/V-029; E-049 |
| F-032 / must_fix | S3 assertions span two mixed unit umbrellas and origins, while a real reset has unit evidence class. Native selected-service tests reference exact existing row IDs. | Three coherent service-free groups (ownership/attachment, readiness/diagnostics, namespace contracts), plus real integration namespace/reset evidence; atomic authored/generated/nested-selector cutover. No aliases or duplicate execution to preserve old names. | C-021; S-023; V-028/V-029; E-049/E-050 |
| F-033 / must_fix | `TestCleanupPrefixPreservesSiblingObjects` and `TestResetBucketPreservesNamespaceAndProvesMutation` accept any HeadObject error as deletion proof; transport or authorization failure can satisfy the assertion. The sibling check proves availability but not payload preservation. | Require the existing typed not-found classification for deleted objects, fail on other errors, and compare surviving sibling bytes with their original payload. Require successful new mutation after stable reset. | C-019/C-021; S-022; V-027/V-029; E-054 |
| F-034 / must_fix | Readiness cancellation during polling sets outcome `cancelled` but drops the cancellation from the returned cause; package setup then reports `harness/fixture_error`. The earlier test cancelled before readiness instead. | Retain polling termination alongside the original probe error; deterministic cancellation between probes must preserve the interrupted envelope and clean probe state. | C-018/C-020; S-022; V-033; E-064 |
| F-035 / must_fix | The readiness controller counts terminal cleanup errors as primary capability failures, counts cleanup timeouts as probe timeouts, and can hide a transient cleanup failure behind a later successful attempt. Pure probe tests did not exercise the controller boundary. | End admission on uncertain cleanup, retaining the original probe stage/cause/class and secondary cleanup status. Qualify transient, deadline and cancellation cases through the controller and setup envelope. | C-018/C-019/C-020; S-022; V-033; E-064 |
| F-036 / should_fix | The cancelled-waiter test can cancel before its goroutine enters admission; it can pass without exercising waiting. Runtime/lease/mode binding changes and post-deletion publication failure also lacked direct negative witnesses. | Coordinate on the wait context's Done observation, cover all required binding categories, and prove journal-write failure remains failure after a real deletion. Keep existing rows and narrow fakes; add no runtime test hooks or routing aliases. | C-017/C-019/C-020; S-025/S-022; V-033; E-064 |
| F-037 / must_fix | A missing fixture CA passes attachment-shape validation. The shared transport then resolves its process-level system trust, potentially using ambient inputs rather than the captured fixture binding. Existing incomplete-binding cases omitted trust. | Require an explicit fixture CA in attachment/recovery binding before readiness or network acquisition. Keep system-trust behavior in the production transport unchanged; managed fixture owners already publish the CA. | C-017/C-020; S-025; V-033; E-064 |

The shared mutation probe, exact-prefix cleanup algorithm, borrowed client cache,
package reuse, broker publication checks and real provider witness each have a
useful purpose. Keep them; simplification is not a deletion quota. Scoped source/routing changes are now authorized. No adopted-owner conflict or
necessary specification amendment was found.

#### 6. Refactor Workstreams

The prerequisites in this table are authoritative. Execution order is WF-20 →
WF-21 → WF-25 → WF-22 → WF-23 → WF-24; existing IDs retain their identity.
There is no optional branch. Save each exit checkpoint before its successor.
Every slice migrates affected selectors with its source, before final routing
consolidation where necessary. S-025 is the separate owner-required correction;
mechanical movement must not silently include it.

| Workflow | Name / class | Prerequisites | Goal / likely files | Validation | Checkpoint and binary exit |
| --- | --- | --- | --- | --- | --- |
| WF-20 | Characterization / root | Current implementation authorization; E-057 revalidation | Complete assertion/caller ledger and focused baseline; s3test and consequential callers | V-025 | All 30 tests and consequential assertions have dispositions, exact baseline roots and classified gaps; known defects are not frozen as requirements |
| WF-21 | Ownership and local dependencies / chain | WF-20 exit | OwnedStore/Harness split, private operations and direct testservices/process migration | V-026 | Ownership, partial settlement and local cleanup qualified; obsolete lifecycle wrappers/global test substitutions removed; admission correction remains separately tracked |
| WF-25 | Bound attachment correction / chain | WF-21 exit | Immutable attachment/dependency/event inputs, reuse validation, terminal failure and cancellable waiting; recovery uses saved binding | V-031 | Negative binding, failure-retention and cancellation witnesses pass; no fresh admission in recovery cleanup or ambient journal redirection |
| WF-22 | Assertion consolidation / chain | WF-25 exit | Repair fake semantics, strengthen typed absence/sibling proof, organize readiness/namespace tests and consolidate real setup | V-027 | Every retained guarantee has qualified successor coverage; incidental string/count/mock constraints retired; arbitrary errors cannot prove deletion |
| WF-23 | Routing consolidation / chain | WF-22 exit | Authored family/verification inputs, nested selections and generated projections | V-028 | Three service-free groups and two real integration scenarios have exact selectors and correct capability/evidence declarations; no duplicate/dangling/alias rows |
| WF-24 | Final qualification and handoff / chain | WF-23 exit | Finalize, freeze source, qualify compatible portfolio and close handoff | V-029 | Required distinct boundaries pass; canonical evidence, cleanup and source limits recorded; no broader acceptance inferred |

#### 7. Proposed Refactor Slice Plan

| Slice / kind | Findings / contracts | Entry and readiness | Intended change / validation | Rollback and binary completion |
| --- | --- | --- | --- | --- |
| S-020 / characterization | F-027–F-033; C-017–C-021 | **complete**, E-058 | Complete assertion/subcase and caller dispositions, establish V-025 baseline, record missing witnesses and known defects separately | Do not encode incidental old behavior as a baseline obligation. Exit is WF-20 checkpoint; unexpected failures require diagnosis |
| S-021 / structural preservation | F-027/F-028; C-017/C-018/C-020 | **complete**, E-059; RB-008 resolved | Separate physical and borrowed lifetimes, localize dependencies and migrate callers; V-026 | Revert source/tests/callers/selectors together, without aliases; ownership and settlement witnesses pass without claiming the S-025 correction |
| S-025 / owner-required admission correction | F-028; C-017/C-020 | **complete**, E-060; RB-010 resolved | Bind inputs and journals, reject changed cached bindings, retain terminal failures and cancel waiting locally; qualify recovery's saved-binding path; V-031 | Revert the correction and its witnesses together if unsafe; keep F-028 open. Exit requires negative-case coverage; no product API/data migration |
| S-022 / test and fixture simplification | F-029/F-030/F-031/F-033; C-018/C-019/C-020/C-021 | **complete**, E-061 | Correct fake semantics, consolidate setup and assertion groups, require typed absence and sibling bytes; V-027 | Restore affected assertions/fixtures together if successor gaps emerge; no compatibility mock behavior; exit is WF-22 checkpoint |
| S-023 / routing cutover | F-032; C-021 | **complete**, E-062 | Consolidate semantic rows and real integration classification; retain verification identities with unchanged meaning; regenerate through Make; V-028 | Restore authored inputs/source and regenerate together; every intended source selector resolves once in the catalog |
| S-024 / final qualification | All current findings/contracts | **complete**, E-063; RB-009 resolved | Finalizer before broader verification, frozen compatible portfolio and saved completion; V-029 | Keep incomplete work explicitly incomplete; no historical-result rewrite or unsupported acceptance claim |

All implementation slices are authorized subject to their entry gates. S-025 corrects existing
REQ-804/410 obligations; no normative amendment or user-facing compatibility
migration is proposed. Introduce its negative witnesses before changing admission,
record expected failures against the defect, and require passing corrected
behavior before WF-22. Historical success cannot satisfy these gates.

##### Implemented ownership and interface

- `OwnedStore` owns container acquisition, startup settlement, TLS material,
  broker probe-bucket lifetime and physical close. One
  `StartOwned(ctx context.Context, options StartOptions) (*OwnedStore, error)`
  returns this owner, including a partial handle when settlement is uncertain.
  Existing before/acquired/settled callbacks retain
  recovery ordering and become owner-typed; testservices retains orchestration.
- `Harness` owns borrowed client access and namespace operations. Composition is
  explicit, without embedding that exposes physical shutdown. Retain `Start` and
  `StartShared` for ordinary consumers. Local transport release must not terminate
  the container, delete the broker bucket, remove owner keys or erase quarantine.
- Replace `StartOwnedWithLabels`, `StartOwnedWithOptions` and `StopShared` at every
  direct caller, then delete them. Process teardown releases its borrowed client;
  it does not report termination of a shared testcontainer.
- S-021 extracts a private process admission object for necessary reuse; tests
  construct independent objects. S-025 then binds a copied set of
  attachment, dependency and event-destination inputs at entry. Validate every
  call against the admitted binding; changed inputs fail instead of using stale
  credentials or a different journal.
  Successful reuse and terminal failure are explicit; waiters honor cancellation
  without shutting down an owner or cancelling another caller's admission.
  Do not add cross-service registries or replacement-generation machinery.
- Keep operation values private and limited to startup, attachment, readiness or
  package lifecycle. Testservices uses its existing dependency style with an
  instance-local object-store seam; no exported fake constructor is necessary.
- Preserve the one mutation probe and cleanup algorithm. Keep the error-returning
  package admission core; its testing.TB wrapper only publishes setup failure and
  registers successful release. Isolate test cases without replacing global
  functions or parsing fatal text.
- A recovery cleanup caller retains its saved exact binding and ownership proof;
  migrating its Harness construction must not trigger fresh admission, provisioning
  or cleanup of unrelated resources. Owner TLS/key cleanup remains recoverable
  through the existing journal/runtime, not a new evidence schema.
- Keep only three service-free assertion groups: ownership/attachment,
  readiness/diagnostics, and namespace lifecycle. Consolidate real roundtrip,
  pagination, sibling protection and final cleanup into one namespace scenario;
  retain stable reset as a separate real scenario. Both are integration evidence.
  Keep publication and native process/lifecycle witnesses distinct.

##### Assertion successor and retirement ledger

Every baseline top-level test is named once below (30 total). These are assertion
dispositions; executed successor evidence is recorded at the slice checkpoints
and final source in E-063. The WF-20 supplement below completes
subcase-level mapping and consequential caller assertions before movement; counts and test
names themselves are not acceptance requirements. Each removed assertion must
name its successor guarantee and observation boundary, or explain why it only
constrains incidental behavior. Missing coverage is a prerequisite, not a reason
to preserve misleading assertions or add compatibility behavior.

| Existing test | Successor / retirement disposition | Contract / validation |
| --- | --- | --- |
| `TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness` | Local admission cases retain explicit binding and no physical acquisition in S-021; S-025 qualifies same-binding reuse and changed-binding rejection | C-017; V-026/V-031 |
| `TestStartSharedRejectsMissingAttachment` | Missing/incomplete attachment and dependency rejection before I/O/publication; recheck against captured inputs in S-025 | C-017; V-026/V-031 |
| `TestBucketNamesAreUniqueAcrossSimulatedProcesses` | Namespace identity cases retain valid bounded names and cross-process noncollision | C-019; V-027 |
| `TestObjectStoreContainerWaitStrategyOnlyWaitsForPortMapping` | Retire description-string match; preserve mapped-endpoint-before-readiness behavior and declared startup bound in ownership cases | C-017/C-018; V-026 |
| `TestSeaweedFSS3AllowedOriginsDefaultSupportsLoopbackBrowserPortRange` | Ownership/configuration cases retain exact secure loopback origins and reject wildcard/plaintext assumptions | C-020; V-026 |
| `TestOwnedObjectStoreAppliesContainerConfiguration` | One owner startup group retains labels, single-master topology and TLS configuration | C-017/C-020; V-026 |
| `TestOwnedObjectStoreReadinessTimeoutIsTerminalForTheStartedLane` | Owner terminal-failure matrix: no replacement after readiness starts; exact cleanup and bounded evidence | C-017/C-018; V-026 |
| `TestOwnedObjectStoreCleanupFailureRemainsSecondaryTypedEvidence` | Same owner group retains partial handle, memoized settlement result, primary cause and diagnostic redaction | C-017/C-020; V-026 |
| `TestOwnedObjectStoreRetriesOnlyPreReadinessStartupAndUsesFreshReadinessContext` | Preserve allowed pre-readiness retry and startup/readiness budget separation; parent cancellation still applies | C-018; V-026 |
| `TestOwnedObjectStoreDoesNotRetryAuthenticationReadinessFailure` | Terminal capability case in owner matrix; no new container, preserved cleanup/classification | C-018/C-020; V-026 |
| `TestObjectStoreMutationProbeRetriesTransientStagesOnSameLane` | Readiness controller cases retain bounded transient put/head retries on one lane, without short-timer call-count assumptions | C-018; V-027 |
| `TestObjectStoreMutationProbeRejectsCapabilitiesImmediately` | Readiness protocol/controller cases retain immediate typed rejection and owned-probe cleanup | C-018/C-020; V-027 |
| `TestObjectStoreMutationProbeDeadlineRetainsCleanupEvidence` | Deterministic deadline case with bounded terminal failure and truthful secondary cleanup | C-018/C-020; V-027 |
| `TestObjectStoreMutationProbePreservesPrimaryFailureDuringCleanupFailure` | Repair fake deletion semantics; assert original put failure, cleanup failure and unresolved state; separate response-loss case | C-018/C-019/C-020; V-027 |
| `TestAttachedObjectStoreUsesMutationProbeContract` | Attached wiring case proves shared probe use, unique child scope and preservation of broker parent bucket | C-017/C-018; V-027 |
| `TestPackageBucketAdmissionProbesWithoutRemovingTheBucket` | Package wiring case retains mutation sequence/size proof, clean probe and preserved reusable bucket | C-018/C-019; V-027 |
| `TestObjectStoreMutationProbeRejectsSizeMismatch` | Protocol negative case retains exact size verification and cleanup outcome | C-018; V-027 |
| `TestCreateBucketRejectsOmittedServiceBeforeClientAcquisition` | Bound dependency guard rejects before client/network acquisition; captured-input correction belongs to S-025 | C-017; V-026/V-031 |
| `TestPreparePackageBucketCoreClassifiesPutDeadlineAndCleanupFailure` | Package failure matrix retains error-returning core and timeout envelope; publication stays independently tested | C-019/C-020; V-027 |
| `TestPreparePackageBucketCoreRejectsCapabilityAndCancellation` | Package failure matrix retains terminal capability/cancellation classes and no successful release | C-019/C-020; V-027 |
| `TestPreparePackageBucketCoreReusesPackageBucket` | Namespace lifecycle group retains same package identity, reset-before-reuse and readiness-before-admission | C-019; V-027 |
| `TestPreparePackageBucketCoreAdmitsDistinctPackagesConcurrently` | Independent-package case retains concurrent admission and separate namespaces; no global test hooks | C-019; V-027 |
| `TestHarnessStartsSeaweedFSS3AndRoundTripsObjects` | Consolidated real namespace lifecycle retains exact payload roundtrip and idempotent final cleanup; retire misleading implication that Start owns startup | C-017/C-019; V-027 |
| `TestCleanupPrefixPreservesSiblingObjects` | Consolidated real namespace lifecycle requires typed not-found for removed keys and exact surviving sibling payloads; other read failures fail the assertion | C-019; V-027 |
| `TestResetBucketPreservesNamespaceAndProvesMutation` | Keep independent real stable-reset scenario: same bucket, typed not-found for prior objects, new mutation/roundtrip succeeds; route as integration | C-019/C-021; V-027/V-028 |
| `TestObjectCleanupBatchesAndProvesExactEmptyNamespace` | Retain 0/1/boundary/multi-page cases, bounded batches and fresh empty proof; retire unnecessarily exact request totals | C-019; V-027 |
| `TestObjectCleanupRejectsUncertainCompletion` | Retain list/delete/verification faults, partial errors, residue, malformed pages/tokens, out-of-prefix keys, cancellation and absent bucket | C-019/C-020; V-027 |
| `TestPackageBucketFailedGenerationCannotBeReadmitted` | Move to package lifecycle group; retain create/reset/readiness/cancelled-reset poisoning, cleanup identity and no repeated mutation | C-019; V-027 |
| `TestPinnedObjectStoreBatchCleanupTraversesMultiplePages` | Consolidated real namespace scenario retains more than one provider page, <=1,000-key batches, fresh empty proof and exact sibling payload; no fixed two-request requirement | C-019; V-027 |
| `TestOwnedBucketReservesCleanupBeforeClientFailure` | Namespace ownership case retains journal reservation before acquisition and no false success event | C-019/C-020; V-027 |

Keep testservices failure-summary/publication assertions distinct from pure probe
assertions: they qualify public classification, retained serialization and exit
propagation. Keep native owner/borrower/selection witnesses distinct from real S3
protocol coverage. Add fault cases for incomplete bindings, separate instances
with conflicting ambient inputs, pending-waiter cancellation, terminal admission
failure, partial acquisition/publication failure and borrower local-close safety
where the baseline lacks them. A preserved test function is not proof that its
assertions cover these boundaries.

##### WF-20 assertion and caller supplement — E-057

The preceding ledger's row numbers (1–30, in order) identify the baseline tests.
These details complete the assertion-level dispositions. No planned successor is
credited as passing before its own checkpoint. Known F-028/F-029/F-030/F-032/F-033
defects are not requirements even when their existing tests pass.

| Ledger rows | Assertions/subcases carried to the named successor | Retirement or missing witness to qualify |
| --- | --- | --- |
| 1–2 | Explicit endpoint/credentials/TLS/probe binding, one admission and no physical start; missing attachment returns no published fixture | Container-field inspection becomes structural absence of shutdown capability plus borrower-close witness. Add incomplete and changed bindings, retained failure and cancelled waiters in S-025 |
| 3 | DNS-safe names, <=63 characters, alphanumeric ends, distinct repeated and simulated-process scopes | Literal sample names/count eight are data, not requirements |
| 4–6 | Startup bound, mapped endpoint available before readiness, labels, single standalone master, exact HTTPS loopback origins and no wildcard/plaintext | Retire wait dependency's description text. Add TLS file confinement and owner/borrower lifetime assertions at S-021 |
| 7–10 | Readiness deadline/capability rejection never replaces a started lane; single settled close, partial handle on uncertain cleanup; internal secondary cause, public redaction and bounded phase/stage/outcome/count fields; permitted transient pre-start retry with fresh readiness budget and parent context | One terminal acquisition/close is causal evidence, not an incidental total. Add before/acquired/settled ordering, cancellation during acquisition and publication failure; keep readiness correction separate |
| 11–14 | Transient put/head recovery on same lane; terminal create capability rejection; deadline and cleanup evidence; primary put cause survives secondary namespace-delete failure | Replace timer-sensitive minimum retries by coordinated transient recovery/deadline cases. Retire false no-residue assertion after failed namespace deletion; separately model mutation then response loss |
| 15–17 | Attached probe uses existing broker namespace; package probe preserves reusable namespace; put → exact-size head → delete → typed absence, no probe residue; size mismatch fails with cleanup | Protocol sequence is causal. Namespace create/delete prohibition protects borrowed ownership. Add typed non-absence failures; failed fake operations cannot silently mutate |
| 18–20 | Omitted object-store dependency before client acquisition; package core returns no release on timeout; infra/readiness_timeout stage/count/cleanup envelope; capability→harness and cancellation→interrupted | S-025 binds guards to saved inputs. Testservices separately proves serialized/public exit behavior; pure probe tests need not duplicate that publication boundary |
| 21–22 | Stable package identity, create once then reset-before-readmission, readiness each admission; distinct packages can hold different leases concurrently | Preserve per-package serialization, not a global lock or global hooks. Use completion channels; timeout only bounds a deadlock failure |
| 23–25 | Real payload roundtrip, repeated final cleanup, owned-prefix removal, sibling protection; independent stable reset keeps bucket and admits a new put/get | Merge roundtrip/prefix into pagination scenario. Replace arbitrary HeadObject errors with typed absence; compare sibling and new-mutation bytes, not mere availability |
| 26 | Empty, single, 1,000-boundary and 2,507-key cleanup; no owned residue, sibling retained, each batch 1–1,000, independent empty proof | Retire exact batch/list totals. Observe final independent listing and bounded complete deletion rather than a prescribed partition |
| 27 | `list`, `delete`, `verify` retain cause; `partial` redacts provider details; `residue`, `nil_page`, `nil_result`, `missing_truncation`, `missing_token`, `duplicate_token`, `cyclic_token` reject uncertainty; `outside_prefix`, `empty_key`, `common_prefix` do not mutate; `cancel_list`, `cancel_delete` retain cancellation; `missing_bucket`, `missing_at_verification` accept typed absence | Preserve finite pagination/cycle rejection; do not freeze a universal three-request ceiling. Cleanup fake remains narrow and separate from readiness |
| 28 | `create`, `reset`, `readiness`, `cancelled_reset`: failed generation retains cause/cleanup identity, releases lock, cannot mutate or retry upon reuse | Move to namespace lifecycle group; no compatibility reset of quarantine |
| 29–30 | Real 1,007-key traversal and sibling survival; journal reserves cleanup identity before failed client acquisition, no false bucket-created event | Merge row 29 into real namespace scenario, bounded batches replace exact [1,000,7]; row 30 keeps journal truth and gains saved-destination failure cases |

| Consequential caller / exact qualification | Assertions and migration boundary |
| --- | --- |
| testservices lifecycle row (A) | Service-specific S3 2-minute acquisition bound, original caller context, before/acquired/settled partial journal ordering, probe publication, primary/secondary readiness diagnostics, retained serialization and public exit 3. Replace only S3 global substitutions with a local seam |
| native physical/borrowed/selected-service rows (B) | Physical lifecycle and persistent borrower nontermination are independent of unit fakes. Selected-service nested rows must follow any renamed S3 selector in the same slice |
| app support guard and evidence-upload process rows (C) | Explicit validated service dependency before fixture construction; standalone process still uploads real evidence. TestMain releases local S3 transport after StartShared; it gains no physical owner |
| managed object-store adapter row (D) | Production adapter exercises the managed verified-TLS fixture binding; keep this independent of fixture protocol unit tests |
| browser recovery cleanup in testservices | Use the saved environment and exact bucket ownership proof, without fresh admission or provisioning; capture journal destination. Qualify binding failures/publication failures in S-025 |
| appsupport and performancefixturelifecycle | Retain borrowed acquisition, scoped namespace cleanup and existing performance policy. Their Harness references must not gain owner methods; no unrelated hook or scheduler redesign |

##### Guarantee and retirement decisions

| Required guarantee / owner | Meaningful successor and observation boundary | Incidental constraint retired / deletion gate |
| --- | --- | --- |
| Physical owner versus borrower; REQ-804/402 | Ownership unit cases plus native physical/borrowed lifecycle; partial ownership, cancellation, one settlement and borrower local-close safety | Mixed Harness flags, redundant start wrappers and StopShared are not compatibility contracts; V-026 qualifies their removal |
| Exact attachment and recovery binding; REQ-804/410 | Independent admission instances; incomplete/changed binding rejection, same-binding reuse, terminal failure, cancelled waiter and saved-binding cleanup | Cached success cannot bypass new binding checks; ambient environment or journal mutation is not a supported rebind; V-031 |
| Readiness mutation and deadlines; §11.4 | Shared probe protocol cases plus owned/attached/package wiring; exact size, typed not-found, bounded same-lane polling and primary/secondary failures | Dependency description strings and scheduling-sensitive retry counts are removed; retain causal ordering and declared budgets; V-026/V-027 |
| Exact cleanup and truthful uncertainty; REQ-804/402, §11 | Narrow cleanup fake covers bounded batches, malformed pages/tokens, partial failures, cancellation, fresh emptiness proof and no out-of-prefix mutation | Exact request totals and fake recursive namespace deletion are removed. Explicit mutation-then-error cases do not redefine default failure semantics; V-027 |
| Isolation, stable reset and data preservation; REQ-804, §11 | Real namespace scenario proves roundtrip, multi-page cleanup, typed absence and unchanged sibling bytes; separate reset proves same identity and new mutation | Duplicate real setup and acceptance of arbitrary HeadObject errors are removed only with these successor observations; F-033/V-027 |
| Package reuse and quarantine; REQ-804, §11.4 | Namespace lifecycle cases prove reset-before-reuse, readiness-before-admission, distinct-package concurrency and failed-generation rejection | Global substitution scaffolds and misplaced quarantine tests are consolidated; failure cause and cleanup identity remain; V-027 |
| Secure transport and truthful publication; Core 04 §1.3, REQ-414, §9 | Configuration/diagnostic unit cases plus distinct testservices serialization/exit and managed-adapter/process witnesses | Repeated protocol assertions at publication boundaries may be removed; serialization, classification, secret confinement and real transport witnesses may not; V-026/V-027/V-029 |
| Exact selection and declared dependencies; §§3/8, REQ-410 | Three service-free groups, two real integration scenarios and nested selection witness | Historical row names, duplicate routes and real-reset-as-unit classification are retired atomically; existing verification identities remain when their meaning is unchanged; V-028 |

Consolidation is justified by guarantees and independent observation boundaries,
not test-count reduction. Do not replace narrow fakes with a general S3 simulator,
expose a public injection API, or introduce a generic clock framework. Use local
controllable waits/contexts where timing matters; production deadlines remain
unchanged. WF-20 establishes the baseline before source or test movement.

#### 8. Validation Plan

**The original implementation passed WF-24 at E-063; the follow-up review is
qualified separately by V-033/E-064.** Public discovery was
revalidated at entry (E-057): `make help ALL=1` and
`make task-guide OWNER=<owner>` for `harness.browser`, `platform.objectstore`,
and `app.server`, with installed-only preparation and diagnostics/accounting off.
Source/catalog inspection supplies the exact
starting selections below. WF-20 revalidated them at E-058; E-063 records the
qualified final successors. No historical pass is promoted to the current source.

V-024 and V-030 retain their historical documentation-check identities. V-031
qualifies the separate admission correction; V-032 covers the prior document
refinement only; E-063 records this implementation's final documentation checks. No historical validation ID or result is repurposed.

| Validation | Findings/contracts/slices / stage | Exact selection or command / expected artifact | Gate |
| --- | --- | --- | --- |
| V-025 | All; S-020; pre-move characterization | Groups A–D below through Make; exact roots, source identity, native exits and assertion ledger | Required baseline before S-021; classify known defects and unexpected failures separately, without freezing incidental behavior; diagnose before rerun/broadening |
| V-026 | F-027/F-028; C-017/C-018/C-020; S-021; post-slice | Ownership/attachment successors in A, testservices row, applicable B/C/D boundary witnesses | No mixed shutdown capability or global hook substitution; uncertain settlement remains recoverable. S-025 owns the separate binding correction |
| V-031 | F-028; C-017/C-020; S-025; post-correction | Attachment successors in A plus testservices recovery/publication and applicable B/C witnesses; exact rows recorded after migration | Changed/incomplete bindings fail before I/O; reuse retains terminal failure; waiter cancellation is local; saved recovery binding and event destination remain exact |
| V-027 | F-029–F-031/F-033; C-018–C-021; S-022; post-slice | Readiness/namespace successors in A; real namespace/reset scenarios; testservices publication witness | Truthful fake failures, typed absence, unchanged sibling payload and successful post-reset mutation qualified before redundant assertions are deleted |
| V-028 | F-032; C-021; S-023; post-slice | `make generate`; `make source-check CHECK=generated`, `CHECK=schema`, `CHECK=catalog`, `CHECK=identity`; `make lint CHECK=go`, `CHECK=backend-boundary`; selected-service row | Authored inputs/projections and exact selectors agree; real reset has integration evidence class |
| V-029 | All; S-024; final regression | `make agent-finalize`, review output, freeze source; compatible successor A–D union plus V-028 static checks; audit every exact root with `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all` | Native exits, valid evidence and required cleanup pass; do not claim whole-owner, release or performance acceptance |
| V-030 | Historical document save only | Earlier documentation checks, retained verbatim in E-051/E-052 | Historical result; not this refinement's validation |
| V-032 | Historical document refinement only | External-output installed-only `make lint-markdown`, direct tables/links/IDs/dependency/assertion review, `git --no-optional-locks diff --check` and baseline comparison | E-055/E-056; configured lint excludes this handoff, so direct review is required; no product assurance claim |
| V-033 | F-034–F-037; post-completion review | Negative ownership/readiness cases, five S3 rows plus publication, finalizer/freeze, Go lint, generated/catalog checks and native borrower boundary; exact-root audits and documentation review | Implementation and documentation checks passed, E-064; explicit fixture CA, preserved primary/secondary failure and cancellation, terminal cleanup uncertainty, deterministic waiting and truthful deletion publication; prior E-063 results remain historical |

All implementation commands run from `/home/jochi/code/cartulary`, with fresh explicit
`CARTULARY_TEST_RUN_ID`, repository-pinned prerequisites and native session/exit
capture. Use `CGO_ENABLED=0` for the existing nested Go-runner profile; retain
current resource claims, fixture scopes and deadlines. Do not widen service
acquisition or make unavailable infrastructure a passing skip. Current-source
pre-move characterization passed at E-058; each successor checkpoint was saved
before its dependent changes.

##### Exact starting selections

A — all five existing S3 rows plus the composition/publication witness:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.s3test_helper_contract,harness.browser.unit.service_free_contract_assertions_for_object_stor_aac986f098,harness.browser.unit.object_store_secure_origins,harness.browser.integration.object_store_fixture_roundtrip,harness.browser.unit.managed_service_assertions_for_object_store_fixt_7b174abeda,harness.browser.unit.testservices_lifecycle_contract
```

B — distinct native physical ownership, persistent borrowing and nested selection:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.native_borrowed_fixture_lifecycle,harness.browser.integration.native_selected_service_lifecycle
```

C — application dependency guard and evidence-upload process consumer:

```sh
make test-slice OWNER=app.server ROWS=app.server.support_unit.incident_create_commit_fault_requires_validated_efdc344831,app.server.process.the_standalone_server_process_can_create_evidenc_4ca93ab308
```

D — production adapter against the managed fixture binding:

```sh
make test-slice OWNER=platform.objectstore ROWS=platform.objectstore.support_integration.managed_service_object_store_binding_4d6aaf4bd9
```

These are exact baseline commands, not permanent names. Record qualified
successors when A is consolidated; migrate nested uses of its old selectors in
B in the same slice. Recompute the final union against target plans so nested or
already compatible execution is not repeated without a missing obligation or
distinct qualification purpose. Changed evidence serialization would additionally
require its existing evidence-accounting owner selection; this plan changes no
schema or serializer. Newly discovered affected callers get exact row mappings
before their source changes, not an automatic full-suite run.

Full CI, visual/golden/UI-review, benchmarks, release qualification and preceding
WF-33 are excluded. No production behavior in those scopes is planned to change.
Finalization left `RESULTS_DIR` unset because there is no compatible successful
full warm-check root. E-063 records the retained-maintenance skips; the scoped
portfolio does not substitute for full-owner or release qualification.

| Invocation group | Side effects / admission | Current result |
| --- | --- | --- |
| Discovery and source inspection | Public help/task guides and file reads; no tests/services/generation | Revalidated for implementation at E-057; E-053/E-054 retain the preceding document-only discovery |
| V-025–V-029 and V-031 | Scoped implementation, services, test results, generation and finalizer writes | V-025/V-026/V-031/V-027/V-028 passed at E-058–E-062; V-029 passed at E-063 |
| V-033 | Five existing S3 source/test files and this handoff; focused tests and required finalization/checks | Passed, E-064: six compatible audited graph roots, seven selected test rows, two generated-policy rows; separate documentation lint/direct review passed |
| V-030 | Historical invocation | E-051/E-052 retained; not current validation |
| V-032 | Existing workspace lock; installed-only Node/frontend readiness; lint fixing disabled; result/cache outputs external | Passed, E-055/E-056; direct review covers the handoff excluded by configured lint |

#### 9. Top-Level Work Tracker

| Work item | Workstream | Status | Dependencies / evidence | Exit condition |
| --- | --- | --- | --- | --- |
| T-018 — initial object-store plan save / documentation | Historical save | DONE | E-048–E-052; V-030 | Sole-file update, reviewed plan/history, truthful checks and preservation proof |
| T-019 — current characterization | WF-20 | DONE | S-020; V-025; E-057/E-058 | Current baseline and complete assertion/caller dispositions saved |
| T-020 — ownership and local dependencies | WF-21 | DONE | T-019; RB-008; S-021/V-026 | Ownership/settlement guarantees qualified; wrappers/globals removed; S-025 remains separate |
| T-025 — bound attachment correction | WF-25 | DONE | T-020; RB-010; S-025/V-031 | Binding rejection, terminal failure, waiter cancellation and saved recovery inputs qualified |
| T-021 — test consolidation | WF-22 | DONE | T-025; RB-009; S-022/V-027 | Meaningful successors qualified; truthful fake behavior and typed absence replace misleading assertions |
| T-022 — routing cutover | WF-23 | DONE | T-021; RB-009; S-023/V-028 | Authored/generated/nested selectors and execution classes agree |
| T-023 — final qualification and handoff | WF-24 | DONE | T-022; RB-009; S-024/V-029; E-063 | Frozen compatible portfolio, cleanup and bounded acceptance recorded |
| T-024 — refine object-store plan / documentation | Historical save | DONE | E-053–E-056; V-032 | Refined plan saved, documentation checks passed and sole-file preservation verified |
| T-026 — post-completion review and correction | Scoped follow-up | DONE | F-034–F-037; V-033; E-064 | Correct identified gaps, qualify exact focused portfolio and preserve historical checkpoints |

T-001–T-017, WF-00–WF-19, F-001–F-026, C-001–C-016,
S-000–S-019, V-001–V-024, E-001–E-047 and RB-001–RB-007 are
reserved historical IDs. Document completion cannot complete a future workflow.

#### 11. Open Questions and Blockers

| ID | Blocker / affected scope | Needed authority or evidence | Status | Resolution condition |
| --- | --- | --- | --- | --- |
| RB-008 | Current pre-move characterization; S-021 | Current authorization (E-057), V-025 roots and completed subcase/caller dispositions | RESOLVED, E-058 | Current A–D baseline and assertion/caller dispositions qualify movement |
| RB-010 | Separate attachment correction; S-025 | WF-21 exit plus intended negative witnesses for F-028; V-031 qualifies the correction | RESOLVED, E-059/E-060 | Ownership split and V-031 correction qualified before consolidation |
| RB-009 | Ordered successor gates; S-022–S-024 | WF-25/WF-22/WF-23 exits, respectively, with V-031/V-027/V-028 evidence | RESOLVED, E-060–E-063 | Every entry followed its saved predecessor exit; final qualification passed |

No unresolved blocker remains for this bounded iteration or the E-064 review. No applicable
primary-owner contradiction was found. Newly encountered physical-start,
attachment, recovery-cleanup or selector consumers would require their own
bounded discovery and qualification; this completion does not claim an exhaustive
suite audit or justify preserving incidental test behavior.
RB-001–RB-007 remain historical and retain their recorded completion evidence.

#### 12. Binary Completion Criteria

| Outcome | Current statement |
| --- | --- |
| Planning completeness | **complete** for the bounded object-store iteration: owners, structural/correction separation, assertion retirement/successor decisions, sequencing, validation and exclusions are specified; not an exhaustive suite audit |
| Per-slice readiness | S-020 **complete**, E-058; S-021 **complete**, E-059; S-025 **complete**, E-060; S-022 **complete**, E-061; S-023 **complete**, E-062; S-024 **complete**, E-063 |
| Post-completion review | **complete**, E-064; F-034–F-037 corrected and qualified, T-026 DONE; no new framework, row, schema or compatibility layer |
| Implementation authorization | **granted by the current implementation request** for the bounded slices and required validation |
| Session write-boundary verification | **Verified**, E-064: this review changed five existing S3 source/test files and the handoff; the original 20-entry implementation delta and 7,654 current entries remain. Final source frozen; all prior logs preserved; HEAD/staged entries/path identity unchanged; raw-index and ignored-tree limits disclosed |
| Implementation / test execution | V-025/V-026/V-031 passed at E-058–E-060; V-027 passed at E-061; V-028 passed at E-062; V-029 passed at E-063; follow-up V-033 passed at E-064 on its corrected source with seven selected test rows and required static/evidence/documentation checks |

The preceding document-only acceptance remains historical; E-063 supplies the
current implementation and final documentation evidence.

| Historical document acceptance group | Result / evidence at that save |
| --- | --- |
| Target/output admission and current baseline | E-053 and pre-write recheck establish the existing repository target, same source commit, dirty document baseline and safe sole-file destination |
| Evidence-backed bounded inspection | Section 2 distinguishes five fully inspected core files, partial caller boundaries, metadata-only entries and excluded ignored trees; E-054 revalidates source posture |
| Required guarantees versus incidental constraints | C-017–C-021 and F-027–F-033 identify owner obligations, meaningful successors and explicit retirement reasons; F-033 excludes arbitrary errors as absence proof |
| Traceability and executable handoff | All 30 baseline tests mapped; WF-20 → WF-21 → WF-25 → WF-22 → WF-23 → WF-24 has explicit gates and rollback; S-025 correction is separate from structural movement |
| Honest verification | Met, E-055: V-032 documentation checks passed with configured lint coverage limits disclosed; E-054 discovery is not passing test evidence; V-030 remains historical |
| Durable history | Prior checkpoints, IDs and assertion dispositions retained; no historical result promoted to current evidence; preceding WF-33 remains open |
| Sole-file preservation | Met, E-056: only this handoff changed; content/path identities/HEAD/index compared with E-053; prior historical log preserved byte-for-byte |

The following implementation criteria are met at E-059–E-063:

- Physical ownership and borrowing have separate types/lifetimes; one owned-start
  API remains; no old wrapper, StopShared alias or borrowed physical shutdown.
- Private dependencies replace mutable test hooks. The separately qualified S-025
  correction binds attachment/dependency/event inputs, rejects stale-binding reuse,
  retains terminal failure and supports local waiter cancellation without destroying
  the owner or changing another caller's admission. Recovery uses saved exact proof.
- Readiness, namespace isolation, quarantine, exact cleanup, stable reset, TLS and
  safe failure publication retain meaningful negative and real-service witnesses.
- Fake failures preserve state unless explicitly modeling response loss after
  mutation; fake namespace deletion cannot erase remaining objects. Incidental
  description and request-count constraints are retired with meaningful successors.
- Deleted-object checks require typed not-found; sibling payloads survive unchanged;
  stable reset preserves bucket identity and supports a successful new mutation.
- Three coherent service-free groups and two real integration scenarios replace
  mixed selectors and redundant setup. Distinct publication, native lifecycle,
  application-process and production-adapter witnesses remain. Authored/generated/
  nested selectors agree without duplicate or compatibility routes.
- Finalization precedes a frozen compatible validation portfolio; required cleanup
  and evidence audits pass, historical failures remain and acceptance limits are
  explicit. Documentation lint alone never establishes production guarantees;
  completion has no prescribed test-count reduction quota.

### Historical checkpoint log — preserved E-001–E-064

### Post-completion review and correction complete — E-064

The user authorized review and correction after E-063. Entry remained dirty
`main` at `9e6fa4053e2de4367ce106606da02e74fafc72d5`, with the same 20 entries
from the completed iteration and no unrelated edits. The review-entry snapshot
and full handoff are under `/tmp/cartulary-s3-review-lt5j7ior/`. All prior log
text, including E-063, remains historical and is preserved verbatim.

Read root/backend/tooling instructions and the test-operations skill and its
references. `make help ALL=1` and `make task-guide OWNER=harness.browser` passed.
Reviewed all S3 implementation/test files, the 30-test assertion ledger, affected
composition/recovery/process callers, authored/nested routing, shared transport
and event-binding boundaries, and REQ-804/402/414 plus §11.4. The separate owner,
borrower, narrow fakes, three unit groups and two real scenarios remain coherent;
no additional framework, compatibility route or service abstraction is needed.

| Finding | Correction and successor assurance |
| --- | --- |
| F-034: polling cancellation lost from error chain | Retain the polling termination alongside the original probe cause. Deterministic cancellation between probes now proves the interrupted setup envelope, both causes and probe cleanup. |
| F-035: secondary cleanup masks or disappears behind readiness | End admission on uncertain cleanup. Preserve primary probe stage, normalized cause and outcome; count probe timeouts separately. Controller cases cover transient cleanup, cleanup-only timeout, primary deadline and primary cancellation, and prohibit another probe from hiding the failure. |
| F-036: incomplete negative-case coverage | The waiter observes entry into its cancellation select before cancellation. Binding cases now include runtime/lease identities and lifecycle modes. The existing real namespace scenario proves deletion can succeed while required publication fails, and the failure still reaches its caller; sibling/final cleanup assertions remain. |
| F-037: omitted fixture CA reaches process-level trust | Require the captured fixture CA before attachment or recovery readiness/network access. The managed owner already publishes it. The missing-CA case proves rejection before verification; the shared production transport retains its separate system-trust mode. |

F-034/F-035 were pre-existing controller defects missed by the earlier
consolidation's guarantee validation. F-037 completes S-025's captured-input
boundary: `cryptography.systemTLSRoots` reads process-level trust on its first
use when no explicit root is supplied. No production transport policy, timeout,
poll interval, permitted startup retry, provider, schema, row identity or caller
API signature changed. No new top-level test was necessary; the existing exact
selectors cover the additional subcases. The only source edits in this review
are `s3test.go`, `attachment.go`, `readiness_test.go`, `ownership_test.go` and
`namespace_integration_test.go` in `internal/testutil/s3test`.

All commands use E-062's pinned Go executable/PATH, `CGO_ENABLED=0`,
`GIT_OPTIONAL_LOCKS=0`, installed-only preparation and fresh explicit run IDs.
Every run ID below resolves under the exact parent
`/home/jochi/code/cartulary/.cartulary/test-results/tr01-s3-review/`.

Negative reproduction and intermediate results remain separate from acceptance:

| Command / run ID | Session / native exit | Outcome / relationship |
| --- | --- | --- |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_readiness,harness.browser.unit.object_store_ownership` / `review-negative-20261026-01` | 19710 / Make 2 | 3,263 ms; readiness row/unit exit 10, `product/test_assertion_failure`; ownership row passed; all 10 cleanup outcomes completed. Added controller negatives expose F-034/F-035. |
| `make format` / `review-format-20261026-01` | 1279 / 0 | 5,459 ms; first correction formatted. |
| Six-row selection below / `review-focused-20261026-01` | 38912 / 0 | 12,790 ms; precedes trust correction, so checkpoint evidence only. |
| `make agent-finalize` / `review-finalize-20261026-01` | 68069 / 0 | 23,411 ms; superseded after the later trust finding. |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_ownership` / `review-trust-negative-20261026-01` | 64584 / Make 2 | 1,834 ms; row/unit exit 10, `product/test_assertion_failure`; all 10 cleanup outcomes completed. Missing-CA witness fails before F-037 correction. |
| `make format` / `review-format-20261026-02` | 66283 / 0 | 5,378 ms; final correction formatted. |

The negative roots' structured summaries and row/unit outcomes were inspected;
`make explain-run` confirmed the first failure. Its retained stdout was empty,
so individual subcase-log completeness is not claimed. Failures are preserved,
not overwritten by later passes.

The final focused command is:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_ownership,harness.browser.unit.object_store_readiness,harness.browser.unit.object_store_namespaces,harness.browser.integration.object_store_namespace_lifecycle,harness.browser.integration.object_store_stable_reset,harness.browser.unit.testservices_lifecycle_contract
```

The finalizer then passed schema shape, tier coverage and generated refresh,
with zero file changes. Both retained canonical-evidence and scheduler-drift
maintenance were skipped as `results-dir-not-provided`: RESULTS_DIR was explicitly
unset because no compatible successful full warm-check exists. Its summary is
`review-finalize-20261026-02/unit-artifacts/finalize-summary.json`. Source froze at
`sha256:d7b53f17a7e34c2ec60fbdb5f8e64e484deefa482346c352c4cb6ddafa0b8be7`;
`/tmp/cartulary-s3-review-lt5j7ior/frozen-source.json` retains non-Markdown content
hashes. The focused result remained compatible, so it was not repeated.

`make target-plan TARGET=test-slice
INPUTS=/tmp/cartulary-s3-review-lt5j7ior/final-selection.json DETAIL=rows JSON=1`
passed (session 23741, exit 0), confirming the native borrower row's existing
self-owned integration profile without adding broker dependencies.

| Accepted command / run ID | Session / native exit | Result / completed cleanup outcomes |
| --- | --- | --- |
| Focused six-row selection / `review-focused-20261026-02` | 60410 / 0 | Pass, 10,828 ms / 16 |
| `make agent-finalize` / `review-finalize-20261026-02` | 29784 / 0 | Pass, 22,885 ms / 10 |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_borrowed_fixture_lifecycle` / `review-native-20261026-01` | 88929 / 0 | Pass, 13,801 ms / 12 |
| `make lint CHECK=go` / `review-lint-20261026-01` | 21232 / 0 | Pass, 9,889 ms / 15 |
| `make source-check CHECK=generated` / `review-generated-20261026-01` | 54426 / 0 | Pass, 18,022 ms / 14 |
| `make source-check CHECK=catalog` / `review-catalog-20261026-01` | 69799 / 0 | Pass, 3,123 ms / 11 |

All six accepted roots share the frozen source digest and passed
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, each with
native exit 0, full_validity=1 and read_only=1. Inspected `run-manifest.json`,
`run-summary.json`, row results and `cleanup-results.json` show zero failed,
skipped or cancelled graph units, seven passing selected test rows plus two
generated-policy rows, and all 78 cleanup outcomes completed. Execution-result
cache hits are zero. Optional execution/resource observations remain unavailable
or not observed; no performance or resource claim is made.

This portfolio qualifies the affected readiness, attachment/trust, package and
namespace boundaries, publication and borrower lifetime. Schema/tier checks also
passed inside finalization. No routing/schema input changed, so separate identity,
full routing, application-process, production-adapter, full-owner, CI or release
runs were not repeated. E-063's wider portfolio remains historical on its own
source. This focused review does not promote it to the corrected source.

`make lint-markdown` passed as `review-markdown-20261026-01` (session 62410,
native exit 0, 17,240 ms); its adhoc summary is
`adhoc/lint-markdown/tool-run-summary.json` beneath that exact root. Configured
lint globs exclude this handoff, so direct review also verified all twelve
sections, tables, local links, IDs and dependencies, the 30 baseline assertion
dispositions, and the review's failure/successor mapping. This documentation run
is separate from the six graph qualification roots. `git --no-optional-locks
diff --check` passed.

Final comparison with the review-entry snapshot finds exactly six modified files:
the five S3 files listed above and this handoff, with no added/removed paths.
No source content changed after finalization/freeze. The original 20-entry
implementation delta and 7,654 tracked/nonignored files remain; HEAD and staged
entries are unchanged, with no staged diff. The handoff remains the same regular,
singly linked inode, and all prior section-10 log text from E-063 onward remains
byte-for-byte preserved. Raw Git-index byte identity and exhaustive ignored-tree
preservation are not claimed. Final direct checks were repeated after saving
completion status.

F-034–F-037, V-033 and T-026 are complete. No unresolved finding remains from this
bounded review. Acceptance is the focused portfolio above, with retained-run
maintenance skips and broader historical evidence kept explicitly separate.

### WF-24 exit / object-store iteration complete — E-063

`make agent-finalize` completed as `wf24-finalize-20261026-01` (session 56369,
native exit 0, 23,045 ms). Its `unit-artifacts/finalize-summary.json` reports
schema shape, tier coverage and generated refresh passed. Retained canonical
evidence and scheduler drift maintenance were skipped with
`results-dir-not-provided`: RESULTS_DIR was explicitly unset because no compatible
full warm-check root exists. These skips do not establish release/whole-owner
acceptance. Comparison against `pre-finalize.json` found no changed source files.

Source is frozen at dirty `main`, commit
`9e6fa4053e2de4367ce106606da02e74fafc72d5`, digest
`sha256:17120726ceef8adf9447044354b3fecf807b8a99432e2800dc921794e1759603`.
The non-Markdown content snapshot is
`/tmp/cartulary-s3-implementation-p5bpj7gf/frozen-source.json`.
All final qualification uses the common E-057 environment plus E-062's pinned
PATH. No timeout, retry, capacity or fixture policy was widened.

Public `make target-plan TARGET=test-slice INPUTS=<file> DETAIL=rows JSON=1`
passed for `/tmp/cartulary-s3-implementation-p5bpj7gf/final-{a,b,c,d}.json`
(sessions 55361/12573/33351/52064, all native exit 0). Each corresponding
`.plan.json` is preserved externally. A selects the five E-062 S3 rows plus
`harness.browser.unit.testservices_lifecycle_contract`; B selects physical and
borrowed native lifecycle; C and D retain the exact Section 8 selections. The
plans confirm three service-free S3 groups and two object-store integration
scenarios. B's native selected-service row is supplied by the already compatible
`wf23-selected-20261026-01` root, preserving its distinct nested selection witness
without repeating it. The independent routing-contract row, Go lint and backend
boundary also retain compatible E-062 roots. Earlier differing source digests
supply checkpoint evidence only. This portfolio is bounded, not full-owner CI.

| Final invocation / run ID | Session / native exit | State |
| --- | --- | --- |
| A / `wf24-a-20261026-01` | 54578 / 0 | Pass, 9,448 ms; 16 completed cleanup outcomes |
| B physical/borrowed / `wf24-b-20261026-01` | 10774 / 0 | Pass, 138,875 ms; 21 completed cleanup outcomes |
| `make source-check CHECK=generated` / `wf24-generated-20261026-01` | 42874 / 0 | Pass, 14,844 ms; 14 completed cleanup outcomes |
| `make source-check CHECK=schema` / `wf24-schema-20261026-01` | 88570 / 0 | Pass, 6,930 ms; 12 completed cleanup outcomes |
| `make source-check CHECK=catalog` / `wf24-catalog-20261026-01` | 27551 / 0 | Pass, 2,084 ms; 11 completed cleanup outcomes |
| `make source-check CHECK=identity` / `wf24-identity-20261026-01` | 11120 / 0 | Pass, 1,124 ms; 12 completed cleanup outcomes |
| C / `wf24-c-20261026-01` | 38717 / 0 | Pass, 73,974 ms; 21 completed cleanup outcomes |
| D / `wf24-d-20261026-01` | 92614 / 0 | Pass, 8,178 ms; 14 completed cleanup outcomes |

Every run ID in this checkpoint and the four retained E-062 roots resolves under
`/home/jochi/code/cartulary/.cartulary/test-results/tr01-s3-implementation/`.
Their exact `run-manifest.json`, `run-summary.json`, `rows/*.json` and
`cleanup-results.json` are the retained evidence. The thirteen accepted graph
roots (four compatible E-062 roots, finalizer, A–D and four static checks) passed
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, each with
native exit 0, full_validity=1 and read_only=1. All thirteen source digests match
the frozen source. Earlier static roots preceded final layout normalization and
were requalified, not silently treated as compatible.

All accepted summaries report pass with zero failed, skipped or cancelled units.
The six test-slice roots cover 13 distinct selected rows: A's six, B's two,
C's two, D's one, native selected-service and routing-contract. The generated
check contributes two additional passing policy rows and backend-boundary lint
one, for 16 distinct retained row results overall. All 179 recorded cleanup
outcomes completed. Finalizer's two internal retained-maintenance skips remain
explicit above; they are not hidden by the graph-unit counts. Optional execution
and resource observations are unavailable or not observed, so no performance or
resource claim is made.

Final A and B commands replace only their retired baseline selections:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_ownership,harness.browser.unit.object_store_readiness,harness.browser.unit.object_store_namespaces,harness.browser.integration.object_store_namespace_lifecycle,harness.browser.integration.object_store_stable_reset,harness.browser.unit.testservices_lifecycle_contract
make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.native_borrowed_fixture_lifecycle
```

C and D are exactly Section 8's commands. The compatible nested-selection root
used `make test-slice OWNER=harness.browser
ROWS=harness.browser.integration.native_selected_service_lifecycle`; the additional
routing witness used `make test-slice OWNER=harness.test_catalog
ROWS=harness.test_catalog.behavior.routing_contract`. These are independent
qualification boundaries, not duplicate S3 selectors.

Comparison with the implementation-entry snapshot finds exactly 20 changed,
added or removed entries: this handoff; eleven entries in `internal/testutil/s3test`
(including seven additions and two removals); the process teardown and native
selection consumers; testservices composition and its tests; authored browser
family, historical row migration and tier-selection assertions; and the
Make-generated topology render index. The 7,654 remaining tracked/nonignored
files include no unrelated changes. Non-Markdown content exactly matches
`frozen-source.json`; HEAD and staged entries are unchanged, with no staged diff.
Raw index metadata changed, so raw byte identity is not claimed. The handoff
remains a regular singly linked file at inode 348826, and all prior log text from
E-053–E-056 through the historical section-10 tail is byte-for-byte preserved.
Ignored runtime/cache/build trees were not exhaustively fingerprinted.

`make lint-markdown` passed as `wf24-markdown-20261026-01` (session 67753,
native exit 0, 17,470 ms). This is a separate adhoc documentation invocation,
not an additional graph qualification root; its authoritative summary is
`adhoc/lint-markdown/tool-run-summary.json` beneath that exact run root.
Configured lint globs exclude this handoff. Direct review therefore checked the
twelve sections, consistent tables, resolvable local links, retained IDs and
ordered gates, the 30-test assertion ledger and its successor coverage, and
agreement among completion/readiness/validation records. `git --no-optional-locks
diff --check` passed. Source preservation and history checks were repeated after
the final document save; only this handoff changed after source freeze.

WF-20/WF-21/WF-25/WF-22/WF-23/WF-24, S-020–S-025 and
T-019/T-020/T-025/T-021/T-022/T-023 are complete. F-027–F-033 are resolved within
this scope. All required production guarantees have meaningful successor
witnesses; acceptance covers the bounded object-store refactor, with no full-suite,
release, performance or preceding WF-33 closure claim. No implementation work
remains in this iteration.

### WF-23 exit / WF-24 entry — E-062

Authored routing now has exactly these S3 groups, all retaining
`harness.browser.verification.support_contract` because its meaning is unchanged:

| Current row | Selection / execution declaration |
| --- | --- |
| `harness.browser.unit.object_store_ownership` | 10 ownership/attachment/configuration tests; service-free, capability `none` |
| `harness.browser.unit.object_store_readiness` | 7 mutation-readiness/diagnostic tests; service-free, capability `none` |
| `harness.browser.unit.object_store_namespaces` | 9 namespace/package/cleanup tests; service-free, capability `none` |
| `harness.browser.integration.object_store_namespace_lifecycle` | `TestObjectStoreNamespaceLifecycle`; integration, capability `object_store_namespace`, dependency `object_store` |
| `harness.browser.integration.object_store_stable_reset` | `TestResetBucketPreservesNamespaceAndProvesMutation`; integration, same capability/dependency |

The native selected-service tests and scheduler tier assertions now name these
current rows. The existing historical migration record points directly to current
successors; no executable alias or new compatibility mechanism was introduced.
Source/catalog inspection finds no old row IDs in executable selectors. The
standalone publication, native, application and production-adapter rows remain.
A review also kept owner-only CORS inputs out of borrowed binding equality: owner
startup reads them once, but changing an irrelevant ambient CORS variable cannot
invalidate an otherwise identical borrower. Imports/field layout were normalized
for readability; no new policy or source owner was introduced.

Common environment and result parent remain E-057. Current validation:

| Command / run ID | Native session / exit | Result |
| --- | --- | --- |
| `make generate` / `wf23-generate-20261026-01` | 29809 / 0 | Pass, 9,980 ms; generated topology render index refreshed |
| `make source-check CHECK=generated` / `wf23-generated-20261026-01` | 83477 / 0 | Pass, 13,627 ms; precedes final layout cleanup and remains checkpoint evidence |
| `make source-check CHECK=schema` / `wf23-schema-20261026-01` | 19113 / 0 | Pass, 6,122 ms |
| `make source-check CHECK=catalog` / `wf23-catalog-20261026-01` | 99254 / 0 | Pass, 1,820 ms |
| `make source-check CHECK=identity` / `wf23-identity-20261026-01` | 53867 / 0 | Pass, 1,036 ms |

`make format` passed as `wf23-format-20261026-01` (session 35691, exit 0,
5,422 ms). `make lint CHECK=go` initially failed as
`wf23-lint-go-20261026-01` (session 85353, Make exit 2; Staticcheck unit exit 1,
`product/test_assertion_failure` normalization). Formatting/vet passed; bounded
Staticcheck stderr identified its internally launched `go` as 1.27.1 with
`GOTOOLCHAIN=local`. It uses PATH independently of Make's GO variable.

Subsequent commands additionally prepend
`/home/jochi/.cache/cartulary/go/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin`
to PATH. No pin, tool or verification rule changed. Corrected Go lint passed as
`wf23-lint-go-20261026-02` (session 66254, exit 0, 7,818 ms).
`make lint CHECK=backend-boundary` passed as `wf23-boundary-20261026-01`
(session 89717, exit 0, 2,890 ms). The exact routing contract passed as `wf23-routing-20261026-01`
(session 84190, exit 0, 98,154 ms; all 11 cleanup outcomes completed).
The native selected-service witness passed as `wf23-selected-20261026-01`
(session 36224, exit 0, 39,425 ms). Go lint's 15 and boundary lint's 12 cleanup
outcomes completed; all these graph summaries have no failed/skipped/cancelled
units. V-028/WF-23/S-023/T-022 are DONE. Earlier lint failure remains historical.

The post-format lint and routing roots share source digest
`sha256:17120726ceef8adf9447044354b3fecf807b8a99432e2800dc921794e1759603`.
WF-24 enters: run `make agent-finalize` with RESULTS_DIR unset (no compatible
successful full warm-check exists), inspect its writes and freeze source. Reuse
focused roots only if compatibility survives; run missing A–D coverage and static
checks, then audit every accepted graph root. Earlier focused results with a
different digest are checkpoint evidence only. This exit was saved before finalization.

### WF-22 exit / WF-23 entry — E-061

Unit assertions now reside in `ownership_test.go`, `readiness_test.go` and
`namespace_test.go`; real scenarios reside in `namespace_integration_test.go`.
Package quarantine moved with namespace lifecycle. Readiness and cleanup retain
separate narrow fakes. No mutable package-global substitutions remain.

The readiness fake leaves state unchanged on ordinary failure, explicitly marks
namespace-delete response loss after mutation, and rejects nonempty namespace
deletion. The old false no-residue assertion is replaced by actual retained-state
and primary/secondary failure assertions. A private per-fixture poll wait lets
unit cases coordinate recovery/expiry while production keeps its existing ticker
and deadlines. Testservices' independent publication witness uses an already
expired deadline instead of sleeping for a nanosecond timer.

The startup-description assertion retires to mapped-endpoint-before-readiness
coverage plus `TestObjectStoreStartupMappingDeadline`. Cleanup request totals
retire to bounded batches, remaining-state inspection, fresh empty observation
and rejection before requesting beyond a repeated token. Required one-time
admission/settlement/quarantine effects and causal mutation order remain.

Three real tests (`TestHarnessStartsSeaweedFSS3AndRoundTripsObjects`,
`TestCleanupPrefixPreservesSiblingObjects`,
`TestPinnedObjectStoreBatchCleanupTraversesMultiplePages`) now have one successor,
`TestObjectStoreNamespaceLifecycle`: exact roundtrip, 1,007-key pagination,
bounded batches, independent empty listing, typed deleted-key absence, unchanged
sibling bytes, saved-binding recovery despite ambient conflict, verified bucket
removal and repeated final cleanup. The independent stable-reset case now requires
typed absence and exact new payload. Transport/authentication errors cannot prove
absence; negative readiness cases explicitly reject them.

The existing integration row selects the new scenario and the existing unit row
selects the renamed deadline test. Authored selectors changed with source;
retired names have no aliases. Final semantic row consolidation remains WF-23.
`make format` passed as `wf22-format-20261026-01` (session 7223, exit 0,
5,541 ms). `make generate` passed as `wf22-generate-20261026-01`
(session 75834, exit 0, 10,420 ms), updating the generated topology render index.
A with the migrated source selectors passed as `wf22-a-20261026-01` (session
60455, exit 0, 11,945 ms): six selected rows, zero failed/skipped/cancelled units,
all 17 cleanup outcomes completed. V-027 and WF-22/S-022/T-021 are DONE.

WF-23 entry inspection additionally found exact references in
`tools/harness/scheduler/work-graph/tests/tier-selection.mjs` and the existing
object-store entry in `tools/test_catalog_row_migrations.json`. These are scoped
routing consumers, not scheduler architecture changes. The migration loader
validates retired/current identities; it does not execute retired aliases. Update
that existing record's successors, all tier/nested selectors, and the authored
family atomically. Add the existing exact
`harness.test_catalog.behavior.routing_contract` row to qualification for the
changed tier assertion boundary; no new harness or policy is needed.
This checkpoint was saved before routing consolidation.

### WF-25 exit / WF-22 entry — E-060

Before correction, new negative subcases were added under the existing
`TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness` selector. The exact
service-free row was run as `wf25-negative-20261026-01` (native session 69868,
Make exit 2, row/unit exit 10). Its run summary reports
`product/test_assertion_failure`; a bounded stdout inspection confirmed changed
bindings reused cached admission. This is an expected defect witness, not a
passing baseline. The terminal row/summary remain authoritative; the subsequently
read stdout file was empty, so no complete retained-log claim is made.

The correction captures a copied allowlist of attachment, dependency, TLS,
runtime and event-destination inputs. One private admission instance retains its
binding and terminal result. Reuse compares the full binding; pending callers
wait on completion with their own context, outside the admission lock. No new
registry or replacement generation exists. Fixture guards, names and journals
use saved inputs; event publication errors propagate instead of reporting success.
Physical startup captures its runtime/CORS inputs once as well.

Recovery now calls `CleanupBucketFromBinding` with the supplied saved environment
only, after testservices' generated-name ownership proof. It cannot fall back to
StartShared or perform fresh admission/provisioning. The saved-binding borrower
closes its local transport. Missing/incomplete bindings fail before network use.
New cases cover each required attachment field, malformed/omitted dependencies,
independent copied bindings, changed credentials/dependencies/destinations,
terminal readiness/publication failure and local waiter cancellation. Existing
selectors include these subcases; no compatibility routing is added.

All roots use E-057's parent and common pinned environment.

| Invocation / run ID | Session / native exit | Result |
| --- | --- | --- |
| Negative service-free row / `wf25-negative-20261026-01` | 69868 / 2 | Expected assertion failure before correction, summary 2,843 ms |
| `make format` / `wf25-format-20261026-01` | 79163 / 0 | Pass, 5,583 ms |
| A / `wf25-a-20261026-01` | 66040 / 0 | Pass, 12,460 ms; six rows including all new negative witnesses and distinct testservices publication |
| Native borrowed lifecycle / `wf25-b-20261026-01` | 2190 / 0 | Pass, 10,207 ms; all 12 cleanup outcomes completed |

C (both original rows) passed as `wf25-c-20261026-01`, session 53694, native
exit 0, summary 76,943 ms; all 21 cleanup outcomes completed. A's 17 cleanup
outcomes completed. A/B/C summaries contain zero failed, cancelled or skipped
units. V-031 is qualified; WF-25/S-025/T-025 are DONE. The expected pre-correction
failure is retained. S-022 may now consolidate the qualified negative cases with
the other assertion groups. This checkpoint was saved before WF-22 edits.

### WF-21 exit / WF-25 entry — E-059

The structural extraction introduces `owned_store.go` (`OwnedStore`, one
`StartOwned(ctx, StartOptions)` constructor, owner-typed settlement callbacks,
physical/TLS close and broker-probe construction), `attachment.go` (one process
admission instance; independent instances in tests), and `operations.go` (private
fixture dependencies). `Harness` cannot terminate containers or access proxy keys;
its close releases local transport and retains package quarantine. No embedding
or compatibility aliases remain. Testservices uses a local start/probe seam;
process TestMain closes its borrowed transport. The existing 30 selector names
remain valid; no routing migration is needed for this extraction.

New subcases under the existing ownership/configuration selector qualify
before-acquisition failure, acquisition cancellation, publication failure,
before→acquire→publish→start→ready ordering, TLS confinement/removal, memoized
owner close and borrower nontermination/quarantine retention. Existing partial
cleanup uncertainty and readiness terminal/retry assertions remain. Admission
still has the old caching/ambient semantics deliberately; WF-25 owns that correction.

All roots use E-057's result parent and common pinned environment.

| Invocation / run ID | Session / native exit | Result |
| --- | --- | --- |
| `make format` / `wf21-format-20261026-01` | 54979 / 2 | Prerequisite Go build failed before formatting: method expression had receiver/context in the wrong order. Step summary and bounded stderr at `harness-runtime-build/build-harness-runtime/` retain the failure; corrected with a local function |
| `make format` / `wf21-format-20261026-02` | 67981 / 0 | Pass, 5,433 ms; only scoped authored Go formatting |
| A / `wf21-a-20261026-01` | 59326 / 0 | Pass, 11,486 ms; all six rows including new ownership subcases and publication witness |
| B physical/borrowed rows / `wf21-b-20261026-01` | 91260 / 0 | Pass, 136,258 ms; both rows, all 21 cleanup outcomes completed. Nested selected-service qualification remains a later routing/final obligation |
| C process row only / `wf21-c-20261026-01` | 24453 / 0 | Pass, 76,561 ms; all 20 cleanup outcomes completed |
| D / `wf21-d-20261026-01` | 77436 / 0 | Pass, 8,078 ms; all 14 cleanup outcomes completed |

A's 17 cleanup outcomes completed. All four test summaries report zero failed,
skipped or cancelled units. The scoped selectors remain unchanged; no generated
routing edits were needed. `git diff --check` passed. WF-21/S-021/T-020 are DONE;
RB-010 is resolved. WF-25 enters with separate negative witnesses before any
admission correction. This exit was saved before those witnesses were added.

### WF-20 exit / WF-21 entry — E-058

All four V-025 baseline selections passed on the unchanged implementation. Commands
use the exact Section 8 rows and E-057 common environment, including the explicit
cached pinned GO executable. Each exact root is
`/home/jochi/code/cartulary/.cartulary/test-results/tr01-s3-implementation/<run-id>`.

| Run ID / group | Native session / exit | Selected rows / summary wall | Cleanup |
| --- | --- | --- | --- |
| `wf20-a-baseline-20261026-02` / A | 10144 / 0 | 6 / 10,354 ms | 17 completed |
| `wf20-b-baseline-20261026-01` / B | 93217 / 0 | 3 / 145,219 ms | 22 completed |
| `wf20-c-baseline-20261026-01` / C | 23974 / 0 | 2 / 73,003 ms | 21 completed |
| `wf20-d-baseline-20261026-01` / D | 51528 / 0 | 1 / 7,399 ms | 14 completed |

All summaries report pass with zero failed, skipped or cancelled units; selected
rows are distinct from setup/work-unit counts. The first A attempt remains a
pre-test toolchain invocation failure (E-057), not a deleted or passing result.
The ledger covers every baseline test/assertion family and affected caller;
known defects and missing negative witnesses remain explicit. No historical
PostgreSQL evidence was promoted. WF-20/S-020/T-019 are DONE and RB-008 is resolved.

WF-21 now enters. Its exit must qualify physical-versus-borrowed ownership,
partial settlement and local dependencies before WF-25 changes admission.
There is no S-025 behavior claim at this checkpoint. This checkpoint was saved
before source movement.

### Object-store implementation entry — E-057

2026-10-26, America/New_York. Current user authorization supersedes the earlier
planning-only restriction. WF-20 is in progress; no successor slice has entered.
The implementation entry baseline is `main` at
`9e6fa4053e2de4367ce106606da02e74fafc72d5`, with pre-existing changes confined to
this handoff. All 7,649 tracked/nonignored entries, HEAD/index and the original
handoff were saved in `/tmp/cartulary-s3-implementation-p5bpj7gf/baseline.json`
and `original-handoff.md`. The E-053/E-056 baseline table above describes the
prior document save, not this invocation's original handoff bytes. No staging or
commit is part of this work. Historical checkpoints below remain unchanged.

Public `make help ALL=1` and task guides for `harness.browser`, `app.server` and
`platform.objectstore` succeeded. The 30-test ledger is supplemented above at
assertion/subcase and consequential-caller level. F-028/F-029/F-030/F-032/F-033
remain known defects; baseline passes do not adopt their incidental behavior.

V-025 uses the exact A–D selections in Section 8. Common command environment:
`CGO_ENABLED=0 GIT_OPTIONAL_LOCKS=0 CARTULARY_PREPARATION_POLICY=installed_only
CARTULARY_TEST_RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr01-s3-implementation`.
Each invocation supplies the run ID below. After the first failure, it also
supplies `GO=/home/jochi/.cache/cartulary/go/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go`.

| Run ID / group | Native session / exit | Result and cleanup |
| --- | --- | --- |
| `wf20-a-baseline-20261026-01` / A | 9411 / 2 | Before tests: cryptographic module toolchain identity rejected. No run root published. Installed launcher VERSION is go1.27.1; required cached executable VERSION is go1.27.3. No pin or verification bypass; corrected explicit GO for the new invocation |
| `wf20-a-baseline-20261026-02` / A | 10144 / 0 | Pass, 6 selected rows; summary wall 10,354 ms; all 17 published cleanup outcomes completed, no failed/skipped units. Root is the result parent above plus this run ID |
| `wf20-b-baseline-20261026-01` / B | 93217 / pending | Native lifecycle validation running; not yet credited |

WF-20 exit is not yet saved. Source refactoring remains gated on the complete
V-025 baseline; no source changes or generation have occurred at this entry.

### Object-store plan refinement — E-053–E-056

2026-10-26, America/New_York. The user selected refinement of the unimplemented
object-store iteration, chose to retain the attachment defect as a separate
correction, and authorized saving this document-only plan. Only this handoff may
change. Source refactoring, tests, generation, services and normative amendments
remain future work. Entries below this refinement record retain their original
historical status, authorization and evidence limits; they are not current results.

#### Refinement decisions and source posture

- E-053 establishes the actual starting worktree: the same `main` commit, clean
  staged entries and pre-existing edits only in this handoff. The previous
  planning pass left its fingerprint unchanged. Do not replace those edits with
  the committed version or reuse the earlier clean-worktree claim.
- E-054 revalidates the five authored S3 files (3,076 lines; 30 top-level tests),
  affected composition/consumer boundaries and routing against the unchanged
  source commit. The focused inventory remains bounded, with partial caller
  inspection and metadata-only remaining entries; no exhaustive suite audit.
- F-033 records weak absence assertions in the prefix-cleanup and stable-reset
  tests. Successors require typed not-found results, surviving sibling bytes and
  successful post-reset mutation; unavailable transport cannot prove deletion.
- S-021/WF-21 now contains structural movement only. S-025/WF-25/V-031 corrects
  bound admission after that checkpoint and before S-022/WF-22 consolidation.
  The owner clauses are REQ-804/410; no specification amendment is proposed.
- F-029 now explicitly rejects implicit recursive fake bucket deletion. Default
  failure preserves state; response loss after mutation is an explicit case.
- The 30-test ledger and guarantee/retirement table distinguish required assurance
  from historical strings, request counts, globals, names and redundant setup.
  Existing verification identities remain when their meaning is unchanged.
- No frontend presentation, production API/data schema, PostgreSQL internals,
  scheduler, provider or performance-policy change is part of this iteration.

#### Refinement evidence register

| Evidence | Source / baseline | Result and applicability |
| --- | --- | --- |
| E-053 | `/tmp/cartulary-s3-plan-refinement-rv0caa_o/baseline.json` and `original-handoff.md` | Fingerprints and file identities for 7,649 tracked/nonignored entries; HEAD, raw index and staged-entry hashes captured before editing; prior document edits retained |
| E-054 | Preceding refinement inspection; current unchanged source commit, applicable root/internal/tools guidance and owner clauses | F-027–F-033, C-017–C-021, separate correction and successor coverage plan; read-only Make help and task-guide discovery succeeded for harness.browser, platform.objectstore and app.server; no tests ran |
| E-055 | V-032 documentation validation; exact command and artifacts below | Outer lint exit 0, configured lint passed; direct document review and whitespace check passed. Configured lint excludes this handoff; no product assurance claim |
| E-056 | Pre-write identity/content check; `/tmp/cartulary-s3-plan-refinement-rv0caa_o/comparison.json` | Only this handoff changed across 7,649 entries; no added/deleted paths or identity changes; HEAD/branch/raw index/staged entries unchanged. Historical log preserved byte-for-byte; ignored-tree limits disclosed |

#### Refinement documentation validation

Executed from `/home/jochi/code/cartulary`, with installed prerequisites,
fixing disabled, external result/cache destinations and network-disabled Corepack:

```sh
GIT_OPTIONAL_LOCKS=0 \
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-s3-plan-refinement-rv0caa_o/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-s3-plan-refinement-rv0caa_o/static-analysis \
CARTULARY_TEST_RUN_ID=tr01-object-store-plan-refinement-20261026-01 \
COREPACK_ENABLE_NETWORK=0 HARNESS_DIAGNOSTICS=off HARNESS_ACCOUNTING=off \
make lint-markdown
```

V-032 passed. The native outer Make exit was **0**; the summary reports **pass**,
zero failures, zero tests and 17,200 ms. Exact summary:
`/tmp/cartulary-s3-plan-refinement-rv0caa_o/results/tr01-object-store-plan-refinement-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
Configured lint excludes this handoff, so that pass applies only to its configured
documents. Supplemental review in `document-review.json` beside the baseline
passed: twelve top-level sections, 17 current tables, eight current local links,
stable IDs and finding traceability, the six-workflow dependency chain, all 30
baseline test mappings and four commands selecting 12 existing catalog rows.
The prior 168,640-byte session log is preserved byte-for-byte. These are document
checks, not executed test successors or product verification.

`git --no-optional-locks diff --check -- docs/handoffs/testing-remediation-01.md`
passed. The E-056 comparison found only this handoff changed across all 7,649
tracked/nonignored entries; no added/deleted paths or file-identity changes, and
unchanged HEAD, branch, raw index and staged entries. Final evidence annotations
were directly reviewed and preservation was rechecked; the linted population did
not change. Ignored build/dependency/output trees were not exhaustively audited;
reviewed command paths used installed prerequisites and external result/cache
outputs. The initial uncommitted plan and historical rationale remain preserved
within the refined document and its retained starting copy.

No format, generation, agent-finalize, product tests, service commands or
installation ran. Those operations are excluded by this document-only scope.
T-024 is DONE; V-025–V-029 and V-031 remain unexecuted future obligations.
The next source-work entry remains later-authorized WF-20 characterization.


### Object-store planning save — E-048–E-052

2026-10-26, America/New_York. The user approved implementing the document-update
plan, explicitly keeping source changes and tests as future work. Revalidation
found the same clean source as the planning turn. The sole admitted write is this
handoff; no instructions embedded in its historical entries expand that boundary.

#### Scope and authority

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-048/E-050 | Existing canonical repository target/output admitted; current owner allocation and instructions revalidated; no normative amendment | Save plan only; later actor revalidates source and obtains implementation authorization |

#### Backend boundary

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-049 | Mixed S3 ownership, globals, ambient binding and misleading fake cleanup identified; all five core files inspected | WF-20 qualifies guarantees before owner split; preserve narrow S3 transport/cleanup owners |

#### Frontend boundary

Not applicable to source changes: browser presentation and UI selectors remain
outside scope. Native fixture/selection boundaries are harness witnesses, not a
request for rendered UI inspection.

#### Contract and codegen

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-049/E-050 | Authored rows and nested selectors identified; real reset currently classed unit; no generated edits | WF-23 updates authored inputs and Make projections after successor qualification |

#### Tests and harness

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-049/E-050 | Thirty baseline tests have planned dispositions; A–D exact starting commands discovered; no tests executed | WF-20 completes subcase/caller ledger and current baseline; preserve distinct physical/publication witnesses |
| E-051 | V-030 passed: outer lint exit 0; direct handoff review and diff check passed | No product test evidence is implied; retain configured lint coverage limit |

#### Security and authorization

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-049/E-050 | Preserve TLS roots/private owner keys, exact namespaces, saved recovery proof, bounded diagnostics and production fail-fast transport | Add/retain fault cases without serializing raw provider errors or adding product retries |

#### Open risks and next session

| Evidence | Current state | Next action |
| --- | --- | --- |
| RB-008/RB-009 | Planning decisions settled; current characterization and predecessor exits are future gates | Begin only later-authorized WF-20; do not resume unrelated WF-33 |
| E-052 | All 7,649 entries compared; only this handoff changed; no identity/path-set/index/HEAD change | Later actor revalidates the baseline; ignored trees were not exhaustively fingerprinted |

#### Current evidence register

| Evidence | Source / baseline | Observation and applicability |
| --- | --- | --- |
| E-048 | Clean `main` at `9e6fa4053e2de4367ce106606da02e74fafc72d5`; `/tmp/cartulary-s3-plan-doc-ohx6z1fv/baseline.json` | All 7,649 path identities/content captured; raw index/entries and original tracker retained; no current test result |
| E-049 | Five s3test files and consequential caller/dependency/routing boundaries in Section 2; unchanged planning/save source | Static findings F-027–F-032, C-017–C-021 and 30-test ledger; no exhaustive repository assertion audit |
| E-050 | Root/internal/tools instructions, Refactor Tracker format/framework, adopted owner clauses, supporting domain/NLSpec guidance, Make/task guides and lint invocation path | Authority distinguished from incidental behavior; future commands known; V-024 collision resolved by preserving it and starting at V-025 |
| E-051 | V-030 document checks; exact command and artifact below | Outer Make exit 0; configured lint passed; direct section/table/link/ledger/catalog review and diff check passed; no product test validation implied |
| E-052 | Pre-write identity/content recheck; `/tmp/cartulary-s3-plan-doc-ohx6z1fv/comparison.json` | Only this handoff content changed across 7,649 entries; HEAD, raw index, staged entries and path identities unchanged; verbatim historical checkpoint log retained |

#### Document validation

Executed from the repository root with installed prerequisites, fixing disabled,
external caches/results and network-disabled Corepack:

```sh
GIT_OPTIONAL_LOCKS=0 \
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-s3-plan-doc-ohx6z1fv/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-s3-plan-doc-ohx6z1fv/static-analysis \
CARTULARY_TEST_RUN_ID=tr01-object-store-plan-doc-20261026-01 \
COREPACK_ENABLE_NETWORK=0 HARNESS_DIAGNOSTICS=off HARNESS_ACCOUNTING=off \
make lint-markdown
```

The native outer Make exit was **0**; the lint summary reports **pass**, zero
failures and 17,280 ms. Exact summary:
`/tmp/cartulary-s3-plan-doc-ohx6z1fv/results/tr01-object-store-plan-doc-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
No test ran. The configured lint globs exclude this handoff, so this pass applies
only to their configured documents. Supplemental document review checked all
twelve top-level sections, tables/fences/whitespace, local links, ID traceability,
all 30 baseline test dispositions, four exact catalog selections and preservation
of the historical checkpoint log. Results are in `document-review.json` beside
the baseline. `git --no-optional-locks diff --check` passed. Final evidence
annotations were directly reviewed; they do not change the linted population.

The comparison in E-052 found no added/removed paths or file-identity changes;
only this handoff's content changed. HEAD, raw index and staged entries remained
identical. Ignored build/dependency/output trees were not exhaustively audited;
the reviewed command paths redirected outputs externally. No formatting,
generation, finalizer, product verification, installation or source refactor ran.
Those operations are excluded by the document-only scope. T-018 is DONE; the
next implementation step remains later-authorized WF-20 characterization.

### Completed PostgreSQL iteration — historical ID index

The following index replaces obsolete current-state summaries; it does not
reopen completed work or discard its evidence. Full original planning detail is
retained in Git at
`9e6fa4053e2de4367ce106606da02e74fafc72d5:docs/handoffs/testing-remediation-01.md`
and in the external original-handoff copy. The detailed checkpoint log beginning
with E-047 below is preserved verbatim. Any “current” or authorization wording in
that historical log refers only to the session it records.

| Historical finding / contract | Completed disposition / slices | Validation / work items / evidence |
| --- | --- | --- |
| F-015; C-011/C-015 | OwnedCluster separates physical PostgreSQL ownership from borrowed Harness; one owned-start API; S-014 | V-017/V-019; T-011; E-039/E-041 |
| F-016; C-012/C-014 | One template recipe, no ordinary migration fallback, confirmed partial-create ownership; S-013/S-017 | V-016/V-021; T-010/T-015; E-038/E-045/E-047 |
| F-017; C-011/C-013 | Local dependencies, bound generation admission and persistent quarantine; S-014/S-018 | V-017/V-022; T-011/T-016; E-039/E-046/E-047 |
| F-018; C-013/C-016 | One real policy resolver and consolidated cleanup cases; S-015/S-018 | V-018/V-022; T-012/T-016; E-040/E-046/E-047 |
| F-019; C-011/C-014 | Capability/meaningful schema assertions replace method counts, smoke-only and impossible borrowed shutdown; S-014/S-015 | V-017/V-018; T-011/T-012; E-039/E-040 |
| F-020; C-011–C-016 | Distinct real transport/reset/ACL/recovery contexts retained; bounded terminology clarified; S-012/S-016 | V-015/V-019; T-009/T-013; E-037/E-041 |
| F-021/F-024; C-012/C-013/C-015 | Context-aware template opens, cancellation-aware wait, one terminal attempt, early confirmed ownership; S-017 | V-021; T-015; E-045 |
| F-022/F-023/F-025/F-026; C-011/C-013/C-014/C-016 | Failed verified generation stays retired; admitted migration capability, immutable inputs and no redundant detach state; S-018 | V-022; T-016; E-046 |
| S-019 / closure | WF-17–WF-19 COMPLETE; frozen compatible final portfolio and handoff | V-023; T-017; E-047; RB-007 DONE |
| Historical planning/review | T-008/T-014 DONE; no implementation authority arose from those document-only saves | V-020/V-024; E-030–E-035/E-042/E-043; RB-005/RB-006 DONE through E-041 |

#### Historical PostgreSQL assertion dispositions


| Existing assertion group | Final disposition / meaningful successor | Retirement gate |
| --- | --- | --- |
| `TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness`, explicit-attachment and omitted-dependency cases | Attachment tests prove missing/malformed binding fails before acquisition; borrower abandonment does not terminate owner; quarantined generation cannot be reused | C-011; V-017; remove invalid mixed Harness construction and unreachable owned-shutdown branch |
| Owned labels, partial startup, readiness/cancellation/authentication failures and diagnostic helpers | OwnedCluster tests use local startup dependencies; retain callback ordering, bounded errors, exact cleanup authority and real transport qualification | C-011/C-015; V-017/V-019; remove old start wrappers only after direct caller migration |
| Duplicate local/service template sequence tests | One operation-order/fault matrix plus thin service publication-order assertion; real `TestTemplatePreparationCompletesCopyBeforePublication` remains | C-012; V-016; no duplicate migration or premature published binding |
| Standalone/attached isolated-clone cleanup and standalone/attached migration-scratch cleanup | **Qualified/replaced, E-040:** `TestDatabaseFixtureCleanup` covers clone/head scratch/targeted scratch × standalone/attached/persistent borrower, exact owned name, one eager drop and handle-before-drop order; independent context/failure authority witnesses remain | C-013/C-014; V-018; keep differences that affect actual cleanup authority |
| Policy resolver-only test plus `ExplicitPostgresFixturePolicyT` | **Qualified, E-040:** public helper uses the sole resolver; its actual attribution path is tested for top-level/package/default selection and rejected policy before acquisition | C-016; V-018; delete unused alternative resolver |
| `TestMigrationDatabaseTargetedOperationValidation` | **Qualified, E-040:** nil/zero/forged rejection; invalid targets before counted source loading or execution; legal canonical-source apply/rollback; no field/method inventory | C-014; V-018; changing private method layout must not become compatibility work |
| `TestHarnessStartsPostgresAndRunsCurrentMigrationPath` and `TestPrepareIsolatedDatabaseTReturnsMigratedDatabase` | **Qualified/replaced, E-040:** `TestPrepareIsolatedDatabaseTReturnsMigratedDatabase` retains schema and `application_crypto_format` queries; smoke function/selector removed | C-012/C-014; V-018; no weaker nonempty-DSN replacement |
| `requireProcessHarnessStartupErrorContract` | Preserve object-store attach failure as primary and perform no PostgreSQL owner operation; physical cleanup-error joins remain in OwnedCluster tests | C-011; V-017; remove synthetic borrowed cleanup error without redesigning S3 |
| Rollback isolation/uncertain settlement, real forced drop/barrier, reset, ACL, process readiness/recovery and performance snapshot lifecycle | Retain separate real contexts and failure boundaries; share setup only when ownership semantics match | C-011–C-016; V-019; no deletion based solely on similar test titles |

Organize focused unit assertions by attachment/policy, owned-cluster lifecycle,
database preparation/cleanup, and migration/transaction capabilities. Keep real
PostgreSQL lifecycle and transport qualification separate from those unit
fixtures. Filenames and top-level test names may change with their routing;
neither their present layout nor their population is an acceptance requirement.

#### Closure assertion successors — F-021–F-026

| Gap / retained guarantee | Qualified successor and exact selection | Disposition |
| --- | --- | --- |
| F-021 cancellation, deadline, one terminal preparation, independent waiters and settlement | `TestPrepareDatabaseTemplateModeClonesWithoutMigrationReplay/template_preparation_contract` in C's certificate_fixture_lifecycle; real template witness in E | Added work/guard stalls, real TCP open, cancellation and waiting deadline, one result, no retry and no premature publication; V-021/V-022 and final C/E requalification pass, E-047. |
| F-024 confirmed CREATE before dependent work; rejected targets never owned | Same template contract plus `TestTemplatePreparationCompletesCopyBeforePublication` in E | Added pre-provision local/journal checks for pristine/clone callbacks; retained primary/close/drop/release and unowned-target boundaries. |
| F-022 terminal verified admission and independent generations | `TestStartSharedRequiresExplicitAttachment` / `TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness` in A | Added readiness failure/cancellation/publication, no repeated probes, pending caller cancellation and concurrent independent generation. |
| F-023 declared migration issuance and canonical target bounds | `TestValidateSelectedPostgresFixturePolicy/migration_issuance_requires_capability` in A; `TestDatabaseFixtureCleanup` and `TestMigrationDatabaseTargetedOperationValidation` in B; recorded caller inventory | Capability/scope cross-product rejects both factories before acquisition; valid issuance, source identity and forged/zero/invalid targets remain. No new policy language or bypass. |
| F-025 stable authority, policy and evidence destination | A's attachment binding/authority/journal cases; `TestMigrationActivityPrivatePublicationAndShutdown` in timing row | Conflicting ambient values do not redirect admission, guards, policy, creation journal or timing publication. Changed cached bindings fail; one journal fault does not poison another registry. |
| F-026 borrower nontermination and initiating error | A's borrowed-handle capability/quarantine assertions; E's real owner usability; D's `requireProcessHarnessStartupErrorContract` | Remove detach field/callback assertions and API together. Existing top-level selectors now protect behavior; no retired selector or forwarding stub is needed. |

The three new row IDs are exactly those approved in the closure plan. Graph
current-head schema moves to its dedicated row; six canonical cutover tests move
together to the reference-data migration row; reporting identity cutover moves
to its migration row. Remaining ordinary tests stay dedicated. The final
portfolio includes both sides of every split. No meaningful assertion was deleted
without its qualified replacement; existing transport/reset/ACL/recovery and
performance snapshot contexts remain separate witnesses.

### Historical checkpoint log — E-001–E-047

### WF-19 completion / E-047 — final handoff

**The closure effort is complete.** WF-17, WF-18 and WF-19 each finished
validation and saved its tracker checkpoint before the next began. All six
findings, reopened F-016–F-019 obligations, T-015–T-017, V-021–V-023 and RB-007
are closed against their actual witnesses. No implementation work remains in
this bounded plan. The next action is human review of the changes and projection
fidelity; there is no new requirement amendment or pending execution gate here.

Finalization ran before the broader portfolio and changed no executable input.
All final runs below share the exact identities:

- Source: `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- System/config capability: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
- Git source: dirty `main` at `867e97ed49829d671c58576232999d1fc9a4e3fb`;
  each manifest retains its selected graph/config, command and exact declared inputs.

The freeze-entry commands immediately below are the exact executed final
portfolio. Every root below has native outer exit **0**, terminal status **pass**,
retained secret scan **pass**, and observability audit exit **0**. For each exact
root, the audit command was
`CGO_ENABLED=0 make harness-observability-check RESULTS_DIR=.cartulary/test-results/<basename> CHECK=all`;
all returned `full_validity=1`. The native command exits were captured separately
from summary status. A later read-only `make test-run-status` on the transport
root exited 0 and reported terminal pass; its observation limitations do not
replace the captured native exit or the subsequent evidence audit.

| Exact root under `/home/jochi/code/cartulary/.cartulary/test-results/` | Result / outer / audit | Completed cleanup records |
| --- | --- | --- |
| `tr01-wf19-finalize-20261026-01` | pass / 0 / 0 | 10 |
| `tr01-wf19-lifecycle-20261026-01` | pass / 0 / 0 | 30 |
| `tr01-wf19-migrations-20261026-01` | pass / 0 / 0 | 21 |
| `tr01-wf19-postgres-20261026-01` | pass / 0 / 0 | 17 |
| `tr01-wf19-process-20261026-01` | pass / 0 / 0 | 21 |
| `tr01-wf19-transport-20261026-01` | pass / 0 / 0 | 12 |
| `tr01-wf19-recovery-20261026-01` | pass / 0 / 0 | 20 |
| `tr01-wf19-assessments-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-entities-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-evidence-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-graphprojection-20261026-01` | pass / 0 / 0 | 17 |
| `tr01-wf19-indicators-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-parties-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-reference_data-20261026-01` | pass / 0 / 0 | 17 |
| `tr01-wf19-reporting-20261026-01` | pass / 0 / 0 | 17 |
| `tr01-wf19-activity-20261026-01` | pass / 0 / 0 | 10 |
| `tr01-wf19-source-generated-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-source-schema-20261026-01` | pass / 0 / 0 | 12 |
| `tr01-wf19-source-catalog-20261026-01` | pass / 0 / 0 | 11 |
| `tr01-wf19-source-identity-20261026-01` | pass / 0 / 0 | 12 |
| `tr01-wf19-lint-go-20261026-01` | pass / 0 / 0 | 15 |
| `tr01-wf19-lint-backend-boundary-20261026-01` | pass / 0 / 0 | 12 |

The portfolio covers **42 distinct rows**. All **338** reported cleanup records
completed, with no failed or blocked cleanup. Transport additionally retains
**64** service scopes with distinct physical allocation identities; each reports
physical settlement completed, passed child cleanup and no failures. The
qualification source/profile remains 16 serial lifetimes, 32 in four-way waves,
and 16 competing lifetimes, with the existing four-lane admission. Template
security/durability, transaction settlement, forced-drop/barrier bounds, reset,
roles/ACLs, process recovery and separately owned performance snapshots remain
selected. No product migration, S3 internals, browser workflow or performance
policy redesign was introduced.

Deliberate internal API changes and migration impact:

- `OpenPurposeDatabase` requires a caller context. All repository call sites were
  migrated atomically; no context-free wrapper remains.
- A failed owner template attempt permanently retires its generation. Start a
  fresh physical owner for another attempt; cleanup authority remains available.
- `DetachShared`, registry active/detach state and process detach callbacks are
  deleted. Borrowers release references; physical ownership stays with OwnedCluster.
- Migration factories require the admitted postgres_migration/exclusive binding.
  Three entity rows and evidence cutover retain IDs with corrected capability;
  graph, reference-data and reporting mixed rows split under the three approved
  IDs. Every assertion is selected exactly once in authored/generated routing.
- Migration timing publication accepts the explicit retained environment, removing
  the last ambient evidence dependency found in this seam. No legacy adapter,
  generic dependency container or new package hierarchy was added.

Failures and limitations are retained, not converted into passes. E-045 preserves
expected template characterization and nil-context lint failure; E-046 preserves
expected attachment characterization, two intermediate build failures and the
quarantine-diagnostic regression. Their corrective witnesses pass on the final
frozen source. The historical 13 E-041 roots still qualify only their original
source/selections. No failed final run was replaced or hidden; all 22 final graph
roots succeeded on their first frozen-source invocation.

`make lint-markdown` exited 0 at `tr01-wf19-lint-docs-20261026-01`. Its configured
globs exclude this tracker, so this document's statuses, IDs, tables, commands,
links, assertion successors and historical block were reviewed directly after
saving. Documentation lint is not a product/specification completeness test.

Retained-run maintenance was **skipped because RESULTS_DIR was unset** during
agent-finalize; no compatible successful full warm-check root exists. No blanket
owner-wide check, release gate, benchmark acceptance or instrumentation-required
performance assessment was run or claimed. The observability audits report
optional execution/resource diagnostics unavailable/not_observed in this normal
profile; that is not a performance claim. Primary-owner semantic review remains
human-owned. Separate preceding-handoff WF-33 remains unchanged and outside scope.

Preservation: all 7,338 frozen non-Markdown source/input hashes still match; the
historical details block is byte-identical to closure entry; staged Git entries
are unchanged. Exactly 33 entries differ from the closure baseline, all within
the approved fixture/test/caller/routing/tracker scope. Existing edits outside
those entries are preserved. Make may refresh index stat metadata and ignored
build/cache/results trees; those reproducible trees were not exhaustively
fingerprinted. No commit or staging was performed. Reproducible evidence and
comparison inventories are under `/tmp/cartulary-pg-closure-_vbuwtfo/`; the
controlling handoff remains this document alone.

### WF-19 freeze and exact final portfolio — E-047 entry

`env -u RESULTS_DIR CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-finalize-20261026-01 make agent-finalize`
exited 0. Inspection found no executable/input changes against the pre-finalize
snapshot. Retained-run maintenance was skipped because RESULTS_DIR was unset;
there is no compatible successful full warm-check root. Executable source is
now frozen. The final portfolio below is recorded before execution. Historical
13-root evidence and WF-17/WF-18 runs are not substituted for these obligations.

Frozen source digest: `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185`.
Toolchain/system identities are those recorded at E-045. The private freeze
inventory is `/tmp/cartulary-pg-closure-_vbuwtfo/final-source.json`; it excludes
Markdown, as do executable source identity inputs. Only tracker/documentation
updates may follow without reopening the final evidence compatibility review.

```sh
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-lifecycle-20261026-01 make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.unit.testservices_lifecycle_contract,harness.browser.integration.postgres_template_preparation,harness.browser.integration.postgres_transaction_isolation,harness.browser.integration.postgres_owned_barrier_observation,harness.browser.integration.postgres_single_drop,harness.browser.integration.postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277,harness.browser.integration.reset_checkpoint_characterization,harness.browser.integration.managed_service_assertions_for_runtime_reset_rec_79940f98ab,harness.browser.integration.performance_fixture_snapshot_lifecycle
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-migrations-20261026-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.concurrent_apply_locking,module.database_migrations.integration.production_ddl_v2_recurrence,module.database_migrations.integration.production_preflight_state_matrix,module.database_migrations.integration.schema_bootstrap,module.database_migrations.integration.schema_readiness_matrix,module.database_migrations.unit.test_harness_targeted_operation_validation,module.database_migrations.integration.cryptographic_state_admission
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-postgres-20261026-01 make test-slice OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle,platform.postgres.integration.fresh_fixture_admission,platform.postgres.integration.purpose_identity_roles_and_acl
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-process-20261026-01 make test-slice OWNER=app.server ROWS=app.server.support_unit.incident_create_commit_fault_requires_validated_efdc344831,app.server.process.effective_deployment_configuration_fails_the_rea_c17cbafdfc,app.server.process.a_fresh_process_reaches_health_and_ready_only_af_c93818219a
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-transport-20261026-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.published_postgre_sql_transport_qualification_ac_38acf197c9
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-recovery-20261026-01 make test-slice OWNER=module.recovery ROWS=module.recovery.process.the_packaged_server_serves_the_coherently_restor_6b1731590e
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-assessments-20261026-01 make test-slice OWNER=module.assessments ROWS=module.assessments.store.assessment_create_idempotency_migration_preserve_b451ac58f2
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-entities-20261026-01 make test-slice OWNER=module.entities ROWS=module.entities.support_integration.entity_identifier_claims_37_3ba3e2aaaf,module.entities.support_integration.entity_source_integrity_migration_36_e6d97d1e54,module.entities.support_integration.timeline_mention_projection_upgrade_preserves_so_dabe8cb065
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-evidence-20261026-01 make test-slice OWNER=module.evidence ROWS=module.evidence.integration.verification_cutover
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-graphprojection-20261026-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.migration_reset,module.graphprojection.storage.graph_projection_current_head_schema_preserves_r_cf74406ebd
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-indicators-20261026-01 make test-slice OWNER=module.indicators ROWS=module.indicators.storage.lifecycle_support_reference_integrity
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-parties-20261026-01 make test-slice OWNER=module.parties ROWS=module.parties.claims.party_active_claims_are_migration_safe_concurren_9ad2db3764
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-reference_data-20261026-01 make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.integration.canonical_reference_data_cutovers_reject_incompa_3682e9346f
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-reporting-20261026-01 make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c,module.reporting.integration.reporting_identity_cutover_preserves_incompatibl_62c39cd14e
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-activity-20261026-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-source-generated-20261026-01 make source-check CHECK=generated
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-source-schema-20261026-01 make source-check CHECK=schema
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-source-catalog-20261026-01 make source-check CHECK=catalog
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-source-identity-20261026-01 make source-check CHECK=identity
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-lint-go-20261026-01 make lint CHECK=go
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-lint-backend-boundary-20261026-01 make lint CHECK=backend-boundary
```

### WF-18 checkpoint / E-046 — admission, binding and routing qualified

WF-18/S-018, T-016 and V-022 are DONE. F-022/F-023/F-025/F-026 are corrected;
all six findings now have implementation witnesses, with overall acceptance
still waiting for WF-19. The adopted requirements were not changed.

Attachment admission now stores a pending/terminal result per advertised
generation. Inspection must confirm identity before readiness/publication
failure can retire that generation. Same-binding callers share the result;
cancelled waiters leave the initiating attempt intact, and another generation
can proceed during stalled admission. Unverified claims confer no ownership.
Readiness cancellation and journal failure retain original causes and quarantine;
subsequent calls neither probe nor prepare again.

Fixtures copy explicit binding inputs at entry. Registry reuse compares connection,
capability, policy, dependency and evidence-destination inputs. Creation guards,
policy resolution, migration admission and events use that retained binding.
Production entrypoints capture process inputs; private registry calls receive
explicit inputs. The existing explicit testservices publisher is preserved.
Migration timing publication was another ambient evidence consumer, discovered
while tracing the same boundary: `suiteservices.MeasureMigration(ctx, env, fn)`
now takes the admitted inputs at its two pgtest callers. Its existing private
publication/shutdown witness passes with deliberately conflicting ambient inputs.
No generic dependency container, scheduler change or compatibility adapter was added.

Both migration factories require postgres_migration plus exclusive scope before
CREATE/open/source access. The negative matrix crosses missing, unknown,
transaction, dedicated and migration capability with absent/shared/exclusive/
unknown scope, for both factories; the valid pair is exercised by existing fake
issuance/canonical-source contracts and real caller rows. Invalid target bounds
and forged/zero capability rejection remain selected. Physical ownership alone
adds no migration authority. Four migration-only rows change capability; the
three mixed rows split exactly as recorded below, retaining IDs, verification
links, aggregate placement, dependencies and runtime/resource profiles.

`DetachShared`, registry active state/detachment and process detach callbacks
are removed without aliases. The process attachment helper returns its original
object-store error. The real template witness checks the owner remains usable
and the same generation remains admitted after borrower use; quarantine tests
check repeated attachment cannot restore eligibility. No physical shutdown is
exposed by Harness. Historical detach descriptions in E-039/E-041 are history,
not current API guidance.

The caller audit traces 43 top-level product tests through local helper calls to
15 migration rows, with no unmapped test. The exact 19-row portfolio below also
includes the unit capability row and ordinary assertions retained by row splits.
All nine owner task guides and harness.evidence_accounting/app.server guides
exited 0. No owner-wide rerun was substituted for exact caller selections.

Validation used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<basename>` from the repo
root. For each root, run-manifest.json retains exact target/OWNER/ROWS/CHECK and
source/tool/system identities. Commands are public `make test-slice` using A/B/D,
the recorded caller selection, or the following additions: harness A appends
`harness.browser.integration.postgres_template_preparation` and
`harness.browser.integration.postgres_transaction_isolation`; migration selection
is the recorded module.database_migrations caller list plus B's cryptographic
admission row; unit-postgres is C's certificate_fixture_lifecycle row;
unit-migrations is B's test_harness_targeted_operation_validation row.
`activity` uses the exact evidence-accounting row recorded below.

| Root under `.cartulary/test-results/` | Public target / selected owner or check | Outer exit | Source digest | Result |
| --- | --- | --- | --- | --- |
| `tr01-wf18-activity-20261026-01` | `make test-slice OWNER=harness.evidence_accounting` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-assessments-20261026-01` | `make test-slice OWNER=module.assessments` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-characterization-20261026-01` | `make test-slice OWNER=harness.browser` | 2 | `sha256:f2c2e40a8ef08268561bbf910d7029ee6bf78f0ffa665b5f19b1999d2347b98e` | fail |
| `tr01-wf18-entities-20261026-01` | `make test-slice OWNER=module.entities` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-evidence-20261026-01` | `make test-slice OWNER=module.evidence` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-format-20261026-01` | `make format` | 0 | `sha256:7536ca4652ec653d2d1a4b58c32bbf539cf4595defc12d0492aef9a275d4c75b` | pass |
| `tr01-wf18-format-20261026-02` | `make format` | 0 | `sha256:468116d6bb0bcdce931be8ea2c157b3466f37d1c3b5c03c85fe1228485bceb45` | pass |
| `tr01-wf18-format-20261026-03` | `make format` | 0 | `sha256:f5a456c76eb789b6cb0cd33181e59e157af77ef3609da6688c0bcf6d7ad32e01` | pass |
| `tr01-wf18-format-20261026-04` | `make format` | 0 | `sha256:aad088c4b1fe25695954555aebe6c74f98f0cbbde9ee98a624a7513c626a53ee` | pass |
| `tr01-wf18-graphprojection-20261026-01` | `make test-slice OWNER=module.graphprojection` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-harness-20261026-01` | `make test-slice OWNER=harness.browser` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-indicators-20261026-01` | `make test-slice OWNER=module.indicators` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-lint-boundary-20261026-01` | `make lint CHECK=backend-boundary` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-lint-go-20261026-01` | `make lint CHECK=go` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-lint-go-20261026-02` | `make lint CHECK=go` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-lint-go-20261026-03` | `make lint CHECK=go` | 0 | `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185` | pass |
| `tr01-wf18-migrations-20261026-01` | `make test-slice OWNER=module.database_migrations` | 2 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | fail |
| `tr01-wf18-migrations-20261026-02` | `make test-slice OWNER=module.database_migrations` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-parties-20261026-01` | `make test-slice OWNER=module.parties` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-process-20261026-01` | `make test-slice OWNER=app.server` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-reference_data-20261026-01` | `make test-slice OWNER=module.reference_data` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-reporting-20261026-01` | `make test-slice OWNER=module.reporting` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-source-catalog-20261026-01` | `make source-check CHECK=catalog` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-source-generated-20261026-01` | `make source-check CHECK=generated` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-source-identity-20261026-01` | `make source-check CHECK=identity` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-source-schema-20261026-01` | `make source-check CHECK=schema` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-unit-20261026-01` | `make test-slice OWNER=harness.browser` | 0 | `sha256:0d13eb2ee84ae2202101891588a2376295c0d6b69c2e7b14e21290562bfb5ff9` | pass |
| `tr01-wf18-unit-harness-20261026-02` | `make test-slice OWNER=harness.browser` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-unit-migrations-20261026-02` | `make test-slice OWNER=module.database_migrations` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-unit-migrations-20261026-03` | `make test-slice OWNER=module.database_migrations` | 0 | `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185` | pass |
| `tr01-wf18-unit-postgres-20261026-01` | `make test-slice OWNER=platform.postgres` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-unit-postgres-20261026-02` | `make test-slice OWNER=platform.postgres` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-unit-postgres-20261026-03` | `make test-slice OWNER=platform.postgres` | 0 | `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185` | pass |

Additional commands: `make generate` exited 0 at
`tr01-wf18-generate-20261026-01`; `make lint-markdown` exited 0 at
`tr01-wf18-lint-docs-20261026-01`. The tracker is outside lint globs and was
reviewed directly. Two intermediate `make format` invocations exited 2 before
the graph could start, roots `20261026T231734Z-p72355` and
`20261026T231758Z-p73018`: incomplete receiver/binding signature migration
prevented building the harness. The build stderr identified the unset binding
in startOwnedAttempt; passing the entry snapshot through that private call
closed the defect. Subsequent format and Go lint roots above pass.

Characterization failed as expected with product/test_assertion_failure in the
helper row. The first full migration selection then found a real diagnostic
regression: new admission returned generic text for quarantine, and policy
resolution could obscure the quarantine witness. The repair gives quarantine
precedence and uses the existing typed fixture-failure publisher. Unit successors
and the entire seven-row migration selection pass afterward. These failed roots
are preserved. Final small test changes add a waiting preparation deadline and
remove unused policy table fields; their affected unit rows and Go lint pass.

All successful roots report completed cleanup, with no failed or blocked cleanup.
Toolchain/system identities remain those in E-045. Source identity changes are
recorded per slice run; WF-19 must freeze and requalify compatible final evidence.
The historical details block remains byte-identical to closure entry. No files
outside the 33 scoped source/test/routing/tracker entries differ from the closure
baseline; generated render-index changes came from Make. The Git index's stat
cache may refresh during canonical tools; no staging was requested or performed.

Next gate: only after this saved checkpoint, run `make agent-finalize` with
RESULTS_DIR unset (there is no compatible successful full warm check root),
inspect its edits, freeze executable inputs, run final A–E plus the exact caller
and timing selections, and audit successful roots. Acceptance, T-017 and V-023
remain open until the final handoff checkpoint is saved.

### WF-18 exact migration caller selections — recorded before execution

The 16 prior caller rows plus three approved splits are the additional caller
portfolio. Includes helper-mediated calls (`migratedSchemaReadinessDatabase`,
`emptyMigrationDatabase`, reference-data cutover helpers, and party clock cutover).
Retained ordinary assertions in split rows are included to detect routing loss.
Run each command with `CGO_ENABLED=0` and a fresh recorded run ID. B and this
inventory overlap intentionally at two migration rows; combine them into one
owner selection for the final portfolio. The narrow evidence-accounting row is
added because migration timing publication now receives explicit bound inputs.

```sh
make test-slice OWNER=module.assessments ROWS=module.assessments.store.assessment_create_idempotency_migration_preserve_b451ac58f2
make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.concurrent_apply_locking,module.database_migrations.integration.production_ddl_v2_recurrence,module.database_migrations.integration.production_preflight_state_matrix,module.database_migrations.integration.schema_bootstrap,module.database_migrations.integration.schema_readiness_matrix,module.database_migrations.unit.test_harness_targeted_operation_validation
make test-slice OWNER=module.entities ROWS=module.entities.support_integration.entity_identifier_claims_37_3ba3e2aaaf,module.entities.support_integration.entity_source_integrity_migration_36_e6d97d1e54,module.entities.support_integration.timeline_mention_projection_upgrade_preserves_so_dabe8cb065
make test-slice OWNER=module.evidence ROWS=module.evidence.integration.verification_cutover
make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.migration_reset,module.graphprojection.storage.graph_projection_current_head_schema_preserves_r_cf74406ebd
make test-slice OWNER=module.indicators ROWS=module.indicators.storage.lifecycle_support_reference_integrity
make test-slice OWNER=module.parties ROWS=module.parties.claims.party_active_claims_are_migration_safe_concurren_9ad2db3764
make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.integration.canonical_reference_data_cutovers_reject_incompa_3682e9346f
make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c,module.reporting.integration.reporting_identity_cutover_preserves_incompatibl_62c39cd14e
make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities
```

### WF-17 checkpoint / E-045 — preparation recovery qualified

WF-17/S-017 and T-015 are DONE. F-021 and F-024 are corrected against the
unchanged adopted obligations. `OpenPurposeDatabase(ctx, dsn, purpose)` now
rejects nil/cancelled contexts before I/O and preserves context causes from
production admission. All repository callers supply an operation/test context.
The independent crypto guard receives the preparation context. One owner-local
attempt runs synchronously with the existing two-minute bound; waiters cancel
independently, and a failed attempt is terminal only after settlement. Invalid
requests do not retire the owner. There is no in-place retry or background recipe.

Pristine CREATE and the existing clone finalization callback synchronously record
local ownership and publish the created event before provisioning. Notification
failure stops dependent work; disposal runs after catalog admission returns.
Rejected CREATE grants no ownership. Exported low-level CloneDatabase remains
available to recovery and performance owners. The real template witness observes
both ownership ledgers at notification for pristine and cloned databases.

Assertion additions are nested beneath the existing selected template contract:
retained terminal failures, successful concurrent preparation, independently
cancelled waiter, both stalled opens with cancellation/deadline, settlement-before-
terminal publication, nil/invalid pre-admission, and real TCP connection-open
cancellation/deadline. Existing close/drop failure and unowned-target assertions
remain. No top-level test or selector was retired in this slice.

Commands below ran from the repository root with `CGO_ENABLED=0` and
`CARTULARY_TEST_RUN_ID=<root basename>`. Exact selections: **A** is Section 8 A;
**B** is Section 8 B; **C-unit** is
`OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle`;
**A+template** is A with
`harness.browser.integration.postgres_template_preparation` appended to ROWS.
All test commands are `make test-slice` with those exact arguments.

| Run root under `.cartulary/test-results/` | Command / selection | Outer exit | Source digest | Result |
| --- | --- | --- | --- | --- |
| `tr01-wf17-characterization-20261026-01` | A | 0 | `sha256:b3365868908ddc3c0d7ad51ec0941edc8995988041bf097b3f68acfa45264807` | pass |
| `tr01-wf17-characterization-20261026-02` | C-unit | 2 | `sha256:b3365868908ddc3c0d7ad51ec0941edc8995988041bf097b3f68acfa45264807` | fail |
| `tr01-wf17-unit-20261026-01` | C-unit | 0 | `sha256:e586dd2240d92f5e92f08c3b0c4ee9e19d2c92231d430957168c2a1ad2433133` | pass |
| `tr01-wf17-unit-20261026-02` | C-unit | 0 | `sha256:d941b537a26c01b20dc30cd6d0522d132c9c84e657009e1c63e42a43ee31bb37` | pass |
| `tr01-wf17-harness-20261026-01` | A+template | 0 | `sha256:d941b537a26c01b20dc30cd6d0522d132c9c84e657009e1c63e42a43ee31bb37` | pass |
| `tr01-wf17-migrations-20261026-01` | B | 0 | `sha256:d941b537a26c01b20dc30cd6d0522d132c9c84e657009e1c63e42a43ee31bb37` | pass |
| `tr01-wf17-lint-go-20261026-01` | make lint CHECK=go | 2 | `sha256:d941b537a26c01b20dc30cd6d0522d132c9c84e657009e1c63e42a43ee31bb37` | fail |
| `tr01-wf17-lint-go-20261026-02` | make lint CHECK=go | 0 | `sha256:70d6064db870a4258e31837e5e077ff77860d39fa68aed2ab0353909bbd5bd6f` | pass |

The first characterization A pass did not select the newly changed template
contract; C-unit then failed with typed `product/test_assertion_failure` (unit
exit 10, outer exit 2), qualifying the retained-failure/waiter gap. No pass was
claimed for that uncovered boundary. Go lint's first failure was SA1012 for the
intentional nil-context rejection test; a one-line test-only annotation explains
that negative case. The corrected lint passes and compiles migrated callers.
No product code suppression was added.

All successful test roots report completed cleanup with no cleanup failures.
Toolchain digest remains `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`;
system digest remains `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Source digests differ during correction and comment-only lint repair; these are
slice evidence, not the frozen final portfolio. WF-19 must requalify final source.
`make help ALL=1`, task guides for harness.browser, platform.postgres and
module.database_migrations all exited 0. `make format` exited 0 at roots
`20261026T230846Z-p15398`, `20261026T231004Z-p22998`, and
`20261026T231100Z-p29881`. `make lint-markdown` exited 0 at
`tr01-wf17-lint-docs-20261026-01`; its configured globs omit this tracker,
which was re-read and reviewed directly before this checkpoint.

Next gate: WF-18 may begin only after this checkpoint is saved. Characterize
verified-generation readmission, binding isolation and migration admission;
remove DetachShared and redundant state; correct/split caller routing; then
regenerate and validate V-022. WF-19 final acceptance remains open. No
complete-owner, release or performance acceptance is asserted.

### Closure entry / E-044 — WF-17 authorized

The implementation request adopts the complete closure plan in this conversation.
Two decisions supersede the review's optional alternatives: remove DetachShared
and its unused state/callbacks, and retain one terminal preparation result per
physical owner, retiring the generation after failed preparation. Cancellation
of a waiting caller does not cancel the initiating attempt. Invalid requests
rejected before an attempt do not retire a healthy owner.

Source baseline is main at `867e97ed49829d671c58576232999d1fc9a4e3fb`, with 33
pre-existing changed/new paths and 7,649 tracked/nonignored entries. Preservation
material is `/tmp/cartulary-pg-closure-_vbuwtfo`; starting tracker SHA-256 is
`d14d9f9130113893ff75b2abc7a86c1cfdf458cb8c3d869304f0ee6a2d7089d5`.
Read root/internal/tools instructions and test-operations skill. REQ-804/405/810/811
already own the required behavior; no normative amendment is needed. Supporting
API guidance will be corrected with callers. No executable check consumes Markdown.

WF-17 implements context-aware purpose opens, one terminal preparation attempt,
and synchronous ownership notification after confirmed CREATE and before
prerequisites/clone grants. WF-18 implements verified-generation failure retention,
immutable fixture inputs/event destinations, strict migration-capability admission,
and DetachShared removal. Its routing correction changes the four migration-only
rows in entities/evidence and splits mixed graph/reference-data/reporting rows.
The three new IDs are the previously derived IDs in the approved plan. Preserve
profiles, dependencies, verification links and exclusive isolation. WF-19 finalizes,
freezes source, validates A–E plus migration-caller rows, audits exact roots and
saves handoff. Every exit updates findings, assertion mapping, selections, work
status, blockers and exact evidence before the next slice begins.

Initial validation is V-021: A, B and real template preparation. Add fault
characterization before implementation and classify expected failures. Run all
repository verification through Make, record native sessions and outer exits,
and retain failures. Broader final verification waits for WF-19's finalization.
No correction validation has run at this entry checkpoint.

### Post-implementation review — E-042

This review supersedes E-041's unconditional completion assessment, not its
historical commands/results or completed workstream checkpoints. F-021–F-025
remain required; F-026 is optional. The existing design is worth retaining, with
bounded corrections in WF-17–WF-19. Implementation authorization is **not granted
by this review invocation**. No production refactor, repair or product rerun was
performed.

Review baseline: `main`, HEAD `867e97ed49829d671c58576232999d1fc9a4e3fb`;
21 already modified tracked files and 12 nonignored new pgtest files; 7,649
path/content/identity entries. Starting tracker SHA-256:
`80eaf2857cdb07dc35a22564206920881aaae4d5262d716b5a353091010fee9f`.
The admitted output is the same regular, singly linked canonical file. External
baseline and review records are at `/tmp/cartulary-pg-review-ghyk5hap`.
A plain initial Git status preceded the optional-write-disabled baseline; any
prior stat-cache refresh is outside the preservation comparison. All subsequent
Git inspection disabled optional writes. No staging or commits were performed.

Source review used bounded `rg`, `sed`, numbered source reads and
`git --no-optional-locks diff`, with relevant dependencies opened directly.
Searches that referenced nonexistent guessed helper paths were corrected by file
and symbol discovery; they produced no edits or validation evidence. Adopted
Harness REQ-804/405/810/811 and the plan govern the findings; `domain.md` is
vocabulary/navigation. No owner contradiction or necessary spec amendment was
found. The existing pool/readmission wording change preserves the requirement.

All thirteen E-041 run manifests and summaries still report pass (70 total units,
zero failed/skipped/cancelled). Source/toolchain/system identities match E-041;
all retained cleanup outcomes are `completed`, secret scans are `pass`, and the
transport root contains 64 service-scope records. Direct hash comparison of the
7,338-file E-041 freeze found no executable/input change. This corroborates
retained results; it does not reproduce faults, rerun observability qualification,
or establish coverage of F-021–F-025. `evidence-review.json` contains the inspected
root identities and results. Tests, services, generation, `agent-finalize` and
new observability runs were intentionally not invoked for this review-only task.

### Review documentation save — E-043

V-024 completed. The inspected lint path used installed-only Node/frontend
readiness, the existing workspace lock, fixing disabled, and external result/cache
directories. Exact command from the repository root:

```sh
env GIT_OPTIONAL_LOCKS=0 \
  CARTULARY_PREPARATION_POLICY=installed_only \
  CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-pg-review-ghyk5hap/results \
  CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-pg-review-ghyk5hap/static-analysis \
  CARTULARY_TEST_RUN_ID=tr01-postgres-review-doc-20261026-01 \
  COREPACK_ENABLE_NETWORK=0 make lint-markdown
```

The lint summary records **pass, exit 0**, 17,650 ms, completed
`2026-10-26T22:27:16.628Z`, zero tests. Exact artifact:
`/tmp/cartulary-pg-review-ghyk5hap/results/tr01-postgres-review-doc-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The asynchronous command's outer Make exit was not retained by this review's
monitor; the recorded exit above is the lint command's tool summary. No retry was
run to replace missing monitor metadata. Current lint globs exclude this tracker.
Direct review checked its twelve sections, balanced fences/tables, local links,
stable IDs, contract/finding/slice/validation references, preserved historical
snapshot and distinction between source findings and executed evidence.
`git --no-optional-locks diff --check` passed. Final evidence annotations do not
change the linted file population and were reviewed directly.

The final comparison against all 7,649 starting entries found **only this tracker
changed**, no new/deleted paths, and identical raw index and index entries. The
embedded historical snapshot remains byte-identical. All other tracked and
nonignored untracked content/file identities, including the prior implementation
edits, are preserved; no staging or commit occurred. External comparison records
are beside the review baseline. Ignored build/dependency/output trees were not
exhaustively fingerprinted; reviewed command paths redirected their outputs.
T-014 is DONE. Next entry is later-authorized WF-17 characterization, not an
implementation action in this review. Required acceptance remains open.

### Current implementation — WF-12 entry / E-036

The user authorized WF-12 through WF-16 in dependency order, with a saved exit
checkpoint before every successor. Starting HEAD is
`867e97ed49829d671c58576232999d1fc9a4e3fb`; the index is unchanged and the sole
pre-existing modification is this tracker. Content/file-identity fingerprints
for 7,637 tracked/nonignored entries, the original tracker, HEAD and index are
retained in `/tmp/cartulary-pg-remediation-kmpj28se`. Ignored outputs and installed
dependencies are not exhaustively fingerprinted. E-030–E-035 remain historical
planning evidence, not this implementation baseline.

Source and owner revalidation confirms F-015–F-020. F-016 additionally covers
discarded close/drop errors and preservation of confirmed CREATE ownership after
later preparation failures. Failed CREATE never grants ownership of an existing
target. The authorized specification edit replaces “returned to the pool” with
“readmitted for reuse” in §11.0; it introduces no pooling or behavioral change.
Domain vocabulary and product specifications require no amendment.

V-015 includes A–D plus real template preparation, transaction isolation,
owned cleanup/barrier and the explicit transport qualification before movement.
Each run uses a fresh literal identity and `CGO_ENABLED=0`; exact commands,
outer exits, artifacts and source identities will be saved at the WF-12 exit.
The existing assertion successor ledger is controlling; no incidental method
count, smoke-only assertion or borrowed-shutdown error becomes a requirement.

Final API decisions: `StartOwned(ctx, StartOptions) (*OwnedCluster, error)`,
`OwnedCluster.Fixture()`, owner-only explicit template preparation and physical
shutdown, attachment-only `Start`/`StartShared`, and no-I/O `DetachShared()`.
There will be no constructor aliases or physical shutdown on `Harness`.

### WF-12 exit / E-037 — characterization qualified

WF-12 completed before WF-13 began. All six commands used
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<ID> make test-slice OWNER=<owner> ROWS=<selection>`
from the repository root, exited 0, and share source
`sha256:daaea890979e0349dc4392aeffd167b41b172981b29d13dca99008fbad112eb9`
and toolchain
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
Run roots are `.cartulary/test-results/<ID>`; the existing selections in §8
give exact A–D row arguments.

| Run ID | Selection | Duration ms | Result |
| --- | --- | --- | --- |
| `tr01-wf12-contract-20261026-01` | A; harness.browser | 10977 | pass, 2/2 units |
| `tr01-wf12-migrations-20261026-01` | B; module.database_migrations | 22391 | pass, 9/9 units |
| `tr01-wf12-postgres-20261026-01` | C; platform.postgres | 14078 | pass, 5/5 units |
| `tr01-wf12-process-20261026-01` | D; app.server | 88243 | pass, 10/10 units |
| `tr01-wf12-lifecycle-20261026-01` | harness.browser: `postgres_template_preparation`, `postgres_transaction_isolation`, `postgres_owned_barrier_observation`, `postgres_single_drop`, `postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277`, all under `.integration.` | 19655 | pass, 6/6 units |
| `tr01-wf12-transport-20261026-01` | harness.browser: `harness.browser.integration.published_postgre_sql_transport_qualification_ac_38acf197c9` | 96331 | pass; 64 fresh lifetimes |

Each exact root passed `make harness-observability-check RESULTS_DIR=<root>
CHECK=all` with outer exit 0 and `full_validity=1`; cleanup completed and retained
secret scans passed. There were no failed/skipped/cancelled units. These are
pre-move evidence only. `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=tr01-wf12-markdown-20261026-01 make lint-markdown` exited 0;
summary: `.cartulary/test-results/tr01-wf12-markdown-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
Its configured globs exclude this tracker, which received direct structure and
diff review; `git diff --check` passed. No product/source movement occurred in
WF-12. The meaningful successor/retirement ledger in §7 remains complete; RB-005
is resolved and S-013 is ready. Human projection-fidelity review is not inferred
from test success. Next: canonical preparation, explicit admission and failure
ownership in WF-13.

### WF-13 exit / E-038 — canonical preparation qualified

One pgtest operation now prepares the ordinary template and returns database/error.
Owned callers prepare explicitly; ordinary clone preparation fails without a sealed
template. Service composition delegates through its physical service capability
and supplies the lifecycle publisher; duplicate create/migrate events and its
SQL/migration recipe were removed. Confirmed CREATE is retained through subsequent
errors; rejected CREATE grants no deletion authority. Template handle/guard close,
drop and event-publication failures are retained with their primary cause and
uncertainty prevents reuse. Fault cases cover create rejection, committed-create
failure, create/migrate event publication, open, initialization, cancellation,
close, seal and combined close/drop failure. Real negative cases now use independent
owned lifetimes rather than accidentally stopping at a previous quarantine.

Validation used `CGO_ENABLED=0`, fresh IDs and public Make commands. Two initial
build failures were diagnosed and corrected: an event receiver typo blocked
`tr01-wf13-format-20261026-01` (build step summary under `harness-runtime-build`),
and an unused import after deleting the duplicate recipe caused
`tr01-wf13-contract-20261026-01` to report `infra/preflight_error` with row
`missing_selector_result`. These are implementation compile failures, not service
instability; their artifacts remain. Format passed at `tr01-wf13-format-20261026-03`.

The exit portfolio shares source
`sha256:9c4615b746a1bab5ae757a313fe7de211200adbf915be1651a22b403d7e6f6a9`:

| Run ID | Exact selection | Duration ms | Result |
| --- | --- | --- | --- |
| `tr01-wf13-lifecycle-20261026-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.unit.testservices_lifecycle_contract,harness.browser.integration.postgres_template_preparation` | 29441 | pass |
| `tr01-wf13-migrations-20261026-01` | Selection B in §8 | 21595 | pass |
| `tr01-wf13-postgres-20261026-02` | Selection C in §8, including the new fault matrix | 20809 | pass |

Each root `.cartulary/test-results/<ID>` passed `make
harness-observability-check RESULTS_DIR=<root> CHECK=all`, outer exit 0,
`full_validity=1`. Native commands exited 0; cleanup completed and retained-secret
scans passed. Earlier WF-13 passes qualify earlier source only. No selector change
was required; the new fault matrix is selected by the existing template contract.
This checkpoint closes WF-13 before WF-14 begins. Next: move physical lifetime and
the explicit preparation method to OwnedCluster, then isolate dependencies and
attachment registries without aliases.

### WF-14 exit / E-039 — ownership and dependencies qualified

`StartOwned(ctx, StartOptions)` now returns `*OwnedCluster`. Its callbacks retain
physical recovery authority before readiness; `Fixture()` and template preparation
remain unavailable until admission completes. The owner holds termination, TLS,
readiness, diagnostics and template preparation. `Harness` has database ownership,
transactions and quarantine, with no physical close or embedded owner. Direct
callers migrated without wrappers or aliases; closure retires fixture admission.
Private startup/database/template/diagnostic operations replace mutable function
hooks. Tests construct independent attachment registries. Every cache hit validates
the complete admitted binding identity; detach preserves generation records,
ownership and quarantine. Independently admitted generations remain usable.

The process fixture now uses `DetachShared()` on object-store attachment failure
and returns the initiating error unchanged. Real template integration proves local
detach leaves PostgreSQL alive and keeps the admitted fixture record. C-011/C-013/
C-015 and F-015/F-017 are implemented and qualified at this checkpoint; F-019's
process boundary is complete, with capability/smoke consolidation left to WF-15.
The original assertion IDs remain selected; owner lifecycle assertions now use
local operations and `OwnedCluster` callbacks. No mutable PG hook substitutions,
borrowed shutdown, `StartOwnedWithLabels`, or `StartOwnedWithOptions` remain.

All commands used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<ID> make test-slice` with
§8 selections. Composite commands use the comma-joined union of the stated rows
under `OWNER=harness.browser`. Results are under `.cartulary/test-results/<ID>`:

| Run ID | Exact row selection | Source digest | Duration ms | Native result / outer exit |
| --- | --- | --- | --- | --- |
| `tr01-wf14-migrations-20261026-01` | B | `sha256:37bb7ebfb485107a7915cd4fb8d6cf5dd7c2e7d0c95fe2d89511cf17273a609b` | 24022 | pass / 0 |
| `tr01-wf14-lifecycle-20261026-02` | A plus first E command | `sha256:d1a13e83352aad66d823eace2650f6d922cf5a170fc28643b078ba0ce3417c8c` | 60262 | pass / 0 |
| `tr01-wf14-process-20261026-01` | D | `sha256:d1a13e83352aad66d823eace2650f6d922cf5a170fc28643b078ba0ce3417c8c` | 78568 | pass / 0 |
| `tr01-wf14-postgres-20261026-02` | C | `sha256:cb6328e926106f4d204cbca33db4c1d5302c8ba253c9a46c96d1a428154ef5a8` | 16978 | pass / 0 |
| `tr01-wf14-transport-20261026-01` | A plus template preparation and explicit transport row | `sha256:cb6328e926106f4d204cbca33db4c1d5302c8ba253c9a46c96d1a428154ef5a8` | 105201 | pass / 0 |

The final transport command selected exactly A's two rows,
`harness.browser.integration.postgres_template_preparation`, and
`harness.browser.integration.published_postgre_sql_transport_qualification_ac_38acf197c9`.
It retained all 64 fresh lifetimes and the existing four-way profile. Each listed
root passed `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`,
outer exit 0, `full_validity=1`; cleanup and retained-secret checks passed. The
source identities above deliberately distinguish checks made before the final
pre-readiness admission guard; C, A, real preparation and transport were rerun
after that guard. WF-16 will establish the compatible final portfolio.

Two implementation build failures were diagnosed before retry: format run
`tr01-wf14-format-20261026-01` had moved-import/receiver signature errors, recorded
in `harness-runtime-build/build-harness-runtime/step-summary.json`; lifecycle run
`tr01-wf14-lifecycle-20261026-01` had two borrowed forced-drop callers incorrectly
changed to `Fixture()`, reported as `infra/preflight_error` and
`missing_selector_result` in its two postgrescleanup unit logs. Both exited 2;
both are corrected and retained. Final format run
`tr01-wf14-format-20261026-04` passed. An attempted `make test-status` observation
was rejected because that target does not exist; corrected to public
`make test-run-status` without executing tests. No unresolved failure remains.

This checkpoint closes WF-14 before WF-15 starts. Next: consolidate policy and
cleanup assertions, qualify capability/smoke successors, update authored selectors,
and regenerate projections. RB-006 now gates only WF-15 exit and WF-16 entry.

### WF-15 exit / E-040 — assertions and routing consolidated

The cleanup successor passed first as a nested case under the existing targeted
capability selector (`tr01-wf15-successors-20261026-01`, selection B, exit 0,
24418 ms). Only then were the four standalone/attached clone/scratch cleanup tests
retired and `TestDatabaseFixtureCleanup` selected directly by the existing
`module.database_migrations.unit.test_harness_targeted_operation_validation` row.
Its nine cases preserve resource kind and standalone/attached/persistent borrowing
contexts, issue real private capabilities, and verify closure before exact eager
drop. Single-drop ownership, caller-context propagation and failed-recovery tests
remain separate because they prove distinct failure boundaries.

The weak `TestHarnessStartsPostgresAndRunsCurrentMigrationPath` and selector were
removed after the existing `TestPrepareIsolatedDatabaseTReturnsMigratedDatabase`
passed C at WF-14; the surviving witness proves schema and crypto admission.
All row IDs, owners, service dependencies, isolation and qualification profiles
remain unchanged. No alias maps were added. Authored changes are only the test
arrays in `module.database_migrations.json` and `platform.postgres.json`;
`make generate` updated only the corresponding generated topology input hashes.

The public policy helper now calls one resolver. Its test exposed attribution
that skipped test files inside pgtest along with implementation frames; retaining
actual `_test.go` caller frames fixes package selection without inventing a test
entrypoint. The first policy run `tr01-wf15-policy-20261026-01` failed with
`product/test_assertion_failure` (outer exit 2, 6354 ms); its logs identify
`package_priority`. After correction `tr01-wf15-policy-20261026-02` passed in
7028 ms. Capability assertions now reject nil/zero/forged identities, check source
and execution counters before rejected targets, and compare legal operations to
the canonical embedded source. The private identity token is nonzero-sized so
its pointer cannot alias an unrelated zero-sized allocation. Reflection API-shape
assertions and obsolete local restore scaffolding are gone. Tests are grouped by
responsibility in the same package, without introducing a package hierarchy.
Concurrent initial template preparation now proves exactly one recipe/publication.

All A–D successors passed on source
`sha256:b3ef01cf9ea1fa6973c4e3aec96dc8522a01c3fff3af3b366bd31bd262e55e1a`:

| Run ID | Exact §8 selection | Duration ms | Native result / outer exit |
| --- | --- | --- | --- |
| `tr01-wf15-contract-20261026-01` | A | 7372 | pass / 0 |
| `tr01-wf15-migrations-20261026-01` | B successor | 30654 | pass / 0 |
| `tr01-wf15-postgres-20261026-01` | C successor | 14558 | pass / 0 |
| `tr01-wf15-process-20261026-01` | D | 78087 | pass / 0 |

The added private canonical-source load counter was then qualified by B at
`tr01-wf15-migrations-20261026-02`, exit 0, 28593 ms, source
`sha256:04e6276a37b9a7e9ca342911b43a12258842424698d8722ff213eac36879c036`.
Each listed test root passed `make harness-observability-check
RESULTS_DIR=.cartulary/test-results/<ID> CHECK=all`, exit 0, `full_validity=1`.
Cleanup completed and retained-secret scans passed. These workstream roots are
not substituted for WF-16's compatible final-source portfolio.

Public commands used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<ID>`. Generation passed
at `tr01-wf15-generate-20261026-01` (10280 ms). `make source-check CHECK=<check>`
passed for `generated`, `schema`, `catalog`, `identity`, using roots
`tr01-wf15-source-<check>-20261026-01` (15645, 6206, 1643, 724 ms respectively),
all exits 0. A format build failed once while moving declarations because two
still-used imports had been removed (`tr01-wf15-format-20261026-02`, exit 2,
`harness-runtime-build/build-harness-runtime/step-summary.json`); corrected before
execution. Final format `tr01-wf15-format-20261026-04` passed (5450 ms).

F-018/F-019 and routing portion of F-020 are implemented. No retired selector or
old API reference remains in executable sources or routing. Human authority and
Markdown isolation are unchanged. This checkpoint closes WF-15 before WF-16
starts. Next: agent-finalize, inspect generated changes, freeze executable source,
run A–E successor union plus required static checks, audit every successful root,
and finish the handoff. No compatible successful full warm-check root exists;
agent-finalize will leave RESULTS_DIR unset and report retained maintenance skipped.

### WF-16 exit / E-041 — final validation and handoff complete

WF-12–WF-16, T-009–T-013, F-015–F-020 and C-011–C-016 are complete for this
bounded PostgreSQL effort. The following evidence is the final acceptance
portfolio; earlier runs retain their own identities and are not substituted.

Finalization ran before broader verification. The initial final Go lint found
`U1000` on the now-unused testservices template-detail helper and the nonzero-sized
capability identity marker (`tr01-wf16-lint-go-20261026-01`, outer exit 2;
`unit-logs/operation-harness.lint_go_staticcheck/stdout.log`). Its normalized result
was `product/test_assertion_failure`; these were implementation cleanup errors,
not a service failure. The helper was removed and the nonzero-sized marker made
unnamed. `make agent-finalize` was repeated, again reporting generated outputs
unchanged (zero updated files), then executable source was frozen. Superseded
WF-16 `-01` lifecycle/migration/static results remain retained but are excluded
from the final table below. No unresolved failure remains.

Final executable/input identity:
`sha256:fbc666088d1ff36c404fa5bc33c090ffab488aa0726f0002e88897436d3a2f97`.
Every graph root below has the same toolchain identity
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and system identity
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Graph/workload digests intentionally differ by exact selection and remain in each
`run-manifest.json` and target summary. Authored configuration, capabilities,
isolation scopes and qualification profiles were frozen with source; no profiles
were replaced or mixed to claim coverage. HEAD remains
`867e97ed49829d671c58576232999d1fc9a4e3fb`; worktree source is intentionally dirty.

Each command used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<ID>`. Each exact run root
is `/home/jochi/code/cartulary/.cartulary/test-results/<ID>`:

| Run ID | Public command / exact §8 selection | Duration ms | Native result / outer exit |
| --- | --- | --- | --- |
| `tr01-wf16-finalize-20261026-02` | `make agent-finalize` (RESULTS_DIR unset) | 23730 | pass / 0 |
| `tr01-wf16-lifecycle-20261026-02` | A plus first E command; exact union below | 54585 | pass / 0 |
| `tr01-wf16-migrations-20261026-02` | B successor, unchanged row ID | 21641 | pass / 0 |
| `tr01-wf16-postgres-20261026-01` | C successor, unchanged row IDs | 20945 | pass / 0 |
| `tr01-wf16-process-20261026-01` | D | 75496 | pass / 0 |
| `tr01-wf16-transport-20261026-01` | Second E command: explicit transport qualification | 91987 | pass / 0 |
| `tr01-wf16-recovery-20261026-01` | Third E command: packaged recovery | 80833 | pass / 0 |
| `tr01-wf16-lint-go-20261026-02` | `make lint CHECK=go` | 15909 | pass / 0 |
| `tr01-wf16-lint-boundary-20261026-02` | `make lint CHECK=backend-boundary` | 2677 | pass / 0 |
| `tr01-wf16-source-generated-20261026-02` | `make source-check CHECK=generated` | 15904 | pass / 0 |
| `tr01-wf16-source-schema-20261026-02` | `make source-check CHECK=schema` | 18294 | pass / 0 |
| `tr01-wf16-source-catalog-20261026-02` | `make source-check CHECK=catalog` | 1744 | pass / 0 |
| `tr01-wf16-source-identity-20261026-02` | `make source-check CHECK=identity` | 690 | pass / 0 |

The lifecycle union command was exactly:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.unit.testservices_lifecycle_contract,harness.browser.integration.postgres_template_preparation,harness.browser.integration.postgres_transaction_isolation,harness.browser.integration.postgres_owned_barrier_observation,harness.browser.integration.postgres_single_drop,harness.browser.integration.postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277,harness.browser.integration.reset_checkpoint_characterization,harness.browser.integration.managed_service_assertions_for_runtime_reset_rec_79940f98ab,harness.browser.integration.performance_fixture_snapshot_lifecycle
```

All 13 graph roots passed `make harness-observability-check
RESULTS_DIR=<exact-run-root> CHECK=all`, outer exit 0, `full_validity=1`.
Every retained cleanup outcome is `completed`; every retained-secret scan passed.
The transport root contains exactly 64 lifetime scopes and preserves the
16 serial / 32 four-way concurrent / 16 competing lifetimes. Recovery qualified
the packaged process boundary independently of reset and snapshot lifecycle.

`make lint-markdown` also passed, root
`.cartulary/test-results/tr01-wf16-markdown-20261026-01`, outer exit 0, 8000 ms;
summary `adhoc/lint-markdown/tool-run-summary.json`. Its legacy adhoc artifact is
separate from graph validity. Current lint globs omit this tracker: the changed
tracker was reviewed directly for structure, links, stable IDs, successor mapping,
phase statuses, source identities, scope and truthful limitations; `git diff
--check` passed. The specification diff only replaces misleading pool terminology
with quarantine/readmission wording. Specification/projection fidelity remains
human-owned; no executable check reads or certifies Markdown requirements.

#### Completed interfaces and assertion handoff

- Physical construction is only `StartOwned(ctx, StartOptions) (*OwnedCluster,
  error)`. Acquisition/settlement callbacks receive the owner. Retain a non-nil
  partial owner on failure for its cleanup receipt and recovery authority.
- Owned template preparation is explicit: `owner.PrepareTemplate(ctx, name)`;
  empty name uses existing canonical naming, service composition supplies its
  existing identity. `owner.Fixture()` becomes available after physical admission.
  Ordinary clones require the sealed template; scratch operations remain explicit.
- `Harness.Close`, physical fields, owned `StartOwnedWithLabels` and
  `StartOwnedWithOptions`, the interim `Harness.PrepareTemplate`, and PostgreSQL
  `StopShared` are removed without aliases. Borrowers use attachment-only
  `Start`/`StartShared`; `DetachShared()` has no context, I/O or shutdown error.
  Detach preserves admitted identity, ownership and generation quarantine.
- `TestDatabaseFixtureCleanup` succeeds the four standalone/attached clone/scratch
  cleanup tests; the prior matrix qualified before removal. The migrated-database
  witness succeeds the weak startup smoke. Capability and policy tests retain
  their row IDs and now test the actual paths. The assertion ledger in §7 and
  E-040 give the detailed mapping. Distinct failure/context witnesses remain.
- Direct callers and authored/generated routing are migrated atomically. No
  product API, schema/data migration, new policy language, new package hierarchy,
  S3 redesign, browser workflow change or performance-policy change is required.

#### Preservation, limits and operational handoff

The execution inventory comparison found 21 changed existing files and 12 new
pgtest files, all inside the authorized boundary; no baseline file was deleted.
The pre-existing tracker history inside `<details>` remains byte-identical.
Seventy generated-policy files had metadata/inode refreshes from public generation
with byte-identical contents; there were no metadata-only changes outside that
policy. Index entries remain byte-identical (digest
`2f8393a2bede4eadc41b7ccc1246f9daa19005bae7c33cc7c84b15d4b9155ce4`);
raw index stat-cache metadata refreshed, with no staging or commits. The final
7,338-file executable/input freeze was rechecked unchanged after validation.
Ignored build/results/dependency trees were not exhaustively fingerprinted.

Both finalizer invocations left RESULTS_DIR unset: no compatible successful full
warm-check root exists. Retained-run maintenance and warm-check scheduler/timing
checks were therefore skipped (`results-dir-not-provided` in
`unit-artifacts/finalize-summary.json`). Successful scoped graph roots were audited
explicitly as above. Full-owner, CI/release and performance acceptance are not
claimed; visual/browser workflows and preceding-handoff WF-33 remain outside scope.
These are bounded acceptance limits, not incomplete PostgreSQL remediation.

The final handoff is this tracker plus the working-tree diff and exact retained
roots. No further implementation workstream remains. Human review can inspect
projection fidelity and the bounded specification wording without treating tests
as specification authority. Rollback remains workstream-sized as described in §7;
preserve the original tracker edits and all failed evidence.

### Earlier planning-session scope and authority (superseded by E-036–E-041)

| Session / baseline | Inspected / touched | Result and supersession | Next action |
| --- | --- | --- | --- |
| 2026-10-26; `867e97ed49829d671c58576232999d1fc9a4e3fb` | Root/internal/tools instructions, tracker framework and adopted owners; sole edit this handoff | E-030/E-032/E-034: completed frontend iteration becomes historical; PostgreSQL plan is current; no owner amendment or implementation authorized | Future WF-12 starts with source and owner revalidation |

### Backend boundary

| Session / baseline | Inspected boundaries | Decision / evidence | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | pgtest, testservices and direct physical/process callers | E-031: split physical lifetime from fixture access; one template recipe; narrow helper owners retained | Qualify pre-move assertions, then WF-13 before WF-14 |

### Frontend boundary

Not applicable to the new iteration: no frontend change. The completed frontend
work and qualification limits are retained in the historical snapshot below.

### Contract and codegen

| Session / baseline | Inspected boundaries | Decision / evidence | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | Core/Harness owners, family rows, generated policy/topology | E-032/E-033: authored routing first, Make projections only; method count, wrappers and duplicate recipes are not requirements | Future WF-15 records exact old-to-successor assertion/selector mapping |

### Tests and harness

| Session / baseline | Inspection / check | Result / evidence | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | Assertions and routing inspected; no tests run | E-031/E-033; RB-005 remains open; historical passes do not clear it | Run V-015 only in a later implementation task |
| Document save | V-020 | E-035; lint scope and manual review limits explicitly recorded | No product verification for this update |

### Security and authorization

| Session / baseline | Inspected requirements | Decision / evidence | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | TLS/purpose admission, migration capability, owned cleanup and redaction boundaries | E-031/E-032: preserve C-011/C-013/C-014/C-015; no public injection or arbitrary-handle migration path | Keep real transport/ACL and capability-negative witnesses in future portfolio |

### Open risks and next session

| Session / baseline | Risk / blocker | Current disposition | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | RB-005/RB-006: no current execution evidence and predecessor gates | Planning complete for bounded seam; movement blocked, future tests explicitly specified | S-012 may enter after authorization; report failure causes before moving or broadening |
| Document save | Concurrent changes and command effects | E-034 pre-write identity check; E-035 final index/content comparison | Revalidate affected sources on any later resumption |

### Current evidence register

| ID | Source / command | Baseline / environment | Observation / artifact / applicability |
| --- | --- | --- | --- |
| E-030 | `git --no-optional-locks` status/index/tree inspection; external content/identity snapshot | Current `main` commit above; clean start | `/tmp/cartulary-pgfixture-plan-bf09ql7x/baseline.json`, `prior-handoff.md`; 7,637 entries; retained fingerprints support preservation, not behavioral coverage |
| E-031 | Source/assertion/import inspections in Section 2 | Same source commit; read-only | Mixed ownership, duplicate recipes, mutable hooks, resolver duplication and incidental assertions directly observed; 116 importer search is not a full product-caller audit |
| E-032 | Adopted status/Core precedence, Harness REQ-804/810/811/405/410 and §11.0; Core 01/04/05; repository instructions | Same source commit | C-011–C-016 distinguish required assurance from private test design; no primary-owner contradiction found in this seam |
| E-033 | Authored family/verification rows, Make surface, policy/topology, test-ops command recipes and lint path | Same source commit; no future command executed | Exact future A–E portfolio; generated outputs and invocation effects identified; V-020 lint exclusion known before execution |
| E-034 | Immediate pre-write canonical path/link/content recheck | Same commit; prior tracker digest from Section 1 | Matched E-030 across all 7,637 entries and both index fingerprints before writing; output remained regular/singly linked; `prewrite-comparison.json` retained beside baseline; historical snapshot retained with heading demotion |
| E-035 | V-020 and final preservation review | Documentation session, 2026-10-26; external cache/results | Lint passed at 20:33:49 UTC, exit 0, 7.570 seconds; no tests. Direct document review and `git --no-optional-locks diff --check` passed. Final comparison found only this tracker changed, with index and all other enumerated content/identities preserved; details below |
| E-042 | Post-implementation source/owner review; read-only final-manifest/summary/cleanup/scan inspection and source-freeze comparison | Review baseline and limits above; same E-041 executable source | Five required gaps and optional redundant attachment state; no product rerun or repair; follow-up chain WF-17–WF-19 |
| E-043 | V-024 documentation validation and preservation | Tracker-only review; external results/cache | Lint tool summary pass/0; outer Make exit not retained; direct review/diff/preservation passed; tracker-only change confirmed |
| E-044 | Closure implementation entry and preservation baseline | Authorized F-021–F-026; WF-17–WF-19; entry backup | Required decisions reconciled before executable edits |
| E-045 | WF-17 preparation recovery checkpoint | V-021 and exact roots above | F-021/F-024 qualified; terminal preparation and early ownership; predecessor saved |
| E-046 | WF-18 admission/binding/routing checkpoint | V-022 and exact caller/static roots above | F-022/F-023/F-025/F-026 qualified; predecessor saved; failed evidence preserved |
| E-047 | WF-19 final validation and handoff | V-023; 22 compatible roots; all native/audit exits 0 | 42 selected rows, 64 fresh transport lifetimes, completed cleanup, preserved source/history and bounded acceptance COMPLETE |

### Documentation save validation — E-035

The executed command from the repository root was:

```sh
env GIT_OPTIONAL_LOCKS=0 \
  CARTULARY_PREPARATION_POLICY=installed_only \
  CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-pgfixture-plan-bf09ql7x/results \
  CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-pgfixture-plan-bf09ql7x/static-analysis \
  CARTULARY_TEST_RUN_ID=tr01-postgres-plan-doc-20261026-01 \
  COREPACK_ENABLE_NETWORK=0 make lint-markdown
```

It passed with summary at
`/tmp/cartulary-pgfixture-plan-bf09ql7x/results/tr01-postgres-plan-doc-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The wrapper labels this documentation invocation `adhoc`; its summary records
zero tests. The pass covers the configured Markdown globs, which exclude
`docs/handoffs/testing-remediation-01.md`. Direct review of this document confirmed
the twelve ordered sections, balanced fences and consistent tables, seven local
links, 21 exact catalog row references, preserved historical text apart from
heading demotion, and coherent finding/contract/slice/validation references.
`document-review.json` records the structural/reference inspection; it is
documentation review, not executable harness or conformance evidence.

`final-comparison.json` beside the baseline records unchanged HEAD/branch,
index bytes and entries, path population, and content/file identities for every
enumerated entry outside this tracker. The tracker retained its file identity.
Ignored dependency/build trees were not exhaustively fingerprinted; the reviewed
invocation used existing dependencies, fixing disabled and external result/cache
paths. No install, format, generation, `agent-finalize`, product check, service
startup or test was run, as required for this document-only update. Final evidence
annotations were reviewed with the same document/baseline checks; they do not
change the linted file population.

### Completed iterations — historical snapshot

The following snapshot retains the preceding document's decisions, stable IDs,
commands, artifacts, failed-attempt diagnoses and completion limits. Heading
levels are demoted for nesting. Statements such as “current”, “authorized” and
“DONE”, and numbered section references, refer to that historical snapshot only.
It is evidence history, not authorization or a validation result for WF-12–WF-16.

<details>
<summary>Completed frontend and graph/scanner iterations (WF-00–WF-11)</summary>

### Testing Remediation 01: Frontend Test-Policy Simplification

#### 1. Scope and Source Posture

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

#### 2. Inspected Baseline Repository Inventory

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

#### 3. Module Boundary Diagnosis

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

#### 4. Public Contract and Behavior Freeze Map

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

#### 5. Coupling and Boundary Findings

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

##### Decision, migration and unresolved-risk ledger

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

#### 6. Refactor Workstreams

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

#### 7. Authorized Refactor Slice Plan

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

##### S-006 — requirement amendment and baseline

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

##### S-007 — one boundary implementation and a real transport seam

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

##### S-008 — qualify E2E reachability through Fallow

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

##### S-009 — behavior instead of selector-expression policing

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

##### S-010 — retire the E2E scanner and complete routing

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

##### Assertion successor ledger

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

#### 8. Validation Plan

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

##### Exact focused selections

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

##### Final portfolio and evidence limits

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

#### 9. Top-Level Work Tracker

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

#### 10. Session Handoff Log

##### Authorized implementation — current checkpoints

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

##### Historical document-only planning session — E-010–E-017

The following subsections retain that session’s observations and limits. Phrases
such as “current” and “next action” below refer to that earlier session; the
implementation checkpoints above supersede its pending work.

###### Scope and authority

| Time / baseline | State and evidence | Next action |
| --- | --- | --- |
| 2026-10-26; `971d6423c1bfc5f1f3ac6569fc524b596cf33f61` | E-010/E-011: user selected in-place update, frontend policies and simplification at source. Prior Plan Mode allowed inspection only. Current task authorizes saving that document-only plan. | Revalidate baseline before later implementation; do not inherit historical implementation permission |

###### Backend boundary

Not applicable: no backend, Go fixture, production assembly, storage or service
change is part of this iteration. Initial exploratory fixture reads selected no
backend workstream.

###### Frontend boundary

| Evidence | Inspected state / affected files | Next action |
| --- | --- | --- |
| E-012/E-013 | Two bespoke scanners duplicate established analysis and encode incidental constraints; Fallow roots hide a real successor-coverage gap | S-007/S-008 establish successors before S-009/S-010 removals |

###### Contract and codegen

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-013/E-015 | REQ-103 grammar is currently adopted; authored row/source inputs and generated outputs are identified; no owner/projection edits occurred | Propose/adopt only the concrete amendment; regenerate from authored inputs during separately authorized implementation |

###### Tests and harness

| Evidence | Commands/results and applicability | Next action |
| --- | --- | --- |
| E-014/E-015 | Assertions and command bindings inspected; read-only help and two task guides passed; no tests ran | Use V-008 as the future baseline, not historical passes |
| E-016 | `make lint-markdown` passed, outer exit 0; installed-only preparation and external result/cache directories. Handoff structure, links and diff reviewed separately because lint globs exclude this file. | No tests, installation, generation or finalizer run; detailed command and coverage below |

###### Security and authorization

| Evidence | Required distinction | Next action |
| --- | --- | --- |
| E-014 | Preserve actual HTTP validation, same-origin boundaries and private diagnostics; current wrapper-name guard is incomplete | Enforce the module that actually performs raw execution; keep success/error/observed-header witnesses |

###### Open risks and next session

| Evidence | Current state / limits | Next action |
| --- | --- | --- |
| E-017 | All 7,638 tracked/non-ignored paths compared: only this handoff content changed; no added/removed paths or file-identity changes; HEAD and index unchanged | Later actor starts with baseline/adoption/characterization revalidation, not broad rediscovery |

###### Historical planning evidence register

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

###### Historical document validation — E-016/E-017

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

##### Historical graph/scanner iteration — E-001–E-009

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

##### Historical final acceptance portfolio

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

#### 11. Open Questions and Blockers

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

#### 12. Binary Completion Criteria

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

</details>





## 11. Open Questions and Blockers

| ID | Gate | Resolution |
| --- | --- | --- |
| RB-017 | Current pre-move characterization | DONE; WF-41 exact results in E-114 |
| RB-018 | Qualified behavioral successors before retirement | DONE; E-118–123 and final-source E-124 |

No unresolved scope, implementation or validation blocker remains for this plan.
Human projection review and historical unrelated WF-33 remain separate.

## 12. Binary Completion Criteria

- [x] All included workstreams exited and were recorded before successors began.
- [x] F-054–060 and D-016–029 have qualified final dispositions in E-124.
- [x] Shared parser is the only import syntax implementation for these consumers.
- [x] Authorization/effect/history/provider guarantees have executed successors.
- [x] SQL policies retain exact-file permissions and reject unresolved closed-scope access.
- [x] No filenames/counts/private spellings replace retired incidental assertions.
- [x] Authored routing and generated projections agree with actual execution.
- [x] Finalizer precedes final-source qualification; maintenance skip is explicit.
- [x] Final roots, native exits, cleanup validity and residual review limits are recorded.

Implementation, selected validation and handoff are complete. No whole-project or
release acceptance is claimed. Changes remain available for review in this checkout.
