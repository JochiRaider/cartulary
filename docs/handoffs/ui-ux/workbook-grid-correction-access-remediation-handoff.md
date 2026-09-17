# Workbook grid correction-access remediation

## Contract and ownership

Design §§7.4, 12.5, and 14.1 and D-AC-056 now require full containment of
each fitting focused editor or correction action, including below-minimum
widths. The full desktop support minimum remains 768 effective CSS pixels.
Oversized editors preserve maximum useful exposure, native text/caret scrolling,
individual correction access, and cancellation. CSS-zoom fixtures do not qualify
native browser zoom or a complete mobile workbook experience.

Core 03 REQ-03-281/300 retain ownership of exact local drafts, rejection,
explicit clear, and cancellation. Domain vocabulary and Testing Harness
mechanics are unchanged. Accessibility rows remain informative; they do not
establish product conformance or Core 05 publication evidence.

## Measured cause and correction

The instrumented baseline at source commit
`3fb4f4d800eda0fb0a04010ecd58abdec5b11c26` is retained under
`.cartulary/test-results/20260917T231621Z-p2140708/`.
At 390×720, CSS zoom 1:

| Stage | Target horizontal bounds | Target width | Grid bounds | Grid scrollLeft |
| --- | --- | --- | --- | --- |
| Display text after setup reveal | 183–389 | 206 | 1–389 | 323 |
| Mounted editor | 176–396 | 220 | 1–389 | 323 |
| Rejected editor | 176–396 | 220 | 1–389 | 323 |

The display text excludes the cell's seven-pixel inline padding. Revealing it
leaves the correctly sized editor seven pixels beyond the grid scrollport.
The clipping ancestor is the grid; rejection preserves the existing geometry.
At CSS zoom 2, the same seven local pixels become fourteen viewport pixels.
The initial diagnostic run also exposed an incorrect new cancellation assertion
against the display span; that assertion now addresses the owning gridcell.

Grid Adapter now reveals the actual mounted control after layout, focus recovery,
feedback changes, and relevant resizing. Shared viewport geometry accounts for
CSS zoom, client bounds, clipping ancestors, borders, scrollbars, and browser
viewport bounds. Editor reveal additionally excludes sticky/frozen grid chrome.
It scrolls only the owned grid and never changes field text, caret, column widths,
or row height. Coalesced work verifies the current focused editor and is cancelled
on disposal. Correction controls wrap and reposition inside the grid work area,
with local overflow for unusually tall content. The sticky creation row is also
excluded from that work area, and clipping-parent resize observations cover panes
whose clipping changes without resizing the grid itself.

The corrected 390px capture in `20260917T233228Z-p2341270` measures the editor at
165–385 after mounting and rejection, with scrollLeft 334. The additional four
pixels beyond the seven-pixel correction accommodate the computed focus outline
and offset. The column remains 220px. Human inspection of the active capture
confirmed visible input and correction controls.

No change to GenericMutationControl sizing was needed: measurements showed its
220px border box already matched the column. No pixel compensation, intersection
threshold reduction, backend change, dependency change, or persistence migration
was introduced.

## Changed implementation and coverage

- Grid Adapter: shared viewport geometry, private editor reveal lifecycle,
  correction-toolbar positioning, primary-control contract comment, and focused
  geometry/lifecycle tests. Column sizing and pointer scrolling share the geometry
  implementation.
- Test utilities: read-only geometry capture includes the target and focused
  control ancestor chains, with no values, text, headers, or credentials. Cell
  setup explicitly does not guarantee containment of a later editor.
- Workbook browser coverage: six independently registered viewport cases,
  pre-assertion geometry and failure-safe screenshots, all correction actions,
  exact draft retention, cancellation anchor, and zero invalid-edit PATCHes.
  Additional cases cover column/viewport resizing, oversized editors, a
  virtualized bottom edge, multiline/select controls, and keyboard reentry.
  Timestamp mutation coverage proves explicit Clear remains local until Commit
  and that successful correction and JSON null clear each submit once.
- Authored owner manifests enumerate exact tests and scenarios; the accessibility
  row names Grid Adapter as a collaborator. Generated routing is refreshed through
  Make. No new public harness schema or evidence family was added.

## Verification evidence

Focused successful runs:

| Command/selection | Result | Retained run directory under `.cartulary/test-results/` |
| --- | --- | --- |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.grid_autosave` (six initial viewport cases) | Pass | `20260917T232219Z-p2178364` |
| Workbook correction accessibility, timestamp retention, editor families, closure, availability, and three column-sizing rows | 17 browser tests passed | `20260917T232731Z-p2260927` |
| Timeline spreadsheet keyboard, pointer, refresh, rejection, and virtualization rows | Pass | `20260917T232904Z-p2296980` |
| `make test-slice OWNER=package.test_utils` | Pass | `20260917T232508Z-p2214048` |
| Grid Adapter mounted-editor reveal row | Pass | `20260917T232833Z-p2295635` |
| Workbook semantic focus and two viewport-continuity rows | Pass | `20260917T232902Z-p2296581` |

The full Grid Adapter owner slice passed its other rows in
`20260917T232508Z-p2214038`; the new reveal row initially failed because its
jsdom fixture lacked scrollTo. The subsequent narrow failure in
`20260917T232731Z-p2260914` was a missing overflow-axis fixture style. Both fixture
defects were corrected before the successful reveal-row run above.

Generation initially rejected unsorted titles and dynamic test-title registration
in `20260917T231433Z-p2131210` and `20260917T231513Z-p2134443`. Literal titles,
distinct scenario IDs, and sorted selectors resolved those routing failures.
Historical failure artifacts remain investigation evidence.

Final maintenance and completed integration checks:

| Target/selection | Result | Retained run directory under `.cartulary/test-results/` |
| --- | --- | --- |
| `make agent-finalize` | Pass; retained-run maintenance skipped because RESULTS_DIR was unset | `20260917T233127Z-p2336626` |
| `make frontend-typecheck` | Pass | `20260917T233228Z-p2340862` |
| `make frontend-import-boundary-check` | Pass | `20260917T233228Z-p2340884` |
| `make lint-biome` | Pass | `20260917T233228Z-p2340913` |
| `make lint-markdown` | Pass | `20260917T233228Z-p2340930` |
| `make generate-drift` | Pass | `20260917T233228Z-p2340580` |
| `make generated-artifact-policy-check` | Pass | `20260917T233228Z-p2340591` |
| `make json-shape-check` | Pass | `20260917T233228Z-p2340601` |
| Grid Adapter support-specimen visual row | Pass | `20260917T233228Z-p2340768` |
| Timeline default, grouped, and active-edit/save-state visual rows | Pass | `20260917T233228Z-p2340796` |
| Workbook default shell and inline-edit/recovery visual rows | Pass | `20260917T233248Z-p2430852` |
| `make browser-e2e-a11y` before final sticky-row refinement | 62 browser tests passed; 20 work units passed | `20260917T233228Z-p2341270` |
| Final reveal unit row, including sticky creation-row exclusion | Pass | `20260917T233618Z-p2482426` |
| Final correction accessibility, role loss/revocation, and session suspension/account replacement rows | 12 browser tests passed | `20260917T233618Z-p2482445` |
| Final `make agent-finalize` | Pass; RESULTS_DIR unset | `20260917T233659Z-p2513993` |
| Final `make frontend-typecheck` | Pass | `20260917T233742Z-p2520828` |
| Final `make lint-biome` | Pass | `20260917T233742Z-p2520848` |
| Final `make browser-e2e-a11y` after all implementation refinements | 62 browser tests passed; 20 work units passed | `20260917T233742Z-p2520977` |
| Final handoff `make lint-markdown` | Pass | `20260917T233858Z-p2557243` |

No golden changed, so no golden refresh or refresh migration was needed.
The final sticky-creation-row and clipping-parent refinements received focused
revalidation separately from the earlier visual runs above. Exact target and row
selections are retained in each run's `run-manifest.json`. The primary reproduction
remains:

```sh
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.grid_autosave
```

Visual inspection also covered the final 260px oversized-editor and virtualized
bottom-edge correction captures from `20260917T233618Z-p2482445`: every correction
action was visible, the bottom-edge toolbar was positioned above the editor, and
the sticky creation row remained outside its bounds. All remediation phases are
complete; no unresolved clipping, draft-retention, focus, or mutation failure was
observed in the selected verification.

Backend, database migration, dependency-security, release, and Core 05 publication
gates were not run: this change introduces no corresponding backend, persistence,
dependency, release, or publication change. Full mobile support and native browser
zoom qualification remain outside this remediation.

Rollback requires a coherent source, specification, test, and generated-routing
revert; no data migration is required.
