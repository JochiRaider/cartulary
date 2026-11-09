package envelope

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
)

type Finding = schema.Finding
type SafeDetails = schema.SafeDetails
type FindingSink = schema.FindingSink
type CheckFunc func(context.Context, FindingSink) error

func LimitFinding(path, id string, maximum, observed int64) Finding {
	return schema.LimitFinding(path, id, maximum, observed)
}
func unsignedLimitFinding(path, id string, maximum, observed uint64) Finding {
	return schema.UnsignedLimitFinding(path, id, maximum, observed)
}

var errShapeMismatch = errors.New("reference pack check mismatch")

type check struct {
	ID   string `json:"check_id"`
	Code string `json:"code"`
}

var checks = func() []check {
	var registry struct {
		SchemaID string  `json:"schema_id"`
		Checks   []check `json:"checks"`
	}
	if json.Unmarshal(projection("checks.v1.json"), &registry) != nil || registry.SchemaID != "cartulary.reference_pack_format_checks.v1" {
		panic("reference format: invalid compiled check registry")
	}
	return registry.Checks
}()

// Boolean convenience admission and enumerating adapters run identical checks
// in the same order. Application retention/publication is not part of this API.
func runCheckRange(ctx context.Context, first, last string, program map[string]CheckFunc) error {
	var selected []check
	active := false
	for _, c := range checks {
		if c.ID == first {
			active = true
		}
		if active {
			selected = append(selected, c)
		}
		if active && c.ID == last {
			active = false
			break
		}
	}
	if active || len(selected) == 0 || selected[len(selected)-1].ID != last || len(selected) != len(program) {
		return errors.New("reference format: incomplete check range")
	}
	for _, c := range selected {
		if program[c.ID] == nil {
			return errors.New("reference format: missing range check")
		}
	}
	for _, c := range selected {
		if err := ctx.Err(); err != nil {
			return err
		}
		var finding Finding
		err := program[c.ID](ctx, func(f Finding) error { finding = f; return errShapeMismatch })
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, errShapeMismatch) {
			return &Failure{Code: c.Code, CheckID: c.ID, Details: finding.Details}
		}
		if err != nil {
			return err
		}
	}
	return nil
}
