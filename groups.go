package main

import "go/ast"

type importType int

const (
	importType_unknown importType = iota
	importType_mixed
	importType_stdlib
	importType_external
	importType_internal
)

type importGroups []importGroup

// lastEnd returns the End() of the last import added to the last group.
func (ig importGroups) lastEnd() int {
	if len(ig) == 0 {
		return 0
	}
	return ig[len(ig)-1].lastEnd()
}

func (ig importGroups) validate(getType getTypeFn) error {
	if len(ig) <= 1 {
		return nil
	}

	for _, gr := range ig {
		if err := gr.validate(getType); err != nil {
			return err
		}
	}

	used := map[importType]bool{}
	for _, gr := range ig {
		grType := gr.grType(getType)
		if used[grType] {
			return ErrDuplicateGroupType
		}
		used[grType] = true
	}

	return nil
}

type importGroup []*ast.ImportSpec

// lastEnd returns the End() of the last import.
func (ig importGroup) lastEnd() int {
	if len(ig) == 0 {
		return 0
	}
	return int(ig[len(ig)-1].End())
}

// type determines if all imports in the group have the same type (stdlib, 3rd party, "own") or they are mixed.
func (ig importGroup) grType(getType getTypeFn) importType {
	if len(ig) == 0 {
		return importType_unknown
	}

	t := getType(ig[0])
	for _, imp := range ig {
		if getType(imp) != t {
			return importType_mixed
		}
	}

	return t
}

func (ig importGroup) validate(getType getTypeFn) error {
	if ig.grType(getType) == importType_mixed {
		return ErrInvalidGrouping
	}
	return nil
}
