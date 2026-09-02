import type { GridNavigationIntent } from "@cartulary/grid-adapter";
import {
  mapWorkbookKeyboardCommand,
  type WorkbookKeyboardEventLike,
} from "../../utils/workbookKeyboard";
import type {
  RowValues,
  TimelineScalarEditorSurface,
  WorkbookRow,
} from "./workbookTimelineModel";

export type TimelineEditorKeyboardIntent =
  | { readonly kind: "none"; readonly preventDefault: boolean }
  | { readonly kind: "restore_prior_grid_focus"; readonly preventDefault: true }
  | { readonly kind: "close_inspector"; readonly preventDefault: true }
  | {
      readonly kind: "navigate";
      readonly navigation: GridNavigationIntent;
      readonly preventDefault: true;
    }
  | {
      readonly kind: "save";
      readonly navigateAfterSave: GridNavigationIntent | null;
      readonly preserveInputFocus: boolean;
      readonly preventDefault: true;
      readonly recordBlankRowTiming: boolean;
    };

export function mapTimelineScalarEditorIntent({
  event,
  focusField,
  hasCommittedAnchor,
  inspectorCanClose,
  priorTimelineGridAnchor,
  surface,
}: {
  readonly event: WorkbookKeyboardEventLike;
  readonly focusField: keyof RowValues;
  readonly hasCommittedAnchor: boolean;
  readonly inspectorCanClose: boolean;
  readonly priorTimelineGridAnchor: boolean;
  readonly surface: TimelineScalarEditorSurface;
}): TimelineEditorKeyboardIntent {
  if (
    surface === "inspector" &&
    event.key === "Escape" &&
    priorTimelineGridAnchor
  ) {
    return { kind: "restore_prior_grid_focus", preventDefault: true };
  }
  const command = mapWorkbookKeyboardCommand(event, {
    closeInspector: hasCommittedAnchor && inspectorCanClose,
    mode: "editor",
    rowKind: hasCommittedAnchor ? "committed" : "draft",
  });
  if (command.kind === "close-inspector") {
    return { kind: "close_inspector", preventDefault: true };
  }
  if (command.kind !== "navigate") {
    return { kind: "none", preventDefault: command.preventDefault };
  }
  if (
    surface === "grid" &&
    command.intent.shiftKey &&
    command.intent.key.startsWith("Arrow")
  ) {
    return { kind: "none", preventDefault: true };
  }
  if (!hasCommittedAnchor) {
    return command.intent.key === "Enter" || command.intent.key === "Tab"
      ? {
          kind: "save",
          navigateAfterSave: null,
          preserveInputFocus: true,
          preventDefault: true,
          recordBlankRowTiming:
            command.intent.key === "Enter" &&
            focusField === "activitySynopsisText" &&
            surface === "grid",
        }
      : { kind: "none", preventDefault: true };
  }
  if (surface === "grid" && command.intent.key === "Tab") {
    return {
      kind: "save",
      navigateAfterSave: null,
      preserveInputFocus: false,
      preventDefault: true,
      recordBlankRowTiming: false,
    };
  }
  return command.intent.key === "Enter" || command.intent.key === "Tab"
    ? {
        kind: "save",
        navigateAfterSave: command.intent,
        preserveInputFocus: false,
        preventDefault: true,
        recordBlankRowTiming: false,
      }
    : {
        kind: "navigate",
        navigation: command.intent,
        preventDefault: true,
      };
}

export function mapTimelineCollectionEditorIntent({
  event,
  hasCommittedAnchor,
  inspectorCanClose,
}: {
  readonly event: WorkbookKeyboardEventLike;
  readonly hasCommittedAnchor: boolean;
  readonly inspectorCanClose: boolean;
}): TimelineEditorKeyboardIntent {
  const command = mapWorkbookKeyboardCommand(event, {
    closeInspector: hasCommittedAnchor && inspectorCanClose,
    mode: "editor",
    rowKind: hasCommittedAnchor ? "committed" : "draft",
  });
  if (command.kind === "close-inspector") {
    return { kind: "close_inspector", preventDefault: true };
  }
  if (command.kind !== "navigate") {
    return { kind: "none", preventDefault: command.preventDefault };
  }
  if (command.intent.shiftKey && command.intent.key.startsWith("Arrow")) {
    return { kind: "none", preventDefault: true };
  }
  if (!hasCommittedAnchor) {
    return command.intent.key === "Enter" || command.intent.key === "Tab"
      ? {
          kind: "save",
          navigateAfterSave: null,
          preserveInputFocus: false,
          preventDefault: true,
          recordBlankRowTiming: false,
        }
      : { kind: "none", preventDefault: true };
  }
  if (command.intent.key === "Tab") {
    return {
      kind: "save",
      navigateAfterSave: null,
      preserveInputFocus: false,
      preventDefault: true,
      recordBlankRowTiming: false,
    };
  }
  return command.intent.key === "Enter"
    ? {
        kind: "save",
        navigateAfterSave: command.intent,
        preserveInputFocus: false,
        preventDefault: true,
        recordBlankRowTiming: false,
      }
    : {
        kind: "navigate",
        navigation: command.intent,
        preventDefault: true,
      };
}

type TimelineKeyboardRow = {
  readonly collectionValues: {
    readonly hostRefs: readonly {
      readonly itemKind: string;
      readonly itemRef: string;
    }[];
    readonly identityRefs: readonly {
      readonly itemKind: string;
      readonly itemRef: string;
    }[];
  };
  readonly recordId: WorkbookRow["recordId"];
  readonly rowVersion: WorkbookRow["rowVersion"];
};

export type TimelineWorkAreaInspectorIntent<Row extends TimelineKeyboardRow> =
  | { readonly kind: "none" }
  | {
      readonly kind: "open_panel";
      readonly panelId: "evidence" | "history";
      readonly row: Row;
    }
  | {
      readonly itemRef: string | null;
      readonly kind: "quick_link";
      readonly row: Row;
    };

export function mapTimelineWorkAreaInspectorIntent<
  Row extends TimelineKeyboardRow,
>({
  event,
  fieldKey,
  row,
}: {
  readonly event: WorkbookKeyboardEventLike;
  readonly fieldKey: string | undefined;
  readonly row: Row | null;
}): TimelineWorkAreaInspectorIntent<Row> {
  const recordId = row?.recordId ?? null;
  const command = mapWorkbookKeyboardCommand(event, {
    cell:
      fieldKey === undefined
        ? null
        : {
            linkResolveCapability:
              fieldKey === "timeline.host_refs" ||
              fieldKey === "timeline.identity_refs",
          },
    inspectorGroups: ["evidence", "history"],
    mode: "grid_navigation",
    committedRowIdentity: recordId,
    previewableEvidenceCount: 0,
    rowKind: recordId === null ? "none" : "committed",
  });
  if (row === null || recordId === null) return { kind: "none" };
  if (command.kind === "open-history") {
    return { kind: "open_panel", panelId: "history", row };
  }
  if (command.kind === "preview-linked-evidence") {
    return { kind: "open_panel", panelId: "evidence", row };
  }
  if (command.kind !== "quick-link") return { kind: "none" };
  const mention = [
    ...row.collectionValues.hostRefs,
    ...row.collectionValues.identityRefs,
  ].find((item) => item.itemKind !== "resolved_ref");
  return { itemRef: mention?.itemRef ?? null, kind: "quick_link", row };
}
