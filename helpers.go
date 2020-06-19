package main

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/tycale/goubus"
)

func ConfigLoad(file string){

	viper.SetConfigName(file) // set config
	viper.SetConfigType("yaml") // as yaml
	viper.AddConfigPath(".") // look directly where executed

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

}

func ConfigUbusAuth(ubus *goubus.Ubus){

	username := viper.GetString("ubus_user")
	if username == "" {
		panic(fmt.Errorf("Fatal error config file does not contain ubus_user\n"))
	}

	password := viper.GetString("ubus_password")
	if password == "" {
		panic(fmt.Errorf("Fatal error config file does not contain ubus_user\n"))
	}

	ubus.Username = username
	ubus.Password = password
}

