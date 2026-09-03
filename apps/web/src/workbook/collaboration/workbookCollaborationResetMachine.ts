export type WorkbookCollaborationResetAdmission = {
  readonly eventGeneration: number;
  readonly generation: number;
  readonly sessionGeneration: number;
  readonly sheetKey: string;
};

export type WorkbookCollaborationResetMachine = {
  readonly active: WorkbookCollaborationResetAdmission | null;
  readonly generation: number;
};

export function initialWorkbookCollaborationResetMachine(): WorkbookCollaborationResetMachine {
  return { active: null, generation: 0 };
}

export function beginWorkbookCollaborationReset(
  machine: WorkbookCollaborationResetMachine,
  input: {
    readonly eventGeneration: number;
    readonly sessionGeneration: number;
    readonly sheetKey: string;
  },
): {
  readonly admission: WorkbookCollaborationResetAdmission;
  readonly machine: WorkbookCollaborationResetMachine;
} {
  const admission = {
    ...input,
    generation: machine.generation + 1,
  };
  return {
    admission,
    machine: { active: admission, generation: admission.generation },
  };
}

export function workbookCollaborationResetIsCurrent(
  machine: WorkbookCollaborationResetMachine,
  admission: WorkbookCollaborationResetAdmission,
  current: { readonly sessionGeneration: number; readonly sheetKey: string },
): boolean {
  return (
    machine.active === admission &&
    current.sessionGeneration === admission.sessionGeneration &&
    current.sheetKey === admission.sheetKey
  );
}

export function cancelWorkbookCollaborationReset(
  machine: WorkbookCollaborationResetMachine,
): WorkbookCollaborationResetMachine {
  return {
    active: null,
    generation: machine.generation + 1,
  };
}
