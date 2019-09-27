package config

import (
	"fmt"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// LoadPattern loads config yaml file(s) by a glob pattern
func LoadPattern(pattern string) (Config, error) {
	cfg := defaultConfig

	fileNames, err := filepath.Glob(pattern)
	if err != nil {
		return cfg, err
	}

	if len(fileNames) == 0 {
		return cfg, fmt.Errorf("no config file found: %s", pattern)
	}

	for _, fileName := range fileNames {
		newCfg, err := loadConfig(fileName)
		if err != nil {
			return cfg, err
		}

		if err := mergo.Merge(&cfg, newCfg, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}

// loadConfig loads a single yaml config file
func loadConfig(filename string) (Config, error) {
	cfg := Config{}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return cfg, errors.Errorf("failed to load config file from %s: %s", filename, err)
	}

	if err := yaml.UnmarshalStrict(content, &cfg); err != nil {
		return cfg, errors.Errorf("failed to parse configuration file: %s", err)
	}

	return cfg, nil
}
