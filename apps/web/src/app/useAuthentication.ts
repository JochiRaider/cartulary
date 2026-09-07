import {
  type FormEvent,
  useEffect,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import type { APIError } from "../services/browserApi";
import {
  type AuthenticationController,
  authBannerForError,
} from "./authenticationModel";

type AuthSurfaceBootstrapState =
  | "loading"
  | "anonymous"
  | "revoked"
  | "public_error_envelope";
type AuthChallengeState = "mfa_required" | "mfa_setup_required";
type AuthGatewayProps = {
  controller: AuthenticationController;
  bootstrapState: AuthSurfaceBootstrapState;
  message: string;
  publicError?: APIError | null;
  readingProfile?: "default" | "hyperlegible" | undefined;
};
export function normalizeTotpCode(value: string) {
  return value.replace(/\D/gu, "").slice(0, 6);
}
function firstFieldError(
  errors: ReturnType<AuthenticationController["getSnapshot"]>["fieldErrors"],
) {
  return (
    errors.username ??
    errors.password ??
    errors.totpCode ??
    errors.bootstrapCompleteCode ??
    ""
  );
}
export function useAuthentication({
  controller,
  bootstrapState,
  message,
  publicError = null,
  readingProfile = "default",
}: AuthGatewayProps) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const usernameRef = useRef<HTMLInputElement>(null);
  const passwordRef = useRef<HTMLInputElement>(null);
  const totpCodeRef = useRef<HTMLInputElement>(null);
  const bootstrapBeginRef = useRef<HTMLButtonElement>(null);
  const bootstrapCompleteCodeRef = useRef<HTMLInputElement>(null);
  useLayoutEffect(() => {
    controller.open();
    return controller.close;
  }, [controller]);
  useEffect(() => {
    if (bootstrapState !== "loading") void controller.discover();
  }, [controller, bootstrapState]);
  const enrollmentReady =
    state.flow.kind === "setup" && state.flow.enrollment.kind === "ready";
  useEffect(() => {
    if (state.fieldErrors.username) usernameRef.current?.focus();
    else if (state.fieldErrors.password) passwordRef.current?.focus();
    else if (state.fieldErrors.totpCode) totpCodeRef.current?.focus();
    else if (state.fieldErrors.bootstrapCompleteCode)
      bootstrapCompleteCodeRef.current?.focus();
    else if (state.flow.kind === "mfa") totpCodeRef.current?.focus();
    else if (state.flow.kind === "setup") {
      if (enrollmentReady) bootstrapCompleteCodeRef.current?.focus();
      else bootstrapBeginRef.current?.focus();
    }
  }, [state.fieldErrors, state.flow.kind, enrollmentReady]);
  const handleLoginSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    controller.login();
  };
  const displayedBootstrapState:
    | AuthChallengeState
    | AuthSurfaceBootstrapState =
    state.flow.kind === "mfa"
      ? "mfa_required"
      : state.flow.kind === "setup"
        ? "mfa_setup_required"
        : bootstrapState;
  const externalBanner =
    publicError === null || state.banner !== null
      ? null
      : authBannerForError(publicError);
  const displayedBanner = state.banner ?? externalBanner;
  const currentFieldError = firstFieldError(state.fieldErrors);
  const authAlertText = displayedBanner?.message ?? currentFieldError;
  const authLiveRole =
    authAlertText === ""
      ? undefined
      : displayedBanner?.tone === "error" || currentFieldError !== ""
        ? "alert"
        : "status";
  const activeErrorCode = state.error?.code ?? publicError?.code ?? "";
  const statusText =
    bootstrapState === "loading"
      ? "Checking current session..."
      : state.operation.kind === "pending" && state.operation.action === "login"
        ? "Signing in..."
        : state.operation.kind === "pending" &&
            state.operation.action === "begin"
          ? "Beginning authenticator enrollment..."
          : state.operation.kind === "pending" &&
              state.operation.action === "complete"
            ? "Completing authenticator enrollment..."
            : state.flow.kind === "mfa"
              ? "Authenticator code required."
              : state.flow.kind === "setup"
                ? "Authenticator setup required."
                : "";
  const rootClassName =
    readingProfile === "hyperlegible"
      ? "cartulary-shell cartulary-auth-shell cartulary-auth-hyperlegible"
      : "cartulary-shell cartulary-auth-shell";
  const supportText =
    state.flow.kind === "mfa"
      ? "Enter the authenticator code for this account."
      : state.flow.kind === "setup"
        ? "Complete authenticator setup before signing in."
        : message;
  const title =
    state.flow.kind === "mfa"
      ? "Verify your identity"
      : state.flow.kind === "setup"
        ? "Set up authenticator"
        : "Sign in to Cartulary";
  const submitLabel =
    state.operation.kind === "pending" && state.operation.action === "login"
      ? "Signing in..."
      : state.flow.kind === "mfa"
        ? "Verify and sign in"
        : "Sign in";
  const canSubmit = controller.canStart() && bootstrapState !== "loading";

  return {
    snapshot: state,
    commands: controller,
    usernameRef,
    passwordRef,
    totpCodeRef,
    bootstrapBeginRef,
    bootstrapCompleteCodeRef,
    handleLoginSubmit,
    displayedBootstrapState,
    displayedBanner,
    authAlertText,
    authLiveRole,
    activeErrorCode,
    statusText,
    rootClassName,
    supportText,
    title,
    submitLabel,
    canSubmit,
  };
}
