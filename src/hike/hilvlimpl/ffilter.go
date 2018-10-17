package hilvlimpl

import (
	"os"
	"path/filepath"
	hlv "hike/hilevel"
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
