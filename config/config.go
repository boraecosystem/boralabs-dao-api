package config

import (
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
	"runtime"
)

var C *viper.Viper

func init() {
	C = viper.New()
	C.AutomaticEnv()
	C.SetConfigType("yaml")
	C.SetConfigName("app")      // name of config file (without extension)
	C.AddConfigPath(basePath()) // optionally look for config in the working directory
	err := C.ReadInConfig()     // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
}

func basePath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(b)
}
