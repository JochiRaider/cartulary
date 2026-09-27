export type WorkbookApplicationShortcutEvent = {
  readonly altKey?: boolean | undefined;
  readonly ctrlKey?: boolean | undefined;
  readonly key: string;
  readonly metaKey?: boolean | undefined;
  readonly shiftKey?: boolean | undefined;
};

export type WorkbookApplicationShortcutContext = {
  readonly capabilities: {
    readonly find?: boolean;
    readonly closeInspector: boolean;
    readonly history: boolean;
    readonly linkedEvidence: boolean;
    readonly quickLink: boolean;
  };
  readonly focusOwner:
    | "find"
    | "editor"
    | "grid_navigation"
    | "inspector"
    | "menu"
    | "overlay";
  readonly rowKind: "committed" | "draft" | "group" | "none";
  readonly selectionIdentity: string | null;
};

type ConsumedApplicationShortcut = {
  readonly preventDefault: true;
  readonly stopPropagation: true;
};

export type WorkbookApplicationShortcutDecision =
  | ({ readonly kind: "open_find" } & ConsumedApplicationShortcut)
  | ({ readonly kind: "quick_link" } & ConsumedApplicationShortcut)
  | ({ readonly kind: "preview_linked_evidence" } & ConsumedApplicationShortcut)
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
  if (
    hasCommandModifier &&
    !event.altKey &&
    !event.shiftKey &&
    event.key.toLowerCase() === "f" &&
    context.capabilities.find &&
    (context.focusOwner === "grid_navigation" || context.focusOwner === "find")
  ) {
    return { kind: "open_find", preventDefault: true, stopPropagation: true };
  }
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
      kind: "preview_linked_evidence",
      preventDefault: true,
      stopPropagation: true,
    };
  }

  return noApplicationShortcut;
}
