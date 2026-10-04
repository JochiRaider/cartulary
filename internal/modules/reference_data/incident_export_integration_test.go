package reference_data

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPortableContainerExport(t *testing.T, ctx context.Context, pool *pgxpool.Pool, storage *coordinatorMemoryStorage, setID string, expected []byte) {
	t.Helper()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	retained := retention{repository: &canonicalRepository{pool: pool, storage: storage}}
	refs, err := retained.ExportReferencesTx(ctx, tx, []string{setID})
	if err != nil {
		t.Fatal(err)
	}
	port, err := newTestIncidentReferences(pool, storage)
	if err != nil {
		t.Fatal(err)
	}
	writes := map[string][]byte{}
	request := IncidentReferenceExportRequest{IncidentID: uuid.New(), References: refs, Embed: true, WriteContainer: func(_ context.Context, path string, size int64, reader io.Reader) error {
		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		if int64(len(data)) != size {
			return errors.New("unexpected stream size")
		}
		writes[path] = data
		return nil
	}}
	manifest, err := port.ExportContentTx(ctx, tx, request)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeIncidentBundleReferences(refs)
	if err != nil {
		t.Fatal(err)
	}
	content, err := packformat.DecodePortableContent(manifest, decoded)
	if err != nil || len(content.Containers) != 1 || len(writes) != 1 {
		t.Fatal("exact container inventory", err)
	}
	path, err := packformat.PortableContainerPath(content.Containers[0].ManifestSHA256)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(writes[path], expected) || content.Containers[0].ContainerSHA256 != packformat.Digest(expected) || content.Containers[0].SizeBytes != int64(len(expected)) {
		t.Fatal("exported bytes differ from independent signed fixture")
	}
	request.Embed = false
	request.WriteContainer = func(context.Context, string, int64, io.Reader) error { t.Fatal("refs-only read container"); return nil }
	if result, err := port.ExportContentTx(ctx, tx, request); err != nil || result != nil {
		t.Fatal("refs-only invented inventory", err)
	}
	request.Embed = true
	request.WriteContainer = func(_ context.Context, _ string, _ int64, r io.Reader) error {
		_, err := io.CopyN(io.Discard, r, 1)
		return err
	}
	if _, err := port.ExportContentTx(ctx, tx, request); err == nil {
		t.Fatal("incomplete sink accepted")
	}
	sentinel := errors.New("sink unavailable")
	request.WriteContainer = func(context.Context, string, int64, io.Reader) error { return sentinel }
	if _, err := port.ExportContentTx(ctx, tx, request); !errors.Is(err, sentinel) {
		t.Fatal("sink failure degraded to optional absence", err)
	}
	var reference string
	if err := tx.QueryRow(ctx, `SELECT e.container_ref FROM reference_pack_envelopes e JOIN reference_pack_candidates c ON c.current_envelope_id=e.envelope_id WHERE e.container_sha256=$1`, packformat.Digest(expected)).Scan(&reference); err != nil {
		t.Fatal(err)
	}
	original := storage.objects[reference]
	corrupt := bytes.Clone(original)
	corrupt[0] ^= 1
	storage.objects[reference] = corrupt
	request.WriteContainer = func(context.Context, string, int64, io.Reader) error {
		t.Fatal("unavailable optional bytes reached sink")
		return nil
	}
	result, err := port.ExportContentTx(ctx, tx, request)
	storage.objects[reference] = original
	if err != nil || result != nil {
		t.Fatal("unavailable optional container invented inventory", err)
	}
	request.WriteContainer = func(_ context.Context, _ string, _ int64, r io.Reader) error {
		original[0] ^= 1
		defer func() { original[0] ^= 1 }()
		_, err := io.Copy(io.Discard, r)
		return err
	}
	if _, err := port.ExportContentTx(ctx, tx, request); err == nil {
		t.Fatal("container changed during copying")
	}
	for _, classification := range []string{"allowed", "restricted", "prohibited", ""} {
		if portableEmbeddingAllowed(packformat.Manifest{License: packformat.License{Redistribution: classification}}) != (classification == "allowed") {
			t.Fatal("redistribution rule", classification)
		}
	}
}
