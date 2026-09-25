# Timeline collection native revision fix

## Boundary and reproduction

Baseline: clean `main` at `f67663d7a486a9d093f6399b9fc53d0055c48ca6`.
This bounded correction follows the completed collection-input cleanup handoff;
that record remains historical evidence. Core 03 REQ-03-111, REQ-03-219 and
REQ-03-298 own authoring and settlement. Design §9.3 owns the collection
presentation. Timeline source sits under `web.workbook`; active focused
verification routes through `module.timeline`, with Workbook regressions under
`web.workbook`.

The new production browser case first failed on the old implementation: after a
submitted Host token was held at the response boundary, a native select-and-replace
gesture emitted one `input` event with identical text. Releasing the older
acknowledgement cleared the field to `""`. The red browser run is
`.cartulary/test-results/20260925T212001Z-p18020`; the red component run is
`.cartulary/test-results/20260925T211940Z-p17170` (revision remained 1 rather
than advancing to 2). Earlier attempts without the repo's pinned Node directory
on `PATH` failed in harness child startup before executing tests; they did not
establish product behavior.

## Correction and compatibility

- `TimelineCollectionCell` now publishes exactly one fresh revision from each
  native `input` event, even when text is unchanged. Its redundant `onChange`
  write is retired. Composition start still protects the pre-input interval;
  composition end, blur, materialization and rerendering do not invent edits.
- The retained store, capture promotion, mutation queue, revision settlement and
  Enter/Tab plus blur deduplication retain their existing ownership. The edit
  event itself does not dispatch a write. No public API, wire, persistence,
  dependency or migration change is needed.
- Component/registry tests and the routed production browser cases cover equal
  native replacement, all three collection fields, independent Inspector text,
  older success and Tag rejection, one logical departure across key and blur,
  native paste/undo/redo, and recordless revision promotion with one create.
  Existing Add, Escape, IME, mention and tag behavior remains covered by the
  selected owner rows.

Authored test routing changed in `tools/test_families/module.timeline.json`;
`make generate` produced its browser batch and topology index projections.
Executable verification has no dependency on this or other Markdown.

## Verification and remaining limits

All commands ran from the repository root. Focused row runs used
`PATH="$PWD/tmp/node-runtime/bin:$PATH"` so harness child processes could find
the repo's pinned Node runtime.

| Check | Result and run root |
| --- | --- |
| `make generate` | PASS, `20260925T213805Z-p47513` |
| `make test-slice OWNER=module.timeline` (collection component and registry rows) | PASS, `20260925T212902Z-p29266` |
| `make service-backed-test-slice OWNER=module.timeline` (new revision, authoring and promotion rows) | PASS, `20260925T213216Z-p95916` |
| `make service-backed-test-slice OWNER=module.timeline` (collection input characterization row) | PASS, `20260925T213954Z-p61120` |
| `make test-slice OWNER=web.workbook` (collection overflow and grid draft lifetime rows) | PASS, `20260925T213336Z-p28876` |
| `make frontend-typecheck` | PASS, `20260925T213349Z-p29822` |
| `make format` | PASS, `20260925T213422Z-p31641` |
| `make generate-drift`; `make lint-biome` | PASS, `20260925T213818Z-p50565`; PASS, `20260925T213917Z-p58965` |
| `make generated-artifact-policy-check`; `make json-shape-check` | PASS, `20260925T214151Z-p95999`; PASS, `20260925T214151Z-p96018` |
| `make agent-finalize` | PASS, `20260925T213848Z-p54760` |
| `make lint-markdown` | PASS, `20260925T213917Z-p58961` |

The first `make generate-drift` attempt (`20260925T213349Z-p29651`) failed in
harness startup without Node on `PATH`; rerun passed. The first `make lint-biome`
attempt (`20260925T213349Z-p29867`) identified formatting in the two touched
test files; `make format` repaired it, and rerun passed. Retained-run maintenance
was skipped because `RESULTS_DIR` was unset.

Intermediate `make generate` failures at `20260925T211221Z-p7089`,
`20260925T212804Z-p19987` and `20260925T212825Z-p23179` were authored routing
order/count errors introduced in this slice; correcting the row and rerunning
generation resolved them. The initial `make service-backed-test-slice` runs at
`20260925T211316Z-p13576` and `20260925T211505Z-p46929`, an existing browser
row at `20260925T211704Z-p79884`, and `make test-slice` at
`20260925T211848Z-p16132` failed before test execution because harness child
processes could not find Node on `PATH`. The expanded browser run at
`20260925T212912Z-p29938` passed success/rejection but exposed an over-strong
recordless test assertion about backward selection direction; the final test
checks the owned selection endpoints and passed at `20260925T213055Z-p63111`.

The recordless promotion check preserves selection endpoints and focus. Its
existing handoff normalizes backward selection direction to forward; this slice
does not change that separate selection transfer. No visual golden, layout or
timed-measurement behavior changed. Rollback is to revert this slice's authored
cell, tests and routing, then rerun `make generate`; no data action is needed.

Acceptance assessment against the advisory digest: A001, A003, A007,
A010–A011, A013–A014, A019, A023–A027 are PASS on the owner map, focused browser/unit
evidence and generated/drift checks above. A002 is N/A because this is an
authorized behavior correction without a new structural abstraction. A004–A006,
A008–A009, A015–A018 and A020–A022 are N/A because their token, layout,
conflict, authority, evidence, virtualization and visual boundaries were not
changed. A012 is N/A because captured request identity and uncertain replay
were unchanged. No applicable acceptance row is blocked.
