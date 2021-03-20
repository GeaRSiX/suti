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
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"os"
	"strings"
)

type options struct {
	RootPath        string
	PartialPaths    []string
	GlobalDataPaths []string
	DataPaths       []string
	DataKey         string
	SortData        string
	ConfigFile      string
}

var opts options
var cwd string

func warn(msg string, args ...interface{}) {
	fmt.Println("WARNING", strings.TrimSuffix(fmt.Sprintf(msg, args...), "\n"))
}

func basedir(path string) string {
	var err error
	if !filepath.IsAbs(path) {
		if path, err = filepath.Rel(cwd, path); err != nil {
			warn("failed to parse path '%s': %s", path, err)
		}
	}
	return path
}

func init() {
	if len(os.Args) <= 1 {
		print("nothing to do")
		os.Exit(0)
	}

	cwd = "."
	opts = parseArgs(os.Args[1:], options{})
	if len(opts.ConfigFile) != 0 {
		cwd = filepath.Dir(opts.ConfigFile)
		opts = parseConfig(opts.ConfigFile, opts)
	}
	opts = setDefaultOptions(opts)
}

func main() {
	gd := LoadDataFiles("", opts.GlobalDataPaths...)
	d := LoadDataFiles(opts.SortData, opts.DataPaths...)
	sd := GenerateSuperData(opts.DataKey, d, gd...)

	if t, e := LoadTemplateFile(opts.RootPath, opts.PartialPaths...); e != nil {
		warn("unable to load templates (%s)", e)
	} else if out, err := ExecuteTemplate(t, sd); err != nil {
		warn("failed to execute template '%s' (%s)", opts.RootPath, err)
	} else {
		fmt.Println(out.String())
	}

	return
}

// custom arg parser because golang.org/pkg/flag doesn't support list args
func parseArgs(args []string, existing options) (o options) {
	o = existing
	var flag string
	for a := 0; a < len(args); a++ {
		arg := args[a]
		if arg[0] == '-' && flag != "--" {
			flag = arg
			ndelims := 0
			for len(flag) > 0 && flag[0] == '-' {
				flag = flag[1:]
				ndelims++
			}

			if ndelims > 2 {
				warn("bad flag syntax: '%s'", arg)
				flag = ""
			}

			// set valid any flags that don't take arguments here
		} else if (flag == "r" || flag == "root") && len(o.RootPath) == 0 {
			o.RootPath = basedir(arg)
		} else if flag == "p" || flag == "partial" {
			o.PartialPaths = append(o.PartialPaths, basedir(arg))
		} else if flag == "gd" || flag == "globaldata" {
			o.GlobalDataPaths = append(o.GlobalDataPaths, basedir(arg))
		} else if flag == "d" || flag == "data" {
			o.DataPaths = append(o.DataPaths, arg)
		} else if flag == "dk" || flag == "datakey" && len(o.DataKey) == 0 {
			o.DataKey = arg
		} else if flag == "sd" || flag == "sortdata" && len(o.SortData) == 0 {
			o.SortData = arg
		} else if flag == "cfg" || flag == "config" && len(o.ConfigFile) == 0 {
			o.ConfigFile = basedir(arg)
		} else if len(flag) == 0 {
			// skip unknown flag arguments
		} else {
			warn("ignoring flag: '%s'", flag)
			flag = ""
		}
	}

	return
}

func parseConfig(fpath string, existing options) options {
	var err error
	var cfgf *os.File
	if cfgf, err = os.Open(fpath); err != nil {
		warn("error loading config file '%s': %s", fpath, err)
		err = io.EOF
	}
	defer cfgf.Close()

	var args []string
	scanf := bufio.NewScanner(cfgf)
	for scanf.Scan() {
		for i, arg := range strings.Split(scanf.Text(), "=") {
			arg = strings.TrimSpace(arg)
			if i == 0 {
				arg = "-" + arg
			}
			args = append(args, arg)
		}
	}
	return parseArgs(args, existing)
}

func setDefaultOptions(o options) options {
	if len(o.SortData) == 0 {
		o.SortData = "filename"
	}
	if len(o.DataKey) == 0 {
		o.DataKey = "data"
	}
	return o
}
