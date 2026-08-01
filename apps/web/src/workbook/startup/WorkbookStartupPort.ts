import type {
  WorkbookStartupQuery,
  WorkbookStartupSelection,
} from "../models/workbookStartup";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";

export type WorkbookStartupAvailability = {
  readonly workspaces: readonly {
    readonly extensionProfileId: string;
    readonly workspaceKey: string;
  }[];
};

export interface WorkbookStartupPort {
  load(input: {
    readonly query: WorkbookStartupQuery;
    readonly signal: AbortSignal;
  }): Promise<
    WorkbookPortResult<{
      readonly availability: WorkbookStartupAvailability;
      readonly selection: WorkbookStartupSelection;
    }>
  >;
}
