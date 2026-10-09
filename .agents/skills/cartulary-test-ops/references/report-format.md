# Report format

Use only fields relevant to the task; do not generate a tracker or empty template.
Unknown, zero, passed and not applicable are different facts.

For a live update, identify the exact root/run/target and selection, native-session
observation, snapshot revision and progress cursor, known work-unit activity,
explicit waits, observed failures, and finalization state. Name limitations such
as absent capability, paused publication or truncated lists. Do not infer an ETA
or process death from unchanged publication.

For a terminal report, include:

- Exact command, run root and selected scope; source identity where relevant.
- Outer exit/signal, or not observed; owning summary and reported outcome.
- Normalized failure class/reason/code when available, separately from observer
  outcomes and the agent's causal hypothesis.
- Evidence actually inspected or audited, missing/incompatible facts, cached or
  unexecuted work, and the same-run cleanup result.
- Relationship to the change, the smallest supported next action, and any
  required verification not performed with its reason.

State "summary reports pass" when that is all that is established. Claim complete
invocation success only after the owning command finishes and the required
evidence and cleanup obligations are satisfied.

Write a durable handoff only when requested or required by the parent task.
Store it outside canonical evidence. Include a still-running session's native
handle and last observation time, exact locator, unresolved questions and next
supported command. Do not imply continued monitoring after the active task ends.
