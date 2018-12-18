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

// ---------------------------------------- AnyFileFilter ----------------------------------------

type AnyFileFilter struct {
	Children []hlv.FileFilter
}

func NewAnyFileFilter(children []hlv.FileFilter) *AnyFileFilter {
	return &AnyFileFilter {
		Children: children,
	}
}

func (filter *AnyFileFilter) AcceptFile(fullPath string, baseDir string, info os.FileInfo) bool {
	for _, child := range filter.Children {
		if child.AcceptFile(fullPath, baseDir, info) {
			return true
		}
	}
	return false
}

func (filter *AnyFileFilter) DumpFilter(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Level(level)
	prn.Print("any {")
	for _, child := range filter.Children {
		prn.Println()
		prn.Indent(1)
		prn.Inject(child.DumpFilter, 0)
	}
	if len(filter.Children) > 0 {
		prn.Println()
		prn.Indent(0)
	}
	prn.Print("}")
	return prn.Done()
}

var _ hlv.FileFilter = &AnyFileFilter{}

// ---------------------------------------- AllFileFilter ----------------------------------------

type AllFileFilter struct {
	Children []hlv.FileFilter
}

func NewAllFileFilter(children []hlv.FileFilter) *AllFileFilter {
	return &AllFileFilter {
		Children: children,
	}
}

func (filter *AllFileFilter) AcceptFile(fullPath string, baseDir string, info os.FileInfo) bool {
	for _, child := range filter.Children {
		if !child.AcceptFile(fullPath, baseDir, info) {
			return false
		}
	}
	return true
}

func (filter *AllFileFilter) DumpFilter(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Level(level)
	prn.Print("all {")
	for _, child := range filter.Children {
		prn.Println()
		prn.Indent(1)
		prn.Inject(child.DumpFilter, 0)
	}
	if len(filter.Children) > 0 {
		prn.Println()
		prn.Indent(0)
	}
	prn.Print("}")
	return prn.Done()
}

var _ hlv.FileFilter = &AllFileFilter{}

// ---------------------------------------- AllFileFilter ----------------------------------------

type NotFileFilter struct {
	Child hlv.FileFilter
}

func NewNotFileFilter(child hlv.FileFilter) *NotFileFilter {
	return &NotFileFilter {
		Child: child,
	}
}

func (filter *NotFileFilter) AcceptFile(fullPath string, baseDir string, info os.FileInfo) bool {
	return !filter.Child.AcceptFile(fullPath, baseDir, info)
}

func (filter *NotFileFilter) DumpFilter(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Print("not ")
	prn.Inject(filter.Child.DumpFilter, level)
	return prn.Done()
}

var _ hlv.FileFilter = &NotFileFilter{}

// ---------------------------------------- misc ----------------------------------------

func AllFileFilters(fullPath string, baseDir string, info os.FileInfo, filters []hlv.FileFilter) bool {
	for _, filter := range filters {
		if !filter.AcceptFile(fullPath, baseDir, info) {
			return false
		}
	}
	return true
}
