# Visual Renderer and Density Remediation Handoff

## Baseline

- Working branch: `main`
- Integration commit: `ba067d5d15c6d06adfce9be3b25e51bb1ab241c6`
- Clean comparison commit: `e894fbfefff1fcf172ce14b6f8f290f6e216a0dc`
- Initial Git state: clean; `main` is one commit ahead of `origin/main`. The existing workbook recovery slice is the committed integration commit and is preserved.
- Tool versions: Git 2.53.0, Node.js 24.15.0, pnpm 10.33.0, Python 3.14.4, Docker 29.5.2, GNU Make 4.4.1.
- Pinned renderer image state: absent locally at baseline; acquisition is part of VR-02.
- Authorized paths: Testing Harness owner text, visual guide and this/existing handoffs; authored harness contracts, schemas, policies, generator/runtime/tests and generated outputs; Playwright configuration and visual tests/goldens; `packages/grid-adapter` implementation/tests; the incident-administration browser test; narrowly required task-routing projections.
- Compatibility boundary: no public API, route, authorization, storage, persisted-data, or migration change.

## Workstream tracker

| Workstream | Status | Checkpoint |
| --- | --- | --- |
| VR-00 — Baseline and tracker | DONE | Both independent failures reproduced at clean parent commit. |
| VR-01 — Owner and machine-contract cleanup | DONE | Owners, current/historical schemas, profile, manifest, and policy agree. |
| VR-02 — Hermetic visual harness | DONE | Pinned remote renderer, provenance, scratch update, cleanup, and tests implemented. |
| VR-03 — Density implementation | DONE | Owned header/selection/gutter/action wrappers pass package and browser density checks. |
| VR-04 — Golden reconciliation | DONE | Atomic refresh reviewed; two fresh ordinary visual runs pass. |
| VR-05 — Integration, validation, and handoff | DONE | Integrated evidence, late-failure transaction audit, and final scope checks pass. |

## VR-00 — Baseline and tracker

- Paths: repository metadata and retained `.cartulary/test-results` evidence; no product source changes.
- Decision: use a detached worktree at the parent commit. Do not stash, reset, amend, or overwrite the committed recovery slice.
- Known retained evidence before reproduction:
  - Visual: `.cartulary/test-results/20260831T144805Z-p47185`; 17 untouched snapshot mismatches across unrelated owners, generally about one percent of pixels and concentrated at glyph edges. Recovery/resolver captures passed.
  - Webserver-backed: `.cartulary/test-results/20260831T144112Z-p68949`; density test expected `2px 5px` and observed `0px 5px 2px`.
- Risk: a fresh detached worktree may require local dependency/bootstrap materialization. Any infrastructure-only failure will be distinguished from the already-retained product failures.
- Rollback: remove the detached worktree and this handoff file; no product or stored-data rollback.
- Next action: run the two public browser targets against `e894fbfefff1fcf172ce14b6f8f290f6e216a0dc` and record exact roots/results.

### Completion checkpoint

- Commands:
  - `git worktree add --detach /tmp/cartulary-vr-baseline.CheOJy/clean e894fbfefff1fcf172ce14b6f8f290f6e216a0dc`
  - `make browser-e2e-visual`
  - `make browser-e2e-webserver-backed`
- Visual result: failed independently at clean commit; retained root `.cartulary/test-results/20260831T162126Z-p68748`, 9/12 graph units. Seventeen expected PNGs differed across Workbook, Timeline, Collaboration, Evidence, Entities, Saved Views, and Network Flow. Differences were approximately one percent and repeated on glyph edges; no recovery-surface source existed at the clean comparison commit.
- Webserver-backed result: failed independently at clean commit; retained root `.cartulary/test-results/20260831T162344Z-p19789`, 58/60 graph units. `updates workbook density from Account Settings while the workbook remains open` observed header padding `0px 5px 2px` instead of `2px 5px` at `incident-administration.spec.ts:479`.
- Setup note: the first visual invocation stopped during preflight because the detached worktree lacked untracked Node dependencies (`ajv/dist/2020`). The clean worktree then used the existing dependency tree; this setup-only failure produced no test run and is not a blocker.
- Diagnostics: Playwright 1.59.1, Chromium 147.0.7727.15/revision 1217, font manifest `c21f8663e6c8fe72681b2be644aa8398538afc59a0f0cda06b94d46d5fbba5fe`. The approved image was absent locally.
- Compatibility: baseline-only work; no runtime or contract change.
- Remaining risk: committed PNGs have no renderer provenance, so the exact environment that produced them cannot be proven.
- Next action: amend TH-HARNESS-REQ-255 and design §3.9; introduce schema-closed renderer/profile/golden/reconciliation-v2 inputs and register them with harness policy.

## VR-01 — Owner and machine-contract cleanup

- Paths: `docs/testing-harness-nlspec.md`, `docs/design.md`, the visual maintenance guide, `tools/frontend_visual_renderer_profile.json`, the generated golden manifest, three current visual schemas, historical reconciliation-v1 schema/registry, schema attachments, generated-artifact and drift inputs, task-surface owner/projections.
- Decisions:
  - Reconciliation v2 is the sole current visual artifact. V1 moved unchanged to diagnostic-only historical storage with its exact SHA-256.
  - The profile fixes the approved image, platform, Playwright/Chromium identities, font-manifest digest, locale, scale, and color scheme. Capture-owned viewport, zoom, and reduced-motion remain separate.
  - `cellPadding` now normatively means the effective logical inset of rendered content; vendor structural padding cannot replace it.
  - The golden manifest is generated and exhaustively contains all 64 committed PNG paths and SHA-256 values; it duplicates no catalog ownership.
- Commands/results:
  - `make playwright-install` — PASS, `.cartulary/test-results/20260831T164638Z-p88175`; pulled and verified digest-pinned `linux/amd64` image.
  - `make generate` — PASS, `.cartulary/test-results/20260831T165659Z-p94343`; generated manifest and topology/task-surface projections.
  - `make json-shape-check` — PASS, `.cartulary/test-results/20260831T165713Z-p97276`; validates profile, exhaustive manifest identity, schemas, and historical registry.
- Compatibility: current retained visual reconciliation changes from v1 to v2; v1 remains readable only as historical diagnostics. Public Make targets and catalog rows are unchanged.
- Risks: the renderer lease and update promotion still require execution/negative-path coverage in VR-02.
- Rollback: revert the owner text, authored profile/schemas/policies, historical move, generator, generated manifest, and generated task/topology outputs together.
- Next action: finish the private renderer lease integration, fail-closed reconciliation, rollback-safe scratch promotion, and harness contract tests.

## VR-02 — Hermetic visual harness

- Paths: Playwright readiness/configuration, browser group runner/finalizer/work graph, renderer lease, golden-manifest generator, reconciliation v2 producer/schema, harness contract tests, visual capture intents, generated topology/task projections.
- Decisions:
  - Copy the exact pinned local Playwright 1.59.1 server package into the digest-pinned container; do not mount the repository and do not download runtime packages.
  - Publish an unguessable endpoint on host `127.0.0.1` and use Playwright `<loopback>` exposure for the existing local services. This is portable across native Docker and Docker Desktop, where container loopback is not host loopback.
  - One renderer lease is scoped to each visual group. Image/platform/package/browser mismatches fail before test execution. SIGINT, SIGTERM, success, and failure converge on forced container cleanup; server logs are never retained.
  - Updates copy all goldens into run-scoped scratch. The finalizer creates and validates a candidate manifest/reconciliation before rollback-safe directory-and-manifest promotion.
- Commands/results:
  - `make test-slice OWNER=harness.browser` — PASS, 28/28, `.cartulary/test-results/20260831T165907Z-p98499`.
  - First lifecycle probe `.cartulary/test-results/20260831T170018Z-p18621` failed closed because Playwright's default server path was `/`; fixed by supplying an unguessable path.
  - Second lifecycle probe `.cartulary/test-results/20260831T170248Z-p74486` failed closed because Docker Desktop isolates container loopback even in host mode; fixed with explicit host-loopback publication and Playwright loopback forwarding.
  - Ordinary pinned visual validation `.cartulary/test-results/20260831T170528Z-p19283` reached all visual scenarios and reproduced the previously classified glyph-edge diffs. Reconciliation v2 recorded two matching renderer attestations and a passing 64-entry manifest identity; no renderer container leaked.
  - `make harness-contract` — PASS, 2/2, `.cartulary/test-results/20260831T170748Z-p64397`.
- Compatibility: nonvisual browser targets continue to use host Chromium. Public target names, row selectors, screenshot tolerances, masks, and scopes are unchanged.
- Risks: promotion is exercised by the authorized refresh in VR-04; until then, candidate-to-tracked promotion has contract coverage but no live successful transaction evidence.
- Rollback: revert renderer/profile/provenance code and generated outputs; no service or data state survives lease cleanup.
- Next action: introduce adapter-owned density wrappers and replace outer-cell shorthand assertions with effective logical-side assertions.

## VR-03 — Density implementation

- Paths: `packages/grid-adapter/src/rdgCompiler.tsx`, `packages/grid-adapter/src/styles.css`, package unit coverage, and the existing incident-administration density scenario.
- Decisions:
  - Standard, selection, gutter, and actions headers now render adapter-owned content wrappers. Selection and gutter body cells use the same inner-content ownership pattern.
  - Outer header cells reset only Cartulary inline/end padding; RDG's inline `padding-block-start` remains available for row-span positioning. No `!important` or vendor-private class selector was introduced.
  - Browser evidence reads the owned wrapper's four logical sides individually; it no longer asserts browser shorthand serialization on a structural cell.
- Commands/results:
  - `make task-guide ROLE=module-author OWNER=package.grid_adapter` — focused routes confirmed.
  - `make test-slice OWNER=package.grid_adapter` — PASS, 38/38, `.cartulary/test-results/20260831T171256Z-p66576`.
  - `make service-backed-test-slice OWNER=package.grid_adapter` — PASS, 13/13, `.cartulary/test-results/20260831T171401Z-p14776`; includes `updates workbook density from Account Settings while the workbook remains open`.
- Compatibility: internal DOM wrappers and density pixels change; roles, accessible names, test IDs, sorting, selection, row height, editor fill, and product interfaces remain intact.
- Risks: wrapper pixel shifts must be reviewed in the pinned-renderer golden refresh; no behavioral regression is present in focused evidence.
- Rollback: revert wrapper markup/CSS and logical-side assertions together.
- Next action: run one ordinary pinned visual baseline, refresh only classified differences through scratch promotion, inspect every PNG, then pass two ordinary runs against one manifest.

## VR-04 — Golden reconciliation

- Paths: the 33 modified PNGs under `apps/web/e2e/workbook.visual.spec.ts-snapshots`, generated `tools/frontend_visual_golden_manifest.json`, v2 reconciliation and renderer attestations under retained run roots, and the visual maintenance guide.
- Decisions:
  - Three ordinary runs were used before refresh while isolating the owned density change. The initial wrapper flex alignment shifted selection content; that was rejected and removed. The accepted wrapper preserves normal inline formatting and changes only the effective density inset.
  - The accepted refresh trigger was **pinned renderer adoption and owned density inset correction**. Tolerances, screenshot scopes, masks, offsets, fixture identities, viewport/zoom settings, and product state were unchanged.
  - All 33 old/current/diff images were reviewed in contact sheets and by exact pixel counts. The changes were glyph-edge rasterization plus the intentional header/gutter/actions inset; no semantic, fixture-state, loading, focus, masking, clipping, or unexpected layout change was accepted.
- Pre-refresh evidence:
  - `.cartulary/test-results/20260831T171532Z-p58889` exposed and rejected selection alignment drift.
  - `.cartulary/test-results/20260831T171830Z-p3893` exposed and rejected flex-induced vertical baseline drift.
  - `.cartulary/test-results/20260831T172110Z-p49742` was the final classified ordinary run; only renderer edges and the owned inset remained.
- Refresh result: `make browser-e2e-visual-update` — PASS, 12/12, `.cartulary/test-results/20260831T172348Z-p95405`. The v2 artifact reconciled 64 intents, 64 committed goldens, 22 fixtures, two exact renderer attestations, and a complete candidate manifest with zero missing, orphan, ambiguous, or unresolved entries. Scratch promotion completed and no renderer container leaked.
- Reviewed changed goldens:
  - Collaboration: `collaboration-conflict-resolver-linux.png`, `collaboration-grid-blocked-conflict-linux.png`, the base/compact/narrow `collaboration-grid-conflict-resolver` variants, and both collaboration presence-marker captures.
  - Entity/evidence/relationships: `entity-mention-chip-states-linux.png`, both evidence timeline/grid captures, and `record-relationships-mention-chips-linux.png`.
  - Workbook shells: the base/compact/narrow incident-directory captures.
  - Network Flow: accepted-inspector, graph-contributors, and rejected-diagnostics captures.
  - Timeline grid/mutation: active-edit, conflict-strip, grouped/default grid, active-edit/empty/pending mutation, and base/compact/narrow transaction-recovery captures.
  - Inspector/query: all four inspector captures plus empty and saved-view query captures.
- Repeatability results:
  - `make browser-e2e-visual` — PASS, 12/12, `.cartulary/test-results/20260831T173059Z-p44513`.
  - `make browser-e2e-visual` — PASS, 12/12, `.cartulary/test-results/20260831T173257Z-p89397`.
- Compatibility: one reviewed provenance-adoption refresh; no public Make target, catalog row, fixture identity, capture contract, product interface, or runtime behavior changed.
- Risks: Docker image availability remains an explicit fail-closed prerequisite. The image digest and manifest make any future renderer or PNG change an intentional review event.
- Rollback: revert the renderer/harness change, wrappers, all 33 PNGs, and generated manifest as one source transaction.
- Next action: run integrated task routing and the complete VR-05 validation matrix, then update the recovery handoff only if both former blockers and all task-owned evidence pass.

## VR-05 — Integration, validation, and handoff

- Paths: all remediation sources and generated artifacts, both handoffs, and the final Git scope; no runtime dependency reads either handoff.
- Decision: do not reuse the baseline failure roots as completion evidence. Current integrated evidence was obtained with the recovery commit present, and the recovery handoff was moved from `BLOCKED` to `DONE` only after both exact broad blockers passed.
- Owner routing/results:
  - Focused: `harness.browser` 28/28 `.cartulary/test-results/20260831T173621Z-p35148`; `package.grid_adapter` 38/38 `.cartulary/test-results/20260831T173727Z-p54710`; `module.workbook` 68/68 `.cartulary/test-results/20260831T173829Z-p1524`; `web.design` 15/15 `.cartulary/test-results/20260831T174043Z-p59348`; `web.workbook` 139/139 `.cartulary/test-results/20260831T174147Z-p5808`.
  - Service-backed: `harness.browser` 6/6 `.cartulary/test-results/20260831T174233Z-p20862`; `package.grid_adapter` 13/13 `.cartulary/test-results/20260831T174336Z-p38152`; `module.workbook` 39/39 `.cartulary/test-results/20260831T174435Z-p81896`; `web.design` 15/15 `.cartulary/test-results/20260831T174643Z-p37191`.
- Contract/generation results: `make generate` `.cartulary/test-results/20260831T174753Z-p83182`; drift 4/4 `.cartulary/test-results/20260831T174806Z-p86057`; generated policy 3/3 `.cartulary/test-results/20260831T174818Z-p89082`; JSON shape 3/3 `.cartulary/test-results/20260831T174823Z-p89517`; catalog and semantic identity passed without retained output; harness contract 2/2 `.cartulary/test-results/20260831T174843Z-p90734`.
- Frontend results: typecheck 2/2 `.cartulary/test-results/20260831T174901Z-p91308`; unit 391/391 `.cartulary/test-results/20260831T174915Z-p91828`; import boundary 2/2 `.cartulary/test-results/20260831T175011Z-p12403`; final Biome 2/2 `.cartulary/test-results/20260831T175048Z-p17808`. Initial Biome root `.cartulary/test-results/20260831T175018Z-p12824` identified formatting/import organization in three touched files; `make format` passed at `.cartulary/test-results/20260831T175040Z-p13588`, and no rule was weakened.
- Browser results: measurement 22/22 `.cartulary/test-results/20260831T175057Z-p18308`; final a11y 12/12 `.cartulary/test-results/20260831T175931Z-p19043`; stateful 34/34 `.cartulary/test-results/20260831T180102Z-p63163`; webserver-backed 60/60 `.cartulary/test-results/20260831T180326Z-p12511`; final visual 12/12 `.cartulary/test-results/20260831T180918Z-p69043`.
- Failure classification: a11y root `.cartulary/test-results/20260831T175528Z-p75813` failed 10/12 because the separate Network Flow group timed out starting test services. Its 12 workbook/design/grid product rows all passed; the immediate full rerun passed 12/12. No task assertion failed.
- Late transaction audit: final review found that the first implementation promoted tracked files before the final target-result write. Promotion now occurs only after reconciliation and target schemas and artifacts pass. A real-filesystem injected fourth-rename failure proves old PNG and manifest bytes are restored; the success case proves the paired swap and backup cleanup. Post-hardening `harness.browser` passed 28/28 at `.cartulary/test-results/20260831T182354Z-p64962`, harness contract 2/2 at `.cartulary/test-results/20260831T182505Z-p84708`, and ordinary visual 12/12 at `.cartulary/test-results/20260831T182527Z-p85989`.
- Final broad results: after `make generate` `.cartulary/test-results/20260831T182803Z-p31393`, JSON shape passed `.cartulary/test-results/20260831T182821Z-p34408`, generation drift passed `.cartulary/test-results/20260831T182821Z-p34405`, generated policy passed `.cartulary/test-results/20260831T182821Z-p34427`, `make agent-finalize` passed `.cartulary/test-results/20260831T182833Z-p38229`, and `make test-fast` passed 443/443 `.cartulary/test-results/20260831T182851Z-p41161`.
- Finalization diagnostics retained: `.cartulary/test-results/20260831T182724Z-p30329` and `.cartulary/test-results/20260831T182748Z-p30920` failed closed because the newly added promotion helper made generated topology inputs stale. The prescribed `make generate` corrected the projection; no generated file was hand-edited.
- Compatibility: remains limited to visual-test infrastructure, density presentation, and internal component boundaries. No public API, route, authorization, storage, persisted-data, or migration change.
- Risks: the digest-pinned Docker image must be available; its absence or any attestation/manifest drift now fails closed. A single non-task Network Flow startup timeout was observed and resolved on immediate rerun.
- Rollback: source-only; no migrations or stored state exist.
- Next action: hand off the completed source-only change. No tracker row remains `TODO`, `IN_PROGRESS`, or `BLOCKED`.

## Compatibility and rollback

- Final compatibility statement: **This remediation changes visual-test infrastructure, workbook density presentation, and internal component boundaries only. It introduces no public API, route, authorization, storage, persisted-data, or migration change.**
- Rollback is source-only: revert harness owner/contracts/profile, renderer lifecycle, generated provenance, grid wrappers/CSS/tests, goldens, and both handoffs together. No data rollback is required.
