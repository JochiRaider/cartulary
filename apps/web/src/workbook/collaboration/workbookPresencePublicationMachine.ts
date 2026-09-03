const presenceDebounceMs = 150;
const maximumPresenceSettlementMs = 1_000;

export type WorkbookPresencePublicationMachine = {
  readonly generation: number;
  readonly pendingSinceMs: number | null;
};

export function initialWorkbookPresencePublicationMachine(): WorkbookPresencePublicationMachine {
  return { generation: 0, pendingSinceMs: null };
}

export function scheduleWorkbookPresencePublication(
  machine: WorkbookPresencePublicationMachine,
  nowMs: number,
): {
  readonly dueAtMs: number;
  readonly generation: number;
  readonly machine: WorkbookPresencePublicationMachine;
} {
  const generation = machine.generation + 1;
  const pendingSinceMs = machine.pendingSinceMs ?? nowMs;
  return {
    dueAtMs: Math.min(
      nowMs + presenceDebounceMs,
      pendingSinceMs + maximumPresenceSettlementMs,
    ),
    generation,
    machine: { generation, pendingSinceMs },
  };
}

export function settleWorkbookPresencePublication(
  machine: WorkbookPresencePublicationMachine,
  generation: number,
):
  | { readonly kind: "stale" }
  | {
      readonly kind: "publish";
      readonly machine: WorkbookPresencePublicationMachine;
    } {
  if (generation !== machine.generation || machine.pendingSinceMs === null) {
    return { kind: "stale" };
  }
  return {
    kind: "publish",
    machine: { generation, pendingSinceMs: null },
  };
}

export function cancelWorkbookPresencePublication(
  machine: WorkbookPresencePublicationMachine,
): WorkbookPresencePublicationMachine {
  return {
    generation: machine.generation + 1,
    pendingSinceMs: null,
  };
}
