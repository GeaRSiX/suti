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
	"time"
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

func validateData(t *testing.T, d Data, e error, lang string) {
	var b []byte

	if e != nil {
		t.Error(e)
	}
	if len(d) == 0 {
		t.Error("no data loaded")
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
	var d Data
	var e error

	for lang, data := range good {
		d, e = LoadData(lang, strings.NewReader(data))
		validateData(t, d, e, lang)

		if d, e = LoadData(lang, strings.NewReader(badData)); e == nil || len(d) > 0 {
			t.Errorf("bad %s passed", lang)
		}

		if d, e = LoadData(lang, strings.NewReader("")); e != nil {
			t.Errorf("empty file failed for json: %s, %s", d, e)
		}
	}

	if d, e = LoadData("invalid", strings.NewReader("shouldn't pass")); e == nil || len(d) > 0 {
		t.Errorf("invalid data language passed: %s, %s", d, e)
	}

	return
}

func validateFileData(t *testing.T, e error, d []Data, dlen int, orderedLangs ...string) {
	if e != nil {
		t.Error(e)
	}

	if dlen != len(orderedLangs) {
		t.Errorf("invalid orderedLangs length (%d should be %d)", len(orderedLangs), dlen)
	}

	if len(d) != dlen {
		t.Errorf("invalid data length (%d should be %d)", len(d), dlen)
	}

	for i, lang := range orderedLangs {
		validateData(t, d[i], nil, lang)
	}
}

func TestLoadDataFiles(t *testing.T) {
	var e error
	var p []string
	var d []Data
	tdir := t.TempDir()

	p = append(p, tdir+"/1.yaml")
	writeTestFile(t, p[len(p)-1], good["yaml"])
	time.Sleep(100 * time.Millisecond)
	p = append(p, tdir+"/good.json")
	writeTestFile(t, p[len(p)-1], good["json"])
	time.Sleep(100 * time.Millisecond)
	p = append(p, tdir+"/good.toml")
	writeTestFile(t, p[len(p)-1], good["toml"])

	d, e = LoadDataFiles("filename", tdir)
	validateFileData(t, e, d, len(p), "yaml", "json", "toml")

	d, e = LoadDataFiles("filename-desc", tdir+"/*")
	validateFileData(t, e, d, len(p), "toml", "json", "yaml")

	d, e = LoadDataFiles("modified", p...)
	validateFileData(t, e, d, len(p), "yaml", "json", "toml")

	d, e = LoadDataFiles("modified-desc", p...)
	validateFileData(t, e, d, len(p), "toml", "json", "yaml")

	p = append(p, tdir+"/bad.json")
	writeTestFile(t, p[len(p)-1], badData)
	if _, e = LoadDataFiles("modified-desc", p...); e == nil {
		t.Error("bad.json passed")
	}
}

func TestGenerateSuperData(t *testing.T) {
	var data Data
	var e error
	var gd Data
	var d []Data
	var sd Data

	if data, e = LoadData("json", strings.NewReader(good["json"])); e == nil {
		gd = data
	} else {
		t.Skip("setup failure:", e)
	}
	if data, e = LoadData("json", strings.NewReader(good["json"])); e == nil {
		d = append(d, data)
	} else {
		t.Skip("setup failure:", e)
	}
	if data, e = LoadData("yaml", strings.NewReader(good["yaml"])); e == nil {
		d = append(d, data)
	} else {
		t.Skip("setup failure:", e)
	}

	sd, e = GenerateSuperData("testdata", gd, d)
	if e != nil {
		t.Error(e)
	}
	if sd["testdata"] == nil {
		t.Log(sd)
		t.Error("datakey is empty")
	}
	if v, ok := sd["testdata"].([]Data); ok == false {
		t.Log(sd)
		t.Error("unable to infer datakey 'testdata'")
	} else if len(v) < len(data) {
		t.Log(sd)
		t.Error("datakey is missing data")
	}
}

func TestMergeData(t *testing.T) {
	var e error
	var d []Data
	var m Data
	var c []string

	if m, e = LoadData("json", strings.NewReader(good["json"])); e == nil {
		d = append(d, m)
	} else {
		t.Skip("setup failure:", e)
	}
	if m, e = LoadData("json", strings.NewReader(good["json"])); e == nil {
		d = append(d, m)
	} else {
		t.Skip("setup failure:", e)
	}
	if m, e = LoadData("yaml", strings.NewReader(good["yaml"])); e == nil {
		d = append(d, m)
	} else {
		t.Skip("setup failure:", e)
	}

	m, c = MergeData(d...)
	if m["eg"] == nil {
		t.Error("missing global keys")
	}
	if len(c) == 0 {
		t.Errorf("conflicting keys were not reported")
	}
}
