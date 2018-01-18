package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var gIndent = ""

func dirTree(out io.Writer, path string, printFiles bool) error {
	indent := "│	"
	prefix := "├───"
	files, err := ioutil.ReadDir(path)

	if err != nil {
		log.Fatal(err)
		return fmt.Errorf("error while read directory")
	}

	if !printFiles {
		tempFiles := make([]os.FileInfo, 0, len(files))
		for _, f := range files {
			if f.IsDir() {
				tempFiles = append(tempFiles, f)
			}
		}
		files = make([]os.FileInfo, len(tempFiles))
		copy(files, tempFiles)
	}

	postfix := ""
	for i, f := range files {
		if i == len(files)-1 {
			prefix = "└───"
			indent = "\t"
		}
		if f.IsDir() {
			fmt.Fprintln(out, gIndent+prefix+f.Name())
			gIndent = gIndent + indent
			dirTree(out, path+string(os.PathSeparator)+f.Name(), printFiles)
			gIndent = gIndent[:len(gIndent)-len(indent)]
		} else {
			if printFiles {
				if f.Size() == 0 {
					postfix = " (empty)"
				} else {
					postfix = " (" + fmt.Sprint(f.Size()) + "b)"
				}
			}
			fmt.Fprintln(out, gIndent+prefix+f.Name()+postfix)
		}
	}
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
