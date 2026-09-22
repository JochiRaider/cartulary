# Workbook visual overhaul, stages 1–2

Status: implementation and technical verification complete; human design review
and participant evaluation remain outstanding. This is a handoff, not a behavior
owner or product acceptance report. Baseline: main at
`d28304bec4d28a6282b0fffcab8ca5312ab4e6d1`, initially clean, 2026-09-21.

## Decisions and boundary

The user approved the stages 1–2 specification: composed incident and working-set
identity, a conditional desktop query strip, grid browsing footer, quiet commands,
and collapsed saved-view Startup controls. Timeline defaults and Reset remain
unchanged; Capture and Review are fixture-local studies. Inspector, recovery,
query admission, transaction identity and draft lifetimes remain with their
existing owners. No data migration or new package was introduced.

The common presentation decisions are command emphasis and allocation of shell
space. Existing feature controllers continue to decide command existence,
eligibility and execution. Replaced local style consumers now use the shared
treatment. No generic form or retry engine was introduced.

Owners: Design §§7.1, 7.4–7.5, 8.3, 12 and 14; Core 03 §14.9 for browsing and
§§3–4 for write versus read recovery. Source: workbook layout, components and
query presentation. Independent verification: web.workbook, web.design and
package.ui; browser rows also route through module.workbook/module.savedviews.

## Stage 1

Production-renderer study attachments compare current, Capture and Review at
1440×900, 1280×720, 1024×720 and 768×640. They use the same synthetic rows and
production Columns commands, with explicit top/left scroll anchors. They neither
change registry defaults nor create new saved-view identities. Existing pressure,
conflict, attached-draft and extension fixtures supply complementary states.

Initial isolated browser run failed before tests because `node` was absent from
PATH: `.cartulary/test-results/20260921T214633Z-p66004`. Use the existing pinned
`tmp/node-runtime/bin` on PATH with public Make targets. No harness policy was
changed to work around this environment issue.

The baseline run passed all 11 harness units at
`.cartulary/test-results/20260921T214914Z-p99910`. Twelve study images are retained
in `.cartulary/design-studies/workbook-overhaul/baseline/`. Agent inspection of all twelve baseline plates confirms narrative-leading layouts
put more narrative fields in the first viewport but do not resolve cramped shell
identity or query hierarchy. All twelve implemented plates were also inspected;
the separate query row and wider identity allocation are visible at desktop bands,
and compact widths retain one query-details route without a stacked toolbar. Keep the approved current-default policy. The same populated state
shows the existing long-label and modified-view pressure. Human usability and
image approval remain outstanding; these attachments are advisory studies.

Stage 1 technical exit: studies captured, governing amendments authored, approved
column/default and query-placement policy recorded. Continue with the user's
already-authorized presentation implementation; no participant result is claimed.

## Stage 2 and verification

Implemented boundaries:

- Design §§7.1/7.5/8.3/12 now declare spare-width incident allocation, saved-view
  minima, conditional chip strip, grid footer and quiet commands. Core 03 §14.9
  explicitly permits conditional continuation controls while preserving pending
  and focused controls.
- `contracts/design/tokens.v1.json` owns new dimensions and quiet-button values.
  `presentation.v2.json` owns chip capacities and initially collapsed Startup;
  its schema and generator project these through existing UI facades.
- `WorkbookSurfaceLayout` owns the strip destination and grid footer placement.
  `WorkbookQuerySummarySlot` moves presentation through a portal without moving
  the query reducer or command ports. The inspector remains on its existing
  shared work area, including its existing overlay and resize behavior.
- `WorkbookIncidentIdentityDisclosure` exposes complete labels with pointer,
  keyboard, Escape and ordinary focus departure. The key takes precedence over
  the title; account navigation remains bounded at the right edge.
- Existing shared form/query styles supply quiet commands for query tools,
  Find/Clear, filter removal, saved-view commands and browsing recovery.
  Replaced local button recipes and inline-chip allocation recipes were removed.
- `SavedViewActionPanel` collapses Startup presentation while retaining operation
  summaries and existing controller-owned confirmation/recovery. Existing helper
  selectors and browser choreography open the disclosure explicitly.
- `WorkbookQueryBrowsingControls` still uses the same browser/controller ports.
  The footer reports committed loaded count during reads, retains pending/focused
  continuation controls and removes exhausted controls only after focus leaves.

No direct vendor integration, schema default, request port, draft store, public
API, persistence format or execution/permission owner changed.

### Verification record

All commands ran at repository root through public Make targets with
`PATH="$PWD/tmp/node-runtime/bin:$PATH"`. Run roots below are beneath
`.cartulary/test-results/`.

| Command or owner selection | Result and evidence |
| --- | --- |
| `make generate` | PASS, `20260921T215150Z-p33406`. Generated facades and topology index were generated, not hand-edited. |
| `make test-slice OWNER=package.ui` | PASS, 10/10 units, `20260921T215850Z-p44442`. |
| Initial focused `web.workbook` rows for saved-view focus, query components/model, preferences, browsing and responsive layout | PASS, 7/7, `20260921T215850Z-p44397`. |
| `make test-slice OWNER=web.workbook` | 292/294 passed, `20260921T220609Z-p60109`; two obsolete presentation expectations were updated. Surface/startup test passed in `20260921T220900Z-p27380`; inspector row passed in `20260921T221253Z-p72849`. |
| Final query model/component and browsing rows | PASS, 4/4, `20260921T222044Z-p50484`. |
| `web.design` global accessibility and workbook frame geometry | PASS, 13/13, `20260921T221254Z-p73070`. |
| `module.savedviews` query-control accessibility and independent discovery continuity | PASS, 11/11, `20260921T221614Z-p51254`; includes text spacing and chip-to-Filters resize focus. |
| `module.workbook` shell accessibility with incident disclosure | PASS, 11/11, `20260921T222045Z-p50705`. |
| Inspector accessibility, Timeline/Notes continuation, live read recovery and preference uncertainty | Selected functional rows PASS in `20260921T221613Z-p51035`; this aggregate failed because visual comparisons were intentionally stale and the preference inspector shortcut still expected expanded Startup. |
| Preferences accessibility, remaining preference browser rows, continuation authorization and session transitions | PASS, 15/15, `20260921T221843Z-p64857`, including the corrected shortcut disclosure. |
| Four `module.timeline.measurement` rows: typing acknowledgement, blank-row creation, ArrowDown and Enter paint | PASS, 20/20 units, `20260921T221615Z-p51772`. This is local owner-selected measurement evidence, not a published benchmark or usability improvement. |
| `make frontend-typecheck frontend-import-boundary-check lint-biome` | PASS, final roots `20260921T221844Z-p65466`, `20260921T221844Z-p65478`, `20260921T221844Z-p65508`. |
| `make generate-drift` | PASS, 4/4, `20260921T220902Z-p27902`. |
| `make agent-finalize` before broader browser verification | PASS, `20260921T215923Z-p47307`; retained-run maintenance skipped because RESULTS_DIR was unset. |

The ordinary visual run `20260921T220859Z-p27254` reconciled 252 captured
intentions against 255 committed goldens with no missing or ambiguous mapping.
Three default-shell captures had not executed because their inspector-bound
assertion still compared against the shortened scrollport rather than the full
grid slot. The corrected ordinary row in `20260921T221613Z-p51035` captured those
three. All 255 paths are accounted for across these ordinary artifacts; none was
retired or deleted as an orphan. The remaining failures were expected image
comparisons after the adopted presentation change.

Earlier diagnostic failures were resolved: missing bare `node` on PATH; an
unused import; semantic-element lint; incorrect nested work-area grid row; tests
assuming always-expanded Startup; and a browser build correctly rejecting inputs
changed by a simultaneously running formatter. Formatter/build operations were
subsequently sequenced. No comparison tolerance, masking or retry was broadened.

The complete `web.design` slice passed 19/19 units at
`20260921T222457Z-p89068`. The final Sort-menu style correction then passed the
saved-view accessibility scenario in `20260921T222713Z-p1055`; its ordinary
visual row failed only the intentionally changed Sort image. That image was
inspected at full size before the final canonical refresh.

Two in-flight ordinary comparisons of the previous visual revision were stopped
when image review found the remaining Sort outlier. Their retained-artifact
finalizers rejected unfinished trace resources (`20260921T222454Z-p88395` and
`20260921T222455Z-p88601`). These interrupted runs are not acceptance evidence;
no artifact policy was weakened. Fresh complete runs passed below.

The final Sort component row passed at `20260921T222827Z-p37323`; final
TypeScript passed at `20260921T223147Z-p78943`, lint at
`20260921T222828Z-p37587`, and repeated generated drift at
`20260921T223027Z-p74827`. Final Markdown lint passed at
`20260921T224000Z-p60653`.

`make generated-artifact-policy-check` and `make json-shape-check` passed at
`20260921T222714Z-p1684` and `20260921T222714Z-p1686`. Finalization was repeated
successfully before the final comparison phase at `20260921T222316Z-p84062`.

See [the complete golden refresh record](workbook-visual-overhaul-golden-review.md)
for all 172 changed filenames, all 36 affected catalog rows and registered fixture
identities. Agent review covered every changed image; no human approval is claimed.
The final canonical refresh passed 12/12 units at `20260921T222849Z-p38514`: all
47 scenarios passed, with 255 captures/goldens, zero orphans, missing paths or
ambiguous mappings. Finalization passed again at `20260921T223346Z-p84434`
before launching the two final ordinary comparisons. Both
`make browser-e2e-visual` runs passed 12/12 units in fresh roots
`20260921T223437Z-p88272` and `20260921T223438Z-p88479`. No comparison tolerance,
mask, viewport, screenshot scope or retry was relaxed.

Final catalog, generated-artifact policy and JSON shape checks passed at
`20260921T223601Z-p58188`, `20260921T223601Z-p58034` and
`20260921T223601Z-p58036`, respectively. Final `git diff --check` passed.

The task-local [before/after study gallery](../../../.cartulary/design-studies/workbook-overhaul/review.html) presents all 24 plates at their declared sizes.

Human usability findings are not yet available. Retained screenshots and agent
inspection must not be described as participant testing or human golden approval.

Advisory dispositions: adopt R002/R013/R014 focus, semantics and stable navigation;
adapt typography/navigation advice through existing roles and owned regions;
reject literal body-offset and marketing-navigation recipes. The narrow offline
query returned four results without fallback.

## Compatibility and rollback

No API, field key, saved-view grammar, persisted layout or default-column migration.
Roll back owner amendments, typed projections, generated facades, presentation and
reviewed goldens coherently. Do not rewrite retained drafts or captured requests.
No indefinite dual implementation is retained.

## Digest acceptance assessment

These dispositions assess this presentation slice, not the entire product.

| Rows | Disposition and evidence |
| --- | --- |
| A001–A003 | PASS: owner/source/verification mapping above; current source and task guides supersede the older digest baseline. The common decisions are command emphasis and shell allocation. |
| A004–A005 | PASS: existing graphite theme, typography and icon family; new dimensions/quiet-command values are authored tokens, projected by existing generators. |
| A006, A008–A009 | PASS: shared density unchanged; responsive/frame accessibility, text-spacing checks and four paint measurement rows pass. Grid and inspector retain their scroll ownership. |
| A007 | PASS within the unchanged creation boundary: full workbook regression evidence and blank-row paint measurement; no new capture prerequisite. |
| A010–A014 | PASS within affected presentation boundaries: inspector/focus regression rows, independent saved-view discovery, preference detachment, uncertain operation and read-recovery scenarios pass. Request/identity and editor owners are unchanged. |
| A015–A018 | Functional PASS: existing full workbook regression rows and stateful continuation/recovery/authorization cases; visual review is separately tracked below. Evidence and conflict execution remain unchanged. |
| A019–A021 | PASS: owner-selected keyboard, contrast, text spacing, focus, continuation and paint measurement evidence above. No new virtualization or permission logic. |
| A022 | Technical PASS: canonical golden refresh, complete agent image review and two fresh ordinary visual passes. Human design approval remains outstanding and is not inferred from automated comparisons. |
| A023–A026 | PASS: semantic selectors, package.ui, import boundaries, generated drift and source review; no executable dependency on Markdown. No migration or duplicated implementation. |
| A027 | PASS: implementation, verification roots, complete golden refresh inventory, unresolved human evaluation and rollback boundaries are recorded in this handoff. |

Retained-run maintenance is skipped because RESULTS_DIR is unset; no successful
full warm check was supplied. Release/conformance suites are outside this visual
slice and no release or Core 05 acceptance is claimed.
