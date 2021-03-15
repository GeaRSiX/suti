package main

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const tmplRootGood = "root {{ template \"tmplPartialGood.tmpl\" . }}"
const tmplPartialGood = "partial"
const tmplResult = "root partial"
const tmplRootBad = "{{ example }}} {{{ template \"partial\" . }}"
const tmplPartialBad = "{{{ .example }}"

const hmplRootGood = "<!DOCTYPE html><html><p>root {{ template \"hmplPartialGood.hmpl\" . }}</p></html>"
const hmplPartialGood = "<b>partial</b>"
const hmplResult = "<!DOCTYPE html><html><p>root <b>partial</b></p></html>"
const hmplRootBad = "{{ example }} {{{ template \"partial\" . }}"
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
		if template, e := LoadTemplateFile(root, gp...); e != nil {
			t.Error(e)
		} else {
			ttype := getTemplateType(root)
			if ttype == "tmpl" || ttype == "gotmpl" {
				ttype = reflect.TypeOf(template).String()
				if ttype != "*template.Template" {
					t.Errorf("invalid tempate type parsed: %s should be *template.Template", ttype)
				}
				rv := reflect.ValueOf(template).MethodByName("Lookup").Call([]reflect.Value{
					reflect.ValueOf("goodRoot.tmpl"),
				})
				if rv[0].IsNil() {
					t.Error("missing defined templates")
				}
				rv = reflect.ValueOf(template).MethodByName("Lookup").Call([]reflect.Value{
					reflect.ValueOf("goodPartial.gotmpl"),
				})
				if rv[0].IsNil() {
					t.Error("missing defined templates")
				}
				rv = reflect.ValueOf(template).MethodByName("Name").Call([]reflect.Value{})
				if rv[0].String() != filepath.Base(root) {
					t.Errorf("invalid template name: %s does not match %s", rv[0].String(), filepath.Base(root))
				}
			} else if ttype == "hmpl" || ttype == "gohmpl" {
				ttype = reflect.TypeOf(template).String()
				if ttype != "*template.Template" {
					t.Errorf("invalid tempate type parsed: %s should be *template.Template", ttype)
				}
				rv := reflect.ValueOf(template).MethodByName("Lookup").Call([]reflect.Value{
					reflect.ValueOf("goodRoot.hmpl"),
				})
				if rv[0].IsNil() {
					t.Error("missing defined templates")
				}
				rv = reflect.ValueOf(template).MethodByName("Lookup").Call([]reflect.Value{
					reflect.ValueOf("goodPartial.gohmpl"),
				})
				if rv[0].IsNil() {
					t.Error("missing defined templates")
				}
				rv = reflect.ValueOf(template).MethodByName("Name").Call([]reflect.Value{})
				if rv[0].String() != filepath.Base(root) {
					t.Errorf("invalid template name: %s does not match %s", rv[0].String(), filepath.Base(root))
				}
			} else {
				t.Errorf("test broken: invalid template type written (%s)", root)
			}
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
	var tmplr, tmplp, hmplr, hmplp string
	var tmpl, hmpl Template
	var results bytes.Buffer
	tdir := t.TempDir()

	tmplr = tdir + "/tmplRootGood.gotmpl"
	if e = writeTestFile(tmplr, tmplRootGood); e != nil {
		t.Skip("setup failure:", e)
	}
	tmplp = tdir + "/tmplPartialGood.tmpl"
	if e = writeTestFile(tmplp, tmplPartialGood); e != nil {
		t.Skip("setup failure:", e)
	}
	hmplr = tdir + "/hmplRootGood.gohmpl"
	if e = writeTestFile(hmplr, hmplRootGood); e != nil {
		t.Skip("setup failure:", e)
	}
	hmplp = tdir + "/hmplPartialGood.hmpl"
	if e = writeTestFile(hmplp, hmplPartialGood); e != nil {
		t.Skip("setup failure:", e)
	}

	if gd, e = LoadData("json", strings.NewReader(goodJson)); e != nil {
		t.Skip("setup failure:", e)
	}
	if d, e = LoadData("yaml", strings.NewReader(goodYaml)); e != nil {
		t.Skip("setup failure:", e)
	}
	if d, e = LoadData("toml", strings.NewReader(goodToml)); e != nil {
		t.Skip("setup failure:", e)
	}

	data := make([]Data, 1)
	data = append(data, d)
	sd = GenerateSuperData("", data, gd)
	if tmpl, e = LoadTemplateFile(tmplr, tmplp); e != nil {
		t.Skip("setup failure:", e)
	}
	if hmpl, e = LoadTemplateFile(hmplr, hmplp); e != nil {
		t.Skip("setup failure:", e)
	}

	if results, e = ExecuteTemplate(tmpl, sd); e != nil {
		t.Error(e)
	} else if results.String() != tmplResult {
		t.Errorf("invalid results: %s should match %s", results.String(), tmplResult)
	}
	if results, e = ExecuteTemplate(hmpl, sd); e != nil {
		t.Error(e)
	} else if results.String() != hmplResult {
		t.Errorf("invalid results: %s should match %s", results.String(), hmplResult)
	}
}
