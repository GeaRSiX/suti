package suti

/*
Copyright (C) 2021 gearsix <gearsix@tuta.io>

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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// SortFileList sorts `filepath` (a list of filepaths) in `order`. `order`
// can be any of the following values: "filename", "filename-asc",
// "filename-desc", "modified", "modified-asc", "modified-desc"; by default
// "filename" and "modified" are in ascending direction, if specified an
// "-asc" suffix will set the direction to ascending and "-desc" will set the
// direction to descending.
// This was originally intended to be used before calling LoadDataFiles on a
// set of "data" files.
func SortFileList(paths []string, order string) (sorted []string, err error) {
	if order == "filename-desc" {
		sorted = sortFileListByName("desc", paths)
	} else if order == "filename-asc" || order == "filename" {
		sorted = sortFileListByName("asc", paths)
	} else if order == "modified-desc" {
		sorted, err = sortFileListByMod("desc", paths)
	} else if order == "modified-asc" || order == "modified" {
		sorted, err = sortFileListByMod("asc", paths)
	} else {
		err = fmt.Errorf("invalid order '%s'", order)
		sorted = paths
	}
	return
}

func sortFileListByName(direction string, paths []string) []string {
	if direction == "desc" {
		sort.Slice(paths, func(i, j int) bool {
			return filepath.Base(paths[i]) > filepath.Base(paths[j])
		})
	} else {
		sort.Slice(paths, func(i, j int) bool {
			return filepath.Base(paths[i]) < filepath.Base(paths[j])
		})
	}
	return paths
}

func sortFileListByMod(direction string, paths []string) ([]string, error) {
	stats := make(map[string]os.FileInfo)
	for _, p := range paths {
		stat, err := os.Stat(p)
		if err != nil {
			return paths, err
		}
		stats[p] = stat
	}

	modtimes := make([]time.Time, 0, len(paths))
	for _, stat := range stats {
		modtimes = append(modtimes, stat.ModTime())
	}
	if direction == "desc" {
		sort.Slice(modtimes, func(i, j int) bool {
			return modtimes[i].After(modtimes[j])
		})
	} else {
		sort.Slice(modtimes, func(i, j int) bool {
			return modtimes[i].Before(modtimes[j])
		})
	}

	sorted := make([]string, 0)
	for _, t := range modtimes {
		for path, stat := range stats {
			if t == stat.ModTime() {
				sorted = append(sorted, path)
				delete(stats, path)
				break
			}
		}
	}
	if len(sorted) != len(paths) {
		fmt.Errorf("sorted length invalid")
	}

	return sorted, nil
}
