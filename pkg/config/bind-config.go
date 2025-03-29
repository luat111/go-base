package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

func GetAllKeys(config any, parts ...string) []string {
	var keys []string

	cnfValue := reflect.ValueOf(config)
	if cnfValue.Kind() == reflect.Ptr {
		cnfValue = cnfValue.Elem()
	}

	for i := range cnfValue.NumField() {
		fieldValue := cnfValue.Field(i)
		fieldType := cnfValue.Type().Field(i)
		fieldTag, ok := fieldType.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.Struct:
			keys = append(keys, GetAllKeys(fieldValue.Interface(), append(parts, fieldTag)...)...)
		default:
			keys = append(keys, strings.Join(append(parts, fieldTag), "."))
		}
	}

	return keys
}

func BindEnv(itf any) {
	for _, key := range GetAllKeys(itf) {
		splitKey := strings.Split(key, ".")
		envKey := splitKey[len(splitKey)-1]

		viper.BindEnv(envKey)
		viper.Set(key, viper.Get(envKey))
	}

	if err := viper.Unmarshal(itf); err != nil {
		fmt.Println(err)
	}
}

func LoadConfig[AppConfig any](path string) (config AppConfig, err error) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	configPath := strings.Join([]string{wd, path, "/.env"}, "")

	viper.SetConfigFile(configPath)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	BindEnv(&config)

	return
}
