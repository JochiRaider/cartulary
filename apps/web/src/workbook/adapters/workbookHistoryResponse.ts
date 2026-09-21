import type {
  RecordHistoryItem as GeneratedHistoryItem,
  GetRecordHistoryResponse,
} from "@cartulary/protocol-ts/http";

type ReadonlyHistory<T> = T extends object
  ? { readonly [Key in keyof T]: ReadonlyHistory<T[Key]> }
  : T;
export type RecordHistoryItem = ReadonlyHistory<GeneratedHistoryItem>;
export type RecordHistoryData = Readonly<
  Omit<GetRecordHistoryResponse["data"], "items">
> & {
  readonly items: readonly RecordHistoryItem[];
};
export type HistoryPaging = Readonly<
  GetRecordHistoryResponse["meta"]["paging"]
>;
