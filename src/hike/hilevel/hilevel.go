package hilevel

import (
	herr "hike/error"
	spc "hike/spec"
	abs "hike/abstract"
)

type TransformFactory interface {
	NewTransform(sources []abs.Artifact, state *spc.State) (abs.Transform, herr.BuildError)
}

type ArtifactFactory interface {
	NewArtifact(oldArtifacts []abs.Artifact, state *spc.State) (abs.Artifact, herr.BuildError)
	RequiresMerge() bool
}
