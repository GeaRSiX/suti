package suti

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestSortFileList(t *testing.T) {
	var err error
	tdir := t.TempDir()
	paths := []string{tdir + "/1", tdir + "/3", tdir + "/2"}
	var sorted []string

	sorted, err = SortFileList(paths, "filename")
	if err != nil {
		t.Error(err)
	}
	for i, p := range sorted {
		if filepath.Base(p) != strconv.Itoa(i+1) {
			t.Errorf("invalid order returned sorted[%d] is %s", i, p)
		}
	}

	sorted, err = SortFileList(paths, "filename-desc")
	if err != nil {
		t.Error(err)
	}
	j := 3
	for i := 0; i < len(sorted); i++ {
		if filepath.Base(sorted[i]) != strconv.Itoa(j) {
			t.Errorf("invalid order returned sorted[%d] is %s", i, sorted[i])
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
		t.Error(err)
	}
	for i := range paths {
		if sorted[i] != paths[i] {
			t.Errorf("invalid order returned %s - %s", sorted, paths)
		}
	}

	sorted, err = SortFileList(paths, "modified-desc")
	if err != nil {
		t.Error(err)
	}
	j = 2
	for i := 0; i < len(paths); i++ {
		if sorted[i] != paths[j] {
			t.Errorf("invalid order returned %s - %s", sorted, paths)
		}
		j--
	}
}
