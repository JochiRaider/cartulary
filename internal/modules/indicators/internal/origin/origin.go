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

type ProducerContext struct {
	origin        ObservationOrigin
	trustedSystem bool
}

func ManualEntryProducer() ProducerContext {
	return ProducerContext{origin: ManualEntry}
}

func ClipboardPasteProducer() ProducerContext {
	return ProducerContext{origin: ClipboardPaste}
}

func CSVImportProducer() ProducerContext {
	return ProducerContext{origin: CSVImport}
}

func XLSXImportProducer() ProducerContext {
	return ProducerContext{origin: XLSXImport}
}

func APIImportProducer() ProducerContext {
	return ProducerContext{origin: APIImport}
}

func ExtractionProducer() ProducerContext {
	return ProducerContext{origin: Extraction}
}

// TrustedSystemProducer is intentionally available only inside the Indicator
// owner tree. The root facade does not expose a corresponding constructor.
func TrustedSystemProducer() ProducerContext {
	return ProducerContext{origin: System, trustedSystem: true}
}

func (context ProducerContext) OriginForWrite() (ObservationOrigin, error) {
	origin, err := Parse(context.origin.String())
	if err != nil || (origin == System && !context.trustedSystem) {
		return "", &ValidationError{Value: context.origin.String()}
	}
	return origin, nil
}
