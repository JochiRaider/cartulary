import { useLayoutEffect, useRef } from "react";
import type { AttachWorkflow, WorkflowAttachment } from "./workflowAttachment";

/** Only committed renders update ports; cleanup belongs to its own attachment. */
export function useWorkflowAttachment<Binding>(
  attach: AttachWorkflow<Binding> | undefined,
  binding: Binding,
  observeVisibility = false,
) {
  const committed = useRef(binding);
  const slot = useRef<{ handle: WorkflowAttachment<Binding> | null } | null>(
    null,
  );
  useLayoutEffect(() => {
    committed.current = binding;
  });
  useLayoutEffect(() => {
    const owned = { handle: attach?.(committed.current) ?? null };
    slot.current = owned;
    const visibility = () => owned.handle?.update(committed.current);
    if (observeVisibility)
      document.addEventListener("visibilitychange", visibility);
    return () => {
      if (observeVisibility)
        document.removeEventListener("visibilitychange", visibility);
      if (slot.current === owned) slot.current = null;
      owned.handle?.detach();
    };
  }, [attach, observeVisibility]);
  useLayoutEffect(() => {
    const owned = slot.current;
    if (!owned || !attach) return;
    if (owned.handle) owned.handle.update(binding);
    else owned.handle = attach(binding);
  });
}
