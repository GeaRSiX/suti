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

func validateTemplateFile(t *testing.T, template Template, root string, partials ...string) {
	types := map[string]string{
		"tmpl":   "*template.Template",
		"gotmpl": "*template.Template",
		"hmpl":   "*template.Template",
		"gohmpl": "*template.Template",
	}
	if reflect.TypeOf(template).String() != types[getTemplateType(root)] {
		t.Error("invalid template loaded")
	}

	var rv []reflect.Value
	for _, p := range partials {
		p = filepath.Base(p)
		rv := reflect.ValueOf(template).MethodByName("Lookup").Call([]reflect.Value{
			reflect.ValueOf(p),
		})
		if rv[0].IsNil() {
			t.Errorf("missing defined template '%s'", p)
			rv = reflect.ValueOf(template).MethodByName("DefinedTemplates").Call([]reflect.Value{})
			t.Log(rv)
		}
	}
	rv = reflect.ValueOf(template).MethodByName("Name").Call([]reflect.Value{})
	if rv[0].String() != filepath.Base(root) {
		t.Errorf("invalid template name: %s does not match %s",
			rv[0].String(), filepath.Base(root))
	}
}

func TestLoadTemplateFile(t *testing.T) {
	var gr, gp, br, bp []string
	tdir := t.TempDir()

	gr = append(gr, tdir+"/goodRoot.tmpl")
	writeTestFile(t, gr[0], tmplRootGood)
	gp = append(gp, tdir+"/goodPartial.gotmpl")
	writeTestFile(t, gp[0], tmplPartialGood)
	br = append(br, tdir+"/badRoot.tmpl")
	writeTestFile(t, br[0], tmplRootBad)
	bp = append(bp, tdir+"/badPartial.gotmpl")
	writeTestFile(t, bp[0], tmplPartialBad)

	gr = append(gr, tdir+"/goodRoot.hmpl")
	writeTestFile(t, gr[1], hmplRootGood)
	gp = append(gp, tdir+"/goodPartial.gohmpl")
	writeTestFile(t, gp[1], hmplPartialGood)
	br = append(br, tdir+"/badRoot.hmpl")
	writeTestFile(t, br[1], hmplRootBad)
	bp = append(bp, tdir+"/badPartial.gohmpl")
	writeTestFile(t, bp[1], hmplPartialBad)

	for g, root := range gr { // good root, good partials
		if template, e := LoadTemplateFile(root, gp[g]); e != nil {
			t.Error(e)
		} else {
			validateTemplateFile(t, template, root, gp[g])
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

func validateExecuteTemplate(t *testing.T, results string, expect string, e error) {
	if e != nil {
		t.Error(e)
	}
	if results != expect {
		t.Errorf("invalid results: %s should match %s", results, expect)
	}
}

func TestExecuteTemplate(t *testing.T) {
	var e error
	var sd, data Data
	var gd, d []Data
	var tmplr, tmplp, hmplr, hmplp string
	var tmpl, hmpl Template
	var results bytes.Buffer
	tdir := t.TempDir()

	tmplr = tdir + "/tmplRootGood.gotmpl"
	writeTestFile(t, tmplr, tmplRootGood)
	tmplp = tdir + "/tmplPartialGood.tmpl"
	writeTestFile(t, tmplp, tmplPartialGood)
	hmplr = tdir + "/hmplRootGood.gohmpl"
	writeTestFile(t, hmplr, hmplRootGood)
	hmplp = tdir + "/hmplPartialGood.hmpl"
	writeTestFile(t, hmplp, hmplPartialGood)

	if data, e = LoadData("json", strings.NewReader(good["json"])); e != nil {
		t.Skip("setup failure:", e)
	}
	gd = append(gd, data)
	if data, e = LoadData("yaml", strings.NewReader(good["yaml"])); e != nil {
		t.Skip("setup failure:", e)
	}
	d = append(d, data)
	if data, e = LoadData("toml", strings.NewReader(good["toml"])); e != nil {
		t.Skip("setup failure:", e)
	}
	d = append(d, data)

	sd = GenerateSuperData("", d, gd...)
	if tmpl, e = LoadTemplateFile(tmplr, tmplp); e != nil {
		t.Skip("setup failure:", e)
	}
	if hmpl, e = LoadTemplateFile(hmplr, hmplp); e != nil {
		t.Skip("setup failure:", e)
	}

	results, e = ExecuteTemplate(tmpl, sd)
	validateExecuteTemplate(t, results.String(), tmplResult, e)

	results, e = ExecuteTemplate(hmpl, sd)
	validateExecuteTemplate(t, results.String(), hmplResult, e)
}
