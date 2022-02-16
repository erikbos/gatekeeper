package config

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Load attempts to load the config and fill it with ENV values
func Load(filename string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(filename)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper read config")
	}

	return v, nil
}
