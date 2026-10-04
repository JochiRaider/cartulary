package reporting

import (
	"bytes"

	reportingjson "github.com/JochiRaider/cartulary/internal/modules/reporting/canonicaljson"
)

func canonicalReportingBytes(raw []byte) ([]byte, error) { return reportingjson.Canonicalize(raw) }
func marshalReportingJSON(value any) ([]byte, error)     { return reportingjson.Marshal(value) }

// Object identities include the owner's schema domain. Exact file checksums
// and Extensions output checksums continue to use hashHex on raw file bytes.
func reportingObjectDigest(schemaID string, canonical []byte) string {
	var input bytes.Buffer
	input.WriteString(schemaID)
	input.WriteByte('\n')
	input.Write(canonical)
	return hashHex(input.Bytes())
}
