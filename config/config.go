package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Load(file string) {

	viper.SetConfigName(file)   // set config
	viper.SetConfigType("yaml") // as yaml
	viper.AddConfigPath(".")    // look directly where executed

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

}

func GetUser() string {
	username := viper.GetString("ubus_user")
	if username == "" {
		panic(fmt.Errorf("Fatal error config file does not contain ubus_user\n"))
	}
	return username
}

func GetPass() string {
	password := viper.GetString("ubus_password")
	if password == "" {
		panic(fmt.Errorf("Fatal error config file does not contain ubus_user\n"))
	}
	return password
}

func GetStringSlice(s string) []string {
	res := viper.GetStringSlice(s)
	if len(res) == 0 {
		panic(fmt.Errorf("Fatal error config file does not contain '" + s + "' or is empty\n"))
	}

	return res
}
