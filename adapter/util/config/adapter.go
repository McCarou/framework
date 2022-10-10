package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"syscall"

	"golang.org/x/exp/slices"

	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

const DefaultJsonConfigPath string = "config.json"
const DefaultConfigEnvPrefix string = "RADIAN"
const DefaultTagConfigName = "config"
const DefaultTagConfigRequiredName = "required"

type ConfigAdapter struct {
	*adapter.BaseAdapter

	config map[string]any
}

func NewConfigAdapter(name string) *ConfigAdapter {
	return &ConfigAdapter{BaseAdapter: adapter.NewBaseAdapter(name), config: make(map[string]any)}
}

func (a *ConfigAdapter) LoadFromJson(cfgStr []byte) error {
	return json.Unmarshal(cfgStr, &a.config)
}

func (a *ConfigAdapter) LoadFromFileJson(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		filePath = DefaultJsonConfigPath
	}

	cnf := make(map[string]any)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Warningf("Cannot read file %s - %s", filePath, err)
		return err
	}

	err = json.Unmarshal(data, &cnf)

	if err != nil {
		logrus.Errorf("Cannot apply config from file %s - %s", filePath, err)
	}

	a.config = cnf // TODO: remove and make union with previous configuration

	return err
}

func (a *ConfigAdapter) LoadFromEnv(prefix string) error {
	return a.loadMap(syscall.Environ(), prefix, "=", "_")
}

func (a *ConfigAdapter) LoadFromArgs(prefix string) error {
	if len(os.Args) < 2 {
		return nil
	}

	args := os.Args[1:]

	return a.loadMap(args, prefix, "=", "-")
}

func (a *ConfigAdapter) loadMap(data []string, prefix, delimKeyValue, delimParams string) error {
	for _, arg := range data {
		if !strings.HasPrefix(arg, prefix) {
			continue
		}

		removeChars := len(prefix) + len(delimParams)

		kv := strings.Split(arg[removeChars:], delimKeyValue)
		keys := strings.Split(kv[0], delimParams)
		if err := a.SetValue(keys, kv[1]); err != nil {
			return err
		}
	}

	return nil
}

func (a *ConfigAdapter) GetAdapter(path []string) (*ConfigAdapter, error) {
	ac := NewConfigAdapter(a.GetName())

	m, err := a.GetValue(path)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(m).Kind() != reflect.TypeOf(map[string]any{}).Kind() {
		return nil, errors.New("invalid config")
	}

	ac.config = maps.Clone(m.(map[string]any))

	return ac, nil
}

func (a *ConfigAdapter) GetValueOrDefault(path []string, defaultValue any) any {
	result, err := a.GetValue(path)

	if err != nil {
		return defaultValue
	}

	return result
}

func (a *ConfigAdapter) GetValue(path []string) (any, error) {
	if len(path) == 0 {
		return a.config, nil
	}

	n := a.config

	for idx, value := range path {
		if idx == len(path)-1 {
			break
		}

		if _, ok := a.config[value]; !ok {
			return nil, errors.New("path not valid")
		}

		if reflect.TypeOf(n).Kind() != reflect.TypeOf(n[value]).Kind() {
			return nil, errors.New("wrong configuration")
		}

		n = n[value].(map[string]any)
	}

	result, ok := n[path[len(path)-1]]

	if !ok {
		return nil, errors.New("value not found")
	}

	return result, nil
}

func (a *ConfigAdapter) SetValue(path []string, val any) error {
	if len(path) == 0 {
		return nil
	}

	n := a.config

	for idx, value := range path {
		if idx == len(path)-1 {
			break
		}

		if _, ok := a.config[value]; !ok {
			n[value] = make(map[string]any)
		}

		if reflect.TypeOf(n).Kind() != reflect.TypeOf(n[value]).Kind() {
			return errors.New("wrong configuration")
		}

		n = n[value].(map[string]any)
	}

	n[path[len(path)-1]] = val

	return nil
}

func (a *ConfigAdapter) Unmarshal(v interface{}) error {
	return a.unmarshalFromMap(a.config, v)
}

func (a *ConfigAdapter) UnmarshalPath(path []string, v interface{}) error {
	m, err := a.GetValue(path)
	if err != nil {
		return err
	}

	if reflect.TypeOf(m).Kind() != reflect.TypeOf(map[string]any{}).Kind() {
		return errors.New("invalid config")
	}

	return a.unmarshalFromMap(m.(map[string]any), v)
}

func (a *ConfigAdapter) unmarshalFromMap(m map[string]any, v interface{}) error {
	if m == nil {
		return errors.New("empty config")
	}
	if v == nil {
		return errors.New("empty structure")
	}

	valueOfV := reflect.ValueOf(v).Elem()
	typeOfV := valueOfV.Type()

	for i := 0; i < typeOfV.NumField(); i++ {
		fieldTagString, ok := typeOfV.Field(i).Tag.Lookup(DefaultTagConfigName)

		if !ok || fieldTagString == "" {
			logrus.Debugf("Field '%s' has no tag = '%s'. Skip", typeOfV.Field(i).Name, DefaultTagConfigName)
			continue
		}

		fieldTags := strings.Split(fieldTagString, ",")

		fieldEnvVal, ok := m[fieldTags[0]]

		if !ok {
			logrus.Debugf("Field '%s' is not in config. Skip", fieldTags[0])

			if slices.Contains(fieldTags, DefaultTagConfigRequiredName) {
				return fmt.Errorf("field '%s' is required", typeOfV.Field(i).Name)
			}

			continue
		}

		if valueOfV.Field(i).CanSet() {
			val := reflect.ValueOf(fieldEnvVal)
			if valueOfV.Field(i).Kind() == reflect.Slice {
				for _, sliceVal := range fieldEnvVal.([]interface{}) {
					valueOfV.Field(i).Set(reflect.Append(valueOfV.Field(i), reflect.ValueOf(sliceVal)))
				}
				continue
			}
			valueOfV.Field(i).Set(val.Convert(valueOfV.Field(i).Type()))
		} else {
			return fmt.Errorf("field '%s' cannot be set", typeOfV.Field(i).Name)
		}
	}

	return nil
}

func (a *ConfigAdapter) Setup() (err error) {
	return nil
}

func (a *ConfigAdapter) Close() error {
	for k := range a.config {
		delete(a.config, k)
	}

	return nil
}
