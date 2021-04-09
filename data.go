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
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Data is the data type used to represent parsed Data (in any format).
type Data map[string]interface{}

func getDataType(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func loadGlobPaths(paths ...string) ([]string, error) {
	var err error
	var glob []string
	for p, path := range paths {
		if strings.Contains(path, "*") {
			if glob, err = filepath.Glob(path); err == nil {
				paths = append(paths, glob...)
				if len(glob) > 0 {
					paths = append(paths[:p], paths[p+1:]...)
				}
			}
		}
	}
	return paths, err
}

// LoadData reads all data from `in` and loads it in the format set in `lang`.
func LoadData(lang string, in io.Reader) (d Data, e error) {
	var inbuf []byte
	if inbuf, e = ioutil.ReadAll(in); e != nil {
		return make(Data), e
	} else if len(inbuf) == 0 {
		return make(Data), nil
	}

	if lang == "json" {
		e = json.Unmarshal(inbuf, &d)
	} else if lang == "yaml" {
		e = yaml.Unmarshal(inbuf, &d)
	} else if lang == "toml" {
		e = toml.Unmarshal(inbuf, &d)
	} else {
		e = fmt.Errorf("'%s' is not a supported data language", lang)
	}

	return
}

// LoadDataFile loads all the data from the file found at `path` into the the
// format of that files file extension (e.g. "x.json" will be loaded as a json).
func LoadDataFile(path string) (d Data, e error) {
	var f *os.File
	if f, e = os.Open(path); e == nil {
		d, e = LoadData(getDataType(path), f)
	}
	f.Close()
	return
}

// LoadDataFiles loads all files in `paths` recursively and sorted them in
// `order`.
func LoadDataFiles(order string, paths ...string) (data []Data, err error) {
	var stat os.FileInfo

	paths, err = loadGlobPaths(paths...)

	loaded := make(map[string]Data)

	for i := 0; i < len(paths) && err == nil; i++ {
		path := paths[i]
		if stat, err = os.Stat(path); err == nil {
			if stat.IsDir() {
				err = filepath.Walk(path,
					func(p string, fi os.FileInfo, e error) error {
						if e == nil && !fi.IsDir() {
							loaded[p], e = LoadDataFile(p)
						}
						return e
					})
			} else {
				loaded[path], err = LoadDataFile(path)
			}
		}
	}

	if err == nil {
		data, err = sortFileData(loaded, order)
	}

	return data, err
}

func sortFileData(data map[string]Data, order string) ([]Data, error) {
	var err error
	sorted := make([]Data, 0, len(data))

	if strings.HasPrefix(order, "filename") {
		if order == "filename-desc" {
			sorted = sortFileDataFilename("desc", data)
		} else if order == "filename-asc" {
			sorted = sortFileDataFilename("asc", data)
		} else {
			sorted = sortFileDataFilename("asc", data)
		}
	} else if strings.HasPrefix(order, "modified") {
		if order == "modified-desc" {
			sorted, err = sortFileDataModified("desc", data)
		} else if order == "modified-asc" {
			sorted, err = sortFileDataModified("asc", data)
		} else {
			sorted, err = sortFileDataModified("asc", data)
		}
	} else {
		for _, d := range data {
			sorted = append(sorted, d)
		}
	}

	return sorted, err
}

func sortFileDataFilename(direction string, data map[string]Data) []Data {
	sorted := make([]Data, 0, len(data))

	fnames := make([]string, 0, len(data))
	for fpath := range data {
		fnames = append(fnames, filepath.Base(fpath))
	}
	sort.Strings(fnames)

	if direction == "desc" {
		for i := len(fnames) - 1; i >= 0; i-- {
			for fpath, d := range data {
				if fnames[i] == filepath.Base(fpath) {
					sorted = append(sorted, d)
				}
			}
		}
	} else {
		for _, fname := range fnames {
			for fpath, d := range data {
				if fname == filepath.Base(fpath) {
					sorted = append(sorted, d)
				}
			}
		}
	}

	return sorted
}

func sortFileDataModified(direction string, data map[string]Data) ([]Data, error) {
	sorted := make([]Data, 0, len(data))

	stats := make(map[string]os.FileInfo)
	for fpath := range data {
		if stat, err := os.Stat(fpath); err != nil {
			return nil, err
		} else {
			stats[fpath] = stat
		}
	}

	modtimes := make([]time.Time, 0, len(data))
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

	for _, t := range modtimes {
		for fpath, stat := range stats {
			if stat.ModTime() == t {
				sorted = append(sorted, data[fpath])
				delete(stats, fpath)
				break
			}
		}
	}

	return sorted, nil
}

// GenerateSuperData simply addeds a key named `datakey` to `global` and assigns `data` to it
func GenerateSuperData(datakey string, global Data, data []Data) (super Data, err error) {
	super = global

	if len(datakey) == 0 {
		datakey = "data"
	}

	if super[datakey] != nil {
		err = fmt.Errorf("datakey '%s' already exists", datakey)
	} else {
		super[datakey] = data
	}

	return
}

// MergeData combines all keys in `data` into a single Data object. If there's
// a conflict (duplicate key), the first found value is kept and the conflicting
// values are ignored.
func MergeData(data ...Data) (merged Data, conflicts []string) {
	merged = make(Data)
	for _, d := range data {
		for key, val := range d {
			if merged[key] == nil {
				merged[key] = val
			} else {
				conflicts = append(conflicts, key)
			}
		}
	}
	return
}
