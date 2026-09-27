# Timeline linked-Evidence review handoff

## Baseline, owners, and decision

The work began on clean `main` at
`ec6c3623201d0a85e892e90be3261dc9b5e015f9`. The root `AGENTS.md` was
the applicable repository instruction. The saved Timeline tag-removal handoff
was read as regression context; this change does not repeat that work. The
seeded editor baseline showed that Space opened an inspector without moving
focus to Evidence, while Manage Attached Evidence exposed a count and upload
control without saved-item actions. No Evidence handle was requested.

Core 01 §7.4.1 owns Timeline projection and collection identity; §16 owns
Evidence handles and fresh issuance. Core 02 §§13 and 18 own Evidence/blob
lifecycle and preview media. Core 03 §8.4, REQ-03-127/128, REQ-03-220, and its
inspector and continuity requirements own local preview, blocked outcomes,
shortcut focus, and retained authoring. Core 04 §2.0A owns current read access,
including viewers and closed incidents. `docs/design.md` guides presentation;
`docs/domain.md` supplies vocabulary. The UI/UX digest and historical handoff
are review aids, not executable authority.

Timeline's `attached_evidence_ids` projection supplies each item's `item_ref`,
`linked_record_id`, and `display_text`. The Evidence view has no preview
capability field or record-ID filter. The existing per-record preview-handle
route resolves capability for an explicit Space intent; no public endpoint was
added. One linked record or `evidence_count` alone does not establish
previewability. For multiple links, Space probes captured linked IDs in order,
one request at a time, stopping on a second previewable item or an
indeterminate result. Only defined Evidence-access blocker reasons count as
definitively blocked. A sole selected item receives a fresh handle; probe
handles are never redeemed. Ordinary rendering issues no probe requests.

The source owner is `web.workbook`; independent focused verification routes
through `web.workbook`, `module.timeline`, `module.workbook`, and
`module.evidence`. Timeline owns source/link identity and entry/return focus.
The shared Evidence access boundary owns handle invocation, typed feedback,
preview lifecycle, explicit download, retry, and access invalidation. The
attachment operation and retained authoring remain with their existing owners.

## Delivered behavior and source boundaries

- The Timeline Evidence section lists saved links by returned display text and
  stable target identity. Its available-empty and unavailable observations are
  distinct. The owner-defined `evidence_count` remains independent of the
  collection length. Existing Attach file and retained attachment recovery
  remain available.
- Each linked item has Preview and Download actions. Preview opens in the
  workbook; blocked or failed issuance reports local typed feedback without a
  download fallback. Download is a separate fresh issuance. Handle transport
  retains the `{}` body, returned identity and same-origin URL validation, and
  no `client_txn_id`. An internal `AbortSignal` cancels obsolete requests.
- Space is admitted for grid-owned focus on a committed selected Timeline row
  with an Evidence group. It opens the inspector, reveals and focuses the
  Evidence list or empty state, and opens a sole previewable target. Editor,
  menu, dialog, and IME ownership remains intact. Semantic focus registration
  fulfills first-open focus when the inspector mounts. Escape closes preview
  before the inspector and restores the list or invoking control; a semantic
  grid fallback remains available.
- Request and focus intents are scoped to account/session/incident read
  authority, source observation, link membership, exact Evidence target, and
  intent sequence. Duplicate activation, row or link changes, refresh,
  inspector closure, surface or incident exit, access loss, and late responses
  cannot retarget a request or reopen stale content. Slow access work does not
  reserve the grid or replay acknowledged writes. Raw Timeline drafts, grid
  selection, and scroll remain with their prior owners.
- The former count-only review copy, hard-coded
  `previewableEvidenceCount: 0`, and one-attempt inspector focus request are
  retired. The Evidence surface now uses the same small access controller as
  Timeline. No second authorization, retry, or preview state owner was added.

Authored source changes are concentrated in
`apps/web/src/workbook/timeline/{components,composition,focus,hooks,models,presentation}`,
`apps/web/src/workbook/features/evidence`, the Evidence handle command port,
`apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx`, and the narrow
workbook facade, shortcuts, and test fixture wiring. Focused unit and shell
tests changed under `apps/web/src/workbook`; service-backed browser scenarios
changed in `apps/web/e2e/evidence.spec.ts`. Authored verification routing
changed in `tools/test_families/{web.workbook,module.timeline,module.workbook,module.evidence}.json`.
`make generate` produced the corresponding
`tools/browser_e2e_batch_manifest.json` and
`tools/execution_topology_render_index.json`. No schema, stored data,
migration, public access route, or specification changed. There is no direct
grid-vendor import or upload redesign.

## Verification and rendered review

| Evidence | Result and run root |
| --- | --- |
| `make help`, `make help-all`, three owner `make task-guide` routes | Public and owner routing inspected before implementation |
| First focused red `make test-slice` | Failed on missing saved-linked-item actions as expected; `.cartulary/test-results/20260927T222927Z-p6571` |
| `make test-slice OWNER=web.workbook` for Evidence presentation/request lifetime, Timeline focus, keyboard, and resolution rows | Passed 6/6; `.cartulary/test-results/20260927T230448Z-p94546` |
| Final focused `make test-slice OWNER=web.workbook` for first-open focus and preview resolution | Passed; `.cartulary/test-results/20260927T230742Z-p5316` |
| `make test-slice OWNER=module.timeline` for semantic collection focus and keyboard decisions | Passed; `.cartulary/test-results/20260927T224705Z-p22107` |
| `make test-slice OWNER=module.workbook` for shortcuts and inspector selection | Passed 3/3; `.cartulary/test-results/20260927T232000Z-p35650` |
| `make service-backed-test-slice OWNER=module.evidence` for existing inline-safe preview plus new linked text and blocked-preview fixtures | Passed 11/11; `.cartulary/test-results/20260927T231853Z-p2430` |
| Focused linked-text browser row after preview and focus corrections | Passed 11/11; `.cartulary/test-results/20260927T232701Z-p19047` |
| Existing Evidence access accessibility row | Passed 11/11; `.cartulary/test-results/20260927T232531Z-p81061` |
| `make format` | Passed; `.cartulary/test-results/20260927T232651Z-p14602` |
| `make generate` | Passed; `.cartulary/test-results/20260927T230403Z-p61055` |
| `make agent-finalize` | Passed; `.cartulary/test-results/20260927T233101Z-p59392`. `RESULTS_DIR` was unset because no qualifying full warm check run was supplied; retained-run maintenance was skipped. |
| `make frontend-typecheck`, `make lint-biome`, `make frontend-import-boundary-check` | Passed; `.cartulary/test-results/20260927T233145Z-p63984`, `...-p64127`, `...-p64184` |
| `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check` | Passed; `.cartulary/test-results/20260927T233145Z-p63828`, `...-p63936`, `...-p64030` |
| `make lint-markdown` | Passed on the handoff; `.cartulary/test-results/20260927T233145Z-p64592` |

The new service-backed scenarios upload real synthetic text and blocked HTML
files. They verify exact Evidence target and fresh `{}` handle requests, in-place
preview, explicit download, blocked local feedback, Escape focus return, raw
inspector draft retention, long labels, both supported densities, and
1440×900 and 1024×720. Controlled promise tests cover zero, one, multiple,
mixed blocked/previewable, incomplete/failed capability observations, duplicate
and late work, source/link/authority changes, and cancellation. Shell coverage
includes viewer and closed-incident read access with attachment disabled.

Intermediate failed runs exposed test locator mistakes and an actual preview
Escape priority bug; both were corrected. They are retained at
`.cartulary/test-results/20260927T225417Z-p42253`,
`20260927T225659Z-p81803`, `20260927T230422Z-p64048`, and
`20260927T230747Z-p5803`. An earlier `make format` failure at
`20260927T225106Z-p27743` exposed hook-dependency and autofocus lint issues,
which were fixed. A seeded 1024×720 visual review exposed dark text on the
browser's text preview canvas, then a focused Preview button just below the
viewport after Escape. The shared Evidence preview iframe now requests the
light system document canvas, and return focus scrolls the invoking control
into view. The affected browser row passed again. Screenshot and automated
accessibility observations are advisory and separate from routed product-test
results.

The final seeded `make ui-review` session at
`.cartulary/test-results/20260927T232808Z-p52887/ui-review` reproduced the
first linked seeded Timeline row under editor access. With a committed grid
cell focused, Space opened the exact linked text preview and one Evidence
preview-handle request. At 1024×720, the text was readable on the light
document canvas; Escape closed only the preview and returned focus to the
visible Evidence list. A 1440×900 capture of the saved list and explicit
actions was inspected in the preceding seeded session. These are manual,
advisory observations. The final page axe scan completed with no violation
attributed to the linked Evidence controls. It did report a serious contrast
issue on the existing Tags empty label and a moderate missing page-level
heading, plus one incomplete contrast result; preview frames were excluded
from axe. No page-wide accessibility conformance claim follows from this
scan. All UI-review sessions were stopped and their foreground processes
exited successfully.

The full `make check` and release suite were not run because the selected
owner slices, service-backed browser and accessibility rows, and final source
and generation checks cover the changed boundary. No visual golden was
refreshed: the inspected change did not require an affected canonical golden.
The final `git diff --check` and scope review found only the focused authored
source, tests, routing, generated routing projections, and this handoff.

## Digest acceptance and compatibility

For this bounded seam, A001–A003, A010–A011, A014, A016–A020, and A023–A027
are supported by the owner mapping, source and routing review, focused state
and race tests, browser journeys, and rendered inspection above. A004 is
supported by the unchanged theme and density owners; the light system canvas
belongs to the browser-rendered Evidence document. A012–A013 concern replayable
writes;
handle issuance is a read-access request and does not acquire a transaction
ID or retry an attachment write. A005–A009, A015, A021–A022 concern unchanged
theme/chrome, creation, conflict, virtualization, or golden registries; no
global claim for those surfaces is made. The advisory page-level axe scan is
reported separately and is not a conformance claim.

Compatibility is unchanged: existing stored Timeline/Evidence rows and public
handle contracts continue to work, and no data migration is needed. Rollback
is the focused authored source, tests, routing inputs, and generated routing
projections. No commit, push, PR, or golden refresh was made.
