import type { GridNavigationKey } from "@cartulary/grid-adapter";

export type WorkbookKeyboardCommand =
  | {
      readonly intent: {
        readonly key: GridNavigationKey;
        readonly shiftKey: boolean;
      };
      readonly kind: "navigate";
      readonly preventDefault: true;
    }
  | {
      readonly kind: "paste-intent";
      readonly preventDefault: false;
    }
  | {
      readonly kind: "quick-link";
      readonly preventDefault: true;
    }
  | {
      readonly kind: "preview-linked-evidence";
      readonly destination: "list_or_empty" | "sole_previewable_item";
      readonly preventDefault: true;
    }
  | {
      readonly kind: "open-history";
      readonly preventDefault: true;
    }
  | {
      readonly kind: "close-inspector";
      readonly preventDefault: true;
    }
  | {
      readonly kind: "none";
      readonly preventDefault: false;
    };

export type WorkbookKeyboardContext = {
  readonly closeInspector?: boolean | undefined;
  readonly cell?:
    | { readonly linkResolveCapability: boolean }
    | null
    | undefined;
  readonly inspectorGroups?: readonly ("evidence" | "history")[] | undefined;
  readonly mode?:
    | "editor"
    | "grid_navigation"
    | "inspector"
    | "menu"
    | undefined;
  readonly rowKind?: "committed" | "draft" | "group" | "none" | undefined;
  readonly committedRowIdentity?: string | null | undefined;
  readonly previewableEvidenceCount?: number | undefined;
};

export type WorkbookKeyboardEventLike = {
  readonly altKey?: boolean | undefined;
  readonly ctrlKey?: boolean | undefined;
  readonly key: string;
  readonly metaKey?: boolean | undefined;
  readonly shiftKey?: boolean | undefined;
};

const navigationKeys = new Set<string>([
  "ArrowDown",
  "ArrowLeft",
  "ArrowRight",
  "ArrowUp",
  "Enter",
  "Tab",
]);

export function mapWorkbookKeyboardCommand(
  event: WorkbookKeyboardEventLike,
  context: WorkbookKeyboardContext = {},
): WorkbookKeyboardCommand {
  const mode = context.mode ?? "grid_navigation";
  const committedGridRow =
    mode === "grid_navigation" &&
    context.rowKind === "committed" &&
    typeof context.committedRowIdentity === "string" &&
    context.committedRowIdentity.length > 0;
  const hasCommandModifier = event.ctrlKey === true || event.metaKey === true;
  const hasUnsupportedModifier =
    event.altKey === true || (hasCommandModifier && event.key !== "v");

  if (
    hasCommandModifier &&
    event.key.toLowerCase() === "v" &&
    event.altKey !== true
  ) {
    return { kind: "paste-intent", preventDefault: false };
  }

  if (
    hasCommandModifier &&
    event.key.toLowerCase() === "k" &&
    event.altKey !== true &&
    committedGridRow &&
    context.cell?.linkResolveCapability === true
  ) {
    return { kind: "quick-link", preventDefault: true };
  }

  if (
    event.altKey === true &&
    !hasCommandModifier &&
    event.key.toLowerCase() === "h" &&
    committedGridRow &&
    context.inspectorGroups?.includes("history") === true
  ) {
    return { kind: "open-history", preventDefault: true };
  }

  if (event.key === "Escape" && context.closeInspector === true) {
    return { kind: "close-inspector", preventDefault: true };
  }

  if (
    event.key === " " &&
    event.altKey !== true &&
    !hasCommandModifier &&
    committedGridRow &&
    context.inspectorGroups?.includes("evidence") === true
  ) {
    return {
      destination:
        context.previewableEvidenceCount === 1
          ? "sole_previewable_item"
          : "list_or_empty",
      kind: "preview-linked-evidence",
      preventDefault: true,
    };
  }

  if (hasUnsupportedModifier || !isGridNavigationKey(event.key)) {
    return { kind: "none", preventDefault: false };
  }

  return {
    intent: {
      key: event.key,
      shiftKey: event.shiftKey === true,
    },
    kind: "navigate",
    preventDefault: true,
  };
}

function isGridNavigationKey(key: string): key is GridNavigationKey {
  return navigationKeys.has(key);
}
