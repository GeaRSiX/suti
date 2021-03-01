package main

import (
	"strings"
	"testing"
	"os"
)

const goodJson = `{"example": 1}`
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
	
	if d, e = LoadData("json", strings.NewReader(goodJson)); e != nil {
		t.Error(e)
	}
	if len(d) == 0 {
		t.Fail()
	} else {
		t.Log(d)
	}
	
	if d, e = LoadData("json", strings.NewReader(badJson)); e == nil {
		t.Error("bad.json passed")
	}
	
	return
}

func TestLoadDataFiles(t *testing.T) {
	var e error
	var p []string
	var d map[string]Data
	tdir := t.TempDir()
	
	p = append(p, tdir+"/good.json")
	if e = writeTestFile(p[0], goodJson); e != nil {
		t.Skip("setup failure:", e)
	}
	p = append(p, tdir+"/good.json")
	if e = writeTestFile(p[1], goodJson); e != nil {
		t.Skip("setup failure:", e)
	}
	p = append(p, tdir+"/bad.json")
	if e = writeTestFile(p[2], badJson); e != nil {
		t.Skip("setup failure:", e)
	}
	
	d = LoadDataFiles(tdir)
	if len(d) == len(p) {
		t.Error("bad.json passed")
	} else if len(d) == 0 {
		t.Error("no data loaded")
	}
	d = LoadDataFiles(p...)
	if len(d) == len(p) {
		t.Error("bad.json passed")
	} else if len(d) == 0 {
		t.Error("no data loaded")
	}
	d = LoadDataFiles(tdir+"/*")
	if len(d) == len(p) {
		t.Error("bad.json passed")
	} else if len(d) == 0 {
		t.Error("no data loaded")
	}
}