package main

import (
	"go/parser"
	"go/token"
	"log"
	"path"
	"strconv"
	"strings"

	"github.com/kr/fs"
)

// scan walks a golang project directory, collecting external package imports.
func scan(dir string) ([]string, error) {

	var pkglist []string

	// create a new walk.
	w := fs.Walk(dir)

	// start the walk down the directory tree.
	for w.Step() {

		// check for errors.
		if w.Err() != nil {
			log.Println("govend scan:", w.Err())
			continue
		}

		// check the file is a .go file.
		if !w.Stat().IsDir() && strings.HasSuffix(w.Path(), ".go") {

			// create an empty fileset.
			fset := token.NewFileSet()

			// parse only the import declarations in the .go file.
			f, err := parser.ParseFile(fset, w.Path(), nil, parser.ImportsOnly)
			if err != nil {
				return nil, err
			}

			// iterate through the files import paths.
			for _, i := range f.Imports {

				// unquote the import path value.
				importpath, err := strconv.Unquote(i.Path.Value)
				if err != nil {
					return nil, err
				}

				// determine the name of the package.
				name := path.Base(importpath)

				// if the package is part of the golang standard library, skip it.
				if stdpkg, ok := stdpkgs[name]; ok {
					for _, pkg := range stdpkg {
						if importpath == pkg.path {
							continue
						}
					}
				}

				// iterate through the known external packages.
				for _, pkg := range pkglist {

					// check if package path already exists, skip the append.
					if importpath == pkg {
						goto SKIP
					}
				}

				// if the import path doens't exists in pkgs, add it.
				pkglist = append(pkglist, importpath)

			SKIP: // skips the appending of packages that are already present.
			}
		}
	}

	return pkglist, nil
}
