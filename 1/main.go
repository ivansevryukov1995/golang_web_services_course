package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func sortFiles(path string, printFiles bool) ([]fs.FileInfo, error) {
	dir, err := os.Open(path)
	if err != nil {
		return []fs.FileInfo{}, err
	}
	defer dir.Close()

	files, err := dir.Readdir(0)
	if err != nil {
		return []fs.FileInfo{}, err
	}

	if !printFiles {
		tempFiles := []os.FileInfo{}
		for _, file := range files {
			if file.IsDir() {
				tempFiles = append(tempFiles, file)
			}
		}
		files = tempFiles
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	return files, nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	var prefix, sizeSufix, endFile string

	files, err := sortFiles(path, printFiles)
	if err != nil {
		return err
	}

	for i, file := range files {
		var tab string

		if i == len(files)-1 {
			prefix = "└───"
		} else {
			prefix = "├───"
		}

		if !file.IsDir() && file.Size() != 0 {
			sizeSufix = fmt.Sprintf(" (%db)", file.Size())
		} else if !file.IsDir() && file.Size() == 0 {
			sizeSufix = " (empty)"
		} else {
			sizeSufix = ""
		}

		for j := 0; j < len(strings.Split(path+string(os.PathSeparator)+file.Name(), "\\"))-2; j++ {

			parent := filepath.Join(strings.Split(path+string(os.PathSeparator)+file.Name(), "\\")[:j+1]...)

			filesParent, err := sortFiles(parent, printFiles)
			if err != nil {
				return err
			}

			for n, fileParent := range filesParent {
				if n == len(filesParent)-1 {
					endFile = fileParent.Name()
				}
			}

			if strings.Split(path+string(os.PathSeparator)+file.Name(), "\\")[j+1] != endFile {
				tab += "│\t"
			} else {
				tab += "\t"
			}

		}

		_, err := fmt.Fprintf(out, fmt.Sprintf("%s%s%s%s\n", tab, prefix, file.Name(), sizeSufix))
		if err != nil {
			return err
		}

		if file.IsDir() {
			err := dirTree(out, path+string(os.PathSeparator)+file.Name(), printFiles)
			if err != nil {
				return err
			}
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
