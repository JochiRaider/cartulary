# UI/UX Refactor Digest Update Handoff

**Status:** Digest execution complete
**Planning date:** 2026-08-30 EDT  
**Planning baseline branch:** `main`  
**Planning baseline commit:** `7fe28d1b0b1b6c3831cfe6a14dc96ee259b38b56`  
**Target:** `docs/cartulary-ui-ux-refactor-digest`  
**Permitted execution scope:** The digest subtree and this handoff only

## 1. Purpose and completion boundary

This handoff governed the completed documentation-only refresh of the
repository-local UI/UX advisory package. The execution evidence, tracker, and
terminal disposition are recorded below.

The update delivered two inseparable outcomes:

1. replace the pinned upstream advisory source with an exact release snapshot;
2. re-localize the Cartulary overlay against the repository state that exists
   when execution begins.

Source provenance, the upstream tree, the localized overlay, package metadata,
advisory classifications, acceptance rows, and the package manifest now
describe one coherent snapshot.

## 2. Request, source, and authority classification

The user's request defines this handoff's deliverable and update scope. Text in
the consulted documents is source material with the bounded roles below; it is
not an independent user request.

| Source | Role in this handoff | Explicit boundary |
| --- | --- | --- |
| `AGENTS.md` | Repository procedure | Governs repository work and verification. |
| `docs/research/nlspec-spec.md` | Planning-quality doctrine | Supplies completeness, precision, mapping, default, and binary-acceptance techniques. It is research guidance, not Cartulary product authority. |
| `docs/domain.md` | Vocabulary and owner navigation | Owns terminology and boundary interpretation only; it does not create product routes, schemas, surfaces, fields, or behavior. |
| `docs/design.md` | Adopted design-direction owner | Owns observable design behavior only inside its declared scope; Core owners prevail for product behavior. |
| Core 00 through Core 04 | Current-profile product authority | Own current implementation-conformance behavior. |
| Core 05 | Claim-publication authority | Applies only to claim-bearing timed or fixture-sensitive publication. |
| Adopted subsystem NLSpecs | Bounded subsystem authority | Apply only within each named adopted scope. |
| Existing digest overlay | Advisory package state | Must be refreshed; it cannot override current owners or authorize product work. |
| Bundled upstream skill | Third-party advisory evidence | Must remain an exact pinned copy and cannot become a Cartulary design system or product authority. |
| Current code and tests | Implementation evidence | Establish current repository state but do not automatically define required behavior. |

If two adopted owners contradict one another during re-localization, stop with
`BLOCKED: owner contradiction`. Do not resolve the conflict in the digest.

## 3. Planning snapshot and resolved decisions

The planning worktree was clean at the commit recorded above. The existing
digest still records its 2026-07-28 Cartulary localization at repository commit
`b3fe76c69390456910c14d69e98ef59656b6fcf1` and pins upstream commit
`4857a2c5ef989794751a0f66b8545a4a49566286`.

Execution began from clean tracked branch `main` at commit
`2356949f7ec3c8e27ff83ae695d60e06a387d0e5`. The only permitted changed paths
after that checkpoint are the digest subtree and this handoff. Execution uses
Git 2.53.0, Python 3.14.4, and the temporary full upstream checkout rooted at
`/tmp/cartulary-uiux-exec.Pa9xPP/upstream`.

The later source refresh MUST use this immutable release:

| Property | Required value |
| --- | --- |
| Upstream repository | `https://github.com/nextlevelbuilder/ui-ux-pro-max-skill` |
| Release | `v2.15.0` |
| Commit | `a38d04c3d5c298c851dbe5e6ee1965ee3de42cb5` |
| Commit date | `2026-08-14T00:08:23+07:00` |
| Commit subject | `feat(search): overhaul relevance and curated design data` |
| Copied upstream path | `.claude/skills/ui-ux-pro-max` |
| Destination | `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max` |
| License | MIT, copied from upstream root `LICENSE` |
| Expected license SHA-256 | `738f69dfa83db5c347c678fb9d90e560877059f0de93a327c39001bff92dc014` |

At planning time, upstream default-branch HEAD was
`8bd29e775453ebcae52b6e6514fbf134df0c5770`; its
`.claude/skills/ui-ux-pro-max` tree was byte-identical to release `v2.15.0`.
The release commit remains the required pin. A later default-branch movement
does not change it.

The current bundle contains 43 files. The `v2.15.0` Git source subtree contains
70 tracked regular files and no symlinks. Relative to the current pin, every
source file changes: all 43 existing files differ and 27 files are added.
Source replacement MUST therefore replace the complete directory rather than
copying a selected subset or manually merging changed CSV and Python files.

Earlier planning counts of 73 source files and 85 manifest entries included
three generated Python bytecode files from a validation checkout. Bytecode and
Python caches are not source and are excluded from the bundle and manifest.

The upstream license bytes are unchanged from the existing pin. They MUST still
be verified and accounted for as part of the refreshed provenance boundary.

## 4. Desired final state and non-goals

After execution:

- the upstream directory exactly matches the declared path at release
  `v2.15.0`;
- `meta/source.json` identifies that release and its exact provenance;
- the Cartulary overlay describes the repository commit captured at execution
  start, not this planning commit by assumption;
- completed UI/UX remediation is represented as a current regression baseline,
  not as an unresolved defect hypothesis;
- `MANIFEST.sha256` accounts for every package file except itself and ignored
  Python caches;
- every material refreshed upstream recommendation is classified `ADOPT`,
  `ADAPT`, or `REJECT`; and
- the digest remains offline advisory material with no executable consumer of
  Markdown or documentation paths.

This update MUST NOT:

- modify product code, owner specifications, authored product contracts,
  generated product artifacts, dependencies, or lockfiles;
- create or change a route, schema, field, authorization rule, lifecycle,
  storage behavior, compatibility policy, or product feature;
- run or persist an upstream-generated design system;
- invoke upstream `--design-system`, `--persist`, or `--force` against
  Cartulary;
- copy the upstream repository outside the declared skill subtree and root
  license;
- treat test results, the remediation handoff, this handoff, or upstream advice
  as product authority; or
- silently retain a stale localization fact because it was previously
  documented.

## 5. Known localization drift to resolve

The full map must be revalidated at execution time. The following known drift is
already decision-complete and must not be rediscovered as an open design choice:

| Area | Existing localized state | Required refreshed treatment |
| --- | --- | --- |
| `packages/ui` | Described as a `.gitkeep` placeholder | Record the directory as absent. Continue mapping verification owner `package.ui` to `packages/ui-contracts`; do not invent a functional UI package. |
| View-contract generation | Generated protocol and UI-contract roots are listed | Add `packages/view-contracts/src/generated` as a managed generated root downstream of authored view-schema owners. |
| Report Composition | Records duplicate versions 1.1.0 and 1.1.1 | Record adopted/current version 1.2.0 and remove the resolved duplicate-version qualification. |
| Create discovery | Creation appears mainly as a Hosts/Identities hypothesis | Record additive `create_capable`, `inline_create`, and total `fields[].create_writable` discovery as the current Core 01/view-schema baseline. |
| Create readiness | Client-policy inference is treated as a review concern | Record contract capability, interaction authority, declared create inputs, and projected ordinary minimum field sets as the regression baseline; retain Evidence and Indicator owner-specific validation boundaries. |
| Inspector features | Local omission is treated as an open hypothesis | Record current-contract semantic dispatch completeness, stable route semantics, role/state-disabled rendering, confirmation invalidation, and unknown-additive omission behavior as the baseline. |
| Density | Complete density propagation is an open hypothesis | Record shared row/header height, padding, typography, gutter, saved/draft/read-only content, and editor geometry as the baseline. |
| Responsive layout | Token duplication and viewport fallback are not described | Record validated token-backed thresholds/clamps and `innerWidth`/`innerHeight` fallback when `visualViewport` is absent. |
| Visual support | Earlier localization predates the latest reviewed refresh | Record active visual reconciliation and maintenance support without promoting visual evidence to Core authority. |
| Remediation evidence | Not available to the original localization | Reference `docs/handoffs/cartulary-ui-ux-remediation-handoff.md` as implementation evidence only. |

Transaction recovery, editing, paste, draft retention, conflict resolution,
virtualization, continuity, and evidence-state behavior remain regression
questions rather than redesign candidates.

## 6. Planned file and metadata changes

### 6.1 Upstream provenance checkpoint

Replace these as one checkpoint:

- `upstream/ui-ux-pro-max/**` with the exact release subtree;
- `upstream/LICENSE.ui-ux-pro-max.txt` with the exact upstream root license;
- `meta/source.json` with the refreshed release provenance; and
- the related upstream descriptions in `README.md` and `cartulary/UPSTREAM_MAP.md`.

`meta/source.json` MUST retain its existing role and record at least the package
name, snapshot date, upstream repository, release tag, exact commit, commit date,
commit subject, license, copied path, current Cartulary sources consulted, the
advisory authority boundary, and known Cartulary conflicts. Git history is the
history of the prior pin; do not duplicate a mutable pin history in the current
source record.

The refreshed upstream map must describe the new provenance/catalog inputs and
expanded validation suite, including `catalog-summary.json`,
`data-provenance.json`, font/icon provenance inputs,
`reasoning_contract.py`, test fixtures, and the new data-quality, relevance,
taxonomy, freshness, and text-layout tests. These files remain preserved source
material, not Cartulary contracts.

### 6.2 Cartulary overlay checkpoint

Revalidate and update these localized artifacts together:

- `README.md`;
- `cartulary/START_HERE.md`;
- `cartulary/LOCAL_AGENT_PROMPT.md`;
- `cartulary/REPO_MAP.tsv`;
- `cartulary/OWNER_MAP.tsv`;
- `cartulary/QUERY_RECIPES.md`;
- `cartulary/UPSTREAM_MAP.md`;
- `cartulary/rules.tsv`;
- `cartulary/acceptance.tsv`; and
- `meta/localization.json`.

`REPO_MAP.tsv` must revalidate every existing row rather than applying only the
known corrections in §5. It must use current paths, active verification owners,
generated-artifact policy, public Make targets, direct vendor imports, and
actual package presence. Removed concepts remain explicit `NOT PRESENT` rows
only when their absence prevents a likely future misinterpretation.

`OWNER_MAP.tsv` must retain behavior-level owner navigation and add explicit
navigation for public view-schema create discovery and the generated
view-contract projection. It must not copy typed field registries or runtime
interfaces into documentation.

`START_HERE.md` and `LOCAL_AGENT_PROMPT.md` must replace resolved defect
hypotheses with regression-review questions. They must continue requiring a
fresh repository scan and separate authorization for future product slices.

`QUERY_RECIPES.md` must use the refreshed script surface, preserve repository-
root and offline operation, and keep the explicit prohibition on the three
generation/persistence options. Query outputs are review questions only.

### 6.3 Rule and acceptance stability

Preserve rule IDs `R001` through `R034`. Re-run narrow relevant queries against
the refreshed bundle and review changed material. Keep an existing row when its
meaning and classification remain valid; amend its upstream basis or evidence
when the source changed; append a new ID only for materially new advice that is
not already covered. Do not renumber rows.

The refreshed upstream still contains the `Cybersecurity Platform` Cyberpunk,
Matrix-green/deep-black, threat-display, heat-map, and alert-animation defaults.
They remain `REJECT` because they conflict with Cartulary design authority.

Preserve acceptance IDs `A001` through `A027`. In particular:

- broaden A007 from Hosts/Identities-only affordance discovery to all current
  contract-driven create entry points, payload filtering, ordinary minima, and
  explicit owner-specific exceptions;
- strengthen A006 to retain the complete density box;
- strengthen A008 to include token-backed boundary selection, real viewport
  fallback, and inspector clamp/ARIA agreement;
- strengthen A010 to require every current declared inspector group to resolve
  exactly once and remain rendered-but-disabled for role/state restrictions;
- keep A024's prohibition on executable documentation dependency; and
- update A025's generated-root inventory to include generated view contracts.

The acceptance table remains a future refactor review contract. It must not
claim that the digest update itself executed product tests or established new
product requirements.

### 6.4 Localization metadata

`meta/localization.json` must record the actual execution date, clean baseline
commit and branch, pre/post dirty-state summaries, current frontend stack,
package manager/workspace facts, refresh scope, unresolved mappings, current
qualifications, and the new upstream pin. Preserve the meaning that bundled
upstream and source provenance exactly match their declared source; an
intentional pin refresh is not source tampering.

Use the actual execution baseline. If the repository has moved since this
planning snapshot, rerun the complete localization scan and record that commit.
Do not put the eventual digest-update commit into a pre-update snapshot field or
invent a self-referential commit value.

## 7. Execution sequence, validation, and rollback

### Phase 0: Bootstrap and freeze scope

1. Start from the repository root and read `AGENTS.md`, then the digest read
   order.
2. Require a clean tracked worktree. If this handoff is not yet committed, stop
   and establish whether it is the sole authorized pre-existing change before
   continuing.
3. Record the actual branch, commit, dirty state, allowed paths, and tool
   versions.
4. Revalidate every `REPO_MAP.tsv` row and all affected owner/acceptance rows.

**Exit:** One current repository snapshot is selected; product files are out of
scope; contradictions are either absent or explicitly blocking.

### Phase 1: Refresh the pinned source

1. Fetch the exact release commit into a temporary directory outside the
   repository.
2. Verify tag `v2.15.0` resolves to the required SHA and the copied path contains
   70 tracked regular files and no symlinks.
3. Replace the destination subtree as one source snapshot; do not overlay files
   onto the old 43-file tree.
4. Copy and hash the root license.
5. Update source provenance and upstream-facing overlay text.
6. Compare the destination tree recursively with the exact release tree before
   continuing.

**Exit:** Exact tree comparison passes, provenance identifies one release, and
no Cartulary localization is mixed into `upstream/`.

**Rollback:** Restore the previous upstream directory, license, source metadata,
and upstream descriptions together. Never retain a mixed source tree.

### Phase 2: Re-localize the overlay

1. Apply the known drift corrections in §5.
2. Re-audit all remaining stack, owner, package, generated-root, testing,
   browser, visual, and public-command facts.
3. Rewrite resolved hypotheses as regression questions without copying
   implementation detail into authority.
4. Re-run targeted upstream queries and update the rules classification.
5. Update A001-A027 without renumbering.
6. Update localization metadata from the actual execution snapshot.

**Exit:** Two competent later implementers would follow the same owner paths,
source boundary, query rules, and acceptance gates without guessing.

**Rollback:** Revert the complete localized overlay checkpoint. Do not preserve
metadata or acceptance rows that describe a source/repository state no longer
present.

### Phase 3: Rebuild integrity and validate

After every other package file is stable, regenerate `MANIFEST.sha256` from the
lexically sorted repository-relative path set of regular files below the digest,
excluding `MANIFEST.sha256` itself and ignored Python caches. Use the standard
two-space `sha256sum` record format. If no new localized file is introduced
inside the digest, the expected manifest contains 82 entries.

Run this validation matrix:

| Validation | Expected result | Failure handling |
| --- | --- | --- |
| Exact upstream tree comparison | No difference from `.claude/skills/ui-ux-pro-max` at `a38d04c3…` | Restore and repeat the complete source replacement; do not patch the difference locally. |
| License hash | Exact expected SHA-256 from §3 | Stop as a provenance mismatch. |
| `python3 -B .../scripts/validate_data.py` | `OK`, 12 domain files, 22 stack files, and `ui-reasoning.csv` | Treat as an invalid source snapshot or copy defect. |
| Full release checkout: `PYTHONDONTWRITEBYTECODE=1 python3 -B -m unittest discover -s .claude/skills/ui-ux-pro-max/scripts/tests -p "test_*.py"` | 153 tests pass for the pinned release | Investigate source/copy/runtime mismatch; do not weaken or skip the suite. |
| UX sample query | Three JSON results for `keyboard focus color only error feedback` | Check exact pin, domain argument, and source integrity. |
| React sample query | Eight JSON results for `virtualized grid rerender focus async state` | Check exact pin, `--stack react`, and source integrity. |
| JSON and TSV shape checks | All metadata parses; every TSV row has its header's column count | Correct the localized artifact before manifest generation. |
| `sha256sum --check docs/cartulary-ui-ux-refactor-digest/MANIFEST.sha256` | Every entry passes and the manifest path set equals the package path set | Regenerate only after locating the extra, missing, or changed file. |
| `make lint-markdown` | Pass | Repair Markdown only; do not broaden into product formatting. |
| `git diff --check` | Pass | Repair whitespace defects. |
| Final scope audit | Only the digest subtree and this handoff changed | Revert or separately authorize unrelated changes. |

The complete upstream suite must run from the full release checkout because
`test_catalog_refresh.py` and `test_relevance_evaluator.py` depend on maintenance
scripts at the upstream repository root that are intentionally outside the
copied subtree. The bundled subtree remains independently usable for data
validation and offline search. `PYTHONDONTWRITEBYTECODE=1` is required in
addition to `python3 -B` because the suite spawns Python subprocesses.

The upstream validation suite may exercise its own persistence behavior inside
temporary test directories. The implementation session itself must never invoke
`--design-system`, `--persist`, or `--force` against Cartulary.

This is documentation and vendored-advisory maintenance. Do not run product
generation or product tests unless an unexpected non-document change appears.
If that occurs, stop the docs-only slice and obtain separate authorization
rather than normalizing the expanded scope through additional tests.

## 8. Work tracker

Status values are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and
`DROPPED`.

| ID | Work item | Status | Depends on | Evidence required | Exit condition |
| --- | --- | --- | --- | --- | --- |
| DU-001 | Capture clean execution baseline and allowed paths | DONE | none | Branch, commit, status, scope record | One non-self-referential localization snapshot is frozen. |
| DU-002 | Revalidate full repository and owner map | DONE | DU-001 | Reviewed `REPO_MAP.tsv`, owner/status checks | Every row is current or explicitly corrected. |
| DU-003 | Export and replace exact `v2.15.0` source tree | DONE | DU-001 | Tag/SHA proof, file count, recursive comparison | Bundle is an exact 70-file source snapshot. |
| DU-004 | Refresh license and source provenance | DONE | DU-003 | License hash and `meta/source.json` review | License, source metadata, and bundle agree. |
| DU-005 | Re-localize README, prompts, and maps | DONE | DU-002, DU-004 | Overlay diff and source/authority review | All current paths, owners, and boundaries are accurate. |
| DU-006 | Reclassify material upstream advice | DONE | DU-003, DU-005 | Query outputs and R001+ ledger review | Every material result is ADOPT, ADAPT, or REJECT. |
| DU-007 | Refresh A001-A027 regression criteria | DONE | DU-002, DU-005 | Acceptance diff and owner map | Criteria reflect current baselines without new authority. |
| DU-008 | Refresh localization metadata | DONE | DU-002, DU-005, DU-006, DU-007 | Parsed metadata and snapshot review | Metadata describes the actual execution state. |
| DU-009 | Regenerate and reconcile package manifest | DONE | DU-003, DU-004, DU-005, DU-006, DU-007, DU-008 | Exact path-set and checksum comparison | Expected package files are accounted for exactly once. |
| DU-010 | Run final validation and scope audit | DONE | DU-009 | Complete validation matrix with outputs | Every required check passes and no product path changed. |
| DU-011 | Complete execution handoff | DONE | DU-010 | Filled log in §10 | Compatibility, rollback, results, and next step are recorded. |

At most one work item is `IN_PROGRESS`. A blocked prerequisite blocks its
dependents; it is not permission to skip validation or weaken source integrity.

## 9. Risks and failure policy

| Risk | Required control |
| --- | --- |
| Large upstream delta obscures local edits | Replace and compare the entire source tree; prohibit local changes below `upstream/`. |
| Default branch advances after planning | Pin the release SHA, not moving HEAD. |
| New upstream design-generation guidance is followed accidentally | Keep Cartulary query recipes first in the read order and retain explicit option prohibitions. |
| Generic advice becomes product authority | Require owner mapping and ADOPT/ADAPT/REJECT classification for every material result. |
| Localization records implementation details as domain terms | Apply `domain.md` vocabulary boundaries and point to owner projections instead of copying them. |
| Design guidance creates Core behavior | Apply `design.md` only within its observable design scope. |
| Remediation handoff is mistaken for authority | Cite it only as current-state and verification evidence. |
| Manifest conceals stale or extra files | Compare both checksums and the exact manifest/package path sets. |
| Python caches contaminate the package | Use `PYTHONDONTWRITEBYTECODE=1` with `python3 -B` and reject unexpected cache paths before manifest generation. |
| Repository moves before execution | Re-run the full localization scan and record the actual baseline. |
| Docs-only scope expands into product changes | Stop and request a separately authorized product slice. |

No compatibility migration is required. The package is documentation and
vendored advisory material. Its compatibility boundary is provenance and read
order: all overlay references, source metadata, and checksums must move together.

## 10. Execution handoff log

Append one row per execution checkpoint. Do not rewrite planning evidence as if
it were execution evidence.

| Date/time | Actor | Work items | Baseline/source snapshot | Paths changed | Commands and results | Advisory disposition | Blockers/deferrals | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-30 18:13 EDT | Codex | DU-001 | `main` at `2356949f7ec3c8e27ff83ae695d60e06a387d0e5`; source `v2.15.0` at `a38d04c3d5c298c851dbe5e6ee1965ee3de42cb5` | This handoff | Clean tracked baseline; Git 2.53.0; Python 3.14.4; immutable tag and source delta verified in `/tmp/cartulary-uiux-exec.Pa9xPP` | Corrected source, cache, manifest, and full-checkout test facts; no advice classified | None | DU-002 |
| 2026-08-30 18:16 EDT | Codex | DU-002 | Execution baseline above | This handoff only; overlay correction ledger recorded | Revalidated all 74 repository-map data rows, owner status, generated roots, public targets, package presence, and direct vendor boundary | No advice classified; found no owner contradiction | Corrections required: absent `packages/ui`, generated view-contract root, Report Composition 1.2.0, and current remediation baselines | DU-003 |
| 2026-08-30 18:16 EDT | Codex | DU-003 | `v2.15.0` at `a38d04c3d5c298c851dbe5e6ee1965ee3de42cb5` | Complete `upstream/ui-ux-pro-max` subtree | Clean archive replacement; 70 files; zero symlinks/caches; no-index path, byte, and mode comparison passed | Source only; no Cartulary advice classified | None | DU-004 |
| 2026-08-30 18:18 EDT | Codex | DU-004 | Release and execution baseline above | License verified; `meta/source.json`, README upstream section, and `UPSTREAM_MAP.md` | Source JSON assertion passed; license SHA-256 `738f69dfa83db5c347c678fb9d90e560877059f0de93a327c39001bff92dc014`; upstream tree remained exact | Expanded provenance/test-topology descriptions remain advisory | None | DU-005 |
| 2026-08-30 18:20 EDT | Codex | DU-005 | Execution baseline plus current owners and implementation evidence | README, `START_HERE.md`, `LOCAL_AGENT_PROMPT.md`, `REPO_MAP.tsv`, `OWNER_MAP.tsv`, `QUERY_RECIPES.md`, `UPSTREAM_MAP.md` | Both TSV shapes passed; all 76 repository-map rows resolve or are explicit `NOT PRESENT`; stale assertion scan passed | Completed create, inspector, density, responsive, and visual work is now a regression baseline | None | DU-006 |
| 2026-08-30 18:21 EDT | Codex | DU-006 | Refreshed `v2.15.0` search/data | `rules.tsv` | Eight narrow UX/icon/React queries reviewed; 35 rule rows have valid shape and unique IDs; R001-R034 retained | Amended R002/R006/R023; added R035 `ADAPT` for text-layout resilience; all other results mapped to existing rows; cybersecurity defaults remain R026-R028 `REJECT` | None | DU-007 |
| 2026-08-30 18:22 EDT | Codex | DU-007 | Current adopted owners and completed remediation evidence | `acceptance.tsv` | 27 rows have valid shape and unique stable A001-A027 IDs | Strengthened A006/A007/A008/A010/A024/A025 without creating product authority; retained remaining regression gates | None | DU-008 |
| 2026-08-30 18:24 EDT | Codex | DU-008 | `main` at `2356949f7ec3c8e27ff83ae695d60e06a387d0e5`; `v2.15.0` at `a38d04c3d5c298c851dbe5e6ee1965ee3de42cb5` | `meta/localization.json` | JSON parse and cross-metadata snapshot assertions passed; current stack/workspace facts revalidated | Added refresh scope, four current qualifications, resolved mappings, and exact upstream pin; records no product changes/tests | None | DU-009 |
| 2026-08-30 18:26 EDT | Codex | DU-009 | Stable localized package above | `MANIFEST.sha256` | 82/82 checksums passed; manifest and package path sets match with zero differences; caches/bytecode excluded | No advisory change | None | DU-010 |
| 2026-08-30 18:28 EDT | Codex | DU-010 | Final package candidate on execution baseline | Digest subtree and this handoff only | Exact tree/license passed; data validation passed; upstream 153/153; queries 3/3 and 8/8; JSON/TSV/IDs/manifest/path set passed; `agent-finalize` 1/1 at `20260830T222748Z-p3671118`; Markdown lint passed at `20260830T222802Z-p3674035`; whitespace and scope passed | All classifications retained; no product tests or generation run | `RESULTS_DIR` unset because no full warm `make check` run applies to this docs-only slice | DU-011 |
| 2026-08-30 18:29 EDT | Codex | DU-011 | `main` at clean baseline `2356949f7ec3c8e27ff83ae695d60e06a387d0e5`; final dirty scope is the digest subtree and this handoff; source `v2.15.0` at `a38d04c3d5c298c851dbe5e6ee1965ee3de42cb5` | `MANIFEST.sha256`; `README.md`; `cartulary/{LOCAL_AGENT_PROMPT.md,OWNER_MAP.tsv,QUERY_RECIPES.md,REPO_MAP.tsv,START_HERE.md,UPSTREAM_MAP.md,acceptance.tsv,rules.tsv}`; `meta/{localization.json,source.json}`; `upstream/ui-ux-pro-max/**`; this handoff | Exact 70-file source tree and license `738f69dfa83db5c347c678fb9d90e560877059f0de93a327c39001bff92dc014`; all DU-010 checks passed; post-handoff Markdown lint passed at `20260830T222955Z-p3675647`; final whitespace, integrity, tracker, acceptance, and scope checks passed | R001-R034 retained; R002/R006/R023 amended; R035 added as `ADAPT`; R026-R028 remain `REJECT`; A001-A027 refreshed as a future regression contract, not a new product claim | No runtime/data migration; rollback source/provenance and the complete overlay as coherent checkpoints; product tests/generation skipped by docs-only scope; retained-run maintenance skipped because `RESULTS_DIR` was unset | none |

The terminal log row must include:

- actual branch, baseline commit, and final dirty-state scope;
- upstream tag/SHA, exact-tree result, file count, and license hash;
- every changed localized path;
- validation commands and retained outputs;
- material ADOPT/ADAPT/REJECT changes;
- A001-A027 disposition;
- compatibility and rollback statement;
- skipped checks with reasons; and
- the next separately authorized slice, or `none`.

## 11. Binary acceptance criteria

- [x] The actual execution baseline is clean, current, and recorded.
- [x] Release `v2.15.0` resolves to the required immutable commit.
- [x] The 70-file upstream destination tree exactly matches the release source.
- [x] The MIT license hash matches the expected value.
- [x] Source and localization metadata describe the same upstream and Cartulary
      snapshots as the package contents.
- [x] Every `REPO_MAP.tsv` row is revalidated against the current repository.
- [x] The known drift in §5 is resolved without inventing a package, owner, or
      product behavior.
- [x] Resolved remediation gaps are represented as regression baselines.
- [x] R001-R034 retain stable IDs and every materially new recommendation has a
      classified appended row or an explicit existing-row mapping.
- [x] A001-A027 retain stable IDs and reflect current create, inspector,
      density, responsive, generated-root, and test-authority behavior.
- [x] No upstream generation/persistence option was invoked against Cartulary.
- [x] Upstream validation, 153 tests, and both sample queries pass.
- [x] JSON, TSV, Markdown, whitespace, manifest, and path-set checks pass.
- [x] No product, owner-specification, generated-product, dependency, or
      lockfile path changed.
- [x] The execution log contains results, compatibility, rollback, deferrals,
      and the next action.

Every applicable criterion is checked and DU-001 through DU-011 are `DONE`.
