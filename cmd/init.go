package cmd

import (
	"encoding/json"
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
		Flags: []cli.Flag{
			stringFlag("packagesetting, ps", "", "The path of package setting json file"),
		},
	}
)

func runInit(c *cli.Context) error {
	jsonPath := c.String("packagesetting")
	pi := PackageInitializer{}
	if err := pi.SetPath(jsonPath); err != nil {
		return err
	}
	return pi.doInit()
}

type PackageSetting struct {
	IgnorePackages []godeps.DepItem `json:"ignorePackages"`
	CustomPackages []godeps.DepItem `json:"customPackages"`
}

type PackageInitializer struct {
	Path           string
	packageSetting PackageSetting
}

func (pi *PackageInitializer) SetPath(path string) error {
	if path != "" {
		if !utils.FileExists(path) {
			logs.Errorf("The custom packages json file %s does not exists", path)
			return common.ErrNotExist
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			logs.Errorf("Failed to read content from file %s, the error is %#v", path, err)
			return err
		}
		err = json.Unmarshal(content, &pi.packageSetting)
		if err != nil {
			logs.Errorf("Failed to unmarshal package settings, the error is %#v", err)
			return err
		}
	}
	return nil
}

func (pi *PackageInitializer) doInit() error {
	systemPackages, err := pi.GetSystemPackages()
	if err != nil {
		logs.Errorf("Failed to get system packages, the error is %v", err)
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		logs.Errorf("Failed to get wd, the error is %v", err)
		return err
	}
	externalPackages, err := pi.GetExternalPackages(wd, systemPackages)
	if err != nil {
		logs.Errorf("Failed to get external packages, the error is %v", err)
	}
	depItems, err := pi.GetDepItems(externalPackages)
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

func (pi *PackageInitializer) GetSystemPackages() ([]string, error) {
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

func (pi *PackageInitializer) GetExternalPackages(wd string, systemPackages []string) ([]string, error) {
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

func (pi *PackageInitializer) GetDepItems(externalPackages []string) ([]godeps.DepItem, error) {
	depItems := []godeps.DepItem{}
	envGoPath := os.Getenv("GOPATH")
	paths := strings.Split(envGoPath, fmt.Sprintf("%c", os.PathListSeparator))
	solvedPackages := []string{}
	for _, externalPackage := range externalPackages {
		if pi.IsIgnorePackage(externalPackage) {
			logs.Debugf("The package %s is in ignore list", externalPackage)
			continue
		}
		found := false
		for index := 0; index < len(solvedPackages) && !found; index++ {
			if strings.HasPrefix(externalPackage, solvedPackages[index]) {
				found = true
			}
		}
		if found {
			continue
		}
		customDepItem, err := pi.GetCustomDepItem(externalPackage)
		if err != nil {
			if err != common.ErrNotExist {
				logs.Errorf("Failed to get custom dep item, the error is %#v", err)
				return depItems, err
			}
		} else {
			depItems = append(depItems, customDepItem)
			solvedPackages = append(solvedPackages, customDepItem.Path)
			continue
		}
		foundGitPath := false
		for _, goPath := range paths {
			for solvePackage := externalPackage; solvePackage != "."; solvePackage = path.Dir(solvePackage) {
				checkGitConfigPath := filepath.Join(goPath, "src", solvePackage, ".git/config")
				if utils.FileExists(checkGitConfigPath) {
					foundGitPath = true
					solvedPackages = append(solvedPackages, solvePackage)
					origin, err := pi.GetGitOrigin(checkGitConfigPath)
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

func (pi *PackageInitializer) IsIgnorePackage(externalPackage string) bool {
	for _, item := range pi.packageSetting.IgnorePackages {
		if strings.HasPrefix(externalPackage, item.Path) {
			return true
		}
	}
	return false
}

func (pi *PackageInitializer) GetCustomDepItem(externalPackage string) (godeps.DepItem, error) {
	if len(pi.packageSetting.CustomPackages) > 0 {
		for _, item := range pi.packageSetting.CustomPackages {
			if strings.HasPrefix(externalPackage, item.Path) {
				logs.Debugf("Find custom dep item %#v", item)
				return item, nil
			}
		}
	}
	return godeps.DepItem{}, common.ErrNotExist
}

func (pi *PackageInitializer) GetGitOrigin(gitConfigPath string) (string, error) {
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
