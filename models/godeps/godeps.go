package godeps

import (
	"encoding/json"
	"io/ioutil"

	"fmt"

	"github.com/swanwish/go-common/logs"
	"github.com/swanwish/go-common/utils"
	"github.com/swanwish/godeps/common"
)

type GoDeps struct {
	Deps []*DepItem
}

func NewGoDeps() (*GoDeps, error) {
	goDeps := GoDeps{}
	err := goDeps.LoadGoDeps()
	return &goDeps, err
}

func (goDeps *GoDeps) LoadGoDeps() error {
	if utils.FileExists(common.DEFAULT_CONFIGURATION_FILE_NAME) {
		content, err := ioutil.ReadFile(common.DEFAULT_CONFIGURATION_FILE_NAME)
		if err != nil {
			logs.Errorf("Failed to read configuration file, the error is %v", err)
			return err
		}
		err = json.Unmarshal(content, &goDeps.Deps)
	}
	return nil
}

func (goDeps *GoDeps) AddItem(path, origin string) error {
	if path == "" || origin == "" {
		logs.Errorf("Failed to add item, the path or origin is empty")
		return common.ErrInvalidParameter
	}
	for _, item := range goDeps.Deps {
		if item.Path == path {
			logs.Warnf("The path %s already exists", path)
			return common.ErrAlreadyExist
		}
		if item.Origin == origin {
			logs.Warnf("The origin %s already exists", origin)
			return common.ErrAlreadyExist
		}
	}
	goDeps.Deps = append(goDeps.Deps, &DepItem{Path: path, Origin: origin})
	return nil
}

func (goDeps *GoDeps) DeleteItem(path string) error {
	if path == "" {
		logs.Errorf("The path is not specified")
		return common.ErrInvalidParameter
	}
	itemIndex := -1
	for index, item := range goDeps.Deps {
		if item.Path == path {
			itemIndex = index
			break
		}
	}
	if itemIndex != -1 {
		goDeps.Deps = append(goDeps.Deps[:itemIndex], goDeps.Deps[itemIndex+1:]...)
	}
	return nil
}

func (goDeps *GoDeps) UpdateItem(path, origin string) error {
	if path == "" || origin == "" {
		logs.Errorf("The path or origin is not specified")
		return common.ErrInvalidParameter
	}
	for _, item := range goDeps.Deps {
		if item.Path == path {
			item.Origin = origin
		}
	}
	return nil
}

func (goDeps *GoDeps) Save() error {
	jsonContent, err := json.MarshalIndent(goDeps.Deps, "", "    ")
	if err != nil {
		logs.Errorf("Failed to marshal deps, the error is %v", err)
		return err
	}
	fmt.Printf("The content of the godeps.json is:\n%s\n", string(jsonContent))
	return utils.SaveFile(common.DEFAULT_CONFIGURATION_FILE_NAME, jsonContent)
}
