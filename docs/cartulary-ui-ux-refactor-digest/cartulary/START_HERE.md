# Cartulary UI/UX Refactor Overlay

## Purpose

This repository-local overlay preserves the reviewed UI/UX remediation baseline
and prepares later, separately authorized slices. It does not authorize
implementation, behavior changes, a re-theme, or a product redesign. The
bundled upstream material is useful as an offline audit workflow, searchable
checklist, and source of review questions. It is not Cartulary authority.

Run commands from the Cartulary repository root. Use [REPO_MAP.tsv](REPO_MAP.tsv) as the compact localization snapshot, then
inspect the selected seam and revalidate changed facts against authored inputs.

## Authority order

Apply this order before every later design or implementation decision:

1. Adopted subsystem NLSpecs for their named scope, using the exact current
   paths and statuses in `REPO_MAP.tsv`.
2. `docs/spec/00_document_set_status_and_precedence.md` through
   `docs/spec/04_security_deployment_and_conformance.md` for current
   implementation-conformance behavior.
3. `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` only for
   claim-bearing timed or fixture-sensitive publication.
4. `docs/design.md` for observable design behavior inside its declared scope.
5. `docs/domain.md` for repository vocabulary and owner navigation inside its
   declared scope.
6. `docs/handoffs/cartulary_modular_refactor_planning_framework.md` and
   implementation guides as subordinate planning or implementation support.
7. Current code and tests as evidence of current state.
8. This localized overlay.
9. Bundled upstream material as advisory evidence only.

If owner documents conflict, stop with `BLOCKED: owner contradiction`; do not
choose silently. Existing code is not automatically required behavior.
Separate an authorized normative correction from behavior-preserving structural
movement.

Product tests, generators, runtime metadata, conformance checks, and release
evidence must not read, stat, hash, or otherwise depend on documentation.
Typed projections and test routing remain owned outside `docs/` as described
by `AGENTS.md`. Manual package integrity and Markdown lint are documentation
maintenance, not product verification.

## Verified repository boundaries

Use [REPO_MAP.tsv](REPO_MAP.tsv) for the reviewed discovery snapshot and
[meta/localization.json](../meta/localization.json) for stack versions and the
actual localization baseline. Workspace patterns are `apps/web` and `packages/*`.
Inspect the authored manifests again when the repository advances.

Three separate questions select the correct boundary:

| Question | Source of truth | Interpretation |
| --- | --- | --- |
| Who owns behavior? | Exact adopted Core/subsystem clauses, with bounded design direction; navigate through [OWNER_MAP.tsv](OWNER_MAP.tsv). | Neither source placement nor test coverage creates requirements. |
| Where does source belong? | `tools/frontend_source_ownership.json` and `tools/frontend_import_boundaries.json`. | Source `web.app` and `web.network_flow` are distinct from verification `web.application` and `web.networkflow`. |
| Which verification runs? | `contracts/verification/registry.json`, `tools/test_catalog_owner.json`, and `tools/test_families`. | Use active catalog IDs with the current task guide; routing does not certify specification completeness. |

Start with the [frontend source overview](../../../apps/web/src/README.md)
and follow its child guides. App owns shell/account/incident composition;
Collaboration owns incident WebSocket lifetime and typed event publication;
Workbook and Network Analysis retain their feature and operation semantics.
Current verification navigation includes `web.architecture`, `web.collaboration`,
and `web.networkflow`. Do not duplicate local guides' file inventories here.

`packages/grid-adapter` is the sole direct `react-data-grid` integration boundary.
`packages/view-contracts` adapts authored view-schema projections and generated
contracts. `packages/ui-contracts` supplies stable selectors and generated token
and presentation facades. Shared semantic test helpers belong in
`packages/test-utils`; browser choreography stays in `apps/web/e2e/support`.
Use `tools/generated_artifact_policy.json` and the authored generation inputs
listed in `REPO_MAP.tsv` to distinguish editable owners from generated output.

The verification ID `package.ui` routes `packages/ui-contracts`; a functional
`packages/ui` is absent. Design §3.11 owns semantic icon identities, while source
uses `lucide-react` directly and has no standalone implementation registry.
`harness.visual` owns fixture contracts but is not an active task-guide owner;
its public fixture entry point is `make browser-e2e-visual`. Report Composition
is adopted/current at 1.2.0; the Reference Pack subsystem NLSpec remains draft.
These qualifications are not defects to repair by inventing implementations.

## Product constraints

Cartulary is a workbook-native incident workspace. Preserve spreadsheet speed
at the view layer while retaining disciplined relational, authorization,
history, and evidence behavior underneath.

- The grid is the protagonist.
- Capture remains grid-first, compact, direct, keyboard-complete, and tolerant
  of incomplete facts.
- Conflict, evidence, validation, save, and recovery feedback appears where it
  changes the user's action.
- The inspector augments the grid; it does not replace it or become a detached
  form workflow.
- The visual language is dense graphite, calm, precise, inspectable, and
  durable.
- Warm accent is scarce and semantic.
- State never relies on color alone.
- Required system views stay reachable inside the workbook shell.
- Vendor coordinates, SQL/projection names, component names, CSS classes, and
  package-specific icon names never become product concepts or test authority.

## Classification rule

Classify every material upstream recommendation before use:

- `ADOPT`: compatible principle that can be applied without changing a
  Cartulary-owned contract.
- `ADAPT`: useful concern whose prescription must be translated through
  Cartulary's desktop workbook, token, interaction, or owner contracts.
- `REJECT`: conflicts with current authority, scope, density, behavior, or
  visual direction.

Do not weaken or bypass this classification. Record it in later change notes
when it materially affects a decision. The baseline matrix is
`docs/cartulary-ui-ux-refactor-digest/cartulary/rules.tsv`.

## Regression baseline for later verification

This is the canonical regression summary for the localized overlay. Exact owner
clauses remain authoritative; use [OWNER_MAP.tsv](OWNER_MAP.tsv) to navigate them.
Historical terminal handoffs establish implementation context and limitations,
not a current product-test pass. The [original remediation record](../../archive/cartulary-ui-ux-remediation-handoff.md)
is archived. Later workflow records are under `docs/handoffs/ui-ux/`.
Reconcile the selected family with current code before calling a concern a defect.

### Established workbook and presentation behavior

| Family | Owner-grounded baseline |
| --- | --- |
| Density | One selection reaches row/header height, padding, typography, gutters, saved/draft/read-only content, and full-cell editor geometry through shared tokens; preserve the declared clear/null surface default. |
| Creation discovery | Use `create_capable`, interaction authority, declared create-writable fields/inputs, and `inline_create` policy for the paths it governs. Preserve projected ordinary minima, Evidence/Indicator owner validation, and declared contextual alternatives. |
| Responsive shell | Validated CSS lengths and token-backed inline-size thresholds govern chrome independently of block size; `innerWidth`/`innerHeight` replace absent `visualViewport`; inspector geometry and ARIA state agree. Keep grid/inspector scrolling inside the shell and navigation reachable. |
| Inspector dispatch | Dispatch current features exactly once by stable semantic route identity. Keep owner-required disabled rendering, review invalidation, and unknown-additive omission behavior. |
| Editing and conflict | Preserve keyboard/paste/edit/commit/cancel/Escape behavior, raw retained drafts, and local error/conflict feedback. Saved values remain distinct from unresolved local work. |
| Evidence, accessibility, and virtualization | Preserve owner-specific Evidence access/lifecycle states, semantic names, non-color cues, visible focus, live regions, reduced motion, stable virtualized identities, and scroll continuity. Long content and supported text/zoom settings must remain usable within owned geometry. Visual evidence uses relevant production renderers and reviewed fixture parameters. |

### Current workflow families

| Family | Required review questions and implementation navigation | Historical terminal evidence |
| --- | --- | --- |
| Contextual creation and retained authoring | Core 01 create/inspector contracts, Core 02 §10, Core 03 §§2.3A,16.4, design §§7.3,12. Review explicit source identity/replacement/clearing, retained raw fields, detachment, declared entry paths, duplicate activation, and current authority. Follow the [Coordination](../../../apps/web/src/workbook/features/coordination/README.md) and [Notes](../../../apps/web/src/workbook/features/notes/README.md) source guides. | [Coordination creation](../../handoffs/ui-ux/workbook-contextual-coordination-create-refactor-handoff.md), [Task/Decision creation](../../handoffs/ui-ux/workbook-contextual-task-decision-create-refactor-handoff.md), [linked Notes](../../handoffs/ui-ux/workbook-linked-note-create-refactor-handoff.md). |
| Exact uncertain replay, acknowledgement, and refresh recovery | Core 01 §3.3.5, Core 03 §§3-4, design §10. Review captured request bytes/identity, duplicate admission, late outcomes, acknowledgement beyond presentation lifetime, and failed refresh after an accepted write. A refresh retry does not resend that write. Follow [runtime](../../../apps/web/src/workbook/runtime/README.md), [feature](../../../apps/web/src/workbook/features/README.md), and [Timeline](../../../apps/web/src/workbook/timeline/README.md) guides; retain source-owner semantics. | [Timeline related Evidence](../../handoffs/ui-ux/workbook-timeline-related-evidence-refactor-handoff.md), [Assessment authoring](../../handoffs/ui-ux/workbook-assessment-authoring-refactor-handoff.md), and the contextual-creation records above. |
| History browsing and corrective actions | Core 01 §3.3.4.2, Core 02 §15, Core 03 §10, Core 04 authorization, design §§7.3,10. Review server-owned continuation/order, retained accepted history, current eligibility, captured rollback/restore, accessible lookup/recovery, and acknowledgement independent of refresh. Follow the [History guide](../../../apps/web/src/workbook/history/README.md). | [History browsing](../../handoffs/ui-ux/workbook-record-history-browsing-refactor-handoff.md) and [History recovery](../../handoffs/ui-ux/workbook-record-history-recovery-refactor-handoff.md). The recovery record includes historical baseline verification limitations; they are neither fresh failures nor passes. |
| Incident versus session authorization loss | Core 01 §3.3.6.2, Core 03 REQ-03-299/100, Core 04 §§1-2, design §12.8. Review scoped incident exit, canonical revocation reasons, owner-required same-account local-work retention, uncertainty handling, and clearing before account replacement. Follow [App](../../../apps/web/src/app/README.md), [Collaboration](../../../apps/web/src/collaboration/README.md), and [Workbook lifecycle](../../../apps/web/src/workbook/lifecycle/README.md) guides. | [Incident/session revocation](../../handoffs/ui-ux/incident-session-revocation-remediation-handoff.md). |
| Query-data and interaction states | Core 03 REQ-03-286/299, design §§7-8,10.6,10.8,12.8. Review data-state/interaction-mode composition, loading versus refresh, successful/filtered empty, unavailable/stale/failure, read-only/closed states, and permission loss. Retain authorized stale content where required; suppress protected material on access loss. Keep local feedback and semantic focus/scroll continuity. Inspect `packages/grid-adapter` and its Workbook/Network Analysis producers. | [Grid operational states](../../handoffs/ui-ux/workbook-grid-operational-state-plane-refactor-handoff.md). |
| Network Analysis imports, tables, pagination, and graphs | Network Flow §§5.4-5.5,7-8,10,13-14,18-19,28; delegated Graph Projection/Extensions owners; design §13.1. Review captured import/table operations, committed pages versus pending destinations, late responses, source identity, retained graph results, bounded observation, authorization loss, and accessible recovery. Follow the [Network Analysis source guide](../../../apps/web/src/networkFlow/README.md); extension resources remain distinct from Core records and saved views. | [Import recovery](../../handoffs/ui-ux/network-analysis-import-recovery-refactor-handoff.md), [table lifecycle](../../handoffs/ui-ux/network-analysis-table-lifecycle-refactor-handoff.md), [pagination recovery](../../handoffs/ui-ux/network-analysis-pagination-recovery-refactor-handoff.md), [saved graph lifecycle](../../handoffs/ui-ux/network-analysis-saved-graph-lifecycle-refactor-handoff.md). |

Apply these questions only where the selected owner requires them. Retained
browser work does not imply reload, cross-tab, or durable local persistence.
Do not infer retry rules, workflow engines, new extension claims, or future-profile
behavior from the similarity of existing implementations. Historical intermediate
failures are not current defects; historical passes are not current execution.

## Refactor selection rubric

This rubric implements the user's structural-design principles. Those principles
are selection guidance, not upstream recommendations or new product authority.
Prefer cohesive ownership, small interfaces that hide meaningful decisions,
security boundaries that remain explicit, and designs that admit future phases
without spreading feature conditionals or duplicate state.

Before selecting a future authorized slice, record each gap using this rubric:

| Decision | Required review evidence |
| --- | --- |
| Observed weakness | A current observation and affected user action; classify confirmed defect, structural weakness, or hypothesis. Historical code and tests do not establish required behavior. |
| Remediation and change areas | State the proposed fix and whether it belongs in adopted specifications, projections, implementation, tests, documentation, or several areas. Separate authorized behavior corrections from structural movement. |
| Responsible owner and boundary | Name exact governing clauses, the common decision hidden by the interface, source ownership, and independent verification routing. Surface/domain semantics remain with their adopted owners. |
| Rationale and long-term benefit | Explain how the design improves stability, maintainability, conceptual clarity, security, extensibility, cohesion, or necessary coupling. Similar-looking forms or large files alone do not justify shared machinery. |
| Future extension path | Explain how a plausible next surface or workflow uses this boundary without duplicate state or scattered feature-specific branches. Do not implement speculative extensions. |
| Capability value | Explain why each carried-forward feature materially improves future use, testability, or maintenance. Existing implementation is not itself a reason to retain it. |
| Retirement | Name redundant paths, adapters, duplicate state owners, and parallel implementations; migrate their callers and remove them together. Each retained compatibility layer needs a concrete reason and an explicit retention or retirement decision. |
| Compatibility and migration | Identify adopted obligations and actual supported consumers. State interface/data impact and migration or cutover steps; say explicitly when none are needed. Do not create aliases or indefinite dual implementations for hypothetical consumers. |
| Risk of leaving the gap | Identify the concrete correctness, security, maintenance, or expansion cost if the weakness remains; distinguish demonstrated risks from hypotheses. |
| Validation and delivery | Define observable pass/fail outcomes, focused owner verification, dependency order, workstream exits, rollout/rollback, limitations, and the final handoff. |

Shared machinery must encapsulate a real common decision while retaining
source-owner semantics. A helper count, framework choice, or visual similarity
is not evidence of a good boundary. Compare simplification and retirement with
retention; carry forward only useful capabilities while honoring adopted
requirements and explicitly supported consumers.

## Workflow for a future authorized product slice

1. Read `AGENTS.md`; capture branch, commit, dirty state, allowed paths, and
   authorized behavior changes. Revalidate the selected seam against authored
   stack, source/import, generated-artifact, and verification inputs.
2. Follow the relevant baseline family, current source guides, and exact owner
   clauses. Characterize risky owner-required behavior; distinguish defects,
   structural weaknesses, and hypotheses. An unresolved adopted-owner conflict
   is `BLOCKED: owner contradiction` and prevents dependent work.
3. Apply the selection rubric. Choose one coherent boundary and state its
   dependencies and binary exit. Preserve semantic identity, authorization,
   retained operation lifetime, and source-owner recovery responsibilities.
4. Use [QUERY_RECIPES.md](QUERY_RECIPES.md) for targeted offline evidence and
   [rules.tsv](rules.tsv) for material `ADOPT`, `ADAPT`, or `REJECT` decisions.
   Never run upstream `--design-system`, `--persist`, or `--force` against
   Cartulary. These can create a second design authority.
5. Implement only the authorized slice. Change authored inputs before generating
   their projections; never hand-edit generated roots. Keep direct vendor imports
   inside the Grid Adapter and test choreography outside product runtime.
6. Validate through the owner-selected loop below. Update the slice's controlling
   tracker with results and its completed exit before beginning its dependent.
7. Complete [acceptance.tsv](acceptance.tsv) as a human review assessment in the
   slice handoff. Record owner clauses, changed behavior/paths, compatibility,
   removed or retained alternatives, commands/results/artifacts, material rule
   dispositions, limitations, rollback, and next action.

### Verification selection

Run public repository commands from the repository root. `make help` and
`make help-all` own the current target inventory. Choose a verified active owner
from the catalog navigation in `REPO_MAP.tsv`, then use:

```bash
make task-guide ROLE=module-author OWNER=<verified-owner-id>
```

Run its recommended focused `make test-slice` or
`make service-backed-test-slice` command with the selected owner and rows.
Broaden only for the changed behavior, boundary, or unresolved risk. Readiness
and generated-artifact checks are selected through the current public surface,
not copied target lists. For visual golden edits, follow
`docs/guides/cartulary_visual_golden_maintenance.md`.

Run `make agent-finalize` before broader end-of-run verification. Supply
`RESULTS_DIR` only for an actual successful full warm check accepted by the
repository; otherwise record retained-run maintenance as skipped because it
was unset. Report each failing target, relevant summary/run root, relation to
this change, and affected dependency. Do not hide unrelated failures or absorb
them into an unrelated slice.

Visual and accessibility artifacts remain implementation-support evidence unless
an applicable owner establishes a conformance or Core 05 publication boundary.
A digest-only update follows the documentation-maintenance gates in the
[controlling localization handoff](../../handoffs/ui-ux/ui-ux-refactor-digest-update-handoff.md),
not the product-test workflow above. Do not mark product acceptance as passed
because advisory criteria were edited.

### Completion semantics

Applicable acceptance rows require `PASS` with evidence. Use `N/A` only with a
specific scope/owner rationale. An applicable `BLOCKED` row prevents completion;
labeling a blocker or deferring required verification does not satisfy the row.
Only `PASS` and justified `N/A` may appear in a completed product-slice assessment.
No new status column or schema is added to the advisory TSV; assessment results
belong in the corresponding handoff.

## Stable verification identifiers

Prefer owner-defined fixture IDs, `view_schema_id`, `record_id`, `field_key`,
semantic icon IDs, and owner-defined capability/state enums. Do not assert
against literal specification text, line numbers, hashes, formatting, document
layout, visible row numbers, incidental DOM hierarchy, CSS classes, component
names, SQL/projection names, vendor coordinates, or package-specific icon
names.

Machine-testable facts derived from specifications belong in versioned,
owner-governed machine-readable artifacts outside documentation directories.
