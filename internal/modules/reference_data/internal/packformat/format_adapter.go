package packformat

import (
	"context"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	formatschema "github.com/JochiRaider/cartulary/referencepacks/format/schema"
)

func formatError(err error) error {
	if formatschema.IsLimit(err) {
		return errShapeLimit
	}
	if f, ok := err.(*envelope.Failure); ok {
		return &Failure{Code: f.Code, CheckID: f.CheckID, Details: f.Details}
	}
	return err
}
func formatChecks(program map[string]envelope.CheckFunc) map[string]CheckFunc {
	result := make(map[string]CheckFunc, len(program))
	for id, run := range program {
		result[id] = func(ctx context.Context, emit FindingSink) error { return formatError(run(ctx, emit)) }
	}
	return result
}
