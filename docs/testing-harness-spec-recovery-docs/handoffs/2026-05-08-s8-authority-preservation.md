---
doc_id: THR-S8-HANDOFF-2026-05-08
title: S8 Authority Preservation Handoff
status: active
role: sprint-handoff
---

# S8 Authority and Preservation Handoff

## Session metadata

| Field | Value |
|---|---|
| Sprint | S8: Authority and preservation |
| Status | `complete` for S4 follow-up routing scope |
| Repository root | `/home/askahn/code/cartulary` |
| HEAD revision | `947e6254d6fbc5154ad4691e0485f6e22e3153e1` |
| Timestamp recorded | `2026-05-08T22:16:39-04:00` |
| Recovery write boundary | `docs/testing-harness-spec-recovery-docs/**` |

## Outputs produced

| Output | Purpose |
|---|---|
| `preservation-matrix.md` | Classifies major S4 follow-up behaviors as preserve, clarify, authority-required, or likely out-of-contract. |
| `harness-authority-map.md` | Names owners and required decisions for reset route, package scripts, local-dev services, env contracts, platform support, stale janitors, and external caches. |
| `main-spec-conflict-list.md` | Lists product-spec conflict risks to avoid while drafting harness NLSpec text. |

## Open owner decisions

- App test runtime reset route visibility/security boundary.
- Direct package-script first-class contract status.
- Local-dev service lifecycle and persistent-state contract.
- Stale janitor destructive bounds and sufficient ownership proof.
- Public harness env-var set and precedence matrix.
- Supported platform/tool profile.
- External Go cache cleanup scope.

## Handoff to NLSpec drafting

NLSpec drafting may preserve Make as canonical command surface, source-limit
discipline, generated-output downstream status, scheduler-lane clarification,
and recovery-doc evidence labels. It must not normalize owner-required items
without a maintainer decision.

## Implementation-change audit

S8 changed recovery documentation only and made no implementation, service,
cleanup, fixture, generated, lockfile, or runtime-state changes.

