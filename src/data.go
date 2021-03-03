package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
	"path/filepath"
	"strings"
	"sort"
)

type Data map[string]interface{}

func getDataType(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

// LoadData TODO
func LoadData(lang string, in io.Reader) (d Data, e error) {
	var fbuf []byte
	if fbuf, e = ioutil.ReadAll(in); e != nil {
		return
	}

	if lang == "json" {
		if json.Valid(fbuf) {
			e = json.Unmarshal(fbuf, &d)
		} else {
			e = fmt.Errorf("invalid json %s", fbuf)
		}
	} else {
		e = fmt.Errorf("'%s' is not a supported data language", lang)
	}

	return
}

// LoadDataFile TODO
func LoadDataFile(path string) (Data, error) {
	if f, e := os.Open(path); e != nil {
		warn("could not load data file '%s' (%s)", path, e)
		return nil, e
	} else {
		defer f.Close()
		return LoadData(getDataType(path), f)
	}
}

// LoadDataFiles TODO
func LoadDataFiles(order string, paths ...string) []Data {
	var err error
	var stat os.FileInfo
	var d Data

	loaded := make(map[string]Data)

	for p, path := range paths {
		if strings.Contains(path, "*") {
			if glob, e := filepath.Glob(path); e == nil {
				paths = append(paths, glob...)
				paths = append(paths[:p], paths[p+1:]...)
			} else {
				warn("error parsing glob '%s': %s", path, err)
			}
		}
	}

	for _, path := range paths {
		err = nil
		stat, err = os.Stat(path)
		if err == nil {
			if stat.IsDir() {
				err = filepath.Walk(path,
					func(p string, fi os.FileInfo, e error) error {
						if e == nil && !fi.IsDir() {
							if d, e = LoadDataFile(p); e == nil {
								loaded[p] = d
							} else {
								warn("skipping data file '%s' (%s)", p, e)
								e = nil
							}
						}
						return e
					})
			} else if d, err = LoadDataFile(path); err == nil {
				loaded[path] = d
			} else {
				warn("skipping data file '%s' (%s)", path, err)
			}
		}
	}
	
	return sortFileData(loaded, order)
}

func sortFileData(data map[string]Data, order string) []Data {
	sorted := make([]Data, 0, len(data))
	
	if order == "filename" {
		fnames := make([]string, 0, len(data))
		for fpath, _ := range data {
			fnames = append(fnames, filepath.Base(fpath))
		}
		sort.Strings(fnames)
		for _, fname := range fnames {
			 for fpath, d := range data {
				 if fname == filepath.Base(fpath) {
					 sorted = append(sorted, d)
				 }
			 }
		}
	} else if order == "modified" {
		stats := make(map[string]os.FileInfo)
		for fpath, _ := range data {
			if stat, err := os.Stat(fpath); err != nil {
				warn("failed to stat %s (%s)", fpath, err)
			} else {
				stats[fpath] = stat
			}
		}
		
		modtimes := make([]time.Time, 0, len(data))
		for _, stat := range stats {
			modtimes = append(modtimes, stat.ModTime())
		}
		sort.Slice(modtimes, func(i, j int) bool {
			return modtimes[i].Before(modtimes[j])
		})
		
		for _, t := range modtimes {
			for fpath, stat := range stats {
				if t == stat.ModTime() {
					sorted = append(sorted, data[fpath])
				}
			}
		}
	} else {
		warn("unrecognised sort option '%s', data will be unsorted", order)
		for _, d := range data {
			sorted = append(sorted, d)
		}
	}
	
	return sorted
}
 