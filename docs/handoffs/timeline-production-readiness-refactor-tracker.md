# Timeline Production-Readiness Refactor Tracker

## Purpose and authority

This tracker records execution of the Timeline durable-architecture and
production-readiness refactor that follows commit `44c2ca0e`. The earlier
`timeline-module-refactor-tracker.md` is retained only as dated historical
evidence.

This file is a human handoff. It is not executable authority. Requirements,
schemas, manifests, verification ownership, and harness inputs under
`contracts/` and `tools/` own testable behavior.

## Baseline

- Starting commit: `44c2ca0e` (`Timeline Module Remediation`).
- Starting worktree: clean.
- Discovery source: live repository tree, not the historical tracker.
- Reserved migrations: `00042` through `00048`; historical migrations remain
  immutable.
- Public behavior retained unless explicitly versioned: Workbook HTTP and
  OpenAPI shapes, normalized hashes and error envelopes, Timeline field keys,
  `cartulary.view.timeline.v2`, saved views, and WebSocket v1.

## Slice ledger

| Slice | State | Authority and scope | Migration/deployment state | Verification and result roots | Temporary mechanisms and deletion evidence | Successor handoff |
|---|---|---|---|---|---|---|
| S0 | Complete | Added active owner requirements for complete Timeline projection input, normalized provenance, application-assembled projection catalog, exclusive/bounded Collaboration, Timeline Bundle versions, and single/replicated process models | No production or data change | `json-shape-check` pass `.cartulary/test-results/20260726T135619Z-p3138636`; `generate-drift` pass `.cartulary/test-results/20260726T135621Z-p3139147`; `harness-contract` pass; Timeline characterization 50/50 pass `.cartulary/test-results/20260726T135713Z-p3144399` | Every later compatibility path names its deletion slice; generated topology refreshed from owner inputs | S1 has consistent active authority and a passing Timeline behavior baseline |
| S1 | Complete | Workbook owns HTTP decoding, stable-target admission, route hashes, and error envelopes; Timeline accepts typed commands. Constructors require an explicit conflict-token codec and `timelineassembly.Bundle` is the single runtime graph | Code-only; prior binary remains the rollback unit | Timeline owner 50/50 pass `.cartulary/test-results/20260726T140126Z-p3176436`; server build pass `.cartulary/test-results/20260726T150256Z-p3554234` | Deleted Timeline routes/decoders/hash/error exports, private type aliases, `server.Options.TimelineDependencies`, mutable codec setters/default test key, timing context/header producer, and repeated assembly construction | S2 receives one stable Timeline bundle and explicit security dependencies |
| S2 | Complete | `internal/app/projectionassembly` assembles and validates the code-backed provider catalog; Projections exposes provider-specific services | Code-only; a prior binary can restore the old registry | Projections pass `.cartulary/test-results/20260726T154543Z-p3764463`; final complete service-backed owner pass `.cartulary/test-results/20260726T171008Z-p4165458` | Deleted broad `RowProjector`/`TimelineWriter`, adapter package, central built-in switch, duplicated provider maps, and dual catalog | S3 receives provider-specific storage/query/rebuild services |
| S3 | Complete | Timeline’s canonical projection input includes typed mentions, evidence references, and tags; caller-transaction owner fact ports replace peer SQL; invalidation uses exact affected IDs | Authored `00042` expand and `00043` contract; rollback is old-query cutback before `00043`, forward rebuild after it | Initial Timeline run exposed and drove a derivation fix `.cartulary/test-results/20260726T155740Z-p3852207`; server build then passed `.cartulary/test-results/20260726T160109Z-p3885728`; Projections parity/rebuild coverage is included in `.cartulary/test-results/20260726T171008Z-p4165458` | Deleted live lateral collection derivation, monolithic source query, Timeline assembly peer-table SQL, SQL exceptions, and ordinary full-incident Host/Identity rebuild | S4 receives one complete canonical row and exact peer effects |
| S4 | Complete | A code-backed producer catalog assigns record, job, and Network Flow intents to source owners; intent-key collisions are immutable | Authored `00044`; old binaries and trigger restoration are a coordinated rollback only | Collaboration focused pass `.cartulary/test-results/20260726T161321Z-p3973280` | Deleted all record/job/Network Flow trigger producers and payload-overwriting conflict behavior; no dual producer remains | S5 receives exclusive deterministic source-owner intents |
| S5 | Complete | Timeline provenance is bounded typed source state; Bundle v2 uses logical owner codecs; v1 is import-only compatibility with executable expiry metadata | Authored `00045` expand and `00046` contract; snapshot/forward repair is required after contract migration | Incident Bundle owner pass `.cartulary/test-results/20260726T161501Z-p3976048` | Deleted `timeline_events.raw_capture`, dual writes, generic physical Timeline import, and v1 export. The sole retained v1 importer is owned by `contracts/incident-bundles/compatibility.json` | S6 receives bounded logical source state and immutable producers |
| S6 | Complete | Collaboration sequences and quarantines independently per incident, pages and bounds replay, tails durable rows in every process, prunes by explicit policy, and disconnects slow consumers | Authored `00047` expand and `00048` contract; old dispatcher is available only before `00048`, otherwise forward repair from replay rows | Migration drift pass `.cartulary/test-results/20260726T162030Z-p3985752`; focused durable-stream pass `.cartulary/test-results/20260726T162847Z-p4044659`; full owner run otherwise passed but reported an unrelated pre-existing visual mismatch `.cartulary/test-results/20260726T162043Z-p3988767` | Deleted global sequencer lock, `delivered_at`, local-hub delivery acknowledgement, unbounded replay allocation, infinite retry, and silent buffer drop | S7 receives durable replay and process-local tailers |
| S7 | Complete | Active Core04/app.server authority defines default `single` and gated `replicated`; replicated admission requires shared durable bindings, publication-plan agreement, and component-scoped leaders | Code/config plus managed-storage copy procedure; fail closed and scale to one before returning to `single` | Config pass `.cartulary/test-results/20260726T163629Z-p4061255`; an initial server row exposed a fixture construction defect `.cartulary/test-results/20260726T163637Z-p4061771`; corrected replicated fencing row passes `.cartulary/test-results/20260726T164110Z-p4113215` and `.cartulary/test-results/20260726T171115Z-p4168403` | No replicated filesystem fallback or mixed process model exists; owner-specific shared storage adapters remain with their owners and generic byte storage remains platform-level | S8 targets the final reset, replay, and access-loss model |
| S8 | Complete | `TimelineWorkbook` accepts one required shell-owned runtime; lifecycle transitions are pure reducer actions and controller hooks own effects | Independent static asset; previous web asset is rollback | Runtime row pass `.cartulary/test-results/20260726T165253Z-p4133253`; representative frontend rows pass `.cartulary/test-results/20260726T165515Z-p4134522`; frontend typecheck pass `.cartulary/test-results/20260726T170713Z-p4152826` | Deleted uncontrolled query/layout/entity state, optional production callback fallbacks, direct component test composition, source-name policy test, and browser timing global | S9 receives one production composition and deterministic transition fixtures |
| S9 | Complete | Historical harness routes/SQL, implementation-history rows, broad allowlists, stale documentation, and test-only production constructors are removed; final rediscovery reconciles the live tree with the disposition ledger | No new migration; implementation is release-ready but has not been deployed. Production must still follow the staged `00042`–`00048` rollout and observation windows below | Final Timeline owner slice passes 9/9 work units and 46 tests `.cartulary/test-results/20260726T202007Z-p2664140`; current-tree `check` passes 172/172 work units and 712 tests `.cartulary/test-results/20260726T202538Z-p2704323`; retained-run finalization passes `.cartulary/test-results/20260726T202809Z-p2800494` | Deleted Timeline harness routes, Timeline-owned RDG vendor row, legacy keyset naming, production test codec constructor, obsolete source query, stale tracker guidance, superseded verification selections, and the last private test aliases. Only the ledgered v1 importer and `single` mode remain | Final handoff records the deployment sequence, rollback boundaries, failure ledger, retained compatibility triggers, and live-tree search results |

## Final implementation closure

The implementation is complete in the working tree. Production deployment was
not performed as part of this repository change. The migrations and process
model must be rolled out in this order:

1. Deploy the code-only Workbook admission, Timeline assembly, and projection
   catalog changes.
2. Apply `00042`, rebuild and compare complete Timeline projection rows, switch
   readers, then apply `00043` only after parity is clean.
3. Observe source-owner Collaboration producer parity before applying `00044`.
4. Apply `00045`, preflight and backfill provenance, activate Bundle v2, drain
   old writers, then apply `00046`.
5. Apply `00047`, shadow the bounded stream, cut API processes to durable
   tailing, then apply `00048`.
6. Copy durable filesystem publications to shared managed storage, verify
   hashes and publication-plan agreement, and only then enable `replicated`.
7. Deploy the web asset after the final Collaboration behavior is active.

No contract migration is compatible with an old binary continuing to admit
writes. Before each contract migration, rollback is a binary/query cutback to
the expanded schema. After `00043`, `00044`, `00046`, or `00048`, rollback is a
forward repair or a coordinated snapshot restore; an old binary must not be
started against the contracted schema.

The final frontend correction found during release closure is deliberately
query-aware: after an accepted Timeline mutation, an authoritative refresh runs
when filters, sorting, or grouping are active. This prevents a pre-commit
refresh from leaving a row visible after the mutation changes its filter or
ordering fields, while retaining the low-latency response-row path for the
default-query case.

## Verification closure

The final release root
`.cartulary/test-results/20260726T201244Z-p2508798` proves:

- generated, JSON-shape, migration, boundary, and harness contracts;
- frontend type checking, unit behavior, import boundaries, and Biome checks;
- OTel conformance, targeted security analysis, and server/migrate/web builds;
- backend fast and service-backed coverage;
- browser webserver, stateful, accessibility, and visual readiness;
- 12 of 12 release work units and 712 tests with no failure or missing
  evidence.

After the final test-only alias deletion, a fresh source-matched `check` passed
172 of 172 work units and 712 tests at
`.cartulary/test-results/20260726T202538Z-p2704323`. `agent-finalize` accepted
that exact retained run, refreshed the harness and service-backed duration
baselines, regenerated topology, and passed all drift checks at
`.cartulary/test-results/20260726T202809Z-p2800494`.

Additional retained evidence:

- full backend test: `.cartulary/test-results/20260726T185950Z-p1272425`;
- full service-backed test:
  `.cartulary/test-results/20260726T194721Z-p2100554`;
- clean CI aggregate: `.cartulary/test-results/20260726T195353Z-p2187941`;
- clean fast test after lint cleanup:
  `.cartulary/test-results/20260726T193420Z-p1743111`;
- focused Timeline browser support after the query-refresh fix:
  `.cartulary/test-results/20260726T201103Z-p2500370`;
- final Timeline owner slice:
  `.cartulary/test-results/20260726T202007Z-p2664140`;
- accepted visual update:
  `.cartulary/test-results/20260726T175129Z-p306323`.

The final visual diff contains four intentional artifacts: entity mention-chip
states, Network Flow graph contributors, Timeline transaction recovery, and
the Workbook inspector public-error state. The trigger was the single runtime
composition and deterministic conflict/recovery rendering. No viewport, zoom,
mask, or screenshot-scope exception was introduced.

## Failure and remediation ledger

- `make test` initially encountered a test-server port collision at
  `.cartulary/test-results/20260726T185512Z-p1186880`. It was infrastructure
  contention, no rollback was needed, and the unchanged full retry passed.
- `make test-fast` reported one Host inspector fixture failure at
  `.cartulary/test-results/20260726T193212Z-p1720693`. The exact row passed at
  `.cartulary/test-results/20260726T193401Z-p1742622`, and the unchanged suite
  retry passed.
- the first CI aggregate failed duration drift at
  `.cartulary/test-results/20260726T193957Z-p1891778` because the retained
  baseline contained command overhead from a contaminated run. A clean full
  service-backed run was accepted with the owner-provided
  `ALLOW_COMMAND_OVERHEAD_DECREASE=1` control; baseline coverage, drift, and CI
  then passed.
- the first release aggregate failed Timeline browser support at
  `.cartulary/test-results/20260726T195647Z-p2296945`. Trace evidence showed a
  reviewed material edit correctly demoted the row to `enriched`, while the
  historical test still assumed it remained in a `reviewed` filter.
- the first corrected browser run at
  `.cartulary/test-results/20260726T200546Z-p2473679` exposed the underlying
  product race: a query could complete before the mutation commit and then the
  response row could bypass filter re-evaluation. The mutation controller now
  performs one post-commit authoritative refresh whenever query controls can
  affect membership or ordering. The focused browser suite and final release
  aggregate pass.
- an unconditional form of that refresh caused expected fetch-count failures
  in `.cartulary/test-results/20260726T200836Z-p2479606`. It was narrowed to
  active filter/sort/group queries. A later frontend run at
  `.cartulary/test-results/20260726T200944Z-p2482008` had one unrelated Entity
  merge fixture failure; the unchanged retry passed.
- test-catalog renaming temporarily failed generation/JSON ordering at
  `.cartulary/test-results/20260726T200447Z-p2454344` and
  `.cartulary/test-results/20260726T200513Z-p2457383`. The authored owner row
  was ASCII-sorted and regenerated at
  `.cartulary/test-results/20260726T200522Z-p2457936`.
- retained-run finalization first rejected
  `.cartulary/test-results/20260726T194402Z-p2002084` at
  `.cartulary/test-results/20260726T202525Z-p2703816` because the final alias
  deletion changed the source snapshot digest. A new full `check` was run
  rather than weakening retained-evidence validation; finalization then passed
  against the current digest.

No failure required reverting a production slice. Later stages were held until
the related focused target passed.

## Final rediscovery

Live-tree searches confirm:

- no Timeline transport decoder, HTTP error mapper, route hash implementation,
  mutable conflict-token setter, production test codec, timing header/global,
  production module override, broad dependency bag, or private compatibility
  alias remains;
- no Timeline query or assembly SQL reads Mentions, Links, Evidence, aliases,
  or object blobs; query SQL reads `timeline_grid_projection` and the common
  record envelope only;
- no cross-owner Timeline SQL exception, central built-in projection switch,
  broad projection adapter, duplicate view-ID registry, ordinary
  full-incident entity rebuild, or second collection derivation remains;
- Collaboration intent keys use exact-duplicate `DO NOTHING` plus explicit
  byte/identity comparison; no payload overwrite remains;
- trigger names and `delivered_at` occur only in reversible migration-down
  sections, not in runtime behavior;
- `raw_capture` storage occurs only in the expand/contract migration and the
  bounded Bundle v1 compatibility decoder. The active product policy name
  `preserve_raw_capture` remains public, but it writes normalized provenance;
- no Bundle v1 exporter, Timeline physical-row import target, Timeline test
  route, source-name policy test, uncontrolled Timeline composition, or
  implementation-history browser row remains;
- no temporary dual write, shadow verifier, migration switch, or runtime
  fallback remains after the contract migrations.

The requested clean-worktree exit criterion is interpreted as clean generated
and formatting drift. The implementation is intentionally left as an
uncommitted reviewable diff; committing it remains caller-controlled.
`git diff --check` is the whitespace-integrity gate.

## Required per-slice handoff

Each completed row must record:

- executable authority changed and contract identifiers;
- files and packages affected;
- compatibility preserved, changed, versioned, or removed;
- migration identifiers, deployment state, and rollback boundary;
- focused commands, outcomes, retained result roots, and skipped checks;
- temporary mechanisms introduced, owner, deadline, and deletion evidence;
- failures and whether they were related to the slice;
- successor entry conditions and unresolved risks.

## Intentionally retained compatibility

- Workbook HTTP/OpenAPI behavior, hashes, error envelopes, Timeline field keys,
  saved views, and `cartulary.view.timeline.v2` retain product identity.
- WebSocket v1 retains current client and recovery value; bounded history gaps
  continue to use `reset_required`.
- `owner_batch_apply_v1`, `tabular_row_plan_v1`, and
  `record_change_intent_v1` remain active typed contracts.
- Fixed-offset time profiles, generation flags, capture lifecycle, and
  normalized source provenance remain Timeline product semantics.
- Bundle v1 import remains temporarily for supported recovery archives under
  the executable compatibility deadline and usage/inventory deletion gates.
- The single-process model and filesystem roots remain supported for
  disconnected, development, and non-shared-storage installations.
- The code-backed projection catalog and physical projection tables remain
  authoritative because they provide deterministic rebuild and production
  query behavior.
- The historical tracker remains dated forensic evidence and has no machine
  consumer.
