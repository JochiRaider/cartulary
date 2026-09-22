import { type ReactNode, useLayoutEffect, useState } from "react";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { WorkbookMutationRuntimeRegistry } from "../runtime/WorkbookMutationRuntimeRegistry";

/** Acquisition may retire a previous lifetime, so it is exclusively a commit effect. */
export function WorkbookMutationRuntimeBoundary({
  registry,
  incidentId,
  clientInstanceId,
  create,
  children,
}: {
  readonly registry: WorkbookMutationRuntimeRegistry;
  readonly incidentId: string;
  readonly clientInstanceId: string;
  readonly create: () => WorkbookMutationRuntime;
  readonly children: (runtime: WorkbookMutationRuntime) => ReactNode;
}) {
  const [binding, setBinding] = useState<{
    registry: WorkbookMutationRuntimeRegistry;
    runtime: WorkbookMutationRuntime;
  } | null>(null);
  useLayoutEffect(() => {
    const runtime = registry.acquire({ incidentId, clientInstanceId }, create);
    setBinding({ registry, runtime });
    // Presentation borrows this lifetime. Only its application registry retires it.
  }, [registry, incidentId, clientInstanceId, create]);
  if (
    binding?.registry !== registry ||
    binding.runtime.scope.incidentId !== incidentId ||
    binding.runtime.scope.clientInstanceId !== clientInstanceId
  )
    return null;
  return children(binding.runtime);
}
