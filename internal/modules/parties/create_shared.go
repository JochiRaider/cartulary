package parties

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
	partysource "github.com/JochiRaider/cartulary/internal/modules/parties/internal/source"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type createValueInput struct {
	present bool
	text    *string
}

type createValueAdmissionError struct {
	field      string
	reasonCode string
}

// admitCreateValues is the sole Party create default/admission routine for
// Workbook and Imports. It iterates the generated registry, supplies optional
// null defaults, and returns one immutable value per field.
func admitCreateValues(inputs map[string]createValueInput) (map[string]policy.Value, *createValueAdmissionError) {
	for fieldKey := range inputs {
		if _, ok := policy.LookupField(fieldKey); !ok {
			return nil, &createValueAdmissionError{field: fieldKey, reasonCode: "unsupported_field_key"}
		}
	}
	values := make(map[string]policy.Value, len(policy.FieldKeys()))
	for _, fieldKey := range policy.FieldKeys() {
		field, _ := policy.LookupField(fieldKey)
		input, present := inputs[fieldKey]
		if !present || !input.present {
			if field.Required {
				return nil, &createValueAdmissionError{field: fieldKey, reasonCode: "missing_required_field"}
			}
			input = createValueInput{present: true}
		}
		value, admissionErr := policy.Admit(fieldKey, input.text)
		if admissionErr != nil {
			return nil, &createValueAdmissionError{field: fieldKey, reasonCode: admissionErr.ReasonCode}
		}
		values[fieldKey] = value
	}
	return values, nil
}

type partyRecordInserter interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
}

type createOrReuseResult struct {
	recordID uuid.UUID
	created  bool
}

// createOrReusePartyTx is the sole caller-transaction Party create body.
// Workbook owns its transaction; Imports borrows its outer unit transaction.
func createOrReusePartyTx(
	ctx context.Context,
	tx pgx.Tx,
	recordsStore partyRecordInserter,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	values map[string]policy.Value,
	now time.Time,
) (createOrReuseResult, error) {
	params := partysource.CreateParams{Values: values}
	recordID, found, err := partysource.FindReusablePartyTx(ctx, tx, incidentID, params)
	if err != nil {
		return createOrReuseResult{}, err
	}
	if found {
		return createOrReuseResult{recordID: recordID}, nil
	}
	recordID, err = recordsStore.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      incidentID,
		RecordType:      "party",
		CreatedByUserID: actorUserID,
		CreatedAt:       now,
		UpdatedByUserID: actorUserID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return createOrReuseResult{}, err
	}
	if err := partysource.InsertPartyTx(ctx, tx, recordID, incidentID, params, now); err != nil {
		return createOrReuseResult{}, err
	}
	if recordID == uuid.Nil {
		return createOrReuseResult{}, fmt.Errorf("parties create: record inserter returned an empty id")
	}
	return createOrReuseResult{recordID: recordID, created: true}, nil
}
