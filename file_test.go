package dati

/*
Copyright (C) 2023 gearsix <gearsix@tuta.io>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestSortFileList(t *testing.T) {
	var err error
	tdir := os.TempDir()
	paths := []string{tdir + "/1", tdir + "/3", tdir + "/2"}
	var sorted []string

	sorted, err = SortFileList(paths, "filename")
	if err != nil {
		t.Fatal(err)
	}
	for i, p := range sorted {
		if filepath.Base(p) != strconv.Itoa(i+1) {
			t.Fatalf("invalid order returned sorted[%d] is %s", i, p)
		}
	}

	sorted, err = SortFileList(paths, "filename-desc")
	if err != nil {
		t.Fatal(err)
	}
	j := 3
	for i := 0; i < len(sorted); i++ {
		if filepath.Base(sorted[i]) != strconv.Itoa(j) {
			t.Fatalf("invalid order returned sorted[%d] is %s", i, sorted[i])
		}
		j--
	}

	for _, path := range paths {
		var f *os.File
		if f, err = os.Create(path); err != nil {
			t.Skip(err)
		}
		defer f.Close()
		time.Sleep(100 * time.Millisecond)
	}

	sorted, err = SortFileList(paths, "modified")
	if err != nil {
		t.Fatal(err)
	}
	for i := range paths {
		if sorted[i] != paths[i] {
			t.Fatalf("invalid order returned %s - %s", sorted, paths)
		}
	}

	sorted, err = SortFileList(paths, "modified-desc")
	if err != nil {
		t.Fatal(err)
	}
	j = 2
	for i := 0; i < len(paths); i++ {
		if sorted[i] != paths[j] {
			t.Fatalf("invalid order returned %s - %s", sorted, paths)
		}
		j--
	}
}
