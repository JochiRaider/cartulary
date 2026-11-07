package producer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type envelopeDestination map[string][]byte

func (d envelopeDestination) Write(_ context.Context, name string, write func(io.Writer) error) error {
	var b bytes.Buffer
	if err := write(&b); err != nil {
		return err
	}
	d[name] = b.Bytes()
	return nil
}

// Use the runtime's archive, trust and content admission with the actual clock
// before exposing a real release. This neither installs trust nor activates it.
func verifyReleaseEnvelope(ctx context.Context, container, bootstrap []byte, report EnvelopeReport, ledgerRoot string, policy publicationPolicy) error {
	files := envelopeDestination{}
	inventory, err := packformat.Extract(ctx, bytes.NewReader(container), int64(len(container)), packformat.DefaultArchiveLimits(), files)
	if err != nil {
		return err
	}
	trust, err := packformat.DecodeBootstrap(bootstrap)
	if err != nil {
		return err
	}
	repository, err := packformat.DecodeHint(files["bundle.json"])
	if err != nil || repository != report.Repository {
		return errors.New("release repository mismatch")
	}
	metadata := map[string][]byte{}
	for name, b := range files {
		if strings.HasPrefix(name, "metadata/") {
			metadata[name] = b
		}
	}
	at := time.Now().UTC()
	proposal, err := packformat.VerifyTrust(metadata, inventory, repository, at, trust[repository])
	if err != nil {
		return err
	}
	m, err := packformat.DecodeManifest(files["manifest.json"], true)
	if err != nil {
		return err
	}
	if err = packformat.ValidateInventory(m, inventory, true); err != nil {
		return err
	}
	if err = packformat.TrustBindingFindings(m, report.ManifestSHA256, report.PayloadSHA256, proposal, func(packformat.Finding) error { return errors.New("release trust binding mismatch") }); err != nil {
		return err
	}
	if err = packformat.ValidateContent(ctx, m, packageSource(files), nil); err != nil {
		return err
	}
	receipt, err := canonicaljson.Marshal(map[string]any{"schema_id": policy.Prefix + "_verification." + policy.Version, "container_sha256": Digest(container), "bootstrap_sha256": Digest(bootstrap), "manifest_sha256": report.ManifestSHA256, "payload_sha256": report.PayloadSHA256, "verified_at": at.Format(time.RFC3339Nano), "trust_valid_until": proposal.ValidUntil.Format(time.RFC3339), "status": "accepted"})
	if err != nil {
		return err
	}
	return publishTree(ledgerRoot, "verification/"+Digest(receipt), map[string][]byte{"verification.json": receipt})
}
