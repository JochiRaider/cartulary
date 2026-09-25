import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import type { WorkbookMutationFeatures } from "./WorkbookMutationFeatureAssembly";

export type WorkbookLifecycleContribution = {
  readonly unsettledMutationCount: number;
  setAuthority(authority: WorkbookMutationAuthority | null): void;
  suspend(): void;
  closeIncident(): void;
  retire(): void;
  subscribe(listener: () => void): () => void;
};

/** Fixed feature membership, with source-specific behavior behind each contribution. */
export class WorkbookFeatureLifecycle {
  private readonly contributions: Record<
    keyof WorkbookMutationFeatures,
    WorkbookLifecycleContribution
  >;
  private retired = false;

  constructor(
    features: WorkbookMutationFeatures,
    private readonly timeline: Readonly<
      Record<
        "actions" | "mentions" | "mutations",
        Omit<WorkbookLifecycleContribution, "subscribe">
      >
    >,
  ) {
    this.contributions = {
      ...features,
      explicitPatches: {
        get unsettledMutationCount() {
          return features.explicitPatches.unsettledMutationCount;
        },
        setAuthority: (authority) =>
          features.explicitPatches.setAuthority(authority),
        suspend: () => features.explicitPatches.suspend(),
        closeIncident: () => {
          const authority = features.explicitPatches.getSnapshot().authority;
          if (authority)
            features.explicitPatches.setAuthority({
              ...authority,
              closed: true,
            });
        },
        retire: () => features.explicitPatches.retire(),
        subscribe: (listener) => features.explicitPatches.subscribe(listener),
      },
    };
  }

  get unsettledMutationCount(): number {
    return (
      Object.values(this.contributions).reduce(
        (total, owner) => total + owner.unsettledMutationCount,
        0,
      ) +
      this.timeline.actions.unsettledMutationCount +
      this.timeline.mentions.unsettledMutationCount +
      this.timeline.mutations.unsettledMutationCount
    );
  }

  subscribe(
    changed: () => void,
    effects: Partial<Record<keyof WorkbookMutationFeatures, () => void>>,
  ): () => void {
    const cleanups = Object.entries(this.contributions).map(([key, owner]) =>
      owner.subscribe(() => {
        if (this.retired) return;
        effects[key as keyof WorkbookMutationFeatures]?.();
        changed();
      }),
    );
    return () => {
      for (const cleanup of cleanups) cleanup();
    };
  }

  setAuthority(authority: WorkbookMutationAuthority | null): void {
    if (this.retired) return;
    for (const owner of [
      ...Object.values(this.contributions),
      ...Object.values(this.timeline),
    ])
      owner.setAuthority(authority);
  }

  suspend(): void {
    if (this.retired) return;
    for (const owner of [
      ...Object.values(this.contributions),
      ...Object.values(this.timeline),
    ])
      owner.suspend();
  }

  closeIncident(): void {
    if (this.retired) return;
    for (const owner of [
      ...Object.values(this.contributions),
      ...Object.values(this.timeline),
    ])
      owner.closeIncident();
  }

  retire(): void {
    if (this.retired) return;
    this.retired = true;
    for (const owner of [
      ...Object.values(this.contributions),
      ...Object.values(this.timeline),
    ])
      owner.retire();
  }
}
