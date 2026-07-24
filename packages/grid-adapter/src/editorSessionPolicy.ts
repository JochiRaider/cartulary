import type {
  GridCellAnchor,
  GridCellTarget,
  GridEditorActivation,
} from "./core";
import { sameGridCellAnchor } from "./semanticPresentation";

export type PendingEditorSeed = {
  readonly activation: GridEditorActivation;
  readonly anchor: GridCellAnchor;
  readonly baseRowVersion: number;
  readonly hasValue: boolean;
  readonly value: unknown;
};

export type ActiveEditorSession = {
  readonly cancel: () => void;
  readonly focus: () => void;
  readonly requestCommit: () => Promise<boolean>;
  readonly target: PendingEditorSeed["anchor"];
};

export function editorSeedForTarget(
  pending: PendingEditorSeed | null,
  target: GridCellTarget,
) {
  if (
    pending === null ||
    pending.baseRowVersion !== target.mutationIdentity.baseRowVersion ||
    !sameGridCellAnchor(pending.anchor, target)
  ) {
    return null;
  }
  return {
    activation: pending.activation,
    hasValue: pending.hasValue,
    value: pending.value,
  };
}
