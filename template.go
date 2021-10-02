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
	"fmt"
	mst "github.com/cbroglie/mustache"
	hmpl "html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	tmpl "text/template"
)

// SupportedTemplateLangs provides a list of supported languages for template files (lower-case)
var SupportedTemplateLangs = []string{"tmpl", "hmpl", "mst"}

// IsSupportedTemplateLang provides the index of `SupportedTemplateLangs` that `lang` is at.
// If `lang` is not in `SupportedTemplateLangs`, `-1` will be returned.
// File extensions can be passed in `lang`, the prefixed `.` will be trimmed.
func IsSupportedTemplateLang(lang string) int {
	lang = strings.ToLower(lang)
	if len(lang) > 0 && lang[0] == '.' {
		lang = lang[1:]
	}
	for i, l := range SupportedTemplateLangs {
		if lang == l {
			return i
		}
	}
	return -1
}

func getTemplateType(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

// Template is a generic interface to any template parsed from LoadTemplateFile
type Template struct {
	Source   string
	Template interface{}
}

// Execute executes `t` against `d`. Reflection is used to determine
// the template type and call it's execution fuction.
func (t *Template) Execute(d interface{}) (result bytes.Buffer, err error) {
	var funcName string
	var params []reflect.Value
	switch (reflect.TypeOf(t.Template).String()) {
	case "*template.Template": // golang templates
		funcName = "Execute"
		params = []reflect.Value{reflect.ValueOf(&result), reflect.ValueOf(d)}
	case "*mustache.Template":
		funcName = "FRender"
		params = []reflect.Value{reflect.ValueOf(&result), reflect.ValueOf(d)}
	default:
		err = fmt.Errorf("unable to infer template type '%s'", reflect.TypeOf(t.Template).String())
	}

	if err == nil {
		rval := reflect.ValueOf(t.Template).MethodByName(funcName).Call(params)
		if !rval[0].IsNil() { // err != nil
			err = rval[0].Interface().(error)
		}
	}

	return
}

func loadTemplateFileTmpl(root string, partials ...string) (*tmpl.Template, error) {
	var stat os.FileInfo
	t, e := tmpl.ParseFiles(root)

	for i := 0; i < len(partials) && e == nil; i++ {
		p := partials[i]
		ptype := getTemplateType(p)

		stat, e = os.Stat(p)
		if e == nil {
			if ptype == "tmpl" || ptype == "gotmpl" {
				t, e = t.ParseFiles(p)
			} else if strings.Contains(p, "*") {
				t, e = t.ParseGlob(p)
			} else if stat.IsDir() {
				e = filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
					if err == nil && !info.IsDir() {
						ptype = getTemplateType(path)
						if ptype == "tmpl" || ptype == "gotmpl" {
							t, err = t.ParseFiles(path)
						}
					}
					return err
				})
			} else {
				return nil, fmt.Errorf("non-matching filetype (%s)", p)
			}
		}
	}

	return t, e
}

func loadTemplateFileHmpl(root string, partials ...string) (*hmpl.Template, error) {
	var stat os.FileInfo
	t, e := hmpl.ParseFiles(root)

	for i := 0; i < len(partials) && e == nil; i++ {
		p := partials[i]
		ptype := getTemplateType(p)

		stat, e = os.Stat(p)
		if e == nil {
			if ptype == "hmpl" || ptype == "gohmpl" {
				t, e = t.ParseFiles(p)
			} else if strings.Contains(p, "*") {
				t, e = t.ParseGlob(p)
			} else if stat.IsDir() {
				e = filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
					if err == nil && !info.IsDir() {
						ptype = getTemplateType(path)
						if ptype == "hmpl" || ptype == "gohmpl" {
							t, err = t.ParseFiles(path)
						}
					}
					return err
				})
			} else {
				return nil, fmt.Errorf("non-matching filetype (%s)", p)
			}
		}
	}

	return t, e
}

func loadTemplateFileMst(root string, partials ...string) (*mst.Template, error) {
	var err error
	for p, partial := range partials {
		if err != nil {
			break
		}

		if stat, e := os.Stat(partial); e != nil {
			partials = append(partials[:p], partials[p+1:]...)
			err = e
		} else if stat.IsDir() == false {
			partials[p] = filepath.Dir(partial)
		} else if strings.Contains(partial, "*") {
			if paths, e := filepath.Glob(partial); e != nil {
				partials = append(partials[:p], partials[p+1:]...)
				err = e
			} else {
				partials = append(partials[:p], partials[p+1:]...)
				partials = append(partials, paths...)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	mstfp := &mst.FileProvider{
		Paths:      partials,
		Extensions: []string{".mst", ".mustache"},
	}
	return mst.ParseFilePartials(root, mstfp)
}

// LoadTemplateFilepath loads a Template from file `root`. All files in `partials`
// that have the same template type (identified by file extension) are also
// parsed and associated with the parsed root template.
func LoadTemplateFilepath(rootPath string, partialPaths ...string) (t Template, e error) {
	if len(rootPath) == 0 {
		e = fmt.Errorf("no rootPath template specified")
	}
	if stat, err := os.Stat(rootPath); err != nil {
		e = err
	} else if stat.IsDir() {
		e = fmt.Errorf("rootPath path must be a file, not a directory: %s", rootPath)
	}

	if e != nil {
		return
	}

	t = Template{Source: rootPath}
	ttype := getTemplateType(rootPath)
	if ttype == "tmpl" || ttype == "gotmpl" {
		t.Template, e = loadTemplateFileTmpl(rootPath, partialPaths...)
	} else if ttype == "hmpl" || ttype == "gohmpl" {
		t.Template, e = loadTemplateFileHmpl(rootPath, partialPaths...)
	} else if ttype == "mst" || ttype == "mustache" {
		t.Template, e = loadTemplateFileMst(rootPath, partialPaths...)
	} else {
		e = fmt.Errorf("'%s' is not a supported template language", ttype)
	}

	return
}

// LoadTemplateString will convert `root` and `partials` data to io.StringReader variables and
// return a `LoadTemplate` call using them as parameters.
// The `partials` map should have the template name to assign the partial template to in the
// string key and the template data in as the value.
func LoadTemplateString(lang string, rootName string, root string, partials map[string]string) (t Template, e error) {
	p := make(map[string]io.Reader)
	for name, partial := range partials {
		p[name] = strings.NewReader(partial)
	}
	return LoadTemplate(lang, rootName, strings.NewReader(root), p)
}

// LoadTemplate loads a Template from `root` of type `lang`, named
// `name`. `lang` must be an element in `SupportedTemplateLangs`.
// `name` is optional, if empty the template name will be "template".
// `root` should be a string of template, with syntax matching that of
// `lang`. `partials` should be a string of template, with syntax
// matching that of `lang`.
func LoadTemplate(lang string, rootName string, root io.Reader, partials map[string]io.Reader) (t Template, e error) {
	if IsSupportedTemplateLang(lang) == -1 {
		e = fmt.Errorf("invalid type '%s'", lang)
		return
	}

	var buf []byte
	switch(lang) {
	case "tmpl":
		var template *tmpl.Template
		if buf, e = io.ReadAll(root); e != nil {
			break
		}
		if template, e = tmpl.New(rootName).Parse(string(buf)); e != nil {
			break
		}
		for name, p := range partials {
			if buf, e = io.ReadAll(p); e != nil {
				break
			}
			if _, e = template.New(name).Parse(string(buf)); e != nil {
				break
			}
		}
		t.Template = template
	case "hmpl":
		var template *hmpl.Template
		if buf, e = io.ReadAll(root); e != nil {
			break
		}
		if template, e = hmpl.New(rootName).Parse(string(buf)); e != nil {
			break
		}
		for name, p := range partials {
			if buf, e = io.ReadAll(p); e != nil {
				break
			}
			if _, e = template.New(name).Parse(string(buf)); e != nil {
				break
			}
		}
		t.Template = template
	case "mst":
		var template *mst.Template
		mstpp := new(mst.StaticProvider)
		mstpp.Partials = make(map[string]string)
		for name, partial := range partials {
			if buf, e = io.ReadAll(partial); e != nil {
				break
			}
			mstpp.Partials[name] = string(buf)
		}
		if e == nil {
			if buf, e = io.ReadAll(root); e != nil {
				break
			}
			template, e = mst.ParseStringPartials(string(buf), mstpp)
			t.Template = template
		}
	default:
		e = fmt.Errorf("'%s' is not a supported template language", lang)
	}

	return
}

