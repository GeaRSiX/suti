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

// LoadDataFiles TODO
func LoadDataFiles(paths ...string) map[string]Data {
	var err error
	var stat os.FileInfo
	var d Data
	var dtype string
	var f *os.File

	loaded := make(map[string]Data)

	for _, path := range paths {
		if stat, err = os.Stat(path); err != nil {
			warn("skipping data file '%s' (%s)", path, err)
			continue
		}
		if f, err = os.Open(path); err != nil {
			warn("skipping data file '%s' (%s)", path, err)
			continue
		}
		defer f.Close()

		if stat.IsDir() {
			_ = filepath.Walk(path,
				func(p string, fi os.FileInfo, e error) error {
					if e == nil && !fi.IsDir() {
						dtype = getDataType(p)
						if d, e = LoadData(dtype, f); e == nil {
							loaded[p] = d
						}
					}

					if e != nil {
						warn("skipping data file '%s' (%s)", p, e)
					}

					return e
				})
		} else {
			dtype = getDataType(path)
			if d, err = LoadData(dtype, f); err == nil {
				loaded[path] = d
			} else {
				warn("skipping data file '%s' (%s)", path, err)
			}
		}
	}

	return loaded
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
			e = fmt.Errorf("invalid json")
		}
	} else {
		e = fmt.Errorf("'%s' is not a supported data language", lang)
	}

	return
}
