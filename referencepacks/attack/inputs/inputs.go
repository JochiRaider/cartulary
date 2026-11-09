// Package attackinputs exposes the current producer machine inputs and source
// validation fixtures. The normal Go import closure binds these exact JSON bytes into both Go
// and harness cache identities. Runtime application assembly must not import it.
package attackinputs

import "embed"

// Files contains no human documentation or deployment credentials.
//
//go:embed audit-report.schema.json dataset-profile.v2.json field-policy.native.v1.json source-lock.v20.0.json source-lock.v20.1.json acquisition.v1.schema.json fixtures/inputs.v1.json fixtures/cases.v1.json
var Files embed.FS
