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

export type WorkbookKeyboardAvailability = {
  readonly closeInspector?: boolean | undefined;
  readonly history?: boolean | undefined;
  readonly previewLinkedEvidence?: boolean | undefined;
  readonly quickLink?: boolean | undefined;
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
  availability: WorkbookKeyboardAvailability = {},
): WorkbookKeyboardCommand {
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
    availability.quickLink === true
  ) {
    return { kind: "quick-link", preventDefault: true };
  }

  if (
    event.altKey === true &&
    !hasCommandModifier &&
    event.key.toLowerCase() === "h" &&
    availability.history === true
  ) {
    return { kind: "open-history", preventDefault: true };
  }

  if (event.key === "Escape" && availability.closeInspector === true) {
    return { kind: "close-inspector", preventDefault: true };
  }

  if (
    event.key === " " &&
    event.altKey !== true &&
    !hasCommandModifier &&
    availability.previewLinkedEvidence === true
  ) {
    return { kind: "preview-linked-evidence", preventDefault: true };
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
