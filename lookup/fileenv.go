package lookup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/utils"
)

type FileEnvLookup struct {
	parent    config.EnvironmentLookup
	variables map[string]string
}

func NewFileEnvLookup(file string, parent config.EnvironmentLookup) (*FileEnvLookup, error) {
	variables := map[string]string{}

	if file != "" {
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var data map[string]interface{}

		if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
			yaml.Unmarshal(contents, &data)
		} else if strings.HasSuffix(file, ".json") {
			json.Unmarshal(contents, &data)
		} else {
			return nil, errors.New("Environment variable file must be suffixed with '.yml', '.yaml', or '.json' and in those respective formats")
		}

		for k, v := range data {
			if stringValue, ok := v.(string); ok {
				variables[k] = stringValue
			} else if intValue, ok := v.(int); ok {
				variables[k] = fmt.Sprintf("%v", intValue)
			} else if int64Value, ok := v.(int64); ok {
				variables[k] = fmt.Sprintf("%v", int64Value)
			} else if float32Value, ok := v.(float32); ok {
				variables[k] = fmt.Sprintf("%v", float32Value)
			} else if float64Value, ok := v.(float64); ok {
				variables[k] = fmt.Sprintf("%v", float64Value)
			} else if boolValue, ok := v.(bool); ok {
				variables[k] = strconv.FormatBool(boolValue)
			} else {
				return nil, fmt.Errorf("Environment variables must be of type string, bool, or int. Key %s is of type %T", k, v)
			}
		}
	}

	logrus.Debugf("Environment Context from file %s: %v", file, variables)
	return &FileEnvLookup{
		parent:    parent,
		variables: variables,
	}, nil
}

func (f *FileEnvLookup) Lookup(key string, config *config.ServiceConfig) []string {
	if v, ok := f.variables[key]; ok {
		return []string{fmt.Sprintf("%s=%s", key, v)}
	}

	if f.parent == nil {
		return nil
	}

	return f.parent.Lookup(key, config)
}

func (f *FileEnvLookup) Variables() map[string]string {
	return utils.MapUnion(f.variables, f.parent.Variables())
}
