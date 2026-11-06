package reporting

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/artifactport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPortableArtifactStreamRetention_Integration(t *testing.T) {
	ctx := context.Background()
	database := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reporting-stream-retention")
	pool, err := pgxpool.New(ctx, database.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	actor, incident, operation := uuid.New(), uuid.New(), uuid.New()
	for _, seed := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO users(id,email,display_name,password_hash) VALUES($1,'stream@example.test','Stream fixture','unused')`, []any{actor}},
		{`INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'STREAM','stream','Stream fixture','active',$2,$2)`, []any{incident, actor}},
		{`INSERT INTO jobs(job_id,scope_kind,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,handler_name,job_kind,progress_unit_id) VALUES($1,'deployment','queued',true,$2,now(),now(),0,'stream','stream','cartulary.stream.progress.v1')`, []any{operation, actor}},
	} {
		if _, err := pool.Exec(ctx, seed.sql, seed.args...); err != nil {
			t.Fatal(err)
		}
	}
	const size = int64(65<<20 + 17)
	digest := sha256.New()
	if _, err := io.CopyBuffer(digest, &portablePatternReader{remaining: size}, make([]byte, 8192)); err != nil {
		t.Fatal(err)
	}
	member := incidentportability.Member{Path: "ext/snapshots/storage-boundary.bin", ByteSize: size, SHA256: hex.EncodeToString(digest.Sum(nil))}
	input := &portablePatternMembers{member: member}
	bound, err := incidentportability.BindMembers(input, []string{member.Path})
	if err != nil {
		t.Fatal(err)
	}
	prepared := &preparedArtifactImport{incident: incident, operation: operation, members: bound}
	for attempt := 0; attempt < 2; attempt++ {
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if err := prepared.ApplyTx(ctx, tx); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	var files, chunks, maximum int64
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reporting_imported_artifact_files WHERE incident_id=$1`, incident).Scan(&files); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*),max(octet_length(content)) FROM reporting_imported_artifact_chunks WHERE incident_id=$1`, incident).Scan(&chunks, &maximum); err != nil {
		t.Fatal(err)
	}
	if files != 1 || chunks != (size+65535)/65536 || maximum > 65536 || input.maximumRead > 65536 {
		t.Fatalf("unbounded or duplicated retention: files=%d chunks=%d max=%d source_read=%d", files, chunks, maximum, input.maximumRead)
	}
	verify := func(tx pgx.Tx) error {
		retained, err := loadImportedArtifactMembers(ctx, tx, incident)
		if err != nil {
			return err
		}
		reader, err := incidentportability.OpenVerifiedMember(ctx, retained, member)
		if err != nil {
			return err
		}
		n, err := io.CopyBuffer(io.Discard, reader, make([]byte, 8192))
		err = errors.Join(err, reader.Close())
		if err == nil && n != size {
			return artifactport.ErrInvalid
		}
		return err
	}
	for _, scenario := range []string{"exact", "missing chunk", "changed digest", "foreign operation"} {
		t.Run(scenario, func(t *testing.T) {
			tx, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			switch scenario {
			case "missing chunk":
				if _, err := tx.Exec(ctx, `DELETE FROM reporting_imported_artifact_chunks WHERE incident_id=$1 AND chunk_ordinal=10`, incident); err != nil {
					t.Fatal(err)
				}
			case "changed digest":
				// Replacing a chunk in the disposable database simulates corruption;
				// ordinary application roles have neither DELETE nor UPDATE access.
				if _, err := tx.Exec(ctx, `DELETE FROM reporting_imported_artifact_chunks WHERE incident_id=$1 AND chunk_ordinal=10`, incident); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `INSERT INTO reporting_imported_artifact_chunks SELECT $1,$2,10,decode(repeat('ff',65536),'hex')`, incident, member.Path); err != nil {
					t.Fatal(err)
				}
			case "foreign operation":
				copy := *prepared
				copy.operation = uuid.New()
				if err := copy.ApplyTx(ctx, tx); !errors.Is(err, artifactport.ErrInvalid) {
					t.Fatalf("foreign replay: %v", err)
				}
				return
			}
			err = verify(tx)
			if (err == nil) != (scenario == "exact") {
				t.Fatalf("%s: %v", scenario, err)
			}
		})
	}
	t.Run("changed source rolls back publication", func(t *testing.T) {
		other := member
		other.Path = "ext/snapshots/changed-source.bin"
		bad := &portablePatternMembers{member: other, corrupt: true}
		bound, err := incidentportability.BindMembers(bad, []string{other.Path})
		if err != nil {
			t.Fatal(err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		err = (&preparedArtifactImport{incident: incident, operation: operation, members: bound}).ApplyTx(ctx, tx)
		_ = tx.Rollback(ctx)
		if err == nil {
			t.Fatal("altered bytes matched the prepared digest")
		}
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM reporting_imported_artifact_files WHERE bundle_path=$1`, other.Path).Scan(&count); err != nil || count != 0 {
			t.Fatal("partial publication", count, err)
		}
	})
}

type portablePatternMembers struct {
	member      incidentportability.Member
	maximumRead int
	corrupt     bool
}

func (m *portablePatternMembers) Members() []incidentportability.Member {
	return []incidentportability.Member{m.member}
}
func (m *portablePatternMembers) OpenMember(_ context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	if path != m.member.Path {
		return nil, incidentportability.Member{}, artifactport.ErrInvalid
	}
	return io.NopCloser(&portablePatternReader{remaining: m.member.ByteSize, maximum: &m.maximumRead, corrupt: m.corrupt}), m.member, nil
}

type portablePatternReader struct {
	remaining, offset int64
	maximum           *int
	corrupt           bool
}

func (r *portablePatternReader) Read(p []byte) (int, error) {
	if r.maximum != nil {
		*r.maximum = max(*r.maximum, len(p))
		if len(p) > 65536 {
			return 0, errors.New("bulk source requested an unbounded read")
		}
	}
	if r.remaining == 0 {
		return 0, io.EOF
	}
	n := int(min(int64(len(p)), r.remaining))
	for i := 0; i < n; i++ {
		p[i] = byte((r.offset+int64(i))*31 + 7)
	}
	if r.corrupt && r.offset == 0 && n > 0 {
		p[0] ^= 1
	}
	r.remaining -= int64(n)
	r.offset += int64(n)
	return n, nil
}
