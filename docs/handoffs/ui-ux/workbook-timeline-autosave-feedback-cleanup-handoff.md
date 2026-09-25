# Timeline autosave-feedback cleanup

> 2026-09-25 qualification: the earlier TAF-01–05 entries below are historical
> observations, not a permanent zero-render contract. Their claims of “zero
> unrelated production grid work” overstate what whole-root DevTools counts can
> establish. The 200-row/open-inspector failure retained three collection renders
> across two commits with stable grid inputs and an intact draft. Its source is
> still unproven. A passing rerun does not resolve that attribution. The remediation
> section at the end records current contracts, changes and verification.


## Control and execution baseline

Execution began on clean `main` at
`7800b9602dbfadaa7e3b8cde6f16da3c583c1504`. The recommended
`b9e182b710f6` baseline is three commits behind. Those commits change Network
Flow and related owners, but not the Workbook implementation in this slice.
The root `AGENTS.md` is the only applicable repository instruction file.

Scope: Timeline pending-save observation and feedback, shared mutation-status
projections and immediate consumers, typed query/error presentation, necessary
shell recovery integration, focused verification inputs and reviewed visual
evidence. Preserve mutation identity, FIFO, coalescing, retry, receipt acceptance,
authorization, and spreadsheet interaction. No public API, persistence,
dependency, preference, notification-system, or analyst-data changes. Do not
edit the digest, commit, push, or deploy.

Source inspection covered the eight requested entry points, their composition
and shell consumers, the queue/conflict/surface owners, semantic selectors and
browser helpers. Source ownership is `web.workbook`; verification starts with
`web.workbook`, `module.timeline`, and `module.workbook`. Selector retirement also
requires `package.ui`. Authored test catalogs, source ownership and topology are
inputs; generated roots and managed outputs are changed only by their owners.

## Authority and state matrix

The localized digest was consulted in its prescribed order. It and historical
save-status, recovery-navigation, grid-state, collection-input, range-entry,
clearing, Find, and frozen-column handoffs are advisory regression evidence.
The research NLSpec description is methodological context, not a new task.

Behavior owners are Core 01 §3.3.5 (mutation identity and dispatch), §3.3.10.1
(exact sheet identity); Core 03 §§3–4, especially REQ-03-035/089/099/100/282/283,
REQ-03-286/299 and §13; Core 04 §§1–2; and design §§10.1, 10.5, 10.8. Domain
§§1, 8–10 supplies vocabulary and navigation. Machine projections and verification
routing do not supersede these owners. Product verification must not consume
Markdown.

Primary labels below assume no higher-priority concurrent condition. Secondary
selection retains design §10.1 precedence: transaction collision, overflow,
active-sheet same-field conflict, other terminal replay failure, authentication,
refresh wait, ordinary progress. Surface candidates match the complete
`SheetRef`; workbook blockers remain global. The shell save-announcement owner
announces save-label transitions; required local error/recovery announcements
remain with their existing semantic owner.

| State | Primary / secondary | Local feedback and recovery target | Announcement / focus |
| --- | --- | --- | --- |
| Ordinary queued autosave | Syncing / queued progress | Retained edit; no action required | Shell save host; preserve editor, selection and scroll. |
| In-flight autosave | Syncing / syncing progress | Existing pending editor state; no replay claim | Shell save host; no automatic movement. |
| Successful acknowledgement | Saved / none | Authoritative values and committed version | Shell save host; existing accepted navigation only. |
| Connectivity pause | Syncing / pending progress | Retained FIFO and draft; automatic owner-permitted replay | Shell save host; preserve focus. |
| Authentication pause | Syncing / authentication required | Existing session recovery; conceal protected presentation while retaining same-account work | Existing session notice plus save host; no progress-driven focus. |
| Refresh blocks replay | Syncing / exact-sheet refresh wait | Existing refresh status/recovery | Existing refresh owner; preserve continuity. |
| Accepted write, failed read | Saved / saved changes, refresh pending where eligible | Stale authorized rows and reads-only recovery; never resend accepted write | Query/recovery error owner; preserve focus and selection. |
| Same-field conflict | Conflict / active-sheet affected count | Cell marker; saved value separate from retained draft; explicit resolver | Existing conflict owner and save host; activation enters resolver, resolution returns to semantic cell. |
| Transaction collision | Conflict / safe transaction recovery | Persistent non-modal notice; new-ID Retry and Discard | Existing recovery panel polite announcement; arrival preserves focus, activation moves it. |
| Other terminal replay failure | Conflict / safe failure summary | Persistent notice; permitted Discard, no re-key retry | Existing failure owner; explicit activation only. |
| Queue overflow | Conflict / queue full | Retain 64 units and refused draft; persistent overflow recovery | Existing assertive overflow feedback; no automatic move. |
| Unsubmitted retained draft | Saved / none from this draft | Exact raw text in original editor or retained-draft recovery | Local validation when applicable; no save-progress announcement. |
| Work on another sheet or saved view | Global label / eligible exact-sheet secondary only | Original operation and recovery identity survive detachment | No stale completion restores old focus or opens Inspector. |
| Read-only transition | Mutation-derived / independent permission description | Authorized rows and drafts remain readable; no new write dispatch | Existing authority owner; no save-driven focus. |
| Incident access loss | Protected workbook withdrawn | Clear protected presentation and exit only addressed incident | Existing root announcement; focus Incidents heading. |
| Session replacement | Old protected work cleared before new account exposure | Same-account recovery retains work; different account retires it | Existing account owner; obsolete callbacks cannot replace current authority. |

No owner change is needed for the source-confirmed cleanup. Query data state
must remain independent of save state and interaction permission; hiding a
duplicate message must never classify a failed query as ready. Security retention
does not authorize exposure of protected rows during account recovery.

## Workstream tracker

| Workstream | Status | Exit and next action |
| --- | --- | --- |
| TAF-01 Characterization and owners | DONE | Production baseline, complete matrix and bounded ledger recorded; no owner decision remains. |
| TAF-02 Observation boundary | DONE | Narrow complete observations pass focused and production proof. |
| TAF-03 Feedback and retirement | DONE | Queue overlay and obsolete projections/selectors retired; recovery passes. |
| TAF-04 Production proof | DONE | Production, accessibility, unchanged AC-043 and reviewed visual refresh plus two ordinary passes complete. |
| TAF-05 Final validation | DONE | Finalizer, terminal gates, acceptance, rollback and final scope complete. |

## Remediation ledger

G1/G2 are confirmed structural costs, G3 is a confirmed copy and obstruction defect,
and G4 is a confirmed independent-state defect. The workstream evidence below
records binary remediation validation; final acceptance remains gated on TAF-05.

| Gap / classification | Remediation and affected areas | Rationale / lasting benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 Duplicate React queue projection; structural cost | Remove Timeline snapshot/setter threading and redundant queue presentation after migrating consumers. | One status owner; fewer independent publications and copy paths. | Internal callers and selectors migrate together; retain pending execution refs and unload warning. | Ordinary progress has one visible presentation and correct counts; unchanged facts retain observation identity. |
| G2 Broad runtime observation; structural cost | Subscribe status/recovery where consumed; Timeline observes only relevant conflicts. | Status changes do not rebuild unrelated grid work. | Execution wakeups and every meaningful conflict/recovery/authority transition must survive caching. | Status-only scenarios preserve grid inputs; real changes update; admitted operations still settle exactly once. |
| G3 In-flight work described as waiting to replay; confirmed copy and obstruction defect | Use existing shared projector and remove Timeline queue-copy helper. | One accurate progress explanation. | No protocol change; exceptional persistent notices stay. | In-flight-only work displays shared syncing copy, with accessible recovery intact. |
| G4 Copy-dependent query presentation; confirmed independent-state defect | Separate typed query failures and operation feedback; remove message equality and blocker suppression. | Stale data remains explicitly stale regardless of save state or wording. | Clipboard/fill/clear feedback must remain local and accessible. | Concurrent failure/blocker and identical-message cases retain correct data state and permitted recovery. |

Material advisory dispositions: R006/R007/R008/R010/R014 ADOPT for local feedback,
independent states, exact recovery and continuity; R012 ADAPT through existing
owner lifetimes and measured work; R034 REJECT incidental selectors. No proposed
optimization is justified by a speculative latency failure.

## Evidence and command ledger

- Planning and execution-entry Git checks: clean branch and HEAD above.
- Planning `make help`, `make help-all`, and task guides for the three primary
  owners: PASS. Revalidate routing when authoring new evidence.
- Production characterization, performance and acceptance: see workstream evidence below.

## Acceptance and delivery

Applicable digest acceptance rows require PASS with current evidence; N/A requires
an owner/scope rationale. An applicable BLOCKED row prevents completion. Final
assessment, changed paths, before/after observations, command run roots,
limitations and skipped-check rationale will be appended during execution.

Rollback is limited to this slice's authored hunks, generated projections and
reviewed goldens; preserve unrelated work. No database or wire migration is
planned. Current next action is recorded at the active workstream exit.

### TAF-01 production observations

The isolated baseline recovery slice passed 13/13 execution units at
`.cartulary/test-results/20260919T221549Z-p27357`: exact saved-view conflict
scope, detached work and announcements, global FIFO recovery, successive
edits, refresh-only recovery, role changes and session recovery.

The production autosave characterization plus existing collection-authoring and
first-input-promotion scenarios passed 13/13 units at
`.cartulary/test-results/20260919T221909Z-p69918`. The autosave group retains
`autosave-feedback-observations` JSON and `held-save-full-workbook` PNG attachments
in its Playwright report; decoded copies are beside that report for review.
Eight cases cover 1/100/200/300 loaded rows and closed/open Inspector.

- Every unchanged runtime notification replaced the status snapshot, triggered
  two commits and six column-prop replacements. It rendered 35/38 collection
  cells for 100–300 loaded rows with Inspector closed/open, without replacing
  row arrays. These are virtualized-window counts, not counts per loaded row.
- Status-only explicit-operation transitions caused 18 column-prop replacements
  and 105/114 collection renders in the same larger windows.
- The queue notice overlapped the grid in all eight cases. Full-workbook review
  shows it obscuring the first-row summary editor while the status strip already
  says `Syncing / Workbook edits are syncing.` The overlay instead says
  `Queued edits are waiting to replay.` during an in-flight request.
- Native collection text, backward selection and focused DOM identity survived
  all eight status-only scenarios. Each case accepted exactly one collection
  write and two scalar writes, including held Tab and rapid Enter acknowledgement.
  Authoritative reads confirm the saved tags and summaries.
- Focus-loss and duplicate-write hypotheses are rejected for these scenarios.
  No measured latency regression or speedup is claimed. The work-count evidence
  confirms G1/G2 structural cost; the screenshot confirms G3 visible obstruction.
  G4 remains a source-confirmed owner disagreement with focused regression proof
  required in TAF-02.

`make generate` passed at `20260919T221813Z-p62269`; topology additions are
owned by the new authored characterization row. `make format` passed at
`20260919T221822Z-p65259`. Initial `make frontend-typecheck` failed at
`20260919T221910Z-p70200` because the new E2E diagnostic's production type import
pulled CSS into the E2E TypeScript project. This was initially suspected to be
baseline tooling drift; investigation established it was test-authoring related.
Replacing it with a small browser diagnostic port fixed the issue without
changing application types or configuration. Typecheck passed at
`20260919T222046Z-p3757`.

Next action: finish the comparable real-row-change counter and baseline AC-043
run, then close TAF-01 before changing production code.

### TAF-01 exit

The final comparable diagnostic scenario passed 11/11 units at
`.cartulary/test-results/20260919T222609Z-p45858`. It additionally records real
collection/scalar mutation work: 27 row-array replacements in each case, distinct
from zero during status-only updates. Its eight-case JSON and full-workbook PNG
are retained in the autosave group's report and decoded beside it. This is the
baseline for final comparison; render counts are diagnostic, not latency.

Baseline four-predicate AC-043 passed 20/20 units at
`.cartulary/test-results/20260919T222101Z-p4249`; aggregate status is qualified,
with one canonical snapshot builder, four clones and zero scheduler overlaps.
P95 ms: selection 32.6, focus/edit 32.9, typing 33.0, blank creation 91.3.
The limits remain 100/100/100/150 ms, with 100 samples per predicate.

Focused `make test-slice OWNER=web.workbook` status, responsibility, committed-grid
and recovery-navigation rows passed 5/5 units at `20260919T222619Z-p62079`.
TAF-01 is DONE. No speculative performance failure, unresolved product choice,
or owner contradiction remains. Next action: begin TAF-02 observation changes.

### TAF-02 exit

Runtime status now caches explicit queue counts, immutable conflict-store entries,
refresh scopes/debts and authority generation. `WorkbookMutationRuntime.emit`
still wakes batches and publishes execution notifications on every call.
`WorkbookPendingQueueState.statusFacts` avoids copying executable payloads;
`pendingUnitFacts` independently supplies retry continuation identity/status/order.
No recursive equality, serialization, debounce, or selector framework was added.

Connected `WorkbookObservedStatusStrip` and `WorkbookActiveSurfaceFrame` own status
subscriptions. Shell infrastructure and Timeline composition no longer mirror
status. `useWorkbookMutationConflicts` selects only the schema's conflict entries.
`useTimelinePendingSaves` is removed; execution refs remain with the owner.
Unload-warning observation moved to the existing shell save-announcement host.
The duplicate mounted publication in `WorkbookTimelineMutationOwner` is removed.

`workbookLifecycleModel` separates operation feedback by family from loader-owned
query errors. Clipboard/fill/clear and local mutation errors no longer masquerade
as failed reads. Presentation removed both message equality and save-blocker
suppression, so an authorized failed refresh remains stale. Terminal queue and
authentication failures keep their existing typed recovery owners.

Changed paths are the runtime queue/conflict/surface/projector and hook modules,
shell and immediate status consumers, Timeline foundation/composition/mutation/
presentation modules, focused tests, and `tools/frontend_source_ownership.json`.
Internal presentation ports changed; mutation and wire contracts did not.

Validation:

- Status, lifecycle, responsibility, autosave and navigation unit rows: PASS 6/6,
  `20260919T223908Z-p90358`.
- Timeline mutation coordinator: PASS 2/2, `20260919T223936Z-p96334`.
- Runtime coalescing, identity, token and security/continuity rows: PASS 5/5,
  `20260919T223938Z-p96754`.
- Queue, FIFO and shell status rows: PASS 4/4, `20260919T224200Z-p43310`.
- Final focused status assertions (same-label counts, replacement conflict/token,
  exact saved-view scope, replacement refresh target, immutable prior observations,
  execution publication and authority invalidation): PASS 2/2,
  `20260919T224201Z-p43535`.
- Typecheck: PASS `20260919T224023Z-p98482`. Earlier failures at
  `20260919T223329Z-p80439`, `20260919T223647Z-p81687`, and
  `20260919T223918Z-p95581` were migration-related missing/obsolete props; fixed.
- Format: PASS `20260919T224149Z-p38905`. Earlier format/Biome failures at
  `20260919T223813Z-p82532`, `20260919T223907Z-p90203`, and
  `20260919T223935Z-p96163` were new assignment-expression/non-null syntax; fixed.
- Generate: PASS `20260919T224116Z-p35515`.
- Production characterization: PASS 11/11 `20260919T224024Z-p99559`.
  All eight unchanged notifications retain identity and produce zero commits.
  Status-only transitions produce three status/recovery commits and zero row,
  column or collection-cell work. Actual writes still replace row arrays 27 times
  per case and preserve editor identity, text and selection, authoritative values,
  one collection write and two scalar writes. This is removed work, not a latency
  claim. The retained queue card still overlaps the grid in all eight cases.

G1/G2 observation validation PASS. G4 independent-message unit validation PASS;
production overlapping-error coverage continues in TAF-04. No unresolved owner
choice or observation blocker remains. Next action: TAF-03 removes the ordinary
queue card and retires its unused projection and selector consumers.

### TAF-03 exit

Removed the ordinary queue card, independent queue-copy function and card-only
styles from `TimelineWorkbookNotices.tsx`. Its auto-resolution disclosure,
source-specific Review and Undo remain unchanged. Removed unused
`workbookPendingQueueSnapshot`, its types, Timeline snapshot props and the obsolete
hook guide registration. Retired both queue selector exports and their unit
assertions from `packages/ui-contracts`.

Migrated production browser and unit consumers in collaboration, public-route,
query, accessibility, visual, mention and row-mutation helpers to semantic save
status, the existing `data-pending-replay-count`, status secondary text, and
persistent recovery controls. Blocked queues assert Conflict/recovery rather
than zero pending work. No hidden compatibility DOM remains; a source audit
finds no consumers of the retired symbols.

The shell status strip retains exactly Saved/Syncing/Conflict, existing responsive
allocation and priority, exact-sheet matching and explicit recovery activation.
`WorkbookSaveAnnouncements` remains the sole routine save announcement owner.
Authentication/session, overflow, conflict, transaction retry/discard, terminal
failure and acknowledged-write refresh recovery remain with their original owners.

Validation:

- Shell status unit PASS 2/2 `20260919T224506Z-p50296`.
- Collaboration queue/recovery units PASS 2/2 `20260919T224522Z-p51120`.
- Selector contract units PASS 2/2 `20260919T224523Z-p51345`.
- Format and typecheck PASS `20260919T224449Z-p45442` and
  `20260919T224450Z-p45648`.
- Production status, overflow, exact-view recovery, refresh, role/session and
  successive-edit slice PASS 13/13 `20260919T224524Z-p51730`.
- Overlay-free production characterization plus collection authoring/promotion
  PASS 13/13 `20260919T224613Z-p86108`. The characterization now asserts stable
  unchanged snapshots, zero unrelated grid work and no notice stack during held
  ordinary saves. Every case preserves the exact native editor and three expected
  writes.

G1/G3 presentation validation PASS. Internal selector migration is complete;
no HTTP or storage migration. Required exceptional recovery remains visible.
Next action: TAF-04 proves overlapping query state, broader spreadsheet/accessibility
baselines, final AC-043 and full-workbook visual evidence.

### TAF-04 behavior and visual review

The final diagnostic artifacts at `20260919T224613Z-p86108` use the same fixture,
viewport, loaded-row windows and instrumentation as `20260919T222609Z-p45858`.
Unchanged publications went from two commits/six column replacements to zero;
status-only changes went from 18 column replacements and 105/114 collection
renders in the larger windows to zero unrelated row/column/cell work. Required
real-row replacements remain 27 per case. These counts do not imply a speedup.
Full-workbook PNG review at 1440×900 confirms the formerly covered active summary
cell is unobstructed, with Syncing and progress detail in the fixed status strip.
All eight cases report no queue overlay/grid intersection.

Broader production results:

- Timeline query/conflict overlap, collection lifetime, range cycles/lifetime/native
  editing/settlement, pointer ranges and continuous virtualized scrolling, rectangle
  paste/fill, batch conflicts, and clear draft/predecessor/recovery/conflict cases:
  PASS 17/17 units, `20260919T224938Z-p28115`.
- Workbook exact replay, CSRF recovery, closure, detached availability, all editor
  families, Timeline collaboration, dependent Inspector saves, frozen columns,
  Find editor/virtualized departure, public mutation/recovery routes, save-state
  accessibility and grid correction accessibility: PASS 21/21 units,
  `20260919T225016Z-p90121`.
- Retained auto-resolution disclosure/Review/Undo through real public routes:
  PASS 11/11, `20260919T225355Z-p44109`; `module.entities` guide selected because
  its immediate mention helper was migrated.
- Composition, stateless presentation, grid environment and Inspector lifecycle:
  PASS 6/6, `20260919T225252Z-p42043`.

The overlapping-error scenario keeps stale query data visible beside Conflict,
then beside Saved after explicit discard, until a successful read. One rejected
PATCH is the only write attempt; recovery does not replay it. Source audit also
narrowed `WorkbookSurfaceRefreshNotice` to the cached debt observation, so it no
longer consumes unrelated status. Its unit assertion initially failed at
`20260919T225325Z-p43133` because the test had not established read authority.
The corrected fixture explicitly tests authorized debt, concealment, same-account
restoration, and immutable settled debt; PASS `20260919T225428Z-p74122`.
Typecheck PASS `20260919T225041Z-p24620`; format PASS `20260919T225429Z-p74433`.

Ordinary visual run `20260919T224939Z-p28404` completed 46 passing scenarios and
one expected comparison failure. Reconciliation v3 accounts for 252 capture
intents and 252 active goldens, zero orphan/missing/ambiguous/unresolved entries;
renderer and manifest validation pass. Its only error is the failed comparison.
Reviewed actual and diff PNGs show only retirement of the queue card over the
first grid row. The fixed status strip and full-grid geometry are unchanged.

Accepted refresh trigger: this requested and functionally validated removal of
redundant ordinary save feedback. Affected row:
`module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67`;
scenario `scenario_79fa228c8b2e`; capture ID
`visual.capture.a5356d72e06448c07560`; capture intent
`timeline-mutation-pending-replay-status`; no registered fixture ID (active
nonregistry capture, fully reconciled). Candidate golden:
`apps/web/e2e/workbook.visual.spec.ts-snapshots/timeline-mutation-pending-replay-status-linux.png`.
Viewport 1440×900, 100% zoom, compact density, dark_graphite, masks, scroll
normalization, renderer, and full-viewport scope are unchanged. Refresh and two
fresh ordinary passes remain required before this workstream is DONE.

### Retention and compatibility review

The shared mutation-status projector remains the sole save-label owner. Its cache
compares explicit scalar counts and owner-held immutable references, including
exact refresh scope, replacement conflicts, debt, and authority generation.
Queue execution/settlement never branches on presentation equality. The queue
still preserves FIFO, captured request identity, coalescing, retry boundaries,
conflict gates, receipts and committed-version acceptance.

Preserved: local raw drafts, native editors and caret/selection, first-input
promotion, Inspector continuity, ranges, clipboard, fill, clearing, Find, frozen
columns, ordinary scrolling, exact saved-view recovery, session boundaries,
source-specific auto-resolution, Review/Undo, and persistent actionable recovery.
Simplified: ordinary progress has one existing strip; status and debt observers
subscribe where consumed; grid coordination observes only conflict facts.
Removed: the queue card/copy/styles, React queue mirror and setters, duplicate
mounted publication, unused projected queue snapshot, old selector exports and
hook registration. No temporary adapter or hidden selector survives migration.

Compatibility is internal TypeScript and verification-only. Consumers in this
repository migrated together. There are no route, protocol, persistence,
dependency, or preference changes. The old `getSnapshot` remains available with
stable identity for unchanged status; execution subscribers still receive all
notifications. Retry continuation and refresh recovery use complete narrow
observations. Account replacement and incident access loss retain their existing
security owners.

Rollback: reverse only this handoff's authored runtime/component/composition/test
and catalog hunks; restore the removed hook/projection/selectors together with
all their consumers if reverting the implementation. Regenerate managed browser
and topology outputs with `make generate`; restore the one reviewed golden and
its generated manifest through the visual owner if reverting presentation.
Do not reset the branch or discard unrelated files. No database rollback exists
or is required. No commits, pushes, deployments or analyst-data changes were made.

### TAF-04 measured hot-path comparison

Final AC-043 PASS 20/20 at `20260919T225558Z-p81883`. Aggregate v3 is qualified,
with the identical `ac043_large_grid_snapshot_v1` key
`55233e62069d0026ce0f0925c1efaa9863cc7e43efd6d45f08722f047be53cb4`, one builder,
four clones, zero scheduler overlaps and 100 samples per predicate. No other
browser or test job ran during this measurement. Diagnostic React instrumentation
is confined to the characterization scenario and absent from these measurements.
Sampling, traffic, paint qualification and thresholds were not modified.

| Predicate | Baseline p95 ms | Final p95 ms | Unchanged limit ms | Result |
| --- | --- | --- | --- | --- |
| Selection | 32.6 | 32.2 | 100 | PASS |
| Focus/edit | 32.9 | 29.8 | 100 | PASS |
| Typing acknowledgement | 33.0 | 34.7 | 100 | PASS |
| Blank creation | 91.3 | 76.8 | 150 | PASS |

All existing hot-path predicates pass. These are two qualified samples, not a
statistical claim of latency improvement; typing varies upward by 1.7 ms and remains
well within its limit. The independently established simplification is elimination
of unrelated grid observation/render work during ordinary status changes.

Additional migrated-consumer proof: collaboration different-field/same-field
resolution, conflict FIFO and same-runtime reauthentication replay PASS 11/11 at
`20260919T230048Z-p24929`; `module.collaboration` guide was selected for changed
browser consumers. Timeline public-query row PASS 11/11 at
`20260919T230117Z-p83359`. Clipboard sentinel, fill and clear planning PASS 4/4 at
`20260919T230116Z-p83138`.

Final source review preserves the original successful owner-event result for an
unchanged conflict draft while retaining its observation identity. This prevents
presentation equality from suppressing `emit`/batch wakeups. Added notification
assertion and responsibility rows PASS 3/3 at `20260919T230238Z-p23698`.

## Acceptance assessment

The digest acceptance rows below are advisory review prompts; adopted owners
and current routed evidence establish required behavior. PASS is limited to this
slice's affected behavior and does not claim a product-wide release or Core 05
conformance packet.

| Rows | Assessment | Scope and evidence |
| --- | --- | --- |
| A001–A003 authority, scope, repository | PASS | Revalidated baseline, localized owners, bounded ledger and current source/verification routing above. |
| A004 tokens | PASS | New local error padding uses existing tokens; no new token registry or component-local design literal. |
| A005 theme | N/A | No theme, palette or theme-selection behavior changed; existing visual cases remain regression evidence. |
| A006 density | PASS | No density projection changes; full visual suite exercises existing variants; pending-save capture remains compact. |
| A007 creation | PASS | Timeline first-input promotion and AC-043 blank creation pass; unrelated contextual create capabilities are unchanged. |
| A008–A009 responsive/overflow | PASS | Status allocation unchanged; narrow/zoom accessibility and full-workbook/range-scroll evidence pass. |
| A010–A011 Inspector/continuity | PASS | Inspector lifecycle, dependent writes, detached surfaces, saved views, native drafts/selection and range scrolling pass. |
| A012–A013 transactions/recovery | PASS | FIFO, exact replay, accepted-write/read recovery, new-ID retry/discard, replacement targets and request counts pass. |
| A014–A015 editing/conflict | PASS | Scalar/collection, composition, clipboard/fill/clear/ranges, conflict marker/resolver and distinct saved/draft values pass. |
| A016–A017 query/authority | PASS | Independent-error regression, stale/failed/missing reads, read-only, scoped revocation, same-account recovery and replacement tests pass. |
| A018 evidence lifecycle | N/A | Evidence lifecycle/preview combinations are untouched; the shared status consumer migration introduces no evidence behavior. |
| A019–A020 accessibility/components | PASS | Save communication, keyboard recovery/focus, narrow/zoom correction, Inspector/frozen column and independent error behavior pass. |
| A021 virtualization | PASS | Comparable 1/100/200/300 windows, range/Find virtualized navigation, zero status-only grid work and all unchanged AC-043 gates pass. |
| A022 visual fixtures | PASS | One explained golden refresh and two fresh ordinary passes reconcile all 252 captures/goldens without errors. |
| A023 selectors | PASS | Retired exports have no consumers; semantic status/count/recovery observations and package.ui selector row pass. |
| A024 test authority | PASS | New tests use machine facts, runtime observations and public UI/routes; no Markdown dependency. |
| A025 generated artifacts | PASS | Make-owned generation, finalizer, generated-policy, drift and JSON-shape checks pass. |
| A026 compatibility | PASS | Internal consumers migrated together; no API/data migration, duplicate paths retired and scoped rollback recorded. |
| A027 handoff | PASS | All five sequential exits saved DONE; final scope, evidence, limitations, compatibility and rollback recorded. |

Visual update PASS 12/12 at `20260919T230046Z-p24139`. The Make-owned transaction
promoted exactly the single reviewed pending-replay PNG and its manifest SHA-256
(`7920251723fa0208c34df99a16a1a34b41699929d6c17cdab2cf3edde51c07be`).
The promoted full-workbook image was reviewed again: only the obsolete queue card
is removed; there is no unexplained layout, typography, clipping, focus or status
allocation change. All other goldens remain byte-identical. Two fresh ordinary
validation runs now use the same promoted manifest in isolated harness stacks.

### TAF-04 exit

Both required fresh ordinary visual runs PASS 12/12:
`20260919T230525Z-p25476` and `20260919T230549Z-p55092`.
Each reconciliation is PASS with 252 active captures/goldens, 29 registered
fixtures, zero orphan/missing/ambiguous/unresolved mappings and no errors.
The promoted manifest is unchanged between runs. Full production screenshots,
rather than only a strip crop, support the unobstructed-grid conclusion.

The same-label/same-count replacement blocker and independent pending-unit
identity assertions PASS at `20260919T230811Z-p94331`; format PASS at
`20260919T230813Z-p94651`. Necessary conflict changes continue to update their
consumers; only unrelated status work is suppressed. All G1–G4 remediations pass
their binary behavior checks. No applicable blocker or unresolved product choice
remains. TAF-04 is DONE. Next action: TAF-05 finalizer and terminal validation.

### TAF-05 exit and final delivery

Rechecked `make help`, `make help-all`, and
`make task-guide ROLE=module-author OWNER=<owner-id>` for `web.workbook`,
`module.timeline`, and `module.workbook`: PASS. Additional owners were limited to
`package.ui` (retired selector exports), `module.entities` (mention helper and
retained auto-resolution), and `module.collaboration` (migrated queue/status waits).
Their focused guidance and applicable rows passed as recorded above.

`env -u RESULTS_DIR make agent-finalize` PASS at `20260919T231042Z-p1923`, before
broader terminal frontend verification. Its retained-run selection is explicitly
skipped: no qualifying current full-warm-check root was supplied. Generated
structure refresh, schema/shape, catalog and tier coverage pass. Canonical retained
evidence and scheduler drift maintenance are skipped with
`results-dir-not-provided`; they are not represented as product passes.

| Final command | Result | Run root suffix under `.cartulary/test-results/` |
| --- | --- | --- |
| `make frontend-typecheck` | PASS 2/2 | `20260919T231110Z-p6220` |
| `make frontend-unit` | PASS 653/653 execution units | `20260919T231110Z-p6273` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260919T231110Z-p6318` |
| `make lint-biome` | PASS 2/2 | `20260919T231110Z-p6383` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260919T231110Z-p5968` |
| `make generate-drift` | PASS 4/4 | `20260919T231110Z-p6031` |
| `make json-shape-check` | PASS 3/3 | `20260919T231110Z-p6004` |
| `make lint-markdown` | PASS; final handoff rerun PASS | `20260919T231224Z-p33238`, `20260919T231655Z-p87077` |
| `git diff --check` | PASS | Repository diff |

Browser verification used the current routed
`make service-backed-test-slice OWNER=<owner> ROWS=<selected rows>` selections
recorded in each run manifest above. It includes production, stateful,
accessibility and the four unchanged AC-043 predicates. Full
`make browser-e2e-visual` passed twice after the single Make-owned refresh.
No product acceptance was weakened to obtain a pass. Earlier typecheck/Biome
failures were implementation migration issues, the debt-test failure was missing
fixture authority, and the sole visual failure was the reviewed intentional
queue-card removal. All have passing dispositions; no unresolved harness failure
or applicable BLOCKED row remains.

Final scope is the runtime observation boundary, immediate connected status and
recovery consumers, Timeline queue/query/operation feedback, the retired internal
selectors and hook guide row, focused regression inputs, this handoff, and one
reviewed golden. Managed changes are limited to
`tools/browser_e2e_batch_manifest.json`,
`tools/execution_topology_render_index.json`, and
`tools/frontend_visual_golden_manifest.json`, generated by their owners.
Authored routing changes are in `tools/test_families/module.timeline.json`,
`tools/test_families/web.workbook.json`, and
`tools/frontend_source_ownership.json`.

G1 PASS: no Timeline React queue mirror or duplicate routine card remains.
G2 PASS: stable unchanged observations and zero unrelated production grid work;
meaningful count, target, conflict, debt and authority changes remain observable,
with execution events preserved. G3 PASS: in-flight progress comes from the shared
projector and no longer obscures the active grid. G4 PASS: failed authorized
query data stays stale independently of mutation labels and local operation errors.
Future status additions should extend explicit owner facts and their narrow
observations; they should not recreate React-owned runtime mirrors.

Limitations: diagnostic render counts cover the stated fixture windows and are
separate from latency evidence. Timing is a qualified baseline/final comparison,
not a statistical speedup claim. Visual/accessibility results are implementation
support, not Core 05 claim publication. Full backend, release, `check`, and global
unrelated browser suites were not run because this bounded frontend change adds
no backend/API/storage behavior; relevant service-backed routes and owner-selected
browser/accessibility rows were run. Owner evidence audit was skipped because no
full `check` plus complete support/a11y evidence packet was requested or produced.
Retained-run maintenance was skipped as explained above.

Final Git baseline remains `main` at
`7800b9602dbfadaa7e3b8cde6f16da3c583c1504`. The working tree contains this authorized
implementation and evidence only; the initially clean user baseline is preserved.
All five workstreams are DONE. Applicable acceptance rows are PASS and N/A rows
have owner/scope rationales. Next action: review the uncommitted diff and handoff;
no further implementation, deployment or analyst-data action is pending.


## 2026-09-25 autosave verification and event-driven readiness remediation

Implementation baseline: `52a88e620` (`Centralize workbook write coordination`).
The user authorized implementation of the seven-gap remediation plan. Historical
execution instructions above describe the previous slice, not new authority.
There is no change to product requirements, HTTP, database, saved-view or retained-
draft formats, and no compatibility alias, feature flag or migration.

### Evidence boundary and review groups

Core 03 REQ-03-035/282 owns committed evidence and dependent writes;
REQ-03-087–089 owns autosave/status; REQ-03-283/298–300 owns continuity, drafts
and authority. Core 04 AC-043 owns the four existing performance predicates.
Testing Harness owns execution and artifact mechanics. Frontend implementation
guidance owns narrow subscriptions, stable snapshots and diagnostic limitations.
`domain.md`, Core and the harness specification require no normative changes.
No executable check consumes these Markdown statements.

Review P2 separately from P3. P2 comprises the controlled composition suite,
record-scoped inspector attention, browser choreography and failure diagnostics.
P3 comprises runtime pending-record readiness, the Timeline adapter and required
cancellation signals in its immediate callers/tests. Routing/docs have separable
hunks for both. To roll back P3, revert those production/test hunks and regenerate
its catalog projection; retain P2's causal tests and browser rewrite. Never restore
the invalid browser render gate as a readiness rollback. P2 can be reverted as its
own group only while retaining equivalent controlled coverage of isolation.

### Gap and phase ledger

| Gap | Implementation and lasting effect | Compatibility and residual risk | Exit evidence |
| --- | --- | --- | --- |
| G1 ownership | Dated qualification above and frontend owner map separate implementation diagnostics from product requirements. | No spec behavior change; original three-render attribution remains unresolved. | All assertions map to owners below; no Markdown-consuming executable inputs. |
| G2 causal isolation | Real runtime/editor/Timeline composition under `act`; commits observed at component boundaries, grid inputs by identity. Explicitly gated editor work is a positive control. | Internal tests only; no arbitrary count tolerance. | Unchanged notification and overlapping status tests; selected drafts, accepted rows and conflicts remain observable. |
| G3 readiness | Eight browser cases settle a public Notes create after establishing Timeline editor state, then verify collection/scalar saves. Operation phases have 25-second deadlines. | Same semantic browser rows; native focus/raw value/caret/direction and exact write counts remain required. | Service-backed row and query overlap passed; full gate recorded below. |
| G4 diagnostics | Case allocated before assertions; semantic phase evidence attached before cleanup; all gates released, all disposers attempted, first failure preserved. Private React/runtime discovery removed. | Attachments are internal implementation evidence; metadata excludes payloads, tokens and raw drafts. | Forced setup/editor/response/attachment/cleanup failures exercised by diagnostic helper tests. |
| G5 composition | Controlled tests exposed broad explicit-patch subscription in parent inspector-section construction. Move attention subscription into inspector and observe only selected record facts; action validity uses the same scoped facts. | Unrelated publications still wake execution. Relevant drafts/operations/authority invalidate attention. No duplicate store or blanket memoization. | Isolation regression fails with a deliberately broadened subscription; positive controls and retained-snapshot tests pass. |
| G6 readiness | Existing `WorkbookWriteCoordinator` now owns pending-record wait; Timeline retains committed lookup and one missing-version refresh. Signals required, authority/retirement cancellation prompt, shared reads survive caller abort. | Internal TS interfaces migrate together. History fallback preserved. No timer or polling fallback. | Immediate/pending/refresh, halt/auth/overflow/conflicts, pre-abort, concurrent waits, authority/disposal, cleanup and continuation races covered. |
| G7 accounting | Existing browser and committed-idle rows retained; new composition and diagnostic suites have authored catalog selectors; generated files refreshed only through Make. | No retired row or equivalence alias. Evidence distinguishes infrastructure, regressions, functional and measurement outcomes. | Generation, catalog and schema checks plus terminal ledger below. |

P0 established an executable baseline and separated infrastructure failure from
product evidence. P1 established owners and the replacement matrix. P2 controlled
coverage preceded browser gate removal, and the broad subscription defect was
repaired only after reproduction. P3 is independent of the original render-source
hypothesis. P4 covers generated projections, finalization, integration and handoff.
Current phase exits and limitations are recorded with terminal results below.

### Replacement assertion map

| Removed or retained observation | Replacement and owner |
| --- | --- |
| Immediate status identity + browser zero collection/row/column counts after two frames | `timelineAutosaveIsolation.test.tsx`: unchanged notification, actual component commits, stable grid inputs with dependencies held fixed. `workbookSaveStatus.test.tsx`: immutable snapshots and execution publications. Implementation support only. |
| Synthetic hidden Notes submissions and count 2→1→0 | Controlled composition suite uses real explicit-patch owner with separate preparation gates, checks same-label count changes and grid/inspector isolation. Browser uses one real held public Notes create for continuity. Core 03 save labels plus implementation isolation. |
| Browser row-change work counters | Controlled positive controls for editor activation, gated drafts, inspector selection, accepted query rows and replacement conflicts. Browser authoritative tag/summary reads retain functional acceptance. |
| Browser root commits/DevTools traversal | Retired as causal evidence; no replacement zero-root requirement. AC-043 remains separate, uninstrumented performance evidence. |
| Native editor identity/raw text/focus/backward selection | Retained against the original element handle before and after the Notes response in every browser case; no whole-root quiescence claim. Core 03 continuity/authoring. |
| Collection and scalar commits / notice obstruction | Retained: exactly one successful collection write and two successful scalar writes per case, persisted values, Syncing/Saved transitions, no routine notice stack; one held-save screenshot. |
| Bounded missing-version refresh | Existing committed-idle catalog row expanded, plus existing row-mutation/source-write coordinator rows prove accepted-version dispatch. Core 03 REQ-03-035/282. |

Canonical new rows are
`web.workbook.regression.timeline_autosave_subscription_isolation` and
`harness.browser.boundary_support.autosave_feedback_diagnostics`. Existing
`web.workbook.regression.timeline_committed_idle_refreshes_once_19e6098c0e`
retains its bounded-refresh purpose and adds event-driven lifecycle cases. The
existing runtime-responsibilities row covers coordinator registration/revision
races and disposal. Existing save-status tests retain exact recovery scope,
refresh debt, same-label transitions and immutable historical snapshots.

### Current verification ledger

All run suffixes below are under `.cartulary/test-results/`. Make commands used
the repository's pinned runtime directory `tmp/node-runtime/bin` on `PATH`; the
shell initially omitted Node/pnpm. This was an environment setup correction, not
a harness/product change. No root-cause claim is based on a passing rerun.

| Evidence | Outcome / run suffix |
| --- | --- |
| Original 200/open failure | Retained `20260925T140418Z-p81602`; attribution remains unproven. |
| Planning baseline unit slice | PASS `20260925T162558Z-p79100`. |
| Planning exact browser attempt | No product verdict, `infra/service_start_error`, `20260925T162304Z-p45337`. |
| Executable unchanged browser baseline | PASS 11/11, `20260925T164856Z-p88760`, after fixing PATH. |
| Controlled status coupling before repair | Deterministic isolation failure `20260925T165723Z-p31682`; narrowed inspector observation fixed it. |
| First repaired controlled composition slice | PASS 2/2, `20260925T165914Z-p33893`. |
| Rewritten autosave and query overlap browser rows | PASS 11/11, `20260925T170135Z-p39919`; all eight continuity cases. |
| Collection/editor registry/row mutation/source coordination slice | PASS 5/5, `20260925T170400Z-p74551`. |
| Diagnostic failure/cleanup slice | PASS 2/2, `20260925T170401Z-p74940`. |

Test-authoring failures included unsorted/duplicate catalog titles, a non-exported
test-binding type, an incorrect inspector selector, a timer assertion that counted
an existing authority timer, and a queue-conflict fixture using the wrong public
error shape. These were fixed without weakening production assertions. Their
superseding results and the final checks follow below.

### Final integration evidence

`make generate` passed at `20260925T171058Z-p90528`; only the generated topology
input index changed. `env -u RESULTS_DIR make agent-finalize` passed at
`20260925T171130Z-p93665` before broader checks. Retained-run canonical evidence
and scheduler maintenance were skipped with `results-dir-not-provided`: no
qualifying successful full warm `check` root was supplied. That skip is not a
product pass.

The deliberate broad-subscription sensitivity mutation failed the new composition
row at `20260925T170943Z-p85210`; the mutation was restored. Final focused
composition/readiness/inspector-draft coverage passed 4/4 at
`20260925T171711Z-p89506`. The full frontend run below passed 676/676. The final
focused run verifies the last test-only type/lint corrections after that suite
started; production sources were unchanged during the suite.

| Command | Result | Run suffix |
| --- | --- | --- |
| `make frontend-typecheck` | PASS 2/2 | `20260925T171446Z-p52630` |
| `make frontend-unit` | PASS 676/676 | `20260925T171202Z-p98138` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260925T171202Z-p98203` |
| `make lint-biome` | PASS 2/2 | `20260925T171447Z-p53014` |
| `make generate-drift` | PASS 4/4 | `20260925T171202Z-p97854` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260925T171202Z-p97956` |
| `make json-shape-check` | PASS 3/3 | `20260925T171202Z-p98058` |
| `make lint-markdown` | PASS | `20260925T171202Z-p98706` |

The first terminal typecheck (`20260925T171202Z-p98045`) rejected an extra
`view_schema_id` on a new inspector test fixture; the first Biome run
(`20260925T171202Z-p98339`) rejected two non-null assertions in the new composition
tests. Both were corrected and superseded above. Neither exposed a production
failure. The intentionally failing sensitivity run is regression evidence, not an
unresolved product failure.

Final test-only additions assert that the Timeline adapter schedules neither native
`setTimeout`/`setInterval` nor coordinator delays, and that cancelling one of two
waiters sharing a refresh leaves the other able to accept the refreshed version.
The expanded committed-idle row passed 2/2 at `20260925T172639Z-p67964`;
`make frontend-typecheck` and `make lint-biome` passed again at
`20260925T172712Z-p73102` and `20260925T172712Z-p73140`.

The first full browser gate (`20260925T171710Z-p89337`) finished 138/140: the
asset-lifetime group and its aggregate summary failed. All other groups, including
the final autosave choreography and history/recovery consumers, passed. The outer
row classified its rejected child build as `product/test_assertion_failure`, but
the nested build (`20260925T172608Z-p60437`) recorded
`artifact/artifact_error`: frontend inputs changed after the run source snapshot.
The implementation session added the last test coverage during that independent
build. This is invalid build evidence, not a demonstrated product regression.
Non-documentation inputs were then frozen. The exact
`harness.browser.boundary_support.frontend_artifact_lifetime` row passed 11/11 at
`20260925T173021Z-p84543` without a code fix. A full gate rerun follows below;
no scenario retry, sleep, threshold or golden adjustment was introduced.

The stateful save-status row
`module.workbook.browser_stateful.workbook_save_status_preserves_authoritative_tra_0fa8b0b16c`
passed 11/11 at `20260925T173021Z-p84558`, covering exact saved-view conflict scope,
global FIFO recovery beside independent work, and queue overflow.

### Changed internal interfaces and review boundaries

- P2 production: `workbook/inspector/workbookInspectorOrdinaryAttention.ts`
  projects selected-record owner facts and keeps action validity scoped to those
  facts. Timeline inspector section construction no longer subscribes in the
  surface parent. Its `detailsOwner` is forwarded through the inspector model and
  region; `TimelineWorkbookInspector` subscribes at the actual consumer.
- P2 evidence: `timelineAutosaveIsolation.test.tsx`, the retained inspector-draft
  test, `e2e/timeline-autosave-feedback.spec.ts`, and
  `e2e/support/workbook/autosaveFeedback{,.test}.ts`. The source ownership entry,
  new composition row and new harness diagnostic row belong to this group.
- P3 production: `WorkbookMutationRuntime.waitForPendingRecordIdle` is the new
  internal operation. `useTimelineCommittedRecordIdle` now accepts
  `mutationRuntime` instead of pending refs and the React conflict projection;
  options require `signal`. Its composition call and History callback type migrate
  together; Mention/source-coordination callers already supplied signals.
- P3 evidence: `useTimelineCommittedRecordIdle.test.tsx`, the immediate
  row-mutation fixture/call, and the additional titles on the existing
  committed-idle catalog row. Coordinator registration/race evidence remains in
  the existing runtime-responsibilities suite.
- Shared handoff: the frontend implementation guide and Timeline READMEs document
  these boundaries. The generated topology input index is regenerated after
  either review group changes its authored catalog; do not hand-edit it.

Review groups remain uncommitted and have independent production/test hunks. No
HTTP, storage, dependency lockfile, visual golden or deployment artifact changed.

Retained autosave JSON from the passing autosave group in the first full run was
inspected directly: all eight 1/100/200/300 × closed/open cases have `complete`
phase records, one Notes write, one collection write and two scalar writes. Each
original editor remained connected and focused with its raw draft matching and
backward selection 3–8 preserved; Notes/collection/scalar gates all show started,
released and completed. The attachments retain semantic booleans/identities and
bounded save events, not raw drafts, payloads or runtime mutation handles.

The clean `make browser-e2e-webserver-backed` rerun passed **140/140 execution
units** at `20260925T173214Z-p50921`. The previously affected asset-lifetime group
and the autosave group both passed inside this complete run. The source-snapshot
failure above is fully dispositioned; no product gate or assertion was weakened.
The complete functional selection includes dependent History, capture, mention,
Evidence, authoring/authority, recovery and editor-continuity consumers.

The four existing AC-043 rows passed through the owner-selected
`make service-backed-test-slice OWNER=module.timeline ROWS=<four measurement rows>`
at `20260925T174425Z-p35752` (20/20 execution units). The aggregate is
`browser-e2e-measurement/frontend-measurement-aggregate.json` under that run.
It is qualified implementation evidence with unchanged fixture/sampling/thresholds,
not a speedup or Core 05 publication claim. The fixture is
`ac043_large_grid_snapshot_v1`, key
`520577706a84ac8a294558004b496faed428370ae5b4d2c7da76a3a5f719fa41`, with one builder,
four clones and zero scheduler overlaps. No DevTools diagnostic instrumentation
was installed for this run.

| Predicate | Samples | p95 (ms) | Existing budget (ms) | Result |
| --- | --- | --- | --- | --- |
| Summary selection down | 100 | 32.3 | 100 | PASS |
| Summary focus edit | 100 | 28.2 | 100 | PASS |
| Committed summary typing acknowledgement | 100 | 31.0 | 100 | PASS |
| Blank row creation | 100 | 74.0 | 150 | PASS |

### Phase exits and final handoff

| Phase | Exit disposition |
| --- | --- |
| P0 evidence baseline | DONE: original render failure retained, attribution unresolved, infrastructure/build-snapshot failures distinguished from functional failures. |
| P1 contract cleanup | DONE: owners and replacement coverage are explicit; no render-count product requirement added. |
| P2 reliable verification | DONE: controlled isolation and positive controls pass; regression sensitivity demonstrated; all eight production continuity cases and failure diagnostics pass. |
| P3 event-driven readiness | DONE: coordinator-backed waiting, required cancellation, refresh/lifecycle/authority/registration and continuation-race coverage, committed acceptance and no polling fallback. |
| P4 integration and handoff | DONE: implementation, routing, full frontend/browser/stateful, AC-043, documentation lint and diff verification complete. |

Limitations remain explicit: the original three renders have not been attributed
by new trace evidence. The discovered broad inspector subscription is a separately
reproduced defect, not proof of that original cause. No latency improvement is
claimed. Full backend/release/`make check`, visual and accessibility gates were not
rerun for this internal frontend coordination slice; it changes no backend contract,
visual design, accessibility contract or golden. Relevant production browser and
stateful consumers, all frontend units and the unchanged performance predicates
were verified. Retained-run maintenance remains skipped because `RESULTS_DIR`
was unset, as required for the absence of a qualifying full warm-check run.

Final documentation check: PASS (`20260925T175022Z-p78575`); `git diff --check` PASS. The next handoff action is review of the
uncommitted P2/P3 groups; no data migration, deployment flag or compatibility
cleanup is required.
