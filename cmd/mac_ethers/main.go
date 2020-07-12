package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"../../goubus"
	"./config"

	"github.com/fatih/color"
)

type conf struct {
	Devices  []string
	User     string
	Password string
	Output   string
}

type etherEntry struct {
	Mac  string
	Name string
}

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

func main() {
	var ethers []etherEntry

	c := loadConfig()

	for _, d := range c.Devices {
		fmt.Println("Retrieve " + d)
		ethers = append(ethers, extractEthersDevice(d, &c)...)
	}

	saveToFile(ethers, c.Output)
}

func saveToFile(ethers []etherEntry, file string) {
	f, err := os.Create(file)
	check(err)

	defer f.Close()
	l := 1
	for _, e := range ethers {
		if e.Mac == "" || e.Name == "" {
			continue
		}
		_, err = f.WriteString(e.Mac + "\t" + e.Name + "\n")
		check(err)
		l++
	}
	fmt.Println(strconv.Itoa(l) + " lines written")
	f.Sync()
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func uciDhcpHostReq(ubus *goubus.Ubus, num int) (empty bool, ether etherEntry) {
	// uci request to current maclist
	// 1/2. Marshal json request
	var uciReq goubus.UbusUciRequest
	err := json.Unmarshal([]byte(`
			{
							 "config": "dhcp",
							 "section": "@host[`+strconv.Itoa(num)+`]",
							 "option": "mac"
			 }
			 `), &uciReq)

	check(err)
	// 2/2. Execute
	res, err := ubus.UciGetConfig(0, uciReq)

	if err != nil {
		if err.Error() == "Empty response" {
			return true, etherEntry{}
		}
		log.Fatal(err)
	}

	mac := fmt.Sprintf("%v", res.Value)

	uciReq.Option = "name"
	// 2/2. Execute
	res, err = ubus.UciGetConfig(0, uciReq)

	check(err)

	name := fmt.Sprintf("%v", res.Value)

	return false, etherEntry{mac, name}
}

func extractEthersDevice(device string, c *conf) []etherEntry {
	var ethers []etherEntry
	ubus := goubus.Ubus{
		URL: "http://" + device + "/ubus",
	}

	setUbusAuth(&ubus, c)

	// Auth
	_, err := ubus.AuthLogin()
	check(err)

	i := 0
	for {
		empty, res := uciDhcpHostReq(&ubus, i)
		ethers = append(ethers, res)
		if empty {
			break
		}
		i++
	}

	return ethers
}

func loadConfig() conf {
	var configF string
	c := conf{}

	flag.StringVar(&configF, "config", "settings.yml", "yaml file used for settings")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config.Load(configF)
	c.Devices = config.GetStringSlice("devices")
	c.User = config.GetUser()
	c.Password = config.GetPass()
	c.Output = config.GetOutput()

	return c
}

func setUbusAuth(ubus *goubus.Ubus, c *conf) {
	ubus.Username = c.User
	ubus.Password = c.Password
}
