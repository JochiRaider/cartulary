import type { APIError } from "../services/browserApi";
import { publicErrorView } from "../services/browserApi";
import {
  activeBadgeStyle,
  closedBadgeStyle,
  errorDetailStyle,
  errorMessageStyle,
  publicErrorStyle,
  unsetValueStyle,
} from "./landingAdminStyles";

export function MutedUnset({ value }: { value: string | null | undefined }) {
  if (value === null || typeof value === "undefined" || value === "") {
    return <span style={unsetValueStyle}>Not set</span>;
  }
  return <span>{value}</span>;
}

export function StatusBadge({ value }: { value: string }) {
  const isClosed = value === "closed";
  return (
    <span style={isClosed ? closedBadgeStyle : activeBadgeStyle}>{value}</span>
  );
}

export function PublicErrorSummary({
  error,
  testIds,
}: {
  error: APIError | null;
  testIds: {
    readonly container: string;
    readonly details: string;
    readonly message: string;
  };
}) {
  const view = publicErrorView(error);
  return (
    <div
      data-testid={testIds.container}
      role={view === null ? undefined : "alert"}
      style={publicErrorStyle}
    >
      <p data-testid={testIds.message} style={errorMessageStyle}>
        {view?.statusText ?? ""}
      </p>
      <p data-testid={testIds.details} style={errorDetailStyle}>
        {view?.details
          .map((detail) => `${detail.label}: ${detail.value}`)
          .join(" · ") ?? ""}
      </p>
    </div>
  );
}

export function formatNullableDateTime(value: string | null | undefined) {
  if (value === null || typeof value === "undefined" || value === "") {
    return "Not recorded";
  }
  return value;
}
