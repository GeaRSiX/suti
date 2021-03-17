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

// Template is a generic interace container for any template type
type Template interface{}

func getTemplateType(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func loadTemplateFileTmpl(root string, partials ...string) (t Template, e error) {
	var gotmpl *tmpl.Template
	if gotmpl, e = tmpl.ParseFiles(root); e == nil {
		for _, p := range partials {
			ptype := getTemplateType(p)
			if ptype == "tmpl" || ptype == "gotmpl" {
				if gotmpl, e = gotmpl.ParseFiles(p); e != nil {
					warn("failed to parse partial '%s': %s", p, e)
				}
			} else {
				warn("skipping partial '%s': non-matching filetype", p)
			}
		}
		t = gotmpl
	}
	return
}

func loadTemplateFileHmpl(root string, partials ...string) (t Template, e error) {
	var gohmpl *hmpl.Template
	if gohmpl, e = hmpl.ParseFiles(root); e == nil {
		for _, p := range partials {
			ptype := getTemplateType(p)
			if ptype == "hmpl" || ptype == "gohmpl" {
				if gohmpl, e = gohmpl.ParseFiles(p); e != nil {
					warn("failed to parse partial '%s': %s", p, e)
				}
			} else {
				warn("skipping partial '%s': non-matching filetype", p)
			}
		}
		t = gohmpl
	}
	return
}

// LoadTemplateFile loads a Template from file `root`. All files in `partials`
// that have the same template type (identified by file extension) are also
// parsed and associated with the parsed root template.
func LoadTemplateFile(root string, partials ...string) (t Template, e error) {
	if len(root) == 0 {
		return nil, fmt.Errorf("no root tempslate specified")
	}

	if stat, err := os.Stat(root); err != nil {
		return nil, err
	} else if stat.IsDir() {
		return nil, fmt.Errorf("root path must be a file, not a directory: %s", root)
	}

	ttype := getTemplateType(root)
	if ttype == "tmpl" || ttype == "gotmpl" {
		t, e = loadTemplateFileTmpl(root, partials...)
	} else if ttype == "hmpl" || ttype == "gohmpl" {
		t, e = loadTemplateFileHmpl(root, partials...)
	} else {
		e = fmt.Errorf("'%s' is not a supported template language", ttype)
	}
	return
}

// ExecuteTemplate executes `t` against `d`. Reflection is used to determine
// the template type and call it's execution fuction.
func ExecuteTemplate(t Template, d Data) (result bytes.Buffer, err error) {
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
