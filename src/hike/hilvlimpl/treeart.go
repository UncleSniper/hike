package hilvlimpl

import (
	"os"
	"time"
	"path/filepath"
	herr "hike/error"
	hlv "hike/hilevel"
	abs "hike/abstract"
	con "hike/concrete"
)

const (
	TreeArtifactDontCache = iota
	TreeArtifactCachePending
	TreeArtifactCacheFilled
)

type TreeArtifact struct {
	con.ArtifactBase
	Root string
	Filters []hlv.FileFilter
	cachedPaths []string
	cacheState int
	earliestModTime time.Time
	latestModTime time.Time
}

func NewTreeArtifact(
	key abs.ArtifactKey,
	name string,
	arise *herr.AriseRef,
	root string,
	filters []hlv.FileFilter,
	noCache bool,
) *TreeArtifact {
	artifact := &TreeArtifact {
		Root: root,
		Filters: filters,
	}
	artifact.Key = key
	artifact.ID = abs.NextArtifactID()
	artifact.Name = name
	artifact.Arise = arise
	if noCache {
		artifact.cacheState = TreeArtifactDontCache
	} else {
		artifact.cacheState = TreeArtifactCachePending
	}
	return artifact
}

func (artifact *TreeArtifact) DisplayName() string {
	if len(artifact.Name) > 0 {
		return artifact.Name
	} else {
		return artifact.Root
	}
}

func (artifact *TreeArtifact) fillCache() herr.BuildError {
	if artifact.cacheState == TreeArtifactCacheFilled {
		return nil
	}
	artifact.cachedPaths = nil
	artifact.earliestModTime = time.Now()
	artifact.latestModTime = time.Now()
	var have bool
	outerr := filepath.Walk(artifact.Root, func(fullPath string, info os.FileInfo, inerr error) error {
		if inerr != nil {
			return inerr
		}
		if !AllFileFilters(fullPath, artifact.Root, info, artifact.Filters) {
			return nil
		}
		mod := info.ModTime()
		if !have {
			artifact.earliestModTime = mod
			artifact.latestModTime = mod
			have = true
		} else {
			if mod.After(artifact.latestModTime) {
				artifact.latestModTime = mod
			}
			if artifact.earliestModTime.After(mod) {
				artifact.earliestModTime = mod
			}
		}
		artifact.cachedPaths = append(artifact.cachedPaths, fullPath)
		return nil
	})
	if outerr != nil {
		return &FSWalkError {
			RootDir: artifact.Root,
			OSError: outerr,
			WalkArise: artifact.Arise,
		}
	}
	if artifact.cacheState == TreeArtifactCachePending {
		artifact.cacheState = TreeArtifactCacheFilled
	}
	return nil
}

func (artifact *TreeArtifact) PathNames(sink []string) ([]string, herr.BuildError) {
	err := artifact.fillCache()
	if err != nil {
		return nil, err
	}
	return artifact.cachedPaths, nil
}

func (artifact *TreeArtifact) EarliestModTime(arise *herr.AriseRef) (time.Time, herr.BuildError, bool) {
	err := artifact.fillCache()
	return artifact.earliestModTime, err, false
}

func (artifact *TreeArtifact) LatestModTime(arise *herr.AriseRef) (time.Time, herr.BuildError, bool) {
	err := artifact.fillCache()
	return artifact.latestModTime, err, false
}

func (artifact *TreeArtifact) Flatten() herr.BuildError {
	return nil
}

func (artifact *TreeArtifact) Require(plan *abs.Plan, requireArise *herr.AriseRef) herr.BuildError {
	return nil
}

func (artifact *TreeArtifact) DumpArtifact(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Level(level)
	prn.Print("tree ")
	con.PrintErrorString(prn, artifact.Key.Unified())
	prn.Println(" {")
	prn.Indent(1)
	con.PrintErrorString(prn, artifact.Root)
	prn.Println()
	if len(artifact.Name) > 0 {
		prn.Indent(1)
		prn.Print("name ")
		con.PrintErrorString(prn, artifact.Name)
		prn.Println()
	}
	if artifact.cacheState == TreeArtifactDontCache {
		prn.Indent(1)
		prn.Println("noCache")
	}
	for _, filter := range artifact.Filters {
		prn.Indent(1)
		prn.Fail(filter.DumpFilter(level + 1))
		prn.Println()
	}
	prn.Indent(0)
	prn.Print("}")
	return prn.Done()
}

var _ abs.Artifact = &TreeArtifact{}
