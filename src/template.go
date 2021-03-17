package main

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
    "github.com/cbroglie/mustache"
	"fmt"
	hmpl "html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	tmpl "text/template"
)

// Template is a generic interface container for any template type
type Template interface{}

func getTemplateType(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func loadTemplateFileTmpl(root string, partials ...string) (t *tmpl.Template, e error) {
	var stat os.FileInfo
	if t, e = tmpl.ParseFiles(root); e == nil {
		for _, p := range partials {
			ptype := getTemplateType(p)
			stat, e = os.Stat(p)
			
			if e == nil {
				if ptype == "tmpl" || ptype == "gotmpl" {
					t, e = t.ParseFiles(p)
				} else if strings.Contains(p, "*") {
					t, e = t.ParseGlob(p)
				} else if stat.IsDir() {
					t, e = t.ParseGlob(p+"/*.tmpl")
					t, e = t.ParseGlob(p+"/*.gotmpl")
				} else {
					e = fmt.Errorf("non-matching filetype")
				}
			}
			
			if e != nil {
				warn("skipping partial '%s': %s", p, e)
			}
		}
	}
	return
}

func loadTemplateFileHmpl(root string, partials ...string) (t *hmpl.Template, e error) {
	var stat os.FileInfo
	if t, e = hmpl.ParseFiles(root); e == nil {
		for _, p := range partials {
			ptype := getTemplateType(p)
			stat, e = os.Stat(p)
			
			if e == nil {
				if ptype == "hmpl" || ptype == "gohmpl" {
					t, e = t.ParseFiles(p)
				} else if strings.Contains(p, "*") {
					t, e = t.ParseGlob(p)
				} else if stat.IsDir() {
					t, e = t.ParseGlob(p+"/*.hmpl")
					t, e = t.ParseGlob(p+"/*.gohmpl")
				} else {
					e = fmt.Errorf("non-matching filetype")
				}
			}
			
			if e != nil {
				warn("skipping partial '%s': %s", p, e)
				e = nil
			}
		}
	}
	return
}

func loadTemplateFileMst(root string, partials ...string) (t *mustache.Template, e error) {
	for p, partial := range partials {
		if stat, err := os.Stat(partial); err != nil {
			partials = append(partials[:p], partials[p+1:]...)
			warn("skipping partial '%s': %s", partial, e)
		} else if stat.IsDir() == false {
			partials[p] = filepath.Dir(partial)
		} else if strings.Contains(partial, "*") {
			if paths, err := filepath.Glob(partial); err != nil {
				partials = append(partials[:p], partials[p+1:]...)
				warn("skipping partial '%s': %s", partial, e)
			} else {
				partials = append(partials[:p], partials[p+1:]...)
				partials = append(partials, paths...)
			}
		}
	}
	
	mstfp := &mustache.FileProvider{
		Paths:      partials,
		Extensions: []string{".mst", ".mustache"},
	}
	t, e = mustache.ParseFilePartials(root, mstfp)
	
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
    } else if ttype == "mst" || ttype == "mustache" {
        t, e = loadTemplateFileMst(root, partials...)
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

	var rval []reflect.Value
	if tt.String() == "*template.Template" { // tmpl or hmpl
		rval = tv.MethodByName("Execute").Call([]reflect.Value{
			reflect.ValueOf(&result), reflect.ValueOf(&d),
		})
	} else if tt.String() == "*mustache.Template" { // mustache
		rval = tv.MethodByName("FRender").Call([]reflect.Value{
			reflect.ValueOf(&result), reflect.ValueOf(&d),
		})
	} else {
		err = fmt.Errorf("unable to infer template type '%s'", tt.String())
	}
	
	if rval[0].IsNil() == false { // rval[0] = err
		err = rval[0].Interface().(error)
	}
	
	return
}
