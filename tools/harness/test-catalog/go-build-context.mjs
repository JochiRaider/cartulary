import { readFileSync } from "node:fs";
import path from "node:path";

// One projection of the ordinary runner profile, shared by static discovery
// and execution. Runtime profiles never add compiler/build tags.
export function goRunnerBuildContext(root) {
  const pins = JSON.parse(readFileSync(path.join(root, "tools/toolchain_pins.json"), "utf8"));
  const version = /^(1)\.(\d+)(?:\.(\d+))?$/u.exec(String(pins.go_version ?? ""));
  if (!version) throw new Error("tools/toolchain_pins.json.go_version must be a Go 1.x or Go 1.x.patch release");
  const environment = Object.freeze({
    GOOS: "linux",
    GOARCH: "amd64",
    GOAMD64: "v1",
    CGO_ENABLED: "0",
    GOFLAGS: "",
    GOENV: "off",
    GOEXPERIMENT: "",
    GOTOOLCHAIN: `go${pins.go_version}`,
    GOWORK: "off",
    GO111MODULE: "on",
  });
  const tags = new Set([environment.GOARCH, `${environment.GOARCH}.${environment.GOAMD64}`, "gc", environment.GOOS, "unix"]);
  for (let minor = 1; minor <= Number(version[2]); minor += 1) tags.add(`go1.${minor}`);
  return Object.freeze({
    goarch: environment.GOARCH,
    goos: environment.GOOS,
    label: `linux/amd64.v1 Go ${pins.go_version}, gc, cgo disabled, no custom build tags`,
    tags,
    environment,
  });
}
