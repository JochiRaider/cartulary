export type WorkbookApplicationShortcutEvent = {
  readonly altKey?: boolean | undefined;
  readonly ctrlKey?: boolean | undefined;
  readonly key: string;
  readonly metaKey?: boolean | undefined;
  readonly shiftKey?: boolean | undefined;
};

export type WorkbookApplicationShortcutContext = {
  readonly capabilities: {
    readonly closeInspector: boolean;
    readonly history: boolean;
    readonly linkedEvidence: boolean;
    readonly quickLink: boolean;
  };
  readonly focusOwner:
    | "editor"
    | "grid_navigation"
    | "inspector"
    | "menu"
    | "overlay";
  readonly previewableEvidenceCount: number;
  readonly rowKind: "committed" | "draft" | "group" | "none";
  readonly selectionIdentity: string | null;
};

type ConsumedApplicationShortcut = {
  readonly preventDefault: true;
  readonly stopPropagation: true;
};

export type WorkbookApplicationShortcutDecision =
  | ({ readonly kind: "quick_link" } & ConsumedApplicationShortcut)
  | ({
      readonly destination: "list_or_empty" | "sole_previewable_item";
      readonly kind: "preview_linked_evidence";
    } & ConsumedApplicationShortcut)
  | ({ readonly kind: "open_history" } & ConsumedApplicationShortcut)
  | ({ readonly kind: "close_inspector" } & ConsumedApplicationShortcut)
  | {
      readonly kind: "none";
      readonly preventDefault: false;
      readonly stopPropagation: false;
    };

const noApplicationShortcut = {
  kind: "none",
  preventDefault: false,
  stopPropagation: false,
} as const;

export function decideWorkbookApplicationShortcut(
  event: WorkbookApplicationShortcutEvent,
  context: WorkbookApplicationShortcutContext,
): WorkbookApplicationShortcutDecision {
  const hasCommandModifier = event.ctrlKey === true || event.metaKey === true;
  const hasCommittedGridSelection =
    context.focusOwner === "grid_navigation" &&
    context.rowKind === "committed" &&
    context.selectionIdentity !== null &&
    context.selectionIdentity.length > 0;

  if (
    event.key === "Escape" &&
    context.capabilities.closeInspector &&
    context.focusOwner !== "menu" &&
    context.focusOwner !== "overlay"
  ) {
    return {
      kind: "close_inspector",
      preventDefault: true,
      stopPropagation: true,
    };
  }

  if (
    hasCommandModifier &&
    event.altKey !== true &&
    event.key.toLowerCase() === "k" &&
    hasCommittedGridSelection &&
    context.capabilities.quickLink
  ) {
    return {
      kind: "quick_link",
      preventDefault: true,
      stopPropagation: true,
    };
  }

  if (
    !hasCommandModifier &&
    event.altKey === true &&
    event.key.toLowerCase() === "h" &&
    hasCommittedGridSelection &&
    context.capabilities.history
  ) {
    return {
      kind: "open_history",
      preventDefault: true,
      stopPropagation: true,
    };
  }

  if (
    !hasCommandModifier &&
    event.altKey !== true &&
    event.key === " " &&
    hasCommittedGridSelection &&
    context.capabilities.linkedEvidence
  ) {
    return {
      destination:
        context.previewableEvidenceCount === 1
          ? "sole_previewable_item"
          : "list_or_empty",
      kind: "preview_linked_evidence",
      preventDefault: true,
      stopPropagation: true,
    };
  }

  return noApplicationShortcut;
}
