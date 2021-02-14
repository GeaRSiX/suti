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

type data interface{}

// LoadDataFiles TODO
func LoadDataFiles(paths ...string) map[string]data {
	var err error
	var stat os.FileInfo
	var d data

	loaded := make(map[string]data)

	for _, dpath := range paths {
		if stat, err = os.Stat(dpath); err != nil {
			warn("skipping data file '%s' (%s)", dpath, err)
			continue
		}

		if stat.IsDir() {
			_ = filepath.Walk(dpath,
				func(path string, info os.FileInfo, e error) error {
					if e == nil && !info.IsDir() {
						if d, e = LoadDataFile(path); e == nil {
							loaded[path] = d
						}
					}

					if e != nil {
						warn("skipping data file '%s' (%s)", path, e)
					}

					return e
				})
		} else {
			if d, err = LoadDataFile(dpath); err == nil {
				loaded[dpath] = d
			} else {
				warn("skipping data file '%s' (%s)", dpath, err)
			}
		}
	}

	return loaded
}

// LoadDataFile TODO
func LoadDataFile(path string) (d data, e error) {
	var f *os.File

	f, e = os.Open(path)
	defer f.Close()
	if e == nil {
		dtype := strings.TrimPrefix(filepath.Ext(path), ".")
		d, e = LoadData(dtype, f)
	}

	return d, e
}

// LoadData TODO
func LoadData(lang string, in io.Reader) (d data, e error) {
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
		e = fmt.Errorf("%s is not a supported data language", lang)
	}

	return
}
