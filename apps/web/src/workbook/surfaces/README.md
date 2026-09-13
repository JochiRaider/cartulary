# workbook/surfaces/

[Parent](../README.md) · [Source overview](../../README.md)

Registration-driven composition of concrete workbook surface renderers.

The facade selects a registered renderer by stable `view_schema_id` and adapts
owner snapshots and commands to its inputs. Concrete surface selection remains
at this composition boundary.

## Files

| File | Responsibility |
| --- | --- |
| [WorkbookSurfacesFacade.tsx](WorkbookSurfacesFacade.tsx) | Registration-driven active-surface selection and adaptation of view-state, query, mutation, collaboration, inspector, continuity, and layout owners into concrete renderer inputs. |
