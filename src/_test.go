package main

import (
	"os"
	"testing"
)

func writeFile(path string, data string) (e error) {
	var f *os.File

	if f, e = os.Create(path); e != nil {
		return
	}
	if _, e = f.WriteString(data); e != nil {
		return
	}
	f.Close()

	return
}
