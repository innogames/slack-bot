package config

import (
	"bytes"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// Load loads config yaml file(s) inside a directory or a single .yaml file
func Load(configFile string) (Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")
	v.AllowEmptyEnv(true)
	v.SetEnvPrefix("BOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	cfg := defaultConfig

	// workaround to ake all keys from struct available
	defaultYaml, _ := yaml.Marshal(defaultConfig)
	v.ReadConfig(bytes.NewBuffer(defaultYaml))

	fileInfo, err := os.Stat(configFile)
	if err != nil {
		// no file/directory
		return cfg, err
	} else if fileInfo.IsDir() {
		// read all files in a directory
		v.AddConfigPath(configFile)
		err := v.MergeInConfig()
		if err != nil {
			return cfg, err
		}
	} else {
		// read a single yaml file
		file, err := os.Open(configFile)

		if err != nil {
			return cfg, err
		}

		defer file.Close()
		err = v.MergeConfig(file)
		if err != nil {
			return cfg, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
