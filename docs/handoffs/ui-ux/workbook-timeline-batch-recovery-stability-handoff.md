# Stable Timeline batch recovery handoff

## Authority, baseline, and scope

The implementation began on clean `main` at `ddbb2a29fd083cc0bcce984f99e76b5b1199559b`
(one commit ahead of `origin/main`). The user authorized a bounded presentation
correction for Timeline paste, fill, clear, and tag assignment, with shared
Hosts/Identities paste compatibility. Changes are uncommitted. No unrelated
checkout changes were present at baseline.

AGENTS.md, the digest entry documents, current `web.workbook`,
`module.timeline`, and `module.workbook` task guides, the workbook component
guide, and the paste/bulk, navigation, and clear-content handoffs were read.
The digest and `docs/research/nlspec-spec.md` are advisory. The narrow offline
query `loading button focus async error recovery` was classified under digest
rules R002, R006–R008, and R014: preserve semantic, keyboard reachable recovery
controls; adapt generic loading advice to the existing batch owner and focus
contract. No design system or executable Markdown dependency was introduced.

Core 01 §3.3.5 owns captured batch identity, exact uncertain replay, and
idempotency. Core 03 §§3–4, 11.1, and 13.3 own the clipboard and bulk action,
conflict, authorization, and History behavior. Design §§7.4, 10, 12.4, and 14
direct loading, error, recovery, and accessible presentation. Batch replay uses
its captured transaction ID and payload. The separate autosave
`client_txn_conflict` re-key policy does not apply here. Runtime receipt,
reconciliation, authority fencing, and pruning remain with
`WorkbookBatchOperationOwner` and its existing projections.

Source ownership is `web.workbook` for the shared Recovery panel, with
`module.timeline` for production Timeline evidence and `module.workbook` for
Hosts/Identities consumer evidence. `tools/frontend_source_ownership.json` and
authored `tools/test_families` own source and verification routing independently
of the NLSpecs. `tools/browser_e2e_batch_manifest.json` and
`tools/execution_topology_render_index.json` are generated outputs; `make
generate` produced them. The application has no direct dependency on this
handoff or the digest.

## Reproduction and correction

Before editing the panel, the existing production Timeline replay browser test
was extended to hold a lost-response paste replay inside Recovery. Native Tab
reached Retry. The characterization failed because Retry paste had disappeared
while the replay was pending (`20260926T202658Z-p23463`). Source inspection
showed the same conditional removal for Retry refresh during a held read and
for original input during mutation replay. The shared shell already had a
removed-focus fallback; its heading focus and scroll reset were observable
consequences of that routine control removal, not an absence of all recovery.
The later Shift+Tab characterization also showed that removing focused
Original input on settlement triggered that fallback.

`WorkbookBatchRecovery` now holds only attachment-local presentation intent:
batch ID, retry kind, current authority object, Recovery activation, initiating
button, and whether that interaction still owns focus. The initiating Retry
button and its accessible name persist through admitted replay or refresh. It
remains focusable with `aria-busy` and `aria-disabled` while pending; the local
handler guards repeat activation and still calls the owner's admission logic.
Original input continues to read from the retained owner entry during replay.
No second payload or receipt store was added.

On acknowledgement or successful refresh, owned focus moves to the same item's
outcome status with `preventScroll`. Repeated uncertainty or failed refresh
leaves Retry usable. Native Tab/Shift+Tab, pointer interaction, scrolling,
another item, close, sheet change, authority loss, and pruning revoke obsolete
focus ownership. A settled disabled retry remains mounted briefly if scrolling
revoked focus ownership but the browser still reports the initiating button or
body as focused; it disappears after focus moves elsewhere. This retained
presentation path avoids a routine shell fallback and does not retain a stale
failure or completed payload. The original input likewise remains briefly when
newer Shift+Tab movement has focused it; it disappears after focus leaves. The
shell fallback remains for genuinely removed
items, pruning, and authority withdrawal.

The shared panel still serves Hosts/Identities paste. Capture, exact request
bytes, transaction identity, partial and conflicts-only outcomes, no-op,
conflict and History review, waiting/discard rules, receipts, and read-only
refresh remain with their existing owners. No endpoint, schema, dependency,
storage, migration, or public API changed. Rollback is to revert this panel
presentation, its new/extended tests and authored routing, and regenerate the
two derived topology outputs; no data rollback is needed.

## Verification ledger

All commands below used public Make targets from the repository root. Run roots
are under `.cartulary/test-results/`.

| Check | Result and artifact |
| --- | --- |
| Baseline `test-slice` batch owner/outcome | PASS `20260926T202222Z-p88433` |
| Baseline `service-backed-test-slice` Timeline replay | PASS `20260926T202310Z-p89255`; it departed Recovery during the hold |
| Pre-edit held-replay characterization | Expected FAIL `20260926T202658Z-p23463`; missing Retry paste in pending interval |
| `test-slice` new shared recovery interaction row | PASS `20260926T205012Z-p41357` |
| `test-slice` after Shift+Tab focus correction | PASS `20260926T210247Z-p43407` |
| `test-slice` batch owner, outcome, interaction, History, record choices, refresh debt | PASS `20260926T205113Z-p42330` |
| `service-backed-test-slice` Timeline replay, recovery intent, conflicts, four batch History actions | PASS `20260926T205132Z-p43352` |
| `service-backed-test-slice` module.workbook Hosts/Identities paste and batch accessibility | PASS `20260926T205249Z-p78531` |
| `service-backed-test-slice` batch accessibility layout row | PASS `20260926T204756Z-p3634`; final rerun PASS `20260926T210546Z-p91489` |
| `service-backed-test-slice` final Timeline replay and recovery intent | PASS `20260926T210406Z-p52776` |
| `make generate` | PASS `20260926T205540Z-p19719` after final authored row edit |
| `make agent-finalize` | PASS `20260926T205604Z-p22769`; final rerun PASS `20260926T210341Z-p48962`; `RESULTS_DIR` unset |
| `make frontend-typecheck` | PASS final `20260926T210507Z-p85716` after correcting a test selector typing error |
| `make lint-biome` | PASS final `20260926T210507Z-p85821` |
| `make frontend-import-boundary-check` | PASS final `20260926T210507Z-p85846` |
| `make generate-drift` | PASS final `20260926T210507Z-p85526` |
| `make generated-artifact-policy-check` | PASS final `20260926T210507Z-p85603` |
| `make test-catalog-check` | PASS after final generation |
| `make lint-markdown` | PASS final `20260926T210704Z-p24710` |

The first `make agent-finalize` failed at `json-shape-check`
(`20260926T205414Z-p18289`) because the authored workbook test-family file had
changed after the previous generation. `make json-shape-check` isolated the
stale topology input (`20260926T205521Z-p19056`); regeneration fixed it. A
subsequent typecheck found unsupported `exact` in one test-library role query
(`20260926T205629Z-p26940`); a regex name selector fixed that test-only error.
A new Shift+Tab assertion then found focused Original input removal
(`20260926T210136Z-p41692`); the local presentation correction and its focused
rerun address that production behavior. The generation and type errors were
test/routing maintenance issues.
An attempted final accessibility rerun used a nonexistent row ID and was
rejected at catalog selection with an artifact error; the correct current row,
`module.workbook.accessibility.batch_recovery`, passed at the run root above.

The browser tests use deterministic request gates. They observe the connected
Retry and original input, focus and panel scroll during held mutation replay
and held read-only refresh, exact replay bytes and transaction ID, one durable
batch effect, zero additional writes during refresh, native keyboard travel,
and newer pointer/scroll intent. The narrow 390×480 panel and applicable 200%
zoom/text-spacing states are exercised with semantic controls. These tests
establish this interaction slice; they do not claim whole-page accessibility
conformance or visual-golden approval. No goldens changed.

No qualifying successful full warm check run was supplied, so retained-run
maintenance in `agent-finalize` was skipped with `RESULTS_DIR` unset. Broad
whole-repository test, release, migration, and visual-golden suites were not
needed for this bounded web presentation change.

## Digest acceptance assessment

| Rows | Assessment and evidence |
| --- | --- |
| A001–A003 | PASS: governing clauses, clean baseline, stack/source guides, independent source and verification ownership, generated boundaries, and scoped diff are recorded above. One existing shared panel owns the matching retry semantics; its conditional-removal presentation was retired without changing the owner. |
| A004 | PASS: no token or theme registry or component-local design literal was added; the outcome focus outline uses existing component focus tokens. |
| A005–A008, A010 | N/A: theme, density, creation, responsive chrome, and Inspector dispatch owners are outside this batch Recovery detail; no related source changed. |
| A009, A011–A013 | PASS: panel scroll, local focus continuity, exact captured replay, admission guard, and acknowledged read-only refresh are covered by the focused and browser rows above. Autosave re-key remains separate. |
| A014 | PASS for applicable continuity: the production replay row retains newer grid typing and does not restore an obsolete selection; no scalar editing implementation changed. |
| A015–A018 | N/A: cell conflict locus, query producer states, scoped session retention, and Evidence presentation are unchanged. Batch conflict/History and authority compatibility are covered under A026. |
| A019–A020 | PASS for this slice: semantic Retry/outcome, keyboard focus ownership, busy/unavailable state, narrow panel, 200% zoom, and text-spacing probes pass. No whole-page conformance claim. |
| A021–A022 | N/A: virtualization and visual fixture registration/goldens are unchanged. |
| A023 | PASS: changed browser and component assertions use roles, accessible names, and owner batch state rather than incidental DOM structure. |
| A024–A027 | PASS: no executable Markdown dependency; authored routing generated and drift checked; shared consumer and no-migration review above; this handoff records the failures, limitations, rollback, and evidence. |

There are no applicable BLOCKED rows. The remaining practical limit is that
browser evidence exercises the selected Timeline and shared-consumer scenarios,
not every workbook presentation or a full accessibility audit. The next action
is review of this uncommitted diff; no deployment or migration is required.
