package main

import (
	"bytes"
	"strings"
	"testing"
)

const tmplRootGood = "{{ .example1 }} {{ template \"tmplPartialGood\" . }}"
const tmplPartialGood = "{{range .data}} .example2 {{ end }}"
const tmplRootBad = "{{ example1 }} {{{ template \"partial\" . }}"
const tmplPartialBad = "{{{ .example }}"

const hmplRootGood = "<!DOCTYPE html><html>{{ .example1 }} {{ template \"hmplPartialGood\" . }}</html>"
const hmplPartialGood = "{{range .data}}<b>.example2</b>{{ end }}"
const hmplRootBad = "{{ example1 }} {{{ template \"partial\" . }}"
const hmplPartialBad = "<b>{{{ .example2 }}</b>"

func TestLoadTemplateFile(t *testing.T) {
	var e error
	var gr, gp, br, bp []string
	tdir := t.TempDir()

	gr = append(gr, tdir+"/goodRoot.tmpl")
	if e = writeTestFile(gr[0], tmplRootGood); e != nil {
		t.Skip("setup failure:", e)
	}
	gp = append(gp, tdir+"/goodPartial.gotmpl")
	if e = writeTestFile(gp[0], tmplPartialGood); e != nil {
		t.Skip("setup failure:", e)
	}
	br = append(br, tdir+"/badRoot.tmpl")
	if e = writeTestFile(br[0], tmplRootBad); e != nil {
		t.Skip("setup failure:", e)
	}
	bp = append(bp, tdir+"/badPartial.gotmpl")
	if e = writeTestFile(bp[0], tmplPartialBad); e != nil {
		t.Skip("setup failure:", e)
	}

	gr = append(gr, tdir+"/goodRoot.hmpl")
	if e = writeTestFile(gr[1], hmplRootGood); e != nil {
		t.Skip("setup failure:", e)
	}
	gp = append(gp, tdir+"/goodPartial.gohmpl")
	if e = writeTestFile(gp[1], hmplPartialGood); e != nil {
		t.Skip("setup failure:", e)
	}
	br = append(br, tdir+"/badRoot.hmpl")
	if e = writeTestFile(br[1], hmplRootBad); e != nil {
		t.Skip("setup failure:", e)
	}
	bp = append(bp, tdir+"/badPartial.gohmpl")
	if e = writeTestFile(bp[1], hmplPartialBad); e != nil {
		t.Skip("setup failure:", e)
	}

	for _, root := range gr { // good root, good partials
		if _, e := LoadTemplateFile(root, gp...); e != nil {
			t.Error(e)
		}
	}
	for _, root := range gr { // good root, bad partials
		if _, e := LoadTemplateFile(root, bp...); e == nil {
			t.Errorf("no error for good template with bad partials\n")
		}
	}
	for _, root := range br { // bad root, good partials
		if _, e := LoadTemplateFile(root, gp...); e == nil {
			t.Errorf("no error for bad template with good partials\n")
		}
	}
	for _, root := range br { // bad root, bad partials
		if _, e := LoadTemplateFile(root, bp...); e == nil {
			t.Errorf("no error for bad template with bad partials\n")
		}
	}
}

func TestExecuteTemplate(t *testing.T) {
	var e error
	var sd, gd, d Data
	var tmpl, hmpl Template
	var results bytes.Buffer

	if gd, e = LoadData("json", strings.NewReader(goodJson1)); e != nil {
		t.Skip("setup failure:", e)
	}
	if d, e = LoadData("json", strings.NewReader(goodJson2)); e != nil {
		t.Skip("setup failure:", e)
	}
	data := make([]Data, 1)
	data = append(data, d)
	sd = GenerateSuperData("", data, gd)
	if tmpl, e = LoadTemplateFile(tmplRootGood, tmplPartialGood); e != nil {
		t.Skip("setup failure:", e)
	}
	if hmpl, e = LoadTemplateFile(hmplRootGood, hmplPartialGood); e != nil {
		t.Skip("setup failure:", e)
	}

	if results, e = ExecuteTemplate(tmpl, sd); e != nil {
		t.Error(e)
	}
	if results, e = ExecuteTemplate(hmpl, sd); e != nil {
		t.Error(e)
	}
	t.Log(results)
}
