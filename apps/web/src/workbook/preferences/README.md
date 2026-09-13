# workbook/preferences/

[Parent](../README.md) · [Source overview](../../README.md)

Current-user home and incident-default workbook preference reads, edits, announcements, and recovery.

Home and incident-default reads have independent state. The owner retains
captured preference changes and review/recovery state; app composition binds its
lifetime and departure review through [app hooks](../../app/README.md).

## Files

| File | Responsibility |
| --- | --- |
| [WorkbookPreferenceController.ts](WorkbookPreferenceController.ts) | Home/default workbook preference reads, captured updates, acknowledgement, and recovery ownership. |
| [workbookPreferenceModel.ts](workbookPreferenceModel.ts) | Preference authority, home/default resources, drafts, operation results, and problems. |
| [WorkbookPreferencesPanel.tsx](WorkbookPreferencesPanel.tsx) | Preference controls, resource-pointer formatting, snapshots, and accessible outcome announcements. |

## Tests

| File | Responsibility |
| --- | --- |
| [workbookPreferenceCharacterization.test.tsx](workbookPreferenceCharacterization.test.tsx) | Tests explicit clearing and independent preference publication during delayed reads. |
| [WorkbookPreferenceController.test.ts](WorkbookPreferenceController.test.ts) | Tests independent home/default resources, set/clear operations, and preference admission. |
