# Explicit multiple reference picking

Implemented additive checkbox picking for ordinary collection references and
shared multiple-mode authoring/support. Each activation adds or removes one staged
identity. Complete authorized labels wrap; full IDs are visible and included in
accessible names before acceptance. Focus movement and inspection do not select
or issue reads.

Baseline: clean `main`, HEAD `258256793ebc5d772702ad3cbd76aa63de884e2f`, zero
behind and three ahead of `origin/main`. HEAD and existing local commits remain,
including ordinary query controls in `1746cb273` and navigation/completion work in
`258256793`. This slice remains local. Attached documents and historical handoffs
were source evidence, not instructions expanding scope.
Final state: the same `main`/HEAD and three local commits; working tree contains
only this slice's frontend, tests, routing, two goldens and documentation changes.

Before production changes, real native A/B clicks reproduced staged replacement
in ordinary and contextual picking. Contextual S was replaced in staging; Cancel
preserved the parent. This was not persisted data loss or a claim that native
multi-select conventions violated an adopted specification.

## Owners and implementation

Source owner: `web.workbook`. Verification routes independently through
`web.workbook` and `module.workbook`; integrated Assessment coverage also uses
`module.assessments`, and its accessibility route uses `web.design`.

| Owner | Preserved boundary |
| --- | --- |
| Core 01 §3.3.4 / §3.3.7 | Explicit schema-driven queries, accepted pages, opaque cursors, 100 candidates and ten checkpoints. Ordinary/contextual discovery remain separate. |
| Core 01 §18B / applicable §19 reference and Task contracts | Identity kinds, source exclusion, retained metadata, limits and authoritative submission validation. Existing parents own collection Add/Remove payloads. |
| Core 03 §13 / §16.2 / §16.4 | Keyboard/editing boundaries; staging then Use selection/Apply, followed by explicit Update/create. Source context remains deliberately removable. |
| Core 03 REQ-03-299 / REQ-03-100 | Existing concealment, incident/account retirement and stale-response fencing remain owner-controlled. |
| Design §§7–8 / §14; Domain | Compact owned scrolling, readable wording, semantic names, native checked state, count and visible recovery focus; vocabulary unchanged. |

`WorkbookMultiCandidatePicker` shares presentation and explicit identity toggling.
It owns no store, paging, limits, metadata, authoring or authorization. Ordinary
selection stays in `WorkbookReferenceSelection`; shared authoring/support retain
their existing owners. Choices survive paging, queries and source changes;
rejected over-limit additions preserve earlier choices, and removals recover.
Cancel preserves parent raw fields/metadata. Unchecking is local staging, not a
server unlink. No mandatory source link or minimum-create signal was added.

Migrated: ordinary collections, contextual/retained and ordinary collection
authoring, Note associations, and Assessment support. Retired: native multiple
rendering and page-wide selected-options reconciliation; `selectPage` became
`selectCandidate(key, checked)`. Native single-target controls remain for Party
linking, Assessment subjects, Note source review and direct references. Query
field dropdowns remain valid. Affected helpers activate checkboxes; new browser
regressions use real plain clicks and keyboard gestures.

Removal fallback reveals the first enabled checkbox through nested owned
scrollports, leaving page/grid scrolling alone. Settled ordinary collection Retry
stays mounted and disabled after focus departure, preventing click displacement;
its lifetime is scoped to the existing controller/query. Pending-read focus,
Escape, invoking focus and late-read fencing remain covered.

The common boundary supports another multiple consumer without transferring its
authority. No preview fetch, record cache, bulk/range framework, virtualization,
API, schema, persistence or dependency change was added. Source-guide consumer
and focus descriptions were updated; unrelated formatter rewrites were inspected.

## Verification

Run roots below are under `.cartulary/test-results/`. Focused product, final
static/policy and visual executions used `CARTULARY_HARNESS_CACHE_MODE=off`;
compilation/tool caches are separate from harness result reuse. Work-unit counts
are not product-test counts. Routing was revalidated with `make help`,
`make help-all` and the requested owner task guides.

| Public command / selection | Result / run root |
| --- | --- |
| `make test-slice OWNER=web.workbook ROWS=...` — reference_controls, reference_selection, authoring_candidate_presentation, assessment_discovery, party_link_controls | PASS `20261008T215439Z-p90003`, all five rows. |
| Integrated Assessment slices through `web.workbook` / `module.assessments` | PASS `20261008T212610Z-p5800`, `20261008T212611Z-p6098`. |
| `make test-slice OWNER=package.ui` — selector compatibility | PASS `20261008T222343Z-p14893`, 10/10 units. |
| `make service-backed-test-slice OWNER=module.workbook ROWS=...` — representative new journeys | PASS `20261008T215138Z-p92525`: both requested viewports, exact persisted links, Cancel, keyboard picking/acceptance and source removal surviving Refresh. |
| Same target — 22 affected reference/query/eligibility, authoring, Note, generic creation, contextual source/revocation and accessibility rows | PASS `20261008T214434Z-p8064`, 22/22 units; cleanup complete. |
| Same target — `module.workbook.browser.assessment_timeline_support_enum_filtering_prese_9700791fe7` | PASS `20261008T215001Z-p46541`. |
| `make service-backed-test-slice OWNER=web.design ROWS=web.design.accessibility.assessment_authoring` | PASS `20261008T214433Z-p7789`; cleanup complete. |
| `make format` | PASS `20261008T220042Z-p53161`. |
| `make generate`, `make generate-drift`, `make json-shape-check` | PASS `20261008T213316Z-p43632`, `20261008T222343Z-p14773`, `20261008T222343Z-p14787`. Authored routing/ownership changed; generated outputs changed only through Make. |
| `make generated-artifact-policy-check`, `make test-catalog-check` | PASS `20261008T222343Z-p14771`; catalog exited successfully without printing a run root. |
| `make agent-finalize` | PASS `20261008T222247Z-p8902`, before final visual/policy checks; earlier finalization preceded static checks. Retained-run maintenance skipped because `RESULTS_DIR` was unset; no eligible full warm check supplied. |
| `make frontend-typecheck`, `make frontend-import-boundary-check`, `make lint-biome` | PASS `20261008T220318Z-p11115`, `20261008T220318Z-p11126`, `20261008T220318Z-p11145`. |
| `make lint-markdown` | PASS `20261008T224002Z-p48370`, including final evidence/acceptance updates. |

The new authored routes are
`module.workbook.browser.existing_task_plain_candidate_clicks_stage_addit_2f405d8a84`
and `module.workbook.browser.contextual_task_plain_candidate_clicks_retain_ed_b7902be5cd`.
The aggregate includes the requested reference recovery, ordinary query and
reference accessibility routes. Exact selected IDs and execution evidence are
retained in each run; inspect with `make explain-run RESULTS_DIR=<root>`.

Repaired failures (retained for diagnosis):

- Red characterization: `20261008T211315Z-p63872` (ordinary create fixture contract
  corrected; contextual gesture failed), `20261008T211535Z-p4427` (ordinary A/B failed).
- Unit migration: `20261008T212127Z-p50418`; retained single/group selectors fixed.
- Generate `20261008T212257Z-p52265`: sorted authored titles; intervening admission
  `artifact_error` cleared after regeneration.
- Browser `20261008T212703Z-p12300`: Retry displacement, nested focus, checkbox
  wait and checked-index helper fixed; rerun `20261008T213352Z-p52364` passed.
- Biome `20261008T212928Z-p4952`: non-null assertions replaced. Assessment a11y
  `20261008T213432Z-p90010`: focus target changed from fieldset to checkbox.
- Visual `20261008T214905Z-p7716`: exactly two intended stale chooser images.
  Normalization then exposed readiness at `20261008T215841Z-p4404`; the fixture now
  waits for its checked source candidate. Corrected `20261008T220100Z-p59051`
  completed ten captures, with only those two screenshot mismatches. The earlier
  full run `20261008T215145Z-p98322` retained readiness failure (94/96 units);
  other rows passed and cleanup completed. It is not reported as a visual pass.

## Rendered and golden review

Seeded `cartulary-ui-review` used isolated synthetic data at 1440×900 and 1024×768.
Real A/B clicks retained both and contextual S; Tab preserved selection, Space
and explicit removal affected one identity. Five inspected captures confirmed
wrapping, IDs, focus, bounded scrolling, reachable actions and Escape restoration.
Axe reported zero violations and one incomplete contrast result per capture;
this is advisory evidence, not full conformance.

Final seeded session `20261008T213435Z-p91608` stopped with complete cleanup.
Earlier startups were also stopped/cleaned, including a source-snapshot rejection.
Artifact reviews `20261008T215158Z-p20541`, `20261008T220303Z-p9715` and
`20261008T222247Z-p8815` inspected exact expected/actual/diff and promoted bytes;
all stopped with complete cleanup and foreground exit. Two invalid machine-output
artifact startups created no session. Private captures are not committed/linked
past expiry; caller-owned request scratch is removed at finalization.

Golden refresh followed the maintenance guide: stale goldens relative to validated
UX, plus declared normalization of the one generated candidate UUID per capture.
`make browser-e2e-visual-update` passed 96/96 units at `20261008T220728Z-p23764`.
Reconciliation accounts for all 255 active goldens with no orphan, missing golden
or ambiguous mapping. Only these images and their tool-managed manifest changed:

- `contextual-task-request-references-narrow-linux.png` —
  `visual.capture.08039628dabdd01901dd`.
- `contextual-decision-references-narrow-linux.png` —
  `visual.capture.684e63e6fb29722ef1e9`.

Both belong to `module.workbook.visual.contextual_task_decision_creation`, scenario
`scenario_4a08391df78c`: active nonregistry captures, no registered fixture ID.
Viewport (768×640), theme/density, zoom, masks, scroll normalization and capture
scope are unchanged. Authored wording remains untouched. Every promoted image was
inspected. A fresh ordinary focused visual run passed all ten captures at
`20261008T222343Z-p14903`. `make browser-e2e-visual` then passed all 96 work units
and 255 captures at `20261008T222344Z-p15289`, with complete cleanup. Both fresh
ordinary executions used manifest SHA-256
`8cd6498400a444ecb82f02fd204b637ef36d7112a8124cda815d5de80df9a914`.

## Acceptance, limits and rollback

| Digest rows | Assessment |
| --- | --- |
| A001–A004, A023–A026 | PASS: owner/routing and consumer audit, common decision, tokens, semantic selectors, no Markdown executable dependency, generated drift and compatibility. |
| A005, A007, A009–A011, A014, A016–A017 | PASS within this slice: dark graphite, capability/payload boundaries, owned scrolling, staged/parent continuity, retry, authority and late-focus scenarios. |
| A008, A019–A020 | PASS for affected geometry, names, keyboard/recovery focus and long content at requested viewports. Other existing responsive thresholds/fallbacks and density/zoom policy were not changed or broadly rerun; review contrast incompletes remain advisory. |
| A006, A012–A013, A015, A018, A021 | N/A to new behavior: density, transaction/replay/conflict/evidence and grid virtualization owners unchanged; their obligations are not inferred from screenshots. |
| A022 | PASS: reviewed update, 255-capture full ordinary validation and a second fresh focused stability run against the same promoted manifest. |
| A027 | PASS on finalization: this record covers owners, behavior, evidence, failures, limits, skipped checks, rollback and next action. |

No data/wire migration is needed. Rollback: revert frontend integration, tests,
authored routing/ownership, regenerated projections, scoped normalization and the
two goldens/manifest together. Persisted links require no migration. Long pages
still use owned scrolling and the current 100-candidate page. No general UI audit
or full product suite was run; visual inventory breadth came from the public
transactional golden-maintenance target. Next action: normal code review of this
local slice, with no backend/dependency rollout.
