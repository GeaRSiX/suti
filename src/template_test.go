package main

import (
	"testing"
)

const tmplRootGood = "hello {{ template \"partial\" . }}"
const tmplPartialGood = "{{ .language }}"
const tmplRootBad = "hello {{{ template \"partial\" . }}"
const tmplPartialBad = "{{{ .language }}"

const hmplRootGood = "hello {{ template \"partial\" . }}"
const hmplPartialGood = "<b>{{ .language }}</b>"
const hmplRootBad = "hello {{{ template \"partial\" . }}"
const hmplPartialBad = "<b>{{{ .language }}</b>"

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
			t.Errorf("good template with bad partials passed\n")
		}
	}
	for _, root := range br { // bad root, good partials
		if tmp, e := LoadTemplateFile(root, gp...); e == nil {
			t.Log(tmp)
			t.Errorf("bad template with good partials passed\n")
		}
	}
	for _, root := range br { // bad root, bad partials
		if _, e := LoadTemplateFile(root, bp...); e == nil {
			t.Errorf("bad template with bad partials passed\n")
		}
	}
}
