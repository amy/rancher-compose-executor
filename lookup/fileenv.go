package lookup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			t := strings.TrimSpace(scanner.Text())
			parts := strings.SplitN(t, "=", 2)
			if len(parts) == 1 {
				variables[parts[0]] = ""
			} else {
				stringCharacter := ""
				if strings.HasPrefix(parts[1], `"`) {
					stringCharacter = `"`
				} else if strings.HasPrefix(parts[1], `'`) {
					stringCharacter = `'`
				} else if strings.HasPrefix(parts[1], "`") {
					stringCharacter = "`"
				}

				if strings.HasPrefix(parts[1], stringCharacter) && strings.HasSuffix(parts[1], stringCharacter) && len(parts[1]) > 1 || stringCharacter == "" {
					variables[parts[0]] = strings.Trim(parts[1], stringCharacter)
				} else {
					for scanner.Scan() {
						parts[1] = parts[1] + "\n" + strings.TrimSpace(scanner.Text())
						if strings.HasSuffix(parts[1], stringCharacter) {
							variables[parts[0]] = strings.Trim(parts[1], stringCharacter)
							break
						}
					}
				}
			}
		}

		if scanner.Err() != nil {
			return nil, scanner.Err()
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
