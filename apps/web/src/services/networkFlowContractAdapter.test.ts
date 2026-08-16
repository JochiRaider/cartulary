import { describe, expect, it } from "vitest";
import {
  decodeNetworkFlowSavedGraphList,
  isNetworkFlowClaimed,
  isSupportedNetworkFlowContract,
  networkFlowContractDescriptor,
} from "./networkFlowContractAdapter";

describe("networkFlowContractAdapter", () => {
  it("requires a claimed profile and a supported compiled contract major", () => {
    expect(
      isNetworkFlowClaimed([
        {
          profile_id: networkFlowContractDescriptor.profile_id,
          claimed: true,
        },
      ]),
    ).toBe(true);
    expect(
      isNetworkFlowClaimed([
        {
          profile_id: networkFlowContractDescriptor.profile_id,
          claimed: false,
        },
      ]),
    ).toBe(false);
    expect(isSupportedNetworkFlowContract({ contract_major: 3 })).toBe(true);
    expect(isSupportedNetworkFlowContract({ contract_major: 2 })).toBe(false);
    expect(isSupportedNetworkFlowContract({ contract_major: 1 })).toBe(false);
  });

  it("decodes saved graphs only through the generated major-3 contract", () => {
    expect(
      decodeNetworkFlowSavedGraphList({
        schema_id: "cartulary.network_flow.graph_view_list.v1",
        graph_views: [],
      }).graph_views,
    ).toEqual([]);
    expect(() =>
      decodeNetworkFlowSavedGraphList({
        schema_id: "cartulary.network_flow.graph_view_list.v1",
        graph_views: [],
        legacy_graph_projection: true,
      }),
    ).toThrow(/validation/u);
  });
});
