# Network Analysis Interaction Chrome Refactor Handoff

## 1. Baseline and authority

| Item | Recorded value |
| --- | --- |
| Started | `2026-08-31T15:02:33-04:00` |
| Branch | `main` (`origin/main` ahead by two commits) |
| Commit | `a8019384c1c10073caea5606d1c34ef2cfe556b0` (`Visual Renderer and Workbook Density`) |
| Initial Git status | Clean |
| Tool versions | Git `2.53.0`; Node `v24.15.0`; pnpm `10.33.0`; Python `3.14.4`; Go `1.26.6`; jq `1.8.1`; GNU Make `4.4.1`; Docker `29.5.2` |
| Digest localization | `meta/localization.json` records repository commit `2356949f7ec3c8e27ff83ae695d60e06a387d0e5`; live owners and repository inputs were revalidated at the current commit. |
| Existing user changes | None at start. Any later unrelated changes remain user-owned and will be preserved. |

The adopted Network Flow NLSpec owns workspace regions, state, identity, query,
graph, lifecycle, authorization, and persistence behavior. `docs/design.md` owns
observable design behavior. The digest is advisory. No contradiction among the
adopted owners was found during planning or this execution baseline.

Authorized paths are limited to Network Analysis presentation and local shared
code under `apps/web/src/networkFlow/**`; relevant unit/browser evidence and
Network Analysis goldens under `apps/web/e2e/**`; additive selector and typed
token-reference facades under `packages/ui-contracts/src/**`; authored source
ownership and visual-fixture inputs under `tools/**`; Make-generated harness
outputs resulting from those inputs; and this handoff. Network Flow APIs,
queries, graph algorithms, imports, mappings, storage, authorization, lifecycle,
extension claims, and persisted data are not authorized for behavioral change.

Runtime code, tests, generators, conformance, and release evidence will not
read or depend on this handoff or another Markdown file. Generated roots will
not be hand edited.

## 2. Workstream ledger

| Workstream | Status | Exit condition |
| --- | --- | --- |
| NA-01 — Baseline and token/control characterization | DONE | Every claimed interaction, token, responsive, and visual gap has deterministic evidence. |
| NA-02 — Typed token and local control foundation | DONE | Typed token/control tests pass and no undefined Network Analysis token reference remains. |
| NA-03 — Workspace migration | DONE | All visible interaction chrome uses the cohesive seam without behavioral regression. |
| NA-04 — Responsive, accessibility, and visual reconciliation | DONE | Applicable browser evidence and reviewed generated golden manifest pass. |
| NA-05 — Final validation and handoff | DONE | Scope, compatibility, generated output, commands, and retained results are complete. |

Only one row may be `IN_PROGRESS`. Before starting the next workstream or
ordered NA-03 group, this ledger and checkpoint log are updated. If an adopted
owner contradiction emerges, the affected row becomes `BLOCKED` and records
exactly `BLOCKED: owner contradiction`.

## 3. Characterization baseline

Confirmed before implementation:

1. Network Analysis references six variables absent from the authored token
   registry: `--ct-colors-accent-strong`, `--ct-colors-danger-strong`,
   `--ct-colors-text-muted`, `--ct-colors-text-primary`,
   `--ct-font-family-mono`, and `--ct-shadow-popover`.
2. Visible inputs, selects, and many buttons use native light browser chrome in
   the dark graphite workspace.
3. Equivalent primary, secondary, destructive, icon, mode, selected, pending,
   and disabled controls use unrelated local styles or no owned styles.
4. Accepted, rejected, and graph query bands do not contain long values or
   action groups reliably at narrow and compact desktop bands.
5. Saved-graph navigation, object selection, pagination, popovers, and dialogs
   use native or hardcoded light presentation.
6. Mapping dialog styles include `#fff`, raw rgba shadows, and non-contract
   radii/spacing; the semantic-grid popover uses the invalid shadow variable.
7. Existing visual evidence covers representative 1440x900 states, but narrow,
   compact, zoom, text spacing, modal, and long-value evidence is incomplete.
8. Repeated local spacing, radius, input, button, and surface literals bypass
   existing component tokens.

Baseline evidence passed before product changes:

- `make test-slice OWNER=web.networkflow`: 37/37 at
  `.cartulary/test-results/20260831T185001Z-p93100`.
- `make test-slice OWNER=package.ui`: 10/10 at
  `.cartulary/test-results/20260831T185006Z-p93614`.
- `make browser-e2e-visual`: 12/12 at
  `.cartulary/test-results/20260831T185017Z-p94433`.

The committed accepted-inspector, rejected-diagnostics, graph-contributor, and
saved-graph-result captures were manually inspected as pre-change evidence.
They show the native-light controls and inconsistent hierarchy described above.

## 4. Advisory query dispositions

The two required offline searches ran with `PYTHONDONTWRITEBYTECODE=1` and
`python3 -B`. The digest was not mutated; `--design-system`, `--persist`, and
`--force` were not used.

| Recommendation | Disposition | Cartulary application |
| --- | --- | --- |
| Associated form labels | ADOPT | Preserve explicit label/control relationships. |
| Visible focus | ADOPT | Use the owned focus ring on every interactive control. |
| Modal containment and return | ADOPT | Preserve the existing focus hook and deterministic invoker return. |
| Semantic headings and native controls | ADOPT | Keep owner labels and native buttons, inputs, and selects. |
| Submit feedback | ADAPT | Express pending/success/error through existing owner states and status precedence. |
| Typography scale | ADAPT | Use existing dense workbook typography tokens only. |
| Local form state | ADAPT | Keep existing drafts; do not introduce a new state architecture. |
| Breadcrumbs and auto-rotation | REJECT | They do not belong to the Network Analysis workflow. |
| New context/custom-hook/generic architecture | REJECT | It is unrelated to this presentation refactor and would expand scope. |

## 5. Checkpoint log

| Time | Workstream | Paths and decisions | Commands and results | Compatibility, rollback, risks, next action |
| --- | --- | --- | --- | --- |
| `2026-08-31T15:02:33-04:00` | Baseline / NA-01 start | Created this handoff before product edits. Recorded clean scope, authority, six invalid tokens, interaction inventory, advisory dispositions, existing visual evidence, and current task routing for `web.networkflow`, `module.networkflow`, `web.design`, and `package.ui`. Existing authored tokens cover every required semantic role; no token addition or alias is planned. | Git/tool baseline passed. Current routing remains focused owner slices plus the requested frontend/browser targets. Existing focused and visual run roots are recorded above. | No product behavior changed. Rollback this handoff with the complete slice. Principal risk is accidental grid or workflow drift; bounded migration and focused checks gate every group. Next: add deterministic characterization and the typed presentation seam. |
| `2026-08-31T15:11:38-04:00` | NA-01 completion / NA-02 start | Added documentation-free boundary assertions for authored token membership and legacy light-theme fallbacks. Added the typed token-reference facade, additive semantic identities, and the initial presentation-only control seam plus component tests; no runtime workspace consumes it yet. | Existing routed boundary row passed at `.cartulary/test-results/20260831T191029Z-p44127`. The pre-generation frontend catalog passed 391/391 at `.cartulary/test-results/20260831T191040Z-p44699`; new test titles are intentionally not claimed until the Make generator runs in NA-02. | Characterization is deterministic without changing Network Flow behavior. The assertions will fail after catalog generation until invalid variables and light fallbacks are removed. Rollback tests only with the complete refactor. Next: generate routing, prove the expected failures, then complete token/control migration. |
| `2026-08-31T15:23:16-04:00` | NA-02 completion / NA-03 groups 1-2 | Added the typed `cartularyDesignTokenReference` facade, closed selectors for graph surfaces/advanced filters/saved graph identity, scoped native control components, dark autofill/focus/disabled/pending rules, and routed component evidence. Removed all six invalid variable names without aliases. Migrated the workspace header, table tabs, top-level modes, lifecycle triggers/dialogs, empty state, graph-surface modes, and accepted/rejected query bands. | `make generate` passed at `.cartulary/test-results/20260831T192254Z-p3474`; control seam row passed at `.cartulary/test-results/20260831T192308Z-p6362`; `package.ui` passed 10/10 at `.cartulary/test-results/20260831T192312Z-p6898`. Typecheck passed at `.cartulary/test-results/20260831T191953Z-p91779`; the full focused Network Flow slice passed 37/37 at `.cartulary/test-results/20260831T192002Z-p92202`. Earlier generation roots `.cartulary/test-results/20260831T192109Z-p96720` and `.cartulary/test-results/20260831T192230Z-p312` identified ASCII ordering requirements in the authored row/collaborator lists; both were corrected. | No owner token or workflow changed. Tabs/modes retain native semantics and keyboard handlers; query draft/apply/clear logic is unchanged. Rollback helper, seam, selectors, routing, workspace/query migration, and generated topology index together. Next: migrate graph and contributor controls. |
| `2026-08-31T15:27:10-04:00` | NA-03 group 3 completion | Migrated graph scope/aggregation choices, bucket select/navigation, retry, bounded paging, vertex/edge selection, row linking, contributor drawer close/link actions, and shared query pagination. Native radio/checkbox semantics, mounting ceilings, selection refs, Escape return, and drawer/grid geometry remain intact. | Typecheck passed at `.cartulary/test-results/20260831T192548Z-p8568`; regenerated routing passed at `.cartulary/test-results/20260831T192644Z-p13760`; focused Network Flow passed 38/38 at `.cartulary/test-results/20260831T192653Z-p16673`. Root `.cartulary/test-results/20260831T192557Z-p8991` failed only the intentionally characterized mapping light-fallback assertion; that title remains unrouted until the ordered mapping group removes the fallback, while the assertion remains in source. | No graph construction, query, selection, paging, or focus behavior changed. Rollback presentation substitutions only. Next: migrate saved-graph list, object controls, paging, and dialogs. |
| `2026-08-31T15:32:02-04:00` | NA-03 group 4 completion | Migrated saved-graph creation, reload, current-list state, rename/refresh/retire, notices, result retry, time-bucket/object paging, vertices/edges, contributor close, and create/rename/retire dialogs. Added graph-view-ID selectors, dense list markers, opaque scrim/dialog surfaces, responsive list/result layout, and token spacing without card presentation. | Typecheck passed at `.cartulary/test-results/20260831T193129Z-p22310`; focused Network Flow passed 38/38 at `.cartulary/test-results/20260831T193142Z-p22807`. | Immutable-result identity, mounting ceilings, stale-result presentation, selection focus, paging, mutations, and modal focus hooks are unchanged. Rollback the saved-panel presentation substitutions and selector together. Next: migrate mapping and indicator-link dialogs and activate the light-fallback boundary assertion. |
| `2026-08-31T15:36:51-04:00` | NA-03 group 5 completion | Migrated the mapping and indicator-link dialogs to opaque scrim/surface/elevation tokens, shared fields/selects/choices/actions, non-color invalid cues, typed monospace, responsive ordinal mapping rows, and explicit primary/secondary/pending hierarchy. Removed all hardcoded light-theme variables, hex fallbacks, raw rgba shadows, and local dialog framing. | Typecheck passed at `.cartulary/test-results/20260831T193604Z-p28509`; generation passed at `.cartulary/test-results/20260831T193625Z-p28991`; focused Network Flow, including the now-routed light-fallback and authored-token assertions, passed 38/38 at `.cartulary/test-results/20260831T193634Z-p31894`. | Preview/apply, ordinal mapping, exact-value link confirmation, modal focus containment, Escape, and return behavior are unchanged. Rollback dialog presentation and boundary routing together. Next: migrate semantic-grid menus/actions and the standalone measurement fixture. |
| `2026-08-31T15:38:31-04:00` | NA-03 group 6 completion / NA-04 start | Migrated semantic-grid column choices/reordering/reset, inspector close, and standalone load-fixture refresh/surface modes. The fixture mounts the same scoped style seam. A source audit now finds no raw visible button/input/select outside the seam; the remaining raw workspace input is the hidden CSV file input. | Typecheck passed at `.cartulary/test-results/20260831T193807Z-p36816`; focused Network Flow passed 38/38 at `.cartulary/test-results/20260831T193817Z-p37259`. | Semantic-grid adapter composition, viewport, row height, virtualization, gutters, density, selection/copy, and layout persistence were not changed. Rollback only the presentation wrappers. Next: add/strengthen responsive, accessibility, measurement, and visual evidence, then reconcile goldens. |
| `2026-08-31T16:38:30-04:00` | NA-04 completion / NA-05 start | Added base, narrow, compact, long-value, 200% zoom, text-spacing, modal, drawer, pending, disabled, and fixture-containment evidence. Corrected advanced-filter and column-menu bounds found by broad accessibility runs without changing document or grid geometry. Added eight Network Analysis capture intents and generated their manifest entries. Manually reviewed all eight retained PNGs; restored two unrelated workbook PNG rewrites produced by the updater and regenerated the manifest from the retained scope. | Targeted accessibility passed 11/11 at `.cartulary/test-results/20260831T200414Z-p92777`; targeted measurement passed 11/11 at `.cartulary/test-results/20260831T201658Z-p87462`; stateful passed 34/34 at `.cartulary/test-results/20260831T201802Z-p30448`; webserver-backed passed 60/60 at `.cartulary/test-results/20260831T202024Z-p79494`; visual update passed 12/12 at `.cartulary/test-results/20260831T202608Z-p36833`. The first ordinary proof root `.cartulary/test-results/20260831T203138Z-p81784` correctly rejected stale hashes after unrelated PNG restoration; `make generate` repaired only the manifest at `.cartulary/test-results/20260831T203418Z-p27629`. Ordinary visual proof then passed twice, 12/12, at `.cartulary/test-results/20260831T203431Z-p30672` and `.cartulary/test-results/20260831T203624Z-p74901`. | Product document overflow assertions pass at all three viewports. The wider measurement debug harness reports its own outer document dimensions but the production fixture is bounded and scrolls only inside its owned work area; this is support-harness framing, not product behavior. No golden changes alter product state or hide controls. Rollback the evidence, authored fixture entries, generated manifest, and eight PNGs with the presentation slice. Next: rerun complete routed, broad, generation, policy, and finalization checks and audit final Git scope. |
| `2026-08-31T17:05:40-04:00` | NA-05 completion | Refreshed all four owner routes, ran every focused route they returned, reran requested broad frontend/browser checks, inspected generated diffs, and audited final source scope. No Go, backend, route, dependency, authorization, persistence, authored token, or generated token file changed. The only raw Network Analysis input outside the seam is the intentionally hidden CSV file picker. | Final retained results are listed in section 7. Full accessibility passed 12/12, measurement 22/22, stateful 34/34, webserver-backed 60/60, frontend unit 392/392, and `test-fast` 444/444. `make agent-finalize` passed with `RESULTS_DIR` unset because no single retained full warm `make check` root was used. | Compatibility remains additive and internal. Rollback all listed source, UI-contract facade, harness input/generated manifest, eight goldens, and this handoff together; no data or operational rollback is required. No TODO or IN_PROGRESS workstream remains. Next: hand off the completed bounded refactor. |

## 6. Visual reconciliation record

The only accepted trigger is this Network Analysis presentation refactor. The
pinned updater ran through `make browser-e2e-visual-update`; the renderer
contract remained
`visual.renderer.playwright_1_59_1_chromium_1217_linux_amd64`.

Retained Network Analysis goldens:

- base desktop: mapping dialog, accepted inspector, rejected diagnostics,
  graph contributors, saved-graph result, and destructive delete dialog;
- narrow desktop: query controls and advanced-filter surface at 1024x720;
- compact desktop: saved-graph interaction at 768x640.

Manual inspection found opaque dialog/popover surfaces, valid elevation,
visible focus, clear primary/secondary/destructive hierarchy, geometric tab and
selection markers, internally bounded grid/drawer scrolling, and reachable
actions at narrow and compact viewports. No capture contains browser-default
white controls, overlap, product-state drift, or a hidden required action.

The updater also rewrote the unrelated
`workbook-inspector-public-error-linux.png` and
`workbook-inspector-rollback-preview-linux.png`; both were restored to their
clean baseline. This intentionally invalidated their generated hashes, so the
ordinary visual run at `.cartulary/test-results/20260831T203138Z-p81784`
failed only manifest reconciliation while both screenshot suites passed. A
subsequent Make-owned generation recomputed the manifest from the retained PNG
scope. Ordinary visual validation then passed twice against that same manifest
at `.cartulary/test-results/20260831T203431Z-p30672` and
`.cartulary/test-results/20260831T203624Z-p74901`.

## 7. Final validation and owner-to-change map

| Owner | Change/evidence boundary |
| --- | --- |
| `web.networkflow` | Local React control seam, workspace migration, boundary/unit tests, source ownership. |
| `module.networkflow` | Browser evidence only; no API, query, graph, lifecycle, authorization, or persistence implementation changed. |
| `web.design` | Existing dark graphite tokens and observable focus/responsive behavior; no authored token added. |
| `package.ui` | Additive typed token-reference helper and semantic selector identities; no React component or generated token edit. |

Final retained Make results:

| Command | Result and retained run root |
| --- | --- |
| `make test-slice OWNER=package.ui` | 10/10 pass, `.cartulary/test-results/20260831T204542Z-p22012` |
| `make test-slice OWNER=web.networkflow` | 38/38 pass, `.cartulary/test-results/20260831T204550Z-p22439` |
| `make test-slice OWNER=web.design` | 15/15 pass, `.cartulary/test-results/20260831T204608Z-p26848` |
| `make service-backed-test-slice OWNER=web.design` | 15/15 pass, `.cartulary/test-results/20260831T204927Z-p31925` |
| `make test-slice OWNER=module.networkflow` | 34/34 pass, `.cartulary/test-results/20260831T204716Z-p72924` |
| `make service-backed-test-slice OWNER=module.networkflow` | 28/28 pass, `.cartulary/test-results/20260831T205034Z-p77993` |
| `make frontend-typecheck` | 2/2 pass, `.cartulary/test-results/20260831T205245Z-p36593` |
| `make frontend-unit` | 392/392 pass, `.cartulary/test-results/20260831T205258Z-p37113` |
| `make frontend-import-boundary-check` | 2/2 pass, `.cartulary/test-results/20260831T205423Z-p70031` |
| `make lint-biome` | 2/2 pass, `.cartulary/test-results/20260831T205431Z-p70459` |
| `make browser-e2e-a11y` | 12/12 pass, `.cartulary/test-results/20260831T203931Z-p19899` |
| `make browser-e2e-measurement` | 22/22 pass, `.cartulary/test-results/20260831T204059Z-p64088` |
| `make browser-e2e-stateful` | 34/34 pass, `.cartulary/test-results/20260831T205530Z-p77776` |
| `make browser-e2e-webserver-backed` | 60/60 pass, `.cartulary/test-results/20260831T205749Z-p27174` |
| `make browser-e2e-visual` twice | 12/12 pass at `.cartulary/test-results/20260831T203431Z-p30672` and `.cartulary/test-results/20260831T203624Z-p74901` |
| `make generate` | Pass, `.cartulary/test-results/20260831T205445Z-p70884` |
| `make generate-drift` | 4/4 pass, `.cartulary/test-results/20260831T205458Z-p73783` |
| `make generated-artifact-policy-check` | 3/3 pass, `.cartulary/test-results/20260831T205510Z-p76798` |
| `make json-shape-check` | 3/3 pass, `.cartulary/test-results/20260831T205516Z-p77238` |
| `make test-fast` | 444/444 pass, `.cartulary/test-results/20260831T210442Z-p84165` |
| `make agent-finalize` | 1/1 pass, `.cartulary/test-results/20260831T210458Z-p84861` |
| `make lint-markdown` | Pass, `.cartulary/test-results/20260831T210707Z-p88209` |

The final generated diff contains only the topology render index derived from
the authored `web.networkflow` family, plus the visual golden manifest derived
from the authored fixture registry and retained PNGs. Generated token output is
unchanged. Generation, drift, artifact policy, and JSON shape checks all pass.

Go security checks were skipped because the final diff is limited to frontend,
UI-contract facade, harness inputs/generated manifests, browser evidence,
goldens, and documentation. It contains no Go, route, authorization, dependency,
backend, or production service change.

## 8. Compatibility, rollback, risks, and deferrals

This is a Network Analysis presentation, accessibility, token-conformance, and
internal frontend-component refactor. It introduces no public API, route,
authorization, extension-profile, graph-semantic, storage, persisted-data, or
migration change.

Rollback is source-only: revert the Network Analysis presentation seam,
additive UI-contract exports and selectors, scoped tests/harness inputs and
generated outputs, eight reconciled Network Analysis goldens, and this handoff
together. No data rollback, migration reversal, or external cleanup is needed.

The test-only load harness owns a wider diagnostic shell than the production
fixture at compact widths. Measurement records that shell but proves the
Network Flow fixture remains bounded and scrolls internally; production
workspace tests separately prove no document-level horizontal or block
overflow. No product risk, behavior deferral, or owner contradiction remains.
