package main

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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"notabug.org/gearsix/dati"
)

// Data is just a generic map for key/value data
type Data map[string]interface{}

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

func warn(err error, msg string, args ...interface{}) {
	warning := "WARNING "
	if len(msg) > 0 {
		warning += strings.TrimSuffix(fmt.Sprintf(msg, args...), "\n")
		if err != nil {
			warning += ": "
		}
	}
	if err != nil {
		warning += err.Error()
	}
	fmt.Println(warning)
}

func assert(err error, msg string, args ...interface{}) {
	if err != nil {
		fmt.Printf("ERROR %s\n%s\n", strings.TrimSuffix(fmt.Sprintf(msg, args...), "\n"), err)
		os.Exit(1)
	}
}

func basedir(path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}
	return path
}

func init() {
	if len(os.Args) <= 1 {
		fmt.Println("nothing to do")
		os.Exit(0)
	}

	opts = parseArgs(os.Args[1:], options{})
	if len(opts.ConfigFile) != 0 {
		cwd = filepath.Dir(opts.ConfigFile)
		opts = parseConfig(opts.ConfigFile, opts)
	}
	opts = setDefaultOptions(opts)
}

func main() {
	var err error
	var global Data
	var data []Data
	var template dati.Template
	var out bytes.Buffer

	opts.GlobalDataPaths = loadFilePaths(opts.GlobalDataPaths...)
	for _, path := range opts.GlobalDataPaths {
		var d Data
		err = dati.LoadDataFile(path, &d)
		assert(err, "failed to load global data '%s'", path)
		data = append(data, d)
	}
	global = mergeData(data)

	opts.DataPaths = loadFilePaths(opts.DataPaths...)
	opts.DataPaths, err = dati.SortFileList(opts.DataPaths, opts.SortData)
	if err != nil {
		warn(err, "failed to sort data files")
	}
	data = make([]Data, 0)
	for _, path := range opts.DataPaths {
		var d Data
		err = dati.LoadDataFile(path, &d)
		assert(err, "failed to load data '%s'", path)
		data = append(data, d)
	}
	global[opts.DataKey] = data

	template, err = dati.LoadTemplateFile(opts.RootPath, opts.PartialPaths...)
	assert(err, "unable to load templates")

	out, err = template.Execute(global)
	assert(err, "failed to execute template '%s'", opts.RootPath)
	fmt.Print(out.String())

	return
}

func help() {
	fmt.Print("Usage: dati [OPTIONS]\n\n")

	fmt.Print("Options")
	fmt.Print(`
   -r path, -root path  
    path of template file to execute against.

  -p path..., -partial path...  
    path of (multiple) template files that are called upon by at least one
    root template. If a directory is passed then all files within that
    directory will (recursively) be loaded.

  -gd path..., -global-data path...  
    path of (multiple) data files to load as "global data". If a directory is
    passed then all files within that directory will (recursively) be loaded.

  -d path..., -data path...  
   path of (multiple) data files to load as "data". If a directory is passed
   then all files within that directory will (recursively) be loaded.

  -dk name, -data-key name  
    set the name of the key used for the generated array of data (default:
    "data")

  -sd attribute, -sort-data attribute  
    The file attribute to order data files by. If no value is provided, the data
    will be provided in the order it's loaded.
    Accepted values: "filename", "modified".
    A suffix can be appended to each value to set the sort order: "-asc" (for
    ascending), "-desc" (for descending). If not specified, this defaults to
    "-asc".
  -cfg file, -config file  
    A data file to provide default values for the above options (see CONFIG).

`)

	fmt.Println("See doc/dati.txt for further details")
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
				warn(nil, "bad flag syntax: '%s'", arg)
				flag = ""
			}

			if strings.Contains(flag, "=") {
				split := strings.SplitN(flag, "=", 2)
				flag = split[0]
				args[a] = split[1]
				a--
			}

			// set valid any flags that don't take arguments here
			if flag == "h" || flag == "help" {
				help()
				os.Exit(0)
			}
		} else if (flag == "r" || flag == "root") && len(o.RootPath) == 0 {
			o.RootPath = basedir(arg)
		} else if flag == "p" || flag == "partial" {
			o.PartialPaths = append(o.PartialPaths, basedir(arg))
		} else if flag == "gd" || flag == "globaldata" {
			o.GlobalDataPaths = append(o.GlobalDataPaths, basedir(arg))
		} else if flag == "d" || flag == "data" {
			o.DataPaths = append(o.DataPaths, basedir(arg))
		} else if flag == "dk" || flag == "datakey" && len(o.DataKey) == 0 {
			o.DataKey = arg
		} else if flag == "sd" || flag == "sortdata" && len(o.SortData) == 0 {
			o.SortData = arg
		} else if flag == "cfg" || flag == "config" && len(o.ConfigFile) == 0 {
			o.ConfigFile = basedir(arg)
		} else if len(flag) == 0 {
			// skip unknown flag arguments
		} else {
			warn(nil, "ignoring flag: '%s'", flag)
			flag = ""
		}
	}

	return
}

func parseConfig(fpath string, existing options) options {
	var err error
	var cfgf *os.File
	if cfgf, err = os.Open(fpath); err != nil {
		warn(err, "error loading config file '%s'", fpath)
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

// load glob & dir filepaths as individual filepaths
func loadFilePaths(paths ...string) (filepaths []string) {
	for _, path := range paths {
		var err error
		if strings.Contains(path, "*") {
			var glob []string
			glob, err = filepath.Glob(path)
			assert(err, "failed to glob '%s'", path)
			for _, p := range glob {
				filepaths = append(filepaths, p)
			}
		} else {
			err = filepath.Walk(path,
				func(p string, info os.FileInfo, e error) error {
					if e == nil && !info.IsDir() {
						filepaths = append(filepaths, p)
					}
					return e
				})
		}
		if err != nil {
			assert(err, "failed to load filepaths for '%s'", path)
		}
	}
	return
}

func mergeData(data []Data) (merged Data) {
	merged = make(Data)
	for _, d := range data {
		for key, val := range d {
			if merged[key] == nil {
				merged[key] = val
			} else {
				warn(nil, "merge conflict for global data key: '%s'", key)
			}
		}
	}
	return
}
