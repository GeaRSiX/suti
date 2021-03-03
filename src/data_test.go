package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const goodJson1 = `{"example1":0}`
const goodJson2 = `{"example2":1}`
const badJson = `{"example":2:]}}`

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

	if d, e = LoadData("json", strings.NewReader(goodJson1)); e != nil {
		t.Error(e)
	} else if len(d) == 0 {
		t.Error("no data loaded")
	}

	if d, e = LoadData("json", strings.NewReader(badJson)); e == nil {
		t.Error("bad.json passed")
	}
	
	if d, e = LoadData("json", strings.NewReader("")); e == nil {
		t.Error("empty file passed")
	}

	return
}

func TestLoadDataFiles(t *testing.T) {
	var e error
	var p []string
	var b []byte
	var d []Data
	tdir := t.TempDir()

	p = append(p, tdir+"/good2.json")
	if e = writeTestFile(p[0], goodJson2); e != nil {
		t.Skip("setup failure:", e)
	}
	p = append(p, tdir+"/good1.json")
	if e = writeTestFile(p[1], goodJson1); e != nil {
		t.Skip("setup failure:", e)
	}
	p = append(p, tdir+"/good1.json")
	if e = writeTestFile(p[2], goodJson1); e != nil {
		t.Skip("setup failure:", e)
	}
	p = append(p, tdir+"/bad.json")
	if e = writeTestFile(p[3], badJson); e != nil {
		t.Skip("setup failure:", e)
	}

	d = LoadDataFiles("filename", tdir)
	if len(d) == len(p) {
		t.Error("bad.json passed")
	} else if len(d) == 0 {
		t.Error("no data loaded")
	} else {
		if b, e = json.Marshal(d[0]); e != nil {
			t.Error(e)
		} else if string(b) == goodJson2 {
			t.Error("data returned out of order")
		} else if string(b) != goodJson1 {
			t.Errorf("incorrect json: %s does not match %s", b, goodJson1)
		}
		
		if b, e = json.Marshal(d[1]); e != nil {
			t.Error(e)
		} else if string(b) == goodJson1 {
			t.Error("data returned out of order")
		} else if string(b) != goodJson2 {
			t.Errorf("incorrect json: %s does not match %s", b, goodJson2)
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
		} else if string(b) == goodJson1 {
			t.Error("data returned out of order")
		} else if string(b) != goodJson2 {
			t.Errorf("incorrect json: %s does not match %s", b, goodJson1)
		}
		
		if b, e = json.Marshal(d[1]); e != nil {
			t.Error(e)
		} else if string(b) == goodJson2 {
			t.Error("data returned out of order")
		} else if string(b) != goodJson1 {
			t.Errorf("incorrect json: %s does not match %s", b, goodJson2)
		}
	}
	
	d = LoadDataFiles("", tdir + "/*")
	if len(d) == len(p) {
		t.Error("bad.json passed")
	} else if len(d) == 0 {
		t.Error("no data loaded")
	}
}
