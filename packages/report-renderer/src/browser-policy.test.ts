import { describe, expect, it } from "vitest";

import { browserDestination } from "./browser-policy.ts";

const endpoint = { port: 31973, portRef: "renderer.http" };

describe("renderer browser destination policy", () => {
  it("admits only the allocated loopback transport without credentials", () => {
    for (const url of [
      "http://127.0.0.1:31973/print",
      "http://[::1]:31973/print",
      "http://localhost:31973/print",
      "ws://127.0.0.1:31973/channel",
    ]) {
      expect(browserDestination(url, endpoint)).toEqual({
        observation_kind: "loopback_allowed",
        destination_class: "loopback",
        allowed: true,
        loopback_port_ref: endpoint.portRef,
      });
    }
    for (const url of [
      "http://127.0.0.1:31974/",
      "http://localhost/",
      "https://localhost:31973/",
      "http://user:password@localhost:31973/",
    ]) {
      expect(browserDestination(url, endpoint)).toMatchObject({
        observation_kind: "loopback_blocked",
        allowed: false,
        loopback_port_ref: null,
      });
    }
  });

  it("classifies external and forbidden schemes without retaining destinations", () => {
    for (const url of [
      "https://private.example.invalid/secret?q=secret",
      "http://127.0.0.1.example.invalid:31973/",
      "http://localhost@private.example.invalid:31973/",
      "ws://192.0.2.1:31973/",
    ]) {
      const result = browserDestination(url, endpoint);
      expect(result.observation_kind).toBe("non_loopback_blocked");
      expect(result.allowed).toBe(false);
      expect(JSON.stringify(result)).not.toContain("secret");
      expect(JSON.stringify(result)).not.toContain("private");
    }
    for (const url of [
      "file:///secret",
      "data:text/plain,secret",
      "javascript:secret",
      "invalid",
    ]) {
      expect(browserDestination(url, endpoint)).toMatchObject({
        observation_kind: "url_scheme_blocked",
        allowed: false,
        loopback_port_ref: null,
      });
    }
  });
});
