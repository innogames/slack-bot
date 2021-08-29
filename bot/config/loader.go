package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/brainexe/viper"
	"gopkg.in/yaml.v3"
)

// Load all yaml config from a directory or a single .yaml file
func Load(configFile string) (Config, error) {
	// don't use '.' or '_' etc as delimiter, as it will block having this chars as map keys
	keyDelimiter := "§"
	v := viper.NewWithOptions(viper.KeyDelimiter(keyDelimiter), viper.KeyPreserveCase())

	v.SetConfigType("yaml")
	v.AllowEmptyEnv(true)
	v.SetEnvPrefix("BOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", keyDelimiter, "_"))
	v.AutomaticEnv()

	cfg := DefaultConfig

	// workaround to take all keys from struct available
	defaultYaml, _ := yaml.Marshal(DefaultConfig)
	_ = v.ReadConfig(bytes.NewBuffer(defaultYaml))

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
	file, err := os.Open(configFile) // #nosec
	if err != nil {
		return err
	}

	defer file.Close()

	return v.MergeConfig(file)
}
