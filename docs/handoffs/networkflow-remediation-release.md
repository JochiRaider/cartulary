# Network Flow remediation release handoff

This release coordinates the internal Network Flow facade, Graph Projection
lease capabilities, Jobs restore pages, migration 44, and the OpenTelemetry
signal registry v2. Execution evidence and final readiness are controlled by
[the remediation tracker](networkflow-module-refactor-tracker.md).
Network Flow public major 6 and persisted state version 4 remain unchanged.

## Telemetry cutover

This is a `breaking_shape_change`. Remove alerts and dashboards using
`cartulary.network_flow.cleanup.eligible` and
`cartulary.network_flow.cleanup.oldest_eligible_result.age` at application cutover.
There are no aliases, approximate replacements under the old names, or dual reads.
Existing cleanup operation, duration and deletion instruments retain their shapes.

Configure these attribute-free signals from registry v2:

| Signal | Interpretation | Operational use |
| --- | --- | --- |
| `cartulary.network_flow.cleanup.examined` | Cumulative committed candidate examinations, including retained results | Compare progress with deletion rates; examination does not imply deletion |
| `cartulary.network_flow.cleanup.continuation` | Last successful dispatcher decision selected paced continuation (1) or base cadence (0) | Identify continued scanning; never interpret as backlog cardinality |
| `cartulary.network_flow.cleanup.last_success.age` | Monotonic elapsed seconds since the last successful sweep | Detect missing successful sweeps; correlate with operation outcomes and duration |

Freshness is absent until the process completes its first successful sweep.
Continuation starts at zero. Failures preserve the last successful continuation
observation; freshness continues aging. Counters retain only work with confirmed
commits, including progress before a later failure. An indeterminate commit is not
counted. No collection callback reads the database. Configure a startup grace
period and a missing-series alert for a claimed Network Flow runtime that never
records a successful sweep. Choose freshness thresholds against the five-minute
base cadence, five-second continuation, and thirty-second failure retry.

## Deployment and rollback

1. Review owner/projection agreement and the tracker's final acceptance evidence.
2. Apply migration 44 through the normal migration binary. It adds the Jobs partial
   index for positive nonterminal status, owner/profile, kind and ascending job ID.
3. Deploy the coordinated application and telemetry configuration. Restore still
   runs under Recovery's existing quiescence and atomic Graph transaction; pages
   of 256 have no deployment-wide total cap.
4. Verify successful cleanup, progress signals, job admission/materialization,
   historical receipt replay, exact-result reads and supported restore readiness.

Rollback pairs the previous application with its previous telemetry configuration.
The additive Jobs index is compatible with the previous application and may remain.
No receipt, result identity, job payload, graph declaration or state-format rewrite
is required. Migration catalog and evidence digests advance with the added index;
their evidence schemas and historical migration bytes remain unchanged. The internal Go API break is coordinated across repository callers;
production compatibility aliases are not part of the release.

Deployment, operator dashboard changes and operational cutover are external steps;
repository validation does not claim they have occurred.
