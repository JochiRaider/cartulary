# Workbook clipboard fidelity refactor handoff

## Execution control

This is the sole WCF execution tracker. The user authorized the complete plan,
including the owner clarifications below. Each workstream depends on its
predecessor being DONE. An unresolved owner contradiction blocks dependents.

| Workstream | Status | Exit |
| --- | --- | --- |
| WCF-01 Characterization and representation contract | DONE | Owner decisions and matrix complete; gaps reproduced. |
| WCF-02 Copy and decoding boundary | DONE | Supported representations and cross-language codec fixtures agree. |
| WCF-03 Admission and retained operations | DONE | Destination agreement, rejection without writes, exact recovery. |
| WCF-04 Integrated fidelity evidence | DONE | Applicable browser, service, codec, and lifecycle scenarios pass. |
| WCF-05 Final validation and handoff | DONE | Final checks and acceptance assessment complete. |

## Baseline and authority

- Execution baseline: `main`, HEAD `2e1f35dc44497be13179f3fc24753713a1b35cee`;
  clean index and working tree, no pre-existing changes. Revalidated before edits.
- Core 01 §§3.3.5, 7.4, 17.2 and 18 own routes, identity, field contracts,
  exact headers, and the parser-neutral mapping kernel. Core 02 §§5–6 and 15
  own source values, entity identity, and history. Core 03 §§3–4, 11 and 13
  own paste dispatch, recovery, and spreadsheet interaction. Core 04 owns
  authorization, resource bounds, and formula-safe export (REQ-04-053).
- `docs/domain.md` and `docs/design.md` retain their vocabulary/design boundaries.
  The localized digest was read in its prescribed order during planning and is
  advisory. Historical handoffs are context, not fresh verification.
- Source ownership and verification routing remain separate. Runtime and tests
  do not consume Markdown. No advisory digest edits, new dependencies, persistent
  browser storage, new paste surfaces, analyst-data edits, or deployment work.

## Representation decisions

The user explicitly selected strict external HTML tables; multiline plain text
as TSV; data-only intent for Cartulary copies; strict rectangular parsing; an
8 MiB UTF-8 bound; and a fixed TSV interpretation for omitted/auto API format.
These decisions amend the adopted owners before dependent implementation.

WCF-01 through WCF-05 are complete. Dedicated production-browser fidelity
evidence is WCF-04; desktop Excel interoperability is not claimed.

## Evidence log

### Planning baseline (not implementation acceptance)

- `make help`, `make help-all`, and task guides for Grid Adapter, Workbook,
  Timeline, Entities, Tabular Ingest, and Imports completed.
- `make test-slice OWNER=module.tabularingest`: PASS, one routed baseline test;
  `.cartulary/test-results/20260915T191700Z-p25555/run-summary.json`.
- `git diff --check`: PASS before implementation.

## Next action

Review the implemented diff and coordinate the documented API migration before
any separately authorized rollout. No implementation work remains.

## WCF-01 characterization

### End-to-end owner trace

Grid columns supply `getClipboardValue` (Timeline source strings; other surfaces
use scalar values or approved collection labels). `planSemanticCopy` serializes
the visible range; `SemanticDataGrid` writes the native clipboard. Its native and
RDG wrapper paste paths converge on `useGridPasteController`. Workbook decoding
feeds semantic dimensions and stable targets, then Timeline or Entity planning.
`WorkbookClipboardPastePort` admits an immutable plan to
`WorkbookBatchOperationOwner`; the transport captures its exact path/body/identity.
Workbook route contributions decode source-owned requests. Tabular Ingest parses
and adapts cells to `cartulary.tabular_mapping_kernel.v1`. Timeline applies
`owner_batch_apply_v1`; Entities applies entity-origin reuse/create in one
transaction. Source owners retain authorization, history, conflicts and receipts.

### Interaction matrix

| Context/input | Disposition |
| --- | --- |
| Active text editor, any source | Native caret/selection copy and paste; no table planning. |
| Grid 1×1 Cartulary copy | Marked scalar HTML preserves exact presentation; formula-safe plain fallback. |
| Grid N×1, 1×N or rectangle copy | Marked data-only HTML plus escaped TSV fallback; preserve every cell. |
| Plain single line, including commas/quotes | Scalar; empty text is a no-op. |
| Plain tabs/newlines/CR | Strict TSV; commas literal; ambiguous scalar line breaks cannot be recovered. |
| Explicit CSV/TSV | Strict format-selected table, including 1×1 and explicit empty cells. |
| External HTML table | One strict rectangle; 1×1 scalar; unsupported table fails without fallback. |
| HTML without table | Continue to offered TSV, CSV, then plain text. |
| Timeline existing records | Stable visible targets, current versions, writable field contracts. |
| Timeline authorized create targets | Current create/group/window policy; no inferred unloaded targets. |
| Hosts/Identities existing scalar | Ordinary source-owned patch and clear/required validation. |
| Hosts/Identities table | Existing entity-origin reuse/create semantics, no new record-target batch route. |
| Generic/Assessment/Network Analysis copy | Same representation export, no additional paste capability. |
| Evidence image/file clipboard | Existing Evidence owner routing; no grid decoding of files. |

### Confirmed gaps and remediation records

Each row includes its affected implementation/tests, rationale and binary exit.
Core 01/03 amendments govern representation/parsing; Core 04 governs byte bounds
and formula neutralization. This handoff records documentation and compatibility.

| Gap | Remediation and affected paths | Benefit and rationale | Compatibility/risk | Binary validation |
| --- | --- | --- | --- | --- |
| Comma-containing copied N×1 becomes N×2 | Replace Workbook punctuation guessing with explicit Grid Adapter representations and fixed TSV fallback; codec/browser tests. | Values and geometry share one interpretation. | Old auto-CSV clients must specify CSV; plain-only ambiguity remains explicit. | Copy/paste keeps N×1 and exact values. |
| Copied quote scalar pastes serialization quotes | Marked 1×1 scalar HTML and raw scalar fallback; Grid Adapter codec and production copy tests. | Separates serialization from analyst text. | Unsupported/stripped metadata cannot reconstruct every scalar. | Quotes survive exactly, without added quoting. |
| Empty records dropped; ragged browser padding differs from server | Strict lexical decoding, explicit empties, clipboard rectangularity before mapping; shared TS/Go corpus and Imports tests. | Prevents row shifts and fabricated clears. | Legacy malformed/ragged clipboard inputs now reject; Imports keeps owned geometry. | Browser/server values, dimensions and failure class agree. |
| Browser accepts malformed quotes rejected by Go | Whole-field quote grammar in both adapters. | No misleading planned selection before server rejection. | Incidental permissive behavior intentionally retired. | Malformed fixture yields no applicable plan or mutation. |
| Header-shaped copied data treated as header | Data-only marker plus captured `header_mode=none`; Timeline admission/hash tests. | Recognition cannot silently delete analyst data. | Additive request member; old default hashes retained. | Exact header-looking data persists as data. |
| No aggregate byte bound; late geometry limits | Authored 8 MiB bound, incremental token/DOM admission and source boundary checks. | Bounds giant cells and HTML before unnecessary allocation. | Previously unbounded payloads reject; no hidden browser-only cap. | Boundary accepted, boundary+1 rejected before write. |
| Formula-leading clipboard exports are unsafe | Formula-neutralized portable text and reversible marked HTML escape. | Satisfies Core 04 without altering Cartulary round-trip values. | Plain-only imports retain apostrophes literally. | Formula-looking values remain strings; no executable export. |
| Grid publishes destination before source admission | Return explicit admission outcome and preserve prior range on rejection. | Feedback matches actual operation. | Retained batch lifetime and event identity unchanged. | Rejection leaves prior focus/range and sends no request. |

### Executable evidence

- TS characterization in `workbookClipboard.test.ts`: emitted `a,b\nc,d`
  becomes two columns; `"a""b"` stays serialization text in scalar dispatch;
  blank records disappear and short rows are padded. The existing malformed
  quote case demonstrates frontend acceptance.
- Go characterization in `tabularingest_test.go`: blank rows disappear, short
  rows remain short, malformed quotes fail. This is code execution, not browser
  execution. Copied bytes are separately checked in Grid Adapter's core test.
- First frontend slice failed due to a characterization import of a non-exported
  helper (`formatGridClipboardTSV`), not product behavior; corrected to use the
  emitted bytes with separate package-local serialization assertions. Failure:
  `.cartulary/test-results/20260915T194534Z-p33082/run-summary.json`.
- Corrected two-row frontend slice PASS:
  `.cartulary/test-results/20260915T194609Z-p34320/run-summary.json`.
- Tabular Ingest characterization PASS:
  `.cartulary/test-results/20260915T194535Z-p33330/run-summary.json`.

Owner normalization is distinct from corruption: Core 01 string contracts may
NFC-normalize, canonicalize CRLF/CR, trim outer whitespace, reject controls or
clear normalized-empty optional fields. Codec fixtures preserve raw input;
authoritative tests compare against the addressed field's contract.

WCF-01 exit: PASS. Copy serialization characterization passed through
`make test-slice OWNER=package.grid_adapter
ROWS=package.grid_adapter.frontend_unit.translate_vendor_row_column_coordinates_to_recor_652aef8a27`;
run `.cartulary/test-results/20260915T194727Z-p35817/run-summary.json`.
The sorted-target baseline also passed at
`.cartulary/test-results/20260915T194622Z-p35031/run-summary.json`.
No owner contradiction remains. Browser reproduction is explicitly deferred to
WCF-04 production-path evidence; all reproductions above are unit executions.

## WCF-02 result

- Added `packages/grid-adapter/src/clipboardCodec.ts` and its focused test,
  plus `internal/modules/tabularingest/clipboard_codec.go`. The authored
  `contracts/tabularingest/clipboard.v1.json` supplies cross-language examples
  and limit facts; tests check both implementations against it.
- Grid copy now offers marked HTML and safe plain text through
  `semanticClipboardPolicy.ts`, `SemanticDataGrid.tsx`, and the test binding.
  HTML uses inert template extraction; metadata contains only version and intent.
- Go `ParseTable` validates strict clipboard geometry; the Imports entry point
  shares lexical decoding while retaining its own rectangle and source limits.
  Removed trimming of record terminators and exact-header row padding.
- Mapping-kernel canonical vectors and fingerprints still pass unchanged.
- `make test-slice OWNER=module.tabularingest`: PASS;
  `.cartulary/test-results/20260915T195634Z-p44527/run-summary.json`.
- `make test-slice OWNER=package.grid_adapter
  ROWS=package.grid_adapter.regression.clipboard_representation`: PASS;
  `.cartulary/test-results/20260915T195634Z-p44530/run-summary.json`.
- Intermediate failures were related: new routing needed ASCII order and matching
  family ID; test corpus lookup needed the runner's repository-relative convention;
  a legacy empty-payload error assertion needed the new deterministic token.
  Failed executable roots: `20260915T195525Z-p43212` and
  `20260915T195525Z-p43219`. Those failures are superseded by the passes above.
- Exit PASS: 17 shared lexical cases, marked scalar/range round-trips, unsafe
  HTML rejection, fallback behavior, formula/apostrophe escaping, and limits.
  Event admission/server route integration remains WCF-03; browser evidence
  remains WCF-04. No desktop Excel interoperability claim is made.

## WCF-03 result

- Grid Adapter snapshots offered MIME types, decodes before target planning, and
  returns source admission to range publication. Decode failures announce locally;
  rejected operations retain the prior selection. Native editor routes stay native.
- Timeline exact headers remain schema-derived; marked copies capture `none`.
  Headers cannot target hidden, noneditable or unavailable columns. Entity scalar
  planning now checks current versions and exact shape; tables retain entity-origin
  reuse/create. No new record-update batch semantics were introduced.
- Captured requests now include explicit format/header interpretation. Authored
  Workbook OpenAPI and unreleased 2.0 changeset preceded Make-owned generation.
  Omitted/default header mode preserves historical hashes; nondefault mode is
  hashed. Committed receipts are retrieved before new lexical parsing, after
  current route authorization and request admission. Invalid mapped writable
  values reject rather than disappearing into unmapped data.
- Retired `formatGridClipboardTSV` and Workbook delimiter guessing. The shared
  codec owns dimensions; source adapters own field mapping. Imports still owns
  discovery geometry. Copy bounds account for escaping before output allocation.
- `make generate`: PASS, `20260915T200206Z-p48541`; first run
  `20260915T200026Z-p46343` correctly required the additive changeset entries.
- `make frontend-typecheck`: PASS, `20260915T201424Z-p32280`.
- Grid codec, both semantic bindings, sorted targets: PASS, narrow Grid Adapter
  slice `20260915T201425Z-p32676` (4/4 graph units).
- Timeline plan/header admission: PASS, `20260915T201300Z-p25451`.
  Timeline Go parsing/hash/provenance rows passed in `20260915T201026Z-p98641`;
  that run's frontend fixture import failure was corrected by the later pass.
- Entity plan admission: PASS, `20260915T201302Z-p25882`;
  Go entity admission/hash row passed in `20260915T201113Z-p1547`.
- `make service-backed-test-slice OWNER=module.workbook
  ROWS=module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770`:
  PASS, `20260915T201115Z-p1774` (3/3 graph units). Includes authoritative
  Timeline updates/creates/conflicts, entity reuse/create/replay and no-effect
  rejection/authority checks.
- Related intermediate failures: legacy padded header fixture, event mock without
  MIME types, scalar fixture with two target columns, and omitted new request
  member expectations. Corrected fixtures reflect the adopted owner contract.
  Roots: `20260915T200505Z-p52872`, `20260915T200506Z-p53091`,
  `20260915T201113Z-p1547`. Full Grid Adapter run passed its existing browser
  accessibility evidence but is not counted as dedicated clipboard fidelity.
- Formatting/lint iterations found an unnecessary stable setter dependency and
  confusing void union; both corrected. Final formatting/lint remains WCF-05.
  A diagnostic-only Make override was rejected by the harness input allowlist;
  no task/harness policy was changed.
- Compatibility: CSV API callers must explicitly send `csv`. During rollout,
  drain or reconcile legacy uncommitted `auto` clipboard requests before upgrade;
  do not retry them under changed semantics. Committed matching receipts remain
  retrievable. Browser retries use captured bytes and transaction identities.
- Exit PASS. No blocked dependency. Next: WCF-04 production clipboard events,
  authoritative edge-case effects, historical receipt compatibility and regression.

## WCF-04 integrated evidence

### Browser environment and interpretation

Production renderers ran through the Make-owned webserver-backed browser stack,
with isolated service fixtures and the built frontend: Ubuntu 26.04 LTS,
Playwright 1.59.1, Chromium 147.0.7727.15 (revision 1217, Linux).
The new Timeline regression uses real Ctrl+C/Ctrl+V and browser ClipboardItems;
DataTransfer events supplement malformed/unsupported representation coverage.
Desktop Excel is unavailable. CSV/TSV quoting, Excel-compatible table HTML,
leading zeros/date-looking strings and formula-neutralized exports are fixtures,
not a desktop Excel interoperability certification.

### Passing evidence

Run roots below are under `.cartulary/test-results/`; each contains its
`run-summary.json`, row logs, and browser artifacts where applicable.

| Acceptance | Evidence | Result |
| --- | --- | --- |
| Strict codec/values/geometry and unchanged kernel vectors | `20260915T202616Z-p75455` Tabular Ingest; `20260915T202617Z-p76731` full Grid Adapter (49/49 graph units) | PASS |
| Production scalar, N×1 comma, rectangle, Unicode, embedded delimiters, formula strings, empty clear; native editor and rejection without writes | `20260915T202629Z-p90127` native clipboard regression | PASS |
| Exact headers, duplicate event, distinct identical-text gestures, fill, delayed/conflicted paste and exact lost-response retry/read-only refresh recovery | Four existing production tests passed in `20260915T202105Z-p79514`; its new fidelity test was still failing an assertion at that point | PASS |
| Timeline and Entity authoritative values, source-owned creation/reuse, blank-value rejection/clear, no rejected writes/history/change sets, old receipt replay | `20260915T202746Z-p51957` expanded service-backed Workbook slice | PASS |
| Entity production create/reuse paths | `20260915T202124Z-p9146` Hosts and Identities browser rows and service slice | PASS |
| Retained operation scope, current authority, exact request capture including header interpretation | `20260915T201925Z-p58044`, plus adapter row pass in `20260915T202216Z-p43539` | PASS |
| Current window/grouped target planning, stable fields, reordered/hidden/noneditable header mapping | `20260915T202340Z-p77296` sorted-target row; grouped correction `20260915T202618Z-p77672`; header-range test `20260915T203100Z-p14466` | PASS |
| Closed-incident reads and exact authorized replay | `20260915T202834Z-p70577` Timeline source store slice | PASS |
| Native editor draft paste and ordinary surfaces | Native draft rows passed in `20260915T202216Z-p43539`; corrected surface suite `20260915T202338Z-p77026` | PASS |
| Evidence image/file routing and first creation orderings | `20260915T202217Z-p43789` production Evidence file-draft browser row | PASS |
| Shared Assessment and Network Analysis presentation consumers | `20260915T202923Z-p88776`, `20260915T202922Z-p88557`; shared production grid codec/binding above | PASS |
| Keyboard, live error announcement, focus, compact grid fixtures | Fidelity test plus `20260915T202355Z-p7104` Grid Adapter accessibility and visual rows | PASS |
| Narrow layouts and 200% zoom recovery | `20260915T202356Z-p7330` Workbook batch and autosave accessibility rows | PASS |
| Imports discovery/mapping/apply workflow | `20260915T201750Z-p39678` service-backed Imports CSV row | PASS |

### Corrections and failed-run accounting

- A newly routed browser test required generated browser topology. `make generate`
  refreshed it successfully at `20260915T201951Z-p76186`; no generated files were
  hand-edited. An attempted unknown semantic test row was rejected before execution;
  the actual catalog row passed at `20260915T203100Z-p14466`.
- Service fixture `20260915T201923Z-p57827` incorrectly passed a UUID's `ID`
  method instead of its UUID value to receipt seeding. Corrected; later service
  runs pass. Receipt seeding only occurs in isolated test databases.
- Browser `20260915T202105Z-p79514` assumed query order was creation order;
  corrected to compare the original stable destination record. Browser
  `20260915T202337Z-p76780` expected null for an explicit Timeline source clear;
  its contract stores the empty string. All earlier fidelity assertions passed;
  the corrected complete test passed at `20260915T202629Z-p90127`.
- Test-binding exports had to expose the real codec. Related frontend failures
  `20260915T202216Z-p43539` and `20260915T202239Z-p73427` are superseded by
  the corrected surface/target/grouped passes. One legacy assertion expected CSV
  for multiline one-column text; it now expects explicit TSV.
- `20260915T203014Z-p89773` stopped before browser execution because frontend
  inputs changed during the build snapshot. No product failure is inferred.
  The subsequent isolated rerun adds native external-HTML precedence evidence.

### Additional confirmed gap

Reordered exact-header mapping retained stable write keys but highlighted the
range in schema order. Remediation: `semanticClipboardPolicy.ts` derives the
visual bounds from current visible column order while leaving request mapping
unchanged; `semanticKernel.test.ts` proves both. Core 01/03 visible-column and
exact-header rules govern this correction. Benefit: selection reflects the
mapped range. Compatibility: no API or persisted-data change. Remaining risk:
noncontiguous header fields use the grid's existing bounding-range presentation.
Binary validation: reordered request keys remain unchanged and visual start/end
follow visible order; PASS at `20260915T203100Z-p14466`.

No screenshots or unrelated visual goldens were regenerated. Broader measurement
baseline regeneration is not applicable: no layout token, CSS, density or
performance-budget contract changed; focused browser accessibility/visual evidence
checks the existing presentation and recovery boundaries.

WCF-04 exit: PASS. Native external HTML and exact-header/repeated-event browser
rerun passed at `20260915T203130Z-p15148` (11/11 graph units). No applicable
acceptance row is blocked. Remaining limitations are desktop Excel/browser
coverage and documented plain-text ambiguity, not unverified product claims.

## WCF-05 final verification and scope

`make help`, `make help-all`, and current task guides for `package.grid_adapter`,
`web.workbook`, `module.workbook`, `module.timeline`, `module.entities`,
`module.tabularingest`, and `module.imports` were re-read. Narrow service/browser
and owner slices preceded broader terminal checks. New tests are routed through
existing owner families; generated topology was refreshed by Make.

`make agent-finalize` PASS at `20260915T203309Z-p46863`.
`unit-artifacts/finalize-summary.json` reports zero generated updates.
RESULTS_DIR was unset: retained-run maintenance, retained performance evidence,
and retained-run checks were explicitly skipped. No successful full warm-check
run is claimed or supplied.

| Terminal check | Result/evidence |
| --- | --- |
| Frontend typecheck | PASS `20260915T203344Z-p53000` |
| Generated drift | PASS `20260915T203344Z-p52580` |
| Generated artifact policy | PASS `20260915T203344Z-p52701` |
| JSON shapes | PASS `20260915T203344Z-p52647` |
| OpenAPI compatibility | PASS `20260915T203344Z-p52735` |
| Frontend import boundaries | PASS `20260915T202631Z-p93624` |
| Backend module boundaries | PASS `20260915T202632Z-p95488` |
| Documentation lint | PASS `20260915T203344Z-p53400` (before terminal evidence append) |
| Broad fast suite | PASS `20260915T203945Z-p45489` (681/681 graph units) |
| Final formatting | PASS `20260915T203834Z-p37014` |
| Final lint, including type and architecture checks | PASS `20260915T204324Z-p27982` (11/11 graph units) |
| Final documentation lint | PASS `20260915T204349Z-p48113` |
| Final service receipt/count regression | PASS `20260915T204300Z-p11427` |

### Final-review corrections

The broad fast suite exposed the required Entity production-export disposition
for `Store.ApplyClipboardPasteRequest`. The authored source-boundary guard now
lists the entry point used by Workbook assembly; the exact owner guard passes at
`20260915T203609Z-p92255`. This is source ownership, not permission to bypass
request authorization or verification routing.

A final-review gap affected invalid mapped relationship values accepted by legacy
API clients. They could still be stored as unmapped provenance when normalization
failed. Remediation: `timeline/clipboard_paste.go` rejects any invalid mapped
writable field, including source-owned mention/collection fields;
`clipboard_paste_test.go` proves the failure. Core 01 field contracts and Core 03
batch admission govern the correction. Benefit: invalid cells cannot silently
become unrelated provenance. Compatibility: invalid legacy payloads now reject;
valid mention-origin behavior and unknown-column raw capture are unchanged.
Unresolved risk: no new relationship write format is admitted. Binary criterion:
invalid mapped relationship returns `ErrInvalidClipboard`; PASS at
`20260915T203531Z-p79720`. No migration or analyst-data rewrite is required.

### Final scope and compatibility

The seam covers offered clipboard representations, strict decoding, grid copy,
source-specific target admission and captured request interpretation. The mapping
kernel, retained-operation owner, history, entity reuse/create, Timeline source
normalization, reference selection and Evidence upload lifetimes remain with
existing owners. Generic Workbook, Assessments, Entities, Timeline and Network
Analysis share the changed copy boundary; approved display labels remain the
only relationship presentation exported.

API change: additive optional `header_mode` (`auto` or `none`). Omission retains
Timeline exact-header recognition. Default-form receipt hashes remain unchanged;
nondefault mode participates in identity. Omitted/`auto` format now means TSV.
CSV clients must send `format: "csv"`; reconcile/drain uncommitted old auto-format
requests before deployment. Matching committed receipts are retrieved before
new lexical decoding, under current authorization/admission. Browser uncertain
retry uses the original captured body and identity. No database migration,
new endpoint, dependency, lockfile edit or persistent browser storage was added.

Retired paths: Workbook punctuation/delimiter heuristics, duplicated browser
parser/dimensions helpers, `formatGridClipboardTSV`, Go blank-record trimming and
header padding. Imports keeps its source-owned missing-cell classification and
ragged discovery geometry while sharing lexical CSV/TSV rules. Its integration
workflow and canonical mapping vectors pass.

### Limitations and scope exclusions

- Desktop Excel and non-Chromium engines were not executed. Browser MIME behavior
  is proven in the recorded Chromium environment; portable plain text remains a
  deliberately limited fallback when richer representations are stripped.
- Plain text cannot identify every multiline scalar, encoded rectangle or removed
  trailing record. Single-line comma-only text remains scalar. No guessing is
  reintroduced to mask those ambiguities.
- Extra or malformed HTML tables, merged/nested tables and unsupported cell
  content reject. This is ordinary table/scalar interoperability, not rich-text
  editing, formula evaluation or a file-import redesign.
- Release/deploy/full `check` and unrelated visual/measurement regeneration are
  out of scope. The changed slices, source boundaries, generated compatibility,
  existing density visuals and keyboard/narrow/200%-zoom checks are the evidence.
- No commits, pushes, deployment, advisory-digest edits or analyst-data changes.
  Service writes were confined to harness-created test databases/object stores.

### Rollback and next action

Baseline has no pre-existing user changes. To roll back, first preserve the final
working diff and untracked seam files, then restore only the tracked paths below
from `2e1f35dc44497be13179f3fc24753713a1b35cee` and remove only the listed new files.
Do not reset the whole repository, remove runtime volumes, or touch analyst data.
Regenerate from restored authored inputs if needed. No database rollback applies.
Next action after terminal PASS: review the diff and coordinate the API auto-CSV
migration before any separately authorized rollout.

### Changed paths

```text
apps/web/e2e/timeline-grid-entry.spec.ts
apps/web/src/workbook/WorkbookShell.sentinel.test.tsx
apps/web/src/workbook/WorkbookShell.surfaces.test.tsx
apps/web/src/workbook/adapters/README.md
apps/web/src/workbook/adapters/createWorkbookClipboardPasteAdapter.test.ts
apps/web/src/workbook/components/EntityWorkbookSurface.tsx
apps/web/src/workbook/components/GenericWorkbookSurface.tsx
apps/web/src/workbook/features/entities/useEntityClipboardPasteController.ts
apps/web/src/workbook/models/entityClipboardPastePlan.test.ts
apps/web/src/workbook/models/entityClipboardPastePlan.ts
apps/web/src/workbook/timeline/composition/useTimelineInteractionComposition.ts
apps/web/src/workbook/timeline/hooks/useTimelineClipboardPasteController.ts
apps/web/src/workbook/timeline/models/timelineClipboardPastePlan.test.ts
apps/web/src/workbook/timeline/models/timelineClipboardPastePlan.ts
apps/web/src/workbook/utils/README.md
apps/web/src/workbook/utils/workbookClipboard.test.ts
apps/web/src/workbook/utils/workbookClipboard.ts
contracts/openapi-releases/2.0.0.change-set.json
contracts/openapi-source/owners/module.workbook/openapi.json
contracts/openapi/cartulary.openapi.yaml
contracts/tabularingest/clipboard.v1.json
docs/handoffs/ui-ux/workbook-clipboard-fidelity-refactor-handoff.md
docs/spec/01_architecture_storage_and_view_contracts.md
docs/spec/03_workbook_interaction_collaboration_and_workflows.md
docs/spec/04_security_deployment_and_conformance.md
internal/app/workbookassembly/action_adapters.go
internal/app/workbookassembly/entity_adapters.go
internal/gen/contractopenapi/artifacts_gen.go
internal/gen/openapioperations/catalog_gen.go
internal/modules/entities/boundary_guard_test.go
internal/modules/entities/hostidentity/clipboard_paste_api.go
internal/modules/entities/hostidentity/clipboard_paste_api_test.go
internal/modules/entities/hostidentity/clipboard_paste_store.go
internal/modules/tabularingest/clipboard_codec.go
internal/modules/tabularingest/tabularingest.go
internal/modules/tabularingest/tabularingest_test.go
internal/modules/timeline/admission/batch.go
internal/modules/timeline/admission/batch_test.go
internal/modules/timeline/batch_mutation_store.go
internal/modules/timeline/clipboard_paste.go
internal/modules/timeline/clipboard_paste_test.go
internal/modules/timeline/commands.go
internal/modules/timeline/facade.go
internal/modules/workbook/clipboard_paste_integration_test.go
packages/grid-adapter/README.md
packages/grid-adapter/src/SemanticDataGrid.tsx
packages/grid-adapter/src/clipboardCodec.test.ts
packages/grid-adapter/src/clipboardCodec.ts
packages/grid-adapter/src/core.test.ts
packages/grid-adapter/src/core.ts
packages/grid-adapter/src/index.test.tsx
packages/grid-adapter/src/index.tsx
packages/grid-adapter/src/rdgCompiler.tsx
packages/grid-adapter/src/semanticClipboardPolicy.ts
packages/grid-adapter/src/semanticKernel.test.ts
packages/grid-adapter/src/semanticPresentation.ts
packages/grid-adapter/src/test-support.tsx
packages/protocol-ts/src/generated/core-http-types.ts
packages/protocol-ts/src/generated/core-http-validators.ts
tools/browser_e2e_batch_manifest.json
tools/execution_topology_render_index.json
tools/test_families/module.timeline.json
tools/test_families/module.workbook.json
tools/test_families/package.grid_adapter.json
tools/test_families/web.workbook.json
tsconfig.base.json
```

The first broader `make test-fast` completed 680/681 units at
`20260915T203344Z-p53238`; its only failure was the related Entity export
allowlist, corrected and passing in the focused rerun above. Final `make format`
PASS at `20260915T203834Z-p37014`; prior Biome failure
`20260915T203344Z-p53087` identified only formatting in the final header-range
implementation and test. No unrelated baseline failure was found.

### Receipt target-count admission

Final review also confirmed that historical Entity request hashes omit targets.
Moving parsing after receipt lookup therefore required preserving target-count
admission through a parser-independent check. `clipboard_paste_store.go` compares
requested create targets with the stored receipt's row count before returning a
replay; the existing prebuilt-plan entry retains its original plan validation.
`clipboard_paste_integration_test.go` proves mismatched targets reject while the
matching legacy receipt still replays. Core 01 request/target admission owns this
rule. Benefit: compatibility does not weaken admission. Migration: no hash or
stored-receipt rewrite. Risk: malformed receipts fail closed. Binary validation:
matching count returns the identical receipt, mismatched count returns 400,
with no new change set. Final service result is recorded below.

New Imports CSV discovery inherits the explicit record-ending grammar: blank
records remain source rows, and quoted line endings reach the Imports adapter
unchanged. Existing discovery artifacts and committed records are not rewritten.
Imports still decides blank-cell classification, mapping omission/null policy,
headers and discovery limits. This lexical correction does not redesign import
approval or application.

## Terminal completion

- Second `make agent-finalize`: PASS `20260915T203857Z-p41335`, before the
  final broad verification. RESULTS_DIR remained unset and retained-run
  maintenance remained skipped.
- `make test-fast`: PASS `20260915T203945Z-p45489`, 681/681 graph units.
  The final receipt-count body check was also executed in the subsequent
  expanded Workbook service slice, PASS `20260915T204300Z-p11427`.
- Final `make lint`: PASS `20260915T204324Z-p27982`, 11/11 graph units.
- Final `make lint-markdown`: PASS `20260915T204349Z-p48113`.
- All applicable acceptance rows above are PASS. No BLOCKED rows or unrelated
  baseline failures remain. Full release/check and unrelated visual/measurement
  regeneration are N/A to this bounded seam for the reasons recorded above.
- Final diff/UTF-8/line-ending/trailing-whitespace checks passed. Branch remains
  `main` at the captured HEAD; all work is uncommitted. No pre-existing changes
  were overwritten. The final changed-path inventory above is complete.
- WCF-05 exit: validation and terminal handoff complete; next action is review.
