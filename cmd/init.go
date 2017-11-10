package cmd

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/swanwish/go-common/logs"
	"github.com/swanwish/go-common/utils"
	"github.com/swanwish/godeps/common"
	"github.com/swanwish/godeps/models/godeps"
	"github.com/urfave/cli"
)

var (
	Init = cli.Command{
		Name:        "init",
		Usage:       "Init godeps.json file according the external packages for local project",
		Description: "This command will create a godeps.json file according the external packages for local project",
		Action:      runInit,
		Flags:       []cli.Flag{},
	}
)

func runInit(c *cli.Context) error {
	systemPackages, err := GetSystemPackages()
	if err != nil {
		logs.Errorf("Failed to get system packages, the error is %v", err)
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		logs.Errorf("Failed to get wd, the error is %v", err)
		return err
	}
	externalPackages, err := GetExternalPackages(wd, systemPackages)
	if err != nil {
		logs.Errorf("Failed to get external packages, the error is %v", err)
	}
	depItems, err := GetDepItems(externalPackages)
	goDeps, err := godeps.NewGoDeps()
	if err != nil {
		logs.Errorf("Failed to create godeps, the error is %v", err)
		return err
	}
	for _, item := range depItems {
		err = goDeps.AddItem(item.Path, item.Origin)
		if err != nil && err != common.ErrAlreadyExist {
			return err
		}
	}
	return goDeps.Save()
}

func GetSystemPackages() ([]string, error) {
	packages := []string{}
	envGoRoot := os.Getenv("GOROOT")
	if envGoRoot == "" {
		logs.Errorf("The GOROOT env is not specified")
		return nil, common.ErrInvalidParameter
	}
	rootSrc := filepath.Join(envGoRoot, "src")
	subFiles, err := ioutil.ReadDir(rootSrc)
	if err != nil {
		logs.Errorf("Failed to list go root, the error is %v", err)
		return nil, err
	}
	for _, subFile := range subFiles {
		if subFile.IsDir() {
			packages = append(packages, fmt.Sprintf("%s", subFile.Name()))
		}
	}
	return packages, nil
}

func GetExternalPackages(wd string, systemPackages []string) ([]string, error) {
	srcIndex := strings.Index(wd, "src/")
	if srcIndex == -1 {
		logs.Errorf("Invalid path, not inside src folder")
		return nil, common.ErrInvalidParameter
	}
	currentPackage := wd[srcIndex+4:]
	importedPackages := ""
	importedSystemPackages := []string{}
	externalPackages := []string{}
	filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		relativePath := strings.TrimPrefix(path, wd)
		if strings.HasPrefix(relativePath, "/") {
			relativePath = relativePath[1:]
		}
		if strings.HasPrefix(relativePath, "vendor/") {
			return nil
		}
		if strings.HasPrefix(relativePath, ".") {
			return nil
		}
		if filepath.Ext(relativePath) != ".go" {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), relativePath, nil, parser.ImportsOnly)
		if err != nil {
			logs.Errorf("Failed to parse file %s, the error is %v", relativePath, err)
			return err
		}
		for _, importSpec := range f.Imports {
			pkg := strings.Trim(importSpec.Path.Value, "\"")
			if strings.Index(importedPackages, fmt.Sprintf(",%s,", pkg)) != -1 {
				continue
			}
			importedPackages = fmt.Sprintf("%s,%s,", importedPackages, pkg)
			if strings.Index(pkg, "/") == -1 {
				continue
			}
			isSystemPackage := false
			for _, systemPackage := range systemPackages {
				if strings.HasPrefix(pkg, systemPackage) {
					importedSystemPackages = append(importedSystemPackages, pkg)
					isSystemPackage = true
				}
			}
			if !isSystemPackage {
				if !strings.HasPrefix(pkg, currentPackage) {
					externalPackages = append(externalPackages, pkg)
				}
			}
		}
		return nil
	})
	sort.Strings(externalPackages)
	return externalPackages, nil
}

func GetDepItems(externalPackages []string) ([]godeps.DepItem, error) {
	depItems := []godeps.DepItem{}
	envGoPath := os.Getenv("GOPATH")
	paths := strings.Split(envGoPath, fmt.Sprintf("%c", os.PathListSeparator))
	solvedPackages := []string{}
	for _, externalPackage := range externalPackages {
		found := false
		for index := 0; index < len(solvedPackages) && !found; index++ {
			if strings.HasPrefix(externalPackage, solvedPackages[index]) {
				found = true
			}
		}
		if found {
			continue
		}
		foundGitPath := false
		for _, goPath := range paths {
			for solvePackage := externalPackage; solvePackage != "."; solvePackage = path.Dir(solvePackage) {
				checkGitConfigPath := filepath.Join(goPath, "src", solvePackage, ".git/config")
				if utils.FileExists(checkGitConfigPath) {
					foundGitPath = true
					solvedPackages = append(solvedPackages, solvePackage)
					origin, err := GetGitOrigin(checkGitConfigPath)
					if err != nil {
						logs.Errorf("Failed to parse origin from path %s, the error is %v", checkGitConfigPath, err)
						return depItems, err
					}
					depItems = append(depItems, godeps.DepItem{Path: solvePackage, Origin: origin})
					break
				}
			}
			if foundGitPath {
				break
			}
		}
		if !foundGitPath {
			logs.Errorf(GetRedMessage(fmt.Sprintf("Failed to find git path for package: %s", externalPackage)))
		}
	}
	return depItems, nil
}

func GetGitOrigin(gitConfigPath string) (string, error) {
	if !utils.FileExists(gitConfigPath) {
		return "", common.ErrNotExist
	}
	content, err := ioutil.ReadFile(gitConfigPath)
	if err != nil {
		logs.Errorf("Failed to read file from %s, the error is %v", gitConfigPath, err)
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "url = ") {
			return line[6:], nil
		}
	}
	return "", common.ErrNotExist
}

func GetRedMessage(message string) string {
	return fmt.Sprintf("\x1b[31;1m%s\x1b[0m", message)
}
