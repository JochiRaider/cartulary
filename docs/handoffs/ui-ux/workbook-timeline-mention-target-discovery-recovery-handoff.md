# Timeline mention target discovery recovery

## Scope and authority

Execution began on clean `main` at `15287ac8d72a6a96018eaeade06d7c2c382d68f9`.
The root `AGENTS.md`, digest README, START_HERE, LOCAL_AGENT_PROMPT,
OWNER_MAP, REPO_MAP, QUERY_RECIPES, rules and acceptance were read as
navigation. Current source and authored test-family inputs were rechecked.
No unrelated change was present; this work remains uncommitted.

Core 01 §3.3.7 owns authoritative opaque query paging and query failures. Core
03 §§2.3A and 9 own the Inspector mention route and explicit action; REQ-03-299
and REQ-03-100 own authorized observation retention and authority loss. Design
§§10.8, 12.4, 12.7–12.8 and 14 direct local feedback, pending controls, focus
and reflow. Domain supplies vocabulary and design supplies direction within
their boundaries. No governing-owner contradiction was found. The research
NLSpec and digest are advisory, not runtime or test inputs.

`web.workbook` owns Timeline source and the focused unit rows. Independent
service-backed browser routing belongs to `module.entities`. The generated
browser batch manifest and topology index are derived from authored
`tools/test_families/{web.workbook,module.entities}.json` through `make generate`.
No generated root was edited by hand. The narrow offline QUERY_RECIPES search
for keyboard focus and async feedback adopted R002/R008/R010 local focus and
recovery advice, adapted R012 to existing shell ownership, and rejected R033
as an unrelated framework prescription. No upstream design system was generated.

## Reproduction and boundary decision

The existing candidate slice passed at
`.cartulary/test-results/20260926T191620Z-p70285`. Source inspection showed
that Retry disappeared while its read loaded, Load more became natively
disabled, a repeated cursor demanded a reload action with no UI, and Retry
reconstructed its cursor from mutable `nextCursor`. These were source
observations before reproduction.

Failing unit evidence at `.cartulary/test-results/20260926T192738Z-p75524`
showed a failed cursor-free restart retrying the old `page-2` cursor and a
pending Retry disconnecting. The production Inspector browser row at
`.cartulary/test-results/20260926T193140Z-p15510` independently reproduced
the disappearing keyboard Retry for Host and Identity while the continuation
response was held. The browser and unit fixes were made after those red runs.
The earlier `.cartulary/test-results/20260926T192620Z-p74156` failed due to an
unstable inline test port and is not product evidence.

The selection rubric finds one bounded owner decision: Timeline needs
accumulated accepted pages and failed-read identity in the same read owner.
The shared `WorkbookCandidateDiscovery` navigates single pages and does not
retain Timeline's loaded eligible set; sharing that state would change the
required selection model. Reuse remains at typed failure/authority boundaries.
This keeps later Timeline pagination extensions local without duplicating
candidate eligibility or creating a generic discovery framework. Remaining
risk is that a future shared paging change must be considered separately for
both models.

## Implementation and compatibility

- `useTimelineMentionCandidates.ts` owns accepted current-chain candidates,
  authorized stale observations, opaque cursor history, active request,
  captured failed `initial`/`continuation`/`restart` read, synchronous admission
  and scope fencing. Retry repeats the captured read. Restart starts a new
  cursor-free chain, including when that restart later fails. Invalid or
  repeated cursors retain a typed unusable-continuation failure. Current
  authority loss conceals candidates through the workbook authority owner;
  obsolete failures cannot affect a replacement scope.
- `TimelineMentionActionControls.tsx` keeps the initiating read control mounted
  and focusable with a stable accessible name and busy/unavailable state.
  Recovery feedback sits beside the target picker. Focus moves to the picker
  when a control must disappear only if no newer keyboard, pointer, focus or
  scroll interaction took ownership. The filter and loaded eligible selection
  remain usable during continuation reads. Selected stale targets retain their
  label with an awaiting-revalidation cue and disabled option; resolution
  admission still requires a current eligible candidate and authority.
- The reader preserves typed server rejections and distinguishes invalid
  accepted pages from transport failure. The picker supports disabled stale
  options. `useTimelineMentionActions.ts` supplies selected identity to the
  discovery owner. Its existing explicit action admission remains unchanged.
  The old `nextCursor`-derived Retry and unused `reload` path were retired
  together.
- Unit and production browser cases live in
  `timelineMentionCandidates.test.tsx` and
  `mentions.discovery-recovery.spec.ts`, routed by authored test-family rows.
  The Timeline actions source guide now records this ownership.

No endpoint, schema, search, dependency, persistence, mention write, or data
migration is introduced. Server ordering, Host/Identity eligibility, raw mention
identity and independently discovered targets remain intact. Create/link and
uncertain-write recovery retain their separate owner, attempts and receipts.
Rollback is the authored source, test and family-row diff; regenerate derived
topology from the reverted family inputs with `make generate`.

## Verification evidence

| Check | Result and retained artifact |
| --- | --- |
| Baseline candidate slice | PASS, `.cartulary/test-results/20260926T191620Z-p70285` |
| Red unit reproduction | Expected failure, `.cartulary/test-results/20260926T192738Z-p75524` |
| Red production browser reproduction | Expected failure, `.cartulary/test-results/20260926T193140Z-p15510` |
| `make generate` after final test-family input | PASS, `.cartulary/test-results/20260926T195640Z-p58443` |
| `make agent-finalize` with `RESULTS_DIR` unset | PASS, `.cartulary/test-results/20260926T195705Z-p61998` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_mention_candidate_creation_review,web.workbook.regression.timeline_mention_candidate_recovery VITEST_MAX_WORKERS=1` | PASS 3/3 harness units after final source edit, `.cartulary/test-results/20260926T195929Z-p38021` |
| `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.timeline_mention_candidate_recovery,module.entities.browser.timeline_mention_creation_recovery,module.entities.browser.the_browser_inspector_resolves_existing_entities_325548131d` | PASS 13/13 harness units, `.cartulary/test-results/20260926T194957Z-p71698` |
| Final `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.timeline_mention_candidate_recovery` | PASS 11/11 harness units after final source edit, `.cartulary/test-results/20260926T195956Z-p41916` |
| `make lint-biome` | PASS 2/2 after final source edit, `.cartulary/test-results/20260926T195929Z-p38187` |
| `make frontend-typecheck` | PASS 2/2 after final source edit, `.cartulary/test-results/20260926T195929Z-p38137` |
| `make frontend-import-boundary-check` | PASS 2/2, `.cartulary/test-results/20260926T195730Z-p66474` |
| `make generate-drift` and `make generated-artifact-policy-check` | PASS 4/4 and 3/3, `.cartulary/test-results/20260926T195729Z-p66064` and `.cartulary/test-results/20260926T195729Z-p66170` |
| `make json-shape-check` and `make test-catalog-check` | PASS 3/3 and exit 0, `.cartulary/test-results/20260926T195729Z-p66264` and public catalog target |
| `make lint-markdown` after final handoff content | PASS, `.cartulary/test-results/20260926T200112Z-p76974` |

The browser row gates production Inspector continuation responses for Host and
Identity, checks keyboard Load more and Retry stay connected, focused, busy and
unavailable to duplicate activation, and confirms focus returns to the picker
at terminal paging. It checks request cursors and zero
mention mutations during discovery, then one successful explicit resolve.
The invalid-cursor case uses the server's rejected continuation, verifies the
disabled stale selected target during a held and then failed restart, and
verifies cursor-free Retry and revalidation. At 768×640 with 200% CSS zoom and increased text
spacing, the recovery control remains within the viewport and trial hit
testing and picker visibility pass. This is scoped interaction and reflow
evidence, not a whole-page accessibility claim.

`make agent-finalize` was attempted before generating the new browser route;
JSON shape validation correctly rejected stale derived topology at
`.cartulary/test-results/20260926T192927Z-p77781`. `make generate` repaired
that order. A later generation attempt failed because newly authored scenario
titles were not ASCII sorted at `.cartulary/test-results/20260926T194045Z-p85811`;
the input order was fixed. `make lint-biome` reported only new-file formatting
and import order at `.cartulary/test-results/20260926T195019Z-p1834`;
`make format-frontend` applied the repository formatter. `RESULTS_DIR` remains unset because no qualifying full
warm check run is used; retained-run maintenance is skipped.

## Digest acceptance assessment

| Rows | Status and evidence or scope rationale |
| --- | --- |
| A001–A003 | PASS: clause/source/routing map above, clean baseline and current source and authored input review. |
| A004–A005 | PASS: no token, theme or design-system literal change in the source diff. |
| A006–A009 | N/A: density, shell breakpoint and grid overflow owners are not changed; narrow Inspector and reflow are checked under A020. |
| A010–A011 | PASS: selected mention and raw identity, filter, focus, scope replacement and stale response checks. |
| A012–A015 | N/A: transaction, queue, editing and conflict owners are unchanged; existing creation/resolve browser rows protect the adjacent action boundary. |
| A016–A017 | PASS for this read seam: local loading/failure/stale eligibility and current versus obsolete authority loss are checked in units and browser. Full producer matrix and account lifecycle are outside this slice. |
| A018 | N/A: Evidence lifecycle is untouched. |
| A019–A020 | PASS for applicable keyboard focus, busy/name, local feedback and 768×640 zoom/text spacing scenarios. No broader conformance claim. |
| A021–A022 | N/A: virtualized grid identity and visual fixtures are unchanged; no golden changed. |
| A023–A025 | PASS: semantic selectors, no Markdown dependency, authored family rows and generated-output/drift verification. |
| A026 | PASS: compatibility and rollback above; no schema/route/data migration. |
| A027 | PASS after final verification and handoff completion below. |

## Final checks and remaining limits

The final focused browser rerun and TypeScript check are recorded above. The source and generated diff passes
`git diff --check`; final uncommitted state is retained for review. No visual
golden changed, so visual-golden maintenance is not applicable. Full-suite,
release, density and broad accessibility checks were skipped because the
focused source, browser, import and generated boundaries passed. There is no
whole-page accessibility conformance claim. Retained-run maintenance was
skipped because `RESULTS_DIR` was unset.
