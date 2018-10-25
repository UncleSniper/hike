package generic

import (
	herr "hike/error"
	abs "hike/abstract"
	con "hike/concrete"
)

// ---------------------------------------- Step ----------------------------------------

type ZipStep struct {
	con.StepBase
	Arise *herr.AriseRef
}

func(zip *ZipStep) Perform() herr.BuildError {
	//TODO
	return nil
}

var _ abs.Step = &ZipStep{}

// ---------------------------------------- Transform ----------------------------------------

type ZipPiece struct {
	Sources []abs.Artifact
	RebaseFrom string
	RebaseTo string
}

type ZipTransform struct {
	con.TransformBase
	Pieces []ZipPiece
}

func (zip *ZipTransform) Plan(destination abs.Artifact, plan *abs.Plan) herr.BuildError {
	//TODO
	return nil
}

func (zip *ZipTransform) DumpTransform(level uint) error {
	//TODO
	return nil
}

var _ abs.Transform = &ZipTransform{}
