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
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"testing"
)

var dataExts = []string{
	".json", "json", "JSON", ".JSON",
	".yaml", "yaml", "YAML", ".YAML",
	".toml", "toml", "TOML", ".TOML",
	".misc", "-", ".", "",
}

func TestIsDataFormat(t *testing.T) {
	for i, ext := range dataExts {
		var target bool
		
		if i < 12 {
			target = true
		}

		is := IsDataFormat(ext)
		if is != target {
			t.Fatalf("%t did not return %t", is, target)
		}
	}
}

func TestReadDataFormat(t *testing.T) {
	for i, ext := range dataExts {
		var target DataFormat

		if i < 4 {
			target = JSON
		} else if i < 8 {
			target = YAML
		} else if i < 12 {
			target = TOML
		}

		fmt := ReadDataFormat(ext)
		if fmt != target {
			if target == "" {
				t.Fatalf("%s is not a supported data language", fmt)
			} else {
				t.Fatalf("%s did not return %s", fmt, target)
			}
		}
	}
}

var good = map[DataFormat]string{
	JSON: `{"eg":0}`,
	YAML: `eg: 0
`,
	TOML: `eg = 0
`,
}

const badData = `{"json"!:2:]}}`

func writeTestFile(t *testing.T, path string, Data string) {
	f, e := os.Create(path)
	defer f.Close()
	if e != nil {
		t.Skipf("setup failure: %s", e)
	}
	_, e = f.WriteString(Data)
	if e != nil {
		t.Skipf("setup failure: %s", e)
	}

	return
}

func validateData(t *testing.T, d interface{}, e error, lang DataFormat) {
	var b []byte

	if e != nil {
		t.Fatal(e)
	}

	switch lang {
	case JSON:
		b, e = json.Marshal(d)
	case YAML:
		b, e = yaml.Marshal(d)
	case TOML:
		b, e = toml.Marshal(d)
	}

	if e != nil {
		t.Error(e)
	}
	if string(b) != good[lang] {
		t.Fatalf("incorrect %s: %s does not match %s", lang, b, good[lang])
	}
}

func TestLoadData(t *testing.T) {
	var d interface{}
	var e error

	for lang, data := range good {
		e = LoadData(lang, strings.NewReader(data), &d)
		validateData(t, d, e, lang)
	}

	if e = LoadData(JSON, strings.NewReader(badData), &d); e == nil {
		t.Fatalf("bad data passed")
	}
	if e = LoadData(TOML, strings.NewReader(""), &d); e != nil {
		t.Fatalf("empty data failed %s, %s", d, e)
	}
	if e = LoadData("", strings.NewReader("shouldn't pass"), &d); e == nil {
		t.Fatalf("invalid data language passed")
	}

	return
}

func TestLoadDataFile(t *testing.T) {
	var d interface{}
	var e error
	var p string
	tdir := os.TempDir()

	for lang, data := range good {
		p = tdir + "/good." + string(lang)
		writeTestFile(t, p, data)
		e = LoadDataFile(p, &d)
		validateData(t, d, e, lang)
	}

	p = tdir + "/bad.json"
	writeTestFile(t, p, badData)
	e = LoadDataFile(p, &d)
	if e == nil {
		t.Fatalf("bad data passed")
	}

	p = tdir + "/empty.json"
	writeTestFile(t, p, "")
	e = LoadDataFile(p, &d)
	if e != nil {
		t.Fatalf("empty file failed: %s", e)
	}

	if e = LoadDataFile("non-existing-file.toml", &d); e == nil {
		t.Fatalf("non-existing file passed: %s, %s", d, e)
	}

	return
}
