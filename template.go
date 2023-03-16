package dati

/*
Copyright (C) 2023 gearsix <gearsix@tuta.io>

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
	"errors"
	"fmt"
	hmpl "html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	tmpl "text/template"

	mst "github.com/cbroglie/mustache"
)

// TemplateLanguage provides a list of supported languages for
// Template files (lower-case)
type TemplateLanguage string

func (t TemplateLanguage) String() string {
	return string(t)
}

const (
	TMPL TemplateLanguage = "tmpl"
	HMPL TemplateLanguage = "hmpl"
	MST  TemplateLanguage = "mst"
)

var (
	ErrUnsupportedTemplate = func(format string) error {
		return fmt.Errorf("template language '%s' is not supported", format)
	}
	ErrUnknownTemplateType = func(templateType string) error {
		return fmt.Errorf("unable to infer template type '%s'", templateType)
	}
	ErrRootPathIsDir = func(path string) error {
		return fmt.Errorf("rootPath path must be a file, not a directory (%s)", path)
	}
	ErrNilTemplate = errors.New("template is nil")
)

// IsTemplateLanguage will return a bool if the file found at `path`
// is a known *TemplateLanguage*, based upon it's file extension.
func IsTemplateLanguage(path string) bool {
	return ReadTemplateLangauge(path) != ""
}

// ReadTemplateLanguage returns the *TemplateLanguage* that the file
// extension of `path` matches. If the file extension of `path` does
// not match any *TemplateLanguage*, then an "" is returned.
func ReadTemplateLangauge(path string) TemplateLanguage {
	if len(path) == 0 {
		return ""
	}

	ext := filepath.Ext(path)
	if len(ext) == 0 {
		ext = path // assume `path` the name of the format
	}

	ext = strings.ToLower(ext)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}

	for _, fmt := range []TemplateLanguage{TMPL, HMPL, MST} {
		if fmt.String() == ext {
			return fmt
		}
	}
	return ""
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
func (t *Template) Execute(data interface{}) (result bytes.Buffer, err error) {
	var funcName string
	var params []reflect.Value
	tType := reflect.TypeOf(t.T)
	if tType == nil {
		err = ErrNilTemplate
		return
	}
	switch tType.String() {
	case "*template.Template": // golang templates
		funcName = "Execute"
		params = []reflect.Value{reflect.ValueOf(&result), reflect.ValueOf(data)}
	case "*mustache.Template":
		funcName = "FRender"
		params = []reflect.Value{reflect.ValueOf(&result), reflect.ValueOf(data)}
	default:
		err = ErrUnknownTemplateType(reflect.TypeOf(t.T).String())
	}

	if err == nil {
		rval := reflect.ValueOf(t.T).MethodByName(funcName).Call(params)
		if !rval[0].IsNil() { // err != nil
			err = rval[0].Interface().(error)
		}
	}

	return
}

// ExecuteToFile writes the result of `(*Template).Execute(data)` to the file at `path` (if no errors occurred).
// If `force` is true, any existing file at `path` will be overwritten.
func (t *Template)ExecuteToFile(data interface{}, path string, force bool) (file *os.File, err error) {
	if f, err := os.Open(path); os.IsNotExist(err) {
		f, err = os.Create(path)
	} else if !force {
		err = os.ErrExist
	} else { // overwrite existing file data
		if err = f.Truncase(0); err == nil {
			_, err = f.Seek(0, 0)i
		}
	}

	if err != nil {
		return
	}
	defer f.Close()

	var out bytes.Buffer
	if out, err = t.Execute(data); err != nil {
		f = nil
	} else {
		_, err = f.Write(out.Bytes())
	}

	return
}

// LoadTemplateFilepath loads a Template from file `root`. All files in `partials`
// that have the same template type (identified by file extension) are also
// parsed and associated with the parsed root template.
func LoadTemplateFile(rootPath string, partialPaths ...string) (t Template, err error) {
	var stat os.FileInfo
	if stat, err = os.Stat(rootPath); err != nil {
		return
	} else if stat.IsDir() {
		err = ErrRootPathIsDir(rootPath)
		return
	}

	lang := ReadTemplateLangauge(rootPath)

	rootName := strings.TrimSuffix(filepath.Base(rootPath), filepath.Ext(rootPath))

	var root *os.File
	if root, err = os.Open(rootPath); err != nil {
		return
	}
	defer root.Close()

	partials := make(map[string]io.Reader)
	for _, path := range partialPaths {
		name := filepath.Base(path)
		if lang == "mst" {
			name = strings.TrimSuffix(name, filepath.Ext(name))
		}

		if _, err = os.Stat(path); err != nil {
			return
		}

		var partial *os.File
		if partial, err = os.Open(path); err != nil {
			return
		}
		defer partial.Close()
		partials[name] = partial
	}

	return LoadTemplate(lang, rootName, root, partials)
}

// LoadTemplateString will convert `root` and `partials` data to io.StringReader variables and
// return a `LoadTemplate` call using them as parameters.
// The `partials` map should have the template name to assign the partial template to in the
// string key and the template data in as the value.
func LoadTemplateString(lang TemplateLanguage, rootName string, root string, partials map[string]string) (t Template, e error) {
	p := make(map[string]io.Reader)
	for name, partial := range partials {
		p[name] = strings.NewReader(partial)
	}
	return LoadTemplate(lang, rootName, strings.NewReader(root), p)
}

// LoadTemplate loads a Template from `root` of type `lang`, named `name`.
// `lang` must be an element in `SupportedTemplateLangs`.
// `name` is optional, if empty the template name will be "template".
// `root` should be a string of template, with syntax matching that of `lang`.
// `partials` should be a string of template, with syntax matching that of `lang`.
func LoadTemplate(lang TemplateLanguage, rootName string, root io.Reader, partials map[string]io.Reader) (t Template, err error) {
	t.Name = rootName

	switch TemplateLanguage(lang) {
	case TMPL:
		t.T, err = loadTemplateTmpl(rootName, root, partials)
	case HMPL:
		t.T, err = loadTemplateHmpl(rootName, root, partials)
	case MST:
		t.T, err = loadTemplateMst(rootName, root, partials)
	default:
		err = ErrUnsupportedTemplate(lang.String())
	}

	return
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

	mstprv := new(mst.StaticProvider)
	mstprv.Partials = make(map[string]string)
	for name, partial := range partials {
		if buf, err := ioutil.ReadAll(partial); err != nil {
			return nil, err
		} else {
			mstprv.Partials[name] = string(buf)
		}
	}

	if buf, err := ioutil.ReadAll(root); err != nil {
		return nil, err
	} else if template, err = mst.ParseStringPartials(string(buf), mstprv); err != nil {
		return nil, err
	}

	return template, nil
}

