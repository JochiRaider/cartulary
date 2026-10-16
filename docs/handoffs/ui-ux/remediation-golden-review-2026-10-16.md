# Remediation visual review — 2026-10-16

Status: reviewed transactional refresh and both required ordinary passes passed.
This is supporting evidence, not product acceptance or Core 05 qualification.

## Reviewed source and reconciliation

The ordinary `remediation-coordinated-release-20261016-01` visual projection
reconciles 255 active capture intents and 255 committed goldens, with no missing,
ambiguous or orphaned mapping. All 29 registered fixtures resolve. The projection
failed because ten screenshot comparisons in two rows changed. Other release
failures remain independent and open.

The exact original-commit inspector row passed in
`remediation-original-inspector-visual-20261016-01` at
`00d24aa5354284dfcf6533b8f650e9eb0291ad44`. Current and original renderer pins,
fonts, styles, viewport and device scale match. Current Evidence metadata and
verification presentation changes the height of preceding inspector content;
retained History, rollback and error anchors differ by about 0.11 CSS pixels.

The artifacts-only session `remediation-release-visual-review-20261016-02`
imported all ten exact canonical capture identities. Actual, expected and diff
components were checked against their byte lengths and SHA-256 digests and
visually inspected at native detail. No DOM or axe evidence was available.
The session closed with outer exit 0, `cleanup=complete`, and 30 imported images.
Caller-owned request scratch was removed. Private images are not handoff links.
The first import encountered capacity contention; the import succeeded after
the concurrent browser-harness verification completed. One subsequent request
used a nonexistent caller filename, failed without an operation, and was
corrected before import. Neither failure changes the source visual result.

## Accepted refresh scope

`web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc`
owns these five active, nonregistry captures:

- `account-menu-controls-narrow-linux.png`
- `account-menu-controls-compact-linux.png`
- `account-menu-controls-short-linux.png`
- `account-menu-long-label-zoom-linux.png`
- `account-menu-long-label-text-spacing-linux.png`

The accepted trigger is stale comparison input relative to the validated
Reporting control entry. “Snapshots and reports” appears between Import and
Promoted fields. The taller scrollable menu retains the focused Membership audit
item within each constrained viewport, including zoom and text spacing. Other
items remain reachable by scrolling. The reviewed differences are insertion and
resulting scroll placement; no unrelated shell, font or color change was found.

`module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`
owns these five registered captures:

| Golden filename | Stable fixture |
| --- | --- |
| `workbook-inspector-history-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-rollback-preview-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-public-error-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-narrow-technical-details-linux.png` | `visual.fixture.inspector_narrow_technical_details` |
| `workbook-inspector-compact-actions-linux.png` | `visual.fixture.inspector_compact_actions` |

The accepted trigger is stale comparison input after validated Evidence panel
presentation changes. The observed text shifts are consistent with the changed
scroll-content geometry; History text, source values, rollback controls, public
error, role-disabled actions, technical details and focus remain present. The
diffs contain small text rasterization/position changes, not missing content or
new overflow. Existing top-edge clipping from intentional scroll positioning
is unchanged in kind. The original baseline comparison supplies a control for
the current-source delta, not evidence of intermittent behavior.

All filenames are under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
Viewport, browser zoom, masks, scroll normalization, screenshot scope, renderer
pins and comparison tolerances are unchanged. No pixel correction is applied.
Only Make's transactional update may publish candidates, and it must pass the
functional assertions and full reconciliation first. Any additional changed
image requires review before acceptance.

## Validation

`remediation-reviewed-golden-update-20261016-01` ran
`make browser-e2e-visual-update` and passed with outer exit 0 in 1,001.044 seconds:
97/97 work units and all 194 cleanup operations complete. The final reconciliation
passes with the same 255 active images, zero missing/ambiguous/orphaned mappings
and 29 resolved registered fixtures. Only the ten listed PNGs changed. The
promoted manifest SHA-256 is
`00bfbafe887edeee7f9ed072146f2ed7dd62a9e61d70d5af15c514234a23eca3`.

All ten promoted images were imported, digest-checked and individually inspected
again in `remediation-promoted-visual-review-20261016-01`. Their final appearance
matches the accepted triggers above; no additional layout or content change was
found. This explicit-image review is reference-only, with no DOM/axe claim. It
closed with outer exit 0 and complete cleanup; caller-owned scratch was removed.

`remediation-post-golden-finalize-20261016-01` passed `make agent-finalize` with
outer exit 0 in 22.203 seconds. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset. Record two subsequent ordinary
`make browser-e2e-visual` passes against the promoted manifest before closing
this refresh.

The first ordinary `make browser-e2e-visual` invocation,
`remediation-promoted-visual-pass-20261016-01`, passed with native exit 0 in
1,116.214 seconds, 97/97 work units and all 194 cleanup operations complete.
Its source is `700d45dd9e4772407495836300b4525ac4887972c792d3f487d19319bd8c3815`.
The subsequent retained-recipient validation repair affects Reporting worker
policy consumption and the package test observer; no frontend, visual fixture,
capture policy or golden bytes change. The second ordinary pass must retain
its own source identity and use the same promoted manifest.

The second ordinary invocation, `remediation-promoted-visual-pass-20261016-02`,
passed with native exit 0 in 943.371 seconds: 97/97 work units and all 194 cleanup
operations complete. Its source is
`4d519ca948da4d9daa88f7d6c14e0fa8bb66f07e5a2fc8199a769bbe58399708` and graph
`5b6951bc0cf6ca8acf7aa5b469d466f00c7a1fbc029c4d55cb16afeed9c0c145`.
The manifest SHA-256 remains exactly the promoted value above. Both ordinary
roots retain `browser-e2e-visual/frontend-visual-reconciliation.json` and pass
complete capture/fixture reconciliation. This closes the reviewed golden refresh;
it does not close the separate integrated release or cost qualification.
