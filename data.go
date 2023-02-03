package dati

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
	"strings"
)

// SupportedDataFormat provides a list of supported languages for
// data files (lower-case)
type SupportedDataFormat string

const (
	JSON SupportedDataFormat = "json"
	YAML SupportedDataFormat = "yaml"
	TOML SupportedDataFormat = "toml"
)

// ReadDataFormat returns the *SupportedDataFormat* that the file
// extension of `path` matches. If the file extension of `path` does
// not match any *SupportedDataFormat*, then an "" is returned.
func ReadDataFormat(path string) SupportedDataFormat {
	if len(path) == 0 {
		return ""
	}

	ext := filepath.Ext(path)
	ext = strings.ToLower(path)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}

	for _, fmt := range []SupportedDataFormat{JSON, YAML, TOML} {
		if string(fmt) == ext {
			return fmt
		}
	}
	return ""
}

// **DEPRECIATED** please use SupportedFormat
var SupportedDataLangs = []string{"json", "yaml", "toml"}

// **DEPRECIATED** please use ReadDataFormat
func IsSupportedDataLang(lang string) int {
	lang = strings.ToLower(lang)
	if len(lang) > 0 && lang[0] == '.' {
		lang = lang[1:]
	}
	for i, l := range []SupportedDataFormat{JSON, YAML, TOML} {
		if lang == string(l) {
			return i
		}
	}
	return -1
}

// LoadData attempts to load all data from `in` as the data language `lang`
// and writes the result in the pointer `outp`.
func LoadData(format string, in io.Reader, outp interface{}) error {
	inbuf, e := ioutil.ReadAll(in)
	if e != nil {
		return e
	} else if len(inbuf) == 0 {
		return nil
	}

	switch IsSupportedDataLang(format) {
	case 0:
		e = json.Unmarshal(inbuf, outp)
	case 1:
		e = yaml.Unmarshal(inbuf, outp)
	case 2:
		e = toml.Unmarshal(inbuf, outp)
	case -1:
		fallthrough
	default:
		e = fmt.Errorf("'%s' is not a supported data language", format)
	}

	return e
}

// LoadDataFile loads all the data from the file found at `path` into the the
// format of that files extension (e.g. "x.json" will be loaded as a json).
// The result is written to the value pointed at by `outp`.
func LoadDataFile(path string, outp interface{}) error {
	f, e := os.Open(path)
	defer f.Close()

	if e == nil {
		lang := filepath.Ext(path)[1:] // don't include '.'
		if e = LoadData(lang, f, outp); e != nil {
			e = fmt.Errorf("failed to load data '%s': %s", path, e.Error())
		}
	}

	return e
}

// **DEPRECIATED** please use LoadDataFile
func LoadDataFilepath(path string, outp interface{}) error {
	return LoadDataFile(path, outp)
}
