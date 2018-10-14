package hilevel

import (
	herr "hike/error"
	abs "hike/abstract"
)

type TransformFactory interface {
	NewTransform(arise *herr.AriseRef, sources []abs.Artifact)
}

type ArtifactFactory interface {
	NewArtifact(arise *herr.AriseRef, oldArtifact abs.Artifact)
}
