package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/exp/slices"

	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

const DefaultJsonConfigPath string = "config.json"
const DefaultConfigEnvPrefix string = "RADIAN"
const TagConfigName = "config"
const TagConfigRequiredName = "required"

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

	data, err := os.ReadFile(filePath)
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

		param := arg[len(prefix)+len(delimParams):]

		splitIndex := strings.Index(param, delimKeyValue)

		kv := param[splitIndex+len(delimKeyValue):]
		keys := strings.Split(param[:splitIndex], delimParams)

		if err := a.SetValue(keys, kv); err != nil {
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

func (a *ConfigAdapter) GetAdapterOrNil(path []string) *ConfigAdapter {
	adp, err := a.GetAdapter(path)

	if err != nil {
		return nil
	}

	return adp
}

func (a *ConfigAdapter) GetValueOrDefault(path []string, defaultValue any) any {
	result, err := a.GetValue(path)

	if err != nil {
		return defaultValue
	}

	return result
}

func (a *ConfigAdapter) GetStringOrDefault(path []string, defaultValue string) string {
	result, err := a.GetString(path)

	if err != nil {
		return defaultValue
	}

	return result
}

func (a *ConfigAdapter) GetString(path []string) (string, error) {
	result, err := a.GetValue(path)

	if err != nil {
		return "", err
	}

	return fmt.Sprint(result), nil
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

		if _, ok := n[value]; !ok {
			return nil, errors.New("path not valid: " + strings.Join(append(path[:idx], value), ".")) // TODO: make more info in error
		}

		if reflect.TypeOf(n).Kind() != reflect.TypeOf(n[value]).Kind() {
			return nil, errors.New("wrong configuration") // TODO: make more info in error
		}

		n = n[value].(map[string]any)
	}

	result, ok := n[path[len(path)-1]]

	if !ok {
		return nil, errors.New("value not found: " + path[len(path)-1]) // TODO: return name of the value
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

		if _, ok := n[value]; !ok {
			n[value] = make(map[string]any)
		}

		if reflect.TypeOf(n).Kind() != reflect.TypeOf(n[value]).Kind() {
			return errors.New("wrong configuration") // TODO: make more info in error
		}

		n = n[value].(map[string]any)
	}

	n[path[len(path)-1]] = val

	return nil
}

func (a *ConfigAdapter) Unmarshal(destination interface{}, skipRequired bool) error {
	return a.unmarshalFromMap(a.config, destination, skipRequired)
}

func (a *ConfigAdapter) UnmarshalPath(path []string, destination interface{}, skipRequired bool) error {
	m, err := a.GetValue(path)
	if err != nil {
		return err
	}

	if reflect.TypeOf(m).Kind() != reflect.TypeOf(map[string]any{}).Kind() {
		return errors.New("invalid config") // TODO: make more info in error
	}

	return a.unmarshalFromMap(m.(map[string]any), destination, skipRequired)
}

func (a *ConfigAdapter) unmarshalFromMap(source map[string]any, destination interface{}, skipRequired bool) error {
	if source == nil {
		return errors.New("empty config")
	}
	if destination == nil {
		return errors.New("empty structure")
	}

	valueOfDest := reflect.ValueOf(destination).Elem()
	typeOfDest := valueOfDest.Type()

	for i := 0; i < typeOfDest.NumField(); i++ {
		fieldTagString, ok := typeOfDest.Field(i).Tag.Lookup(TagConfigName)

		if !ok || fieldTagString == "" {
			logrus.Debugf("Field '%s' has no tag = '%s'. Skip", typeOfDest.Field(i).Name, TagConfigName)
			continue
		}

		fieldTags := strings.Split(fieldTagString, ",")

		sourceValue, ok := source[fieldTags[0]]

		if !ok {
			logrus.Debugf("Field '%s' is not in config. Skip", fieldTags[0])

			if !skipRequired && slices.Contains(fieldTags, TagConfigRequiredName) {
				return fmt.Errorf("field '%s' is required", typeOfDest.Field(i).Name)
			}

			continue
		}

		if valueOfDest.Field(i).CanSet() {
			if valueOfDest.Field(i).Kind() == reflect.Slice {
				for _, sliceVal := range sourceValue.([]interface{}) {
					valueOfDest.Field(i).Set(reflect.Append(valueOfDest.Field(i), reflect.ValueOf(sliceVal)))
				}
				continue
			}

			sourceReflectValue := reflect.ValueOf(sourceValue)

			if (valueOfDest.Field(i).Kind() == reflect.Int ||
				valueOfDest.Field(i).Kind() == reflect.Int16 ||
				valueOfDest.Field(i).Kind() == reflect.Int32 ||
				valueOfDest.Field(i).Kind() == reflect.Int64 ||
				valueOfDest.Field(i).Kind() == reflect.Int8) &&
				sourceReflectValue.Kind() == reflect.String {
				intVal, err := strconv.ParseInt(sourceReflectValue.String(), 10, int(valueOfDest.Field(i).Type().Size())*8)

				if err != nil {
					return err
				}

				valueOfDest.Field(i).Set(reflect.ValueOf(intVal).Convert(valueOfDest.Field(i).Type()))

				continue
			}

			if (valueOfDest.Field(i).Kind() == reflect.Uint ||
				valueOfDest.Field(i).Kind() == reflect.Uint16 ||
				valueOfDest.Field(i).Kind() == reflect.Uint32 ||
				valueOfDest.Field(i).Kind() == reflect.Uint64 ||
				valueOfDest.Field(i).Kind() == reflect.Uint8) &&
				sourceReflectValue.Kind() == reflect.String {
				uintVal, err := strconv.ParseUint(sourceReflectValue.String(), 10, int(valueOfDest.Field(i).Type().Size())*8)

				if err != nil {
					return err
				}

				valueOfDest.Field(i).Set(reflect.ValueOf(uintVal).Convert(valueOfDest.Field(i).Type()))

				continue
			}

			if valueOfDest.Field(i).Kind() == reflect.Bool &&
				sourceReflectValue.Kind() == reflect.String {
				boolVal, err := strconv.ParseBool(sourceReflectValue.String())

				if err != nil {
					return err
				}

				valueOfDest.Field(i).Set(reflect.ValueOf(boolVal).Convert(valueOfDest.Field(i).Type()))

				continue
			}

			if (valueOfDest.Field(i).Kind() == reflect.Float32 ||
				valueOfDest.Field(i).Kind() == reflect.Float64) &&
				sourceReflectValue.Kind() == reflect.String {
				floatVal, err := strconv.ParseFloat(sourceReflectValue.String(), int(valueOfDest.Field(i).Type().Size())*8)

				if err != nil {
					return err
				}

				valueOfDest.Field(i).Set(reflect.ValueOf(floatVal).Convert(valueOfDest.Field(i).Type()))

				continue
			}

			valueOfDest.Field(i).Set(sourceReflectValue.Convert(valueOfDest.Field(i).Type()))
		} else {
			return fmt.Errorf("field '%s' cannot be set", typeOfDest.Field(i).Name)
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
