package main

import (
	"fmt"
	"os"
)

type Options struct {
	RootPaths       []string
	PartialPaths    []string
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

	options = parseArgs(os.Args[1:])
	if len(options.SortData) == 0 {
		options.SortData = "filename"
	}
}

func main() {

	_ = LoadDataFiles("", options.GlobalDataPaths...)
	_ = LoadDataFiles(options.SortData, options.DataPaths...)

	templates := make([]Template, 0)
	for _, r := range options.RootPaths {
		if t, e := LoadTemplateFile(r, options.PartialPaths...); e != nil {
			warn("unable to load templates (%s)", e)
		} else {
			templates = append(templates, t)
		}
	}

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
		} else if flag == "r" || flag == "root" {
			o.RootPaths = append(o.RootPaths, arg)
		} else if flag == "p" || flag == "partial" {
			o.PartialPaths = append(o.PartialPaths, arg)
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
