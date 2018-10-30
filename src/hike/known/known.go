package known

import (
	prs "hike/parser"
	syn "hike/syntax"
)

func RegisterCoreStructures(known *prs.KnownStructures) {
	// TopParser
	known.RegisterTopParser("goal", syn.TopGoal)
	known.RegisterTopParser("artifact", syn.ToplevelArtifact)
	known.RegisterTopParser("set", syn.TopSetVar)
	known.RegisterTopParser("setdef", syn.TopSetVarDef)
	known.RegisterTopParser("include", syn.ParseInclude)
	// ActionParser
	known.RegisterActionParser("attain", syn.TopAttainAction)
	known.RegisterActionParser("require", syn.TopRequireAction)
	known.RegisterActionParser("delete", syn.ParseDeleteAction)
	// ArtifactParser
	known.RegisterArtifactParser("file", syn.TopFileArtifact)
	known.RegisterArtifactParser("artifacts", syn.TopGroupArtifact)
	known.RegisterArtifactParser("pipeline", syn.ParsePipelineArtifact)
	known.RegisterArtifactParser("tree", syn.TopTreeArtifact)
	known.RegisterArtifactParser("split", syn.TopSplitArtifact)
	// TransformParser
	known.RegisterTransformParser("exec", syn.TopCommandTransform)
	known.RegisterTransformParser("copy", syn.TopCopyTransform)
	known.RegisterTransformParser("zip", syn.TopZipTransform)
	known.RegisterTransformParser("unzip", syn.TopUnzipTransform)
	// ArtifactSetParser
	known.RegisterArtifactSetParser("each", syn.ParseArtifactEach)
	known.RegisterArtifactSetParser("scandir", syn.ParseArtifactScanDir)
	// ArtifactFactoryParser
	known.RegisterArtifactFactoryParser("file", syn.TopStaticFile)
	known.RegisterArtifactFactoryParser("regex", syn.TopRegexFile)
	// TransformFactoryParser
	known.RegisterTransformFactoryParser("exec", syn.TopCommandTransformFactory)
	known.RegisterTransformFactoryParser("copy", syn.TopCopyTransformFactory)
	// FileFilterParser
	known.RegisterFileFilterParser("files", syn.TopFilesFileFilter)
	known.RegisterFileFilterParser("directories", syn.TopDirectoriesFileFilter)
	known.RegisterFileFilterParser("wildcard", syn.TopWildcardFileFilter)
}

func RegisterAllKnownStructures(known *prs.KnownStructures) {
	RegisterCoreStructures(known)
}
