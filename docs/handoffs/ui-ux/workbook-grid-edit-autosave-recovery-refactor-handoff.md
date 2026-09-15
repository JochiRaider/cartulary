# Workbook grid edit and autosave recovery refactor

## Control and baseline

This is the controlling GEA-01 through GEA-05 execution record. Only the current
workstream may be IN_PROGRESS. A dependent starts only after its predecessor's
exit evidence is recorded and its status is saved as DONE.

- Authorized scope: existing committed-cell direct-value editing, retained raw
  authoring, autosave settlement, first-dispatch preparation, exact replay,
  accepted-row reconciliation, and their shared Grid Adapter boundaries.
- Execution baseline: clean `main`, HEAD
  `29be6f83f4a6f05d2bd848d47ff4399328d92ac1`; no pre-existing edits.
- Baseline revalidated with `git status --short --branch`, `git rev-parse HEAD`,
  and repository AGENTS.md discovery before creating this file.
- Read order completed: digest README, START_HERE, LOCAL_AGENT_PROMPT, REPO_MAP,
  OWNER_MAP, rules, acceptance, QUERY_RECIPES, and UPSTREAM_MAP. Digest and research
  material are advisory; they do not establish product requirements.
- No commits, pushes, deployments, analyst-data edits, dependencies, browser
  persistence, or new editing capabilities are authorized by this slice.

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| GEA-01 Characterization and contracts | DONE | Owner decisions adopted; focused failures reproduced below. |
| GEA-02 Retained drafts | DONE | Store, scalar intent and Timeline draft tests pass; typecheck passes. |
| GEA-03 Autosave and continuity | DONE | Focused queue, transport, draft, adapter and source tests pass; evidence below. |
| GEA-04 Integrated evidence | DONE | Owning unit/browser/security/regression and four Timeline measurements pass. |
| GEA-05 Terminal validation | DONE | Required terminal checks, completed-handoff formatting and final scope review pass. |

## Owners and decisions

Behavior: Core 01 §3.3.5, §7.4, §18A–B and §19; Core 02 source-field, lifecycle
and revision sections; Core 03 §§3–4, REQ-03-298/299 and §13; Core 04 §§1–2.
Domain vocabulary and design direction remain within their declared boundaries.
Source placement comes from `tools/frontend_source_ownership.json` and
`tools/frontend_import_boundaries.json`; test routing comes independently from
`contracts/verification`, `tools/test_catalog_owner.json`, and `tools/test_families`.

The user approved these owner clarifications during planning and authorized their
implementation:

1. An explicit activation of the original eligible cell restores exact raw text
   inline. Remount alone does not activate an editor or move focus.
2. Original authoring baseline and first-dispatch request base are distinct.
   Current-version preparation is allowed only for unchanged edited fields and
   declared dependencies, or advancement through the work's accepted predecessor.
   Other relevant changes require local review. Uncertain attempts never rebase.
3. Core 03 REQ-03-099's authoritative-acceptance navigation gate remains.
   Admission, mutation acknowledgement, and navigation permission are distinct.
   Explicit sheet/saved-view departure detaches presentation and retains work.

## Lifecycle and identity matrix

| Transition | Owner and identity | Required behavior |
| --- | --- | --- |
| Raw authoring | Workbook runtime; account/incident/client, schema, record, field, editor context, revision | Retain exact input and baseline; no queue or authoritative effect. |
| Local validation | Workbook/source field contract; draft revision | Produce validated intent separately; preserve invalid text and explicit clear. |
| Admission | Existing pending queue; stable unit and contributor revisions | FIFO capacity 64; refusal retains local work; no server acceptance. |
| Unsent coalescing | Queue; contiguous same-record run | Final direct values and ordered collection actions; preserve contributing ownership. |
| First dispatch | Mutation owner and committed-record capability | Guard current authority/dependencies, prepare base, capture request once. |
| Uncertainty | Captured attempt and transaction identity | Replay identical route, body, ID and base; no coalescing into captured work. |
| Rejection | Source validation, conflict queue or terminal queue blocker | Preserve local revisions; never auto-resolve same-field conflict. |
| Acknowledgement | Retained mutation owner; complete correlated receipt | Publish committed row/version before releasing dependent work. |
| Projection refresh | Query owner and WorkbookSurfaceRegistry | Version floors and generation fences; retain read debt after failure/detachment. |
| Retirement | Matching operation and authoring revision | Clear only owned work; stale completion cannot close/focus another editor. |

Parked work stays bound to its original semantic identity. Filtering, deletion,
merge, hiding, sorting, grouping and virtualization never retarget it. Eligibility
must be revalidated before resume. Closure and readable role loss prohibit writes;
session uncertainty conceals protected presentation while preserving same-account
work. Incident retirement and account replacement clear protected retained state.
No reload, crash, restart or cross-tab retention is promised.

## Gap register

Each row records remediation, affected areas, rationale/benefit, compatibility,
remaining risk and binary validation. Source observations are not browser passes.

| Gap | Observation / remediation and areas | Benefit / compatibility / risk | Binary validation |
| --- | --- | --- | --- |
| G1 | Managed enqueue returns Grid `accepted` before dispatch. Separate admission ticket from terminal outcome; Core 03 clarification, runtime, surface callers, adapter tests. | Correct navigation authority; internal TypeScript callers migrate together, HTTP unchanged; risk: pending keyboard focus. | Destination opens only for its accepted current revision; sheet detachment remains usable. |
| G2 | Generic raw text lives in mounted adapter state. Add retained workbook draft owner and neutral attachment callbacks; Core 03, models, surfaces, source guides, unit/browser tests. | One lifetime owner; no durable migration; risk: hidden/deleted targets and concealment. | Exact invalid/unsubmitted text returns only at original eligible target. |
| G3 | Managed settlement deletes overlays by field without revision identity. Track all contributed revisions through coalescing and settlement; driver/store/queue tests. | A cannot erase B or refused work; internal metadata changes only; risk: coalesced contributors. | Same/different-field B remains after A acknowledgement or discard. |
| G4 | Managed dispatch uses admission base; accepted rows are distributed through source-specific stores. Add guarded first capture and complete committed-row publication; Core 03, queue, runtime, query/source ports, API-backed tests. | Correct dependent bases and monotonic reads; wire contract unchanged; risk: external same-field changes and sparse events. | Follow-on uses accepted predecessor version; uncertain replay preserves original bytes/base. |
| G5 | Accepted work clears local overlays before mounted refresh; refresh debt already has a retained owner. Preserve complete receipt and separate acceptance/read obligations; runtime/query tests and browser evidence. | Failed reads cannot undo saves; retain existing debt registry; risk: stale/missing rows and late registrations. | Failed/stale/detached refresh preserves accepted row and B; recovery emits reads only. |
| G6 | Generic builder trims all input and maps blank clearable fields to null. Use declared scalar/reference/string rules with explicit clear intent; field adapter/model tests and browser matrix. | Exact raw authoring with owner-valid payloads; HTTP unchanged; risk: changing existing inspector consumers. | Invalid timestamps/references remain local; explicit permitted clear sends JSON null. |
| G7 | Adapter has local revision/navigation guards; cross-attachment late completion remains a characterization target. Fence semantic attachment and revision, retaining Timeline owners; adapter and browser tests. | Stable focus and uninterrupted input; no vendor change; risk: IME, virtualization and creation promotion. | Late A cannot close B, change its draft, or focus an obsolete destination. |

## Verification and evidence

Planning/execution inspection read the named surface/editor/driver/runtime paths,
pending queue, committed-record port, query owners, Timeline draft registry,
SurfaceRegistry, authored manifests and browser committed-response-loss fixtures.

Task-guide commands completed successfully for `package.grid_adapter`,
`web.workbook`, `module.workbook`, `module.entities`, `module.timeline`,
`module.artifacts`, `module.tasksdecisions`, `module.parties`, and `module.evidence`.
Historical product results are not fresh validation. Current product acceptance
is recorded in the integrated matrix and terminal results below.

## Completion, compatibility and rollback

Applicable acceptance rows require PASS with current evidence; N/A needs an owner
and scope rationale. Any applicable BLOCKED row prevents completion. Test data
must use isolated harness fixtures rather than existing analyst incidents.

No public HTTP route, payload, database schema, dependency or storage migration is
planned. Internal adapter/runtime callers migrate together. Rollback must restore
owner text, projections, implementation, generated derivatives and tests together;
it must not reverse accepted source edits, history or attribution.

Current disposition: GEA-01 through GEA-05 are DONE; stop at this seam.
Retained-run maintenance is skipped because RESULTS_DIR is unset.

## Authored committed-field inventory

All listed fields have `grid_editable=true` and `write_kind=direct_value`.

Hidden means eligible only when explicitly made visible. Timeline visible text
is source-owned free text, including its date/time-looking fields. Task status
guards are the coupled fields `task.status`, `task.owner_user_id`,
`task.blocked_reason` and `task.completed_at`; changing any of these validates
against all four. Other Task scalar fields do not acquire that dependency set.
Reference eligibility is revalidated by the existing source/reference owner;
retaining an identifier does not preserve its authorization or lifecycle.

| Surface / source verification | Field | Editor | Clear | Hidden | Dependencies |
| --- | --- | --- | --- | --- | --- |
| assessments / module.assessments | None | Owner-specific actions only | N/A | N/A | No ordinary grid mutation |
| comm_log / module.artifacts | `comm_log.timestamp_utc` | timestamp | No | No | Field contract and lifecycle |
| comm_log / module.artifacts | `comm_log.comm_type` | enum | No | No | Field contract and lifecycle |
| comm_log / module.artifacts | `comm_log.audience` | single-line | No | No | Field contract and lifecycle |
| comm_log / module.artifacts | `comm_log.channel_or_meeting` | single-line | No | No | Field contract and lifecycle |
| comm_log / module.artifacts | `comm_log.summary` | single-line | No | No | Field contract and lifecycle |
| comm_log / module.artifacts | `comm_log.next_report_at` | timestamp | Yes | No | Field contract and lifecycle |
| comm_log / module.artifacts | `comm_log.privilege_tag` | single-line | Yes | Yes | Field contract and lifecycle |
| decisions / module.tasksdecisions | `decision.summary` | single-line | No | No | Field contract and lifecycle |
| decisions / module.tasksdecisions | `decision.status` | enum | No | No | Field contract and lifecycle |
| decisions / module.tasksdecisions | `decision.owner_user_id` | stable-reference | No | No | Current reference eligibility |
| decisions / module.tasksdecisions | `decision.decision_type` | enum | No | No | Field contract and lifecycle |
| decisions / module.tasksdecisions | `decision.decided_at` | timestamp | No | No | Field contract and lifecycle |
| decisions / module.tasksdecisions | `decision.rationale` | multiline | No | No | Field contract and lifecycle |
| evidence / module.evidence | `evidence.title` | single-line | No | No | Field contract and lifecycle |
| evidence / module.evidence | `evidence.lifecycle_state` | enum | No | No | Field contract and lifecycle |
| evidence / module.evidence | `evidence.requested_at` | timestamp | Yes | No | Field contract and lifecycle |
| evidence / module.evidence | `evidence.received_at` | timestamp | Yes | No | Field contract and lifecycle |
| evidence / module.evidence | `evidence.storage_ref` | single-line | No | No | Field contract and lifecycle |
| evidence / module.evidence | `evidence.collector_party_text` | single-line | Yes | No | Field contract and lifecycle |
| evidence / module.evidence | `evidence.collector_party_id` | stable-reference | Yes | Yes | Current reference eligibility |
| evidence / module.evidence | `evidence.source_party_text` | single-line | Yes | No | Field contract and lifecycle |
| evidence / module.evidence | `evidence.source_party_id` | stable-reference | Yes | Yes | Current reference eligibility |
| findings / module.artifacts | `finding.statement` | multiline | No | No | Field contract and lifecycle |
| findings / module.artifacts | `finding.kind` | enum | No | No | Field contract and lifecycle |
| findings / module.artifacts | `finding.state` | enum | No | No | Field contract and lifecycle |
| findings / module.artifacts | `finding.owner_user_id` | stable-reference | No | No | Current reference eligibility |
| findings / module.artifacts | `finding.confidence_score` | number | Yes | No | Field contract and lifecycle |
| forensic_keywords / module.artifacts | `forensic_keyword.pattern` | single-line | No | No | Field contract and lifecycle |
| forensic_keywords / module.artifacts | `forensic_keyword.reason` | single-line | No | No | Field contract and lifecycle |
| forensic_keywords / module.artifacts | `forensic_keyword.match_mode` | enum | No | No | Field contract and lifecycle |
| forensic_keywords / module.artifacts | `forensic_keyword.case_sensitive` | boolean | No | No | Field contract and lifecycle |
| handoff / module.artifacts | `handoff.timestamp_utc` | timestamp | No | No | Field contract and lifecycle |
| handoff / module.artifacts | `handoff.outgoing_owner_user_id` | stable-reference | No | No | Current reference eligibility |
| handoff / module.artifacts | `handoff.incoming_owner_user_id` | stable-reference | No | No | Current reference eligibility |
| handoff / module.artifacts | `handoff.current_state_summary` | multiline | No | No | Field contract and lifecycle |
| handoff / module.artifacts | `handoff.next_checks` | multiline | Yes | No | Field contract and lifecycle |
| handoff / module.artifacts | `handoff.acknowledged_at` | timestamp | Yes | No | Field contract and lifecycle |
| hosts / module.entities | `host.display_name` | single-line | No | No | Field contract and lifecycle |
| hosts / module.entities | `host.hostname` | single-line | No | No | Field contract and lifecycle |
| hosts / module.entities | `host.location` | single-line | Yes | No | Field contract and lifecycle |
| hosts / module.entities | `host.os_platform` | single-line | Yes | No | Field contract and lifecycle |
| hosts / module.entities | `host.business_owner` | single-line | Yes | No | Field contract and lifecycle |
| hosts / module.entities | `host.criticality` | single-line | Yes | No | Field contract and lifecycle |
| hosts / module.entities | `host.containment_status` | single-line | Yes | No | Field contract and lifecycle |
| identities / module.entities | `identity.display_name` | single-line | No | No | Field contract and lifecycle |
| identities / module.entities | `identity.upn` | single-line | No | No | Field contract and lifecycle |
| identities / module.entities | `identity.email` | single-line | No | No | Field contract and lifecycle |
| identities / module.entities | `identity.sam_account_name` | single-line | No | No | Field contract and lifecycle |
| identities / module.entities | `identity.privilege_level` | single-line | Yes | No | Field contract and lifecycle |
| identities / module.entities | `identity.mfa_state` | single-line | Yes | No | Field contract and lifecycle |
| identities / module.entities | `identity.reset_status` | single-line | Yes | No | Field contract and lifecycle |
| indicators / module.indicators | None | Owner-specific actions only | N/A | N/A | No ordinary grid mutation |
| investigative_queries / module.artifacts | `investigative_query.platform` | single-line | No | No | Field contract and lifecycle |
| investigative_queries / module.artifacts | `investigative_query.purpose` | single-line | No | No | Field contract and lifecycle |
| investigative_queries / module.artifacts | `investigative_query.query_text` | multiline | No | No | Field contract and lifecycle |
| lesson / module.artifacts | `lesson.timestamp_utc` | timestamp | No | No | Field contract and lifecycle |
| lesson / module.artifacts | `lesson.summary` | single-line | No | No | Field contract and lifecycle |
| lesson / module.artifacts | `lesson.owner_user_id` | stable-reference | No | No | Current reference eligibility |
| lesson / module.artifacts | `lesson.closure_state` | enum | No | No | Field contract and lifecycle |
| notes / module.artifacts | `note.title` | single-line | No | No | Field contract and lifecycle |
| notes / module.artifacts | `note.body` | multiline | No | No | Field contract and lifecycle |
| parties / module.parties | `party.display_name` | single-line | No | No | Field contract and lifecycle |
| parties / module.parties | `party.party_kind` | enum | No | No | Field contract and lifecycle |
| parties / module.parties | `party.organization_name` | single-line | Yes | No | Field contract and lifecycle |
| parties / module.parties | `party.role_title` | single-line | Yes | No | Field contract and lifecycle |
| parties / module.parties | `party.primary_email` | single-line | Yes | No | Field contract and lifecycle |
| parties / module.parties | `party.timezone_name` | single-line | Yes | No | Field contract and lifecycle |
| parties / module.parties | `party.external_ref` | single-line | Yes | No | Field contract and lifecycle |
| parties / module.parties | `party.notes` | multiline | Yes | Yes | Field contract and lifecycle |
| status_review / module.artifacts | `status_review.timestamp_utc` | timestamp | No | No | Field contract and lifecycle |
| status_review / module.artifacts | `status_review.review_owner_user_id` | stable-reference | No | No | Current reference eligibility |
| status_review / module.artifacts | `status_review.current_state_summary` | multiline | No | No | Field contract and lifecycle |
| status_review / module.artifacts | `status_review.active_risks_summary` | multiline | Yes | No | Field contract and lifecycle |
| status_review / module.artifacts | `status_review.next_report_at` | timestamp | Yes | No | Field contract and lifecycle |
| task_requests / module.tasksdecisions | `task.title` | single-line | No | No | Field contract and lifecycle |
| task_requests / module.tasksdecisions | `task.status` | enum | No | No | Task status guards |
| task_requests / module.tasksdecisions | `task.owner_user_id` | stable-reference | No | No | Task status guards |
| task_requests / module.tasksdecisions | `task.priority` | enum | No | No | Field contract and lifecycle |
| task_requests / module.tasksdecisions | `task.task_kind` | enum | No | No | Field contract and lifecycle |
| task_requests / module.tasksdecisions | `task.workstream` | single-line | No | No | Field contract and lifecycle |
| task_requests / module.tasksdecisions | `task.due_at` | timestamp | Yes | No | Field contract and lifecycle |
| task_requests / module.tasksdecisions | `task.requester_party_text` | single-line | Yes | No | Field contract and lifecycle |
| task_requests / module.tasksdecisions | `task.requester_party_id` | stable-reference | Yes | Yes | Current reference eligibility |
| task_requests / module.tasksdecisions | `task.blocked_reason` | single-line | No | No | Task status guards |
| task_requests / module.tasksdecisions | `task.completed_at` | timestamp | Yes | No | Task status guards |
| task_requests / module.tasksdecisions | `task.external_ticket_ref` | single-line | No | No | Field contract and lifecycle |
| task_requests / module.tasksdecisions | `task.closure_summary` | multiline | No | Yes | Field contract and lifecycle |
| task_requests / module.tasksdecisions | `task.decision_record_id` | stable-reference | Yes | Yes | Current reference eligibility |
| timeline / module.timeline | `timeline.date_entered_text` | single-line | Yes | No | Field contract and lifecycle |
| timeline / module.timeline | `timeline.analyst_text` | single-line | Yes | No | Field contract and lifecycle |
| timeline / module.timeline | `timeline.mitre_stage_text` | single-line | Yes | No | Field contract and lifecycle |
| timeline / module.timeline | `timeline.device_object_text` | single-line | Yes | No | Field contract and lifecycle |
| timeline / module.timeline | `timeline.ip_address_text` | single-line | Yes | No | Field contract and lifecycle |
| timeline / module.timeline | `timeline.activity_utc_text` | single-line | Yes | No | Field contract and lifecycle |
| timeline / module.timeline | `timeline.activity_local_text` | single-line | Yes | No | Field contract and lifecycle |
| timeline / module.timeline | `timeline.raw_activity_text` | single-line | Yes | No | Field contract and lifecycle |
| timeline / module.timeline | `timeline.activity_synopsis_text` | single-line | Yes | No | Field contract and lifecycle |
| timeline / module.timeline | `timeline.data_source_text` | single-line | Yes | No | Field contract and lifecycle |

## GEA-01 exit evidence

- Core 03 REQ-03-298 and REQ-03-100 now explicitly separate retained authoring
  baseline, guarded first dispatch and immutable uncertain replay, using the
  user's approved decisions. REQ-03-099 already explicitly separates local
  admission from authoritative navigation acceptance. No unresolved owner
  contradiction remains for this seam.
- The matrix above covers all 97 direct-value fields in 15 eligible schemas and
  explicitly dispositions both non-editable specialized schemas. Read-only and
  action/collection fields acquire no new editor.
- `make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.committed_grid_autosave` reproduced all three
  intended regression failures: admission returned accepted; A removed B's
  visible edit; B dispatched base 1 after A accepted version 2. Run root:
  `.cartulary/test-results/20260915T004624Z-p35719`; diagnostic:
  `unit-logs/row-web.workbook.regression.committed_grid_autosave/vitest-failure-details.json`.
  These are related characterization failures, not product passes or an
  unrelated environment failure. Their repair belongs to GEA-03.
- Inspection confirms Generic/Entity do not connect raw editor writes to the
  retained runtime. The generic builder trims raw input and maps empty clearable
  timestamps/references to null. GEA-02 owns focused retention/validation tests
  alongside their replacements. Cross-attachment focus and stale/missing refresh
  remain targeted race hypotheses until the integrated scenarios run.
- Next action: start GEA-02, establish retained authoring and neutral attachment
  interfaces. No dependent implementation was started before this exit record.

## GEA-02 exit evidence

- Added `WorkbookGridDraftStore` and declared grid-value conversion; Generic and
  Entity controls now retain raw input in runtime ownership, restore it on
  explicit activation, preserve explicit null, and expose inline saved-change
  review. Inspector draft/submission paths remain independent.
- Grid Adapter has a neutral retain callback for both production and test
  bindings. Selection events preserve the semantic value, including explicit
  null, instead of replacing it with a control's empty display string.
- Timeline's existing neutral draft store now tracks revisions. Its mutation
  context captures source editor revisions. Query exclusion retires only mounted
  element references, preserving raw authoring. Source creation/mutation owners
  remain in place. Added context/revision settlement evidence.
- `make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.grid_draft_lifetime,web.workbook.regression.grid_field_intent`
  PASS, `.cartulary/test-results/20260915T005321Z-p44360`.
- `make test-slice OWNER=module.timeline
  ROWS=module.timeline.frontend.timeline_editor_draft_registry_0de7a147c1` PASS,
  `.cartulary/test-results/20260915T005602Z-p50475`.
- `make frontend-typecheck` PASS,
  `.cartulary/test-results/20260915T005602Z-p50554`. An earlier run
  `.cartulary/test-results/20260915T005321Z-p44441` found a missing required
  capture-state value in a new fixture; corrected before the passing rerun.
- `make format` PASS, `.cartulary/test-results/20260915T005541Z-p46109`.
  Initial format/slice admission rejected unsorted authored test titles; sorted
  those titles before execution. No assertion was weakened.
- GEA-01 documentation maintenance: `make lint-markdown` PASS,
  `.cartulary/test-results/20260915T004656Z-p36488`. Later documentation edits
  still require terminal validation.
- Next action: GEA-03 managed admission/settlement, guarded request capture and
  accepted-row reconciliation. Integrated browser/security acceptance is pending.

## GEA-03 exit evidence

- `WorkbookManagedPatchDriver` returns an admission ticket with an observable
  terminal result. Per-field contributors own overlay and raw-draft revisions;
  coalesced replacements settle as superseded. Acknowledgement and discard cannot
  remove newer, different-field or refused work. Generic/Entity callers await
  authoritative settlement and leave semantic navigation to Grid Adapter.
- `workbookPendingQueue` prepares and captures a base atomically before first
  dispatch. Replay skips preparation. Explicit fresh-ID recovery starts a new
  attempt. The existing FIFO, capacity, coalescing and reservation boundaries
  remain. Managed retryable failures now respect the existing scheduler delay.
- Complete correlated receipts are retained through the existing committed-record
  capability in `WorkbookExplicitPatchOwner`; runtime source coordination publishes
  accepted rows before releasing dependent work. Refresh remains owned by
  `WorkbookSurfaceRegistry` and is a separate read obligation.
- `createWorkbookPendingMutationAdapter` retains captured route/body/transaction
  identity through uncertain responses. Malformed acknowledgements are uncertain,
  and replay does not use a newly observed version. Captures retire with the
  account/incident runtime. A receipt must advance the captured request base.
- Timeline retains original scalar baselines and editor-context revisions. Its
  materialization no longer includes unrelated inspector/grid authoring. Own
  accepted predecessors advance only their fields; relevant independent changes
  require inline review. Coalescing merges captured contributors. Uncertain
  replay uses the captured base. Conflict registration no longer replaces newer
  raw text; conflict retirement carries exact source draft revisions.
- These additional confirmed findings extend G2/G3/G4/G5/G7: query omission
  retired Timeline raw text; inspector materialization mixed grid text; malformed
  success was treated as rejection; retry bypassed its delay; conflict callbacks
  restored/cleared raw values without revision ownership. Remediation areas are
  the neutral draft store, Timeline registry/driver/conflict bindings, common
  pending adapter/runtime and owning tests/source guides. They improve predictable
  continuation without wire/schema migration. Integrated source/security/focus
  behavior remains the GEA-04 risk. Binary validation is exact retained text and
  original-target focus, immutable replay, one durable effect and monotonic rows.
- Grid Adapter deduplicates commits by revision and value, retains seeded direct
  input, and fences detached completion. The live RDG test now also replaces an
  editor at the same semantic cell with identical text before its old save settles.
  Only the original operation settles; the replacement retains text and focus.
- `make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.committed_grid_autosave,web.workbook.regression.workbookz_mutation_pending_adapter_0f28ab91e4`
  PASS, `.cartulary/test-results/20260915T012557Z-p2350`.
- `make test-slice OWNER=package.grid_adapter
  ROWS=package.grid_adapter.regression.index_suite_d804ef9789` PASS,
  `.cartulary/test-results/20260915T012557Z-p2301`.
- `make test-slice OWNER=module.timeline
  ROWS=module.timeline.frontend.timeline_editor_draft_registry_0de7a147c1` PASS,
  `.cartulary/test-results/20260915T012557Z-p2325`.
- `make test-slice OWNER=module.workbook
  ROWS=module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_8c99e88779`
  PASS, `.cartulary/test-results/20260915T012137Z-p89458`.
- `make test-slice OWNER=module.evidence
  ROWS=module.evidence.frontend_unit.timeline_file_draft_recovery` PASS,
  `.cartulary/test-results/20260915T012137Z-p89540`.
- `make frontend-typecheck` PASS,
  `.cartulary/test-results/20260915T012557Z-p2416`.
- Common coalescing, surface/account continuity and Timeline driver planning
  rows passed in `.cartulary/test-results/20260915T011620Z-p77178`.
  Draft validation-state and transport tests also passed in
  `.cartulary/test-results/20260915T012137Z-p89434`.
- Related intermediate failures: missing fixture/TypeScript declarations were
  corrected; an outdated admission assertion was migrated; replay timing exposed
  and verified the scheduler fix. Diagnostic runs:
  `.cartulary/test-results/20260915T010458Z-p57844`,
  `.cartulary/test-results/20260915T010736Z-p63909`, and
  `.cartulary/test-results/20260915T011450Z-p75502`.
  A generated-in-memory test-title update initially included duplicate entries;
  authored selectors were corrected before Make execution. No product assertion
  or tolerance was weakened.
- Internal TypeScript interfaces migrate with their consumers. No new queue,
  inspector submission workflow, persistence, vendor, dependency or migration.
  Obsolete unconditional field clearing and admission-as-acceptance callbacks
  have been replaced. Next action: GEA-04 integrated evidence.

## GEA-04 integrated evidence (in progress)

The contract-driven scalar test enumerates all 97 editable fields and all seven
families, and explicitly rejects non-editable fields: PASS,
`.cartulary/test-results/20260915T013318Z-p11601`.

The first API-backed run used `make service-backed-test-slice OWNER=module.workbook`
with `ROWS=module.workbook.browser.grid_autosave_successive,module.workbook.browser.grid_autosave_exact_replay,module.workbook.browser.grid_autosave_timestamp_retention,module.workbook.browser.grid_autosave_surface_matrix,module.workbook.browser.grid_autosave_editor_families`.
Run `.cartulary/test-results/20260915T013613Z-p23258`: two scenarios PASS, three FAIL.

- Successive A/B authoring, real committed receipts, request bases 1/2, distinct
  transaction/change-set identities, history count 3 and final next-cell focus
  passed on Hosts, Evidence and Timeline.
- Ordinary direct editing passed on all 15 surfaces, with real request bodies,
  durable saved values and history increments.
- Sheet detachment triggered blur submission of newer B: confirmed G1/G7 gap.
  Explicit sheet, system-view and saved-view controls now declare the existing
  semantic external-action boundary. Ordinary navigation retains its acceptance
  gate. Benefit: detachment preserves unsubmitted text without adding a queue.
  Compatibility: private DOM boundary only. Binary validation: two identical
  replay bodies and one durable edit, with B restored locally.
- Timestamp Clear was obscured by the neighboring cell: confirmed G6/G7 gap.
  Source correction controls use a neutral adapter geometry slot above adjacent
  cells, within the original column width. No row-height change. Binary
  validation: pointer and keyboard can reach correction without forced events.
- Boolean activation called a text-only caret API: confirmed G7 gap. Guard caret
  placement by actual text-selection support. Binary validation: boolean editor
  activation, checkbox intent and authoritative navigation pass without errors.

A focused rerun of exact replay, timestamps and families at
`.cartulary/test-results/20260915T014432Z-p61632` passed replay (including malformed
acknowledgement) and all additional editor families. Timestamp clear exposed a
second draft-lifetime defect: the attachment effect re-retained completed text on
subsequent renders, retaining an obsolete baseline. Initial seed retention is now
limited to attachment. This extends G2/G3/G7; affected areas are adapter lifecycle,
inline recovery, browser timestamp evidence and this handoff. Binary validation:
clear after an acknowledged timestamp dispatches null and retires only that work.
Risk remains until the focused browser rerun passes.

Unavailable original targets now expose a compact, copyable local-text disclosure
with discard, backed directly by the generic draft store or Timeline registry.
This implements the approved closure/role-loss/hidden/deleted-target disposition;
it grants no submission authority. Source ownership and the components guide were
updated together. Security and presentation tests below must establish concealment,
original-target identity and explicit reactivation before this workstream exits.

Intermediate checks: `make frontend-typecheck` PASS at
`.cartulary/test-results/20260915T014432Z-p61703`; Grid Adapter focused suite PASS at
`20260915T014432Z-p61608`. `make format` initially failed on an obsolete hook
dependency; after correction PASS at `20260915T014822Z-p95884`.
`make lint-biome` at `20260915T014602Z-p94585` identified formatting and four new
non-null assertions; corrections are pending fresh static verification.

Further GEA-04 evidence and corrections:

- Timestamp lexical input and explicit null clear PASS at
  `20260915T015318Z-p11356`. Managed autosave unit tests (including denial recovery)
  PASS at `20260915T020850Z-p7995`; current 12-test suite with contiguous/sparse
  collaboration cache evidence PASS at `20260915T022658Z-p54690`.
- Session suspension/recovery, in-place account replacement, real viewer denial
  followed by incident revocation, and inspector/grid same-record sequencing
  PASS at `20260915T020907Z-p9440`. The new composition/keyboard correction,
  focus visibility, narrow viewport and 200% zoom accessibility scenario also
  PASS in that run. Initial session failure was a fixture API authentication
  mismatch after UI login; the authoritative isolated administrator read fixed
  the assertion without changing the application session.
- G1/G7: inspector focus was stolen by an earlier Enter completion. Adapter
  navigation generation now advances when focus leaves or another semantic
  transition starts. Source callbacks settle independently of presentation.
  Package tests PASS at `20260915T022220Z-p15687`; typecheck PASS at
  `20260915T022221Z-p16072` and `20260915T022844Z-p55654`.
- G2/G7: incident-closed notifications did not refresh the current incident
  resource, and the vendor retained internal edit mode through read-only state.
  The identity hook uses its existing version-fenced read; Grid Adapter detaches
  vendor presentation without source discard or focus movement. Closure/reopen
  across Hosts, Evidence and Timeline PASS at `20260915T022538Z-p18715`.
  Compatibility: existing lifecycle routes and authority remain controlling.
  Binary validation: zero PATCH, copyable exact closed draft, no automatic editor
  on reopen, explicit original-cell activation restores exact text.
- G4: detached/inactive-view collaboration did not advance the shared committed
  cache on every surface. Validated contiguous patches now update cached complete
  rows before active-surface projection; sparse gaps retain only a version floor.
  They do not acknowledge a request. Real Host different-field preparation and
  same-field review PASS at `20260915T022538Z-p18715`: bases 1/3/5, durable versions
  through 6, unique client transactions, remote fields preserved. Areas:
  committed owner, collaboration coordinator, unit/browser coverage, source guides
  and this record. Benefit: dependent writes use current evidence independent of
  presentation. Risk: sparse gaps require a full read or local review; replay
  remains captured. No public protocol or storage migration.
- G5: failed, stale and missing post-acceptance queries PASS at
  `20260915T022137Z-p83755` on all three representative surfaces. Detached recovery
  performs reads; accepted values and newer exact raw text survive. Timeline's
  default unsorted scalar patch does not require a query; the browser fixture now
  sorts the edited field to exercise the actual refresh obligation. An offscreen
  saved-value assertion now uses the existing virtualization helper.
- Fresh inspector recovery (three scenarios), ordinary creation, and Entity
  paste replay (two scenarios) PASS in component browser-group artifacts under
  `20260915T022739Z-p63007`. The enclosing command failed on the new CSRF scenario;
  its error-context artifact also triggered the harness secret-syntax guard.
  The fixture now removes the proof header using its named support constant,
  rather than embedding a synthetic header value in captured test source.
  CSRF continuity remains pending; no aggregate pass is claimed.
- Timeline regression run `20260915T022738Z-p62782`: 10/12 PASS, including pointer,
  keyboard, refresh, virtualization, uncertain creation, paste/fill and grouped
  conflict recovery. Rejection/discard and creation duplicate-Tab remain under
  correction. The former exposed recovery-button blur resubmission; the latter
  exposed first preparation replacing source admission signatures with normalized
  transport signatures. Recovery actions now declare semantic external action;
  preparation preserves admission identity separately from captured request base.
- Saved-view availability fixture corrections use the complete authored layout
  schema. Base presentation retains working column visibility, so resumption
  explicitly reveals the original hidden field before activation. This is a
  fixture correction, not a universal navigation or visibility policy change.
- `make generate` PASS at `20260915T022712Z-p59510`. `make lint-biome` at
  `20260915T022845Z-p55855` found only the newly edited browser file's formatting;
  formatting and fresh static verification remain in progress.

Next action remains GEA-04: resolve CSRF/availability and the two Timeline
regressions, complete fresh Evidence/History/Network Analysis and accessibility
coverage, then mark this workstream DONE before terminal GEA-05 verification.

GEA-04 follow-up evidence:

- Full Timeline creation continuation/duplicate Tab PASS at
  `20260915T023348Z-p66254`; rejection and conflict-discard continuity PASS at
  `20260915T023710Z-p96133`. Preparation must preserve the source admission
  signature because normalized transport intent omits the view discriminator.
  The owning queue test now asserts duplicate admission after prepared capture:
  PASS at `20260915T023820Z-p92980`.
- Conflict settlement first lets mounted source presentation test its captured
  revision, then retires detached remainder. Conflict token replacement preserves
  those revision captures. This fixes G3/G7 without clearing later authoring.
  Full Collaboration resolver sequencing PASS at `20260915T024152Z-p58910`;
  its accessibility scenario PASS at `20260915T023821Z-p93201`.
- Evidence existing-upload acknowledgement-loss/remount PASS at
  `20260915T023526Z-p32640`. Four selected History scenarios PASS at
  `20260915T023527Z-p32888`: retained Timeline replay against a newer row,
  detached Entity and Generic delete recovery, and attached-Evidence rollback
  preserving later edits.
- Network Analysis exact temporal navigation, selection clearing announcements
  and claimed keyboard/focus/read-only accessibility PASS at
  `20260915T023711Z-p96370`. Shared adapter changes grant no editing capability.
- Saved-view hiding/filtering and deleted original-target retention/discard PASS
  at `20260915T023709Z-p95866`. The fixture returns to an explicit unfiltered saved
  view; selecting Unsaved view intentionally keeps the current working query.
- CSRF denial now recovers through the existing authorization scheduler. A
  confirmed authority can precede React attachment or be superseded by a newer
  query; cancelled confirmation reads must retry while replay remains paused.
  Cancellation of authorization itself still pauses. Areas: coordinator, existing
  confirmation classification tests, real browser scenario, source guide.
  Binary validation: identical captured request, one history increment, preserved
  newer raw text, no incident revocation. PASS alongside refresh-only recovery at
  `20260915T024151Z-p58520`. Initial `20260915T023914Z-p27677` was an infrastructure
  source-snapshot failure caused by editing during the build; later browser
  builds use frozen implementation input until completion.
- Reviewed newer baselines are not regressed by older predecessor receipts.
  Generic drafts/contributors keep their newer baseline version; Timeline
  advances only baseline field values matching that predecessor's authored base.
  Registry PASS at `20260915T024349Z-p24119`; managed test PASS at
  `20260915T024150Z-p58270`. An initial test omitted the explicit review step and
  correctly halted the original baseline; the fixture now performs review.
- Coordinator unit suite PASS at `20260915T024347Z-p22904`, including unchanged
  authority recovery, cancellation distinctions and revocation. A remaining
  old admission-as-acceptance assertion was migrated to the admission ticket.
  Typecheck PASS at `20260915T024350Z-p25131`.
- Timeline refresh debt now arises from required query reconciliation or detached
  application, not every successfully applied unsorted scalar receipt. Required
  refresh uses `WorkbookSurfaceRegistry`; ordinary saves no longer leave a
  spurious refresh notice. The notice occupies the existing top-bar recovery
  slot rather than introducing an extra shell grid row.

Current next action: finish the consolidated 13 functional grid scenarios and
shared accessibility checks, then run focused Timeline entry measurements.

Consolidated GEA-04 checks:

- Shared runtime/queue/transport/draft suites PASS at
  `20260915T024433Z-p29111`; Grid Adapter and semantic focus suites PASS at
  `20260915T024434Z-p32281`; Grid Adapter accessibility PASS at
  `20260915T024432Z-p28851`; Biome PASS at `20260915T024542Z-p96107`.
- The 13-scenario functional consolidation at `20260915T024431Z-p28626` passed
  12 scenarios, plus both Workbook accessibility rows. A's socket echo could
  arrive before its HTTP receipt, causing premature local rejection of B.
  G1/G4 correction: permit local admission behind existing grid work, then perform
  field/dependency review at first preparation after predecessor settlement.
  The original baseline and uncertain request remain unchanged. Pending feedback
  suppresses premature review of the predecessor echo. A/B, collaboration review
  and timestamp rerun PASS at `20260915T024820Z-p3928`.
- Timeline creation, keyboard and rejection/conflict continuity PASS together at
  `20260915T024821Z-p4147`. Current typecheck PASS at
  `20260915T024748Z-p97828`; managed autosave suite PASS at
  `20260915T024750Z-p99010`.
- Timeline first preparation now contributes its complete known row to the common
  committed cache and respects that cache's version floor. This lets contiguous
  collaboration patches observed during a captured attempt inform the next
  never-dispatched write. The dedicated real API scenario is pending below.

- Dedicated Timeline collaboration scenario PASS at
  `20260915T025222Z-p75606`: real A commit at version 2, collaborator change at
  version 3 before A HTTP acknowledgement, then B prepares version 3 and commits
  version 4. Both B and the collaborator field survive with four History entries.
- Evidence file-draft recovery unit row PASS at `20260915T025223Z-p75824`;
  Timeline driver plans and runtime surface continuity PASS at
  `20260915T025225Z-p76062`. Current generation PASS at
  `20260915T025123Z-p67795`, formatting PASS at `20260915T025122Z-p67649`, and
  typecheck PASS at `20260915T025124Z-p69471`.

## Integrated acceptance matrix

Run identifiers below resolve under `.cartulary/test-results/`; each retained
`run-summary.json`, target summary and browser group result supplies actual row
accounting. Machine coverage remains in authored test-family JSON and executable
contract-driven tests, independent of this document.

| Acceptance boundary | Result | Current owning evidence |
| --- | --- | --- |
| All 97 declared fields, seven editor families, exact clear/lexical rules | PASS | `grid_field_intent` unit matrix; `grid_autosave_surface_matrix`, `editor_families`, and `timestamp_retention` browser rows; p28626 and p3928 runs above. |
| Raw lifetime, original target, no automatic editor/focus on remount | PASS | Draft store/registry and adapter tests; exact replay, availability, closure and session browser rows. |
| A/B same-field settlement and authoritative navigation | PASS | `grid_autosave_successive` on Hosts/Evidence/Timeline, `20260915T024820Z-p3928`; owning queue/driver tests. |
| Different-field follow-on, collaboration before preparation/during dispatch | PASS | Host collaboration `20260915T024820Z-p3928`; Timeline collaboration `20260915T025222Z-p75606`; field/dependency unit guards. |
| Pre-dispatch coalescing, FIFO capacity 64, refused/newer text and discard | PASS | Queue capture/signature suite, managed autosave suite, adapter stale-completion tests; Timeline duplicate Tab and rejection browser rows. |
| Captured replay after real commit/response loss | PASS | `grid_autosave_exact_replay`; body/base/transaction equality, History and effect counts at `20260915T024431Z-p28626`; Timeline/Evidence/History recovery regressions. |
| Accepted row floor, failed/stale/missing/detached refresh, reads-only recovery | PASS | `grid_autosave_refresh` at `20260915T024151Z-p58520`; common cache/receipt and refresh unit suites. |
| Inspector and specialized owner coordination | PASS | Grid/inspector browser sequencing, three inspector recovery scenarios, Evidence file recovery, History rollback and runtime merge-admission tests. |
| Selection, sorting/grouping, paste/fill, virtualization and scroll | PASS | Existing Timeline browser rows at `20260915T022738Z-p62782`; creation/rejection/keyboard rerun `20260915T024821Z-p4147`; availability row. |
| Hidden/filter-excluded/deleted original targets | PASS | `grid_autosave_availability`; exact non-actionable parked text, explicit original-cell return, no retargeting or PATCH. |
| Suspension/recovery, replacement, closure/reopen, role loss/revocation | PASS | Real API-backed session, closure and role rows at `20260915T024431Z-p28626`; authority/epoch/store unit tests. |
| Operation denial, CSRF and stale callbacks | PASS | Real CSRF recovery at `20260915T024151Z-p58520`; adapter navigation/attachment/revision fences and coordinator recovery suite. |
| Creation, inspector, bulk, Evidence, History and conflict regressions | PASS | Fresh owning browser groups and corrections recorded above; no historical pass is substituted. |
| Keyboard, IME, visible focus, announcements, compact/narrow/200% zoom | PASS | New grid autosave accessibility, existing Workbook and Grid Adapter accessibility; Collaboration and Network Analysis accessibility runs above. |
| Network Analysis read-only/navigation | PASS | `20260915T023711Z-p96370`; no new grid-edit capabilities. |
| Timeline entry latency | PASS | All four current owner measurements, `20260915T025618Z-p9331` (20/20 execution units). |
| Assessment/Indicator existing-row editors; collections/action payload editors | N/A | No declared ordinary direct-value grid editors in scope; retained specialized owners are unchanged. |
| Reload/crash/cross-tab recovery | N/A | Explicitly excluded; memory lifetime ends with the existing runtime/security boundary. |

Coverage is representative in the real browser (all 15 surfaces and all seven
families), with exhaustive authored-field eligibility/intent coverage in unit
tests. It does not claim a browser cross-product of every field, navigation and
race. Service fixtures are isolated synthetic incidents; existing analyst data
was not used or modified.

## GEA-04 exit

`make service-backed-test-slice OWNER=module.timeline` with the four authored
measurement rows (`committed_timeline_summary_typing_acknowledgment_b615aabfe6`,
`timeline_blank_row_creation_satisfies_the_paint_afddd2ce13`,
`timeline_summary_arrow_down_selection_satisfies_961a4ec1d3`,
`timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95`) PASS at
`.cartulary/test-results/20260915T025618Z-p9331`: 20/20 execution units, all four
paint-qualified measurements within their existing budgets. No assertion,
threshold, tolerance, fixture workload or renderer pin was weakened.

GEA-04 is DONE. Confirmed gaps have current owning-boundary evidence; the
integrated matrix has no unresolved applicable blocker. Next action: GEA-05,
starting with `make agent-finalize` before broader terminal verification.

## GEA-05 execution

Task guides were inspected for the requested five owners and source owners from
the matrix; additional risk routing covers authorization, collaboration,
Revisions, Network Analysis and application composition. `RESULTS_DIR` is unset:
no successful full-warm run is being claimed, so retained-run maintenance is
skipped. Broader checks begin only after the finalizer completes.

- `make agent-finalize` PASS at `20260915T030103Z-p51203` (1/1 execution
  unit; generated refresh, schema/catalog and tier coverage completed). Retained
  canonical-evidence and scheduler maintenance were skipped with `RESULTS_DIR`
  unset. No full-warm retained-run evidence was supplied.
- Terminal verification now includes the full frontend unit projection, typecheck,
  import boundaries, Biome and advisory Fallow; focused current visual rows; and
  real service rows for every source family in the editable surface matrix.

Terminal `make frontend-unit` at `20260915T030132Z-p55784` completed with
613/623 execution units passing and ten failed routed units. Causes and actions:

- Three old enqueue assertions now expect an admission ticket; terminal effects
  remain asserted separately by their owning tests.
- The new browser file now imports every view identity from the public contract
  facade. Architecture policy remains unchanged.
- The isolated Timeline fixture now supplies current authority to the shared
  retained owner, matching application assembly; this restores authorized query
  refresh in its existing sorted/filtered test.
- Conflict fixture now includes the required token, record, resolution class and
  attribution. A malformed conflict is uncertainty, not valid conflict evidence.
- Grid editor tests deliver the native input event after changing DOM input.
  Unannounced synthetic DOM writes do not update semantic raw authoring. Current
  editor value still differs from committed row state when Enter/Tab dispatches.
- The older-response fixture now asserts retained raw input and Syncing alongside
  the monotonic committed row version. Its response predates the new request's
  base and cannot legitimately close the editor as accepted.
- Confirmed regression: Observation readiness read only the default grid draft
  after source-specific materialization was introduced. Its source owner now
  checks independent inspector drafts too; the test exercises both surfaces.
  Areas: Observation hook, owning unit test, Timeline source guide and this
  record. Benefit: specialized actions cannot consume uncommitted source text.
  Compatibility: existing action authority and source selection remain unchanged;
  no public or storage migration. Binary validation: either local source draft
  blocks preparation, clearing it restores readiness, changed saved source still
  rejects. Pending fresh validation; no applicable failure is waived.

Focused reruns resolved all ten terminal unit failures: web.workbook selection
PASS at `20260915T030833Z-p15205` (7/7); Timeline selection PASS at
`20260915T030833Z-p15230` (3/3); Collaboration selection PASS at
`20260915T030844Z-p40120` (2/2); harness architecture policy PASS at
`20260915T030853Z-p63191` (2/2). The full frontend suite is rerunning after the
fixture and Observation-owner corrections.

Visual comparisons PASS without golden changes: Workbook shell and inspector
rows at `20260915T030132Z-p55683`; Timeline active/syncing/saved/conflict and
grouped-grid rows at `20260915T030132Z-p55705`; Grid Adapter support specimens at
`20260915T030833Z-p15259`. Renderer, fixture scope, masks and tolerance are unchanged.

Fallow target PASS at `20260915T030132Z-p55828`; its package-surface blocking
check has zero findings. The repository-wide advisory report retains 398 findings
and a nonblocking health exit 1 in `frontend-fallow-static/fallow-static-summary.json`.
Changed-path review found source calls for the flagged runtime methods; retained
receipt access is an intentional capability, covered by receipt tests. No broad
cleanup or suppression was added, and advisory findings are not claimed resolved.

Terminal typecheck at `20260915T030915Z-p66018` caught a missing required
`sessionIdentity` in the revised isolated Timeline authority fixture. The fixture
now uses its existing `fixture-session` identity, matching its batch authority.
This is a test composition correction; production source is unchanged.

### Final product and service results

Full `make frontend-unit` PASS at `20260915T030915Z-p65977`: 623/623 execution
units. Final typecheck, import boundaries and Biome PASS respectively at
`20260915T031221Z-p97104`, `20260915T031221Z-p97129` and
`20260915T031221Z-p97156`. The missing fixture session identity is corrected.

All 14 new functional grid-autosave scenarios and the new accessibility scenario
PASS together at `20260915T031036Z-p86921` (13/13 execution units, 15 catalog
rows across functional and accessibility groups). This includes complete surface
and editor-family matrices, captured response loss, same/different-field races,
collaboration, refresh-only remount recovery, inspector sequencing, exact parked
text, hidden/filter/deleted targets, closure, session/account/role/incident
boundaries, CSRF, keyboard/IME, narrow layout and 200% zoom.

Existing inspector, batch and ordinary-creation accessibility PASS at
`20260915T030915Z-p65855` (13/13); Collaboration visual rows PASS at
`20260915T030932Z-p83125` (11/11). No golden changed. The source-readiness
correction also has fresh real-browser Observation paging/source-edit regression
PASS at `20260915T031128Z-p33203` (11/11).

The following commands ran from the repository root. Exact selected rows are
recorded here and in each run's `run-manifest.json`; statuses come from the actual
run summaries, not task-guide routing.

| Make target / owner / selected rows | Result | Run |
| --- | --- | --- |
| `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.support_integration.integration_surface_envelope_1366b4ce6c` | PASS | `20260915T030158Z-p97651` |
| `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.store.attached_evidence_create_and_patch_aa33ea0168` | PASS | `20260915T030200Z-p98913` |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.integration.real_http_create_patch_review_supersede_replay_m_37e5c78ca1,module.timeline.integration.same_field_timeline_patch_conflicts_are_transpor_808ccc2585,module.timeline.store.patch_replay_returns_the_original_committed_resu_b93ddcbc77,module.timeline.store.stale_timeline_patches_derive_committed_writable_b89a3c27fa` | PASS | `20260915T030334Z-p83616` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.integration.full_row_and_sparse_patch_wire_families_preserve_32a4ab3be1` | PASS | `20260915T030500Z-p26757` |
| `make service-backed-test-slice OWNER=module.artifacts ROWS=module.artifacts.source_mutations.artifact_source_create_and_patch_preserve_all_ei_c8d63cf5cd` | PASS | `20260915T030504Z-p28347` |
| `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.idempotency_adapter.application_workbook_idempotency_adapter_boundar_90a3399bf6` | PASS | `20260915T030611Z-p75411` |
| `make service-backed-test-slice OWNER=module.parties ROWS=module.parties.idempotency.party_replay_storage_preserves_semantic_results_478b92a552` | PASS | `20260915T030648Z-p92689` |
| `make service-backed-test-slice OWNER=module.auth ROWS=module.auth.support_integration.user_list_continuation_uses_live_rows_7ed03d0063` | PASS | `20260915T030833Z-p15285` |
| `make service-backed-test-slice OWNER=module.revisions ROWS=module.revisions.store.http_security_precedence_4e7a9c2d13,module.revisions.integration.retained_history_evidence_for_an_extant_record_r_b590f9c4b1` | PASS | `20260915T030948Z-p2602` |
| `make service-backed-test-slice OWNER=module.collaboration ROWS=module.collaboration.integration.resume_with_a_valid_replay_token_replays_replaya_1826b38095` | PASS | `20260915T031133Z-p36045` |
| `make service-backed-test-slice OWNER=app.server ROWS=app.server.process.process_smoke_exercises_login_csrf_session_revoc_c6fd6eb90f` | PASS | `20260915T030939Z-p90180` |
| `make test-slice OWNER=module.workbook ROWS=module.workbook.unit.registry_derived_writable_conflict_capabilities_c0614db52a,module.workbook.unit.timestamp_instant_v1_accepts_rfc_3339_json_strin_ba11c9c407` | PASS | `20260915T030915Z-p65901` |
| `make test-slice OWNER=module.entities ROWS=module.entities.unit.entity_mutation_admission_and_hashing_remain_own_80f4147eb7` | PASS | `20260915T030926Z-p76255` |
| `make test-slice OWNER=module.auth ROWS=module.auth.unit.an_inventory_driven_loop_covers_every_cookie_aut_0b6e53ff68` | PASS | `20260915T030938Z-p89197` |
| `make test-slice OWNER=module.networkflow ROWS=module.networkflow.frontend_integration.verify_production_network_flow_grids_read_only_b_f662be335e` | PASS | `20260915T030955Z-p13557` |

### Terminal artifact checks

- `make generate-drift` PASS at `20260915T031221Z-p96768` (4/4).
- `make generated-artifact-policy-check` PASS at `20260915T031221Z-p96779` (3/3).
- `make json-shape-check` PASS at `20260915T031221Z-p96782` (3/3).
- `make lint-markdown` PASS at `20260915T031221Z-p97185`; artifact:
  `adhoc/lint-markdown/tool-run-summary.json`. The completed final record below
  receives a final formatting pass before GEA-05 closes.
- `git diff --check` PASS. Final branch/HEAD remain `main` /
  `29be6f83f4a6f05d2bd848d47ff4399328d92ac1`; working tree contains this seam's
  uncommitted changes. Baseline had no pre-existing edits.

### Final substantive change and ownership review

The shared workbook runtime owns exact scalar authoring and revision captures.
Generic/Entity editors attach to it; Timeline's retained source registry uses the
same revision-safe semantic boundary while keeping creation, mutation, collection,
paste/fill and Evidence ownership. The neutral local store carries opaque keys,
raw values, revisions and baseline facts without feature policy.

Managed autosave returns a stable admission ticket and settles only contributed
revisions. Never-dispatched work prepares against validated current committed
facts and relevant-field/dependency guards; an uncertain operation reuses its
captured route, payload, base and transaction. Complete receipts and monotonic
rows are published before dependent release. Refresh debt stays in
`WorkbookSurfaceRegistry`, with compact read-only recovery after detachment.

Grid Adapter owns editor attachment, composition/caret, vendor translation,
duplicate-commit prevention and semantic focus/navigation generations. It cannot
acknowledge admission or retire another operation's draft. Ordinary editor
navigation keeps the Core 03 acceptance gate; explicit sheet/saved-view changes
can park work. Closure/role loss preserves readable text; suspension conceals it;
account/incident retirement clears the retained runtime.

Inspected/changed areas include both workbook surfaces, scalar control and parked
feedback, runtime/queue/transport/committed-record coordination, Timeline retained
editing and mutation bindings, collaboration/current-incident attachment, shared
Grid Adapter, source guides, Core 03 clarifications, authored source/test manifests
and Make-generated browser/topology derivatives. New executable tests cover the
97-field contract inventory and retained/settlement boundaries; new browser
scenarios use real isolated API commits. No executable input reads Markdown.

Retired behavior: mount-only generic draft ownership; blanket grid trimming and
blank-to-null conversion; admission-as-acceptance; revision-blind overlay and
editor clearing; dispatch-time rewriting of uncertain request bases; unguarded
late navigation; and mounted-only required Timeline refresh. No hidden grids are
kept alive for state retention. No retained feature owner was replaced by an
inspector confirmation flow or parallel autosave queue.

Internal TypeScript callers and tests migrate together. Public HTTP routes,
payload schemas, field declarations, database schema and dependencies are
unchanged. The approved Core 03 clarification is projected into the local queue,
draft, receipt and adapter interfaces; no new wire/projection field was needed.
Generated changes are limited to derivatives of authored test routing. No digest,
lockfile, vendor, golden, migration, backend source or analyst dataset changed.
No commit, push, deployment or external publication occurred.

### Terminal acceptance and limits

| Acceptance | Result / disposition |
| --- | --- |
| Contract reconciliation, all editor families and transitions | PASS; GEA-01 decisions and full declared-field inventory. |
| Exact retained authoring and semantic/security attribution | PASS; integrated store, adapter and API/browser evidence. |
| Successive settlement, guarded preparation and immutable replay | PASS; unit races plus real commit/response-loss and History counts. |
| Accepted row reconciliation and retained refresh-only recovery | PASS; stale/missing/failed/detached scenarios. |
| Spreadsheet and retained workflow regressions | PASS; fresh Timeline, inspector, creation, bulk, Evidence, History, conflict, Observation and Network Analysis checks. |
| Accessibility, density, narrow layout, zoom and measurements | PASS; current owning rows and four existing Timeline measurement budgets. |
| Visual regressions | PASS; six selected owner rows covering shell, inspector, Timeline states/grouping, adapter and collaboration; no golden update. |
| Source/import/test ownership, typecheck, full frontend unit suite | PASS; current manifests, boundaries and 623/623 units. |
| Source/security/revision/application service contracts | PASS; selected actual commands and run summaries above. |
| Generation, policy, JSON shape and whitespace | PASS; terminal artifact checks above. |
| Final completed-handoff Markdown validation | PASS; `make lint-markdown` at `20260915T031736Z-p32066`, with `git diff --check` passing. |
| Retained full-warm run maintenance | SKIPPED; RESULTS_DIR unset, no qualifying full-warm run supplied. |
| Repository-wide advisory Fallow findings | Target PASS; existing nonblocking profile retains findings, blocking package-surface check has zero. No claim of repository-wide cleanup. |
| Full release/CI, unrelated backend suites and optional profiles | N/A; no backend/protocol/dependency/deployment change; focused source/service checks cover the changed frontend boundaries. |
| Golden update / two post-update ordinary runs | N/A; every selected existing comparison passed and no golden changed. |
| Browser cross-product of every field and every race | N/A; all fields receive contract-driven coverage, all surfaces/families receive representative browser evidence. |
| Reload, crash, cross-tab or persistent recovery | N/A; explicitly excluded; only the existing account/incident/client runtime lifetime is retained. |

No unresolved applicable product or verification blocker remains. The earlier
failures are documented with causes and fresh passing reruns; routing itself is
not offered as proof. Retention grants neither read nor write authorization.

### Rollback and next action

Rollback must restore the Core 03 clarification, internal TypeScript contracts,
implementation, source/test ownership inputs, generated browser/topology outputs,
source guides and tests as one coherent change. Regenerate derivatives through
Make and rerun the same affected checks. Do not perform data rollback: accepted
source edits, History, attribution and server idempotency records remain valid.
Resolve retained uncertain attempts through their captured identity before a
client lifecycle change; never manufacture a replacement transaction to erase an
accepted edit. Memory-local unsent drafts are not a reload migration mechanism.

GEA-05 is DONE after the required terminal checks and completed-handoff Markdown
and whitespace validation passed. All five workstreams are complete. Changes
remain uncommitted for review on the original branch. Stop here: no broader
refactor, publication or new editing capability follows from this handoff.
