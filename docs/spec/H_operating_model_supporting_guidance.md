# Appendix H: Operating-Model Supporting Guidance

This appendix is **non-normative**.

It describes recommended operator practice for using the current coordination surfaces.
It does not define implementation conformance. When it differs from Core 00 through Core 04, Core 00 through Core 04 govern.

## 1. Purpose and scope

This appendix captures operating-model guidance that belongs outside normative product behavior. It is intended for tracker hygiene, companion findings-document discipline, handoff quality, status-review cadence, workload redistribution, debrief follow-through, and challenge or escalation practice.

The current product contract remains centered on workbook-native coordination surfaces. This appendix describes how teams can use those surfaces well without turning routine row capture into a ritualized workflow.

## 2. Coordination surface map

Use the standardized workbook-native surfaces below as the durable coordination layer:

- `cartulary.view.timeline.v2` for Timeline operational capture, including date entered, analyst, MITRE source text, device/object source text, IP address source text, UTC/local activity text, raw activity, activity synopsis, and data source source text.
- `cartulary.view.task_requests.v1` for Task Requests, including ownership, status, priority, workstream, due tracking, blocked work, and follow-through.
- `cartulary.view.decisions.v1` for Decisions, including owner, status, rationale, review state, and supersession.
- `cartulary.view.comm_log.v1` for Communications Logs, including audience, channel or meeting context, summary, referenced decisions, and action follow-up.
- `cartulary.view.handoff.v1` for Handoffs, including current state, open work, open risks, and next checks.
- `cartulary.view.status_review.v1` for Status Reviews, including blocked work, pending evidence, open decisions, risk summary, and next report timing.
- `cartulary.view.lesson.v1` for Lessons, including follow-up tasks, evidence references, and closure state.

Use user-facing labels after the first canonical `view_schema_id` reference when that improves readability, but keep the canonical surface identity available in runbooks, training, and screenshots.

## 3. Tracker hygiene

Recommended routine:

- Review unowned, blocked, stale, or overdue task-request rows at least once per shift or once per working day.
- Review unresolved mentions, duplicate-looking rows, timeline gaps, and incomplete evidence links before each status point.
- Keep timeline rows compact and queryable. Put long analyst narrative, excerpts, and exploratory notes in the companion findings document instead of overloading timeline cells.
- Treat RAW Activity and the other Timeline operational cells as preserved source text. Use inspector links, suggestions, tags, evidence, and canonical entity records for structure rather than changing source strings to make the row look cleaner.
- Prefer linking to existing task, decision, party, and evidence records over repeating the same information in free text.
- Treat `next_report_at`, `next_checks`, and similar forward-looking fields as coordination prompts, not as substitutes for team judgment.

## 3A. Inspector workflow practice

The inspector is useful for deeper row-context cleanup, but it should remain off the ordinary capture hot path.

Recommended practice:

- Use the inspector during review windows to resolve mentions, inspect relationship overflow, check row history, and clean up evidence links.
- Create task, decision, handoff, communications-log, or status-review records from selected rows when the selected row creates real follow-up work.
- Use Handoff inspector features to acknowledge receipt and verify linked open tasks, open decisions, open risks, and next checks.
- Use Status Review inspector features to review blocked work, pending evidence, open decisions, active risks, and next-report timing.
- Treat destructive confirmations, rollback previews, merge plans, and supersede actions as deliberate review actions, not as routine row-editing steps.

Do not require the inspector for ordinary Timeline, Host, Identity, Evidence, or Notes row creation, inline editing, paste, or rough capture.

## 4. Companion findings-document discipline

Use the workbook as the compact, queryable control surface for incident facts, relationships, and next actions.
Use the companion findings document for:

- longer-form analytic narrative,
- evidence excerpts,
- analyst reasoning that is still too rough for external release,
- draft report language,
- investigation dead ends worth retaining.

A practical split is:

- workbook: current state, identifiers, ownership, links, statuses, timestamps, and actionability;
- findings document: explanatory detail, quotations, screenshots, investigative context, and narrative synthesis.

When a fact becomes durable and operationally important, summarize it back into the workbook and link the longer narrative rather than leaving the workbook blind to it.

## 5. Incident-start and phase-change briefs

A useful incident-start or phase-change brief usually covers:

- current scope and confidence,
- immediate risks and containment constraints,
- named owners for active workstreams,
- required external or leadership updates,
- next decision points,
- the next scheduled review or report checkpoint.

Recommended supporting records:

- one status-review record for the shared checkpoint,
- one or more task-request rows for newly assigned work,
- decision rows for strategy or release decisions that should remain inspectable later,
- a communications-log record when the brief itself drives stakeholder commitments.

## 6. Shift handoff quality

A good handoff is specific enough that the next analyst can resume work without reconstructing the case from scratch.

Recommended minimum contents:

- `current_state_summary`
- `open_task_ids[]`
- `open_decision_ids[]`
- `open_risk_refs[]`
- `next_checks`
- an explicit acknowledgement or confirmation that the handoff was received

Recommended practice:

- capture one handoff record per real handoff boundary, not per routine row edit;
- link open tasks and decisions directly instead of summarizing them only in prose;
- call out blockers, assumptions, and time-sensitive checks explicitly;
- treat the handoff record as a continuity artifact, not a substitute for ordinary timeline capture.

## 7. Status-review cadence

A status review is the right place for coordination ritual. Routine row editing is not.

Recommended status-review contents:

- `blocked_task_ids[]`
- `pending_evidence_ids[]`
- `open_decision_ids[]`
- `active_risks_summary`
- `next_report_at`

Recommended practice:

- run status reviews at a cadence appropriate to incident tempo;
- use the review to rebalance work, surface blockers, and prepare stakeholder updates;
- prefer one well-linked status-review record over scattered reminder notes;
- update task ownership and due expectations during or immediately after the review.

## 8. Communications and stakeholder updates

Use Communications Logs for durable communication memory, especially when messages change scope, commitment, or expectations.

Recommended contents:

- audience or attendee party refs where available
- `channel_or_meeting`
- `summary`
- `decision_ids[]`
- `action_task_ids[]`
- `next_report_at`

Recommended practice:

- log stakeholder-impacting calls, emails, chat summaries, and decision-bearing meetings;
- capture what was communicated, what was asked for, and what follow-up is due;
- keep sensitive or draft narrative in the findings document until it is ready for curated release.

## 9. Workload redistribution and no-owner review

Use Task Requests and saved or system views to expose:

- no-owner work,
- blocked work,
- high-priority work,
- due or overdue work,
- workstream-specific queues,
- requester-specific or external-ticket follow-up.

Recommended practice:

- review no-owner and blocked queues explicitly during status reviews;
- redistribute work before owners become single-threaded bottlenecks;
- create a new task-request row when redistribution changes accountability in a durable way;
- prefer owner changes and linked tasks over informal chat-only delegation.

## 10. Challenge and escalation practice

Challenge and escalation belong in team practice, not in mandatory row-by-row product ritual.

Recommended practice:

- raise concerns explicitly when scope, evidence handling, containment, or release posture appears unsafe or incomplete;
- acknowledge the concern, assign an owner, and record the follow-up path in a decision, task request, status review, or communications log as appropriate;
- use selective second-person review for high-impact transitions such as destructive containment, externally consequential releases, or fragile evidence handoffs;
- avoid adding mandatory challenge checklists to ordinary timeline entry.

The goal is visible follow-through, not ceremony.

## 11. Debrief and lesson follow-through

Use Lesson records to capture what changed, what failed, what nearly failed, and what should be improved after the incident or after a major phase.

Recommended contents:

- concise lesson statement,
- affected workstream or area,
- `follow_up_task_ids[]`,
- `evidence_refs[]` where helpful,
- `closure_state`

Recommended practice:

- create lessons when there is a real process, tooling, communication, or coordination learning point;
- link each actionable lesson to one or more follow-up task requests;
- close the lesson only when the linked follow-up work is complete or explicitly retired;
- keep retrospective facilitation style in SOPs or training material rather than embedding it in the product contract.

## 12. Recovery operator runbook guidance

Core 01 and Core 04 own the recovery CLI contract. This section is only operator practice.

Recommended backup creation practice:

- invoke `operator backup create` through deployment-owned orchestration often enough to keep at least one successful retained backup no older than 24 hours;
- use a six-hour backup creation interval as the ordinary operational starting point unless the deployment has a stricter local recovery objective;
- preserve the final JSON result, any JSONL progress stream, and the encrypted recovery journal reference for every backup creation attempt;
- treat any failed candidate as diagnostic-only material that cannot be selected for restore, latest-backup inspection, or restore verification.

Recommended backup inspection practice:

- run `operator backup inspect latest` before any restore or verification drill;
- copy the exact selected `backup_set_id` from the final JSON result when preparing a real restore;
- treat a failed inspect result as a recovery readiness issue, not as permission to choose an older timestamp manually.

Recommended restore confirmation practice:

- use `operator restore latest --confirm-backup-set-id <exact-id>` with the ID selected by the latest inspect result;
- do not substitute an interactive `yes`, a timestamp, a local note, or a guessed backup identifier for the confirmation ID;
- keep the source and target configuration files visible in the runbook step so an operator can verify they are distinct before invocation.

Recommended restore-verification practice:

- keep an isolated target configuration ready for `operator restore-verify latest` and `operator restore-verify due`;
- mark verification targets explicitly and keep them outside ordinary traffic-serving paths;
- treat `restore-verify due` returning `no_op` as a successful cadence check only when the final JSON result is preserved.

Recommended failure recovery:

- if preflight fails with `unsafe_restore_target`, stop and reinitialize the target before retrying;
- if a restore or verification times out, leave the target not-ready and rebuild it from a fresh database and object namespace before reuse;
- if backup creation fails after staging artifacts, do not copy the diagnostic artifact reference into a restore command or runbook selector;
- preserve the final result JSON, any JSONL progress stream, the encrypted recovery journal reference, and the safe administrative-audit summary for incident-independent operational review;
- do not paste raw DSNs, bucket names, object keys, recovery keys, secret references, or incident content into runbook notes.

## 13. Administrative audit review guidance

Core 01, Core 02, and Core 04 own the administrative audit read projections,
storage invariants, authorization split, and redaction rules. This section is
only operating practice.

Recommended deployment audit review practice:

- use the deployment administrative-audit view for account administration, credential recovery, bootstrap, backup, and restore review;
- start with bounded timestamp filters and then narrow by actor, action code, target kind, or target identifier;
- use the server filters rather than free-text search, screenshots, exported spreadsheets, or copied database excerpts as the primary review path;
- treat redacted before/after values as evidence that a sensitive field changed, not as missing evidence;
- keep account-recovery runbooks free of passwords, password hashes, TOTP secrets, bootstrap tokens, session tokens, recovery keys, raw DSNs, object keys, object-store credentials, and storage secrets.

Recommended incident membership audit review practice:

- use the incident membership-audit view only from the addressed incident and only for membership create, role-change, and delete review;
- do not ask deployment administrators without incident membership to inspect membership activity for an incident they cannot otherwise access;
- preserve the incident identifier, applied filters, relevant event identifiers, and redacted-value status when recording review notes;
- keep membership audit review separate from incident content review unless the reviewer also has ordinary incident access for that content.

## 14. Incident-bundle import operations guidance

Core 01 owns incident-bundle import admission, final publication, initial importer membership, failure reasons, and terminal job results. Core 03 owns workbook startup after an imported incident is opened. This section is only operating practice.

Recommended import handoff practice:

- keep the importer account active and deployment-admin eligible until the import job reaches a terminal state;
- after success, use the **Open imported incident** action from Deployment administration rather than waiting for background completion to navigate away automatically;
- after opening the incident, add the intended local incident team through the ordinary incident membership route or UI;
- verify that historical actors imported as historical context only and are not mistaken for current incident members;
- record any team-admission notes in ordinary incident coordination surfaces once the imported incident is open.
