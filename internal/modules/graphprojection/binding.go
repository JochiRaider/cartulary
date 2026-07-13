package graphprojection

import "context"

type ProjectionBinding struct {
	ProjectionRunID        string
	GraphViewID            string
	SourceSnapshotID       string
	ProjectionVersion      string
	State                  RunState
	ProjectionConfigDigest string
	ProjectionSourceDigest string
	ProjectionOutputDigest string
}

type ProjectionBindingReader interface {
	LookupProjectionBinding(context.Context, string) (ProjectionBinding, error)
}
