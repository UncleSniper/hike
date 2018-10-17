package hilevel

import (
	"os"
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

type FileFilter interface {
	AcceptFile(fullPath string, baseDir string, info os.FileInfo) bool
	DumpFilter(level uint) error
}
