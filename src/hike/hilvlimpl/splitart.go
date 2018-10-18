package hilvlimpl

import (
	"os"
	"time"
	herr "hike/error"
	abs "hike/abstract"
	con "hike/concrete"
)

type SplitArtifact struct {
	OwnKey *abs.ArtifactKey
	ID abs.ArtifactID
	Arise *herr.AriseRef
	StartChild abs.Artifact
	EndChild abs.Artifact
	Flipped bool
}

func NewSplitArtifact(
	ownKey *abs.ArtifactKey,
	startChild abs.Artifact,
	endChild abs.Artifact,
	arise *herr.AriseRef,
) *SplitArtifact {
	return &SplitArtifact {
		OwnKey: ownKey,
		ID: abs.NextArtifactID(),
		Arise: arise,
		StartChild: startChild,
		EndChild: endChild,
	}
}

func (split *SplitArtifact) ArtifactKey() *abs.ArtifactKey {
	switch {
		case split.OwnKey != nil:
			return split.OwnKey
		case split.Flipped:
			return split.EndChild.ArtifactKey()
		default:
			return split.StartChild.ArtifactKey()
	}
}

func (split *SplitArtifact) ArtifactID() abs.ArtifactID {
	return split.ID
}

func (split *SplitArtifact) DisplayName() string {
	if split.Flipped {
		return split.EndChild.DisplayName()
	} else {
		return split.StartChild.DisplayName()
	}
}

func (split *SplitArtifact) ArtifactArise() *herr.AriseRef {
	return split.Arise
}

func (split *SplitArtifact) PathNames(sink []string) ([]string, herr.BuildError) {
	if split.Flipped {
		return split.EndChild.PathNames(sink)
	} else {
		return split.StartChild.PathNames(sink)
	}
}

func (split *SplitArtifact) EarliestModTime(arise *herr.AriseRef) (time.Time, herr.BuildError, bool) {
	if split.Flipped {
		return split.EndChild.EarliestModTime(arise)
	} else {
		return split.StartChild.EarliestModTime(arise)
	}
}

func (split *SplitArtifact) LatestModTime(arise *herr.AriseRef) (time.Time, herr.BuildError, bool) {
	if split.Flipped {
		return split.EndChild.LatestModTime(arise)
	} else {
		return split.StartChild.LatestModTime(arise)
	}
}

func (split *SplitArtifact) Flatten() herr.BuildError {
	return nil
}

func (split *SplitArtifact) Require(plan *abs.Plan, arise *herr.AriseRef) herr.BuildError {
	if split.Flipped {
		return split.EndChild.Require(plan, arise)
	} else {
		err := split.StartChild.Require(plan, arise)
		if err == nil {
			split.Flipped = true
		}
		return err
	}
}

func (split *SplitArtifact) DumpArtifact(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Level(level)
	prn.Println("split {")
	prn.Indent(1)
	con.PrintErrorString(prn, split.StartChild.ArtifactKey().Unified())
	prn.Println()
	prn.Indent(1)
	con.PrintErrorString(prn, split.EndChild.ArtifactKey().Unified())
	prn.Println()
	prn.Indent(0)
	prn.Print("}")
	return prn.Done()
}

var _ abs.Artifact = &SplitArtifact{}
