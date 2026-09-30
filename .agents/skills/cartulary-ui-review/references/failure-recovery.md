# Failure recovery

Load this after a failure or uncertain outcome. Read the structured result's
`failures`, `state`, `epoch`, and `exit_code` before interpreting the shell status.
The [result schema](../../../../tools/schemas/cartulary.ui_review_command_result.v2.schema.json)
defines the fields. Make's generic exit code is not the underlying diagnosis.

## Choose recovery from the evidence

| Situation | Next step |
| --- | --- |
| Invalid request | Read the selected request schema and correct that input before retrying. Do not add unsupported fields or compatibility aliases. |
| Session or epoch mismatch | Inspect status using the exact locator. If still ready, obtain the current epoch and take a fresh snapshot before targeting again. Never attach to another session by discovery. |
| Missing or ambiguous target | Inspect fresh observations and select an exact target. Explicitly reveal virtualized content if the task requires it. |
| Action timeout or interrupted response | Effects may already have occurred. If the session survives, inspect a fresh snapshot and relevant state before deciding whether any further action is needed. |
| Unstable capture | Inspect the app state; wait for a known transition to settle or explicitly adjust the reviewed state, then take a bounded retry. Do not mutate during capture or retry indefinitely. |
| Missing dependencies or unavailable origin | Use `make doctor` for the reported prerequisite. Repair only within authorized setup scope; preserve the chosen mode and origin. |
| Unsafe/unsupported artifact or comparison | Resolve exact source identity, permissions, supported schema, and comparison intent. Do not rewrite metadata, bypass limits, or automatically downgrade provenance. |
| Capacity/resource contention | Check session status and reported ownership. Wait or stop this task's session as appropriate; do not kill unrelated owners. |
| Browser/controller loss or terminal failure | Invoke stop through the exact locator so the harness can recover owned resources. Inspect the terminal receipt before considering a fresh session. |
| Axe or image-analysis execution failure | Treat analysis as failed. Do not report an empty finding list as successful coverage. Inspect the diagnostic before deciding a justified retry. |

Retry only when the cause is understood and the next attempt changes the relevant
condition. Stop repeated attempts when the same condition remains unresolved;
report the blocker and complete all available cleanup. A failed operation may
consume an epoch or terminate the session, so do not assume the next request can
reuse either its target references or its prior state.

## Confirm terminal cleanup

Bind `review_session` to the original exact locator, then use:

```bash
CARTULARY_OUTPUT_MODE=machine make ui-review-stop UI_SESSION="$review_session"
```

Stop can interrupt busy work. Do not wait for a hung operation to finish before
requesting cancellation. Keep or poll the stop command's process until it
completes, then wait for the foreground startup process to exit. A subsequent stop
preserves the original terminal outcome; it cannot turn an earlier failure into
success.

Status/stop use minimal control readiness: damaged or absent browser/image/axe
packages do not require repair before stopping an exact session. Status is
observational. If both supervisors are lost, exact stop may recover a proven-dead
nonterminal session through its persisted resource handles. It never discovers or
terminates another session. Each resource owner applies its bounded teardown.
Safe detail is disposed even when another release fails; only unresolved recovery
proof remains. Age alone does not authorize deleting that proof or a live root.

A later exact stop may finish unresolved resource cleanup while preserving the
original failed terminal receipt byte-for-byte. A missing/invalid receipt is an
artifact failure, never successful cleanup evidence. Restart a drained session
after private layout changes; never reinterpret old proof or delete it for rollback.

Inspect the returned structural receipt using the exact run identity. The
[receipt schema](../../../../tools/schemas/cartulary.ui_review_receipt.v2.schema.json)
distinguishes cleanup completion from command outcome. `closed` or successful
status inspection alone does not establish successful cleanup. Preserve both the
primary failure and any cleanup failure in the handoff.

If safe receipt publication fails, report that cleanup evidence is unavailable.
If cleanup fails, preserve ownership records for diagnosis; do not manually
remove runtime roots, private service data, or leases to make the result appear
clean. Remove only your caller-owned request scratch files after their use ends.
Borrowed application services and imported originals remain with their owners.
