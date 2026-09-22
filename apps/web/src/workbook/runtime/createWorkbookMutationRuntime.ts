import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import type { PendingReplayScope } from "./pending/workbookPendingQueue";
import { assembleWorkbookMutationFeatures } from "./WorkbookMutationFeatureAssembly";
import { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";
import {
  browserWorkbookRuntimeDependencies,
  type WorkbookRuntimeDependencies,
} from "./workbookRuntimePorts";

export function createWorkbookMutationRuntime(
  scope: PendingReplayScope,
  transactionIds: SecureTransactionIdPort,
  pendingMutationPort: WorkbookPendingMutationPort,
  dependencies: WorkbookRuntimeDependencies = browserWorkbookRuntimeDependencies,
): WorkbookMutationRuntime {
  return new WorkbookMutationRuntime(
    scope,
    transactionIds,
    pendingMutationPort,
    assembleWorkbookMutationFeatures,
    dependencies,
  );
}
