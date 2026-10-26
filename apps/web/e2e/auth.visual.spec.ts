import { authTestId } from "@cartulary/ui-contracts";
import type { Page, Route } from "@playwright/test";
import { expect } from "./fixtures";
import { withPresentation } from "./support/accessibility/presentation";
import {
  navigateVisualApplication,
  reloadVisualApplication,
} from "./support/runtime/applicationReadiness";
import { responseBarrier } from "./support/runtime/cleanup";
import {
  awaitResponseBarrier,
  installOwnedRoute,
} from "./support/transport/requestInterception";
import { assertViewportVisualRegression } from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import { attachFontManifestDigest } from "./support/visual/fonts";

async function fillAuthVisualCredentials(page: Page) {
  await page
    .getByTestId(authTestId("login-username"))
    .fill("visual.operator@example.test");
  await page.getByTestId(authTestId("login-password")).fill("VisualPass1!");
}

function releaseAuthVisualStep(release: (() => void) | null) {
  release?.();
}

async function assertAuthGatewayVisual(page: Page, name: string) {
  await assertViewportVisualRegression(page, name);
  await test.info().attach(`${name}.png`, {
    body: await page.screenshot({ fullPage: false }),
    contentType: "image/png",
  });
}

async function fulfillAuthVisualLogin(route: Route, mode: AuthVisualLoginMode) {
  if (mode === "mfa_required") {
    await fulfillAuthVisualError(route, {
      code: "mfa_required",
      details: {
        required_second_factor_kinds: ["totp"],
      },
      message: "MFA is required.",
      status: 401,
    });
    return;
  }
  if (mode === "mfa_setup_required") {
    await fulfillAuthVisualError(route, {
      code: "mfa_setup_required",
      details: {
        bootstrap_token: "visual-bootstrap-token",
        bootstrap_expires_at: new Date(Date.now() + 300_000).toISOString(),
        required_setup_kinds: ["totp"],
      },
      message: "Authenticator setup is required.",
      status: 401,
    });
    return;
  }
  if (mode === "invalid_mfa") {
    await fulfillAuthVisualError(route, {
      code: "invalid_second_factor",
      message: "Invalid second factor.",
      status: 401,
    });
    return;
  }
  if (mode === "service_unavailable") {
    await fulfillAuthVisualError(route, {
      code: "service_unavailable",
      message: "Authentication service unavailable.",
      status: 503,
    });
    return;
  }
  await fulfillAuthVisualError(route, {
    code: "invalid_credentials",
    message: "Invalid credentials.",
    status: 401,
  });
}

async function fulfillAuthVisualJSON(
  route: Route,
  body: Record<string, unknown>,
  status = 200,
) {
  await route.fulfill({
    contentType: "application/json",
    status,
    body: JSON.stringify(body),
  });
}

async function fulfillAuthVisualError(
  route: Route,
  options: {
    code: string;
    details?: Record<string, unknown>;
    message: string;
    status: number;
  },
) {
  await fulfillAuthVisualJSON(
    route,
    {
      error: {
        code: options.code,
        status: options.status,
        message: options.message,
        details: options.details ?? {},
      },
    },
    options.status,
  );
}

type AuthVisualLoginMode =
  | "invalid_credentials"
  | "invalid_mfa"
  | "mfa_required"
  | "mfa_setup_required"
  | "pending"
  | "service_unavailable";

test("Capture auth gateway initial, focused, loading, invalid credentials, MFA required, invalid MFA, MFA setup required, service unavailable, mobile, reduced-motion, and 200%-zoom states.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await attachFontManifestDigest();

  let sessionPending = true;
  let releaseSession: (() => void) | null = null;
  let loginMode: AuthVisualLoginMode = "invalid_credentials";
  let pendingLoginResult: AuthVisualLoginMode = "invalid_credentials";
  let releaseLogin: (() => void) | null = null;

  await installOwnedRoute(page, "**/api/v1/auth/session", async (route) => {
    if (sessionPending) {
      const heldResponse = responseBarrier(page);
      releaseSession = heldResponse.release;
      if (!(await awaitResponseBarrier(route, heldResponse.promise))) return;
    }
    await fulfillAuthVisualError(route, {
      code: "session_required",
      message: "Session required.",
      status: 401,
    });
  });
  await installOwnedRoute(page, "**/api/v1/auth/providers", async (route) => {
    await fulfillAuthVisualJSON(route, {
      meta: { request_id: "auth-visual-providers" },
      data: { providers: [] },
    });
  });
  await installOwnedRoute(page, "**/api/v1/auth/login", async (route) => {
    const mode = loginMode;
    if (mode === "pending") {
      const heldResponse = responseBarrier(page);
      releaseLogin = heldResponse.release;
      if (!(await awaitResponseBarrier(route, heldResponse.promise))) return;
      await fulfillAuthVisualLogin(route, pendingLoginResult);
      return;
    }
    await fulfillAuthVisualLogin(route, mode);
  });

  await navigateVisualApplication(page, "/");
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "loading",
  );
  await assertAuthGatewayVisual(page, "auth-loading");
  sessionPending = false;
  releaseAuthVisualStep(releaseSession);

  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "anonymous",
  );
  await expect(page.getByTestId(authTestId("login-totp-code"))).toHaveCount(0);
  await assertAuthGatewayVisual(page, "auth-initial");

  await page.getByTestId(authTestId("login-username")).focus();
  await assertAuthGatewayVisual(page, "auth-focused");

  await fillAuthVisualCredentials(page);
  loginMode = "pending";
  pendingLoginResult = "invalid_credentials";
  await page.getByTestId(authTestId("login-submit")).click();
  await expect(page.getByTestId(authTestId("login-submit"))).toHaveText(
    "Signing in...",
  );
  await assertAuthGatewayVisual(page, "auth-submitting");
  releaseAuthVisualStep(releaseLogin);
  await expect(page.getByTestId(authTestId("feedback"))).toHaveText(
    "Email or password is incorrect.",
  );
  await assertAuthGatewayVisual(page, "auth-invalid-credentials");

  loginMode = "mfa_required";
  await fillAuthVisualCredentials(page);
  await page.getByTestId(authTestId("login-submit")).click();
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_required",
  );
  await assertAuthGatewayVisual(page, "auth-mfa-required");

  loginMode = "invalid_mfa";
  await page.getByTestId(authTestId("login-totp-code")).fill("000000");
  await page.getByTestId(authTestId("login-submit")).click();
  await expect(page.getByTestId(authTestId("feedback"))).toHaveText(
    "The verification code is incorrect or expired.",
  );
  await assertAuthGatewayVisual(page, "auth-invalid-mfa");

  loginMode = "mfa_setup_required";
  await test.step("reload visual application", () =>
    reloadVisualApplication(page));
  await fillAuthVisualCredentials(page);
  await page.getByTestId(authTestId("login-submit")).click();
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_setup_required",
  );
  await assertAuthGatewayVisual(page, "auth-mfa-setup-required");

  loginMode = "service_unavailable";
  await test.step("reload visual application", () =>
    reloadVisualApplication(page));
  await fillAuthVisualCredentials(page);
  await page.getByTestId(authTestId("login-submit")).click();
  await expect(page.getByTestId(authTestId("feedback"))).toHaveText(
    "The earlier authentication remains unconfirmed. Check the current session again or enter fresh credentials after the request settles; reload if it does not finish.",
  );
  await expect(page.getByTestId(authTestId("login-password"))).toHaveValue("");
  await assertAuthGatewayVisual(page, "auth-service-unavailable");

  await page.setViewportSize({ width: 390, height: 844 });
  await test.step("reload visual application", () =>
    reloadVisualApplication(page));
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "anonymous",
  );
  await assertAuthGatewayVisual(page, "auth-mobile");

  await page.setViewportSize({ width: 1440, height: 900 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  await test.step("reload visual application", () =>
    reloadVisualApplication(page));
  await assertAuthGatewayVisual(page, "auth-reduced-motion");
  await page.emulateMedia({ reducedMotion: "no-preference" });

  await withPresentation(page, { zoom: "200%" }, async () => {
    await assertAuthGatewayVisual(page, "auth-200-zoom");
  });
});
