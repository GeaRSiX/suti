package main

import (
	"strings"
	"testing"
	"os"
)

const goodJson = `{"example": 1}`
const badJson = `{"example":2:]}}`

func writeTestFile(path string, data string) (e error) {
	var f *os.File
	
	if f, e = os.Create(path); e != nil {
		return
	}
	if _, e = f.WriteString(data); e != nil {
		return
	}
	f.Close()
	
	return
}

func TestLoadDataFiles(t *testing.T) {
	var e error
	var p []string
	var d map[string]data
	tdir := t.TempDir()
	
	p = append(p, tdir+"/good.json")
	if e = writeTestFile(p[0], goodJson); e != nil {
		t.Skip("setup failure:", e)
	}
	p = append(p, tdir+"/bad.json")
	if e = writeTestFile(p[1], badJson); e != nil {
		t.Skip("setup failure:", e)
	}
	
	d = LoadDataFiles(tdir)
	if len(d) == len(p) {
		t.Error("bad.json passed")
	}
	d = LoadDataFiles(p...)
	if len(d) == len(p) {
		t.Error("bad.json passed")
	}
}

func TestLoadData(t *testing.T) {
	var e error
	
	if _, e = LoadData("json", strings.NewReader(goodJson)); e != nil {
		t.Error(e)
	}
	if _, e = LoadData("json", strings.NewReader(badJson)); e == nil {
		t.Error("bad.json passed")
	}
	
	return
}
