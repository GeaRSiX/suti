package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type data interface{}
type Options struct {
	TemplatePaths   []string
	GlobalDataPaths []string
	DataPaths       []string
	DataKey         string
	SortData        string
	ConfigFile      string
}

var options Options

func init() {
	if len(os.Args) <= 1 {
		print("nothing to do")
		os.Exit(0)
	}
}

func main() {
	options := parseArgs(os.Args[1:])
	d := LoadDataFiles(options.DataPaths)
	fmt.Println(d)

	return
}

func warn(msg string, args ...interface{}) {
	fmt.Println("WARNING", fmt.Sprintf(msg, args...))
	return
}

// custom arg parser because golang.org/pkg/flag doesn't support list args
func parseArgs(args []string) (o Options) {
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

		} else if flag == "t" || flag == "template" {
			o.TemplatePaths = append(o.TemplatePaths, arg)
		} else if flag == "gd" || flag == "globaldata" {
			o.GlobalDataPaths = append(o.GlobalDataPaths, arg)
		} else if flag == "d" || flag == "data" {
			o.DataPaths = append(o.DataPaths, arg)
		} else if flag == "dk" || flag == "datakey" {
			o.DataKey = arg
		} else if flag == "sd" || flag == "sortdata" {
			o.SortData = arg
		} else if flag == "cfg" || flag == "config" {
			o.ConfigFile = arg
		} else if len(flag) == 0 {
			// skip unknown flag arguments
		} else {
			warn("unknown flag: '%s'", flag)
			flag = ""
		}
	}

	return
}

// LoadDataFiles TODO
func LoadDataFiles(paths []string) map[string]data {
	var err error
	var stat os.FileInfo
	var d data
	
	loaded := make(map[string]data)

	for _, dpath := range paths {
		if stat, err = os.Stat(dpath); err != nil {
			warn("skipping data file '%s' (%s)", dpath, err)
			continue
		}

		if stat.IsDir() {
			_ = filepath.Walk(dpath,
				func(path string, info os.FileInfo, e error) error {
					if e == nil && !info.IsDir() {
						if d, e = LoadDataFile(path); e == nil {
							loaded[path] = d
						}
					}

					if e != nil {
						warn("unable to load data from file '%s' (%s)", path, e)
					}

					return e
				})
		} else {
			if d, err = LoadDataFile(dpath); err == nil {
				loaded[dpath] = d
			} else {
				warn("unable to load data from file '%s' (%s)", dpath, err)
			}
		}
	}

	return loaded
}

// LoadDataFile TODO
func LoadDataFile(path string) (d data, e error) {
	var f *os.File

	f, e = os.Open(path)
	defer f.Close()
	if e == nil {
		dtype := strings.TrimPrefix(filepath.Ext(path), ".")
		d, e = LoadData(dtype, f)
	}

	return d, e
}

// LoadData TODO
func LoadData(lang string, in io.Reader) (d data, e error) {
	var fbuf []byte
	if fbuf, e = ioutil.ReadAll(in); e != nil {
		return
	}

	if lang == "json" {
		e = json.Unmarshal(fbuf, &d)
	} else {
		e = fmt.Errorf("%s is not a supported data language", lang)
	}

	return
}
