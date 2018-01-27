package main

import (
	stderr "errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var (
	errNoGOPATH      = stderr.New("GOPATH is not set")
	errOutsideGOPATH = stderr.New("current directory is outside GOPATH")
	errTooShallow    = stderr.New("you need to be at least two level deep into GOPATH to check import grouping")

	ErrInvalidGrouping    = stderr.New("invalid grouping")
	ErrTooManyGroups      = stderr.New("too many import groups")
	ErrDuplicateGroupType = stderr.New("multiple import groups of the same type")
)

func main() {
	// Find root package to determine what is an "own" package.
	root, err := rootPackage()
	if err != nil {
		fmt.Printf("Failed to determine root package: %v", err)
		os.Exit(1)
	}
	isOwn := mkIsOwnImport(root)
	getType := mkGetType(isOwn)

	// Check if the return error should be non-zero.
	var hasErr bool

	// Walk through all files an subdirectories.
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		// Check only *.go files.
		if !isGoFile(info) {
			return nil
		}

		// Skip vendor directory.
		if isVendor(path) {
			return filepath.SkipDir
		}

		// Check import grouping, print errors next to the filepath.
		// Also, swich flag to indicate non-zero exit code, if necessary.
		if err := processFile(path, getType); err != nil {
			fmt.Printf("%s: %v\n", path, err)
			hasErr = true
		}

		return nil
	})

	if hasErr {
		os.Exit(1)
	}
}

func processFile(filename string, getType getTypeFn) error {
	// Open the file.
	in, err := os.Open(filename)
	if err != nil {
		return errors.Errorf("failed to open file: %v", err)
	}
	defer in.Close()

	// Read file contents.
	src, err := ioutil.ReadAll(in)
	if err != nil {
		return errors.Errorf("failed to read file contents: %v", err)
	}

	// Parse file into a token tree.
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, src, parser.ImportsOnly)
	if err != nil {
		return errors.Errorf("failed to parse file: %v", err)
	}

	// Group imports as they are grouped in the file.
	groups := importGroups{}
	for _, imp := range file.Imports {
		if hadEmptyLineBefore(imp, groups) {
			// Create a new group
			groups = append(groups, []*ast.ImportSpec{imp})
			continue
		}

		// Add to the last group that already exists.
		lastGroupIndex := len(groups) - 1
		lastGroup := groups[lastGroupIndex]
		groups[lastGroupIndex] = append(lastGroup, imp)
	}

	// There should be at most three import groups (stdlib, 3rd party, "own").
	if len(groups) > 3 {
		return errors.Wrapf(ErrTooManyGroups, "found %d groups", len(groups))
	}

	// If any of the groups is invalid (has mixed import types), or there are multiple groups of the same type,
	// an error is returned.
	if err := groups.validate(getType); err != nil {
		return err
	}

	return nil
}

// Determine what should be interpreted as "own" package. By default it is a two-level deep path:
// github.com/slomek/go3mports -> github.com/slomek
func rootPackage() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrapf(err, "failed to read current directory")
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", errNoGOPATH
	}

	if !strings.HasPrefix(pwd, gopath) {
		return "", errOutsideGOPATH
	}

	pwd = strings.TrimPrefix(pwd, filepath.Join(gopath, "src", ""))
	pwd = strings.TrimPrefix(pwd, string(filepath.Separator))

	parts := strings.SplitN(pwd, string(filepath.Separator), 3)
	if len(parts) < 1 {
		return "", errTooShallow
	}

	return strings.Join(parts[:2], string(filepath.Separator)), nil
}

func isGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func isVendor(path string) bool {
	return strings.HasPrefix(path, "vendor/")
}

type ownImportFn func(*ast.ImportSpec) bool

func mkIsOwnImport(ownRoot string) ownImportFn {
	return func(i *ast.ImportSpec) bool {
		return strings.HasPrefix(i.Path.Value[1:], ownRoot)
	}
}

type getTypeFn func(*ast.ImportSpec) importType

func mkGetType(isOwn ownImportFn) getTypeFn {
	return func(i *ast.ImportSpec) importType {
		path := i.Path.Value

		if !strings.Contains(path, ".") {
			return importType_stdlib
		}

		if isOwn(i) {
			return importType_internal
		}

		return importType_external
	}
}

// If Pos() of the current import is less than two characters away from the previous' End(),
// they should belong to the same group.
func hadEmptyLineBefore(imp *ast.ImportSpec, groups importGroups) bool {
	return int(imp.Pos())-groups.lastEnd() > 2
}
