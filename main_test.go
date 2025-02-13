/*
ptrcmp
Copyright (C) 2025  loveholidays

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU Lesser General Public
License as published by the Free Software Foundation; either
version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with this program; if not, write to the Free Software Foundation,
Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/
package main

import (
	"github.com/stretchr/testify/assert"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestPointerComparisonFinderWorking(t *testing.T) {
	fset := token.NewFileSet()
	pathWithPointer := "./tests/with_pointer_comparison.go"
	finder := NewPointerComparisonFinder(fset)
	file, err := parser.ParseFile(fset, pathWithPointer, nil, parser.ParseComments)
	assert.Nil(t, err)
	err = finder.checkFile(pathWithPointer, file)
	assert.Nil(t, err)
	assert.Len(t, finder.issues, 1)
	assert.True(t, strings.Contains("tests/with_pointer_comparison.go:7: Direct pointer comparison found. Consider comparing the dereferenced values instead.", finder.issues[0].message))
}

func TestNoPointerComparisons(t *testing.T) {
	fset := token.NewFileSet()
	pathWithPointer := "./tests/without_pointer_comparison.go"
	finder := NewPointerComparisonFinder(fset)
	file, err := parser.ParseFile(fset, pathWithPointer, nil, parser.ParseComments)
	assert.Nil(t, err)
	err = finder.checkFile(pathWithPointer, file)
	assert.Nil(t, err)
	assert.Len(t, finder.issues, 0)
}
