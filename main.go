package main

import (
	"fmt"

	"github.com/pborman/getopt/v2"
	//  "log"
)

// Cfg is the main config object
var Cfg Config

func main() {
	var err error
	optConfig := getopt.StringLong("config", 'c', "", "Config file ini")
	getopt.Parse()

	if *optConfig != "" {
		err = Cfg.load(*optConfig)
		if err != nil {
			panic(err)
		}
	}

	controller := Controller{}
	fmt.Println("Running controller now")

	controller.start()
	fmt.Println("Controller started")
}
