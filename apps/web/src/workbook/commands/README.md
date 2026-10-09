# workbook/commands/

[Parent](../README.md) · [Source overview](../../README.md)

Commands discovers currently attached owner actions and dispatches captured
semantic targets. It owns neither source workflows nor authorization decisions.
[Core 03 §2.5.2](../../../../../docs/spec/03_workbook_interaction_collaboration_and_workflows.md#252-commands)
defines matching, paging, target capture and invocation; §2.5.7 defines borrowed
editor focus and retained work. Design §§8.3A and 14 supply presentation direction.

| File | Responsibility |
| --- | --- |
| [workbookCommandIndex.ts](workbookCommandIndex.ts) | Descriptor/target types and pure deterministic metadata search. |
| [WorkbookCommands.tsx](WorkbookCommands.tsx) | Shell provider, private registration lifetime, owner contribution hook and Commands control. |
| [workbookCommandIndex.test.ts](workbookCommandIndex.test.ts) | Normalization, token matching, rank, family and stable identity ordering. |
| [WorkbookCommands.test.tsx](WorkbookCommands.test.tsx) | Captured-target revalidation, twenty-result pages, borrowed focus and concealment. |

`WorkbookCommandsProvider` creates one private index per mounted shell.
`useWorkbookCommand` registers stable descriptor metadata, rejects duplicate
identities and invokes the current attached owner's callbacks. Detachment removes
the contribution and fences captured callbacks. New commands supply a descriptor
and owner binding; they do not add feature execution to the index.

Opening captures the surface and semantic cell/selection. Invocation rechecks the
current surface, selection, read authority and owner availability. Search operates
on labels and terms, never protected record bodies. Matching and family/page
limits consume the existing design projection, without another registry.
Commands borrows editor focus; Escape restores eligible origin focus. Opening or
dismissing it does not submit or discard an editor or owner operation.

Navigation commands call the existing [navigation boundary](../navigation/README.md).
Source actions remain with their authoring/recovery owners. Uncertainty conceals
Commands and drops its captured presentation target; runtime owners decide which
drafts, receipts and session metadata survive.

Use `make task-guide ROLE=module-author OWNER=web.workbook` for current unit
routing and the `module.workbook` guide for responsive keyboard/accessibility
coverage. The [controlling handoff](../../../../../docs/handoffs/ui-ux/workbook-workbench-visual-refresh.md)
records scoped execution evidence; it is not an executable input.
