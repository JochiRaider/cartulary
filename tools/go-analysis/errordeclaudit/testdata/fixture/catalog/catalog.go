// Package catalog mirrors the generated operation catalog's literal shape.
package catalog

type Operation struct {
	OperationID string
	ErrorCodes  map[int][]string
}

var catalog = []Operation{
	{OperationID: "getThing", ErrorCodes: map[int][]string{404: []string{"thing_not_found"}, 500: []string{"internal_error"}}},
	{OperationID: "lostThing", ErrorCodes: map[int][]string{500: []string{"internal_error"}}},
	{OperationID: "putThing", ErrorCodes: map[int][]string{404: []string{"thing_not_found"}, 500: []string{"internal_error"}}},
}
