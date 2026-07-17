function originFromEnv(name: string, fallback: string) {
  return (process.env[name] ?? fallback).replace(/\/+$/, "");
}

export const apiBase = originFromEnv(
  "CARTULARY_WEB_E2E_API_ORIGIN",
  "http://127.0.0.1:8080",
);

export const webBase = originFromEnv(
  "CARTULARY_WEB_E2E_PUBLIC_ORIGIN",
  "http://127.0.0.1:4173",
);
