package indicatorassembly

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type projectionRows interface {
	RefreshRowTx(context.Context, pgx.Tx, string, uuid.UUID) error
	LoadRowTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
}

type sourceTextPort struct {
	projections projectionRows
}

func NewSourceTextPort(projections projectionRows) indicators.SourceTextPort {
	return sourceTextPort{projections: projections}
}

func (port sourceTextPort) LoadTextTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, recordType string, fieldKey string) (indicators.SourceTextValue, error) {
	viewSchemaID, err := sourceTextViewSchema(recordType, fieldKey)
	if err != nil {
		return indicators.SourceTextValue{}, err
	}
	if err := port.projections.RefreshRowTx(ctx, tx, viewSchemaID, recordID); err != nil {
		return indicators.SourceTextValue{}, fmt.Errorf("refresh source text projection: %w", err)
	}
	row, err := port.projections.LoadRowTx(ctx, tx, viewSchemaID, recordID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return indicators.SourceTextValue{}, indicators.ErrSourceTextUnavailable
		}
		return indicators.SourceTextValue{}, fmt.Errorf("load source text projection: %w", err)
	}
	text, ok := projectedText(row, fieldKey)
	if !ok {
		return indicators.SourceTextValue{}, indicators.ErrSourceTextUnavailable
	}
	return indicators.SourceTextValue{ViewSchemaID: viewSchemaID, Text: text, Row: row}, nil
}

func (port sourceTextPort) LoadRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, recordType string, fieldKey string) (map[string]any, error) {
	viewSchemaID, err := sourceTextViewSchema(recordType, fieldKey)
	if err != nil {
		return nil, err
	}
	row, err := port.projections.LoadRowTx(ctx, tx, viewSchemaID, recordID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, indicators.ErrSourceTextUnavailable
	}
	if err != nil {
		return nil, fmt.Errorf("load source projection: %w", err)
	}
	return row, nil
}

func (port sourceTextPort) RefreshAndLoadRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, recordType string, fieldKey string) (map[string]any, error) {
	viewSchemaID, err := sourceTextViewSchema(recordType, fieldKey)
	if err != nil {
		return nil, err
	}
	if err := port.projections.RefreshRowTx(ctx, tx, viewSchemaID, recordID); err != nil {
		return nil, fmt.Errorf("refresh source text projection: %w", err)
	}
	row, err := port.projections.LoadRowTx(ctx, tx, viewSchemaID, recordID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, indicators.ErrSourceTextUnavailable
	}
	if err != nil {
		return nil, fmt.Errorf("load refreshed source text projection: %w", err)
	}
	return row, nil
}

func sourceTextViewSchema(recordType string, fieldKey string) (string, error) {
	var matches []string
	for _, resource := range viewschema.ListPublicResources() {
		if !slices.Contains(resource.SourceRecordTypes, recordType) {
			continue
		}
		for _, field := range resource.Fields {
			if field.FieldKey == fieldKey && field.ReadKind == "text" {
				matches = append(matches, resource.ViewSchemaID)
				break
			}
		}
	}
	if len(matches) != 1 {
		return "", indicators.ErrSourceTextUnavailable
	}
	return matches[0], nil
}

func projectedText(row map[string]any, fieldKey string) (string, bool) {
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return "", false
	}
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return "", false
	}
	value, ok := cell["value"].(string)
	return value, ok
}
