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
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"testing"
)

var good = map[string]string{
	"json": `{"eg":0}`,
	"yaml": `eg: 0
`,
	"toml": `eg = 0
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

func validateData(t *testing.T, d interface{}, e error, lang string) {
	var b []byte

	if e != nil {
		t.Error(e)
	}

	switch lang {
	case "json":
		b, e = json.Marshal(d)
	case "yaml":
		b, e = yaml.Marshal(d)
	case "toml":
		b, e = toml.Marshal(d)
	}

	if e != nil {
		t.Error(e)
	}
	if string(b) != good[lang] {
		t.Errorf("incorrect %s: %s does not match %s", lang, b, good[lang])
	}
}

func TestLoadData(t *testing.T) {
	var d interface{}
	var e error

	for lang, data := range good {
		e = LoadData(lang, strings.NewReader(data), &d)
		validateData(t, d, e, lang)
	}

	if e = LoadData("json", strings.NewReader(badData), &d); e == nil {
		t.Errorf("bad data passed")
	}
	if e = LoadData("toml", strings.NewReader(""), &d); e != nil {
		t.Errorf("empty data failed %s, %s", d, e)
	}
	if e = LoadData("void", strings.NewReader("shouldn't pass"), &d); e == nil {
		t.Errorf("invalid data language passed: %s, %s", d, e)
	}

	return
}

func TestLoadDataFile(t *testing.T) {
	var d interface{}
	var e error
	var p string
	tdir := t.TempDir()

	for lang, data := range good {
		p = tdir + "/good." + lang
		writeTestFile(t, p, data)
		e = LoadDataFile(p, &d)
		validateData(t, d, e, lang)
	}

	p = tdir + "/bad.json"
	writeTestFile(t, p, badData)
	e = LoadDataFile(p, &d)
	if e == nil {
		t.Errorf("bad data passed")
	}

	p = tdir + "/empty.json"
	writeTestFile(t, p, "")
	e = LoadDataFile(p, &d)
	if e != nil {
		t.Errorf("empty file failed: %s", e)
	}

	if e = LoadDataFile("non-existing-file.toml", &d); e == nil {
		t.Errorf("non-existing file passed: %s, %s", d, e)
	}

	return
}
