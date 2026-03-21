# Cartulary Document Set

This directory contains the derived document set for **Cartulary**, produced from the exploratory design artifact [`sod_replacement_exploratory_design_report.md`](appendices/G_source_archive_exploratory_design_artifact.md).

The document set is split into:

- a **normative core**, which is authoritative for closed decisions and current conformance,
- a set of **non-normative appendices**, which preserve rationale, diagrams, roadmap context, open questions, source extracts, and traceability.

## Precedence

When the normative core and a non-normative appendix differ, the normative core governs.

When an appendix records an unresolved question or a future option, that text does not change current conformance unless a normative core document explicitly says so.

## Normative core

1. [`normative/00_document_set_status_and_precedence.md`](normative/00_document_set_status_and_precedence.md)
2. [`normative/01_architecture_storage_and_view_contracts.md`](normative/01_architecture_storage_and_view_contracts.md)
3. [`normative/02_domain_model_schema_and_history.md`](normative/02_domain_model_schema_and_history.md)
4. [`normative/03_workbook_interaction_collaboration_and_workflows.md`](normative/03_workbook_interaction_collaboration_and_workflows.md)
5. [`normative/04_security_deployment_and_conformance.md`](normative/04_security_deployment_and_conformance.md)

## Non-normative appendices

A. [`appendices/A_problem_framing_rationale_tradeoffs_and_sanity_check.md`](appendices/A_problem_framing_rationale_tradeoffs_and_sanity_check.md)  
B. [`appendices/B_architecture_diagrams_and_explanatory_source_extract.md`](appendices/B_architecture_diagrams_and_explanatory_source_extract.md)  
C. [`appendices/C_schema_reference_and_ddl_source_extract.md`](appendices/C_schema_reference_and_ddl_source_extract.md)  
D. [`appendices/D_workflow_and_ui_illustrations_source_extract.md`](appendices/D_workflow_and_ui_illustrations_source_extract.md)  
E. [`appendices/E_roadmap_open_questions_and_decision_backlog.md`](appendices/E_roadmap_open_questions_and_decision_backlog.md)  
F. [`appendices/F_source_traceability_matrix.md`](appendices/F_source_traceability_matrix.md)  
G. [`appendices/G_source_archive_exploratory_design_artifact.md`](appendices/G_source_archive_exploratory_design_artifact.md)

## Notes on editorial treatment

The source artifact mixes current requirements, future roadmap items, and open questions in a few areas. To keep the normative core internally consistent while preserving all source information, this document set uses:

- **conformance profiles** for capabilities that the source described as important but not uniformly day-one,
- a **decision backlog appendix** for unresolved questions,
- a **source traceability matrix** so every source section lands in at least one derived target,
- a **source archive appendix** for complete preservation of the original artifact text.
