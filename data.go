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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// DataFormat provides a list of supported languages for
// data files (lower-case)
type DataFormat string

const (
	JSON DataFormat = "json"
	YAML DataFormat = "yaml"
	TOML DataFormat = "toml"
)

// IsDataFile checks if `path` is one of the known *DatFormat*s.
func IsDataFormat(path string) bool {
	return ReadDataFormat(path) != ""
}

// ReadDataFormat returns the *DataFormat* that the file
// extension of `path` matches. If the file extension of `path` does
// not match any *DataFormat*, then an "" is returned.
func ReadDataFormat(path string) DataFormat {
	if len(path) == 0 {
		return ""
	}

	ext := filepath.Ext(path)
	if len(ext) == 0 {
		ext = path // assume `path` the name of the format
	}

	ext = strings.ToLower(ext)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}

	for _, fmt := range []DataFormat{JSON, YAML, TOML} {
		if string(fmt) == ext {
			return fmt
		}
	}
	return ""
}

// LoadData attempts to load all data from `in` as `format` and writes
// the result in the pointer `out`.
func LoadData(format DataFormat, in io.Reader, out interface{}) error {
	inbuf, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	} else if len(inbuf) == 0 {
		return nil
	}

	switch format {
	case JSON:
		err = json.Unmarshal(inbuf, out)
	case YAML:
		err = yaml.Unmarshal(inbuf, out)
	case TOML:
		err = toml.Unmarshal(inbuf, out)
	default:
		err = fmt.Errorf("'%s' is not a supported data language", format)
	}

	return err
}

// LoadDataFile loads all the data from the file found at `path` into
// the the format of that files extension (e.g. "x.json" will be loaded
// as a json). The result is written to the value pointed at by `outp`.
func LoadDataFile(path string, outp interface{}) error {
	file, err := os.Open(path)
	defer file.Close()

	if err == nil {
		if err = LoadData(ReadDataFormat(path), file, outp); err != nil {
			err = fmt.Errorf("failed to load data '%s': %s", path, err.Error())
		}
	}

	return err
}

// WriteData attempts to write `data` as `format` to `outp`.
func WriteData(format DataFormat, data interface{}, w io.Writer) error {
	var err error

	switch format {
	case JSON:
		err = json.NewEncoder(w).Encode(data)
	case YAML:
		err = yaml.NewEncoder(w).Encode(data)
	case TOML:
		err = toml.NewEncoder(w).Encode(data)
	default:
		err = fmt.Errorf("'%s' is not a supported data language", format)
	}

	return err
}

// WriteDataFile attempts to write `data` as `format` to the file at `path`.
// If `force` is *true*, then any existing files will be overwritten.
func WriteDataFile(format DataFormat, data interface{}, path string) (f *os.File, err error) {
	f, err = os.Open(path)
	defer f.Close()

	if err == nil {
		if err = WriteData(format, data, f); err != nil {
			err = fmt.Errorf("faild to write data '%s': %s", path, err.Error())
		}
	}
	if err != nil {
		f = nil
	}

	return
}
