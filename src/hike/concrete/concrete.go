package concrete

import (
	abs "hike/abstract"
)

type ArtifactBase struct {
	Key abs.ArtifactKey
	Name string
}

func (artifact *ArtifactBase) ArtifactKey() abs.ArtifactKey {
	return artifact.Key
}

type FileArtifact struct {
	ArtifactBase
	Path string
	GeneratingTransform *abs.Transform
}

func (artifact *FileArtifact) DisplayName() string {
	if len(artifact.Name) > 0 {
		return artifact.Name
	} else {
		return artifact.Path
	}
}

func (artifact *FileArtifact) PathNames(sink []string) []string {
	return append(sink, artifact.Path)
}

func (artifact *FileArtifact) Flatten() abs.BuildError {
	return nil
}

func (artifact *FileArtifact) Require(plan *abs.Plan) abs.BuildError {
	//TODO
	return nil
}

var _ abs.Artifact = &FileArtifact{}
