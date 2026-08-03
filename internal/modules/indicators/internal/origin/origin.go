package origin

import (
	"errors"
	"fmt"
)

type ObservationOrigin string

const (
	ManualEntry    ObservationOrigin = "manual_entry"
	ClipboardPaste ObservationOrigin = "clipboard_paste"
	CSVImport      ObservationOrigin = "csv_import"
	XLSXImport     ObservationOrigin = "xlsx_import"
	APIImport      ObservationOrigin = "api_import"
	Extraction     ObservationOrigin = "extraction"
	System         ObservationOrigin = "system"
)

var ErrInvalidObservationOrigin = errors.New("indicators: invalid observation origin")

type ValidationError struct {
	Value string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid indicator observation origin %q", e.Value)
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidObservationOrigin
}

func Parse(raw string) (ObservationOrigin, error) {
	origin := ObservationOrigin(raw)
	switch origin {
	case ManualEntry, ClipboardPaste, CSVImport, XLSXImport, APIImport, Extraction, System:
		return origin, nil
	default:
		return "", &ValidationError{Value: raw}
	}
}

func (origin ObservationOrigin) String() string {
	return string(origin)
}
