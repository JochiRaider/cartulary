package reporting

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sort"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/artifactport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const portableArtifactChunkBytes = 64 << 10

func (p *preparedArtifactImport) ApplyTx(ctx context.Context, tx pgx.Tx) error {
	if tx == nil || p == nil || p.members == nil {
		return artifactport.ErrInvalid
	}
	for _, member := range p.members.Members() {
		tag, err := tx.Exec(ctx, `INSERT INTO reporting_imported_artifact_files(incident_id,bundle_path,operation_id,size_bytes,sha256) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, p.incident, member.Path, p.operation, member.ByteSize, member.SHA256)
		if err != nil {
			return err
		}
		inserted := tag.RowsAffected() == 1
		if !inserted {
			var same bool
			if err := tx.QueryRow(ctx, `SELECT operation_id=$3 AND size_bytes=$4 AND sha256=$5 FROM reporting_imported_artifact_files WHERE incident_id=$1 AND bundle_path=$2`, p.incident, member.Path, p.operation, member.ByteSize, member.SHA256).Scan(&same); err != nil {
				return err
			}
			if !same {
				return artifactport.ErrInvalid
			}
		}
		source, err := incidentportability.OpenVerifiedMember(ctx, p.members, member)
		if err != nil {
			return err
		}
		ordinal, writeErr := writePortableArtifactChunks(ctx, tx, p.incident, member.Path, source, inserted)
		err = errors.Join(writeErr, source.Close())
		if err != nil {
			return err
		}
		var count int64
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM reporting_imported_artifact_chunks WHERE incident_id=$1 AND bundle_path=$2`, p.incident, member.Path).Scan(&count); err != nil {
			return err
		}
		if count != ordinal {
			return artifactport.ErrInvalid
		}
	}
	return nil
}

func writePortableArtifactChunks(ctx context.Context, tx pgx.Tx, incident uuid.UUID, path string, source io.Reader, insert bool) (int64, error) {
	buffer := make([]byte, portableArtifactChunkBytes)
	ordinal := int64(0)
	for {
		n, readErr := io.ReadFull(source, buffer)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return ordinal, readErr
		}
		if n > 0 {
			if insert {
				if _, err := tx.Exec(ctx, `INSERT INTO reporting_imported_artifact_chunks(incident_id,bundle_path,chunk_ordinal,content) VALUES($1,$2,$3,$4)`, incident, path, ordinal, buffer[:n]); err != nil {
					return ordinal, err
				}
			} else {
				var retained []byte
				if err := tx.QueryRow(ctx, `SELECT content FROM reporting_imported_artifact_chunks WHERE incident_id=$1 AND bundle_path=$2 AND chunk_ordinal=$3`, incident, path, ordinal).Scan(&retained); err != nil {
					return ordinal, err
				}
				if !bytes.Equal(retained, buffer[:n]) {
					return ordinal, artifactport.ErrInvalid
				}
			}
			ordinal++
		}
		if readErr != nil {
			return ordinal, nil
		}
	}
}

type retainedArtifactMembers struct {
	query    incidentportability.Queryer
	incident uuid.UUID
	members  []incidentportability.Member
}

func loadImportedArtifactMembers(ctx context.Context, query incidentportability.Queryer, incident uuid.UUID) (*retainedArtifactMembers, error) {
	rows, err := query.Query(ctx, `SELECT bundle_path,sha256,size_bytes FROM reporting_imported_artifact_files WHERE incident_id=$1 ORDER BY bundle_path COLLATE "C"`, incident)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := &retainedArtifactMembers{query: query, incident: incident}
	for rows.Next() {
		var member incidentportability.Member
		if err := rows.Scan(&member.Path, &member.SHA256, &member.ByteSize); err != nil {
			return nil, err
		}
		if !portableMemberPath(member.Path) || !portableDigestPattern.MatchString(member.SHA256) || member.ByteSize < 0 || len(result.members) >= 10000 {
			return nil, artifactport.ErrInvalid
		}
		result.members = append(result.members, member)
	}
	return result, rows.Err()
}
func (r *retainedArtifactMembers) Members() []incidentportability.Member {
	return append([]incidentportability.Member(nil), r.members...)
}
func (r *retainedArtifactMembers) OpenMember(ctx context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	i := sort.Search(len(r.members), func(i int) bool { return r.members[i].Path >= path })
	if i == len(r.members) || r.members[i].Path != path {
		return nil, incidentportability.Member{}, artifactport.ErrInvalid
	}
	m := r.members[i]
	return &retainedArtifactReader{ctx: ctx, owner: r, member: m}, m, nil
}

// One SQL row is read per fixed-size chunk. Consumers wrap this reader in the
// shared digest verifier, including export and Recovery admission.
type retainedArtifactReader struct {
	ctx             context.Context
	owner           *retainedArtifactMembers
	member          incidentportability.Member
	ordinal, loaded int64
	buffer          []byte
	closed          bool
}

func (r *retainedArtifactReader) Read(p []byte) (int, error) {
	if r.closed {
		return 0, io.ErrClosedPipe
	}
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}
	if len(r.buffer) == 0 {
		if r.loaded == r.member.ByteSize {
			var count int64
			if err := r.owner.query.QueryRow(r.ctx, `SELECT count(*) FROM reporting_imported_artifact_chunks WHERE incident_id=$1 AND bundle_path=$2`, r.owner.incident, r.member.Path).Scan(&count); err != nil {
				return 0, err
			}
			if count != r.ordinal {
				return 0, artifactport.ErrInvalid
			}
			return 0, io.EOF
		}
		if err := r.owner.query.QueryRow(r.ctx, `SELECT content FROM reporting_imported_artifact_chunks WHERE incident_id=$1 AND bundle_path=$2 AND chunk_ordinal=$3`, r.owner.incident, r.member.Path, r.ordinal).Scan(&r.buffer); err != nil {
			return 0, err
		}
		if int64(len(r.buffer)) != min(int64(portableArtifactChunkBytes), r.member.ByteSize-r.loaded) {
			return 0, artifactport.ErrInvalid
		}
		r.loaded += int64(len(r.buffer))
		r.ordinal++
	}
	n := copy(p, r.buffer)
	r.buffer = r.buffer[n:]
	return n, nil
}
func (r *retainedArtifactReader) Close() error { r.closed = true; r.buffer = nil; return nil }
