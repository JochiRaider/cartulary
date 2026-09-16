import type {
  WorkbookSavedViewLayoutJson,
  WorkbookSavedViewQueryJson,
} from "../models/workbookQuery";
import type { SavedViewResource } from "../models/workbookSavedViews";

export type SavedViewObserver = <T>(
  request: (signal: AbortSignal) => Promise<T>,
) => {
  readonly result: Promise<
    | { kind: "completed"; value: T }
    | { kind: "timeout" | "transport" | "cancelled" }
  >;
  readonly settled: Promise<void>;
  readonly cancel: () => void;
};

export type SavedViewProblem = {
  readonly kind:
    | "validation"
    | "conflict"
    | "authentication_required"
    | "authorization_denied"
    | "unavailable_target"
    | "transport"
    | "invalid_contract"
    | "terminal";
  readonly message: string;
  readonly publicCode?: string;
  readonly field?: string;
  readonly reason?: string;
  readonly conflict?: {
    readonly savedViewId: string;
    readonly baseVersion: number;
    readonly currentVersion: number;
  };
};

export type SavedViewResult<T> =
  | { readonly kind: "accepted"; readonly value: T }
  | {
      readonly kind: "rejected" | "uncertain";
      readonly failure: SavedViewProblem;
    };

export type WorkbookSavedViewDefinition = {
  readonly displayName: string;
  readonly layoutJson: WorkbookSavedViewLayoutJson;
  readonly queryJson: WorkbookSavedViewQueryJson;
  readonly scope: "private" | "shared";
  readonly viewSchemaId: string;
};

export type WorkbookSavedViewChanges = Partial<
  Omit<WorkbookSavedViewDefinition, "viewSchemaId">
>;

export interface WorkbookSavedViewPort {
  getResource(input: {
    readonly savedViewId: string;
    readonly signal: AbortSignal;
  }): Promise<SavedViewResult<SavedViewResource>>;
  listPage(input: {
    readonly viewSchemaId: string;
    readonly cursorToken: string | null;
    readonly limit: number;
    readonly signal: AbortSignal;
  }): Promise<
    SavedViewResult<{
      readonly nextCursor: string | null;
      readonly savedViews: readonly SavedViewResource[];
    }>
  >;
  create(input: {
    readonly definition: WorkbookSavedViewDefinition;
    readonly signal: AbortSignal;
  }): Promise<SavedViewResult<SavedViewResource>>;
  patch(input: {
    readonly base: SavedViewResource;
    readonly changes: WorkbookSavedViewChanges;
    readonly signal: AbortSignal;
  }): Promise<SavedViewResult<SavedViewResource>>;
  delete(input: {
    readonly savedViewId: string;
    readonly scope: SavedViewResource["scope"];
    readonly signal: AbortSignal;
  }): Promise<SavedViewResult<undefined>>;
}
