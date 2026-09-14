# Ordinary workbook row creation and recovery handoff

## Baseline and authorized boundary

Execution baseline: clean `main` at
`83fd4165de4f8490235c5c5ad3b7564cedfe6e20`, revalidated before edits.
Only root `AGENTS.md` applies. Planning followed the localized digest read order,
Core 00 owner mapping, Core 01 mutation/discovery and Parties/coordination
addenda, Core 02 entity/artifact/Indicator/Evidence/history owners, Core 03
creation/continuity/security lifetime and spreadsheet interaction, Core 04
authorization/concealment, and the domain/design boundaries. Source placement
was inspected through the web source overview and child guides; verification
routing was inspected independently through authored catalogs.

Authored manifests identify React/React DOM 19.2.5, TypeScript 6.0.2, Vite 8.0.8,
Vitest 4.1.4 and Playwright 1.59.1. Grid Adapter encapsulates
react-data-grid 7.0.0-beta.59; workbook code adds no direct vendor-grid import.

The approved seam covers ordinary single-row creation on the fourteen schemas
below. One editable draft exists per canonical schema. Pending or uncertain
submission blocks only another ordinary create on that schema. Text changed
after dispatch becomes an unsent next draft after acceptance, with its own
explicit Commit. Timeline, Notes, Assessments and contextual creation retain
their specialized owners. No new routes, navigation surfaces, dependencies,
browser persistence, database migration, clipboard/import redesign, upload
redesign, or existing-record editing redesign is authorized. Do not edit the
digest, commit, push, deploy, or modify analyst data. Tests use isolated fixtures.

Planning evidence: three create-model rows passed through
`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.genericworkbookmodel_builds_generic_creates_with_07b632b33d,web.workbook.regression.genericworkbookmodel_seeds_owner_defaults_and_kn_82f50ea384,web.workbook.regression.genericworkbookmodel_uses_contract_create_minima_3376430d26`
at `.cartulary/test-results/20260913T202214Z-p26245`. This is not lifecycle or
browser evidence. Historical specialized-workflow handoffs are regression
navigation, not fresh passes.

## Sequential tracker

| Workstream | Dependency | State | Exit |
| --- | --- | --- | --- |
| ORC-01 characterization and reconciliation | none | DONE | Every entry/transition disposed; confirmed gaps reproduced. |
| ORC-02 retained authoring and readiness | ORC-01 | DONE | One authoritative draft and correct readiness for every entry. |
| ORC-03 attempts and migration | ORC-02 | DONE | Exact uncertain replay, retained acceptance, no superseded ordinary lifecycle. |
| ORC-04 integrated evidence | ORC-03 | DONE | All gaps and affected regressions pass applicable verification. |
| ORC-05 terminal validation | ORC-04 | DONE | Completed handoff bytes validated with all applicable acceptance PASS. |

Only the current row may be IN_PROGRESS. Save each exit, changed paths,
commands/results, risks, decisions and next action before starting its dependent.
An unresolved primary-owner contradiction or applicable blocked acceptance row
prevents advancement.

## Frozen entry and owner matrix

Schema suffixes below mean `cartulary.view.<suffix>.v1`. All fourteen have
authored schema projections and registered renderers. Runtime unavailable
discovery is distinct from a broken exposed entry. Creation is explicit Commit;
navigation alone creates nothing. Hosts/Identities use EntityWorkbookSurface;
the remaining entries use the generic renderer. Generic inspector creation is
supplementary to the grid: hidden draft fields and a shared Commit control.
Hosts/Identities have no standalone inspector-create path (N/A, no added path).

| Schema | Minimum | Defaults / special inputs | Behavioral owner | Verification owners |
| --- | --- | --- | --- | --- |
| hosts | display name, hostname, FQDN or AAD device ID | Entity-origin; aliases/defaults do not qualify; exact reuse | Core 01 REQ-01-323; Core 02 §§6,8 | web.workbook, module.entities, module.workbook |
| identities | display name, AAD object ID, SID, UPN, email or SAM account | Entity-origin; aliases/defaults do not qualify; exact reuse | Core 01 REQ-01-326; Core 02 §§6,8 | web.workbook, module.entities, module.workbook |
| parties | display name and kind | Exact email/external-reference reuse; no implicit text linking | Core 01 §19 REQ-01-501; Core 02 §8 | web.workbook, module.parties, module.workbook |
| task_requests | title and task kind | Open, normal priority, actor owner; no fabricated context | Core 01 REQ-01-336; Core 02 §10.4.1 | web.workbook, module.workbook |
| decisions | summary, type and rationale | Proposed, actor owner, commit-time decided timestamp | Core 01 REQ-01-339; Core 02 §10.4.2 | web.workbook, module.workbook |
| comm_log | type, audience, channel/meeting and summary | Commit timestamp/ID; empty refs; null optional text/time | Core 01 REQ-01-503/674 | web.workbook, module.artifacts, module.workbook |
| handoff | incoming member and summary | Outgoing actor, commit timestamp/ID; empty refs; nullable next checks/acknowledgement | Core 01 REQ-01-504/674 | web.workbook, module.artifacts, module.workbook |
| status_review | summary | Review actor, commit timestamp/ID; empty refs; nullable risks/time | Core 01 REQ-01-505/674 | web.workbook, module.artifacts, module.workbook |
| lesson | summary | Actor owner, open closure, commit timestamp/ID; empty refs | Core 01 REQ-01-506/674 | web.workbook, module.artifacts, module.workbook |
| findings | statement | Finding/open, actor owner, nullable confidence/closure | Core 01 REQ-01-507; Core 02 §10.4.6 | web.workbook, module.artifacts, module.workbook |
| investigative_queries | platform, purpose and query text | Server ID/actor/time | Core 01 REQ-01-508; Core 02 §10.4.6 | web.workbook, module.artifacts, module.workbook |
| forensic_keywords | pattern and reason | Literal match, false case sensitivity | Core 01 REQ-01-509; Core 02 §10.4.6 | web.workbook, module.artifacts, module.workbook |
| evidence | qualifying authored metadata | Omitted lifecycle requested; omitted requested time filled; explicit null suppresses timestamp default; Party IDs alone do not qualify | Core 01 REQ-01-328; Core 02 §13 | web.workbook, module.evidence, module.workbook |
| indicators | determinate canonical identity | Type/kind/display, applicable normalization, paired hashes; reuse does not enrich | Core 01 REQ-01-331; Core 02 §10.2 | web.workbook, module.indicators, module.workbook |

All entries use the existing `createViewRow` operation and incident/view route,
actor/incident/view/transaction idempotency, complete `view_row_v1` acceptance,
and current editor/reviewer/admin authority. Discovery and field behavior are
owned by Core 01 §§7.4/19, not frontend policies or tests. Fields and create-only
inputs are separate. Ordinary coordination omits `coordination.source_record_id`;
its nullable contextual input remains owned by REQ-01-674. Ordinary Evidence
metadata does not expose raw `evidence.initial_object_blob_id`; the existing
attachment owner retains that capability. General Evidence-only discovery prose
is a stale restatement against the specific primary registries (Core 00
REQ-00-044 and Core 01 REQ-01-309), not authority to remove coordination inputs.

Source ownership: web.workbook, composed with existing transport/services and
source-owned feature preparation. Query placement, semantic grid continuity,
runtime refresh debt and independent verification ownership remain unchanged.

### Declared ordinary field reachability

The following inventory uses authored field keys (schema prefix omitted only
inside each cell). Visible fields use the trailing draft row; hidden fields use
the existing Columns control and, on generic surfaces, supplementary inspector
authoring. Hidden Entity seeds remain grid fields. Technical identity/version
columns never become authoring inputs. Explicit Commit is the only trigger.

| Schema | Visible create fields | Hidden create fields |
| --- | --- | --- |
| hosts | `host.display_name`, `host.hostname`, `host.aliases`, `host.location`, `host.os_platform`, `host.business_owner`, `host.criticality`, `host.containment_status` | `host.aad_device_id`, `host.fqdn` |
| identities | `identity.display_name`, `identity.upn`, `identity.email`, `identity.sam_account_name`, `identity.aliases`, `identity.privilege_level`, `identity.mfa_state`, `identity.reset_status` | `identity.aad_object_id`, `identity.sid` |
| parties | `party.display_name`, `party.party_kind`, `party.organization_name`, `party.role_title`, `party.primary_email`, `party.timezone_name`, `party.external_ref` | `party.notes` |
| task_requests | `task.title`, `task.status`, `task.owner_user_id`, `task.priority`, `task.task_kind`, `task.workstream`, `task.due_at`, `task.requester_party_text`, `task.blocked_reason`, `task.completed_at`, `task.external_ticket_ref` | `task.requester_party_id`, `task.closure_summary`, `task.linked_record_ids`, `task.decision_record_id` |
| decisions | `decision.summary`, `decision.status`, `decision.owner_user_id`, `decision.decision_type`, `decision.decided_at`, `decision.rationale`, `decision.support_refs` | `decision.affected_record_ids` |
| comm_log | `comm_log.timestamp_utc`, `comm_log.comm_type`, `comm_log.audience`, `comm_log.channel_or_meeting`, `comm_log.summary`, `comm_log.next_report_at` | `comm_log.privilege_tag`, `comm_log.decision_ids`, `comm_log.action_task_ids`, `comm_log.audience_party_ids`, `comm_log.attendee_party_ids` |
| handoff | `handoff.timestamp_utc`, `handoff.outgoing_owner_user_id`, `handoff.incoming_owner_user_id`, `handoff.current_state_summary`, `handoff.next_checks`, `handoff.acknowledged_at` | `handoff.open_task_ids`, `handoff.open_decision_ids`, `handoff.open_risk_refs` |
| status_review | `status_review.timestamp_utc`, `status_review.review_owner_user_id`, `status_review.current_state_summary`, `status_review.active_risks_summary`, `status_review.next_report_at` | `status_review.blocked_task_ids`, `status_review.pending_evidence_ids`, `status_review.open_decision_ids` |
| lesson | `lesson.timestamp_utc`, `lesson.summary`, `lesson.owner_user_id`, `lesson.closure_state`, `lesson.follow_up_task_ids`, `lesson.evidence_refs` | none |
| findings | `finding.statement`, `finding.kind`, `finding.state`, `finding.owner_user_id`, `finding.confidence_score` | `finding.supporting_refs`, `finding.contradictory_refs` |
| investigative_queries | `investigative_query.platform`, `investigative_query.purpose`, `investigative_query.query_text` | none |
| forensic_keywords | `forensic_keyword.pattern`, `forensic_keyword.reason`, `forensic_keyword.match_mode`, `forensic_keyword.case_sensitive` | none |
| evidence | `evidence.title`, `evidence.lifecycle_state`, `evidence.requested_at`, `evidence.received_at`, `evidence.storage_ref`, `evidence.collector_party_text`, `evidence.source_party_text` | `evidence.collector_party_id`, `evidence.source_party_id` |
| indicators | `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, `indicator.normalized_value`, `indicator.defanged_value`, `indicator.hash_algorithm`, `indicator.hash_value`, `indicator.stix_pattern` | none |

All raw values must remain separate from their normalized request values.
String bindings follow Core 01 REQ-01-489–498 (including multiline reasons,
email and the packaged timezone registry); timestamps retain declared RFC 3339
semantics; enums, direct references and collection actions follow each field.
Neither defaults nor optional references replace an authored minimum.

Characterized baseline placement: `EntityWorkbookSurface.tsx` and
`GenericWorkbookSurface.tsx` supply draft cells; generic inspector composition
and `GenericWorkbookInspectorPresentation.tsx` supply hidden fields and Commit.
Both surfaces already use semantic grid continuity and the workbook Columns
control. Current ordinary commands build new transaction identities per call;
generic feedback already distinguishes accepted writes from refresh failure.
Creation currently ends with component-owned reset/refresh; Entity creation
also selects the returned row without a retained attachment fence.

Reference controls currently use `useOwnerReferenceOptions` and
`GenericMutationControl`. Paged authoring controls already exist in
`WorkbookAuthoringReferenceControl`; ordinary integration must retain selected
identities independently of page membership. G6 is a coverage/robustness gap,
not yet a reproduced off-page loss claim.

## Lifecycle disposition

| Transition | Draft/attempt disposition | Presentation and admission |
| --- | --- | --- |
| Selection change / inspector close | Retain | Detach; restore semantic focus only for current attachment |
| Sheet / saved-view navigation | Retain by canonical schema | Same schema shares authoring; late completion never navigates |
| In-app refresh / failed refresh | Retain raw text and accepted receipt | Preserve authorized rows, focus and selection; retry reads |
| Role downgrade with visibility | Retain copyable work | No unauthorized writes; current reads permitted |
| Incident closure | Retain copyable drafts and immutable uncertain attempts | No fresh/automatic writes; exact committed replay remains explicit |
| Reopening | Retain | No automatic replay; definitive closed rejection requires fresh action/request |
| Session loss / uncertain authorization | Retain privately in same account runtime | Conceal protected snapshots; suspend writes; authoritative recovery |
| Same-account recovery | Retain | Revalidate session, incident authority and required reads; explicit recovery |
| Incident access revocation | Retain only within existing suspended incident runtime | Immediately clear protected presentation, exit to directory, preserve account session, explicit reentry |
| Different incident / account replacement / disposal | Retire | Clear previous protected material and fence obsolete callbacks |

Core 03 REQ-03-099/100/287/288/299/301/302, Core 01 §3.3.5 and REQ-01-591,
Core 04 §§1–2 and design §§8/10/12 govern these dispositions. In-app refresh is
not full document reload; this seam provides no persistence or cross-tab archive.
Acknowledged results cannot become create retries. Late prior-account responses
cannot restore protected content, and retirement never reverses server commits.

## Gap ledger

Initially these are structural findings or hypotheses; execution evidence below
must establish each claimed user-visible defect before remediation.

| Gap | Remediation / affected areas | Rationale and benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 presentation-owned ordinary drafts | Retained schema-keyed owner; remove Entity reset and generic local state | Navigation and closure preserve unfinished work | Memory-only same-account/incident lifetime; no new Entity inspector | Exact raw values survive detach/navigation/refresh and grid/inspector access |
| G2 duplicate/uncertain dispatch | Synchronous reservation and immutable attempts; migrate ordinary command ports | One logical write through response loss | Existing route and receipt hashes unchanged; preserve Entity merge admission | Same-frame activation sends once; replay bytes/ID equal original |
| G3 acceptance tied to component completion | Store complete receipt before effects; integrate refresh debt/version owners | Reads can fail without losing accepted results | Preserve newer rows, unrelated selection, and source-authoritative reuse | Delayed receipt retained; acknowledged recovery sends no creation |
| G4 newer text erased by earlier response | Separate draft identity/revision from attempt | Continued authoring remains safe | Later text is next draft, never implicit patch | Older completion cannot clear changed draft |
| G5 omission/normalization/readiness conflation | Target-owned preparation over explicit raw/null/omission values | Owner-correct defaults, identity and raw-text preservation | Shared helper consumers must retain specialized semantics | All target fields/minima/defaults and explicit null versus omission pass |
| G6 reference discovery and hidden fields | Existing neutral paged identity controls, target restrictions and supported field visibility | Retained off-page references and usable optional inputs | No new associations or navigation; role loss must conceal labels | Paging cannot drop chosen IDs; loading/empty/failure/unavailable usable |

## Execution evidence

### ORC-01

Initial inspection reconfirmed the clean baseline and source-supported matrix.
Task guides for web.workbook, module.workbook, module.entities and module.artifacts
were read during planning. The new narrow rows were run against unchanged product source:

`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.ordinary_create_lifecycle,web.workbook.regression.ordinary_create_omission`

Run `.cartulary/test-results/20260913T203955Z-p33324` is expected red
(1/3 units passed). The target summary is `target-summaries/test-slice.json`;
assertions are retained under each row's `unit-logs/.../vitest-failure-details.json`.
Eight lifecycle assertions fail: Host/Evidence navigation returns empty drafts;
same-frame activation sends twice; earlier acceptance clears newer text;
response loss and malformed success provide no exact-replay action; detached
acceptance has no retained feedback; refresh failure has no read-only recovery.
The omission assertion confirms untouched Evidence requested/received timestamps
are sent as null. Closing an empty Entity inspector PASSES; the more general
reset-callback concern remains structural and is not claimed as a universal
closure defect. Fixture responses are cloned so duplicate dispatch is observed
without a response-body reuse error.

Initial harness routing attempts were rejected before test execution for
unsorted row/title inputs and interpolated test titles. Authored routing now
uses sorted literal selectors. These setup failures are related and corrected.
No generation or product behavior was changed to make the reproductions run.

G1/G4 are governed by Core 03 continuity and REQ-03-299/301/302; G2/G3 by
Core 01 §3.3.5 and Core 03 pending/acknowledgement recovery; G5/G6 by
Core 01 §§7.4/19 and the target Core 02 owners in the entry matrix. G1–G5 are
confirmed for the specific reproductions above; G6 is prospective protection.
Each remediation retains public routes, server deduplication/attribution and
stored receipts. No indefinite local-owner compatibility path is required.

Production browser characterization used
`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.the_browser_workbook_exercises_parties_notes_tas_8ca985fff2`.
Run `.cartulary/test-results/20260913T204430Z-p97617` completed all fourteen
new production-renderer discovery steps, including generic supplementary
inspector Commit controls. The subsequent existing Party validation assertion
failed on a stale/ambiguous selector; its expected message is present in the
retained accessibility snapshot. The assertion now selects the alert containing
the expected text. Earlier runs `20260913T204036Z-p34420` and
`20260913T204329Z-p66196` exposed that stale selector and an offscreen virtualized
Commit control; the characterization now uses the existing grid-scroll helper.
These are related test repairs, not unavailable production paths. Browser
summary/report/trace are under `browser-e2e-webserver-backed/browser-groups/functional-support-default-workbook-generic/`.

ORC-01 exit: DONE. Changed paths are this handoff,
`WorkbookShell.surfaces.test.tsx`, `genericWorkbookModel.test.ts`,
`workbook.generic.spec.ts` and authored `tools/test_families/web.workbook.json`.
No product files changed. All entries/transitions have explicit dispositions and
confirmed gaps have deterministic red assertions and binary remediation exits.
The common boundary will own draft lifetime, authority, immutable attempts and
receipts; source contributions will own preparation and reference restrictions.
A future adopted schema would register a preparation/reference contribution at
composition, without a new lifecycle branch. No extension is implemented.
Remaining risks are stale-read admission, reference paging, security recovery
and specialized regression coverage; these are explicit ORC-02–04 obligations.
Next action: begin retained authoring and target preparation in ORC-02.

 The source matrix and
all lifecycle transitions have explicit dispositions; no primary-owner
contradiction requires a new decision.

### ORC-02

Exit: DONE. The existing runtime/registry now retains one ordinary draft per
schema. Generic and Entity presentation state borrows that owner. Inspector
resets no longer clear ordinary Entity authoring. Raw/null/omission intent,
displayed defaults and normalized preparation are distinct. Copyable drafts
remain visible for readable role loss/closure; suspension hides protected
snapshots. Paged reference identity lives with the draft, with compact grid
pickers, semantic focus return and existing staged reference behavior.

Changed implementation areas: `features/ordinary/*`, source contributions under
`features/{entities,coordination,parties,artifacts,evidence,indicators}`, neutral
`models/workbookAuthoringValues.ts`, runtime/shell/infrastructure composition,
Entity/generic presentation and creation hooks, generic mutation control and
`WorkbookAuthoringReferenceControl.tsx`. Source guides and authored frontend
ownership/routing are updated. Source-owned initial Task/Decision guards and
Coordination text/timestamp preparation share neutral decisions with existing
contextual consumers. No specialized lifecycle owner was replaced.

The shared generic builder now omits untouched fields; real specialized helper
consumers remain. Party timezone readiness consumes the existing packaged
registry through a generated TypeScript projection registered in
`contracts/index.json` and the existing protocol HTTP facade. `make generate`
PASS at `.cartulary/test-results/20260913T210114Z-p71685` generated the new
registry and routing index; generated files were not hand-edited.

Focused command:
`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.ordinary_create_authoring,web.workbook.regression.ordinary_create_omission,web.workbook.regression.ordinary_create_continuity,web.workbook.regression.ordinary_create_controls,web.workbook.regression.coordination_create_authoring,web.workbook.regression.contextual_task_decision_authoring,web.workbook.regression.note_create_authoring`
PASS 8/8 units at `.cartulary/test-results/20260913T210445Z-p75753`.
Additional timezone, initial lifecycle and Indicator constraint assertions passed
with `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.ordinary_create_authoring`
at `.cartulary/test-results/20260913T210553Z-p77978` (2/2 units).
`make frontend-typecheck` PASS at
`.cartulary/test-results/20260913T210720Z-p79213` (2/2 units).
`git diff --check` PASS.

Earlier related typecheck/format failures identified fixture typing, a promise
condition and an attachment-hook dependency; these are corrected. `make format`
at `20260913T205823Z-p64984` and `make lint-biome` at
`20260913T210053Z-p71244` failed on those two new diagnostics (existing informational
suggestions were not changed). Subsequent `make format-frontend` passed.
The lifecycle row at `20260913T205913Z-p69673` now passes navigation/closure
assertions but remains red for submission/recovery; those cases were separated
from the continuity row and are the explicit ORC-03 dependency.

Production characterization subsequently reached Notes after all fourteen
ordinary discovery steps and Party creation. Run `20260913T204629Z-p29410`
failed only the old title-case expectation against the existing Note minimum
message; that assertion now uses case-insensitive semantic text. This is not a
new Notes defect or a claimed full browser pass.

Decisions/compatibility: keep source-specific normalization and canonical reuse
on existing wire contracts, no browser persistence, no new standalone Entity
inspector, no invented coordination source. Server authority still decides
reference existence, exact reuse and commit-time guards. Risks remaining for
ORC-03/04 are immutable transport, late receipts, reconciliation, current-authority
rechecks and full service/browser/security regression evidence. Next action:
implement captured attempts and migrate the old one-shot ordinary commands.

### Dependent execution

Each exit below was saved before its dependent started. Historical in-progress
and blocker entries preserve the execution record; the top tracker and terminal
assessment give the current disposition.

## Final acceptance and rollback

The terminal assessment follows the execution evidence. Applicable digest
A001–A027 require PASS; N/A needs a scope/owner rationale. Historical failed runs
remain recorded with their superseding passes; they are not fresh success claims.

Rollback reverses only this seam's coherent authored specifications/projections,
generated output, source, tests and routing. Preserve unrelated changes and
committed analyst records; never compensate with deletes or recreation. No data
migration, persistent browser store, commit, push or deployment is planned.

## ORC-03 exit evidence

DONE. Every ordinary grid and supplementary inspector submits through
`WorkbookOrdinaryCreateOwner`. The incident runtime retains schema-local
admission, captured operation/path/body/request/actor/incident/draft revision,
complete receipts and HTTP outcomes, and read-only reconciliation obligations.
Capture and reservation precede asynchronous authority and discovery reads.
Transport loss, ambiguous failures and invalid success remain uncertain; exact
replay uses captured bytes with current credentials. Rejection preserves raw
text. Matching accepted drafts reset; newer revisions become a separate unsent
draft. Acceptance never selects a row, steals focus, opens an inspector, or
inserts rows into a filtered/paged query that has not admitted them.

Changed paths: ordinary owner/contracts/controls/tests; ordinary transport and
`workbookCreateCapability`; Entity and generic presentations/composition; mutation
command ports; runtime and shell infrastructure; Entity/generic query owners and
`ordinaryCreateQuery.test.tsx`; shared discovery fixture and source guides;
authored frontend ownership and web.workbook routing. Entity reservations remain
held through uncertainty. History versions, transaction acknowledgement, surface
refresh debt and reference refresh integrate through their existing owners.
Shared discovery comparison now serves ordinary and contextual creation.
Indicator receipt validation remains Indicator-owned, including neutral HTTP
status semantics and unchanged canonical reuse.

Retired: generic/entity `canCreateRecord` and one-shot `createRecord` commands,
component-owned ordinary drafts, ordinary reset callbacks, and ordinary local
pending/error/receipt settlement. Existing-record patch, paste and merge commands
remain. The pure generic builder still has Party creation, Indicator observation
and Timeline-related consumers. Notes, Assessments, contextual Task/Decision,
coordination and Timeline Evidence retain their lifecycle owners. Semantic focus
stays with the existing grid/continuity owners; no Timeline key/trigger policy was
copied to ordinary sheets.

Evidence (all commands from repository root):

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.ordinary_create_submission,web.workbook.regression.ordinary_create_lifecycle,web.workbook.regression.ordinary_create_continuity`
  at `20260913T212822Z-p85228`: owner submission and continuity PASS; lifecycle
  fixture lacked the new public discovery read. Added a typed-contract-derived
  public schema fixture; lifecycle PASS at `20260913T213036Z-p87232` (2/2).
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.ordinary_create_query,web.workbook.regression.ordinary_create_submission,web.workbook.regression.ordinary_create_transport,web.workbook.regression.ordinary_create_lifecycle,web.workbook.regression.ordinary_create_continuity,web.workbook.regression.ordinary_create_controls,web.workbook.regression.ordinary_create_authoring,web.workbook.regression.ordinary_create_omission,web.workbook.regression.contextual_task_decision_discovery,web.workbook.regression.indicator_canonical_transport,web.workbook.regression.workbookz_mutation_runtime_semantic_command_ports_7b2f1d02a4`
  at `20260913T214024Z-p96697`: 11/12 PASS. The new authorization query gate
  exposed an initial Entity-load edge. Queries now resume authorized reads on
  authority restoration, without dispatching mutations.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.ordinary_create_continuity,web.workbook.regression.ordinary_create_lifecycle,web.workbook.regression.ordinary_create_query`
  PASS 4/4 at `20260913T214205Z-p99864`. Query tests cover every generic schema,
  both Entity schemas, stale version floors, filtered membership, concealment and
  obsolete reads. Owner tests cover uncertain authority and explicit revalidation.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.note_create_recovery,web.workbook.regression.coordination_create_recovery,web.workbook.regression.contextual_task_decision_recovery,web.workbook.regression.timeline_related_evidence_recovery,web.workbook.regression.entity_merge_admission,web.workbook.regression.entity_merge_query_concealment,web.workbook.regression.ordinary_create_transport,web.workbook.regression.ordinary_create_submission`
  at `20260913T214255Z-p1574`: 7/9 PASS. Two transport assertions needed to include
  the newly retained HTTP status; full receipt assertions remain exact.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.contextual_task_decision_recovery,web.workbook.regression.timeline_related_evidence_recovery,web.workbook.regression.ordinary_create_transport`
  PASS 4/4 at `20260913T214438Z-p4019`, including fresh CSRF credentials on exact
  replay. No specialized lifecycle was migrated.
- `make frontend-typecheck` PASS 2/2 at `20260913T214530Z-p5048`.
  `make lint-biome` PASS 2/2 at `20260913T214239Z-p1099`.
  `make format-frontend` and `git diff --check` PASS.

Run roots above are under `.cartulary/test-results/`. Row diagnostics are under
`unit-logs/row-<row-id>/vitest-failure-details.json`; graph summaries are retained
in their run roots. Earlier related failures at `212505Z-p83973`,
`213301Z-p88702`, `213310Z-p89139`, `213409Z-p91062`, `213503Z-p92407`,
`213856Z-p94460` and `214002Z-p96139` identified integration typing/dependencies,
fixture content type/minimum changes and missing fixture arguments. Their
corrections have focused passing evidence above. No unrelated repair was made.

Compatibility: public routes, requests, envelopes, stored replay identities and
Indicator legacy receipts are unchanged. The shared internal sender adds HTTP
status to accepted outcomes; affected exact assertions were updated. No persistent
storage migration. Uncertain work remains schema-local and is excluded from
Timeline dispatch admission. A future adopted surface would contribute its
preparation/reference/receipt rules at the composition boundary, with no lifecycle
branch. No future surface was implemented.

Remaining risk: production renderer geometry, the full field/minimum matrix,
service-backed identity/authorization/history contracts, spreadsheet capture,
accessibility and reviewed visual evidence require ORC-04. Next action: exercise
those integrated scenarios and close every applicable regression before terminal
validation. No blocked owner decision remains.


## ORC-04 — Integrated evidence

The complete fourteen-schema minimum matrix and the existing generic mutation
browser path pass through production controls. Reference test drivers now mount
virtualized columns and use the paged picker's accessible combobox/listbox names.
The old implicit-label lookup included option text and was not a valid selector
for this presentation. All eight current Timeline spreadsheet rows pass, including
uncertain creation, virtualization, keyboard, pointer, rejection, refresh and bulk
operations. The new ordinary uncertain-recovery browser case also proves continued
Timeline capture while another schema retains uncertainty.

The complete declared-field preparation test covers every ordinary create field,
explicit null disposition and excluded create input. Task blocked/completed fields
require their matching state; closure summary, Decision decided-at and Handoff
acknowledged-at remain allowed by their source contracts. Timezones use the adopted
registry (the fixture uses America/New_York); hash identity uses the registered
sha256 type. Test guesses of UTC/file_hash and broader initial-state prohibitions
were corrected against the source owners, without changing product contracts.

### G7 — Read-only production draft reachability

- Reproduction: service-backed browser role downgrade followed by explicit
  authority revalidation retained owner values but the production Grid Adapter
  suppressed its trailing draft. Component-only owner tests did not expose that
  producer/renderer interaction. Run `20260913T221113Z-p77123` failed the copyable
  draft assertion.
- Affected areas: ordinary Entity and generic sheets, role downgrade and incident
  closure; the Grid Adapter's read-only creation policy remains its own boundary.
- Owner/rationale: Core 03 REQ-03-299/301/302 and §§13/18/19 require authorized
  retained local work to remain usable within its security lifetime. Retention
  in an inaccessible store does not satisfy readable/copyable authoring.
- Remediation/benefit: the local ordinary recovery area exposes a bounded,
  keyboard-reachable read-only draft with exact raw text and retained reference
  labels when writing is unavailable. All declared authored fields remain
  copyable without introducing another editor or Timeline policy branch.
- Compatibility/migration: no wire/storage migration, new navigation surface or
  Grid Adapter capability. The production grid continues suppressing creation
  in read-only mode. Authorized drafts remain in the existing ordinary owner.
- Unresolved risk: production closure/revocation and read-only recovery geometry
  are being revalidated; focused tests also require immediate concealment.
- Binary acceptance: viewer/closed authoring is copyable; create dispatch is
  prohibited; suspension/revocation removes protected text immediately; obsolete
  completion cannot reopen the retired presentation. All assertions must PASS.

Current successful service/renderer runs (under `.cartulary/test-results/`):

- `20260913T214947Z-p39492`: `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120,module.workbook.integration.evidence_row_create_and_patch_through_workbook_p_d074e288e1,module.workbook.integration.parties_and_coordination_system_views_query_rout_e0808abf39` PASS 3/3.
- `20260913T214947Z-p39498`: `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.integration.host_and_identity_create_routes_reuse_exact_matc_f443e2591f,module.entities.support_integration.integration_entity_create_idempotency_is_actor_s_3c8ed1fe5c,module.entities.store.exact_match_precedence_resolves_reuse_decisions_949743e789` PASS 5/5.
- `20260913T215727Z-p79856`: `make service-backed-test-slice OWNER=module.artifacts ROWS=module.artifacts.source_mutations.artifact_source_create_and_patch_preserve_all_ei_c8d63cf5cd,module.artifacts.source_mutations.coordination_source_atomic_creation,module.artifacts.linked_notes.artifact_linked_note_creation_is_atomic_across_f_bce1ac123b` PASS 3/3.
- `20260913T215727Z-p79865`: `make service-backed-test-slice OWNER=module.parties ROWS=module.parties.store.party_create_reuses_exactly_one_active_same_inci_f7f18f1c5c,module.parties.claims.party_active_claims_are_migration_safe_concurren_9ad2db3764` PASS 4/4.
- `20260913T220357Z-p82413`: `make service-backed-test-slice OWNER=module.indicators ROWS=module.indicators.integration.indicator_create_and_query_routes_persist_canoni_e3a29b7aaf,module.indicators.identity.direct_create_serializes_transaction_identity_an_8f843da442,module.indicators.target_resolution.role_failures_are_atomic` PASS 5/5.
- `20260913T220356Z-p82175`: `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.ordinary_create_matrix,module.workbook.browser.the_browser_workbook_exercises_parties_notes_tas_8ca985fff2` PASS 13/13.
- `20260913T220304Z-p50388`: `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.spreadsheet_bulk,module.timeline.browser.spreadsheet_creation,module.timeline.browser.spreadsheet_keyboard,module.timeline.browser.spreadsheet_pointer,module.timeline.browser.spreadsheet_refresh,module.timeline.browser.spreadsheet_rejection,module.timeline.browser.spreadsheet_uncertain_create,module.timeline.browser.spreadsheet_virtualization` PASS 11/11.
- `20260913T220656Z-p36083`: `make test-slice OWNER=module.evidence ROWS=module.evidence.unit.create_signal_and_initial_lifecycle_matrix_4f9a8b671c` PASS 1/1.
- `20260913T220658Z-p36345`: `make test-slice OWNER=module.indicators ROWS=module.indicators.unit.create_admission_preserves_wire_and_replay_7ef2ad611c,module.indicators.identity.indicator_identity_normalization_and_dedupe_are_50da60e4b0` PASS 2/2.
- `20260913T221522Z-p45225`: `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.ordinary_create_authoring` PASS 2/2, including all declared fields.

New browser rows cover shared grid/inspector authoring and same-frame activation,
newer typing, lost/malformed committed responses, exact replay, detached acceptance
and read-only refresh recovery. New security rows cover closure/reopening and
role loss plus scoped revocation with delayed acceptance. The runtime registry
unit row now covers ordinary same-account recovery and incident/account/disposal
retirement. Accessibility and visual cases exercise production reference controls
and local recovery at 1280x720 and 390x480. Visual evidence is still pending review
and Make-owned golden maintenance; ORC-04 is not complete.


### ORC-04 validation dependency and follow-up

The production read-only reachability correction now passes:
`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.ordinary_create_closure,module.workbook.browser.ordinary_create_revocation`
PASS 11/11 at `20260913T221904Z-p68254`. Closure preserves exact uncertain replay;
reopening uses the existing Incident Controls action and never auto-creates the
next draft. Role downgrade is established at the existing current-authority read;
unauthorized fresh dispatch is prevented. Scoped revocation immediately exits and
conceals authoring, retains the account session, and fences delayed acceptance.

`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.ordinary_create_authoring_recovery,module.workbook.accessibility.coordination_create_authoring_recovery,module.workbook.visual.ordinary_create_authoring_recovery`
at `20260913T222213Z-p40347` passes both accessibility rows, including Tab traversal
inside the compact picker, Escape return, visible focus, names, reduced motion and
narrow recovery. Visual functional assertions pass; four newly authored captures
lack committed goldens. The retained reconciliation v3 report is
`browser-e2e-visual/frontend-visual-reconciliation.json` under that root: four
capture intents, zero ambiguous mappings, four missing new goldens. Its 248
unobserved existing goldens are not deletion candidates: this was a narrow run.
The new desktop/narrow reference, uncertain recovery and read-only authoring
images have been inspected. Read-only draft content is bounded and scrollable;
recovery actions now precede the copyable fields so they remain reachable.

The first full `make browser-e2e-visual` at `20260913T221151Z-p9044` failed on
expected ordinary-control diffs/new captures and three existing conflict fixture
assertions. It also exposed incorrect new scenario-ID spelling; new authored
scenario IDs now use the required stable twelve-hex form, generated routing was
regenerated, and reconciliation schema validation passes in the later narrow run.
No golden has been promoted. Existing-control differences are concentrated in
ordinary reference controls/defaults and supplementary inspector content.

**Blocked dependency, decision requested:** the public
`make browser-e2e-visual-update` target selects the full 44-row visual inventory
and exposes no row-selection input. The three conflict failures reproduce at clean
`83fd4165` in isolated `/tmp/cartulary-orc-baseline`, after Make-owned frontend
installation. `make service-backed-test-slice OWNER=module.collaboration ROWS=module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1,module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c,module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc`
fails identically at baseline run
`/tmp/cartulary-orc-baseline/.cartulary/test-results/20260913T222052Z-p4890`:
`stabilizeConflictResolverVisual` finds the existing blocked-edit message where it
asserts absence. Baseline presence, Conflict-strip and recovered-strip captures
pass. Removing the fixture's later Escape in an isolated investigation did not
remove the failure (`20260913T222448Z-p74616`); that experimental change is not in
the implementation checkout. This is not an ordinary-creation regression. The
user's no-unrelated-repairs instruction requires a scope decision before absorbing
its fix. ORC-04 remains IN_PROGRESS; ORC-05 remains NOT_STARTED. No assertion,
comparison tolerance, golden or product conflict behavior has been weakened.

Additional focused evidence:

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.ordinary_create_authoring,web.workbook.regression.ordinary_create_controls,web.workbook.regression.ordinary_create_submission,web.workbook.regression.ordinary_create_transport,web.workbook.regression.ordinary_create_query,web.workbook.regression.ordinary_create_continuity,web.workbook.regression.ordinary_create_lifecycle,web.workbook.regression.ordinary_create_omission,web.workbook.regression.note_create_recovery,web.workbook.regression.coordination_create_recovery,web.workbook.regression.contextual_task_decision_recovery,web.workbook.regression.timeline_related_evidence_recovery,web.workbook.regression.entity_merge_admission,web.workbook.regression.entity_merge_query_concealment`
  PASS 15/15 at `20260913T222928Z-p12040`.
- `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19,web.architecture.boundary_support.transportboundarypolicy_suite_aab5aaaedb,web.architecture.support.enforce_frontend_import_boundaries_apps_web_must_cc56151788`
  PASS 4/4 at `20260913T222927Z-p11344`. Earlier failure at
  `20260913T222547Z-p6933` identified a type-import cycle (removed by using the
  existing protocol receipt type), missing timezone HTTP-facade registration
  (authored owner input updated), and stale wire-intent placement assertions
  (exact source-owner paths updated, including retained Note/coordination owners).
- `make test-slice OWNER=package.view_contracts ROWS=package.view_contracts.frontend_unit.contracts`
  PASS 2/2 at `20260913T222548Z-p7158`.
- `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.index_suite_d804ef9789,package.grid_adapter.regression.core_maps_filtered_overflow_to_explicit_create_r_e4b08f0386,package.grid_adapter.regression.core_orders_invalid_pending_active_selected_read_cd0d921838`
  PASS 4/4 at `20260913T222550Z-p7541`.
- `make generate` PASS at `20260913T222925Z-p10801`; derivatives follow authored
  routing/contract inputs. `make frontend-typecheck` PASS 2/2 at
  `20260913T222154Z-p39678`.

A browser attempt at `20260913T221602Z-p46014` was correctly rejected before
execution because frontend source changed during its artifact capture; the later
passing security run uses a stable completed build. Other related intermediate
fixture/preparation failures at `20260913T215733Z-p91446`,
`20260913T220655Z-p35860`, `20260913T221113Z-p77123`,
`20260913T220923Z-p73904`, `20260913T221100Z-p76538`, and
`20260913T221206Z-p36318` are superseded by the explicit passing evidence above.


Further specialized regression evidence, all from this implementation checkout:

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.assessment_append_recovery,web.workbook.regression.assessment_reconciliation,web.workbook.regression.assessment_socket_order,web.workbook.regression.workbookshell_assessments_suite_f2e4a841a2,web.workbook.regression.workbookz_conflict_saved_merge_start_8bb0ee5972,web.workbook.regression.workbookz_conflict_typed_collection_actions_16d0ff6b08,web.workbook.regression.workbookz_conflict_unknown_class_atomic_60e26b5ba2`
  PASS 8/8 at `20260913T223158Z-p17376`.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.note_create_recovery,module.workbook.browser.note_create_revocation,module.workbook.browser.coordination_create_response_loss,module.workbook.browser.coordination_create_revocation,module.workbook.browser.ordinary_create_closure,module.workbook.browser.ordinary_create_revocation,module.workbook.browser.ordinary_create_continuity,module.workbook.browser.ordinary_create_uncertain_timeline,module.workbook.browser.ordinary_create_detached_refresh`
  PASS 17/17 at `20260913T223159Z-p17603`.
- `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.contextual_task_replay`
  PASS 11/11 at `20260913T223242Z-p50908`.
- `make service-backed-test-slice OWNER=module.assessments ROWS=module.assessments.browser.contextual_decision_refresh`
  PASS 11/11 at `20260913T223243Z-p51153`.
- `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser.contextual_creation_discard,module.evidence.browser.timeline_related_creation_replay,module.evidence.browser.timeline_related_partial_creation,module.evidence.browser.timeline_related_role_loss`
  PASS 13/13 at `20260913T223244Z-p51471`.
- `make lint-biome` PASS 2/2 at `20260913T223346Z-p46670`.
  `make lint-markdown` PASS at `20260913T223223Z-p47589`, summary
  `adhoc/lint-markdown/tool-run-summary.json`; later handoff edits require renewal.
  `git diff --check` PASS.

The scoped ordinary implementation and specialized functional regressions are
passing. Final visual promotion, applicable generation/drift/finalization and the
completed-byte acceptance assessment remain dependent on ORC-04 closure. No
historical result or successful sibling row is used to mark the blocked visual
acceptance PASS.


Interim acceptance gates: A022 (reviewed visual fixtures) is BLOCKED by the
full-inventory conflict prerequisite and pending golden promotion; A027 (terminal
handoff) is BLOCKED by its ORC-04 dependency. The remaining final acceptance
assessment belongs to ORC-05 and has not been declared complete. `RESULTS_DIR`
remains unset. `make agent-finalize`, terminal generation/drift/policy checks,
post-promotion visual runs and completed-byte validation have not been run as a
terminal bundle because that dependent workstream has not started. Retained-run
maintenance has not been requested. No scope extension is assumed from elapsed
waiting time.

### Latest bounded review while the scope decision is pending

`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.ordinary_create_matrix,module.workbook.browser.ordinary_create_continuity`
PASS 11/11 at `20260913T224327Z-p84415`. The matrix now uses hidden Host AAD
device and Identity SID direct seeds through the production Columns menu, whose
semantic control is `menuitemcheckbox`. The continuity scenario activates the
inspector and grid Commit controls in the same frame and observes exactly one
POST. Other direct minima retain parameterized source-owned unit coverage.
Changed paths are the ordinary browser spec and its ordinary-create support
helper; no new product behavior or test row was introduced.

The second full `make browser-e2e-visual` FAIL 10/12 at
`20260913T223453Z-p47451` confirms the same three baseline conflict assertion
failures, fifteen expected changed comparisons and four missing new captures.
Its `browser-e2e-visual/frontend-visual-reconciliation.json` records 246 capture
intents, 248 committed goldens, 242 active mappings, six unobserved conflict
goldens, four missing ordinary goldens, zero ambiguous mappings and one unresolved
registered conflict fixture. The six unobserved conflict captures follow failed
fixture execution; they are retained, not candidates for deletion.

All fifteen diff images and all four new ordinary desktop/narrow images were
reviewed from that run. Task/Decision draft defaults, paged ordinary references
and supplementary inspector content explain the substantive changes. Contextual
authoring geometry remains usable; small vertical shifts are visible in the
existing contextual comparisons. The latest closed-incident narrow capture
confirms that recovery precedes the bounded, copyable read-only draft. No image
has been promoted, masked or given a looser comparison threshold. These reviews
support a future authorized golden update; they do not turn failed comparisons
into PASS.

The isolated baseline fixture experiment has been reverted, leaving its checkout
at pristine tracked baseline bytes. Its failing run artifacts remain available.
Next action remains the pending user scope decision on the independently
reproduced conflict prerequisite. ORC-04 cannot close, and ORC-05 cannot start,
until the applicable visual gate can pass.

Blocked prerequisite disposition: remediation is a separately authorized narrow
repair of the existing conflict failure, or an upstream repair before resuming.
Affected areas are the collaboration visual fixture, its existing Timeline
conflict owner, golden maintenance and this handoff's visual gate. The adopted
spreadsheet/conflict contract remains controlling; catalog selection and the
visual golden maintenance guide route verification without changing behavior.
The rationale and long-term benefit are trustworthy conflict regression evidence
without bypassing assertions or deleting unobserved captures. No compatibility
change is authorized. Root cause beyond the reproduced blocked-edit assertion
remains unresolved. Binary validation is all three failing rows passing with
their assertions intact, followed by successful Make-owned full-inventory golden
maintenance and the required fresh visual verification runs.

The final two-row browser edit required formatting only. `make lint-biome`
initially failed at `20260913T224549Z-p16590` and
`20260913T224644Z-p18918` on that spec's formatting; `make format` PASS 2/2 at
`20260913T224658Z-p19466` corrected it, and `make lint-biome` PASS 2/2 at
`20260913T224712Z-p23798`. No Go source changed. `git diff --check` PASS.
`make lint-markdown` PASS at `20260913T224606Z-p17188`, with subsequent handoff
edits requiring one further documentation-only validation reported in the
interruption response. Branch and HEAD remain `main` / `83fd4165`; the isolated
baseline checkout is clean. Product, routing and test changes remain uncommitted.

### G8 — Authorized conflict prerequisite repair

The user explicitly authorized a narrow fix for the three pre-existing conflict
visual failures, preserving their assertions. This resolves the scope-decision
blocker; ORC-04 remains IN_PROGRESS until the repair and visual validation pass.
Revalidation on resumption found `main` at `83fd4165` with the preceding seam
changes now staged. Preserve that index state and unrelated work; subsequent
repair edits remain reviewable without committing or changing analyst records.
The prior baseline reproductions remain evidence, not a fresh pass. Next action:
trace the redundant blocked-edit transition, add regression evidence at its
existing owner, and rerun the three unchanged visual assertions before golden
maintenance.

The authorized prerequisite's cause is now confirmed: `WorkbookActiveSurfaceFrame`
applied `inert` and `aria-hidden` to all active content when conflict recovery
opened. A real browser then blurred the original Timeline editor, attempted its
blur commit against the registered conflict, and failed to restore focus into
the inert subtree. The diagnostic run at `20260913T225650Z-p28585` reproduces
this sequence; temporary focus instrumentation was removed after inspection.

Remediation removes the redundant inert wrapper from the existing shell frame;
recovery panels keep their existing mutually exclusive presentation and conflict
owners keep their dispatch guards. Core 03 REQ-03-300/218 require the exact draft
and original focusable editor after rejection, and §3.3 prohibits automatic
same-field conflict retry. Preserving that editor avoids both the redundant blur
attempt and the inaccessible correction surface. No grid-adapter, Timeline
mutation/keyboard algorithm, wire contract, or stored record changes are needed.
Existing inspector-overlay inert behavior and authorization concealment remain
owned by their respective layout and security boundaries.

`WorkbookActiveSurfaceFrame.test.tsx` fails before the fix because its original
editor becomes inaccessible (`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.conflict_editor_accessibility`,
`20260913T230117Z-p61501`), then passes 2/2 at `20260913T230210Z-p67138` after
the fix. It checks draft and element identity, accessibility, focus, resolver
dismissal and retained unresolved conflict. The first existing visual fixture
now additionally asserts retained focus and absence of hidden/inert ancestors;
the three original assertions are unchanged. Source ownership, the component
guide and the new unit row are authored inputs; `make generate` PASS at
`20260913T230209Z-p66574`. Remaining risk is shared-shell focus regression,
covered next by Timeline, conflict action/accessibility and full visual rows.

The three originally failing visual rows PASS 11/11 at `20260913T230238Z-p70291`
with all original assertions and goldens intact. The eight Timeline spreadsheet
rows PASS 11/11 at `20260913T230238Z-p70302`. Follow-up conflict action coverage
identified that resolver controls also need the Grid Adapter's existing
`data-grid-editor-external-action` contract: transferring focus into recovery
must not resubmit the rejected editor and pull focus back. The conflict resolver
and status activation button now declare that existing boundary.

The first mixed action/accessibility run failed at `20260913T230311Z-p28380`:
the action row exposed the focus-transfer defect; the accessibility group had
`service_readiness_timeout`, so its missing-result summary was secondary
infrastructure failure. Independent reruns at `20260913T230653Z-p65375` and
`20260913T230654Z-p65633` reached additional stale fixture assumptions: automatic
Timeline focus theft into the resolver, and movement from an unacknowledged edit
without explicit detachment. The fixtures now follow REQ-03-300/218: assert the
retained editor, explicitly activate conflict recovery before asserting summary
focus, and use Escape to detach a queued edit before capturing another. Resolver
request bodies, exact merged text, saved/local comparisons, queue count and FIFO
assertions remain intact. No mutation or keyboard policy was changed for these
fixture corrections. These scoped follow-ups remain under the authorized
conflict prerequisite repair and require fresh passing evidence.

Conflict public-action and accessibility rows now PASS separately, 11/11 each at
`20260913T231007Z-p29117` and `20260913T231008Z-p29387`. The unchanged FIFO,
stale-token and merged-request assertions pass. The related conflict unit row
initially failed at `20260913T231034Z-p86751` (2/3) and
`20260913T231233Z-p92485` on the same stale automatic-focus assumption in its
test fixture. The fixture now borrows `useWorkbookRecoveryFocus` instead of
duplicating its focus owner; its test focuses Retry and explicitly activates the
new resolver before asserting summary focus. That row passes 2/2 at
`20260913T231440Z-p97986`. The pending-queue sibling passed in the original
two-row run. The fixture source guide records the retained production owner.

### G9 — Ordinary control token and viewport review

The final ORC-04 token audit found fixed popover sizing in the new compact
reference presentation. Remediation reuses the existing workbook menu styles and
their token-backed width, then positions from actual panel measurements. A
ResizeObserver and viewport/scroll listeners keep the panel inside the viewport
as candidates load or available height changes; detachment removes all observers
and listeners. The ordinary read-only notice reserves half of the available
workbook height after token-defined shell bars. No second design, density or
breakpoint registry is introduced. This applies design direction within its
boundary and digest R011/A004/A008; those advisory rows do not become product
wire authority.

Affected paths are `WorkbookAuthoringReferenceControl`, `OrdinaryCreateNotice`
and the ordinary accessibility scenario. The rationale and benefit are shared
theme/density maintenance and operable recovery under vertical resize, without
fixed-size drift or a new navigation surface. Wire compatibility, authoring
lifetimes and stored data are unchanged. Remaining risk is popup focus/geometry
under production layout; binary acceptance is keyboard focus and full control
visibility after vertical resize, all ordinary reference unit tests, and reviewed
desktop/narrow production screenshots. The scenario adds those assertions.
`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.conflict_editor_accessibility,web.workbook.regression.ordinary_create_controls`
PASS 3/3 at `20260913T231441Z-p98227`; `make frontend-typecheck` PASS 2/2 at
`20260913T231454Z-p99200`. A fresh full visual run and ordinary browser/accessibility
selection are running before golden maintenance.

### Completed conflict and control follow-up

The user-authorized conflict repair now passes its original visual assertions,
public actions, accessibility and FIFO recovery checks. The final pending-queue
unit fixture had the same stale assumption that an unacknowledged editor closes
automatically. Explicit Escape detachment preserves the original queue ordering,
request and acknowledgement assertions. Its first run failed 1/2 at
`20260913T231902Z-p69691`; the corrected scenario passes 2/2 at
`20260913T232044Z-p76357` using
`make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_999bdc7b60`.
`make format` PASS 2/2 at `20260913T232035Z-p72104`.

Further fresh results, all under `.cartulary/test-results/` with
`run-summary.json` and `target-summaries/<target>.json`:

- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.ordinary_create_authoring_recovery,module.workbook.accessibility.coordination_create_authoring_recovery,module.workbook.browser.ordinary_create_matrix,module.workbook.browser.ordinary_create_continuity`
  PASS 13/13 at `20260913T231509Z-p99854`. This closes G9's keyboard/vertical-resize
  criterion and revalidates all fourteen production minima, hidden Entity seeds,
  ordinary reference controls and simultaneous grid/inspector activation.
- `make test-slice OWNER=package.ui ROWS=package.ui.boundary_support.design_tokens_exposes_explicit_font_role_tokens_55f6ac41b0,package.ui.boundary_support.design_tokens_exposes_the_adopted_default_theme_4f6994fbc4,package.ui.boundary_support.design_tokens_renders_css_variables_without_unre_7aba7b88a9,package.ui.frontend_unit.design_presentation_projection_8e15b85b40,package.ui.frontend_unit.workbook_interaction_contracts_dcad48388a,package.ui.frontend_unit.workbook_shell_grid_contracts_07bd0aad94`
  PASS 7/7 at `20260913T231743Z-p67546`.
- `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19,web.architecture.boundary_support.transportboundarypolicy_suite_aab5aaaedb,web.architecture.support.enforce_frontend_import_boundaries_apps_web_must_cc56151788`
  PASS 4/4 at `20260913T231744Z-p67786`.
- `make test-slice OWNER=module.collaboration ROWS=module.collaboration.frontend.save_state_labels_use_only_syncing_saved_and_con_b9c4789169,module.collaboration.frontend.the_resolver_keeps_the_grid_visible_leaves_confl_41a75525b7`
  PASS 3/3 at `20260913T231902Z-p69681`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell_resolver_keep_saved_action_submits_2d011ebd3f,web.workbook.regression.workbookshell_resolver_merged_value_action_submi_5d8f831f1e,web.workbook.regression.workbookshell_resolver_use_unsaved_action_submit_f7537f46a6`
  PASS 4/4 at `20260913T231902Z-p69711`.

The full ordinary visual run at `20260913T231508Z-p99727` fails comparisons only
(10/12 execution units). Every functional assertion, including the three original
conflict scenarios, passes. Its reconciliation v3 accounts for 252 capture
intents, 248 existing active goldens, zero orphans, zero ambiguous mappings, all
29 registered fixtures, zero unresolved fixtures, and four intentional newly
authored ordinary captures awaiting creation. No existing image is unaccounted
for or eligible for deletion. The fifteen stale comparisons follow the reviewed
ordinary-control/default/inspector change. All four latest ordinary screenshots
were inspected at their declared 1280x720 or 390x480 viewport: reference popup,
uncertain submission at both widths, and copyable closed-incident authoring.
Recovery remains operable, names and focus are visible, and the workbook grid
retains the available shell space. The accepted update trigger is the implemented
ordinary creation contract and its newly authored production fixtures. The
previous scope blocker is resolved; post-promotion validation remains pending.


### Visual golden maintenance record

`make browser-e2e-visual-update` PASS 12/12 at
`20260913T232417Z-p77771`. The Make-owned candidate promotion updated fifteen
stale comparisons, added the four new ordinary captures, and regenerated
`tools/frontend_visual_golden_manifest.json`. It retained every already-passing
image, including all original conflict goldens. Reconciliation v3 PASS accounts
for all 252 active captures/goldens, all 29 registered fixtures, zero missing,
zero ambiguous, zero orphaned and zero unresolved registered fixtures.

Accepted trigger: owner-correct ordinary draft defaults, paged references and
supplementary inspector authoring intentionally supersede the previous ordinary
controls; the four new fixtures cover that retained creation/recovery behavior.
No renderer/font/theme/density pin, viewport, zoom, mask, scroll normalization,
screenshot scope or comparison threshold changed. Existing contextual rows retain
their source/lifecycle owners; their small layout shifts follow shared reference
control geometry. No image was copied manually or deleted.

Every promoted image below was inspected after promotion, following the earlier
actual/diff review. Desktop and narrow controls, focus and recovery remain usable;
no unexplained typography, clipping, overflow or state regression was accepted.
These are implementation-support observations, not Core 05 publication claims.
The exact owner/fixture mapping comes from the retained reconciliation and authored
catalog; nonregistry captures are active, exactly reconciled consumers.
All filenames below are under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.

- Row: `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`.
  Fixture: `visual.fixture.evidence_affordance`.
  Files: `evidence-affordance-states-linux.png`.
- Row: `module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc`.
  Fixture: `visual.fixture.task_requests_or_decisions`.
  Files: `record-relationships-task-requests-linux.png`.
- Row: `module.workbook.visual.contextual_task_decision_creation`.
  Fixture: active nonregistry capture; no registered fixture identity.
  Files: `contextual-decision-authoring-linux.png`, `contextual-decision-authoring-narrow-linux.png`, `contextual-decision-recovery-linux.png`, `contextual-decision-recovery-narrow-linux.png`, `contextual-decision-references-narrow-linux.png`, `contextual-task-request-authoring-linux.png`, `contextual-task-request-authoring-narrow-linux.png`, `contextual-task-request-recovery-linux.png`, `contextual-task-request-recovery-narrow-linux.png`, `contextual-task-request-references-narrow-linux.png`.
- Row: `module.workbook.visual.decision_supersession_review_recovery`.
  Fixture: active nonregistry capture; no registered fixture identity.
  Files: `decision-supersession-accepted-linux.png`, `decision-supersession-review-linux.png`, `decision-supersession-review-narrow-linux.png`.
- Row: `module.workbook.visual.ordinary_create_authoring_recovery`.
  Fixture: active nonregistry capture; no registered fixture identity.
  Files: `ordinary-closed-retained-narrow-linux.png`, `ordinary-recovery-1280-linux.png`, `ordinary-recovery-390-linux.png`, `ordinary-reference-authoring-linux.png`.

Two fresh ordinary visual passes against the promoted manifest are required
before ORC-04 can close. Terminal policy/drift and final documentation validation
remain the ORC-05 dependency.

Conflict follow-up commands for the passing runs above:

- `make service-backed-test-slice OWNER=module.collaboration ROWS=module.collaboration.browser.verify_conflict_resolver_actions_submit_public_m_0c77fae932`
  PASS 11/11 at `20260913T231007Z-p29117`.
- `make service-backed-test-slice OWNER=module.collaboration ROWS=module.collaboration.accessibility.verify_conflict_state_resolver_controls_presence_2ec686e8dd`
  PASS 11/11 at `20260913T231008Z-p29387`.
- `make test-slice OWNER=module.collaboration ROWS=module.collaboration.frontend_unit.verify_same_field_conflict_anchors_conflict_queu_e44b13cfa1`
  PASS 2/2 at `20260913T231440Z-p97986`.

G8's binary criteria are satisfied by those runs, the three unchanged conflict
visual rows and the complete successful golden update. No unresolved product
risk remains from the repaired inert/blur transition; broader terminal regression
checks will still cover the shared shell and test fixture consumers.

First fresh post-promotion `make browser-e2e-visual` PASS 12/12 at
`20260913T232900Z-p14781`; all 252 capture/golden mappings pass with no missing,
ambiguous, orphaned or unresolved fixture entries. The second required fresh run
is in progress. Security was independently rerun after the shared frame repair:
`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.ordinary_create_closure,module.workbook.browser.ordinary_create_revocation`
PASS 11/11 at `20260913T233255Z-p49080`. This confirms closure/reopening,
copyable draft lifetime, scoped revocation/concealment and obsolete acknowledgement
fencing on the final product source.

### ORC-04 exit

DONE. Second fresh `make browser-e2e-visual` PASS 12/12 at
`20260913T233320Z-p78044`, after the first PASS at `20260913T232900Z-p14781`.
Both reconcile all 252 active images and all 29 registered fixtures without
missing, ambiguous, orphaned or unresolved entries. The terminal target summaries
and `browser-e2e-visual/frontend-visual-reconciliation.json` are retained in each
run root. Every promoted image was inspected; original conflict goldens and
assertions remain intact. This supersedes the historical visual blocker and its
interim acceptance disposition above.

G1–G5 confirmed defects pass the original deterministic reproductions and new
owner/service/browser coverage. G6's reference/hidden-field protection passes
unit and production controls. G7's read-only reachability passes closure/role
and concealment tests. G8's authorized conflict correction passes original visual,
keyboard, public action and FIFO assertions. G9's token/viewport correction passes
resize/accessibility and reviewed production images. No primary-owner decision
or applicable ORC-04 criterion remains blocked.

Changed paths since ORC-03: ordinary field/query/control/submission tests and
read-only notice; ordinary browser/security/support scenarios; a11y/visual and
generic fixtures; reference popup presentation; the authorized active-frame,
resolver and status focus boundary; production recovery-focus test fixture and
its collaboration scenarios; source guides/ownership and authored row routing;
Make-generated routing, nineteen goldens and their generated manifest; this
handoff. No Grid Adapter or Timeline keyboard/mutation algorithm changed.

Compatibility remains the existing wire/storage/replay contract, including
Indicator legacy receipts. No migrations or new dependencies are needed. The
remaining work is terminal repository consistency and completed-byte handoff
validation. Next action: start ORC-05 and run `make agent-finalize` with
`RESULTS_DIR` unset before broader terminal verification. Retained-run maintenance
will be reported as skipped; these focused runs are not a successful full warm
check source.

## ORC-05 — Terminal validation

DONE, following the saved ORC-04 DONE exit and terminal validation below. The owner-guided narrow
selections above precede this terminal bundle. Additional source owners reached
by the seam are Parties, Indicators, Evidence, Timeline, collaboration, UI/view
contracts, Grid Adapter, architecture and authored catalog routing. Broader
frontend checks cover the shared runtime/transport and recovery fixture consumers;
generation/policy/JSON checks cover authored projections, routing and promoted
goldens. Final documentation checks cover completed handoff bytes separately.

### G10 — Packaged timezone projection registration

The initial `env -u RESULTS_DIR make agent-finalize` failed 0/1 at
`20260913T233813Z-p15254`; its `unit-artifacts/finalize-summary.json` identifies
the JSON-shape prerequisite. Standalone `make json-shape-check` reproduced it
(2/3, `20260913T233839Z-p15805`): the old family guard prohibited every frontend
string-contract projection, including the newly registered timezone value.
This failure is related to this seam.

Core 01 REQ-01-498 requires exact packaged timezone membership; host Intl/tzdb
is expressly not admission authority. The authorized readiness and projection
registration work needs that public value. Remediation permits exactly the
declared timezone registry, output and identifier while continuing to reject
provenance, other neutral string artifacts and all protected Party/backend
families. The existing JSON-shape smoke test adds negative exposure fixtures;
the active harness boundary suite now executes it through a named semantic case.
No new family, duplicate registry, dependency or wire/storage contract is needed.
The benefit is deterministic readiness from the same reviewed source as the
server. Remaining risk is accidental widening of exposure; binary acceptance is
the exact positive projection plus rejected provenance/other string/Party
projections, current shape/policy/drift and frontend checks.

Changed paths are `tools/harness/generated-artifacts/check-json-shapes.mjs`,
`tools/harness/generated-artifacts/tests/test-json-shapes.mjs` and
`tools/harness/tests/contract-suite-support.mjs`. Source/verification ownership
remains the harness's existing generated-artifact and boundary suites.
`make generate` PASS at `20260913T234048Z-p17089` and
`20260913T234516Z-p98469`. No product bytes changed from this checker correction.
The `make explain-target TARGET=harness-smoke-json-shapes DETAIL=summary` lookup
returned usage error because that name is a harness check, not a public target.
The public `make harness-contract` wrapper now executes its negative tests.

`env -u RESULTS_DIR make agent-finalize` PASS 1/1 at
`20260913T234125Z-p20455`, before the broader terminal bundle. Its summary records
generated structure unchanged (zero updated files), no failures, and retained-run
selection/performance/scheduler maintenance SKIPPED because `RESULTS_DIR` was
unset. No focused run was represented as a successful full warm check.
After connecting the negative fixtures to the active boundary suite, finalization
passed again 1/1 at `20260913T234610Z-p3625`.

Terminal commands/results under `.cartulary/test-results/`:

| Command | Result | Run root suffix |
| --- | --- | --- |
| `make frontend-typecheck` | PASS 2/2 | `20260913T234148Z-p25136` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260913T234148Z-p25303` |
| `make lint-biome` | PASS 2/2 | `20260913T234148Z-p25446` |
| `make lint-scripts` | PASS 2/2 | `20260913T234148Z-p25507` |
| `make generate-drift` | PASS 4/4 | `20260913T234147Z-p24880` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260913T234148Z-p24941` |
| `make json-shape-check` | PASS 3/3 | `20260913T234148Z-p25034` |
| `make harness-contract` | PASS 2/2 | `20260913T234148Z-p25813` |
| `make harness-contract` with active negative exposure case | PASS 2/2 | `20260913T234648Z-p7633` |

Each graph run retains `run-summary.json`, its exact
`target-summaries/<target>.json`, and unit logs. Generation retains
`generate/tool-run-summary.json`. The latter harness log explicitly records
`public_timezone_projection_boundary` PASS, including its negative fixtures.
These results close G10; later test/routing edits require renewed terminal checks.

### G11 — Timeline verification fixture reconciliation

Broader `make test-slice OWNER=web.workbook` failed 247/251 at
`20260913T234147Z-p24868`. The four failing rows reproduce together both in this
checkout (`20260913T234514Z-p98084`, 1/5) and pristine baseline
`/tmp/cartulary-orc-baseline/.cartulary/test-results/20260913T234514Z-p98101`
(1/5). Exact selection:

`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell__grid_preserves_draft_row_edits_ac_cd490037ce,web.workbook.regression.workbookshell__sentinel_grid_anchor_shell_suppor_428c33b58f,web.workbook.regression.workbookshell__autosave_commits_enter_from_the_c_6dfe4292ef,web.workbook.regression.workbookshell__autosave_commits_tab_from_the_cur_f5c4976958`

The first fixture assumed qualifying Timeline input waits for Enter; it now uses
nonqualifying whitespace to verify exact pre-creation draft retention across
refresh, then qualifying text to verify automatic creation. Its accepted value
remains in the active editor through acknowledgement. The two autosave fixtures
exposed a test-renderer parity gap: production captures the current DOM control
value on Enter/Tab, but test support captured stale React state. The test renderer
now captures input/textarea/select values (including checkbox booleans) at that
same boundary. The original autosave tests and exact request/version assertions
are unchanged. The keyboard-anchor fixture now asserts horizontal Tab movement
to Data Source, the next declared visible column after Activity Synopsis;
its former immediate grid-exit expectation contradicted Core 03 REQ-03-300 and
the user's controlling spreadsheet contract. The semantic row identity remains;
its authored selector title describes vertical/horizontal movement accurately.
The declared outer shell-focus boundary remains covered by the passing production
spreadsheet keyboard row. No Timeline/Grid Adapter algorithm changes.

Broader `make browser-e2e-a11y` failed 12/14 at `20260913T234148Z-p26328` on the
global design-readiness fixture's focus-anchor assertion. The same failure
reproduces at pristine baseline using
`make service-backed-test-slice OWNER=web.design ROWS=web.design.accessibility.verify_global_accessibility_matrix_for_keyboard_90088a3666`,
run `/tmp/cartulary-orc-baseline/.cartulary/test-results/20260913T234646Z-p7358`
(9/11). It programmatically focused an unselected cell after navigation, then
expected a semantic selection anchor. The fixture now selects the cell through
its public click action and closes its editor before inspecting focus recovery.
Its anchor, accessible-name and Escape-restoration assertions remain intact.

These are owner-grounded verification corrections inside the required Timeline
regression seam, not new product behavior. Changed paths are the grid/sentinel
unit specs, Grid Adapter test-support renderer, global a11y fixture and authored
sentinel title routing. Benefit: executable evidence follows first-qualifying-input creation, editor control-value
ownership and current spreadsheet navigation. No data/wire compatibility impact.
Risk: corrected drivers might fail to exercise the intended state. Binary
acceptance requires all four focused rows and the global accessibility row to
pass, followed by renewed broader affected evidence. `make generate` PASS at
`20260913T234759Z-p41721`; `make format` PASS 2/2 at
`20260913T234801Z-p42370`. The first corrected selection passed 3/5 at
`20260913T234828Z-p49227`; remaining fixture adjustments recognize the accepted
active editor and the actual next declared column. The global accessibility row
passes 11/11 at `20260913T234829Z-p49448`. `make format` passes 2/2 at
`20260913T235130Z-p81957`; the final four-row rerun is pending.

G11 focused exit: all four original Timeline unit rows PASS 5/5 at
`20260913T235143Z-p86312`, and global accessibility PASS 11/11 at
`20260913T234829Z-p49448`. The autosave specs are byte-identical to baseline;
only the test-support renderer needed parity with production's current-control
capture. Product Grid Adapter and Timeline keyboard/mutation source remain
unchanged. The testing source guide records this distinction.
`env -u RESULTS_DIR make agent-finalize` PASS 1/1 at
`20260913T235237Z-p87435`, generated structure unchanged and retained-run
maintenance skipped. The renewed broad workbook/Grid Adapter and accessibility
checks run after that finalization.

### Terminal acceptance assessment

The digest is advisory; PASS below means the applicable seam requirement has
owner-grounded evidence in this handoff, not a new profile or release claim.
Historical interim BLOCKED entries above are superseded by the recorded repairs
and fresh passes. Unavailable Entity inspector-create entries remain N/A in the
frozen matrix because their owners expose grid creation only.

| Row | Disposition | Evidence and applicable boundary |
| --- | --- | --- |
| A001 authority | PASS | Core 00–04 and domain/design mapping, schema matrix and G1–G11 owner clauses; routing never defines behavior. |
| A002 cohesive scope | PASS | One incident/account ordinary owner; source preparation contributions; actual local paths retired; future contribution described without implementing it. |
| A003 repository state | PASS | Clean initial main/HEAD captured; later staged state preserved; authored manifests, source guides, generated roots and vendor-grid boundary reviewed. |
| A004 tokens | PASS | Existing workbook menu styles and token-defined shell dimensions; G9, UI token tests and reviewed images. |
| A005 theme | PASS | Existing dark_graphite profile and renderer/font pins unchanged; full visual comparisons pass. |
| A006 density | PASS | Shared density projection and Grid Adapter geometry retained; UI/Grid Adapter unit evidence and all visual density fixtures. No new density policy. |
| A007 creation | PASS | Fourteen-schema discovery/minimum/field/input matrix, direct seeds, no default readiness, exact reuse/dedupe, Evidence and contextual regression selections. |
| A008 responsive | PASS | Existing viewport bands/clamps preserved; popup visualViewport fallback and vertical resize; narrow/desktop accessibility and visual fixtures. |
| A009 overflow | PASS | Existing shell-owned grid/inspector scrolling, bounded copyable draft and operable popup; status/navigation remain available. |
| A010 inspector | PASS | Shared ordinary draft, supplementary hidden fields, detach/late outcomes, retained source workflows and inspector regression evidence. |
| A011 continuity | PASS | Navigation/refresh/closure/return, newer drafts, off-page identities, version floors and unrelated selection; semantic focus stays with existing owners. |
| A012 transactions | PASS | Secure IDs, synchronous admission before awaits, complete immutable capture, same-frame grid/inspector activation and exact byte replay. |
| A013 acknowledgement/recovery | PASS | Complete receipt checkpoint, detached/late acceptance, read-only refresh debt; explicit ordinary conflict recovery and independent existing FIFO actions. |
| A014 editing | PASS | Eight production Timeline spreadsheet rows, exact rejection/editor retention, paste/fill and uncertain typing; G8/G11 verify focus and current-control capture. |
| A015 conflict | PASS | Three original conflict visual assertions/goldens, exact saved/local/merged bodies, FIFO and accessibility pass after authorized narrow frame repair. |
| A016 composed states | PASS | Owner/query/readiness controls distinguish refresh/failure/unavailable from read-only/closed/concealed; current UI presentation projection tests. |
| A017 authorization lifetime | PASS | Same-account conceal/recovery, role/closure/reopen, scoped revocation, account/incident retirement and obsolete callback fencing; fresh final-source security browser pass. |
| A018 Evidence | PASS | Metadata requirements and reserved locator/lifecycle rules; service and Timeline Evidence regressions; reviewed affordance matrix retains textual state distinctions. |
| A019 accessibility | PASS | Ordinary/coordination/conflict and global keyboard/focus/name/contrast/reduced-motion rows pass with production renderers; no upstream conformance claim imported. |
| A020 components | PASS | Existing tokens, controls and compound-state contracts; popup/read-only area and inspector/reference states reviewed under declared desktop/narrow settings. |
| A021 virtualization | PASS | Production spreadsheet virtualization/refresh/uncertain rows and query version/page-membership tests; no new sizing algorithm or benchmark claim. |
| A022 visual fixtures | PASS | Make-owned update, all nineteen changed images reviewed, two fresh full visual passes, 252 captures/goldens and 29 registered fixtures reconciled. |
| A023 selectors | PASS | Source-owned schema/record/field IDs and semantic controls; package.ui selectors/interaction contract tests. |
| A024 test authority | PASS | Executable source/test audit and source ownership/import checks; no Markdown dependency, digest edit or prose-bound test. |
| A025 generated outputs | PASS | Authored contract/facade/routing inputs before Make generation; exact timezone projection guard; finalizer, generation drift, JSON and generated-policy checks. |
| A026 compatibility | PASS | Existing HTTP/routes/receipts and Indicator legacy replay retained; specialized helper consumers justified; no data/browser-store migration; coherent rollback below. |
| A027 terminal handoff | PASS | Final broad results, exact-byte documentation validation, coherent rollback and scope review complete; every sequential exit was saved before its dependent. |

Rule dispositions: R001–R015 PASS through the corresponding authority,
accessibility, state, continuity, transaction and handoff assessments. R016 and
R019 are N/A: no touch/mobile profile or typography policy is introduced; dense
desktop geometry follows its owner. R017–R018 PASS for existing desktop bands,
fallback sizing and owned horizontal grid scrolling. R020 PASS: no fake rows or
record-shaped loading data is introduced. R021 PASS: no new animation; existing
reduced-motion behavior remains verified. R022–R024 PASS: local Commit/recovery,
existing semantic controls and local feedback preserve grid geometry. R025 is
N/A to ordinary creation: no destructive action or new confirmation workflow;
existing specialized confirmations retain their owners and visual regressions.
R026–R034 PASS: no rejected decorative/marketing pattern, parallel design registry,
new behavior authority or incidental-selector policy was introduced. R035 PASS:
compact reference/read-only controls retain usable text, focus and scrolling under
the applicable existing spacing/zoom/overflow fixtures. N/A applies only to the
explicit unsupported profiles/workflows, not to any ordinary path in the matrix.

### Compatibility, limits and rollback

The supported matrix is exactly the fourteen schemas above, with Entity grid
creation and twelve generic grid/supplementary-inspector presentations. All are
registered and exposed in this checkout's production discovery. Missing runtime
capability remains an unavailable state and cannot dispatch; no fallback creates
an undeclared path. There is no standalone Host/Identity inspector create.

One raw editable draft per schema is retained only in the existing in-memory
incident/account runtime. Browser/process disposal ends that lifetime. Accepted
records remain server records; this work neither deletes nor recreates them to
compensate for lost presentation. No persistent browser storage or migration was
introduced. Server authorization/reference existence/history/exact-match rules
remain authoritative. Ordinary uncertainty blocks only its own schema; Timeline
retains its independent automatic creation and spreadsheet owner.

Retired paths: Entity/generic local ordinary drafts and create reset callbacks,
ordinary local pending/error/acceptance ownership, and one-shot ordinary create
command methods. Retained paths: Note, Assessment, contextual Task/Decision,
coordination, Timeline capture/Evidence/mentions, Party creation and Indicator
observation owners; their pure shared request helpers have real consumers.
Entity merge/write admission, FIFO conflict recovery, history/invalidation and
semantic grid focus remain existing owners. The internal sender's accepted
result adds HTTP status; every affected specialized consumer was checked.

Rollback is a coherent reverse of this seam's source contributions, runtime and
presentation migration, query/transport integration, authored contract/facade and
routing inputs, generated timezone/routing output, test/support corrections,
reviewed goldens/manifest, guides and handoff. Reverse the narrow authorized
conflict-frame fix with its resolver/status focus boundary and tests as one set.
Regenerate derivatives from restored authored inputs through Make; never restore
only a generated file or leave two ordinary lifecycle owners. Preserve unrelated
work and the pre-existing index state. Issue no compensating deletes, recreation,
SQL migration, session reset or analyst-data operation. No commit/push/deployment
was performed.

Skipped checks: retained-run performance/scheduler maintenance (`RESULTS_DIR`
unset); full warm `make check`/CI/release gates and owner evidence publication
(no release/benchmark claim or complete warm source requested); unrelated backend
lint/vulnerability and migration suites (no Go/SQL/dependency change); separate
new performance measurements (production Grid Adapter sizing/virtualization
algorithms unchanged and current production regression evidence passes).
Required affected service, browser, accessibility, visual, source-boundary and
projection checks are represented above; none is waived to close a failure.

### Final affected verification

All remaining terminal selections pass on the final product/test bytes:

| Command | Result | Run root suffix |
| --- | --- | --- |
| `make test-slice OWNER=web.workbook` | PASS 251/251 | `20260913T235317Z-p91396` |
| `make test-slice OWNER=package.grid_adapter` | PASS 44/44 | `20260913T235317Z-p91448` |
| `make browser-e2e-a11y` | PASS 14/14 | `20260913T235317Z-p92582` |
| `make frontend-typecheck` | PASS 2/2 | `20260913T235317Z-p91924` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260913T235317Z-p92147` |
| `make lint-biome` | PASS 2/2 | `20260913T235317Z-p92075` |
| `make lint-scripts` | PASS 2/2 | `20260913T235317Z-p92144` |
| `make generate-drift` | PASS 4/4 | `20260913T235317Z-p91602` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260913T235317Z-p91693` |
| `make json-shape-check` | PASS 3/3 | `20260913T235317Z-p91769` |

The full accessibility result and the broad workbook/Grid Adapter results close
G11 and supersede its historical failed targets. Browser accounting is retained
under `browser-e2e-a11y/`, including production Playwright reports and
`frontend-accessibility-summary.json`. Visual product bytes remain those reviewed
and validated twice at `20260913T232900Z-p14781` and
`20260913T233320Z-p78044`; subsequent changes affected verification fixtures,
projection registration checks and documentation, not production rendering.

### G12 — Exact handoff documentation coverage

`make lint-markdown` passed at `20260913T235543Z-p97252`, summary
`adhoc/lint-markdown/tool-run-summary.json`. Final coverage inspection found that
its existing globs included source guides and selected handoffs but omitted this
new path. That pass therefore does not establish validation of this handoff.
The authorized final-byte requirement is controlling: remediation adds this
exact handoff path to `.markdownlint-cli2.jsonc`, without broadening product
checks or making any executable product artifact consume Markdown.

Affected area is documentation validation only. Rationale/benefit: future public
Markdown lint includes the controlling artifact and cannot silently omit it.
Compatibility/migration impact is none for application behavior, stored records,
public protocol or generated product files. Residual risk is a final edit after
lint; binary acceptance requires the complete DONE candidate bytes to pass the
public lint target, promotion of exactly those bytes, and a canonical-path lint
and diff check with no subsequent edit.

### ORC-05 terminal disposition

Validation candidate prepared from the complete handoff. The canonical tracker
remains IN_PROGRESS until `make lint-markdown` validates the identical complete
DONE candidate at repository root. Promotion uses byte-for-byte replacement,
followed by `make lint-markdown`, `git diff HEAD --check` and final scope review
on the canonical path. The final response supplies that last documentation run
root so no post-validation edit is needed merely to insert its own run identity.
The temporary candidate is removed by promotion and is not a delivered artifact.

Final scope review: `main` remains at
`83fd4165de4f8490235c5c5ad3b7564cedfe6e20`. The user-visible staged state found on
resumption is preserved; new repairs and goldens remain unstaged/untracked as
applicable. No index manipulation, commit, push, deployment, dependency or
lockfile change, SQL/data migration, digest edit or analyst-data modification
occurred. Product Grid Adapter and Timeline algorithms are unchanged; the only
Grid Adapter edit is its test-support control-value capture. All primary-owner
and applicable acceptance decisions are resolved. G1–G12 have closed binary
criteria; G12 is finalized by the candidate/canonical documentation procedure.
Next action after that verification is delivery of this handoff and stopping at
this seam.

ORC-05 exit: DONE. All applicable A001–A027 rows are PASS. The validated
complete bytes were promoted only after candidate validation; canonical-path
documentation and diff validation complete the delivery record. No applicable
BLOCKED row, unresolved owner decision or authorized implementation work remains.
