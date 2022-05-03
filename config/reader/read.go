package reader

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	viperDefaultDelimiter = "."
	defaultTagName        = "default"
	squashTagValue        = ",squash"
	mapStructureTagName   = "mapstructure"
)

// Read reads configs to use during run
func Read(config interface{}, opts ...viper.DecoderConfigOption) error {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(viperDefaultDelimiter, "_")) // replace default viper delimiter for env vars
	v.AutomaticEnv()
	v.SetTypeByDefaultValue(true)
	err := setDefaults("", v, reflect.StructField{}, reflect.ValueOf(config).Elem())
	if err != nil {
		return errors.WithMessage(err, "failed to apply defaults")
	}
	err = v.Unmarshal(config, opts...)
	if err != nil {
		return errors.WithMessage(err, "failed to parse configuration")
	}
	return nil
}

// setDefaults sets default values for struct fields based using value from default tag
// nolint:gocyclo
func setDefaults(parentName string, vip *viper.Viper, t reflect.StructField, v reflect.Value) error {
	if v.Kind() == reflect.Struct {
		value, ok := t.Tag.Lookup(mapStructureTagName)
		if ok && value != squashTagValue {
			if parentName != "" {
				parentName += viperDefaultDelimiter
			}
			parentName += strings.ToUpper(value)
		}
		for i := 0; i < v.NumField(); i++ {
			if err := setDefaults(parentName, vip, v.Type().Field(i), v.Field(i)); err != nil {
				return err
			}
		}
		return nil
	}
	value, _ := t.Tag.Lookup(defaultTagName)
	fieldName, ok := t.Tag.Lookup(mapStructureTagName)
	if ok && fieldName != squashTagValue {
		if parentName != "" {
			fieldName = parentName + viperDefaultDelimiter + strings.ToUpper(fieldName)
		}
		vip.SetDefault(strings.ToUpper(fieldName), value)
	}
	return nil
}
