package known

import (
	prs "hike/parser"
	syn "hike/syntax"
)

func RegisterCoreStructures(known *prs.KnownStructures) {
	known.RegisterTopParser("goal", syn.TopGoal)
	known.RegisterActionParser("attain", syn.TopAction)
}

func RegisterAllKnownStructures(known *prs.KnownStructures) {
	RegisterCoreStructures(known)
}
