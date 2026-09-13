# Workbook presence refactor handoff

## Control and baseline

This is the controlling record for the authorized collaboration presence seam.
No commit is authorized. The digest is read-only. Completed relationship,
collection, Evidence, inspector, recovery, view-bar, density, and grid-state
seams are outside this work.

- Branch: `main`.
- Commit: `aa7c9611318a88c927019827129d8589a05f258e`.
- Initial dirty state: clean, reconfirmed at implementation entry.
- Applicable instructions: root `AGENTS.md`; no nested instructions found.
- Stack: existing React/TypeScript/Vite, Vitest, Playwright, Go; no dependency
  or toolchain changes are permitted.
- Allowed paths: presence code and directly affected workbook consumers under
  `apps/web/src`, presence tests/browser support under `apps/web/e2e`, narrow
  existing lifecycle code/tests under `internal/modules/collaboration`, authored
  design projections and their generator/schema support, authored verification
  routing, affected generated outputs through Make, reviewed presence and
  directly affected consumer goldens,
  this handoff, narrow design clarification, and the explicitly approved Core 01
  timestamp clarification.
- Preserve unrelated work. No APIs, storage schemas, public modes, matching
  identities, authorization policy, durable sequencing/replay, or mutation
  behavior changes are authorized.

## Tracker

Only one row may be `IN_PROGRESS`. Complete and record an exit before starting
its successor; never advance past a blocked dependency.

| Workstream | Status | Exit and next action |
| --- | --- | --- |
| CP-01 Authority and characterization | DONE | Owners frozen; frontend and server failures reproduced. Next: CP-02. |
| CP-02 Scoped presentation | DONE | Pure and canonical-state checks passed; next CP-03 integration. |
| CP-03 Renewal, expiry, integration | DONE | Lifecycle, owner suites, typecheck and import boundaries passed; next CP-04. |
| CP-04 Production and visual evidence | DONE | Multi-client, lifecycle, accessibility, responsive, measurement and two reviewed ordinary visual passes complete. Next: CP-05. |
| CP-05 Final validation and closure | DONE | Finalizer, owner/browser/frontend checks, generation/policy, lint and scope audit passed; handoff closed. No successor seam. |

## Authority and decisions

Inspected Core 00 precedence, Core 01 §3.3.10, Core 03 §4.3, Core 04 session
lifetime rules, the adopted Collaboration Module Boundary Decision,
`docs/domain.md` vocabulary/owner navigation, design §§10.2–10.3 and 14.2 with
density/responsive/component/accessibility requirements, authored WebSocket
contracts, generated-artifact policy, verification routing, and the visual
golden maintenance guide. The digest START_HERE, LOCAL_AGENT_PROMPT, REPO_MAP,
OWNER_MAP, rules, and acceptance files are advisory only. Its localization
commit `2356949f7ec3c8e27ff83ae695d60e06a387d0e5` is older than this checkout;
relevant paths and runnable owners were revalidated without refreshing it.

| Decision | Authority and disposition |
| --- | --- |
| Canonical state | REQ-01-265/266 and REQ-03-094: exact connection-keyed snapshots/deltas; no presentation ordering in replay/cache identity. |
| Scopes | REQ-01-260/262 and REQ-03-092/094: exact sheet, materialized row, writable editing field; Network Analysis remains header-only. |
| Self exclusion | Preserve current-connection exclusion. Owners do not prescribe an account-wide exclusion; another tab of the same account remains a candidate. |
| Focused | Design drift, not a new wire mode: derive row focus from `viewing` with a materialized row. Keep `editing`, `viewing`, `idle` on the wire. |
| Renewal timestamp | User explicitly approved narrow Core 01 clarification during planning: heartbeat renews only expiry; observation remains last accepted presence publication. |
| Lifetime | REQ-01-264/276/277 and REQ-04-010: existing 15-second heartbeat and 45-second lifetime; keepalive never extends session idle expiry. |
| Presentation | Design §10.3: duplicate-state selection, priority/UTC/NFC/code-point ordering, unique-user capacities 5/3/2. |
| Quiet state | REQ-03-095 and design §14.2: no presence-only save/queue/query/focus effects or live announcements. |

No unresolved owner contradiction was found. Runtime/tests/generators
must not consume, stat, or hash Markdown. Owner-to-projection review is human.

## Characterization and lifecycle trace

At baseline, publication went from surface focus/editor bindings through the
coordinator's bounded publication task and the sole incident session into
`hello`, `resume`, or `presence_update`. The server hub stores one record per
connection with observation and expiry. Snapshots prune expired entries and
serialize exact connection order; deltas are ephemeral. The client retains a
connection map; baseline presentation filtered the active sheet/self connection,
then components sorted again and counted connections.

Source-proven gaps before implementation:

- `visiblePresence` counts connections and uses locale collation; the canonical
  projection's presentation helper uses a different string ordering.
- No client expiry check or expiry task exists.
- `pong` does not renew hub presence. `lastSent` advances on every outgoing
  message, allowing broadcasts to suppress a heartbeat request.
- Header presence has implicit live announcements through `role="status"`.
  The shared status-strip slot is already a non-live section; retain that fact.
- Generic/Entity/Assessment consume same-cell markers but publish only
  viewing/idle from selection and lack row-gutter presence.
- `D-VFIX-005` declares full-viewport coverage, but its helper crops the grid
  and resets horizontal scroll left after asserting the edited cell. Manual
  baseline image inspection shows neither the header nor the edited cell.
- Existing shell presence fixtures use historical expired timestamps without
  a controlled wall clock. Correct their clocks, not the expiry rule.

Initial runtime hypotheses were heartbeat starvation under continual broadcasts,
timer-throttling recovery, marker clipping across densities, and production
editor/focus continuity during renewal. The scheduling correction removes
outgoing traffic from the ping-decision inputs and has exact-boundary tests;
actual multi-client renewal proves sustained liveness. Controlled clock-jump
and stale-callback tests prove delayed-timer recovery; browser density/spacing
geometry and editor-identity tests resolved the clipping/focus hypotheses.

## Verification ledger

Task guides resolved `web.collaboration`, `module.collaboration`,
`web.workbook`, `web.design`, `package.ui`, `package.grid_adapter`, and
`module.timeline`. The narrow visual-fixture synchronization also resolved
`module.entities`; its existing authored visual row runs in the ordinary suite.
Use owner `test-slice`/`service-backed-test-slice` first;
select exact authored rows. Browser visual/a11y/stateful and Timeline
measurement targets apply to this seam. Run `make agent-finalize` before broad
terminal checks. Leave `RESULTS_DIR` unset absent qualifying full warm evidence.

Planning baseline passes (existing tests, not proof of corrected behavior):

| Command selection | Result and run root |
| --- | --- |
| `make test-slice OWNER=web.collaboration` with transition-model and coordinator rows | PASS, `.cartulary/test-results/20260905T030917Z-p62977` |
| `make test-slice OWNER=web.workbook` with Timeline presence and shell presence rows | PASS, `.cartulary/test-results/20260905T031007Z-p64182` |
| `make test-slice OWNER=module.collaboration` with semantic hub/protocol row | PASS, `.cartulary/test-results/20260905T031024Z-p64942` |

Implementation commands, failures, corrections, and exits are appended below.

### CP-05 evidence and exit

- CP-04 exit was recorded before starting this successor.
- `env -u RESULTS_DIR make agent-finalize` PASS, 1/1 unit,
  `.cartulary/test-results/20260905T050528Z-p26831`, before broader terminal
  checks. Its `generate-drift`, JSON-shape and catalog/tier checks passed;
  generated structure was unchanged. No qualifying full warm successful check
  evidence exists, so retained-run canonical-evidence and scheduler timing/event
  maintenance were explicitly skipped (`results-dir-not-provided`).
- Terminal frontend typecheck PASS, 2/2 units, `20260905T050614Z-p30554`;
  import boundaries PASS, 2/2 units, `20260905T050614Z-p30596`; script lint
  PASS, 2/2 units, `20260905T050614Z-p30686`; generated-artifact policy PASS,
  3/3 units, `20260905T050614Z-p30287`, under `.cartulary/test-results/`.
- Biome initially failed only formatting of the added visual-fixture assertion,
  `.cartulary/test-results/20260905T050614Z-p30638`. Correct through
  `make format` PASS at `20260905T050721Z-p59105`; Biome rerun PASS,
  2/2 units, `20260905T050815Z-p82973`. Pre-closure Markdown lint PASS,
  `20260905T050815Z-p82975`, under `.cartulary/test-results/`.
- Broad `make frontend-unit` failed 3 of 440 units,
  `.cartulary/test-results/20260905T050614Z-p30646`. Corrections: register both
  new TypeScript files under the existing `web.workbook` source container in
  authored `tools/frontend_source_ownership.json`; replace raw browser editor
  IDs with shared row-cell selectors plus textbox roles and raw marker prefixes
  with accessible scope roles; remove the invalid viewing-plus-field anchor
  from the existing save-state fixture and align its clock to the observation.
  No policy check or save-state behavior was changed. New owner task-guide:
  `web.architecture`, with its source/selector policy slice. Focused verification
  precedes repeating affected browser checks and the broad frontend target.
- Corrected architecture/source-selector slice PASS, 12/12 units,
  `.cartulary/test-results/20260905T051112Z-p89766`; save-state fixture row
  PASS, 2/2 units, `.cartulary/test-results/20260905T051113Z-p89774`.
  All final corrections affect test/ownership inputs only. Production rendering
  and the reviewed golden bytes remain unchanged after the two ordinary passes.
- Repeated `env -u RESULTS_DIR make agent-finalize` after the authored source
  registry correction: PASS, 1/1 unit,
  `.cartulary/test-results/20260905T051212Z-p91962`. Generation/drift,
  JSON-shape and catalog checks passed again; retained-run maintenance remained
  explicitly skipped. Repeated browser and broad frontend results follow.
- Corrected Network Analysis browser route PASS, 11/11 units,
  `.cartulary/test-results/20260905T051308Z-p95216`; the corrected main
  collaboration route failed at `.cartulary/test-results/20260905T051308Z-p95218`:
  its replacement locator was scoped under the display cell's inner test ID,
  which is unmounted during editing. Corrected to the existing editor group
  using its field-contract accessible name within the shared grid selector.
  Focus and geometry assertions are preserved. No runtime change was needed.

- Final broad `make frontend-unit` PASS, 440/440 units,
  `.cartulary/test-results/20260905T051506Z-p85210`. Typecheck PASS, 2/2 units,
  `20260905T051506Z-p85209`; Biome PASS, 2/2 units,
  `20260905T051506Z-p85263`; generated-artifact policy PASS, 3/3 units,
  `20260905T051506Z-p85007`, under `.cartulary/test-results/`.
  The last editor-locator refinement is verified through the affected browser,
  architecture, typecheck and lint routes; production and golden bytes are
  unchanged. Latest architecture slice PASS, 12/12 units,
  `20260905T051747Z-p32285`; typecheck PASS, 2/2 units,
  `20260905T051747Z-p32422`; Biome PASS, 2/2 units,
  `20260905T051747Z-p32473`, under `.cartulary/test-results/`.

- Final corrected collaboration browser route PASS, 11/11 units,
  `.cartulary/test-results/20260905T051747Z-p32273`, including quiet actual
  heartbeats, editor-group focus/geometry, removal and authorization loss.
  Pre-closure Markdown lint PASS,
  `.cartulary/test-results/20260905T051850Z-p79109`; `git diff --check` PASS.
- CP-05 exit complete. All required affected-owner checks have successful
  terminal evidence; failed attempts and their corrections are retained above.
  The scope is 60 related paths, with no unresolved owner contradiction or
  required product change deferred. The final record is closed and re-read for
  UTF-8, final newline, trailing whitespace and Markdown validation after the
  tracker status update. Closed-record Markdown lint PASS at
  `.cartulary/test-results/20260905T052141Z-p81019`; the final prose cleanup
  is re-read and linted once more without changing source or generated artifacts.
  No commit or successor seam is authorized or started.

### CP-04 evidence and exit

- Extended the existing authored real-browser scenario with another tab of the
  same authenticated user on a different row. It verifies one unique user in
  the header and each scope, actual pongs across more than two 45-second
  lifetimes, advancing expiry with identical observation/activity/anchors,
  unchanged publication/query counts and focus/selection/scroll, then tab
  removal, reconnect and final removal. No synthetic wire publications or
  observer-derived lifetime extensions are used.
- First browser attempt failed only on the test's guessed field label
  `Summary`; the production accessible owner label is `Activity Synopsis`.
  Corrected it. Initial run: `.cartulary/test-results/20260905T035044Z-p39313`.
- Quiet renewal scenario PASS (1.7 minutes) and existing multi-client
  reset/replay/conflict scenario PASS in
  `.cartulary/test-results/20260905T035234Z-p19416`. The enclosing selected run
  failed its new accessibility geometry assertions; do not report it as a
  wholly successful run.
- Added a duplicate-user tab to the six-user overflow visual fixture and
  corrected `D-VFIX-005` to a full viewport with production cell-anchor scroll.
  Assert all three indicators and unique counts immediately before capture.
  The second presence golden now retains the edited cell in its grid crop.
- Real browser overflow exposed a new integration defect: inserting the cell
  marker wrapper replaced the focused editor DOM node, triggering editing
  stop publication. Added an expected-failure DOM-identity/focus assertion
  before correction: `.cartulary/test-results/20260905T035503Z-p77687`.
  Kept editor wrappers stable across peer arrival/removal; regression PASS at
  `.cartulary/test-results/20260905T035533Z-p81082`. Six-user cell overflow then
  reached the correct `+4`; no stale presence was retained to obtain that result.
- Geometry characterization distinguished horizontal scrolling from clipping:
  bringing the value into view does not guarantee its trailing marker is in
  view. Tests now use the existing production scrolling helper on the marker.
  Zoom checks retain a supported effective desktop viewport. Density,
  responsive, text-spacing, editor separation and live-ancestor checks were
  still under verification at this point; final passes are recorded below.
- Ordinary visual run `.cartulary/test-results/20260905T035258Z-p75120` failed
  expected screenshot differences and the then-unfixed overflow assertion.
  Reconciliation reviewed: no missing active golden or ambiguous mapping;
  captures after failed assertions were not reached (nine reported orphans,
  three unresolved registered fixtures). None are deletion candidates. No
  golden was mutated during this attempt. Initial diff review ties base-surface
  column shifts to the fixed presence gutter, and inspector fixture changes to presence
  glyphs/colors; completed inspector/Evidence behavior remains untouched.
- Timeline measurement selection PASS, 19/19 units,
  `.cartulary/test-results/20260905T035259Z-p76463`, using the four authored
  Timeline typing, blank-row, arrow-selection and Enter-focus measurement rows.
- Browser test type errors (optional tuple indexes and DOM Node removal) were
  corrected; frontend typecheck PASS,
  `.cartulary/test-results/20260905T035545Z-p5127`.
- The expanded real-browser scenario also covers Hosts and Notes editor
  start/stop plus Assessment's derived row focus, exact sheet exclusion and
  unchanged save state. Whole selected route PASS, 11/11 units,
  `.cartulary/test-results/20260905T040327Z-p83420`.
- Accessibility geometry exposed Timeline's absolutely positioned input using
  the whole cell as its containing block. Reserved editor slots now establish
  their own containing block, retaining existing input and density styles.
  Presence accessibility/conflict scenario PASS, 11/11 units,
  `.cartulary/test-results/20260905T040144Z-p27817`. Earlier geometry failures
  retained at `20260905T035534Z-p81342`, `20260905T035750Z-p38507` and
  `20260905T040009Z-p35444` under `.cartulary/test-results/`.
- Responsive frame geometry route PASS, 11/11 units,
  `.cartulary/test-results/20260905T035825Z-p85726`.
- Screenshot comparisons now accumulate failures through `expect.soft` while
  later captures run. Every mismatch still fails its test and owner row;
  functional assertions remain immediate. This lets ordinary reconciliation
  account for captures after an intentionally changed golden. Authored owner
  `harness.browser` test slice PASS, 28/28 units,
  `.cartulary/test-results/20260905T040638Z-p70244`.
- Added actual same-account extension-tab coverage to the existing claimed
  Network Analysis scenario and its `web.collaboration` collaborator routing.
  Navigation uses the established incident entry plus workspace tab. Initial
  attempts using an explicit extension URL did not render the second tab's
  workspace; retained failures are `20260905T040344Z-p20699` and
  `20260905T040541Z-p22611`. Extension launch behavior is not redesigned here.
  The initially guessed owner `module.network_flow` was rejected by task-guide;
  the resolved authored owner is `module.networkflow`.
- Further ordinary visual runs at `20260905T040011Z-p35773` and
  `20260905T040458Z-p78090` retained failures. The first timed out before a
  socket was created; the latter collected 93/94 captures and exposed a missing
  scrolling-helper import, now corrected. No update was attempted on incomplete
  functional evidence. The capture now scrolls through the neighboring Data
  Source field to leave visible context after the same-cell marker.
- Claimed Network Analysis presence route PASS, 11/11 units,
  `.cartulary/test-results/20260905T040941Z-p92462`; both tabs share one
  authenticated user and current-connection self exclusion is preserved.
- Complete ordinary visual reconciliation reviewed at
  `.cartulary/test-results/20260905T041057Z-p42586`: 94 capture intents, 94
  active committed goldens, 26 resolved registered fixtures, zero orphan,
  missing or ambiguous mappings. All functional capture assertions passed;
  the run failed nine expected screenshot comparisons. This is the controlling
  pre-refresh reconciliation, superseding incomplete attempts above.
- Accepted refresh triggers: readable quiet presence glyphs and other-user
  color; fixed row-presence gutter on shared base surfaces; corrected full
  viewport and edited-field scroll for `D-VFIX-005`; retained edited field in
  the second presence grid crop. No viewport, zoom, density, renderer or mask
  changes elsewhere. Existing nonregistry captures remain active.
- Strengthened overflow geometry with all six users and the duplicate tab
  across compact/default/comfortable density and text spacing. Every glyph's
  box remains inside its owning row/cell. All functional visual assertions
  passed in `.cartulary/test-results/20260905T042113Z-p40450`; its visual row
  failed only the expected old cropped-image comparison.
- The same run passed the quiet renewal/base-surface scenario with real
  session revocation, socket closure, removal and revoked-tab reload. An earlier
  attempt guessed immediate auth-shell presentation and failed at
  `20260905T041341Z-p20574`; the test now waits for the actual revocation and
  socket-close contract, then verifies no restored presence after reload.
- The quiet fixture now seeds both rows before either editor starts. Creating
  its second row during the quiet phase could legitimately cause a live row
  refresh and end the first editor, demonstrated at
  `20260905T041757Z-p87437`; presence-only behavior is measured without that
  unrelated mutation. No query/recovery implementation was changed.
- `make generate` refreshed authored projection/routing outputs successfully,
  `.cartulary/test-results/20260905T041834Z-p34185`. Web design slice PASS,
  15/15 units, `20260905T041328Z-p91856`; UI package slice PASS, 10/10 units,
  `20260905T041329Z-p92114`, under `.cartulary/test-results/`.
- Shared base editor fieldsets retain the adapter's direct-child geometry;
  the presence layout is nested inside. The browser checks both analysts'
  input bounds, separate marker space, and retained focus. Geometry sampling
  waits for the required peer marker, avoiding an input box sampled before
  arrival and a marker box sampled after it. Final selected functional route
  PASS, 11/11 units, `.cartulary/test-results/20260905T042818Z-p48115`;
  prior mixed-frame geometry failure: `20260905T042501Z-p97504`.
- All 163 workbook units PASS after the editor correction,
  `.cartulary/test-results/20260905T042820Z-p48345`.
- Final wire audit found the authored `date-time` timestamp projection admits
  RFC 3339 offset forms. Added a precision-sensitive equivalent-instant case
  before correction: expected FAIL `20260905T043218Z-p11476`. The parser now
  validates calendar/offset syntax and compares normalized UTC instants while
  preserving fractional precision. No wire schema changed. All eight
  collaboration units PASS at `20260905T043240Z-p12179`; final typecheck PASS
  at `20260905T043303Z-p19475`, under `.cartulary/test-results/`.
- Fresh final-source ordinary visual run `20260905T043302Z-p17720` retained
  the nine expected screenshot differences and an intermittent existing
  entity-mention fixture failure: the dismissed-mention assertion observed a
  resolved target and then no mounted item. That completed seam is unchanged;
  its golden is not a refresh or deletion candidate. Presence's full fixture,
  density/spacing assertions and all 26 registered fixture mappings completed.
  Repeat ordinary verification before any mutation of goldens.
- Final-source repeat `.cartulary/test-results/20260905T043725Z-p66844`
  completed all functional captures, including the unchanged entity-mention
  fixture. Reconciliation again accounts for all 94 active goldens and 26
  registered fixtures with zero missing/ambiguous/orphan mappings. Only the
  same nine expected screenshot comparisons failed. This is the final
  controlling ordinary reconciliation for the authorized refresh.

- `make browser-e2e-visual-update` PASS, 12/12 units,
  `.cartulary/test-results/20260905T044018Z-p14069`. It refreshed the nine
  mapped images plus a ten-pixel background-text raster difference in
  `network-flow-analysis-delete-dialog-linux.png`. Reviewed both versions and
  restored only that unrelated PNG from the baseline. No unrelated golden was
  accepted. `make generate` then regenerated the manifest successfully at
  `.cartulary/test-results/20260905T044427Z-p14809`; no generated output was
  hand-edited. The two required fresh ordinary passes are recorded below.
- Final server audit aligned the new incoming-frame session-expiry guard with
  the existing ticker's expiry disposition: revoke the expired session in the
  auth store and hub before closing. This preserves existing cleanup and does
  not advance or alter authenticated session timing. All 31 collaboration
  owner units PASS at `.cartulary/test-results/20260905T044214Z-p60429`;
  final service-backed lifetime/revocation selection PASS, 5/5 units,
  `.cartulary/test-results/20260905T044812Z-p50617`.

- First post-refresh ordinary run failed, 10/12 units,
  `.cartulary/test-results/20260905T044800Z-p19161`, on the repeated unrelated
  mention fixture race. Source comparison found its visual fixture clicked
  Dismiss immediately after Resolve, while the existing accessibility fixture
  already awaited the resolved accessible name. Applied that same three-line
  synchronization assertion to the visual fixture. This tightens fixture
  preparation without changing product behavior, its golden, or comparisons.
  Waiting only for the resolved label still failed at
  `.cartulary/test-results/20260905T045052Z-p84495`. Trace evidence shows both
  resolve and dismiss requests accepted with HTTP 200. The fixture now also
  waits for `Saved`, whose existing action hook sets it after projection
  settlement, before starting dismissal. No mention runtime was changed.
  A guessed visual-row suffix was rejected before execution; corrected it
  from `tools/test_families/module.entities.json`. Focused verification PASS,
  11/11 units, `.cartulary/test-results/20260905T045429Z-p34228`, using
  `make service-backed-test-slice OWNER=module.entities` with its existing
  `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`
  row. First fresh ordinary `make browser-e2e-visual` PASS, 12/12 units,
  `.cartulary/test-results/20260905T045554Z-p79535`; second fresh ordinary
  PASS, 12/12 units, `.cartulary/test-results/20260905T045817Z-p26921`.
  Both account for all 94 active goldens and 26 registered fixtures, with zero
  missing/orphan/ambiguous mappings. Every refreshed PNG was reviewed before
  these passes. No comparison threshold was relaxed.

- Final-source Timeline typing, blank-row creation, arrow-selection and
  Enter-focus measurement selection PASS, 19/19 units,
  `.cartulary/test-results/20260905T050046Z-p73641`. It ran without concurrent
  builds after the final editor geometry correction.
- CP-04 exit complete: duplicate-user production tabs, quiet actual heartbeat
  renewal across more than two lifetimes, removal/reconnect/reset/revocation,
  extension scope, all capacities, density/responsive/zoom/spacing geometry,
  quiet accessibility, state continuity, reviewed goldens and two fresh ordinary
  passes are complete. Next action: CP-05 finalizer and terminal validation.

#### Golden responsibility map

Before final refresh, the source audit found that wrapping the shared base
editor fieldset breaks the adapter's existing direct-child editor-geometry
selector (`packages/grid-adapter/src/styles.css`). Preserve that fieldset as
the direct child and put the presence layout inside it; do not duplicate its
geometry rules or change the adapter's vendor boundary. Add browser assertions
for the input staying inside its owning cell and separate from peer markers.

All filenames below live in `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
Rows and fixture identities are copied from the controlling reconciliation,
not inferred from filenames. Each of these nine refreshed PNGs was manually
reviewed after the Make-owned update: header/row/cell overflow and value context
are visible in the main capture; the second capture retains its cell marker;
base-grid captures show the reserved gutter; the narrow inspector capture
changes presence glyphs without changing inspector behavior.

| Golden filename | Authored semantic owner row | Registered fixture |
| --- | --- | --- |
| `collaboration-presence-markers-linux.png` | `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `visual.fixture.presence_overflow` (`D-VFIX-005`) |
| `collaboration-grid-presence-markers-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_deterministic_timelin_22b64f5dec` | Active nonregistry capture |
| `evidence-affordance-states-linux.png` | `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `visual.fixture.evidence_affordance` |
| `evidence-grid-available-evidence-linux.png` | `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af` | Active nonregistry capture |
| `evidence-grid-requested-evidence-linux.png` | `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af` | Active nonregistry capture |
| `evidence-grid-blocked-preview-linux.png` | `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830` | Active nonregistry capture |
| `record-relationships-evidence-access-linux.png` | `module.evidence.visual.the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc` | Active nonregistry capture |
| `record-relationships-task-requests-linux.png` | `module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc` | `visual.fixture.task_requests_or_decisions` |
| `workbook-inspector-narrow-technical-details-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_narrow_technical_details` |

### CP-03 evidence

- Server: added private live-record renewal, preserved observation and scope,
  broadcast renewed expiry using existing upserts, prevented late renewal after
  expiry/removal, and removed outgoing-traffic suppression from ping scheduling.
  Existing timeout constants remain unchanged. Incoming frames recheck the
  existing session deadline, incident access and closure before renewing.
- Client: canonical map retained; one generation-guarded expiry task derives
  all scopes. Visibility/focus restoration only recomputes time. Renewal-only
  changes retain the rendered snapshot. Old session callbacks and socket close
  events cannot restore invalidated projections. Accepted handshake connection
  identity is applied before React attachment catches up.
- Presentation: removed duplicated sorting and raw presentation arrays;
  Timeline and shared base surfaces consume scoped unique-user presentations.
  Header has a non-live accessible identity/count label. Row markers use
  readable token-backed single initials; cells reserve an inline marker slot.
  Cells without presence retain their previous layout. Editor slots preserve
  focus overflow. No detailed collaborator expansion was added.
- Added shared base row-gutter bindings and Generic/Entity editor lifecycle
  publication. Assessment has no writable grid editor in this checkout, so its
  row presentation was completed without inventing an editor or changing the
  inspector workflow. This adapts the plan to actual owner-backed capability.
- Added controlled expiry/renewal, stale callback, removal, sheet change,
  authorization loss, editor cleanup, and handshake identity tests. Corrected
  old shell fixtures with explicit time and replaced the backend test's invalid
  five-minute heartbeat gap with an in-lifetime pong.
- `make test-slice OWNER=web.collaboration` PASS (8/8 units),
  `.cartulary/test-results/20260905T034154Z-p9120`.
- `make test-slice OWNER=web.workbook` PASS (163/163 units),
  `.cartulary/test-results/20260905T034153Z-p8870`.
- Focused server semantic row PASS,
  `.cartulary/test-results/20260905T032838Z-p78854`.
- `make service-backed-test-slice OWNER=module.collaboration` with presence
  lifetime, two-real-client replay, and revocation rows PASS (5/5 units),
  `.cartulary/test-results/20260905T033948Z-p85588`.
- Initial typecheck failed on test fixtures still using raw arrays; migrated
  them to scoped presentations. Corrected typecheck PASS,
  `.cartulary/test-results/20260905T033844Z-p83475`.
- First `make format` failed on two returning `forEach` callbacks and three
  memo dependencies. Corrected loops and made row rendering depend directly
  on its presentation snapshot. `make format` PASS,
  `.cartulary/test-results/20260905T034235Z-p26832`.
- Final typecheck PASS, `.cartulary/test-results/20260905T034354Z-p32472`;
  import-boundary PASS, `.cartulary/test-results/20260905T034354Z-p32480`.
  CP-03 exit complete. Browser/golden passes are recorded under CP-04. No
  unrelated source edits resulted from formatting.

### CP-02 exit

- Added `workbookPresencePresentation.ts`: independent header/row/cell
  projection, structured duplicate-state keys, exact self/sheet filtering,
  precision-preserving UTC comparisons, NFC/code-point ties, unique-user
  capacities, next expiry, and presentation equality independent of expiry.
- Added capacities/priority/quietness to authored design presentation JSON,
  schema, and generator. `make generate` PASS, run
  `.cartulary/test-results/20260905T032340Z-p72743`.
- Pure presentation plus canonical transition rows PASS, run
  `.cartulary/test-results/20260905T032639Z-p76426`; nine presentation cases
  cover scope-local user collapse, overflow, Unicode/permutations, fractional
  observations, invalid inputs, exact identities, renewal equality, and expiry.
- `make test-slice OWNER=package.ui` PASS, run
  `.cartulary/test-results/20260905T032639Z-p76432`.
- Split the quiet component characterization into its own accessibility row;
  it was then a known failing CP-03 obligation, subsequently corrected. Legacy
  component helpers were removed when production consumers switched in CP-03.
- CP-02 exit complete; canonical snapshots remain keyed by exact connection.

### CP-01 exit

- Authored `workbookPresence.test.tsx` with four desired-behavior assertions
  before production edits: unique counts, activity priority, expiry, and quiet
  accessibility. Routed as `web.collaboration.regression.presence_presentation`.
- `make test-slice OWNER=web.collaboration ROWS=web.collaboration.regression.presence_presentation`:
  expected FAIL, all four assertions; run
  `.cartulary/test-results/20260905T032205Z-p70321`.
- Added a direct production `handleClientMessage` pong/hub test under the
  existing semantic hub owner row. `make test-slice OWNER=module.collaboration
  ROWS=module.collaboration.support_unit.semantic_hub_and_payload_ownership_d1e2f3a4b5`:
  expected FAIL for quiet presence expiring; run
  `.cartulary/test-results/20260905T032214Z-p70895`.
- Initial attempts stopped at catalog ASCII ordering checks (row and title
  arrays), before product execution. Corrected authored ordering; no check was
  weakened. These preparation failures had no completed run root.
- Missing base-grid row/editor bindings and incomplete capture were established
  by production source and manual baseline PNG inspection. Runtime clipping
  and heartbeat-under-broadcast verification remain CP-03/04 obligations.
- Applied the separately approved Core 01 timestamp clarification and narrow
  subordinate design reconciliation. No wire shapes or public modes changed.
- CP-01 exit complete. Next: implement the pure scoped projector and tests.

## Digest dispositions and acceptance

- ADOPT R001–R005, R009–R015: keyboard/accessibility, stable virtualization,
  shared tokens, state continuity, and owner-aligned evidence.
- ADAPT R016–R019, R021, R023, R035: desktop density/responsive geometry,
  quiet motion, existing icon provenance, and zoom/text-spacing resilience.
- REJECT R026–R034: retain rejection of cyberpunk, Matrix green, threat/alert
  theater, replacement palettes, marketing/dashboard layouts, glow, generic
  behavior authority, and incidental selectors.
- Other advice is not material to this seam. Digest files remain unchanged.
- A001–A003 ADOPT: authority/scope/baseline maps and reviewed source diff.
- A004–A006 ADAPT: existing dark graphite tokens and density registry retained;
  inherited fixed gutter geometry and font-relative marker spacing provide
  readable capacities. No new palette, theme, typography or density registry.
  Three-density browser geometry and reviewed images provide scoped evidence.
- A008–A009 ADAPT: existing responsive placement and scroll ownership retained;
  responsive boundary and presence zoom/text-spacing routes passed.
- A011, A014–A017 ADAPT: preserve existing continuity, editing, conflicts,
  refresh and authorization states. Workbook units, the real multi-client
  stateful scenario, editor identity/geometry and quiet-query assertions pass.
- A019–A023 ADOPT within this seam: named non-live identities/counts, stable
  scope selectors, production virtualization, density/spacing geometry and
  registered capture settings. Two ordinary visual passes and the UI package
  slice pass; final-source Timeline measurements also pass.
- A024–A027 ADOPT: no runtime/test/generator dependency on Markdown; authored
  projections and Make-owned outputs only; owner boundary/rollback maps and
  this handoff. Final drift/policy checks passed; closure includes the final
  handoff-byte and Markdown checks. These are implementation-support results,
  not a new Core 05 conformance or release claim.
- A007/A010/A012/A013/A018: regression boundaries only; no creation, inspector,
  transaction identity, recovery workflow, or Evidence behavior redesign.

## Compatibility and rollback

Existing presence lifecycle/presentation corrections only; no new public
interface, authentication behavior, durable collaboration semantics, stored-data
migration, dependency, or second socket. Detailed collaborator expansion is
omitted. No locking, authorization, save-state, or activity-audit meaning is
added.

## File responsibility and rollback boundary

- `apps/web/src/workbook/collaboration/workbookPresencePresentation.ts` owns
  pure scoped selection and lifetime-independent render identities.
  `workbookPresenceProjection.ts` retains canonical connection state;
  `WorkbookCollaborationCoordinator.ts` owns the single expiry task, session
  generations, selectors and base editor publication. Its React hook owns
  visibility/focus restoration listeners. `IncidentCollaborationSession.tsx`
  rejects old socket-close callbacks.
- `apps/web/src/workbook/utils/workbookPresence.ts` validates admitted records
  and supplies Unicode initials. `WorkbookPresenceMarkers.tsx`,
  `WorkbookStatusStrip.tsx`, `WorkbookShellTopBar.tsx`, and `workbookStyles.ts`
  own quiet accessible indicators and reserved geometry. Shared base gutters
  reuse the existing 58-pixel Timeline geometry; typography/colors use existing
  tokens, with no new theme or density registry.
- `AssessmentWorkbookSurface.tsx`, `EntityWorkbookSurface.tsx`,
  `GenericWorkbookSurface.tsx`, `WorkbookGridEditorControl.tsx`, and
  `models/workbookContractRows.ts` provide existing surface bindings and editor
  lifecycle. Timeline controller/renderers/models/compositions carry the same
  scoped presentation. No grid-adapter implementation or vendor import changed.
- `internal/modules/collaboration/hub.go` and `socket_session.go` own the
  demonstrated renewal/liveness corrections; their existing hub/session test
  files own backend regressions. No durable store or auth policy changed.
- The new presence test, coordinator/transition tests, Timeline presence and
  binding tests, and controlled shell fixture clocks own frontend regressions.
  Browser collaboration, accessibility, visual, Network Analysis and replay
  support changes own production verification. The visual fixture's additional
  resolved-name assertion only synchronizes existing fixture preparation.
- `contracts/design/presentation.v1.json`, its schema and design-presentation
  generator own typed capacities/priorities/quietness. Authored catalogs under
  `tools/test_families/`, `tools/frontend_source_ownership.json`, and
  `tools/frontend_visual_fixture_registry.json`
  own routing/capture declarations. Make owns the generated UI presentation,
  topology render index and golden manifest. The nine PNGs are enumerated in
  the reviewed golden responsibility map above.
- Only `docs/design.md`, the approved Core 01 clarification, and this handoff
  are documentation changes. The digest and completed-seam owner documents
  remain untouched.

Atomic source rollback must restore the audited changed tracked paths in all
these responsibility groups to baseline commit
`aa7c9611318a88c927019827129d8589a05f258e` together, and remove the three new
files: the pure presentation projector, `utils/workbookPresence.test.tsx`, and
this handoff. Restore the nine reviewed PNGs and generated manifest alongside
source, tests, authored projections/routing, generator/schema and generated
outputs; never restore only expiry filtering or only server renewal. Preserve
any unrelated changes added after this audit. Re-run owner verification and
generation/drift through Make. No database rollback or migration is required.
No atomic source rollback or commit has been performed.

Final scope audit: 60 paths (57 tracked modifications and the three new files
listed above), including exactly nine PNGs. Branch and commit remain the
recorded baseline. `git diff --check` passes; no digest, dependency/lockfile,
wire-contract, database, or grid-adapter source changes are present. All changes
fit the responsibility groups above, including the source-owner registry added
when its policy check identified the missing new files.

Full `make check`, release gates, and unrelated browser suites were not
selected for this bounded seam; verification used the affected owner routes,
the broad frontend target and full visual target. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset, as recorded by both finalizer results. No unsupported
conformance/release claim is made. CP-01 through CP-05 are DONE; there is no
next seam in progress. Closed-record Markdown and byte validation are the final
maintenance checks.
