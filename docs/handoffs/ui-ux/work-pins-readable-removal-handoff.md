# Readable Work pins and continuous removal focus

## Boundary and authority

Baseline: `main`, `d4e7e67638756858d0a46aaff566665334576dfa`, initially clean,
one commit ahead of `origin/main`. Source owner: `web.workbook`. Verification
owners: `web.workbook` and `module.workbook`, confirmed against the current
catalogs and their module-author task guides.

Core 03 §2.5.1/REQ-03-311 owns auxiliary attachment; §2.5.3/REQ-03-313 owns
session pins; §2.5.4/REQ-03-314 and §2.5.5/REQ-03-315 own navigation and Return.
REQ-03-299/100 retain authority and authoring lifetimes. Core 04
§2.2/REQ-04-169/170 governs navigation security. Design §§7, 14, 15.2A guides
presentation; Domain supplies vocabulary. Research and earlier handoffs were
evidence, not additional authorization. No adopted specification changed.

The user chose complete wrapping, context plus an Identity disclosure, and
next/previous/Pin view removal focus. That focus sequence is task acceptance,
not an assertion that Core mandates this exact sequence.

## Implementation and retirement

- `WorkbookWorkPanel.tsx` replaces the old inline pin row with one private
  `WorkPinItem`. The existing Open button wraps prose and unbroken tokens locally.
  Static surface titles identify record/base-surface context; saved views and
  extension roots show their kind. Native Identity disclosures show the complete
  descriptor identifiers. Remove and Identity wrap beneath the label.
- `WorkbookNavigationIntent.ts` and `WorkbookSessionNavigation.ts` add and set a
  session-only `unavailable` marker when concealing a pin. Generic presentation
  suppresses labels, context, identifiers, and accessible descriptions. An
  authorized record literally named “Unavailable item” still has its identity.
- `useWorkPinRemovalFocus.ts` owns only a short-lived Work focus intention.
  Semantic identity, expected survivor order, session, navigation attempt,
  attachment, authority, and intervening interaction fence recovery. It uses
  `preventScroll` and adjusts only the existing dock scrollport, including at
  supported zoom. A token-based indicator remains visible after pointer removal.
- `WorkbookWorkPanel.test.tsx` tests the production renderer, cancellation,
  concealment, descriptor kinds, exact removal, and keyboard close behavior.
  `workbook-destination-handoff.spec.ts` adds production browser scenarios for
  the three requested viewports and a viewer.
- `tools/frontend_source_ownership.json` registers the two new source files.
  `tools/test_families/{web.workbook,module.workbook}.json` owns new routing;
  `make generate` produced the batch manifest and topology render index.

The common decision is how an authorized descriptor is read, identified, opened,
and removed. The old pin rendering is retired in place. The new hook is local to
Work; the reference-specific removal helper remains with its existing callers.
There is no shared collection framework, second pin store, metadata cache, or
parallel renderer. Future descriptor kinds can extend this private presentation
without taking navigation or record ownership.

Global quiet-command styling, auxiliary-dock ownership, navigation presentation
and host, query/saved-view owners, drafts, operations, receipts, and Recovery
counts remain unchanged. Pins retain insertion order, twenty-pin capacity,
duplicate-before-capacity semantics, and semantic identity independent of label
and field hint. Display/disclosure/removal introduce no record reads or writes.

## Verification commands and evidence

Commands run from the repository root. Discovery included `make help`,
`make help-all`, and `make task-guide ROLE=module-author OWNER=web.workbook`
and `OWNER=module.workbook`. New IDs came from `make author-test-row-id`.
Each run below used the explicit `CARTULARY_TEST_RUN_ID` shown; its root is
`.cartulary/test-results/<run-id>`. Product runs retain `run-summary.json`,
`target-summaries/<target>.json`, and exact row results under that root.
Observation used the exact run locator through `make test-run-status` and
`make explain-run`; outer process exits were collected separately.

The initial unit selection was:

```text
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.work_pin_presentation_removal_27944912ba,web.workbook.regression.session_navigation,web.workbook.regression.navigation_feedback,web.workbook.regression.workbench_navigation_admission,web.workbook.regression.navigation_presentation
```

The initial browser selection was:

```text
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.work_pin_readability_removal_1440_9594ec3857,module.workbook.browser.work_pin_readability_removal_1024_b97af532c7,module.workbook.browser.work_pin_readability_removal_768_741d03e297,module.workbook.browser.work_pin_viewer_removal_9f4c8a0b9c,module.workbook.browser.destination_handoff_1024_963c4de116,module.workbook.browser.destination_handoff_1440_920aaf64d5,module.workbook.browser.destination_handoff_768_8f85fa718d,module.workbook.browser.destination_handoff_viewer_64fc8d19e2,module.workbook.browser.destination_handoff_hidden_a9564ecda1,module.workbook.browser.destination_handoff_supersession_62c776e360
```

| Run ID | Command/selection | Result; outer exit |
| --- | --- | --- |
| `work-pins-generate-20261009-01` | `make generate` | PASS; 0 |
| `work-pins-format-20261009-03` | `make format` | PASS; 0; authored diff inspected |
| `work-pins-finalize-20261009-02` | `make agent-finalize` | PASS; 0 |
| `work-pins-unit-20261009-01` | Five unit rows above | PASS; 0 |
| `work-pins-unit-20261009-02` | New renderer row only, after final focus/selector edits | PASS; 0 |
| `work-pins-browser-20261009-01` | Ten browser rows above | Six existing rows PASS; four new rows FAIL; 2 |
| `work-pins-browser-20261009-02` | Four new browser rows only | PASS; 0; four scenarios, eleven harness units |
| `work-pins-typecheck-20261009-02` | `make frontend-typecheck` | PASS; 0 |
| `work-pins-boundaries-20261009-01` | `make frontend-import-boundary-check` | PASS; 0 |
| `work-pins-biome-20261009-02` | `make lint-biome` | PASS; 0 |
| `work-pins-policy-20261009-01` | `make generated-artifact-policy-check` | PASS; 0 |
| `work-pins-markdown-20261009-01` | `make lint-markdown` | PASS; 0; `adhoc/lint-markdown/tool-run-summary.json` |

The second browser run used exactly the first four row IDs in the browser
command above, with the same target and owner. The second unit run used exactly
`web.workbook.regression.work_pin_presentation_removal_27944912ba` with the same
target and owner. Commands from finalize onward set
`CARTULARY_OUTPUT_MODE=machine`. Finalize's
`unit-artifacts/finalize-summary.json` records successful JSON shape, catalog/tier
coverage, and generated-structure drift checks with no further generated changes.

The new browser scenarios exercise twenty pins, duplicate admission at capacity,
overflow refusal, equal prose labels, a 199-character token, native disclosures,
first/middle/last/sole removal, pointer and keyboard activation, recovered ring
geometry, immediate Escape, empty Work, pin Open/Return, and a retained Notes
draft. They assert no descriptor reads during display/disclosure/removal, no
source writes, unchanged grid selection/scroll and page/footer geometry, and no
reopened editor. The viewer scenario exercises applicable pin navigation and
removal without edit capability. Unit evidence adds disabled-target skipping,
intervening pointer/keyboard/focus cancellation, closure, detachment, unmount,
authority loss/recovery, attempt changes, unexpected pin changes, and concealed
identity/ARIA. Geometry claims come from browsers, not jsdom.

Earlier failed runs were resolved, not discarded: format `-01` and
`work-pins-biome-diagnose-20261009-01` found the intentional session cleanup
dependency; the narrowly explained lint suppression retains that cleanup.
Finalize `-01` and `work-pins-shape-20261009-01` found stale authored-catalog
projections, repaired by `make generate`. Typecheck `-01` found five unsupported
Testing Library role-selector options, removed. Browser `-01` found a fixture
password below the supported minimum and a duplicate-feedback assertion after
inspector closure clears feedback. The corrected browser test checks duplicate
position/count; the session unit row checks its immediate hint and capacity
ordering. Failed commands exited 2. Exact diagnostics remain in their run roots.

`RESULTS_DIR` was unset for finalize: retained-run maintenance was skipped because
there was no eligible successful full warm check. Finalize ran before broader
end-of-run checks. No unrelated product suites or golden refreshes were run.

## Rendered review

Before source edits, seeded editor review session `20261009T120548Z-p72754`
completed 66 serialized operations. It reproduced clipped prose at 1440×900 and
1024×768, equal-label ambiguity, body focus after first/middle/last/sole removal,
failed immediate Escape, and a clipped 199-character token at 768×640 and 125%.
Its safe receipt is
`.cartulary/test-results/20261009T120548Z-p72754/ui-review/terminal.json`:
cleanup complete, stop and foreground exits 0.

Post-change seeded review used the exact locator
`/home/jochi/code/cartulary/.cartulary/test-results/20261009T124708Z-p15146/ui-review/session.json`.
Its sealed source digest matches the successful second browser run:
`3bbda206c2b9f4ed174dc635e6f91d1ba6fa570135278aafe4c6988525042542`.
Editor and viewer inspection covered all three requested viewports, equal-label
disclosures, long prose/tokens, 125% zoom, first/middle/last/sole removal, and
empty Work. Complete prose occupied three lines at 1440 and two at 1024;
the token wrapped without horizontal overflow at 768 and 125%. Full distinct
record identifiers were visible before activation. Removal visibly revealed the
next/previous/fallback control within Work. Immediate Escape restored Work trigger
focus, with the footer remaining at the viewport bottom.

Ten original captures were inspected after verifying manifest bytes and SHA-256.
Advisory axe reported zero violations and five incomplete color-contrast results
across captures; those incomplete results concerned existing surrounding chrome,
not a claim of full accessibility conformance. Changing zoom while the same
control already owns focus does not itself refocus/scroll it; ordinary Tab reveals
it, and removal recovery reveals its own target at the new zoom.

One viewer action used the editor's gridcell name and returned
`target_unavailable`; subsequent requests with the stale epoch were rejected
without mutation. Status and a fresh snapshot recovered the exact viewer
`Read-only Activity Date (UTC)` name, and inspection continued successfully.
This review-command error is separate from the passing viewer product test.

`make ui-review-stop UI_SESSION=<exact locator above>` completed successfully;
its safe `ui-review/terminal.json` reports `status: ok`, `cleanup: complete`, and
no failed requests or console errors. Stop and foreground process exits were 0.
Only caller-owned request scratch files were removed afterward. Browser-review
observations remain separate from product-test evidence. No Work capture exists
in the current visual golden spec; no golden was refreshed. Private review images
expired on cleanup and are not retained handoff links or conformance claims.

## Acceptance and compatibility

Assessment against the maintained digest is bounded to this seam:

| Rows | Assessment | Evidence or scope rationale |
| --- | --- | --- |
| A001–A003 | PASS | Governing clauses and source/routing owners above; baseline, manifests, guides, generated boundaries, and final diff inspected. One descriptor-presentation decision, explicit retirement, no owner relocation. |
| A004 | PASS | Typography, spacing, and focus use authored tokens. Local wrapping and dock reveal add no design registry. |
| A008–A009 | PASS, scoped | Three viewport scenarios and 125% zoom; visible/reachable actions, dock-owned scrolling, stable page/grid/footer. No responsive-threshold or inspector-clamp change. |
| A011, A014 | PASS, scoped | Draft retention, pin-open/Return, selection/scroll continuity, removal focus and Escape; existing destination/hidden-field/supersession rows pass. |
| A016–A017 | PASS, scoped | Viewer/readability admission remains owned by navigation; concealment and authority/attempt/detachment cancellation tests pass. No query-state or account-lifetime redesign. |
| A019–A020 | PASS, scoped | Real keyboard/pointer actions, native disclosures, visible recovered rings, complete prose/token rendering, empty/twenty-pin and zoom coverage. Advisory scan limitations above. |
| A022 | PASS, scoped | New production-renderer browser scenarios and ten inspected seeded originals; no affected canonical Work golden or golden refresh. |
| A023–A025 | PASS | Semantic descriptor selectors, no Markdown executable dependency, authored row routing with generated projections; import, catalog, drift, and generated-policy checks pass. |
| A026 | PASS | No navigation/authorization/write/persistence contract changes; retirement, helper retention and rollback stated. |
| A027 | PASS | Evidence, failures, cleanup, compatibility and rollback recorded here; Markdown lint and final diff checks pass. |
| A005–A007, A010, A012–A013, A015, A018, A021 | N/A | No theme/density/creation/dispatcher/transaction/recovery/conflict/evidence/virtualization change. Existing owners stay in place; draft and navigation effects relevant to this seam are covered above. |

Broader suites were not needed: no query/coordinator, writer, density, theme,
virtualization or persistence owner changed. The twenty-pin case is automated
product evidence; the interactive seeded review used a smaller representative
list. No new unresolved product risk was found.

No API, wire format, database, persistence, or migration change. The concealment
marker records already-known session state and grants no authority. Read
permission remains independent of edit permission. Rollback reverts this cohesive
frontend/test/authored-routing/generated-projection slice together; no persisted
data rollback is needed. No commit, publication, or deployment was requested.

Final repository review: `git diff --check` passed, HEAD remains the baseline,
and status contains only this frontend/test/routing/generated slice and this
handoff. No unrelated file was reformatted or removed.
