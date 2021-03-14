package main

import (
	"bytes"
	"fmt"
	hmpl "html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	tmpl "text/template"
)

type Template interface{}

func getTemplateType(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func LoadTemplateFile(root string, partials ...string) (t Template, e error) {
	if len(root) == 0 {
		e = fmt.Errorf("no root template specified")
	}
	if stat, err := os.Stat(root); err != nil {
		e = err
	} else if stat.IsDir() {
		e = fmt.Errorf("root path must be a file, not a directory: %s", root)
	}
	if e != nil {
		return
	}

	ttype := getTemplateType(root)

	if ttype == "tmpl" || ttype == "gotmpl" {
		var gotmpl *tmpl.Template
		if gotmpl, e = tmpl.ParseFiles(root); e != nil {
			return nil, e
		}
		for _, p := range partials {
			ptype := getTemplateType(p)
			if ptype == "tmpl" || ptype == "gotmpl" {
				gotmpl, e = gotmpl.ParseFiles(p)
			}
		}
		t = gotmpl
	} else if ttype == "hmpl" || ttype == "gohmpl" {
		var gohmpl *hmpl.Template
		if gohmpl, e = hmpl.ParseFiles(root); e != nil {
			return nil, e
		}
		for _, p := range partials {
			ptype := getTemplateType(p)
			if ptype == "tmpl" || ptype == "gotmpl" {
				gohmpl, e = gohmpl.ParseFiles(p)
			}
		}
		t = gohmpl
	} else {
		e = fmt.Errorf("'%s' is not a supported template language", ttype)
	}

	return
}

func ExecuteTemplate(t Template, d Data) (result bytes.Buffer, err error) {
	if t == nil || d == nil {
		err = fmt.Errorf("missing parameters")
		return
	}

	tv := reflect.ValueOf(t)
	tt := reflect.TypeOf(t)
	if tt.String() == "*template.Template" { // tmpl or hmpl
		rval := tv.MethodByName("Execute").Call([]reflect.Value{
			reflect.ValueOf(&result), reflect.ValueOf(&d),
		})
		if rval[0].IsNil() == false {
			err = rval[0].Interface().(error)
		}
	} else {
		err = fmt.Errorf("unable to infer template type '%s'", tt.String())
	}

	return
}
