package packformat

import (
	"context"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"time"
)

func VerifyTrust(files map[string][]byte, inventory Inventory, repository string, at time.Time, snapshot TrustSnapshot) (TrustProposal, error) {
	return verifyTrust(context.Background(), files, inventory, repository, at, snapshot, nil, false)
}
func verifyTrust(ctx context.Context, files map[string][]byte, inventory Inventory, repository string, at time.Time, snapshot TrustSnapshot, scratch DiagnosticScratch, diagnostics bool) (TrustProposal, error) {
	program, result := TrustChecks(files, inventory, repository, at, snapshot)
	if err := runCheckRange(ctx, "tuf_schema", "target_hash", scratch, program, diagnostics); err != nil {
		return TrustProposal{}, err
	}
	return result(), nil
}
func TrustChecks(files map[string][]byte, inventory Inventory, repository string, at time.Time, snapshot TrustSnapshot) (map[string]CheckFunc, func() TrustProposal) {
	p, r := envelope.TrustChecks(files, inventory, repository, at, snapshot)
	return formatChecks(p), r
}
