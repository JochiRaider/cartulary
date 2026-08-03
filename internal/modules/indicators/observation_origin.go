package indicators

import indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"

type ObservationOrigin = indicatororigin.ObservationOrigin

const (
	ManualEntryObservationOrigin    ObservationOrigin = indicatororigin.ManualEntry
	ClipboardPasteObservationOrigin ObservationOrigin = indicatororigin.ClipboardPaste
	CSVImportObservationOrigin      ObservationOrigin = indicatororigin.CSVImport
	XLSXImportObservationOrigin     ObservationOrigin = indicatororigin.XLSXImport
	APIImportObservationOrigin      ObservationOrigin = indicatororigin.APIImport
	ExtractionObservationOrigin     ObservationOrigin = indicatororigin.Extraction
	SystemObservationOrigin         ObservationOrigin = indicatororigin.System
)

var ErrInvalidObservationOrigin = indicatororigin.ErrInvalidObservationOrigin

func ParseObservationOrigin(raw string) (ObservationOrigin, error) {
	return indicatororigin.Parse(raw)
}

// ObservationProducerContext binds an observation to a source-authorized
// producer class. Its state is intentionally opaque to ordinary callers.
type ObservationProducerContext struct {
	value indicatororigin.ProducerContext
}

func ManualEntryObservationProducer() ObservationProducerContext {
	return ObservationProducerContext{value: indicatororigin.ManualEntryProducer()}
}

func ClipboardPasteObservationProducer() ObservationProducerContext {
	return ObservationProducerContext{value: indicatororigin.ClipboardPasteProducer()}
}

func CSVImportObservationProducer() ObservationProducerContext {
	return ObservationProducerContext{value: indicatororigin.CSVImportProducer()}
}

func XLSXImportObservationProducer() ObservationProducerContext {
	return ObservationProducerContext{value: indicatororigin.XLSXImportProducer()}
}

func APIImportObservationProducer() ObservationProducerContext {
	return ObservationProducerContext{value: indicatororigin.APIImportProducer()}
}

func ExtractionObservationProducer() ObservationProducerContext {
	return ObservationProducerContext{value: indicatororigin.ExtractionProducer()}
}

func (context ObservationProducerContext) originForWrite() (ObservationOrigin, error) {
	return context.value.OriginForWrite()
}
