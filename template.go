package dati

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
	"io/ioutil"
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

// Template is a wrapper to interface with any template parsed by dati.
// Ideally it would have just been an interface{} that defines Execute but
// the libaries being used aren't that uniform.
type Template struct {
	Name string
	T    interface{}
}

// Execute executes `t` against `d`. Reflection is used to determine
// the template type and call it's execution fuction.
func (t *Template) Execute(d interface{}) (result bytes.Buffer, err error) {
	var funcName string
	var params []reflect.Value
	tType := reflect.TypeOf(t.T)
	if tType == nil {
		err = fmt.Errorf("template.T is nil")
		return
	}
	switch tType.String() {
	case "*template.Template": // golang templates
		funcName = "Execute"
		params = []reflect.Value{reflect.ValueOf(&result), reflect.ValueOf(d)}
	case "*mustache.Template":
		funcName = "FRender"
		params = []reflect.Value{reflect.ValueOf(&result), reflect.ValueOf(d)}
	default:
		err = fmt.Errorf("unable to infer template type '%s'", reflect.TypeOf(t.T).String())
	}

	if err == nil {
		rval := reflect.ValueOf(t.T).MethodByName(funcName).Call(params)
		if !rval[0].IsNil() { // err != nil
			err = rval[0].Interface().(error)
		}
	}

	return
}

// LoadTemplateFilepath loads a Template from file `root`. All files in `partials`
// that have the same template type (identified by file extension) are also
// parsed and associated with the parsed root template.
func LoadTemplateFilepath(rootPath string, partialPaths ...string) (t Template, e error) {
	var stat os.FileInfo
	if stat, e = os.Stat(rootPath); e != nil {
		return
	} else if stat.IsDir() {
		e = fmt.Errorf("rootPath path must be a file, not a directory: %s", rootPath)
		return
	}

	lang := strings.TrimPrefix(filepath.Ext(rootPath), ".")

	rootName := strings.TrimSuffix(filepath.Base(rootPath), filepath.Ext(rootPath))

	var root *os.File
	if root, e = os.Open(rootPath); e != nil {
		return
	}
	defer root.Close()

	partials := make(map[string]io.Reader)
	for _, path := range partialPaths {
		name := filepath.Base(path)
		if lang == "mst" {
			name = strings.TrimSuffix(name, filepath.Ext(name))
		}

		if stat, e = os.Stat(path); e != nil {
			return
		}

		var p *os.File
		if p, e = os.Open(path); e != nil {
			return
		}
		defer p.Close()
		partials[name] = p
	}

	return LoadTemplate(lang, rootName, root, partials)
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

func loadTemplateTmpl(rootName string, root io.Reader, partials map[string]io.Reader) (*tmpl.Template, error) {
	var template *tmpl.Template

	if buf, err := ioutil.ReadAll(root); err != nil {
		return nil, err
	} else if template, err = tmpl.New(rootName).Parse(string(buf)); err != nil {
		return nil, err
	}

	for name, partial := range partials {
		if buf, err := ioutil.ReadAll(partial); err != nil {
			return nil, err
		} else if _, err = template.New(name).Parse(string(buf)); err != nil {
			return nil, err
		}
	}

	return template, nil
}

func loadTemplateHmpl(rootName string, root io.Reader, partials map[string]io.Reader) (*hmpl.Template, error) {
	var template *hmpl.Template

	if buf, err := ioutil.ReadAll(root); err != nil {
		return nil, err
	} else if template, err = hmpl.New(rootName).Parse(string(buf)); err != nil {
		return nil, err
	}

	for name, partial := range partials {
		if buf, err := ioutil.ReadAll(partial); err != nil {
			return nil, err
		} else if _, err = template.New(name).Parse(string(buf)); err != nil {
			return nil, err
		}
	}

	return template, nil
}

func loadTemplateMst(rootName string, root io.Reader, partials map[string]io.Reader) (*mst.Template, error) {
	var template *mst.Template

	mstpp := new(mst.StaticProvider)
	mstpp.Partials = make(map[string]string)
	for name, partial := range partials {
		if buf, err := ioutil.ReadAll(partial); err != nil {
			return nil, err
		} else {
			mstpp.Partials[name] = string(buf)
		}
	}

	if buf, err := ioutil.ReadAll(root); err != nil {
		return nil, err
	} else if template, err = mst.ParseStringPartials(string(buf), mstpp); err != nil {
		return nil, err
	}

	return template, nil
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

	t.Name = rootName
	switch lang {
	case "tmpl":
		t.T, e = loadTemplateTmpl(rootName, root, partials)
	case "hmpl":
		t.T, e = loadTemplateHmpl(rootName, root, partials)
	case "mst":
		t.T, e = loadTemplateMst(rootName, root, partials)
	default:
		e = fmt.Errorf("'%s' is not a supported template language", lang)
	}

	return
}
