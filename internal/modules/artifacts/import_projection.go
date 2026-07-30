package artifacts

import (
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
)

// newArtifactImportProjectionPort keeps projection-provider construction at the
// approved source-owner integration seam. Import mutation itself depends only
// on the narrow artifactProjectionRowPort.
func newArtifactImportProjectionPort() artifactProjectionRowPort {
	return projections.NewArtifactRows(nil, artifactprojection.QuerySurfaces()...)
}
