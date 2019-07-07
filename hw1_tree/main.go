package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

func dirTree(out io.Writer, path string, printFiles bool) error {
	return walk(out, path, printFiles, 0, make([]string, 1))
}

func walk(out io.Writer, path string, printFiles bool, level int, prefixes []string) error {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	files := make([]os.FileInfo, 0)
	for _, v := range dir {
		if !printFiles && !v.IsDir() {
			continue
		}
		files = append(files, v)
	}
	sort.Sort(byName(files))
	len := len(files)
	for i, f := range files {
		if !printFiles && !f.IsDir() {
			continue
		}
		if i == len-1 {
			prefixes[level] = prefixLast
		} else {
			prefixes[level] = prefixCur
		}
		var size string
		if !f.IsDir() {
			size = " (empty)"
			if f.Size() != 0 {
				size = fmt.Sprintf(" (%vb)", f.Size())
			}
		}
		fmt.Fprint(out, strings.Join(prefixes, ""), f.Name(), size, "\n")
		if i == len-1 {
			prefixes[level] = prefixEmpty
		} else {
			prefixes[level] = prefixPrev
		}
		if f.IsDir() {
			err = walk(out, path+string(os.PathSeparator)+f.Name(), printFiles, level+1, append(prefixes, prefixPrev))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const prefixEmpty = "\t"
const prefixPrev = "│\t"
const prefixCur = "├───"
const prefixLast = "└───"

type byName []os.FileInfo

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }

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
