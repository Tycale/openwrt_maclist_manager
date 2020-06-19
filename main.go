package main

import (
	"fmt"
	"log"
	"flag"

	"github.com/tycale/goubus"
)

func main() {
	var configF string

	flag.StringVar(&configF, "config", "settings.yml", "yaml file used for settings")

  flag.Parse()


	ubus := goubus.Ubus{
	  URL:      "http://192.168.2.1/ubus",
	}

	ConfigLoad(configF)
	ConfigUbusAuth(&ubus)

	call, err := ubus.AuthLogin()
	if err != nil {
	  log.Fatal(err)
	}
	fmt.Println(call)

}


