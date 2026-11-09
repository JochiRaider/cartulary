package dataset

import (
	"embed"
	"strconv"

	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
)

// These are engine-owned declarations. Pack bytes cannot add a schema or
// resolve a file, URL, expression or executable normalization algorithm.
//
//go:embed schemas/*.json
var schemaFiles embed.FS

var datasetShapes = loadShapes()

func loadShapes() map[string]*schema.Shape {
	resources := map[string][]byte{}
	for _, kind := range []string{"descriptor", "object", "relationship"} {
		name := "dataset." + kind + ".v1.schema.json"
		raw, err := schemaFiles.ReadFile("schemas/" + name)
		if err != nil {
			panic(err)
		}
		resources[name] = raw
	}
	limits := map[string]int{"max_source_refs": 64, "max_dataset_lookups": MaxLookups, "max_dataset_components": MaxComponents, "max_dataset_object_postings": MaxObjectPostings}
	result := map[string]*schema.Shape{}
	for _, kind := range []string{"descriptor", "object", "relationship"} {
		result[kind] = schema.Compile(resources, limits, "dataset."+kind+".v1.schema.json")
	}
	return result
}

func firstFailure(code string) schema.FindingSink {
	return func(f schema.Finding) error {
		err := &Failure{Code: code}
		if f.Details.LimitID != nil {
			err.Limit = *f.Details.LimitID
			if f.Details.ExpectedToken != nil {
				err.Maximum, _ = strconv.ParseUint(*f.Details.ExpectedToken, 10, 64)
			}
			if f.Details.ActualToken != nil {
				err.Observed, _ = strconv.ParseUint(*f.Details.ActualToken, 10, 64)
			}
		}
		return err
	}
}
