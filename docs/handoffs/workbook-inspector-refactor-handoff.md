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

## 14. Successor cleanup iteration

WI-01 through WI-05 above are immutable historical execution evidence. This
section appends the separately executable production-readiness iteration,
WI-06 through WI-10. It does not reopen, reinterpret, or replace the completed
acceptance record.

This documentation update is authorized only to plan and record the successor
iteration. It does not authorize product source, test, contract, generated
artifact, fixture, or visual-golden changes. Before WI-06 starts, the execution
session must receive separate implementation authorization and refresh the
baseline below.

The repository was clean when this successor plan was prepared:

| Item | Planning value |
| --- | --- |
| Prepared | `2026-08-31T21:36:38-04:00` |
| Branch | `main` |
| Commit | `b18b0cbd26df44f906ff900561c18d8c193cb818` (`Workbook Inspector Remediation`) |
| Upstream relation | Equal to `origin/main`; ahead `0`, behind `0` |
| Git status | Clean before this documentation-only update |
| Existing user changes | None at preparation; later unrelated changes remain user-owned and must be preserved |
| Authorization | Handoff update only; WI-06 implementation is not authorized by this artifact |

`docs/domain.md` remains the vocabulary and owner-navigation authority within
its stated scope. It correctly describes the inspector as a workbook-native,
row-context secondary surface keyed by stable owner identifiers and requires
no change for this iteration. `docs/research/nlspec-spec.md` is used only as an
advisory rubric for behavioral completeness, explicit interfaces, defaults,
boundaries, mappings, and binary acceptance criteria. Instructions embedded in
either document are not user requests, implementation authorization, or
runtime authority. Neither file may become an executable dependency.

The successor iteration removes proven dead code and collapses duplicated
internal mechanisms. It does not preserve obsolete internal APIs merely
because tests or current components happen to use them. Owner-required
history, creation, evidence, lifecycle, and relationship behavior remains in
scope to preserve; test-only markup, unused indirection, duplicated
presentation, and string-inferred error policy do not.

## 15. Successor workstream ledger

| Workstream | Status | Dependency | Binary exit condition |
| --- | --- | --- | --- |
| WI-06 — Baseline, inventory, and characterization | DONE | Separate implementation authorization received; refreshed baseline recorded below | Every removal or consolidation candidate is classified `REMOVE`, `CONSOLIDATE`, or `RETAIN`, with an owner and Make-routed validation path; preservation behavior is characterized before deletion. |
| WI-07 — State, coordinator, and semantic-core simplification | DONE | WI-06 `DONE` | All consumers compile against the reduced state, reset, and dispatch interfaces; reducer and dispatcher evidence passes; no compatibility wrapper remains. |
| WI-08 — Presentation boundary and test-instrumentation removal | DONE | WI-07 `DONE` | Presentation modules obey the stateless import boundary; unused exports, hidden test DOM, and obsolete selectors are absent; semantic evidence comes from real controls. |
| WI-09 — Shared workflows, history, and safety hardening | DONE | WI-08 `DONE` | Both history owners and all four related-workflow consumers use the shared pure presentation models; typed failures and actor metadata rules pass; command and state ownership remains separate. |
| WI-10 — Final validation and handoff | DONE | WI-09 `DONE` | Routed and terminal evidence passes; generated provenance and Git scope are clean; the successor acceptance matrix and terminal record are complete. |

Only one successor workstream may be `IN_PROGRESS`. Immediately before work on
a slice begins, record the refreshed checkpoint and mark that slice
`IN_PROGRESS`. After the slice meets every exit criterion, append its paths,
commands, run roots, failures and classifications, compatibility, rollback,
risks, and next seam; mark it `DONE`; only then mark its successor
`IN_PROGRESS`. A required failure leaves the slice `IN_PROGRESS`. Use
`BLOCKED` only for a verified owner contradiction and record both clauses.

## 16. Successor scope and governing decisions

Future implementation is limited to the Workbook Inspector package and its
direct dependency cone:

- `apps/web/src/workbook/inspector/**`;
- direct Timeline, generic, entity, and assessment inspector consumers under
  `apps/web/src/workbook/**`;
- shared operation-failure presentation used by those consumers;
- inspector selectors and internal types under
  `packages/ui-contracts/src/**` and other already-owned package paths;
- focused unit, accessibility, stateful, measurement, support,
  webserver-backed, and visual tests affected by those paths;
- authored source-ownership, import-boundary, verification-routing, and visual
  fixture inputs under `tools/**` when the structural change requires them;
- Make-generated output from an authorized owning input; and
- this handoff as the successor execution ledger.

Do not change public routes, APIs, schemas, payloads, authorization decisions,
lifecycle transitions, persistence, stored data, history legality, evidence
semantics, or concurrency rules. Do not add dependencies, client-side feature
label registries, generic command buses, universal inspector controllers, or
compatibility aliases. Do not hand-edit generated roots or the visual manifest.

The cleanup must use the smallest durable internal interfaces:

- inspector state contains only state that production consumers read;
- the coordinator emits one owner reset event and requests focus restoration,
  without knowing feature-specific form or selection names;
- semantic dispatch receives the inspector config and a feature-group key and
  returns the canonical contract feature with a typed disposition;
- presentation modules receive already-derived models and callbacks, without
  importing owner hooks, mutation ports, reducers, or stateful controllers;
- shared history and related-workflow code owns pure mapping and presentation,
  while Timeline and generic owners retain their command, loading, refresh,
  and mutation state; and
- public error copy is derived from typed failures at the rejection boundary,
  never by searching arbitrary strings.

Internal selector and component removals are atomic. Tests migrate in the same
slice as the production removal. No alias, deprecated export, hidden marker,
parallel renderer, or temporary adapter may survive a workstream exit.

## 17. Successor remediation matrix

### G01 — Dead inspector state and model API

- **Remediation:** Remove `activePanelId`, `configViewSchemaId`,
  `select_panel`, `selectPanel`, the panel argument to `open`, and
  `firstPanelId`. Remove test-only `inspectorFeatureGroupsForPanel` and
  `inspectorNoRowState`. Delete the trivial `selectInspectorConfig` wrapper
  and read `contract.inspectorConfig` at its consumers.
- **Areas and owners:** Implementation and unit tests; `web.workbook` owns the
  model and direct consumers. No specification change is planned.
- **Rationale and long-term benefit:** State and commands with no production
  reader manufacture lifecycle cases and imply panel navigation that the
  product does not have. A smaller model makes invalidation, future panels,
  and reducer evolution easier to reason about and exhaustively test.
- **Compatibility or migration impact:** Internal TypeScript callers and tests
  migrate atomically. Stable machine status, selected record, invalidation
  generation/cause, feature identity, routes, and owner behavior remain.
- **Risk if unresolved:** Dead state can acquire accidental authority, tests
  preserve imaginary behavior, and future panel work must distinguish real
  invariants from abandoned scaffolding.
- **Binary validation:** Repository search finds none of the removed symbols;
  model consumers compile; reducer tests still prove open, close, retarget,
  no-row, invalidation, and stale-state clearing.

### G02 — Feature-specific coordinator reset ports

- **Remediation:** Replace `clearLocalForm`, `clearLifecycleState`,
  `clearMergePlan`, `clearPendingConfirmation`, `clearPreview`,
  `clearSelection`, and `clearWorkflowForm` with one required owner reset
  callback receiving `{ cause, scope }` and one required focus-restoration
  callback. The scope is `row_local` or `surface`. Close, retarget, and action
  completion use `row_local`; surface/config change, authorization loss,
  incident closure, deletion, merge, and hard refresh use `surface`.
- **Areas and owners:** Implementation, coordinator/model unit tests, and
  consumer integration tests; `web.workbook` owns the coordinator, while each
  consumer owns the effects of its reset callback.
- **Rationale and long-term benefit:** A generic coordinator should express
  invalidation extent, not enumerate every current feature. One semantic reset
  event allows future owner state to grow without widening the coordinator.
- **Compatibility or migration impact:** Every owner installs the new required
  callback in the same slice. Surface resets additionally clear owner
  selection and lifecycle state. No authorization, lifecycle, selection, or
  focus policy changes.
- **Risk if unresolved:** Each future feature adds another optional port;
  callers silently omit cleanup, and stale forms, previews, selections, or
  confirmations can survive subject changes.
- **Binary validation:** Compile-time construction requires both callbacks;
  no feature-named reset port remains; tests prove the expected scope for
  close, retarget, action completion, surface change, authorization loss,
  deletion, merge, and hard refresh, including focus fallback.

### G03 — Copied semantic features and duplicated contract types

- **Remediation:** Resolve semantic features by
  `(inspectorConfig, featureGroupKey)` and return the canonical feature object
  held by that config. Delete copied-feature deep equality and the discarded
  per-render resolution loop. Use generated `InspectorDisabledCondition`
  instead of a local token union. Replace the string-keyed route-owner map with
  an exhaustive typed disposition table: panel read, contextual workflow or
  pivot, direct history action, existing owner control, or unsupported.
- **Areas and owners:** Inspector implementation and dispatcher tests under
  `web.workbook`; generated view-contract types are consumed, not edited.
- **Rationale and long-term benefit:** Canonical objects prevent silent drift
  when the contract grows. Exhaustive dispositions make unsupported additions
  visible at compile/test time without duplicating route or label knowledge.
- **Compatibility or migration impact:** Call sites pass the key instead of a
  copied feature. All 17 current schemas and 247 current feature-group
  instances retain their semantic identity and owner locus. Unknown additive
  groups remain omitted without inference.
- **Risk if unresolved:** New contract fields can be ignored by hand-written
  equality, copied token unions diverge from generated owners, and dead runtime
  resolution adds complexity without enforcing a behavior.
- **Binary validation:** Corpus characterization resolves every current
  instance exactly once to the canonical object and expected disposition;
  an unknown additive group returns no presentation; searches find no local
  disabled-token union, deep feature comparator, or discarded resolution loop.

### G04 — Stateful and oversized presentation seam

- **Remediation:** Move contextual-action pending-confirmation state out of
  `presentation/` into an inspector orchestration container. Split the current
  presentation file into focused shell/panel, action, feedback/confirmation,
  history, and pure-model modules. Delete unused
  `WorkbookInspectorStatus` and `WorkbookInspectorMessage`; make helpers that
  are not public presentation contracts private. Add an import-boundary rule
  that forbids owner hooks, mutation ports, runtime state, reducers, and
  stateful controllers from presentation modules.
- **Areas and owners:** Implementation, source-ownership/import-boundary
  manifests, unit tests, and accessibility tests; `web.workbook` and
  `web.design` share verification ownership.
- **Rationale and long-term benefit:** Presentation can remain reusable only
  when it is pure about owner state and commands. Focused modules improve
  cohesion, reduce exported surface area, and stop a new universal inspector
  controller from forming accidentally.
- **Compatibility or migration impact:** Internal imports and tests change
  atomically. Confirmation identity, nested `Esc`, safe focus, outer-inspector
  close, and focus restoration remain. Focus-oriented effects and refs may
  remain in confirmation presentation; orchestration state may not.
- **Risk if unresolved:** Presentation and control ownership continue to blur,
  future actions add branches to a monolith, and accessibility changes become
  coupled to unrelated mutation behavior.
- **Binary validation:** Presentation files contain no `useState` or
  `useReducer` and pass the new import boundary; removed exports have no
  definitions or imports; contextual confirmation and focus tests pass.

### G05 — Production DOM and selectors created only for tests

- **Remediation:** Delete hidden feature-group marker nodes and
  `workbookInspectorFeatureGroupTestId`. Delete the hidden entity-subject node
  and `entityInspectorSubjectTestId`. Remove both selectors from
  `@cartulary/ui-contracts` without aliases. Retarget evidence to real action
  controls, the inspector `data-record-id`, panel selectors, or pure
  disposition results.
- **Areas and owners:** Inspector/entity implementation, `package.ui`, focused
  frontend and browser tests, and affected test-support helpers.
- **Rationale and long-term benefit:** The production accessibility tree and
  component API should represent product semantics, not private test
  bookkeeping. Tests grounded in real interaction loci are more durable.
- **Compatibility or migration impact:** Internal test selectors are removed
  intentionally. Actual controls retain the feature key, route kind, route
  owner, accessible disabled reason, record identity, and stable action
  selector needed by behavior tests.
- **Risk if unresolved:** Hidden marker contracts become permanent public
  baggage, tests can pass while real controls regress, and duplicated semantic
  nodes complicate accessibility and future refactors.
- **Binary validation:** Source and built DOM contain neither hidden marker nor
  selector; UI-contract exports and tests compile; semantic tests locate and
  operate the real control or inspector element.

### G06 — Duplicated history presentation mapping

- **Remediation:** Share one pure history-event mapper, rollback-label mapping,
  and technical-field builder between Timeline and generic history. Keep
  loading, command ports, mutation state, deleted-row behavior, refresh logic,
  and exact server selectors in their existing owners. Treat actor IDs as
  technical metadata unless a genuine human actor label is available.
- **Areas and owners:** Inspector presentation and Timeline/generic consumer
  implementation; history, rollback, deletion, restoration, concurrency,
  accessibility, and visual tests.
- **Rationale and long-term benefit:** Identical event hierarchy and metadata
  rules should have one pure source, while distinct state machines should not
  be forced into a universal controller. Human/technical separation prevents
  opaque identifiers from becoming user-facing names.
- **Compatibility or migration impact:** History APIs, legal actions, rollback
  scope, selectors, commands, and refresh behavior remain. Only actor-ID
  prominence may change visually; the exact value remains available in
  technical details.
- **Risk if unresolved:** Two mappers drift on destructive language, metadata,
  new event types, and accessibility; actor UUIDs continue to masquerade as
  human context.
- **Binary validation:** Both owners render the same pure event model and
  technical field order; actor UUIDs are never primary labels; exact rollback,
  delete, restore, concurrency, deleted-history, and refresh tests pass.

### G07 — Duplicated related-record workflows

- **Remediation:** Extract one pure seed/draft binding model over
  `{ recordId, cells }` and one shared related-record form presentation for
  Timeline, generic, entity, and assessment consumers. Retain separate owner
  controllers and submission logic. Timeline continues to own evidence
  post-create linking and refresh.
- **Areas and owners:** Inspector presentation and direct consumer hooks/tests;
  routed ownership includes Workbook, Timeline, Entities, Evidence, and
  Assessments where their behaviors are exercised.
- **Rationale and long-term benefit:** Draft initialization, target validation,
  and form rendering are one concept. Command sequencing and post-create
  effects are owner-specific. This boundary reduces duplicate fixes without
  creating a generic workflow engine.
- **Compatibility or migration impact:** Existing creation commands, payloads,
  append-only assessment semantics, selection, validation, and evidence links
  remain unchanged. Internal hook and form props may be replaced atomically.
- **Risk if unresolved:** Four consumers drift on seed binding, validation,
  labels, and disabled behavior; a later consolidation becomes broader and
  riskier.
- **Binary validation:** All four consumers use the shared pure model and form;
  generic, entity, assessment, and Timeline creation tests pass; Timeline
  evidence creation still links to the source record and refreshes once.

### G08 — Local presentation literals and unshared actions

- **Remediation:** Replace component-local `rem` spacing in related-workflow
  and history presentation with existing design tokens. Replace unstyled local
  inspector buttons with the established inspector action presentation. Do not
  add a new token, density rule, card, or component-local design registry.
- **Areas and owners:** Inspector implementation, `web.design` tests, source
  policy checks, and directly affected visual evidence.
- **Rationale and long-term benefit:** Shared tokens and action semantics keep
  hierarchy, hit targets, disabled treatment, and responsive behavior coherent
  as panels grow.
- **Compatibility or migration impact:** No command or interaction locus moves.
  Narrow, reviewed spacing/button drift is permitted only where removing a
  literal changes pixels.
- **Risk if unresolved:** Local visual policy multiplies, responsive fixes must
  be repeated, and actions with identical meaning present inconsistently.
- **Binary validation:** Searches and policy tests find no replaced local
  literal or unshared inspector button; affected accessibility, measurement,
  and reviewed visual scenarios pass without new token definitions.

### G09 — String-inferred public error policy

- **Remediation:** Carry the decoded `publicCode` on internal transport
  `WorkbookOperationFailure` values. Build a typed inspector error
  presentation at the failure boundary; do not inspect arbitrary strings with
  `includes`. Preserve primary copy exactly as
  `This row changed; refresh it before retrying.` for a row-version conflict.
  Put the public code and sanitized server message in operable
  `Technical details`. Delete `workbookInspectorSafePublicMessage` when every
  inspector caller uses typed failure presentation or a deliberately local
  message.
- **Areas and owners:** Internal operation-executor and inspector
  implementation, failure-model unit tests, accessibility tests, and directly
  affected visual evidence. This is not a public API change.
- **Rationale and long-term benefit:** Error safety should be exhaustive over a
  typed failure model, not dependent on text fragments. Carrying the decoded
  code preserves operational diagnostics while preventing technical values
  from becoming primary public copy.
- **Compatibility or migration impact:** Internal failure objects gain an
  optional field only where a decoded transport code exists; local failures
  legitimately omit it. Server messages are shown only after existing
  sanitization. The intentional visible change is safe typed copy instead of a
  raw code or actor-style technical label.
- **Risk if unresolved:** Copy changes can bypass conflict handling, unrelated
  text can be misclassified, raw public codes can dominate the UI, and future
  failure types require more brittle string branches.
- **Binary validation:** Tests map each typed failure deterministically; raw
  `row_version_conflict` is never primary copy; code and sanitized detail stay
  operable in technical disclosure; repository search finds no inspector
  substring inference or removed helper.

## 18. WI-06 — Baseline, inventory, and characterization

**Dependencies:** Separate implementation authorization. Refresh branch,
commit, upstream relation, Git status, toolchain, applicable instructions,
generated policy, allowed paths, relevant task guides, and unrelated changes.

**Work:**

1. Inventory every field, reducer action, command, callback port, export,
   helper, hidden marker, selector, duplicate mapper/form, design literal, and
   string-based error path named in G01-G09. Use production references,
   test-only references, generated ownership, and consumer construction sites
   to distinguish dead code from indirect use.
2. Assign each candidate `REMOVE`, `CONSOLIDATE`, or `RETAIN`; record its
   current owner, future owner, reason, compatibility effect, and narrowest
   Make validation route. `RETAIN` requires an owner-backed production need,
   not merely a test reference.
3. Add characterization before deletion for model invalidation, reset scope,
   focus restoration, exact feature disposition, unknown omission, history
   legality, actor metadata, related creation, and Timeline evidence linking.
4. Confirm all 17 schemas and 247 current feature-group instances at the live
   baseline. Recount rather than treating the historical totals as executable
   authority.
5. Reconcile the cleanup boundary with current Core owners, design direction,
   domain vocabulary, generated types, and source-ownership rules. On a real
   contradiction, record `BLOCKED: owner contradiction` and both clauses.

### WI-06 live classified inventory

| Gap and live count | Disposition | Current owner to future owner | Rationale, compatibility, and narrow Make route |
| --- | --- | --- | --- |
| G01: two unread state fields, two reducer actions, two commands, one `open` argument, one initial-state config argument, one config wrapper, two test-only helpers | `REMOVE` | `web.workbook` model/tests to no owner | They have no production reader or semantic control. Atomic internal compile break with no alias; preserve open/status/subject/invalidation and production `inspectorPanelIsDeclared`. Validate with `make test-slice OWNER=web.workbook`. |
| G02: seven optional feature-named ports across four production consumers | `CONSOLIDATE` | `web.workbook` coordinator plus Timeline/generic/entity/assessment owner callbacks to one coordinator reset event; state remains in each consumer | Scope is the durable shared concept; owner state and focus remain excluded from the coordinator. Internal callback migration only. Validate with `make test-slice OWNER=web.workbook` and affected module slices. |
| G03: one local disabled union, one route-owner wildcard map, one deep comparator family, one discarded render lookup, and one test-only production completeness helper over 17 schemas/247 feature instances | `REMOVE` plus `CONSOLIDATE` | `web.workbook` dispatcher to canonical config lookup plus an explicit stable-tuple disposition registry; generated `InspectorDisabledCondition` remains `package.view` owned | Canonical identity and fail-closed tuples prevent silent contract drift. No generated edit; unknown additions remain omitted. Validate with `make test-slice OWNER=web.workbook`. |
| G04: one 698-line presentation module, one pending-confirmation state cell, two unused component exports, and non-contract public helpers | `CONSOLIDATE` plus `REMOVE` | `web.workbook` presentation to focused shell/panel, action, feedback/confirmation, history, and pure-model modules; pending state moves to inspector orchestration | Focus effects/refs remain presentation-owned; mutations, controllers, and state orchestration remain excluded. Internal imports only. Validate with `web.workbook`, `web.architecture`, and `frontend-import-boundary-check`. |
| G05: two hidden production nodes and two authored UI selector exports | `REMOVE` | `web.workbook`/entity presentation and `package.ui` selectors to no owner; real action controls and inspector-root record/version attributes remain | Intentional internal test API break without alias. Tests move to real semantic loci. Validate with `web.workbook`, `package.ui`, and affected browser support. |
| G06: two event mappers, two rollback-label maps, and two technical-field policies | `CONSOLIDATE` | generic and Timeline history presentation to one pure inspector history model | Generic and Timeline loading, requests, commands, deleted-row handling, mutations, selectors, and refresh remain separate stateful owners. Actor IDs become technical-only. Validate through `web.workbook`, `module.workbook`, and `module.timeline`. |
| G07: two seed interpreters and two form presentations serving four consumers | `CONSOLIDATE` | generic/entity/assessment shared controller plus Timeline controller to one pure seed/draft model and form presentation | Submission remains owner-local; Timeline evidence link, accepted-row apply, and one refresh are explicit exclusions. No payload change. Validate through `web.workbook`, Timeline, Entities, Evidence, and Assessments slices. |
| G08: seven inventoried local `rem` spacing literals and unshared history/related-workflow buttons | `CONSOLIDATE` | local component styles to existing `web.design` tokens and inspector action presentation | No new token registry and no command-locus move; only narrow token/button pixel drift is permitted. Validate through `web.design`, accessibility, measurement, and visual evidence. |
| G09: one substring helper with five production presentation callers and rejection paths in Timeline/generic/entity/assessment/history | `REMOVE` plus `CONSOLIDATE` | arbitrary strings to typed `WorkbookOperationFailure` mapping and inspector error presentation | Preserve existing typed family presentation and sanitizer; add optional decoded `publicCode`; deliberate local messages remain a separate path. Internal-only type growth. Validate through `web.workbook`, affected modules, accessibility, and visual evidence. |

The live corpus recount is 17 view schemas and 247 schema-qualified semantic
tuples; every tuple is unique under view schema, feature key, panel, route kind,
route owner, and action key. Characterization covers canonical object identity,
unknown omission, reducer invalidation and row-version retarget, reset/focus,
server-advertised history actions, actor metadata, all four related-create
surfaces, Timeline evidence create-link-apply-refresh sequencing, and current
typed error-family behavior. Every `RETAIN` item above has a production owner;
every `CONSOLIDATE` item names its pure owner and stateful exclusions.

Core 01 REQ-01-615/617, Core 03 REQ-03-291/292, `docs/domain.md`, and
`docs/design.md` already own the required vocabulary, config identity,
invalidation, focus, feature execution, and design policy. No contradiction or
missing normative behavior was found, so no specification, domain, or design
edit is needed.

**Primary risks:** Deleting an indirectly constructed owner callback; treating
test usage as product authority; using this handoff or advisory research as a
runtime input; relying on stale counts after contract growth.

**Exit criteria:** The complete classified inventory and preservation tests
pass; every removal has zero production owner; every consolidation names one
pure shared owner and the stateful owners it must not absorb; every retained
symbol has a current owner-backed purpose; no unresolved contradiction remains.

**Tracker gate:** Append WI-06 evidence and mark it `DONE`. Mark WI-07
`IN_PROGRESS` before changing model, coordinator, or dispatcher source.

## 19. WI-07 — State, coordinator, and semantic-core simplification

**Dependencies:** WI-06 `DONE` with characterization passing.

**Work:**

1. Remove the G01 model fields, reducer action, commands, arguments, wrappers,
   and test-only helpers in one compile-safe change; update consumers and tests
   to use contract state directly.
2. Define the reset event with closed cause and scope semantics. Install the
   required reset and focus callbacks for every owner before removing the
   feature-specific port bag.
3. Convert semantic resolution to config-plus-key lookup returning the
   canonical feature, adopt generated disabled-condition typing, and implement
   the exhaustive disposition table.
4. Delete copied-feature comparison, string-keyed owner dispatch, discarded
   render-time lookup, and production-only completeness helpers whose sole
   consumer is characterization.
5. Run reducer, coordinator, dispatcher, consumer compilation, and exact
   corpus tests before broadening to routed `web.workbook` evidence.

**Primary risks:** Over-broad surface resets; focus restoration occurring twice;
loss of exact feature identity; silently accepting an unsupported additive
group; retaining adapters that recreate the deleted API.

**Exit criteria:** All consumers compile against only the new interfaces;
reset-scope and focus tests pass; every live contract feature resolves exactly
once; unknown additions remain omitted; repository search finds no removed
state, port, copied-feature comparator, or compatibility wrapper.

**Tracker gate:** Record paths, searches, focused commands, run roots, and any
failure classification. Mark WI-07 `DONE`, then WI-08 `IN_PROGRESS`, before
moving orchestration or deleting test instrumentation.

## 20. WI-08 — Presentation boundary and test-instrumentation removal

**Dependencies:** WI-07 `DONE` with all consumers on the reduced core.

**Work:**

1. Move contextual-action pending-confirmation orchestration out of
   `presentation/` while retaining owner commands and inspector-local focus
   behavior.
2. Split the presentation seam by responsibility and reduce its exported API
   to components and models that production consumers use.
3. Add authored source-ownership entries and an import-boundary assertion for
   the presentation directory. The rule must reject owner hooks, mutation
   ports, runtime model state, reducers, controllers, `useState`, and
   `useReducer`.
4. Remove hidden feature and entity nodes, delete both UI-contract selectors,
   and migrate all tests/support to actual controls, record identity, panels,
   or pure disposition.
5. Prove confirmation `Esc`, safe focus, outer close, and restoration behavior
   after moving state. Inspect rendered DOM and accessibility output to confirm
   no test-only node remains.

**Primary risks:** Accidentally weakening focus behavior while moving state;
making private presentation helpers a new public package; tests losing semantic
coverage when markers disappear; hand-editing generated UI contracts instead
of an authored owner.

**Exit criteria:** Import-boundary and source-ownership checks pass; presentation
contains no stateful orchestration; removed exports/selectors/DOM are absent;
real controls expose all required identity and disabled semantics; focused
unit and accessibility tests pass.

**Tracker gate:** Record selector migrations and exact semantic replacements,
not just deleted names. Mark WI-08 `DONE`, then WI-09 `IN_PROGRESS`, before
consolidating history, related workflows, design literals, or failure policy.

## 21. WI-09 — Shared workflows, history, and safety hardening

**Dependencies:** WI-08 `DONE` and the presentation boundary enforced.

**Work:**

1. Extract and adopt the shared pure history mapper, rollback labels, and
   technical metadata builder. Keep Timeline and generic controllers separate
   and verify their exact selectors, legal actions, mutation state, and refresh.
2. Move actor identifiers to technical details. Supply a primary actor label
   only from an owner-provided human label; never infer one from an ID.
3. Extract the pure related-record seed model and form presentation; migrate
   Timeline, generic, entity, and assessment consumers atomically while
   retaining owner submission and Timeline evidence link/refresh behavior.
4. Carry typed transport failure codes, adopt typed inspector error
   presentation at every caller, and remove string-based safe-message logic.
5. Replace the inventoried local spacing and button presentation with existing
   tokens/components. Run ordinary visual reconciliation before any update.
   Promote only narrow intentional drift with the Make-owned update, manually
   inspect every changed image, and run ordinary visual proof twice.

**Primary risks:** Shared helpers absorbing loading or command state; losing
Timeline evidence post-create sequencing; exposing unsanitized server text;
accepting broad visual drift caused by the module split rather than the two
intended visible changes.

**Exit criteria:** Both history owners and all four creation consumers use the
shared pure models; actor IDs are technical-only; typed failure and exact
conflict copy tests pass; evidence linking and owner refresh tests pass; no
inventoried local design literal remains; all retained golden changes are
manually accepted and proven twice.

**Tracker gate:** Record changed fixture IDs, golden paths, run roots, manual
review, and visual compatibility. Mark WI-09 `DONE`, then WI-10 `IN_PROGRESS`,
before terminal validation.

## 22. WI-10 — Final validation and handoff

**Dependencies:** WI-09 `DONE` with no unreviewed generated or visual output.

**Work:**

1. Refresh task-guide routing for every touched owner. At minimum, reassess
   `web.workbook`, `web.design`, `package.ui`, `module.workbook`,
   `module.timeline`, `module.entities`, `module.evidence`,
   `module.revisions`, `module.assessments`, and any owner added by the live
   dependency cone. Run every returned focused and service-backed route.
2. Run `make agent-finalize` before terminal broad verification. Use a retained
   successful warm-run root only when one actually exists.
3. Verify authored/generated provenance, UI-contract generation where
   applicable, source ownership, import boundaries, Git scope, absence of
   removed symbols, absence of TODOs/adapters, and absence of Markdown runtime
   dependencies.
4. Complete the successor acceptance matrix with `PASS`, justified `N/A`, or
   evidenced `BLOCKED`. Record before/after behavior, files changed, failures
   and classification, compatibility, rollback, residual risks, deferrals,
   and the next safe seam.

**Primary risks:** Treating a stale run root as current proof; running
generation without an owning-input change; overlooking generated selector or
visual-manifest provenance; declaring completion with an adapter or dead
export still present.

**Exit criteria:** Applicable focused, service-backed, and terminal checks
pass; final Git scope contains only authorized paths; every generated change
has an authored cause and Make provenance; successor acceptance is complete;
WI-06 through WI-10 are `DONE`; no temporary path or implied cleanup remains.

**Final tracker gate:** Mark WI-10 `DONE` only after the terminal checkpoint,
acceptance matrix, compatibility statement, rollback unit, risks, and next seam
are complete.

## 23. Successor validation plan

During implementation, use the narrowest Make-owned route returned by each
current task guide. Required focused scenarios are:

- no removed reducer field, action, command, reset port, helper, component
  export, hidden marker, or selector remains;
- presentation modules contain no owner state/controller imports and no
  `useState` or `useReducer`;
- every current feature resolves once to its canonical disposition and unknown
  additions remain omitted;
- actual action controls retain feature key, route kind, route owner,
  accessible disabled reason, and stable action selector;
- close, retarget, action completion, surface change, authorization loss,
  incident closure, deletion, merge, and hard refresh select the correct reset
  scope and restore focus once;
- Timeline and generic history produce identical hierarchy and technical
  metadata while retaining separate controllers, selectors, and legal actions;
- actor UUIDs never become primary human labels;
- related creation works on Timeline, generic, entity, and assessment
  surfaces, and Timeline evidence creation still links and refreshes;
- raw `row_version_conflict` is never primary copy, while the typed code and
  sanitized detail remain operable;
- no new component-local design literal, Markdown runtime dependency,
  generated-file hand edit, compatibility layer, or duplicate inspector path
  exists; and
- selection, stale-state clearing, grid scroll, responsive overlay, nested
  `Esc`, focus visibility/restoration, disabled descriptions, long technical
  values, and non-color state cues continue to pass.

The applicable terminal matrix is:

```text
make format
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

Run `make agent-finalize` before the broader terminal checks even though it is
listed with the complete target inventory above. Run `make format` only when
authored Go or frontend source changed. Run `make generate` only when an
authored owning input or generated accounting changed, and record the owning
input. For visual drift, follow the repository visual-golden guide: ordinary
reconciliation first, Make-owned update only for reviewed intentional changes,
manual review of every changed image, and two ordinary proofs after promotion.

For this immediate documentation-only update, run only:

```text
make lint-markdown
git diff --check
```

The handoff must be the sole changed file. Product, contract, generated,
conformance, fixture, and release checks are intentionally not changed or run.

## 24. Successor acceptance matrix

WI-10 must classify each row `PASS`, justified `N/A`, or evidenced `BLOCKED`.

| Acceptance | Result | Required evidence |
| --- | --- | --- |
| C001 — Authority and scope | PASS | Refreshed instructions and owners permit the bounded cleanup; no Markdown or advisory content becomes executable authority. |
| C002 — Complete inventory | PASS | Every candidate in G01-G09 is classified with current/future owner, disposition, rationale, compatibility, and Make route. |
| C003 — Dead model removal | PASS | Removed state, actions, arguments, wrappers, and helpers have zero definitions/references; preservation reducer tests pass. |
| C004 — Reset semantics | PASS | One required reset callback and one focus callback serve every owner; all listed causes select the expected scope and clear no more or less than owned state. |
| C005 — Canonical dispatch | PASS | The live corpus resolves exactly once to canonical features and typed dispositions; unknown additions are omitted; duplicated types/comparators are absent. |
| C006 — Presentation boundary | PASS | Authored ownership and import-boundary checks reject owner/stateful dependencies; presentation has no orchestration state. |
| C007 — Real semantic DOM | PASS | Hidden markers and obsolete selectors are absent; real controls and inspector record identity provide the required test evidence. |
| C008 — History consolidation | PASS | Timeline and generic history share pure mapping and metadata while preserving separate state, commands, legal actions, selectors, and refresh. |
| C009 — Related workflow consolidation | PASS | Timeline, generic, entity, and assessment use the shared seed/form model; all command behavior and Timeline evidence linking pass. |
| C010 — Safe technical metadata | PASS | Actor IDs are technical-only; typed failures drive public copy; exact conflict copy and operable sanitized detail pass. |
| C011 — Design and accessibility | PASS | No inventoried local literal or unshared action remains; responsive, keyboard, focus, description, and reviewed visual evidence passes. |
| C012 — Compatibility | PASS | No route, API, schema, payload, authorization, lifecycle, persistence, history, evidence, concurrency, dependency, or stored-data behavior changed. |
| C013 — Generated provenance | PASS | Every generated change has an authorized authored input and Make-owned provenance; prohibited generated roots were not hand-edited. |
| C014 — No compatibility residue | PASS | Searches find no alias, adapter, deprecated export, hidden test node, duplicate path, TODO, or string-inferred inspector error policy. |
| C015 — Verification and handoff | PASS | All routed/applicable terminal checks pass and the final checkpoint records paths, runs, failures, review, rollback, risks, deferrals, and next seam. |

## 25. Successor checkpoint log

| Time | Workstream | Paths and decisions | Commands and results | Compatibility, rollback, risks, next action |
| --- | --- | --- | --- | --- |
| `2026-08-31T21:36:38-04:00` | Successor planning; WI-06 not started | Appended sections 14-26 to `docs/handoffs/workbook-inspector-refactor-handoff.md`. Preserved WI-01-WI-05 unchanged; recorded the clean planning baseline, G01-G09 remediation matrix, ordered WI-06-WI-10 workstreams, validation, acceptance, and rollback. `docs/domain.md` and `docs/research/nlspec-spec.md` were reference inputs only and were not edited. | `make lint-markdown` PASS at `.cartulary/test-results/20260901T013943Z-p31673`; `git diff --check` PASS. No product check was warranted or authorized in this documentation-only step. | No runtime, test, contract, generated, fixture, golden, selector, or stored-data behavior changed. This appended documentation is the sole rollback unit. Next: obtain separate implementation authorization, refresh the baseline, and mark WI-06 `IN_PROGRESS`. |
| `2026-08-31T22:09:47-04:00` | WI-06 started | Received implementation authorization. Refreshed `AGENTS.md`, branch, commit, upstream, Git state, generated-artifact policy, nested-instruction scope, and the `web.workbook` task guide before source or test work. Baseline is `main` at `b18b0cbd26df44f906ff900561c18d8c193cb818`, equal to `origin/main` (ahead `0`, behind `0`). The staged successor addition to this handoff is the sole pre-existing user-owned change and remains preserved in place; no nested `AGENTS.md` applies below `apps/`, `packages/`, or `tools/`. WI-06 is the only `IN_PROGRESS` workstream. | Read-only baseline commands PASS: `git status --short --branch`, revision/upstream queries, staged/unstaged stats, `find apps packages tools -name AGENTS.md -print`, generated policy inspection, and `make task-guide ROLE=module-author OWNER=web.workbook` (`focused make test-slice OWNER=web.workbook`; `broader make test-fast`). | No product behavior changed. The active rollback unit remains the staged successor handoff addition. Next: complete the G01-G09 live inventory, authority comparison, and preservation characterization before implementation removal. |
| `2026-08-31T22:12:15-04:00` | WI-06 complete; WI-07 started | Recorded the complete classified G01-G09 inventory, live consumer and duplication counts, pure-owner/stateful-exclusion decisions, compatibility, and Make routes. Reconciled Core 01 REQ-01-615/617, Core 03 REQ-03-291/292, domain vocabulary, design direction, generated ownership, and source boundaries: no contradiction and no normative edit required. Strengthened live-corpus characterization so all 247 instances must resolve by canonical object identity as well as unique semantic key; existing focused evidence already covers invalidation/focus, legal history actions, actor metadata, four creation surfaces, Timeline evidence linking/refresh, and typed error behavior. | `make doctor` PASS at `.cartulary/test-results/20260901T021128Z-p46495`; `make toolchain-drift` PASS at `.cartulary/test-results/20260901T021129Z-p46976`; `make test-slice OWNER=web.workbook` PASS (139/139) at `.cartulary/test-results/20260901T021130Z-p47360`; `git diff --check` and staged diff check PASS. No failures. | Characterization-only test and tracker changes; no runtime, public, stored-data, or generated behavior changed. Rollback is the WI-06 test assertion plus successor tracker rows. Residual implementation risks are over-reset, double focus, copied identity, and wildcard admission. WI-07 is now the sole `IN_PROGRESS` workstream; next: implement G01-G03 atomically. |
| `2026-08-31T22:33:13-04:00` | WI-07 complete; WI-08 started | Reduced `workbookInspectorModel` to open/status/subject/invalidation facts; removed config/panel state, actions, arguments, wrappers, and test helpers. Replaced the optional feature-named coordinator bag with required `resetOwnerState({cause, scope})` and `restoreFocus`, migrated Timeline/generic/entity/assessment consumers, and delegated Timeline close cleanup to coordinator invalidation. Replaced copied-feature resolution with `(config, featureGroupKey)`, canonical object identity, generated disabled-condition typing, four dispositions, and a fail-closed authored fingerprint for every schema's ordered complete stable tuple corpus. Removed the wildcard owner map, comparator family, discarded lookup, and test-only production completeness helper. | `make test-slice OWNER=web.workbook` PASS (139/139) at `.cartulary/test-results/20260901T022404Z-p97838`; focused tombstone regression rerun PASS at `.cartulary/test-results/20260901T022353Z-p97271`; `module.entities` PASS (42/42) at `.cartulary/test-results/20260901T022500Z-p13620`; `module.timeline` PASS (53/53) at `.cartulary/test-results/20260901T022500Z-p13630`; serial `module.assessments` PASS (28/28) at `.cartulary/test-results/20260901T023006Z-p86348`; serial `module.workbook` PASS (68/68) at `.cartulary/test-results/20260901T023103Z-p29188`; removed-symbol searches, `git diff --check`, and staged diff check PASS. Introduced failures and resolutions: two catalog-title selectors were restored without restoring removed APIs; copied-feature Timeline tests migrated to key/canonical semantics; one Timeline tombstone regression was caused by duplicate synchronous history reset and fixed by retaining its existing stateful invalidation owner. Parallel `module.workbook` (`.cartulary/test-results/20260901T022500Z-p13616`) and `module.assessments` (`.cartulary/test-results/20260901T022500Z-p13641`) failed from concurrent Vite output-directory contention (`FONT_MANIFEST.json` copy `ENOENT`), classified unrelated to product and proven by serial passes. | Internal TypeScript break only; no alias or wrapper, and no public, route, payload, authorization, lifecycle, persistence, history, evidence, dependency, generated, or stored-data change. Rollback is atomic across model/coordinator/dispatcher, four consumers, Timeline policy/lifecycle, and tests. Residual risks move to confirmation focus and selector migration. WI-08 is now the sole `IN_PROGRESS` workstream; next: split presentation, enforce its stateless boundary, and remove hidden DOM/selectors atomically. |
| `2026-08-31T22:43:47-04:00` | WI-08 complete; WI-09 started | Moved contextual-confirmation state to `inspector/WorkbookInspectorContextualActions.tsx`; replaced the monolith with focused `WorkbookInspectorShell`, `WorkbookInspectorActions`, `WorkbookInspectorFeedback`, existing history presentation, and pure model modules. Deleted unused status/message exports, kept confirmation focus effects local, added an authored stateless import boundary and source-hook policy, and updated source ownership. Removed both hidden production nodes and the authored `workbookInspectorFeatureGroupTestId`/`entityInspectorSubjectTestId` exports without aliases. Exact selector migrations: feature characterization now uses the real `workbookInspectorFeatureActionTestId` button with feature/route attributes; entity readiness and diagnostics use `entityInspectorTestId` plus inspector-root `data-record-id`, `data-row-version`, state, and schema attributes. | `make frontend-typecheck` PASS at `.cartulary/test-results/20260901T023955Z-p87148`; `make test-slice OWNER=web.workbook` PASS (139/139) at `.cartulary/test-results/20260901T024014Z-p87758`; `web.architecture` PASS (12/12) at `.cartulary/test-results/20260901T024014Z-p87765`; `package.ui` PASS (10/10) at `.cartulary/test-results/20260901T024014Z-p87782`; `web.design` PASS (15/15) at `.cartulary/test-results/20260901T024105Z-p6151`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260901T024105Z-p6247`; `make browser-e2e-a11y` PASS (12/12) at `.cartulary/test-results/20260901T024214Z-p54188`; removed-export, hidden-node, old-import, state-hook, source, and diff checks PASS. One introduced typecheck failure at `.cartulary/test-results/20260901T023901Z-p86294` identified an unsupported matcher, unused characterization values, and the obsolete Timeline close callback parameter; all were removed and the rerun passed. No visual update or generated output occurred. | Internal component/test API migration only; confirmation identity, safe initial focus, nested `Escape`, inspector close/focus return, real action identity, routes, disabled descriptions, and entity readiness remain. Rollback is atomic across the split modules, four consumers, selectors/package tests, entity test support, ownership/import policy, and tracker. Residual risks move to shared mapper/form semantics, typed error detail, and reviewed token/button drift. WI-09 is now the sole `IN_PROGRESS` workstream; next: implement G06-G09 while preserving separate stateful controllers. |
| `2026-08-31T23:51:39-04:00` | WI-09 complete; WI-10 started | Added pure `workbookHistoryPresentationModel`, `inspectorRelatedRecordModel`, and `workbookInspectorErrorModel` owners with routed unit rows and source ownership. Timeline and generic history now share event, rollback-label, pending-field, technical-order, and action presentation while retaining separate state, requests, commands, deletion, and refresh. Timeline, generic, entity, and assessment creation now share the seed/draft model and `InspectorCreateRelatedWorkflow`; Timeline alone retains create/link/apply/single-refresh sequencing. `WorkbookOperationFailure.publicCode` is preserved only from decoded envelopes, every inspector rejection boundary maps the typed failure immediately, actor IDs are technical-only, and string inference was deleted. Existing spacing tokens and `WorkbookInspectorActionButton` replace the inventoried local literals/buttons. Visual capture stabilization replaced ambiguous history `scrollIntoViewIfNeeded` behavior with the existing explicit scroll-container anchor. No owner, Core, domain, design, API, contract, payload, lifecycle, persistence, history, evidence, concurrency, dependency, or stored-data change was required. | Final focused PASS: `frontend-typecheck` `.cartulary/test-results/20260901T030048Z-p7031`; `web.workbook` 142/142 `.cartulary/test-results/20260901T031514Z-p37201`; `web.architecture` 12/12 `.cartulary/test-results/20260901T031332Z-p87073`; `web.design` 15/15 `.cartulary/test-results/20260901T031342Z-p88598`; import boundaries `.cartulary/test-results/20260901T031322Z-p86686`; `frontend-unit` 395/395 `.cartulary/test-results/20260901T034922Z-p62783`; `module.timeline` 53/53 `.cartulary/test-results/20260901T031601Z-p52477`; `module.entities` 42/42 `.cartulary/test-results/20260901T033501Z-p95390`; `module.evidence` 36/36 `.cartulary/test-results/20260901T033646Z-p46976`; `module.indicators` 20/20 `.cartulary/test-results/20260901T033800Z-p96940`; `module.revisions` 26/26 `.cartulary/test-results/20260901T033849Z-p15118`; `module.assessments` 28/28 `.cartulary/test-results/20260901T034002Z-p59918`; `module.workbook` 68/68 `.cartulary/test-results/20260901T034104Z-p3451`; accessibility 12/12 `.cartulary/test-results/20260901T034318Z-p61114`; measurement 22/22 `.cartulary/test-results/20260901T034447Z-p5983`. Introduced assertion/type migrations failed at `frontend-typecheck` `.cartulary/test-results/20260901T025806Z-p4916`, `web.workbook` `.cartulary/test-results/20260901T030151Z-p7817` and `.cartulary/test-results/20260901T030257Z-p23196`, `frontend-unit` `.cartulary/test-results/20260901T030438Z-p39140`, assessment `.cartulary/test-results/20260901T030718Z-p79814`, Workbook `.cartulary/test-results/20260901T030755Z-p81018`, and the new-model rows `.cartulary/test-results/20260901T031247Z-p85226`; each identified stale string/selector or missing-field expectations and passed after migration. `lint-biome` at `.cartulary/test-results/20260901T031456Z-p36734` reports only import organization/formatting intentionally queued for WI-10 `make format`. The first `module.workbook` route `.cartulary/test-results/20260901T032037Z-p12393` failed only the intentional visual delta; the first post-refresh proof `.cartulary/test-results/20260901T032612Z-p15544` exposed the introduced ambiguous scroll anchor. The explicit-anchor correction passed update 12/12 at `.cartulary/test-results/20260901T032848Z-p60767` and two fresh ordinary proofs 12/12 at `.cartulary/test-results/20260901T033110Z-p6338` and `.cartulary/test-results/20260901T033305Z-p51320`. Removed-symbol, local-literal/button, presentation-state/import, hidden-node, and diff checks pass. | Accepted trigger: G08 intentionally adopted the established inspector action presentation and spacing tokens; G09 intentionally changed conflict copy and technical-detail hierarchy. Affected rows are `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` and `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`; capture IDs are `visual.capture.06903c2d94c98d2d2eaf`, `visual.capture.1aebc4c7b4cfc2965a5f`, `visual.capture.21a59ffec6a398738353`, `visual.capture.26d4c9bb1434fa99f29b`, `visual.capture.3451d6d508d3782f0df3`, `visual.capture.bda571d19ea0163393fc`, and `visual.capture.f5ca1ba5368f9538cbb9`. Registered fixture IDs are `visual.fixture.destructive_actions`, `visual.fixture.inspector_compact_actions`, `visual.fixture.inspector_narrow_technical_details`, and `visual.fixture.mention_chip_state_matrix`; the three history/error captures are active reconciled nonregistry captures. Changed goldens: `entity-mention-chip-states-linux.png`, `workbook-inspector-compact-actions-linux.png`, `workbook-inspector-destructive-confirmation-linux.png`, `workbook-inspector-history-linux.png`, `workbook-inspector-narrow-technical-details-linux.png`, `workbook-inspector-public-error-linux.png`, and `workbook-inspector-rollback-preview-linux.png`, plus the Make-generated visual manifest. Viewport, zoom, masks, and screenshot scope did not change; scroll normalization changed only for the history capture. Manual review of all seven promoted images found only intended button spacing, typed conflict copy, technical hierarchy, and stabilized scrolling, with no clipping, overflow, focus, typography, loading, or unrelated drift. Compatibility remains internal and source-only; no alias or public migration. Rollback is atomic across shared models/presentation, separate consumer adaptations, typed failure plumbing, focused tests/catalog/ownership, the anchor correction, seven goldens, generated manifest, and this row. WI-10 is now the sole `IN_PROGRESS` workstream; next: refresh every owner guide, format, generate from the authored `web.workbook` catalog input, finalize, and run the complete terminal matrix. |
| `2026-09-01T00:44:39-04:00` | WI-10 complete | Refreshed live task guides for `web.workbook`, `web.design`, `web.architecture`, `package.ui`, `module.workbook`, `module.timeline`, `module.entities`, `module.evidence`, `module.indicators`, `module.revisions`, and `module.assessments`; ran every returned focused and service-backed route. `make format` normalized authored frontend imports. The authorized authored input `tools/test_families/web.workbook.json` required `make generate`, which changed only the generated `tools/execution_topology_render_index.json`; the visual-update target separately owns `tools/frontend_visual_golden_manifest.json`. Source ownership contains every new module; UI-contract removed exports have no definition or reference; generated roots have no diff; runtime/tests have no Markdown dependency; final Git scope is limited to authorized `apps/web`, `packages/ui-contracts`, `tools`, goldens, and this handoff. Removed-symbol, state/import, hidden-DOM, duplicate seed/mapping, local design residue, substring inference, TODO/alias/temporary-adapter, provenance, manifest-hash, and `git diff --check` audits pass. C001-C015 are `PASS`; there are no deferrals or blocked criteria. | Focused PASS roots: `web.workbook` 142/142 `.cartulary/test-results/20260901T035410Z-p7539`; `web.design` 15/15 `.cartulary/test-results/20260901T035451Z-p22916` and service 15/15 `.cartulary/test-results/20260901T035603Z-p70535`; `web.architecture` 12/12 `.cartulary/test-results/20260901T035707Z-p17276`; `package.ui` 10/10 `.cartulary/test-results/20260901T035715Z-p18813`; `module.workbook` 68/68 `.cartulary/test-results/20260901T035723Z-p19261` and service 39/39 `.cartulary/test-results/20260901T035934Z-p76957`; `module.timeline` 53/53 `.cartulary/test-results/20260901T040143Z-p32496` and service 30/30 `.cartulary/test-results/20260901T040624Z-p91935`; `module.entities` 42/42 `.cartulary/test-results/20260901T041059Z-p50491` and service 33/33 `.cartulary/test-results/20260901T041249Z-p1989`; `module.evidence` 36/36 `.cartulary/test-results/20260901T041437Z-p52793` and service 25/25 `.cartulary/test-results/20260901T041551Z-p3323`; `module.indicators` 20/20 `.cartulary/test-results/20260901T041702Z-p52719` and service 8/8 `.cartulary/test-results/20260901T041749Z-p70390`; `module.revisions` 26/26 `.cartulary/test-results/20260901T041837Z-p87662` and service 20/20 `.cartulary/test-results/20260901T041949Z-p32996`; `module.assessments` 28/28 `.cartulary/test-results/20260901T042100Z-p77797` and service 19/19 `.cartulary/test-results/20260901T042201Z-p21472`. `make format` PASS `.cartulary/test-results/20260901T035341Z-p99789`; `make generate` PASS `.cartulary/test-results/20260901T035350Z-p4490`; `make agent-finalize` PASS `.cartulary/test-results/20260901T042302Z-p63879`, with retained-run maintenance skipped because `RESULTS_DIR` was unset. Terminal PASS: generation drift `.cartulary/test-results/20260901T042320Z-p66772`; generated policy `.cartulary/test-results/20260901T042331Z-p69791`; JSON shape `.cartulary/test-results/20260901T042335Z-p70241`; typecheck `.cartulary/test-results/20260901T042341Z-p70743`; frontend unit 395/395 `.cartulary/test-results/20260901T042356Z-p71288`; import boundary `.cartulary/test-results/20260901T042442Z-p87043`; Biome `.cartulary/test-results/20260901T042448Z-p87487`; accessibility 12/12 `.cartulary/test-results/20260901T042453Z-p88029`; measurement 22/22 `.cartulary/test-results/20260901T042622Z-p32921`; stateful 34/34 `.cartulary/test-results/20260901T043050Z-p89871`; support 19/19 `.cartulary/test-results/20260901T043314Z-p39629`; webserver-backed 60/60 `.cartulary/test-results/20260901T043440Z-p83223`; visual 12/12 `.cartulary/test-results/20260901T044022Z-p40657`; `test-fast` 447/447 `.cartulary/test-results/20260901T044217Z-p84883`; post-tracker Markdown lint `.cartulary/test-results/20260901T044608Z-p87955`; `git diff --check` PASS. No WI-10 command failed. | Before/after: dead and feature-specific core surfaces are gone; semantic dispatch is canonical and fail-closed; presentation is state-orchestration-free; tests use real DOM; history, creation, design, and error policy have cohesive shared owners without merging state machines. Compatibility remains the recorded internal TypeScript/test API break only; public and stored-data interfaces are unchanged. The complete source-only rollback unit is the WI-06-WI-10 model/coordinator/dispatcher cleanup, presentation split, selector removals, shared models, typed failure plumbing, consumer/test/catalog/ownership changes, explicit visual anchor, seven reviewed goldens, two Make-generated manifests/indexes, and successor ledger. Residual risk is limited to ordinary future contract growth: each new feature must add a canonical tuple, real owner control, reset participation, and focused evidence. The next safe extension seam is the canonical contract feature plus explicit dispatcher disposition, owner reset callback, shared pure presentation model where applicable, and an owner-held controller; no wildcard or universal controller is needed. WI-06 through WI-10 are `DONE`. |

Append a checkpoint after every workstream and before its successor begins.
Every failure entry must identify the target, retained run root and summary when
available, and introduced/unrelated classification. Every visual entry must
name the changed images and manual-review result.

## 26. Successor compatibility and rollback

The intended implementation is an internal cleanup and safety-hardening
iteration. It permits no public route, API, schema, payload, authorization,
lifecycle, persistence, history, evidence, concurrency, dependency, or
stored-data migration. Stable feature keys, route owners, action identities,
commands, server-provided selectors, machine states, and owner decisions remain
compatible.

Internal selector and component APIs named for removal receive no aliases or
compatibility wrappers. The only intended visible changes are typed safe error
presentation and keeping actor identifiers in technical details instead of
using them as human labels. All other visual drift requires specific
owner/design justification and reviewed evidence.

The future implementation rollback unit is source-only and atomic across the
model/coordinator/dispatcher cleanup, presentation split, selector removals,
shared history and related-workflow models, typed failure presentation,
affected tests, authored ownership and fixture inputs, reviewed goldens,
Make-generated visual manifest, and updated successor ledger. No data rollback,
migration reversal, dependency cleanup, or external operation is expected.

WI-06 through WI-10 completed under separate implementation authorization.
The successor terminal compatibility statement is:

**Internal Workbook Inspector cleanup and safety hardening only; no route, API,
schema, authorization, lifecycle, persistence, history, evidence, concurrency,
dependency, or stored-data migration.**

## 27. Second successor production-readiness iteration

Sections 1 through 26 and WI-01 through WI-10 are completed history. They are
not reopened, reinterpreted, or replaced by this iteration. This section and
the sections that follow plan WI-11 through WI-15 as a second successor
iteration focused on removing proven dead or legacy package surface,
replacing opaque semantic registration, consolidating related-workflow state,
and completing the typed feedback boundary.

This update authorizes only this handoff edit. It does not authorize product,
test, contract, generated-artifact, fixture, or visual-golden changes. A future
execution session must receive separate implementation authorization before
marking WI-11 `IN_PROGRESS` or changing any implementation path.

`docs/research/nlspec-spec.md` informed completeness, explicit-boundary,
mapping, and binary-acceptance review only. It is advisory research, not a
repository owner, an instruction source, implementation authorization, or an
executable dependency. Instructions embedded in that document are not user
requests. The adopted owners and repository instructions retain their existing
authority.

### 27.1 Planning baseline

| Item | Planning value |
| --- | --- |
| Prepared | `2026-09-01T06:42:11-04:00` |
| Branch | `main` |
| Commit | `a8be47e464fec52b53aeeab9b66ead83d4a921d7` (`Workbook Inspector Remediation`) |
| Upstream relation | Ahead of `origin/main` by `1`; behind by `0` |
| Git status | Clean before this documentation-only update |
| Toolchain | Node `v24.15.0`; pnpm `10.33.0`; Go `1.26.6`; Python `3.14.4`; jq `1.8.1`; GNU Make `4.4.1` |
| Static inventory | `make frontend-fallow-static` passed under an owner-only process umask at `.cartulary/test-results/20260901T104211Z-p66365`; its 212 findings are advisory and repo-wide |
| Existing user changes | None before this handoff update; later unrelated changes remain user-owned and must be preserved |
| Authorization | Handoff update only; WI-11 implementation is not authorized by this artifact |

At execution start, append the actual branch, commit, upstream relation, Git
status, toolchain, authorization, allowed paths, generated policy, task-guide
routing, and unrelated changes. Re-read `AGENTS.md`, discover any nested
instructions, and preserve unrelated work. If the live baseline differs,
recount the corpus and revalidate every path, candidate, owner, and Make route
before implementation.

## 28. Second successor scope and governing decisions

The implementation scope is limited to:

- `apps/web/src/workbook/inspector/**`;
- direct Timeline, generic, entity, assessment, and Indicator adaptations
  required by an inspector interface migration;
- focused tests and browser evidence that exercise those paths;
- authored source-ownership, import-boundary, or test-routing inputs only when
  a new source or enforcement rule requires them;
- Make-generated output from an authorized authored input; and
- this handoff as the live WI-11 through WI-15 tracker.

The following work is explicitly deferred:

- broad decomposition of `GenericWorkbookSurface`, `EntityWorkbookSurface`,
  and `AssessmentWorkbookSurface`;
- adjacent Timeline adapter casts or transport normalization not required by
  an inspector interface migration;
- repo-wide Fallow findings outside the direct inspector dependency cone; and
- unrelated harness, dependency, generated-contract, design-system, or test-
  support cleanup.

File size alone is not evidence of dead code or authorization to redesign an
owner. Fallow findings are candidate evidence only. A production caller or an
adopted owner may justify retention; test-only use does not. Do not add a
suppression to preserve an unused interface.

No public route, API, schema, payload, authorization, lifecycle, persistence,
history, evidence, concurrency, dependency, generated-contract, or stored-data
change is permitted. Do not introduce a universal inspector controller,
generic command bus, client-side feature-label registry, compatibility alias,
deprecated wrapper, or parallel legacy path. Generated roots remain Make-
owned and must not be hand-edited.

Core 01, Core 03, Core 04, `docs/design.md`, and `docs/domain.md` currently
provide sufficient behavioral, presentation, security, and vocabulary
authority. No normative edit is planned. WI-11 must confirm that conclusion.
If adopted owners assert incompatible requirements, record
`BLOCKED: owner contradiction`, cite both clauses, and stop the affected
workstream rather than choosing an interpretation in code.

## 29. Verified planning inventory and classification rules

The planning pass reconfirmed 17 current view schemas and 247 inspector
feature-group instances. The package has no Fallow-reported unused file,
circular dependency, boundary violation, policy violation, stale suppression,
or unused catalog entry. The relevant Fallow candidates are:

| Candidate | Planning disposition | Current production owner after remediation |
| --- | --- | --- |
| `WorkbookInspectorActions.workbookInspectorActionSemanticProps` | `PRIVATIZE` | `WorkbookInspectorContextualAction` in the defining presentation module |
| `semanticInspectorDispatcher.semanticInspectorFeatureKey` | `PRIVATIZE` | Exact semantic registry lookup inside the dispatcher package |
| `workbookInspectorErrorModel.workbookInspectorFeedbackPresentation` | `REMOVE` | Replaced in WI-14 by the used discriminated-feedback renderer |
| `useInspectorCreateRelatedWorkflow.InspectorCreateRelatedWorkflowState` | `PRIVATIZE` in WI-12, then replace in WI-13 | Shared pure related-workflow state owner |
| `IndicatorInspectorWorkflow.isIndicatorInspectorAction` re-export | `REMOVE` | No caller; the handler module retains its private validator |
| `indicatorInspectorHandlers.isIndicatorInspectorAction` export modifier | `PRIVATIZE` | Internal binding construction in the defining module |

Source inspection additionally confirmed these manually identified candidates:

| Candidate | Planning disposition | Reason |
| --- | --- | --- |
| `WorkbookInspectorActionBinding.viewSchemaId` | `REMOVE` | Written during binding and never read |
| `WorkbookInspectorActionGroup.bindings` | `REMOVE` | Used only for a redundant empty check already owned by the caller |
| `registeredInspectorConfigFingerprints` and `inspectorConfigFingerprint` | `REPLACE` | Opaque, schema-wide registration with an unnecessarily broad failure domain |
| `semanticInspectorDisposition` and the test-local `expectedDisposition` | `REPLACE` | Production and test duplicate the same inferred policy instead of consuming an explicit mapping |
| Generic and Timeline related-workflow state transitions | `CONSOLIDATE` | Same pure lifecycle; different commands and post-create effects remain stateful exclusions |
| `string | WorkbookInspectorErrorPresentation` feedback | `REPLACE` | Primitive overloading requires duplicate render-time discrimination |

At WI-11 refresh, classify any newly discovered candidate by these rules:

- `REMOVE` when no production caller or adopted owner requires the symbol;
- `PRIVATIZE` when all legitimate use is inside its defining module;
- `CONSOLIDATE` only when the common owner is pure and stateful commands,
  requests, refresh, or owner effects can remain excluded; and
- `RETAIN` only with a named production caller, current/future owner,
  compatibility rationale, and Make-routed validation path.

## 30. Second successor remediation matrix

### G10 — Dead package surface and redundant internal APIs

- **Remediation:** Remove or privatize the six verified Fallow findings. Remove
  `WorkbookInspectorActionBinding.viewSchemaId`,
  `WorkbookInspectorActionGroup.bindings`, and its redundant empty guard.
  Delete rather than alias obsolete interfaces.
- **Areas and owners:** Inspector implementation, Indicator direct consumer,
  focused tests, static-analysis evidence, and this tracker. `web.workbook`
  owns the package; `module.indicators` owns the affected specialized workflow.
- **Rationale and long-term benefit:** These members expose implementation
  details, duplicate caller knowledge, or survive only because the previous
  split left public modifiers behind. A smaller package surface makes
  ownership explicit and future refactoring safer.
- **Compatibility or migration:** Atomic internal TypeScript migration. No
  public or stored-data interface changes and no compatibility aliases.
- **Risk if unresolved:** Unused APIs can acquire accidental consumers and
  become permanent compatibility obligations.
- **Binary validation:** Removed-symbol searches are empty; scoped Fallow
  output no longer lists the named exports or type; TypeScript and focused
  tests pass; every remaining package export has a production consumer or
  cross-module owner.

### G11 — Opaque whole-config semantic registration

- **Remediation:** Replace config fingerprints, hashing, inferred
  dispositions, and the test's duplicate disposition algorithm with one
  readable authored registry of complete tuples: view schema, feature key,
  panel, route kind, route owner, action key, and disposition. Lookup the exact
  tuple and return the canonical feature held by the supplied config.
- **Areas and owners:** Inspector dispatcher and corpus tests under
  `web.workbook`; generated view contracts are consumed, not edited.
- **Rationale and long-term benefit:** A magic config hash obscures review and
  makes one unknown additive tuple disable all otherwise supported tuples in
  that schema. Per-tuple registration is independently reviewable and
  fail-closed without a schema-wide failure domain.
- **Compatibility or migration:** Every current tuple retains canonical object
  identity and disposition. An unknown or altered tuple remains unsupported;
  registered sibling tuples in the same config continue resolving.
- **Risk if unresolved:** Updating an opaque hash can admit an unintended group,
  while ordinary additive contract growth can remove unrelated controls.
- **Binary validation:** Recount the live corpus; every tuple resolves exactly
  once by object identity; duplicate and stale registry entries fail; altered
  tuples are unsupported; an additive unknown tuple does not affect a
  registered sibling; fingerprint, hash, inferred fallback, and duplicated
  test policy are absent.

### G12 — Duplicated imperative related-workflow state

- **Remediation:** Add one pure reducer shared by the generic and Timeline
  hooks. State contains an opaque workflow ID, subject key including row-
  version identity, canonical feature, form model, and `editing` or
  `submitting` phase. Update, rejection, completion, cancellation, and
  retarget events apply only to the matching workflow ID.
- **Areas and owners:** Pure inspector model, both stateful hooks, generic,
  entity, assessment, Timeline, and Evidence-linked creation tests. The pure
  state belongs to `web.workbook`; commands and post-create effects remain
  with their current owner modules.
- **Rationale and long-term benefit:** Both hooks duplicate the same state
  transitions, and an asynchronous result can currently arrive after cancel
  or retarget. One exhaustive reducer prevents drift and rejects obsolete
  results without creating a universal command controller.
- **Compatibility or migration:** Generic creation retains `onCreated`.
  Timeline retains create, evidence link, accepted-row apply, and exactly one
  refresh. Command payloads, validation, selection, and append-only assessment
  behavior do not change.
- **Risk if unresolved:** Controllers diverge and stale completion can
  resurrect an obsolete workflow or overwrite a new subject's state.
- **Binary validation:** Pure reducer tests cover begin, update, submit,
  reject, retry, cancel, complete, retarget, and stale-result omission; both
  hooks use the reducer; all four consumers pass; Timeline Evidence linking
  and refresh remain exact.

### G13 — Primitive-or-object feedback contract

- **Remediation:** Replace `string | WorkbookInspectorErrorPresentation` with
  a discriminated `WorkbookInspectorFeedback` containing either a plain
  message or typed error presentation. Add used constructors and one shared
  renderer. Migrate Timeline and Indicator producers and remove render-time
  `typeof` branches.
- **Areas and owners:** Inspector error model and presentation, Timeline and
  Indicator direct consumers, unit and accessibility tests, and this tracker.
- **Rationale and long-term benefit:** Primitive overloading hides intent and
  makes each renderer rediscover the variant. A discriminated model can grow
  future status variants without string inference or another breaking union.
- **Compatibility or migration:** Preserve exact copy, neutral-message
  rendering, error styling, live-region behavior, and technical details.
  Local form errors remain separate typed form state.
- **Risk if unresolved:** A new consumer can present operation failures as
  ordinary text or discard structured technical detail.
- **Binary validation:** TypeScript rejects raw feedback strings; one shared
  renderer serves every shared feedback locus; every message and failure path
  retains its expected presentation; no feedback `typeof` branch or obsolete
  helper remains.

## 31. Second successor workstream ledger and tracker protocol

| Workstream | Status | Dependency | Binary exit condition |
| --- | --- | --- | --- |
| WI-11 — Authority, baseline, inventory, and characterization | DONE | Separate implementation authorization and refreshed checkpoint | Scoped inventory and preservation characterization are complete; every retained interface has a production owner; no unresolved owner contradiction remains. |
| WI-12 — Semantic registry and package-surface minimization | DONE | WI-11 `DONE` | Exact tuple lookup is the only semantic path; scoped dead exports, fields, props, fingerprints, and inferred fallback are absent. |
| WI-13 — Related-workflow state consolidation | DONE | WI-12 `DONE` | Both stateful hooks use the pure reducer; stale results cannot revive canceled or retargeted workflows; all creation paths pass. |
| WI-14 — Typed feedback convergence | DONE | WI-13 `DONE` | The discriminated feedback model and shared renderer are the only shared feedback path; focused accessibility and visual evidence pass without unintended change. |
| WI-15 — Final validation and handoff | DONE | WI-14 `DONE` | All applicable routed and terminal evidence passes; generated provenance and Git scope are clean; C016-C025 and the terminal checkpoint are complete. |

Only one workstream may be `IN_PROGRESS`. Immediately before a workstream
begins, append its refreshed checkpoint and mark it `IN_PROGRESS`. After its
exit criteria pass, append paths, decisions, commands, run roots, failures and
classifications, compatibility, rollback, residual risks, and next action;
mark it `DONE`; only then mark its successor `IN_PROGRESS`. A required failure
leaves the workstream `IN_PROGRESS`. Use `BLOCKED` only for a verified owner
contradiction and record both clauses.

## 32. WI-11 — Authority, baseline, inventory, and characterization

1. Refresh repository instructions, nested-instruction scope, branch, commit,
   upstream relation, Git state, toolchain, generated policy, source
   ownership, task guides, and the live contract corpus.
2. Mark WI-11 `IN_PROGRESS` only after separate implementation authorization
   and the refreshed checkpoint are recorded.
3. Recount view schemas, feature tuples, direct consumers, feedback producers,
   related-workflow controllers, exported package surface, and scoped static
   findings.
4. Classify every scoped candidate by section 29. Test-only use cannot justify
   `RETAIN`. Confirm the stateful exclusions for each consolidation.
5. Add or strengthen documentation-free preservation characterization for
   canonical identity, current dispositions, unknown omission, sibling
   isolation, related creation, Timeline Evidence linking, feedback copy and
   roles, cancellation, retargeting, and owner focus/reset behavior.
6. Run Fallow with an owner-only process umask. Record the scoped report and
   distinguish it from unrelated repo-wide findings; do not widen scope or
   add suppressions to make the count zero.
7. Confirm no Core, domain, or design edit is required. A contradiction uses
   the blocking protocol in section 28.

**Exit:** The inventory and preservation evidence are complete, current
behavior passes, every retained interface has a production owner and Make
route, and each consolidation has a pure owner with explicit stateful
exclusions.

**Tracker gate:** Mark WI-11 `DONE`, record all evidence, then mark WI-12
`IN_PROGRESS` before semantic or package-surface edits.

## 33. WI-12 — Semantic registry and package-surface minimization

1. Implement G10 and G11 atomically.
2. Add the readable exact-tuple registry in authored inspector source; do not
   move an implementation disposition into generated or normative contracts.
3. Keep the semantic-key builder private to the registry/dispatcher package.
4. Reject duplicate registry keys and prove every live tuple is registered
   exactly once without deriving expected disposition through a second
   algorithm.
5. Prove an unknown additive tuple is omitted while a registered sibling in
   the same config remains supported.
6. Remove the verified dead exports, type modifier, binding field, action-group
   prop, empty guard, hash, fingerprint, and inferred fallback with no alias.
7. Run `web.workbook`, `web.architecture`, `module.indicators`, frontend
   typecheck/unit, import-boundary, Biome, and scoped Fallow evidence.

**Primary risks:** A transcription error in the explicit registry, lost
canonical identity, a stale registry entry, accidental admission of a future
tuple, or retaining a compatibility wrapper.

**Exit:** Exact tuple lookup is the only semantic path; all current tuples and
additive-isolation cases pass; scoped static findings and manual dead members
are absent.

**Tracker gate:** Record paths, corpus counts, searches, reports, run roots,
and failure classifications; mark WI-12 `DONE`, then WI-13 `IN_PROGRESS`.

## 34. WI-13 — Related-workflow state consolidation

1. Implement G12 as a pure reducer with exhaustive actions and no command,
   transport, refresh, or React dependency.
2. Use an owner-created opaque workflow ID so a canceled and reopened workflow
   for the same feature and subject cannot accept the earlier request's
   result.
3. Include row-version identity in each subject key. Retarget clears only a
   workflow whose subject key no longer matches.
4. Migrate the generic hook and the Timeline hook to the reducer. Do not merge
   their submission controllers.
5. Preserve generic `onCreated`; preserve Timeline create, Evidence link,
   accepted-row apply, single refresh, failure retention, and success copy.
6. Add reducer and hook evidence for editing, validation rejection, operation
   rejection, retry, cancellation, row-version retarget, subject retarget,
   completion, and late-result omission.
7. Run focused and service-backed Workbook, Timeline, Entities, Assessments,
   and Evidence routes returned by their refreshed task guides.

**Primary risks:** Dispatching a result to a newer workflow, clearing too much
owner state, changing creation payloads, or losing Timeline Evidence
sequencing.

**Exit:** Both hooks use the pure reducer, stale asynchronous results are
ignored, all four creation consumers pass, and Timeline Evidence linking and
refresh remain exact.

**Tracker gate:** Record reducer cases, consumer paths, run roots, failures,
compatibility, and rollback; mark WI-13 `DONE`, then WI-14 `IN_PROGRESS`.

## 35. WI-14 — Typed feedback convergence

1. Implement G13 with a closed discriminated feedback type and constructors
   for plain messages and typed operation failures.
2. Add one shared renderer that preserves the current neutral message and
   public-error presentation branches.
3. Migrate Timeline selection, related creation, mentions, Evidence attach,
   keyboard and feature controllers, plus Indicator workflow feedback.
4. Keep local related-form errors in `InspectorRelatedRecordFormModel`; do not
   route them through transient shared feedback.
5. Delete raw-string assignments, duplicate `InspectorMessage` renderers,
   render-time `typeof` discrimination, and the obsolete conversion helper.
6. Add source-policy and focused unit coverage for the closed type, renderer,
   every failure family, neutral copy, technical detail, role, and live-region
   behavior.
7. Run `web.workbook`, `web.architecture`, `web.design`, `module.timeline`,
   `module.indicators`, frontend typecheck/unit, import-boundary, Biome,
   accessibility, and ordinary visual evidence.

**Primary risks:** Changing copy or announcement behavior, losing technical
detail, misclassifying a neutral status as an operation failure, or leaving a
parallel renderer.

**Exit:** The discriminated type and shared renderer are the only shared
feedback path; focused and browser evidence passes; no unexpected visual
change exists.

**Tracker gate:** Record every migrated producer and renderer, commands, run
roots, failures, accessibility/visual outcome, compatibility, and rollback;
mark WI-14 `DONE`, then WI-15 `IN_PROGRESS`.

## 36. WI-15 — Final validation and handoff

1. Refresh task-guide routing for every touched owner. Expected owners are
   `web.workbook`, `web.architecture`, `web.design`, `module.workbook`,
   `module.timeline`, `module.entities`, `module.evidence`,
   `module.assessments`, and `module.indicators`; add an owner only when the
   live dependency cone identifies it.
2. Run every returned focused and service-backed slice and record its run root.
3. Run `make format`. Run `make generate` only when an authored generator input
   changed, and record that input and every generated output.
4. Run `make agent-finalize` before the broader terminal matrix. Pass
   `RESULTS_DIR` only when a real successful retained warm-run root exists.
5. Run the applicable terminal matrix in section 37.
6. Verify authored/generated provenance, source ownership, exact registry
   coverage, scoped Fallow cleanup, final Git scope, removed-symbol searches,
   no Markdown runtime dependency, and absence of TODOs, aliases, deprecated
   wrappers, suppressions, or parallel legacy paths.
7. Classify C016 through C025 as `PASS`, justified `N/A`, or evidenced
   `BLOCKED`. Record before/after behavior, failures, compatibility, rollback,
   residual risks, deferrals, and the next safe extension seam.
8. Mark WI-15 `DONE` only after the terminal checkpoint is complete. Rerun
   Markdown lint and `git diff --check` after the final tracker edit.

**Exit:** All applicable routed and terminal evidence passes, generated changes
have Make-owned provenance, Git scope is authorized, and WI-11 through WI-15
are `DONE`.

## 37. Second successor validation plan

Run repository commands from the root through public Make targets. During
WI-11 and WI-12, run Fallow under an owner-only process umask:

```text
umask 077
make frontend-fallow-static
```

Use the refreshed task guides to select focused rows. The expected routes are:

```text
make task-guide ROLE=module-author OWNER=web.workbook
make task-guide ROLE=module-author OWNER=web.architecture
make task-guide ROLE=module-author OWNER=web.design
make task-guide ROLE=module-author OWNER=module.workbook
make task-guide ROLE=module-author OWNER=module.timeline
make task-guide ROLE=module-author OWNER=module.entities
make task-guide ROLE=module-author OWNER=module.evidence
make task-guide ROLE=module-author OWNER=module.assessments
make task-guide ROLE=module-author OWNER=module.indicators
```

The final applicable terminal matrix is:

```text
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
make lint-markdown
git diff --check
```

No visual change is expected. Treat any visual difference as a regression
unless a separately justified owner/design decision makes it intentional. An
intentional update must follow the repository visual-golden guide, name the
authored trigger and affected fixtures/images, inspect every changed image
manually, and pass two fresh ordinary visual proofs after promotion.

For this immediate documentation-only update, run only:

```text
make lint-markdown
git diff --check
```

The handoff must be the sole tracked change. Product, contract, generated,
fixture, golden, conformance, and release evidence is not changed by this
planning step.

## 38. Second successor acceptance matrix

WI-15 must classify every row as `PASS`, justified `N/A`, or evidenced
`BLOCKED`.

| Acceptance | Result | Required evidence |
| --- | --- | --- |
| C016 — Authority and bounded scope | PASS | Refreshed owners authorize the inspector-cone cleanup; no owner contradiction or normative edit was found, advisory Markdown is not an executable dependency, and deferred broader work remains untouched. |
| C017 — Complete production-owner inventory | PASS | Every scoped candidate is classified by section 29; retained interfaces have named production owners and Make routes, while test-only use justified no retention. |
| C018 — Minimal package surface | PASS | Named dead exports, fields, props, guards, obsolete helpers, and the final unused registry type export have zero remaining public definitions or references; no alias, deprecated wrapper, or suppression replaced them. |
| C019 — Exact semantic registration | PASS | The authored registry and live corpus contain the same 247 exact tuples across 17 schemas; tests prove unique canonical identity and explicit disposition, duplicate/stale rejection, and absence of hashes or inferred fallback. |
| C020 — Additive isolation | PASS | Focused negative tests prove unknown, altered, ambiguous, and duplicate addressed tuples are unsupported while a registered sibling in the same config remains supported. |
| C021 — Shared workflow state | PASS | Generic and Timeline hooks use one pure exhaustive reducer while their mutation ports, payloads, refresh, selection, and owner effects remain separate. |
| C022 — Stale-result rejection | PASS | Pure and deferred-promise tests cover cancel/reopen, row-version and subject retarget, retry, late rejection, late completion, and captured Timeline Evidence work; only the matching workflow ID commits UI state or feedback. |
| C023 — Typed feedback | PASS | Raw strings cannot inhabit shared feedback; constructors and one shared renderer preserve neutral/error styling, technical details, test IDs, assertive errors, Timeline non-announcement, and Indicator polite announcements; no duplicate discriminator remains. |
| C024 — Compatibility and provenance | PASS | Public and stored-data interfaces are unchanged; behavior changes are limited to additive tuple isolation and stale UI-result rejection; the sole generated change is the Make-owned topology index derived from the authored Workbook catalog. |
| C025 — Verification and handoff | PASS | Every applicable routed and terminal check passes; final scope, failures, visual result, rollback, residual risk, deferrals, and next seam are recorded; WI-11-WI-15 are `DONE`. |

## 39. Second successor checkpoint log

| Time | Workstream | Paths and decisions | Commands and results | Compatibility, rollback, risks, next action |
| --- | --- | --- | --- | --- |
| `2026-09-01T06:42:11-04:00` | Second successor planning; WI-11 not started | Appended sections 27-40 to this handoff. Preserved sections 1-26 and WI-01-WI-10 unchanged; recorded the clean planning baseline, bounded inspector-cone scope, G10-G13 remediation, ordered WI-11-WI-15 workstreams, validation, acceptance, compatibility, rollback, and explicit deferrals. `docs/research/nlspec-spec.md` was advisory input only and was not edited. | Planning inspection confirmed the baseline and generated policy. Owner-only-umask `make frontend-fallow-static` PASS at `.cartulary/test-results/20260901T104211Z-p66365`; its 212 repo-wide findings were treated as advisory, with six scoped unused export/type candidates recorded. `make lint-markdown` PASS at `.cartulary/test-results/20260901T104448Z-p67587`; `git diff --check` PASS. | No product, test, contract, generated, fixture, golden, route, API, schema, authorization, lifecycle, persistence, history, evidence, concurrency, dependency, or stored-data behavior changed. This handoff addition is the sole rollback unit. Next: obtain separate implementation authorization, refresh the baseline, and mark WI-11 `IN_PROGRESS`. |
| `2026-09-01T06:56:54-04:00` | WI-11 started | The user explicitly authorized implementation of WI-11-WI-15. Re-read the root `AGENTS.md`; no nested instructions apply. Refreshed `main` at `a8be47e464fec52b53aeeab9b66ead83d4a921d7`, ahead of `origin/main` by one and behind by zero. The pre-existing sections 27-40 handoff edit is the only tracked user change and remains preserved. Authorized paths are the inspector dependency cone, focused tests, required authored ownership/routing inputs and their Make-generated outputs, and this tracker. Generated roots and files remain governed by `tools/generated_artifact_policy.json`. | Toolchain remains Git `2.53.0`, Node `v24.15.0`, pnpm `10.33.0`, Go `1.26.6`, Python `3.14.4`, jq `1.8.1`, and GNU Make `4.4.1`. Live authored corpus recount PASS: 17 view schemas and 247 feature-group instances; `contracts/view-schemas/index.json` is the non-schema index. Baseline inventory and owner-routed characterization are in progress. | No product or contract behavior changed at this checkpoint. Core 01 REQ-01-615/617, Core 03 REQ-03-291/292, Core 04 acceptance, domain vocabulary, and design direction have no identified contradiction and require no planned normative edit. WI-11 is the only `IN_PROGRESS` workstream. Next: refresh task routes, scoped static evidence, and preservation tests. |
| `2026-09-01T06:58:06-04:00` | WI-11 complete; WI-12 started | Reconfirmed the live 17-schema/247-instance corpus, direct inspector and Indicator consumers, two related-workflow controllers, Timeline/Indicator shared-feedback producers, two render-time discriminators, and the complete G10-G13 candidate set. Classified the six Fallow findings and the unused binding field/action-group prop as `REMOVE` or `PRIVATIZE`; exact registration as `REPLACE`; generic and Timeline pure workflow lifecycle as `CONSOLIDATE`; command, request, accepted-row application, refresh, selection, and post-create owner effects as `RETAIN`; and primitive feedback as `REPLACE`. Existing routed evidence preserves canonical object identity, unknown omission, invalidation/focus, all four creation consumers, Timeline Evidence link/apply/one-refresh sequencing, and feedback copy/roles. | Refreshed all nine expected task guides. `make doctor` PASS at `.cartulary/test-results/20260901T105731Z-p77185`; `make toolchain-drift` PASS at `.cartulary/test-results/20260901T105732Z-p77649`; owner-only-umask `make frontend-fallow-static` PASS at `.cartulary/test-results/20260901T105737Z-p78106` with 196 advisory dead-code findings, exactly six scoped candidates, and no scoped unused file, cycle, boundary, policy, or stale-suppression finding; `make test-slice OWNER=web.workbook` PASS 142/142 at `.cartulary/test-results/20260901T105747Z-p78754`; `git diff --check` PASS. No command failed. | Characterization and tracker changes only; no product, contract, generated, fixture, or golden behavior changed. No owner contradiction or normative edit is required. Stateful consolidation explicitly excludes mutation commands, requests, selection, accepted-row application, refresh, and post-create effects. Rollback is the WI-11 ledger evidence only. WI-12 is now the sole `IN_PROGRESS` workstream; next: implement exact per-tuple admission and remove the scoped dead surface atomically. |
| `2026-09-01T07:05:12-04:00` | WI-12 complete; WI-13 started | Added the authored 247-row exact semantic registry and source-ownership entry. The dispatcher now resolves one addressed config-owned feature, admits its exact view/key/panel/kind/owner/action tuple, returns canonical object identity and explicit disposition, and isolates unknown, altered, additive, or duplicate addressed tuples from registered siblings. Removed schema fingerprints, hashing, inferred disposition, duplicate test policy, the unused action-binding schema field/action-group prop/guard, the obsolete feedback conversion helper, the Indicator validator re-export, and unnecessary public modifiers without aliases. | PASS: `frontend-typecheck` `.cartulary/test-results/20260901T110058Z-p80142`; `web.workbook` 142/142 `.cartulary/test-results/20260901T110113Z-p80651`; `web.architecture` 12/12 `.cartulary/test-results/20260901T110153Z-p95984`; `module.indicators` 20/20 `.cartulary/test-results/20260901T110158Z-p97479`; import boundaries `.cartulary/test-results/20260901T110246Z-p15630`; `make format` `.cartulary/test-results/20260901T110305Z-p16754`; Biome rerun `.cartulary/test-results/20260901T110308Z-p20943`; frontend unit 395/395 `.cartulary/test-results/20260901T110315Z-p21413`; owner-only-umask scoped Fallow `.cartulary/test-results/20260901T110449Z-p58862`, with all six named findings absent and advisory dead-code count reduced from 196 to 191; removed-member and `git diff --check` audits PASS. The first Biome run failed only unformatted new registry source at `.cartulary/test-results/20260901T110249Z-p16026`; classified introduced formatting drift and resolved by the Make-owned formatter and passing rerun. | Internal TypeScript admission and package-surface migration only. Current tuples, labels, selectors, routes, authorization, and stored data remain compatible; intended correction is sibling isolation for unknown additions. Rollback is atomic across registry, dispatcher, G10 removals, focused tests, source ownership, and this row. Residual risk moves to asynchronous workflow identity. WI-13 is now the sole `IN_PROGRESS` workstream; next: install the shared pure reducer and migrate both controllers while preserving owner effects. |
| `2026-09-01T07:52:27-04:00` | WI-13 complete; WI-14 started | Added one pure exhaustive related-record reducer with structured view/record/row-version subjects, opaque `symbol` workflow identity, canonical feature/target/draft/error state, and `editing`/`submitting` phases. Generic and Timeline hooks now share only that reducer; their mutation ports, payloads, callbacks, selection, accepted-row application, refresh, and Evidence command tails remain separate. Generic, Entity, Assessment, and Timeline consumers supply exact subject identity. Timeline distinguishes a captured accepted owner sequence from UI cancellation so create/link/apply/one-refresh completes while only the current workflow may commit reducer state or feedback. Added pure reducer cases and deferred hook proofs for cancel/reopen, late acceptance, late rejection, row-version/surface retarget, and stale Timeline Evidence effects. Added the two test sources to authored ownership and `web.workbook` routing; `make generate` changed only Make-owned `tools/execution_topology_render_index.json`. | PASS: `make generate` `.cartulary/test-results/20260901T111631Z-p73514`; typecheck `.cartulary/test-results/20260901T111645Z-p76543` and final `.cartulary/test-results/20260901T112955Z-p71408`; new reducer/hook rows `.cartulary/test-results/20260901T113019Z-p72449`; `web.workbook` 144/144 `.cartulary/test-results/20260901T111710Z-p77808`; `module.workbook` 68/68 `.cartulary/test-results/20260901T113026Z-p73204`; `module.timeline` 53/53 `.cartulary/test-results/20260901T113242Z-p32586`; `module.entities` 42/42 `.cartulary/test-results/20260901T113719Z-p92042`; `module.evidence` 36/36 `.cartulary/test-results/20260901T113908Z-p43575`; `module.assessments` 28/28 `.cartulary/test-results/20260901T114019Z-p93735`; service-backed `module.workbook` 39/39 `.cartulary/test-results/20260901T114141Z-p38995`, `module.timeline` 30/30 `.cartulary/test-results/20260901T114348Z-p94020`, `module.entities` 33/33 `.cartulary/test-results/20260901T114823Z-p52492`, `module.evidence` 25/25 `.cartulary/test-results/20260901T115008Z-p3719`, and `module.assessments` 19/19 `.cartulary/test-results/20260901T115117Z-p53093`; final format `.cartulary/test-results/20260901T112951Z-p67258`; `git diff --check` PASS. `web.workbook` has no service-backed rows, so its direct invocation returned the catalog artifact error without a retained root and was classified expected `N/A` by the refreshed guide. Introduced failures: typecheck `.cartulary/test-results/20260901T110843Z-p60689` found one wrong Assessment identifier and nullable-key narrowing, resolved at `.cartulary/test-results/20260901T111056Z-p62105`; catalog format preflight reported unsorted authored rows/titles/collaborators without a retained root, and formatter `.cartulary/test-results/20260901T111536Z-p64739` found two exhaustive-dependency errors, all resolved at `.cartulary/test-results/20260901T111627Z-p69415`; first `module.workbook` `.cartulary/test-results/20260901T111757Z-p93099` exposed loss of normal Evidence success feedback after the accepted row-version reset. Focused diagnostic reruns `.cartulary/test-results/20260901T112039Z-p55178`, `.cartulary/test-results/20260901T112200Z-p56221`, `.cartulary/test-results/20260901T112232Z-p56908`, `.cartulary/test-results/20260901T112300Z-p57570`, `.cartulary/test-results/20260901T112648Z-p64266`, `.cartulary/test-results/20260901T112713Z-p64933`, `.cartulary/test-results/20260901T112816Z-p65718`, and `.cartulary/test-results/20260901T112834Z-p66354` retained the same introduced assertion failure while isolating the lifecycle race; the captured-owner-sequence correction passed the exact row at `.cartulary/test-results/20260901T113009Z-p71899` and both stale-result rows remained green. | Internal TypeScript/test/routing migration only. Creation payloads, validation, generic `onCreated`, append-only Assessment behavior, success/error copy, focus/reset behavior, Evidence linking, accepted-row application, and refresh count remain compatible; the intended correction is rejection of stale workflow UI results. No public, route, API, schema, authorization, lifecycle, persistence, history, evidence, dependency, generated-contract, or stored-data migration occurred. Rollback is atomic across the reducer, both hooks, five consumer/composition adaptations, feature lifecycle coordination, three test files, authored ownership/catalog, generated topology index, and this row. Residual risk moves to primitive shared feedback; WI-14 is now the sole `IN_PROGRESS` workstream. Next: install the closed feedback type and shared renderer, then migrate every Timeline and Indicator producer without changing copy, role, live-region, or pixels. |
| `2026-09-01T08:13:18-04:00` | WI-14 complete; WI-15 started | Replaced primitive-or-error shared feedback with closed `message` and `error` variants, explicit `none`/`polite` announcement policy, neutral-message and typed-operation-failure constructors, and one shared renderer. Timeline selection, related creation, mentions, Evidence attach, keyboard, feature control, selection-target feedback, and both mention-panel direct producers now construct non-announced messages; Indicator producers construct polite messages and typed failures. Timeline retains its muted `bodyStyle`, Indicator retains its status style, and errors reuse the assertive public-error path with technical details. Field-local related-form errors and Indicator load errors remain typed local state. Deleted both render-time discriminators and the duplicate Timeline `InspectorMessage`; production searches find no raw shared-feedback setter, `typeof` feedback discriminator, obsolete conversion helper, or duplicate renderer. Added constructor/renderer/announcement tests, a source-policy assertion, authored routing for the new evidence, and regenerated the Make-owned topology index. | PASS: format `.cartulary/test-results/20260901T115604Z-p96851`, `.cartulary/test-results/20260901T115821Z-p3350`, and final `.cartulary/test-results/20260901T121304Z-p2245`; generation `.cartulary/test-results/20260901T115825Z-p7456`; focused feedback/model/source-policy rows `.cartulary/test-results/20260901T115842Z-p10426`; `module.indicators` 20/20 `.cartulary/test-results/20260901T115850Z-p11195`; final frontend unit 398/398 `.cartulary/test-results/20260901T120140Z-p66842`; import boundaries `.cartulary/test-results/20260901T120323Z-p5044`; Biome `.cartulary/test-results/20260901T120326Z-p5440`; `web.workbook` 145/145 `.cartulary/test-results/20260901T120331Z-p5858`; `web.architecture` 12/12 `.cartulary/test-results/20260901T120337Z-p6269`; `web.design` 15/15 `.cartulary/test-results/20260901T120342Z-p6731`; `module.timeline` 53/53 `.cartulary/test-results/20260901T120450Z-p54357`; accessibility 12/12 `.cartulary/test-results/20260901T120929Z-p13052`; ordinary visual 12/12 `.cartulary/test-results/20260901T121059Z-p57500`; production source searches and `git diff --check` PASS. Introduced typecheck `.cartulary/test-results/20260901T115608Z-p1503` found one missed raw Timeline selection-target message and passed at `.cartulary/test-results/20260901T115642Z-p2180`. First frontend unit `.cartulary/test-results/20260901T115940Z-p28929` passed 397/398 and failed only because the new renderer test used an unowned literal test ID; the exact selector-policy row passed after removing the unnecessary literal at `.cartulary/test-results/20260901T120132Z-p66258`, followed by the complete unit pass. | Internal TypeScript feedback migration only. Exact copy, test IDs, neutral/error colors, technical details, Timeline non-announced messages, Indicator polite status announcements, and assertive failures remain compatible. No pixels or goldens changed; ordinary visual evidence passed without update or manual promotion. No public, route, API, payload, schema, authorization, lifecycle, persistence, history, evidence, dependency, generated-contract, or stored-data change occurred. Rollback is atomic across the feedback model/constructors, shared renderer, Timeline/Indicator producers, focused tests/source policy, authored catalog, generated topology index, and this row. Residual risk is terminal integration/provenance only. WI-15 is now the sole `IN_PROGRESS` workstream; next: refresh all nine task guides, run every applicable focused/service route, finalize, execute the terminal matrix, audit C016-C025, and close the handoff. |
| `2026-09-01T09:04:01-04:00` | WI-15 complete | Refreshed all nine task guides and ran every applicable focused and service-backed route. Final state replaces dead package surface with private owners, a readable exact 247-tuple registry, one pure related-workflow reducer with opaque identities, and one discriminated feedback model/renderer. Registry/live-corpus equality is 247 tuples across 17 schemas; authored source ownership covers the registry and both new hook tests. The authored `tools/test_families/web.workbook.json` change generated only `tools/execution_topology_render_index.json` through `make generate`. Final Git scope is bounded to 32 `apps/web/src/workbook` files, three authorized `tools` files, and this handoff; no generated contract, database, dependency, fixture, golden, or image changed. Searches find no removed symbol, raw feedback setter, `typeof` discriminator, duplicate renderer, fingerprint/hash/fallback, Markdown runtime dependency, TODO/FIXME, suppression, alias, deprecated wrapper, or parallel legacy path. C016-C025 are `PASS`. | Focused PASS: `web.workbook` 145/145 `.cartulary/test-results/20260901T121442Z-p8659`; `web.architecture` 12/12 `.cartulary/test-results/20260901T121444Z-p9056`; `web.design` 15/15 `.cartulary/test-results/20260901T121449Z-p9496`; `module.workbook` 68/68 `.cartulary/test-results/20260901T121556Z-p55774`; `module.timeline` 53/53 `.cartulary/test-results/20260901T121806Z-p11480`; `module.entities` 42/42 `.cartulary/test-results/20260901T122250Z-p69374`; `module.evidence` 36/36 `.cartulary/test-results/20260901T122436Z-p20595`; `module.assessments` 28/28 `.cartulary/test-results/20260901T122546Z-p70010`; `module.indicators` 20/20 `.cartulary/test-results/20260901T122644Z-p12893`. Service-backed PASS: `web.design` 15/15 `.cartulary/test-results/20260901T122733Z-p30326`; `module.workbook` 39/39 `.cartulary/test-results/20260901T122834Z-p76552`; `module.timeline` 30/30 `.cartulary/test-results/20260901T123039Z-p32031`; `module.entities` 33/33 `.cartulary/test-results/20260901T123516Z-p89812`; `module.evidence` 25/25 `.cartulary/test-results/20260901T123701Z-p41013`; `module.assessments` 19/19 `.cartulary/test-results/20260901T123810Z-p90331`; `module.indicators` 8/8 `.cartulary/test-results/20260901T123908Z-p33230`. `web.workbook` and `web.architecture` have no service-backed rows. Format `.cartulary/test-results/20260901T123956Z-p50566`; generation `.cartulary/test-results/20260901T124000Z-p54658`; agent finalize `.cartulary/test-results/20260901T124013Z-p57571`, with retained-run maintenance skipped because `RESULTS_DIR` was unset. Terminal PASS: generation drift `.cartulary/test-results/20260901T124032Z-p60461`; generated policy `.cartulary/test-results/20260901T124040Z-p63446`; JSON shape `.cartulary/test-results/20260901T124041Z-p63857`; typecheck `.cartulary/test-results/20260901T124045Z-p64314`; frontend unit 398/398 `.cartulary/test-results/20260901T124055Z-p64801`; import boundaries `.cartulary/test-results/20260901T124058Z-p65173`; Biome `.cartulary/test-results/20260901T124101Z-p65577`; accessibility 12/12 `.cartulary/test-results/20260901T124112Z-p66138`; measurement 22/22 `.cartulary/test-results/20260901T124242Z-p11031`; stateful 34/34 `.cartulary/test-results/20260901T124715Z-p67878`; support 19/19 `.cartulary/test-results/20260901T124939Z-p17613`; webserver-backed 60/60 `.cartulary/test-results/20260901T125104Z-p61220`; visual 12/12 `.cartulary/test-results/20260901T125648Z-p18474`; `test-fast` 450/450 `.cartulary/test-results/20260901T125844Z-p62691`. No WI-15 target failed. Owner-only-umask Fallow target PASS at `.cartulary/test-results/20260901T125912Z-p63593` surfaced one introduced advisory unused registry type export; it was privatized without an alias. Post-fix format `.cartulary/test-results/20260901T130023Z-p64653`, typecheck `.cartulary/test-results/20260901T130026Z-p68805`, `web.workbook` 145/145 `.cartulary/test-results/20260901T130036Z-p69222`, Biome `.cartulary/test-results/20260901T130116Z-p84921`, `test-fast` 450/450 `.cartulary/test-results/20260901T130118Z-p85385`, and final Fallow `.cartulary/test-results/20260901T130222Z-p2607` pass with zero scoped findings; its remaining 190 repo-wide dead-code findings are advisory and deferred. Final generation drift `.cartulary/test-results/20260901T130310Z-p3485`, generated policy `.cartulary/test-results/20260901T130318Z-p6477`, JSON shape `.cartulary/test-results/20260901T130319Z-p6881`, post-tracker Markdown lint `.cartulary/test-results/20260901T130703Z-p8412`, and `git diff --check` PASS. | Before/after: schema-wide opaque admission, duplicated stale-prone lifecycle state, primitive feedback, and dead public mechanics are replaced by exact per-tuple isolation, one exhaustive pure state machine, workflow-ID guarded UI commits, typed feedback, and the smallest production-owned surface. Compatibility is unchanged except for the two intended corrections: unknown additive tuples no longer disable registered siblings, and stale workflow results no longer overwrite current UI state. No visual delta or golden update occurred. The rollback unit is atomic across registry/dispatcher cleanup, package-surface removals, reducer and hook migrations, feedback model/renderer and producers, focused tests, authored ownership/catalog, generated topology index, and WI-11-WI-15 ledger; no data rollback is required. Residual risk is ordinary future tuple and workflow growth. Deferred work remains broad surface decomposition, adjacent Timeline casts, repo-wide Fallow cleanup, dependencies, and unrelated harness/test support. The next safe seam requires an exact tuple and disposition, real owner control, reset/invalidation participation, and focused evidence; it must not add wildcard dispatch or a universal controller. WI-11 through WI-15 are `DONE`. |

Append a checkpoint after every workstream and before its successor begins.
Every failure entry must identify the target, retained run root and summary when
available, and introduced or unrelated classification. Every visual entry must
state whether images changed and record the manual-review result when they did.

## 40. Second successor compatibility, rollback, and deferrals

The completed implementation is an internal TypeScript and test migration. It
permits no public route, API, schema, payload, authorization, lifecycle,
persistence, history, evidence, concurrency, dependency, generated-contract,
or stored-data migration. Removed internal interfaces receive no alias or
deprecated compatibility wrapper.

The only intended behavioral corrections are:

- an unknown additive semantic tuple no longer disables registered siblings
  in the same inspector config; and
- a canceled, reopened, or retargeted related workflow ignores asynchronous
  results belonging to the obsolete workflow ID.

Feedback copy, neutral-message presentation, error presentation, technical
details, roles, and live-region behavior remain compatible. No visual change
is expected.

The completed rollback unit is source-only and atomic across the exact semantic
registry, package-surface removals, shared pure workflow reducer, both hook
migrations, discriminated feedback model and renderer, direct consumer
adaptations, focused tests, authored ownership/routing inputs, Make-generated
outputs derived from those inputs, and the WI-11 through WI-15 ledger. No data
rollback, migration reversal, dependency cleanup, or external operation is
expected.

Broad Workbook surface decomposition, adjacent Timeline adapter casts,
repo-wide Fallow cleanup, and unrelated harness or test-support refactors are
explicitly deferred. They require their own owner inventory, authorization,
compatibility analysis, and workstream plan after this package boundary is
stable.

The second successor terminal compatibility statement is:

**Internal Workbook Inspector package cleanup and stale-state hardening only;
no route, API, schema, authorization, lifecycle, persistence, history,
evidence, concurrency, dependency, generated-contract, or stored-data
migration.**

## 41. Third successor authority, authorization, and baseline

Sections 1 through 40 are completed history and MUST remain unchanged. This
section begins the third-successor production-readiness plan. WI-16 through
WI-22 are the sole active execution ledger for this iteration, and all seven
workstreams are `PLANNED`. This planning update does not authorize product,
test, contract, generated, fixture, golden, conformance, or release-evidence
changes. Separate implementation authorization is required before WI-16 may
be marked `IN_PROGRESS`.

The adopted Core owners remain normative. Core 01 owns exact configured
feature identity and unsupported-feature omission; Core 03 owns row-version
identity, invalidation, and retained recovery behavior; Core 04 owns
authorization. `docs/domain.md` continues to own Workbook-native vocabulary,
and `docs/design.md` continues to own presentation and accessibility direction.
`docs/research/nlspec-spec.md` was used only as advisory drafting guidance. It
is neither a requirement owner nor an executable dependency. No Core, domain,
or design change is planned. If WI-16 finds contradictory owner clauses, mark
the affected workstream `BLOCKED: owner contradiction`, record both clauses,
and stop that workstream.

The documentation-planning checkpoint is:

- branch `main` at `fa784e85d7e7f14a56e1548958272dd76d321429`, equal to
  `origin/main`, with a clean worktree before this handoff edit;
- 17 live view schemas and 247 inspector feature-group instances;
- all 17 schemas declare `record.delete`, `record.restore`, and
  `history.rollback`;
- four surface-local `featureGroups.filter` predicates independently select
  visible inspector actions;
- Timeline implements delete, restore, and rollback, while Generic, Entity,
  and Assessment expose delete and rollback mechanisms without a retained
  tombstone restore path;
- Generic, Entity, and Assessment contain their inspector panel assembly and
  dispatch inside their surface roots;
- 11 production structural assertions remain in the scoped Timeline
  clipboard, mention, Evidence, and history adapters;
- owner-only-umask `make frontend-fallow-static` passed at
  `.cartulary/test-results/20260901T211002Z-p5486` with 190 advisory repo-wide
  findings. The scoped useful findings are the unnecessary recovery-message
  exports; the remaining inspector candidates were identified by direct
  production-consumer searches. Owner-required collaboration, mutation
  runtime, and pending-queue members are explicit static-analysis false
  positives and remain out of scope.

The implementation boundary is internal Workbook Inspector production
readiness plus the directly required Timeline adapters and private
`package.view_contracts` literal-preservation helper. No public route, wire
API, payload, schema, authorization, persistence, Evidence semantics,
dependency, generated contract, or stored-data migration is authorized.

## 42. Third successor gap register and disposition rules

| Gap | Current failure | Durable disposition | Areas | Intended behavior change |
| --- | --- | --- | --- | --- |
| G14 | Same-file and test-only mechanics remain exported. | `PRIVATIZE` or move only to a production-consumed pure owner; do not alias. | Implementation, tests, static analysis, tracker. | None. |
| G15 | Exact semantic registration is treated as executable support while four owners independently filter and dispatch features. | `CONSOLIDATE` owner capability resolution while retaining exact semantic admission. | Inspector model, four owners, tests, tracker. | Registered-but-ownerless features are omitted. |
| G16 | Core history behavior is duplicated and restore is unavailable outside Timeline. | `CONSOLIDATE` pure state and presentation; `RETAIN` separate command sequences. | Model, presentation, adapters, consumers, tests, tracker. | Non-Timeline restore and stale-result rejection become available. |
| G17 | Ordinary successes still enter error-only state and presentation. | `REPLACE` primitive/error callbacks with typed feedback at the producer. | Feedback model, controllers, presentation, accessibility tests, tracker. | Success urgency is corrected. |
| G18 | Inspector composition remains embedded in three broad surface roots. | `EXTRACT` owner-specific composition without creating a universal controller. | Generic, Entity, Assessment, tests, boundaries, tracker. | None. |
| G19 | Structural casts conceal mismatches at Timeline protocol boundaries. | `REPLACE` assertions with exact construction and validation. | Timeline adapters/models, `package.view_contracts`, tests, tracker. | Malformed responses fail closed. |

`REMOVE` and `PRIVATIZE` require a zero-consumer or same-file proof.
`RETAIN` requires a named production owner and a Make-owned verification
route. Test-only use never justifies retention. A generated type or schema is
consumed but never hand-edited. Do not add a suppression, deprecated wrapper,
compatibility alias, wildcard dispatcher, broad command facade, or Markdown
runtime dependency to complete a disposition.

## 43. G14 and G15 — Minimal surface and truthful owner capability

### G14 remediation

Privatize `SemanticInspectorFeatureResolution`,
`InspectorCreateRelatedSubject`, `InspectorRelatedRecordSubject`,
`InspectorRelatedRecordDraftResult`, `GenericMutationSaveState`,
`GenericPatchMutationRequest`, `AssessmentCreationController`, the
same-file-only Workbook Inspector owner/reset types, and the two edit-recovery
message constants when WI-16 reconfirms their current consumer graph. The
reset-scope algorithm may remain independently testable only if it moves to a
pure module that the production coordinator also imports; otherwise test it
through coordinator behavior. Do not preserve any removed export through a
barrel, alias, or deprecation wrapper.

This change reduces accidental package contracts without changing runtime
behavior. Leaving the exports in place lets tests and future callers turn
implementation details into permanent compatibility obligations. Completion
requires removed-symbol searches, a production owner for every remaining
export, frontend typecheck, focused owner tests, import-boundary checks, and a
scoped Fallow report with no G14 finding.

### G15 remediation

Keep `resolveSemanticInspectorFeature` responsible only for exact tuple
admission and canonical config-owned object identity. Add owner capability
resolution over the admitted feature with the closed outcomes
`create_related`, `indicator`, `record_history`, `existing_owner_control`, and
`unsupported`. A `record_history` result includes exactly one action:
`delete`, `restore`, or `rollback`.

Each of Generic, Entity, Assessment, and Timeline MUST use its owner
resolution both to derive visible invokable controls and to dispatch an
invocation. `existing_owner_control` remains implemented by its existing
panel-local owner and is not duplicated as a contextual action. An admitted
feature with no current owner resolution is `unsupported` and omitted without
affecting registered siblings. Delete the four surface-local route-kind
filters and all divergent click-time reclassification.

The semantic registry remains unchanged at 247 tuples unless the live corpus
itself changes under separate authority. This is an internal TypeScript
boundary migration; it changes no contract or payload. The long-term seam for
a new feature is exact registration plus one explicit owner capability, not a
new filter or universal action controller. Completion requires corpus equality,
one capability result per tuple, render/dispatch parity, unsupported omission,
canonical identity, and zero remaining surface-local route filters.

## 44. G16 — Shared core record-history capability

Introduce one pure record-history model with:

- a subject union of `live` and `deleted`, each containing `recordId` and
  `rowVersion`;
- an exhaustive rollback-target union of `history_entry`, `change_set`, and
  `row_restore` with only its required server-authored selector;
- `idle`, `loading`, `ready`, and `submitting` phases;
- loaded history, a typed local error, an optional pending confirmation, and
  optional typed local feedback;
- owner-created opaque request and operation IDs; and
- exhaustive load, preview, cancel, submit, reject, accept, retarget, and
  clear transitions that commit only for the matching subject and ID.

Use one shared record-history presentation for metadata, loading, failures,
delete/restore controls, rollback actions, confirmations, technical details,
and local feedback. Controls are enabled only when the matching registered
`record_history` capability is present, the server advertises the rollback
selector where applicable, authorization permits mutation, and the subject
state permits the action.

Generic, Entity, and Assessment share a narrow controller over their existing
`RecordRouteCommandPort`. Timeline adapts its existing history controller to
the same pure state and presentation but retains socket transaction tracking,
save-state transitions, committed-record waiting, viewport continuity, and
load ordering. Do not merge those command sequences.

After accepted deletion, retain only a `deleted` history subject with the
accepted tombstone row version, clear row details and workflows, and render
only the History panel. Restore uses that tombstone version, refreshes the
owner once, reselects the restored record ID, and returns to a live subject.
Rollback uses only a server-advertised target, refreshes according to the
owner sequence, and reloads current history. Once a mutation crosses its port,
complete its captured owner effects; only matching current IDs may update
inspector state or feedback.

Routes and wire payloads remain unchanged. The intentional compatibility
changes are non-Timeline restore availability and rejection of stale history
results. Leaving the gap unresolved preserves misleading declared actions,
inconsistent confirmation behavior, and stale-result races. Completion
requires all three actions on all 17 schemas; live and tombstone row-version
tests; duplicate/stale load and mutation tests; authorization-loss tests;
server-selector tests; and exact Timeline link, continuity, save-state, and
refresh evidence.

## 45. G17 and G18 — Feedback semantics and owner composition

### G17 remediation

Change the generic related-workflow callback to accept
`WorkbookInspectorFeedback | null`. Begin-time and operation failures use the
typed error variant. A successful related-record creation keeps its current
copy, neutral styling, and `announcement: "none"`. Standalone Assessment
creation success is a neutral message with `announcement: "polite"`.

Record-history completion produces local neutral, polite feedback with the
existing rollback copy and explicit delete/restore completion copy. No success
may pass through `WorkbookInspectorPublicError`. Entity merge errors, mutation
errors, and ordinary inspector feedback become distinct state so a precedence
expression cannot coerce messages into errors. Field-local related-form errors
remain in the related-form model.

This corrects internal and accessibility semantics without changing routes,
payloads, technical error details, or failure urgency. Completion requires
compile-time rejection of raw feedback strings; constructor and renderer
tests; related success with no live announcement; Assessment and history
success with polite announcement; assertive failures; and zero success paths
through the public-error renderer.

### G18 remediation

Extract owner-specific Generic, Entity, and Assessment inspector composition
into dedicated modules. Each owner module receives narrow subject, capability,
authorization, mutation, refresh, selection, focus/reset, and presentation
ports. It owns panel assembly, capability selection, contextual dispatch,
related workflow presentation, record history, Indicator composition where
applicable, and inspector-local feedback.

The surface roots retain grid/query composition, cell editing, persistence
commands, selection authority, mutation runtime, and owner-specific non-
inspector workflows. Do not introduce a shared mega-prop object, a universal
inspector controller, or broad surface decomposition. Preserve DOM contracts,
test IDs, panel order, focus/reset behavior, and visuals.

Completion requires no inspector route filtering or panel assembly in the
three surface roots, independent focused evidence for each owner module,
passing import boundaries, and unchanged functional and ordinary visual
evidence. The long-term benefit is a cohesive inspector extension boundary
without coupling unrelated grid and mutation behavior.

## 46. G19 — Timeline adapter type integrity

Remove the scoped production structural assertions through exact typed
construction:

1. Make the private registry validator generic over its string-literal input,
   validate membership, and return the input rather than the registry's
   widened string. Exported built-in view-schema constants consequently retain
   source-compatible literal types.
2. Validate and destructure clipboard columns and targets into non-empty
   tuples. Use the canonical Workbook same-field-conflict payload and one
   parser that returns that type; validate the conflicts collection without
   an array cast.
3. Type the mention request builder to the generated request-compatible shape.
   Read the generated mention resource through field validators without an
   `unknown` round-trip and reject invalid row-version or entity shapes.
4. Make Evidence create and patch builders return their exact request type or
   `null`. Split create and patch execution branches so TypeScript narrows the
   request and accepted response without a request cast.
5. Carry the exhaustive rollback-target union through the Timeline history
   port and structurally assign it to the generated rollback request.

Literal `as const` expressions that establish discriminants are not the
target; `as unknown as`, generated-request assertions, and collection-shape
assertions are. Generated sources remain untouched. Runtime requests must
remain byte-for-byte compatible, while malformed responses newly fail closed.
Completion requires zero scoped structural assertions, exact adapter request
tests, negative response tests, `package.view_contracts` evidence, frontend
typecheck, and Timeline owner tests.

## 47. Third successor workstream ledger and tracker protocol

| Workstream | Status | Dependency | Binary exit condition |
| --- | --- | --- | --- |
| WI-16 — Authority, baseline, inventory, and characterization | DONE | Separate implementation authorization | Complete owner/inventory evidence, passing characterization, and no owner contradiction. |
| WI-17 — Minimal exports and truthful capability resolution | DONE | WI-16 `DONE` | Remaining exports are production-owned and one owner resolution drives render and dispatch. |
| WI-18 — Shared core record-history capability | DONE | WI-17 `DONE` | Shared pure state/presentation serves all four owners and all 17 schemas implement the core action triple. |
| WI-19 — Feedback semantic correction | DONE | WI-18 `DONE` | Every success and failure uses its typed feedback and announcement path. |
| WI-20 — Owner-specific inspector composition | DONE | WI-19 `DONE` | Three surface roots delegate inspector composition through narrow owner boundaries. |
| WI-21 — Timeline adapter type integrity | DONE | WI-20 `DONE` | Scoped adapters construct and validate exact contract shapes without structural casts. |
| WI-22 — Final validation and handoff | DONE | WI-21 `DONE` | All routed and terminal evidence passes and C026-C035 are resolved. |

Only one workstream may be `IN_PROGRESS`. Immediately before beginning a
workstream, append its refreshed checkpoint and change only that workstream to
`IN_PROGRESS`. After its exit criteria pass, append paths, decisions,
commands, run roots, failures and classifications, compatibility, rollback,
residual risks, and next action; mark it `DONE`; only then mark its successor
`IN_PROGRESS`. A required failure leaves the workstream `IN_PROGRESS`.
`BLOCKED` is reserved for a verified owner contradiction and must name both
clauses.

## 48. WI-16 — Authority, baseline, inventory, and characterization

1. Refresh repository and nested instructions, branch, commit, upstream
   relation, Git state, toolchain, generated policy, source ownership, task
   guides, and the live contract corpus.
2. Reconfirm every G14 candidate and every retained static-analysis false
   positive by direct production-consumer search. Classify each candidate
   `REMOVE`, `PRIVATIZE`, `CONSOLIDATE`, `EXTRACT`, or `RETAIN`.
3. Produce a tuple-to-owner-capability inventory for all 247 feature groups
   and prove all 17 schemas contain the core history action triple.
4. Inventory the four render filters and matching execution predicates, both
   history controller families, every related/Assessment/history/Entity
   feedback producer and renderer, the three embedded inspector compositions,
   and every scoped adapter assertion.
5. Strengthen documentation-free pre-change characterization for canonical
   identity, supported omission, render/dispatch parity, live and deleted
   history subjects, delete/restore/rollback versions, stale loads and
   mutations, feedback copy/roles, owner focus/reset, and exact adapter wire
   shapes.
6. Run Fallow with an owner-only process umask and record scoped findings
   separately from the repo-wide advisory count. Add no suppression.
7. Confirm no normative owner edit is required or invoke the contradiction
   protocol from section 41.

**Exit:** The inventory is complete, current behavior is characterized by
passing evidence, every retained interface has a production owner and Make
route, and no owner contradiction remains.

**Tracker gate:** Mark WI-16 `DONE`, append all evidence, then mark WI-17
`IN_PROGRESS` before changing production interfaces.

## 49. WI-17 — Minimal exports and truthful capability resolution

1. Implement G14 and G15 atomically so removal cannot leave a second feature
   policy behind.
2. Add corpus-driven owner-capability tests before deleting the four render
   filters and divergent execution predicates.
3. Preserve exact semantic registration and config-owned object identity.
   Owner capability consumes only an admitted canonical feature.
4. Keep existing-owner controls panel-owned, core actions history-owned, and
   related/Indicator actions with their current owner controllers.
5. Privatize the confirmed dead surface without aliases, wrappers, barrels,
   or suppressions.
6. Run focused `web.workbook`, `web.architecture`, `module.timeline`,
   `module.entities`, `module.assessments`, and `module.indicators` evidence,
   plus typecheck, import boundaries, Biome, removed-symbol searches, and
   scoped Fallow.

**Primary risks:** Treating semantic admission as execution, omitting a live
owner capability, duplicating an existing-owner control, or retaining a
test-only compatibility wrapper.

**Exit:** Every live tuple has its intended closed capability result, rendered
actions have exactly one execution path, registered-but-ownerless features
are omitted individually, and the G14 surface is private or removed.

**Tracker gate:** Record capability counts, removed symbols, reports, run
roots, compatibility, and rollback; mark WI-17 `DONE`, then WI-18
`IN_PROGRESS`.

## 50. WI-18 — Shared core record-history capability

1. Add the pure subject, rollback-target, state, action, and reducer model
   described in section 44, with exhaustive tests for every legal and illegal
   transition.
2. Replace the duplicate history presentations with one shared component
   driven by registered owner capabilities and server-advertised selectors.
3. Migrate Generic, Entity, and Assessment to the simple command adapter and
   retained tombstone lifecycle. Preserve each owner's refresh, reselection,
   mutation-state, focus, and reset responsibilities.
4. Migrate Timeline state and presentation while retaining its existing
   socket, committed-record, viewport-continuity, save-state, and refresh
   command sequence.
5. Add deferred-promise evidence for stale loads, stale rejections, stale
   acceptances, cancel/reopen, subject retarget, row-version retarget, and an
   accepted captured owner sequence.
6. Prove delete uses a live version, restore uses the accepted tombstone
   version, rollback targets contain only their required selectors, and a
   restored record is refreshed and reselected once.
7. Run focused and service-backed Workbook, Timeline, Entity, Assessment, and
   applicable design routes returned by refreshed task guides.

**Primary risks:** Losing a tombstone after selection clears, applying an old
operation to a new subject, duplicating Timeline refresh/socket effects, or
constructing a rollback target not advertised by the server.

**Exit:** All four owners share the pure model and presentation, all 17 schema
action triples execute through their real owner, stale results cannot affect a
new subject, and Timeline sequencing remains exact.

**Tracker gate:** Record reducer cases, owner paths, action coverage, run
roots, compatibility, and rollback; mark WI-18 `DONE`, then WI-19
`IN_PROGRESS`.

## 51. WI-19 — Feedback semantic correction

1. Implement the G17 producer and state migrations without adding a second
   renderer or primitive compatibility callback.
2. Preserve related-record success copy and render it as non-announced neutral
   feedback.
3. Render standalone Assessment success and record-history completion as
   neutral polite feedback at their local locus.
4. Keep all operation failures typed, assertive, and capable of rendering
   current technical details. Keep related-form validation local to the form.
5. Separate Entity merge, mutation, and ordinary inspector feedback state and
   remove the overloaded public-error precedence path.
6. Add unit and accessibility evidence for copy, style, test IDs, roles,
   announcement policy, retarget/reset, and failure retention.
7. Run `web.workbook`, `web.design`, `module.timeline`, `module.entities`,
   `module.assessments`, frontend checks, accessibility, and ordinary visual
   evidence.

**Primary risks:** Copy drift, duplicate announcements, lost technical detail,
or clearing an actionable failure during an unrelated feedback update.

**Exit:** Raw feedback strings are rejected, no success uses the public-error
renderer, announcement policy matches section 45, failures remain assertive,
and no unintended visual change exists.

**Tracker gate:** Record every migrated producer and renderer, accessibility
and visual outcomes, run roots, compatibility, and rollback; mark WI-19
`DONE`, then WI-20 `IN_PROGRESS`.

## 52. WI-20 — Owner-specific inspector composition

1. Extract Generic inspector composition behind a narrow owner module while
   leaving Generic query, grid, edit, and persistence mechanics in the surface
   root.
2. Extract Host/Identity inspector composition behind the Entity owner module
   while leaving entity grid, merge authority, alias editing, and persistence
   in their current owners.
3. Extract Assessment inspector composition behind the Assessment owner
   module while preserving append-only creation and selection authority.
4. Move panel mapping, owner capability usage, contextual dispatch, related
   workflow presentation, record history, Indicator composition, and local
   feedback into those modules.
5. Use explicit narrow props/ports. Reject a shared universal controller,
   mega-prop facade, or broad grid/mutation decomposition.
6. Preserve DOM/test-ID contracts, panel order, authorization, focus/reset,
   payloads, refresh counts, and visuals through focused and browser evidence.

**Primary risks:** Moving selection authority into presentation, creating a
large cross-owner facade, changing focus restoration, or accidentally
reordering inspector content.

**Exit:** The three surface roots contain no inspector route filtering or
panel assembly, owner modules are independently testable, import boundaries
pass, and functional/accessibility/visual behavior remains compatible.

**Tracker gate:** Record moved responsibilities, retained owners, run roots,
visual outcome, compatibility, and rollback; mark WI-20 `DONE`, then WI-21
`IN_PROGRESS`.

## 53. WI-21 — Timeline adapter type integrity

1. Implement the five structural changes in section 46 without editing a
   generated root.
2. Add compile-time and runtime evidence for non-empty clipboard structures,
   canonical conflict payloads, mention request/response validation, exact
   Evidence create/patch branches, and exhaustive rollback targets.
3. Confirm built-in view-schema constants retain literal types and all
   existing package consumers typecheck without a compatibility cast.
4. Preserve exact accepted request bodies and response application for valid
   data; add negative cases that reject malformed data before owner state is
   updated.
5. Run `package.view_contracts`, `module.timeline`, `module.evidence`,
   `web.architecture`, frontend typecheck/unit, import boundaries, Biome, and
   structural-cast searches.

**Primary risks:** Replacing a cast with a second untyped builder, widening a
literal elsewhere, subtly changing request serialization, or accepting a
partial response.

**Exit:** No scoped production structural assertion remains, valid wire
behavior is unchanged, invalid shapes fail closed, and package plus owner
evidence passes.

**Tracker gate:** Record every eliminated assertion, type refinement, negative
case, run root, compatibility statement, and rollback; mark WI-21 `DONE`, then
WI-22 `IN_PROGRESS`.

## 54. WI-22 — Final validation, acceptance, and handoff

Refresh task guides for these expected owners, adding no owner unless the live
dependency cone proves it necessary:

```text
make task-guide ROLE=module-author OWNER=web.workbook
make task-guide ROLE=module-author OWNER=web.architecture
make task-guide ROLE=module-author OWNER=web.design
make task-guide ROLE=module-author OWNER=module.workbook
make task-guide ROLE=module-author OWNER=module.timeline
make task-guide ROLE=module-author OWNER=module.entities
make task-guide ROLE=module-author OWNER=module.evidence
make task-guide ROLE=module-author OWNER=module.assessments
make task-guide ROLE=module-author OWNER=module.indicators
make task-guide ROLE=module-author OWNER=package.view_contracts
```

Run every returned focused and service-backed slice. Update authored source
ownership and test-catalog inputs for new tests, then run `make generate` and
verify every changed generated output has Make-owned provenance. Run
`make format`, followed by `make agent-finalize` before terminal verification.
Pass `RESULTS_DIR` only for a real successful retained warm-run root.

Run the terminal matrix:

```text
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
```

Run scoped Fallow under an owner-only process umask. Audit final Git scope,
source ownership, generated provenance, corpus/capability equality, core
action coverage, removed exports, surface-root boundaries, scoped casts, and
the absence of Markdown runtime dependencies, TODOs, aliases, deprecated
wrappers, suppressions, wildcard dispatch, and parallel legacy paths.

WI-22 must classify every row as `PASS`, justified `N/A`, or evidenced
`BLOCKED`:

| Acceptance | Planning result | Required evidence |
| --- | --- | --- |
| C026 — Authority and bounded scope | PASS | Core 01, Core 03, Core 04, domain, and design ownership was reconfirmed without contradiction. Changes remain within internal TypeScript, focused evidence, authored ownership/catalog inputs, one Make-generated topology digest, and this ledger; no normative owner or unplanned subsystem changed. |
| C027 — Minimal package surface | PASS | All G14 candidates are private or removed without aliases, wrappers, barrels, or suppressions. Final scoped Fallow has no finding in the implementation cone; its repo-wide advisory findings remain explicitly deferred. |
| C028 — Truthful capability | PASS | Corpus-driven evidence proves 247 exact tuples across 17 schemas, one intended closed capability per tuple, canonical config identity, render/dispatch parity, and sibling-safe omission. Root policy tests prevent owner capability dispatch from returning to broad surface roots. |
| C029 — Complete core history actions | PASS | All 17 schemas expose delete, restore, and rollback through shared presentation and real owner adapters; rollback accepts only server-advertised `history_entry`, `change_set`, or `row_restore` selectors. |
| C030 — Tombstone and stale-result safety | PASS | Reducer, deferred-promise, owner, and browser evidence covers positive live/tombstone versions, history-only deleted subjects, restore reselection, cancel/reopen, retarget, duplicate work, and stale load/mutation rejection while captured accepted owner effects finish. |
| C031 — Typed feedback semantics | PASS | Producers accept only `WorkbookInspectorFeedback | null`; related success is neutral and silent, Assessment/history success is neutral and polite, failures remain assertive with technical detail, and success never enters `WorkbookInspectorPublicError`. |
| C032 — Owner surface cohesion | PASS | Generic, Entity, and Assessment roots delegate panel assembly, capability selection and dispatch, related/history presentation, applicable Indicator composition, and local feedback to narrow owner modules. Static policy forbids shell, panel, and capability-dispatch logic in those roots; no universal facade exists. |
| C033 — Adapter type integrity | PASS | All 11 scoped assertions are absent. Literal-preserving registry validation, non-empty clipboard validation, direct mention validation, exact Evidence request variants, canonical conflicts, and exhaustive rollback targets preserve valid wire behavior and reject malformed data before state application. |
| C034 — Compatibility and provenance | PASS | Public routes, valid payloads, schemas, authorization, persistence, dependencies, generated contracts, stored data, DOM/test IDs, focus, and ordinary pixels remain compatible. The only generated change is the Make-produced topology render-index digest; no image or golden changed. Intended corrections are isolated omission, non-Timeline restore availability, stale-result rejection, typed announcements, and malformed-response rejection. |
| C035 — Verification and terminal handoff | PASS | Every applicable focused, service-backed, terminal, Fallow, finalize, Markdown, and diff gate passes; all failures are classified and resolved, WI-16-WI-22 are `DONE`, and the final checkpoint is recorded. |

No visual change is expected. Treat a difference as a regression unless an
owner/design decision explicitly authorizes it. Any intentional update must
follow the visual-golden guide, include manual review of every changed image,
and pass two fresh ordinary visual proofs.

After the terminal tracker edit, run:

```text
make lint-markdown
git diff --check
```

Mark WI-22 `DONE` only after C026-C035, all run roots, failure
classifications, compatibility, rollback, residual risks, deferrals, and the
next safe extension seam are recorded.

## 55. Third successor checkpoint, compatibility, rollback, and deferrals

| Time | Workstream | Paths and decisions | Commands and results | Compatibility, rollback, risks, next action |
| --- | --- | --- | --- | --- |
| `2026-09-01T17:33:08-04:00` | Third successor planning; WI-16 not started | Appended sections 41-55 to this handoff and preserved sections 1-40 as completed history. Recorded G14-G19, WI-16-WI-22, C026-C035, the clean baseline, 17-schema/247-feature corpus, core history action declarations, four duplicated filters, 11 scoped adapter assertions, explicit static-analysis exclusions, validation, compatibility, rollback, and deferrals. `docs/research/nlspec-spec.md` was advisory only; no normative document was edited. | Planning inspection confirmed `main` at `fa784e85d7e7f14a56e1548958272dd76d321429`, equal to `origin/main`, and the prior owner-only-umask Fallow PASS at `.cartulary/test-results/20260901T211002Z-p5486` with 190 advisory repo-wide findings. Pre-tracker `git diff --check` passed; `make lint-markdown` passed at `.cartulary/test-results/20260901T213609Z-p16022`. | Documentation only; no product, test, contract, generated, fixture, golden, route, API, payload, schema, authorization, persistence, Evidence, dependency, or stored-data behavior changed. This handoff addition is the sole rollback unit. Next: obtain separate implementation authorization, refresh the baseline, and mark only WI-16 `IN_PROGRESS`. |
| `2026-09-01T17:49:35-04:00` | WI-16 started | Received separate implementation authorization. Refreshed repository instructions, Git/upstream state, generated-artifact policy, live inspector corpus, initial source ownership, and Make-owned verification routing. Sections 1-40 remain unchanged; no normative owner edit is authorized. | `main` remains at `fa784e85d7e7f14a56e1548958272dd76d321429`, equal to `origin/main`; the only pre-existing change is this staged handoff. `git diff --cached --check` passed. The machine projection contains 17 schemas, 247 feature keys, and no missing delete/restore/rollback triple. Task guides confirmed focused routes for `web.workbook`, `module.timeline`, and `package.view_contracts`; the remaining owner guides, inventories, characterization, and scoped Fallow are in progress. | Tracker-only transition. No product, test, contract, generated, fixture, golden, route, API, payload, schema, authorization, persistence, Evidence, dependency, or stored-data behavior changed. WI-16 is the sole `IN_PROGRESS` workstream. |
| `2026-09-01T17:56:58-04:00` | WI-16 complete; WI-17 started | Reconfirmed the adopted Core 01/Core 03/Core 04 and domain/design boundaries with no contradiction. The inventory remains 17 schemas, 247 semantic registrations, four owner-local render filters plus matching dispatch classification, two history controller/presentation families, three embedded owner compositions, typed-feedback exceptions, and exactly 11 scoped adapter assertions. Direct searches proved every G14 candidate same-file or test-only; retained collaboration, mutation-runtime, and pending-queue members remain explicit out-of-scope false positives. | PASS: `web.workbook` 145/145 `.cartulary/test-results/20260901T215034Z-p24737`; `web.architecture` 12/12 `.cartulary/test-results/20260901T215034Z-p24745`; `package.view_contracts` 5/5 `.cartulary/test-results/20260901T215034Z-p24736`; `module.entities` 42/42 `.cartulary/test-results/20260901T215048Z-p27095`; `module.assessments` 28/28 `.cartulary/test-results/20260901T215048Z-p27106`; `module.timeline` 53/53 `.cartulary/test-results/20260901T215048Z-p27131`; `module.indicators` 20/20 `.cartulary/test-results/20260901T215048Z-p27150`; owner-only-umask Fallow `.cartulary/test-results/20260901T215618Z-p1228` retained the expected 190 advisory dead-code findings and no suppression. | Characterization and tracker only; no product, contract, generated, fixture, golden, route, API, payload, schema, authorization, persistence, Evidence, dependency, or stored-data behavior changed. WI-16 is independently revertible through this tracker row. WI-17 is now the sole `IN_PROGRESS` workstream; next: privatize the confirmed surface and make one owner capability result drive rendering and dispatch. |
| `2026-09-01T18:25:50-04:00` | WI-17 complete; WI-18 started | Privatized every confirmed G14 same-file/test-only type, helper, and recovery constant without an alias or wrapper. Added a closed owner-capability union over exact semantic admission, including typed history actions; all four surfaces now derive contextual rendering from canonical capability objects and dispatch those same objects. Removed four surface-local route filters, click-time route reclassification, and the redundant Timeline feature resolver. Updated the authored Timeline policy test selector to the replacement target-contract ownership evidence. The live registry remains 247 exact tuples across 17 schemas, with 17 delete, 17 restore, and 17 rollback capabilities; unknown additive tuples are omitted without disabling siblings. | PASS: format `.cartulary/test-results/20260901T220757Z-p5993`; typecheck `.cartulary/test-results/20260901T220843Z-p11192`; `web.workbook` 145/145 `.cartulary/test-results/20260901T220900Z-p11767`; `web.architecture` 12/12 `.cartulary/test-results/20260901T220946Z-p27734`; `module.entities` 42/42 `.cartulary/test-results/20260901T220946Z-p27777`; `module.assessments` 28/28 `.cartulary/test-results/20260901T220946Z-p27799`; `module.indicators` 20/20 `.cartulary/test-results/20260901T220946Z-p27827`; `module.timeline` 53/53 `.cartulary/test-results/20260901T222058Z-p65463`; import boundaries `.cartulary/test-results/20260901T221944Z-p63211`; Biome `.cartulary/test-results/20260901T222023Z-p64163`; owner-only-umask Fallow `.cartulary/test-results/20260901T222033Z-p64660` with 188 advisory dead-code findings and no G14 candidate. `git diff --check` and removed-symbol/filter searches passed. The first typecheck failed at `.cartulary/test-results/20260901T220809Z-p10246` because one test import used the wrong relative path; introduced and fixed. The first Biome run failed at `.cartulary/test-results/20260901T221944Z-p63229` on that test's import order; introduced and fixed. A parallel Timeline run was manually cancelled at 51/53 (`.cartulary/test-results/20260901T220946Z-p27749`) after cross-run browser contention; the first sequential rerun then failed only the pre-existing blank-row measurement qualification at `.cartulary/test-results/20260901T221444Z-p4453`; unchanged rerun passed 53/53. | Internal TypeScript ownership and authored test routing only. Public routes, payloads, schemas, authorization, persistence, generated contracts, stored data, DOM/test IDs, and visuals are unchanged. The intended correction is that only capabilities with a real contextual owner render, while config discovery and registered siblings remain intact. Rollback is atomic across the capability dispatcher, four consumers, private export cleanup, Timeline policy simplification, focused tests, catalog selector, and this row; no data rollback is required. Residual history duplication now moves to WI-18, which is the sole `IN_PROGRESS` workstream. Next: introduce the shared pure history state/presentation and retain owner-specific command sequencing. |
| `2026-09-01T19:39:21-04:00` | WI-18 complete; WI-19 started | Added `workbookRecordHistoryModel.ts`, its exhaustive reducer tests, the narrow simple-owner controller, and one shared history presentation. Generic, Entity, and Assessment now retain accepted tombstones, clear live inspector work on delete, restore from the accepted version, refresh/reselect once, and reject stale load/mutation commits while completing captured owner effects. Timeline now uses the same pure state and presentation while retaining its private socket, save-state, committed-record, viewport-continuity, selection, and load sequence. Exact rollback selectors flow through the shared union and the Timeline port. The presentation preserves the prior Timeline idle, spacing, confirmation order/copy, focus restoration, DOM IDs, and ordinary goldens. Added authored source ownership for the new production and test paths. | PASS: format `.cartulary/test-results/20260901T233056Z-p75208`; typecheck `.cartulary/test-results/20260901T233206Z-p23242`; `web.workbook` 145/145 `.cartulary/test-results/20260901T233255Z-p26362`; `web.architecture` 12/12 `.cartulary/test-results/20260901T233255Z-p26355`; `module.assessments` 28/28 `.cartulary/test-results/20260901T225748Z-p99457`; `module.indicators` 20/20 `.cartulary/test-results/20260901T225852Z-p43207`; full `module.timeline` 53/53 `.cartulary/test-results/20260901T233340Z-p42588`; focused Workbook history visual route 11/11 `.cartulary/test-results/20260901T233100Z-p79355`; focused Entity visual rerun 11/11 `.cartulary/test-results/20260901T233824Z-p2553`; import boundaries `.cartulary/test-results/20260901T233206Z-p23270`; Biome `.cartulary/test-results/20260901T233206Z-p23307`; `git diff --check`. Introduced and fixed failures: nullable Timeline subject and exact-optional props in typecheck roots `.cartulary/test-results/20260901T230439Z-p24470`, `.cartulary/test-results/20260901T230553Z-p45114`, and `.cartulary/test-results/20260901T231251Z-p22361`; empty initial Timeline history section in `module.workbook` `.cartulary/test-results/20260901T225942Z-p60875`; source-owner ordering in `web.architecture` `.cartulary/test-results/20260901T233206Z-p23141`; and visual compatibility deltas successively isolated at `.cartulary/test-results/20260901T231425Z-p43294`, `.cartulary/test-results/20260901T231812Z-p91769`, `.cartulary/test-results/20260901T232118Z-p40162`, `.cartulary/test-results/20260901T232304Z-p87591`, and `.cartulary/test-results/20260901T232424Z-p35487` through `.cartulary/test-results/20260901T232842Z-p31583`. The Entity chip golden failures at `.cartulary/test-results/20260901T225545Z-p46748` and `.cartulary/test-results/20260901T230638Z-p50397` were unrelated renderer capture noise; unchanged focused rerun passed and no image changed. | Internal TypeScript state, controller, presentation, owner integration, focused tests, and authored ownership only. Routes, request/response bodies, schemas, authorization, persistence, generated contracts, dependencies, stored data, and goldens are unchanged. Intended corrections are restore availability for non-Timeline owners and rejection of stale history commits; successful history feedback is already typed and polite for the WI-19 migration. Rollback is source-only across the shared model/controller/presentation, four owner integrations, Timeline adapter/state migration, tests, ownership input, and this row. Residual risk is limited to the still-authorized Timeline history target assertion, which remains for WI-21. WI-19 is now the sole `IN_PROGRESS` workstream; next: migrate every remaining success/error producer to typed feedback without a primitive callback or second renderer. |
| `2026-09-01T19:53:32-04:00` | WI-19 complete; WI-20 started | Changed the related-workflow port to `WorkbookInspectorFeedback | null`; the producer now emits typed assertive local failures and preserves `Created <title> record <id>.` as neutral, non-announced feedback. Standalone Assessment creation now emits neutral polite `Assessment created.` feedback. History already emits polite delete/restore/rollback feedback from WI-18. Entity merge, mutation error, and ordinary action feedback are independent states; merge and paste successes are neutral while operation failures retain typed technical detail. All four loci use the existing shared feedback renderer, and tests assert constructors, exact copy, announcement policy, stale completion, and error roles. | PASS: format `.cartulary/test-results/20260901T234815Z-p52675`; typecheck `.cartulary/test-results/20260901T234819Z-p56845`; `web.workbook` 145/145 `.cartulary/test-results/20260901T234843Z-p57453`; `web.design` 15/15 `.cartulary/test-results/20260901T234843Z-p57459`; `module.assessments` 28/28 `.cartulary/test-results/20260901T235017Z-p21175`; `module.entities` 42/42 `.cartulary/test-results/20260901T235118Z-p64547`; import boundaries `.cartulary/test-results/20260901T235317Z-p16278`; Biome `.cartulary/test-results/20260901T235317Z-p16293`; typed-producer/public-error searches and `git diff --check`. The first typecheck failed at `.cartulary/test-results/20260901T234713Z-p51803` because three retained Entity mutation-error calls needed their existing constructor import and Generic lacked a local neutral style; introduced and fixed. No golden changed, and the ordinary design and Entity visual evidence passed without update. | Internal TypeScript feedback types, state separation, shared-renderer consumption, and focused tests only. Copy, test IDs, technical failure details, assertive error behavior, routes, payloads, schemas, persistence, authorization, dependencies, generated contracts, and stored data remain compatible. Announcement changes are intentional: related success is silent-neutral, while Assessment and history success are polite-neutral. Rollback is source-only across the feedback constructor, related/Assessment/Entity producers and consumers, tests, and this row. WI-20 is now the sole `IN_PROGRESS` workstream; next: extract three narrow owner-specific inspector composition modules while retaining grid, query, selection, mutation runtime, and persistence in their roots. |
| `2026-09-01T20:10:38-04:00` | WI-20 complete; WI-21 started | Added dedicated Generic, Entity, and Assessment inspector composition modules with explicit subject, capability, authorization, history, related-workflow, owner-content, feedback, and close/action ports. The modules now own shell and panel assembly, capability selection, contextual controls, shared history, related-workflow presentation, Indicator composition where applicable, and local feedback. The broad roots retain grid/query state, persistence and mutation commands, selection, refresh effects, and owner-specific edit/merge/create controls supplied as explicit content. Added static boundary evidence proving all three roots delegate and cannot reintroduce shell or panel assembly; registered all new production paths in authored frontend source ownership. Existing DOM/test IDs, panel order, focus/close behavior, feedback placement and styling were preserved. | PASS: format `.cartulary/test-results/20260902T000347Z-p26095`; typecheck `.cartulary/test-results/20260902T000213Z-p24912`; `web.architecture` 12/12 `.cartulary/test-results/20260902T000408Z-p31260`; `web.workbook` 145/145 `.cartulary/test-results/20260902T000408Z-p31304`; `module.assessments` 28/28 `.cartulary/test-results/20260902T000408Z-p31283`; `module.entities` 42/42 `.cartulary/test-results/20260902T000408Z-p31314`; import boundaries `.cartulary/test-results/20260902T000646Z-p46040`; Biome `.cartulary/test-results/20260902T000646Z-p46057`; accessibility 12/12 `.cartulary/test-results/20260902T000656Z-p46924`; ordinary visual 12/12 `.cartulary/test-results/20260902T000834Z-p91720`; root-assembly searches found no inspector shell, panel map, contextual resolver, related presentation, or history presentation in the three roots. The first post-extraction typecheck failed at `.cartulary/test-results/20260902T000137Z-p23925` only for an introduced unused style left by the move; removed and the rerun passed. No image or golden changed. | Internal TypeScript composition, static architecture evidence, authored source ownership, and tracker only. Routes, payloads, schemas, persistence, authorization, dependencies, generated contracts, stored data, test IDs, DOM behavior, focus/reset, and pixels remain compatible. Rollback is source-only across the three owner modules, their root delegation, the architecture assertion, ownership input, and this row. The remaining scoped structural assertions are isolated to WI-21, now the sole `IN_PROGRESS` workstream; next: replace each assertion with literal-preserving construction or fail-closed runtime validation without editing generated sources. |
| `2026-09-01T20:28:05-04:00` | WI-21 complete; WI-22 started | Made the private view-schema registry validator generic and literal-preserving. Clipboard construction now validates columns and targets into non-empty tuples, uses the literal Timeline schema ID, and parses conflicts through the canonical same-field payload without collection casts. Mention requests are structurally generated-compatible at the adapter boundary; mention identity, positive row versions, entity type, and optional metadata fields are validated directly. Added one production-owned adapter builder for exact Timeline Evidence create and patch variants, split their execution branches, and retained the exact Evidence-row request variant. The exhaustive history rollback target now assigns directly to the generated request. Removed all 11 scoped structural assertions without replacement casts; generated sources were untouched. Added exact request-body, empty-collection, malformed mention/Evidence/conflict, literal-type, and rollback-target evidence. | PASS: final format `.cartulary/test-results/20260902T002505Z-p71286`; final typecheck `.cartulary/test-results/20260902T002509Z-p75448`; import boundaries `.cartulary/test-results/20260902T002519Z-p75922`; `package.view_contracts` 5/5 `.cartulary/test-results/20260902T002532Z-p76517`; focused Timeline adapter evidence 2/2 `.cartulary/test-results/20260902T002532Z-p76522`; focused Evidence adapter evidence 2/2 `.cartulary/test-results/20260902T002532Z-p76537`; full `module.timeline` 53/53 `.cartulary/test-results/20260902T001834Z-p52227`; full `module.evidence` 36/36 `.cartulary/test-results/20260902T001834Z-p52231`; `web.architecture` 12/12 `.cartulary/test-results/20260902T002617Z-p79766`; frontend unit 398/398 `.cartulary/test-results/20260902T002617Z-p79871`; Biome `.cartulary/test-results/20260902T002617Z-p79916`; scoped structural-cast, removed-builder, protocol-layer, and `git diff --check` searches passed. Introduced and fixed failures: literal preservation exposed two overly narrow test containers in typecheck `.cartulary/test-results/20260902T001519Z-p43090`; the first architecture run `.cartulary/test-results/20260902T001834Z-p52236` correctly rejected protocol imports outside adapters, so exact Evidence builders moved to a production-owned adapter module and the mention model retained a protocol-independent structural payload; the intervening typecheck `.cartulary/test-results/20260902T002401Z-p70442` exposed the remaining builder consumer and exact-optional mismatch; the next architecture run `.cartulary/test-results/20260902T002532Z-p76514` found only authored ownership ordering, which was corrected. | Valid request serialization, operation IDs, routes, payload fields, response application, public types, schemas, persistence, authorization, dependencies, generated contracts, and stored data remain compatible. Malformed adapter responses now intentionally fail closed, and constants retain narrower source-compatible literals. Rollback is source-only across the private registry helper, adapter request/response validation, canonical conflict parsing, Evidence request owner, tests, ownership input, and this row. No generated file or golden changed. WI-22 is now the sole `IN_PROGRESS` workstream; next: refresh all task guides, reconcile authored catalogs/topology through Make, run the full terminal matrix and Fallow audit, resolve C026-C035, finalize, and close the handoff. |
| `2026-09-01T21:48:20-04:00` | WI-22 validation complete; closure gate pending | Refreshed all ten task guides and ran every applicable focused and service-backed slice. Reconciled authored source ownership and the `web.workbook` catalog; `make generate` changed only `tools/execution_topology_render_index.json` through declared Make provenance. The final architecture audit found residual capability dispatch in the three broad roots, so it moved into the already extracted Generic, Entity, and Assessment owner modules and the static ownership policy now rejects root `InspectorContextualCapability`/`capability.kind` use. Final Fallow then identified the same-file `workbookRecordHistorySubjectsEqual` export; it was privatized without an alias. Searches prove no scoped cast, removed export, root panel/dispatch path, Markdown runtime dependency, added TODO/FIXME, suppression, wildcard dispatch, deprecated wrapper, compatibility alias, parallel legacy path, or golden/image change. C026-C034 are `PASS`; C035 awaits the post-ledger Markdown and diff gates. | Generation/provenance PASS: `make generate` `.cartulary/test-results/20260902T002915Z-p20321`; generation drift `.cartulary/test-results/20260902T002936Z-p23436`; generated policy `.cartulary/test-results/20260902T002945Z-p26442`; JSON shape `.cartulary/test-results/20260902T002946Z-p26862`; initial format `.cartulary/test-results/20260902T003000Z-p27411`; initial finalize `.cartulary/test-results/20260902T003004Z-p31537`, with retained-run maintenance skipped because `RESULTS_DIR` was unset. Focused PASS: `web.workbook` 145/145 `.cartulary/test-results/20260902T003034Z-p34680`; `web.architecture` 12/12 `.cartulary/test-results/20260902T003034Z-p34687`; `web.design` 15/15 `.cartulary/test-results/20260902T003034Z-p34678`; `package.view_contracts` 5/5 `.cartulary/test-results/20260902T003034Z-p34702`; `module.workbook` 68/68 `.cartulary/test-results/20260902T003144Z-p83762`; `module.entities` 42/42 `.cartulary/test-results/20260902T003144Z-p83774`; `module.assessments` 28/28 `.cartulary/test-results/20260902T003144Z-p83775`; `module.indicators` 20/20 `.cartulary/test-results/20260902T003144Z-p83796`; `module.evidence` 36/36 `.cartulary/test-results/20260902T003417Z-p50462`; `module.timeline` 53/53 `.cartulary/test-results/20260902T003417Z-p50456`. Service-backed PASS: `web.design` 15/15 `.cartulary/test-results/20260902T003906Z-p58627`; `module.workbook` 39/39 `.cartulary/test-results/20260902T004015Z-p5571`; `module.timeline` 30/30 `.cartulary/test-results/20260902T004226Z-p61057`; `module.entities` 33/33 `.cartulary/test-results/20260902T004716Z-p20002`; `module.evidence` 25/25 `.cartulary/test-results/20260902T004908Z-p71008`; `module.assessments` 19/19 `.cartulary/test-results/20260902T005021Z-p21130`; `module.indicators` 8/8 `.cartulary/test-results/20260902T005125Z-p63771`; `web.workbook`, `web.architecture`, and `package.view_contracts` expose no service-backed rows. Pre-correction terminal PASS: typecheck `.cartulary/test-results/20260902T005218Z-p81442`; unit 398/398 `.cartulary/test-results/20260902T005218Z-p81456`; import boundaries `.cartulary/test-results/20260902T005218Z-p81468`; Biome `.cartulary/test-results/20260902T005218Z-p81509`; accessibility retry 12/12 `.cartulary/test-results/20260902T005640Z-p27008`; measurement 22/22 `.cartulary/test-results/20260902T005814Z-p71677`; stateful 34/34 `.cartulary/test-results/20260902T010244Z-p29537`; support 19/19 `.cartulary/test-results/20260902T010510Z-p79062`; webserver-backed 60/60 `.cartulary/test-results/20260902T010639Z-p23382`; visual 12/12 `.cartulary/test-results/20260902T011223Z-p80836`; `test-fast` 450/450 `.cartulary/test-results/20260902T011423Z-p25797`; owner-only-umask Fallow `.cartulary/test-results/20260902T011449Z-p26704`. The first accessibility attempt failed 10/12 at `.cartulary/test-results/20260902T005233Z-p83022` because service readiness timed out before two groups produced results; classified unrelated infrastructure, and the immediate unchanged retry passed. | Post-audit cohesion PASS: format `.cartulary/test-results/20260902T012129Z-p29923`; typecheck `.cartulary/test-results/20260902T012138Z-p34165`; `web.workbook` 145/145 `.cartulary/test-results/20260902T012157Z-p34855`; `web.architecture` 12/12 `.cartulary/test-results/20260902T012157Z-p34865`; `module.entities` 42/42 `.cartulary/test-results/20260902T012157Z-p34860`; `module.assessments` 28/28 `.cartulary/test-results/20260902T012157Z-p34883`; `module.indicators` 20/20 `.cartulary/test-results/20260902T012432Z-p49457`; import boundaries `.cartulary/test-results/20260902T012432Z-p49585`; Biome `.cartulary/test-results/20260902T012432Z-p49598`. Fresh terminal PASS: generation drift `.cartulary/test-results/20260902T012530Z-p68104`; generated policy `.cartulary/test-results/20260902T012530Z-p68129`; JSON shape `.cartulary/test-results/20260902T012530Z-p68163`; typecheck `.cartulary/test-results/20260902T012530Z-p68603`; frontend unit 398/398 `.cartulary/test-results/20260902T012530Z-p68578`; import boundaries `.cartulary/test-results/20260902T012531Z-p68677`; Biome `.cartulary/test-results/20260902T012531Z-p68760`; accessibility 12/12 `.cartulary/test-results/20260902T012636Z-p93660`; measurement 22/22 `.cartulary/test-results/20260902T012808Z-p38869`; stateful 34/34 `.cartulary/test-results/20260902T013243Z-p96449`; support 19/19 `.cartulary/test-results/20260902T013505Z-p46492`; webserver-backed 60/60 `.cartulary/test-results/20260902T013630Z-p90290`; visual 12/12 `.cartulary/test-results/20260902T014217Z-p48374`; `test-fast` 450/450 `.cartulary/test-results/20260902T014415Z-p92846`. Owner-only-umask Fallow passed at `.cartulary/test-results/20260902T014442Z-p93735` but surfaced the introduced same-file history helper export; after privatization, format `.cartulary/test-results/20260902T014530Z-p94813`, typecheck `.cartulary/test-results/20260902T014542Z-p99298`, `web.workbook` 145/145 `.cartulary/test-results/20260902T014542Z-p99161`, Biome `.cartulary/test-results/20260902T014542Z-p99370`, `test-fast` 450/450 `.cartulary/test-results/20260902T014542Z-p99476`, and final Fallow `.cartulary/test-results/20260902T014715Z-p32969` pass with zero scoped findings; its 205 repo-wide issues are advisory and deferred. Final `agent-finalize` passed at `.cartulary/test-results/20260902T014749Z-p33842`, with retained-run maintenance skipped because `RESULTS_DIR` was unset. Valid wire bodies, routes, schemas, authorization, persistence, dependencies, generated contracts, stored data, DOM/test IDs, focus/reset, and ordinary visuals remain compatible. Intended changes are truthful unsupported omission, non-Timeline restore availability, stale-result rejection, typed success announcements, and malformed-response rejection. No image or golden changed. Rollback is source-only and workstream-granular across the capability model, package-surface removals, shared history, four adapters, typed feedback, owner modules, Timeline adapter refinements, focused evidence, authored ownership/catalog, the Make-generated render-index digest, and this ledger; no data or external rollback exists. Residual risk is ordinary future feature growth. Deferred work remains broad grid/mutation decomposition, non-adapter Timeline casts, repo-wide Fallow cleanup, dependencies, unrelated harness cleanup, and normative specification edits. Next safe extension seam: exact semantic identity, explicit owner capability, real owner control, reset/invalidation participation, and focused evidence—never wildcard dispatch or a universal controller. |
| `2026-09-01T21:51:21-04:00` | WI-22 complete | The terminal ledger resolves C026-C035 and records all scoped paths, decisions, run roots, failure classifications, compatibility, rollback, residual risks, deferrals, and the next extension seam. WI-16 through WI-22 are `DONE`; sections 1-40 remain unchanged. | Post-ledger `make lint-markdown` PASS at `.cartulary/test-results/20260902T015006Z-p37366`; completed-state rerun PASS at `.cartulary/test-results/20260902T015217Z-p39029`; `git diff --check` PASS after each tracker edit. | No additional product or generated change occurred during closure. The complete source-only rollback remains workstream-granular and requires no data migration or external rollback. Handoff complete. |

Append a checkpoint after every workstream and before its successor begins.
Every failure entry must identify the target, retained run root and summary
when available, and whether the failure was introduced or unrelated. Every
visual entry must state whether images changed and record the manual-review
result when they did.

The intended future behavior changes are limited to truthful owner-capability
omission, restore availability on non-Timeline surfaces, rejection of stale
history results, malformed adapter-response rejection, and corrected success
announcement semantics. Routes, wire payloads, schemas, authorization,
persistence, Evidence semantics, dependencies, generated contracts, and
stored data remain compatible. The private view-schema validator's generic
return is a source-compatible type refinement for exported constants.

Each completed workstream is independently source-revertible after its tracker
gate. The complete source-only rollback unit includes the capability model,
package-surface removals, shared history model and presentation, four owner
adapters, typed feedback migration, three owner composition extractions,
Timeline adapter corrections, focused tests, authored ownership/catalog
inputs, their Make-generated topology outputs, and WI-16 through WI-22 ledger
entries. No data rollback or external operation is required.

Broad grid or mutation decomposition, non-adapter Timeline casts, repo-wide
Fallow cleanup, dependency changes, unrelated harness/test-support cleanup,
and normative specification edits remain deferred. A future inspector feature
must add one exact semantic tuple, one explicit owner capability, real owner
control, reset/invalidation participation, and focused evidence. It must not
introduce wildcard dispatch or a universal inspector controller.

For this immediate documentation-only update, run only:

```text
make lint-markdown
git diff --check
```

The third-successor planning compatibility statement is:

**Documentation-only production-readiness planning; no product, test,
contract, generated, fixture, golden, route, API, payload, schema,
authorization, persistence, Evidence, dependency, or stored-data change.**
