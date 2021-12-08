package config

import (
	logger "badcoin/src/helper/logger"
	"path/filepath"

	viper "github.com/spf13/viper"
)

// var (
// 	Configs Configurations
// )

// InitConfig load and marshal config file
func Init(configFile string) (Configs *Configurations, err error){

	// Set undefined variables
	configFilePath := configFile
	if configFilePath == "" {
		configFilePath, _ = filepath.Abs("../../config")
	}
	
	// Set the file name of the configurations file
	viper.SetConfigName("config")

	// Set the path to look for the configurations file
	viper.AddConfigPath(configFilePath)

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetConfigType("yml")

	if err = viper.ReadInConfig(); err != nil {
		logger.Error("Error reading config file, ", err)
		return nil,err
	}

	viper.SetDefault("ConfigFile", configFilePath)

	Configs = new(Configurations)
	err = viper.Unmarshal(Configs)
	if err != nil {
		logger.Error("Unable to decode into struct, ", err)
		return nil,err
	}

	return Configs,nil
}
