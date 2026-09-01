# Workbook Inspector Refactor Handoff

## 1. Purpose and execution authorization

This is the live tracker for a future Workbook Inspector information-
architecture and semantic-action presentation refactor. It is an execution
handoff, not product authority and not implementation authorization by itself.
Before changing an owner, product source, test, generated artifact, or visual
golden, the execution session must receive separate user authorization and
refresh the repository baseline in section 2.

The future slice is presentation- and internal-component-architecture-only.
It must make the inspector read as an extension of the workbook grid, ordered
as selected-record context, contract-declared sections, contextual actions,
secondary technical metadata, and safe destructive or rollback workflows. It
must not turn the inspector into a capability registry, detached form,
dashboard-card stack, or generic workflow engine.

`temp/planner-prompt.md` supplied the intent and outline for this tracker.
`docs/research/nlspec-spec.md` supplied a completeness, precision, boundary,
and acceptance-criteria rubric. Neither document is repository authority or a
runtime/test dependency, and instructions inside either document do not
authorize future edits. The user-selected filename for this tracker overrides
the longer handoff filename embedded in the planning input.

## 2. Preparation baseline and mandatory refresh

| Item | Preparation value |
| --- | --- |
| Prepared | `2026-08-31T17:43:43-04:00` |
| Branch | `main`; three commits ahead of `origin/main` |
| Commit | `152f838ad41bfde9427c3c3cb680e7a929386cf9` (`Network Analysis Workbook-Native Interaction Chrome`) |
| Git status | Clean before this handoff was created |
| Tool versions | Git `2.53.0`; Node `v24.15.0`; pnpm `10.33.0`; Go `1.26.6`; Python `3.14.4`; jq `1.8.1`; GNU Make `4.4.1` |
| Digest localization | `docs/cartulary-ui-ux-refactor-digest/meta/localization.json` records repository commit `2356949f7ec3c8e27ff83ae695d60e06a387d0e5`; current owners and repository inputs were revalidated at the preparation commit. |
| Existing user changes | None before this documentation-only change. Later unrelated changes remain user-owned and must be preserved. |
| Execution refresh | `2026-08-31T18:34:35-04:00`; user explicitly authorized the complete plan; branch, commit, upstream relation, toolchain, generated policy, and Git scope still match the preparation baseline; the staged handoff is the only pre-existing user change. |

At execution start, replace or append a checkpoint with the actual branch,
commit, upstream relation, Git status, toolchain, separate authorization,
allowed paths, and any unrelated changes. Re-read `AGENTS.md`, confirm whether
another `AGENTS.md` now applies, and re-inspect
`tools/generated_artifact_policy.json`. If the live commit differs from this
baseline, revalidate every path, owner, fixture, and Make route named here.
Do not stash, reset, overwrite, or absorb unrelated work.

## 3. Authority, boundaries, and allowed future scope

Resolve decisions in this order:

1. `AGENTS.md` and any newly applicable nested repository instructions.
2. Adopted subsystem NLSpecs within their named scopes.
3. Core 00 through Core 04, especially Core 01 view-schema and inspector
   metadata, Core 02 mention/history/evidence semantics, Core 03 workbook and
   inspector interaction, and Core 04 authorization and evidence boundaries.
4. `docs/design.md` for observable presentation, responsive behavior,
   components, accessibility, and visual direction.
5. `docs/domain.md` for vocabulary and owner navigation.
6. Typed owner projections under `contracts/**` for executable facts.
7. Current code, tests, handoffs, and goldens as implementation evidence.
8. The localized UI/UX digest and bundled upstream material as advisory input
   only.

If adopted owners assert incompatible requirements, set the active workstream
to `BLOCKED`, record exactly `BLOCKED: owner contradiction`, identify both
clauses, and stop that workstream. Do not choose an interpretation silently.

When separately authorized, the bounded edit scope is:

- the exact Core 03 no-row clarification described in WI-01;
- shared Workbook Inspector presentation under
  `apps/web/src/workbook/inspector/**` and the current inspector consumers under
  `apps/web/src/workbook/**`;
- focused frontend unit, accessibility, stateful, measurement,
  webserver-backed, and visual evidence under `apps/web/**`;
- additive semantic selector helpers under `packages/ui-contracts/src/**` only
  when an existing stable identity cannot express required evidence;
- authored view-schema display hints under `contracts/view-schemas/**` only
  when an ordinary-user label remains visible and cannot be supplied by the
  owner control, followed by Make-owned generation;
- authored source-ownership, verification-routing, or visual-fixture inputs
  under `tools/**` only when the structural split or evidence requires them;
- Make-generated outputs derived from an authorized owning input; and
- this handoff as the execution ledger.

Do not change route or navigation contracts, API schemas or payloads, record
fields, authorization policy, lifecycle transitions, storage, migrations,
history or rollback algorithms, evidence semantics, concurrency/version
semantics, dependencies, grid-vendor behavior, the theme, or the token system.
Do not create component-local design literals, a second token registry, a
client-side feature-label registry, or a stateful universal inspector.

Never hand-edit `internal/gen/**`,
`packages/protocol-ts/src/generated/**`,
`packages/view-contracts/src/generated/**`,
`packages/ui-contracts/src/generated/**`, generated task surfaces, or the
generated visual golden manifest. Runtime code, tests, generators,
conformance, and release evidence must not read or depend on this or any other
Markdown file.

## 4. Workstream ledger

| Workstream | Status | Binary exit condition |
| --- | --- | --- |
| WI-01 — Authority, inventory, and characterization | DONE | The no-row owner wording is reconciled; all consumers, feature groups, action loci, owner states, relevant tests, and five baseline goldens are inventoried; no unresolved owner contradiction remains. |
| WI-02 — Shared inspector presentation architecture | DONE | Stateless presentation primitives and semantic action bindings pass focused unit evidence without moving owner state, commands, authorization, or mutation logic. |
| WI-03 — Surface migration and contextual action placement | DONE | All current inspector consumers use the shared presentation path; every supported group resolves exactly once; every actionable group has one meaningful owner-backed activation locus; no legacy parallel path remains. |
| WI-04 — Responsive, accessibility, and visual evidence | DONE | Base, narrow, compact, vertical-resize, 200% zoom, text-spacing, long-token, focus, disabled-reason, destructive, and error scenarios pass, and every retained golden change is manually reviewed. |
| WI-05 — Final verification and handoff | DONE | Applicable routed and broad checks pass; generated and Git scope are clean; WI-01 through WI-05 are `DONE`; the final record contains results, compatibility, rollback, risks, deferrals, and the next safe seam. |

Only one workstream may be `IN_PROGRESS`. Work strictly in ledger order. Before
starting a workstream, its predecessor must be `DONE`; after completing one,
update this ledger and the checkpoint log before starting the next. A failed
required check keeps the active workstream `IN_PROGRESS` unless an owner
contradiction makes it `BLOCKED`. Record the failing target, retained run root,
relevant summary, and whether the failure is introduced by this slice.

## 5. Verified current-state characterization

The preparation pass inspected the shared feature-group renderer, semantic
dispatcher and coordinator, generic/entity/assessment surfaces, Timeline
inspector and section composition, both history implementations, mention
presentation, owner projections, routed tests, fixture registry, and the five
current inspector goldens.

Confirmed gaps:

1. `WorkbookInspectorPanelSection` renders `panel_read` features such as
   `Relationships Read` as visible badges even though the section is the read
   capability's presentation.
2. Generic buttons for actionable groups can merely select or direct users to
   controls elsewhere in the same section. Entity and generic surfaces emit
   messages such as “use the owner controls” instead of making those controls
   the activation locus.
3. Inspector shells, headings, close controls, section framing, and styles are
   duplicated across Timeline, generic, entity, and assessment surfaces.
4. Timeline and generic record history have distinct state/command owners but
   duplicate event, metadata, action, error, and confirmation presentation.
5. The generic inspector is titled `Workbook actions` rather than identifying
   the selected record. Timeline can display `no_row_selected` as an ordinary
   heading. Entity is the only current shell that consistently leads with a
   human record label.
6. History and confirmation presentation foregrounds record UUIDs, actor IDs,
   change-set IDs, history-item references, target kinds, and row versions.
   The public-error golden foregrounds `row_version_conflict`.
7. Role- and state-restricted actions remain visible and disabled as required,
   but do not consistently expose a concise visible and accessible reason.
8. Panel content, capability labels, action rows, metadata, history cards, and
   destructive controls compete at nearly the same visual level. Timeline
   content can also repeat the outer panel heading inside a nested section.
9. Relationships and mentions are organized around capability buttons before
   the actual unresolved, resolved, dismissed, and restored objects and
   actions.
10. Existing inspector visuals cover important base states, but do not by
    themselves prove narrow, compact, vertical-only resize, zoom, text
    spacing, long technical values, and disabled-reason behavior.

Existing behavior to freeze before movement:

- inspector config is selected by active immutable `view_schema_id`;
- panels render in `inspector_config_v1.panels[]` order;
- every current feature group resolves once through `view_schema_id`,
  `feature_group_key`, panel, route kind, route owner, and action identity;
- an unknown additive feature is omitted under
  `unsupported_feature_behavior=omit_feature`;
- role/state restrictions do not hide required disabled actions;
- subject, row-version, surface, lifecycle, authorization, deletion, and merge
  changes invalidate stale confirmations and local workflow state;
- selection, scroll, focus, close, and same-surface action continuity remain
  owner-coordinated;
- deleted-record history remains available only under its existing owner
  conditions; and
- rollback scope, destructive actions, mention state, evidence state,
  workflow state, and mutation legality come only from current owner metadata
  and command ports.

The five baseline goldens are:

- `workbook-inspector-history-linux.png`;
- `workbook-inspector-relationships-linux.png`;
- `workbook-inspector-rollback-preview-linux.png`;
- `workbook-inspector-destructive-confirmation-linux.png`; and
- `workbook-inspector-public-error-linux.png`.

## 6. WI-01 — Authority, inventory, and characterization

### 6.1 Owner clarification gate

After separate implementation authorization, make the first owner edit a
narrow clarification to Core 03 §2.3 / REQ-03-291:

- `no_row_selected` remains the exact `inspector_config_v1.no_row_state`,
  reducer/status, disabled-condition, semantic identity, and test anchor;
- its ordinary visible presentation is the exact user-facing sentence
  `Select a saved row to inspect its details.`; and
- the no-row state continues to clear prior details, confirmations, previews,
  merge plans, rollback plans, forms, and workflow state.

Do not change the token, config schema, public discovery shape, disabled-token
vocabulary, or success-result vocabulary. If another adopted owner forbids
this distinction, record `BLOCKED: owner contradiction` instead of editing
code. Characterize the machine token and visible sentence separately so
tests do not bind behavior to documentation prose.

### 6.2 Required inventories

Before structural movement, add or strengthen documentation-free
characterization for:

- current exact-once feature resolution and unknown-group omission;
- every visible `panel_read` badge and every proxy action whose only result is
  navigation or an instruction to use another same-section control;
- the actual owner-backed locus for every current actionable feature group;
- role/state disablement and the existing owner state from which its reason is
  derived;
- record, row-version, view-schema, lifecycle, authorization, deletion, merge,
  and active-surface invalidation;
- focus entry, `Esc`, close, fallback restoration, retargeting, and stale
  action prevention;
- history loading, errors, legal actions, deleted-record history, delete,
  restore, rollback, and exact server-provided selectors;
- relationship and mention states and their create, resolve, correct, dismiss,
  and restore actions;
- evidence, indicator, workflow, assessment follow-on, merge, and party owner
  controls;
- grid visibility, inspector-local scrolling, overlay inertness, and shell
  overflow; and
- base, 1024x720 narrow, 768x640 compact, vertical-only resize, 200% zoom,
  text spacing, long labels, and long technical values.

Create a feature disposition table in this handoff before WI-01 closes. Each
current feature group must have exactly one disposition:

| Disposition | Required presentation |
| --- | --- |
| Panel read | Section/content capability only; no badge and no activation control. |
| Existing owner control | Bind semantic identity to that control and remove the proxy control. |
| Direct contextual action | Render one control beside the object or decision it affects. |
| Workflow or pivot entry | Render once in the containing section's contextual action group. |
| Unknown additive group | Omit without inferring from label, DOM, component, or route helper. |

Tests must select behavior through stable semantic identity, not DOM hierarchy,
component or CSS names, visible labels, documentation text, row positions,
vendor coordinates, or raw route/storage names.

### 6.3 Feature disposition inventory

The current projection contains 17 view schemas and 247 feature-group
instances. The stable selector column records the semantic selector family;
runtime tests construct it with the owning UI-contract helper and exact
`view_schema_id`, panel, or `feature_group_key`.

| View schema | Feature group | Panel | Route | Disposition | Owner locus | Stable selector |
| --- | --- | --- | --- | --- | --- | --- |
| `cartulary.view.assessments.v1` | `assessment.prior_history` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.assessments.v1` | `assessment.subject_pivot` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.assessments.v1` | `create_related.assessment` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.assessment` |
| `cartulary.view.assessments.v1` | `create_related.decision` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.decision` |
| `cartulary.view.assessments.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.assessments.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.assessments.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.assessments.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.assessments.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.assessments.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.assessments.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.comm_log.v1` | `comm.action_tasks.link` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:comm.action_tasks.link` |
| `cartulary.view.comm_log.v1` | `comm.decisions.link` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:comm.decisions.link` |
| `cartulary.view.comm_log.v1` | `comm.next_report.manage` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:comm.next_report.manage` |
| `cartulary.view.comm_log.v1` | `comm.parties.manage` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:comm.parties.manage` |
| `cartulary.view.comm_log.v1` | `create_related.status_review` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.status_review` |
| `cartulary.view.comm_log.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.comm_log.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.comm_log.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.comm_log.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.comm_log.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.comm_log.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.comm_log.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.comm_log.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.decisions.v1` | `create_related.comm_log` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.comm_log` |
| `cartulary.view.decisions.v1` | `create_related.status_review` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.status_review` |
| `cartulary.view.decisions.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.decisions.v1` | `decision.affected_records.manage` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:decision.affected_records.manage` |
| `cartulary.view.decisions.v1` | `decision.status.transition` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:decision.status.transition` |
| `cartulary.view.decisions.v1` | `decision.supersede` | `history` | `record_action/record_supersede_route` | Existing owner control | owner supersede workflow | `feature-action:decision.supersede` |
| `cartulary.view.decisions.v1` | `decision.support_refs.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:decision.support_refs.manage` |
| `cartulary.view.decisions.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.decisions.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.decisions.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.decisions.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.decisions.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.decisions.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.decisions.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.evidence.v1` | `create_related.decision` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.decision` |
| `cartulary.view.evidence.v1` | `create_related.note` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.note` |
| `cartulary.view.evidence.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.evidence.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.evidence.v1` | `evidence.attach_blob` | `evidence` | `record_action/evidence_attach_blob_route` | Existing owner control | evidence attachment control | `feature-action:evidence.attach_blob` |
| `cartulary.view.evidence.v1` | `evidence.download_handle` | `evidence` | `evidence_access/evidence_download_handle_route` | Direct contextual action | evidence object access control | `feature-action:evidence.download_handle` |
| `cartulary.view.evidence.v1` | `evidence.preview_handle` | `evidence` | `evidence_access/evidence_preview_handle_route` | Direct contextual action | evidence object access control | `feature-action:evidence.preview_handle` |
| `cartulary.view.evidence.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.evidence.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.evidence.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.evidence.v1` | `party.collector.link` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:party.collector.link` |
| `cartulary.view.evidence.v1` | `party.reference.clear` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:party.reference.clear` |
| `cartulary.view.evidence.v1` | `party.source.link` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:party.source.link` |
| `cartulary.view.evidence.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.evidence.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.evidence.v1` | `relationships.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:relationships.manage` |
| `cartulary.view.evidence.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.evidence.v1` | `surface_pivot.linked_records` | `workflow` | `surface_pivot/view_query_route` | Workflow or pivot entry | workbook surface-pivot control | `feature-action:surface_pivot.linked_records` |
| `cartulary.view.evidence.v1` | `surface_pivot.timeline` | `workflow` | `surface_pivot/view_query_route` | Workflow or pivot entry | workbook surface-pivot control | `feature-action:surface_pivot.timeline` |
| `cartulary.view.findings.v1` | `create_related.decision` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.decision` |
| `cartulary.view.findings.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.findings.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.findings.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.findings.v1` | `finding.close_or_reopen` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:finding.close_or_reopen` |
| `cartulary.view.findings.v1` | `finding.contradictory_refs.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:finding.contradictory_refs.manage` |
| `cartulary.view.findings.v1` | `finding.evidence_refs.manage` | `evidence` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:finding.evidence_refs.manage` |
| `cartulary.view.findings.v1` | `finding.owner.manage` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:finding.owner.manage` |
| `cartulary.view.findings.v1` | `finding.support_refs.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:finding.support_refs.manage` |
| `cartulary.view.findings.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.findings.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.findings.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.findings.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.findings.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.forensic_keywords.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.forensic_keywords.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.forensic_keywords.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.forensic_keywords.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.forensic_keywords.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.forensic_keywords.v1` | `keyword.evidence_refs.manage` | `evidence` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:keyword.evidence_refs.manage` |
| `cartulary.view.forensic_keywords.v1` | `keyword.findings.link` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:keyword.findings.link` |
| `cartulary.view.forensic_keywords.v1` | `keyword.timeline_rows.link` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:keyword.timeline_rows.link` |
| `cartulary.view.forensic_keywords.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.forensic_keywords.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.forensic_keywords.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.handoff.v1` | `create_related.status_review` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.status_review` |
| `cartulary.view.handoff.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.handoff.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.handoff.v1` | `handoff.acknowledge` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:handoff.acknowledge` |
| `cartulary.view.handoff.v1` | `handoff.next_checks.manage` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:handoff.next_checks.manage` |
| `cartulary.view.handoff.v1` | `handoff.open_decisions.review` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.handoff.v1` | `handoff.open_tasks.review` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.handoff.v1` | `handoff.risks.review` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.handoff.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.handoff.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.handoff.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.handoff.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.handoff.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.hosts.v1` | `create_related.decision` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.decision` |
| `cartulary.view.hosts.v1` | `create_related.note` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.note` |
| `cartulary.view.hosts.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.hosts.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.hosts.v1` | `entity.aliases.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.hosts.v1` | `entity.merge` | `relationships` | `record_action/record_merge_route` | Existing owner control | entity merge workflow | `feature-action:entity.merge` |
| `cartulary.view.hosts.v1` | `entity.relationships.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:entity.relationships.manage` |
| `cartulary.view.hosts.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.hosts.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.hosts.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.hosts.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.hosts.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.hosts.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.hosts.v1` | `surface_pivot.assessments` | `workflow` | `surface_pivot/view_query_route` | Workflow or pivot entry | workbook surface-pivot control | `feature-action:surface_pivot.assessments` |
| `cartulary.view.hosts.v1` | `surface_pivot.evidence` | `workflow` | `surface_pivot/view_query_route` | Workflow or pivot entry | workbook surface-pivot control | `feature-action:surface_pivot.evidence` |
| `cartulary.view.hosts.v1` | `surface_pivot.timeline` | `workflow` | `surface_pivot/view_query_route` | Workflow or pivot entry | workbook surface-pivot control | `feature-action:surface_pivot.timeline` |
| `cartulary.view.identities.v1` | `create_related.decision` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.decision` |
| `cartulary.view.identities.v1` | `create_related.note` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.note` |
| `cartulary.view.identities.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.identities.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.identities.v1` | `entity.aliases.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.identities.v1` | `entity.merge` | `relationships` | `record_action/record_merge_route` | Existing owner control | entity merge workflow | `feature-action:entity.merge` |
| `cartulary.view.identities.v1` | `entity.relationships.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:entity.relationships.manage` |
| `cartulary.view.identities.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.identities.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.identities.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.identities.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.identities.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.identities.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.identities.v1` | `surface_pivot.assessments` | `workflow` | `surface_pivot/view_query_route` | Workflow or pivot entry | workbook surface-pivot control | `feature-action:surface_pivot.assessments` |
| `cartulary.view.identities.v1` | `surface_pivot.evidence` | `workflow` | `surface_pivot/view_query_route` | Workflow or pivot entry | workbook surface-pivot control | `feature-action:surface_pivot.evidence` |
| `cartulary.view.identities.v1` | `surface_pivot.timeline` | `workflow` | `surface_pivot/view_query_route` | Workflow or pivot entry | workbook surface-pivot control | `feature-action:surface_pivot.timeline` |
| `cartulary.view.indicators.v1` | `create_related.decision` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.decision` |
| `cartulary.view.indicators.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.indicators.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.indicators.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.indicators.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.indicators.v1` | `indicator.lifecycle.manage` | `history` | `indicator_lifecycle/indicator_lifecycle_route` | Existing owner control | indicator lifecycle workflow | `feature-action:indicator.lifecycle.manage` |
| `cartulary.view.indicators.v1` | `indicator.lifecycle.read` | `history` | `indicator_lifecycle/indicator_lifecycle_route` | Existing owner control | indicator lifecycle workflow | `feature-action:indicator.lifecycle.read` |
| `cartulary.view.indicators.v1` | `indicator.observations.pivot` | `relationships` | `indicator_observations/indicator_observations_route` | Existing owner control | indicator observation workflow | `feature-action:indicator.observations.pivot` |
| `cartulary.view.indicators.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.indicators.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.indicators.v1` | `relationships.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:relationships.manage` |
| `cartulary.view.indicators.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.investigative_queries.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.investigative_queries.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.investigative_queries.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.investigative_queries.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.investigative_queries.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.investigative_queries.v1` | `query.evidence_refs.manage` | `evidence` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:query.evidence_refs.manage` |
| `cartulary.view.investigative_queries.v1` | `query.findings.link` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:query.findings.link` |
| `cartulary.view.investigative_queries.v1` | `query.result.link` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:query.result.link` |
| `cartulary.view.investigative_queries.v1` | `query.source.link` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:query.source.link` |
| `cartulary.view.investigative_queries.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.investigative_queries.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.investigative_queries.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.lesson.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.lesson.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.lesson.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.lesson.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.lesson.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.lesson.v1` | `lesson.close_or_reopen` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:lesson.close_or_reopen` |
| `cartulary.view.lesson.v1` | `lesson.evidence_refs.manage` | `evidence` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:lesson.evidence_refs.manage` |
| `cartulary.view.lesson.v1` | `lesson.followup_tasks.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:lesson.followup_tasks.manage` |
| `cartulary.view.lesson.v1` | `lesson.owner.manage` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:lesson.owner.manage` |
| `cartulary.view.lesson.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.lesson.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.lesson.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.notes.v1` | `artifact.evidence_refs.manage` | `evidence` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:artifact.evidence_refs.manage` |
| `cartulary.view.notes.v1` | `artifact.related_notes.manage` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:artifact.related_notes.manage` |
| `cartulary.view.notes.v1` | `artifact.source_links.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:artifact.source_links.manage` |
| `cartulary.view.notes.v1` | `artifact.tags.manage` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:artifact.tags.manage` |
| `cartulary.view.notes.v1` | `create_related.decision` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.decision` |
| `cartulary.view.notes.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.notes.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.notes.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.notes.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.notes.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.notes.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.notes.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.notes.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.notes.v1` | `surface_pivot.source_records` | `workflow` | `surface_pivot/view_query_route` | Workflow or pivot entry | workbook surface-pivot control | `feature-action:surface_pivot.source_records` |
| `cartulary.view.parties.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.parties.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.parties.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.parties.v1` | `party.reference.clear` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:party.reference.clear` |
| `cartulary.view.parties.v1` | `party.reference.clear_both` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.parties.v1` | `party.reference.link` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:party.reference.link` |
| `cartulary.view.parties.v1` | `party.usage_pivot.audience_attendee` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.parties.v1` | `party.usage_pivot.collector_source` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.parties.v1` | `party.usage_pivot.owner_stakeholder` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.parties.v1` | `party.usage_pivot.requester` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.parties.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.parties.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.parties.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.status_review.v1` | `create_related.comm_log` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.comm_log` |
| `cartulary.view.status_review.v1` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.status_review.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.status_review.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.status_review.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.status_review.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.status_review.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.status_review.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.status_review.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.status_review.v1` | `status_review.blocked_tasks.review` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.status_review.v1` | `status_review.next_report.manage` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:status_review.next_report.manage` |
| `cartulary.view.status_review.v1` | `status_review.open_decisions.review` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.status_review.v1` | `status_review.pending_evidence.review` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.status_review.v1` | `status_review.risks.review` | `workflow` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:workflow` |
| `cartulary.view.task_requests.v1` | `create_related.comm_log` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.comm_log` |
| `cartulary.view.task_requests.v1` | `create_related.lesson` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.lesson` |
| `cartulary.view.task_requests.v1` | `create_related.status_review` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.status_review` |
| `cartulary.view.task_requests.v1` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.task_requests.v1` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.task_requests.v1` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.task_requests.v1` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.task_requests.v1` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.task_requests.v1` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.task_requests.v1` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.task_requests.v1` | `task.decision.clear` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:task.decision.clear` |
| `cartulary.view.task_requests.v1` | `task.decision.link` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:task.decision.link` |
| `cartulary.view.task_requests.v1` | `task.links.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:task.links.manage` |
| `cartulary.view.task_requests.v1` | `task.requester_party.clear` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:task.requester_party.clear` |
| `cartulary.view.task_requests.v1` | `task.requester_party.link` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:task.requester_party.link` |
| `cartulary.view.task_requests.v1` | `task.status.transition` | `workflow` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:task.status.transition` |
| `cartulary.view.timeline.v2` | `create_related.comm_log` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.comm_log` |
| `cartulary.view.timeline.v2` | `create_related.decision` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.decision` |
| `cartulary.view.timeline.v2` | `create_related.evidence` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.evidence` |
| `cartulary.view.timeline.v2` | `create_related.handoff` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.handoff` |
| `cartulary.view.timeline.v2` | `create_related.lesson` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.lesson` |
| `cartulary.view.timeline.v2` | `create_related.note` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.note` |
| `cartulary.view.timeline.v2` | `create_related.status_review` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.status_review` |
| `cartulary.view.timeline.v2` | `create_related.task_request` | `workflow` | `view_row_create/view_row_create_route` | Workflow or pivot entry | related-record workflow | `feature-action:create_related.task_request` |
| `cartulary.view.timeline.v2` | `details.read` | `details` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:details` |
| `cartulary.view.timeline.v2` | `entity_mentions.create_host` | `relationships` | `entity_mention_action/entity_mention_resolve_route` | Direct contextual action | selected mention control | `feature-action:entity_mentions.create_host` |
| `cartulary.view.timeline.v2` | `entity_mentions.create_identity` | `relationships` | `entity_mention_action/entity_mention_resolve_route` | Direct contextual action | selected mention control | `feature-action:entity_mentions.create_identity` |
| `cartulary.view.timeline.v2` | `entity_mentions.dismiss` | `relationships` | `entity_mention_action/entity_mention_resolve_route` | Direct contextual action | selected mention control | `feature-action:entity_mentions.dismiss` |
| `cartulary.view.timeline.v2` | `entity_mentions.resolve` | `relationships` | `entity_mention_action/entity_mention_resolve_route` | Direct contextual action | selected mention control | `feature-action:entity_mentions.resolve` |
| `cartulary.view.timeline.v2` | `entity_mentions.restore` | `relationships` | `entity_mention_action/entity_mention_resolve_route` | Direct contextual action | selected mention control | `feature-action:entity_mentions.restore` |
| `cartulary.view.timeline.v2` | `evidence.attach_blob` | `evidence` | `record_action/evidence_attach_blob_route` | Existing owner control | evidence attachment control | `feature-action:evidence.attach_blob` |
| `cartulary.view.timeline.v2` | `evidence.download_handle` | `evidence` | `evidence_access/evidence_download_handle_route` | Direct contextual action | evidence object access control | `feature-action:evidence.download_handle` |
| `cartulary.view.timeline.v2` | `evidence.preview_handle` | `evidence` | `evidence_access/evidence_preview_handle_route` | Direct contextual action | evidence object access control | `feature-action:evidence.preview_handle` |
| `cartulary.view.timeline.v2` | `evidence.read` | `evidence` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:evidence` |
| `cartulary.view.timeline.v2` | `history.read` | `history` | `panel_read/record_history_route` | Panel read | owning panel content | `panel:history` |
| `cartulary.view.timeline.v2` | `history.rollback` | `history` | `record_action/record_rollback_route` | Direct contextual action | row-history object control | `feature-action:history.rollback` |
| `cartulary.view.timeline.v2` | `indicator.observations.manage` | `relationships` | `indicator_observations/indicator_observations_route` | Existing owner control | indicator observation workflow | `feature-action:indicator.observations.manage` |
| `cartulary.view.timeline.v2` | `record.delete` | `history` | `record_action/record_delete_route` | Direct contextual action | row-history object control | `feature-action:record.delete` |
| `cartulary.view.timeline.v2` | `record.restore` | `history` | `record_action/record_restore_route` | Direct contextual action | row-history object control | `feature-action:record.restore` |
| `cartulary.view.timeline.v2` | `relationships.manage` | `relationships` | `record_patch/record_patch_route` | Existing owner control | owner field/action control | `feature-action:relationships.manage` |
| `cartulary.view.timeline.v2` | `relationships.read` | `relationships` | `panel_read/current_row_projection` | Panel read | owning panel content | `panel:relationships` |
| `cartulary.view.timeline.v2` | `timeline.mark_reviewed` | `history` | `record_action/record_mark_reviewed_route` | Existing owner control | Timeline review control | `feature-action:timeline.mark_reviewed` |
| `cartulary.view.timeline.v2` | `timeline.supersede` | `history` | `record_action/record_supersede_route` | Existing owner control | owner supersede workflow | `feature-action:timeline.supersede` |

### 6.4 Advisory input

Run the two narrow offline searches from the planning input with
`PYTHONDONTWRITEBYTECODE=1` and `python3 -B`. Do not use `--design-system`,
`--persist`, or `--force`, and do not mutate the bundled digest. Record every
material result as `ADOPT`, `ADAPT`, or `REJECT`; reject the bundled
Cybersecurity Platform visual defaults.

The implementation-start searches produced these material dispositions:

- `ADOPT`: confirm destructive actions before execution; preserve visible
  non-color state cues; manage focus on confirmation entry and return; use
  semantic HTML controls; derive presentation instead of storing duplicate
  state.
- `ADAPT`: show disabled state with native semantics and adjacent owner-backed
  reason copy rather than relying on opacity/cursor styling; keep confirmation
  inspector-local rather than adopting generic modal ceremony.
- `REJECT`: generic state-lifting/reducer guidance that would centralize owner
  workflow state; bundled Cybersecurity Platform visual defaults; generic
  utility-class examples that bypass Cartulary tokens.

## 7. WI-02 — Shared presentation architecture

Create a shared, stateless presentation seam under
`apps/web/src/workbook/inspector/presentation/`. Keep it small and explicit:

- `WorkbookInspectorPresentation.tsx` owns `WorkbookInspectorShell`,
  `WorkbookInspectorRecordContext`, `WorkbookInspectorPanelSection`,
  `WorkbookInspectorActionGroup`, `WorkbookInspectorStatus`,
  `WorkbookInspectorCompactMetadata`, `WorkbookInspectorTechnicalDetails`,
  `WorkbookInspectorMessage`, and `WorkbookInspectorConfirmation`;
- `WorkbookHistoryPresentation.tsx` owns `WorkbookHistoryList` and
  `WorkbookHistoryEvent`; and
- `workbookInspectorPresentationModel.ts` owns pure presentation types and
  derivation of closed disabled-reason copy from existing owner state.

The seam may be factored further only when source ownership or focused tests
require it; do not create a general component library, context provider,
reducer, command bus, or universal inspector state machine.

Use one internal subject-presentation value containing the human label,
surface or record-type label, optional owner-backed state label, `recordId`,
and `rowVersion`. Human context is primary. The technical values render only
as compact metadata or inside an operable `Technical details` disclosure,
using the existing monospace role, safe wrapping, full accessible value, and a
copy control where operationally useful. Do not fetch new data to obtain a
label: reuse the selected row and existing row-label helpers.

The shared shell must remain an `aside` with a stable accessible name, visible
`Close inspector` control, one primary heading, internal scrolling, and the
existing layout-owned width, clamp, separator, overlay, and inertness
behavior. It must use existing design tokens and
`workbookSurfaceInspectorPanelStyle` or its shared equivalent; it must not
introduce local raw color, spacing, radius, shadow, typography, or breakpoint
values.

`WorkbookInspectorPanelSection` renders the contract panel heading and owner
content once. It consumes resolved `panel_read` groups without visible chrome.
It accepts contextual owner controls but does not manufacture a generic row of
feature buttons. Existing owner controls are the sole activation locus for
actions they already implement. The contextual action group is reserved for
genuine workflow or pivot entries; direct object actions remain beside the
object or decision they affect. Direct actions retain config order within
their panel.

Each action-bearing owner control receives the existing stable feature action
test ID and semantic data for feature key, route kind, route owner, and action
identity. A shared action-binding type accounts for every current actionable
group exactly once. The binding points to the real owner control or workflow;
it never adds a second activation path.

Disabled actions remain visible when the owner requires them, use native
disabled semantics, and have adjacent visible reason text associated through
an accessible description. Reasons are derived only from the current closed
disabled token, minimum incident role, and owner state. This presentation copy
is not an authorization decision or a feature-label registry.

Local messages render at the section/control that can recover from them.
Loading, empty, error, dirty, pending, disabled, destructive, and successful
states use the existing design component variants and non-color cues. One
local decision group has at most one primary action.

## 8. WI-03 — Atomic consumer migration

Migrate the following consumers in one workstream without retaining a legacy
and new presentation path:

- `TimelineWorkbookInspector` and its Timeline section composition;
- `GenericWorkbookSurface` for all contract-backed generic surfaces;
- `EntityWorkbookSurface` for Hosts and Identities; and
- `AssessmentWorkbookSurface`, preserving append-only creation and follow-on
  semantics.

Use these subject-label rules:

- Timeline: current activity synopsis, falling back to `Selected timeline row`;
- Hosts and Identities: the existing entity label;
- generic surfaces: the existing contract-aware row-label helper and active
  surface title; and
- Assessments: the current selected subject/assessment context for follow-on
  work, while the create draft title remains the creation heading.

Move actual owner controls into the contract panel they serve. Remove
capability badges and proxy buttons/messages that only prepare or point to
those controls. Preserve existing owner workflows, including mention
resolution, entity creation, relationship management, evidence access and
attachment, indicator observations/lifecycle, party linking/clearing,
assessment follow-on creation, lifecycle review/supersede, merge, and related
record creation/pivots.

For history, retain separate Timeline and generic state/command owners while
using the shared stateless event and metadata presentation. Lead each event
with a human-readable operation/diff summary, then timestamp and available
actor context. Put actor IDs, history-item references, change-set IDs,
revision numbers, rollback selectors, target kinds, record IDs, and row
versions in technical details. Never infer legal actions from displayed text.

Destructive and rollback confirmation must:

- name the human subject and operation first;
- retain exact record/version identity as secondary technical context when the
  owner requires it;
- distinguish confirm from cancel and use destructive styling only for the
  destructive action;
- invalidate on every current owner-required subject, version, view,
  lifecycle, authorization, deletion, and merge transition;
- move focus into the confirmation and restore it predictably; and
- remain inspector-local unless an adopted owner requires a modal dialog.

Update `tools/frontend_source_ownership.json` if new authored files require
accounting. Add semantic selectors only through the authored UI-contract
package. If a still-visible direct action needs better ordinary copy, edit only
the owning view-schema display hint and regenerate; do not map feature keys to
labels in client code.

## 9. WI-04 — Responsive, accessibility, and visual evidence

Preserve the grid as the primary work surface and prove:

- the base inspector remains adjacent and leaves the grid visible;
- narrow and compact inspectors use the existing overlay contract and keep the
  grid inert until close;
- inspector content scrolls internally without shell or document overflow;
- width, clamp, resize separator, and vertical-only resize behavior remain
  layout-owned;
- required information and actions remain reachable at 1024x720 and 768x640,
  at 200% zoom, with required text spacing, and with long unbroken technical
  values;
- headings, landmarks, names, descriptions, live regions, disabled semantics,
  non-color cues, and focus visibility are correct;
- `Esc` closes one layer at a time and focus restoration follows the existing
  semantic fallback ladder; and
- no responsive treatment hides required information or corrupts fixed grid
  geometry.

Reconcile the five baseline goldens through the documented Make-owned visual
workflow. Add narrow, compact, open technical-details, disabled-reason, or
long-token capture intents only when the existing fixture registry cannot
evidence those states. Review every changed image manually. Restore unrelated
command-produced PNG changes before generating the manifest, and run ordinary
visual proof twice against the retained manifest. Goldens are implementation
evidence, not product authority.

## 10. Verification routing and command matrix

At execution start, refresh current routing with:

```text
make task-guide ROLE=module-author OWNER=web.workbook
make task-guide ROLE=module-author OWNER=web.design
make task-guide ROLE=module-author OWNER=package.ui
make task-guide ROLE=module-author OWNER=module.workbook
make task-guide ROLE=module-author OWNER=module.timeline
make task-guide ROLE=module-author OWNER=module.entities
make task-guide ROLE=module-author OWNER=module.evidence
make task-guide ROLE=module-author OWNER=module.indicators
make task-guide ROLE=module-author OWNER=module.revisions
make task-guide ROLE=module-author OWNER=module.assessments
make task-guide ROLE=module-author OWNER=module.collaboration
make task-guide ROLE=module-author OWNER=web.collaboration
```

Run every focused command returned for a touched owner. The preparation-time
routes use `make test-slice OWNER=<owner>` for every listed owner; all listed
module owners also expose a service-backed slice except `web.workbook`,
`package.ui`, and `web.collaboration`.

Use the narrowest loop during each workstream, then complete the applicable
matrix in WI-05:

```text
make format
make generate
make generate-drift
make generated-artifact-policy-check
make json-shape-check
make frontend-typecheck
make frontend-unit
make frontend-import-boundary-check
make lint-biome
make browser-e2e-a11y
make browser-e2e-measurement
make browser-e2e-stateful
make browser-e2e-support
make browser-e2e-webserver-backed
make browser-e2e-visual
make test-fast
make agent-finalize
make lint-markdown
git diff --check
```

Run `make format` only when authored Go or frontend source needs formatting.
Run generation only when an owning input or generated accounting changes.
Follow the visual maintenance guide for any golden update; do not invoke test
runners directly. Run `make agent-finalize` before the broader terminal checks.
If no successful full warm-check root is retained, record that
`RESULTS_DIR` was unset rather than inventing retained evidence.

## 11. Acceptance matrix

The final execution must classify every row below as `PASS`, `N/A` with a
rationale, or `BLOCKED` with evidence.

| Acceptance | Result | Evidence |
| --- | --- | --- |
| A001 — Authority | PASS | The only normative edit is the Core 03 REQ-03-291 machine-state/visible-copy clarification; WI-01 records Core 01, Core 04, design, and domain reconciliation and rejects advisory authority. |
| A002 — Scope | PASS | One shared stateless presentation seam under `workbook/inspector/presentation/` serves all consumers; owner state and commands remain in their prior owners. |
| A004 — Tokens | PASS | Shared design tokens are reused; source review, frontend unit policy, and Biome found no component-local token, theme, density, or label registry. |
| A008 — Responsive | PASS | Measurement, accessibility, stateful, and reviewed 1024×720 and 768×640 visual evidence preserve inline-size selection, clamp, ARIA, overlay, and vertical-resize behavior. |
| A009 — Overflow | PASS | Measurement and constrained-overlay assertions prove shell-owned scrolling, no viewport overflow, and reachable controls and disclosures. |
| A010 — Inspector | PASS | Characterization covers all 17 view schemas and 247 current feature-group instances, exact-once dispatch, disabled restrictions, stale-state invalidation, and unknown additive omission. |
| A011 — Continuity | PASS | Focus, selection, scroll, retargeting, nested `Esc`, refresh, and same-shell transition tests pass in unit, support, stateful, and webserver-backed suites. |
| A018 — Evidence | PASS | Focused and service-backed Evidence owner slices plus stateful and webserver-backed evidence scenarios pass without lifecycle or command changes. |
| A019 — Accessibility | PASS | The final accessibility target passes all 12 units, including keyboard parity, names, descriptions, live regions, contrast, non-color cues, disabled reasons, and focus restoration. |
| A020 — Components | PASS | Presentation primitive units and all consumer units deterministically cover sections, bindings, messages, disclosures, confirmations, history, and compound states. |
| A022 — Visual fixtures | PASS | Five existing inspector goldens, two constrained inspector additions, and the affected mention-state golden were manually reviewed; two post-promotion proofs and the final visual target pass. |
| A023 — Selectors | PASS | Stable UI-contract and owner selectors remain the interaction anchors; stale label assertions discovered in WI-05 were replaced with semantic owner-control selectors or exact technical identity. |
| A024 — Test authority | PASS | Repository search confirms no executable artifact reads this handoff, the Core Markdown owner, or advisory content; tests characterize projected behavior directly. |
| A025 — Generated artifacts | PASS | Only the visual golden manifest changed, through the Make-owned visual update from authored fixtures and reviewed PNGs; generation drift and generated-artifact policy pass. |
| A026 — Boundary | PASS | Git review and routed owner suites confirm no route, API, schema, write, authorization, lifecycle, persistence, history, evidence, dependency, or compatibility behavior was invented. |
| A027 — Handoff | PASS | Sections 2, 6, 10-13 and the checkpoint log record owners, before/after behavior, paths, commands, evidence, failures, compatibility, rollback, risks, deferrals, and the next safe seam. |

Also apply R001-R004, R006, R008, R011, R013, R014, R018, R022, R024,
R025, and R033-R035. Record material advisory results in the WI-01 disposition
table rather than copying generic UI guidance into product behavior.

Completion requires all of the following:

- panel-read groups have no visible pseudo-action chrome;
- every meaningful action has one contextual owner-backed activation locus;
- panel order and all stable semantic identities remain intact;
- human record context is primary and technical identifiers remain available
  but secondary;
- no visible `no_row_selected` token remains while the machine token is
  unchanged;
- disabled actions explain the owner-backed restriction;
- history and destructive workflows retain exact legal-action, selector,
  concurrency, invalidation, and focus behavior;
- all current consumers use the shared presentation seam;
- applicable responsive, accessibility, stateful, service-backed,
  measurement, and visual evidence passes;
- final Git scope contains only authorized files; and
- no TODO, temporary compatibility layer, parallel inspector path, or implied
  product follow-up remains when WI-05 is marked `DONE`.

## 12. Checkpoint log

| Time | Workstream | Paths and decisions | Commands and results | Compatibility, rollback, risks, next action |
| --- | --- | --- | --- | --- |
| `2026-08-31T17:43:43-04:00` | Handoff preparation / WI-01 start | Created this documentation-only tracker at the clean preparation baseline. Recorded the verified gaps, selected Core 03 no-row clarification, bounded target architecture, ordered workstreams, evidence, and acceptance requirements. No owner, product, test, generated, or golden file changed. | Read-only Git/tool baseline and current task-guide routing were inspected. Product verification was not run because this session is documentation-only. | No product interface or behavior changed. This handoff alone is the rollback unit for this session. Next: obtain separate implementation authorization, refresh the baseline, and perform WI-01 owner reconciliation and characterization. |
| `2026-08-31T18:34:35-04:00` | WI-01 complete / WI-02 start | Refreshed the exact baseline and authorization; reconciled Core 03 REQ-03-291 so `no_row_selected` remains machine identity while ordinary copy is `Select a saved row to inspect its details.`; inventoried all 17 view schemas and 247 feature-group instances with disposition, owner locus, and semantic selector; classified advisory results without adopting advisory authority; strengthened documentation-free corpus characterization. Core 01, Core 04, design, and domain owners contain no contradiction. | `make test-slice OWNER=web.workbook` PASS, run root `.cartulary/test-results/20260831T223326Z-p18899`; `make lint-markdown` PASS, run root `.cartulary/test-results/20260831T223327Z-p19003`; `git diff --check` PASS. | Only presentation wording authority changes; no token, contract, route, state, or stored data changes. WI-02 may now add the stateless presentation seam; owner state and commands remain out of scope. |
| `2026-08-31T18:45:48-04:00` | WI-02 complete / WI-03 start | Added the stateless shared shell, record context, panel, action-binding, status/message, technical-details, disabled-reason, confirmation, and history-event presentation seam; added pure subject, history, binding, and closed disabled-copy models; accounted new sources without changing owner state or commands. | Initial `make format` failed on an introduced unused import and semantic-group lint error at `.cartulary/test-results/20260831T224101Z-p36655`; fixed. Initial `make frontend-typecheck` failed on unsupported matcher typings at `.cartulary/test-results/20260831T224135Z-p45633`; fixed. A concurrent pre-format `make lint-biome` attempt failed at `.cartulary/test-results/20260831T224229Z-p62209`; rerun passed. Final PASS: `make format` `.cartulary/test-results/20260831T224229Z-p62207`; `make frontend-typecheck` `.cartulary/test-results/20260831T224229Z-p62161`; `make lint-biome` `.cartulary/test-results/20260831T224247Z-p67110`; `make frontend-unit` `.cartulary/test-results/20260831T224257Z-p67632`; `make test-slice OWNER=web.workbook` `.cartulary/test-results/20260831T224135Z-p45496`; `make test-slice OWNER=web.design` `.cartulary/test-results/20260831T224439Z-p5301`; JSON shape, import boundary, generated policy, and `git diff --check` PASS. | Internal additive presentation types only; no route, schema, authorization, lifecycle, persistence, history, evidence, or command changes. Next: atomically migrate every consumer and remove the legacy renderer. |
| `2026-08-31T19:08:57-04:00` | WI-03 complete / WI-04 start | Atomically migrated Timeline, all generic contract surfaces, Hosts/Identities, and Assessment creation/follow-on to the shared shell and panel path; made active generic grid identity the inspector subject; removed the legacy generic capability renderer, visible panel-read badges, proxy messages, and the duplicate shell styles; bound workflow/indicator/direct lifecycle entries to their existing owners; consolidated Timeline and generic history event/confirmation presentation without moving their state or command ports; made subject labels and safe conflict copy primary, moved identifiers into technical disclosures, and retained stable selectors and lifecycle invalidation. | Introduced migration failures were classified and fixed: `make format` `.cartulary/test-results/20260831T225801Z-p56887`; `make frontend-typecheck` `.cartulary/test-results/20260831T225906Z-p65883`; `make frontend-unit` `.cartulary/test-results/20260831T230054Z-p71625`; first `make test-slice OWNER=web.workbook` `.cartulary/test-results/20260831T230449Z-p15917`. Final PASS: `make format` `.cartulary/test-results/20260831T230555Z-p31391`; focused affected row `.cartulary/test-results/20260831T230603Z-p35580`; `make test-slice OWNER=web.workbook` 139/139 `.cartulary/test-results/20260831T230641Z-p36403`; `make test-slice OWNER=web.design` 15/15 `.cartulary/test-results/20260831T230724Z-p51422`; `make frontend-typecheck` `.cartulary/test-results/20260831T230839Z-p99259`; `make lint-biome` `.cartulary/test-results/20260831T230839Z-p99302`; `make frontend-import-boundary-check` `.cartulary/test-results/20260831T230839Z-p99282`. | Presentation and internal component migration only. Stable routes, payloads, authorization, lifecycle, history selectors, evidence behavior, and stored data remain compatible. Intentional visible changes receive no compatibility wrapper. Next: add constrained-overlay, keyboard, accessibility, measurement, and reviewed visual evidence. |
| `2026-08-31T19:17:24-04:00` | WI-04 evidence in progress | Added authored 1024×720 right-overlay and 768×640 full-overlay visual intents, plus inspector-specific no-row identity/copy, safe conflict copy, confirmation focus, nested `Esc`, focus restoration, disabled-reason description, text-spacing, and 200% zoom assertions. | Introduced `make format` failure at `.cartulary/test-results/20260831T231642Z-p3714`: frontend parser found a stray delimiter in the no-row assertion update; formatter unit summary is `target-summaries/format.json`. Fixed and reran PASS at `.cartulary/test-results/20260831T231710Z-p8361`. Introduced `make frontend-typecheck` failure at `.cartulary/test-results/20260831T231744Z-p12821`: missing pure error-helper import and missing close-selector identity argument; fixed, with PASS at `.cartulary/test-results/20260831T231834Z-p17914`. Introduced `make browser-e2e-a11y` failure at `.cartulary/test-results/20260831T231903Z-p19506`: both inspector rows exposed 1:1 destructive-button contrast because the shared confirmation paired the semantic-destructive background with the danger-text token; 18/20 tests passed. Classified as an introduced product defect and repaired by using the coherent transparent danger-button background token. Second run `.cartulary/test-results/20260831T232122Z-p69767` passed 19/20 and showed that the previously latent use of generic read-only mode as `incident_closed` supplied false lifecycle copy to a viewer. Classified as an implementation defect exposed by the new evidence; repaired by carrying the actual incident-closed boolean through the existing layout owner while retaining role-based authorization derivation. The first typecheck of that structural fix failed at `.cartulary/test-results/20260831T232509Z-p21106` because the Timeline test runtime fixture lacked the new explicit lifecycle fact; classified as an introduced fixture-accounting failure and fixed with a default-false fixture input. | WI-04 remains `IN_PROGRESS`. No runtime contract changed; fixtures and accessibility evidence are authored test inputs. Next: rerun accessibility, reconcile ordinary visual output, use the Make-owned golden update, manually review retained images, and prove the promoted manifest twice. |
| `2026-08-31T19:32:17-04:00` | WI-04 pre-update visual reconciliation | Accessibility passed at `.cartulary/test-results/20260831T232613Z-p26551`; measurement passed all 22 units at `.cartulary/test-results/20260831T232746Z-p72119`. The required ordinary visual run then reconciled the refactor before any update. | `make browser-e2e-visual` failed at `.cartulary/test-results/20260831T233217Z-p29349`: 22/26 scenarios passed. Expected drift was limited to the object-first mention-state capture and first inspector baseline; the two new golden files were correctly reported absent. One introduced assertion still required a duplicated inner `Evidence` heading; classified as a test defect and retargeted to the semantic evidence-section label. Expected/actual mention and relationship images were manually compared before promotion and showed only the authorized removal of capability-first badges, raw-token framing, and repeated headings plus shared history presentation. | WI-04 remains `IN_PROGRESS`. No broad screenshot churn was accepted. Next: format, run the Make-owned visual update, inspect every command-produced changed image, and run the ordinary visual proof twice. |
| `2026-08-31T19:39:38-04:00` | WI-04 manual-review correction | The Make-owned visual update passed at `.cartulary/test-results/20260831T233527Z-p77912` and changed only the five inspector baselines, two new inspector captures, the object-first mention-state capture, and the generated manifest. Manual review accepted seven images but rejected the first 1024×720 capture because its disabled reason was below the frame. | The post-review ordinary run at `.cartulary/test-results/20260831T233938Z-p28657` passed 25/26 and proved the record ID and disabled reason could not both occupy the viewport while technical metadata was incorrectly placed before contract panels. Classified as an implementation hierarchy defect, not acceptable golden drift. Record technical metadata was moved to the secondary position after contract panels, and the narrow fixture now pairs the final owner-backed workflow restriction with the adjacent record disclosure. | WI-04 remains `IN_PROGRESS`; the first promoted narrow image is not accepted. Next: rerun focused/unit compilation, ordinary reconciliation, Make-owned update, review the corrected narrow image and all resulting inspector drift, then prove the manifest twice. |
| `2026-08-31T19:44:10-04:00` | WI-04 hierarchy verification in progress | The secondary-metadata correction passed format and typecheck at `.cartulary/test-results/20260831T234332Z-p74106` and `.cartulary/test-results/20260831T234344Z-p78300`. | `make frontend-unit` then passed 391/392 units and failed the selector-policy row at `.cartulary/test-results/20260831T234410Z-p78889`: the new text-spacing style used the raw shared `timeline-inspector` test ID. Classified as an introduced test-boundary defect; replaced with `dataTestIdSelector(timelineInspectorTestId())` passed into the browser evaluation. | WI-04 remains `IN_PROGRESS`. Stable selector ownership is preserved; no product behavior changed. Next: rerun formatting and frontend units before visual reconciliation. |
| `2026-08-31T19:57:12-04:00` | WI-04 complete / WI-05 start | Added `visual.fixture.inspector_narrow_technical_details` at 1024×720 and `visual.fixture.inspector_compact_actions` at 768×640; added explicit text-spacing, 200% zoom, disabled-description, nested-`Esc`, safe-focus, and no-overflow coverage; separated actual incident lifecycle from authorization-only read-only mode; moved record IDs/version to secondary metadata after contract panels. Manually reviewed and accepted `entity-mention-chip-states-linux.png`, all five existing `workbook-inspector-*.png` baselines, `workbook-inspector-narrow-technical-details-linux.png`, and `workbook-inspector-compact-actions-linux.png`. | Final WI-04 PASS evidence: format `.cartulary/test-results/20260831T234624Z-p16694`; frontend unit 392/392 `.cartulary/test-results/20260831T234637Z-p20912`; accessibility `.cartulary/test-results/20260831T232613Z-p26551`; measurement `.cartulary/test-results/20260831T232746Z-p72119`; pre-promotion visual reconciliation `.cartulary/test-results/20260831T234823Z-p58088` failed only on the two expected first-capture diffs; Make-owned update `.cartulary/test-results/20260831T235036Z-p4027`; ordinary visual proofs `.cartulary/test-results/20260831T235315Z-p48752` and `.cartulary/test-results/20260831T235513Z-p94032`, both PASS. The generated golden manifest was updated only by the Make target. Earlier introduced failures and fixes are recorded in the preceding WI-04 checkpoints. | Presentation/test evidence only; no route, payload, authorization decision, lifecycle transition, persistence, or stored-data contract changed. Eight reviewed images are the visual rollback set, together with the authored fixture registry and generated manifest. WI-05 may now run routed and terminal validation, complete the acceptance matrix, and close the handoff. |
| `2026-09-01T00:02:45Z` | WI-05 focused owner validation in progress | Refreshed all 12 handoff task guides. Initial focused PASS roots: `package.ui` `.cartulary/test-results/20260831T235759Z-p41828`; `web.workbook` `.cartulary/test-results/20260831T235759Z-p41827`; `web.design` `.cartulary/test-results/20260831T235759Z-p41839`. | `module.workbook` failed 65/68 at `.cartulary/test-results/20260831T235759Z-p41850`; an isolated rerun reproduced 65/68 at `.cartulary/test-results/20260901T000245Z-p64342`. The three failures were generic creation workflows whose controls were absent while `no_row_selected`. Classified as an introduced shared-shell regression: the shell suppressed all owner children in no-row mode, even though generic/entity/assessment owners intentionally retain non-row creation workflows and separately guard row-bound content. Fixed by always rendering owner children while preserving the exact machine state and ordinary no-row sentence; added unit characterization. | WI-05 remains `IN_PROGRESS`. No compatibility wrapper is added. The fix restores existing owner-scoped creation behavior without restoring stale row-bound state or changing the machine token. Next: rerun format, unit evidence, and isolated `module.workbook` before resuming other owner slices. |
| `2026-08-31T20:15:46-04:00` | WI-05 focused owner validation in progress | The no-row shell correction passed `make format`, `make frontend-typecheck`, `make test-slice OWNER=web.workbook` (139/139), and the isolated `module.workbook` slice (68/68). Focused Timeline validation then exercised its complete routed unit, integration, browser-support, webserver-backed, visual, and measurement set. | PASS roots: format `.cartulary/test-results/20260901T000740Z-p23117`; typecheck `.cartulary/test-results/20260901T000752Z-p27436`; `web.workbook` `.cartulary/test-results/20260901T000752Z-p27367`; `module.workbook` `.cartulary/test-results/20260901T000836Z-p42971`. `make test-slice OWNER=module.timeline` passed 51/53 units and failed at `.cartulary/test-results/20260901T001055Z-p2492`: the single authoritative row failure was `module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13`, which accepted the server timing mark but did not observe the generated summary before its qualification deadline; the second failed unit was the derived measurement target summary. The other three Timeline measurement rows and all 49 non-summary rows passed. Classified provisionally as an isolated timing/fixture observation, not an inspector assertion; an exact-row rerun is required before final classification. | WI-05 remains `IN_PROGRESS`; no source change is justified by one isolated measurement miss. Next: rerun the exact failed Timeline row through the Make-owned slice, then either record a clean transient recovery or investigate a reproducible product defect before any later owner begins. |
| `2026-08-31T20:25:52-04:00` | WI-05 focused owner validation in progress | The exact Timeline measurement rerun passed, closing the prior failure as transient observation. Focused Entities then exposed a stale browser expectation for the removed `Raw token` subheading; it was replaced with stable mention-resolution owner-control selectors while retaining the exact mention value assertion. Evidence and Indicators passed. Revisions then exposed a stale history string expectation for the removed `Record <uuid>` heading; the test now asserts the exact record ID within the semantic history panel and the stable restore control. | Timeline exact-row PASS 13/13 `.cartulary/test-results/20260901T001603Z-p62886`. Entities initial 40/42 failure `.cartulary/test-results/20260901T001910Z-p5711`; exact corrected row PASS 11/11 `.cartulary/test-results/20260901T002131Z-p61231`; format PASS `.cartulary/test-results/20260901T002123Z-p57055`. Evidence PASS 36/36 `.cartulary/test-results/20260901T002222Z-p4081`; Indicators PASS 20/20 `.cartulary/test-results/20260901T002336Z-p54029`. Revisions initial 24/26 failure `.cartulary/test-results/20260901T002424Z-p71760`, isolated to `module.revisions.browser_stateful.verify_inspector_details_relationships_evidence_e32cd188c5` plus its derived target summary; classified as stale presentation-copy evidence because the record ID and restore command remained present and operable. | Both assertion changes intentionally remove dependencies on obsolete presentation chrome; no route, command, selector, or lifecycle behavior changed. WI-05 remains `IN_PROGRESS`. Next: format and rerun the exact Revisions row, then continue focused Assessment, Collaboration, and web-collaboration routing. |
| `2026-08-31T20:45:22-04:00` | WI-05 routed owner validation complete | The corrected Revisions stateful row passed. The remaining focused Assessment, Collaboration, and web-collaboration slices passed. Every independently routed service-backed slice returned by the refreshed task guides also passed for web.design and the eight applicable module owners. | Format PASS `.cartulary/test-results/20260901T002612Z-p17535`; Revisions exact row 11/11 `.cartulary/test-results/20260901T002620Z-p21713`; focused Assessments 28/28 `.cartulary/test-results/20260901T002713Z-p64121`; focused Collaboration 31/31 `.cartulary/test-results/20260901T002813Z-p7574`; focused web-collaboration 4/4 `.cartulary/test-results/20260901T002949Z-p57832`. Service-backed PASS: web.design 15/15 `.cartulary/test-results/20260901T002958Z-p58620`; Workbook 39/39 `.cartulary/test-results/20260901T003103Z-p5242`; Timeline 30/30 `.cartulary/test-results/20260901T003311Z-p60216`; Entities 33/33 `.cartulary/test-results/20260901T003748Z-p18523`; Evidence 25/25 `.cartulary/test-results/20260901T003936Z-p69211`; Indicators 8/8 `.cartulary/test-results/20260901T004049Z-p18961`; Revisions 20/20 `.cartulary/test-results/20260901T004136Z-p36277`; Assessments 19/19 `.cartulary/test-results/20260901T004249Z-p80988`; Collaboration 22/22 `.cartulary/test-results/20260901T004348Z-p23774`. | Routed evidence now covers every touched owner with no remaining failed row. WI-05 remains `IN_PROGRESS`. Next: run `make agent-finalize` before the broader terminal matrix, then complete generated provenance, Git scope, acceptance, rollback, and handoff records. |
| `2026-08-31T21:07:48-04:00` | WI-05 terminal validation complete; final tracker gate pending | Final source scope comprises the Core 03 clarification; the shared presentation seam and migrated Timeline/generic/entity/assessment consumers under `apps/web/src/workbook/**`; focused unit and browser evidence under `apps/web/**`; source ownership and authored visual fixtures under `tools/**`; the Make-generated visual manifest; eight manually accepted PNGs; and this tracker. Before: duplicated shells/history, capability badges and proxy actions, raw machine copy, primary technical IDs, unexplained disablement, and capability-first mentions. After: one stateless shell/section/history/confirmation path, owner-backed contextual controls, human subject hierarchy, secondary operable technical details, closed disabled reasons, object-first relationships, and preserved owner state/commands. | `make agent-finalize` PASS `.cartulary/test-results/20260901T004547Z-p73278` with `RESULTS_DIR` unset, so retained full-warm-run maintenance was not requested. Terminal PASS roots: format `.cartulary/test-results/20260901T004607Z-p76243`; generate-drift `.cartulary/test-results/20260901T004613Z-p80377`; generated policy `.cartulary/test-results/20260901T004624Z-p83399`; JSON shape `.cartulary/test-results/20260901T004627Z-p83849`; typecheck `.cartulary/test-results/20260901T004634Z-p84352`; frontend unit 392/392 `.cartulary/test-results/20260901T004646Z-p84837`; import boundary `.cartulary/test-results/20260901T004819Z-p20806`; Biome `.cartulary/test-results/20260901T004824Z-p21242`; accessibility 12/12 `.cartulary/test-results/20260901T004832Z-p21813`; measurement 22/22 `.cartulary/test-results/20260901T005001Z-p66138`; stateful 34/34 `.cartulary/test-results/20260901T005431Z-p23439`; support 19/19 `.cartulary/test-results/20260901T005655Z-p72617`; webserver-backed 60/60 `.cartulary/test-results/20260901T005819Z-p16595`; visual 12/12 `.cartulary/test-results/20260901T010401Z-p73581`; test-fast 444/444 `.cartulary/test-results/20260901T010555Z-p18198`; Markdown lint `.cartulary/test-results/20260901T010612Z-p19012`; staged and unstaged `git diff --check` PASS. `make generate` was not run because no generated code owner input changed; the visual manifest was already regenerated by the Make-owned visual update. | Generated provenance and Git scope are authorized; `docs/domain.md`, routes, schemas, payloads, authorization, lifecycle, persistence, history, evidence, dependencies, and stored data are unchanged. Rollback is the Core clarification, shared seam/migration, tests and authored fixtures/accounting, generated manifest, eight PNGs, and tracker as one source-only unit. Residual risk is limited to normal renderer-specific golden maintenance and explicit future bindings for additive groups that are intentionally omitted until owned; no product deferral, TODO, compatibility layer, or parallel inspector path remains. Next safe seam: add a future owner-declared group through the canonical feature object and an explicit owner-control binding, with its routed characterization. |
| `2026-08-31T21:09:34-04:00` | WI-05 complete | Classified every acceptance row `PASS`, finalized the compatibility and rollback statement, and confirmed the terminal record contains all required owners, behavior, paths, failures, evidence, residual risks, deferrals, and next seam. WI-01 through WI-05 are now `DONE`. | Post-record `make lint-markdown` PASS at `.cartulary/test-results/20260901T010915Z-p21115`. All focused, service-backed, terminal, visual, generated-provenance, and Git-scope results are retained in the preceding rows. | No temporary execution path remains. The complete source-only change set is ready for review or atomic rollback; there is no implied product follow-up. |

Append one row at every workstream transition and material decision. Include
actual paths, owner clauses, before/after behavior, commands, retained run
roots, failures and classification, manual visual review, compatibility,
rollback, risk, deferral, and the next safe action.

## 13. Compatibility, rollback, and final statement

The completed implementation preserves public contracts and stable semantic
identities. It introduces no route, API, schema, payload, selector,
authorization, lifecycle, persistence, history, evidence, dependency, or
stored-data migration. Internal TypeScript presentation types are additive;
obsolete component-private interfaces and the legacy generic renderer have no
compatibility wrapper. The Core 03 clarification changes presentation wording
authority but not the `no_row_selected` token or `inspector_config_v1` shape.

The implementation rollback unit is source-only: revert the Core 03
clarification, shared presentation seam, atomic consumer migration, focused
tests and authored harness inputs, Make-generated visual manifest, eight
reviewed goldens, and this updated handoff together. No data rollback,
migration reversal, dependency removal, or external cleanup is required.

The terminal compatibility statement is:

**Presentation and internal component-architecture refactor only; no route,
API, authorization, lifecycle, persistence, history, evidence, or stored-data
migration.**
