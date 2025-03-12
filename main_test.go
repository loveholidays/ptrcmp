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
	"strings"
	"testing"
)

func TestPointerComparisonFinderWorking(t *testing.T) {
	results, err := parseDir("./tests")
	assert.Nil(t, err)
	assert.Equal(t, len(results), 1)
	assert.True(t, strings.Contains(results[0], "ptrcmp/tests/with_pointer_comparison.go:25:5: comparing pointers to basic types: int and int\n"))
}
