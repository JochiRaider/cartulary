package reporting

import (
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/outputcontract"
)

type renderOutputOptions = outputcontract.Options

func parseRenderOutputOptions(raw json.RawMessage, kind, scope string) (renderOutputOptions, error) {
	options, err := outputcontract.DecodeAdmitted(raw, kind, scope)
	if err != nil {
		return renderOutputOptions{}, newRenderValidationError("export_model_invalid", err.ReasonCode)
	}
	return options, nil
}
