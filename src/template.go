package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	tmpl "text/template"
	hmpl "html/template"
)

type template interface{}

func getTemplateType(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func LoadTemplateFile(root string, partials ...string) (t template, e error) {
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
