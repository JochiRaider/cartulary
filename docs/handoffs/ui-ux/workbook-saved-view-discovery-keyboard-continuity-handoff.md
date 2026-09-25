# Saved-view discovery keyboard continuity handoff

## Boundary and reproduction

Entry was clean `main` at `9b04ef324733492abe97d49b202022c56c1650d6`.
Core 03 REQ-03-022A owns bounded discovery, addressed resource observation,
activation and cancellation. Design §§8.3, 12.2 and 12.4 own picker presentation
and loading versus disabled button states. Workbook source remains under
`web.workbook`; `module.savedviews` independently routes the public browser and
accessibility evidence. The research NLSpec essay, prior discovery handoff and
localized UI/UX digest were advisory navigation, not product authority.

Failure-first regression: the expanded selector test failed with missing
`aria-busy` on held Next at
`.cartulary/test-results/20260925T184428Z-p5852`; the expanded service-backed
browser scenario found Next disabled and unfocused during its held read at
`.cartulary/test-results/20260925T184533Z-p7134`. Initial baseline row runs
stopped at harness `service_start_error` because the pinned Node binary was not
on this shell's `PATH`. Supplying `tmp/node-runtime/bin` to the public Make
invocations let both regressions execute. `make doctor` passed independently.

## Change and retirement

- `SavedViewDiscovery` now publishes the initiating paging or retry action
  separately from the requested cursor, distinguishes First from Refresh,
  guards duplicate reads, and clears action state on settlement or cancellation.
  Exact retry cursors, one accepted page, ten checkpoints, page-local counts and
  existing read generations remain with that owner.
- `SavedViewBrowser` keeps the initiating action mounted, tabbable, named and
  visibly focused with a token-backed pending cue and `aria-busy`. Other blocked
  actions use the unavailable state. Retry remains rendered during its read;
  success moves focus to the first candidate, or Unsaved view when empty, if
  its action disappears or becomes unavailable. Candidate focus follows
  `saved_view_id` through reordering and uses a local positional fallback only
  while the picker owns focus. External focus is never reclaimed.
- Retired native pending-disable on the initiating action, `problem`-only Retry
  rendering, and page-index focus ownership. Query-footer and Columns focus
  patterns informed ownership checks; their domain state machines remain
  separate. Controller selection, addressed-resource admission, selected label,
  Modified state, Unsaved view, Escape, outside dismissal, Close and generation
  fences remain in their existing owners.

No public API, persisted configuration, route, record write, generated artifact,
dependency or migration changed. The internal discovery snapshot gained a typed
`pendingAction`, and Refresh has a distinct internal method. The browser test
proves an uninterrupted keyboard path at 768 px through held Next, repeated
failure, held Retry, success, backward/forward browsing and explicit activation.
Before activation, working grouping/layout and the selected identity remained
unchanged, the inspector stayed closed, and no incident writes were sent. The
existing accessibility row also covers supported 200% zoom.

## Verification and limitations

All commands ran from the repository root through public Make targets with the
pinned Node directory on `PATH` where the harness needed it.

| Check | Result and artifact root |
| --- | --- |
| `make test-slice OWNER=web.workbook` (selector, independent reads, operation owner rows; final external-focus selector check) | PASS, `.cartulary/test-results/20260925T185653Z-p54738` and `...T190253Z-p1137` |
| `make test-slice OWNER=module.savedviews` (frontend pagination row) | PASS, `.cartulary/test-results/20260925T185653Z-p54761` |
| `make service-backed-test-slice OWNER=module.savedviews` (discovery browser and accessibility rows) | PASS, `.cartulary/test-results/20260925T185705Z-p56381` |
| `make agent-finalize` before broader checks and after the handoff | PASS, `.cartulary/test-results/20260925T185535Z-p44449` and `...T185902Z-p92016` |
| `make frontend-typecheck` and `make frontend-import-boundary-check` | PASS, `.cartulary/test-results/20260925T185601Z-p48570` and `...-p48596` |
| `make format` and `make lint-biome` | PASS, `.cartulary/test-results/20260925T190307Z-p1787` and `...T185653Z-p54927` |
| `make lint-markdown` and `git diff --check` | PASS, `.cartulary/test-results/20260925T185925Z-p95919` and the final worktree check |

The first Biome run reported formatting in three touched files at
`.cartulary/test-results/20260925T185601Z-p48646`; the public formatter resolved
it. Retained-run maintenance was skipped because `RESULTS_DIR` was unset. This
slice makes memory-local focus guarantees; it adds no reload persistence or new
screen-reader conformance claim. No visual golden was changed.

## Digest acceptance and rollback

| Digest rows | Disposition |
| --- | --- |
| A001–A004 | PASS: adopted owners, current source boundaries and token-backed states were reviewed; one discovery-action/focus decision was changed without a new shared machine. |
| A008–A009, A011, A019–A020 | PASS: 768 px and supported zoom, picker bounds, semantic focus, keyboard recovery, names, busy/unavailable states and local feedback have owner-routed evidence above. |
| A023–A027 | PASS: existing semantic selectors and routed rows, Markdown-independent executable checks, no generated edits or migration, and this evidence-backed handoff. |
| A005–A007, A010, A012–A018, A021–A022 | N/A: theme, density system, creation, inspector dispatch, transaction/edit/conflict/query-grid/security/virtualization and visual-golden behavior were not changed by this picker slice. |

Advisory R001/R002/R007 were adopted for keyboard focus and distinct loading
feedback. The upstream loading-button suggestion was adapted: duplicate
activation is blocked without native-disabling the focused loading control, as
Cartulary design §12.2 requires. R012 was adapted by keeping read identity in
the discovery owner and transient focus ownership in the component.

Rollback is a targeted revert of this slice's five authored TypeScript files
and this handoff; no data rollback is needed. That revert restores the confirmed
keyboard focus defect. Next action is review of this bounded change.
