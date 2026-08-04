import {
  networkFlowContractDescriptor,
  networkFlowDecoders,
  networkFlowErrorRegistry,
  networkFlowMappingRegistry,
  networkFlowPresentationRegistry,
  type TableList,
} from "@cartulary/protocol-ts/network-flow";

declare const tableList: TableList;

export const networkFlowCompileSurface = {
  decoded: networkFlowDecoders.tableList.decode(tableList),
  descriptor: networkFlowContractDescriptor,
  errors: networkFlowErrorRegistry,
  mapping: networkFlowMappingRegistry,
  presentation: networkFlowPresentationRegistry,
};
