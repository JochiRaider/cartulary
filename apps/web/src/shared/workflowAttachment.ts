export type WorkflowAttachment<Binding> = {
  update: (binding: Binding) => boolean;
  detach: () => void;
};
export type AttachWorkflow<Binding> = (
  binding: Binding,
) => WorkflowAttachment<Binding> | null;

/** Owns presentation identity only. Feature adapters supply authority and policy. */
export function createWorkflowAttachment<Binding>(
  admit: (binding: Binding) => (() => boolean) | null,
  apply: (binding: Binding | null) => void,
) {
  let generation = 0;
  const attach: AttachWorkflow<Binding> = (binding) => {
    const current = admit(binding);
    if (!current) return null;
    const token = ++generation;
    const live = () => token === generation && current();
    apply(binding);
    return {
      update: (next) => {
        if (!live() || !admit(next)) return false;
        apply(next);
        return true;
      },
      detach: () => {
        if (!live()) return;
        ++generation;
        apply(null);
      },
    };
  };
  return {
    attach,
    invalidate: () => {
      ++generation;
    },
  };
}
