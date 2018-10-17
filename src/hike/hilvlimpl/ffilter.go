package hilvlimpl

import (
	"os"
	"path/filepath"
	herr "hike/error"
	hlv "hike/hilevel"
	con "hike/concrete"
)

// ---------------------------------------- FileTypeFilter ----------------------------------------

type FileTypeFilter struct {
	IsDir bool
}

func NewFileTypeFilter(isDir bool) *FileTypeFilter {
	return &FileTypeFilter {
		IsDir: isDir,
	}
}

func (filter *FileTypeFilter) AcceptFile(fullPath string, baseDir string, info os.FileInfo) bool {
	return info.IsDir() == filter.IsDir
}

func (filter *FileTypeFilter) DumpFilter(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	if filter.IsDir {
		prn.Print("directories")
	} else {
		prn.Print("files")
	}
	return prn.Done()
}

var _ hlv.FileFilter = &FileTypeFilter{}

// ---------------------------------------- WildcardFileFilter ----------------------------------------

type WildcardFileFilter struct {
	Pattern string
}

func NewWildcardFileFilter(pattern string) *WildcardFileFilter {
	return &WildcardFileFilter {
		Pattern: pattern,
	}
}

func (filter *WildcardFileFilter) AcceptFile(fullPath string, baseDir string, info os.FileInfo) bool {
	matched, err := filepath.Match(filter.Pattern, info.Name())
	return err == nil && matched
}

func (filter *WildcardFileFilter) DumpFilter(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Print("wildcard ")
	con.PrintErrorString(prn, filter.Pattern)
	return prn.Done()
}

var _ hlv.FileFilter = &WildcardFileFilter{}

// ---------------------------------------- misc ----------------------------------------

func AllFileFilters(fullPath string, baseDir string, info os.FileInfo, filters []hlv.FileFilter) bool {
	for _, filter := range filters {
		if !filter.AcceptFile(fullPath, baseDir, info) {
			return false
		}
	}
	return true
}
