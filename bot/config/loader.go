package config

import (
	"bytes"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

// Load loads config yaml file(s) inside a directory or a single .yaml file
func Load(configFile string) (Config, error) {
	v := viper.NewWithOptions(viper.KeyDelimiter("#"))

	v.SetConfigType("yaml")
	v.AllowEmptyEnv(true)
	v.SetEnvPrefix("BOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	cfg := defaultConfig

	// workaround to take all keys from struct available
	defaultYaml, _ := yaml.Marshal(defaultConfig)
	v.ReadConfig(bytes.NewBuffer(defaultYaml))

	fileInfo, err := os.Stat(configFile)
	if err != nil {
		// no file/directory
		return cfg, err
	} else if fileInfo.IsDir() {
		// read all files in a directory
		files, err := filepath.Glob(configFile + "/*.yaml")
		if err != nil {
			return cfg, err
		}
		for _, file := range files {
			err := loadFile(v, file)
			if err != nil {
				return cfg, err
			}
		}
	} else {
		err := loadFile(v, configFile)
		if err != nil {
			return cfg, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func loadFile(v *viper.Viper, configFile string) error {
	// read a single yaml file
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}

	defer file.Close()

	return v.MergeConfig(file)
}
