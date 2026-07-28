import { describe, expect, it } from "vitest";
import {
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
    expect(isSupportedNetworkFlowContract({ contract_major: 2 })).toBe(true);
    expect(isSupportedNetworkFlowContract({ contract_major: 1 })).toBe(false);
  });
});
