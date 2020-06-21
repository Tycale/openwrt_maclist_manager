package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"

	"./config"
	"./goubus"
)

type conf struct {
	Devices     []string
	SSIDs       []string
	AllowedMacs []string
	User        string
	Password    string
}

func filterIfaces(ubus goubus.Ubus, uwd goubus.UbusWirelessDevice, ssids []string) []string {
	selectedIfaces := []string{}

	for _, s := range uwd.Devices {
		info, err := ubus.WirelessInfo(0, s)

		if err != nil {
			log.Fatal(err)
		}

		for _, ssid := range ssids {
			if ssid == info.SSID {
				selectedIfaces = append(selectedIfaces, s)
			}
		}
	}

	return selectedIfaces
}

func loadConfig() conf {
	var configF string
	c := conf{}

	flag.StringVar(&configF, "config", "settings.yml", "yaml file used for settings")
	flag.Parse()

	config.Load(configF)
	c.SSIDs = config.GetStringSlice("ssids")
	c.AllowedMacs = config.GetStringSlice("allowed_macs")
	c.Devices = config.GetStringSlice("devices")
	c.User = config.GetUser()
	c.Password = config.GetPass()

	return c
}

func setUbusAuth(ubus *goubus.Ubus, c conf) {
	ubus.Username = c.User
	ubus.Password = c.Password
}

func main() {

	c := loadConfig()

	for _, d := range c.Devices {
		log.Println("------ Updating " + d + " ------")
		updateMacs(d, c)
	}
}

func checkWifisExist(ubus goubus.Ubus, c conf) int {
	// Find setup wifis
	d, err := ubus.WirelessDevices(0)
	if err != nil {
		log.Fatal(err)
	}

	// Filter out wifi
	numWifiFound := len(d.Devices)
	selIfaces := filterIfaces(ubus, d, c.SSIDs)
	numFilteredWifi := len(selIfaces)

	// log Results
	log.Printf("Found " + strconv.Itoa(numFilteredWifi) + " out of " + strconv.Itoa(numWifiFound) + " WiFi corresponding to your filter via ubus.\n")

	return numFilteredWifi
}

func findUciWifiSection(ubus goubus.Ubus, c conf, numWifis int) []int {

	selIfaces := []int{}

	for i := 0; i < numWifis; i++ {

		// 1/2. Marshal json request
		var uciReq goubus.UbusUciRequest
		err := json.Unmarshal([]byte(`
		{
			"config": "wireless",
			"section": "@wifi-iface[`+strconv.Itoa(i)+`]",
			"option": "ssid"
		}
		`), &uciReq)

		if err != nil {
			log.Fatal(err)
		}

		// 2/2. Execute
		res, err := ubus.UciGetConfig(0, uciReq)

		for i, s := range c.SSIDs {
			if s == res.Value {
				selIfaces = append(selIfaces, i)
				log.Println("Found '" + s + "' SSID in UCI as 'wireless.@wifi-iface[" + strconv.Itoa(i) + "]'.")
			}
		}
	}

	return selIfaces
}

func updateMacs(device string, c conf) {

	ubus := goubus.Ubus{
		URL: "http://" + device + "/ubus",
	}

	setUbusAuth(&ubus, c)

	// Auth
	_, err := ubus.AuthLogin()
	if err != nil {
		log.Fatal(err)
	}

	numWifis := checkWifisExist(ubus, c)

	if numWifis == 0 {
		log.Fatal("Could not find any wifi to configure on this device.")
	}

	uciIfaces := findUciWifiSection(ubus, c, numWifis)

	if len(uciIfaces) != numWifis {
		log.Fatal("UCI does not reflex ubus infos, abort.")
	}

	for _, uIface := range uciIfaces {
		toDelMacs, toAddMacs := listDelAddMacs(ubus, c, uIface)
		log.Println(toDelMacs)
		log.Println(toAddMacs)
	}

}

func listDelAddMacs(ubus goubus.Ubus, c conf, iface int) ([]string, []string) {
	var toDel []string
	var toAdd []string

	// uci request to current maclist
	// 1/2. Marshal json request
	var uciReq goubus.UbusUciRequest
	err := json.Unmarshal([]byte(`
	{
		"config": "wireless",
		"section": "@wifi-iface[`+strconv.Itoa(iface)+`]",
		"option": "maclist"
	}
	`), &uciReq)

	if err != nil {
		log.Fatal(err)
	}

	// 2/2. Execute
	macList, err := ubus.UciGetConfig(0, uciReq)

	if err != nil {
		if err.Error() == "Empty response" {
			log.Println("No maclist at the moment.")
			return []string{}, c.AllowedMacs
		} else {
			log.Fatal(err)
		}
	}

	// Find existing macs that should be deleted
	for _, m := range macList.Value.([]interface{}) {
		mac := fmt.Sprintf("%v", m)
		keep := false
		for _, wm := range c.AllowedMacs {
			if wm == mac {
				keep = true
			}
		}
		if !keep {
			toDel = append(toDel, mac)
		}
	}

	// Find existing macs that should be added
	for _, wm := range c.AllowedMacs {
		exist := false

		for _, m := range macList.Value.([]interface{}) {
			mac := fmt.Sprintf("%v", m)
			if wm == mac {
				exist = true
			}
		}
		if !exist {
			toAdd = append(toAdd, wm)
		}
	}

	return toDel, toAdd
}
