import type {
  RecordHistoryItem as GeneratedHistoryItem,
  GetRecordHistoryResponse,
} from "@cartulary/protocol-ts/http";

export type RecordHistoryItem = Readonly<
  Omit<GeneratedHistoryItem, "available_rollback_actions" | "diff_summary">
> & {
  readonly available_rollback_actions: readonly GeneratedHistoryItem["available_rollback_actions"][number][];
  readonly diff_summary: {
    readonly summary: string;
    readonly units: readonly Record<string, unknown>[];
  };
};
export type RecordHistoryData = Readonly<
  Omit<GetRecordHistoryResponse["data"], "items">
> & {
  readonly items: readonly RecordHistoryItem[];
};
export type HistoryPaging = Readonly<
  GetRecordHistoryResponse["meta"]["paging"]
>;
