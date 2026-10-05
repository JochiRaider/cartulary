package reference_data

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/jackc/pgx/v5"
)

// Frozen verification reads are private persistence operations. They never select
// current replacements for admitted members or trust roots.
func (c *verificationService) frozenMember(ctx context.Context, a executionAttempt, ordinal int64) (frozenMember, error) {
	m := frozenMember{Ordinal: ordinal}
	var envelope []byte
	err := c.pool.QueryRow(ctx, `SELECT m.pack_key,m.pack_version,e.canonical_envelope FROM reference_pack_operation_members m LEFT JOIN reference_pack_envelopes e ON e.envelope_id=m.envelope_id WHERE m.operation_id=$1 AND m.ordinal=$2`, a.OperationID, ordinal).Scan(&m.Key, &m.Version, &envelope)
	if errors.Is(err, pgx.ErrNoRows) && a.Frozen.Kind == "import" && ordinal == 1 {
		return m, nil
	}
	if err != nil {
		return m, err
	}
	if envelope != nil {
		decoded, err := decodeSuccessfulEnvelope(envelope)
		if err != nil {
			return m, err
		}
		m.Envelope = &decoded
	}
	return m, nil
}

// Resolve only historical roots actually carried by this container, plus the
// exact root captured at admission. Trust history may grow without making
// verification allocation proportional to the repository's lifetime.
func (c *verificationService) frozenTrust(ctx context.Context, a executionAttempt, m frozenMember, id string, versions []int64) (packformat.TrustSnapshot, bool, error) {
	s := packformat.TrustSnapshot{RootHistory: map[int64][]byte{}, Highest: map[string]packformat.RetainedMetadata{}}
	var current int64
	err := c.pool.QueryRow(ctx, `SELECT root_version FROM reference_pack_operation_repositories WHERE operation_id=$1 AND repository_id=$2`, a.OperationID, id).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, false, nil
	}
	if err != nil {
		return s, false, err
	}
	versions = append(versions, current)
	var retainedBytes int64
	if err := c.pool.QueryRow(ctx, `SELECT coalesce(sum(octet_length(canonical_bytes)),0)::bigint FROM reference_pack_roots WHERE repository_id=$1 AND root_version <= $2 AND root_version=ANY($3)`, id, current, versions).Scan(&retainedBytes); err != nil {
		return s, false, err
	}
	// Exact historical replay cannot reference more bytes than the admitted
	// metadata inventory, plus the current root. Tiny hostile root members must
	// not cause materialization of arbitrarily large retained trust history.
	if retainedBytes > 8388608+2097152 {
		return s, false, &ContentRejection{Code: "tuf_metadata_invalid", CheckID: "tuf_schema", CandidateKey: m.Key, CandidateVersion: m.Version}
	}
	rows, err := c.pool.Query(ctx, `SELECT root_version,canonical_bytes FROM reference_pack_roots WHERE repository_id=$1 AND root_version <= $2 AND root_version=ANY($3) ORDER BY root_version`, id, current, versions)
	if err != nil {
		return s, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var v int64
		var data []byte
		if err := rows.Scan(&v, &data); err != nil {
			return s, false, err
		}
		s.RootHistory[v] = data
		if v == current {
			s.Root = data
		}
	}
	if err := rows.Err(); err != nil {
		return s, false, err
	}
	if len(s.Root) == 0 {
		return s, false, errors.New("reference pack: frozen root lost")
	}
	if m.Envelope != nil && m.Envelope.TrustProposal != nil {
		for role, metadata := range m.Envelope.TrustProposal.Metadata {
			s.Highest[role] = metadata
		}
	}
	return s, true, nil
}
