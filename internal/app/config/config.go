package config

import (
	"log"

	"github.com/go-viper/encoding/ini"
	"github.com/spf13/viper"
)

var (
	viperCfg *viper.Viper
)

func init() {
	viper.GetViper().AllKeys()
	// Configure Viper to read the config file
	codecRegistry := viper.NewCodecRegistry()
	codecRegistry.RegisterCodec("ini", ini.Codec{})

	viperCfg = viper.NewWithOptions(
		viper.WithCodecRegistry(codecRegistry),
	)

	viperCfg.SetConfigName("config")
	viperCfg.SetConfigType("ini")
	viperCfg.AddConfigPath(".")
	viperCfg.AddConfigPath("$HOME/.cosmos")
	if err := viperCfg.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}

func GetAppConfig() *viper.Viper {
	return viperCfg
}
