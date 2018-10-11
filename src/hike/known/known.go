package known

import (
	prs "hike/parser"
	syn "hike/syntax"
)

func RegisterCoreStructures(known *prs.KnownStructures) {
	known.RegisterTopParser("goal", syn.TopGoal)
	known.RegisterActionParser("attain", syn.TopAttainAction)
	known.RegisterActionParser("require", syn.TopRequireAction)
	known.RegisterArtifactParser("file", syn.TopFileArtifact)
	known.RegisterArtifactParser("artifacts", syn.TopGroupArtifact)
	known.RegisterTransformParser("exec", syn.TopCommandTransform)
}

func RegisterAllKnownStructures(known *prs.KnownStructures) {
	RegisterCoreStructures(known)
}
