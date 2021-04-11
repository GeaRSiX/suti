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
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const tmplRootGood = "{{.eg}} {{ template \"tmplPartialGood.tmpl\" . }}"
const tmplPartialGood = "{{range .data}}{{.eg}}{{end}}"
const tmplResult = "0 00"
const tmplRootBad = "{{ example }}} {{{ template \"tmplPartialBad.tmpl\" . }}"
const tmplPartialBad = "{{{ .example }}"

const hmplRootGood = "<!DOCTYPE html><html><p>{{.eg}} {{ template \"hmplPartialGood.hmpl\" . }}</p></html>"
const hmplPartialGood = "<b>{{range .data}}{{.eg}}{{end}}</b>"
const hmplResult = "<!DOCTYPE html><html><p>0 <b>00</b></p></html>"
const hmplRootBad = "{{ example }} {{{ template \"hmplPartialBad.hmpl\" . }}"
const hmplPartialBad = "<b>{{{ .example2 }}</b>"

const mstRootGood = "{{eg}} {{> mstPartialGood}}"
const mstPartialGood = "{{#data}}{{eg}}{{/data}}"
const mstResult = "0 00"
const mstRootBad = "{{> badPartial.mst}}{{#doesnt-exist}}{{/exit}}"
const mstPartialBad = "p{{$}{{ > noexist}}"

func validateTemplateFile(t *testing.T, template Template, root string, partials ...string) {
	types := map[string]string{
		"tmpl":     "*template.Template",
		"gotmpl":   "*template.Template",
		"hmpl":     "*template.Template",
		"gohmpl":   "*template.Template",
		"mst":      "*mustache.Template",
		"mustache": "*mustache.Template",
	}

	ttype := getTemplateType(root)
	if reflect.TypeOf(template.Template).String() != types[ttype] {
		t.Error("invalid template loaded")
	}

	if types[ttype] == "*template.Template" {
		var rv []reflect.Value
		for _, p := range partials {
			p = filepath.Base(p)
			rv := reflect.ValueOf(template.Template).MethodByName("Lookup").Call([]reflect.Value{
				reflect.ValueOf(p),
			})
			if rv[0].IsNil() {
				t.Errorf("missing defined template '%s'", p)
				rv = reflect.ValueOf(template.Template).MethodByName("DefinedTemplates").Call([]reflect.Value{})
				t.Log(rv)
			}
		}
		rv = reflect.ValueOf(template.Template).MethodByName("Name").Call([]reflect.Value{})
		if rv[0].String() != filepath.Base(root) {
			t.Errorf("invalid template name: %s does not match %s",
				rv[0].String(), filepath.Base(root))
		}
	}
}

func TestLoadTemplateFile(t *testing.T) {
	var gr, gp, br, bp []string
	tdir := t.TempDir()
	i := 0

	gr = append(gr, tdir+"/goodRoot.tmpl")
	writeTestFile(t, gr[i], tmplRootGood)
	gp = append(gp, tdir+"/goodPartial.gotmpl")
	writeTestFile(t, gp[i], tmplPartialGood)
	br = append(br, tdir+"/badRoot.tmpl")
	writeTestFile(t, br[i], tmplRootBad)
	bp = append(bp, tdir+"/badPartial.gotmpl")
	writeTestFile(t, bp[i], tmplPartialBad)
	i++

	gr = append(gr, tdir+"/goodRoot.hmpl")
	writeTestFile(t, gr[i], hmplRootGood)
	gp = append(gp, tdir+"/goodPartial.gohmpl")
	writeTestFile(t, gp[i], hmplPartialGood)
	br = append(br, tdir+"/badRoot.hmpl")
	writeTestFile(t, br[i], hmplRootBad)
	bp = append(bp, tdir+"/badPartial.gohmpl")
	writeTestFile(t, bp[i], hmplPartialBad)
	i++

	gr = append(gr, tdir+"/goodRoot.mustache")
	writeTestFile(t, gr[i], mstRootGood)
	gp = append(gp, tdir+"/goodPartial.mst")
	writeTestFile(t, gp[i], mstPartialGood)
	br = append(br, tdir+"/badRoot.mst")
	writeTestFile(t, br[i], mstRootBad)
	bp = append(bp, tdir+"/badPartial.mst")
	writeTestFile(t, bp[i], mstPartialBad)

	for g, root := range gr { // good root, good partials
		if template, e := LoadTemplateFile(root, gp[g]); e != nil {
			t.Error(e)
		} else {
			validateTemplateFile(t, template, root, gp[g])
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

func validateExecute(t *testing.T, results string, expect string, e error) {
	if e != nil {
		t.Error(e)
	} else if results != expect {
		t.Errorf("invalid results: '%s' should match '%s'", results, expect)
	}
}

func TestExecute(t *testing.T) {
	var e error
	var sd, gd, data Data
	var d []Data
	var tmplr, tmplp, hmplr, hmplp, mstr, mstp string
	var tmpl1, tmpl2, hmpl1, hmpl2, mst1, mst2 Template
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
	mstr = tdir + "/mstRootGood.mustache"
	writeTestFile(t, mstr, mstRootGood)
	mstp = tdir + "/mstPartialGood.mst"
	writeTestFile(t, mstp, mstPartialGood)

	if data, e = LoadData("json", strings.NewReader(good["json"])); e != nil {
		t.Skip("setup failure:", e)
	}
	gd = data
	if data, e = LoadData("yaml", strings.NewReader(good["yaml"])); e != nil {
		t.Skip("setup failure:", e)
	}
	d = append(d, data)
	if data, e = LoadData("toml", strings.NewReader(good["toml"])); e != nil {
		t.Skip("setup failure:", e)
	}
	d = append(d, data)

	if sd, e = GenerateSuperData("", gd, d); e != nil {
		t.Skip("setup failure:", e)
	}
	if tmpl1, e = LoadTemplateFile(tmplr, tmplp); e != nil {
		t.Skip("setup failure:", e)
	}
	if tmpl2, e = LoadTemplateFile(tmplr, tdir); e != nil {
		t.Skip("setup failure:", e)
	}
	if hmpl1, e = LoadTemplateFile(hmplr, hmplp); e != nil {
		t.Skip("setup failure:", e)
	}
	if hmpl2, e = LoadTemplateFile(tmplr, tdir); e != nil {
		t.Skip("setup failure:", e)
	}
	if mst1, e = LoadTemplateFile(mstr, mstp); e != nil {
		t.Skip("setup failure:", e)
	}
	if mst2, e = LoadTemplateFile(tmplr, tdir); e != nil {
		t.Skip("setup failure:", e)
	}

	results, e = tmpl1.Execute(sd)
	validateExecute(t, results.String(), tmplResult, e)
	results, e = tmpl2.Execute(sd)
	validateExecute(t, results.String(), tmplResult, e)

	results, e = hmpl1.Execute(sd)
	validateExecute(t, results.String(), hmplResult, e)
	results, e = hmpl2.Execute(sd)
	validateExecute(t, results.String(), tmplResult, e)

	results, e = mst1.Execute(sd)
	validateExecute(t, results.String(), mstResult, e)
	results, e = mst2.Execute(sd)
	validateExecute(t, results.String(), mstResult, e)
}
