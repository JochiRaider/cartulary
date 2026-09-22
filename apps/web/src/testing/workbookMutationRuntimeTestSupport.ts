import { useLayoutEffect, useState } from "react";
import { WorkbookMutationRuntimeRegistry } from "../workbook/runtime/WorkbookMutationRuntimeRegistry";

/** Standalone shell tests own the application lifetime explicitly. */
export function useWorkbookMutationRuntimeTestRegistry(
  provided?: WorkbookMutationRuntimeRegistry,
) {
  const [owned] = useState(() => new WorkbookMutationRuntimeRegistry());
  useLayoutEffect(() => () => owned.dispose(), [owned]);
  return provided ?? owned;
}
