# Workbook workbench implementation

Status: implemented and verified through the routed checks below; ready for review.
This record is implementation evidence, not behavioral authority.

## Scope and owners

The implementation applies the adopted redesign in Core 01 REQ-01-680–684,
Core 03 REQ-03-311–318, Core 04 REQ-04-169–170 and AC-572–597, and Design 0.6.0.
Work began on `main` at `886750d965278a32452dbbeced77f55910f758f5`. The existing
nine document-adoption edits were preserved. No commit or deployment was made.

| Decision | Behavioral owner | Implementation boundary | Verification routing |
| --- | --- | --- | --- |
| Bounded, authorized record location and ordinary cursor continuation | Core 01 REQ-01-680–684; Core 04 REQ-04-169–170 | Workbook transport, querypage algorithm, source-owned read transactions | `module.workbook`, `platform.viewquery` |
| Shell geometry, one auxiliary destination, full source reader | Design 0.6; Core 03 workbench requirements | Authored design projections; shared workbook presentation | `web.workbook`, `module.workbook`, `package.ui` |
| Commands and five Timeline presets | Core 03 REQ-03-311–318 | Command index with source-owned action bindings; existing query owner | `web.workbook`, `module.workbook` |
| Session pins, staged pivots and bounded Return | Core 03 REQ-03-311–318; Core 04 protected-state requirements | Incident mutation-runtime session metadata; browsing registry; source selection callbacks | `web.workbook`, `module.workbook` |
| Extension destination composition | Existing extension availability owner plus workbench requirements | Injected work-area frame; existing Network Flow lifecycle | `module.networkflow` |

The common decisions are destination placement, command discovery, and bounded
navigation admission. Source mutation, editor, conflict, authoring, and evidence
owners remain responsible for their existing behavior. Adding an action requires
an owner binding; adding a record surface requires its declared view contract and
query/locator provider. No generic record-store API, second task model, grid
vendor migration, workflow rewrite, database migration, or persistent session
archive was introduced.

## Delivered behavior

- The incident/view/footer bands are 40/40/32 CSS pixels. The footer contains
  primary sheets and grouped More views; save state remains separately labeled.
  Record, Work and Recovery share one destination. Desktop adjacency preserves
  the grid minimum; narrower overlays make the covered grid inert and contain
  keyboard focus. The dock starts closed.
- Commands use deterministic metadata matching, family ordering, 20-result
  paging, captured semantic targets and invocation-time revalidation. Commands
  do not search incident content or introduce a global shortcut. Query controls
  use compact menus at narrower widths and retain their drafts during resizing.
- Full committed source reading preserves RAW whitespace, offers exact Copy,
  local soft wrapping and owner-declared Edit. Mention, reference and Note
  navigation uses explicit admitted record identities and target surfaces.
- Work contains a 20-pin, insertion-ordered session working set and links to the
  existing coordination surfaces. Duplicate pins are detected before capacity;
  full capacity refuses insertion without eviction.
- Return retains at most 32 metadata anchors, independent of the existing
  20-checkpoint/three-page query browser. Record-directed navigation stages its
  read before replacing context. Ordinary surface selection retains its existing
  creation-first behavior, browsing checkpoint and accepted-write recovery.
  Located records are selected through the source owner before inspection.
- The locator distinguishes `located`, `outside_query`, and `unavailable` without
  exposing inaccessible record existence. It validates the closed request and
  current authorization before lookup. Located windows contain at most 100 rows
  and issue normal query cursors using the preceding tuple. Source providers use
  a consistent read-only transaction; no browser scan or hidden prefetch occurs.
- Pins and Return are memory-local. Authorization uncertainty conceals protected
  state; confirmed incident exit, access loss and account replacement retire it
  under the existing owners. Detaching presentation does not cancel admitted
  writes or acknowledge unresolved work.

## Compatibility and correction record

The locator is an additive OpenAPI change. Authored OpenAPI, error, HTTP-operation,
design and verification projections were updated before their generated outputs.
Existing source fields, write routes, transaction identities, retained draft
lifetimes and density choices remain in their owners.

Rendered and regression review found and corrected: ordinary sheet entry losing
its checkpoint; Return focusing a cell without selecting its inspector record;
responsive menus remounting filter drafts; overlay width inheriting the adjacent
45-percent clamp; a focus trap omitting native disclosure controls; and nested
query panels extending outside the viewport at zoom. Obsolete top-bar/System
views test entry paths were migrated to the footer/More views path. Assertions
for source writes, replay, saved content and recovery were retained.

The optional legacy Note lookup path remains for standalone presentations without
a workbench provider; the application path uses the bounded locator. It is not a
second authoritative navigation store. No persisted data migration is necessary.
Rollback consists of reverting this implementation and its generated projections
and approved goldens together while retaining the independently adopted owner
documents for an explicit follow-up decision.

## Verification evidence

All paths below are repository-relative retained run roots. Failed attempts are
not evidence of a pass; superseding checks are identified explicitly.

| Check | Result and retained evidence |
| --- | --- |
| `make generate` | PASS, `.cartulary/test-results/20261007T051444Z-p85694` |
| Locator integration slice, including all registered surfaces, grouped/tied Hosts, null Timeline time, cursor scope and strict validation | PASS, 3/3 units, `.cartulary/test-results/20261007T040003Z-p45177` |
| Navigation admission, semantic grid entry and query controls slice | PASS, 4/4 units, `.cartulary/test-results/20261007T041600Z-p37945` |
| Network Analysis discovery/import slice | PASS, 11/11 units, `.cartulary/test-results/20261007T041559Z-p37664` |
| Note pivots, authoring focus, detached accepted-write recovery, column controls | Functional rows PASS in `.cartulary/test-results/20261007T041601Z-p38220`; its column accessibility row failed and was corrected |
| Column keyboard/zoom and saved-view layout rerun | PASS, 13/13 units, `.cartulary/test-results/20261007T042454Z-p23032` |
| Frontend typecheck | PASS, 2/2 units, `.cartulary/test-results/20261007T051544Z-p94661` |
| Script lint | PASS, 2/2 units, `.cartulary/test-results/20261007T050700Z-p744` |
| Biome | PASS, 2/2 units, `.cartulary/test-results/20261007T051544Z-p94693` |
| Frontend import boundary | PASS, `.cartulary/test-results/20261007T035002Z-p13812` |
| Backend module boundary | PASS, `.cartulary/test-results/20261007T035127Z-p17096` |
| `make agent-finalize` | PASS, `.cartulary/test-results/20261007T051655Z-p36118`; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| Full frontend owner regression | 316/318 units passed in `.cartulary/test-results/20261007T042631Z-p6116`; both failed rows plus navigation admission PASS in the 4/4-unit rerun `.cartulary/test-results/20261007T043341Z-p85642` |
| Shared UI contracts | PASS, 10/10 units, `.cartulary/test-results/20261007T043407Z-p7145` |
| View query owner | PASS, 18/18 units, `.cartulary/test-results/20261007T043408Z-p8618` |
| Enum-filter responsive rerun | PASS, 11/11 units, `.cartulary/test-results/20261007T043342Z-p85951` |
| Generated drift | PASS, 4/4 units, `.cartulary/test-results/20261007T052043Z-p26288` |
| Generated policy | PASS, 3/3 units, `.cartulary/test-results/20261007T052043Z-p26277` |
| Markdown lint | PASS after the final evidence update, `.cartulary/test-results/20261007T052716Z-p50688` |
| Full workbook owner baseline | 87/109 units passed, `.cartulary/test-results/20261007T035002Z-p13696`; exposed selector, navigation, geometry and expected golden differences; not final acceptance |
| Coordination/grid accessibility, Timeline clear, held filters and History continuation | Routed rows PASS, `.cartulary/test-results/20261007T041649Z-p12373` |
| Final surface-navigation corrections | System views, delayed entry and coordination navigation PASS in `.cartulary/test-results/20261007T043839Z-p33699`; its remaining obsolete footer focus expectation was corrected, then desktop keyboard row PASS, 11/11 units, `.cartulary/test-results/20261007T044311Z-p26645` |
| Frozen-column query dismissal | PASS, 11/11 units, `.cartulary/test-results/20261007T043915Z-p70683` |
| Query browser attempt during source editing | Not executed: source-snapshot protection rejected changed build inputs, `.cartulary/test-results/20261007T042550Z-p65173`; superseded by the successful sealed query/column reruns above |
| Commands interaction: target revalidation, bounded paging, retained draft/focus, authority concealment | PASS, 2/2 units, `.cartulary/test-results/20261007T051544Z-p94555` |
| Commands responsive admission, hit testing, focus, explicit dispatch and axe | PASS, 11/11 units, `.cartulary/test-results/20261007T051543Z-p94081` |
| OpenAPI compatibility | PASS, 4/4 units, `.cartulary/test-results/20261007T045217Z-p57523` |
| JSON shape and visual fixture registry | PASS, 3/3 units, `.cartulary/test-results/20261007T052043Z-p26306` |
| Authored test catalog | PASS, 1/1 unit, `.cartulary/test-results/20261007T052043Z-p26392` |
| Final ordinary visual comparison 1 | PASS, 12/12 units, `.cartulary/test-results/20261007T051745Z-p41032` |
| Final ordinary visual comparison 2 | PASS, 12/12 units, `.cartulary/test-results/20261007T051745Z-p41039` |
| Whitespace/diff validation | `git diff --check` PASS |

Earlier typecheck/format failures from duplicate imports and menu accessibility
semantics were corrected. A misnamed backend target was replaced by the public
`backend-module-boundary-check` target. Private UI-review attempts that failed
capacity or fixture preparation were closed and are not rendered evidence.

## Rendered review and visual refresh

The initial desktop baseline and subsequent default-profile review used the
canonical Make UI-review harness. Review covered the compact shell, Work at
1024 pixels, full RAW reading, pinning, task navigation and Return. The review
at `.cartulary/test-results/20261007T035943Z-p34555/ui-review/session.json` found
the selection and overlay-width defects subsequently corrected. Its three
consumed captures reported zero automated accessibility violations; that does
not establish whole-application accessibility. The session closed successfully.
Private screenshots were inspected during the session and are not durable links.

A fresh sealed review at
`.cartulary/test-results/20261007T042610Z-p84107/ui-review/session.json`
confirmed pin → Task Requests → pin → Return → Return restores the original
Timeline MITRE cell and the correct selected inspector record. The 1024×720
capture shows an inspector from x=464 to x=1024, satisfying the 560-pixel overlay
width. Captures 14 and 17 were digest-verified and visually consumed; both had
zero axe violations and zero console records (capture 14 had one incomplete
contrast rule; capture 17 had none). The session closed successfully.

The accepted visual refresh trigger is the adopted shell, footer, inspector and
menu redesign. Browser pins, fonts, zoom, masks, scroll normalization and capture
scope are not being changed to hide differences. The ordinary visual
reconciliation in `.cartulary/test-results/20261007T035002Z-p13696` reported zero
missing goldens and zero ambiguous mappings; its selected visual target failed
pixel comparison. No orphan deletion is authorized by that partial run.
The first transactional update passed 12/12 units in
`.cartulary/test-results/20261007T044457Z-p76167`. Reconciliation accounted for
255 active captures and goldens, 29 registered fixtures, zero missing goldens,
zero ambiguous mappings and zero orphans. It refreshed 232 goldens.

All 232 candidate image digests were inspected through three private artifact
sessions. The review found one remaining defect: the lifecycle label wrapped and
clipped inside the 40-pixel band at 390 pixels. The label now stays on one line
without flex shrinking; the owning visual scenario additionally checks that its
full height lies within the band. The first ordinary comparison passed 12/12 units in
`.cartulary/test-results/20261007T050041Z-p98896` against that initial manifest.
Because the clipping correction followed, this pass is a baseline rather than
final acceptance. The corrective transactional update passed 12/12 units at
`.cartulary/test-results/20261007T050710Z-p2162`; only the lifecycle golden changed
from the initial refresh. Its final bytes were reimported and inspected: the full
lifecycle label is visible within the incident band. Two fresh ordinary comparisons
against the same corrected manifest passed 12/12 units each at
`.cartulary/test-results/20261007T051745Z-p41032` and
`.cartulary/test-results/20261007T051745Z-p41039`. Both reconciliations verify the
same final manifest SHA-256 and account for all 255 active goldens. The
[visual refresh record](workbook-workbench-visual-refresh.md) lists all 232 changed
filenames, exact catalog owner rows, fixture IDs and final manifest digest.

The administration and directory changes follow from their existing use of the
shared top-bar minimum-height token. An explicit old/current administrative audit
comparison showed the header/content origin moving from 64 to 58 pixels while
preserving the controls, labels and data. This reviewed shared-token effect is
included in the refresh scope. Renderer identity, fonts and capture declarations
are unchanged. Artifact-only imports provide pixels, not DOM or accessibility
evidence. A subsequent seeded Commands review found that the below-minimum header
scrollport clipped the absolute-positioned command panel. Commands now uses
viewport positioning with the shared zoom-aware menu-bounds logic. A browser row
checks retained search focus, unobstructed hit testing, viewport bounds, explicit
dispatch and automated accessibility at desktop, 390 pixels and 200-percent zoom.
The browser row passed 11/11 units at
`.cartulary/test-results/20261007T051543Z-p94081`. Its three original PNG
attachments were imported with digest verification and visually consumed at
1440×900, 390×480 and 1280×900 with 200-percent zoom. The search field, result
action and paging/cancel controls remain visible without clipping in all three.
These screenshots confirm presentation; the browser assertions supply focus,
hit-testing and dispatch evidence.
No interactive HTML report was consumed. Artifact sessions closed with successful
cleanup receipts; no private screenshot paths are retained in this handoff.

## Acceptance assessment

| Digest rows | Current assessment |
| --- | --- |
| A001–A003 authority, scope and repository boundary | PASS: adopted owner map, current implementation/projection inspection and preserved initial dirty state above |
| A004–A005 tokens and theme | PASS: authored design projections and generated facade; existing graphite theme and icon family |
| A006, A008–A009 density, responsiveness and overflow | PASS: density/viewport/zoom matrix, focused responsive browser checks, and both final visual comparisons |
| A007 creation | PASS for the routed ordinary/contextual and detached recovery checks above; no source-write change |
| A010–A011 inspector and continuity | PASS: focused navigation tests and fresh rendered record-selection confirmation above |
| A012–A017 transactions, recovery, editing, conflict, query and authorization | PASS for the routed checks: all initial nonvisual workbook failures have successful corrective reruns; the full original run itself remains failed |
| A018 evidence | PASS for retained source/evidence behavior and reviewed evidence states in the final visual matrix |
| A019–A022 accessibility, compound states, virtualization and visuals | PASS for routed keyboard/axe, query-window and compound-state checks plus reviewed final visuals; this is not a whole-application accessibility certification or analyst-study result |
| A023 selectors | PASS: stable surface/record/field identities and 10/10 shared UI verification units |
| A024 test authority | PASS by dependency/diff audit: no executable checks read Markdown |
| A025 generated outputs | PASS: generation, drift and generated policy checks above |
| A026 compatibility | PASS: additive read interface, preserved source ownership, explicit no-migration and rollback decision |
| A027 handoff | PASS: owner mapping, correction record, exact visual inventory, verification evidence, compatibility/rollback decision and limitations recorded |

The eight-analyst formative study requires human participants and has not been
conducted. No timed, fixture-sensitive public performance or release-conformance
claim is made. Deployment and release checks are outside this implementation run.
