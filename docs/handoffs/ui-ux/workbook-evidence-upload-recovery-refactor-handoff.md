# Workbook Evidence upload recovery refactor

## Baseline and authority

Execution started on `main` at
`bd59fe9660601c78906657dc5ac9f39544e81246`. The working tree was clean;
there were no pre-existing changes. The implementation plan and its three
clarifications were approved in the conversation. No commit, push, deployment,
digest edit, or analyst-data modification is authorized by this handoff.

Behavior follows Core 01 §3.3.8 and Timeline/Evidence view contracts, Core 02
§13, Core 03 §§2.3A, 3–4, 8 and spreadsheet interaction owners, and Core 04
§2.0A. Domain vocabulary and design direction stay within their declared
boundaries. The digest and completed workflow handoffs are advisory and
historical evidence. Source placement follows the frontend ownership/import
manifests; verification routing follows the current catalog and family inputs.
No executable artifact may depend on this document or other Markdown.

## Sequential workstreams

| Workstream | Dependency | Status | Exit |
| --- | --- | --- | --- |
| EUR-01 Characterization and owner reconciliation | None | DONE | Every family and recovery transition has an owner-backed disposition and passing focused characterization. |
| EUR-02 Retained file authoring and ownership | EUR-01 | DONE | Unfinished work and accepted receipts survive permitted detachment with explicit resource lifetimes. |
| EUR-03 Stage recovery and spreadsheet integration | EUR-02 | DONE | Uncertainty, detachment, late acceptance and failed refresh cannot duplicate or retarget effects. |
| EUR-04 Integrated behavioral and security evidence | EUR-03 | DONE | Every confirmed gap passes evidence at its owning boundary. |
| EUR-05 Final validation and handoff | EUR-04 | DONE | Applicable terminal checks and completed handoff validation pass. |

Only the current workstream may be `IN_PROGRESS`. Its actual evidence must be
appended and its `DONE` status saved before the dependent starts. An unresolved
owner contradiction or failing required boundary prevents advancement.

## Approved owner decisions

1. Uncertain byte transfer uses explicit finalization-first recovery through
   the existing attach or atomic create route. PUT is not replayed. Uncertain
   finalization retains the captured request; definitive incomplete-upload
   rejection may offer a fresh slot.
2. Multiple-file gestures start no upload, retain existing work, and report
   `Choose one file at a time.`
3. Source edits and presentation detachment invalidate undispatched Timeline
   linking readiness. Review resumes against the original source. Ordinary
   draft promotion preserves that identity.
4. File attachment does not determine custody. Existing custody is preserved;
   new file-backed Evidence defaults to `requested`. Subsequent custody changes
   remain explicit ordinary actions.

## EUR-01 inspection and gap register

Inspection confirmed the baseline and read the applicable root `AGENTS.md`.
Planning followed the localized digest read order and inspected the current
frontend guides, source/import manifests, owner clauses, adapters, retained
runtime, backend upload lease/finalization paths and test families.

| Gap | Remediation and affected areas | Rationale and long-term benefit | Compatibility and risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 Normalized hints compared with raw file metadata | Evidence slot validation, transport tests; Core 01 REQ-01-244 | Preserve server-approved normalization without reconstructing a stale contract. | No wire change; malformed/cross-incident responses must remain rejected. | Trimmed/empty hints and zero-byte contracts are accepted correctly; altered contract identities are rejected. |
| G2 Zero-byte files rejected locally | Existing-Evidence admission and tests; Core 01 REQ-01-243 | Apply the declared inclusive size range consistently. | Newly admits owner-valid empty files; server ceiling stays authoritative. | Zero bytes succeeds; configured overflow creates no slot. |
| G3 Attachment automatically PATCHes custody | Existing attachment and Timeline create adapters, lifecycle tests; Core 02 §13.1 | Separate byte availability from analyst custody intent. | Removes incidental automatic promotion; accepted records are preserved. | Attachment preserves lifecycle and makes no custody PATCH; omitted create lifecycle is requested. |
| G4 Chain-local file/attempt/receipt state | Retained feature owners, semantic stage ports, runtime integration, recovery tests and source guides; Core 03 §§3–4, 8 | Preserve accepted progress and exact replay across presentation changes. | Internal ports change; all actual callers must migrate together. | Lost responses and remount preserve the exact attempt and produce one effect. |
| G5 First-file-only admission | Picker/drop/paste admission and tests; approved Core 03 clarification | Prevent silent loss without adding bulk upload. | Multiple-file input now receives explicit refusal. | Two files create no slot and preserve existing work. |
| G6 Timeline transfer runs inside save work | Timeline retained owner and source coordination, browser continuity tests; Core 03 first-input and concurrency owners | Keep unrelated spreadsheet capture responsive and source identity stable. | Draft promotion must use the existing creation authority. | Typing continues during transfer and either completion order creates exactly one intended Timeline row. |
| G7 Backend finalization paths differ | Evidence finalization preflight, failure accounting and transaction tests; Core 02 §13 and Core 04 §2.0A | Keep atomic creation and existing attachment under the same blob invariants. | No planned schema migration; authorization denial must not consume failure budget. | Both routes reject incomplete transfer and enforce terminal integrity and non-terminal failure budgets without partial record effects. |

G1–G6 are confirmed by source inspection. G7 is confirmed by the added public
route characterization and corrected within EUR-01. No adopted-owner
contradiction was found. Completed ordinary, paste/bulk, inspector, related
metadata Evidence and workbook-remediation handoffs remain regression boundaries;
their historical intermediate failures are not current defects.

### Operation and recovery inventory

| Family and current entry points | Local stages | Durable stages and replay | Retained recovery disposition |
| --- | --- | --- | --- |
| Existing Evidence: grid/inspector `EvidenceAccessActions` picker through `useEvidenceWorkbookBindings` | File and selected Evidence/version | Slot `(actor, incident, txn)`; one PUT lease; record-scoped attach `(actor, record, txn)` | Preserve file and slot; exact slot/attach replay; uncertain transfer finalizes first; attachment preserves custody. |
| Existing Timeline: `TimelineEvidencePanel` picker/drop/clipboard and grid file drop/paste | File and original Timeline/version | Slot; one PUT; generic atomic Evidence create `(actor, incident, Evidence schema, txn)`; Timeline collection PATCH | Keep full Evidence receipt after link failure; source review and link-only recovery; no second create. |
| Timeline draft: `TimelineDraftRowActions` picker and grid file drop/paste | File and local draft identity without authoritative row | Slot; one PUT; atomic Evidence create; Timeline generic create with Evidence reference | Ordinary first input and screenshot completion share draft creation identity; promotion redirects only that same draft's continuation to its committed row. |

Grid text paste belongs to the Grid Adapter/Timeline clipboard owner. Grid
file gestures are admitted separately without consuming text paste. Existing
Evidence has picker presentations in the grid and inspector; it had no file
drop/clipboard presentation to migrate, and this seam adds none.
All file lists use the approved reject-together rule. Picker accept hints do not
establish security, type policy or preview eligibility.

Slot acceptance includes server normalization and independent 60-minute target
and 24-hour pending expiry. Slot replay after expiry returns the original target;
unused targets from a former issuing session cannot upload after reauthentication.
Successful transfer is not Evidence creation, attachment or custody. Finalization
requires completed transfer and atomically associates one blob with at most one
Evidence record. Failed finalization retains no placeholder record. Incomplete
lease preflight rejects without consuming a finalization attempt; admitted
object-store observation failures consume the existing budget. The fourth such
failure becomes terminal; successful idempotent replay consumes none.

Malformed success and transport uncertainty retain exact requests. A definitive
rejection permits only its owner-specific review, correction, fresh-slot or
transaction-conflict action. Selection, filtering and navigation detach
presentation without changing source identity; deletion blocks fresh association.
Accepted writes retain success through failed refresh and permit reads-only
reconciliation. Protected presentation and retained operation lifetimes follow
the current runtime authority boundaries.

### EUR-01 changes and evidence

- Added the approved scoped recovery/admission/source clauses to Core 03 §8.1.
- Corrected `createUploadedEvidenceBlob.ts` normalized hint comparison and
  removed zero-byte rejection and automatic custody PATCH from existing
  attachment. Timeline atomic create now omits implicit `available` custody.
- Corrected `initial_blob_create.go`, `mutation_create.go` and
  `route_blob_handlers.go`: incomplete transfer is rejected; actual observation
  failures use durable failure accounting and leave Evidence effects atomic.
- Extended `upload_lifecycle_integration_test.go` with real uploaded bytes,
  incomplete-lease rejection, consumed PUT rejection, and both public
  finalization failure budgets. Extended `workbookEvidence.test.ts` with
  normalized filenames, empty-file admission and absence of custody PATCH.
- Registered `module.evidence.frontend_unit.attachment_preserves_custody` in
  the authored Evidence test family.
- PASS: all four requested `make task-guide ROLE=module-author OWNER=...`
  commands.
- PASS: `make service-backed-test-slice OWNER=module.evidence` with rows
  `initial_blob_create_atomic_replay_concealment_73c81fe329`,
  `the_object_blob_create_route_enforces_request_sh_17b6e1b674`, and
  `expired_slot_replay_returns_the_same_expired_slo_6d29ac9f3a` (each prefixed
  `module.evidence.integration.`). Run root:
  `.cartulary/test-results/20260914T214709Z-p92979`; 3/3 units passed.
- PASS: `make test-slice OWNER=module.evidence` with frontend-unit rows
  `attachment_preserves_custody`,
  `private_object_blob_upload_hides_target_4b2f89cd31`, and
  `upload_retry_body_discard_697516c4d8` (each prefixed
  `module.evidence.frontend_unit.`). Run root:
  `.cartulary/test-results/20260914T214831Z-p11672`; 4/4 units passed.

These runs use isolated harness fixtures, not analyst data. Remaining risks are
frontend lifetime, source ordering and integrated security recovery; EUR-02–04
own their implementation and evidence.

## EUR-02 changes and evidence

- Added `EvidenceUploadSession` and `EvidenceFileFinalization` for private
  file/capability retention, captured stage requests and complete receipts.
  `WorkbookEvidenceAttachmentOwner` and `WorkbookTimelineFileOwner` own their
  distinct finalization/association policies. Public snapshots exclude files,
  capability targets and response bodies.
- Added semantic file/link transports and registered authored frontend paths.
  The incident runtime retains both owners, coordinates short source mutations,
  forwards authority retirement and accepted version observations, and records
  shared surface refresh debt before presentation effects.
- Unused old-session targets are released. Uncertain transfers retain files for
  explicit finalization-first recovery. Accepted finalization releases file and
  target references. Account/incident retirement invalidates late callbacks.
- PASS: focused `retained_file_recovery` row, six behavioral tests;
  `.cartulary/test-results/20260914T221021Z-p20331` (2/2 units).
- PASS: `make frontend-typecheck`,
  `.cartulary/test-results/20260914T221142Z-p21264` (2/2).
- PASS: `make frontend-import-boundary-check`,
  `.cartulary/test-results/20260914T221153Z-p21787` (2/2).
- Earlier implementation failures were corrected: optional-property/payload
  validation typing (`20260914T215730Z-p14668`), collection cell access typing
  (`20260914T220223Z-p16383`), and slot route capture
  (`20260914T220516Z-p18200`). Catalog title declaration/order errors prevented
  execution and were corrected in the authored test-family row.
- Remaining integration risk: controls and draft promotion still use the old
  adapters until EUR-03 migrates all consumers. No claim of reload persistence.

## EUR-03 characterization and resolved dependency

The first real-service selected-row and screenshot-only browser run confirmed
successful Evidence and Timeline associations, but failed both historical count
assertions: `module.evidence.browser` selected rows in
`.cartulary/test-results/20260914T222355Z-p28160`.
The production Timeline projection counts only available blobs with
available/released Evidence custody. New upload creation correctly remains
requested. This uncovered a material count-policy question; the adopted count
clauses do not explicitly settle that custody filter.

G8: proposed remediation is to count successfully finalized associated files
independently of custody, leaving metadata-only/incomplete-upload behavior alone.
Affected areas are the Timeline projection owner clarification, collection-facts
projection, focused unit/service/browser assertions and this handoff. Benefit:
attachment feedback and projected file counts agree without implicit custody
changes. Compatibility: existing requested/received/quarantined-custody records
with available blobs may gain a count; no data migration. Unresolved risk:
custody-filtered counts may be intentional. Binary validation: the approved
count rule holds after atomic create and existing-row association while pending
blobs never count. **Owner decision approved in the conversation:** count finalized associated
files regardless of custody. Core 01 and the Timeline collection-facts projection
now carry this rule. No unresolved count-policy dependency remains; focused
projection and browser validation is required before EUR-03 exits.

### EUR-03 completed implementation and evidence

- Migrated every picker and Timeline panel/drop/clipboard consumer to the
  retained owners. Added file gestures to the grid work area without intercepting
  clipboard text. Recovery controls identify the original source and expose
  explicit review, resume, fresh-slot, request-conflict, discard and refresh
  actions. Rejected multi-file admission is visible with the inspector closed.
- Captured slot, finalization/create and collection-link attempts before dispatch;
  exact uncertainty recovery retains complete receipts independently of refresh.
  A replay denial cannot disprove an earlier uncertain acceptance. Discard stops
  future stages while dispatched uncertain work retains its outcome owner.
- Long transfers no longer occupy Timeline's save queue. Original draft keys
  resolve through the retained Timeline creation owner, sharing screenshot and
  ordinary-input creation while preserving follow-on edits and presentation
  detachment. Source coordination and accepted version floors precede linking.
- Removed `createUploadedEvidenceBlob.ts`, `createEvidenceAttachmentPort.ts`,
  `createTimelineEvidenceAttachmentAdapter.ts` and
  `TimelineEvidenceAttachmentPort.ts` after migrating their callers. Removed
  ID-only attachment ports and the automatic custody PATCH.
- Rechecked completed upload leases inside existing-Evidence finalization and
  concurrent exact replay under the record lock. Extended attachment fixtures
  with actual completed lease state. Applied the approved G8 projection rule
  and custody/blob-state count matrix.
- PASS: `make service-backed-test-slice OWNER=module.evidence` for the attached
  create/patch and blob-attach store rows plus both selected-row and
  screenshot-only browser rows, 13/13 units;
  `.cartulary/test-results/20260914T223117Z-p64170`.
- PASS: the real existing-Evidence attach browser row
  `module.evidence.browser.verify_attach_flow_uses_generated_protocol_types_7a778c9178`,
  11/11 units; `.cartulary/test-results/20260914T223330Z-p97683`.
- PASS: `make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.workbookshell_surfaces_suite_668e482b1e`,
  2/2 units; `.cartulary/test-results/20260914T223619Z-p31713`.
- PASS: retained file and Timeline draft recovery rows, 11 behavioral tests,
  3/3 units; `.cartulary/test-results/20260914T224502Z-p50441`.
- PASS: `make frontend-typecheck`, 2/2;
  `.cartulary/test-results/20260914T224502Z-p50501`.
- PASS: `make format`, 2/2;
  `.cartulary/test-results/20260914T224442Z-p46036`.
- Corrected intermediate failures: outdated shell receipt version fixture
  (`223331Z-p97905`), missing retained draft reader/refresh fixture
  (`223455Z-p30566`), fixture receipt types/private refresh access and unused
  import (`224217Z-p42638`), semantic markup and dependency/non-null lint
  (`224326Z-p45048`; earlier format runs `223735Z-p32876` and
  `224155Z-p38276`). A diagnostic invocation with `VERBOSE` as a Make input
  was rejected by public preflight; normal public invocations remain in use.

All EUR-03 exit checks above passed before EUR-04 began. Integrated failure
injection and broader security/visual regressions remain EUR-04 obligations.

## EUR-04 integrated evidence

Additional confirmed gaps and their closed remediation criteria:

| Gap | Remediation and affected areas | Rationale and long-term benefit | Compatibility and remaining risk | Binary validation |
| --- | --- | --- | --- | --- |
| G9 Accepted link replay skipped refresh | Timeline file owner continuation and production browser recovery test; Core 03 acknowledgement/refresh owners | Keep accepted association and projection reconciliation independent. | No API change; failed reads retain debt. | Lost link acknowledgement followed by failed query exposes Refresh; recovery adds zero writes. |
| G10 Narrow inspector made grid recovery unreachable | Shared retained-owner context and local inspector recovery presentation; Evidence/Timeline components and accessibility fixture | Reach the same operation from the active protected presentation without another state authority. | Same commands and record identity; grid stays mounted. | Keyboard controls remain reachable at 1280, 768, 390 pixels and 200% zoom; discard does not delete accepted data. |
| G11 Concurrent atomic exact create returned rejection | Evidence create association-rejection reconciliation and real API/object-store concurrency test; Core 01 operation idempotency and Core 02 atomic association | Return the already accepted complete receipt after waiting for a competing identical create. | No new port, route or schema; current non-association denials remain denials. | Two exact concurrent requests return one record/change-set receipt; distinct requests associate the blob once. |
| G12 Invalid accepted upload target could strand a ready stage | Slot receipt validation, neutral same-origin target validation and transport tests; Core 01 opaque capability and Core 04 concealment | Keep malformed success uncertain and exactly recoverable before byte dispatch. | Existing target method/headers remain authoritative; no URL reconstruction. | Foreign-origin, traversal and malformed targets never dispatch bytes or expose private bodies. |
| G13 Old negative callbacks could overwrite a newer observation | Slot, finalization and link dispatch fencing; retained recovery tests; Core 03 late-outcome ownership | Accept late correlated success while rejecting obsolete negative state changes. | Exact request identity remains unchanged; uncertainty never becomes proof of rollback. | An old rejection cannot reset an in-flight replay; late acceptance is retained once and clears pending state even if another observation times out. |
| G14 Definitive operation denial needed reviewed recovery | Feature-owned authority recheck and explicit original-source confirmation, retained-owner tests; Core 04 failure scope | Resume permitted work after authoritative recovery without inventing incident revocation. | Fresh rejected operations use new identities; uncertain operations retain exact identities. | Unknown authority dispatches nothing; current permitted authority plus review resumes the original source. |
| G15 Legacy fill verification expected retired focus behavior | Existing keyboard scenario and its authored Timeline routing title; Core 03 continuity and completed bulk remediation | Keep verification aligned with the current endpoint-preserving implementation. | No grid product change; no reinstatement of unconditional source focus. | Fill preserves the selected range endpoint after acceptance and refresh. |

G1 was further tightened during integrated review: the slot adapter retains
schema-validated server-normalized advisory hints instead of reconstructing
them with JavaScript `trim()`. Server and browser Unicode whitespace rules can
differ. Exact incident, byte-size, expected-hash, expiry and capability checks
remain in place. The U+0085 filename normalization regression passes with the
private upload/body rows in `20260914T234211Z-p31068` (3/3 units).

Source/version observations invalidate review only for the original matching
source. Retained status appears locally in grid and inspector; these snapshots
use `aria-live="off"`, leaving live save acknowledgement with the existing
runtime owner. Source review candidates clear on detachment and authority
changes. No accepted upload or association promotes Evidence custody.

### Boundary evidence

All run roots below are under `.cartulary/test-results/`. Run manifests retain
exact row selections and command inputs; `run-summary.json`, per-unit logs and
browser group reports retain results.

| Public command and selected boundary | Result and run root |
| --- | --- |
| `make test-slice OWNER=module.evidence ROWS=...` retained file, draft promotion, custody, private upload and body-discard rows | PASS 6/6; `20260914T231907Z-p27373`. Includes 12 retained workflow tests and both actual Timeline creation-owner orderings. |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell_surfaces_suite_668e482b1e` | PASS 2/2; `20260914T231907Z-p27395`. |
| Existing Evidence access binding regression row | PASS in `20260914T230746Z-p22643`; unrelated shell fixture failure in that combined run was subsequently corrected. |
| Atomic create, incomplete lease, consumed transfer, both failure budgets, terminal size/hash mismatch and concurrent distinct/exact association | PASS in `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.integration.initial_blob_create_atomic_replay_concealment_73c81fe329,...`; `20260914T231358Z-p7922`, 12/12 with original-source browser recovery. |
| Existing attachment store, transactional replay and quarantine eligibility | PASS 4/4 focused units; `20260914T225506Z-p9054`. |
| Expired target replay, fresh slot identity, completed bytes after target expiry and pending timeout | PASS integration unit in `20260914T230700Z-p89327`. |
| New real-service `file_stage_recovery`, `existing_file_recovery`, `file_draft_orderings` browser rows | PASS in `20260914T230019Z-p72389`; `file_original_source_review` PASS in `20260914T231358Z-p7922`. |
| Evidence slot authorization/request-shape, protected-handle redemption, stateful revocation/closed attachment and Evidence icon accessibility | PASS selected groups in `20260914T230044Z-p3305`; only stale visual setup failed that combined run. |
| New `module.evidence.accessibility.file_recovery` | PASS in `20260914T230700Z-p89327`; keyboard, focus, contrast, reduced motion, narrow layouts and 200% zoom. |
| Existing Evidence count, blocked access and requested access visual rows | PASS in `20260914T230700Z-p89327`, with unchanged golden bytes. Fixtures explicitly request any custody change and dismiss completed local recovery before capturing their declared historical states. |
| Workbook ordinary uncertain creation, inspector exact replay/detached refresh, related metadata Evidence, grid and inspector accessibility | PASS selected groups in `20260914T231358Z-p7962`; default shell visual setup was the only failing group. |
| `make test-slice OWNER=module.timeline` | 67/69 units passed in `20260914T230816Z-p45396`; sole failing browser scenario was the obsolete fill assertion (plus its group summary). Projection, keyboard, paste, grouping, virtualization and measurement rows passed. |
| Corrected fill endpoint scenario | PASS 11/11; `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser_support.timeline_keyboard_fill_down_preserves_the_select_66fb142159`, `20260914T231907Z-p27457`. |
| Timeline typing acknowledgement and blank-row creation measurement rows | PASS 16/16; `20260914T230043Z-p3086`. 100 qualified samples each: p95 30.8 ms against 100 ms; p95 87.5 ms against 150 ms. |
| `make frontend-import-boundary-check` | PASS 2/2; `20260914T230815Z-p45053`. |
| `make generate` and `make format` | PASS `20260914T231832Z-p19454` and `20260914T231844Z-p22886`. |

The browser stage tests transfer actual bytes through the public API and object
store. They deliberately drop acknowledgements after service acceptance and
assert exact request bytes and durable counts. Original-source conflict testing
injects a typed link rejection after a real independent source edit; transfer,
atomic Evidence creation and the subsequent link use real services. It does not
claim that an unrelated source-field edit necessarily conflicts server-side.

Manually reviewed the recovery screenshots retained as PNG attachments in the
accessibility `playwright-report.json` for `20260914T230700Z-p89327`: local
controls wrap inside the 390-pixel drawer, focus is visible, and desktop grid
and inspector retain independent scrolling. Contrast evidence is in that group's
`contrast-checks/` directory. These are implementation-support artifacts, not
Core 05 publication claims. No golden had been regenerated at that checkpoint;
the later reviewed refresh is recorded below.

Latest focused additions: retained timeout/replay observations PASS 2/2 in
`20260914T232857Z-p7645`; frontend typecheck PASS 2/2 in
`20260914T232252Z-p94881`; long-filename recovery accessibility PASS 11/11 in
`20260914T233433Z-p54019`. The 390-pixel long-filename capture was manually
reviewed: text wraps without hiding Resume/Discard or the status strip.

### Intermediate failures and disposition

- `20260914T225111Z-p66009`: real accepted-link replay missed refresh (G9);
  hidden picker scrolling and source-rebase assumptions were test setup issues.
- `20260914T225545Z-p26906` and `20260914T230019Z-p72389`: malformed synthetic
  rejection omitted required error status; narrow inspector needed G10; clock
  advancement needed a still-valid isolated test session.
- `20260914T230700Z-p89327`: source recovery locator matched both presentations;
  corrected to the exact semantic group. Accessibility, expiry and visual groups
  passed independently.
- `20260914T230745Z-p21967`: G11 was a confirmed service race, now fixed and passed.
- `20260914T230746Z-p22643` and `20260914T231358Z-p7931`: shell assertion needed
  a scoped group and its `within` import after adding inspector presentation.
- `20260914T230044Z-p3305`: visual fixture retained completed recovery and assumed
  automatic custody promotion; declared fixture state is now explicit.
- Full-shell visual investigation: `20260914T231358Z-p7962`,
  `20260914T231907Z-p27426`, `20260914T232251Z-p93413`,
  `20260914T232544Z-p31195` and `20260914T232942Z-p12794` exposed a fixture
  dependency on the removed completion-time Summary-cell selection. Attempts
  to normalize scrolling or alter wrapper clipping did not resolve it and were
  removed. Native `.focus()` alone does not establish semantic grid selection.
  The fixture now explicitly selects and exits editing on its declared cell
  before existing scroll preparation, positions the pointer outside controls,
  and prepares the registry-declared top-left scroll for each narrower capture.
  An isolated original-HEAD worktree passed the unmodified visual row
  (`20260914T232729Z-p67511`) and diagnostic repetition
  (`20260914T232942Z-p12806`); preserved under
  `.cartulary/eur-baseline-control/`. The temporary worktree and its borrowed
  dependency symlinks were removed after preserving these results. The diagnostic confirmed identical cell,
  grid and ancestor dimensions, but old selected `tabindex=0` versus an
  unselected `tabindex=-1` in the new flow. Diagnostic runs `20260914T233124Z-p81313`
  and `20260914T233340Z-p18366` narrowed the remaining differences to scroll
  preparation and nine raster pixels. No product focus-stealing path or
  speculative clipping fix was reinstated; no golden changed during diagnosis.
  Final paired controls `20260914T234103Z-p66809` and
  `20260914T234317Z-p36447` also passed and are preserved under the same
  baseline-control directory. Both temporary checkouts were removed; no
  diagnostic attachment code remains in the authored visual scenario.
- `20260914T230816Z-p45396` and `20260914T231425Z-p67123`: G15 was a stale current
  verification projection. The completed bulk handoff explicitly preserves the
  endpoint; no historical product defect was reopened.
- Typecheck failures `20260914T232029Z-p91586` (Testing Library uses exact role
  names without Playwright's `exact` option), `20260914T225506Z-p9111`, `20260914T230018Z-p72086`, and
  `20260914T230700Z-p89383` were corrected test types/imports. Catalog generation
  preflight failures were resolved through authored routing and `make generate`.

- Service regression `20260914T235133Z-p99946` passed 3/4 units; its upload-to-
  Timeline projection scenario still expected requested-custody files to count
  zero after finalization. The scenario now asserts requested custody is
  preserved and count/has-evidence becomes 1/true on finalization and replay,
  before the separate explicit custody PATCH. This aligns verification with
  the approved G8 count clarification; no product count behavior was reverted.
- Broad regression run `20260914T234958Z-p10301` passed 11/13 units; its
  stage-recovery selector still matched both grid and inspector. All recovery
  locators now use their exact semantic presentation name. The complete five-row
  upload/accessibility selection subsequently passed 13/13 units in
  `20260914T235313Z-p81559`, including the current long-filename and
  nonduplicated live-region assertions.
- `make frontend-unit` in `20260914T234958Z-p10411` exposed text-paste fixtures
  whose synthetic clipboard omits `files`. The file capture handler now treats
  an absent/empty file list as ordinary text paste. Although the harness
  classified the unhandled callback as preflight infrastructure failure, it
  was caused by this seam and was repaired here; passing follow-up is required.
  The run finished 618/620; both failures came from that same callback.

### Reviewed visual refresh decision

The remaining ordinary comparison in `20260914T234317Z-p36456` differs by nine
high-contrast pixels around the bottom draft controls. Baseline control
`20260914T234317Z-p36447` passes. Both builds report identical button coordinates
`(56, 847.796875)`, size `22.390625` square, parent geometry, Inter 14.4px/700,
font features, line height, and transforms. The glyph differs by one raster
row after the changed upload/recovery and explicit-selection sequence. The
other two responsive shell captures pass after their declared top-left framing
is prepared. This is a reviewed stale implementation-support golden, not a
product geometry or typography change. No tolerance or mask is loosened.

Accepted trigger under the visual maintenance guide: the upload interaction
and its fixture preparation intentionally changed; the old golden is stale
relative to the validated renderer state. Decision: accept the current measured
layout and refresh through `make browser-e2e-visual-update` only. The exact owner
row is `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0`;
stable fixture is `visual.fixture.default_timeline_workbook_shell`; affected
candidate is `apps/web/e2e/workbook.visual.spec.ts-snapshots/incident-directory-default-timeline-workbook-shell-linux.png`.
Viewport 1440x900, zoom 100%, dark_graphite, compact density, masks, scope and
renderer/font pins remain unchanged. Scroll and selection preparation are now
explicit, and the pointer is outside controls before anchoring. The ordinary
reconciliation artifact was reviewed before mutation. Fresh update results and
two ordinary validation roots must be recorded before accepting the refresh.
The selected-row reconciliation accounts for three active captures and 249
other committed PNGs outside that selection, with no missing/ambiguous mappings
or unresolved registered fixtures. The complete ordinary run `20260914T234958Z-p10467` then reconciled all
252 active captures/committed PNGs, all 29 registered fixtures, zero orphans,
zero missing/ambiguous mappings and zero unresolved registered fixtures.
Its 46 passing scenarios and one shell comparison failure confirm the same
affected golden. In the full catalog sequence the comparison differs by two
strong pixels on the bottom draft control, within the already reviewed glyph
area; current full image review confirms unchanged layout and typography.
The update began only after reviewing this complete reconciliation.

Refresh result: `make browser-e2e-visual-update` PASS 12/12 in
`20260914T235858Z-p51337`. All 47 scenarios and 252 captures passed; its
`browser-e2e-visual-update/frontend-visual-reconciliation.json` reports all 252
active PNGs, all 29 fixtures, zero missing/ambiguous/orphan mappings and no
errors. The target promoted only the default-shell PNG above and its one
SHA-256 entry in `tools/frontend_visual_golden_manifest.json`. Reviewed the
promoted full image and its old/new pixel bounds: `(56, 863)` through
`(100, 870)`, with two strong differing pixels. No layout, clipping, overflow,
focus, density or typography change is present. No other golden changed.
Fresh ordinary validations subsequently passed; terminal roots are recorded in
EUR-05 below.

Finalization preflight: `make agent-finalize` PASS 1/1 in
`20260914T234930Z-p6510`, before broader terminal checks. `RESULTS_DIR` was
unset; retained successful-run maintenance was skipped.

### Current-source integrated regression results

The current measurement selection (`20260914T235134Z-p909`) passed 19/21
units. Its four predicates each collected 100 qualified samples: typing
acknowledgement p95 48.4 ms, Enter focus 28.8 ms, and ArrowDown selection
62.4 ms passed their 100 ms limits. Blank-row creation p95 349.7 ms exceeded
150 ms while full visual/frontend/other service verification ran concurrently.
This is retained as a real failed measurement, not relabeled as a harness pass;
an isolated rerun without other heavy suites is required. No budget or
measurement algorithm changed.

- Full `make frontend-unit`: PASS 620/620 in
  `20260914T235456Z-p58725`, including both corrected text-paste callback rows.

- `make service-backed-test-slice OWNER=module.evidence ROWS=...`: PASS 4/4
  in `20260914T235457Z-p59084`; seven authored upload, projection, attachment,
  authorization, expiry, atomic-create and concurrent-association rows.
- `make test-slice OWNER=module.evidence ROWS=...`: PASS 3/3 in
  `20260914T235314Z-p82036`; seven Evidence unit rows for canonical mapping,
  creation effects, default lifecycle, derived fields, attachment/access,
  failure policy and mutation admission.
- `make test-slice OWNER=module.workbook ROWS=...`: PASS 4/4 in
  `20260914T235326Z-p89812`; public mutation schema, closed-incident admission,
  Evidence generic creation/patching and collection mutation integration.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell_surfaces_suite_668e482b1e`:
  PASS 2/2 in `20260914T235316Z-p82387`.
- `make test-slice OWNER=package.ui`: PASS 10/10 in
  `20260914T235131Z-p99391`; supported theme, tokens, selector and interaction
  contracts. `make test-slice OWNER=harness.test_catalog`: PASS 1/1 in
  `20260914T235109Z-p89126`.
- Current source `make frontend-typecheck`, `make frontend-import-boundary-check`
  and `make lint-biome`: PASS 2/2 each in `20260914T235548Z-p88296`,
  `20260914T235548Z-p88303`, and `20260914T235548Z-p88320`.
- Intermediate `make lint-markdown`: PASS in
  `20260914T235754Z-p28940`, with summary
  `adhoc/lint-markdown/tool-run-summary.json`. Final completed-handoff lint
  remains required.
- `make generate-drift`: PASS 4/4 in `20260914T235535Z-p80200`;
  `make generated-artifact-policy-check`: PASS 3/3 in
  `20260914T235535Z-p80215`; `make json-shape-check`: PASS 3/3 in
  `20260914T235535Z-p80219`; `make test-catalog-check` exited zero in the same
  sequential invocation. These checks will be revalidated after golden
  promotion where their inputs change.

## Verification and next action

Planning task guides passed for `module.evidence`, `module.timeline`,
`web.workbook`, and `module.workbook`. These are routing observations, not fresh
product passes. Execution evidence is recorded per workstream above.

EUR-04 exit: all confirmed gaps have passing owner-boundary evidence, including
real API/object-store transfer/finalization, exact recovery, retained lifetimes,
source/draft identity, spreadsheet continuity, accessibility, and reviewed visual
presentation. The isolated blank-row measurement passed 14/14 units in
`20260915T000400Z-p93426`: 100 qualified samples, p95 87.1 ms against
150 ms. No performance budget was changed.

EUR-04 was saved DONE before beginning EUR-05. Its dependent terminal
maintenance, ordinary visual validation and handoff checks are complete below.

## Compatibility, limitations and rollback

Public routes, request/response schemas and authorization precedence remain
unchanged. No database migration, dependency, persistent browser storage,
renewal/resumable protocol or client cleanup API was added. Retention is limited
to the current live browser runtime and its owner-defined account/incident
boundaries; reload, crash and cross-tab recovery are not provided.

Observable compatibility changes are intentional: zero-byte files are admitted;
multiple-file gestures are refused together; attachment preserves custody; new
file-backed Evidence starts requested; finalized associated files count regardless
of custody. Preview/download remains subject to its existing access rules, so
an explicit ordinary custody action may be needed for access.

Existing materialized Timeline projections can retain the old count until a
source refresh or the existing server-owned incident projection rebuild. The
Timeline provider already declares RefreshRow and IncidentRebuild; no new
backfill or migration mechanism is introduced. Adoption on an existing database
should use that existing projection-maintenance boundary where needed. This
implementation performed no rebuild or modification of analyst data.

Rollback is one coherent patch across the Core 01/Core 03 clarifications,
authored source/test ownership and routing, Go/TypeScript implementation,
generated topology/batch outputs, tests, and the golden/manifest pair. Restore
removed chain helpers only together with their old callers and ports; do not
leave mixed mutation authorities. Use repository generators for derived files.
Re-derive materialized projections through their existing owner if rolling back
count semantics on an adopted deployment. Preserve accepted Evidence/Timeline
records, revisions and associations: no data rollback, destructive cleanup or
client deletion is part of this plan. Abandoned blob cleanup remains with its
existing server owner.

## EUR-05 final validation and scope review

`make agent-finalize` passed again after golden promotion: 1/1 in
`20260915T000729Z-p26509`, before terminal verification. `RESULTS_DIR` remained
unset, so retained successful full-warm run maintenance was skipped. No prior
run was represented as a current full-warm check.

After promotion, `make generate-drift` passed 4/4 in
`20260915T000758Z-p30216`; `make generated-artifact-policy-check` passed 3/3 in
`20260915T000758Z-p30222`; `make json-shape-check` passed 3/3 in
`20260915T000758Z-p30228`; the sequential `make test-catalog-check` exited zero.
The full frontend, focused owner/service, browser, accessibility and measurement
results above use the final executable source. Golden promotion changed only
the reviewed PNG and its generated digest; their fresh visual results follow.

Fresh ordinary visual validations against the promoted manifest:

- `make browser-e2e-visual`: PASS 12/12 in `20260915T000758Z-p30454`;
- `make browser-e2e-visual`: PASS 12/12 in `20260915T000758Z-p30462`.

Both runs passed all 47 scenarios and all 252 active captures. Each retained
`browser-e2e-visual/frontend-visual-reconciliation.json` accounts for 252 PNGs,
29 registered fixtures, zero missing/ambiguous/orphan mappings and no errors.
The viewport, renderer/font pins, masks and comparison tolerances were unchanged.
These are fresh independent ordinary runs, not update-mode or retained-run
substitutes. All applicable executable terminal checks are now passing.

### Independent authority, placement and verification map

| Boundary | Behavioral authority | Authored implementation/source placement | Verification routing and evidence |
| --- | --- | --- | --- |
| Slot and single-upload lease | Core 01 §3.3.8, Core 02 §13, Core 04 §2.0A | Evidence upload/finalization files; `workbook/adapters/createEvidenceFileTransport.ts`; `services/workbookEvidence.ts` | `tools/test_families/module.evidence.json`; private transport, retained recovery and real API/object-store rows |
| Existing Evidence attachment | Core 01 Evidence view/access contracts, Core 02 custody/association, Core 03 §8 | `WorkbookEvidenceAttachmentOwner.ts`, `EvidenceFileFinalization.ts`, `useEvidenceWorkbookBindings.tsx`; Evidence Go attach/create routes | Evidence unit/store/browser and access authorization rows |
| Timeline file association and draft promotion | Core 03 source coordination, retained attempts, first-input creation and §8 clarification | `WorkbookTimelineFileOwner.ts`; Timeline composition and `WorkbookTimelineMutationOwner.ts`; `createTimelineFileLinkTransport.ts` | Evidence file browser/draft rows; Timeline creation, paste, keyboard, projection and measurement rows |
| Custody-independent file count | Approved Core 01 Timeline `evidence_count` clarification, Core 02 lifecycle distinction | `timeline/workbookprojection/collection_facts.go` | Timeline state matrix/derivation equivalence; Evidence upload-to-projection and projection rebuild rows |
| Lifetime, acknowledgement and read debt | Core 03 REQ-03-100/299 and current acknowledgement/source owners; Core 04 §2.0A | `WorkbookMutationRuntime.ts`, shell infrastructure/providers, retained feature owners | Retained timeout/late-callback/security tests; real-service remount and failed-refresh scenarios |
| Compact recovery and spreadsheet interaction | Current Core 03 spreadsheet owners; `docs/design.md` within its design boundary | `EvidenceFileRecovery.tsx`, `TimelineWorkbookGrid.tsx`, panel and editor bindings | Accessibility, keyboard/paste/fill, 620-unit frontend run, measurements and production visual catalog |
| Source and generated boundaries | Root AGENTS and adopted subsystem architecture/harness owners | `tools/frontend_source_ownership.json`, authored family manifests and local source READMEs | Import/source ownership, catalog, selector, JSON-shape and generated-policy/drift checks |

The source-placement manifest assigns files; it does not supply behavioral
requirements. The test families select evidence; they do not define behavior.
The digest and historical handoffs supplied navigation and regression boundaries.
No executable source/test/generator reads, stats, hashes or otherwise depends on
Markdown. The new recovery tests and adapters contain no Markdown dependency;
source/import and generated-policy checks passed.

### Structural disposition and removed paths

Neutral slot validation, one-transfer capability handling, exact attempts and
full receipt retention are shared because those decisions are common. Existing
Evidence finalization policy and Timeline association/draft promotion remain
with their feature owners. Both subscribe through the current retained incident
runtime; there is no generic workflow framework, second Timeline queue or
parallel mutation authority. Future compatible stages can extend the typed
ports and receipts without moving feature policy into transport helpers.

All supported consumers migrated before removing:

- `apps/web/src/workbook/features/evidence/createUploadedEvidenceBlob.ts`;
- `apps/web/src/workbook/features/evidence/createEvidenceAttachmentPort.ts`;
- `apps/web/src/workbook/timeline/adapters/createTimelineEvidenceAttachmentAdapter.ts`;
- `apps/web/src/workbook/timeline/ports/TimelineEvidenceAttachmentPort.ts`.

The old ID-only attachment command was removed from the mutation port. Obsolete
chain references are absent from `apps/web`, `packages`, `internal` and `tools`.
The current source READMEs and ownership manifest describe the retained owners,
ports and lifecycle boundaries. Required custody/access, ordinary creation,
metadata-only related Evidence, inspector recovery and bulk/grid behavior remain
covered by their existing supported paths.

### Skipped and excluded work

No full `make check`/release/deploy run was claimed: current task guides selected
focused service/owner rows plus the full affected frontend and visual catalogs.
Unchanged migration/toolchain pins, dependency installation artifacts and broad
security scanners were not independently rerun; applicable generation and
schema/policy checks did run. Persistent storage, lease renewal, multipart,
resumable uploads, new endpoints/dependencies, bulk upload and analyst-data
maintenance are outside this seam. Preview/access security remains covered by
the selected existing Evidence owner rows. Visual/accessibility outputs are
implementation support, not Core 05 publication evidence.

Existing projection rebuild regression: `make service-backed-test-slice
OWNER=module.evidence ROWS=module.evidence.integration.attached_evidence_projection_storage_can_be_corr_933714dcfe`
passed 3/3 in `20260915T000937Z-p1852`, including preservation of source/history
and blob associations while repairing projection storage.

### Digest acceptance assessment

The applicable advisory acceptance rows are assessed against adopted owners and
actual evidence. A PASS below is implementation evidence, not a new normative
requirement or publication claim. No row is excluded as N/A.

| Row | Status | Evidence and disposition |
| --- | --- | --- |
| A001 Authority | PASS | Approved decisions and independent owner/source/verification map above; Core 01/Core 03 changes are narrow clarifications. |
| A002 Scope | PASS | Common upload/receipt decisions share neutral mechanics; feature policies remain separate; all four retired paths and migrated consumers are accounted for. |
| A003 Repository state | PASS | Clean actual main/HEAD captured before edits; current source/import manifests and local guides inspected; final diff remains at the same HEAD. |
| A004 Tokens | PASS | Recovery reuses existing Evidence button/message styles and spacing tokens; flex/overflow allocation adds no color, typography, theme or density registry. |
| A005 Theme | PASS | Existing dark_graphite remains the exposed theme; token/selector tests and full production visual fixtures cover it. |
| A006 Density | PASS | Shared density unchanged; grid measurements, frontend geometry tests and responsive production captures pass. |
| A007 Creation | PASS | Atomic initial-blob create, ordinary minimum signals, metadata-only regressions and both shared Timeline draft-promotion orderings pass. |
| A008 Responsive | PASS | 1280/768/390 recovery controls, long filename, zoom and current shell responsive fixtures; current frontend responsive tests pass. |
| A009 Overflow | PASS | Grid and inspector keep owned scrolling; local recovery wraps and the status/navigation remain reachable in reviewed captures. |
| A010 Inspector | PASS | Both presentations subscribe to one retained owner; source/detachment/authority review invalidation and existing inspector recovery tests pass. |
| A011 Continuity | PASS | Original source binding, off-page reads, source deletion, navigation, late acceptance and shared draft promotion pass; Timeline remains mounted. |
| A012 Transactions | PASS | Existing random transaction generator; immutable operation/route/body/authority/version capture; same-frame duplicate guards and exact-byte replay assertions. |
| A013 Acknowledgement and recovery | PASS | Complete receipts precede effects; late link failure preserves Evidence; accepted/failed-refresh recovery is reads-only, including remount. Existing queue retry/discard tests remain passing. |
| A014 Editing | PASS | Full frontend and owner-selected real keyboard/paste/fill rows; text clipboard handler regression repaired; ordinary input proceeds during transfer. |
| A015 Conflict | PASS | Local retained source-review/rejection feedback plus existing cell/inspector conflict fixtures; saved source and authored work stay distinct. |
| A016 Query and interaction states | PASS | Existing data/interaction matrix in full frontend; read debt and authority admission remain separate; malformed/current-source recovery tests pass. |
| A017 Refresh and authorization scope | PASS | Same-account recovery, session replacement, role/closure concealment, scoped revocation, account retirement and obsolete callbacks covered; operation denial does not invent revocation. |
| A018 Evidence | PASS | Custody/default lifecycle and blob state matrix, protected access regressions, atomic association, finalized-file count and no placeholder/incomplete count. |
| A019 Accessibility | PASS | Exact semantic controls, keyboard focus, contrast, reduced motion, zoom/narrow layout and one live save-acknowledgement owner; accessibility and keyboard rows pass. |
| A020 Components | PASS | Long-content recovery, wrapping, compound state precedence and shared component/density/zoom fixtures pass. |
| A021 Virtualization | PASS | Current Timeline virtualization and full frontend continuity rows; typing/focus/movement measurements pass and isolated creation p95 is 87.1 ms. |
| A022 Visual fixtures | PASS | Reviewed Make-owned one-golden refresh and two fresh ordinary full-catalog runs passed: 47 scenarios and 252 reconciled captures in each run. |
| A023 Selectors | PASS | Stable record/field/view and named owner-state controls; `make test-slice OWNER=package.ui` passed 10/10. |
| A024 Test authority | PASS | No executable Markdown dependency introduced; source, catalog and generated-policy checks pass. |
| A025 Generated artifacts | PASS | Authored ownership/routing edited first; Make generation and golden maintenance produced derivatives; post-promotion drift/schema/policy checks pass. |
| A026 Compatibility | PASS | Public contracts unchanged; approved custody/count/admission corrections, memory lifetime, existing projection maintenance and coherent rollback are explicit. |
| A027 Handoff | PASS | Sequential exits, actual commands/artifacts, failures and recovery, acceptance, compatibility, limitations and rollback recorded; terminal visual, Markdown and whitespace checks passed. |

Final scope review: all changes are this seam's specification clarifications,
implementation, supported-consumer migration, focused tests, ownership/routing,
source guides, generated derivatives, reviewed golden and controlling handoff.
No digest, lockfile, generated root by hand, commit, push, deploy or analyst-data
mutation occurred. `git diff --check` passed at review. The temporary baseline
worktree and borrowed symlinks were removed; its evidence remains in ignored
`.cartulary/eur-baseline-control/`.

Final document validation: `make lint-markdown` passed in
`20260915T001352Z-p20621`, with
`adhoc/lint-markdown/tool-run-summary.json`. The whitespace preflight found one
extra blank line at this new handoff's EOF; it was removed. `git diff --check`
and `git diff --no-index --check /dev/null <new-file>` for every untracked
authored file then passed (the latter uses exit 1 for a clean nonempty new
file, with no whitespace diagnostic). Branch remains `main` at the original
HEAD; all changes remain uncommitted.

EUR-05 is DONE. All A001–A027 rows are PASS; there is no applicable owner,
product or harness blocker. Intermediate failures retain their actual results
and are resolved by the passing boundary evidence above; none was waived or
converted into a pass. Next action: review/adopt this uncommitted seam. No
further implementation, data maintenance, commit, push or deployment is part of
this execution. Stop after this seam.
