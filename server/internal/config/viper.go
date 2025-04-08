package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func (a *AppConfig) mustLoadSpecs(cfgFile *string) *AppConfig {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("pxgrider")

	if cfgFile != nil && *cfgFile != "" {
		viper.SetConfigFile(*cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}
	viper.AutomaticEnv()

	a.Specs.prepareViper()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	if err := a.Specs.loadFromViper(); err != nil {
		panic(err)
	}

	a.Specs.Lock()
	defer a.Specs.Unlock()

	if err := a.Specs.DB.Validate(); err != nil {
		panic(err)
	}

	a.Specs.Version.BuildStamp = BuildStamp
	a.Specs.Version.GitHash = GitHash
	a.Specs.Version.GitVersion = V
	a.Specs.Version.V = Version

	return a
}
