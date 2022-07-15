package main

import (
	"fmt"

	"gopkg.in/ini.v1"
	//  "log"
)

// Config is the main ini config
var Config *ini.File

func main() {
	var err error
	Config, err = loadConfig("~/.sotd.ini")
	if err != nil {
		panic(err)
	}

	controller := Controller{}
	fmt.Println("Running controller now")

	controller.start()
	fmt.Println("Controller started")
}
