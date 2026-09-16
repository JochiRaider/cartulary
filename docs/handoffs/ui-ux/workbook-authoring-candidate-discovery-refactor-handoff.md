# Workbook authoring candidate discovery and selection continuity

## Baseline and execution control

Execution starts on clean `main` at
`5d19723bcf711da6478ae022e4d58c3854581e4f`. Branch, HEAD, dirty state and
applicable AGENTS.md were revalidated before editing. Only root AGENTS.md applies;
there is no pre-existing work to overwrite. The localized digest read order,
adopted owners, source guides and preceding reference, query, saved-view,
grid/autosave, clipboard, recovery and creation handoffs were inspected during
planning at this same HEAD. Historical evidence is not a current acceptance pass.

Authorized scope is all production consumers of useWorkbookCandidates, their
read/presentation boundaries, necessary parent selection metadata, tests and
source guides. The user also approved compact schema-declared ordering/filtering
for all record pickers and test-only repair of the 24 reproduced selector-policy
readiness waits. Membership discovery remains paged only. No digest edits,
dependencies, persistent browser storage, endpoints, commits, pushes, deployment
or analyst-data changes are authorized. Tests use isolated fixtures.

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| ACD-01 Characterization and contract reconciliation | DONE | Owner/consumer matrix, bounded policy and binary gap criteria recorded; diff check passes. |
| ACD-02 Bounded discovery ownership | DONE | Bounded controller, typed reads, semantic fencing and focused lifecycle evidence pass. |
| ACD-03 Consumer integration and retirement | DONE | All five families migrated; bounded staging, raw entry, source review and parent recovery evidence pass. |
| ACD-04 Integrated behavior and spreadsheet evidence | DONE | Production, accessibility, spreadsheet and all changed-selector rows pass; eleven reviewed goldens pass two fresh full visual runs. |
| ACD-05 Terminal validation and completed handoff | DONE | Terminal verification, completed-handoff Markdown, acceptance and scope review pass; ready for review. |

Workstreams advanced sequentially: only the active row was IN_PROGRESS, and its
actual exit evidence and DONE status were saved before its successor began. All
five rows are now DONE; no owner contradiction or applicable blocked acceptance
remains.

## Authority and verification boundaries

Core 01 §§3.3.4–3.3.7, 7.4, 18B and 19 own query, creation, reference and
replay contracts, including REQ-01-674's distinct coordination source input.
Core 02 §§3, 10, 12–13, 15 and 19 own identity, artifact variants, source
relationships, Assessments, Evidence, history and Parties. Core 03 §§2.3A,
3–4, 13 and 16 own authoring continuity, interaction, keyboard and contextual
workflows, especially REQ-03-100/219/256/258/299/304/305. Core 04 §§1–2 owns
authorization and protected content. Domain vocabulary and design §§7–8,12,14
apply only within their declared boundaries. No normative change is planned.

Source ownership is web.workbook; shared source stays within current frontend
import rules. Verification is independently routed by contracts/verification,
tools/test_catalog_owner.json and tools/test_families. No executable artifact
reads, stats or hashes Markdown. Source and test manifests are authored inputs;
generated derivatives are changed only through Make.

## Evidence log

Planning at the captured HEAD (baseline evidence only):

- `make help`, `make help-all`, and task guides for web.workbook,
  module.workbook and web.architecture passed.
- Focused Workbook discovery/authoring/reference slice: PASS 8/8 units,
  `.cartulary/test-results/20260916T030013Z-p1004722`.
- Selector policy row: FAIL with 24 existing heading-name readiness waits,
  `.cartulary/test-results/20260916T030041Z-p1006233`.
  Failure details are in its row's `unit-logs/.../vitest-failure-details.json`.
  This was reproduced at clean HEAD, rather than inherited from Saved Views.
- Execution revalidation: clean same HEAD; no nested AGENTS.md found.

## Acceptance and completion

Implementation, consumer integration and terminal code/browser verification are
complete. Current evidence is recorded per workstream below; historical planning
evidence is not used as a completion pass. ACD-01 through ACD-05 are DONE. Every
applicable acceptance row is PASS; N/A rows have explicit scope/owner rationale.

## ACD-01 consumer and field reconciliation

Production search confirms exactly five hook consumers. Shared ordinary reference
presentation is also used by contextual coordination source selection. Sources
below mean the existing authorized workbook query; members use the existing
incident-membership list. All query pages are 100, server ordered, opaque-cursor
pages. Record sources support their declared query capabilities; membership does
not support record filtering. Public direct references preserve exact identifiers.

| Consumer / entries / parent | Target fields | Candidate identity and source | Cardinality / clear / regression boundary |
| --- | --- | --- | --- |
| OrdinaryCreateControl; grid and supplementary inspector; WorkbookOrdinaryCreateOwner | task.owner_user_id, decision.owner_user_id, finding.owner_user_id, handoff.incoming_owner_user_id, handoff.outgoing_owner_user_id, status_review.review_owner_user_id, lesson.owner_user_id | user_id from incident_members | Single; omission/default differs from invalid null on non-clearable fields. Parent owns raw input and Commit; existing-edit registry is separate. |
| Same ordinary owner | task.requester_party_id, evidence.collector_party_id, evidence.source_party_id | party_id from Parties | Single; explicit null clears optional reference, preserved Party text is independent. |
| Same ordinary owner | task.decision_record_id | Decision record_id from Decisions | Single; explicit null clears, no implicit relationship beyond the field contract. |
| Same ordinary owner | task.linked_record_ids, decision.support_refs, decision.affected_record_ids, finding.supporting_refs, finding.contradictory_refs | record_id from discovered implemented record surfaces | Up to 64 create actions; empty omits the collection. No existing-record item removal semantics are imported. |
| Ordinary owner and CoordinationCreateForm / WorkbookCoordinationCreateOwner | comm_log.decision_ids, handoff.open_decision_ids, status_review.open_decision_ids | record_id from Decisions | Up to 64 additions; empty omits. Field names do not invent lifecycle filters. |
| Same ordinary/coordination owners | comm_log.action_task_ids, handoff.open_task_ids, status_review.blocked_task_ids, lesson.follow_up_task_ids | record_id from Task Requests | Up to 64 additions; empty omits. Draft and source context remain independent. |
| Same ordinary/coordination owners | status_review.pending_evidence_ids, lesson.evidence_refs | record_id from Evidence | Up to 64 additions; empty omits. Evidence lifecycle remains parent/domain owned. |
| Same ordinary/coordination owners | comm_log.audience_party_ids, comm_log.attendee_party_ids | party_id from Parties | Up to 64 Party additions; empty omits; required audience text remains independent. |
| CoordinationCreateForm / retained coordination owner; 12 declared contextual actions | coordination.source_record_id | record_id, schema and row version from REQ-01-674's target-specific source matrix | Single; explicit clear selects unlinked creation. Replacement preserves target fields and requires current source review. Origin is not the editable source. |
| ContextualCreateForm / WorkbookContextualCreateOwner; 23 Task/Decision actions | task.owner_user_id, decision.owner_user_id; task.requester_party_id; task.decision_record_id; task.linked_record_ids, decision.support_refs, decision.affected_record_ids | Members; Parties; Decisions; discovered record surfaces respectively | Singles or at most 64 collection actions. Empty owner means current-actor default; optional scalars/collections retain existing omission preparation. Seeds remain editable; source verification, attempts and receipts stay parent-owned. |
| NoteCreateForm and NoteSheetAuthoring / WorkbookNoteCreateOwner | Explicit Note source | record_id, schema and row version from Timeline, Hosts, Identities, Evidence | Single staged selection; clear creates an unlinked Note. Source alone does not meet Note minima. No Note-specific storage or extra link write. |
| TimelineRelatedEvidenceForm / WorkbookTimelineRelatedEvidenceOwner | evidence.collector_party_id, evidence.source_party_id | party_id from Parties | Separate singles; empty control preserves parent omission/clear preparation. Metadata text, source Timeline and separate create/link attempts are unaffected. |
| Assessment inspector / WorkbookAssessmentAuthoringOwner | assessment.subject_ref with subject_type | Host or Identity record_id and display text | Single deliberate selection updates the draft immediately. Explicit subject-type change clears the old subject; paging/filtering does not. Append semantics remain unchanged. |
| Same Assessment owner | assessment.support_refs | Timeline record_id and display text only, REQ-03-304 | Staged up to 64 additions; empty omits; Cancel changes nothing. Existing non-Timeline references stay readable. |

Core 01 field registries §§7.4/18B/19 and Core 02 §§10/19 back the field
mapping. The existing feature preparation owners already enforce collection
action limits; these are not inferred from WorkbookReferenceSelection's edit
limit. Shared selection will receive a feature-supplied maximum.

Read ports are WorkbookAuthoringReadPort (including Note and contextual aliases)
and AssessmentCandidateReadPort. createWorkbookAuthoringReader already invokes
authority rechecks, and createAssessmentCandidateReader already classifies
authority failures; neither protection is absent. boundedRead already enforces
a 30-second deadline. useWorkbookCandidates already cancels, rejects obsolete
keys, retains rows after continuation failure and prevents overlapping reads.
Parent owners already retain selected IDs, source high-water marks, drafts,
captured attempts and receipts. The reconciliation preserves these capabilities.

## Gap ledger

| Gap / classification | Remediation / affected area | Rationale and long-term benefit | Compatibility / risk if unresolved | Binary validation |
| --- | --- | --- | --- | --- |
| G1 observed structural accumulation | Replace accumulating hook with one-page discovery; bounded checkpoints and minimal selected metadata | Dataset growth no longer grows browser retention | Frontend-only cutover; unbounded payload/cursor/label growth remains otherwise | After 12+ pages, rows <=100, prior checkpoints <=10, labels limited to current selections |
| G2 observed presentation coupling | Separate accepted observations from pending/failure; migrate all five controls | A failed later read does not disable usable earlier observations or unrelated Apply | Apply still only edits drafts; otherwise read failure blocks deliberate authoring | Select/apply accepted candidates during continuation and its retryable failure |
| G3 observed typed-outcome erasure | Preserve WorkbookPortResult in discovery and surface inventory | Distinguish recovery from target/authority loss without a second authority owner | Internal port change; otherwise generic error text obscures required lifecycle | Typed outcomes survive to presentation; current authority owner receives scoped recheck |
| G4 observed oversized selection metadata | Retain selected identity/label and source rowVersion only; prune feature label maps | Selected records survive eviction without retaining visited rows | Parent preparation unchanged; otherwise selections keep unnecessary payloads and labels | Multi-page selected IDs survive; no unrelated label or row payload survives |
| G5 observed ordinary picker-only reference control | Restore exact raw-ID authoring alongside optional picker | Honors REQ-03-219 and keeps creation immediately editable | Existing parser/clear rules apply; otherwise reference entry requires browsing | Raw input edits parent authoring without lookup/creation or focus loss |
| G6 observed repeated query UI / missing controls | Compact shared schema-declared query presentation for all record pickers | Responsive bounded narrowing without feature query copies | User-approved UI extension; members stay paged only | Query edits dispatch only on Apply; selection and parent fields unchanged |
| G7 reproduced baseline test-policy defect | Replace 24 heading readiness waits with semantic readiness selectors | Restores architecture gate without weakening policy | Explicitly approved test-only expansion; gate otherwise fails | Existing policy and every changed behavioral owner row pass |
| H1 lifecycle/security/focus hypotheses | Characterize query/authority replacement, late completion, dismissal, invalid cursor, loss/restoration and source changes | Evidence distinguishes existing protections from actual defects | No unobserved defect claimed; races could expose stale targets or steal focus | All named controller, consumer and production browser scenarios pass |

Common decision: one current authorized observation plus bounded navigation and
exact read recovery. Feature-owned selection, eligibility, maximum count, clear
semantics, source review and write lifecycle are excluded from that controller.
A further authoring target supplies its reader, scope and narrow candidate shape;
it uses the same browser without Timeline branches or a workflow framework.

Accepted policy: one 100-row page, ten earlier request checkpoints, explicit
Previous/Next/First/Refresh/Retry, no prefetch, no dataset-wide label cache.
Dismissal releases browsing and unapplied staging; reopening starts fresh with
parent selections. Candidate membership never establishes reference existence.
No normative contradiction was found. No endpoint, storage or migration change
is needed. Existing edits, main query windows, Saved Views, mutation ownership,
clipboard and global recovery remain regression boundaries.

ACD-01 exit: source/owner inspection and the focused baseline tests cover all
inventoried consumers. The gaps above have explicit binary criteria; suspected
lifecycle races remain hypotheses pending focused reproduction. Only this
handoff changed. Next action: implement and test bounded discovery in ACD-02.

## ACD-02 implementation and exit evidence

WorkbookCandidateDiscovery owns one accepted page, bounded navigation and exact
read recovery. The thin useWorkbookCandidateDiscovery binding fences replacement
by semantic scope and controller generation. WorkbookCandidateReadPort carries
canonical query correlation and a current-scope predicate through adapters;
WorkbookCandidateBrowsing presents typed local read recovery without mutations.
Disabled reads cannot be activated through navigation methods. Disposal aborts
work and releases retained pages/checkpoints. Source-verification callers retain
the existing full-row authoring port; a separate narrow selection type excludes
rows. Feature integration and selection reduction are ACD-03 work.

Both authoring adapters now verify canonical query metadata through the established
query primitive. Inventory reads preserve WorkbookPortResult; current consumers
and fixtures were adjusted to the internal port change. Adapter authority
callbacks check cancellation and the caller's semantic scope before rechecking.
Retryable read errors no longer request an unrelated authority refresh. No public
API, server behavior or normative owner changed.

New source/test files are registered in frontend_source_ownership and the Workbook
verification family. Source guides describe the boundary. Make generation updated
the topology render index; no generated file was hand edited.

- `make generate`: PASS, `.cartulary/test-results/20260916T031539Z-p1013484`.
- `make format`: PASS 2/2, `.cartulary/test-results/20260916T032412Z-p1039511`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.authoring_candidate_discovery,web.workbook.regression.assessment_discovery`:
  PASS 3/3 units (13 tests), `.cartulary/test-results/20260916T032417Z-p1043721`.
- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260916T032423Z-p1044415`.
- `git diff --check`: PASS.

The seven new lifecycle tests establish 16-page finite retention, Previous/First,
usable pending/failed accepted pages, exact retry, duplicate admission protection,
invalid/cyclic page rejection, typed authority/target retirement, cancellation and
30-second deadlines. The adapter fixture now supplies the actual effective sort;
it previously supplied an empty canonical sort. During implementation a redundant
aborted-result branch failed type checking (032211 run), and that fixture's first
canonical-sort correction duplicated record_id (032257/032334/032403 runs). Both
were corrected; the successful current runs above supersede them. Initial new
catalog title ordering also failed generation (031526 run); ASCII ordering was
repaired before generation. These are implementation failures, not inherited
baseline classifications.

Exit criteria PASS. Remaining risk is production composition and feature-specific
selection continuity; next action is ACD-03 integration and retirement.

## ACD-03 integration and retirement evidence

All five production families now use WorkbookCandidateDiscovery through its thin
binding. WorkbookAuthoringReferencePicker shares the ordinary, coordination,
contextual, Note and related-Evidence staged presentation. Assessment subject and
support retain their distinct direct/staged update boundaries while using the same
query, page-selection and read-recovery presentation. No main-sheet, Saved View,
existing-edit registry, clipboard or mutation recovery implementation was changed.

Current-page membership and retained selection are separate. Explicit removal
handles off-page IDs; rejected over-limit choices leave existing selections intact.
Apply requires current authoring permission and selection cardinality, not complete
discovery. Record filters/order use declared schemas and existing query builders;
query editing makes no request until Apply. Source changes reset local query
controls without changing selected IDs. Membership skips record inventory/filter
controls. Note and coordination source controls alone capture rowVersion; other
selections retain identity, schema and necessary display text. Full source review
continues through the existing deadline-bounded readWorkbookAuthoringRecord.

The shell supplies account/session/incident/authority identity to active and
recovery attachments. Adapters delegate scoped authority rechecks once per read;
source-verification callers still use their existing adapter recheck. Target
replacement, detachment, dismissal and read-only transitions release discovery.
Freshness changes discard stale staged labels. Parent contextual, coordination
and Evidence label maps now retain only IDs referenced across their own fields.
Ordinary raw-ID edits invalidate only the changed field's optional presentation;
other fields, attempts and receipts remain intact. Raw creation controls remain
immediately editable and own grid focus. Explicit picker dismissal restores its
invoking control without a read-completion focus effect. Popover positioning uses
the established inherited-zoom calculation and viewport bounds.

Retirement dispositions:

- useWorkbookCandidates.ts: deleted after its final production and test consumer
  moved; accumulating rows/cursor sets and stale/loading selection gates retired.
- createContextualCreateReader.ts: deleted alias; supported callers import the
  shared authoring adapter directly.
- WorkbookRecordCandidatePicker: retained native presentation for authoring,
  supported Party linking and Timeline mention controls.
- readWorkbookAuthoringRecord: retained supported complete-row source verification;
  it is not a browsing cache or a new record-lookup capability.
- Source READMEs, frontend source inventory and Workbook test routing updated.
  Make generation updated derived topology accounting.

Current evidence:

- All eight discovery/consumer rows: PASS 9/9 units,
  `.cartulary/test-results/20260916T034952Z-p1109801`.
- Contextual authoring, controller, presentation and ordinary raw-input slices:
  PASS 5/5, `.cartulary/test-results/20260916T034814Z-p1103822`.
- Parent recovery and ordinary lifecycle/omission slice:
  `.cartulary/test-results/20260916T034553Z-p1097076` passed all eight recovery/
  lifecycle rows; contextual presentation assertions failed there and passed in
  the subsequent focused and all-consumer runs above.
- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260916T035033Z-p1111790`.
- `make format` and `make lint-biome`: PASS 2/2 each,
  `.cartulary/test-results/20260916T035113Z-p1112899` and
  `.cartulary/test-results/20260916T035118Z-p1117170`.
- `make generate`: PASS,
  `.cartulary/test-results/20260916T035121Z-p1117552`.
- `git diff --check`: PASS; production search finds no retired hook/alias imports.

The new production-control tests traverse thirteen 100-row pages, retain selections
through eviction, recover delayed/failed reads, enforce selection limits without
replacement, apply queries explicitly, cancel/reopen, fence authority callbacks,
conceal protected labels, exercise StrictMode cleanup, and preserve reviewed
source versions while reducing replacement metadata. The ordinary raw-input test
also proves independent fields and labels survive without any discovery read.

Migration failures were investigated in retained run artifacts: 033252/033438 and
034553 runs exposed old accumulating-option/status/name expectations; these were
replaced with retained-selection and current-page assertions. Assessment's query
test additionally used an incorrect field key (host.state instead of the declared
host.host_state); its Apply/read readiness now uses the actual supported field and
request count. Type runs 033121, 033951/034101 and 034959 caught internal test/typing
migration errors, now corrected. The 033636 type run overlapped alias retirement
and observed its removed file; subsequent complete type runs pass. Formatting
033856 exposed three redundant effect dependencies and two non-null assertions;
these were repaired. Lint 035051 required formatting the final test correction,
then passed. A diagnostic attempt to override BIOME_CHECK_FLAGS was rejected as an
unsupported Make input; no harness policy was changed.

Exit criteria PASS. Compatibility is internal TypeScript/read presentation only:
existing creation serialization, operations and authoritative validation remain
parent-owned. No endpoint, dependency, storage or persisted-data migration was
introduced. Remaining work is production browser, accessibility, visual and
spreadsheet evidence in ACD-04, followed by terminal checks in ACD-05.

## ACD-04 implementation and exit evidence

Three new production browser rows in authoring-candidates.spec.ts use isolated
service-backed fixtures with 105 candidates (106 including a reviewed seed).
They cover ordinary, contextual and related-Evidence Parties; reviewed Note
sources; and distinct Assessment subject/support rules. They exercise actual
adapters, continuation failure/exact retry, delayed source replacement,
filter Apply, off-page retention, explicit cancellation, focus restoration and
unchanged parent fields. All three PASS in
`.cartulary/test-results/20260916T040852Z-p1251296` (11/11 execution units).
The thirteen-page presentation test and sixteen-page controller test cover
checkpoint-budget traversal; service tests deliberately do not exhaust catalogs.

Production evidence exposed and corrected two implementation defects:

| Gap | Owner / remediation / affected area | Benefit / compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- |
| G8 reproduced transport classification defect | Core 01 authorized query recovery and Core 03 continuity; both candidate adapters now distinguish thrown transport reads from invalid accepted payloads | Exact Retry remains available after a network disconnect; internal-only change; otherwise retryable failures incorrectly demand restart | Real browser abort followed by byte-equivalent request retry passes; both adapter unit assertions pass |
| G9 reproduced raw-control overlap | Core 03 REQ-219 immediate raw entry and keyboard/pointer accessibility; OrdinaryCreateControl provides a relative input slot beside the picker button inside the dense cell | Both controls retain independent hit areas without changing row geometry; frontend-only; otherwise input intercepts picker pointer clicks | Ordinary visual scenario passes functional opening, keyboard a11y passes, raw input test passes |

The shared picker now focuses its explicit Cancel control on attachment, rather
than querying a potentially hidden ordering/filter control. No read-completion
focus effect was added. Deadline tests additionally fence authority callbacks
after a read has settled, independent of abort cooperation.

Investigation record (not completion evidence): 040430 production failures exposed
G8 plus a test using undeclared host.display_name filtering and a narrow screenshot
before scrolling the resized inspector. The test now uses declared host.host_state
and verifies reachable controls after scrolling. 035520/040003 runs exposed old
ordinary wrapper selectors; production composition now scopes the raw input's
semantic gridcell. 041011 visual run exposed G9 and intentional picker differences.
041041 unit run exposed nested feature headings in Recovery; scoped selectors now
identify its own focus anchor. 041039/041415 spreadsheet runs exposed a test that
attempted to discard without opening recovery and then clicked through the still
open recovery surface. It now uses the existing open/close controls. No Timeline
or global recovery implementation changed; no unsupported baseline classification
is assigned to these newly observed stale test assumptions.

The approved 24 baseline heading waits are replaced without changing selector
policy. The initial selector gate PASS is 035449 (2/2). The corrected affected unit
owner slices PASS at 041313 (module.collaboration), 041321 (module.networkflow),
041329 (module.workbook), 041339 (web.networkflow), and 041347 (web.workbook).
Complete run IDs remain in the retained run roots. Additional browser rows cover
the test-only Network Analysis changes, history, entity merge, and capture actions.

Visual review follows the visual maintenance and browser design readiness guides.
The seven changed images from 041011 show explicit pages, selected identities,
query disclosures and visible Cancel focus. Narrow controls wrap within the
inspector. The 390x640 Note and 1280x720 Assessment production attachments in
040852 show reachable controls and wrapped long labels. Those attachments are
embedded in playwright-report.json; diagnostic extracted copies are under
/tmp/acd-review. They are implementation evidence, not conformance claims.
At that review point, golden refresh and terminal visual validation were pending;
the completed refresh and two validation runs are recorded below.

Additional current ACD-04 evidence:

- 041428 production a11y passed contextual, Evidence, Note and ordinary cases;
  coordination resize readiness was corrected and PASS at
  `20260916T042010Z-p1828636` (11/11). Two animation frames settle the responsive
  host before deliberate test focus; product focus assertions remain unchanged.
- Assessment a11y PASS `20260916T041650Z-p1658008` (11/11). These cases cover named
  controls, keyboard focus, narrow layouts, zoom and text spacing as applicable.
- All twelve spreadsheet scenarios passed across 041039's eleven passing cases
  and `20260916T042041Z-p1865188`'s corrected rejection case (11/11 units). The
  latter uses explicit recovery opening/closing for both rejected and conflicting
  edits; no input, acceptance, clipboard or movement assertion was removed.
- Changed-selector browser owners PASS at `20260916T041126Z-p1351054` (Entities),
  `20260916T041222Z-p1382797` (Network Analysis),
  `20260916T041327Z-p1422637` (Timeline capture),
  `20260916T041450Z-p1546680` (collaboration a11y), and
  `20260916T041547Z-p1620668` (all eight history scenarios).
- The 041750 a11y run passed the Indicator lifecycle/observation rows and mutation
  lifecycle row. Indicator tests now await their asserted completed refresh before
  changing attachments; earlier 041421 intentionally detached their refresh and
  then incorrectly expected completion. Their existing exact status assertions
  remain. Coordination's final resize check passes in 042010 as above.
- `20260916T042204Z-p1903318` PASS 13/13 covers existing-edit reference recovery
  and coordination exact replay/refresh-only recovery. The existing-edit test now
  preserves already-selected current-page native options when adding a choice;
  its retained seed's opaque server order never guaranteed it was off-page.
  Coordination completion asserts its own recovery disappears and the independent
  inactive-view refresh debt remains visible, instead of incorrectly requiring
  global Recovery to disappear. No reference-edit or recovery owner changed.
- Evidence creation replay, partial linking and role-loss cases passed in 041840;
  its contextual replacement case was corrected to close recovery and reopen the
  source inspector after discarding, then PASS `20260916T042216Z-p1927651`.
- Contextual Task exact replay PASS `20260916T042303Z-p1997258` (11/11).
- Grid Adapter keyboard decision/anchor rows PASS
  `20260916T042315Z-p2021612` (3/3). Architecture selector, source-ownership and
  workbook-layout rows PASS `20260916T042321Z-p2029002` (4/4).
- Workbook service-backed Parties/coordination queries and registry-derived
  query/create coverage PASS `20260916T042328Z-p2029740` (3/3).

The complete ordinary visual run `20260916T041449Z-p1546314` reconciles all 252
active captures/goldens, zero missing, zero ambiguous mappings, zero orphans, and
all 29 registered fixtures. All functional assertions passed; comparison changes
are limited to new authoring browsing controls and ordinary raw-reference cells
or supplementary inputs visible in existing Evidence/Task/Decision scenes.
No viewport, zoom, mask, scroll normalization or capture scope was changed.
The Make-owned update target retains comparisons that already pass, so unaffected
goldens are not regenerated. The resulting changes required inspection and
two fresh ordinary validation runs; both are recorded in the exit evidence below.

- Assessment contextual Decision refresh and append recovery PASS
  `20260916T042353Z-p2049554` (13/13).
- Final combined discovery/consumer slice PASS
  `20260916T042544Z-p2081740` (10/10), including the real-adapter transport
  corrections, settled-deadline callback fence, and ordinary raw-control layout.

The resulting extension boundary requires an authorized reader, minimal candidate
projection, semantic target/source/query identity, authority context and freshness
revision. Feature code supplies selection cardinality, eligibility, clear rules and
parent Apply. Another target can compose the same hook/query/page-selection UI
without changing Timeline, mutation execution or the global recovery catalog.

### Visual refresh record

`make browser-e2e-visual-update` PASS 12/12 at
`.cartulary/test-results/20260916T042215Z-p1924633`. It promoted eleven explained
changes and retained the other 241 goldens byte-for-byte. All eleven promoted
images were inspected. Decision: accept the bounded candidate UI and raw-reference
input changes. The existing dense row geometry, typography, visible focus and
responsive inspector scrollports remain intact. No renderer pin, viewport,
zoom, masks, scroll normalization or screenshot scope changed. The following
identity mapping comes from that run's reconciliation artifact, independently of
behavioral ownership. `none` means an active nonregistry capture, not an orphan.

Paths below are under apps/web/e2e/workbook.visual.spec.ts-snapshots/.

| Golden | Catalog owner row | Stable fixture IDs |
| --- | --- | --- |
| contextual-decision-references-narrow-linux.png | module.workbook.visual.contextual_task_decision_creation | none; active capture |
| contextual-task-request-references-narrow-linux.png | module.workbook.visual.contextual_task_decision_creation | none; active capture |
| coordination-source-narrow-linux.png | module.workbook.visual.coordination_create_authoring_recovery | visual.fixture.contextual_coordination_creation |
| decision-supersession-accepted-linux.png | module.workbook.visual.decision_supersession_review_recovery | none; active capture |
| decision-supersession-review-linux.png | module.workbook.visual.decision_supersession_review_recovery | none; active capture |
| decision-supersession-review-narrow-linux.png | module.workbook.visual.decision_supersession_review_recovery | none; active capture |
| evidence-affordance-states-linux.png | module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4 | visual.fixture.evidence_affordance |
| linked-note-source-narrow-linux.png | module.workbook.visual.note_create_authoring_recovery | none; active capture |
| ordinary-reference-authoring-linux.png | module.workbook.visual.ordinary_create_authoring_recovery | none; active capture |
| record-relationships-task-requests-linux.png | module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc | visual.fixture.task_requests_or_decisions |
| timeline-related-evidence-party-narrow-linux.png | module.workbook.visual.timeline_related_evidence | none; active capture |

### Additional verification drift disposition

G10 is observed downstream test drift, distinct from a new product defect:
Core 03 recovery focus, explicit navigation and independent refresh ownership
require tests to use actual attachment controls. Affected areas are
contextual-create.spec.ts, contextual-coordination-create.spec.ts,
workbook-inspector-edit.spec.ts, timeline-grid-entry.spec.ts and the already
approved workbook.a11y.spec.ts. Remediation preserves all mutation, identity,
focus and persistence assertions while opening/closing real recovery controls,
waiting for completed refresh before detachment where completion is asserted,
settling responsive layout before deliberate focus, and retaining current-page
native selections when adding a choice. Benefit: deterministic evidence of the
adopted behavior, without production compatibility changes. Unresolved risk:
false regression failures or accidentally asserting global recovery completion
when an independent obligation remains. Binary validation: every affected exact
catalog row PASS; achieved in the current run roots listed above. This does not
weaken the selector policy or classify an untested historical baseline.

Final source review confirms Assessment subject-type replacement already keys the
subject picker in useAssessmentWorkbookInspectorComposition. Its existing parent
remount resets query controls and detaches old reads. Fresh draft admission already
requires explicit discard of an unfinished draft. These are retained protections,
not newly claimed defects or additional browsing owners.

ACD-04 exit PASS. Two fresh `make browser-e2e-visual` runs pass all 12/12 execution
units (47 browser scenarios, 252 captures):
`.cartulary/test-results/20260916T042703Z-p2083893` and
`.cartulary/test-results/20260916T042715Z-p2109950`. Both use independent run-owned
builds and isolated service fixtures against the same promoted manifest. Functional,
a11y, spreadsheet, recovery and visual evidence is current. There are no remaining
applicable ACD-04 blockers. Next action: ACD-05 finalization and terminal checks.


## ACD-05 terminal validation

Current task discovery was repeated with `make help`, `make help-all` and the
applicable module-author task guides. Routing started with web.workbook and
module.workbook and added the affected Assessment, Evidence, collaboration,
Entities, UI/Grid Adapter, architecture, application and Network Analysis rows.
This routing does not establish behavioral authority.

`make agent-finalize` PASS 1/1 at
`.cartulary/test-results/20260916T043207Z-p2153524`, before broader terminal checks. It also passes again after the final presentation
refinement and installation repair at `20260916T045339Z-p2431425`. Its `unit-artifacts/finalize-summary.json` reports generated outputs unchanged,
zero updated files, and no rollback needed. RESULTS_DIR was unset because this
work has no qualifying successful full warm-check root. Retained-run selection,
canonical evidence maintenance, scheduler timing/order maintenance and performance
evidence checks were therefore skipped (`results-dir-not-provided`). No unrelated
finalizer changes needed isolation.

### Terminal failure investigation and remediation

The first full `make frontend-unit` run failed 10 units (626/636 passed), root
`.cartulary/test-results/20260916T043249Z-p2157613`. Each failing unit's runner and
vitest-failure-details artifacts were inspected. A separate detached baseline
worktree at the captured HEAD ran the same public target with the same installed
toolchain/dependencies and unchanged tracked source. It failed 10 units
(624/634 passed) at `20260916T043945Z-p2242914`. Its complete evidence is retained
under `.cartulary/acd-baseline-control/20260916T043945Z-p2242914/`.
The baseline reproduces nine terminal failures plus the previously confirmed
selector-policy failure. The changed tree replaces the selector-policy failure
with one Assessment presentation assertion requiring migration. This is current
baseline evidence, not an inherited classification.

G11 is confirmed downstream verification drift, not a new production lifecycle
defect. Core 03 §§2.3A, 3–4 and 16 require explicit recovery attachment, semantic
focus restoration, independent reads and retained authoring selection; Core 01
§§3.3.4–3.3.7 supplies the declared query/schema boundary. Remediation and affected
areas:

- WorkbookShell.inspector.test.tsx removes an obsolete eager membership response
  from its positional mock; the actual create receipt and seeded payload remain
  asserted. Membership discovery is activated deliberately.
- WorkbookShell.assessments.test.tsx checks the retained subject identity before
  the candidate page arrives. It retains all append/default/support/create checks.
- WorkbookShell.history.test.tsx explicitly opens conflict recovery and checks its
  own heading/invoker focus; cancellation still preserves draft and request counts.
- WorkbookShell.actionSequencing.test.tsx and WorkbookShell.support.test.tsx open
  the appropriate retained operation after asserting grid continuity. Queued-write
  gating, no duplicate mutation and current committed identity assertions remain.
- Decision supersession and Indicator observation presentation fixtures compose
  the existing WorkbookRecoveryFixture and select their retained entry. They
  assert current owner status, exact attempts, detached draft retention and focus.
- App.landing.test.tsx's three route/access-loss cases correct the stable request
  count from five to four before loss and seven to six afterward, reflecting the
  existing absence of an eager incident-directory request on an incident route.
  Workbook is mocked there; no application or authorization implementation changed.
- workbook-inspector-edit.spec.ts takes current schema IDs from the existing
  view-contract facade. The unchanged E2E architecture policy passes.

Benefit: regression evidence now exercises the adopted production composition
and independent reads. Compatibility/migration impact: test fixtures only; no
application, recovery, domain or API behavior changes. Risk if unresolved: terminal
gates fail on stale attachment/request assumptions and cannot validate this seam.
Binary criterion: all ten affected exact owner rows and the full frontend unit
gate pass. Narrow corrected rows and the full 636/636-unit terminal rerun pass at
20260916T044509Z-p2331025.

Narrow post-correction evidence (roots under `.cartulary/test-results/`):

| Command / owner rows | Result | Run root |
| --- | --- | --- |
| test-slice web.workbook: decision_supersession_runtime, indicator_observations_presentation, mention_undo_current_committed_identity_ddf69011e7 | PASS 4/4 | 20260916T044440Z-p2328707 |
| test-slice module.collaboration: verify_same_field_conflict_anchors_conflict_queu_8bcc6e7d75 | PASS 2/2 | 20260916T044440Z-p2328716 |
| test-slice module.workbook: verify_sync_engine_pending_queue_orders_creates_8992eb8931 | PASS 2/2 | 20260916T044449Z-p2329921 |
| test-slice module.workbook: verify_inspector_panels_and_feature_groups_rende_b947f0007c | Row PASS in the initial corrected two-row run; the other row was corrected separately | 20260916T044227Z-p2298624 |
| test-slice module.assessments: the_assessment_surface_preserves_stable_selectio_ab289a0995 | PASS 2/2 | 20260916T044322Z-p2317727 |
| test-slice web.application: both affected app_landing rows | PASS 3/3 | 20260916T044253Z-p2307303 |
| test-slice harness.browser: architecturepolicy_suite_4e0bacb131 | PASS 2/2 | 20260916T044343Z-p2323428 |

The exact full row IDs and declared OWNER/ROWS arguments are in each root's
run-manifest.json; row outcomes are in unit-results and target-summaries.
Intermediate 044226/044227/044253 failures identified remaining stale recovery
names/focus targets. Those assertions were aligned with the existing semantic
heading and explicit invoker, then passed above. A terminal type failure at 043249
was an unused import left by an earlier test edit. Formatting at 044140 and lint
at 044252 exposed a test-only createElement children argument and unfinished
formatting; both were repaired. These are investigated implementation/fixture
corrections, not ignored failures or weakened checks.

| Terminal command | Result | Run root under .cartulary/test-results/ |
| --- | --- | --- |
| make frontend-unit | PASS 636/636 on final production source | 20260916T045407Z-p2435423 |
| make frontend-typecheck | PASS 2/2 after final browser assertion correction | 20260916T045809Z-p2565410 |
| make format | PASS 2/2 | 20260916T045055Z-p2404381 |
| make lint-biome | PASS 2/2 | 20260916T045833Z-p2576011 |
| make frontend-import-boundary-check | PASS 2/2 | 20260916T045507Z-p2481224 |
| make json-shape-check | PASS 3/3 | 20260916T045522Z-p2488714 |
| make generated-artifact-policy-check | PASS 3/3 | 20260916T045534Z-p2493359 |
| make generate-drift | PASS 4/4 | 20260916T045544Z-p2496857 |
| make lint-markdown | PASS on completed handoff body before DONE save | 20260916T050016Z-p2580281 |

G12 is an observed typed-message presentation gap found in final review:
reference-surface inventory concealed protected selections correctly but displayed
a generic authorization message. The candidate-page UI already distinguished
session loss from incident access uncertainty. Core 04 §§1–2 and Core 03
continuity require that distinction through presentation. The existing shared
browsing presenter now exports the same typed authorization wording for inventory.
Affected files: WorkbookCandidateBrowsing.tsx, WorkbookAuthoringReferencePicker.tsx
and its existing authority test. Long-term benefit: both discovery reads explain
the appropriate recovery without creating another authority owner. Compatibility:
error copy only; authorization callbacks, concealment and parent drafts are intact.
Unresolved risk: analysts cannot distinguish an expired session from an incident
access check. Binary validation: each inventory failure displays its specific
message, conceals labels, disables Apply, starts no candidate read and invokes the
existing authority callback once. The new assertions first failed at
`20260916T045018Z-p2403579`, reproducing the missing session message. After the
presentation correction, controller/picker slice PASS 3/3 at
`20260916T045100Z-p2408658`; formatting PASS at 045055. Terminal checks pass on the final production source at the run roots in the
terminal table. All three production composition scenarios pass again in
20260916T045407Z-p2435333; its independent existing-edit fixture race is documented
below and has its own corrected passing row.

The temporary baseline worktree was removed after its complete run artifacts were
copied to the retained baseline-control path. No source or dependency artifacts
from that worktree were copied into the implementation.

An execution-environment failure interrupted the final reruns at
`20260916T045151Z-p2410117` (agent-finalize, missing ajv) and
`20260916T045150Z-p2409900` (browser service setup, 6/13 units passed before the
frontend build failed). The baseline checkout's Make-owned installation had
redirected shared dependency symlinks into that temporary checkout. Its removal
made those installed links unavailable. This was caused by the baseline setup,
not the implementation. The temporary link target was restored, then
`CARTULARY_FORCE_REINSTALL=1 make frontend-install` repaired the workspace's frozen
installation at `20260916T045308Z-p2430631`. Inspection confirmed zero dependency
links remaining to the temporary path, and the temporary links were removed.
No package manifest, lockfile or generated source changed. Neither interrupted attempt counts as completion evidence. The repaired-install
finalizer passes at `20260916T045339Z-p2431425` (1/1), again with zero generated
changes and RESULTS_DIR unset. The final browser and frontend checks follow it.

The expanded final browser run `20260916T045407Z-p2435333` passed all three new
candidate-discovery scenarios and four of five existing-inspector scenarios. Its
reference-selection scenario reached its final second-write assertion and ended
while the intercepted response was still in flight (`route.fetch` reported a
closed page at teardown). This is an additional G10/G11 fixture readiness race:
counting a dispatched PATCH does not prove acknowledgement. The test now awaits
that exact PATCH response and the expected saved row version 3 before finishing;
all request payload and operation-count assertions remain. The corrected exact
row passes 11/11 execution units at `20260916T045741Z-p2538436`. No production
operation or recovery owner changed. Selector policy and E2E architecture policy
pass again at `20260916T045841Z-p2577691` and `20260916T045848Z-p2579136` (2/2 each).

### Acceptance assessment

| Applicable requirement / gap | Owner basis and binary evidence | Assessment |
| --- | --- | --- |
| Consumer/field reconciliation, G1–G6 | Core 01 creation/query/reference clauses; Core 02 identity/Parties/source/Assessment clauses; every inventoried field has the disposition above | PASS |
| Finite rows, checkpoints, retained metadata, G1/G4 | Core 01 bounded authorized query; sixteen-page controller and thirteen-page production-control tests prove <=100 rows, <=10 checkpoints and selected-only payloads | PASS |
| Responsive accepted observations and exact read recovery, G2/G3/G8 | Core 01 query and Core 03 continuity; delayed/failed continuation keeps accepted selections usable, exact retry repeats the read, invalid cursor needs explicit First | PASS |
| Scope replacement, deadlines, target/authority outcomes, H1 | Core 03 continuity and Core 04 §§1–2; obsolete success/failure/rechecks cannot publish; deletion/access/session outcomes remain distinct; protected labels conceal | PASS |
| Feature-owned staging and authoring, G4/G5/G6 | Core 03 REQ-219/256/258/299/304/305; every family passes explicit identity, clear, source review, selection-limit and draft/attempt/receipt continuity tests | PASS |
| Production adapters and composition | Three new service-backed rows plus affected authoring/recovery rows pass for all five families, with separate Assessment subject/support evidence | PASS |
| Main Timeline spreadsheet regression | Core 03 keyboard/creation owners; twelve scenarios and Grid Adapter evidence pass input, movement, accepted commit, caret, Escape, range/clipboard/fill and creation admission | PASS |
| Accessibility, focus and visual behavior, G9 | Core 03 interaction plus design direction; affected family a11y rows and two complete visual validation runs pass; eleven promoted images individually reviewed | PASS |
| Selector policy and corrected verification, G7/G10 | Unchanged architecture policy and affected unit/browser owners pass, including Network Analysis | PASS |
| Typed surface-inventory authorization messages, G12 | Core 04 §§1–2; red/green production-control assertions distinguish session/incident outcomes while preserving concealment and authority delegation | PASS |
| Terminal verification fixture drift, G11 | Nine failures reproduced at captured baseline; ten corrected exact owner rows and full 636-unit gate pass | PASS |
| Retirement and extension boundary | No retired hook/alias imports; supported native picker and full source verifier retained; source/import and verification manifests updated | PASS |
| Terminal documentation, generated drift and scope | Final handoff lint, generated policy/drift, git diff --check and 101-path scope review pass; captured branch/HEAD unchanged | PASS |
| New endpoint, persisted-data/dependency migration, browser storage | N/A: no such change; existing authorized query and membership capabilities suffice | N/A |
| Deployment, release claim, full backend/security/release suite | N/A: frontend/read-presentation and test changes only; required domain service slices and browser builds pass, no Core 05 release/conformance claim is made | N/A |
| Retained full warm-check/performance maintenance | N/A: no qualifying successful full warm-check root; RESULTS_DIR deliberately unset and finalizer records skip | N/A |

### Compatibility, limitations and rollback

All changes are frontend implementation, internal TypeScript read/selection
contracts, source guidance and verification artifacts. Parent mutation preparation,
transaction identity, receipts, uncertain replay and refresh debt remain owned by
the existing authoring owners. Full-row source verification remains available to
its supported callers. No public protocol, endpoint, persisted-data migration,
dependency, browser storage or analyst-data update is involved.

Browsing is deliberately a current page, not a complete catalog or stable count.
Previous can re-fetch only the ten retained checkpoints; First is the explicit way
back beyond that budget. Opaque server cursors may expire or be rejected; the UI
retains justified selections and offers explicit restart. Selected identities may
later become unavailable, so authoritative validation remains in the parent
submission/source-review path. This task adds no generic lookup, prefetch,
exhaustive traversal or secondary mutation/recovery owner.

Rollback is a review of this work's authored diff against
`5d19723bcf711da6478ae022e4d58c3854581e4f`: reverse only the listed changes, restore
the two retired source files and corresponding owner entries, remove the added
boundary/tests, and regenerate affected projections through Make. Restore only the
eleven listed visual goldens and their tool-owned manifest entries if reverting
the UI. There was no pre-existing work at capture; any subsequent user edits must
be preserved. Do not reset the branch, overwrite a working tree or revert analyst
data. No database or deployment rollback is necessary.

A future target supplies its authorized reader, minimal candidate projection,
semantic target/source identity, current authority and freshness revision. It owns
selection limits, eligibility, clear/Apply semantics and parent submission. The
same controller, React binding and browsing/query presentation then provide finite
navigation, cancellation and exact read recovery without Timeline branches or a
second browsing-state owner.

### Changed files

This inventory includes authored additions, modifications, deletions and Make-owned
derivatives. Generated browser batching/topology and visual manifest entries were
updated by their tooling; generated roots and lockfiles were not hand edited.

- `apps/web/e2e/authoring-candidates.spec.ts`
- `apps/web/e2e/contextual-coordination-create.spec.ts`
- `apps/web/e2e/contextual-create.spec.ts`
- `apps/web/e2e/history-recovery.spec.ts`
- `apps/web/e2e/merge-recovery.spec.ts`
- `apps/web/e2e/network-flow.spec.ts`
- `apps/web/e2e/timeline-grid-entry.spec.ts`
- `apps/web/e2e/timeline-workbook.spec.ts`
- `apps/web/e2e/workbook-inspector-edit.spec.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/contextual-decision-references-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/contextual-task-request-references-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/coordination-source-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/decision-supersession-accepted-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/decision-supersession-review-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/decision-supersession-review-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/evidence-affordance-states-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/linked-note-source-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/ordinary-reference-authoring-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/record-relationships-task-requests-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/timeline-related-evidence-party-narrow-linux.png`
- `apps/web/src/app/App.landing.test.tsx`
- `apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx`
- `apps/web/src/workbook/WorkbookShell.actionSequencing.test.tsx`
- `apps/web/src/workbook/WorkbookShell.assessments.test.tsx`
- `apps/web/src/workbook/WorkbookShell.collaboration.test.tsx`
- `apps/web/src/workbook/WorkbookShell.history.test.tsx`
- `apps/web/src/workbook/WorkbookShell.inspector.test.tsx`
- `apps/web/src/workbook/WorkbookShell.support.test.tsx`
- `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/adapters/README.md`
- `apps/web/src/workbook/adapters/createAssessmentCandidateReader.ts`
- `apps/web/src/workbook/adapters/createContextualCreateReader.ts`
- `apps/web/src/workbook/adapters/createWorkbookAuthoringReader.ts`
- `apps/web/src/workbook/components/README.md`
- `apps/web/src/workbook/components/WorkbookAuthoringReferenceControl.tsx`
- `apps/web/src/workbook/components/WorkbookAuthoringReferencePicker.test.tsx`
- `apps/web/src/workbook/components/WorkbookAuthoringReferencePicker.tsx`
- `apps/web/src/workbook/components/WorkbookCandidateBrowsing.tsx`
- `apps/web/src/workbook/components/WorkbookCandidateQueryControl.tsx`
- `apps/web/src/workbook/components/WorkbookCandidateSelection.tsx`
- `apps/web/src/workbook/features/assessments/AssessmentDiscovery.tsx`
- `apps/web/src/workbook/features/assessments/README.md`
- `apps/web/src/workbook/features/assessments/assessmentCandidatePort.ts`
- `apps/web/src/workbook/features/assessments/assessmentDiscovery.test.tsx`
- `apps/web/src/workbook/features/coordination/ContextualCreateForm.tsx`
- `apps/web/src/workbook/features/coordination/ContextualReferenceControl.tsx`
- `apps/web/src/workbook/features/coordination/CoordinationCreateForm.tsx`
- `apps/web/src/workbook/features/coordination/README.md`
- `apps/web/src/workbook/features/coordination/WorkbookContextualTaskDecisionCreateOwner.ts`
- `apps/web/src/workbook/features/coordination/WorkbookCoordinationCreateOwner.ts`
- `apps/web/src/workbook/features/coordination/contextualCreateAuthoring.test.tsx`
- `apps/web/src/workbook/features/coordination/contextualCreateDiscovery.test.ts`
- `apps/web/src/workbook/features/coordination/contextualCreateRecovery.test.tsx`
- `apps/web/src/workbook/features/coordination/coordinationCreateAuthoring.test.tsx`
- `apps/web/src/workbook/features/coordination/coordinationCreateRecovery.test.tsx`
- `apps/web/src/workbook/features/coordination/decisionSupersessionRuntime.test.ts`
- `apps/web/src/workbook/features/entities/WorkbookEntityMergeRecovery.test.tsx`
- `apps/web/src/workbook/features/evidence/README.md`
- `apps/web/src/workbook/features/evidence/RelatedEvidencePartyControl.tsx`
- `apps/web/src/workbook/features/evidence/TimelineRelatedEvidenceForm.tsx`
- `apps/web/src/workbook/features/evidence/WorkbookTimelineRelatedEvidenceOwner.ts`
- `apps/web/src/workbook/features/evidence/evidenceFileRecovery.test.ts`
- `apps/web/src/workbook/features/evidence/timelineRelatedEvidenceAuthoring.test.tsx`
- `apps/web/src/workbook/features/evidence/timelineRelatedEvidenceRecovery.test.tsx`
- `apps/web/src/workbook/features/indicators/observationPresentation.test.tsx`
- `apps/web/src/workbook/features/notes/NoteCreateForm.tsx`
- `apps/web/src/workbook/features/notes/NoteSheetAuthoring.tsx`
- `apps/web/src/workbook/features/notes/NoteSourceControl.tsx`
- `apps/web/src/workbook/features/notes/README.md`
- `apps/web/src/workbook/features/notes/noteCreateAuthoring.test.tsx`
- `apps/web/src/workbook/features/notes/noteCreateRecovery.test.tsx`
- `apps/web/src/workbook/features/ordinary/OrdinaryCreateControl.tsx`
- `apps/web/src/workbook/features/ordinary/README.md`
- `apps/web/src/workbook/features/ordinary/WorkbookOrdinaryCreateOwner.ts`
- `apps/web/src/workbook/features/ordinary/ordinaryCreateContract.ts`
- `apps/web/src/workbook/features/ordinary/ordinaryCreateControls.test.tsx`
- `apps/web/src/workbook/hooks/README.md`
- `apps/web/src/workbook/hooks/useWorkbookAuthoringInventory.ts`
- `apps/web/src/workbook/hooks/useWorkbookCandidateDiscovery.ts`
- `apps/web/src/workbook/hooks/useWorkbookCandidates.ts`
- `apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts`
- `apps/web/src/workbook/ports/README.md`
- `apps/web/src/workbook/ports/WorkbookAuthoringReadPort.ts`
- `apps/web/src/workbook/ports/WorkbookCandidateReadPort.ts`
- `apps/web/src/workbook/services/README.md`
- `apps/web/src/workbook/services/WorkbookCandidateDiscovery.test.ts`
- `apps/web/src/workbook/services/WorkbookCandidateDiscovery.ts`
- `apps/web/src/workbook/timeline/useTimelineCreateRelatedWorkflow.test.tsx`
- `apps/web/src/workbook/utils/README.md`
- `apps/web/src/workbook/utils/retainWorkbookReferenceLabels.ts`
- `apps/web/src/workbook/workbookRecoveryNavigation.test.tsx`
- `docs/handoffs/ui-ux/workbook-authoring-candidate-discovery-refactor-handoff.md`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/test_families/module.workbook.json`
- `tools/test_families/web.workbook.json`

### Exact representative verification commands

Commands below are the actual retained run selections, not a copied task inventory.
Other exact row selections are available in the cited run-manifest.json files.

Run `20260916T042544Z-p2081740`:

```sh
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.authoring_candidate_discovery,web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.assessment_discovery,web.workbook.regression.contextual_task_decision_authoring,web.workbook.regression.contextual_task_decision_discovery,web.workbook.regression.coordination_create_authoring,web.workbook.regression.note_create_authoring,web.workbook.regression.timeline_related_evidence_authoring,web.workbook.regression.ordinary_create_controls
```

Run `20260916T040852Z-p1251296`:

```sh
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.authoring_candidates_parties,module.workbook.browser.authoring_candidates_note_source,module.workbook.browser.authoring_candidates_assessments
```

Run `20260916T042328Z-p2029740`:

```sh
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.integration.parties_and_coordination_system_views_query_rout_e0808abf39,module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120
```

### Final scope and handoff checklist

- Branch remains main at the captured HEAD; no commit, push or deployment occurred.
- Scope review covers all 101 changed paths listed above. No intervening user work
  was overwritten. The digest, package manifests, lockfiles and generated code
  roots are unchanged. The temporary baseline checkout is gone; retained evidence
  remains under .cartulary/acd-baseline-control.
- Adopted ownership and field-specific compatibility decisions are explicit;
  no normative contradiction or amendment remains.
- G1–G12 remediations have passing binary evidence. H1 hypotheses have lifecycle
  evidence without claiming previously existing protections were absent.
- Every production consumer migrated; retirement and supported retained helpers
  have explicit dispositions. No candidate read enters global recovery routing.
- Controller, consumer, parent recovery, production browser, accessibility,
  spreadsheet, visual, domain query, source/import and architecture checks pass.
  Full frontend unit gate passes 636/636. Relevant final browser/selector
  corrections have their own subsequent passing owner rows.
- Eleven visual changes are explained and inspected, with two complete ordinary
  visual runs passing after promotion. Later G12 changes affect only error wording,
  preserve existing page wording and successful rendering, and require no golden
  update. The final real-browser discovery rows also pass.
- Terminal type, lint, JSON shape, generated policy/drift and finalizer pass.
  RESULTS_DIR remains unset; retained full warm-check/performance maintenance is
  explicitly skipped. No broad backend/release claim is made.
- Compatibility, limitations, rollback, exact representative commands, run roots
  and complete changed-file inventory are recorded. Final Markdown and diff/scope
  confirmation passed before saving the active workstream DONE.

ACD-05 exit PASS. `make lint-markdown` passed at
`.cartulary/test-results/20260916T050016Z-p2580281`; final `git diff --check` and
scope review passed. The handoff checklist is complete. The only validation
limitations are the explicitly scoped N/A/skipped checks above.

Next action: review the implementation and this completed handoff. No further
implementation, deployment, commit or analyst-data action is pending.
