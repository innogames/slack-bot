package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

// envPrefix is used to overwrite config values via environment, e.g. BOT_SLACK_TOKEN for slack.token
const envPrefix = "BOT"

// Load all yaml config from a directory or a single .yaml file.
// All map keys keep their case (e.g. Jenkins job names) and map entries without a value are kept as well.
func Load(configFile string) (Config, error) {
	cfg := DefaultConfig

	// workaround to take all keys from struct available, so they can be overwritten via environment
	defaultYaml, _ := yaml.Marshal(DefaultConfig)
	raw, _ := parseYaml(defaultYaml)
	cfg.raw = raw

	fileInfo, err := os.Stat(configFile)
	if err != nil {
		// no file/directory
		return cfg, err
	}

	files := []string{configFile}
	if fileInfo.IsDir() {
		// read all files in a directory
		files, err = filepath.Glob(configFile + "/*.yaml")
		if err != nil {
			return cfg, err
		}
	}

	for _, file := range files {
		fileConfig, err := loadFile(file)
		if err != nil {
			return cfg, err
		}
		mergeMaps(raw, fileConfig)
	}

	applyEnv(raw, nil)

	if err := decode(raw, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// loadFile reads a single yaml file
func loadFile(configFile string) (map[string]any, error) {
	content, err := os.ReadFile(configFile) // #nosec
	if err != nil {
		return nil, err
	}

	m, err := parseYaml(content)
	if err != nil {
		return nil, fmt.Errorf("while parsing config %s: %w", configFile, err)
	}

	return m, nil
}

func parseYaml(content []byte) (map[string]any, error) {
	m := map[string]any{}
	if err := yaml.Unmarshal(content, &m); err != nil {
		return nil, err
	}

	return normalize(m).(map[string]any), nil
}

// normalize converts all nested map[any]any (e.g. from numeric yaml keys) into map[string]any
func normalize(value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, val := range v {
			v[key] = normalize(val)
		}
		return v
	case map[any]any:
		m := make(map[string]any, len(v))
		for key, val := range v {
			m[fmt.Sprint(key)] = normalize(val)
		}
		return m
	case []any:
		for i, val := range v {
			v[i] = normalize(val)
		}
		return v
	default:
		return v
	}
}

// mergeMaps merges src deeply into dst: nested maps get merged, all other values (including lists) are replaced
func mergeMaps(dst map[string]any, src map[string]any) {
	for key, srcVal := range src {
		srcMap, srcIsMap := srcVal.(map[string]any)
		dstMap, dstIsMap := dst[key].(map[string]any)
		if srcIsMap && dstIsMap {
			mergeMaps(dstMap, srcMap)
			continue
		}
		dst[key] = srcVal
	}
}

// applyEnv overwrites all known leaf values with the matching environment variable, e.g. BOT_JENKINS_HOST.
// An empty environment variable also overwrites the value.
func applyEnv(m map[string]any, path []string) {
	for key, val := range m {
		currentPath := append(slices.Clone(path), key)
		if subMap, ok := val.(map[string]any); ok {
			applyEnv(subMap, currentPath)
			continue
		}

		envName := envPrefix + "_" + strings.ToUpper(strings.ReplaceAll(strings.Join(currentPath, "_"), ".", "_"))
		if envValue, ok := os.LookupEnv(envName); ok {
			m[key] = envValue
		}
	}
}

// decode the raw config map into the given struct, with support of time.Duration values and comma separated lists
func decode(input any, output any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			stringToSliceHookFunc(","),
		),
		WeaklyTypedInput: true,
		Result:           output,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// stringToSliceHookFunc splits a string (e.g. from environment) into a slice, an empty string results in an empty slice
func stringToSliceHookFunc(sep string) mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String || to.Kind() != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		return strings.Split(raw, sep), nil
	}
}
