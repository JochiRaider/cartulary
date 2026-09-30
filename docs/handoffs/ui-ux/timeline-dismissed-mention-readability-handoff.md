# Readable dismissed Timeline mentions

Date: 2026-09-30. Status: source implementation and focused behavior acceptance
pass; visual acceptance is blocked by unrelated baseline differences. No commit,
push, or deployment was performed.

## Baseline, authority, and boundary

The checkout started clean on `main` at
`5e26c3f47db687f485ec67ae1520f88f34c16506`, one commit ahead of `origin/main`.
No unrelated work was present. Branch and HEAD remain unchanged. This task
preserves that commit's mention focus reveal and preceding collection behavior.
The applicable root AGENTS.md and current Make help/task guides were read.

Core 00 establishes precedence. Core 03 §9, REQ-03-129–134, owns explicit
resolve/dismiss/restore, raw source preservation, exclusion from active
collections/unresolved counts, and Restore without prior-target relinking.
Core 03 §2.3A, REQ-03-219, REQ-03-283, and REQ-03-299/100 retain Inspector,
draft, focus, operation, and scoped authority lifetimes. Domain supplies
vocabulary and owner navigation. Research, historical handoffs, bundled prompts,
and the digest were advice/evidence, without becoming authorization or executable
dependencies.

The authorized design correction reconciles design §12.2's former tertiary
dismissed foreground with §14.1's 4.5:1 ordinary-text requirement and §14.3's
closed accepted pairing matrix. `chip.dismissed` now uses existing
`colors.ink-muted`. `cp.muted.surface1` explicitly covers inspectable dismissed
mention labels and markers and adds `D-VFIX-004` coverage. §9.2's low emphasis,
transparent background, hairline border, visible marker, inspection behavior,
and closed four-state model remain intact. No disabled-text exemption was used.

Source ownership is `web.workbook`; verification routing is independently owned
by the authored `web.workbook`, `module.entities`, `module.workbook`, and visual
family manifests. The shared chip already owns the common state-to-style
decision, so the correction uses that boundary without a new renderer,
abstraction, theme, or state owner. Shared Timeline grid and entity Inspector
consumers were inspected. Their active-collection filtering remains unchanged.

The selection-rubric finding is a confirmed readability defect during dismissal
and inspection. Existing shared styling is the remediation seam. Keeping the
accepted token binding in one place prevents the same defect in another consumer;
future callers can use the existing chip. Leaving the gap would preserve unreadable
informational text. Retired bindings are the dismissed tertiary foreground and
the outer-button assertion used as dismissed-text contrast evidence. Existing
selection, focus, inspection, source, operation, and recovery owners remain useful
and retain their responsibilities. There is no API or data migration.

Both `cartulary-ui-ux-refactor` and `cartulary-ui-review` were applied. The requested
digest README, START_HERE, LOCAL_AGENT_PROMPT, maps, rules, acceptance and query
recipes were consulted. Advice to use contrast and non-color cues was adopted
within the existing owner profile; palette/framework replacement was rejected.
There is no separate controlling tracker for this slice.

## Changed and inspected paths

Paths below are relative to `/home/jochi/code/cartulary`.

- `docs/design.md`: the two bounded owner table corrections above.
- `apps/web/src/workbook/components/WorkbookRelationshipChip.tsx`: only the
  dismissed foreground binding changes. Background, border, ring, text and event
  handling are unchanged.
- `apps/web/src/workbook/components/WorkbookRelationshipChip.test.tsx`: the
  existing scenario now covers selected/unselected dismissed Host and Identity
  variants, exact raw text/names, lowercase marker, muted binding, transparent
  background, hairline border, and selection ring.
- `apps/web/e2e/workbook.a11y.spec.ts`: extends the existing entity accessibility
  scenario and contrast helper with required label/marker descendants, diagnostics
  before pass assertions, unique context identifiers, opaque ancestor-background
  resolution, and explicit rejection of missing/hidden/ambiguous/unparseable or
  unsupported alpha/image/opacity/filter/blending paint. It removes the incorrect
  dismissed outer-button target. Optional legacy targets keep their old behavior.
  It covers both entities, long source inspection, five layouts, and closure in
  the same runtime followed by reopening the Inspector.
- `apps/web/e2e/mentions.lifecycle.spec.ts`: extends the existing scenario to both
  entity types; checks dismissal exclusion, preserved raw source, empty resolution
  metadata, and explicit unresolved restoration without prior-target relinking.

Inspected and unchanged: `WorkbookShell.mentionChips.test.ts`,
`models/workbookRelationshipChip.ts`, `timeline/models/workbookMentionChips.ts`,
`TimelineMentionsPanel.tsx`, `TimelineCollectionCell.tsx`,
`features/entities/useEntityWorkbookInspectorComposition.tsx`,
`TimelineMentionActionControls.tsx`, `useTimelineMentionActions.ts`,
`timelineMentionOperationModel.ts`, `createTimelineMentionResolutionAdapter.ts`,
`e2e/support/entities/mentions.ts`, and `e2e/workbook.visual.spec.ts`.
Captured operation identity, duplicate admission, uncertain replay,
acknowledgement/refresh and retained recovery ownership have no source changes.
Endpoints, schemas, persistence, queries, payloads, discovery, History and draft
lifetimes are unchanged. View-model coverage needs no extension because its
behavior did not change. Existing test titles/catalog identities are preserved.
No typed token or state projection needs generation for this substitution.

## Rendered evidence and acceptance

The pre-fix required-descendant regression failed as intended. Both Host label
and lowercase marker computed `rgb(100, 116, 139)` (`#64748B`) over actual
Inspector paint `rgb(17, 19, 24)` (`#111318`): **3.90:1**, below 4.5:1.
The transparent text/chip/button/collection/section/panel ancestors were recorded
through the opaque `timeline-inspector` background. The prior outer-button
measurement was 17.76:1 and did not represent the nested text.

The passing expanded scenario records **29 distinct contexts / 58 required text
measurements**. Every label and marker uses `rgb(203, 213, 225)` (`#CBD5E1`)
over the same actual background, measuring **12.52:1**. Each target must exist
exactly once and yield a measurement; missing targets cannot disappear silently.

| Acceptance | Result and evidence |
| --- | --- |
| Host and Identity dismissal and inspection | PASS; selected and unselected descendants at 1440×900 and 1024×720, names and non-color marker retained |
| Long label / full source | PASS; 150-character Identity source remains exact and fully inspectable in Mention details |
| Enlargement, reflow, spacing | PASS; 1024×720 at 200%, 320×720 vertical content, and simultaneous 1.5 line height / .12em letter / .16em word spacing / 2em paragraph spacing; no Inspector horizontal overflow |
| Ordinary Restore | PASS for both entity types; unresolved raw source, empty resolution metadata, no restored prior link |
| Active collections and unresolved indicator | PASS; dismissed entries excluded, `timeline.has_unresolved_mentions=false` |
| Arrow focus reveal and draft retention | PASS; existing focus guard includes 1440×900, 1024×720 and enlargement, selection/caret/candidate/scroll retention and no unintended request |
| Inspector keyboard / Tab / Escape / continuity | PASS; existing module.workbook guard and lifecycle continuity assertions |
| Closed authority | PASS; closure closes Inspector under existing semantics, reopening in the same runtime retains observed entries; both entity descendants remain readable at both viewports and Restore remains disabled |
| Durable dismissed listing | N/A; no reload/account transition listing was introduced or claimed; History remains the durable distinction |

Before/after manual seeded review independently reproduced and inspected both
entities at 1440×900 and 1024×720. Bundle metadata, component byte lengths and
SHA-256 were verified before viewing images. The before review had serious Axe
color-contrast findings on all four text descendants; the after scans completed
without those findings. Remaining heading/table-header observations are outside
this scope. These are bounded diagnostics, not a whole-page accessibility or
user-effort study. Both exact sessions were stopped, terminal cleanup was complete,
foreground processes exited successfully, and task-owned scratch files were removed.
Private captures and expired links are intentionally absent from this handoff.

## Commands and retained results

Commands ran from `/home/jochi/code/cartulary` through public Make routing. Fresh
unit/browser/visual evidence used `CARTULARY_HARNESS_CACHE_MODE=off`. Help,
help-all, task-guide for `web.workbook`, `module.entities`, `web.design` and
`module.workbook`, and explain-target for ordinary/update visual scope were
revalidated before use.

Selections (existing titles/IDs unchanged):

```sh
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbook_relationship_chip_presentation_f96d227ab5
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.accessibility.verify_mention_chip_states_and_manual_resolution_e5964739d3,module.entities.accessibility.mention_arrow_navigation_reveals_focused_inspect_524ff6af49,module.entities.browser.the_browser_inspector_dismisses_a_mention_and_re_6150fd43cd
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.verify_keyboard_open_close_panel_navigation_esc_42b98cf08e
```

Every run root below has prefix
`/home/jochi/code/cartulary/.cartulary/test-results/`.

| Command/selection | Result | Exact run ID |
| --- | --- | --- |
| Planning chip unit / existing entity accessibility | PASS, baseline only | `20260930T141837Z-p52353` / `20260930T141924Z-p53123` |
| Required descendant regression before fix | Expected FAIL, both text ratios 3.90 | `20260930T143726Z-p21303` |
| Expanded chip unit | PASS | `20260930T144428Z-p60444` |
| Entity lifecycle / arrow guards | PASS; accessibility test expectation needed correction | `20260930T144448Z-p61448` |
| Complete actual-text entity accessibility | PASS, 58 required measurements | `20260930T145948Z-p79790` |
| Inspector keyboard guard | PASS | `20260930T145140Z-p10123` |
| Authored source formatting | PASS; diff inspected | `20260930T145537Z-p39488` |
| `make agent-finalize`, RESULTS_DIR unset | PASS | `20260930T145253Z-p75646` |
| Final `make format` / `make agent-finalize`, RESULTS_DIR unset | PASS; final source diff inspected | `20260930T151314Z-p55927` / `20260930T151422Z-p60778` |
| Final `make frontend-typecheck` | PASS | `20260930T151522Z-p65584` |
| Final `make frontend-import-boundary-check` | PASS | `20260930T151522Z-p65620` |
| Final `make lint-biome` | PASS | `20260930T151522Z-p65677` |
| `make json-shape-check` | PASS | `20260930T151522Z-p65318` |
| `make generated-artifact-policy-check` | PASS | `20260930T151522Z-p65357` |
| `make generate-drift` | PASS | `20260930T151522Z-p65327` |
| Final `make lint-markdown` | PASS | `20260930T151722Z-p72081` |
| Final `git diff --check` and scope review | PASS; six intended paths including this handoff, no generated/golden/lockfile changes | No harness run |
| Manual before / after seeded review | Closed, cleanup complete | `20260930T142636Z-p88048` / `20260930T145402Z-p80203` |
| Ordinary visual validation with correction | FAIL, expected mention difference plus unrelated comparisons | `20260930T145403Z-p80367` |
| Ordinary visual comparison with only this task's foreground change temporarily reverted | FAIL, same ten unrelated comparisons; mention fixture PASS | `20260930T150433Z-p17833` |

The red measurement JSON is under the first failing root's
`browser-e2e-a11y/browser-groups/a11y-workbook-a11y/contrast-checks/`; its filename
starts `dismissed-host-selected-1440x900-`. The accepted descendant matrix uses
the corresponding directory under `20260930T145948Z-p79790` and unique
`dismissed-*` filenames. Row results and Playwright reports distinguish selected
passes from target aggregation.

Earlier implementation iterations were not acceptance: formatting failed at
`20260930T144404Z-p55897` and `20260930T144601Z-p98065`, and type checking failed
at `20260930T144505Z-p87398` / `20260930T144601Z-p98030` on test typing/cleanup.
Those were corrected; type checking subsequently passed at
`20260930T145140Z-p10216`. Accessibility runs `20260930T145140Z-p10107` and
`20260930T145536Z-p39070` reached all open-layout measurements but failed on
test setup assumptions about closed Inspector presentation. Reopening the existing
Inspector without assuming expanded section navigation corrected the setup; no
product geometry or authority change was made.

Retained-run maintenance was skipped because `RESULTS_DIR` was unset: no eligible
successful full warm check was available. Full backend, release and unrelated
browser suites are unnecessary for this presentation change.

## Visual reconciliation and remaining work

The registered affected mapping is unchanged:
`visual.fixture.mention_chip_state_matrix` / `D-VFIX-004`, owner row
`module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`,
capture `visual.capture.bda571d19ea0163393fc`, golden
`apps/web/e2e/workbook.visual.spec.ts-snapshots/entity-mention-chip-states-linux.png`.
The current pinned renderer is Playwright 1.59.1 / Chromium 147.0.7727.15,
linux/amd64. The fixture keeps 1280×720, 100% zoom, dark_graphite, compact density,
and its existing masks, selector crop, focus and scroll anchor.

The ordinary run's `browser-e2e-visual/frontend-visual-reconciliation.json`
accounts for 255 capture intents / 255 active committed goldens, zero orphans,
missing goldens or ambiguous mappings, and all 29 registered fixtures. The
manifest/profile checks pass. Reconciliation's terminal status fails because
comparison rows fail, not because capture/golden mapping is incomplete.

The affected actual/diff images were reviewed: only dismissed label/marker paint
changes. All ten unrelated diff images were also inspected. They cover contextual
Task Request/Decision authoring and reference selectors, Inspector attached editing
and retained draft positioning, Evidence failed-upload text, and Decision
supersession's Open history text. None follows from the dismissed foreground
substitution. They prevent acceptance through the public update route, which
selects all 44 rows. Goldens, generated manifest, capture parameters, masks and
tolerances remain unchanged.

For causal isolation, only this task's one foreground line was temporarily
returned to the current HEAD style; no reset or preceding change was reverted.
The fresh ordinary visual run `20260930T150433Z-p17833` passed the mention row
and reproduced all ten unrelated diff PNGs with byte-identical SHA-256 to the
corrected run. The muted line was restored immediately afterward. This confirms
baseline failures independently of the substitution, rather than inferring them
from the changed-source run.

Unrelated affected golden names, all under the existing snapshot directory:

- `contextual-task-request-authoring-linux.png`,
  `contextual-task-request-authoring-narrow-linux.png`,
  `contextual-task-request-references-narrow-linux.png`,
  `contextual-decision-authoring-linux.png`,
  `contextual-decision-authoring-narrow-linux.png`,
  `contextual-decision-references-narrow-linux.png`:
  `module.workbook.visual.contextual_task_decision_creation`.
- `workbook-inspector-attached-edit-linux.png`,
  `workbook-inspector-retained-draft-linux.png`:
  `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`.
- `decision-supersession-review-linux.png`:
  `module.workbook.visual.decision_supersession_review_recovery`.
- `evidence-affordance-states-linux.png`:
  `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`.

The target and exact retained reconciliation artifact are
`make browser-e2e-visual` and
`/home/jochi/code/cartulary/.cartulary/test-results/20260930T145403Z-p80367/browser-e2e-visual/frontend-visual-reconciliation.json`.
Actual/diff PNGs and Playwright reports are under that target's
`browser-groups/visual-workbook-visual/` directory. Functional preparation reached
every capture; failures are screenshot comparisons. No baseline fix or general
UI audit was attempted.

The remaining dependency is resolution of those baseline visual comparisons in
their own scope. Then run ordinary reconciliation, the public
`make browser-e2e-visual-update`, review every changed image, and obtain two fresh
ordinary passes. This handoff does not claim those steps completed.

## Digest acceptance

Assessments apply to this bounded presentation correction, not the full product.

| Criterion | Result and rationale |
| --- | --- |
| A001 Authority | PASS; exact Core/design ownership and bounded contradiction correction recorded |
| A002 Scope | PASS; existing shared decision retained; tertiary binding and wrong contrast assertion retired; no new abstraction |
| A003 Repository state | PASS; clean baseline, owner/import/generated/routing inputs and consumers inspected; no new grid imports |
| A004 Tokens | PASS; existing muted token only; no palette literal or registry |
| A005 Theme | PASS; current dark_graphite renderer and diff only |
| A006 Density | N/A; no density implementation change; current compact fixtures retained |
| A007 Creation | N/A; product creation capabilities and payloads unchanged |
| A008 Responsive | PASS for selected scope; both requested viewports, enlargement and vertical reflow; no responsive routing change |
| A009 Overflow | PASS for selected scope; Inspector overflow checks and focus/scroll guards |
| A010 Inspector | PASS for selected scope; inspection, closed disabled Restore and semantic subject remain intact; dispatch unchanged |
| A011 Continuity | PASS; existing focus/draft/caret/scroll and lifecycle continuity guards |
| A012 Transactions | N/A to changed behavior; captured requests/duplicate protection inspected and unchanged |
| A013 Acknowledgement/recovery | N/A to changed behavior; replay, acknowledged refresh and retained ownership inspected and unchanged |
| A014 Editing | PASS for preserved focus/draft/Tab/Escape scope; no editor lifetime change |
| A015 Conflict | N/A; no conflict implementation or presentation change |
| A016 Query/interaction states | PASS for selected closed/readable/blocked-action scope; query producer unchanged |
| A017 Authorization scope | PASS for same-runtime closure; account/revocation lifetimes unchanged and outside this change |
| A018 Evidence | N/A; unrelated Evidence visual mismatch is a blocker, not an accepted change |
| A019 Accessibility | PASS for selected text, names, cues, focus and keyboard checks; no whole-page claim |
| A020 Components | PASS; four states preserved, dismissed entity/selection variants, long source, zoom/reflow/spacing |
| A021 Virtualization | N/A to changed implementation; existing semantic focus guard retained and passed |
| A022 Visual fixtures | BLOCKED; unrelated baseline comparisons prevent public update acceptance and two ordinary passes |
| A023 Selectors | PASS; existing semantic record/field/mention/chip selectors and scoped descendants |
| A024 Test authority | PASS; no executable Markdown dependency; owner correction is human-reviewed |
| A025 Generated artifacts | PASS; no generated file hand-edited or projection needed; policy, JSON shape and drift checks passed |
| A026 Compatibility | PASS; no schema/API/data migration or new persistence; existing consumers retained |
| A027 Handoff | PASS as an honest blocked handoff; completion withheld for A022 |

Rollback is a coordinated revert of this task's design, source, test and handoff
changes, plus any future accepted golden/generated-manifest changes. Preserve
the preceding focus and collection corrections. No service or data rollback is
needed.
