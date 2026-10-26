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
- For a multi-command acceptance portfolio, which obligations each invocation
  adds, why any overlap is necessary, and whether the union closes the requested
  scope on compatible source. Keep earlier-source results identified separately.
- For browser preparation failure, the consuming cohort/allocation and observed
  phase, cases not launched, quarantine and cleanup outcome. Do not report a
  reset attempt for fresh initialization or count preparation as a product test.

Report planning success as compilation, retained-run validity as evidence
admission, and complete-owner acceptance only for a passing audit of the required
compatible union. A child projection's presence alone establishes none of the
latter claims. Correct historical diagnostic mistakes without rewriting outcomes.

State "summary reports pass" when that is all that is established. Claim complete
invocation success only after the owning command finishes and the required
evidence and cleanup obligations are satisfied.

Write a durable handoff only when requested or required by the parent task.
Store it outside canonical evidence. Include a still-running session's native
handle and last observation time, exact locator, unresolved questions and next
supported command. Do not imply continued monitoring after the active task ends.
