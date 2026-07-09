package links

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links/valuecodec"
)

func (s *Store) LoadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	value, err := s.LoadRecordLinkMutationValueTx(ctx, tx, recordLinkID)
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func (s *Store) LoadRecordLinkMutationValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (valuecodec.RecordLinkMutationValue, error) {
	value, err := valuecodec.LoadRecordLinkMutationValueTx(ctx, tx, recordLinkID)
	if errors.Is(err, pgx.ErrNoRows) {
		return valuecodec.RecordLinkMutationValue{}, ErrRecordLinkNotFound
	}
	if err != nil {
		return valuecodec.RecordLinkMutationValue{}, fmt.Errorf("load record link value: %w", err)
	}
	return value, nil
}

func (s *Store) LoadRecordTagValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (map[string]any, error) {
	value, err := s.LoadRecordTagMutationValueTx(ctx, tx, recordTagID)
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func (s *Store) LoadRecordTagMutationValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (valuecodec.RecordTagMutationValue, error) {
	value, err := valuecodec.LoadRecordTagMutationValueTx(ctx, tx, recordTagID)
	if errors.Is(err, pgx.ErrNoRows) {
		return valuecodec.RecordTagMutationValue{}, ErrTagNotFound
	}
	if err != nil {
		return valuecodec.RecordTagMutationValue{}, fmt.Errorf("load record tag value: %w", err)
	}
	return value, nil
}

func formatMutationTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatMutationUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}
