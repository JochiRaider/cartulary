// @vitest-environment node

import { describe, expect, it, vi } from "vitest";

import { reconcileSuiteAdminTotpState } from "./helpers";

describe("reconcileSuiteAdminTotpState", () => {
  it("reuses the stored suite-admin secret when it still authenticates", async () => {
    const client = {
      loginLocal: vi.fn(async () => ({ kind: "success" as const })),
      provisionTotpFromBootstrap: vi.fn(async () => "unused"),
    };

    const secret = await reconcileSuiteAdminTotpState(
      client,
      "JBSWY3DPEHPK3PXP",
      {
        externalServerMode: true,
        sharedStateDir: "/tmp/cartulary-shared",
        stateFilePath: "/tmp/cartulary-shared/admin-totp.txt",
      },
    );

    expect(secret).toBe("JBSWY3DPEHPK3PXP");
    expect(client.loginLocal).toHaveBeenCalledOnce();
    expect(client.provisionTotpFromBootstrap).not.toHaveBeenCalled();
  });

  it("enrolls a new suite-admin secret when the backend requires TOTP setup", async () => {
    const client = {
      loginLocal: vi.fn(async () => ({
        kind: "error" as const,
        status: 401,
        code: "mfa_setup_required",
        details: { bootstrap_token: "bootstrap-token-123" },
      })),
      provisionTotpFromBootstrap: vi.fn(async () => "JBSWY3DPEHPK3PXP"),
    };

    const secret = await reconcileSuiteAdminTotpState(client, null, {
      externalServerMode: true,
      sharedStateDir: "/tmp/cartulary-shared",
      stateFilePath: "/tmp/cartulary-shared/admin-totp.txt",
    });

    expect(secret).toBe("JBSWY3DPEHPK3PXP");
    expect(client.provisionTotpFromBootstrap).toHaveBeenCalledWith(
      "bootstrap-token-123",
    );
  });

  it("fails explicitly when shared harness state is missing for an active MFA setup", async () => {
    const client = {
      loginLocal: vi.fn(async () => ({
        kind: "error" as const,
        status: 401,
        code: "mfa_required",
        details: {},
      })),
      provisionTotpFromBootstrap: vi.fn(async () => "unused"),
    };

    await expect(
      reconcileSuiteAdminTotpState(client, null, {
        externalServerMode: true,
        sharedStateDir: "/tmp/cartulary-shared",
        stateFilePath: "/tmp/cartulary-shared/admin-totp.txt",
      }),
    ).rejects.toThrow(
      "suite admin MFA is already active but no stored harness TOTP secret is available",
    );
    await expect(
      reconcileSuiteAdminTotpState(client, null, {
        externalServerMode: true,
        sharedStateDir: "/tmp/cartulary-shared",
        stateFilePath: "/tmp/cartulary-shared/admin-totp.txt",
      }),
    ).rejects.toThrow("CARTULARY_PLAYWRIGHT_STATE_DIR=/tmp/cartulary-shared");
  });
});
