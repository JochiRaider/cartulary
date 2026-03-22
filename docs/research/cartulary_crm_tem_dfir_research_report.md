
# Aviation CRM/TEM and Cartulary’s DFIR Workbook: A Non-Normative Research Report on Transfer, Fit, and Limits

## 1. Executive summary

This report adds value. Cartulary’s current corpus is already strong on workbook mechanics, low-friction capture, progressive structuring, projections, evidence objects, row-centric history, rollback, and collaboration. The thinner areas are coordination and operating-model concerns: authority gradients, challenge and escalation, handoffs, workload redistribution, shared situational awareness, communication cadence, explicit ownership, and lessons-learned follow-through.[^14]

**Direct evidence:** Recent aviation work still supports CRM as a disciplined package of non-technical skills centered on communication, decision-making, situational awareness, leadership/followership, workload management, and error management. Recent aviation accident analysis and safety-science work also shows that hierarchy and poor listening can suppress or neutralize challenge behavior, and that monitoring, role clarity, and workload distribution matter to safety outcomes.[^1][^2][^3][^4][^5][^6] Recent cyber and CSIRT literature points to parallel gaps: shared understanding of the incident and planned tasks, communication tooling for larger incidents, structured incident reports, visual overviews, escalation decisions, and better incident-learning practice remain uneven.[^8][^9][^10][^11]

**Inference:** The strongest transfer is not “CRM means generic collaboration.” It is narrower. Aviation CRM/TEM maps most cleanly to DFIR where teams must coordinate under uncertainty, time pressure, incomplete information, and uneven authority. The highest-value translated concepts are: authority-gradient management; safety voice plus safety listening; mutual monitoring for high-risk actions; workload redistribution and task shedding; briefing, handoff, and debrief discipline; and shared situational awareness through explicit owners, current status, and next actions.[^2][^3][^4][^8][^9]

**Recommendation:** These concepts justify a dedicated non-normative report because they sharpen exactly the parts of Cartulary that are comparatively thin, while leaving current core mechanics intact. The main product implications are lightweight and adjacent to the grid: candidate task/request, decision, communications-log, handoff, status-review, and lessons-learned surfaces; explicit owner fields on coordination objects; and saved/system views that surface risk, workload, and no-owner gaps. The main non-product implications belong in playbooks, SOPs, and coaching: challenge behavior, listening behavior, escalation discipline, briefing quality, and debrief practice.[^8][^9][^10][^11][^14]

**Recommendation:** CRM/TEM should remain non-normative for now. The evidence is strong enough to guide design and operating-model choices, but not strong enough to justify importing cockpit ritual into the workbook hot path. Cartulary should not require challenge-and-response on row edits, mandatory modal checklists on capture, or rigid pre-entry structure. The import should be selective, milestone-based, and artifact-supported, not ritualized row-by-row.[^2][^4][^6][^7][^14]

## 2. Method note and evidence posture

**Direct evidence:** Source selection prioritized recent peer-reviewed aviation and aircrew evidence, with emphasis on 2024-2026 material available as of March 21, 2026. Older canonical aviation sources were used only for stable concept definitions such as CRM, TEM, briefing, and monitoring. Adjacent transfer cases, especially the 2025 endoscopy CRM/TEM study, were treated as secondary corroboration rather than primary proof. Cyber and DFIR sources were used to test transferability rather than to replace aviation evidence.[^1][^2][^3][^4][^5][^6][^7][^8][^9][^10][^11][^12]

**Inference:** The evidence is strongest for broad operating-model ideas and weakest for direct software-object requirements. Aviation research can support claims such as “authority gradients degrade challenge behavior” or “briefing and monitoring improve team coordination,” but it cannot by itself prove that Cartulary must add a specific object or panel. Product implications in this report are therefore translations, not direct read-throughs.

**Evidence posture:** Conclusions are strongest where three conditions align: aviation evidence is direct, cyber evidence identifies a similar coordination problem, and the proposed implication can sit off the workbook hot path. Conclusions are tentative where one or more of those conditions fail. Handoffs in cyber are a good example: the operational need is clear, but direct cybersecurity handoff research is still thin, so the transfer should stay non-normative and validation-oriented.[^9][^13]

## 3. What CRM and TEM mean in aviation

### CRM

**Direct evidence:** CRM in aviation is a structured approach to using all available human and technical resources effectively. In current training and guidance, it centers on communication, situational awareness, teamwork, task allocation, decision-making, and error management rather than on stick-and-rudder technique alone.[^1][^5]

**Inference:** For Cartulary, the useful part of CRM is not aviation jargon. It is the discipline of making coordination observable and recoverable under stress.

### TEM

**Direct evidence:** TEM treats threats, errors, and undesired states as normal features of operations that crews must anticipate, detect, trap, and recover from. Threats are factors outside the crew’s immediate control that increase operational complexity; errors are crew actions or omissions; undesired states are the dangerous conditions that result when threats and errors are not managed.[^5]

**Inference:** In DFIR, the closest analogue is not “attack indicators.” It is the management of coordination risk: ambiguous facts, stale assumptions, overloaded analysts, weak handoffs, and premature decisions that can degrade incident handling even when the underlying technical analysis is sound.

### Relevant CRM/TEM concepts for DFIR

**Direct evidence:** The aviation concepts most relevant to this report are authority gradient, safety voice and safety listening, mutual monitoring, workload management and task shedding, briefing and debriefing, shared situational awareness, communication discipline, and threat/error trapping and recovery.[^1][^2][^3][^4][^5][^6]

**Inference:** These concepts matter to Cartulary only where they improve incident coordination without slowing rough capture. That boundary rules out literal import of many cockpit practices.

## 4. Transferability to DFIR

### Direct conceptual transfer

### Authority gradient, safety voice, and safety listening

**Direct evidence:** Aviation accident analysis still supports the old CRM concern that hierarchy can suppress challenge. Becker and Ayton link steep authority gradients to suppression of critical concerns, and Noort et al. show that flight crews spoke up in almost all accidents studied, yet poor safety listening still reduced the effectiveness of that voice and correlated with worse outcomes.[^3][^4]

**Inference:** This maps cleanly to DFIR. The failure mode is not only “junior analyst does not speak up.” It is also “concern is raised but not acknowledged, tracked, or acted on.” That is directly relevant to incident leads, senior responders, outside specialists, and vendor or client relationships.

**Recommendation:** Cartulary should support operating-model guidance and lightweight artifacts for challenge acknowledgement, escalation ownership, and decision visibility. This does not require a “challenge” button on every row. It does suggest that challengeable coordination objects such as tasks, decisions, evidence requests, and status reviews should have an owner, timestamp, and visible follow-through path.

**Judgment:** relevance high; likely layer playbook/training first with hybrid product support; hot-path risk low if attached to coordination objects and high if attached to ordinary row editing; evidence strength strong in aviation and moderate in cyber transfer; confidence high.

### Mutual monitoring and cross-check

**Direct evidence:** Aviation guidance treats monitoring and cross-checking as a last barrier against error, especially when workload is high. Becker and Ayton’s role-assignment study also suggests that concentrating command and control in the same person can degrade monitoring and increase systemic risk.[^3][^6]

**Inference:** The clean DFIR analogue is not “every row needs a second analyst.” It is narrower: high-impact containment, public statements, external release, risky scoping decisions, and fragile evidence handoffs benefit from an explicit second set of eyes or a watch function.

**Recommendation:** Product support should stay selective. Reviewer/history surfaces, high-risk decision artifacts, and evidence handoff records are plausible places for cross-check support. Routine capture and routine editing should remain single-operator.

**Judgment:** relevance high for selected actions and low for routine capture; likely layer hybrid; hot-path risk low if scoped to high-impact transitions and high if generalized to ordinary edits; evidence strength strong in aviation and moderate in cyber transfer; confidence moderate-high.

### Workload management and task shedding

**Direct evidence:** CRM and TEM treat workload management as a core safety skill. Cookson’s analysis of United 232 identifies workload management and expansion of the working team boundary as central to successful emergency coordination, while Becker and Ayton’s findings reinforce that role and workload distribution affect safety outcomes.[^2][^3]

**Inference:** This transfers well to DFIR because incident work is bursty, distributed, and often bottlenecked by a few analysts or leads. In cyber practice, CSIRTs report a need to handle workload peaks, stay aware of each other’s tasks, and use better communication/reporting tools during larger incidents.[^9]

**Recommendation:** Cartulary should treat workload as a coordination problem, not just a staffing problem. Candidate implications include explicit task/request objects, owner and status fields, saved views for no-owner or overloaded-work queues, and incident-level status-review artifacts that make redistribution visible.

**Judgment:** relevance high; likely layer hybrid with strong product support; hot-path risk low if surfaces stay adjacent to capture; evidence strength moderate-strong; confidence high.

### Briefing, handoff, and debrief discipline

**Direct evidence:** Aviation guidance treats briefings as a means to build mutual understanding, clarify expectations, and promote effective teamwork. Recent cyber guidance also stresses communication, reporting, and lessons learned, while CSIRT research identifies handoff, shared task awareness, structured reporting, and visual overviews as practical needs. Direct cyber handoff research remains limited, but current exploratory work treats the need as real.[^6][^8][^9][^13]

**Inference:** Incident-start briefs, phase-change briefs, shift handoffs, and debriefs map well to DFIR. Unlike cockpit checklists, these can occur at natural coordination boundaries rather than on the row-edit hot path.

**Recommendation:** Cartulary should not productize briefing behavior as a mandatory wizard. It should, however, support handoff and status-review artifacts that can hold current state, open tasks, unresolved questions, current risks, and next checkpoints. Debrief surfaces should connect lessons to follow-up tasks rather than ending as free text.

**Judgment:** relevance high; likely layer hybrid, with SOP/training leading and product support via artifacts; hot-path risk low because these are milestone actions; evidence strength moderate in aviation and cyber, weak-to-moderate for direct cyber handoff specifics; confidence moderate-high.

### Shared situational awareness and communication cadence

**Direct evidence:** Cookson’s United 232 analysis highlights continual updating of the shared mental model. NIST 800-61r3 emphasizes that incident response now depends on many internal and external parties and requires coordinated communication and reporting. Van der Kleij et al. likewise identify shared awareness of the incident and of ongoing and planned tasks as essential, especially in distributed teams.[^2][^8][^9]

**Inference:** Cartulary already provides part of the common operating picture through workbook projections and collaboration. The thinner gap is the coordination layer above the rows: who owns what, what is blocked, what changed since the last status point, and what must be said to whom next.

**Recommendation:** Product support belongs in saved/system views, status-review artifacts, communications logs, and decision/task surfaces, not in more mandatory fields on every timeline row.

**Judgment:** relevance high; likely layer hybrid with strong product support; hot-path risk low; evidence strength strong; confidence high.

### Transfer by analogy

### TEM as a coordination lens for threat, error, and recovery

**Direct evidence:** TEM in aviation is a model for recognizing that threats and errors are normal and that safety depends on detecting, trapping, and recovering from them rather than assuming perfect prevention.[^5] Ebert et al. make a related point for cybersecurity reporting by arguing for recognition of near misses and latent factors, not only critical incidents.[^10]

**Inference:** The closest Cartulary analogue is not a new per-row taxonomy. It is a design stance: incident work should preserve and expose open risks, near misses, weak controls, ambiguous states, and coordination failures early enough to recover from them.

**Recommendation:** Use TEM mainly as an operating-model lens for debriefs, lessons learned, near-miss capture, and status reviews. Do not force analysts to label every row as “threat” or “error.”

**Judgment:** relevance moderate-high; likely layer playbook/debrief first with light product support; hot-path risk medium if literalized into row schema; evidence strength strong for aviation concept and moderate for cyber analogy; confidence moderate.

### Team time-outs for high-risk transitions

**Direct evidence:** Aviation uses structured pauses and briefings around critical phases. In the 2025 endoscopy study, CRM/TEM-derived dialogic time-outs and communication guidelines improved communication, task clarity, and information transfer, while staff generally did not perceive the added steps as materially delaying work.[^7]

**Inference:** A DFIR analogue is plausible only at phase boundaries: destructive containment, evidence release, stakeholder briefing, or major strategy shift. It is not plausible on every timeline edit.

**Recommendation:** If Cartulary supports time-outs at all, they should be optional or playbook-driven milestone artifacts, not mandatory modal steps on the capture path.

**Judgment:** relevance moderate; likely layer SOP/training first with optional product support; hot-path risk low at milestones and extreme if moved onto row capture; evidence strength moderate and mostly transferred by analogy; confidence moderate.

### Role separation analogous to pilot flying versus pilot monitoring

**Direct evidence:** Becker and Ayton’s work suggests that combining control and monitoring in one cockpit role can degrade safety, particularly when hierarchy weakens intervention.[^3]

**Inference:** The closest DFIR fit is selective role separation on high-impact actions: one person drives, another watches or reviews. The analogy weakens fast outside those cases because incident teams are less standardized, more fluid, and not always paired.

**Recommendation:** Treat this as a bounded operating-model pattern for specific destructive or externally consequential actions, not as a general team structure.

**Judgment:** relevance moderate; likely layer SOP/reviewer workflow; hot-path risk medium if overused; evidence strength moderate; confidence moderate.

### Speculative or weak transfer

### Closed-loop challenge-response, standardized phraseology, and checklist discipline

**Direct evidence:** Challenge-response and phrase discipline are effective in tightly standardized flight operations, and CRM-derived dialogic time-outs can work in adjacent domains when adapted carefully.[^6][^7] Cookson also cautions that plain-language success in United 232 was context-specific and does not generalize automatically.[^2]

**Inference:** These practices do not transfer cleanly to everyday DFIR capture. Incident work is less standardized, less role-stable, and more heterogeneous than cockpit operations. Literal import would likely create ceremony, not clarity.

**Recommendation:** Keep these practices out of Cartulary’s default product surface. At most, adapt them narrowly for playbooks covering destructive containment, external release, or other rare, high-consequence transitions.

**Judgment:** relevance selective only; likely layer SOP/training; hot-path risk high; evidence strength weak-to-moderate for DFIR transfer; confidence high that routine import would be harmful.

### Transfer risks and domain mismatch

**Direct evidence:** Aviation CRM/TEM was developed for tightly standardized crews, explicit role separations, recurrent training, and highly scripted operating environments. Cyber incident response, by contrast, often involves ad hoc teams, distributed collaboration, and uneven tooling, while research on handoffs and social coordination in cyber remains comparatively immature.[^5][^8][^9][^13]

**Inference:** This is the central mismatch. Cartulary should borrow aviation’s coordination principles, not its ritual density. The product model can support shared awareness, ownership, handoff, and review. It cannot replace training in assertiveness, listening, escalation judgment, or debrief quality.

**Recommendation:** The right transfer posture is “support the behavior, do not simulate the cockpit.”

## 5. Where the current Cartulary corpus is strong already

**Direct evidence from the current corpus:** Cartulary already handles much of the mechanics that CRM/TEM is not needed to solve: workbook-shaped interaction, grid-first capture, progressive structure, denormalized projections over auditable source state, typed links, evidence objects, row-centric history, rollback, collaboration, and a deliberate separation between hot-path editing and deeper structure.[^14]

**Inference:** A CRM/TEM report therefore should not try to re-argue the grid, the relational substrate, or the history model. Its job is to sharpen the operating model and the adjacent surfaces that make teams safer and more coordinated around the grid.

## 6. Where CRM/TEM could sharpen the current corpus

### Authority gradients, challenge, and escalation behavior

**Direct evidence:** Aviation evidence is strongest here. Noort et al. show that speaking up alone is not enough; listening quality and hierarchy affect whether concerns change outcomes. CSIRT work also identifies escalation decisions as a real team-level need.[^4][^9]

**Inference:** Cartulary’s current corpus would benefit from a clearer non-normative operating model for when to challenge, when to escalate, who must acknowledge, and how unresolved concerns remain visible.

**Recommendation:** Candidate implications are an explicit escalation note or decision artifact, an owner on high-impact decisions, and saved views for unresolved or blocked items. This is one of the clearest areas where CRM adds value.

### Handoff discipline

**Direct evidence:** Cyber literature consistently points to the need for shared task awareness and structured reporting; direct cyber handoff evidence is still limited, but emerging work treats handoff as a real operational weak point.[^9][^13]

**Inference:** Cartulary’s backlog around ownership, task handling, and communications logs is directly relevant here. Handoff is not just a note; it is a bounded state transfer.

**Recommendation:** Candidate support includes a handoff artifact with outgoing owner, incoming owner, current picture, open tasks, open decisions, unresolved questions, and next checks. This belongs near tasks and decisions, not inside timeline-row capture.

### Workload redistribution

**Direct evidence:** CRM/TEM treats workload management as a first-order safety concern, and CSIRTs report the need to manage peaks, underload, and distributed coordination.[^2][^3][^9]

**Inference:** Cartulary’s current corpus could use a more explicit model for “who is carrying what.” Without that, the workbook can describe the incident but not the work allocation around it.

**Recommendation:** Candidate product support includes tasks or requests with owner and status, status-review artifacts, and saved views for overloaded or no-owner work queues.

### Shared situational awareness

**Direct evidence:** Aviation briefing, mental-model updating, and monitoring all point to a common principle: everyone solving the current problem needs the same essential picture. Cyber guidance and CSIRT interviews make the same point in different language.[^2][^6][^8][^9]

**Inference:** Cartulary already has the raw substrate for a common operating picture. It is thinner on the derived coordination picture.

**Recommendation:** Candidate implications include system views for current status, open risks, no-owner items, pending evidence, and current decisions; incident-level status-review artifacts; and communications logs for what was said externally and when.

### Communications cadence and status reporting

**Direct evidence:** NIST 800-61r3 explicitly treats incident coordination, notification, public communication, information sharing, and after-action reporting as distinct communication concerns. CSIRTs also report the need for better communication tools and structured reports during larger incidents.[^8][^9]

**Inference:** This is a direct fit for Cartulary’s thin areas. The workbook can hold facts, but status reporting requires explicit cadence, audience, and summary artifacts.

**Recommendation:** Candidate support includes a communications or meeting log, a status-review artifact, and lightweight views that surface “what changed since last status” without forcing analysts out of the grid during capture.

### Explicit ownership, task/request handling, and decision capture

**Direct evidence:** CRM focuses strongly on task allocation, role clarity, and joint problem solving. Cyber sources similarly stress shared understanding of current and planned tasks, structured reporting, and follow-through.[^2][^5][^8][^9]

**Inference:** Among the possible explicit work objects already visible in Cartulary’s backlog, tasks, requests, decisions, and ownership have the strongest CRM/TEM support. Hypotheses have some support through joint problem solving and shared mental models, but that support is weaker and more indirect.

**Recommendation:** If Cartulary promotes explicit work objects, tasks/requests and decisions should come before a dedicated hypotheses object. Ownership should attach first to those coordination objects and to evidence requests, not necessarily to every timeline row.

### Communications logs

**Direct evidence:** Aviation briefing/debrief practice and cyber incident communication guidance both point to the need for durable, audience-aware communication records.[^6][^8]

**Inference:** A communications log is a stronger CRM/TEM fit than a generic notes expansion because it directly addresses cadence, shared awareness, and handoff.

**Recommendation:** Candidate support includes timestamp, audience, summary, linked tasks or decisions, and optional next-report time. This is product-adjacent and low-risk to the hot path.

### Lessons-learned follow-through

**Direct evidence:** NIST recommends after-action reporting and continuous sharing of lessons learned. Patterson et al. show that cyber incident learning often remains superficial, defensive, or symptom-focused, while Ebert et al. argue for broader reporting that includes near misses and latent factors.[^8][^10][^11]

**Inference:** CRM/TEM adds a useful lens here because it emphasizes not only technical failure but also coordination failure, weak listening, poor monitoring, and missed recovery opportunities.

**Recommendation:** Cartulary’s corpus would benefit from an explicit debrief or lessons-learned surface tied to follow-up tasks and closure state, not just narrative notes.


**Inference:** Relative to the current Cartulary backlog, CRM/TEM gives the clearest support to explicit tasks or requests, decisions, ownership, communications logs, handoff records, status-review artifacts, and lessons-learned follow-through. It gives weaker support to a dedicated hypotheses object and almost no direct help on indicator-storage promotion, assessment vocabulary, or import-boundary questions.[^14]


## 7. Translation matrix

| CRM/TEM principle | Aviation meaning | DFIR analogue | Candidate artifact / UI / procedure | Product vs playbook / training | Hot-path risk | Evidence strength | Candidate normative implication |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Authority gradient and safety listening | Hierarchy can suppress challenge; listening quality determines whether challenge changes action | Junior or peripheral responders raise concerns that leads or specialists must acknowledge and resolve | Decision artifact, escalation note, explicit owner, unresolved-concern view, escalation SOP | Hybrid; training/SOP first | Low if attached to coordination objects; high if attached to every edit | Strong aviation; moderate cyber transfer | Candidate: explicit owner and acknowledgement path for high-impact decisions or escalations |
| Mutual monitoring and cross-check | Second crew member watches for drift, error, or missed cues | Reviewer/watch function on destructive containment, public statements, or evidence release | Reviewer surface, high-risk action review step, evidence handoff check | Hybrid | Low if selective; high if generalized | Strong aviation; moderate transfer | Candidate: optional second-person review for selected high-impact workflows only |
| Workload management and task shedding | Redistribute work to prevent overload and maintain safe control | Reassign tasks, requests, and follow-ups when lead or analyst load spikes | Task/request object, owner/status fields, overload or no-owner views, status review | Hybrid with strong product support | Low | Moderate-strong | Candidate: explicit task/request records with owner, status, due state, and linked records |
| Task allocation and role clarity | Crew coordination depends on clear task sharing and who is responsible for what next | Explicit owner on coordination objects, and limited approver/reviewer use on selected transitions | Owner field, reassignment action, selective approver/reviewer on high-impact workflows | Hybrid | Low if scoped; medium if generalized | Moderate | Candidate: explicit owner on tasks, evidence requests, decisions, and handoff records |
| Briefing discipline | Short, shared pre-phase plan clarifying risks, roles, and expectations | Incident-start and phase-change brief | Status-review artifact, playbook brief template | SOP/training first; product-adjacent | Low | Moderate | Candidate: incident-start and phase-change brief template with stated owner, priorities, and next report time |
| Handoff discipline | Structured transfer of current picture, open issues, and next actions | Shift handoff or analyst transfer during active incident | Handoff record, outgoing/incoming owner fields, open-task snapshot | Hybrid | Low | Moderate, with thin direct cyber evidence | Candidate: handoff artifact with current state, open tasks, open risks, and next checks |
| Shared situational awareness | Crew keeps a current mutual mental model of state, risks, and plan | Common operating picture of incident state, ownership, and pending actions | Saved/system views, status-review surface, communications log | Hybrid with strong product support | Low | Strong | Candidate: system views for no-owner, blocked, pending-evidence, and high-risk items |
| Communication cadence and status reporting | Briefings and updates keep all parties aligned on current state and intent | Internal status updates, stakeholder updates, meeting notes, decision trace | Communications or meeting log, “changed since last status” view | Hybrid | Low | Moderate-strong | Candidate: communications log with timestamp, audience, summary, and linked tasks/decisions |
| Debrief and lessons-learned follow-through | Post-flight learning closes the loop between event, error, and procedural change | After-action review linked to corrective work | Debrief surface, lesson record, follow-up task linkage | Hybrid; operating-model led | Low | Moderate-strong | Candidate: lessons-learned records linked to follow-up tasks and closure state |
| TEM threat/error/recovery framing | Anticipate threats, trap errors, recover before unsafe state persists | Surface near misses, latent conditions, coordination risks, and recovery points | Debrief template, near-miss field in review artifact, risk watchlist | Playbook/debrief first | Medium if literalized in-row | Strong aviation concept; moderate cyber analogy | Candidate: debrief artifacts that capture near misses, latent conditions, and recovery points without changing row schema |

## 8. Product-model versus playbook/training split

### Product model

**Recommendation:** The strongest product-model candidates are the ones that make coordination state explicit without changing the capture path. That includes tasks or requests, decisions, communications logs, handoff records, explicit owner fields on coordination objects, status-review artifacts, and lessons-learned surfaces. These are durable artifacts with clear identity, timestamps, and linkability to evidence, findings, and tasks. They fit Cartulary’s existing record-and-projection architecture and directly address the coordination gaps identified above.[^8][^9][^10][^11][^14]

**Inference:** A dedicated hypotheses object is a weaker product candidate than tasks or decisions. Joint problem solving and shared mental models support it indirectly, but the evidence is not as crisp, and hypotheses can often remain artifacts until repeated operational use justifies promotion.

### Playbook or SOP layer

**Recommendation:** Briefing cadence, escalation ladders, role expectations, challenge triggers, handoff protocol, time-out use for high-risk transitions, and debrief timing belong primarily in playbooks or SOPs. These are sequence and behavior rules, not durable case facts.

**Inference:** Cartulary can host the artifacts those procedures use, but it should not be the sole source of the behavior itself.

### Training or coaching

**Recommendation:** Assertive challenge, non-defensive listening, mutual monitoring habits, brevity and clarity in escalation language, and workload redistribution judgment should remain primarily training and coaching concerns. The product can surface whether an owner exists or a decision was recorded; it cannot teach a lead to listen well.

**Inference:** This is the most important boundary in the report. A product can support CRM/TEM behavior, but cannot substitute for it.

### Hybrid boundary

**Recommendation:** Handoff records, status-review artifacts, communications logs, and decision logs sit at the hybrid boundary. They are product-relevant because they need stable structure, timestamps, and linking. They are also procedure-relevant because their value depends on when and how teams use them.

**Recommendation:** The minimum hybrid rule is: product stores the artifact; playbook defines when it is used; training improves how well it is used.

## 9. Implications for specific Cartulary surfaces

### Timeline and workbook grid

**Recommendation:** Keep the grid mechanically unchanged. Do not add mandatory CRM/TEM fields, challenge prompts, or checklists to timeline-row capture. Use optional columns, adjacent saved views, and linked coordination objects instead. Possible low-risk additions are optional owner or follow-up indicators on selected views, but not mandatory row-level ritual.  
**Status:** hot-path compatible for optional indicators and linked objects; hot-path hostile for mandatory CRM/TEM entry fields.[^14]

### Inspector and resolution flows

**Recommendation:** The inspector is the right place for explicit owner transfer, linked decisions, challenge or escalation notes, handoff context, and selective review status on high-impact actions. It is also the right place to surface whether a concern was acknowledged or remains unresolved.  
**Status:** hot-path adjacent.

### Evidence request, attachment, and handoff artifacts

**Recommendation:** Evidence already has a natural lifecycle. CRM/TEM suggests sharpening its coordination side: requester, current owner, expected handoff, current status, linked decision or task, and whether receipt or review is blocked. Evidence handoff is one of the clearest places for selective cross-check.  
**Status:** hot-path adjacent, with compatible attachment mechanics.

### Saved views and system views

**Recommendation:** Saved and system views are the safest product surface for CRM/TEM ideas. Good candidates are “no owner,” “blocked task,” “pending evidence,” “open decision,” “unacknowledged escalation,” and “high-workload” views. These improve shared situational awareness without changing row-entry behavior.  
**Status:** hot-path compatible.

### Reviewer, history, and rollback surfaces

**Recommendation:** Reviewer/history surfaces can support selective mutual monitoring by making high-impact changes, evidence handoffs, and decision changes easy to inspect. This should remain targeted, not a generalized approval workflow on all edits.  
**Status:** hot-path adjacent.

### Communications or meeting-log surfaces

**Recommendation:** A communications log is a strong candidate surface. It should record timestamp, audience, summary, linked decisions or actions, and next-report timing where relevant. This directly supports cadence, status reporting, and handoff.  
**Status:** hot-path adjacent.

### Possible explicit work objects for tasks, hypotheses, decisions, and ownership

**Recommendation:** Promote tasks or requests, decisions, and ownership before promoting hypotheses. Tasks and decisions have the clearest CRM/TEM and cyber-operational support. Ownership should attach to those objects and to evidence requests or status reviews. Hypotheses should remain lower priority unless usage data shows a real need.  
**Status:** hot-path compatible if exposed as adjacent objects rather than mandatory row fields.

## 10. What should not be imported

### Practices that are too ceremonial

Do not import cockpit-style ceremony into routine capture. That includes mandatory challenge-response on ordinary edits, modal time-outs on row entry, or checklist completion before rough capture. The evidence does not support making routine DFIR documentation behave like pre-takeoff procedure.[^6][^7]

### Practices that assume cockpit crew structure too literally

Do not assume a stable captain/first-officer pair, a pilot-flying/pilot-monitoring split for all work, or tightly standardized recurrent training across all users. DFIR teams are more fluid, more distributed, and often cross organizational boundaries.[^3][^8][^9]

### Practices that would burden rough capture

Do not require analysts to classify each row as threat, error, undesired state, challenge, or decision before saving it. Do not require explicit owner, approver, or second-person acknowledgement on every timeline event. These would damage the non-negotiable hot path.[^5][^14]

### Practices that would create false precision

Do not create pseudo-rigorous CRM/TEM scoring fields, force confidence scales where no operational decision depends on them, or treat analyst communication quality as a product metric captured row by row. That would produce noise and compliance theater.

### Practices that make sense only in tightly standardized aviation operations

Do not import sterile-cockpit assumptions, standardized phraseology as a general rule, or universal challenge-response scripts. Selective use for narrow high-consequence playbooks may be reasonable. General use is not.

## 11. Failure modes and anti-goals

### Performative checklisting

The most obvious failure mode is adding visible ritual without changing coordination quality. A check box that says “brief complete” or “handoff done” is worse than useless if no shared picture, owner change, or next action actually becomes visible.[^7][^11]

### Over-structuring the capture path

Cartulary’s design objective is to beat spreadsheet failure modes without losing spreadsheet speed. CRM/TEM translation fails if it adds mandatory structure before analysts can capture the fact.

### Confusing product responsibilities with training responsibilities

A second failure mode is expecting the product to solve authority gradients, listening problems, or escalation judgment on its own. Those are operating-model and coaching problems first.

### Importing vocabulary without operational substance

Another failure mode is sprinkling aviation words into the corpus without clear contracts. “TEM,” “challenge,” or “briefing” only help if they point to a specific artifact, procedure, or validation experiment.

### Flattening important differences between aviation crews and incident teams

The last major anti-goal is domain flattening. Aircraft crews act in tightly standardized, temporally compressed, physically coupled systems. Incident teams do not. Any transfer that forgets that difference will over-ritualize routine work and underfit real incident variability.

## 12. Candidate normative implications

This section is non-authoritative. Each item below is a candidate only.

### Likely product-surface candidates

Candidate: support a task or request object with stable identity, owner, status, due state, linked records, and last-update time.

Candidate: support a decision record or decision artifact with timestamp, owner, short rationale, and support links to evidence, findings, events, or tasks.

Candidate: support a communications or meeting-log record with timestamp, audience, summary, linked actions, and optional next-report time.

Candidate: support a handoff record with outgoing owner, incoming owner, current state summary, open tasks, open risks, open decisions, and next checks.

Candidate: support saved/system views for no-owner items, blocked items, pending evidence, open decisions, and high-risk unresolved items.

Candidate: support a lessons-learned or after-action surface that links lessons to follow-up tasks and closure state.

Candidate: support explicit owner fields on coordination artifacts such as tasks, evidence requests, decisions, and handoff records, rather than on every timeline row.

### Likely operating-model candidates

Candidate: define a non-normative incident-start brief and phase-change brief that state incident lead, priorities, open risks, current unknowns, and next reporting time.

Candidate: define a non-normative escalation rule that requires an explicit owner and visible disposition for high-impact challenge items.

Candidate: define a non-normative shift-handoff practice using a stable handoff artifact rather than free-form chat alone.

Candidate: define a non-normative status-review cadence for major incidents, with one artifact per review rather than implicit state drift.

Candidate: define a non-normative debrief practice that records coordination failures, near misses, latent conditions, and follow-up actions.

Candidate: define a non-normative selective second-person review practice for destructive containment, externally consequential decisions, or evidence release, but not for routine row edits.

### Likely non-candidates

Candidate non-candidate: do not require mandatory CRM/TEM fields on timeline rows.

Candidate non-candidate: do not require challenge-response acknowledgement on ordinary workbook edits.

Candidate non-candidate: do not generalize cockpit role mirroring such as pilot-flying/pilot-monitoring to routine incident work.

Candidate non-candidate: do not make briefing or time-out wizards a prerequisite for rough capture.

Candidate non-candidate: do not force standardized aviation phraseology into general analyst communication.

Candidate non-candidate: do not create a generalized approval workflow for all edits under the banner of cross-checking.

## 13. Open questions and validation work

### Product experiments

Run a bounded experiment on explicit task/request objects versus artifact-only follow-up notes. The primary signals are whether ownership becomes clearer, stale work declines, and status reviews become easier to prepare.

Run a bounded experiment on a decision artifact for high-impact incident decisions. The key question is whether teams record decisions close to the moment they are made, and whether reviewer confidence in “why did we do this?” improves.

Run a bounded experiment on a handoff artifact for active incidents. Compare it with free-form notes or chat-based handoff and measure missed follow-ups, repeated questioning, and time to regain incident context.

Run a bounded experiment on a communications log and status-review surface. Measure whether stakeholder updates become easier to produce and whether analysts voluntarily keep the surface current.

### Observational studies

Observe live or simulated incident shifts and note when the team loses shared situational awareness, when challenges are raised but not acknowledged, when work becomes ownerless, and when handoffs create duplication or omission.

Observe whether analysts naturally create workaround artifacts for decisions, handoffs, or status reporting outside Cartulary. If they do, those are strong candidates for promotion into product surfaces.

### Interview prompts for analysts and incident leads

Ask: when during an incident do you most often realize the team is no longer looking at the same picture?

Ask: what information do you repeat at every handoff, and where do you currently store it?

Ask: which decisions create the most later confusion about rationale, ownership, or timing?

Ask: when a junior or peripheral responder raises a concern, what makes it more likely to be acknowledged and tracked?

Ask: which work items become stale because nobody clearly owns them?

Ask: which communication artifacts are produced outside the main workbook because the workbook is not the right place today?

### Signals that a concept should be promoted

Promote when teams voluntarily use the artifact without heavy enforcement, when stale or ownerless work declines, when handoff errors decline, when status updates become faster to prepare, and when reviewers can reconstruct decision rationale and follow-up more reliably.

Promote when the new surface remains adjacent to the grid and does not measurably slow rough capture.

### Signals that a concept should be rejected

Reject when users backfill the artifact after the fact, bypass it with chat or separate docs, or describe it as performative.

Reject when row-entry latency, paste behavior, or capture flow worsens.

Reject when the concept forces false precision, duplicates existing structures without better outcomes, or requires substantial coaching just to avoid obvious misuse.

## Sources

[^1]: Seda Ceken, "Non-Technical Skills Proficiency in Aviation Pilots: A Systematic Review," *International Journal of Aviation, Aeronautics, and Aerospace* 11(3), 2024. Used here for the recent review finding that technical skills alone are insufficient and that communication, decision-making, situational awareness, stress management, and leadership dominate current non-technical skill emphasis. citeturn601770view4turn601770view2

[^2]: Simon Cookson, "CRM in the Cockpit: An Analysis of Crew Communication in the Crash of United Airlines Flight 232," *Theoretical and Applied Ergonomics* 1(1), 2025. Used here for active participation in communication, shared mental-model updating, joint problem solving, team-boundary expansion, workload management, and the caution that communication patterns observed in that emergency do not generalize mechanically. citeturn744722view0turn744722view1turn744722view2

[^3]: T. Becker and P. Ayton, "Effects of flight crew role assignment on aviation accidents and incidents: Evidence of a systemic safety issue," *Safety Science* 171, 2024. Used here for evidence on role assignment, monitoring, status hierarchy effects, authority gradients, and the modern CRM/TEM training bundle. citeturn617018view0turn617018view1turn617018view2turn617018view4

[^4]: Merel C. Noort, Niels C. Reader, Alex Gillespie, "Safety voice and safety listening during aviation accidents: Cockpit voice recordings reveal that speaking-up to power is not enough," *Safety Science* 139, 2021. Used here for the distinction between speaking up and being listened to, the role of power distance, and the finding that poor listening reduced subsequent voice and was associated with worse outcomes. citeturn882324view1turn882324view2turn882324view3turn882324view4

[^5]: FAA, *Advisory Circular 120-51E: Crew Resource Management Training*; CASA / ICAO-aligned TEM guidance; and related official TEM materials. Used here for stable, older definitions of CRM and TEM and their current training scope. citeturn798089view2turn711398view0turn299532search4turn299532search6

[^6]: ICAO model advisory circular material on crew briefings and monitoring/cross-checking, plus Flight Safety Foundation checklist/CRM briefing notes. Used here for canonical briefing, monitoring, and challenge-response concepts and for the caution that these are deeply tied to standardized operations. citeturn107279view0turn107279view1turn107279view2turn107279view3turn798089view0turn798089view1

[^7]: Philipp Schweikart et al., "Crew resource management and threat and error management improve team communication in endoscopy: a prospective study," *Scientific Reports* 15, 2025. Used here only as secondary transfer evidence showing that adapted CRM/TEM interventions improved communication, task clarity, and information transfer outside aviation, with limited workflow cost. citeturn912750view0turn912750view2turn912750view3turn912750view4

[^8]: NIST SP 800-61 Revision 3, *Incident Response Recommendations and Considerations for Cybersecurity Risk Management*, April 2025. Used here for roles and responsibilities, playbooks, incident communication, after-action reporting, and continuous lessons-learned expectations. citeturn238993view0turn238993view1turn238993view2turn238993view3turn238993view4

[^9]: Rick van der Kleij, Kleinhuis, and Young, "Computer Security Incident Response Team Effectiveness: A Needs Assessment," *Frontiers in Psychology* 8, 2017. Used here for shared incident awareness, handoffs, workload variation, structured reporting, visual overviews, and the gap between ticketing/logging tools and larger-scale teamwork needs. citeturn955913view0turn378959view0turn378959view2turn378959view3

[^10]: Nico Ebert et al., "Learning from safety science: designing incident reporting systems in cybersecurity," *Journal of Cybersecurity* 11(1), 2025. Used here for near misses, latent factors, feedback, shared ownership/accountability, and incident-reporting design principles borrowed from safety science. citeturn848684view0turn848684view1

[^11]: Clare M. Patterson, Jason R. C. Nurse, and Virginia N. L. Franqueira, "\"I don't think we're there yet\": The practices and challenges of organisational learning from cyber security incidents," *Computers & Security* 139, 2024. Used here for the barriers to meaningful post-incident learning, including defensiveness, superficial reviews, and weak follow-through. citeturn137087view0turn137087view1

[^12]: J. Tilbury and S. Flowerday, "Humans and Automation: Augmenting Security Operation Centers," *Computers* 4(3), 2024. Used here as a caution that SOC design is often overly technical, that automation bias/complacency are real, and that ambiguous processes should remain under analyst oversight rather than being over-automated. citeturn965200view0turn965200view1turn965200view3turn965200view4

[^13]: Liberty Kent, Nilufer Tuptuk, and Ingolf Becker, "Passing the Baton: Shift Handovers within Cybersecurity Incident Response Teams," 2026 preprint. Used here only as emerging evidence that handoff is important in cyber incident response and that direct guidance remains limited. citeturn594685search0turn594685search1turn594685search3

[^14]: Cartulary internal corpus: *Cartulary Normative Core 00: Document Set Status, Precedence, and Conformance*; *Cartulary Normative Core 01: Architecture, Storage, and View Contracts*; *Cartulary Normative Core 02: Domain Model, Schema, and History*; *Cartulary Normative Core 04: Security, Deployment, and Conformance*; Appendix A; Appendix E; and *Spreadsheet of Doom (SoD): Architecture and Operating Model for Spreadsheet-Centric DFIR Case Management*. These sources establish the current Cartulary design boundary, strong areas, thinner areas, and current backlog context.
