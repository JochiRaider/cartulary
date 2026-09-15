import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";

import type { WorkbookQueryRow } from "./WorkbookQueryRow";

/** Applied server metadata, including sorts that are not authored overrides. */
export type WorkbookCanonicalQuery = {
  readonly filters: WorkbookQueryState["filters"];
  readonly sort: WorkbookQueryState["sort"];
  readonly groupBy?: string;
};

export type WorkbookQueryRequestIdentity = {
  readonly authorityGeneration: number;
  readonly requestGeneration: number;
  readonly surfaceIdentity: string;
};

export type WorkbookProducingQueryRequest = {
  readonly queryState: WorkbookQueryState;
  readonly limit: number;
  readonly cursorToken?: string;
  readonly identity?: WorkbookQueryRequestIdentity;
};

export type WorkbookQueryPaging = {
  readonly limit: number;
  readonly hasMore: boolean;
  readonly nextCursor: string | null;
};

export type WorkbookViewQueryAccepted = {
  readonly incidentId: string;
  readonly rows: readonly WorkbookQueryRow[];
  readonly viewSchemaId: string;
  readonly canonicalQuery: WorkbookCanonicalQuery;
  readonly paging: WorkbookQueryPaging;
  readonly producingRequest: WorkbookProducingQueryRequest;
};

export type WorkbookViewQueryResult =
  WorkbookPortResult<WorkbookViewQueryAccepted>;

export interface WorkbookViewQueryPort {
  query(input: {
    readonly contract: ViewContract;
    readonly queryState: WorkbookQueryState;
    readonly signal: AbortSignal;
    readonly limit?: number;
    readonly cursorToken?: string;
    readonly expectedCanonicalQuery?: WorkbookCanonicalQuery;
    readonly identity?: WorkbookQueryRequestIdentity;
  }): Promise<WorkbookViewQueryResult>;
}
