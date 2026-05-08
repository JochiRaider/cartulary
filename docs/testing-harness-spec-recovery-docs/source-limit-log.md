---
doc_id: THR-S0-SOURCE-LIMITS
title: Testing Harness Recovery Source-Limit Log
status: active
role: source-limit-log
---

# Testing Harness Recovery Source-Limit Log

## Document role

This log records unavailable, incomplete, or intentionally uninspected sources
during testing harness reverse-specification recovery. Do not delete unresolved
rows. Mark resolved limits with a follow-up note instead.

## Source limits

| Source-limit ID | Surface | Limit type | What was inspected | What was not inspected | Impact | Follow-up needed | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|
| SL-0001 | `.github/**` | `inaccessible_file` | Repository root was checked with `test -d .github && find .github -maxdepth 3 -type f \| sort || printf '.github absent\n'`. | CI workflow files under `.github/**`; directory was absent. | S0 cannot record repository CI workflow behavior from `.github` sources. Later sprints must classify CI as absent, external, or located outside `.github` if evidence appears. | During S1/S2, search for any alternate CI or release validation configuration and update this row if CI evidence is found. | `source_limit` | Required S0 input named CI configuration, but no `.github/` directory exists in the working tree. |
