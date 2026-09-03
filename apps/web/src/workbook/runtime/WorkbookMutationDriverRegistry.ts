import type { PendingReplayUnitState } from "../utils/workbookPendingQueue";

export type WorkbookMutationOwnerEnvelope =
  | {
      readonly kind: "managed_patch";
      readonly viewSchemaId: string;
    }
  | {
      readonly kind: "timeline_row";
      readonly viewSchemaId: string;
    };

export type WorkbookMutationOwnerKind = WorkbookMutationOwnerEnvelope["kind"];

export type WorkbookManagedPatchMutationDriver = {
  readonly kind: "managed_patch";
  readonly drain: (
    unit: PendingReplayUnitState,
    envelope: Extract<
      WorkbookMutationOwnerEnvelope,
      { readonly kind: "managed_patch" }
    >,
  ) => Promise<void>;
};

export type WorkbookTimelineRowMutationDriver = {
  readonly kind: "timeline_row";
  readonly drain: (
    unit: PendingReplayUnitState,
    envelope: Extract<
      WorkbookMutationOwnerEnvelope,
      { readonly kind: "timeline_row" }
    >,
  ) => Promise<void>;
};

export type WorkbookMutationDriver =
  | WorkbookManagedPatchMutationDriver
  | WorkbookTimelineRowMutationDriver;

export type WorkbookMutationDriverRegistration =
  | {
      readonly accepted: true;
      readonly status: "registered";
      readonly unregister: () => void;
    }
  | {
      readonly accepted: false;
      readonly status: "duplicate";
      readonly kind: WorkbookMutationOwnerKind;
    };

export type WorkbookMutationDriverDispatch =
  | { readonly status: "dispatched" }
  | {
      readonly status: "driver_absent";
      readonly kind: WorkbookMutationOwnerKind;
    }
  | { readonly status: "owner_absent" };

/** Routes the FIFO head to its one exact owner without replacing live drivers. */
export class WorkbookMutationDriverRegistry {
  #managedPatchDriver: WorkbookManagedPatchMutationDriver | null = null;
  #timelineRowDriver: WorkbookTimelineRowMutationDriver | null = null;
  readonly #owners = new Map<string, WorkbookMutationOwnerEnvelope>();

  register(driver: WorkbookMutationDriver): WorkbookMutationDriverRegistration {
    if (driver.kind === "managed_patch") {
      if (this.#managedPatchDriver !== null) {
        return { accepted: false, status: "duplicate", kind: driver.kind };
      }
      this.#managedPatchDriver = driver;
    } else {
      if (this.#timelineRowDriver !== null) {
        return { accepted: false, status: "duplicate", kind: driver.kind };
      }
      this.#timelineRowDriver = driver;
    }
    let active = true;
    return {
      accepted: true,
      status: "registered",
      unregister: () => {
        if (!active) return;
        active = false;
        if (
          driver.kind === "managed_patch" &&
          this.#managedPatchDriver === driver
        ) {
          this.#managedPatchDriver = null;
        } else if (
          driver.kind === "timeline_row" &&
          this.#timelineRowDriver === driver
        ) {
          this.#timelineRowDriver = null;
        }
      },
    };
  }

  claim(unitId: string, envelope: WorkbookMutationOwnerEnvelope): void {
    this.#owners.set(unitId, envelope);
  }

  release(unitId: string): void {
    this.#owners.delete(unitId);
  }

  envelope(unitId: string): WorkbookMutationOwnerEnvelope | null {
    return this.#owners.get(unitId) ?? null;
  }

  async drain(
    unit: PendingReplayUnitState,
  ): Promise<WorkbookMutationDriverDispatch> {
    const envelope = this.#owners.get(unit.id);
    if (envelope === undefined) return { status: "owner_absent" };
    if (envelope.kind === "managed_patch") {
      const driver = this.#managedPatchDriver;
      if (driver === null) {
        return { status: "driver_absent", kind: envelope.kind };
      }
      await driver.drain(unit, envelope);
    } else {
      const driver = this.#timelineRowDriver;
      if (driver === null) {
        return { status: "driver_absent", kind: envelope.kind };
      }
      await driver.drain(unit, envelope);
    }
    return { status: "dispatched" };
  }
}
