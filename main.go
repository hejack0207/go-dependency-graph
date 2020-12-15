package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	showStdLib = flag.Bool("show-std", false, "Show dependencies os Standard Library")
	depLevel   = flag.Int("level", -1, "Dept of Dependency Graph")
	ignorePkgs = flag.String("ignore", "", "Ignore packages in dependency graph")

	ignoredPkgs  = map[string]bool{}
	pkgList      = map[string]bool{}
	graphList    = map[string]bool{}
	pkgDeps      = make(map[string][]string)
	buildContext = build.Default
)

func getImports(pkg *packages.Package) []string {
	allImports := pkg.Imports
	// fmt.Println("All Imports", allImports)
	ret := []string{}
	for key, _ := range allImports {
		ret = append(ret, key)
	}
	return ret
}

func processEachPackage(dir string, pkgName string) error {
	// fmt.Println("Directory: ", dir)
	// fmt.Println("Current Package Processing: ", pkgName)

	if ignoredPkgs[pkgName] {
		return nil
	}

	// pkg, err := buildContext.Import(pkgName, dir, 0)
	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: false,
		Dir:   dir,
	}

	initial, err := packages.Load(cfg, dir)
	if err != nil {
		return fmt.Errorf("Failed to import: %s", pkgName)
	}

	pkgList[pkgName] = true

	// if initial.Goroot && !*showStdLib {
	//         return nil
	// }
	if packages.PrintErrors(initial) > 0 {
		return fmt.Errorf("packages contain errors")
	}

	pkg := initial[0]
	pkgImports := getImports(pkg)
	pkgDeps[pkgName] = pkgImports

	for _, subPack := range pkgImports {
		if _, pkgExist := pkgList[subPack]; !pkgExist {
			if err := processEachPackage(pkg.PkgPath, subPack); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Invalid Arguments!")
	} else {
		fmt.Println("Arguments: ", args)
	}

	if *ignorePkgs != "" {
		for _, pkg := range strings.Split(*ignorePkgs, ",") {
			ignoredPkgs[pkg] = true
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current working directory")
	}

	for _, pkgName := range args {
		err := processEachPackage(cwd, pkgName)
		ShowGoDeps(pkgName, *depLevel)
		ProcessGoGraph(pkgName, *depLevel)
		if err != nil {
			fmt.Println("Error while processing the: ", pkgName, err)
			break
		}
	}

	// fmt.Println("Package Details:")
	// fmt.Println(pkgDeps)
}
