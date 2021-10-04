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

const tmplRootGood = `{{.eg}} {{ template "tmplPartialGood" . }}`
const tmplPartialGood = `{{.eg}}`
const tmplResult = `0 0`
const tmplRootBad = `{{ example }}} {{{ template \"tmplPartialBad\" . }}`
const tmplPartialBad = `{{{ .example }}`

const hmplRootGood = `<!DOCTYPE html><html><p>{{.eg}} {{ template "hmplPartialGood" . }}</p></html>`
const hmplPartialGood = `<b>{{.eg}}</b>`
const hmplResult = `<!DOCTYPE html><html><p>0 <b>0</b></p></html>`
const hmplRootBad = `{{ example }} {{{ template "hmplPartialBad" . }}`
const hmplPartialBad = `<b>{{{ .example2 }}</b>`

const mstRootGood = `{{eg}} {{> mstPartialGood}}`
const mstPartialGood = `{{eg}}`
const mstResult = `0 0`
const mstRootBad = `{{> badPartial.mst}}{{#doesnt-exist}}{{/exit}}`
const mstPartialBad = `p{{$}}{{ > noexist}`

func TestIsSupportedTemplateLang(t *testing.T) {
	exts := []string{
		".tmpl", "tmpl", "TMPL", ".TMPL",
		".hmpl", "hmpl", "HMPL", ".HMPL",
		".mst", "mst", "MST", ".MST",
		".NONE", "-", ".", "",
	}

	for i, ext := range exts {
		var target int
		if i < 4 {
			target = 0
		} else if i < 8 {
			target = 1
		} else if i < 12 {
			target = 2
		} else {
			target = -1
		}

		if IsSupportedTemplateLang(ext) != target {
			if target == -1 {
				t.Fatalf("%s is not a supported data language", ext)
			} else {
				t.Fatalf("%s did not return %s", ext, SupportedTemplateLangs[target])
			}
		}
	}
}

func validateTemplate(t *testing.T, template Template, templateType string, rootName string, partialNames ...string) {
	types := map[string]string{
		"tmpl":     "*template.Template",
		"gotmpl":   "*template.Template",
		"hmpl":     "*template.Template",
		"gohmpl":   "*template.Template",
		"mst":      "*mustache.Template",
		"mustache": "*mustache.Template",
	}

	rt := reflect.TypeOf(template.Template).String()
	if rt != types[templateType] {
		t.Fatalf("invalid template type '%s' loaded, should be '%s'", rt, types[templateType])
	}

	if types[templateType] == "*template.Template" {
		var rv []reflect.Value
		for _, p := range partialNames {
			rv := reflect.ValueOf(template.Template).MethodByName("Lookup").Call([]reflect.Value{reflect.ValueOf(p)})
			if rv[0].IsNil() {
				t.Fatalf("missing defined template '%s'", p)
				rv = reflect.ValueOf(template.Template).MethodByName("DefinedTemplates").Call([]reflect.Value{})
				t.Log(rv)
			}
		}
		rv = reflect.ValueOf(template.Template).MethodByName("Name").Call([]reflect.Value{})
		if rv[0].String() != rootName {
			t.Fatalf("invalid template name: %s does not match %s", rv[0].String(), rootName)
		}
	}
}

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
		t.Fatal("invalid template loaded")
	}

	if types[ttype] == "*template.Template" {
		var rv []reflect.Value
		for _, p := range partials {
			p = filepath.Base(p)
			rv := reflect.ValueOf(template.Template).MethodByName("Lookup").Call([]reflect.Value{
				reflect.ValueOf(p),
			})
			if rv[0].IsNil() {
				t.Fatalf("missing defined template '%s'", p)
				rv = reflect.ValueOf(template.Template).MethodByName("DefinedTemplates").Call([]reflect.Value{})
				t.Log(rv)
			}
		}
		rv = reflect.ValueOf(template.Template).MethodByName("Name").Call([]reflect.Value{})
		if rv[0].String() != filepath.Base(root) {
			t.Fatalf("invalid template name: %s does not match %s",
				rv[0].String(), filepath.Base(root))
		}
	}
}

func TestLoadTemplateFilepath(t *testing.T) {
	var gr, gp, br, bp []string
	tdir := t.TempDir()
	i := 0

	gr = append(gr, tdir+"/goodRoot.tmpl")
	writeTestFile(t, gr[i], tmplRootGood)
	gp = append(gp, tdir+"/goodPartial.tmpl")
	writeTestFile(t, gp[i], tmplPartialGood)
	br = append(br, tdir+"/badRoot.tmpl")
	writeTestFile(t, br[i], tmplRootBad)
	bp = append(bp, tdir+"/badPartial.tmpl")
	writeTestFile(t, bp[i], tmplPartialBad)
	i++

	gr = append(gr, tdir+"/goodRoot.hmpl")
	writeTestFile(t, gr[i], hmplRootGood)
	gp = append(gp, tdir+"/goodPartial.hmpl")
	writeTestFile(t, gp[i], hmplPartialGood)
	br = append(br, tdir+"/badRoot.hmpl")
	writeTestFile(t, br[i], hmplRootBad)
	bp = append(bp, tdir+"/badPartial.hmpl")
	writeTestFile(t, bp[i], hmplPartialBad)
	i++

	gr = append(gr, tdir+"/goodRoot.mst")
	writeTestFile(t, gr[i], mstRootGood)
	gp = append(gp, tdir+"/goodPartial.mst")
	writeTestFile(t, gp[i], mstPartialGood)
	br = append(br, tdir+"/badRoot.mst")
	writeTestFile(t, br[i], mstRootBad)
	bp = append(bp, tdir+"/badPartial.mst")
	writeTestFile(t, bp[i], mstPartialBad)

	for g, root := range gr { // good root, good partials
		if template, e := LoadTemplateFilepath(root, gp[g]); e != nil {
			t.Fatal(e)
		} else {
			validateTemplateFile(t, template, root, gp[g])
		}
	}
	for _, root := range br { // bad root, good partials
		if _, e := LoadTemplateFilepath(root, gp...); e == nil {
			t.Fatalf("no error for bad template with good partials\n")
		}
	}
	for _, root := range br { // bad root, bad partials
		if _, e := LoadTemplateFilepath(root, bp...); e == nil {
			t.Fatalf("no error for bad template with bad partials\n")
		}
	}
}

func TestLoadTemplateString(t *testing.T) {
	var err error
	var template Template
	var templateType string

	testInvalid := func(templateType string, template Template) {
		t.Logf("invalid '%s' template managed to load", templateType)
		if buf, err := template.Execute(""); err == nil {
			t.Fatalf("invalid '%s' template managed to execute: %s", templateType, buf.String())
		}
	}

	name := "test"
	templateType = "tmpl"
	if template, err = LoadTemplateString(templateType, name, tmplRootGood,
		map[string]string{"tmplPartialGood.tmpl": tmplPartialGood}); err != nil {
			t.Fatalf("'%s' template failed to load", templateType)
		}
	if template, err = LoadTemplateString(templateType, name, tmplRootBad,
		map[string]string{"tmplPartialGood.tmpl": tmplPartialGood}); err == nil {
			testInvalid(templateType, template)
		}
	if template, err = LoadTemplateString(templateType, name, tmplRootGood,
		map[string]string{"tmplPartialGood.tmpl": tmplPartialBad}); err == nil {
			testInvalid(templateType, template)
		}

	templateType = "hmpl"
	if template, err = LoadTemplateString(templateType, name, hmplRootGood,
		map[string]string{"hmplPartialGood.hmpl": hmplPartialGood}); err != nil {
			t.Fatalf("'%s' template failed to load", templateType)
		}
	if template, err = LoadTemplateString(templateType, name, hmplRootBad,
		map[string]string{"hmplPartialGood.hmpl": hmplPartialGood}); err == nil {
			testInvalid(templateType, template)
		}
	if template, err = LoadTemplateString(templateType, name, hmplRootGood,
		map[string]string{"hmplPartialGood.hmpl": hmplPartialBad}); err == nil {
			testInvalid(templateType, template)
		}

	templateType = "mst"
	if template, err = LoadTemplateString(templateType, name, mstRootGood,
		map[string]string{"mstPartialGood": mstPartialGood}); err != nil {
			t.Fatalf("'%s' template failed to load", templateType)
		}
	if template, err = LoadTemplateString(templateType, name, mstRootBad,
		map[string]string{"mstPartialGood": mstPartialGood}); err == nil {
			testInvalid(templateType, template)
		}
	if template, err = LoadTemplateString(templateType, name, mstRootGood,
		map[string]string{"mstPartialGood": mstPartialBad}); err == nil {
			testInvalid(templateType, template)
		}
}

// func TestLoadTemplateString(t *testing.T) {} // This is tested by TestLoadTemplateFilepath and TestLoadTemplateString

func validateExecute(t *testing.T, results string, expect string, e error) {
	if e != nil {
		t.Fatal(e)
	} else if results != expect {
		t.Fatalf("invalid results: '%s' should match '%s'", results, expect)
	}
}

func TestExecute(t *testing.T) {
	t.Parallel()
	
	var err error
	var tmpl Template
	var data map[string]interface{}
	var results bytes.Buffer
	
	if tmpl, err = LoadTemplateString("tmpl", "tmplRootGood", tmplRootGood,
		map[string]string{"tmplPartialGood": tmplPartialGood}); err != nil {
			t.Skip("setup failure:", err)
		}
	if err = LoadData("json", strings.NewReader(good["json"]), &data); err != nil {
		t.Skip("setup failure:", err)
	}
	results, err = tmpl.Execute(data)
	validateExecute(t, results.String(), tmplResult, err)
	
	if tmpl, err = LoadTemplateString("hmpl", "hmplRootGood", hmplRootGood,
		map[string]string{"hmplPartialGood": hmplPartialGood}); err != nil {
			t.Skip("setup failure:", err)
		}
	if err = LoadData("yaml", strings.NewReader(good["yaml"]), &data); err != nil {
		t.Skip("setup failure:", err)
	}
	results, err = tmpl.Execute(data)
	validateExecute(t, results.String(), hmplResult, err)
	
	if tmpl, err = LoadTemplateString("mst", "mstRootGood", mstRootGood,
		map[string]string{"mstPartialGood": mstPartialGood}); err != nil {
			t.Skip("setup failure:", err)
		}
	if err = LoadData("toml", strings.NewReader(good["toml"]), &data); err != nil {
		t.Skip("setup failure:", err)
	}
	results, err = tmpl.Execute(data)
	validateExecute(t, results.String(), mstResult, err)
}
