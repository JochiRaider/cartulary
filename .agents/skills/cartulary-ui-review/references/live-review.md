# Live review

Read this for seeded/dev interaction and page capture. Read the matching section
of the [workflow guide](../../../../docs/guides/cartulary_browser_design_readiness_workflow.md#agent-facing-private-review)
only when additional setup detail is needed.

## Start the selected mode

Use `make doctor` when readiness is unknown or startup identifies a dependency
problem. `make bootstrap` owns dependency installation when setup is within the
user's task. Ordinary review commands never install tools; do not substitute an
ambient browser or raw Playwright script.

For reproducible review after explicit setup, start `make ui-review` directly;
a smoke run is not a prerequisite. Seeded startup validates installed prerequisites
and builds current-source artifacts without installing tools or pulling images.
Failures provide closed phase/subject/condition/recovery fields in v2 results and
receipts. Drain active sessions by exact locator before upgrading or rolling back.

Start `make ui-review`. Select `REVIEW_PROFILE` explicitly
only when the task requires a different profile; consult the guide for current
profile meaning. Source edits require stopping and starting a new seeded session
so the review covers a new sealed build.

For an explicitly selected running development app, use:

```bash
make ui-review UI_MODE=dev UI_ORIGIN=http://127.0.0.1:5173
```

Replace the port with the selected app's actual port. The origin must satisfy the
current loopback contract; do not guess one or fall back to another app. Dev mode
does not accept `REVIEW_PROFILE`. It must not reset, migrate, or stop borrowed
services. UI actions can still change borrowed application data; limit them to
the user's authorized task.

Keep the foreground exec session open. Use a second shell for finite commands.
In the examples below, `review_session` is the exact printed locator and
`review_request` is the absolute path to the private JSON request you wrote.
Shell variables do not automatically persist between separate tool invocations;
carry the resolved paths explicitly.

```bash
CARTULARY_OUTPUT_MODE=machine make ui-review-status UI_SESSION="$review_session"
```

Proceed when the session is ready. Use the returned epoch instead of assuming
that startup or another caller left it at zero.

## Observe, target, act

Read the selected action variant in the
[action schema](../../../../tools/schemas/cartulary.ui_review_action.v2.schema.json).
A snapshot request for a session whose current epoch is zero is:

```json
{"schema_id":"cartulary.ui_review_action.v2","expected_epoch":0,"action":"snapshot","parameters":{}}
```

Replace `expected_epoch` with the actual current value, then submit:

```bash
CARTULARY_OUTPUT_MODE=machine make ui-browser UI_SESSION="$review_session" UI_REQUEST="$review_request"
```

Read the observations through returned `private_refs`. Prefer an observed exact
test ID or role/name pair; use session element references when necessary. Consult
the [target definition](../../../../tools/schemas/cartulary.ui_review_bundle.v1.schema.json#/$defs/target)
for its closed variants. Zero or multiple matches require refining the target;
there is no first-match, force-click, or arbitrary JavaScript escape hatch.
Snapshots replace references, and admitted actions invalidate them even when an
action fails. Use returned epochs and newly observed references, not arithmetic
guesses. Recovery after uncertain effects requires a fresh snapshot.

For seeded role review, use the `authenticate` action's actor variant. It is
unavailable in dev mode. Do not extract credentials or manufacture authentication
state. Navigate and interact only as needed to reach the requested state.

## Capture and assess

Read the page variant of the
[capture schema](../../../../tools/schemas/cartulary.ui_review_capture_request.v1.schema.json).
The minimal example below assumes current epoch zero; replace that value:

```json
{"schema_id":"cartulary.ui_review_capture_request.v1","source":"page","expected_epoch":0}
```

```bash
CARTULARY_OUTPUT_MODE=machine make ui-capture UI_SESSION="$review_session" UI_REQUEST="$review_request"
```

Capture observes rendered content without changing focus or scrolling. Reveal
virtualized content with an explicit action before capturing. Add only relevant
targets or scope; do not invent canonical bindings. Load detailed geometry from
the [observations schema](../../../../tools/schemas/cartulary.ui_review_observations.v1.schema.json)
only when resolving an actual coordinate question.

Page capture includes axe by default. Use `include_axe: false` only when the task
does not need a fresh accessibility scan and explain any resulting coverage
limitation. Distinguish completed violations/incomplete results from disabled,
unavailable, and execution failure; frames are excluded. Automated scans do not
replace keyboard, focus, or canonical contrast checks.

Resolve an image through the returned bundle manifest's component reference,
verify its bytes/digest, and open it before drawing visual conclusions. Load
[artifact analysis](artifact-analysis.md) only if a crop, comparison, or report
would help. Finish through the cleanup path in `SKILL.md`.
