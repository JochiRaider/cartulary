import type { AuthorizationRecoveryResult } from "../../shared/authorizationRecovery";

const authorizationRecoveryDelayMs = 1_000;

export type WorkbookAuthorizationRecoveryAdmission = {
  readonly generation: number;
};

export type WorkbookAuthorizationRecoveryMachine = {
  readonly authorizationConfirmed: boolean;
  readonly canResumeMutations: boolean;
  readonly generation: number;
  readonly phase:
    | "idle"
    | "scheduled"
    | "recovering"
    | "refreshing"
    | "terminal";
  readonly scheduledForMs: number | null;
};

export type WorkbookAuthorizationRecoveryResultPlan =
  | {
      readonly kind: "stale";
      readonly machine: WorkbookAuthorizationRecoveryMachine;
    }
  | {
      readonly kind: "retry";
      readonly machine: WorkbookAuthorizationRecoveryMachine;
    }
  | {
      readonly kind: "access_lost";
      readonly machine: WorkbookAuthorizationRecoveryMachine;
    }
  | {
      readonly admission: WorkbookAuthorizationRecoveryAdmission;
      readonly canResumeMutations: boolean;
      readonly kind: "authorized";
      readonly machine: WorkbookAuthorizationRecoveryMachine;
      readonly result: Extract<
        AuthorizationRecoveryResult,
        { readonly kind: "authorized" }
      >;
    };

export function initialWorkbookAuthorizationRecoveryMachine(): WorkbookAuthorizationRecoveryMachine {
  return {
    authorizationConfirmed: true,
    canResumeMutations: true,
    generation: 0,
    phase: "idle",
    scheduledForMs: null,
  };
}

export function scheduleWorkbookAuthorizationRecovery(
  machine: WorkbookAuthorizationRecoveryMachine,
  nowMs: number,
): WorkbookAuthorizationRecoveryMachine {
  if (machine.phase === "terminal") return machine;
  return {
    authorizationConfirmed: false,
    canResumeMutations: false,
    generation: machine.generation + 1,
    phase: "scheduled",
    scheduledForMs: nowMs + authorizationRecoveryDelayMs,
  };
}

export function beginWorkbookAuthorizationRecovery(
  machine: WorkbookAuthorizationRecoveryMachine,
  generation: number,
):
  | {
      readonly kind: "stale";
      readonly machine: WorkbookAuthorizationRecoveryMachine;
    }
  | {
      readonly admission: WorkbookAuthorizationRecoveryAdmission;
      readonly kind: "recover";
      readonly machine: WorkbookAuthorizationRecoveryMachine;
    } {
  if (machine.phase !== "scheduled" || machine.generation !== generation) {
    return { kind: "stale", machine };
  }
  return {
    admission: { generation },
    kind: "recover",
    machine: { ...machine, phase: "recovering", scheduledForMs: null },
  };
}

function admissionIsCurrent(
  machine: WorkbookAuthorizationRecoveryMachine,
  admission: WorkbookAuthorizationRecoveryAdmission,
  phase: "recovering" | "refreshing",
): boolean {
  return machine.phase === phase && machine.generation === admission.generation;
}

export function planWorkbookAuthorizationRecoveryResult(
  machine: WorkbookAuthorizationRecoveryMachine,
  admission: WorkbookAuthorizationRecoveryAdmission,
  result: AuthorizationRecoveryResult,
  nowMs: number,
): WorkbookAuthorizationRecoveryResultPlan {
  if (!admissionIsCurrent(machine, admission, "recovering")) {
    return { kind: "stale", machine };
  }
  if (result.kind === "unavailable") {
    return {
      kind: "retry",
      machine: scheduleWorkbookAuthorizationRecovery(machine, nowMs),
    };
  }
  if (result.kind === "access_lost") {
    return {
      kind: "access_lost",
      machine: {
        ...machine,
        authorizationConfirmed: false,
        canResumeMutations: false,
        phase: "idle",
        scheduledForMs: null,
      },
    };
  }
  const canResumeMutations =
    result.role === "editor" ||
    result.role === "reviewer" ||
    result.role === "admin";
  return {
    admission,
    canResumeMutations,
    kind: "authorized",
    machine: {
      ...machine,
      authorizationConfirmed: false,
      canResumeMutations,
      phase: "refreshing",
      scheduledForMs: null,
    },
    result,
  };
}

export function completeWorkbookAuthorizationRecovery(
  machine: WorkbookAuthorizationRecoveryMachine,
  admission: WorkbookAuthorizationRecoveryAdmission,
):
  | {
      readonly kind: "stale";
      readonly machine: WorkbookAuthorizationRecoveryMachine;
    }
  | {
      readonly canResumeMutations: boolean;
      readonly kind: "complete";
      readonly machine: WorkbookAuthorizationRecoveryMachine;
    } {
  if (!admissionIsCurrent(machine, admission, "refreshing")) {
    return { kind: "stale", machine };
  }
  return {
    canResumeMutations: machine.canResumeMutations,
    kind: "complete",
    machine: { ...machine, authorizationConfirmed: true, phase: "idle" },
  };
}

export function retryWorkbookAuthorizationRecovery(
  machine: WorkbookAuthorizationRecoveryMachine,
  admission: WorkbookAuthorizationRecoveryAdmission,
  nowMs: number,
): WorkbookAuthorizationRecoveryMachine {
  return admissionIsCurrent(machine, admission, "refreshing")
    ? scheduleWorkbookAuthorizationRecovery(machine, nowMs)
    : machine;
}

export function terminateWorkbookAuthorizationRecovery(
  machine: WorkbookAuthorizationRecoveryMachine,
): WorkbookAuthorizationRecoveryMachine {
  return {
    authorizationConfirmed: false,
    canResumeMutations: false,
    generation: machine.generation + 1,
    phase: "terminal",
    scheduledForMs: null,
  };
}
