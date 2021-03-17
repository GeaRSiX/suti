package main

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
	"json": `{"json":0}`,
	"yaml": `yaml: 0
`,
	"toml": `toml = 0
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

func validateFileData(t *testing.T, d []Data, dlen int, orderedLangs ...string) {
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
	time.Sleep(100 * time.Millisecond)
	p = append(p, tdir+"/bad.json")
	writeTestFile(t, p[len(p)-1], badData)

	d = LoadDataFiles("filename", tdir)
	validateFileData(t, d, len(p)-1, "yaml", "json", "toml")

	d = LoadDataFiles("filename-desc", tdir+"/*")
	validateFileData(t, d, len(p)-1, "toml", "json", "yaml")

	d = LoadDataFiles("modified", p...)
	validateFileData(t, d, len(p)-1, "yaml", "json", "toml")

	d = LoadDataFiles("modified-desc", p...)
	validateFileData(t, d, len(p)-1, "toml", "json", "yaml")
}

func TestGenerateSuperData(t *testing.T) {
	var data Data
	var e error
	var gd []Data
	var d []Data
	var sd Data

	if data, e = LoadData("json", strings.NewReader(good["json"])); e == nil {
		gd = append(gd, data)
	} else {
		t.Skip("setup failure:", e)
	}
	if data, e = LoadData("json", strings.NewReader(good["json"])); e == nil { // test duplicate
		gd = append(gd, data)
	} else {
		t.Skip("setup failure:", e)
	}
	if data, e = LoadData("yaml", strings.NewReader(good["yaml"])); e == nil {
		d = append(d, data)
	} else {
		t.Skip("setup failure:", e)
	}

	sd = GenerateSuperData("testdata", d, gd...)
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

	m = MergeData(d...)
	if m["json"] == nil || m["yaml"] == nil {
		t.Error("missing global keys")
	}
}
