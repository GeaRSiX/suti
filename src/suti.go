package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Options provides all the different variables that need to be set by the
// user calling suti.
type Options struct {
	RootPath        string
	PartialPaths    []string
	GlobalDataPaths []string
	DataPaths       []string
	DataKey         string
	SortData        string
	ConfigFile      string
}

var options Options

func warn(msg string, args ...interface{}) {
	fmt.Println("WARNING", strings.TrimSuffix(fmt.Sprintf(msg, args...), "\n"))
}

func init() {
	if len(os.Args) <= 1 {
		print("nothing to do")
		os.Exit(0)
	}

	options = parseArgs(os.Args[1:], Options{})
	if len(options.ConfigFile) != 0 {
		var cfgln string
		cfgargs := make([]string, 0)
		if cfgf, err := os.Open(options.ConfigFile); err == nil {
			defer cfgf.Close()
			var err error
			for err != io.EOF {
				_, err = fmt.Fscanln(cfgf, &cfgln)
				for i, a := range strings.Split(cfgln, "=") {
					if i == 0 {
						a = "-" + a
					}
					cfgargs = append(cfgargs, a)
				}
			}
		} else {
			warn("unable to open config file (%s): %s", options.ConfigFile, err)
		}
		if len(cfgargs) > 0 {
			options = parseArgs(cfgargs, options)
		}
	}
	if len(options.SortData) == 0 {
		options.SortData = "filename"
	}
}

func main() {
	gd := LoadDataFiles("", options.GlobalDataPaths...)
	d := LoadDataFiles(options.SortData, options.DataPaths...)
	sd := GenerateSuperData(options.DataKey, d, gd...)

	if t, e := LoadTemplateFile(options.RootPath, options.PartialPaths...); e != nil {
		warn("unable to load templates (%s)", e)
	} else if out, err := ExecuteTemplate(t, sd); err != nil {
		warn("failed to execute template '%s' (%s)", options.RootPath, err)
	} else {
		fmt.Println(out.String())
	}

	return
}

// custom arg parser because golang.org/pkg/flag doesn't support list args
func parseArgs(args []string, existing Options) (o Options) {
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
			o.RootPath = arg
		} else if flag == "p" || flag == "partial" {
			o.PartialPaths = append(o.PartialPaths, arg)
		} else if flag == "gd" || flag == "globaldata" {
			o.GlobalDataPaths = append(o.GlobalDataPaths, arg)
		} else if flag == "d" || flag == "data" {
			o.DataPaths = append(o.DataPaths, arg)
		} else if flag == "dk" || flag == "datakey" && len(o.DataKey) == 0 {
			o.DataKey = arg
		} else if flag == "sd" || flag == "sortdata" && len(o.SortData) == 0 {
			o.SortData = arg
		} else if flag == "cfg" || flag == "config" && len(o.ConfigFile) == 0 {
			o.ConfigFile = arg
		} else if len(flag) == 0 {
			// skip unknown flag arguments
		} else {
			warn("ignoring flag: '%s'", flag)
			flag = ""
		}
	}

	return
}
