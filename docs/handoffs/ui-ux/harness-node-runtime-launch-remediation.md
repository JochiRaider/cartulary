# Scheduler Node runtime launch remediation

This is a human handoff record. The adopted Testing Harness NLSpec owns the
behavior; this file is not an executable test, generation, or release input.

## Decision and change

The scheduler previously compiled logical `node` work commands and carried
`NODE_BIN` in the runtime environment, but the executor spawned bare `node`.
A non-interactive caller without Node on `PATH` could start the browser stack
and then fail to launch a test command.

The harness contract now binds logical graph `node` commands to the selected,
validated `NODE_BIN` at process launch. The scheduler validates that binary
before managed service startup. The same runtime directory leads `PATH` for
managed-suite setup and graph children, including after fixture and unit
environment overlays. Runtime paths do not replace commands in the work graph
or its semantic digest. Invalid or unlaunchable selected Node fails as
`configuration_error` with exit code `2`, without falling back to another Node.

This changes no public Make target, command ID, graph schema, retained artifact
path, database state, or product API. The generated topology render index was
refreshed with `make generate`; generated files were not edited by hand.
Shell startup files are not part of the remediation.

## Verification record

| Check | Result | Evidence |
| --- | --- | --- |
| `make generate` | Pass | `.cartulary/test-results/20260926T041932Z-p7939` |
| `make lint-markdown` | Pass | `.cartulary/test-results/20260926T044448Z-p87809` |
| `make lint-scripts` | Pass | `.cartulary/test-results/20260926T044440Z-p87387` |
| `make generate-drift` | Pass | `.cartulary/test-results/20260926T044423Z-p83447` |
| `make generated-artifact-policy-check` | Pass | `.cartulary/test-results/20260926T042032Z-p12423` |
| `make agent-finalize` | Pass | `.cartulary/test-results/20260926T044511Z-p89493`; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make harness-contract` | Pass | `.cartulary/test-results/20260926T044535Z-p93304` |
| `make harness-contract-tests` | Pass after fixture repair | `.cartulary/test-results/20260926T044307Z-p79286`; the direct Vitest diagnostics fixture now owns a suite runtime when none is supplied and uses the executing pinned Node on its child `PATH` |

An initial browser invocation used `PATH=/usr/bin:/bin`, which also omitted
the required Go launcher. Browser fixtures failed to start the backend with
`env: 'go': No such file or directory`; that run was cancelled and retained at
`.cartulary/test-results/20260926T042142Z-p20338`. The replacement invocation
uses a non-interactive `PATH` where `go` resolves and `node` does not.

The contract fixture covers direct and nested Node execution with Node-free and
competing-Node `PATH` values, a non-Node graph child invoking Node, invalid and
non-executable binary paths, a launch-time missing interpreter, and unchanged
logical command identity. Human review establishes owner-to-projection fidelity;
the tests do not read specification Markdown.

## Browser verdict

`make browser-e2e-webserver-backed` ran with a non-interactive caller `PATH`
where `go` resolves and `node` does not. The retained run at
`.cartulary/test-results/20260926T042905Z-p87597` finished with 137 of 140
units passing. Two browser groups failed product assertions, and the target
summary therefore failed as the third unit. Both failures are Timeline focus
continuity assertions in `mentions.resolve.spec.ts` and
`mentions.lifecycle.spec.ts`; their Playwright workers and nested Node
processes launched successfully. They do not indicate a Node runtime launch
failure. The exact Playwright stdout and traces are retained under the two
`functional-support-default-mentions-*` browser group directories in that run.
