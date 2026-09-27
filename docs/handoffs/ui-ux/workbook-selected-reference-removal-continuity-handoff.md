# Workbook selected-reference removal continuity

## Scope and authority

Work began from clean `main` at `01ffd9847a94e51d538988e9d0caff74341731ce`,
ahead of assessment baseline `c4fe8d5e0`. The root `AGENTS.md` was the applicable
repository procedure. The localized UI/UX digest, its bundled upstream advice,
and the prior authoring-candidate-discovery handoff were read as advisory
navigation. Current source and adopted owners were checked independently. One
offline focus query, `keyboard focus removed control list`, supported R001 and
R002; R012 was adapted to the existing owner lifetimes. No documentation enters
runtime or verification.

Core 01 §§7.4 and 19 govern exact references and creation inputs; Core 02 §10
governs Assessment subject and support semantics; Core 03 §§2.3A and 16.4,
including REQ-03-304, govern retained authoring and keyboard continuity.
`docs/design.md` §§12 and 14 direct local focus and visible keyboard behavior.
Source ownership is `web.workbook`. Unit verification routes through
`web.workbook`, and browser verification through `module.workbook`. The authored
source and test manifests under `tools/` own their generated projections;
`tools/generated_artifact_policy.json` marks generated roots. No generated
output was hand-edited.

## Reproduction and source findings

At baseline, a Timeline contextual Task with three selected references from the
first Party page and one from the second page lost focus to `BODY` after each
native Tab then Enter/Space Remove action: first, middle, last, and sole.
Document and Timeline grid scroll offsets stayed at zero. The original parent
Linked Records Remove had the same unhandled button-unmount path in source; its
baseline browser focus capture was ambiguous and is not used as observation.
The initial focused
component regression failed with `BODY` rather than the next Remove control.
The passing baseline presentation slice did not assert removal focus.

Current source confirms separate state lifetimes: the common candidate list is
staged inside a picker; its Cancel leaves the parent unchanged and Apply updates
only the owning draft. The ordinary and Coordination parent list applies its
owner callback immediately. Contextual Task/Decision parent lists retain raw
form input and source seeds. Note source Clear means explicit unlinked source;
related Evidence Party removal retains independent Party text. Assessment
subject selection updates its parent directly, while Timeline support is staged.
Candidate paging, filters, and background reads do not own the selection.

## Correction and boundary

`useSelectedReferenceRemovalFocus` is one Workbook-local presentation helper.
It records only an explicit focused Remove intention: semantic record ID, prior
displayed ID order, source element, and target/authority scope. It calls the
existing owner callback and waits until rendered selection reflects the exact
removal. It then focuses the next surviving enabled Remove in the prior order,
otherwise the previous, otherwise the same field's enabled selector or Choose
trigger. The labelled field group is a programmatically focusable fallback
when an ordinary empty candidate page disables its selector; this preserves a
visible local focus destination without claiming a candidate exists. Authority
loss retires the intention and retains the existing concealment path.

The intention is retired on focus or input elsewhere, conflicting selection,
target or authority replacement, disabled presentation, and unmount. Candidate
responses and unrelated rerenders cannot create or replay it. Focus uses
`preventScroll`; the helper reveals the destination only in its nearest local
scrollport and stops before the Timeline grid. The group uses the existing
focus-ring token. Stable IDs handle duplicate labels and off-page selections.
No selection state, authoring store, global focus manager, server call, or
candidate read was added.

The common staged list, ordinary/Coordination parent list, contextual
Task/Decision parent list, Note source Clear, and related Evidence Party Remove
use the same decision. The previous unhandled button-unmount paths are retired.
Existing picker close/Apply/Cancel focus behavior, single/multiple limits,
metadata, order, raw fields, source context, omission/default and explicit
clear semantics, authorization, and disabled states remain with their owners.
Unrelated reference editors, pagination/retry, initial picker focus, and visual
restyling are outside this slice. No API, schema, storage, or data migration is
needed. Rollback is the focused source/test/route diff, followed by the public
generator; no data conversion is required.

## Evidence and verification

Source conclusions above come from current consumers and owner code. Browser
observations below are separate observations from a real Chromium keyboard run.

| Evidence | Result |
| --- | --- |
| Baseline presentation slice | PASS, `.cartulary/test-results/20260927T063301Z-p99388`; no removal assertion. |
| Baseline Timeline keyboard reproduction | FAIL as expected, `.cartulary/test-results/20260927T064017Z-p68398`; attached JSON recorded four `BODY` focus results and unchanged page/grid offsets. |
| Baseline focused component regression | FAIL as expected, `.cartulary/test-results/20260927T064258Z-p2226`; `BODY` instead of next Remove. |
| Exploratory browser setup | FAIL, `.cartulary/test-results/20260927T063734Z-p2501` and `.cartulary/test-results/20260927T063902Z-p35801`; initial Tab path and last-page option count in the new test were corrected before the focus reproduction. |
| Early corrected browser assertion | FAIL, `.cartulary/test-results/20260927T064630Z-p6664`; the test read only `aria-label` from the text-labelled parent Choose trigger. Focus was local; the attachment and assertion were corrected. |
| Corrected presentation component slice | PASS, `.cartulary/test-results/20260927T065100Z-p41401`. |
| Corrected Timeline browser slice | PASS, `.cartulary/test-results/20260927T065104Z-p42186`; next/previous/selector focus and unchanged page/grid offsets. |
| Contextual plus component unit slice | PASS, `.cartulary/test-results/20260927T065512Z-p9953`. |
| Frontend type check | PASS, `.cartulary/test-results/20260927T065513Z-p10028`. |
| Timeline plus Assessment browser slice | PASS, `.cartulary/test-results/20260927T065345Z-p76083`. |
| Narrow Timeline, Assessment, Note and related Evidence browser rerun | PASS, `.cartulary/test-results/20260927T070106Z-p71785`; real keyboard activation, Cancel/Apply, source clearing, parent removal, focus visibility and raw-field retention. Attached JSON records first/middle/last Remove focus, sole selector focus, page/grid offsets `0 → 0`, and Inspector offsets `3416 → 3416` or `3443 → 3443` for each removal. |
| Final compact inline plus named consumer browser slice | PASS, `.cartulary/test-results/20260927T070258Z-p12234`; all three authored rows passed, including the new Timeline scenario. Compact picker Remove, Cancel and retained raw cell input were observed. Its Timeline attachment repeated the same focus and scroll results above. |
| Final authority-state browser verification | Assessment and Note rows passed in `.cartulary/test-results/20260927T070639Z-p67325`. The Party row failed once because the contextual test saw one selected item after page-two selection instead of four; it did not fail a focus assertion. The selector was strengthened with explicit counts before and after paging. Two subsequent Party-row runs passed at `.cartulary/test-results/20260927T070854Z-p7266` and `.cartulary/test-results/20260927T071007Z-p40035`. The intermittent selection result was not reproduced or attributed to the focus helper; it remains a test/consumer stability limitation. |
| Initial `make format` | FAIL, `.cartulary/test-results/20260927T065542Z-p11150`: six new semantic-element/hook-dependency errors and one test warning; corrected in this slice. Unrelated unsafe-template suggestions were informational. |
| Intermediate type and Biome checks | FAIL, `.cartulary/test-results/20260927T065100Z-p41460` for overly narrow test state typing, `.cartulary/test-results/20260927T070106Z-p71841` for fieldset ref types, and `.cartulary/test-results/20260927T065854Z-p60061` for the new semantic/hook/test diagnostics. All were fixed; final checks pass below. |
| Final `make format` | PASS, `.cartulary/test-results/20260927T070844Z-p2858`. |
| `make generate` | PASS, `.cartulary/test-results/20260927T065608Z-p15771`; updated the execution topology render index from authored test rows. |
| `make agent-finalize` | PASS, `.cartulary/test-results/20260927T070614Z-p63169`; `RESULTS_DIR` unset. |
| Generation drift, generated policy and JSON shape | PASS, `.cartulary/test-results/20260927T070303Z-p16925`, `.cartulary/test-results/20260927T065849Z-p57800`, `.cartulary/test-results/20260927T065849Z-p57898`. |
| Test catalog check | PASS, `make test-catalog-check`; all authored titles and scenario IDs route. |
| Final Biome, type and import boundary | PASS, `.cartulary/test-results/20260927T071015Z-p49769`, `.cartulary/test-results/20260927T071015Z-p49679`, `.cartulary/test-results/20260927T070639Z-p67633`. |
| Markdown lint | PASS, `.cartulary/test-results/20260927T070515Z-p56665`; final handoff lint follows. |

The selected unit row's runner result lists all five new component regression
titles as passed; the final browser group and per-row results list every routed
scenario as passed in the successful browser runs. No required verification
remains blocked; the one intermittent Party-row failure is retained above.

### Digest acceptance mapping

| Row | Evidence |
| --- | --- |
| A001, A003, A024 | Exact adopted owners and source/test/generated boundaries above; digest stays advisory and outside executable dependencies. |
| A002, A026 | One shared presentation decision across existing consumers; no new public contract, migration, or unrelated editor scope. |
| A007, A011, A014 | Focused staged Cancel/Apply, parent removal, source-clear, raw-field, retained identity and off-page checks. |
| A009, A019, A020 | Native keyboard browser focus, token-backed fallback and narrow long-label viewport; page/grid scroll comparisons. |
| A017 | Target/authority retirement and concealed/disabled behavior in focused tests. |
| A025 | Authored routing/source inputs and public `make generate`; generated index is a projection. |

`RESULTS_DIR` was unset because no qualifying successful full warm run was
provided; retained-run maintenance was skipped. Broad full warm suites and
visual golden maintenance were skipped because the focused source, browser and
static routes covered the changed behavior and no visual restyling occurred.
No commit or publication was performed.
