import type {
  WorkbookSavedViewLayoutJson,
  WorkbookSavedViewQueryJson,
} from "../models/workbookQuery";
import type { SavedViewResource } from "../models/workbookSavedViews";
import type { WorkbookPortResult } from "./WorkbookPortResult";

export type WorkbookSavedViewDefinition = {
  readonly displayName: string;
  readonly layoutJson: WorkbookSavedViewLayoutJson;
  readonly queryJson: WorkbookSavedViewQueryJson;
  readonly scope: "private" | "shared";
  readonly viewSchemaId: string;
};

export interface WorkbookSavedViewPort {
  listPage(input: {
    readonly cursorToken: string | null;
    readonly limit: number;
    readonly signal: AbortSignal;
  }): Promise<
    WorkbookPortResult<{
      readonly nextCursor: string | null;
      readonly savedViews: readonly SavedViewResource[];
    }>
  >;
  create(input: {
    readonly definition: WorkbookSavedViewDefinition;
    readonly signal: AbortSignal;
  }): Promise<WorkbookPortResult<SavedViewResource>>;
  patch(input: {
    readonly baseVersion: number;
    readonly definition: Omit<WorkbookSavedViewDefinition, "viewSchemaId">;
    readonly savedViewId: string;
    readonly scope: SavedViewResource["scope"];
    readonly signal: AbortSignal;
    readonly viewSchemaId: string;
  }): Promise<WorkbookPortResult<SavedViewResource>>;
  delete(input: {
    readonly savedViewId: string;
    readonly scope: SavedViewResource["scope"];
    readonly signal: AbortSignal;
  }): Promise<WorkbookPortResult<void>>;
}
