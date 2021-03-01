package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
func LoadDataFiles(paths ...string) map[string]Data {
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
			} else if d, err  = LoadDataFile(path); err == nil {
				loaded[path] = d
			} else {
				warn("skipping data file '%s' (%s)", path, err)
			}
		}
	}

	return loaded
}
