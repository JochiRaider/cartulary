import type { RecordLifecycleAccepted } from "../mutations/workbookMutationCommandPorts";

export type WorkbookRecordHistoryOwnerEffects = {
  readonly deleteAccepted: (
    accepted: RecordLifecycleAccepted,
  ) => Promise<void> | void;
  readonly restoreAccepted: (
    accepted: RecordLifecycleAccepted,
  ) => Promise<void> | void;
  readonly rollbackAccepted: (
    accepted: RecordLifecycleAccepted,
  ) => Promise<void> | void;
};
