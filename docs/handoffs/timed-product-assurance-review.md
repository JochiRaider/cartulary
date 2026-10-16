# Timed product assurance review

Human review only. Core 04 REQ-04-157–159 owns fixtures, action acceptance, visible two-frame completion and sampling. Core 05 separately owns qualification for published benchmark claims. A browser title, eventual assertion, response timer or test timeout does not satisfy these predicates. Review: Codex under the approved remediation plan; statuses below are deliberately independent of catalog reachability.

| Criterion | Predicate and exact threshold | Fixture / sampling | Current exercising evidence and gap | Closure owner |
| --- | --- | --- | --- | --- |
| AC-003 | `perf.timeline_paste_20x5.v2`; elapsed <2,000 ms | Large-grid; exact 20×5 TSV; one operation, no warm-up/retry; fixed 1920×1080/DSF1 | New `timeline-paste.spec.ts` title `measures 20 new Timeline rows and 100 visible TSV cells within AC-003`; exact catalog row `module.timeline.measurement.paste_20x5_v2`. Product acceptance mark and 100-cell/20-new-ID two-frame detector implemented. Negative detector and evidence tests pass; retained browser run `remediation-timeline-measurements-20261014-01` passed with one 356.3 ms observation and qualified v4 aggregate/v5 target evidence | Timeline / harness browser |
| AC-008, AC-132 | Three `perf.timeline_presence_{surface,row,cell}.rendered.v1` predicates; matching remote indicator within 1,000 ms | Large-grid, 25 sessions; single observation | `timeline-presence.spec.ts` measures coordinator acceptance through two-frame remote glyph completion in one browser clock domain. `remediation-presence-browser-20261014-04` passed all three rows and aggregate qualification: surface 28.9 ms, row 178.1 ms, cell 180.1 ms. Negative DOM/clock, instrumentation and retained-evidence regressions pass. The older two-client `collaboration.spec.ts` case remains functional support, not this workload's assurance | Collaboration / Workbook |
| AC-011 | Three `perf.timeline_{history_entry_rollback,change_set_rollback,row_restore}.rendered.v2` predicates; elapsed <2,000 ms | Large-grid, ordinary reviewer, 25 sessions; independent single observations | `timeline-history.spec.ts` implements exact single-field, two-field change-set, and whole-row restore workloads. Existing legal actions, admission and reconciliation remain the product path. The observer requires restored cells plus new visible attribution together on two frames, with fixed grid scroll and geometry; product-managed inspector focus reveal remains in the measured interval. Owner, presentation, detector and evidence regressions pass. Qualified v2 runs: `remediation-history-v2-browser-20261014-01` change-set 348 ms; `remediation-history-v2-remaining-20261014-01` single-entry 241.3 ms and whole-row 231.7 ms. The generic predicate and the three superseded history v1 identities are retired | Revisions / Workbook |
| AC-016, AC-046 | `perf.evidence_verification.visible_with_cancel.v1`; progress/cancel and distinct-row editing within 1,000 ms | Evidence-heavy v2; full 100-preview preflight, one 2 MiB file, one editor, single observation | `evidence-verification.spec.ts` exercises product acceptance after transfer, one fresh verification admission, exact Job/blob bindings and two visible frames with an uncommitted draft on a different row. Detector and retained-upload tests pass. Physical root `remediation-evidence-measurement-browser-20261016-04` passed with one qualified 125.2 ms observation, zero overlap and all 28 cleanup operations complete; held-response browser/accessibility tests remain separate functional evidence | Evidence / Workbook |
| AC-027 | `perf.import_csv_discovery.visible_with_cancel.v1` and `perf.import_xlsx_discovery.visible_with_cancel.v1`; <=1,000 ms | Large-grid, exact 20,000-row two-column files, single observation per format | Core 04 now pins file content, ordinary admission, geometry, concurrent grid editing, exact Job/cancel two-frame completion and no retries. Existing `import-assistant.spec.ts` paths remain functional support. Instrumentation and detector/evidence validation are implemented. `remediation-import-discovery-browser-20261015-04` qualified CSV at 609.6 ms and XLSX at 462.2 ms, one fresh observation each; discovery does not establish apply timing. Separate apply predicates qualified in `remediation-import-apply-repaired-browser-20261016-01`: CSV 343.4 ms, XLSX 302.7 ms, with successful ordinary cancellation and all cleanup complete | Tabular Ingest / Workbook |
| AC-030 | `perf.reporting_{snapshot,render}.visible_with_cancel.v1`; progress/cancel within 1,000 ms | Snapshot/report workload, single observation per distinct admission path | Exact large-grid snapshot and report predicates qualify in `remediation-reporting-complete-owner-20261016-01`: 173.7 ms and 179.1 ms, one fresh observation each, ordinary cancellation and all 41 cleanup operations complete. The current-source Reporting owner audit passed; see the render recipe qualification record | Reporting / Composition / Workbook |
| AC-033 | Three `perf.reference_pack_{import,reverify,refresh}.visible_with_cancel.v1` predicates; <=1,000 ms | Exact signed host-registry bundle and large-grid/25-session fixture; one independent observation per mechanism | Core 04 now defines separate administration and Timeline pages sharing one ordinary session, exact prepared pack state, and two frames in both documents. Full owner root `remediation-reference-progress-owner-20261016-01` qualified import 84.1 ms, reverify 109.9 ms and refresh 150.4 ms, one observation each with zero overlap. Explicit full-owner evidence audit passed. Existing landing responsiveness remains functional support only | Reference Data / Workbook |
| AC-043 | `perf.timeline_summary_selection_down.v1`; p95 <=100 ms | Large-grid, 25 sessions; one discarded warm-up plus 100 operations | `timeline-grid.spec.ts`: `measures paint-qualified Timeline summary ArrowDown selection within AC-043`. Product mark/paint observer and typed policy exist. Qualified run `remediation-timeline-measurements-20261014-01`: 100 measured samples, p95 29.0 ms | Timeline |
| AC-043 | `perf.timeline_summary_focus_edit.v1`; p95 <=100 ms | Same large-grid and p95 protocol | `timeline-grid.spec.ts`: `measures paint-qualified Timeline summary pointer focus within AC-043`. Anchored same-record/field editor and caret predicate; qualified run `remediation-timeline-measurements-20261014-01`: 100 measured samples, p95 34.9 ms | Timeline |
| AC-043 | `perf.typing_ack.v1`; p95 <=100 ms | Same large-grid and p95 protocol | `timeline-grid.spec.ts`: `measures paint-qualified committed Timeline summary typing acknowledgment within AC-043`. Visible trailing `x` in same editor; request completion never substitutes; qualified run `remediation-timeline-measurements-20261014-01`: 100 measured samples, p95 25.8 ms | Timeline |
| AC-043 | `perf.timeline_blank_row_create.v1`; p95 <=150 ms | Same large-grid and p95 protocol | `timeline-grid.spec.ts`: `measures paint-qualified Timeline blank-row creation within AC-043`. New committed row/value/version visible across two frames; qualified run `remediation-timeline-measurements-20261014-01`: 100 measured samples, p95 131.1 ms | Timeline |
| AC-044 | Six exact `perf.timeline_{sort,filter,group}.{first_useful_viewport,stable_viewport}.v1` identities; <=250 / <=1,000 ms p95 respectively | Large-grid; independent one warm-up plus 100 operations; exact Core 04 Timeline workloads at 1920×1080/DSF1 | New `timeline-view-change.spec.ts` has six exact measurement rows. Acceptance marks, committed query revisions and two-frame viewport observers are implemented. Filter/first-useful passed `remediation-view-change-filter-20261014-03` with 100 samples, p95 151.4 ms and all 24 remote analysts visible. The other five predicates passed `remediation-view-change-remaining-20261014-01`: filter/stable 220.7 ms, group/first-useful 150.4 ms, group/stable 165.4 ms, sort/first-useful 161.8 ms and sort/stable 194.5 ms p95, each with 100 measured samples and qualified evidence. Generic `perf.view_change.*.v1` identities are retired. Single-group Timeline evidence does not close other surface or multiple-group obligations | View Query / Projections / Workbook |
| AC-045 | `perf.evidence_inspector.metadata_shell.v2`; p95 <=300 ms | Evidence-heavy v2: 10,000 physical 2 MiB PNG objects, 5,000 Timeline rows and a selected 100-link target; one warm-up plus 100 operations | Core 01 REQ-01-328a and Core 03 REQ-03-248a now own the bounded metadata read and five-record inspector window. Current-source backend, adapter, lifetime, presentation and real-browser paging tests pass (see the remediation handoff). Metadata eligibility is not handle availability. Core 04 now defines v2 eligibility timing separately from required real preview-access preflight. The full physical snapshot and exact observer are implemented. Run `remediation-ac045-measurement-v2-20261014-06` qualified one warm-up plus 100 measured opens: p50 45.1 ms, p95 57.1 ms, zero scheduler overlap, all 27 cleanup steps complete. See the full-workload record below for source binding; Core 05 publication remains separate | Evidence / Workbook |
| AC-046 | Exact owner predicates in Core 04 REQ-04-158; within 1,000 ms | Each admitted interactive import, processing, snapshot, report and Reference Pack mechanism | Reuses the mechanism-specific evidence obligations above; bulk projection rebuild is explicitly maintenance-owned under Core 01 REQ-01-353 and AC-472, with import rebuild inside the existing import Job. One umbrella verification ID cannot establish all interactive contexts | Each job owner / Workbook |
| AC-047 | `perf.anchor_stability_under_live_updates.v2`; seeded pass/fail trace, no timing threshold | Large-grid, 25 sessions, exact 11-step trace; no action retry | `timeline-anchor-trace.spec.ts`, row `module.timeline.measurement.ac047_anchor_trace`. Qualified run `remediation-anchor-trace-browser-20261014-04` completed all 11 steps over 154 observed frames; target versions 2→3→4; zero violations; at most 42 mounted rows; zero scheduler overlap and all 27 cleanup operations completed. Existing small-fixture scenarios remain functional support. | Timeline / Grid Adapter / harness browser |

## Common measurement boundary

Preparation qualifies fixture counts, seed, ordinary security, geometry, runtime profile and traffic before the measured action. Product instrumentation marks acceptance once. Browser observers verify the owner-defined visible state across two consecutive animation frames. Evidence serialization records samples and predicate proof without raw source payloads. The harness validates current schemas, exact profile bindings, recomputed statistics, no retries, resource isolation, artifact digests and completed cleanup before qualification. Repeated or contaminated actions remain diagnostic evidence and cannot become a pass through summary editing.

Single-operation evidence uses its own elapsed statistic. P95 evidence retains the existing nearest-rank estimator and exactly 100 measured operations. Changed predicates get new identities; unrelated unchanged predicates keep theirs. Old evidence generations are retained for investigation and rejected by current qualification.

## Remaining measurement closure

Core 04 now pins CSV/XLSX discovery and apply, Reporting snapshot and render,
and Evidence byte-verification workloads. Reporting's workbook workflow and
both progress predicates have been implemented and qualified; the retained
Reporting windows below identify their source applicability. Evidence uses the
base Common Job runtime, with a non-modal progress and cancellation surface;
its exact single-observation workload qualified in the physical Evidence root
recorded above. Later source applicability remains a separate review.

Reference Pack import, reverify and refresh use their deployment administration
route. Their exact grid-context workloads and cross-page evidence qualified in
`remediation-reference-progress-owner-20261016-01`, as recorded in the final
section below. Core 01 REQ-01-353/452 and Core 04 AC-046 now explicitly reconcile bulk
projection rebuild as internal maintenance under Recovery or the existing import
Job, without a standalone browser route or Common Job. Recovery progress,
cancellation and fail-closed readiness remain required under AC-472 and the
operator contract; this is a normative scope decision, not a silent exclusion.
The Graph Projection v2 owner independently forbids a public Graph route or
hidden worker; source-owned extension graph actions retain their own contracts.

## Retained measurement window

`remediation-timeline-measurements-20261014-01` completed through `make test-slice OWNER=module.timeline` with the five exact measurement rows above, outer exit 0, and wall time 324.841 s. Source digest: `sha256:566e1bdbd6f7b5ecd6170732671cf97457b1e483f6d2debab9aeb30c39e5af7a`; graph digest: `sha256:6554f1c3a71811b3033c9f431c7e72aee0226cf8ca42fe6d054fb177503b9068`. Each summary reports qualified evidence with zero scheduler overlap. All 43 cleanup operations completed. AC-043 retains one discarded warm-up plus 100 measured samples per predicate. AC-003 retains one synthesized TSV paste observation. The source state was dirty; this is predicate qualification, not a clean-source cost portfolio or Core 05 publication claim. Later executable-source changes require evidence applicability review or fresh qualification.

After hardening every retained measurement-artifact read against parent symlinks, aliases and concurrent changes, `remediation-secure-paste-measurement-20261014-01` passed the AC-003 v2 row through the real finalizer in 153.033 s (outer exit 0). Its single observation was 335.2000000476837 ms, qualification `qualified`, scheduler overlap zero, and all 27 cleanup operations completed. Source: `sha256:a76003e360c36ca18ede67854715861318d02ccb563310bf7a9ecb5dd5793e88`; graph: `sha256:9734d47fa32dacf731d7d7498b06232cdff786b3f6a32171479c8ca5891cc0e9`. This is a fresh implementation-predicate observation; it is not a clean-source cost window or Core 05 publication claim.

## AC-044 completed Timeline window

The six exact Timeline predicates passed at source
`sha256:d483562d4649036fcb1fc1a2a0b7e85427691bb0a5fae103ccdb7a764ec521e8`.
Filter/first-useful used `remediation-view-change-filter-20261014-03`
(graph `sha256:9d3325c37fa5f56777d421607313acb0b00ca36b581268088b592fdd390ed46e`).
The other five used `remediation-view-change-remaining-20261014-01`
(graph `sha256:c9a839e0c0c1da2f2cd09a0b92fe4c534889387bd0fc0be7d6670fec3cab6acc`),
which passed in 713.241 s with outer exit 0 and all 43 cleanup operations
completed. Each row has one discarded warm-up, 100 measured observations,
zero scheduler overlap, and `qualified` evidence. The presence count is checked
before and after every observation. These are implementation measurements from
a dirty source tree, not Core 05 publication or a clean execution-cost portfolio.
The exact nonselective filter and single-group workload limitations remain.

## Presence clock and evidence boundary

The three presence predicates each retain one operation, without warm-up or
retry. A and B use distinct authenticated contexts in the same Chromium
execution. B is one of the 24 updating background sessions. Product acceptance
marks precede presence publication debounce. The remote observer requires the
exact actor glyph and, for row/cell cases, record, positive version and field
bindings across two frames. Surface presence must show the actor individually;
a total or overflow count cannot qualify. Absolute timestamps and display names
are excluded from retained proof. The clock conversion follows
[W3C High Resolution Time](https://www.w3.org/TR/hr-time-3/#time-origin): each
context's origin and relative time describe the same browser monotonic clock.
The versioned presence proof is an additional explicit variant in measurement
observation v3; existing paste and percentile proof semantics are unchanged.

Generation `remediation-presence-generate-20261014-01` passed in 10.190 s;
format `remediation-presence-format-20261014-01` in 4.741 s; paint qualification
`remediation-presence-paint-20261014-01` in 2.547 s; and the coordinator row
`remediation-presence-coordinator-20261014-01` in 3.883 s. All outer exits were 0.
These checks establish detector and instrumentation behavior, not browser latency.

Presence type checking `remediation-presence-types-20261014-01` passed in
20.134 s, and the focused graph contract root
`remediation-presence-graph-20261014-01` passed (outer exit 0). Finalization
`remediation-presence-finalize-20261014-01` passed in 18.625 s; retained-run
maintenance was skipped because `RESULTS_DIR` was unset.

The first browser root `remediation-presence-browser-20261014-01` failed in
213.063 s (outer exit 2, `product` / `test_assertion_failure`), with all 35
cleanup operations completed. Source was
`sha256:88be8c0cfab05e359bdb133c86a441e1ac64750b57701024a133a03db090728e`.
Surface preparation could not select Hosts using the overflow menu; preparation
now uses the ordinary visible Hosts tab and measures the visible Timeline tab.
Row and cell observers timed out without a qualifying two-frame state. Their
rejection now reports bounded binding/visibility reasons; boxless CSS
`display: contents` ancestors are excluded from clipping geometry, with an
explicit detector regression. No latency or product-stability conclusion is
inferred from that failed run.

The second browser root `remediation-presence-browser-20261014-02` failed in
215.268 s (outer exit 2), with all 34 cleanup operations completed. Cell proof
reported a nonintersecting target; preparation now reveals the summary field in
both browsers before timing. Row proof found no matching focused target;
preparation now waits specifically for focused rather than any row-presence mode
and failures retain bounded semantic diagnostics. Surface browser assertions
passed, but evidence collection rejected the newly admitted nullable synthetic
`record_id`. The collector now permits that key only at the exact closed
presence-proof path; its end-to-end collection test covers all three variants
and rejects additional payload members. This run remains failed and supplies no
qualified presence acceptance claim.

The collection regression root `remediation-presence-graph-20261014-02`
passed in 25.710 s, and type checking `remediation-presence-types-20261014-02`
passed in 20.687 s (outer exits 0). Markdown lint
`remediation-presence-markdown-20261014-01` passed in 21.240 s. Finalization
`remediation-presence-finalize-20261014-03` rejected stale generated topology
inputs after the evidence collector changed (3.017 s, outer exit 2); no
measurement was launched from that state. `make generate` in
`remediation-presence-generate-20261014-02` then passed in 9.260 s.

The third root `remediation-presence-browser-20261014-03` failed in 217.069 s
(outer exit 2), with all 35 cleanup operations completed. Its surface row has
individually qualified evidence: 31.20019531296566 ms, one observation, zero
scheduler overlap. Source:
`sha256:0dd1d17a5c0ba948a6bf7d995ec30587f856c63e4007af08d8c1711d6cc15a61`.
The overall run is failed. Row diagnostics proved local focus reached the
correct record while no presence acceptance was emitted. Timeline's active-cell
callback updated only its focus anchor. It now also publishes accepted focus
through its presence owner; stale editor-stop notifications cannot overwrite a
different record's current state. This repairs existing Core 03 REQ-03-092,
without changing debounce or wire semantics. The focused regression root
`remediation-presence-focus-unit-20261014-01` passed in 2.259 s (outer exit 0).
Cell publication and remote identity binding succeeded, but clipping
qualification still failed; exact bounded geometry and an observer failure
capture now support the next diagnosis.

The target-cell preparation now qualifies the whole grid cell inside the clipped
viewport before timing; revealing only its text could leave the presence marker
offscreen. Type checking initially caught a missing projection of the new focus
command through Workbook composition (`remediation-presence-types-20261014-03`,
outer exit 2); after that projection was repaired,
`remediation-presence-types-20261014-04` passed in 20.326 s, outer exit 0.

`remediation-presence-targets-20261014-01` exercised row and cell presence after
those fixes. Both browser rows and individual measurement summaries passed, but
the aggregate failed as `artifact` / `artifact_error` (outer exit 2, 176.339 s).
All 31 cleanup operations completed. Source:
`sha256:26433173c354bf40ef39db81c1dc3ef93bfc31e6fadf99c68f3b0261427d554b`;
graph: `sha256:4a9209333bcba33b4e5a957cce53d6ed92796c029e7acc2120cb27998f2f8747`.
Two retained-artifact validation callers still used the old boolean paste flag.
Collection and retained finalization now derive the exact synthetic proof policy
from the same schema-validated observation, with no caller-supplied flag or path.
Regression cases cover paste and all three presence variants through both
collection and retained finalization, including forbidden top-level identities
and unknown payload members. The failed root is retained unchanged.

`remediation-presence-graph-20261014-03` passed in 23.930 s;
`remediation-presence-generate-20261014-03` in 9.200 s;
`remediation-presence-format-20261014-07` in 4.761 s;
`remediation-presence-markdown-20261014-02` in 8.850 s; and
`remediation-presence-finalize-20261014-08` in 18.224 s. All outer exits were 0.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

The complete three-row successor `remediation-presence-browser-20261014-04`
passed in 201.513 s with outer exit 0 and all 35 cleanup operations completed.
Source: `sha256:a52dc95262a0e12f625cdc3213d79754370ed5719c4c946b327502dce92d78e8`;
graph: `sha256:53de69e32f989352799042daccb26654d958d6b2985cde6aae270ece16b63bd4`.
Surface elapsed 28.899951172061265 ms, row 178.10009765671566 ms and cell
180.099853515625 ms. Every predicate has one observation, no warm-up or retry,
zero scheduler overlap and qualified current-generation aggregate evidence.
This is a dirty-source implementation measurement, not a clean cost portfolio or
Core 05 publication claim. Historical failed roots remain failed.

## History measurement implementation

Core 04 now fixes separate Timeline single-entry, whole-change-set and whole-row
restore workloads. The foreground is admitted as an ordinary reviewer through
the existing membership route. One confirmation is timed from the existing
history owner's accepted intent, before coordination/revalidation/dispatch.
The history presentation exposes typed semantic bindings already available to
that authorized view. Its paint observer requires two restored visible cells,
the new exact row version, a new history/change-set identity and visible reviewer
attribution and committed time together on two frames. It rejects retargeting,
stale entries, hidden or clipped content, one-frame completion, repeated action,
document changes, driver scrolling, grid scrolling and geometry changes. Ordinary post-measurement
reads independently check the rendered history identity, timestamp and values.
No navigation or route completion supplies the stop predicate.

`remediation-history-generate-20261014-01` passed in 9.070 s;
`remediation-history-format-20261014-01` in 4.638 s;
`remediation-history-paint-20261014-01` in 2.656 s;
`remediation-history-owner-20261014-01` in 3.074 s; and
`remediation-history-graph-20261014-01` in 23.400 s (all outer exits 0).
The graph regression includes all three proof variants through collection and
retained finalization, strict threshold rejection at 2,000 ms, and incorrect
action/version/attribution/geometry/cardinality cases.

`remediation-history-types-20261014-01` failed (19.405 s, outer exit 2,
`harness` / `child_target_failure`) because the preparation used an unbounded
array where the public patch type requires at least one change. The workload now
constructs an explicit nonempty one- or two-field payload. All 11 cleanup
operations completed. `remediation-history-format-20261014-02` then passed in
4.729 s, `remediation-history-finalize-20261014-01` in 17.921 s, and
`remediation-history-types-20261014-02` in 19.799 s (outer exits 0).
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

`remediation-history-browser-20261014-01` failed in 221.783 s (outer exit 2,
`product` / `test_assertion_failure`, all 35 cleanup operations completed).
All three history predicates timed out before visible completion; this root
supplies no passing AC-011 evidence. Source:
`sha256:fac096039a316f3693e06fc736466b44fbc0cf80ca187f1535f782d710b2b0d0`;
graph: `sha256:23721f5a396b3df810f6318dec9543d13fe693eefb4acfe5c84b30f34e127b56`.
The retained change-set trace shows the ordinary rollback POST returning HTTP
500 after 3,147.097 ms, with the source values and history unchanged. The public
error correctly withholds private diagnostics. The retained evidence does not
identify a PostgreSQL error code, so deadlock attribution remains unconfirmed.
Inspection found that all interactive history mutation paths rebuilt every
incident projection in their transaction, including unrelated large-grid rows.

The corrective slice removes the incident-rebuild capability from Revisions.
Source-owned Links and Entities readers identify direct projection dependents;
Revisions keeps that refresh/protection scope distinct from authoritative
mutation/history identities. The coordinator protects and revalidates the scope
before mutation, then invokes the existing exact row refresh capability. It
retains inactive relationship endpoints for restore and avoids transitive graph
expansion. Maintenance/import/recovery keep their own explicit rebuild ports.
Core 01 REQ-01-351 and REQ-01-104 now state this scope and lock treatment; the
expanded rollback/restore protected set is an intentional compatibility change
that needs human projection-fidelity review. Ordinary soft-delete retains
ordinary transaction locks. No whole-incident compatibility fallback remains.

`remediation-history-scoped-revisions-20261014-01` failed in 20.643 s (outer exit
2, `product` / `test_assertion_failure`, all 18 cleanup operations completed).
Construction, delete/restore adapter and atomic-consequence checks, existing
rollback selectors, lock precedence, and injected failure atomicity passed.
New trigger-backed assertions proved that single-entry and change-set rollback
do not write an unrelated projection whose envelope remains locked. The new
whole-row assertion selected an unseeded revision 1 and received the expected
404. Its preparation now makes an ordinary patch after retained revision 2 and
restores that actual prior revision. A further public-route regression checks
Evidence-title propagation to Timeline, dependent lock precedence, soft-delete
removal and restore via the retained inactive-endpoint relation.

`remediation-history-scoped-revisions-20261014-02` failed in 19.095 s (outer exit
2). All three unrelated-projection assertions and the new dependency scenario
passed; the existing outer scenario subsequently timed out waiting for its
fixed-clock WebSocket event. The added scenario had shared that server's clock
and pending publication state. An attempted nested independent server then
blocked on the package-scoped object-store fixture mutex: the parent retains
that mutex until cleanup. `remediation-history-scoped-revisions-20261014-03`
was therefore intentionally interrupted through its verified native execution
handle after inspecting the fixture ownership (outer exit 130, 120.207 s,
`interrupted` / `cancelled_or_interrupted`). Neither root is passing evidence.
The scenario is now an independently cataloged top-level test,
`TestHistoryProjectionDependencies_Integration`, under
`module.revisions.integration.projection_dependencies`, so it acquires and
releases its own fixture before another top-level scenario begins.

Generation `remediation-history-scoped-generate-20261014-01` passed in 9.266 s;
`remediation-history-scoped-finalize-20261014-01` passed in 18.130 s (outer exits
0). Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

Both scoped Revisions runs 02 and 03 completed all 14 cleanup operations.
`remediation-history-scoped-revisions-20261014-04` selected all 46 current Go
rows and finished in 38.567 s with outer exit 2. Its product tests, including the
independently isolated dependency scenario, passed. Three assembly rows lacked
selector results because their test double still implemented the retired
incident-rebuild method; the harness classified the build failure as `infra` /
`preflight_error`. That double now implements the exact row-refresh capability.
`remediation-history-scoped-projections-20261014-01` passed all 22 owner rows
(19 work units) in 29.851 s with outer exit 0, covering source ownership, public
facades, projection/query parity, deterministic rebuild and recovery behavior.
These concurrent backend runs are verification evidence, not timing windows.


## History observer correction and v2 qualification

The missing three assembly rows passed through
`remediation-history-scoped-assembly-20261014-01` in 1.216 s; the history-owner
regression passed through `remediation-history-scoped-owner-20261014-01` in
3.778 s, and Markdown lint through
`remediation-history-scoped-markdown-20261014-01` in 21.290 s (outer exits 0).
Revisions run 04 and the Projections run completed all 33 and 48 cleanup
operations respectively.

`remediation-history-browser-20261014-02` failed in 212.689 s (outer exit 2).
Its three observers rejected inspector scrolling after confirmation. The
change-set request now returned HTTP 200 in 53.977 ms; this request duration
alone is not the visible performance predicate. The retained trace shows grid
horizontal scroll unchanged at 69 CSS pixels and inspector-body scroll changing
from 2,419 to 2,119 as confirmation closes and ordinary focus is restored.
The inspector focus behavior is owned by the existing design contract. The
initial detector incorrectly froze every scroll container, including the
inspector, rather than preventing driver scrolling from manufacturing grid
completion. No product focus behavior was changed to make this test pass.
Source: `sha256:f007ec7314e228f001b4b7213e91ce76869e155eb11223c000a1560e8fe81955`;
graph: `sha256:23721f5a396b3df810f6318dec9543d13fe693eefb4acfe5c84b30f34e127b56`.

Core 04 and the closed history v2 contract explicitly distinguish fixed grid
scroll/geometry from product-managed inspector focus reveal. Preparation focuses
the real confirm control; the driver then sends one ordinary Enter without
click auto-scrolling. All restored cells and the new visible history attribution
must still qualify together across two frames. Inspector movement that hides
required content fails. New predicate, contract and proof identities retire v1;
no compatibility reader translates old evidence. Negative tests reject old
identities, driver-managed scrolling, changed grid scroll and non-keyboard
action metadata, in addition to the existing threshold and binding cases.

`remediation-history-v2-generate-20261014-01` passed in 8.990 s;
format 01 in 4.775 s; paint qualification 01 in 2.497 s; finalization 01 in
17.993 s; and TypeScript checking 01 in 23.310 s (outer exits 0).
`remediation-history-v2-harness-20261014-01` failed in 44.024 s (outer exit 2,
`product` / `test_assertion_failure`): seven units passed, including routing
contracts and the v2 proof regressions. Owner-boundary checks found three imports
of private catalog files; callers now use the existing public catalog facade.
The cleanup mount-boundary negative inherited global machine-output mode and
was rejected before its intended safety assertion. That subprocess now uses the
same deliberately bounded fixture environment as the other cleanup cases.
Neither fix weakens the invariant being tested.

After the Windows interruption, exact-root observation recovered finalization
02 as failed in 2.897 s with all 10 cleanup operations completed. Its owning
outer process exit was unavailable. The retained finalizer diagnosis identifies
stale generated topology inputs after the public facade changed; no browser
measurement ran from that state. Generation 02 subsequently passed in 9.140 s,
outer exit 0. Retained-run maintenance remains skipped because `RESULTS_DIR`
is unset. These roots remain implementation evidence rather than a clean-source
cost window or a Core 05 publication claim.


Finalization 03 passed in 18.066 s (outer exit 0). The two-row boundary run
`remediation-history-v2-boundaries-20261014-01` failed in 41.848 s (outer exit
2): cleanup/public-registry parity and private-import checks passed, while a
separate fixture-vector assertion exposed registration-order-dependent generated
output. The renderer now orders profile descriptors and key vectors at its
boundary. The existing assertion is retained; a new reversed-registration
comparison requires byte-identical output. This preserves deterministic
extension to additional profiles. Generation 03 passed in 9.210 s, outer exit 0.


Boundary runs 02 and 03 failed in 17.994 s and 18.013 s (outer exits 2): a
negative source-identity fixture violated contract-ID order, then path order,
before reaching the intended mismatch check. Its final mutation changes only
the existing contract's version token, preserving both collections' structural
preconditions. Fixtures now locate their profile by identity instead of list
position. `remediation-history-v2-boundaries-20261014-04` passed in 17.858 s,
outer exit 0. No invalid input was admitted and no expected rejection weakened.
Finalization 04 passed in 18.123 s, outer exit 0; maintenance was skipped because
`RESULTS_DIR` was unset.


## AC-011 qualified history v2 window

Finalization 05 passed in 18.242 s (outer exit 0; retained-run maintenance
skipped because `RESULTS_DIR` was unset). Exact change-set selection
`remediation-history-v2-browser-20261014-01` passed in 148.532 s, outer exit 0,
all 14 work units and 27 cleanup operations complete. Its single observation is
348 ms, with `qualified` evidence and zero scheduler overlap. Graph:
`sha256:17536ce26f05351e0f9542d42555bf8128c223e1bee5f0a9e2d11a825445de36`.

Exact single-entry and row-restore selections
`remediation-history-v2-remaining-20261014-01` passed in 173.983 s, outer exit 0,
all 16 work units and 31 cleanup operations complete. Single-entry elapsed is
241.2999999988824 ms; row restore elapsed is 231.69999999925494 ms. Each has
one observation, zero warm-ups, no retries, zero scheduler overlap and
`qualified` evidence. Graph:
`sha256:a23d471f97a23b46d889de869c4e75e89f38c69f5d98c813c52cca98a5dfe238`.
Both invocations use source
`sha256:653a49094836e1ba02af76207c7c6b66e9c33ebed9d951ed2416d0399b2a4763`.
Each action completes below the unchanged strict 2,000 ms threshold, including
product-managed inspector focus reveal and visible new history attribution.
This closes these three predicates for the retained implementation source; it
is neither a clean-source execution-cost portfolio nor Core 05 publication.

### Later projection-owner boundary repair

`remediation-reporting-payload-boundary-20261014-01` found a direct `records`
read in the newly added Entities projection-dependency reader. It failed
`product/test_assertion_failure` (1.989 s, outer exit 2), with all 12 cleanup
operations complete. The reader now queries only Entities' mention relation;
globally unique resolved record IDs select direct dependents. Revisions already
validates every returned envelope's incident and source owner before locking or
mutating the scope. No Records access exemption was added.

`TestProjectionScopeIsDirectAndIncidentBound` adds explicit cross-incident,
missing-envelope, unsupported-owner, nil-identity, source-reader-failure and
empty-mutation denial. Valid duplicate contributions close to a sorted direct
scope; readers receive only the original changed IDs, without recursive
expansion. That unit and the existing dependent rollback/delete/restore
integration passed in
`remediation-reporting-scope-boundary-revisions-20261014-01` (16.310 s,
outer exit 0). The earlier browser numbers remain bound to their recorded
source; this later query change is not silently included in those windows.

### Measurement guard validation after restart

The Reporting follow-up lint run
`remediation-reporting-independent-binding-lint-20261014-01` failed
(33.027 s, outer exit 2, `harness/child_target_failure`): an unused
Composition default and 53 non-null assertions in the measurement helpers and
their tests. The default was removed. Measurement preparation and observation
now explicitly reject missing predicates, fixture records, DOM nodes, timing
marks and observer pages. Fixed-size DOM fixtures use checked tuples; no lint
suppressions were added. These guards do not alter workload thresholds.

Focused `make test-slice` checks completed with outer exit 0:

| Exact run ID | Selected owner scope | Duration |
| --- | --- | --- |
| `remediation-measurement-guards-history-20261014-01` | `module.revisions.frontend.history_paint_qualification` | 2.750 s |
| `remediation-measurement-guards-presence-20261014-01` | `module.collaboration.frontend.presence_paint_qualification` | 2.682 s |
| `remediation-measurement-guards-timeline-20261014-01` | Timeline snapshot parser and AC-044 detector rows | 3.361 s |

`make format` passed as `remediation-measurement-guards-format-20261014-01`
(4.704 s). `make agent-finalize` passed as
`remediation-measurement-guards-finalize-20261014-01` (18.286 s); retained-run
maintenance was skipped because `RESULTS_DIR` was unset. `make lint` then
passed all 11 units as `remediation-measurement-guards-lint-20261014-01`
(25.093 s, outer exit 0), including frontend type checking and backend
boundaries. Its source identity is
`sha256:c46136604766ea22e54f80c29eeb3b224e67b10305afe3e2fc432d705010d1ca`.
Roots are under `.cartulary/test-results/`. Earlier browser measurements retain
their recorded source applicability; these unit checks are not fresh timing
observations or final frozen-source qualification.

### AC-047 seeded trace implementation

Core 04 now defines `perf.anchor_stability_under_live_updates.v2` as an exact
ordered large-grid trace. It preserves the pass/fail character of AC-047: no
latency threshold, timing sample, percentile, warm-up or product retry is
invented. The first unlinked target in the default first page is excluded from
the normal 24-session update pool. The test retains a real server query response
and releases it after the local commit; request/response bodies are not changed.
Scrolling, pending sort/filter/group, stale-query delivery and a subsequent
ordinary target update remain observable product paths.

`timeline-anchor-trace.spec.ts` owns one exact measurement row,
`module.timeline.measurement.ac047_anchor_trace`. The independent frame monitor
in `anchorTraceSupport.ts` records only typed synthetic identity, integer
versions, ordered step/frame counts, mounted-row counts, grid replacements and
closed violation codes. It does not alter rendering. Its unit row
`module.timeline.frontend.ac047_trace_qualification` exercises retargeted
editors, lost draft text, version rollback, missing frames, grid replacement and
full-sheet mounting. This is a concrete DOM/frame assurance boundary, not an
assertion about every internal React render.

Current evidence contracts are observation v4, summary v5, aggregate v5 and
browser-target result v6. `seeded_trace` has null `threshold_ms`, zero timing
samples and explicit expected/completed steps. `predicate_failed` records an
observed invariant failure. The same normal fixture lease, cleanup, scheduler
exclusion, artifact digest, retry rejection and redaction checks qualify trace
evidence. Prior schemas remain unsupported for current qualification. Earlier
latency predicates retain their workloads and identities; their earlier-format
results remain historical evidence.

Implementation validation so far (roots under `.cartulary/test-results/`):

- `remediation-anchor-trace-generate-20261014-01` and `-02` failed (8.650 s and
  8.800 s, outer exit 2) because the new unit row initially retained copied
  titles and then unsorted titles. Catalog preflight rejected both; exact titles
  were corrected. `-03` passed (10.050 s); `-04` passed (8.930 s), outer exit 0.
- `remediation-anchor-trace-unit-20261014-01` passed the new detector row
  (3.232 s, outer exit 0).
- `remediation-anchor-trace-contract-20261014-01` failed
  `product/test_assertion_failure` (21.100 s, outer exit 2): the new target-result
  schema still admitted the old aggregate artifact kind. `-02` failed the same
  class (26.909 s, outer exit 2): a contract assertion still expected the old
  summary filename. Both exact version projections were corrected.
- `remediation-anchor-trace-finalize-20261014-01` failed
  `harness/child_target_failure` (0.791 s, outer exit 2): the new proof schema
  lacked the required explicit `schema_id`. That schema and its producers were
  corrected. `-02` passed (18.108 s, outer exit 0); retained-run maintenance was
  skipped because `RESULTS_DIR` was unset.
- `remediation-anchor-trace-lint-20261014-01` passed (22.501 s, outer exit 0).

Browser qualification and retained evidence review are still required for this
trace. The implementation and contract tests alone do not close AC-047.

The first retained browser trace,
`remediation-anchor-trace-browser-20261014-01`, failed
`product/test_assertion_failure` (158.947 s, outer exit 2; all 27 cleanup
operations completed). It exposed two detector/workload assumptions: grouping
legitimately replaces the bounded grid viewport, and sorting the edited synopsis
can remove the target from the current query window. Neither observation alone
proves an adopted product defect. The observer now reacquires the semantic grid
while retaining its record-version history and records replacement counts as
diagnostics. The workload prepares a non-null Analyst sort key for the target
so the real held response must contain its older version. It verifies that fact
before releasing the pending commit. No product retry, threshold relaxation,
rendering mutation or product compatibility branch was added.

The `cartulary-ui-ux-refactor` guidance was applied to this review: Core 03 and
Core 04 own edit/refresh/virtualization semantics; Grid Adapter owns vendor
components; the Timeline/harness catalogs own execution. Private DOM-instance
identity is not an adopted invariant. Product presentation, density, icon,
authorization and authoring changes are N/A for this detector-only correction.
The new browser workload still requires a passing qualified execution.

The corrected trace detector passed `remediation-anchor-trace-unit-20261014-02`
(2.976 s); the full routing/measurement contract row passed
`remediation-anchor-trace-contract-20261014-04` (24.795 s), both outer exit 0.
The preceding `-03` contract run also passed (24.481 s) before the viewport
correction. `remediation-anchor-trace-finalize-20261014-03` failed the required
stale-topology preflight (2.960 s, outer exit 2) after the unit title changed;
`make generate` was then run explicitly as `remediation-anchor-trace-generate-20261014-06`
and passed (8.940 s). This preserves the rule that authored catalog changes
require regeneration before broader validation.

`remediation-anchor-trace-browser-20261014-02` failed
`product/test_assertion_failure` (160.083 s, outer exit 2; all 27 cleanup
operations completed). The retained trace had no invariant violations and
observed the committed target version, but the synopsis cell was horizontally
virtualized after interacting with the Analyst sort header. The test now returns
to the edited column by ordinary scrolling after grouping and before the saved
value assertion. This is permitted trace choreography, not DOM fabrication or
a retry of the product mutation. It does not establish product instability.

`remediation-anchor-trace-browser-20261014-03` failed
`product/test_assertion_failure` (153.725 s, outer exit 2; all 27 cleanup
operations completed). Retained DOM evidence identifies the missing display
cell precisely: the submitted synopsis remains in a retained editor. The
prior horizontal-virtualization explanation alone was incomplete. The trace
now reveals and checks the semantic field in either editor or display form,
checks the actual mutation response's record/version/value, and continues
checking any retained editor's record binding after submission. Invalid DOM
row-version bindings now fail the frame monitor explicitly. These are detector
corrections; no retry or product-behavior exemption was introduced.

`remediation-anchor-trace-unit-20261014-03` passed (2.995 s), including retained
editor binding after commit admission and explicit invalid-version rejection.
`remediation-anchor-trace-contract-20261014-05` passed (24.853 s), including
trace cardinality/order, non-timed policy, identity/version, full-sheet mount,
retry, sensitive-field and retained-artifact rejection. Both outer exits were 0.
`remediation-anchor-trace-finalize-20261014-06` passed (18.011 s, outer exit 0);
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

### Qualified AC-047 implementation evidence

`CARTULARY_TEST_RUN_ID=remediation-anchor-trace-browser-20261014-04
CARTULARY_OUTPUT_MODE=ci make test-slice OWNER=module.timeline
ROWS=module.timeline.measurement.ac047_anchor_trace` passed in 149.466 s with
outer exit 0. All 14 work units passed, no cache hit contributed execution, and
all 27 cleanup operations completed. The exact source and graph identities are:

- Source: `sha256:05876452340916412c2359425837e565e0bfc5902fc3882261ca84b548d41fdb`.
- Graph: `sha256:8cf7f6eb6e95c5f7afe82b2c33fc5e9b044eee14d530e43eb242e63f4392944c`.

The retained observation completed all 11 ordered steps over 154 animation
frames. The target's prepared, committed and final versions were 2, 3 and 4.
The stale full-query response contained version 2; after delivery, the mounted
row remained at version 3. The later ordinary live update advanced it to 4.
The exact request record identity matched the target, there was one foreground
PATCH, and the draft/committed value checks passed. Maximum mounted rows were
42; one bounded grid replacement occurred on grouping and no invariant violation
was recorded. There are no timing samples or fabricated latency statistics.

The post-cleanup v5 summary and v5 aggregate report `qualified`, with scheduler
overlap count 0. They are retained beneath the exact run's
`browser-e2e-measurement/browser-groups/measurement-measurement-timeline-anchor-trace-module.timeline.measurement.ac047_anchor_trace/`
and `browser-e2e-measurement/`, respectively. This establishes the exact seeded
trace for the recorded implementation source. It is not a clean-source execution
cost portfolio, a claim about all possible concurrent schedules, or Core 05
publication qualification. Later source changes still require applicability
review and the final integrated/frozen-source qualification remains open.

The final pre-browser `make lint` run
`remediation-anchor-trace-lint-20261014-03` passed (20.498 s, outer exit 0).

The post-browser maintenance checks also passed with outer exit 0:
`remediation-anchor-trace-drift-20261014-01` (`generate-drift`, 12.072 s),
`remediation-anchor-trace-policy-20261014-01`
(`generated-artifact-policy-check`, 0.803 s),
`remediation-anchor-trace-shapes-20261014-01` (`json-shape-check`, 4.744 s), and
`remediation-anchor-trace-markdown-20261014-01` (`lint-markdown`, 22.160 s).


## AC-045 physical-fixture qualification

`remediation-ac045-measurement-v2-20261014-06` passed the exact
`module.timeline.measurement.evidence_inspector_v2` service-backed slice in
600.980 s, outer exit 0. The source was
`sha256:109cec88b4f539f3888c3027e7caf35444245846273cf9bf762bc30c04ec5f72`;
the graph was
`sha256:da8f84bd2aac03ea996ed132855f5818a7d7f46587a3c6dd4d9db08a0775b857`.
All 14 work units and all 27 cleanup steps completed successfully. The
retained-secret scan passed. The finalized summary is under the exact run's
`browser-e2e-measurement/browser-groups/measurement-measurement-evidence-inspector-module.timeline.measurement.evidence_inspector_v2/`
directory as `frontend-measurement-summary.v5.json`.

The physical v3 snapshot retained 10,000 distinct real 2 MiB PNG objects and
5,000 Timeline rows. Its final Collaboration contribution sequenced
construction intents through the ordinary stream before sealing, with exact
replay and quiescence validation. The ordinary editor preflight redeemed and
verified all 100 linked preview bodies, then disposed of the capabilities.
The measured reopen uses the workbook toggle to retain the prepared Evidence
section; row-menu Inspect intentionally opens Details and is not that start
state. No product layout or navigation behavior changed.

The observation retains one discarded warm-up and 100 measured samples with
exact fresh metadata and two-frame proof. P50 is 45.100000001490116 ms and
p95 is 57.100000001490116 ms against <=300 ms. Qualification is `qualified`
and scheduler overlap is zero. Earlier failures and their mechanism-specific
repairs are recorded in the main remediation handoff. This result closes this
predicate's implementation execution evidence at the stated source; it does
not establish Core 05 publication or a clean-source cost portfolio. Later
changes still require an applicability review or fresh qualification.


Fresh browser receipt validation passed in
`remediation-benchmark-v2-paste-20261014-02` through the exact service-backed
AC-003 row: 148.335 s, outer exit 0, all 14 units passed, all 27 cleanup
operations completed, and retained-secret scanning passed (86 files). The
single observation qualified at 352.5 ms with zero scheduler overlap.
Source: `sha256:fc464a302bd0e4627cb1bc79760aa5cfae2828aa46f2c910e3eb7ef4c72a61d7`.
Graph: `sha256:99d850dba4bbd5b360c050c0dbe168af313926e0dc6ac3c38b7bc68d35fdb0d8`.
The retained and inline browser receipt reports Chromium 153.0.8010.12,
Playwright 1.63.0, Node v24.15.0, Linux/x64 and `browser_mode=headless`.
This run cannot satisfy the headed Core 05 publication profile. It is a fresh
implementation-predicate observation with accurate execution context.

Finalizer `remediation-benchmark-v2-finalize-20261014-04` passed (18.891 s),
with retained-run maintenance skipped because `RESULTS_DIR` was unset. Lint
`remediation-benchmark-v2-lint-20261014-02` passed (22.363 s), and Markdown lint
`remediation-benchmark-v2-markdown-20261014-01` passed (23.700 s), before this
annotation. No publication claim or qualified execution-cost portfolio is
inferred from these results.


Presence binding validation passed through the exact public owner routes:
`remediation-presence-binding-contract-20261014-01` (33.748 s),
`remediation-presence-binding-routing-20261014-01` (30.959 s), and
`remediation-presence-binding-lint-20261014-01` (22.371 s), all outer exit 0.
Finalizer `remediation-presence-binding-finalize-20261014-01` passed (18.415 s);
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

The exact three-row service-backed run
`remediation-presence-binding-browser-20261014-01` passed (205.828 s, outer
exit 0): 18/18 units, 35 completed cleanup operations, one shared physical
snapshot and three isolated clones, zero scheduler overlap. The qualified
single observations were surface 46.100000001490116 ms, row
179.99985351413488 ms, and cell 205.79995116591454 ms, all <=1,000 ms.
Source: `sha256:32bf06bcb19c3e3182497e47f7fa67ddc4cdaec2c6d5642dd3fadaa4261c47be`.
Graph: `sha256:14a7c11704839f544c15830dc4931c168a23b6bdddb708f8c62bd939f8a3e378`.
Snapshot: `856723721e445931dd81d7d9d2c9ea4793c9f13088765088d4a5d8c7ecb5e2a7`.
This supplies the timed presence predicates for AC-008/AC-132 at the retained
source. It does not replace other AC-132 functional contexts or qualify a
headed Core 05 publication or clean-source execution-cost claim.

## CSV and XLSX discovery measurement implementation

Core 04 REQ-04-158 now pins the two 20,000-row discovery workloads. The authored
`contracts/performance/import-discovery.v1.json` projects those decisions; it
is not derived from Markdown. The existing Import controller emits separate
acceptance, dispatch and exact Job-admission marks, with opaque correlation
identities only. No new job engine or alternate admission route was introduced.
The presentation exposes its existing validated job binding to the observer.
The same serialized DOM observer is exercised by adversarial unit tests.

The new `timeline_large_grid_import_snapshot_v1` fixture profile reuses the
existing large-grid contributions and traffic driver, with explicit
`network_flow_claimed` consumer compatibility. Its database-only builder keeps
runtime `default`; the server's enabled claims belong to the consumer context.
The canonical observation v4 admits an explicit closed Import proof variant.
The collector and retained-artifact validator admit a synthetic record identity
only in that schema-validated proof location. The format, job state, action
count, geometry, grid binding, two frames and <=1,000 ms predicate are checked;
there is no timing slack, retry or terminal-job substitution.

| Observable obligation | Exercising identity and context | Review status |
| --- | --- | --- |
| One accepted discovery, exact immutable dispatch/replay and Job correlation, no payload in timing detail | `Workbook import lifecycle marks accepted discovery separately from dispatch replay and exact job admission`; existing `module.imports.frontend.lifecycle`, Vitest browser-unit | Passed with existing lifecycle assertions in the unit run below; replay is distinguishable and cannot qualify as one fresh measurement. |
| Reject wrong/terminal Job, apply instead of discovery, disabled/hidden cancel, missing progress, stale/readonly/offscreen editor, changed geometry, missing/repeated marks and one-frame completion | `requires one fresh discovery job with visible cancel and a stable editable Timeline binding`; `module.imports.frontend.discovery_paint_qualification`, Vitest harness-node with DOM | Passed; this validates detection, not browser performance. |
| CSV admission through visible progress/cancel and ordinary Timeline editing | `measures CSV discovery progress and cancel while Timeline accepts editing`; `module.imports.measurement.csv_discovery`, isolated Playwright measurement | Qualified in `remediation-import-discovery-browser-20261015-04`; exact observations below. |
| Independent XLSX discovery with the same grid-interaction predicate | `measures XLSX discovery progress and cancel while Timeline accepts editing`; `module.imports.measurement.xlsx_discovery`, isolated Playwright measurement | Qualified in `remediation-import-discovery-browser-20261015-04`; exact observations below. |
| Closed evidence, matching fixture/format, <= threshold, no warm-up or retries, retained collection parity and sensitive-field rejection | Existing `test-browser-work-graph.mjs` through `make harness-graph-contract` | Passed in `remediation-import-discovery-contract-20261015-03`; browser performance remains separate. |

The XLSX preparation helper writes only the five fixed trusted OOXML parts with
inline strings, fixed ZIP metadata and native raw DEFLATE/CRC. It is not a
production archive reader. File preparation, scrolling and focus preparation
precede timing. The measured action uses ordinary keyboard/pointer input;
draft discard, ordinary cancellation and terminal cleanup occur afterward.
No existing browser scenario or assertion is removed in this slice.

Retained roots use `.cartulary/test-results/` and the prefix
`remediation-import-discovery-`, all dated `20261015`:

- `format-20261015-01`: outer 2, configuration error before graph publication;
  new verification profile used the invalid token `import`. Corrected to the
  admitted `extension.import`. `explain-run` confirmed no published summary.
- `format-20261015-02`: outer 0, 6.075 s; `generate-20261015-01`: outer 0,
  10.510 s.
- `unit-20261015-01`: outer 2, `product/test_assertion_failure`, 3.468 s,
  12 cleanup steps completed. The public UI-contract facade omitted the new
  generated export, so controller tests could not obtain the timing contract.
  The standalone observer row passed. Added the facade export.
- `format-20261015-03`: outer 0, 5.913 s. `unit-20261015-02`: outer 2,
  `product/test_assertion_failure`, 2.952 s, all 12 cleanup steps completed.
  The new replay test supplied a definite failure rather than an uncertain
  transport outcome; corrected the fixture without changing production replay.
- `format-20261015-04`: outer 0, 5.772 s. `unit-20261015-03`: outer 0,
  3.369 s, both selected rows passed.
- `finalize-20261015-01`: outer 0, 20.450 s. Retained-run maintenance skipped
  because `RESULTS_DIR` was unset.
- `contract-20261015-01`: outer 2, tool exit 1, 17.700 s,
  `harness/unknown_failure`. The fixture builder incorrectly inherited the
  consumer runtime, and synthetic retained-proof fixtures assumed the default
  profile. Repaired those owner/test bindings; no schema relaxation.
- `lint-20261015-01`: outer 2, `harness/child_target_failure`, 23.122 s,
  10/11 units passed and all 20 cleanup steps completed. Type checking passed;
  Biome required an optional-chain simplification in the new observer.
- `format-20261015-05`: outer 0, 5.932 s. `finalize-20261015-02`: outer 2,
  `harness/child_target_failure`, 3.077 s, all 10 cleanup steps completed.
  Its initial shape gate rejected stale generated topology after the builder
  policy correction. `generate-20261015-02` refreshed it (outer 0, 9.890 s).
- `finalize-20261015-03`: outer 0, 20.302 s. Retained-run maintenance skipped
  because `RESULTS_DIR` was unset. Subsequent validation remains in progress.

These are scoped implementation checks. Neither CSV/XLSX performance nor the
remaining Reporting/Reference Pack/evidence/projection progress obligations
are closed by this enrollment or by a unit-test pass.

The next checks retained these additional outcomes:

- `contract-20261015-02`: outer 0, 34.610 s; harness graph, profile routing,
  collector/retained-proof parity and evidence-negative cases passed.
- `lint-20261015-02`: outer 2, `harness/child_target_failure`, 23.671 s,
  10/11 units passed and all 20 cleanup steps completed. The remaining warning
  was the nullable start-mark guard; the exact format diagnostic identified
  its optional-chain simplification. `format-20261015-06` passed (5.898 s).
- `finalize-20261015-04`: outer 0, 20.177 s; retained-run maintenance skipped
  because `RESULTS_DIR` was unset. `lint-20261015-03` passed (outer 0,
  22.417 s); `observer-20261015-01` passed (outer 0, 3.454 s).
- `browser-20261015-01`: outer 2, `product/test_assertion_failure`, 261.134 s;
  all 17 units terminated and all 32 cleanup steps completed. Both cases
  admitted their uploads, but the generic grid-scroll preparation left the
  target beneath the support drawer. Playwright's actionability log identifies
  the drawer intercepting the target pointer. Both jobs ultimately succeeded;
  both measurement observations are `incomplete`, with no elapsed sample.
  This deterministic preparation defect is not a measured threshold failure
  or evidence of product instability. Source:
  `sha256:cb70508f8c12a3709c9091926d2cf040b4c65474e1d9442883c9b1cc3b0e6eb3`;
  graph:
  `sha256:a9c952d0b69c8776b3dfa0f1ebf94c2b63e026756988b732a6952637bc6dd4e7`.

Preparation now positions the complete target field beside the drawer, verifies
uncovered hit targets after final focus preparation, and captures the single
pointer destination before timing. The observer also rejects occluded progress,
cancel controls and editors. The driver uses the Timeline-owned scalar editor
identity. Neither the drawer nor product action behavior changed, and there
is no force click, response hold, or action retry. The retained diagnostic was
textual page evidence; no screenshot assessment or UI-review session is claimed.
`format-20261015-07` passed (5.931 s), the expanded occlusion regression passed
in `observer-20261015-02` (2.531 s), and `format-20261015-08` passed (5.871 s),
all outer exit 0. Fresh browser validation of this repair remains pending.


## Import discovery geometry and ordinary editing correction

The first complete preparation repair did not qualify. Retained roots continue
to use the `remediation-import-discovery-` prefix under
`.cartulary/test-results/`:

- `finalize-20261015-05`: outer 0, 19.692 s; `lint-20261015-04`: outer 0,
  25.666 s, all 11 units passed.
- `browser-20261015-02`: outer 2, `product/test_assertion_failure`, 236.957 s,
  14/17 units passed and all 32 cleanup steps completed. Both cases failed the
  preparation hit-test before upload. Neither produced a timing sample.
  Source: `sha256:91f8620b19441cc0dfc2da9444c8cde33057be5d2e821c95a016cc3ddae9b694`.
- `format-20261015-09`: outer 0, 5.886 s; `generate-20261015-03`: outer 0,
  10.270 s; `finalize-20261015-06`: outer 0, 20.519 s.
- `contract-20261015-03`: outer 0, 35.400 s; `lint-20261015-05`: outer 0,
  23.849 s, all 11 units passed.
- `browser-20261015-03`: outer 2, `product/test_assertion_failure`, 254.445 s,
  14/17 units passed and all 32 cleanup steps completed. Analyst preparation
  and upload admission succeeded; both cases then timed out waiting for the
  editor after an extra Enter keypress. The grid interaction owner admits
  ordinary scalar editing on the pointer click itself. Enter immediately
  closes that editor. These are driver defects, not observed threshold
  violations or confirmed intermittent outcomes.
  Source: `sha256:a2fc686fd61c6375e95045fd5f16410e686ad5aad1b4fb65f9301c3e66ee6d2b`.

Core 04 and the typed workload now pin the initially null Analyst field
(`timeline.analyst_text`) of the first record in record-ID order. It is a default
writable scalar visible beside the non-modal drawer. The uploaded 20,000-row
file is unchanged. The preparation-only custom centering calculation is removed;
whole-field hit tests remain mandatory before timing. This corrects a new
workload without changing product layout or manufacturing visibility during
measurement. Neither earlier workload produced qualifying evidence.

The driver now performs exactly one upload activation, one ordinary pointer
click that selects and opens the scalar editor, and keyboard input of `x`.
It does not issue a second Enter, force a click, retry an action, hold a response,
or alter the threshold. `format-20261015-10` passed (outer 0, 5.834 s), the
current-field observer regression passed in `observer-20261015-03` (outer 0,
2.587 s), and `finalize-20261015-07` passed (outer 0, 20.451 s). All finalizer
runs above skipped retained-run maintenance because `RESULTS_DIR` was unset.
Fresh browser validation remains required.


## Qualified CSV and XLSX discovery observations

`make service-backed-test-slice OWNER=module.imports
ROWS=module.imports.measurement.csv_discovery,module.imports.measurement.xlsx_discovery`
passed in `remediation-import-discovery-browser-20261015-04` with outer exit 0,
17/17 work units passed and wall time 226.239 s. Both v5 summaries have
`qualification_outcome=qualified`, one measured observation, no warm-up or action
retry, and zero scheduler overlap:

| Predicate | Elapsed | Threshold | Observed completion |
| --- | ---: | ---: | --- |
| `perf.import_csv_discovery.visible_with_cancel.v1` | 609.5999999940395 ms | <=1,000 ms | Exact running discovery Job, enabled cancel and active Analyst editor containing `x` over two consecutive frames. |
| `perf.import_xlsx_discovery.visible_with_cancel.v1` | 462.19999998807907 ms | <=1,000 ms | Same independently exercised predicate for the actual XLSX upload. |

Both proofs bind the original row version `1`, stable record/field identity,
1920×1080/DSF1 geometry, unchanged scroll, one action and the uncommitted draft.
The fixture uses 20,000 Timeline rows, the exact two-column input, 25 authenticated
sessions and prescribed background updates. All 32 cleanup steps completed.
Ordinary post-measurement cancellation was requested for each Job; both Jobs
completed successfully before cancellation took effect. This is the retained
terminal outcome, not a claim that cancellation interrupted processing.

Source: `sha256:63732c750d0e1f173a9dec47fc7218ec4c51985917d810fcf04ed46248cfda00`.
Graph: `sha256:a9c952d0b69c8776b3dfa0f1ebf94c2b63e026756988b732a6952637bc6dd4e7`.
Snapshot key: `d732680a7919737e9ae0e80858e46b3d5d61f5b3d44edbf8d8f7b545dafba597`.
The source tree was dirty; these are scoped product-predicate observations, not
a clean-source cost portfolio or Core 05 benchmark publication. Import review,
mapping/apply, other Job mechanisms and coordinated adoption retain their
separate obligations. No production behavior or threshold was changed in the
last driver correction, and no failed observation was discarded or reclassified.

The Import discovery follow-up documentation gate
`remediation-import-discovery-markdown-20261015-01` passed with native exit `0`
(24,000 ms). Generated drift validation
`remediation-import-discovery-drift-20261015-01` passed with native exit `0`,
4/4 work units (13,500 ms). These checks validate documentation/projections;
the qualifying CSV/XLSX observations remain the explicitly retained browser04
root above.

## Import apply measurement implementation

Core 04 now separately defines CSV and XLSX apply under
`perf.import_{csv,xlsx}_apply.visible_with_cancel.v1`. The workload uses the
same 20,000-row source and large-grid fixture as discovery. The one-based
Import column ordinals are 1 for Activity Synopsis and 2 for the unmapped
source note. Exactly one approved, selected unit must be current before Apply.

The controller records acceptance before current-session/selection preflight,
then exact fresh dispatch and job admission using the same opaque transaction.
Replay records another dispatch with `replay=true`, never another accepted
action. The common paint observer independently requires the correct job
purpose and identity, ordered marks, enabled visible cancellation, and stable
editable Timeline binding across two frames. Discovery and apply keep distinct
predicate and proof identities. The existing observation envelope remains v4;
the new closed `cartulary.import_apply_measurement_proof.v1` is an explicit
additional variant. Earlier discovery proof semantics are unchanged; unknown
proof shapes still fail admission.

`importProgressScenario.ts` shares the ordinary preparation and measurement
choreography. Apply preparation performs real discovery, review, approval and
selection. The single Apply control now precedes unit review beside Job status;
there is no post-start scroll or measurement-only rendering. Cleanup refreshes
ordinary outcomes and records enum-only unit state separately from cancellation
and terminal Job status. An enabled cancel control alone proves no rollback.

Initial roots share prefix `.cartulary/test-results/remediation-import-apply-`
and suffix `-20261016-01`:

| Root | Command / result |
| --- | --- |
| `format` | `make format`, native 0, 6.019 s. |
| `generate` | `make generate`, native 0, 9.970 s. |
| `unit` | Six exact Import frontend rows covering discovery/apply paint, lifecycle and all current assistant presentation tests; native 0, 4.323 s, four units. |
| `finalize` | `make agent-finalize`, native 0, 20.618 s; maintenance skipped because `RESULTS_DIR` was unset. |
| `graph` | `make harness-graph-contract`, native 0, 35.310 s; schema, exact predicate routing, threshold, altered proof and retained evidence rejection. |
| `shape` | `make json-shape-check`, native 0, 6.162 s. |
| `lint` | `make lint`, native 2, 23.526 s, `harness/child_target_failure`; 10/11 units passed, all 20 cleanup steps completed. Frontend typecheck found the missing new predicate type in the shared traffic helper. This is an introduced projection error, not a stability incident. |

The traffic helper type was corrected; paint validation additionally rejects
admission preceding dispatch. These final edits require fresh narrow validation
and browser qualification. No apply timing observation is claimed yet.

Final pre-browser checks passed: `format-20261016-02` (6.125 s),
`finalize-20261016-02` (19.890 s; maintenance unset),
`unit-20261016-02` (4.677 s), `typecheck-20261016-01` (23.016 s),
`drift-20261016-01` (16.281 s), and `catalog-20261016-01` (35.258 s),
all native exit 0 under the same prefix above.

`remediation-import-apply-browser-20261016-01` failed with native 2,
`product/test_assertion_failure`, in 241.949 s. Both formats completed ordinary
discovery, but the new exact `getByLabel` selector did not match the wrapped
select labels during mapping preparation. Retained accessibility state shows
the three correctly named comboboxes and the expected 20,000-row unit. No Apply
activation or timed observation occurred. The selector was corrected to exact
accessible combobox roles/names, without changing workload or product timing.
The run completed all 32 cleanup operations. Its source is
`sha256:e70bd71f44e49becc55e2c462a073f92db528f45f293806d46699e7c8c66424f`
and graph is
`sha256:187729a324b39f91114752e92d21cdac497ab279abb7693e6b7db121f5bf986d`.

### Import unit cancellation defect

`remediation-import-apply-csv-browser-20261016-02` reached a 206.9 ms single
observation, but the run failed (`product/test_assertion_failure`, native 2,
358.457 s). The Job stayed running, cancellation became uncertain, and four
owned-session revocations failed. Harness cleanup recorded three failures and
one blocked services-close operation; the exact fixture lease was quarantined.
The observation is investigative, not qualified timing evidence.

Inspection found that Import took the Jobs transition lock before all per-unit
owner work and retained it through commit. A 20,000-row transaction therefore
blocked cancellation and lease renewal. The repair moves execution fencing and
actor locking to the short outcome commit barrier and observes cancellation or
execution loss while work remains uncommitted. Rollback gets an independent
bounded context. Unit atomicity, earlier committed outcomes, final current
actor checks and the Jobs-locked commit/cancel race remain mandatory. Core 01
REQ-01-620c now makes that scheduling boundary explicit. Mechanism-specific
regression and fresh browser qualification remain required.

## Import cancellation and exact recovery

The CSV apply run `remediation-import-apply-csv-browser-20261016-02` exposed
a deterministic product defect, not confirmed intermittent instability. Its
206.9 ms timing observation is unqualified because terminal cancellation and
owned cleanup failed. The import unit held the Jobs transition lock and actor
row lock throughout its large transaction, preventing ordinary cancellation,
lease renewal and session revocation. No retry or longer deadline repairs that
mechanism.

Imports now observes its exact active execution while staging owner work and
cancels the transaction context when cancellation is requested. The Jobs lock
and current active-actor revalidation form a short commit barrier after staging.
Atomic unit publication, current incident authorization, prior committed-unit
outcomes and indeterminate-commit recovery are preserved. Rollback has its own
bounded context. Core 01 REQ-01-620c makes this boundary explicit.

`make test-slice OWNER=module.imports` with exact rows
`uncommitted_unit_cancellation`, `partial_cancellation_outcomes_a14b66f830`,
`apply_transaction_current_state_bc8ad0ac32` and
`unit_commit_crash_recovery_108cd054a6` (all under `module.imports.integration`)
passed in `remediation-import-cancellation-integration-20261016-01`: native
exit 0, 52.743 s, three work units, all 14 cleanup steps completed. The new
regression blocks actual owner SQL before commit and requires ordinary authorized
cancellation to interrupt it before the test releases the blocking lock.

The retained browser resource could not be recovered through a public operation.
`test-run-recover` now exposes the existing native fixture recovery engine with
exact original run, private runtime and lease inputs. It does not discover or
signal processes by name, select newest runs, reset shared services, or rewrite
original results. The native owner validates the binding and process authority;
cleanup uses the existing dependency graph and physical adapters. Unit cases
reject missing inputs, foreign run/lease and symlinks and preserve original
execution evidence. The command is helper-only, outside the fixed cost roster.

Recovery validation retained these attempts:

| Root suffix after `remediation-exact-recovery-` | Result and significance |
| --- | --- |
| `generate-20261016-01` | Native 2; new helper was missing its explicit observability disposition. Added outside-roster membership; no silent scope widening. |
| `generate-20261016-02` | Native 0; generation after membership correction. |
| `unit-20261016-01` | Native 0, 2.006 s; `harness.browser.unit.native_fixtures`, including recovery entrypoint cases. |
| `import-20261016-01` | Native 2 before physical cleanup; the command-mode Make wrapper dropped required inputs before its own admission. Fixed wrapper and child forwarding together with regression coverage. |
| `import-20261016-02` | Native 2 / command 12, 0.080 s; physical browser adapter lacked the pinned Node environment. Browser failed and its suite was correctly blocked. Added the existing Make-owned Node binding; retained proof remained intact. |
| `import-20261016-03` | Native 0, 0.940 s; exact browser stack and managed suite completed recovery, private runtime removed. Original failed browser evidence remains failed. |
| `command-20261016-01` | Native 0, 57.725 s; public command-registry parity slice. |
| `shape-20261016-01` | Native 0, 6.372 s; JSON shape check. |
| `drift-20261016-01` | Native 0, 26.488 s; generated drift check. |

`remediation-import-cancellation-lint-20261016-01` passed `make lint` with
native exit 0, 34.799 s and all 11 work units. Finalizers
`remediation-import-cancellation-finalize-20261016-01` and
`remediation-exact-recovery-finalize-20261016-01/02` passed with native exit 0.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
The attempted helper name `make harness-smoke-task-surface-report` has no public
Make rule and did not execute a test; its script remains in the declared harness
tier. No pass is inferred for that unexecuted helper.

Fresh CSV/XLSX apply qualification remains required after this repair. These
mechanism tests and recovery results do not alone satisfy AC-027 or AC-046.

### Qualified CSV and XLSX apply observations

`remediation-import-apply-repaired-browser-20261016-01` completed with native
exit 0 in 232.095 s: 17/17 work units, all 32 cleanup operations completed,
no cached units, qualified measurement aggregate v5. Source:
`sha256:5deb44873aba69184685552f34fd227f60b97fc91bef51e59ded210528b04658`;
graph: `sha256:187729a324b39f91114752e92d21cdac497ab279abb7693e6b7db121f5bf986d`.
Both exact AC-027/AC-046 predicates used one observation without warm-up or retry:
CSV 343.4 ms; XLSX 302.7 ms, each <=1,000 ms. Each proof binds its exact running
apply Job, 20,000-row/two-column source, one selected unit, column-1 mapping,
25 authenticated sessions, 4.8 background updates/second, two visible stable
frames, enabled cancellation and an uncommitted Timeline draft. Both ordinary
cancellations reached `canceled`; unit outcomes were `failed`, as required for
canceled uncommitted units. Those enum states do not represent committed row
loss. The separate SQL-blocking regression proves rollback and absence of owner
rows/journal effects.

`remediation-exact-recovery-finalize-20261016-03` passed before this run
(native 0, 20.570 s); retained-run maintenance was skipped because `RESULTS_DIR`
was unset. The previous failed roots remain separate evidence, and no finite
pass streak or Core 05 publication claim is made.

## Reporting progress predicate implementation

The proposed Core 04 workload now owns
`perf.reporting_snapshot.visible_with_cancel.v1` and
`perf.reporting_render.visible_with_cancel.v1`. Typed input is
`contracts/performance/reporting-progress.v1.json`; the independently admitted
fixture is `timeline_large_grid_reporting_snapshot_v1`. It reuses the physical
large-grid owner contributions without adding rows or bypassing product
admission. Snapshot and internal Slidev render admission are measured separately.
The report snapshot is prepared through the ordinary UI before background traffic;
its current supporting read remains inside the render interval.

Reporting's retained controller emits acceptance, fresh dispatch and exact Job
admission marks. A shared pure browser detector now checks Import and Reporting
Job/grid visibility, identity, bounds, geometry and two-frame completion. The
Reporting observation variant is closed and preserves synthetic record/job and
snapshot identities only at its validated proof location. The owner adapter checks
fixture, predicate, sampling, traffic, exact threshold and outcome. The real
browser scenarios also assert the actual single POST body, discard the draft and
retain ordinary cancellation/terminal cleanup. Cleanup errors preserve the primary
measurement error.

Initial validation, `.cartulary/test-results` roots:

| Root | Result |
| --- | --- |
| `remediation-reporting-progress-generate-20261016-01` | Native exit 0. |
| `remediation-reporting-progress-format-20261016-01` | Exit 2; lint rejected throwing from `finally`. Aggregate error propagation moved after cleanup, preserving primary/secondary errors. |
| `remediation-reporting-progress-format-20261016-02` / `-03` | Exit 0, 6.156 / 6.120 s. |
| `remediation-reporting-progress-unit-20261016-01` | Exit 2; Reporting's new generated timing value was missing from the intermediate authored facade. Shared detector passed; controller/application tests exposed the missing export. |
| `remediation-reporting-progress-unit-20261016-02` | Exit 0, 4.058 s; controller, application binding and Reporting paint detector. |
| `remediation-reporting-progress-import-unit-20261016-01` | Exit 0, 2.571 s; unchanged discovery/apply detector obligations after extraction. |
| `remediation-reporting-progress-finalize-20261016-01` | Exit 0, 20.318 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |

No qualifying Reporting browser observation has yet been recorded at this checkpoint.
AC-030 / the Reporting portion of AC-046 and coordinated adoption remain open.

### Reporting admission defect and repair qualification

The first ordinary Reporting progress measurement,
`remediation-reporting-progress-browser-20261016-01`, exited 2 after 352.367 s:
14/17 harness units passed, both product measurements failed, and all 32 cleanup
activities completed. The render fixture's snapshot POST exceeded the client's
30 s observation window before a Job was returned; the snapshot measurement
never received a Job within its visible-progress observation. This is not
qualifying timing evidence. The exact Playwright reports and traces remain in
that run's `browser-e2e-measurement/browser-groups` directories.

Inspection found incident-sized source collection, projection, validation and
canonicalization in the synchronous snapshot admission path. The coordinated
candidate now separates the small durable request admission from the worker's
consistent source capture. Core 01 owns the changed capture-time and explicit
boundary semantics. Reporting retains its own closed admission, its digest and
Reference Pack pin independently of the immutable capture. Migration 92 admits
this state in the existing owner payload row; no second lifecycle table or
legacy reader is introduced. `snapshot_reporting.snapshot_create_v3` retires v2.
A pending capture is explicit SQL NULL, never a fabricated empty export model.

The regression blocks a source-table read while requiring the authenticated
admission, exact pending replay and cancellation to complete. Interrupted work
must publish neither a materialized payload nor a snapshot. Broader replay,
recovery, concurrent admission and migration checks remain required. The browser
predicates remain unqualified until the repaired product passes fresh execution.

The repaired run `remediation-reporting-progress-browser-20261016-02` exited 2
with 15/17 units passed after 349.093 s. All 32 cleanup activities completed.
The exact snapshot row passed with a 208 ms observation under the prescribed
25-session / 4.8 updates-per-second workload. The render row remained incomplete:
its setup snapshot was admitted in about 31 ms but terminally failed after about
93 s, before a render action was attempted. The Job's safe public summary exposed
only `internal_error`; the failing stage was not retained there. This is a
separate unresolved capture-completion defect, not a report timing result.

The next diagnostic slice retains closed snapshot-stage/failure-kind tokens,
without raw SQL, source values, paths or arbitrary error messages. Setup now
observes terminal snapshot status directly and fails promptly when it is failed
or canceled, instead of waiting the full readiness timeout after terminal failure.
`agent-finalize` for the admission repair passed in 20.287 s; retained-run
maintenance was skipped because `RESULTS_DIR` was unset.


## Qualified Reporting progress after snapshot publication repair

Both ordinary Reporting mechanisms now have fresh passing run roots on source
`sha256:13b270115003b640e3881c2f375cb3821d95ee394026de51efdccaf75577e831`:

| Predicate | Exact run | Observation | Execution |
| --- | --- | --- | --- |
| `perf.reporting_render.visible_with_cancel.v1` | `remediation-snapshot-publication-report-progress-20261016-01` | 200.8 ms, one sample, qualified, scheduler overlap zero | Native exit 0, 209.781 seconds, 15/15 units, all 28 cleanup steps complete |
| `perf.reporting_snapshot.visible_with_cancel.v1` | `remediation-snapshot-publication-snapshot-progress-20261016-01` | 177.4 ms, one sample, qualified, scheduler overlap zero | Native exit 0, 198.905 seconds, 15/15 units, all 28 cleanup steps complete |

Both use the ordinary claimed-profile UI, authentication and authorization,
large-grid fixture, 25 authenticated sessions, prescribed background traffic,
anchored Timeline edit and visible two-frame progress/cancel predicate. Report
preparation creates its real large-grid snapshot before measurement. Neither
observation retries the measured action. Cancellation and terminal cleanup are
retained separately. These observations establish admission responsiveness and
continued editing, not complete-render latency, arbitrary-size export support,
a clean-source execution-cost portfolio or Core 05 publication qualification.

Snapshot publication integration passed in 67.917 seconds; corrected worker and
codec regression root `remediation-snapshot-publication-unit-20261016-02`
passed in 7.135 seconds. Finalizer
`remediation-snapshot-publication-finalize-20261016-01` passed in 20.761 seconds;
maintenance was skipped because `RESULTS_DIR` was unset. Generated drift,
artifact policy and JSON shape roots `remediation-render-recipe-{drift,policy,shape}-20261016-01`
passed in 14.689, 0.897 and 5.214 seconds respectively. All native exits were 0.


## Reference Pack progress implementation

Core 04 now names independent import, reverify and selected-refresh predicates
with one exact signed host-registry input. The administration page and Timeline
page share one foreground authenticated session; the existing 24-session traffic
remains active. The shared paint observer reads both same-origin documents,
awaits animation frames from each, and rejects replaced documents, changed
geometry, missing/altered action marks, repeated actions, mismatched Job purpose,
disabled cancellation and stale record bindings. Preparation Jobs cannot satisfy
the measured action. The observer admits no synthetic worker or held response.

The ordinary controller records acceptance before request capture and mandatory
preflight, dispatch with replay identity, and the admitted Job. Presentation adds
semantic Job bindings to the existing progress and cancel controls. Exact typed
proof includes only synthetic identities and bundle metadata; the existing
collector and retained finalizer apply the same closed redaction policy. The
profile adapter reconstructs the declared input and rejects changed bytes, size,
preparation date or payload digest. Bundle role metadata now shares one captured
preparation date. Independent production TUF validation remains authoritative.

The shared bundle producer moved from browser test support into
`tools/harness/browser/reference-pack-fixture.mjs`, with no legacy wrapper. This
repairs the actual isolated-generation failure in finalize 01: an evidence
adapter imported a browser-tree module absent from the generator's admitted
inputs. Startup, browser tests and evidence validation now consume that one
harness-owned producer. No documentation input or additional runner was added.

All roots below use prefix `remediation-reference-progress-` and date
`20261016`; native exits were observed independently:

| Suffix | Result | Scope |
| --- | --- | --- |
| `generate-20261016-01` | 0, 10.480 s | Initial projections |
| `format-20261016-01` | 0, 6.213 s | Authored source formatting |
| `finalize-20261016-01` | 2, 19.675 s | Isolated generation rejected the cross-tree producer import; 10 cleanup operations completed |
| `generate-20261016-02` | 0, 10.090 s | Shared producer ownership repair |
| `finalize-20261016-02` | 0, 20.293 s | Generated structure and selection closure |
| `unit-20261016-01` | 0, 5.902 s | Exact controller recovery/instrumentation and companion-frame adversarial cases |
| `types-20261016-01` | 0, 21.321 s | Frontend type checking |
| `graph-20261016-01` | 0, 38.461 s | Collection/retention and malformed-proof rejection |
| `markdown-20261016-01` | 0, 24.660 s | Specification/support prose lint only |

Retained-run maintenance was skipped because `RESULTS_DIR` was unset. Physical
qualification remains pending. Before physical execution, the preparation reset
was corrected to clear the three exact administration marks; Workbook-only mark
resetting does not clear a different owner's namespace. This neither retries an
action nor changes its timing interval.


The first physical root,
`remediation-reference-progress-browser-20261016-01`, returned native exit 2
in 264.996 s, with all 36 cleanup operations complete. Import produced a
qualified 117.6 ms observation, one sample and zero overlap. Reverify and refresh
stopped before measurement because the test treated the GET version resource's
`pack_version` string as a nested resource. The assertion now checks the actual
`data` resource; the production API and timing contract are unchanged. Aggregate
qualification correctly rejected the incomplete selection. Source was
`sha256:fddbbaa787648258877abed4ea111a7703b88784ea5b0e1012e7ec33685ca022`.

The second physical root, `remediation-reference-progress-browser-20261016-02`,
returned native exit 2 in 260.262 s, with all 36 cleanup operations complete.
Refresh and reverify qualified at 125.7 ms and 113.3 ms, one observation each,
zero overlap. Import reached 174.3 ms but failed the added post-measurement
multipart parser: Chromium's network report omits uploaded file bodies. Its
incomplete outcome and the failed aggregate cannot qualify the import predicate.
Source: `sha256:050b44a87bab4f7c3f1c3b9e01021f41a2d45238af00975266c72bab43a8606b`.

The repaired driver compares the actual browser input File bytes before timing,
requires one ordinary multipart dispatch, then reads the successfully admitted
exact version through the public API after the Job finishes. It checks payload
digest, availability, inactive/staged status and no pending work. Existing service
test `preserves upload bytes filename media hint and staged-only metadata across
lost-response replay` independently validates exact multipart construction.
This combination does not claim that Playwright retained multipart network bytes.
No request interception, response hold or timing alteration is introduced.

Format 02 and finalize 03 passed in 6.415 s and 20.701 s; format 03 and finalize 04
passed in 6.724 s and 20.817 s. Format 04 rejected a throw lexically inside finally
(outer 2, 6.283 s, all 11 cleanup operations complete). The shared public-version
assertion now lives outside finally and is still caught as cleanup evidence.
Format 05 passed in 6.142 s and finalize 05 in 21.721 s. All successful native
exits were 0; retained-run maintenance remained skipped because `RESULTS_DIR`
was unset. The exact full Reference Data owner is the next validation scope.

## Qualified Reference Pack progress and owner evidence

`remediation-reference-progress-owner-20261016-01` ran
`make test-slice OWNER=module.reference_data` and returned native exit 0 in
225.956 s: 33/33 work units and all 56 cleanup operations complete. Import
qualified at 84.1 ms, reverify at 109.9 ms and refresh at 150.4 ms. Each predicate
has one fresh observation, a 1,000 ms inclusive threshold and zero scheduler
overlap. File preparation, actual public admission, ordinary trust validation,
concurrent Timeline editing, two-document paint qualification and successful
terminal Job/version checks all completed without retries or held responses.

Source: `sha256:20fcde567eb385f659db47e5ee0195d3a833cd1d8b611dc2b4b7c0e56e032901`.
Graph: `sha256:f52af56df58a9593f73050b1cc8fd4cfa65baf1a254c8581f2838843c752cdb0`.
The explicit manifest
`.cartulary/remediation-evidence/reference-progress-owner-20261016.json`
selects only that root. `make test-evidence-audit OWNER=module.reference_data`
with that manifest passed in `remediation-reference-progress-audit-20261016-01`,
native exit 0, 2.793 s. These results qualify that owner/source; the dirty tree
does not establish a clean-source execution-cost portfolio or Core 05 claim.
Subsequent executable changes require explicit applicability review or rerun.

## Qualified Import discovery after admission repair

`remediation-import-discovery-physical-20261016-01` ran the exact
`module.imports.measurement.csv_discovery` and
`module.imports.measurement.xlsx_discovery` rows through
`make test-slice OWNER=module.imports`. Native exit was 0 in 422.623 seconds:
17/17 work units and all 32 cleanup operations complete. The physical fixtures
used the dedicated scratch root outside `/tmp`.

`perf.import_csv_discovery.visible_with_cancel.v1` qualified at
604.5999999940395 ms; `perf.import_xlsx_discovery.visible_with_cancel.v1`
qualified at 371.40000000596046 ms. Each has one observation, a 1,000 ms
threshold and zero scheduler overlap. Fresh in-time admission enables the
visible cancellation affordance; mutation still obtains current authorized Job
state. Replay and late acknowledgement do not enable that affordance. No action
retry, delayed response or threshold adjustment was introduced.

Source: `sha256:700d45dd9e4772407495836300b4525ac4887972c792d3f487d19319bd8c3815`.
Graph: `sha256:1454b4d899b8143a8ca133b9dfd08ac7ce4e3bf43e58d8c91b2639ba482a4e02`.
The current cleanup is complete; the prior aggregate release's cleanup symptom
remains unconfirmed and is not declared repaired by this later result.

## Frozen integrated measurement qualification

`remediation-cost-reference-release-check-20261016-01-cold` completed with native
exit 0 in 6,257.716 seconds, all 1,275 work units passed, all 1,980 cleanup
operations completed and retained-secret scan passed. The clean executable
source is `b3463443139a16a16aa1ec2d820c9ab712295bfdf19ca70f4f6b45fb8642fcb2`
at commit `7ccec27a9d905bcd4d8024df58d67c7d5a73f53d`; graph
`195b03c43c1af2124fcb2108cfdf55cd04a1f2d5ee09621d9b71f0e6480addc9`.
The retained `browser-e2e-measurement/frontend-measurement-aggregate.json` is
v5 and `qualified`. Its four profile groups each have one builder, respectively
2/4/2/21 isolated clones, and zero scheduler overlap. All 29 rollups passed.
This closes the integrated applicability review at this source; later source
changes retain their own impact review. It is not Core 05 publication evidence
or a multi-observation execution-cost portfolio. Values below are rounded to
one decimal millisecond; exact samples and statistics remain in retained evidence.

| Predicate | Sampling / observed statistic | Outcome |
| --- | --- | --- |
| `perf.evidence_inspector.metadata_shell.v2` | 100 observations: p95 46.2 ms | Passed |
| `perf.evidence_verification.visible_with_cancel.v1` | One observation: 159.2 ms | Passed |
| `perf.import_csv_apply.visible_with_cancel.v1` | One observation: 218.3 ms | Passed |
| `perf.import_csv_discovery.visible_with_cancel.v1` | One observation: 544.0 ms | Passed |
| `perf.import_xlsx_apply.visible_with_cancel.v1` | One observation: 230.9 ms | Passed |
| `perf.import_xlsx_discovery.visible_with_cancel.v1` | One observation: 335.0 ms | Passed |
| `perf.reporting_render.visible_with_cancel.v1` | One observation: 216.3 ms | Passed |
| `perf.reporting_snapshot.visible_with_cancel.v1` | One observation: 214.5 ms | Passed |
| `perf.anchor_stability_under_live_updates.v2` | Seeded trace: 11/11 steps | Passed |
| `perf.reference_pack_import.visible_with_cancel.v1` | One observation: 215.8 ms | Passed |
| `perf.reference_pack_refresh.visible_with_cancel.v1` | One observation: 183.8 ms | Passed |
| `perf.reference_pack_reverify.visible_with_cancel.v1` | One observation: 123.5 ms | Passed |
| `perf.timeline_blank_row_create.v1` | 100 observations: p95 122.6 ms | Passed |
| `perf.timeline_change_set_rollback.rendered.v2` | One observation: 356.9 ms | Passed |
| `perf.timeline_filter.first_useful_viewport.v1` | 100 observations: p95 166.9 ms | Passed |
| `perf.timeline_filter.stable_viewport.v1` | 100 observations: p95 185.8 ms | Passed |
| `perf.timeline_group.first_useful_viewport.v1` | 100 observations: p95 170.0 ms | Passed |
| `perf.timeline_group.stable_viewport.v1` | 100 observations: p95 160.6 ms | Passed |
| `perf.timeline_history_entry_rollback.rendered.v2` | One observation: 264.9 ms | Passed |
| `perf.timeline_paste_20x5.v2` | One observation: 391.2 ms | Passed |
| `perf.timeline_presence_cell.rendered.v1` | One observation: 183.1 ms | Passed |
| `perf.timeline_presence_row.rendered.v1` | One observation: 182.2 ms | Passed |
| `perf.timeline_presence_surface.rendered.v1` | One observation: 36.7 ms | Passed |
| `perf.timeline_row_restore.rendered.v2` | One observation: 392.8 ms | Passed |
| `perf.timeline_sort.first_useful_viewport.v1` | 100 observations: p95 183.0 ms | Passed |
| `perf.timeline_sort.stable_viewport.v1` | 100 observations: p95 198.0 ms | Passed |
| `perf.timeline_summary_focus_edit.v1` | 100 observations: p95 36.8 ms | Passed |
| `perf.timeline_summary_selection_down.v1` | 100 observations: p95 28.3 ms | Passed |
| `perf.typing_ack.v1` | 100 observations: p95 26.6 ms | Passed |
