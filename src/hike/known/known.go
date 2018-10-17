package known

import (
	prs "hike/parser"
	syn "hike/syntax"
)

func RegisterCoreStructures(known *prs.KnownStructures) {
	known.RegisterTopParser("goal", syn.TopGoal)
	known.RegisterTopParser("artifact", syn.ToplevelArtifact)
	known.RegisterActionParser("attain", syn.TopAttainAction)
	known.RegisterActionParser("require", syn.TopRequireAction)
	known.RegisterArtifactParser("file", syn.TopFileArtifact)
	known.RegisterArtifactParser("artifacts", syn.TopGroupArtifact)
	known.RegisterArtifactParser("pipeline", syn.ParsePipelineArtifact)
	known.RegisterTransformParser("exec", syn.TopCommandTransform)
	known.RegisterArtifactSetParser("each", syn.ParseArtifactEach)
	known.RegisterArtifactFactoryParser("file", syn.TopStaticFile)
	known.RegisterArtifactFactoryParser("regex", syn.TopRegexFile)
	known.RegisterTransformFactoryParser("exec", syn.TopCommandTransformFactory)
	known.RegisterFileFilterParser("files", syn.TopFilesFileFilter)
	known.RegisterFileFilterParser("directories", syn.TopDirectoriesFileFilter)
	known.RegisterFileFilterParser("wildcard", syn.TopWildcardFileFilter)
}

func RegisterAllKnownStructures(known *prs.KnownStructures) {
	RegisterCoreStructures(known)
}
