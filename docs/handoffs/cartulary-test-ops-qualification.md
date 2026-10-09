# Cartulary test operations qualification

## Delivery

The repository skill lives in `.agents/skills/cartulary-test-ops`. The adopted
interface is Testing Harness NLSpec Section 8.9 (TH-HARNESS-REQ-289 and
TH-HARNESS-AC-136). Two closed schemas, the task surface, helper ownership and
evidence verification routing project that contract. Generated projections are
maintained through Make.

The scheduler supplies diagnostic activity directly to a serialized publisher.
The reader selects one exact invocation, prefers terminal summaries, and does
not audit acceptance evidence. Private canonical-event staging remains private.
Publication revisions and scheduler progress cursors remain separate. Quiet-lock
boundaries pause publication, and a matching paused cursor returns immediately.
The reader does not acquire host, fixture, service or workspace admission.

`unsupported_live` is reserved for an explicitly identified producer without
live support. Current tool-only runs have no preterminal identity artifact;
their absence is `not_published`, and their published summary is terminal.
No inference from missing files is used to manufacture the reserved state.

## Executable verification

All result roots below are under `.cartulary/test-results/`. Native Make outcomes
were observed separately from reported harness summaries. Initial failing runs
are retained; successful reruns do not erase them.

| Command | Most recent result | Run ID |
| --- | --- | --- |
| `make test-slice OWNER=harness.command_surface` | Pass, 2/2 units | `test-ops-command-03` |
| `make test-slice OWNER=harness.evidence_accounting` | Pass, 1/1 unit | `test-ops-evidence-04` |
| `make generated-artifact-policy-check` | Pass, 3/3 units | `test-ops-policy-01` |
| `make lint-scripts` | Pass, 2/2 units | `test-ops-scripts-01` |
| `make lint-markdown` | Pass, including this qualification report | `test-ops-markdown-03` |
| `make agent-finalize` | Pass, 1/1 unit; retained maintenance skipped | `test-ops-finalize-04` |
| `make harness-contract` | Pass, 2/2 units; all 117 contract tests pass | `test-ops-contract-02` |
| `make json-shape-check` | Pass, 3/3 units | `test-ops-shape-02` |
| `make generate-drift` | Pass, 4/4 units | `test-ops-drift-02` |

Finalization has no qualifying full warm check root: `RESULTS_DIR` is unset and
retained-run maintenance is skipped. The focused suites cover exact selection,
unpublished and tool-only evidence, cursor/heartbeat distinctions, waits,
signal interruption, malformed/unsafe identity, 16 MiB overflow, repeated large
replacement, four concurrent readers, bounded omissions, quiet lifecycle order,
sibling completion and dependency skips. The publisher stop case verifies no
later timer/update writes. The runner stops it before retained-secret scanning.

## Skill scenario review

This is a manual procedure review, separate from executable harness evidence.
It is not an independent agent evaluation or a release acceptance claim.

| Scenario supplied to the procedure | Resulting decision |
| --- | --- |
| Explain a confirmed Vitest hook timeout with preserved original error | Preserve `timing/timeout_failure` and original cause; do not use registration location as assertion evidence. |
| Explain a Playwright per-test timeout | Preserve its product assertion classification; do not apply Vitest mapping. |
| Explain `STACK_TRACE_ERROR` without original cause | Report cause unresolved; no invented timeout. |
| Report a cached exact-row pass | State cached/subset scope, no fresh execution or full-owner closure. |
| Observe a checkout lacking `test-run-status` | Use the native session and existing terminal diagnosis; no replacement script or new test. |
| Diagnose a failed invocation without rerun authority | Read exact structured evidence and bounded referenced logs; do not rerun, repair or finalize. |
| Native command still runs after a pass summary | Report the summary only; await outer outcome and required finalization. |
| Matching paused cursor while measurement runs | Stop structured polling, use native session, and avoid known sibling polling. |
| Development task requires finalization | Preserve that existing authorization; diagnosis-only restrictions do not override it. |

Skill frontmatter/name validation passed with the skill-creator validator.
Three relative reference links and the interface metadata were inspected.

## Measurement qualification

All three accepted profiles passed the unchanged public target. Each native
Make exit was 0; all 24 units passed; the frontend aggregate was `qualified`;
all 35 cleanup steps completed; retained-secret scanning passed. The explicit
read-only `make harness-observability-check RESULTS_DIR=<exact-root>` also
passed for each root. These are measurement qualifications, not full-owner or
release closure.

| Readers | Run ID | Wall time (s) | Observation behavior |
| --- | --- | --- | --- |
| 0 | `test-ops-measure-0` | 359.861 | Native session only; no status calls. |
| 1 | `test-ops-measure-1` | 358.502 | Seven bounded status calls; revision 17 reported `paused_measurement`; structured polling stopped. |
| 4 | `test-ops-measure-4-native` | 381.813 | Four independent readers queried exact sibling `test-ops-measure-0`, with one second idle after each completed call. |

All three used source digest
`sha256:da79c43949aa44fc9cc85b0f9bf12e2b237ccea552018f747a6cae4960812921`
and graph digest
`sha256:23673b5786a209d36a974ab03f073413eba5d4eeb5344af6b03fe6b699fdb201`.
The four-reader profile is intentional sibling-reader pressure during quiet
intervals; it is not the skill's recommended monitoring loop. Four-way overlap
was confirmed by invocation timestamps. Its readers completed 212, 211, 211,
and 211 calls, with all 845 responses reporting reader success and the exact
sibling's retained result. No reader acquires admission or controls the tests.

Each timing predicate retained 100 samples per run and its existing threshold.
The other four selected Network Flow measurement rows also passed.

| Predicate | Threshold (ms) | p95, zero readers (ms) | p95, one reader (ms) | p95, four readers (ms) |
| --- | --- | --- | --- | --- |
| `perf.timeline_blank_row_create.v1` | 150 | 115.4 | 110.1 | 123.8 |
| `perf.timeline_summary_focus_edit.v1` | 100 | 35.0 | 35.1 | 37.0 |
| `perf.timeline_summary_selection_down.v1` | 100 | 27.6 | 29.0 | 27.4 |
| `perf.typing_ack.v1` | 100 | 27.8 | 27.2 | 32.3 |

Each aggregate retained one fixture builder, four clones and zero scheduler
overlap for `ac043_large_grid_snapshot_v1`. Its declared measurement artifact is
`browser-e2e-measurement/frontend-measurement-aggregate.json` inside each run.

GNU `time` measured each public Make observer invocation and its descendants.
CPU totals include command startup/preflight. RSS is the largest per-invocation
peak, not simultaneous aggregate memory. Physical read bytes are filesystem
input blocks multiplied by 512; cached logical reads are not measured.

| Readers | Status calls | User CPU (s) | System CPU (s) | Maximum RSS (KiB) | Physical read bytes | Reader failures |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| 1 | 7 | 3.25 | 2.55 | 91,148 | 0 | 0 |
| 4 | 845 | 382.72 | 370.18 | 92,240 | 0 | 0 |

The four-reader workload used about two CPU cores on average over the invocation.
Zero physical disk reads reflect warm filesystem caching, not zero read volume
or zero interference. This single-run matrix qualifies the recorded host,
selection, reader cadence and source sizes; it does not establish a universal
performance guarantee or justify higher polling rates.

The earlier exploratory `test-ops-measure-4` passed tests, but its API-side
collector discarded stdout chunks and stopped two loops at their fourth calls.
It is excluded from four-reader qualification. The accepted repeat used one
native job, retained full per-call stdout/resource records, and joined all
readers after the measurement's native exit. No harness repair or threshold
change was used to make qualification pass.

Raw resource records and the consolidated noncanonical report are retained in
`.cartulary/retained-work/test-ops-qualification/qualification-summary.json`,
with accepted four-reader records under `four-native-zVNFZI/` beside it. They
are qualification instrumentation, separate from canonical test evidence.

## Preservation

The preexisting workbook handoff edit at
`docs/handoffs/ui-ux/workbook-workbench-visual-refresh.md` is unchanged by this
work (442 added lines and one removed line at initial inspection).
