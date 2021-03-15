package main

import (
	"encoding/json"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"testing"
)

const goodJson = `{"json":0}`
const goodJson2 = `{"json":1}`
const goodYaml = `yaml: 0
`
const goodToml = `toml = 0
`
const badData = `{"json"!:2:]}}`

func writeTestFile(path string, Data string) (e error) {
	var f *os.File

	if f, e = os.Create(path); e != nil {
		return
	}
	if _, e = f.WriteString(Data); e != nil {
		return
	}
	f.Close()

	return
}

func TestLoadData(t *testing.T) {
	var d Data
	var e error
	var b []byte

	// json
	if d, e = LoadData("json", strings.NewReader(goodJson)); e != nil {
		t.Error(e)
	} else if len(d) == 0 {
		t.Error("no data loaded")
	} else {
		if b, e = json.Marshal(d); e != nil {
			t.Error(e)
		} else if string(b) != goodJson {
			t.Errorf("incorrect json: %s does not match %s", b, goodJson)
		}
	}
	if d, e = LoadData("json", strings.NewReader(badData)); e == nil || len(d) > 0 {
		t.Error("bad json passed")
	}

	// yaml
	if d, e = LoadData("yaml", strings.NewReader(goodYaml)); e != nil {
		t.Error(e)
	} else if len(d) == 0 {
		t.Error("no data loaded")
	} else {
		if b, e = yaml.Marshal(d); e != nil {
			t.Error(e)
		} else if string(b) != goodYaml {
			t.Errorf("incorrect yaml: %s does not match %s", b, goodYaml)
		}
	}
	if d, e = LoadData("yaml", strings.NewReader(badData)); e == nil || len(d) > 0 {
		t.Error("bad yaml passed")
	}

	// toml
	if d, e = LoadData("toml", strings.NewReader(goodToml)); e != nil {
		t.Error(e)
	} else if len(d) == 0 {
		t.Error("no data loaded")
	} else {
		if b, e = toml.Marshal(d); e != nil {
			t.Error(e)
		} else if string(b) != goodToml {
			t.Errorf("incorrect toml: %s does not match %s", b, goodToml)
		}
	}
	if d, e = LoadData("toml", strings.NewReader(badData)); e == nil || len(d) > 0 {
		t.Error("bad toml passed")
	}
	
	// misc
	if d, e = LoadData("json", strings.NewReader("")); e != nil {
		t.Errorf("empty file failed for json: %s, %s", d, e)
	}
	if d, e = LoadData("yaml", strings.NewReader("")); e != nil {
		t.Errorf("empty file failed for yaml: %s, %s", d, e)
	}
	if d, e = LoadData("toml", strings.NewReader("")); e != nil {
		t.Errorf("empty file failed toml: %s, %s", d, e)
	}
	if d, e = LoadData("ebrgji", strings.NewReader(goodJson)); e == nil || len(d) > 0 {
		t.Errorf("invalid data language passed: %s, %s", d, e)
	}
	
	return
}

func TestLoadDataFiles(t *testing.T) {
	var e error
	var p []string
	var b []byte
	var d []Data
	tdir := t.TempDir()

	p = append(p, tdir+"/good.json")
	if e = writeTestFile(p[len(p)-1], goodJson); e != nil {
		t.Skip("setup failure:", e)
	}
	p = append(p, tdir+"/1.yaml")
	if e = writeTestFile(p[len(p)-1], goodYaml); e != nil {
		t.Skip("setup failure:", e)
	}
	p = append(p, tdir+"/good.toml")
	if e = writeTestFile(p[len(p)-1], goodToml); e != nil {
		t.Skip("setup failure:", e)
	}
	p = append(p, tdir+"/bad.json")
	if e = writeTestFile(p[len(p)-1], badData); e != nil {
		t.Skip("setup failure:", e)
	}

	d = LoadDataFiles("filename", tdir)
	if len(d) == len(p) {
		t.Error("bad.json passed")
	} else if len(d) == 0 {
		t.Error("no data loaded")
	} else {
		if b, e = yaml.Marshal(d[0]); e != nil {
			t.Error(e)
		} else if string(b) != goodYaml {
			t.Error("data returned out of order")
		}
		if b, e = json.Marshal(d[1]); e != nil {
			t.Error(e)
		} else if string(b) != goodJson {
			t.Error("data returned out of order")
		}
		if b, e = toml.Marshal(d[2]); e != nil {
			t.Error(e)
		} else if string(b) != goodToml {
			t.Error("data returned out of order")
		}
	}

	d = LoadDataFiles("filename-desc", tdir+"/*")
	if len(d) == len(p) {
		t.Error("bad.json passed")
	} else if len(d) == 0 {
		t.Error("no data loaded")
	} else {
		if b, e = toml.Marshal(d[0]); e != nil {
			t.Error(e)
		} else if string(b) != goodToml {
			t.Error("data returned out of order")
		}
		if b, e = json.Marshal(d[1]); e != nil {
			t.Error(e)
		} else if string(b) != goodJson {
			t.Error("data returned out of order")
		}
		if b, e = yaml.Marshal(d[2]); e != nil {
			t.Error(e)
		} else if string(b) != goodYaml {
			t.Error("data returned out of order")
		}
	}

	d = LoadDataFiles("modified", p...)
	if len(d) == len(p) {
		t.Error("bad.json passed")
	} else if len(d) == 0 {
		t.Error("no data loaded")
	} else {
		if b, e = json.Marshal(d[0]); e != nil {
			t.Error(e)
		} else if string(b) != goodJson {
			t.Error("data returned out of order")
		}
		if b, e = yaml.Marshal(d[1]); e != nil {
			t.Error(e)
		} else if string(b) != goodYaml {
			t.Error("data returned out of order")
		}
		if b, e = toml.Marshal(d[2]); e != nil {
			t.Error(e)
		} else if string(b) != goodToml {
			t.Error("data returned out of order")
		}
	}

	d = LoadDataFiles("modified-desc", p...)
	if len(d) == len(p) {
		t.Error("bad.json passed")
	} else if len(d) == 0 {
		t.Error("no data loaded")
	} else {
		if b, e = toml.Marshal(d[0]); e != nil {
			t.Error(e)
		} else if string(b) != goodToml {
			t.Error("data returned out of order")
		}
		if b, e = yaml.Marshal(d[1]); e != nil {
			t.Error(e)
		} else if string(b) != goodYaml {
			t.Error("data returned out of order")
		}
		if b, e = json.Marshal(d[2]); e != nil {
			t.Error(e)
		} else if string(b) != goodJson {
			t.Error("data returned out of order")
		}
	}
}

func TestGenerateSuperData(t *testing.T) {
	var data Data
	var e error
	var gd []Data
	var d []Data
	var sd Data

	if data, e = LoadData("json", strings.NewReader(goodJson)); e == nil {
		gd = append(gd, data)
	} else {
		t.Skip("setup failure:", e)
	}
	if data, e = LoadData("json", strings.NewReader(goodJson)); e == nil {
		gd = append(gd, data)
	} else {
		t.Skip("setup failure:", e)
	}
	if data, e = LoadData("yaml", strings.NewReader(goodYaml)); e == nil {
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

	if m, e = LoadData("json", strings.NewReader(goodJson)); e == nil {
		d = append(d, m)
	} else {
		t.Skip("setup failure:", e)
	}
	if m, e = LoadData("json", strings.NewReader(goodJson2)); e == nil {
		d = append(d, m)
	} else {
		t.Skip("setup failure:", e)
	}
	if m, e = LoadData("yaml", strings.NewReader(goodYaml)); e == nil {
		d = append(d, m)
	} else {
		t.Skip("setup failure:", e)
	}

	m = nil
	m = MergeData(d...)
	if m["json"] == nil || m["yaml"] == nil {
		t.Error("missing global keys")
	}
}
