# Appendix H: Operating-Model Supporting Guidance

This appendix is **non-normative**.

It describes recommended operator practice for using the current coordination surfaces.
It does not define implementation conformance. When it differs from Core 00 through Core 04, Core 00 through Core 04 govern.

## 1. Purpose and scope

This appendix captures operating-model guidance that belongs outside normative product behavior. It is intended for tracker hygiene, companion findings-document discipline, handoff quality, status-review cadence, workload redistribution, debrief follow-through, and challenge or escalation practice.

The current product contract remains centered on workbook-native coordination surfaces. This appendix describes how teams can use those surfaces well without turning routine row capture into a ritualized workflow.

## 2. Coordination surface map

Use the standardized workbook-native surfaces below as the durable coordination layer:

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
- Prefer linking to existing task, decision, party, and evidence records over repeating the same information in free text.
- Treat `next_report_at`, `next_checks`, and similar forward-looking fields as coordination prompts, not as substitutes for team judgment.

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
